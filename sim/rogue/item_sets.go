package rogue

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The rows ItemSetSpell names for each set's thresholds.
var (
	bonescytheHeadRush = spelldata.MustFind(28812) // 4pc: Head Rush, the proc
	bonescytheThreat   = spelldata.MustFind(28811) // 6pc: less threat
)

// Tier 3
var ItemSetBonescytheArmor = core.NewItemSet(core.ItemSet{
	ID:    524,
	Name:  "Bonescythe Armor",
	Slots: shared.TierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		// TODO: 2pc (28816) Invigorate: normal melee swings have a chance to heal the wearer for 100.
		4: applyBonescytheHeadRush,
		6: shared.ParsedSetBonus(bonescytheThreat),
		// TODO: 8pc (28814) Revealed Flaw: Eviscerate has a chance per combo point to give the next Backstab,
		// Sinister Strike or Hemorrhage a 100% critical strike chance.
	},
})

// "Your Backstab, Sinister Strike, and Hemorrhage critical hits cause you to regain 5 energy."
//
// The row hears every landed melee ability, and its class flags name Eviscerate, Shank and Mutilate too. The
// tooltip's three spells and its crit are set here; the cooldown is the row's.
func applyBonescytheHeadRush(agent core.Agent, setBonusAura *core.Aura) {
	rogue := agent.(RogueAgent).GetRogue()
	trigger := spelldata.ProcTrigger(&rogue.Character, bonescytheHeadRush, nil)
	trigger.Handler = shared.FiredSpellHandler(&rogue.Character, core.ActionID{SpellID: bonescytheHeadRush.FiredSpell().ID}, bonescytheHeadRush)
	trigger.Name = "Bonescythe Armor - 4PC"
	trigger.ClassSpellMask = RogueSpellBackstab | RogueSpellSinisterStrike | RogueSpellHemorrhage
	trigger.Outcome = core.OutcomeCrit

	setBonusAura.AttachProcTrigger(trigger)
}
