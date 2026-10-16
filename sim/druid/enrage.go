package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The tooltip's "reduces base armor by 27% in Bear Form and 16% in Dire Bear Form". The row carries the
// penalty as effect 4, an A_DUMMY with no value. Base armor is the equipment share.
const (
	enrageBearArmorScale     = 0.73
	enrageDireBearArmorScale = 0.84
)

func (druid *Druid) registerEnrageSpell() {
	rank := spellData.Enrage.Highest()
	periodic := rank.PeriodicEffect()
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.Hot = spelldata.DotConfig(rank, periodic)
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
		druid.AddRage(sim, periodic.Tenths(), druid.enrageMetrics())
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		if instant := rank.EnergizeEffect(); instant != spelldata.NilEffect {
			druid.AddRage(sim, instant.Tenths()*(1+druid.WolfsheadEnrageBonus), druid.enrageMetrics())
		}
		spell.SelfHot().Apply(sim)
	}
	druid.Enrage = druid.RegisterSpell(config)
	druid.EnrageAura = druid.Enrage.SelfHot().Aura
	druid.Enrage.RelatedSelfBuff = druid.EnrageAura

	var armorScale float64
	druid.EnrageAura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		armorScale = core.TernaryFloat64(druid.InForm(DireBear), enrageDireBearArmorScale, enrageBearArmorScale)
		druid.ApplyDynamicEquipScaling(sim, stats.Armor, armorScale)
	})
	druid.EnrageAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.RemoveDynamicEquipScaling(sim, stats.Armor, armorScale)
	})
	for _, form := range []*core.Aura{druid.BearFormAura, druid.DireBearFormAura} {
		if form != nil {
			form.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
				druid.EnrageAura.Deactivate(sim)
			})
		}
	}
}

func (druid *Druid) enrageMetrics() *core.ResourceMetrics {
	if druid.enrageMetric == nil {
		druid.enrageMetric = druid.NewRageMetrics(core.ActionID{SpellID: spellData.Enrage.Highest().ID})
	}
	return druid.enrageMetric
}
