package paladin

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Seal of Command (talent)
// https://www.wowhead.com/forever/spell=20920
//
// Gives the Paladin a chance to deal additional Holy damage equal to 70% of normal weapon damage.
// Only one Seal can be active on the Paladin at any one time. Lasts 30 sec.
//
// Unleashing this Seal's energy will judge an enemy, instantly causing 194 to 211 Holy damage, 388 to
// 422 if the target is stunned or incapacitated.
//
// The game files state the stunned number, and the game halves it otherwise. The judgement's damage
// sits on the spell its dummy names, one rank above the proc on the family's triggered ladder, which
// the proc spell 20424 heads with no rank subtext. 7 procs per minute is the Classic sim's number.
//
// Spell power share in EV-0061. The Holy damage the target takes extra is added after the percent.
func (paladin *Paladin) registerSealOfCommand(_ int32, rank *spelldata.Spell) {
	judgeRank := spellData.JudgementOfCommand.Rank(rank.RankNumber())
	judgeDamageRank := spellData.SealOfCommandTriggered.ByID(int32(judgeRank.EffectN(1).BaseValue()))
	judgeDamage := judgeDamageRank.DamageEffect()

	// Melee in SpellCategories with No Active Defense: hit and crit on the melee table, never dodged,
	// parried or blocked.
	judgementConfig := spelldata.SpellConfig(&paladin.Unit, judgeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.RemoveFlags(core.SpellFlagAPL))
	judgementConfig.ClassSpellMask = SpellMaskJudgementOfCommand
	// The dummy states no class flags; modifiers find the judgement by the damage spell's.
	judgementConfig.ClassFlags = judgeDamageRank.ClassFlags
	judgementConfig.BonusCoefficient = judgeDamage.Coeff()
	judgementConfig.DirectVariance = judgeDamage.DirectVariance()
	// The dummy lands on the enemy before the damage it names. Proc rolls in EV-0155, EV-0156. The
	// spell's flags leave the suppress bits off a dummy, so they are set here.
	dummyConfig := spelldata.SpellConfig(&paladin.Unit, judgeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial),
		spelldata.Proc(), spelldata.Tag(1), spelldata.Flags(core.SpellFlagNoMetrics|core.SpellFlagNoLogs), spelldata.Triggered())
	if judgeRank.SuppressesWeaponProcs() {
		dummyConfig.Flags |= core.SpellFlagSuppressWeaponProcs
	}
	dummyLanding := paladin.RegisterSpell(dummyConfig)

	judgementConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := judgeDamage.Average()
		if !target.PseudoStats.Stunned {
			baseDamage /= 2
		}
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		if result.Landed() {
			dummyLanding.CalcAndDealOutcome(sim, target, dummyLanding.OutcomeAlwaysHit)
		}
		spell.DealDamage(sim, result)
	}
	judgement := paladin.RegisterSpell(judgementConfig)

	procRank := spellData.SealOfCommandTriggered.Rank(1)
	procEffect := procRank.EffectN(1)
	weaponPercent := procEffect.Percent() * spellData.ImprovedSeals.MultiplierAt(paladin.Talents.ImprovedSeals)
	coefficient := procEffect.Coeff()
	// 20424 carries Not a Proc: an ability hit to every listener, weapon procs included. No
	// BonusCoefficient: spell power goes through the weapon percent, and core adds the Holy damage the
	// target takes extra at the full coefficient.
	procConfig := spelldata.SpellConfig(&paladin.Unit, procRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc())
	procConfig.ClassSpellMask = SpellMaskSealOfCommandProc
	procConfig.TakenBonusCoefficient = coefficient
	procConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		weaponDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
		baseDamage := (weaponDamage + coefficient*spell.SpellDamage(target)) * weaponPercent

		baseDamage += coefficient * spell.Unit.AttackTables[target.UnitIndex].MobTypeBonusStats[target.MobType][stats.SpellDamage]

		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		dealAfterBatch(sim, spell, result)
	}
	procSpell := paladin.RegisterSpell(procConfig)
	echoSpell := paladin.registerEchoHit(procConfig, echoOfCommandID)

	// The Echo rolls the seal's chance ("a chance to activate Seal of Command"); its stored copies ignore
	// the seal's 1 sec ICD and do not start it.
	dpm := paladin.NewLegacyPPMManager(7, core.ProcMaskMeleeWhiteHit)
	icd := core.Cooldown{Timer: paladin.NewTimer(), Duration: time.Second}
	rollProc := func(sim *core.Simulation) bool {
		return dpm.Proc(sim, core.ProcMaskMeleeMHAuto, "Seal of Command")
	}

	aura := paladin.makeSealExclusive(paladin.RegisterAura(core.Aura{
		Label:    rank.RankedLabel(&paladin.Unit),
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: sealDuration,
	}).AttachProcTrigger(core.ProcTrigger{
		Callback:       core.CallbackOnSpellHitDealt,
		ProcMask:       swingMask,
		Outcome:        core.OutcomeLanded,
		ExtraCondition: hearsSwing,
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			if icd.IsReady(sim) && rollProc(sim) {
				procSpell.Cast(sim, result.Target)
				icd.Use(sim)
			}
		},
	}))

	paladin.registerSealSpell(&sealConfig{
		rank:      rank,
		classMask: SpellMaskSealOfCommand,
		aura:      aura,
		judgement: judgement,
		echoID:    echoOfCommandID,
		echo: func(sim *core.Simulation, target *core.Unit) {
			if rollProc(sim) {
				echoSpell.Cast(sim, target)
			}
		},
	})
}
