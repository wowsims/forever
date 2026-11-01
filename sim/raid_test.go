package sim

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
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
	var preset *proto.PresetEncounter
	for _, encounter := range core.PresetEncounters {
		if encounter.Path == "Default/Dynamic Adds" {
			preset = encounter
		}
	}
	if preset == nil {
		t.Fatal("the Dynamic Adds preset encounter is not registered")
	}
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
