package paladin

import (
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var RetributionAuraRankMap = spellData.RetributionAura

// Retribution Aura
// https://www.wowhead.com/forever/spell=10301
//
// Causes 30 Holy damage to any creature that strikes a party member within 30 yards. Players may
// only have one Aura on them per Paladin at any one time.
//
// TODO: In game the aura scales with the target's spell power (a known issue); the sim uses the
// paladin's, as the tooltip intends. Recheck after the next build.
func (paladin *Paladin) registerRetributionAura() {
	RetributionAuraRankMap.Each(func(_ int32, rank *spelldata.Spell) {
		// The damage shield's number is the rank's first effect.
		r := auraRank(rank)
		r.Value = rank.EffectN(1).Average()
		aura := buffs.NewSpellDataRetributionAuraRank(&paladin.Character, true, r, 0)
		paladin.registerAuraSpell(rank, aura, SpellMaskRetributionAura)
	})
}
