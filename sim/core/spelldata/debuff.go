package spelldata

import (
	"slices"

	"github.com/wowsims/forever/sim/core/dbcenums"
)

// The positions, counted from 1 the way EffectN counts, of the effects that slow the enemy the spell
// lands on: a speed aura of a negative value on an enemy target that changes its attack or cast speed,
// as Frostguard's Chilled 16927 states. A row that stacks answers none, since an exclusive slow cannot
// follow stacks.
func (s *Spell) SlowEffects() []int32 {
	return s.enemyEffects(s.slows)
}

// The positions of the slows and stat changes the spell puts on the enemy it lands on, in order.
func (s *Spell) DebuffEffects() []int32 {
	return s.enemyEffects(func(e *Effect) bool { return s.slows(e) || debuffsAStat(e) })
}

func (s *Spell) enemyEffects(matches func(*Effect) bool) []int32 {
	return slices.DeleteFunc(EffectsOn(s, AuraOnEnemy), func(i int32) bool { return !matches(s.EffectN(int(i))) })
}

func (s *Spell) slows(e *Effect) bool {
	return s.MaxStack <= 0 && (e.ChangesAttackSpeed() || e.ChangesCastSpeed()) && e.BasePoints < 0
}

// Whether the effect changes a stat of the enemy the spell lands on, in the stats ParseEffects applies
// to a unit: armor and resistances, attack power, flat damage done. Annihilator's Armor Shatter 16928
// takes 165 armor per stack.
func debuffsAStat(e *Effect) bool {
	switch e.Aura {
	case dbcenums.A_MOD_RESISTANCE:
		return len(resistanceStats(e.Misc)) > 0
	case dbcenums.A_MOD_ARMOR_NO_MODS:
		return e.ArmorNoMods()
	case dbcenums.A_MOD_DAMAGE_DONE:
		return len(damageDoneStats(e.Misc)) > 0
	case dbcenums.A_MOD_ATTACK_POWER, dbcenums.A_MOD_RANGED_ATTACK_POWER:
		return true
	}
	return false
}

// Whether the row puts a debuff the sim models on the enemy it lands on.
func (s *Spell) DebuffsTheTarget() bool {
	return len(s.DebuffEffects()) > 0
}

// Whether any aura the row applies lands on an enemy. Such a row is never a buff on the wearer.
func (s *Spell) AppliesAnAuraToAnEnemy() bool {
	return len(EffectsOn(s, AuraOnEnemy)) > 0
}
