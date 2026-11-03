package encounters

import (
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
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: &proto.Encounter{
			Duration: 180,
			Targets: []*proto.Target{
				googleproto.Clone(core.GetPresetTargetWithID(dynamicBossID).Config).(*proto.Target),
				googleproto.Clone(core.GetPresetTargetWithID(dynamicAddID).Config).(*proto.Target),
			},
		},
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
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
