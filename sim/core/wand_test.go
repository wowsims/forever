package core

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A wand-like Shoot: one 100 damage shot per 2 s wand speed, for as long as the fight lasts. Like
// RegisterShoot's Shoot it is hasted by cast speed and refused while moving; unlike it, it ignores
// resists, so every shot deals exactly 100.
func fakeShoot(char *Character, shots *[]time.Duration) *Spell {
	spell := char.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 5019},
		Flags:            SpellFlagChanneled | SpellFlagAPL | SpellFlagIgnoreResists,
		SpellSchool:      SpellSchoolShadow,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskRangedAuto,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ExtraCastCondition: func(_ *Simulation, _ *Unit) bool {
			return !char.Moving
		},
		Dot: WandChannel(),
	})
	char.EnableWand(spell, func() time.Duration { return 2 * time.Second }, func(sim *Simulation, target *Unit) {
		spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		*shots = append(*shots, sim.CurrentTime)
	})
	return spell
}

// Ready once from readyAt.
type timedStubAction struct {
	defaultAPLActionImpl
	readyAt time.Duration
	ranAt   []time.Duration
}

func (action *timedStubAction) IsReady(sim *Simulation) bool {
	return sim.CurrentTime >= action.readyAt && len(action.ranAt) == 0
}
func (action *timedStubAction) Execute(sim *Simulation) {
	action.ranAt = append(action.ranAt, sim.CurrentTime)
}
func (action *timedStubAction) String() string { return "timed test action" }

type channelTest struct {
	sim        *Simulation
	fw         *FakeRageWarrior
	rot        *APLRotation
	spell      *Spell
	shots      []time.Duration
	channelEnd []time.Duration
}

func newChannelTest(register func(ct *channelTest, char *Character) *Spell) *channelTest {
	ct := &channelTest{}
	ct.sim, ct.fw, ct.rot = setupFakeRageRotationWith(func(char *Character) {
		ct.spell = register(ct, char)
	})
	ct.spell.finalize()
	ct.sim.Cleanup()
	ct.sim.Reset()
	ct.fw.OnChannelEnd(func(sim *Simulation, _ *Dot) {
		ct.channelEnd = append(ct.channelEnd, sim.CurrentTime)
	})
	return ct
}

func newShootTest() *channelTest {
	return newChannelTest(func(ct *channelTest, char *Character) *Spell {
		return fakeShoot(char, &ct.shots)
	})
}

// Shoot and one more spell, which gets its own cooldown timer when the config names a cooldown.
func newSpellShootTest(config SpellConfig) (*channelTest, *Spell) {
	var spell *Spell
	ct := newChannelTest(func(ct *channelTest, char *Character) *Spell {
		if config.Cast.CD.Duration > 0 {
			config.Cast.CD.Timer = char.NewTimer()
		}
		spell = char.RegisterSpell(config)
		return fakeShoot(char, &ct.shots)
	})
	spell.finalize()
	return ct, spell
}

func (ct *channelTest) wanding() bool {
	return ct.fw.Wand.wanding()
}

func (ct *channelTest) castOf(spell *Spell) APLActionImpl {
	return ct.rot.newActionCastSpell(&proto.APLActionCastSpell{SpellId: spell.ActionID.ToProto()})
}

func (ct *channelTest) channelOf(spell *Spell, interruptIf string, allowRecast bool) APLActionImpl {
	return ct.rot.newActionChannelSpell(&proto.APLActionChannelSpell{SpellId: spell.ActionID.ToProto(), InterruptIf: constValue(interruptIf), AllowRecast: allowRecast})
}

func (ct *channelTest) autocastAction() APLActionImpl {
	return ct.rot.newActionAutocastOtherCooldowns(&proto.APLActionAutocastOtherCooldowns{})
}

// The pull's rotation action stays pending at 0; the melee swings are turned off.
func (ct *channelTest) start(actions ...APLActionImpl) {
	ct.rot.priorityList = nil
	for _, impl := range actions {
		action := newAPLActionEntry(nil, impl)
		action.Finalize(ct.rot)
		ct.rot.priorityList = append(ct.rot.priorityList, action)
	}
	ct.sim.PrePull()
	ct.sim.Step()
	ct.fw.AutoAttacks.CancelAutoSwing(ct.sim)
}

// Runs every pending action and task up to and including until, then moves the clock to until.
func (ct *channelTest) runUntil(until time.Duration) {
	sim := ct.sim
	for (sim.pendingActions[len(sim.pendingActions)-1].NextActionAt <= until || sim.minTaskTime <= until) && !sim.Step() {
	}
	sim.advance(until)
}

func (ct *channelTest) damage() float64 {
	metrics := ct.spell.SpellMetrics[ct.fw.CurrentTarget.UnitIndex]
	return metrics.TotalDamage
}

func seconds(values ...float64) []time.Duration {
	return MapSlice(values, DurationFromSeconds)
}

// A spell that only records when it is cast.
func recordedSpellConfig(id int32, castAt *[]time.Duration) SpellConfig {
	return SpellConfig{
		ActionID: ActionID{SpellID: id},
		Flags:    SpellFlagAPL,
		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			*castAt = append(*castAt, sim.CurrentTime)
		},
	}
}

// A one-minute major cooldown, on the GCD when gcd is set, ready from 3 s.
type testCooldown struct {
	gcd       time.Duration
	priority  int32
	explosive bool // An explosive, which goes right after the GCD starts.
	spell     *Spell
	castAt    []time.Duration
}

func newCooldownShootTest(cooldowns ...*testCooldown) *channelTest {
	ct := newChannelTest(func(ct *channelTest, char *Character) *Spell {
		for i, cooldown := range cooldowns {
			config := recordedSpellConfig(910100+int32(i), &cooldown.castAt)
			config.Cast = CastConfig{DefaultCast: Cast{GCD: cooldown.gcd}, CD: Cooldown{Timer: char.NewTimer(), Duration: time.Minute}}
			cooldown.spell = char.RegisterSpell(config)
			char.AddMajorCooldown(MajorCooldown{Spell: cooldown.spell, Type: CooldownTypeDPS, Priority: cooldown.priority})
		}
		return fakeShoot(char, &ct.shots)
	})
	ct.fw.majorCooldownManager.finalize()
	ct.fw.majorCooldownManager.reset(ct.sim)
	for _, cooldown := range cooldowns {
		cooldown.spell.finalize()
		cooldown.spell.CD.Set(3 * time.Second)
	}
	ct.fw.UpdateMajorCooldowns()
	return ct
}

// A Drain Life-like channel of five 1 s ticks, with its ticks recorded.
func drainConfig(drainTicks *[]time.Duration, cast CastConfig) SpellConfig {
	config := fakeWarlockDot(689, "Drain Life", ClassFlags{Family: 5, Mask: [4]uint32{8}}, SpellFlagChanneled|SpellFlagAPL)
	config.Cast = cast
	onTick := config.Dot.OnTick
	config.Dot.OnTick = func(sim *Simulation, target *Unit, dot *Dot) {
		onTick(sim, target, dot)
		*drainTicks = append(*drainTicks, sim.CurrentTime)
	}
	return config
}

// An instant on the GCD that records its casts.
func nukeConfig(castAt *[]time.Duration) SpellConfig {
	config := recordedSpellConfig(910110, castAt)
	config.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault}}
	return config
}

func TestWandShootScheduler(t *testing.T) {
	t.Run("a cast fires one shot", func(t *testing.T) {
		ct := newShootTest()
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)
		ct.castOf(ct.spell).Execute(ct.sim)

		ct.runUntil(499 * time.Millisecond)
		if len(ct.shots) != 0 || !ct.wanding() {
			t.Fatalf("a cast Shoot should still wind up at 0.499 s, shots at %v", ct.shots)
		}
		ct.runUntil(2001 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5)) || ct.damage() != 100 {
			t.Fatalf("a cast Shoot should fire one 100 damage shot at 0.5 s, shots at %v for %.1f", ct.shots, ct.damage())
		}
		if ct.fw.ChanneledDot != nil {
			t.Fatalf("a cast Shoot should stop after its one shot")
		}
	})

	t.Run("a channel keeps shooting", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.channelOf(ct.spell, "true", false))

		ct.runUntil(3500 * time.Millisecond)
		wand := ct.fw.Wand
		if !ct.wanding() || wand.timeToNext(ct.sim) != time.Second || wand.timeSinceLast(ct.sim) != time.Second || wand.swingSpeed() != 2*time.Second {
			t.Fatalf("at 3.5 s the wand should be shooting with 1 s to the next shot, 1 s since the last and a 2 s speed, got %t, %s, %s, %s",
				ct.wanding(), wand.timeToNext(ct.sim), wand.timeSinceLast(ct.sim), wand.swingSpeed())
		}

		ct.runUntil(6501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5, 2.5, 4.5, 6.5)) || !ct.wanding() || len(ct.channelEnd) != 0 {
			t.Fatalf("a Shoot channel with nothing else ready should shoot at 0.5 s and then every 2 s and keep going, shots at %v, ended at %v", ct.shots, ct.channelEnd)
		}

		ct.sim.Cleanup()
		ct.sim.Reset()
		ct.sim.advance(500 * time.Millisecond)
		if got := wand.timeSinceLast(ct.sim); got != 500*time.Millisecond {
			t.Fatalf("after a reset the time since the last shot should count from the pull, got %s at 0.5 s", got)
		}
	})

	t.Run("what restarts the wand timer", func(t *testing.T) {
		var castAt, drainTicks []time.Duration
		racial := recordedSpellConfig(910120, &castAt)
		racial.Flags |= SpellFlagNoOnCastComplete
		racial.Cast = CastConfig{CD: Cooldown{Duration: time.Minute}}
		proc := recordedSpellConfig(910121, &castAt)
		proc.Flags = SpellFlagNoOnCastComplete
		hardcast := recordedSpellConfig(910122, &castAt)
		hardcast.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: time.Second}}

		for _, c := range []struct {
			name      string
			config    SpellConfig
			proc      bool // Cast straight through Spell.Cast rather than by the rotation.
			shootAt   time.Duration
			firstShot time.Duration
		}{
			{"a spell at 0, a wand speed after it", nukeConfig(&castAt), false, time.Second, 2 * time.Second},
			{"a spell at 0, Shoot 3 s later only waits for the wind-up", nukeConfig(&castAt), false, 3 * time.Second, 3500 * time.Millisecond},
			{"an off-GCD racial or potion cast at 0", racial, false, 500 * time.Millisecond, 2 * time.Second},
			{"a proc at 0 does not", proc, true, 500 * time.Millisecond, time.Second},
			{"a hardcast from its end at 1 s", hardcast, false, time.Second, 3 * time.Second},
			{"another channel from its end at 5 s", drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}}), false, 5 * time.Second, 7 * time.Second},
			{"a proc cast of a spell on the GCD does not", nukeConfig(&castAt), true, 500 * time.Millisecond, time.Second},
			{"a proc cast of a spell with a cooldown does not", racial, true, 500 * time.Millisecond, time.Second},
			{"a proc hardcast does not", hardcast, true, time.Second, 1500 * time.Millisecond},
		} {
			ct, spell := newSpellShootTest(c.config)
			ct.start(&timedStubAction{readyAt: time.Minute})
			ct.runUntil(0)
			if c.proc {
				if !spell.Cast(ct.sim, ct.fw.CurrentTarget) {
					t.Fatalf("%s: the spell could not be cast", c.name)
				}
			} else {
				ct.castOf(spell).Execute(ct.sim)
			}
			ct.runUntil(c.shootAt)
			ct.castOf(ct.spell).Execute(ct.sim)
			ct.runUntil(c.firstShot + time.Millisecond)
			if !slices.Equal(ct.shots, []time.Duration{c.firstShot}) {
				t.Fatalf("%s: Shoot cast at %s should fire at %s, shots at %v", c.name, c.shootAt, c.firstShot, ct.shots)
			}
		}
	})

	t.Run("an interrupt takes the shot's slot", func(t *testing.T) {
		ct := newShootTest()
		stub := &timedStubAction{readyAt: 3 * time.Second}
		ct.start(stub, ct.channelOf(ct.spell, "true", false))

		ct.runUntil(4501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5, 2.5)) || ct.damage() != 200 {
			t.Fatalf("Shoot should fire at 0.5 and 2.5 s and give its 4.5 s slot away, shots at %v for %.1f", ct.shots, ct.damage())
		}
		if !slices.Equal(ct.channelEnd, seconds(4.5)) || !slices.Equal(stub.ranAt, seconds(4.5)) {
			t.Fatalf("an action ready at 3 s should end Shoot and run at the 4.5 s slot, channel ended at %v, action ran at %v", ct.channelEnd, stub.ranAt)
		}
	})

	t.Run("no interrupt when the GCD ends", func(t *testing.T) {
		ct := newShootTest()
		stub := &timedStubAction{readyAt: time.Second}
		ct.start(stub, ct.channelOf(ct.spell, "true", false))
		ct.runUntil(0)
		ct.fw.SetGCDTimer(ct.sim, 1500*time.Millisecond)

		ct.runUntil(2501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5)) {
			t.Fatalf("Shoot should fire at 0.5 s and give its 2.5 s slot to the waiting action, shots at %v", ct.shots)
		}
		if !slices.Equal(ct.channelEnd, seconds(2.5)) || !slices.Equal(stub.ranAt, seconds(2.5)) {
			t.Fatalf("the action waiting on the GCD should end Shoot at the 2.5 s slot, not at 1.5 s, channel ended at %v, action ran at %v", ct.channelEnd, stub.ranAt)
		}
	})

	t.Run("an interrupt inside the spell queue window still takes the slot", func(t *testing.T) {
		ct := newShootTest()
		stub := &timedStubAction{readyAt: time.Second}
		ct.start(stub, ct.channelOf(ct.spell, "true", false))
		ct.runUntil(time.Second)
		ct.fw.SetGCDTimer(ct.sim, 2800*time.Millisecond)
		ct.runUntil(1200 * time.Millisecond)
		ct.fw.ReactToEvent(ct.sim, false, false)
		ct.runUntil(1200 * time.Millisecond)

		ct.runUntil(2501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5)) || !slices.Equal(ct.channelEnd, seconds(2.5)) {
			t.Fatalf("with the GCD 0.3 s out at the 2.5 s slot Shoot should end there with no shot, shots at %v, ended at %v", ct.shots, ct.channelEnd)
		}
		ct.runUntil(2801 * time.Millisecond)
		if !slices.Equal(stub.ranAt, seconds(2.8)) {
			t.Fatalf("the action should run when the GCD ends at 2.8 s, ran at %v", stub.ranAt)
		}
	})

	t.Run("a normal channel ticks before its interrupt", func(t *testing.T) {
		var drainTicks []time.Duration
		ct := newChannelTest(func(_ *channelTest, char *Character) *Spell {
			return char.RegisterSpell(drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}}))
		})
		stub := &timedStubAction{readyAt: 500 * time.Millisecond}
		ct.start(stub, ct.channelOf(ct.spell, "true", false))

		ct.runUntil(1001 * time.Millisecond)
		metrics := ct.spell.SpellMetrics[ct.fw.CurrentTarget.UnitIndex]
		if !slices.Equal(drainTicks, seconds(1)) || metrics.TotalTickDamage == 0 {
			t.Fatalf("Drain Life should deal its 1 s tick before it ends, ticks at %v for %.1f", drainTicks, metrics.TotalTickDamage)
		}
		if !slices.Equal(ct.channelEnd, seconds(1)) {
			t.Fatalf("Drain Life should end at its 1 s tick, ended at %v", ct.channelEnd)
		}
		ct.runUntil(1500 * time.Millisecond)
		if !slices.Equal(stub.ranAt, seconds(1.5)) {
			t.Fatalf("the action that ended Drain Life should run when the GCD ends at 1.5 s, ran at %v", stub.ranAt)
		}
	})

	t.Run("a channel allowed to recast keeps shooting", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.channelOf(ct.spell, "true", true))

		ct.runUntil(4501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5, 2.5, 4.5)) || len(ct.channelEnd) != 0 {
			t.Fatalf("a Shoot channel allowed to recast should shoot at 0.5, 2.5 and 4.5 s without being cut, shots at %v, ended at %v", ct.shots, ct.channelEnd)
		}
	})

	t.Run("no cooldown is used while wanding", func(t *testing.T) {
		onGCD := &testCooldown{gcd: GCDDefault, priority: 1}
		potion := &testCooldown{}
		ct := newCooldownShootTest(onGCD, potion)
		ct.start(ct.autocastAction(), ct.channelOf(ct.spell, "false", false))

		ct.runUntil(6501 * time.Millisecond)
		if len(potion.castAt) != 0 || len(onGCD.castAt) != 0 {
			t.Fatalf("with Shoot never interrupted no cooldown should be used, the potion at %v, the one on the GCD at %v", potion.castAt, onGCD.castAt)
		}
		if !slices.Equal(ct.shots, seconds(0.5, 2.5, 4.5, 6.5)) || len(ct.channelEnd) != 0 {
			t.Fatalf("Shoot should keep shooting, shots at %v, ended at %v", ct.shots, ct.channelEnd)
		}
	})

	t.Run("a ready cooldown above Shoot takes the shot's slot", func(t *testing.T) {
		potion := &testCooldown{}
		ct := newCooldownShootTest(potion)
		ct.start(ct.autocastAction(), ct.channelOf(ct.spell, "true", false))

		ct.runUntil(6501 * time.Millisecond)
		if !slices.Equal(potion.castAt, seconds(4.5)) {
			t.Fatalf("a potion ready at 3 s should be used at the 4.5 s slot, used at %v", potion.castAt)
		}
		if !slices.Equal(ct.shots, seconds(0.5, 2.5, 6.5)) || !slices.Equal(ct.channelEnd, seconds(4.5)) || !ct.wanding() {
			t.Fatalf("Shoot should give its 4.5 s slot to the potion and start again, shooting a wand speed after it, shots at %v, ended at %v", ct.shots, ct.channelEnd)
		}
	})

	t.Run("cooldowns take one slot each", func(t *testing.T) {
		onGCD := &testCooldown{gcd: GCDDefault, priority: 1}
		potion := &testCooldown{}
		ct := newCooldownShootTest(onGCD, potion)
		if ct.fw.majorCooldownManager.majorCooldowns[0].Spell != onGCD.spell {
			t.Fatalf("the cooldown on the GCD should sort ahead of the potion")
		}
		ct.start(ct.autocastAction(), ct.channelOf(ct.spell, "true", false))

		ct.runUntil(8501 * time.Millisecond)
		if !slices.Equal(onGCD.castAt, seconds(4.5)) || !slices.Equal(potion.castAt, seconds(6.5)) {
			t.Fatalf("the one on the GCD should take the 4.5 s slot and the potion, held by that GCD, the 6.5 s slot: the one on the GCD at %v, the potion at %v", onGCD.castAt, potion.castAt)
		}
		if !slices.Equal(ct.shots, seconds(0.5, 2.5, 8.5)) || !slices.Equal(ct.channelEnd, seconds(4.5, 6.5)) {
			t.Fatalf("Shoot should shoot at 0.5, 2.5 and 8.5 s around the two cooldowns, shots at %v, ended at %v", ct.shots, ct.channelEnd)
		}
	})

	t.Run("a cast fires its shot with another action ready", func(t *testing.T) {
		ct := newShootTest()
		stub := &timedStubAction{readyAt: 200 * time.Millisecond}
		ct.start(stub, ct.castOf(ct.spell))

		ct.runUntil(501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5)) || !slices.Equal(stub.ranAt, seconds(0.5)) || !slices.Equal(ct.channelEnd, seconds(0.5)) {
			t.Fatalf("a cast Shoot should fire at 0.5 s although an action is ready from 0.2 s, then end and let it run, shots at %v, action ran at %v, ended at %v", ct.shots, stub.ranAt, ct.channelEnd)
		}
		ct.runUntil(2501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5, 2.5)) {
			t.Fatalf("the cast Shoot after the action should fire a wand speed after the last shot, shots at %v", ct.shots)
		}
	})

	t.Run("a cast Shoot fires on the wand speed with a clip delay set", func(t *testing.T) {
		ct := newShootTest()
		ct.fw.ChannelClipDelay = 150 * time.Millisecond
		ct.start(ct.castOf(ct.spell))

		ct.runUntil(4501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5, 2.5, 4.5)) {
			t.Fatalf("Shoot cast after each shot should fire every 2 s with no clip delay, shots at %v", ct.shots)
		}
	})

	t.Run("starting Shoot ends the rotation pass", func(t *testing.T) {
		var nukeAt []time.Duration
		ct, nuke := newSpellShootTest(nukeConfig(&nukeAt))
		ct.start(ct.channelOf(ct.spell, "true", false), ct.castOf(nuke))

		ct.runUntil(0)
		if !ct.wanding() || len(nukeAt) != 0 {
			t.Fatalf("at 0 only Shoot should start, wanding %t, spell on the GCD cast at %v", ct.wanding(), nukeAt)
		}
		ct.runUntil(2501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5, 2.5)) || len(nukeAt) != 0 {
			t.Fatalf("Shoot above the spell should keep shooting, shots at %v, spell cast at %v", ct.shots, nukeAt)
		}
	})

	t.Run("a spell after an instant and Shoot waits for the shot slot", func(t *testing.T) {
		var nukeAt []time.Duration
		ct, nuke := newSpellShootTest(nukeConfig(&nukeAt))
		ct.start(ct.castOf(nuke), ct.channelOf(ct.spell, "true", false))

		ct.runUntil(0)
		if !slices.Equal(nukeAt, seconds(0)) || !ct.wanding() {
			t.Fatalf("at 0 the instant and then Shoot should start in one pass, instant at %v, wanding %t", nukeAt, ct.wanding())
		}
		ct.runUntil(2001 * time.Millisecond)
		if !slices.Equal(nukeAt, seconds(0, 2)) || len(ct.shots) != 0 || !slices.Equal(ct.channelEnd, seconds(2)) {
			t.Fatalf("the next instant should wait past its GCD at 1.5 s for the 2 s slot and take it, instants at %v, shots at %v, ended at %v", nukeAt, ct.shots, ct.channelEnd)
		}
	})

	t.Run("Shoot above a normal channel runs alone", func(t *testing.T) {
		var drainTicks []time.Duration
		ct, drain := newSpellShootTest(drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}}))
		stub := &timedStubAction{readyAt: 3 * time.Second}
		ct.start(stub, ct.channelOf(ct.spell, "true", false), ct.channelOf(drain, "true", false))

		ct.runUntil(0)
		if dot := ct.fw.ChanneledDot; dot == nil || dot.Spell != ct.spell || drain.Dot(ct.fw.CurrentTarget).IsActive() {
			t.Fatalf("at 0 Shoot should be the only channel")
		}
		ct.runUntil(4501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5, 2.5)) || len(drainTicks) != 0 {
			t.Fatalf("only Shoot should run, shots at %v, Drain Life ticks at %v", ct.shots, drainTicks)
		}
		if !slices.Equal(ct.channelEnd, seconds(4.5)) || !slices.Equal(stub.ranAt, seconds(4.5)) {
			t.Fatalf("Shoot should end once, at 4.5 s for the action ready at 3 s, ended at %v, action ran at %v", ct.channelEnd, stub.ranAt)
		}
	})

	t.Run("a failed channel leaves no interrupt condition to a cast Shoot", func(t *testing.T) {
		var drainTicks []time.Duration
		ct, drain := newSpellShootTest(drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}, CD: Cooldown{Duration: time.Minute}}))
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)
		drain.CD.Set(time.Minute)

		ct.channelOf(drain, "true", false).Execute(ct.sim)
		if ct.fw.ChanneledDot != nil {
			t.Fatalf("Drain Life on cooldown should not start")
		}
		ct.castOf(ct.spell).Execute(ct.sim)

		ct.runUntil(2501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5)) || !slices.Equal(ct.channelEnd, seconds(0.5)) {
			t.Fatalf("the cast Shoot should fire its one shot at 0.5 s and end, shots at %v, ended at %v", ct.shots, ct.channelEnd)
		}
	})

	t.Run("a channel cast that never starts its channel leaves no interrupt condition to a cast Shoot", func(t *testing.T) {
		var drainTicks []time.Duration
		config := drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}})
		config.ApplyEffects = func(_ *Simulation, _ *Unit, _ *Spell) {}
		ct, drain := newSpellShootTest(config)
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)

		ct.channelOf(drain, "false", false).Execute(ct.sim)
		if ct.fw.ChanneledDot != nil {
			t.Fatalf("a cast that applies no dot should start no channel")
		}
		ct.castOf(ct.spell).Execute(ct.sim)

		ct.runUntil(6501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(2)) || !slices.Equal(ct.channelEnd, seconds(2)) {
			t.Fatalf("the cast Shoot should fire its one shot a wand speed after the cast at 0 and end, shots at %v, ended at %v", ct.shots, ct.channelEnd)
		}
	})

	t.Run("the rotation wakes at the GCD after Shoot ends inside it", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.castOf(ct.spell))
		ct.runUntil(0)
		ct.fw.SetGCDTimer(ct.sim, 3*time.Second)
		ct.runUntil(200 * time.Millisecond)
		ct.fw.ReactToEvent(ct.sim, false, false)
		ct.runUntil(200 * time.Millisecond)

		ct.runUntil(3501 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(0.5, 3.5)) {
			t.Fatalf("after the 0.5 s shot ends Shoot inside the GCD, the rotation should cast it again when the GCD ends at 3 s, shots at %v", ct.shots)
		}
	})

	t.Run("moving ends Shoot with no shot", func(t *testing.T) {
		for _, moveAt := range seconds(0.25, 1) {
			ct := newShootTest()
			ct.start(ct.channelOf(ct.spell, "true", false))
			ct.runUntil(moveAt)
			before := len(ct.shots)
			ct.fw.MoveDuration(1500*time.Millisecond, ct.sim)
			if ct.wanding() || !slices.Equal(ct.channelEnd, []time.Duration{moveAt}) {
				t.Fatalf("moving at %s should end Shoot at once, ended at %v", moveAt, ct.channelEnd)
			}

			moveEnd := moveAt + 1500*time.Millisecond
			ct.runUntil(moveEnd - time.Millisecond)
			if ct.wanding() || len(ct.shots) != before {
				t.Fatalf("Shoot should not start or shoot while moving from %s, shots at %v", moveAt, ct.shots)
			}
			ct.runUntil(moveEnd + time.Second)
			if !ct.wanding() || len(ct.shots) != before+1 || ct.shots[before] < moveEnd+500*time.Millisecond {
				t.Fatalf("after the move ending at %s the rotation should start Shoot again, shots at %v", moveEnd, ct.shots)
			}
		}
	})

	t.Run("a channel with nothing else ready keeps shooting", func(t *testing.T) {
		var nukeAt []time.Duration
		ct, nuke := newSpellShootTest(nukeConfig(&nukeAt))
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)
		ct.castOf(nuke).Execute(ct.sim)
		ct.channelOf(ct.spell, "true", false).Execute(ct.sim)

		ct.runUntil(6001 * time.Millisecond)
		if !slices.Equal(ct.shots, seconds(2, 4, 6)) || len(ct.channelEnd) != 0 {
			t.Fatalf("a Shoot channel whose interrupt condition holds with nothing else to do should shoot at 2, 4 and 6 s, shots at %v, ended at %v", ct.shots, ct.channelEnd)
		}
	})

	t.Run("Shoot inside a strict sequence shoots until the next subaction is ready", func(t *testing.T) {
		for _, c := range []struct {
			name         string
			nukeReadyAt  time.Duration
			shots, ended []time.Duration
			nukeCasts    []time.Duration
		}{
			{"next subaction on cooldown", time.Minute, seconds(0.5, 2.5, 4.5), nil, nil},
			{"next subaction ready at 3 s", 3 * time.Second, seconds(0.5, 2.5), seconds(4.5), seconds(4.5)},
		} {
			var nukeAt []time.Duration
			config := nukeConfig(&nukeAt)
			config.Cast.CD = Cooldown{Duration: time.Minute}
			ct, nuke := newSpellShootTest(config)
			nuke.CD.Set(c.nukeReadyAt)
			sequence := &APLActionStrictSequence{unit: &ct.fw.Unit, subactions: []*APLAction{
				{impl: ct.channelOf(ct.spell, "true", false)},
				{impl: ct.castOf(nuke)},
			}}
			ct.start(sequence)

			ct.runUntil(4501 * time.Millisecond)
			if !slices.Equal(ct.shots, c.shots) || !slices.Equal(ct.channelEnd, c.ended) || !slices.Equal(nukeAt, c.nukeCasts) {
				t.Fatalf("%s: shots at %v, Shoot ended at %v, nuke cast at %v, want %v, %v, %v", c.name, ct.shots, ct.channelEnd, nukeAt, c.shots, c.ended, c.nukeCasts)
			}
		}
	})

	t.Run("a strict sequence waits for a spell queued on its own cooldown", func(t *testing.T) {
		for _, gcd := range []time.Duration{0, GCDDefault} {
			t.Run(gcd.String(), func(t *testing.T) {
				var casts []int32
				var first, second, fallback *Spell
				ct := newChannelTest(func(ct *channelTest, char *Character) *Spell {
					register := func(id int32, gcd time.Duration) *Spell {
						return char.RegisterSpell(SpellConfig{
							ActionID: ActionID{SpellID: id},
							Flags:    SpellFlagAPL,
							Cast: CastConfig{
								DefaultCast: Cast{GCD: gcd},
								CD:          Cooldown{Timer: char.NewTimer(), Duration: time.Minute},
							},
							ApplyEffects: func(_ *Simulation, _ *Unit, _ *Spell) { casts = append(casts, id) },
						})
					}
					first = register(910124, gcd)
					second = register(910125, GCDDefault)
					fallback = register(910126, GCDDefault)
					return fakeShoot(char, &ct.shots)
				})
				for _, spell := range []*Spell{first, second, fallback} {
					spell.finalize()
				}
				ct.start(&timedStubAction{readyAt: time.Minute})
				ct.runUntil(750 * time.Millisecond)
				first.CD.Set(time.Second)
				sequence := &APLAction{impl: &APLActionStrictSequence{unit: &ct.fw.Unit, subactions: []*APLAction{
					{impl: ct.castOf(first)},
					{impl: ct.castOf(second)},
				}}}
				sequence.Finalize(ct.rot)
				ct.rot.priorityList = []*APLAction{sequence, {impl: ct.castOf(fallback)}}
				ct.fw.RotationTimer.Set(ct.sim.CurrentTime)
				ct.rot.DoNextAction(ct.sim)
				if len(casts) != 0 {
					t.Fatalf("a queued sequence cast before its cooldown ended: %v", casts)
				}
				ct.runUntil(time.Second + gcd + time.Millisecond)
				if len(casts) < 2 || !slices.Equal(casts[:2], []int32{first.SpellID, second.SpellID}) {
					t.Fatalf("the queued spell must be followed by the next sequence step, casts %v", casts)
				}
			})
		}
	})

	t.Run("a spell queueable at the slot takes it", func(t *testing.T) {
		var nukeAt []time.Duration
		ct, nuke := newSpellShootTest(nukeConfig(&nukeAt))
		ct.start(ct.castOf(nuke), ct.channelOf(ct.spell, "true", false))
		ct.runUntil(0)
		ct.fw.SetGCDTimer(ct.sim, 2300*time.Millisecond)

		ct.runUntil(2001 * time.Millisecond)
		if len(ct.shots) != 0 || !slices.Equal(ct.channelEnd, seconds(2)) {
			t.Fatalf("with the GCD 0.3 s out at the 2 s slot the spell above Shoot should take the slot, shots at %v, ended at %v", ct.shots, ct.channelEnd)
		}
		ct.runUntil(2301 * time.Millisecond)
		if !slices.Equal(nukeAt, seconds(0, 2.3)) {
			t.Fatalf("the spell should be cast when the GCD ends at 2.3 s, cast at %v", nukeAt)
		}
	})

	t.Run("a queued channel that never starts leaves no interrupt condition to a cast Shoot", func(t *testing.T) {
		for _, failed := range []bool{false, true} {
			var castAt []time.Duration
			hardcast := recordedSpellConfig(910122, &castAt)
			hardcast.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: time.Second}}
			ct, spell := newSpellShootTest(hardcast)
			ct.start(&timedStubAction{readyAt: time.Minute})
			ct.runUntil(0)
			ct.castOf(spell).Execute(ct.sim)
			ct.runUntil(700 * time.Millisecond)
			ct.channelOf(ct.spell, "true", false).Execute(ct.sim)
			if ct.fw.QueuedSpell == nil || ct.fw.QueuedSpell.spell != ct.spell {
				t.Fatalf("the Shoot channel should queue behind the hardcast ending at 1 s")
			}

			shootAt := time.Second
			if failed {
				ct.runUntil(time.Second)
				ct.fw.MoveDuration(500*time.Millisecond, ct.sim)
				shootAt = 1500 * time.Millisecond
			} else {
				ct.fw.CancelQueuedSpell(ct.sim)
			}
			ct.runUntil(shootAt)
			if ct.wanding() {
				t.Fatalf("failed %t: the queued Shoot channel should not have started", failed)
			}
			ct.castOf(ct.spell).Execute(ct.sim)

			ct.runUntil(6 * time.Second)
			if !slices.Equal(ct.shots, seconds(3)) || !slices.Equal(ct.channelEnd, seconds(3)) {
				t.Fatalf("failed %t: the cast Shoot should fire its one shot a wand speed after the hardcast and end, shots at %v, ended at %v", failed, ct.shots, ct.channelEnd)
			}
		}
	})

	t.Run("a rotation hardcast that fails at completion leaves no cast for a proc of the spell to complete", func(t *testing.T) {
		var castAt []time.Duration
		hardcast := recordedSpellConfig(910123, &castAt)
		hardcast.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: time.Second}}
		ct, spell := newSpellShootTest(hardcast)
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)
		ct.castOf(spell).Execute(ct.sim)
		ct.runUntil(500 * time.Millisecond)
		ct.fw.PseudoStats.Incapacitated = true
		ct.runUntil(time.Second)
		ct.fw.PseudoStats.Incapacitated = false
		if len(castAt) != 0 {
			t.Fatalf("the hardcast should fail when it completes, effects applied at %v", castAt)
		}

		spell.applyEffects(ct.sim, ct.fw.CurrentTarget)
		if len(castAt) != 1 || ct.fw.Wand.timerAt != -NeverExpires {
			t.Fatalf("a proc cast of the spell should leave the wand timer alone, effects at %v, timer at %v", castAt, ct.fw.Wand.timerAt)
		}
	})

	t.Run("a channel whose dot sits on its related spell takes its interrupt condition", func(t *testing.T) {
		var related *Spell
		var drainTicks []time.Duration
		ct := newChannelTest(func(_ *channelTest, char *Character) *Spell {
			related = char.RegisterSpell(drainConfig(&drainTicks, CastConfig{}))
			return char.RegisterSpell(SpellConfig{
				ActionID:        ActionID{SpellID: 910130},
				Flags:           SpellFlagChanneled | SpellFlagAPL,
				Cast:            CastConfig{DefaultCast: Cast{GCD: GCDDefault}},
				RelatedDotSpell: related,
				ApplyEffects: func(sim *Simulation, target *Unit, _ *Spell) {
					related.Dot(target).Apply(sim)
				},
			})
		})
		related.finalize()
		stub := &timedStubAction{readyAt: 500 * time.Millisecond}
		ct.start(stub, ct.channelOf(ct.spell, "true", false))

		ct.runUntil(1001 * time.Millisecond)
		if !slices.Equal(drainTicks, seconds(1)) || !slices.Equal(ct.channelEnd, seconds(1)) {
			t.Fatalf("the channel should end at its 1 s tick for the action ready at 0.5 s, ticks at %v, ended at %v", drainTicks, ct.channelEnd)
		}
		ct.runUntil(1500 * time.Millisecond)
		if !slices.Equal(stub.ranAt, seconds(1.5)) {
			t.Fatalf("the action should run when the GCD ends at 1.5 s, ran at %v", stub.ranAt)
		}
	})

	t.Run("Shoot without the unit's wand state still shoots", func(t *testing.T) {
		for _, channel := range []bool{false, true} {
			ct := newShootTest()
			ct.fw.Wand = nil
			action := ct.castOf(ct.spell)
			want := seconds(0.5)
			if channel {
				action = ct.channelOf(ct.spell, "true", false)
				want = seconds(0.5, 2.5, 4.5)
			}
			ct.start(action)

			ct.runUntil(4501 * time.Millisecond)
			if channel && ct.fw.ChanneledDot == nil {
				t.Fatalf("the Shoot channel should still run")
			}
			if !slices.Equal(ct.shots[:len(want)], want) {
				t.Fatalf("channel %t: Shoot should shoot at %v, shots at %v", channel, want, ct.shots)
			}
		}
	})
}

func TestValueWand(t *testing.T) {
	type reads struct {
		swing, toNext, sinceLast time.Duration
	}
	read := func(ct *channelTest) reads {
		rot, sim := ct.rot, ct.sim
		return reads{
			swing:     rot.newValueAutoSwingTime(&proto.APLValueAutoSwingTime{AutoType: proto.APLValueAutoSwingTime_Wand}, nil).GetDuration(sim),
			toNext:    rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoAttackType_WandAuto}, nil).GetDuration(sim),
			sinceLast: rot.newValueAutoTimeSinceLast(&proto.APLValueAutoTimeSinceLast{AutoType: proto.APLValueAutoAttackType_WandAuto}, nil).GetDuration(sim),
		}
	}

	t.Run("mid channel, also in Any", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.channelOf(ct.spell, "true", false))
		ct.runUntil(3 * time.Second)
		if got, want := read(ct), (reads{2 * time.Second, 1500 * time.Millisecond, 500 * time.Millisecond}); got != want {
			t.Fatalf("at 3 s into a Shoot channel the wand values read %+v, want %+v", got, want)
		}
		toNext := ct.rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{}, nil).GetDuration(ct.sim)
		sinceLast := ct.rot.newValueAutoTimeSinceLast(&proto.APLValueAutoTimeSinceLast{}, nil).GetDuration(ct.sim)
		if toNext != 1500*time.Millisecond || sinceLast != 500*time.Millisecond {
			t.Fatalf("with the melee swings off, Any should read the wand at 3 s into a Shoot channel: %s to the next, %s since the last, want 1.5 s and 0.5 s", toNext, sinceLast)
		}
	})

	t.Run("swing time at the current cast speed", func(t *testing.T) {
		ct := newShootTest()
		ct.start()
		ct.fw.MultiplyCastSpeed(ct.sim, 1.1)
		if got := read(ct).swing; got != 1818*time.Millisecond {
			t.Fatalf("with 10%% cast speed the wand swing time reads %s, want 1.818 s", got)
		}
	})

	t.Run("swing time while wanding stays the running speed", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.channelOf(ct.spell, "true", false))
		ct.runUntil(3 * time.Second)
		ct.fw.MultiplyCastSpeed(ct.sim, 1.1)
		if got := read(ct).swing; got != 2*time.Second {
			t.Fatalf("with cast speed gained mid-channel the wand swing time reads %s while wanding, want the running 2 s", got)
		}
		ct.fw.MoveDuration(time.Second, ct.sim)
		if ct.wanding() {
			t.Fatalf("moving should end Shoot")
		}
		if got := read(ct).swing; got != 1818*time.Millisecond {
			t.Fatalf("after the channel ended the wand swing time reads %s, want the hasted 1.818 s", got)
		}
	})

	t.Run("before any cast", func(t *testing.T) {
		ct := newShootTest()
		ct.start()
		if got := read(ct); got != (reads{swing: 2 * time.Second}) {
			t.Fatalf("before any cast the wand reads %+v, want only the 2 s swing time", got)
		}
	})

	t.Run("no wand", func(t *testing.T) {
		ct := newShootTest()
		ct.fw.Wand = nil
		ct.start()
		ct.runUntil(3 * time.Second)
		if got := read(ct); got != (reads{}) {
			t.Fatalf("a unit without a wand reads %+v, want all zero", got)
		}
	})
}
