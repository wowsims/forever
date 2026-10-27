package core_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Each walk lands in its own action row, tag 0 for one the unit starts by hand and tag 1 for one a spell
// causes, and counts one cast, for the owner and for its pet.
func TestWalksSplitByHandFromSpellCaused(t *testing.T) {
	sim, owner := setupPetOwnerSim(&proto.RaidBuffs{}, &proto.PartyBuffs{})
	pet := &owner.pet.Unit

	owner.MoveDurationByHand(time.Second, sim)
	owner.StopMoving(sim)
	owner.MoveTo(owner.DistanceFromTarget+5, sim)
	owner.StopMoving(sim)
	pet.MoveDurationByHand(time.Second, sim)
	pet.StopMoving(sim)
	pet.MoveTo(pet.DistanceFromTarget+5, sim)
	pet.StopMoving(sim)

	sim.Cleanup()

	action := func(unit *core.Unit, id core.ActionID) *proto.ActionMetrics {
		for _, action := range unit.Metrics.ToProto().Actions {
			if core.ProtoToActionID(action.Id).SameAction(id) {
				return action
			}
		}
		t.Fatalf("%s has no metrics for %s", unit.Label, id)
		return nil
	}
	move := core.ActionID{OtherID: proto.OtherAction_OtherActionMove}
	for _, tc := range []struct {
		name string
		unit *core.Unit
		tag  int32
	}{
		{"the owner's walk by hand", &owner.Unit, 0},
		{"the owner's walk a spell causes", &owner.Unit, 1},
		{"the pet's walk by hand", pet, 0},
		{"the pet's walk its own logic causes", pet, 1},
	} {
		walks := action(tc.unit, move.WithTag(tc.tag))
		casts := int32(0)
		for _, target := range walks.Targets {
			casts += target.Casts
		}
		if casts != 1 {
			t.Errorf("%s has %d casts, want 1", tc.name, casts)
		}
	}
}
