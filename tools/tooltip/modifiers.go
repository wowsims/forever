package tooltip

import (
	"maps"
	"math"
	"slices"
	"strconv"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

// A passive talent's spell modifier on one value. Values holds one entry per rank, rank 1 first, and the player's
// points in Talent pick one. A Flat one adds its value; any other is a percentage, and two percentages multiply.
// A modifier on the base points applies before the attack power and spell power shares are added, a Whole one to
// the damage or healing amount with its shares.
type Modifier struct {
	Talent int64     `json:"t"`
	Values []float64 `json:"v"`
	Flat   bool      `json:"f,omitempty"`
	Whole  bool      `json:"w,omitempty"`
}

type spellModifier struct {
	op dbcenums.SpellModOp
	Modifier
}

// The modifiers the trees' passive talents state, by the spell they reach.
type Modifiers map[int64][]spellModifier

// The properties whose modifiers a tooltip shows: an effect's value, the duration, charges and radius in the
// text, and the cost, range, cast time and cooldown in the header.
var shownOps = map[dbcenums.SpellModOp]bool{
	dbcenums.SPELLMOD_DAMAGE: true, dbcenums.SPELLMOD_DOT: true, dbcenums.SPELLMOD_ALL_EFFECTS: true,
	dbcenums.SPELLMOD_EFFECT1: true, dbcenums.SPELLMOD_EFFECT2: true, dbcenums.SPELLMOD_EFFECT3: true,
	dbcenums.SPELLMOD_EFFECT4: true, dbcenums.SPELLMOD_EFFECT5: true,
	dbcenums.SPELLMOD_DURATION: true, dbcenums.SPELLMOD_CHARGES: true, dbcenums.SPELLMOD_RADIUS: true,
	dbcenums.SPELLMOD_COST: true, dbcenums.SPELLMOD_RANGE: true, dbcenums.SPELLMOD_CASTING_TIME: true,
	dbcenums.SPELLMOD_COOLDOWN: true,
}

// Every A_ADD_FLAT_MODIFIER and A_ADD_PCT_MODIFIER of the talents, the trees' spell ids with their points, on
// the spells their class flags name. A talent that is cast, such as Arcane Power, modifies only while its buff
// is up, and one bound to a form only in it, so neither is a modifier the player always has.
func NewModifiers(instance *dbc.DBC, talents map[int32]int32) Modifiers {
	byFamily := map[int32][]int64{}
	for _, id := range slices.Sorted(maps.Keys(instance.Spells)) {
		if family := instance.Spells[id].SpellClassSet; family != 0 {
			byFamily[family] = append(byFamily[family], int64(id))
		}
	}

	mods := Modifiers{}
	for _, id := range slices.Sorted(maps.Keys(talents)) {
		row := spelldata.Find(id)
		if row == spelldata.Nil || !row.IsPassive() || row.StanceMask != 0 {
			continue
		}
		ladder, ok := talentLadder(id, talents[id])
		if !ok {
			continue
		}
		for i, effect := range row.Effects {
			op := dbcenums.SpellModOp(effect.Misc)
			if effect.Aura != dbcenums.A_ADD_FLAT_MODIFIER && effect.Aura != dbcenums.A_ADD_PCT_MODIFIER || !shownOps[op] {
				continue
			}
			mod := spellModifier{op: op, Modifier: Modifier{Talent: int64(id), Flat: effect.Aura == dbcenums.A_ADD_FLAT_MODIFIER}}
			for n := int32(1); n <= ladder.Len(); n++ {
				mod.Values = append(mod.Values, ladder.Rank(n).Effects[i].Average())
			}
			for _, target := range byFamily[effect.ClassFlags.Family] {
				if classFlags(instance.Spells[int(target)]).Matches(effect.ClassFlags) {
					mods[target] = append(mods[target], mod)
				}
			}
		}
	}
	return mods
}

// A talent's ranks off the store's ladder; a curve whose length disagrees with the tree's points has none.
func talentLadder(id int32, points int32) (ladder spelldata.Ladder, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return spelldata.Talent(id, points), true
}

func classFlags(spell dbc.Spell) core.ClassFlags {
	flags := core.ClassFlags{Family: spell.SpellClassSet}
	for i := range min(len(spell.SpellClassMask), len(flags.Mask)) {
		flags.Mask[i] = uint32(spell.SpellClassMask[i])
	}
	return flags
}

// The modifiers of the ops on a spell, in the talents' order.
func (m Modifiers) For(spellId int64, ops ...dbcenums.SpellModOp) []Modifier {
	var out []Modifier
	for _, mod := range m[spellId] {
		if slices.Contains(ops, mod.op) {
			out = append(out, mod.Modifier)
		}
	}
	return out
}

// The ops that name one effect by its index.
var effectOps = []dbcenums.SpellModOp{dbcenums.SPELLMOD_EFFECT1, dbcenums.SPELLMOD_EFFECT2, dbcenums.SPELLMOD_EFFECT3,
	dbcenums.SPELLMOD_EFFECT4, dbcenums.SPELLMOD_EFFECT5}

// The modifiers on one effect's value: those on its base points, then the damage or healing ones, which
// SPELLMOD_DAMAGE states for a direct amount and SPELLMOD_DOT for a periodic one.
func (m Modifiers) forEffect(spellId int64, effectIdx int64, kind string) []Modifier {
	ops := []dbcenums.SpellModOp{dbcenums.SPELLMOD_ALL_EFFECTS}
	if effectIdx >= 0 && int(effectIdx) < len(effectOps) {
		ops = append(ops, effectOps[effectIdx])
	}
	out := m.For(spellId, ops...)
	var whole dbcenums.SpellModOp
	switch kind {
	case amountDamage, amountHeal:
		whole = dbcenums.SPELLMOD_DAMAGE
	case amountPeriodicDamage, amountPeriodicHeal:
		whole = dbcenums.SPELLMOD_DOT
	default:
		return out
	}
	for _, mod := range m.For(spellId, whole) {
		mod.Whole = true
		out = append(out, mod)
	}
	return out
}

// The kinds of EffectAmount. A damage kind takes its school's done multiplier, a healing kind the healing one.
const (
	amountDamage         = "dmg"
	amountPeriodicDamage = "dot"
	amountHeal           = "heal"
	amountPeriodicHeal   = "hot"
)

type modifierProvider interface {
	GetModifiers() Modifiers
}

func (c *compiler) modifiers() Modifiers {
	if p, ok := c.provider.(modifierProvider); ok {
		return p.GetModifiers()
	}
	return nil
}

// An effect amount the player's talents or done multipliers can change, with its parts kept apart for the UI:
// base, stat shares, kind, school and the modifiers that reach it. Any other amount is value unchanged.
//
// The node is {"fx": {"b": base, "s": [shares], "k": kind, "c": school mask, "m": [modifiers]}}, with every key
// but b left out where it has nothing. With no point in any of the talents and done multipliers of 1, the UI
// evaluates it to the same sum as value.
func (c *compiler) modifiedAmount(spellId int64, effectIdx int64, amount EffectAmount, base float64, shares []symbol, value symbol) symbol {
	mods := c.modifiers().forEffect(spellId, effectIdx, amount.Kind)
	if len(mods) == 0 && amount.Kind == "" {
		return value
	}
	node := map[string]Expr{"b": roundFloat(base)}
	if len(shares) > 0 {
		exprs := make([]Expr, len(shares))
		for i, share := range shares {
			exprs[i] = share.json()
		}
		node["s"] = exprs
	}
	if amount.Kind != "" {
		node["k"] = amount.Kind
	}
	if amount.Kind == amountDamage || amount.Kind == amountPeriodicDamage {
		node["c"] = amount.School
	}
	if len(mods) > 0 {
		node["m"] = mods
	}
	return symbol{expr: map[string]Expr{"fx": node}, plain: &value}
}

// A value of the spell that is not an effect amount, with the modifiers of its op on it: the duration in
// seconds, charges, a radius or the range. The game files state a flat duration in milliseconds.
func (c *compiler) modifiedValue(s *SimpleSpellValue, spellId int64, value symbol) symbol {
	var op dbcenums.SpellModOp
	scale := 1.0
	switch s.Selector.EffectColumn {
	case "d", "D":
		op, scale = dbcenums.SPELLMOD_DURATION, 0.001
	case "n", "N":
		op = dbcenums.SPELLMOD_CHARGES
	case "a", "A":
		op = dbcenums.SPELLMOD_RADIUS
	case "r", "R":
		op = dbcenums.SPELLMOD_RADIUS
		if !s.Selector.HasIndex {
			op = dbcenums.SPELLMOD_RANGE
		}
	default:
		return value
	}
	mods := c.modifiers().For(spellId, op)
	if len(mods) == 0 || !value.isConst() {
		return value
	}
	return symbol{expr: Modified(value.c, scaledModifiers(mods, scale)), plain: &value}
}

// Flat values in another unit: a duration's milliseconds as seconds.
func scaledModifiers(mods []Modifier, scale float64) []Modifier {
	if scale == 1 {
		return mods
	}
	out := make([]Modifier, len(mods))
	for i, mod := range mods {
		out[i] = mod
		if mod.Flat {
			out[i].Values = make([]float64, len(mod.Values))
			for j, v := range mod.Values {
				out[i].Values[j] = v * scale
			}
		}
	}
	return out
}

// The value without the player's modifiers, which decides how a value is written: a range or one number, the
// plural after it.
func (s symbol) unmodified() symbol {
	if s.plain != nil {
		return *s.plain
	}
	return s
}

// A value that the talents' modifiers in mods can change, as the UI reads it.
func Modified(base float64, mods []Modifier) Expr {
	return map[string]Expr{"fx": map[string]Expr{"b": roundFloat(base), "m": mods}}
}

// A number with at most decimals places, its trailing zeros dropped, as the header and a duration write one.
func trimmed(value Expr, decimals int) Segment {
	return map[string]Expr{"n": value, "p": decimals, "t": true}
}

func exprOp(name string, args ...Expr) Expr {
	return map[string]Expr{"o": name, "a": args}
}

func branch(cond Expr, yes []Segment, no []Segment) Segment {
	return map[string]Expr{"if": cond, "y": yes, "n": no}
}

// One way to write a value: the test that picks it, as holds on a number for the generator and as test on the
// expression for the UI, and the segments that write it. The last form takes every value no earlier one does.
type form struct {
	holds func(v float64) bool
	test  func(v Expr) Expr
	write func(v Expr) []Segment
}

// The value written in the first form whose test it passes. Only the forms some reachable value takes are kept:
// one form is written as it is, several become branches on their tests.
func written(value Expr, reachable []float64, forms []form) []Segment {
	used := make([]bool, len(forms))
	for _, v := range reachable {
		for i, f := range forms {
			if i == len(forms)-1 || f.holds(v) {
				used[i] = true
				break
			}
		}
	}
	var out []Segment
	for i := len(forms) - 1; i >= 0; i-- {
		if !used[i] {
			continue
		}
		if out == nil {
			out = forms[i].write(value)
			continue
		}
		out = []Segment{branch(forms[i].test(value), forms[i].write(value), out)}
	}
	return out
}

// Every value the modifiers can make of base, for every rank of each talent, none taken included. A talent
// with two modifiers on the value takes the same rank in both.
func reachable(base float64, mods []Modifier, round func(float64) float64) []float64 {
	var talents []int64
	for _, mod := range mods {
		if !slices.Contains(talents, mod.Talent) {
			talents = append(talents, mod.Talent)
		}
	}
	ranks := map[int64]int{}
	var out []float64
	var walk func(i int)
	walk = func(i int) {
		if i == len(talents) {
			flat, pct := 0.0, 1.0
			for _, mod := range mods {
				if rank := ranks[mod.Talent]; rank > 0 {
					if mod.Flat {
						flat += mod.Values[rank-1]
					} else {
						pct *= 1 + mod.Values[rank-1]/100
					}
				}
			}
			out = append(out, round((base+flat)*pct))
			return
		}
		points := 0
		for _, mod := range mods {
			if mod.Talent == talents[i] {
				points = max(points, len(mod.Values))
			}
		}
		for rank := 0; rank <= points; rank++ {
			ranks[talents[i]] = rank
			walk(i + 1)
		}
	}
	walk(0)
	return out
}

func unrounded(v float64) float64 { return v }

// A duration in seconds the modifiers can change, the way durationText writes a fixed one.
func (c *compiler) modifiedDuration(s *SimpleSpellValue, seconds float64) []Segment {
	spellId := s.getSpellId(&TooltipContext{DataProvider: c.provider, SpellId: c.spellId})
	mods := scaledModifiers(c.modifiers().For(spellId, dbcenums.SPELLMOD_DURATION), 0.001)
	return written(Modified(seconds, mods), reachable(seconds, mods, unrounded), []form{
		{
			holds: func(v float64) bool { return v <= 0 },
			test:  func(v Expr) Expr { return exprOp("not", exprOp(">", v, 0.0)) },
			write: func(Expr) []Segment { return []Segment{untilCanceled} },
		},
		{
			holds: func(v float64) bool { return v >= 7200 },
			test:  func(v Expr) Expr { return exprOp(">=", v, 7200.0) },
			write: func(v Expr) []Segment {
				return []Segment{map[string]Expr{"n": exprOp("/", v, 3600.0), "p": 0}, " hrs"}
			},
		},
		{
			holds: func(v float64) bool { return v >= 3600 },
			test:  func(v Expr) Expr { return exprOp(">=", v, 3600.0) },
			write: func(Expr) []Segment { return []Segment{"1 hour"} },
		},
		{
			holds: func(v float64) bool { return v >= 60 },
			test:  func(v Expr) Expr { return exprOp(">=", v, 60.0) },
			write: func(v Expr) []Segment { return []Segment{trimmed(exprOp("/", v, 60.0), 1), " min"} },
		},
		{write: func(v Expr) []Segment { return []Segment{trimmed(v, 1), " sec"} }},
	})
}

// A cost, range or other plain amount the modifiers can change, with its unit after it. A cost rounds down, as
// the game's does.
func ModifiedAmount(base float64, mods []Modifier, unit string, roundDown bool) []Segment {
	value := Modified(base, mods)
	if roundDown {
		value = exprOp("floor", value)
	}
	return []Segment{trimmed(value, 2), unit}
}

// The game's SPELL_DURATION_UNTIL_CANCELLED, which a duration of none or a permanent aura's reads.
const untilCanceled = "until canceled"

// A unit a header time is written in, as the game's strings word it: SPELL_RECAST_TIME_DAYS, _HOURS, _MIN and _SEC
// for a cooldown, SPELL_CAST_TIME_MIN and _SEC for a cast time, each as "%.2g day cooldown" and so on.
type timeUnit struct {
	ms   float64
	text string
}

// Largest first; a time takes the first it reaches, and one shorter than all of them the last.
var (
	castTimeUnits = []timeUnit{{60000, " min cast"}, {1000, " sec cast"}}
	cooldownUnits = []timeUnit{{86400000, " day cooldown"}, {3600000, " hour cooldown"}, {60000, " min cooldown"}, {1000, " sec cooldown"}}
)

func unitOf(ms float64, units []timeUnit) timeUnit {
	for _, unit := range units {
		if ms >= unit.ms {
			return unit
		}
	}
	return units[len(units)-1]
}

// A time in milliseconds in its unit, with the two significant digits the game's %.2g writes: 1.5 min, 12 sec.
func timeText(ms float64, units []timeUnit) string {
	unit := unitOf(ms, units)
	return strconv.FormatFloat(ms/unit.ms, 'g', 2, 64) + unit.text
}

// The forms of a time the modifiers can change, one per unit, with its number in two significant digits as
// timeText writes it: one decimal where every value the unit takes is under 10, none where every one is 10 or more,
// and a branch on 10 where both are reached. The UI rounds a half up where %.2g rounds it to even, so a value on a
// half (Vanish's 4.25 min) takes a form of its own ahead of them, with timeText's text.
func timeForms(units []timeUnit, values []float64) []form {
	var forms []form
	for _, v := range values {
		n := v / unitOf(v, units).ms
		if n < 10 {
			n *= 10
		}
		if v <= 0 || n-math.Floor(n) != 0.5 || slices.ContainsFunc(forms, func(f form) bool { return f.holds(v) }) {
			continue
		}
		forms = append(forms, form{
			holds: func(x float64) bool { return x == v },
			test:  func(x Expr) Expr { return exprOp("=", x, v) },
			write: func(Expr) []Segment { return []Segment{timeText(v, units)} },
		})
	}
	for _, unit := range units {
		under, over := false, false
		for _, v := range values {
			if v > 0 && unitOf(v, units) == unit {
				under, over = under || v < 10*unit.ms, over || v >= 10*unit.ms
			}
		}
		forms = append(forms, form{
			holds: func(v float64) bool { return v >= unit.ms },
			test:  func(v Expr) Expr { return exprOp(">=", v, unit.ms) },
			write: func(v Expr) []Segment {
				n := exprOp("/", v, unit.ms)
				switch {
				case !over:
					return []Segment{trimmed(n, 1), unit.text}
				case !under:
					return []Segment{trimmed(n, 0), unit.text}
				}
				return []Segment{trimmed(exprOp("round", n, exprOp("cond", exprOp(">=", n, 10.0), 0.0, 1.0)), 1), unit.text}
			},
		})
	}
	return forms
}

// A cast time in milliseconds as the header writes it, Instant for none.
func CastTimeText(ms float64) string {
	if ms <= 0 {
		return "Instant"
	}
	return timeText(ms, castTimeUnits)
}

// A cooldown in milliseconds as the header writes it, nothing for none.
func CooldownText(ms float64) string {
	if ms <= 0 {
		return ""
	}
	return timeText(ms, cooldownUnits)
}

// A cast time in milliseconds the modifiers can change, written as CastTimeText writes a fixed one.
func ModifiedCastTime(ms float64, mods []Modifier) []Segment {
	values := reachable(ms, mods, unrounded)
	return written(Modified(ms, mods), values, append([]form{{
		holds: func(v float64) bool { return v <= 0 },
		test:  func(v Expr) Expr { return exprOp("not", exprOp(">", v, 0.0)) },
		write: func(Expr) []Segment { return []Segment{"Instant"} },
	}}, timeForms(castTimeUnits, values)...))
}

// A cooldown in milliseconds the modifiers can change, written as CooldownText writes a fixed one.
func ModifiedCooldown(ms float64, mods []Modifier) []Segment {
	values := reachable(ms, mods, math.Round)
	return written(exprOp("round", Modified(ms, mods), 0.0), values, append([]form{{
		holds: func(v float64) bool { return v <= 0 },
		test:  func(v Expr) Expr { return exprOp("not", exprOp(">", v, 0.0)) },
		write: func(Expr) []Segment { return []Segment{} },
	}}, timeForms(cooldownUnits, values)...))
}
