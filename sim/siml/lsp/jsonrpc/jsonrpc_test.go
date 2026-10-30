package jsonrpc

import (
	"io"
	"strings"
	"testing"
)

func TestOversizedFrameIsRefused(t *testing.T) {
	if _, err := NewFrameReader(strings.NewReader("Content-Length: 99999999999\r\n\r\n")).Read(); err == nil || err == io.EOF {
		t.Errorf("a 99999999999 byte frame gave %v", err)
	}
}
