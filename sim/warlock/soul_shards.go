package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

const soulShardItemID = 6265

// The Soul Shards carried, as the stacks of an aura on the warlock, so the timeline and the APL's aura stack
// values read them. Of the reagents the warlock's rows name, only the Soul Shard is counted: the Infernal
// Stone (5565, Inferno) and the Demonic Figurine (16583, Ritual of Doom) belong to spells the sim does not
// cast. Each iteration starts with the shards carried into the fight. No shard comes back from a
// kill, Drain Soul's or Shadowburn's within 8 sec, as the target does not die. With none left the aura
// falls off, and a refund brings it back.
func (warlock *Warlock) registerSoulShards() {
	carried := warlock.Options.SoulShards
	warlock.SoulShardsAura = warlock.RegisterAura(core.Aura{
		Label:     "Soul Shards",
		ActionID:  core.ActionID{ItemID: soulShardItemID},
		Duration:  core.NeverExpires,
		MaxStacks: carried,
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

// The cast needs and spends the Soul Shards its row states as reagents. They are checked where the sim
// checks mana, when the cast starts and when it completes, and spent first in ApplyEffects, which the sim
// runs when the cast completes, right after it spends the mana. Wraps the config's ExtraCastCondition and
// ApplyEffects, so it is called after both are set.
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
