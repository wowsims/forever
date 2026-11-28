package rogue

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The rows ItemSetSpell names for each set's thresholds.
var (
	bonescytheHeadRush = spelldata.MustFind(28812) // 4pc: Head Rush, the proc
	bonescytheThreat   = spelldata.MustFind(28811) // 6pc: less threat
	bonescytheFlaw     = spelldata.MustFind(28814) // 8pc: the chance per combo point on Eviscerate
	revealedFlaw       = spelldata.MustFind(28815) // 8pc: Revealed Flaw, the crit it grants
)

// Tier 3
var ItemSetBonescytheArmor = core.NewItemSet(core.ItemSet{
	ID:    524,
	Name:  "Bonescythe Armor",
	Slots: shared.TierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		// TODO: 2pc (28816) Invigorate: normal melee swings have a chance to heal the wearer for 100.
		// TODO: Ingame test needed -- whether only Backstab, Sinister Strike and Hemorrhage crits restore energy.
		4: applyBonescytheHeadRush,
		6: shared.ParsedSetBonus(bonescytheThreat),
		// TODO: Ingame test needed -- the 5% per combo point.
		8: applyBonescytheRevealedFlaw,
	},
})

const bonescytheStrikes = RogueSpellBackstab | RogueSpellSinisterStrike | RogueSpellHemorrhage

// "Your Backstab, Sinister Strike, and Hemorrhage critical hits cause you to regain 5 energy."
//
// The row hears every landed melee ability, and its class flags name Eviscerate, Shank and Mutilate too. The
// tooltip's three spells and its crit are set here.
func applyBonescytheHeadRush(agent core.Agent, setBonusAura *core.Aura) {
	rogue := agent.(RogueAgent).GetRogue()
	trigger := spelldata.ProcTrigger(&rogue.Character, bonescytheHeadRush, nil)
	trigger.Handler = shared.FiredSpellHandler(&rogue.Character, core.ActionID{SpellID: bonescytheHeadRush.FiredSpell().ID}, bonescytheHeadRush)
	trigger.Name = "Bonescythe Armor - 4PC"
	trigger.ClassSpellMask = bonescytheStrikes
	trigger.Outcome = core.OutcomeCrit

	setBonusAura.AttachProcTrigger(trigger)
}

// "Your Eviscerate has a chance per combo point to reveal a flaw in your opponent's armor, granting a 100%
// critical hit chance for your next Backstab, Sinister Strike, or Hemorrhage."
//
// 28815's cooldown does not limit the roll. The buff's class flags name Shank, Mutilate and the Season of Discovery
// strikes too; the crit and the charge are the tooltip's three here.
func applyBonescytheRevealedFlaw(agent core.Agent, setBonusAura *core.Aura) {
	rogue := agent.(RogueAgent).GetRogue()
	buff := rogue.RegisterAura(spelldata.AuraConfig(revealedFlaw)).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ClassMask:  bonescytheStrikes,
		FloatValue: revealedFlaw.EffectN(1).BaseValue(),
	})
	buff.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, aura.MaxStacks)
	})
	spelldata.AttachChargeSpender(&rogue.Character, buff, revealedFlaw, "Revealed Flaw - Spend",
		func(_ *core.Character, trigger *core.ProcTrigger) {
			trigger.ClassSpellMask = bonescytheStrikes
		})

	// Eviscerate spends its combo points before its hit is dealt, so they are read as it applies.
	var comboPoints int32
	eviscerate := bonescytheFlaw.EffectN(1).ClassFlags
	setBonusAura.AttachProcTrigger(core.ProcTrigger{
		Name:               "Bonescythe Armor - 8PC Combo Points",
		Callback:           core.CallbackOnApplyEffects,
		ClassFlags:         eviscerate,
		TriggerImmediately: true,
		Handler: func(_ *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			comboPoints = rogue.ComboPoints(result.Target)
		},
	})
	chancePerPoint := float64(bonescytheFlaw.EffectN(1).PointsPerResource) / 100
	setBonusAura.AttachProcTrigger(core.ProcTrigger{
		Name:       "Bonescythe Armor - 8PC",
		ActionID:   core.ActionID{SpellID: bonescytheFlaw.ID},
		Callback:   core.CallbackOnSpellHitDealt,
		ClassFlags: eviscerate,
		Outcome:    core.OutcomeLanded,
		ExtraCondition: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
			return sim.Proc(chancePerPoint*float64(comboPoints), "Revealed Flaw")
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Activate(sim)
		},
	}).ExposeToAPL(bonescytheFlaw.ID)
}
