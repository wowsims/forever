package core_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// An action is pressed when its unit presses it: a player's APL spell, a SpellFlagPressed spell or the hit
// a queue spell of the same action presses, a player's walk by hand, and a pet's ability. Auto attacks,
// procs, a walk a spell causes and a pet's walks are not.
func TestActionMetricsPressedIsWhatTheUnitPresses(t *testing.T) {
	sim, owner := setupPetOwnerSim(&proto.RaidBuffs{}, &proto.PartyBuffs{})
	pet := &owner.pet.Unit

	register := func(unit *core.Unit, id core.ActionID, flags core.SpellFlag) core.ActionID {
		return unit.RegisterSpell(core.SpellConfig{ActionID: id, Flags: flags}).ActionID
	}
	aplSpell := register(&owner.Unit, core.ActionID{SpellID: 1}, core.SpellFlagAPL)
	queuedHit := register(&owner.Unit, core.ActionID{SpellID: 2}, core.SpellFlagMeleeMetrics)
	register(&owner.Unit, queuedHit.WithTag(1), core.SpellFlagAPL|core.SpellFlagNoMetrics)
	ownerProc := register(&owner.Unit, core.ActionID{SpellID: 3}, core.SpellFlagProc)
	petAbility := register(pet, core.ActionID{SpellID: 4}, core.SpellFlagNone)
	petProc := register(pet, core.ActionID{SpellID: 5}, core.SpellFlagProc)
	pressedThrough := register(&owner.Unit, core.ActionID{SpellID: 6}, core.SpellFlagPressed)
	petSwing := *pet.AutoAttacks.MHConfig()
	petSwing.ActionID = petSwing.ActionID.WithTag(5)
	petSwingCopy := pet.RegisterSpell(petSwing).ActionID
	sharedAction := core.ActionID{SpellID: 7}
	register(&owner.Unit, sharedAction, core.SpellFlagNone)
	register(&owner.Unit, sharedAction, core.SpellFlagPressed)

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
	isPressed := func(unit *core.Unit, id core.ActionID) bool {
		return action(unit, id).IsPressed
	}
	for _, tc := range []struct {
		name string
		unit *core.Unit
		id   core.ActionID
		want bool
	}{
		{"an APL spell", &owner.Unit, aplSpell, true},
		{"a SpellFlagPressed spell", &owner.Unit, pressedThrough, true},
		{"two spells of one action, the unpressed first", &owner.Unit, sharedAction, true},
		{"a hit pressed through its tag 1 queue spell", &owner.Unit, queuedHit, true},
		{"the owner's proc", &owner.Unit, ownerProc, false},
		{"the owner's melee", &owner.Unit, owner.AutoAttacks.MHAuto().ActionID, false},
		{"the pet's ability", pet, petAbility, true},
		{"the pet's proc", pet, petProc, false},
		{"a tagged copy of the pet's melee", pet, petSwingCopy, false},
		{"the pet's melee", pet, pet.AutoAttacks.MHAuto().ActionID, false},
	} {
		if got := isPressed(tc.unit, tc.id); got != tc.want {
			t.Errorf("%s is pressed %t, want %t", tc.name, got, tc.want)
		}
	}

	// Each walk lands in its own row: tag 0 for one by hand, tag 1 for one a spell causes.
	move := core.ActionID{OtherID: proto.OtherAction_OtherActionMove}
	for _, tc := range []struct {
		name string
		unit *core.Unit
		tag  int32
		want bool
	}{
		{"the owner's walk by hand", &owner.Unit, 0, true},
		{"the owner's walk a spell causes", &owner.Unit, 1, false},
		{"the pet's walk by hand", pet, 0, false},
		{"the pet's walk its own logic causes", pet, 1, false},
	} {
		walks := action(tc.unit, move.WithTag(tc.tag))
		casts := int32(0)
		for _, target := range walks.Targets {
			casts += target.Casts
		}
		if casts != 1 || walks.IsPressed != tc.want {
			t.Errorf("%s has %d casts and is pressed %t, want 1 cast and pressed %t", tc.name, casts, walks.IsPressed, tc.want)
		}
	}
}
