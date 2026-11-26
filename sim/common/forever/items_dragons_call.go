package forever

import (
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The Emerald Dragon Whelp 8776 that Dragon's Call's 13049 summons is level 57 whatever its owner's level. Health is
// 1.3 times the creature health for that level. Armor is the creature armor at ExpectedStatMod 1's share in the game
// files, 0.51. It never crits (EV-0021).
//
// The 2.0 s swing is 2 s of the level 57 creature dps at 0.73, ExpectedStatMod 1's share, 14% either side (EV-0021).
// That row is inferred from the whelp's armor matching its 0.51. A smaller share plus the whelp's in-game attack power
// fits as well (0.5 gives 82), so the log does not tell the two apart. The attack power the whelp shows in game is
// left out; the swing carries all its damage.
//
// It glances 2% of its landed swings for each point of defense the target has over 250, so against level 60 and up
// every landed swing glances, and a glancing blow deals the share of a hit the logs read for the target's level, down
// to 1% to 18% against levels 62 and 63 (EV-0021).
//
// The first Acid Spit 9591 starts 0 to 1.8 s after the summon, each next wait is 0.2 to 3.5 s, and a spit's
// hardcast holds the swing, as in the log (EV-0021).
var emeraldDragonWhelp = shared.GuardianCreature{
	Name:   "Emerald Dragon Whelp",
	Level:  57,
	Health: 1.3,
	Armor:  0.51,
	Swing: shared.GuardianSwing{Share: 0.73, Speed: 2.0, Variance: 0.14, Glance: &shared.GuardianGlance{
		Skill: 250,
		Damage: map[int32][2]float64{
			56: {0.60, 0.98},
			57: {0.57, 0.92},
			58: {0.34, 0.75},
			59: {0.08, 0.59},
			60: {0.02, 0.45},
			61: {0.01, 0.29},
			62: {0.01, 0.18},
		},
	}},
	Cast: &shared.GuardianCast{
		FirstMax:   1800 * time.Millisecond,
		WaitMin:    200 * time.Millisecond,
		WaitMax:    3500 * time.Millisecond,
		FirstLabel: "Whelp First Spit",
		WaitLabel:  "Whelp Spit Idle",
		Config:     acidSpit,
	},
}

// Acid Spit 9591 deals its base points as a percent of the creature dps for the caster's level, which
// is what the row's creature level scaling attribute (0x80000 on Attributes 0) means: 249.7 at the whelp's
// level 57 (EV-0021). Its 5340 mana at level 57 pays for more spits than a summon lasts, so the cast takes no
// cost.
func acidSpit(whelp *shared.Guardian) core.SpellConfig {
	row := spelldata.MustFind(9591)
	baseDamage := row.DamageEffect().Percent() * whelp.CreatureStats().AutoAttackDps
	config := spelldata.SpellConfig(&whelp.Unit, row, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ManaCost = core.ManaCostOptions{}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
	}
	return config
}
