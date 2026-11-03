package rogue

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (rogue *Rogue) registerSubtletyTalents() {
	// Tier 1
	rogue.registerCamouflage()
	// Master of Deception NYI
	rogue.registerOpportunity()

	// Tier 2
	rogue.registerSetup()
	rogue.registerElusiveness()
	// Dirty Tricks NYI
	rogue.registerImprovedAmbush()

	// Tier 3
	rogue.registerInitiative()
	rogue.registerGhostlyStrike()
	// Improved Distract NYI

	// Tier 4
	// Heightened Senses NYI
	rogue.registerQuickGetaway()
	rogue.registerPremeditation()
	rogue.registerSerratedBlades()

	// Tier 5
	rogue.registerDirtyDeeds()
	rogue.registerPreparation()
	rogue.registerHemorrhage()

	rogue.registerBootKnife()
	rogue.registerDeathMark()
}

func (rogue *Rogue) registerOpportunity() {
	rank := rogue.Talents.Opportunity
	if rank == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellBackstab | RogueSpellGarrote | RogueSpellAmbush | RogueSpellMutilateHit,
		FloatValue: 0.04 * float64(rank),
	})
}

func (rogue *Rogue) registerSetup() {
	rank := rogue.Talents.Setup
	if rank == 0 {
		return
	}

	metrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 13983})
	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Setup",
		ActionID:           core.ActionID{SpellID: 13983},
		Callback:           core.CallbackOnSpellHitTaken,
		Outcome:            core.OutcomeDodge | core.OutcomeMiss,
		ProcChance:         0.15 * float64(rank),
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Outcome.Matches(core.OutcomeMiss) && spell.DefenseType != core.DefenseTypeMagic {
				return
			}
			rogue.AddComboPoints(sim, 1, metrics)
		},
	})
}

func (rogue *Rogue) registerCamouflage() {
	rank := rogue.Talents.Camouflage
	if rank == 0 {
		return
	}

	rogue.camouflageMovementSpeed = 0.05 * float64(rank)
	rogue.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		ClassMask: RogueSpellStealth,
		TimeValue: -2 * time.Second * time.Duration(rank),
	})
}

func (rogue *Rogue) registerQuickGetaway() {
	rank := rogue.Talents.QuickGetaway
	if rank == 0 {
		return
	}

	speedAura := rogue.RegisterAura(core.Aura{
		Label:    "Quick Getaway",
		ActionID: core.ActionID{SpellID: 14084},
		Duration: 3 * time.Second,
		OnGain: func(_ *core.Aura, sim *core.Simulation) {
			rogue.MultiplyMovementSpeed(sim, 1+0.15*float64(rank))
		},
		OnExpire: func(_ *core.Aura, sim *core.Simulation) {
			rogue.MultiplyMovementSpeed(sim, 1/(1+0.15*float64(rank)))
		},
	})
	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Quick Getaway - Sap",
		ActionID:           core.ActionID{SpellID: 14084},
		Callback:           core.CallbackOnCastComplete,
		ProcChance:         1,
		TriggerImmediately: true,
		ExtraCondition: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spellData.Sap.Contains(spell)
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			speedAura.Activate(sim)
		},
	})
}

func (rogue *Rogue) registerBootKnife() {
	rank := rogue.Talents.BootKnife
	if rank == 0 {
		return
	}

	metrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 1310728})
	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Boot Knife",
		ActionID:           core.ActionID{SpellID: 1310728},
		Callback:           core.CallbackOnSpellHitDealt,
		ClassSpellMask:     RogueSpellShank,
		Outcome:            core.OutcomeLanded,
		ProcChance:         0.2 * float64(rank),
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			rogue.AddComboPoints(sim, 2, metrics)
		},
	})
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
	rogue.GhostlyStrikeAura = rogue.RegisterAura(core.Aura{
		Label:    "Ghostly Strike",
		ActionID: core.ActionID{SpellID: ghostlyStrikeRank.ID},
		Duration: 7 * time.Second,
	})

	// Dodge aura not implemented; maybe if rogue tanking comes back

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) // Data suggests un-normalized

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if result.Landed() {
			rogue.GhostlyStrikeAura.Activate(sim)
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
		hemoAura.AttachDDBC(DDBC_Garrote, DDBC_Total, &rogue.AttackTables, func(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
			if spell.ClassSpellMask&RogueSpellGarrote != 0 {
				return 1.15
			}
			return 1
		})
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
