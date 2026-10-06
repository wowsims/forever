package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Seal of Justice
// https://www.wowhead.com/forever/spell=20164
//
// Fills the Paladin with the spirit of justice for 30 sec, giving each melee attack a chance to stun
// for 2 sec. Only one Seal can be active on the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy for 10 sec, preventing them from fleeing. Your
// melee strikes will refresh the spell's duration. Only one Judgement per Paladin can be active at
// any one time.
func (paladin *Paladin) registerSealOfJustice() {
	rank := spellData.SealOfJustice.Highest()
	judgementRank := spellData.SealOfJusticeTriggered.ByID(int32(rank.EffectN(2).BaseValue()))
	judgementAuras := paladin.newJudgementAuras(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Judgement of Justice",
			ActionID: core.ActionID{SpellID: judgementRank.ID},
			Tag:      buffs.JudgementAuraTag,
			Duration: judgementRank.Duration(),
		})
	})

	// Melee in SpellCategories and Always Hit: the debuff lands without a roll.
	judgementConfig := spelldata.SpellConfig(&paladin.Unit, judgementRank, spelldata.Melee(core.ProcMaskEmpty), spelldata.RemoveFlags(core.SpellFlagAPL))
	judgementConfig.ClassSpellMask = SpellMaskJudgementOfJustice
	judgementConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
		judgementAuras.Get(target).Activate(sim)
	}
	judgementConfig.RelatedAuraArrays = judgementAuras.ToMap()
	judgement := paladin.RegisterSpell(judgementConfig)

	aura := paladin.makeSealExclusive(paladin.RegisterAura(core.Aura{
		Label:    rank.RankedLabel(&paladin.Unit),
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: sealDuration,
	}))

	paladin.registerSealSpell(&sealConfig{
		rank:      rank,
		classMask: SpellMaskSealOfJustice,
		aura:      aura,
		judgement: judgement,
		echoID:    echoOfJusticeID,
	})
}
