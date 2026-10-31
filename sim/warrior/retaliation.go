package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var retaliationRank = spellData.Retaliation.Highest()
var retaliationHit = spellData.RetaliationTriggered.Highest()
var retaliationHitBaseDamage = retaliationHit.DamageEffect().Average()

func (warrior *Warrior) registerRetaliation() {
	// The aura casts the counterattack, but it does not take spelldata.Proc(): that marks the spell
	// passive, and the metrics aggregator counts no cast for a passive spell, while the sim reports
	// every counterattack as a cast.
	hitConfig := spelldata.SpellConfig(&warrior.Unit, retaliationHit,
		spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.RemoveFlags(core.SpellFlagAPL))

	hitConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := retaliationHitBaseDamage + warrior.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
	}

	attackSpell := warrior.RegisterSpell(hitConfig)

	aura := warrior.RegisterAura(spelldata.AuraConfig(retaliationRank))
	trigger := spelldata.ProcTrigger(&warrior.Character, retaliationRank,
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			attackSpell.Cast(sim, spell.Unit)
			aura.RemoveStack(sim)
		})
	trigger.TriggerImmediately = true
	aura.AttachProcTrigger(trigger)

	config := spelldata.SpellConfig(&warrior.Unit, retaliationRank)

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		aura.Activate(sim)
		aura.SetStacks(sim, aura.MaxStacks)
	}

	config.RelatedSelfBuff = aura

	spell := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell:  spell,
		Type:   core.CooldownTypeDPS,
		Manual: true,
	})
}
