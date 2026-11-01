package dbc

import (
	"slices"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type Consumable struct {
	Id                 int             // Item ID
	Name               string          // Item name
	ItemLevel          int             // Item level
	RequiredLevel      int             // Required level to use
	ClassId            int             // Item class ID (should be 0 for consumables)
	SubClassId         ConsumableClass // Item subclass ID
	IconFileDataID     int             // Icon file data ID
	SpellCategoryID    int             // Spell category ID
	SpellCategoryFlags int             // Spell category flags
	ItemEffects        []int           // Item effect IDs
	ClassMask          int
	RaceMask           int64
	RequiredSkill      int
	RequiredSkillRank  int
	RequiredHoliday    int // ItemSparse.RequiredHoliday, 0 when the item names no world event.
	TemporaryEnchant   TemporaryEnchant
}

// The SpellItemEnchantment row a consumable's temporary enchant (Effect 54, or the imbue Effect 360)
// applies, by its three slots. ID is 0 when the consumable enchants nothing.
type TemporaryEnchant struct {
	ID      int
	Effects [3]int
	Points  [3]int
	Args    [3]int
	// SpellCastingRequirements.RequiredAreasID of the spells the slots name: the area group outside
	// which they do nothing. 0 when they work everywhere.
	RequiredAreasID int
}

func IsTemporaryEnchantEffect(effect dbcenums.SpellEffectType) bool {
	return effect == dbcenums.E_ENCHANT_ITEM_TEMPORARY || effect == dbcenums.E_360
}

// The enchant as the enchant table reads a permanent one: stats and pseudo stats from its stat,
// resistance and equip spell slots, and its flat weapon damage. A combat spell slot that states a
// chance adds the spell it casts off the enchanted weapon's hits, by its id alone; the store writes
// the chance onto that spell's row.
func (t TemporaryEnchant) ToProto() *proto.SimEnchant {
	enchant := Enchant{EffectId: t.ID, Effects: t.Effects[:], EffectPoints: t.Points[:], EffectArgs: t.Args[:], IsWeaponEnchant: true}
	parsed, slots := enchant.ToProtoWithSlots()
	simEnchant := &proto.SimEnchant{
		EffectId:       parsed.EffectId,
		Stats:          parsed.Stats,
		PseudoStats:    parsed.PseudoStats,
		WeaponDamage:   parsed.WeaponDamage,
		EnchantEffects: parsed.EnchantEffects,
	}
	for _, slot := range slots {
		spellID := int32(slot.SpellID)
		if !slot.IsCombatSpell || slot.ChancePct <= 0 || slices.ContainsFunc(simEnchant.EnchantEffects, func(e *proto.ItemEffect) bool {
			return e.BuffId == spellID
		}) {
			continue
		}
		simEnchant.EnchantEffects = append(simEnchant.EnchantEffects, &proto.ItemEffect{BuffId: spellID})
	}
	return simEnchant
}

func (c *Consumable) ToMap() map[string]any {
	return map[string]any{
		"Id":                 c.Id,
		"Name":               c.Name,
		"ItemLevel":          c.ItemLevel,
		"RequiredLevel":      c.RequiredLevel,
		"ClassId":            c.ClassId,
		"SubClassId":         c.SubClassId,
		"IconFileDataID":     c.IconFileDataID,
		"SpellCategoryID":    c.SpellCategoryID,
		"SpellCategoryFlags": c.SpellCategoryFlags,
		"ItemEffects":        c.ItemEffects,
	}
}

// ToProto converts the Consumable to a proto representation.
func (c *Consumable) ToProto() *proto.Consumable {
	onUses := c.OnUseEffects()
	enchant, enchantEffect := temporaryEnchantSpell(onUses)
	category := onUseCategory(onUses)
	consumable := &proto.Consumable{
		Id:                 int32(c.Id),
		Type:               c.consumableType(enchantEffect, category),
		Stats:              c.GetStatModifiers().ToProtoArray(),
		Name:               c.Name,
		Lvl:                int32(c.ItemLevel),
		BuffsMainStat:      false, // Todo: Should be food currently, might be more in MoP, figure out how to tell
		CategoryId:         int32(c.SpellCategoryID),
		EffectIds:          c.GetNonStatEffectIds(),
		ItemEffects:        onUses,
		ClassAllowlist:     GetClassesFromClassMask(c.ClassMask),
		RequiredProfession: GetProfession(c.RequiredSkill),
	}
	if category > 0 {
		consumable.CategoryId = category
	}
	// A required level of 1 requires nothing: the EZ-Thro and SAF-T bombs and the Restored potions
	// state 1 beside item levels up to 60.
	if c.RequiredLevel > 1 {
		consumable.Lvl = int32(c.RequiredLevel)
	}
	if consumable.RequiredProfession != proto.Profession_ProfessionUnknown {
		consumable.ReqProfRank = int32(c.RequiredSkillRank)
	}
	if c.RaceMask != -1 && c.RaceMask != 0 {
		consumable.RaceAllowlist = GetRacesFromRaceMask(c.RaceMask)
	}
	if enchant != nil && enchant.EquippedItemClass == ITEM_CLASS_WEAPON {
		consumable.WeaponSubclassMask = enchant.EquippedItemSubclass
		consumable.WeaponInventoryTypeMask = enchant.EquippedItemInvTypes
	}
	if c.TemporaryEnchant.ID != 0 && enchant != nil {
		consumable.TemporaryEnchant = c.TemporaryEnchant.ToProto()
	}
	return consumable
}

// The spell each on-use row casts, with the cooldowns and category the item states for it.
func (c *Consumable) OnUseEffects() []*proto.ItemEffect {
	var effects []*proto.ItemEffect
	for _, id := range slices.Sorted(slices.Values(c.ItemEffects)) {
		e := GetItemEffect(id)
		if e.ID == 0 || e.TriggerType != ITEM_SPELLTRIGGER_ON_USE || e.SpellID == 0 {
			continue
		}
		effects = append(effects, &proto.ItemEffect{
			BuffId: int32(e.SpellID),
			Effect: &proto.ItemEffect_OnUse{OnUse: &proto.OnUseEffect{
				CooldownMs:         int32(e.CoolDownMSec),
				CategoryId:         int32(e.SpellCategoryID),
				CategoryCooldownMs: int32(e.CategoryCoolDownMSec),
			}},
		})
	}
	return effects
}

func temporaryEnchantSpell(onUses []*proto.ItemEffect) (*Spell, dbcenums.SpellEffectType) {
	for _, e := range onUses {
		for _, se := range dbcInstance.SpellEffects[int(e.BuffId)] {
			if IsTemporaryEnchantEffect(se.EffectType) {
				spell := dbcInstance.Spells[int(e.BuffId)]
				return &spell, se.EffectType
			}
		}
	}
	return nil, 0
}

// The category the first on-use shares: the item's row, or where it states none, its spell's. The
// Spellblasting, Frenzy and Mender's potions state theirs on the spell alone.
func onUseCategory(onUses []*proto.ItemEffect) int32 {
	if len(onUses) == 0 {
		return 0
	}
	if category := onUses[0].GetOnUse().CategoryId; category > 0 {
		return category
	}
	return dbcInstance.Spells[int(onUses[0].BuffId)].Category
}

// Whether an on-use row shares the explosives' category 24.
func (c *Consumable) SharesExplosivesCategory() bool {
	return slices.ContainsFunc(c.OnUseEffects(), func(e *proto.ItemEffect) bool {
		return e.GetOnUse().CategoryId == EXPLOSIVES_SPELL_CATEGORY
	})
}

// A consumable's type, which the settings pick it by: an imbue by what it does, one sharing the potion
// or conjured category a potion or conjured item (Potion of Demon Slaying 9224 is an Elixir-subclass
// item whose spell 11406 shares category 4), and otherwise its item subclass, with what shares the
// food or drink category food. Whether it is an explosive takes how its spell routes, which the
// database package decides.
func (c *Consumable) consumableType(enchantEffect dbcenums.SpellEffectType, category int32) proto.ConsumableType {
	switch {
	case enchantEffect == dbcenums.E_360:
		return proto.ConsumableType_ConsumableTypeWeaponImbue
	case enchantEffect != 0:
		return proto.ConsumableType_ConsumableTypeImbue
	case category == POTION_SPELL_CATEGORY:
		return proto.ConsumableType_ConsumableTypePotion
	case category == CONJURED_SPELL_CATEGORY:
		return proto.ConsumableType_ConsumableTypeConjured
	}
	itemType := c.itemType()
	switch itemType {
	case proto.ConsumableType_ConsumableTypePotion, proto.ConsumableType_ConsumableTypeFlask, proto.ConsumableType_ConsumableTypeElixir:
		return itemType
	}
	if category == FOOD_SPELL_CATEGORY || category == DRINK_SPELL_CATEGORY {
		return proto.ConsumableType_ConsumableTypeFood
	}
	return itemType
}

func (c *Consumable) itemType() proto.ConsumableType {
	if c.ClassId != ITEM_CLASS_CONSUMABLE {
		return proto.ConsumableType_ConsumableTypeUnknown
	}
	if c.SubClassId == ELIXIR {
		return proto.ConsumableType_ConsumableTypeElixir
	}
	if c.SubClassId == FOOD {
		return proto.ConsumableType_ConsumableTypeFood
	}
	if val, ok := consumableClassToProto[c.SubClassId]; ok {
		return val
	}
	return proto.ConsumableType_ConsumableTypeUnknown
}

func (s ConsumableClass) ToProto() proto.ConsumableType {
	if val, ok := consumableClassToProto[s]; ok {
		return val
	}
	return proto.ConsumableType_ConsumableTypeUnknown
}

// A consumable's effects that restore a resource rather than grant a stat, which the sim wires up
// separately.
var consumableRestoreEffectTypes = map[SpellEffectType]bool{
	dbcenums.E_HEAL:         true,
	dbcenums.E_HEAL_PCT:     true,
	dbcenums.E_ENERGIZE:     true,
	dbcenums.E_ENERGIZE_PCT: true,
}

func (consumable *Consumable) GetNonStatEffectIds() []int32 {
	var effectIds []int32

	slices.Sort(consumable.ItemEffects)
	for _, effectID := range consumable.ItemEffects {
		effect := GetItemEffect(effectID)
		if effect.ID != 0 {
			if spellEffects, ok := dbcInstance.SpellEffects[effect.SpellID]; ok {
				for _, spellEffect := range spellEffects {
					if consumableRestoreEffectTypes[spellEffect.EffectType] || spellEffect.EffectAura == dbcenums.A_PERIODIC_ENERGIZE || spellEffect.EffectAura == dbcenums.A_PERIODIC_HEAL {
						effectIds = append(effectIds, int32(spellEffect.ID))
					}
					// Aura effects triggered on use (e.g. Fel Mana Potion's spell damage
					// reduction) carry the stats the sim applies as a temporary aura.
					if spellEffect.EffectType == dbcenums.E_TRIGGER_SPELL {
						for _, subEffect := range dbcInstance.SpellEffects[spellEffect.EffectTriggerSpell] {
							if _, ok := subEffect.ParseStatEffect(false, 0); ok {
								effectIds = append(effectIds, int32(subEffect.ID))
							}
						}
					}
				}
			}
		}
	}
	slices.Sort(effectIds)
	return effectIds
}
func (consumable *Consumable) GetStatModifiers() *stats.Stats {
	stats := &stats.Stats{}
	for _, effectID := range consumable.ItemEffects {
		effect := GetItemEffect(effectID)
		if effect.ID != 0 {
			if spellEffects, ok := dbcInstance.SpellEffects[effect.SpellID]; ok {
				for _, spellEffect := range spellEffects {
					if stat, ok := spellEffect.ParseStatEffect(spellEffect.Coefficient != 0, 0); ok {
						stats.AddInplace(&stat)
					}
					// A periodic trigger with a value hands its amount to every effect of the spell it
					// triggers: Nutritious Food 1248401 states Well Fed 1248422's Strength that way.
					if spellEffect.EffectAura == dbcenums.A_PERIODIC_TRIGGER_SPELL_WITH_VALUE {
						for _, triggered := range dbcInstance.SpellEffects[spellEffect.EffectTriggerSpell] {
							triggered.EffectBasePoints = spellEffect.EffectBasePoints
							if stat, ok := triggered.ParseStatEffect(false, 0); ok {
								stats.AddInplace(&stat)
							}
						}
					}
				}
			}
		}
	}
	return stats
}
