package buffs

import (
	"time"

	"github.com/wowsims/forever/sim/core"
)

func SlowAura(target *core.Unit) *core.Aura {
	return castSlowReductionAura(target, "Slow", 31589, 1.5, time.Second*15)
}

func castSlowReductionAura(target *core.Unit, label string, spellID int32, multiplier float64, duration time.Duration) *core.Aura {
	aura := target.GetOrRegisterAura(core.Aura{Label: label, ActionID: core.ActionID{SpellID: spellID}, Duration: duration})
	aura.NewExclusiveEffect("CastSpdReduction", false, core.ExclusiveEffect{
		// How far from 1 the applied factor is, which is the scale every member of the category bids on.
		Priority: 1 - 1/multiplier,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, 1/multiplier)
			ee.Aura.Unit.MultiplyRangedSpeed(sim, 1/multiplier)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, multiplier)
			ee.Aura.Unit.MultiplyRangedSpeed(sim, multiplier)
		},
	})
	return aura
}

// A slow on casts alone, in the category Slow's cast and ranged slow takes: only the strongest applies.
func CastSpeedReductionEffect(aura *core.Aura, castTimeMultiplier float64) *core.ExclusiveEffect {
	return aura.NewExclusiveEffect("CastSpdReduction", false, core.ExclusiveEffect{
		Priority: 1 - 1/castTimeMultiplier,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, 1/castTimeMultiplier)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, castTimeMultiplier)
		},
	})
}

// The slot every slow on attacks bids in: the one Thunder Clap's melee haste takes, so only the
// strongest applies.
var AtkSpeedReductionCategory = core.ExclusiveStatCategory(ThunderClapCategory, "MeleeSpeedMultiplier", true)

func AtkSpeedReductionEffect(aura *core.Aura, speedMultiplier float64) *core.ExclusiveEffect {
	return aura.NewExclusiveEffect(AtkSpeedReductionCategory, false, core.ExclusiveEffect{
		// How far from 1 the applied factor is, which is the scale every member of the category bids on.
		Priority: 1 - 1/speedMultiplier,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyAttackSpeed(sim, 1/speedMultiplier)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyAttackSpeed(sim, speedMultiplier)
		},
	})
}
