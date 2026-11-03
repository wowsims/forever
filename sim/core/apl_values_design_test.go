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
		for _, add := range wave.Adds {
			add.Stats = designHealthStats()
		}
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

func designBossAbilityValues(t *testing.T, rot *APLRotation, name string) (APLValue, APLValue) {
	t.Helper()
	active := designValue(t, rot, name+" active", &proto.APLValue{Value: &proto.APLValue_BossAbilityIsActive{BossAbilityIsActive: &proto.APLValueBossAbilityIsActive{AbilityName: name}}})
	until := designValue(t, rot, name+" until", &proto.APLValue{Value: &proto.APLValue_TimeUntilBossAbility{TimeUntilBossAbility: &proto.APLValueTimeUntilBossAbility{AbilityName: name}}})
	return active, until
}

// A window kind is active while its window is open, the instant kinds never are, and the time until one is the
// soonest start of the abilities of its name.
func TestValueDesignReadsBossAbilities(t *testing.T) {
	named := func(name string, ability *proto.DesignBossAbility) *proto.DesignBossAbility {
		ability.Name = name
		return ability
	}
	bolt := designBossWindow(proto.DesignBossAbility_SpellHit, 0.7, 1)
	bolt.MinDamage = 100
	bolt.MaxDamage = 100
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{
		named("Fire", designPulse(0.2, 0.3, 1)),
		named("Fire", designAreaHit(0.15)),
		named("Slow", designSwingTiming(0.5, 0.6, 2)),
		named("Bolt", bolt),
	}}
	sim, fw := setupDesignSim(designTankRequest(design), nil)
	fireActive, untilFire := designBossAbilityValues(t, fw.Rotation, "Fire")
	slowActive, untilSlow := designBossAbilityValues(t, fw.Rotation, "Slow")
	boltActive, untilBolt := designBossAbilityValues(t, fw.Rotation, "Bolt")

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
	until := func(next float64) time.Duration { return designAt(next) - sim.CurrentTime }

	runLifecycleUntil(sim, designAt(0.1))
	verify(
		check{"Fire active", fireActive.GetBool(sim), false},
		check{"until Fire, the area hit first", untilFire.GetDuration(sim), until(0.15)},
		check{"Slow active", slowActive.GetBool(sim), false},
		check{"until Slow", untilSlow.GetDuration(sim), until(0.5)},
		check{"until Bolt", untilBolt.GetDuration(sim), until(0.7)},
	)
	runLifecycleUntil(sim, designAt(0.17))
	verify(check{"until Fire, the pulse next", untilFire.GetDuration(sim), until(0.2)})
	runLifecycleUntil(sim, designAt(0.25))
	verify(
		check{"Fire active", fireActive.GetBool(sim), true},
		check{"until Fire", untilFire.GetDuration(sim), time.Duration(0)},
	)
	runLifecycleUntil(sim, designAt(0.35))
	verify(
		check{"Fire active", fireActive.GetBool(sim), false},
		check{"until Fire", untilFire.GetDuration(sim), NeverExpires},
	)
	runLifecycleUntil(sim, designAt(0.55))
	verify(
		check{"Slow active", slowActive.GetBool(sim), true},
		check{"until Slow", untilSlow.GetDuration(sim), time.Duration(0)},
	)
	runLifecycleUntil(sim, designAt(0.65))
	verify(
		check{"Slow active", slowActive.GetBool(sim), false},
		check{"until Slow", untilSlow.GetDuration(sim), NeverExpires},
		check{"until Bolt", untilBolt.GetDuration(sim), until(0.7)},
	)
	runLifecycleUntil(sim, designAt(0.75))
	verify(
		check{"Bolt active", boltActive.GetBool(sim), false},
		check{"until Bolt", untilBolt.GetDuration(sim), NeverExpires},
	)
}

// An APL line with no name names no ability, an unnamed one included.
func TestValueDesignWarnsForAnUnknownBossAbility(t *testing.T) {
	named := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{designAreaHit(0.1)}}
	named.BossAbilities[0].Name = "Fire"
	unnamed := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{designAreaHit(0.1)}}
	for label, c := range map[string]struct {
		request *proto.RaidSimRequest
		name    string
	}{
		"another name": {designTestRequest(1, named), "Frost"},
		"no name":      {designTestRequest(1, unnamed), ""},
		"no design":    {designTestRequest(1, nil), "Frost"},
	} {
		_, fw := setupDesignSim(c.request, nil)
		configs := map[string]*proto.APLValue{
			"Boss Ability Is Active":           {Uuid: &proto.UUID{Value: "a"}, Value: &proto.APLValue_BossAbilityIsActive{BossAbilityIsActive: &proto.APLValueBossAbilityIsActive{AbilityName: c.name}}},
			"Time Until Boss Ability":          {Uuid: &proto.UUID{Value: "b"}, Value: &proto.APLValue_TimeUntilBossAbility{TimeUntilBossAbility: &proto.APLValueTimeUntilBossAbility{AbilityName: c.name}}},
			"Boss Ability Cast Time Remaining": {Uuid: &proto.UUID{Value: "c"}, Value: &proto.APLValue_BossAbilityCastTimeRemaining{BossAbilityCastTimeRemaining: &proto.APLValueBossAbilityCastTimeRemaining{AbilityName: c.name}}},
		}
		for name, config := range configs {
			if value := fw.Rotation.newAPLValue(config); value != nil {
				t.Errorf("%s, %s: got a value for an unknown ability, want none", label, name)
			}
			if warnings := fw.Rotation.uuidValidations[config.Uuid]; len(warnings) != 1 || warnings[0].LogLevel != proto.LogLevel_Warning {
				t.Errorf("%s, %s: validations = %v, want one warning", label, name, warnings)
			}
		}
		known := &proto.APLValue{Uuid: &proto.UUID{Value: "d"}, Value: &proto.APLValue_BossAbilityIsKnown{BossAbilityIsKnown: &proto.APLValueBossAbilityIsKnown{AbilityName: c.name}}}
		if value, ok := fw.Rotation.newAPLValue(known).(*APLValueConst); !ok || value.boolVal {
			t.Errorf("%s, Boss Ability Is Known: expected a constant false, got %v", label, value)
		}
		if warnings := fw.Rotation.uuidValidations[known.Uuid]; len(warnings) != 0 {
			t.Errorf("%s, Boss Ability Is Known: validations = %v, want none", label, warnings)
		}
	}
}

func designNamedSpellHit(name string, targetIndex int32, start float64, castTime time.Duration) *proto.DesignBossAbility {
	ability := designBossWindow(proto.DesignBossAbility_SpellHit, start, 1)
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

func bossAbilityIsKnownConfig(name string, target *proto.UnitReference) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_BossAbilityIsKnown{BossAbilityIsKnown: &proto.APLValueBossAbilityIsKnown{AbilityName: name, TargetUnit: target}}}
}

// The time left on a spell hit's hardcast, 0 before it starts and once it lands. A name with no spell hit that
// hardcasts reads a constant 0 with a warning.
func TestValueDesignReadsBossAbilityCastTimeRemaining(t *testing.T) {
	const castTime = 2 * time.Second
	fire := designAreaHit(0.2)
	fire.Name = "Fire"
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{
		designNamedSpellHit("Bolt", 0, 0.4, castTime),
		fire,
		designNamedSpellHit("Instant", 0, 0.5, 0),
	}}
	sim, fw := setupDesignSim(designTestRequest(1, design), nil)
	remaining := designValue(t, fw.Rotation, "Bolt", castTimeRemainingConfig("Bolt", nil))

	for _, c := range []struct {
		name string
		at   time.Duration
		want time.Duration
	}{
		{"before the cast", designAt(0.3), 0},
		{"during the cast", designAt(0.4) + 500*time.Millisecond, castTime - 500*time.Millisecond},
		{"as it lands", designAt(0.4) + castTime, 0},
		{"after it lands", designAt(0.4) + castTime + time.Second, 0},
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

// Is Known folds to a constant with no unit or a fixed target: true for a name the design has, on the unit when one
// is set. An area hit is on no unit, and an ability construction skipped is not known. A current target folds to
// false when no target has an ability of the name.
func TestValueDesignBossAbilityIsKnownFolds(t *testing.T) {
	fire := designAreaHit(0.2)
	fire.Name = "Fire"
	// The targets do not swing, so construction skips a swing timing.
	slow := designSwingTiming(0.2, 0.4, 2)
	slow.Name = "Slow"
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{designNamedSpellHit("Bolt", 1, 0.4, time.Second), fire, slow}}
	sim, fw := setupDesignSim(designTestRequest(2, design), nil)
	current := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}

	for _, c := range []struct {
		label  string
		name   string
		target *proto.UnitReference
		want   bool
	}{
		{"a known name", "Bolt", nil, true},
		{"an unknown name", "Frost", nil, false},
		{"on the add, its target", "Bolt", designTargetRef(1), true},
		{"on the boss", "Bolt", designTargetRef(0), false},
		{"an area hit with no unit", "Fire", nil, true},
		{"an area hit on the boss", "Fire", designTargetRef(0), false},
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

	slowConfig := &proto.APLValue{Value: &proto.APLValue_BossAbilityIsActive{BossAbilityIsActive: &proto.APLValueBossAbilityIsActive{AbilityName: "Slow"}}}
	slowActive := designValue(t, fw.Rotation, "Slow active", slowConfig)
	if warnings := fw.Rotation.uuidValidations[slowConfig.Uuid]; len(warnings) != 0 {
		t.Errorf("Boss Ability Is Active of a skipped ability: validations = %v, want none", warnings)
	}
	runLifecycleUntil(sim, designAt(0.3))
	if slowActive.GetBool(sim) {
		t.Error("Boss Ability Is Active of a skipped ability inside its window: expected false")
	}
}

// Is Known with a current target reads the target at each evaluation.
func TestValueDesignBossAbilityIsKnownReadsTheCurrentTarget(t *testing.T) {
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{designNamedSpellHit("Bolt", 1, 0.4, time.Second)}}
	sim, fw := setupDesignSim(designTestRequest(2, design), nil)
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
func TestValueDesignBossAbilityIsActiveOnASetTarget(t *testing.T) {
	window := func(targetIndex int32, start, end float64) *proto.DesignBossAbility {
		ability := designSwingTiming(start, end, 2)
		ability.Name = "Slow"
		ability.TargetIndex = targetIndex
		return ability
	}
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{window(0, 0.2, 0.4), window(1, 0.5, 0.7)}}
	request := designTestRequest(2, design)
	for _, target := range request.Encounter.Targets {
		target.SwingSpeed = designBossSwingSpeed.Seconds()
		target.MinBaseDamage = 1000
	}
	sim, fw := setupDesignSim(request, nil)
	isActive := func(name string, target *proto.UnitReference) APLValue {
		return designValue(t, fw.Rotation, name, &proto.APLValue{Value: &proto.APLValue_BossAbilityIsActive{BossAbilityIsActive: &proto.APLValueBossAbilityIsActive{AbilityName: "Slow", TargetUnit: target}}})
	}
	onAdd := isActive("on the add", designTargetRef(1))
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

	runLifecycleUntil(sim, designAt(0.3))
	verify(
		check{"on the add, the boss's window open", onAdd.GetBool(sim), false},
		check{"on the current target, the boss", onCurrent.GetBool(sim), true},
	)
	runLifecycleUntil(sim, designAt(0.6))
	verify(
		check{"on the add, its window open", onAdd.GetBool(sim), true},
		check{"on the current target, the boss", onCurrent.GetBool(sim), false},
	)
	fw.CurrentTarget = &sim.Encounter.AllTargets[1].Unit
	verify(check{"on the current target, the add", onCurrent.GetBool(sim), true})
}

// With a unit the values read only that unit's copy of a shared name, a current or next target the one it is at
// each read; with none, every copy.
func TestValueDesignBossAbilityUnitFilter(t *testing.T) {
	const castTime = 2 * time.Second
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{
		designNamedSpellHit("Bolt", 0, 0.3, castTime),
		designNamedSpellHit("Bolt", 1, 0.5, castTime),
	}}
	sim, fw := setupDesignSim(designTestRequest(2, design), nil)
	rot := fw.Rotation
	current := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}
	untilAny := designValue(t, rot, "until any", timeUntilBossAbilityConfig("Bolt", nil))
	untilAdd := designValue(t, rot, "until add", timeUntilBossAbilityConfig("Bolt", designTargetRef(1)))
	untilCurrent := designValue(t, rot, "until current", timeUntilBossAbilityConfig("Bolt", current))
	untilNext := designValue(t, rot, "until next", timeUntilBossAbilityConfig("Bolt", &proto.UnitReference{Type: proto.UnitReference_NextTarget}))
	castAny := designValue(t, rot, "cast any", castTimeRemainingConfig("Bolt", nil))
	castAdd := designValue(t, rot, "cast add", castTimeRemainingConfig("Bolt", designTargetRef(1)))

	until := func(next float64) time.Duration { return designAt(next) - sim.CurrentTime }
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

	runLifecycleUntil(sim, designAt(0.1))
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

	runLifecycleUntil(sim, designAt(0.3)+time.Second)
	verify(
		check{"any cast, the boss casting", castAny.GetDuration(sim), castTime - time.Second},
		check{"the add's cast, the boss casting", castAdd.GetDuration(sim), 0},
	)
	runLifecycleUntil(sim, designAt(0.5)+time.Second)
	verify(
		check{"any cast, the add casting", castAny.GetDuration(sim), castTime - time.Second},
		check{"the add's cast", castAdd.GetDuration(sim), castTime - time.Second},
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
	health := &proto.APLValue{Uuid: &proto.UUID{Value: "a"}, Value: &proto.APLValue_CurrentHealth{CurrentHealth: &proto.APLValueCurrentHealth{SourceUnit: boss}}}
	if value := fw.Rotation.newAPLValue(health); value == nil {
		t.Error("Current Health: expected the value to be kept")
	}
	if warnings := fw.Rotation.uuidValidations[health.Uuid]; len(warnings) != 1 || warnings[0].LogLevel != proto.LogLevel_Warning {
		t.Errorf("Current Health: validations = %v, want one warning", warnings)
	}

	// The percent reads the target's curve, which needs no Health stat.
	percent := &proto.APLValue{Uuid: &proto.UUID{Value: "b"}, Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{SourceUnit: boss}}}
	if value := fw.Rotation.newAPLValue(percent); value == nil {
		t.Error("Current Health %: expected the value to be kept")
	}
	if warnings := fw.Rotation.uuidValidations[percent.Uuid]; len(warnings) != 0 {
		t.Errorf("Current Health %%: validations = %v, want none", warnings)
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
