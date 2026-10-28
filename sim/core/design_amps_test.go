package core

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

func designAmpWindow(ampType proto.DamageAmpType, percent, start, end float64) *proto.DesignDamageAmp {
	return &proto.DesignDamageAmp{
		Type:    ampType,
		Percent: percent,
		Start:   &proto.DesignTime{At: start},
		End:     &proto.DesignTime{At: end},
	}
}

func expectMultiplier(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("%s: expected %v, got %v", what, want, got)
	}
}

func TestDesignCasterAmpRaisesDamageDealtInsideItsWindow(t *testing.T) {
	design := &proto.EncounterDesign{DamageAmps: []*proto.DesignDamageAmp{
		designAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.2, 0.4),
	}}
	sim, fw := setupDesignSim(designTestRequest(1, design), nil)
	base := fw.PseudoStats.DamageDealtMultiplier
	caster := proto.DamageAmpType_DamageAmpTypeCasterBuff
	environment := proto.DamageAmpType_DamageAmpTypeEnvironmentBuff

	if sim.Encounter.AmpActive(sim, caster) {
		t.Error("before the window: expected no active caster amp")
	}
	if got := sim.Encounter.NextAmpAt(sim, caster); got != designAt(0.2) {
		t.Errorf("next caster amp before the window: expected %s, got %s", designAt(0.2), got)
	}
	if got := sim.Encounter.NextAmpAt(sim, environment); got != NeverExpires {
		t.Errorf("next environment amp of a design without one: expected never, got %s", got)
	}

	runLifecycleUntil(sim, designAt(0.2)-time.Millisecond)
	expectMultiplier(t, "damage dealt before the window", fw.PseudoStats.DamageDealtMultiplier, base)

	runLifecycleUntil(sim, designAt(0.2))
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

	runLifecycleUntil(sim, designAt(0.4))
	expectMultiplier(t, "damage dealt after the window", fw.PseudoStats.DamageDealtMultiplier, base)
	if sim.Encounter.AmpActive(sim, caster) {
		t.Error("after the window: expected no active caster amp")
	}
	if got := sim.Encounter.NextAmpAt(sim, caster); got != NeverExpires {
		t.Errorf("next caster amp after the only window: expected never, got %s", got)
	}
}

func TestDesignAmpWindowsOfOneTypeAddUp(t *testing.T) {
	design := &proto.EncounterDesign{DamageAmps: []*proto.DesignDamageAmp{
		designAmpWindow(proto.DamageAmpType_DamageAmpTypeEnvironmentBuff, 10, 0.2, 0.6),
		designAmpWindow(proto.DamageAmpType_DamageAmpTypeEnvironmentBuff, 2.5, 0.4, 0.8),
	}}
	sim, fw := setupDesignSim(designTestRequest(1, design), nil)
	base := fw.PseudoStats.DamageDealtMultiplier

	runLifecycleUntil(sim, designAt(0.3))
	expectMultiplier(t, "first window alone", fw.PseudoStats.DamageDealtMultiplier, base*1.1)
	runLifecycleUntil(sim, designAt(0.5))
	expectMultiplier(t, "both windows", fw.PseudoStats.DamageDealtMultiplier, base*1.125)
	runLifecycleUntil(sim, designAt(0.7))
	expectMultiplier(t, "second window alone", fw.PseudoStats.DamageDealtMultiplier, base*1.025)
	runLifecycleUntil(sim, designAt(0.9))
	expectMultiplier(t, "after both windows", fw.PseudoStats.DamageDealtMultiplier, base)
}

func TestDesignAmpWindowThatEndsAsAnotherStartsKeepsTheSum(t *testing.T) {
	design := &proto.EncounterDesign{DamageAmps: []*proto.DesignDamageAmp{
		designAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.2, 0.4),
		designAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.4, 0.6),
	}}
	sim, fw := setupDesignSim(designTestRequest(1, design), nil)
	base := fw.PseudoStats.DamageDealtMultiplier
	caster := proto.DamageAmpType_DamageAmpTypeCasterBuff

	runLifecycleUntil(sim, designAt(0.4))
	expectMultiplier(t, "at the hand-over", fw.PseudoStats.DamageDealtMultiplier, base*1.1)
	if !sim.Encounter.AmpActive(sim, caster) {
		t.Error("at the hand-over: expected the caster amp active")
	}
	runLifecycleUntil(sim, designAt(0.6))
	expectMultiplier(t, "after the second window", fw.PseudoStats.DamageDealtMultiplier, base)
}

func TestDesignAmpRepeatsEveryInterval(t *testing.T) {
	amp := designAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.1, 0.2)
	amp.Repeat = &proto.DesignRepeat{Interval: 0.3, Count: 1}
	sim, fw := setupDesignSim(designTestRequest(1, &proto.EncounterDesign{DamageAmps: []*proto.DesignDamageAmp{amp}}), nil)
	base := fw.PseudoStats.DamageDealtMultiplier
	caster := proto.DamageAmpType_DamageAmpTypeCasterBuff

	runLifecycleUntil(sim, designAt(0.1))
	if got := sim.Encounter.NextAmpAt(sim, caster); got != designAt(0.1) {
		t.Errorf("inside the first window: expected now, got %s", got)
	}
	runLifecycleUntil(sim, designAt(0.25))
	if got := sim.Encounter.NextAmpAt(sim, caster); got != designAt(0.4) {
		t.Errorf("between windows: expected the repeat at %s, got %s", designAt(0.4), got)
	}
	expectMultiplier(t, "between windows", fw.PseudoStats.DamageDealtMultiplier, base)
	runLifecycleUntil(sim, designAt(0.45))
	expectMultiplier(t, "inside the repeat", fw.PseudoStats.DamageDealtMultiplier, base*1.1)
	runLifecycleUntil(sim, designAt(0.6))
	if got := sim.Encounter.NextAmpAt(sim, caster); got != NeverExpires {
		t.Errorf("after the last repeat: expected never, got %s", got)
	}
}

func TestDesignTargetAmpHitsOnlyItsTargets(t *testing.T) {
	amp := designAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0.2, 0.4)
	amp.Targets = &proto.DesignTargets{TargetIndices: []int32{1}}
	sim, _ := setupDesignSim(designTestRequest(3, &proto.EncounterDesign{DamageAmps: []*proto.DesignDamageAmp{amp}}), nil)
	targets := sim.Encounter.AllTargets
	base := targets[1].PseudoStats.DamageTakenMultiplier

	runLifecycleUntil(sim, designAt(0.3))
	for i, want := range []float64{base, base * 1.1, base} {
		expectMultiplier(t, "damage taken of target "+string(rune('0'+i)), targets[i].PseudoStats.DamageTakenMultiplier, want)
	}
	if aura := targets[1].GetAura("Damage Taken % (Target Debuff)"); aura == nil || !aura.IsActive() {
		t.Error("inside the window: expected the debuff aura on the named target")
	}
	if targets[0].GetAura("Damage Taken % (Target Debuff)") != nil {
		t.Error("expected no debuff aura on a target the amp does not name")
	}

	runLifecycleUntil(sim, designAt(0.5))
	expectMultiplier(t, "damage taken after the window", targets[1].PseudoStats.DamageTakenMultiplier, base)
}

func TestDesignTargetAmpWithoutTargetsHitsTheBoss(t *testing.T) {
	amp := designAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 20, 0.2, 0.4)
	sim, _ := setupDesignSim(designTestRequest(2, &proto.EncounterDesign{DamageAmps: []*proto.DesignDamageAmp{amp}}), nil)
	targets := sim.Encounter.AllTargets
	base := targets[0].PseudoStats.DamageTakenMultiplier

	runLifecycleUntil(sim, designAt(0.3))
	expectMultiplier(t, "damage taken of the boss", targets[0].PseudoStats.DamageTakenMultiplier, base*1.2)
	expectMultiplier(t, "damage taken of the second target", targets[1].PseudoStats.DamageTakenMultiplier, base)
}

func TestDesignTargetAmpReachesAddsThatSpawnInsideTheWindow(t *testing.T) {
	amp := designAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0.2, 0.6)
	amp.Targets = &proto.DesignTargets{WaveIndices: []int32{0}}
	design := &proto.EncounterDesign{
		AddWaves:   []*proto.AddWave{testAddWave(1, 0.3, 0.1, nil)},
		DamageAmps: []*proto.DesignDamageAmp{amp},
	}
	sim, _ := setupDesignSim(designTestRequest(1, design), nil)
	boss, add := sim.Encounter.AllTargets[0], sim.Encounter.AllTargets[1]
	base := boss.PseudoStats.DamageTakenMultiplier

	runLifecycleUntil(sim, designAt(0.25))
	expectMultiplier(t, "boss inside a window naming only the wave", boss.PseudoStats.DamageTakenMultiplier, base)

	runLifecycleUntil(sim, designAt(0.3))
	if !add.IsEnabled() {
		t.Fatal("at the spawn: expected the add active")
	}
	expectMultiplier(t, "add spawned inside the window", add.PseudoStats.DamageTakenMultiplier, base*1.1)
	expectMultiplier(t, "boss with the add up", boss.PseudoStats.DamageTakenMultiplier, base)

	runLifecycleUntil(sim, designAt(0.7))
	expectMultiplier(t, "add after the window", add.PseudoStats.DamageTakenMultiplier, base)
}

func TestDesignAmpsRunEveryIteration(t *testing.T) {
	amp := designAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0.2, 2)
	amp.Targets = &proto.DesignTargets{WaveIndices: []int32{0}}
	caster := designAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 5, 0.5, 1)
	design := &proto.EncounterDesign{
		AddWaves:   []*proto.AddWave{testAddWave(1, 0.3, 0.1, nil)},
		DamageAmps: []*proto.DesignDamageAmp{amp, caster},
	}
	sim, fw := setupDesignSim(designTestRequest(1, design), nil)
	base := fw.PseudoStats.DamageDealtMultiplier
	sim.Cleanup()
	for range 3 {
		sim.runOnce()
		expectMultiplier(t, "player damage dealt after an iteration", fw.PseudoStats.DamageDealtMultiplier, base)
	}
}

func TestDesignTargetAmpReachesEveryLifeOfARepeatingWave(t *testing.T) {
	amp := designAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0, 1)
	amp.Targets = &proto.DesignTargets{WaveIndices: []int32{0}}
	design := &proto.EncounterDesign{
		AddWaves:   []*proto.AddWave{testAddWave(1, 0.1, 0.05, &proto.DesignRepeat{Interval: 0.2})},
		DamageAmps: []*proto.DesignDamageAmp{amp},
	}
	sim, _ := setupDesignSim(designTestRequest(1, design), nil)
	add := sim.Encounter.AllTargets[1]
	base := add.PseudoStats.DamageTakenMultiplier

	for _, spawn := range []float64{0.1, 0.3, 0.5, 0.7} {
		runLifecycleUntil(sim, designAt(spawn+0.01))
		if !add.IsEnabled() {
			t.Fatalf("at %v of the fight: expected the add up", spawn+0.01)
		}
		expectMultiplier(t, "add alive after its spawn at "+designAt(spawn).String(), add.PseudoStats.DamageTakenMultiplier, base*1.1)
	}
}

func TestDesignTargetAmpKeepsTheOpenWindowsAcrossADeath(t *testing.T) {
	first := designAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0.1, 0.55)
	second := designAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 5, 0.3, 0.9)
	for _, amp := range []*proto.DesignDamageAmp{first, second} {
		amp.Targets = &proto.DesignTargets{WaveIndices: []int32{0}}
	}
	design := &proto.EncounterDesign{
		AddWaves:   []*proto.AddWave{testAddWave(1, 0.2, 0.3, &proto.DesignRepeat{Interval: 0.4, Count: 1})},
		DamageAmps: []*proto.DesignDamageAmp{first, second},
	}
	sim, _ := setupDesignSim(designTestRequest(1, design), nil)
	add := sim.Encounter.AllTargets[1]
	base := add.PseudoStats.DamageTakenMultiplier

	runLifecycleUntil(sim, designAt(0.4))
	expectMultiplier(t, "add with both windows", add.PseudoStats.DamageTakenMultiplier, base*1.15)
	runLifecycleUntil(sim, designAt(0.57))
	if add.IsEnabled() {
		t.Fatal("after its life: expected the add gone")
	}
	runLifecycleUntil(sim, designAt(0.7))
	if !add.IsEnabled() {
		t.Fatal("at its second spawn: expected the add up")
	}
	expectMultiplier(t, "add respawned with only the second window open", add.PseudoStats.DamageTakenMultiplier, base*1.05)
}

func TestDesignTargetAmpReturnsAfterADowntimeThatClearsDebuffs(t *testing.T) {
	amp := designAmpWindow(proto.DamageAmpType_DamageAmpTypeTargetDebuff, 10, 0.2, 0.9)
	design := &proto.EncounterDesign{
		Downtimes:  []*proto.DesignDowntime{{TargetIndex: 0, Start: &proto.DesignTime{At: 0.3}, End: &proto.DesignTime{At: 0.5}, ClearDebuffs: true}},
		DamageAmps: []*proto.DesignDamageAmp{amp},
	}
	sim, _ := setupDesignSim(designTestRequest(1, design), nil)
	boss := sim.Encounter.AllTargets[0]
	base := boss.PseudoStats.DamageTakenMultiplier

	runLifecycleUntil(sim, designAt(0.25))
	expectMultiplier(t, "boss before the downtime", boss.PseudoStats.DamageTakenMultiplier, base*1.1)
	runLifecycleUntil(sim, designAt(0.55))
	if !boss.IsEnabled() {
		t.Fatal("after the downtime: expected the boss back")
	}
	expectMultiplier(t, "boss after the downtime", boss.PseudoStats.DamageTakenMultiplier, base*1.1)
	runLifecycleUntil(sim, designAt(0.95))
	expectMultiplier(t, "boss after the window", boss.PseudoStats.DamageTakenMultiplier, base)
}

func TestDesignAmpEndsLeaveNoIdleAura(t *testing.T) {
	design := &proto.EncounterDesign{DamageAmps: []*proto.DesignDamageAmp{
		designAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.2, 0.4),
		designAmpWindow(proto.DamageAmpType_DamageAmpTypeCasterBuff, 10, 0.4, 0.6),
	}}
	sim, fw := setupDesignSim(designTestRequest(1, design), nil)
	aura := fw.GetAura("Damage Done % (Caster Buff)")

	runLifecycleUntil(sim, designAt(0.4))
	if !aura.IsActive() || aura.GetStacks() != 1000 {
		t.Errorf("at the hand-over: expected the aura up with 1000 stacks, got active %t with %d", aura.IsActive(), aura.GetStacks())
	}
	if aura.metrics.Procs != 1 {
		t.Errorf("gains over two back-to-back windows: expected 1, got %d", aura.metrics.Procs)
	}
	runLifecycleUntil(sim, designAt(0.7))
	if aura.IsActive() {
		t.Error("after the last window: expected the aura off")
	}
}
