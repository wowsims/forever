package core

import (
	"fmt"
	"maps"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	googleproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func setupFakeRageRotation() (*Simulation, *FakeRageWarrior, *APLRotation) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	return sim, fw, fw.Rotation
}

// Auras and spells only register while the environment is constructing, so register runs with the
// state set back to Constructed on the already reset sim.
func setupFakeRageRotationWith(register func(char *Character)) (*Simulation, *FakeRageWarrior, *APLRotation) {
	return setupFakeRageRotationFrom(fakeRageSimRequest(), register)
}

func setupFakeRageRotationFrom(request *proto.RaidSimRequest, register func(char *Character)) (*Simulation, *FakeRageWarrior, *APLRotation) {
	sim := setupFakeRageSimFrom(request)
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	sim.Environment.State = Constructed
	register(&fw.Character)
	sim.Environment.State = Finalized
	return sim, fw, fw.Rotation
}

const (
	testPlainAuraID   = 910001
	testStackAuraID   = 910002
	testICDAuraID     = 910003
	testMissingAuraID = 910099
)

func constValue(val string) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: val}}}
}

// A yes/no value no constant stands in for.
func yesNoValue() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
		Op: proto.APLValueCompare_OpLt, Lhs: &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}}, Rhs: constValue("1s"),
	}}}
}

// The validations are one warning per wanted text, in order, each containing it.
func wantWarnings(t *testing.T, label string, validations []*proto.APLValidation, want ...string) {
	t.Helper()
	ok := len(validations) == len(want)
	for i := 0; ok && i < len(want); i++ {
		ok = validations[i].LogLevel == proto.LogLevel_Warning && strings.Contains(validations[i].Validation, want[i])
	}
	if !ok {
		t.Errorf("%s: validations = %v, want warnings %q", label, validations, want)
	}
}

func wantNoValue(t *testing.T, values map[string]APLValue) {
	t.Helper()
	for name, value := range values {
		if value != nil {
			t.Errorf("%s: got a value, want none", name)
		}
	}
}

func renderValue(sim *Simulation, value APLValue) string {
	switch value.Type() {
	case proto.APLValueType_ValueTypeBool:
		return fmt.Sprint(value.GetBool(sim))
	case proto.APLValueType_ValueTypeInt:
		return fmt.Sprint(value.GetInt(sim))
	case proto.APLValueType_ValueTypeFloat:
		return fmt.Sprint(value.GetFloat(sim))
	case proto.APLValueType_ValueTypeDuration:
		return value.GetDuration(sim).String()
	default:
		return value.Type().String()
	}
}

func newStatAura(char *Character, label string, id int32, stat stats.Stat, duration time.Duration) *StatBuffAura {
	var buffs stats.Stats
	buffs[stat] = 100
	return char.NewTemporaryStatsAura(label, ActionID{SpellID: id}, buffs, duration)
}

type stubAction struct {
	defaultAPLActionImpl
	ready bool
}

func (action *stubAction) IsReady(*Simulation) bool { return action.ready }
func (action *stubAction) Execute(*Simulation)      {}
func (action *stubAction) String() string           { return "test action" }

func TestValueConst(t *testing.T) {
	sim := &Simulation{}
	unit := &Unit{}
	rot := &APLRotation{
		unit: unit,
	}

	stringVal := rot.newValueConst(&proto.APLValueConst{Val: "test str"}, &proto.UUID{Value: ""})
	if stringVal.GetString(sim) != "test str" {
		t.Fatalf("Unexpected string value %s", stringVal.GetString(sim))
	}

	intVal := rot.newValueConst(&proto.APLValueConst{Val: "10"}, &proto.UUID{Value: ""})
	if intVal.GetInt(sim) != 10 {
		t.Fatalf("Unexpected int value %d", intVal.GetInt(sim))
	}

	floatVal := rot.newValueConst(&proto.APLValueConst{Val: "10.123"}, &proto.UUID{Value: ""})
	if floatVal.GetFloat(sim) != 10.123 {
		t.Fatalf("Unexpected float value %f", floatVal.GetFloat(sim))
	}

	durVal := rot.newValueConst(&proto.APLValueConst{Val: "10.123s"}, &proto.UUID{Value: ""})
	if durVal.GetDuration(sim) != time.Millisecond*10123 {
		t.Fatalf("Unexpected duration value %s", durVal.GetDuration(sim))
	}

	coercedDurVal := rot.coerceTo(floatVal, proto.APLValueType_ValueTypeDuration)
	if _, ok := coercedDurVal.(*APLValueConst); !ok {
		t.Fatalf("Failed to skip coerce wrapper for duration value")
	}
	if coercedDurVal.GetDuration(sim) != time.Millisecond*10123 {
		t.Fatalf("Unexpected coerced duration value %s", coercedDurVal.GetDuration(sim))
	}
}

func TestValueOperators(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()
	lt, ge, eq := proto.APLValueCompare_OpLt, proto.APLValueCompare_OpGe, proto.APLValueCompare_OpEq
	add, mul, div := proto.APLValueMath_OpAdd, proto.APLValueMath_OpMul, proto.APLValueMath_OpDiv
	consts := func(vals ...string) []*proto.APLValue {
		out := make([]*proto.APLValue, len(vals))
		for i, val := range vals {
			out[i] = constValue(val)
		}
		return out
	}
	cmpValue := func(op proto.APLValueCompare_ComparisonOperator, lhs, rhs string) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: op, Lhs: constValue(lhs), Rhs: constValue(rhs)}}}
	}
	mathValue := func(op proto.APLValueMath_MathOperator, lhs, rhs string) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: op, Lhs: constValue(lhs), Rhs: constValue(rhs)}}}
	}
	maxValue := func(vals ...string) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_Max{Max: &proto.APLValueMax{Vals: consts(vals...)}}}
	}
	minValue := func(vals ...string) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_Min{Min: &proto.APLValueMin{Vals: consts(vals...)}}}
	}
	not := func(val *proto.APLValue) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: val}}}
	}
	and := func(vals ...*proto.APLValue) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: vals}}}
	}
	or := func(vals ...*proto.APLValue) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_Or{Or: &proto.APLValueOr{Vals: vals}}}
	}
	yes, no := cmpValue(lt, "1", "2"), cmpValue(lt, "2", "1")

	tests := []struct {
		name string
		cfg  *proto.APLValue
		want string // empty means the constructor returns nil
	}{
		{"cmp int lt", cmpValue(lt, "1", "2"), "true"},
		{"cmp int ge", cmpValue(ge, "1", "2"), "false"},
		{"cmp int against float", cmpValue(lt, "1", "1.5"), "true"},
		{"cmp duration", cmpValue(ge, "3s", "2s"), "true"},
		{"cmp bool eq", cmpValue(eq, "true", "true"), "true"},
		{"cmp bool lt is rejected", cmpValue(lt, "true", "false"), ""},

		{"math int add", mathValue(add, "3", "4"), "7"},
		{"math int div truncates", mathValue(div, "6", "4"), "1"},
		{"math float mul", mathValue(mul, "1.5", "2"), "3"},
		{"math duration times int", mathValue(mul, "2s", "3"), "6s"},
		{"math duration over duration", mathValue(div, "6s", "2s"), "3"},
		{"math duration times duration is rejected", mathValue(mul, "2s", "3s"), ""},
		{"math int over duration is rejected", mathValue(div, "2", "3s"), ""},
		{"math bool is rejected", mathValue(add, "true", "1"), ""},
		{"math string is rejected", mathValue(add, "abc", "abc"), ""},

		{"max int", maxValue("3", "7", "5"), "7"},
		{"min int", minValue("3", "7", "5"), "3"},
		{"max coerces int and float", maxValue("3", "2.5"), "3"},
		{"min of durations", minValue("3s", "2s"), "2s"},
		{"max single value", maxValue("4"), "4"},
		{"min empty", minValue(), ""},

		{"and both true", and(yes, yes), "true"},
		{"and one false", and(yes, no), "false"},
		{"and const false short-circuits", and(yes, constValue("false")), "false"},
		{"and empty", and(), ""},
		{"or both false", or(no, no), "false"},
		{"or one true", or(no, yes), "true"},
		{"or const true short-circuits", or(no, constValue("true")), "true"},
		{"or empty", or(), ""},
		{"not of true", not(yes), "false"},
		{"not of false", not(no), "true"},
		{"not of const true", not(constValue("true")), "false"},
		{"not of nothing", not(nil), ""},
	}

	for _, tc := range tests {
		got := rot.newAPLValue(tc.cfg)
		if tc.want == "" {
			if got != nil {
				t.Errorf("%s: want no value, got %s", tc.name, got)
			}
			continue
		}
		if got == nil {
			t.Errorf("%s: got no value, want %s", tc.name, tc.want)
			continue
		}
		if rendered := renderValue(sim, got); rendered != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, rendered, tc.want)
		}
	}
}

func TestValueVariables(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()
	rot.valueVariables = []*APLValueVariable{{name: "five", value: constValue("5")}}

	ref := rot.newValueVariableRef(&proto.APLValueVariableRef{Name: "five"}, nil)
	if ref == nil || ref.Type() != proto.APLValueType_ValueTypeInt || ref.GetInt(sim) != 5 {
		t.Errorf("a reference to a variable holding 5 should resolve to the int 5, got %v", ref)
	}

	placeholder := rot.newValueVariablePlaceholder(&proto.APLValueVariablePlaceholder{Name: "x"}, nil)
	if placeholder == nil || placeholder.Type() != proto.APLValueType_ValueTypeUnknown {
		t.Errorf("a named placeholder should construct with the unknown type, got %v", placeholder)
	}

	wantNoValue(t, map[string]APLValue{
		"Variable Ref to an unknown variable": rot.newValueVariableRef(&proto.APLValueVariableRef{Name: "missing"}, nil),
		"Variable Placeholder without a name": rot.newValueVariablePlaceholder(&proto.APLValueVariablePlaceholder{}, nil),
	})
}

// A group reference fills the group's placeholders in its actions' values, those of an action inside a sequence
// included, as it does in their conditions, and checks what it builds as the constructors do. A fill of a type the
// value cannot take, or one that is empty, leaves the placeholder reading zero with a warning, and a placeholder
// left unfilled is an error.
func TestGroupFillsActionValues(t *testing.T) {
	placeholder := func(name string) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_VariablePlaceholder{VariablePlaceholder: &proto.APLValueVariablePlaceholder{Name: name}}}
	}
	reference := func(group string, variables ...*proto.APLValueVariable) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_GroupReference{GroupReference: &proto.APLActionGroupReference{GroupName: group, Variables: variables}}}}
	}
	move := func(moveRange *proto.APLValue) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: moveRange}}}}
	}
	moveDuration := func(duration *proto.APLValue) *proto.APLAction {
		return &proto.APLAction{Condition: placeholder("when"), Action: &proto.APLAction_MoveDuration{MoveDuration: &proto.APLActionMoveDuration{Duration: duration}}}
	}
	sim := setupFakeRageSimFrom(fakeRageSimRequest())
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	rot := fw.newAPLRotation(&proto.APLRotation{
		Type: proto.APLRotation_TypeAPL,
		PriorityList: []*proto.APLListItem{
			reference("filled",
				&proto.APLValueVariable{Name: "time", Value: constValue("2s")},
				&proto.APLValueVariable{Name: "range", Value: constValue("7")},
				&proto.APLValueVariable{Name: "far", Value: constValue("3s")},
				&proto.APLValueVariable{Name: "when", Value: constValue("true")},
				&proto.APLValueVariable{Name: "yesNo", Value: yesNoValue()},
				&proto.APLValueVariable{Name: "blank", Value: &proto.APLValue{}},
			),
			reference("unfilled"),
		},
		Groups: []*proto.APLGroup{
			{
				Name: "filled",
				Actions: []*proto.APLListItem{
					{Action: &proto.APLAction{Action: &proto.APLAction_Wait{Wait: &proto.APLActionWait{Duration: placeholder("time")}}}},
					move(placeholder("range")),
					move(placeholder("far")),
					{Action: &proto.APLAction{Action: &proto.APLAction_Sequence{Sequence: &proto.APLActionSequence{Name: "walk", Actions: []*proto.APLAction{moveDuration(placeholder("time"))}}}}},
					{Action: moveDuration(placeholder("yesNo"))},
					{Action: &proto.APLAction{
						Condition: &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: proto.APLValueCompare_OpGt, Lhs: placeholder("blank"), Rhs: constValue("1")}}},
						Action:    &proto.APLAction_Wait{Wait: &proto.APLActionWait{Duration: placeholder("blank")}},
					}},
					{Action: &proto.APLAction{Action: &proto.APLAction_Wait{Wait: &proto.APLActionWait{Duration: &proto.APLValue{Value: &proto.APLValue_Max{Max: &proto.APLValueMax{
						Vals: []*proto.APLValue{placeholder("time"), {Value: &proto.APLValue_NumberTargets{NumberTargets: &proto.APLValueNumberTargets{}}}},
					}}}}}}},
					{Action: &proto.APLAction{Action: &proto.APLAction_Wait{Wait: &proto.APLActionWait{Duration: &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{
						Op: proto.APLValueMath_OpMul, Lhs: placeholder("time"), Rhs: placeholder("time"),
					}}}}}}},
					{Action: &proto.APLAction{Action: &proto.APLAction_Wait{Wait: &proto.APLActionWait{Duration: &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{
						Op: proto.APLValueMath_OpMul, Lhs: placeholder("blank"), Rhs: constValue("2s"),
					}}}}}}},
				},
			},
			{Name: "unfilled", Actions: []*proto.APLListItem{move(placeholder("missing"))}},
		},
	})

	group := rot.groups[0].actions
	inSequence := group[3].impl.(*APLActionSequence).subactions[0]
	for _, tc := range []struct {
		name      string
		got, want any
	}{
		{"Wait", group[0].impl.(*APLActionWait).duration.GetDuration(sim), 2 * time.Second},
		{"Move", group[1].impl.(*APLActionMove).moveRange.GetFloat(sim), 7.0},
		{"Move on a duration", group[2].impl.(*APLActionMove).moveRange.GetFloat(sim), 3.0},
		{"Move Duration in a sequence", inSequence.impl.(*APLActionMoveDuration).moveDuration.GetDuration(sim), 2 * time.Second},
		{"its condition", inSequence.condition.GetBool(sim), true},
		{"Move Duration of a yes/no", group[4].impl.(*APLActionMoveDuration).moveDuration.GetDuration(sim), time.Duration(0)},
		{"Wait on an empty variable", group[5].impl.(*APLActionWait).duration.GetDuration(sim), time.Duration(0)},
		{"a comparison of it", group[5].condition.GetBool(sim), false},
		{"Wait on the longer of a duration and a count", group[6].impl.(*APLActionWait).duration.GetDuration(sim), 2 * time.Second},
		{"Wait on a duration times a duration", group[7].impl.(*APLActionWait).duration.GetDuration(sim), time.Duration(0)},
		{"Wait on an empty variable times a duration", group[8].impl.(*APLActionWait).duration.GetDuration(sim), time.Duration(0)},
	} {
		if tc.got != tc.want {
			t.Errorf("%s reads %v, want %v", tc.name, tc.got, tc.want)
		}
	}

	wantWarnings(t, "the filled reference", rot.priorityListValidations[0], "Cannot read a yes/no value as a duration", "variable 'blank'", "variable 'blank'", "Duration * Duration", "variable 'blank'")
	if validations := rot.priorityListValidations[1]; len(validations) != 1 || validations[0].LogLevel != proto.LogLevel_Error || !strings.Contains(validations[0].Validation, "placeholder 'missing'") {
		t.Errorf("the unfilled reference: validations = %v, want the error naming 'missing'", validations)
	}
}

// A placeholder in Aura Should Refresh's overlap is filled like one in any other value.
func TestGroupFillsAuraShouldRefreshOverlap(t *testing.T) {
	_, _, rot := setupFakeRageRotation()
	overlap := rot.coerceTo(&APLValueVariablePlaceholder{name: "time"}, proto.APLValueType_ValueTypeDuration)
	filled := (&APLActionGroupReference{}).replacePlaceholders(&APLValueAuraShouldRefresh{maxOverlap: overlap}, map[string]*proto.APLValue{"time": constValue("2s")}, rot)
	if got := filled.(*APLValueAuraShouldRefresh).maxOverlap.GetDuration(nil); got != 2*time.Second {
		t.Errorf("the overlap reads %v, want 2s", got)
	}
}

func TestValueEncounter(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()
	sim.Duration = 100 * time.Second
	sim.CurrentTime = 25 * time.Second

	currentTime := rot.newValueCurrentTime(&proto.APLValueCurrentTime{}, nil)
	currentTimePercent := rot.newValueCurrentTimePercent(&proto.APLValueCurrentTimePercent{}, nil)
	remainingTime := rot.newValueRemainingTime(&proto.APLValueRemainingTime{}, nil)
	remainingTimePercent := rot.newValueRemainingTimePercent(&proto.APLValueRemainingTimePercent{}, nil)
	numberTargets := rot.newValueNumberTargets(&proto.APLValueNumberTargets{}, nil)

	if got := currentTime.GetDuration(sim); got != 25*time.Second {
		t.Fatalf("Current Time = %s, want 25s", got)
	}
	if got := currentTimePercent.GetFloat(sim); got != 0.25 {
		t.Fatalf("Current Time Percent = %f, want 0.25", got)
	}
	if got := remainingTime.GetDuration(sim); got != 75*time.Second {
		t.Fatalf("Remaining Time = %s, want 75s", got)
	}
	if got := remainingTimePercent.GetFloat(sim); got != 0.75 {
		t.Fatalf("Remaining Time Percent = %f, want 0.75", got)
	}
	if numberTargets.GetInt(sim) != 1 || numberTargets.GetFloat(sim) != 1 {
		t.Fatalf("Number of Targets = %d, want 1", numberTargets.GetInt(sim))
	}

	thresholds := []struct {
		threshold proto.APLValueIsExecutePhase_ExecutePhaseThreshold
		phase     int32
		want      bool
	}{
		{proto.APLValueIsExecutePhase_E20, 20, true},
		{proto.APLValueIsExecutePhase_E20, 35, false},
		{proto.APLValueIsExecutePhase_E35, 20, true},
		{proto.APLValueIsExecutePhase_E35, 35, true},
		{proto.APLValueIsExecutePhase_E35, 0, false},
	}
	for _, tt := range thresholds {
		value := rot.newValueIsExecutePhase(&proto.APLValueIsExecutePhase{Threshold: tt.threshold}, nil)
		rot.unit.CurrentTarget.executePhase = tt.phase
		if got := value.GetBool(sim); got != tt.want {
			t.Errorf("Is Execute Phase %v in the %d phase = %t, want %t", tt.threshold, tt.phase, got, tt.want)
		}
	}
	wantNoValue(t, map[string]APLValue{
		"Is Execute Phase with an unknown threshold":         rot.newValueIsExecutePhase(&proto.APLValueIsExecutePhase{}, nil),
		"Is Execute Phase with a threshold outside the enum": rot.newValueIsExecutePhase(&proto.APLValueIsExecutePhase{Threshold: 4}, nil),
	})
}

func TestValueTargetIsMobTypeAndEncounterInArea(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()
	sim.Encounter.AreaTypes = []proto.AreaType{proto.AreaType_AreaTypeHaunted}

	isMobType := func(mobType proto.MobType, targetUnit *proto.UnitReference) APLValue {
		return rot.newValueTargetIsMobType(&proto.APLValueTargetIsMobType{TargetUnit: targetUnit, MobType: mobType}, nil)
	}
	firstTarget := &proto.UnitReference{Type: proto.UnitReference_Target, Index: 0}
	for _, tc := range []struct {
		mobType    proto.MobType
		targetUnit *proto.UnitReference
		want       bool
	}{
		{proto.MobType_MobTypeHumanoid, nil, true},
		{proto.MobType_MobTypeDemon, nil, false},
		{proto.MobType_MobTypeUnknown, nil, false},
		{proto.MobType_MobTypeHumanoid, firstTarget, true},
		{proto.MobType_MobTypeUndead, firstTarget, false},
	} {
		if got := isMobType(tc.mobType, tc.targetUnit).GetBool(sim); got != tc.want {
			t.Errorf("Target Is Mob Type(%v, %v) = %t, want %t", tc.targetUnit, tc.mobType, got, tc.want)
		}
	}

	inHaunted := rot.newValueEncounterInArea(&proto.APLValueEncounterInArea{AreaType: proto.AreaType_AreaTypeHaunted}, nil)
	inDesert := rot.newValueEncounterInArea(&proto.APLValueEncounterInArea{AreaType: proto.AreaType_AreaTypeDesert}, nil)
	if !inHaunted.GetBool(sim) || inDesert.GetBool(sim) {
		t.Errorf("Encounter In Area: haunted = %t, desert = %t, want true, false", inHaunted.GetBool(sim), inDesert.GetBool(sim))
	}

	wantNoValue(t, map[string]APLValue{
		"Target Is Mob Type for a missing target":    isMobType(proto.MobType_MobTypeHumanoid, &proto.UnitReference{Type: proto.UnitReference_Target, Index: 5}),
		"Encounter In Area for an unknown area type": rot.newValueEncounterInArea(&proto.APLValueEncounterInArea{}, nil),
	})
}

func TestValueBoss(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	target := sim.Encounter.AllTargetUnits[0]
	bossSpellID := ActionID{SpellID: 424242}
	bossSpell := target.RegisterSpell(SpellConfig{
		ActionID: bossSpellID,
		Cast: CastConfig{
			CD: Cooldown{Timer: target.NewTimer(), Duration: 10 * time.Second},
		},
	})
	id := bossSpellID.ToProto()
	bossRef := &proto.UnitReference{Type: proto.UnitReference_Target, Index: 0}

	isCasting := rot.newValueBossSpellIsCasting(&proto.APLValueBossSpellIsCasting{SpellId: id}, nil)
	timeToReady := rot.newValueBossSpellTimeToReady(&proto.APLValueBossSpellTimeToReady{SpellId: id}, nil)
	currentTarget := rot.newValueBossCurrentTarget(&proto.APLValueBossCurrentTarget{TargetUnit: bossRef}, nil)
	if isCasting == nil || timeToReady == nil || currentTarget == nil {
		t.Fatalf("Expected the boss values to resolve a spell registered on the target")
	}

	if isCasting.GetBool(sim) {
		t.Fatalf("Boss Spell Is Casting before any cast = true, want false")
	}
	target.Hardcast = Hardcast{ActionID: bossSpellID, Expires: 3 * time.Second}
	if !isCasting.GetBool(sim) {
		t.Fatalf("Boss Spell Is Casting during its hardcast = false, want true")
	}
	sim.CurrentTime = 3 * time.Second
	if isCasting.GetBool(sim) {
		t.Fatalf("Boss Spell Is Casting once the hardcast expired = true, want false")
	}

	if got := timeToReady.GetDuration(sim); got != 0 {
		t.Fatalf("Boss Spell Time To Ready off cooldown = %s, want 0s", got)
	}
	bossSpell.CD.Set(sim.CurrentTime + 10*time.Second)
	if got := timeToReady.GetDuration(sim); got != 10*time.Second {
		t.Fatalf("Boss Spell Time To Ready on cooldown = %s, want 10s", got)
	}

	target.CurrentTarget = &fw.Unit
	if !currentTarget.GetBool(sim) {
		t.Fatalf("Boss Current Target while the boss targets the player = false, want true")
	}
	target.CurrentTarget = target
	if currentTarget.GetBool(sim) {
		t.Fatalf("Boss Current Target while the boss targets another unit = true, want false")
	}
}

func TestValueHealthAndMana(t *testing.T) {
	sim := &Simulation{}
	target := &Unit{}
	target.healthBar = healthBar{unit: target, currentHealth: 250}
	target.stats[stats.Health] = 1000
	target.manaBar = manaBar{unit: target, currentMana: 300}
	target.stats[stats.Mana] = 1200
	target.manaRegenWhileCasting = 7

	rot := &APLRotation{unit: &Unit{CurrentTarget: target}}
	targetRef := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}

	if got := rot.newValueCurrentHealth(&proto.APLValueCurrentHealth{SourceUnit: targetRef}, nil).GetFloat(sim); got != 250 {
		t.Errorf("Current Health of target = %v, want 250", got)
	}
	if got := rot.newValueCurrentHealthPercent(&proto.APLValueCurrentHealthPercent{SourceUnit: targetRef}, nil).GetFloat(sim); got != 0.25 {
		t.Errorf("Current Health %% of target = %v, want 0.25", got)
	}
	if got := rot.newValueCurrentMana(&proto.APLValueCurrentMana{SourceUnit: targetRef}, nil).GetFloat(sim); got != 300 {
		t.Errorf("Current Mana of target = %v, want 300", got)
	}
	if got := rot.newValueCurrentManaPercent(&proto.APLValueCurrentManaPercent{SourceUnit: targetRef}, nil).GetFloat(sim); got != 0.25 {
		t.Errorf("Current Mana %% of target = %v, want 0.25", got)
	}

	targetRot := &APLRotation{unit: target}
	if got := targetRot.newValueMaxHealth(&proto.APLValueMaxHealth{}, nil).GetFloat(sim); got != 1000 {
		t.Errorf("Max Health = %v, want 1000", got)
	}
	if got := targetRot.newValueMaxMana(&proto.APLValueMaxMana{}, nil).GetFloat(sim); got != 1200 {
		t.Errorf("Max Mana = %v, want 1200", got)
	}
	if got := targetRot.newValueCurrentManaRegen(&proto.APLValueCurrentManaRegen{}, nil).GetFloat(sim); got != 7 {
		t.Errorf("Current Mana Regen = %v, want 7", got)
	}

	wantNoValue(t, map[string]APLValue{
		"Current Health without the bar":   rot.newValueCurrentHealth(&proto.APLValueCurrentHealth{}, nil),
		"Current Health % without the bar": rot.newValueCurrentHealthPercent(&proto.APLValueCurrentHealthPercent{}, nil),
		"Max Health without the bar":       rot.newValueMaxHealth(&proto.APLValueMaxHealth{}, nil),
		"Current Mana without the bar":     rot.newValueCurrentMana(&proto.APLValueCurrentMana{}, nil),
		"Current Mana % without the bar":   rot.newValueCurrentManaPercent(&proto.APLValueCurrentManaPercent{}, nil),
		"Max Mana without the bar":         rot.newValueMaxMana(&proto.APLValueMaxMana{}, nil),
		"Mana Regen without the bar":       rot.newValueCurrentManaRegen(&proto.APLValueCurrentManaRegen{}, nil),
	})
}

func TestValueRageAndEnergyAndComboPoints(t *testing.T) {
	sim := &Simulation{CurrentTime: 30 * time.Millisecond}
	unit := &Unit{}
	unit.rageBar = rageBar{unit: unit, currentRage: 40, maxRage: 100}
	unit.energyBar = energyBar{
		unit:           unit,
		currentEnergy:  60,
		maxEnergy:      110,
		comboPoints:    3,
		maxComboPoints: 5,
		nextEnergyTick: 100 * time.Millisecond,
	}
	rot := &APLRotation{unit: unit}

	if got := rot.newValueCurrentRage(&proto.APLValueCurrentRage{}, nil).GetFloat(sim); got != 40 {
		t.Errorf("Current Rage = %v, want 40", got)
	}
	if got := rot.newValueMaxRage(&proto.APLValueMaxRage{}, nil).GetFloat(sim); got != 100 {
		t.Errorf("Max Rage = %v, want 100", got)
	}
	if got := rot.newValueCurrentEnergy(&proto.APLValueCurrentEnergy{}, nil).GetFloat(sim); got != 60 {
		t.Errorf("Current Energy = %v, want 60", got)
	}
	if got := rot.newValueMaxEnergy(&proto.APLValueMaxEnergy{}, nil).GetFloat(sim); got != 110 {
		t.Errorf("Max Energy = %v, want 110", got)
	}
	if got := rot.newValueCurrentComboPoints(&proto.APLValueCurrentComboPoints{}, nil).GetInt(sim); got != 3 {
		t.Errorf("Current Combo Points = %v, want 3", got)
	}
	if got := rot.newValueMaxComboPoints(&proto.APLValueMaxComboPoints{}, nil).GetInt(sim); got != 5 {
		t.Errorf("Max Combo Points = %v, want 5", got)
	}
	if got := rot.newValueTimeToNextEnergyTick(&proto.APLValueTimeToNextEnergyTick{}, nil).GetDuration(sim); got != 70*time.Millisecond {
		t.Errorf("Time To Next Energy Tick = %v, want 70ms", got)
	}

	empty := &APLRotation{unit: &Unit{}}
	wantNoValue(t, map[string]APLValue{
		"Current Rage without the bar":         empty.newValueCurrentRage(&proto.APLValueCurrentRage{}, nil),
		"Max Rage without the bar":             empty.newValueMaxRage(&proto.APLValueMaxRage{}, nil),
		"Current Energy without the bar":       empty.newValueCurrentEnergy(&proto.APLValueCurrentEnergy{}, nil),
		"Max Energy without the bar":           empty.newValueMaxEnergy(&proto.APLValueMaxEnergy{}, nil),
		"Current Combo Points without the bar": empty.newValueCurrentComboPoints(&proto.APLValueCurrentComboPoints{}, nil),
		"Max Combo Points without the bar":     empty.newValueMaxComboPoints(&proto.APLValueMaxComboPoints{}, nil),
		"Next Energy Tick without the bar":     empty.newValueTimeToNextEnergyTick(&proto.APLValueTimeToNextEnergyTick{}, nil),
	})
}

func TestValueEnergyTimeToTarget(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	config := func(target string) *proto.APLValueEnergyTimeToTarget {
		return &proto.APLValueEnergyTimeToTarget{TargetEnergy: constValue(target)}
	}

	wantNoValue(t, map[string]APLValue{
		"Energy Time To Target without an energy bar": rot.newValueEnergyTimeToTarget(config("50"), nil),
	})

	fw.EnableEnergyBar(EnergyBarOptions{MaxEnergy: 100, MaxComboPoints: 5})
	wantNoValue(t, map[string]APLValue{
		"Energy Time To Target without a target energy": rot.newValueEnergyTimeToTarget(&proto.APLValueEnergyTimeToTarget{}, nil),
	})

	sim.CurrentTime = 0
	fw.energyBar.currentEnergy = 40
	fw.energyBar.nextEnergyTick = 50 * time.Millisecond

	for _, c := range []struct {
		target string
		want   time.Duration
	}{
		{"30", 0},
		{"43", 250 * time.Millisecond},
		{"150", NeverExpires},
	} {
		value := rot.newValueEnergyTimeToTarget(config(c.target), nil)
		if got := value.GetDuration(sim); got != c.want {
			t.Errorf("Energy Time To Target %s from 40 energy = %v, want %v", c.target, got, c.want)
		}
	}
}

func TestValueGCD(t *testing.T) {
	sim := &Simulation{}
	unit := &Unit{GCD: new(Timer)}
	rot := &APLRotation{unit: unit}
	isReady := rot.newValueGCDIsReady(&proto.APLValueGCDIsReady{}, nil)
	timeToReady := rot.newValueGCDTimeToReady(&proto.APLValueGCDTimeToReady{}, nil)

	for _, c := range []struct {
		name      string
		now       time.Duration
		readyAt   time.Duration
		wantReady bool
		wantTime  time.Duration
	}{
		{"off cooldown", time.Second, 500 * time.Millisecond, true, 0},
		{"inside the queue window", time.Second, time.Second + MaxSpellQueueWindow, true, MaxSpellQueueWindow},
		{"outside the queue window", time.Second, time.Second + MaxSpellQueueWindow + time.Millisecond, false, MaxSpellQueueWindow + time.Millisecond},
	} {
		sim.CurrentTime = c.now
		*unit.GCD = Timer(c.readyAt)
		if got := isReady.GetBool(sim); got != c.wantReady {
			t.Errorf("GCD Is Ready, %s = %t, want %t", c.name, got, c.wantReady)
		}
		if got := timeToReady.GetDuration(sim); got != c.wantTime {
			t.Errorf("GCD Time To Ready, %s = %v, want %v", c.name, got, c.wantTime)
		}
	}
}

func TestValueAutoAttacks(t *testing.T) {
	sim := &Simulation{CurrentTime: 10 * time.Second}
	unit := &Unit{}
	aa := &unit.AutoAttacks
	aa.mh = WeaponAttack{enabled: true, swingAt: 12 * time.Second, previousSwing: 9 * time.Second, curSwingDuration: 3 * time.Second}
	aa.oh = WeaponAttack{enabled: true, swingAt: 11 * time.Second, previousSwing: 8 * time.Second, curSwingDuration: 2 * time.Second}
	aa.ranged = WeaponAttack{enabled: true, swingAt: 15 * time.Second, previousSwing: 9500 * time.Millisecond, curSwingDuration: 4 * time.Second}
	rot := &APLRotation{unit: unit}

	for _, c := range []struct {
		swing proto.APLValueAutoSwingTime_SwingType
		want  time.Duration
	}{
		{proto.APLValueAutoSwingTime_MainHand, 3 * time.Second},
		{proto.APLValueAutoSwingTime_OffHand, 2 * time.Second},
		{proto.APLValueAutoSwingTime_Ranged, 4 * time.Second},
	} {
		got := rot.newValueAutoSwingTime(&proto.APLValueAutoSwingTime{AutoType: c.swing}, nil).GetDuration(sim)
		if got != c.want {
			t.Errorf("Auto Swing Time %v = %v, want %v", c.swing, got, c.want)
		}
	}

	for _, c := range []struct {
		auto          proto.APLValueAutoAttackType
		wantToNext    time.Duration
		wantSinceLast time.Duration
	}{
		{proto.APLValueAutoAttackType_MeleeAuto, time.Second, 1 * time.Second},
		{proto.APLValueAutoAttackType_MainHandAuto, 2 * time.Second, time.Second},
		{proto.APLValueAutoAttackType_OffHandAuto, time.Second, 2 * time.Second},
		{proto.APLValueAutoAttackType_RangedAuto, 5 * time.Second, 500 * time.Millisecond},
	} {
		toNext := rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{AutoType: c.auto}, nil).GetDuration(sim)
		if toNext != c.wantToNext {
			t.Errorf("Auto Time To Next %v = %v, want %v", c.auto, toNext, c.wantToNext)
		}
		sinceLast := rot.newValueAutoTimeSinceLast(&proto.APLValueAutoTimeSinceLast{AutoType: c.auto}, nil).GetDuration(sim)
		if sinceLast != c.wantSinceLast {
			t.Errorf("Auto Time Since Last %v = %v, want %v", c.auto, sinceLast, c.wantSinceLast)
		}
	}

	anyToNext := rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoAttackType_AnyAuto}, nil).GetDuration(sim)
	anySinceLast := rot.newValueAutoTimeSinceLast(&proto.APLValueAutoTimeSinceLast{AutoType: proto.APLValueAutoAttackType_AnyAuto}, nil).GetDuration(sim)
	if anyToNext != time.Second || anySinceLast != 500*time.Millisecond {
		t.Errorf("Any auto: time to next %v, since last %v, want 1s, 500ms", anyToNext, anySinceLast)
	}

	sim.CurrentTime = 20 * time.Second
	if got := rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoAttackType_MainHandAuto}, nil).GetDuration(sim); got != 0 {
		t.Errorf("Auto Time To Next past the swing time = %v, want 0", got)
	}
}

func TestValueUnitDistanceAndMoving(t *testing.T) {
	sim := &Simulation{}
	rot := &APLRotation{unit: &Unit{DistanceFromTarget: 12.5, CurrentTarget: &Unit{Moving: true}}}
	targetRef := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}

	if got := rot.newValueUnitDistance(&proto.APLValueUnitDistance{}, nil).GetFloat(sim); got != 12.5 {
		t.Errorf("Unit Distance = %v, want 12.5", got)
	}

	selfMoving := rot.newValueUnitIsMoving(&proto.APLValueUnitIsMoving{}, nil)
	targetMoving := rot.newValueUnitIsMoving(&proto.APLValueUnitIsMoving{SourceUnit: targetRef}, nil)
	if selfMoving.GetBool(sim) || !targetMoving.GetBool(sim) {
		t.Errorf("Unit Is Moving: self = %t, target = %t, want false, true", selfMoving.GetBool(sim), targetMoving.GetBool(sim))
	}

	noTarget := &APLRotation{unit: &Unit{}}
	wantNoValue(t, map[string]APLValue{
		"Unit Is Moving without a current target": noTarget.newValueUnitIsMoving(&proto.APLValueUnitIsMoving{SourceUnit: targetRef}, nil),
		"Movement Speed without a current target": noTarget.newValueUnitMovementSpeed(&proto.APLValueUnitMovementSpeed{SourceUnit: targetRef}, nil),
	})
}

func TestValueUnitMovementSpeed(t *testing.T) {
	sim := &Simulation{}
	player := &Unit{Type: PlayerUnit}
	player.PseudoStats.MovementSpeedMultiplier = 1.3
	target := &Unit{Type: EnemyUnit}
	target.PseudoStats.MovementSpeedMultiplier = 1
	player.CurrentTarget = target
	rot := &APLRotation{unit: player}

	if got := rot.newValueUnitMovementSpeed(&proto.APLValueUnitMovementSpeed{}, nil).GetFloat(sim); math.Abs(got-9.1) > 1e-9 {
		t.Errorf("Movement Speed of a player at +30%% = %v, want 9.1", got)
	}
	targetRef := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}
	if got := rot.newValueUnitMovementSpeed(&proto.APLValueUnitMovementSpeed{SourceUnit: targetRef}, nil).GetFloat(sim); got != 8 {
		t.Errorf("Movement Speed of the target = %v, want 8", got)
	}
}

func TestValueUnknownIDs(t *testing.T) {
	_, _, rot := setupFakeRageRotation()
	spell := ActionID{SpellID: 999999}.ToProto()
	aura := ActionID{SpellID: testMissingAuraID}.ToProto()

	wantNoValue(t, map[string]APLValue{
		"Spell Can Cast for an unknown spell":            rot.newValueSpellCanCast(&proto.APLValueSpellCanCast{SpellId: spell}, nil),
		"Spell Is Ready for an unknown spell":            rot.newValueSpellIsReady(&proto.APLValueSpellIsReady{SpellId: spell}, nil),
		"Spell Time To Ready for an unknown spell":       rot.newValueSpellTimeToReady(&proto.APLValueSpellTimeToReady{SpellId: spell}, nil),
		"Spell Cast Time for an unknown spell":           rot.newValueSpellCastTime(&proto.APLValueSpellCastTime{SpellId: spell}, nil),
		"Spell Travel Time for an unknown spell":         rot.newValueSpellTravelTime(&proto.APLValueSpellTravelTime{SpellId: spell}, nil),
		"Spell CPM for an unknown spell":                 rot.newValueSpellCPM(&proto.APLValueSpellCPM{SpellId: spell}, nil),
		"Spell Is Casting for an unknown spell":          rot.newValueSpellIsCasting(&proto.APLValueSpellIsCasting{SpellId: spell}, nil),
		"Spell Is Channeling for an unknown spell":       rot.newValueSpellIsChanneling(&proto.APLValueSpellIsChanneling{SpellId: spell}, nil),
		"Spell Channeled Ticks for an unknown spell":     rot.newValueSpellChanneledTicks(&proto.APLValueSpellChanneledTicks{SpellId: spell}, nil),
		"Spell Current Cost for an unknown spell":        rot.newValueSpellCurrentCost(&proto.APLValueSpellCurrentCost{SpellId: spell}, nil),
		"Spell Num Charges for an unknown spell":         rot.newValueSpellNumCharges(&proto.APLValueSpellNumCharges{SpellId: spell}, nil),
		"Spell Time To Charge for an unknown spell":      rot.newValueSpellTimeToCharge(&proto.APLValueSpellTimeToCharge{SpellId: spell}, nil),
		"Spell GCD Hasted Duration for an unknown spell": rot.newValueSpellGCDHastedDuration(&proto.APLValueSpellGCDHastedDuration{SpellId: spell}, nil),
		"Spell Full Cooldown for an unknown spell":       rot.newValueSpellFullCooldown(&proto.APLValueSpellFullCooldown{SpellId: spell}, nil),
		"Spell In Flight for an unknown spell":           rot.newValueSpellInFlight(&proto.APLValueSpellInFlight{SpellId: spell}, nil),
		"exported Spell In Flight for an unknown spell":  rot.NewValueSpellInFlight(&proto.APLValueSpellInFlight{SpellId: spell}, nil),
		"Spell Is Hit Capped for an unknown spell":       rot.newValueSpellIsHitCapped(&proto.APLValueSpellIsHitCapped{SpellId: spell}, nil),
		"Spell Is Expertise Capped for an unknown spell": rot.newValueSpellIsExpertiseCapped(&proto.APLValueSpellIsExpertiseCapped{SpellId: spell}, nil),

		"Boss Spell Is Casting for a spell the boss does not know":    rot.newValueBossSpellIsCasting(&proto.APLValueBossSpellIsCasting{SpellId: spell}, nil),
		"Boss Spell Time To Ready for a spell the boss does not know": rot.newValueBossSpellTimeToReady(&proto.APLValueBossSpellTimeToReady{SpellId: spell}, nil),

		"Dot Is Active for an unknown spell":                rot.newValueDotIsActive(&proto.APLValueDotIsActive{SpellId: spell}, nil),
		"Dot Is Active On All Targets for an unknown spell": rot.newValueDotIsActiveOnAllTargets(&proto.APLValueDotIsActiveOnAllTargets{SpellId: spell}, nil),
		"Dot Remaining Time for an unknown spell":           rot.newValueDotRemainingTime(&proto.APLValueDotRemainingTime{SpellId: spell}, nil),
		"Dot Lowest Remaining Time for an unknown spell":    rot.newValueDotLowestRemainingTime(&proto.APLValueDotLowestRemainingTime{SpellId: spell}, nil),
		"Dot Tick Frequency for an unknown spell":           rot.newValueDotTickFrequency(&proto.APLValueDotTickFrequency{SpellId: spell}, nil),
		"Dot Time To Next Tick for an unknown spell":        rot.newValueDotTimeToNextTick(&proto.APLValueDotTimeToNextTick{SpellId: spell}, nil),
		"Dot Base Duration for an unknown spell":            rot.newValueDotBaseDuration(&proto.APLValueDotBaseDuration{SpellId: spell}, nil),

		"Aura Is Active without an aura id":      rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{}, nil),
		"Aura Is Inactive without an aura id":    rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{}, nil),
		"Aura Remaining Time without an aura id": rot.newValueAuraRemainingTime(&proto.APLValueAuraRemainingTime{}, nil),
		"Aura Duration without an aura id":       rot.newValueAuraDuration(&proto.APLValueAuraDuration{}, nil),

		"Aura Is Active for a missing aura":         rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{AuraId: aura}, nil),
		"Aura Is Inactive for a missing aura":       rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{AuraId: aura}, nil),
		"Aura Remaining Time for a missing aura":    rot.newValueAuraRemainingTime(&proto.APLValueAuraRemainingTime{AuraId: aura}, nil),
		"Aura Duration for a missing aura":          rot.newValueAuraDuration(&proto.APLValueAuraDuration{AuraId: aura}, nil),
		"Aura Num Stacks for a missing aura":        rot.newValueAuraNumStacks(&proto.APLValueAuraNumStacks{AuraId: aura}, nil),
		"Aura Internal Cooldown for a missing aura": rot.newValueAuraInternalCooldown(&proto.APLValueAuraInternalCooldown{AuraId: aura}, nil),
		"Aura ICD Is Ready for a missing aura":      rot.newValueAuraICDIsReady(&proto.APLValueAuraICDIsReady{AuraId: aura}, nil),
		"Aura Should Refresh for a missing aura":    rot.newValueAuraShouldRefresh(&proto.APLValueAuraShouldRefresh{AuraId: aura}, nil),
	})
}

func TestValueSpellIsKnown(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()

	known := rot.newValueSpellIsKnown(&proto.APLValueSpellIsKnown{SpellId: fakeRendActionID.ToProto()}, nil)
	unknown := rot.newValueSpellIsKnown(&proto.APLValueSpellIsKnown{SpellId: ActionID{SpellID: 1}.ToProto()}, nil)
	if !known.GetBool(sim) || unknown.GetBool(sim) {
		t.Fatalf("Is Known: Rend = %t, unknown = %t, want true, false", known.GetBool(sim), unknown.GetBool(sim))
	}
}

// Three WarriorTalents fields picked by kind, with the spell ids their talent_spell_id options carry:
// the first one-point talent with its point, the first ranked talent at 3 points, and the next ranked
// talent with none. A talents string with every field in one tree sets field N from digit N.
type testTalents struct {
	single, ranked, untaken int32
	fill                    func(char *Character)
}

func newTestTalents(t *testing.T) testTalents {
	t.Helper()
	fields := (&proto.WarriorTalents{}).ProtoReflect().Descriptor().Fields()
	digits := []byte(strings.Repeat("0", fields.Len()))
	spellID := func(fd protoreflect.FieldDescriptor) int32 {
		return googleproto.GetExtension(fd.Options(), proto.E_TalentSpellId).(int32)
	}

	var tt testTalents
	for i := range fields.Len() {
		fd := fields.Get(i)
		switch {
		case fd.Kind() == protoreflect.BoolKind && tt.single == 0:
			tt.single = spellID(fd)
			digits[fd.Number()-1] = '1'
		case fd.Kind() == protoreflect.Int32Kind && tt.ranked == 0:
			tt.ranked = spellID(fd)
			digits[fd.Number()-1] = '3'
		case fd.Kind() == protoreflect.Int32Kind && tt.untaken == 0:
			tt.untaken = spellID(fd)
		}
	}
	if tt.single == 0 || tt.ranked == 0 || tt.untaken == 0 {
		t.Fatalf("WarriorTalents lacks a one-point talent or two ranked talents with talent_spell_id: %+v", tt)
	}

	tt.fill = func(char *Character) {
		char.FillTalentsProto((&proto.WarriorTalents{}).ProtoReflect(), string(digits), [3]int{fields.Len(), 0, 0})
	}
	return tt
}

func TestValueTalentIsKnown(t *testing.T) {
	talents := newTestTalents(t)
	sim, _, rot := setupFakeRageRotationWith(talents.fill)

	talent := func(spellID int32, rank int32) *proto.ActionID {
		id := ActionID{SpellID: spellID}.ToProto()
		id.Rank = rank
		return id
	}
	for _, tc := range []struct {
		name string
		id   *proto.ActionID
		want bool
	}{
		{"one-point talent with no rank", talent(talents.single, 0), true},
		{"one-point talent at its rank", talent(talents.single, 1), true},
		{"ranked talent below the held rank", talent(talents.ranked, 1), true},
		{"ranked talent at the held rank", talent(talents.ranked, 3), true},
		{"ranked talent above the held rank", talent(talents.ranked, 4), false},
		{"talent with no points", talent(talents.untaken, 1), false},
	} {
		rot.curValidations = nil
		value := rot.newValueTalentIsKnown(&proto.APLValueTalentIsKnown{TalentId: tc.id}, nil)
		if got := value.GetBool(sim); got != tc.want {
			t.Errorf("%s: got %t, want %t", tc.name, got, tc.want)
		}
		if warned := len(rot.curValidations) > 0; warned == tc.want {
			t.Errorf("%s: warned %t, want a warning only for an unknown talent", tc.name, warned)
		}
	}
	rot.curValidations = nil
}

func TestValueTalentRank(t *testing.T) {
	talents := newTestTalents(t)
	sim, _, rot := setupFakeRageRotationWith(talents.fill)

	for _, tc := range []struct {
		name    string
		spellID int32
		want    int32
	}{
		{"one-point talent", talents.single, 1},
		{"ranked talent", talents.ranked, 3},
		{"talent with no points", talents.untaken, 0},
	} {
		rot.curValidations = nil
		// The id's rank is not read, so one above the held rank changes nothing.
		id := ActionID{SpellID: tc.spellID}.ToProto()
		id.Rank = 5
		value := rot.newValueTalentRank(&proto.APLValueTalentRank{TalentId: id}, nil)
		if value.Type() != proto.APLValueType_ValueTypeInt || value.GetInt(sim) != tc.want {
			t.Errorf("%s: got %s, want %d", tc.name, renderValue(sim, value), tc.want)
		}
		if warned := len(rot.curValidations) > 0; warned != (tc.want == 0) {
			t.Errorf("%s: warned %t, want a warning only for a talent with no points", tc.name, warned)
		}
	}
	rot.curValidations = nil
}

func TestPlayerMetadataTalents(t *testing.T) {
	talents := newTestTalents(t)
	_, fw, _ := setupFakeRageRotationWith(talents.fill)

	playerStats := &proto.PlayerStats{}
	fw.FillPlayerStats(playerStats)

	got := map[int32]int32{}
	for _, id := range playerStats.Metadata.Talents {
		got[id.GetSpellId()] = id.Rank
	}
	if want := map[int32]int32{talents.single: 1, talents.ranked: 3}; !maps.Equal(got, want) {
		t.Errorf("metadata talents = %v, want %v", got, want)
	}
}

func TestValueSpellCastState(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	target := sim.Encounter.ActiveTargetUnits[0]
	rend := fw.GetSpell(fakeRendActionID)
	cdSpell := fw.GetSpell(fakeCategorySpellActionIDs[0])
	rendID := fakeRendActionID.ToProto()
	cdID := fakeCategorySpellActionIDs[0].ToProto()

	canCast := rot.newValueSpellCanCast(&proto.APLValueSpellCanCast{SpellId: rendID}, nil)
	if canCast.GetBool(sim) {
		t.Errorf("Can Cast(Rend) without Rage: got true, want false")
	}
	fw.AddRage(sim, 50, rend.Cost.ResourceCostImpl.(*RageCost).ResourceMetrics)
	if !canCast.GetBool(sim) {
		t.Errorf("Can Cast(Rend) with 50 Rage: got false, want true")
	}

	cost := rot.newValueSpellCurrentCost(&proto.APLValueSpellCurrentCost{SpellId: rendID}, nil)
	if cost.GetFloat(sim) != 30 {
		t.Errorf("Current Cost(Rend) = %f, want 30", cost.GetFloat(sim))
	}
	noCost := rot.newValueSpellCurrentCost(&proto.APLValueSpellCurrentCost{SpellId: cdID}, nil)
	if noCost.GetFloat(sim) != 0 {
		t.Errorf("Current Cost of a spell without a cost = %f, want 0", noCost.GetFloat(sim))
	}

	isReady := rot.newValueSpellIsReady(&proto.APLValueSpellIsReady{SpellId: cdID}, nil)
	timeToReady := rot.newValueSpellTimeToReady(&proto.APLValueSpellTimeToReady{SpellId: cdID}, nil)
	fullCD := rot.newValueSpellFullCooldown(&proto.APLValueSpellFullCooldown{SpellId: cdID}, nil)
	if !isReady.GetBool(sim) || timeToReady.GetDuration(sim) != 0 {
		t.Errorf("before the cast: Is Ready = %t, Time To Ready = %s, want true, 0s", isReady.GetBool(sim), timeToReady.GetDuration(sim))
	}
	if !cdSpell.Cast(sim, target) {
		t.Fatal("the cooldown spell did not cast")
	}
	if isReady.GetBool(sim) || timeToReady.GetDuration(sim) != time.Minute || timeToReady.GetFloat(sim) != 60 {
		t.Errorf("after the cast: Is Ready = %t, Time To Ready = %s, want false, 1m0s", isReady.GetBool(sim), timeToReady.GetDuration(sim))
	}
	if fullCD.GetDuration(sim) != time.Minute {
		t.Errorf("Full Cooldown = %s, want 1m0s", fullCD.GetDuration(sim))
	}
	rend.SharedCD.Duration = 120 * time.Second
	sharedFullCD := rot.newValueSpellFullCooldown(&proto.APLValueSpellFullCooldown{SpellId: rendID}, nil)
	if sharedFullCD.GetDuration(sim) != 120*time.Second {
		t.Errorf("Full Cooldown of a spell with only a 120s shared cooldown = %s, want 2m0s", sharedFullCD.GetDuration(sim))
	}
	rend.SharedCD.Duration = 0

	sim.CurrentTime = 90 * time.Second
	cdSpell.casts = 3
	cpm := rot.newValueSpellCPM(&proto.APLValueSpellCPM{SpellId: cdID}, nil)
	if cpm.GetFloat(sim) != 2 {
		t.Errorf("CPM after 3 casts in 1.5 minutes = %f, want 2", cpm.GetFloat(sim))
	}

	castTime := rot.newValueSpellCastTime(&proto.APLValueSpellCastTime{SpellId: rendID}, nil)
	if castTime.GetDuration(sim) != 0 {
		t.Errorf("Cast Time(Rend) = %s, want 0s", castTime.GetDuration(sim))
	}

	travel := rot.newValueSpellTravelTime(&proto.APLValueSpellTravelTime{SpellId: rendID}, nil)
	if travel.GetDuration(sim) != 0 {
		t.Errorf("Travel Time without a missile = %s, want 0s", travel.GetDuration(sim))
	}
	rend.MissileSpeed = 20
	fw.DistanceFromTarget = 40
	if travel.GetDuration(sim) != 2*time.Second {
		t.Errorf("Travel Time at 20 speed over 40 yards = %s, want 2s", travel.GetDuration(sim))
	}
}

func TestValueSpellGCDHastedDuration(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	spell := fw.GetSpell(fakeGCDSpellActionID)
	value := rot.newValueSpellGCDHastedDuration(&proto.APLValueSpellGCDHastedDuration{SpellId: fakeGCDSpellActionID.ToProto()}, nil)

	if value.GetDuration(sim) != GCDDefault || value.GetFloat(sim) != GCDDefault.Seconds() {
		t.Errorf("GCD without haste = %s, want %s", value.GetDuration(sim), GCDDefault)
	}

	spell.DefaultCast.GCD = 3 * time.Second
	spell.IgnoreHaste = true
	fw.MultiplyCastSpeed(sim, 2)
	if value.GetDuration(sim) != 3*time.Second {
		t.Errorf("GCD of a haste-ignoring spell = %s, want 3s", value.GetDuration(sim))
	}
	spell.IgnoreHaste = false
	if value.GetDuration(sim) != 1500*time.Millisecond {
		t.Errorf("GCD at double cast speed = %s, want 1.5s", value.GetDuration(sim))
	}
}

func TestValueSpellCharges(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	spell := fw.GetSpell(fakeGCDSpellActionID)
	id := fakeGCDSpellActionID.ToProto()
	numCharges := rot.newValueSpellNumCharges(&proto.APLValueSpellNumCharges{SpellId: id}, nil)
	timeToCharge := rot.newValueSpellTimeToCharge(&proto.APLValueSpellTimeToCharge{SpellId: id}, nil)

	spell.MaxCharges = 2
	spell.RechargeTime = 10 * time.Second
	spell.charges = 2
	if numCharges.GetInt(sim) != 2 || timeToCharge.GetDuration(sim) != 0 {
		t.Errorf("full charges: count = %d, time to charge = %s, want 2, 0s", numCharges.GetInt(sim), timeToCharge.GetDuration(sim))
	}

	spell.ConsumeCharge(sim)
	sim.CurrentTime = 4 * time.Second
	if numCharges.GetInt(sim) != 1 || timeToCharge.GetDuration(sim) != 6*time.Second || timeToCharge.GetFloat(sim) != 6 {
		t.Errorf("one charge spent 4s ago: count = %d, time to charge = %s, want 1, 6s", numCharges.GetInt(sim), timeToCharge.GetDuration(sim))
	}
}

func TestValueSpellChannelingAndFlight(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	rend := fw.GetSpell(fakeRendActionID)
	other := fw.GetSpell(fakeGCDSpellActionID)
	rendID := fakeRendActionID.ToProto()
	target := sim.Encounter.ActiveTargetUnits[0]

	isChanneling := rot.newValueSpellIsChanneling(&proto.APLValueSpellIsChanneling{SpellId: rendID}, nil)
	ticks := rot.newValueSpellChanneledTicks(&proto.APLValueSpellChanneledTicks{SpellId: rendID}, nil)
	inFlight := rot.newValueSpellInFlight(&proto.APLValueSpellInFlight{SpellId: rendID}, nil)

	if isChanneling.GetBool(sim) || ticks.GetInt(sim) != 0 || inFlight.GetBool(sim) {
		t.Errorf("idle: channeling = %t, ticks = %d, in flight = %t, want false, 0, false", isChanneling.GetBool(sim), ticks.GetInt(sim), inFlight.GetBool(sim))
	}

	dot := rend.Dot(target)
	dot.Apply(sim)
	fw.ChanneledDot = dot
	if !isChanneling.GetBool(sim) {
		t.Errorf("Is Channeling(Rend) while channeling Rend: got false, want true")
	}
	if ticks.GetInt(sim) != dot.TickCount() {
		t.Errorf("Channeled Ticks = %d, want %d", ticks.GetInt(sim), dot.TickCount())
	}
	fw.ChanneledDot = &Dot{Spell: other}
	if isChanneling.GetBool(sim) {
		t.Errorf("Is Channeling(Rend) while channeling another spell: got true, want false")
	}

	fw.SpellsInFlight = map[*Spell]int32{rend: 1}
	if !inFlight.GetBool(sim) {
		t.Errorf("Spell In Flight with a missile in the air: got false, want true")
	}
}

func TestValueSelectedConsumables(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	fw.Consumables.ResourcePotId = 100
	fw.Consumables.OffensivePotId = 102
	fw.Consumables.ConjuredId = 200

	selected := rot.newValueSelectedPotion(&proto.APLValueSelectedPotion{PotionId: ActionID{ItemID: 100}.ToProto()}, nil)
	selectedOffensive := rot.newValueSelectedPotion(&proto.APLValueSelectedPotion{PotionId: ActionID{ItemID: 102}.ToProto()}, nil)
	other := rot.newValueSelectedPotion(&proto.APLValueSelectedPotion{PotionId: ActionID{ItemID: 101}.ToProto()}, nil)
	if !selected.GetBool(sim) || !selectedOffensive.GetBool(sim) || other.GetBool(sim) {
		t.Errorf("Selected Potion: resource pick = %t, offensive pick = %t, other = %t, want true, true, false",
			selected.GetBool(sim), selectedOffensive.GetBool(sim), other.GetBool(sim))
	}
	inSlot := func(id int32, slot proto.OtherAction) bool {
		return rot.newValueSelectedPotion(&proto.APLValueSelectedPotion{PotionId: ActionID{ItemID: id}.ToProto(), Slot: slot}, nil).GetBool(sim)
	}
	offensive := inSlot(102, proto.OtherAction_OtherActionOffensivePotion)
	resourceInOffensive := inSlot(100, proto.OtherAction_OtherActionOffensivePotion)
	offensiveInDefensive := inSlot(102, proto.OtherAction_OtherActionDefensivePotion)
	if !offensive || resourceInOffensive || offensiveInDefensive {
		t.Errorf("Selected Potion in a slot: offensive pick = %t, resource pick = %t, offensive pick in the defensive slot = %t, want true, false, false",
			offensive, resourceInOffensive, offensiveInDefensive)
	}

	selectedConjured := rot.newValueSelectedConjured(&proto.APLValueSelectedConjured{ConjuredId: ActionID{ItemID: 200}.ToProto()}, nil)
	otherConjured := rot.newValueSelectedConjured(&proto.APLValueSelectedConjured{ConjuredId: ActionID{ItemID: 201}.ToProto()}, nil)
	if !selectedConjured.GetBool(sim) || otherConjured.GetBool(sim) {
		t.Errorf("Selected Conjured: matching = %t, other = %t, want true, false", selectedConjured.GetBool(sim), otherConjured.GetBool(sim))
	}

	wantNoValue(t, map[string]APLValue{
		"Selected Potion without a potion":          rot.newValueSelectedPotion(&proto.APLValueSelectedPotion{}, nil),
		"Selected Conjured without a conjured item": rot.newValueSelectedConjured(&proto.APLValueSelectedConjured{}, nil),
	})
}

// Each potion placeholder casts the potion the consumes pick for its class, and one whose class has
// no pick casts nothing.
func TestPotionPlaceholdersCastTheirClassPick(t *testing.T) {
	const resourceID, offensiveID int32 = 910201, 910202
	potions := map[int32]*Spell{}
	_, fw, rot := setupFakeRageRotationWith(func(char *Character) {
		for _, id := range []int32{resourceID, offensiveID} {
			potions[id] = char.RegisterSpell(SpellConfig{ActionID: ActionID{ItemID: id}, Flags: SpellFlagPotion | SpellFlagAPL})
		}
	})
	fw.Consumables.ResourcePotId = resourceID
	fw.Consumables.OffensivePotId = offensiveID

	placeholder := func(action proto.OtherAction) *Spell {
		return rot.GetAPLSpell(ActionID{OtherID: action}.ToProto())
	}
	if got := placeholder(proto.OtherAction_OtherActionResourcePotion); got != potions[resourceID] {
		t.Errorf("Resource Potion casts %v, want %v", got, potions[resourceID])
	}
	if got := placeholder(proto.OtherAction_OtherActionOffensivePotion); got != potions[offensiveID] {
		t.Errorf("Offensive Potion casts %v, want %v", got, potions[offensiveID])
	}
	if got := placeholder(proto.OtherAction_OtherActionDefensivePotion); got != nil {
		t.Errorf("Defensive Potion without a pick casts %v, want nothing", got)
	}
}

func TestValueActiveItemSwapSet(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()

	wantNoValue(t, map[string]APLValue{
		"Active Item Swap Set for an unknown swap set":   rot.newValueActiveItemSwapSet(&proto.APLValueActiveItemSwapSet{}, nil),
		"Active Item Swap Set without item swap enabled": rot.newValueActiveItemSwapSet(&proto.APLValueActiveItemSwapSet{SwapSet: proto.APLActionItemSwap_Swap1}, nil),
	})

	fw.ItemSwap.character = &fw.Character
	fw.ItemSwap.slots = []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}
	fw.ItemSwap.swapSet = proto.APLActionItemSwap_Main

	main := rot.newValueActiveItemSwapSet(&proto.APLValueActiveItemSwapSet{SwapSet: proto.APLActionItemSwap_Main}, nil)
	swap := rot.newValueActiveItemSwapSet(&proto.APLValueActiveItemSwapSet{SwapSet: proto.APLActionItemSwap_Swap1}, nil)
	if !main.GetBool(sim) || swap.GetBool(sim) {
		t.Errorf("main set active: Main = %t, Swap1 = %t, want true, false", main.GetBool(sim), swap.GetBool(sim))
	}
	fw.ItemSwap.swapSet = proto.APLActionItemSwap_Swap1
	if main.GetBool(sim) || !swap.GetBool(sim) {
		t.Errorf("swap set active: Main = %t, Swap1 = %t, want false, true", main.GetBool(sim), swap.GetBool(sim))
	}
}

func registerTestAuras(char *Character) (plain, stacking, icdAura *Aura) {
	plain = char.RegisterAura(Aura{Label: "Plain", ActionID: ActionID{SpellID: testPlainAuraID}, Duration: 10 * time.Second})
	stacking = char.RegisterAura(Aura{Label: "Stacking", ActionID: ActionID{SpellID: testStackAuraID}, Duration: 10 * time.Second, MaxStacks: 5})
	icdAura = char.RegisterAura(Aura{
		Label:    "Icd",
		ActionID: ActionID{SpellID: testICDAuraID},
		Duration: 10 * time.Second,
		Icd:      &Cooldown{Timer: char.NewTimer(), Duration: 30 * time.Second},
	})
	return plain, stacking, icdAura
}

func TestValueAuras(t *testing.T) {
	var plain, stacking, icdAura *Aura
	sim, _, rot := setupFakeRageRotationWith(func(char *Character) {
		plain, stacking, icdAura = registerTestAuras(char)
	})
	plainID := ActionID{SpellID: testPlainAuraID}.ToProto()
	stackID := ActionID{SpellID: testStackAuraID}.ToProto()
	icdID := ActionID{SpellID: testICDAuraID}.ToProto()

	wantNoValue(t, map[string]APLValue{
		"Aura Internal Cooldown for an aura without an ICD": rot.newValueAuraInternalCooldown(&proto.APLValueAuraInternalCooldown{AuraId: plainID}, nil),
		"Aura Num Stacks for a non-stacking aura":           rot.newValueAuraNumStacks(&proto.APLValueAuraNumStacks{AuraId: plainID}, nil),
	})

	known := rot.newValueAuraIsKnown(&proto.APLValueAuraIsKnown{AuraId: plainID}, nil)
	unknown := rot.newValueAuraIsKnown(&proto.APLValueAuraIsKnown{AuraId: ActionID{SpellID: testMissingAuraID}.ToProto()}, nil)
	if !known.GetBool(sim) || unknown.GetBool(sim) {
		t.Errorf("Aura Is Known: registered = %t, missing = %t, want true, false", known.GetBool(sim), unknown.GetBool(sim))
	}

	isActive := rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{AuraId: plainID}, nil)
	isInactive := rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{AuraId: plainID}, nil)
	remaining := rot.newValueAuraRemainingTime(&proto.APLValueAuraRemainingTime{AuraId: plainID}, nil)
	duration := rot.newValueAuraDuration(&proto.APLValueAuraDuration{AuraId: plainID}, nil)
	numStacks := rot.newValueAuraNumStacks(&proto.APLValueAuraNumStacks{AuraId: stackID}, nil)
	icdRemaining := rot.newValueAuraInternalCooldown(&proto.APLValueAuraInternalCooldown{AuraId: icdID}, nil)
	icdReady := rot.newValueAuraICDIsReady(&proto.APLValueAuraICDIsReady{AuraId: icdID}, nil)

	if isActive.GetBool(sim) || !isInactive.GetBool(sim) || remaining.GetDuration(sim) != 0 {
		t.Errorf("inactive aura: active = %t, inactive = %t, remaining = %s, want false, true, 0s", isActive.GetBool(sim), isInactive.GetBool(sim), remaining.GetDuration(sim))
	}
	if duration.GetDuration(sim) != 10*time.Second {
		t.Errorf("Aura Duration = %s, want 10s", duration.GetDuration(sim))
	}
	if numStacks.GetInt(sim) != 0 {
		t.Errorf("Aura Num Stacks while down = %d, want 0", numStacks.GetInt(sim))
	}
	if !icdReady.GetBool(sim) || icdRemaining.GetDuration(sim) != 0 {
		t.Errorf("fresh ICD: ready = %t, remaining = %s, want true, 0s", icdReady.GetBool(sim), icdRemaining.GetDuration(sim))
	}

	sim.CurrentTime = time.Second
	plain.Activate(sim)
	stacking.Activate(sim)
	stacking.SetStacks(sim, 3)
	icdAura.Icd.Use(sim)
	sim.CurrentTime = 3 * time.Second

	if !isActive.GetBool(sim) || isInactive.GetBool(sim) {
		t.Errorf("active aura: active = %t, inactive = %t, want true, false", isActive.GetBool(sim), isInactive.GetBool(sim))
	}
	if remaining.GetDuration(sim) != 8*time.Second {
		t.Errorf("Aura Remaining Time = %s, want 8s", remaining.GetDuration(sim))
	}
	if numStacks.GetInt(sim) != 3 {
		t.Errorf("Aura Num Stacks = %d, want 3", numStacks.GetInt(sim))
	}
	if icdReady.GetBool(sim) || icdRemaining.GetDuration(sim) != 28*time.Second {
		t.Errorf("ICD in use: ready = %t, remaining = %s, want false, 28s", icdReady.GetBool(sim), icdRemaining.GetDuration(sim))
	}
}

func TestValueAuraReactionTime(t *testing.T) {
	var plain, icdAura *Aura
	sim, fw, rot := setupFakeRageRotationWith(func(char *Character) {
		plain, _, icdAura = registerTestAuras(char)
	})
	fw.ReactionTime = 100 * time.Millisecond
	plainID := ActionID{SpellID: testPlainAuraID}.ToProto()

	isActive := rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{AuraId: plainID, IncludeReactionTime: true}, nil)
	isInactive := rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{AuraId: plainID, IncludeReactionTime: true}, nil)
	icdReady := rot.newValueAuraICDIsReady(&proto.APLValueAuraICDIsReady{AuraId: ActionID{SpellID: testICDAuraID}.ToProto(), IncludeReactionTime: true}, nil)

	sim.CurrentTime = time.Second
	plain.Activate(sim)
	icdAura.Icd.Use(sim)
	icdAura.Activate(sim)
	sim.CurrentTime = time.Second + 50*time.Millisecond
	if isActive.GetBool(sim) || isInactive.GetBool(sim) || !icdReady.GetBool(sim) {
		t.Errorf("inside the reaction time: active = %t, inactive = %t, ICD ready = %t, want false, false, true", isActive.GetBool(sim), isInactive.GetBool(sim), icdReady.GetBool(sim))
	}
	sim.CurrentTime = time.Second + 200*time.Millisecond
	if !isActive.GetBool(sim) || isInactive.GetBool(sim) || icdReady.GetBool(sim) {
		t.Errorf("past the reaction time: active = %t, inactive = %t, ICD ready = %t, want true, false, false", isActive.GetBool(sim), isInactive.GetBool(sim), icdReady.GetBool(sim))
	}
}

func TestValueAuraNumStacksFollowsCurrentTarget(t *testing.T) {
	request := fakeRageSimRequest()
	request.Encounter.Targets = append(request.Encounter.Targets, &proto.Target{Name: "second target", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2})
	sim := setupFakeRageSimFrom(request)
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	fw.ReactionTime = 100 * time.Millisecond
	first, second := sim.Encounter.AllTargetUnits[0], sim.Encounter.AllTargetUnits[1]

	sim.Environment.State = Constructed
	registerDebuff := func(target *Unit) *Aura {
		return target.RegisterAura(Aura{Label: "Stacking Debuff", ActionID: ActionID{SpellID: testStackAuraID}, Duration: 30 * time.Second, MaxStacks: 5})
	}
	firstDebuff, secondDebuff := registerDebuff(first), registerDebuff(second)
	sim.Environment.State = Finalized

	config := &proto.APLValueAuraNumStacks{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
		AuraId:     ActionID{SpellID: testStackAuraID}.ToProto(),
	}
	numStacks := fw.Rotation.newValueAuraNumStacks(config, nil)
	config.IncludeReactionTime = true
	reactedStacks := fw.Rotation.newValueAuraNumStacks(config, nil)

	sim.CurrentTime = time.Second
	firstDebuff.Activate(sim)
	firstDebuff.SetStacks(sim, 3)
	sim.CurrentTime = 2 * time.Second
	if numStacks.GetInt(sim) != 3 || reactedStacks.GetInt(sim) != 3 {
		t.Errorf("first target: stacks = %d, with reaction time = %d, want 3, 3", numStacks.GetInt(sim), reactedStacks.GetInt(sim))
	}

	fw.CurrentTarget = second
	if numStacks.GetInt(sim) != 0 || reactedStacks.GetInt(sim) != 0 {
		t.Errorf("second target before its debuff: stacks = %d, with reaction time = %d, want 0, 0", numStacks.GetInt(sim), reactedStacks.GetInt(sim))
	}
	secondDebuff.Activate(sim)
	secondDebuff.SetStacks(sim, 2)
	sim.CurrentTime = 2*time.Second + 50*time.Millisecond
	if numStacks.GetInt(sim) != 2 || reactedStacks.GetInt(sim) != 0 {
		t.Errorf("second target inside the reaction time: stacks = %d, with reaction time = %d, want 2, 0", numStacks.GetInt(sim), reactedStacks.GetInt(sim))
	}
	sim.CurrentTime = 2*time.Second + 200*time.Millisecond
	if reactedStacks.GetInt(sim) != 2 {
		t.Errorf("second target past the reaction time: with reaction time = %d, want 2", reactedStacks.GetInt(sim))
	}

	fw.CurrentTarget = first
	if numStacks.GetInt(sim) != 3 || reactedStacks.GetInt(sim) != 3 {
		t.Errorf("back on the first target: stacks = %d, with reaction time = %d, want 3, 3", numStacks.GetInt(sim), reactedStacks.GetInt(sim))
	}
}

func TestValueAuraShouldRefresh(t *testing.T) {
	var buff *Aura
	sim, _, rot := setupFakeRageRotationWith(func(char *Character) {
		buff, _, _ = registerTestAuras(char)
		buff.NewExclusiveEffect("aplAuraCategory", false, ExclusiveEffect{Priority: 1})
	})

	value := rot.newValueAuraShouldRefresh(&proto.APLValueAuraShouldRefresh{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_Self},
		AuraId:     ActionID{SpellID: testPlainAuraID}.ToProto(),
	}, nil)
	if value == nil {
		t.Fatalf("expected a Should Refresh value for a known aura")
	}

	if !value.GetBool(sim) {
		t.Errorf("Should Refresh with nothing active = false, want true")
	}
	buff.Activate(sim)
	if value.GetBool(sim) {
		t.Errorf("Should Refresh with the only effect active and a full duration left = true, want false")
	}
}

func TestValueAuraSets(t *testing.T) {
	const (
		strID     = 920001
		agiID     = 920002
		cooldown  = 920003
		cdSpellID = 920004
	)
	var strProc, cdAura *StatBuffAura

	sim, _, rot := setupFakeRageRotationWith(func(char *Character) {
		strProc = newStatAura(char, "Str Proc", strID, stats.Strength, 15*time.Second)
		strProc.Icd = &Cooldown{Timer: char.NewTimer(), Duration: 45 * time.Second}
		agiProc := newStatAura(char, "Agi Proc", agiID, stats.Agility, 10*time.Second)
		agiProc.Icd = &Cooldown{Timer: char.NewTimer(), Duration: 30 * time.Second}
		char.ItemProcBuffs = append(char.ItemProcBuffs, strProc, agiProc)

		cdAura = newStatAura(char, "Str Cooldown", cooldown, stats.Strength, 20*time.Second)
		spell := char.RegisterSpell(SpellConfig{ActionID: ActionID{SpellID: cdSpellID}})
		char.initialMajorCooldowns = append(char.initialMajorCooldowns, MajorCooldown{Spell: spell, BuffAura: cdAura})
	})

	const none = -1
	valuesFor := func(stat int32) map[string]APLValue {
		return map[string]APLValue{
			"All Item Stat Procs Active":           rot.newValueAllItemStatProcsActive(&proto.APLValueAllItemStatProcsActive{StatType1: stat, StatType2: none, StatType3: none}, nil),
			"Any Item Stat Procs Active":           rot.newValueAnyItemStatProcsActive(&proto.APLValueAnyItemStatProcsActive{StatType1: stat, StatType2: none, StatType3: none}, nil),
			"Any Item Stat Procs Available":        rot.newValueAnyItemStatProcsAvailable(&proto.APLValueAnyItemStatProcsAvailable{StatType1: stat, StatType2: none, StatType3: none}, nil),
			"Item Procs Min Remaining Time":        rot.newValueItemProcsMinRemainingTime(&proto.APLValueItemProcsMinRemainingTime{StatType1: stat, StatType2: none, StatType3: none}, nil),
			"Item Procs Max Remaining ICD":         rot.newValueItemsProcsMaxRemainingICD(&proto.APLValueItemProcsMaxRemainingICD{StatType1: stat, StatType2: none, StatType3: none}, nil),
			"Any Stat Buff Cooldowns Active":       rot.newValueAnyStatBuffCooldownsActive(&proto.APLValueAnyStatBuffCooldownsActive{StatType1: stat, StatType2: none, StatType3: none}, nil),
			"Any Stat Buff Cooldowns Min Duration": rot.newValueAnyStatBuffCooldownsMinDuration(&proto.APLValueAnyStatBuffCooldownsMinDuration{StatType1: stat, StatType2: none, StatType3: none}, nil),
			"Num Equipped Stat Proc Items":         rot.newValueNumEquippedStatProcItems(&proto.APLValueNumEquippedStatProcItems{StatType1: stat, StatType2: none, StatType3: none}, nil),
			"Num Stat Buff Cooldowns":              rot.newValueNumStatBuffCooldowns(&proto.APLValueNumStatBuffCooldowns{StatType1: stat, StatType2: none, StatType3: none}, nil),
		}
	}
	str := int32(proto.Stat_StatStrength)

	unbuffed := valuesFor(int32(proto.Stat_StatDodgeRating))
	for _, name := range []string{"Num Equipped Stat Proc Items", "Num Stat Buff Cooldowns"} {
		if got := unbuffed[name].GetInt(sim); got != 0 {
			t.Errorf("%s for an unbuffed stat = %d, want 0", name, got)
		}
		delete(unbuffed, name)
	}
	wantNoValue(t, unbuffed)

	values := valuesFor(str)
	for name, value := range values {
		if value == nil {
			t.Fatalf("%s for a stat that a proc and a cooldown buff: got no value", name)
		}
	}
	allActive, anyActive, anyAvailable := values["All Item Stat Procs Active"], values["Any Item Stat Procs Active"], values["Any Item Stat Procs Available"]
	minRemaining, maxICD := values["Item Procs Min Remaining Time"], values["Item Procs Max Remaining ICD"]
	cdActive, cdMinDuration := values["Any Stat Buff Cooldowns Active"], values["Any Stat Buff Cooldowns Min Duration"]
	numItems, numCooldowns := values["Num Equipped Stat Proc Items"], values["Num Stat Buff Cooldowns"]
	numItemsHighIcd := rot.newValueNumEquippedStatProcItems(&proto.APLValueNumEquippedStatProcItems{StatType1: str, StatType2: none, StatType3: none, MinIcdSeconds: 50}, nil)

	if got := numItems.GetInt(sim); got != 1 {
		t.Errorf("Num Equipped Stat Proc Items = %d, want 1", got)
	}
	if got := numItemsHighIcd.GetInt(sim); got != 0 {
		t.Errorf("Num Equipped Stat Proc Items with a minimum ICD above the proc's = %d, want 0", got)
	}
	if got := numCooldowns.GetInt(sim); got != 1 {
		t.Errorf("Num Stat Buff Cooldowns = %d, want 1", got)
	}

	if allActive.GetBool(sim) || anyActive.GetBool(sim) || !anyAvailable.GetBool(sim) || cdActive.GetBool(sim) {
		t.Errorf("nothing active: all = %t, any = %t, available = %t, cooldown = %t, want false, false, true, false",
			allActive.GetBool(sim), anyActive.GetBool(sim), anyAvailable.GetBool(sim), cdActive.GetBool(sim))
	}
	if got := minRemaining.GetDuration(sim); got != NeverExpires {
		t.Errorf("Item Procs Min Remaining Time with nothing active = %s, want never", got)
	}
	if got := cdMinDuration.GetDuration(sim); got != NeverExpires {
		t.Errorf("Any Stat Buff Cooldowns Min Duration with nothing active = %s, want never", got)
	}
	if got := maxICD.GetDuration(sim); got != 0 {
		t.Errorf("Item Procs Max Remaining ICD with a ready proc = %s, want 0s", got)
	}

	sim.CurrentTime = time.Second
	strProc.Icd.Use(sim)
	strProc.Activate(sim)
	cdAura.Activate(sim)
	sim.CurrentTime = 6 * time.Second

	if !allActive.GetBool(sim) || !anyActive.GetBool(sim) || anyAvailable.GetBool(sim) || !cdActive.GetBool(sim) {
		t.Errorf("proc and cooldown up: all = %t, any = %t, available = %t, cooldown = %t, want true, true, false, true",
			allActive.GetBool(sim), anyActive.GetBool(sim), anyAvailable.GetBool(sim), cdActive.GetBool(sim))
	}
	if got := minRemaining.GetDuration(sim); got != 10*time.Second {
		t.Errorf("Item Procs Min Remaining Time = %s, want 10s", got)
	}
	if got := cdMinDuration.GetDuration(sim); got != 15*time.Second {
		t.Errorf("Any Stat Buff Cooldowns Min Duration = %s, want 15s", got)
	}

	sim.CurrentTime = 20 * time.Second
	strProc.Deactivate(sim)
	if got := maxICD.GetDuration(sim); got != 26*time.Second {
		t.Errorf("Item Procs Max Remaining ICD = %s, want 26s", got)
	}
	if anyAvailable.GetBool(sim) {
		t.Errorf("Any Item Stat Procs Available while the proc is on its ICD = true, want false")
	}

	strProc.IsSwapped = true
	if got := numItems.GetInt(sim); got != 0 {
		t.Errorf("Num Equipped Stat Proc Items with the item swapped out = %d, want 0", got)
	}
	if !allActive.GetBool(sim) {
		t.Errorf("All Item Stat Procs Active with the only proc swapped out = false, want true")
	}
}

func TestValueDots(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	dot := fw.GetSpell(fakeRendActionID).Dot(sim.Encounter.AllTargetUnits[0])
	id := fakeRendActionID.ToProto()

	isActive := rot.newValueDotIsActive(&proto.APLValueDotIsActive{SpellId: id}, nil)
	activeOnAll := rot.newValueDotIsActiveOnAllTargets(&proto.APLValueDotIsActiveOnAllTargets{SpellId: id}, nil)
	remaining := rot.newValueDotRemainingTime(&proto.APLValueDotRemainingTime{SpellId: id}, nil)
	lowest := rot.newValueDotLowestRemainingTime(&proto.APLValueDotLowestRemainingTime{SpellId: id}, nil)
	tickFrequency := rot.newValueDotTickFrequency(&proto.APLValueDotTickFrequency{SpellId: id}, nil)
	nextTick := rot.newValueDotTimeToNextTick(&proto.APLValueDotTimeToNextTick{SpellId: id}, nil)
	baseDuration := rot.newValueDotBaseDuration(&proto.APLValueDotBaseDuration{SpellId: id}, nil)

	if isActive == nil || activeOnAll == nil || remaining == nil || lowest == nil || tickFrequency == nil || nextTick == nil || baseDuration == nil {
		t.Fatalf("Expected every dot value to resolve the fake Rend dot")
	}

	if isActive.GetBool(sim) || activeOnAll.GetBool(sim) || remaining.GetDuration(sim) != 0 || lowest.GetDuration(sim) != 0 {
		t.Fatalf("Inactive dot: is active = %t, active on all = %t, remaining = %s, lowest remaining = %s, want false, false, 0s, 0s",
			isActive.GetBool(sim), activeOnAll.GetBool(sim), remaining.GetDuration(sim), lowest.GetDuration(sim))
	}
	if tickFrequency.GetDuration(sim) != 3*time.Second || baseDuration.GetDuration(sim) != 15*time.Second {
		t.Fatalf("Inactive dot: tick frequency = %s, base duration = %s, want 3s, 15s", tickFrequency.GetDuration(sim), baseDuration.GetDuration(sim))
	}

	dot.Apply(sim)
	sim.CurrentTime = time.Second

	if !isActive.GetBool(sim) || !activeOnAll.GetBool(sim) {
		t.Fatalf("Active dot: is active = %t, active on all = %t, want true, true", isActive.GetBool(sim), activeOnAll.GetBool(sim))
	}
	if remaining.GetDuration(sim) != 14*time.Second || lowest.GetDuration(sim) != 14*time.Second {
		t.Fatalf("Active dot: remaining = %s, lowest remaining = %s, want 14s, 14s", remaining.GetDuration(sim), lowest.GetDuration(sim))
	}
	if tickFrequency.GetDuration(sim) != 3*time.Second || nextTick.GetDuration(sim) != 2*time.Second {
		t.Fatalf("Active dot: tick frequency = %s, time to next tick = %s, want 3s, 2s", tickFrequency.GetDuration(sim), nextTick.GetDuration(sim))
	}
}

// Disabled for now, with the DoT increase values in apl_values_dot.go.
/*
func TestValueDotPercentIncrease(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	rend := fw.GetSpell(fakeRendActionID)
	id := fakeRendActionID.ToProto()

	wantNoValue(t, map[string]APLValue{
		"Dot Percent Increase without an expected tick damage calculator":           rot.newValueDotPercentIncrease(&proto.APLValueDotPercentIncrease{SpellId: id}, nil),
		"Dot Tick Rate Percent Increase without an expected tick damage calculator": rot.newValueDotTickRatePercentIncrease(&proto.APLValueDotPercentIncrease{SpellId: id}, nil),
	})

	snapshotDamage, currentDamage := 0.0, 0.0
	result := &SpellResult{}
	rend.expectedTickDamageInternal = func(_ *Simulation, _ *Unit, _ *Spell, useSnapshot bool) *SpellResult {
		result.Damage = currentDamage
		if useSnapshot {
			result.Damage = snapshotDamage
		}
		return result
	}

	percent := rot.newValueDotPercentIncrease(&proto.APLValueDotPercentIncrease{SpellId: id}, nil)
	tickRate := rot.newValueDotTickRatePercentIncrease(&proto.APLValueDotPercentIncrease{SpellId: id}, nil)
	if percent == nil || tickRate == nil {
		t.Fatalf("Expected both increase values to resolve once the spell has an expected tick damage calculator")
	}

	if got := percent.GetFloat(sim); got != 1 {
		t.Fatalf("Dot Percent Increase with no snapshot damage = %f, want 1", got)
	}
	if got := tickRate.GetFloat(sim); got != 1 {
		t.Fatalf("Dot Tick Rate Percent Increase on an inactive dot = %f, want 1", got)
	}

	snapshotDamage, currentDamage = 100, 125
	if got := percent.GetFloat(sim); got != 0.25 {
		t.Fatalf("Dot Percent Increase for 100 snapshot and 125 current damage = %f, want 0.25", got)
	}

	rend.Dot(sim.Encounter.AllTargetUnits[0]).Apply(sim)
	if got := tickRate.GetFloat(sim); got != 0 {
		t.Fatalf("Dot Tick Rate Percent Increase on an active dot without haste = %f, want 0", got)
	}
}
*/

func TestValueDotIncreaseDisabled(t *testing.T) {
	_, _, rot := setupFakeRageRotation()
	id := fakeRendActionID.ToProto()
	configs := map[string]*proto.APLValue{
		"Dot Percent Increase": {
			Uuid:  &proto.UUID{Value: "dot-percent-increase"},
			Value: &proto.APLValue_DotPercentIncrease{DotPercentIncrease: &proto.APLValueDotPercentIncrease{SpellId: id}},
		},
		"Dot Tick Rate Percent Increase": {
			Uuid:  &proto.UUID{Value: "dot-tick-rate-percent-increase"},
			Value: &proto.APLValue_DotTickRatePercentIncrease{DotTickRatePercentIncrease: &proto.APLValueDotPercentIncrease{SpellId: id}},
		},
	}

	for name, config := range configs {
		if value := rot.newAPLValue(config); value != nil {
			t.Errorf("%s: got a value, want none", name)
		}
		if warnings := rot.uuidValidations[config.Uuid]; len(warnings) != 1 || warnings[0].LogLevel != proto.LogLevel_Warning {
			t.Errorf("%s: validations = %v, want one warning", name, warnings)
		}
	}
}

func TestValueSpellExpectedDpet(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	rend := fw.GetSpell(fakeRendActionID)
	config := &proto.APLValueSpellExpectedDpet{SpellId: fakeRendActionID.ToProto()}

	wantNoValue(t, map[string]APLValue{
		"Spell Expected DPET without an expected damage calculator": rot.newValueSpellExpectedDpet(config, nil),
	})

	hit600 := func(_ *Simulation, _ *Unit, _ *Spell, _ bool) *SpellResult {
		return &SpellResult{Damage: 600}
	}
	rend.expectedInitialDamageInternal = hit600
	dpet := rot.newValueSpellExpectedDpet(config, nil)
	if dpet == nil {
		t.Fatalf("Expected Spell Expected DPET to resolve once the spell has an expected initial damage calculator")
	}

	rend.IgnoreHaste = false
	rend.DefaultCast.GCD = GCDDefault
	rend.DefaultCast.CastTime = 2 * time.Second
	if got := dpet.GetFloat(sim); got != 300 {
		t.Errorf("a 600 hit with a 2s cast = %f, want 300", got)
	}

	rend.expectedTickDamageInternal = func(_ *Simulation, _ *Unit, _ *Spell, _ bool) *SpellResult {
		return &SpellResult{Damage: 150}
	}
	if got := dpet.GetFloat(sim); got != 675 {
		t.Errorf("a 600 hit and 5 ticks of 150 with a 2s cast = %f, want 675", got)
	}

	rend.expectedInitialDamageInternal = nil
	rend.DefaultCast.CastTime = 0
	if got := dpet.GetFloat(sim); got != 500 {
		t.Errorf("5 ticks of 150 on an instant with a 1.5s GCD = %f, want 500", got)
	}

	rend.Flags |= SpellFlagChanneled
	if got := dpet.GetFloat(sim); got != 50 {
		t.Errorf("5 ticks of 150 channeled every 3s = %f, want 50", got)
	}

	rend.Flags &^= SpellFlagChanneled
	dot := rend.Dot(sim.Encounter.AllTargetUnits[0])
	dot.BaseDurationFlat = 6 * time.Second
	if got := dpet.GetFloat(sim); got != 700 {
		t.Errorf("7 ticks of 150 (two ticks of extra base duration) on an instant with a 1.5s GCD = %f, want 700", got)
	}
	dot.BaseDurationFlat = 0

	sim.CurrentTime = sim.Duration - 7*time.Second
	if got := dpet.GetFloat(sim); got != 200 {
		t.Errorf("2 of 5 ticks of 150 landing in the last 7s on an instant with a 1.5s GCD = %f, want 200", got)
	}
	sim.CurrentTime = sim.Duration - 6*time.Second
	if got := dpet.GetFloat(sim); got != 200 {
		t.Errorf("2 of 5 ticks of 150 in the last 6s, the second on the last moment of the fight = %f, want 200", got)
	}
	rend.Flags |= SpellFlagChanneled
	sim.CurrentTime = sim.Duration - 7*time.Second
	if got := dpet.GetFloat(sim); got != 50 {
		t.Errorf("2 of 5 ticks of 150 channeled every 3s in the last 7s = %f, want 50", got)
	}
	rend.Flags &^= SpellFlagChanneled
	rend.expectedInitialDamageInternal = hit600
	rend.DefaultCast.CastTime = 2 * time.Second
	sim.CurrentTime = sim.Duration - 9*time.Second
	if got := dpet.GetFloat(sim); got != 450 {
		t.Errorf("a 600 hit and 2 of 5 ticks of 150 after a 2s cast in the last 9s = %f, want 450", got)
	}
	sim.CurrentTime = sim.Duration - time.Second
	if got := dpet.GetFloat(sim); got != 0 {
		t.Errorf("a 2s cast in the last 1s = %f, want 0", got)
	}
	rend.MissileMinDuration = time.Second
	sim.CurrentTime = sim.Duration - 8500*time.Millisecond
	if got := dpet.GetFloat(sim); got != 375 {
		t.Errorf("a 600 hit and 1 of 5 ticks of 150 after a 2s cast and 1s travel in the last 8.5s = %f, want 375", got)
	}
	sim.CurrentTime = sim.Duration - 2500*time.Millisecond
	if got := dpet.GetFloat(sim); got != 0 {
		t.Errorf("a 2s cast with 1s travel in the last 2.5s = %f, want 0", got)
	}
	rend.MissileMinDuration = 0
	sim.CurrentTime = 0
	rend.expectedInitialDamageInternal = nil
	rend.DefaultCast.CastTime = 0

	rend.DefaultCast.GCD = 0
	if got := dpet.GetFloat(sim); got != 0 {
		t.Errorf("a spell with neither a cast time nor a GCD = %f, want 0", got)
	}
}

var (
	testMeleeSpecialActionID = ActionID{SpellID: 910101}
	testFireSpellActionID    = ActionID{SpellID: 910102}
	testWandShotActionID     = ActionID{SpellID: 910103}
)

func registerCapTestSpells(char *Character) {
	char.RegisterSpell(SpellConfig{
		ActionID:    testMeleeSpecialActionID,
		SpellSchool: SpellSchoolPhysical,
		DefenseType: DefenseTypeMelee,
		ProcMask:    ProcMaskMeleeMHSpecial,
		Flags:       SpellFlagAPL,
	})
	char.RegisterSpell(SpellConfig{
		ActionID:       testFireSpellActionID,
		SpellSchool:    SpellSchoolFire,
		DefenseType:    DefenseTypeMagic,
		ProcMask:       ProcMaskSpellDamage,
		Flags:          SpellFlagAPL,
		ClassSpellMask: 1,
	})
	char.RegisterSpell(SpellConfig{
		ActionID:    testWandShotActionID,
		SpellSchool: SpellSchoolFire,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskRangedAuto,
		Flags:       SpellFlagAPL,
	})
}

// Against the level 63 target a level 60 attacker with no weapon skill bonus misses a special 8% of the time
// and a spell 17%.
func TestValueSpellIsHitCapped(t *testing.T) {
	sim, fw, rot := setupFakeRageRotationWith(registerCapTestSpells)
	hitCapped := func(id ActionID) APLValue {
		return rot.newValueSpellIsHitCapped(&proto.APLValueSpellIsHitCapped{SpellId: id.ToProto()}, nil)
	}
	special := hitCapped(testMeleeSpecialActionID)
	rend := hitCapped(fakeRendActionID)
	fire := hitCapped(testFireSpellActionID)
	wand := hitCapped(testWandShotActionID)

	if !fw.AutoAttacks.IsDualWielding {
		t.Fatal("the fake warrior should dual wield")
	}

	fw.AddStatDynamic(sim, stats.PhysicalHitPercent, 7)
	if special.GetBool(sim) {
		t.Errorf("Hit Capped(melee special) with 7%% hit: got true, want false")
	}
	fw.AddStatDynamic(sim, stats.PhysicalHitPercent, 1)
	if !special.GetBool(sim) {
		t.Errorf("Hit Capped(melee special) with 8%% hit while dual wielding: got false, want true")
	}
	if !rend.GetBool(sim) {
		t.Errorf("Hit Capped(Rend), a melee special without a DefenseType, with 8%% hit: got false, want true")
	}
	if !wand.GetBool(sim) {
		t.Errorf("Hit Capped(wand shot) with 8%% physical hit: got false, want true")
	}
	if fire.GetBool(sim) {
		t.Errorf("Hit Capped(Fire spell) with 8%% physical hit: got true, want false")
	}
	fw.AddStatDynamic(sim, stats.PhysicalHitPercent, -8)

	fw.AddStatDynamic(sim, stats.SpellHitPercent, 10)
	if fire.GetBool(sim) {
		t.Errorf("Hit Capped(Fire spell) with 10%% spell hit: got true, want false")
	}
	fw.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexFire] = 7
	if !fire.GetBool(sim) {
		t.Errorf("Hit Capped(Fire spell) with 10%% spell hit and 7%% Fire hit: got false, want true")
	}
	fw.AddStatDynamic(sim, stats.SpellHitPercent, 7)
	if wand.GetBool(sim) {
		t.Errorf("Hit Capped(wand shot) with 17%% spell hit and no physical hit: got true, want false")
	}
}

// Against the level 63 target a level 60 attacker with no weapon skill bonus is dodged 6.5% of the time
// and parried 14%.
func TestValueSpellIsExpertiseCapped(t *testing.T) {
	sim, fw, rot := setupFakeRageRotationWith(registerCapTestSpells)
	expertiseCapped := func(id ActionID) APLValue {
		return rot.newValueSpellIsExpertiseCapped(&proto.APLValueSpellIsExpertiseCapped{SpellId: id.ToProto()}, nil)
	}
	special := expertiseCapped(testMeleeSpecialActionID)

	if !expertiseCapped(testFireSpellActionID).GetBool(sim) || !expertiseCapped(testWandShotActionID).GetBool(sim) {
		t.Errorf("Expertise Capped of a Fire spell and a wand shot without expertise: got false, want true")
	}

	fw.PseudoStats.InFrontOfTarget = false
	fw.AddStatDynamic(sim, stats.ExpertisePercent, 6)
	if special.GetBool(sim) {
		t.Errorf("Expertise Capped behind the target with 6%% expertise: got true, want false")
	}
	fw.AddStatDynamic(sim, stats.ExpertisePercent, 1)
	if !special.GetBool(sim) {
		t.Errorf("Expertise Capped behind the target with 7%% expertise: got false, want true")
	}

	fw.PseudoStats.InFrontOfTarget = true
	if special.GetBool(sim) {
		t.Errorf("Expertise Capped in front of the target with 7%% expertise: got true, want false")
	}
	fw.AddStatDynamic(sim, stats.ExpertisePercent, 7)
	if !special.GetBool(sim) {
		t.Errorf("Expertise Capped in front of the target with 14%% expertise: got false, want true")
	}

	fw.AddStatDynamic(sim, stats.ExpertisePercent, -14)
	fw.GetSpell(testMeleeSpecialActionID).Flags |= SpellFlagNoActiveDefense
	if !special.GetBool(sim) {
		t.Errorf("Expertise Capped in front of the target without expertise, flagged No Active Defense: got false, want true")
	}
}

// Against a level 62 target the special's miss and dodge are 0.05 + 10*0.001, which sums to slightly above
// 0.06 in floats, so 6% hit and 6% expertise reach the cap only within that remainder.
func TestValueSpellCappedAtFloatRemainder(t *testing.T) {
	request := fakeRageSimRequest()
	request.Encounter.Targets[0].Level = 62
	sim, fw, rot := setupFakeRageRotationFrom(request, registerCapTestSpells)
	special := fw.GetSpell(testMeleeSpecialActionID)
	if miss := special.physicalMissChanceNoDWPenalty(fw.AttackTables[sim.Encounter.ActiveTargetUnits[0].UnitIndex]); miss <= 0.06 {
		t.Fatalf("the special's miss against a level 62 target = %v, want slightly above 0.06", miss)
	}

	fw.AddStatDynamic(sim, stats.PhysicalHitPercent, 6)
	fw.AddStatDynamic(sim, stats.ExpertisePercent, 6)
	fw.PseudoStats.InFrontOfTarget = false
	id := testMeleeSpecialActionID.ToProto()
	if !rot.newValueSpellIsHitCapped(&proto.APLValueSpellIsHitCapped{SpellId: id}, nil).GetBool(sim) {
		t.Errorf("Hit Capped(melee special) with 6%% hit against a level 62 target: got false, want true")
	}
	if !rot.newValueSpellIsExpertiseCapped(&proto.APLValueSpellIsExpertiseCapped{SpellId: id}, nil).GetBool(sim) {
		t.Errorf("Expertise Capped behind a level 62 target with 6%% expertise: got false, want true")
	}
}

func TestValueSequences(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	castSpell := &APLActionCastSpell{spell: fw.GetSpell(fakeCategorySpellActionIDs[0])}
	sequences := map[string]*APLActionSequence{
		"empty":    {unit: &fw.Unit, name: "empty"},
		"waiting":  {unit: &fw.Unit, name: "waiting", subactions: []*APLAction{{impl: &stubAction{ready: false}}}},
		"ready":    {unit: &fw.Unit, name: "ready", subactions: []*APLAction{{impl: &stubAction{ready: true}}}},
		"casting":  {unit: &fw.Unit, name: "casting", subactions: []*APLAction{{impl: castSpell}}},
		"finished": {unit: &fw.Unit, name: "finished", subactions: []*APLAction{{impl: &stubAction{ready: true}}}, curIdx: 1},
	}
	for _, sequence := range sequences {
		rot.priorityList = append(rot.priorityList, &APLAction{impl: sequence})
	}

	wantNoValue(t, map[string]APLValue{
		"Sequence Is Complete without a sequence name":   rot.newValueSequenceIsComplete(&proto.APLValueSequenceIsComplete{}, nil),
		"Sequence Is Ready without a sequence name":      rot.newValueSequenceIsReady(&proto.APLValueSequenceIsReady{}, nil),
		"Sequence Time To Ready without a sequence name": rot.newValueSequenceTimeToReady(&proto.APLValueSequenceTimeToReady{}, nil),
	})

	tests := []struct {
		name            string
		complete, ready bool
		timeToReady     time.Duration
	}{
		{"empty", true, false, NeverExpires},
		{"waiting", false, false, 3 * time.Second},
		{"ready", false, true, 0},
		{"casting", false, false, 0},
		{"finished", true, false, NeverExpires},
		{"unknown", false, false, NeverExpires},
	}
	for _, tc := range tests {
		complete := rot.newValueSequenceIsComplete(&proto.APLValueSequenceIsComplete{SequenceName: tc.name}, nil)
		ready := rot.newValueSequenceIsReady(&proto.APLValueSequenceIsReady{SequenceName: tc.name}, nil)
		timeToReady := rot.newValueSequenceTimeToReady(&proto.APLValueSequenceTimeToReady{SequenceName: tc.name}, nil)
		for _, value := range []APLValue{complete, ready, timeToReady} {
			value.Finalize(rot)
		}
		if got := complete.GetBool(sim); got != tc.complete {
			t.Errorf("%s: Sequence Is Complete = %t, want %t", tc.name, got, tc.complete)
		}
		if got := ready.GetBool(sim); got != tc.ready {
			t.Errorf("%s: Sequence Is Ready = %t, want %t", tc.name, got, tc.ready)
		}
		if got := timeToReady.GetDuration(sim); got != tc.timeToReady {
			t.Errorf("%s: Sequence Time To Ready = %s, want %s", tc.name, got, tc.timeToReady)
		}
	}
}

func TestValueActionGroupUsed(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()
	rot.groups = []*APLGroup{
		{name: "referenced", referencedBy: &APLActionGroupReference{}},
		{name: "unreferenced"},
	}

	for name, want := range map[string]bool{"referenced": true, "unreferenced": false, "missing": false} {
		value := rot.newValueActionGroupUsed(&proto.APLValueActionGroupUsed{Name: name}, nil)
		value.Finalize(rot)
		if got := value.GetBool(sim); got != want {
			t.Errorf("group %s: Action Group Is Used = %t, want %t", name, got, want)
		}
	}
}

func TestValueProperties(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	fw.ChannelClipDelay = 40 * time.Millisecond
	fw.ReactionTime = 200 * time.Millisecond

	clip := rot.newValueChannelClipDelay(&proto.APLValueChannelClipDelay{}, nil)
	if got := clip.GetDuration(sim); got != 40*time.Millisecond {
		t.Errorf("Channel Clip Delay = %s, want 40ms", got)
	}
	input := rot.newValueInputDelay(&proto.APLValueInputDelay{}, nil)
	if got := input.GetDuration(sim); got != 200*time.Millisecond {
		t.Errorf("Input Delay = %s, want 200ms", got)
	}
}

func TestValueMultipleCdUsages(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	target := fw.GetSpell(fakeCategorySpellActionIDs[0])
	plain := fw.GetSpell(fakeCategorySpellActionIDs[1])
	target.RelatedSelfBuff = &Aura{Duration: 10 * time.Second}
	sim.Duration = 180 * time.Second
	targetID := target.ActionID.ToProto()

	wantNoValue(t, map[string]APLValue{
		"Multiple Cd Usages without a target spell":                  rot.newValueMultipleCdUsages(&proto.APLValueMultipleCdUsages{}, nil),
		"Multiple Cd Usages for a spell without a related self buff": rot.newValueMultipleCdUsages(&proto.APLValueMultipleCdUsages{TargetSpellId: plain.ActionID.ToProto()}, nil),
	})

	always := rot.newValueMultipleCdUsages(&proto.APLValueMultipleCdUsages{TargetSpellId: targetID}, nil)
	aligned := rot.newValueMultipleCdUsages(&proto.APLValueMultipleCdUsages{TargetSpellId: targetID, AlignCdEnd: true}, nil)
	if always == nil || aligned == nil {
		t.Fatal("a target spell with a related self buff should construct")
	}

	// The UI sends an empty base spell id when none is picked, which counts as no base spell.
	for _, tc := range []struct {
		name        string
		base        *proto.ActionID
		wantWarning bool
	}{
		{"an empty base spell id", &proto.ActionID{}, false},
		{"an unknown base spell", ActionID{SpellID: 999999}.ToProto(), true},
		{"a base spell without a related self buff", plain.ActionID.ToProto(), true},
	} {
		rot.curValidations = nil
		value := rot.newValueMultipleCdUsages(&proto.APLValueMultipleCdUsages{BaseSpellId: tc.base, TargetSpellId: targetID, AlignCdEnd: true}, nil)
		if value == nil {
			t.Errorf("%s should construct", tc.name)
		}
		if !tc.wantWarning {
			if len(rot.curValidations) != 0 {
				t.Errorf("%s should log no validation, got %v", tc.name, rot.curValidations)
			}
		} else if len(rot.curValidations) == 0 || !strings.HasSuffix(rot.curValidations[len(rot.curValidations)-1].Validation, "Ignoring the base spell and offset.") {
			t.Errorf("%s should log the ignored base spell warning, got %v", tc.name, rot.curValidations)
		}
	}
	rot.curValidations = nil

	// With a 60s cooldown and a 10s buff, the aligned value casts while a later use would still get at least
	// half its buff (more than 65s left) or the buff would run to the end, and holds in between. The
	// cooldown is read when the value is evaluated, so a cooldown change applies at once.
	for _, tc := range []struct {
		elapsed  time.Duration
		cooldown time.Duration
		fight    time.Duration
		aligned  bool
	}{
		{0, time.Minute, 180 * time.Second, true},
		{114 * time.Second, time.Minute, 180 * time.Second, true},
		{115 * time.Second, time.Minute, 180 * time.Second, false},
		{119 * time.Second, time.Minute, 180 * time.Second, false},
		{120 * time.Second, time.Minute, 180 * time.Second, false},
		{150 * time.Second, time.Minute, 180 * time.Second, false},
		{170 * time.Second, time.Minute, 180 * time.Second, true},
		{175 * time.Second, time.Minute, 180 * time.Second, true},
		{144 * time.Second, 30 * time.Second, 180 * time.Second, true},
		{0, time.Minute, 8 * time.Second, true},
	} {
		target.CD.Duration, sim.Duration, sim.CurrentTime = tc.cooldown, tc.fight, tc.elapsed
		if !always.GetBool(sim) {
			t.Errorf("at %s of %s without a base spell: got false, want true", tc.elapsed, tc.fight)
		}
		if got := aligned.GetBool(sim); got != tc.aligned {
			t.Errorf("at %s of %s with a %s cooldown aligned to the end of combat: got %t, want %t", tc.elapsed, tc.fight, tc.cooldown, got, tc.aligned)
		}
	}

	// A spell with only a shared category lockout, like a potion, counts that lockout as its cooldown.
	target.CD.Duration, target.SharedCD.Duration, sim.Duration = 0, 120*time.Second, 300*time.Second
	for _, tc := range []struct {
		remaining time.Duration
		aligned   bool
	}{
		{126 * time.Second, true},
		{125 * time.Second, false},
		{10 * time.Second, true},
		{11 * time.Second, false},
	} {
		sim.CurrentTime = sim.Duration - tc.remaining
		if got := aligned.GetBool(sim); got != tc.aligned {
			t.Errorf("with %s left and a 120s shared cooldown aligned to the end of combat: got %t, want %t", tc.remaining, got, tc.aligned)
		}
	}
	target.SharedCD.Duration = 0

	target.CD.Duration = 5 * time.Second
	sim.Duration = 180 * time.Second
	for elapsed := time.Duration(0); elapsed < 180*time.Second; elapsed += time.Second {
		sim.CurrentTime = elapsed
		if !aligned.GetBool(sim) {
			t.Errorf("at %s with a cooldown shorter than the buff, aligned should cast", elapsed)
		}
	}
}

func TestCastSpellWithoutSpellID(t *testing.T) {
	rot := &APLRotation{
		unit: &Unit{},
	}

	if action := rot.newActionCastSpell(&proto.APLActionCastSpell{}); action != nil {
		t.Fatalf("Expected no action for a cast without a spell, got %v", action)
	}
}
