package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var HolyStrikeRankMap = spellData.HolyStrike

// Holy Strike
// https://www.wowhead.com/forever/spell=10333
//
// An instant strike that causes 50% weapon damage plus an additional 97 to 121 as Holy damage.
//
// The row states the flat part as its normalized-weapon-damage effect and the percentage as a
// second effect, and the percentage scales the flat part too: the 2 to 4 is Rank 1's 12 ± 1.5 at 25%.
//
// The spell power coefficient sits on the flat effect, but the paladin's own spell power goes
// through the weapon percent, as it does for the Seal of Command proc: Rank 3's 0.429 at 32% is
// 0.137 per point of spell power. The damage-against-mob-type gear is added after the percent at
// the full coefficient, and the Holy damage the target takes extra, Judgement of the Crusader, is
// taken at the full coefficient after the done multipliers, where core adds it.
func (paladin *Paladin) registerHolyStrike(_ int32, rank *spelldata.Spell) {
	coefficient := rank.DamageEffect().Coeff()

	// No BonusCoefficient: the spell power goes through the weapon percent, so the strike adds it by
	// hand below rather than letting CalcDamage add it on top.
	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ClassSpellMask = SpellMaskHolyStrike
	config.TakenBonusCoefficient = coefficient

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		weaponDamage := spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		baseDamage := rank.WeaponStrike(sim, spell, weaponDamage, coefficient*spell.SpellDamage(target))

		baseDamage += coefficient * spell.Unit.AttackTables[target.UnitIndex].MobTypeBonusStats[target.MobType][stats.SpellDamage]

		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
	}

	paladin.RegisterSpell(config)
}
