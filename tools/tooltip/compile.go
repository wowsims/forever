package tooltip

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/participle/v2"
)

// Slips in the game files' texts the grammar does not read, corrected before parsing: a doubled $, a $
// before an operator (Frostbolt's $m2$*), $s~1 for $s1, a bare $1 for effect 1's value (Increased Pet Armor),
// $?$p for $?p (Tranquility), a $ before a % (Screeching Frenzy's $s$%), $10/;s1 for $/10;s1 (Revitalizing
// Waters), $25695/5;s1 for $/5;25695s1 (Arathi Basin Enriched Ration), a line break between a condition and
// its branch (Earth Totems' $?pc166190\r\n[) and between its two branches (Chronostatic Preservation).
var compileFixes = []tooltipFix{
	{Regex: regexp.MustCompile(`\$\$`), Replace: "$$"},
	{Regex: regexp.MustCompile(`(\$[mMsS]\d)\$\*`), Replace: "${1}*"},
	{Regex: regexp.MustCompile(`\$([sSmMwW])~(\d)`), Replace: "$$${1}${2}"},
	{Regex: regexp.MustCompile(`\$(\d)(\D|$)`), Replace: "$$s${1}${2}"},
	{Regex: regexp.MustCompile(`\$\?\$([aps]\d+)`), Replace: "$$?${1}"},
	{Regex: regexp.MustCompile(`(\$\d*[a-zA-Z]\d{0,2})\$%`), Replace: "${1}%"},
	{Regex: regexp.MustCompile(`\$(\d+)/;`), Replace: "$$/${1};"},
	{Regex: regexp.MustCompile(`\$(\d+)/(\d+);([a-zA-Z]\d)`), Replace: "$$/${2};${1}${3}"},
	{Regex: regexp.MustCompile(`(\$\?[a-z]{1,2}\d+)\s+\[`), Replace: "${1}["},
	{Regex: regexp.MustCompile(`\]\s+\[`), Replace: "]["},
	// Colours, |cFFFFFFFF...|r and |cnGREEN_FONT_COLOR:...|r, which the site's tooltips do not take.
	{Regex: regexp.MustCompile(`(?i)\|c(?:[0-9a-f]{8}|n[a-z_]+:)`), Replace: ""},
	{Regex: regexp.MustCompile(`(?i)\|r`), Replace: ""},
	// A mount's speed, $j1g for the ground, comes from its MountCapability rows, which the extracted game
	// files tables do not hold, so the condition on it takes its other branch.
	{Regex: regexp.MustCompile(`\$\?j\d[a-z]\[[^\[\]]*\]\[([^\[\]]*)\]`), Replace: "${1}"},
}

func fixForCompile(text string) string {
	for _, f := range compileFixes {
		text = f.Regex.ReplaceAllString(text, f.Replace)
	}
	for _, f := range fixes {
		if f.Regex != lineBreakFix.Regex {
			text = f.Regex.ReplaceAllString(text, f.Replace)
		}
	}
	return text
}

// Building the grammar takes longer than a parse, so the compiler builds it once.
var tooltipParser = sync.OnceValue(func() *participle.Parser[TooltipAST] {
	return participle.MustBuild[TooltipAST](participle.Lexer(getLexer()), participle.Elide("Whitespace", "Comment", "CommentStart"), participle.UseLookahead(-1))
})

// A description compiled for the site: text, with the numbers that depend on the player left as expressions
// the UI evaluates with the player's stats and talents. Everything else is worked out here, the way String
// renders it.
//
// A segment is a string; a number {"n": expr, "p": decimals}, with "abs": true where the game writes the value
// unsigned; or a branch {"if": expr, "y": segments, "n": segments}. An expression is a number; a player value
// {"v": name}, named as the game's own tokens name them (AP, RAP, SP, SPFR, SPI, PL, mws, ...) plus HEAL for
// healing power; a talent {"t": spellId, "r": rank}, 1 when the player has rank points in it (r is 1 when
// absent); or an operation
// {"o": op, "a": [expressions]}.
type Segment = any

type Expr = any

// A value while compiling: a constant, or an expression over the player.
type symbol struct {
	c     float64
	expr  Expr    // nil for a constant
	plain *symbol // for an amount the player's modifiers can change, the value without them
}

func constant(c float64) symbol { return symbol{c: c} }

func (s symbol) isConst() bool { return s.expr == nil }

func (s symbol) json() Expr {
	if s.isConst() {
		return roundFloat(s.c)
	}
	return s.expr
}

// Seven significant digits, which drops the float32 widening of a game files coefficient: 0.857, not
// 0.8569999933242798.
func roundFloat(v float64) float64 {
	rounded, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'g', 7, 64), 64)
	return rounded
}

// Folds an operation whose arguments are all constants; otherwise keeps it for the UI.
func operation(op string, fold func(...float64) float64, args ...symbol) symbol {
	values := make([]float64, len(args))
	exprs := make([]Expr, len(args))
	allConst := true
	for i, a := range args {
		values[i], exprs[i] = a.c, a.json()
		allConst = allConst && a.isConst()
	}
	if allConst {
		return constant(fold(values...))
	}
	return symbol{expr: map[string]Expr{"o": op, "a": exprs}}
}

func boolValue(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

var folds = map[string]func(...float64) float64{
	"+":     func(v ...float64) float64 { return v[0] + v[1] },
	"-":     func(v ...float64) float64 { return v[0] - v[1] },
	"*":     func(v ...float64) float64 { return v[0] * v[1] },
	"/":     func(v ...float64) float64 { return v[0] / v[1] },
	"^":     func(v ...float64) float64 { return math.Pow(v[0], v[1]) },
	"neg":   func(v ...float64) float64 { return -v[0] },
	"max":   func(v ...float64) float64 { return math.Max(v[0], v[1]) },
	"min":   func(v ...float64) float64 { return math.Min(v[0], v[1]) },
	"abs":   func(v ...float64) float64 { return math.Abs(v[0]) },
	"floor": func(v ...float64) float64 { return math.Floor(v[0]) },
	"clamp": func(v ...float64) float64 { return math.Min(math.Max(v[0], v[1]), v[2]) },
	"round": func(v ...float64) float64 { f := math.Pow(10, v[1]); return math.Round(v[0]*f) / f },
	"cond": func(v ...float64) float64 {
		if v[0] > 0 {
			return v[1]
		}
		return v[2]
	},
	">":   func(v ...float64) float64 { return boolValue(v[0] > v[1]) },
	">=":  func(v ...float64) float64 { return boolValue(v[0] >= v[1]) },
	"<":   func(v ...float64) float64 { return boolValue(v[0] < v[1]) },
	"=":   func(v ...float64) float64 { return boolValue(v[0] == v[1]) },
	"!=":  func(v ...float64) float64 { return boolValue(v[0] != v[1]) },
	"and": func(v ...float64) float64 { return boolValue(v[0] != 0 && v[1] != 0) },
	"or":  func(v ...float64) float64 { return boolValue(v[0] != 0 || v[1] != 0) },
	"not": func(v ...float64) float64 { return boolValue(v[0] == 0) },
}

func op(name string, args ...symbol) symbol {
	// x*1 and x+0 are x: a coefficient of 1 or a talent multiplier no player can raise leaves the bare value.
	if len(args) == 2 {
		identity := map[string]float64{"*": 1, "+": 0}
		if id, ok := identity[name]; ok {
			if args[0].isConst() && args[0].c == id {
				return args[1]
			}
			if args[1].isConst() && args[1].c == id {
				return args[0]
			}
		}
	}
	return operation(name, folds[name], args...)
}

// The game's tokens for the player's stats and weapons, as the UI names them. apm is the melee attack power
// Revenge scales with; rap the ranged attack power Summon Hawk's dive does; bh the healing power Prayer of
// Mending's ${($m1+($bh*$bc))*$<mult>} adds.
var playerValues = map[string]string{
	"AP": "AP", "apm": "AP", "RAP": "RAP", "rap": "RAP",
	"bh": "HEAL",
	"SP": "SP", "SPA": "SPA", "SPFI": "SPFI", "SPFR": "SPFR", "SPH": "SPH", "SPN": "SPN", "SPS": "SPS",
	"sps": "SPS", "spfi": "SPFI",
	"SPI": "SPI", "STR": "STR", "AGI": "AGI", "INT": "INT", "STA": "STA",
	"PL": "PL", "pl": "PL",
	"mwb": "mwb", "MWB": "MWB", "mws": "mws", "MWS": "mws",
	"owb": "owb", "OWB": "OWB", "ows": "ows", "OWS": "ows",
}

// The spell power token of a school mask's lowest school.
func schoolPower(school int32) string {
	switch {
	case school&2 != 0:
		return "SPH"
	case school&4 != 0:
		return "SPFI"
	case school&8 != 0:
		return "SPN"
	case school&16 != 0:
		return "SPFR"
	case school&32 != 0:
		return "SPS"
	case school&64 != 0:
		return "SPA"
	}
	return "SP"
}

// The effect amounts a provider can split into the shares of the player's stats.
type amountProvider interface {
	GetEffectAmount(spellId int64, effectIdx int64) EffectAmount
}

// The spell power coefficient a provider reads off an effect's row, whatever the effect does: $bc, which
// Prayer of Mending's dummy effect carries.
type coefficientProvider interface {
	GetEffectBonusCoefficient(spellId int64, effectIdx int64) float64
}

type compiler struct {
	provider TooltipDataProvider
	talents  map[int64]Talent // the talent ranks a condition on is left to the player
	spellId  int64
	vars     map[string]symbol
	last     float64 // the last number written, which a plural form after it agrees with
	depth    int
	included []error // the texts this one includes that did not compile
	text     string  // the text as parsed, whose spaces the compiled text keeps
}

// A talent rank's spell as a condition reads it: Spell is the talent the trees offer, whose points must reach
// Rank. Spell is 0 for a rank of a talent no tree offers, which no player has.
type Talent struct {
	Spell int64
	Rank  int
}

// The text of a spell's description, or of any text in its format, compiled for the site. A condition on a
// talent rank in talents is left to the player's points; every other condition is decided the way String
// decides it.
func Compile(text string, provider TooltipDataProvider, spellId int64, talents map[int64]Talent) ([]Segment, error) {
	return compileAt(text, provider, spellId, talents, 0)
}

func compileAt(text string, provider TooltipDataProvider, spellId int64, talents map[int64]Talent, depth int) (segments []Segment, err error) {
	// A row the grammar reads but this pass does not cover fails one tooltip, not the run.
	defer func() {
		if r := recover(); r != nil {
			segments, err = nil, fmt.Errorf("spell %d: %v", spellId, r)
		}
	}()

	parser := tooltipParser()
	c := &compiler{provider: provider, talents: talents, spellId: spellId, vars: map[string]symbol{}, depth: depth}
	vars, err := parser.ParseString("", fixForCompile(provider.GetDescriptionVariableString(spellId)))
	if err != nil {
		return nil, err
	}
	if vars.Values != nil {
		for _, v := range *vars.Values {
			if v.VariableAssignment != nil {
				c.vars[v.VariableAssignment.VariableName] = c.complexNumber(v.VariableAssignment.Value)
			}
		}
	}

	c.text = fixForCompile(text)
	ast, err := parser.ParseString("", c.text)
	if err != nil {
		return nil, err
	}
	segments = c.ast(ast)
	if depth == 0 {
		segments = tidy(segments)
	}
	// A text that includes one that did not compile keeps the rest of its text, and fails with the included one.
	return segments, errors.Join(c.included...)
}

// The game files' line breaks as newlines, with none left at either end.
func tidy(segments []Segment) []Segment {
	for i, s := range segments {
		if text, ok := s.(string); ok {
			text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
			if i == 0 {
				text = strings.TrimLeft(text, " \n")
			}
			if i == len(segments)-1 {
				text = strings.TrimRight(text, " \n")
			}
			segments[i] = text
		}
	}
	return segments
}

// The values with a space where the text has one: before a value, and after the last one of a branch, as
// $?$PL<30[ and ][, ] puts them.
func (c *compiler) ast(t *TooltipAST) []Segment {
	var out []Segment
	if t == nil || t.Values == nil {
		return out
	}
	values := *t.Values
	for i, v := range values {
		parts := c.complex(v)
		if len(parts) == 0 {
			continue
		}
		if text, ok := parts[0].(string); ok && text == "" && len(parts) == 1 {
			continue
		}
		if c.spaceAt(v.Pos.Offset - 1) {
			out = append(out, " ")
		}
		out = append(out, parts...)
		if i == len(values)-1 && c.spaceAt(v.EndPos.Offset) {
			out = append(out, " ")
		}
	}
	return merge(out)
}

// Whether the text has a space or a tab at offset, which the lexer drops between tokens.
func (c *compiler) spaceAt(offset int) bool {
	return offset >= 0 && offset < len(c.text) && (c.text[offset] == ' ' || c.text[offset] == '\t')
}

// Joins neighbouring strings.
func merge(segments []Segment) []Segment {
	var out []Segment
	for _, s := range segments {
		if text, ok := s.(string); ok && len(out) > 0 {
			if prev, ok := out[len(out)-1].(string); ok {
				out[len(out)-1] = prev + text
				continue
			}
		}
		out = append(out, s)
	}
	return out
}

func (c *compiler) number(value symbol, decimals int, abs bool) Segment {
	if value.isConst() {
		c.last = value.c
		v := value.c
		if abs {
			v = math.Abs(v)
		}
		// A half rounds away from zero, as the UI's toFixed rounds the numbers it works out.
		scale := math.Pow(10, float64(decimals))
		return strconv.FormatFloat(math.Round(v*scale)/scale, 'f', decimals, 64)
	}
	c.last = 2 // a value only the player knows reads as plural
	if plain := value.unmodified(); plain.isConst() {
		c.last = plain.c // one only the player's modifiers change agrees with its value without them
	}
	segment := map[string]Expr{"n": value.expr, "p": decimals}
	if abs {
		segment["abs"] = true
	}
	return segment
}

func (c *compiler) complex(v ComplexValue) []Segment {
	switch {
	case v.VariableAssignment != nil:
		return nil
	case v.Terniary != nil:
		return c.ternary(v.Terniary)
	case v.Word != nil:
		return []Segment{*v.Word}
	case v.ComputedValue != nil:
		return c.computedText(v.ComputedValue)
	case v.Punctuation != nil:
		return []Segment{*v.Punctuation}
	case v.MathExpression != nil:
		return []Segment{c.number(c.mathExpression(v.MathExpression), int(v.MathExpression.GetDecimalPlace()), false)}
	case v.ShortTernary != nil:
		ctx := &TooltipContext{DataProvider: c.provider, LastEval: c.last}
		return []Segment{v.ShortTernary.String(ctx)}
	case v.DescriptionRef != nil:
		if c.depth > 3 {
			return nil
		}
		ref := v.DescriptionRef.SpellId
		segments, err := compileAt(v.DescriptionRef.Text(c.provider), c.provider, ref, c.talents, c.depth+1)
		if err != nil {
			c.included = append(c.included, fmt.Errorf("the text of %d it includes: %w", ref, err))
		}
		return segments
	case v.SpellNameRef != nil:
		return []Segment{c.provider.GetSpellName(v.SpellNameRef.SpellId)}
	case v.AuraCaster:
		return []Segment{auraCasterText}
	}
	return nil
}

func (c *compiler) ternary(t *Ternary) []Segment {
	cond := c.boolExpression(t.BoolExpr)
	first := c.ast(t.FirstValue)
	var second []Segment
	if t.SecondValue.Chained != nil {
		second = c.ternary(t.SecondValue.Chained)
	} else {
		second = c.ast(t.SecondValue.SecondValue)
	}
	if cond.isConst() {
		if cond.c != 0 {
			return first
		}
		return second
	}
	// An empty branch, $?$PL<10[]?, is an empty list: the UI renders whichever branch the player takes.
	return []Segment{map[string]Expr{"if": cond.expr, "y": append([]Segment{}, first...), "n": append([]Segment{}, second...)}}
}

// A text segment's value as a number, for a variable assigned from a branch or a word.
func (c *compiler) astNumber(t *TooltipAST) symbol {
	if t != nil && t.Values != nil && len(*t.Values) == 1 {
		return c.complexNumber(&(*t.Values)[0])
	}
	text := ""
	for _, s := range c.ast(t) {
		if str, ok := s.(string); ok {
			text += str
		}
	}
	f, _ := strconv.ParseFloat(strings.TrimSpace(text), 64)
	return constant(f)
}

func (c *compiler) complexNumber(v *ComplexValue) symbol {
	switch {
	case v.MathExpression != nil:
		return c.mathExpression(v.MathExpression)
	case v.ComputedValue != nil:
		return c.computed(v.ComputedValue)
	case v.Terniary != nil:
		return c.ternaryNumber(v.Terniary)
	case v.Word != nil:
		f, _ := strconv.ParseFloat(*v.Word, 64)
		return constant(f)
	}
	return constant(0)
}

func (c *compiler) ternaryNumber(t *Ternary) symbol {
	cond := c.boolExpression(t.BoolExpr)
	first := c.astNumber(t.FirstValue)
	var second symbol
	if t.SecondValue.Chained != nil {
		second = c.ternaryNumber(t.SecondValue.Chained)
	} else {
		second = c.astNumber(t.SecondValue.SecondValue)
	}
	return op("cond", cond, first, second)
}

func (c *compiler) mathExpression(m *MathExpression) symbol {
	value := c.expression(m.Expression)
	if m.Round == nil {
		return value
	}
	return op("round", value, constant(*m.Round*10))
}

func (c *compiler) expression(e *Expression) symbol {
	value := c.simpleTerm(e.Left)
	for _, r := range e.Right {
		value = op(r.Operator.String(), value, c.simpleTerm(r.Term))
	}
	return value
}

func (c *compiler) simpleTerm(t *MathSimpleTerm) symbol {
	value := c.factor(t.Left)
	for _, r := range t.Right {
		value = op(r.Operator.String(), value, c.factor(r.Factor))
	}
	return value
}

func (c *compiler) factor(f *MathFactorTerm) symbol {
	base := c.mathValue(f.Base)
	if f.Exponent != nil {
		return op("^", base, c.mathValue(f.Exponent))
	}
	return base
}

func (c *compiler) mathValue(v *MathValue) symbol {
	switch {
	case v.Number != nil:
		return constant(*v.Number)
	case v.Variable != nil:
		return c.variable(*v.Variable)
	case v.ComputedValue != nil:
		return c.computed(v.ComputedValue)
	case v.Subexpression != nil:
		return c.expression(v.Subexpression)
	}
	return constant(0)
}

func (c *compiler) variable(name string) symbol {
	if value, ok := c.vars[name]; ok {
		return value
	}
	if player, ok := playerValues[name]; ok {
		return symbol{expr: map[string]Expr{"v": player}}
	}
	switch name {
	case "proccooldown":
		return constant(c.provider.GetSpellProcCooldown(c.spellId).Seconds())
	case "procrppm":
		return constant(c.provider.GetSpellPPM(c.spellId))
	case "bc":
		if p, ok := c.provider.(coefficientProvider); ok {
			return constant(p.GetEffectBonusCoefficient(c.spellId, 0))
		}
	}
	return constant(0)
}

func (c *compiler) computed(v *ComputedValue) symbol {
	switch {
	case v.Negative != nil:
		return op("neg", c.computed(v.Negative.Value))
	case v.SimpleCompute != nil:
		return scaled(v.SimpleCompute, c.spellValue(v.SimpleCompute.Value))
	case v.SpellValue != nil:
		return c.spellValue(v.SpellValue)
	case v.VariableRef != nil:
		name := v.VariableRef.DynamicVarName
		if name == "" {
			name = v.VariableRef.StaticVarName
		}
		return c.variable(name)
	case v.Function != nil:
		return c.function(v.Function)
	}
	return constant(0)
}

func (c *compiler) function(f *Function) symbol {
	args := make([]symbol, len(*f.Args))
	for i := range *f.Args {
		args[i] = c.expression(&(*f.Args)[i])
	}
	name := strings.ToLower(f.Name)
	switch name {
	case "gt":
		name = ">"
	case "gte":
		name = ">="
	case "lt":
		name = "<"
	}
	if _, ok := folds[name]; !ok {
		panic(f.Name + " is not a function the compiler knows")
	}
	return op(name, args...)
}

// A spell value; a damage or healing amount carries the player's attack power and spell power shares.
func (c *compiler) spellValue(s *SimpleSpellValue) symbol {
	low, _ := c.spellRange(s)
	return low
}

// The bottom and top of a spell value. An amount the sim rolls in a range, the store's, has two; any other
// value is both.
func (c *compiler) spellRange(s *SimpleSpellValue) (symbol, symbol) {
	ctx := &TooltipContext{DataProvider: c.provider, SpellId: c.spellId}
	column := s.Selector.EffectColumn
	scaling, ok := c.provider.(amountProvider)
	if !ok || !strings.Contains("swomM", column) && column != "S" && column != "W" && column != "O" {
		value := c.modifiedValue(s, s.getSpellId(ctx), constant(s.Eval(ctx)))
		return value, value
	}

	// $m and $M are the bottom and top of the roll, $s and $w both ends of it; the shares of the player's stats
	// ride on each, so Frostbolt's ${$m2*$<frostdamage>} grows with frost spell power too.
	spellId := s.getSpellId(ctx)
	amount := scaling.GetEffectAmount(spellId, s.Selector.EffectIndex-1)
	low, high := amount.Base, amount.Base
	switch {
	case column == "m" || column == "M":
		low = s.Eval(ctx)
		high = low
	case (amount.Low != 0 || amount.High != 0) && strings.Contains("sSwW", column):
		low, high = amount.Low, amount.High
	}
	return c.withShares(s, spellId, amount, low), c.withShares(s, spellId, amount, high)
}

func (c *compiler) withShares(s *SimpleSpellValue, spellId int64, amount EffectAmount, base float64) symbol {
	value := constant(base)
	var shares []symbol
	if amount.AP > 0 {
		power := "AP"
		if amount.Ranged {
			power = "RAP"
		}
		shares = append(shares, op("*", constant(amount.AP), symbol{expr: map[string]Expr{"v": power}}))
		value = op("+", value, shares[len(shares)-1])
	}
	// Spell power adds nothing to a physical effect, whatever coefficient its row states (Revenge states 1).
	if amount.SP > 0 && (amount.Heal || amount.School&^1 != 0) {
		power := schoolPower(amount.School)
		if amount.Heal {
			power = "HEAL"
		}
		shares = append(shares, op("*", constant(amount.SP), symbol{expr: map[string]Expr{"v": power}}))
		value = op("+", value, shares[len(shares)-1])
	}
	value = c.modifiedAmount(spellId, s.Selector.EffectIndex-1, amount, base, shares, value)
	if strings.ToLower(s.Selector.EffectColumn) == "o" {
		period := c.provider.GetEffectPeriod(spellId, s.Selector.EffectIndex-1)
		if period == 0 {
			return constant(0)
		}
		ticks := float64(c.provider.GetSpellDuration(spellId)) / float64(period)
		value = op("*", value, constant(ticks))
	}
	return value
}

// $/10;s1 and $*2;s1: the value divided or multiplied by the number.
func scaled(sc *SimpleCompute, value symbol) symbol {
	if sc.Op == "/" {
		return op("/", value, constant(float64(sc.Num)))
	}
	return op("*", value, constant(float64(sc.Num)))
}

// Whether two values are the same number for every player, which the text then writes once.
func sameValue(a, b symbol) bool {
	return fmt.Sprint(a.json()) == fmt.Sprint(b.json())
}

// A spell value as the game writes it: durations in words, periods with one decimal, a radius with its
// decimals, the rest unsigned.
func (c *compiler) computedText(v *ComputedValue) []Segment {
	switch {
	case v.Negative != nil:
		// The game writes the minus itself before an unsigned value: -$/10;s1 rage cost reads -2.
		return append([]Segment{"-"}, c.computedText(v.Negative.Value)...)
	case v.SpellValue != nil:
		column := v.SpellValue.Selector.EffectColumn
		value := c.spellValue(v.SpellValue)
		switch column {
		case "d", "D":
			c.last = value.unmodified().c
			if !value.isConst() {
				return c.modifiedDuration(v.SpellValue, value.unmodified().c)
			}
			return []Segment{durationText(value.c)}
		case "t", "T":
			seconds := value.c / float64(time.Second)
			c.last = seconds
			return []Segment{strconv.FormatFloat(math.Round(seconds*100)/100, 'f', -1, 64)}
		case "e":
			return []Segment{c.number(value, 1, false)}
		case "s", "S", "w", "W":
			if low, high := c.spellRange(v.SpellValue); !sameValue(low.unmodified(), high.unmodified()) {
				return []Segment{c.number(low, 0, true), " to ", c.number(high, 0, true)}
			}
		case "a", "A", "r", "R":
			if value.isConst() {
				c.last = value.c
				return []Segment{strconv.FormatFloat(math.Round(math.Abs(value.c)*100)/100, 'f', -1, 64)}
			}
		}
		return []Segment{c.number(value, 0, true)}
	case v.SimpleCompute != nil:
		// One decimal, none on a whole number: $/1000;S1 reads 0.5 and 4, $/2;20467s1 69 to 73.
		write := func(value symbol) Segment {
			if !value.isConst() {
				// A value the player changes keeps its decimal only where it has one: 69, or 75.9 with a talent.
				segment := c.number(value, 1, true)
				if s, ok := segment.(map[string]Expr); ok {
					s["t"] = true
				}
				return segment
			}
			decimals := 1
			if math.Mod(math.Round(value.c*10), 10) == 0 {
				decimals = 0
			}
			return c.number(value, decimals, true)
		}
		low, high := c.spellRange(v.SimpleCompute.Value)
		if !sameValue(low, high) {
			return []Segment{write(scaled(v.SimpleCompute, low)), " to ", write(scaled(v.SimpleCompute, high))}
		}
		return []Segment{write(scaled(v.SimpleCompute, low))}
	case v.VariableRef != nil:
		return []Segment{c.number(c.computed(v), 0, false)}
	}
	return nil
}

// A duration in seconds the way the game's tooltips write one: 3 sec, 1.5 sec, 10 min, 1 hour, 2 hrs, and until
// canceled for none, which a permanent aura's -1 reads as too.
func durationText(seconds float64) string {
	switch {
	case seconds <= 0:
		return untilCanceled
	case seconds >= 7200:
		return strconv.FormatFloat(math.Round(seconds/3600), 'f', -1, 64) + " hrs"
	case seconds >= 3600:
		return "1 hour"
	case seconds >= 60:
		return strconv.FormatFloat(math.Round(seconds/60*10)/10, 'f', -1, 64) + " min"
	}
	return strconv.FormatFloat(math.Round(seconds*10)/10, 'f', -1, 64) + " sec"
}

func (c *compiler) boolExpression(b *BoolExpression) symbol {
	value := c.boolTerm(b.Left)
	for _, r := range b.Right {
		name := "and"
		if r.Operator == "|" {
			name = "or"
		}
		value = op(name, value, c.boolTerm(r.Term))
	}
	return value
}

func (c *compiler) boolTerm(b *BoolGeneralTerm) symbol {
	if len(b.Right) == 0 {
		return c.boolTerminal(b.Left)
	}
	left := c.boolNumber(b.Left)
	right := c.boolNumber(b.Right[0].Value)
	switch b.Right[0].Operator {
	case ">", ">=", "<", "=", "!=":
		return op(b.Right[0].Operator, left, right)
	case "==":
		return op("=", left, right)
	case "<=":
		return op(">=", right, left)
	}
	return constant(0)
}

func (c *compiler) boolNumber(b *BoolTerminalValue) symbol {
	switch {
	case b.Number != nil:
		return constant(*b.Number)
	case b.SpellRef != nil:
		return c.spellValue(b.SpellRef)
	case b.Variable != nil:
		return c.variable(*b.Variable)
	}
	return c.boolTerminal(b)
}

func (c *compiler) boolTerminal(b *BoolTerminalValue) symbol {
	var value symbol
	switch {
	case b.SpellCondition != nil:
		cond := b.SpellCondition
		kind := strings.TrimPrefix(cond.Op, "?")
		if talent, ok := c.talents[cond.SpellId]; ok && kind != "c" && kind != "pc" {
			value = constant(0)
			if talent.Spell != 0 {
				expr := map[string]Expr{"t": talent.Spell}
				if talent.Rank > 1 {
					expr["r"] = talent.Rank
				}
				value = symbol{expr: expr}
			}
		} else if known, ok := c.playerCondition(kind, cond.SpellId); ok {
			value = known
		} else {
			value = constant(boolValue(cond.EvalBool(&TooltipContext{DataProvider: c.provider, SpellId: c.spellId})))
		}
	case b.Subexpression != nil:
		value = c.boolExpression(b.Subexpression)
	case b.Number != nil:
		value = constant(boolValue(*b.Number != 0))
	case b.SpellRef != nil:
		value = op("!=", c.spellValue(b.SpellRef), constant(0))
	case b.Variable != nil:
		value = op("!=", c.variable(*b.Variable), constant(0))
	default:
		value = constant(0)
	}
	if b.Negated != nil {
		return op("not", value)
	}
	return value
}
