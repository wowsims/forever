package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var earthShockRanks = castableRanks(spellData.EarthShock)
var flameShockRanks = castableRanks(spellData.FlameShock)
var frostShockRanks = castableRanks(spellData.FrostShock)

// The three shocks share category 19, so the spell's category cooldown is one timer for all of them.
func (shaman *Shaman) newShockSpellConfig(rank *spelldata.Spell) core.SpellConfig {
	return spelldata.SpellConfig(&shaman.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Flags(SpellFlagShamanSpell|SpellFlagShock|SpellFlagInstant))
}

func (shaman *Shaman) registerEarthShockSpell(rank *spelldata.Spell) {
	config := shaman.newShockSpellConfig(rank)
	config.ClassSpellMask = SpellMaskEarthShock

	damage := rank.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	shaman.EarthShock = append(shaman.EarthShock, shaman.RegisterSpell(config))
}

func (shaman *Shaman) registerFlameShockSpell(rank *spelldata.Spell) {
	config := shaman.newShockSpellConfig(rank)
	config.ClassSpellMask = SpellMaskFlameShock

	config.Dot = spelldata.DotConfig(rank, rank.PeriodicEffect())

	damage := rank.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealDamage(sim, result)
	}

	shaman.FlameShock = append(shaman.FlameShock, shaman.RegisterSpell(config))
}

// The slow, effect 1, is not modelled.
func (shaman *Shaman) registerFrostShockSpell(rank *spelldata.Spell) {
	config := shaman.newShockSpellConfig(rank)
	config.ClassSpellMask = SpellMaskFrostShock

	damage := rank.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	shaman.FrostShock = append(shaman.FrostShock, shaman.RegisterSpell(config))
}

func (shaman *Shaman) registerShocks() {
	earthShockRanks.Each(func(_ int32, rank *spelldata.Spell) { shaman.registerEarthShockSpell(rank) })
	flameShockRanks.Each(func(_ int32, rank *spelldata.Spell) { shaman.registerFlameShockSpell(rank) })
	frostShockRanks.Each(func(_ int32, rank *spelldata.Spell) { shaman.registerFrostShockSpell(rank) })
}
