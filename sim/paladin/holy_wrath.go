package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var HolyWrathRankMap = spellData.HolyWrath

// Holy Wrath
// https://www.wowhead.com/forever/spell=10318
//
// Sends bolts of holy power in all directions, causing 498 to 584 Holy damage to all Undead and Demon
// targets within 20 yds and stunning them for 2 sec.
func (paladin *Paladin) registerHolyWrath(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = SpellMaskHolyWrath
	// The 20 yd area is SpellRadius 9 on effect 1's EffectRadiusIndex_1, which the store does not carry.
	// TODO: Read rank.EffectN(1).RadiusMax once the store falls back to EffectRadiusIndex_1.
	config.MaxRange = 20

	config.Cast.ModifyCast = func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
		castTime := paladin.ApplyCastSpeedForSpell(cast.CastTime, spell)
		paladin.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+castTime)
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		results := []*core.SpellResult{}
		for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
			if aoeTarget.MobType == proto.MobType_MobTypeUndead || aoeTarget.MobType == proto.MobType_MobTypeDemon {
				results = append(results, spell.CalcDamage(sim, aoeTarget, damage.Average(), spell.OutcomeMagicHitAndCrit))
			}
		}

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			for _, result := range results {
				spell.DealDamage(sim, result)
			}
		})
	}

	paladin.RegisterSpell(config)
}
