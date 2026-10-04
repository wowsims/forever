package shaman

import (
	"time"

	"github.com/wowsims/forever/sim/core"
)

// Package-level state the commented-out implementations used:
// var waterShieldRank = spellData.WaterShield.ByID(33736)

func (shaman *Shaman) registerShieldsSpells() {
	shaman.registerWaterShieldSpell()
	shaman.registerLightningShieldSpell()
	shaman.registerShieldEffectTriggerSpell()
}

// TODO: To be implemented. Port the TBC Shield Effect Trigger Spell implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) registerShieldEffectTriggerSpell() {
	// The TBC implementation, kept for the port:
	// shaman.ShieldSelfProcSpell = shaman.RegisterSpell(core.SpellConfig{
	// 	Flags:          core.SpellFlagNoMetrics | core.SpellFlagNoLogs,
	// 	ClassSpellMask: SpellMaskShieldSelfProc,
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		spell.CalcAndDealDamage(sim, target, 0, spell.OutcomeAlwaysHit)
	// 	},
	// })
}

func (shaman *Shaman) startShieldProcPeriodicAction(sim *core.Simulation) {
	// TODO: ShieldSelfProcSpell is nil until the shields are implemented.
	if shaman.SelfBuffs.ShieldProcrate > 0 && shaman.ShieldSelfProcSpell != nil {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   60 * time.Second / time.Duration(shaman.SelfBuffs.ShieldProcrate),
			Priority: core.ActionPriorityGCD,
			OnAction: func(sim *core.Simulation) {
				shaman.ShieldSelfProcSpell.Cast(sim, &shaman.Unit)
			},
		})
	}
}

// TODO: To be implemented. The ability exists: spell 408510 on the Restoration line. No rank subtext,
// so no generated table -- pin the id directly.
func (shaman *Shaman) registerWaterShieldSpell() {
	// The TBC implementation, kept for the port:
	// bonusManaReturned := 0.0
	// mp5 := 50.0
	// if shaman.CouldHaveSetBonus(ItemSetTidefuryRaiment, 4) {
	// 	bonusManaReturned = 56
	// }
	//
	// actionID := core.ActionID{SpellID: waterShieldRank.ID}
	// waterShieldManaMetrics := shaman.NewManaMetrics(actionID)
	//
	// shaman.WaterShieldAura = shaman.RegisterAura(core.Aura{
	// 	Label:     "Water Shield",
	// 	ActionID:  actionID,
	// 	Duration:  10 * time.Minute,
	// 	MaxStacks: 3,
	// }).AttachProcTrigger(core.ProcTrigger{
	// 	Name:           "Water Shield Trigger",
	// 	Callback:       core.CallbackOnSpellHitTaken,
	// 	ICD:            3500 * time.Millisecond,
	// 	ClassSpellMask: SpellMaskShieldSelfProc,
	// 	Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
	// 		shaman.WaterShieldAura.RemoveStack(sim)
	// 		shaman.AddMana(sim, waterShieldRank.DamageEffect().Average(core.CharacterLevel)+bonusManaReturned, waterShieldManaMetrics)
	// 	},
	// }).AttachStatBuff(stats.MP5, mp5)
	//
	// shaman.RegisterSpell(core.SpellConfig{
	// 	ActionID:    actionID,
	// 	SpellSchool: core.SpellSchoolNature,
	// 	DefenseType: core.DefenseTypeMagic,
	// 	Flags:       core.SpellFlagAPL | SpellFlagInstant,
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: waterShieldRank.GCD(),
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		shaman.LightningShieldAura.Deactivate(sim)
	// 		shaman.WaterShieldAura.Activate(sim)
	// 		shaman.WaterShieldAura.SetStacks(sim, 3)
	// 	},
	// 	RelatedSelfBuff: shaman.WaterShieldAura,
	// })
}

var lightningShieldRank = spellData.LightningShield.Highest()

// TODO: To be implemented. Port the TBC Lightning Shield Spell implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) registerLightningShieldSpell() {
	// The TBC implementation, kept for the port:
	// actionID := core.ActionID{SpellID: lightningShieldRank.ID}
	//
	// // The Direct amount sits on a generic dummy aura rather than a school-damage effect.
	// lightningShieldEffect := lightningShieldRank.EffectN(1)
	//
	// lsDamage := shaman.RegisterSpell(core.SpellConfig{
	// 	ActionID:         core.ActionID{SpellID: lightningShieldRank.ID},
	// 	SpellSchool:      lightningShieldRank.SpellSchool(),
	// 	DefenseType:      lightningShieldRank.DefenseTypeCore(),
	// 	ProcMask:         core.ProcMaskEmpty,
	// 	Flags:            SpellFlagShamanSpell,
	// 	ClassSpellMask:   SpellMaskLightningShield,
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	BonusCoefficient: lightningShieldEffect.Coeff(),
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := lightningShieldEffect.Average(core.CharacterLevel)
	// 		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	// 	},
	// })
	//
	// shaman.LightningShieldAura = shaman.RegisterAura(core.Aura{
	// 	Label:     "Lightning Shield",
	// 	ActionID:  actionID,
	// 	Duration:  10 * time.Minute,
	// 	MaxStacks: 3,
	// }).AttachProcTrigger(core.ProcTrigger{
	// 	Name:           "Lightning Shield Trigger",
	// 	Callback:       core.CallbackOnSpellHitTaken,
	// 	ICD:            3500 * time.Millisecond,
	// 	ClassSpellMask: SpellMaskShieldSelfProc,
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
	// 		shaman.LightningShieldAura.RemoveStack(sim)
	// 		lsDamage.Cast(sim, shaman.CurrentTarget)
	// 	},
	// })
	//
	// shaman.RegisterSpell(core.SpellConfig{
	// 	ActionID:    actionID,
	// 	SpellSchool: core.SpellSchoolNature,
	// 	DefenseType: core.DefenseTypeMagic,
	// 	Flags:       core.SpellFlagAPL | SpellFlagInstant,
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(lightningShieldRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: lightningShieldRank.GCD(),
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		shaman.WaterShieldAura.Deactivate(sim)
	// 		shaman.LightningShieldAura.Activate(sim)
	// 		shaman.LightningShieldAura.SetStacks(sim, 3)
	// 	},
	// 	RelatedSelfBuff: shaman.LightningShieldAura,
	// })
}
