package database

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

func shippedConsumable(t *testing.T, instance *dbc.DBC, id int) *proto.Consumable {
	t.Helper()
	catalogued, ok := instance.Consumables[id]
	if !ok {
		t.Fatalf("%d is not in the catalogue", id)
	}
	routing := RouteConsumable(&catalogued, instance)
	if !ShipsConsumable(routing) {
		t.Fatalf("%d %s does not ship", id, catalogued.Name)
	}
	return ShippedConsumable(&catalogued, routing, instance)
}

// The ranks of a scroll share one group, keyed by their first aura, and a scroll of another stat is
// another group. An elixir has none, and neither have the Jujus nor Bogling Root: the sim counts each
// stat they grant at its strongest one's amount.
func TestAConsumableTakesItsExclusiveGroup(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	groupNames := map[string]string{}
	for _, tc := range []struct {
		name string
		ids  []int
	}{
		{"Strength: Scroll of Strength IV and Scroll of Strength", []int{10310, 954}},
		{"Agility: Scroll of Agility IV, Scroll of Agility II and Scroll of Agility", []int{10309, 1477, 3012}},
		{"Intellect: Scroll of Intellect IV and Scroll of Intellect", []int{10308, 955}},
		{"Armor: Scroll of Protection IV and Scroll of Protection", []int{10305, 3013}},
		{"Spirit: Scroll of Spirit IV and Scroll of Spirit", []int{10306, 1181}},
		{"Stamina: Scroll of Stamina IV and Scroll of Stamina", []int{10307, 1180}},
	} {
		first := shippedConsumable(t, instance, tc.ids[0])
		if first.GroupKey == "" || first.GroupName == "" {
			t.Errorf("%s: %d takes group %q %q, want one", tc.name, tc.ids[0], first.GroupKey, first.GroupName)
			continue
		}
		for _, id := range tc.ids[1:] {
			if got := shippedConsumable(t, instance, id); got.GroupKey != first.GroupKey || got.GroupName != first.GroupName {
				t.Errorf("%s: %d takes group %q %q, want %d's %q %q", tc.name, id, got.GroupKey, got.GroupName, tc.ids[0], first.GroupKey, first.GroupName)
			}
		}
		if other, taken := groupNames[first.GroupKey]; taken {
			t.Errorf("%s: group %q is also %s's", tc.name, first.GroupKey, other)
		}
		groupNames[first.GroupKey] = tc.name
	}

	for _, tc := range []struct {
		name       string
		id         int
		key, label string
	}{
		{"Spirit of Zanza", 20079, "Zanza", "Zanza"},
		{"R.O.I.D.S., grouped by category in its rows, Zanza by the override", 8410, "Zanza", "Zanza"},
		{"Rumsey Rum Black Label, a food, takes Rum and leaves the food pick", 21151, "Rum", "Rum"},
		{"Gordok Green Grog", 18269, "Rum", "Rum"},
		{"Kreeg's Stout Beatdown", 18284, "Rum", "Rum"},
		{"Fire-toasted Bun", 23327, "Hit food", "Hit food"},
		{"Dark Desire", 22237, "Hit food", "Hit food"},
		{"Elixir of Greater Strength", 9206, "", ""},
		{"Elixir of the Mongoose (Agility then crit)", 13452, "", ""},
		{"Arcane Elixir", 9155, "", ""},
		{"Gift of Arthas", 9088, "", ""},
		{"Juju Power", 12451, "", ""},
		{"Juju Might", 12460, "", ""},
		{"Bogling Root", 5206, "", ""},
		{"Flask of the Titans stacks with everything", 13510, "", ""},
		{"Mighty Rage Potion stacks with everything", 13442, "", ""},
	} {
		if got := shippedConsumable(t, instance, tc.id); got.GroupKey != tc.key || got.GroupName != tc.label {
			t.Errorf("%s: group %q %q, want %q %q", tc.name, got.GroupKey, got.GroupName, tc.key, tc.label)
		}
	}
}

// A scroll's group is keyed by its first aura's type and misc value and named after what that aura
// grants.
func TestScrollGroupNamesWhatTheAuraGrants(t *testing.T) {
	for _, tc := range []struct {
		aura      dbcenums.EffectAuraType
		misc      int
		key, name string
	}{
		{dbcenums.A_MOD_STAT, 1, "29/1", "Agility"},
		{dbcenums.A_MOD_STAT, 3, "29/3", "Intellect"},
		{dbcenums.A_MOD_DAMAGE_DONE, 126, "13/126", "Spell Power"},
		{dbcenums.A_MOD_DAMAGE_DONE, 4, "13/4", "Fire Spell Power"},
		{dbcenums.A_MOD_RESISTANCE, 32, "22/32", "Shadow Resistance"},
		{dbcenums.A_MOD_RESISTANCE, 1, "22/1", "Armor"},
		{dbcenums.A_MOD_ATTACK_POWER, 0, "99/0", "Attack Power"},
		{dbcenums.A_PERIODIC_ENERGIZE, 0, "24/0", "Mana Regeneration"},
	} {
		if got, want := ScrollGroup(tc.aura, tc.misc), (ConsumableGroupKey{Key: tc.key, Name: tc.name}); got != want {
			t.Errorf("aura %v misc %d: group %+v, want %+v", tc.aura, tc.misc, got, want)
		}
	}
}

// The category a stat consumable is used one at a time with the others of, read off its rows. The
// item's on-use states the category and its cooldown, or its spell does where the item states none
// (a category of 0 or less, a cooldown below 0). Each buff the consumable puts on the user has to
// last no longer than the category cooldown, or using another would stack them, and no shorter than
// the longest cooldown the item runs on, or it is one used in combat. 0 where any of that fails.
func categoryOneAtATime(onUse *proto.ItemEffect) int32 {
	row := spelldata.Find(onUse.BuffId)
	use := onUse.GetOnUse()
	category, categoryMs, cooldownMs := use.GetCategoryId(), use.GetCategoryCooldownMs(), use.GetCooldownMs()
	if category <= 0 {
		category = int32(row.Category)
	}
	if categoryMs < 0 {
		categoryMs = row.CategoryCooldownMs
	}
	if cooldownMs < 0 {
		cooldownMs = row.CooldownMs
	}
	if category <= 0 || categoryMs <= 0 {
		return 0
	}

	buffs := 0
	for _, s := range shared.ConsumableSpells(row) {
		if s == spelldata.Nil || s.WhileSeated() || !slices.ContainsFunc(s.Effects, func(e spelldata.Effect) bool {
			return e.ParsedAura() && shared.LandsOnTheUser(&e)
		}) {
			continue
		}
		buffs++
		if s.DurationMs <= 0 || s.DurationMs > categoryMs || s.DurationMs < max(cooldownMs, categoryMs) {
			return 0
		}
	}
	if buffs == 0 {
		return 0
	}
	return category
}

// No elixir db.json ships has a group from its rows, and every scroll has one. A stat consumable
// other than an elixir or a scroll is grouped by its category exactly where its rows make it one used
// one at a time with the others of that category.
func TestTheRowsGroupEveryScrollNoElixirAndByCategoryOnlyWhereTheCategoryOutlastsTheBuff(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	grouped, ungrouped := 0, 0
	for _, id := range slices.Sorted(maps.Keys(instance.Consumables)) {
		catalogued := instance.Consumables[id]
		routing := RouteConsumable(&catalogued, instance)
		if !ShipsConsumable(routing) {
			continue
		}
		consumable := ShippedConsumable(&catalogued, routing, instance)
		group := rowGroupOf(&catalogued, consumable, routing, instance)
		if consumable.Type == proto.ConsumableType_ConsumableTypeElixir && group.Key != "" {
			t.Errorf("elixir %d %s is grouped as %q, want no group", id, catalogued.Name, group.Key)
		}
		if consumable.Type == proto.ConsumableType_ConsumableTypeScroll && group.Key == "" {
			t.Errorf("scroll %d %s applies no aura to group it by", id, catalogued.Name)
		}
		if consumable.Type == proto.ConsumableType_ConsumableTypeElixir || consumable.Type == proto.ConsumableType_ConsumableTypeScroll ||
			registrationFamilyOf(consumable, &catalogued, routing) != statFamily {
			if strings.HasPrefix(group.Key, "category ") {
				t.Errorf("%v %d %s is grouped by its category %q, want only a stat consumable grouped so", consumable.Type, id, catalogued.Name, group.Key)
			}
			continue
		}

		want := ""
		if category := categoryOneAtATime(consumable.ItemEffects[0]); category > 0 {
			want = fmt.Sprintf("category %d", category)
			grouped++
		} else {
			ungrouped++
		}
		if group.Key != want {
			t.Errorf("%d %s is grouped as %q, want %q", id, catalogued.Name, group.Key, want)
		}
	}
	if grouped == 0 || ungrouped == 0 {
		t.Errorf("%d stat consumables grouped by category and %d not, want some of each", grouped, ungrouped)
	}
}

func TestEveryExclusiveGroupOverrideShips(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	seen := map[int32]bool{}
	for _, group := range ConsumableExclusiveGroups {
		for _, id := range group.Items {
			if seen[id] {
				t.Errorf("%d is in more than one group", id)
			}
			seen[id] = true
			shippedConsumable(t, instance, int(id))
		}
	}
}
