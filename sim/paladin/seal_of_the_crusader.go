package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Seal of the Crusader
// https://www.wowhead.com/forever/spell=20308
//
// Fills the Paladin with the spirit of a crusader for 30 sec, granting 325 melee attack power. The
// Paladin also attacks 40% faster, but deals less damage with each attack. Only one Seal can be
// active on the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy for 40 sec, increasing Holy damage taken by up to
// 161. Your melee strikes will refresh the spell's duration. Only one Judgement per Paladin can be
// active at any one time.
//
// The game files state no number for the damage each attack loses. In game the seal swings the weapon
// as a 1.4 times faster, lighter one: each white hit deals 1/1.4 of its damage, attack power's share
// included, and Vindication's chance on each swing drops by the same 1.4, both measured at 1.6 and 3.6
// speed. The white damage a second and the procs a minute hold.
func (paladin *Paladin) registerSealOfTheCrusader(_ int32, rank *spelldata.Spell) {
	judgementRank := spellData.SealOfTheCrusaderTriggered.ByID(int32(rank.EffectN(3).BaseValue()))
	judgementAuras := paladin.newJudgementAuras(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataJudgementOfTheCrusaderAuraRank(target, buffs.JudgementRank{
			SpellID: judgementRank.ID,
			Rank:    rank.RankNumber(),
			Value:   judgementRank.Effect(dbcenums.A_MOD_DAMAGE_TAKEN, 2).Average() + paladin.judgementOfTheCrusaderBonus,
		})
	})

	// Melee in SpellCategories and Always Hit: the debuff lands without a roll.
	judgementConfig := spelldata.SpellConfig(&paladin.Unit, judgementRank, spelldata.Melee(core.ProcMaskEmpty), spelldata.RemoveFlags(core.SpellFlagAPL))
	judgementConfig.ClassSpellMask = SpellMaskJudgementOfTheCrusader
	judgementConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
		judgementAuras.Get(target).Activate(sim)
	}
	judgementConfig.RelatedAuraArrays = judgementAuras.ToMap()
	judgement := paladin.RegisterSpell(judgementConfig)

	// The high end of the attack power, which is the number the game shows where the two differ.
	attackPower := rank.Effect(dbcenums.A_MOD_ATTACK_POWER, 0).Max() + paladin.sealOfTheCrusaderBonusAttackPower
	speed := 1 + rank.Effect(dbcenums.A_MOD_ATTACKSPEED, 0).Percent()

	aura := paladin.makeSealExclusive(paladin.RegisterAura(core.Aura{
		Label:    rank.RankedLabel(&paladin.Unit),
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: sealDuration,
	}).
		AttachStatBuff(stats.AttackPower, attackPower).
		AttachMultiplyWeaponSpeed(speed).
		AttachSpellMod(core.SpellModConfig{
			ProcMask:   core.ProcMaskMeleeMHAuto,
			Kind:       core.SpellMod_DamageDone_Pct,
			FloatValue: 1/speed - 1,
		}))

	paladin.registerSealSpell(&sealConfig{
		rank:      rank,
		classMask: SpellMaskSealOfTheCrusader,
		aura:      aura,
		judgement: judgement,
	})
}
