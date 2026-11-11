package core

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

func planAmpWindow(ampType proto.DamageAmpType, percent, start, end float64) *proto.PlanDamageAmp {
	return &proto.PlanDamageAmp{
		Type:    ampType,
		Percent: percent,
		Start:   &proto.PlanTime{At: start},
		End:     &proto.PlanTime{At: end},
	}
}

func expectMultiplier(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("%s: expected %v, got %v", what, want, got)
	}
}

func TestPlanCasterAmpRaisesDamageDealtInsideItsWindow(t *testing.T) {
	plan := &proto.EncounterPlan{DamageAmps: []*proto.PlanDamageAmp{
		planAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.2, 0.4),
	}}
	sim, fw := setupPlanSim(planTestRequest(1, plan), nil)
	base := fw.PseudoStats.DamageDealtMultiplier
	caster := proto.DamageAmpType_DamageAmpTypeCasterBuff
	environment := proto.DamageAmpType_DamageAmpTypeEnvironmentBuff

	if sim.Encounter.AmpActive(sim, caster) {
		t.Error("before the window: expected no active caster amp")
	}
	if got := sim.Encounter.NextAmpAt(sim, caster); got != planAt(0.2) {
		t.Errorf("next caster amp before the window: expected %s, got %s", planAt(0.2), got)
	}
	if got := sim.Encounter.NextAmpAt(sim, environment); got != NeverExpires {
		t.Errorf("next environment amp of a plan without one: expected never, got %s", got)
	}

	runLifecycleUntil(sim, planAt(0.2)-time.Millisecond)
	expectMultiplier(t, "damage dealt before the window", fw.PseudoStats.DamageDealtMultiplier, base)

	runLifecycleUntil(sim, planAt(0.2))
	expectMultiplier(t, "damage dealt inside a +10% window", fw.PseudoStats.DamageDealtMultiplier, base*1.1)
	if !sim.Encounter.AmpActive(sim, caster) || sim.Encounter.AmpActive(sim, environment) {
		t.Error("inside the window: expected the caster amp active and the environment amp not")
	}
	if got := sim.Encounter.NextAmpAt(sim, caster); got != sim.CurrentTime {
		t.Errorf("next caster amp inside its window: expected now, got %s", got)
	}
	if aura := fw.GetAura("Damage Done % (Caster Buff)"); aura == nil || !aura.IsActive() {
		t.Error("inside the window: expected the caster buff aura on the player")
	}

	runLifecycleUntil(sim, planAt(0.4))
	expectMultiplier(t, "damage dealt after the window", fw.PseudoStats.DamageDealtMultiplier, base)
	if sim.Encounter.AmpActive(sim, caster) {
		t.Error("after the window: expected no active caster amp")
	}
	if got := sim.Encounter.NextAmpAt(sim, caster); got != NeverExpires {
		t.Errorf("next caster amp after the only window: expected never, got %s", got)
	}
}

func TestPlanAmpWindowsOfOneTypeAddUp(t *testing.T) {
	plan := &proto.EncounterPlan{DamageAmps: []*proto.PlanDamageAmp{
		planAmpWindow(proto.DamageAmpType_DamageAmpTypeEnvironmentBuff, 10, 0.2, 0.6),
		planAmpWindow(proto.DamageAmpType_DamageAmpTypeEnvironmentBuff, 2.5, 0.4, 0.8),
	}}
	sim, fw := setupPlanSim(planTestRequest(1, plan), nil)
	base := fw.PseudoStats.DamageDealtMultiplier

	runLifecycleUntil(sim, planAt(0.3))
	expectMultiplier(t, "first window alone", fw.PseudoStats.DamageDealtMultiplier, base*1.1)
	runLifecycleUntil(sim, planAt(0.5))
	expectMultiplier(t, "both windows", fw.PseudoStats.DamageDealtMultiplier, base*1.125)
	runLifecycleUntil(sim, planAt(0.7))
	expectMultiplier(t, "second window alone", fw.PseudoStats.DamageDealtMultiplier, base*1.025)
	runLifecycleUntil(sim, planAt(0.9))
	expectMultiplier(t, "after both windows", fw.PseudoStats.DamageDealtMultiplier, base)
}

func TestPlanAmpWindowThatEndsAsAnotherStartsKeepsTheSum(t *testing.T) {
	plan := &proto.EncounterPlan{DamageAmps: []*proto.PlanDamageAmp{
		planAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.2, 0.4),
		planAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.4, 0.6),
	}}
	sim, fw := setupPlanSim(planTestRequest(1, plan), nil)
	base := fw.PseudoStats.DamageDealtMultiplier
	caster := proto.DamageAmpType_DamageAmpTypeCasterBuff

	runLifecycleUntil(sim, planAt(0.4))
	expectMultiplier(t, "at the hand-over", fw.PseudoStats.DamageDealtMultiplier, base*1.1)
	if !sim.Encounter.AmpActive(sim, caster) {
		t.Error("at the hand-over: expected the caster amp active")
	}
	runLifecycleUntil(sim, planAt(0.6))
	expectMultiplier(t, "after the second window", fw.PseudoStats.DamageDealtMultiplier, base)
}

func TestPlanAmpRepeatsEveryInterval(t *testing.T) {
	amp := planAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.1, 0.2)
	amp.Repeat = &proto.PlanRepeat{Interval: 0.3, Count: 1}
	sim, fw := setupPlanSim(planTestRequest(1, &proto.EncounterPlan{DamageAmps: []*proto.PlanDamageAmp{amp}}), nil)
	base := fw.PseudoStats.DamageDealtMultiplier
	caster := proto.DamageAmpType_DamageAmpTypeCasterBuff

	runLifecycleUntil(sim, planAt(0.1))
	if got := sim.Encounter.NextAmpAt(sim, caster); got != planAt(0.1) {
		t.Errorf("inside the first window: expected now, got %s", got)
	}
	runLifecycleUntil(sim, planAt(0.25))
	if got := sim.Encounter.NextAmpAt(sim, caster); got != planAt(0.4) {
		t.Errorf("between windows: expected the repeat at %s, got %s", planAt(0.4), got)
	}
	expectMultiplier(t, "between windows", fw.PseudoStats.DamageDealtMultiplier, base)
	runLifecycleUntil(sim, planAt(0.45))
	expectMultiplier(t, "inside the repeat", fw.PseudoStats.DamageDealtMultiplier, base*1.1)
	runLifecycleUntil(sim, planAt(0.6))
	if got := sim.Encounter.NextAmpAt(sim, caster); got != NeverExpires {
		t.Errorf("after the last repeat: expected never, got %s", got)
	}
}

func TestPlanTargetAmpHitsOnlyItsTargets(t *testing.T) {
	amp := planAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0.2, 0.4)
	amp.Targets = &proto.PlanTargets{TargetIndices: []int32{1}}
	sim, _ := setupPlanSim(planTestRequest(3, &proto.EncounterPlan{DamageAmps: []*proto.PlanDamageAmp{amp}}), nil)
	targets := sim.Encounter.AllTargets
	base := targets[1].PseudoStats.DamageTakenMultiplier

	runLifecycleUntil(sim, planAt(0.3))
	for i, want := range []float64{base, base * 1.1, base} {
		expectMultiplier(t, "damage taken of target "+string(rune('0'+i)), targets[i].PseudoStats.DamageTakenMultiplier, want)
	}
	if aura := targets[1].GetAura("Damage Taken % (Target Debuff)"); aura == nil || !aura.IsActive() {
		t.Error("inside the window: expected the debuff aura on the named target")
	}
	if targets[0].GetAura("Damage Taken % (Target Debuff)") != nil {
		t.Error("expected no debuff aura on a target the amp does not name")
	}

	runLifecycleUntil(sim, planAt(0.5))
	expectMultiplier(t, "damage taken after the window", targets[1].PseudoStats.DamageTakenMultiplier, base)
}

func TestPlanTargetAmpWithoutTargetsHitsTheBoss(t *testing.T) {
	amp := planAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 20, 0.2, 0.4)
	sim, _ := setupPlanSim(planTestRequest(2, &proto.EncounterPlan{DamageAmps: []*proto.PlanDamageAmp{amp}}), nil)
	targets := sim.Encounter.AllTargets
	base := targets[0].PseudoStats.DamageTakenMultiplier

	runLifecycleUntil(sim, planAt(0.3))
	expectMultiplier(t, "damage taken of the boss", targets[0].PseudoStats.DamageTakenMultiplier, base*1.2)
	expectMultiplier(t, "damage taken of the second target", targets[1].PseudoStats.DamageTakenMultiplier, base)
}

func TestPlanTargetAmpReachesAddsThatSpawnInsideTheWindow(t *testing.T) {
	amp := planAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0.2, 0.6)
	amp.Targets = &proto.PlanTargets{WaveIndices: []int32{0}}
	plan := &proto.EncounterPlan{
		AddWaves:   []*proto.AddWave{testAddWave(1, 0.3, 0.1, nil)},
		DamageAmps: []*proto.PlanDamageAmp{amp},
	}
	sim, _ := setupPlanSim(planTestRequest(1, plan), nil)
	boss, add := sim.Encounter.AllTargets[0], sim.Encounter.AllTargets[1]
	base := boss.PseudoStats.DamageTakenMultiplier

	runLifecycleUntil(sim, planAt(0.25))
	expectMultiplier(t, "boss inside a window naming only the wave", boss.PseudoStats.DamageTakenMultiplier, base)

	runLifecycleUntil(sim, planAt(0.3))
	if !add.IsEnabled() {
		t.Fatal("at the spawn: expected the add active")
	}
	expectMultiplier(t, "add spawned inside the window", add.PseudoStats.DamageTakenMultiplier, base*1.1)
	expectMultiplier(t, "boss with the add up", boss.PseudoStats.DamageTakenMultiplier, base)

	runLifecycleUntil(sim, planAt(0.7))
	expectMultiplier(t, "add after the window", add.PseudoStats.DamageTakenMultiplier, base)
}

func TestPlanAmpsRunEveryIteration(t *testing.T) {
	amp := planAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0.2, 2)
	amp.Targets = &proto.PlanTargets{WaveIndices: []int32{0}}
	caster := planAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 5, 0.5, 1)
	plan := &proto.EncounterPlan{
		AddWaves:   []*proto.AddWave{testAddWave(1, 0.3, 0.1, nil)},
		DamageAmps: []*proto.PlanDamageAmp{amp, caster},
	}
	sim, fw := setupPlanSim(planTestRequest(1, plan), nil)
	base := fw.PseudoStats.DamageDealtMultiplier
	sim.Cleanup()
	for range 3 {
		sim.runOnce()
		expectMultiplier(t, "player damage dealt after an iteration", fw.PseudoStats.DamageDealtMultiplier, base)
	}
}

func TestPlanTargetAmpReachesEveryLifeOfARepeatingWave(t *testing.T) {
	amp := planAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0, 1)
	amp.Targets = &proto.PlanTargets{WaveIndices: []int32{0}}
	plan := &proto.EncounterPlan{
		AddWaves:   []*proto.AddWave{testAddWave(1, 0.1, 0.05, &proto.PlanRepeat{Interval: 0.2})},
		DamageAmps: []*proto.PlanDamageAmp{amp},
	}
	sim, _ := setupPlanSim(planTestRequest(1, plan), nil)
	add := sim.Encounter.AllTargets[1]
	base := add.PseudoStats.DamageTakenMultiplier

	for _, spawn := range []float64{0.1, 0.3, 0.5, 0.7} {
		runLifecycleUntil(sim, planAt(spawn+0.01))
		if !add.IsEnabled() {
			t.Fatalf("at %v of the fight: expected the add up", spawn+0.01)
		}
		expectMultiplier(t, "add alive after its spawn at "+planAt(spawn).String(), add.PseudoStats.DamageTakenMultiplier, base*1.1)
	}
}

func TestPlanTargetAmpKeepsTheOpenWindowsAcrossADeath(t *testing.T) {
	first := planAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0.1, 0.55)
	second := planAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 5, 0.3, 0.9)
	for _, amp := range []*proto.PlanDamageAmp{first, second} {
		amp.Targets = &proto.PlanTargets{WaveIndices: []int32{0}}
	}
	plan := &proto.EncounterPlan{
		AddWaves:   []*proto.AddWave{testAddWave(1, 0.2, 0.3, &proto.PlanRepeat{Interval: 0.4, Count: 1})},
		DamageAmps: []*proto.PlanDamageAmp{first, second},
	}
	sim, _ := setupPlanSim(planTestRequest(1, plan), nil)
	add := sim.Encounter.AllTargets[1]
	base := add.PseudoStats.DamageTakenMultiplier

	runLifecycleUntil(sim, planAt(0.4))
	expectMultiplier(t, "add with both windows", add.PseudoStats.DamageTakenMultiplier, base*1.15)
	runLifecycleUntil(sim, planAt(0.57))
	if add.IsEnabled() {
		t.Fatal("after its life: expected the add gone")
	}
	runLifecycleUntil(sim, planAt(0.7))
	if !add.IsEnabled() {
		t.Fatal("at its second spawn: expected the add up")
	}
	expectMultiplier(t, "add respawned with only the second window open", add.PseudoStats.DamageTakenMultiplier, base*1.05)
}

func TestPlanTargetAmpReturnsAfterADowntimeThatClearsDebuffs(t *testing.T) {
	amp := planAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0.2, 0.9)
	plan := &proto.EncounterPlan{
		Downtimes:  []*proto.PlanDowntime{{TargetIndex: 0, Start: &proto.PlanTime{At: 0.3}, End: &proto.PlanTime{At: 0.5}, ClearDebuffs: true}},
		DamageAmps: []*proto.PlanDamageAmp{amp},
	}
	sim, _ := setupPlanSim(planTestRequest(1, plan), nil)
	boss := sim.Encounter.AllTargets[0]
	base := boss.PseudoStats.DamageTakenMultiplier

	runLifecycleUntil(sim, planAt(0.25))
	expectMultiplier(t, "boss before the downtime", boss.PseudoStats.DamageTakenMultiplier, base*1.1)
	runLifecycleUntil(sim, planAt(0.55))
	if !boss.IsEnabled() {
		t.Fatal("after the downtime: expected the boss back")
	}
	expectMultiplier(t, "boss after the downtime", boss.PseudoStats.DamageTakenMultiplier, base*1.1)
	runLifecycleUntil(sim, planAt(0.95))
	expectMultiplier(t, "boss after the window", boss.PseudoStats.DamageTakenMultiplier, base)
}

func TestPlanAmpEndsLeaveNoIdleAura(t *testing.T) {
	plan := &proto.EncounterPlan{DamageAmps: []*proto.PlanDamageAmp{
		planAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.2, 0.4),
		planAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.4, 0.6),
	}}
	sim, fw := setupPlanSim(planTestRequest(1, plan), nil)
	aura := fw.GetAura("Damage Done % (Caster Buff)")

	runLifecycleUntil(sim, planAt(0.4))
	if !aura.IsActive() || aura.GetStacks() != 1000 {
		t.Errorf("at the hand-over: expected the aura up with 1000 stacks, got active %t with %d", aura.IsActive(), aura.GetStacks())
	}
	if aura.metrics.Procs != 1 {
		t.Errorf("gains over two back-to-back windows: expected 1, got %d", aura.metrics.Procs)
	}
	runLifecycleUntil(sim, planAt(0.7))
	if aura.IsActive() {
		t.Error("after the last window: expected the aura off")
	}
}
