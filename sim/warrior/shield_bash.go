package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shieldBashRank = spellData.ShieldBash.Highest()

// Shield Bash's threat is 1.5 times its damage, times the stance. Rank 1's crits for 8, 8, 9, 10 and 11
// read as 12.0, 12.2, 13.5, 15.0 and 16.7 threat; a factor of 1.5 predicts 12, 12, 13.5, 15 and 16.5,
// where a flat +2 would predict 10, 10, 11, 12 and 13.
const shieldBashThreatMultiplier = 1.5

func (warrior *Warrior) registerShieldBash() {
	config := spelldata.SpellConfig(&warrior.Unit, shieldBashRank, spelldata.Melee(core.ProcMaskMeleeOHSpecial))
	config.ThreatMultiplier = shieldBashThreatMultiplier

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := shieldBashRank.DamageEffect().Roll(sim)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
