package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shankRank = spellData.Shank.Highest()
var venomBuff = spellData.VenomTriggered.Highest()
var bootKnifePoints = spellData.BootKnifeTriggered.Highest()

// The row is Always Hit, so only the crit is rolled. Shank strikes with the off hand and carries an off-hand special's
// mask, so the off hand's poison and weapon procs, the procs that hear melee specials, Blade Flurry and the modifiers
// scoped to the off hand (Dual Wield Specialization, Hack and Slash) all reach it. Its row adds no combo point, so it
// is no builder and Seal Fate does not hear it. Venom goes up before the damage, so the poison the hit procs rolls at
// Venom's raised chance and deals its raised damage (EV-0222).
func (rogue *Rogue) registerShank() {
	config := spelldata.SpellConfig(&rogue.Unit, shankRank, spelldata.Melee(core.ProcMaskMeleeOHSpecial))
	config.DamageMultiplier = shankRank.DamageEffect().Percent()

	// Flash of Steel's SPELLMOD_EFFECT2 raises the energy of the second effect; the parse would read it as
	// damage, so it is added here.
	energy := shankRank.EffectN(2).Average() + spellData.FlashOfSteel.EffectAt(1).ValueAt(rogue.Talents.FlashOfSteel)

	var venomAura *core.Aura
	if rogue.Talents.Venom {
		venomAura = rogue.RegisterAura(spelldata.AuraConfig(venomBuff))
		spelldata.ParseEffects(&rogue.Character, venomAura, venomBuff)
	}

	// Boot Knife's proc row names no class flags, so a trigger built from it would hear every melee special, while its
	// tooltip gives the points to Shank alone. Shank rolls them itself, as finishers roll Ruthlessness.
	bootKnifeChance := spellData.BootKnife.EffectAt(1).FractionAt(rogue.Talents.BootKnife)
	var bootKnifeMetrics *core.ResourceMetrics
	if rogue.Talents.BootKnife > 0 {
		bootKnifeMetrics = rogue.NewComboPointMetrics(core.ActionID{SpellID: bootKnifePoints.ID})
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		if venomAura != nil {
			venomAura.Activate(sim)
		}
		baseDamage := rogue.OHWeaponDamage(sim, spell.MeleeAttackPower(target))
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialCritOnly)
		rogue.AddEnergy(sim, energy, spell.EnergyMetrics())
		if !result.Landed() {
			return
		}

		if bootKnifeMetrics != nil && sim.Proc(bootKnifeChance, "Boot Knife") {
			rogue.AddComboPoints(sim, target, int32(bootKnifePoints.ProcEnergizeEffect().Average()), bootKnifeMetrics)
		}
	}

	rogue.Shank = rogue.RegisterSpell(config)
}
