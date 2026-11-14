package core

import "testing"

// A cost stated as a share of base mana rounds to the nearest whole mana, up as well as down.
func TestBaseManaShareCostRoundsToNearest(t *testing.T) {
	for _, c := range []struct {
		baseMana float64
		percent  float64
		want     int32
	}{
		{914, 35, 320},
		{899, 35, 315},
		{869, 35, 304},
		{899, 55, 494},
	} {
		ct := newManaTickTest(func(char *Character) *Spell {
			char.BaseMana = c.baseMana
			config := paidSpellConfig(910400, 0)
			config.ManaCost = ManaCostOptions{BaseCostPercent: c.percent}
			return char.RegisterSpell(config)
		})
		if got := ct.spell.Cost.BaseCost; got != c.want {
			t.Errorf("%v%% of %v base mana costs %d, want %d", c.percent, c.baseMana, got, c.want)
		}
	}
}
