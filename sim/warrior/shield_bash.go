package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shieldBashRank = spellData.ShieldBash.Highest()

func (warrior *Warrior) registerShieldBash() {
	// TODO: Ingame test needed -- Rank 1's logs fit 1.5 times the damage as well as the store's flat bonus.
	config := spelldata.SpellConfig(&warrior.Unit, shieldBashRank, spelldata.Melee(core.ProcMaskMeleeOHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := shieldBashRank.DamageEffect().Roll(sim)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
