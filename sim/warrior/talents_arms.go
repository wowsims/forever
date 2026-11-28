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

// The rage regen aura on effect 1 states its amount on the rage bar's tenths, per five seconds as every
// A_MOD_POWER_REGEN does. No tick for it in game files, so the rage arrives as one gain every five
// seconds. The tooltip's "$m2 Rage every $m3 sec" reads the two dummies instead, 1 rage every 3 seconds.
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

	// The bleed does not take spelldata.Proc(): that marks the spell passive, and the metrics aggregator
	// counts no cast for a passive spell. Attacker modifiers in EV-0139.
	config := spelldata.SpellConfig(&warrior.Unit, deepWoundsBleed,
		spelldata.Flags(core.SpellFlagNoOnCastComplete|core.SpellFlagIgnoreResists|core.SpellFlagProc|core.SpellFlagIgnoreAttackerModifiers)) // 12162 and 412609 lack Not a Proc.
	config.ProcMask = core.ProcMaskEmpty

	// The dot is written out rather than taken from spelldata.DotConfig: 412609's periodic effect is a
	// dummy of one base point carrying a spell power coefficient of 1, and the resolver would tick that
	// amount with spell power on top of it.
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

	var application float64

	// A crit while the bleed is up adds the unticked damage to the new application and spreads the sum
	// over a full set of ticks; the ticks keep the first application's timer.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHitNoHitCounter)
		dot := spell.Dot(target)
		totalDamage := dot.OutstandingDmg() + application

		if dot.IsActive() {
			dot.DurationExtend(sim, deepWoundsBleed.Duration())
		} else {
			dot.Apply(sim)
		}
		dot.SnapshotBaseDamage = totalDamage / float64(dot.RemainingTicks())
	}

	warrior.DeepWounds = warrior.RegisterSpell(config)

	// Bleed starting crits in EV-0139.
	trigger := spelldata.ProcTrigger(&warrior.Character,
		spellData.DeepWounds.Rank(warrior.Talents.DeepWounds),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// Application amount in EV-0139.
			weapon := core.Ternary(spell.IsOH() && !spell.IsMH(), warrior.AutoAttacks.OH(), warrior.AutoAttacks.MH())
			application = weapon.AverageDamage() * share
			warrior.DeepWounds.Cast(sim, result.Target)
		})
	trigger.Name = "Deep Wounds - Trigger"
	trigger.TriggerImmediately = true
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.SpellSchool.Matches(core.SpellSchoolPhysical) && spell.DefenseType != core.DefenseTypeMagic
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
	config.VersusBonus = &core.VersusBonus{Multiplier: spearingStrikeGiantMultiplier, MobTypes: []proto.MobType{proto.MobType_MobTypeGiant, proto.MobType_MobTypeDragonkin}}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := spearingStrikeWeaponShare * spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

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

	// Hits that proc in EV-0056. Main-hand white swings include extra attacks.
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

	// The Mace/Staff aura's armor ignore reaches every attack. A mace in each hand ignores the share once.
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

	// 1290261 states no proc flags at all, so the shape, the mask and the rate's effect are the caller's.
	// Hits that roll in EV-0002, EV-0054. Cleave rolls it once a cast, on its first landed hit; no log
	// measures Cleave. Swings other procs grant (Hand of Justice) roll it as well.
	firstCleaveHit, resetFirstCleaveHit := core.OncePerCast()
	firstWhirlwindGrant, resetFirstWhirlwindGrant := core.OncePerCast()
	swordTrigger := core.ProcTrigger{
		Name:            "Weaponmaster (Sword)",
		ActionID:        actionID.WithTag(3),
		MetricsActionID: actionID.WithTag(3),
		Callback:        core.CallbackOnSpellHitDealt,
		ProcMask:        core.ProcMaskMelee,
		Outcome:         core.OutcomeLanded,
		ProcChance:      spellData.Weaponmaster.EffectAt(3).FractionAt(rank),
		Duration:        core.NeverExpires,
		ExtraCondition: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			if !spell.ProcMask.Matches(swordMask) || spell.ActionID.SpellID == sweepingStrikesHitID {
				return false
			}
			if spell.ProcMask.Matches(core.ProcMaskMeleeSpecial) && spellData.Cleave.Contains(spell) {
				return firstCleaveHit(sim, spell)
			}
			return true
		},
	}
	onInit := warrior.ExtraAttackProc(&swordTrigger, actionID.SpellID, 1)
	// The handler queues a swing on every success, so a Whirlwind's later successes are dropped here.
	grant := swordTrigger.Handler
	swordTrigger.Handler = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if spellData.Whirlwind.Contains(spell) && !firstWhirlwindGrant(sim, spell) {
			return
		}
		grant(sim, spell, result)
	}
	swordAura := warrior.MakeProcTriggerAura(swordTrigger).ApplyOnInit(onInit).ApplyOnInit(func(aura *core.Aura, _ *core.Simulation) {
		aura.OnReset = func(aura *core.Aura, sim *core.Simulation) {
			resetFirstCleaveHit()
			resetFirstWhirlwindGrant()
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

	// The spell's five rank-swap effects have no sim kind and are reported as skipped.
	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedSlam.Rank(warrior.Talents.ImprovedSlam))
}

var sweepingStrikesRank = spellData.SweepingStrikes.Highest()

// The strike both of Sweeping Strikes' hits report under.
const sweepingStrikesHitID = 12723

func (warrior *Warrior) registerSweepingStrikes() {
	if !warrior.Talents.SweepingStrikes {
		return
	}

	actionID := core.ActionID{SpellID: sweepingStrikesHitID}

	var copyDamage float64
	// The copy takes the hit's damage with the attacker's damage bonuses already in it, and those bonuses
	// apply to it again, except one that needs a weapon equipped (EV-0182). The logs measure Enrage and
	// Two-Handed Weapon Specialization; that every other bonus follows the same rule, Death Wish among
	// them, is an owner decision.
	twoHandSpec := 1.0
	var twoHandSpecWeapon core.EquipRequirement
	if warrior.Talents.TwoHandedWeaponSpecialization > 0 {
		row := spellData.TwoHandedWeaponSpecialization.Rank(warrior.Talents.TwoHandedWeaponSpecialization)
		twoHandSpec = 1 + row.EffectN(1).Percent()
		twoHandSpecWeapon = core.EquipRequirement{SubclassMask: row.EquipSubclass, InvTypeMask: row.EquipInvType, Hand: spelldata.RequiredHand(row)}
	}
	// Proc listeners in EV-0007. Equip procs still hear the copied hit;
	// no log measures them on 12723.
	hitSpell := warrior.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		ClassFlags:  SpellFlagsSweepingStrikes,
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeSpecial,
		DefenseType: core.DefenseTypeNone,
		Flags:       core.SpellFlagIgnoreTargetModifiers | core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := copyDamage
			if twoHandSpec != 1 && twoHandSpecWeapon.Equipped(&warrior.Character) {
				damage /= twoHandSpec
			}
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeAlwaysHit)
		},
	})

	warrior.SweepingStrikesNormalizedAttack = warrior.RegisterSpell(core.SpellConfig{
		ActionID:    actionID.WithTag(1), // Real SpellID: 26654
		ClassFlags:  SpellFlagsSweepingStrikes,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		// A main-hand strike, so only the main-hand weapon's procs hear it.
		ProcMask: core.ProcMaskMeleeMHSpecial,
		Flags:    core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			// Strike outcome in EV-0213.
			// TODO: Ingame test needed: whether the strike an Execute brings below 20% rolls a crit too; the
			// logs measure the Whirlwind ones.
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialCritOnly)
		},
	})

	sweepingStrikes := spelldata.ProcTrigger(&warrior.Character, sweepingStrikesRank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if warrior.Env.ActiveTargetCount() < 2 || warrior.SweepingStrikesAura.GetStacks() == 0 || result.PostOutcomeDamage <= 0 {
				return
			}

			// Whirlwind brings its one strike through CastNormalizedSweepingStrikesAttack.
			// Hits and charges in EV-0054.
			// TODO: Whirlwind's off-hand strike also brings off-hand strikes of Sweeping Strikes (462890 in
			// game), each using a charge, and the sim casts none; how many a cast brings is open (EV-0216).
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
