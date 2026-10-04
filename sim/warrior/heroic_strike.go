package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var heroicStrikeRank = spellData.HeroicStrike.Highest()

func (warrior *Warrior) registerHeroicStrike() {
	config := spelldata.SpellConfig(&warrior.Unit, heroicStrikeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	// TODO: Ingame research needed if this adds flat threat
	config.FlatThreatBonus = 0

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := heroicStrikeRank.WeaponStrike(sim, spell, warrior.MHWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
