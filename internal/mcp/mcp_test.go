package mcp

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forgeprint/forgelore/internal/record"
	"github.com/forgeprint/forgelore/internal/store"
)

const modernMeta = `"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{}}`

// session drives a server over a scripted set of lines and returns the
// replies, the way a client would.
func session(t *testing.T, s *Server, lines ...string) []map[string]any {
	t.Helper()
	var out strings.Builder
	if err := s.Serve(strings.NewReader(strings.Join(lines, "\n")+"\n"), &out); err != nil {
		t.Fatalf("serve: %v", err)
	}
	var replies []map[string]any
	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			t.Fatalf("the server wrote a line that is not JSON: %v\n%s", err, scanner.Text())
		}
		replies = append(replies, m)
	}
	return replies
}

// newServer returns a server over a store holding one fix.
func newServer(t *testing.T) (*Server, *record.Record) {
	t.Helper()
	root := filepath.Join(t.TempDir(), ".forgelore")
	st := store.New(root)
	if err := st.Init(); err != nil {
		t.Fatal(err)
	}
	id, err := record.NewULID()
	if err != nil {
		t.Fatal(err)
	}
	r := &record.Record{
		Schema: record.CurrentSchema, ID: id, Type: record.TypeFix,
		Scope: record.ScopeTeam, Title: "greet lives in internal/greeter; import it",
		Created: time.Now().UTC().Truncate(time.Second), Source: record.SourceUser,
		Fingerprint: "cd023fb609411574", Body: "the body",
	}
	if _, err := st.Put(r); err != nil {
		t.Fatal(err)
	}
	return New(st, "test"), r
}

func result(t *testing.T, reply map[string]any) map[string]any {
	t.Helper()
	if e, ok := reply["error"]; ok {
		t.Fatalf("expected a result, got error %v", e)
	}
	res, ok := reply["result"].(map[string]any)
	if !ok {
		t.Fatalf("no result in %v", reply)
	}
	return res
}

func rpcErrorOf(t *testing.T, reply map[string]any) (int, map[string]any) {
	t.Helper()
	e, ok := reply["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected an error, got %v", reply)
	}
	code, _ := e["code"].(float64)
	return int(code), e
}

func TestModernDiscover(t *testing.T) {
	s, _ := newServer(t)
	replies := session(t, s, `{"jsonrpc":"2.0","id":"d1","method":"server/discover","params":{`+modernMeta+`}}`)
	if len(replies) != 1 {
		t.Fatalf("got %d replies", len(replies))
	}
	res := result(t, replies[0])

	if res["resultType"] != "complete" {
		t.Errorf("resultType = %v", res["resultType"])
	}
	versions, _ := res["supportedVersions"].([]any)
	if len(versions) != 2 || versions[0] != Modern {
		t.Errorf("supportedVersions = %v", versions)
	}
	meta, _ := res["_meta"].(map[string]any)
	if _, ok := meta[metaServerInfo]; !ok {
		t.Errorf("no server info in _meta: %v", res["_meta"])
	}
}

// TestIDsSurviveUnchanged: Claude Code sends a string id. A server that
// assumed numbers would answer requests nobody asked.
func TestIDsSurviveUnchanged(t *testing.T) {
	s, _ := newServer(t)
	for _, id := range []string{`"server-discover-probe-1"`, `7`} {
		replies := session(t, s, `{"jsonrpc":"2.0","id":`+id+`,"method":"server/discover","params":{`+modernMeta+`}}`)
		got, _ := json.Marshal(replies[0]["id"])
		if string(got) != id {
			t.Errorf("id %s came back as %s", id, got)
		}
	}
}

func TestModernToolsListAndCall(t *testing.T) {
	s, _ := newServer(t)
	replies := session(t, s,
		`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{`+modernMeta+`}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{`+modernMeta+`,"name":"search","arguments":{"query":"greet"}}}`)

	list := result(t, replies[0])
	tools, _ := list["tools"].([]any)
	var names []string
	for _, tool := range tools {
		names = append(names, tool.(map[string]any)["name"].(string))
	}
	want := []string{"search", "get", "recall_error", "propose"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("tools = %v, want %v", names, want)
	}

	// A cache hint is only valid with a lifetime beside it; the real client
	// refuses the list otherwise.
	if _, ok := list["cacheScope"]; ok {
		if _, ok := list["ttlMs"].(float64); !ok {
			t.Error("cacheScope was sent without ttlMs")
		}
	}

	call := result(t, replies[1])
	if call["isError"] != false {
		t.Errorf("isError = %v", call["isError"])
	}
	content, _ := call["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "greet lives in internal/greeter") {
		t.Errorf("search did not find the record: %s", text)
	}
}

// TestLegacySessionNeedsNoMeta: a handshake client sends no _meta on
// anything after initialize, and must still be served.
func TestLegacySession(t *testing.T) {
	s, _ := newServer(t)
	replies := session(t, s,
		`{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"probe","version":"1"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)

	if len(replies) != 2 {
		t.Fatalf("got %d replies, want 2 (a notification is not answered)", len(replies))
	}
	init := result(t, replies[0])
	if init["protocolVersion"] != Legacy {
		t.Errorf("protocolVersion = %v", init["protocolVersion"])
	}
	// A legacy client predates both of these and should not meet them.
	if _, ok := init["resultType"]; ok {
		t.Error("resultType was sent to a legacy client")
	}
	if _, ok := result(t, replies[1])["_meta"]; ok {
		t.Error("_meta was sent to a legacy client")
	}
}

func TestLegacyClientIsOfferedAVersionItCanUse(t *testing.T) {
	s, _ := newServer(t)
	replies := session(t, s,
		`{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2024-01-01"}}`)
	if got := result(t, replies[0])["protocolVersion"]; got != Legacy {
		t.Errorf("protocolVersion = %v, want the newest legacy revision", got)
	}
}

func TestErrorPaths(t *testing.T) {
	cases := []struct {
		why  string
		line string
		code int
	}{
		{"not JSON", `{`, codeParse},
		{"not JSON-RPC", `{"id":1,"method":"tools/list"}`, codeInvalidRequest},
		{"unknown method", `{"jsonrpc":"2.0","id":1,"method":"nope","params":{` + modernMeta + `}}`, codeMethodNotFound},
		{"no protocol version and no handshake", `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`, codeInvalidParams},
		{
			"no client capabilities",
			`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28"}}}`,
			codeInvalidParams,
		},
		{
			"a version nobody speaks",
			`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"1900-01-01","io.modelcontextprotocol/clientCapabilities":{}}}}`,
			codeUnsupportedProtocolVersion,
		},
		{"unknown tool", `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{` + modernMeta + `,"name":"nope"}}`, codeInvalidParams},
	}
	for _, c := range cases {
		s, _ := newServer(t)
		replies := session(t, s, c.line)
		if len(replies) != 1 {
			t.Errorf("%s: got %d replies", c.why, len(replies))
			continue
		}
		code, _ := rpcErrorOf(t, replies[0])
		if code != c.code {
			t.Errorf("%s: code = %d, want %d", c.why, code, c.code)
		}
	}
}

// TestUnsupportedVersionAdvertisesWhatIsSupported: a client has no way to
// retry without the list.
func TestUnsupportedVersionAdvertisesTheAlternatives(t *testing.T) {
	s, _ := newServer(t)
	replies := session(t, s,
		`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"1900-01-01","io.modelcontextprotocol/clientCapabilities":{}}}}`)
	_, e := rpcErrorOf(t, replies[0])
	data, _ := e["data"].(map[string]any)
	supported, _ := data["supported"].([]any)
	if len(supported) != 2 {
		t.Errorf("supported = %v", data["supported"])
	}
	if data["requested"] != "1900-01-01" {
		t.Errorf("requested = %v", data["requested"])
	}
}

// TestNotificationsAreNeverAnswered: a reply to a notification is a protocol
// violation and confuses a client's id bookkeeping.
func TestNotificationsAreNeverAnswered(t *testing.T) {
	s, _ := newServer(t)
	replies := session(t, s,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":1}}`,
		`{"jsonrpc":"2.0","id":null,"method":"notifications/whatever"}`)
	if len(replies) != 0 {
		t.Errorf("got %d replies to notifications: %v", len(replies), replies)
	}
}

// TestStagedAccess: get only works with an id, and an id only comes from a
// search. There is no call that returns the store.
func TestGetNeedsAnID(t *testing.T) {
	s, _ := newServer(t)
	replies := session(t, s,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{`+modernMeta+`,"name":"get","arguments":{"id":"everything"}}}`)
	res := result(t, replies[0])
	if res["isError"] != true {
		t.Errorf("a bogus id was accepted: %v", res)
	}
}

func TestRecallErrorFindsAFixByFingerprint(t *testing.T) {
	s, _ := newServer(t)
	replies := session(t, s,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{`+modernMeta+`,"name":"recall_error","arguments":{"command":"go build ./...","output":"./main.go:5:14: undefined: greet"}}}`)
	res := result(t, replies[0])
	text := res["content"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "greet lives in internal/greeter") {
		t.Errorf("the fix was not recalled:\n%s", text)
	}
}

func TestRecallErrorOnNonsenseSaysSo(t *testing.T) {
	s, _ := newServer(t)
	replies := session(t, s,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{`+modernMeta+`,"name":"recall_error","arguments":{"output":"everything is fine"}}}`)
	res := result(t, replies[0])
	if res["isError"] != false {
		t.Errorf("unrecognised output was reported as a tool error: %v", res)
	}
}

// TestProposeRecordsNothing is the whole point of propose: it leaves a
// request for a person, not a memory.
func TestProposeRecordsNothing(t *testing.T) {
	s, _ := newServer(t)
	before, _, err := s.store.All()
	if err != nil {
		t.Fatal(err)
	}
	session(t, s,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{`+modernMeta+`,"name":"propose","arguments":{"title":"Do the thing","fingerprint":"aaaabbbbccccdddd"}}}`)

	after, _, err := s.store.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Errorf("propose wrote a record: %d then %d", len(before), len(after))
	}
}

func TestProposeNeedsATitle(t *testing.T) {
	s, _ := newServer(t)
	replies := session(t, s,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{`+modernMeta+`,"name":"propose","arguments":{"command":"go build"}}}`)
	if result(t, replies[0])["isError"] != true {
		t.Error("propose accepted a proposal with no title")
	}
}

// TestNothingButProtocolReachesStdout: the specification is explicit, and a
// stray line corrupts the stream for the rest of the session.
func TestEveryStdoutLineIsAMessage(t *testing.T) {
	s, _ := newServer(t)
	var out strings.Builder
	in := strings.Join([]string{
		`{`,
		`{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{` + modernMeta + `}}`,
		``,
		`not json at all`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
	}, "\n") + "\n"
	if err := s.Serve(strings.NewReader(in), &out); err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimRight(out.String(), "\n"), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Errorf("non-message on stdout: %q", line)
			continue
		}
		if m["jsonrpc"] != "2.0" {
			t.Errorf("line is not JSON-RPC: %q", line)
		}
		if strings.Contains(line, "\n") {
			t.Errorf("a message contains a newline: %q", line)
		}
	}
}

// TestCapturedClientTraffic replays what a real client sent. A reply shape
// that would stop it connecting fails here instead of at a connection.
func TestCapturedClientTraffic(t *testing.T) {
	const corpus = "../../testdata/mcp/claude-code/2.1.289"
	files, err := os.ReadDir(corpus)
	if err != nil {
		t.Fatalf("no captured traffic: %v", err)
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".jsonl" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(corpus, f.Name()))
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")

		s, _ := newServer(t)
		replies := session(t, s, lines...)

		answered := 0
		for _, line := range lines {
			var m map[string]any
			if err := json.Unmarshal([]byte(line), &m); err != nil {
				t.Fatalf("%s: %v", f.Name(), err)
			}
			if _, ok := m["id"]; ok {
				answered++
			}
		}
		if len(replies) != answered {
			t.Errorf("%s: %d replies for %d requests", f.Name(), len(replies), answered)
		}
		for i, reply := range replies {
			if _, bad := reply["error"]; bad {
				t.Errorf("%s: reply %d is an error: %v", f.Name(), i, reply["error"])
			}
		}
		t.Logf("%-14s %d request(s) replayed", f.Name(), answered)
	}
}

// TestTheCommandChangesTheFingerprint is a trap worth a test rather than a
// comment: the tool a command names is part of the fingerprint, so recalling
// the same output without the command does not find a memory recorded with
// it. The tool description says so; this keeps it true.
func TestTheCommandChangesTheFingerprint(t *testing.T) {
	s, _ := newServer(t)
	const output = "./main.go:5:14: undefined: greet"

	call := func(args string) string {
		replies := session(t, s,
			`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{`+modernMeta+`,"name":"recall_error","arguments":{`+args+`}}}`)
		return result(t, replies[0])["content"].([]any)[0].(map[string]any)["text"].(string)
	}

	withCommand := call(`"output":"` + output + `","command":"go build ./..."`)
	if !strings.Contains(withCommand, "greet lives in internal/greeter") {
		t.Errorf("the fix was not found with the command:\n%s", withCommand)
	}
	withoutCommand := call(`"output":"` + output + `"`)
	if strings.Contains(withoutCommand, "greet lives in internal/greeter") {
		t.Errorf("the fingerprint ignored the command after all:\n%s", withoutCommand)
	}
}
