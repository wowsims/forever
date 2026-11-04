package warlock

import (
	"github.com/wowsims/forever/sim/core"
)

// Effect 1 (A_CHANNEL_DEATH_ITEM) returns a Soul Shard when the target dies with the channel on it,
// whoever lands the killing blow. The row's PROC_FLAG_KILL at 0% chance grants nothing itself; Soul
// Harvest reads that kill.
// TODO: The tooltip's chance of a Soul Shard from a target Drain Soul damages has no row behind it and is not modelled.
func (warlock *Warlock) registerDrainSoul() {
	rankedSpell := spellData.DrainSoul.ByID(11675)
	drainSoulTick := rankedSpell.PeriodicEffect()

	config := warlock.soulSiphonChannelConfig(rankedSpell, drainSoulTick, nil)
	config.ClassSpellMask = WarlockSpellDrainSoul

	warlock.DrainSoul = warlock.RegisterSpell(config)

	warlock.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Drain Soul - Soul Shard",
		Callback: core.CallbackOnTargetDies,
		ExtraCondition: func(_ *core.Simulation, _ *core.Spell, result *core.SpellResult) bool {
			return warlock.DrainSoul.Dot(result.Target).IsActive()
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			warlock.gainSoulShard(sim)
		},
	})
}
