package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerProwlSpell() {
	rank := spellData.Prowl.Highest()
	druid.ProwlAura = druid.RegisterAura(spelldata.AuraConfig(rank))
	speed := 1 + rank.EffectN(2).Percent()
	// Attribute word 0 cancels auto attack (0x100000), which core does not read, so the aura pauses
	// melee until a cast ends it. Dash carries ATTR_EX_1_NO_STEALTH_BREAK and does not.
	// The not-in-combat bit keeps Prowl itself before the pull.
	druid.ProwlAura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		druid.MultiplyMovementSpeed(sim, speed)
		druid.AutoAttacks.PauseMeleeBy(sim, core.NeverExpires)
	})
	druid.ProwlAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.MultiplyMovementSpeed(sim, 1/speed)
		druid.AutoAttacks.ResumeMeleeAt(sim, 0, 0)
	})
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return !druid.ProwlAura.IsActive()
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		druid.ProwlAura.Activate(sim)
	}
	druid.Prowl = druid.RegisterSpell(config)
}
