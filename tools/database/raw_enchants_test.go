package database

// Runs the raw enchant query on a fixture shaped like the game files, holding only the columns it
// reads, so it needs no game files database.

import (
	"database/sql"
	"path/filepath"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

const rawEnchantSchema = `
CREATE TABLE Spell (ID INTEGER PRIMARY KEY, NameSubtext_lang TEXT);
CREATE TABLE SpellName (ID INTEGER PRIMARY KEY, Name_lang TEXT);
CREATE TABLE SpellEffect (ID INTEGER PRIMARY KEY, SpellID INTEGER, Effect INTEGER, EffectMiscValue_0 INTEGER);
CREATE TABLE SpellItemEnchantment (
	ID INTEGER PRIMARY KEY, Name_lang TEXT, Flags INTEGER, RequiredSkillID INTEGER, RequiredSkillRank INTEGER,
	Effect TEXT, EffectPointsMin TEXT, EffectArg TEXT,
	Effect_0 INTEGER GENERATED ALWAYS AS (json_extract(Effect, '$[0]')),
	Effect_1 INTEGER GENERATED ALWAYS AS (json_extract(Effect, '$[1]')),
	Effect_2 INTEGER GENERATED ALWAYS AS (json_extract(Effect, '$[2]')),
	EffectArg_0 INTEGER GENERATED ALWAYS AS (json_extract(EffectArg, '$[0]')),
	EffectArg_1 INTEGER GENERATED ALWAYS AS (json_extract(EffectArg, '$[1]')),
	EffectArg_2 INTEGER GENERATED ALWAYS AS (json_extract(EffectArg, '$[2]'))
);
CREATE TABLE ItemEffect (ID INTEGER PRIMARY KEY, SpellID INTEGER);
CREATE TABLE ItemXItemEffect (ID INTEGER PRIMARY KEY, ItemID INTEGER, ItemEffectID INTEGER);
CREATE TABLE SpellEquippedItems (ID INTEGER PRIMARY KEY, SpellID INTEGER, EquippedItemClass INTEGER, EquippedItemInvTypes INTEGER, EquippedItemSubclass INTEGER);
CREATE TABLE SkillLineAbility (ID INTEGER PRIMARY KEY, Spell INTEGER, ClassMask INTEGER, SkillLine INTEGER);
CREATE TABLE Item (ID INTEGER PRIMARY KEY, IconFileDataID INTEGER);
CREATE TABLE ItemSparse (ID INTEGER PRIMARY KEY, Display_lang TEXT, OverallQualityID INTEGER, ItemLevel INTEGER DEFAULT 1, AllowableClass INTEGER DEFAULT -1);
`

// Scourge shoulder enchant 2717 (grant 29467, a profession ability) and its re-issue 7884 (grant
// 1219512, on an item without an ItemSparse row) share a name. Enchant 8203 is granted by a bracer
// and a boots recipe under two names; each is taught by two items, and the boots' lower item has no
// ItemSparse row. Enchant 1523 (grant 15427) is used only by a placeholder ([PH]) item. Grand
// Crusader 7940 is taught by a Season of Discovery formula of item level 80 and applied by a level 1
// scroll; Minor Agility 247 has only a level 1 scroll, and Superior Intellect 7659 a level 1 scroll and
// a level 1 formula. Presence of Sight 2588 is applied only by a mage item.
const rawEnchantFixture = `
INSERT INTO SpellItemEnchantment (ID, Name_lang, Flags, RequiredSkillID, RequiredSkillRank, Effect, EffectPointsMin, EffectArg) VALUES
	(2717, 'Attack Power +26', 0, 0, 0, '[5,0,0]', '[26,0,0]', '[38,0,0]'),
	(7884, 'Attack Power +26', 0, 0, 0, '[5,0,0]', '[26,0,0]', '[38,0,0]'),
	(8203, 'Spirit +$k1', 0, 0, 0, '[5,0,0]', '[3,0,0]', '[6,0,0]'),
	(1523, 'MANA/FR +85/14', 0, 0, 0, '[3,0,0]', '[0,0,0]', '[15408,0,0]');
INSERT INTO Spell (ID, NameSubtext_lang) VALUES (29467, ''), (1219512, ''), (1248001, ''), (1248002, ''), (15427, '');
INSERT INTO SpellName (ID, Name_lang) VALUES
	(29467, 'Might of the Scourge'), (1219512, 'Might of the Scourge'),
	(1248001, 'Enchant Bracer - Lesser Spirit'), (1248002, 'Enchant Boots - Lesser Spirit'),
	(15427, 'Greater Arcane Amalgamation');
INSERT INTO SpellEffect (SpellID, Effect, EffectMiscValue_0) VALUES
	(29467, 53, 2717), (1219512, 53, 7884), (1248001, 53, 8203), (1248002, 53, 8203), (15427, 53, 1523);
INSERT INTO SkillLineAbility (Spell, ClassMask, SkillLine) VALUES (29467, 0, 0);
INSERT INTO ItemEffect (ID, SpellID) VALUES (1, 1219512), (2, 1248001), (3, 1248001), (4, 1248002), (5, 1248002), (6, 15427);
INSERT INTO ItemXItemEffect (ItemID, ItemEffectID) VALUES (236326, 1), (273601, 2), (249500, 3), (249501, 4), (249499, 5), (11663, 6);
INSERT INTO Item (ID, IconFileDataID) VALUES (236326, 10), (273601, 463531), (249500, 134327), (249501, 134327), (249499, 134327), (11663, 134094);
INSERT INTO ItemSparse (ID, Display_lang, OverallQualityID) VALUES (273601, 'Enchant Bracer - Lesser Spirit', 1), (249500, 'Formula', 2), (249501, 'Formula', 2),
	(11663, '[PH] Greater Arcane Amalgamation (MANA/FR)', 3);
INSERT INTO SpellItemEnchantment (ID, Name_lang, Flags, RequiredSkillID, RequiredSkillRank, Effect, EffectPointsMin, EffectArg) VALUES
	(7940, 'Grand Crusader', 0, 0, 0, '[1,0,0]', '[0,0,0]', '[1231124,0,0]'),
	(247, 'Agility +1', 0, 0, 0, '[5,0,0]', '[1,0,0]', '[3,0,0]'),
	(7659, 'Intellect +12', 0, 0, 0, '[5,0,0]', '[12,0,0]', '[5,0,0]');
INSERT INTO Spell (ID, NameSubtext_lang) VALUES (1231128, ''), (7779, ''), (1219577, '');
INSERT INTO SpellName (ID, Name_lang) VALUES
	(1231128, 'Enchant Weapon - Grand Crusader'), (7779, 'Enchant Bracer - Minor Agility'), (1219577, 'Enchant Off-Hand - Superior Intellect');
INSERT INTO SpellEffect (SpellID, Effect, EffectMiscValue_0) VALUES (1231128, 53, 7940), (7779, 53, 247), (1219577, 53, 7659);
INSERT INTO SkillLineAbility (Spell, ClassMask, SkillLine) VALUES (1231128, 0, 333), (7779, 0, 333), (1219577, 0, 333);
INSERT INTO ItemEffect (ID, SpellID) VALUES (7, 1231128), (8, 1231128), (9, 7779), (10, 1219577), (11, 1219577);
INSERT INTO ItemXItemEffect (ItemID, ItemEffectID) VALUES (241191, 7), (273607, 8), (273470, 9), (273598, 10), (274391, 11);
INSERT INTO Item (ID, IconFileDataID) VALUES (241191, 134327), (273607, 463531), (273470, 463531), (273598, 463531), (274391, 134327);
INSERT INTO ItemSparse (ID, Display_lang, OverallQualityID, ItemLevel) VALUES
	(241191, 'Formula', 3, 80), (273607, 'Enchant Weapon - Grand Crusader', 1, 1), (273470, 'Enchant Bracer - Minor Agility', 1, 1),
	(273598, 'Enchant Off-Hand - Superior Intellect', 1, 1), (274391, 'Formula', 1, 1);
INSERT INTO SpellItemEnchantment (ID, Name_lang, Flags, RequiredSkillID, RequiredSkillRank, Effect, EffectPointsMin, EffectArg) VALUES
	(2588, 'Presence of Sight', 0, 0, 0, '[5,0,0]', '[18,0,0]', '[45,0,0]');
INSERT INTO Spell (ID, NameSubtext_lang) VALUES (24156, '');
INSERT INTO SpellName (ID, Name_lang) VALUES (24156, 'Presence of Sight');
INSERT INTO SpellEffect (SpellID, Effect, EffectMiscValue_0) VALUES (24156, 53, 2588);
INSERT INTO ItemEffect (ID, SpellID) VALUES (12, 24156);
INSERT INTO ItemXItemEffect (ItemID, ItemEffectID) VALUES (19787, 12);
INSERT INTO Item (ID, IconFileDataID) VALUES (19787, 136153);
INSERT INTO ItemSparse (ID, Display_lang, OverallQualityID, ItemLevel, AllowableClass) VALUES (19787, 'Presence of Sight', 3, 60, 128);
`

func loadRawEnchantFixture(t *testing.T) []dbc.Enchant {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "client.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(rawEnchantSchema + rawEnchantFixture); err != nil {
		t.Fatalf("building the fixture: %v", err)
	}
	enchants, err := LoadAndWriteRawEnchants(&DBHelper{db: db}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return enchants
}

type rawEnchantKey struct {
	id   int
	name string
}

func rawEnchantsByKey(enchants []dbc.Enchant) map[rawEnchantKey]dbc.Enchant {
	byKey := map[rawEnchantKey]dbc.Enchant{}
	for _, enchant := range enchants {
		byKey[rawEnchantKey{enchant.EffectId, enchant.Name}] = enchant
	}
	return byKey
}

// One row per enchant and recipe name: enchants sharing a name stay apart, and an enchant granted
// under two names keeps both.
func TestRawEnchantsKeepOneRowPerEnchantAndName(t *testing.T) {
	enchants := loadRawEnchantFixture(t)
	byKey := rawEnchantsByKey(enchants)

	want := []rawEnchantKey{
		{2717, "Might of the Scourge"},
		{7884, "Might of the Scourge"},
		{8203, "Enchant Bracer - Lesser Spirit"},
		{8203, "Enchant Boots - Lesser Spirit"},
		{1523, "Greater Arcane Amalgamation"},
		{7940, "Enchant Weapon - Grand Crusader"},
		{247, "Enchant Bracer - Minor Agility"},
		{7659, "Enchant Off-Hand - Superior Intellect"},
		{2588, "Presence of Sight"},
	}
	for _, key := range want {
		if _, ok := byKey[key]; !ok {
			t.Errorf("no row for enchant %d %q", key.id, key.name)
		}
	}
	if len(enchants) != len(want) {
		t.Errorf("%d rows, want %d: %+v", len(enchants), len(want), enchants)
	}
}

// A grant taught by several items names the lowest item id the game files ship, or the lowest item id
// where it ships none, a Season of Discovery item last, and the icon and quality are that item's.
func TestRawEnchantNamesItsLowestItem(t *testing.T) {
	byKey := rawEnchantsByKey(loadRawEnchantFixture(t))

	for key, want := range map[rawEnchantKey]struct{ item, icon, quality int }{
		{8203, "Enchant Bracer - Lesser Spirit"}:  {249500, 134327, 2},
		{8203, "Enchant Boots - Lesser Spirit"}:   {249501, 134327, 2},
		{7884, "Might of the Scourge"}:            {236326, 10, 1},
		{7940, "Enchant Weapon - Grand Crusader"}: {273607, 463531, 1},
	} {
		got := byKey[key]
		if got.ItemId != want.item || got.FDID != want.icon || int(got.Quality) != want.quality {
			t.Errorf("enchant %d %q: item %d, icon %d, quality %d; want item %d, icon %d, quality %d",
				key.id, key.name, got.ItemId, got.FDID, got.Quality, want.item, want.icon, want.quality)
		}
	}
}

// A grant is live when a profession teaches it or any item using it has an ItemSparse row that does
// not name a placeholder ([PH]) item.
func TestRawEnchantIsLiveWhenTaughtOrOnAShippedItem(t *testing.T) {
	byKey := rawEnchantsByKey(loadRawEnchantFixture(t))

	for key, want := range map[rawEnchantKey]bool{
		{2717, "Might of the Scourge"}:           true,
		{7884, "Might of the Scourge"}:           false,
		{8203, "Enchant Bracer - Lesser Spirit"}: true,
		{8203, "Enchant Boots - Lesser Spirit"}:  true,
		{1523, "Greater Arcane Amalgamation"}:    false,
	} {
		if got := byKey[key].IsLive; got != want {
			t.Errorf("enchant %d %q reads live %v, want %v", key.id, key.name, got, want)
		}
	}
}

// The level is the highest item level above 1 of an item carrying the grant outside the Season of
// Discovery item range, else 0.
func TestRawEnchantLvl(t *testing.T) {
	byKey := rawEnchantsByKey(loadRawEnchantFixture(t))

	for key, want := range map[rawEnchantKey]int{
		{7940, "Enchant Weapon - Grand Crusader"}:       0,
		{7659, "Enchant Off-Hand - Superior Intellect"}: 0,
		{247, "Enchant Bracer - Minor Agility"}:         0,
	} {
		if got := byKey[key].Lvl; got != want {
			t.Errorf("enchant %d %q: lvl %d, want %d", key.id, key.name, got, want)
		}
	}
}

// An enchant no skill line teaches is open to the classes that may use an item applying it.
func TestRawEnchantClassAllowlist(t *testing.T) {
	byKey := rawEnchantsByKey(loadRawEnchantFixture(t))

	for key, want := range map[rawEnchantKey][]proto.Class{
		{2588, "Presence of Sight"}:               {proto.Class_ClassMage},
		{8203, "Enchant Bracer - Lesser Spirit"}:  nil,
		{7940, "Enchant Weapon - Grand Crusader"}: nil,
	} {
		if got := dbc.GetClassesFromClassMask(byKey[key].ClassMask); !slices.Equal(got, want) {
			t.Errorf("enchant %d %q: class allowlist %v, want %v", key.id, key.name, got, want)
		}
	}
}
