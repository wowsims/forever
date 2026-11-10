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
// Fills the Paladin with divine wisdom for 30 sec, giving each melee attack a chance to restore 90 of
// the Paladin's mana. Only one Seal can be active on the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy for 40 sec, granting attacks and spells used
// against the judged enemy a chance to restore 59 mana to the attacker. Your melee strikes will
// refresh the spell's duration. Only one Judgement per Paladin can be active at any one time.
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
	judgementConfig := spelldata.SpellConfig(&paladin.Unit, spelldata.MustFind(judgementID), spelldata.Melee(core.ProcMaskEmpty), spelldata.RemoveFlags(core.SpellFlagAPL))
	judgementConfig.ClassSpellMask = SpellMaskJudgementOfWisdom
	judgementConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
		judgementAuras.Get(target).Activate(sim)
	}
	judgementConfig.RelatedAuraArrays = judgementAuras.ToMap()
	judgement := paladin.RegisterSpell(judgementConfig)

	procRank := rank.Refs()[0]
	mana := procRank.EnergizeEffect().Average()
	procID := core.ActionID{SpellID: procRank.ID}
	manaMetrics := paladin.NewManaMetrics(procID)
	procConfig := spelldata.SpellConfig(&paladin.Unit, procRank, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc(),
		spelldata.Flags(core.SpellFlagProc)) // The restores lack Not a Proc.
	procConfig.ClassSpellMask = SpellMaskSealOfWisdomProc
	procConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.Unit.AddMana(sim, mana, manaMetrics)
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
		classMask: SpellMaskSealOfWisdom,
		aura:      aura,
		judgement: judgement,
		echoID:    echoOfWisdomID,
		// A stored copy rolls the seal's own rate for a main-hand swing.
		echo: func(sim *core.Simulation, _ *core.Unit) {
			if trigger.DPM.Proc(sim, core.ProcMaskMeleeMHAuto, "Seal of Wisdom") {
				procSpell.Cast(sim, &paladin.Unit)
			}
		},
	})
}
