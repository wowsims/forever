package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (rogue *Rogue) registerCombatTalents() {
	// Tier 1
	rogue.registerImprovedEviscerate()
	rogue.registerImprovedSinisterStrike()
	// Lightning Reflexes NYI

	// Tier 2
	rogue.registerPuncturingWounds()
	// Deflection NYI
	rogue.registerPrecision()

	// Tier 3
	// None in this tier implemented

	// Tier 4
	// Improved Kick NYI
	rogue.registerFlawlessExecution()
	rogue.registerDualWieldSpecialization()

	// Tier 5
	rogue.registerBladeFlurry()
	rogue.registerHackAndSlash()

	// Tier 6
	rogue.registerWeaponExpertise()
	rogue.registerAggression()

	// Tier 7
	rogue.registerAdrenalineRush()
}

func (rogue *Rogue) registerImprovedSinisterStrike() {
	if rogue.Talents.ImprovedSinisterStrike == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.ImprovedSinisterStrike.Rank(rogue.Talents.ImprovedSinisterStrike))
}

func (rogue *Rogue) registerPuncturingWounds() {
	if rogue.Talents.PuncturingWounds == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.PuncturingWounds.Rank(rogue.Talents.PuncturingWounds))

	rogue.puncturingWoundsComboPointMetrics = rogue.NewComboPointMetrics(core.ActionID{SpellID: spellData.PuncturingWounds.Rank(rogue.Talents.PuncturingWounds).ID})
}

func (rogue *Rogue) registerPrecision() {
	if rogue.Talents.Precision == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Precision.Rank(rogue.Talents.Precision))
}

func (rogue *Rogue) registerFlawlessExecution() {
	if !rogue.Talents.FlawlessExecution {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.FlawlessExecution.Rank(1))
}

func (rogue *Rogue) registerDualWieldSpecialization() {
	if rogue.Talents.DualWieldSpecialization == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.DualWieldSpecialization.Rank(rogue.Talents.DualWieldSpecialization))
}

var bladeFlurryRank = spellData.BladeFlurry.Highest()

func (rogue *Rogue) registerBladeFlurry() {
	if !rogue.Talents.BladeFlurry {
		return
	}

	var curDmg float64
	bfHit := rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 22482},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty, // No proc mask, so it won't proc itself.
		Flags:       core.SpellFlagIgnoreResists | core.SpellFlagIgnoreModifiers | core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, curDmg, spell.OutcomeAlwaysHit)
		},
	})

	rogue.BladeFlurryAura = rogue.RegisterAura(spelldata.AuraConfig(bladeFlurryRank))
	spelldata.ParseEffects(&rogue.Character, rogue.BladeFlurryAura, bladeFlurryRank)
	rogue.BladeFlurryAura.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if sim.ActiveTargetCount() < 2 {
			return
		}
		if result.Damage == 0 || !spell.ProcMask.Matches(core.ProcMaskMelee) {
			return
		}
		curDmg = result.Damage
		next := rogue.Env.NextActiveTargetUnit(result.Target)
		bfHit.Cast(sim, next)
		bfHit.SpellMetrics[next.UnitIndex].Casts--
	}

	config := spelldata.SpellConfig(&rogue.Unit, bladeFlurryRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMCD))
	config.ClassSpellMask = RogueSpellBladeFlurry
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BladeFlurryAura.Activate(sim)
	}
	rogue.BladeFlurry = rogue.RegisterSpell(config)

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.BladeFlurry,
		Type:  core.CooldownTypeDPS,
	})
}

func (rogue *Rogue) registerHackAndSlash() {
	if rogue.Talents.HackAndSlash == 0 {
		return
	}

	rank := rogue.Talents.HackAndSlash
	actionID := core.ActionID{SpellID: spellData.HackAndSlash.Rank(rogue.Talents.HackAndSlash).ID}

	// The three branches share A_DUMMY and misc 0, so each is named by its effect index.
	mainHandIs := func(weaponTypes ...proto.WeaponType) bool {
		return rogue.GetProcMaskForTypes(weaponTypes...).Matches(core.ProcMaskMeleeMH)
	}
	offHandIs := func(weaponTypes ...proto.WeaponType) bool {
		return rogue.GetProcMaskForTypes(weaponTypes...).Matches(core.ProcMaskMeleeOH)
	}
	var armorIgnoreOn bool
	var critMask core.ProcMask
	var swordMask core.ProcMask
	readWeapons := func() {
		armorIgnoreOn = mainHandIs(proto.WeaponType_WeaponTypeMace) || offHandIs(proto.WeaponType_WeaponTypeMace)
		swordMask = rogue.GetProcMaskForTypes(proto.WeaponType_WeaponTypeSword, proto.WeaponType_WeaponTypeAxe)
		critMask = rogue.GetProcMaskForTypes(proto.WeaponType_WeaponTypeFist, proto.WeaponType_WeaponTypeDagger)
	}
	readWeapons()

	critAura := rogue.RegisterAura(core.Aura{
		Label:      "Hack and Slash (Fist/Dagger)",
		ActionID:   actionID.WithTag(2),
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(critMask != core.ProcMaskUnknown, core.CharacterBuildPhaseBase, core.CharacterBuildPhaseNone),
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: spellData.HackAndSlash.EffectAt(1).ValueAt(rank),
		ProcMask:   critMask,
	})
	if critMask != core.ProcMaskUnknown {
		core.MakePermanent(critAura)
	}

	armorIgnore := spellData.HackAndSlash.EffectAt(2).FractionAt(rank)
	armorIgnoreAura := rogue.RegisterAura(core.Aura{
		Label:    "Hack and Slash (Mace)",
		ActionID: actionID.WithTag(3),
		Duration: core.NeverExpires,
		OnGain: func(_ *core.Aura, _ *core.Simulation) {
			rogue.AddArmorIgnoreFactor(armorIgnore)
		},
		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			rogue.AddArmorIgnoreFactor(-armorIgnore)
		},
	})
	if armorIgnoreOn {
		core.MakePermanent(armorIgnoreAura)
	}

	setActive := func(sim *core.Simulation, aura *core.Aura, on bool) {
		if on {
			aura.Activate(sim)
		} else {
			aura.Deactivate(sim)
		}
	}

	var extraAttack *core.Spell
	swordAura := rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Hack and Slash (Axe/Sword)",
		ActionID:           actionID.WithTag(1),
		MetricsActionID:    actionID.WithTag(1),
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMelee,
		Outcome:            core.OutcomeLanded,
		ProcChance:         spellData.HackAndSlash.EffectAt(3).FractionAt(rank),
		ICD:                spellData.HackAndSlash.Rank(rank).ICD(),
		TriggerImmediately: true,
		Duration:           core.NeverExpires,
		ExtraCondition: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spell.ProcMask.Matches(swordMask)
		},
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// We were told these don't reset swings anymore, don't need to use MaybeReplaceSwing
			extraAttack.Cast(sim, result.Target)
		},
	}).ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
		config := *rogue.AutoAttacks.MHConfig()
		config.ActionID = config.ActionID.WithTag(actionID.SpellID)
		extraAttack = rogue.GetOrRegisterSpell(config)
		aura.OnReset = func(aura *core.Aura, sim *core.Simulation) {
			if swordMask == core.ProcMaskUnknown {
				aura.Deactivate(sim)
			} else {
				aura.Activate(sim)
			}
		}
	})

	rogue.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
		readWeapons()
		setActive(sim, critAura, critMask != core.ProcMaskUnknown)
		setActive(sim, armorIgnoreAura, armorIgnoreOn)
		setActive(sim, swordAura, swordMask != core.ProcMaskUnknown)
	})
}

// "Reduces the chance for your attacks to be Dodged or Parried by 1%/2%." A_MOD_EXPERTISE
// carries the percent directly.
func (rogue *Rogue) registerWeaponExpertise() {
	if rogue.Talents.WeaponExpertise == 0 {
		return
	}

	rogue.AddStat(stats.ExpertisePercent, spellData.WeaponExpertise.ValueAt(rogue.Talents.WeaponExpertise))
}

func (rogue *Rogue) registerAggression() {
	if rogue.Talents.Aggression == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Aggression.Rank(rogue.Talents.Aggression))
}

var adrenalineRushRank = spellData.AdrenalineRush.Highest()

func (rogue *Rogue) registerAdrenalineRush() {
	if !rogue.Talents.AdrenalineRush {
		return
	}

	config := spelldata.SpellConfig(&rogue.Unit, adrenalineRushRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMCD))
	adrenalineRushAura := rogue.RegisterAura(spelldata.AuraConfig(adrenalineRushRank))
	spelldata.ParseEffects(&rogue.Character, adrenalineRushAura, adrenalineRushRank)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		adrenalineRushAura.Activate(sim)
	}

	rogue.AdrenalineRush = rogue.RegisterSpell(config)

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.AdrenalineRush,
		Type:  core.CooldownTypeDPS,
	})
}
