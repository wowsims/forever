package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (priest *Priest) registerHolyTalents() {
	// Tier 1
	priest.applyTwilightFocus()
	priest.applyImprovedRenew()
	priest.applyHolySpecialization()

	// Tier 2
	priest.applySpellWarding()
	priest.applyDivineFury()

	// Tier 3
	priest.applyHolyNova()
	priest.applyBlessedRecovery()
	priest.applyInspiration()

	// Tier 4
	priest.applyHolyReach()
	priest.applyImprovedHealing()
	priest.applySearingLight()
	priest.applyBindingHeal()

	// Tier 5
	priest.applyLitanyOfLight()
	priest.applySpiritOfRedemption()
	priest.applySpiritualGuidance()

	// Tier 6
	priest.applySpiritualHealing()

	// Tier 7
	priest.applyPrayerOfMending()
}

// Gives you a 70% chance to avoid interruption caused by damage while casting any spell.
// The caster takes no hits in a damage sim, so it changes nothing.
func (priest *Priest) applyTwilightFocus() {
	if priest.Talents.TwilightFocus == 0 {
		return
	}
}

// Increases the amount healed by your Renew spell by 5%.
// TODO: Renew rank 1289450 states 35 per tick with a 0 coefficient and no level requirement; verify in game before porting Renew.
func (priest *Priest) applyImprovedRenew() {
	if priest.Talents.ImprovedRenew == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.ImprovedRenew.Rank(priest.Talents.ImprovedRenew))
}

// Increases the critical effect chance of your Holy spells by X%.
func (priest *Priest) applyHolySpecialization() {
	if priest.Talents.HolySpecialization == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.HolySpecialization.Rank(priest.Talents.HolySpecialization))
}

// Reduces all spell damage taken by X%
func (priest *Priest) applySpellWarding() {
	if priest.Talents.SpellWarding == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.SpellWarding.Rank(priest.Talents.SpellWarding))
}

// Reduces the casting time of your Smite, Holy Fire, Heal and Greater Heal spells by x sec
func (priest *Priest) applyDivineFury() {
	if priest.Talents.DivineFury == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.DivineFury.Rank(priest.Talents.DivineFury))
}

// Grants the holy nova spell
func (priest *Priest) applyHolyNova() {
	if !priest.Talents.HolyNova {
		return
	}

	HolyNovaRankMap.Each(priest.registerHolyNovaSpell)
}

var HolyNovaRankMap = spellData.HolyNova

// After being struck by a melee or ranged critical hit, or suffering more than 30% of your maximum Health from a single attack,
// heal 8% of the damage taken over 6 sec. Refreshing this effect carries over any remaining healing.
// TODO: The sim does not model the incoming damage events that trigger this heal.
func (priest *Priest) applyBlessedRecovery() {
	if priest.Talents.BlessedRecovery == 0 {
		return
	}
}

// Your non-periodic critical heals increase your target's Armor by X% for 15 sec.
func (priest *Priest) applyInspiration() {
	if priest.Talents.Inspiration == 0 {
		return
	}

	triggered := spellData.InspirationTriggered.Highest()
	armorAuras := priest.NewAllyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.GetOrRegisterAura(spelldata.AuraConfig(triggered, spelldata.RankLabel(triggered, &priest.Unit)))
		spelldata.ParseEffects(nil, aura, triggered)
		return aura
	})
	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Inspiration",
		Callback: core.CallbackOnHealDealt,
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			if result.DidCrit() {
				armorAuras.Get(result.Target).Activate(sim)
			}
		},
	})
}

// Increases the range of your Smite and Holy Fire spells and the radius of your Prayer of Healing and Holy Nova spells by X%.
// The radius (effect 2) has no spell mod in the sim; Holy Nova already hits every enemy.
func (priest *Priest) applyHolyReach() {
	if priest.Talents.HolyReach == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.HolyReach.Rank(priest.Talents.HolyReach), spelldata.Effects(1))
}

// Reduces the Mana cost of your Lesser Heal, Heal, Greater Heal, Penance, and Prayer of Mending spells by x%.
func (priest *Priest) applyImprovedHealing() {
	if priest.Talents.ImprovedHealing == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.ImprovedHealing.Rank(priest.Talents.ImprovedHealing))
}

// Increases your Holy damage done by X%, and gives a Y% chance each time your Holy Fire spell deals periodic damage for your next Holy Nova to cost no Mana.
func (priest *Priest) applySearingLight() {
	if priest.Talents.SearingLight == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.SearingLight.Rank(priest.Talents.SearingLight), spelldata.Effects(1))

	holyPurposeRow := spellData.SearingLightTriggered.Highest()
	holyPurposeConfig := spelldata.AuraConfig(holyPurposeRow)
	holyPurposeConfig.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spellData.HolyNova.Contains(spell) {
			aura.Deactivate(sim)
		}
	}
	holyPurpose := priest.RegisterAura(holyPurposeConfig)
	spelldata.ParseEffects(&priest.Character, holyPurpose, holyPurposeRow, spelldata.Effects(1))

	// The row's listener names no spell; the tooltip limits it to Holy Fire's periodic damage.
	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Searing Light",
		ClassSpellMask: PriestSpellHolyFire,
		ProcChance:     spellData.SearingLight.EffectAt(2).FractionAt(priest.Talents.SearingLight),
		Callback:       core.CallbackOnPeriodicDamageDealt,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			holyPurpose.Activate(sim)
		},
	})
}

// Grants Binding Heal.
func (priest *Priest) applyBindingHeal() {
	if !priest.Talents.BindingHeal {
		return
	}

	spellData.BindingHeal.Each(priest.registerBindingHeal)
}

// When you cast a healing spell, gain Mana equal to 20% of the base cost of the spell if your previous heal was a different spell.
func (priest *Priest) applyLitanyOfLight() {
	if priest.Talents.LitanyOfLight == 0 {
		return
	}

	// One mask per spell, so every rank of a heal is the same previous heal.
	var previousHeal int64
	priest.RegisterResetEffect(func(_ *core.Simulation) {
		previousHeal = 0
	})
	refundFraction := spellData.LitanyOfLight.EffectAt(1).FractionAt(priest.Talents.LitanyOfLight)
	refundMetrics := priest.NewManaMetrics(core.ActionID{SpellID: spellData.LitanyOfLight.Highest().ID})
	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Litany of Light",
		Callback:       core.CallbackOnCastComplete,
		ClassSpellMask: priestLitanyOfLightHeals,
		Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			heal := spell.ClassSpellMask
			if previousHeal != 0 && previousHeal != heal && spell.Cost != nil {
				spell.Unit.AddMana(sim, float64(spell.Cost.BaseCost)*refundFraction, refundMetrics)
			}
			previousHeal = heal
		},
	})
}

// Power Word: Shield absorbs rather than heals, so it is not a healing spell here.
const priestLitanyOfLightHeals = PriestSpellFlashHeal |
	PriestSpellLesserHeal |
	PriestSpellHeal |
	PriestSpellGreaterHeal |
	PriestSpellBindingHeal |
	PriestSpellRenew |
	PriestSpellPenance |
	PriestSpellPrayerOfHealing |
	PriestSpellPrayerOfMending

// Grants spirit of redemption on death
// Acts after death, which a damage sim does not model.
func (priest *Priest) applySpiritOfRedemption() {
	if !priest.Talents.SpiritOfRedemption {
		return
	}
}

// Increases your spell healing by up to X% of your total Spirit and your spell damage by up to Y% of your total Spirit.
// TODO: 14901 excludes Shadowform; unverified whether Shadowform suppresses it in game.
func (priest *Priest) applySpiritualGuidance() {
	if priest.Talents.SpiritualGuidance == 0 {
		return
	}

	dmgCoeff := spellData.SpiritualGuidance.Effect(dbcenums.A_MOD_SPELL_DAMAGE_OF_STAT_PERCENT, 126).FractionAt(priest.Talents.SpiritualGuidance)
	healCoeff := spellData.SpiritualGuidance.Effect(dbcenums.A_MOD_SPELL_HEALING_OF_STAT_PERCENT, 4).FractionAt(priest.Talents.SpiritualGuidance)
	priest.AddStatDependency(stats.Spirit, stats.SpellDamage, dmgCoeff)
	priest.AddStatDependency(stats.Spirit, stats.HealingPower, healCoeff)
}

// Increases the amount healed by your spells by X%.
func (priest *Priest) applySpiritualHealing() {
	if priest.Talents.SpiritualHealing == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.SpiritualHealing.Rank(priest.Talents.SpiritualHealing))
}

// Grants Prayer of Mending.
func (priest *Priest) applyPrayerOfMending() {
	if !priest.Talents.PrayerOfMending {
		return
	}

	spellData.PrayerOfMending.Each(priest.registerPrayerOfMending)
}
