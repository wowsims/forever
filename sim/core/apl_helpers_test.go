package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestRandomFriendlyReferencesShareOneTargetPerActionPick(t *testing.T) {
	request := fakeRageSimRequest()
	request.Raid.TargetDummies = 4
	sim := setupFakeRageSimFrom(request)
	unit := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
	rotation := unit.newAPLRotation(&proto.APLRotation{Type: proto.APLRotation_TypeAPL})
	ref := &proto.UnitReference{Type: proto.UnitReference_RandomFriendlyTarget}
	condition := rotation.GetSourceUnit(ref)
	cast := rotation.GetTargetUnit(ref)
	seen := make(map[*Unit]bool)
	for range 40 {
		rotation.getNextAction(sim)
		target := cast.Get()
		if target == nil || target == unit || target.IsOpponent(unit) {
			t.Fatal("random friendly target is not an initialized target dummy")
		}
		if condition.Get() != target || cast.Get() != target {
			t.Fatal("condition and cast read different targets during one action pick")
		}
		seen[target] = true
	}
	if len(seen) != 4 {
		t.Errorf("random target visited %d dummies, want all four", len(seen))
	}

	plain := unit.newAPLRotation(&proto.APLRotation{Type: proto.APLRotation_TypeAPL})
	plain.getNextAction(sim)
	if plain.randomFriendlyTarget != nil {
		t.Fatal("a rotation without random references created a random target")
	}
}
