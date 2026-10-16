package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Granted by the Bane of Havoc talent: while the Bane is on a target, 15% of the damage the warlock deals
// to other targets is dealt to it as well. The effect is a dummy, so the share is read off its value. It
// holds one target at a time, and is a Bane: landing it takes any other Bane of this warlock's off the target.
func (warlock *Warlock) registerBaneOfHavoc() {
	rankedSpell := spellData.BaneOfHavoc.ByID(1225228)
	havocShare := rankedSpell.EffectN(1).Percent()

	// The target the Bane is on, if any
	var cursed *core.Unit

	warlock.BaneOfHavocAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		auraConfig := spelldata.AuraConfig(rankedSpell, spelldata.RankLabel(rankedSpell, &warlock.Unit))
		auraConfig.OnGain = func(aura *core.Aura, _ *core.Simulation) {
			cursed = aura.Unit
		}
		auraConfig.OnExpire = func(aura *core.Aura, _ *core.Simulation) {
			if cursed == aura.Unit {
				cursed = nil
			}
		}
		return target.GetOrRegisterAura(auraConfig)
	})

	// The share dealt again: already the damage the warlock dealt, so nothing modifies it, and it hears no
	// damage callbacks so it cannot feed itself.
	havocDamage := warlock.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: rankedSpell.ID, Tag: 1},
		SpellSchool:      rankedSpell.SpellSchool(),
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagIgnoreModifiers | core.SpellFlagIgnoreResists | core.SpellFlagNoOnDamageDealt | core.SpellFlagPassiveSpell,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	})
	mirror := func(sim *core.Simulation, result *core.SpellResult) {
		if cursed != nil && result.Target != cursed && result.Damage > 0 {
			havocDamage.CalcAndDealDamage(sim, cursed, result.Damage*havocShare, havocDamage.OutcomeAlwaysHit)
		}
	}
	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "Bane of Havoc Listener",
		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			mirror(sim, result)
		},
		OnPeriodicDamageDealt: func(_ *core.Aura, sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			mirror(sim, result)
		},
	}))

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	config.RelatedAuraArrays = warlock.BaneOfHavocAuras.ToMap()
	// TODO: 1225228 has no SpellCooldowns row, so it is off the GCD where the other Banes have a 1.5 s one; confirm in game.

	// Roll to hit, and move the Bane onto the target in place of any other Bane there
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			warlock.deactivateOtherBanes(sim, spell, target)
			if cursed != nil && cursed != target {
				warlock.BaneOfHavocAuras.Get(cursed).Deactivate(sim)
			}
			warlock.BaneOfHavocAuras.Get(target).Activate(sim)
		}
		spell.DealOutcome(sim, result)
	}

	warlock.BaneOfHavoc = warlock.RegisterSpell(config)
	warlock.Banes = append(warlock.Banes, warlock.BaneOfHavoc)
}
