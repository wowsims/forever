package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shieldWallRank = spellData.ShieldWall.Highest()

func (warrior *Warrior) registerShieldWall() {
	aura := warrior.RegisterAura(spelldata.AuraConfig(shieldWallRank))
	spelldata.ParseEffects(&warrior.Character, aura, shieldWallRank)

	config := spelldata.SpellConfig(&warrior.Unit, shieldWallRank)

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		aura.Activate(sim)
	}

	config.RelatedSelfBuff = aura

	spell := warrior.RegisterSpell(config)

	warrior.deactivateWithoutShield(aura)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeSurvival,
		ShouldActivate: func(s *core.Simulation, c *core.Character) bool {
			return warrior.CurrentHealthPercent() < 0.4
		},
	})
}
