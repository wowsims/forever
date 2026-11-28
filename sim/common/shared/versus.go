package shared

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The factor the spell deals or restores against its targets, kept for the source on the stats sheet,
// or nil where the spell states none.
func RegisterVersus(character *core.Character, source core.ActionID, kind proto.VersusEffectKind, row *spelldata.Spell) *core.VersusBonus {
	bonus, ok := row.VersusBonus()
	if !ok {
		return nil
	}
	return character.RegisterVersus(source, kind, bonus)
}

// The spell damage the spell's A_MOD_DAMAGE_DONE_CREATURE effect grants against the mob types its mask
// names while the aura is up, credited to the spell. The parse table reads this aura on an item, a
// consumable or a set bonus as physical and spell damage. A class set bonus that raises "damage done by
// your spells and effects" against a creature type grants spell damage alone, and registers through this
// helper.
func AttachSpellDamageVersus(aura *core.Aura, row *spelldata.Spell) *core.Aura {
	for i := range row.Effects {
		if effect := &row.Effects[i]; effect.Aura == dbcenums.A_MOD_DAMAGE_DONE_CREATURE {
			return aura.AttachMobTypeBonusStats(core.ActionID{SpellID: row.ID}, stats.Stats{stats.SpellDamage: effect.BaseValue()}, spelldata.MobTypes(effect.Misc)...)
		}
	}
	panic(fmt.Sprintf("spell %d has no A_MOD_DAMAGE_DONE_CREATURE effect", row.ID))
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
