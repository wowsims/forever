package paladin

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (paladin *Paladin) registerProtectionTalents() {
	// Tier 1
	paladin.applyToughness()
	paladin.applyRedoubt()

	// Tier 2
	paladin.applyPrecision()
	// Guardian's Favor changes Blessing of Protection and Blessing of Freedom, which are not modelled.
	paladin.applyAnticipation()

	// Tier 3
	// Improved Seal of Fury attaches to the seal's shield in seal_of_fury.go
	paladin.applyImprovedRighteousFury()
	paladin.applyShieldSpecialization()
	paladin.applySacredDuty()

	// Tier 4
	// Swift Judgement registered in registerTalentSpells
	paladin.applyOneHandedWeaponSpecialization()
	// Improved Hammer of Justice shortens a cooldown the sim does not model.

	// Tier 5
	// Templar's Bulwark registered in registerTalentSpells
	paladin.applyReckoning()

	// Tier 6
	paladin.applyIronCreed()

	// Tier 7
	// Holy Shield registered in registerTalentSpells
}

// Toughness - Increases your armor value from items by 2/4/6/8/10%.
func (paladin *Paladin) applyToughness() {
	if paladin.Talents.Toughness == 0 {
		return
	}

	paladin.ApplyEquipScaling(stats.Armor, spellData.Toughness.Effect(dbcenums.A_MOD_BASE_RESISTANCE_PCT, 1).MultiplierAt(paladin.Talents.Toughness))
}

// Redoubt - Damaging melee attacks against you have a 10% chance to increase your chance to block
// by 4/8/12/16/20%. Lasts 10 sec or 5 blocks.
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

	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Redoubt - Trigger" + paladin.Label,
		Callback:           core.CallbackOnSpellHitTaken,
		ProcMask:           core.ProcMaskMelee,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
		// The column is the roll on this row: 10%, as the tooltip states. Rank 0 is untaken and reads 0.
		ProcChance: float64(spellData.Redoubt.Rank(paladin.Talents.Redoubt).ProcChance) / 100,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			redoubt.Activate(sim)
			redoubt.SetStacks(sim, redoubt.MaxStacks)
		},
	})
}

// Precision - Improves your chance to hit by 1/2/3%, with melee weapons and spells alike.
func (paladin *Paladin) applyPrecision() {
	if paladin.Talents.Precision == 0 {
		return
	}

	paladin.AddStat(stats.PhysicalHitPercent, spellData.Precision.Effect(dbcenums.A_MOD_HIT_CHANCE, 0).ValueAt(paladin.Talents.Precision))
	paladin.AddStat(stats.SpellHitPercent, spellData.Precision.Effect(dbcenums.A_MOD_SPELL_HIT_CHANCE, 0).ValueAt(paladin.Talents.Precision))
}

// Anticipation - Increases your Defense Skill by 4/8/12/16/20.
func (paladin *Paladin) applyAnticipation() {
	if paladin.Talents.Anticipation == 0 {
		return
	}

	paladin.AddStat(stats.DefenseRating, spellData.Anticipation.ValueAt(paladin.Talents.Anticipation)*core.DefenseRatingPerDefenseLevel)
}

// Improved Seal of Fury - When Seal of Fury's shield is fully absorbed, restore 60 Mana, increased
// by 15% per level the attacker is above you, up to 45%. The level is the current target's. The
// game files state the mana as one point per caster level, so it is read folded.
func (paladin *Paladin) applyImprovedSealOfFury(shield *core.DamageAbsorptionAura) {
	if !paladin.Talents.ImprovedSealOfFury {
		return
	}

	rank := spellData.ImprovedSealOfFury.Highest()
	mana := rank.EffectN(1).Average()
	perLevel := rank.EffectN(2).Percent()
	maxLevels := rank.EffectN(3).Average()
	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: rank.ID})

	shield.AttachOnDamageAbsorbed(func(sim *core.Simulation, aura *core.DamageAbsorptionAura, _ *core.SpellResult, _ float64) {
		if aura.ShieldStrength > 0 {
			return
		}
		levels := min(maxLevels, max(0, float64(paladin.CurrentTarget.Level-paladin.Level)))
		paladin.AddMana(sim, mana*(1+perLevel*levels), manaMetrics)
	})
}

// Improved Righteous Fury - While Righteous Fury is active, all damage taken is reduced by 2/4/6%.
func (paladin *Paladin) applyImprovedRighteousFury() {
	if paladin.Talents.ImprovedRighteousFury == 0 {
		return
	}

	// The game files state this as a negative percentage per rank: -2 / -4 / -6, so MultiplierAt
	// gives 0.98 / 0.96 / 0.94 and the minus is never written here.
	multiplier := spellData.ImprovedRighteousFury.
		Effect(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_EFFECT2)).
		MultiplierAt(paladin.Talents.ImprovedRighteousFury)

	paladin.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Matches(SpellMaskRighteousFury) {
			spell.RelatedSelfBuff.AttachMultiplicativePseudoStatBuff(&paladin.PseudoStats.DamageTakenMultiplier, multiplier)
		}
	})
}

// Shield Specialization - Increases the amount of damage absorbed by your shield by 10/20/30%, and
// gives your blocks a 33/66/100% chance to restore 6% of your maximum Mana. May only occur once
// every 3 sec.
func (paladin *Paladin) applyShieldSpecialization() {
	if paladin.Talents.ShieldSpecialization == 0 {
		return
	}

	paladin.PseudoStats.BlockValueMultiplier *= spellData.ShieldSpecialization.Effect(dbcenums.A_MOD_BLOCK_VALUE_PCT, 0).MultiplierAt(paladin.Talents.ShieldSpecialization)

	rank := spellData.ShieldSpecializationTriggered.Highest()
	manaShare := rank.EffectN(1).Percent()
	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: rank.ID})

	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Shield Specialization" + paladin.Label,
		Callback:           core.CallbackOnSpellHitTaken,
		Outcome:            core.OutcomeBlock,
		ProcChance:         spellData.ShieldSpecialization.EffectAt(2).FractionAt(paladin.Talents.ShieldSpecialization),
		ICD:                time.Second * 3,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			paladin.AddMana(sim, paladin.MaxMana()*manaShare, manaMetrics)
		},
	})
}

// Sacred Duty - Increases your total Stamina by 2/4% and reduces the cooldown of your Divine
// Shield, Divine Protection, and Templar's Bulwark spells by 30/60 sec. Only Templar's Bulwark is
// modelled; the cooldown is the second effect.
func (paladin *Paladin) applySacredDuty() {
	if paladin.Talents.SacredDuty == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.SacredDuty.Rank(paladin.Talents.SacredDuty), spelldata.Effects(1))
	paladin.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskTemplarsBulwark,
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: time.Duration(spellData.SacredDuty.EffectAt(2).ValueAt(paladin.Talents.SacredDuty)) * time.Millisecond,
	})
}

// One-Handed Weapon Specialization - Increases the damage you deal with one-handed melee weapons
// by 3/7/10%.
func (paladin *Paladin) applyOneHandedWeaponSpecialization() {
	if paladin.Talents.OneHandedWeaponSpecialization == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character,
		spellData.OneHandedWeaponSpecialization.Rank(paladin.Talents.OneHandedWeaponSpecialization))
}

// Reckoning - Gives you a 8/16/24/32/40% chance to gain an extra attack after Blocking a melee
// attack and a 20/40/60/80/100% chance to gain an extra attack after being the victim of a
// non-periodic critical strike.
func (paladin *Paladin) applyReckoning() {
	if paladin.Talents.Reckoning == 0 {
		return
	}

	rank := spellData.ReckoningTriggered.Highest()
	blockChance := spellData.Reckoning.FractionAt(paladin.Talents.Reckoning)
	critChance := blockChance * 2.5

	// Both triggers swing the same main-hand copy, tagged with Reckoning, at the attacker.
	block := core.ProcTrigger{
		Name:       "Reckoning - Block" + paladin.Label,
		Callback:   core.CallbackOnSpellHitTaken,
		Outcome:    core.OutcomeBlock,
		ProcChance: blockChance,
	}
	onBlock := paladin.ExtraAttackProc(&block, rank.ID, 1)
	paladin.MakeProcTriggerAura(block).ApplyOnInit(onBlock)

	crit := core.ProcTrigger{
		Name:       "Reckoning - Crit" + paladin.Label,
		Callback:   core.CallbackOnSpellHitTaken,
		Outcome:    core.OutcomeCrit,
		ProcChance: critChance,
	}
	onCrit := paladin.ExtraAttackProc(&crit, rank.ID, 1)
	paladin.MakeProcTriggerAura(crit).ApplyOnInit(onCrit)
}

// Iron Creed - Increases the threat generated by your Holy Strike ability 5/10/15/20/25%. While
// Righteous Fury is active, Holy Strike also reduces your damage taken by 3/6/9/12/15% for 6 sec.
func (paladin *Paladin) applyIronCreed() {
	if paladin.Talents.IronCreed == 0 {
		return
	}

	paladin.AddStaticMod(core.SpellModConfig{
		ClassMask:  SpellMaskHolyStrike,
		Kind:       core.SpellMod_ThreatMultiplier_Pct,
		FloatValue: spellData.IronCreed.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_THREAT)).FractionAt(paladin.Talents.IronCreed),
	})

	rank := spellData.IronCreedTriggered.Highest()
	reduction := spellData.IronCreed.Effect(dbcenums.A_PROC_TRIGGER_SPELL_WITH_VALUE, 0).FractionAt(paladin.Talents.IronCreed)

	ironCreed := paladin.RegisterAura(core.Aura{
		Label:    "Iron Creed" + paladin.Label,
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: rank.Duration(),
	}).AttachMultiplicativePseudoStatBuff(&paladin.PseudoStats.DamageTakenMultiplier, 1-reduction)

	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Iron Creed - Trigger" + paladin.Label,
		Callback:       core.CallbackOnSpellHitDealt,
		ClassSpellMask: SpellMaskHolyStrike,
		Outcome:        core.OutcomeLanded,
		ExtraCondition: func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
			return paladin.RighteousFuryAura.IsActive()
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			ironCreed.Activate(sim)
		},
	})
}
