package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

const soulShardItemID = 6265
const soulShardsMax = 32

// The Soul Shards carried, as the stacks of an aura on the warlock, so the timeline and the APL's aura stack
// values read them. Of the reagents the warlock's spells name, only the Soul Shard is counted: the Infernal
// Stone (5565, Inferno) and the Demonic Figurine (16583, Ritual of Doom) belong to spells the sim does not
// cast. A warlock carries at most soulShardsMax, so a shard gained at the cap is lost.
func (warlock *Warlock) registerSoulShards() {
	carried := warlock.Options.SoulShards
	warlock.SoulShardsAura = warlock.RegisterAura(core.Aura{
		Label:     "Soul Shards",
		ActionID:  core.ActionID{ItemID: soulShardItemID},
		Duration:  core.NeverExpires,
		MaxStacks: soulShardsMax,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			if carried > 0 {
				aura.Activate(sim)
				aura.SetStacks(sim, carried)
			}
		},
	})
}

// An effect that lets the spells its class flags name cast without their reagents while its aura is up.
type reagentWaiver struct {
	aura  *core.Aura
	flags core.ClassFlags
}

// The cast needs and uses the Soul Shards its spell states as reagents. Wraps the config's
// ExtraCastCondition and ApplyEffects, so it is called after both are set.
func (warlock *Warlock) useReagents(config *core.SpellConfig, row *spelldata.Spell) {
	shards := row.ReagentCount(soulShardItemID)
	if shards == 0 {
		return
	}

	waived := func() bool {
		for _, waiver := range warlock.reagentWaivers {
			if waiver.aura.IsActive() && waiver.flags.Matches(row.ClassFlags) {
				return true
			}
		}
		return false
	}

	extraCastCondition := config.ExtraCastCondition
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return (extraCastCondition == nil || extraCastCondition(sim, target)) &&
			(waived() || warlock.SoulShardsAura.GetStacks() >= shards)
	}

	applyEffects := config.ApplyEffects
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if !waived() {
			warlock.SoulShardsAura.RemoveStacks(sim, shards)
		}
		applyEffects(sim, target, spell)
	}
}

func (warlock *Warlock) gainSoulShard(sim *core.Simulation) {
	if !warlock.SoulShardsAura.IsActive() {
		warlock.SoulShardsAura.Activate(sim)
	}
	warlock.SoulShardsAura.AddStack(sim)
}
