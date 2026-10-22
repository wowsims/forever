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
type castBarTest struct {
	*channelTest
	slow, nuke, offGCD, short         *Spell
	slowAt, nukeAt, offGCDAt, shortAt []time.Duration
	slowAura                          *Aura
	dotCast, buffCast                 *Spell
	buff                              *Aura
	drain                             *Spell
	drainTicks                        []time.Duration
	order                             []string
}

// The config, noting name in the test's order as its effects apply.
func (cb *castBarTest) ordered(name string, config SpellConfig) SpellConfig {
	applyEffects := config.ApplyEffects
	config.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
		cb.order = append(cb.order, name)
		applyEffects(sim, target, spell)
	}
	return config
}

func newCastBarTest(cooldowns ...*testCooldown) *castBarTest {
	cb := &castBarTest{}
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
		dotCast := fakeWarlockDot(910160, "Cast Bar Dot", ClassFlags{}, SpellFlagAPL)
		dotCast.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: 2 * time.Second}}
		cb.dotCast = char.RegisterSpell(dotCast)
		cb.buff = char.RegisterAura(Aura{Label: "Cast Bar Buff", ActionID: ActionID{SpellID: 910161}, Duration: 10 * time.Second, MaxStacks: 3})
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
	for _, spell := range []*Spell{cb.slow, cb.nuke, cb.offGCD, cb.short, cb.dotCast, cb.buffCast, cb.drain} {
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
func (cb *castBarTest) slowAtPull() APLActionImpl {
	return &gatedAction{APLActionImpl: cb.castOf(cb.slow), ready: func(sim *Simulation) bool { return sim.CurrentTime == 0 }}
}

func TestRotationQueuesBeforeCastBarEnds(t *testing.T) {
	const castEnd = 2 * time.Second
	earlyLook := castEnd - SpellBatchWindow

	t.Run("the rotation looks one batch before the cast bar ends and its pick goes off at the end", func(t *testing.T) {
		cb := newCastBarTest()
		probe := &lookProbe{}
		cb.start(probe, cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(earlyLook - 1)
		if probe.looked(earlyLook) || cb.fw.QueuedSpell != nil {
			t.Fatalf("the rotation should not look before 1.99 s, looked at %v", probe.lookedAt)
		}
		cb.runUntil(earlyLook)
		if !probe.looked(earlyLook) || cb.fw.QueuedSpell == nil || cb.fw.QueuedSpell.spell != cb.nuke || len(cb.slowAt) != 0 {
			t.Fatalf("at 1.99 s, before the 2 s cast lands, the rotation should look and queue the instant, looked at %v, cast landed at %v", probe.lookedAt, cb.slowAt)
		}
		cb.runUntil(castEnd)
		if !slices.Equal(cb.slowAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) {
			t.Fatalf("the cast should land at 2 s and the queued instant go off as it ends, cast at %v, instant at %v", cb.slowAt, cb.nukeAt)
		}
	})

	t.Run("the pick does not see the aura the ending cast puts up", func(t *testing.T) {
		cb := newCastBarTest()
		slowIfNoAura := &gatedAction{APLActionImpl: cb.castOf(cb.slow), ready: func(*Simulation) bool { return !cb.slowAura.IsActive() }}
		cb.start(slowIfNoAura, cb.castOf(cb.nuke))

		cb.runUntil(2 * castEnd)
		if !slices.Equal(cb.slowAt, []time.Duration{castEnd, 2 * castEnd}) {
			t.Fatalf("with the aura down at the 1.99 s look the cast should be queued again and land twice, at 2 s and 4 s, landed at %v", cb.slowAt)
		}
		if !slices.Equal(cb.nukeAt, []time.Duration{2 * castEnd}) {
			t.Fatalf("the second cast's look, with the aura up, should queue the instant for its end, the instant went off at %v", cb.nukeAt)
		}
	})

	t.Run("a spell off the GCD waits at the cast end until the pick has gone off", func(t *testing.T) {
		cb := newCastBarTest()
		cb.offGCD.CD.Set(1500 * time.Millisecond)
		cb.start(cb.castOf(cb.offGCD), cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(castEnd)
		if !slices.Equal(cb.offGCDAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) ||
			!slices.Equal(cb.order, []string{"nuke", "off the GCD"}) {
			t.Fatalf("the spell off the GCD, ready from 1.5 s, should go at the 2 s cast end right after the instant queued at 1.99 s, off the GCD at %v, instant at %v, in the order %v", cb.offGCDAt, cb.nukeAt, cb.order)
		}
	})

	t.Run("a move below the pick waits for it at the cast end", func(t *testing.T) {
		cb := newCastBarTest()
		slowTwice := &gatedAction{APLActionImpl: cb.castOf(cb.slow), ready: func(*Simulation) bool { return len(cb.slowAt) < 2 }}
		move := cb.rot.newActionMove(&proto.APLActionMove{RangeFromTarget: constValue("20")})
		cb.start(slowTwice, move)

		cb.runUntil(castEnd)
		if cb.fw.Moving || cb.fw.Hardcast.ActionID != cb.slow.ActionID || cb.fw.Hardcast.Expires != 2*castEnd {
			t.Fatalf("at the 2 s cast end the slow cast queued at 1.99 s should start before the move below it, moving %t, casting %v until %v", cb.fw.Moving, cb.fw.Hardcast.ActionID, cb.fw.Hardcast.Expires)
		}
		cb.runUntil(2 * castEnd)
		if !slices.Equal(cb.slowAt, []time.Duration{castEnd, 2 * castEnd}) {
			t.Fatalf("both casts should land, at 2 s and 4 s, landed at %v", cb.slowAt)
		}
	})

	t.Run("a major cooldown off the GCD goes at the cast end, ahead of the pick", func(t *testing.T) {
		potion := &testCooldown{}
		cb := newCastBarTest(potion)
		cb.start(cb.autocastAction(), cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(castEnd)
		if !slices.Equal(potion.castAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) ||
			!slices.Equal(cb.order, []string{"cooldown", "nuke"}) {
			t.Fatalf("the potion, ready from 1.5 s, should go at the 2 s cast end ahead of the instant queued at 1.99 s, the potion at %v, instant at %v, in the order %v", potion.castAt, cb.nukeAt, cb.order)
		}
	})

	t.Run("an explosive goes in the cast-end pass right after a queued channel starts", func(t *testing.T) {
		explosive := &testCooldown{explosive: true}
		cb := newCastBarTest(explosive)
		cb.start(cb.autocastAction(), cb.slowAtPull(), cb.channelOf(cb.drain, "false", false))

		cb.runUntil(castEnd)
		if cb.fw.ChanneledDot == nil || !slices.Equal(explosive.castAt, []time.Duration{castEnd}) || !slices.Equal(cb.order, []string{"drain", "cooldown"}) {
			t.Fatalf("the channel queued at 1.99 s should start at the 2 s cast end and the explosive, ready from 1.5 s, go right after it, the explosive at %v, in the order %v", explosive.castAt, cb.order)
		}
	})

	t.Run("a major cooldown on the GCD takes the pick's place at the cast end", func(t *testing.T) {
		onGCD := &testCooldown{gcd: GCDDefault}
		cb := newCastBarTest(onGCD)
		cb.start(cb.autocastAction(), cb.slowAtPull(), cb.castOf(cb.nuke))

		cb.runUntil(castEnd + GCDDefault + 1)
		if !slices.Equal(onGCD.castAt, []time.Duration{castEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd + GCDDefault}) {
			t.Fatalf("the cooldown on the GCD should go at the 2 s cast end in place of the queued instant, which goes when its GCD ends at 3.5 s, the cooldown at %v, instant at %v", onGCD.castAt, cb.nukeAt)
		}
	})

	t.Run("a look with nothing to queue looks again at the cast end", func(t *testing.T) {
		cb := newCastBarTest()
		probe := &lookProbe{}
		after := &timedStubAction{readyAt: castEnd}
		cb.start(probe, cb.slowAtPull(), after)

		cb.runUntil(castEnd)
		if !probe.looked(earlyLook) || !probe.looked(castEnd) || !slices.Equal(after.ranAt, []time.Duration{castEnd}) {
			t.Fatalf("the rotation should look at 1.99 s and again at the 2 s cast end, where the action ready from 2 s runs, looked at %v, the action ran at %v", probe.lookedAt, after.ranAt)
		}
	})

	t.Run("instants, channels and casts shorter than the GCD keep the look at their end", func(t *testing.T) {
		lookedEarly := func(probe *lookProbe, end time.Duration) bool {
			return slices.ContainsFunc(probe.lookedAt, func(at time.Duration) bool { return at > end-SpellBatchWindow-1 && at < end })
		}

		cb := newCastBarTest()
		probe := &lookProbe{}
		cb.start(probe, cb.castOf(cb.nuke))
		cb.runUntil(GCDDefault)
		if lookedEarly(probe, GCDDefault) || !probe.looked(GCDDefault) {
			t.Fatalf("after an instant the rotation should look when the GCD ends at 1.5 s, looked at %v", probe.lookedAt)
		}

		cb = newCastBarTest()
		probe = &lookProbe{}
		shortAtPull := &gatedAction{APLActionImpl: cb.castOf(cb.short), ready: func(sim *Simulation) bool { return sim.CurrentTime == 0 }}
		cb.start(probe, shortAtPull)
		cb.runUntil(GCDDefault)
		if lookedEarly(probe, time.Second) || !probe.looked(GCDDefault) || !slices.Equal(cb.shortAt, []time.Duration{time.Second}) {
			t.Fatalf("after a 1 s cast on the 1.5 s GCD the rotation should look when the GCD ends, looked at %v, the cast landed at %v", probe.lookedAt, cb.shortAt)
		}

		var drainTicks []time.Duration
		ct, drain := newSpellShootTest(drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}}))
		probe = &lookProbe{}
		drainAtPull := &gatedAction{APLActionImpl: ct.channelOf(drain, "false", false), ready: func(sim *Simulation) bool { return sim.CurrentTime == 0 }}
		ct.start(probe, drainAtPull)
		ct.runUntil(5 * time.Second)
		if lookedEarly(probe, 5*time.Second) || !probe.looked(5*time.Second) || len(drainTicks) != 5 {
			t.Fatalf("after a 5 s channel the rotation should look when it ends, looked at %v, ticks at %v", probe.lookedAt, drainTicks)
		}
	})

	t.Run("a major cooldown goes at the cast end inside a sequence of casts", func(t *testing.T) {
		potion := &testCooldown{}
		cb := newCastBarTest(potion)
		slow := &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: cb.slow.ActionID.ToProto()}}}
		sequence := cb.rot.newActionSequence(&proto.APLActionSequence{Actions: []*proto.APLAction{slow, slow}})
		cb.start(cb.autocastAction(), sequence)

		cb.runUntil(2 * castEnd)
		if !slices.Equal(cb.slowAt, []time.Duration{castEnd, 2 * castEnd}) {
			t.Fatalf("the sequence's two casts should land at 2 s and 4 s, landed at %v", cb.slowAt)
		}
		if !slices.Equal(potion.castAt, []time.Duration{castEnd}) {
			t.Fatalf("the potion, ready from 1.5 s, should go at the first cast's end at 2 s, ahead of the second cast queued at 1.99 s, went at %v", potion.castAt)
		}
	})

	t.Run("a major cooldown goes at the cast end ahead of a Strict Multidot cast queued behind the bar", func(t *testing.T) {
		potion := &testCooldown{}
		cb := newCastBarTest(potion)
		strictMultidot := cb.rot.newActionStrictMultidot(&proto.APLActionStrictMultidot{SpellId: cb.dotCast.ActionID.ToProto(), MaxDots: 1})
		cb.start(cb.autocastAction(), cb.slowAtPull(), strictMultidot)

		cb.runUntil(castEnd)
		if !slices.Equal(potion.castAt, []time.Duration{castEnd}) || cb.fw.Hardcast.ActionID != cb.dotCast.ActionID || cb.fw.Hardcast.Expires != 2*castEnd {
			t.Fatalf("at the 2 s cast end the potion, ready from 1.5 s, should go and the dot cast queued at 1.99 s start after it, the potion at %v, casting %v until %v", potion.castAt, cb.fw.Hardcast.ActionID, cb.fw.Hardcast.Expires)
		}
		if len(cb.rot.controllingActions) != 0 {
			t.Fatalf("the Strict Multidot should hand the rotation back once its cast went off")
		}
	})

	t.Run("a pet's cast bar keeps the look at its end", func(t *testing.T) {
		cb := newCastBarTest()
		probe := &lookProbe{}
		nukeAfterPull := &gatedAction{APLActionImpl: cb.castOf(cb.nuke), ready: func(sim *Simulation) bool { return sim.CurrentTime > 0 }}
		cb.start(probe, nukeAfterPull)
		cb.runUntil(0)
		cb.fw.Type = PetUnit
		cb.castOf(cb.slow).Execute(cb.sim)

		cb.runUntil(castEnd)
		if probe.looked(earlyLook) || !probe.looked(castEnd) || !slices.Equal(cb.nukeAt, []time.Duration{castEnd}) {
			t.Fatalf("a pet should look when its 2 s cast ends, not at 1.99 s, looked at %v, the instant went off at %v", probe.lookedAt, cb.nukeAt)
		}
	})

	t.Run("a pushback after the early look drops the queued cast and looks again before the new end", func(t *testing.T) {
		cb := newCastBarTest()
		// Nothing would queue the instant again at the old end.
		nukeOffOldEnd := &gatedAction{APLActionImpl: cb.castOf(cb.nuke), ready: func(sim *Simulation) bool { return sim.CurrentTime != castEnd }}
		cb.start(cb.slowAtPull(), nukeOffOldEnd)

		cb.runUntil(earlyLook + 5*time.Millisecond)
		if cb.fw.QueuedSpell == nil || cb.fw.QueuedSpell.spell != cb.nuke {
			t.Fatalf("the instant should be queued at 1.99 s")
		}
		hc := &cb.fw.Hardcast
		pushback := hc.pushBack(cb.sim.CurrentTime)
		cb.fw.newHardcastAction(cb.sim)
		cb.fw.lookAgainAfterPushback(cb.sim)
		newEnd := castEnd + pushback

		cb.runUntil(newEnd - SpellBatchWindow - 1)
		if len(cb.nukeAt) != 0 || cb.fw.rotationCast.spell != cb.slow {
			t.Fatalf("pushed back to %v, the slow cast should still run with nothing cast at the old end, instant at %v", newEnd, cb.nukeAt)
		}
		cb.runUntil(newEnd)
		if !slices.Equal(cb.slowAt, []time.Duration{newEnd}) || !slices.Equal(cb.nukeAt, []time.Duration{newEnd}) {
			t.Fatalf("the cast should land at the new end %v and the instant, queued again before it, go then, cast at %v, instant at %v", newEnd, cb.slowAt, cb.nukeAt)
		}
	})
}

func TestRotationReadsOwnCastAsLanded(t *testing.T) {
	const castEnd = 2 * time.Second
	earlyLook := castEnd - SpellBatchWindow

	t.Run("the cast's own dot reads as landed at the cast end", func(t *testing.T) {
		cb := newCastBarTest()
		spellID := cb.dotCast.ActionID.ToProto()
		isActive := cb.rot.newValueDotIsActive(&proto.APLValueDotIsActive{SpellId: spellID}, nil)
		remaining := cb.rot.newValueDotRemainingTime(&proto.APLValueDotRemainingTime{SpellId: spellID}, nil)
		multidot := cb.rot.newActionMultidot(&proto.APLActionMultidot{SpellId: spellID, MaxDots: 1})
		onAllTargets := cb.rot.newValueDotIsActiveOnAllTargets(&proto.APLValueDotIsActiveOnAllTargets{SpellId: spellID}, nil)
		lowestRemaining := cb.rot.newValueDotLowestRemainingTime(&proto.APLValueDotLowestRemainingTime{SpellId: spellID}, nil)
		tickFrequency := cb.rot.newValueDotTickFrequency(&proto.APLValueDotTickFrequency{SpellId: spellID}, nil)
		timeToNextTick := cb.rot.newValueDotTimeToNextTick(&proto.APLValueDotTimeToNextTick{SpellId: spellID}, nil)
		dot := cb.dotCast.Dot(cb.fw.CurrentTarget)
		cb.start(&timedStubAction{readyAt: time.Minute})
		cb.runUntil(0)
		cb.castOf(cb.dotCast).Execute(cb.sim)

		for _, at := range []time.Duration{time.Second, earlyLook} {
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
			t.Fatalf("at 1.99 s the cast could be queued again, but Multidot should read its dot as landed and not cast it")
		}
		cb.runUntil(castEnd + time.Second)
		if !dot.IsActive() || remaining.GetDuration(cb.sim) != 4*time.Second {
			t.Fatalf("after the cast lands the rotation should read the dot itself, 4 s left at 3 s, read %v", remaining.GetDuration(cb.sim))
		}
	})

	t.Run("the cast's own buff reads as landed, other auras as they are", func(t *testing.T) {
		cb := newCastBarTest()
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

		cb.runUntil(earlyLook)
		if cb.buff.IsActive() || !buffActive.GetBool(cb.sim) || buffInactive.GetBool(cb.sim) || buffRemaining.GetDuration(cb.sim) != castEnd+10*time.Second-earlyLook || buffStacks.GetInt(cb.sim) != 1 {
			t.Fatalf("at 1.99 s the buff should read as up with one stack until 12 s, read %t (inactive %t), %v left, %d stacks", buffActive.GetBool(cb.sim), buffInactive.GetBool(cb.sim), buffRemaining.GetDuration(cb.sim), buffStacks.GetInt(cb.sim))
		}
		cb.runUntil(castEnd)
		cb.runUntil(2 * castEnd)
		cb.buffCast.Cast(cb.sim, &cb.fw.Unit)
		cb.runUntil(3*castEnd - SpellBatchWindow)
		if cb.buff.GetStacks() != 1 || buffStacks.GetInt(cb.sim) != 2 {
			t.Fatalf("during the second cast the buff's one stack should read as two, read %d", buffStacks.GetInt(cb.sim))
		}

		cb.runUntil(4 * castEnd)
		cb.castOf(cb.slow).Execute(cb.sim)
		cb.runUntil(5*castEnd - SpellBatchWindow)
		if otherActive.GetBool(cb.sim) {
			t.Fatalf("the aura of another id that the slow cast puts up should read as down until it lands")
		}
	})

	t.Run("Spell Is Casting is true for the cast bar only", func(t *testing.T) {
		cb := newCastBarTest()
		isCasting := func(spell *Spell) APLValue {
			return cb.rot.newValueSpellIsCasting(&proto.APLValueSpellIsCasting{SpellId: spell.ActionID.ToProto()}, nil)
		}
		slowCasting, nukeCasting := isCasting(cb.slow), isCasting(cb.nuke)
		cb.start(&timedStubAction{readyAt: time.Minute})
		cb.runUntil(0)
		cb.castOf(cb.slow).Execute(cb.sim)

		cb.runUntil(earlyLook)
		if !slowCasting.GetBool(cb.sim) || nukeCasting.GetBool(cb.sim) {
			t.Fatalf("at 1.99 s only the slow cast should be casting")
		}
		cb.runUntil(castEnd)
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
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)
		paid.Cast(ct.sim, ct.fw.CurrentTarget)

		ct.runUntil(earlyLook)
		if want := ct.fw.CurrentMana() - 100; mana.GetFloat(ct.sim) != want || manaPercent.GetFloat(ct.sim) != want/ct.fw.MaxMana() {
			t.Fatalf("during the 2 s cast mana should read %.1f, what is left once it pays 100, read %.1f (%.4f)", want, mana.GetFloat(ct.sim), manaPercent.GetFloat(ct.sim))
		}
		ct.runUntil(castEnd)
		if mana.GetFloat(ct.sim) != ct.fw.CurrentMana() {
			t.Fatalf("once the cast pays at 2 s mana should read as it is, %.1f, read %.1f", ct.fw.CurrentMana(), mana.GetFloat(ct.sim))
		}
	})
}
