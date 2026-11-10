package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The tooltip's "reduces base armor by 27% in Bear Form and 16% in Dire Bear Form". The row carries the
// penalty as effect 4, an A_DUMMY with no value. In Bear Form the game takes 27% of the item armor and the
// items' bonus armor, both after the form's multiplier, and of the 2 armor per Agility: 1206 armor drops to 962
// on a sheet whose item and Agility share is 904, and sheets whose bonus armor reads 60 and 100 lose 27% of 2.8
// times it on top. Buff armor stays whole and the level adds nothing. Thick Hide's armor stays whole, as no
// measured druid had it.
const (
	enrageBearArmorScale = 0.73
	// TODO: Ingame test needed -- Dire Bear takes its 16% from the same share as Bear Form, its bonus armor included.
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
		direBear := druid.InForm(DireBear)
		armorScale = core.TernaryFloat64(direBear, enrageDireBearArmorScale, enrageBearArmorScale)
		agiArmor = core.Ternary(direBear, direBearAgiArmor, bearAgiArmor)
		druid.ApplyDynamicEquipScaling(sim, stats.Armor, armorScale)
		druid.ApplyDynamicEquipScaling(sim, stats.BonusArmor, armorScale)
		druid.EnableBuildPhaseStatDep(sim, agiArmor)
	})
	druid.EnrageAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.RemoveDynamicEquipScaling(sim, stats.Armor, armorScale)
		druid.RemoveDynamicEquipScaling(sim, stats.BonusArmor, armorScale)
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
