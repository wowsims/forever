package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shieldBashRank = spellData.ShieldBash.Highest()

// Shield Bash's threat is 1.5 times its damage, times the stance: Rank 1's crits for 8 to 11 make 12 to
// 16.7, where a flat +2 would make 10 to 13.
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
