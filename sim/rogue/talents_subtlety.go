package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (rogue *Rogue) registerSubtletyTalents() {
	// Tier 1
	// Camouflage NYI
	// Master of Deception NYI
	rogue.registerOpportunity()

	// Tier 2
	// Setup NYI
	rogue.registerElusiveness()
	// Dirty Tricks NYI
	rogue.registerImprovedAmbush()

	// Tier 3
	rogue.registerInitiative()
	rogue.registerGhostlyStrike()
	// Improved Distract NYI

	// Tier 4
	// Heightened Senses NYI
	rogue.registerPremeditation()
	rogue.registerSerratedBlades()

	// Tier 5
	rogue.registerDirtyDeeds()
	rogue.registerPreparation()
	rogue.registerHemorrhage()

	// Tier 6
	rogue.registerQuietus()
	rogue.registerCutthroat()

	// Tier 7
	rogue.registerThousandCuts()
}

func (rogue *Rogue) registerOpportunity() {
	if rogue.Talents.Opportunity == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Opportunity.Rank(rogue.Talents.Opportunity))
}

func (rogue *Rogue) registerInitiative() {
	if rogue.Talents.Initiative == 0 {
		return
	}

	initiativeRank := spellData.Initiative.Rank(rogue.Talents.Initiative)
	initMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: initiativeRank.ID})

	trigger := spelldata.ProcTrigger(&rogue.Character, initiativeRank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			rogue.AddComboPoints(sim, 1, initMetrics)
		})
	trigger.ClassSpellMask = RogueSpellGarrote | RogueSpellAmbush
	trigger.TriggerImmediately = true
	trigger.Outcome = core.OutcomeLanded
	trigger.RequireDamageDealt = false
	rogue.MakeProcTriggerAura(trigger)
}

var ghostlyStrikeRank = spellData.GhostlyStrike.Highest()

func (rogue *Rogue) registerGhostlyStrike() {
	if !rogue.Talents.GhostlyStrike {
		return
	}

	config := spelldata.SpellConfig(&rogue.Unit, ghostlyStrikeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder
	config.ClassSpellMask = RogueSpellGhostlyStrike
	config.DamageMultiplier = core.TernaryFloat64(rogue.HasDagger(core.MainHand), ghostlyStrikeRank.EffectN(4).Percent(), ghostlyStrikeRank.EffectN(1).Percent())

	// Dodge aura not implemented; maybe if rogue tanking comes back

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) // Data suggests un-normalized

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if result.Landed() {
			rogue.AddComboPoints(sim, int32(ghostlyStrikeRank.ProcEnergizeEffect().Roll(sim)), spell.ComboPointMetrics())
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.GhostlyStrike = rogue.RegisterSpell(config)

	rogue.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(s *core.Simulation, slot proto.ItemSlot) {
		rogue.GhostlyStrike.DamageMultiplier = core.TernaryFloat64(rogue.HasDagger(core.MainHand), ghostlyStrikeRank.EffectN(4).Percent(), ghostlyStrikeRank.EffectN(1).Percent())
	})
}

func (rogue *Rogue) registerImprovedAmbush() {
	if rogue.Talents.ImprovedAmbush == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.ImprovedAmbush.Rank(rogue.Talents.ImprovedAmbush))
}

func (rogue *Rogue) registerElusiveness() {
	if rogue.Talents.Elusiveness == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Elusiveness.Rank(rogue.Talents.Elusiveness))
}

func (rogue *Rogue) registerSerratedBlades() {
	if rogue.Talents.SerratedBlades == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.SerratedBlades.Rank(rogue.Talents.SerratedBlades))
}

func (rogue *Rogue) registerPreparation() {
	if !rogue.Talents.Preparation {
		return
	}

	// Prep purely runs a server-side script, so no client data to reference.

	config := spelldata.SpellConfig(&rogue.Unit, spellData.Preparation.Highest(), spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMCD))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if rogue.ColdBlood != nil {
			rogue.ColdBlood.CD.Set(0)
		}
		if rogue.Premeditation != nil {
			rogue.Premeditation.CD.Set(0)
		}
		if rogue.Vanish != nil {
			rogue.Vanish.CD.Set(0)
		}
		if rogue.GhostlyStrike != nil {
			rogue.GhostlyStrike.CD.Set(0)
		}
		if rogue.BladeFlurry != nil {
			rogue.BladeFlurry.CD.Set(0)
		}
		if rogue.AdrenalineRush != nil {
			rogue.AdrenalineRush.CD.Set(0)
		}
	}

	rogue.Preparation = rogue.RegisterSpell(config)

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.Preparation,
		Type:  core.CooldownTypeDPS,
	})
}

func (rogue *Rogue) registerDirtyDeeds() {
	if rogue.Talents.DirtyDeeds == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.DirtyDeeds.Rank(rogue.Talents.DirtyDeeds))
}

var hemorrhageRank = spellData.Hemorrhage.Highest()

func (rogue *Rogue) registerHemorrhage() {
	if !rogue.Talents.Hemorrhage {
		return
	}

	config := spelldata.SpellConfig(&rogue.Unit, hemorrhageRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder
	config.ClassSpellMask = RogueSpellHemorrhage
	config.DamageMultiplier = core.TernaryFloat64(rogue.HasDagger(core.MainHand), hemorrhageRank.EffectN(5).Percent(), hemorrhageRank.EffectN(4).Percent())

	hemoAuras := rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		hemoAura := target.GetOrRegisterAura(spelldata.AuraConfig(hemorrhageRank))
		spelldata.ParseEffects(nil, hemoAura, spellData.Hemorrhage.Highest(),
			spelldata.FromCaster(&rogue.Unit, DDBC_Hemorrhage, DDBC_Total))
		return hemoAura
	})

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := rogue.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		if result.Landed() {
			hemoAuras.Get(target).Activate(sim)
			rogue.AddComboPoints(sim, int32(hemorrhageRank.ProcEnergizeEffect().Roll(sim)), spell.ComboPointMetrics())
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.Hemorrhage = rogue.GetOrRegisterSpell(config)

	rogue.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(s *core.Simulation, slot proto.ItemSlot) {
		rogue.Hemorrhage.DamageMultiplier = core.TernaryFloat64(rogue.HasDagger(core.MainHand), hemorrhageRank.EffectN(5).Percent(), hemorrhageRank.EffectN(4).Percent())
	})
}

var premeditationRank = spellData.Premeditation.Highest()

func (rogue *Rogue) registerPremeditation() {
	if !rogue.Talents.Premeditation {
		return
	}

	config := spelldata.SpellConfig(&rogue.Unit, premeditationRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMCD))
	premedMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: premeditationRank.ID})

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.AddComboPoints(sim, int32(premeditationRank.ProcEnergizeEffect().Roll(sim)), premedMetrics)
	}

	rogue.Premeditation = rogue.GetOrRegisterSpell(config)

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.Premeditation,
		Type:  core.CooldownTypeDPS,
	})
}

func (rogue *Rogue) registerQuietus() {
	if rogue.Talents.Quietus == 0 {
		return
	}

	// The talent's effects are dummy auras: effect 1 is the damage bonus, effect 2 the health threshold.
	rank := spellData.Quietus.Rank(rogue.Talents.Quietus)
	threshold := int32(spellData.Quietus.EffectAt(2).ValueAt(rogue.Talents.Quietus))

	executeBonus := rogue.RegisterAura(core.Aura{
		Label:    "Quietus Damage Bonus",
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: core.NeverExpires,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellQuietus,
		FloatValue: spellData.Quietus.EffectAt(1).FractionAt(rogue.Talents.Quietus),
	})

	rogue.RegisterResetEffect(func(sim *core.Simulation) {
		executeBonus.Deactivate(sim)
		sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
			if isExecute <= threshold && !executeBonus.IsActive() {
				executeBonus.Activate(sim)
			}
		})
	})
}

func (rogue *Rogue) registerCutthroat() {
	if rogue.Talents.Cutthroat == 0 {
		return
	}

	cutthroatTrigger := spellData.CutthroatTriggered.Highest()
	rogue.CutthroatAura = rogue.RegisterAura(spelldata.AuraConfig(cutthroatTrigger)).
		ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
			rogue.Ambush.CastRequirement.CasterForm = true
		}).
		ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
			rogue.Ambush.CastRequirement.CasterForm = false
		})

	cutthroatRank := spellData.Cutthroat.Rank(rogue.Talents.Cutthroat)
	trigger := spelldata.ProcTrigger(&rogue.Character, cutthroatRank, func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		rogue.CutthroatAura.Activate(sim)
	})
	trigger.Name = "Cutthroat Trigger"
	trigger.TriggerImmediately = true
	// trigger.RequireDamageDealt = true - Unknown
	trigger.ClassSpellMask = RogueSpellBackstab
	rogue.MakeProcTriggerAura(trigger)

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Cutthroat - Cast Watcher",
		ActionID:           core.ActionID{SpellID: cutthroatRank.ID},
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     RogueSpellAmbush,
		TriggerImmediately: true,
		ProcChance:         1,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			rogue.CutthroatAura.Deactivate(sim)
		},
	})
}

var thousandCutsRank = spellData.ThousandCuts.Highest()
var thousandCutsTriggered = spellData.ThousandCutsTriggered.Highest()

func (rogue *Rogue) registerThousandCuts() {
	if !rogue.Talents.ThousandCuts {
		return
	}

	aura := rogue.RegisterAura(spelldata.AuraConfig(thousandCutsTriggered))
	spelldata.ParseEffects(&rogue.Character, aura, thousandCutsTriggered)

	trigger := spelldata.ProcTrigger(&rogue.Character, thousandCutsRank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.Activate(sim)
			aura.AddStack(sim)
		})
	trigger.Name = "Thousand Cuts - Trigger"
	trigger.ProcMask = core.ProcMaskUnknown
	trigger.ClassSpellMask = RogueSpellRupture
	trigger.TriggerImmediately = true
	rogue.MakeProcTriggerAura(trigger)

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Thousand Cuts - Cast Watcher",
		ActionID:           core.ActionID{SpellID: thousandCutsRank.ID},
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     RogueSpellHemorrhage | RogueSpellBackstab,
		TriggerImmediately: true,
		ProcChance:         1,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.Deactivate(sim)
		},
	})
}
