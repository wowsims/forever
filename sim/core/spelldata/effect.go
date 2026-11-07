package spelldata

import (
	"math"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
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
// and the per-level gain is added on top of the fraction.
//
// A damage or heal amount keeps the fraction: in game the amount dealt rounds up with a chance equal
// to the fraction of the whole amount, power and multipliers included, so it averages the exact
// value. Every other amount, a stat, a resource, an absorb or a threat, is the whole number the
// tooltip shows, the floor.
//
// float32 is load-bearing: EffectRealPointsPerLevel is a float32 widened into the DB
// (3.79999995231628), and multiplying in float64 moves the result off the tooltip on six rows.
func (e *Effect) AverageAt(level int32) float64 {
	if e.damageOrHeal() {
		return e.exactAt(level)
	}
	return math.Floor(e.exactAt(level))
}

func (e *Effect) exactAt(level int32) float64 {
	return float64(float32(e.BasePoints) + e.levelGain(level))
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

	delta := max(lvl-int32(e.BaseLevel), 0)

	return float32(float32(delta) * float32(e.PPL))
}

// The low end of the roll. EffectVariance spreads the base points, whose ends are rounded to whole
// numbers, a half to the even one; the per-level gain is added to both, and each end is rounded
// outward. An effect that does not roll answers the floored amount at both ends. For a spell-school
// hit or a heal these are the ends the game files state, which core spreads over the whole amount
// instead (DirectVariance).
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
		whole := math.Floor(e.exactAt(level))
		return whole, whole
	}
	base := float32(e.BasePoints)
	// A physical hit rolls its base points by the whole Variance either side; anything else by half.
	side := float32(float32(math.Abs(e.BasePoints)) * float32(e.Variance) / 2)
	if e.physicalDamage() {
		side *= 2
	}
	gain := e.levelGain(level)
	low := float32(math.RoundToEven(float64(base-side))) + gain
	high := float32(math.RoundToEven(float64(base+side))) + gain
	return math.Floor(float64(low)), math.Ceil(float64(high))
}

// The amount for one cast: rolled where the game files state a spread, the average where they do not.
func (e *Effect) RollAt(sim *core.Simulation, level int32) float64 {
	if e.Variance == 0 {
		return e.AverageAt(level)
	}
	return sim.Roll(e.spread(level))
}

// What a cast averages before power: the average for an effect that does not roll and for a
// spell-school hit or a heal, whose roll core centres on it; elsewhere the middle of the roll, which
// can sit half a point off Average.
func (e *Effect) MeanAt(level int32) float64 {
	if e.Variance == 0 || e.DirectVariance() > 0 {
		return e.AverageAt(level)
	}
	low, high := e.spread(level)
	return (low + high) / 2
}

// The amount to hand CalcDamage or CalcHealing for one cast: the average where core rolls the whole
// amount (DirectVariance), one roll of the base points otherwise.
func (e *Effect) DirectAmount(sim *core.Simulation) float64 {
	if e.DirectVariance() > 0 {
		return e.Average()
	}
	return e.Roll(sim)
}

// What core.SpellConfig.DirectVariance takes: the Variance of a spell-school hit or of a heal of any
// school, which roll their whole amount once power is added; 0 for every other effect.
func (e *Effect) DirectVariance() float64 {
	switch {
	case e.Type == dbcenums.E_HEAL:
		return e.Variance
	case e.Type == dbcenums.E_SCHOOL_DAMAGE && Find(e.SpellID).SpellSchool() != core.SpellSchoolPhysical:
		return e.Variance
	}
	return 0
}

func (e *Effect) damageOrHeal() bool {
	switch e.Type {
	case dbcenums.E_SCHOOL_DAMAGE, dbcenums.E_HEAL, dbcenums.E_HEALTH_LEECH, dbcenums.E_WEAPON_DAMAGE,
		dbcenums.E_NORMALIZED_WEAPON_DMG, dbcenums.E_WEAPON_DAMAGE_NOSCHOOL:
		return true
	}
	switch e.Aura {
	case dbcenums.A_PERIODIC_DAMAGE, dbcenums.A_PERIODIC_HEAL, dbcenums.A_PERIODIC_LEECH,
		dbcenums.A_DAMAGE_SHIELD, dbcenums.A_PROC_TRIGGER_DAMAGE:
		return true
	}
	return false
}

func (e *Effect) physicalDamage() bool {
	switch e.Type {
	case dbcenums.E_SCHOOL_DAMAGE, dbcenums.E_WEAPON_DAMAGE, dbcenums.E_WEAPON_PERCENT_DAMAGE,
		dbcenums.E_NORMALIZED_WEAPON_DMG, dbcenums.E_WEAPON_DAMAGE_NOSCHOOL:
		return Find(e.SpellID).SpellSchool() == core.SpellSchoolPhysical
	}
	return false
}
