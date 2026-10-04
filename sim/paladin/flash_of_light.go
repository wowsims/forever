package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var FlashOfLightRankMap = spellData.FlashOfLight

// Flash of Light
// https://www.wowhead.com/forever/spell=19943
//
// Heals a friendly target for 308.
func (paladin *Paladin) registerFlashOfLight(_ int32, rank *spelldata.Spell) {
	heal := rank.HealEffect()

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.ID},
		SpellSchool:    rank.SpellSchool(),
		DefenseType:    rank.DefenseTypeCore(),
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskFlashOfLight,
		Rank:           rank.RankNumber(),
		MaxRange:       float64(rank.MaxRange),

		ManaCost: manaCost(rank),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      rank.GCD(),
				CastTime: rank.CastTime(),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: heal.Coeff(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			amount := heal.Roll(sim) + paladin.flashOfLightBonusHealing
			if target.HasActiveAuraWithTag(buffs.GreaterBlessingOfLightCategory) {
				amount += blessingOfLightFlashOfLightBonus
			}
			spell.CalcAndDealHealing(sim, target, amount, spell.OutcomeHealingCrit)
		},
	})
}
