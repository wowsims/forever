package siml

import (
	"encoding/json"
	"slices"
	"sync"
	"unicode/utf16"
	"unicode/utf8"
)

// A file's text read for an editor: the File that Parse gives, with where the text writes each part of what it
// loads, the names it writes, and the statements the language server reads.

// A stretch of a file's text. Its positions count lines and columns from 1, columns in UTF-16 code units, as a
// Diagnostic's do.
type Range struct{ Start, End Position }

// A JSON path into what a file loads, and the text that writes it.
type PathSpan struct {
	Path  []any // JSON names (string) and list indices (int).
	Range Range
}

// A file's text, read. Its spans and the paths of its names lead into the loader's message (SPEC.md §1): the
// APLRotation of an x.apl.siml, the EquipmentSpec of an x.gear.siml, and the IndividualSimSettings of any other file.
type Doc struct {
	path  string
	src   string // The text as read (SPEC.md §2), which byte offsets count in.
	lines *lineMap
	ix    *schemaIndex
	names *Names
	open  Opener     // How the text's use lines are read, which the language services read used files with too.
	typed *typedFile // Nil when the schema does not read.
	file  *File
	err   error
	root  []any // The loader's message's JSON path from IndividualSimSettings.

	spansOnce sync.Once
	spanIndex map[string]Range // The first span of each path within the loader's message, for SpanFor.

	occOnce   sync.Once
	occ       []occurrence // What occurrences reads, once: a Doc does not change after Analyze.
	wordsOnce sync.Once
	words     []wordRole // What classify reads, once.

	pathOnce  sync.Once
	pathLines map[int][]pathSpan // The rotation's spans, with their paths within it, by each line they cover from 0, for PathAt.
}

// Reads src as the file at path, as Parse does; the result holds the diagnostics, and an error only when the schema
// itself does not read.
func Analyze(path string, src []byte, names *Names, open Opener) *Doc {
	d := &Doc{path: path, src: sourceText(src), names: names, open: open}
	d.lines = newLineMap(d.src)
	ix, err := loadIndex()
	if err != nil {
		d.err = err
		return d
	}
	d.ix = ix
	kind := kindOf(path)
	switch kind.Section {
	case sectionAPL:
		d.root = jsonKeys(ix.rotationPath)
	case sectionGear:
		for _, key := range ix.sections[sectionGear].Presence {
			d.root = append(d.root, key)
		}
	}
	self := resolveUse("", path)
	d.typed = readFile(ix, names, d.src, kind, &uses{refused: open == nil, path: self, open: open, chain: []string{self}})
	d.file, d.err = d.read()
	return d
}

func (d *Doc) read() (*File, error) {
	typed := d.typed
	f := &File{Path: d.path, Kind: typed.kind}
	if typed.spec != nil {
		f.Spec = typed.spec.Spec
	}
	for _, st := range typed.tree.top {
		if use, ok := st.(*useStmt); ok {
			f.Uses = append(f.Uses, resolveUse(d.path, use.path.value))
		}
	}
	for _, r := range d.Refs() {
		if r.Known {
			f.Names = append(f.Names, r)
		}
	}
	var raw []diag
	for _, at := range typed.typer.rawRefs {
		raw = append(raw, errorAt("raw-ref", "a file under ui/specs writes ids by name: format writes the names the table holds, and make siml-names names what the sim registers", at))
	}
	f.RawRefs = d.lines.diagnostics(d.path, raw)
	diags := d.lines.diagnostics(d.path, typed.diags)
	if hasError(typed.diags) {
		return f, diags
	}
	f.Settings, f.Categories, f.Warnings = typed.settings, typed.categories, diags
	return f, nil
}

// What Parse gives for the text.
func (d *Doc) File() (*File, error) { return d.file, d.err }

func (d *Doc) Path() string { return d.path }

func (d *Doc) Kind() Kind { return kindOf(d.path) }

// The text as read: a byte order mark at its start dropped, and ill-formed UTF-8 replaced (SPEC.md §2).
func (d *Doc) Text() string { return d.src }

// The errors and warnings, in the order SPEC.md §13 fixes.
func (d *Doc) Diagnostics() Diagnostics {
	if d.file == nil {
		return nil
	}
	var diags Diagnostics
	if d.err != nil {
		diags, _ = d.err.(Diagnostics)
		return diags
	}
	return d.file.Warnings
}

// The text of each JSON path into the loader's message that the text writes: each row, var line and group header,
// each value, action, unit, spell and field value within them, each entry's values and slot row, and the talents
// line. A path can have more than one span; they come in the order typing reads the nodes.
func (d *Doc) Spans() []PathSpan {
	if d.typed == nil {
		return nil
	}
	var out []PathSpan
	for _, s := range d.typed.typer.spans {
		if path, ok := d.within(s.path); ok {
			out = append(out, PathSpan{Path: path, Range: d.rangeOf(s.span)})
		}
	}
	return out
}

// The names the text writes, in file order, with the ones no namespace holds (Known false). Each one's Path is within
// the loader's message, and nil for a name outside it.
func (d *Doc) Refs() []NameRef {
	if d.typed == nil {
		return nil
	}
	refs := slices.Clone(d.typed.typer.refs)
	slices.SortStableFunc(refs, func(a, b nameRef) int { return a.start - b.start })
	// A name read more than once, in a when line or a def's value, is listed once, as it was read last.
	for i := len(refs) - 1; i > 0; i-- {
		if refs[i-1].span == refs[i].span {
			refs = slices.Delete(refs, i-1, i)
		}
	}
	out := make([]NameRef, len(refs))
	for i, r := range refs {
		path, _ := d.within(r.path)
		out[i] = NameRef{
			Name:      r.name,
			Namespace: r.ns,
			Start:     d.lines.position(r.start),
			End:       d.lines.position(r.end),
			ID:        r.id,
			Row:       r.row,
			Path:      path,
			Known:     r.known,
		}
	}
	return out
}

// The rotation's JSON, in the form of SPEC.md §5.4; nil when the text has an error or holds no rotation.
func (d *Doc) RotationJSON() json.RawMessage {
	if d.err != nil || !slices.Contains(d.file.Categories, rotationCategory) {
		return nil
	}
	return json.RawMessage(d.ix.jsonText(d.file.Rotation().ProtoReflect()))
}

// The position of a byte offset of Text.
func (d *Doc) Position(offset int) Position { return d.lines.position(min(max(offset, 0), len(d.src))) }

// The byte offset of Text at p, kept within p's line.
func (d *Doc) Offset(p Position) int {
	line := min(max(p.Line-1, 0), len(d.lines.starts)-1)
	end := len(d.src)
	if line+1 < len(d.lines.starts) {
		end = d.lines.starts[line+1] - 1
	}
	at, col := d.lines.starts[line], 1
	for at < end && col < p.Col {
		r, size := utf8.DecodeRuneInString(d.src[at:])
		at += size
		col += utf16.RuneLen(r)
	}
	return at
}

// The text of the innermost node a rotation path leads into: the path's own span, else the deepest span that holds
// it. The path is within the rotation, in a file of any kind.
func (d *Doc) SpanFor(path []any) (Range, bool) {
	d.spansOnce.Do(func() {
		d.spanIndex = map[string]Range{}
		for _, s := range d.Spans() {
			if _, ok := d.spanIndex[pathKey(s.Path)]; !ok {
				d.spanIndex[pathKey(s.Path)] = s.Range
			}
		}
	})
	full := slices.Concat(d.rotationBase(), path)
	for n := len(full); n >= 0; n-- {
		if r, ok := d.spanIndex[pathKey(full[:n])]; ok {
			return r, true
		}
	}
	return Range{}, false
}

// The rotation path the text at p writes: the path of the shortest span around it, the deeper one of two the same
// length. Empty outside the rotation's spans.
func (d *Doc) PathAt(p Position) []any {
	if d.typed == nil {
		return []any{}
	}
	offset := d.Offset(p)
	d.pathOnce.Do(func() {
		// Each span the rotation holds, with its path within the rotation, by the lines it covers.
		base := d.rotationBase()
		d.pathLines = map[int][]pathSpan{}
		for _, s := range d.typed.typer.spans {
			path, ok := d.within(s.path)
			if !ok || !hasPrefix(path, base) {
				continue
			}
			at := pathSpan{span: s.span, path: path[len(base):]}
			for line := d.lines.line(s.start); line <= d.lines.line(s.end); line++ {
				d.pathLines[line] = append(d.pathLines[line], at)
			}
		}
	})
	var best *pathSpan
	line := d.pathLines[d.lines.line(offset)]
	for i, s := range line {
		if s.start > offset || offset > s.end {
			continue
		}
		if best == nil || s.end-s.start < best.end-best.start || s.end-s.start == best.end-best.start && len(s.path) > len(best.path) {
			best = &line[i]
		}
	}
	if best == nil {
		return []any{}
	}
	return slices.Clone(best.path)
}

// The rotation's path within the loader's message: none in an x.apl.siml, the settings' rotation field otherwise.
func (d *Doc) rotationBase() []any {
	if d.ix == nil || d.Kind().Section == sectionAPL {
		return nil
	}
	return jsonKeys(d.ix.rotationPath)
}

// A path from IndividualSimSettings as a path within the loader's message; false when it leads elsewhere.
func (d *Doc) within(path []any) ([]any, bool) {
	if !hasPrefix(path, d.root) {
		return nil, false
	}
	return path[len(d.root):], true
}

func (d *Doc) rangeOf(s span) Range {
	return Range{Start: d.lines.position(s.start), End: d.lines.position(s.end)}
}

func hasPrefix(path, prefix []any) bool {
	return len(prefix) <= len(path) && slices.Equal(path[:len(prefix)], prefix)
}

// The JSON path of the field from IndividualSimSettings, through the string that holds a spec's Rotation as JSON.
func (p *placed) keys() []any {
	keys := jsonKeys(p.at)
	if p.encoded != nil {
		keys = append(keys, p.encoded.JSONName())
	}
	return append(keys, p.JSON)
}
