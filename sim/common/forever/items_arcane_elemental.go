package forever

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/stats"
)

// The Arcane Elemental 277729 that Violet Sorcerer's Vestments' 5-piece bonus summons: a guardian that only
// melees, a physical swing every 2.0 s. The game files carry no creature row for it. Its normal hit before
// armor is 0.62 of 2 s of the creature dps for its owner's level, ±10%, with no spell power share (EV-0169);
// the target's armor, the attack table and the crit multiplier apply as for any pet's melee.
type arcaneElemental struct {
	core.Pet
}

var arcaneElementalSwingAverage = 0.62 * core.CreatureAutoAttackDpsMaxLevel * 2.0

var arcaneElementalSwing = core.Weapon{
	BaseDamageMin: arcaneElementalSwingAverage * (1 - 0.1),
	BaseDamageMax: arcaneElementalSwingAverage * (1 + 0.1),
	SwingSpeed:    2.0,
	MaxRange:      core.MaxMeleeRange,
}

// One piece carries the elemental, since every piece is worn when the 5-piece bonus is. A summon while the
// last one is up keeps that one for the new 20 s.
func init() {
	core.NewItemPet(273044, newArcaneElemental) // Violet Sorcerer's Robes
}

func newArcaneElemental(owner *core.Character) core.PetAgent {
	elemental := &arcaneElemental{Pet: core.NewPet(core.PetConfig{
		Name:  "Arcane Elemental",
		Owner: owner,
		StatInheritance: func(_ stats.Stats) stats.Stats {
			return stats.Stats{}
		},
		IsGuardian:         true,
		StartsInMeleeRange: true,
	})}
	elemental.EnableAutoAttacks(elemental, core.AutoAttackOptions{
		MainHand:       arcaneElementalSwing,
		AutoSwingMelee: true,
	})
	return elemental
}

func (elemental *arcaneElemental) GetPet() *core.Pet                   { return &elemental.Pet }
func (elemental *arcaneElemental) Reset(_ *core.Simulation)            {}
func (elemental *arcaneElemental) OnEncounterStart(_ *core.Simulation) {}

// The elemental only swings: its rotation sleeps for as long as it can be up.
func (elemental *arcaneElemental) ExecuteCustomRotation(sim *core.Simulation) {
	elemental.WaitUntil(sim, sim.CurrentTime+arcaneElementalCall.Duration())
}
