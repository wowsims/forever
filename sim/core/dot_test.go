package core

import (
	"math"
	"slices"
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

// Checks the ticks against their times and their shares of a full tick: each is a full one but the last,
// which deals lastShare of one.
func checkDotTicks(t *testing.T, got []dotTick, start time.Duration, at []time.Duration, lastShare float64) {
	t.Helper()
	if len(got) != len(at) {
		t.Fatalf("%d ticks at %v, want %d at %v", len(got), got, len(at), at)
	}
	for i, tick := range got {
		if tick.at-start != at[i] {
			t.Errorf("tick %d at %v, want %v", i+1, tick.at-start, at[i])
		}
		share := 1.0
		if i == len(got)-1 {
			share = lastShare
		}
		if want := share * got[0].damage; math.Abs(tick.damage-want) > 1e-9 {
			t.Errorf("tick %d at %v dealt %v, want %v (%v of a full tick)", i+1, tick.at-start, tick.damage, want, share)
		}
	}
}

// A dot re-timed while it runs keeps its expiry and its share of progress to the next tick, both ways: a tick
// 0.5 s away at 1 s ticks is 0.25 s away at 0.5 s ticks. One whose expiry then falls part-way to a tick deals
// that tick at the expiry, truncated to the share of the period it ran; one back on its grid, or never re-timed,
// ends on a whole tick. One refreshed or cancelled before its expiry deals no truncated tick. An ended dot
// ticks at its normal period again.
func TestDotRetimedEndsWithATruncatedTick(t *testing.T) {
	ms := time.Millisecond
	type event struct {
		at        time.Duration
		do        func(ct *channelTest, dot *Dot)
		untilNext time.Duration // The time to the next tick after the event, when checked.
	}
	retime := func(at, period, untilNext time.Duration) event {
		return event{at: at, untilNext: untilNext, do: func(ct *channelTest, dot *Dot) { dot.SetTickPeriod(ct.sim, period) }}
	}
	endsPartWay := []event{retime(1500*ms, 500*ms, 250*ms), retime(1600*ms, 1000*ms, 300*ms)}
	for _, tc := range []struct {
		name      string
		events    []event
		ticks     []time.Duration
		lastShare float64
	}{
		{
			name:      "never re-timed",
			ticks:     []time.Duration{1000 * ms, 2000 * ms, 3000 * ms, 4000 * ms, 5000 * ms},
			lastShare: 1,
		},
		{
			name:      "re-timed back onto its grid",
			events:    []event{retime(1500*ms, 500*ms, 250*ms), retime(2500*ms, 1000*ms, 500*ms)},
			ticks:     []time.Duration{1000 * ms, 1750 * ms, 2250 * ms, 3000 * ms, 4000 * ms, 5000 * ms},
			lastShare: 1,
		},
		{
			name:      "ends part-way to a tick at the new period",
			events:    []event{retime(1500*ms, 500*ms, 250*ms)},
			ticks:     []time.Duration{1000 * ms, 1750 * ms, 2250 * ms, 2750 * ms, 3250 * ms, 3750 * ms, 4250 * ms, 4750 * ms, 5000 * ms},
			lastShare: 0.5,
		},
		{
			name:      "ends part-way to a tick",
			events:    endsPartWay,
			ticks:     []time.Duration{1000 * ms, 1900 * ms, 2900 * ms, 3900 * ms, 4900 * ms, 5000 * ms},
			lastShare: 0.1,
		},
		{
			name: "refreshed before its expiry",
			events: append(slices.Clone(endsPartWay), event{at: 4950 * ms, do: func(ct *channelTest, dot *Dot) {
				ct.spell.Cast(ct.sim, ct.fw.CurrentTarget)
			}}),
			ticks:     []time.Duration{1000 * ms, 1900 * ms, 2900 * ms, 3900 * ms, 4900 * ms, 5950 * ms, 6950 * ms, 7950 * ms, 8950 * ms, 9950 * ms},
			lastShare: 1,
		},
		{
			name: "cancelled before its expiry",
			events: append(slices.Clone(endsPartWay), event{at: 4950 * ms, do: func(ct *channelTest, dot *Dot) {
				dot.Deactivate(ct.sim)
			}}),
			ticks:     []time.Duration{1000 * ms, 1900 * ms, 2900 * ms, 3900 * ms, 4900 * ms},
			lastShare: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
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

			for _, e := range tc.events {
				ct.runUntil(start + e.at)
				e.do(ct, dot)
				if e.untilNext != 0 {
					if got := dot.TimeUntilNextTick(ct.sim); got != e.untilNext {
						t.Errorf("next tick %v after the event at %v, want %v", got, e.at, e.untilNext)
					}
				}
			}
			ct.runUntil(start + 12*time.Second)
			checkDotTicks(t, ticks, start, tc.ticks, tc.lastShare)
			if got := dot.TickPeriod(); got != time.Second {
				t.Errorf("the ended dot ticks every %v, want its normal %v", got, time.Second)
			}
		})
	}
}

// A dot re-timed from inside its own tick deals that tick once and ticks at the new period from it, and its
// tick stays queued once, so an action due between two ticks runs on time.
func TestDotRetimedFromItsOwnTick(t *testing.T) {
	ms := time.Millisecond
	var ticks []dotTick
	ct := newChannelTest(func(_ *channelTest, char *Character) *Spell {
		config := recordedDotConfig(910311, &ticks)
		onTick := config.Dot.OnTick
		config.Dot.OnTick = func(sim *Simulation, target *Unit, dot *Dot) {
			onTick(sim, target, dot)
			if dot.TickCount() == 1 {
				dot.SetTickPeriod(sim, 500*ms)
			}
		}
		return char.RegisterSpell(config)
	})
	ct.start()
	start := ct.sim.CurrentTime
	ct.spell.Cast(ct.sim, ct.fw.CurrentTarget)
	var probeRanAt time.Duration
	ct.sim.AddPendingAction(&PendingAction{NextActionAt: start + 1750*ms, OnAction: func(sim *Simulation) {
		probeRanAt = sim.CurrentTime
	}})

	ct.runUntil(start + 12*time.Second)
	if got, want := probeRanAt-start, 1750*ms; got != want {
		t.Errorf("an action due at %v ran at %v", want, got)
	}
	var at []time.Duration
	for tick := 1000 * ms; tick <= 5000*ms; tick += 500 * ms {
		at = append(at, tick)
	}
	checkDotTicks(t, ticks, start, at, 1)
}

// A dot re-timed off its grid whose period is set again at the instant it expires, before its own expiry is
// handled (as by a channel that ends with it), still deals its truncated tick at the expiry: half a tick, the
// share of the period it ran, half of that progress made at twice the normal speed.
func TestDotRetimedAtItsExpiryDealsItsTruncatedTick(t *testing.T) {
	ms := time.Millisecond
	var ticks []dotTick
	var lastSpedUp float64
	var endsWithDot *Aura
	dotUpAtEnd := false
	ct := newChannelTest(func(_ *channelTest, char *Character) *Spell {
		config := recordedDotConfig(910312, &ticks)
		onTick := config.Dot.OnTick
		config.Dot.OnTick = func(sim *Simulation, target *Unit, dot *Dot) {
			onTick(sim, target, dot)
			lastSpedUp = dot.SpedUpShare(sim)
		}
		spell := char.RegisterSpell(config)
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

	ct.runUntil(start + 1500*ms)
	dot.SetTickPeriod(ct.sim, 500*ms)
	ct.runUntil(start + 12*time.Second)
	if !dotUpAtEnd {
		t.Fatal("the dot had expired when its period was set at its expiry, so the case is not reached")
	}
	checkDotTicks(t, ticks, start, []time.Duration{1000 * ms, 1750 * ms, 2250 * ms, 2750 * ms, 3250 * ms, 3750 * ms, 4250 * ms, 4750 * ms, 5000 * ms}, 0.5)
	if math.Abs(lastSpedUp-0.5) > 1e-9 {
		t.Errorf("the truncated tick's sped-up share is %v, want 0.5", lastSpedUp)
	}
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
