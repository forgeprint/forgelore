package redact

import (
	"strings"
	"testing"
)

// Sample credentials are assembled at run time rather than written as
// literals. A file full of realistic-looking tokens would trip the repository's
// own secret scanner, and a test that forces an exception in that scanner is a
// hole someone will eventually walk through.
func sample(prefix string, n int, alphabet string) string {
	var b strings.Builder
	b.WriteString(prefix)
	for i := 0; i < n; i++ {
		b.WriteByte(alphabet[i%len(alphabet)])
	}
	return b.String()
}

const (
	lower = "abcdefghijklmnopqrstuvwxyz"
	upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alnum = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func TestMasksKnownCredentialShapes(t *testing.T) {
	cases := []struct {
		kind   string
		secret string
	}{
		{"github-token", sample("ghp_", 36, alnum)},
		{"slack-token", sample("xoxb-", 24, alnum)},
		{"aws-access-key-id", sample("AKIA", 16, upper)},
		{"google-api-key", sample("AIza", 35, alnum)},
		{"stripe-key", sample("sk_live_", 24, alnum)},
		{"anthropic-key", sample("sk-ant-", 30, alnum)},
		{"openai-key", sample("sk-", 32, alnum)},
	}

	for _, c := range cases {
		t.Run(c.kind, func(t *testing.T) {
			in := "the build failed with " + c.secret + " in the log"
			out, found := Text(in)

			if strings.Contains(out, c.secret) {
				t.Errorf("the secret survived: %q", out)
			}
			if !strings.Contains(out, mark(c.kind)) {
				t.Errorf("output = %q, want it to contain %s", out, mark(c.kind))
			}
			if len(found) == 0 {
				t.Fatal("no findings reported")
			}
			if found[0].Kind != c.kind {
				t.Errorf("reported %q, want %q", found[0].Kind, c.kind)
			}
		})
	}
}

func TestMasksOnlyTheSecretPart(t *testing.T) {
	token := sample("", 32, alnum)

	out, found := Text("Authorization: Bearer " + token)
	if strings.Contains(out, token) {
		t.Errorf("the token survived: %q", out)
	}
	if !strings.Contains(out, "Bearer ") {
		t.Errorf("output = %q, want the header name kept", out)
	}
	if len(found) != 1 || found[0].Kind != "bearer-token" {
		t.Errorf("findings = %v, want one bearer-token", found)
	}

	out, _ = Text("cloning https://user:" + sample("", 20, lower) + "@example.invalid/repo.git")
	if !strings.Contains(out, "https://user:") || !strings.Contains(out, "@example.invalid") {
		t.Errorf("output = %q, want the URL structure kept", out)
	}
	if !strings.Contains(out, mark("url-credentials")) {
		t.Errorf("output = %q, want the password masked", out)
	}
}

func TestMasksAssignedSecrets(t *testing.T) {
	for _, line := range []string{
		`DB_PASSWORD=hunter2hunter2`,
		`api_key: "` + sample("", 24, lower) + `"`,
		`client_secret = ` + sample("", 24, lower),
		`ACCESS_KEY: ` + sample("", 24, lower),
	} {
		out, found := Text(line)
		if len(found) == 0 {
			t.Errorf("Text(%q) found nothing", line)
			continue
		}
		if !strings.Contains(out, markPrefix) {
			t.Errorf("Text(%q) = %q, want a redaction mark", line, out)
		}
	}
}

func TestMasksPrivateKeyBlocks(t *testing.T) {
	in := "before\n-----BEGIN RSA PRIVATE KEY-----\n" + sample("", 40, alnum) + "\n-----END RSA PRIVATE KEY-----\nafter"
	out, found := Text(in)

	if strings.Contains(out, "BEGIN RSA PRIVATE KEY") {
		t.Errorf("the key block survived: %q", out)
	}
	if !strings.HasPrefix(out, "before") || !strings.HasSuffix(out, "after") {
		t.Errorf("output = %q, want the surrounding text kept", out)
	}
	if len(found) != 1 || found[0].Kind != "private-key-block" {
		t.Errorf("findings = %v", found)
	}
}

func TestPrivateSectionsAreNeverStored(t *testing.T) {
	out, found := Text("keep this <private>drop this</private> and this")
	if strings.Contains(out, "drop this") {
		t.Errorf("output = %q, want the private section gone", out)
	}
	if out != "keep this  and this" {
		t.Errorf("output = %q", out)
	}
	if len(found) != 1 || found[0].Kind != "private-section" || found[0].Count != 1 {
		t.Errorf("findings = %v", found)
	}

	out, found = Text("a <private>one</private> b <private>two</private> c")
	if strings.Contains(out, "one") || strings.Contains(out, "two") {
		t.Errorf("output = %q", out)
	}
	if len(found) != 1 || found[0].Count != 2 {
		t.Errorf("findings = %v, want two private sections", found)
	}
}

func TestUnclosedPrivateTagDropsTheRest(t *testing.T) {
	// When it is unclear where the private section ends, the safe reading is
	// that all of it is private.
	out, found := Text("public part <private>secret and everything after it")
	if out != "public part " {
		t.Errorf("output = %q, want everything from the tag onwards dropped", out)
	}
	if len(found) != 1 || found[0].Kind != "private-section" {
		t.Errorf("findings = %v", found)
	}
}

func TestOrdinaryTextIsUntouched(t *testing.T) {
	in := "GOARCH=arm64 go build ./... fails in the link step; set CGO_ENABLED=0.\n\nSee docs/plan.md."
	out, found := Text(in)
	if out != in {
		t.Errorf("Text changed ordinary prose:\n got %q\nwant %q", out, in)
	}
	if found != nil {
		t.Errorf("findings = %v, want none", found)
	}
}

func TestRedactingTwiceChangesNothing(t *testing.T) {
	in := "token " + sample("ghp_", 36, alnum) + " and <private>x</private>"
	once, _ := Text(in)
	twice, found := Text(once)
	if twice != once {
		t.Errorf("second pass changed the text:\n first %q\nsecond %q", once, twice)
	}
	if found != nil {
		t.Errorf("second pass reported %v, want nothing left to mask", found)
	}
}

func TestMergeAndSummary(t *testing.T) {
	a := []Finding{{Kind: "github-token", Count: 1}, {Kind: "private-section", Count: 1}}
	b := []Finding{{Kind: "github-token", Count: 2}}

	merged := Merge(a, b)
	if len(merged) != 2 {
		t.Fatalf("merged = %v", merged)
	}
	if merged[0].Kind != "github-token" || merged[0].Count != 3 {
		t.Errorf("merged[0] = %v, want github-token x3", merged[0])
	}
	if got := Summary(merged); got != "github-token x3, private-section x1" {
		t.Errorf("Summary = %q", got)
	}
	if got := Summary(nil); got != "" {
		t.Errorf("Summary(nil) = %q, want empty", got)
	}
}
