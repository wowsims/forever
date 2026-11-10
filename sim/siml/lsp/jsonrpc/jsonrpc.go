// Package jsonrpc is JSON-RPC 2.0 over a stream, framed as the language server protocol frames it: a Content-Length
// header, a blank line and the body. The SIML language server and the spell data server both speak it.
package jsonrpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	CodeParseError       = -32700
	CodeInvalidRequest   = -32600
	CodeMethodNotFound   = -32601
	CodeInvalidParams    = -32602
	CodeInternalError    = -32603
	CodeRequestCancelled = -32800
	CodeContentModified  = -32801
	CodeRequestFailed    = -32803
)

type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// An error a request answers with.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ResponseError) Error() string { return e.Message }

type FrameReader struct{ r *textproto.Reader }

func NewFrameReader(in io.Reader) *FrameReader {
	return &FrameReader{textproto.NewReader(bufio.NewReaderSize(in, 1<<16))}
}

const maxFrame = 64 << 20

// The body of the next frame; io.EOF when the stream ends between frames.
func (f *FrameReader) Read() ([]byte, error) {
	header, err := f.r.ReadMIMEHeader()
	if err != nil {
		if len(header) == 0 && (err == io.EOF || err == io.ErrUnexpectedEOF) {
			return nil, io.EOF
		}
		return nil, err
	}
	length, err := strconv.Atoi(header.Get("Content-Length"))
	if err != nil || length < 0 {
		return nil, fmt.Errorf("a frame without a Content-Length: %v", header)
	}
	if length > maxFrame {
		return nil, fmt.Errorf("a frame of %d bytes, over the %d MB the server reads", length, maxFrame>>20)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(f.r.R, body); err != nil {
		return nil, err
	}
	return body, nil
}

// Writes whole frames, one at a time, from any goroutine.
type FrameWriter struct {
	mu  sync.Mutex
	out io.Writer
}

func NewFrameWriter(out io.Writer) *FrameWriter { return &FrameWriter{out: out} }

func (w *FrameWriter) Write(msg Message) {
	msg.JSONRPC = "2.0"
	body, err := json.Marshal(msg)
	if err != nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	fmt.Fprintf(w.out, "Content-Length: %d\r\n\r\n%s", len(body), body)
}

func (w *FrameWriter) Respond(id json.RawMessage, result any, rpcErr *ResponseError) {
	msg := Message{ID: id, Error: rpcErr}
	if rpcErr == nil {
		encoded, err := json.Marshal(result)
		if err != nil {
			msg.Error = &ResponseError{Code: CodeInternalError, Message: err.Error()}
		} else {
			msg.Result = encoded
		}
	}
	w.Write(msg)
}

func (w *FrameWriter) Notify(method string, params any) {
	encoded, err := json.Marshal(params)
	if err != nil {
		return
	}
	w.Write(Message{Method: method, Params: encoded})
}

// A request to the other side, whose answer the writer does not wait for.
func (w *FrameWriter) Request(id int, method string, params any) {
	encoded, err := json.Marshal(params)
	if err != nil {
		return
	}
	w.Write(Message{ID: json.RawMessage(strconv.Itoa(id)), Method: method, Params: encoded})
}

// The path of a file: URI; false for another scheme. A URI with a host names a UNC path.
func URIPath(uri string) (string, bool) {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return "", false
	}
	path := u.Path
	if u.Host != "" && u.Host != "localhost" {
		path = "//" + u.Host + path
	} else if runtime.GOOS == "windows" {
		path = strings.TrimPrefix(path, "/")
	}
	return filepath.FromSlash(path), true
}

// The file: URI of a path, as editors write it.
func PathURI(path string) string {
	slashed := filepath.ToSlash(path)
	if host, rest, unc := strings.Cut(strings.TrimPrefix(slashed, "//"), "/"); unc && strings.HasPrefix(slashed, "//") {
		return (&url.URL{Scheme: "file", Host: host, Path: "/" + rest}).String()
	}
	if !strings.HasPrefix(slashed, "/") {
		slashed = "/" + slashed
	}
	return (&url.URL{Scheme: "file", Path: slashed}).String()
}
