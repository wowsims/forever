package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerPowerWordShield() {
	weakenedSoul := spellData.PowerWordShieldTriggered.ByID(6788)
	priest.weakenedSoulAuras = priest.NewAllyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(spelldata.AuraConfig(weakenedSoul))
	})

	// Every rank puts up the same shield on a target, so a cast of any rank replaces the one up. The
	// rank receives credit for added capacity after subtracting the unused shield it replaces.
	highest := spellData.PowerWordShield.Highest()
	schools := core.SpellSchool(highest.AbsorbEffect().Misc)
	absorbLabel := spelldata.Label(highest.Name + " - " + priest.Label)
	var strength float64
	castBy := make([]*core.Spell, len(priest.Env.AllUnits))
	absorbs := priest.NewAllyDamageAbsorptionAuraArray(func(target *core.Unit) *core.DamageAbsorptionAura {
		absorb := target.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
			Aura: spelldata.AuraConfig(highest, absorbLabel),
			ShieldStrengthCalculator: func(_ *core.Unit) float64 {
				return strength
			},
			ShouldApplyToResult: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult, _ bool) bool {
				return spell.SpellSchool.Matches(schools)
			},
		})
		// The APL's Multishield reads the spell's Shield, which ends with the absorb.
		absorb.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
			castBy[target.UnitIndex].Shield(target).Deactivate(sim)
		})
		return absorb
	})
	shieldAura := spelldata.AuraConfig(highest)
	shieldAura.ActionID = shieldAura.ActionID.WithTag(1)

	spellData.PowerWordShield.Each(func(_ int32, rank *spelldata.Spell) {
		baseAbsorb := rank.AbsorbEffect()

		config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
		config.ClassSpellMask = PriestSpellPowerWordShield
		config.BonusCoefficient = baseAbsorb.Coeff()
		config.RelatedAuraArrays = priest.weakenedSoulAuras.ToMap()
		config.Shield = core.ShieldConfig{Aura: shieldAura}
		config.ExtraCastCondition = func(_ *core.Simulation, target *core.Unit) bool {
			target = priest.healTarget(target)
			return !priest.weakenedSoulAuras.Get(target).IsActive()
		}
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = priest.healTarget(target)
			priest.weakenedSoulAuras.Get(target).Activate(sim)
			// A shield takes the spell's own modifiers (Improved Power Word: Shield), not healing done
			// or healing taken.
			strength = (baseAbsorb.Average() + spell.BonusCoefficient*spell.HealingPower(target)) *
				spell.DamageMultiplier * spell.DamageMultiplierAdditive
			absorb := absorbs.Get(target)
			gained := max(0, strength-absorb.ShieldStrength)
			castBy[target.UnitIndex] = spell
			absorb.Activate(sim)
			spell.Shield(target).Activate(sim)
			spell.RecordShielding(target, gained)
			spell.SpellMetrics[target.UnitIndex].Hits++
		}

		priest.RegisterSpell(config)
	})
}
