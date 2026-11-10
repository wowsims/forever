package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/siml"
)

// The grammar the extension colours files with, which `siml grammar` writes from the schema.

func TestCommittedGrammarIsTheGeneratedOne(t *testing.T) {
	text, err := simlGrammarText()
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile(filepath.Join("..", "..", "..", ".vscode", "extensions", "wowsims-apl", "syntaxes", "siml.tmLanguage.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(text, committed) {
		t.Error("the committed grammar is not what `go run ./cmd/wowsimcli siml grammar -o .vscode/extensions/wowsims-apl/syntaxes/siml.tmLanguage.json` writes")
	}
}

// A pattern of the grammar, with its lookarounds checked by hand: Go's regexp has none. The grammar's patterns hold
// a lookbehind only at their start and a lookahead only at their end.
type grammarPattern struct {
	name     string
	captures map[int]string
	re       *regexp.Regexp
	behind   *regexp.Regexp // The text before the match must not end in one of these.
	ahead    *regexp.Regexp // The text after the match must (or with negate, must not) start with this.
	negate   bool
	anchored bool // The pattern matches only at the start of the text.
}

var (
	lookbehind = regexp.MustCompile(`^\(\?<!(\[[^\]]*\])\)`)
	lookahead  = regexp.MustCompile(`\(\?(=|!)((?:[^()]|\\\(|\\\))*)\)$`)
)

func compileGrammarPattern(t *testing.T, raw map[string]any) grammarPattern {
	t.Helper()
	p := grammarPattern{captures: map[int]string{}}
	p.name, _ = raw["name"].(string)
	if captures, ok := raw["captures"].(map[string]any); ok {
		for k, v := range captures {
			n, _ := strconv.Atoi(k)
			p.captures[n] = v.(map[string]any)["name"].(string)
		}
	}
	source := raw["match"].(string)
	if m := lookbehind.FindStringSubmatch(source); m != nil {
		p.behind = regexp.MustCompile(m[1] + `$`)
		source = source[len(m[0]):]
	}
	if m := lookahead.FindStringSubmatch(source); m != nil {
		p.ahead, p.negate = regexp.MustCompile(`^`+m[2]), m[1] == "!"
		source = source[:len(source)-len(m[0])]
	}
	p.anchored = strings.HasPrefix(source, "^")
	// A match is tried on every stretch of the text, the whole stretch matching.
	p.re = regexp.MustCompile(`^(?:` + source + `)$`)
	return p
}

// The first match in text whose lookarounds hold, as group indices.
func (p grammarPattern) match(text string) []int {
	for start := 0; start <= len(text); start++ {
		for end := len(text); end >= start; end-- {
			m := p.re.FindStringSubmatchIndex(text[start:end])
			if m == nil {
				continue
			}
			if p.anchored && start > 0 {
				continue
			}
			if p.behind != nil && p.behind.MatchString(text[:start]) {
				continue
			}
			if p.ahead != nil && p.ahead.MatchString(text[end:]) == p.negate {
				continue
			}
			for i := range m {
				if m[i] >= 0 {
					m[i] += start
				}
			}
			return m
		}
	}
	return nil
}

func grammarPatterns(t *testing.T) []grammarPattern {
	t.Helper()
	text, err := simlGrammarText()
	if err != nil {
		t.Fatal(err)
	}
	var grammar struct {
		Patterns   []map[string]any `json:"patterns"`
		Repository map[string]struct {
			Patterns []map[string]any `json:"patterns"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(text, &grammar); err != nil {
		t.Fatal(err)
	}
	var out []grammarPattern
	for _, raw := range append(grammar.Patterns, grammar.Repository["value"].Patterns...) {
		if _, ok := raw["match"]; ok {
			out = append(out, compileGrammarPattern(t, raw))
		}
	}
	return out
}

// The scope the first pattern to match text gives part of it: the whole match, or one of its captures.
func scopeOf(patterns []grammarPattern, text, part string) string {
	for _, p := range patterns {
		m := p.match(text)
		if m == nil {
			continue
		}
		if p.name != "" && text[m[0]:m[1]] == part {
			return p.name
		}
		for group, name := range p.captures {
			if 2*group+1 < len(m) && m[2*group] >= 0 && text[m[2*group]:m[2*group+1]] == part {
				return name
			}
		}
	}
	return ""
}

func schemaWords(t *testing.T) siml.Words {
	t.Helper()
	words, err := siml.GrammarWords()
	if err != nil {
		t.Fatal(err)
	}
	return words
}

func TestGrammarColoursEveryKeyword(t *testing.T) {
	patterns, words := grammarPatterns(t), schemaWords(t)
	for _, word := range words.Keywords {
		if got := scopeOf(patterns, word, word); !strings.HasPrefix(got, "keyword.") && !strings.HasPrefix(got, "constant.language.") {
			t.Errorf("%s is coloured %q", word, got)
		}
	}
}

func TestGrammarColoursHeadsBeforeTheirParenthesis(t *testing.T) {
	patterns, words := grammarPatterns(t), schemaWords(t)
	for _, head := range words.Heads {
		if got := scopeOf(patterns, head+"(", head); got != "support.function.head.siml" {
			t.Errorf("%s( is coloured %q", head, got)
		}
	}
	if got := scopeOf(patterns, `sequence("S")`, "sequence"); got != "support.function.head.siml" {
		t.Errorf(`sequence("S") is coloured %q`, got)
	}
	if got := scopeOf(patterns, `sequence "S" [none]`, "sequence"); got != "keyword.other.action.siml" {
		t.Errorf(`sequence "S" is coloured %q`, got)
	}
}

func TestGrammarColoursUnits(t *testing.T) {
	patterns, words := grammarPatterns(t), schemaWords(t)
	for _, word := range words.Units {
		if got := scopeOf(patterns, word, word); got != "variable.language.unit.siml" {
			t.Errorf("%s is coloured %q", word, got)
		}
	}
	for _, prefix := range words.UnitPrefixes {
		if got := scopeOf(patterns, prefix+"2", prefix+"2"); got != "variable.language.unit.siml" {
			t.Errorf("%s2 is coloured %q", prefix, got)
		}
	}
}

func TestGrammarColoursSectionHeaders(t *testing.T) {
	patterns, words := grammarPatterns(t), schemaWords(t)
	for _, s := range words.Sections {
		if s == "group" {
			continue
		}
		if got := scopeOf(patterns, s+":", s); got != "entity.name.section.siml" {
			t.Errorf("%s: is coloured %q", s, got)
		}
	}
	if got := scopeOf(patterns, `group "Weave":`, `"Weave"`); got != "entity.name.section.group.siml" {
		t.Errorf(`group "Weave": is coloured %q`, got)
	}
}

func TestGrammarColoursLiterals(t *testing.T) {
	patterns := grammarPatterns(t)
	for _, literal := range []string{"5", "1.5s", "400ms", "30%", "1m30s", "5.00"} {
		if got := scopeOf(patterns, literal, literal); got != "constant.numeric.siml" {
			t.Errorf("%s is coloured %q", literal, got)
		}
	}
	for text, want := range map[string]string{
		"${Melee weave}": "variable.other.siml",
		"%{Two words}":   "variable.parameter.placeholder.siml",
		"// a comment":   "comment.line.double-slash.siml",
		"/// a note":     "comment.line.triple-slash.documentation.siml",
	} {
		if got := scopeOf(patterns, text, text); got != want {
			t.Errorf("%s is coloured %q, want %q", text, got, want)
		}
	}
}

func TestGrammarColoursDefLines(t *testing.T) {
	patterns := grammarPatterns(t)
	for _, c := range []struct{ text, part, want string }{
		{"def fits(x, y) = spell(x).cast_time < y", "def", "keyword.control.siml"},
		{"def fits(x, y) = spell(x).cast_time < y", "fits", "entity.name.function.siml"},
		{"def fits(x, y) = spell(x).cast_time < y", "x, y", "meta.parameters.siml"},
		{"def early = current_time < 10s", "early", "entity.name.function.siml"},
	} {
		if got := scopeOf(patterns, c.text, c.part); got != c.want {
			t.Errorf("%s in %s is coloured %q, want %q", c.part, c.text, got, c.want)
		}
	}
}
