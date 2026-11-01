package shaman

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var windfuryTotemRank = spellData.WindfuryTotem.Highest()

// Strength of Earth's rank 4 and Grace of Air's rank 2 are the last ranks the trainer teaches; the Tablet of
// Strength of Earth Totem V and Tablet of Grace of Air Totem III books teach the ranks above them.
var strengthOfEarthTotemTrainerRank = spellData.StrengthOfEarthTotem.ByID(10442)
var strengthOfEarthTotemBookRank = spellData.StrengthOfEarthTotem.ByID(25361)
var graceOfAirTotemTrainerRank = spellData.GraceOfAirTotem.ByID(10627)
var graceOfAirTotemBookRank = spellData.GraceOfAirTotem.ByID(25359)

var manaSpringTotemRank = spellData.ManaSpringTotem.Highest()
var flametongueTotemRank = spellData.FlametongueTotem.Highest()
var frostResistanceTotemRank = spellData.FrostResistanceTotem.Highest()

// A totem's cast: the row's cost, GCD and cooldown, and nothing it hits.
func (shaman *Shaman) newTotemSpellConfig(rank *spelldata.Spell) core.SpellConfig {
	config := spelldata.SpellConfig(&shaman.Unit, rank, spelldata.Flags(core.SpellFlagAPL|SpellFlagShamanSpell|SpellFlagInstant))
	config.ProcMask = core.ProcMaskEmpty
	return config
}

// Stands a totem in its element's slot for the totem row's duration, taking down the one there.
func (shaman *Shaman) dropTotem(sim *core.Simulation, slot int, aura *core.Aura, duration time.Duration) {
	if current := shaman.TotemAuras[slot]; current != nil && current != aura {
		current.Deactivate(sim)
	}
	shaman.TotemAuras[slot] = aura
	shaman.TotemExpirations[slot] = sim.CurrentTime + duration
	aura.Activate(sim)
}

// A totem the sim models as the buff it gives its caster, lasting as long as the totem stands.
func (shaman *Shaman) registerBuffTotem(rank *spelldata.Spell, slot int, buff *core.Aura) *core.Spell {
	duration := rank.Duration()
	buff.Duration = duration

	config := shaman.newTotemSpellConfig(rank)
	config.ClassSpellMask = SpellMaskBasicTotem
	config.RelatedSelfBuff = buff
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		shaman.dropTotem(sim, slot, buff, duration)
	}

	return shaman.RegisterSpell(config)
}

// A weapon totem stands as an aura of the totem row's duration that holds the totem's category, which
// turns its proc on. The proc is core/buffs', shared with the raid buff of the same totem.
func (shaman *Shaman) registerWeaponTotem(rank *spelldata.Spell, slot int, category string, priority float64, procTrigger *core.Aura) *core.Spell {
	duration := rank.Duration()

	totemAura := shaman.RegisterAura(core.Aura{
		Label:    rank.Name + " (Self)",
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: duration,
	})
	totemAura.NewExclusiveEffect(category, false, core.ExclusiveEffect{
		Priority: priority,
		OnGain: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
			procTrigger.Activate(sim)
		},
		OnExpire: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
			procTrigger.Deactivate(sim)
		},
	})

	config := shaman.newTotemSpellConfig(rank)
	config.ClassSpellMask = SpellMaskBasicTotem
	config.RelatedSelfBuff = totemAura
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		shaman.dropTotem(sim, slot, totemAura, duration)
	}
	return shaman.RegisterSpell(config)
}

// The rows already carry Improved Weapon Totems: Windfury's attack power and Flametongue's damage are
// its 30% and 12% over their untalented values, so nothing applies the talent.
func (shaman *Shaman) registerWindfuryTotemSpell() {
	shaman.WindfuryTotem = shaman.registerWeaponTotem(windfuryTotemRank, AirTotem, buffs.WindfuryTotemCategory,
		buffs.WindfuryTotemValue(0), buffs.NewWindfuryTotemProcTrigger(&shaman.Character))
	buffs.HoldOverWindfuryTotemRivals(shaman.WindfuryTotem.RelatedSelfBuff)
}

// Under a standing Windfury Totem, Flametongue Totem's proc does nothing.
func (shaman *Shaman) registerFlametongueTotemSpell() {
	shaman.FlametongueTotem = shaman.registerWeaponTotem(flametongueTotemRank, FireTotem, buffs.FlametongueTotemCategory,
		buffs.FlametongueTotemPriority(), buffs.NewFlametongueTotemProcTrigger(&shaman.Character))
}

func (shaman *Shaman) registerStrengthOfEarthTotemSpell() {
	// The rows already carry Enhancing Totems: Strength of Earth's and Grace of Air's amounts are its 15%
	// over their untalented values, so nothing applies the talent.
	shaman.registerBuffTotem(strengthOfEarthTotemTrainerRank, EarthTotem,
		buffs.NewSpellDataStrengthOfEarthTotemRankAura(&shaman.Unit, true, spellData.StrengthOfEarthTotemTriggered.ByID(10441)))
	shaman.registerBuffTotem(strengthOfEarthTotemBookRank, EarthTotem,
		buffs.NewSpellDataStrengthOfEarthTotemAura(&shaman.Unit, true, 0))
}

func (shaman *Shaman) registerGraceOfAirTotemSpell() {
	shaman.registerBuffTotem(graceOfAirTotemTrainerRank, AirTotem,
		buffs.NewSpellDataGraceOfAirTotemRankAura(&shaman.Unit, true, spellData.GraceOfAirTotemTriggered.ByID(10626)))
	shaman.registerBuffTotem(graceOfAirTotemBookRank, AirTotem,
		buffs.NewSpellDataGraceOfAirTotemAura(&shaman.Unit, true, 0))
}

// The aura reads Restorative Totems' first effect, the Mana Spring share, from the talent points.
func (shaman *Shaman) registerManaSpringTotemSpell() {
	shaman.registerBuffTotem(manaSpringTotemRank, WaterTotem,
		buffs.NewSpellDataManaSpringTotemAura(&shaman.Unit, true, shaman.Talents.RestorativeTotems))
}

// A fire totem, so it stands in the slot Fire Nova radiates from.
func (shaman *Shaman) registerFrostResistanceTotemSpell() {
	shaman.registerBuffTotem(frostResistanceTotemRank, FireTotem,
		buffs.NewSpellDataFrostResistanceTotemAura(&shaman.Unit, true, 0))
}

// Commented out upstream of this fork by "only 1 totem per type" (53970b1d84), not by
// the Forever stubbing pass, and never called since. It is kept rather than reduced to
// the usual empty-body no-op because the body is the only record of how the totem was
// modelled; shaman.HealingStreamTotem stays nil until someone revisits totem slots.
/* func (shaman *Shaman) registerHealingStreamTotemSpell() {
	config := shaman.newTotemSpellConfig(3, 5394, SpellMaskBasicTotem, time.Second)
	hsHeal := shaman.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 5394},
		SpellSchool:      core.SpellSchoolNature,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | SpellFlagInstant,
		DamageMultiplier: 1,
		CritMultiplier:   1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			healing := 28 + spell.HealingPower(target)*0.08272
			spell.CalcAndDealHealing(sim, target, healing, spell.OutcomeHealing)
		},
	})
	config.Hot = core.DotConfig{
		Aura: core.Aura{
			Label: "HealingStreamHot",
		},
		NumberOfTicks: 150,
		TickLength:    time.Second * 2,
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			hsHeal.Cast(sim, target)
		},
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + time.Second*300
		for _, agent := range shaman.Party.Players {
			spell.Hot(&agent.GetCharacter().Unit).Activate(sim)
		}
	}
	shaman.HealingStreamTotem = shaman.RegisterSpell(config)
} */
