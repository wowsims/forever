package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerHexOfWeakness(_ int32, rank *spelldata.Spell) {
	// The attack power reduction bids in the raid's Hex of Weakness category, so only the strongest
	// of it and Demoralizing Shout counts; the healing reduction bids on its own.
	auras := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.RegisterAura(spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &priest.Unit)))
		spelldata.ParseEffects(nil, aura, rank, spelldata.Level(priest.Level),
			spelldata.Exclusive(buffs.HexOfWeaknessCategory, false))
		return aura
	})

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskEmpty))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			auras.Get(target).Activate(sim)
		}
	}
	config.RelatedAuraArrays = auras.ToMap()

	priest.RegisterSpell(config)
}
