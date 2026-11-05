package sim

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	googleproto "google.golang.org/protobuf/proto"
)

func init() {
	RegisterAll()
}

var SimOptions = &proto.SimOptions{
	Iterations: 1,
	IsTest:     true,
}

var StandardTarget = &proto.Target{
	Stats:   stats.Stats{stats.Armor: 3731}.ToProtoArray(),
	MobType: proto.MobType_MobTypeDemon,
}

var STEncounter = &proto.Encounter{
	Duration: 180,
	Targets: []*proto.Target{
		StandardTarget,
	},
}

var BasicRaid = &proto.Raid{
	Parties: []*proto.Party{
		{
			Players: []*proto.Player{},
		},
		{
			Players: []*proto.Player{},
		},
	},
}

// Tests that we don't crash with various combinations of empty parties / blank players.
func TestSparseRaid(t *testing.T) {
	sparseRaid := &proto.Raid{
		Parties: []*proto.Party{
			{},
			{
				Players: []*proto.Player{
					{},
					{},
				},
			},
			{
				Players: []*proto.Player{
					{},
					{},
				},
			},
		},
	}

	rsr := &proto.RaidSimRequest{
		Raid:       sparseRaid,
		Encounter:  STEncounter,
		SimOptions: SimOptions,
	}

	core.RunRaidSim(rsr)
	// Don't need to check results, as long as it doesn't crash we're fine.
}

func TestBasicRaid(t *testing.T) {
	t.Skip()
	rsr := &proto.RaidSimRequest{
		Raid:       BasicRaid,
		Encounter:  STEncounter,
		SimOptions: SimOptions,
	}

	core.RaidSimTest("P1 ST", t, rsr, 6323.79)
}

// To quickly debug raid sim issues, uncomment this test and copy in a request string.
/*
func testRaidString(t *testing.T, raidString string) {
	rsr := &proto.RaidSimRequest{}

	data := []byte(raidString)
	if err := protojson.Unmarshal(data, rsr); err != nil {
		panic(err)
	}

	core.RunRaidSim(rsr)
	//core.RaidSimTest("Fixed Raid", t, rsr, 10000.00)
}

func TestFixedRaid(t *testing.T) {
 	testRaidString(t, `
 	`)
}
*/

// Every preset target simulates on its own as a plain target, with no design.
func TestPresetTargetsAlone(t *testing.T) {
	for _, path := range []string{"Default/Raid Target", "Default/Movement", "Default/Dynamic Boss", "Default/Dynamic Add", "Default/Boss"} {
		preset := core.GetPresetTargetWithPath(path)
		if preset == nil {
			t.Fatalf("the preset target %s is not registered", path)
		}
		t.Run(preset.Config.Name, func(t *testing.T) {
			result := core.RunRaidSim(&proto.RaidSimRequest{
				Raid:       BasicRaid,
				Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{preset.Config}},
				SimOptions: SimOptions,
			})
			if result.Error != nil {
				t.Fatalf("the encounter with the target alone failed: %s", result.Error.Message)
			}
		})
	}
}

// The Dynamic Adds preset encounter, run with its targets and its design, spawns an add wave and despawns it:
// a 180 s fight spawns at 10 s, 40 s, ... and each add lives 20 s.
func TestDynamicAddsPresetEncounter(t *testing.T) {
	preset := dynamicAddsPreset(t)
	if len(preset.Targets) != 1 || preset.Design == nil || len(preset.Design.AddWaves) != 1 {
		t.Fatalf("the preset encounter has %d targets and design %v, want one target and one add wave", len(preset.Targets), preset.Design)
	}
	targets := make([]*proto.Target, len(preset.Targets))
	for i, target := range preset.Targets {
		targets[i] = target.Target
	}
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       BasicRaid,
		Encounter:  &proto.Encounter{Duration: 180, Targets: targets, Design: preset.Design},
		SimOptions: &proto.SimOptions{Iterations: 1, IsTest: true, Debug: true},
	})
	if result.Error != nil {
		t.Fatalf("the preset encounter failed: %s", result.Error.Message)
	}
	spawns := logTimes(result.Logs, "wave Add spawned 1 adds")
	despawns := logTimes(result.Logs, "despawns from wave Add")
	wantSpawns := []float64{10, 40, 70, 100, 130, 160}
	wantDespawns := []float64{30, 60, 90, 120, 150, 180}
	if !slices.Equal(spawns, wantSpawns) {
		t.Errorf("the add wave spawned at %v, want %v", spawns, wantSpawns)
	}
	if !slices.Equal(despawns, wantDespawns) && !slices.Equal(despawns, wantDespawns[:5]) {
		t.Errorf("the adds despawned at %v, want 20 s after each spawn: %v", despawns, wantDespawns)
	}
}

func dynamicAddsPreset(t *testing.T) *proto.PresetEncounter {
	t.Helper()
	for _, encounter := range core.PresetEncounters {
		if encounter.Path == "Default/Dynamic Adds" {
			return encounter
		}
	}
	t.Fatal("the Dynamic Adds preset encounter is not registered")
	return nil
}

// The raid's debuffs the debuff tests tick. The warrior in the raid makes the raid's Sunder the player's own aura.
var raidCurseAndSunder = &proto.Debuffs{CurseOfElements: true, SunderArmor: true}

// The Dynamic Adds preset in a 180 s fight, its design changed by setDesign when it is not nil.
func dynamicAddsEncounter(t *testing.T, setDesign func(design *proto.EncounterDesign)) *proto.Encounter {
	t.Helper()
	preset := dynamicAddsPreset(t)
	targets := make([]*proto.Target, len(preset.Targets))
	for i, target := range preset.Targets {
		targets[i] = target.Target
	}
	design := googleproto.Clone(preset.Design).(*proto.EncounterDesign)
	if setDesign != nil {
		setDesign(design)
	}
	return &proto.Encounter{Duration: 180, Targets: targets, Design: design}
}

// A sim of the Dynamic Adds preset with the raid's Curse of the Elements and Sunder Armor ticked and the add wave's
// debuff settings set.
func newDynamicAddsDebuffSim(t *testing.T, apply bool, delay, jitter float64) *core.Simulation {
	return newWarriorSim(raidCurseAndSunder, dynamicAddsEncounter(t, func(design *proto.EncounterDesign) {
		setWaveDebuffs(design.AddWaves[0], apply, delay, jitter)
	}))
}

func setWaveDebuffs(wave *proto.AddWave, apply bool, delay, jitter float64) {
	wave.HoldDebuffs = !apply
	wave.DebuffDelay = delay
	wave.DebuffJitter = jitter
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
// starting off at the reset and rounded to the millisecond, since a design's times are fractions of the fight. With
// stacksAt set, it also returns the Sunder Armor stacks at that time.
func raidDebuffFlips(t *testing.T, sim *core.Simulation, targetIndex int, stacksAt time.Duration) (coe, sunder []time.Duration, stacks int32) {
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
		now := sim.CurrentTime.Round(time.Millisecond)
		if coeAura.IsActive() != coeOn {
			coeOn = !coeOn
			coe = append(coe, now)
		}
		if sunderAura.IsActive() != sunderOn {
			sunderOn = !sunderOn
			sunder = append(sunder, now)
		}
		if stacksAt > 0 && sim.CurrentTime < stacksAt {
			stacks = sunderAura.GetStacks()
		}
		if sim.Step() {
			break
		}
	}
	sim.Cleanup()
	return coe, sunder, stacks
}

// The add's spawns at the preset's schedule in a 180 s fight: up at 10 s, 40 s, ... and down 20 s later. The last
// life lasts to the fight's end, which no despawn comes before.
var dynamicAddsSpawns = []time.Duration{10 * time.Second, 40 * time.Second, 70 * time.Second, 100 * time.Second, 130 * time.Second, 160 * time.Second}

func dynamicAddsLives(landAfter time.Duration) []time.Duration {
	var flips []time.Duration
	for _, spawn := range dynamicAddsSpawns {
		flips = append(flips, spawn+landAfter, spawn+20*time.Second)
	}
	return flips[:len(flips)-1]
}

// The raid's debuffs land on the add the wave's Debuff delay after each spawn, respawns included, and are off while it
// is down, before the first spawn too, where Sunder Armor's ramp at the pull gives the add no stack. A despawn while
// the ramp runs (delay 15 s, the ramp from 25 s to 31 s) leaves Sunder off. A second iteration starts the same.
func TestDynamicAddsRaidDebuffsLandAfterTheDelay(t *testing.T) {
	for _, tc := range []struct {
		delay      float64
		stacksAt29 int32
	}{{0, 5}, {3, 5}, {15, 3}} {
		sim := newDynamicAddsDebuffSim(t, true, tc.delay, 0)
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
		sim := newDynamicAddsDebuffSim(t, true, tc.delay, tc.jitter)
		low := core.DurationFromSeconds(max(tc.delay-tc.jitter, 0))
		high := core.DurationFromSeconds(tc.delay + tc.jitter)
		seen := map[time.Duration]bool{}
		for iteration := range 2 {
			coe, _, _ := raidDebuffFlips(t, sim, 1, 0)
			if len(coe) != 2*len(dynamicAddsSpawns)-1 {
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
		encounter := dynamicAddsEncounter(t, func(design *proto.EncounterDesign) {
			setWaveDebuffs(design.AddWaves[0], true, 3, 2)
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

// With the wave's Raid debuffs off the add never gets the raid's debuffs, and a delay longer than the add's life never
// lands either: each despawn cancels its landing and each respawn starts over.
func TestDynamicAddsRaidDebuffsNeverLand(t *testing.T) {
	for name, sim := range map[string]*core.Simulation{
		"Raid debuffs off":         newDynamicAddsDebuffSim(t, false, 0, 0),
		"delay longer than a life": newDynamicAddsDebuffSim(t, true, 25, 0),
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
	sim := newWarriorSim(raidCurseAndSunder, dynamicAddsEncounter(t, func(design *proto.EncounterDesign) {
		design.AddWaves[0].Lifetime = 4.0 / 180
		design.AddWaves[0].Repeat.Interval = 6.0 / 180
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
				got = append(got, sunderStacksAt{sim.CurrentTime.Round(time.Millisecond), stacks})
			}
		}
		sim.Cleanup()
		if !slices.Equal(got, want) {
			t.Errorf("iteration %d: Sunder Armor stacks %v, want %v", iteration, got, want)
		}
	}
}

// The wave holds the raid's debuffs before the add's own reset, so that reset puts none on, and a spawn at the pull
// with no debuff delay puts them on at the spawn.
func TestDynamicAddsRaidDebuffsOnAtThePullSpawn(t *testing.T) {
	sim := newWarriorSim(raidCurseAndSunder, dynamicAddsEncounter(t, func(design *proto.EncounterDesign) {
		design.AddWaves[0].Spawn.At = 0
	}))
	add := sim.Encounter.AllTargets[1]
	coe := add.GetAura("Curse of the Elements (External)")
	gains := 0
	coe.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) { gains++ })
	for iteration := range 2 {
		gains = 0
		sim.Reset()
		if gains != 0 || coe.IsActive() {
			t.Errorf("iteration %d: Curse of the Elements came on %d times on the add in the reset, active after it %t, want never", iteration, gains, coe.IsActive())
		}
		sim.PrePull()
		for !add.IsEnabled() && !sim.Step() {
		}
		if sim.CurrentTime != 0 || !coe.IsActive() {
			t.Errorf("iteration %d: the add spawned at %s with Curse of the Elements on %t, want at 0 with it on", iteration, sim.CurrentTime, coe.IsActive())
		}
		sim.Cleanup()
	}
}

// A downtime of the boss from start to end, fractions of the fight.
func bossDowntime(start, end float64, clear bool) *proto.DesignDowntime {
	return &proto.DesignDowntime{Start: &proto.DesignTime{At: start}, End: &proto.DesignTime{At: end}, ClearDebuffs: clear}
}

// A downtime that clears debuffs takes the raid's permanent debuffs off the boss at its start and puts them back at
// its end, where Sunder Armor ramps up anew; one that keeps them leaves them on.
func TestDesignClearDebuffsDowntimeRemovesAndRestoresTheRaidDebuffs(t *testing.T) {
	for _, clear := range []bool{true, false} {
		sim := newWarriorSim(raidCurseAndSunder, dynamicAddsEncounter(t, func(design *proto.EncounterDesign) {
			design.Downtimes = []*proto.DesignDowntime{bossDowntime(0.25, 0.5, clear)}
		}))
		want, wantStacks := []time.Duration{0}, int32(5)
		if clear {
			want, wantStacks = []time.Duration{0, 45 * time.Second, 90 * time.Second}, 2
		}
		for iteration := range 2 {
			coe, sunder, stacks := raidDebuffFlips(t, sim, 0, 92*time.Second)
			if !slices.Equal(coe, want) || !slices.Equal(sunder, want) {
				t.Errorf("clear %t, iteration %d: Curse of the Elements on and off on the boss at %v and Sunder Armor at %v, want %v", clear, iteration, coe, sunder, want)
			}
			if stacks != wantStacks {
				t.Errorf("clear %t, iteration %d: Sunder Armor stacks on the boss at 92 s = %d, want %d", clear, iteration, stacks, wantStacks)
			}
		}
	}
}

// Overlapping downtimes that clear debuffs put the raid's back on the boss as the last of them ends, one that keeps
// them leaves them on the boss while it is still down, and one that never ends leaves them off until the next reset.
func TestDesignOverlappingClearDebuffsDowntimes(t *testing.T) {
	for name, tc := range map[string]struct {
		downtimes []*proto.DesignDowntime
		want      []time.Duration
	}{
		"both clear":            {[]*proto.DesignDowntime{bossDowntime(0.25, 0.5, true), bossDowntime(1.0/3, 2.0/3, true)}, []time.Duration{0, 45 * time.Second, 120 * time.Second}},
		"the second keeps them": {[]*proto.DesignDowntime{bossDowntime(0.25, 0.5, true), bossDowntime(1.0/3, 2.0/3, false)}, []time.Duration{0, 45 * time.Second, 90 * time.Second}},
		"never ends":            {[]*proto.DesignDowntime{bossDowntime(0.25, 1, true)}, []time.Duration{0, 45 * time.Second}},
	} {
		sim := newWarriorSim(raidCurseAndSunder, dynamicAddsEncounter(t, func(design *proto.EncounterDesign) { design.Downtimes = tc.downtimes }))
		for iteration := range 2 {
			coe, sunder, _ := raidDebuffFlips(t, sim, 0, 0)
			if !slices.Equal(coe, tc.want) || !slices.Equal(sunder, tc.want) {
				t.Errorf("%s, iteration %d: Curse of the Elements on and off on the boss at %v and Sunder Armor at %v, want %v", name, iteration, coe, sunder, tc.want)
			}
		}
	}
}

// With the boss down from the pull, the add's raid debuffs land while the add is still disabled, so no player's
// rotation woken by the add's enabling acts on it before they are on.
func TestDynamicAddsRaidDebuffsLandBeforeTheAddIsEnabled(t *testing.T) {
	sim := newWarriorSim(raidCurseAndSunder, dynamicAddsEncounter(t, func(design *proto.EncounterDesign) {
		design.AddWaves[0].Spawn.At = 0
		design.Downtimes = []*proto.DesignDowntime{bossDowntime(0, 1, false)}
	}))
	add := sim.Encounter.AllTargets[1]
	var enabledAtGain []bool
	add.GetAura("Curse of the Elements (External)").ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
		enabledAtGain = append(enabledAtGain, add.IsEnabled())
	})
	for iteration := range 2 {
		enabledAtGain = nil
		raidDebuffFlips(t, sim, 1, 0)
		if len(enabledAtGain) == 0 {
			t.Fatalf("iteration %d: Curse of the Elements never came on the add", iteration)
		}
		for _, enabled := range enabledAtGain {
			if enabled {
				t.Errorf("iteration %d: Curse of the Elements came on the add after it was enabled, want before: %v", iteration, enabledAtGain)
				break
			}
		}
	}
}

// The encounter's metrics list the boss and then every add of every wave, each naming the wave and add it spawns
// as, and an area spell cast while the adds are up damages each of them.
func TestDesignAddsReachTheEncounterMetrics(t *testing.T) {
	add := &proto.Target{Name: "Son of Flame", Level: 60, MobType: proto.MobType_MobTypeElemental}
	design := &proto.EncounterDesign{AddWaves: []*proto.AddWave{
		{Name: "Adds 1", Adds: []*proto.Target{add, add, add}, Spawn: &proto.DesignTime{}, Lifetime: 0.5},
		{Name: "Adds 2", Adds: []*proto.Target{add}, Spawn: &proto.DesignTime{At: 0.5}, Lifetime: 0.5},
	}}
	arcaneExplosion := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 10202}},
	}}}}
	result := core.RunRaidSimConcurrent(&proto.RaidSimRequest{
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{
			Name:      "Mage",
			Class:     proto.Class_ClassMage,
			Race:      proto.Race_RaceHuman,
			Equipment: &proto.EquipmentSpec{},
			Buffs:     &proto.IndividualBuffs{},
			Spec:      &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{ClassOptions: &proto.MageOptions{}}}},
			Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: []*proto.APLListItem{arcaneExplosion}},
		}}, Buffs: &proto.PartyBuffs{}}}, Buffs: &proto.RaidBuffs{}},
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{StandardTarget}, Design: design},
		SimOptions: &proto.SimOptions{Iterations: 8, IsTest: true},
	})
	if result.Error != nil {
		t.Fatalf("the sim failed: %s", result.Error.Message)
	}

	targets := result.EncounterMetrics.Targets
	if len(targets) != 5 {
		t.Fatalf("encounter metrics: expected the boss and 4 adds, got %d targets", len(targets))
	}
	if targets[0].DesignAdd != nil {
		t.Errorf("the boss: expected no design add, got %v", targets[0].DesignAdd)
	}
	wantAdds := [][2]int32{{0, 0}, {0, 1}, {0, 2}, {1, 0}}
	for i, want := range wantAdds {
		got := targets[i+1].DesignAdd
		if got == nil || got.WaveIndex != want[0] || got.AddIndex != want[1] {
			t.Errorf("target %d: expected wave %d add %d, got %v", i+2, want[0], want[1], got)
		}
	}

	damage := map[int32]float64{}
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		for _, target := range action.Targets {
			damage[target.UnitIndex] += target.Damage
		}
	}
	for _, target := range targets {
		if damage[target.UnitIndex] <= 0 {
			t.Errorf("%s: expected damage from Arcane Explosion, got none", target.Name)
		}
	}
}

// The seconds of the debug log lines that contain text.
func logTimes(logs string, text string) []float64 {
	var times []float64
	for _, line := range strings.Split(logs, "\n") {
		if !strings.Contains(line, text) {
			continue
		}
		match := logTimeRegexp.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		seconds, _ := strconv.ParseFloat(match[1], 64)
		times = append(times, seconds)
	}
	return times
}

var logTimeRegexp = regexp.MustCompile(`^\[(\d+(?:\.\d+)?)\]`)
