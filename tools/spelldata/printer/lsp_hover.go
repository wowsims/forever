package printer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/wowsims/forever/sim/siml/lsp/jsonrpc"
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
	function string

	parameter   *parameterSource
	spellConfig *ast.CallExpr
}

type parameterSource struct {
	function       string
	index          int
	rankIndex      int
	rankParameter  string
	spellParameter string
	chain          *chain
	file           string
	line           int
}

// A Go file as the hover reads it. The AST is partial where the text does not parse, and nil where it
// states no package clause.
type parsedFile struct {
	path  string
	text  string
	file  *ast.File
	tok   *token.File
	decls []declaration
	// The source ladder for function parameters inferred from Ladder.Each callbacks.
	parameterSources []parameterSource
	spellConfigs     []declaration

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

		parameter := f.enclosingEachRankParameter(ident.Pos())

		c, err := walkChainWithParameter(value, f.tok.Offset, parameter)

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

	bindSpellConfig := func(name ast.Expr, call *ast.CallExpr) bool {
		ident, ok := name.(*ast.Ident)
		if !ok || ident.Name == "_" {
			return false
		}

		if name, ok := pkgSelector(call.Fun, "spelldata"); !ok || name != "SpellConfig" {
			return false
		}

		if len(call.Args) < 2 {
			return false
		}

		d := declaration{
			name:        ident.Name,
			file:        path,
			line:        f.tok.Line(ident.Pos()),
			spellConfig: call,
		}

		if body := enclosingBody(f.enclosing(ident.Pos())); body != nil {
			d.from = f.tok.Offset(ident.Pos())
			d.to = f.tok.Offset(body.End())
		}

		f.spellConfigs = append(f.spellConfigs, d)
		return true
	}

	bindEach := func(call *ast.CallExpr) {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Each" || len(call.Args) != 1 {
			return
		}

		// Ladder.Each can receive either:
		//
		//     ladder.Each(func(_ int32, rank *spelldata.Spell) {
		//         registerSpell(rank)
		//     })
		//
		// or a method/function value:
		//
		//     ladder.Each(paladin.registerSpell)
		//
		// Parentheses around a function literal are harmless, so unwrap them first.
		callback := call.Args[0]
		for {
			paren, ok := callback.(*ast.ParenExpr)
			if !ok {
				break
			}
			callback = paren.X
		}

		ladder, err := walkChain(sel.X, f.tok.Offset)
		if err != nil {
			return
		}

		// Inline callback:
		//
		//     ladder.Each(func(_ int32, rank *spelldata.Spell) { ... })
		//
		// Find calls inside the callback which pass the callback spell parameter
		// as an argument.
		if fn, ok := callback.(*ast.FuncLit); ok {
			if fn.Type == nil || fn.Type.Params == nil {
				return
			}

			// Ladder.Each has the shape:
			//
			//     func(rank int32, spell *spelldata.Spell)
			//
			// We only care about the second parameter. The first is the numeric rank.
			if fn.Type.Params.List == nil || len(fn.Type.Params.List) < 2 {
				return
			}

			// Ladder.Each has the callback shape:
			//
			//	func(rank int32, spell *spelldata.Spell)
			//
			// The parameter names are arbitrary; parameter position is the API contract.
			rankParam := fn.Type.Params.List[0]
			if types.ExprString(rankParam.Type) != "int32" {
				return
			}

			rankName := ""
			if len(rankParam.Names) > 0 && rankParam.Names[0].Name != "_" {
				rankName = rankParam.Names[0].Name
			}

			spellParam := fn.Type.Params.List[1]
			if len(spellParam.Names) == 0 {
				return
			}

			// We deliberately require the second callback parameter to be
			// *spelldata.Spell.
			if types.ExprString(spellParam.Type) != "*spelldata.Spell" {
				return
			}

			spellName := spellParam.Names[0]
			if spellName.Name == "_" {
				return
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				nested, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				for index, arg := range nested.Args {
					ident, ok := arg.(*ast.Ident)
					if !ok || ident.Name != spellName.Name {
						continue
					}
					functionName := calledFunctionName(nested)
					if functionName == "" {
						continue
					}

					f.parameterSources = append(f.parameterSources, parameterSource{
						function:       functionName,
						index:          index,
						rankParameter:  rankName,
						spellParameter: spellName.Name,
						chain:          ladder,
						file:           path,
						line:           f.tok.Line(spellName.Pos()),
					})

				}

				return true
			})

			return
		}

		// Method/function value callback:
		//
		//	ladder.Each(paladin.registerConsecration)
		//
		// The callback has no explicit argument identifiers at the call site.
		// Instead, record the callback function and let parameterDeclaration
		// resolve its *spelldata.Spell parameter later. The callback may be
		// declared in a different file in the same package.
		callbackSelector, ok := callback.(*ast.SelectorExpr)
		if !ok {
			return
		}

		functionName := callbackSelector.Sel.Name
		if functionName == "" {
			return
		}

		f.parameterSources = append(f.parameterSources, parameterSource{
			function: functionName,
			index:    -1,
			chain:    ladder,
			file:     path,
			line:     f.tok.Line(callbackSelector.Sel.Pos()),
		})

	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {

		case *ast.ValueSpec:
			if len(n.Names) == len(n.Values) {
				for i, name := range n.Names {
					if call, ok := n.Values[i].(*ast.CallExpr); ok &&
						bindSpellConfig(name, call) {
						continue
					}
					bind(name, n.Values[i])
				}
			}

		case *ast.AssignStmt:
			if (n.Tok == token.DEFINE || n.Tok == token.ASSIGN) && len(n.Lhs) == len(n.Rhs) {
				for i, name := range n.Lhs {
					if call, ok := n.Rhs[i].(*ast.CallExpr); ok &&
						bindSpellConfig(name, call) {
						continue
					}
					bind(name, n.Rhs[i])
				}
			}

		case *ast.CallExpr:
			bindEach(n)
		}

		return true
	})
	return f
}

func calledFunctionName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		return fn.Sel.Name
	default:
		return ""
	}
}

func enclosingBody(nodes []ast.Node) *ast.BlockStmt {
	for _, node := range slices.Backward(nodes) {
		switch fn := node.(type) {
		case *ast.FuncDecl:
			return fn.Body
		case *ast.FuncLit:
			return fn.Body
		}
	}
	return nil
}

func (f *parsedFile) enclosingEachRankParameter(pos token.Pos) string {
	nodes := f.enclosing(pos)

	for _, node := range nodes {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			continue
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Each" || len(call.Args) != 1 {
			continue
		}

		callback, ok := call.Args[0].(*ast.FuncLit)
		if !ok || callback.Type == nil || callback.Type.Params == nil {
			continue
		}

		params := callback.Type.Params.List
		if len(params) < 2 {
			continue
		}

		rankParam := params[0]
		spellParam := params[1]

		// Ladder.Each's callback contract is:
		//
		//     func(rank int32, spell *spelldata.Spell)
		//
		// The names are arbitrary; the parameter positions and types
		// define their meaning.
		if types.ExprString(rankParam.Type) != "int32" ||
			types.ExprString(spellParam.Type) != "*spelldata.Spell" {
			continue
		}

		if len(rankParam.Names) == 0 || rankParam.Names[0].Name == "_" {
			return ""
		}

		return rankParam.Names[0].Name
	}

	return ""
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
	key              any
	decls            []declaration
	parameterSources []parameterSource
	spellConfigs     []declaration
	text             string
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

	// Infer function parameters from SpellData ladder .Each callbacks.
	if own != "" {
		// Find the function containing the hover position. Parameter names such
		// as "rank" are local to that function, so package-wide callback sources
		// must not overwrite one another.
		currentFunction := ""
		ast.Inspect(current.file, func(node ast.Node) bool {
			fn, ok := node.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Body == nil {
				return true
			}

			if current.tok.Offset(fn.Pos()) <= at && at <= current.tok.Offset(fn.End()) {
				currentFunction = fn.Name.Name
				return false
			}

			return true
		})

		for _, path := range paths {
			for _, source := range files[path].parameterSources {
				if source.function == "" || source.function != currentFunction {
					continue
				}

				// Inline callbacks have their parameter declaration in the
				// current file.
				if source.index >= 0 {
					d, ok := parameterDeclaration(current, source)
					if !ok || d.from > at || at > d.to {
						continue
					}

					resolved, err := resolveChain(d.chain, out, trace)
					if err != nil {
						trace.add("  %s parameter source unresolved: %v", d.name, err)
						continue
					}

					d.chain = resolved
					out[d.name] = d

					trace.add(
						"  %s = %s  (%s:%d)",
						d.name,
						d.chain.text(false),
						trace.rel(d.file),
						d.line,
					)
					continue
				}

				// Method-value callbacks may be declared in any file in the
				// same package. Search for the callback with the same function
				// name as the function containing the hover.
				for _, callbackPath := range paths {
					text, err := w.text(callbackPath)
					if err != nil {
						continue
					}

					callbackFile := current
					if callbackPath != current.path {
						callbackFile = parseGo(callbackPath, text)
					}

					d, ok := parameterDeclaration(callbackFile, source)
					if !ok {
						continue
					}

					// For a callback in another file, d.from/d.to are offsets
					// in that other file and cannot be compared with at.
					// The function-name match above provides the scope check.
					if callbackPath == current.path && (d.from > at || at > d.to) {
						continue
					}

					resolved, err := resolveChain(d.chain, out, trace)
					if err != nil {
						trace.add("  %s parameter source unresolved: %v", d.name, err)
						continue
					}

					d.chain = resolved
					out[d.name] = d

					trace.add(
						"  %s = %s  (%s:%d)",
						d.name,
						d.chain.text(false),
						trace.rel(d.file),
						d.line,
					)

					break
				}
			}
		}
	}

	return out
}

func parameterDeclaration(file *parsedFile, source parameterSource) (declaration, bool) {
	var result declaration

	ast.Inspect(file.file, func(node ast.Node) bool {
		fn, ok := node.(*ast.FuncDecl)
		if !ok || fn.Name.Name != source.function || fn.Type == nil || fn.Type.Params == nil {
			return true
		}

		params := fn.Type.Params.List
		index := 0

		var rankParameter string

		if source.index < 0 &&
			len(params) > 0 &&
			len(params[0].Names) > 0 &&
			params[0].Names[0].Name != "_" &&
			types.ExprString(params[0].Type) == "int32" {
			rankParameter = params[0].Names[0].Name
		}

		for _, field := range params {
			names := field.Names

			// A method receiver isn't part of Type.Params, so this index
			// exactly matches the Ladder.Each callback argument position.
			//
			// For method-value callbacks we don't know the index when
			// bindEach sees the callback, so find the *spelldata.Spell
			// parameter directly.

			if source.index < 0 && types.ExprString(field.Type) == "*spelldata.Spell" {
				if len(names) == 0 {
					return false
				}

				name := names[0]
				source.rankParameter = rankParameter
				result = declaration{
					name:      name.Name,
					chain:     source.chain,
					file:      file.path,
					line:      file.tok.Line(name.Pos()),
					from:      file.tok.Offset(fn.Pos()),
					to:        file.tok.Offset(fn.End()),
					function:  fn.Name.Name,
					parameter: &source,
				}
				return false
			}

			for _, name := range names {
				if index == source.index {
					source.rankParameter = rankParameter
					result = declaration{
						name:      name.Name,
						chain:     source.chain,
						file:      file.path,
						line:      file.tok.Line(name.Pos()),
						from:      file.tok.Offset(fn.Pos()),
						to:        file.tok.Offset(fn.End()),
						function:  fn.Name.Name,
						parameter: &source,
					}
					return false
				}

				index++
			}

			// An unnamed parameter still occupies one call argument.
			if len(names) == 0 {
				index++
			}
		}

		return false
	})

	return result, result.name != ""
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
		parsed := parseGo(path, text)
		files[path] = cachedFile{
			key:              key,
			decls:            file.decls,
			spellConfigs:     parsed.spellConfigs,
			parameterSources: parsed.parameterSources,
			text:             text,
		}
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
	chain       *chain
	name        string
	parameter   *parameterSource
	spellConfig *ast.CallExpr
	rank        *virtualRank
}

type virtualRank struct {
	uri           string
	parameter     string
	rankParameter string
	function      string
	ladder        *chain
	rank          int32
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

// The ranks the editor picked for Ladder.Each callback parameters, and how a hover offers another.
type rankPicks struct {
	ranks map[string]int32 // By rankKey; a parameter with no pick reads the highest rank.
	links bool             // The editor runs command:wowsims-spelldata.changeRank links in a hover.
}

func (p *rankPicks) rank(key string) int32 {
	if p == nil {
		return 0
	}
	return p.ranks[key]
}

func (p *rankPicks) linked() bool {
	return p != nil && p.links
}

// The hover for a position in a document, as markdown, and the lines saying how it was reached. The
// line and column are 0-based and the column counts UTF-16 code units, as LSP positions do.
func (w *workspace) hover(text string, line, col int, uri string, picks *rankPicks) (string, []string, bool) {
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
		result, err := evalSpellConfig(call, w.declarations(folder, f, at, trace), pkg, trace, nil)
		if err != nil {
			return trace.fail("%v", err)
		}
		trace.add("✓ %s → %d fields", title(result.spell), len(result.rows))
		return result.table().markdown(uri), trace.lines, true
	}

	if lit := spellModConfigAt(nodes, f.tok.Pos(at)); lit != nil {
		trace.add("SpellModConfig")
		result := evalSpellModConfig(lit, w.declarations(folder, f, at, trace), pkg, trace)
		trace.add("✓ SpellModConfig → %d fields", len(result.rows))
		return result.table().markdown(uri), trace.lines, true
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

	if hover.rank != nil {
		hover.rank.rank = picks.rank(rankKey(uri, hover.rank.function, hover.rank.parameter))
	}

	if hover.spellConfig != nil {
		trace.add("SpellConfig declaration %s", hover.name)

		result, err := evalSpellConfig(
			hover.spellConfig,
			declarations,
			pkg,
			trace,
			hover.rank,
		)

		if err != nil {
			return trace.fail("%v", err)
		}
		result.assignedLater = configAssignments(f.file, f.tok, hover.spellConfig, hover.name)
		trace.add("✓ %s → %d fields", title(result.spell), len(result.rows))

		return result.table().markdown(uri), trace.lines, true
	}
	trace.add("expr %s", hover.chain.text(false))
	result, err := evalExpr(ladderFamilies(), hover.chain, pkg, hover.rank)
	if err != nil {
		return trace.fail("%v", err)
	}
	trace.add("✓ %s", resultSummary(result))
	return exprMarkdown(result, hover, picks.linked()), trace.lines, true
}

func rankKey(uri, function, parameter string) string {
	return fmt.Sprintf("%s:%s:%s", uri, function, parameter)
}

// The Ladder.Each callback parameter a hover at the position reads through, with the rank it reads
// now, and the ladder it ranges over. It writes no trace: an editor asks on every cursor move.
func (w *workspace) rankAt(text string, line, col int, uri string, picks *rankPicks) (*virtualRank, *ladderFamily, bool) {
	path := uriPath(uri)
	if filepath.Ext(path) != ".go" {
		return nil, nil, false
	}
	lineText, start, ok := lineAt(text, line)
	if !ok {
		return nil, nil, false
	}
	f := w.parse(path, text)
	if f.file == nil {
		return nil, nil, false
	}
	folder := filepath.Dir(path)
	at := start + byteOffsetOfUTF16Column(lineText, col)
	trace := &tracer{root: w.root}
	hover, ok := chainHoverAt(f, f.enclosing(f.tok.Pos(at)), at, w.declarations(folder, f, at, trace), trace.rel(folder), trace)
	if !ok || hover.rank == nil || hover.rank.ladder == nil {
		return nil, nil, false
	}
	family, err := findFamily(ladderFamilies(), hover.rank.ladder.root.family, filepath.Base(folder))
	if err != nil || family.err != nil {
		return nil, nil, false
	}
	hover.rank.rank = picks.rank(rankKey(uri, hover.rank.function, hover.rank.parameter))
	if hover.rank.rank == 0 {
		hover.rank.rank = family.ladder.Len()
	}
	return hover.rank, family, true
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

		// The cursor on an argument chain is on that chain, which a node further in states.
		if row := c.root.byID; row != nil && row.argumentCovers(at) ||
			slices.ContainsFunc(c.segments, func(seg segment) bool { return seg.argumentCovers(at) }) {
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
		return resolved(&chain{root: c.root}, c.root.name, declarations, trace)
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

	for _, declaration := range f.spellConfigs {
		if declaration.name != ident.Name {
			continue
		}
		if declaration.from != 0 && (at < declaration.from || at > declaration.to) {
			continue
		}

		trace.add("SpellConfig declaration %s", ident.Name)

		hover := chainHover{
			name:        ident.Name,
			spellConfig: declaration.spellConfig,
		}

		if len(declaration.spellConfig.Args) >= 2 {
			rankExpr := declaration.spellConfig.Args[1]

			rankChain, err := walkChain(rankExpr, f.tok.Offset)
			if err == nil {
				rankHover, ok := resolved(rankChain, "", declarations, trace)
				if ok {
					hover.rank = rankHover.rank
				}
			}
		}

		return hover, true
	}

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

func virtualParameterForChain(c *chain, declarations map[string]declaration) (*parameterSource, string, bool) {
	if c == nil {
		return nil, "", false
	}

	for _, seg := range c.segments {
		if seg.name != "Rank" {
			continue
		}

		for _, arg := range seg.args {
			if arg.parameter == "" {
				continue
			}

			for name, declaration := range declarations {
				source := declaration.parameter
				if source == nil {
					continue
				}

				if source.rankParameter != arg.parameter {
					continue
				}

				return source, name, true
			}
		}
	}

	return nil, "", false
}

func resolved(c *chain, name string, declarations map[string]declaration, trace *tracer) (chainHover, bool) {
	return resolvedWithin(c, name, declarations, maxSubstitutions, trace)
}

// resolved, reading argument chains at most depth deep: a name bound through its own argument, which an
// unsaved buffer can state, would otherwise recurse without end.
func resolvedWithin(c *chain, name string, declarations map[string]declaration, depth int, trace *tracer) (chainHover, bool) {
	var parameter *parameterSource
	var parameterName string
	var parameterLadder *chain

	current := c
	seen := map[string]bool{}

	for current.root.name != "" && !seen[current.root.name] {
		declaration, ok := declarations[current.root.name]
		if !ok {
			break
		}
		seen[current.root.name] = true

		if declaration.parameter != nil {
			parameter = declaration.parameter
			parameterName = current.root.name
			parameterLadder = declaration.chain
			break
		}

		// A declaration can refer to the virtual rank indirectly through
		// a deferred Ladder.Each rank argument:
		//
		//	tickSpell := spellData.ConsecrationTriggered.Rank(n)
		//
		// The numeric value of n is intentionally not resolved here.
		// The existing virtual-rank state supplies the selected rank later.
		if source, name, ok := virtualParameterForChain(declaration.chain, declarations); ok && parameter == nil {
			parameter = source
			parameterName = name
			parameterLadder = source.chain
		}

		trace.add("  %s = %s  (%s:%d)", current.root.name, declaration.chain.text(false), trace.rel(declaration.file), declaration.line)

		current = &chain{
			root:     declaration.chain.root,
			segments: append(slices.Clip(declaration.chain.segments), current.segments...),
		}
	}

	// An argument chain can read through the callback parameter too:
	//
	//	judgementRank := spellData.SealOfFuryTriggered.ByID(int32(rank.EffectN(3).BaseValue()))
	current, through := resolveArguments(current, declarations, depth, trace)
	if parameter == nil && through.rank != nil {
		parameter = through.parameter
		parameterName = through.rank.parameter
		parameterLadder = through.rank.ladder
	}

	// A Ladder.Each callback parameter is a virtual spell rank.
	// Its declaration points at the ladder itself, so it cannot be
	// resolved as an ordinary chain.
	if parameter != nil {
		return chainHover{
			chain:     current,
			name:      name,
			parameter: parameter,
			rank: &virtualRank{
				parameter:     parameterName,
				rankParameter: parameter.rankParameter,
				function:      parameter.function,
				ladder:        parameterLadder,
				rank:          0,
			},
		}, true
	}

	return chainHover{
		chain: current,
		name:  name,
	}, true
}

// The chain with every name the package binds substituted by the chain it stands for, down to a
// ladder: at most maxSubstitutions names deep, and never through a name twice.
func resolveChain(c *chain, declarations map[string]declaration, trace *tracer) (*chain, error) {
	c, err := resolveFrom(c, declarations, maxSubstitutions, map[string]bool{}, trace)
	if err != nil {
		return nil, err
	}
	c, _ = resolveArguments(c, declarations, maxSubstitutions, trace)
	return c, nil
}

// The chain with each argument chain resolved the way a hovered chain is, on copies: the declaration
// cache shares its chains between hovers. Answers the hover of the first argument that reads through
// a Ladder.Each callback parameter, whose rank the whole chain then reads at. Arguments nested deeper
// than depth stay as written.
func resolveArguments(c *chain, declarations map[string]declaration, depth int, trace *tracer) (*chain, chainHover) {
	var through chainHover
	resolve := func(seg *segment) {
		if depth <= 0 || !slices.ContainsFunc(seg.args, func(arg argument) bool { return arg.chain != nil }) {
			return
		}
		seg.args = slices.Clone(seg.args)
		for i, arg := range seg.args {
			if arg.chain == nil {
				continue
			}
			hover, _ := resolvedWithin(arg.chain, "", declarations, depth-1, trace)
			seg.args[i].chain = hover.chain
			if through.rank == nil && hover.rank != nil {
				through = hover
			}
		}
	}

	out := &chain{root: c.root, segments: slices.Clone(c.segments)}
	if c.root.byID != nil {
		row := *c.root.byID
		resolve(&row)
		out.root.byID = &row
	}
	for i := range out.segments {
		resolve(&out.segments[i])
	}
	return out, through
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
	if path, ok := jsonrpc.URIPath(uri); ok {
		return filepath.Clean(path)
	}
	return filepath.Clean(uri)
}

func pathURI(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return jsonrpc.PathURI(path)
}
