package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var evocationRank = spellData.Evocation.Highest()

// While channeling, spirit regen is increased by 1500% (effect 1) and all of it continues while
// casting (effect 2). Both are parsed onto the channel's aura; the casting share stops at all of
// the spirit regen, so Mage Armor and Arcane Meditation add nothing during it.
func (mage *Mage) registerEvocation() {
	config := spelldata.SpellConfig(&mage.Unit, evocationRank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = MageSpellEvocation

	config.Hot = spelldata.DotConfig(evocationRank, evocationRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0))
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(_ *core.Simulation, _ *core.Unit, _ *core.Dot) {}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.SelfHot().Apply(sim)
	}

	spell := mage.RegisterSpell(config)
	spelldata.ParseEffects(&mage.Character, spell.SelfHot().Aura, evocationRank, spelldata.Effects(1, 2))

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
			return character.CurrentManaPercent() < 0.1
		},
	})
}
