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
// It reports as the totem's spell tagged 1, so the results show the totem and its hits as one row.
func (shaman *Shaman) newTotemAttackConfig(totem *spelldata.Spell, rank *spelldata.Spell) core.SpellConfig {
	config := spelldata.SpellConfig(&shaman.Unit, rank,
		spelldata.Magic(core.ProcMaskEmpty),
		spelldata.Proc(),
		spelldata.Flags(SpellFlagShamanSpell))
	config.ActionID = core.ActionID{SpellID: totem.ID, Tag: 1}
	return config
}

// A fire totem that acts every period while it stands. Nothing it does lands before the pull. The
// totem stands as long as its aura lasts, which a duration mod on the cast reaches (Trial by Fire's
// on Magma Totem).
func (shaman *Shaman) registerFireTotem(rank *spelldata.Spell, mask int64, period core.PeriodicActionOptions) *core.Spell {
	onAction := period.OnAction
	period.OnAction = func(sim *core.Simulation) {
		if sim.CurrentTime >= 0 {
			onAction(sim)
		}
	}

	totemAura := shaman.RegisterAura(spelldata.AuraConfig(rank)).AttachPeriodicAction(period)

	config := shaman.newTotemSpellConfig(rank)
	config.ClassSpellMask = mask
	config.RelatedSelfBuff = totemAura
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		shaman.dropTotem(sim, FireTotem, totemAura, totemAura.Duration)
	}

	return shaman.RegisterSpell(config)
}

// The totem attacks its target once a cast of its attack, 10436's 2.2 s.
func (shaman *Shaman) registerSearingTotemSpell() {
	damage := searingAttackRank.DamageEffect()

	attackConfig := shaman.newTotemAttackConfig(searingTotemRank, searingAttackRank)
	attackConfig.ClassSpellMask = SpellMaskSearingTotem
	attackConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
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

	pulseConfig := shaman.newTotemAttackConfig(magmaTotemRank, magmaPulseRank)
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

// "Instantly inflicts $408428s1 fire damage to enemies within $408428a1 yd of your active Fire totem."
// The cast deals nothing itself: it casts the Fire Nova row of its own rank, 408428, which hits. Each
// row carries its own class flags, so Trial by Fire's cooldown reaches the cast and its damage the
// hit, and Call of Flame and Elemental Fury name the hit alone.
// The hit lacks Not a Proc, so only listeners that can proc from procs hear it. It reports as the
// cast's spell tagged 1, so the results show the cast and its hits as one row.
func (shaman *Shaman) registerFireNovaSpell() {
	damage := fireNovaDamageRank.DamageEffect()

	novaConfig := spelldata.SpellConfig(&shaman.Unit, fireNovaDamageRank,
		spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Proc(),
		spelldata.Flags(SpellFlagShamanSpell|core.SpellFlagProc))
	novaConfig.ActionID = core.ActionID{SpellID: fireNovaRank.ID, Tag: 1}
	novaConfig.ClassSpellMask = SpellMaskFireNova
	novaConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamage(sim, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}
	nova := shaman.RegisterSpell(novaConfig)

	config := spelldata.SpellConfig(&shaman.Unit, fireNovaRank,
		spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Flags(SpellFlagShamanSpell|SpellFlagInstant))
	config.ClassSpellMask = SpellMaskFireNova
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		totem := shaman.TotemAuras[FireTotem]
		return totem != nil && totem.IsActive()
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
		nova.Cast(sim, target)
	}

	shaman.FireNova = shaman.RegisterSpell(config)
}
