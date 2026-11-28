package forever

import (
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The Emerald Dragon Whelp 8776 that Dragon's Call's 13049 summons is level 57 whatever its owner's level.
// Armor is the creature armor at ExpectedStatMod 1's share in the game files, 0.51. It never crits (EV-0021).
//
// Swing damage in EV-0021. The swing's 0.73 is ExpectedStatMod 1's share, a row inferred from the
// whelp's armor matching that row's 0.51. A smaller share plus the whelp's in-game attack power fits
// as well, so the log does not tell the two apart.
//
// Glancing blows in EV-0021.
//
// Spit timing in EV-0021. A spit's hardcast holds the swing.
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

// The base points are a percent of the creature dps for the caster's level (EV-0021), which is what the
// spell's creature level scaling attribute (0x80000 on Attributes 0) means.
// The whelp's mana pays for more spits than a summon lasts, so the cast takes no cost.
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
