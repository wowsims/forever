package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (rogue *Rogue) registerSubtletyTalents() {
	// Tier 1
	rogue.registerSetup()
	// Master of Deception NYI
	rogue.registerOpportunity()

	// Tier 2
	rogue.registerDirtyDeeds()
	// Camouflage NYI
	rogue.registerImprovedAmbush()
	// Dirty Tricks NYI

	// Tier 3
	rogue.registerInitiative()
	rogue.registerPremeditation()
	rogue.registerSerratedBlades()

	// Tier 4
	rogue.registerElusiveness()
	rogue.registerGhostlyStrike()
	// Quick Getaway NYI

	// Tier 5
	// Heightened Senses NYI
	rogue.registerPreparation()
	rogue.registerHemorrhage()

	// Tier 6
	// Boot Knife modeled in shank.go

	// Tier 7
	rogue.registerDeathMark()
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
		func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			rogue.AddComboPoints(sim, result.Target, 1, initMetrics)
		})
	trigger.ClassSpellMask = RogueSpellGarrote | RogueSpellAmbush
	trigger.TriggerImmediately = true
	trigger.Outcome = core.OutcomeLanded
	trigger.RequireDamageDealt = false
	rogue.MakeProcTriggerAura(trigger)
}

var ghostlyStrikeRank = spellData.GhostlyStrike.Highest()
var ghostlyStrikeTriggered = spellData.GhostlyStrikeTriggered.Highest()

func (rogue *Rogue) registerGhostlyStrike() {
	if !rogue.Talents.GhostlyStrike {
		return
	}

	config := spelldata.SpellConfig(&rogue.Unit, ghostlyStrikeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder
	config.ClassSpellMask = RogueSpellGhostlyStrike
	config.DamageMultiplier = core.TernaryFloat64(rogue.HasDagger(core.MainHand), ghostlyStrikeRank.EffectN(4).Percent(), ghostlyStrikeRank.EffectN(1).Percent())

	// The dodge buff is the row's own aura: its duration and the dodge its second effect adds.
	dodgeBuff := rogue.RegisterAura(spelldata.AuraConfig(ghostlyStrikeRank))
	spelldata.ParseEffects(&rogue.Character, dodgeBuff, ghostlyStrikeRank, spelldata.Effects(2))
	config.RelatedSelfBuff = dodgeBuff

	// The strike's last effect triggers a buff that lets the next Ambush or Garrote skip Stealth.
	rogue.GhostlyStrikeAura = rogue.RegisterAura(spelldata.AuraConfig(ghostlyStrikeTriggered, spelldata.Label("Ghostly Strike - Next Ambush or Garrote"))).
		ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
			rogue.Ambush.CastRequirement.CasterForm = true
			rogue.Garrote.CastRequirement.CasterForm = true
		}).
		ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
			rogue.Ambush.CastRequirement.CasterForm = false
			rogue.Garrote.CastRequirement.CasterForm = false
		})
	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Ghostly Strike - Cast Watcher",
		ActionID:           core.ActionID{SpellID: ghostlyStrikeTriggered.ID},
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     RogueSpellAmbush | RogueSpellGarrote,
		TriggerImmediately: true,
		ProcChance:         1,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			rogue.GhostlyStrikeAura.Deactivate(sim)
		},
	})

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		// The row's E_NORMALIZED_WEAPON_DMG effect normalizes the weapon damage, as on Hemorrhage.
		baseDamage := spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		// An avoided strike grants no buff, as it grants no combo point.
		if result.Landed() {
			dodgeBuff.Activate(sim)
			rogue.GhostlyStrikeAura.Activate(sim)
			rogue.AddComboPoints(sim, target, int32(ghostlyStrikeRank.ProcEnergizeEffect().Roll(sim)), spell.ComboPointMetrics())
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
		if rogue.Riposte != nil {
			rogue.Riposte.CD.Set(0)
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
			rogue.AddComboPoints(sim, target, int32(hemorrhageRank.ProcEnergizeEffect().Roll(sim)), spell.ComboPointMetrics())
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
		rogue.AddComboPoints(sim, target, int32(premeditationRank.ProcEnergizeEffect().Roll(sim)), premedMetrics)
	}

	rogue.Premeditation = rogue.GetOrRegisterSpell(config)

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.Premeditation,
		Type:  core.CooldownTypeDPS,
	})
}

var setupTriggered = spellData.SetupTriggered.Highest()

func (rogue *Rogue) registerSetup() {
	if rogue.Talents.Setup == 0 {
		return
	}

	setupRank := spellData.Setup.Rank(rogue.Talents.Setup)
	setupMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: setupRank.ID})

	trigger := spelldata.ProcTrigger(&rogue.Character, setupRank,
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			// The dodged attack's result is on the rogue; the point goes on the attacker, the rogue's target at the
			// dodge, even if the target changes before this deferred handler runs.
			rogue.AddComboPoints(sim, spell.Unit, int32(setupTriggered.ProcEnergizeEffect().Roll(sim)), setupMetrics)
		})
	// Only the dodge half is modelled. The fully resisted spell half is not.
	trigger.Outcome = core.OutcomeDodge
	trigger.RequireDamageDealt = false
	// The tooltip limits it to your target's attacks.
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.Unit == rogue.CurrentTarget
	}
	rogue.MakeProcTriggerAura(trigger)
}
