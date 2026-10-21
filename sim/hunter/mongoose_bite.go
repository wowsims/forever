package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var mongooseBiteRank = spellData.MongooseBite.Highest()

// The window the hunter's own dodge opens: the aura Defensive State fires when the hunter dodges an attack.
var mongooseBiteWindow = spellData.DefensiveStateTriggered.Highest()

func (hunter *Hunter) registerMongooseBite() {
	hunter.MongooseBiteAura = hunter.RegisterAura(spelldata.AuraConfig(mongooseBiteWindow))

	hunter.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Mongoose Bite - Trigger",
		TriggerImmediately: true,
		// Defensive State (13358) hears melee swings and abilities taken, not ranged attacks.
		ProcMask: core.ProcMaskMelee,
		Outcome:  core.OutcomeDodge,
		Callback: core.CallbackOnSpellHitTaken,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			hunter.MongooseBiteAura.Activate(sim)
		},
	})

	exposePrey := hunter.registerExposePrey()

	config := spelldata.SpellConfig(&hunter.Unit, mongooseBiteRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return hunter.MongooseBiteAura.IsActive() || exposePrey.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := mongooseBiteRank.WeaponStrike(sim, spell, hunter.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		// TODO: Ingame test needed: which window a bite spends when both are up.
		hunter.MongooseBiteAura.Deactivate(sim)
		exposePrey.Deactivate(sim)
	}

	hunter.RegisterSpell(config)
}

// The window Expose Prey opens: the aura its proc casts.
var exposePreyWindow = spellData.ExposePreyTriggered.Highest()

// Expose Prey opens its own Mongoose Bite window off any attack on a target carrying Hunter's Mark.
// The rate is the effect ladder the tooltip's $m1 names, 5/10 by rank, and the mark is the caller's
// condition. Without the talent the window is nil, which reads as an aura that is down.
func (hunter *Hunter) registerExposePrey() *core.Aura {
	if hunter.Talents.ExposePrey == 0 {
		return nil
	}

	window := hunter.RegisterAura(spelldata.AuraConfig(exposePreyWindow))
	trigger := spelldata.ProcTrigger(&hunter.Character, spellData.ExposePrey.Rank(hunter.Talents.ExposePrey),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			window.Activate(sim)
		})
	trigger.Name = "Expose Prey - Trigger"
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
		return result.Target.HasActiveAuraWithTag(buffs.HuntersMarkCategory)
	}

	hunter.MakeProcTriggerAura(trigger)
	return window
}
