package lsp

import (
	"fmt"

	"github.com/wowsims/forever/sim/siml/lsp/jsonrpc"
)

const (
	codeParseError       = jsonrpc.CodeParseError
	codeInvalidRequest   = jsonrpc.CodeInvalidRequest
	codeMethodNotFound   = jsonrpc.CodeMethodNotFound
	codeInvalidParams    = jsonrpc.CodeInvalidParams
	codeInternalError    = jsonrpc.CodeInternalError
	codeRequestCancelled = jsonrpc.CodeRequestCancelled
	codeContentModified  = jsonrpc.CodeContentModified
	codeRequestFailed    = jsonrpc.CodeRequestFailed
)

type message = jsonrpc.Message

// An error a request answers with.
type ResponseError = jsonrpc.ResponseError

func failed(format string, args ...any) *ResponseError {
	return &ResponseError{Code: codeRequestFailed, Message: fmt.Sprintf(format, args...)}
}

func invalidParams(err error) *ResponseError {
	return &ResponseError{Code: codeInvalidParams, Message: err.Error()}
}

func uriPath(uri string) (string, bool) { return jsonrpc.URIPath(uri) }

func pathURI(path string) string { return jsonrpc.PathURI(path) }
