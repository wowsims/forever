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

// Two elementals can be up at once, since one lasts 20 s and the summon rests 15 s (EV-0169). Two pieces carry
// one each, since every piece is worn when the 5-piece bonus is.
func init() {
	core.NewItemPet(273044, newArcaneElemental("Arcane Elemental 1")) // Violet Sorcerer's Robes
	core.NewItemPet(273051, newArcaneElemental("Arcane Elemental 2")) // Violet Sorcerer's Mantle
}

func newArcaneElemental(name string) func(owner *core.Character) core.PetAgent {
	return func(owner *core.Character) core.PetAgent {
		elemental := &arcaneElemental{Pet: core.NewPet(core.PetConfig{
			Name:  name,
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
}

func (elemental *arcaneElemental) GetPet() *core.Pet                   { return &elemental.Pet }
func (elemental *arcaneElemental) Reset(_ *core.Simulation)            {}
func (elemental *arcaneElemental) OnEncounterStart(_ *core.Simulation) {}

// The elemental only swings: its rotation sleeps for as long as it can be up.
func (elemental *arcaneElemental) ExecuteCustomRotation(sim *core.Simulation) {
	elemental.WaitUntil(sim, sim.CurrentTime+arcaneElementalCall.Duration())
}
