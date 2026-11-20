package core_test

import (
	"math"
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

// The health 10 HP5 credits a tick: HP5 / 5 a second over the 2.02 s tick.
const tenHP5Tick = 10 * 2.02 / 5

// The health n ticks of 10 HP5 credit, one entry per n.
func tenHP5Ticks(n ...float64) []float64 {
	health := make([]float64, len(n))
	for i, ticks := range n {
		health[i] = ticks * tenHP5Tick
	}
	return health
}

// Whether each gain matches its wanted health, within float error.
func healthGainsMatch(gained, want []float64) bool {
	return slices.EqualFunc(gained, want, func(a, b float64) bool { return math.Abs(a-b) < 1e-9 })
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

// The sim setupHealthRegenSim builds, with a temporary stats aura the owner registers in Initialize.
func setupHealthRegenSimWithAura(t *testing.T, hp5 float64, name string, spellID int32, bonus stats.Stats) (*core.Simulation, *petOwner, *core.StatBuffAura) {
	t.Helper()
	var aura *core.StatBuffAura
	ownerSetup = func(owner *petOwner) {
		aura = owner.NewTemporaryStatsAura(name, core.ActionID{SpellID: spellID}, bonus, core.NeverExpires)
	}
	defer func() { ownerSetup = nil }()
	sim, owner := setupHealthRegenSim(t, hp5)
	return sim, owner, aura
}

// 10 HP5 credits 4.04 health on each 2.02 s tick, the first at 2.02 s.
func TestHealthRegenTicksEvery2020Milliseconds(t *testing.T) {
	sim, owner := setupHealthRegenSim(t, 10)
	owner.RemoveHealth(sim, owner.CurrentHealth()/2)

	gained := recordHealthGained(sim, &owner.Unit, healthRegenTick-time.Millisecond, healthRegenTick, 2*healthRegenTick-time.Millisecond, 2*healthRegenTick)
	stepTo(t, sim, 5*time.Second)
	if want := tenHP5Ticks(0, 1, 1, 2); !healthGainsMatch(gained, want) {
		t.Errorf("health gained by 2.019 s, 2.02 s, 4.039 s and 4.04 s: %v, want %v", gained, want)
	}

	metrics := otherActionMetrics(&owner.Unit, proto.OtherAction_OtherActionHealthRegen)
	if metrics == nil {
		t.Fatal("no Health Regen health metric")
	}
	if metrics.Type != proto.ResourceType_ResourceTypeHealth || metrics.Events != 2 || !healthGainsMatch([]float64{metrics.ActualGain}, tenHP5Ticks(2)) {
		t.Errorf("Health Regen metric: %v events of %v %v, want 2 events of %v health", metrics.Events, metrics.ActualGain, metrics.Type, 2*tenHP5Tick)
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
		if want := tenHP5Ticks(0, 0, 0, 0, 1); !healthGainsMatch(gained, want) {
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
	sim, owner, aura := setupHealthRegenSimWithAura(t, 0, "Test HP5", 910700, stats.Stats{stats.HP5: 10})
	owner.SetCurrentHealthPercent(0.5)

	gained := recordHealthGained(sim, &owner.Unit, 3*time.Second, 2*healthRegenTick-time.Millisecond, 2*healthRegenTick, 3*healthRegenTick)
	stepTo(t, sim, 3*time.Second)
	aura.Activate(sim)
	stepTo(t, sim, 7*time.Second)
	if want := tenHP5Ticks(0, 0, 1, 2); !healthGainsMatch(gained, want) {
		t.Errorf("health gained by 3 s, 4.039 s, 4.04 s and 6.06 s: %v, want %v", gained, want)
	}
}

// HP5 lost at 3 s credits nothing at 4.04 s; HP5 regained at 5 s credits from the 6.06 s tick.
func TestHealthRegenSleepsWhenHP5Lost(t *testing.T) {
	sim, owner, aura := setupHealthRegenSimWithAura(t, 0, "Test HP5", 910701, stats.Stats{stats.HP5: 10})
	aura.Activate(sim)
	owner.RemoveHealth(sim, owner.CurrentHealth()/2)

	gained := recordHealthGained(sim, &owner.Unit, healthRegenTick, 3*time.Second, 2*healthRegenTick, 5*time.Second, 3*healthRegenTick-time.Millisecond, 3*healthRegenTick)
	stepTo(t, sim, 3*time.Second)
	aura.Deactivate(sim)
	stepTo(t, sim, 5*time.Second)
	aura.Activate(sim)
	stepTo(t, sim, 7*time.Second)
	if want := tenHP5Ticks(1, 1, 1, 1, 1, 2); !healthGainsMatch(gained, want) {
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
	if want := tenHP5Ticks(1, 2); !healthGainsMatch(summoned, want) {
		t.Errorf("the summoned pet gained %v health by 2.02 s and 4.04 s, want %v", summoned, want)
	}
	if want := []float64{0, 0}; !slices.Equal(away, want) {
		t.Errorf("the pet away gained %v health by 2.02 s and 4.04 s, want %v", away, want)
	}
}

// A pet with HP5 summoned at 5 s regenerates from the 6.06 s tick.
func TestHealthRegenPetSummonedMidFight(t *testing.T) {
	sim, owner := setupHealthRegenSim(t, 0)
	pet := owner.dynamic
	pet.AddStatsDynamic(sim, stats.Stats{stats.Health: 1000, stats.HP5: 10})

	stepTo(t, sim, 5*time.Second)
	pet.Enable(sim, pet)
	pet.RemoveHealth(sim, pet.CurrentHealth()/2)

	gained := recordHealthGained(sim, &pet.Unit, 3*healthRegenTick-time.Millisecond, 3*healthRegenTick, 4*healthRegenTick)
	stepTo(t, sim, 9*time.Second)
	if want := tenHP5Ticks(0, 1, 2); !healthGainsMatch(gained, want) {
		t.Errorf("the pet summoned at 5 s gained %v health by 6.059 s, 6.06 s and 8.08 s, want %v", gained, want)
	}
}

// At full health the ticks sleep, so none is an advance point: with the pets, whose rotations step
// every 10 ms, sent away, an aura due to expire at 3 s runs its expiry at the next thing the sim does,
// here at 10 s.
func TestHealthRegenSleepsAtFullHealth(t *testing.T) {
	var aura *core.Aura
	expiredAt := time.Duration(-1)
	ownerSetup = func(owner *petOwner) {
		aura = owner.RegisterAura(core.Aura{
			Label:    "Test Expiry",
			ActionID: core.ActionID{SpellID: 910704},
			Duration: 3 * time.Second,
			OnExpire: func(_ *core.Aura, sim *core.Simulation) { expiredAt = sim.CurrentTime },
		})
	}
	defer func() { ownerSetup = nil }()
	sim, owner := setupHealthRegenSim(t, 10)
	owner.pet.Disable(sim)
	owner.guardian.Disable(sim)
	aura.Activate(sim)

	stepTo(t, sim, 10*time.Second)
	if expiredAt != 10*time.Second {
		t.Errorf("the aura due at 3 s expired at %v, want 10 s with no tick in between", expiredAt)
	}
	if metrics := otherActionMetrics(&owner.Unit, proto.OtherAction_OtherActionHealthRegen); metrics != nil {
		t.Errorf("Health Regen metric at full health through 10 s: %v events, want none", metrics.Events)
	}
}

// Health lost at 3 s, to damage, a self cost or a raised maximum, wakes the ticks at 4.04 s.
func TestHealthRegenWakesWhenHurt(t *testing.T) {
	for _, tc := range []struct {
		name string
		hurt func(*core.Simulation, *petOwner)
	}{
		{"damage taken", func(sim *core.Simulation, owner *petOwner) { owner.RemoveHealth(sim, 100) }},
		{"a self cost", func(sim *core.Simulation, owner *petOwner) {
			owner.SpendHealth(sim, 100, owner.NewHealthMetrics(core.ActionID{SpellID: 910705}))
		}},
		{"a raised maximum", func(sim *core.Simulation, owner *petOwner) {
			owner.AddStatsDynamic(sim, stats.Stats{stats.Health: 100})
		}},
	} {
		sim, owner := setupHealthRegenSim(t, 10)
		stepTo(t, sim, 3*time.Second)
		tc.hurt(sim, owner)
		gained := recordHealthGained(sim, &owner.Unit, 2*healthRegenTick-time.Millisecond, 2*healthRegenTick)
		stepTo(t, sim, 5*time.Second)
		if want := tenHP5Ticks(0, 1); !healthGainsMatch(gained, want) {
			t.Errorf("hurt by %s at 3 s, health gained by 4.039 s and 4.04 s: %v, want %v", tc.name, gained, want)
		}
	}
}

// Losing a Stamina buff caps current health at the lower maximum at once, so the next tick gains nothing.
func TestHealthRegenAfterMaxHealthDrops(t *testing.T) {
	sim, owner, aura := setupHealthRegenSimWithAura(t, 10, "Test Stamina", 910702, stats.Stats{stats.Stamina: 100})
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

// A raise of 3000 maximum health that fades caps current health at the old maximum and books no damage
// taken: hit for 0 or 2000 in between, the unit ends at its old maximum; hit for 3500, 500 below it.
func TestHealthRegenMaxHealthRaiseFades(t *testing.T) {
	for _, tc := range []struct {
		hit   float64
		below float64
	}{{0, 0}, {2000, 0}, {3500, 500}} {
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
		if got, want := owner.CurrentHealth(), baseMax-tc.below; got != want || owner.MaxHealth() != baseMax {
			t.Errorf("hit for %v, health %v of %v once the raise fades, want %v of %v", tc.hit, got, owner.MaxHealth(), want, baseMax)
		}
		taken := otherActionMetrics(&owner.Unit, proto.OtherAction_OtherActionDamageTaken)
		if tc.hit == 0 && taken != nil {
			t.Errorf("never hit, damage taken metric %v once the raise fades, want none", taken)
		}
		if tc.hit > 0 && (taken == nil || taken.Events != 1 || taken.Gain != -tc.hit) {
			t.Errorf("hit for %v, damage taken metric %v once the raise fades, want the hit alone", tc.hit, taken)
		}
	}
}
