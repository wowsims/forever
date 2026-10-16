package shaman

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// One shield is up at a time, so the shield proc rate option spends a charge of whichever it is.
func (shaman *Shaman) registerShieldsSpells() {
	spendWaterShield := shaman.registerWaterShieldSpell()
	spendLightningShield := shaman.registerLightningShieldSpell()
	shaman.ShieldSelfProc = func(sim *core.Simulation) {
		spendLightningShield(sim, shaman.CurrentTarget)
		spendWaterShield(sim)
	}
}

// The shield proc rate option stands for the hits a shaman takes when nothing attacks it in the sim.
func (shaman *Shaman) startShieldProcPeriodicAction(sim *core.Simulation) {
	if shaman.SelfBuffs.ShieldProcrate > 0 && shaman.ShieldSelfProc != nil {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   time.Duration(float64(time.Minute) / shaman.SelfBuffs.ShieldProcrate),
			Priority: core.ActionPriorityGCD,
			OnAction: func(sim *core.Simulation) {
				shaman.ShieldSelfProc(sim)
			},
		})
	}
}

var waterShieldRank = spellData.WaterShield.Highest()
var waterShieldMana = spellData.WaterShieldTriggered.Highest()

// The talent teaches Water Shield: three globes for 10 minutes, one of which a hit taken or one of the
// shaman's spell crits spends, at most once every 3.5 s, to restore 408511's 2% of maximum mana. The
// row hears both on one mask, but its tooltip asks for a crit only of the shaman's own spells ("when a
// spell, melee, or ranged attack hits the caster or when one of the caster's spells gets a critical
// result"), so the two are separate triggers sharing the globes and the cooldown. Answers what spends a
// globe for the shield proc rate option.
func (shaman *Shaman) registerWaterShieldSpell() func(sim *core.Simulation) {
	if !shaman.Talents.WaterShield {
		return func(*core.Simulation) {}
	}

	manaMetrics := shaman.NewManaMetrics(core.ActionID{SpellID: waterShieldMana.ID})
	manaShare := waterShieldMana.EffectN(1).Percent()

	shaman.WaterShieldAura = shaman.RegisterAura(spelldata.AuraConfig(waterShieldRank))

	icd := core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: waterShieldRank.ICD(),
	}
	spendGlobe := func(sim *core.Simulation) {
		if !shaman.WaterShieldAura.IsActive() || !icd.IsReady(sim) {
			return
		}
		icd.Use(sim)
		shaman.WaterShieldAura.RemoveStack(sim)
		shaman.AddMana(sim, shaman.MaxMana()*manaShare, manaMetrics)
	}
	handler := func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		spendGlobe(sim)
	}

	taken := spelldata.ProcTrigger(&shaman.Character, waterShieldRank, handler)
	taken.Name = "Water Shield Trigger - Hit Taken"
	taken.ICD = 0
	taken.Callback = core.CallbackOnSpellHitTaken
	taken.ProcMask &^= core.ProcMaskSpellHealing
	taken.Outcome = core.OutcomeLanded
	shaman.WaterShieldAura.AttachProcTrigger(taken)

	crit := spelldata.ProcTrigger(&shaman.Character, waterShieldRank, handler)
	crit.Name = "Water Shield Trigger - Spell Crit"
	crit.ICD = 0
	crit.Callback = core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt
	crit.ProcMask = core.ProcMaskSpellDamage | core.ProcMaskSpellHealing
	shaman.WaterShieldAura.AttachProcTrigger(crit)

	config := spelldata.SpellConfig(&shaman.Unit, waterShieldRank, spelldata.Flags(core.SpellFlagAPL|SpellFlagShamanSpell|SpellFlagInstant))
	config.ProcMask = core.ProcMaskEmpty
	config.RelatedSelfBuff = shaman.WaterShieldAura
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		shaman.LightningShieldAura.Deactivate(sim)
		shaman.WaterShieldAura.Activate(sim)
		shaman.WaterShieldAura.SetStacks(sim, shaman.WaterShieldAura.MaxStacks)
	}
	shaman.RegisterSpell(config)

	return spendGlobe
}

var lightningShieldRank = spellData.LightningShield.Highest()

// The shield's effect 1 fires 26545, a dummy, and the orb its tooltip names deals the damage.
var lightningShieldOrb = lightningShieldRank.Refs()[0]

// 10432 holds three charges for 10 minutes, and a hit taken spends one, at most once every 3.5 s, to
// strike the attacker with the orb.
func (shaman *Shaman) registerLightningShieldSpell() func(sim *core.Simulation, attacker *core.Unit) {
	orbConfig := spelldata.SpellConfig(&shaman.Unit, lightningShieldOrb,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc(), spelldata.Flags(core.SpellFlagProc|SpellFlagShamanSpell))
	orbConfig.ClassSpellMask = SpellMaskLightningShield
	damage := lightningShieldOrb.DamageEffect()
	orbConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// The row cannot crit.
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHit)
	}
	orb := shaman.RegisterSpell(orbConfig)

	shaman.LightningShieldAura = shaman.RegisterAura(spelldata.AuraConfig(lightningShieldRank))

	// The cooldown is the trigger's own, shared with the shield proc rate option.
	icd := core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: lightningShieldRank.ICD(),
	}
	spendCharge := func(sim *core.Simulation, attacker *core.Unit) {
		if !shaman.LightningShieldAura.IsActive() || !icd.IsReady(sim) {
			return
		}
		icd.Use(sim)
		shaman.LightningShieldAura.RemoveStack(sim)
		orb.Cast(sim, attacker)
	}
	trigger := spelldata.ProcTrigger(&shaman.Character, lightningShieldRank, func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
		spendCharge(sim, spell.Unit)
	})
	trigger.Name = "Lightning Shield Trigger"
	trigger.ICD = 0
	shaman.LightningShieldAura.AttachProcTrigger(trigger)

	config := spelldata.SpellConfig(&shaman.Unit, lightningShieldRank, spelldata.Flags(core.SpellFlagAPL|SpellFlagShamanSpell|SpellFlagInstant))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = SpellMaskLightningShield
	config.RelatedSelfBuff = shaman.LightningShieldAura
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		if shaman.WaterShieldAura != nil {
			shaman.WaterShieldAura.Deactivate(sim)
		}
		shaman.LightningShieldAura.Activate(sim)
		shaman.LightningShieldAura.SetStacks(sim, shaman.LightningShieldAura.MaxStacks)
	}
	shaman.LightningShield = shaman.RegisterSpell(config)

	return spendCharge
}
