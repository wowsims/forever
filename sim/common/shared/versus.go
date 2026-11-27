package shared

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The factor the row deals or restores against its targets, kept for the source on the stats sheet,
// or nil where the row states none.
func RegisterVersus(character *core.Character, source core.ActionID, kind proto.VersusEffectKind, row *spelldata.Spell) *core.VersusBonus {
	bonus, ok := row.VersusBonus()
	if !ok {
		return nil
	}
	return character.RegisterVersus(source, kind, bonus)
}

// Narrows a proc to hits on the targets, before any condition it already has.
func OnlyAgainst(targets core.VersusTargets) spelldata.ProcOpt {
	return func(_ *core.Character, trigger *core.ProcTrigger) {
		condition := trigger.ExtraCondition
		trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			return targets.Matches(result.Target) && (condition == nil || condition(sim, spell, result))
		}
	}
}
