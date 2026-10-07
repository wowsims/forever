package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"strings"
)

const (
	codeParseError     = -32700
	codeMethodNotFound = -32601
	codeInternalError  = -32603

	messageTypeLog = 4
)

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

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
}

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type lspServer struct {
	out      io.Writer
	ws       *workspace
	trace    bool
	shutdown bool

	ranks map[string]int32
}

func serveLSP(in io.Reader, out io.Writer) (shutdownRequested bool, err error) {
	server := &lspServer{out: out, ws: newWorkspace(), trace: true, ranks: map[string]int32{}}
	reader := textproto.NewReader(bufio.NewReader(in))

	for {
		body, err := readFrame(reader)
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

func readFrame(reader *textproto.Reader) ([]byte, error) {
	header, err := reader.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	length, err := strconv.Atoi(header.Get("Content-Length"))
	if err != nil || length < 0 {
		return nil, fmt.Errorf("a frame without a Content-Length: %v", header)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(reader.R, body); err != nil {
		return nil, err
	}
	return body, nil
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
	case "wowsims-spelldata/setRank":
		var params struct {
			URI       string `json:"uri"`
			Parameter string `json:"parameter"`
			Function  string `json:"function"`
			Rank      int32  `json:"rank"`
		}

		if err := json.Unmarshal(msg.Params, &params); err != nil {
			if request {
				s.respond(msg.ID, nil, &rpcError{
					Code:    codeParseError,
					Message: err.Error(),
				})
			}
			return
		}

		key := fmt.Sprintf(
			"%s:%s:%s",
			params.URI,
			params.Function,
			params.Parameter,
		)

		s.ranks[key] = params.Rank

		if request {
			s.respond(msg.ID, nil, nil)
		}

	case "initialize":
		var params struct {
			InitializationOptions struct {
				Trace string `json:"trace"`
			} `json:"initializationOptions"`
		}
		_ = json.Unmarshal(msg.Params, &params)
		s.trace = params.InitializationOptions.Trace != "off"
		s.respond(msg.ID, map[string]any{
			"capabilities": map[string]any{
				"hoverProvider": true,
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

	case "textDocument/didOpen", "textDocument/didChange", "textDocument/didClose", "textDocument/hover":
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

	markdown, trace, ok := s.ws.hover(text, at.Line, at.Character, uri, s.ranks)
	s.log(strings.Join(trace, "\n"))
	if !ok {
		return nil
	}
	return map[string]any{"contents": map[string]any{"kind": "markdown", "value": markdown}}
}

func (s *lspServer) log(message string) {
	if !s.trace || message == "" {
		return
	}
	s.notify("window/logMessage", map[string]any{"type": messageTypeLog, "message": message})
}

func (s *lspServer) respond(id json.RawMessage, result any, rpcErr *rpcError) {
	msg := rpcMessage{JSONRPC: "2.0", ID: id, Error: rpcErr}
	if rpcErr == nil {
		encoded, err := json.Marshal(result)
		if err != nil {
			msg.Error = &rpcError{Code: codeInternalError, Message: err.Error()}
		} else {
			msg.Result = encoded
		}
	}
	s.write(msg)
}

func (s *lspServer) notify(method string, params any) {
	encoded, err := json.Marshal(params)
	if err != nil {
		return
	}
	s.write(rpcMessage{JSONRPC: "2.0", Method: method, Params: encoded})
}

func (s *lspServer) write(msg rpcMessage) {
	body, err := json.Marshal(msg)
	if err != nil {
		return
	}
	fmt.Fprintf(s.out, "Content-Length: %d\r\n\r\n%s", len(body), body)
}
