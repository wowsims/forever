package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Package-level state the commented-out implementations used:
// var faerieFireFeralRank = spellData.FaerieFireFeral.ByID(27011)

var faerieFireRank = spellData.FaerieFire.Highest()

func (druid *Druid) registerFaerieFireSpell() {
	druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		// The druid's own copy of the debuff; the game files have no talent that raises it.
		return buffs.NewSpellDataFaerieFireAura(target, true, 0)
	})

	config := spelldata.SpellConfig(&druid.Unit, faerieFireRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = DruidSpellFaerieFire
	config.FlatThreatBonus = 132 // The game files state no threat for Faerie Fire.

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			druid.FaerieFireAuras.Get(target).Activate(sim)
		}
	}
	config.RelatedAuraArrays = druid.FaerieFireAuras.ToMap()

	druid.FaerieFire = druid.RegisterSpell(config)
}

// TODO: uncalled -- Forever drops the Faerie Fire (Feral) talent; re-gate before wiring
// back into RegisterFeralCatSpells/RegisterFeralTankSpells.
// TODO: To be implemented. Forever has no separate Feral version. Plain Faerie Fire exists on the
// Balance line as spells 770, 778, 9749, 9907, and the feral variant folded into it.
func (druid *Druid) registerFaerieFireFeralSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 	// The druid's own copy of the aura; the game files have no Improved Faerie
	// 	// Fire node, so there are no talent points to pass.
	// 	return buffs.NewSpellDataFaerieFireAura(target, true, 0)
	// })
	//
	// druid.FaerieFireFeral = druid.RegisterSpell(core.SpellConfig{
	// 	ClassSpellMask:  DruidSpellFaerieFireFeral,
	// 	ActionID:        core.ActionID{SpellID: faerieFireFeralRank.ID},
	// 	CastRequirement: faerieFireFeralRank.CastRequirement(),
	// 	SpellSchool:     faerieFireFeralRank.SpellSchool(),
	// 	DefenseType:     faerieFireFeralRank.DefenseTypeCore(),
	// 	ProcMask:        core.ProcMaskSpellDamage,
	// 	Flags:           core.SpellFlagAPL,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: faerieFireFeralRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    druid.NewTimer(),
	// 			Duration: max(faerieFireFeralRank.Cooldown(), faerieFireFeralRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	ThreatMultiplier: 1,
	// 	FlatThreatBonus:  132,
	// 	MaxRange:         float64(faerieFireFeralRank.MaxRange),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
	//
	// 		if result.Landed() {
	// 			druid.FaerieFireAuras.Get(target).Activate(sim)
	// 		}
	// 	},
	//
	// 	RelatedAuraArrays: druid.FaerieFireAuras.ToMap(),
	// })
}
