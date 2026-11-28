package classic

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	// Thunderfury, Blessed Blade of the Windseeker
	//
	// No rate in game files; the item's chance-on-hit entry carries it.
	core.NewItemEffect(19019, func(agent core.Agent) {
		character := agent.GetCharacter()
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

		singleTargetConfig := spelldata.SpellConfig(&character.Unit, spelldata.MustFind(procActionID.SpellID),
			spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
		singleTargetConfig.Flags |= core.SpellFlagProc
		singleTargetConfig.ThreatMultiplier = 0.5
		singleTargetConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, 300, spell.OutcomeMagicHitAndCrit)
			if result.Landed() {
				attackSpeedDebuffAura.Get(result.Target).Activate(sim)
			}
		}
		singleTargetSpell := character.RegisterSpell(singleTargetConfig)

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

		// The bounce is the same spell's resistance chain, so it resists all or nothing as the spell does.
		bounceSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    procActionID.WithTag(2),
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       singleTargetConfig.Flags & core.SpellFlagBinary,

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

		aura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:         "Thunderfury Proc",
			Callback:     core.CallbackOnSpellHitDealt,
			Outcome:      core.OutcomeLanded,
			DPM:          character.NewDynamicLegacyProcForWeapon(19019, entryPPM(core.GetItemByID(19019).ItemEffects, procActionID.SpellID), 0),
			IsWeaponProc: true,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				singleTargetSpell.Cast(sim, result.Target)
				bounceSpell.Cast(sim, result.Target)
			},
		})

		character.ItemSwap.RegisterProc(19019, aura)
	})

}

// The procs per minute an item's or an enchant's chance-on-hit entry carries for the spell.
func entryPPM(effects []*proto.ItemEffect, spellID int32) float64 {
	for _, effect := range effects {
		if effect.BuffId == spellID && effect.GetProc().GetPpm() > 0 {
			return effect.GetProc().GetPpm()
		}
	}
	panic(fmt.Sprintf("no entry states procs per minute for spell %d", spellID))
}
