package printer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

type configOption struct {
	label  string
	source string
	opt    spelldata.SpellOpt
	err    error
}

// The options a SpellConfig call can pass, called with the arguments converted to what each takes.
var optionBuilders = map[string]any{
	"Melee": spelldata.Melee,
	"Magic": spelldata.Magic,
	"Proc":  spelldata.Proc,
	"Flags": spelldata.Flags,
	"Tag":   spelldata.Tag,
}

func readOption(expr ast.Expr) configOption {
	source := types.ExprString(expr)
	option := configOption{label: source, source: source}

	call, ok := expr.(*ast.CallExpr)
	var name string
	if ok {
		name, ok = pkgSelector(call.Fun, "spelldata")
	}
	if !ok {
		option.err = fmt.Errorf("not a spelldata option call")
		return option
	}

	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, strings.ReplaceAll(types.ExprString(arg), "core.", ""))
	}
	option.label = name + "(" + strings.Join(args, " | ") + ")"
	if len(option.label) > 48 {
		option.label = name + "(…)"
	}

	builder, known := optionBuilders[name]
	if !known {
		option.err = fmt.Errorf("spelldata.%s is not an option this reads", name)
		return option
	}
	build := reflect.ValueOf(builder)
	if len(call.Args) != build.Type().NumIn() {
		option.err = fmt.Errorf("spelldata.%s takes %s", name, arguments(build.Type().NumIn()))
		return option
	}

	in := make([]reflect.Value, 0, len(call.Args))
	for i, arg := range call.Args {
		value, err := evalConst(arg)
		if err != nil {
			option.err = err
			return option
		}
		converted, err := convertArg(argument{source: types.ExprString(arg), value: value}, build.Type().In(i))
		if err != nil {
			option.err = err
			return option
		}
		in = append(in, converted)
	}
	option.opt = build.Call(in)[0].Interface().(spelldata.SpellOpt)
	return option
}

type configRow struct {
	field string
	value string
	from  string
}

type configResult struct {
	spell         *spelldata.Spell
	pick          string
	rows          []configRow
	skipped       []configOption
	missing       []string
	assignedLater []configAssignment
}

func evalSpellConfig(expr *ast.CallExpr, declarations map[string]declaration, pkg string, trace *tracer, virtual *virtualRank) (*configResult, error) {
	if len(expr.Args) < 2 {
		return nil, fmt.Errorf("%s is not a SpellConfig call with a row", types.ExprString(expr))
	}

	s, pick, err := configPick(expr.Args[1], declarations, pkg, trace, virtual)
	if err != nil {
		return nil, err
	}

	result := &configResult{spell: s, pick: pick}
	labels := []string{fmt.Sprintf("SpellData · %s", title(s))}
	var opts []spelldata.SpellOpt
	for _, arg := range expr.Args[2:] {
		option := readOption(arg)
		if option.err != nil {
			trace.add("  option %s unevaluated: %v", option.source, option.err)
			result.skipped = append(result.skipped, option)
			continue
		}
		trace.add("  option %s", option.label)
		opts = append(opts, option.opt)
		labels = append(labels, option.label)
	}

	stages := make([]core.SpellConfig, 0, len(opts)+1)
	for i := 0; i <= len(opts); i++ {
		stages = append(stages, spelldata.SpellConfig(&core.Unit{}, s, opts[:i]...))
	}
	result.rows = attribute(stages, labels)
	result.missing = missingConfigFields(&stages[len(stages)-1])
	return result, nil
}

func configPick(arg ast.Expr, declarations map[string]declaration, pkg string, trace *tracer, virtual *virtualRank) (*spelldata.Spell, string, error) {
	trace.add("  row %s", types.ExprString(arg))
	_, result, err := evalDeclaredExpr(arg, declarations, pkg, trace, virtual)
	if err != nil {
		return nil, "", err
	}
	if result.kind != kindSpell {
		return nil, "", fmt.Errorf("%s reads a %s, not a row", types.ExprString(arg), result.kind)
	}
	return result.spell, title(result.spell), nil
}

// A chain of the store's accessors, resolved through the package's declarations and evaluated.
func evalDeclaredExpr(expr ast.Expr, declarations map[string]declaration, pkg string, trace *tracer, virtual *virtualRank) (*chain, *exprResult, error) {
	c, err := walkChain(expr, func(token.Pos) int { return 0 })
	if err != nil {
		return nil, nil, err
	}

	if virtual != nil {
		// Preserve the virtual parameter as the root of the expression.
		// This matters not only for "rank" itself, but also for aliases such
		// as "holyFireHit" whose declaration eventually resolves through rank.
		resolved, err := resolveToVirtual(c, declarations, virtual.parameter, trace)
		if err != nil {
			return nil, nil, err
		}
		c, _ = resolveArguments(resolved, declarations, maxSubstitutions, trace)
	} else {
		c, err = resolveChain(c, declarations, trace)
		if err != nil {
			return nil, nil, err
		}
	}

	result, err := evalExpr(ladderFamilies(), c, pkg, virtual)
	if err != nil {
		return nil, nil, err
	}

	return c, result, nil
}

func resolveToVirtual(c *chain, declarations map[string]declaration, parameter string, trace *tracer) (*chain, error) {
	if c.root.name == parameter {
		return c, nil
	}

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

	trace.add("  %s = %s  (%s:%d)", name, bound.chain.text(false), trace.rel(bound.file), bound.line)

	head, err := resolveToVirtual(bound.chain, declarations, parameter, trace)
	if err != nil {
		return nil, err
	}

	return &chain{
		root:     head.root,
		segments: append(slices.Clip(head.segments), c.segments...),
	}, nil
}

type configField struct {
	name    string
	read    func(*core.SpellConfig) string
	bits    func(*core.SpellConfig) uint64
	bitName func(bit uint64) (string, bool)
}

func nonZero[T comparable](v T, format func(T) string) string {
	var zero T
	if v == zero {
		return ""
	}
	return format(v)
}

func floatText(v float64) string          { return strconv.FormatFloat(v, 'f', -1, 64) }
func durationText(v time.Duration) string { return v.String() }
func intText[T int32 | int](v T) string   { return strconv.Itoa(int(v)) }
func boolText(bool) string                { return "true" }

func procMaskBit(bit uint64) (string, bool)  { return dbcenums.Named(core.ProcMask(bit)) }
func spellFlagBit(bit uint64) (string, bool) { return dbcenums.Named(core.SpellFlag(bit)) }

var configFields = []configField{
	{name: "ActionID", read: func(c *core.SpellConfig) string {
		if c.ActionID.Tag != 0 {
			return fmt.Sprintf("SpellID %d, Tag %d", c.ActionID.SpellID, c.ActionID.Tag)
		}
		return nonZero(c.ActionID.SpellID, func(id int32) string { return fmt.Sprintf("SpellID %d", id) })
	}},
	{name: "Rank", read: func(c *core.SpellConfig) string { return nonZero(c.Rank, intText[int32]) }},
	{name: "SpellSchool", read: func(c *core.SpellConfig) string { return nonZero(c.SpellSchool, core.SpellSchool.String) }},
	{name: "DefenseType", read: func(c *core.SpellConfig) string { return nonZero(c.DefenseType, core.DefenseType.String) }},
	{name: "Flags", bits: func(c *core.SpellConfig) uint64 { return uint64(c.Flags) }, bitName: spellFlagBit},
	{name: "ProcMask", bits: func(c *core.SpellConfig) uint64 { return uint64(c.ProcMask) }, bitName: procMaskBit},
	{name: "Cast.DefaultCast.CastTime", read: func(c *core.SpellConfig) string { return nonZero(c.Cast.DefaultCast.CastTime, durationText) }},
	{name: "Cast.DefaultCast.GCD", read: func(c *core.SpellConfig) string { return nonZero(c.Cast.DefaultCast.GCD, durationText) }},
	{name: "Cast.DefaultCast.NonEmpty", read: func(c *core.SpellConfig) string { return nonZero(c.Cast.DefaultCast.NonEmpty, boolText) }},
	{name: "Cast.IgnoreHaste", read: func(c *core.SpellConfig) string { return nonZero(c.Cast.IgnoreHaste, boolText) }},
	{name: "Cast.CD", read: func(c *core.SpellConfig) string { return nonZero(c.Cast.CD.Duration, durationText) }},
	{name: "Cast.SharedCD", read: func(c *core.SpellConfig) string { return nonZero(c.Cast.SharedCD.Duration, durationText) }},
	{name: "ManaCost.FlatCost", read: func(c *core.SpellConfig) string { return nonZero(c.ManaCost.FlatCost, intText[int32]) }},
	{name: "ManaCost.BaseCostPercent", read: func(c *core.SpellConfig) string { return nonZero(c.ManaCost.BaseCostPercent, floatText) }},
	{name: "RageCost.Cost", read: func(c *core.SpellConfig) string { return nonZero(c.RageCost.Cost, intText[int32]) }},
	{name: "RageCost.Refund", read: func(c *core.SpellConfig) string { return nonZero(c.RageCost.Refund, floatText) }},
	{name: "EnergyCost.Cost", read: func(c *core.SpellConfig) string { return nonZero(c.EnergyCost.Cost, intText[int32]) }},
	{name: "EnergyCost.Refund", read: func(c *core.SpellConfig) string { return nonZero(c.EnergyCost.Refund, floatText) }},
	{name: "FocusCost.Cost", read: func(c *core.SpellConfig) string { return nonZero(c.FocusCost.Cost, intText[int32]) }},
	{name: "FocusCost.Refund", read: func(c *core.SpellConfig) string { return nonZero(c.FocusCost.Refund, floatText) }},
	{name: "DamageMultiplier", read: func(c *core.SpellConfig) string { return nonZero(c.DamageMultiplier, floatText) }},
	{name: "ThreatMultiplier", read: func(c *core.SpellConfig) string { return nonZero(c.ThreatMultiplier, floatText) }},
	{name: "FlatThreatBonus", read: func(c *core.SpellConfig) string { return nonZero(c.FlatThreatBonus, floatText) }},
	{name: "ThreatAPCoefficient", read: func(c *core.SpellConfig) string { return nonZero(c.ThreatAPCoefficient, floatText) }},
	{name: "BonusCoefficient", read: func(c *core.SpellConfig) string { return nonZero(c.BonusCoefficient, floatText) }},
	{name: "MinRange", read: func(c *core.SpellConfig) string { return nonZero(c.MinRange, floatText) }},
	{name: "MaxRange", read: func(c *core.SpellConfig) string { return nonZero(c.MaxRange, floatText) }},
	{name: "MissileSpeed", read: func(c *core.SpellConfig) string { return nonZero(c.MissileSpeed, floatText) }},
	{name: "ClassFlags", read: func(c *core.SpellConfig) string { return classFlagsPhrase(c.ClassFlags) }},
	{name: "CastRequirement.Forms", read: func(c *core.SpellConfig) string { return nonZero(c.CastRequirement.Forms, formsText) }},
	{name: "CastRequirement.ExcludedForms", read: func(c *core.SpellConfig) string { return nonZero(c.CastRequirement.ExcludedForms, formsText) }},
	{name: "CastRequirement.CasterForm", read: func(c *core.SpellConfig) string { return nonZero(c.CastRequirement.CasterForm, boolText) }},
	{name: "CastRequirement.NotShapeshifted", read: func(c *core.SpellConfig) string { return nonZero(c.CastRequirement.NotShapeshifted, boolText) }},
	{name: "CastRequirement.CasterAura", read: func(c *core.SpellConfig) string { return nonZero(c.CastRequirement.CasterAura, intText[int32]) }},
	{name: "CastRequirement.ExcludeCasterAura", read: func(c *core.SpellConfig) string { return nonZero(c.CastRequirement.ExcludeCasterAura, intText[int32]) }},
	{name: "CastRequirement.Weapon.SubclassMask", read: func(c *core.SpellConfig) string {
		return nonZero(c.CastRequirement.Weapon.SubclassMask, func(mask int32) string { return fmt.Sprintf("%d (%s)", mask, weaponRequirement(mask)) })
	}},
	{name: "CastRequirement.Weapon.InvTypeMask", read: func(c *core.SpellConfig) string {
		return nonZero(c.CastRequirement.Weapon.InvTypeMask, func(mask int32) string { return fmt.Sprintf("%d (%s)", mask, maskNames(mask, inventoryTypes)) })
	}},
	{name: "CastRequirement.Weapon.Hand", read: func(c *core.SpellConfig) string { return nonZero(c.CastRequirement.Weapon.Hand, procMaskNames) }},
	{name: "CastRequirement.Shield", read: func(c *core.SpellConfig) string { return nonZero(c.CastRequirement.Shield, boolText) }},
	{name: "CastRequirement.NotInCombat", read: func(c *core.SpellConfig) string { return nonZero(c.CastRequirement.NotInCombat, boolText) }},
}

func missingConfigFields(config *core.SpellConfig) []string {
	var missing []string

	if config.ApplyEffects == nil {
		missing = append(missing, "ApplyEffects")
	}
	if config.ExpectedInitialDamage == nil {
		missing = append(missing, "ExpectedInitialDamage")
	}
	if config.ExpectedTickDamage == nil {
		missing = append(missing, "ExpectedTickDamage")
	}
	if config.ExtraCastCondition == nil {
		missing = append(missing, "ExtraCastCondition")
	}

	if reflect.ValueOf(config.Dot).IsZero() {
		missing = append(missing, "Dot")
	}
	if reflect.ValueOf(config.Hot).IsZero() {
		missing = append(missing, "Hot")
	}
	if reflect.ValueOf(config.Shield).IsZero() {
		missing = append(missing, "Shield")
	}

	if reflect.ValueOf(config.RelatedAuraArrays).IsZero() {
		missing = append(missing, "RelatedAuraArrays")
	}
	if config.RelatedDotSpell == nil {
		missing = append(missing, "RelatedDotSpell")
	}
	if config.RelatedSelfBuff == nil {
		missing = append(missing, "RelatedSelfBuff")
	}

	if reflect.ValueOf(config.Source).IsZero() {
		missing = append(missing, "Source")
	}

	return missing
}

// These are SpellConfig fields that SpellData intentionally does not populate.
// They are normally supplied by the caller when the spell is registered.
var configMissingFields = []string{
	"ApplyEffects",
	"ExpectedInitialDamage",
	"ExpectedTickDamage",
	"ExtraCastCondition",
	"Dot",
	"Hot",
	"Shield",
	"RelatedAuraArrays",
	"RelatedDotSpell",
	"RelatedSelfBuff",
	"Source",
}

func formsText(mask uint64) string { return fmt.Sprintf("%#x (%s)", mask, stanceList(mask)) }

func attribute(stages []core.SpellConfig, labels []string) []configRow {
	final := &stages[len(stages)-1]
	var rows []configRow

	for _, field := range configFields {
		if field.bits != nil {
			value := field.bits(final)
			if value == 0 {
				continue
			}
			var from []string
			for i := range stages {
				var previous uint64
				if i > 0 {
					previous = field.bits(&stages[i-1])
				}
				added := field.bits(&stages[i]) &^ previous & value
				if added != 0 && !slices.Contains(from, labels[i]) {
					from = append(from, labels[i])
				}
			}
			rows = append(rows, configRow{field.name, strings.Join(setBits(value, field.bitName), " | "), strings.Join(from, ", ")})
			continue
		}

		value := field.read(final)
		if value == "" {
			continue
		}
		var from []string
		previous := ""
		for i := range stages {
			current := field.read(&stages[i])
			if current != previous && current != "" {
				from = append(from, labels[i])
			}
			previous = current
		}
		rows = append(rows, configRow{field.name, value, strings.Join(from, ", ")})
	}
	return rows
}

func (result *configResult) table() configTable {
	return configTable{
		kind:          "SpellConfig",
		of:            title(result.spell),
		pick:          result.pick,
		rows:          result.rows,
		missing:       result.missing,
		skipped:       result.skipped,
		assignedLater: result.assignedLater,
	}
}

// A config the way the tool prints it: the type, the spell it is of and the row it picks, the fields
// with where each came from, a note under them, the options left unevaluated and a footnote.
type configTable struct {
	kind, of, pick string
	rows           []configRow
	codeFrom       bool
	note           string
	skipped        []configOption
	footnote       string
	missing        []string
	assignedLater  []configAssignment
}

func (t configTable) markdown(uri string) string {
	var md strings.Builder

	fmt.Fprintf(&md, "`%s`", t.kind)
	if t.of != "" {
		fmt.Fprintf(&md, " of %s", t.of)
	}
	md.WriteString("\n\n")

	if t.pick != "" {
		fmt.Fprintf(&md, "`%s`\n\n", t.pick)
	}

	md.WriteString("---\n| field | value | from |\n|---|---|---|\n")
	for _, row := range t.rows {
		from := cell(row.from)
		if t.codeFrom {
			from = codeCell(row.from)
		}
		fmt.Fprintf(&md, "| **%s** | %s | %s |\n", row.field, codeCell(row.value), from)
	}

	if len(t.assignedLater) > 0 {
		md.WriteString("\n### Assigned later\n\n")
		md.WriteString("| field | value | source |\n|---|---|---|\n")
		for _, assignment := range t.assignedLater {
			fmt.Fprintf(
				&md,
				"| **%s** | %s | [%d](%s#L%d) |\n",
				assignment.field,
				codeCell(assignment.value),
				assignment.line,
				uri,
				assignment.line,
			)
		}
	}

	if t.note != "" {
		fmt.Fprintf(&md, "\n%s\n", t.note)
	}

	for _, option := range t.skipped {
		fmt.Fprintf(&md, "\nunevaluated: `%s` (%s)\n", option.source, option.err)
	}

	if t.footnote != "" {
		fmt.Fprintf(&md, "\n%s\n", t.footnote)
	}

	return md.String()
}

func (t configTable) writeText(out io.Writer) {
	fmt.Fprint(out, t.kind)
	if t.of != "" {
		fmt.Fprintf(out, " of %s", t.of)
	}
	fmt.Fprintln(out)
	if t.pick != "" {
		fmt.Fprintf(out, "%s\n", t.pick)
	}
	fmt.Fprintln(out)
	fieldWidth, valueWidth := 0, 0
	for _, row := range t.rows {
		fieldWidth = max(fieldWidth, len(row.field))
		valueWidth = max(valueWidth, len(row.value))
	}
	for _, row := range t.rows {
		fmt.Fprintf(out, "%-*s  %-*s  %s\n", fieldWidth, row.field, valueWidth, row.value, row.from)
	}
	if t.note != "" {
		fmt.Fprintf(out, "\n%s\n", t.note)
	}
	for _, option := range t.skipped {
		fmt.Fprintf(out, "\nunevaluated: %s (%s)\n", option.source, option.err)
	}
	if t.footnote != "" {
		fmt.Fprintf(out, "\n%s\n", t.footnote)
	}
}

type configAssignment struct {
	field string
	value string
	line  int
}

func configAssignments(file *ast.File, tok *token.File, call *ast.CallExpr, name string) []configAssignment {
	var assignments []configAssignment

	var selectorPath func(ast.Expr) (string, bool)
	selectorPath = func(expr ast.Expr) (string, bool) {
		switch expr := expr.(type) {
		case *ast.Ident:
			if expr.Name != name {
				return "", false
			}
			return "", true

		case *ast.SelectorExpr:
			parent, ok := selectorPath(expr.X)
			if !ok {
				return "", false
			}
			if parent == "" {
				return expr.Sel.Name, true
			}
			return parent + "." + expr.Sel.Name, true

		default:
			return "", false
		}
	}

	// Find the function containing the SpellConfig call. We only inspect
	// assignments in that function, and do not descend into nested functions.
	var body *ast.BlockStmt
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncDecl:
			if node.Body != nil &&
				node.Body.Pos() <= call.Pos() &&
				call.End() <= node.Body.End() {
				body = node.Body
				return false
			}
		case *ast.FuncLit:
			if node.Body != nil &&
				node.Body.Pos() <= call.Pos() &&
				call.End() <= node.Body.End() {
				body = node.Body
				return false
			}
		}
		return true
	})

	if body == nil {
		return nil
	}

	ast.Inspect(body, func(node ast.Node) bool {
		// Don't inspect nested function bodies. Assignments inside a closure
		// are not assignments to the config in this function's flow.
		switch node := node.(type) {
		case *ast.FuncLit:
			return false

		case *ast.AssignStmt:
			// Only consider assignments that occur after the SpellConfig call.
			if node.Pos() <= call.End() {
				return true
			}

			for i, lhs := range node.Lhs {
				field, ok := selectorPath(lhs)
				if !ok || field == "" {
					continue
				}
				if i >= len(node.Rhs) {
					continue
				}

				assignments = append(assignments, configAssignment{
					field: field,
					value: types.ExprString(node.Rhs[i]),
					line:  tok.Line(node.Pos()),
				})
			}
		}

		return true
	})

	return assignments
}
