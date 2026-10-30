package lsp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/siml/lsp/jsonrpc"
)

// A client of the server over in-memory pipes, driving it as an editor's client does.

var repository = func() string {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		panic(err)
	}
	return root
}()

type response struct {
	Result json.RawMessage
	Error  *ResponseError
}

type client struct {
	t       *testing.T
	in      *io.PipeWriter
	w       *jsonrpc.FrameWriter
	done    chan error
	mu      sync.Mutex
	nextID  int
	pending map[string]chan response
	notes   []message
	wake    chan struct{}
}

// Starts a server with opts and a client of it; the client shuts it down when the test ends.
func connect(t *testing.T, opts Options) *client {
	t.Helper()
	toServer, fromClient := io.Pipe()
	toClient, fromServer := io.Pipe()
	c := &client{
		t:       t,
		in:      fromClient,
		w:       jsonrpc.NewFrameWriter(fromClient),
		done:    make(chan error, 1),
		pending: map[string]chan response{},
		wake:    make(chan struct{}, 1),
	}
	go func() {
		_, err := Serve(toServer, fromServer, opts)
		fromServer.Close()
		c.done <- err
	}()
	go c.read(toClient)
	t.Cleanup(func() {
		c.request("shutdown", nil)
		c.w.Notify("exit", nil)
		select {
		case err := <-c.done:
			if err != nil {
				t.Errorf("the server ended with %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("the server did not exit")
		}
		c.in.Close()
	})
	return c
}

func (c *client) read(from io.Reader) {
	reader := jsonrpc.NewFrameReader(from)
	for {
		body, err := reader.Read()
		if err != nil {
			return
		}
		var msg message
		if json.Unmarshal(body, &msg) != nil {
			continue
		}
		if msg.Method == "" {
			c.mu.Lock()
			ch := c.pending[string(msg.ID)]
			delete(c.pending, string(msg.ID))
			c.mu.Unlock()
			if ch != nil {
				ch <- response{msg.Result, msg.Error}
			}
			continue
		}
		// A request from the server, such as client/registerCapability, is answered as a client answers it.
		if len(msg.ID) > 0 {
			c.w.Respond(msg.ID, nil, nil)
		}
		c.mu.Lock()
		c.notes = append(c.notes, msg)
		c.mu.Unlock()
		select {
		case c.wake <- struct{}{}:
		default:
		}
	}
}

func (c *client) send(method string, params any) response {
	c.mu.Lock()
	c.nextID++
	n := c.nextID
	ch := make(chan response, 1)
	c.pending[fmt.Sprint(n)] = ch
	c.mu.Unlock()
	c.w.Request(n, method, params)
	select {
	case r := <-ch:
		return r
	case <-time.After(60 * time.Second):
		c.t.Fatalf("no answer to %s", method)
		return response{}
	}
}

// A request's result, decoded into out when out is not nil; the test fails on an error.
func (c *client) request(method string, params any, out ...any) response {
	c.t.Helper()
	r := c.send(method, params)
	if len(out) > 0 {
		if r.Error != nil {
			c.t.Fatalf("%s: %s", method, r.Error.Message)
		}
		if err := json.Unmarshal(r.Result, out[0]); err != nil {
			c.t.Fatalf("%s: %v in %s", method, err, r.Result)
		}
	}
	return r
}

func (c *client) notify(method string, params any) { c.w.Notify(method, params) }

// The first notification of method whose params match, waiting up to 10 s for it.
func (c *client) notification(method string, matches func(json.RawMessage) bool) json.RawMessage {
	c.t.Helper()
	end := time.Now().Add(10 * time.Second)
	for {
		c.mu.Lock()
		for _, n := range c.notes {
			if n.Method == method && (matches == nil || matches(n.Params)) {
				c.mu.Unlock()
				return n.Params
			}
		}
		c.mu.Unlock()
		if time.Now().After(end) {
			c.t.Fatalf("the server sent no %s that matches within 10 s", method)
		}
		select {
		case <-c.wake:
		case <-time.After(100 * time.Millisecond):
		}
	}
}

type published struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// The diagnostics published for uri that pass ok, waiting for them.
func (c *client) diagnostics(uri string, ok func([]Diagnostic) bool) []Diagnostic {
	c.t.Helper()
	var found published
	c.notification("textDocument/publishDiagnostics", func(params json.RawMessage) bool {
		var p published
		return json.Unmarshal(params, &p) == nil && p.URI == uri && (ok == nil || ok(p.Diagnostics)) && func() bool { found = p; return true }()
	})
	return found.Diagnostics
}

func (c *client) initialize(root string, capabilities any) {
	c.t.Helper()
	var result map[string]any
	c.request("initialize", map[string]any{
		"processId":    os.Getpid(),
		"rootUri":      pathURI(root),
		"capabilities": capabilities,
	}, &result)
	c.notify("initialized", map[string]any{})
}

func (c *client) open(uri, text string) {
	c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "siml", "version": 1, "text": text}})
}

func (c *client) change(uri string, version int, text string) {
	c.notify("textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": uri, "version": version},
		"contentChanges": []any{map[string]any{"text": text}},
	})
}

func codes(diags []Diagnostic) []string {
	out := []string{}
	for _, d := range diags {
		out = append(out, d.Code)
	}
	return out
}

func lines(parts ...string) string { return strings.Join(parts, "\n") }

func fixture(t *testing.T, name string) (path, uri, text string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return path, pathURI(path), string(data)
}

func textDocument(uri string) map[string]any { return map[string]any{"uri": uri} }

func pos(line, character int) Position { return Position{line, character} }

// The text with the edits made, as an editor makes a workspace edit. The text is ASCII, so characters are bytes.
func applied(text string, edits []TextEdit) string {
	offset := func(p Position) int {
		ls := strings.SplitAfter(text, "\n")
		n := 0
		for i := 0; i < p.Line && i < len(ls); i++ {
			n += len(ls[i])
		}
		return n + p.Character
	}
	sorted := append([]TextEdit{}, edits...)
	for i := range sorted {
		for j := i + 1; j < len(sorted); j++ {
			if offset(sorted[j].Range.Start) > offset(sorted[i].Range.Start) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	for _, e := range sorted {
		text = text[:offset(e.Range.Start)] + e.NewText + text[offset(e.Range.End):]
	}
	return text
}
