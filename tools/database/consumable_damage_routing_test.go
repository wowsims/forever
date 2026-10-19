package database

import (
	"testing"

	"github.com/wowsims/forever/tools/database/dbc"
)

func TestADamageOverTimeOnAnAreaIsRoutedAsDamage(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	radiationBomb := instance.Consumables[215168]
	if routing := RouteConsumable(&radiationBomb, instance); !routing.Supported() || routing.Kind != ConsumableDamage {
		t.Errorf("Ez-Thro Radiation Bomb 215168 routes as %q (%s), want damage with 435968's damage over time on an area",
			routing.Kind, routing.Reason())
	}

	highYield := instance.Consumables[215127]
	if routing := RouteConsumable(&highYield, instance); !routing.Supported() || routing.Kind != ConsumableDamage {
		t.Errorf("High-Yield Radiation Bomb 215127 routes as %q (%s), want damage", routing.Kind, routing.Reason())
	}
}

// A potion that damages enemies registers through the potion builder, which runs it.
func TestADamagePotionRegistersAsAPotion(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		id   int
		name string
	}{
		{250951, "Caustic Smog Potion"},
		{250953, "Dragonfire Potion"},
	} {
		catalogued := instance.Consumables[tc.id]
		routing := RouteConsumable(&catalogued, instance)
		shipped := ShippedConsumable(&catalogued, routing, instance)
		if family := registrationFamilyOf(shipped, &catalogued, routing); family != potionFamily {
			t.Errorf("%d %s registers in family %d, want the potions", tc.id, tc.name, family)
		}
		if unsupported := onUseUnsupported(shipped); unsupported != nil {
			t.Errorf("%d %s is refused by the potion builder: %v", tc.id, tc.name, unsupported)
		}
	}
}
