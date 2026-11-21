package core

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A request whose boss the first tank slot tanks, a tank the sim does not model, with the fake Protection Warrior
// in the second.
func planOffTankRequest(plan *proto.EncounterPlan) *proto.RaidSimRequest {
	request := planTankRequest(plan)
	request.Raid.Tanks = []*proto.UnitReference{{}, {Type: proto.UnitReference_Player, Index: 0}}
	return request
}

func testTankSwap(start, end float64, repeat *proto.PlanRepeat) *proto.PlanTankSwap {
	startAt, endAt := planWindow(start, end)
	return &proto.PlanTankSwap{Start: startAt, End: endAt, Repeat: repeat}
}

type planSpan struct{ from, to time.Duration }

func (span planSpan) holds(at time.Duration) bool {
	return at >= span.from && at < span.to
}

// The times on the boss's swing grid through first, up to the end, that fall inside the spans or, with inside
// false, outside every one of them.
func planSwingGrid(first, end time.Duration, spans []planSpan, inside bool) []time.Duration {
	var grid []time.Duration
	for at := first; at < end; at += planBossSwingSpeed {
		if slices.ContainsFunc(spans, func(span planSpan) bool { return span.holds(at) }) == inside {
			grid = append(grid, at)
		}
	}
	return grid
}

func planHitsBefore(hits []planHitTaken, end time.Duration) []planHitTaken {
	var before []planHitTaken
	for _, hit := range hits {
		if hit.at < end {
			before = append(before, hit)
		}
	}
	return before
}

func isPlanSpellHit(index int) func(ActionID) bool {
	tag := planBossAbilityTag + int32(index)
	return func(id ActionID) bool { return id.SpellID == planSpellHitSpellID && id.Tag == tag }
}

// The boss the player tanks swings at the other tank through the window on its running timer, so the player
// takes every swing of the grid outside it and none inside, and the first one after it lands on the grid. A Strike
// armed inside the window is spent there, a Spell Hit cast inside it does not reach the player, and an Area Hit
// inside it still does.
func TestPlanTankSwapTakesTheBossOffItsTank(t *testing.T) {
	window := planSpan{planAt(0.21), planAt(0.35)} // Off the swing grid, so no swing ties with an edge.
	strike := planBossWindow(proto.PlanBossAbility_Strike, 0.25, 1)
	strike.WeaponPercent = 150
	plan := &proto.EncounterPlan{
		TankSwaps: []*proto.PlanTankSwap{testTankSwap(0.21, 0.35, nil)},
		BossAbilities: []*proto.PlanBossAbility{
			strike,
			planNamedSpellHit("", 0, 0.27, 0),
			planNamedSpellHit("", 0, 0.5, 0),
			planAreaHit(0.3),
		},
	}
	sim, hits := runPlanHitsTaken(t, planTankRequest(plan))
	end := sim.Duration - planBossSwingSpeed

	swings := planHitsBefore(planHitsOf(hits, isBossMeleeSwing), end)
	if len(swings) == 0 || swings[0].at >= window.from {
		t.Fatalf("test setup: expected the boss to swing at the player before the window, got %v", swings)
	}
	expectPlanHitTimes(t, "swings", swings, planSwingGrid(swings[0].at, end, []planSpan{window}, false))
	if strikes := planHitsOf(hits, func(id ActionID) bool { return id.SpellID == planStrikeSpellID }); len(strikes) != 0 {
		t.Errorf("expected the strike spent on the other tank inside the window, got %v", strikes)
	}
	expectPlanHitTimes(t, "spell hit inside the window", planHitsOf(hits, isPlanSpellHit(1)), nil)
	expectPlanHitTimes(t, "spell hit after the window", planHitsOf(hits, isPlanSpellHit(2)), []time.Duration{planAt(0.5)})
	expectPlanHitTimes(t, "area hit inside the window", planAreaHitsOf(hits, 3), []time.Duration{planAt(0.3)})
}

// A Spell Hit at the other tank is still a hardcast: it holds the swing through the cast, and the first swing
// after the window lands one full swing after the cast's end.
func TestPlanTankSwapSpellHitAtTheOtherTankHoldsTheSwing(t *testing.T) {
	const castTime = 5 * time.Second
	castStart := planAt(0.251)
	plan := &proto.EncounterPlan{
		TankSwaps:     []*proto.PlanTankSwap{testTankSwap(0.21, 0.35, nil)},
		BossAbilities: []*proto.PlanBossAbility{planNamedSpellHit("", 0, 0.251, castTime)},
	}
	_, hits := runPlanHitsTaken(t, planTankRequest(plan))

	expectPlanHitTimes(t, "spell hit", planHitsOf(hits, isPlanSpellHit(0)), nil)
	swings := planHitsOf(hits, isBossMeleeSwing)
	next := slices.IndexFunc(swings, func(hit planHitTaken) bool { return hit.at >= planAt(0.35) })
	if want := castStart + castTime + 3*planBossSwingSpeed; next < 0 || swings[next].at != want {
		t.Errorf("expected the first swing after the window on the timer the cast set, at %s; got %v", want, swings)
	}
}

// A window runs from its start up to its end for a Spell Hit on either edge: one at the start goes to the other tank
// and one at the end to the target's own.
func TestPlanTankSwapSpellHitOnAWindowEdge(t *testing.T) {
	plan := &proto.EncounterPlan{
		TankSwaps: []*proto.PlanTankSwap{testTankSwap(0.21, 0.35, nil)},
		BossAbilities: []*proto.PlanBossAbility{
			planNamedSpellHit("", 0, 0.21, 0),
			planNamedSpellHit("", 0, 0.35, 0),
		},
	}

	t.Run("tanked by the player", func(t *testing.T) {
		_, hits := runPlanHitsTaken(t, planTankRequest(plan))
		expectPlanHitTimes(t, "spell hit at the start", planHitsOf(hits, isPlanSpellHit(0)), nil)
		expectPlanHitTimes(t, "spell hit at the end", planHitsOf(hits, isPlanSpellHit(1)), []time.Duration{planAt(0.35)})
	})

	t.Run("tanked by another tank slot", func(t *testing.T) {
		_, hits := runPlanHitsTaken(t, planOffTankRequest(plan))
		expectPlanHitTimes(t, "spell hit at the start", planHitsOf(hits, isPlanSpellHit(0)), []time.Duration{planAt(0.21)})
		expectPlanHitTimes(t, "spell hit at the end", planHitsOf(hits, isPlanSpellHit(1)), nil)
	})
}

// The boss another tank slot tanks swings from the pull at that tank, so the player in a tank slot of their own
// takes its swings only inside each window, on the one grid its pull swing set, which runs through both. The
// player counts as tanking, a Spell Hit reaches them only inside a window, a Strike armed outside one is spent on
// the other tank while one armed inside lands on them, a Thrash's extra attacks follow the swings onto them, and an
// Area Hit reaches them outside a window as well.
func TestPlanTankSwapBringsTheBossOntoTheOffTank(t *testing.T) {
	windows := []planSpan{{planAt(0.21), planAt(0.35)}, {planAt(0.61), planAt(0.75)}}
	strike := func(start float64) *proto.PlanBossAbility {
		ability := planBossWindow(proto.PlanBossAbility_Strike, start, 1)
		ability.WeaponPercent = 150
		return ability
	}
	plan := &proto.EncounterPlan{
		TankSwaps: []*proto.PlanTankSwap{testTankSwap(0.21, 0.35, &proto.PlanRepeat{Interval: 0.4, Count: 1})},
		BossAbilities: []*proto.PlanBossAbility{
			planNamedSpellHit("", 0, 0.1, 0),
			planNamedSpellHit("", 0, 0.3, 0),
			planAreaHit(0.1),
			planThrash(0.15, 0.4),
			strike(0.1),
			strike(0.25),
		},
	}
	sim, hits := setupPlanHitsTaken(planOffTankRequest(plan), nil)
	fw := sim.Raid.Parties[0].Players[0].(*FakeFearWarrior)
	if !fw.Metrics.isTanking {
		t.Error("expected the player a swap brings the boss onto to count as tanking")
	}
	if fw.HardcastAvoidanceAura == nil {
		t.Error("expected the player a swap brings the boss onto to have the tank's reduced avoidance aura")
	}
	boss := sim.Encounter.AllTargets[0]
	runLifecycleUntil(sim, planAt(0.2))
	if !boss.AutoAttacks.mh.enabled {
		t.Fatal("expected the boss swinging at the other tank before the window")
	}
	pull := boss.AutoAttacks.mh.swingAt % planBossSwingSpeed
	bossTargetsMe := fw.Rotation.newValueBossCurrentTarget(&proto.APLValueBossCurrentTarget{
		TargetUnit: &proto.UnitReference{Type: proto.UnitReference_Target, Index: 0},
	}, nil)
	for _, step := range []struct {
		at   time.Duration
		want bool
	}{{planAt(0.2), false}, {planAt(0.3), true}, {planAt(0.4), false}} {
		runLifecycleUntil(sim, step.at)
		if got := bossTargetsMe.GetBool(sim); got != step.want {
			t.Errorf("Boss Current Target at %s: got %t, want %t", step.at, got, step.want)
		}
	}
	runLifecycleUntil(sim, sim.Duration)

	isStrike := func(id ActionID) bool { return id.SpellID == planStrikeSpellID }
	grid := planSwingGrid(pull, sim.Duration, windows, true)
	attacks := planHitsOf(*hits, func(id ActionID) bool { return isBossMeleeSwing(id) || isStrike(id) })
	expectPlanHitTimes(t, "swings and strikes", attacks, grid)
	strikeAt := grid[slices.IndexFunc(grid, func(at time.Duration) bool { return at >= planAt(0.25) })]
	expectPlanHitTimes(t, "strikes", planHitsOf(*hits, isStrike), []time.Duration{strikeAt})
	expectPlanHitTimes(t, "spell hit outside a window", planHitsOf(*hits, isPlanSpellHit(0)), nil)
	expectPlanHitTimes(t, "spell hit inside a window", planHitsOf(*hits, isPlanSpellHit(1)), []time.Duration{planAt(0.3)})
	expectPlanHitTimes(t, "area hit outside a window", planAreaHitsOf(*hits, 2), []time.Duration{planAt(0.1)})
	extra := planHitsOf(*hits, func(id ActionID) bool {
		return id.OtherID == proto.OtherAction_OtherActionAttack && id.Tag == planBossAbilityTag+3
	})
	if len(extra) == 0 {
		t.Error("expected extra attacks on the player inside the window")
	}
	for _, hit := range extra {
		if !windows[0].holds(hit.at) {
			t.Errorf("extra attack on the player at %s, outside the window %v", hit.at, windows[0])
		}
	}
}

// A downtime that ends inside a window brings the boss back still swapped: the boss the player tanks stays off them
// until the window ends, and the boss another tank slot tanks comes back swinging at the player until then.
func TestPlanTankSwapHoldsThroughADowntime(t *testing.T) {
	window := planSpan{planAt(0.21), planAt(0.45)}
	plan := &proto.EncounterPlan{
		TankSwaps: []*proto.PlanTankSwap{testTankSwap(0.21, 0.45, nil)},
		Downtimes: []*proto.PlanDowntime{{TargetIndex: 0, Start: &proto.PlanTime{At: 0.25}, End: &proto.PlanTime{At: 0.3}}},
	}

	t.Run("tanked by the player", func(t *testing.T) {
		_, hits := runPlanHitsTaken(t, planTankRequest(plan))
		swings := planHitsOf(hits, isBossMeleeSwing)
		if got := bossSwingsBetween(hits, window.from, window.to); got != 0 {
			t.Errorf("expected no swing on the player inside the window, the downtime's end included; got %d", got)
		}
		if got := bossSwingsBetween(hits, window.to-time.Nanosecond, window.to+planBossSwingSpeed); got != 1 {
			t.Errorf("expected the swing back on the player once the window ends, got %d in its first swing: %v", got, swings)
		}
	})

	t.Run("tanked by another tank slot", func(t *testing.T) {
		_, hits := runPlanHitsTaken(t, planOffTankRequest(plan))
		swings := planHitsOf(hits, isBossMeleeSwing)
		back := planAt(0.3)
		if got := bossSwingsBetween(hits, back-time.Nanosecond, window.to); got == 0 {
			t.Errorf("expected the boss back from its downtime to swing at the player inside the window, got %v", swings)
		}
		for _, swing := range swings {
			if swing.at >= planAt(0.25) && swing.at < back {
				t.Errorf("swing at %s while the boss is down", swing.at)
			}
			if !window.holds(swing.at) {
				t.Errorf("swing on the player at %s, outside the window %v", swing.at, window)
			}
		}
	})
}

// A window still open at the fight's end is gone by the next iteration, which plays out as the first did: the
// boss swings at the player from its pull swing on, every swing of the grid up to the window.
func TestPlanTankSwapStartsEachIterationClean(t *testing.T) {
	window := []planSpan{{planAt(0.9), planTestDuration}}
	plan := &proto.EncounterPlan{TankSwaps: []*proto.PlanTankSwap{testTankSwap(0.9, 1, nil)}}
	sim, hits := setupPlanHitsTaken(planTankRequest(plan), nil)
	state := &sim.Encounter.plan.targets[0]

	for iteration := range 2 {
		if state.swapWindows != 0 {
			t.Fatalf("iteration %d: %d windows running at the pull, expected none", iteration+1, state.swapWindows)
		}
		*hits = (*hits)[:0]
		runLifecycleUntil(sim, sim.Duration)
		if state.swapWindows != 1 {
			t.Fatalf("iteration %d, test setup: expected the window open at the fight's end, got %d running", iteration+1, state.swapWindows)
		}
		swings := planHitsOf(*hits, isBossMeleeSwing)
		if len(swings) == 0 || swings[0].at >= planBossSwingSpeed {
			t.Fatalf("iteration %d: expected the boss's first swing on the player within a swing of the pull, got %v", iteration+1, swings)
		}
		expectPlanHitTimes(t, fmt.Sprintf("iteration %d swings", iteration+1), swings, planSwingGrid(swings[0].at, sim.Duration, window, false))
		sim.Cleanup()
		sim.Reset()
		sim.PrePull()
	}
}

// A plan whose only event is a tank swap is a plan, and its swap acts. A swap on a target with neither tank in the
// sim does nothing, so a Spell Hit inside it reaches the player as one with no tank does.
func TestPlanTankSwapAloneIsAPlan(t *testing.T) {
	plan := &proto.EncounterPlan{TankSwaps: []*proto.PlanTankSwap{testTankSwap(0.21, 0.35, nil)}}
	sim, hits := runPlanHitsTaken(t, planTankRequest(plan))
	if sim.Encounter.plan == nil {
		t.Fatal("expected a plan with only a tank swap to run")
	}
	if got := bossSwingsBetween(hits, planAt(0.21), planAt(0.35)); got != 0 {
		t.Errorf("expected no swing on the player inside the window, got %d", got)
	}

	noTank := &proto.EncounterPlan{
		TankSwaps:     []*proto.PlanTankSwap{testTankSwap(0.21, 0.35, nil)},
		BossAbilities: []*proto.PlanBossAbility{planNamedSpellHit("", 0, 0.3, 0)},
	}
	sim, hits = runPlanHitsTaken(t, planTestRequest(1, noTank))
	if !sim.Encounter.plan.tankSwaps[0].skip {
		t.Error("expected the swap between two tanks outside the sim skipped")
	}
	expectPlanHitTimes(t, "spell hit without a tank", planHitsOf(hits, isPlanSpellHit(0)), []time.Duration{planAt(0.3)})

	untanked := planOffTankRequest(noTank)
	untanked.Encounter.Targets[0].TankIndex = -1
	sim, hits = runPlanHitsTaken(t, untanked)
	if !sim.Encounter.plan.tankSwaps[0].skip {
		t.Error("expected the swap on a target no tank slot tanks skipped")
	}
	if got := bossSwingsBetween(hits, 0, sim.Duration); got != 0 {
		t.Errorf("expected the untanked boss never to swing at the player in a tank slot, got %d", got)
	}
	expectPlanHitTimes(t, "spell hit on an untanked boss", planHitsOf(hits, isPlanSpellHit(0)), []time.Duration{planAt(0.3)})
}
