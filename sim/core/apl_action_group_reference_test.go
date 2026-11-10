package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// A group referenced from a prepull action, a priority list entry, a sequence in the priority list and a sequence
// in another group resolves at every reference, each to a copy of its own.
func TestGroupReferencedFromPrepullAndSequencesResolvesEverywhere(t *testing.T) {
	const spellID = 910107
	refTo := func(name string) *proto.APLAction {
		return &proto.APLAction{Action: &proto.APLAction_GroupReference{GroupReference: &proto.APLActionGroupReference{GroupName: name}}}
	}
	sequenceOf := func(name string, actions ...*proto.APLAction) *proto.APLAction {
		return &proto.APLAction{Action: &proto.APLAction_Sequence{Sequence: &proto.APLActionSequence{Name: name, Actions: actions}}}
	}
	config := &proto.APLRotation{
		Type: proto.APLRotation_TypeAPL,
		PrepullActions: []*proto.APLPrepullAction{
			{Action: refTo("shared"), DoAtValue: constValue("-1s")},
		},
		PriorityList: aplItems(
			refTo("opener"),
			sequenceOf("in priority", aplCast(spellID), refTo("shared")),
			refTo("shared"),
		),
		Groups: []*proto.APLGroup{
			{Name: "shared", Actions: aplItems(aplCast(spellID))},
			{Name: "opener", Actions: aplItems(sequenceOf("in opener", aplCast(spellID), refTo("shared")))},
		},
	}

	sim := setupFakeRageSimFrom(fakeRageSimRequest())
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	sim.Environment.State = Constructed
	fw.RegisterSpell(SpellConfig{
		ActionID: ActionID{SpellID: spellID},
		Flags:    SpellFlagAPL,
		Cast:     CastConfig{DefaultCast: Cast{GCD: GCDDefault}},
	})
	rot := fw.newAPLRotation(config)

	var validations [][]*proto.APLValidation
	validations = append(validations, rot.prepullValidations...)
	validations = append(validations, rot.priorityListValidations...)
	for _, group := range rot.groupListValidations {
		validations = append(validations, group...)
	}
	for _, entry := range validations {
		for _, validation := range entry {
			if validation.LogLevel == proto.LogLevel_Error {
				t.Errorf("rotation logged an error: %s", validation.Validation)
			}
		}
	}

	var references []*APLActionGroupReference
	for _, entry := range rot.allPrepullActions() {
		for _, action := range entry.GetAllActions() {
			if reference, ok := action.impl.(*APLActionGroupReference); ok {
				references = append(references, reference)
			}
		}
	}
	for _, action := range rot.allAPLActions() {
		if reference, ok := action.impl.(*APLActionGroupReference); ok {
			references = append(references, reference)
		}
	}

	copies := map[*APLGroup]bool{}
	resolved := 0
	for _, reference := range references {
		if reference.groupName != "shared" || copies[reference.group] {
			continue
		}
		if reference.group == nil {
			t.Fatalf("a reference to %q did not resolve", reference.groupName)
		}
		copies[reference.group] = true
		resolved++
	}
	if resolved != 4 {
		t.Errorf("%d references to the shared group resolved to a copy of their own, want 4", resolved)
	}
}
