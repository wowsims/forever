package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Thunderbrew's Boot Flask 744's on-use 12257 and Cloak of Fire 14134's 18364 each keep an aura on the
// wearer for the row's duration that casts the spell it triggers every period. Each tick deals the
// triggered row's amount to all three targets.
func TestPeriodicDamageOnUseTicksOnEveryTarget(t *testing.T) {
	for _, tc := range []struct {
		name     string
		itemID   int32
		spellID  int32
		cooldown time.Duration
	}{
		{"Thunderbrew's Boot Flask", 991701, 12257, 5 * time.Minute},
		{"Cloak of Fire", 991702, 18364, 2 * time.Minute},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row := spelldata.MustFind(tc.spellID)
			periodic := row.FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL)
			period := periodic.Period()
			ticks := int(row.Duration() / period)
			damageID := periodic.TriggerID
			amount := spelldata.MustFind(damageID).DamageEffect()
			if ticks < 2 || amount.Min() != amount.Max() || amount.Average() <= 0 {
				t.Fatalf("%d ticks %d times for %v to %v, want at least 2 ticks of one amount", tc.spellID, ticks, amount.Min(), amount.Max())
			}

			sim, caster := newOnUseSimAgainst(t, NewSpellDataPeriodicDamageOnUse,
				map[int32]*proto.ItemEffect{tc.itemID: testOnUse(tc.spellID, int32(tc.cooldown.Milliseconds()), 0, 0)}, &proto.ItemSpec{}, 3)
			onUse := onUseSpell(t, caster, tc.itemID)
			targets := sim.Encounter.ActiveTargetUnits

			if got := caster.GetInitialMajorCooldown(onUse.ActionID); !got.Type.Matches(core.CooldownTypeDPS) {
				t.Errorf("the on-use is a major cooldown of type %v, want a DPS one", got.Type)
			}

			if !onUse.Cast(sim, targets[0]) {
				t.Fatalf("%s could not be used", tc.name)
			}
			start := sim.CurrentTime
			damage := caster.GetSpell(core.ActionID{SpellID: damageID})
			if damage == nil {
				t.Fatalf("the on-use registered no spell for %d", damageID)
			}
			aura := onUse.SelfHot()
			if aura == nil || !aura.IsActive() {
				t.Fatalf("using %s put no aura on the wearer", tc.name)
			}

			dealt := func(i int) float64 {
				metrics := damage.SpellMetrics[targets[i].UnitIndex]
				if metrics.Misses+metrics.ResistedHits > 0 {
					t.Fatalf("%d missed target %d", damageID, i)
				}
				return metrics.TotalDamage
			}
			for i := range targets {
				if got := dealt(i); got != 0 {
					t.Errorf("target %d took %v when the item was used, want nothing until the first tick", i, got)
				}
			}

			for tick := 1; tick <= ticks; tick++ {
				stepPast(t, sim, start+time.Duration(tick)*period+time.Millisecond)
				for i := range targets {
					if got, want := dealt(i), float64(tick)*amount.Average(); got != want {
						t.Errorf("after tick %d target %d took %v in all, want %v", tick, i, got, want)
					}
				}
			}
			if aura.IsActive() {
				t.Errorf("the aura is still up after its %d ticks", ticks)
			}

			stepPast(t, sim, start+time.Duration(ticks+1)*period+time.Millisecond)
			for i := range targets {
				if got, want := dealt(i), float64(ticks)*amount.Average(); got != want {
					t.Errorf("target %d took %v once the aura was gone, want the %v of its ticks", i, got, want)
				}
			}

			if got := onUse.CD.ReadyAt(); got != start+tc.cooldown {
				t.Errorf("%s is ready again at %v, want %v after its use at %v", tc.name, got, tc.cooldown, start)
			}
		})
	}
}
