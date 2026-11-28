package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerFaerieFireSpell() {
	rank := spellData.FaerieFire.Highest()
	druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataFaerieFireAura(target, true, 0)
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	// The timer must exist before the form passive adds its recovery time.
	config.Cast.CD = core.Cooldown{Timer: druid.NewTimer(), Duration: spellData.CatFormTriggered.Highest().EffectN(4).TimeValue()}
	// Cost, cooldown and GCD in cat or bear come from the form passive's spell mods.
	// The row's E_THREAT effect states 0. In game a landed cast in Bear or Dire Bear Form threatens
	// (EV-0034, EV-0037).
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.FlatThreatBonus = core.TernaryFloat64(druid.InForm(Bear), 2*float64(druid.level()), 0)
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			druid.FaerieFireAuras.Get(target).Activate(sim)
		}
	}
	config.RelatedAuraArrays = druid.FaerieFireAuras.ToMap()
	druid.FaerieFire = druid.RegisterSpell(config)
	druid.FaerieFire.CD.Duration = rank.Cooldown()
}
