package itemhelpers

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
)

// Weapon proc helpers, ported from the SoD sim. Every helper rolls the proc on the weapon's own
// PPM manager, which follows the item through item swaps, and registers the trigger aura with
// ItemSwap so the proc toggles with the weapon.

type WeaponProcTrigger struct {
	ItemID int32
	Name   string
	PPM    float64

	// Set for an "Equip" proc, which is an aura the game matches by proc flags and which never
	// hears proc hits. Left unset it means a "Chance on hit" proc: a weapon proc that rolls on
	// every hit of its weapon except those suppressing weapon procs.
	EquipProc          bool
	TriggerImmediately bool

	// Runs once per character and returns the proc handler, or nil to opt the character out.
	Handler func(character *core.Character) core.ProcHandler
}

// Registers a weapon proc whose handler runs on every landed hit that passes the weapon's PPM
// roll. The other helpers build on this one.
func CreateWeaponProcTrigger(config WeaponProcTrigger) {
	core.NewItemEffect(config.ItemID, func(agent core.Agent) {
		character := agent.GetCharacter()

		handler := config.Handler(character)
		if handler == nil {
			return
		}

		aura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               config.Name + " Proc",
			Callback:           core.CallbackOnSpellHitDealt,
			Outcome:            core.OutcomeLanded,
			DPM:                character.NewDynamicLegacyProcForWeapon(config.ItemID, config.PPM, 0),
			IsWeaponProc:       !config.EquipProc,
			TriggerImmediately: config.TriggerImmediately,
			Handler:            handler,
		})

		character.ItemSwap.RegisterProc(config.ItemID, aura)
	})
}

type WeaponProcDamage struct {
	ItemID int32
	Name   string
	PPM    float64

	SpellID int32
	School  core.SpellSchool
	// From SpellCategories. Picks the hit table and crit multiplier.
	DefenseType      core.DefenseType
	MinDmg           float64
	MaxDmg           float64
	BonusCoefficient float64

	// See WeaponProcTrigger. CreateWeaponCoHProcDamage and CreateWeaponEquipProcDamage set it.
	EquipProc bool
}

// Registers a weapon proc that deals flat damage.
func CreateWeaponProcDamage(config WeaponProcDamage) {
	shared.NewProcDamageEffect(shared.ProcDamageEffect{
		ItemID:           config.ItemID,
		SpellID:          config.SpellID,
		School:           config.School,
		DefenseType:      config.DefenseType,
		MinDmg:           config.MinDmg,
		MaxDmg:           config.MaxDmg,
		BonusCoefficient: config.BonusCoefficient,
		Flags:            core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagProc,
		Trigger: core.ProcTrigger{
			Name:         config.Name + " Proc",
			Callback:     core.CallbackOnSpellHitDealt,
			Outcome:      core.OutcomeLanded,
			IsWeaponProc: !config.EquipProc,
		},
		TriggerDPM: func(character *core.Character) *core.DynamicProcManager {
			return character.NewDynamicLegacyProcForWeapon(config.ItemID, config.PPM, 0)
		},
	})
}

// Registers a "Chance on hit" weapon damage proc.
func CreateWeaponCoHProcDamage(config WeaponProcDamage) {
	config.EquipProc = false
	CreateWeaponProcDamage(config)
}

// Registers an "Equip" weapon damage proc.
func CreateWeaponEquipProcDamage(config WeaponProcDamage) {
	config.EquipProc = true
	CreateWeaponProcDamage(config)
}

type WeaponProcSpell struct {
	ItemID int32
	Name   string
	PPM    float64

	// Runs once per character and returns the spell to cast at the hit target, or nil to opt
	// the character out (e.g. a resource proc on a class without that resource).
	Spell func(character *core.Character) *core.Spell
}

// Registers a "Chance on hit" weapon proc that casts a custom spell on the target that was hit.
func CreateWeaponProcSpell(config WeaponProcSpell) {
	CreateWeaponProcTrigger(WeaponProcTrigger{
		ItemID:             config.ItemID,
		Name:               config.Name,
		PPM:                config.PPM,
		TriggerImmediately: true,
		Handler: func(character *core.Character) core.ProcHandler {
			procSpell := config.Spell(character)
			if procSpell == nil {
				return nil
			}
			procSpell.Flags |= core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagProc

			return func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				procSpell.Cast(sim, result.Target)
			}
		},
	})
}

type WeaponProcAura struct {
	ItemID int32
	Name   string
	PPM    float64

	// Runs once per character and returns the aura the proc activates on the wearer. A stacking
	// aura is filled to its maximum on every proc.
	Aura func(character *core.Character) *core.Aura
}

// Registers a "Chance on hit" weapon proc that activates a custom aura on the wearer.
func CreateWeaponProcAura(config WeaponProcAura) {
	core.NewItemEffect(config.ItemID, func(agent core.Agent) {
		AddWeaponProcAura(agent.GetCharacter(), config)
	})
}

// Adds a "Chance on hit" weapon proc for a custom aura to an existing item effect.
func AddWeaponProcAura(character *core.Character, config WeaponProcAura) {
	procAura := config.Aura(character)

	aura := character.MakeProcTriggerAura(core.ProcTrigger{
		Name:         config.Name + " Proc",
		Callback:     core.CallbackOnSpellHitDealt,
		Outcome:      core.OutcomeLanded,
		DPM:          character.NewDynamicLegacyProcForWeapon(config.ItemID, config.PPM, 0),
		IsWeaponProc: true,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			procAura.Activate(sim)
			if procAura.MaxStacks > 0 {
				procAura.SetStacks(sim, procAura.MaxStacks)
			}
		},
	})

	character.ItemSwap.RegisterProc(config.ItemID, aura)
}
