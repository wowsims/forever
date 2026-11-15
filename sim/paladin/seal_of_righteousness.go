package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The damage spell each rank fires on a hit. No row triggers them and none is on a skill line, so no
// ladder holds them and the store carries no row for them; the proc is built by hand. Rank 1 is
// 21084, which fires 25741 (subtext "Rank 1.5"); 25742 ("Rank 1") belongs to the level 1 starter
// seal 20154, which 21084 supersedes and the ladder leaves out.
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
// The judgement's dummy, its second effect, carries the number the per-hit damage is built from:
// the seal's tooltip renders that number over 87 and over 25, which is the flat "value per hundred
// times swing speed" at speeds 1.15 and 4.0. Its description variables add the spell power part,
// $1h=${0.0765*$SPH} to the low end and $2h=${0.12*$SPH} to the high end. The damage spells carry
// no spell power coefficient, and the judgement dummy's per-rank coefficient (0.058 to 0.2) is not
// read.
//
//	proc = value/100 * speed * hand(speed) + coefficient * (spell power + Judgement of the Crusader)
//	  two-hander:  hand = 1.2,                                  coefficient 0.12
//	  one-hander:  hand = 0.85 - 0.2 * max(0, 2.0 - speed),     coefficient 0.1
//
// The two-hander line is exact. One-handers sit on 0.85 from 2.0 up and lose a little per tenth of
// a second below that; the straight line through the fast points is one of two shapes the data
// allows (EV-0059). The spell power part takes one coefficient per hand type at every speed and
// rank, outside the fast-weapon penalty, and Judgement of the Crusader counts as spell power at
// the same coefficient. A crit doubles the unrounded total.
//
// The tooltip's 0.0765 fits none of the one-handers, whose coefficient is 0.1 (EV-0060).
func (paladin *Paladin) registerSealOfRighteousness(_ int32, rank *spelldata.Spell) {
	judgeRank := spellData.JudgementOfRighteousness.ByID(int32(rank.EffectN(2).BaseValue()))
	judgeDamage := judgeRank.DamageEffect()
	perHit := judgeRank.EffectN(2)

	// The judgement is Ranged in SpellCategories and carries No Active Defense: it rolls hit and
	// crit on the ranged table with the paladin's ranged hit chance, cannot be blocked, and is a
	// ranged attack to every proc that listens for one.
	judgementConfig := spelldata.SpellConfig(&paladin.Unit, judgeRank, spelldata.Melee(core.ProcMaskRangedSpecial), spelldata.RemoveFlags(core.SpellFlagAPL), spelldata.Flags(core.SpellFlagBinary))
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

		// No BonusCoefficient: the coefficient depends on the hand, and Judgement of the Crusader
		// sits inside the base with the spell power, under the done multipliers, so the proc adds
		// both by hand.
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

// The measured weapon multiplier and spell power coefficient of a Seal of Righteousness hit for the
// weapon in the main hand: see registerSealOfRighteousness.
func sealOfRighteousnessHand(mh *core.Item) (hand float64, coefficient float64) {
	if mh.HandType == proto.HandType_HandTypeTwoHand {
		return 1.2, 0.12
	}
	return 0.85 - 0.2*max(0, 2.0-mh.SwingSpeed), 0.1
}
