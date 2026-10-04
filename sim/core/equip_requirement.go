package core

import (
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The game files' SpellEquippedItems for a weapon: the weapon subclass and inventory-type masks, and the
// hand an ATTR_EX_3 hand bit names. A zero field is no restriction.
type EquipRequirement struct {
	SubclassMask int32
	InvTypeMask  int32
	Hand         ProcMask
}

// The requirement the config's Equip fields name.
func (config SpellModConfig) Equip() EquipRequirement {
	return EquipRequirement{SubclassMask: config.EquipSubclass, InvTypeMask: config.EquipInvType, Hand: config.EquipHand}
}

func (req EquipRequirement) IsZero() bool {
	return req.SubclassMask == 0 && req.InvTypeMask == 0 && req.Hand == 0
}

// Whether the weapon behind an attack with this proc mask meets the requirement. The weapon is the
// main hand's for a main-hand bit, else the off hand's, else the ranged slot's. An attack with no
// weapon bit, nil equipment (a unit that is not a player), and a shield or held item in the off hand
// never match.
func (req EquipRequirement) matchesEquipment(equipment *Equipment, spellProcMask ProcMask) bool {
	if equipment == nil {
		return false
	}
	if req.Hand != 0 {
		spellProcMask &= req.Hand
	}
	var slot proto.ItemSlot
	switch {
	case spellProcMask.Matches(ProcMaskMeleeMH):
		slot = proto.ItemSlot_ItemSlotMainHand
	case spellProcMask.Matches(ProcMaskMeleeOH):
		slot = proto.ItemSlot_ItemSlotOffHand
	case spellProcMask.Matches(ProcMaskRanged):
		slot = proto.ItemSlot_ItemSlotRanged
	default:
		return false
	}

	return req.matchesItem(slot, &equipment[slot])
}

// The equipment of the player the unit is, or nil for any other unit.
func (unit *Unit) playerEquipment() *Equipment {
	if unit == nil || unit.Type != PlayerUnit || unit.Env == nil {
		return nil
	}
	agent := unit.Env.Raid.GetPlayerFromUnit(unit)
	if agent == nil {
		return nil
	}
	return &agent.GetCharacter().Equipment
}

func (req EquipRequirement) matchesItem(slot proto.ItemSlot, item *Item) bool {
	subclass := item.WeaponSubclassBit()
	if subclass == 0 {
		return false
	}
	if req.SubclassMask != 0 && req.SubclassMask&subclass == 0 {
		return false
	}
	if req.InvTypeMask != 0 && req.InvTypeMask&invTypeBit(slot, item) == 0 {
		return false
	}
	return true
}

// Whether a weapon meeting the requirement sits in the main-hand, off-hand or ranged slot. A hand bit
// narrows the slots to that hand's. A unit that is not a player never meets it.
func (req EquipRequirement) Equipped(character *Character) bool {
	if character == nil || character.Type != PlayerUnit {
		return false
	}

	slots := AllWeaponSlots()
	if req.Hand != 0 {
		slots = nil
		if req.Hand.Matches(ProcMaskMeleeMH) {
			slots = append(slots, proto.ItemSlot_ItemSlotMainHand)
		}
		if req.Hand.Matches(ProcMaskMeleeOH) {
			slots = append(slots, proto.ItemSlot_ItemSlotOffHand)
		}
	}
	for _, slot := range slots {
		if req.matchesItem(slot, &character.Equipment[slot]) {
			return true
		}
	}
	return false
}

// A level that holds only while a weapon meeting the requirement is equipped: set hears the level while
// one is and 0 while none is, again on every weapon change.
type EquipGate struct {
	character *Character
	req       EquipRequirement
	level     float64
	set       func(sim *Simulation, level float64)
}

func (character *Character) NewEquipGate(req EquipRequirement, set func(sim *Simulation, level float64)) *EquipGate {
	gate := &EquipGate{character: character, req: req, set: set}
	character.RegisterWeaponChangeCallback(gate.refresh)
	return gate
}

func (gate *EquipGate) SetLevel(sim *Simulation, level float64) {
	gate.level = level
	gate.refresh(sim)
}

func (gate *EquipGate) Equipped() bool {
	return gate.req.Equipped(gate.character)
}

func (gate *EquipGate) refresh(sim *Simulation) {
	if gate.Equipped() {
		gate.set(sim, gate.level)
	} else {
		gate.set(sim, 0)
	}
}

// Stats the character holds while a weapon meeting the requirement is equipped, each counted at its
// own level. They follow the weapon slots through item swaps.
type EquipGatedStats struct {
	character *Character
	gate      *EquipGate
	stats     []gatedStat
}

type gatedStat struct {
	stat         stats.Stat
	value        float64
	level        float64
	appliedLevel float64
}

func (character *Character) NewEquipGatedStats(req EquipRequirement, bonus stats.Stats) *EquipGatedStats {
	gated := &EquipGatedStats{character: character}
	for stat, value := range bonus {
		if value != 0 {
			gated.stats = append(gated.stats, gatedStat{stat: stats.Stat(stat), value: value})
		}
	}
	gated.gate = character.NewEquipGate(req, func(sim *Simulation, _ float64) { gated.refresh(sim) })
	return gated
}

// Counts every stat level times over: 0 is off and 1 the plain amount. A nil Simulation adds them to
// the character as it is built.
func (gated *EquipGatedStats) SetLevel(sim *Simulation, level float64) {
	for i := range gated.stats {
		gated.stats[i].level = level
	}
	gated.refresh(sim)
}

// Counts one stat level times over, the way SetLevel counts them all.
func (gated *EquipGatedStats) SetStatLevel(sim *Simulation, stat stats.Stat, level float64) {
	for i := range gated.stats {
		if gated.stats[i].stat == stat {
			gated.stats[i].level = level
		}
	}
	gated.refresh(sim)
}

func (gated *EquipGatedStats) refresh(sim *Simulation) {
	equipped := gated.gate.Equipped()
	var delta stats.Stats
	changed := false
	for i := range gated.stats {
		gs := &gated.stats[i]
		want := 0.0
		if equipped {
			want = gs.level
		}
		if want == gs.appliedLevel {
			continue
		}
		delta[gs.stat] += gs.value*want - gs.value*gs.appliedLevel
		gs.appliedLevel = want
		changed = true
	}
	if !changed {
		return
	}
	if sim == nil {
		gated.character.AddStats(delta)
	} else {
		gated.character.AddStatsDynamic(sim, delta)
	}
}

// The item's bit in a spell's weapon subclass mask, or 0 for an item that is not a weapon.
func (item *Item) WeaponSubclassBit() int32 {
	switch item.RangedWeaponType {
	case proto.RangedWeaponType_RangedWeaponTypeBow:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_BOW
	case proto.RangedWeaponType_RangedWeaponTypeGun:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_GUNS
	case proto.RangedWeaponType_RangedWeaponTypeCrossbow:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_CROSSBOW
	case proto.RangedWeaponType_RangedWeaponTypeThrown:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_THROWN
	case proto.RangedWeaponType_RangedWeaponTypeWand:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_WAND
	}

	twoHand := item.HandType == proto.HandType_HandTypeTwoHand
	switch item.WeaponType {
	case proto.WeaponType_WeaponTypeAxe:
		return Ternary[int32](twoHand, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE)
	case proto.WeaponType_WeaponTypeMace:
		return Ternary[int32](twoHand, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_2H, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_1H)
	case proto.WeaponType_WeaponTypeSword:
		return Ternary[int32](twoHand, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H)
	case proto.WeaponType_WeaponTypePolearm:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM
	case proto.WeaponType_WeaponTypeStaff:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF
	case proto.WeaponType_WeaponTypeFist:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST
	case proto.WeaponType_WeaponTypeDagger:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS
	}
	return 0
}

// The item's bits in a spell's inventory-type mask while it sits in this slot. A one-hand weapon also
// counts as the main-hand or off-hand weapon of the hand holding it.
func invTypeBit(slot proto.ItemSlot, item *Item) int32 {
	switch item.RangedWeaponType {
	case proto.RangedWeaponType_RangedWeaponTypeBow:
		return dbcenums.INVTYPE_BIT_RANGED
	case proto.RangedWeaponType_RangedWeaponTypeThrown:
		return dbcenums.INVTYPE_BIT_THROWN
	case proto.RangedWeaponType_RangedWeaponTypeGun, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeWand:
		return dbcenums.INVTYPE_BIT_RANGEDRIGHT
	}

	switch item.WeaponType {
	case proto.WeaponType_WeaponTypeShield:
		return dbcenums.INVTYPE_BIT_SHIELD
	case proto.WeaponType_WeaponTypeOffHand:
		return dbcenums.INVTYPE_BIT_HOLDABLE
	}

	switch item.HandType {
	case proto.HandType_HandTypeTwoHand:
		return dbcenums.INVTYPE_BIT_2HWEAPON
	case proto.HandType_HandTypeMainHand:
		return dbcenums.INVTYPE_BIT_WEAPONMAINHAND
	case proto.HandType_HandTypeOffHand:
		return dbcenums.INVTYPE_BIT_WEAPONOFFHAND
	case proto.HandType_HandTypeOneHand:
		switch slot {
		case proto.ItemSlot_ItemSlotMainHand:
			return dbcenums.INVTYPE_BIT_WEAPON | dbcenums.INVTYPE_BIT_WEAPONMAINHAND
		case proto.ItemSlot_ItemSlotOffHand:
			return dbcenums.INVTYPE_BIT_WEAPON | dbcenums.INVTYPE_BIT_WEAPONOFFHAND
		}
		return dbcenums.INVTYPE_BIT_WEAPON
	}
	return 0
}

// The weapon subclass mask naming every weapon of these types, one- and two-handed alike.
func WeaponSubclassMask(types ...proto.WeaponType) int32 {
	var mask int32
	for _, weaponType := range types {
		for _, handType := range []proto.HandType{proto.HandType_HandTypeOneHand, proto.HandType_HandTypeTwoHand} {
			mask |= (&Item{WeaponType: weaponType, HandType: handType}).WeaponSubclassBit()
		}
	}
	return mask
}
