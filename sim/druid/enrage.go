package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The tooltip's "reduces base armor by 27% in Bear Form and 16% in Dire Bear Form". The row carries the
// penalty as effect 4, an A_DUMMY with no value. Bear Form's share of the armor (EV-0032, EV-0109) and
// Dire Bear Form's (EV-0034). Thick Hide's armor stays whole, as no measured druid had it.
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
	// The parse puts a SPELLMOD_EFFECT1 on Enrage (Enraged Idol's 1291060) on BonusBaseDamage, in the
	// row's tenths of rage. The combat log writes no line for these ticks, though each threatens as any
	// rage gained does.
	config.Hot.OnTick = func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
		druid.AddRage(sim, periodic.Tenths()+dot.Spell.BonusBaseDamage/10, druid.enrageMetrics())
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
