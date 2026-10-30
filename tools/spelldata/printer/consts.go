package printer

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"sync"
)

// Where the constants a class file passes to an accessor or an option are declared.
var constantSources = []struct {
	path  string
	files []string
}{
	{"github.com/wowsims/forever/sim/core/stats", []string{"sim/core/stats/stats.go"}},
	{"github.com/wowsims/forever/sim/core/dbcenums", []string{"sim/core/dbcenums/*.go"}},
	{"github.com/wowsims/forever/sim/core", []string{"sim/core/flags.go", "sim/core/constants.go",
		"sim/core/base_stats_auto_gen.go", "sim/core/spell_mod.go"}},
}

var (
	loadConstants sync.Once
	constFset     = token.NewFileSet()
	constPackages = map[string]*types.Package{}
	evalPackage   *types.Package
	evalPos       token.Pos
	// Each text once: types.Eval parses into constFset for good, which a long-lived server would grow
	// on every hover. Nil is a text that is not a constant.
	constValues = map[string]constant.Value{}

	procFlagConsts  = map[int64]string{}
	procFlag2Consts = map[int64]string{}
	spellModKinds   = map[int64]string{}
)

// Only the packages above are imported, and only by each other: type-checking proto and stats from
// source takes seconds, and the constants do not use them. The errors that leaves behind are ignored;
// the constants still resolve.
type checkedImports map[string]*types.Package

func (c checkedImports) Import(path string) (*types.Package, error) {
	if pkg, ok := c[path]; ok {
		return pkg, nil
	}
	return nil, fmt.Errorf("%s is not imported", path)
}

func scanConstants() {
	root, err := moduleRoot()
	if err != nil {
		return
	}
	conf := types.Config{Importer: checkedImports(constPackages), Error: func(error) {}}

	for _, source := range constantSources {
		var files []*ast.File
		for _, pattern := range source.files {
			paths, _ := filepath.Glob(filepath.Join(root, pattern))
			for _, path := range paths {
				if strings.HasSuffix(path, "_test.go") {
					continue
				}
				file, err := parser.ParseFile(constFset, path, nil, parser.SkipObjectResolution)
				if err != nil {
					return
				}
				files = append(files, file)
			}
		}
		constPackages[source.path], _ = conf.Check(source.path, constFset, files, nil)
	}

	var imports strings.Builder
	for _, source := range constantSources {
		fmt.Fprintf(&imports, "import %q\n", source.path)
	}
	file, err := parser.ParseFile(constFset, "eval.go", "package eval\n"+imports.String(), 0)
	if err != nil {
		return
	}
	evalPackage, _ = conf.Check("eval", constFset, []*ast.File{file}, nil)
	evalPos = file.Name.Pos()

	scanProcFlags(constPackages["github.com/wowsims/forever/sim/core/dbcenums"])
	scanTypedConsts(constPackages["github.com/wowsims/forever/sim/core"], "SpellModType", spellModKinds)
}

// The constants of one named type a package states, by value.
func scanTypedConsts(pkg *types.Package, typeName string, into map[int64]string) {
	if pkg == nil {
		return
	}
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		c, ok := scope.Lookup(name).(*types.Const)
		if !ok {
			continue
		}
		if named, ok := c.Type().(*types.Named); !ok || named.Obj().Name() != typeName {
			continue
		}
		if value, exact := intValue(c.Val()); exact {
			into[value] = name
		}
	}
}

// The PROC_FLAG_ and PROC_FLAG_2_ constants dbcenums states, by value, each word's in its own map. A
// zero is left out: PROC_FLAG_NONE would claim every unset bit.
func scanProcFlags(pkg *types.Package) {
	if pkg == nil {
		return
	}
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		c, ok := scope.Lookup(name).(*types.Const)
		into := procFlagConsts
		switch {
		case !ok:
			continue
		case strings.HasPrefix(name, "PROC_FLAG_2_"):
			into = procFlag2Consts
		case !strings.HasPrefix(name, "PROC_FLAG_"):
			continue
		}
		if value, exact := intValue(c.Val()); exact && value != 0 {
			into[value] = name
		}
	}
}

func constName(names map[int64]string, value int64) (string, bool) {
	loadConstants.Do(scanConstants)
	name, ok := names[value]
	return name, ok
}

// A variable or a call is refused: this reads, it does not run the package around it. A composite or
// function literal is never a constant, and is refused before its text reaches the cache.
func evalConst(expr ast.Expr) (constant.Value, error) {
	loadConstants.Do(scanConstants)
	if evalPackage == nil {
		return nil, fmt.Errorf("sim/core's constants did not load")
	}
	text := types.ExprString(expr)
	if holdsLiteral(expr) {
		return nil, fmt.Errorf("%s is not a literal or a constant of sim/core, stats or dbcenums", text)
	}
	value, seen := constValues[text]
	if !seen {
		if tv, err := types.Eval(constFset, evalPackage, evalPos, text); err == nil {
			value = tv.Value
		}
		constValues[text] = value
	}
	if value == nil {
		return nil, fmt.Errorf("%s is not a literal or a constant of sim/core, stats or dbcenums", text)
	}
	return value, nil
}

func evalInt(expr ast.Expr) (int64, error) {
	value, err := evalConst(expr)
	if err != nil {
		return 0, err
	}
	n, exact := intValue(value)
	if !exact {
		return 0, fmt.Errorf("%s is not an integer", types.ExprString(expr))
	}
	return n, nil
}

// A constant as an integer, where it is one exactly: 2.0 is 2, and 2.5 is none.
func intValue(v constant.Value) (int64, bool) {
	return constant.Int64Val(constant.ToInt(v))
}

func holdsLiteral(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.CompositeLit, *ast.FuncLit:
			found = true
		}
		return !found
	})
	return found
}
