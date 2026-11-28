package core

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A boost for the test: gained at start for 10 s, multiplying spirit regen by mult and adding share to
// the share that continues while casting.
type testRegenBoost struct {
	start time.Duration
	mult  float64
	share float64
}

// Runs 15 s at 10 mana per second of spirit regen with the boosts, casting throughout when casting
// is set, and answers the unit's own regen, what each boost recorded and the mana gained. fieldsFirst
// changes the fields before the boost is listed, as an exclusive category does.
func runRegenBoosts(boosts []testRegenBoost, casting bool, fieldsFirst bool) (own float64, added []float64, gained float64) {
	var auras []*Aura
	ct := newManaTickTest(func(char *Character) *Spell {
		for i, b := range boosts {
			aura := char.RegisterAura(Aura{Label: fmt.Sprintf("Regen Boost %d", i), ActionID: ActionID{SpellID: 910400 + int32(i)}, Duration: 10 * time.Second})
			var boost *ManaRegenBoost
			apply := func(sim *Simulation, on bool) {
				mult, share := b.mult, b.share
				if !on {
					mult, share = 1/b.mult, -b.share
				}
				boost.SpiritMult *= mult
				boost.CastingShare += share
				char.PseudoStats.SpiritRegenMultiplier *= mult
				char.PseudoStats.SpiritRegenRateCasting += share
				char.UpdateManaRegenRates(sim)
			}
			// Handlers run in the order they are added.
			if !fieldsFirst {
				boost = char.ManaRegenBoost(aura)
			}
			aura.ApplyOnGain(func(_ *Aura, sim *Simulation) { apply(sim, true) })
			aura.ApplyOnExpire(func(_ *Aura, sim *Simulation) { apply(sim, false) })
			if fieldsFirst {
				boost = char.ManaRegenBoost(aura)
			}
			auras = append(auras, aura)
		}
		return char.RegisterSpell(paidSpellConfig(910499, 0))
	})
	ct.start()
	unit := ct.fw
	unit.SpendMana(ct.sim, 5000, unit.NewManaMetrics(ActionID{SpellID: 910498}))
	before := unit.CurrentMana()
	if casting {
		unit.PseudoStats.FiveSecondRuleRefreshTime = time.Minute
	}

	for t := manaTestStep; t <= 15*time.Second; t += manaTestStep {
		ct.runUntil(t)
		for i, b := range boosts {
			if b.start == t {
				auras[i].Activate(ct.sim)
			}
		}
	}

	ownMetrics := unit.manaNotCastingMetrics
	if casting {
		ownMetrics = unit.manaCastingMetrics
	}
	for _, aura := range auras {
		added = append(added, unit.manaRegenBoosts[aura].metrics.Gain)
	}
	return ownMetrics.Gain, added, unit.CurrentMana() - before
}

// What each boost adds goes to its own metrics and the regen row keeps the unit's own regen, whichever
// order the boost's fields and its listing take.
func TestManaRegenBoostsTakeWhatTheyAdd(t *testing.T) {
	for _, c := range []struct {
		name    string
		boosts  []testRegenBoost
		casting bool
		own     float64
		added   []float64
	}{
		// 10 s at 5 times 10 mana a second add 400; the unit's own 15 s are 150.
		{"one boost", []testRegenBoost{{time.Second, 5, 0}}, false, 150, []float64{400}},
		// A boost that changes no field, gained while another is up, takes none of it.
		{"a boost that adds nothing", []testRegenBoost{{time.Second, 5, 0}, {2 * time.Second, 1, 0}}, false, 150, []float64{400, 0}},
		// From 2 s to 11 s both are up: 10 * 2 * 3 = 60 a second, of which the second takes 60 - 20.
		{"two boosts", []testRegenBoost{{time.Second, 2, 0}, {2 * time.Second, 3, 0}}, false, 150, []float64{1*10 + 9*10, 9*40 + 1*20}},
		// While casting, the unit keeps none of its spirit regen; the boost's full share is its own.
		{"casting share", []testRegenBoost{{time.Second, 1, 1}}, true, 0, []float64{100}},
	} {
		for _, fieldsFirst := range []bool{false, true} {
			own, added, gained := runRegenBoosts(c.boosts, c.casting, fieldsFirst)
			total := own
			for i, want := range c.added {
				total += added[i]
				if math.Abs(added[i]-want) > 1e-6 {
					t.Errorf("%s, fields first %v: boost %d recorded %v, want %v", c.name, fieldsFirst, i, added[i], want)
				}
			}
			if math.Abs(own-c.own) > 1e-6 {
				t.Errorf("%s, fields first %v: the regen row recorded %v, want %v", c.name, fieldsFirst, own, c.own)
			}
			if math.Abs(gained-total) > 1e-6 {
				t.Errorf("%s, fields first %v: the unit gained %v, but the rows add up to %v", c.name, fieldsFirst, gained, total)
			}
		}
	}
}

// Two auras of one spell, the player's copy and the external one, record under one boost row.
func TestManaRegenBoostsOfOneSpellShareTheirMetrics(t *testing.T) {
	var player, external *ManaRegenBoost
	newManaTickTest(func(char *Character) *Spell {
		player = char.ManaRegenBoost(char.RegisterAura(Aura{Label: "Boost (Player)", ActionID: ActionID{SpellID: 910400}}))
		external = char.ManaRegenBoost(char.RegisterAura(Aura{Label: "Boost (External)", ActionID: ActionID{SpellID: 910400, Tag: -1}}))
		return char.RegisterSpell(paidSpellConfig(910499, 0))
	})
	if player.metrics != external.metrics {
		t.Error("the player's and the external copy record under two rows")
	}
	if want := (ActionID{SpellID: 910400, Tag: manaRegenBoostTag}); player.metrics.ActionID != want || player.metrics.Type != proto.ResourceType_ResourceTypeMana {
		t.Errorf("the boost row is %v of %v", player.metrics.ActionID, player.metrics.Type)
	}
}
