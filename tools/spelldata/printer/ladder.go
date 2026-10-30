package printer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"sync"

	"github.com/wowsims/forever/sim/core/spelldata"
)

// Where a spell id sits in a class file's ladders: the package, the field name the generator gave the
// family, and the call that reaches this rank of it.
type ladderRef struct {
	pkg   string
	field string
	call  string
}

func (r ladderRef) String() string {
	return fmt.Sprintf("%s spellData.%s.%s", r.pkg, r.field, r.call)
}

// One generated ladder, built the way the class file builds it: the ids its constructor names and the
// Ladder that constructor answers, or the panic it answered with instead.
type ladderFamily struct {
	pkg    string
	field  string
	talent bool
	ids    []int32
	ladder spelldata.Ladder
	err    error
}

func (f *ladderFamily) key() string {
	return f.pkg + "/" + f.field
}

var (
	loadLadders sync.Once
	ladderIndex = map[int32][]ladderRef{}
	familyIndex = map[string]*ladderFamily{}
	// The field names every class states a ladder under.
	familyFields = map[string]bool{}
)

// The ladder calls that reach this id, in package and field order. A file that does not parse states
// no ladder and answers nothing - another session editing a class package must not stop the
// printer.
func ladderRefs(id int32) []string {
	loadLadders.Do(scanLadders)

	refs := ladderIndex[id]
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		out = append(out, ref.String())
	}
	return out
}

// Every ladder the generated class files state, keyed `<package>/<field>`.
func ladderFamilies() map[string]*ladderFamily {
	loadLadders.Do(scanLadders)
	return familyIndex
}

// Every `spelldata.Ranked(...)` and `spelldata.Talent(...)` in the generated class files, read as
// source rather than through an import: tools/spelldata must not depend on a class package.
func scanLadders() {
	root, err := moduleRoot()
	if err != nil {
		return
	}

	files, err := filepath.Glob(filepath.Join(root, "sim", "*", "spell_data_auto_gen.go"))
	if err != nil {
		return
	}
	sort.Strings(files)

	for _, path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			continue
		}
		collectLadders(file)
	}
}

func collectLadders(file *ast.File) {
	pkg := file.Name.Name

	ast.Inspect(file, func(node ast.Node) bool {
		kv, ok := node.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		field, ok := kv.Key.(*ast.Ident)
		if !ok {
			return true
		}
		family := newLadderFamily(pkg, field.Name, kv.Value)
		if family == nil {
			return true
		}
		familyIndex[family.key()] = family
		familyFields[family.field] = true
		for id, call := range family.refs() {
			ladderIndex[id] = append(ladderIndex[id], ladderRef{pkg: pkg, field: field.Name, call: call})
		}
		return true
	})
}

// The family a `spelldata.Ranked(ids...)` or `spelldata.Talent(id, ranks)` states, or nil for any other
// value and for arguments that are not the constants the generator writes.
func newLadderFamily(pkg, field string, value ast.Expr) *ladderFamily {
	call, ok := value.(*ast.CallExpr)
	if !ok {
		return nil
	}
	constructor, ok := pkgSelector(call.Fun, "spelldata")
	if !ok {
		return nil
	}

	args := make([]int32, 0, len(call.Args))
	for _, arg := range call.Args {
		n, err := evalInt(arg)
		if err != nil {
			return nil
		}
		args = append(args, int32(n))
	}

	family := &ladderFamily{pkg: pkg, field: field}
	switch {
	case constructor == "Talent" && len(args) == 2 && args[1] > 0:
		family.talent, family.ids = true, args[:1]
		family.ladder, family.err = buildLadder(func() spelldata.Ladder { return spelldata.Talent(args[0], args[1]) })
	case constructor == "Ranked" && len(args) > 0:
		family.ids = args
		family.ladder, family.err = buildLadder(func() spelldata.Ladder { return spelldata.Ranked(args...) })
	default:
		return nil
	}
	return family
}

// Ranked and Talent panic on a spell the store does not carry and on a curve whose rank count
// disagrees with the talent's, and that panic is answered as an error.
func buildLadder(build func() spelldata.Ladder) (ladder spelldata.Ladder, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return build(), nil
}

// How a class file names each of the ladder's ids. A rank below the top is named by its own id rather
// than by position, which is what a reader holding that id is looking for; a talent's ranks are one
// spell, named by rank number.
func (f *ladderFamily) refs() map[int32]string {
	if f.talent {
		return map[int32]string{f.ids[0]: fmt.Sprintf("Rank(n), n up to %d", f.ladder.Len())}
	}
	out := map[int32]string{}
	for i, id := range f.ids {
		out[id] = fmt.Sprintf("ByID(%d)", id)
		if i == len(f.ids)-1 {
			out[id] = "Highest()"
		}
	}
	return out
}
