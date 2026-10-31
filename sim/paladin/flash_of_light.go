package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var FlashOfLightRankMap = spellData.FlashOfLight

// Flash of Light
// https://www.wowhead.com/forever/spell=19943
//
// Heals a friendly target for 299 to 333.
func (paladin *Paladin) registerFlashOfLight(_ int32, rank *spelldata.Spell) {
	heal := rank.HealEffect()

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = SpellMaskFlashOfLight
	// In game Flash of Light makes 0.25 threat a point healed, half of what other heals make.
	config.ThreatMultiplier = 0.5

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealHealing(sim, target, heal.Average()+paladin.flashOfLightBonusHealing, spell.OutcomeHealingCrit)
	}

	paladin.RegisterSpell(config)
}
