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

// Elixirs whose first aura is the same share one group, and a different first aura, another stat or
// another school, is another group. The elixirs granting a crit chance share one group whatever
// their first aura. A scroll shares its group with its ranks alone.
func TestAConsumableTakesItsExclusiveGroup(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	groupNames := map[string]string{}
	for _, tc := range []struct {
		name string
		ids  []int
	}{
		{"Strength: Elixir of Greater Strength, Elixir of Ogre Strength", []int{9206, 3391}},
		{"Agility: Elixir of Greater Agility, Elixir of Agility", []int{9187, 8949}},
		{"Critical Strike: the Mongoose (Agility then crit), the Grizzly, the Owl, Sages, the Whale", []int{13452, 250351, 250337, 250338, 250348}},
		{"Intellect: Elixir of the Sages (Intellect then Spirit), Elixir of Greater Intellect", []int{13447, 9179}},
		{"Armor: Elixir of Greater Defense", []int{13445}},
		{"Attack Power: Winterfall Firewater, Distilled Firewater", []int{12820, 246948}},
		{"Physical Damage: Elixir of Minor Force", []int{247755}},
		{"Arcane Elixir, every school", []int{9155}},
		{"Elixir of Fire Power, fire alone", []int{6373}},
		{"Gift of Arthas, shadow resistance", []int{9088}},
		{"Scroll of Strength IV and Scroll of Strength", []int{10310, 954}},
		{"Scroll of Agility IV and Scroll of Agility", []int{10309, 3012}},
		{"Scroll of Intellect IV", []int{10308}},
		{"Scroll of Protection IV", []int{10305}},
		{"Scroll of Spirit IV", []int{10306}},
		{"Scroll of Stamina IV", []int{10307}},
		{"Scroll of Minor Evocation, A_PERIODIC_ENERGIZE", []int{213564}},
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
		{"Juju Power, which counts with the elixirs instead", 12451, "", ""},
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

// An elixir's group is keyed by its first aura's type and misc value and named after what that aura
// grants, and a scroll's the same way, apart from the elixirs'.
func TestElixirGroupNamesWhatTheAuraGrants(t *testing.T) {
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
		{dbcenums.A_MOD_CRIT_PCT, 0, "290/0", "Critical Strike"},
	} {
		if got, want := ElixirGroup(tc.aura, tc.misc), (ConsumableGroupKey{Key: tc.key, Name: tc.name}); got != want {
			t.Errorf("aura %v misc %d: group %+v, want %+v", tc.aura, tc.misc, got, want)
		}
	}
	if got, want := ScrollGroup(dbcenums.A_MOD_STAT, 1), (ConsumableGroupKey{Key: "scroll 29/1", Name: "Agility"}); got != want {
		t.Errorf("scroll of Agility: group %+v, want %+v", got, want)
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

// Every elixir and scroll db.json ships has a group, no scroll one an elixir has, and the crit group
// holds the five elixirs that grant a crit chance. A stat consumable other than an elixir or a scroll
// is grouped by its category exactly where its rows make it one used one at a time with the others of
// that category.
func TestTheRowsGroupEveryElixirAndByCategoryOnlyWhereTheCategoryOutlastsTheBuff(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	grouped, ungrouped := 0, 0
	elixirKeys, scrollKeys := map[string]bool{}, map[string]bool{}
	var crit []int
	for _, id := range slices.Sorted(maps.Keys(instance.Consumables)) {
		catalogued := instance.Consumables[id]
		routing := RouteConsumable(&catalogued, instance)
		if !ShipsConsumable(routing) {
			continue
		}
		consumable := ShippedConsumable(&catalogued, routing, instance)
		if !KeepConsumable(consumable) {
			continue
		}
		group := rowGroupOf(&catalogued, consumable, routing, instance)
		if (consumable.Type == proto.ConsumableType_ConsumableTypeElixir || consumable.Type == proto.ConsumableType_ConsumableTypeScroll) && group.Key == "" {
			t.Errorf("%v %d %s applies no aura to group it by", consumable.Type, id, catalogued.Name)
		}
		switch consumable.Type {
		case proto.ConsumableType_ConsumableTypeElixir:
			elixirKeys[group.Key] = true
			if group == ElixirCritGroup {
				crit = append(crit, id)
			}
		case proto.ConsumableType_ConsumableTypeScroll:
			scrollKeys[group.Key] = true
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
	for key := range scrollKeys {
		if elixirKeys[key] {
			t.Errorf("scroll group %q is an elixir group too", key)
		}
	}
	if want := []int{13452, 250337, 250338, 250348, 250351}; !slices.Equal(crit, want) {
		t.Errorf("crit group holds %v, want %v", crit, want)
	}
}

// The consumables ConsumablesCountWithElixirs names ship marked, and no other consumable is.
func TestTheConsumablesThatCountWithTheElixirsShipMarked(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, id := range slices.Sorted(maps.Keys(instance.Consumables)) {
		catalogued := instance.Consumables[id]
		routing := RouteConsumable(&catalogued, instance)
		if !ShipsConsumable(routing) {
			if slices.Contains(ConsumablesCountWithElixirs, int32(id)) {
				t.Errorf("%d %s counts with the elixirs but does not ship", id, catalogued.Name)
			}
			continue
		}
		if got, want := ShippedConsumable(&catalogued, routing, instance).CountsWithElixirs, slices.Contains(ConsumablesCountWithElixirs, int32(id)); got != want {
			t.Errorf("%d %s counts with the elixirs: %v, want %v", id, catalogued.Name, got, want)
		}
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
