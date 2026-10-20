package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

var fakeHealActionID = ActionID{SpellID: 2050}

func (fw *FakeRageWarrior) registerFakeHeal() {
	fw.RegisterSpell(SpellConfig{
		ActionID:         fakeHealActionID,
		SpellSchool:      SpellSchoolHoly,
		ProcMask:         ProcMaskSpellHealing,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealHealing(sim, target, 300, spell.OutcomeAlwaysHit)
		},
	})
}

// Only the health a heal restores makes threat, half a point per point healed, split evenly between
// the enemies in combat: a 300 heal on a unit missing 100 makes 50 threat, 25 on each of two targets.
func TestHealThreatCountsTheHealthRestoredSplitBetweenTargets(t *testing.T) {
	request := fakeRageSimRequest()
	request.Encounter.Targets = append(request.Encounter.Targets,
		&proto.Target{Name: "second target", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2})
	sim := setupFakeRageSimFrom(request)
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	heal := fw.GetSpell(fakeHealActionID)
	// The fake warrior has no stamina and starts below zero health; it gets a 1000 pool at full.
	fw.AddStatDynamic(sim, stats.Health, 1000-fw.MaxHealth())
	fw.GainHealth(sim, fw.MaxHealth()-fw.CurrentHealth(), fw.DamageTakenHealthMetrics)

	threatOn := func() []float64 {
		var threat []float64
		for _, target := range sim.Encounter.ActiveTargetUnits {
			threat = append(threat, heal.SpellMetrics[target.UnitIndex].TotalThreat)
		}
		return threat
	}

	heal.Cast(sim, &fw.Unit)
	for i, threat := range threatOn() {
		if threat != 0 {
			t.Errorf("a heal at full health made %v threat on target %d, want 0", threat, i)
		}
	}

	fw.RemoveHealth(sim, 100)
	heal.Cast(sim, &fw.Unit)
	for i, threat := range threatOn() {
		if !WithinToleranceFloat64(25, threat, 0.001) {
			t.Errorf("a heal restoring 100 health made %v threat on target %d, want 25", threat, i)
		}
	}
}
