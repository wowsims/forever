package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type pendingProcTest struct {
	*channelTest
	buff         *Aura
	debuffs      AuraArray
	undeclared   *Aura
	immediate    *Aura
	rolled       *Aura
	managed      *Aura
	refused      *Aura
	refusedNow   *Aura
	stacking     *Aura
	statProc     *StatBuffAura
	aplSpell     *Spell
	procSpell    *Spell
	mcdSpell     *Spell
	castTriggers int
}

// A spell cast from the APL, one that is not and a major cooldown, each a landed hit. Deferred triggers on those
// hits put up a buff and a debuff on the hit target, both declared, and a second buff left undeclared; an
// immediate trigger puts up a third, declared, at the hit. Two more deferred triggers declare a buff each, one
// rolling a stated chance below 1 and one a proc manager that always fires. A deferred and an immediate trigger
// declare a buff each whose handler never puts it up. Two more deferred triggers declare a stacking buff whose
// handler only adds a stack while it is up, and a Strength item proc. A cast trigger declares the debuff too,
// though a cast carries no target to find it on, and a hit-taken trigger declares it on the attacker.
func newPendingProcTest() *pendingProcTest {
	pt := &pendingProcTest{}
	pt.channelTest = newChannelTest(func(_ *channelTest, char *Character) *Spell {
		hit := func(id int32, flags SpellFlag) *Spell {
			config := castResultSpellConfig(id, flags)
			config.MissileSpeed = 0
			config.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
				spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
			}
			return char.RegisterSpell(config)
		}
		pt.aplSpell = hit(910304, SpellFlagAPL)
		pt.procSpell = hit(910305, SpellFlagNone)
		pt.mcdSpell = hit(910311, SpellFlagAPL|SpellFlagMCD)

		buff := func(label string, id int32) *Aura {
			return char.RegisterAura(Aura{Label: label, ActionID: ActionID{SpellID: id}, Duration: 10 * time.Second})
		}
		pt.buff = buff("Pending Buff", 910301)
		pt.undeclared = buff("Undeclared Buff", 910303)
		pt.immediate = buff("Immediate Buff", 910306)
		pt.rolled = buff("Rolled Buff", 910307)
		pt.managed = buff("Managed Buff", 910308)
		pt.refused = buff("Refused Buff", 910309)
		pt.refusedNow = buff("Refused Immediate Buff", 910310)
		pt.stacking = char.RegisterAura(Aura{Label: "Stacking Buff", ActionID: ActionID{SpellID: 910312}, Duration: time.Minute, MaxStacks: 5})
		pt.statProc = newStatAura(char, "Stat Proc", 910313, stats.Strength, 10*time.Second)
		char.ItemProcBuffs = append(char.ItemProcBuffs, pt.statProc)
		pt.debuffs = char.NewEnemyAuraArray(func(target *Unit) *Aura {
			return target.RegisterAura(Aura{Label: "Pending Debuff", ActionID: ActionID{SpellID: 910302}, Duration: 10 * time.Second})
		})

		trigger := func(name string, config ProcTrigger) {
			config.Name = name
			if config.Callback == CallbackEmpty {
				config.Callback = CallbackOnSpellHitDealt
				config.Outcome = OutcomeLanded
			}
			config.ProcMask = ProcMaskSpellDamage
			char.MakeProcTriggerAura(config)
		}
		trigger("Pending Buff Trigger", ProcTrigger{
			ActivatesAura: pt.buff,
			Handler:       func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.buff.Activate(sim) },
		})
		trigger("Pending Debuff Trigger", ProcTrigger{
			ActivatesAuras: pt.debuffs,
			Handler:        func(sim *Simulation, _ *Spell, result *SpellResult) { pt.debuffs.Get(result.Target).Activate(sim) },
		})
		trigger("Immediate Buff Trigger", ProcTrigger{
			TriggerImmediately: true,
			ActivatesAura:      pt.immediate,
			Handler:            func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.immediate.Activate(sim) },
		})
		trigger("Rolled Buff Trigger", ProcTrigger{
			ProcChance:    0.9999999,
			ActivatesAura: pt.rolled,
			Handler:       func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.rolled.Activate(sim) },
		})
		trigger("Managed Buff Trigger", ProcTrigger{
			DPM:           char.NewFixedProcChanceManager(1, ProcMaskSpellDamage),
			ActivatesAura: pt.managed,
			Handler:       func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.managed.Activate(sim) },
		})
		trigger("Refusing Buff Trigger", ProcTrigger{
			ActivatesAura: pt.refused,
			Handler:       func(*Simulation, *Spell, *SpellResult) {},
		})
		trigger("Refusing Immediate Buff Trigger", ProcTrigger{
			TriggerImmediately: true,
			ActivatesAura:      pt.refusedNow,
			Handler:            func(*Simulation, *Spell, *SpellResult) {},
		})
		trigger("Stacking Buff Trigger", ProcTrigger{
			ActivatesAura: pt.stacking,
			Handler: func(sim *Simulation, _ *Spell, _ *SpellResult) {
				if pt.stacking.IsActive() {
					pt.stacking.AddStack(sim)
				}
			},
		})
		trigger("Stat Proc Trigger", ProcTrigger{
			ActivatesAura: pt.statProc.Aura,
			Handler:       func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.statProc.Activate(sim) },
		})
		trigger("Undeclared Buff Trigger", ProcTrigger{
			Handler: func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.undeclared.Activate(sim) },
		})
		trigger("Pending Debuff Cast Trigger", ProcTrigger{
			Callback:       CallbackOnCastComplete,
			ActivatesAuras: pt.debuffs,
			Handler:        func(*Simulation, *Spell, *SpellResult) { pt.castTriggers++ },
		})
		trigger("Pending Debuff Hit Taken Trigger", ProcTrigger{
			Callback:       CallbackOnSpellHitTaken,
			Outcome:        OutcomeLanded,
			ActivatesAuras: pt.debuffs,
			Handler:        func(sim *Simulation, spell *Spell, _ *SpellResult) { pt.debuffs.Get(spell.Unit).Activate(sim) },
		})
		return pt.aplSpell
	})
	pt.procSpell.finalize()
	pt.mcdSpell.finalize()
	pt.fw.ReactionTime = 100 * time.Millisecond
	pt.start()
	return pt
}

func (pt *pendingProcTest) hit(spell *Spell) {
	spell.SkipCastAndApplyEffects(pt.sim, pt.fw.CurrentTarget)
}

func (pt *pendingProcTest) isActive(aura *Aura, includeReactionTime bool) APLValue {
	return pt.rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{AuraId: aura.ActionID.ToProto(), IncludeReactionTime: includeReactionTime}, nil)
}

func (pt *pendingProcTest) isInactive(aura *Aura, includeReactionTime bool) APLValue {
	return pt.rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{AuraId: aura.ActionID.ToProto(), IncludeReactionTime: includeReactionTime}, nil)
}

func (pt *pendingProcTest) debuffIsActive(includeReactionTime bool) APLValue {
	return pt.rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{
		SourceUnit:          &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
		AuraId:              pt.debuffs.Get(pt.fw.CurrentTarget).ActionID.ToProto(),
		IncludeReactionTime: includeReactionTime,
	}, nil)
}

// Inside the spell batch window a declared aura reads as up to the APL, and an undeclared one does not.
func TestAuraIsActiveSeesADeclaredProcBeforeItsHandlerRuns(t *testing.T) {
	pt := newPendingProcTest()
	active, inactive := pt.isActive(pt.buff, false), pt.isInactive(pt.buff, false)
	debuffActive := pt.debuffIsActive(false)
	undeclaredActive, undeclaredInactive := pt.isActive(pt.undeclared, false), pt.isInactive(pt.undeclared, false)

	pt.runUntil(time.Second)
	pt.hit(pt.procSpell)
	if pt.buff.IsActive() {
		t.Fatalf("buff active before its handler ran, want the handler deferred")
	}
	if !active.GetBool(pt.sim) || inactive.GetBool(pt.sim) || !debuffActive.GetBool(pt.sim) {
		t.Errorf("declared in the batch window: active = %t, inactive = %t, debuff active = %t, want true, false, true",
			active.GetBool(pt.sim), inactive.GetBool(pt.sim), debuffActive.GetBool(pt.sim))
	}
	if undeclaredActive.GetBool(pt.sim) || !undeclaredInactive.GetBool(pt.sim) {
		t.Errorf("undeclared in the batch window: active = %t, inactive = %t, want false, true", undeclaredActive.GetBool(pt.sim), undeclaredInactive.GetBool(pt.sim))
	}

	pt.runUntil(time.Second + SpellBatchWindow)
	if !pt.buff.IsActive() || pt.buff.pendingActivations != 0 || !active.GetBool(pt.sim) || inactive.GetBool(pt.sim) {
		t.Errorf("after the handler ran: buff up = %t, pending = %d, active = %t, inactive = %t, want true, 0, true, false",
			pt.buff.IsActive(), pt.buff.pendingActivations, active.GetBool(pt.sim), inactive.GetBool(pt.sim))
	}
}

// Each deferral counts until its handler runs, and the ones still waiting at the end of an iteration do not
// carry into the next.
func TestPendingProcCountsEachDeferralAndClearsAtTheIterationEnd(t *testing.T) {
	pt := newPendingProcTest()
	active := pt.isActive(pt.buff, false)
	pending := func() int32 { return pt.buff.pendingActivations }

	pt.runUntil(time.Second)
	pt.hit(pt.procSpell)
	pt.runUntil(time.Second + SpellBatchWindow/2)
	pt.hit(pt.procSpell)
	if pending() != 2 {
		t.Errorf("two handlers waiting: pending = %d, want 2", pending())
	}
	pt.runUntil(time.Second + SpellBatchWindow)
	if pending() != 1 {
		t.Errorf("one handler run, one waiting: pending = %d, want 1", pending())
	}

	pt.sim.Cleanup()
	pt.sim.Reset()
	if pending() != 0 || active.GetBool(pt.sim) {
		t.Errorf("next iteration: pending = %d, active = %t, want 0, false", pending(), active.GetBool(pt.sim))
	}

	pt.start()
	pt.runUntil(time.Second)
	pt.hit(pt.procSpell)
	pt.runUntil(time.Second + SpellBatchWindow)
	if pending() != 0 || !pt.buff.IsActive() {
		t.Errorf("a proc in the next iteration: pending = %d, buff up = %t, want 0, true", pending(), pt.buff.IsActive())
	}
}

// With the reaction time, a proc the player cannot count on stays a proc: one off a spell the APL did not cast,
// one off a major cooldown, which the cooldown manager may cast on its own, and one off an APL spell that rolls
// a stated chance or a proc manager. It is unseen while it waits and for the reaction time after it lands.
func TestReactionTimeHidesAProcThePlayerCannotCountOn(t *testing.T) {
	for _, c := range []struct {
		name  string
		spell func(*pendingProcTest) *Spell
		aura  func(*pendingProcTest) *Aura
	}{
		{"off the other spell", func(pt *pendingProcTest) *Spell { return pt.procSpell }, func(pt *pendingProcTest) *Aura { return pt.buff }},
		{"off a major cooldown", func(pt *pendingProcTest) *Spell { return pt.mcdSpell }, func(pt *pendingProcTest) *Aura { return pt.buff }},
		{"stated chance", func(pt *pendingProcTest) *Spell { return pt.aplSpell }, func(pt *pendingProcTest) *Aura { return pt.rolled }},
		{"proc manager", func(pt *pendingProcTest) *Spell { return pt.aplSpell }, func(pt *pendingProcTest) *Aura { return pt.managed }},
	} {
		pt := newPendingProcTest()
		aura := c.aura(pt)
		reacted := pt.isActive(aura, true)

		pt.runUntil(time.Second)
		pt.hit(c.spell(pt))
		if aura.pendingActivations != 1 {
			t.Fatalf("%s: pending = %d after the hit, want the proc to fire", c.name, aura.pendingActivations)
		}
		if reacted.GetBool(pt.sim) {
			t.Errorf("%s in the batch window: active = true, want false", c.name)
		}
		landed := time.Second + SpellBatchWindow
		pt.runUntil(landed + 50*time.Millisecond)
		if !aura.IsActive() || reacted.GetBool(pt.sim) {
			t.Errorf("%s inside the reaction time after landing: buff up = %t, active = %t, want true, false", c.name, aura.IsActive(), reacted.GetBool(pt.sim))
		}
		pt.runUntil(landed + 100*time.Millisecond)
		if !reacted.GetBool(pt.sim) {
			t.Errorf("%s at the reaction time after landing: active = false, want true", c.name)
		}
	}
}

// With the reaction time, a declared proc off a spell cast from the APL counts as the player's own: seen while it
// waits and as it lands, and for the reaction time after it runs out; Aura Is Inactive agrees while it waits. A
// cast trigger declaring a per-target aura marks nothing, having no target to find it on.
func TestReactionTimeCountsAProcOffAnAPLSpellAsOwn(t *testing.T) {
	pt := newPendingProcTest()
	reacted, debuffReacted := pt.isActive(pt.buff, true), pt.debuffIsActive(true)
	inactiveReacted := pt.isInactive(pt.buff, true)
	debuff := pt.debuffs.Get(pt.fw.CurrentTarget)

	pt.runUntil(time.Second)
	if !inactiveReacted.GetBool(pt.sim) {
		t.Errorf("before the cast: inactive = false, want true")
	}
	pt.aplSpell.Cast(pt.sim, pt.fw.CurrentTarget)
	if pt.castTriggers != 0 || debuff.pendingActivations != 1 {
		t.Errorf("after the cast: cast triggers run = %d, debuff pending = %d, want 0, 1", pt.castTriggers, debuff.pendingActivations)
	}
	if !reacted.GetBool(pt.sim) || !debuffReacted.GetBool(pt.sim) || inactiveReacted.GetBool(pt.sim) {
		t.Errorf("in the batch window: active = %t, debuff active = %t, inactive = %t, want true, true, false",
			reacted.GetBool(pt.sim), debuffReacted.GetBool(pt.sim), inactiveReacted.GetBool(pt.sim))
	}
	pt.runUntil(time.Second + SpellBatchWindow)
	if pt.castTriggers != 1 || !reacted.GetBool(pt.sim) || !debuffReacted.GetBool(pt.sim) {
		t.Errorf("as it lands: cast triggers run = %d, active = %t, debuff active = %t, want 1, true, true", pt.castTriggers, reacted.GetBool(pt.sim), debuffReacted.GetBool(pt.sim))
	}

	expires := pt.buff.ExpiresAt()
	pt.runUntil(expires)
	pt.runUntil(expires + 50*time.Millisecond)
	if pt.buff.IsActive() || !reacted.GetBool(pt.sim) {
		t.Errorf("inside the reaction time after running out: buff up = %t, active = %t, want false, true", pt.buff.IsActive(), reacted.GetBool(pt.sim))
	}
	pt.runUntil(expires + 100*time.Millisecond)
	if reacted.GetBool(pt.sim) {
		t.Errorf("at the reaction time after running out: active = true, want false")
	}
}

// With the reaction time, an immediate declared proc off a spell cast from the APL counts as the player's own as
// it lands, and one off a spell the APL did not cast stays a proc.
func TestReactionTimeCountsAnImmediateProcOffAnAPLSpellAsOwn(t *testing.T) {
	pt := newPendingProcTest()
	reacted := pt.isActive(pt.immediate, true)

	pt.runUntil(time.Second)
	pt.hit(pt.aplSpell)
	if !pt.immediate.IsActive() || !reacted.GetBool(pt.sim) {
		t.Errorf("off the APL spell, as it lands: buff up = %t, active = %t, want true, true", pt.immediate.IsActive(), reacted.GetBool(pt.sim))
	}

	pt.runUntil(pt.immediate.ExpiresAt() + time.Second)
	pt.hit(pt.procSpell)
	if !pt.immediate.IsActive() || reacted.GetBool(pt.sim) {
		t.Errorf("off the other spell, as it lands: buff up = %t, active = %t, want true, false", pt.immediate.IsActive(), reacted.GetBool(pt.sim))
	}
}

// With the reaction time, an aura something else put up stays a proc when an own trigger's handler refuses to
// put it up again, deferred or immediate.
func TestReactionTimeIgnoresAnOwnProcWhoseHandlerRefuses(t *testing.T) {
	pt := newPendingProcTest()
	cases := []struct {
		name    string
		aura    *Aura
		reacted APLValue
	}{
		{"deferred", pt.refused, pt.isActive(pt.refused, true)},
		{"immediate", pt.refusedNow, pt.isActive(pt.refusedNow, true)},
	}

	pt.runUntil(time.Second)
	for _, c := range cases {
		c.aura.Activate(pt.sim)
	}
	pt.runUntil(time.Second + 50*time.Millisecond)
	pt.hit(pt.aplSpell)
	pt.runUntil(time.Second + 50*time.Millisecond + SpellBatchWindow)
	for _, c := range cases {
		if !c.aura.IsActive() || c.aura.pendingActivations != 0 || c.reacted.GetBool(pt.sim) {
			t.Errorf("%s, after the handler refused inside the reaction time: buff up = %t, pending = %d, active = %t, want true, 0, false",
				c.name, c.aura.IsActive(), c.aura.pendingActivations, c.reacted.GetBool(pt.sim))
		}
	}

	pt.runUntil(time.Second + 100*time.Millisecond)
	for _, c := range cases {
		if !c.reacted.GetBool(pt.sim) {
			t.Errorf("%s at the reaction time after it went up: active = false, want true", c.name)
		}
	}
}

// With the reaction time, an aura something else put up counts as the player's own once an own trigger's handler
// adds a stack to it.
func TestReactionTimeCountsAnOwnProcThatOnlyAddsAStack(t *testing.T) {
	pt := newPendingProcTest()
	reacted := pt.isActive(pt.stacking, true)

	pt.runUntil(time.Second)
	pt.stacking.Activate(pt.sim)
	pt.stacking.SetStacks(pt.sim, 1)
	pt.runUntil(time.Second + 50*time.Millisecond)
	pt.hit(pt.aplSpell)
	pt.runUntil(time.Second + 50*time.Millisecond + SpellBatchWindow)
	if pt.stacking.GetStacks() != 2 || pt.stacking.pendingActivations != 0 || !reacted.GetBool(pt.sim) {
		t.Errorf("after the handler added a stack inside the reaction time: stacks = %d, pending = %d, active = %t, want 2, 0, true",
			pt.stacking.GetStacks(), pt.stacking.pendingActivations, reacted.GetBool(pt.sim))
	}
}

// A hit-taken trigger declaring a per-target aura marks the one on the attacker pending.
func TestPendingProcOffAHitTakenMarksTheAttacker(t *testing.T) {
	pt := newPendingProcTest()
	attacker := pt.fw.CurrentTarget
	debuff := pt.debuffs.Get(attacker)
	active := pt.debuffIsActive(false)

	pt.runUntil(time.Second)
	pt.fw.Unit.OnSpellHitTaken(pt.sim, &Spell{Unit: attacker, ProcMask: ProcMaskSpellDamage}, &SpellResult{Target: &pt.fw.Unit, Outcome: OutcomeHit})
	if debuff.IsActive() || debuff.pendingActivations != 1 || !active.GetBool(pt.sim) {
		t.Errorf("in the batch window: debuff up = %t, pending = %d, active = %t, want false, 1, true", debuff.IsActive(), debuff.pendingActivations, active.GetBool(pt.sim))
	}
	pt.runUntil(time.Second + SpellBatchWindow)
	if !debuff.IsActive() || debuff.pendingActivations != 0 {
		t.Errorf("after the handler ran: debuff up = %t, pending = %d, want true, 0", debuff.IsActive(), debuff.pendingActivations)
	}
}

// Inside the spell batch window a declared item stat proc reads as up to Any and All Item Stat Procs Active.
func TestItemStatProcsActiveSeeADeclaredProcBeforeItsHandlerRuns(t *testing.T) {
	pt := newPendingProcTest()
	str := int32(proto.Stat_StatStrength)
	anyActive := pt.rot.newValueAnyItemStatProcsActive(&proto.APLValueAnyItemStatProcsActive{StatType1: str, StatType2: -1, StatType3: -1}, nil)
	allActive := pt.rot.newValueAllItemStatProcsActive(&proto.APLValueAllItemStatProcsActive{StatType1: str, StatType2: -1, StatType3: -1}, nil)

	pt.runUntil(time.Second)
	if anyActive.GetBool(pt.sim) || allActive.GetBool(pt.sim) {
		t.Errorf("before the hit: any = %t, all = %t, want false, false", anyActive.GetBool(pt.sim), allActive.GetBool(pt.sim))
	}
	pt.hit(pt.procSpell)
	if pt.statProc.IsActive() || !anyActive.GetBool(pt.sim) || !allActive.GetBool(pt.sim) {
		t.Errorf("in the batch window: proc up = %t, any = %t, all = %t, want false, true, true", pt.statProc.IsActive(), anyActive.GetBool(pt.sim), allActive.GetBool(pt.sim))
	}
}
