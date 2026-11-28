package core

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

const manaTestStep = 100 * time.Millisecond

// A channel of 1 s ticks, paying 100 mana when cost is set.
func manaChannelConfig(ticks int32, cost bool) SpellConfig {
	config := fakeWarlockDot(910300, "Mana Channel", ClassFlags{}, SpellFlagChanneled|SpellFlagAPL)
	config.Dot.NumberOfTicks = ticks
	if cost {
		config.ManaCost = ManaCostOptions{FlatCost: 100}
	}
	return config
}

// A spell paying 100 mana, instant when castTime is 0.
func paidSpellConfig(id int32, castTime time.Duration) SpellConfig {
	config := fakeWarlockDot(id, "Paid Spell", ClassFlags{}, SpellFlagAPL)
	config.Dot = DotConfig{}
	config.ApplyEffects = func(_ *Simulation, _ *Unit, _ *Spell) {}
	config.ManaCost = ManaCostOptions{FlatCost: 100}
	config.Cast = CastConfig{DefaultCast: Cast{CastTime: castTime}}
	return config
}

// The warrior gets a mana bar with spirit regen, none of which goes on while casting.
func newManaTickTest(register func(char *Character) *Spell) *channelTest {
	return newChannelTest(func(ct *channelTest, char *Character) *Spell {
		char.manaBar.unit = &char.Unit
		char.manaCastingMetrics = char.NewManaMetrics(ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 1})
		char.manaNotCastingMetrics = char.NewManaMetrics(ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 2})
		char.SetSpiritManaRegen(10, 0)
		char.initialStats[stats.Mana] = 10000
		char.initialStatsWithoutDeps[stats.Mana] = 10000
		return register(char)
	})
}

// Casts the spell at 0 and runs to until in 100 ms steps, calling at[t] right after the step to t.
// Returns the spans with the regen of a unit outside the five-second rule as from, to pairs, each
// span running from just after from up to and including to.
func (ct *channelTest) fullRegenSpans(until time.Duration, at map[time.Duration]func(*channelTest)) []time.Duration {
	ct.start()
	if !ct.spell.Cast(ct.sim, ct.fw.CurrentTarget) {
		panic("the spell under test was not cast")
	}
	var spans []time.Duration
	for t := manaTestStep; t <= until; t += manaTestStep {
		before := ct.fw.manaNotCastingMetrics.Events
		ct.runUntil(t)
		if ct.fw.manaNotCastingMetrics.Events > before {
			if n := len(spans); n > 0 && spans[n-1] == t-manaTestStep {
				spans[n-1] = t
			} else {
				spans = append(spans, t-manaTestStep, t)
			}
		}
		if fn := at[t]; fn != nil {
			fn(ct)
		}
	}
	return spans
}

func TestFiveSecondRuleWindows(t *testing.T) {
	for _, tc := range []struct {
		name     string
		register func(char *Character) *Spell
		at       map[time.Duration]func(ct *channelTest)
		want     []float64
	}{
		{
			name:     "a paid 15 s channel holds regen until it ends",
			register: func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(15, true)) },
			want:     []float64{15, 20},
		},
		{
			name:     "a paid 5 s channel holds regen until it ends",
			register: func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(5, true)) },
			want:     []float64{5, 20},
		},
		{
			name:     "a paid channel cut short at 2 s holds regen until 5 s",
			register: func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(15, true)) },
			at:       map[time.Duration]func(ct *channelTest){2 * time.Second: func(ct *channelTest) { ct.fw.Interrupt(ct.sim) }},
			want:     []float64{5, 20},
		},
		{
			name:     "a costless channel leaves regen alone",
			register: func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(15, false)) },
			want:     []float64{0, 20},
		},
		{
			name:     "Shoot leaves regen alone",
			register: func(char *Character) *Spell { return fakeShoot(char, new([]time.Duration)) },
			want:     []float64{0, 20},
		},
		{
			name:     "a paid 3 s hardcast holds regen for 5 s from its end",
			register: func(char *Character) *Spell { return char.RegisterSpell(paidSpellConfig(910301, 3*time.Second)) },
			want:     []float64{0, 3, 8, 20},
		},
		{
			name:     "a paid instant holds regen for 5 s",
			register: func(char *Character) *Spell { return char.RegisterSpell(paidSpellConfig(910302, 0)) },
			want:     []float64{5, 20},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ct := newManaTickTest(tc.register)
			if got, want := ct.fullRegenSpans(20*time.Second, tc.at), seconds(tc.want...); !slices.Equal(got, want) {
				t.Errorf("full regen over the spans %v, want %v", got, want)
			}
		})
	}
}

// A channel whose dot sits on its related spell is paid for by the spell cast.
func TestPaidChannelOnItsRelatedSpellHoldsRegen(t *testing.T) {
	var related *Spell
	ct := newManaTickTest(func(char *Character) *Spell {
		related = char.RegisterSpell(manaChannelConfig(15, false))
		config := paidSpellConfig(910303, 0)
		config.Flags |= SpellFlagChanneled
		config.RelatedDotSpell = related
		config.ApplyEffects = func(sim *Simulation, target *Unit, _ *Spell) {
			related.Dot(target).Apply(sim)
		}
		return char.RegisterSpell(config)
	})
	related.finalize()
	if got, want := ct.fullRegenSpans(20*time.Second, nil), seconds(15, 20); !slices.Equal(got, want) {
		t.Errorf("full regen over the spans %v, want %v", got, want)
	}
}

// A paid cast whose channel never starts leaves no payment for a later free cast of it to claim.
func TestFreeChannelAfterAMissedPaidOneLeavesRegenAlone(t *testing.T) {
	landed := false
	ct := newManaTickTest(func(char *Character) *Spell {
		config := manaChannelConfig(15, true)
		applyEffects := config.ApplyEffects
		config.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
			if landed {
				applyEffects(sim, target, spell)
			}
		}
		return char.RegisterSpell(config)
	})
	castFree := func(ct *channelTest) {
		landed = true
		ct.spell.Cost.FlatModifier = -100
		if !ct.spell.Cast(ct.sim, ct.fw.CurrentTarget) || !ct.fw.IsChanneling() {
			panic("the free channel did not start")
		}
	}
	if got, want := ct.fullRegenSpans(25*time.Second, map[time.Duration]func(*channelTest){6 * time.Second: castFree}), seconds(5, 25); !slices.Equal(got, want) {
		t.Errorf("full regen over the spans %v, want %v", got, want)
	}
}

type manaAt struct {
	at   time.Duration
	mana float64
}

// Runs to each time in turn and checks the mana there.
func (ct *channelTest) wantManaAt(t *testing.T, checks ...manaAt) {
	t.Helper()
	for _, c := range checks {
		ct.runUntil(c.at)
		if got := ct.fw.CurrentMana(); math.Abs(got-c.mana) > 1e-9 {
			t.Errorf("%v mana at %v, want %v", got, c.at, c.mana)
		}
	}
}

// Outside the five-second rule mana rises by the regen rate times the time passed, at every step.
func TestManaRegenAccruesEveryTick(t *testing.T) {
	ct := newManaTickTest(func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(5, false)) })
	ct.start()
	ct.fw.currentMana = 1000
	var checks []manaAt
	for at := manaTestStep; at <= 10*time.Second; at += manaTestStep {
		checks = append(checks, manaAt{at, 1000 + 10*at.Seconds()})
	}
	ct.wantManaAt(t, checks...)
}

// Using mana at 1.25 s opens the five-second rule there: the regen up to the use counts in full, none
// accrues up to 6.25 s, and the full regen resumes at 6.25 s.
func TestFiveSecondRuleCountsFromTheSpend(t *testing.T) {
	ct := newManaTickTest(func(char *Character) *Spell { return char.RegisterSpell(paidSpellConfig(910302, 0)) })
	ct.start()
	ct.fw.currentMana = 1000
	ct.runUntil(1250 * time.Millisecond)
	if !ct.spell.Cast(ct.sim, ct.fw.CurrentTarget) {
		t.Fatal("the paid instant was not cast")
	}
	if got, want := ct.fw.TimeUntilManaRegen(ct.sim, 912.5+10), 6050*time.Millisecond; got != want {
		t.Errorf("%v until 10 more mana, want %v: a second of full regen from 6.25 s, credited on the tick at 7.3 s", got, want)
	}
	ct.wantManaAt(t,
		manaAt{1250 * time.Millisecond, 1000 + 12.5 - 100},
		manaAt{3 * time.Second, 912.5},
		manaAt{6250 * time.Millisecond, 912.5},
		manaAt{6300 * time.Millisecond, 912.5 + 0.5},
		manaAt{7300 * time.Millisecond, 912.5 + 10.5},
	)
}

// A regen rate change between two steps counts the time before it at the old rate.
func TestManaRegenRateChangeCreditsTheOldRate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(ct *channelTest)
	}{
		{"a regen speed multiplier", func(ct *channelTest) { ct.fw.MultiplyManaRegenSpeed(ct.sim, 2) }},
		{"an MP5 gain", func(ct *channelTest) { ct.fw.AddStatDynamic(ct.sim, stats.MP5, 50) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ct := newManaTickTest(func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(5, false)) })
			ct.start()
			ct.fw.currentMana = 1000
			ct.runUntil(1050 * time.Millisecond)
			tc.change(ct)
			ct.wantManaAt(t, manaAt{2 * time.Second, 1000 + 10*1.05 + 20*0.95})
		})
	}
}

// The MP5 regen that goes on inside the five-second rule accrues at every step too.
func TestMP5AccruesInsideTheFiveSecondRule(t *testing.T) {
	ct := newManaTickTest(func(char *Character) *Spell { return char.RegisterSpell(paidSpellConfig(910302, 0)) })
	ct.start()
	ct.fw.AddStatDynamic(ct.sim, stats.MP5, 50)
	ct.fw.currentMana = 1000
	if !ct.spell.Cast(ct.sim, ct.fw.CurrentTarget) {
		t.Fatal("the paid instant was not cast")
	}
	var checks []manaAt
	for at := manaTestStep; at <= 5*time.Second; at += manaTestStep {
		checks = append(checks, manaAt{at, 900 + 10*at.Seconds()})
	}
	ct.wantManaAt(t, checks...)
}

// A mana bar reset mid-fight, as when a pet is summoned again, keeps ticking and regenerates from there.
func TestManaBarResetMidFightKeepsTicking(t *testing.T) {
	ct := newManaTickTest(func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(5, false)) })
	ct.start()
	ct.runUntil(time.Second)
	ct.fw.manaBar.reset()
	ct.fw.currentMana = 1000
	ct.wantManaAt(t, manaAt{3 * time.Second, 1000 + 10*2})
}
