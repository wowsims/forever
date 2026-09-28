package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var insectSwarmRank = spellData.InsectSwarm.Highest()

func (druid *Druid) registerInsectSwarmSpell() {
	// Binary: the row's second effect lowers the target's chance to hit, and a spell with a
	// non-damage effect is resisted whole or not at all. That hit debuff is not modelled.
	config := spelldata.SpellConfig(&druid.Unit, insectSwarmRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = DruidSpellInsectSwarm
	config.Dot = spelldata.DotConfig(insectSwarmRank, insectSwarmRank.PeriodicEffect())

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	druid.InsectSwarm = druid.RegisterSpell(config)
}
