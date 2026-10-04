package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The mana each rank's judgement grants attackers. It carries no rank subtext, so the pairing is
// by hand; the restore the seal itself fires is the first spell its tooltip names.
var judgementOfWisdomManaIDs = map[int32]int32{1: 20268, 2: 20352, 3: 20353}

// Seal of Wisdom
// https://www.wowhead.com/forever/spell=20357
//
// Fills the Paladin with divine wisdom for 30 sec, giving each melee attack a chance to restore
// mana to the Paladin. Only one Seal can be active on the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy for 40 sec, granting attacks and spells used
// against the judged enemy a chance to restore mana to the attacker. Your melee strikes will
// refresh the spell's duration. Only one Judgement per Paladin can be active at any one time.
//
// The seal's second effect names its judgement (the row has no effect at the game files' index 1).
func (paladin *Paladin) registerSealOfWisdom(_ int32, rank *spelldata.Spell) {
	judgementID := int32(rank.EffectN(2).BaseValue())
	manaRank := spellData.SealOfWisdomTriggered.ByID(judgementOfWisdomManaIDs[rank.RankNumber()])
	judgementAuras := paladin.newJudgementAuras(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataJudgementOfWisdomAuraRank(target, buffs.JudgementRank{
			SpellID: judgementID,
			Rank:    rank.RankNumber(),
			Value:   manaRank.EnergizeEffect().Average(),
		})
	})

	// Melee in SpellCategories and Always Hit: the debuff lands without a roll.
	judgement := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: judgementID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagMeleeMetrics,
		ClassSpellMask: SpellMaskJudgementOfWisdom,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			judgementAuras.Get(target).Activate(sim)
		},

		RelatedAuraArrays: judgementAuras.ToMap(),
	})

	procRank := rank.Refs()[0]
	mana := procRank.EnergizeEffect().Average()
	procID := core.ActionID{SpellID: procRank.ID}
	manaMetrics := paladin.NewManaMetrics(procID)
	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       procID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagHelpful | core.SpellFlagPassiveSpell | core.SpellFlagProc, // The restores lack Not a Proc.
		ClassSpellMask: SpellMaskSealOfWisdomProc,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.Unit.AddMana(sim, mana, manaMetrics)
		},
	})

	aura := paladin.makeSealExclusive(paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:            sealLabel("Seal of Wisdom", paladin, rank),
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
		classMask: SpellMaskSealOfWisdom,
		aura:      aura,
		judgement: judgement,
	})
}
