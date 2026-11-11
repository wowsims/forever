package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerPrimalBiteSpell() {
	if !druid.Talents.PrimalBite {
		return
	}
	rank := spellData.PrimalBite.Highest()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ThreatMultiplier = 3
	// The tooltip's "generating a high amount of threat", which the row states no effect for: in game the
	// hit threatens for 3 times its damage with no flat part, measured on the first two ranks.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, druid.weaponSpecialDamage(sim, spell, target, rank), spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		druid.refundOnMiss(sim, spell, result)
	}
	druid.PrimalBite = druid.RegisterSpell(config)
}
