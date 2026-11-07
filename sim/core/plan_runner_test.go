package core

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

const planTestDuration = 100 * time.Second

func planTestTarget() *proto.Target {
	return &proto.Target{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}
}

func planTestRequest(numTargets int, plan *proto.EncounterPlan) *proto.RaidSimRequest {
	targets := make([]*proto.Target, numTargets)
	for i := range targets {
		targets[i] = planTestTarget()
	}
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name:      "Warrior",
					Class:     proto.Class_ClassWarrior,
					Race:      proto.Race_RaceHuman,
					Buffs:     &proto.IndividualBuffs{},
					Spec:      &proto.Player_ProtectionWarrior{},
					Equipment: &proto.EquipmentSpec{},
					Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
				}},
				Buffs: &proto.PartyBuffs{},
			}},
		},
		Encounter: &proto.Encounter{Targets: targets, Duration: planTestDuration.Seconds(), Plan: plan},
	}
}

// A sim of the fake Protection Warrior, which only auto attacks, reset and through the prepull. register runs
// while the environment is constructing, so it may register auras.
func setupPlanSim(request *proto.RaidSimRequest, register func(char *Character)) (*Simulation, *FakeFearWarrior) {
	sim := NewSim(request, simsignals.CreateSignals())
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

func planAt(fraction float64) time.Duration {
	return time.Duration(fraction * float64(planTestDuration))
}

func testAddWave(count int32, spawn, lifetime float64, repeat *proto.PlanRepeat) *proto.AddWave {
	adds := make([]*proto.Target, count)
	for i := range adds {
		adds[i] = planTestTarget()
	}
	return &proto.AddWave{
		Name:                 "adds",
		Adds:                 adds,
		Spawn:                &proto.PlanTime{At: spawn},
		Lifetime:             lifetime,
		Repeat:               repeat,
		ExecuteProportion_20: 0.25,
		ExecuteProportion_35: 0.5,
		KillShare:            1,
	}
}

// The times of every log line holding text, over one iteration.
func logTimes(sim *Simulation, text string) *[]time.Duration {
	times := &[]time.Duration{}
	sim.Log = func(message string, vals ...interface{}) {
		if strings.Contains(fmt.Sprintf(message, vals...), text) {
			*times = append(*times, sim.CurrentTime)
		}
	}
	return times
}

// Checks through Is Execute Phase on the target that it enters 35% at at35 and 20% at at20, fractions of the fight,
// each to the millisecond.
func checkExecuteCrossings(t *testing.T, sim *Simulation, rot *APLRotation, targetIndex int32, at35, at20 float64) {
	t.Helper()
	phase := func(threshold proto.APLValueIsExecutePhase_ExecutePhaseThreshold) APLValue {
		return planValue(t, rot, threshold.String(), &proto.APLValue{Value: &proto.APLValue_IsExecutePhase{
			IsExecutePhase: &proto.APLValueIsExecutePhase{Threshold: threshold, TargetUnit: planTargetRef(targetIndex)},
		}})
	}
	in35, in20 := phase(proto.APLValueIsExecutePhase_E35), phase(proto.APLValueIsExecutePhase_E20)
	for _, step := range []struct {
		at     time.Duration
		want35 bool
		want20 bool
	}{
		{planAt(at35) - time.Millisecond, false, false},
		{planAt(at35) + time.Millisecond, true, false},
		{planAt(at20) - time.Millisecond, true, false},
		{planAt(at20) + time.Millisecond, true, true},
	} {
		runLifecycleUntil(sim, step.at)
		if got35, got20 := in35.GetBool(sim), in20.GetBool(sim); got35 != step.want35 || got20 != step.want20 {
			t.Errorf("execute phases at %s: expected 35%% %v and 20%% %v, got %v and %v", step.at, step.want35, step.want20, got35, got20)
		}
	}
}

func TestPlanDowntimeTakesTheTargetAwayAndBringsItBack(t *testing.T) {
	plan := &proto.EncounterPlan{Downtimes: []*proto.PlanDowntime{
		{TargetIndex: 0, Start: &proto.PlanTime{At: 0.25}, End: &proto.PlanTime{At: 0.5}, ClearDebuffs: true},
		{TargetIndex: 1, Start: &proto.PlanTime{At: 0.25}, End: &proto.PlanTime{At: 0.5}},
	}}
	var cleared, kept *Aura
	sim, _ := setupPlanSim(planTestRequest(3, plan), func(char *Character) {
		// Timed, since clearing a target's debuffs leaves its permanent ones.
		cleared = char.Env.Encounter.AllTargets[0].RegisterAura(Aura{Label: "Cleared Debuff", Duration: 5 * time.Minute})
		kept = char.Env.Encounter.AllTargets[1].RegisterAura(Aura{Label: "Kept Debuff", Duration: 5 * time.Minute})
	})
	first, second := sim.Encounter.AllTargets[0], sim.Encounter.AllTargets[1]
	cleared.Activate(sim)
	kept.Activate(sim)

	runLifecycleUntil(sim, planAt(0.25)-time.Millisecond)
	if !first.IsEnabled() || !second.IsEnabled() {
		t.Fatal("before the downtime: expected both targets active")
	}

	runLifecycleUntil(sim, planAt(0.25))
	if first.IsEnabled() || second.IsEnabled() {
		t.Fatal("at the downtime's start: expected both targets gone")
	}
	if cleared.IsActive() {
		t.Error("a downtime that clears debuffs left the debuff on its target")
	}
	if !kept.IsActive() {
		t.Error("a downtime that keeps debuffs took the debuff off its target")
	}
	if got := first.NextActiveAt(sim); got != planAt(0.5) {
		t.Errorf("next active during the downtime: expected %s, got %s", planAt(0.5), got)
	}
	if got := first.DiesAt(sim); got != sim.Duration {
		t.Errorf("death of a target in a downtime: expected the fight's end %s, got %s", sim.Duration, got)
	}

	runLifecycleUntil(sim, planAt(0.5))
	if !first.IsEnabled() || !second.IsEnabled() {
		t.Fatal("at the downtime's end: expected both targets back")
	}
	if got := first.NextActiveAt(sim); got != sim.CurrentTime {
		t.Errorf("next active of an active target: expected now, got %s", got)
	}
}

func TestPlanBossKeyframesDriveHealthAndExecutePhases(t *testing.T) {
	plan := &proto.EncounterPlan{BossHealth: []*proto.HealthKeyframe{
		{At: 1, HealthPercent: 0},
		{At: 0, HealthPercent: 100},
		{At: 0.5, HealthPercent: 50},
		{At: 0.8, HealthPercent: 20},
	}}
	sim, _ := setupPlanSim(planTestRequest(1, plan), nil)
	boss := sim.Encounter.AllTargets[0]

	for _, step := range []struct {
		at    time.Duration
		phase int32
	}{
		{planAt(0.65) - time.Millisecond, 0},
		{planAt(0.65), 35},
		{planAt(0.8) - time.Millisecond, 35},
		{planAt(0.8), 20},
	} {
		runLifecycleUntil(sim, step.at)
		if boss.executePhase != step.phase {
			t.Errorf("execute phase at %s: expected %d, got %d", step.at, step.phase, boss.executePhase)
		}
	}
}

// A keyframe below 35% crosses it from full health at the pull, not at the keyframe, and one above it falls on to 0
// at the fight's end, when the boss dies.
func TestPlanBossCurveStartsFullAndEndsDead(t *testing.T) {
	for _, tc := range []struct {
		percent    float64
		at35, at20 float64
	}{{20, 0.40625, 0.5}, {40, 0.5625, 0.75}} {
		plan := &proto.EncounterPlan{BossHealth: []*proto.HealthKeyframe{{At: 0.5, HealthPercent: tc.percent}}}
		sim, fw := setupPlanSim(planTestRequest(1, plan), nil)
		boss := planTargetRef(0)
		active := planValue(t, fw.Rotation, "active", &proto.APLValue{Value: &proto.APLValue_TargetIsActive{TargetIsActive: &proto.APLValueTargetIsActive{TargetUnit: boss}}})
		untilDies := planValue(t, fw.Rotation, "untilDies", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetDies{TimeUntilTargetDies: &proto.APLValueTimeUntilTargetDies{TargetUnit: boss}}})
		if !active.GetBool(sim) {
			t.Errorf("keyframe at %v%%: expected the boss active at the pull", tc.percent)
		}
		if got := untilDies.GetDuration(sim); got != planTestDuration {
			t.Errorf("keyframe at %v%%: expected the boss to die at the fight's end %s, got %s", tc.percent, planTestDuration, got)
		}
		checkExecuteCrossings(t, sim, fw.Rotation, 0, tc.at35, tc.at20)
	}
}

func TestPlanBossCurveKeepsTheKeyframesThatFit(t *testing.T) {
	var keyframes []*proto.HealthKeyframe
	for i := range 20 {
		keyframes = append(keyframes, &proto.HealthKeyframe{At: 0.05 * float64(i+1), HealthPercent: 95 - 5*float64(i)})
	}
	sim, _ := setupPlanSim(planTestRequest(1, &proto.EncounterPlan{BossHealth: keyframes}), nil)
	sim.Cleanup()
	notes := logTimes(sim, "the first 16 are kept")
	sim.runOnce()
	if len(*notes) != 1 {
		t.Errorf("too many keyframes: expected one log line, got %d", len(*notes))
	}
	if got, want := sim.Encounter.AllTargets[0].health.executeAt35, planAt(0.65); got != want {
		t.Errorf("35%% crossing from the kept knots: expected %s, got %s", want, got)
	}
}

func TestPlanBossCurveKeepsTwelveKeyframes(t *testing.T) {
	var keyframes []*proto.HealthKeyframe
	for i := range 12 {
		keyframes = append(keyframes, &proto.HealthKeyframe{At: 0.08 * float64(i+1), HealthPercent: 100 - 8*float64(i+1)})
	}
	sim, fw := setupPlanSim(planTestRequest(1, &proto.EncounterPlan{BossHealth: keyframes}), nil)
	sim.Cleanup()
	notes := logTimes(sim, "are kept")
	sim.reset()
	sim.PrePull()
	if len(*notes) != 0 {
		t.Errorf("12 keyframes fit: expected no log line, got %d", len(*notes))
	}
	// 35% falls between the eighth and ninth keyframes, 20% on the tenth.
	checkExecuteCrossings(t, sim, fw.Rotation, 0, 0.65, 0.8)
}

func TestPlanWaveSpawnsAddsThatDieAtTheirDespawn(t *testing.T) {
	plan := &proto.EncounterPlan{AddWaves: []*proto.AddWave{testAddWave(2, 0.1, 0.2, nil)}}
	kills := 0
	var listener *Aura
	sim, _ := setupPlanSim(planTestRequest(1, plan), func(char *Character) {
		listener = registerDeathListener(char, "Death Listener", nil, func(_ *Aura, _ *Simulation, _ *Unit) { kills++ })
	})
	listener.Activate(sim)

	if got := len(sim.Encounter.AllTargets); got != 3 {
		t.Fatalf("targets: expected the boss and a pool of 2, got %d", got)
	}
	adds := sim.Encounter.AllTargets[1:]
	if sim.Encounter.NumActiveAdds() != 0 {
		t.Fatal("before the spawn: expected no adds")
	}
	if got := adds[0].NextActiveAt(sim); got != planAt(0.1) {
		t.Errorf("next active of an add before its spawn: expected %s, got %s", planAt(0.1), got)
	}

	runLifecycleUntil(sim, planAt(0.15))
	if got := sim.Encounter.NumActiveAdds(); got != 2 {
		t.Fatalf("after the spawn: expected 2 adds, got %d", got)
	}
	for _, add := range adds {
		if got := add.DiesAt(sim); got != planAt(0.3) {
			t.Errorf("add death: expected %s, got %s", planAt(0.3), got)
		}
	}

	for _, step := range []struct {
		at    time.Duration
		phase int32
	}{
		{planAt(0.2) - time.Millisecond, 0},
		{planAt(0.2), 35},
		{planAt(0.25) - time.Millisecond, 35},
		{planAt(0.25), 20},
	} {
		runLifecycleUntil(sim, step.at)
		for i, add := range adds {
			if add.executePhase != step.phase {
				t.Errorf("execute phase of add %d at %s: expected %d, got %d", i, step.at, step.phase, add.executePhase)
			}
		}
	}

	runLifecycleUntil(sim, planAt(0.3))
	if got := sim.Encounter.NumActiveAdds(); got != 0 {
		t.Fatalf("after the despawn: expected no adds, got %d", got)
	}
	if kills != 2 {
		t.Errorf("kills at a kill share of 1: expected 2, got %d", kills)
	}
	for _, add := range adds {
		if got := add.DiesAt(sim); got != NeverExpires {
			t.Errorf("death of a dead add with no spawn left: expected never, got %s", got)
		}
	}
}

func TestPlanWaveSpawnsEachAddFromItsOwnTemplate(t *testing.T) {
	wave := testAddWave(0, 0.1, 0.2, nil)
	for i, mobType := range []proto.MobType{proto.MobType_MobTypeBeast, proto.MobType_MobTypeDemon, proto.MobType_MobTypeUndead} {
		armor := make([]float64, int(stats.Armor)+1)
		armor[stats.Armor] = float64(1000 * (i + 1))
		wave.Adds = append(wave.Adds, &proto.Target{Name: fmt.Sprintf("add %d", i+1), Level: int32(60 + i), MobType: mobType, Stats: armor})
	}
	plan := &proto.EncounterPlan{AddWaves: []*proto.AddWave{wave}}
	kills := 0
	var listener *Aura
	sim, _ := setupPlanSim(planTestRequest(1, plan), func(char *Character) {
		listener = registerDeathListener(char, "Death Listener", nil, func(_ *Aura, _ *Simulation, _ *Unit) { kills++ })
	})
	listener.Activate(sim)

	if got := len(sim.Encounter.AllTargets); got != 1+len(wave.Adds) {
		t.Fatalf("targets: expected the boss and one slot per add, got %d", got)
	}
	adds := sim.Encounter.AllTargets[1:]
	for i, add := range adds {
		template := wave.Adds[i]
		if add.Level != template.Level || add.MobType != template.MobType || add.GetStat(stats.Armor) != template.Stats[stats.Armor] {
			t.Errorf("add %d: expected level %d, %s, armor %v from its template, got level %d, %s, armor %v",
				i+1, template.Level, template.MobType, template.Stats[stats.Armor], add.Level, add.MobType, add.GetStat(stats.Armor))
		}
	}

	runLifecycleUntil(sim, planAt(0.15))
	if got := sim.Encounter.NumActiveAdds(); got != len(adds) {
		t.Fatalf("after the spawn: expected %d adds, got %d", len(adds), got)
	}
	for i, add := range adds {
		if got := add.DiesAt(sim); got != planAt(0.3) {
			t.Errorf("add %d death: expected %s, got %s", i+1, planAt(0.3), got)
		}
	}

	runLifecycleUntil(sim, planAt(0.3))
	if got := sim.Encounter.NumActiveAdds(); got != 0 {
		t.Errorf("after the despawn: expected no adds, got %d", got)
	}
	if kills != len(adds) {
		t.Errorf("kills at a kill share of 1: expected %d, got %d", len(adds), kills)
	}
}

func TestPlanMetricsNameTheAddOfEachPoolTarget(t *testing.T) {
	wave := testAddWave(0, 0.1, 0.15, &proto.PlanRepeat{Interval: 0.1})
	wave.Adds = []*proto.Target{planTestTarget(), nil, planTestTarget()}
	encounter := NewEncounter(planTestRequest(1, &proto.EncounterPlan{AddWaves: []*proto.AddWave{wave}}).Encounter)

	targets := encounter.GetMetricsProto().Targets
	if len(targets) != 5 {
		t.Fatalf("targets: expected the boss and two slots for each add with a target, got %d", len(targets))
	}
	if targets[0].PlanAdd != nil {
		t.Errorf("the boss: expected no add of the plan, got %v", targets[0].PlanAdd)
	}
	for i, wantAdd := range []int32{0, 0, 2, 2} {
		if got := targets[i+1].PlanAdd; got == nil || got.WaveIndex != 0 || got.AddIndex != wantAdd {
			t.Errorf("pool target %d: expected wave 0 add %d, got %v", i+1, wantAdd, got)
		}
	}
}

func TestPlanWaveRepeatsCountTimesOrUntilTheEnd(t *testing.T) {
	for _, tc := range []struct {
		count int32
		want  []time.Duration
	}{
		{2, []time.Duration{planAt(0.1), planAt(0.3), planAt(0.5)}},
		{0, []time.Duration{planAt(0.1), planAt(0.3), planAt(0.5), planAt(0.7), planAt(0.9)}},
	} {
		plan := &proto.EncounterPlan{AddWaves: []*proto.AddWave{
			testAddWave(1, 0.1, 0.05, &proto.PlanRepeat{Interval: 0.2, Count: tc.count}),
		}}
		sim, _ := setupPlanSim(planTestRequest(1, plan), nil)
		sim.Cleanup()
		spawns := logTimes(sim, "spawned")
		sim.runOnce()
		if !slices.Equal(*spawns, tc.want) {
			t.Errorf("count %d: expected spawns at %v, got %v", tc.count, tc.want, *spawns)
		}
	}
}

func TestPlanJitterIsDeterministicAndIndependentOfGear(t *testing.T) {
	plan := &proto.EncounterPlan{
		Downtimes: []*proto.PlanDowntime{
			{TargetIndex: 0, Start: &proto.PlanTime{At: 0.4, Jitter: 0.05}, End: &proto.PlanTime{At: 0.5, Jitter: 0.05}},
		},
		AddWaves: []*proto.AddWave{
			testAddWave(1, 0.1, 0.1, &proto.PlanRepeat{Interval: 0.2, Jitter: 0.05}),
		},
	}
	plan.AddWaves[0].Spawn.Jitter = 0.05
	plan.AddWaves[0].KillShare = 0.5

	type iteration struct {
		spawns, downtimes []time.Duration
		kills             int
	}
	run := func(bonus stats.Stats) ([]iteration, float64) {
		request := planTestRequest(1, plan)
		request.Raid.Parties[0].Players[0].BonusStats = &proto.UnitStats{Stats: bonus.ToProtoArray()}
		kills := 0
		var listener *Aura
		sim, fw := setupPlanSim(request, func(char *Character) {
			listener = registerDeathListener(char, "Death Listener", nil, func(_ *Aura, _ *Simulation, _ *Unit) { kills++ })
		})
		sim.Cleanup()
		var iterations []iteration
		for i := range 20 {
			sim.reseedRands(int64(i))
			spawns := logTimes(sim, "spawned")
			logged := sim.Log
			var downtimes []time.Duration
			sim.Log = func(message string, vals ...interface{}) {
				if strings.Contains(message, "Downtime starts") {
					downtimes = append(downtimes, sim.CurrentTime)
				}
				logged(message, vals...)
			}
			kills = 0
			sim.reset()
			listener.Activate(sim)
			sim.PrePull()
			sim.runPendingActions()
			sim.Cleanup()
			iterations = append(iterations, iteration{spawns: *spawns, downtimes: downtimes, kills: kills})
		}
		return iterations, fw.Metrics.ToProto().Dps.Avg
	}

	base, baseDamage := run(stats.Stats{})
	again, _ := run(stats.Stats{})
	geared, gearedDamage := run(stats.Stats{stats.AttackPower: 1000, stats.HasteRating: 500})
	if baseDamage == gearedDamage {
		t.Fatal("test setup: the bonus stats deal the same damage, so the gear change moves nothing")
	}

	moved := false
	for i := range base {
		if !slices.Equal(base[i].spawns, again[i].spawns) || !slices.Equal(base[i].downtimes, again[i].downtimes) || base[i].kills != again[i].kills {
			t.Errorf("iteration %d: the same seed rolled a different timeline", i)
		}
		if !slices.Equal(base[i].spawns, geared[i].spawns) || !slices.Equal(base[i].downtimes, geared[i].downtimes) || base[i].kills != geared[i].kills {
			t.Errorf("iteration %d: other stats rolled a different timeline: %v / %v kills %d, against %v / %v kills %d",
				i, base[i].spawns, base[i].downtimes, base[i].kills, geared[i].spawns, geared[i].downtimes, geared[i].kills)
		}
		if i > 0 && !slices.Equal(base[i].spawns, base[0].spawns) {
			moved = true
		}
	}
	if !moved {
		t.Error("jitter: expected the spawns to move between iterations")
	}
}

func TestPlanPoolExhaustionSkipsTheSpawn(t *testing.T) {
	plan := &proto.EncounterPlan{AddWaves: []*proto.AddWave{
		testAddWave(1, 0, 1, &proto.PlanRepeat{Interval: 0.01}),
		{Name: "no adds", Lifetime: 0.1},
		{Name: "no template", Adds: []*proto.Target{nil}, Lifetime: 0.1},
		testAddWave(planMaxPoolTargets+1, 0.5, 0.1, nil),
		testAddWave(1, 0.5, 0, nil),
	}}
	sim, _ := setupPlanSim(planTestRequest(1, plan), nil)
	if got := len(sim.Encounter.AllTargets) - 1; got != planMaxPoolTargets {
		t.Fatalf("pool: expected the cap of %d targets, got %d", planMaxPoolTargets, got)
	}
	sim.Cleanup()
	skipped := logTimes(sim, "spawn skipped")
	skips := map[string]int{"spawns no adds": 0, "has no target": 0, "more than the": 0, "has no lifetime": 0}
	logged := sim.Log
	sim.Log = func(message string, vals ...interface{}) {
		line := fmt.Sprintf(message, vals...)
		for text := range skips {
			if strings.Contains(line, text) {
				skips[text]++
			}
		}
		logged(message, vals...)
	}
	sim.runOnce()
	if len(*skipped) == 0 {
		t.Error("pool exhaustion: expected a log line for each skipped spawn")
	}
	for text, lines := range skips {
		if lines != 1 {
			t.Errorf("a skipped wave that %q: expected one log line, got %d", text, lines)
		}
	}
	if got := sim.Encounter.NumActiveAdds(); got != planMaxPoolTargets {
		t.Errorf("adds at the end: expected the whole pool, got %d", got)
	}
}

func downtime(index int32, start, end float64, clear bool) *proto.PlanDowntime {
	return &proto.PlanDowntime{TargetIndex: index, Start: &proto.PlanTime{At: start}, End: &proto.PlanTime{At: end}, ClearDebuffs: clear}
}

func TestPlanOverlappingDowntimesAreAUnion(t *testing.T) {
	plan := &proto.EncounterPlan{Downtimes: []*proto.PlanDowntime{
		downtime(0, 0.2, 0.6, false),
		downtime(0, 0.3, 0.5, true),
	}}
	var debuff *Aura
	sim, _ := setupPlanSim(planTestRequest(2, plan), func(char *Character) {
		debuff = char.Env.Encounter.AllTargets[0].RegisterAura(Aura{Label: "Debuff", Duration: 5 * time.Minute})
	})
	target := sim.Encounter.AllTargets[0]
	debuff.Activate(sim)

	runLifecycleUntil(sim, planAt(0.25))
	if target.IsEnabled() || !debuff.IsActive() {
		t.Fatal("in the first downtime, which keeps debuffs: expected the target gone with its debuff")
	}
	if got := target.NextActiveAt(sim); got != planAt(0.6) {
		t.Errorf("next active in the first downtime: expected %s, got %s", planAt(0.6), got)
	}
	runLifecycleUntil(sim, planAt(0.35))
	if debuff.IsActive() {
		t.Error("a second downtime that clears debuffs on a target already down: expected the debuff gone")
	}
	if got := target.NextActiveAt(sim); got != planAt(0.6) {
		t.Errorf("next active in both downtimes: expected the latest end %s, got %s", planAt(0.6), got)
	}
	runLifecycleUntil(sim, planAt(0.55))
	if target.IsEnabled() {
		t.Fatal("after the inner downtime ends: expected the target still down for the outer one")
	}
	runLifecycleUntil(sim, planAt(0.6))
	if !target.IsEnabled() {
		t.Fatal("after the last downtime ends: expected the target back")
	}
}

func TestPlanBackToBackJitteredDowntimesComeBackOnce(t *testing.T) {
	plan := &proto.EncounterPlan{Downtimes: []*proto.PlanDowntime{
		downtime(0, 0.2, 0.4, false),
		downtime(0, 0.4, 0.6, false),
	}}
	plan.Downtimes[0].End.Jitter = 0.05
	plan.Downtimes[1].Start.Jitter = 0.05
	sim, _ := setupPlanSim(planTestRequest(2, plan), nil)
	target := sim.Encounter.AllTargets[0]
	overlapped, apart := 0, 0
	for i := range 30 {
		sim.Cleanup()
		sim.reseedRands(int64(i))
		backs := logTimes(sim, "target back: true")
		sim.reset()
		sim.PrePull()
		first, second := sim.Encounter.plan.downtimes[0], sim.Encounter.plan.downtimes[1]
		joined := first.end >= second.start

		probes := []time.Duration{planAt(0.1), (first.start + first.end) / 2, (second.start + second.end) / 2, planAt(0.7)}
		want := []bool{true, false, false, true}
		if !joined {
			probes = append(probes, (first.end+second.start)/2)
			want = append(want, true)
			slices.SortFunc(probes, func(a, b time.Duration) int { return int(a - b) })
			want = []bool{true, false, true, false, true}
		}
		for j, at := range probes {
			runLifecycleUntil(sim, at)
			if target.IsEnabled() != want[j] {
				t.Errorf("iteration %d at %s: expected active %t, got %t", i, at, want[j], target.IsEnabled())
			}
		}
		wantBacks := 2
		if joined {
			wantBacks = 1
			overlapped++
		} else {
			apart++
		}
		if len(*backs) != wantBacks {
			t.Errorf("iteration %d: expected the target back %d times, got %d", i, wantBacks, len(*backs))
		}
	}
	if overlapped == 0 || apart == 0 {
		t.Errorf("test setup: expected both joined and apart downtimes, got %d and %d", overlapped, apart)
	}
}

func TestPlanDowntimeNeverTurnsOnATargetDisabledAtStart(t *testing.T) {
	request := planTestRequest(2, &proto.EncounterPlan{Downtimes: []*proto.PlanDowntime{downtime(1, 0.2, 0.4, false)}})
	request.Encounter.Targets[1].DisabledAtStart = true
	sim, _ := setupPlanSim(request, nil)
	target := sim.Encounter.AllTargets[1]

	runLifecycleUntil(sim, planAt(0.3))
	if got := target.NextActiveAt(sim); got != NeverExpires {
		t.Errorf("next active of a target off before its downtime: expected never, got %s", got)
	}
	runLifecycleUntil(sim, planAt(0.5))
	if target.IsEnabled() {
		t.Error("after the downtime: expected the target disabled at start to stay off")
	}
}

func TestPlanBossKeyframesAreClampedAndNeverRise(t *testing.T) {
	plan := &proto.EncounterPlan{BossHealth: []*proto.HealthKeyframe{
		{At: -0.1, HealthPercent: 120}, {At: 0.5, HealthPercent: 60}, {At: 0.7, HealthPercent: 80}, {At: 1.2, HealthPercent: 0},
	}}
	sim, fw := setupPlanSim(planTestRequest(1, plan), nil)
	sim.Cleanup()
	notes := logTimes(sim, "are clamped")
	sim.reset()
	sim.PrePull()
	if len(*notes) != 1 {
		t.Errorf("clamped keyframes: expected one log line, got %d", len(*notes))
	}
	// Clamped, the boss holds 60% from 0.5 to 0.7 and falls to 0 at the fight's end.
	checkExecuteCrossings(t, sim, fw.Rotation, 0, 0.825, 0.9)
}

func TestPlanResetBringsBackATargetTheLastIterationEndedWithout(t *testing.T) {
	plan := &proto.EncounterPlan{Downtimes: []*proto.PlanDowntime{
		{TargetIndex: 0, Start: &proto.PlanTime{At: 0.5}, End: &proto.PlanTime{At: 1}},
	}}
	request := planTestRequest(1, plan)
	request.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	sim, fw := setupPlanSim(request, nil)
	boss := sim.Encounter.AllTargets[0]
	sim.runPendingActions()
	if boss.IsEnabled() {
		t.Fatal("test setup: expected the boss gone at the end of the first iteration")
	}
	sim.Cleanup()

	sim.reseedRands(1)
	sim.reset()
	sim.PrePull()
	if !boss.IsEnabled() || sim.ActiveTargetCount() != 1 {
		t.Fatal("after the reset: expected the boss active again")
	}
	if boss.CurrentTarget != &fw.Unit {
		t.Fatalf("boss's target after the reset: expected the tank, got %s", unitLabel(boss.CurrentTarget))
	}
	if fw.CurrentTarget != &boss.Unit {
		t.Fatalf("tank's target after the reset: expected the boss, got %s", unitLabel(fw.CurrentTarget))
	}
}

func TestPlanDowntimeOnTheOnlyTargetIdlesThePlayer(t *testing.T) {
	plan := &proto.EncounterPlan{Downtimes: []*proto.PlanDowntime{
		{TargetIndex: 0, Start: &proto.PlanTime{At: 0}, End: &proto.PlanTime{At: 1}},
	}}
	sim, fw := setupPlanSim(planTestRequest(1, plan), nil)
	target := sim.Encounter.AllTargets[0]

	runLifecycleUntil(sim, time.Second)
	if sim.ActiveTargetCount() != 0 {
		t.Fatalf("active targets: expected none, got %d", sim.ActiveTargetCount())
	}
	if aa := &fw.AutoAttacks; aa.mh.enabled || aa.oh.enabled {
		t.Fatal("auto attacks: expected them stopped with no target")
	}
	swings := mhSwingsOn(fw, target)
	sim.runPendingActions()
	if got := mhSwingsOn(fw, target); got != swings {
		t.Errorf("MH swings with no target: expected %d, got %d", swings, got)
	}
	if target.IsEnabled() {
		t.Error("a downtime to the fight's end: expected the target to stay gone")
	}
}

func TestPlanOnlyTargetDisabledAtStartIdlesThePlayerUntilATargetComes(t *testing.T) {
	for _, tc := range []struct {
		name    string
		waves   []*proto.AddWave
		bringUp func(sim *Simulation) *Target
	}{
		{"an add spawns", []*proto.AddWave{testAddWave(1, 0.1, 0.5, nil)}, func(sim *Simulation) *Target { return sim.Encounter.AllTargets[1] }},
		{"the target is enabled", nil, func(sim *Simulation) *Target {
			sim.Encounter.AllTargets[0].Enable(sim)
			return sim.Encounter.AllTargets[0]
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := planTestRequest(1, &proto.EncounterPlan{AddWaves: tc.waves})
			request.Encounter.Targets[0].DisabledAtStart = true
			sim, fw := setupPlanSim(request, nil)
			boss := sim.Encounter.AllTargets[0]

			runLifecycleUntil(sim, planAt(0.1)-time.Millisecond)
			if aa := &fw.AutoAttacks; aa.mh.enabled || aa.oh.enabled {
				t.Error("auto attacks before a target comes: expected them stopped")
			}
			if got := mhSwingsOn(fw, boss); got != 0 {
				t.Errorf("MH swings at the target disabled at start: expected none, got %d", got)
			}

			runLifecycleUntil(sim, planAt(0.1))
			target := tc.bringUp(sim)
			runLifecycleUntil(sim, planAt(0.15))
			if fw.CurrentTarget != &target.Unit {
				t.Errorf("player's target once a target comes: expected %s, got %s", target.Label, unitLabel(fw.CurrentTarget))
			}
			if got := mhSwingsOn(fw, target); got == 0 {
				t.Error("MH swings once a target comes: expected some, got none")
			}
		})
	}
}

func TestPlanRaidStartsOnTheLowestTargetNotDisabledAtStart(t *testing.T) {
	request := planTestRequest(3, &proto.EncounterPlan{Downtimes: []*proto.PlanDowntime{downtime(2, 0.5, 0.6, false)}})
	request.Encounter.Targets[0].DisabledAtStart = true
	sim, fw := setupPlanSim(request, nil)
	first := sim.Encounter.AllTargets[1]

	runLifecycleUntil(sim, planAt(0.05))
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
