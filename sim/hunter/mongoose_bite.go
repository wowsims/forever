package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var mongooseBiteRank = spellData.MongooseBite.Highest()

// The window a dodge opens: the aura Defensive State fires on the dodge.
var mongooseBiteWindow = spellData.DefensiveStateTriggered.Highest()

func (hunter *Hunter) registerMongooseBite() {
	hunter.MongooseBiteAura = hunter.RegisterAura(spelldata.AuraConfig(mongooseBiteWindow))

	hunter.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Mongoose Bite - Trigger",
		TriggerImmediately: true,
		Outcome:            core.OutcomeDodge,
		Callback:           core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			hunter.MongooseBiteAura.Activate(sim)
		},
	})

	hunter.registerExposePrey()

	config := spelldata.SpellConfig(&hunter.Unit, mongooseBiteRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return hunter.MongooseBiteAura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := mongooseBiteRank.WeaponStrike(sim, spell, hunter.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		hunter.MongooseBiteAura.Deactivate(sim)
	}

	hunter.RegisterSpell(config)
}

// Expose Prey opens the same window off any attack on a target carrying Hunter's Mark. The rate is
// the effect ladder the tooltip's $m1 names - 5/10 by rank - and the mark is the caller's condition.
func (hunter *Hunter) registerExposePrey() {
	if hunter.Talents.ExposePrey == 0 {
		return
	}

	trigger := spelldata.ProcTrigger(&hunter.Character, spellData.ExposePrey.Rank(hunter.Talents.ExposePrey),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			hunter.MongooseBiteAura.Activate(sim)
		})
	trigger.Name = "Expose Prey - Trigger"
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
		return result.Target.HasActiveAuraWithTag(buffs.HuntersMarkCategory)
	}

	hunter.MakeProcTriggerAura(trigger)
}
