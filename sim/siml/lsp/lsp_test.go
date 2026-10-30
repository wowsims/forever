package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/siml"
	"github.com/wowsims/forever/tools/spelldata/printer"
)

// The server driven over the protocol, the way any editor's client drives it.

var smokeURI = pathURI(filepath.Join(repository, "ui", "specs", "hunter", "dps", "apls", "smoke.apl.siml"))

func started(t *testing.T) *client {
	c := connect(t, Options{Row: printer.Row})
	c.initialize(repository, map[string]any{})
	return c
}

func TestInitialize(t *testing.T) {
	c := connect(t, Options{})
	var result struct {
		Capabilities struct {
			CompletionProvider struct {
				TriggerCharacters []string `json:"triggerCharacters"`
			} `json:"completionProvider"`
			HoverProvider              bool `json:"hoverProvider"`
			DocumentFormattingProvider bool `json:"documentFormattingProvider"`
			SemanticTokensProvider     struct {
				Legend struct {
					TokenTypes []string `json:"tokenTypes"`
				} `json:"legend"`
			} `json:"semanticTokensProvider"`
			CodeActionProvider struct {
				CodeActionKinds []string `json:"codeActionKinds"`
			} `json:"codeActionProvider"`
			DocumentLinkProvider map[string]any `json:"documentLinkProvider"`
		} `json:"capabilities"`
	}
	c.request("initialize", map[string]any{"rootUri": pathURI(repository), "capabilities": map[string]any{}}, &result)
	caps := result.Capabilities
	if !slices.Contains(caps.CompletionProvider.TriggerCharacters, ".") || !caps.HoverProvider || !caps.DocumentFormattingProvider {
		t.Errorf("capabilities = %+v", caps)
	}
	if len(caps.SemanticTokensProvider.Legend.TokenTypes) == 0 {
		t.Error("the semantic tokens legend has no token types")
	}
	if !slices.Contains(caps.CodeActionProvider.CodeActionKinds, "quickfix") {
		t.Errorf("code action kinds = %v", caps.CodeActionProvider.CodeActionKinds)
	}
	if !reflect.DeepEqual(caps.DocumentLinkProvider, map[string]any{"resolveProvider": false}) {
		t.Errorf("document link provider = %v", caps.DocumentLinkProvider)
	}
}

func TestOpenedFileGetsDiagnosticsAndQuickFix(t *testing.T) {
	c := started(t)
	c.open(smokeURI, lines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mrk", "  ", ""))
	diags := c.diagnostics(smokeURI, nil)
	if len(diags) != 1 || diags[0].Code != "unknown-name" || diags[0].Range.Start != pos(4, 7) {
		t.Fatalf("diagnostics = %+v, want unknown-name at 4:7", diags)
	}
	if !strings.Contains(diags[0].Message, "did you mean hunters_mark") {
		t.Errorf("message = %q", diags[0].Message)
	}

	// A quick fix changes the unknown name to the parser's candidate.
	var actions []CodeAction
	c.request("textDocument/codeAction", map[string]any{"textDocument": textDocument(smokeURI), "range": diags[0].Range, "context": map[string]any{"diagnostics": diags}}, &actions)
	i := slices.IndexFunc(actions, func(a CodeAction) bool { return a.Kind == "quickfix" })
	if i < 0 {
		t.Fatalf("no quick fix in %+v", actions)
	}
	if actions[i].Title != "Change to hunters_mark" {
		t.Errorf("title = %q", actions[i].Title)
	}
	if want := []TextEdit{{Range: diags[0].Range, NewText: "hunters_mark"}}; !reflect.DeepEqual(actions[i].Edit.Changes[smokeURI], want) {
		t.Errorf("edits = %+v, want %+v", actions[i].Edit.Changes[smokeURI], want)
	}
}

func TestUsePathIsALink(t *testing.T) {
	c := started(t)
	path := filepath.Join(repository, "ui", "specs", "hunter", "dps", "builds", "mm.siml")
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	c.open(pathURI(path), string(text))
	var links []DocumentLink
	c.request("textDocument/documentLink", map[string]any{"textDocument": textDocument(pathURI(path))}, &links)
	var targets []string
	for _, l := range links {
		targets = append(targets, strings.TrimPrefix(l.Target, pathURI(repository)+"/"))
	}
	if want := []string{"ui/specs/hunter/dps/gear_sets/p1.gear.siml", "ui/specs/hunter/dps/apls/turret.apl.siml"}; !reflect.DeepEqual(targets, want) {
		t.Errorf("links = %v, want %v", targets, want)
	}
}

func completionLabels(t *testing.T, c *client, uri string, at Position) []string {
	t.Helper()
	var items []CompletionItem
	c.request("textDocument/completion", map[string]any{"textDocument": textDocument(uri), "position": at}, &items)
	var labels []string
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	return labels
}

func TestCompletionOffersWhatThePositionTakes(t *testing.T) {
	c := started(t)
	c.open(smokeURI, lines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mrk", "  ", ""))
	labels := completionLabels(t, c, smokeURI, pos(5, 2))
	if !slices.Contains(labels, "cast") || !slices.Contains(labels, "wait until") {
		t.Errorf("at the start of a priority row: %v, want cast and wait until", labels)
	}
	if labels := completionLabels(t, c, smokeURI, pos(4, 12)); !slices.Contains(labels, "hunters_mark") {
		t.Errorf("after cast: %v, want hunters_mark", labels)
	}
}

// A document with CRLF line endings, as Git for Windows checks a file out, formats with no edit when only its line
// endings differ from format's, and keeps them when it formats.
func TestFormattingKeepsCRLF(t *testing.T) {
	c := started(t)
	crlf := strings.ReplaceAll(lines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mark", ""), "\n", "\r\n")
	format := func(text string, version int) string {
		t.Helper()
		if version == 1 {
			c.open(smokeURI, text)
		} else {
			c.change(smokeURI, version, text)
		}
		var edits []TextEdit
		c.request("textDocument/formatting", map[string]any{"textDocument": textDocument(smokeURI), "options": map[string]any{"tabSize": 2, "insertSpaces": true}}, &edits)
		if version == 1 && len(edits) > 0 {
			t.Errorf("a formatted document gets edits: %+v", edits)
		}
		return applied(text, edits)
	}
	if got := format(crlf, 1); got != crlf {
		t.Errorf("formatted: %q, want %q", got, crlf)
	}
	if got := format(strings.Replace(crlf, "cast hunters_mark", "cast  hunters_mark", 1), 2); got != crlf {
		t.Errorf("formatted: %q, want %q", got, crlf)
	}
}

func TestHoverNamesASpellOfTheTable(t *testing.T) {
	c := started(t)
	c.open(smokeURI, lines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mrk", ""))
	c.change(smokeURI, 2, lines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mark", ""))
	var hover Hover
	c.request("textDocument/hover", map[string]any{"textDocument": textDocument(smokeURI), "position": pos(4, 9)}, &hover)
	id := tableSpell(t, siml.NamespaceSpell, "hunters_mark").GetSpellId()
	if !strings.Contains(hover.Contents.Value, "Hunter's Mark") || !strings.Contains(hover.Contents.Value, fmt.Sprintf("spell %d", id)) ||
		!strings.Contains(hover.Contents.Value, fmt.Sprintf("https://www.wowhead.com/forever/spell=%d", id)) {
		t.Errorf("hover = %q", hover.Contents.Value)
	}
}

func TestUntitledDocumentIsServed(t *testing.T) {
	c := started(t)
	uri := "untitled:Untitled-1"
	c.open(uri, lines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mrk", ""))
	if diags := c.diagnostics(uri, nil); !reflect.DeepEqual(codes(diags), []string{"unknown-name"}) {
		t.Errorf("diagnostics = %v", codes(diags))
	}
	if labels := completionLabels(t, c, uri, pos(4, 9)); !slices.Contains(labels, "hunters_mark") {
		t.Errorf("completion = %v, want hunters_mark", labels)
	}
}

func TestOtherSchemesAreNotServed(t *testing.T) {
	c := started(t)
	uri := "git:/x/y.apl.siml?ref=HEAD"
	c.open(uri, lines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mrk", ""))
	c.open(smokeURI, lines("siml 1", "spec hunter", ""))
	c.diagnostics(smokeURI, nil)
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, n := range c.notes {
		if n.Method == "textDocument/publishDiagnostics" && strings.Contains(string(n.Params), "git:") {
			t.Errorf("a git: document got diagnostics: %s", n.Params)
		}
	}
}

func TestUsedFileClosedIsReadFromDisk(t *testing.T) {
	c := started(t)
	_, usedURI, usedText := fixture(t, "default.apl.siml")
	_, usingURI, usingText := fixture(t, "hawks.apl.siml")
	c.open(usedURI, strings.Replace(usedText, "var ${Use Summon Hawk} = false\n", "", 1))
	c.open(usingURI, usingText)
	if diags := c.diagnostics(usingURI, nil); !reflect.DeepEqual(codes(diags), []string{"with-unknown-variable"}) {
		t.Errorf("diagnostics = %v, want with-unknown-variable", codes(diags))
	}
	c.notify("textDocument/didClose", map[string]any{"textDocument": textDocument(usedURI)})
	c.diagnostics(usingURI, func(d []Diagnostic) bool { return len(d) == 0 })
}

func TestRenameFromWithList(t *testing.T) {
	c := started(t)
	_, usingURI, usingText := fixture(t, "hawks.apl.siml")
	c.open(usingURI, usingText)
	at := pos(3, 30)
	var edit WorkspaceEdit
	c.request("textDocument/rename", map[string]any{"textDocument": textDocument(usingURI), "position": at, "newName": "Summon hawks"}, &edit)
	counts := map[string]int{}
	for uri, edits := range edit.Changes {
		path, _ := uriPath(uri)
		counts[filepath.Base(path)] = len(edits)
	}
	if want := map[string]int{"default.apl.siml": 5, "hawks.apl.siml": 1}; !reflect.DeepEqual(counts, want) {
		t.Errorf("edits per file = %v, want %v", counts, want)
	}
	refused := c.request("textDocument/rename", map[string]any{"textDocument": textDocument(usingURI), "position": at, "newName": "Use Aimed Shot"})
	if refused.Error == nil || refused.Error.Message != "default.apl.siml has a variable named Use Aimed Shot already." {
		t.Errorf("rename to a taken name = %+v", refused.Error)
	}
}

func TestRegeneratedTableIsRead(t *testing.T) {
	root := t.TempDir()
	tablePath := filepath.Join(root, tableFile)
	table, err := os.ReadFile(filepath.Join(repository, tableFile))
	if err != nil {
		t.Fatal(err)
	}
	var older map[string]any
	if err := json.Unmarshal(table, &older); err != nil {
		t.Fatal(err)
	}
	delete(older["classes"].(map[string]any)["hunter"].(map[string]any)["spell"].(map[string]any), "hunters_mark")
	olderText, _ := json.Marshal(older)
	if err := os.MkdirAll(filepath.Dir(tablePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tablePath, olderText, 0o644); err != nil {
		t.Fatal(err)
	}

	c := connect(t, Options{})
	c.initialize(root, map[string]any{"workspace": map[string]any{"didChangeWatchedFiles": map[string]any{"dynamicRegistration": true}}})
	var registration struct {
		Registrations []struct {
			Method          string `json:"method"`
			RegisterOptions struct {
				Watchers []struct {
					GlobPattern string `json:"globPattern"`
				} `json:"watchers"`
			} `json:"registerOptions"`
		} `json:"registrations"`
	}
	if err := json.Unmarshal(c.notification("client/registerCapability", nil), &registration); err != nil {
		t.Fatal(err)
	}
	var globs []string
	for _, r := range registration.Registrations {
		for _, w := range r.RegisterOptions.Watchers {
			globs = append(globs, w.GlobPattern)
		}
	}
	if !slices.Contains(globs, "**/sim/siml/names.json") {
		t.Errorf("watched = %v, want the table", globs)
	}

	uri := pathURI(filepath.Join(root, "x.apl.siml"))
	c.open(uri, lines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mark", ""))
	if diags := c.diagnostics(uri, nil); !reflect.DeepEqual(codes(diags), []string{"unknown-name"}) {
		t.Errorf("diagnostics = %v, want unknown-name", codes(diags))
	}
	var before rotationAnswer
	c.request("siml/rotation", map[string]any{"uri": uri}, &before)
	if err := os.WriteFile(tablePath, table, 0o644); err != nil {
		t.Fatal(err)
	}
	if labels := completionLabels(t, c, uri, pos(4, 12)); !slices.Contains(labels, "hunters_mark") {
		t.Errorf("the next request does not read the new table: %v", labels)
	}
	c.notify("workspace/didChangeWatchedFiles", map[string]any{"changes": []any{map[string]any{"uri": pathURI(tablePath), "type": 2}}})
	c.diagnostics(uri, func(d []Diagnostic) bool { return len(d) == 0 })
	var after rotationAnswer
	c.request("siml/rotation", map[string]any{"uri": uri}, &after)
	if before.Names == "" || after.Names == before.Names {
		t.Errorf("the panel's names = %q, then %q; want a new value for the new table", before.Names, after.Names)
	}
}

func TestCancelledRequestAnswersCancelled(t *testing.T) {
	release := make(chan struct{})
	c := connect(t, Options{Sim: func(ctx context.Context, root string, req json.RawMessage) (json.RawMessage, error) {
		<-ctx.Done()
		close(release)
		return nil, ctx.Err()
	}})
	c.initialize(repository, map[string]any{})
	path := filepath.Join(repository, "ui", "specs", "hunter", "dps", "apls", "default.apl.siml")
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	c.open(pathURI(path), string(text))
	answer := make(chan response, 1)
	go func() { answer <- c.send("siml/sim", map[string]any{"uri": pathURI(path), "iterations": 10}) }()
	// The request's id is the client's last.
	for {
		c.mu.Lock()
		n := len(c.pending)
		id := c.nextID
		c.mu.Unlock()
		if n > 0 {
			c.notify("$/cancelRequest", map[string]any{"id": id})
			break
		}
		time.Sleep(time.Millisecond)
	}
	r := <-answer
	if r.Error == nil || r.Error.Code != codeRequestCancelled {
		t.Errorf("a cancelled sim answered %+v", r)
	}
	<-release
}

func TestUsedFileErrorPointsIntoIt(t *testing.T) {
	c := started(t)
	_, usedURI, usedText := fixture(t, "default.apl.siml")
	_, usingURI, usingText := fixture(t, "hawks.apl.siml")
	c.open(usedURI, strings.Replace(usedText, "cast mongoose_bite", "cast mongoose_bit", 1))
	c.open(usingURI, usingText)
	diags := c.diagnostics(usingURI, func(d []Diagnostic) bool { return len(d) > 0 })
	if len(diags[0].RelatedInformation) != 1 || diags[0].RelatedInformation[0].Location.URI != usedURI || diags[0].RelatedInformation[0].Location.Range.Start.Line != 26 {
		t.Errorf("diagnostics = %+v, want one related to line 26 of default.apl.siml", diags)
	}
	var folds []FoldingRange
	c.request("textDocument/foldingRange", map[string]any{"textDocument": textDocument(usedURI)}, &folds)
	if !slices.ContainsFunc(folds, func(f FoldingRange) bool { return f.Kind == "region" }) {
		t.Errorf("no section folds as a region: %+v", folds)
	}
}

func TestUsedFileThatIsNotARegularFileIsRefused(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "big.apl.siml")
	if err := os.WriteFile(big, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(big, 17<<20); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{dir, big} {
		if _, err := siml.ReadSource(path); err == nil {
			t.Errorf("%s is read", path)
		}
	}
}

func TestBrokenTableIsReadAgainWhenItChanges(t *testing.T) {
	root := t.TempDir()
	tablePath := filepath.Join(root, tableFile)
	table, err := os.ReadFile(filepath.Join(repository, tableFile))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(tablePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tablePath, []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := started(t)
	uri := pathURI(filepath.Join(root, "x.apl.siml"))
	c.open(uri, lines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mark", ""))
	c.diagnostics(uri, func(d []Diagnostic) bool { return len(d) == 1 && d[0].Severity == severityError })
	if err := os.WriteFile(tablePath, table, 0o644); err != nil {
		t.Fatal(err)
	}
	c.notify("workspace/didChangeWatchedFiles", map[string]any{"changes": []any{map[string]any{"uri": pathURI(tablePath), "type": 2}}})
	c.diagnostics(uri, func(d []Diagnostic) bool { return len(d) == 0 })
}

func TestUNCPathRoundTrips(t *testing.T) {
	const uri = "file://wsl.localhost/Ubuntu/home/a.apl.siml"
	path, ok := uriPath(uri)
	if !ok || filepath.ToSlash(path) != "//wsl.localhost/Ubuntu/home/a.apl.siml" {
		t.Fatalf("uriPath(%s) = %q, %v; want //wsl.localhost/Ubuntu/home/a.apl.siml", uri, path, ok)
	}
	if got := pathURI(path); got != uri {
		t.Errorf("pathURI(%q) = %s, want %s", path, got, uri)
	}
}
