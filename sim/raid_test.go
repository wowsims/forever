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
	for _, id := range []int32{31146, 31147, 99999, 99998, 100000} {
		preset := core.GetPresetTargetWithID(id)
		if preset == nil {
			t.Fatalf("the preset target %d is not registered", id)
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
