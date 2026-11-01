package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 4 is the last rank the trainer teaches; rank 5 comes from the Handbook of Feint V book.
var feintTrainerRank = spellData.Feint.ByID(11303)
var feintBookRank = spellData.Feint.ByID(25302)

// This exists purely to model the DPS loss of having to occasionally feint
func (rogue *Rogue) registerFeintSpell() {
	rogue.registerFeintRank(feintTrainerRank)
	rogue.registerFeintRank(feintBookRank)
}

func (rogue *Rogue) registerFeintRank(rank *spelldata.Spell) {
	config := spelldata.SpellConfig(&rogue.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

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
