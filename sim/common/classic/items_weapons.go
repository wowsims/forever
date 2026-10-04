package classic

import (
	"time"

	"github.com/wowsims/forever/sim/common/itemhelpers"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	// Thunderfury, Blessed Blade of the Windseeker
	itemhelpers.CreateWeaponProcTrigger(itemhelpers.WeaponProcTrigger{
		ItemID: 19019,
		Name:   "Thunderfury",
		PPM:    6,
		Handler: func(character *core.Character) core.ProcHandler {
			procActionID := core.ActionID{SpellID: 21992}

			attackSpeedDebuffAura := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
				aura := target.GetOrRegisterAura(core.Aura{
					Label:    "Cyclone",
					ActionID: core.ActionID{SpellID: 27648},
					Duration: time.Second * 12,
				})

				buffs.AtkSpeedReductionEffect(aura, core.MultiplierFromPercent(20))

				return aura
			})

			singleTargetSpell := character.RegisterSpell(core.SpellConfig{
				ActionID:    procActionID.WithTag(1),
				SpellSchool: core.SpellSchoolNature,
				DefenseType: core.DefenseTypeMagic,
				ProcMask:    core.ProcMaskSpellDamage,
				Flags:       core.SpellFlagProc,

				DamageMultiplier: 1,
				ThreatMultiplier: 0.5,

				ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					result := spell.CalcAndDealDamage(sim, target, 300, spell.OutcomeMagicHitAndCrit)
					if result.Landed() {
						attackSpeedDebuffAura.Get(result.Target).Activate(sim)
					}
				},
			})

			resistanceDebuffAura := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
				return target.GetOrRegisterAura(core.Aura{
					Label:    "Thunderfury",
					ActionID: procActionID,
					Duration: time.Second * 12,
					OnGain: func(aura *core.Aura, sim *core.Simulation) {
						target.AddStatDynamic(sim, stats.NatureResistance, -25)
					},
					OnExpire: func(aura *core.Aura, sim *core.Simulation) {
						target.AddStatDynamic(sim, stats.NatureResistance, 25)
					},
				})
			})

			bounceSpell := character.RegisterSpell(core.SpellConfig{
				ActionID:    procActionID.WithTag(2),
				SpellSchool: core.SpellSchoolNature,
				DefenseType: core.DefenseTypeMagic,
				ProcMask:    core.ProcMaskEmpty,

				ThreatMultiplier: 1,
				FlatThreatBonus:  63,

				ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					results := spell.CalcCleaveDamage(sim, target, 5, 0, spell.OutcomeMagicHit)
					for _, result := range results {
						if result.Landed() {
							resistanceDebuffAura.Get(result.Target).Activate(sim)
						}
					}
					spell.DealBatchedAoeDamage(sim)
				},
			})

			return func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				singleTargetSpell.Cast(sim, result.Target)
				bounceSpell.Cast(sim, result.Target)
			}
		},
	})

}
