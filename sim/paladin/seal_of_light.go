package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The heal each rank's judgement grants attackers. It carries no rank subtext, so the pairing is
// by hand; the heal the seal itself fires is the first spell its tooltip names.
var judgementOfLightHealIDs = map[int32]int32{1: 20267, 2: 20341, 3: 20342, 4: 20343}

// Seal of Light
// https://www.wowhead.com/forever/spell=20349
//
// Fills the Paladin with divine light for 30 sec, giving each melee attack a chance to heal the
// Paladin for 94. Only one Seal can be active on the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy for 40 sec, granting melee attacks made against
// the judged enemy a chance of healing the attacker for 61. Your melee strikes will refresh the
// spell's duration. Only one Judgement per Paladin can be active at any one time.
func (paladin *Paladin) registerSealOfLight(_ int32, rank *spelldata.Spell) {
	judgementID := int32(rank.EffectN(2).BaseValue())
	healID := judgementOfLightHealIDs[rank.RankNumber()]
	judgementAuras := paladin.newJudgementAuras(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataJudgementOfLightAuraRank(target, &paladin.Unit, buffs.JudgementRank{
			SpellID: judgementID,
			Rank:    rank.RankNumber(),
			HealID:  healID,
		})
	})

	// Melee in SpellCategories and Always Hit: the debuff lands without a roll.
	judgementConfig := spelldata.SpellConfig(&paladin.Unit, spelldata.MustFind(judgementID), spelldata.Melee(core.ProcMaskEmpty), spelldata.RemoveFlags(core.SpellFlagAPL))
	judgementConfig.ClassSpellMask = SpellMaskJudgementOfLight
	judgementConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
		judgementAuras.Get(target).Activate(sim)
	}
	judgementConfig.RelatedAuraArrays = judgementAuras.ToMap()
	judgement := paladin.RegisterSpell(judgementConfig)

	procRank := rank.Refs()[0]
	heal := procRank.HealEffect().Average()
	procConfig := spelldata.SpellConfig(&paladin.Unit, procRank, spelldata.Magic(core.ProcMaskSpellHealing), spelldata.Proc(),
		spelldata.Flags(core.SpellFlagProc)) // The heals lack Not a Proc.
	procConfig.ClassSpellMask = SpellMaskSealOfLightProc
	procConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealHealing(sim, target, heal, spell.OutcomeAlwaysHit)
	}
	procSpell := paladin.RegisterSpell(procConfig)

	// The row hears melee swings and abilities; its rate is an override.
	trigger := spelldata.ProcTrigger(&paladin.Character, rank, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		procSpell.Cast(sim, &paladin.Unit)
	})
	trigger.Name = rank.RankedLabel(&paladin.Unit)
	trigger.MetricsActionID = core.ActionID{SpellID: rank.ID}
	trigger.Duration = sealDuration
	aura := paladin.makeSealExclusive(paladin.MakeProcTriggerAura(trigger))

	paladin.registerSealSpell(&sealConfig{
		rank:      rank,
		classMask: SpellMaskSealOfLight,
		aura:      aura,
		judgement: judgement,
	})
}
