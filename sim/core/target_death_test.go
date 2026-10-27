package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// Two targets, so the first can die while the second stays active, and labelled rands, so a test can
// tell whether the kill roll drew.
func setupTargetDeathSim(register func(char *Character)) (*Simulation, *Target) {
	request := fakeRageSimRequest()
	request.SimOptions.IsTest = true
	request.Encounter.Targets = append(request.Encounter.Targets,
		&proto.Target{Name: "second target", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2})
	sim, _, _ := setupFakeRageRotationFrom(request, register)
	return sim, sim.Encounter.AllTargets[0]
}

func registerDeathListener(char *Character, label string, onTargetDies, onKill OnTargetDeath) *Aura {
	return char.RegisterAura(Aura{
		Label:        label,
		Duration:     NeverExpires,
		OnTargetDies: onTargetDies,
		OnKill:       onKill,
	})
}

func TestTargetDieRunsTheCallbacksBeforeTheDebuffsFade(t *testing.T) {
	var debuff *Aura
	var dyingTarget *Unit
	debuffActiveAtDeath := false
	var listener *Aura
	sim, target := setupTargetDeathSim(func(char *Character) {
		debuff = char.Env.Encounter.AllTargets[0].RegisterAura(Aura{Label: "Test Debuff", Duration: time.Second * 30})
		listener = registerDeathListener(char, "Death Listener", func(_ *Aura, _ *Simulation, target *Unit) {
			dyingTarget = target
			debuffActiveAtDeath = debuff.IsActive()
		}, nil)
	})
	listener.Activate(sim)
	debuff.Activate(sim)

	target.Die(sim, 1)

	if dyingTarget != &target.Unit {
		t.Fatalf("OnTargetDies got %v, want the dying target", dyingTarget)
	}
	if !debuffActiveAtDeath {
		t.Errorf("the debuff on the dying target is gone when OnTargetDies runs, want it still active")
	}
	if debuff.IsActive() {
		t.Errorf("the debuff on the dead target is still active after Die")
	}
	if target.IsEnabled() {
		t.Errorf("the target is still enabled after Die")
	}
}

func TestTargetDieCreditsTheKillAtTheKillShare(t *testing.T) {
	const deaths = 1000
	for _, tc := range []struct {
		killShare float64
		min, max  int
		draws     bool
	}{
		{killShare: 0, min: 0, max: 0, draws: false},
		{killShare: 1, min: deaths, max: deaths, draws: false},
		{killShare: 0.5, min: deaths * 45 / 100, max: deaths * 55 / 100, draws: true},
	} {
		diedCount, killCount := 0, 0
		var listener *Aura
		sim, target := setupTargetDeathSim(func(char *Character) {
			listener = registerDeathListener(char, "Death Listener",
				func(_ *Aura, _ *Simulation, _ *Unit) { diedCount++ },
				func(_ *Aura, _ *Simulation, _ *Unit) { killCount++ })
		})
		listener.Activate(sim)

		for range deaths {
			target.Die(sim, tc.killShare)
			target.Enable(sim)
		}

		if diedCount != deaths {
			t.Errorf("kill share %v: OnTargetDies ran %d times over %d deaths, want every death", tc.killShare, diedCount, deaths)
		}
		if killCount < tc.min || killCount > tc.max {
			t.Errorf("kill share %v: OnKill ran %d times over %d deaths, want %d to %d", tc.killShare, killCount, deaths, tc.min, tc.max)
		}
		if _, drew := sim.testRands["Killing Blow"]; drew != tc.draws {
			t.Errorf("kill share %v: the kill roll drew: %t, want %t", tc.killShare, drew, tc.draws)
		}
	}
}

func TestProcTriggerOnKillCallsTheHandlerWithTheDeadTarget(t *testing.T) {
	var killedTarget, diedTarget *Unit
	var killSpell *Spell
	kills := 0
	var onKill, onTargetDies *Aura
	sim, target := setupTargetDeathSim(func(char *Character) {
		onKill = char.MakeProcTriggerAura(ProcTrigger{
			Name:               "Kill Trigger",
			Callback:           CallbackOnKill,
			TriggerImmediately: true,
			Handler: func(_ *Simulation, spell *Spell, result *SpellResult) {
				kills++
				killSpell = spell
				killedTarget = result.Target
			},
		})
		onTargetDies = char.MakeProcTriggerAura(ProcTrigger{
			Name:               "Death Trigger",
			Callback:           CallbackOnTargetDies,
			TriggerImmediately: true,
			Handler: func(_ *Simulation, _ *Spell, result *SpellResult) {
				diedTarget = result.Target
			},
		})
	})
	onKill.Activate(sim)
	onTargetDies.Activate(sim)

	target.Die(sim, 0)
	if kills != 0 {
		t.Errorf("the OnKill trigger ran on a death with no kill credit")
	}
	if diedTarget != &target.Unit {
		t.Errorf("the OnTargetDies trigger got %v, want the dead target", diedTarget)
	}

	target.Enable(sim)
	target.Die(sim, 1)
	if kills != 1 {
		t.Fatalf("the OnKill trigger ran %d times on a credited kill, want once", kills)
	}
	if killedTarget != &target.Unit {
		t.Errorf("the OnKill trigger got %v, want the dead target", killedTarget)
	}
	if killSpell != nil {
		t.Errorf("the OnKill trigger got spell %v, want none", killSpell.ActionID)
	}
}

func TestTargetDieOnADisabledTargetDoesNothing(t *testing.T) {
	calls := 0
	var listener *Aura
	sim, target := setupTargetDeathSim(func(char *Character) {
		listener = registerDeathListener(char, "Death Listener",
			func(_ *Aura, _ *Simulation, _ *Unit) { calls++ },
			func(_ *Aura, _ *Simulation, _ *Unit) { calls++ })
	})
	listener.Activate(sim)
	target.Disable(sim, true)

	target.Die(sim, 1)

	if calls != 0 {
		t.Errorf("Die on a disabled target ran %d callbacks, want none", calls)
	}
}
