package spelldata

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// What Power answers for a bar the spell does not use. Shared, like Nil and NilEffect, so a caller
// must not write through it.
var nilPower = &Power{}

// The client marks 51 instants - the hunter's shots and stings among them - with a negative cast
// time, which is no cast at all.
func (s *Spell) CastTime() time.Duration {
	return max(0, core.DurationFromMillis(s.CastTimeMs))
}

// SpellCooldowns.RecoveryTime. A spell gated by its category instead - Fire Blast and Cone of Cold
// share one - states that in CategoryCooldown.
func (s *Spell) Cooldown() time.Duration {
	return core.DurationFromMillis(s.CooldownMs)
}

func (s *Spell) CategoryCooldown() time.Duration {
	return core.DurationFromMillis(s.CategoryCooldownMs)
}

func (s *Spell) GCD() time.Duration {
	return core.DurationFromMillis(s.GCDMs)
}

// The game files state a permanent aura as -1, which is core.NeverExpires here.
func (s *Spell) Duration() time.Duration {
	if s.DurationMs == -1 {
		return core.NeverExpires
	}
	return core.DurationFromMillis(s.DurationMs)
}

// SpellMisc.MinDuration, to the millisecond: the game files state it in seconds as a float.
func (s *Spell) MissileMinDuration() time.Duration {
	return core.DurationFromMillis(math.Round(float64(s.MinDuration) * 1000))
}

// SpellAuraOptions.ProcCategoryRecovery: the internal cooldown between two procs.
func (s *Spell) ICD() time.Duration {
	return core.DurationFromMillis(s.ICDMs)
}

func (s *Spell) SpellSchool() core.SpellSchool {
	return s.School
}

func (s *Spell) DefenseTypeCore() core.DefenseType {
	return s.DefenseType
}

// The i-th effect the row carries, counted from 1 by position rather than by the game files'
// EffectIndex, which has gaps. Out of range answers NilEffect.
func (s *Spell) EffectN(i int) *Effect {
	if i < 1 || i > len(s.Effects) {
		return NilEffect
	}
	return &s.Effects[i-1]
}

// The effect at the game files' EffectIndex i, or NilEffect when the row carries none there.
func (s *Spell) EffectAtIndex(i uint8) *Effect {
	for j := range s.Effects {
		if s.Effects[j].Index == i {
			return &s.Effects[j]
		}
	}
	return NilEffect
}

// The one effect with this aura and misc value. Panics when none matches, and when two do: reading
// the first of several silently is the bug this shape exists to prevent, and EffectN is the way past
// it.
func (s *Spell) Effect(aura dbcenums.EffectAuraType, misc int32) *Effect {
	found, matches := -1, 0
	for i := range s.Effects {
		if s.Effects[i].Aura == aura && s.Effects[i].Misc == misc {
			matches++
			if found < 0 {
				found = i
			}
		}
	}
	if matches > 1 {
		panic(fmt.Sprintf("spell %d has %d effects with aura %d misc %d - index them instead",
			s.ID, matches, aura, misc))
	}
	if found < 0 {
		panic(fmt.Sprintf("spell %d has no effect with aura %d misc %d, in %d effects",
			s.ID, aura, misc, len(s.Effects)))
	}
	return &s.Effects[found]
}

// The first effect matching all three, or NilEffect. Zero matches anything it is given as: an aura of
// 0 on a direct effect is what the game files state there.
func (s *Spell) FindEffect(typ dbcenums.SpellEffectType, aura dbcenums.EffectAuraType, misc int32) *Effect {
	for i := range s.Effects {
		e := &s.Effects[i]
		if e.Type == typ && e.Aura == aura && e.Misc == misc {
			return e
		}
	}
	return NilEffect
}

// The effect carrying the spell's direct damage, whether it states an amount or a weapon multiplier.
// It sits on any of the weapon effects as readily as on school damage: a weapon effect states a
// multiplier or a flat bonus rather than an amount.
func (s *Spell) DamageEffect() *Effect {
	return s.firstOfType(dbcenums.E_SCHOOL_DAMAGE, dbcenums.E_WEAPON_DAMAGE,
		dbcenums.E_WEAPON_PERCENT_DAMAGE, dbcenums.E_NORMALIZED_WEAPON_DMG,
		dbcenums.E_WEAPON_DAMAGE_NOSCHOOL)
}

// A weapon strike's damage: its flat weapon-damage roll, the weapon damage and extra, all scaled by
// the spell's weapon percent, or taken whole where it states none. In game the percent scales the
// flat part too: Holy Strike rank 1's 12 ± 1.5 at 25% shows as 2 to 4.
func (s *Spell) WeaponStrike(sim *core.Simulation, spell *core.Spell, weaponDamage, extra float64) float64 {
	flat := s.firstOfType(dbcenums.E_NORMALIZED_WEAPON_DMG, dbcenums.E_WEAPON_DAMAGE_NOSCHOOL,
		dbcenums.E_WEAPON_DAMAGE)
	percent := 1.0
	if p := s.firstOfType(dbcenums.E_WEAPON_PERCENT_DAMAGE); p != NilEffect {
		percent = p.Percent()
	}
	return (flat.RollAt(sim, spell.Unit.Level) + weaponDamage + extra) * percent
}

// The effect granting the caster extra attacks, whose BasePoints is how many.
func (s *Spell) ExtraAttacksEffect() *Effect {
	return s.firstOfType(dbcenums.E_ADD_EXTRA_ATTACKS)
}

func (s *Spell) HealEffect() *Effect {
	return s.firstOfType(dbcenums.E_HEAL)
}

// The effect a proc's heal lands through: E_HEAL_PCT, a percentage of the target's maximum health
// (Recovery 1248759 reads 5), E_HEAL, an amount the effect rolls, or an A_PERIODIC_HEAL aura that
// heals the amount it rolls every period (Julie's Blessing 8348 reads 13 every 2 s).
func (s *Spell) ProcHealEffect() *Effect {
	if e := s.firstOfType(dbcenums.E_HEAL_PCT, dbcenums.E_HEAL); e != NilEffect {
		return e
	}
	return s.FirstAura(dbcenums.A_PERIODIC_HEAL)
}

// The A_PERIODIC_DAMAGE aura the spell applies, or NilEffect where it applies none.
func (s *Spell) PeriodicDamageEffect() *Effect {
	return s.FirstAura(dbcenums.A_PERIODIC_DAMAGE)
}

// Whether the spell deals damage to an enemy: a damage effect or a periodic damage effect that hits one.
func (s *Spell) DamagesAnEnemy() bool {
	for _, e := range []*Effect{s.DamageEffect(), s.PeriodicDamageEffect()} {
		if e != NilEffect && e.HitsAnEnemy() {
			return true
		}
	}
	return false
}

// The A_SCHOOL_ABSORB aura the spell applies, or NilEffect where it applies none. Its Misc is the mask
// of the schools it absorbs, in core.SpellSchool's bits.
func (s *Spell) AbsorbEffect() *Effect {
	return s.FirstAura(dbcenums.A_SCHOOL_ABSORB)
}

func (s *Spell) EnergizeEffect() *Effect {
	return s.firstOfType(dbcenums.E_ENERGIZE)
}

func (s *Spell) CategoryCooldownEffect() *Effect {
	return s.firstOfType(dbcenums.E_MODIFY_COOLDOWNS_BY_CATEGORY)
}

// The effect a resource gain lands through: E_ENERGIZE, an amount into the bar its misc value names
// (Burst of Energy 24532 reads 60 energy), or an A_PERIODIC_ENERGIZE aura that restores the amount
// every period (Earthen Sigil 24884 reads 40 mana every 1 s).
func (s *Spell) ProcEnergizeEffect() *Effect {
	if e := s.EnergizeEffect(); e != NilEffect {
		return e
	}
	return s.FirstAura(dbcenums.A_PERIODIC_ENERGIZE)
}

// The effect that ticks: an aura application whose aura carries a per-tick value, which is damage,
// healing, mana or a spell fired each tick.
func (s *Spell) PeriodicEffect() *Effect {
	for i := range s.Effects {
		e := &s.Effects[i]
		if e.Type != dbcenums.E_APPLY_AURA {
			continue
		}
		switch e.Aura {
		case dbcenums.A_PERIODIC_DAMAGE, dbcenums.A_PERIODIC_HEAL, dbcenums.A_PERIODIC_ENERGIZE,
			dbcenums.A_PERIODIC_TRIGGER_SPELL, dbcenums.A_PERIODIC_LEECH:
			return e
		}
	}
	return NilEffect
}

func (s *Spell) firstOfType(types ...dbcenums.SpellEffectType) *Effect {
	for i := range s.Effects {
		if slices.Contains(types, s.Effects[i].Type) {
			return &s.Effects[i]
		}
	}
	return NilEffect
}

func (s *Spell) FirstAura(auras ...dbcenums.EffectAuraType) *Effect {
	for i := range s.Effects {
		if e := &s.Effects[i]; e.Type == dbcenums.E_APPLY_AURA && slices.Contains(auras, e.Aura) {
			return e
		}
	}
	return NilEffect
}

// The cost out of this bar, or a zero Power where the spell does not use it. Both the row's own and
// the shared zero one are the store's, so a caller must not write through what it gets back.
func (s *Spell) Power(t dbcenums.PowerType) *Power {
	for i := range s.Powers {
		if s.Powers[i].Type == t {
			return &s.Powers[i]
		}
	}
	return nilPower
}

// The cost in the units the sim spends: rage off the game files' 0-1000 bar, everything else as
// stated.
func (s *Spell) PowerCost(t dbcenums.PowerType) float64 {
	cost := float64(s.Power(t).Cost)
	if t.InTenths() {
		return cost / 10
	}
	return cost
}

// The cost of the bar the spell spends - the first SpellPower row - in the units PowerCost converts
// to. 0 for a spell with no SpellPower row.
func (s *Spell) Cost() float64 {
	if len(s.Powers) == 0 {
		return 0
	}
	return s.PowerCost(s.Powers[0].Type)
}

// How many of the item each cast consumes; 0 when it is not one of the spell's reagents.
func (s *Spell) ReagentCount(item int32) int32 {
	count := int32(0)
	for _, r := range s.Reagents {
		if r.Item == item {
			count += int32(r.Count)
		}
	}
	return count
}

// Spell.NameSubtext_lang's "Rank N" as a number; 0 for a spell the game files show no rank on, whose
// subtext is "Passive" or empty.
func (s *Spell) RankNumber() int32 {
	return rankOf(s)
}

// The name with its "Rank N" subtext, "Holy Fire Rank 8"; a spell the game files show no rank on keeps
// its name.
func (s *Spell) RankedName() string {
	if s.RankNumber() == 0 {
		return s.Name
	}
	return s.Name + " " + s.Rank
}

// The ranked name with the caster's label, "Holy Fire Rank 8 - Priest (#1)".
func (s *Spell) RankedLabel(owner *core.Unit) string {
	return s.RankedName() + " - " + owner.Label
}

// The spells the tooltip names, in the order it names them. An id the store does not carry is left
// out rather than answered as Nil.
func (s *Spell) Refs() []*Spell {
	return resolve(s.RefIDs)
}

// The spell an A_OVERRIDE_ACTIONBAR_SPELLS effect of overrider puts on the action bar in place of
// this one. It panics when overrider does not replace this spell.
func (s *Spell) OverriddenBy(overrider *Spell) *Spell {
	return MustFind(int32(overrider.Effect(dbcenums.A_OVERRIDE_ACTIONBAR_SPELLS, s.ID).BasePoints))
}

// The form an A_MOD_SHAPESHIFT effect puts the caster in, or none.
func (s *Spell) ShapeshiftForm() dbcenums.ShapeshiftForm {
	for i := range s.Effects {
		if s.Effects[i].Aura == dbcenums.A_MOD_SHAPESHIFT {
			return dbcenums.ShapeshiftForm(s.Effects[i].Misc)
		}
	}
	return 0
}

// The spells whose effects fire this one, from the trigger index.
func (s *Spell) Drivers() []*Spell {
	return resolve(drivers[s.ID])
}

// Every spell this one's effects fire, deduped, in effect order, and then the ones a server-side
// handler casts off it.
func (s *Spell) Triggered() []*Spell {
	var ids []int32
	for _, e := range s.Effects {
		if e.TriggerID != 0 && !slices.Contains(ids, e.TriggerID) {
			ids = append(ids, e.TriggerID)
		}
	}
	for _, id := range handTriggers[s.ID] {
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return resolve(ids)
}

// A tick of a damage over time. The hit is rolled once, when the dot is applied, so the tick rolls
// none: a crit only where the row states Periodic Can Crit, on the crit of the dot spell's defense
// type, and the partial resist its school takes.
func (s *Spell) TickOutcome(dot *core.Dot) core.OutcomeApplier {
	switch tickOutcomeKind(s.PeriodicCanCrit(), dot.Spell.DefenseType == core.DefenseTypeMagic) {
	case tickOutcomeMagicCrit:
		return dot.Spell.OutcomeTickMagicCrit
	case tickOutcomePhysicalCrit:
		return dot.Spell.OutcomeTickPhysicalCrit
	default:
		return dot.OutcomeTick
	}
}

// Which of the three ticks the two flags pick. Split out so the choice can be asserted without a Dot.
const (
	tickOutcomeMagicCrit = iota
	tickOutcomePhysicalCrit
	tickOutcomePlain
)

func tickOutcomeKind(canCrit bool, magic bool) int {
	switch {
	case canCrit && magic:
		return tickOutcomeMagicCrit
	case canCrit:
		return tickOutcomePhysicalCrit
	default:
		return tickOutcomePlain
	}
}

func resolve(ids []int32) []*Spell {
	var out []*Spell
	for _, id := range ids {
		if s := Find(id); s != Nil {
			out = append(out, s)
		}
	}
	return out
}
