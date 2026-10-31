package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The tooltip's "up to 10 Rage per second". The row states only the health each point of rage heals,
// as effect 1's $s1 percent; its periodic trigger names no spell, and the heal row 22845 states 0.
const frenziedRegenerationRagePerTick = 10.0

func (druid *Druid) registerFrenziedRegenerationSpell() {
	rank := spellData.FrenziedRegeneration.Highest()
	tick := rank.PeriodicEffect()
	rageMetrics := druid.NewRageMetrics(core.ActionID{SpellID: rank.ID})

	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ProcMask = core.ProcMaskSpellHealing
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1
	config.Hot = spelldata.DotConfig(rank, tick)
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		rage := min(druid.CurrentRage(), frenziedRegenerationRagePerTick)
		if rage <= 0 {
			return
		}
		druid.SpendRage(sim, rage, rageMetrics)
		dot.Spell.CalcAndDealPeriodicHealing(sim, target, rage*tick.Percent()*druid.MaxHealth(), dot.OutcomeTick)
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.SelfHot().Apply(sim)
	}
	druid.FrenziedRegeneration = druid.RegisterSpell(config)
	druid.FrenziedRegenerationAura = druid.FrenziedRegeneration.SelfHot().Aura
	druid.FrenziedRegeneration.RelatedSelfBuff = druid.FrenziedRegenerationAura

	for _, form := range []*core.Aura{druid.BearFormAura, druid.DireBearFormAura} {
		if form != nil {
			form.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
				druid.FrenziedRegenerationAura.Deactivate(sim)
			})
		}
	}
}
