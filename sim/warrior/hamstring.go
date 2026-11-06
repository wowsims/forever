package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var hamstringRank = spellData.Hamstring.Highest()
var hamstringBaseDamage = hamstringRank.DamageEffect().Average()

// Hamstring's threat is 1.25 times its damage, times the stance: Rank 1's 4-damage hits make 5.
// TODO: Ingame test needed -- a hit of other damage tells 1.25 times from a flat +1.
const hamstringThreatMultiplier = 1.25

func (warrior *Warrior) registerHamstring() {
	config := spelldata.SpellConfig(&warrior.Unit, hamstringRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ThreatMultiplier = hamstringThreatMultiplier

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, hamstringBaseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
