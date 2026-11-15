package bulk

import (
	"fmt"
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

type bulkSimCandidateGenerator struct {
	settings           *proto.BulkSettings
	playerClass        proto.Class
	playerSpec         proto.Spec
	playerCanDualWield bool
	playerIsEnchanter  bool
	baseEquipment      core.Equipment
	selectedByBulkSlot map[BulkSimItemSlot][]bulkSimCandidateOption
	groupedPairsBySlot map[BulkSimItemSlot][][2]bulkSimCandidateOption
	frozenItems        map[BulkSimItemSlot]*core.Item
	frozenSlots        map[BulkSimItemSlot]proto.ItemSlot // The slot each frozen ring or trinket is worn in.
	frozenWeaponSlot   proto.ItemSlot
	frozenWeaponItem   *core.Item
	weaponTypeFilters  map[proto.ItemSlot][]proto.WeaponType
	copyCounts         map[itemSpecCacheKey]int
	weaponCombos       [][2]*bulkSimCandidateOption
}

func newBulkSimCandidateGenerator(request *proto.BulkSimRequest, player *proto.Player) (*bulkSimCandidateGenerator, error) {
	if player.GetEquipment() == nil {
		return nil, fmt.Errorf("bulk request is missing player equipment")
	}

	playerSpec, err := core.PlayerProtoToSpecSafe(player)
	if err != nil {
		return nil, err
	}

	isEnchanter := player.GetProfession1() == proto.Profession_Enchanting || player.GetProfession2() == proto.Profession_Enchanting
	return newBulkSimCandidateGeneratorForSpec(request.GetBulkSettings(), player.GetClass(), playerSpec, isEnchanter, core.ProtoToEquipment(player.GetEquipment()))
}

func newBulkSimCandidateGeneratorForSpec(settings *proto.BulkSettings, playerClass proto.Class, playerSpec proto.Spec, isEnchanter bool, equipment core.Equipment) (*bulkSimCandidateGenerator, error) {
	generator := &bulkSimCandidateGenerator{
		settings:           settings,
		playerClass:        playerClass,
		playerSpec:         playerSpec,
		playerCanDualWield: core.SpecCanDualWieldCapabilities[playerSpec],
		playerIsEnchanter:  isEnchanter,
		baseEquipment:      equipment,
		selectedByBulkSlot: make(map[BulkSimItemSlot][]bulkSimCandidateOption),
		groupedPairsBySlot: make(map[BulkSimItemSlot][][2]bulkSimCandidateOption),
		frozenItems:        make(map[BulkSimItemSlot]*core.Item),
		frozenSlots:        make(map[BulkSimItemSlot]proto.ItemSlot),
		weaponTypeFilters: map[proto.ItemSlot][]proto.WeaponType{
			proto.ItemSlot_ItemSlotMainHand: settings.GetFreezeMainhandWeaponSlots(),
			proto.ItemSlot_ItemSlotOffHand:  settings.GetFreezeOffhandWeaponSlots(),
		},
	}

	generator.initFrozenSettings()
	if err := generator.initSelectedItems(); err != nil {
		return nil, err
	}
	generator.initGroupedSlotPairs()
	if err := generator.validateGroupedSlots(); err != nil {
		return nil, err
	}
	generator.weaponCombos = generator.getAllWeaponCombos()

	return generator, nil
}

func (generator *bulkSimCandidateGenerator) validateGroupedSlots() error {
	for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
		if len(generator.selectedByBulkSlot[bulkSlot]) == 0 {
			continue
		}
		if len(generator.groupedPairsBySlot[bulkSlot]) == 0 {
			return fmt.Errorf("no equippable pair of items available for grouped bulk slot %d", bulkSlot)
		}
	}
	return nil
}

func (generator *bulkSimCandidateGenerator) buildCandidates() []*proto.BulkGearCandidate {
	rawCombinations := generator.rawCombinationsCount()
	table := generator.newCombinationTable()
	// Never reserve the whole raw space: it is an unbounded product over the bulk slots,
	// so a few extra selections take it past anything that can be allocated for. Append
	// grows geometrically from here anyway.
	candidates := make([]*proto.BulkGearCandidate, 0, min(rawCombinations, maxBulkCandidatePreallocation))
	for comboIdx := range rawCombinations {
		candidates = append(candidates, &proto.BulkGearCandidate{
			Index: int32(len(candidates)),
			Gear:  table.gear(comboIdx),
		})
	}
	return candidates
}

func (generator *bulkSimCandidateGenerator) initFrozenSettings() {
	if slot := generator.settings.GetFreezeRingSlot(); slot == int32(proto.ItemSlot_ItemSlotFinger1) || slot == int32(proto.ItemSlot_ItemSlotFinger2) {
		item := generator.baseEquipment.GetItemBySlot(proto.ItemSlot(slot))
		if item != nil && item.ID != 0 {
			itemCopy := *item
			generator.frozenItems[BulkSimItemSlotFinger] = &itemCopy
			generator.frozenSlots[BulkSimItemSlotFinger] = proto.ItemSlot(slot)
		}
	}
	if slot := generator.settings.GetFreezeTrinketSlot(); slot == int32(proto.ItemSlot_ItemSlotTrinket1) || slot == int32(proto.ItemSlot_ItemSlotTrinket2) {
		item := generator.baseEquipment.GetItemBySlot(proto.ItemSlot(slot))
		if item != nil && item.ID != 0 {
			itemCopy := *item
			generator.frozenItems[BulkSimItemSlotTrinket] = &itemCopy
			generator.frozenSlots[BulkSimItemSlotTrinket] = proto.ItemSlot(slot)
		}
	}
	if slot := generator.settings.GetFreezeWeaponSlot(); slot == int32(proto.ItemSlot_ItemSlotMainHand) || slot == int32(proto.ItemSlot_ItemSlotOffHand) {
		generator.frozenWeaponSlot = proto.ItemSlot(slot)
		if item := generator.baseEquipment.GetItemBySlot(generator.frozenWeaponSlot); item != nil && item.ID != 0 {
			itemCopy := *item
			generator.frozenWeaponItem = &itemCopy
		}
	}
}

func (generator *bulkSimCandidateGenerator) initSelectedItems() error {
	equippedItemsBySlot := make(map[proto.ItemSlot]*core.Item)
	equippedCounts := make(map[itemSpecFingerprintKey]int)
	for slot := range core.NumItemSlots {
		equippedItem := generator.baseEquipment.GetItemBySlot(slot)
		if equippedItem == nil || equippedItem.ID == 0 {
			continue
		}
		itemCopy := *equippedItem
		equippedItemsBySlot[slot] = &itemCopy
		equippedCounts[buildItemSpecFingerprintKey(equippedItem.ToItemSpecProto())]++
	}

	weaponSelected := false
	for _, selectedItem := range generator.settings.GetItems() {
		if selectedItem == nil || selectedItem.GetId() == 0 {
			continue
		}
		selectedFingerprint := buildItemSpecFingerprintKey(selectedItem)
		if equippedCounts[selectedFingerprint] > 0 {
			// Items filling a pair of slots stack: 1 equipped + 1 added = 2 total, enabling
			// same-item combos like [Sp,Sp] or two Rings of Ancient Knowledge. For single-slot
			// items a user-added duplicate of an equipped item is redundant.
			baseItem := core.GetItemByID(selectedItem.GetId())
			if baseItem == nil || !canStackTwoCopies(*baseItem, generator.playerCanDualWield) {
				equippedCounts[selectedFingerprint]--
				continue
			}
			// paired slot: fall through to add it alongside the equipped copy
		}
		baseItem := core.GetItemByID(selectedItem.GetId())
		if baseItem == nil {
			return fmt.Errorf("selected bulk item %d is missing from the database", selectedItem.GetId())
		}

		option := bulkSimCandidateOption{
			spec: googleProto.Clone(selectedItem).(*proto.ItemSpec),
			item: core.NewItem(core.ItemSpec{
				ID:           selectedItem.GetId(),
				RandomSuffix: selectedItem.GetRandomSuffix(),
				Enchant:      selectedItem.GetEnchant(),
			}),
		}

		// Several physical slots can share one bulk slot (Finger1/Finger2, and both hands for a
		// dual-wielder), so dedupe on the bulk slot rather than skipping the secondary physical
		// slot: an off-hand-only item has no other eligible slot, and skipping it dropped the
		// item from the batch entirely.
		addedBulkSlots := make(map[BulkSimItemSlot]struct{}, 2)
		for _, slot := range core.EligibleSlotsForItem(&option.item) {
			if !canEquipItem(option.item, generator.playerClass, generator.playerSpec, slot) {
				continue
			}
			bulkSlot := getBulkItemSlotFromSlot(slot, generator.playerCanDualWield)
			if _, added := addedBulkSlots[bulkSlot]; added {
				continue
			}
			addedBulkSlots[bulkSlot] = struct{}{}
			generator.selectedByBulkSlot[bulkSlot] = append(generator.selectedByBulkSlot[bulkSlot], option)
			if isHandWeaponBulkSlot(bulkSlot) {
				weaponSelected = true
			}
		}
	}

	// Swapping the worn weapons between the hands is a choice only once a weapon is selected;
	// otherwise each stays in the hand it is worn in.
	for slot := range core.NumItemSlots {
		equippedItem := equippedItemsBySlot[slot]
		if equippedItem == nil {
			continue
		}
		bulkSlot := getBulkItemSlotFromSlot(slot, generator.playerCanDualWield && weaponSelected)
		generator.selectedByBulkSlot[bulkSlot] = append(generator.selectedByBulkSlot[bulkSlot], bulkSimCandidateOption{
			spec: equippedItem.ToItemSpecProto(),
			item: *equippedItem,
		})
	}

	generator.initCopyCounts()
	for bulkSlot, options := range generator.selectedByBulkSlot {
		generator.selectedByBulkSlot[bulkSlot] = dedupeCandidateOptions(options)
	}
	generator.initEnchantVariants()
	return nil
}

// Whether both of an item's slots can hold it at once, which is what makes a same-item combo (two
// identical rings, or one weapon in each hand) a valid input.
func canStackTwoCopies(item core.Item, playerCanDualWield bool) bool {
	if item.Unique || item.LimitCategory != 0 {
		return false
	}
	switch item.Type {
	case proto.ItemType_ItemTypeFinger, proto.ItemType_ItemTypeTrinket:
		return true
	case proto.ItemType_ItemTypeWeapon:
		return playerCanDualWield &&
			item.HandType != proto.HandType_HandTypeTwoHand &&
			item.HandType != proto.HandType_HandTypeMainHand &&
			item.HandType != proto.HandType_HandTypeOffHand
	default:
		return false
	}
}

// Capture how many copies of each item are available (selected + equipped) before dedup
// collapses them. An item may only fill both of its slots when at least two copies exist.
func (generator *bulkSimCandidateGenerator) initCopyCounts() {
	generator.copyCounts = make(map[itemSpecCacheKey]int)
	for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotHandWeapon, BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
		for _, option := range generator.selectedByBulkSlot[bulkSlot] {
			generator.copyCounts[buildItemSpecKey(option.spec)]++
		}
	}
}

// Whether this item may occupy both of its slots at once: the item allows it and the batch
// actually holds two copies.
func (generator *bulkSimCandidateGenerator) hasTwoCopies(option bulkSimCandidateOption, item core.Item) bool {
	return canStackTwoCopies(item, generator.playerCanDualWield) && generator.copyCounts[buildItemSpecKey(option.spec)] >= 2
}

func (generator *bulkSimCandidateGenerator) initGroupedSlotPairs() {
	for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
		options := generator.selectedByBulkSlot[bulkSlot]
		if len(options) == 0 {
			continue
		}
		var pairs [][2]bulkSimCandidateOption
		if frozenItem := generator.frozenItems[bulkSlot]; frozenItem != nil {
			pairs = make([][2]bulkSimCandidateOption, 0, len(options))
			frozenSpec := frozenItem.ToItemSpecProto()
			for _, option := range options {
				if candidateOptionEqualsItem(option, *frozenItem) &&
					!generator.hasTwoCopies(option, *frozenItem) {
					continue
				}
				if frozenItem.Unique && frozenItem.ID == option.item.ID {
					continue
				}
				if frozenItem.LimitCategory != 0 && frozenItem.LimitCategory == option.item.LimitCategory {
					continue
				}
				frozenOption := bulkSimCandidateOption{spec: frozenSpec, item: *frozenItem}
				// A frozen item stays in the slot it is worn in: replaceItem takes the enchant from that slot.
				if generator.frozenSlots[bulkSlot] == BulkSimItemSlotToItemSlotPairs[bulkSlot][1] {
					pairs = append(pairs, [2]bulkSimCandidateOption{option, frozenOption})
				} else {
					pairs = append(pairs, [2]bulkSimCandidateOption{frozenOption, option})
				}
			}
		} else {
			pairs = make([][2]bulkSimCandidateOption, 0, len(options)*(len(options)-1)/2)
			for i := range options {
				// Wearing two copies of the same item is a valid input, but only when two copies
				// were actually selected.
				twoCopies := generator.hasTwoCopies(options[i], options[i].item)
				if twoCopies {
					pairs = append(pairs, [2]bulkSimCandidateOption{options[i], options[i]})
				}
				for j := i + 1; j < len(options); j++ {
					if !twoCopies && candidateOptionsShareItem(options[i], options[j]) {
						continue
					}
					if options[i].item.Unique && options[i].item.ID == options[j].item.ID {
						continue
					}
					lc := options[i].item.LimitCategory
					if lc != 0 && lc == options[j].item.LimitCategory {
						continue
					}
					pairs = append(pairs, [2]bulkSimCandidateOption{options[i], options[j]})
				}
			}
		}
		generator.groupedPairsBySlot[bulkSlot] = pairs
	}
}

// The raw combination space is a plain product over the bulk slots and nothing bounds it.
// The product is clamped only to keep it inside the int32 result fields and out of
// overflow - never to refuse the request.
const maxBulkRawCombinations = math.MaxInt32
const maxBulkCandidatePreallocation = 1 << 16

func saturatingCombinationsMul(rawCombinations int, factor int) int {
	if factor == 0 {
		return 0
	}
	if rawCombinations > maxBulkRawCombinations/factor {
		return maxBulkRawCombinations
	}
	return rawCombinations * factor
}

func (generator *bulkSimCandidateGenerator) rawCombinationsCount() int {
	rawCombinations := len(generator.weaponCombos)
	if rawCombinations == 0 {
		rawCombinations = 1
	}
	for _, bulkSlot := range bulkSimNonWeaponOrder {
		numOptions := len(generator.selectedByBulkSlot[bulkSlot])
		if numOptions == 0 {
			continue
		}
		if bulkSlot == BulkSimItemSlotFinger || bulkSlot == BulkSimItemSlotTrinket {
			rawCombinations = saturatingCombinationsMul(rawCombinations, len(generator.groupedPairsBySlot[bulkSlot]))
		} else {
			rawCombinations = saturatingCombinationsMul(rawCombinations, numOptions)
		}
	}
	return rawCombinations
}

type slotItem struct {
	slot proto.ItemSlot
	spec core.ItemSpec
}

// Every part of a candidate, resolved once: the baseline items, and per mixed-radix digit of the
// combination index the slots each choice fills. A choice resolves the same way in every
// combination, as its carry-over enchant comes from the baseline item in its slot.
type combinationTable struct {
	base   core.EquipmentSpec
	digits [][][]slotItem
}

func (generator *bulkSimCandidateGenerator) newCombinationTable() combinationTable {
	var table combinationTable
	for slot, item := range generator.baseEquipment {
		table.base[slot] = generator.slotSpec(proto.ItemSlot(slot), item)
	}
	if generator.baseEquipment[proto.ItemSlot_ItemSlotMainHand].HandType == proto.HandType_HandTypeTwoHand {
		table.base[proto.ItemSlot_ItemSlotOffHand] = core.ItemSpec{}
	}

	if len(generator.weaponCombos) > 0 {
		choices := make([][]slotItem, 0, len(generator.weaponCombos))
		for _, combo := range generator.weaponCombos {
			mainHand := generator.baseEquipment[proto.ItemSlot_ItemSlotMainHand]
			if combo[0] != nil {
				mainHand = generator.resolveItem(proto.ItemSlot_ItemSlotMainHand, *combo[0])
			}
			offHand := generator.baseEquipment[proto.ItemSlot_ItemSlotOffHand]
			if combo[1] != nil {
				offHand = generator.resolveItem(proto.ItemSlot_ItemSlotOffHand, *combo[1])
			}
			if mainHand.HandType == proto.HandType_HandTypeTwoHand {
				offHand = core.Item{}
			}
			choices = append(choices, []slotItem{
				{slot: proto.ItemSlot_ItemSlotMainHand, spec: itemSpecOf(mainHand)},
				{slot: proto.ItemSlot_ItemSlotOffHand, spec: itemSpecOf(offHand)},
			})
		}
		table.digits = append(table.digits, choices)
	}

	for _, bulkSlot := range bulkSimNonWeaponOrder {
		options := generator.selectedByBulkSlot[bulkSlot]
		if len(options) == 0 {
			continue
		}
		if bulkSlot == BulkSimItemSlotFinger || bulkSlot == BulkSimItemSlotTrinket {
			slots := BulkSimItemSlotToItemSlotPairs[bulkSlot]
			pairs := generator.groupedPairsBySlot[bulkSlot]
			choices := make([][]slotItem, 0, len(pairs))
			for _, pair := range pairs {
				choices = append(choices, []slotItem{
					{slot: slots[0], spec: generator.slotSpec(slots[0], generator.resolveItem(slots[0], pair[0]))},
					{slot: slots[1], spec: generator.slotSpec(slots[1], generator.resolveItem(slots[1], pair[1]))},
				})
			}
			table.digits = append(table.digits, choices)
			continue
		}
		slot := BulkSimItemSlotToSingleItemSlot[bulkSlot]
		choices := make([][]slotItem, 0, len(options))
		for _, option := range options {
			choices = append(choices, []slotItem{{slot: slot, spec: itemSpecOf(generator.resolveItem(slot, option))}})
		}
		table.digits = append(table.digits, choices)
	}
	return table
}

func (generator *bulkSimCandidateGenerator) resolveItem(slot proto.ItemSlot, option bulkSimCandidateOption) core.Item {
	if existing := generator.baseEquipment[slot]; existing.ID != 0 {
		return replaceItem(existing, option)
	}
	return createSelectedItem(option)
}

func itemSpecOf(item core.Item) core.ItemSpec {
	return core.ItemSpec{ID: item.ID, RandomSuffix: item.RandomSuffix.ID, Enchant: item.Enchant.EffectID}
}

// A ring enchant needs Enchanting: without it both ring slots come out unenchanted, the baseline's
// and a selected ring's own enchant included.
func (generator *bulkSimCandidateGenerator) slotSpec(slot proto.ItemSlot, item core.Item) core.ItemSpec {
	spec := itemSpecOf(item)
	if !generator.playerIsEnchanter && (slot == proto.ItemSlot_ItemSlotFinger1 || slot == proto.ItemSlot_ItemSlotFinger2) {
		spec.Enchant = 0
	}
	return spec
}

// Builds the gear for one combination index. Its item specs share one backing array, so a
// candidate costs a fixed handful of allocations.
func (table *combinationTable) gear(comboIdx int) *proto.EquipmentSpec {
	items := table.base
	for _, choices := range table.digits {
		for _, placed := range choices[comboIdx%len(choices)] {
			items[placed.slot] = placed.spec
		}
		comboIdx /= len(choices)
	}

	specs := make([]proto.ItemSpec, len(items))
	gear := &proto.EquipmentSpec{Items: make([]*proto.ItemSpec, len(items))}
	for slot, item := range items {
		specs[slot].Id = item.ID
		specs[slot].RandomSuffix = item.RandomSuffix
		specs[slot].Enchant = item.Enchant
		gear.Items[slot] = &specs[slot]
	}
	return gear
}
