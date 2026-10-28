package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var feintRank = spellData.Feint.Rank(4)

// This exists purely to model the DPS loss of having to occasionally feint
func (rogue *Rogue) registerFeintSpell() {
	config := spelldata.SpellConfig(&rogue.Unit, feintRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	// The row lacks No Active Defense, so the target can dodge it and, from the front, parry it.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		result := spell.CalcAndDealDamage(sim, target, 0, spell.OutcomeMeleeSpecialHit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	rogue.RegisterSpell(config)
}
