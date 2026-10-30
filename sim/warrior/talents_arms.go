package warrior

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warrior *Warrior) registerArmsTalents() {
	// Tier 1
	warrior.registerImprovedHeroicStrike()
	warrior.registerDeflection()
	warrior.registerImprovedRend()

	// Tier 2
	// Improved Charge: charge.go
	// Improved Tactical Mastery: stances.go
	warrior.registerImprovedOverpower()

	// Tier 3
	warrior.registerAngerManagement()
	warrior.registerDeepWounds()

	// Tier 4
	warrior.registerSpearingStrike()
	warrior.registerTwoHandedWeaponSpecialization()
	warrior.registerImpale()

	// Tier 5
	warrior.registerBloodthrill()
	warrior.registerSweepingStrikes()
	warrior.registerWeaponmaster()

	// Tier 6
	warrior.registerImprovedSlam()
	warrior.registerImprovedHamstring()

	// Tier 7
	warrior.registerMortalStrike()
}

/*
 * Arms
 */
func (warrior *Warrior) registerImprovedHeroicStrike() {
	if warrior.Talents.ImprovedHeroicStrike == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character,
		spellData.ImprovedHeroicStrike.Rank(warrior.Talents.ImprovedHeroicStrike))
}
func (warrior *Warrior) registerDeflection() {
	if warrior.Talents.Deflection == 0 {
		return
	}

	// The row states parry as a percentage, which the parse adds to ParryPercent on top of the base
	// chance.
	spelldata.ParseStatic(&warrior.Character, spellData.Deflection.Rank(warrior.Talents.Deflection))
}

func (warrior *Warrior) registerImprovedRend() {
	if warrior.Talents.ImprovedRend == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedRend.Rank(warrior.Talents.ImprovedRend))
}

func (warrior *Warrior) registerImprovedOverpower() {
	if warrior.Talents.ImprovedOverpower == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character,
		spellData.ImprovedOverpower.Rank(warrior.Talents.ImprovedOverpower))
}

var angerManagementRank = spellData.AngerManagement.Highest()

// The rage regen aura on effect 1 states its amount on the rage bar's tenths, per five seconds as
// every A_MOD_POWER_REGEN does. The game files state no tick for it, so the rage arrives as one
// gain every five seconds while in combat. The tooltip's "$m2 Rage every $m3 sec" reads the two
// dummies instead, 1 rage every 3 seconds.
var angerManagementRage = angerManagementRank.EffectN(1).Tenths()

const angerManagementPeriod = 5 * time.Second

func (warrior *Warrior) registerAngerManagement() {
	if !warrior.Talents.AngerManagement {
		return
	}

	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: angerManagementRank.ID})

	warrior.RegisterResetEffect(func(sim *core.Simulation) {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period: angerManagementPeriod,
			OnAction: func(sim *core.Simulation) {
				if sim.CurrentTime > 0 {
					warrior.AddRage(sim, angerManagementRage, rageMetrics)
				}
			},
		})
	})
}

var deepWoundsBleed = spellData.DeepWoundsTriggered.ByID(412609)

func (warrior *Warrior) registerDeepWounds() {
	if warrior.Talents.DeepWounds == 0 {
		return
	}

	share := spellData.DeepWounds.FractionAt(warrior.Talents.DeepWounds)
	tick := deepWoundsBleed.EffectN(1)

	// A crit casts the bleed, but it does not take spelldata.Proc(): that marks the spell passive,
	// and the metrics aggregator counts no cast for a passive spell, while the sim reports every
	// application as a cast.
	config := spelldata.SpellConfig(&warrior.Unit, deepWoundsBleed,
		spelldata.Flags(core.SpellFlagNoOnCastComplete|core.SpellFlagIgnoreResists|core.SpellFlagProc)) // 12162 and 412609 lack Not a Proc.
	config.ProcMask = core.ProcMaskEmpty

	// The tick is a share of weapon damage, not the row's amount or its coefficient, which is why the
	// dot is written out rather than taken from spelldata.DotConfig: 412609's periodic effect is a
	// dummy of one base point carrying a spell power coefficient of 1, and the resolver would tick
	// that amount with spell power on top of it.
	config.Dot = core.DotConfig{
		Aura: core.Aura{
			Label: "DeepWounds",
		},
		NumberOfTicks: int32(deepWoundsBleed.Duration() / tick.Period()),
		TickLength:    tick.Period(),

		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			dot.Spell.CalcAndDealPeriodicDamage(sim, target, dot.SnapshotBaseDamage, deepWoundsBleed.TickOutcome(dot))
		},
	}

	// A crit while the bleed is up adds the unticked damage to the new application and spreads the sum
	// over a full set of ticks; the ticks keep the first application's timer.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHitNoHitCounter)
		dot := spell.Dot(target)
		// One application is the share of the main hand's average weapon damage, without Attack Power,
		// whichever hand crit.
		totalDamage := dot.OutstandingDmg() + warrior.AutoAttacks.MH().AverageDamage()*share

		if dot.IsActive() {
			dot.DurationExtend(sim, deepWoundsBleed.Duration())
		} else {
			dot.Apply(sim)
		}
		dot.SnapshotBaseDamage = totalDamage / float64(dot.RemainingTicks())
	}

	warrior.DeepWounds = warrior.RegisterSpell(config)

	// The proc shape with no roll: 12834 states its rate as "always" and its crit hint carries the
	// tooltip's condition. The physical school the tooltip's melee weapon means is the caller's.
	trigger := spelldata.ProcTrigger(&warrior.Character,
		spellData.DeepWounds.Rank(warrior.Talents.DeepWounds),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			warrior.DeepWounds.Cast(sim, result.Target)
		})
	trigger.Name = "Deep Wounds - Trigger"
	trigger.TriggerImmediately = true
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.SpellSchool.Matches(core.SpellSchoolPhysical)
	}

	warrior.MakeProcTriggerAura(trigger)
}

func (warrior *Warrior) registerTwoHandedWeaponSpecialization() {
	if warrior.Talents.TwoHandedWeaponSpecialization == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character,
		spellData.TwoHandedWeaponSpecialization.Rank(warrior.Talents.TwoHandedWeaponSpecialization))
}

func (warrior *Warrior) registerImpale() {
	if warrior.Talents.Impale == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.Impale.Rank(warrior.Talents.Impale))
}

var mortalStrikeRank = spellData.MortalStrike.Highest()

func (warrior *Warrior) registerMortalStrike() {
	if !warrior.Talents.MortalStrike {
		return
	}

	config := spelldata.SpellConfig(&warrior.Unit, mortalStrikeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := mortalStrikeRank.WeaponStrike(sim, spell, spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.MortalStrike = warrior.RegisterSpell(config)
}

var spearingStrikeRank = spellData.SpearingStrike.Highest()

// The tooltip reads "deals $s2% weapon damage" and "an additional ${$s2*$s3}%" against Giants and
// Dragonkin, and the effects share an aura and misc value, so both are taken by effect index.
var spearingStrikeWeaponShare = spearingStrikeRank.EffectN(2).Percent()
var spearingStrikeGiantMultiplier = 1 + spearingStrikeRank.EffectN(3).Average()

func (warrior *Warrior) registerSpearingStrike() {
	if !warrior.Talents.SpearingStrike {
		return
	}

	config := spelldata.SpellConfig(&warrior.Unit, spearingStrikeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := spearingStrikeWeaponShare * spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		if target.MobType == proto.MobType_MobTypeGiant || target.MobType == proto.MobType_MobTypeDragonkin {
			baseDamage *= spearingStrikeGiantMultiplier
		}

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}

func (warrior *Warrior) registerBloodthrill() {
	if warrior.Talents.Bloodthrill == 0 {
		return
	}

	talent := spellData.Bloodthrill.Rank(warrior.Talents.Bloodthrill)

	// The proc opens Overpower's window as a dodge would, for the window's own duration, and the
	// cast consumes it. The tooltip's $s1 names an effect, so the row's own ProcChanceEffectN
	// reads the rate ladder. Main-hand white swings, extra attacks included, and the on-next-swing
	// Heroic Strike and Cleave proc it on a target with the warrior's Rend; the row's melee-ability
	// bit would also admit the instant specials, which in game do not proc it.
	trigger := spelldata.ProcTrigger(&warrior.Character, talent,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			warrior.OverpowerAura.Activate(sim)
		})
	trigger.Name = "Bloodthrill - Trigger"
	trigger.RequireDamageDealt = false
	trigger.RequireDamageComponent = true
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
		swing := spell.ProcMask.Matches(core.ProcMaskMeleeMHAuto) ||
			spellData.HeroicStrike.Contains(spell) || spellData.Cleave.Contains(spell)
		return swing && warrior.Rend.Dot(result.Target).IsActive()
	}

	warrior.MakeProcTriggerAura(trigger)
}

func (warrior *Warrior) registerWeaponmaster() {
	if warrior.Talents.Weaponmaster == 0 {
		return
	}

	rank := warrior.Talents.Weaponmaster
	actionID := core.ActionID{SpellID: spellData.Weaponmaster.Rank(rank).ID}

	// The three branches share A_DUMMY and misc 0, so each is named by its effect index.
	var critOn, armorIgnoreOn bool
	var swordMask core.ProcMask
	readWeapons := func() {
		critOn = warrior.GetProcMaskForTypes(proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypePolearm).Matches(core.ProcMaskMelee)
		armorIgnoreOn = warrior.GetProcMaskForTypes(proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeStaff).Matches(core.ProcMaskMelee)
		swordMask = warrior.GetProcMaskForTypes(proto.WeaponType_WeaponTypeSword)
	}
	readWeapons()

	// The aura is up while either hand holds an axe or a polearm, and its crit reaches the attacks
	// made with one.
	critAura := warrior.RegisterAura(core.Aura{
		Label:      "Weaponmaster (Axe/Polearm)",
		ActionID:   actionID.WithTag(1),
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(critOn, core.CharacterBuildPhaseTalents, core.CharacterBuildPhaseNone),
	}).AttachSpellMod(core.SpellModConfig{
		Kind:          core.SpellMod_BonusCrit_Percent,
		FloatValue:    spellData.Weaponmaster.EffectAt(1).ValueAt(rank),
		ProcMask:      core.ProcMaskMelee,
		EquipSubclass: dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM,
	})
	if critOn {
		core.MakePermanent(critAura)
	}

	// The aura is up while either hand holds a one- or two-handed mace or a staff, and its armor
	// ignore reaches every attack. A mace in each hand ignores the share once.
	armorIgnore := spellData.Weaponmaster.EffectAt(2).FractionAt(rank)
	armorIgnoreAura := warrior.RegisterAura(core.Aura{
		Label:    "Weaponmaster (Mace/Staff)",
		ActionID: actionID.WithTag(2),
		Duration: core.NeverExpires,
		OnGain: func(_ *core.Aura, _ *core.Simulation) {
			warrior.AddArmorIgnoreFactor(armorIgnore)
		},
		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			warrior.AddArmorIgnoreFactor(-armorIgnore)
		},
	})
	if armorIgnoreOn {
		core.MakePermanent(armorIgnoreAura)
	}

	// 1290261 states no proc flags at all, so the row decodes to a listener that hears nothing:
	// the shape, the mask and the rate's effect are all the caller's.
	swordTrigger := core.ProcTrigger{
		Name:            "Weaponmaster (Sword)",
		ActionID:        actionID.WithTag(3),
		MetricsActionID: actionID.WithTag(3),
		Callback:        core.CallbackOnSpellHitDealt,
		ProcMask:        core.ProcMaskMelee,
		Outcome:         core.OutcomeLanded,
		ProcChance:      spellData.Weaponmaster.EffectAt(3).FractionAt(rank),
		Duration:        core.NeverExpires,
		ExtraCondition: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spell.ProcMask.Matches(swordMask)
		},
	}
	onInit := warrior.ExtraAttackProc(&swordTrigger, actionID.SpellID, 1)
	swordAura := warrior.MakeProcTriggerAura(swordTrigger).ApplyOnInit(onInit).ApplyOnInit(func(aura *core.Aura, _ *core.Simulation) {
		aura.OnReset = func(aura *core.Aura, sim *core.Simulation) {
			if swordMask == core.ProcMaskUnknown {
				aura.Deactivate(sim)
			} else {
				aura.Activate(sim)
			}
		}
	})

	setActive := func(sim *core.Simulation, aura *core.Aura, on bool) {
		if on {
			aura.Activate(sim)
		} else {
			aura.Deactivate(sim)
		}
	}
	warrior.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
		readWeapons()
		setActive(sim, critAura, critOn)
		setActive(sim, armorIgnoreAura, armorIgnoreOn)
		setActive(sim, swordAura, swordMask != core.ProcMaskUnknown)
	})
}

func (warrior *Warrior) registerImprovedHamstring() {
	if warrior.Talents.ImprovedHamstring == 0 {
		return
	}

	talent := spellData.ImprovedHamstring.Rank(warrior.Talents.ImprovedHamstring)
	root := talent.EffectN(1).Trigger()

	immobilizeAuras := warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(spelldata.AuraConfig(root, spelldata.Label("Improved Hamstring-"+warrior.Label)))
	})

	// The rate is the effect ladder the tooltip's $m1 names. The one ability it fires on is a
	// shape no proc mask states, so the row's listener is narrowed to Hamstring by hand.
	trigger := spelldata.ProcTrigger(&warrior.Character, talent,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			immobilizeAuras.Get(result.Target).Activate(sim)
		})
	trigger.Name = "Improved Hamstring - Trigger"
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spellData.Hamstring.Contains(spell)
	}

	warrior.MakeProcTriggerAura(trigger)
}

func (warrior *Warrior) registerImprovedSlam() {
	if warrior.Talents.ImprovedSlam == 0 {
		return
	}

	// The five rank-swap effects the row states past the cast time, the global cooldown and Slam's
	// cooldown have no sim kind and are reported as skipped.
	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedSlam.Rank(warrior.Talents.ImprovedSlam))
}

var sweepingStrikesRank = spellData.SweepingStrikes.Highest()

func (warrior *Warrior) registerSweepingStrikes() {
	if !warrior.Talents.SweepingStrikes {
		return
	}

	actionID := core.ActionID{SpellID: 12723}

	var copyDamage float64
	hitSpell := warrior.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		ClassFlags:  SpellFlagsSweepingStrikes,
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeSpecial,
		Flags:       core.SpellFlagIgnoreModifiers | core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, copyDamage, spell.OutcomeAlwaysHit)
		},
	})

	warrior.SweepingStrikesNormalizedAttack = warrior.RegisterSpell(core.SpellConfig{
		ActionID:    actionID.WithTag(1), // Real SpellID: 26654
		ClassFlags:  SpellFlagsSweepingStrikes,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeAlwaysHit)
		},
	})

	// The proc shape with no roll: 12292 states its rate as "always" and the buff's charges are
	// what run out. The duration and the charge count are the row's; which hits spend a charge is
	// the handler's.
	sweepingStrikes := spelldata.ProcTrigger(&warrior.Character, sweepingStrikesRank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if warrior.Env.ActiveTargetCount() < 2 || warrior.SweepingStrikesAura.GetStacks() == 0 || result.PostOutcomeDamage <= 0 {
				return
			}

			if spell == hitSpell || spell == warrior.SweepingStrikesNormalizedAttack || spellData.ThunderClap.Contains(spell) || spellData.Whirlwind.Contains(spell) {
				return
			}

			nextTarget := warrior.Env.NextActiveTargetUnit(result.Target)
			if spellData.Execute.Contains(spell) && result.Target.IsExecutePhase20() {
				warrior.SweepingStrikesNormalizedAttack.Cast(sim, nextTarget)
			} else {
				copyDamage = result.Damage / result.ArmorAndResistanceMultiplier
				hitSpell.Cast(sim, nextTarget)
			}

			warrior.SweepingStrikesAura.RemoveStack(sim)
		})
	sweepingStrikes.MetricsActionID = actionID
	sweepingStrikes.Duration = sweepingStrikesRank.Duration()
	sweepingStrikes.TriggerImmediately = true

	warrior.SweepingStrikesAura = warrior.MakeProcTriggerAura(sweepingStrikes)
	warrior.SweepingStrikesAura.MaxStacks = int32(sweepingStrikesRank.ProcCharges)

	config := spelldata.SpellConfig(&warrior.Unit, sweepingStrikesRank)
	// The sim casts the ability under the id of the strike it grants, which is what the APL names.
	config.ActionID = actionID

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.RelatedSelfBuff.Activate(sim)
		warrior.SweepingStrikesAura.SetStacks(sim, int32(sweepingStrikesRank.ProcCharges))
	}

	config.RelatedSelfBuff = warrior.SweepingStrikesAura

	ssCD := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: ssCD,
		Type:  core.CooldownTypeDPS,
	})
}
