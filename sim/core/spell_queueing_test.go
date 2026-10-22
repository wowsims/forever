package core

import (
	"slices"
	"testing"
	"time"
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
