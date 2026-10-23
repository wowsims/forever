package core_test

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

const healthRegenTick = time.Millisecond * 2020

// The hunter owner with its pets and hp5 HP5 on its sheet, nothing swinging until the pull; prepull
// holds the rotation's prepull actions.
func setupHealthRegenSim(t *testing.T, hp5 float64, prepull ...*proto.APLPrepullAction) (*core.Simulation, *petOwner) {
	t.Helper()
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{
			Name:       "Owner",
			Race:       proto.Race_RaceHuman,
			Class:      proto.Class_ClassHunter,
			Spec:       &proto.Player_Hunter{},
			Equipment:  &proto.EquipmentSpec{},
			BonusStats: &proto.UnitStats{Stats: stats.Stats{stats.HP5: hp5}.ToProtoArray()},
			Rotation:   &proto.APLRotation{PrepullActions: prepull},
		}}}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	for _, unit := range sim.Raid.AllUnits {
		unit.AutoAttacks.CancelAutoSwing(sim)
	}
	return sim, sim.Raid.Parties[0].Players[0].(*petOwner)
}

// A prepull action at doAt, so the sim starts there.
func prepullMoveAt(doAt string) *proto.APLPrepullAction {
	return &proto.APLPrepullAction{
		DoAtValue: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: doAt}}},
		Action: &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{
			RangeFromTarget: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "5"}}},
		}}},
	}
}

// The health the unit gains from now up to each time, read once everything else due then has run.
func recordHealthGained(sim *core.Simulation, unit *core.Unit, times ...time.Duration) []float64 {
	start := unit.CurrentHealth()
	gained := make([]float64, len(times))
	for i, at := range times {
		sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{
			DoAt:     at,
			Priority: core.ActionPriorityLow,
			OnAction: func(*core.Simulation) { gained[i] = unit.CurrentHealth() - start },
		}))
	}
	return gained
}

// The unit's metric for the other action, nil if it recorded none.
func otherActionMetrics(unit *core.Unit, id proto.OtherAction) *proto.ResourceMetrics {
	for _, metrics := range unit.Metrics.ToProto().Resources {
		if metrics.Id.GetOtherId() == id {
			return metrics
		}
	}
	return nil
}

// A temporary stats aura on the owner; auras can only be built while the sim is being constructed.
func newTestStatsAura(sim *core.Simulation, owner *petOwner, name string, spellID int32, bonus stats.Stats) *core.StatBuffAura {
	defer func(state core.EnvironmentState) { sim.Environment.State = state }(sim.Environment.State)
	sim.Environment.State = core.Constructed
	return owner.NewTemporaryStatsAura(name, core.ActionID{SpellID: spellID}, bonus, core.NeverExpires)
}

// 10 HP5 credits 4 health on each 2.02 s tick, the first at 2.02 s.
func TestHealthRegenTicksEvery2020Milliseconds(t *testing.T) {
	sim, owner := setupHealthRegenSim(t, 10)
	owner.SetCurrentHealthPercent(0.5)

	gained := recordHealthGained(sim, &owner.Unit, healthRegenTick-time.Millisecond, healthRegenTick, 2*healthRegenTick-time.Millisecond, 2*healthRegenTick)
	stepTo(t, sim, 5*time.Second)
	if want := []float64{0, 4, 4, 8}; !slices.Equal(gained, want) {
		t.Errorf("health gained by 2.019 s, 2.02 s, 4.039 s and 4.04 s: %v, want %v", gained, want)
	}

	metrics := otherActionMetrics(&owner.Unit, proto.OtherAction_OtherActionHealthRegen)
	if metrics == nil {
		t.Fatal("no Health Regen health metric")
	}
	if metrics.Type != proto.ResourceType_ResourceTypeHealth || metrics.Events != 2 || metrics.ActualGain != 8 {
		t.Errorf("Health Regen metric: %v events of %v %v, want 2 events of 8 health", metrics.Events, metrics.ActualGain, metrics.Type)
	}
}

// HP5 gained in the prepull, even more than a tick before the pull, waits for the 2.02 s tick, with none at 0.
func TestHealthRegenNoneBeforeCombat(t *testing.T) {
	for _, start := range []time.Duration{-1500 * time.Millisecond, -3 * time.Second} {
		sim, owner := setupHealthRegenSim(t, 10, prepullMoveAt(start.String()))
		sim.PrePull()
		if sim.CurrentTime != start {
			t.Fatalf("the prepull starts at %v, want %v", sim.CurrentTime, start)
		}
		owner.RemoveHealth(sim, 100)

		gained := recordHealthGained(sim, &owner.Unit, 0, time.Second, 2*time.Second, healthRegenTick-time.Millisecond, healthRegenTick)
		stepTo(t, sim, 3*time.Second)
		if want := []float64{0, 0, 0, 0, 4}; !slices.Equal(gained, want) {
			t.Errorf("prepull from %v, health gained by 0 s, 1 s, 2 s, 2.019 s and 2.02 s: %v, want %v", start, gained, want)
		}
	}
}

// A tick credits only up to the maximum; the metric records the share that landed.
func TestHealthRegenCapsAtMaxHealth(t *testing.T) {
	sim, owner := setupHealthRegenSim(t, 10)
	owner.RemoveHealth(sim, 1)

	stepTo(t, sim, 3*time.Second)
	if got, want := owner.CurrentHealth(), owner.MaxHealth(); got != want {
		t.Errorf("health %v after the tick, want the maximum %v", got, want)
	}
	if metrics := otherActionMetrics(&owner.Unit, proto.OtherAction_OtherActionHealthRegen); metrics == nil || metrics.ActualGain != 1 {
		t.Errorf("Health Regen metric %v, want 1 health gained", metrics)
	}
}

// A dead unit gains nothing from its HP5.
func TestHealthRegenStopsAfterDeath(t *testing.T) {
	sim, owner := setupHealthRegenSim(t, 10)
	owner.RemoveHealth(sim, owner.CurrentHealth())
	owner.Died(sim)

	gained := recordHealthGained(sim, &owner.Unit, healthRegenTick, 2*healthRegenTick, 10*time.Second)
	stepTo(t, sim, 11*time.Second)
	if want := []float64{0, 0, 0}; !slices.Equal(gained, want) || owner.CurrentHealth() != 0 {
		t.Errorf("dead, the unit gained %v health by 2.02 s, 4.04 s and 10 s, want %v", gained, want)
	}
}

// HP5 gained at 3 s waits for the 4.04 s tick; the time before it earns nothing.
func TestHealthRegenStartsWhenHP5Gained(t *testing.T) {
	sim, owner := setupHealthRegenSim(t, 0)
	aura := newTestStatsAura(sim, owner, "Test HP5", 910700, stats.Stats{stats.HP5: 10})
	owner.SetCurrentHealthPercent(0.5)

	gained := recordHealthGained(sim, &owner.Unit, 3*time.Second, 2*healthRegenTick-time.Millisecond, 2*healthRegenTick, 3*healthRegenTick)
	stepTo(t, sim, 3*time.Second)
	aura.Activate(sim)
	stepTo(t, sim, 7*time.Second)
	if want := []float64{0, 0, 4, 8}; !slices.Equal(gained, want) {
		t.Errorf("health gained by 3 s, 4.039 s, 4.04 s and 6.06 s: %v, want %v", gained, want)
	}
}

// HP5 lost at 3 s credits nothing at 4.04 s; HP5 regained at 5 s credits from the 6.06 s tick.
func TestHealthRegenSleepsWhenHP5Lost(t *testing.T) {
	sim, owner := setupHealthRegenSim(t, 0)
	aura := newTestStatsAura(sim, owner, "Test HP5", 910701, stats.Stats{stats.HP5: 10})
	aura.Activate(sim)
	owner.SetCurrentHealthPercent(0.5)

	gained := recordHealthGained(sim, &owner.Unit, healthRegenTick, 3*time.Second, 2*healthRegenTick, 5*time.Second, 3*healthRegenTick-time.Millisecond, 3*healthRegenTick)
	stepTo(t, sim, 3*time.Second)
	aura.Deactivate(sim)
	stepTo(t, sim, 5*time.Second)
	aura.Activate(sim)
	stepTo(t, sim, 7*time.Second)
	if want := []float64{4, 4, 4, 4, 4, 8}; !slices.Equal(gained, want) {
		t.Errorf("health gained by 2.02 s, 3 s, 4.04 s, 5 s, 6.059 s and 6.06 s: %v, want %v", gained, want)
	}
}

// A pet regenerates like a player while summoned and not while away.
func TestHealthRegenDisabledPet(t *testing.T) {
	sim, owner := setupHealthRegenSim(t, 0)
	if !owner.pet.IsEnabled() || owner.dynamic.IsEnabled() {
		t.Fatalf("the true pet is enabled %v and the dynamic pet %v, want only the true pet", owner.pet.IsEnabled(), owner.dynamic.IsEnabled())
	}
	for _, pet := range []*testPet{owner.pet, owner.dynamic} {
		pet.AddStatsDynamic(sim, stats.Stats{stats.Health: 1000, stats.HP5: 10})
		pet.SetCurrentHealthPercent(0.5)
	}

	summoned := recordHealthGained(sim, &owner.pet.Unit, healthRegenTick, 2*healthRegenTick)
	away := recordHealthGained(sim, &owner.dynamic.Unit, healthRegenTick, 2*healthRegenTick)
	stepTo(t, sim, 5*time.Second)
	if want := []float64{4, 8}; !slices.Equal(summoned, want) {
		t.Errorf("the summoned pet gained %v health by 2.02 s and 4.04 s, want %v", summoned, want)
	}
	if want := []float64{0, 0}; !slices.Equal(away, want) {
		t.Errorf("the pet away gained %v health by 2.02 s and 4.04 s, want %v", away, want)
	}
}

// At full health a tick credits nothing and records no event; the first tick after damage credits.
func TestHealthRegenSkipsTicksAtFullHealth(t *testing.T) {
	sim, owner := setupHealthRegenSim(t, 10)

	stepTo(t, sim, 7*time.Second)
	if metrics := otherActionMetrics(&owner.Unit, proto.OtherAction_OtherActionHealthRegen); metrics != nil {
		t.Errorf("Health Regen metric at full health through 6.06 s: %v events, want none", metrics.Events)
	}

	owner.RemoveHealth(sim, 100)
	gained := recordHealthGained(sim, &owner.Unit, 4*healthRegenTick-time.Millisecond, 4*healthRegenTick)
	stepTo(t, sim, 9*time.Second)
	if want := []float64{0, 4}; !slices.Equal(gained, want) {
		t.Errorf("health gained by 8.079 s and 8.08 s: %v, want %v", gained, want)
	}
	if metrics := otherActionMetrics(&owner.Unit, proto.OtherAction_OtherActionHealthRegen); metrics == nil || metrics.Events != 1 || metrics.ActualGain != 4 {
		t.Errorf("Health Regen metric %v, want 1 event of 4 health", metrics)
	}
}

// Losing a Stamina buff caps current health at the lower maximum at once, so the next tick gains nothing.
func TestHealthRegenAfterMaxHealthDrops(t *testing.T) {
	sim, owner := setupHealthRegenSim(t, 10)
	aura := newTestStatsAura(sim, owner, "Test Stamina", 910702, stats.Stats{stats.Stamina: 100})
	baseMax := owner.MaxHealth()
	aura.Activate(sim)
	if owner.MaxHealth() <= baseMax {
		t.Fatalf("maximum health %v with the Stamina buff, want above %v", owner.MaxHealth(), baseMax)
	}
	owner.SetCurrentHealthPercent(1)

	aura.Deactivate(sim)
	if got, want := owner.CurrentHealth(), owner.MaxHealth(); got != want {
		t.Errorf("health %v once the buff fades, want the maximum %v", got, want)
	}
	stepTo(t, sim, 3*time.Second)
	if metrics := otherActionMetrics(&owner.Unit, proto.OtherAction_OtherActionHealthRegen); metrics != nil {
		t.Errorf("Health Regen metric %v events of %v health after the buff fades, want none", metrics.Events, metrics.ActualGain)
	}
}

// A raise of 3000 maximum health that fades takes 3000 from current health once, then lowers the
// maximum: at full health the unit ends at its old maximum; hit for 2000 in between, it ends 2000 below it.
func TestHealthRegenMaxHealthRaiseFades(t *testing.T) {
	for _, tc := range []struct {
		hit    float64
		events int32
	}{{0, 1}, {2000, 2}} {
		sim, owner := setupHealthRegenSim(t, 0)
		metrics := owner.NewHealthMetrics(core.ActionID{SpellID: 910703})
		baseMax := owner.MaxHealth()

		owner.UpdateMaxHealth(sim, 3000, metrics)
		if owner.MaxHealth() != baseMax+3000 || owner.CurrentHealth() != baseMax+3000 {
			t.Fatalf("after the raise, health %v of %v, want %v of %v", owner.CurrentHealth(), owner.MaxHealth(), baseMax+3000, baseMax+3000)
		}
		if tc.hit > 0 {
			owner.RemoveHealth(sim, tc.hit)
		}

		owner.UpdateMaxHealth(sim, -3000, metrics)
		if got, want := owner.CurrentHealth(), baseMax-tc.hit; got != want || owner.MaxHealth() != baseMax {
			t.Errorf("hit for %v, health %v of %v once the raise fades, want %v of %v", tc.hit, got, owner.MaxHealth(), want, baseMax)
		}
		taken := otherActionMetrics(&owner.Unit, proto.OtherAction_OtherActionDamageTaken)
		if taken == nil || taken.Events != tc.events || taken.Gain != -(3000+tc.hit) {
			t.Errorf("hit for %v, damage taken metric %v, want %v events totalling %v", tc.hit, taken, tc.events, 3000+tc.hit)
		}
	}
}
