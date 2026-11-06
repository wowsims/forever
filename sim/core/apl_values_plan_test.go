package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

func planTargetRef(index int32) *proto.UnitReference {
	return &proto.UnitReference{Type: proto.UnitReference_Target, Index: index}
}

func planValue(t *testing.T, rot *APLRotation, name string, config *proto.APLValue) APLValue {
	t.Helper()
	config.Uuid = &proto.UUID{Value: name}
	got := rot.newAPLValue(config)
	if got == nil {
		t.Fatalf("%s: got no value", name)
	}
	return got
}

type planValues struct {
	active0, active1           APLValue
	adds                       APLValue
	untilActive0, untilActive1 APLValue
	untilDies0, untilDies1     APLValue
	untilMove, moveDuration    APLValue
	ampActive, untilAmp        APLValue
}

func buildPlanValues(t *testing.T, rot *APLRotation, addIndex int32) *planValues {
	t.Helper()
	boss, add := planTargetRef(0), planTargetRef(addIndex)
	amp := proto.DamageAmpType_DamageAmpTypeEnvironmentBuff
	return &planValues{
		active0:      planValue(t, rot, "active0", &proto.APLValue{Value: &proto.APLValue_TargetIsActive{TargetIsActive: &proto.APLValueTargetIsActive{TargetUnit: boss}}}),
		active1:      planValue(t, rot, "active1", &proto.APLValue{Value: &proto.APLValue_TargetIsActive{TargetIsActive: &proto.APLValueTargetIsActive{TargetUnit: add}}}),
		adds:         planValue(t, rot, "adds", &proto.APLValue{Value: &proto.APLValue_NumActiveAdds{NumActiveAdds: &proto.APLValueNumActiveAdds{}}}),
		untilActive0: planValue(t, rot, "untilActive0", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetActive{TimeUntilTargetActive: &proto.APLValueTimeUntilTargetActive{TargetUnit: boss}}}),
		untilActive1: planValue(t, rot, "untilActive1", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetActive{TimeUntilTargetActive: &proto.APLValueTimeUntilTargetActive{TargetUnit: add}}}),
		untilDies0:   planValue(t, rot, "untilDies0", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetDies{TimeUntilTargetDies: &proto.APLValueTimeUntilTargetDies{TargetUnit: boss}}}),
		untilDies1:   planValue(t, rot, "untilDies1", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetDies{TimeUntilTargetDies: &proto.APLValueTimeUntilTargetDies{TargetUnit: add}}}),
		untilMove:    planValue(t, rot, "untilMove", &proto.APLValue{Value: &proto.APLValue_TimeUntilNextMove{TimeUntilNextMove: &proto.APLValueTimeUntilNextMove{}}}),
		moveDuration: planValue(t, rot, "moveDuration", &proto.APLValue{Value: &proto.APLValue_NextMoveDuration{NextMoveDuration: &proto.APLValueNextMoveDuration{}}}),
		ampActive:    planValue(t, rot, "ampActive", &proto.APLValue{Value: &proto.APLValue_AmpIsActive{AmpIsActive: &proto.APLValueAmpIsActive{AmpType: amp}}}),
		untilAmp:     planValue(t, rot, "untilAmp", &proto.APLValue{Value: &proto.APLValue_TimeUntilAmp{TimeUntilAmp: &proto.APLValueTimeUntilAmp{AmpType: amp}}}),
	}
}

func TestValuePlanPlainEncounter(t *testing.T) {
	sim, fw := setupPlanSim(planTestRequest(2, nil), nil)
	values := buildPlanValues(t, fw.Rotation, 1)

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
	runLifecycleUntil(sim, planAt(0.25))
	for name, value := range map[string]APLValue{"time until dies 0": values.untilDies0, "time until dies 1": values.untilDies1} {
		if got := value.GetDuration(sim); got != planAt(0.75) {
			t.Errorf("%s: expected the fight's remainder %s, got %s", name, planAt(0.75), got)
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
}

func TestValuePlanReadsDowntimesAndWaves(t *testing.T) {
	plan := &proto.EncounterPlan{
		Downtimes: []*proto.PlanDowntime{{TargetIndex: 0, Start: &proto.PlanTime{At: 0.25}, End: &proto.PlanTime{At: 0.5}}},
		AddWaves:  []*proto.AddWave{testAddWave(1, 0.6, 0.2, nil)},
	}
	sim, fw := setupPlanSim(planTestRequest(1, plan), nil)
	values := buildPlanValues(t, fw.Rotation, 1)

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

	runLifecycleUntil(sim, planAt(0.1))
	verify("before the downtime",
		check{"boss active", values.active0.GetBool(sim), true},
		check{"add active", values.active1.GetBool(sim), false},
		check{"adds", values.adds.GetInt(sim), int32(0)},
		check{"boss until active", values.untilActive0.GetDuration(sim), time.Duration(0)},
		check{"add until active", values.untilActive1.GetDuration(sim), planAt(0.5)},
		check{"boss until dies", values.untilDies0.GetDuration(sim), planAt(0.9)},
		check{"add until dies", values.untilDies1.GetDuration(sim), planAt(0.7)},
	)

	runLifecycleUntil(sim, planAt(0.3))
	verify("inside the downtime",
		check{"boss active", values.active0.GetBool(sim), false},
		check{"boss until active", values.untilActive0.GetDuration(sim), planAt(0.2)},
		check{"boss until dies", values.untilDies0.GetDuration(sim), planAt(0.7)},
	)

	runLifecycleUntil(sim, planAt(0.7))
	verify("inside the wave",
		check{"boss active", values.active0.GetBool(sim), true},
		check{"add active", values.active1.GetBool(sim), true},
		check{"adds", values.adds.GetInt(sim), int32(1)},
		check{"add until active", values.untilActive1.GetDuration(sim), time.Duration(0)},
		check{"add until dies", values.untilDies1.GetDuration(sim), planAt(0.1)},
	)

	runLifecycleUntil(sim, planAt(0.85))
	verify("after the wave",
		check{"add active", values.active1.GetBool(sim), false},
		check{"adds", values.adds.GetInt(sim), int32(0)},
		check{"add until active", values.untilActive1.GetDuration(sim), NeverExpires},
		check{"add until dies", values.untilDies1.GetDuration(sim), NeverExpires},
	)
}

func planBossAbilityValues(t *testing.T, rot *APLRotation, name string) (APLValue, APLValue) {
	t.Helper()
	active := planValue(t, rot, name+" active", bossAbilityIsActiveConfig(name, nil))
	until := planValue(t, rot, name+" until", timeUntilBossAbilityConfig(name, nil))
	return active, until
}

// A window kind is active while its window is open, the instant kinds never are, and the time until one is the
// soonest start of the abilities of its name.
func TestValuePlanReadsBossAbilities(t *testing.T) {
	named := func(name string, ability *proto.PlanBossAbility) *proto.PlanBossAbility {
		ability.Name = name
		return ability
	}
	bolt := planBossWindow(proto.PlanBossAbility_SpellHit, 0.7, 1)
	bolt.MinDamage = 100
	bolt.MaxDamage = 100
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{
		named("Fire", planPulse(0.2, 0.3, 1)),
		named("Fire", planAreaHit(0.15)),
		named("Slow", planSwingTiming(0.5, 0.6, 2)),
		named("Bolt", bolt),
	}}
	sim, fw := setupPlanSim(planTankRequest(plan), nil)
	fireActive, untilFire := planBossAbilityValues(t, fw.Rotation, "Fire")
	slowActive, untilSlow := planBossAbilityValues(t, fw.Rotation, "Slow")
	boltActive, untilBolt := planBossAbilityValues(t, fw.Rotation, "Bolt")

	type check struct {
		name string
		got  any
		want any
	}
	verify := func(checks ...check) {
		t.Helper()
		for _, c := range checks {
			if c.got != c.want {
				t.Errorf("at %s, %s: expected %v, got %v", sim.CurrentTime, c.name, c.want, c.got)
			}
		}
	}
	until := func(next float64) time.Duration { return planAt(next) - sim.CurrentTime }

	runLifecycleUntil(sim, planAt(0.1))
	verify(
		check{"Fire active", fireActive.GetBool(sim), false},
		check{"until Fire, the area hit first", untilFire.GetDuration(sim), until(0.15)},
		check{"Slow active", slowActive.GetBool(sim), false},
		check{"until Slow", untilSlow.GetDuration(sim), until(0.5)},
		check{"until Bolt", untilBolt.GetDuration(sim), until(0.7)},
	)
	runLifecycleUntil(sim, planAt(0.17))
	verify(check{"until Fire, the pulse next", untilFire.GetDuration(sim), until(0.2)})
	runLifecycleUntil(sim, planAt(0.25))
	verify(
		check{"Fire active", fireActive.GetBool(sim), true},
		check{"until Fire", untilFire.GetDuration(sim), time.Duration(0)},
	)
	runLifecycleUntil(sim, planAt(0.35))
	verify(
		check{"Fire active", fireActive.GetBool(sim), false},
		check{"until Fire", untilFire.GetDuration(sim), NeverExpires},
	)
	runLifecycleUntil(sim, planAt(0.55))
	verify(
		check{"Slow active", slowActive.GetBool(sim), true},
		check{"until Slow", untilSlow.GetDuration(sim), time.Duration(0)},
	)
	runLifecycleUntil(sim, planAt(0.65))
	verify(
		check{"Slow active", slowActive.GetBool(sim), false},
		check{"until Slow", untilSlow.GetDuration(sim), NeverExpires},
		check{"until Bolt", untilBolt.GetDuration(sim), until(0.7)},
	)
	runLifecycleUntil(sim, planAt(0.75))
	verify(
		check{"Bolt active", boltActive.GetBool(sim), false},
		check{"until Bolt", untilBolt.GetDuration(sim), NeverExpires},
	)
}

// An APL line with no name names no ability, an unnamed one included.
func TestValuePlanWarnsForAnUnknownBossAbility(t *testing.T) {
	named := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{planAreaHit(0.1)}}
	named.BossAbilities[0].Name = "Fire"
	unnamed := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{planAreaHit(0.1)}}
	for label, c := range map[string]struct {
		request *proto.RaidSimRequest
		name    string
	}{
		"another name": {planTestRequest(1, named), "Frost"},
		"no name":      {planTestRequest(1, unnamed), ""},
		"no plan":      {planTestRequest(1, nil), "Frost"},
	} {
		_, fw := setupPlanSim(c.request, nil)
		withUuid := func(config *proto.APLValue, uuid string) *proto.APLValue {
			config.Uuid = &proto.UUID{Value: uuid}
			return config
		}
		configs := map[string]*proto.APLValue{
			"Boss Ability Is Active":           withUuid(bossAbilityIsActiveConfig(c.name, nil), "a"),
			"Time Until Boss Ability":          withUuid(timeUntilBossAbilityConfig(c.name, nil), "b"),
			"Boss Ability Cast Time Remaining": withUuid(castTimeRemainingConfig(c.name, nil), "c"),
		}
		for name, config := range configs {
			if value := fw.Rotation.newAPLValue(config); value != nil {
				t.Errorf("%s, %s: got a value for an unknown ability, want none", label, name)
			}
			if warnings := fw.Rotation.uuidValidations[config.Uuid]; len(warnings) != 1 || warnings[0].LogLevel != proto.LogLevel_Warning {
				t.Errorf("%s, %s: validations = %v, want one warning", label, name, warnings)
			}
		}
		known := withUuid(bossAbilityIsKnownConfig(c.name, nil), "d")
		if value, ok := fw.Rotation.newAPLValue(known).(*APLValueConst); !ok || value.boolVal {
			t.Errorf("%s, Boss Ability Is Known: expected a constant false, got %v", label, value)
		}
		if warnings := fw.Rotation.uuidValidations[known.Uuid]; len(warnings) != 0 {
			t.Errorf("%s, Boss Ability Is Known: validations = %v, want none", label, warnings)
		}
	}
}

func planNamedSpellHit(name string, targetIndex int32, start float64, castTime time.Duration) *proto.PlanBossAbility {
	ability := planBossWindow(proto.PlanBossAbility_SpellHit, start, 1)
	ability.Name = name
	ability.TargetIndex = targetIndex
	ability.MinDamage, ability.MaxDamage = 100, 100
	ability.CastTime = castTime.Seconds()
	return ability
}

func castTimeRemainingConfig(name string, target *proto.UnitReference) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_BossAbilityCastTimeRemaining{BossAbilityCastTimeRemaining: &proto.APLValueBossAbilityCastTimeRemaining{AbilityName: name, TargetUnit: target}}}
}

func timeUntilBossAbilityConfig(name string, target *proto.UnitReference) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_TimeUntilBossAbility{TimeUntilBossAbility: &proto.APLValueTimeUntilBossAbility{AbilityName: name, TargetUnit: target}}}
}

func bossAbilityIsActiveConfig(name string, target *proto.UnitReference) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_BossAbilityIsActive{BossAbilityIsActive: &proto.APLValueBossAbilityIsActive{AbilityName: name, TargetUnit: target}}}
}

func bossAbilityIsKnownConfig(name string, target *proto.UnitReference) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_BossAbilityIsKnown{BossAbilityIsKnown: &proto.APLValueBossAbilityIsKnown{AbilityName: name, TargetUnit: target}}}
}

// The time left on a spell hit's hardcast, 0 before it starts and once it lands. A name with no spell hit that
// hardcasts reads a constant 0 with a warning.
func TestValuePlanReadsBossAbilityCastTimeRemaining(t *testing.T) {
	const castTime = 2 * time.Second
	fire := planAreaHit(0.2)
	fire.Name = "Fire"
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{
		planNamedSpellHit("Bolt", 0, 0.4, castTime),
		fire,
		planNamedSpellHit("Instant", 0, 0.5, 0),
	}}
	sim, fw := setupPlanSim(planTestRequest(1, plan), nil)
	remaining := planValue(t, fw.Rotation, "Bolt", castTimeRemainingConfig("Bolt", nil))

	for _, c := range []struct {
		name string
		at   time.Duration
		want time.Duration
	}{
		{"before the cast", planAt(0.3), 0},
		{"during the cast", planAt(0.4) + 500*time.Millisecond, castTime - 500*time.Millisecond},
		{"as it lands", planAt(0.4) + castTime, 0},
		{"after it lands", planAt(0.4) + castTime + time.Second, 0},
	} {
		runLifecycleUntil(sim, c.at)
		if got := remaining.GetDuration(sim); got != c.want {
			t.Errorf("%s: expected %s, got %s", c.name, c.want, got)
		}
	}

	for label, name := range map[string]string{"an area hit": "Fire", "a spell hit with no cast time": "Instant"} {
		config := castTimeRemainingConfig(name, nil)
		config.Uuid = &proto.UUID{Value: name}
		value, ok := fw.Rotation.newAPLValue(config).(*APLValueConst)
		if !ok || value.Type() != proto.APLValueType_ValueTypeDuration || value.GetDuration(sim) != 0 {
			t.Errorf("%s: expected a constant 0ms, got %v", label, value)
		}
		if warnings := fw.Rotation.uuidValidations[config.Uuid]; len(warnings) != 1 || warnings[0].LogLevel != proto.LogLevel_Warning {
			t.Errorf("%s: validations = %v, want one warning", label, warnings)
		}
	}
}

// Is Known folds to a constant with no unit or a fixed target: true for a name the plan has, on the unit when one
// is set. An area hit is on no unit, and an ability construction skipped is not known. A current target folds to
// false when no target has an ability of the name.
func TestValuePlanBossAbilityIsKnownFolds(t *testing.T) {
	fire := planAreaHit(0.2)
	fire.Name = "Fire"
	// The targets do not swing, so construction skips a swing timing.
	slow := planSwingTiming(0.2, 0.4, 2)
	slow.Name = "Slow"
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{planNamedSpellHit("Bolt", 1, 0.4, time.Second), fire, slow}}
	sim, fw := setupPlanSim(planTestRequest(2, plan), nil)
	current := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}

	for _, c := range []struct {
		label  string
		name   string
		target *proto.UnitReference
		want   bool
	}{
		{"a known name", "Bolt", nil, true},
		{"an unknown name", "Frost", nil, false},
		{"on the add, its target", "Bolt", planTargetRef(1), true},
		{"on the boss", "Bolt", planTargetRef(0), false},
		{"an area hit with no unit", "Fire", nil, true},
		{"an area hit on the boss", "Fire", planTargetRef(0), false},
		{"a skipped ability", "Slow", nil, false},
		{"an unknown name on the current target", "Frost", current, false},
		{"an area hit on the current target", "Fire", current, false},
	} {
		config := bossAbilityIsKnownConfig(c.name, c.target)
		config.Uuid = &proto.UUID{Value: c.label}
		value, ok := fw.Rotation.newAPLValue(config).(*APLValueConst)
		if !ok || value.boolVal != c.want {
			t.Errorf("%s: expected a constant %v, got %v", c.label, c.want, value)
		}
		if warnings := fw.Rotation.uuidValidations[config.Uuid]; len(warnings) != 0 {
			t.Errorf("%s: validations = %v, want none", c.label, warnings)
		}
	}

	slowConfig := bossAbilityIsActiveConfig("Slow", nil)
	slowActive := planValue(t, fw.Rotation, "Slow active", slowConfig)
	if warnings := fw.Rotation.uuidValidations[slowConfig.Uuid]; len(warnings) != 0 {
		t.Errorf("Boss Ability Is Active of a skipped ability: validations = %v, want none", warnings)
	}
	runLifecycleUntil(sim, planAt(0.3))
	if slowActive.GetBool(sim) {
		t.Error("Boss Ability Is Active of a skipped ability inside its window: expected false")
	}
}

// Is Known with a current target reads the target at each evaluation.
func TestValuePlanBossAbilityIsKnownReadsTheCurrentTarget(t *testing.T) {
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{planNamedSpellHit("Bolt", 1, 0.4, time.Second)}}
	sim, fw := setupPlanSim(planTestRequest(2, plan), nil)
	config := bossAbilityIsKnownConfig("Bolt", &proto.UnitReference{Type: proto.UnitReference_CurrentTarget})
	config.Uuid = &proto.UUID{Value: "known"}
	known := fw.Rotation.newAPLValue(config)
	if _, folded := known.(*APLValueConst); folded || known == nil {
		t.Fatalf("expected a value read at each evaluation, got %v", known)
	}
	if warnings := fw.Rotation.uuidValidations[config.Uuid]; len(warnings) != 0 {
		t.Errorf("validations = %v, want none", warnings)
	}
	if known.GetBool(sim) {
		t.Error("on the boss, the current target: expected false")
	}
	fw.CurrentTarget = &sim.Encounter.AllTargets[1].Unit
	if !known.GetBool(sim) {
		t.Error("on the add, the current target: expected true")
	}
}

// With a set unit Is Active reads only the windows of that unit's copy of the name.
func TestValuePlanBossAbilityIsActiveOnASetTarget(t *testing.T) {
	window := func(targetIndex int32, start, end float64) *proto.PlanBossAbility {
		ability := planSwingTiming(start, end, 2)
		ability.Name = "Slow"
		ability.TargetIndex = targetIndex
		return ability
	}
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{window(0, 0.2, 0.4), window(1, 0.5, 0.7)}}
	request := planTestRequest(2, plan)
	for _, target := range request.Encounter.Targets {
		target.SwingSpeed = planBossSwingSpeed.Seconds()
		target.MinBaseDamage = 1000
	}
	sim, fw := setupPlanSim(request, nil)
	isActive := func(name string, target *proto.UnitReference) APLValue {
		return planValue(t, fw.Rotation, name, bossAbilityIsActiveConfig("Slow", target))
	}
	onAdd := isActive("on the add", planTargetRef(1))
	onCurrent := isActive("on the current target", &proto.UnitReference{Type: proto.UnitReference_CurrentTarget})

	type check struct {
		name string
		got  bool
		want bool
	}
	verify := func(checks ...check) {
		t.Helper()
		for _, c := range checks {
			if c.got != c.want {
				t.Errorf("at %s, %s: expected %v, got %v", sim.CurrentTime, c.name, c.want, c.got)
			}
		}
	}

	runLifecycleUntil(sim, planAt(0.3))
	verify(
		check{"on the add, the boss's window open", onAdd.GetBool(sim), false},
		check{"on the current target, the boss", onCurrent.GetBool(sim), true},
	)
	runLifecycleUntil(sim, planAt(0.6))
	verify(
		check{"on the add, its window open", onAdd.GetBool(sim), true},
		check{"on the current target, the boss", onCurrent.GetBool(sim), false},
	)
	fw.CurrentTarget = &sim.Encounter.AllTargets[1].Unit
	verify(check{"on the current target, the add", onCurrent.GetBool(sim), true})
}

// With a unit the values read only that unit's copy of a shared name, a current or next target the one it is at
// each read; with none, every copy.
func TestValuePlanBossAbilityUnitFilter(t *testing.T) {
	const castTime = 2 * time.Second
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{
		planNamedSpellHit("Bolt", 0, 0.3, castTime),
		planNamedSpellHit("Bolt", 1, 0.5, castTime),
	}}
	sim, fw := setupPlanSim(planTestRequest(2, plan), nil)
	rot := fw.Rotation
	current := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}
	untilAny := planValue(t, rot, "until any", timeUntilBossAbilityConfig("Bolt", nil))
	untilAdd := planValue(t, rot, "until add", timeUntilBossAbilityConfig("Bolt", planTargetRef(1)))
	untilCurrent := planValue(t, rot, "until current", timeUntilBossAbilityConfig("Bolt", current))
	untilNext := planValue(t, rot, "until next", timeUntilBossAbilityConfig("Bolt", &proto.UnitReference{Type: proto.UnitReference_NextTarget}))
	castAny := planValue(t, rot, "cast any", castTimeRemainingConfig("Bolt", nil))
	castAdd := planValue(t, rot, "cast add", castTimeRemainingConfig("Bolt", planTargetRef(1)))

	until := func(next float64) time.Duration { return planAt(next) - sim.CurrentTime }
	type check struct {
		name string
		got  time.Duration
		want time.Duration
	}
	verify := func(checks ...check) {
		t.Helper()
		for _, c := range checks {
			if c.got != c.want {
				t.Errorf("at %s, %s: expected %s, got %s", sim.CurrentTime, c.name, c.want, c.got)
			}
		}
	}

	runLifecycleUntil(sim, planAt(0.1))
	verify(
		check{"until any, the boss's first", untilAny.GetDuration(sim), until(0.3)},
		check{"until the add's", untilAdd.GetDuration(sim), until(0.5)},
		check{"until the current target's, the boss's", untilCurrent.GetDuration(sim), until(0.3)},
		check{"until the next target's, the add's", untilNext.GetDuration(sim), until(0.5)},
	)
	fw.CurrentTarget = &sim.Encounter.AllTargets[1].Unit
	verify(
		check{"until the current target's, the add's", untilCurrent.GetDuration(sim), until(0.5)},
		check{"until the next target's, the boss's", untilNext.GetDuration(sim), until(0.3)},
	)

	runLifecycleUntil(sim, planAt(0.3)+time.Second)
	verify(
		check{"any cast, the boss casting", castAny.GetDuration(sim), castTime - time.Second},
		check{"the add's cast, the boss casting", castAdd.GetDuration(sim), 0},
	)
	runLifecycleUntil(sim, planAt(0.5)+time.Second)
	verify(
		check{"any cast, the add casting", castAny.GetDuration(sim), castTime - time.Second},
		check{"the add's cast", castAdd.GetDuration(sim), castTime - time.Second},
	)
}

func TestValuePlanWarnsForUnitsThatAreNotEnemyTargets(t *testing.T) {
	_, fw := setupPlanSim(planTestRequest(1, nil), nil)
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

func TestValuePlanWarnsForAnUnknownAmp(t *testing.T) {
	_, fw := setupPlanSim(planTestRequest(1, nil), nil)
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

func TestValuePlanFallsBackWhenTheCurrentTargetIsNotAnEnemy(t *testing.T) {
	sim, fw := setupPlanSim(planTestRequest(1, nil), nil)
	rot := fw.Rotation
	current := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}
	active := planValue(t, rot, "active", &proto.APLValue{Value: &proto.APLValue_TargetIsActive{TargetIsActive: &proto.APLValueTargetIsActive{TargetUnit: current}}})
	untilActive := planValue(t, rot, "untilActive", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetActive{TimeUntilTargetActive: &proto.APLValueTimeUntilTargetActive{TargetUnit: current}}})
	untilDies := planValue(t, rot, "untilDies", &proto.APLValue{Value: &proto.APLValue_TimeUntilTargetDies{TimeUntilTargetDies: &proto.APLValueTimeUntilTargetDies{TargetUnit: current}}})

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
}
