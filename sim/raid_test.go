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
	googleProto "google.golang.org/protobuf/proto"
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

// The Dynamic Adds preset's two targets each work on their own: an encounter with the add alone, or the boss alone,
// simulates as plain targets.
func TestDynamicAddsTargetsAlone(t *testing.T) {
	boss := core.GetPresetTargetWithID(99999)
	add := core.GetPresetTargetWithID(99998)
	if boss == nil || add == nil {
		t.Fatal("the Dynamic Adds preset targets are not registered")
	}
	for name, target := range map[string]*proto.Target{"add": add.Config, "boss": boss.Config} {
		t.Run(name, func(t *testing.T) {
			result := core.RunRaidSim(&proto.RaidSimRequest{
				Raid:       BasicRaid,
				Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{target}},
				SimOptions: SimOptions,
			})
			if result.Error != nil {
				t.Fatalf("the encounter with the %s alone failed: %s", name, result.Error.Message)
			}
		})
	}
}

// The Dynamic Adds preset's targets. A killSharePercent other than nil sets the boss's Kill share % input.
func dynamicAddsTargets(t *testing.T, killSharePercent *float64) []*proto.Target {
	t.Helper()
	var preset *proto.PresetEncounter
	for _, encounter := range core.PresetEncounters {
		if encounter.Path == "Default/Dynamic Adds" {
			preset = encounter
		}
	}
	if preset == nil || len(preset.Targets) != 2 {
		t.Fatal("the Dynamic Adds preset encounter with its boss and add is not registered")
	}
	boss := googleProto.Clone(preset.Targets[0].Target).(*proto.Target)
	if killSharePercent != nil {
		boss.TargetInputs[3].NumberValue = *killSharePercent
	}
	return []*proto.Target{boss, preset.Targets[1].Target}
}

func runDynamicAdds(t *testing.T, duration float64, killSharePercent *float64) string {
	t.Helper()
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       BasicRaid,
		Encounter:  &proto.Encounter{Duration: duration, Targets: dynamicAddsTargets(t, killSharePercent)},
		SimOptions: &proto.SimOptions{Iterations: 1, IsTest: true, Debug: true},
	})
	if result.Error != nil {
		t.Fatalf("the Dynamic Adds encounter failed: %s", result.Error.Message)
	}
	return result.Logs
}

// The Dynamic Adds boss spawns its add 10 s in and every 30 s after, and the add dies 20 s after each spawn: a 180 s
// fight spawns at 10 s, 40 s, ... and the add dies at 30 s, 60 s, ... and at the fight's end.
func TestDynamicAddsPresetEncounter(t *testing.T) {
	logs := runDynamicAdds(t, 180, nil)
	spawns := logTimes(logs, "Spawned 1 adds")
	deaths := logTimes(logs, "Died (kill credited")
	wantSpawns := []float64{10, 40, 70, 100, 130, 160}
	wantDeaths := []float64{30, 60, 90, 120, 150, 180}
	if !slices.Equal(spawns, wantSpawns) {
		t.Errorf("the add spawned at %v, want %v", spawns, wantSpawns)
	}
	if !slices.Equal(deaths, wantDeaths) {
		t.Errorf("the add died at %v, want 20 s after each spawn: %v", deaths, wantDeaths)
	}
}

// The boss's Kill share % input is the chance the raid gets each despawning add's kill: 0 never credits it, 100
// always does, and 20 credits some and not others.
func TestDynamicAddsKillShare(t *testing.T) {
	for _, tc := range []struct {
		percent       float64
		credited, not bool
	}{
		{0, false, true},
		{100, true, false},
		{20, true, true},
	} {
		logs := runDynamicAdds(t, 3000, &tc.percent)
		credited, not := len(logTimes(logs, "kill credited: true")), len(logTimes(logs, "kill credited: false"))
		if credited+not < 90 {
			t.Fatalf("kill share %v%%: %d deaths in a 3000 s fight, want about 100", tc.percent, credited+not)
		}
		if (credited > 0) != tc.credited || (not > 0) != tc.not {
			t.Errorf("kill share %v%%: %d deaths credited and %d not", tc.percent, credited, not)
		}
	}
}

// The seconds of the debug log lines that contain text.
func logTimes(logs string, text string) []float64 {
	var times []float64
	for line := range strings.SplitSeq(logs, "\n") {
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
