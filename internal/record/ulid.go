package record

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ULIDs identify records: 48 bits of millisecond timestamp followed by 80 bits
// of randomness, written as 26 characters of Crockford base32. Two properties
// matter here. They sort by creation time as plain strings, which makes a
// directory listing chronological, and they need no coordination, which is
// what lets two developers add records at the same time without conflicting
// (ADR-0009).

// alphabet is Crockford base32: the digits, and the letters without I, L, O
// and U, which are the ones people mistype.
const alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// ULIDLen is the length of a ULID in its canonical text form.
const ULIDLen = 26

// ErrRandomnessExhausted means more than 2^80 identifiers were requested
// within one millisecond. It cannot happen in practice; returning an error
// rather than wrapping around keeps the ordering guarantee honest.
var ErrRandomnessExhausted = errors.New("record: ULID randomness exhausted within one millisecond")

// Generator produces ULIDs that are strictly increasing, including within a
// single millisecond and across a clock that steps backwards.
type Generator struct {
	mu     sync.Mutex
	lastMS uint64
	rnd    [10]byte
	seeded bool
}

var defaultGenerator Generator

// NewULID returns a new identifier from the shared generator.
func NewULID() (string, error) { return defaultGenerator.At(time.Now()) }

// At returns an identifier carrying t's millisecond timestamp.
func (g *Generator) At(t time.Time) (string, error) {
	ms := uint64(t.UTC().UnixMilli())

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.seeded && ms <= g.lastMS {
		// The same millisecond, or a clock that went backwards. Keep the
		// previous timestamp and step the random part, so identifiers stay
		// ordered whatever the clock does.
		ms = g.lastMS
		if !increment(&g.rnd) {
			return "", ErrRandomnessExhausted
		}
	} else if _, err := rand.Read(g.rnd[:]); err != nil {
		return "", fmt.Errorf("record: reading randomness: %w", err)
	}
	g.lastMS, g.seeded = ms, true

	var b [16]byte
	for i := 0; i < 6; i++ {
		b[i] = byte(ms >> (8 * (5 - i)))
	}
	copy(b[6:], g.rnd[:])
	return encodeULID(b), nil
}

// increment adds one to the randomness, returning false on overflow.
func increment(b *[10]byte) bool {
	for i := len(b) - 1; i >= 0; i-- {
		b[i]++
		if b[i] != 0 {
			return true
		}
	}
	return false
}

// encodeULID writes 128 bits as 26 base32 characters, most significant first.
// The 26 characters hold 130 bits, so the value is read as if left-padded with
// two zero bits; that is why a valid ULID never starts above '7'.
func encodeULID(b [16]byte) string {
	var out [ULIDLen]byte
	for i := 0; i < ULIDLen; i++ {
		var v byte
		for j := 0; j < 5; j++ {
			k := 5*i + j - 2 // bit index into the 128-bit value
			var bit byte
			if k >= 0 {
				bit = b[k/8] >> (7 - uint(k%8)) & 1
			}
			v = v<<1 | bit
		}
		out[i] = alphabet[v]
	}
	return string(out[:])
}

// ValidULID reports whether s is a ULID in canonical form: 26 uppercase
// Crockford base32 characters, not overflowing 128 bits.
func ValidULID(s string) bool {
	if len(s) != ULIDLen || s[0] > '7' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if decodeChar(s[i]) < 0 {
			return false
		}
	}
	return true
}

// NormalizeULID puts user input into canonical form. Lookups are
// case-insensitive, and Crockford's confusable letters are folded the way the
// encoding intends: O is zero, I and L are one.
func NormalizeULID(s string) (string, bool) {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		switch c {
		case 'O':
			c = '0'
		case 'I', 'L':
			c = '1'
		}
		out = append(out, c)
	}
	id := string(out)
	return id, ValidULID(id)
}

// decodeChar returns the value of a canonical base32 character, or -1.
func decodeChar(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'H':
		return int(c-'A') + 10
	case c == 'J', c == 'K':
		return int(c-'J') + 18
	case c >= 'M' && c <= 'N':
		return int(c-'M') + 20
	case c >= 'P' && c <= 'T':
		return int(c-'P') + 22
	case c >= 'V' && c <= 'Z':
		return int(c-'V') + 27
	}
	return -1
}
