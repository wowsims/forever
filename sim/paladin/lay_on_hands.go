package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var LayOnHandsRankMap = spellData.LayOnHands

// Lay on Hands
// https://www.wowhead.com/forever/spell=10310
//
// Heals a friendly target for an amount equal to the Paladin's maximum health and restores 550
// of their mana. Drains all of the Paladin's remaining Mana when used, but does not interrupt
// Mana regeneration.
func (paladin *Paladin) registerLayOnHands(_ int32, rank *spelldata.Spell) {
	// Rank 1 restores no mana and has no energize effect, which reads as 0.
	manaRestore := rank.EnergizeEffect().Average()
	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: rank.ID})

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = SpellMaskLayOnHands

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Unit.SpendMana(sim, spell.Unit.CurrentMana(), manaMetrics)

		if target.HasManaBar() {
			target.AddMana(sim, manaRestore, manaMetrics)
		}
		spell.CalcAndDealHealing(sim, target, spell.Unit.MaxHealth(), spell.OutcomeHealingCrit)
	}

	paladin.RegisterSpell(config)
}
