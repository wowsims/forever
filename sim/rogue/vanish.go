package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var vanishRank = spellData.Vanish.Highest()

func (rogue *Rogue) registerVanishSpell() {
	config := spelldata.SpellConfig(&rogue.Unit, vanishRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMCD))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.AutoAttacks.CancelAutoSwing(sim)
		rogue.StealthAura.Activate(sim)
	}

	rogue.Vanish = rogue.RegisterSpell(config)

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.Vanish,
		Type:  core.CooldownTypeDPS,
	})

	// The TBC implementation, kept for the port:
	// rogue.Vanish = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: vanishRank.ID},
	// 	SpellSchool:    vanishRank.SpellSchool(),
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellVanish,
	// 	ClassFlags:     vanishRank.ClassFlags,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: 0,
	// 		},
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    rogue.NewTimer(),
	// 			Duration: max(vanishRank.Cooldown(), vanishRank.CategoryCooldown()),
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		// Pause auto attacks
	// 		rogue.AutoAttacks.CancelAutoSwing(sim)
	// 		// Apply stealth
	// 		rogue.StealthAura.Activate(sim)
	// 	},
	// })
}
