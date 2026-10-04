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
// Sends bolts of holy power in all directions, causing 533 Holy damage to all Undead and Demon
// targets within 20 yds and stunning them for 2 sec. The stun has no place in the sim.
func (paladin *Paladin) registerHolyWrath(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.ID},
		SpellSchool:    rank.SpellSchool(),
		DefenseType:    rank.DefenseTypeCore(),
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHolyWrath,
		Rank:           rank.RankNumber(),
		MaxRange:       20,
		MissileSpeed:   float64(rank.Speed),

		ManaCost: manaCost(rank),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      rank.GCD(),
				CastTime: rank.CastTime(),
			},
			CD: core.Cooldown{
				Timer:    paladin.sharedTimer(&paladin.holyWrathTimer),
				Duration: cooldown(rank),
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				castTime := paladin.ApplyCastSpeedForSpell(cast.CastTime, spell)
				paladin.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+castTime)
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: damage.Coeff(),

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			results := []*core.SpellResult{}
			for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
				if aoeTarget.MobType == proto.MobType_MobTypeUndead || aoeTarget.MobType == proto.MobType_MobTypeDemon {
					results = append(results, spell.CalcDamage(sim, aoeTarget, damage.Roll(sim), spell.OutcomeMagicHitAndCrit))
				}
			}

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				for _, result := range results {
					spell.DealDamage(sim, result)
				}
			})
		},
	})
}
