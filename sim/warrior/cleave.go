package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var cleaveRank = spellData.Cleave.Highest()
var cleaveTargets = int32(cleaveRank.DamageEffect().ChainTargets)

func (warrior *Warrior) registerCleave() {
	config := spelldata.SpellConfig(&warrior.Unit, cleaveRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	// TODO: Ingame research needed if this adds flat threat
	config.FlatThreatBonus = 0

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := cleaveRank.WeaponStrike(sim, spell, warrior.MHWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		results := spell.CalcCleaveDamage(sim, target, cleaveTargets, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		spell.DealBatchedAoeDamage(sim)
		if !results[0].Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
