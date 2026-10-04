package warlock

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type WarlockPet struct {
	core.Pet

	AutoCastAbilities []*core.Spell
	MinMana           float64 // The minimum amount of energy needed to the AI casts a spell
	ManaIntRatio      float64
}

var petBaseStats = map[proto.WarlockOptions_Summon]*stats.Stats{
	proto.WarlockOptions_Imp: {
		stats.Mana:                2988,
		stats.Stamina:             101,
		stats.Strength:            153, //fix these later
		stats.Agility:             108, //fix these later
		stats.Intellect:           327,
		stats.Spirit:              263,
		stats.AttackPower:         135,
		stats.MP5:                 123,
		stats.PhysicalCritPercent: 5,
		stats.SpellCritPercent:    5,
	},
	proto.WarlockOptions_Voidwalker: {
		stats.Stamina:             280,
		stats.Strength:            153,
		stats.Agility:             108,
		stats.Intellect:           133,
		stats.Spirit:              122,
		stats.AttackPower:         286,
		stats.MP5:                 48,
		stats.PhysicalCritPercent: 5,
		stats.SpellCritPercent:    5,
	},
	proto.WarlockOptions_Succubus: {
		stats.Mana:                3862,
		stats.Stamina:             280,
		stats.Strength:            154,
		stats.Agility:             108,
		stats.Intellect:           133,
		stats.Spirit:              122,
		stats.AttackPower:         286,
		stats.MP5:                 48,
		stats.PhysicalCritPercent: 5,
		stats.SpellCritPercent:    5,
	},
	proto.WarlockOptions_Felhunter: {},
}

func (warlock *Warlock) SimplePetStatInheritanceWithScale() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		const resistScale = 0.4
		const baseStatScale = 0.3

		return stats.Stats{
			stats.Stamina:          ownerStats[stats.Stamina] * 0.3,
			stats.Intellect:        ownerStats[stats.Intellect] * 0.3,
			stats.Armor:            ownerStats[stats.Armor] * 0.35,
			stats.SpellPiercing:    ownerStats[stats.SpellPiercing], // not 100% on this one
			stats.SpellDamage:      max(ownerStats[stats.ShadowDamage], ownerStats[stats.FireDamage]) * 0.15,
			stats.AttackPower:      max(ownerStats[stats.ShadowDamage], ownerStats[stats.FireDamage]) * 0.57,
			stats.ArcaneResistance: ownerStats[stats.ArcaneResistance] * resistScale,
			stats.FireResistance:   ownerStats[stats.FireResistance] * resistScale,
			stats.FrostResistance:  ownerStats[stats.FrostResistance] * resistScale,
			stats.NatureResistance: ownerStats[stats.NatureResistance] * resistScale,
			stats.ShadowResistance: ownerStats[stats.ShadowResistance] * resistScale,
		}
	}
}

func AutoAttackConfig(min float64, max float64) *core.AutoAttackOptions {
	return &core.AutoAttackOptions{
		MainHand: core.Weapon{
			BaseDamageMin: float64(min),
			BaseDamageMax: float64(max),
			SwingSpeed:    2.0,
		},
		AutoSwingMelee: true,
	}
}

func (warlock *Warlock) makePet(
	summon proto.WarlockOptions_Summon,
	name string,
	enabledOnStart bool,
	baseStats stats.Stats,
	aaOptions *core.AutoAttackOptions,
	statInheritance core.PetStatInheritance,
	isGuardian bool,
) *WarlockPet {
	pet := &WarlockPet{
		Pet: core.NewPet(core.PetConfig{
			Name:            name,
			Owner:           &warlock.Character,
			BaseStats:       baseStats,
			StatInheritance: statInheritance,
			EnabledOnStart:  enabledOnStart,
			IsGuardian:      isGuardian,
		}),
	}

	pet.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[proto.Class_ClassPaladin])

	pet.AddStatDependency(stats.Intellect, stats.SpellCritPercent, core.CritPerIntMaxLevel[proto.Class_ClassPaladin])
	pet.AddStatDependency(stats.Intellect, stats.Mana, 15)

	// set pet class for proper scaling values
	if enabledOnStart {
		warlock.ActivePet = pet
		pet.OnPetEnable = func(sim *core.Simulation) {
			if warlock.Talents.DemonicKnowledge > 0 {
				warlock.DemonicKnowledgeAura.Activate(sim)
				warlock.updateDemonicKnowledge(sim)
			}
			if warlock.Talents.MasterDemonologist > 0 {
				if warlock.MasterDemonologistAura.IsActive() {
					warlock.MasterDemonologistAura.Deactivate(sim)
				}
				warlock.MasterDemonologistAura.Activate(sim)
			}
		}
		pet.OnPetDisable = func(sim *core.Simulation) {
			if warlock.Talents.DemonicKnowledge > 0 {
				warlock.updateDemonicKnowledge(sim)
				warlock.DemonicKnowledgeAura.Deactivate(sim)
			}
		}
		warlock.RegisterResetEffect(func(sim *core.Simulation) {
			warlock.ActivePet = pet
		})
	}

	warlock.setPetOptions(pet, aaOptions)

	if summon == proto.WarlockOptions_Imp {
		pet.SetSpiritManaRegen(0, 1.0/5)
	}

	return pet
}

func (warlock *Warlock) setPetOptions(petAgent core.PetAgent, aaOptions *core.AutoAttackOptions) {
	pet := petAgent.GetPet()
	if aaOptions != nil {
		pet.EnableAutoAttacks(petAgent, *aaOptions)
	}

	pet.EnableManaBar()
	warlock.AddPet(petAgent)
}

func (warlock *Warlock) registerPets() {
	warlock.Imp = warlock.registerImp()
	warlock.Succubus = warlock.registerSuccubus()
	warlock.Felhunter = warlock.registerFelHunter()
	warlock.Voidwalker = warlock.registerVoidWalker()
}

// TODO: To be implemented. Port the TBC Imp implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) registerImp() *WarlockPet {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// name := proto.WarlockOptions_Summon_name[int32(proto.WarlockOptions_Imp)]
	// enabledOnStart := proto.WarlockOptions_Imp == warlock.Options.Summon
	// return warlock.registerImpWithName(name, enabledOnStart, false)
}

// TODO: To be implemented. Port the TBC Imp With Name implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) registerImpWithName(name string, enabledOnStart bool, isGuardian bool) *WarlockPet {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// pet := warlock.RegisterPet(proto.WarlockOptions_Imp, 0, 0, name, enabledOnStart, isGuardian)
	// pet.registerFireboltSpell()
	// pet.MinMana = 145
	// return pet
}

// TODO: To be implemented. Port the TBC Fel Hunter implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) registerFelHunter() *WarlockPet {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// name := proto.WarlockOptions_Summon_name[int32(proto.WarlockOptions_Felhunter)]
	// enabledOnStart := proto.WarlockOptions_Felhunter == warlock.Options.Summon
	// return warlock.registerFelHunterWithName(name, enabledOnStart, false)
}

// TODO: To be implemented. Port the TBC Fel Hunter With Name implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) registerFelHunterWithName(name string, enabledOnStart bool, isGuardian bool) *WarlockPet {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// pet := warlock.RegisterPet(proto.WarlockOptions_Felhunter, 2, 3.5, name, enabledOnStart, isGuardian)
	// //add felhunter ability
	// pet.MinMana = 130
	// return pet
}

// TODO: To be implemented. Port the TBC Void Walker implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) registerVoidWalker() *WarlockPet {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// name := proto.WarlockOptions_Summon_name[int32(proto.WarlockOptions_Voidwalker)]
	// enabledOnStart := proto.WarlockOptions_Voidwalker == warlock.Options.Summon
	// return warlock.registerVoidWalkerWithName(name, enabledOnStart, false)
}

// TODO: To be implemented. Port the TBC Void Walker With Name implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) registerVoidWalkerWithName(name string, enabledOnStart bool, isGuardian bool) *WarlockPet {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// pet := warlock.RegisterPet(proto.WarlockOptions_Voidwalker, 2, 3.5, name, enabledOnStart, isGuardian)
	// pet.registerTormentSpell()
	// pet.MinMana = 120
	// return pet
}

// TODO: To be implemented. Port the TBC Succubus implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) registerSuccubus() *WarlockPet {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// name := proto.WarlockOptions_Summon_name[int32(proto.WarlockOptions_Succubus)]
	// enabledOnStart := proto.WarlockOptions_Succubus == warlock.Options.Summon
	// return warlock.registerSuccubusWithName(name, enabledOnStart, false)
}

// TODO: To be implemented. Port the TBC Succubus With Name implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) registerSuccubusWithName(name string, enabledOnStart bool, isGuardian bool) *WarlockPet {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// pet := warlock.RegisterPet(proto.WarlockOptions_Succubus, 173, 216, name, enabledOnStart, isGuardian)
	// pet.registerLashOfPainSpell()
	// pet.MinMana = 190
	// return pet
}

// TODO: To be implemented. Not verified against Forever.
func (warlock *Warlock) RegisterPet(
	t proto.WarlockOptions_Summon,
	min float64,
	max float64,
	name string,
	enabledOnStart bool,
	isGuardian bool,
) *WarlockPet {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// baseStats, ok := petBaseStats[t]
	// if !ok {
	// panic("Undefined base stats for pet")
	// }
	//
	// var attackOptions *core.AutoAttackOptions = nil
	// if t > 1 {
	// attackOptions = AutoAttackConfig(min, max)
	// }
	//
	// inheritance := warlock.SimplePetStatInheritanceWithScale()
	//
	// return warlock.makePet(t, name, enabledOnStart, *baseStats, attackOptions, inheritance, isGuardian)
}

func (pet *WarlockPet) GetPet() *core.Pet {
	return &pet.Pet
}

func (pet *WarlockPet) Reset(_ *core.Simulation) {
}

func (pet *WarlockPet) OnEncounterStart(_ *core.Simulation) {
}

func (pet *WarlockPet) ExecuteCustomRotation(sim *core.Simulation) {
	waitUntil := time.Duration(0)

	for _, spell := range pet.AutoCastAbilities {
		if spell.CanCast(sim, pet.CurrentTarget) && pet.CurrentMana() > pet.MinMana {
			spell.Cast(sim, pet.CurrentTarget)
			return
		}

		// calculate energy required
		cost := max(pet.MinMana, spell.Cost.GetCurrentCost())
		regen := pet.ManaRegenPerSecondWhileCasting()
		if regen > 0 {
			timeTillMana := max(0, (cost-pet.CurrentMana())/regen)
			waitUntil = min(waitUntil, time.Duration(float64(time.Second)*timeTillMana))
		}
	}

	// for now average the delay out to 100 ms so we don't need to roll random every time
	pet.WaitUntil(sim, sim.CurrentTime+waitUntil+time.Millisecond*100)
}

var petActionFireBolt = core.ActionID{SpellID: 27267}

// TODO: To be implemented. Port the TBC Firebolt Spell implementation below; not yet verified against the Forever game files.
func (pet *WarlockPet) registerFireboltSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// pet.AutoCastAbilities = append(pet.AutoCastAbilities, pet.RegisterSpell(core.SpellConfig{
	// 	ActionID:       petActionFireBolt,
	// 	SpellSchool:    core.SpellSchoolFire,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: WarlockSpellImpFireBolt,
	// 	MissileSpeed:   16,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: 145,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      core.GCDMin,
	// 			CastTime: time.Second * 2,
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	DefenseType:      core.DefenseTypeMagic,
	// 	ThreatMultiplier: 1,
	// 	BonusCoefficient: 0.571,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		dmgRoll := pet.CalcAndRollDamageRange(sim, 112, 127)
	// 		result := spell.CalcDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			spell.DealDamage(sim, result)
	// 		})
	// 	},
	// }))
}

var petActionLashOfPain = core.ActionID{SpellID: 27274}

// TODO: To be implemented. Port the TBC Lash Of Pain Spell implementation below; not yet verified against the Forever game files.
func (pet *WarlockPet) registerLashOfPainSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// pet.AutoCastAbilities = append(pet.AutoCastAbilities, pet.RegisterSpell(core.SpellConfig{
	// 	ActionID:       petActionLashOfPain,
	// 	SpellSchool:    core.SpellSchoolShadow,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: WarlockSpellSuccubusLashOfPain,
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: 190,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: core.GCDDefault,
	// 		},
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    pet.NewTimer(),
	// 			Duration: 12 * time.Second,
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	DefenseType:      core.DefenseTypeMagic,
	// 	ThreatMultiplier: 1,
	// 	BonusCoefficient: 0.429,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcDamage(sim, target, 123, spell.OutcomeMagicHitAndCrit)
	// 		spell.DealDamage(sim, result)
	// 	},
	// }))
	//
}

var petActionTorment = core.ActionID{SpellID: 27270}

// TODO: To be implemented. Port the TBC Torment Spell implementation below; not yet verified against the Forever game files.
func (pet *WarlockPet) registerTormentSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// pet.AutoCastAbilities = append(pet.AutoCastAbilities, pet.RegisterSpell(core.SpellConfig{
	// 	ActionID:       petActionTorment,
	// 	SpellSchool:    core.SpellSchoolShadow,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: WarlockSpellVoidwalkerTorment,
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: 130,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: core.GCDDefault,
	// 		},
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    pet.NewTimer(),
	// 			Duration: time.Second * 5,
	// 		},
	// 	},
	// 	DefenseType: core.DefenseTypeMagic,
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcDamage(sim, target, 1000, spell.OutcomeMagicHitAndCrit)
	// 		spell.DealDamage(sim, result)
	// 	},
	// }))
}
