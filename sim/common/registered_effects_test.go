//go:build with_db

// Which items and enchants the sim simulates an effect for, as a local baseline that is not tracked.
// The generated registrations are a hundred lines of ids; the baseline answers the only question a
// regeneration has to answer - what became live, and what stopped being live. A checkout without it
// writes it on the first run.
//
// Needs the database: an effect registered for an item the game files do not ship is skipped, so
// without it the list would name items no sim can equip.
package common_test

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"

	assetsdb "github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
)

const registeredEffectsPath = "forever/registered_effects.txt"

var registerAllEffects = sync.OnceFunc(common.RegisterAllEffects)

func TestRegisteredEffects(t *testing.T) {
	registerAllEffects()
	db := assetsdb.Load()

	var itemIDs []int32
	for _, item := range db.Items {
		itemIDs = append(itemIDs, item.Id)
	}
	var enchantIDs []int32
	for _, enchant := range db.Enchants {
		enchantIDs = append(enchantIDs, enchant.EffectId)
	}

	var out strings.Builder
	out.WriteString("# The items and enchants sim/common registers an effect for, written by\n")
	out.WriteString("# TestRegisteredEffects in sim/common. Regenerate with\n")
	out.WriteString("#   UPDATE_REGISTERED_EFFECTS=1 GOARCH=amd64 go test --tags=with_db ./sim/common/ -run TestRegisteredEffects -count=1\n")

	out.WriteString("\n# items\n")
	for _, id := range registered(itemIDs, core.HasItemEffect) {
		fmt.Fprintf(&out, "%d %s\n", id, itemName(id))
	}

	out.WriteString("\n# enchants\n")
	for _, id := range registered(enchantIDs, core.HasEnchantEffect) {
		fmt.Fprintf(&out, "%d %s\n", id, enchantName(id))
	}

	committed, err := os.ReadFile(registeredEffectsPath)
	if os.IsNotExist(err) || os.Getenv("UPDATE_REGISTERED_EFFECTS") != "" {
		if err := os.WriteFile(registeredEffectsPath, []byte(out.String()), 0644); err != nil {
			t.Fatalf("%v", err)
		}
		t.Logf("wrote the baseline %s", registeredEffectsPath)
		return
	}
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(committed) == out.String() {
		return
	}

	for _, line := range changedLines(string(committed), out.String()) {
		t.Error(line)
	}
	t.Errorf("%s is not what the sim registers; once the moves above are the intended ones, rewrite the baseline:\n"+
		"  UPDATE_REGISTERED_EFFECTS=1 GOARCH=amd64 go test --tags=with_db ./sim/common/ -run TestRegisteredEffects -count=1",
		registeredEffectsPath)
}

// The ids the game files ship that an effect is registered for, each once, in id order.
func registered(shipped []int32, hasEffect func(int32) bool) []int32 {
	slices.Sort(shipped)
	return slices.DeleteFunc(slices.Compact(shipped), func(id int32) bool { return !hasEffect(id) })
}

// The lines one side has and the other does not, which on a sorted list is exactly what appeared and
// what vanished.
func changedLines(committed string, registered string) []string {
	was := map[string]bool{}
	for line := range strings.SplitSeq(committed, "\n") {
		was[line] = true
	}

	var changed []string
	now := map[string]bool{}
	for line := range strings.SplitSeq(registered, "\n") {
		now[line] = true
		if !was[line] && line != "" && !strings.HasPrefix(line, "#") {
			changed = append(changed, "now registered: "+line)
		}
	}

	for line := range strings.SplitSeq(committed, "\n") {
		if !now[line] && line != "" && !strings.HasPrefix(line, "#") {
			changed = append(changed, "no longer registered: "+line)
		}
	}

	return changed
}

func itemName(id int32) string {
	if item := core.GetItemByID(id); item != nil {
		return item.Name
	}
	return "?"
}

func enchantName(id int32) string {
	if enchant := core.GetEnchantByEffectID(id); enchant != nil {
		return enchant.Name
	}
	return "?"
}
