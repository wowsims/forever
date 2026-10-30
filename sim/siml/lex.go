package siml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// The tokens and trivia of SIML text (SPEC.md §2). Offsets are byte offsets into the source; a diagnostic's
// column counts UTF-16 code units (lineMap).

type tokenKind int

const (
	tokIdent tokenKind = iota
	tokNumber
	tokString
	tokRaw
	tokVar
	tokPlaceholder
	tokPunct
	tokBad
)

type token struct {
	kind       tokenKind
	text       string
	start, end int
	value      string      // A string's decoded text, a raw token's JSON text, a variable's or placeholder's name.
	err        *tokenError // A token that does not read as one; the parser stops in front of it.
}

type tokenError struct {
	code, message string
	span
}

type triviaKind int

const (
	triviaComment triviaKind = iota // `//` and the rest of the line; text is everything after the `//`.
	triviaNote                      // `///` and the rest of the line; text drops one space after the `///` and the blanks that end it.
	triviaBlank                     // A line holding only spaces and tabs.
)

type trivia struct {
	kind triviaKind
	text string
	line int // 0-based.
	span
}

// Source bytes as text, read as the WHATWG UTF-8 decoder of browsers and Node reads them (SPEC.md §2): a byte order
// mark at the start is skipped, and each maximal subpart of an ill-formed sequence becomes one U+FFFD.
func sourceText(src []byte) string {
	src = bytes.TrimPrefix(src, []byte("\uFEFF"))
	if utf8.Valid(src) {
		return string(src)
	}
	var b strings.Builder
	for len(src) > 0 {
		r, n := utf8.DecodeRune(src)
		if r == utf8.RuneError && n == 1 {
			n = maximalSubpart(src)
		}
		b.WriteRune(r)
		src = src[n:]
	}
	return b.String()
}

// The length of the longest start of a well-formed UTF-8 sequence at the start of s, at least 1, by the byte ranges
// of the Unicode standard's table 3-7.
func maximalSubpart(s []byte) int {
	need, lo, hi := 0, byte(0x80), byte(0xBF)
	switch c := s[0]; {
	case 0xC2 <= c && c <= 0xDF:
		need = 1
	case c == 0xE0:
		need, lo = 2, 0xA0
	case c == 0xED:
		need, hi = 2, 0x9F
	case 0xE1 <= c && c <= 0xEF:
		need = 2
	case c == 0xF0:
		need, lo = 3, 0x90
	case c == 0xF4:
		need, hi = 3, 0x8F
	case 0xF1 <= c && c <= 0xF3:
		need = 3
	}
	n := 1
	for n <= need && n < len(s) && lo <= s[n] && s[n] <= hi {
		n, lo, hi = n+1, 0x80, 0xBF
	}
	return n
}

// Byte offsets into the source.
type span struct{ start, end int }

func (s span) pos() span { return s }

// The start of every line, for line numbers and UTF-16 columns.
type lineMap struct {
	src    string
	starts []int
	ends   []int // The end of each character that takes more bytes than UTF-16 code units, in order.
	saved  []int // How many more bytes than code units the characters up to ends[i] take.
}

func newLineMap(src string) *lineMap {
	m := &lineMap{src: src, starts: []int{0}}
	saved := 0
	for i := 0; i < len(src); i++ {
		switch {
		case src[i] == '\n':
			m.starts = append(m.starts, i+1)
		case src[i] >= utf8.RuneSelf:
			r, size := utf8.DecodeRuneInString(src[i:])
			if n := size - utf16.RuneLen(r); n > 0 {
				saved += n
				m.ends = append(m.ends, i+size)
				m.saved = append(m.saved, saved)
			}
			i += size - 1
		}
	}
	return m
}

// The 0-based line holding offset.
func (m *lineMap) line(offset int) int {
	return sort.Search(len(m.starts), func(i int) bool { return m.starts[i] > offset }) - 1
}

// The byte column of offset in its line, from 0.
func (m *lineMap) column(offset int) int { return offset - m.starts[m.line(offset)] }

// Line i as written, without its line break.
func (m *lineMap) text(i int) string {
	end := len(m.src)
	if i+1 < len(m.starts) {
		end = m.starts[i+1] - 1
	}
	return strings.TrimSuffix(m.src[m.starts[i]:end], "\r")
}

// The 1-based line and UTF-16 column of offset.
func (m *lineMap) position(offset int) Position {
	line := m.line(offset)
	start := m.starts[line]
	return Position{Line: line + 1, Col: offset - start - m.savedBy(offset) + m.savedBy(start) + 1}
}

// How many more bytes than UTF-16 code units the characters that end by offset take.
func (m *lineMap) savedBy(offset int) int {
	i := sort.SearchInts(m.ends, offset+1)
	if i == 0 {
		return 0
	}
	return m.saved[i-1]
}

var (
	punct2   = []string{">=", "<=", "==", "!="}
	punct1   = "<>+-*/()[],=.:#"
	rejected = []struct{ text, code, message string }{
		{"&&", "write-and", "write and in place of &&"},
		{"||", "write-or", "write or in place of ||"},
		{"!", "write-not", "write not in place of !"},
		{"&", "write-and", "write and in place of &"},
		{"|", "write-or", "write or in place of |"},
	}
	durationUnits = []string{"ns", "us", "µs", "ms", "s", "m", "h"}
)

func isIdentStart(c byte) bool { return isUpper(c) || isLower(c) || c == '_' }
func isIdentChar(c byte) bool  { return isIdentStart(c) || isDigit(c) }

// Where the line holding pos ends, before its \n or \r\n.
func contentEnd(src string, pos int) int {
	nl := strings.IndexByte(src[pos:], '\n')
	if nl < 0 {
		return len(src)
	}
	nl += pos
	if nl > pos && src[nl-1] == '\r' {
		return nl - 1
	}
	return nl
}

func lex(src string) ([]token, []trivia, *lineMap) {
	lines := newLineMap(src)
	var tokens []token
	var triv []trivia
	pos := 0
	lineStart := true
	for pos < len(src) {
		if lineStart {
			lineStart = false
			p := pos
			for p < len(src) && (src[p] == ' ' || src[p] == '\t') {
				p++
			}
			crlf := strings.HasPrefix(src[p:], "\r\n")
			if p < len(src) && src[p] == '\n' || crlf {
				triv = append(triv, trivia{kind: triviaBlank, line: lines.line(pos), span: span{pos, p}})
				pos = p + 1
				if crlf {
					pos++
				}
				lineStart = true
				continue
			}
			pos = p
			continue
		}
		c := src[pos]
		switch {
		case c == '\n':
			pos++
			lineStart = true
		case c == ' ' || c == '\t' || strings.HasPrefix(src[pos:], "\r\n"):
			pos++
		case strings.HasPrefix(src[pos:], "//"):
			end := contentEnd(src, pos)
			t := trivia{kind: triviaComment, text: src[pos+2 : end], line: lines.line(pos), span: span{pos, end}}
			if strings.HasPrefix(src[pos:], "///") {
				t.kind = triviaNote
				t.text = strings.TrimPrefix(strings.TrimRight(src[pos+3:end], " \t"), " ")
			}
			triv = append(triv, t)
			pos = end
		default:
			t := scan(src, pos)
			tokens = append(tokens, t)
			pos = t.end
		}
	}
	return tokens, triv, lines
}

func scan(src string, pos int) token {
	c := src[pos]
	switch {
	case c == '`':
		return scanRaw(src, pos)
	case c == '"':
		return scanString(src, pos)
	case c == '$' || c == '%':
		return scanName(src, pos)
	case isDigit(c):
		return scanNumber(src, pos)
	case isIdentStart(c):
		end := pos + 1
		for end < len(src) && isIdentChar(src[end]) {
			end++
		}
		return token{kind: tokIdent, text: src[pos:end], start: pos, end: end}
	}
	for _, p := range punct2 {
		if strings.HasPrefix(src[pos:], p) {
			return token{kind: tokPunct, text: p, start: pos, end: pos + 2}
		}
	}
	for _, r := range rejected {
		if strings.HasPrefix(src[pos:], r.text) {
			end := pos + len(r.text)
			return token{kind: tokPunct, text: r.text, start: pos, end: end, err: &tokenError{r.code, r.message, span{pos, end}}}
		}
	}
	if strings.IndexByte(punct1, c) >= 0 {
		return token{kind: tokPunct, text: src[pos : pos+1], start: pos, end: pos + 1}
	}
	_, size := utf8.DecodeRuneInString(src[pos:])
	end := pos + size
	shown := src[pos:end]
	switch shown {
	case "\r":
		shown = `\r`
	case "\uFEFF": // A byte order mark past the start is invisible, so the message writes its escape.
		shown = `\uFEFF`
	}
	message := fmt.Sprintf("'%s' starts nothing SIML reads", shown)
	return token{kind: tokBad, text: src[pos:end], start: pos, end: end, err: &tokenError{"bad-character", message, span{pos, end}}}
}

// Backticked JSON, which may span lines.
func scanRaw(src string, pos int) token {
	end := strings.IndexByte(src[pos+1:], '`')
	if end < 0 {
		return token{
			kind: tokRaw, text: src[pos:], start: pos, end: len(src), value: src[pos+1:],
			err: &tokenError{"unclosed-raw", "the raw JSON has no closing `", span{pos, pos + 1}},
		}
	}
	end += pos + 1
	return token{kind: tokRaw, text: src[pos : end+1], start: pos, end: end + 1, value: src[pos+1 : end]}
}

// A JSON string on one line.
func scanString(src string, pos int) token {
	i := pos + 1
	for i < len(src) && src[i] != '"' && src[i] != '\n' {
		if src[i] == '\\' && (i+1 >= len(src) || src[i+1] != '\n') {
			i += 2
		} else {
			i++
		}
	}
	if i >= len(src) || src[i] != '"' {
		end := contentEnd(src, pos)
		return token{
			kind: tokString, text: src[pos:end], start: pos, end: end,
			err: &tokenError{"unclosed-string", `the string has no closing " on its line`, span{pos, end}},
		}
	}
	t := token{kind: tokString, text: src[pos : i+1], start: pos, end: i + 1}
	if t.err = badEscape(src, pos, i); t.err == nil {
		_ = json.Unmarshal([]byte(t.text), &t.value) // A string badEscape passes is a JSON string.
	}
	return t
}

// The first control character or escape that makes the string between open and close no JSON string, or an escape
// of a lone surrogate, which names no character; nil when there is none.
func badEscape(src string, open, close int) *tokenError {
	for i := open + 1; i < close; i++ {
		if src[i] < 0x20 {
			return &tokenError{"bad-character", `a control character in a string is written as an escape, such as \t`, span{i, i + 1}}
		}
		if src[i] != '\\' {
			continue
		}
		if strings.IndexByte(`"\/bfnrt`, src[i+1]) >= 0 {
			i++
			continue
		}
		if c, ok := unicodeEscape(src, i, close); ok {
			if utf16.IsSurrogate(c) {
				if low, ok := unicodeEscape(src, i+6, close); c >= 0xDC00 || !ok || low < 0xDC00 || low > 0xDFFF {
					return &tokenError{"bad-character", fmt.Sprintf(`'%s' is a lone surrogate, which names no character`, src[i:i+6]), span{i, i + 6}}
				}
				i += 6
			}
			i += 5
			continue
		}
		end := min(i+2, close)
		if src[i+1] == 'u' {
			end = min(i+6, close)
		}
		message := fmt.Sprintf(`'%s' is no escape: a string escapes only \" \\ \/ \b \f \n \r \t and \uXXXX`, src[i:end])
		return &tokenError{"bad-character", message, span{i, end}}
	}
	return nil
}

// The code unit of the \uXXXX escape at i, before close.
func unicodeEscape(src string, i, close int) (rune, bool) {
	if i+6 > close || src[i] != '\\' || src[i+1] != 'u' || !isHex4(src[i+2:i+6]) {
		return 0, false
	}
	c, _ := strconv.ParseUint(src[i+2:i+6], 16, 16)
	return rune(c), true
}

func isHex4(s string) bool {
	if len(s) != 4 {
		return false
	}
	for i := range len(s) {
		if !isDigit(s[i]) && strings.IndexByte("abcdefABCDEF", s[i]) < 0 {
			return false
		}
	}
	return true
}

// `$name`, `${Name}`, `%name`, `%{Name}`. Inside the braces `\}`, `\\` and `\n` are `}`, `\` and a line feed.
func scanName(src string, pos int) token {
	sigil := src[pos : pos+1]
	kind := tokPlaceholder
	if sigil == "$" {
		kind = tokVar
	}
	if strings.HasPrefix(src[pos+1:], "{") {
		end := contentEnd(src, pos)
		var name strings.Builder
		var bad *tokenError
		for i := pos + 2; i < end; {
			switch c := src[i]; {
			case c == '}':
				return token{kind: kind, text: src[pos : i+1], start: pos, end: i + 1, value: name.String(), err: bad}
			case c != '\\':
				name.WriteByte(c)
				i++
			case i+1 < end && strings.IndexByte(`}\n`, src[i+1]) >= 0:
				if src[i+1] == 'n' {
					name.WriteByte('\n')
				} else {
					name.WriteByte(src[i+1])
				}
				i += 2
			default:
				_, size := utf8.DecodeRuneInString(src[i+1 : end])
				if bad == nil {
					message := fmt.Sprintf(`'%s' is no escape: %s{…} escapes only \} \\ and \n`, src[i:i+1+size], sigil)
					bad = &tokenError{"bad-character", message, span{i, i + 1 + size}}
				}
				i += 1 + size
			}
		}
		message := sigil + "{ needs its } on the same line"
		return token{kind: kind, text: src[pos:end], start: pos, end: end, err: &tokenError{"expected-name", message, span{pos, end}}}
	}
	if pos+1 < len(src) && isIdentStart(src[pos+1]) {
		end := pos + 2
		for end < len(src) && isIdentChar(src[end]) {
			end++
		}
		return token{kind: kind, text: src[pos:end], start: pos, end: end, value: src[pos+1 : end]}
	}
	message := fmt.Sprintf("expected a name after %s: write %[1]sname or %[1]s{Name}", sigil)
	return token{kind: kind, text: sigil, start: pos, end: pos + 1, err: &tokenError{"expected-name", message, span{pos, pos + 1}}}
}

func scanNumber(src string, pos int) token {
	if end, ok := literalEnd(src, pos); ok && (end >= len(src) || src[end] != '.') {
		return token{kind: tokNumber, text: src[pos:end], start: pos, end: end}
	}
	end := pos + 1
	for end < len(src) {
		if isIdentChar(src[end]) || src[end] == '.' {
			end++
		} else if strings.HasPrefix(src[end:], "µ") {
			end += len("µ")
		} else {
			break
		}
	}
	message := fmt.Sprintf("'%s' is no number: a number may end in one of %% ns us µs ms s m h, or be a duration such as 1m30s", src[pos:end])
	return token{kind: tokNumber, text: src[pos:end], start: pos, end: end, err: &tokenError{"bad-literal", message, span{pos, end}}}
}

// Whether s is one literal as the lexer reads it (SPEC.md §2), a leading minus allowed.
func isLiteral(s string) bool {
	s = strings.TrimPrefix(s, "-")
	end, ok := literalEnd(s, 0)
	return ok && end == len(s)
}

// Where the literal at pos ends (SPEC.md §2): a number, then %, a unit, or a compound duration such as 1h2m3.5s,
// with no letter, digit or _ after it. The fraction is tried first and the whole number alone when that fails, so
// the literal of 1.5x ends at its dot, and scanNumber reads a literal that ends at a dot as part of a bad-literal
// run. A loop, not a recursion, so a run such as 1s1s1s… costs no stack however long it is.
func literalEnd(src string, pos int) (int, bool) {
	whole := digits(src, pos)
	if whole == pos {
		return 0, false
	}
	if whole+1 < len(src) && src[whole] == '.' && isDigit(src[whole+1]) {
		if end, ok := suffixEnd(src, digits(src, whole+1)); ok {
			return end, true
		}
	}
	return suffixEnd(src, whole)
}

// The end of the digits at pos.
func digits(src string, pos int) int {
	for pos < len(src) && isDigit(src[pos]) {
		pos++
	}
	return pos
}

// Where a literal whose number ends at pos ends: after %, or after a unit, with any number of number-and-unit pairs
// between, when no letter, digit or _ follows; else at pos, the number alone. Only ms and m share a start, and m then
// needs a digit, which the s is not, so the first unit that fits is the only one that can.
func suffixEnd(src string, pos int) (int, bool) {
	ends := func(e int) bool { return e >= len(src) || !isIdentChar(src[e]) }
	if strings.HasPrefix(src[pos:], "%") && ends(pos+1) {
		return pos + 1, true
	}
	for end := pos; ; {
		unit := ""
		for _, u := range durationUnits {
			if strings.HasPrefix(src[end:], u) {
				unit = u
				break
			}
		}
		if unit == "" {
			break
		}
		end += len(unit)
		if ends(end) {
			return end, true
		}
		next := digits(src, end)
		if next == end {
			break
		}
		if next+1 < len(src) && src[next] == '.' && isDigit(src[next+1]) {
			next = digits(src, next+1)
		}
		end = next
	}
	return pos, ends(pos)
}
