package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/wowsims/forever/sim/siml"
	"google.golang.org/protobuf/encoding/protojson"
)

type simlServeRequest struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
	Text string `json:"text"`
}

type serveDiagnostic struct {
	Line     int    `json:"line"` // Line and col count from 1.
	Col      int    `json:"col"`
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type simlServeResponse struct {
	ID       int               `json:"id"`
	JSON     json.RawMessage   `json:"json,omitempty"`
	Errors   []serveDiagnostic `json:"errors"`
	Warnings []serveDiagnostic `json:"warnings"`
	Uses     []string          `json:"uses"`
	Error    string            `json:"error,omitempty"`
}

// Reads .siml texts for the Vite plugin, tools/vite/siml.mts, one JSON request a line on stdin and one JSON response a
// line on stdout, until stdin closes. The name table is the file -names, read again when its time or size changes, so
// no with_db build is needed.
//
//	go run ./cmd/wowsimcli siml serve -names sim/siml/names.json
//
// Request: {"id": 1, "path": "/abs/ui/specs/hunter/dps/defaults.siml", "text": "siml 1\n..."}. Use lines are read
// against path from the file system.
//
// Response:
//
//	{
//	  "id": 1,
//	  "json": {...},     // the protojson of the message the file's kind loads as (SPEC.md §1), absent when it does not load
//	  "errors": [{"line": 5, "col": 8, "severity": "error", "code": "unknown-name", "message": "..."}],
//	                     // what fails the import: every error, and under ui/specs every warning and raw ref too
//	  "warnings": [...], // the warnings that do not
//	  "uses": ["..."],   // every file the use lines read, or tried to
//	  "error": "..."     // set when the text cannot be read at all, such as a name table that does not load
//	}
func simlServe(out, errOut io.Writer, args []string) error {
	flags := newFlags("serve [-names <names.json>]", errOut)
	namesPath := flags.String("names", "sim/siml/names.json", "the name table")
	if err := parseFlags(flags, args, 0); err != nil {
		return err
	}
	return serveSiml(os.Stdin, out, *namesPath)
}

func serveSiml(in io.Reader, out io.Writer, namesPath string) error {
	table := siml.NamesFile{Path: namesPath}
	reader := bufio.NewReader(in)
	encoder := json.NewEncoder(out)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			var req simlServeRequest
			res := newResponse()
			if err := json.Unmarshal(line, &req); err != nil {
				res.Error = fmt.Sprintf("reading the request: %v", err)
			} else if names, err := table.Get(); err != nil {
				res.Error = err.Error()
			} else {
				res = serveFile(req, names)
			}
			res.ID = req.ID
			if err := encoder.Encode(res); err != nil {
				return err
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// A response with empty lists, which JSON writes as [] rather than null.
func newResponse() simlServeResponse {
	return simlServeResponse{Errors: []serveDiagnostic{}, Warnings: []serveDiagnostic{}, Uses: []string{}}
}

func serveFile(req simlServeRequest, names *siml.Names) (res simlServeResponse) {
	res = newResponse()
	defer func() {
		if r := recover(); r != nil {
			res = newResponse()
			res.Error = fmt.Sprintf("reading %s panicked: %v", req.Path, r)
		}
	}()
	open := func(path string) ([]byte, error) {
		res.Uses = append(res.Uses, path)
		return siml.ReadSource(path)
	}
	f, err := siml.Parse(req.Path, []byte(req.Text), names, open)
	fail, warn, err := failures(req.Path, f, err)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.Errors, res.Warnings = serveDiagnostics(fail), serveDiagnostics(warn)
	if f.Settings != nil {
		if res.JSON, err = protojson.Marshal(f.Message()); err != nil {
			res.Error = err.Error()
		}
	}
	return res
}

func serveDiagnostics(diags siml.Diagnostics) []serveDiagnostic {
	out := []serveDiagnostic{}
	for _, d := range diags {
		out = append(out, serveDiagnostic{Line: d.Start.Line, Col: d.Start.Col, Severity: d.Severity.String(), Code: d.Code, Message: d.Message})
	}
	return out
}
