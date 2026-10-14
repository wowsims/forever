package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerFade(_ int32, rank *spelldata.Spell) {
	threatDrop := rank.EffectN(1).Average()
	var spell *core.Spell

	// A_MOD_TOTAL_THREAT lowers the priest's threat on every enemy while the aura is up and gives it
	// back when it ends. The sim keeps threat as each spell's running total, so the drop is booked
	// against Fade on gain and returned on expire.
	auraConfig := spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &priest.Unit))
	auraConfig.OnGain = func(_ *core.Aura, _ *core.Simulation) {
		spell.ApplyAOEThreatIgnoreMultipliers(threatDrop)
	}
	auraConfig.OnExpire = func(_ *core.Aura, _ *core.Simulation) {
		spell.ApplyAOEThreatIgnoreMultipliers(-threatDrop)
	}
	aura := priest.RegisterAura(auraConfig)
	aura.NewExclusiveEffect(rank.Name, true, core.ExclusiveEffect{})

	spell = priest.registerSelfBuffSpell(rank, aura)
}
