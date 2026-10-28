package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerRavageSpell() {
	rank := spellData.Ravage.Highest()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	// The row's stealth requirement (ATTR_REQ_STEALTH) is not read by CastRequirement, so Prowl or
	// Shadowmeld is checked here; "behind the target" is the tooltip's.
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.stealthed() && druid.BehindTarget()
	}
	// No Active Defense (Attributes 0x200000): it misses or lands, never dodged, parried or blocked.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, druid.weaponSpecialDamage(sim, spell, target, rank), spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		if result.Landed() {
			druid.awardComboPoints(sim, spell, rank)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Ravage = druid.RegisterSpell(config)
}
