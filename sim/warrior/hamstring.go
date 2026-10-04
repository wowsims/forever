package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var hamstringRank = spellData.Hamstring.Highest()
var hamstringBaseDamage = hamstringRank.DamageEffect().Average()

func (warrior *Warrior) registerHamstring() {
	// TODO: Manual review needed -- the game files state no threat coefficient; 1 until measured in game.
	config := spelldata.SpellConfig(&warrior.Unit, hamstringRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	// TODO: Ingame research needed if this adds flat threat
	config.FlatThreatBonus = 0

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, hamstringBaseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
