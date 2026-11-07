package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The tooltip's "reduces base armor by 27% in Bear Form and 16% in Dire Bear Form". The row carries the
// penalty as effect 4, an A_DUMMY with no value. In Bear Form the game takes 27% of the sum of the item armor
// after the form's multiplier and the 2 armor per Agility: 1206 armor drops to 962 on a sheet whose item and
// Agility share is 904. Buff armor stays whole. Only Bear Form was measured; Dire Bear takes its 16% from the same
// share. Bonus armor and Thick Hide's armor stay whole, as no measured druid had either.
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

	// The 2 is the armor per Agility every character gets (core's Agility to armor dependency).
	bearAgiArmor := druid.NewDynamicStatDependency(stats.Agility, stats.Armor, -2*(1-enrageBearArmorScale))
	direBearAgiArmor := druid.NewDynamicStatDependency(stats.Agility, stats.Armor, -2*(1-enrageDireBearArmorScale))
	var armorScale float64
	var agiArmor *stats.StatDependency
	druid.EnrageAura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		armorScale, agiArmor = enrageBearArmorScale, bearAgiArmor
		if druid.InForm(DireBear) {
			armorScale, agiArmor = enrageDireBearArmorScale, direBearAgiArmor
		}
		druid.ApplyDynamicEquipScaling(sim, stats.Armor, armorScale)
		druid.EnableBuildPhaseStatDep(sim, agiArmor)
	})
	druid.EnrageAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.RemoveDynamicEquipScaling(sim, stats.Armor, armorScale)
		druid.DisableBuildPhaseStatDep(sim, agiArmor)
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
