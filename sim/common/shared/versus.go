package shared

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The factor the spell deals or restores against its targets, kept for the item on the stats sheet, or
// nil where the spell states none.
func ItemVersusBonus(character *core.Character, itemID int32, kind proto.VersusEffectKind, row *spelldata.Spell) *core.VersusBonus {
	bonus, ok := row.VersusBonus()
	if !ok {
		return nil
	}
	return character.NewItemVersusBonus(itemID, kind, bonus)
}
