package siml

import (
	"errors"
	"fmt"

	"github.com/wowsims/forever/sim/core/proto"
)

// The site's requests (proto/ui.proto), which it sends to lib.wasm or to the web server. Neither reads other files,
// so a use line is refused.

// The error of a request made before the name table is read.
const noNames = "no SIML name table is read"

// Reads req.Text as the file at req.Path, with names. Text over the size a file may have is refused.
func ParseRequest(req *proto.SimlParseRequest, names *Names) (result *proto.SimlParseResult) {
	defer recovered(func(message string) { result = &proto.SimlParseResult{Error: message} })
	if names == nil {
		return &proto.SimlParseResult{Error: noNames}
	}
	if err := textSize(req.Text); err != nil {
		return &proto.SimlParseResult{Error: err.Error()}
	}
	f, err := Parse(req.Path, []byte(req.Text), names, nil)
	if diags, ok := errors.AsType[Diagnostics](err); ok {
		return &proto.SimlParseResult{Spec: f.Spec, Diagnostics: diagnosticProtos(diags)}
	} else if err != nil {
		return &proto.SimlParseResult{Error: err.Error()}
	}
	return &proto.SimlParseResult{Settings: f.Settings, Spec: f.Spec, Categories: f.Categories, Diagnostics: diagnosticProtos(f.Warnings)}
}

// Prints req.Settings as a file of req.Section, with names. A print that does not read back as a file of that kind
// gives no text, and the diagnostics of the read.
func PrintRequest(req *proto.SimlPrintRequest, names *Names) (result *proto.SimlPrintResult) {
	defer recovered(func(message string) { result = &proto.SimlPrintResult{Error: message} })
	if names == nil {
		return &proto.SimlPrintResult{Error: noNames}
	}
	path := "wowsims.siml"
	if req.Section != "" {
		path = "x." + req.Section + ".siml"
	}
	kind, err := KindOf(path)
	if err != nil {
		return &proto.SimlPrintResult{Error: err.Error()}
	}
	text := Print(kind, req.Spec, req.Settings, req.Categories, names)
	_, err = Parse(path, text, names, nil)
	if diags, ok := errors.AsType[Diagnostics](err); ok {
		return &proto.SimlPrintResult{Diagnostics: diagnosticProtos(diags)}
	} else if err != nil {
		return &proto.SimlPrintResult{Error: err.Error()}
	}
	return &proto.SimlPrintResult{Text: string(text)}
}

// Why text is too large to read: the limit of a file (maxSourceSize), so a request costs the reader no more than
// a file does.
func textSize(text string) error {
	if len(text) > maxSourceSize {
		return fmt.Errorf("the text is larger than %d MB", maxSourceSize>>20)
	}
	return nil
}

// Hands a panic's message to fail, so a request the reader or printer cannot take answers with an error rather than
// ending the process that serves it. Deferred.
func recovered(fail func(message string)) {
	if r := recover(); r != nil {
		fail(fmt.Sprint(r))
	}
}

func diagnosticProtos(diags Diagnostics) []*proto.SimlDiagnostic {
	out := make([]*proto.SimlDiagnostic, len(diags))
	for i, d := range diags {
		out[i] = &proto.SimlDiagnostic{
			StartLine: int32(d.Start.Line),
			StartCol:  int32(d.Start.Col),
			EndLine:   int32(d.End.Line),
			EndCol:    int32(d.End.Col),
			Warning:   d.Severity == Warning,
			Code:      d.Code,
			Message:   d.Message,
		}
	}
	return out
}
