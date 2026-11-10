package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var volleyRank = spellData.Volley.Highest()

// A wave's damage is its own row, the one the cast row's tooltip names.
var volleyWaveRank = volleyRank.Refs()[0]

// A ground channel: each tick of the cast row's A_PERIODIC_DUMMY casts the wave row at every enemy,
// the first a tick after the cast. Cast speed shortens the ticks and ranged haste does not. A wave
// takes 3% of ranged attack power against its target, the share the cast row's first effect states
// as a spell power coefficient, and rolls the ranged hit and crit tables. In game the missile's
// flight delays the channel's aura but not the waves, so the channel starts at the cast.
func (hunter *Hunter) registerVolley() {
	waveBase := volleyWaveRank.DamageEffect().Average()
	apShare := volleyRank.EffectN(1).Coeff()

	waveConfig := spelldata.SpellConfig(&hunter.Unit, volleyWaveRank, spelldata.Melee(core.ProcMaskRangedSpecial), spelldata.Proc())
	if volleyWaveRank.IsAProc() {
		waveConfig.Flags |= core.SpellFlagProc
	}
	waveConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, target := range sim.Encounter.ActiveTargetUnits {
			spell.CalcAndDealDamage(sim, target, waveBase+apShare*spell.RangedAttackPower(target), spell.OutcomeRangedHitAndCrit)
		}
	}
	wave := hunter.RegisterSpell(waveConfig)

	config := spelldata.SpellConfig(&hunter.Unit, volleyRank, spelldata.Melee(core.ProcMaskEmpty))
	config.Dot = spelldata.DotConfig(volleyRank, volleyRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0))
	config.Dot.IsAOE = true
	// Cast speed shortens the ticks and the channel with them; the wave count stays the row's.
	config.Dot.AffectedByCastSpeed = true
	config.Dot.HasteReducesDuration = true
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
		wave.Cast(sim, target)
	}
	// Auto Shot winds up no shot while the channel runs. Its swing timer keeps running, so a shot
	// that came due winds up when the channel ends, whether it runs out or is cut short.
	config.Dot.Aura.OnExpire = func(_ *core.Aura, sim *core.Simulation) {
		hunter.AutoAttacks.ResumeRangedAt(sim, sim.CurrentTime)
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		channel := spell.AOEDot()
		channel.Apply(sim)
		hunter.AutoAttacks.DelayRangedUntil(sim, sim.CurrentTime+channel.Duration)
	}

	hunter.RegisterSpell(config)
}
