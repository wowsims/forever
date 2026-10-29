package forever

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Revelation's buff per class. Warriors and rogues have none.
var revelationBuffs = map[proto.Class]int32{
	proto.Class_ClassMage:    1248808,
	proto.Class_ClassPriest:  1323377,
	proto.Class_ClassWarlock: 1323392,
	proto.Class_ClassDruid:   1323400,
	proto.Class_ClassHunter:  1323410,
	proto.Class_ClassShaman:  1323418,
	proto.Class_ClassPaladin: 1323419,
}

func init() {

	// Felsteel Shield Spike
	// EffectID: 2714, Proc SpellID: 29455
	// Permanently attaches a felsteel spike to your shield that deals 26 to 38 damage to attackers whose melee attacks you block.
	// https://www.wowhead.com/forever/spell=29455
	shared.NewProcDamageEffect(shared.ProcDamageEffect{
		EnchantID: 2714,
		SpellID:   29455,
		School:    core.SpellSchoolPhysical,
		MinDmg:    26,
		MaxDmg:    38,
		IsMelee:   true,
		Flags:     core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagNoOnDamageDealt,
		Trigger: core.ProcTrigger{
			Name:       "Felsteel Shield Spike",
			Callback:   core.CallbackOnSpellHitTaken,
			ProcMask:   core.ProcMaskMelee,
			Outcome:    core.OutcomeBlock,
			ProcChance: 1,
		},
	})

	// Mongoose
	// EffectID: 2673, Proc SpellID: 28093
	// PPM: 1, ICD: 0
	// Permanently enchant a Melee Weapon to occasionally increase Agility by 120 and attack speed slightly (2%).
	core.NewEnchantEffect(2673, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 15

		createMongooseAuras := func(slot proto.ItemSlot) *core.StatBuffAura {
			labelSuffix := core.Ternary(slot == proto.ItemSlot_ItemSlotMainHand, " (MH)", " (OH)")
			aura := character.NewTemporaryStatsAuraWrapped(
				"Lightning Speed"+labelSuffix,
				core.ActionID{SpellID: 28093}.WithHand(slot),
				stats.Stats{stats.Agility: 120},
				duration,
				func(aura *core.Aura) {
					aura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
						character.MultiplyAttackSpeed(sim, 1.02)
					})
					aura.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
						character.MultiplyAttackSpeed(sim, 1/1.02)
					})
				},
			)
			character.AddStatProcBuff(2673, aura, true, []proto.ItemSlot{slot})
			character.ItemSwap.RegisterWeaponEnchantBuff(aura.Aura, 2673)
			return aura
		}

		mhAuras := createMongooseAuras(proto.ItemSlot_ItemSlotMainHand)
		ohAuras := createMongooseAuras(proto.ItemSlot_ItemSlotOffHand)

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:         "Enchant Weapon - Mongoose",
			Callback:     core.CallbackOnSpellHitDealt,
			ActionID:     core.ActionID{SpellID: 28093},
			IsWeaponProc: true,
			DPM:          character.NewDynamicLegacyProcForEnchant(2673, 1.0, 0),
			Outcome:      core.OutcomeLanded,
			Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				core.Ternary(spell.IsOH(), ohAuras, mhAuras).Activate(sim)
			},
		})
	})

	// Executioner
	// EffectID: 3225, Proc SpellID: 42976
	// PPM: ?, ICD: 0
	// Permanently enchant a Melee Weapon to occasionally ignore 840 of your enemy's armor.  Requires a level 60 or higher item.
	core.NewEnchantEffect(3225, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 15

		aura := character.NewTemporaryStatsAura(
			"Executioner",
			core.ActionID{SpellID: 42976},
			stats.Stats{stats.ArmorPenetration: 840},
			duration,
		)
		character.AddStatProcBuff(3225, aura, true, core.AllMeleeWeaponSlots())
		character.ItemSwap.RegisterWeaponEnchantBuff(aura.Aura, 3225)

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:         "Enchant Weapon - Executioner",
			Callback:     core.CallbackOnSpellHitDealt,
			ActionID:     core.ActionID{SpellID: 28093},
			IsWeaponProc: true,
			DPM:          character.NewDynamicLegacyProcForEnchant(3225, 1.0, 0),
			Outcome:      core.OutcomeLanded,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})
	})

	// Deathfrost
	// EffectID: 3273, Proc SpellID: 46579, Damage SpellID: 46579, Debuff SpellID: 46629
	// Proc Chance: 50%, ICD: 25s
	// Permanently enchant a weapon so your damaging spells and melee weapon hits occasionally inflict an additional 150 Frost damage
	// and reduce the target's melee, ranged, and casting speed by 15% for 8 sec.  Requires a level 60 or higher item.
	core.NewEnchantEffect(3273, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 8
		effect := 0.85

		debuffArray := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return target.GetOrRegisterAura(core.Aura{
				Label:    "Deathfrost",
				Duration: duration,
				ActionID: core.ActionID{SpellID: 46629},
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					aura.Unit.MultiplyAttackSpeed(sim, effect)
					aura.Unit.MultiplyRangedSpeed(sim, effect)
					aura.Unit.MultiplyCastSpeed(sim, effect)
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					aura.Unit.MultiplyAttackSpeed(sim, 1/effect)
					aura.Unit.MultiplyRangedSpeed(sim, 1/effect)
					aura.Unit.MultiplyCastSpeed(sim, 1/effect)
				},
			})
		})

		dfSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 46579},
			SpellSchool: core.SpellSchoolFrost,
			DefenseType: core.DefenseTypeMagic,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagProc,
			ProcMask:    core.ProcMaskSpellDamage,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, 150, spell.OutcomeMagicCrit)
			},
		})

		// Enchant 3273 carries two effects: a combat spell (Effect 1) that rolls on melee hits, and an
		// Equip aura 46662 (Effect 3) that rolls on spell damage, direct and periodic (ProcTypeMask
		// 2424832), without Can Proc From Procs. Both cast 46579. The 25 s lockout is 46662's proc
		// recovery; the combat half has none in the DBC, so the two halves share one here rather than
		// doubling the rate the old single trigger had.
		//
		// The roll and the lockout are taken in ExtraCondition, which runs synchronously in the
		// callback. The handler runs a batch window later, and hits landing on the same timestamp
		// (both hands, extra attacks, multiple targets) would all pass a lockout consumed there.
		lockout := core.Cooldown{
			Timer:    character.NewTimer(),
			Duration: time.Second * 25,
		}
		meleeDPM := character.NewDynamicLegacyProcForEnchant(3273, 0, 0.5)
		handler := func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			debuffArray.Get(result.Target).Activate(sim)
			dfSpell.Cast(sim, result.Target)
		}

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:         "Enchant Weapon - Deathfrost (Melee)",
			Callback:     core.CallbackOnSpellHitDealt,
			ActionID:     core.ActionID{SpellID: 46579},
			IsWeaponProc: true,
			Outcome:      core.OutcomeLanded,
			ExtraCondition: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
				if !lockout.IsReady(sim) || !meleeDPM.Proc(sim, spell.ProcMask, "Deathfrost (Melee)") {
					return false
				}
				lockout.Use(sim)
				return true
			},
			Handler: handler,
		})

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Enchant Weapon - Deathfrost (Spell)",
			Callback: core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt,
			ActionID: core.ActionID{SpellID: 46662},
			ProcMask: core.ProcMaskSpellDamage,
			Outcome:  core.OutcomeLanded,

			RequireDamageComponent: true,

			ExtraCondition: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
				if !lockout.IsReady(sim) || !sim.Proc(0.5, "Deathfrost (Spell)") {
					return false
				}
				lockout.Use(sim)
				return true
			},
			Handler: handler,
		})
	})

	// Enchant Weapon - Revelation
	// EffectID: 8217, Proc SpellID: 1248806
	// The proc spell names no buff: each class with spells to crit has a Revelation of its own, and the
	// wearer gets its class's, with that row's class flags and charge. The game files state no rate,
	// only that a direct spell or heal landing without a crit rolls it, and that the chance falls as
	// the wearer's crit chance rises; the enchant's effect entry carries the chance. Only a hit that deals damage or heals rolls it: a landing for
	// nothing, such as a channel's start or an aura's application, does not. The spell a channel
	// casts each tick (Blizzard's) rolls it too.
	// TODO: Verify in game that warriors and rogues get no buff.
	core.NewEnchantEffect(8217, func(agent core.Agent) {
		character := agent.GetCharacter()
		buffID, ok := revelationBuffs[character.Class]
		if !ok {
			return
		}

		row := spelldata.MustFind(buffID)
		buff := character.RegisterAura(spelldata.AuraConfig(row))
		spelldata.ParseEffects(character, buff, row)
		character.ItemSwap.RegisterWeaponEnchantBuff(buff, 8217)

		// The charge goes to the next cast the crit modifier reaches, whatever its proc mask. A channel
		// the mask names (Blizzard, a hunter's Volley) spends it.
		spender := spelldata.AttachChargeSpender(character, buff, row, "Revelation Charge",
			func(_ *core.Character, trigger *core.ProcTrigger) {
				trigger.ProcMask = core.ProcMaskUnknown
			})

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1248806), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Activate(sim)
			buff.SetStacks(sim, buff.MaxStacks)
		}, spelldata.Chance(enchantProcChance(8217, 1248806))) // TODO: Model how the chance falls with the wearer's crit chance.
		trigger.Name = "Enchant Weapon - Revelation"
		trigger.Outcome = core.OutcomeLanded
		trigger.RequireDamageDealt = true
		// The row's listener hears direct spells and heals. A ranged Shot the buff's class flags name
		// rolls the same chance, which the hunter buff's Arcane Shot is.
		// TODO: Verify in game that a hunter's Arcane Shot procs Revelation.
		spellHits, classFlags := trigger.ProcMask, spender.ClassFlags
		trigger.ProcMask = core.ProcMaskUnknown
		trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			shot := spell.DefenseType == core.DefenseTypeRanged && spell.MatchesFlags(classFlags)
			if !spell.ProcMask.Matches(spellHits) && !shot {
				return false
			}
			// A landing while the buff is up rolls too and refreshes it.
			return !result.DidCrit()
		}
		character.ItemSwap.RegisterEnchantProc(8217, character.MakeProcTriggerAura(trigger))
	})

	// Scopes
	core.NewEnchantEffect(2722, func(agent core.Agent) {
		character := agent.GetCharacter()
		ranged := character.AutoAttacks.Ranged()
		ranged.BaseDamageMin += 10
		ranged.BaseDamageMax += 10
	})

	core.NewEnchantEffect(2723, func(agent core.Agent) {
		character := agent.GetCharacter()
		ranged := character.AutoAttacks.Ranged()
		ranged.BaseDamageMin += 12
		ranged.BaseDamageMax += 12
	})

	movementSpeedEnchants := []int32{
		911,  // Enchant Boots - Minor Speed
		2939, // Enchant Boots - Cat's Swiftness
		2940, // Enchant Boots - Boar's Speed
	}

	for _, enchantID := range movementSpeedEnchants {
		core.NewEnchantEffect(enchantID, func(agent core.Agent) {
			character := agent.GetCharacter()
			// The game files state the run speed as A_MOD_INCREASE_SPEED, as Strider Kick and Dash do,
			// and only the highest of those counts, so it bids in the active category with them.
			aura := core.MakePermanent(character.GetOrRegisterAura(core.Aura{
				Label:    "Minor Run Speed",
				ActionID: core.ActionID{SpellID: 13889},
			}))
			aura.NewActiveMovementSpeedEffect(0.08)

			character.ItemSwap.RegisterEnchantProc(enchantID, aura)
		})
	}
}

// The chance per hit the enchant's effect entry carries for the spell.
func enchantProcChance(effectID int32, spellID int32) float64 {
	for _, effect := range core.GetEnchantByEffectID(effectID).EnchantEffects {
		if effect.BuffId == spellID && effect.GetProc().GetProcChance() > 0 {
			return effect.GetProc().GetProcChance()
		}
	}
	panic(fmt.Sprintf("enchant %d states no proc chance for spell %d", effectID, spellID))
}
