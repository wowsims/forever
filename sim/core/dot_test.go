package core

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	RegisterAgentFactory(
		proto.Player_ElementalShaman{},
		proto.Spec_SpecElementalShaman,
		NewFakeElementalShaman,
		func(player *proto.Player, spec any) {
			playerSpec, ok := spec.(*proto.Player_ElementalShaman)
			if !ok {
				panic("Invalid spec value for Elemental Shaman!")
			}
			player.Spec = playerSpec
		},
	)
}

type FakeAgent struct {
	Spell *Spell
	Dot   *Dot
	Character
	Init func()
}

func (fa *FakeAgent) GetCharacter() *Character {
	return &fa.Character
}

func (fa *FakeAgent) Initialize() {
	if fa.Init != nil {
		fa.Init()
	}
}

func (fa *FakeAgent) ApplyTalents()                  {}
func (fa *FakeAgent) Reset(_ *Simulation)            {}
func (fa *FakeAgent) OnGCDReady(_ *Simulation)       {}
func (fa *FakeAgent) OnEncounterStart(_ *Simulation) {}

func NewFakeElementalShaman(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	fa := &FakeAgent{
		Character: *char,
	}

	fa.Init = func() {
		fa.Spell = fa.RegisterSpell(SpellConfig{
			ActionID:    ActionID{SpellID: 42},
			SpellSchool: SpellSchoolShadow,
			ProcMask:    ProcMaskSpellDamage,
			Flags:       SpellFlagIgnoreResists,
			Cast:        CastConfig{},

			BonusCritPercent: 3,
			DamageMultiplier: 1.5,
			ThreatMultiplier: 1,

			Dot: DotConfig{
				Aura: Aura{
					Label: "fakedot",
				},
				NumberOfTicks:       6,
				TickLength:          time.Second * 3,
				AffectedByCastSpeed: false,
				BonusCoefficient:    1,

				OnSnapshot: func(sim *Simulation, target *Unit, dot *Dot) {
					dot.Snapshot(target, 100)
				},
				OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				},
			},

			ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
				result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
				if result.Landed() {
					spell.Dot(target).Apply(sim)
				}
				spell.DealOutcome(sim, result)
			},
		})
		fa.Dot = fa.Spell.CurDot()
	}

	return fa
}

func SetupFakeSim() *Simulation {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{
			RandomSeed: 100,
		},
		Raid: &proto.Raid{
			Parties: []*proto.Party{
				{
					Players: []*proto.Player{
						{
							Name:      "Caster",
							Class:     proto.Class_ClassShaman,
							Buffs:     &proto.IndividualBuffs{},
							Spec:      &proto.Player_ElementalShaman{},
							Equipment: &proto.EquipmentSpec{},
						},
					},
					Buffs: &proto.PartyBuffs{},
				},
			},
		},
		Encounter: &proto.Encounter{
			Targets: []*proto.Target{
				{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon},
			},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim
}

func expectDotTickDamage(t *testing.T, sim *Simulation, dot *Dot, expectedDamage float64) {
	damageBefore := dot.Spell.SpellMetrics[0].TotalDamage
	dot.TickOnce(sim)
	damageAfter := dot.Spell.SpellMetrics[0].TotalDamage
	delta := damageAfter - damageBefore

	if !WithinToleranceFloat64(expectedDamage, delta, 0.01) {
		t.Fatalf("Incorrect tick damage applied: Expected: %0.3f, Actual: %0.3f", expectedDamage, delta)
	}
}

func TestBleedDoesNotLandOnAnImmuneTarget(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	fa.Dot.Unit.BleedImmune = true

	fa.Dot.Apply(sim)
	if !fa.Dot.IsActive() {
		t.Fatal("a Dot that is not a bleed did not land on a bleed-immune target")
	}
	fa.Dot.Deactivate(sim)

	fa.Spell.Flags |= SpellFlagBleed
	fa.Dot.Apply(sim)
	if fa.Dot.IsActive() {
		t.Fatal("a bleed landed on a bleed-immune target")
	}

	fa.Dot.Unit.BleedImmune = false
	fa.Dot.Apply(sim)
	if !fa.Dot.IsActive() {
		t.Fatal("a bleed did not land on a target that takes bleeds")
	}
}

func TestBleedHitMissesAnImmuneTarget(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	target := sim.Encounter.AllTargetUnits[0]
	target.BleedImmune = true
	always := func(_ *Simulation, result *SpellResult, _ *AttackTable) { result.Outcome = OutcomeHit }

	if result := fa.Spell.CalcDamage(sim, target, 100, always); !result.Landed() {
		t.Fatal("a spell with no bleed hit missed a bleed-immune target")
	}
	fa.Spell.Flags |= SpellFlagBleed
	if result := fa.Spell.CalcDamage(sim, target, 100, always); result.Landed() || result.Damage != 0 {
		t.Fatalf("a bleed hit landed for %v on a bleed-immune target", result.Damage)
	}
	if result := fa.Spell.CalcOutcome(sim, target, always); result.Landed() {
		t.Fatal("a bleed outcome landed on a bleed-immune target")
	}
	target.BleedImmune = false
	if result := fa.Spell.CalcDamage(sim, target, 100, always); !result.Landed() {
		t.Fatal("a bleed hit missed a target that takes bleeds")
	}
}

func TestTargetReadsBleedImmunity(t *testing.T) {
	if !NewTarget(&proto.Target{BleedImmune: true}, 0).BleedImmune {
		t.Fatal("the target setting did not reach the unit")
	}
	if NewTarget(&proto.Target{}, 0).BleedImmune {
		t.Fatal("a target is bleed-immune by default")
	}
}

func TestChanneledTicksKeepTheTicksTheChannelStartedWith(t *testing.T) {
	var shorterTicks *SpellMod
	ct := newChannelTest(func(_ *channelTest, char *Character) *Spell {
		shorterTicks = char.AddDynamicMod(SpellModConfig{Kind: SpellMod_DotTickLength_Flat, SpellFlag: SpellFlagChanneled, TimeValue: -500 * time.Millisecond})
		return char.RegisterSpell(manaChannelConfig(5, false))
	})
	ticks := ct.rot.newValueSpellChanneledTicks(&proto.APLValueSpellChanneledTicks{SpellId: ct.spell.ActionID.ToProto()}, nil)

	ct.start()
	shorterTicks.Activate()
	ct.spell.Cast(ct.sim, ct.fw.CurrentTarget)
	shorterTicks.Deactivate() // the mod drops once the channel has taken its 0.5 s ticks

	for want := int32(1); want <= 4; want++ {
		ct.runUntil(time.Duration(want) * 500 * time.Millisecond)
		if got := ticks.GetInt(ct.sim); got != want {
			t.Fatalf("Channeled Ticks after tick %d = %d, want %d", want, got, want)
		}
	}
}

func TestDotTickCountKeepsTheTicksDoneOnExtend(t *testing.T) {
	ct := newChannelTest(func(_ *channelTest, char *Character) *Spell {
		return char.RegisterSpell(fakeWarlockDot(910301, "Dot", ClassFlags{}, 0))
	})
	ct.start()
	ct.spell.Cast(ct.sim, ct.fw.CurrentTarget)
	dot := ct.spell.Dot(ct.fw.CurrentTarget)

	ct.runUntil(3 * time.Second)
	dot.DurationExtend(ct.sim, 2*time.Second)
	if got := dot.TickCount(); got != 3 {
		t.Fatalf("tick count after 3 ticks and an extension = %d, want 3", got)
	}
	ct.runUntil(4 * time.Second)
	if got := dot.TickCount(); got != 4 {
		t.Fatalf("tick count after the next tick = %d, want 4", got)
	}
}

type dotTick struct {
	at     time.Duration
	damage float64
}

// A fake dot that records the time and damage of each tick it deals.
func recordedDotConfig(id int32, ticks *[]dotTick) SpellConfig {
	config := fakeWarlockDot(id, "Dot", ClassFlags{}, SpellFlagIgnoreResists)
	onTick := config.Dot.OnTick
	config.Dot.OnTick = func(sim *Simulation, target *Unit, dot *Dot) {
		metrics := &dot.Spell.SpellMetrics[target.UnitIndex]
		damage := metrics.TotalDamage
		onTick(sim, target, dot)
		*ticks = append(*ticks, dotTick{sim.CurrentTime, metrics.TotalDamage - damage})
	}
	return config
}

// Checks the ticks against their times from the dot's application, each a full tick.
func checkDotTicks(t *testing.T, got []dotTick, start time.Duration, at []time.Duration) {
	t.Helper()
	times := MapSlice(got, func(tick dotTick) time.Duration { return tick.at - start })
	if len(got) != len(at) {
		t.Fatalf("%d ticks at %v, want %d at %v", len(got), times, len(at), at)
	}
	for i, tick := range got {
		if times[i] != at[i] {
			t.Errorf("tick %d at %v, want %v", i+1, times[i], at[i])
		}
		if math.Abs(tick.damage-got[0].damage) > 1e-9 {
			t.Errorf("tick %d at %v dealt %v, want a full tick's %v", i+1, times[i], tick.damage, got[0].damage)
		}
	}
}

// Times in seconds, to the millisecond.
func gridTimes(values ...float64) []time.Duration {
	return MapSlice(values, func(s float64) time.Duration { return time.Duration(math.Round(s*1000)) * time.Millisecond })
}

// A change made to the dot at a time from its application, queued to run before the dot's own tick due at that
// instant (first) or after it.
type dotEvent struct {
	at     time.Duration
	first  bool
	retime bool
	do     func(ct *channelTest, dot *Dot)
}

func retimeDot(at, period time.Duration, first bool) dotEvent {
	return dotEvent{at: at, first: first, retime: true, do: func(ct *channelTest, dot *Dot) { dot.SetTickPeriod(ct.sim, period) }}
}

// Applies a recorded dot of 5 ticks 1 s apart, runs the events and returns its ticks, the time it was applied
// and, for each event, whether the dot had a tick still queued at its instant. After a last event that re-times
// the dot, its remaining ticks are the ticks it deals later, and once it has ended it holds its normal period.
func runRetimedDot(t *testing.T, events []dotEvent) ([]dotTick, time.Duration, []bool) {
	t.Helper()
	var ticks []dotTick
	ct := newChannelTest(func(_ *channelTest, char *Character) *Spell {
		return char.RegisterSpell(recordedDotConfig(910310, &ticks))
	})
	ct.start()
	start := ct.sim.CurrentTime
	ct.spell.Cast(ct.sim, ct.fw.CurrentTarget)
	dot := ct.spell.Dot(ct.fw.CurrentTarget)
	if got, want := dot.ExpiresAt()-start, 5*time.Second; got != want {
		t.Fatalf("the dot expires at %v, want %v", got, want)
	}
	queued := make([]bool, len(events))
	var remaining int32
	for i, e := range events {
		priority := ActionPriorityLow
		if e.first {
			priority = ActionPriorityDOT
		}
		ct.sim.AddPendingAction(&PendingAction{NextActionAt: start + e.at, Priority: priority, OnAction: func(sim *Simulation) {
			queued[i] = dot.IsActive() && dot.NextTickAt() == sim.CurrentTime
			e.do(ct, dot)
			remaining = dot.RemainingTicks()
		}})
	}
	ct.runUntil(start + 12*time.Second)
	if n := len(events); n > 0 && events[n-1].retime {
		later := int32(0)
		for _, tick := range ticks {
			if tick.at > start+events[n-1].at {
				later++
			}
		}
		if remaining != later {
			t.Errorf("%d ticks remaining after the last re-timing, want the %d it dealt later", remaining, later)
		}
	}
	if dot.IsActive() {
		t.Fatalf("the dot is still up at %v", ct.sim.CurrentTime-start)
	}
	return ticks, start, queued
}

// A dot re-timed while it runs ticks at its expiry less whole periods, which for a dot applied at its normal
// period is its application plus whole periods. Half its period keeps every normal tick and adds the points
// between, the points at the instants it starts and ends included and one already passed not; its period again
// puts it back on its normal ticks. Re-timed at its application, it does not tick there. The last tick lands on
// the expiry in full, also at a period that does not divide the duration. One refreshed or cancelled ticks as any
// dot does.
func TestDotRetimedTicksOnItsGrid(t *testing.T) {
	ms := time.Millisecond
	at := func(at time.Duration, do func(ct *channelTest, dot *Dot)) dotEvent {
		return dotEvent{at: at, do: do}
	}
	recast := func(ct *channelTest, _ *Dot) { ct.spell.Cast(ct.sim, ct.fw.CurrentTarget) }
	cancel := func(ct *channelTest, dot *Dot) { dot.Deactivate(ct.sim) }
	for _, tc := range []struct {
		name   string
		events []dotEvent
		ticks  []time.Duration
	}{
		{name: "never re-timed", ticks: gridTimes(1, 2, 3, 4, 5)},
		{
			name:   "half its period from 1.2 s to 3.2 s",
			events: []dotEvent{retimeDot(1200*ms, 500*ms, false), retimeDot(3200*ms, time.Second, false)},
			ticks:  gridTimes(1, 1.5, 2, 2.5, 3, 4, 5),
		},
		{
			name:   "half its period from 1.5 s to 3.5 s, both half points",
			events: []dotEvent{retimeDot(1500*ms, 500*ms, false), retimeDot(3500*ms, time.Second, false)},
			ticks:  gridTimes(1, 1.5, 2, 2.5, 3, 3.5, 4, 5),
		},
		{
			name:   "half its period from 1.6 s, past a half point",
			events: []dotEvent{retimeDot(1600*ms, 500*ms, false)},
			ticks:  gridTimes(1, 2, 2.5, 3, 3.5, 4, 4.5, 5),
		},
		{
			name:   "half its period from its application",
			events: []dotEvent{retimeDot(0, 500*ms, false)},
			ticks:  gridTimes(0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5),
		},
		{
			name:   "half its period from 3.7 s to its expiry",
			events: []dotEvent{retimeDot(3700*ms, 500*ms, false)},
			ticks:  gridTimes(1, 2, 3, 4, 4.5, 5),
		},
		{
			name:   "a period that does not divide its duration",
			events: []dotEvent{retimeDot(1200*ms, 400*ms, false)},
			ticks:  gridTimes(1, 1.4, 1.8, 2.2, 2.6, 3, 3.4, 3.8, 4.2, 4.6, 5),
		},
		{
			name:   "refreshed before its expiry",
			events: []dotEvent{retimeDot(1200*ms, 500*ms, false), at(4950*ms, recast)},
			ticks:  gridTimes(1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5.95, 6.95, 7.95, 8.95, 9.95),
		},
		{
			name:   "cancelled before its expiry",
			events: []dotEvent{retimeDot(1200*ms, 500*ms, false), at(4950*ms, cancel)},
			ticks:  gridTimes(1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ticks, start, _ := runRetimedDot(t, tc.events)
			checkDotTicks(t, ticks, start, tc.ticks)
		})
	}
}

// A re-timing at an instant the dot has a tick due deals that tick once, whether it runs before or after the
// tick, and the next tick falls on the new period's grid. Ending a half period on a half point and starting it
// again there, as a channel recast on the instant it ends, deals that point once.
func TestDotRetimedAtATickDealsItOnce(t *testing.T) {
	ms := time.Millisecond
	for _, tc := range []struct {
		name   string
		events func(first bool) []dotEvent
		at     int // The event at the instant of the tick.
		ticks  []time.Duration
	}{
		{
			name:   "half its period from a tick",
			events: func(first bool) []dotEvent { return []dotEvent{retimeDot(2000*ms, 500*ms, first)} },
			ticks:  gridTimes(1, 2, 2.5, 3, 3.5, 4, 4.5, 5),
		},
		{
			name: "its period again on a half point",
			events: func(first bool) []dotEvent {
				return []dotEvent{retimeDot(1200*ms, 500*ms, false), retimeDot(3500*ms, time.Second, first)}
			},
			at:    1,
			ticks: gridTimes(1, 1.5, 2, 2.5, 3, 3.5, 4, 5),
		},
		{
			name: "half its period again on the half point it ended on",
			events: func(first bool) []dotEvent {
				return []dotEvent{retimeDot(1200*ms, 500*ms, false), retimeDot(3500*ms, time.Second, first), retimeDot(3500*ms, 500*ms, first)}
			},
			at:    1,
			ticks: gridTimes(1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5),
		},
	} {
		for _, first := range []bool{true, false} {
			t.Run(tc.name+map[bool]string{true: ", before the tick", false: ", after the tick"}[first], func(t *testing.T) {
				ticks, start, queued := runRetimedDot(t, tc.events(first))
				if queued[tc.at] != first {
					t.Fatalf("the dot's tick was queued at the event: %v, want %v", queued[tc.at], first)
				}
				checkDotTicks(t, ticks, start, tc.ticks)
			})
		}
	}
}

// A dot re-timed from inside its own tick deals that tick once, and its next tick is the next point of the new
// period's grid, also when that tick is a half point and the period goes back to the normal one. Its tick stays
// queued once, so an action due between two ticks runs on time.
func TestDotRetimedFromItsOwnTick(t *testing.T) {
	ms := time.Millisecond
	for _, tc := range []struct {
		name   string
		before bool  // Half its period from 1.2 s, before the tick.
		atTick int32 // The tick that sets the period.
		period time.Duration
		ticks  []time.Duration
	}{
		{name: "half its period from its tick at 1 s", atTick: 1, period: 500 * ms, ticks: gridTimes(1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5)},
		{name: "its period again from its tick on the half point 1.5 s", before: true, atTick: 2, period: time.Second, ticks: gridTimes(1, 1.5, 2, 3, 4, 5)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var ticks []dotTick
			var remaining []int32
			ct := newChannelTest(func(_ *channelTest, char *Character) *Spell {
				config := recordedDotConfig(910311, &ticks)
				onTick := config.Dot.OnTick
				config.Dot.OnTick = func(sim *Simulation, target *Unit, dot *Dot) {
					onTick(sim, target, dot)
					if dot.TickCount() == tc.atTick {
						dot.SetTickPeriod(sim, tc.period)
					}
					if dot.TickCount() >= tc.atTick {
						remaining = append(remaining, dot.RemainingTicks())
					}
				}
				return char.RegisterSpell(config)
			})
			ct.start()
			start := ct.sim.CurrentTime
			ct.spell.Cast(ct.sim, ct.fw.CurrentTarget)
			dot := ct.spell.Dot(ct.fw.CurrentTarget)
			if tc.before {
				ct.sim.AddPendingAction(&PendingAction{NextActionAt: start + 1200*ms, OnAction: func(sim *Simulation) {
					dot.SetTickPeriod(sim, 500*ms)
				}})
			}
			var probeRanAt time.Duration
			ct.sim.AddPendingAction(&PendingAction{NextActionAt: start + 1750*ms, OnAction: func(sim *Simulation) {
				probeRanAt = sim.CurrentTime
			}})

			ct.runUntil(start + 12*time.Second)
			if got, want := probeRanAt-start, 1750*ms; got != want {
				t.Errorf("an action due at %v ran at %v", want, got)
			}
			checkDotTicks(t, ticks, start, tc.ticks)
			for i, got := range remaining {
				tick := int(tc.atTick) + i
				if want := int32(len(ticks) - tick); got != want {
					t.Errorf("%d ticks remaining after tick %d, want %d", got, tick, want)
				}
			}
		})
	}
}

// A dot whose period is set at the instant it expires, before its own expiry is handled (as by a channel that
// ends with it), deals its last tick at the expiry once, in full.
func TestDotRetimedAtItsExpiryDealsItsLastTickOnce(t *testing.T) {
	ms := time.Millisecond
	var ticks []dotTick
	var endsWithDot *Aura
	dotUpAtEnd := false
	ct := newChannelTest(func(_ *channelTest, char *Character) *Spell {
		spell := char.RegisterSpell(recordedDotConfig(910312, &ticks))
		endsWithDot = char.CurrentTarget.RegisterAura(Aura{
			Label:    "Ends With The Dot",
			Duration: 5 * time.Second,
			OnExpire: func(aura *Aura, sim *Simulation) {
				dot := spell.Dot(aura.Unit)
				dotUpAtEnd = dot.IsActive()
				dot.SetTickPeriod(sim, time.Second)
			},
		})
		return spell
	})
	ct.start()
	start := ct.sim.CurrentTime
	// Activated before the dot, so its expiry is handled first at the instant both expire.
	endsWithDot.Activate(ct.sim)
	ct.spell.Cast(ct.sim, ct.fw.CurrentTarget)
	dot := ct.spell.Dot(ct.fw.CurrentTarget)
	if dot.ExpiresAt() != endsWithDot.ExpiresAt() {
		t.Fatalf("the dot expires at %v and the aura at %v, want the same instant", dot.ExpiresAt()-start, endsWithDot.ExpiresAt()-start)
	}

	ct.runUntil(start + 1200*ms)
	dot.SetTickPeriod(ct.sim, 500*ms)
	ct.runUntil(start + 12*time.Second)
	if !dotUpAtEnd {
		t.Fatal("the dot had expired when its period was set at its expiry, so the case is not reached")
	}
	checkDotTicks(t, ticks, start, gridTimes(1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5))
}

func TestDotSnapshot(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)

	fa.Dot.Apply(sim)
	expectDotTickDamage(t, sim, fa.Dot, 150) // (100) * 1.5
}

func TestDotSnapshotSpellDamage(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)

	fa.Dot.Apply(sim)
	expectDotTickDamage(t, sim, fa.Dot, 150) // (100) * 1.5

	// Spell power shouldn't get applied because dot was already snapshot.
	fa.GetCharacter().AddStatDynamic(sim, stats.SpellDamage, 100)
	expectDotTickDamage(t, sim, fa.Dot, 150) // (100) * 1.5

	fa.Dot.Deactivate(sim)
	fa.Dot.Apply(sim)
	expectDotTickDamage(t, sim, fa.Dot, 300) // (100 + 100) * 1.5
}

func TestDotSnapshotSpellMultiplier(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	spell := fa.GetCharacter().Spellbook[0]
	spell.DamageMultiplier *= 2

	fa.Dot.Apply(sim)
	expectDotTickDamage(t, sim, fa.Dot, 300) // (100) * 1.5 * 2
}
