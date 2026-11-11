package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: Faerie Fire does not reset the swing timer.
func (druid *Druid) registerFaerieFireSpell() {
	rank := spellData.FaerieFire.Highest()
	druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataFaerieFireAura(target, true, 0)
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	// The timer must exist before the form passive adds its recovery time.
	config.Cast.CD = core.Cooldown{Timer: druid.NewTimer(), Duration: spellData.CatFormTriggered.Highest().EffectN(4).TimeValue()}
	// Cost, cooldown and GCD in cat or bear come from the form passive's spell mods. A cast on a target
	// that has the debuff refreshes it; when to refresh is the rotation's.
	// The row's E_THREAT effect states 0. In game a landed cast in Bear or Dire Bear Form threatens for
	// twice the caster's level, which the form's threat multiplier raises, and in Cat Form or caster form
	// for none; measured at caster levels 38 to 50 on the lower ranks.
	// TODO: Ingame research needed -- the level 60 rank's threat, which this takes to follow the same rule.
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
