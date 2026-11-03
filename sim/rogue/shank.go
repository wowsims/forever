package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shankRank = spellData.Shank.Highest()
var venomBuff = spellData.VenomTriggered.Highest()
var bootKnifePoints = spellData.BootKnifeTriggered.Highest()

// The row is Always Hit, so only the crit is rolled. It suppresses the caster's, the target's and the
// weapon procs, so Shank carries no proc mask: no trinket, talent or poison hears it, and the row's Suppress
// Weapon Procs flag keeps the weapon procs off as well. With no hand in its mask, a modifier scoped to the
// off hand does not reach it either. Shank is no builder, so Seal Fate does not hear it.
func (rogue *Rogue) registerShank() {
	// TODO: validate in game that Suppress Caster Procs and Suppress Target Procs leave Shank heard by no proc.
	config := spelldata.SpellConfig(&rogue.Unit, shankRank, spelldata.Melee(core.ProcMaskEmpty))
	config.DamageMultiplier = shankRank.DamageEffect().Percent()

	// Flash of Steel's SPELLMOD_EFFECT2 raises the energy of the second effect; the parse would read it as
	// damage, so it is added here.
	energy := shankRank.EffectN(2).Average() + spellData.FlashOfSteel.EffectAt(1).ValueAt(rogue.Talents.FlashOfSteel)

	var venomAura *core.Aura
	if rogue.Talents.Venom {
		venomAura = rogue.RegisterAura(spelldata.AuraConfig(venomBuff))
		spelldata.ParseEffects(&rogue.Character, venomAura, venomBuff)
	}

	bootKnifeChance := spellData.BootKnife.EffectAt(1).FractionAt(rogue.Talents.BootKnife)
	var bootKnifeMetrics *core.ResourceMetrics
	if rogue.Talents.BootKnife > 0 {
		bootKnifeMetrics = rogue.NewComboPointMetrics(core.ActionID{SpellID: bootKnifePoints.ID})
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := rogue.OHWeaponDamage(sim, spell.MeleeAttackPower(target))
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialCritOnly)
		rogue.AddEnergy(sim, energy, spell.EnergyMetrics())
		if !result.Landed() {
			return
		}

		if venomAura != nil {
			venomAura.Activate(sim)
		}
		if bootKnifeMetrics != nil && sim.Proc(bootKnifeChance, "Boot Knife") {
			rogue.AddComboPoints(sim, int32(bootKnifePoints.ProcEnergizeEffect().Average()), bootKnifeMetrics)
		}
	}

	rogue.Shank = rogue.RegisterSpell(config)
}
