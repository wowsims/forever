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

// What a hawk is, read off combat logs: 5% of the creature health for its level and no armor or attack
// power of its own; 80% of the owner's attack power and 30% of the owner's armor on top (EV-0074, EV-0073).
// The row that scales a hawk, Hunter Guardian Hawk Pet Scaling 01 (1293586), is not in the store, so the
// swing below is fitted to those logs (EV-0073).
var hawkBaseStats = stats.Stats{
	stats.Health: 0.05 * core.CreatureHealthMaxLevel,
}

// The attack power a hawk takes its 80% of is the owner's ranged attack power and what the main-hand
// weapon adds to it in game, 14 a point of its dps; the ranged weapon and the off hand add nothing,
// and with no main-hand weapon it adds 7 (EV-0073).
const hawkOwnerAPPerMainHandDPS = 14.0
const hawkOwnerUnarmedWeaponAP = 7.0

// A swing before armor is 0.153 of the creature dps over the 2.5 s at level 60, spanning 0.22 either side,
// and 60 attack power a point of dps (EV-0073). Both are fitted, as the game files state
// neither, and were read with Unleashed Fury 5/5 and Focused Fire 2/2, which reach the hawks, so the
// swing here is before them.
var hawkLogDamageMultiplier = spellData.UnleashedFury.MultiplierAt(5) * spellData.FocusedFire.MultiplierAt(2)
var hawkSwingAverage = 0.153 * core.CreatureStatsByLevel[core.CharacterLevel].AutoAttackDps * 2.5 / hawkLogDamageMultiplier

var hawkSwing = core.Weapon{
	BaseDamageMin:     hawkSwingAverage * (1 - 0.22),
	BaseDamageMax:     hawkSwingAverage * (1 + 0.22),
	SwingSpeed:        2.5,
	AttackPowerPerDPS: 60 * hawkLogDamageMultiplier,
	MaxRange:          core.MaxMeleeRange,
}

// Effect 3 is the number of hawks up at once. A hawk's swing follows its owner's ranged speed while
// it is up, a quiver's and Quick Shots' included; Berserking's attack speed does not reach it in
// game. Its crit is half the tamed pet's, the higher of melee and ranged (inheritedCrit), with none
// from Agility: with Ferocity 5/5 it reads 10 plus half the owner's crit, and with no Ferocity half the
// tamed pet's (EV-0074). Its melee hit and expertise are the tamed pet's.
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
					stats.AttackPower:         (ownerStats[stats.RangedAttackPower] + hunter.hawkWeaponAP()) * 0.8,
					stats.PhysicalCritPercent: hunter.inheritedCrit(ownerStats) * 0.5,
					stats.PhysicalHitPercent:  inheritedHit(ownerStats),
					stats.ExpertisePercent:    ownerStats[stats.ExpertisePercent],
				}
			},
			IsGuardian:                       true,
			IsDynamic:                        true,
			HasDynamicRangedSpeedInheritance: true,
			StartsInMeleeRange:               true,
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

func (hunter *Hunter) hawkWeaponAP() float64 {
	if hunter.GetMHWeapon() == nil {
		return hawkOwnerUnarmedWeaponAP
	}
	mainHand := hunter.WeaponFromMainHand()
	return hawkOwnerAPPerMainHandDPS * mainHand.DPS()
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

// The dive's damage, then a hawk for 18 s. The dive flies the row's minimum of 1 s at any distance
// in range, and the hawk appears as it lands and swings at once, on a fresh swing timer when it
// replaces another (EV-0127). The hawks take turns, so one past the count replaces the one with the
// least time left.
func (hunter *Hunter) registerSummonHawk() {
	if !hunter.Talents.SummonHawk {
		return
	}

	config := spelldata.SpellConfig(&hunter.Unit, summonHawkRank, spelldata.Melee(core.ProcMaskRangedSpecial))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := summonHawkBaseDamage + summonHawkAPShare*spell.RangedAttackPower(target)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedCritOnly)
		spell.DealCastResult(sim, result)
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)

			hawk := hunter.Hawks[hunter.nextHawk]
			hunter.nextHawk = (hunter.nextHawk + 1) % len(hunter.Hawks)
			if hawk.IsEnabled() {
				hawk.Disable(sim)
			}
			hawk.EnableWithTimeout(sim, hawk, hawkSummon.Duration())
			hawk.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime-hawk.AutoAttacks.MainhandSwingSpeed())
		})
	}

	hunter.RegisterSpell(config)
}
