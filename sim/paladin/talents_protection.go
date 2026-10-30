package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (paladin *Paladin) registerProtectionTalents() {
	// Tier 1
	paladin.applyDivineIntellect()
	paladin.applyRedoubt()

	// Tier 2
	paladin.applyPrecision()
	// Guardian's Favor changes Blessing of Protection and Blessing of Freedom, which are not modelled.
	paladin.applySacredDuty()

	// Tier 3
	// Templar's Bulwark registered in registerTalentSpells
	// Improved Hammer of Justice shortens a cooldown the sim does not model.
	paladin.applyShieldSpecialization()
	paladin.applyToughness()

	// Tier 4
	// Swift Judgement registered in registerTalentSpells
	paladin.applyImprovedRighteousFury()
	paladin.applyAnticipation()

	// Tier 5
	// Stoicism shortens Stun and Charm effects, which the sim never suffers.
	// Holy Shield registered in registerTalentSpells
	paladin.applyReckoning()

	// Tier 6
	paladin.applyIronCreed()

	// Tier 7
	paladin.applyShieldsworn()
}

// Divine Intellect - Increases your total Intellect by 2/4/6/8/10%.
func (paladin *Paladin) applyDivineIntellect() {
	if paladin.Talents.DivineIntellect == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.DivineIntellect.Rank(paladin.Talents.DivineIntellect))
}

// Toughness - Increases your armor value from items by 2/4/6/8/10%.
func (paladin *Paladin) applyToughness() {
	if paladin.Talents.Toughness == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.Toughness.Rank(paladin.Talents.Toughness))
}

// Redoubt - Damaging melee attacks against you have a 10% chance to increase your chance to block
// by 6/12/18/24/30%. Lasts 10 sec or 5 blocks.
func (paladin *Paladin) applyRedoubt() {
	if paladin.Talents.Redoubt == 0 {
		return
	}

	rank := spellData.RedoubtTriggered.Highest()

	var redoubt *core.Aura
	redoubt = paladin.RegisterAura(core.Aura{
		Label:     "Redoubt" + paladin.Label,
		ActionID:  core.ActionID{SpellID: rank.ID},
		Duration:  rank.Duration(),
		MaxStacks: int32(rank.ProcCharges),
	}).AttachStatBuff(
		stats.BlockPercent, spellData.Redoubt.ValueAt(paladin.Talents.Redoubt),
	).AttachProcTrigger(core.ProcTrigger{
		Callback:           core.CallbackOnSpellHitTaken,
		Outcome:            core.OutcomeBlock,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			redoubt.RemoveStack(sim)
		},
	})

	trigger := spelldata.ProcTrigger(&paladin.Character, spellData.Redoubt.Rank(paladin.Talents.Redoubt), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		redoubt.Activate(sim)
		redoubt.SetStacks(sim, redoubt.MaxStacks)
	})
	trigger.Name = "Redoubt - Trigger" + paladin.Label
	paladin.MakeProcTriggerAura(trigger)
}

// Precision - Improves your chance to hit by 1/2/3%.
func (paladin *Paladin) applyPrecision() {
	if paladin.Talents.Precision == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.Precision.Rank(paladin.Talents.Precision))
}

// Anticipation - Increases your Defense Skill by 4/8/12/16/20.
func (paladin *Paladin) applyAnticipation() {
	if paladin.Talents.Anticipation == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.Anticipation.Rank(paladin.Talents.Anticipation))
}

// Improved Righteous Fury - While Righteous Fury is active, all damage taken is reduced by 3/6/9%.
func (paladin *Paladin) applyImprovedRighteousFury() {
	if paladin.Talents.ImprovedRighteousFury == 0 {
		return
	}

	// The game files state this as a negative percentage per rank: -3 / -6 / -9, so MultiplierAt
	// gives 0.97 / 0.94 / 0.91 and the minus is never written here.
	multiplier := spellData.ImprovedRighteousFury.
		Effect(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_EFFECT2)).
		MultiplierAt(paladin.Talents.ImprovedRighteousFury)

	paladin.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Matches(SpellMaskRighteousFury) {
			spell.RelatedSelfBuff.AttachMultiplicativePseudoStatBuff(&paladin.PseudoStats.DamageTakenMultiplier, multiplier)
		}
	})
}

// Shield Specialization - Increases the amount of damage absorbed by your shield by 15/30%, and
// gives your blocks a 50/100% chance to restore 4% of your maximum Mana. May only occur once
// every 3 sec. With Shieldsworn, a Holy Strike that lands rolls the same chance without the
// cooldown.
func (paladin *Paladin) applyShieldSpecialization() {
	if paladin.Talents.ShieldSpecialization == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.ShieldSpecialization.Rank(paladin.Talents.ShieldSpecialization), spelldata.Effects(1))

	rank := spellData.ShieldSpecializationTriggered.Highest()
	manaShare := rank.EffectN(1).Percent()
	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: rank.ID})
	manaMetrics.NoThreat = rank.GeneratesNoThreat()

	trigger := spelldata.ProcTrigger(&paladin.Character, spellData.ShieldSpecialization.Rank(paladin.Talents.ShieldSpecialization), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		paladin.AddMana(sim, paladin.MaxMana()*manaShare, manaMetrics)
	})
	trigger.Name = "Shield Specialization" + paladin.Label
	trigger.Outcome = core.OutcomeBlock
	// The tooltip's "your blocks" counts a block that stops the whole hit.
	trigger.RequireDamageDealt = false
	trigger.TriggerImmediately = true
	paladin.MakeProcTriggerAura(trigger)

	if paladin.Talents.Shieldsworn {
		paladin.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Shield Specialization - Shieldsworn" + paladin.Label,
			Callback:           core.CallbackOnSpellHitDealt,
			ClassSpellMask:     SpellMaskHolyStrike,
			Outcome:            core.OutcomeLanded,
			ProcChance:         trigger.ProcChance,
			TriggerImmediately: true,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				paladin.AddMana(sim, paladin.MaxMana()*manaShare, manaMetrics)
			},
		})
	}
}

// Sacred Duty - Reduces the cooldown of your Divine Shield, Divine Protection, and Templar's Bulwark
// spells by 30/60 sec, and reduces the duration of Forbearance by 5/10 sec.
func (paladin *Paladin) applySacredDuty() {
	if paladin.Talents.SacredDuty == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.SacredDuty.Rank(paladin.Talents.SacredDuty))
}

// Reckoning - Gives you a 8/16/24/32/40% chance after Blocking and a 20/40/60/80/100% chance after
// receiving a non-periodic critical strike to gain an extra attack. Stores up to 4 extra attacks, which
// the next main-hand swing spends. Cannot occur from Blocks more than once every 1.5 sec.
func (paladin *Paladin) applyReckoning() {
	if paladin.Talents.Reckoning == 0 {
		return
	}

	talent := spellData.Reckoning.Rank(paladin.Talents.Reckoning)
	rank := spellData.ReckoningTriggered.Highest()
	blockChance := spellData.Reckoning.FractionAt(paladin.Talents.Reckoning)
	critChance := blockChance * 2.5

	stored := paladin.RegisterAura(spelldata.AuraConfig(rank, spelldata.Label("Reckoning"+paladin.Label)))
	gain := func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		stored.Activate(sim)
		stored.AddStack(sim)
	}

	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Reckoning - Block" + paladin.Label,
		Callback:   core.CallbackOnSpellHitTaken,
		Outcome:    core.OutcomeBlock,
		ProcChance: blockChance,
		ICD:        talent.ICD(),
		Handler:    gain,
	})
	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Reckoning - Crit" + paladin.Label,
		Callback:   core.CallbackOnSpellHitTaken,
		Outcome:    core.OutcomeCrit,
		ProcChance: critChance,
		Handler:    gain,
	})

	// The next main-hand swing spends every stored charge as a copy of itself, tagged with
	// Reckoning, at the same target. The copies spend none.
	var extraAttack *core.Spell
	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Reckoning - Swing" + paladin.Label,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto,
		TriggerImmediately: true,
		ExtraCondition: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return extraAttack != nil && spell != extraAttack && stored.IsActive()
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			count := stored.GetStacks()
			stored.Deactivate(sim)
			target := result.Target
			sim.RunAfterCurrentEvent(func(sim *core.Simulation) {
				mainHand := paladin.AutoAttacks.MHAuto()
				extraAttack.SetSpellSchool(mainHand.SpellSchool)
				extraAttack.BonusCoefficient = mainHand.BonusCoefficient
				for range count {
					extraAttack.Cast(sim, target)
				}
			})
		},
	}).ApplyOnInit(func(_ *core.Aura, _ *core.Simulation) {
		if paladin.AutoAttacks.MHAuto() == nil {
			return
		}
		swing := *paladin.AutoAttacks.MHConfig()
		swing.ActionID = swing.ActionID.WithTag(rank.ID)
		extraAttack = paladin.GetOrRegisterSpell(swing)
	})
}

// Iron Creed - Increases the threat generated by your Holy Strike ability by 10/20/30%. While
// Righteous Fury is active, Holy Strike also reduces your damage taken by 2/4/6% for 6 sec.
func (paladin *Paladin) applyIronCreed() {
	if paladin.Talents.IronCreed == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.IronCreed.Rank(paladin.Talents.IronCreed), spelldata.Effects(1))

	rank := spellData.IronCreedTriggered.Highest()
	// The game files state the reduction as -2 / -4 / -6, so MultiplierAt gives 0.98 / 0.96 / 0.94.
	multiplier := spellData.IronCreed.Effect(dbcenums.A_PROC_TRIGGER_SPELL_WITH_VALUE, 0).MultiplierAt(paladin.Talents.IronCreed)

	ironCreed := paladin.RegisterAura(core.Aura{
		Label:    "Iron Creed" + paladin.Label,
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: rank.Duration(),
	}).AttachMultiplicativePseudoStatBuff(&paladin.PseudoStats.DamageTakenMultiplier, multiplier)

	// The row hears any melee special and names its ability only in the tooltip. In game a Holy
	// Strike that misses or is avoided still grants the reduction, so the trigger hears every outcome.
	trigger := spelldata.ProcTrigger(&paladin.Character, spellData.IronCreed.Rank(paladin.Talents.IronCreed), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		ironCreed.Activate(sim)
	})
	trigger.Name = "Iron Creed - Trigger" + paladin.Label
	trigger.ClassSpellMask = SpellMaskHolyStrike
	trigger.Outcome = core.OutcomeEmpty
	trigger.RequireDamageDealt = false
	trigger.ExtraCondition = func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
		return paladin.RighteousFuryAura.IsActive()
	}
	paladin.MakeProcTriggerAura(trigger)
}

// Shieldsworn - Reduces the cooldown of your Holy Strike ability by 4 sec. Additionally, Holy Strike
// now strikes with your shield, triggering your active Holy Shield and Shield Specialization talents
// without consuming their charges or triggering their cooldowns, and allowing you to Block spells
// for 6 sec.
//
// Holy Shield and Shield Specialization read the talent where they register. The sim lets only
// melee and ranged attacks be blocked, so the 6 sec spell block is not modelled.
// TODO: Verify in game whether a Holy Strike that misses still triggers Holy Shield and Shield
// Specialization; both here hear only a hit that lands.
func (paladin *Paladin) applyShieldsworn() {
	if !paladin.Talents.Shieldsworn {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.Shieldsworn.Rank(1))
}
