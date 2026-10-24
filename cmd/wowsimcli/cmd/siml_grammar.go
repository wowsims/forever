package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/siml"
)

// The TextMate grammar of SIML (sim/siml/SPEC.md §2-§4), which the VS Code extension colours files with:
//
//	go run ./cmd/wowsimcli siml grammar -o .vscode/extensions/wowsims-apl/syntaxes/siml.tmLanguage.json
//
// Its keywords, accessor heads, unit words and section names come from the schema, so a word the schema adds is
// coloured on the next build. The language server's semantic tokens colour what only the schema and the name table
// know on top. With no -o it writes to stdout.
func simlGrammar(out, errOut io.Writer, args []string) error {
	flags := newFlags("grammar [-o <file>]", errOut)
	output := flags.String("o", "", "the file to write; stdout when empty")
	if err := parseFlags(flags, args, 0); err != nil {
		return err
	}
	text, err := simlGrammarText()
	if err != nil {
		return err
	}
	return writeOutput(out, *output, text)
}

// Writes text to the file path, or to out when path is empty.
func writeOutput(out io.Writer, path string, text []byte) error {
	if path == "" {
		_, err := out.Write(text)
		return err
	}
	return os.WriteFile(path, text, 0o644)
}

var (
	grammarActionWords   = []string{"cast", "channel", "wait", "move", "group", "sequence", "schedule"}
	grammarLogicWords    = []string{"and", "or", "not"}
	grammarConstantWords = []string{"true", "false", "none"}
)

const (
	grammarNumber   = `[0-9]+(?:\.[0-9]+)?`
	grammarIdentEnd = `(?![A-Za-z0-9_])`
)

// A JSON object whose keys keep their order.
type orderedObject []orderedField

type orderedField struct {
	key   string
	value any
}

func (o orderedObject) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, f := range o {
		if i > 0 {
			b.WriteByte(',')
		}
		key, err := marshalNoEscape(f.key)
		if err != nil {
			return nil, err
		}
		value, err := marshalNoEscape(f.value)
		if err != nil {
			return nil, err
		}
		b.Write(key)
		b.WriteByte(':')
		b.Write(value)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func marshalNoEscape(v any) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(b.Bytes(), []byte("\n")), nil
}

func grammarWords(list []string) string {
	quoted := make([]string, len(list))
	for i, w := range list {
		quoted[i] = regexp.QuoteMeta(w)
	}
	return strings.Join(quoted, "|")
}

func grammarScope(name string) orderedObject { return orderedObject{{"name", name}} }

func grammarNamed(name, match string) orderedObject {
	return orderedObject{{"name", name}, {"match", match}}
}

// The grammar, built from the schema the protos derive.
func simlGrammarObject() (orderedObject, error) {
	words, err := siml.GrammarWords()
	if err != nil {
		return nil, err
	}
	unit := `(?:` + grammarWords(words.DurationUnits) + `)`
	var control []string
	for _, w := range words.Keywords {
		if !slices.Contains(grammarActionWords, w) && !slices.Contains(grammarLogicWords, w) && !slices.Contains(grammarConstantWords, w) {
			control = append(control, w)
		}
	}
	units, prefixes := words.Units, words.UnitPrefixes
	var sections []string
	for _, s := range words.Sections {
		if s != "group" {
			sections = append(sections, s)
		}
	}
	captures := func(names ...string) orderedObject {
		out := orderedObject{}
		for i, n := range names {
			out = append(out, orderedField{string(rune('1' + i)), grammarScope(n)})
		}
		return out
	}
	return orderedObject{
		{"$schema", "https://raw.githubusercontent.com/martinring/tmlanguage/master/tmlanguage.json"},
		{"name", "SIML"},
		{"scopeName", "source.siml"},
		{"patterns", []any{
			grammarNamed("comment.line.triple-slash.documentation.siml", `///.*$`),
			grammarNamed("comment.line.double-slash.siml", `//.*$`),
			orderedObject{
				{"match", `^(group)\s*("(?:[^"\\]|\\.)*")\s*(:)`},
				{"captures", captures("keyword.other.section.siml", "entity.name.section.group.siml", "punctuation.separator.colon.siml")},
			},
			orderedObject{
				{"match", `^\s*(` + grammarWords(sections) + `)\s*(:)`},
				{"captures", captures("entity.name.section.siml", "punctuation.separator.colon.siml")},
			},
			orderedObject{{"match", `^\s*(#)(?=\s)`}, {"captures", captures("keyword.other.hidden.siml")}},
			// `def name(p, …) =`: the def's name, and its parameters (SPEC.md §6.10).
			orderedObject{
				{"match", `^\s*(def)\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:\(([^()]*)\))?\s*(?==)`},
				{"captures", orderedObject{
					{"1", grammarScope("keyword.control.siml")},
					{"2", grammarScope("entity.name.function.siml")},
					{"3", orderedObject{{"name", "meta.parameters.siml"}, {"patterns", []any{grammarNamed("variable.parameter.siml", `[A-Za-z_][A-Za-z0-9_]*`)}}}},
				}},
			},
			orderedObject{{"include", "#value"}},
		}},
		{"repository", orderedObject{{"value", orderedObject{{"patterns", []any{
			orderedObject{
				{"name", "string.quoted.double.siml"},
				{"begin", `"`},
				{"end", `"|$`},
				{"patterns", []any{grammarNamed("constant.character.escape.siml", `\\(?:["\\/bfnrt]|u[0-9A-Fa-f]{4})`)}},
			},
			orderedObject{
				{"name", "meta.embedded.block.json.siml"},
				{"begin", "`"},
				{"end", "`"},
				{"beginCaptures", orderedObject{{"0", grammarScope("punctuation.definition.raw.begin.siml")}}},
				{"endCaptures", orderedObject{{"0", grammarScope("punctuation.definition.raw.end.siml")}}},
				{"patterns", []any{orderedObject{{"include", "source.json"}}}},
			},
			grammarNamed("variable.other.siml", `\$(?:\{[^}\n]*\}|[A-Za-z_][A-Za-z0-9_]*)`),
			grammarNamed("variable.parameter.placeholder.siml", `%(?:\{[^}\n]*\}|[A-Za-z_][A-Za-z0-9_]*)`),
			grammarNamed("constant.numeric.siml", `(?<![A-Za-z0-9_.])`+grammarNumber+`(?:%|`+unit+`(?:`+grammarNumber+unit+`)*)?`+grammarIdentEnd),
			// Before the keywords: `sequence(` is the accessor head, `sequence "S"` the action.
			grammarNamed("support.function.head.siml", `\b(?:`+grammarWords(words.Heads)+`)(?=\s*\()`),
			grammarNamed("keyword.other.action.siml", `\b(?:`+grammarWords(grammarActionWords)+`)`+grammarIdentEnd),
			grammarNamed("keyword.operator.word.siml", `\b(?:`+grammarWords(grammarLogicWords)+`)`+grammarIdentEnd),
			grammarNamed("constant.language.siml", `\b(?:`+grammarWords(grammarConstantWords)+`)`+grammarIdentEnd),
			grammarNamed("keyword.control.siml", `\b(?:`+grammarWords(control)+`)`+grammarIdentEnd),
			grammarNamed("variable.language.unit.siml", `\b(?:`+grammarWords(units)+`|(?:`+grammarWords(prefixes)+`)[1-9][0-9]*)`+grammarIdentEnd),
			grammarNamed("entity.name.function.siml", `\b[a-z_][a-z0-9_]*(?=\()`),
			grammarNamed("keyword.operator.comparison.siml", `==|!=|>=|<=|>|<`),
			grammarNamed("keyword.operator.arithmetic.siml", `[+\-*/]`),
			grammarNamed("keyword.operator.assignment.siml", `=`),
		}}}}}},
	}, nil
}

// The grammar as the file holds it: JSON indented with tabs, and a line break at its end.
func simlGrammarText() ([]byte, error) {
	grammar, err := simlGrammarObject()
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")
	if err := enc.Encode(grammar); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
