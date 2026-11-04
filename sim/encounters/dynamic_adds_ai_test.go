package encounters

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/warrior/dps"
	googleproto "google.golang.org/protobuf/proto"
)

func init() {
	dps.RegisterDpsWarrior()
	common.RegisterAllEffects()
}

// The add's readings follow the boss's schedule at the preset's inputs: spawn after 10 s, despawn 20 s later, respawn
// 10 s after that. The boss stays to the fight's end.
func TestDynamicAddsSchedule(t *testing.T) {
	sim := newDynamicAddsSim(&proto.Debuffs{}, nil)
	sim.Reset()
	boss, add := sim.Encounter.AllTargets[0], sim.Encounter.AllTargets[1]

	check := func(phase string, target *core.Target, wantUp bool, wantUntilActive, wantUntilDies time.Duration) {
		t.Helper()
		if target.IsEnabled() != wantUp {
			t.Fatalf("%s at %s: %s is enabled = %t, want %t", phase, sim.CurrentTime, target.Label, target.IsEnabled(), wantUp)
		}
		if got := target.NextActiveAt(sim) - sim.CurrentTime; got != wantUntilActive {
			t.Errorf("%s at %s: Time Until Target Active(%s) = %s, want %s", phase, sim.CurrentTime, target.Label, got, wantUntilActive)
		}
		if got := target.DiesAt(sim) - sim.CurrentTime; got != wantUntilDies {
			t.Errorf("%s at %s: Time Until Target Dies(%s) = %s, want %s", phase, sim.CurrentTime, target.Label, got, wantUntilDies)
		}
	}
	stepUntilAddIs := func(up bool) {
		for add.IsEnabled() != up && !sim.Step() {
		}
	}

	check("before the first spawn", add, false, 10*time.Second, 30*time.Second)

	stepUntilAddIs(true)
	if sim.CurrentTime != 10*time.Second {
		t.Fatalf("the add spawned at %s, want 10s", sim.CurrentTime)
	}
	check("at the spawn", add, true, 0, 20*time.Second)

	stepUntilAddIs(false)
	if sim.CurrentTime != 30*time.Second {
		t.Fatalf("the add despawned at %s, want 30s", sim.CurrentTime)
	}
	check("down awaiting the respawn", add, false, 10*time.Second, 30*time.Second)
	check("down awaiting the respawn", boss, true, 0, 150*time.Second)
}

// A sim of the Dynamic Adds preset with the raid's Curse of the Elements and Sunder Armor ticked and the boss's debuff
// inputs set, and the boss's other inputs changed by setBoss when it is not nil. The warrior in the raid makes the raid's
// Sunder the player's own aura.
func newDynamicAddsDebuffSim(applyDebuffs bool, delay, jitter float64, setBoss func(boss *proto.Target)) *core.Simulation {
	return newDynamicAddsSim(&proto.Debuffs{CurseOfElements: true, SunderArmor: true}, func(boss *proto.Target) {
		boss.TargetInputs[4].BoolValue = applyDebuffs
		boss.TargetInputs[5].NumberValue = delay
		boss.TargetInputs[6].NumberValue = jitter
		if setBoss != nil {
			setBoss(boss)
		}
	})
}

// A sim of the Dynamic Adds preset with a warrior and the raid's debuffs, and the boss changed by setBoss when it is
// not nil.
func newDynamicAddsSim(debuffs *proto.Debuffs, setBoss func(boss *proto.Target)) *core.Simulation {
	player := &proto.Player{
		Name:          "Warrior",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: "30303103-050530355022100031",
		Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
			ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker},
		}}},
		Rotation: &proto.APLRotation{},
	}
	boss := googleproto.Clone(core.GetPresetTargetWithID(dynamicBossID).Config).(*proto.Target)
	if setBoss != nil {
		setBoss(boss)
	}
	return core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuffs),
		Encounter: &proto.Encounter{
			Duration: 180,
			Targets:  []*proto.Target{boss, googleproto.Clone(core.GetPresetTargetWithID(dynamicAddID).Config).(*proto.Target)},
		},
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
}

// The times in one iteration at which each aura comes on and goes off, read after every step, starting off at the
// reset. With sunderStacksAt set, it also returns the Sunder Armor stacks at that time.
func dynamicAddsDebuffFlips(t *testing.T, sim *core.Simulation, sunderStacksAt time.Duration) (coe, sunder []time.Duration, stacks int32) {
	t.Helper()
	sim.Reset()
	add := sim.Encounter.AllTargets[1]
	coeAura, sunderAura := add.GetAura("Curse of the Elements (External)"), add.GetAura("Sunder Armor (Player)")
	if coeAura == nil || sunderAura == nil {
		t.Fatal("test setup: the add carries no raid Curse of the Elements or Sunder Armor")
	}
	coeOn, sunderOn := false, false
	stacks = -1
	for {
		if coeAura.IsActive() != coeOn {
			coeOn = !coeOn
			coe = append(coe, sim.CurrentTime)
		}
		if sunderAura.IsActive() != sunderOn {
			sunderOn = !sunderOn
			sunder = append(sunder, sim.CurrentTime)
		}
		if stacks < 0 && sunderStacksAt > 0 && sim.CurrentTime >= sunderStacksAt {
			stacks = sunderAura.GetStacks()
		}
		if sim.Step() {
			break
		}
	}
	sim.Cleanup()
	return coe, sunder, stacks
}

// The add's spawns and deaths at the preset's schedule in a 180 s fight: up at 10 s, 40 s, ... and down 20 s later,
// the last death at the fight's end.
var dynamicAddsSpawns = []time.Duration{10 * time.Second, 40 * time.Second, 70 * time.Second, 100 * time.Second, 130 * time.Second, 160 * time.Second}

func dynamicAddsLives(landAfter time.Duration) []time.Duration {
	var flips []time.Duration
	for _, spawn := range dynamicAddsSpawns {
		flips = append(flips, spawn+landAfter, spawn+20*time.Second)
	}
	return flips
}

// The raid's debuffs land on the add the Debuff delay after each spawn, respawns included, and are off while it is
// down, before the first spawn too, where Sunder Armor's ramp at the pull gives the add no stack. A despawn while the
// ramp runs (delay 15 s, the ramp from 25 s to 31 s) leaves Sunder off. A second iteration starts the same.
func TestDynamicAddsRaidDebuffsLandAfterTheDelay(t *testing.T) {
	for _, tc := range []struct {
		delay      float64
		stacksAt29 int32
	}{{0, 5}, {3, 5}, {15, 3}} {
		sim := newDynamicAddsDebuffSim(true, tc.delay, 0, nil)
		want := dynamicAddsLives(core.DurationFromSeconds(tc.delay))
		for iteration := range 2 {
			coe, sunder, stacks := dynamicAddsDebuffFlips(t, sim, 29*time.Second)
			if !slices.Equal(coe, want) {
				t.Errorf("delay %v s, iteration %d: Curse of the Elements on and off at %v, want %v", tc.delay, iteration, coe, want)
			}
			if !slices.Equal(sunder, want) {
				t.Errorf("delay %v s, iteration %d: Sunder Armor on and off at %v, want %v", tc.delay, iteration, sunder, want)
			}
			if stacks != tc.stacksAt29 {
				t.Errorf("delay %v s, iteration %d: Sunder Armor stacks at 29 s = %d, want %d from its ramp", tc.delay, iteration, stacks, tc.stacksAt29)
			}
		}
	}
}

// The Debuff jitter moves each landing by up to its value either way, a roll for each spawn, and never before the
// spawn.
func TestDynamicAddsRaidDebuffsJitter(t *testing.T) {
	for _, tc := range []struct{ delay, jitter float64 }{{3, 2}, {1, 3}} {
		sim := newDynamicAddsDebuffSim(true, tc.delay, tc.jitter, nil)
		low := core.DurationFromSeconds(max(tc.delay-tc.jitter, 0))
		high := core.DurationFromSeconds(tc.delay + tc.jitter)
		seen := map[time.Duration]bool{}
		for iteration := range 2 {
			coe, _, _ := dynamicAddsDebuffFlips(t, sim, 0)
			if len(coe) != 2*len(dynamicAddsSpawns) {
				t.Fatalf("delay %v s, jitter %v s, iteration %d: Curse of the Elements on and off at %v, want once each life", tc.delay, tc.jitter, iteration, coe)
			}
			for life, spawn := range dynamicAddsSpawns {
				after := coe[2*life] - spawn
				if after < low || after > high {
					t.Errorf("delay %v s, jitter %v s: landed %s after the spawn at %s, want %s to %s", tc.delay, tc.jitter, after, spawn, low, high)
				}
				seen[after] = true
			}
		}
		if len(seen) < 2 {
			t.Errorf("delay %v s, jitter %v s: every landing came %v after its spawn, want a roll for each", tc.delay, tc.jitter, seen)
		}
	}
}

// With Apply debuffs off the add never gets the raid's debuffs, and a delay longer than the add's life never lands
// either: each despawn cancels its landing and each respawn starts over.
func TestDynamicAddsRaidDebuffsNeverLand(t *testing.T) {
	for name, sim := range map[string]*core.Simulation{
		"Apply debuffs off":        newDynamicAddsDebuffSim(false, 0, 0, nil),
		"delay longer than a life": newDynamicAddsDebuffSim(true, 25, 0, nil),
	} {
		coe, sunder, _ := dynamicAddsDebuffFlips(t, sim, 0)
		if len(coe) != 0 || len(sunder) != 0 {
			t.Errorf("%s: Curse of the Elements on and off at %v and Sunder Armor at %v, want never on", name, coe, sunder)
		}
	}
}

type sunderStacksAt struct {
	at     time.Duration
	stacks int32
}

// With a life and a respawn shorter than Sunder Armor's ramp (4 s up, 2 s down), each spawn starts one ramp, a stack
// each global cooldown from the spawn, and the ramp of the life before adds none.
func TestDynamicAddsRaidSunderRampsOncePerLife(t *testing.T) {
	sim := newDynamicAddsDebuffSim(true, 0, 0, func(boss *proto.Target) {
		boss.TargetInputs[0].NumberValue = 2
		boss.TargetInputs[1].NumberValue = 4
	})
	var want []sunderStacksAt
	for spawn := 10 * time.Second; spawn < 40*time.Second; spawn += 6 * time.Second {
		want = append(want,
			sunderStacksAt{spawn, 1},
			sunderStacksAt{spawn + core.GCDDefault, 2},
			sunderStacksAt{spawn + 2*core.GCDDefault, 3},
			sunderStacksAt{spawn + 4*time.Second, 0})
	}
	for iteration := range 2 {
		sim.Reset()
		sunder := sim.Encounter.AllTargets[1].GetAura("Sunder Armor (Player)")
		var got []sunderStacksAt
		stacks := int32(0)
		for !sim.Step() && sim.CurrentTime < 40*time.Second {
			if sunder.GetStacks() != stacks {
				stacks = sunder.GetStacks()
				got = append(got, sunderStacksAt{sim.CurrentTime, stacks})
			}
		}
		sim.Cleanup()
		if !slices.Equal(got, want) {
			t.Errorf("iteration %d: Sunder Armor stacks %v, want %v", iteration, got, want)
		}
	}
}

// With no spawn delay the add spawns inside the sim's reset, and a debuff delay of 0 puts the raid's debuffs on in that
// reset, before the raid's own reset and the prepull read them.
func TestDynamicAddsRaidDebuffsOnAtTheReset(t *testing.T) {
	sim := newDynamicAddsDebuffSim(true, 0, 0, func(boss *proto.Target) {
		boss.TargetInputs[2].NumberValue = 0
	})
	coe := sim.Encounter.AllTargets[1].GetAura("Curse of the Elements (External)")
	for iteration := range 2 {
		sim.Reset()
		if !coe.IsActive() {
			t.Errorf("iteration %d: Curse of the Elements is off on the add right after the reset, want on", iteration)
		}
		sim.Cleanup()
	}
}
