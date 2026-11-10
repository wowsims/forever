package lsp

import (
	"github.com/wowsims/forever/sim/siml"
)

// The spell and item names and icons, by id, for the rotation panel: [name, icon].
type panelDatabase struct {
	Spells map[int32][2]string `json:"spells"`
	Items  map[int32][2]string `json:"items"`
}

// The items with a use, gear and then consumables: what a rotation can cast as item.<name>.
func usableItems(c *siml.DatabaseCatalog) []int32 {
	var out []int32
	for _, item := range c.Items() {
		if item.OnUse {
			out = append(out, item.ID)
		}
	}
	for _, x := range c.Consumables() {
		if x.OnUse {
			out = append(out, x.ID)
		}
	}
	return out
}
