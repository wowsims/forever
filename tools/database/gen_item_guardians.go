package database

import (
	"fmt"
	"go/format"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/database/overrides"
)

const itemGuardiansFileName = "sim/common/forever/item_guardians_auto_gen.go"

// An item whose use, equip or chance-on-hit spell summons a creature: the spell carries the E_SUMMON, or a spell it
// triggers does.
type itemSummon struct {
	itemID   int32
	name     string
	onUse    bool  // The summon comes from the item's use.
	spellID  int32 // The item effect's spell.
	summonID int32 // The spell whose E_SUMMON names the creature.
	creature int32
	count    int
	byHand   bool // The item's effect is written by hand, as Dragon's Call's is.
}

// Every summon an item in db reaches through its use, equip and chance-on-hit spells, in item id order. Totems and
// overrides.NotGuardianSummons stay out.
func itemSummons(instance *dbc.DBC, db *WowDatabase, byHand func(int32) bool) []itemSummon {
	var summons []itemSummon
	for _, id := range slices.Sorted(maps.Keys(db.Items)) {
		seen := map[int32]bool{}
		for _, effect := range instance.ItemEffectsByParentID[int(id)] {
			if effect.SpellID == 0 || effect.TriggerType != dbc.ITEM_SPELLTRIGGER_ON_USE &&
				effect.TriggerType != dbc.ITEM_SPELLTRIGGER_ON_EQUIP && effect.TriggerType != dbc.ITEM_SPELLTRIGGER_CHANCE_ON_HIT {
				continue
			}
			summonID, summon, ok := summonOf(instance, effect.SpellID)
			if !ok || seen[summonID] {
				continue
			}
			seen[summonID] = true
			if _, excluded := overrides.NotGuardianSummons[summonID]; excluded || summonsIntoTotemSlot(summon) {
				continue
			}
			summons = append(summons, itemSummon{
				itemID:   id,
				name:     db.Items[id].Name,
				onUse:    effect.TriggerType == dbc.ITEM_SPELLTRIGGER_ON_USE,
				spellID:  int32(effect.SpellID),
				summonID: summonID,
				creature: int32(summon.EffectMiscValues[0]),
				count:    max(1, int(summon.EffectBasePoints)),
				byHand:   byHand(id),
			})
		}
	}
	return summons
}

// The spell's E_SUMMON, or that of a spell it triggers, with the spell that carries it.
func summonOf(instance *dbc.DBC, spellID int) (int32, dbc.SpellEffect, bool) {
	for _, id := range append([]int{spellID}, spellTriggers(instance, spellID)...) {
		for _, effect := range instance.SpellEffects[id] {
			if effect.EffectType == dbcenums.E_SUMMON && len(effect.EffectMiscValues) > 1 {
				return int32(id), effect, true
			}
		}
	}
	return 0, dbc.SpellEffect{}, false
}

// The spells the spell's effects trigger, in effect order.
func spellTriggers(instance *dbc.DBC, spellID int) []int {
	var ids []int
	for _, index := range slices.Sorted(maps.Keys(instance.SpellEffects[spellID])) {
		if trigger := instance.SpellEffects[spellID][index].EffectTriggerSpell; trigger > 0 {
			ids = append(ids, trigger)
		}
	}
	return ids
}

func summonsIntoTotemSlot(summon dbc.SpellEffect) bool {
	return slices.Contains(totemSlotSummonProperties, int32(summon.EffectMiscValues[1]))
}

// Writes item_guardians_auto_gen.go: a NewGuardianOnUse call for every use item whose spell summons a guardian, and
// as comments the summons it does not register, those on a proc and those written by hand. Every summon whose
// creature forever.GuardianCreatures does not hold, and that no hand-written effect covers, goes on the missing
// effects list, so the gear tab says its creature is not simulated.
func GenerateItemGuardians(instance *dbc.DBC, db *WowDatabase) error {
	summons := itemSummons(instance, db, implementedByHand)
	for _, s := range summons {
		if _, measured := forever.GuardianCreatures[s.creature]; s.byHand || s.onUse && measured {
			continue
		}
		tooltip, rendered := renderItemEffectTooltip(db.Items[s.itemID], &proto.ItemEffect{BuffId: s.spellID}, instance)
		if !rendered {
			tooltip = instance.Spells[int(s.spellID)].NameLang
		}
		StoreMissingEffect("ItemEffects", s.name, Variant{ID: int(s.itemID), Name: tooltip, SpellID: int(s.spellID)})
	}
	source, err := format.Source([]byte(renderItemGuardians(summons, func(id int32) bool { return spelldata.Find(id) != spelldata.Nil })))
	if err != nil {
		return fmt.Errorf("failed to format %s: %w", itemGuardiansFileName, err)
	}
	return os.WriteFile(itemGuardiansFileName, source, 0o644)
}

// The generated file's text. A use whose summon the spell store does not carry is written as a comment, since
// NewGuardianOnUse reads the row.
func renderItemGuardians(summons []itemSummon, inStore func(int32) bool) string {
	var uses, others strings.Builder
	for _, s := range summons {
		what := fmt.Sprintf("%s: %d summons %d × creature %d", s.name, s.summonID, s.count, s.creature)
		switch {
		case s.byHand:
			fmt.Fprintf(&others, "\t// %d %s, written by hand\n", s.itemID, what)
		case !s.onUse:
			fmt.Fprintf(&others, "\t// %d %s on a proc, which needs its trigger wired by hand\n", s.itemID, what)
		case !inStore(s.summonID):
			fmt.Fprintf(&uses, "\t\t// not in the spell store: {%d, %d}, // %s\n", s.itemID, s.summonID, what)
		default:
			fmt.Fprintf(&uses, "\t\t{%d, %d}, // %s\n", s.itemID, s.summonID, what)
		}
	}

	var out strings.Builder
	out.WriteString("// Code generated by tools/database/gen_item_guardians.go. DO NOT EDIT.\n\npackage forever\n\n")
	out.WriteString("import \"github.com/wowsims/forever/sim/common/shared\"\n\n")
	out.WriteString("// Every use item whose spell summons a guardian. One whose creature GuardianCreatures does not hold registers\n")
	out.WriteString("// nothing and keeps its stats.\n")
	out.WriteString("func RegisterAllItemGuardians() {\n")
	out.WriteString("\tfor _, use := range []struct{ itemID, summonID int32 }{\n")
	out.WriteString(uses.String())
	out.WriteString("\t} {\n\t\tshared.NewGuardianOnUse(use.itemID, use.summonID, GuardianCreatures)\n\t}\n")
	if others.Len() > 0 {
		out.WriteString("\n\t// Not registered here:\n")
		out.WriteString(others.String())
	}
	out.WriteString("}\n")
	return out.String()
}
