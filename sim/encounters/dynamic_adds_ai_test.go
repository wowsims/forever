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
	sim := newWarriorSim(&proto.Debuffs{}, dynamicAddsEncounter(nil))
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

// The raid's debuffs the debuff tests tick. The warrior in the raid makes the raid's Sunder the player's own aura.
var raidCurseAndSunder = &proto.Debuffs{CurseOfElements: true, SunderArmor: true}

// A sim of the Dynamic Adds preset with the raid's Curse of the Elements and Sunder Armor ticked and the boss's add
// debuff inputs set.
func newDynamicAddsDebuffSim(applyDebuffs bool, delay, jitter float64) *core.Simulation {
	return newWarriorSim(raidCurseAndSunder, dynamicAddsEncounter(func(boss *proto.Target) {
		setAddDebuffInputs(boss, applyDebuffs, delay, jitter)
	}))
}

// Sets the boss's Apply debuffs to adds, Debuff delay and Debuff jitter inputs.
func setAddDebuffInputs(boss *proto.Target, applyDebuffs bool, delay, jitter float64) {
	boss.TargetInputs[4].BoolValue = applyDebuffs
	boss.TargetInputs[5].NumberValue = delay
	boss.TargetInputs[6].NumberValue = jitter
}

// The Dynamic Adds preset in a 180 s fight, the boss changed by setBoss when it is not nil.
func dynamicAddsEncounter(setBoss func(boss *proto.Target)) *proto.Encounter {
	boss := googleproto.Clone(core.GetPresetTargetWithID(dynamicBossID).Config).(*proto.Target)
	if setBoss != nil {
		setBoss(boss)
	}
	return &proto.Encounter{
		Duration: 180,
		Targets:  []*proto.Target{boss, googleproto.Clone(core.GetPresetTargetWithID(dynamicAddID).Config).(*proto.Target)},
	}
}

// A sim of the encounter with a warrior and the raid's debuffs.
func newWarriorSim(debuffs *proto.Debuffs, encounter *proto.Encounter) *core.Simulation {
	return newWarriorSimWithGear(debuffs, encounter, &proto.EquipmentSpec{})
}

// A sim of the encounter with a warrior wearing the gear and the raid's debuffs.
func newWarriorSimWithGear(debuffs *proto.Debuffs, encounter *proto.Encounter, gear *proto.EquipmentSpec) *core.Simulation {
	player := &proto.Player{
		Name:          "Warrior",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     gear,
		TalentsString: "30303103-050530355022100031",
		Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
			ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker},
		}}},
		Rotation: &proto.APLRotation{},
	}
	return core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuffs),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
}

// The times in one iteration at which each raid debuff on the target comes on and goes off, read after every step,
// starting off at the reset. With sunderStacksAt set, it also returns the Sunder Armor stacks at that time.
func raidDebuffFlips(t *testing.T, sim *core.Simulation, targetIndex int, sunderStacksAt time.Duration) (coe, sunder []time.Duration, stacks int32) {
	t.Helper()
	sim.Reset()
	sim.PrePull()
	target := sim.Encounter.AllTargets[targetIndex]
	coeAura, sunderAura := target.GetAura("Curse of the Elements (External)"), target.GetAura("Sunder Armor (Player)")
	if coeAura == nil || sunderAura == nil {
		t.Fatalf("test setup: %s carries no raid Curse of the Elements or Sunder Armor", target.Label)
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

// The raid's debuffs land on the add the boss's Debuff delay after each spawn, respawns included, and are off while it is
// down, before the first spawn too, where Sunder Armor's ramp at the pull gives the add no stack. A despawn while the
// ramp runs (delay 15 s, the ramp from 25 s to 31 s) leaves Sunder off. A second iteration starts the same.
func TestDynamicAddsRaidDebuffsLandAfterTheDelay(t *testing.T) {
	for _, tc := range []struct {
		delay      float64
		stacksAt29 int32
	}{{0, 5}, {3, 5}, {15, 3}} {
		sim := newDynamicAddsDebuffSim(true, tc.delay, 0)
		want := dynamicAddsLives(core.DurationFromSeconds(tc.delay))
		for iteration := range 2 {
			coe, sunder, stacks := raidDebuffFlips(t, sim, 1, 29*time.Second)
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
		sim := newDynamicAddsDebuffSim(true, tc.delay, tc.jitter)
		low := core.DurationFromSeconds(max(tc.delay-tc.jitter, 0))
		high := core.DurationFromSeconds(tc.delay + tc.jitter)
		seen := map[time.Duration]bool{}
		for iteration := range 2 {
			coe, _, _ := raidDebuffFlips(t, sim, 1, 0)
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

// The Debuff jitter rolls apart from the combat rolls: a warrior who swings a weapon rolls more in combat than one who
// fights unarmed, and with the same seed the add's debuffs land at the same times for both.
func TestDynamicAddsRaidDebuffsJitterApartFromCombatRolls(t *testing.T) {
	newSim := func(gear *proto.EquipmentSpec) *core.Simulation {
		encounter := dynamicAddsEncounter(func(boss *proto.Target) {
			setAddDebuffInputs(boss, true, 3, 2)
		})
		return newWarriorSimWithGear(raidCurseAndSunder, encounter, gear)
	}
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 9449}
	coeLandings := func(sim *core.Simulation) []time.Duration {
		coe, _, _ := raidDebuffFlips(t, sim, 1, 0)
		var landings []time.Duration
		for i := 0; i < len(coe); i += 2 {
			landings = append(landings, coe[i])
		}
		return landings
	}
	unarmed, armed := newSim(&proto.EquipmentSpec{}), newSim(&proto.EquipmentSpec{Items: items})
	for iteration := range 2 {
		unarmedCoe, armedCoe := coeLandings(unarmed), coeLandings(armed)
		if len(unarmedCoe) != len(dynamicAddsSpawns) || !slices.Equal(unarmedCoe, armedCoe) {
			t.Errorf("iteration %d: Curse of the Elements landed at %v unarmed and at %v with a weapon, want once each life at the same times", iteration, unarmedCoe, armedCoe)
		}
	}
}

// With Apply debuffs off the add never gets the raid's debuffs, and a delay longer than the add's life never lands
// either: each despawn cancels its landing and each respawn starts over.
func TestDynamicAddsRaidDebuffsNeverLand(t *testing.T) {
	for name, sim := range map[string]*core.Simulation{
		"Apply debuffs off":        newDynamicAddsDebuffSim(false, 0, 0),
		"delay longer than a life": newDynamicAddsDebuffSim(true, 25, 0),
	} {
		coe, sunder, _ := raidDebuffFlips(t, sim, 1, 0)
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
	sim := newWarriorSim(raidCurseAndSunder, dynamicAddsEncounter(func(boss *proto.Target) {
		boss.TargetInputs[0].NumberValue = 2
		boss.TargetInputs[1].NumberValue = 4
	}))
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
	sim := newWarriorSim(raidCurseAndSunder, dynamicAddsEncounter(func(boss *proto.Target) {
		boss.TargetInputs[2].NumberValue = 0
	}))
	coe := sim.Encounter.AllTargets[1].GetAura("Curse of the Elements (External)")
	for iteration := range 2 {
		sim.Reset()
		if !coe.IsActive() {
			t.Errorf("iteration %d: Curse of the Elements is off on the add right after the reset, want on", iteration)
		}
		sim.Cleanup()
	}
}

// A boss saved with only the respawn, lifetime, spawn delay and Kill share % inputs applies the raid's debuffs to the
// add at each spawn with no delay.
func TestDynamicAddsRaidDebuffsWithoutTheDebuffInputs(t *testing.T) {
	sim := newWarriorSim(raidCurseAndSunder, dynamicAddsEncounter(func(boss *proto.Target) {
		boss.TargetInputs = boss.TargetInputs[:4]
	}))
	want := dynamicAddsLives(0)
	for iteration := range 2 {
		coe, sunder, _ := raidDebuffFlips(t, sim, 1, 0)
		if !slices.Equal(coe, want) || !slices.Equal(sunder, want) {
			t.Errorf("iteration %d: Curse of the Elements on and off at %v and Sunder Armor at %v, want %v", iteration, coe, sunder, want)
		}
	}
}
