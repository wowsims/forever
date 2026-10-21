package core

import (
	"fmt"
	"math"
	"slices"
	"sync"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
)

var WITH_DB = false

var itemsByID = map[int32]Item{}
var randomSuffixesByID = map[int32]RandomSuffix{}
var enchantRowsByEffectID = map[int32][]Enchant{} // Every row with distinct slot rules per effect id, in load order.
var itemEffectRandPropPointsByIlvl = map[int32]ItemEffectRandPropPoints{}
var consumablesByID = map[int32]Consumable{}
var spellEffectsById = map[int32]*proto.SpellEffect{}

// These maps are written at request time by AddToDatabase while other requests are
// mid-flight, and a concurrent map read/write is an unrecoverable fatal error rather than
// a panic something can recover from. They are unexported so every access has to go
// through an accessor below that holds dbMu; do not export them.
var dbMu sync.RWMutex

// The only bulk scan of the item map - everything else looks items up by ID. Runs fn for
// every item that belongs to a set, under the read lock, stopping early if fn returns false.
func forEachSetItem(fn func(item Item) bool) {
	dbMu.RLock()
	defer dbMu.RUnlock()
	for _, item := range itemsByID {
		if item.SetName == "" {
			continue
		}
		if !fn(item) {
			return
		}
	}
}

func AddToDatabase(newDB *proto.SimDatabase) {
	addToDatabase(newDB)
}

func addToDatabase(newDB *proto.SimDatabase) {
	dbMu.Lock()
	defer dbMu.Unlock()

	for _, v := range newDB.Items {
		if _, ok := itemsByID[v.Id]; !ok {
			itemsByID[v.Id] = ItemFromProto(v)
		}
	}

	for _, v := range newDB.RandomSuffixes {
		if _, ok := randomSuffixesByID[v.Id]; !ok {
			randomSuffixesByID[v.Id] = RandomSuffixFromProto(v)
		}
	}

	for _, v := range newDB.Enchants {
		enchant := EnchantFromProto(v)
		if !slices.ContainsFunc(enchantRowsByEffectID[v.EffectId], enchant.sameSlotRules) {
			enchantRowsByEffectID[v.EffectId] = append(enchantRowsByEffectID[v.EffectId], enchant)
		}
	}

	for _, v := range newDB.ItemEffectRandPropPoints {
		if _, ok := itemEffectRandPropPointsByIlvl[v.Ilvl]; !ok {
			itemEffectRandPropPointsByIlvl[v.Ilvl] = ItemEffectRandPropPointsFromProto(v)
		}
	}
	for _, v := range newDB.Consumables {
		if _, ok := consumablesByID[v.Id]; !ok {
			consumablesByID[v.Id] = ConsumableFromProto(v)
		}
	}
	for _, v := range newDB.SpellEffects {
		if _, ok := spellEffectsById[v.Id]; !ok {
			spellEffectsById[v.Id] = v
		}
	}
}

type ItemEffectRandPropPoints struct {
	Ilvl           int32
	RandPropPoints int32
}

// Converts a protobuf ItemEffectRandPropPoints to a Go ItemEffectRandPropPoints
func ItemEffectRandPropPointsFromProto(ieRpp *proto.ItemEffectRandPropPoints) ItemEffectRandPropPoints {
	return ItemEffectRandPropPoints{
		Ilvl:           ieRpp.GetIlvl(),
		RandPropPoints: ieRpp.GetRandPropPoints(),
	}
}

// Converts a Go ItemEffectRandPropPoints to a protobuf ItemEffectRandPropPoints
func ItemEffectRandPropPointsToProto(ieRpp ItemEffectRandPropPoints) *proto.ItemEffectRandPropPoints {
	return &proto.ItemEffectRandPropPoints{
		Ilvl:           ieRpp.Ilvl,
		RandPropPoints: ieRpp.RandPropPoints,
	}
}

type Consumable struct {
	Id                 int32
	Type               proto.ConsumableType
	Stats              stats.Stats
	BuffsMainStat      bool
	Name               string
	ItemEffects        []*proto.ItemEffect
	RequiredProfession proto.Profession
	ReqProfRank        int32
	// The category the first on-use shares: the item row's, or where it states none, its spell's.
	CategoryId int32
	// The enchant a temporary enchant puts on a weapon, nil on consumables that enchant nothing, and
	// the weapon subclasses and inventory types it takes, where a zero mask takes any.
	TemporaryEnchant        *Enchant
	WeaponSubclassMask      int32
	WeaponInventoryTypeMask int32
	// The exclusive group, of which only the first consumable the spec names applies; empty on
	// consumables that stack with everything.
	GroupKey string
}

// The item's on-use rows.
func (consumable Consumable) OnUseEffects() []*proto.ItemEffect {
	return FilterSlice(consumable.ItemEffects, func(effect *proto.ItemEffect) bool { return effect.GetOnUse() != nil })
}

// The first on-use row, the one a consumable's builder runs.
func (consumable Consumable) OnUse() *proto.ItemEffect {
	onUses := consumable.OnUseEffects()
	if len(onUses) == 0 {
		panic(fmt.Sprintf("consumable %d (%s) states no on-use row", consumable.Id, consumable.Name))
	}
	return onUses[0]
}

func ConsumableFromProto(consumable *proto.Consumable) Consumable {
	var temporaryEnchant *Enchant
	if consumable.TemporaryEnchant != nil {
		enchant := EnchantFromProto(consumable.TemporaryEnchant)
		temporaryEnchant = &enchant
	}
	return Consumable{
		Id:                      consumable.Id,
		Type:                    consumable.Type,
		Stats:                   stats.FromProtoArray(consumable.Stats),
		BuffsMainStat:           consumable.BuffsMainStat,
		Name:                    consumable.Name,
		ItemEffects:             consumable.ItemEffects,
		RequiredProfession:      consumable.RequiredProfession,
		ReqProfRank:             consumable.ReqProfRank,
		CategoryId:              consumable.CategoryId,
		TemporaryEnchant:        temporaryEnchant,
		WeaponSubclassMask:      consumable.WeaponSubclassMask,
		WeaponInventoryTypeMask: consumable.WeaponInventoryTypeMask,
		GroupKey:                consumable.GroupKey,
	}
}

type Item struct {
	ID        int32
	Type      proto.ItemType
	ArmorType proto.ArmorType
	// Weapon Stats
	WeaponType       proto.WeaponType
	HandType         proto.HandType
	RangedWeaponType proto.RangedWeaponType
	WeaponDamageMin  float64
	WeaponDamageMax  float64
	UnroundedDPS     float64
	SwingSpeed       float64
	QualityModifier  float64           // Per-item offset to weapon "average damage"; negative for caster weapons.
	SpellSchool      proto.SpellSchool // The weapon's damage school; unset means physical.

	Name          string
	Stats         stats.Stats // Stats applied to wearer
	PseudoStats   []float64   // Indexed by proto.PseudoStat; percents through stats.FromPseudoStatsProto, haste through registerEquipSpeedAuras
	Quality       proto.ItemQuality
	Unique        bool
	LimitCategory int32
	SetName       string // Empty string if not part of a set.
	SetID         int32  // 0 if not part of a set.

	ClassAllowlist []proto.Class

	// Modified for each instance of the item.
	RandomSuffix RandomSuffix
	Enchant      Enchant

	//Internal use
	TempEnchant             int32
	TempEnchantWeaponDamage float64
	WeaponImbue             int32        // The Effect 360 imbue (poison, firestone, class imbue), beside the Effect 54 TempEnchant.
	TempEnchantStats        *stats.Stats // The stats the TempEnchant and the WeaponImbue put on the item, counted with its enchant's.
	TempEnchantPseudoStats  []float64    // The TempEnchant's pseudo stats, whose haste registerEquipSpeedAuras applies.
	WeaponImbuePseudoStats  []float64    // The WeaponImbue's pseudo stats, whose haste registerEquipSpeedAuras applies.
	ScalingOptions          map[int32]*proto.ScalingItemProperties
	RandPropPoints          int32
	ItemEffects             []*proto.ItemEffect
}

// Matches either slot: the Effect 54 TempEnchant or the Effect 360 WeaponImbue.
func (item *Item) HasTempEnchant(effectID int32) bool {
	return item.TempEnchant == effectID || item.WeaponImbue == effectID
}

func ItemFromProto(pData *proto.SimItem) Item {
	return Item{
		ID:               pData.Id,
		Name:             pData.Name,
		Type:             pData.Type,
		ArmorType:        pData.ArmorType,
		WeaponType:       pData.WeaponType,
		HandType:         pData.HandType,
		RangedWeaponType: pData.RangedWeaponType,
		SwingSpeed:       pData.WeaponSpeed,
		QualityModifier:  pData.QualityModifier,
		SpellSchool:      pData.SpellSchool,
		PseudoStats:      pData.PseudoStats,
		Unique:           pData.Unique,
		LimitCategory:    pData.LimitCategory,
		SetName:          pData.SetName,
		SetID:            pData.SetId,
		ClassAllowlist:   pData.ClassAllowlist,
		ScalingOptions:   pData.ScalingOptions,
		ItemEffects:      pData.ItemEffects,
	}
}

func (item *Item) UsableBy(class proto.Class) bool {
	return len(item.ClassAllowlist) == 0 || slices.Contains(item.ClassAllowlist, class)
}

func (item *Item) ToItemSpecProto() *proto.ItemSpec {
	itemSpec := &proto.ItemSpec{
		Id:           item.ID,
		RandomSuffix: item.RandomSuffix.ID,
		Enchant:      item.Enchant.EffectID,
	}

	return itemSpec
}

type RandomSuffix struct {
	ID               int32
	Name             string
	Stats            stats.Stats // Allocations per 10000 of the item's random property points.
	PseudoStats      []float64   // Indexed by proto.PseudoStat; flat, through stats.FromPseudoStatsProto.
	ReducedCasterDps bool        // A weapon with the suffix deals its scaling options' ReducedCasterWeaponDamage.
}

func RandomSuffixFromProto(pData *proto.ItemRandomSuffix) RandomSuffix {
	return RandomSuffix{
		ID:               pData.Id,
		Name:             pData.Name,
		Stats:            stats.FromProtoArray(pData.Stats),
		PseudoStats:      pData.PseudoStats,
		ReducedCasterDps: pData.ReducedCasterDps,
	}
}

type Enchant struct {
	EffectID       int32 // Used by UI to apply effect to tooltip
	Stats          stats.Stats
	PseudoStats    []float64 // Indexed by proto.PseudoStat; percents through stats.FromPseudoStatsProto, haste through registerEquipSpeedAuras
	WeaponDamage   float64   // Applied by newWeaponFromItem
	EnchantEffects []*proto.ItemEffect
	Name           string         // Only needed for unit tests
	Type           proto.ItemType // Only needed for unit tests
	EnchantType    proto.EnchantType
	ExtraTypes     []proto.ItemType
	ProcKind       proto.EnchantProcKind
}

func EnchantFromProto(pData *proto.SimEnchant) Enchant {
	return Enchant{
		EffectID:       pData.EffectId,
		Stats:          stats.FromProtoArray(pData.Stats),
		PseudoStats:    pData.PseudoStats,
		WeaponDamage:   pData.WeaponDamage,
		EnchantEffects: pData.EnchantEffects,
		Name:           pData.Name,
		Type:           pData.Type,
		EnchantType:    pData.EnchantType,
		ExtraTypes:     append([]proto.ItemType(nil), pData.ExtraTypes...),
		ProcKind:       pData.ProcKind,
	}
}

type ItemSpec struct {
	ID           int32
	RandomSuffix int32
	Enchant      int32
}

type Equipment [NumItemSlots]Item

func (equipment *Equipment) MainHand() *Item {
	return &equipment[proto.ItemSlot_ItemSlotMainHand]
}

func (equipment *Equipment) OffHand() *Item {
	return &equipment[proto.ItemSlot_ItemSlotOffHand]
}

func (equipment *Equipment) Ranged() *Item {
	return &equipment[proto.ItemSlot_ItemSlotRanged]
}

func (equipment *Equipment) Head() *Item {
	return &equipment[proto.ItemSlot_ItemSlotHead]
}

func (equipment *Equipment) Neck() *Item {
	return &equipment[proto.ItemSlot_ItemSlotNeck]
}

func (equipment *Equipment) Shoulder() *Item {
	return &equipment[proto.ItemSlot_ItemSlotShoulder]
}

func (equipment *Equipment) Back() *Item {
	return &equipment[proto.ItemSlot_ItemSlotBack]
}

func (equipment *Equipment) Chest() *Item {
	return &equipment[proto.ItemSlot_ItemSlotChest]
}

func (equipment *Equipment) Wrist() *Item {
	return &equipment[proto.ItemSlot_ItemSlotWrist]
}

func (equipment *Equipment) Hands() *Item {
	return &equipment[proto.ItemSlot_ItemSlotHands]
}

func (equipment *Equipment) Waist() *Item {
	return &equipment[proto.ItemSlot_ItemSlotWaist]
}

func (equipment *Equipment) Legs() *Item {
	return &equipment[proto.ItemSlot_ItemSlotLegs]
}

func (equipment *Equipment) Feet() *Item {
	return &equipment[proto.ItemSlot_ItemSlotFeet]
}

func (equipment *Equipment) Trinket1() *Item {
	return &equipment[proto.ItemSlot_ItemSlotTrinket1]
}

func (equipment *Equipment) Trinket2() *Item {
	return &equipment[proto.ItemSlot_ItemSlotTrinket2]
}

func (equipment *Equipment) Finger1() *Item {
	return &equipment[proto.ItemSlot_ItemSlotFinger1]
}

func (equipment *Equipment) Finger2() *Item {
	return &equipment[proto.ItemSlot_ItemSlotFinger2]
}

func (equipment *Equipment) GetItemBySlot(slot proto.ItemSlot) *Item {
	if (slot < 0) || (slot >= NumItemSlots) {
		panic(fmt.Sprintf("%d is an invalid item slot index!", slot))
	}

	return &equipment[slot]
}

func (equipment *Equipment) EquipItem(item Item) {
	if item.Type == proto.ItemType_ItemTypeFinger {
		if equipment.Finger1().ID == 0 {
			*equipment.Finger1() = item
		} else {
			*equipment.Finger2() = item
		}
	} else if item.Type == proto.ItemType_ItemTypeTrinket {
		if equipment.Trinket1().ID == 0 {
			*equipment.Trinket1() = item
		} else {
			*equipment.Trinket2() = item
		}
	} else if item.Type == proto.ItemType_ItemTypeWeapon {
		if item.WeaponType == proto.WeaponType_WeaponTypeShield && equipment.MainHand().HandType != proto.HandType_HandTypeTwoHand {
			*equipment.OffHand() = item
		} else if item.HandType == proto.HandType_HandTypeMainHand || item.HandType == proto.HandType_HandTypeUnknown {
			*equipment.MainHand() = item
		} else if item.HandType == proto.HandType_HandTypeOffHand {
			*equipment.OffHand() = item
		} else if item.HandType == proto.HandType_HandTypeOneHand || item.HandType == proto.HandType_HandTypeTwoHand {
			if equipment.MainHand().ID == 0 {
				*equipment.MainHand() = item
			} else if equipment.OffHand().ID == 0 {
				*equipment.OffHand() = item
			}
		}
	} else {
		equipment[ItemTypeToSlot(item.Type)] = item
	}
}

func (equipment *Equipment) EquipEnchant(enchant Enchant) {
	// Some shield enchants parse as ItemTypeUnknown, so default those to
	// the OH slot to ensure they still get tested.
	if enchant.Type == proto.ItemType_ItemTypeUnknown {
		equipment.OffHand().Enchant = enchant
	} else {
		equipment[ItemTypeToSlot(enchant.Type)].Enchant = enchant
	}
}

func (equipment *Equipment) containsEnchantInSlot(effectID int32, slot proto.ItemSlot) bool {
	return (equipment[slot].Enchant.EffectID == effectID) || equipment[slot].HasTempEnchant(effectID)
}

func (equipment *Equipment) containsEnchantInSlots(effectID int32, possibleSlots []proto.ItemSlot) bool {
	return slices.ContainsFunc(possibleSlots, func(slot proto.ItemSlot) bool {
		return equipment.containsEnchantInSlot(effectID, slot)
	})
}

func (equipment *Equipment) containsItemInSlots(itemID int32, possibleSlots []proto.ItemSlot) bool {
	return slices.ContainsFunc(possibleSlots, func(slot proto.ItemSlot) bool {
		return equipment[slot].ID == itemID
	})
}

func GetEnchantByEffectID(effectID int32) *Enchant {
	dbMu.RLock()
	rows := enchantRowsByEffectID[effectID]
	dbMu.RUnlock()
	if len(rows) == 0 {
		return nil
	}
	enchant := rows[0]
	return &enchant
}

// Every row under effectID, as AddToDatabase stores them.
func GetEnchantRowsByEffectID(effectID int32) []Enchant {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return slices.Clone(enchantRowsByEffectID[effectID])
}

// The first row under effectID that fits item, for the stats an enchant gives that item.
func FittingEnchantRow(effectID int32, item Item) (Enchant, bool) {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return fittingEnchantRow(effectID, item)
}

// The game files carry several rows under some effect ids (1897 is both 2H Lesser Impact and
// Superior Striking), differing in the items they fit. Takes no lock: it runs under the read
// lock in FittingEnchantRow and NewItem.
func fittingEnchantRow(effectID int32, item Item) (Enchant, bool) {
	for _, row := range enchantRowsByEffectID[effectID] {
		if enchantRowAppliesToItem(row, item) {
			return row, true
		}
	}
	return Enchant{}, false
}

// A static enchant gives only stats, pseudo-stats or weapon damage and has no temporary effect.
func (enchant Enchant) IsStatic() bool {
	if enchant.ProcKind != proto.EnchantProcKind_EnchantProcKindNone {
		return false
	}
	return enchant.Stats != (stats.Stats{}) || enchant.WeaponDamage != 0 || slices.ContainsFunc(enchant.PseudoStats, func(v float64) bool { return v != 0 })
}

func EnchantAppliesToItem(effectID int32, item Item) bool {
	_, ok := FittingEnchantRow(effectID, item)
	return ok
}

// Takes no lock: it runs under the read lock through fittingEnchantRow.
func enchantRowAppliesToItem(enchant Enchant, item Item) bool {
	if enchant.EnchantType == proto.EnchantType_EnchantTypeTwoHand && item.HandType != proto.HandType_HandTypeTwoHand {
		return false
	}
	if enchant.EnchantType == proto.EnchantType_EnchantTypeStaff && item.WeaponType != proto.WeaponType_WeaponTypeStaff {
		return false
	}
	if (enchant.EnchantType == proto.EnchantType_EnchantTypeShield) != (item.WeaponType == proto.WeaponType_WeaponTypeShield) {
		return false
	}
	if (enchant.EnchantType == proto.EnchantType_EnchantTypeOffHand) != (item.WeaponType == proto.WeaponType_WeaponTypeOffHand) {
		return false
	}
	if enchant.Type == proto.ItemType_ItemTypeRanged {
		if item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeBow && item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeCrossbow && item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeGun {
			return false
		}
	} else if item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeUnknown && item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeWand {
		return false
	}
	return CheckSliceOverlap(eligibleEnchantSlots(enchant), EligibleSlotsForItem(&item))
}

func eligibleEnchantSlots(enchant Enchant) []proto.ItemSlot {
	types := append([]proto.ItemType{enchant.Type}, enchant.ExtraTypes...)
	slots := make([]proto.ItemSlot, 0, len(types)*2)
	for _, itemType := range types {
		if typeSlots, ok := ItemTypeToSlotsMap[itemType]; ok {
			slots = append(slots, typeSlots...)
			continue
		}
		if itemType == proto.ItemType_ItemTypeWeapon {
			slots = append(slots, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand)
		}
	}
	return slots
}

func (enchant Enchant) sameSlotRules(other Enchant) bool {
	return enchant.Type == other.Type && enchant.EnchantType == other.EnchantType && slices.Equal(enchant.ExtraTypes, other.ExtraTypes)
}

func GetConsumableByID(id int32) Consumable {
	dbMu.RLock()
	c := consumablesByID[id]
	dbMu.RUnlock()
	return c
}

func GetSpellEffectByID(id int32) *proto.SpellEffect {
	dbMu.RLock()
	e := spellEffectsById[id]
	dbMu.RUnlock()
	return e
}

func GetItemEffectRandPropPointsByIlvl(ilvl int32) *ItemEffectRandPropPoints {
	dbMu.RLock()
	p, ok := itemEffectRandPropPointsByIlvl[ilvl]
	dbMu.RUnlock()
	if !ok {
		return nil
	}
	return &p
}

func (equipment *Equipment) ToEquipmentSpecProto() *proto.EquipmentSpec {
	return &proto.EquipmentSpec{
		Items: MapSlice(equipment[:], func(item Item) *proto.ItemSpec {
			return item.ToItemSpecProto()
		}),
	}
}

// Structs used for looking up items/enchants
type EquipmentSpec [NumItemSlots]ItemSpec

func ProtoToEquipmentSpec(es *proto.EquipmentSpec) EquipmentSpec {
	var coreEquip EquipmentSpec
	for i, item := range es.Items {
		coreEquip[i] = ItemSpec{
			ID:           item.Id,
			RandomSuffix: item.RandomSuffix,
			Enchant:      item.Enchant,
		}
	}
	return coreEquip
}

func NewItem(itemSpec ItemSpec) Item {
	// AddToDatabase writes these maps while other requests are mid-flight, so every
	// lookup below has to be under the lock. Held across the whole function rather
	// than reacquired per lookup; nothing in here takes dbMu again.
	dbMu.RLock()
	defer dbMu.RUnlock()

	item := Item{}
	if foundItem, ok := itemsByID[itemSpec.ID]; ok {
		item = foundItem
	} else {
		panic(fmt.Sprintf("No item with id: %d", itemSpec.ID))
	}

	scalingOptions := item.ScalingOptions[0]
	item.Stats = stats.FromProtoMap(scalingOptions.Stats)

	item.WeaponDamageMax = scalingOptions.WeaponDamageMax
	item.WeaponDamageMin = scalingOptions.WeaponDamageMin
	item.UnroundedDPS = scalingOptions.UnroundedDps
	item.RandPropPoints = scalingOptions.RandPropPoints

	if itemSpec.RandomSuffix != 0 {
		if randomSuffix, ok := randomSuffixesByID[itemSpec.RandomSuffix]; ok {
			item.RandomSuffix = randomSuffix
		} else {
			panic(fmt.Sprintf("No random suffix with id: %d", itemSpec.RandomSuffix))
		}
		if reduced := scalingOptions.ReducedCasterWeaponDamage; item.RandomSuffix.ReducedCasterDps && reduced != nil {
			item.WeaponDamageMin = reduced.WeaponDamageMin
			item.WeaponDamageMax = reduced.WeaponDamageMax
			item.UnroundedDPS = reduced.UnroundedDps
		}
	}

	if itemSpec.Enchant != 0 {
		if rows := enchantRowsByEffectID[itemSpec.Enchant]; len(rows) == 1 {
			item.Enchant = rows[0]
		} else if enchant, ok := fittingEnchantRow(itemSpec.Enchant, item); ok {
			item.Enchant = enchant
		} else if len(rows) > 0 {
			item.Enchant = rows[0]
		}
		// else {
		// 	panic(fmt.Sprintf("No enchant with id: %d", itemSpec.Enchant))
		// }
	}

	return item
}

func NewEquipmentSet(equipSpec EquipmentSpec) Equipment {
	equipment := Equipment{}
	for _, itemSpec := range equipSpec {
		if itemSpec.ID != 0 {
			equipment.EquipItem(NewItem(itemSpec))
		}
	}

	return equipment
}

func ProtoToEquipment(es *proto.EquipmentSpec) Equipment {
	return NewEquipmentSet(ProtoToEquipmentSpec(es))
}

// Folds the stats of every matching area into Stats.
func (item Item) inArea(areaTypes []proto.AreaType) Item {
	if item.ID == 0 {
		return item
	}
	for _, areaStats := range item.ScalingOptions[0].GetAreaStats() {
		if slices.Contains(areaTypes, areaStats.AreaType) {
			item.Stats = item.Stats.Add(stats.FromProtoMap(areaStats.Stats))
		}
	}
	return item
}

func (equipment Equipment) inArea(areaTypes []proto.AreaType) Equipment {
	for i := range equipment {
		equipment[i] = equipment[i].inArea(areaTypes)
	}
	return equipment
}

// Like ItemSpec, but uses names for reference instead of ID.
type ItemStringSpec struct {
	Name    string
	Enchant string
}

func EquipmentSpecFromJsonString(jsonString string) *proto.EquipmentSpec {
	es := &proto.EquipmentSpec{}

	data := []byte(jsonString)
	if err := protojson.Unmarshal(data, es); err != nil {
		panic(err)
	}
	return es
}

func ItemSwapFromJsonString(jsonString string) *proto.ItemSwap {
	is := &proto.ItemSwap{}

	data := []byte(jsonString)
	if err := protojson.Unmarshal(data, is); err != nil {
		panic(err)
	}
	return is
}

func (equipment *Equipment) Stats(spec proto.Spec) stats.Stats {
	equipStats := stats.Stats{}

	for _, item := range equipment {
		equipStats = equipStats.Add(ItemEquipmentBaseStats(item))
		equipStats = equipStats.Add(ItemEquipmentEnchantStats(item))
	}

	return equipStats
}

// Returns the base stats on the equipment. That is all stats without Enchants
func ItemEquipmentBaseStats(item Item) stats.Stats {
	equipStats := stats.Stats{}

	if item.ID == 0 {
		return equipStats
	}

	equipStats = equipStats.Add(item.Stats)
	equipStats = equipStats.Add(stats.FromPseudoStatsProto(item.PseudoStats))

	// The game rounds each suffix stat to the nearest point: 4106 Intellect allocation on a ring of 16
	// random property points is 6.57, shown and applied as 7. Multiplying before dividing keeps an exact
	// half exact (47500 at 6 points is 28.5, shown as 29).
	for stat, allocation := range item.RandomSuffix.Stats {
		equipStats[stat] += math.Round(allocation * float64(item.RandPropPoints) / 10000)
	}
	equipStats = equipStats.Add(stats.FromPseudoStatsProto(item.RandomSuffix.PseudoStats))

	return equipStats
}

func ItemEquipmentEnchantStats(item Item) stats.Stats {
	if item.ID == 0 {
		return stats.Stats{}
	}

	equipStats := stats.Stats{}
	equipStats = equipStats.Add(item.Enchant.Stats)
	equipStats = equipStats.Add(stats.FromPseudoStatsProto(item.Enchant.PseudoStats))
	if item.TempEnchantStats != nil {
		equipStats = equipStats.Add(*item.TempEnchantStats)
	}

	return equipStats
}

func GetItemByID(id int32) *Item {
	dbMu.RLock()
	item, ok := itemsByID[id]
	dbMu.RUnlock()
	if !ok {
		return nil
	}
	return &item
}

func ItemTypeToSlot(it proto.ItemType) proto.ItemSlot {
	switch it {
	case proto.ItemType_ItemTypeHead:
		return proto.ItemSlot_ItemSlotHead
	case proto.ItemType_ItemTypeNeck:
		return proto.ItemSlot_ItemSlotNeck
	case proto.ItemType_ItemTypeShoulder:
		return proto.ItemSlot_ItemSlotShoulder
	case proto.ItemType_ItemTypeBack:
		return proto.ItemSlot_ItemSlotBack
	case proto.ItemType_ItemTypeChest:
		return proto.ItemSlot_ItemSlotChest
	case proto.ItemType_ItemTypeWrist:
		return proto.ItemSlot_ItemSlotWrist
	case proto.ItemType_ItemTypeHands:
		return proto.ItemSlot_ItemSlotHands
	case proto.ItemType_ItemTypeWaist:
		return proto.ItemSlot_ItemSlotWaist
	case proto.ItemType_ItemTypeLegs:
		return proto.ItemSlot_ItemSlotLegs
	case proto.ItemType_ItemTypeFeet:
		return proto.ItemSlot_ItemSlotFeet
	case proto.ItemType_ItemTypeFinger:
		return proto.ItemSlot_ItemSlotFinger1
	case proto.ItemType_ItemTypeTrinket:
		return proto.ItemSlot_ItemSlotTrinket1
	case proto.ItemType_ItemTypeWeapon:
		return proto.ItemSlot_ItemSlotMainHand
	case proto.ItemType_ItemTypeRanged:
		return proto.ItemSlot_ItemSlotRanged
	}

	return 255
}

// gen_db -gen=go-to-ts writes this to the UI as ITEM_TYPE_TO_SLOTS, read by getEligibleItemSlots in
// ui/sim/proto/items.ts.
var ItemTypeToSlotsMap = map[proto.ItemType][]proto.ItemSlot{
	proto.ItemType_ItemTypeUnknown:  {},
	proto.ItemType_ItemTypeHead:     {proto.ItemSlot_ItemSlotHead},
	proto.ItemType_ItemTypeNeck:     {proto.ItemSlot_ItemSlotNeck},
	proto.ItemType_ItemTypeShoulder: {proto.ItemSlot_ItemSlotShoulder},
	proto.ItemType_ItemTypeBack:     {proto.ItemSlot_ItemSlotBack},
	proto.ItemType_ItemTypeChest:    {proto.ItemSlot_ItemSlotChest},
	proto.ItemType_ItemTypeWrist:    {proto.ItemSlot_ItemSlotWrist},
	proto.ItemType_ItemTypeHands:    {proto.ItemSlot_ItemSlotHands},
	proto.ItemType_ItemTypeWaist:    {proto.ItemSlot_ItemSlotWaist},
	proto.ItemType_ItemTypeLegs:     {proto.ItemSlot_ItemSlotLegs},
	proto.ItemType_ItemTypeFeet:     {proto.ItemSlot_ItemSlotFeet},
	proto.ItemType_ItemTypeFinger:   {proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2},
	proto.ItemType_ItemTypeTrinket:  {proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2},
	proto.ItemType_ItemTypeRanged:   {proto.ItemSlot_ItemSlotRanged},
	// ItemType_ItemTypeWeapon is excluded intentionally - the slot cannot be decided based on type alone for weapons.
}

// The item slots an item can occupy.
func EligibleSlotsForItem(item *Item) []proto.ItemSlot {
	if item == nil {
		return nil
	}
	if slots, ok := ItemTypeToSlotsMap[item.Type]; ok {
		return slots
	}

	if item.Type == proto.ItemType_ItemTypeWeapon {
		switch item.HandType {
		case proto.HandType_HandTypeTwoHand, proto.HandType_HandTypeMainHand:
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}
		case proto.HandType_HandTypeOffHand:
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}
		case proto.HandType_HandTypeOneHand:
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}
		}
	}

	return nil
}
