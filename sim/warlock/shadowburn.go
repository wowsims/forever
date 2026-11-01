package warlock

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Granted by the Shadowburn talent. The 15 sec cooldown is the row's category cooldown, and the Soul Shard
// each cast takes the row's reagent. A hit that lands leaves effect 1 (A_CHANNEL_DEATH_ITEM) on the target
// for the row's 8 sec, and the target dying under it returns a Soul Shard, whoever lands the killing blow.
// The effect is kept as the time its window ends on each target, not as an aura: an aura's expiry on the
// target would add a step to every fight and shift the order of the events at that time.
func (warlock *Warlock) registerShadowBurn() {
	rankedSpell := spellData.Shadowburn.ByID(18871)
	shadowburnHit := rankedSpell.DamageEffect()

	deathItemEnds := make([]time.Duration, len(warlock.Env.AllUnits))
	warlock.RegisterResetEffect(func(_ *core.Simulation) {
		clear(deathItemEnds)
	})
	// The window goes with the target that dies, so a later spawn in the same pool slot starts without it.
	warlock.RegisterAura(*core.MakePermanent(&core.Aura{
		Label: "Shadowburn - Soul Shard",
		OnTargetDies: func(_ *core.Aura, sim *core.Simulation, target *core.Unit) {
			if sim.CurrentTime < deathItemEnds[target.UnitIndex] {
				warlock.gainSoulShard(sim)
			}
			deathItemEnds[target.UnitIndex] = 0
		},
	}))

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellShadowBurn

	// Roll the damage and to hit
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, shadowburnHit.Average(), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			deathItemEnds[target.UnitIndex] = sim.CurrentTime + rankedSpell.Duration()
		}
	}

	// Calculate expected damage from the row's average hit
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, shadowburnHit.Average(), spell.OutcomeExpectedMagicCrit)
	}

	warlock.useReagents(&config, rankedSpell)
	warlock.Shadowburn = warlock.RegisterSpell(config)
}
