package bulk

import (
	"fmt"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

const (
	scHeadWorn      int32 = 910600
	scHeadNewA      int32 = 910601
	scHeadNewB      int32 = 910602
	scCloakWorn     int32 = 910603
	scCloakNew      int32 = 910604
	scRingWornA     int32 = 910610
	scRingWornB     int32 = 910611
	scRingNew       int32 = 910612
	scRingLimitWorn int32 = 910613
	scRingLimitNew  int32 = 910614
	scRingUniqueNew int32 = 910615
	scRingUniqueRS  int32 = 910616 // Unique, rolled with a random suffix.
	scTrinketWornA  int32 = 910620
	scTrinketWornB  int32 = 910621
	scTrinketNew    int32 = 910622
)

const (
	scSwordWorn       int32 = 910630
	scDaggerWorn      int32 = 910631
	scSwordNew        int32 = 910632
	scMainHandNew     int32 = 910633
	scShieldWorn      int32 = 910634
	scShieldNew       int32 = 910635
	scTwoHandWorn     int32 = 910636
	scTwoHandNew      int32 = 910637
	scLimitDaggerNew  int32 = 910638
	scLimitDaggerWorn int32 = 910639
	scBowWorn         int32 = 910640
	scBowNew          int32 = 910641
)

const scLimitCategory int32 = 4343

const (
	scSuffixA int32 = 910630
	scSuffixB int32 = 910631
)

func registerSingleChangeFixtures() {
	item := func(id int32, itemType proto.ItemType) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           fmt.Sprintf("SingleChangeTest-%d", id),
			Type:           itemType,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
		}
	}
	limitRing := func(id int32) *proto.SimItem {
		simItem := item(id, proto.ItemType_ItemTypeFinger)
		simItem.LimitCategory = scLimitCategory
		return simItem
	}
	uniqueRing := func(id int32) *proto.SimItem {
		simItem := item(id, proto.ItemType_ItemTypeFinger)
		simItem.Unique = true
		return simItem
	}
	weapon := func(id int32, weaponType proto.WeaponType, handType proto.HandType) *proto.SimItem {
		simItem := item(id, proto.ItemType_ItemTypeWeapon)
		simItem.WeaponType = weaponType
		simItem.HandType = handType
		return simItem
	}
	limitDagger := func(id int32) *proto.SimItem {
		simItem := weapon(id, proto.WeaponType_WeaponTypeDagger, proto.HandType_HandTypeOneHand)
		simItem.LimitCategory = scLimitCategory
		return simItem
	}
	bow := func(id int32) *proto.SimItem {
		simItem := item(id, proto.ItemType_ItemTypeRanged)
		simItem.RangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeBow
		return simItem
	}

	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
		item(scHeadWorn, proto.ItemType_ItemTypeHead),
		item(scHeadNewA, proto.ItemType_ItemTypeHead),
		item(scHeadNewB, proto.ItemType_ItemTypeHead),
		item(scCloakWorn, proto.ItemType_ItemTypeBack),
		item(scCloakNew, proto.ItemType_ItemTypeBack),
		item(scRingWornA, proto.ItemType_ItemTypeFinger),
		item(scRingWornB, proto.ItemType_ItemTypeFinger),
		item(scRingNew, proto.ItemType_ItemTypeFinger),
		limitRing(scRingLimitWorn),
		limitRing(scRingLimitNew),
		uniqueRing(scRingUniqueNew),
		uniqueRing(scRingUniqueRS),
		item(scTrinketWornA, proto.ItemType_ItemTypeTrinket),
		item(scTrinketWornB, proto.ItemType_ItemTypeTrinket),
		item(scTrinketNew, proto.ItemType_ItemTypeTrinket),
		weapon(scSwordWorn, proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeOneHand),
		weapon(scDaggerWorn, proto.WeaponType_WeaponTypeDagger, proto.HandType_HandTypeOneHand),
		weapon(scSwordNew, proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeOneHand),
		weapon(scMainHandNew, proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeMainHand),
		weapon(scShieldWorn, proto.WeaponType_WeaponTypeShield, proto.HandType_HandTypeOffHand),
		weapon(scShieldNew, proto.WeaponType_WeaponTypeShield, proto.HandType_HandTypeOffHand),
		weapon(scTwoHandWorn, proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeTwoHand),
		weapon(scTwoHandNew, proto.WeaponType_WeaponTypeMace, proto.HandType_HandTypeTwoHand),
		limitDagger(scLimitDaggerNew),
		limitDagger(scLimitDaggerWorn),
		bow(scBowWorn),
		bow(scBowNew),
	}, RandomSuffixes: []*proto.ItemRandomSuffix{
		{Id: scSuffixA, Name: "of Test A"},
		{Id: scSuffixB, Name: "of Test B"},
	}})
}

// Builds the generator, checks the count and the indices against the candidates, and returns each
// candidate's gear.
func singleChangeGear(t *testing.T, spec proto.Spec, class proto.Class, isEnchanter bool, equipment core.Equipment, settings *proto.BulkSettings) (*bulkSimCandidateGenerator, []*proto.EquipmentSpec) {
	t.Helper()
	generator, err := newBulkSimCandidateGeneratorForSpec(settings, class, spec, isEnchanter, equipment)
	if err != nil {
		t.Fatalf("newBulkSimCandidateGeneratorForSpec: %v", err)
	}
	candidates := generator.buildCandidates()
	if count := generator.rawCombinationsCount(); count != len(candidates) {
		t.Fatalf("combination count %d, candidates %d", count, len(candidates))
	}
	gear := make([]*proto.EquipmentSpec, 0, len(candidates))
	for i, candidate := range candidates {
		if int(candidate.GetIndex()) != i {
			t.Fatalf("candidate %d has index %d", i, candidate.GetIndex())
		}
		gear = append(gear, candidate.GetGear())
	}
	return generator, gear
}

func changedSlots(base core.EquipmentSpec, gear *proto.EquipmentSpec) map[proto.ItemSlot]core.ItemSpec {
	changed := make(map[proto.ItemSlot]core.ItemSpec)
	for slot, spec := range core.ProtoToEquipmentSpec(gear) {
		if spec != base[slot] {
			changed[proto.ItemSlot(slot)] = spec
		}
	}
	return changed
}

type scChange struct {
	slot    proto.ItemSlot
	id      int32
	enchant int32
}

// Checks that each candidate changes exactly the one slot listed for it, in order. A negative
// enchant skips the enchant check.
func assertSingleChanges(t *testing.T, generator *bulkSimCandidateGenerator, gear []*proto.EquipmentSpec, want []scChange) {
	t.Helper()
	if len(gear) != len(want) {
		t.Fatalf("%d candidates, want %d", len(gear), len(want))
	}
	base := generator.baseGearSpec()
	for i, candidate := range gear {
		changed := changedSlots(base, candidate)
		spec, ok := changed[want[i].slot]
		if len(changed) != 1 || !ok {
			t.Fatalf("candidate %d changes %v, want only slot %s", i, changed, want[i].slot)
		}
		if spec.ID != want[i].id {
			t.Fatalf("candidate %d wears %d in slot %s, want %d", i, spec.ID, want[i].slot, want[i].id)
		}
		if want[i].enchant >= 0 && spec.Enchant != want[i].enchant {
			t.Fatalf("candidate %d enchants slot %s with %d, want %d", i, want[i].slot, spec.Enchant, want[i].enchant)
		}
	}
}

func scWornArmorRingsTrinkets() core.Equipment {
	return bulkEnchantEquipment(
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotHead, id: scHeadWorn},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotBack, id: scCloakWorn},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingWornA},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger2, id: scRingWornB},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotTrinket1, id: scTrinketWornA},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotTrinket2, id: scTrinketWornB},
	)
}

func TestSingleChangeCountIsASum(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	settings := &proto.BulkSettings{
		Items:        bulkEnchantSelection(scHeadNewA, scHeadNewB, scCloakNew, scRingNew, scTrinketNew),
		SingleChange: true,
	}

	generator, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, scWornArmorRingsTrinkets(), settings)

	// Two heads, a cloak, and a ring and a trinket in either slot; the cross product would be 3*2*3*3 = 54.
	if len(gear) != 7 {
		t.Fatalf("%d candidates, want 7", len(gear))
	}
	base := generator.baseGearSpec()
	for i, candidate := range gear {
		if changed := changedSlots(base, candidate); len(changed) != 1 {
			t.Fatalf("candidate %d changes %d slots, want 1: %v", i, len(changed), changed)
		}
	}
	assertDistinctGear(t, gear)
}

func TestSingleChangeRingsAndTrinketsTryBothSlots(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	settings := &proto.BulkSettings{
		Items:        bulkEnchantSelection(scRingNew, scRingUniqueNew, scTrinketNew),
		SingleChange: true,
	}

	generator, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, scWornArmorRingsTrinkets(), settings)

	assertSingleChanges(t, generator, gear, []scChange{
		{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingNew, enchant: -1},
		{slot: proto.ItemSlot_ItemSlotFinger2, id: scRingNew, enchant: -1},
		{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingUniqueNew, enchant: -1},
		{slot: proto.ItemSlot_ItemSlotFinger2, id: scRingUniqueNew, enchant: -1},
		{slot: proto.ItemSlot_ItemSlotTrinket1, id: scTrinketNew, enchant: -1},
		{slot: proto.ItemSlot_ItemSlotTrinket2, id: scTrinketNew, enchant: -1},
	})
}

// Putting the new ring on Finger2 would wear two rings of one limit category.
func TestSingleChangeLimitCategoryKeepsTheOtherSlot(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	equipment := bulkEnchantEquipment(
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingLimitWorn},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger2, id: scRingWornB},
	)
	settings := &proto.BulkSettings{Items: bulkEnchantSelection(scRingLimitNew), SingleChange: true}

	generator, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, equipment, settings)

	assertSingleChanges(t, generator, gear, []scChange{
		{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingLimitNew, enchant: -1},
	})
}

// Another suffix of a worn unique ring is a new item, and it cannot go on beside the worn copy.
func TestSingleChangeUniqueKeepsTheOtherSlot(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	var equipment core.Equipment
	equipment[proto.ItemSlot_ItemSlotFinger1] = core.NewItem(core.ItemSpec{ID: scRingUniqueRS, RandomSuffix: scSuffixA})
	equipment[proto.ItemSlot_ItemSlotFinger2] = core.NewItem(core.ItemSpec{ID: scRingWornB})
	settings := &proto.BulkSettings{
		Items:        []*proto.ItemSpec{{Id: scRingUniqueRS, RandomSuffix: scSuffixB}},
		SingleChange: true,
	}

	generator, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, equipment, settings)

	assertSingleChanges(t, generator, gear, []scChange{
		{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingUniqueRS, enchant: -1},
	})
	if suffix := gear[0].GetItems()[proto.ItemSlot_ItemSlotFinger1].GetRandomSuffix(); suffix != scSuffixB {
		t.Fatalf("Finger1 wears suffix %d, want %d", suffix, scSuffixB)
	}
}

func TestSingleChangeSkipsWornItems(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	equipment := bulkEnchantEquipment(
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotBack, id: scCloakWorn, enchant: bulkEnchantBackA},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingWornA},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger2, id: scRingWornB},
	)
	settings := &proto.BulkSettings{
		Items: []*proto.ItemSpec{
			{Id: scCloakWorn, Enchant: bulkEnchantBackB},
			{Id: scRingWornA},
			{Id: scRingWornB, Enchant: bulkEnchantRingA},
		},
		Enchants:     []*proto.BulkEnchantSlotOptions{bulkEnchantPool(proto.ItemSlot_ItemSlotBack, bulkEnchantBackA, bulkEnchantBackB)},
		SingleChange: true,
	}

	generator, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, equipment, settings)

	if len(gear) != 0 || generator.rawCombinationsCount() != 0 {
		t.Fatalf("%d candidates, count %d; want none for a selection of worn items", len(gear), generator.rawCombinationsCount())
	}
}

func TestSingleChangePoolsOnlyTheNewItem(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	equipment := bulkEnchantEquipment(
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotBack, id: scCloakWorn, enchant: bulkEnchantBackA},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotHead, id: scHeadWorn},
	)
	settings := &proto.BulkSettings{
		Items:        bulkEnchantSelection(scCloakNew, scHeadNewA),
		Enchants:     []*proto.BulkEnchantSlotOptions{bulkEnchantPool(proto.ItemSlot_ItemSlotBack, bulkEnchantBackA, bulkEnchantBackB)},
		SingleChange: true,
	}

	generator, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, equipment, settings)

	assertSingleChanges(t, generator, gear, []scChange{
		{slot: proto.ItemSlot_ItemSlotBack, id: scCloakNew, enchant: bulkEnchantBackA},
		{slot: proto.ItemSlot_ItemSlotBack, id: scCloakNew, enchant: bulkEnchantBackB},
		{slot: proto.ItemSlot_ItemSlotHead, id: scHeadNewA, enchant: 0},
	})
	for i, candidate := range gear {
		back := candidate.GetItems()[proto.ItemSlot_ItemSlotBack]
		if back.GetId() == scCloakWorn && back.GetEnchant() != bulkEnchantBackA {
			t.Fatalf("candidate %d re-enchants the worn cloak with %d", i, back.GetEnchant())
		}
	}
}

func TestSingleChangeWithoutPoolCarriesTheSlotEnchant(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	equipment := bulkEnchantEquipment(
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotBack, id: scCloakWorn, enchant: bulkEnchantBackA},
	)
	settings := &proto.BulkSettings{Items: bulkEnchantSelection(scCloakNew), SingleChange: true}

	generator, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, equipment, settings)

	assertSingleChanges(t, generator, gear, []scChange{
		{slot: proto.ItemSlot_ItemSlotBack, id: scCloakNew, enchant: bulkEnchantBackA},
	})
}

func TestSingleChangeRingPoolNeedsEnchanting(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	equipment := bulkEnchantEquipment(
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingWornA},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger2, id: scRingWornB},
	)
	settings := &proto.BulkSettings{
		Items:        bulkEnchantSelection(scRingNew),
		Enchants:     []*proto.BulkEnchantSlotOptions{bulkEnchantPool(proto.ItemSlot_ItemSlotFinger1, bulkEnchantRingA, bulkEnchantRingB)},
		SingleChange: true,
	}

	enchanter, enchanterGear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, equipment, settings)
	assertSingleChanges(t, enchanter, enchanterGear, []scChange{
		{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingNew, enchant: bulkEnchantRingA},
		{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingNew, enchant: bulkEnchantRingB},
		{slot: proto.ItemSlot_ItemSlotFinger2, id: scRingNew, enchant: bulkEnchantRingA},
		{slot: proto.ItemSlot_ItemSlotFinger2, id: scRingNew, enchant: bulkEnchantRingB},
	})

	nonEnchanter, nonEnchanterGear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, false, equipment, settings)
	assertSingleChanges(t, nonEnchanter, nonEnchanterGear, []scChange{
		{slot: proto.ItemSlot_ItemSlotFinger1, id: scRingNew, enchant: 0},
		{slot: proto.ItemSlot_ItemSlotFinger2, id: scRingNew, enchant: 0},
	})
	for i, candidate := range nonEnchanterGear {
		for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2} {
			if enchant := candidate.GetItems()[slot].GetEnchant(); enchant != 0 {
				t.Fatalf("candidate %d wears ring enchant %d in slot %s without Enchanting", i, enchant, slot)
			}
		}
	}
}

func TestSingleChangeEntryPointsAgree(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	items := make([]*proto.ItemSpec, core.NumItemSlots)
	for slot := range items {
		items[slot] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotBack] = &proto.ItemSpec{Id: scCloakWorn}
	items[proto.ItemSlot_ItemSlotFinger1] = &proto.ItemSpec{Id: scRingWornA}
	items[proto.ItemSlot_ItemSlotFinger2] = &proto.ItemSpec{Id: scRingWornB}
	items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: scTrinketWornA}
	items[proto.ItemSlot_ItemSlotTrinket2] = &proto.ItemSpec{Id: scTrinketWornB}
	equipment := &proto.EquipmentSpec{Items: items}
	baseRequest := &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Class:     proto.Class_ClassWarrior,
			Race:      proto.Race_RaceOrc,
			Equipment: equipment,
			Spec:      &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}},
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
	}
	settings := &proto.BulkSettings{
		Items:        bulkEnchantSelection(scCloakNew, scRingNew, scTrinketNew),
		Enchants:     []*proto.BulkEnchantSlotOptions{bulkEnchantPool(proto.ItemSlot_ItemSlotBack, bulkEnchantBackA, bulkEnchantBackB)},
		SingleChange: true,
	}

	count := BulkCombinationCount(&proto.BulkCombinationCountRequest{BaseRequest: baseRequest, BulkSettings: settings})
	if count.GetError() != nil {
		t.Fatalf("BulkCombinationCount: %s", count.GetError().GetMessage())
	}
	candidatesRequest := &proto.BulkCandidatesRequest{BaseRequest: baseRequest, BulkSettings: settings}
	first := BulkCandidates(candidatesRequest)
	second := BulkCandidates(candidatesRequest)
	for _, result := range []*proto.BulkCandidatesResult{first, second} {
		if result.GetError() != nil {
			t.Fatalf("BulkCandidates: %s", result.GetError().GetMessage())
		}
	}

	if count.GetCombinations() != 6 || count.GetRawCombinations() != 6 || first.GetRawCombinations() != 6 || len(first.GetCandidates()) != 6 {
		t.Fatalf("count %d/%d, candidates %d/%d; want 6 everywhere",
			count.GetCombinations(), count.GetRawCombinations(), first.GetRawCombinations(), len(first.GetCandidates()))
	}
	if len(second.GetCandidates()) != len(first.GetCandidates()) {
		t.Fatalf("second call returns %d candidates, the first %d", len(second.GetCandidates()), len(first.GetCandidates()))
	}
	baseKey := bulkEnchantGearKey(equipment)
	for i, candidate := range first.GetCandidates() {
		key := bulkEnchantGearKey(candidate.GetGear())
		if other := bulkEnchantGearKey(second.GetCandidates()[i].GetGear()); other != key {
			t.Fatalf("candidate %d is %s on one call and %s on the other", i, key, other)
		}
		if key == baseKey {
			t.Fatalf("candidate %d is the worn gear", i)
		}
	}
}

// The main-hand and off-hand ids of each candidate.
func weaponPairs(gear []*proto.EquipmentSpec) [][2]int32 {
	pairs := make([][2]int32, 0, len(gear))
	for _, candidate := range gear {
		items := candidate.GetItems()
		pairs = append(pairs, [2]int32{items[proto.ItemSlot_ItemSlotMainHand].GetId(), items[proto.ItemSlot_ItemSlotOffHand].GetId()})
	}
	return pairs
}

// The main-hand id and enchant, then the off-hand id and enchant, of each candidate.
func weaponSpecs(gear []*proto.EquipmentSpec) [][4]int32 {
	specs := make([][4]int32, 0, len(gear))
	for _, candidate := range gear {
		mainHand := candidate.GetItems()[proto.ItemSlot_ItemSlotMainHand]
		offHand := candidate.GetItems()[proto.ItemSlot_ItemSlotOffHand]
		specs = append(specs, [4]int32{mainHand.GetId(), mainHand.GetEnchant(), offHand.GetId(), offHand.GetEnchant()})
	}
	return specs
}

func scWornHands(mainHand int32, offHand int32) core.Equipment {
	return bulkEnchantEquipment(
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotMainHand, id: mainHand},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotOffHand, id: offHand},
	)
}

func TestSingleChangeWeaponPlacements(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	const (
		n  = scSwordNew
		m  = scMainHandNew
		h  = scShieldNew
		tw = scTwoHandNew
		s  = scSwordWorn
		d  = scDaggerWorn
		sh = scShieldWorn
		w  = scTwoHandWorn
	)
	type player struct {
		spec  proto.Spec
		class proto.Class
	}
	warrior := player{proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior}
	paladin := player{proto.Spec_SpecProtectionPaladin, proto.Class_ClassPaladin}
	rows := []struct {
		name     string
		player   player
		mainHand int32
		offHand  int32
		selected int32
		want     [][2]int32
	}{
		{"dual wield, one-hander goes into either hand", warrior, s, d, n, [][2]int32{{n, d}, {n, s}, {s, n}, {d, n}}},
		{"dual wield, main-hand-only weapon", warrior, s, d, m, [][2]int32{{m, d}, {m, s}}},
		{"dual wield, shield", warrior, s, d, h, [][2]int32{{s, h}, {d, h}}},
		{"dual wield, two-hander", warrior, s, d, tw, [][2]int32{{tw, 0}}},
		{"dual wield over a shield, the shield stays off hand", warrior, s, sh, n, [][2]int32{{n, sh}, {n, s}, {s, n}}},
		{"dual wield under a two-hander, one-hander", warrior, w, sh, n, nil},
		{"dual wield under a two-hander, main-hand-only weapon", warrior, w, sh, m, nil},
		{"dual wield under a two-hander, two-hander", warrior, w, sh, tw, [][2]int32{{tw, 0}}},
		{"dual wield under a two-hander, shield", warrior, w, sh, h, nil},
		{"no dual wield, one-hander", paladin, s, sh, n, [][2]int32{{n, sh}}},
		{"no dual wield, two-hander", paladin, s, sh, tw, [][2]int32{{tw, 0}}},
		{"no dual wield, shield", paladin, s, sh, h, [][2]int32{{s, h}}},
		{"no dual wield, dagger the class cannot use", paladin, s, sh, d, nil},
		{"no dual wield under a two-hander, one-hander", paladin, w, sh, n, nil},
		{"no dual wield under a two-hander, shield", paladin, w, sh, h, nil},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			settings := &proto.BulkSettings{Items: bulkEnchantSelection(row.selected), SingleChange: true}

			_, gear := singleChangeGear(t, row.player.spec, row.player.class, true, scWornHands(row.mainHand, row.offHand), settings)

			if got := weaponPairs(gear); !slices.Equal(got, row.want) {
				t.Fatalf("weapon pairs %v, want %v", got, row.want)
			}
			assertDistinctGear(t, gear)
		})
	}
}

// The new dagger cannot be worn beside the worn one of its limit category, in either hand.
func TestSingleChangeLimitCategoryAcrossHands(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	settings := &proto.BulkSettings{Items: bulkEnchantSelection(scLimitDaggerNew), SingleChange: true}

	_, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, scWornHands(scSwordWorn, scLimitDaggerWorn), settings)

	want := [][2]int32{{scLimitDaggerNew, scSwordWorn}, {scSwordWorn, scLimitDaggerNew}}
	if got := weaponPairs(gear); !slices.Equal(got, want) {
		t.Fatalf("weapon pairs %v, want %v", got, want)
	}
}

func scWornEnchantedSwordAndDagger() core.Equipment {
	return bulkEnchantEquipment(
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotMainHand, id: scSwordWorn, enchant: bulkEnchantWeaponKept},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotOffHand, id: scDaggerWorn},
	)
}

func TestSingleChangeShuffledWeaponKeepsItsEnchant(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	settings := &proto.BulkSettings{Items: bulkEnchantSelection(scSwordNew), SingleChange: true}

	_, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, scWornEnchantedSwordAndDagger(), settings)

	want := [][4]int32{
		{scSwordNew, bulkEnchantWeaponKept, scDaggerWorn, 0},
		{scSwordNew, bulkEnchantWeaponKept, scSwordWorn, bulkEnchantWeaponKept},
		{scSwordWorn, bulkEnchantWeaponKept, scSwordNew, 0},
		{scDaggerWorn, 0, scSwordNew, 0},
	}
	if got := weaponSpecs(gear); !slices.Equal(got, want) {
		t.Fatalf("weapons %v, want %v", got, want)
	}
}

func TestSingleChangeWeaponPoolsPerHand(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	settings := &proto.BulkSettings{
		Items: bulkEnchantSelection(scSwordNew),
		Enchants: []*proto.BulkEnchantSlotOptions{
			bulkEnchantPool(proto.ItemSlot_ItemSlotMainHand, bulkEnchantWeaponA),
			bulkEnchantPool(proto.ItemSlot_ItemSlotOffHand, bulkEnchantWeaponB),
		},
		SingleChange: true,
	}

	_, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, scWornEnchantedSwordAndDagger(), settings)

	want := [][4]int32{
		{scSwordNew, bulkEnchantWeaponA, scDaggerWorn, 0},
		{scSwordNew, bulkEnchantWeaponA, scSwordWorn, bulkEnchantWeaponKept},
		{scSwordWorn, bulkEnchantWeaponKept, scSwordNew, bulkEnchantWeaponB},
		{scDaggerWorn, 0, scSwordNew, bulkEnchantWeaponB},
	}
	if got := weaponSpecs(gear); !slices.Equal(got, want) {
		t.Fatalf("weapons %v, want %v", got, want)
	}
}

func TestSingleChangeRangedTakesTheRangedSlot(t *testing.T) {
	registerBulkEnchantFixtures()
	registerSingleChangeFixtures()
	equipment := bulkEnchantEquipment(bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotRanged, id: scBowWorn})
	settings := &proto.BulkSettings{Items: bulkEnchantSelection(scBowNew), SingleChange: true}

	generator, gear := singleChangeGear(t, proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, true, equipment, settings)

	assertSingleChanges(t, generator, gear, []scChange{
		{slot: proto.ItemSlot_ItemSlotRanged, id: scBowNew, enchant: 0},
	})
}
