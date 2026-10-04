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
// Paladin. Only one Seal can be active on the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy for 40 sec, granting melee attacks made
// against the judged enemy a chance of healing the attacker. Your melee strikes will refresh the
// spell's duration. Only one Judgement per Paladin can be active at any one time.
//
// The seal's second effect names its judgement (the row has no effect at the game files' index 1).
func (paladin *Paladin) registerSealOfLight(_ int32, rank *spelldata.Spell) {
	judgementID := int32(rank.EffectN(2).BaseValue())
	healRank := spellData.SealOfLightTriggered.ByID(judgementOfLightHealIDs[rank.RankNumber()])
	judgementAuras := paladin.newJudgementAuras(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataJudgementOfLightAuraRank(target, &paladin.Unit, buffs.JudgementRank{
			SpellID: judgementID,
			Rank:    rank.RankNumber(),
			Value:   healRank.HealEffect().Average(),
		})
	})

	// Melee in SpellCategories and Always Hit: the debuff lands without a roll.
	judgement := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: judgementID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagMeleeMetrics,
		ClassSpellMask: SpellMaskJudgementOfLight,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			judgementAuras.Get(target).Activate(sim)
		},

		RelatedAuraArrays: judgementAuras.ToMap(),
	})

	procRank := rank.Refs()[0]
	heal := procRank.HealEffect().Average()
	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: procRank.ID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagHelpful | core.SpellFlagPassiveSpell | core.SpellFlagProc, // The heals lack Not a Proc.
		ClassSpellMask: SpellMaskSealOfLightProc,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, heal, spell.OutcomeAlwaysHit)
		},
	})

	aura := paladin.makeSealExclusive(paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:            sealLabel("Seal of Light", paladin, rank),
		ActionID:        core.ActionID{SpellID: rank.ID},
		MetricsActionID: core.ActionID{SpellID: rank.ID},
		Duration:        sealDuration,
		Callback:        core.CallbackOnSpellHitDealt,
		ProcMask:        core.ProcMaskMelee,
		Outcome:         core.OutcomeLanded,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			procSpell.Cast(sim, &paladin.Unit)
		},
	}))

	paladin.registerSealSpell(&sealConfig{
		rank:      rank,
		classMask: SpellMaskSealOfLight,
		aura:      aura,
		judgement: judgement,
	})
}
