package printer

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// A core.SpellModConfig literal as core reads it: each field the evaluator could read, with the text it
// read it from, and the weapon rule the Equip fields set.
type modConfigResult struct {
	rows    []configRow
	weapon  string
	skipped []configOption
}

func isSpellModConfig(lit *ast.CompositeLit) bool {
	if name, ok := pkgSelector(lit.Type, "core"); ok {
		return name == "SpellModConfig"
	}
	ident, ok := lit.Type.(*ast.Ident)
	return ok && ident.Name == "SpellModConfig"
}

func evalSpellModConfig(lit *ast.CompositeLit, declarations map[string]declaration, pkg string, trace *tracer) *modConfigResult {
	result := &modConfigResult{}
	var config core.SpellModConfig
	target := reflect.ValueOf(&config).Elem()
	sources := map[string]string{}

	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		var key *ast.Ident
		if ok {
			key, ok = kv.Key.(*ast.Ident)
		}
		if !ok {
			result.skipped = append(result.skipped, configOption{source: types.ExprString(elt), err: fmt.Errorf("not a field: value pair")})
			continue
		}
		source := types.ExprString(kv.Value)
		field, known := target.Type().FieldByName(key.Name)
		if !known {
			result.skipped = append(result.skipped, configOption{source: key.Name + ": " + source, err: fmt.Errorf("SpellModConfig has no field %s", key.Name)})
			continue
		}
		value, err := modFieldValue(kv.Value, field.Type, declarations, pkg, trace)
		if err != nil {
			trace.add("  field %s unevaluated: %v", key.Name, err)
			result.skipped = append(result.skipped, configOption{source: key.Name + ": " + source, err: err})
			continue
		}
		trace.add("  field %s = %s", key.Name, source)
		target.FieldByIndex(field.Index).Set(value)
		sources[key.Name] = source
	}

	for _, field := range modConfigFields {
		if value := field.read(&config); value != "" {
			result.rows = append(result.rows, configRow{field.name, value, sources[field.name]})
		}
	}
	result.weapon = modWeaponRule(&config)
	return result
}

// A field's value: a constant, or a chain of the store's accessors the package's declarations resolve.
// A function, a composite and anything reading a variable are refused.
func modFieldValue(expr ast.Expr, want reflect.Type, declarations map[string]declaration, pkg string, trace *tracer) (reflect.Value, error) {
	switch want.Kind() {
	case reflect.Func, reflect.Struct:
		return reflect.Value{}, fmt.Errorf("a %s is not read", want)
	}

	value, err := evalConst(expr)
	if err != nil {
		_, read, err := evalDeclaredExpr(expr, declarations, pkg, trace, nil)
		if err != nil {
			return reflect.Value{}, err
		}
		if read.kind != kindValue {
			return reflect.Value{}, fmt.Errorf("%s reads a %s, not a value", types.ExprString(expr), read.kind)
		}
		value = chainConstant(read.value, want)
	}
	return convertArg(argument{source: types.ExprString(expr), value: value}, want)
}

// The text an accessor's value prints as, back as a constant of the kind the field takes.
func chainConstant(text string, want reflect.Type) constant.Value {
	if want == reflect.TypeFor[time.Duration]() {
		if d, err := time.ParseDuration(text); err == nil {
			return constant.MakeInt64(int64(d))
		}
	}
	if b, err := strconv.ParseBool(text); err == nil {
		return constant.MakeBool(b)
	}
	if n, err := strconv.ParseInt(text, 10, 64); err == nil {
		return constant.MakeInt64(n)
	}
	if f, err := strconv.ParseFloat(text, 64); err == nil {
		return constant.MakeFloat64(f)
	}
	return constant.MakeString(text)
}

type modConfigField struct {
	name string
	read func(*core.SpellModConfig) string
}

func procMaskNames(mask core.ProcMask) string {
	return strings.Join(setBits(uint64(mask), procMaskBit), " | ")
}

var modConfigFields = []modConfigField{
	{"Kind", func(c *core.SpellModConfig) string {
		return nonZero(c.Kind, func(kind core.SpellModType) string {
			if name, ok := constName(spellModKinds, int64(kind)); ok {
				return name
			}
			return fmt.Sprintf("kind %#x", uint64(kind))
		})
	}},
	{"ClassMask", func(c *core.SpellModConfig) string {
		return nonZero(c.ClassMask, func(mask int64) string { return fmt.Sprintf("%#x", mask) })
	}},
	{"School", func(c *core.SpellModConfig) string { return nonZero(c.School, core.SpellSchool.String) }},
	{"DefenseType", func(c *core.SpellModConfig) string { return nonZero(c.DefenseType, core.DefenseType.String) }},
	{"MultiplierSchool", func(c *core.SpellModConfig) string { return nonZero(c.MultiplierSchool, core.SpellSchool.String) }},
	{"ProcMask", func(c *core.SpellModConfig) string { return nonZero(c.ProcMask, procMaskNames) }},
	{"SpellFlag", func(c *core.SpellModConfig) string {
		return strings.Join(setBits(uint64(c.SpellFlag), spellFlagBit), " | ")
	}},
	{"ResourceType", func(c *core.SpellModConfig) string {
		return nonZero(c.ResourceType, proto.ResourceType.String)
	}},
	{"IntValue", func(c *core.SpellModConfig) string { return nonZero(c.IntValue, intText[int32]) }},
	{"TimeValue", func(c *core.SpellModConfig) string { return nonZero(c.TimeValue, durationText) }},
	{"FloatValue", func(c *core.SpellModConfig) string { return nonZero(c.FloatValue, floatText) }},
	{"KeyValue", func(c *core.SpellModConfig) string { return c.KeyValue }},
	{"ShouldApplyToPets", func(c *core.SpellModConfig) string { return nonZero(c.ShouldApplyToPets, boolText) }},
	{"EquipSubclass", func(c *core.SpellModConfig) string {
		return nonZero(c.EquipSubclass, func(mask int32) string { return fmt.Sprintf("%d (%s)", mask, weaponRequirement(mask)) })
	}},
	{"EquipInvType", func(c *core.SpellModConfig) string {
		return nonZero(c.EquipInvType, func(mask int32) string { return fmt.Sprintf("%d (%s)", mask, maskNames(mask, inventoryTypes)) })
	}},
	{"EquipHand", func(c *core.SpellModConfig) string { return nonZero(c.EquipHand, procMaskNames) }},
	{"FromRating", func(c *core.SpellModConfig) string { return nonZero(c.FromRating, boolText) }},
}

// What the Equip fields make of the mod in core's shouldApply: the weapon behind each attack is checked,
// a magic-defense spell and a pet never match, and an item swap re-filters the spells it reaches.
func modWeaponRule(c *core.SpellModConfig) string {
	if c.Equip().IsZero() {
		return ""
	}
	rule := perWeapon(*c) + ". Magic-defense spells and pets never match; an item swap re-filters it."
	if c.FromRating {
		rule += " The sheet counts it as rating."
	}
	return rule
}

func (result *modConfigResult) table() configTable {
	return configTable{kind: "SpellModConfig", rows: result.rows, codeFrom: true, note: result.weapon, skipped: result.skipped}
}
