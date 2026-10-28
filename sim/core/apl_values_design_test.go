package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func designTargetRef(index int32) *proto.UnitReference {
	return &proto.UnitReference{Type: proto.UnitReference_Target, Index: index}
}

func designValue(t *testing.T, rot *APLRotation, name string, config *proto.APLValue) APLValue {
	t.Helper()
	config.Uuid = &proto.UUID{Value: name}
	got := rot.newAPLValue(config)
	if got == nil {
		t.Fatalf("%s: got no value", name)
	}
	return got
}

type designValues struct {
	active0, active1                      APLValue
	adds                                  APLValue
	untilActive0, untilActive1            APLValue
	untilDies0, untilDies1                APLValue
	untilMove, moveDuration               APLValue
	ampActive, untilAmp                   APLValue
	currentPercent0, currentHealth0       APLValue
	currentPercentAdd1, currentHealthAdd1 APLValue
}

func buildDesignValues(t *testing.T, rot *APLRotation, addIndex int32) *designValues {
	t.Helper()
	boss, add := designTargetRef(0), designTargetRef(addIndex)
	amp := proto.DamageAmpType_DamageAmpTypeEnvironmentBuff
	return &designValues{
		active0:            designValue(t, rot, "active0", &proto.APLValue{Value: &proto.APLValue_TargetIsActive{TargetIsActive: &proto.APLValueTargetIsActive{TargetUnit: boss}}}),
		active1:            designValue(t, rot, "active1", &proto.APLValue{Value: &proto.APLValue_TargetIsActive{TargetIsActive: &proto.APLValueTargetIsActive{TargetUnit: add}}}),
		adds:               designValue(t, rot, "adds", &proto.APLValue{Value: &proto.APLValue_NumActiveAdds{NumActiveAdds: &proto.APLValueNumActiveAdds{}}}),
		untilActive0:       designValue(t, rot, "untilActive0", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetActive{TimeUntilTargetActive: &proto.APLValueTimeUntilTargetActive{TargetUnit: boss}}}),
		untilActive1:       designValue(t, rot, "untilActive1", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetActive{TimeUntilTargetActive: &proto.APLValueTimeUntilTargetActive{TargetUnit: add}}}),
		untilDies0:         designValue(t, rot, "untilDies0", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetDies{TimeUntilTargetDies: &proto.APLValueTimeUntilTargetDies{TargetUnit: boss}}}),
		untilDies1:         designValue(t, rot, "untilDies1", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetDies{TimeUntilTargetDies: &proto.APLValueTimeUntilTargetDies{TargetUnit: add}}}),
		untilMove:          designValue(t, rot, "untilMove", &proto.APLValue{Value: &proto.APLValue_TimeUntilNextMove{TimeUntilNextMove: &proto.APLValueTimeUntilNextMove{}}}),
		moveDuration:       designValue(t, rot, "moveDuration", &proto.APLValue{Value: &proto.APLValue_NextMoveDuration{NextMoveDuration: &proto.APLValueNextMoveDuration{}}}),
		ampActive:          designValue(t, rot, "ampActive", &proto.APLValue{Value: &proto.APLValue_AmpIsActive{AmpIsActive: &proto.APLValueAmpIsActive{AmpType: amp}}}),
		untilAmp:           designValue(t, rot, "untilAmp", &proto.APLValue{Value: &proto.APLValue_TimeUntilAmp{TimeUntilAmp: &proto.APLValueTimeUntilAmp{AmpType: amp}}}),
		currentPercent0:    designValue(t, rot, "percent0", &proto.APLValue{Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{SourceUnit: boss}}}),
		currentHealth0:     designValue(t, rot, "health0", &proto.APLValue{Value: &proto.APLValue_CurrentHealth{CurrentHealth: &proto.APLValueCurrentHealth{SourceUnit: boss}}}),
		currentPercentAdd1: designValue(t, rot, "percentAdd", &proto.APLValue{Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{SourceUnit: add}}}),
		currentHealthAdd1:  designValue(t, rot, "healthAdd", &proto.APLValue{Value: &proto.APLValue_CurrentHealth{CurrentHealth: &proto.APLValueCurrentHealth{SourceUnit: add}}}),
	}
}

const designTestHealth = 50000

func designHealthStats() []float64 {
	targetStats := make([]float64, int(stats.Health)+1)
	targetStats[stats.Health] = designTestHealth
	return targetStats
}

// A request whose targets and wave templates carry a Health stat.
func designHealthRequest(numTargets int, design *proto.EncounterDesign) *proto.RaidSimRequest {
	request := designTestRequest(numTargets, design)
	for _, target := range request.Encounter.Targets {
		target.Stats = designHealthStats()
	}
	for _, wave := range design.GetAddWaves() {
		wave.Target.Stats = designHealthStats()
	}
	return request
}

func TestValueDesignPlainEncounter(t *testing.T) {
	sim, fw := setupDesignSim(designHealthRequest(2, nil), nil)
	values := buildDesignValues(t, fw.Rotation, 1)

	if !values.active0.GetBool(sim) || !values.active1.GetBool(sim) {
		t.Error("plain targets: expected both active")
	}
	if got := values.adds.GetInt(sim); got != 1 {
		t.Errorf("active adds with a second plain target: expected 1, got %d", got)
	}
	for name, value := range map[string]APLValue{"time until active 0": values.untilActive0, "time until active 1": values.untilActive1} {
		if got := value.GetDuration(sim); got != 0 {
			t.Errorf("%s of an active target: expected 0, got %s", name, got)
		}
	}
	runLifecycleUntil(sim, designAt(0.25))
	for name, value := range map[string]APLValue{"time until dies 0": values.untilDies0, "time until dies 1": values.untilDies1} {
		if got := value.GetDuration(sim); got != designAt(0.75) {
			t.Errorf("%s: expected the fight's remainder %s, got %s", name, designAt(0.75), got)
		}
	}
	if got := values.untilMove.GetDuration(sim); got != NeverExpires {
		t.Errorf("time until next move: expected never, got %s", got)
	}
	if got := values.moveDuration.GetDuration(sim); got != 0 {
		t.Errorf("next move duration: expected 0, got %s", got)
	}
	if values.ampActive.GetBool(sim) {
		t.Error("amp is active: expected false")
	}
	if got := values.untilAmp.GetDuration(sim); got != NeverExpires {
		t.Errorf("time until amp: expected never, got %s", got)
	}
	percent := sim.Encounter.AllTargets[0].HealthPercent(sim) / 100
	if got := values.currentPercent0.GetFloat(sim); got != percent || percent >= 1 {
		t.Errorf("enemy health %%: expected the curve's %v below 1, got %v", percent, got)
	}
	if got, want := values.currentHealth0.GetFloat(sim), percent*designTestHealth; got != want || want == 0 {
		t.Errorf("enemy health: expected %v, got %v", want, got)
	}
}

func TestValueDesignReadsDowntimesAndWaves(t *testing.T) {
	design := &proto.EncounterDesign{
		Downtimes: []*proto.DesignDowntime{{TargetIndex: 0, Start: &proto.DesignTime{At: 0.25}, End: &proto.DesignTime{At: 0.5}}},
		AddWaves:  []*proto.AddWave{testAddWave(1, 0.6, 0.2, nil)},
	}
	sim, fw := setupDesignSim(designHealthRequest(1, design), nil)
	values := buildDesignValues(t, fw.Rotation, 1)

	type check struct {
		name string
		got  any
		want any
	}
	verify := func(label string, checks ...check) {
		t.Helper()
		for _, c := range checks {
			if c.got != c.want {
				t.Errorf("%s, %s: expected %v, got %v", label, c.name, c.want, c.got)
			}
		}
	}

	runLifecycleUntil(sim, designAt(0.1))
	verify("before the downtime",
		check{"boss active", values.active0.GetBool(sim), true},
		check{"add active", values.active1.GetBool(sim), false},
		check{"adds", values.adds.GetInt(sim), int32(0)},
		check{"boss until active", values.untilActive0.GetDuration(sim), time.Duration(0)},
		check{"add until active", values.untilActive1.GetDuration(sim), designAt(0.5)},
		check{"boss until dies", values.untilDies0.GetDuration(sim), designAt(0.9)},
		check{"add until dies", values.untilDies1.GetDuration(sim), designAt(0.7)},
		check{"add health %", values.currentPercentAdd1.GetFloat(sim), 1.0},
	)

	runLifecycleUntil(sim, designAt(0.3))
	verify("inside the downtime",
		check{"boss active", values.active0.GetBool(sim), false},
		check{"boss until active", values.untilActive0.GetDuration(sim), designAt(0.2)},
		check{"boss until dies", values.untilDies0.GetDuration(sim), designAt(0.7)},
	)

	runLifecycleUntil(sim, designAt(0.7))
	addHealthPercent := sim.Encounter.AllTargets[1].HealthPercent(sim) / 100
	verify("inside the wave",
		check{"boss active", values.active0.GetBool(sim), true},
		check{"add active", values.active1.GetBool(sim), true},
		check{"adds", values.adds.GetInt(sim), int32(1)},
		check{"add until active", values.untilActive1.GetDuration(sim), time.Duration(0)},
		check{"add until dies", values.untilDies1.GetDuration(sim), designAt(0.1)},
		check{"add health %", values.currentPercentAdd1.GetFloat(sim), addHealthPercent},
		check{"add health", values.currentHealthAdd1.GetFloat(sim), addHealthPercent * designTestHealth},
	)

	runLifecycleUntil(sim, designAt(0.85))
	verify("after the wave",
		check{"add active", values.active1.GetBool(sim), false},
		check{"adds", values.adds.GetInt(sim), int32(0)},
		check{"add until active", values.untilActive1.GetDuration(sim), NeverExpires},
		check{"add until dies", values.untilDies1.GetDuration(sim), NeverExpires},
	)
}

func TestValueDesignWarnsForUnitsThatAreNotEnemyTargets(t *testing.T) {
	_, fw := setupDesignSim(designTestRequest(1, nil), nil)
	self := &proto.UnitReference{Type: proto.UnitReference_Self}
	configs := map[string]*proto.APLValue{
		"Target Is Active":         {Uuid: &proto.UUID{Value: "a"}, Value: &proto.APLValue_TargetIsActive{TargetIsActive: &proto.APLValueTargetIsActive{TargetUnit: self}}},
		"Time Until Target Active": {Uuid: &proto.UUID{Value: "b"}, Value: &proto.APLValue_TimeUntilTargetActive{TimeUntilTargetActive: &proto.APLValueTimeUntilTargetActive{TargetUnit: self}}},
		"Time Until Target Dies":   {Uuid: &proto.UUID{Value: "c"}, Value: &proto.APLValue_TimeUntilTargetDies{TimeUntilTargetDies: &proto.APLValueTimeUntilTargetDies{TargetUnit: self}}},
	}
	for name, config := range configs {
		if value := fw.Rotation.newAPLValue(config); value != nil {
			t.Errorf("%s: got a value for the player, want none", name)
		}
		if warnings := fw.Rotation.uuidValidations[config.Uuid]; len(warnings) != 1 || warnings[0].LogLevel != proto.LogLevel_Warning {
			t.Errorf("%s: validations = %v, want one warning", name, warnings)
		}
	}
}

func TestValueDesignWarnsForEnemyWithoutHealthAndUnknownAmp(t *testing.T) {
	_, fw := setupDesignSim(designTestRequest(1, nil), nil)
	boss := designTargetRef(0)
	configs := map[string]*proto.APLValue{
		"Current Health":   {Uuid: &proto.UUID{Value: "a"}, Value: &proto.APLValue_CurrentHealth{CurrentHealth: &proto.APLValueCurrentHealth{SourceUnit: boss}}},
		"Current Health %": {Uuid: &proto.UUID{Value: "b"}, Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{SourceUnit: boss}}},
	}
	for name, config := range configs {
		if value := fw.Rotation.newAPLValue(config); value == nil {
			t.Errorf("%s: expected the value to be kept", name)
		}
		if warnings := fw.Rotation.uuidValidations[config.Uuid]; len(warnings) != 1 || warnings[0].LogLevel != proto.LogLevel_Warning {
			t.Errorf("%s: validations = %v, want one warning", name, warnings)
		}
	}

	amps := map[string]*proto.APLValue{
		"Amp Is Active":  {Uuid: &proto.UUID{Value: "c"}, Value: &proto.APLValue_AmpIsActive{AmpIsActive: &proto.APLValueAmpIsActive{AmpType: 7}}},
		"Time Until Amp": {Uuid: &proto.UUID{Value: "d"}, Value: &proto.APLValue_TimeUntilAmp{TimeUntilAmp: &proto.APLValueTimeUntilAmp{AmpType: -1}}},
	}
	for name, config := range amps {
		if value := fw.Rotation.newAPLValue(config); value != nil {
			t.Errorf("%s: got a value for an unknown amp type, want none", name)
		}
		if warnings := fw.Rotation.uuidValidations[config.Uuid]; len(warnings) != 1 || warnings[0].LogLevel != proto.LogLevel_Warning {
			t.Errorf("%s: validations = %v, want one warning", name, warnings)
		}
	}
}

func TestValueDesignFallsBackWhenTheCurrentTargetIsNotAnEnemy(t *testing.T) {
	sim, fw := setupDesignSim(designHealthRequest(1, nil), nil)
	rot := fw.Rotation
	current := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}
	active := designValue(t, rot, "active", &proto.APLValue{Value: &proto.APLValue_TargetIsActive{TargetIsActive: &proto.APLValueTargetIsActive{TargetUnit: current}}})
	untilActive := designValue(t, rot, "untilActive", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetActive{TimeUntilTargetActive: &proto.APLValueTimeUntilTargetActive{TargetUnit: current}}})
	untilDies := designValue(t, rot, "untilDies", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetDies{TimeUntilTargetDies: &proto.APLValueTimeUntilTargetDies{TargetUnit: current}}})
	health := designValue(t, rot, "health", &proto.APLValue{Value: &proto.APLValue_CurrentHealth{CurrentHealth: &proto.APLValueCurrentHealth{SourceUnit: current}}})
	percent := designValue(t, rot, "percent", &proto.APLValue{Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{SourceUnit: current}}})

	fw.CurrentTarget = &fw.Unit
	if active.GetBool(sim) {
		t.Error("target is active for a non-enemy current target: expected false")
	}
	if got := untilActive.GetDuration(sim); got != NeverExpires {
		t.Errorf("time until active for a non-enemy: expected never, got %s", got)
	}
	if got := untilDies.GetDuration(sim); got != NeverExpires {
		t.Errorf("time until dies for a non-enemy: expected never, got %s", got)
	}
	if got := health.GetFloat(sim); got != fw.CurrentHealth() {
		t.Errorf("health of a non-enemy: expected %v, got %v", fw.CurrentHealth(), got)
	}
	if got := percent.GetFloat(sim); got != fw.CurrentHealthPercent() {
		t.Errorf("health %% of a non-enemy: expected %v, got %v", fw.CurrentHealthPercent(), got)
	}
}
