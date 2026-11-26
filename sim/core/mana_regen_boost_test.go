package core

import (
	"math"
	"testing"
	"time"
)

// What a boost adds to the regen goes to its own metrics and the regen rows keep the unit's own,
// whether the boost's fields change before its aura counts as up or after. Spirit regen is 10 mana
// per second and the boost multiplies it by 5 for 10 s, so the unit's own regen over 15 s is 150 and
// the boost's is 400.
func TestManaRegenBoostTakesWhatItAdds(t *testing.T) {
	for _, fieldsFirst := range []bool{false, true} {
		var aura *Aura
		ct := newManaTickTest(func(char *Character) *Spell {
			aura = char.RegisterAura(Aura{Label: "Regen Boost", ActionID: ActionID{SpellID: 910400}, Duration: 10 * time.Second})
			var boost *ManaRegenBoost
			apply := func(sim *Simulation, factor float64) {
				boost.MultiplySpiritRegen(factor)
				char.PseudoStats.SpiritRegenMultiplier *= factor
				char.UpdateManaRegenRates(sim)
			}
			// Handlers run in the order they are added, so adding the field changes first changes the
			// fields before the boost is listed, as an exclusive category does.
			if !fieldsFirst {
				boost = char.ManaRegenBoost(aura)
			}
			aura.ApplyOnGain(func(_ *Aura, sim *Simulation) { apply(sim, 5) })
			aura.ApplyOnExpire(func(_ *Aura, sim *Simulation) { apply(sim, 1.0/5) })
			if fieldsFirst {
				boost = char.ManaRegenBoost(aura)
			}
			return char.RegisterSpell(paidSpellConfig(910401, 0))
		})
		ct.start()
		ct.fw.SpendMana(ct.sim, 5000, ct.fw.manaNotCastingMetrics)

		ct.runUntil(time.Second)
		aura.Activate(ct.sim)
		ct.runUntil(15 * time.Second)

		boost := ct.fw.manaRegenBoosts[aura].metrics
		own := ct.fw.manaNotCastingMetrics.Gain + 5000
		if math.Abs(own-150) > 1e-6 || math.Abs(boost.Gain-400) > 1e-6 {
			t.Errorf("fields first %v: own regen %v and boost %v, want 150 and 400", fieldsFirst, own, boost.Gain)
		}
		if got, want := ct.fw.CurrentMana(), 10000-5000+550.0; math.Abs(got-want) > 1e-6 {
			t.Errorf("fields first %v: %v mana, want %v", fieldsFirst, got, want)
		}
		if !boost.ActionID.SameAction(ActionID{SpellID: 910400, Tag: ManaRegenBoostTag}) {
			t.Errorf("the boost records under %v", boost.ActionID)
		}
	}
}
