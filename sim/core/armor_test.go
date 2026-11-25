package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

func TestArmorDamageReductionCap(t *testing.T) {
	// Boss attacker: armorConstant = 400 + 85*63 = 5755, so 75% reduction is
	// reached at 3*5755 = 17265 armor.
	attacker := Unit{Type: EnemyUnit, Level: 63}
	tolerance := 0.0001

	modifierForArmor := func(armor float64) float64 {
		defender := Unit{
			Type:         PlayerUnit,
			Level:        60,
			initialStats: stats.Stats{stats.Armor: armor},
			PseudoStats:  stats.NewPseudoStats(),
		}
		defender.stats = defender.initialStats
		return NewAttackTable(&attacker, &defender).GetArmorDamageModifier(nil)
	}

	if modifier := modifierForArmor(11510); !WithinToleranceFloat64(1.0/3.0, modifier, tolerance) {
		t.Fatalf("Expected %f damage taken below the cap, got %f", 1.0/3.0, modifier)
	}
	if modifier := modifierForArmor(17265); !WithinToleranceFloat64(0.25, modifier, tolerance) {
		t.Fatalf("Expected %f damage taken at the cap, got %f", 0.25, modifier)
	}
	if modifier := modifierForArmor(50000); !WithinToleranceFloat64(0.25, modifier, tolerance) {
		t.Fatalf("Expected %f damage taken above the cap, got %f", 0.25, modifier)
	}
}

func TestArmorReductionFromLevel60Attacker(t *testing.T) {
	attacker := Unit{Type: PlayerUnit, Level: 60}
	defender := Unit{Type: EnemyUnit, Level: 63, PseudoStats: stats.NewPseudoStats()}
	defender.stats[stats.Armor] = 3731

	reduction := 1 - NewAttackTable(&attacker, &defender).GetArmorDamageModifier(nil)
	if !WithinToleranceFloat64(0.4042, reduction, 0.0001) {
		t.Fatalf("3731 armor against a level 60 attacker reduces damage by %.4f, want 0.4042", reduction)
	}
}

// Periodic physical damage ignores armor, meets all of it under SpellFlagApplyArmorReduction, and meets a
// tenth of what armor penetration leaves under SpellFlagTenthResistance, below the 75% cap that the whole
// armor reaches; a direct hit of a spell flagged for a tenth meets the whole armor.
func TestPeriodicPhysicalDamageMeetsTheArmorItsFlagsName(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60}
	attacker.stats[stats.ArmorPenetration] = 1000
	defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender.stats[stats.Armor] = 21000
	attackTable := NewAttackTable(attacker, defender)
	armorConstant := 400 + 85*float64(attacker.Level)
	tenth := (defender.Armor() - attacker.stats[stats.ArmorPenetration]) / 10

	for _, tc := range []struct {
		name       string
		flags      SpellFlag
		isPeriodic bool
		want       float64
	}{
		{"a tick", 0, true, 1},
		{"a tick flagged for the whole armor", SpellFlagApplyArmorReduction, true, 0.25},
		{"a tick flagged for a tenth", SpellFlagTenthResistance, true, 1 - tenth/(tenth+armorConstant)},
		{"a direct hit flagged for a tenth", SpellFlagTenthResistance, false, 0.25},
	} {
		spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolPhysical, SchoolIndex: stats.SchoolIndexPhysical, Flags: tc.flags}
		if got, _ := spell.ResistanceMultiplier(nil, tc.isPeriodic, attackTable); !WithinToleranceFloat64(tc.want, got, 1e-9) {
			t.Errorf("%s keeps %v of its damage, want %v", tc.name, got, tc.want)
		}
	}
}
