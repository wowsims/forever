package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var kickRank = spellData.Kick.Highest()

// This exists purely to model the DPS loss of having to occasionally kick
func (rogue *Rogue) registerKickSpell() {
	kickBaseDamage := kickRank.DamageEffect().Average()
	config := spelldata.SpellConfig(&rogue.Unit, kickRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.VersusStats = []stats.Stat{stats.PhysicalDamage}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := kickBaseDamage
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	rogue.RegisterSpell(config)
}
