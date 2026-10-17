package paladin

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (paladin *Paladin) registerHolyTalents() {
	// Tier 1
	paladin.applyDivineStrength()
	paladin.applyDivineIntellect()

	// Tier 2
	paladin.applyHealingLight()
	paladin.applySpiritualFocus()
	paladin.applyImprovedSeals()
	// Unyielding Faith shortens Fear and Disorient effects, which the sim never suffers.

	// Tier 3
	// Voice of Truth grants immunity to Silence and Interrupt effects, which the sim never suffers.
	paladin.applyReverence()
	paladin.applyPurifyingPower()

	// Tier 4
	paladin.applyInfusionOfLight()
	paladin.applyIllumination()
	// Divine Favor registered in registerTalentSpells

	// Tier 5
	paladin.applyDivinePrecision()
	// Holy Shock registered in registerTalentSpells
	paladin.applyConsecratedGround()

	// Tier 6
	paladin.applyHolyPower()

	// Tier 7
	// Light's Vigil registered in registerTalentSpells
}

// Divine Strength - Increases your Strength by 2/4/6/8/10%.
func (paladin *Paladin) applyDivineStrength() {
	if paladin.Talents.DivineStrength == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.DivineStrength.Rank(paladin.Talents.DivineStrength))
}

// Divine Intellect - Increases your total Intellect by 2/4/6/8/10%.
func (paladin *Paladin) applyDivineIntellect() {
	if paladin.Talents.DivineIntellect == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.DivineIntellect.Rank(paladin.Talents.DivineIntellect))
}

// Healing Light - Increases the amount healed by your Holy Light, Flash of Light, and Holy Shock
// spells by 4/8/12%.
func (paladin *Paladin) applyHealingLight() {
	if paladin.Talents.HealingLight == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.HealingLight.Rank(paladin.Talents.HealingLight))
}

// Spiritual Focus - Gives your Flash of Light, Holy Light, and Light's Vigil spells a 35/70% chance
// to not lose casting time when you take damage.
func (paladin *Paladin) applySpiritualFocus() {
	if paladin.Talents.SpiritualFocus == 0 {
		return
	}

	paladin.PseudoStats.PushbackChance -= spellData.SpiritualFocus.FractionAt(paladin.Talents.SpiritualFocus)
}

// Improved Seals - Increases the damage done by your Seals and Judgements by 5/10/15%.
func (paladin *Paladin) applyImprovedSeals() {
	if paladin.Talents.ImprovedSeals == 0 {
		return
	}

	// The row names the Righteousness and Fury hits and the damaging Judgements only. The Seal of Command
	// proc takes the talent through its weapon percent, see registerSealOfCommand.
	paladin.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskSealOfRighteousnessProc | SpellMaskSealOfFuryProc |
			SpellMaskJudgementOfRighteousness | SpellMaskJudgementOfCommand | SpellMaskJudgementOfFury,
		Kind:       core.SpellMod_DirectDamageDone_Pct,
		FloatValue: spellData.ImprovedSeals.FractionAt(paladin.Talents.ImprovedSeals),
	})
}

// Reverence - Allows 10/20/30% of your Mana regeneration to continue while casting.
func (paladin *Paladin) applyReverence() {
	if paladin.Talents.Reverence == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.Reverence.Rank(paladin.Talents.Reverence))
}

// Purifying Power - Reduces the mana cost of your Cleanse and Purify spells by 10/20% and reduces
// the cooldown of your Exorcism and Holy Wrath spells by 17/33%.
func (paladin *Paladin) applyPurifyingPower() {
	if paladin.Talents.PurifyingPower == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.PurifyingPower.Rank(paladin.Talents.PurifyingPower))
}

// Infusion of Light - Your Holy Shock and Flash of Light critical hits reduce the cast time of your
// next Holy Light cast within 15 sec by 0.5/1.0 sec.
func (paladin *Paladin) applyInfusionOfLight() {
	if paladin.Talents.InfusionOfLight == 0 {
		return
	}

	rank := spellData.InfusionOfLightTriggered.Highest()

	var infusion *core.Aura
	infusion = paladin.RegisterAura(core.Aura{
		Label:    "Infusion of Light" + paladin.Label,
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: rank.Duration(),
	}).AttachSpellMod(core.SpellModConfig{
		ClassMask: SpellMaskHolyLight,
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Duration(spellData.InfusionOfLight.ValueAt(paladin.Talents.InfusionOfLight)) * time.Millisecond,
	}).AttachProcTrigger(core.ProcTrigger{
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     SpellMaskHolyLight,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			infusion.Deactivate(sim)
		},
	})

	// The row hears every spell crit and states no class flags; the tooltip names Holy Shock and
	// Flash of Light.
	trigger := spelldata.ProcTrigger(&paladin.Character, spellData.InfusionOfLight.Rank(paladin.Talents.InfusionOfLight), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		infusion.Activate(sim)
	})
	trigger.Name = "Infusion of Light - Trigger" + paladin.Label
	trigger.ClassSpellMask = SpellMaskHolyShock | SpellMaskHolyShockHeal | SpellMaskFlashOfLight
	paladin.MakeProcTriggerAura(trigger)
}

// Illumination - After getting a critical effect from your Flash of Light, Holy Light, Light's
// Vigil, or Holy Shock heal spell you have a 20/40/60/80/100% chance to gain Mana equal to 50% of
// the base cost of the spell.
func (paladin *Paladin) applyIllumination() {
	if paladin.Talents.Illumination == 0 {
		return
	}

	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: spellData.Illumination.Highest().ID})
	refund := spellData.Illumination.EffectAt(3).FractionAt(paladin.Talents.Illumination)

	// The row hears every heal crit and states no class flags; the tooltip names the healing spells.
	trigger := spelldata.ProcTrigger(&paladin.Character, spellData.Illumination.Rank(paladin.Talents.Illumination), func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
		paladin.AddMana(sim, float64(spell.Cost.BaseCost)*refund, manaMetrics)
	})
	trigger.Name = "Illumination" + paladin.Label
	trigger.ClassSpellMask = SpellMaskHealingSpells
	trigger.Outcome = core.OutcomeCrit
	paladin.MakeProcTriggerAura(trigger)
}

// Divine Precision - Improves your chance to hit with Holy spells by 6/12/18%.
func (paladin *Paladin) applyDivinePrecision() {
	if paladin.Talents.DivinePrecision == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.DivinePrecision.Rank(paladin.Talents.DivinePrecision), spelldata.ShowAsSchoolHit(core.SpellSchoolHoly))
}

// Consecrated Ground - Gives your Holy spells 5/10% increased damage against the first 4 enemies
// that enter your Consecration.
func (paladin *Paladin) applyConsecratedGround() {
	if paladin.Talents.ConsecratedGround == 0 {
		return
	}

	rank := spellData.ConsecratedGround.Highest()
	multiplier := spellData.ConsecratedGround.MultiplierAt(paladin.Talents.ConsecratedGround)

	paladin.consecratedGroundAuras = paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Consecrated Ground" + paladin.Label,
			ActionID: core.ActionID{SpellID: rank.ID},
			Duration: spellData.ConsecratedGroundTriggered.Highest().Duration(),
		}).AttachDDBC(0, 1, &paladin.AttackTables, func(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
			if spell.SpellSchool.Matches(core.SpellSchoolHoly) {
				return multiplier
			}
			return 1
		})
	})
}

// Holy Power - Increases the critical strike chance of your Holy Shock and Holy Strike spells by
// 3/6/9/12/15%, and all other spells by 1/2/3/4/5%.
func (paladin *Paladin) applyHolyPower() {
	if paladin.Talents.HolyPower == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.HolyPower.Rank(paladin.Talents.HolyPower))
}
