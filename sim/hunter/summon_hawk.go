package hunter

import (
	"strconv"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var summonHawkRank = spellData.SummonHawk.Highest()
var summonHawkBaseDamage = summonHawkRank.DamageEffect().Average()

// The creature the dive leaves behind, up for 18 s.
var hawkSummon = spellData.SummonHawkTriggered.ByID(1293248)

// The tooltip's ${$s1+($rap*($s4/100))}: effect 4 is the share of ranged attack power.
var summonHawkAPShare = summonHawkRank.EffectN(4).Percent()

// A hawk the dive leaves behind: a guardian with a pet's base stats and swing damage, summoned at
// the enemy by the cast.
type HawkPet struct {
	core.Pet
}

func (hawk *HawkPet) GetPet() *core.Pet                   { return &hawk.Pet }
func (hawk *HawkPet) Reset(_ *core.Simulation)            {}
func (hawk *HawkPet) OnEncounterStart(_ *core.Simulation) {}

// A hawk only swings: its rotation sleeps for as long as a hawk can be up.
func (hawk *HawkPet) ExecuteCustomRotation(sim *core.Simulation) {
	hawk.WaitUntil(sim, sim.CurrentTime+hawkSummon.Duration())
}

// What a hawk is, read off level 30 combat logs of 2862 swings at 407 to 1023 owner ranged attack power:
// 5% of the creature health for its level and no armor or attack power of its own; 80% of the owner's
// ranged attack power and 30% of the owner's armor on top. The row that scales a hawk, Hunter Guardian
// Hawk Pet Scaling 01 (1293586), is not in the store, so the swing below is fitted to those logs.
var hawkBaseStats = stats.Stats{
	stats.Health: 0.05 * core.CreatureHealthMaxLevel,
}

// A swing before armor read 0.0815 of the creature dps over the 2.5 s, spanning a third either side, plus
// 0.0179 a point of the owner's ranged attack power: 112 attack power a point of dps at 80% of it. Both
// are fitted (3.3-6.7 plus 0.0179 × ranged attack power at level 30), not stated in the game files, and
// were read with Unleashed Fury 5/5, which reaches the hawks, so the swing here is before it.
var hawkLogDamageMultiplier = spellData.UnleashedFury.MultiplierAt(5)
var hawkSwingAverage = 0.0815 * core.CreatureAutoAttackDpsMaxLevel * 2.5 / hawkLogDamageMultiplier

var hawkSwing = core.Weapon{
	BaseDamageMin:     hawkSwingAverage * (1 - 1.0/3),
	BaseDamageMax:     hawkSwingAverage * (1 + 1.0/3),
	SwingSpeed:        2.5,
	AttackPowerPerDPS: 112 * hawkLogDamageMultiplier,
	MaxRange:          core.MaxMeleeRange,
}

// Effect 3 is the number of hawks up at once. A hawk's swing follows its owner's ranged speed while
// it is up, a quiver's and Quick Shots' included; Berserking's attack speed does not reach it in
// game. Its crit is half the tamed pet's, the higher of melee and ranged (inheritedCrit), with none
// from Agility: 4340 swings with Ferocity 5/5 read 10 plus half the owner's crit against targets 2
// levels below to 3 above it, and 0-Ferocity logs half the tamed pet's. Its melee hit and expertise are
// the tamed pet's: 1% hit took a hawk from 1 miss in 32 swings to none in 515 against a target 8
// levels below it.
func (hunter *Hunter) newHawks() []*HawkPet {
	if !hunter.Talents.SummonHawk {
		return nil
	}

	hawks := make([]*HawkPet, int(summonHawkRank.EffectN(3).BaseValue()))
	for i := range hawks {
		hawk := &HawkPet{Pet: core.NewPet(core.PetConfig{
			Name:      "Hawk " + strconv.Itoa(i+1),
			Owner:     &hunter.Character,
			BaseStats: hawkBaseStats,
			StatInheritance: func(ownerStats stats.Stats) stats.Stats {
				return stats.Stats{
					stats.Armor:               ownerStats[stats.Armor] * 0.3,
					stats.AttackPower:         ownerStats[stats.RangedAttackPower] * 0.8,
					stats.PhysicalCritPercent: hunter.inheritedCrit(ownerStats) * 0.5,
					stats.PhysicalHitPercent:  inheritedHit(ownerStats),
					stats.ExpertisePercent:    ownerStats[stats.ExpertisePercent],
				}
			},
			IsGuardian:                       true,
			IsDynamic:                        true,
			HasDynamicRangedSpeedInheritance: true,
		})}
		hawk.EnableAutoAttacks(hawk, core.AutoAttackOptions{
			MainHand:       hawkSwing,
			AutoSwingMelee: true,
		})
		// The damage done is the owner's as the hawk is summoned, on top of the hawk's own.
		inherited := 1.0
		hawk.OnPetEnable = func(sim *core.Simulation) {
			inherited = hunter.PseudoStats.DamageDealtMultiplier
			hawk.PseudoStats.DamageDealtMultiplier *= inherited
		}
		hawk.OnPetDisable = func(sim *core.Simulation) {
			hawk.PseudoStats.DamageDealtMultiplier /= inherited
		}
		hunter.AddPet(hawk)
		hawks[i] = hawk
	}
	return hawks
}

// The tamed pet and the hawks: what the talents that name pets and hawks reach.
func (hunter *Hunter) pets() []*core.Pet {
	var pets []*core.Pet
	if hunter.Pet != nil {
		pets = append(pets, &hunter.Pet.Pet)
	}
	for _, hawk := range hunter.Hawks {
		pets = append(pets, &hawk.Pet)
	}
	return pets
}

// The dive's damage, then a hawk for 18 s. The hawks take turns, so one past the count replaces
// the one with the least time left.
func (hunter *Hunter) registerSummonHawk() {
	if !hunter.Talents.SummonHawk {
		return
	}

	config := spelldata.SpellConfig(&hunter.Unit, summonHawkRank, spelldata.Melee(core.ProcMaskRangedSpecial))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := summonHawkBaseDamage + summonHawkAPShare*spell.RangedAttackPower(target)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedCritOnly)
		spell.DealDamageAfterTravel(sim, result)

		hawk := hunter.Hawks[hunter.nextHawk]
		hunter.nextHawk = (hunter.nextHawk + 1) % len(hunter.Hawks)
		if hawk.IsEnabled() {
			hawk.Disable(sim)
		}
		hawk.EnableWithTimeout(sim, hawk, hawkSummon.Duration())
	}

	hunter.RegisterSpell(config)
}
