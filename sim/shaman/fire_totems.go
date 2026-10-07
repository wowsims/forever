package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var searingTotemRank = spellData.SearingTotem.Highest()
var searingAttackRank = spellData.SearingTotemTriggered.Highest()

// 10587's tooltip names the pulse, 10581, and the passive that casts it, 10584.
var magmaTotemRank = spellData.MagmaTotem.Highest()
var magmaPulseRank = spellData.MagmaTotemTriggered.ByID(10581)
var magmaPassiveRank = spellData.MagmaTotemTriggered.ByID(10584)

var fireNovaRank = spellData.FireNova.Highest()
var fireNovaDamageRank = spellData.FireNovaTriggered.Highest()

// A fire totem's own attack: the totem's, not the caster's, so it lands no proc of the caster's.
func (shaman *Shaman) newTotemAttackConfig(rank *spelldata.Spell) core.SpellConfig {
	return spelldata.SpellConfig(&shaman.Unit, rank,
		spelldata.Magic(core.ProcMaskEmpty),
		spelldata.Proc(),
		spelldata.Flags(SpellFlagShamanSpell))
}

// A fire totem that acts every period while it stands. Nothing it does lands before the pull.
func (shaman *Shaman) registerFireTotem(rank *spelldata.Spell, mask int64, period core.PeriodicActionOptions) *core.Spell {
	duration := rank.Duration()

	onAction := period.OnAction
	period.OnAction = func(sim *core.Simulation) {
		if sim.CurrentTime >= 0 {
			onAction(sim)
		}
	}

	totemAura := shaman.RegisterAura(spelldata.AuraConfig(rank)).AttachPeriodicAction(period)

	config := shaman.newTotemSpellConfig(rank)
	config.ClassSpellMask = mask
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		shaman.dropTotem(sim, FireTotem, totemAura, duration)
	}

	return shaman.RegisterSpell(config)
}

// The totem attacks its target once a cast of its attack, 10436's 2.2 s.
func (shaman *Shaman) registerSearingTotemSpell() {
	damage := searingAttackRank.DamageEffect()

	attackConfig := shaman.newTotemAttackConfig(searingAttackRank)
	attackConfig.ClassSpellMask = SpellMaskSearingTotem
	attackConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}
	attack := shaman.RegisterSpell(attackConfig)

	shaman.SearingTotem = shaman.registerFireTotem(searingTotemRank, SpellMaskSearingTotem, core.PeriodicActionOptions{
		Period: searingAttackRank.CastTime(),
		OnAction: func(sim *core.Simulation) {
			attack.Cast(sim, shaman.CurrentTarget)
		},
	})
}

// The passive casts the pulse on every enemy around the totem each of its periods.
func (shaman *Shaman) registerMagmaTotemSpell() {
	damage := magmaPulseRank.DamageEffect()

	pulseConfig := shaman.newTotemAttackConfig(magmaPulseRank)
	pulseConfig.ClassSpellMask = SpellMaskMagmaTotem
	pulseConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamage(sim, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}
	pulse := shaman.RegisterSpell(pulseConfig)

	shaman.MagmaTotem = shaman.registerFireTotem(magmaTotemRank, SpellMaskMagmaTotem, core.PeriodicActionOptions{
		Period: magmaPassiveRank.EffectN(1).Period(),
		OnAction: func(sim *core.Simulation) {
			pulse.Cast(sim, shaman.CurrentTarget)
		},
	})
}

// "Instantly inflicts $11307s1 fire damage to enemies within $11307a1 yd of your active Fire totem."
// The cast deals 11307's damage itself, so it carries both rows' class flags: Improved Fire Nova
// names the cast, 408345, and Call of Flame and Elemental Fury name the damage, 11307.
func (shaman *Shaman) registerFireNovaSpell() {
	damage := fireNovaDamageRank.DamageEffect()

	config := spelldata.SpellConfig(&shaman.Unit, fireNovaRank,
		spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Flags(SpellFlagShamanSpell|SpellFlagInstant))
	config.ClassSpellMask = SpellMaskFireNova
	config.BonusCoefficient = damage.Coeff()
	config.DirectVariance = damage.DirectVariance()
	for i := range config.ClassFlags.Mask {
		config.ClassFlags.Mask[i] |= fireNovaDamageRank.ClassFlags.Mask[i]
	}

	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		totem := shaman.TotemAuras[FireTotem]
		return totem != nil && totem.IsActive()
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamage(sim, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	shaman.FireNova = shaman.RegisterSpell(config)
}
