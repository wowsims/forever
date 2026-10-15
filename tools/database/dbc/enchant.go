package dbc

import (
	"fmt"
	"slices"
	"sort"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/tools/database/overrides"
)

type Enchant struct {
	EffectId           int
	Name               string
	SpellId            int
	ItemId             int
	ProfessionId       int
	Effects            []int
	EffectPoints       []int
	EffectArgs         []int
	IsWeaponEnchant    bool
	InventoryType      InventoryTypeFlag
	SubClassMask       int
	ClassMask          int
	FDID               int
	Quality            ItemQuality
	RequiredProfession int
	EffectName         string
	IsLive             bool
	Lvl                int
}

// A spell an enchant slot casts off a hit, or hangs a listener on.
type EnchantProcSlot struct {
	SpellID int
	// A combat spell (Effect 1), which the game casts off the weapon's hit rather than through a
	// proc mask.
	IsCombatSpell bool
	// The chance in percent a combat spell slot states for it, 0 where it states none.
	ChancePct int
	// What the slot applies: the combat spell itself, or the spell its equip aura triggers.
	AppliesSpellID int
	// The effect entry the slot's spell resolves to, where it resolves stats.
	Effect *proto.ItemEffect
}

// The auras through which an equip spell answers a hit: the proc triggers, the retaliation of a
// damage shield, and the dummy a server-side proc hangs off.
var EnchantProcAuras = []EffectAuraType{
	dbcenums.A_PROC_TRIGGER_SPELL, dbcenums.A_PROC_TRIGGER_SPELL_WITH_VALUE, dbcenums.A_PROC_TRIGGER_SPELL_COPY,
	dbcenums.A_PROC_TRIGGER_DAMAGE, dbcenums.A_DAMAGE_SHIELD, dbcenums.A_DUMMY,
}

// Every combat spell (Effect 1) and every equip spell (Effect 3) answering a hit, one per slot.
func (enchant *Enchant) ProcSlots() []EnchantProcSlot {
	var slots []EnchantProcSlot
	for idx, effect := range enchant.Effects {
		if idx >= len(enchant.EffectArgs) || enchant.EffectArgs[idx] == 0 {
			continue
		}
		spellID := enchant.EffectArgs[idx]

		switch effect {
		case ITEM_ENCHANTMENT_COMBAT_SPELL:
			slot := EnchantProcSlot{SpellID: spellID, IsCombatSpell: true, AppliesSpellID: spellID}
			if idx < len(enchant.EffectPoints) {
				slot.ChancePct = enchant.EffectPoints[idx]
			}
			slots = append(slots, slot)
		case ITEM_ENCHANTMENT_EQUIP_SPELL:
			for _, spellEffect := range dbcInstance.SpellEffectsInOrder(spellID) {
				if slices.Contains(EnchantProcAuras, spellEffect.EffectAura) {
					slots = append(slots, EnchantProcSlot{SpellID: spellID, AppliesSpellID: spellEffect.EffectTriggerSpell})
					break
				}
			}
		}
	}

	for i := range slots {
		eff := ItemEffect{TriggerType: ITEM_SPELLTRIGGER_CHANCE_ON_HIT, SpellID: slots[i].SpellID}
		slots[i].Effect, _ = eff.ToProto(0)
	}

	return slots
}

// The slot's effect carrying a hand-kept rate, refused once the game files state a chance, so the
// number does not outlive its reason. A slot whose spell resolves no stats gets an entry of its own,
// which carries the rate alone.
func withEnchantProcRate(effectID int, slot EnchantProcSlot, rate *overrides.EnchantProc) *proto.ItemEffect {
	if chance := dbcInstance.Spells[slot.SpellID].ProcChance; slot.ChancePct > 0 || (chance > 0 && chance < 100) {
		panic(fmt.Sprintf("enchant %d's spell %d now states a chance, so its hand-supplied rate is stale", effectID, slot.SpellID))
	}
	effect := slot.Effect
	if effect == nil {
		effect = &proto.ItemEffect{BuffId: int32(slot.SpellID), Effect: &proto.ItemEffect_Proc{Proc: &proto.ProcEffect{}}}
	}
	proc := effect.GetProc()
	switch rate.Kind {
	case overrides.RatePPM:
		proc.ProcRate = &proto.ProcEffect_Ppm{Ppm: rate.Value}
	case overrides.RateChance:
		proc.ProcRate = &proto.ProcEffect_ProcChance{ProcChance: rate.Value}
	}
	return effect
}

// SpellItemEnchantment 8203 is "Spirit +$k1", applied by Enchant Bracer/Boots - Lesser Spirit, but
// its EffectArg is 124, the all-resistances index. The label and the enchant are right and the
// argument is not, so it is corrected here rather than read as five resistances.
var enchantEffectArgFixes = map[int]map[int]int{
	8203: {0: ITEM_MOD_SPIRIT},
}

// The effects an enchant's slots resolve stats or a hand-kept rate for, one per buff.
func (enchant *Enchant) SlotEffects(slots []EnchantProcSlot) []*proto.ItemEffect {
	var effects []*proto.ItemEffect
	for _, slot := range slots {
		effect := slot.Effect
		if rate := overrides.EnchantProcRate(int32(enchant.EffectId), int32(slot.SpellID)); rate != nil {
			effect = withEnchantProcRate(enchant.EffectId, slot, rate)
		}
		if effect != nil && !slices.ContainsFunc(effects, func(e *proto.ItemEffect) bool { return e.BuffId == effect.BuffId }) {
			effects = append(effects, effect)
		}
	}
	return effects
}

// Enchants the proc rules classify wrongly, keyed by effect id.
var enchantProcKindOverrides = map[int]proto.EnchantProcKind{
	// Lotus Claw: its equip spell 1294053, a chance at Death Lotus when gathering herbs, is an A_DUMMY with no SpellAuraOptions row.
	8695: proto.EnchantProcKind_EnchantProcKindNone,
}

// An enchant is Survival when every proc slot's spell only absorbs, heals or adds armor, and Combat
// otherwise. An enchant with no proc slot is None.
func (enchant *Enchant) procKind(slots []EnchantProcSlot) proto.EnchantProcKind {
	if kind, ok := enchantProcKindOverrides[enchant.EffectId]; ok {
		return kind
	}
	if len(slots) == 0 {
		return proto.EnchantProcKind_EnchantProcKindNone
	}
	for _, slot := range slots {
		if !appliesSurvivalOnly(slot.AppliesSpellID) {
			return proto.EnchantProcKind_EnchantProcKindCombat
		}
	}
	return proto.EnchantProcKind_EnchantProcKindSurvival
}

func appliesSurvivalOnly(spellID int) bool {
	effects := dbcInstance.SpellEffectsInOrder(spellID)
	if len(effects) == 0 {
		return false
	}
	for _, effect := range effects {
		switch {
		case effect.EffectType == dbcenums.E_HEAL, effect.EffectType == dbcenums.E_HEAL_PCT:
		case effect.EffectAura == dbcenums.A_SCHOOL_ABSORB:
		case (effect.EffectAura == dbcenums.A_MOD_RESISTANCE || effect.EffectAura == dbcenums.A_MOD_ARMOR_NO_MODS) &&
			len(effect.EffectMiscValues) > 0 && effect.EffectMiscValues[0] == int(core.SpellSchoolPhysical):
		default:
			return false
		}
	}
	return true
}

func (enchant *Enchant) ToProto() *proto.UIEnchant {
	uiEnchant, _ := enchant.ToProtoWithSlots()
	return uiEnchant
}

func (enchant *Enchant) ToProtoWithSlots() (*proto.UIEnchant, []EnchantProcSlot) {
	uiEnchant := &proto.UIEnchant{
		Name:               enchant.Name,
		ItemId:             int32(enchant.ItemId),
		SpellId:            int32(enchant.SpellId),
		EffectId:           int32(enchant.EffectId),
		ClassAllowlist:     GetClassesFromClassMask(enchant.ClassMask),
		ExtraTypes:         []proto.ItemType{},
		Stats:              stats.Stats{}.ToProtoArray(),
		Quality:            enchant.Quality.ToProto(),
		Lvl:                int32(enchant.Lvl),
		RequiredProfession: GetProfession(enchant.RequiredProfession),
	}

	slots := enchant.ProcSlots()
	uiEnchant.ProcKind = enchant.procKind(slots)

	if enchant.FDID == 0 {
		uiEnchant.Icon = "trade_engraving"
	}

	if enchant.IsWeaponEnchant {
		// Process weapon enchants.
		uiEnchant.Type = proto.ItemType_ItemTypeWeapon
		if enchant.SubClassMask == dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF {
			// Staff only.
			uiEnchant.EnchantType = proto.EnchantType_EnchantTypeStaff
		} else if enchant.SubClassMask != 0 && enchant.SubClassMask&^allTwoHandMask == 0 {
			// Two-handed weapons only, whichever of them the mask names.
			uiEnchant.EnchantType = proto.EnchantType_EnchantTypeTwoHand
		}
		if enchant.SubClassMask == dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING {
			uiEnchant.Type = proto.ItemType_ItemTypeRanged
		}
	} else {
		// Process non-weapon enchants.
		if enchant.SubClassMask == OffHandValue || enchant.InventoryType.Has(HOLDABLE) {
			uiEnchant.EnchantType = proto.EnchantType_EnchantTypeOffHand
			uiEnchant.Type = proto.ItemType_ItemTypeWeapon
		}
		// Shield enchants target the shield subclass alone; shield spikes also set the obsolete buckler bit (mask 96).
		// Others name the shield inventory type instead (OFF_HAND, 0x4000) with no subclass: 7663.
		if enchant.SubClassMask&ITEM_SUBCLASS_BIT_ARMOR_SHIELD != 0 || enchant.InventoryType.Has(OFF_HAND) {
			uiEnchant.EnchantType = proto.EnchantType_EnchantTypeShield
			uiEnchant.Type = proto.ItemType_ItemTypeWeapon
		}
		// Sort flags for consistent generation
		var flags []int
		for flag := range MapInventoryTypeToEnchantMetaType {
			flags = append(flags, int(flag))
		}

		sort.Ints(flags)

		for _, f := range flags {
			flag := InventoryTypeFlag(f)
			m := MapInventoryTypeToEnchantMetaType[flag]
			if enchant.InventoryType.Has(flag) {
				if uiEnchant.Type != proto.ItemType_ItemTypeUnknown {
					uiEnchant.ExtraTypes = append(uiEnchant.ExtraTypes, m.ItemType)
				} else {
					uiEnchant.Type = m.ItemType
				}
			}
		}
		slices.Sort(uiEnchant.ExtraTypes)
	}
	effectArgs := enchant.EffectArgs
	if fixes, ok := enchantEffectArgFixes[enchant.EffectId]; ok {
		effectArgs = slices.Clone(effectArgs)
		for index, arg := range fixes {
			effectArgs[index] = arg
		}
	}
	enchantStats := stats.Stats{}
	pseudoStats := make([]float64, stats.PseudoStatsLen)
	processEnchantmentEffects(enchant.Effects, effectArgs, enchant.EffectPoints, &enchantStats, pseudoStats, true)
	uiEnchant.Stats = enchantStats.ToProtoArray()
	uiEnchant.PseudoStats = NullFloat(pseudoStats)
	for i, effect := range enchant.Effects {
		if effect == ITEM_ENCHANTMENT_DAMAGE {
			uiEnchant.WeaponDamage += float64(enchant.EffectPoints[i])
		}
	}
	return uiEnchant, slots
}
