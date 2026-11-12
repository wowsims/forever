package warlock

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Bane of Agony's base ramps in thirds of its duration by the time since it landed, also under Wrack's added
// ticks: a tick in the first third, its end included, deals half, one in the second the base, and one in the last
// one and a half. The spell power share and the damage multipliers are read as each tick lands, and the share does
// not ramp. Amplify Curse, which the cast spends, raises the whole tick, base and share, for the whole of the Bane
// it landed with. The share read at each tick comes from the Nov 3 patch notes; the game files do not state it.
// TODO: Ingame test needed: confirm on live that each tick reads the spell power of the moment.
func (warlock *Warlock) registerBaneOfAgony() {
	rankedSpell := spellData.BaneOfAgony.ByID(11713)
	agonyTick := rankedSpell.PeriodicEffect()

	// The Amplify Curse bonus each target's Bane landed with.
	ampByTarget := map[int32]float64{}

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellBaneOfAgony

	config.Dot = spelldata.DotConfig(rankedSpell, agonyTick)

	// The bonus of a cast now, if Amplify Curse is up.
	ampNow := func() float64 {
		if warlock.AmplifyCurseAura.IsActive() {
			return warlock.AmplifyCurseMultiplier
		}
		return 1
	}

	// Dot.Apply snapshots on every application, whatever applied the Bane, and runs inside the cast's
	// effects, before the cast spends Amplify Curse.
	config.Dot.OnSnapshot = func(_ *core.Simulation, target *core.Unit, _ *core.Dot) {
		ampByTarget[target.UnitIndex] = ampNow()
	}

	// Sets the dot's snapshot to a tick at the ramp, with the spell power and damage multipliers of now.
	setTick := func(target *core.Unit, dot *core.Dot, ampMult float64, ramp float64) {
		attackTable := warlock.AttackTables[target.UnitIndex]
		spellPowerShare := dot.BonusCoefficient * dot.Spell.BonusDamage(attackTable)
		dot.SnapshotBaseDamage = (agonyTick.Average()*ramp + spellPowerShare) * ampMult
		dot.SnapshotAttackerMultiplier = dot.Spell.AttackerDamageMultiplier(attackTable, true) * dot.PeriodicDamageMultiplier
	}

	// A dot activated without Apply has no entry, and ticks as if cast now.
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		ampMult, ok := ampByTarget[target.UnitIndex]
		if !ok {
			ampMult = ampNow()
		}
		setTick(target, dot, ampMult, baneOfAgonyRamp(sim.CurrentTime-dot.StartedAt(), dot.ExpiresAt()-dot.StartedAt()))
		dot.CalcAndDealPeriodicSnapshotDamage(sim, target, rankedSpell.TickOutcome(dot))
	}

	// Roll to hit and apply as dot in place of any other Bane
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			warlock.deactivateOtherBanes(sim, spell, target)
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	// Calculate expected damage from the average tick, the base at a ramp of one, which is what a tick is worth
	// over the whole duration: with the Amplify Curse bonus of the dot on the target (useSnapshot), or of a cast
	// now, so Dot Percent Increase sees what re-applying gains. Each tick sets the snapshot again, so setting it
	// here costs a running dot nothing.
	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
		dot := spell.Dot(target)
		ampMult, ok := ampByTarget[target.UnitIndex]
		if !useSnapshot || !ok {
			ampMult = ampNow()
		}
		setTick(target, dot, ampMult, 1)
		return dot.CalcSnapshotDamage(sim, target, rankedSpell.ExpectedMagicTickOutcome(spell))
	}

	warlock.BaneOfAgony = warlock.RegisterSpell(config)
	warlock.Banes = append(warlock.Banes, warlock.BaneOfAgony)
	warlock.afflictionDots = append(warlock.afflictionDots, warlock.BaneOfAgony)
}

// The share of the base a tick deals, by the third of the duration its time since the Bane landed falls in. A
// tick on a third's end deals that third's share.
func baneOfAgonyRamp(elapsed time.Duration, duration time.Duration) float64 {
	switch {
	case 3*elapsed <= duration:
		return 0.5
	case 3*elapsed <= 2*duration:
		return 1
	default:
		return 1.5
	}
}
