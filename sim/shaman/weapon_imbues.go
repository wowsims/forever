package shaman

import (
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Windfury Weapon and Flametongue Weapon put the same passive on the weapon at every rank, and the
// rank decides only which spells the passive fires, so they are read off the top rank's refs.
var (
	windfuryWeaponPassive    = spelldata.MustFind(439431)
	windfuryWeaponAttack     = spellData.WindfuryWeapon.Highest().Refs()[0]
	flametongueWeaponPassive = spelldata.MustFind(436519)
	flametongueWeaponProc    = spellData.FlametongueWeapon.Highest().Refs()[0]
	flametongueWeaponAttack  = spellData.FlametongueWeapon.Highest().Refs()[1]
	frostbrandWeaponAttack   = spellData.FrostbrandWeapon.Highest().Refs()[0]
	rockbiterWeaponPassive   = spellData.RockbiterWeapon.Highest().Refs()[0]
	rockbiterWeaponProc      = spelldata.MustFind(20871)
)

// Every imbue has its top rank's cast and an aura, so a rotation can put any of them on the main hand.
// The effects register wherever the shaman has a main hand, equipped or in the swap set, and each
// does nothing while the main hand in use carries another imbue. A shaman's imbue sits beside an oil
// or a stone on the same weapon, so it is kept apart from the weapon's temporary enchant.
func (shaman *Shaman) registerImbues() {
	shaman.mainHandID = shaman.Equipment[proto.ItemSlot_ItemSlotMainHand].ID
	if shaman.ItemSwap.IsEnabled() {
		shaman.swapMainHandID = shaman.ItemSwap.GetUnequippedItemBySlot(proto.ItemSlot_ItemSlotMainHand).ID
	}
	shaman.imbues = [2]proto.ShamanImbue{shaman.SelfBuffs.ImbueMH, shaman.SelfBuffs.ImbueMHSwap}

	shaman.registerImbueCast(proto.ShamanImbue_WindfuryWeapon, spellData.WindfuryWeapon.Highest())
	shaman.registerImbueCast(proto.ShamanImbue_FlametongueWeapon, spellData.FlametongueWeapon.Highest())
	shaman.registerImbueCast(proto.ShamanImbue_FrostbrandWeapon, spellData.FrostbrandWeapon.Highest())
	shaman.registerImbueCast(proto.ShamanImbue_RockbiterWeapon, spellData.RockbiterWeapon.Highest())

	if shaman.mainHandID == 0 && shaman.swapMainHandID == 0 {
		return
	}

	shaman.RegisterWindfuryImbue()
	shaman.RegisterFlametongueImbue()
	shaman.RegisterFrostbrandImbue()
	shaman.RegisterRockbiterImbue()

	shaman.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, _ proto.ItemSlot) {
		shaman.applyImbues(sim)
	})
}

// The aura is the cast's buff, so a rotation can ask which imbue the main hand carries. The imbue
// lasts 60 minutes, longer than any fight, so the aura never runs out. The cast replaces the imbue on
// the main hand in use, and the GCD, the mana and the melee weapon it needs come from the row.
func (shaman *Shaman) registerImbueCast(imbue proto.ShamanImbue, cast *spelldata.Spell) {
	aura := shaman.RegisterAura(core.Aura{
		Label:    cast.Name + " (Main Hand)",
		ActionID: core.ActionID{SpellID: cast.ID},
		Duration: core.NeverExpires,
	})
	shaman.imbueAuras[imbue] = aura

	config := spelldata.SpellConfig(&shaman.Unit, cast, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.RelatedSelfBuff = aura
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		shaman.imbues[shaman.mainHandIndex()] = imbue
		shaman.applyImbues(sim)
	}
	shaman.RegisterSpell(config)
}

// Which of the two main hands is in use. They are told apart by item id, so a swap set holding the
// same weapon reads the equipped one's imbue.
func (shaman *Shaman) mainHandIndex() int {
	if id := shaman.MainHand().ID; id == shaman.swapMainHandID && id != shaman.mainHandID {
		return 1
	}
	return 0
}

// The imbue on the main hand in use, and none while that hand is empty.
func (shaman *Shaman) mainHandImbue() proto.ShamanImbue {
	if shaman.MainHand().ID == 0 {
		return proto.ShamanImbue_NoImbue
	}
	return shaman.imbues[shaman.mainHandIndex()]
}

// Puts up the aura of the imbue the main hand in use carries, takes down the others, and lets each
// imbue's effects follow, after a cast or a weapon swap.
func (shaman *Shaman) applyImbues(sim *core.Simulation) {
	current := shaman.mainHandImbue()
	for imbue, aura := range shaman.imbueAuras {
		if aura == nil {
			continue
		}
		if proto.ShamanImbue(imbue) == current {
			aura.Activate(sim)
		} else {
			aura.Deactivate(sim)
		}
	}
	for _, onChange := range shaman.onImbueChange {
		onChange(sim)
	}
}

// Rolls on the main hand's hits while it carries the imbue, at the weapon's speed for a rate per
// minute, and nothing rolls while it does not.
func (shaman *Shaman) imbueProcManager(imbue proto.ShamanImbue, ppm float64, chance float64) *core.DynamicProcManager {
	build := func() core.DynamicProcManager {
		procMask := core.ProcMaskUnknown
		if shaman.mainHandImbue() == imbue {
			procMask = core.ProcMaskMeleeMH
		}
		if ppm != 0 {
			return *shaman.NewStaticLegacyPPMManager(ppm, procMask)
		}
		return *shaman.NewFixedProcChanceManager(chance, procMask)
	}
	dpm := build()
	shaman.onImbueChange = append(shaman.onImbueChange, func(_ *core.Simulation) {
		dpm = build()
	})
	return &dpm
}

// "When applied to main hand, disables any benefit you personally receive from" the totem of the
// same name: while the main hand carries the imbue, its aura outbids the totem for its category.
func (shaman *Shaman) holdTotemCategoryOnMainHand(category string, imbue proto.ShamanImbue) {
	shaman.imbueAuras[imbue].NewExclusiveEffect(category, false, core.ExclusiveEffect{
		Priority: math.MaxFloat64,
	})
}

// The extra attacks are logged as 439440, with 16361's attack power added. They are white in name
// only: auto attack procs such as Flametongue Totem hear them, but they roll the weapon special's
// table of miss, dodge, parry and crit, never glance, and spend no Flurry charge. Like every extra
// attack they report as a main-hand attack tagged with the spell that grants them.
func (shaman *Shaman) newWindfuryImbueSpell() *core.Spell {
	// Elemental Weapons raises effect 1, which the parse would read as damage.
	attackPower := windfuryWeaponAttack.EffectN(1).Average() * shaman.windfuryWeaponMultiplier

	return shaman.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: windfuryWeaponAttack.ID},
		SpellSchool:      core.SpellSchoolPhysical,
		DefenseType:      core.DefenseTypeMelee,
		ProcMask:         core.ProcMaskMeleeMHAuto,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := shaman.MHWeaponDamage(sim, spell.MeleeAttackPower(target)+attackPower)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	})
}

// 439431 rolls its 20% on the main hand's melee hits, once per 1.5 s, and fires the rank's 16361,
// whose extra attacks count is how many strikes follow. It rolls off the hit's cast result, and the
// strikes follow once that hit has been dealt, in the same instant.
func (shaman *Shaman) RegisterWindfuryImbue() {
	strike := shaman.newWindfuryImbueSpell()
	shaman.windfuryWeaponStrike = strike
	count := int(windfuryWeaponAttack.ExtraAttacksEffect().BaseValue())

	chance := windfuryWeaponPassive.StatedChance()
	trigger := spelldata.ProcTrigger(&shaman.Character, windfuryWeaponPassive, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		target := result.Target
		sim.RunAfterCurrentEvent(func(sim *core.Simulation) {
			for range count {
				strike.Cast(sim, target)
			}
		})
	}, func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = shaman.imbueProcManager(proto.ShamanImbue_WindfuryWeapon, 0, chance)
		trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spell != strike
		}
	})
	trigger.Name = "Windfury Weapon Trigger"
	trigger.Callback = core.CallbackOnCastResult
	trigger.TriggerImmediately = true
	shaman.MakeProcTriggerAura(trigger)

	shaman.holdTotemCategoryOnMainHand(buffs.WindfuryTotemCategory, proto.ShamanImbue_WindfuryWeapon)
}

// The hit is 10444, and the tooltip's ${$16344m1/77-1+($10444m1)} to ${$16344m1/25+($10444m1)}
// over weapon speeds of 1.3 to 4.0 s adds 16344's m1 for every 100 s of swing to its damage.
func (shaman *Shaman) newFlametongueImbueSpell() *core.Spell {
	config := spelldata.SpellConfig(&shaman.Unit, flametongueWeaponAttack,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc(),
		spelldata.Flags(core.SpellFlagProc|SpellFlagShamanSpell))
	config.ClassSpellMask = SpellMaskFlametongueWeapon

	perSwing := flametongueWeaponProc.EffectN(1)
	damage := flametongueWeaponAttack.EffectN(1)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := perSwing.Average()*shaman.AutoAttacks.MH().SwingSpeed/100 + damage.Average()
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	}
	return shaman.RegisterSpell(config)
}

// 436519 has no roll: every melee hit of the main hand adds its fire damage.
func (shaman *Shaman) RegisterFlametongueImbue() {
	flametongueSpell := shaman.newFlametongueImbueSpell()

	trigger := spelldata.ProcTrigger(&shaman.Character, flametongueWeaponPassive, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		flametongueSpell.Cast(sim, result.Target)
	}, func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = shaman.imbueProcManager(proto.ShamanImbue_FlametongueWeapon, 0, 1)
	})
	trigger.Name = "Flametongue Weapon Trigger"
	trigger.TriggerImmediately = true
	shaman.MakeProcTriggerAura(trigger)

	shaman.holdTotemCategoryOnMainHand(buffs.FlametongueTotemCategory, proto.ShamanImbue_FlametongueWeapon)
}

// Enchant 1668 casts 16353 off the main hand's hits. The slow, effect 1, is not
// modelled.
func (shaman *Shaman) RegisterFrostbrandImbue() {
	config := spelldata.SpellConfig(&shaman.Unit, frostbrandWeaponAttack,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc(), spelldata.Flags(core.SpellFlagProc|SpellFlagShamanSpell))
	config.ClassSpellMask = SpellMaskFrostbrandWeapon
	damage := frostbrandWeaponAttack.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}
	frostbrandSpell := shaman.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(&shaman.Character, frostbrandWeaponAttack, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		frostbrandSpell.Cast(sim, result.Target)
	}, spelldata.WeaponProc(), func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = shaman.imbueProcManager(proto.ShamanImbue_FrostbrandWeapon, float64(frostbrandWeaponAttack.RPPM), 0)
	})
	trigger.Name = "Frostbrand Weapon Trigger"
	trigger.TriggerImmediately = true
	shaman.MakeProcTriggerAura(trigger)
}

// Enchant 1664 puts 16313 on the main hand as an equip aura: the attack power, and the threat Spirit
// Weapons gives effect 2, which the row states as 0. Both ride on the imbue's aura.
func (shaman *Shaman) RegisterRockbiterImbue() {
	// Elemental Weapons raises effect 1, which the parse would read as damage.
	attackPower := rockbiterWeaponPassive.EffectN(1).Average() * shaman.rockbiterWeaponMultiplier
	threat := 1 + rockbiterWeaponPassive.EffectN(2).Percent() + shaman.spiritWeaponsRockbiterThreat

	shaman.imbueAuras[proto.ShamanImbue_RockbiterWeapon].
		AttachStatBuff(stats.AttackPower, attackPower).
		AttachMultiplicativePseudoStatBuff(&shaman.PseudoStats.ThreatMultiplier, threat)

	threatSpell := shaman.newRockbiterThreatSpell()
	trigger := spelldata.ProcTrigger(&shaman.Character, rockbiterWeaponPassive, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		threatSpell.Cast(sim, result.Target)
	}, func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = shaman.imbueProcManager(proto.ShamanImbue_RockbiterWeapon, 0, 1)
	})
	trigger.Name = "Rockbiter Weapon Trigger"
	trigger.TriggerImmediately = true
	shaman.MakeProcTriggerAura(trigger)
}

// The top rank's passive, 16313, casts 20871 off each melee hit, the top of the 20860 to 20871 ladder;
// its script effect adds its base points of threat for every second of the main hand's swing.
func (shaman *Shaman) newRockbiterThreatSpell() *core.Spell {
	config := spelldata.SpellConfig(&shaman.Unit, rockbiterWeaponProc, spelldata.Proc())
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	config.FlatThreatBonus = rockbiterWeaponProc.EffectN(1).Average()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		perSecond := spell.FlatThreatBonus
		spell.FlatThreatBonus *= shaman.AutoAttacks.MH().SwingSpeed
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
		spell.FlatThreatBonus = perSecond
	}
	return shaman.RegisterSpell(config)
}
