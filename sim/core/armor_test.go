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
