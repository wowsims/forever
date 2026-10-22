package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var HolyLightRankMap = spellData.HolyLight

// What Blessing of Light adds to a Holy Light or Flash of Light on a target that carries it: its
// first effect states the Holy Light bonus, its second the Flash of Light one.
var blessingOfLightRank = spellData.GreaterBlessingOfLight.Highest()
var blessingOfLightHolyLightBonus = blessingOfLightRank.EffectN(1).Average()
var blessingOfLightFlashOfLightBonus = blessingOfLightRank.EffectN(2).Average()

// Holy Light
// https://www.wowhead.com/forever/spell=25292
//
// Heals a friendly target for 1524 to 1694.
func (paladin *Paladin) registerHolyLight(_ int32, rank *spelldata.Spell) {
	heal := rank.HealEffect()

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = SpellMaskHolyLight
	// In game Holy Light makes 0.25 threat a point healed, half of what other heals make.
	config.ThreatMultiplier = 0.5

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		amount := heal.Average()
		if target.HasActiveAuraWithTag(buffs.GreaterBlessingOfLightCategory) {
			amount += blessingOfLightHolyLightBonus
		}
		spell.CalcAndDealHealing(sim, target, amount, spell.OutcomeHealingCrit)
	}

	paladin.RegisterSpell(config)
}
