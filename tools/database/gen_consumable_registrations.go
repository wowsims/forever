package database

import (
	"fmt"
	"go/format"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/tools/database/dbc"
)

type consumableRegistration struct {
	id   int32
	name string
	// The imbue's equip spells, which the builder takes after the consumable's own id.
	args []int32
	// The spell penetration the imbue states on one school as a stat.
	piercing shared.SchoolPiercing
	// Why the builder refuses the consumable; its table entry is then written commented out beneath it.
	unsupported []string
}

type registrationGroup struct {
	title   string
	entries []consumableRegistration
}

// Writes the registration of every consumable a builder in sim/common/shared runs from its rows, one
// file per family, and each one a builder refuses as a commented-out table entry beneath the reason. The ids
// are spelled out rather than read from the loaded database, which the browser's sim does not have
// when the effects register. routes holds the routings gen_db already made, by item id.
func GenerateConsumableRegistrations(instance *dbc.DBC, db *WowDatabase, routes map[int]*ConsumableRouting) error {
	var explosives, refusedExplosives, potions, conjured, others, stats, imbues []consumableRegistration
	ids := map[int32]bool{}
	for id := range instance.Consumables {
		ids[int32(id)] = true
	}
	for id := range db.Consumables {
		ids[id] = true
	}

	for _, id := range slices.Sorted(maps.Keys(ids)) {
		c, shipped := db.Consumables[id]
		catalogued, inCatalogue := instance.Consumables[int(id)]
		routing := routes[int(id)]
		if routing == nil && inCatalogue {
			routing = RouteConsumable(&catalogued, instance)
		}
		name := catalogued.Name
		if shipped {
			name = c.Name
		}
		entry := consumableRegistration{id: id, name: name}

		var catalogueEntry *dbc.Consumable
		if inCatalogue {
			catalogueEntry = &catalogued
		}
		switch registrationFamilyOf(c, catalogueEntry, routing) {
		case imbueFamily:
			if inCatalogue {
				entry.args = parsedEquipSpells(catalogued.TemporaryEnchant)
				entry.piercing = enchantPiercing(catalogued.TemporaryEnchant)
			}
			imbues = append(imbues, entry)
		case explosiveFamily:
			if routing.Supported() {
				explosives = append(explosives, entry)
				continue
			}
			entry.unsupported = routing.Unsupported
			refusedExplosives = append(refusedExplosives, entry)
		case potionFamily:
			entry.unsupported = onUseUnsupported(c)
			potions = append(potions, entry)
		case conjuredFamily:
			entry.unsupported = onUseUnsupported(c)
			conjured = append(conjured, entry)
		case otherOnUseFamily:
			entry.unsupported = onUseUnsupported(c)
			others = append(others, entry)
		case statFamily:
			stats = append(stats, entry)
		}
	}

	for _, file := range []struct {
		path, function, constructor string
		equipSpells                 bool
		groups                      []registrationGroup
	}{
		{"sim/common/forever/consumables_explosives_auto_gen.go", "RegisterAllExplosives", "NewSpellDataExplosive", false,
			[]registrationGroup{{entries: explosives}, {entries: refusedExplosives}}},
		{"sim/common/forever/consumables_potions_auto_gen.go", "RegisterPotionsAndConjured", "NewSpellDataOnUseConsumable", false,
			[]registrationGroup{{"Potions", potions}, {"Conjured", conjured}, {"Other on-use consumables", others}}},
		{"sim/common/forever/consumables_imbues_auto_gen.go", "RegisterAllTemporaryEnchantConsumables", "NewSpellDataTemporaryEnchant", true,
			[]registrationGroup{{entries: imbues}}},
		{"sim/common/forever/consumable_stats_auto_gen.go", "RegisterAllStatConsumables", "NewSpellDataStatConsumable", false,
			[]registrationGroup{{entries: stats}}},
	} {
		if err := writeRegistrationFile(file.path, file.function, file.constructor, file.equipSpells, file.groups); err != nil {
			return err
		}
	}
	return nil
}

type registrationFamily int

const (
	unregisteredFamily registrationFamily = iota
	imbueFamily
	explosiveFamily
	potionFamily
	conjuredFamily
	otherOnUseFamily
	statFamily
)

// The family GenerateConsumableRegistrations registers the consumable in, which names its builder. c
// is nil for a consumable db.json does not ship, catalogued for one the catalogue does not carry. An
// imbue ClassBuiltImbues names is the class's own, and registers in none.
func registrationFamilyOf(c *proto.Consumable, catalogued *dbc.Consumable, routing *ConsumableRouting) registrationFamily {
	_, classBuilt := ClassBuiltImbues[c.GetId()]
	switch {
	case c != nil && classBuilt:
		return unregisteredFamily
	case c != nil && (c.Type == proto.ConsumableType_ConsumableTypeImbue || c.Type == proto.ConsumableType_ConsumableTypeWeaponImbue):
		return imbueFamily
	case catalogued != nil && IsExplosive(catalogued, routing):
		return explosiveFamily
	case c == nil:
		return unregisteredFamily
	case c.CategoryId == dbc.POTION_SPELL_CATEGORY:
		return potionFamily
	case c.Type == proto.ConsumableType_ConsumableTypeConjured:
		return conjuredFamily
	case catalogued == nil:
		return unregisteredFamily
	case routing.Kind == ConsumableHeal, routing.Kind == ConsumableAbsorb, routing.Kind == ConsumableEnergize:
		return otherOnUseFamily
	case routing.Supported() && routing.Kind == ConsumableStatAura:
		return statFamily
	}
	return unregisteredFamily
}

// The equip spells of a temporary enchant that the imbue builder parses: those that add no stat,
// resistance or weapon damage to the enchant's own fields, whose every effect lands on the user, and
// whose every aura the builder's parse, counted once per hand, attaches. Consecrated Sharpening
// Stone's enchant 2684 names 28893, its attack power against Undead; a wizard oil's equip spell is
// already the enchant's spell damage. An equip spell that procs is taken where the builder can
// build its listener: Scroll of Imbue Manablade's enchant 8715 names 1308913, mana on melee hits.
func parsedEquipSpells(enchant dbc.TemporaryEnchant) []int32 {
	var spellIDs []int32
	for slot, effect := range enchant.Effects {
		spellID := int32(enchant.Args[slot])
		if effect != dbc.ITEM_ENCHANTMENT_EQUIP_SPELL || spellID == 0 {
			continue
		}
		alone := dbc.TemporaryEnchant{ID: enchant.ID}
		alone.Effects[0], alone.Points[0], alone.Args[0] = effect, enchant.Points[slot], enchant.Args[slot]
		fields := alone.ToProto()
		row := spelldata.Find(spellID)
		if slices.ContainsFunc(fields.Stats, isNonZero) || slices.ContainsFunc(fields.PseudoStats, isNonZero) ||
			fields.WeaponDamage != 0 || len(fields.EnchantEffects) > 0 || row == spelldata.Nil {
			continue
		}
		if shared.TemporaryEnchantEquipSpellProcs(row) {
			if len(spelldata.EnchantAuraUnsupported(row)) > 0 || len(shared.FiredSpellUnsupported(row)) > 0 {
				continue
			}
		} else if slices.ContainsFunc(row.Effects, landsOffTheUser) || len(spelldata.DryRun(row, spelldata.Count(2)).Skipped) > 0 {
			continue
		}
		spellIDs = append(spellIDs, spellID)
	}
	return spellIDs
}

// The stat slot of a temporary enchant that pierces one school's resistance, which the stat array
// the enchant's other stats ship in has no place for: Scroll of Imbue Baleflame's enchant 8706.
func enchantPiercing(enchant dbc.TemporaryEnchant) shared.SchoolPiercing {
	for slot, effect := range enchant.Effects {
		if effect != dbc.ITEM_ENCHANTMENT_STAT {
			continue
		}
		if school, ok := dbc.MapBonusStatIndexToPiercingSchool(enchant.Args[slot]); ok {
			return shared.SchoolPiercing{School: school, Amount: float64(enchant.Points[slot])}
		}
	}
	return shared.SchoolPiercing{}
}

func landsOffTheUser(e spelldata.Effect) bool {
	return !shared.LandsOnTheUser(&e)
}

func isNonZero(v float64) bool {
	return v != 0
}

// Why the potion builder cannot run the spell the consumable's first on-use row casts.
func onUseUnsupported(c *proto.Consumable) []string {
	for _, effect := range c.ItemEffects {
		if effect.GetOnUse() != nil {
			return shared.ConsumableUnsupported(spelldata.Find(effect.BuffId))
		}
	}
	return nil
}

// The name a generated file reaches a school index by.
var schoolIndexNames = map[stats.SchoolIndex]string{
	stats.SchoolIndexFire:  "stats.SchoolIndexFire",
	stats.SchoolIndexFrost: "stats.SchoolIndexFrost",
}

// Each family is one table and one loop that calls its builder: the compiler inlines the builder, and
// the closure it hands core, into every call site, so a call per consumable copies the builder once
// per consumable.
func writeRegistrationFile(path, function, constructor string, equipSpells bool, groups []registrationGroup) error {
	var table strings.Builder
	usesStats := false
	for i, group := range groups {
		switch {
		case group.title != "" && i > 0:
			fmt.Fprintf(&table, "\n\t\t// %s\n", group.title)
		case group.title != "":
			fmt.Fprintf(&table, "\t\t// %s\n", group.title)
		case i > 0 && len(group.entries) > 0:
			table.WriteString("\n")
		}
		for _, e := range group.entries {
			entry := fmt.Sprintf("%d,", e.id)
			if equipSpells {
				fields := []string{fmt.Sprintf("itemID: %d", e.id)}
				if e.piercing.Amount != 0 {
					fields = append(fields, fmt.Sprintf("piercing: shared.SchoolPiercing{School: %s, Amount: %s}",
						schoolIndexNames[e.piercing.School], num(e.piercing.Amount)))
					usesStats = usesStats || len(e.unsupported) == 0
				}
				if len(e.args) > 0 {
					args := make([]string, len(e.args))
					for j, arg := range e.args {
						args[j] = fmt.Sprint(arg)
					}
					fields = append(fields, fmt.Sprintf("equipSpellIDs: []int32{%s}", strings.Join(args, ", ")))
				}
				entry = "{" + strings.Join(fields, ", ") + "},"
			}
			if len(e.unsupported) > 0 {
				fmt.Fprintf(&table, "\t\t// unsupported: %s\n\t\t// %s // %s\n", strings.Join(e.unsupported, "; "), entry, e.name)
				continue
			}
			fmt.Fprintf(&table, "\t\t%s // %s\n", entry, e.name)
		}
	}

	var out strings.Builder
	out.WriteString("// Code generated by tools/database/gen_consumable_registrations.go. DO NOT EDIT.\n\npackage forever\n\n")
	if usesStats {
		out.WriteString("import (\n\t\"github.com/wowsims/forever/sim/common/shared\"\n\t\"github.com/wowsims/forever/sim/core/stats\"\n)\n\n")
	} else {
		out.WriteString("import \"github.com/wowsims/forever/sim/common/shared\"\n\n")
	}
	fmt.Fprintf(&out, "func %s() {\n", function)
	if equipSpells {
		fmt.Fprintf(&out, "\tfor _, imbue := range []struct {\n\t\titemID int32\n\t\tpiercing shared.SchoolPiercing\n\t\tequipSpellIDs []int32\n\t}{\n%s\t} {\n", table.String())
		fmt.Fprintf(&out, "\t\tshared.%s(imbue.itemID, imbue.piercing, imbue.equipSpellIDs...)\n\t}\n}\n", constructor)
	} else {
		fmt.Fprintf(&out, "\tfor _, itemID := range []int32{\n%s\t} {\n", table.String())
		fmt.Fprintf(&out, "\t\tshared.%s(itemID)\n\t}\n}\n", constructor)
	}

	formatted, err := format.Source([]byte(out.String()))
	if err != nil {
		return fmt.Errorf("generated %s is not valid Go: %w", path, err)
	}
	return os.WriteFile(path, formatted, 0644)
}
