package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var stormstrikeRank = spellData.Stormstrike.Highest()

// The slots on the shaman's attack table against a target that the debuffs parsed with
// spelldata.FromCaster hold.
const (
	DDBC_Stormstrike int = iota
	DDBC_Total
)

// Effect 2 raises the shaman's own Lightning Bolt, Chain Lightning and Earth Shock against the target
// for its one charge, which the next of them to hit spends.
func (shaman *Shaman) stormstrikeDebuffAura(target *core.Unit) *core.Aura {
	consumes := stormstrikeRank.EffectN(2).ClassFlags

	config := spelldata.AuraConfig(stormstrikeRank, spelldata.Label("Stormstrike-"+shaman.Label))
	config.OnSpellHitTaken = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if spell.Unit == &shaman.Unit && result.Landed() && consumes.Reaches(spell) {
			aura.RemoveStack(sim)
		}
	}
	aura := target.GetOrRegisterAura(config)
	spelldata.ParseEffects(nil, aura, stormstrikeRank, spelldata.Effects(2),
		spelldata.FromCaster(&shaman.Unit, DDBC_Stormstrike, DDBC_Total))
	return aura
}

// TODO: patch preview: the patch notes raise the strike to 150% normalized weapon damage, which the
// row's effect does not state.
const stormstrikeWeaponPercent = 1.5

// Effect 1 is a main-hand strike for normalized weapon damage.
func (shaman *Shaman) registerStormstrikeSpell() {
	shaman.StormstrikeDebuffAuras = shaman.NewEnemyAuraArray(shaman.stormstrikeDebuffAura)

	config := spelldata.SpellConfig(&shaman.Unit, stormstrikeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ClassSpellMask = SpellMaskStormstrike
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := stormstrikeRank.WeaponStrike(sim, spell, shaman.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		baseDamage *= stormstrikeWeaponPercent
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		if result.Landed() {
			aura := shaman.StormstrikeDebuffAuras.Get(target)
			aura.Activate(sim)
			aura.SetStacks(sim, aura.MaxStacks)
		}
	}

	shaman.Stormstrike = shaman.RegisterSpell(config)
}
