package database

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/tools/database/dbc"
)

func rawEnchantByName(t *testing.T, effectID int, name string) dbc.Enchant {
	t.Helper()
	for _, enchant := range dbc.GetDBC().Enchants {
		if enchant.EffectId == effectID && enchant.Name == name {
			return enchant
		}
	}
	t.Fatalf("enchant %d %q is not in the enchant inputs", effectID, name)
	return dbc.Enchant{}
}

func enchantByName(t *testing.T, effectID int, name string) *proto.UIEnchant {
	t.Helper()
	enchant := rawEnchantByName(t, effectID, name)
	return enchant.ToProto()
}

func enchantSlotPoints(t *testing.T, enchant dbc.Enchant, effect, arg int) float64 {
	t.Helper()
	for i, e := range enchant.Effects {
		if e == effect && enchant.EffectArgs[i] == arg {
			return float64(enchant.EffectPoints[i])
		}
	}
	t.Fatalf("enchant %d %q has no slot %d with argument %d", enchant.EffectId, enchant.Name, effect, arg)
	return 0
}

func enchantEquipSpell(t *testing.T, enchant dbc.Enchant) int {
	t.Helper()
	for i, e := range enchant.Effects {
		if e == dbc.ITEM_ENCHANTMENT_EQUIP_SPELL {
			return enchant.EffectArgs[i]
		}
	}
	t.Fatalf("enchant %d %q has no equip spell slot", enchant.EffectId, enchant.Name)
	return 0
}

// The amount an equip spell's aura states, read the way the generator reads it: base points plus die
// sides.
func equipAuraPoints(t *testing.T, spellID int, aura dbcenums.EffectAuraType) float64 {
	t.Helper()
	for _, e := range dbc.GetDBC().SpellEffectsInOrder(spellID) {
		if e.EffectType == dbcenums.E_APPLY_AURA && e.EffectAura == aura {
			return e.EffectBasePoints + e.EffectDieSides
		}
	}
	t.Fatalf("spell %d states no %v", spellID, aura)
	return 0
}

func equipDefenseRating(t *testing.T, spellID int) float64 {
	return equipAuraPoints(t, spellID, dbcenums.A_MOD_SKILL) * core.DefenseRatingPerDefenseLevel
}

// Which stats each slot lands on: spell power as spell damage and healing, an armor kit's resistance
// slot as bonus armor, an equip spell's defense skill as defense rating, its block value and its
// Stamina as themselves, and its mana as Mana.
func TestEnchantStats(t *testing.T) {
	withDBCInputs(t)

	spellPower := func(t *testing.T, e dbc.Enchant) stats.Stats {
		sp := enchantSlotPoints(t, e, dbc.ITEM_ENCHANTMENT_STAT, dbc.ITEM_MOD_SPELL_POWER)
		return stats.Stats{stats.SpellDamage: sp, stats.HealingPower: sp}
	}
	presence := func(t *testing.T, e dbc.Enchant) stats.Stats {
		spell := enchantEquipSpell(t, e)
		return stats.Stats{stats.Stamina: equipAuraPoints(t, spell, dbcenums.A_MOD_STAT), stats.DefenseRating: equipDefenseRating(t, spell),
			stats.BlockValue: equipAuraPoints(t, spell, dbcenums.A_MOD_BLOCK_VALUE_FLAT)}
	}
	for _, tc := range []struct {
		effectID int
		name     string
		want     func(t *testing.T, e dbc.Enchant) stats.Stats
	}{
		{7655, "Enchant Bracer - Spell Power", spellPower},
		{8211, "Enchant 2H Weapon - Mighty Spell Power", spellPower},
		{8482, "Mystic Medium Armor Kit", func(t *testing.T, e dbc.Enchant) stats.Stats {
			want := spellPower(t, e)
			want[stats.BonusArmor] = enchantSlotPoints(t, e, dbc.ITEM_ENCHANTMENT_RESISTANCE, 0)
			return want
		}},
		{923, "Enchant Bracer - Deflection", func(t *testing.T, e dbc.Enchant) stats.Stats {
			return stats.Stats{stats.DefenseRating: equipDefenseRating(t, enchantEquipSpell(t, e))}
		}},
		{2583, "Presence of Might", presence},
		{7633, "Presence of Valor", presence},
		{8214, "Enchant Bracer - Superior Deflection", func(t *testing.T, e dbc.Enchant) stats.Stats {
			return stats.Stats{stats.DefenseRating: enchantSlotPoints(t, e, dbc.ITEM_ENCHANTMENT_STAT, dbc.ITEM_MOD_DEFENSE_SKILL_RATING)}
		}},
		{8719, "Wild Leather Armor Kit", func(t *testing.T, e dbc.Enchant) stats.Stats {
			return stats.Stats{stats.DefenseRating: enchantSlotPoints(t, e, dbc.ITEM_ENCHANTMENT_STAT, dbc.ITEM_MOD_DEFENSE_SKILL_RATING),
				stats.Stamina: enchantSlotPoints(t, e, dbc.ITEM_ENCHANTMENT_STAT, dbc.ITEM_MOD_STAMINA)}
		}},
		{1483, "Lesser Arcane Amalgamation", func(t *testing.T, e dbc.Enchant) stats.Stats {
			return stats.Stats{stats.Mana: equipAuraPoints(t, enchantEquipSpell(t, e), dbcenums.A_MOD_INCREASE_ENERGY)}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := rawEnchantByName(t, tc.effectID, tc.name)
			want := tc.want(t, raw)
			if got := stats.FromProtoArray(raw.ToProto().Stats); got != want {
				t.Errorf("stats %v, want %v", got.FlatString(), want.FlatString())
			}
		})
	}
}

// Enchant 1483, applied to a head or legs item by Lesser Arcanum of Rumination (11622), ships: it is
// live and the global filters keep it.
func TestLesserArcanumOfRuminationShips(t *testing.T) {
	withDBCInputs(t)

	raw := rawEnchantByName(t, 1483, "Lesser Arcane Amalgamation")
	if !raw.IsLive {
		t.Error("reads not live")
	}
	if !KeepEnchant(raw.ToProto()) {
		t.Error("the global filters drop it")
	}
}

// A damage slot lands as weapon damage alone.
func TestEnchantWeaponDamage(t *testing.T) {
	withDBCInputs(t)

	for _, tc := range []struct {
		effectID int
		name     string
		damage   bool
	}{
		{963, "Enchant 2H Weapon - Greater Impact", true},
		{1897, "Enchant Weapon - Superior Striking", true},
		{664, "Sniper Scope", true},
		{7944, "Poultry Precision Scope", true},
		{7655, "Enchant Bracer - Spell Power", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := rawEnchantByName(t, tc.effectID, tc.name)
			want := 0.0
			if tc.damage {
				want = enchantSlotPoints(t, raw, dbc.ITEM_ENCHANTMENT_DAMAGE, 0)
			}
			enchant := raw.ToProto()
			if enchant.WeaponDamage != want {
				t.Errorf("weapon damage %v, want %v", enchant.WeaponDamage, want)
			}
			if tc.damage && stats.FromProtoArray(enchant.Stats) != (stats.Stats{}) {
				t.Errorf("stats %v, want none", enchant.Stats)
			}
		})
	}
}

// Which haste channels an equip spell's aura reaches: A_MOD_MELEE_HASTE_3 melee alone,
// A_MOD_MELEE_RANGED_HASTE_2 melee and ranged, A_MOD_RANGED_HASTE ranged, and
// A_MOD_CASTING_SPEED_NOT_STACK spells.
func TestEnchantHaste(t *testing.T) {
	withDBCInputs(t)

	for _, tc := range []struct {
		effectID             int
		name                 string
		melee, ranged, spell dbcenums.EffectAuraType
	}{
		{34, "Weapon Counterweight", dbcenums.A_MOD_MELEE_HASTE_3, 0, 0},
		{931, "Enchant Gloves - Minor Haste", dbcenums.A_MOD_MELEE_RANGED_HASTE_2, dbcenums.A_MOD_MELEE_RANGED_HASTE_2, dbcenums.A_MOD_CASTING_SPEED_NOT_STACK},
		{2543, "Arcanum of Rapidity", dbcenums.A_MOD_MELEE_HASTE_3, dbcenums.A_MOD_RANGED_HASTE, 0},
		{7123, "Automatic Crowd Pummeler", dbcenums.A_MOD_MELEE_HASTE_3, 0, 0},
		{7655, "Enchant Bracer - Spell Power", 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := rawEnchantByName(t, tc.effectID, tc.name)
			enchant := raw.ToProto()
			pseudoStat := func(pseudoStat proto.PseudoStat) float64 {
				if int(pseudoStat) < len(enchant.PseudoStats) {
					return enchant.PseudoStats[pseudoStat]
				}
				return 0
			}
			amount := func(aura dbcenums.EffectAuraType) float64 {
				if aura == 0 {
					return 0
				}
				return equipAuraPoints(t, enchantEquipSpell(t, raw), aura)
			}
			for _, haste := range []struct {
				pseudoStat proto.PseudoStat
				aura       dbcenums.EffectAuraType
			}{
				{proto.PseudoStat_PseudoStatMeleeHastePercent, tc.melee},
				{proto.PseudoStat_PseudoStatRangedHastePercent, tc.ranged},
				{proto.PseudoStat_PseudoStatSpellHastePercent, tc.spell},
			} {
				if got, want := pseudoStat(haste.pseudoStat), amount(haste.aura); got != want {
					t.Errorf("%s %v, want %v", haste.pseudoStat, got, want)
				}
			}
			if tc.melee != 0 && stats.FromProtoArray(enchant.Stats) != (stats.Stats{}) {
				t.Errorf("stats %v, want none", enchant.Stats)
			}
		})
	}
}

// The slot a non-weapon enchant takes, from its subclass mask and inventory types.
func TestEnchantSlot(t *testing.T) {
	for _, tc := range []struct {
		name        string
		enchant     dbc.Enchant
		itemType    proto.ItemType
		enchantType proto.EnchantType
		extraTypes  []proto.ItemType
	}{
		{"shield by its subclass", dbc.Enchant{SubClassMask: dbc.ITEM_SUBCLASS_BIT_ARMOR_SHIELD},
			proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeShield, nil},
		{"shield spike, the shield and buckler bits", dbc.Enchant{SubClassMask: dbc.ITEM_SUBCLASS_BIT_ARMOR_SHIELD | dbc.ITEM_SUBCLASS_BIT_ARMOR_BUCKLER},
			proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeShield, nil},
		{"shield by the off-hand inventory type alone", dbc.Enchant{InventoryType: dbc.OFF_HAND},
			proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeShield, nil},
		{"off-hand by the held in off-hand inventory type", dbc.Enchant{InventoryType: dbc.HOLDABLE},
			proto.ItemType_ItemTypeWeapon, proto.EnchantType_EnchantTypeOffHand, nil},
		{"cloak", dbc.Enchant{InventoryType: dbc.CLOAK},
			proto.ItemType_ItemTypeBack, proto.EnchantType_EnchantTypeNormal, nil},
		{"head or legs", dbc.Enchant{InventoryType: dbc.HEAD | dbc.LEGS},
			proto.ItemType_ItemTypeHead, proto.EnchantType_EnchantTypeNormal, []proto.ItemType{proto.ItemType_ItemTypeLegs}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			enchant := tc.enchant.ToProto()
			if enchant.Type != tc.itemType || enchant.EnchantType != tc.enchantType || !slices.Equal(enchant.ExtraTypes, tc.extraTypes) {
				t.Errorf("type %v, enchant type %v, extra types %v; want %v, %v, %v",
					enchant.Type, enchant.EnchantType, enchant.ExtraTypes, tc.itemType, tc.enchantType, tc.extraTypes)
			}
		})
	}
}

// The global filters drop an enchant whose item lies in the Season of Discovery range, both ends
// included, and keep one on either side of it.
func TestEnchantSourcesOutsideTheSoDItemRange(t *testing.T) {
	for _, tc := range []struct {
		name   string
		itemID int32
		keep   bool
	}{
		{"below the range", sodItemIDMin - 1, true},
		{"the range's first item", sodItemIDMin, false},
		{"inside the range", (sodItemIDMin + sodItemIDMax) / 2, false},
		{"the range's last item", sodItemIDMax, false},
		{"above the range", sodItemIDMax + 1, true},
	} {
		enchant := &proto.UIEnchant{EffectId: 990001, Name: "Enchant Weapon - Crusader", ItemId: tc.itemID}
		if got := KeepEnchant(enchant); got != tc.keep {
			t.Errorf("%s: item %d kept %v, want %v", tc.name, tc.itemID, got, tc.keep)
		}
	}
}

func TestWeaponEnchantSubclassType(t *testing.T) {
	for _, tc := range []struct {
		name         string
		subClassMask int
		want         proto.EnchantType
	}{
		{"staff alone", dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF, proto.EnchantType_EnchantTypeStaff},
		{"two-hand axe, mace, polearm and sword", 354, proto.EnchantType_EnchantTypeTwoHand},
		{"two-hand sword and one-hand sword", dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H, proto.EnchantType_EnchantTypeNormal},
		{"any weapon", 0, proto.EnchantType_EnchantTypeNormal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			enchant := dbc.Enchant{IsWeaponEnchant: true, SubClassMask: tc.subClassMask}
			if got := enchant.ToProto().EnchantType; got != tc.want {
				t.Errorf("enchant type %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEquipSpellStats(t *testing.T) {
	withDBCInputs(t)
	dbc.GetDBC()

	for _, tc := range []struct {
		name    string
		spellID int
		want    func(t *testing.T, spellID int) stats.Stats
	}{
		{"13922 Enchant Bracer - Deflection, defense skill as defense rating", 13922, func(t *testing.T, spellID int) stats.Stats {
			return stats.Stats{stats.DefenseRating: equipDefenseRating(t, spellID)}
		}},
		{"24148 Presence of Might, defense skill and block value; its Stamina is not this reader's", 24148, func(t *testing.T, spellID int) stats.Stats {
			return stats.Stats{stats.DefenseRating: equipDefenseRating(t, spellID), stats.BlockValue: equipAuraPoints(t, spellID, dbcenums.A_MOD_BLOCK_VALUE_FLAT)}
		}},
		{"7778 Enchant Boots - Minor Agility adds none", 7778, func(*testing.T, int) stats.Stats { return stats.Stats{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.want(t, tc.spellID)
			got := stats.Stats{}
			added, _ := dbc.AddEquipSpellStats(&got, nil, tc.spellID)
			if got != want {
				t.Errorf("stats %v, want %v", got.FlatString(), want.FlatString())
			}
			if added != (want != stats.Stats{}) {
				t.Errorf("reported added=%v", added)
			}
		})
	}
}

// Healing items state spell power (45) and their extra healing (41) in separate slots, and the
// two add up: 240040 Inquisition Headpiece.
func TestItemSpellPowerAddsToHealing(t *testing.T) {
	withDBCInputs(t)

	item, ok := dbc.GetDBC().Items[240040]
	if !ok {
		t.Fatal("item 240040 is not in the item inputs")
	}
	slot := func(mod int) float64 {
		i := slices.Index(item.BonusStat, mod)
		if i < 0 {
			t.Fatalf("item 240040 states no stat %d", mod)
		}
		return item.GetScaledStat(i, item.ItemLevel)
	}
	spellPower, healing := slot(dbc.ITEM_MOD_SPELL_POWER), slot(dbc.ITEM_MOD_SPELL_HEALING_DONE)
	if spellPower == 0 || healing == 0 {
		t.Fatalf("item 240040 states spell power %v and healing %v, want both", spellPower, healing)
	}
	got := item.GetStats(item.ItemLevel)
	if got[stats.SpellDamage] != spellPower || got[stats.HealingPower] != spellPower+healing {
		t.Errorf("spell damage %v and healing %v, want %v and %v", got[stats.SpellDamage], got[stats.HealingPower], spellPower, spellPower+healing)
	}
}
