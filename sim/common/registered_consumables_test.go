//go:build with_db

package common_test

import (
	"testing"

	assetsdb "github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core"
	_ "github.com/wowsims/forever/sim/rogue"
)

// Every consumable db.json ships is one the sim applies: a hand-written or data-driven registration
// for its id. The rogue registers its poisons itself, on import.
func TestEveryShippedConsumableIsRegistered(t *testing.T) {
	registerAllEffects()

	var missing []int32
	for _, consumable := range assetsdb.Load().Consumables {
		if !core.HasConsumableEffect(consumable.Id) {
			missing = append(missing, consumable.Id)
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d consumables in db.json register no effect: %v", len(missing), missing)
	}
}
