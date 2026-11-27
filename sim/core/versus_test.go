package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// Every creature family but the two that span several mob types counts as a mob type on a target with
// none of its own, so a family added to the enum needs its mob type.
func TestATargetOfAFamilyCountsAsTheFamilysMobType(t *testing.T) {
	spanning := map[proto.CreatureFamily]bool{
		proto.CreatureFamily_CreatureFamilyUnknown:    true,
		proto.CreatureFamily_CreatureFamilyAquatic:    true,
		proto.CreatureFamily_CreatureFamilySerpentine: true,
	}
	for value, name := range proto.CreatureFamily_name {
		family := proto.CreatureFamily(value)
		got := NewTarget(&proto.Target{CreatureFamily: family}, 0).MobType
		if spanning[family] != (got == proto.MobType_MobTypeUnknown) {
			t.Errorf("a %s target with no mob type counts as %v", name, got)
		}
	}
	if got := NewTarget(&proto.Target{MobType: proto.MobType_MobTypeUndead, CreatureFamily: proto.CreatureFamily_CreatureFamilyTroll}, 0).MobType; got != proto.MobType_MobTypeUndead {
		t.Errorf("an Undead Troll counts as %v, want Undead alone", got)
	}
}

// A Fire Elemental with no mob type takes every bonus against Elementals: a target match and the
// Skyborne's Elemental Insight on the attack table.
func TestAFireElementalTakesTheBonusesAgainstElementals(t *testing.T) {
	sim, fw := setupRacialSimAgainst(proto.Race_RaceHighOrderSkyborne,
		&proto.Target{Name: "target", Level: 63, CreatureFamily: proto.CreatureFamily_CreatureFamilyFireElemental})
	target := sim.Encounter.ActiveTargetUnits[0]

	if !(VersusTargets{MobTypes: []proto.MobType{proto.MobType_MobTypeElemental}}).Matches(target) {
		t.Error("a Fire Elemental with no mob type is no Elemental to a versus check")
	}
	if got := fw.AttackTables[target.UnitIndex].DamageDealtMultiplier; !WithinToleranceFloat64(1.05, got, 0.0001) {
		t.Errorf("Elemental Insight deals x%0.4f to a Fire Elemental with no mob type, want x1.05", got)
	}
}
