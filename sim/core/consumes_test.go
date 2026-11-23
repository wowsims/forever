package core

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestSelectedConsumableIDs(t *testing.T) {
	const (
		pot int32 = 990411 + iota
		potion
		flask
		food
		explosive
		mhImbue
		mhWeaponImbue
		ohImbue
		conjured
		extra1
		extra2
		extra3
		extra4
		swapMhImbue
		swapOhWeaponImbue
		resourcePot
		defensivePot
	)
	consumes := &proto.ConsumesSpec{
		ResourcePotId:       resourcePot,
		DefensivePotId:      defensivePot,
		OffensivePotId:      pot,
		Potions:             []int32{pot, potion},
		FlaskId:             flask,
		FoodId:              food,
		ExplosiveId:         explosive,
		MhImbueId:           mhImbue,
		MhWeaponImbueId:     mhWeaponImbue,
		OhImbueId:           ohImbue,
		SwapMhImbueId:       swapMhImbue,
		SwapOhImbueId:       ohImbue,
		SwapOhWeaponImbueId: swapOhWeaponImbue,
		ConjuredItems:       []int32{conjured},
		ExtraIds:            []int32{extra1, extra2, extra3, extra4},
	}

	got := selectedConsumableIDs(consumes)
	want := []int32{resourcePot, defensivePot, pot, flask, food, explosive, ohImbue, mhImbue, mhWeaponImbue, swapOhWeaponImbue, swapMhImbue,
		extra1, extra2, extra3, extra4, potion, conjured}
	if !slices.Equal(got, want) {
		t.Errorf("selected %v, want %v", got, want)
	}
}

func TestSelectedConsumableIDsKeepTheFirstOfEachGroup(t *testing.T) {
	const mongoose, agility, scroll, scrollIV, roids, zanza, otherScroll = 990401, 990402, 990403, 990404, 990405, 990406, 990407
	addToDatabase(&proto.SimDatabase{Consumables: []*proto.Consumable{
		{Id: mongoose, Type: proto.ConsumableType_ConsumableTypeElixir},
		{Id: agility, Type: proto.ConsumableType_ConsumableTypeElixir},
		{Id: scroll, Type: proto.ConsumableType_ConsumableTypeScroll, GroupKey: "29/1"},
		{Id: scrollIV, Type: proto.ConsumableType_ConsumableTypeScroll, GroupKey: "29/1"},
		{Id: roids, Type: proto.ConsumableType_ConsumableTypeOther, GroupKey: "Zanza"},
		{Id: zanza, Type: proto.ConsumableType_ConsumableTypeOther, GroupKey: "Zanza"},
		{Id: otherScroll, Type: proto.ConsumableType_ConsumableTypeScroll, GroupKey: "29/2"},
	}})

	got := selectedConsumableIDs(&proto.ConsumesSpec{ElixirIds: []int32{mongoose, agility}, ExtraIds: []int32{scroll, scrollIV, roids, zanza, otherScroll}})
	if want := []int32{mongoose, agility, scroll, roids, otherScroll}; !slices.Equal(got, want) {
		t.Errorf("selected %v, want %v", got, want)
	}
}

func TestHandWrittenConsumableEffectReplacesDataDriven(t *testing.T) {
	const potion = 990430
	addToDatabase(&proto.SimDatabase{Consumables: []*proto.Consumable{{Id: potion}}})
	unregister := func() {
		delete(consumableEffects, potion)
		delete(dataDrivenConsumableEffects, potion)
	}
	t.Cleanup(unregister)

	for _, handFirst := range []bool{true, false} {
		applied := ""
		hand := func() { NewConsumableEffect(potion, func(Agent) { applied = "hand" }) }
		dataDriven := func() { NewSpellDataConsumableEffect(potion, func(Agent) { applied = "rows" }) }
		if handFirst {
			hand()
			dataDriven()
		} else {
			dataDriven()
			hand()
		}

		if !HasConsumableEffect(potion) {
			t.Fatalf("hand first %v: %d has no effect registered", handFirst, potion)
		}
		apply, _ := consumableEffect(potion)
		apply(nil)
		if applied != "hand" {
			t.Errorf("hand first %v: the %s effect applied, want the hand-written one", handFirst, applied)
		}
		unregister()
	}
}

func TestASecondEffectOfTheSameOriginPanics(t *testing.T) {
	const potion = 990431
	addToDatabase(&proto.SimDatabase{Consumables: []*proto.Consumable{{Id: potion}}})
	for name, register := range map[string]func(int32, ApplyEffect){
		"hand-written": NewConsumableEffect,
		"data-driven":  NewSpellDataConsumableEffect,
	} {
		t.Run(name, func(t *testing.T) {
			t.Cleanup(func() {
				delete(consumableEffects, potion)
				delete(dataDrivenConsumableEffects, potion)
			})

			register(potion, func(Agent) {})
			defer func() {
				if recover() == nil {
					t.Errorf("a second %s effect for one consumable registered without a panic", name)
				}
			}()
			register(potion, func(Agent) {})
		})
	}
}
