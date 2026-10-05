package bulk

import (
	"fmt"
	"sort"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Enchant and item-rule helpers.

func addBulkTestEnchant(effectID int32, itemType proto.ItemType, enchantType proto.EnchantType) int32 {
	core.AddToDatabase(&proto.SimDatabase{
		Enchants: []*proto.SimEnchant{{EffectId: effectID, Type: itemType, EnchantType: enchantType}},
	})
	return effectID
}

func TestBulkSimEnchantAppliesToItem_UsesWeaponTypeRules(t *testing.T) {
	weaponEffectID := int32(910001)

	addBulkTestEnchant(weaponEffectID, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeNormal)

	twoHandSword := core.Item{
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeSword,
		HandType:   proto.HandType_HandTypeTwoHand,
	}
	oneHandSword := core.Item{
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeSword,
		HandType:   proto.HandType_HandTypeOneHand,
	}
	staff := core.Item{
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeStaff,
		HandType:   proto.HandType_HandTypeTwoHand,
	}
	shield := core.Item{
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeShield,
		HandType:   proto.HandType_HandTypeOffHand,
	}
	offHand := core.Item{
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeOffHand,
		HandType:   proto.HandType_HandTypeOffHand,
	}

	if !enchantAppliesToItem(weaponEffectID, twoHandSword) {
		t.Fatalf("expected weapon enchant to apply to two-handed weapon")
	}
	if !enchantAppliesToItem(weaponEffectID, oneHandSword) {
		t.Fatalf("expected weapon enchant to apply to one-handed weapon")
	}
	if !enchantAppliesToItem(weaponEffectID, staff) {
		t.Fatalf("expected weapon enchant to apply to staff")
	}
	if enchantAppliesToItem(weaponEffectID, shield) {
		t.Fatalf("expected weapon enchant to not apply to shield")
	}
	if enchantAppliesToItem(weaponEffectID, offHand) {
		t.Fatalf("expected weapon enchant to not apply to off-hand frill")
	}
}

func TestBulkSimEnchantAppliesToItem_UsesEnchantTypeRules(t *testing.T) {
	twoHand := addBulkTestEnchant(910010, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeTwoHand)
	shieldOnly := addBulkTestEnchant(910011, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeShield)
	staffOnly := addBulkTestEnchant(910012, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeStaff)
	offHandOnly := addBulkTestEnchant(910013, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeOffHand)

	oneHandSword := core.Item{Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand}
	twoHandSword := core.Item{Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeTwoHand}
	staff := core.Item{Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand}
	mace := core.Item{Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeTwoHand}
	shield := core.Item{Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand}
	frill := core.Item{Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeOffHand, HandType: proto.HandType_HandTypeOffHand}

	cases := []struct {
		name    string
		enchant int32
		item    core.Item
		want    bool
	}{
		{"two-hand enchant on two-hander", twoHand, twoHandSword, true},
		{"two-hand enchant on one-hander", twoHand, oneHandSword, false},
		{"shield enchant on shield", shieldOnly, shield, true},
		{"shield enchant on weapon", shieldOnly, oneHandSword, false},
		{"staff enchant on staff", staffOnly, staff, true},
		{"staff enchant on two-hand mace", staffOnly, mace, false},
		{"off-hand enchant on off-hand frill", offHandOnly, frill, true},
		{"off-hand enchant on shield", offHandOnly, shield, false},
	}
	for _, c := range cases {
		if got := enchantAppliesToItem(c.enchant, c.item); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestBulkSimEnchantAppliesToItem_AnyRowFits(t *testing.T) {
	sharedEffectID := addBulkTestEnchant(910020, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeTwoHand)
	addBulkTestEnchant(sharedEffectID, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeNormal)

	oneHandSword := core.Item{Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand}
	twoHandSword := core.Item{Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeTwoHand}
	shield := core.Item{Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand}

	if !enchantAppliesToItem(sharedEffectID, oneHandSword) {
		t.Errorf("expected the one-hand row to fit a one-hander")
	}
	if !enchantAppliesToItem(sharedEffectID, twoHandSword) {
		t.Errorf("expected the enchant to fit a two-hander")
	}
	if enchantAppliesToItem(sharedEffectID, shield) {
		t.Errorf("expected neither row to fit a shield")
	}
}

func TestBulkSimEnchantAppliesToItem_UsesTypedRangedRules(t *testing.T) {
	rangedEffectID := int32(910005)
	weaponEffectID := int32(910006)

	addBulkTestEnchant(rangedEffectID, proto.ItemType_ItemTypeRanged, proto.EnchantType_EnchantTypeNormal)
	addBulkTestEnchant(weaponEffectID, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeNormal)

	bow := core.Item{
		Type:             proto.ItemType_ItemTypeRanged,
		RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
	}
	wand := core.Item{
		Type:             proto.ItemType_ItemTypeRanged,
		RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
	}
	gun := core.Item{
		Type:             proto.ItemType_ItemTypeRanged,
		RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeGun,
	}

	if !enchantAppliesToItem(rangedEffectID, bow) {
		t.Fatalf("expected ranged enchant to apply to bow")
	}
	if enchantAppliesToItem(rangedEffectID, wand) {
		t.Fatalf("expected ranged enchant to not apply to wand")
	}
	if enchantAppliesToItem(weaponEffectID, gun) {
		t.Fatalf("expected non-ranged enchant to not apply to non-wand ranged weapon")
	}
}

func TestBulkSimEnchantAppliesToItem_RejectsNonMatchingItemTypes(t *testing.T) {
	extraTypeEffectID := int32(910007)
	addBulkTestEnchant(extraTypeEffectID, proto.ItemType_ItemTypeChest, proto.EnchantType_EnchantTypeNormal)

	wrist := core.Item{Type: proto.ItemType_ItemTypeWrist}
	legs := core.Item{Type: proto.ItemType_ItemTypeLegs}
	chest := core.Item{Type: proto.ItemType_ItemTypeChest}

	if !enchantAppliesToItem(extraTypeEffectID, chest) {
		t.Fatalf("expected enchant to apply to matching item type")
	}
	if enchantAppliesToItem(extraTypeEffectID, wrist) {
		t.Fatalf("expected chest enchant to not apply to wrist")
	}
	if enchantAppliesToItem(extraTypeEffectID, legs) {
		t.Fatalf("expected enchant to not apply to unrelated item type")
	}
}

func oneHandOption(id int32) bulkSimCandidateOption {
	return bulkSimCandidateOption{
		spec: &proto.ItemSpec{Id: id},
		item: core.Item{
			ID:         id,
			Type:       proto.ItemType_ItemTypeWeapon,
			HandType:   proto.HandType_HandTypeOneHand,
			WeaponType: proto.WeaponType_WeaponTypeSword,
		},
	}
}

func countSameWeaponBothHands(combos [][2]*bulkSimCandidateOption, id int32) int {
	matches := 0
	for _, combo := range combos {
		if combo[0] == nil || combo[1] == nil {
			continue
		}
		if combo[0].spec.GetId() == id && combo[1].spec.GetId() == id {
			matches++
		}
	}
	return matches
}

// Two copies of the same 1H weapon (one equipped, one user-added) must allow wielding it in
// both hands; a single copy must not. This pins the dual-wield stacking behaviour so the
// candidate-option dedup cannot silently collapse it.
func TestGetAllWeaponCombosSameWeaponBothHandsNeedsTwoCopies(t *testing.T) {
	testCases := []struct {
		name       string
		copies     int
		wantCombos int
	}{
		{name: "one copy", copies: 1, wantCombos: 0},
		{name: "two copies", copies: 2, wantCombos: 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			options := make([]bulkSimCandidateOption, 0, testCase.copies)
			for range testCase.copies {
				options = append(options, oneHandOption(100))
			}
			generator := &bulkSimCandidateGenerator{
				playerCanDualWield: true,
				selectedByBulkSlot: map[BulkSimItemSlot][]bulkSimCandidateOption{
					BulkSimItemSlotHandWeapon: options,
				},
			}
			generator.initCopyCounts()

			got := countSameWeaponBothHands(generator.getAllWeaponCombos(), 100)
			if got != testCase.wantCombos {
				t.Fatalf("same-weapon-both-hands combos for %d copies = %d, want %d", testCase.copies, got, testCase.wantCombos)
			}
		})
	}
}

// Distinct 1H weapons must still pair in both orders regardless of copy counts.
func TestGetAllWeaponCombosPairsDistinctWeaponsBothOrders(t *testing.T) {
	generator := &bulkSimCandidateGenerator{
		playerCanDualWield: true,
		selectedByBulkSlot: map[BulkSimItemSlot][]bulkSimCandidateOption{
			BulkSimItemSlotHandWeapon: {oneHandOption(100), oneHandOption(200)},
		},
	}
	generator.initCopyCounts()

	forward, backward := 0, 0
	for _, combo := range generator.getAllWeaponCombos() {
		if combo[0] == nil || combo[1] == nil {
			continue
		}
		if combo[0].spec.GetId() == 100 && combo[1].spec.GetId() == 200 {
			forward++
		}
		if combo[0].spec.GetId() == 200 && combo[1].spec.GetId() == 100 {
			backward++
		}
	}
	if forward != 1 || backward != 1 {
		t.Fatalf("distinct weapon pairings: forward=%d backward=%d, want 1 and 1", forward, backward)
	}
}

func addBulkTestWeapon(id int32, handType proto.HandType, weaponType proto.WeaponType) {
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{
			{
				Id:             id,
				Type:           proto.ItemType_ItemTypeWeapon,
				WeaponType:     weaponType,
				HandType:       handType,
				ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
			},
		},
	})
}

// A 2H mainhand combo carries a nil offhand, so the base gear's offhand has to be cleared:
// without the reset the equipped shield stays wielded alongside the 2H weapon.
func TestBuildGearForComboClearsOffHandForTwoHandMainHand(t *testing.T) {
	const twoHandID, oneHandID, shieldID = int32(910201), int32(910202), int32(910203)
	addBulkTestWeapon(twoHandID, proto.HandType_HandTypeTwoHand, proto.WeaponType_WeaponTypeSword)
	addBulkTestWeapon(oneHandID, proto.HandType_HandTypeOneHand, proto.WeaponType_WeaponTypeSword)
	addBulkTestWeapon(shieldID, proto.HandType_HandTypeOffHand, proto.WeaponType_WeaponTypeShield)

	var baseEquipment core.Equipment
	baseEquipment[proto.ItemSlot_ItemSlotMainHand] = core.NewItem(core.ItemSpec{ID: oneHandID})
	baseEquipment[proto.ItemSlot_ItemSlotOffHand] = core.NewItem(core.ItemSpec{ID: shieldID})

	generator := &bulkSimCandidateGenerator{
		baseEquipment: baseEquipment,
		selectedByBulkSlot: map[BulkSimItemSlot][]bulkSimCandidateOption{
			BulkSimItemSlotMainHand: {{
				spec: &proto.ItemSpec{Id: twoHandID},
				item: core.NewItem(core.ItemSpec{ID: twoHandID}),
			}},
		},
	}
	generator.initCopyCounts()
	generator.weaponCombos = generator.getAllWeaponCombos()

	if len(generator.weaponCombos) != 1 {
		t.Fatalf("expected 1 weapon combo, got %d", len(generator.weaponCombos))
	}

	gear, err := generator.buildGearForCombo(0)
	if err != nil {
		t.Fatalf("buildGearForCombo failed: %v", err)
	}
	items := gear.GetItems()
	if got := items[proto.ItemSlot_ItemSlotMainHand].GetId(); got != twoHandID {
		t.Fatalf("expected 2H weapon in mainhand, got %d", got)
	}
	if got := items[proto.ItemSlot_ItemSlotOffHand].GetId(); got != 0 {
		t.Fatalf("expected offhand to be cleared for 2H mainhand, got %d", got)
	}
}

// TBC has no Titan's Grip, so a 2H weapon never pairs with an offhand for any spec, DPS
// warriors included.
func TestGetAllWeaponCombosNeverPairsTwoHandWithOffHand(t *testing.T) {
	const firstTwoHandID, secondTwoHandID = int32(910301), int32(910302)
	addBulkTestWeapon(firstTwoHandID, proto.HandType_HandTypeTwoHand, proto.WeaponType_WeaponTypeSword)
	addBulkTestWeapon(secondTwoHandID, proto.HandType_HandTypeTwoHand, proto.WeaponType_WeaponTypeAxe)

	twoHandOption := func(id int32) bulkSimCandidateOption {
		return bulkSimCandidateOption{
			spec: &proto.ItemSpec{Id: id},
			item: core.NewItem(core.ItemSpec{ID: id}),
		}
	}

	generator := &bulkSimCandidateGenerator{
		playerSpec:         proto.Spec_SpecDpsWarrior,
		playerCanDualWield: true,
		selectedByBulkSlot: map[BulkSimItemSlot][]bulkSimCandidateOption{
			BulkSimItemSlotHandWeapon: {twoHandOption(firstTwoHandID), twoHandOption(secondTwoHandID)},
		},
	}
	generator.initCopyCounts()

	combos := generator.getAllWeaponCombos()
	if len(combos) != 2 {
		t.Fatalf("expected 2 weapon combos (one per 2H weapon), got %d", len(combos))
	}
	for _, combo := range combos {
		if combo[1] != nil {
			t.Fatalf("expected nil offhand for 2H combo, got item %d", combo[1].spec.GetId())
		}
	}
}

// Per-spec mainhand/offhand combination matrix.

// Feeds every spec a pool of synthetic weapons covering all (weaponType, handType) shapes the
// class can equip, then asserts the generated MH/OH pairs are all wearable in game. Run with
// -v to print the per-spec combination table.

var comboTableWeaponTypes = []proto.WeaponType{
	proto.WeaponType_WeaponTypeAxe,
	proto.WeaponType_WeaponTypeDagger,
	proto.WeaponType_WeaponTypeFist,
	proto.WeaponType_WeaponTypeMace,
	proto.WeaponType_WeaponTypeOffHand,
	proto.WeaponType_WeaponTypePolearm,
	proto.WeaponType_WeaponTypeShield,
	proto.WeaponType_WeaponTypeStaff,
	proto.WeaponType_WeaponTypeSword,
}

var comboTableHandTypes = []proto.HandType{
	proto.HandType_HandTypeMainHand,
	proto.HandType_HandTypeOneHand,
	proto.HandType_HandTypeOffHand,
	proto.HandType_HandTypeTwoHand,
}

// (weaponType, handType) pairs that actually occur in assets/database/db.json, so the synthetic
// pool cannot invent shapes like a one-handed off-hand frill.
var comboTableRealPairs = map[proto.WeaponType][]proto.HandType{
	proto.WeaponType_WeaponTypeAxe:     {proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand, proto.HandType_HandTypeTwoHand},
	proto.WeaponType_WeaponTypeDagger:  {proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand},
	proto.WeaponType_WeaponTypeFist:    {proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand},
	proto.WeaponType_WeaponTypeMace:    {proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand, proto.HandType_HandTypeTwoHand},
	proto.WeaponType_WeaponTypeOffHand: {proto.HandType_HandTypeOffHand},
	proto.WeaponType_WeaponTypePolearm: {proto.HandType_HandTypeTwoHand},
	proto.WeaponType_WeaponTypeShield:  {proto.HandType_HandTypeOffHand},
	proto.WeaponType_WeaponTypeStaff:   {proto.HandType_HandTypeTwoHand},
	proto.WeaponType_WeaponTypeSword:   {proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand, proto.HandType_HandTypeTwoHand},
}

func comboTableItemID(wt proto.WeaponType, ht proto.HandType, copyIdx int) int32 {
	return 990000 + int32(wt)*100 + int32(ht)*10 + int32(copyIdx)
}

func comboTableItemName(wt proto.WeaponType, ht proto.HandType, copyIdx int) string {
	wtNames := map[proto.WeaponType]string{
		proto.WeaponType_WeaponTypeAxe:     "Axe",
		proto.WeaponType_WeaponTypeDagger:  "Dagger",
		proto.WeaponType_WeaponTypeFist:    "Fist",
		proto.WeaponType_WeaponTypeMace:    "Mace",
		proto.WeaponType_WeaponTypeOffHand: "Frill",
		proto.WeaponType_WeaponTypePolearm: "Polearm",
		proto.WeaponType_WeaponTypeShield:  "Shield",
		proto.WeaponType_WeaponTypeStaff:   "Staff",
		proto.WeaponType_WeaponTypeSword:   "Sword",
	}
	htNames := map[proto.HandType]string{
		proto.HandType_HandTypeMainHand: "MHonly",
		proto.HandType_HandTypeOneHand:  "1H",
		proto.HandType_HandTypeOffHand:  "OHonly",
		proto.HandType_HandTypeTwoHand:  "2H",
	}
	return fmt.Sprintf("%s-%s#%d", wtNames[wt], htNames[ht], copyIdx)
}

func registerComboTableItems() {
	items := make([]*proto.SimItem, 0, len(comboTableWeaponTypes)*len(comboTableHandTypes)*2)
	for _, wt := range comboTableWeaponTypes {
		for _, ht := range comboTableRealPairs[wt] {
			for copyIdx := range 2 {
				items = append(items, &proto.SimItem{
					Id:             comboTableItemID(wt, ht, copyIdx),
					Name:           comboTableItemName(wt, ht, copyIdx),
					Type:           proto.ItemType_ItemTypeWeapon,
					WeaponType:     wt,
					HandType:       ht,
					ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
				})
			}
		}
	}
	core.AddToDatabase(&proto.SimDatabase{Items: items})
}

var comboTableSpecs = []struct {
	spec  proto.Spec
	class proto.Class
	label string
}{
	{proto.Spec_SpecBalanceDruid, proto.Class_ClassDruid, "Balance Druid"},
	{proto.Spec_SpecFeralCatDruid, proto.Class_ClassDruid, "Feral Cat Druid"},
	{proto.Spec_SpecFeralBearDruid, proto.Class_ClassDruid, "Feral Bear Druid"},
	{proto.Spec_SpecRestorationDruid, proto.Class_ClassDruid, "Restoration Druid"},
	{proto.Spec_SpecHunter, proto.Class_ClassHunter, "Hunter"},
	{proto.Spec_SpecMage, proto.Class_ClassMage, "Mage"},
	{proto.Spec_SpecHolyPaladin, proto.Class_ClassPaladin, "Holy Paladin"},
	{proto.Spec_SpecProtectionPaladin, proto.Class_ClassPaladin, "Protection Paladin"},
	{proto.Spec_SpecRetributionPaladin, proto.Class_ClassPaladin, "Retribution Paladin"},
	{proto.Spec_SpecDpsPriest, proto.Class_ClassPriest, "Priest"},
	{proto.Spec_SpecHealerPriest, proto.Class_ClassPriest, "Healer Priest"},
	{proto.Spec_SpecRogue, proto.Class_ClassRogue, "Rogue"},
	{proto.Spec_SpecElementalShaman, proto.Class_ClassShaman, "Elemental Shaman"},
	{proto.Spec_SpecEnhancementShaman, proto.Class_ClassShaman, "Enhancement Shaman"},
	{proto.Spec_SpecRestorationShaman, proto.Class_ClassShaman, "Restoration Shaman"},
	{proto.Spec_SpecWarlock, proto.Class_ClassWarlock, "Warlock"},
	{proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, "DPS Warrior"},
	{proto.Spec_SpecProtectionWarrior, proto.Class_ClassWarrior, "Protection Warrior"},
}

// classComboTableItems returns every synthetic weapon the class can actually equip, so the
// selection fed to the generator mirrors what the item search would let through.
func classComboTableItems(class proto.Class) []*proto.ItemSpec {
	specs := make([]*proto.ItemSpec, 0)
	for _, wt := range comboTableWeaponTypes {
		if _, ok := core.ClassWeaponTypeCapabilities[class][wt]; !ok {
			continue
		}
		for _, ht := range comboTableRealPairs[wt] {
			for copyIdx := range 2 {
				specs = append(specs, &proto.ItemSpec{Id: comboTableItemID(wt, ht, copyIdx)})
			}
		}
	}
	return specs
}

func comboLabel(option *bulkSimCandidateOption) string {
	if option == nil {
		return "(keeps base gear)"
	}
	return option.item.Name
}

func TestWeaponComboMatrixPerSpec(t *testing.T) {
	registerComboTableItems()

	for _, testCase := range comboTableSpecs {
		selected := classComboTableItems(testCase.class)
		generator := &bulkSimCandidateGenerator{
			settings:           &proto.BulkSettings{Items: selected},
			playerClass:        testCase.class,
			playerSpec:         testCase.spec,
			playerCanDualWield: core.SpecCanDualWieldCapabilities[testCase.spec],
			baseEquipment:      core.Equipment{},
			selectedByBulkSlot: make(map[BulkSimItemSlot][]bulkSimCandidateOption),
			groupedPairsBySlot: make(map[BulkSimItemSlot][][2]bulkSimCandidateOption),
			frozenItems:        make(map[BulkSimItemSlot]*core.Item),
			weaponTypeFilters:  make(map[proto.ItemSlot][]proto.WeaponType),
		}
		if err := generator.initSelectedItems(); err != nil {
			t.Fatalf("%s: initSelectedItems: %v", testCase.label, err)
		}

		shapes := make(map[string]int)
		illegal := make([]string, 0)
		for _, combo := range generator.getAllWeaponCombos() {
			mh, oh := comboLabel(combo[0]), comboLabel(combo[1])
			shapes[shapeOf(combo)]++
			if isIllegalCombo(combo) {
				illegal = append(illegal, fmt.Sprintf("%s + %s", mh, oh))
			}
		}

		shapeKeys := make([]string, 0, len(shapes))
		for key := range shapes {
			shapeKeys = append(shapeKeys, key)
		}
		sort.Strings(shapeKeys)

		t.Logf("=== %s (dualWield=%v)", testCase.label, generator.playerCanDualWield)
		t.Logf("    bulkSlots: MainHand=%d OffHand=%d HandWeapon=%d",
			len(generator.selectedByBulkSlot[BulkSimItemSlotMainHand]),
			len(generator.selectedByBulkSlot[BulkSimItemSlotOffHand]),
			len(generator.selectedByBulkSlot[BulkSimItemSlotHandWeapon]))
		for _, key := range shapeKeys {
			t.Logf("    %-22s x%d", key, shapes[key])
		}
		if len(illegal) > 0 {
			t.Errorf("%s: %d unwearable combos, e.g. %v", testCase.label, len(illegal), illegal[:min(len(illegal), 8)])
		}
		if !generator.playerCanDualWield {
			for _, combo := range generator.getAllWeaponCombos() {
				if combo[0] != nil && combo[1] != nil && combo[1].item.WeaponType != proto.WeaponType_WeaponTypeShield &&
					combo[1].item.WeaponType != proto.WeaponType_WeaponTypeOffHand {
					t.Errorf("%s cannot dual wield but got %s + %s", testCase.label, comboLabel(combo[0]), comboLabel(combo[1]))
					break
				}
			}
		}
	}
}

func handTypeShape(option *bulkSimCandidateOption) string {
	if option == nil {
		return "none"
	}
	switch option.item.HandType {
	case proto.HandType_HandTypeMainHand:
		return "MHonly"
	case proto.HandType_HandTypeOneHand:
		return "1H"
	case proto.HandType_HandTypeOffHand:
		if option.item.WeaponType == proto.WeaponType_WeaponTypeShield {
			return "Shield"
		}
		return "OHonly"
	case proto.HandType_HandTypeTwoHand:
		return "2H"
	}
	return "unknown"
}

func shapeOf(combo [2]*bulkSimCandidateOption) string {
	return fmt.Sprintf("MH=%s OH=%s", handTypeShape(combo[0]), handTypeShape(combo[1]))
}

// Any pairing WoW itself would reject: a 2H in either hand alongside another weapon, an
// off-hand-only item in the mainhand, or a mainhand-only item in the offhand.
func isIllegalCombo(combo [2]*bulkSimCandidateOption) bool {
	mh, oh := combo[0], combo[1]
	if mh != nil && oh != nil {
		if mh.item.HandType == proto.HandType_HandTypeTwoHand || oh.item.HandType == proto.HandType_HandTypeTwoHand {
			return true
		}
	}
	if mh != nil && mh.item.HandType == proto.HandType_HandTypeOffHand {
		return true
	}
	if oh != nil && oh.item.HandType == proto.HandType_HandTypeMainHand {
		return true
	}
	return false
}

// Per-spec paired-slot (ring / trinket / both hands) matrix.

// Covers the paired bulk slots (ring, trinket, both hands) for every spec: which same-slot
// pairings the candidate generator offers when the batch holds two copies of one item, two items
// sharing a limit category, unique items, and plain distinct items. Run with -v to print the
// table.

const (
	pairRingPlainID   int32 = 992001
	pairRingOtherID   int32 = 992002
	pairRingUniqueID  int32 = 992003
	pairRingLimitAID  int32 = 992004
	pairRingLimitBID  int32 = 992005
	pairTrinketPlain  int32 = 992011
	pairTrinketOther  int32 = 992012
	pairTrinketUnique int32 = 992013
	pairWeaponPlainID int32 = 992021
	pairWeaponUniqID  int32 = 992022
)

const pairLimitCategory int32 = 4242

func registerPairedSlotItems() {
	item := func(id int32, itemType proto.ItemType, unique bool, limitCategory int32) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           fmt.Sprintf("PairTest-%d", id),
			Type:           itemType,
			Unique:         unique,
			LimitCategory:  limitCategory,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
		}
	}
	weapon := func(id int32, unique bool) *proto.SimItem {
		simItem := item(id, proto.ItemType_ItemTypeWeapon, unique, 0)
		simItem.WeaponType = proto.WeaponType_WeaponTypeDagger
		simItem.HandType = proto.HandType_HandTypeOneHand
		return simItem
	}

	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
		item(pairRingPlainID, proto.ItemType_ItemTypeFinger, false, 0),
		item(pairRingOtherID, proto.ItemType_ItemTypeFinger, false, 0),
		item(pairRingUniqueID, proto.ItemType_ItemTypeFinger, true, 0),
		item(pairRingLimitAID, proto.ItemType_ItemTypeFinger, false, pairLimitCategory),
		item(pairRingLimitBID, proto.ItemType_ItemTypeFinger, false, pairLimitCategory),
		item(pairTrinketPlain, proto.ItemType_ItemTypeTrinket, false, 0),
		item(pairTrinketOther, proto.ItemType_ItemTypeTrinket, false, 0),
		item(pairTrinketUnique, proto.ItemType_ItemTypeTrinket, true, 0),
		weapon(pairWeaponPlainID, false),
		weapon(pairWeaponUniqID, true),
	}})
}

// Two copies of everything that may legitimately be doubled, one copy of the rest.
func pairedSlotSelection() []*proto.ItemSpec {
	ids := []int32{
		pairRingPlainID, pairRingPlainID,
		pairRingOtherID,
		pairRingUniqueID, pairRingUniqueID,
		pairRingLimitAID, pairRingLimitBID,
		pairTrinketPlain, pairTrinketPlain,
		pairTrinketOther,
		pairTrinketUnique, pairTrinketUnique,
		pairWeaponPlainID, pairWeaponPlainID,
		pairWeaponUniqID, pairWeaponUniqID,
	}
	specs := make([]*proto.ItemSpec, 0, len(ids))
	for _, id := range ids {
		specs = append(specs, &proto.ItemSpec{Id: id})
	}
	return specs
}

type pairKey struct {
	first  int32
	second int32
}

func newPairKey(first int32, second int32) pairKey {
	if second < first {
		first, second = second, first
	}
	return pairKey{first: first, second: second}
}

func TestPairedSlotMatrixPerSpec(t *testing.T) {
	registerPairedSlotItems()

	cases := []struct {
		label string
		pair  pairKey
		want  bool
	}{
		{label: "ring: two copies of a plain item", pair: newPairKey(pairRingPlainID, pairRingPlainID), want: true},
		{label: "ring: one copy of a plain item", pair: newPairKey(pairRingOtherID, pairRingOtherID), want: false},
		{label: "ring: two copies of a unique item", pair: newPairKey(pairRingUniqueID, pairRingUniqueID), want: false},
		{label: "ring: two items sharing a limit category", pair: newPairKey(pairRingLimitAID, pairRingLimitBID), want: false},
		{label: "ring: two distinct plain items", pair: newPairKey(pairRingPlainID, pairRingOtherID), want: true},
		{label: "ring: distinct plain + unique items", pair: newPairKey(pairRingPlainID, pairRingUniqueID), want: true},
		{label: "trinket: two copies of a plain item", pair: newPairKey(pairTrinketPlain, pairTrinketPlain), want: true},
		{label: "trinket: two copies of a unique item", pair: newPairKey(pairTrinketUnique, pairTrinketUnique), want: false},
		{label: "trinket: two distinct plain items", pair: newPairKey(pairTrinketPlain, pairTrinketOther), want: true},
	}

	for _, testCase := range comboTableSpecs {
		generator := &bulkSimCandidateGenerator{
			settings:           &proto.BulkSettings{Items: pairedSlotSelection()},
			playerClass:        testCase.class,
			playerSpec:         testCase.spec,
			playerCanDualWield: core.SpecCanDualWieldCapabilities[testCase.spec],
			baseEquipment:      core.Equipment{},
			selectedByBulkSlot: make(map[BulkSimItemSlot][]bulkSimCandidateOption),
			groupedPairsBySlot: make(map[BulkSimItemSlot][][2]bulkSimCandidateOption),
			frozenItems:        make(map[BulkSimItemSlot]*core.Item),
			weaponTypeFilters:  make(map[proto.ItemSlot][]proto.WeaponType),
		}
		if err := generator.initSelectedItems(); err != nil {
			t.Fatalf("%s: initSelectedItems: %v", testCase.label, err)
		}
		generator.initGroupedSlotPairs()

		got := make(map[pairKey]bool)
		for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
			for _, pair := range generator.groupedPairsBySlot[bulkSlot] {
				got[newPairKey(pair[0].spec.GetId(), pair[1].spec.GetId())] = true
			}
		}
		// Both hands are a paired slot too, just built by getAllWeaponCombos instead.
		weaponPairs := make(map[pairKey]bool)
		for _, combo := range generator.getAllWeaponCombos() {
			if combo[0] == nil || combo[1] == nil {
				continue
			}
			weaponPairs[newPairKey(combo[0].spec.GetId(), combo[1].spec.GetId())] = true
		}

		t.Logf("=== %s (dualWield=%v)", testCase.label, generator.playerCanDualWield)
		for _, expectation := range cases {
			if got[expectation.pair] != expectation.want {
				t.Errorf("%s: %s = %v, want %v", testCase.label, expectation.label, got[expectation.pair], expectation.want)
			}
			t.Logf("    %-46s %v", expectation.label, got[expectation.pair])
		}

		// A weapon may fill both hands only for a dual-wielder, and never when it is unique.
		wantSameWeapon := generator.playerCanDualWield
		samePlain := weaponPairs[newPairKey(pairWeaponPlainID, pairWeaponPlainID)]
		sameUnique := weaponPairs[newPairKey(pairWeaponUniqID, pairWeaponUniqID)]
		if samePlain != wantSameWeapon {
			t.Errorf("%s: two copies of a plain 1H weapon in both hands = %v, want %v", testCase.label, samePlain, wantSameWeapon)
		}
		if sameUnique {
			t.Errorf("%s: two copies of a unique 1H weapon wielded in both hands", testCase.label)
		}
		t.Logf("    %-46s %v", "weapon: two copies of a plain 1H item", samePlain)
		t.Logf("    %-46s %v", "weapon: two copies of a unique 1H item", sameUnique)

		pairCounts := make([]string, 0, 3)
		for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
			pairCounts = append(pairCounts, fmt.Sprintf("%s=%d", BulkSimItemSlotNames[bulkSlot], len(generator.groupedPairsBySlot[bulkSlot])))
		}
		sort.Strings(pairCounts)
		t.Logf("    pairs: %v, weaponCombos=%d", pairCounts, len(generator.getAllWeaponCombos()))
	}
}

// Enchant pools.

const (
	bulkEnchantBackA       int32 = 910100
	bulkEnchantBackB       int32 = 910101
	bulkEnchantRingA       int32 = 910102
	bulkEnchantRingB       int32 = 910103
	bulkEnchantRingKept    int32 = 910104
	bulkEnchantWeaponA     int32 = 910105
	bulkEnchantWeaponB     int32 = 910106
	bulkEnchantWeaponKept  int32 = 910107
	bulkEnchantTwoHandOnly int32 = 910108
	bulkEnchantTwoRows     int32 = 910109 // A 2H row and a one-hand row under one effect id, 2H loaded first.
	bulkEnchantUnknown     int32 = 910199
)

const (
	bulkEnchantCloakID    int32 = 910500
	bulkEnchantCloakAltID int32 = 910501
	bulkEnchantRingRID    int32 = 910502
	bulkEnchantRingSID    int32 = 910503
	bulkEnchantRingTID    int32 = 910504
	bulkEnchantDaggerAID  int32 = 910505
	bulkEnchantDaggerBID  int32 = 910506
	bulkEnchantDaggerCID  int32 = 910507
	bulkEnchantSwordAID   int32 = 910508
	bulkEnchantSwordBID   int32 = 910509
)

func registerBulkEnchantFixtures() {
	item := func(id int32, itemType proto.ItemType) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           fmt.Sprintf("EnchantTest-%d", id),
			Type:           itemType,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
		}
	}
	weapon := func(id int32, weaponType proto.WeaponType) *proto.SimItem {
		simItem := item(id, proto.ItemType_ItemTypeWeapon)
		simItem.WeaponType = weaponType
		simItem.HandType = proto.HandType_HandTypeOneHand
		return simItem
	}
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{
			item(bulkEnchantCloakID, proto.ItemType_ItemTypeBack),
			item(bulkEnchantCloakAltID, proto.ItemType_ItemTypeBack),
			item(bulkEnchantRingRID, proto.ItemType_ItemTypeFinger),
			item(bulkEnchantRingSID, proto.ItemType_ItemTypeFinger),
			item(bulkEnchantRingTID, proto.ItemType_ItemTypeFinger),
			weapon(bulkEnchantDaggerAID, proto.WeaponType_WeaponTypeDagger),
			weapon(bulkEnchantDaggerBID, proto.WeaponType_WeaponTypeDagger),
			weapon(bulkEnchantDaggerCID, proto.WeaponType_WeaponTypeDagger),
			weapon(bulkEnchantSwordAID, proto.WeaponType_WeaponTypeSword),
			weapon(bulkEnchantSwordBID, proto.WeaponType_WeaponTypeSword),
		},
	})
	addBulkTestEnchant(bulkEnchantBackA, proto.ItemType_ItemTypeBack, proto.EnchantType_EnchantTypeNormal)
	addBulkTestEnchant(bulkEnchantBackB, proto.ItemType_ItemTypeBack, proto.EnchantType_EnchantTypeNormal)
	addBulkTestEnchant(bulkEnchantRingA, proto.ItemType_ItemTypeFinger, proto.EnchantType_EnchantTypeNormal)
	addBulkTestEnchant(bulkEnchantRingB, proto.ItemType_ItemTypeFinger, proto.EnchantType_EnchantTypeNormal)
	addBulkTestEnchant(bulkEnchantRingKept, proto.ItemType_ItemTypeFinger, proto.EnchantType_EnchantTypeNormal)
	addBulkTestEnchant(bulkEnchantWeaponA, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeNormal)
	addBulkTestEnchant(bulkEnchantWeaponB, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeNormal)
	addBulkTestEnchant(bulkEnchantWeaponKept, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeNormal)
	addBulkTestEnchant(bulkEnchantTwoHandOnly, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeTwoHand)
	addBulkTestEnchant(bulkEnchantTwoRows, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeTwoHand)
	addBulkTestEnchant(bulkEnchantTwoRows, proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeNormal)
}

type bulkEnchantEquipped struct {
	slot    proto.ItemSlot
	id      int32
	enchant int32
}

func bulkEnchantEquipment(items ...bulkEnchantEquipped) core.Equipment {
	var equipment core.Equipment
	for _, equipped := range items {
		equipment[equipped.slot] = core.NewItem(core.ItemSpec{ID: equipped.id, Enchant: equipped.enchant})
	}
	return equipment
}

func bulkEnchantPool(slot proto.ItemSlot, effectIDs ...int32) *proto.BulkEnchantSlotOptions {
	return &proto.BulkEnchantSlotOptions{Slot: slot, EffectIds: effectIDs}
}

func bulkEnchantSelection(ids ...int32) []*proto.ItemSpec {
	specs := make([]*proto.ItemSpec, 0, len(ids))
	for _, id := range ids {
		specs = append(specs, &proto.ItemSpec{Id: id})
	}
	return specs
}

func newBulkEnchantTestGenerator(t *testing.T, spec proto.Spec, class proto.Class, equipment core.Equipment, settings *proto.BulkSettings) *bulkSimCandidateGenerator {
	t.Helper()
	generator, err := newBulkSimCandidateGeneratorForSpec(settings, class, spec, equipment)
	if err != nil {
		t.Fatalf("newBulkSimCandidateGeneratorForSpec: %v", err)
	}
	return generator
}

func bulkEnchantCandidateGear(t *testing.T, generator *bulkSimCandidateGenerator) []*proto.EquipmentSpec {
	t.Helper()
	candidates, err := generator.buildCandidates()
	if err != nil {
		t.Fatalf("buildCandidates: %v", err)
	}
	gear := make([]*proto.EquipmentSpec, 0, len(candidates))
	for _, candidate := range candidates {
		gear = append(gear, candidate.GetGear())
	}
	return gear
}

func bulkEnchantGearKey(gear *proto.EquipmentSpec) string {
	key := ""
	for slot, item := range gear.GetItems() {
		key += fmt.Sprintf("%d=%d/%d/%d;", slot, item.GetId(), item.GetRandomSuffix(), item.GetEnchant())
	}
	return key
}

func assertDistinctGear(t *testing.T, gear []*proto.EquipmentSpec) {
	t.Helper()
	seen := make(map[string]bool, len(gear))
	for _, candidate := range gear {
		key := bulkEnchantGearKey(candidate)
		if seen[key] {
			t.Fatalf("two candidates share the gear set %s", key)
		}
		seen[key] = true
	}
}

func TestBulkEnchantExpandsEquippedItem(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantEquipment(bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotBack, id: bulkEnchantCloakID}),
		&proto.BulkSettings{Enchants: []*proto.BulkEnchantSlotOptions{
			bulkEnchantPool(proto.ItemSlot_ItemSlotBack, bulkEnchantBackA, bulkEnchantBackB),
		}})

	gear := bulkEnchantCandidateGear(t, generator)
	if len(gear) != 2 {
		t.Fatalf("expected 2 candidates (cloak x 2 pooled enchants), got %d", len(gear))
	}
	got := []int32{
		gear[0].GetItems()[proto.ItemSlot_ItemSlotBack].GetEnchant(),
		gear[1].GetItems()[proto.ItemSlot_ItemSlotBack].GetEnchant(),
	}
	if got[0] != bulkEnchantBackA || got[1] != bulkEnchantBackB {
		t.Fatalf("cloak enchants = %v, want [%d %d]", got, bulkEnchantBackA, bulkEnchantBackB)
	}
}

func TestBulkEnchantPoolsDedupeOverlaps(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantEquipment(bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotBack, id: bulkEnchantCloakID}),
		&proto.BulkSettings{Enchants: []*proto.BulkEnchantSlotOptions{
			bulkEnchantPool(proto.ItemSlot_ItemSlotBack, bulkEnchantBackA, bulkEnchantBackA),
			bulkEnchantPool(proto.ItemSlot_ItemSlotBack, bulkEnchantBackA, bulkEnchantBackB),
		}})

	gear := bulkEnchantCandidateGear(t, generator)
	if len(gear) != 2 {
		t.Fatalf("expected 2 candidates from the overlapping pools, got %d", len(gear))
	}
	assertDistinctGear(t, gear)
}

func TestBulkEnchantSkipsUnknownEffectID(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantEquipment(bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotBack, id: bulkEnchantCloakID}),
		&proto.BulkSettings{Enchants: []*proto.BulkEnchantSlotOptions{
			bulkEnchantPool(proto.ItemSlot_ItemSlotBack, bulkEnchantUnknown, bulkEnchantBackA),
		}})

	gear := bulkEnchantCandidateGear(t, generator)
	if len(gear) != 1 {
		t.Fatalf("expected 1 candidate (the unknown effect id skipped), got %d", len(gear))
	}
	if got := gear[0].GetItems()[proto.ItemSlot_ItemSlotBack].GetEnchant(); got != bulkEnchantBackA {
		t.Fatalf("cloak enchant = %d, want %d", got, bulkEnchantBackA)
	}
}

// A 2H-only enchant fits neither one-hander, so both keep the main hand's carry-over enchant.
func TestBulkEnchantFallsBackWhenNothingFits(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRetributionPaladin, proto.Class_ClassPaladin,
		bulkEnchantEquipment(bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotMainHand, id: bulkEnchantSwordAID, enchant: bulkEnchantWeaponKept}),
		&proto.BulkSettings{
			Items: bulkEnchantSelection(bulkEnchantSwordBID),
			Enchants: []*proto.BulkEnchantSlotOptions{
				bulkEnchantPool(proto.ItemSlot_ItemSlotMainHand, bulkEnchantTwoHandOnly),
			},
		})

	gear := bulkEnchantCandidateGear(t, generator)
	if len(gear) != 2 {
		t.Fatalf("expected 2 candidates (one per sword), got %d", len(gear))
	}
	for _, candidate := range gear {
		mainHand := candidate.GetItems()[proto.ItemSlot_ItemSlotMainHand]
		if mainHand.GetEnchant() != bulkEnchantWeaponKept {
			t.Fatalf("main hand %d enchant = %d, want the carry-over %d", mainHand.GetId(), mainHand.GetEnchant(), bulkEnchantWeaponKept)
		}
	}
}

func TestBulkEnchantSharedEffectIDFitsOneHander(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRetributionPaladin, proto.Class_ClassPaladin,
		bulkEnchantEquipment(bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotMainHand, id: bulkEnchantSwordAID, enchant: bulkEnchantWeaponKept}),
		&proto.BulkSettings{Enchants: []*proto.BulkEnchantSlotOptions{
			bulkEnchantPool(proto.ItemSlot_ItemSlotMainHand, bulkEnchantTwoRows),
		}})

	gear := bulkEnchantCandidateGear(t, generator)
	if len(gear) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(gear))
	}
	if got := gear[0].GetItems()[proto.ItemSlot_ItemSlotMainHand].GetEnchant(); got != bulkEnchantTwoRows {
		t.Fatalf("one-hander enchant = %d, want %d", got, bulkEnchantTwoRows)
	}
}

func TestBulkEnchantDualWieldHandPools(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantEquipment(
			bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotMainHand, id: bulkEnchantDaggerAID},
			bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotOffHand, id: bulkEnchantDaggerBID},
		),
		&proto.BulkSettings{Enchants: []*proto.BulkEnchantSlotOptions{
			bulkEnchantPool(proto.ItemSlot_ItemSlotMainHand, bulkEnchantWeaponA),
			bulkEnchantPool(proto.ItemSlot_ItemSlotOffHand, bulkEnchantWeaponB),
		}})

	gear := bulkEnchantCandidateGear(t, generator)
	if len(gear) != 2 {
		t.Fatalf("expected 2 candidates (both dagger orders), got %d", len(gear))
	}
	for _, candidate := range gear {
		mainHand := candidate.GetItems()[proto.ItemSlot_ItemSlotMainHand]
		offHand := candidate.GetItems()[proto.ItemSlot_ItemSlotOffHand]
		if mainHand.GetEnchant() != bulkEnchantWeaponA || offHand.GetEnchant() != bulkEnchantWeaponB {
			t.Fatalf("hands %d+%d / %d+%d, want main hand enchant %d and off hand enchant %d",
				mainHand.GetId(), mainHand.GetEnchant(), offHand.GetId(), offHand.GetEnchant(), bulkEnchantWeaponA, bulkEnchantWeaponB)
		}
		if mainHand.GetId() == offHand.GetId() {
			t.Fatalf("dagger %d wielded in both hands from one copy", mainHand.GetId())
		}
	}
}

// With a pool for the main hand only, the off hand keeps the carry-over: every candidate
// still wields an off-hand dagger with the off hand's existing enchant.
func TestBulkEnchantDualWieldMainHandPoolOnly(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantEquipment(
			bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotMainHand, id: bulkEnchantDaggerAID},
			bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotOffHand, id: bulkEnchantDaggerBID, enchant: bulkEnchantWeaponKept},
		),
		&proto.BulkSettings{Enchants: []*proto.BulkEnchantSlotOptions{
			bulkEnchantPool(proto.ItemSlot_ItemSlotMainHand, bulkEnchantWeaponA),
		}})

	gear := bulkEnchantCandidateGear(t, generator)
	if len(gear) != 2 {
		t.Fatalf("expected 2 candidates (both dagger orders), got %d", len(gear))
	}
	for _, candidate := range gear {
		mainHand := candidate.GetItems()[proto.ItemSlot_ItemSlotMainHand]
		offHand := candidate.GetItems()[proto.ItemSlot_ItemSlotOffHand]
		if offHand.GetId() == 0 || offHand.GetId() == mainHand.GetId() {
			t.Fatalf("hands %d / %d, want two different daggers", mainHand.GetId(), offHand.GetId())
		}
		if mainHand.GetEnchant() != bulkEnchantWeaponA || offHand.GetEnchant() != bulkEnchantWeaponKept {
			t.Fatalf("hand enchants %d / %d, want %d / %d", mainHand.GetEnchant(), offHand.GetEnchant(), bulkEnchantWeaponA, bulkEnchantWeaponKept)
		}
	}
}

func TestBulkEnchantRingVariantsNeedTwoCopies(t *testing.T) {
	registerBulkEnchantFixtures()
	equipment := bulkEnchantEquipment(
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger1, id: bulkEnchantRingRID},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger2, id: bulkEnchantRingSID},
	)
	pools := []*proto.BulkEnchantSlotOptions{
		bulkEnchantPool(proto.ItemSlot_ItemSlotFinger1, bulkEnchantRingA, bulkEnchantRingB),
	}
	ringPair := func(gear *proto.EquipmentSpec) (*proto.ItemSpec, *proto.ItemSpec) {
		return gear.GetItems()[proto.ItemSlot_ItemSlotFinger1], gear.GetItems()[proto.ItemSlot_ItemSlotFinger2]
	}

	oneCopy := bulkEnchantCandidateGear(t, newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue, equipment,
		&proto.BulkSettings{Enchants: pools}))
	if len(oneCopy) != 4 {
		t.Fatalf("expected 4 candidates (2 R variants x 2 S variants), got %d", len(oneCopy))
	}
	for _, gear := range oneCopy {
		first, second := ringPair(gear)
		if first.GetId() == second.GetId() {
			t.Fatalf("ring %d worn twice from one copy (enchants %d, %d)", first.GetId(), first.GetEnchant(), second.GetEnchant())
		}
	}

	twoCopies := bulkEnchantCandidateGear(t, newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue, equipment,
		&proto.BulkSettings{Items: bulkEnchantSelection(bulkEnchantRingRID), Enchants: pools}))
	foundMixedPair := false
	for _, gear := range twoCopies {
		first, second := ringPair(gear)
		if first.GetId() == bulkEnchantRingRID && second.GetId() == bulkEnchantRingRID &&
			first.GetEnchant() != second.GetEnchant() {
			foundMixedPair = true
		}
	}
	if !foundMixedPair {
		t.Fatalf("expected R with both pooled enchants once two copies are selected")
	}
}

func TestBulkEnchantFrozenSlotUntouched(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantEquipment(
			bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger1, id: bulkEnchantRingRID, enchant: bulkEnchantRingKept},
			bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger2, id: bulkEnchantRingSID},
		),
		&proto.BulkSettings{
			FreezeRingSlot: int32(proto.ItemSlot_ItemSlotFinger1),
			Enchants: []*proto.BulkEnchantSlotOptions{
				bulkEnchantPool(proto.ItemSlot_ItemSlotFinger1, bulkEnchantRingA, bulkEnchantRingB),
			},
		})

	gear := bulkEnchantCandidateGear(t, generator)
	if len(gear) != 2 {
		t.Fatalf("expected 2 candidates (frozen R with S x 2 pooled enchants), got %d", len(gear))
	}
	for _, candidate := range gear {
		frozen := candidate.GetItems()[proto.ItemSlot_ItemSlotFinger1]
		if frozen.GetId() != bulkEnchantRingRID || frozen.GetEnchant() != bulkEnchantRingKept {
			t.Fatalf("frozen ring slot holds %d+%d, want %d+%d", frozen.GetId(), frozen.GetEnchant(), bulkEnchantRingRID, bulkEnchantRingKept)
		}
	}
}

func TestBulkFrozenSecondRingKeepsItsSlotAndEnchant(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantEquipment(
			bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger1, id: bulkEnchantRingRID, enchant: bulkEnchantRingA},
			bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger2, id: bulkEnchantRingSID, enchant: bulkEnchantRingKept},
		),
		&proto.BulkSettings{FreezeRingSlot: int32(proto.ItemSlot_ItemSlotFinger2)})

	gear := bulkEnchantCandidateGear(t, generator)
	if len(gear) == 0 {
		t.Fatalf("expected candidates")
	}
	for _, candidate := range gear {
		frozen := candidate.GetItems()[proto.ItemSlot_ItemSlotFinger2]
		if frozen.GetId() != bulkEnchantRingSID || frozen.GetEnchant() != bulkEnchantRingKept {
			t.Fatalf("frozen Finger 2 holds %d+%d, want %d+%d", frozen.GetId(), frozen.GetEnchant(), bulkEnchantRingSID, bulkEnchantRingKept)
		}
		other := candidate.GetItems()[proto.ItemSlot_ItemSlotFinger1]
		if other.GetId() == bulkEnchantRingRID && other.GetEnchant() != bulkEnchantRingA {
			t.Fatalf("Finger 1 ring R holds enchant %d, want its own %d", other.GetEnchant(), bulkEnchantRingA)
		}
	}
}

func TestBulkEnchantFrozenWeaponUntouched(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantEquipment(
			bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotMainHand, id: bulkEnchantDaggerAID, enchant: bulkEnchantWeaponKept},
			bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotOffHand, id: bulkEnchantDaggerBID},
		),
		&proto.BulkSettings{
			FreezeWeaponSlot: int32(proto.ItemSlot_ItemSlotMainHand),
			Enchants: []*proto.BulkEnchantSlotOptions{
				bulkEnchantPool(proto.ItemSlot_ItemSlotMainHand, bulkEnchantWeaponA),
				bulkEnchantPool(proto.ItemSlot_ItemSlotOffHand, bulkEnchantWeaponB),
			},
		})

	gear := bulkEnchantCandidateGear(t, generator)
	if len(gear) != 1 {
		t.Fatalf("expected 1 candidate (frozen A with B + off-hand enchant), got %d", len(gear))
	}
	mainHand := gear[0].GetItems()[proto.ItemSlot_ItemSlotMainHand]
	offHand := gear[0].GetItems()[proto.ItemSlot_ItemSlotOffHand]
	if mainHand.GetId() != bulkEnchantDaggerAID || mainHand.GetEnchant() != bulkEnchantWeaponKept {
		t.Fatalf("frozen main hand holds %d+%d, want %d+%d", mainHand.GetId(), mainHand.GetEnchant(), bulkEnchantDaggerAID, bulkEnchantWeaponKept)
	}
	if offHand.GetId() != bulkEnchantDaggerBID || offHand.GetEnchant() != bulkEnchantWeaponB {
		t.Fatalf("off hand holds %d+%d, want %d+%d", offHand.GetId(), offHand.GetEnchant(), bulkEnchantDaggerBID, bulkEnchantWeaponB)
	}
}

func bulkEnchantMixedEquipment() core.Equipment {
	return bulkEnchantEquipment(
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotBack, id: bulkEnchantCloakID, enchant: bulkEnchantBackA},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger1, id: bulkEnchantRingRID, enchant: bulkEnchantRingKept},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotFinger2, id: bulkEnchantRingSID},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotMainHand, id: bulkEnchantDaggerAID, enchant: bulkEnchantWeaponKept},
		bulkEnchantEquipped{slot: proto.ItemSlot_ItemSlotOffHand, id: bulkEnchantDaggerBID},
	)
}

func TestBulkEnchantNoPoolUnchanged(t *testing.T) {
	registerBulkEnchantFixtures()
	selection := bulkEnchantSelection(bulkEnchantCloakAltID, bulkEnchantRingTID, bulkEnchantDaggerCID, bulkEnchantDaggerAID)

	withoutField := bulkEnchantCandidateGear(t, newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantMixedEquipment(), &proto.BulkSettings{Items: selection}))
	withEmptyField := bulkEnchantCandidateGear(t, newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantMixedEquipment(), &proto.BulkSettings{
			Items: selection,
			Enchants: []*proto.BulkEnchantSlotOptions{
				bulkEnchantPool(proto.ItemSlot_ItemSlotBack),
				bulkEnchantPool(proto.ItemSlot_ItemSlotFinger1),
				bulkEnchantPool(proto.ItemSlot_ItemSlotMainHand),
			},
		}))

	if len(withoutField) != len(withEmptyField) {
		t.Fatalf("candidate count %d without pools, %d with empty pools", len(withoutField), len(withEmptyField))
	}
	for idx := range withoutField {
		if left, right := bulkEnchantGearKey(withoutField[idx]), bulkEnchantGearKey(withEmptyField[idx]); left != right {
			t.Fatalf("candidate %d differs:\n  %s\n  %s", idx, left, right)
		}
	}
}

func TestBulkEnchantCountMatchesCandidates(t *testing.T) {
	registerBulkEnchantFixtures()
	generator := newBulkEnchantTestGenerator(t, proto.Spec_SpecRogue, proto.Class_ClassRogue,
		bulkEnchantMixedEquipment(), &proto.BulkSettings{
			Items: bulkEnchantSelection(bulkEnchantCloakAltID, bulkEnchantRingTID, bulkEnchantDaggerCID, bulkEnchantDaggerAID),
			Enchants: []*proto.BulkEnchantSlotOptions{
				bulkEnchantPool(proto.ItemSlot_ItemSlotBack, bulkEnchantBackA, bulkEnchantBackB),
				bulkEnchantPool(proto.ItemSlot_ItemSlotFinger1, bulkEnchantRingA, bulkEnchantRingB),
				bulkEnchantPool(proto.ItemSlot_ItemSlotMainHand, bulkEnchantWeaponA, bulkEnchantWeaponB),
				bulkEnchantPool(proto.ItemSlot_ItemSlotOffHand, bulkEnchantWeaponB),
			},
		})

	combinations := generator.rawCombinationsCount()
	gear := bulkEnchantCandidateGear(t, generator)
	if combinations != len(gear) {
		t.Fatalf("combination count %d, candidates %d", combinations, len(gear))
	}
	// Without pools: 2 cloaks x 3 ring pairs x 7 dagger pairs (A twice, B, C).
	if combinations <= 2*3*7 {
		t.Fatalf("expected the pools to multiply the combinations, got %d", combinations)
	}
	assertDistinctGear(t, gear)
}
