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
// sits on the spell its own dummy names, which the family's triggered ladder carries one rank above
// the proc. The proc spell 20424 has no rank subtext and heads that ladder: its 70% of weapon damage
// and the 0.29 coefficient on that same effect are read from it, and 7 procs per minute is the
// number the Classic sim carries.
//
// The coefficient sits on the weapon-percent effect, and the percent applies to the paladin's own
// spell power as well as the weapon: at 27 spell power the Forever beta measured 5.5 of each proc
// coming from spell power, 20.4%, against the 29% the coefficient states. Improved Seals raises the
// percent itself (the tooltip reads 80% with the talent, 70 * 1.15). The Holy damage the target
// takes extra, Judgement of the Crusader and the damage-against-mob-type gear, is added after the
// percent at the full coefficient, so it scales at 29% whatever the talent.
func (paladin *Paladin) registerSealOfCommand(_ int32, rank *spelldata.Spell) {
	judgeRank := spellData.JudgementOfCommand.Rank(rank.RankNumber())
	judgeDamageRank := spellData.SealOfCommandTriggered.ByID(int32(judgeRank.EffectN(1).BaseValue()))
	judgeDamage := judgeDamageRank.DamageEffect()

	// Melee in SpellCategories with No Active Defense: hit and crit on the melee table, never
	// dodged, parried or blocked.
	judgementConfig := spelldata.SpellConfig(&paladin.Unit, judgeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.RemoveFlags(core.SpellFlagAPL))
	judgementConfig.ClassSpellMask = SpellMaskJudgementOfCommand
	// The judgement's dummy states no class flags; the damage spell it names does, and modifiers find
	// the judgement by those.
	judgementConfig.ClassFlags = judgeDamageRank.ClassFlags
	judgementConfig.BonusCoefficient = judgeDamage.Coeff()
	judgementConfig.DirectVariance = judgeDamage.DirectVariance()
	judgementConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := judgeDamage.Average()
		if !target.PseudoStats.Stunned {
			baseDamage /= 2
		}
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
	}
	judgement := paladin.RegisterSpell(judgementConfig)

	procRank := spellData.SealOfCommandTriggered.Rank(1)
	procEffect := procRank.EffectN(1)
	weaponPercent := procEffect.Percent() * spellData.ImprovedSeals.MultiplierAt(paladin.Talents.ImprovedSeals)
	coefficient := procEffect.Coeff()
	// 20424 carries Not a Proc: it is an ability hit to every listener, weapon procs included. No
	// BonusCoefficient: the spell power goes through the weapon percent, so the proc adds it by hand
	// below rather than letting CalcDamage add it on top. The Holy damage the target takes extra is
	// taken at the full coefficient after the done multipliers, where core adds it.
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

	// The seal and its Echo roll the same chance: an Echo of Command "empowers your next melee
	// attack with a chance to activate Seal of Command". The Echo's stored copies ignore the seal's
	// 1 sec ICD and do not start it, so a spend of three copies rolls three times and can proc three
	// times.
	dpm := paladin.NewLegacyPPMManager(7, core.ProcMaskMeleeWhiteHit)
	icd := core.Cooldown{Timer: paladin.NewTimer(), Duration: time.Second}
	rollProc := func(sim *core.Simulation, target *core.Unit) bool {
		if !dpm.Proc(sim, core.ProcMaskMeleeMHAuto, "Seal of Command") {
			return false
		}
		procSpell.Cast(sim, target)
		return true
	}

	aura := paladin.makeSealExclusive(paladin.RegisterAura(core.Aura{
		Label:    rank.RankedLabel(&paladin.Unit),
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: sealDuration,
	}).AttachProcTrigger(core.ProcTrigger{
		Callback: core.CallbackOnSpellHitDealt,
		ProcMask: core.ProcMaskMeleeWhiteHit,
		Outcome:  core.OutcomeLanded,
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			if icd.IsReady(sim) && rollProc(sim, result.Target) {
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
			rollProc(sim, target)
		},
	})
}
