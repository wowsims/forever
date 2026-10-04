package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var disarmRank = spellData.Disarm.Highest()

func (warrior *Warrior) registerDisarm() {
	// TODO: core has no disarm effect, so the aura only tracks the debuff's uptime.
	auras := warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(spelldata.AuraConfig(disarmRank, spelldata.Label("Disarm-"+warrior.Label)))
	})

	config := spelldata.SpellConfig(&warrior.Unit, disarmRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)

		if result.Landed() {
			auras.Get(target).Activate(sim)
		} else {
			spell.IssueRefund(sim)
		}
	}

	config.RelatedAuraArrays = auras.ToMap()

	warrior.Disarm = warrior.RegisterSpell(config)
}
