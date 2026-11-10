package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var rendRank = spellData.Rend.Highest()

// The Attack Power share per tick (0.02) is measured in game on Ranks 2, 4 and 5, with the attack power read at
// each tick. TODO: Ranks 1, 3, 6 and 7 carry the same share unmeasured. Verify it on them once logs at their level exist.
func (warrior *Warrior) registerRend() {
	tick := rendRank.PeriodicEffect()

	config := spelldata.SpellConfig(&warrior.Unit, rendRank,
		spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Flags(core.SpellFlagNoOnCastComplete))

	config.Dot = spelldata.DotConfig(rendRank, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damage := tick.Average() + tick.APCoeff()*dot.Spell.MeleeAttackPower(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, rendRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		} else {
			spell.IssueRefund(sim)
		}
	}

	warrior.Rend = warrior.RegisterSpell(config)
}
