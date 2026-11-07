package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shieldBashRank = spellData.ShieldBash.Highest()

// Shield Bash's threat is 1.5 times its damage, times the stance, from Rank 1's logs.
// TODO: Ingame test needed -- those logs fit a flat +2 as well; a hit of other damage tells them apart.
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
