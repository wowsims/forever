package core

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// Every action of an APL rotation names the entry it was built from by the config's own positions: hidden
// rows count, a group segment names the group's place in the config rather than the sim's copy of it, a
// group another group references adds a segment of its own, and a step keeps its place when an earlier
// step fails to build. Auto and custom rotations name no entry path.
func TestAPLEntryPathsFollowConfigPositions(t *testing.T) {
	const dotID, unknownID = 910105, 910106
	refTo := func(name string) *proto.APLAction {
		return &proto.APLAction{Action: &proto.APLAction_GroupReference{GroupReference: &proto.APLActionGroupReference{GroupName: name}}}
	}
	config := func(rotationType proto.APLRotation_Type) *proto.APLRotation {
		return &proto.APLRotation{
			Type: rotationType,
			PrepullActions: []*proto.APLPrepullAction{
				{Action: aplCast(dotID), DoAtValue: constValue("-2s"), Hide: true},
				{Action: aplCast(dotID), DoAtValue: constValue("-1s")},
			},
			PriorityList: append([]*proto.APLListItem{{Action: aplCast(dotID), Hide: true}}, aplItems(
				aplCast(dotID),
				refTo("b"),
				refTo("b"),
				&proto.APLAction{Action: &proto.APLAction_Schedule{Schedule: &proto.APLActionSchedule{Schedule: "0s", InnerAction: aplCast(dotID)}}},
				&proto.APLAction{Action: &proto.APLAction_StrictMultidot{StrictMultidot: &proto.APLActionStrictMultidot{SpellId: ActionID{SpellID: dotID}.ToProto(), MaxDots: 2}}},
				refTo("a"),
			)...),
			Groups: []*proto.APLGroup{
				{Name: "a", Actions: append([]*proto.APLListItem{{Action: aplCast(dotID), Hide: true}}, aplItems(refTo("b"))...)},
				{Name: "b", Actions: append([]*proto.APLListItem{{Action: aplCast(dotID), Hide: true}}, aplItems(
					&proto.APLAction{Action: &proto.APLAction_Sequence{Sequence: &proto.APLActionSequence{
						Actions: []*proto.APLAction{aplCast(unknownID), aplCast(dotID), aplCast(dotID)},
					}}},
				)...)},
			},
		}
	}

	request := fakeRageSimRequest()
	request.Encounter.Targets = append(request.Encounter.Targets, request.Encounter.Targets[0])
	sim := setupFakeRageSimFrom(request)
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	sim.Environment.State = Constructed
	fw.RegisterSpell(SpellConfig{
		ActionID: ActionID{SpellID: dotID},
		Flags:    SpellFlagAPL,
		Cast:     CastConfig{DefaultCast: Cast{GCD: GCDDefault}},
		Dot:      DotConfig{Aura: Aura{Label: "Test Dot"}, NumberOfTicks: 5, TickLength: 3 * time.Second},
	})

	want := func(action *APLAction, entryPath string) {
		t.Helper()
		if action.entryPath != entryPath {
			t.Errorf("%s has entry path %q, want %q", action.impl, action.entryPath, entryPath)
		}
	}
	sequenceOf := func(ref *APLAction) *APLActionSequence {
		return ref.impl.(*APLActionGroupReference).group.actions[0].impl.(*APLActionSequence)
	}

	rot := fw.newAPLRotation(config(proto.APLRotation_TypeAPL))
	want(rot.priorityList[0], "priority 2")
	steps := sequenceOf(rot.priorityList[1]).subactions
	if len(steps) != 2 {
		t.Fatalf("sequence kept %d steps, want 2", len(steps))
	}
	want(steps[0], "priority 3 > group 2.2 > step 2")
	want(steps[1], "priority 3 > group 2.2 > step 3")
	want(rot.priorityList[2].impl.(*APLActionGroupReference).group.actions[0], "priority 4 > group 2.2")
	schedule := rot.priorityList[3]
	want(schedule, "priority 5")
	want(schedule.impl.(*APLActionSchedule).innerAction, "priority 5")
	multidot := rot.priorityList[4].impl.(*APLActionStrictMultidot)
	if len(multidot.actions) != 2 {
		t.Fatalf("strict multidot has %d actions, want 2", len(multidot.actions))
	}
	for _, action := range multidot.actions {
		want(action, "priority 6")
	}
	nested := rot.priorityList[5].impl.(*APLActionGroupReference).group.actions[0]
	want(nested, "priority 7 > group 1.2")
	want(nested.impl.(*APLActionGroupReference).group.actions[0], "priority 7 > group 1.2 > group 2.2")
	nestedSteps := sequenceOf(nested).subactions
	if len(nestedSteps) != 2 {
		t.Fatalf("nested sequence kept %d steps, want 2", len(nestedSteps))
	}
	want(nestedSteps[0], "priority 7 > group 1.2 > group 2.2 > step 2")
	want(nestedSteps[1], "priority 7 > group 1.2 > group 2.2 > step 3")
	want(rot.prepullActions[0], "prepull 2")

	want(fw.newCustomRotation().priorityList[0], "")

	untyped := fw.newAPLRotation(config(proto.APLRotation_TypeUnknown))
	want(untyped.priorityList[0], "priority 2")
	want(untyped.prepullActions[0], "prepull 2")

	for _, rotationType := range []proto.APLRotation_Type{proto.APLRotation_TypeAuto, proto.APLRotation_TypeSimple} {
		generated := fw.newAPLRotation(config(rotationType))
		actions := append(generated.allAPLActions(), generated.allPrepullActions()...)
		actions = append(actions, generated.priorityList[4].impl.(*APLActionStrictMultidot).actions...)
		for _, action := range actions {
			want(action, "")
		}
	}
}

// A logged fake warrior sim at the pull, running this APL rotation with the spells and major cooldowns
// register adds, its melee swings off. Every line the run logs is kept, formatted.
type entryPathTest struct {
	*channelTest
	logs []string
}

func newEntryPathTest(register func(char *Character), config *proto.APLRotation) *entryPathTest {
	sim := setupFakeRageSimFrom(fakeRageSimRequest())
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	known := len(fw.Spellbook)
	sim.Environment.State = Constructed
	register(&fw.Character)
	sim.Environment.State = Finalized
	for _, spell := range fw.Spellbook[known:] {
		spell.finalize()
	}
	fw.majorCooldownManager.finalize()
	config.Type = proto.APLRotation_TypeAPL
	fw.Rotation = fw.newAPLRotation(config)
	sim.Cleanup()
	sim.Reset()

	lt := &entryPathTest{channelTest: &channelTest{sim: sim, fw: fw, rot: fw.Rotation}}
	sim.Log = func(message string, vals ...any) {
		lt.logs = append(lt.logs, fmt.Sprintf(message, vals...))
	}
	sim.PrePull()
	sim.Step()
	fw.AutoAttacks.CancelAutoSwing(sim)
	return lt
}

// The spell's Casting lines, in the order they were logged.
func (lt *entryPathTest) casting(spell *Spell) []string {
	return FilterSlice(lt.logs, func(line string) bool {
		return strings.Contains(line, fmt.Sprintf("Casting %s (", spell.ActionID))
	})
}

// Whether the spell was queued to go off later.
func (lt *entryPathTest) queued(spell *Spell) bool {
	return len(FilterSlice(lt.logs, func(line string) bool {
		return strings.Contains(line, fmt.Sprintf("Queueing up %s ", spell.ActionID))
	})) > 0
}

func aplCast(id int32) *proto.APLAction {
	return &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: ActionID{SpellID: id}.ToProto()}}}
}

func aplItems(actions ...*proto.APLAction) []*proto.APLListItem {
	return MapSlice(actions, func(action *proto.APLAction) *proto.APLListItem { return &proto.APLListItem{Action: action} })
}

func wantEntryPaths(t *testing.T, name string, got []string, suffixes ...string) {
	t.Helper()
	if len(got) < len(suffixes) {
		t.Fatalf("%s logged %d Casting lines, want at least %d: %q", name, len(got), len(suffixes), got)
	}
	for i, suffix := range suffixes {
		if !strings.HasSuffix(got[i], suffix) {
			t.Errorf("%s's Casting line %d is %q, want it to end %q", name, i+1, got[i], suffix)
		}
	}
}

// A cast the rotation starts ends its Casting line with the path of the entry that started it, after the
// text the line has always had.
func TestRotationCastLogsItsEntryPath(t *testing.T) {
	var nuke *Spell
	hidden := &proto.APLListItem{Action: aplCast(910110), Hide: true}
	lt := newEntryPathTest(func(char *Character) {
		nuke = char.RegisterSpell(nukeConfig(&[]time.Duration{}))
	}, &proto.APLRotation{PriorityList: append([]*proto.APLListItem{hidden}, aplItems(aplCast(910110))...)})

	lt.runUntil(0)
	want := fmt.Sprintf("[%s] Casting %s (Cost = 0.000, Cast Time = 0s, GCD = 1.5s, Effective Time = 1.5s) APL: priority 2", lt.fw.Label, nuke.ActionID)
	if got := lt.casting(nuke); len(got) != 1 || got[0] != want {
		t.Fatalf("the nuke's Casting lines are %q, want [%q]", got, want)
	}
}

// A spell another spell's effects cast is not the rotation's cast, and its Casting line names no entry path.
func TestCastInsideARotationCastLogsNoEntryPath(t *testing.T) {
	var outer, inner *Spell
	lt := newEntryPathTest(func(char *Character) {
		inner = char.RegisterSpell(SpellConfig{ActionID: ActionID{SpellID: 910121}, ApplyEffects: func(*Simulation, *Unit, *Spell) {}})
		config := nukeConfig(&[]time.Duration{})
		config.ApplyEffects = func(sim *Simulation, target *Unit, _ *Spell) {
			inner.Cast(sim, target)
		}
		outer = char.RegisterSpell(config)
	}, &proto.APLRotation{PriorityList: aplItems(aplCast(910110))})

	lt.runUntil(0)
	wantEntryPaths(t, "the outer spell", lt.casting(outer), ") APL: priority 1")
	if got := lt.casting(inner); len(got) != 1 || strings.Contains(got[0], " APL: ") {
		t.Fatalf("the inner spell's Casting lines are %q, want one that names no entry path", got)
	}
}

// A queued spell goes off with no entry running; its Casting line names the entry that queued it.
func TestQueuedCastCarriesTheEntryPathThatQueuedIt(t *testing.T) {
	t.Run("a spell picked at a hardcast's end", func(t *testing.T) {
		var slow, nuke *Spell
		firstSecond := &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
			Op:  proto.APLValueCompare_OpLt,
			Lhs: &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}},
			Rhs: constValue("1s"),
		}}}
		slowCast := aplCast(910122)
		slowCast.Condition = firstSecond
		lt := newEntryPathTest(func(char *Character) {
			config := recordedSpellConfig(910122, &[]time.Duration{})
			config.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: 2 * time.Second}}
			slow = char.RegisterSpell(config)
			nuke = char.RegisterSpell(nukeConfig(&[]time.Duration{}))
		}, &proto.APLRotation{PriorityList: aplItems(slowCast, aplCast(910110))})

		lt.runUntil(2*time.Second + 1)
		wantEntryPaths(t, "the hardcast", lt.casting(slow), ") APL: priority 1")
		if !lt.queued(nuke) {
			t.Fatalf("the nuke should be queued at the 2 s cast end, logs: %q", lt.logs)
		}
		wantEntryPaths(t, "the queued nuke", lt.casting(nuke), ") APL: priority 2")
	})

	// The nuke at the pull starts the GCD; the second spell's entries, priority 2 and then 3, press it 100 ms
	// apart from 1.2 s and queue it for the GCD end at 1.5 s, with the rotation's own look held back past it.
	inQueueWindow := func(entries int) (*entryPathTest, *Spell) {
		var second *Spell
		lt := newEntryPathTest(func(char *Character) {
			config := nukeConfig(&[]time.Duration{})
			config.Cast.CD = Cooldown{Timer: char.NewTimer(), Duration: time.Minute}
			char.RegisterSpell(config)
			secondConfig := recordedSpellConfig(910123, &[]time.Duration{})
			secondConfig.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault}}
			second = char.RegisterSpell(secondConfig)
		}, &proto.APLRotation{PriorityList: aplItems(aplCast(910110), aplCast(910123), aplCast(910123))})
		for i := 1; i <= entries; i++ {
			lt.runUntil(1200*time.Millisecond + time.Duration(i-1)*100*time.Millisecond)
			lt.rot.priorityList[i].Execute(lt.sim)
		}
		lt.fw.SetRotationTimer(lt.sim, 1550*time.Millisecond)
		if lt.fw.QueuedSpell == nil || lt.fw.QueuedSpell.spell != second {
			t.Fatalf("the second spell should be queued for the GCD end at 1.5 s")
		}
		lt.runUntil(1501 * time.Millisecond)
		return lt, second
	}

	t.Run("a spell queued in the spell queue window", func(t *testing.T) {
		lt, second := inQueueWindow(1)
		if got := lt.casting(second); len(got) != 1 || !strings.HasSuffix(got[0], ") APL: priority 2") {
			t.Fatalf("the queued spell's Casting lines are %q, want one ending \") APL: priority 2\"", got)
		}
	})

	t.Run("a spell queued again by another entry", func(t *testing.T) {
		lt, second := inQueueWindow(2)
		if got := lt.casting(second); len(got) != 1 || !strings.HasSuffix(got[0], ") APL: priority 3") {
			t.Fatalf("the spell queued again at 1.3 s has Casting lines %q, want one ending \") APL: priority 3\"", got)
		}
	})
}

// Steps after the first run in later passes of the rotation, outside the sequence's own run, and each
// logs its own step.
func TestStrictSequenceLaterStepsLogTheirStep(t *testing.T) {
	var nuke *Spell
	sequence := &proto.APLAction{Action: &proto.APLAction_StrictSequence{StrictSequence: &proto.APLActionStrictSequence{
		Actions: []*proto.APLAction{aplCast(910110), aplCast(910110), aplCast(910110)},
	}}}
	lt := newEntryPathTest(func(char *Character) {
		nuke = char.RegisterSpell(nukeConfig(&[]time.Duration{}))
	}, &proto.APLRotation{PriorityList: aplItems(sequence)})

	lt.runUntil(3 * time.Second)
	wantEntryPaths(t, "the nuke", lt.casting(nuke), ") APL: priority 1 > step 1", ") APL: priority 1 > step 2", ") APL: priority 1 > step 3")
}

// A major cooldown Autocast Other Cooldowns fires names the Autocast Other Cooldowns entry.
func TestAutocastCooldownLogsTheAutocastEntryPath(t *testing.T) {
	var cooldown *Spell
	hidden := &proto.APLListItem{Action: aplCast(910110), Hide: true}
	autocast := &proto.APLAction{Action: &proto.APLAction_AutocastOtherCooldowns{AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{}}}
	lt := newEntryPathTest(func(char *Character) {
		char.RegisterSpell(nukeConfig(&[]time.Duration{}))
		config := recordedSpellConfig(910124, &[]time.Duration{})
		config.Cast = CastConfig{CD: Cooldown{Timer: char.NewTimer(), Duration: time.Minute}}
		cooldown = char.RegisterSpell(config)
		char.AddMajorCooldown(MajorCooldown{Spell: cooldown, Type: CooldownTypeDPS})
	}, &proto.APLRotation{PriorityList: append([]*proto.APLListItem{hidden}, aplItems(autocast)...)})

	lt.runUntil(0)
	wantEntryPaths(t, "the cooldown", lt.casting(cooldown), ") APL: priority 2")
}

// A press that queues an attack for the next swing credits it, and the swing's cast of the attack names the
// press's entry path. A later swing that casts it again, with no new press, names none.
func TestNextSwingAttackCarriesTheQueuingEntryPath(t *testing.T) {
	var press, attack *Spell
	armed := false
	lt := newEntryPathTest(func(char *Character) {
		attack = char.RegisterSpell(SpellConfig{ActionID: ActionID{SpellID: 910131}, ApplyEffects: func(*Simulation, *Unit, *Spell) {}})
		press = char.RegisterSpell(SpellConfig{
			ActionID: ActionID{SpellID: 910130},
			Flags:    SpellFlagAPL,
			ExtraCastCondition: func(*Simulation, *Unit) bool {
				return !armed
			},
			ApplyEffects: func(_ *Simulation, target *Unit, _ *Spell) {
				attack.CountManualCast(target)
				armed = true
			},
		})
	}, &proto.APLRotation{PriorityList: aplItems(aplCast(910130))})
	lt.fw.AutoAttacks.SetReplaceMHSwing(func(_ *Simulation, swing *Spell) *Spell {
		if armed {
			return attack
		}
		return swing
	})
	lt.fw.AutoAttacks.EnableAutoSwing(lt.sim)

	lt.runUntil(6 * time.Second)
	wantEntryPaths(t, "the press", lt.casting(press), ") APL: priority 1")
	got := lt.casting(attack)
	wantEntryPaths(t, "the queued attack", got, ") APL: priority 1")
	if len(got) < 2 || strings.Contains(got[1], " APL: ") {
		t.Fatalf("the attack's Casting lines are %q, want a second that names no entry path", got)
	}
}

// A walk the rotation picks at a hardcast's end is no press of the hardcast, so no later walk names the
// hardcast's entry path.
func TestWalkAtHardcastEndTakesNoEntryPath(t *testing.T) {
	var slow *Spell
	slowCast := aplCast(910122)
	slowCast.Condition = &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
		Op:  proto.APLValueCompare_OpLt,
		Lhs: &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}},
		Rhs: constValue("1s"),
	}}}
	walk := &proto.APLAction{Action: &proto.APLAction_MoveDuration{MoveDuration: &proto.APLActionMoveDuration{Duration: constValue("500ms")}}}
	lt := newEntryPathTest(func(char *Character) {
		config := recordedSpellConfig(910122, &[]time.Duration{})
		config.Flags |= SpellFlagCanCastWhileMoving
		config.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: 2 * time.Second}}
		slow = char.RegisterSpell(config)
	}, &proto.APLRotation{PriorityList: aplItems(slowCast, walk)})

	lt.runUntil(3 * time.Second)
	wantEntryPaths(t, "the hardcast", lt.casting(slow), ") APL: priority 1")
	walks := lt.casting(lt.fw.moveSpell)
	if len(walks) < 2 {
		t.Fatalf("the walks' Casting lines are %q, want at least 2", walks)
	}
	for _, line := range walks {
		if strings.Contains(line, " APL: ") {
			t.Fatalf("the walks' Casting lines are %q, want none that names an entry path", walks)
		}
	}
}

func TestWalkInsideQueuedCastTakesNoEntryPath(t *testing.T) {
	var queued *Spell
	qid := ActionID{SpellID: 910123}.ToProto()
	late := &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
		Op:  proto.APLValueCompare_OpGt,
		Lhs: &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}},
		Rhs: constValue("1.3s"),
	}}}
	notReady := &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: &proto.APLValue{
		Value: &proto.APLValue_SpellIsReady{SpellIsReady: &proto.APLValueSpellIsReady{SpellId: qid}},
	}}}}
	cond := &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: []*proto.APLValue{late, notReady}}}}
	walk := &proto.APLAction{Condition: cond, Action: &proto.APLAction_MoveDuration{MoveDuration: &proto.APLActionMoveDuration{Duration: constValue("500ms")}}}
	lt := newEntryPathTest(func(char *Character) {
		config := recordedSpellConfig(910123, &[]time.Duration{})
		config.Cast = CastConfig{CD: Cooldown{Timer: char.NewTimer(), Duration: 1400 * time.Millisecond}}
		config.ApplyEffects = func(sim *Simulation, _ *Unit, _ *Spell) {
			char.ReactToEvent(sim, false, false)
		}
		queued = char.RegisterSpell(config)
	}, &proto.APLRotation{PriorityList: aplItems(aplCast(910123), walk)})

	lt.runUntil(1200 * time.Millisecond)
	lt.rot.priorityList[0].Execute(lt.sim)
	if !lt.queued(queued) {
		t.Fatalf("the spell was not queued")
	}
	lt.runUntil(4 * time.Second)
	walks := lt.casting(lt.fw.moveSpell)
	if len(walks) == 0 {
		t.Fatalf("no walk was picked")
	}
	for _, line := range walks {
		if strings.Contains(line, " APL: ") {
			t.Fatalf("the walks' Casting lines are %q, want none that names an entry path", walks)
		}
	}
}

// With logging off the rotation keeps no running entry path, so a cast reads none while its entry runs.
func TestUnloggedRunTracksNothing(t *testing.T) {
	run := func(logged bool) []string {
		var seen []string
		lt := newEntryPathTest(func(char *Character) {
			config := nukeConfig(&[]time.Duration{})
			config.ApplyEffects = func(_ *Simulation, _ *Unit, spell *Spell) {
				seen = append(seen, spell.Unit.Rotation.runningEntryPath)
			}
			char.RegisterSpell(config)
		}, &proto.APLRotation{PriorityList: aplItems(aplCast(910110))})
		if !logged {
			lt.sim.Log = nil
		}
		lt.runUntil(5 * time.Second)
		return seen
	}

	if seen := run(true); len(seen) == 0 || seen[0] != "priority 1" {
		t.Fatalf("a logged cast read the running entry paths %q, want \"priority 1\"", seen)
	}
	seen := run(false)
	if len(seen) < 3 {
		t.Fatalf("the unlogged run cast %d times, want at least 3", len(seen))
	}
	for _, entryPath := range seen {
		if entryPath != "" {
			t.Fatalf("an unlogged cast read the running entry paths %q, want none", seen)
		}
	}
}
