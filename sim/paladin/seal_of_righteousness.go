package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The damage spell each rank fires on a hit; no row triggers them and none is on a skill line, so the
// store has none and the proc is built by hand. Rank 1 (21084) fires 25741 ("Rank 1.5"); 25742
// ("Rank 1") is the starter seal 20154's.
var sealOfRighteousnessProcIDs = map[int32]int32{1: 25741, 2: 25740, 3: 25739, 4: 25738, 5: 25737, 6: 25736, 7: 25735, 8: 25713}

// Seal of Righteousness
// https://www.wowhead.com/forever/spell=20293
//
// Fills the Paladin with holy spirit for 30 sec, granting each melee attack additional Holy damage.
// Slower weapons cause more Holy damage per swing. Only one Seal can be active on the Paladin at
// any one time.
//
// Unleashing this Seal's energy will cause 187 to 203 Holy damage to an enemy.
//
// The seal's second effect names its judgement (the row has no effect at the game files' index 1).
// The per-hit value is the judgement dummy's second effect; the tooltip divides it by 87 and by 25,
// value/100 * speed at speeds 1.15 and 4.0. The damage spells carry no spell power coefficient and the
// dummy's is not read. Hand factors and coefficients in EV-0059.
func (paladin *Paladin) registerSealOfRighteousness(_ int32, rank *spelldata.Spell) {
	judgeRank := spellData.JudgementOfRighteousness.ByID(int32(rank.EffectN(2).BaseValue()))
	judgeDamage := judgeRank.DamageEffect()
	perHit := judgeRank.EffectN(2)

	// Ranged in SpellCategories with No Active Defense: the ranged table and hit chance, never blocked, and
	// a ranged attack to every proc that listens for one.
	judgementConfig := spelldata.SpellConfig(&paladin.Unit, judgeRank, spelldata.Melee(core.ProcMaskRangedSpecial), spelldata.RemoveFlags(core.SpellFlagAPL))
	judgementConfig.ClassSpellMask = SpellMaskJudgementOfRighteousness
	judgementConfig.BonusCoefficient = judgeDamage.Coeff()
	judgementConfig.DirectVariance = judgeDamage.DirectVariance()
	judgementConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, judgeDamage.Average(), spell.OutcomeRangedHitAndCrit)
	}
	judgement := paladin.RegisterSpell(judgementConfig)

	value := perHit.Average()
	procConfig := core.SpellConfig{
		ActionID:    core.ActionID{SpellID: sealOfRighteousnessProcIDs[rank.RankNumber()]},
		SpellSchool: core.SpellSchoolHoly,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		// The damage spells are procs, so auras without Can Proc From Procs never hear them. They
		// carry no Suppress Weapon Procs, unlike their TBC rows: a "Chance on hit" weapon effect
		// rolls on the seal's hit as well as on the swing, and the game shows both procs landing.
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagProc,
		ClassSpellMask: SpellMaskSealOfRighteousnessProc,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		// No BonusCoefficient: the coefficient depends on the hand, and Judgement of the Crusader sits in
		// the base with the spell power, under the done multipliers.
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			mh := paladin.MainHand()
			hand, coefficient := sealOfRighteousnessHand(mh)
			attackTable := spell.Unit.AttackTables[target.UnitIndex]
			baseDamage := value/100*mh.SwingSpeed*hand +
				coefficient*(spell.BonusDamage(attackTable)+target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly])
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialCritOnly)
			dealAfterBatch(sim, spell, result)
		},
	}
	procSpell := paladin.RegisterSpell(procConfig)
	echoSpell := paladin.registerEchoHit(procConfig, echoOfRighteousnessID)

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
		classMask: SpellMaskSealOfRighteousness,
		aura:      aura,
		judgement: judgement,
		echoID:    echoOfRighteousnessID,
		echo: func(sim *core.Simulation, target *core.Unit) {
			echoSpell.Cast(sim, target)
		},
	})
}

// The measured hand factor and spell power coefficient of the main-hand weapon (registerSealOfRighteousness).
func sealOfRighteousnessHand(mh *core.Item) (hand float64, coefficient float64) {
	if mh.HandType == proto.HandType_HandTypeTwoHand {
		return 1.2, 0.12
	}
	return 0.85 - 0.17*max(0, 2.0-mh.SwingSpeed), 0.1
}
