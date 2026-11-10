package siml

import (
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/wowsims/forever/sim/core/itemfit"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// The item database as the editor reads it (assets/database/db.json, then leftover_db.json): names and icons, which
// slots an item goes in, which enchants fit it and which random suffixes it rolls, and the kind of each consumable.

type CatalogConsumable struct {
	ID    int32
	Name  string
	Type  proto.ConsumableType
	OnUse bool
}

type CatalogSuffix struct {
	ID    int32
	Name  string
	Stats []float64 // Indexed by Stat.
}

// The database files, from the repository root, in the order they are read: the first holds an id both hold.
var catalogFiles = []string{filepath.Join("assets", "database", "db.json"), filepath.Join("assets", "database", "leftover_db.json")}

// The catalog read from the database files.
type DatabaseCatalog struct {
	items       map[int32]CatalogItem
	itemList    []CatalogItem
	consumables []CatalogConsumable
	names       map[int32]string // Item names: gear, then consumables, then item icons.
	spells      map[int32]string
	spellIcons  map[int32][2]string // [name, icon], for the rotation panel.
	itemIcons   map[int32][2]string // [name, icon]: a gear item's own over its item icon's.
	rows        map[int32][]CatalogEnchant
	suffixes    map[int32]CatalogSuffix
	itemSuffix  map[int32][]int32
}

// Reads the database files of the repository at root. A file that is missing or does not read is left out.
func ReadCatalog(root string) *DatabaseCatalog {
	c := &DatabaseCatalog{
		items: map[int32]CatalogItem{}, names: map[int32]string{}, spells: map[int32]string{},
		rows: map[int32][]CatalogEnchant{}, suffixes: map[int32]CatalogSuffix{}, itemSuffix: map[int32][]int32{},
		spellIcons: map[int32][2]string{}, itemIcons: map[int32][2]string{},
	}
	consumables := map[int32]bool{}
	var iconNames []*proto.IconData
	for _, file := range catalogFiles {
		text, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			continue
		}
		db := &proto.UIDatabase{}
		if (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(text, db) != nil {
			continue
		}
		for _, item := range db.Items {
			if _, ok := c.items[item.Id]; ok {
				continue
			}
			ci := CatalogItem{
				ID: item.Id, Name: item.Name, Icon: item.Icon, Type: item.Type, RangedType: item.RangedWeaponType,
				HandType: item.HandType, WeaponType: item.WeaponType, Quality: item.Quality, Ilvl: item.ScalingOptions[0].GetIlvl(),
				Phase: item.Phase, SetName: item.SetName, OnUse: onUse(item.ItemEffects),
			}
			for _, e := range item.ItemEffects {
				if e.BuffName != "" {
					ci.Effects = append(ci.Effects, e.BuffName)
				}
			}
			c.items[item.Id] = ci
			c.itemList = append(c.itemList, ci)
			c.itemSuffix[item.Id] = item.RandomSuffixOptions
		}
		for _, x := range db.Consumables {
			if !consumables[x.Id] {
				consumables[x.Id] = true
				c.consumables = append(c.consumables, CatalogConsumable{ID: x.Id, Name: x.Name, Type: x.Type, OnUse: onUse(x.ItemEffects)})
			}
		}
		iconNames = append(iconNames, db.ItemIcons...)
		for _, icon := range db.SpellIcons {
			if _, ok := c.spells[icon.Id]; !ok && icon.Name != "" {
				c.spells[icon.Id], c.spellIcons[icon.Id] = icon.Name, [2]string{icon.Name, icon.Icon}
			}
		}
		for _, e := range db.Enchants {
			row := CatalogEnchant{
				EffectID: e.EffectId, Name: e.Name, Type: e.Type, EnchantType: e.EnchantType, ExtraTypes: e.ExtraTypes,
				SpellID: e.SpellId, ItemID: e.ItemId,
			}
			c.rows[e.EffectId] = append(c.rows[e.EffectId], row)
		}
		for _, s := range db.RandomSuffixes {
			if _, ok := c.suffixes[s.Id]; !ok {
				c.suffixes[s.Id] = CatalogSuffix{ID: s.Id, Name: s.Name, Stats: s.Stats}
			}
		}
	}
	for _, item := range c.itemList {
		if item.Name != "" {
			c.names[item.ID] = item.Name
		}
	}
	for _, x := range c.consumables {
		if _, ok := c.names[x.ID]; !ok && x.Name != "" {
			c.names[x.ID] = x.Name
		}
	}
	for _, icon := range iconNames {
		if _, ok := c.names[icon.Id]; !ok && icon.Name != "" {
			c.names[icon.Id] = icon.Name
		}
		if _, ok := c.itemIcons[icon.Id]; !ok && icon.Name != "" {
			c.itemIcons[icon.Id] = [2]string{icon.Name, icon.Icon}
		}
	}
	for _, item := range c.itemList {
		if item.Name != "" {
			c.itemIcons[item.ID] = [2]string{item.Name, item.Icon}
		}
	}
	return c
}

func onUse(effects []*proto.ItemEffect) bool {
	return slices.ContainsFunc(effects, func(e *proto.ItemEffect) bool { return e.GetOnUse() != nil })
}

func (c *DatabaseCatalog) Item(id int32) (CatalogItem, bool) {
	item, ok := c.items[id]
	return item, ok
}

func (c *DatabaseCatalog) Items() []CatalogItem { return c.itemList }

// Spell names and icons by id, [name, icon], from the database's spell icons.
func (c *DatabaseCatalog) SpellIcons() map[int32][2]string { return c.spellIcons }

// Item names and icons by id, [name, icon]: a gear item's own, else its item icon's.
func (c *DatabaseCatalog) ItemIcons() map[int32][2]string { return c.itemIcons }

func (c *DatabaseCatalog) Suffixes(itemID int32) []int32 { return c.itemSuffix[itemID] }

func (c *DatabaseCatalog) SpellName(spellID int32) (string, bool) {
	name, ok := c.spells[spellID]
	return name, ok
}

func (c *DatabaseCatalog) ItemName(id int32) (string, bool) {
	name, ok := c.names[id]
	return name, ok
}

func (c *DatabaseCatalog) Consumables() []CatalogConsumable { return c.consumables }

func (c *DatabaseCatalog) Suffix(id int32) (CatalogSuffix, bool) {
	s, ok := c.suffixes[id]
	return s, ok
}

func (c *DatabaseCatalog) EnchantRows(effectID int32) []CatalogEnchant { return c.rows[effectID] }

// The catalog of the repository at root, read again whenever a database file changes on disk.
func CatalogAt(root string) *DatabaseCatalog {
	catalogs.Lock()
	held := catalogs.byRoot[root]
	if held == nil {
		held = &Reread[*DatabaseCatalog]{}
		catalogs.byRoot[root] = held
	}
	catalogs.Unlock()
	paths := make([]string, len(catalogFiles))
	for i, file := range catalogFiles {
		paths[i] = filepath.Join(root, file)
	}
	c, _ := held.Get(paths, func() (*DatabaseCatalog, error) { return ReadCatalog(root), nil })
	return c
}

var catalogs = struct {
	sync.Mutex
	byRoot map[string]*Reread[*DatabaseCatalog]
}{byRoot: map[string]*Reread[*DatabaseCatalog]{}}

// --- What goes together, by the rules the sim checks a gear set with (sim/core/itemfit).

func (item CatalogItem) fit() itemfit.Item {
	return itemfit.Item{Type: item.Type, HandType: item.HandType, WeaponType: item.WeaponType, RangedType: item.RangedType}
}

// The slots an item goes in.
func (item CatalogItem) slots() []proto.ItemSlot { return itemfit.ItemSlots(item.fit()) }

// Whether an enchant row fits an item.
func (row CatalogEnchant) fits(item CatalogItem) bool {
	return itemfit.EnchantFits(itemfit.Enchant{Type: row.Type, EnchantType: row.EnchantType, ExtraTypes: row.ExtraTypes}, item.fit())
}

// Whether a row of the enchant effect fits the item.
func enchantEffectFits(c *DatabaseCatalog, effectID int32, item CatalogItem) bool {
	return slices.ContainsFunc(c.EnchantRows(effectID), func(row CatalogEnchant) bool { return row.fits(item) })
}
