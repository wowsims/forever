package core

import (
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func applyRaceEffects(agent Agent) {
	character := agent.GetCharacter()

	switch character.Race {
	case proto.Race_RaceDwarf:
		// Earth Fury 1259719 holds while any weapon slot holds a mace, one- or two-handed, and follows
		// item swaps.
		earthFuryActive := false
		character.NewEquipGate(EquipRequirement{SubclassMask: WeaponSubclassMask(proto.WeaponType_WeaponTypeMace)}, func(_ *Simulation, level float64) {
			if (level > 0) == earthFuryActive {
				return
			}
			earthFuryActive = level > 0
			character.PseudoStats.CritDamageMultiplier *= Ternary(earthFuryActive, 1.02, 1/1.02)
		}).SetLevel(nil, 1)
		applyCreatureTypeSlaying(character, proto.MobType_MobTypeBeast)

		actionID := ActionID{SpellID: 20594}

		stoneFormAura := character.RegisterAura(Aura{
			Label:    "Stoneform",
			ActionID: actionID,
			Duration: time.Second * 8,
		}).AttachMultiplicativePseudoStatBuff(&character.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical], 0.9)

		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				DefaultCast: Cast{
					GCD: GCDDefault,
				},
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 3,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				stoneFormAura.Activate(sim)
			},

			RelatedSelfBuff: stoneFormAura,
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeSurvival,
		})
	case proto.Race_RaceGnome:
		applyExpansiveMind(character)
		applyEureka(character)
	case proto.Race_RaceHuman:
		character.MultiplyStat(stats.Spirit, 1.05)
		character.AddStats(stats.Stats{stats.PhysicalCritPercent: 2, stats.SpellCritPercent: 2}) // Ruthless 20597
	case proto.Race_RaceNightElf:
		character.PseudoStats.BaseDodgeChance += 0.01
		character.PseudoStats.MovementSpeedMultiplier *= 1.02

		RegisterTemporaryStatsOnUseCD(character,
			"Elune's Light",
			stats.Stats{
				stats.PhysicalCritPercent: 10,
				stats.SpellCritPercent:    10,
			},
			time.Second*15,
			SpellConfig{
				ActionID: ActionID{SpellID: 1259799},
				Cast: CastConfig{
					CD: Cooldown{
						Timer:    character.NewTimer(),
						Duration: time.Minute * 3,
					},
				},
			})
	case proto.Race_RaceOrc:
		character.AddStats(stats.Stats{stats.PhysicalCritPercent: 1, stats.SpellCritPercent: 1}) // Savage 20574

		bloodFuryID := ActionID{SpellID: 20572}

		bloodFuryAura := character.NewTemporaryStatMultiplierAura(Aura{
			Label:    "Blood Fury",
			ActionID: bloodFuryID,
			Duration: time.Second * 15,
		}, []StatMultiplier{
			{Stat: stats.AttackPower, Multiplier: 1.1},
			{Stat: stats.RangedAttackPower, Multiplier: 1.1},
			{Stat: stats.SpellDamage, Multiplier: 1.1},
			{Stat: stats.HealingPower, Multiplier: 1.1},
		})

		bloodFuryCD := Cooldown{
			Timer:    character.NewTimer(),
			Duration: time.Minute * 3,
		}
		bloodFuryAura.Icd = &bloodFuryCD

		bloodFury := character.RegisterSpell(SpellConfig{
			ActionID: bloodFuryID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: bloodFuryCD,
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				bloodFuryAura.Activate(sim)
			},

			RelatedSelfBuff: bloodFuryAura.Aura,
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell:    bloodFury,
			Type:     CooldownTypeDPS,
			BuffAura: bloodFuryAura,
		})

		shatterCurseID := ActionID{SpellID: 1299026}

		shatterCurseAura := character.RegisterAura(Aura{
			Label:    "Shatter Curse",
			ActionID: shatterCurseID,
			Duration: time.Second * 8,
		})
		for _, school := range []stats.SchoolIndex{
			stats.SchoolIndexArcane,
			stats.SchoolIndexFire,
			stats.SchoolIndexFrost,
			stats.SchoolIndexHoly,
			stats.SchoolIndexNature,
			stats.SchoolIndexShadow,
		} {
			shatterCurseAura.AttachMultiplicativePseudoStatBuff(&character.PseudoStats.SchoolDamageTakenMultiplier[school], 0.95)
		}

		shatterCurse := character.RegisterSpell(SpellConfig{
			ActionID: shatterCurseID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				DefaultCast: Cast{
					GCD: GCDDefault,
				},
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 3,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				shatterCurseAura.Activate(sim)
			},

			RelatedSelfBuff: shatterCurseAura,
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: shatterCurse,
			Type:  CooldownTypeSurvival,
		})
	case proto.Race_RaceTauren:
		character.MultiplyStat(stats.Health, 1.05)
		character.PseudoStats.CritDamageMultiplier *= 1.02 // Endurance 20550
	case proto.Race_RaceTroll:
		applyCreatureTypeSlaying(character, proto.MobType_MobTypeBeast)

		actionID := ActionID{SpellID: 20554}

		berserkingAura := character.RegisterAura(Aura{
			Label:    "Berserking",
			ActionID: actionID,
			Duration: time.Second * 10,
		}).
			AttachMultiplyAttackSpeed(1.1).
			AttachMultiplyCastSpeed(1.1)

		berserking := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 3,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				berserkingAura.Activate(sim)
			},

			RelatedSelfBuff: berserkingAura,
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: berserking,
			Type:  CooldownTypeDPS,
		})
	case proto.Race_RaceUndead:
		applyTouchOfTheGrave(character)
	case proto.Race_RaceHighOrderSkyborne:
		applySkyborneSharedRacials(character)

		energizedAura := character.RegisterAura(Aura{
			Label:    "Energized",
			ActionID: ActionID{SpellID: 1270842},
			Duration: time.Second * 30, // SpellMisc 1270842 DurationIndex 9, 30000 ms
			OnGain: func(aura *Aura, sim *Simulation) {
				if aura.Unit.HasManaBar() {
					aura.Unit.MultiplyManaRegenSpeed(sim, 2)
				}
			},
			OnExpire: func(aura *Aura, sim *Simulation) {
				if aura.Unit.HasManaBar() {
					aura.Unit.MultiplyManaRegenSpeed(sim, 0.5)
				}
			},
		})

		character.RegisterSpell(SpellConfig{
			ActionID: ActionID{SpellID: 1259705},
			Flags:    SpellFlagAPL | SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				DefaultCast: Cast{
					GCD:      GCDDefault,
					CastTime: DurationFromMillis(2000.0),
				},
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				energizedAura.Activate(sim)
			},

			RelatedSelfBuff: energizedAura,
		})
	case proto.Race_RaceWindshaperSkyborne:
		applySkyborneSharedRacials(character)

		elementalBlessingAura := character.RegisterAura(Aura{
			Label:    "Elemental Blessing",
			ActionID: ActionID{SpellID: 1259688},
			Duration: time.Second * 30,
			OnGain: func(aura *Aura, sim *Simulation) {
				aura.Unit.MultiplyMovementSpeed(sim, 1.1)
			},
			OnExpire: func(aura *Aura, sim *Simulation) {
				aura.Unit.MultiplyMovementSpeed(sim, 1/1.1)
			},
		})

		character.RegisterSpell(SpellConfig{
			ActionID: ActionID{SpellID: 1259686},
			Flags:    SpellFlagAPL | SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				DefaultCast: Cast{
					GCD:      GCDDefault,
					CastTime: DurationFromMillis(500.0),
				},
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				elementalBlessingAura.Activate(sim)
			},

			RelatedSelfBuff: elementalBlessingAura,
		})
	}
}

func applySkyborneSharedRacials(character *Character) {
	applyCreatureTypeSlaying(character, proto.MobType_MobTypeElemental)

	character.PseudoStats.AttackSpeedMultiplier *= 1.03 // Wind Blessed 1259710
	character.PseudoStats.CastSpeedMultiplier *= 1.03
}

// Beast Slaying 20557, Big Game Hunter 1259721 and Elemental Insight 1259707 state one effect, aura 168:
// 5% damage against the creature type. Their rows carry no aura 547, the critical damage half that
// Humanoid Slaying 19151 and Monster Slaying 24294 state.
func applyCreatureTypeSlaying(character *Character, mobType proto.MobType) {
	character.Env.RegisterPostFinalizeEffect(func() {
		character.MultiplyDamageVersus(1.05, mobType)
	})
}

func applyExpansiveMind(character *Character) {
	switch character.Class {
	case proto.Class_ClassWarrior:
		character.rageBar.maxRage *= 1.05
	case proto.Class_ClassRogue:
		character.energyBar.maxEnergy *= 1.05
	case proto.Class_ClassPriest, proto.Class_ClassMage, proto.Class_ClassWarlock:
		character.MultiplyStat(stats.Mana, 1.05)
	}
}

// One class's Eureka! row: the class mask and the fraction of its idx0 SPELLMOD_COST, idx1
// SPELLMOD_DAMAGE and idx2 SPELLMOD_DOT, zero where the row states no SPELLMOD_DOT, and the row's
// duration and cooldown.
type EurekaMods struct {
	CostFlags, DamageFlags, ChannelFlags ClassFlags
	Cost, Damage, Channel                float64
	Duration, Cooldown                   time.Duration
}

var eurekaMods func(spellID int32) EurekaMods

// The rows sit in the spell store, which imports core, so sim/core/buffs registers their reader from
// its init.
func RegisterEurekaMods(read func(spellID int32) EurekaMods) {
	eurekaMods = read
}

// Eureka! scopes its mods to each class's abilities by the game files' class masks, for as long as
// the buff is up: idx0 of each row names the spells whose cost it lowers, idx1 the ones whose direct
// damage it raises, and idx2, on the warlock and priest rows only, the channels whose ticks it raises.
// Other periodic damage takes no bonus. A tick takes the bonus when Eureka! is up as it lands, so a
// channel started before the buff gains it and one running past its end loses it.
// TODO: Ingame test needed: whether a channel still running when the buff runs out keeps the bonus to
// its end.
func applyEureka(character *Character) {
	var spellID int32
	var resourceType proto.ResourceType
	procMask := ProcMaskSpecial

	switch character.Class {
	case proto.Class_ClassRogue:
		spellID, resourceType = 1259812, proto.ResourceType_ResourceTypeEnergy
	case proto.Class_ClassWarrior:
		spellID, resourceType = 1259813, proto.ResourceType_ResourceTypeRage
	case proto.Class_ClassMage:
		spellID, resourceType = 1259817, proto.ResourceType_ResourceTypeMana
	case proto.Class_ClassWarlock:
		spellID, resourceType = 1259821, proto.ResourceType_ResourceTypeMana
	case proto.Class_ClassPriest:
		spellID, resourceType = 1259823, proto.ResourceType_ResourceTypeMana
		procMask |= ProcMaskSpellHealing
	default:
		return
	}

	if eurekaMods == nil {
		panic(`no Eureka! rows are registered: import _ "github.com/wowsims/forever/sim/core/buffs"`)
	}
	mods := eurekaMods(spellID)
	actionID := ActionID{SpellID: spellID}

	aura := character.RegisterAura(Aura{
		Label:    "Eureka!",
		ActionID: actionID,
		Duration: mods.Duration,
	}).AttachSpellMod(SpellModConfig{
		Kind:         SpellMod_PowerCost_Pct,
		ClassFlags:   mods.CostFlags,
		ProcMask:     procMask,
		ResourceType: resourceType,
		FloatValue:   mods.Cost,
	}).AttachSpellMod(SpellModConfig{
		Kind:       SpellMod_DirectDamageDone_Pct,
		ClassFlags: mods.DamageFlags,
		ProcMask:   procMask,
		FloatValue: mods.Damage,
	})
	// A SPELLMOD_DOT row that names no spells reaches none, where a mod with no class flags reaches all.
	if !mods.ChannelFlags.IsZero() {
		aura.AttachSpellMod(SpellModConfig{
			Kind:       SpellMod_DotDamageDone_Pct,
			ClassFlags: mods.ChannelFlags,
			ProcMask:   procMask,
			SpellFlag:  SpellFlagChanneled,
			FloatValue: mods.Channel,
		})
	}

	spell := character.RegisterSpell(SpellConfig{
		ActionID: actionID,
		Flags:    SpellFlagNoOnCastComplete,
		Cast: CastConfig{
			CD: Cooldown{
				Timer:    character.NewTimer(),
				Duration: mods.Cooldown,
			},
		},
		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			aura.Activate(sim)
		},

		RelatedSelfBuff: aura,
	})

	character.AddMajorCooldown(MajorCooldown{
		Spell: spell,
		Type:  CooldownTypeDPS,
	})
}

func applyTouchOfTheGrave(character *Character) {
	auraID, procChance := int32(1260201), 0.3
	switch character.Class {
	case proto.Class_ClassWarrior, proto.Class_ClassPaladin, proto.Class_ClassRogue:
		auraID, procChance = 1260189, 0.2
	}

	drainID := ActionID{SpellID: 1260198}
	healthMetrics := character.NewHealthMetrics(drainID)

	drain := character.RegisterSpell(SpellConfig{
		ActionID:    drainID,
		SpellSchool: SpellSchoolShadow,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskEmpty,
		Flags:       SpellFlagProc | SpellFlagPassiveSpell | SpellFlagIgnoreModifiers | SpellFlagNoSpellMods,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			result := spell.CalcAndDealDamage(sim, target, 0.05*spell.Unit.MaxHealth(), spell.OutcomeMagicHit)
			if result.Landed() && spell.Unit.HasHealthBar() {
				spell.Unit.GainHealth(sim, result.Damage, healthMetrics)
			}
		},
	})

	character.MakeProcTriggerAura(ProcTrigger{
		Name:       "Touch of the Grave",
		ActionID:   ActionID{SpellID: auraID},
		Callback:   CallbackOnSpellHitDealt,
		ProcMask:   ProcMaskMelee | ProcMaskRanged | ProcMaskSpellDamage,
		Outcome:    OutcomeLanded,
		ProcChance: procChance,
		ICD:        time.Second * 25,

		RequireDamageComponent: true,

		Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
			drain.Cast(sim, result.Target)
		},
	})
}
