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
// Heals a friendly target for 1580.
func (paladin *Paladin) registerHolyLight(_ int32, rank *spelldata.Spell) {
	heal := rank.HealEffect()

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.ID},
		SpellSchool:    rank.SpellSchool(),
		DefenseType:    rank.DefenseTypeCore(),
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskHolyLight,
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
			amount := heal.Roll(sim)
			if target.HasActiveAuraWithTag(buffs.GreaterBlessingOfLightCategory) {
				amount += blessingOfLightHolyLightBonus
			}
			spell.CalcAndDealHealing(sim, target, amount, spell.OutcomeHealingCrit)
		},
	})
}
