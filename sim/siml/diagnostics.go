package siml

import (
	"fmt"
	"sort"
	"strings"
)

// Whether a diagnostic stops a file from loading: a file with an error yields no proto, and one with only
// warnings does. Information is the editor's: a rotation lint that points something out (Lint's info).
type Severity int

const (
	Error Severity = iota
	Warning
	Information
)

func (s Severity) String() string {
	switch s {
	case Warning:
		return "warning"
	case Information:
		return "info"
	}
	return "error"
}

// A place in a file. Line and Col count from 1, and Col counts UTF-16 code units, as editors do.
type Position struct{ Line, Col int }

type Diagnostic struct {
	Path       string // The file it is in.
	Start, End Position
	Severity   Severity
	Code       string // A code of SPEC.md §13 such as unknown-name, a lint code, or unregistered-name.
	Message    string
	Related    *RelatedLocation // For an error of a used file reported at the use line: where it is in that file.
}

// A place in another file that a diagnostic is about.
type RelatedLocation struct {
	Path  string
	Range Range
}

// "path:line:col code message", without the path when there is none.
func (d Diagnostic) String() string {
	at := fmt.Sprintf("%d:%d", d.Start.Line, d.Start.Col)
	if d.Path != "" {
		at = d.Path + ":" + at
	}
	return at + " " + d.Code + " " + d.Message
}

// A file's diagnostics in the order SPEC.md §13 fixes. As an error, one "path:line:col code message" per line.
type Diagnostics []Diagnostic

func (d Diagnostics) Error() string {
	lines := make([]string, len(d))
	for i, x := range d {
		lines[i] = x.String()
	}
	return strings.Join(lines, "\n")
}

// A diagnostic while reading a file, at byte offsets of its text.
type diag struct {
	span
	severity      Severity
	code, message string
	related       *RelatedLocation
}

func hasError(diags []diag) bool {
	for _, d := range diags {
		if d.severity == Error {
			return true
		}
	}
	return false
}

// diags in the order SPEC.md §13 fixes (by where they start, then by where they end, then by code), at their
// positions in the file at path.
func (m *lineMap) diagnostics(path string, diags []diag) Diagnostics {
	sort.SliceStable(diags, func(i, j int) bool {
		a, b := diags[i], diags[j]
		if a.start != b.start {
			return a.start < b.start
		}
		if a.end != b.end {
			return a.end < b.end
		}
		return a.code < b.code
	})
	out := make(Diagnostics, len(diags))
	for i, d := range diags {
		out[i] = Diagnostic{
			Path:     path,
			Start:    m.position(d.start),
			End:      m.position(d.end),
			Severity: d.severity,
			Code:     d.code,
			Message:  d.message,
			Related:  d.related,
		}
	}
	return out
}
