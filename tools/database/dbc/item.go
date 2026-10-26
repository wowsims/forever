package dbc

import (
	"math"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

const UPGRADE_SYSTEM_ACTIVE = false

type Item struct {
	Id                    int
	Name                  string
	InventoryType         int
	ItemDelay             int
	OverallQuality        ItemQuality
	DmgVariance           float64
	ItemLevel             int
	ItemClass             int
	ItemSubClass          int
	StatAlloc             []float64
	BonusStat             []int
	ArmorValue            int
	FireResistance        int
	NatureResistance      int
	FrostResistance       int
	ShadowResistance      int
	ArcaneResistance      int
	Flags0                ItemStaticFlags0
	Flags1                ItemStaticFlags1
	Flags2                ItemStaticFlags2
	Flags3                ItemStaticFlags3
	Flags4                ItemStaticFlags4
	FDID                  int
	ItemSetName           string
	ItemSetId             int
	ClassMask             int
	RaceMask              int
	QualityModifier       float64
	DamageType            int // ItemSparse.DamageType, the weapon's school in the game files' order.
	RandomSuffixOptions   []int32
	BonusAmountCalculated []float64
	NameDescription       string // Contains information for i.E. Thunderforging. Normal = Thunderforged, HC = Heroic Thunderforged
	UpgradeID             int
	UpgradePath           []int
	LimitCategory         int
	Bonding               int
}

func (item *Item) ToUIItem() *proto.UIItem {
	return item.ToScaledUIItem(item.ItemLevel)
}

func (item *Item) ToScaledUIItem(itemLevel int) *proto.UIItem {
	scalingProperties := make(map[int32]*proto.ScalingItemProperties)
	var weaponType, handType, rangedType = item.GetWeaponTypes()
	uiItem := &proto.UIItem{
		Type:                MapInventoryTypeToItemType[item.InventoryType],
		Quality:             item.OverallQuality.ToProto(),
		SetName:             item.ItemSetName,
		SetId:               int32(item.ItemSetId),
		Name:                item.Name,
		ClassAllowlist:      GetClassesFromClassMask(item.ClassMask),
		Id:                  int32(item.Id),
		RandomSuffixOptions: item.RandomSuffixOptions,
		RangedWeaponType:    rangedType,
		HandType:            handType,
		WeaponType:          weaponType,
		WeaponSpeed:         float64(item.ItemDelay) / 1000,
		NameDescription:     item.NameDescription,
		LimitCategory:       int32(item.LimitCategory),
		QualityModifier:     item.GetWeaponQualityModifier(),
	}

	item.ParseItemFlags(uiItem)

	if item.ItemClass == ITEM_CLASS_WEAPON {
		uiItem.SpellSchool = item.GetSpellSchool()
	}

	if item.ItemClass == ITEM_CLASS_ARMOR {
		uiItem.ArmorType = MapArmorSubclassToArmorType[item.ItemSubClass]
	}

	if item.ItemSubClass == ITEM_SUBCLASS_ARMOR_SHIELD && uiItem.HandType == proto.HandType_HandTypeUnknown {
		uiItem.HandType = proto.HandType_HandTypeOffHand
		uiItem.WeaponType = proto.WeaponType_WeaponTypeShield
	}

	// Append base itemlevel stats
	scalingProperties[int32(0)] = &proto.ScalingItemProperties{
		WeaponDamageMin: item.WeaponDmgMin(item.ItemLevel),
		WeaponDamageMax: item.WeaponDmgMax(item.ItemLevel),
		Stats:           item.GetStats(item.ItemLevel).ToProtoMap(),
		RandPropPoints:  item.GetRandPropPoints(item.ItemLevel),
		Ilvl:            int32(item.ItemLevel),
	}
	if item.ItemClass == ITEM_CLASS_WEAPON {
		// Equipped damage is rounded; a form that swings at its own speed needs the table cell itself.
		dps, _ := dbcInstance.WeaponDPS(item, item.ItemLevel)
		scalingProperties[0].UnroundedDps = float32(dps)
		if item.hasReducedCasterDpsSuffix() {
			dps, _ := dbcInstance.ReducedCasterWeaponDPS(item, item.ItemLevel)
			low, high := item.WeaponDamage(dps)
			scalingProperties[0].ReducedCasterWeaponDamage = &proto.ReducedCasterWeaponDamage{WeaponDamageMin: low, WeaponDamageMax: high, UnroundedDps: float32(dps)}
		}
	}

	uiItem.ScalingOptions = scalingProperties
	return uiItem
}

func (item *Item) ParseItemFlags(uiItem *proto.UIItem) {
	if item.Flags1.Has(HORDE_SPECIFIC) {
		uiItem.FactionRestriction = proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY
	}

	if item.Flags1.Has(ALLIANCE_SPECIFIC) {
		uiItem.FactionRestriction = proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY
	}

	if item.Flags0.Has(UNIQUE_EQUIPPABLE) {
		uiItem.Unique = true
	}
}

func (item *Item) GetStats(itemLevel int) *stats.Stats {
	stats := &stats.Stats{}
	for i, alloc := range item.BonusStat {
		mapped, success := MapBonusStatIndexToStats(alloc)
		if !success {
			// Skip this stat then
			continue
		}
		value := item.GetScaledStat(i, itemLevel)
		for _, stat := range mapped {
			stats[stat] += value
			if stat == proto.Stat_StatArmorPenetration {
				stats[stat] = math.Abs(stats[stat])
			}
		}
	}

	armor := item.GetArmorValue(itemLevel)
	if armor > 0 {
		stats[proto.Stat_StatArmor] = float64(armor)
	}

	if item.FireResistance > 0 {
		stats[proto.Stat_StatFireResistance] = float64(item.FireResistance)
	}
	if item.NatureResistance > 0 {
		stats[proto.Stat_StatNatureResistance] = float64(item.NatureResistance)
	}
	if item.FrostResistance > 0 {
		stats[proto.Stat_StatFrostResistance] = float64(item.FrostResistance)
	}
	if item.ShadowResistance > 0 {
		stats[proto.Stat_StatShadowResistance] = float64(item.ShadowResistance)
	}
	if item.ArcaneResistance > 0 {
		stats[proto.Stat_StatArcaneResistance] = float64(item.ArcaneResistance)
	}

	if item.ItemClass == ITEM_CLASS_ARMOR && item.QualityModifier > 0 {
		stats[proto.Stat_StatBonusArmor] = item.QualityModifier
	}

	blockValue := item.GetBlockValue(itemLevel)
	if blockValue > 0 {
		stats[proto.Stat_StatBlockValue] += blockValue
	}

	item.addCasterSpellPower(stats, itemLevel)

	return stats
}

// Caster weapons carry spell power that no stat slot states, worked out from the first column of the
// item level budget with the allocations of the "of Magic" and "of Healing" bonus lists: 20000 spell
// power, or on a healing weapon 37600 healing (1.88 times the spell power) and a third of that as spell
// damage, each rounded down. It adds to the weapon's own stats; the rule reproduces the in-game tooltip
// of every weapon with one of the two flags.
const (
	casterSpellPowerAllocation   = 20000
	casterHealingPowerAllocation = 37600
)

func (item *Item) addCasterSpellPower(stats *stats.Stats, itemLevel int) {
	if !item.Flags4.Has(CASTER_SPELL_POWER) && !item.Flags4.Has(CASTER_HEALING_POWER) {
		return
	}
	budget := int(GetDBC().RandomPropertiesByIlvl[itemLevel][item.OverallQuality.ToProto()][0])
	add := func(mod int, value int) {
		mapped, _ := MapBonusStatIndexToStats(mod)
		for _, stat := range mapped {
			stats[stat] += float64(value)
		}
	}
	if item.Flags4.Has(CASTER_HEALING_POWER) {
		healing := casterHealingPowerAllocation * budget / 10000
		add(ITEM_MOD_SPELL_HEALING_DONE, healing)
		add(ITEM_MOD_SPELL_DAMAGE_DONE, healing/3)
		return
	}
	add(ITEM_MOD_SPELL_POWER, casterSpellPowerAllocation*budget/10000)
}

// GetWeaponQualityModifier returns the per-item damage offset for weapons.
// For armor the value is repurposed as bonus armor (see GetStats), so don't expose it as a weapon quality modifier.
func (item *Item) GetWeaponQualityModifier() float64 {
	if item.ItemClass != ITEM_CLASS_WEAPON {
		return 0
	}
	return item.QualityModifier
}

func (item *Item) GetRandPropPoints(itemLevel int) int32 {
	suffixType := item.GetRandomSuffixType()
	randomProperty := GetDBC().RandomPropertiesByIlvl[itemLevel]
	if suffixType < 0 {
		return 0
	}
	return randomProperty[item.OverallQuality.ToProto()][suffixType]
}
func (item *Item) GetScaledStat(index int, itemLevel int) float64 {
	//Todo check if overflow array

	// The game files ship no precomputed StatModifier_bonusAmount, so the allocation
	// budget is the only source: 99.8% of stat slots carry a non-zero StatAlloc.
	slotType := item.GetRandomSuffixType()
	itemBudget := 0.0

	if slotType != -1 && item.OverallQuality > 0 {

		randomProperty := GetDBC().RandomPropertiesByIlvl[itemLevel]
		itemBudget = float64(randomProperty[item.OverallQuality.ToProto()][slotType])

		if item.StatAlloc[index] > 0 && itemBudget > 0 {
			rawValue := math.Round(item.StatAlloc[index] * itemBudget * 0.0001)

			return rawValue
		}
	}
	return 0
}

// Whether one of the item's random suffixes reduces caster DPS, so the item needs the damage it deals with it.
func (item *Item) hasReducedCasterDpsSuffix() bool {
	for _, id := range item.RandomSuffixOptions {
		if dbcInstance.RandomSuffix[int(id)].ReducedCasterDps {
			return true
		}
	}
	return false
}

func (item *Item) WeaponDmgMin(itemLevel int) float64 {
	low, _ := item.WeaponDamage(item.WeaponDps(itemLevel))
	return low
}
func (item *Item) GetBlockValue(itemLevel int) float64 {
	weaponType, _, _ := item.GetWeaponTypes()
	if weaponType == proto.WeaponType_WeaponTypeShield {
		blockValues := GetDBC().ShieldBlockValues[itemLevel]
		blockBudget := float64(blockValues.Values[item.OverallQuality.ToProto()])
		return blockBudget
	}
	return 0
}

func (item *Item) WeaponDmgMax(itemLevel int) float64 {
	_, high := item.WeaponDamage(item.WeaponDps(itemLevel))
	return high
}

func (item *Item) ApproximateScaleCoeff(currIlvl int, newIlvl int) float64 {
	if currIlvl == 0 || newIlvl == 0 {
		return 1.0
	}

	diff := float64(currIlvl - newIlvl)
	return 1.0 / math.Pow(1.15, diff/15.0)
}

// GetArmorValue computes base armor from the item level tables, since the game files carry no armor
// value on the item: ItemArmorShield by quality for a shield, otherwise ItemArmorTotal for the armor
// type, scaled by ItemArmorQuality and the slot's ArmorLocation modifier. A robe takes the chest row.
func (item *Item) GetArmorValue(itemLevel int) int {
	if item.ItemClass != ITEM_CLASS_ARMOR || item.OverallQuality > 6 {
		return 0
	}
	ilvl := item.ItemLevel
	if itemLevel > 0 {
		ilvl = itemLevel
	}

	if item.ItemSubClass == ITEM_SUBCLASS_ARMOR_SHIELD {
		shield, ok := GetDBC().ItemArmorShield[ilvl]
		if !ok {
			return 0
		}
		return int(math.Floor(shield.Quality[item.OverallQuality] + 0.5))
	}

	var totalArmor float64
	total := GetDBC().ItemArmorTotal[ilvl]
	switch item.ItemSubClass {
	case ITEM_SUBCLASS_ARMOR_CLOTH:
		totalArmor = total.Cloth
	case ITEM_SUBCLASS_ARMOR_LEATHER:
		totalArmor = total.Leather
	case ITEM_SUBCLASS_ARMOR_MAIL:
		totalArmor = total.Mail
	case ITEM_SUBCLASS_ARMOR_PLATE:
		totalArmor = total.Plate
	default:
		return 0
	}

	switch item.InventoryType {
	case INVTYPE_HEAD, INVTYPE_SHOULDERS, INVTYPE_CHEST, INVTYPE_WAIST, INVTYPE_LEGS, INVTYPE_FEET, INVTYPE_WRISTS, INVTYPE_HANDS, INVTYPE_CLOAK, INVTYPE_ROBE:
	default:
		return 0
	}
	slot := item.InventoryType
	if slot == INVTYPE_ROBE {
		slot = INVTYPE_CHEST
	}
	quality, ok := GetDBC().ItemArmorQuality[ilvl]
	if !ok {
		return 0
	}
	location := GetDBC().ArmorLocation[slot]
	return int(math.Floor(totalArmor*quality.Quality[item.OverallQuality]*location.Modifier[item.ItemSubClass-1] + 0.5))
}

func (item *Item) GetWeaponTypes() (proto.WeaponType, proto.HandType, proto.RangedWeaponType) {
	weaponType := proto.WeaponType_WeaponTypeUnknown
	rangedWeaponType := proto.RangedWeaponType_RangedWeaponTypeUnknown
	handType := proto.HandType_HandTypeUnknown

	switch item.ItemClass {
	case ITEM_CLASS_ARMOR:
		switch item.ItemSubClass {
		case ITEM_SUBCLASS_ARMOR_MISC:
			if item.InventoryType == INVTYPE_HOLDABLE {
				handType = proto.HandType_HandTypeOffHand
				weaponType = proto.WeaponType_WeaponTypeOffHand
			}
		case ITEM_SUBCLASS_ARMOR_TOTEM:
			rangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeTotem
		case ITEM_SUBCLASS_ARMOR_IDOL:
			rangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeIdol
		case ITEM_SUBCLASS_ARMOR_LIBRAM:
			rangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeLibram
		case ITEM_SUBCLASS_ARMOR_SIGIL:
			rangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeSigil
		}

	case ITEM_CLASS_WEAPON:
		switch item.ItemSubClass {
		case ITEM_SUBCLASS_WEAPON_BOW:
			rangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeBow
		case ITEM_SUBCLASS_WEAPON_CROSSBOW:
			rangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeCrossbow
		case ITEM_SUBCLASS_WEAPON_GUN:
			rangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeGun
		case ITEM_SUBCLASS_WEAPON_WAND:
			rangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeWand
		case ITEM_SUBCLASS_WEAPON_THROWN:
			rangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeThrown
		case ITEM_SUBCLASS_WEAPON_AXE2, ITEM_SUBCLASS_WEAPON_FISHING_POLE, ITEM_SUBCLASS_WEAPON_SPEAR, ITEM_SUBCLASS_WEAPON_STAFF, ITEM_SUBCLASS_WEAPON_POLEARM, ITEM_SUBCLASS_WEAPON_MACE2,
			ITEM_SUBCLASS_WEAPON_EXOTIC2, ITEM_SUBCLASS_WEAPON_SWORD2:
			handType = proto.HandType_HandTypeTwoHand
			weaponType = MapWeaponSubClassToWeaponType[item.ItemSubClass]
		case ITEM_SUBCLASS_WEAPON_EXOTIC, ITEM_SUBCLASS_WEAPON_MACE, ITEM_SUBCLASS_WEAPON_WARGLAIVE, ITEM_SUBCLASS_WEAPON_FIST, ITEM_SUBCLASS_WEAPON_SWORD,
			ITEM_SUBCLASS_WEAPON_DAGGER, ITEM_SUBCLASS_WEAPON_AXE:
			if item.InventoryType == INVTYPE_WEAPON {
				handType = proto.HandType_HandTypeOneHand
			} else if item.InventoryType == INVTYPE_WEAPONOFFHAND {
				handType = proto.HandType_HandTypeOffHand
			} else {
				handType = proto.HandType_HandTypeMainHand
			}
			weaponType = MapWeaponSubClassToWeaponType[item.ItemSubClass]
		}
	}
	if item.ItemSubClass == ITEM_SUBCLASS_ARMOR_SHIELD && handType == proto.HandType_HandTypeUnknown {
		handType = proto.HandType_HandTypeOffHand
		weaponType = proto.WeaponType_WeaponTypeShield
	}
	return weaponType, handType, rangedWeaponType
}

func (item *Item) GetRandomSuffixType() int {
	switch item.ItemClass {
	case ITEM_CLASS_WEAPON:
		switch item.ItemSubClass {
		case ITEM_SUBCLASS_WEAPON_AXE2,
			ITEM_SUBCLASS_WEAPON_MACE2,
			ITEM_SUBCLASS_WEAPON_POLEARM,
			ITEM_SUBCLASS_WEAPON_SWORD2,
			ITEM_SUBCLASS_WEAPON_STAFF:
			return 0

		case ITEM_SUBCLASS_WEAPON_THROWN,
			ITEM_SUBCLASS_WEAPON_GUN,
			ITEM_SUBCLASS_WEAPON_BOW,
			ITEM_SUBCLASS_WEAPON_CROSSBOW,
			ITEM_SUBCLASS_WEAPON_WAND:
			return 4

		default:
			return 3
		}

	case ITEM_CLASS_ARMOR:
		switch item.InventoryType {
		case INVTYPE_HEAD,
			INVTYPE_CHEST,
			INVTYPE_LEGS,
			INVTYPE_ROBE:
			return 0

		case INVTYPE_SHOULDERS,
			INVTYPE_WAIST,
			INVTYPE_FEET,
			INVTYPE_HANDS,
			INVTYPE_TRINKET:
			return 1

		case INVTYPE_NECK,
			INVTYPE_WEAPONOFFHAND,
			INVTYPE_HOLDABLE,
			INVTYPE_FINGER,
			INVTYPE_CLOAK,
			INVTYPE_WRISTS,
			INVTYPE_SHIELD,
			INVTYPE_RELIC:
			return 2

		default:
			return -1
		}

	default:
		return -1
	}
}

// ItemSparse.DamageType counts the schools in the game files' order, which the proto enum does not share.
var damageTypeSchools = [...]proto.SpellSchool{
	proto.SpellSchool_SpellSchoolPhysical,
	proto.SpellSchool_SpellSchoolHoly,
	proto.SpellSchool_SpellSchoolFire,
	proto.SpellSchool_SpellSchoolNature,
	proto.SpellSchool_SpellSchoolFrost,
	proto.SpellSchool_SpellSchoolShadow,
	proto.SpellSchool_SpellSchoolArcane,
}

func (item *Item) GetSpellSchool() proto.SpellSchool {
	if item.DamageType < 0 || item.DamageType >= len(damageTypeSchools) {
		return proto.SpellSchool_SpellSchoolPhysical
	}
	return damageTypeSchools[item.DamageType]
}
