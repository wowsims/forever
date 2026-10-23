package database

// Every consumable the catalogue selects, and the shape its on-use spell routes to. One that fits no
// shape is left out of db.json and named in unsupported_consumables.txt with the reason, beside the
// count routed to each shape, so a routing that moves is reviewed rather than discovered. Only the
// damage and heal shapes have builders so far. That file is a local baseline, not tracked: a
// checkout without it writes it on the first run.
//
// Reads the consumables the last gen_db run captured and the committed store.

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/tools/database/dbc"
)

const unsupportedConsumablesPath = "tools/database/unsupported_consumables.txt"

func TestConsumablesRouteByTheirSpell(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID int
		want   ConsumableKind
	}{
		{10646, ConsumableDamage},            // Goblin Sapper Charge: 13241
		{11566, ConsumableDamage},            // Crystal Charge, a Quest-class item: 15239
		{215127, ConsumableDamage},           // High-Yield Radiation Bomb: 435963
		{250951, ConsumableDamage},           // Caustic Smog Potion: 1252525's cloud 1252526 puts 1252543 on every enemy
		{250953, ConsumableDamage},           // Dragonfire Potion: 1252554, a cone of fire
		{13446, ConsumableHeal},              // Major Healing Potion: 17534
		{13444, ConsumableEnergize},          // Major Mana Potion: 17531
		{13510, ConsumableStatAura},          // Flask of the Titans: 17626
		{954, ConsumableStatAura},            // Scroll of Strength, cast on a friendly target: 8118
		{20749, ConsumableTemporaryEnchant},  // Brilliant Wizard Oil: 25122, enchant 2628
		{3824, ConsumableTemporaryEnchant},   // Shadow Oil: 3594, enchant 25 casts 350026 at 15%
		{277502, ConsumableTemporaryEnchant}, // Scroll of Imbue Precision: Effect 360, enchant 8718
		{20452, ConsumableStatAura},          // Smoked Desert Dumplings, Well Fed through 1248401: 1248422
		{13457, ConsumableAbsorb},            // Greater Fire Protection Potion: 17543
		{4376, ConsumableAbsorb},             // Flame Deflector: 4057
		{213564, ConsumableEnergize},         // Scroll of Minor Evocation: 434851
		{3382, ConsumableStatAura},           // Minor Troll's Blood Elixir: 3219
		{3388, ConsumableStatAura},           // Lesser Troll's Blood Elixir: 3222
		{3826, ConsumableStatAura},           // Troll's Blood Elixir: 3223
		{20004, ConsumableStatAura},          // Major Troll's Blood Elixir: 24361
		{250336, ConsumableStatAura},         // Elixir of Wicked Regeneration: 1250932
	} {
		consumable, ok := instance.Consumables[tc.itemID]
		if !ok {
			t.Errorf("%d is not in the catalogue", tc.itemID)
			continue
		}
		if routing := RouteConsumable(&consumable, instance); !routing.Supported() || routing.Kind != tc.want {
			t.Errorf("%d %s routes as %q (%s), want %q", tc.itemID, consumable.Name, routing.Kind, routing.Reason(), tc.want)
		}
	}

	wickedRegeneration := instance.Consumables[250336]
	if got, want := *wickedRegeneration.GetStatModifiers(), (stats.Stats{
		stats.HP5: equipAuraPoints(t, 1250932, dbcenums.A_MOD_HEALTH_REGEN_IN_COMBAT),
		stats.MP5: equipAuraPoints(t, 1250932, dbcenums.A_MOD_POWER_REGEN),
	}); got != want {
		t.Errorf("250336 carries %v, want 1250932's %v", got.FlatString(), want.FlatString())
	}

	blastingCharge := instance.Consumables[221325]
	if RouteConsumable(&blastingCharge, instance).Supported() {
		t.Error("221325 routes as supported, but its spell 446504 only opens a lock")
	}

	blackfathomStone := instance.Consumables[211845]
	if RouteConsumable(&blackfathomStone, instance).Supported() {
		t.Error("211845 routes as supported, but its equip spell 430391 works only in area group 7864")
	}

	for _, tc := range []struct {
		itemID int
		kind   ConsumableKind
		why    string
	}{
		{8079, ConsumableStatAura, "22734's mana comes in only while seated"},         // Conjured Crystal Water
		{9088, ConsumableStatAura, "11371's A_PROC_TRIGGER_SPELL is not parsed"},      // Gift of Arthas
		{274273, ConsumableStatAura, "1293740 states its zone bonus through A_DUMMY"}, // Flask of Natural Accuracy
		{4953, ConsumableStatAura, "5021 rolls its health"},                           // Trogg Brew
		{1401, ConsumableHeal, "833's mana comes in only while seated"},               // Green Tea Leaf
		{4596, ConsumableHeal, "1244863's E_DUMMY is not modelled"},                   // Minor Discolored Healing Potion
		{20388, ConsumableEnergize, "24707 restores only while seated"},               // Lollipop
	} {
		consumable := instance.Consumables[tc.itemID]
		if routing := RouteConsumable(&consumable, instance); routing.Supported() || routing.Kind != tc.kind {
			t.Errorf("%d %s routes as %q, supported %v, want an unsupported %q: %s", tc.itemID, consumable.Name,
				routing.Kind, routing.Supported(), tc.kind, tc.why)
		}
	}
}

func TestEveryConsumableIsRoutedOrListed(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	var out strings.Builder
	out.WriteString("# Every consumable the catalogue selects whose on-use spell fits no shape the sim builds, as\n")
	out.WriteString("# \"<item id> <spell id> <item name>: <reason>\", after the count routed to each shape. Written by\n")
	out.WriteString("# TestEveryConsumableIsRoutedOrListed in tools/database. Regenerate with\n")
	out.WriteString("#   UPDATE_UNSUPPORTED_CONSUMABLES=1 go test ./tools/database/ -run TestEveryConsumableIsRoutedOrListed -count=1\n\n")

	routed := map[ConsumableKind]int{}
	var listed strings.Builder
	for _, id := range slices.Sorted(maps.Keys(instance.Consumables)) {
		consumable := instance.Consumables[id]
		routing := RouteConsumable(&consumable, instance)
		if routing.Supported() {
			routed[routing.Kind]++
			continue
		}
		fmt.Fprintf(&listed, "%d %d %s: %s\n", id, routing.SpellID, consumable.Name, routing.Reason())
	}

	for _, kind := range ConsumableKinds {
		fmt.Fprintf(&out, "# %s: %d\n", kind, routed[kind])
	}
	out.WriteString("\n")
	out.WriteString(listed.String())

	if routed[ConsumableDamage] == 0 || routed[ConsumableStatAura] == 0 {
		t.Fatal("no consumable routes as damage or as a stat aura, so this proves nothing")
	}

	committed, err := os.ReadFile(unsupportedConsumablesPath)
	if os.IsNotExist(err) || os.Getenv("UPDATE_UNSUPPORTED_CONSUMABLES") != "" {
		if err := os.WriteFile(unsupportedConsumablesPath, []byte(out.String()), 0644); err != nil {
			t.Fatalf("%v", err)
		}
		t.Logf("wrote the baseline %s", unsupportedConsumablesPath)
		return
	}
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(committed) == out.String() {
		return
	}

	for _, line := range changedProcLines(string(committed), out.String()) {
		t.Error(line)
	}
	t.Errorf("%s is not what the rows say; once the moves above are the intended ones, rewrite the baseline:\n"+
		"  UPDATE_UNSUPPORTED_CONSUMABLES=1 go test ./tools/database/ -run TestEveryConsumableIsRoutedOrListed -count=1",
		unsupportedConsumablesPath)
}
