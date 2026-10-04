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

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.ID},
		SpellSchool:    rank.SpellSchool(),
		DefenseType:    rank.DefenseTypeCore(),
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskLayOnHands,
		Rank:           rank.RankNumber(),
		MaxRange:       float64(rank.MaxRange),

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: rank.GCD(),
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: cooldown(rank),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Unit.SpendMana(sim, spell.Unit.CurrentMana(), manaMetrics)

			if target.HasManaBar() {
				target.AddMana(sim, manaRestore, manaMetrics)
			}
			spell.CalcAndDealHealing(sim, target, spell.Unit.MaxHealth(), spell.OutcomeHealingCrit)
		},
	})
}
