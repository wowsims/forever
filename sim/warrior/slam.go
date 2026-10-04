package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var slamRank = spellData.Slam.Highest()

var improvedSlamRank = slamRank.OverriddenBy(spellData.ImprovedSlam.Highest())

func (warrior *Warrior) registerSlam() {
	rank := core.Ternary(warrior.Talents.ImprovedSlam > 0, improvedSlamRank, slamRank)
	stopsSwing := !rank.HasAttr(dbcenums.ATTR_INDEX_EX_2, dbcenums.ATTR_EX_2_DO_NOT_RESET_COMBAT_TIMERS)

	config := spelldata.SpellConfig(&warrior.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	// The APL names Slam by the spell the talent replaces.
	config.ActionID = core.ActionID{SpellID: slamRank.ID}

	config.Cast.ModifyCast = func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
		if cast.CastTime > 0 && stopsSwing {
			warrior.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+cast.CastTime)
		}
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := rank.WeaponStrike(sim, spell, spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
