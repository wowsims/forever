package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerDemoralizingRoarSpell() {
	rank := spellData.DemoralizingRoar.Highest()
	druid.DemoralizingRoarAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataDemoralizingRoarAura(target, true, 0)
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskEmpty))
	// Effect 1 hits every enemy around the caster, and the row states no target cap.
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, target := range sim.Encounter.ActiveTargetUnits {
			// The row's defense type is magic, so a target cannot dodge it.
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				druid.DemoralizingRoarAuras.Get(target).Activate(sim)
			}
		}
	}
	config.RelatedAuraArrays = druid.DemoralizingRoarAuras.ToMap()
	druid.DemoralizingRoar = druid.RegisterSpell(config)
}
