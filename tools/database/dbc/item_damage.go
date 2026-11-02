package dbc

import "math"

type ItemDamageTable struct {
	ItemLevel int
	Quality   []float64
}

// A weapon flagged REDUCED_CASTER_DPS, or carrying a random suffix whose bonus list has an
// ITEM_BONUS_REDUCED_CASTER_DPS row, deals a fixed share of its damage table in game, while the
// game files' caster tables match the regular ones. The shares reproduce the in-game damage range of
// all 165 flagged weapons, QualityModifier included; 29/39 of the two-hand table is 29/30 of the one-hand one.
const (
	casterOneHandDpsShare = 2.0 / 3.0
	casterTwoHandDpsShare = 29.0 / 39.0
)

// The share of its damage table the weapon deals when its caster DPS is reduced; 1 off a weapon slot.
func (item *Item) reducedCasterDpsShare() float64 {
	switch item.InventoryType {
	case INVTYPE_WEAPON, INVTYPE_WEAPONMAINHAND, INVTYPE_WEAPONOFFHAND:
		return casterOneHandDpsShare
	case INVTYPE_2HWEAPON:
		return casterTwoHandDpsShare
	}
	return 1
}

func (item *Item) WeaponDps(itemLevel int) float64 {
	dps, _ := dbcInstance.WeaponDPS(item, itemLevel)
	return dps
}

// The equipped damage range from a damage table cell, rounded as the game shows it.
func (item *Item) WeaponDamage(dps float64) (float64, float64) {
	speed := float64(item.ItemDelay) / 1000
	low := dps*speed*(1-item.DmgVariance/2) + item.QualityModifier*speed
	high := dps*speed*(1+item.DmgVariance/2) + item.QualityModifier*speed
	if low < 0 {
		low = 1
	}
	if high < 0 {
		high = 1
	}
	return math.Floor(low), math.Floor(high + 0.5)
}

// The DPS the weapon deals as the game files flag it; false when the game files have no table cell for it.
func (d *DBC) WeaponDPS(item *Item, itemLevel int) (float64, bool) {
	dps, present := d.weaponTableDPS(item, itemLevel)
	if item.Flags4.Has(REDUCED_CASTER_DPS) {
		dps *= item.reducedCasterDpsShare()
	}
	return dps, present
}

// The DPS the weapon deals with a random suffix that reduces caster DPS: the share of its table cell,
// whatever the item's own flags.
func (d *DBC) ReducedCasterWeaponDPS(item *Item, itemLevel int) (float64, bool) {
	dps, present := d.weaponTableDPS(item, itemLevel)
	return dps * item.reducedCasterDpsShare(), present
}

// The damage table cell for the item's quality and level; false when the game files have no such cell.
func (d *DBC) weaponTableDPS(item *Item, itemLevel int) (float64, bool) {
	quality := item.OverallQuality
	if quality > 6 {
		quality = 4 // Heirlooms = epic
	}
	if itemLevel <= 0 {
		itemLevel = item.ItemLevel
	}
	table := item.weaponDamageTable()
	row, present := d.ItemDamageTable[table][itemLevel]
	if !present || quality < 0 || int(quality) >= len(row.Quality) {
		return 0, false
	}
	return row.Quality[quality], true
}

func (item *Item) weaponDamageTable() string {
	switch item.InventoryType {
	case INVTYPE_WEAPON, INVTYPE_WEAPONMAINHAND, INVTYPE_WEAPONOFFHAND:
		if item.Flags1.Has(CASTER_WEAPON) {
			return "ItemDamageOneHandCaster"
		}
		return "ItemDamageOneHand"
	case INVTYPE_2HWEAPON:
		if item.Flags1.Has(CASTER_WEAPON) {
			return "ItemDamageTwoHandCaster"
		}
		return "ItemDamageTwoHand"
	case INVTYPE_RANGED, INVTYPE_THROWN, INVTYPE_RANGEDRIGHT:
		switch item.ItemSubClass {
		case ITEM_SUBCLASS_WEAPON_BOW, ITEM_SUBCLASS_WEAPON_GUN, ITEM_SUBCLASS_WEAPON_CROSSBOW:
			return "ItemDamageRanged"
		case ITEM_SUBCLASS_WEAPON_THROWN:
			return "ItemDamageThrown"
		case ITEM_SUBCLASS_WEAPON_WAND:
			return "ItemDamageWand"
		}
	}
	return ""
}
