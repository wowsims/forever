package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var earthShockRank = spellData.EarthShock.Highest()
var flameShockRank = spellData.FlameShock.Highest()
var frostShockRank = spellData.FrostShock.Highest()

// The three shocks share category 19, so the row's category cooldown is one timer for all of them.
func (shaman *Shaman) newShockSpellConfig(rank *spelldata.Spell) core.SpellConfig {
	return spelldata.SpellConfig(&shaman.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Flags(SpellFlagShamanSpell|SpellFlagShock|SpellFlagInstant))
}

func (shaman *Shaman) registerEarthShockSpell() {
	config := shaman.newShockSpellConfig(earthShockRank)
	config.ClassSpellMask = SpellMaskEarthShock

	damage := earthShockRank.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	shaman.EarthShock = shaman.RegisterSpell(config)
}

func (shaman *Shaman) registerFlameShockSpell() {
	config := shaman.newShockSpellConfig(flameShockRank)
	config.ClassSpellMask = SpellMaskFlameShock

	config.Dot = spelldata.DotConfig(flameShockRank, flameShockRank.PeriodicEffect())

	damage := flameShockRank.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealDamage(sim, result)
	}

	shaman.FlameShock = shaman.RegisterSpell(config)
}

// The slow, effect 1, is not modelled.
func (shaman *Shaman) registerFrostShockSpell() {
	config := shaman.newShockSpellConfig(frostShockRank)
	config.ClassSpellMask = SpellMaskFrostShock

	damage := frostShockRank.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	shaman.FrostShock = shaman.RegisterSpell(config)
}

func (shaman *Shaman) registerShocks() {
	shaman.registerEarthShockSpell()
	shaman.registerFlameShockSpell()
	shaman.registerFrostShockSpell()
}
