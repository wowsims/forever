package spelldata

import (
	"math"
	"time"

	"github.com/wowsims/forever/sim/core"
)

// The game files' own number, before any unit the aura gives it.
func (e *Effect) BaseValue() float64 {
	return e.BasePoints
}

// The game files state a percentage as an integer: Improved Righteous Fury reads 16, not 0.16.
func (e *Effect) Percent() float64 {
	return e.BasePoints / 100
}

// The game files state rage on a 0-1000 bar: a -30 cost modifier is 3 rage, Charge's energize of
// 150 is 15.
func (e *Effect) Tenths() float64 {
	return e.BasePoints / 10
}

// The value read as a time, which is the unit a duration or delay modifier states it in.
func (e *Effect) TimeValue() time.Duration {
	return core.DurationFromMillis(e.BasePoints)
}

// EffectAuraPeriod: how long one tick of the aura lasts.
func (e *Effect) Period() time.Duration {
	return core.DurationFromMillis(e.PeriodMs)
}

func (e *Effect) Coeff() float64 {
	return e.SPCoef
}

func (e *Effect) APCoeff() float64 {
	return e.APCoef
}

// The spell this effect fires, or Nil where it fires nothing.
func (e *Effect) Trigger() *Spell {
	return Find(e.TriggerID)
}

// The amount at a caster level: the base points plus the per-level gain over the spell's BaseLevel,
// stopping at the level the spell stops scaling at. Both levels are the owning spell's, stamped onto
// the effect at generation, so an effect answers the same amount however the caller reached it.
//
// The base keeps the game files' own EffectBasePointsF, which is fractional on about one effect in 55,
// and the per-level gain is added on top of the fraction before the whole is floored.
//
// float32 is load-bearing: EffectRealPointsPerLevel is a float32 widened into the DB
// (3.79999995231628), and multiplying in float64 moves the result off the tooltip on six rows. The
// game resolves the amount to a whole number, which is the floor below.
func (e *Effect) AverageAt(level int32) float64 {
	return math.Floor(float64(float32(e.BasePoints) + e.levelGain(level)))
}

// Average, Min, Max, Mean and Roll read the sim's character level; the At forms take one.
func (e *Effect) Average() float64 {
	return e.AverageAt(core.CharacterLevel)
}

func (e *Effect) Min() float64 {
	return e.MinAt(core.CharacterLevel)
}

func (e *Effect) Max() float64 {
	return e.MaxAt(core.CharacterLevel)
}

func (e *Effect) Mean() float64 {
	return e.MeanAt(core.CharacterLevel)
}

func (e *Effect) Roll(sim *core.Simulation) float64 {
	return e.RollAt(sim, core.CharacterLevel)
}

func (e *Effect) levelGain(level int32) float32 {
	// MaxLevel 0 is the game files' "no cap", which a spell with no SpellLevels row carries.
	lvl := level
	if e.MaxLevel > 0 && int32(e.MaxLevel) < lvl {
		lvl = int32(e.MaxLevel)
	}

	delta := lvl - int32(e.BaseLevel)
	if delta < 0 {
		delta = 0
	}

	return float32(float32(delta) * float32(e.PPL))
}

// The low end of the roll. EffectVariance spreads the base points, whose ends are rounded to whole
// numbers, a half to the even one; the per-level gain is added to both, and each end is rounded
// outward. An effect that does not roll answers the average at both ends.
func (e *Effect) MinAt(level int32) float64 {
	low, _ := e.spread(level)
	return low
}

func (e *Effect) MaxAt(level int32) float64 {
	_, high := e.spread(level)
	return high
}

// float32 for the reason Average is: Wrath rank 3's 30 ± 2.5 ends on halves there, 27.5 and 32.5,
// which float64 reads a hair to either side of.
func (e *Effect) spread(level int32) (float64, float64) {
	if e.Variance == 0 {
		average := e.AverageAt(level)
		return average, average
	}
	base := float32(e.BasePoints)
	half := float32(float32(math.Abs(e.BasePoints)) * float32(e.Variance) / 2)
	gain := e.levelGain(level)
	low := float32(math.RoundToEven(float64(base-half))) + gain
	high := float32(math.RoundToEven(float64(base+half))) + gain
	return math.Floor(float64(low)), math.Ceil(float64(high))
}

// The amount for one cast: rolled where the game files state a spread, the average where they do not.
func (e *Effect) RollAt(sim *core.Simulation, level int32) float64 {
	if e.Variance == 0 {
		return e.AverageAt(level)
	}
	return sim.Roll(e.spread(level))
}

// What a cast averages: the middle of the roll where the game files state a spread, which can sit
// half a point above Average, and the average where they do not.
func (e *Effect) MeanAt(level int32) float64 {
	low, high := e.spread(level)
	return (low + high) / 2
}
