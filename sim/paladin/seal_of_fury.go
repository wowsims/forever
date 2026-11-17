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
// attack you for 3 sec.
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

	// Ranged in SpellCategories with No Active Defense: hit and crit on the ranged table, never
	// blocked, and a ranged attack to procs. No BonusCoefficient, as on the seal's hit: the
	// judgement takes no spell damage against the target's creature type (EV-0143) and adds its
	// spell power by hand, and core adds the Holy damage the target takes extra at the coefficient.
	judgementCoefficient := judgementDamage.Coeff()
	judgementConfig := spelldata.SpellConfig(&paladin.Unit, judgementRank, spelldata.Melee(core.ProcMaskRangedSpecial), spelldata.RemoveFlags(core.SpellFlagAPL), spelldata.Flags(core.SpellFlagBinary))
	judgementConfig.ClassSpellMask = SpellMaskJudgementOfFury
	judgementConfig.TakenBonusCoefficient = judgementCoefficient
	judgementConfig.DirectVariance = judgementDamage.DirectVariance()
	judgementConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, judgementDamage.Average()+judgementCoefficient*spell.SpellDamage(target), spell.OutcomeRangedHitAndCrit)
		paladin.sealOfFuryTaunt.Cast(sim, target)
	}
	judgement := paladin.RegisterSpell(judgementConfig)

	// The shield is this share of the hit in the open world, which is what the sim gives every target
	// (EV-0022). Inside a dungeon the game scales it by the creature health the
	// game files expect at a base level over the one at the mob's level (ExpectedStat.CreatureHealth),
	// where the base is the dungeon's lowest mob level in Shadowfang Keep, the Stockade, Gnomeregan,
	// Scarlet Monastery and Ruins of Lordaeron. The sim's targets carry no base level, so it models
	// none of that.
	shieldShare := rank.EffectN(2).Percent()
	var pendingShield float64
	shield := paladin.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
		Aura: core.Aura{
			Label:     fmt.Sprintf("Seal of Fury Shield%s Rank %d", paladin.Label, rank.RankNumber()),
			ActionID:  core.ActionID{SpellID: rank.ID}.WithTag(1),
			Duration:  sealDuration,
			MenuGroup: menuPaladinSeals,
		},
		ShieldStrengthCalculator: func(_ *core.Unit) float64 {
			return pendingShield
		},
	})

	damage := procDamage.Mean()
	procCoefficient := procDamage.Coeff()
	// The damage spells carry the same flags as Seal of Righteousness's: procs themselves, and no
	// Suppress Weapon Procs, so a weapon's "Chance on hit" rolls on the seal's hit too.
	// No BonusCoefficient: the hit takes the paladin's spell power but none of the spell damage
	// against the target's creature type (EV-0143), so the proc adds the spell power by hand. The
	// Holy damage the target takes extra is taken at the coefficient after the done multipliers,
	// where core adds it.
	procConfig := spelldata.SpellConfig(&paladin.Unit, procRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc(), spelldata.Flags(core.SpellFlagProc))
	procConfig.ClassSpellMask = SpellMaskSealOfFuryProc
	procConfig.TakenBonusCoefficient = procCoefficient
	procConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage+procCoefficient*spell.SpellDamage(target), spell.OutcomeMeleeSpecialCritOnly)
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
	echoSpell := paladin.registerEchoHit(procConfig, echoOfFuryID)

	aura := paladin.makeSealExclusive(paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:            rank.RankedLabel(&paladin.Unit),
		ActionID:        core.ActionID{SpellID: rank.ID},
		MetricsActionID: core.ActionID{SpellID: rank.ID},
		Duration:        sealDuration,
		Callback:        core.CallbackOnSpellHitDealt,
		ProcMask:        swingMask,
		Outcome:         core.OutcomeLanded,
		ExtraCondition:  hearsSwing,
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
			echoSpell.Cast(sim, target)
		},
	})
}
