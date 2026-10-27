package core

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestQueuedSpellRotationPass(t *testing.T) {
	// The GCD ends at 1.5 s and the spell is queued for it at 1.2 s. The rotation's next look, a reaction
	// time poll, falls at 1.55 s, past the GCD end; the action below the spell becomes ready at 1.5 s.
	queueAt := func(config SpellConfig) (*channelTest, *timedStubAction) {
		ct, spell := newSpellShootTest(config)
		after := &timedStubAction{readyAt: 1500 * time.Millisecond}
		ct.start(after)
		ct.runUntil(0)
		ct.fw.SetGCDTimer(ct.sim, 1500*time.Millisecond)
		ct.runUntil(1200 * time.Millisecond)
		spell.CastOrQueue(ct.sim, ct.fw.CurrentTarget)
		if ct.fw.QueuedSpell == nil || ct.fw.QueuedSpell.spell != spell {
			t.Fatalf("the spell should be queued for the GCD end at 1.5 s")
		}
		ct.fw.SetRotationTimer(ct.sim, 1550*time.Millisecond)
		return ct, after
	}

	t.Run("an instant lets the rotation act at once", func(t *testing.T) {
		var castAt []time.Duration
		ct, after := queueAt(nukeConfig(&castAt))

		ct.runUntil(1501 * time.Millisecond)
		if len(castAt) != 1 || castAt[0] < 1500*time.Millisecond || castAt[0] > 1501*time.Millisecond {
			t.Fatalf("the queued instant should go off as the GCD ends at 1.5 s, cast at %v", castAt)
		}
		if len(after.ranAt) != 1 || after.ranAt[0] != castAt[0] {
			t.Fatalf("the action below the spell should run as the queued instant goes off at %v, ran at %v (the GCD it starts ends at 3 s)", castAt[0], after.ranAt)
		}
	})

	t.Run("a hardcast keeps the rotation for the cast's end", func(t *testing.T) {
		var castAt []time.Duration
		config := nukeConfig(&castAt)
		config.Cast.DefaultCast.CastTime = time.Second
		ct, after := queueAt(config)

		ct.runUntil(2400 * time.Millisecond)
		if len(after.ranAt) != 0 {
			t.Fatalf("the rotation should wait out the 1 s cast that started at 1.5 s, the action ran at %v", after.ranAt)
		}
		ct.runUntil(3001 * time.Millisecond)
		if !slices.Equal(castAt, []time.Duration{2500*time.Millisecond + 1}) || len(after.ranAt) != 1 {
			t.Fatalf("the cast should land at 2.5 s and the rotation act by its GCD end at 3 s, cast at %v, the action ran at %v", castAt, after.ranAt)
		}
	})
}

// Never ready; records each time the rotation looks at it.
type lookProbe struct {
	defaultAPLActionImpl
	lookedAt []time.Duration
}

func (probe *lookProbe) IsReady(sim *Simulation) bool {
	if n := len(probe.lookedAt); n == 0 || probe.lookedAt[n-1] != sim.CurrentTime {
		probe.lookedAt = append(probe.lookedAt, sim.CurrentTime)
	}
	return false
}
func (probe *lookProbe) Execute(*Simulation) {}
func (probe *lookProbe) String() string      { return "look probe" }

func (probe *lookProbe) looked(at time.Duration) bool {
	return slices.Contains(probe.lookedAt, at)
}

// Ready while ready says so and the action it wraps is.
type gatedAction struct {
	APLActionImpl
	ready func(sim *Simulation) bool
}

func (action *gatedAction) IsReady(sim *Simulation) bool {
	return action.ready(sim) && action.APLActionImpl.IsReady(sim)
}

// A 2 s cast on the 1.5 s GCD that puts up a one-minute aura of another id as it lands, an instant on the GCD,
// an instant off it with a one-minute cooldown and a 1 s cast on the GCD, each recording when it lands; and two
// more 2 s casts: one applying its own 5 s dot, one putting up its own 10 s buff of up to 3 stacks on the caster;
// and a Drain Life channel. The instants, the major cooldowns and the channel also note their names in order.
// Blast is a 2 s cast with an 8 s cooldown and a 4 s cooldown it shares with the instant shock, both starting
// when it lands.
type hardcastTest struct {
	*channelTest
	slow, nuke, offGCD, short         *Spell
	slowAt, nukeAt, offGCDAt, shortAt []time.Duration
	blast, shock                      *Spell
	blastAt                           []time.Duration
	slowAura                          *Aura
	dotCast, buffCast                 *Spell
	buff                              *Aura
	drain                             *Spell
	drainTicks                        []time.Duration
	order                             []string
}

// The config, noting name in the test's order as its effects apply.
func (cb *hardcastTest) ordered(name string, config SpellConfig) SpellConfig {
	applyEffects := config.ApplyEffects
	config.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
		cb.order = append(cb.order, name)
		applyEffects(sim, target, spell)
	}
	return config
}

func newHardcastTest(cooldowns ...*testCooldown) *hardcastTest {
	cb := &hardcastTest{}
	cb.channelTest = newChannelTest(func(ct *channelTest, char *Character) *Spell {
		cb.slowAura = char.RegisterAura(Aura{Label: "Slow Cast Aura", ActionID: ActionID{SpellID: 910141}, Duration: time.Minute})
		slow := recordedSpellConfig(910140, &cb.slowAt)
		slow.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: 2 * time.Second}}
		slow.ApplyEffects = func(sim *Simulation, _ *Unit, _ *Spell) {
			cb.slowAt = append(cb.slowAt, sim.CurrentTime)
			cb.slowAura.Activate(sim)
		}
		cb.slow = char.RegisterSpell(slow)
		cb.nuke = char.RegisterSpell(cb.ordered("nuke", nukeConfig(&cb.nukeAt)))
		offGCD := recordedSpellConfig(910142, &cb.offGCDAt)
		offGCD.Cast = CastConfig{CD: Cooldown{Timer: char.NewTimer(), Duration: time.Minute}}
		cb.offGCD = char.RegisterSpell(cb.ordered("off the GCD", offGCD))
		cb.drain = char.RegisterSpell(cb.ordered("drain", drainConfig(&cb.drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}})))
		short := recordedSpellConfig(910143, &cb.shortAt)
		short.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: time.Second}}
		cb.short = char.RegisterSpell(short)
		sharedTimer := char.NewTimer()
		blast := recordedSpellConfig(910144, &cb.blastAt)
		blast.Cast = CastConfig{
			DefaultCast: Cast{GCD: GCDDefault, CastTime: 2 * time.Second},
			CD:          Cooldown{Timer: char.NewTimer(), Duration: 8 * time.Second},
			SharedCD:    Cooldown{Timer: sharedTimer, Duration: 4 * time.Second},
		}
		cb.blast = char.RegisterSpell(blast)
		shock := recordedSpellConfig(910145, &[]time.Duration{})
		shock.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault}, SharedCD: Cooldown{Timer: sharedTimer, Duration: 4 * time.Second}}
		cb.shock = char.RegisterSpell(shock)
		dotCast := fakeWarlockDot(910160, "Hardcast Dot", ClassFlags{}, SpellFlagAPL)
		dotCast.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: 2 * time.Second}}
		cb.dotCast = char.RegisterSpell(dotCast)
		cb.buff = char.RegisterAura(Aura{Label: "Hardcast Buff", ActionID: ActionID{SpellID: 910161}, Duration: 10 * time.Second, MaxStacks: 3})
		buffCast := recordedSpellConfig(910161, &[]time.Duration{})
		buffCast.Flags |= SpellFlagHelpful
		buffCast.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: 2 * time.Second}}
		buffCast.ApplyEffects = func(sim *Simulation, _ *Unit, _ *Spell) {
			cb.buff.Activate(sim)
			cb.buff.AddStack(sim)
		}
		cb.buffCast = char.RegisterSpell(buffCast)
		for i, cooldown := range cooldowns {
			config := recordedSpellConfig(910150+int32(i), &cooldown.castAt)
			config.Cast = CastConfig{DefaultCast: Cast{GCD: cooldown.gcd}, CD: Cooldown{Timer: char.NewTimer(), Duration: time.Minute}}
			cooldown.spell = char.RegisterSpell(cb.ordered("cooldown", config))
			char.AddMajorCooldown(MajorCooldown{Spell: cooldown.spell, Type: Ternary(cooldown.explosive, CooldownTypeExplosive, CooldownTypeDPS), Priority: cooldown.priority})
		}
		return fakeShoot(char, &ct.shots)
	})
	for _, spell := range []*Spell{cb.slow, cb.nuke, cb.offGCD, cb.short, cb.blast, cb.shock, cb.dotCast, cb.buffCast, cb.drain} {
		spell.finalize()
	}
	cb.fw.majorCooldownManager.finalize()
	cb.fw.majorCooldownManager.reset(cb.sim)
	for _, cooldown := range cooldowns {
		cooldown.spell.finalize()
		cooldown.spell.CD.Set(1500 * time.Millisecond)
	}
	cb.fw.UpdateMajorCooldowns()
	return cb
}

// The slow cast, only as the rotation's first action at 0.
func (cb *hardcastTest) slowAtPull() APLActionImpl {
	return &gatedAction{APLActionImpl: cb.castOf(cb.slow), ready: func(sim *Simulation) bool { return sim.CurrentTime == 0 }}
}

// The instant's action, noting how many slow casts had landed each time the rotation examines it at `at`.
func (cb *hardcastTest) nukeNotingLandings(at time.Duration, landed *[]int) APLActionImpl {
	return &gatedAction{APLActionImpl: cb.castOf(cb.nuke), ready: func(sim *Simulation) bool {
		if sim.CurrentTime == at {
			*landed = append(*landed, len(cb.slowAt))
		}
		return true
	}}
}

func TestRotationPicksAtCastEndBeforeLanding(t *testing.T) {
	const castEnd = 2 * time.Second

	t.Run("the rotation picks at the cast end before the cast lands and its pick goes off as the cast ends", func(t *testing.T) {
		cb := newHardcastTest()
		probe := &lookProbe{}
		var landed []int
		cb.start(probe, cb.slowAtPull(), cb.nukeNotingLandings(castEnd, &landed))

		cb.runUntil(castEnd - 1)
		if !slices.Equal(probe.lookedAt, []time.Duration{0}) || cb.fw.QueuedSpell != nil {
			t.Fatalf("the rotation should not look during the 2 s cast, looked at %v", probe.lookedAt)
		}
		cb.runUntil(castEnd)
		if !probe.looked(castEnd) || len(landed) < 2 || landed[0] != 0 || slices.Contains(landed[1:], 0) {
			t.Fatalf("at the 2 s cast end the rotation should pick the instant before the cast lands and look again after, looked at %v, casts landed when the instant was examined: %v", probe.lookedAt, landed)
		}
		if !slices.Equal(cb.slowAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) {
			t.Fatalf("the cast should land at 2 s and the instant go off as it ends, cast at %v, instant at %v", cb.slowAt, cb.nukeAt)
		}
	})

	t.Run("the pick does not see the aura the ending cast puts up", func(t *testing.T) {
		cb := newHardcastTest()
		slowIfNoAura := &gatedAction{APLActionImpl: cb.castOf(cb.slow), ready: func(*Simulation) bool { return !cb.slowAura.IsActive() }}
		cb.start(slowIfNoAura, cb.castOf(cb.nuke))

		cb.runUntil(2 * castEnd)
		if !slices.Equal(cb.slowAt, []time.Duration{castEnd, 2 * castEnd}) {
			t.Fatalf("with the aura down at the 2 s pick the cast should be queued again and land twice, at 2 s and 4 s, landed at %v", cb.slowAt)
		}
		if !slices.Equal(cb.nukeAt, []time.Duration{2 * castEnd}) {
			t.Fatalf("the second cast's pick, with the aura up, should queue the instant for its end, the instant went off at %v", cb.nukeAt)
		}
	})

	t.Run("a spell off the GCD waits at the cast end until the pick has gone off", func(t *testing.T) {
		cb := newHardcastTest()
		cb.offGCD.CD.Set(1500 * time.Millisecond)
		cb.start(cb.castOf(cb.offGCD), cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(castEnd)
		if !slices.Equal(cb.offGCDAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) ||
			!slices.Equal(cb.order, []string{"nuke", "off the GCD"}) {
			t.Fatalf("the spell off the GCD, ready from 1.5 s, should go at the 2 s cast end right after the instant picked there, off the GCD at %v, instant at %v, in the order %v", cb.offGCDAt, cb.nukeAt, cb.order)
		}
	})

	t.Run("a move below the pick waits for it at the cast end", func(t *testing.T) {
		cb := newHardcastTest()
		slowTwice := &gatedAction{APLActionImpl: cb.castOf(cb.slow), ready: func(*Simulation) bool { return len(cb.slowAt) < 2 }}
		move := cb.rot.newActionMove(&proto.APLActionMove{RangeFromTarget: constValue("20")})
		cb.start(slowTwice, move)

		cb.runUntil(castEnd)
		if cb.fw.Moving || cb.fw.Hardcast.ActionID != cb.slow.ActionID || cb.fw.Hardcast.Expires != 2*castEnd {
			t.Fatalf("at the 2 s cast end the slow cast picked there should start before the move below it, moving %t, casting %v until %v", cb.fw.Moving, cb.fw.Hardcast.ActionID, cb.fw.Hardcast.Expires)
		}
		cb.runUntil(2 * castEnd)
		if !slices.Equal(cb.slowAt, []time.Duration{castEnd, 2 * castEnd}) {
			t.Fatalf("both casts should land, at 2 s and 4 s, landed at %v", cb.slowAt)
		}
	})

	t.Run("a major cooldown off the GCD goes at the cast end, ahead of the pick", func(t *testing.T) {
		potion := &testCooldown{}
		cb := newHardcastTest(potion)
		cb.start(cb.autocastAction(), cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(castEnd)
		if !slices.Equal(potion.castAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) ||
			!slices.Equal(cb.order, []string{"cooldown", "nuke"}) {
			t.Fatalf("the potion, ready from 1.5 s, should go at the 2 s cast end ahead of the instant picked there, the potion at %v, instant at %v, in the order %v", potion.castAt, cb.nukeAt, cb.order)
		}
	})

	t.Run("an explosive goes in the cast-end pass right after a picked channel starts", func(t *testing.T) {
		explosive := &testCooldown{explosive: true}
		cb := newHardcastTest(explosive)
		cb.start(cb.autocastAction(), cb.slowAtPull(), cb.channelOf(cb.drain, "false", false))

		cb.runUntil(castEnd)
		if cb.fw.ChanneledDot == nil || !slices.Equal(explosive.castAt, []time.Duration{castEnd}) || !slices.Equal(cb.order, []string{"drain", "cooldown"}) {
			t.Fatalf("the channel picked at the 2 s cast end should start there and the explosive, ready from 1.5 s, go right after it, the explosive at %v, in the order %v", explosive.castAt, cb.order)
		}
	})

	t.Run("a major cooldown on the GCD takes the pick's place at the cast end", func(t *testing.T) {
		onGCD := &testCooldown{gcd: GCDDefault}
		cb := newHardcastTest(onGCD)
		cb.start(cb.autocastAction(), cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(castEnd + GCDDefault + 1)
		if !slices.Equal(onGCD.castAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd + GCDDefault}) {
			t.Fatalf("the cooldown on the GCD should go at the 2 s cast end in place of the picked instant, which goes when its GCD ends at 3.5 s, the cooldown at %v, instant at %v", onGCD.castAt, cb.nukeAt)
		}
	})

	t.Run("a major cooldown's own cast line goes at the cast end, ahead of the pick", func(t *testing.T) {
		potion := &testCooldown{}
		cb := newHardcastTest(potion)
		cb.start(cb.castOf(potion.spell), cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(castEnd)
		if !slices.Equal(potion.castAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) ||
			!slices.Equal(cb.order, []string{"cooldown", "nuke"}) {
			t.Fatalf("the potion's Cast Spell line, ready from 1.5 s, should go at the 2 s cast end ahead of the instant picked there, the potion at %v, instant at %v, in the order %v", potion.castAt, cb.nukeAt, cb.order)
		}
	})

	t.Run("the pick leaves a major cooldown on the GCD to the cast end, where it takes the pick's place", func(t *testing.T) {
		onGCD := &testCooldown{gcd: GCDDefault}
		cb := newHardcastTest(onGCD)
		cb.start(cb.castOf(onGCD.spell), cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(castEnd + GCDDefault + 1)
		if !slices.Equal(onGCD.castAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd + GCDDefault}) {
			t.Fatalf("the cooldown's Cast Spell line should go at the 2 s cast end in place of the picked instant, which goes when its GCD ends at 3.5 s, the cooldown at %v, instant at %v", onGCD.castAt, cb.nukeAt)
		}

		// Ready only until the slow cast's aura goes up as the cast lands.
		onGCD = &testCooldown{gcd: GCDDefault}
		cb = newHardcastTest(onGCD)
		untilLanded := &gatedAction{APLActionImpl: cb.castOf(onGCD.spell), ready: func(*Simulation) bool { return !cb.slowAura.IsActive() }}
		cb.start(untilLanded, cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(castEnd + GCDDefault + 1)
		if len(onGCD.castAt) != 0 || !slices.Equal(cb.nukeAt, []time.Duration{castEnd, castEnd + GCDDefault}) {
			t.Fatalf("the pick should not queue the cooldown, which the landed cast's aura rules out at the 2 s cast end, so the picked instant goes then and again at 3.5 s, the cooldown at %v, instant at %v", onGCD.castAt, cb.nukeAt)
		}
	})

	t.Run("the pick goes off at the cast end without a look at the list unless a major cooldown is ready to go ahead of it", func(t *testing.T) {
		// Never ready; counts the looks at the cast end after the slow cast lands and before the picked instant
		// goes off, and those after it.
		probeAtCastEnd := func(cb *hardcastTest, before, after *int) APLActionImpl {
			return &gatedAction{APLActionImpl: &lookProbe{}, ready: func(sim *Simulation) bool {
				if sim.CurrentTime == castEnd && len(cb.slowAt) == 1 {
					*Ternary(len(cb.nukeAt) == 0, before, after)++
				}
				return false
			}}
		}

		potion := &testCooldown{}
		cb := newHardcastTest(potion)
		potion.spell.CD.Set(time.Minute)
		var before, after int
		cb.start(cb.castOf(potion.spell), cb.autocastAction(), probeAtCastEnd(cb, &before, &after), cb.slowAtPull(), cb.castOf(cb.nuke))
		cb.runUntil(castEnd)
		if !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) || len(potion.castAt) != 0 || before != 0 || after == 0 {
			t.Fatalf("with the potion on its cooldown the instant picked at the 2 s cast end should go off with no look at the list before it and one after, instant at %v, the potion at %v, looks before %d, after %d", cb.nukeAt, potion.castAt, before, after)
		}

		group := func(cb *hardcastTest, potion *testCooldown) APLActionImpl {
			cb.rot.groups = append(cb.rot.groups, &APLGroup{name: "cooldowns", actions: []*APLAction{newAPLActionEntry(nil, cb.castOf(potion.spell))}})
			return cb.rot.newActionGroupReference(&proto.APLActionGroupReference{GroupName: "cooldowns"})
		}
		for _, line := range []struct {
			name   string
			action func(cb *hardcastTest, potion *testCooldown) APLActionImpl
		}{
			{"Cast Spell line", func(cb *hardcastTest, potion *testCooldown) APLActionImpl { return cb.castOf(potion.spell) }},
			{"Autocast Other Cooldowns line", func(cb *hardcastTest, _ *testCooldown) APLActionImpl { return cb.autocastAction() }},
			{"group reference holding its Cast Spell line", group},
		} {
			potion := &testCooldown{}
			cb := newHardcastTest(potion)
			var before, after int
			cb.start(probeAtCastEnd(cb, &before, &after), line.action(cb, potion), cb.slowAtPull(), cb.castOf(cb.nuke))
			cb.runUntil(castEnd)
			if !slices.Equal(cb.order, []string{"cooldown", "nuke"}) || !slices.Equal(potion.castAt, []time.Duration{castEnd}) || before == 0 {
				t.Fatalf("the potion, ready from 1.5 s, should go through its %s at the 2 s cast end ahead of the instant picked there after a look at the list, in the order %v, the potion at %v, looks before the instant %d", line.name, cb.order, potion.castAt, before)
			}
		}
	})

	t.Run("a pick with nothing to queue leaves the cast-end look to act after the cast lands", func(t *testing.T) {
		cb := newHardcastTest()
		probe := &lookProbe{}
		after := &timedStubAction{readyAt: castEnd}
		cb.start(probe, cb.slowAtPull(), after)

		cb.runUntil(castEnd)
		if !slices.Equal(probe.lookedAt, []time.Duration{0, castEnd}) || !slices.Equal(after.ranAt, []time.Duration{castEnd}) {
			t.Fatalf("the rotation should look at the 2 s cast end only, where the action ready from 2 s runs, looked at %v, the action ran at %v", probe.lookedAt, after.ranAt)
		}
	})

	t.Run("after an instant, a channel or a cast shorter than the GCD the look stays where it was, after the outcome", func(t *testing.T) {
		cb := newHardcastTest()
		probe := &lookProbe{}
		cb.start(probe, cb.castOf(cb.nuke))
		cb.runUntil(GCDDefault)
		if !slices.Equal(probe.lookedAt, []time.Duration{0, GCDDefault}) {
			t.Fatalf("after an instant the rotation should look when the GCD ends at 1.5 s, looked at %v", probe.lookedAt)
		}

		cb = newHardcastTest()
		probe = &lookProbe{}
		examinedAtCastEnd := 0
		shortAtPull := &gatedAction{APLActionImpl: cb.castOf(cb.short), ready: func(sim *Simulation) bool { return sim.CurrentTime == 0 }}
		cb.start(probe, shortAtPull, &gatedAction{APLActionImpl: cb.castOf(cb.nuke), ready: func(sim *Simulation) bool {
			if sim.CurrentTime == time.Second {
				examinedAtCastEnd++
			}
			return false
		}})
		cb.runUntil(GCDDefault)
		if !slices.Equal(probe.lookedAt, []time.Duration{0, GCDDefault}) || !slices.Equal(cb.shortAt, []time.Duration{time.Second}) || examinedAtCastEnd != 0 {
			t.Fatalf("after a 1 s cast on the 1.5 s GCD the rotation should look when the GCD ends, not at the cast end, looked at %v, the cast landed at %v, examined the instant at 1 s %d times", probe.lookedAt, cb.shortAt, examinedAtCastEnd)
		}

		var drainTicks []time.Duration
		ct, drain := newSpellShootTest(drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}}))
		probe = &lookProbe{}
		drainAtPull := &gatedAction{APLActionImpl: ct.channelOf(drain, "false", false), ready: func(sim *Simulation) bool { return sim.CurrentTime == 0 }}
		ct.start(probe, drainAtPull)
		ct.runUntil(5 * time.Second)
		if !probe.looked(5*time.Second) || len(drainTicks) != 5 {
			t.Fatalf("after a 5 s channel the rotation should look when it ends, looked at %v, ticks at %v", probe.lookedAt, drainTicks)
		}
	})

	t.Run("a major cooldown goes at the cast end inside a sequence of casts", func(t *testing.T) {
		potion := &testCooldown{}
		cb := newHardcastTest(potion)
		slow := &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: cb.slow.ActionID.ToProto()}}}
		sequence := cb.rot.newActionSequence(&proto.APLActionSequence{Actions: []*proto.APLAction{slow, slow}})
		cb.start(cb.autocastAction(), sequence)

		cb.runUntil(2 * castEnd)
		if !slices.Equal(cb.slowAt, []time.Duration{castEnd, 2 * castEnd}) {
			t.Fatalf("the sequence's two casts should land at 2 s and 4 s, landed at %v", cb.slowAt)
		}
		if !slices.Equal(potion.castAt, []time.Duration{castEnd}) {
			t.Fatalf("the potion, ready from 1.5 s, should go at the first cast's end at 2 s, ahead of the second cast picked there, went at %v", potion.castAt)
		}
	})

	t.Run("a major cooldown goes at the cast end ahead of a Strict Multidot cast queued after the hardcast", func(t *testing.T) {
		potion := &testCooldown{}
		cb := newHardcastTest(potion)
		strictMultidot := cb.rot.newActionStrictMultidot(&proto.APLActionStrictMultidot{SpellId: cb.dotCast.ActionID.ToProto(), MaxDots: 1})
		cb.start(cb.autocastAction(), cb.slowAtPull(), strictMultidot)

		cb.runUntil(castEnd)
		if !slices.Equal(potion.castAt, []time.Duration{castEnd}) || cb.fw.Hardcast.ActionID != cb.dotCast.ActionID || cb.fw.Hardcast.Expires != 2*castEnd {
			t.Fatalf("at the 2 s cast end the potion, ready from 1.5 s, should go and the dot cast picked there start after it, the potion at %v, casting %v until %v", potion.castAt, cb.fw.Hardcast.ActionID, cb.fw.Hardcast.Expires)
		}
		if len(cb.rot.controllingActions) != 0 {
			t.Fatalf("the Strict Multidot should hand the rotation back once its cast went off")
		}
	})

	t.Run("a major cooldown goes at the cast end ahead of a Strict Sequence step queued after the hardcast", func(t *testing.T) {
		potion := &testCooldown{}
		cb := newHardcastTest(potion)
		castOf := func(spell *Spell) *proto.APLAction {
			return &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spell.ActionID.ToProto()}}}
		}
		strictSequence := cb.rot.newActionStrictSequence(&proto.APLActionStrictSequence{Actions: []*proto.APLAction{castOf(cb.slow), castOf(cb.nuke)}})
		cb.start(cb.autocastAction(), strictSequence)

		cb.runUntil(castEnd)
		if !slices.Equal(potion.castAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) ||
			!slices.Equal(cb.order, []string{"cooldown", "nuke"}) {
			t.Fatalf("at the 2 s cast end the potion, ready from 1.5 s, should go ahead of the sequence's instant queued there, the potion at %v, instant at %v, in the order %v", potion.castAt, cb.nukeAt, cb.order)
		}
		if len(cb.rot.controllingActions) != 0 {
			t.Fatalf("the Strict Sequence should hand the rotation back once its last step went off")
		}
	})

	t.Run("a Strict Sequence holds a major cooldown step after a hardcast for the cast-end pass", func(t *testing.T) {
		potion := &testCooldown{}
		cb := newHardcastTest(potion)
		castOf := func(spell *Spell) *proto.APLAction {
			return &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spell.ActionID.ToProto()}}}
		}
		strictSequence := cb.rot.newActionStrictSequence(&proto.APLActionStrictSequence{Actions: []*proto.APLAction{castOf(cb.slow), castOf(potion.spell)}})
		cb.start(&gatedAction{APLActionImpl: strictSequence, ready: func(sim *Simulation) bool { return sim.CurrentTime == 0 }}, cb.castOf(cb.nuke))

		cb.runUntil(castEnd)
		if !slices.Equal(potion.castAt, []time.Duration{castEnd}) || !slices.Equal(cb.order, []string{"cooldown", "nuke"}) {
			t.Fatalf("the sequence's potion step, ready from 1.5 s, should go at the 2 s cast end ahead of the instant below the sequence, the potion at %v, in the order %v", potion.castAt, cb.order)
		}
		if len(cb.rot.controllingActions) != 0 {
			t.Fatalf("the Strict Sequence should hand the rotation back once its last step went off")
		}
	})

	t.Run("a prepull cast ending at the pull gets its pick at 0 before it lands", func(t *testing.T) {
		// The pick at 0 goes before the cast's outcome, as at any cast end; a prepull action at 0 completes
		// the cast first, so it runs after the pick.
		cb := newHardcastTest()
		var landed []int
		cb.sim.CurrentTime = -castEnd
		cb.castOf(cb.slow).Execute(cb.sim)
		cb.start(cb.nukeNotingLandings(0, &landed))

		cb.runUntil(0)
		if len(landed) < 2 || landed[0] != 0 || slices.Contains(landed[1:], 0) || !slices.Equal(cb.slowAt, []time.Duration{0}) || !slices.Equal(cb.nukeAt, []time.Duration{0}) {
			t.Fatalf("the cast from -2 s should land at 0 with the instant picked before it lands and going off at 0, casts landed when the instant was examined: %v, cast at %v, instant at %v", landed, cb.slowAt, cb.nukeAt)
		}
	})

	t.Run("a pet's hardcast keeps the look at the cast end, after the cast lands", func(t *testing.T) {
		cb := newHardcastTest()
		probe := &lookProbe{}
		var landed []int
		nukeAfterPull := &gatedAction{APLActionImpl: cb.nukeNotingLandings(castEnd, &landed), ready: func(sim *Simulation) bool { return sim.CurrentTime > 0 }}
		cb.start(probe, nukeAfterPull)
		cb.runUntil(0)
		cb.fw.Type = PetUnit
		cb.castOf(cb.slow).Execute(cb.sim)

		cb.runUntil(castEnd)
		if !probe.looked(castEnd) || len(landed) == 0 || slices.Contains(landed, 0) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) {
			t.Fatalf("a pet should look once its 2 s cast has landed, looked at %v, casts landed when the instant was examined: %v, the instant went off at %v", probe.lookedAt, landed, cb.nukeAt)
		}
	})

	t.Run("a pushback drops the cast queued at the old end for a fresh pick at the new end", func(t *testing.T) {
		cb := newHardcastTest()
		var landed []int
		cb.start(cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(castEnd - 5*time.Millisecond)
		hc := &cb.fw.Hardcast
		pushback := hc.pushBack(cb.sim.CurrentTime)
		cb.fw.newHardcastAction(cb.sim)
		newEnd := castEnd + pushback
		cb.rot.priorityList[1] = newAPLActionEntry(nil, cb.nukeNotingLandings(newEnd, &landed))

		cb.runUntil(castEnd)
		if cb.fw.QueuedSpell == nil || cb.fw.QueuedSpell.spell != cb.nuke || len(cb.nukeAt) != 0 || cb.fw.rotationCast.spell != cb.slow {
			t.Fatalf("at the old 2 s end the rotation should queue the instant under the running cast, pushed back to %v, instant at %v", newEnd, cb.nukeAt)
		}
		queuedAtOldEnd := cb.fw.QueuedSpell.queueAction
		cb.runUntil(newEnd)
		if len(landed) < 2 || landed[0] != 0 || slices.Contains(landed[1:], 0) || !queuedAtOldEnd.cancelled {
			t.Fatalf("at the new end %v the rotation should drop the instant queued at 2 s and pick again before the cast lands, casts landed when the instant was examined: %v", newEnd, landed)
		}
		if !slices.Equal(cb.slowAt, []time.Duration{newEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{newEnd}) {
			t.Fatalf("the cast should land at the new end %v and the instant picked there go then, cast at %v, instant at %v", newEnd, cb.slowAt, cb.nukeAt)
		}
	})
}

// Never ready; each time the rotation looks at it at the end of the hardcast of unit, before the hardcast
// lands, notes what read says.
func readAtPick(unit *Unit, read func(sim *Simulation) bool, reads *[]bool) APLActionImpl {
	return &gatedAction{APLActionImpl: &timedStubAction{readyAt: time.Minute}, ready: func(sim *Simulation) bool {
		if unit.Hardcast.Expires == sim.CurrentTime {
			*reads = append(*reads, read(sim))
		}
		return false
	}}
}

func TestRotationReadsOwnCastAsLanded(t *testing.T) {
	const castEnd = 2 * time.Second
	const beforeEnd = 10 * time.Millisecond
	lateInCast := castEnd - beforeEnd

	t.Run("the cast's own dot reads as landed at the cast end", func(t *testing.T) {
		cb := newHardcastTest()
		spellID := cb.dotCast.ActionID.ToProto()
		isActive := cb.rot.newValueDotIsActive(&proto.APLValueDotIsActive{SpellId: spellID}, nil)
		remaining := cb.rot.newValueDotRemainingTime(&proto.APLValueDotRemainingTime{SpellId: spellID}, nil)
		multidot := cb.rot.newActionMultidot(&proto.APLActionMultidot{SpellId: spellID, MaxDots: 1})
		onAllTargets := cb.rot.newValueDotIsActiveOnAllTargets(&proto.APLValueDotIsActiveOnAllTargets{SpellId: spellID}, nil)
		lowestRemaining := cb.rot.newValueDotLowestRemainingTime(&proto.APLValueDotLowestRemainingTime{SpellId: spellID}, nil)
		tickFrequency := cb.rot.newValueDotTickFrequency(&proto.APLValueDotTickFrequency{SpellId: spellID}, nil)
		timeToNextTick := cb.rot.newValueDotTimeToNextTick(&proto.APLValueDotTimeToNextTick{SpellId: spellID}, nil)
		dot := cb.dotCast.Dot(cb.fw.CurrentTarget)
		var atPick []bool
		cb.start(readAtPick(&cb.fw.Unit, func(sim *Simulation) bool { return !dot.IsActive() && isActive.GetBool(sim) }, &atPick))
		cb.runUntil(0)
		cb.castOf(cb.dotCast).Execute(cb.sim)

		for _, at := range []time.Duration{time.Second, lateInCast} {
			cb.runUntil(at)
			if dot.IsActive() {
				t.Fatalf("the dot itself should only go up when the cast lands at 2 s, up at %v", at)
			}
			if !isActive.GetBool(cb.sim) || remaining.GetDuration(cb.sim) != castEnd+5*time.Second-at {
				t.Fatalf("at %v the rotation should read the dot as up until 7 s, 5 s from the cast end, read %t with %v left", at, isActive.GetBool(cb.sim), remaining.GetDuration(cb.sim))
			}
			if !onAllTargets.GetBool(cb.sim) || lowestRemaining.GetDuration(cb.sim) != castEnd+5*time.Second-at {
				t.Fatalf("at %v the dot should read as up on the one target until 7 s, read %t with %v lowest left", at, onAllTargets.GetBool(cb.sim), lowestRemaining.GetDuration(cb.sim))
			}
			if tickFrequency.GetDuration(cb.sim) != time.Second || timeToNextTick.GetDuration(cb.sim) != castEnd+time.Second-at {
				t.Fatalf("at %v the dot should read as ticking every 1 s from the cast end, read %v with the next tick in %v", at, tickFrequency.GetDuration(cb.sim), timeToNextTick.GetDuration(cb.sim))
			}
		}
		if multidot.IsReady(cb.sim) || !cb.castOf(cb.dotCast).IsReady(cb.sim) {
			t.Fatalf("late in the 2 s cast the dot cast could be queued again, but Multidot should read its dot as landed and not cast it")
		}
		cb.runUntil(castEnd)
		if !slices.Equal(atPick, []bool{true}) {
			t.Fatalf("the pick at the 2 s cast end should read the dot as up before it goes up, read %v", atPick)
		}
		cb.runUntil(castEnd + time.Second)
		if !dot.IsActive() || remaining.GetDuration(cb.sim) != 4*time.Second {
			t.Fatalf("after the cast lands the rotation should read the dot itself, 4 s left at 3 s, read %v", remaining.GetDuration(cb.sim))
		}
	})

	t.Run("the cast's own buff reads as landed, other auras as they are", func(t *testing.T) {
		cb := newHardcastTest()
		auraValues := func(id ActionID) (APLValue, APLValue, APLValue) {
			return cb.rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{AuraId: id.ToProto()}, nil),
				cb.rot.newValueAuraRemainingTime(&proto.APLValueAuraRemainingTime{AuraId: id.ToProto()}, nil),
				cb.rot.newValueAuraNumStacks(&proto.APLValueAuraNumStacks{AuraId: id.ToProto()}, nil)
		}
		buffActive, buffRemaining, buffStacks := auraValues(cb.buff.ActionID)
		buffInactive := cb.rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{AuraId: cb.buff.ActionID.ToProto()}, nil)
		otherActive, _, _ := auraValues(cb.slowAura.ActionID)
		cb.start(&timedStubAction{readyAt: time.Minute})
		cb.runUntil(0)
		cb.buffCast.Cast(cb.sim, &cb.fw.Unit)

		cb.runUntil(lateInCast)
		if cb.buff.IsActive() || !buffActive.GetBool(cb.sim) || buffInactive.GetBool(cb.sim) || buffRemaining.GetDuration(cb.sim) != castEnd+10*time.Second-lateInCast || buffStacks.GetInt(cb.sim) != 1 {
			t.Fatalf("late in the 2 s cast the buff should read as up with one stack until 12 s, read %t (inactive %t), %v left, %d stacks", buffActive.GetBool(cb.sim), buffInactive.GetBool(cb.sim), buffRemaining.GetDuration(cb.sim), buffStacks.GetInt(cb.sim))
		}
		cb.runUntil(castEnd)
		cb.runUntil(2 * castEnd)
		cb.buffCast.Cast(cb.sim, &cb.fw.Unit)
		cb.runUntil(3*castEnd - beforeEnd)
		if cb.buff.GetStacks() != 1 || buffStacks.GetInt(cb.sim) != 2 {
			t.Fatalf("during the second cast the buff's one stack should read as two, read %d", buffStacks.GetInt(cb.sim))
		}

		cb.runUntil(4 * castEnd)
		cb.castOf(cb.slow).Execute(cb.sim)
		cb.runUntil(5*castEnd - beforeEnd)
		if otherActive.GetBool(cb.sim) {
			t.Fatalf("the aura of another id that the slow cast puts up should read as down until it lands")
		}
	})

	t.Run("the cooldowns the cast starts as it lands read as running from the cast end", func(t *testing.T) {
		cb := newHardcastTest()
		isReady := func(spell *Spell) APLValue {
			return cb.rot.newValueSpellIsReady(&proto.APLValueSpellIsReady{SpellId: spell.ActionID.ToProto()}, nil)
		}
		timeToReady := func(spell *Spell) APLValue {
			return cb.rot.newValueSpellTimeToReady(&proto.APLValueSpellTimeToReady{SpellId: spell.ActionID.ToProto()}, nil)
		}
		blastReady, blastTimeToReady, shockTimeToReady := isReady(cb.blast), timeToReady(cb.blast), timeToReady(cb.shock)
		offGCDReady, offGCDTimeToReady := isReady(cb.offGCD), timeToReady(cb.offGCD)
		cb.start(cb.castOf(cb.blast), cb.castOf(cb.nuke))

		cb.runUntil(time.Second)
		if !cb.blast.IsReady(cb.sim) {
			t.Fatalf("the blast's own cooldown should only start when it lands at 2 s")
		}
		if blastReady.GetBool(cb.sim) || blastTimeToReady.GetDuration(cb.sim) != castEnd+8*time.Second-time.Second {
			t.Fatalf("at 1 s the rotation should read the blast as ready at 10 s, 8 s from its cast end, read ready %t in %v", blastReady.GetBool(cb.sim), blastTimeToReady.GetDuration(cb.sim))
		}
		if shockTimeToReady.GetDuration(cb.sim) != castEnd+4*time.Second-time.Second {
			t.Fatalf("at 1 s the shock should read the 4 s cooldown it shares with the blast as running from 2 s, read ready in %v", shockTimeToReady.GetDuration(cb.sim))
		}
		if !offGCDReady.GetBool(cb.sim) || offGCDTimeToReady.GetDuration(cb.sim) != 0 {
			t.Fatalf("a spell on a cooldown of its own should read as it is, ready now, read ready %t in %v", offGCDReady.GetBool(cb.sim), offGCDTimeToReady.GetDuration(cb.sim))
		}

		cb.runUntil(castEnd)
		if !slices.Equal(cb.blastAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) {
			t.Fatalf("the blast should land at 2 s and the instant below it, picked at the cast end as the blast's cooldown is about to start, go off as it ends, blast at %v, instant at %v", cb.blastAt, cb.nukeAt)
		}

		cb.runUntil(castEnd + time.Second)
		if cb.blast.IsReady(cb.sim) || blastReady.GetBool(cb.sim) || blastTimeToReady.GetDuration(cb.sim) != 7*time.Second {
			t.Fatalf("after the blast lands its cooldown should run until 10 s, 7 s left at 3 s, read ready %t in %v", blastReady.GetBool(cb.sim), blastTimeToReady.GetDuration(cb.sim))
		}
	})

	t.Run("Spell Is Casting is true for the hardcast only", func(t *testing.T) {
		cb := newHardcastTest()
		isCasting := func(spell *Spell) APLValue {
			return cb.rot.newValueSpellIsCasting(&proto.APLValueSpellIsCasting{SpellId: spell.ActionID.ToProto()}, nil)
		}
		slowCasting, nukeCasting := isCasting(cb.slow), isCasting(cb.nuke)
		var atPick []bool
		cb.start(readAtPick(&cb.fw.Unit, func(sim *Simulation) bool { return len(cb.slowAt) == 0 && slowCasting.GetBool(sim) }, &atPick))
		cb.runUntil(0)
		cb.castOf(cb.slow).Execute(cb.sim)

		cb.runUntil(lateInCast)
		if !slowCasting.GetBool(cb.sim) || nukeCasting.GetBool(cb.sim) {
			t.Fatalf("late in the 2 s cast only the slow cast should be casting")
		}
		cb.runUntil(castEnd)
		if !slices.Equal(atPick, []bool{true}) {
			t.Fatalf("the pick at the 2 s cast end should read the slow cast as casting before it lands, read %v", atPick)
		}
		if slowCasting.GetBool(cb.sim) {
			t.Fatalf("once the cast lands at 2 s it should no longer be casting")
		}

		var drainTicks []time.Duration
		ct, drain := newSpellShootTest(drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}}))
		drainCasting := ct.rot.newValueSpellIsCasting(&proto.APLValueSpellIsCasting{SpellId: drain.ActionID.ToProto()}, nil)
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)
		ct.channelOf(drain, "false", false).Execute(ct.sim)
		ct.runUntil(time.Second + time.Millisecond)
		if ct.fw.ChanneledDot == nil || drainCasting.GetBool(ct.sim) {
			t.Fatalf("a running channel should not read as casting")
		}
	})

	t.Run("current mana reads what is left once the running cast pays", func(t *testing.T) {
		var paid *Spell
		ct := newManaTickTest(func(char *Character) *Spell {
			paid = char.RegisterSpell(paidSpellConfig(910170, 2*time.Second))
			return paid
		})
		mana := ct.rot.newValueCurrentMana(&proto.APLValueCurrentMana{}, nil)
		manaPercent := ct.rot.newValueCurrentManaPercent(&proto.APLValueCurrentManaPercent{}, nil)
		var atPick []bool
		ct.start(readAtPick(&ct.fw.Unit, func(sim *Simulation) bool { return mana.GetFloat(sim) == ct.fw.CurrentMana()-100 }, &atPick))
		ct.runUntil(0)
		paid.Cast(ct.sim, ct.fw.CurrentTarget)

		ct.runUntil(lateInCast)
		if want := ct.fw.CurrentMana() - 100; mana.GetFloat(ct.sim) != want || manaPercent.GetFloat(ct.sim) != want/ct.fw.MaxMana() {
			t.Fatalf("during the 2 s cast mana should read %.1f, what is left once it pays 100, read %.1f (%.4f)", want, mana.GetFloat(ct.sim), manaPercent.GetFloat(ct.sim))
		}
		ct.runUntil(castEnd)
		if !slices.Equal(atPick, []bool{true}) {
			t.Fatalf("the pick at the 2 s cast end should read the mana left once the cast pays, before it pays, read %v", atPick)
		}
		if mana.GetFloat(ct.sim) != ct.fw.CurrentMana() {
			t.Fatalf("once the cast pays at 2 s mana should read as it is, %.1f, read %.1f", ct.fw.CurrentMana(), mana.GetFloat(ct.sim))
		}
	})
}

// A rotation with a move keeps looking through the GCD once an event wakes it: after an event at 0.5 s under the
// GCD of the instant cast at the pull (as an Auto Shot, a swing or a proc would), a walk whose condition turns true
// at 0.9 s starts then rather than at the GCD's end.
func TestMoveConditionTurningTrueDuringTheGcdStartsTheWalk(t *testing.T) {
	cb := newHardcastTest()
	fromNinth := cb.rot.newAPLValue(&proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
		Op:  proto.APLValueCompare_OpGe,
		Lhs: &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}},
		Rhs: constValue("0.9s"),
	}}})
	nukeAtPull := &gatedAction{APLActionImpl: cb.castOf(cb.nuke), ready: func(sim *Simulation) bool { return sim.CurrentTime == 0 }}
	cb.rot.priorityList = nil
	for _, action := range []*APLAction{
		newAPLActionEntry(fromNinth, cb.rot.newActionMove(&proto.APLActionMove{RangeFromTarget: constValue("20")})),
		newAPLActionEntry(nil, nukeAtPull),
	} {
		action.Finalize(cb.rot)
		cb.rot.priorityList = append(cb.rot.priorityList, action)
	}
	cb.sim.PrePull()
	cb.sim.Step()
	cb.fw.AutoAttacks.CancelAutoSwing(cb.sim)
	cb.sim.AddPendingAction(NewDelayedAction(DelayedActionOptions{
		DoAt:     500 * time.Millisecond,
		OnAction: func(sim *Simulation) { cb.fw.ReactToEvent(sim, false, true) },
	}))

	cb.runUntil(950 * time.Millisecond)
	if cb.fw.GCD.IsReady(cb.sim) {
		t.Fatalf("the instant's GCD should still run at 0.95 s, the instant went off at %v", cb.nukeAt)
	}
	if !cb.fw.Moving {
		t.Fatalf("at 0.95 s the walk whose condition turned true at 0.9 s should have started, under the GCD, at %.1f yd", cb.fw.DistanceFromTarget)
	}
}
