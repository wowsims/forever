package core

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func lifecycleTargets(count int) []*proto.Target {
	targets := make([]*proto.Target, count)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}
	}
	return targets
}

// A sim of the fake Protection Warrior, which only auto attacks, against the targets.
func setupLifecycleSim(targets []*proto.Target) (*Simulation, *FakeFearWarrior) {
	return setupLifecycleSimWithTanks(targets, nil)
}

func setupLifecycleSimWithTanks(targets []*proto.Target, tanks []*proto.UnitReference) (*Simulation, *FakeFearWarrior) {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: &proto.Raid{
			Tanks: tanks,
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name:      "Warrior",
					Class:     proto.Class_ClassWarrior,
					Buffs:     &proto.IndividualBuffs{},
					Spec:      &proto.Player_ProtectionWarrior{},
					Equipment: &proto.EquipmentSpec{},
					Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
				}},
				Buffs: &proto.PartyBuffs{},
			}},
		},
		Encounter: &proto.Encounter{Targets: targets, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	return sim, sim.Raid.Parties[0].Players[0].(*FakeFearWarrior)
}

func runLifecycleUntil(sim *Simulation, until time.Duration) {
	for (sim.pendingActions[len(sim.pendingActions)-1].NextActionAt <= until || sim.minTaskTime <= until || sim.minWeaponAttackTime <= until) && !sim.Step() {
	}
	sim.advance(until)
}

func mhSwingsOn(fw *FakeFearWarrior, target *Target) int32 {
	return fw.AutoAttacks.MHAuto().SpellMetrics[target.UnitIndex].Casts
}

func unitLabel(unit *Unit) string {
	if unit == nil {
		return "nil"
	}
	return unit.Label
}

func targetLabel(target *Target) string {
	if target == nil {
		return "nil"
	}
	return target.Label
}

func TestDisablingTheOnlyTargetIdlesTheRaid(t *testing.T) {
	sim, fw := setupLifecycleSim(lifecycleTargets(1))
	target := sim.Encounter.AllTargets[0]
	aa := &fw.AutoAttacks

	runLifecycleUntil(sim, 3*time.Second)
	if mhSwingsOn(fw, target) == 0 {
		t.Fatal("test setup: the warrior does not swing at the target")
	}

	target.Disable(sim, false)
	if sim.ActiveTargetCount() != 0 {
		t.Fatalf("active targets: expected none, got %d", sim.ActiveTargetCount())
	}
	if fw.CurrentTarget != &target.Unit {
		t.Fatalf("current target: expected it kept on the disabled target, got %s", unitLabel(fw.CurrentTarget))
	}
	if aa.mh.enabled || aa.oh.enabled {
		t.Fatal("auto attacks: expected them stopped with no target left")
	}

	swings := mhSwingsOn(fw, target)
	runLifecycleUntil(sim, 10*time.Second)
	if got := mhSwingsOn(fw, target); got != swings {
		t.Fatalf("MH swings while no target is active: expected %d, got %d", swings, got)
	}

	target.Enable(sim)
	if fw.CurrentTarget != &target.Unit {
		t.Fatalf("current target: expected the enabled target, got %s", unitLabel(fw.CurrentTarget))
	}
	if !aa.mh.enabled || !aa.oh.enabled {
		t.Fatal("auto attacks: expected them running again once the target is back")
	}
	runLifecycleUntil(sim, 16*time.Second)
	if mhSwingsOn(fw, target) <= swings {
		t.Fatal("MH swings: expected the warrior to swing at the target again")
	}
}

func TestDisablingTheTargetMovesTheRaidToTheNextOne(t *testing.T) {
	sim, fw := setupLifecycleSim(lifecycleTargets(2))
	first, second := sim.Encounter.AllTargets[0], sim.Encounter.AllTargets[1]
	aa := &fw.AutoAttacks

	runLifecycleUntil(sim, 2*time.Second)
	first.Disable(sim, false)
	if fw.CurrentTarget != &second.Unit {
		t.Fatalf("current target after the first goes: expected the second, got %s", unitLabel(fw.CurrentTarget))
	}
	if !aa.mh.enabled {
		t.Fatal("auto attacks: expected them running with a target left")
	}
	runLifecycleUntil(sim, 6*time.Second)
	if mhSwingsOn(fw, second) == 0 {
		t.Fatal("MH swings: expected the warrior to swing at the second target")
	}

	first.Enable(sim)
	if fw.CurrentTarget != &second.Unit {
		t.Fatalf("current target after the first comes back: expected it kept on the second, got %s", unitLabel(fw.CurrentTarget))
	}

	first.Disable(sim, false)
	second.Disable(sim, false)
	if fw.CurrentTarget != &second.Unit {
		t.Fatalf("current target with none left: expected it kept on the second, got %s", unitLabel(fw.CurrentTarget))
	}

	first.Enable(sim)
	if fw.CurrentTarget != &first.Unit {
		t.Fatalf("current target of the stranded warrior: expected the first target, got %s", unitLabel(fw.CurrentTarget))
	}
	if !aa.mh.enabled {
		t.Fatal("auto attacks: expected them running again on the first target")
	}
	swings := mhSwingsOn(fw, first)
	runLifecycleUntil(sim, 12*time.Second)
	if mhSwingsOn(fw, first) <= swings {
		t.Fatal("MH swings: expected the warrior to swing at the first target again")
	}
}

func TestEveryTargetDisabledAtStartLeavesTheRaidOnTheFirst(t *testing.T) {
	targets := lifecycleTargets(2)
	for _, target := range targets {
		target.DisabledAtStart = true
	}
	sim, fw := setupLifecycleSim(targets)

	if fw.CurrentTarget != sim.Encounter.AllTargetUnits[0] {
		t.Fatalf("current target: expected the first target, got %s", unitLabel(fw.CurrentTarget))
	}

	runLifecycleUntil(sim, time.Second)
	second := sim.Encounter.AllTargets[1]
	second.Enable(sim)
	if fw.CurrentTarget != &second.Unit {
		t.Fatalf("current target: expected the target that came, got %s", unitLabel(fw.CurrentTarget))
	}
	runLifecycleUntil(sim, 6*time.Second)
	if mhSwingsOn(fw, second) == 0 {
		t.Fatal("MH swings: expected the warrior to swing at the target that came")
	}
}

func TestTankStartsOnTheFirstEnabledTargetItTanks(t *testing.T) {
	tanks := []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	for _, tc := range []struct {
		name     string
		disabled []bool
		want     int
	}{
		{"every target enabled", []bool{false, false}, 0},
		{"the first target disabled at start", []bool{true, false}, 1},
		{"every target disabled at start", []bool{true, true}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			targets := lifecycleTargets(len(tc.disabled))
			for i, disabled := range tc.disabled {
				targets[i].DisabledAtStart = disabled
			}
			sim, fw := setupLifecycleSimWithTanks(targets, tanks)
			want := sim.Encounter.AllTargetUnits[tc.want]
			if fw.CurrentTarget != want {
				t.Fatalf("tank's target: expected %s, got %s", want.Label, unitLabel(fw.CurrentTarget))
			}
			sim.Cleanup()
			sim.Reset()
			if fw.CurrentTarget != want {
				t.Fatalf("tank's target after a reset: expected %s, got %s", want.Label, unitLabel(fw.CurrentTarget))
			}
		})
	}
}

func TestDisablingTheTargetWakesTheRotationOfTheUnitsMoved(t *testing.T) {
	sim, fw := setupLifecycleSim(lifecycleTargets(2))
	runLifecycleUntil(sim, 2*time.Second)
	fw.SetRotationTimer(sim, sim.CurrentTime+30*time.Second)

	sim.Encounter.AllTargets[0].Disable(sim, false)
	if at, latest := fw.NextRotationActionAt(), sim.CurrentTime+fw.ReactionTime; at > latest {
		t.Fatalf("next rotation look: expected it by %s, got %s", latest, at)
	}
}

func TestActiveTargetWalksWithNoTargetEnabled(t *testing.T) {
	sim, _ := setupLifecycleSim(lifecycleTargets(3))
	targets := sim.Encounter.AllTargets

	targets[0].Disable(sim, false)
	targets[2].Disable(sim, false)
	if next := targets[1].NextActiveTarget(); next != targets[1] {
		t.Fatalf("next active target of the only enabled target: expected itself, got %s", targetLabel(next))
	}
	if prev := targets[0].PreviousActiveTarget(); prev != targets[1] {
		t.Fatalf("previous active target: expected the second, got %s", targetLabel(prev))
	}

	targets[1].Disable(sim, false)
	for _, target := range targets {
		if next := target.NextActiveTarget(); next != nil {
			t.Fatalf("next active target with none enabled: expected nil, got %s", targetLabel(next))
		}
		if prev := target.PreviousActiveTarget(); prev != nil {
			t.Fatalf("previous active target with none enabled: expected nil, got %s", targetLabel(prev))
		}
		if unit := sim.Environment.NextActiveTargetUnit(&target.Unit); unit != &target.Unit {
			t.Fatalf("next active target unit with none enabled: expected the unit itself, got %s", unitLabel(unit))
		}
		if unit := sim.Environment.PreviousActiveTargetUnit(&target.Unit); unit != &target.Unit {
			t.Fatalf("previous active target unit with none enabled: expected the unit itself, got %s", unitLabel(unit))
		}
	}
	if preferred := sim.Encounter.preferredActiveTarget(); preferred != nil {
		t.Fatalf("preferred active target with none enabled: expected nil, got %s", targetLabel(preferred))
	}

	targets[2].Enable(sim)
	targets[1].Enable(sim)
	if preferred := sim.Encounter.preferredActiveTarget(); preferred != targets[1] {
		t.Fatalf("preferred active target: expected the lowest enabled index, got %s", targetLabel(preferred))
	}
	sim.Encounter.resetActiveTargetOrder()
	if !slices.Equal(sim.Encounter.ActiveTargets, []*Target{targets[1], targets[2]}) {
		t.Fatal("active targets: expected them in index order after the reorder")
	}
	if !slices.Equal(sim.Encounter.ActiveTargetUnits, []*Unit{&targets[1].Unit, &targets[2].Unit}) {
		t.Fatal("active target units: expected them in index order after the reorder")
	}
}

type lifecyclePresetAI struct{}

func (ai *lifecyclePresetAI) Initialize(*Target, *proto.Target) {}
func (ai *lifecyclePresetAI) Reset(*Simulation)                 {}
func (ai *lifecyclePresetAI) ExecuteCustomRotation(*Simulation) {}

const lifecyclePresetID = 990301

func init() {
	AddPresetTarget(&PresetTarget{
		PathPrefix: "Lifecycle Test",
		Config:     &proto.Target{Id: lifecyclePresetID, Name: "Pool Template"},
		AI:         func() TargetAI { return &lifecyclePresetAI{} },
	})
}

func TestTargetConfigsLineUpWithTargets(t *testing.T) {
	options := &proto.Encounter{Targets: lifecycleTargets(3)}
	options.Targets[2].DisabledAtStart = true
	encounter := NewEncounter(options)

	if len(encounter.targetConfigs) != len(encounter.AllTargets) {
		t.Fatalf("configs: expected %d, got %d", len(encounter.AllTargets), len(encounter.targetConfigs))
	}
	for i, config := range encounter.targetConfigs {
		if config != options.Targets[i] {
			t.Fatalf("config %d: expected the encounter's target config", i)
		}
	}

	empty := NewEncounter(&proto.Encounter{})
	if len(empty.targetConfigs) != 1 || empty.targetConfigs[0] != nil {
		t.Fatalf("configs of an encounter without targets: expected one nil config, got %v", empty.targetConfigs)
	}

	template := &proto.Target{Id: lifecyclePresetID, Name: "Add", Level: 61}
	pool := encounter.allocatePoolTargets(template, 2)
	if len(pool) != 2 || len(encounter.AllTargets) != 5 || len(encounter.AllTargetUnits) != 5 || len(encounter.targetConfigs) != 5 {
		t.Fatalf("pool: expected 2 targets appended to 3, got %d of %d", len(pool), len(encounter.AllTargets))
	}
	for i, target := range pool {
		index := int32(3 + i)
		if target.Index != index || encounter.AllTargets[index] != target || encounter.AllTargetUnits[index] != &target.Unit {
			t.Fatalf("pool target %d: expected it at index %d of the target lists", i, index)
		}
		if encounter.targetConfigs[index] != template {
			t.Fatalf("pool target %d: expected the template as its config", i)
		}
		if target.IsEnabled() {
			t.Fatalf("pool target %d: expected it disabled", i)
		}
		if target.AI != nil {
			t.Fatalf("pool target %d: expected no AI from the preset with the template's id", i)
		}
		if target.Level != template.Level {
			t.Fatalf("pool target %d: expected the template's level %d, got %d", i, template.Level, target.Level)
		}
	}
	if len(encounter.ActiveTargets) != 2 {
		t.Fatalf("active targets: expected the pool to add none, got %d", len(encounter.ActiveTargets))
	}
	if NewTarget(template, 0).AI == nil {
		t.Fatal("test setup: the preset with the template's id has no AI")
	}
}
