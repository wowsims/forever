package printer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/wowsims/forever/sim/siml/lsp/jsonrpc"
)

const (
	codeParseError     = jsonrpc.CodeParseError
	codeMethodNotFound = jsonrpc.CodeMethodNotFound
	codeInvalidParams  = jsonrpc.CodeInvalidParams
	codeInternalError  = jsonrpc.CodeInternalError

	messageTypeLog = 4

	// The command a code action runs to pick the rank a Ladder.Each callback parameter hovers as.
	setRankCommand = "wowsims-spelldata.setRank"
)

type rpcMessage = jsonrpc.Message

type rpcError = jsonrpc.ResponseError

// The params of every textDocument method the server reads.
type documentParams struct {
	TextDocument struct {
		URI  string `json:"uri"`
		Text string `json:"text"`
	} `json:"textDocument"`
	ContentChanges []struct {
		// Nil where the change is the whole text.
		Range *struct {
			Start position `json:"start"`
			End   position `json:"end"`
		} `json:"range"`
		Text string `json:"text"`
	} `json:"contentChanges"`
	Position position `json:"position"`
	Range    struct {
		Start position `json:"start"`
	} `json:"range"`
	Context struct {
		Only []string `json:"only"`
	} `json:"context"`
}

// The argument of setRankCommand: the rank hovers in the document read for the callback parameter.
type rankPick struct {
	URI       string `json:"uri"`
	Parameter string `json:"parameter"`
	Function  string `json:"function"`
	Rank      int32  `json:"rank"`
}

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type lspServer struct {
	w        *jsonrpc.FrameWriter
	ws       *workspace
	trace    bool
	shutdown bool
	picks    *rankPicks
}

func serveLSP(in io.Reader, out io.Writer) (shutdownRequested bool, err error) {
	server := &lspServer{w: jsonrpc.NewFrameWriter(out), ws: newWorkspace(), trace: true, picks: &rankPicks{ranks: map[string]int32{}}}
	reader := jsonrpc.NewFrameReader(in)

	for {
		body, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return server.shutdown, nil
		}
		if err != nil {
			return server.shutdown, err
		}

		var msg rpcMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			server.respond(json.RawMessage("null"), nil, &rpcError{Code: codeParseError, Message: err.Error()})
			continue
		}
		if msg.Method == "exit" {
			return server.shutdown, nil
		}
		server.handle(msg)
	}
}

func (s *lspServer) handle(msg rpcMessage) {
	request := len(msg.ID) > 0
	defer func() {
		if r := recover(); r != nil {
			s.log(fmt.Sprintf("✗ %s panicked: %v", msg.Method, r))
			if request {
				s.respond(msg.ID, nil, &rpcError{Code: codeInternalError, Message: fmt.Sprint(r)})
			}
		}
	}()

	switch msg.Method {
	case "initialize":
		var params struct {
			InitializationOptions struct {
				Trace     string `json:"trace"`
				RankLinks bool   `json:"rankLinks"`
			} `json:"initializationOptions"`
		}
		_ = json.Unmarshal(msg.Params, &params)
		s.trace = params.InitializationOptions.Trace != "off"
		s.picks.links = params.InitializationOptions.RankLinks
		s.respond(msg.ID, map[string]any{
			"capabilities": map[string]any{
				"hoverProvider":          true,
				"codeActionProvider":     true,
				"executeCommandProvider": map[string]any{"commands": []string{setRankCommand}},
				"textDocumentSync": map[string]any{
					"openClose": true,
					"change":    2,
				},
			},
			"serverInfo": map[string]any{"name": "wowsims-spelldata"},
		}, nil)

	case "shutdown":
		s.shutdown = true
		s.respond(msg.ID, nil, nil)

	case "workspace/executeCommand":
		var params struct {
			Command   string            `json:"command"`
			Arguments []json.RawMessage `json:"arguments"`
		}
		var pick rankPick
		if err := json.Unmarshal(msg.Params, &params); err != nil || params.Command != setRankCommand || len(params.Arguments) != 1 {
			s.respond(msg.ID, nil, &rpcError{Code: codeInvalidParams, Message: "the command is " + setRankCommand + " with one {uri, function, parameter, rank}"})
			return
		}
		if err := json.Unmarshal(params.Arguments[0], &pick); err != nil {
			s.respond(msg.ID, nil, &rpcError{Code: codeInvalidParams, Message: err.Error()})
			return
		}
		s.picks.ranks[rankKey(pick.URI, pick.Function, pick.Parameter)] = pick.Rank
		s.respond(msg.ID, nil, nil)

	case "textDocument/didOpen", "textDocument/didChange", "textDocument/didClose", "textDocument/hover", "textDocument/codeAction":
		var params documentParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			if request {
				s.respond(msg.ID, nil, &rpcError{Code: codeParseError, Message: err.Error()})
			}
			return
		}
		uri := params.TextDocument.URI
		switch msg.Method {
		case "textDocument/didOpen":
			s.ws.update(uri, params.TextDocument.Text, true)
		case "textDocument/didChange":
			text := s.ws.buffers[uriPath(uri)]
			for _, change := range params.ContentChanges {
				if change.Range == nil {
					text = change.Text
					continue
				}
				start := offsetOf(text, change.Range.Start.Line, change.Range.Start.Character)
				end := max(start, offsetOf(text, change.Range.End.Line, change.Range.End.Character))
				text = text[:start] + change.Text + text[end:]
			}
			s.ws.update(uri, text, true)
		case "textDocument/didClose":
			s.ws.update(uri, "", false)
		case "textDocument/hover":
			s.respond(msg.ID, s.hover(uri, params.Position), nil)
		case "textDocument/codeAction":
			s.respond(msg.ID, s.rankActions(uri, params.Range.Start, params.Context.Only), nil)
		}

	default:
		if request {
			s.respond(msg.ID, nil, &rpcError{Code: codeMethodNotFound, Message: "method not found: " + msg.Method})
		}
	}
}

func (s *lspServer) hover(uri string, at position) any {
	text, err := s.ws.text(uriPath(uri))
	if err != nil {
		s.log("✗ " + err.Error())
		return nil
	}

	markdown, trace, ok := s.ws.hover(text, at.Line, at.Character, uri, s.picks)
	s.log(strings.Join(trace, "\n"))
	if !ok {
		return nil
	}
	return map[string]any{"contents": map[string]any{"kind": "markdown", "value": markdown}}
}

// A setRankCommand for each other rank of the Ladder.Each callback parameter a hover at the position
// reads through. A request for some kinds of action gets none: Zed asks for kinds on save and runs
// what comes back, and a bare command has no kind to filter it out.
func (s *lspServer) rankActions(uri string, at position, only []string) []any {
	actions := []any{}
	if len(only) > 0 {
		return actions
	}
	text, err := s.ws.text(uriPath(uri))
	if err != nil {
		return actions
	}
	rank, family, ok := s.ws.rankAt(text, at.Line, at.Character, uri, s.picks)
	if !ok {
		return actions
	}
	for n := int32(1); n <= family.ladder.Len(); n++ {
		if n == rank.rank {
			continue
		}
		actions = append(actions, map[string]any{
			"title":     fmt.Sprintf("Hover %s as %s", rank.parameter, title(family.ladder.Rank(n))),
			"command":   setRankCommand,
			"arguments": []rankPick{{URI: uri, Parameter: rank.parameter, Function: rank.function, Rank: n}},
		})
	}
	return actions
}

func (s *lspServer) log(message string) {
	if !s.trace || message == "" {
		return
	}
	s.notify("window/logMessage", map[string]any{"type": messageTypeLog, "message": message})
}

func (s *lspServer) respond(id json.RawMessage, result any, rpcErr *rpcError) {
	s.w.Respond(id, result, rpcErr)
}

func (s *lspServer) notify(method string, params any) { s.w.Notify(method, params) }
