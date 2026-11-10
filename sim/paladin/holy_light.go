package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var HolyLightRankMap = spellData.HolyLight

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
		if paladin.hasOathOfLight(target) {
			amount += paladin.oathOfLight.holyLight
		}
		spell.CalcAndDealHealing(sim, target, amount, spell.OutcomeHealingCrit)
	}

	paladin.RegisterSpell(config)
}
