package main

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/doc"
	"go/parser"
	"go/token"
	"go/types"
	"math"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// What -expr answers: a row, an effect of it, or a value read off one.
const (
	kindSpell  = "spell"
	kindEffect = "effect"
	kindValue  = "value"
)

// A root followed by selectors and calls, with where each part sits in the text it was read from.
type chain struct {
	root     root
	segments []segment
}

// Where a chain opens: a ladder, `spellData.<Family>`; a row named by id, `spelldata.MustFind(id)` or
// `spelldata.Find(id)`, the way item and set-bonus spells are; or a name still to be resolved to one
// of those.
type root struct {
	family string
	byID   *segment
	name   string
}

func (r root) text(resolved bool) string {
	switch {
	case r.family != "":
		return "spellData." + r.family
	case r.byID != nil:
		return "spelldata." + r.byID.text(resolved)
	}
	return r.name
}

// One `.Name(args)` of a chain, a bare `.Name` where the caller wrote no call, or an index `[n]` into
// the list the segment before it answered, its one argument the index.
type segment struct {
	name       string
	args       []argument
	call       bool
	index      bool
	start, end int
}

// An argument as the caller wrote it and as the evaluator reads it: `core.CharacterLevel` is passed as
// 60 and prints as 60 in the trail, so the trail is the chain with every name resolved.
//
// A Ladder.Each callback parameter is deferred until evalExpr has the virtual rank context.
// Parameter contains the identifier name when value is nil.
//
// An argument can be a chain of its own, bare or under a numeric conversion:
// `ByID(int32(rank.EffectN(3).BaseValue()))`. evalExpr reads it to its value before the call, and start
// and end say where it sits, so a hover on it answers for the inner chain.
type argument struct {
	source    string
	value     constant.Value
	parameter string

	chain      *chain
	convert    string
	start, end int
}

// The segment as written, or with every argument resolved to its value.
func (s segment) text(resolved bool) string {
	if s.index {
		return "[" + s.args[0].text(resolved) + "]"
	}
	if !s.call {
		return s.name
	}
	parts := make([]string, 0, len(s.args))
	for _, arg := range s.args {
		parts = append(parts, arg.text(resolved))
	}
	return s.name + "(" + strings.Join(parts, ", ") + ")"
}

func (a argument) text(resolved bool) string {
	if resolved && a.value != nil {
		return a.value.String()
	}
	return a.source
}

// The segment as it follows the one before it in a chain.
func (s segment) link(resolved bool) string {
	if s.index {
		return s.text(resolved)
	}
	return "." + s.text(resolved)
}

func (s segment) covers(at int) bool {
	return at >= s.start && at <= s.end
}

// Whether at falls inside an argument that is a chain of its own, which answers for itself.
func (s segment) argumentCovers(at int) bool {
	return slices.ContainsFunc(s.args, func(arg argument) bool {
		return arg.chain != nil && at >= arg.start && at <= arg.end
	})
}

// Whether the chain opens on a ladder or a row rather than on a name.
func (c *chain) rooted() bool {
	return c.root.family != "" || c.root.byID != nil
}

func (c *chain) text(resolved bool) string {
	var out strings.Builder
	out.WriteString(c.root.text(resolved))
	for _, seg := range c.segments {
		out.WriteString(seg.link(resolved))
	}
	return out.String()
}

// A chain read to the end: the row the card states, the effect the chain went through, and the value
// the last accessor answered, with the type it was called on.
type exprResult struct {
	kind  string
	trail string
	owner string

	family     *ladderFamily
	spell      *spelldata.Spell
	readEffect int

	spellConfig *core.SpellConfig

	value     string
	accessors []string
	// What the last accessor answered, which an argument chain hands to the call it sits in.
	answer reflect.Value
}

// The chain a text states, its offsets counted from the start of the text.
func parseChain(text string) (*chain, error) {
	node, err := parser.ParseExpr(text)
	if err != nil {
		return nil, fmt.Errorf("%q is not a chain of accessor calls", text)
	}
	return walkChain(node, func(pos token.Pos) int { return int(pos) - 1 })
}

// The chain an expression states, with offset turning a node's position into the caller's offsets.
// Anything that is not a name followed by selectors, calls and constant indexes - an operator, a
// conversion - is refused whole, and so is an argument that is neither a constant nor a chain.
func walkChain(node ast.Expr, offset func(token.Pos) int) (*chain, error) {
	switch n := node.(type) {
	case *ast.Ident:
		return &chain{root: root{name: n.Name}}, nil

	case *ast.SelectorExpr:
		c, err := walkChain(n.X, offset)
		if err != nil {
			return nil, err
		}
		if c.root.name == "spellData" && len(c.segments) == 0 {
			c.root = root{family: n.Sel.Name}
			return c, nil
		}
		c.segments = append(c.segments, segment{name: n.Sel.Name, start: offset(n.X.End()), end: offset(n.End())})
		return c, nil

	case *ast.CallExpr:
		sel, ok := n.Fun.(*ast.SelectorExpr)
		if !ok {
			return nil, fmt.Errorf("%s is not an accessor call", types.ExprString(n.Fun))
		}
		c, err := walkChain(sel.X, offset)
		if err != nil {
			return nil, err
		}
		seg := segment{name: sel.Sel.Name, call: true, start: offset(sel.X.End()), end: offset(n.End())}

		for _, arg := range n.Args {
			parsed, err := chainArgument(arg, offset)
			if err != nil {
				return nil, err
			}
			seg.args = append(seg.args, parsed)
		}

		if c.root.name == "spelldata" && len(c.segments) == 0 && (seg.name == "MustFind" || seg.name == "Find") && len(seg.args) == 1 {
			c.root = root{byID: &seg}
			return c, nil
		}
		c.segments = append(c.segments, seg)
		return c, nil

	case *ast.IndexExpr:
		c, err := walkChain(n.X, offset)
		if err != nil {
			return nil, err
		}
		return appendIndex(c, n, offset)
	}
	return nil, fmt.Errorf("%s is not a chain of accessor calls: write spellData.<Family> and the accessors on it", types.ExprString(node))
}

func appendIndex(c *chain, n *ast.IndexExpr, offset func(token.Pos) int) (*chain, error) {
	value, err := evalConst(n.Index)
	if err != nil {
		return nil, err
	}
	c.segments = append(c.segments, segment{
		args:  []argument{{source: types.ExprString(n.Index), value: value}},
		index: true,
		start: offset(n.X.End()),
		end:   offset(n.End()),
	})
	return c, nil
}

// A constant, else a chain that opens on a ladder, a row or a name, bare or under a conversion to a
// number type. A lone name is refused with the constant's error: it is no chain to read.
func chainArgument(arg ast.Expr, offset func(token.Pos) int) (argument, error) {
	source := types.ExprString(arg)
	value, err := evalConst(arg)
	if err == nil {
		return argument{source: source, value: value}, nil
	}

	inner, convert := arg, ""
	if call, ok := arg.(*ast.CallExpr); ok && len(call.Args) == 1 {
		if ident, ok := call.Fun.(*ast.Ident); ok && numberType(ident.Name) != nil {
			inner, convert = call.Args[0], ident.Name
		}
	}
	c, chainErr := walkChain(inner, offset)
	if chainErr != nil || (c.root.name != "" && len(c.segments) == 0) {
		return argument{}, err
	}
	return argument{source: source, chain: c, convert: convert, start: offset(arg.Pos()), end: offset(arg.End())}, nil
}

// The basic number type a name stands for in Go's universe scope, or nil.
func numberType(name string) *types.Basic {
	typeName, ok := types.Universe.Lookup(name).(*types.TypeName)
	if !ok {
		return nil
	}
	basic, ok := typeName.Type().(*types.Basic)
	if !ok || basic.Info()&types.IsNumeric == 0 {
		return nil
	}
	return basic
}

func walkChainWithParameter(node ast.Expr, offset func(token.Pos) int, parameter string) (*chain, error) {
	switch n := node.(type) {
	case *ast.CallExpr:
		sel, ok := n.Fun.(*ast.SelectorExpr)
		if !ok {
			return nil, fmt.Errorf("%s is not an accessor call", types.ExprString(n.Fun))
		}

		c, err := walkChainWithParameter(sel.X, offset, parameter)
		if err != nil {
			return nil, err
		}

		seg := segment{
			name:  sel.Sel.Name,
			call:  true,
			start: offset(sel.X.End()),
			end:   offset(n.End()),
		}

		for _, arg := range n.Args {
			source := types.ExprString(arg)

			ident, ok := arg.(*ast.Ident)
			if !ok {
				parsed, err := chainArgument(arg, offset)
				if err != nil {
					return nil, err
				}
				seg.args = append(seg.args, parsed)
				continue
			}

			value, err := evalConst(arg)
			if err == nil {
				seg.args = append(seg.args, argument{
					source: source,
					value:  value,
				})
				continue
			}

			// Rank(...) inside a virtual-rank evaluation uses the rank
			// selected by the hover rather than resolving the identifier
			// as a sim/core constant.
			if seg.name == "Rank" {
				seg.args = append(seg.args, argument{
					source:    source,
					parameter: ident.Name,
				})
				continue
			}

			if parameter == "" || ident.Name != parameter {
				return nil, err
			}

			seg.args = append(seg.args, argument{
				source:    source,
				parameter: ident.Name,
			})
		}

		if c.root.name == "spelldata" && len(c.segments) == 0 &&
			(seg.name == "MustFind" || seg.name == "Find") && len(seg.args) == 1 {
			c.root = root{byID: &seg}
			return c, nil
		}

		c.segments = append(c.segments, seg)
		return c, nil

	case *ast.IndexExpr:
		c, err := walkChainWithParameter(n.X, offset, parameter)
		if err != nil {
			return nil, err
		}
		return appendIndex(c, n, offset)

	default:
		return walkChain(node, offset)
	}
}

// A ladder followed by the store's own accessors, evaluated by name over the exported method set: an
// accessor added to sim/core/spelldata is readable here without a list to keep in step.
func evalExpr(index map[string]*ladderFamily, c *chain, pkg string, virtual *virtualRank) (result *exprResult, err error) {
	// The store panics where a caller asks for a rank or an effect it does not carry, on purpose. The
	// panic is this reader's answer, as an error.
	defer func() {
		if r := recover(); r != nil {
			result, err = nil, fmt.Errorf("%v", r)
		}
	}()

	c, err = readArguments(index, c, pkg, virtual)
	if err != nil {
		return nil, err
	}

	var family *ladderFamily
	var res *exprResult
	var current reflect.Value

	switch {
	case virtual != nil && c.root.name == virtual.parameter:
		family, err = findFamily(index, virtual.ladder.root.family, pkg)
		if err != nil {
			return nil, err
		}
		if family.err != nil {
			return nil, family.err
		}

		spell := family.ladder.Highest()
		// if virtual.rank > 0
		if virtual.rank > 0 {
			spell = family.ladder.Rank(virtual.rank)
		}

		if spell == spelldata.Nil {
			return nil, fmt.Errorf("%s names a spell the store does not carry", c.root.name)
		}

		res = &exprResult{
			trail:  c.root.name,
			family: family,
			spell:  spell,
		}
		current = reflect.ValueOf(spell)

	case c.root.byID != nil:
		id, err := convertArg(c.root.byID.args[0], reflect.TypeFor[int32]())
		if err != nil {
			return nil, err
		}
		row := spelldata.Find(int32(id.Int()))
		if row == spelldata.Nil {
			return nil, fmt.Errorf("%s names a spell the store does not carry", c.root.text(true))
		}
		res = &exprResult{trail: c.root.text(true), spell: row}
		current = reflect.ValueOf(row)
	case c.root.family != "":
		family, err = findFamily(index, c.root.family, pkg)
		if err != nil {
			return nil, err
		}
		if family.err != nil {
			return nil, family.err
		}
		if len(c.segments) > 0 {
			if err := family.checkPick(c.segments[0]); err != nil {
				return nil, err
			}
		}
		res = &exprResult{trail: c.root.text(true), family: family, spell: family.ladder.Highest()}
		current = reflect.ValueOf(family.ladder)
	default:
		return nil, fmt.Errorf("%q is not a ladder call: write spellData.<Family> and the accessors on it", c.text(false))
	}
	var effectOf func(*spelldata.Spell) *spelldata.Effect

	for _, seg := range c.segments {
		answer, err := callSegment(current, seg, virtual)
		if err != nil {
			return nil, err
		}
		res.trail += seg.link(true)
		res.owner = baseTypeName(current.Type())

		switch value := answer.Interface().(type) {
		case *spelldata.Spell:
			if value == spelldata.Nil {
				return nil, fmt.Errorf("%s answers a spell the store does not carry", res.trail)
			}
			res.spell, res.readEffect = value, 0
		case *spelldata.Effect:
			res.readEffect = effectPosition(res.spell, value)
		case spelldata.LadderEffect:
			effectOf = spellEffect(seg)
		default:
			if rank, ok := rankArgument(current, seg, virtual); ok {
				res.spell = family.ladder.Rank(rank)
				res.readEffect = 1
				if effectOf != nil {
					res.readEffect = effectPosition(res.spell, effectOf(res.spell))
				}
			}
		}
		current = answer
	}

	res.answer = current
	switch value := current.Interface().(type) {
	case *spelldata.Spell:
		res.kind, res.value = kindSpell, strconv.Itoa(int(value.ID))
	case *spelldata.Effect:
		res.kind, res.accessors = kindEffect, effectAccessors(value)
		res.value = "no effect"
		if res.readEffect > 0 {
			res.value = fmt.Sprintf("effect %d", res.readEffect)
		}
	case spelldata.Ladder, spelldata.LadderEffect:
		return nil, fmt.Errorf("%s names a ladder, not a rank: read a rank off it with Highest(), Rank(n), ByID(id) or a ...At(rank)", res.trail)
	default:
		text, err := formatValue(current)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", res.trail, err)
		}
		res.kind, res.value = kindValue, text
	}
	return res, nil
}

// The chain with each argument that is a chain of its own read to its value, on copies: the
// declaration cache shares its chains between hovers.
func readArguments(index map[string]*ladderFamily, c *chain, pkg string, virtual *virtualRank) (*chain, error) {
	read := func(seg *segment) error {
		if !slices.ContainsFunc(seg.args, func(arg argument) bool { return arg.chain != nil }) {
			return nil
		}
		seg.args = slices.Clone(seg.args)
		for i, arg := range seg.args {
			if arg.chain == nil {
				continue
			}
			value, err := chainValue(index, arg, pkg, virtual)
			if err != nil {
				return err
			}
			seg.args[i].value = value
		}
		return nil
	}

	out := &chain{root: c.root, segments: slices.Clone(c.segments)}
	if c.root.byID != nil {
		row := *c.root.byID
		if err := read(&row); err != nil {
			return nil, err
		}
		out.root.byID = &row
	}
	for i := range out.segments {
		if err := read(&out.segments[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// The number an argument chain answers, converted the way Go converts a variable: a float to an
// integer type drops its fraction.
func chainValue(index map[string]*ladderFamily, arg argument, pkg string, virtual *virtualRank) (constant.Value, error) {
	result, err := evalExpr(index, arg.chain, pkg, virtual)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", arg.source, err)
	}

	answer := result.answer
	var value constant.Value
	switch {
	case answer.CanInt():
		value = constant.MakeInt64(answer.Int())
	case answer.CanUint():
		value = constant.MakeUint64(answer.Uint())
	case answer.CanFloat():
		value = constant.MakeFloat64(answer.Float())
		if basic := numberType(arg.convert); basic != nil && basic.Info()&types.IsInteger != 0 {
			value = constant.MakeFloat64(math.Trunc(answer.Float()))
		}
	default:
		return nil, fmt.Errorf("%s answers a %s, not a number", arg.source, answer.Type())
	}
	return value, nil
}

// Rank and ByID refused in the ladder's own terms, where the store would answer Nil or panic.
func (f *ladderFamily) checkPick(seg segment) error {
	if len(seg.args) != 1 {
		return nil
	}
	if seg.args[0].value == nil {
		return nil
	}

	n, ok := intValue(seg.args[0].value)
	if !ok {
		return nil
	}
	switch seg.name {
	case "Rank":
		if n < 1 || n > int64(f.ladder.Len()) {
			return fmt.Errorf("%s spellData.%s has %d ranks, not rank %d", f.pkg, f.field, f.ladder.Len(), n)
		}
	case "ByID":
		if int64(int32(n)) != n || !slices.Contains(f.ids, int32(n)) {
			return fmt.Errorf("%s spellData.%s has no rank with id %d", f.pkg, f.field, n)
		}
	}
	return nil
}

// The effect of a rank that a LadderEffect reads, found the way the store finds it.
func spellEffect(seg segment) func(*spelldata.Spell) *spelldata.Effect {
	method := map[string]string{"EffectAt": "EffectN", "Effect": "Effect"}[seg.name]
	return func(s *spelldata.Spell) *spelldata.Effect {
		answer, err := callSegment(
			reflect.ValueOf(s),
			segment{name: method, args: seg.args, call: true},
			nil,
		)
		if err != nil {
			return spelldata.NilEffect
		}
		return answer.Interface().(*spelldata.Effect)
	}
}

// The rank a value read off a ladder is at: the one argument of a ...At(rank). Rank 0 is untaken and
// reads no rank.

func rankArgument(recv reflect.Value, seg segment, virtual *virtualRank) (int32, bool) {
	switch recv.Interface().(type) {
	case spelldata.Ladder, spelldata.LadderEffect:
	default:
		return 0, false
	}

	if len(seg.args) != 1 || (seg.name != "Rank" && !strings.HasSuffix(seg.name, "At")) {
		return 0, false
	}

	// A virtual-rank hover already has the rank selected by the
	// Set Rank / Change Rank markdown control. The argument to
	// Rank(...) or an *At accessor may therefore be the callback's
	// rank parameter rather than a constant.
	if virtual != nil && seg.args[0].parameter != "" {
		return virtual.rank, virtual.rank > 0
	}

	value, err := argumentValue(seg.args[0], virtual)
	if err != nil {
		return 0, false
	}

	n, exact := intValue(value)
	return int32(n), exact && n > 0
}

// Resolve a chain argument to a constant. Ladder.Each callback parameters are
// the one exception: the rank-number parameter is resolved from virtual.rank
// when evaluating a virtual rank.
func argumentValue(arg argument, virtual *virtualRank) (constant.Value, error) {
	if arg.value != nil {
		return arg.value, nil
	}

	if virtual != nil && arg.parameter != "" && arg.parameter == virtual.rankParameter {
		return constant.MakeInt64(int64(virtual.rank)), nil
	}

	return nil, fmt.Errorf("%s is not a literal or a constant of sim/core, stats or dbcenums", arg.source)
}

func callSegment(recv reflect.Value, seg segment, virtual *virtualRank) (reflect.Value, error) {
	owner := recv.Type().String()
	if seg.index {
		list := reflect.Indirect(recv)
		if list.Kind() != reflect.Slice && list.Kind() != reflect.Array {
			return reflect.Value{}, fmt.Errorf("a %s is not a list to index", owner)
		}
		value, err := argumentValue(seg.args[0], virtual)
		if err != nil {
			return reflect.Value{}, err
		}
		if i, exact := intValue(value); exact && i >= 0 && i < int64(list.Len()) {
			return list.Index(int(i)), nil
		}
		return reflect.Value{}, fmt.Errorf("the list has no index %s: it holds %d", seg.args[0].source, list.Len())
	}
	if !seg.call {
		row := reflect.Indirect(recv)
		if row.Kind() == reflect.Struct {
			if field, ok := row.Type().FieldByName(seg.name); ok && field.IsExported() {
				return row.FieldByIndex(field.Index), nil
			}
		}
		return reflect.Value{}, fmt.Errorf("%q is not a field of %s", seg.name, owner)
	}

	method := recv.MethodByName(seg.name)
	if !method.IsValid() {
		return reflect.Value{}, fmt.Errorf("%q is not a method of %s", seg.name, owner)
	}
	signature := method.Type()
	label := fmt.Sprintf("%s.%s", owner, seg.name)

	if signature.IsVariadic() {
		return reflect.Value{}, fmt.Errorf("%s takes a list of arguments, which -expr does not write", label)
	}
	if signature.NumIn() != len(seg.args) {
		return reflect.Value{}, fmt.Errorf("%s takes %s, not %d", label, arguments(signature.NumIn()), len(seg.args))
	}
	if signature.NumOut() != 1 {
		return reflect.Value{}, fmt.Errorf("%s answers %d values, and -expr reads one", label, signature.NumOut())
	}

	in := make([]reflect.Value, 0, len(seg.args))
	for i, arg := range seg.args {
		resolved, err := argumentValue(arg, virtual)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%s: %w", label, err)
		}

		value, err := convertArg(
			argument{
				source: arg.source,
				value:  resolved,
			},
			signature.In(i),
		)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%s: %w", label, err)
		}
		in = append(in, value)
	}
	return method.Call(in)[0], nil
}

// A constant as the parameter's own type. An integer widens into a float, a fractional number does not
// narrow into an integer, and an integer the type cannot hold does not wrap: truncating it silently is
// the reading a caller would not notice.
func convertArg(arg argument, want reflect.Type) (reflect.Value, error) {
	zero := reflect.Zero(want)
	switch {
	case zero.CanUint():
		if n, exact := constant.Uint64Val(constant.ToInt(arg.value)); exact && !zero.OverflowUint(n) {
			return reflect.ValueOf(n).Convert(want), nil
		}
	case zero.CanInt():
		if n, exact := intValue(arg.value); exact && !zero.OverflowInt(n) {
			return reflect.ValueOf(n).Convert(want), nil
		}
	case zero.CanFloat():
		if v := constant.ToFloat(arg.value); v.Kind() == constant.Float {
			f, _ := constant.Float64Val(v)
			return reflect.ValueOf(f).Convert(want), nil
		}
	case want.Kind() == reflect.Bool:
		if arg.value.Kind() == constant.Bool {
			return reflect.ValueOf(constant.BoolVal(arg.value)).Convert(want), nil
		}
	case want.Kind() == reflect.String:
		if arg.value.Kind() == constant.String {
			return reflect.ValueOf(constant.StringVal(arg.value)).Convert(want), nil
		}
	}
	return reflect.Value{}, fmt.Errorf("%s is not the %s it takes", arg.source, want)
}

func isInteger(t reflect.Type) bool {
	zero := reflect.Zero(t)
	return zero.CanInt() || zero.CanUint()
}

func arguments(n int) string {
	if n == 1 {
		return "1 argument"
	}
	return fmt.Sprintf("%d arguments", n)
}

// The value the last accessor answered, in the units the text form states elsewhere: a time as the
// duration it is, a number read back as the float32 the game files' columns are.
func formatValue(v reflect.Value) (string, error) {
	switch value := v.Interface().(type) {
	case time.Duration:
		return value.String(), nil
	case core.SpellSchool:
		return value.String(), nil
	case bool:
		return strconv.FormatBool(value), nil
	case string:
		return value, nil
	case []*spelldata.Spell:
		if len(value) == 0 {
			return "none", nil
		}
		titles := make([]string, len(value))
		for i, s := range value {
			titles[i] = fmt.Sprintf("[%d] %s", i, title(s))
		}
		return strings.Join(titles, ", "), nil
	}

	switch {
	case v.CanInt():
		return strconv.FormatInt(v.Int(), 10), nil
	case v.CanUint():
		return strconv.FormatUint(v.Uint(), 10), nil
	case v.CanFloat():
		return number(v.Float()), nil
	}
	return "", fmt.Errorf("a %s is not a value to read", v.Type())
}

func (r *exprResult) card() card {
	return newCard(r.spell, rankLabel(r.family, r.spell), r.readEffect)
}

// Where an effect sits in the row the chain came through, counted the way EffectN counts. 0 for an
// effect the row does not carry, which is what a finder answers where the row has none.
func effectPosition(s *spelldata.Spell, e *spelldata.Effect) int {
	for i := range s.Effects {
		if &s.Effects[i] == e {
			return i + 1
		}
	}
	return 0
}

// Every Effect accessor that reads something off this effect, with what it answers there: the
// zero-argument ones and the ones that take a caster level. One that answers nothing - a zero amount,
// no period - is left out, so the list is the readings this effect actually states.
func effectAccessors(e *spelldata.Effect) []string {
	v := reflect.ValueOf(e)

	out := []string{}
	for i := range v.NumMethod() {
		name := v.Type().Method(i).Name
		method := v.Method(i)
		signature := method.Type()
		if signature.IsVariadic() || signature.NumOut() != 1 {
			continue
		}

		var in []reflect.Value
		label := name + "()"
		switch {
		case signature.NumIn() == 0:
		case signature.NumIn() == 1 && isInteger(signature.In(0)):
			in = []reflect.Value{reflect.ValueOf(int64(core.CharacterLevel)).Convert(signature.In(0))}
			label = fmt.Sprintf("%s(%d)", name, core.CharacterLevel)
		default:
			continue
		}

		value, err := formatValue(method.Call(in)[0])
		if err != nil || value == "" || value == "0" || value == "0s" || value == "false" {
			continue
		}
		out = append(out, label+" = "+value)
	}
	return out
}

var (
	loadDocs   sync.Once
	methodDocs = map[string]string{}
)

// The doc comment sim/core/spelldata writes on an accessor, keyed `<Type>.<Method>`, so the reader
// states what the store itself says about the value. Empty for an accessor with no comment.
func methodDoc(owner, method string) string {
	loadDocs.Do(scanMethodDocs)
	return methodDocs[owner+"."+method]
}

func scanMethodDocs() {
	root, err := moduleRoot()
	if err != nil {
		return
	}
	paths, err := filepath.Glob(filepath.Join(root, "sim", "core", "spelldata", "*.go"))
	if err != nil {
		return
	}

	fset := token.NewFileSet()
	var files []*ast.File
	for _, path := range paths {
		name := filepath.Base(path)
		// The generated table is one long literal and states no accessor, and a test file's helpers are
		// not what a caller reads.
		if name == "spells_auto_gen.go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	pkg, err := doc.NewFromFiles(fset, files, "github.com/wowsims/forever/sim/core/spelldata")
	if err != nil {
		return
	}
	for _, t := range pkg.Types {
		for _, m := range t.Methods {
			methodDocs[t.Name+"."+m.Name] = strings.TrimSpace(m.Doc)
		}
	}
}

// The type name the doc comments are keyed by: `*spelldata.Effect` is written on `Effect`.
func baseTypeName(t reflect.Type) string {
	name := strings.TrimPrefix(t.String(), "*")
	if _, after, ok := strings.Cut(name, "."); ok {
		return after
	}
	return name
}
