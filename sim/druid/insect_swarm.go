package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var insectSwarmRank = spellData.InsectSwarm.Highest()

func (druid *Druid) registerInsectSwarmSpell() {
	auras := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataInsectSwarmAura(target, true, 0)
	})

	config := spelldata.SpellConfig(&druid.Unit, insectSwarmRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.Dot = spelldata.DotConfig(insectSwarmRank, insectSwarmRank.PeriodicEffect())
	// A tick takes its Spell Piercing bonus at the full share (EV-0142).
	config.Flags |= core.SpellFlagFullResistanceTicks
	// A duration mod such as Nature's Splendor's lengthens the hit debuff through it, as it does the dot.
	config.RelatedAuraArrays = auras.ToMap()

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			auras.Get(target).Activate(sim)
		}
		spell.DealOutcome(sim, result)
	}

	druid.InsectSwarm = druid.RegisterSpell(config)
}
