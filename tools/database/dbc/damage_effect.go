package dbc

import (
	"slices"

	"github.com/wowsims/forever/sim/core/dbcenums"
)

// Describes a direct-damage effect reached from an item or enchant effect chain. Damage procs
// carry no stats, so the stat-based resolution in ParseStatEffect returns nothing for them and
// the effect is dropped before it is ever reported.
type DamageEffect struct {
	SpellID          int     // spell that deals the damage
	SchoolMask       int32   // DBC school mask of the damage spell
	DefenseType      int32   // SpellCategories.DefenseType of the damage spell, which picks its hit table and crit multiplier
	BonusCoefficient float64 // spell power coefficient
	IsLeech          bool    // E_HEALTH_LEECH also heals the caster, which is not modelled here
}

// The effect types whose value is flat damage. E_WEAPON_PERCENT_DAMAGE is deliberately absent: it
// is a percentage of weapon damage rather than an amount.
var directDamageEffectTypes = []SpellEffectType{dbcenums.E_SCHOOL_DAMAGE, dbcenums.E_HEALTH_LEECH}

// Walks the trigger chain below spellID and returns the first direct damage effect it reaches, or
// nil if the chain deals no flat damage.
func ResolveDamageEffect(spellID int) *DamageEffect {
	return newChainWalker().resolveDamageEffect(spellID)
}

func (w *chainWalker) resolveDamageEffect(spellID int) *DamageEffect {
	effects := w.effects(spellID)

	for _, se := range effects {
		// A_PROC_TRIGGER_DAMAGE carries its amount on the aura itself rather than on a separate
		// triggered spell, which is how the shield spikes are described.
		if !slices.Contains(directDamageEffectTypes, se.EffectType) && se.EffectAura != dbcenums.A_PROC_TRIGGER_DAMAGE {
			continue
		}

		if se.EffectBasePoints == 0 {
			continue
		}

		return &DamageEffect{
			SpellID:          spellID,
			SchoolMask:       dbcInstance.Spells[spellID].SchoolMask,
			DefenseType:      dbcInstance.Spells[spellID].DefenseType,
			BonusCoefficient: se.EffectBonusCoefficient,
			IsLeech:          se.EffectType == dbcenums.E_HEALTH_LEECH,
		}
	}

	for _, se := range effects {
		if !se.CastsItsTrigger() {
			continue
		}

		if damage := w.resolveDamageEffect(se.EffectTriggerSpell); damage != nil {
			return damage
		}
	}

	return nil
}
