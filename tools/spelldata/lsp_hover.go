package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Each pattern captures the id in group 1.
var spellIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\b(?:Must)?Find\(\s*(\d+)\s*\)`),
	regexp.MustCompile(`\.ByID\(\s*(\d+)\s*\)`),
	regexp.MustCompile(`\bSpellID:\s*(\d+)`),
	regexp.MustCompile(`"spellId":\s*(\d+)`),
	regexp.MustCompile(`\bfromSpellId\(\s*(\d+)\s*\)`),
	regexp.MustCompile(`\bspellId:\s*(\d+)`),
}

const maxSubstitutions = 4

// A name a Go file binds to a chain, with `var`, `=` or `:=`. A name bound inside a function is seen
// from its binding to the end of that function's body, as offsets in its file; to is 0 on a name
// bound at package level, which every file of the package sees.
type declaration struct {
	name     string
	chain    *chain
	file     string
	line     int
	from, to int
}

// A Go file as the hover reads it. The AST is partial where the text does not parse, and nil where it
// states no package clause.
type parsedFile struct {
	path  string
	text  string
	file  *ast.File
	tok   *token.File
	decls []declaration
	// The names the file binds to a value that is not a chain, with why.
	unread map[*ast.Ident]error
}

func parseGo(path, text string) *parsedFile {
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, path, text, parser.SkipObjectResolution)
	f := &parsedFile{path: path, text: text, unread: map[*ast.Ident]error{}}
	// ParseFile answers an empty file, not nil, where the text states no package clause.
	if file == nil || !file.Package.IsValid() {
		return f
	}
	f.file, f.tok = file, fset.File(file.Pos())

	bind := func(name ast.Expr, value ast.Expr) {
		ident, ok := name.(*ast.Ident)
		if !ok || ident.Name == "_" {
			return
		}
		c, err := walkChain(value, f.tok.Offset)
		if err != nil {
			f.unread[ident] = err
			return
		}
		d := declaration{name: ident.Name, chain: c, file: path, line: f.tok.Line(ident.Pos())}
		if body := enclosingBody(f.enclosing(ident.Pos())); body != nil {
			d.from, d.to = f.tok.Offset(ident.Pos()), f.tok.Offset(body.End())
		}
		f.decls = append(f.decls, d)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.ValueSpec:
			if len(n.Names) == len(n.Values) {
				for i, name := range n.Names {
					bind(name, n.Values[i])
				}
			}
		case *ast.AssignStmt:
			if (n.Tok == token.DEFINE || n.Tok == token.ASSIGN) && len(n.Lhs) == len(n.Rhs) {
				for i, name := range n.Lhs {
					bind(name, n.Rhs[i])
				}
			}
		}
		return true
	})
	return f
}

func enclosingBody(nodes []ast.Node) *ast.BlockStmt {
	for i := len(nodes) - 1; i >= 0; i-- {
		switch fn := nodes[i].(type) {
		case *ast.FuncDecl:
			return fn.Body
		case *ast.FuncLit:
			return fn.Body
		}
	}
	return nil
}

// The nodes that hold pos, outermost first. An end counts as inside, so the cursor just past a name
// is on it.
func (f *parsedFile) enclosing(pos token.Pos) []ast.Node {
	var out []ast.Node
	ast.Inspect(f.file, func(node ast.Node) bool {
		if node == nil || pos < node.Pos() || pos > node.End() {
			return false
		}
		out = append(out, node)
		return true
	})
	return out
}

type workspace struct {
	buffers map[string]string
	// The declarations of each Go file read, by folder.
	folders map[string]map[string]cachedFile
	root    string
	current *parsedFile
}

// A file's declarations and what they were read at: the editor's text for a file it holds, else the
// modification time on disk, which is how a change the editor does not report - a checkout, a
// generator, another session - reaches the cache.
type cachedFile struct {
	key   any
	decls []declaration
}

func newWorkspace() *workspace {
	root, _ := moduleRoot()
	return &workspace{buffers: map[string]string{}, folders: map[string]map[string]cachedFile{}, root: root}
}

// The text an editor holds for a file, or with open false, the file as it stands on disk again.
func (w *workspace) update(uri, text string, open bool) {
	path := uriPath(uri)
	if open {
		w.buffers[path] = text
	} else {
		delete(w.buffers, path)
	}
}

// A file as the editor holds it, else as it stands on disk.
func (w *workspace) text(path string) (string, error) {
	if text, open := w.buffers[path]; open {
		return text, nil
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

// The file being hovered, parsed once per text.
func (w *workspace) parse(path, text string) *parsedFile {
	if w.current == nil || w.current.path != path || w.current.text != text {
		w.current = parseGo(path, text)
	}
	return w.current
}

// Every name the package folder binds at package level, and the names the current file binds in the
// function around the offset at, the last binding before at winning. With no current file there is
// no position to scope by, and every binding of the folder is read.
func (w *workspace) declarations(folder string, current *parsedFile, at int, trace *tracer) map[string]declaration {
	files, cached := w.folders[folder]
	if !cached {
		files = map[string]cachedFile{}
		w.folders[folder] = files
	}
	read := w.refresh(folder, files, trace)
	if cached && read == 0 {
		trace.add("declarations %s: cache hit", trace.rel(folder))
	} else {
		trace.add("declarations %s: cache miss, read %d files", trace.rel(folder), read)
	}

	own := current.pathIn(folder)
	paths := slices.Collect(maps.Keys(files))
	if _, listed := files[own]; own != "" && !listed {
		paths = append(paths, own)
	}
	slices.Sort(paths)

	out := map[string]declaration{}
	for _, path := range paths {
		decls := files[path].decls
		if path == own {
			decls = current.decls
		}
		for _, d := range decls {
			if d.to == 0 || current == nil {
				out[d.name] = d
			}
		}
	}
	if own != "" {
		for _, d := range current.decls {
			if d.to != 0 && d.from <= at && at <= d.to {
				out[d.name] = d
			}
		}
	}
	return out
}

// Reads again each Go file of the folder whose text or modification time moved since it was read, and
// drops the ones that are gone. Answers how many files it read or dropped.
func (w *workspace) refresh(folder string, files map[string]cachedFile, trace *tracer) int {
	keys := map[string]any{}
	for path, text := range w.buffers {
		if filepath.Dir(path) == folder && strings.HasSuffix(path, ".go") {
			keys[path] = text
		}
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		trace.add("declarations %s: %v", trace.rel(folder), err)
	}
	for _, entry := range entries {
		path := filepath.Join(folder, entry.Name())
		if _, open := keys[path]; open || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if info, err := entry.Info(); err == nil {
			keys[path] = info.ModTime().UnixNano()
		}
	}

	read := 0
	for path := range files {
		if _, ok := keys[path]; !ok {
			delete(files, path)
			read++
		}
	}
	for path, key := range keys {
		if file, ok := files[path]; ok && file.key == key {
			continue
		}
		read++
		text, err := w.text(path)
		if err != nil {
			delete(files, path)
			continue
		}
		file := w.current
		if file == nil || file.path != path || file.text != text {
			file = parseGo(path, text)
		}
		files[path] = cachedFile{key: key, decls: file.decls}
	}
	return read
}

func (f *parsedFile) pathIn(folder string) string {
	if f == nil || filepath.Dir(f.path) != folder {
		return ""
	}
	return f.path
}

// What a chain hover evaluates, and the name it answers for where the cursor was on a bound name
// rather than on a part of a chain.
type chainHover struct {
	chain *chain
	name  string
}

type tracer struct {
	root  string
	lines []string
}

func (t *tracer) add(format string, args ...any) {
	t.lines = append(t.lines, fmt.Sprintf(format, args...))
}

func (t *tracer) rel(path string) string {
	if t.root != "" {
		if rel, err := filepath.Rel(t.root, path); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return path
}

func (t *tracer) fail(format string, args ...any) (string, []string, bool) {
	t.add("✗ "+format, args...)
	return "", t.lines, false
}

// The hover for a position in a document, as markdown, and the lines saying how it was reached. The
// line and column are 0-based and the column counts UTF-16 code units, as LSP positions do.
func (w *workspace) hover(text string, line, col int, uri string) (string, []string, bool) {
	path := uriPath(uri)
	trace := &tracer{root: w.root}

	lineText, start, ok := lineAt(text, line)
	if !ok {
		return trace.fail("%s:%d is past the end of the document", trace.rel(path), line+1)
	}
	column := byteOffsetOfUTF16Column(lineText, col)
	trace.add("%s:%d:%d", trace.rel(path), line+1, col+1)

	if match := matchCovering(lineText, column, spellIDPatterns...); match != nil {
		id, _ := strconv.ParseInt(lineText[match[2]:match[3]], 10, 32)
		trace.add("id %d", id)
		s, err := findSpell(int32(id))
		if err != nil {
			return trace.fail("%v", err)
		}
		trace.add("✓ %s", title(s))
		return idMarkdown(s), trace.lines, true
	}

	if filepath.Ext(path) != ".go" {
		return trace.fail("no spell id under the cursor")
	}
	f := w.parse(path, text)
	if f.file == nil {
		return trace.fail("no spell id, family or name under the cursor")
	}
	folder := filepath.Dir(path)
	pkg := filepath.Base(folder)
	at := start + column
	nodes := f.enclosing(f.tok.Pos(at))

	if call := spellConfigAt(nodes, f.tok.Pos(at)); call != nil {
		trace.add("SpellConfig")
		result, err := evalSpellConfig(call, w.declarations(folder, f, at, trace), pkg, trace)
		if err != nil {
			return trace.fail("%v", err)
		}
		trace.add("✓ %s → %d fields", title(result.spell), len(result.rows))
		return result.table().markdown(), trace.lines, true
	}

	if lit := spellModConfigAt(nodes, f.tok.Pos(at)); lit != nil {
		trace.add("SpellModConfig")
		result := evalSpellModConfig(lit, w.declarations(folder, f, at, trace), pkg, trace)
		trace.add("✓ SpellModConfig → %d fields", len(result.rows))
		return result.table().markdown(), trace.lines, true
	}

	if field := familyAt(nodes); field != "" {
		trace.add("family %s/%s", pkg, field)
		family, err := findFamily(ladderFamilies(), field, pkg)
		if err != nil {
			return trace.fail("%v", err)
		}
		if family.err != nil {
			return trace.fail("%v", family.err)
		}
		trace.add("✓ %s, %d ranks", family.key(), family.ladder.Len())
		return familyMarkdown(family), trace.lines, true
	}

	declarations := w.declarations(folder, f, at, trace)
	hover, ok := chainHoverAt(f, nodes, at, declarations, trace.rel(folder), trace)
	if !ok {
		return "", trace.lines, false
	}

	trace.add("expr %s", hover.chain.text(false))
	result, err := evalExpr(ladderFamilies(), hover.chain, pkg)
	if err != nil {
		return trace.fail("%v", err)
	}
	trace.add("✓ %s", resultSummary(result))
	return exprMarkdown(result, hover), trace.lines, true
}

func spellConfigAt(nodes []ast.Node, pos token.Pos) *ast.CallExpr {
	for _, node := range nodes {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			continue
		}
		if name, ok := pkgSelector(call.Fun, "spelldata"); ok && name == "SpellConfig" && pos >= call.Fun.Pos() && pos <= call.Fun.End() {
			return call
		}
	}
	return nil
}

// The SpellModConfig literal whose type name the cursor is on.
func spellModConfigAt(nodes []ast.Node, pos token.Pos) *ast.CompositeLit {
	for _, node := range nodes {
		lit, ok := node.(*ast.CompositeLit)
		if ok && lit.Type != nil && isSpellModConfig(lit) && pos >= lit.Type.Pos() && pos <= lit.Type.End() {
			return lit
		}
	}
	return nil
}

func familyAt(nodes []ast.Node) string {
	for _, node := range nodes {
		if expr, ok := node.(ast.Expr); ok {
			if field, ok := pkgSelector(expr, "spellData"); ok {
				return field
			}
		}
	}
	return ""
}

// The name a `<pkg>.<Name>` selector picks, where expr is one.
func pkgSelector(expr ast.Expr, pkg string) (string, bool) {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	if ident, ok := sel.X.(*ast.Ident); !ok || ident.Name != pkg {
		return "", false
	}
	return sel.Sel.Name, true
}

func resultSummary(result *exprResult) string {
	parts := []string{result.card().heading()}
	if result.readEffect > 0 {
		parts = append(parts, fmt.Sprintf("effect %d", result.readEffect))
	}
	if result.kind == kindValue {
		parts = append(parts, result.value)
	}
	return strings.Join(parts, " → ")
}

// The outermost chain under the cursor answers for the accessor the cursor is on, arguments included,
// and a name on its own answers for what the package binds it to.
func chainHoverAt(f *parsedFile, nodes []ast.Node, at int, declarations map[string]declaration, folder string, trace *tracer) (chainHover, bool) {
	for _, node := range nodes {
		expr, ok := node.(ast.Expr)
		if !ok {
			continue
		}
		c, err := walkChain(expr, f.tok.Offset)
		if err != nil {
			continue
		}
		if row := c.root.byID; row != nil && row.covers(at) {
			trace.add("segment %s", row.text(false))
			return resolved(&chain{root: c.root}, "", declarations, trace)
		}
		for i, seg := range c.segments {
			if seg.covers(at) {
				trace.add("segment %s", seg.text(false))
				return resolved(&chain{root: c.root, segments: c.segments[:i+1]}, "", declarations, trace)
			}
		}
		if c.root.name == "" || len(c.segments) == 0 {
			break
		}
		trace.add("ident %s", c.root.name)
		return resolved(&chain{root: c.root}, "", declarations, trace)
	}

	var ident *ast.Ident
	for _, node := range nodes {
		if name, ok := node.(*ast.Ident); ok {
			ident = name
		}
	}
	if ident == nil {
		trace.add("✗ no spell id, family or name under the cursor")
		return chainHover{}, false
	}
	trace.add("ident %s", ident.Name)
	if _, bound := declarations[ident.Name]; bound {
		return resolved(&chain{root: root{name: ident.Name}}, ident.Name, declarations, trace)
	}
	if err, ok := f.unread[ident]; ok {
		trace.add("✗ %s is bound to no chain the evaluator reads: %v", ident.Name, err)
	} else {
		trace.add("✗ no ladder-shaped declaration of %s in %s", ident.Name, folder)
	}
	return chainHover{}, false
}

func resolved(c *chain, name string, declarations map[string]declaration, trace *tracer) (chainHover, bool) {
	resolved, err := resolveChain(c, declarations, trace)
	if err != nil {
		trace.add("✗ %v", err)
		return chainHover{}, false
	}
	return chainHover{chain: resolved, name: name}, true
}

// The chain with every name the package binds substituted by the chain it stands for, down to a
// ladder: at most maxSubstitutions names deep, and never through a name twice.
func resolveChain(c *chain, declarations map[string]declaration, trace *tracer) (*chain, error) {
	return resolveFrom(c, declarations, maxSubstitutions, map[string]bool{}, trace)
}

func resolveFrom(c *chain, declarations map[string]declaration, depth int, seen map[string]bool, trace *tracer) (*chain, error) {
	if c.rooted() {
		return c, nil
	}

	name := c.root.name
	bound, ok := declarations[name]
	if !ok {
		if isFamilyName(name) {
			return &chain{root: root{family: name}, segments: c.segments}, nil
		}
		return nil, fmt.Errorf("no ladder-shaped declaration of %s in the package", name)
	}
	if seen[name] {
		return nil, fmt.Errorf("%s stands on itself", name)
	}
	if depth <= 0 {
		return nil, fmt.Errorf("%s stands on more than %d names", c.text(false), maxSubstitutions)
	}

	seen[name] = true
	trace.add("  %s = %s  (%s:%d)", name, bound.chain.text(false), trace.rel(bound.file), bound.line)
	head, err := resolveFrom(bound.chain, declarations, depth-1, seen, trace)
	if err != nil {
		return nil, err
	}
	return &chain{root: head.root, segments: append(slices.Clip(head.segments), c.segments...)}, nil
}

// A line of the text, counted from 0, without its line ending, and the byte offset it starts at.
func lineAt(text string, line int) (string, int, bool) {
	if line < 0 {
		return "", 0, false
	}
	start := 0
	for ; line > 0; line-- {
		next := strings.IndexByte(text[start:], '\n')
		if next < 0 {
			return "", 0, false
		}
		start += next + 1
	}
	lineText, _, _ := strings.Cut(text[start:], "\n")
	return strings.TrimSuffix(lineText, "\r"), start, true
}

// The byte offset of an LSP position, clamped to the end of its line and of the text.
func offsetOf(text string, line, col int) int {
	lineText, start, ok := lineAt(text, line)
	if !ok {
		return len(text)
	}
	return start + byteOffsetOfUTF16Column(lineText, col)
}

func matchCovering(lineText string, column int, patterns ...*regexp.Regexp) []int {
	for _, pattern := range patterns {
		for _, match := range pattern.FindAllStringSubmatchIndex(lineText, -1) {
			if column >= match[0] && column <= match[1] {
				return match
			}
		}
	}
	return nil
}

func byteOffsetOfUTF16Column(lineText string, units int) int {
	offset := 0
	for offset < len(lineText) && units > 0 {
		r, size := utf8.DecodeRuneInString(lineText[offset:])
		units -= utf16.RuneLen(r)
		offset += size
	}
	return offset
}

// A Windows URI states the drive after the path's leading slash, file:///c:/x.
func uriPath(uri string) string {
	if u, err := url.Parse(uri); err == nil && u.Scheme == "file" {
		path := u.Path
		if filepath.VolumeName(strings.TrimPrefix(path, "/")) != "" {
			path = strings.TrimPrefix(path, "/")
		}
		return filepath.Clean(filepath.FromSlash(path))
	}
	return filepath.Clean(uri)
}

func pathURI(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	abs = filepath.ToSlash(abs)
	if !strings.HasPrefix(abs, "/") {
		abs = "/" + abs
	}
	return (&url.URL{Scheme: "file", Path: abs}).String()
}
