package database

import (
	"fmt"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/tools/database/dbc"
)

// An exclusive group the rows do not state: a row group the items join, or one of their own.
type ConsumableGroup struct {
	Group ConsumableGroupKey
	Items []int32
}

type ConsumableGroupKey struct {
	Key  string
	Name string
}

// The group of the elixirs and scrolls whose on-use spell's first aura is the aura type with the misc
// value.
func ElixirGroup(aura dbcenums.EffectAuraType, misc int) ConsumableGroupKey {
	return ConsumableGroupKey{Key: fmt.Sprintf("%d/%d", aura, misc), Name: auraGroupName(aura, misc)}
}

// Writes the consumable's exclusive group. ConsumableExclusiveGroups names it first. Otherwise an
// elixir's or a scroll's is the aura type and misc value of the first aura its on-use spell applies,
// so a scroll joins the elixirs of the same key, and a scroll whose key no elixir has shares a group
// with its ranks alone. A stat consumable used one at a time with the others of its category
// (shared.OneAtATimeCategory) is grouped by that category. Any other consumable has no group.
func setConsumableGroup(c *dbc.Consumable, consumable *proto.Consumable, routing *ConsumableRouting, instance *dbc.DBC) {
	group, ok := exclusiveGroupOverride(consumable.Id)
	if !ok {
		group = rowGroupOf(c, consumable, routing, instance)
	}
	consumable.GroupKey, consumable.GroupName = group.Key, group.Name
}

func rowGroupOf(c *dbc.Consumable, consumable *proto.Consumable, routing *ConsumableRouting, instance *dbc.DBC) ConsumableGroupKey {
	if len(consumable.ItemEffects) == 0 {
		return ConsumableGroupKey{}
	}
	onUse := consumable.ItemEffects[0]
	switch consumable.Type {
	case proto.ConsumableType_ConsumableTypeElixir, proto.ConsumableType_ConsumableTypeScroll:
		aura, misc, ok := firstAura(instance, int(onUse.BuffId))
		if !ok {
			return ConsumableGroupKey{}
		}
		return ElixirGroup(aura, misc)
	}
	if registrationFamilyOf(consumable, c, routing) == statFamily {
		if category := shared.OneAtATimeCategory(onUse); category > 0 {
			return ConsumableGroupKey{Key: fmt.Sprintf("category %d", category), Name: fmt.Sprintf("Category %d", category)}
		}
	}
	return ConsumableGroupKey{}
}

func exclusiveGroupOverride(id int32) (ConsumableGroupKey, bool) {
	for _, group := range ConsumableExclusiveGroups {
		if slices.Contains(group.Items, id) {
			return group.Group, true
		}
	}
	return ConsumableGroupKey{}, false
}

func firstAura(instance *dbc.DBC, spellID int) (dbcenums.EffectAuraType, int, bool) {
	for _, e := range instance.SpellEffectsInOrder(spellID) {
		if e.EffectType == dbcenums.E_APPLY_AURA {
			misc := 0
			if len(e.EffectMiscValues) > 0 {
				misc = e.EffectMiscValues[0]
			}
			return e.EffectAura, misc, true
		}
	}
	return 0, 0, false
}

const allSpellSchools = 126

func auraGroupName(aura dbcenums.EffectAuraType, misc int) string {
	switch aura {
	case dbcenums.A_MOD_STAT:
		if stat, ok := dbc.MapMainStatToStat(misc); ok {
			return stats.Stat(stat).StatName()
		}
		return "All Stats"
	case dbcenums.A_MOD_DAMAGE_DONE:
		switch misc {
		case allSpellSchools:
			return "Spell Power"
		case int(core.SpellSchoolPhysical):
			return "Physical Damage"
		}
		return schoolLabel(misc) + " Spell Power"
	case dbcenums.A_MOD_RESISTANCE, dbcenums.A_MOD_ARMOR_NO_MODS:
		if misc == int(core.SpellSchoolPhysical) {
			return "Armor"
		}
		return schoolLabel(misc) + " Resistance"
	case dbcenums.A_MOD_INCREASE_HEALTH:
		return "Health"
	case dbcenums.A_MOD_POWER_REGEN, dbcenums.A_PERIODIC_ENERGIZE:
		return "Mana Regeneration"
	case dbcenums.A_MOD_HEALTH_REGEN_IN_COMBAT:
		return "Health Regeneration"
	case dbcenums.A_MOD_ATTACK_POWER:
		return "Attack Power"
	case dbcenums.A_MOD_HEALING_DONE:
		return "Healing"
	case dbcenums.A_MOD_CRIT_PCT:
		return "Critical Strike"
	}
	return "Other"
}

func schoolLabel(mask int) string {
	name := core.SpellSchool(mask).String()
	if strings.HasPrefix(name, "SpellSchool(") {
		return "School"
	}
	return strings.TrimPrefix(name, "SpellSchool")
}
