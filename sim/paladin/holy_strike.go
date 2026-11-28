package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var HolyStrikeRankMap = spellData.HolyStrike

// Holy Strike
// https://www.wowhead.com/forever/spell=10333
//
// An instant strike that causes 50% weapon damage plus an additional 97 to 121 as Holy damage.
//
// The row states the flat part as its normalized-weapon-damage effect and the percentage as a
// second effect, and the percentage scales the flat part too.
//
// The spell power coefficient sits on the flat effect, but the paladin's own spell power goes
// through the weapon percent. The strike takes no spell damage against the target's creature type
// (EV-0189), and Judgement of the Crusader is taken at the full coefficient after the done multipliers.
func (paladin *Paladin) registerHolyStrike(_ int32, rank *spelldata.Spell) {
	coefficient := rank.DamageEffect().Coeff()

	// No BonusCoefficient: the spell power goes through the weapon percent, so the strike adds it by hand.
	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ClassSpellMask = SpellMaskHolyStrike
	config.TakenBonusCoefficient = coefficient

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		weaponDamage := spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		baseDamage := rank.WeaponStrike(sim, spell, weaponDamage, coefficient*spell.SpellDamage(target))
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
	}

	paladin.RegisterSpell(config)
}
