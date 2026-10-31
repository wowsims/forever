package paladin

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Seal of Fury
// https://www.wowhead.com/forever/spell=20423
//
// Fills the Paladin with divine fury for 30 sec, causing melee attacks to deal an additional 35 Holy
// damage. While a shield is equipped, each attack also grants an absorb shield equal to 50% of the
// Holy damage dealt. Only one Seal can be active on the Paladin at any one time.
//
// Unleashing this Seal's energy causes 168 to 182 Holy damage to an enemy and taunts the target to
// attack you for 4 sec.
func (paladin *Paladin) registerSealOfFury(_ int32, rank *spelldata.Spell) {
	judgementRank := spellData.SealOfFuryTriggered.ByID(int32(rank.EffectN(3).BaseValue()))
	judgementDamage := judgementRank.DamageEffect()
	procRank := rank.Refs()[0]
	procDamage := procRank.DamageEffect()

	if paladin.sealOfFuryTaunt == nil {
		tauntRank := spellData.SealOfFuryTriggered.Rank(3)

		tauntAuras := paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return target.GetOrRegisterAura(spelldata.AuraConfig(tauntRank))
		})

		tauntSpellConfig := spelldata.SpellConfig(&paladin.Unit, tauntRank, spelldata.Flags(core.SpellFlagNoOnCastComplete))
		tauntSpellConfig.ProcMask = core.ProcMaskEmpty
		tauntSpellConfig.Cast.DefaultCast.NonEmpty = true
		tauntSpellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				tauntAuras.Get(target).Activate(sim)
			}
		}

		paladin.sealOfFuryTaunt = paladin.RegisterSpell(tauntSpellConfig)
	}

	// Melee in SpellCategories with No Active Defense: hit and crit on the melee table, never
	// dodged, parried or blocked.
	judgementConfig := spelldata.SpellConfig(&paladin.Unit, judgementRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.RemoveFlags(core.SpellFlagAPL), spelldata.Flags(core.SpellFlagBinary))
	judgementConfig.ClassSpellMask = SpellMaskJudgementOfFury
	judgementConfig.BonusCoefficient = judgementDamage.Coeff()
	judgementConfig.DirectVariance = judgementDamage.DirectVariance()
	judgementConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, judgementDamage.Average(), spell.OutcomeMeleeSpecialHitAndCrit)
		paladin.sealOfFuryTaunt.Cast(sim, target)
	}
	judgement := paladin.RegisterSpell(judgementConfig)

	shieldShare := rank.EffectN(2).Percent()
	var pendingShield float64
	shield := paladin.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
		Aura: core.Aura{
			Label:    fmt.Sprintf("Seal of Fury Shield%s Rank %d", paladin.Label, rank.RankNumber()),
			ActionID: core.ActionID{SpellID: rank.ID}.WithTag(1),
			Duration: sealDuration,
		},
		ShieldStrengthCalculator: func(_ *core.Unit) float64 {
			return pendingShield
		},
	})

	damage := procDamage.Mean()
	// The damage spells carry the same flags as Seal of Righteousness's: procs themselves, and no
	// Suppress Weapon Procs, so a weapon's "Chance on hit" rolls on the seal's hit too.
	procConfig := spelldata.SpellConfig(&paladin.Unit, procRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc(), spelldata.Flags(core.SpellFlagProc))
	procConfig.ClassSpellMask = SpellMaskSealOfFuryProc
	procConfig.BonusCoefficient = procDamage.Coeff()
	procConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage, spell.OutcomeMeleeSpecialCritOnly)
		sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{
			DoAt:     sim.CurrentTime + core.SpellBatchWindow,
			Priority: core.ActionPriorityLow,
			OnAction: func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
				if paladin.PseudoStats.CanBlock && result.Damage > 0 {
					pendingShield = result.Damage * shieldShare
					shield.Activate(sim)
				}
			},
		}))
	}
	procSpell := paladin.RegisterSpell(procConfig)

	aura := paladin.makeSealExclusive(paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:            rank.RankedLabel(&paladin.Unit),
		ActionID:        core.ActionID{SpellID: rank.ID},
		MetricsActionID: core.ActionID{SpellID: rank.ID},
		Duration:        sealDuration,
		Callback:        core.CallbackOnSpellHitDealt,
		ProcMask:        core.ProcMaskMeleeWhiteHit,
		Outcome:         core.OutcomeLanded,
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			procSpell.Cast(sim, result.Target)
		},
	}))

	paladin.registerSealSpell(&sealConfig{
		rank:      rank,
		classMask: SpellMaskSealOfFury,
		aura:      aura,
		judgement: judgement,
		echoID:    echoOfFuryID,
		echo: func(sim *core.Simulation, target *core.Unit) {
			procSpell.Cast(sim, target)
		},
	})
}
