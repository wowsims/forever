package core

import (
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
	return setupLifecycleSimWith(targets, tanks, nil)
}

// register runs while the environment is constructing, so it may register spells and auras.
func setupLifecycleSimWith(targets []*proto.Target, tanks []*proto.UnitReference, register func(char *Character)) (*Simulation, *FakeFearWarrior) {
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
	fw := sim.Raid.Parties[0].Players[0].(*FakeFearWarrior)
	if register != nil {
		sim.Environment.State = Constructed
		register(&fw.Character)
		sim.Environment.State = Finalized
	}
	sim.Reset()
	sim.PrePull()
	return sim, fw
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
}

// A sim against two targets with the spell registered, at the pull, and a count of its ticks.
func setupTickCounter(config SpellConfig) (*Simulation, *FakeFearWarrior, *Spell, *int) {
	ticks := new(int)
	config.Dot.OnTick = func(*Simulation, *Unit, *Dot) { *ticks++ }
	var spell *Spell
	sim, fw := setupLifecycleSimWith(lifecycleTargets(2), nil, func(char *Character) {
		spell = char.RegisterSpell(config)
	})
	spell.finalize()
	runLifecycleUntil(sim, 0)
	return sim, fw, spell, ticks
}

func TestChannelEndsWhenItsTargetGoesDown(t *testing.T) {
	for _, tc := range []struct {
		name   string
		goDown func(sim *Simulation, target *Target)
	}{
		{"disabled keeping its debuffs", func(sim *Simulation, target *Target) { target.Disable(sim, false) }},
		{"disabled clearing its debuffs", func(sim *Simulation, target *Target) { target.Disable(sim, true) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, fw, spell, ticks := setupTickCounter(manaChannelConfig(5, false))
			target := sim.Encounter.AllTargets[0]
			if !spell.Cast(sim, &target.Unit) || !fw.IsChanneling() {
				t.Fatal("test setup: expected the channel running")
			}

			// The target goes as a tick comes due, before the tick runs.
			runLifecycleUntil(sim, 3*time.Second-time.Millisecond)
			sim.advance(3 * time.Second)
			tc.goDown(sim, target)
			if fw.IsChanneling() || spell.Dot(&target.Unit).IsActive() {
				t.Fatal("expected the channel ended with its target gone")
			}
			if fw.CurrentTarget != sim.Encounter.AllTargetUnits[1] {
				t.Errorf("current target: expected the second, got %s", unitLabel(fw.CurrentTarget))
			}

			runLifecycleUntil(sim, 10*time.Second)
			if *ticks != 2 {
				t.Errorf("ticks: expected the 2 before the target went, got %d", *ticks)
			}
		})
	}
}

func TestShootEndsWhenItsTargetGoesDown(t *testing.T) {
	ct := newShootTest()
	ct.start(ct.channelOf(ct.spell, "true", false))
	ct.runUntil(3 * time.Second)
	shots := len(ct.shots)
	if shots == 0 || !ct.wanding() {
		t.Fatalf("test setup: expected Shoot running with shots fired, got %d", shots)
	}

	ct.sim.Encounter.AllTargets[0].Disable(ct.sim, false)
	if ct.wanding() {
		t.Fatal("expected Shoot ended with its target gone")
	}
	ct.runUntil(10 * time.Second)
	if len(ct.shots) != shots {
		t.Errorf("shots: expected the %d before the target went, got %d", shots, len(ct.shots))
	}
}

// Only the target going down ends anything: a channel on another target, a ground channel and a dot of the
// target disabled without clearing its debuffs all tick on.
func TestTargetGoingDownLeavesOtherPeriodicsTicking(t *testing.T) {
	ground := manaChannelConfig(5, false)
	ground.Dot.IsAOE = true
	ground.ApplyEffects = func(sim *Simulation, _ *Unit, spell *Spell) { spell.AOEDot().Apply(sim) }
	for _, tc := range []struct {
		name   string
		config SpellConfig
		target int
	}{
		{"a channel on another target", manaChannelConfig(5, false), 1},
		{"a ground channel", ground, 0},
		{"a dot on the target", fakeWarlockDot(910310, "Plain Dot", ClassFlags{}, SpellFlagAPL), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, _, spell, ticks := setupTickCounter(tc.config)
			if !spell.Cast(sim, sim.Encounter.AllTargetUnits[tc.target]) {
				t.Fatal("test setup: expected the spell cast")
			}
			runLifecycleUntil(sim, 1500*time.Millisecond)
			sim.Encounter.AllTargets[0].Disable(sim, false)
			runLifecycleUntil(sim, 10*time.Second)
			if *ticks != 5 {
				t.Errorf("ticks: expected all 5, got %d", *ticks)
			}
		})
	}
}

// A target disabled at start takes no swings at the pull: the raid has nothing to swing at until it is
// enabled, and then swings at it.
func TestOnlyTargetDisabledAtStartIdlesTheRaidUntilItComes(t *testing.T) {
	targets := lifecycleTargets(1)
	targets[0].DisabledAtStart = true
	sim, fw := setupLifecycleSim(targets)
	target := sim.Encounter.AllTargets[0]

	runLifecycleUntil(sim, 5*time.Second)
	if aa := &fw.AutoAttacks; aa.mh.enabled || aa.oh.enabled {
		t.Error("auto attacks before the target comes: expected them stopped")
	}
	if got := mhSwingsOn(fw, target); got != 0 {
		t.Errorf("MH swings at the target disabled at start: expected none, got %d", got)
	}

	target.Enable(sim)
	runLifecycleUntil(sim, 10*time.Second)
	if got := mhSwingsOn(fw, target); got == 0 {
		t.Error("MH swings once the target comes: expected some, got none")
	}
}

func TestRaidStartsOnTheLowestTargetNotDisabledAtStart(t *testing.T) {
	targets := lifecycleTargets(3)
	targets[0].DisabledAtStart = true
	sim, fw := setupLifecycleSim(targets)
	first := sim.Encounter.AllTargets[1]

	runLifecycleUntil(sim, 5*time.Second)
	if fw.CurrentTarget != &first.Unit {
		t.Fatalf("player's target at the pull: expected %s, got %s", first.Label, unitLabel(fw.CurrentTarget))
	}
	if got := mhSwingsOn(fw, first); got == 0 {
		t.Error("MH swings at the first active target: expected some, got none")
	}
	if got := mhSwingsOn(fw, sim.Encounter.AllTargets[0]); got != 0 {
		t.Errorf("MH swings at the target disabled at start: expected none, got %d", got)
	}
}

// A target disabled mid-hardcast stops the cast, so the spell does not land when the hardcast would have ended,
// nor once the target is back.
func TestDisablingATargetInterruptsItsHardcast(t *testing.T) {
	const castTime = 4 * time.Second
	landed := 0
	var spell *Spell
	sim, fw := setupLifecycleSimWith(lifecycleTargets(1), nil, func(char *Character) {
		spell = char.Env.Encounter.AllTargets[0].RegisterSpell(SpellConfig{
			ActionID:     ActionID{SpellID: 910400},
			Cast:         CastConfig{DefaultCast: Cast{CastTime: castTime}},
			ApplyEffects: func(*Simulation, *Unit, *Spell) { landed++ },
		})
	})
	spell.finalize()
	target := sim.Encounter.AllTargets[0]
	runLifecycleUntil(sim, time.Second)
	if !spell.Cast(sim, &fw.Unit) || target.Hardcast.Expires <= sim.CurrentTime {
		t.Fatal("test setup: expected the target mid-hardcast")
	}

	target.Disable(sim, false)
	if target.Hardcast.Expires > sim.CurrentTime {
		t.Errorf("the disable left the hardcast running until %s", target.Hardcast.Expires)
	}
	runLifecycleUntil(sim, 3*time.Second)
	target.Enable(sim)
	runLifecycleUntil(sim, 10*time.Second)
	if landed != 0 {
		t.Errorf("the interrupted hardcast landed %d times, want never", landed)
	}
}
