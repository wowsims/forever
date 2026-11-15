package spelldata

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// The effect whose base points are the row's factor against its targets, or NilEffect where the row
// states none or the factor is the override's own.
func (s *Spell) VersusFactorEffect() *Effect {
	if len(s.VersusMobTypes) == 0 && len(s.VersusFamilies) == 0 || s.VersusMultiplier != 0 {
		return NilEffect
	}
	return s.EffectAtIndex(s.VersusEffectIndex)
}

// The factor the row deals or restores against its targets, and false where it has none.
func (s *Spell) VersusBonus() (core.VersusBonus, bool) {
	if len(s.VersusMobTypes) == 0 && len(s.VersusFamilies) == 0 {
		return core.VersusBonus{}, false
	}
	bonus := core.VersusBonus{Multiplier: float64(s.VersusMultiplier)}
	if bonus.Multiplier == 0 {
		bonus.Multiplier = s.EffectAtIndex(s.VersusEffectIndex).BaseValue()
	}
	for _, mobType := range s.VersusMobTypes {
		bonus.MobTypes = append(bonus.MobTypes, proto.MobType(mobType))
	}
	for _, family := range s.VersusFamilies {
		bonus.Families = append(bonus.Families, proto.CreatureFamily(family))
	}
	return bonus, true
}
