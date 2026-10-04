package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var raptorStrikeRank = spellData.RaptorStrike.Highest()

func (hunter *Hunter) registerRaptorStrike() {
	config := spelldata.SpellConfig(&hunter.Unit, raptorStrikeRank,
		spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Flags(core.SpellFlagNoOnCastComplete))
	// The swing replacement below casts it whenever it is ready, so the APL is not offered it.
	config.Flags &^= core.SpellFlagAPL

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := raptorStrikeRank.WeaponStrike(sim, spell, hunter.MHWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
	}

	hunter.RaptorStrike = hunter.RegisterSpell(config)
}

// The swing a main-hand auto becomes: Raptor Strike when it can be cast, the plain swing otherwise.
func (hunter *Hunter) TryRaptorStrike(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if mhSwingSpell.ActionID.Tag != 1 || !hunter.RaptorStrike.CanCast(sim, hunter.CurrentTarget) {
		return mhSwingSpell
	}

	return hunter.RaptorStrike
}
