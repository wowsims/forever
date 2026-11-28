package forever

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	// Force Reactive Disk
	// https://www.wowhead.com/forever/item=18168/force-reactive-disk
	//
	// When the shield blocks, it releases an electrical charge that damages all nearby enemies.
	core.NewItemEffect(18168, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(22600)
		damage := row.DamageEffect()

		config := procSpellConfig(character, row)
		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, result := range spell.CalcAoeDamage(sim, damage.Average()*sim.Encounter.AOECapMultiplier(), spell.OutcomeMagicHitAndCrit) {
				spell.DealCastResult(sim, result)
			}
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealBatchedAoeDamage(sim)
			})
		}
		charge := character.RegisterSpell(config)

		// The block is the tooltip's, and a block that leaves no damage still fires it.
		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(22618), func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			charge.Cast(sim, spell.Unit)
		})
		trigger.Name = "Force Reactive Disk"
		trigger.ActionID = core.ActionID{ItemID: 18168}
		trigger.Outcome = core.OutcomeBlock
		trigger.RequireDamageDealt = false
		character.ItemSwap.RegisterProc(18168, character.MakeProcTriggerAura(trigger))
	})

	// Black Grasp of the Destroyer
	// https://www.wowhead.com/forever/item=22194/black-grasp-of-the-destroyer
	//
	// On successful melee or ranged attack gain 8 mana and if possible drain 8 mana from the target.
	//
	// 27522 casts a scripted dummy; the gain is 29471's. The drain (27526) takes mana from a target
	// the sim gives none, so it is not simulated.
	core.NewItemEffect(22194, func(agent core.Agent) {
		character := agent.GetCharacter()
		if !character.HasManaBar() {
			return
		}
		gain := spelldata.MustFind(29471).EnergizeEffect()
		metrics := character.NewManaMetrics(core.ActionID{SpellID: 29471})

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(27522), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			character.AddMana(sim, gain.Average(), metrics)
		})
		trigger.Name = "Black Grasp of the Destroyer"
		trigger.ActionID = core.ActionID{ItemID: 22194}
		character.ItemSwap.RegisterProc(22194, character.MakeProcTriggerAura(trigger))
	})

	// Swine Fists
	// https://www.wowhead.com/forever/item=10760/swine-fists
	//
	// Spells and attacks against Swine deal 2 Nature damage.
	//
	// The rows deal the Nature damage to any target; the tooltip limits it to Swine. The tooltip reads
	// its amount from Cursed Murloc Eye's 1292675; the damage dealt is 1293782's own.
	core.NewItemEffect(10760, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(1293782)
		damage := row.DamageEffect()

		config := procSpellConfig(character, row)
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeMagicHitAndCrit)
		}
		swineFists := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1293783), func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			if result.Target.CreatureFamily == proto.CreatureFamily_CreatureFamilySwine {
				swineFists.Cast(sim, result.Target)
			}
		})
		trigger.Name = "Swine Fists"
		trigger.ActionID = core.ActionID{ItemID: 10760}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(10760, character.MakeProcTriggerAura(trigger))
	})

	// Gloves of Holy Might
	// https://www.wowhead.com/forever/item=867/gloves-of-holy-might
	//
	// Smite your target for 112 Holy damage. Damage is tripled against Undead.
	//
	// The row states no power coefficient, so tripling the amount triples the hit. Its defense type is
	// ranged, so it rolls on the ranged hit table.
	core.NewItemEffect(867, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(1325229)
		damage := row.DamageEffect()

		config := procSpellConfig(character, row)
		// The tooltip's tripling, which no row states, is an override on 1325229.
		config.VersusBonus = shared.ItemVersusBonus(character, 867, proto.VersusEffectKind_VersusEffectKindDamage, row)
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeRangedHitAndCrit)
		}
		smite := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1330855), func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			smite.Cast(sim, result.Target)
		})
		trigger.Name = "Gloves of Holy Might"
		trigger.ActionID = core.ActionID{ItemID: 867}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(867, character.MakeProcTriggerAura(trigger))
	})

	// Thick Obsidian Breastplate
	// https://www.wowhead.com/forever/item=22196/thick-obsidian-breastplate
	//
	// When struck by a non-periodic damage spell you have a 30% chance of getting a 6s spell shield
	// that absorbs 400 of that school of damage.
	//
	// The shield is one row per school; only the Fire one, 27533, is in the spell store, and the other
	// five state the same absorb, roll and duration on their own school.
	core.NewItemEffect(22196, func(agent core.Agent) {
		character := agent.GetCharacter()
		fire := spelldata.MustFind(27533)
		absorb := fire.AbsorbEffect()

		type schoolShield struct {
			school core.SpellSchool
			aura   *core.DamageAbsorptionAura
		}
		var amount float64
		var shields []schoolShield
		for _, row := range []struct {
			id     int32
			school core.SpellSchool
		}{
			{27533, core.SpellSchoolFire},
			{27534, core.SpellSchoolFrost},
			{27535, core.SpellSchoolShadow},
			{27536, core.SpellSchoolHoly},
			{27538, core.SpellSchoolNature},
			{27540, core.SpellSchoolArcane},
		} {
			school := row.school
			aura := character.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
				Aura: core.Aura{
					Label:    "Obsidian Armor " + school.String(),
					ActionID: core.ActionID{SpellID: row.id},
					Duration: fire.Duration(),
				},
				ShieldStrengthCalculator: func(_ *core.Unit) float64 {
					return amount
				},
				ShouldApplyToResult: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult, _ bool) bool {
					return spell.SpellSchool.Matches(school)
				},
			})
			shields = append(shields, schoolShield{school, aura})
		}
		shieldFor := func(spell *core.Spell) *core.DamageAbsorptionAura {
			for _, shield := range shields {
				if spell.SpellSchool.Matches(shield.school) {
					return shield.aura
				}
			}
			return nil
		}

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(27539), func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			amount = absorb.Roll(sim)
			shieldFor(spell).Activate(sim)
		})
		trigger.Name = "Thick Obsidian Breastplate"
		trigger.ActionID = core.ActionID{ItemID: 22196}
		// A Physical spell has no shield, and does not spend the chance or the cooldown.
		trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return shieldFor(spell) != nil
		}
		character.ItemSwap.RegisterProc(22196, character.MakeProcTriggerAura(trigger))
	})

	// Depleted Eye of Influence
	// https://www.wowhead.com/forever/item=275630/depleted-eye-of-influence
	//
	// Chance on harmful spell cast to reduce target's attack power by 60 for 30s.
	//
	// TODO: waiting for PPM data
	// core.NewItemEffect(275630, func(agent core.Agent) {
	// 	character := agent.GetCharacter()
	// 	row := spelldata.MustFind(1297082)
	// 	debuffs := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 		aura := target.RegisterAura(spelldata.AuraConfig(row, spelldata.Label("Depleted Eye of Influence")))
	// 		spelldata.ParseEffects(nil, aura, row)
	// 		return aura
	// 	})
	//
	// 	trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1297085), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
	// 		debuffs.Get(character.CurrentTarget).Activate(sim)
	// 	}, spelldata.Chance(depletedEyeOfInfluenceChance))
	// 	trigger.Name = "Depleted Eye of Influence"
	// 	trigger.ActionID = core.ActionID{ItemID: 275630}
	// 	trigger.TriggerImmediately = true
	// 	character.ItemSwap.RegisterProc(275630, character.MakeProcTriggerAura(trigger))
	// })

	// Demon Forged Breastplate
	// https://www.wowhead.com/forever/item=12628/demon-forged-breastplate
	//
	// When struck has a 3% chance of stealing 240 life from the attacker over 4s.
	core.NewItemEffect(12628, func(agent core.Agent) {
		character := agent.GetCharacter()
		drain := registerLeechDrain(character, spelldata.MustFind(16608), "Demon Forged Breastplate")

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(16611), func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			drain.Cast(sim, spell.Unit)
		})
		trigger.Name = "Demon Forged Breastplate"
		trigger.ActionID = core.ActionID{ItemID: 12628}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(12628, character.MakeProcTriggerAura(trigger))
	})

	// Skullflame Shield
	// https://www.wowhead.com/forever/item=1168/skullflame-shield
	//
	// When struck in melee combat, has a 3% chance of stealing 270 life from target enemy.
	// When struck in combat has a 2% chance of dealing 115 Fire damage to all targets around you.
	//
	// Both roll on each landed melee hit taken. The sim has no attacker facing, so Drain Life takes the
	// attacker's life only while the wearer is In Front of Target, an owner ruling; Flamestrike hits every
	// enemy around the wearer (EV-0227).
	core.NewItemEffect(1168, func(agent core.Agent) {
		character := agent.GetCharacter()

		drainRow := spelldata.MustFind(18817)
		leech := drainRow.DamageEffect()
		healthMetrics := character.NewHealthMetrics(core.ActionID{SpellID: drainRow.ID})
		drainConfig := procSpellConfig(character, drainRow)
		drainConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, leech.DirectAmount(sim), spell.OutcomeMagicHitAndCrit)
			shared.GainLeechHealth(sim, character, result.Damage, leech, healthMetrics)
		}
		drainLife := character.RegisterSpell(drainConfig)

		drainTrigger := spelldata.ProcTrigger(character, spelldata.MustFind(18815), func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			drainLife.Cast(sim, spell.Unit)
		})
		drainTrigger.Name = "Skullflame Shield - Drain Life"
		drainTrigger.ActionID = core.ActionID{ItemID: 1168}
		drainTrigger.TriggerImmediately = true
		drainTrigger.ExtraCondition = func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
			return character.PseudoStats.InFrontOfTarget
		}
		character.ItemSwap.RegisterProc(1168, character.MakeProcTriggerAura(drainTrigger))

		flameRow := spelldata.MustFind(18818)
		flame := flameRow.DamageEffect()
		flameConfig := procSpellConfig(character, flameRow)
		flameConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.CalcAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, func(sim *core.Simulation, _ *core.Spell) float64 {
				return flame.DirectAmount(sim) * sim.Encounter.AOECapMultiplier()
			})
			spell.DealBatchedAoeDamage(sim)
		}
		flamestrike := character.RegisterSpell(flameConfig)

		flameTrigger := spelldata.ProcTrigger(character, spelldata.MustFind(18816), func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			flamestrike.Cast(sim, spell.Unit)
		})
		flameTrigger.Name = "Skullflame Shield - Flamestrike"
		flameTrigger.ActionID = core.ActionID{ItemID: 1168}
		flameTrigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(1168, character.MakeProcTriggerAura(flameTrigger))
	})

	// Stoneform Shoulders
	// https://www.wowhead.com/forever/item=13955/stoneform-shoulders
	//
	// After entering combat, reduce the next instance of Physical damage taken within 20s by 60.
	//
	// 1298230 applies 1298231 on entering combat, which is the encounter start in the sim. 1298231
	// states no charges: the first Physical hit taken that deals damage takes the 60 off, half of the hit
	// at most, and removes it.
	core.NewItemEffect(13955, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(1298231)
		buff := character.RegisterAura(spelldata.AuraConfig(row))
		spelldata.ParseEffects(character, buff, row)

		spend := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Deactivate(sim)
		})
		spend.Name = "Stoneform Shoulders - Spend"
		spend.TriggerImmediately = true
		spend.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spell.SpellSchool.Matches(core.SpellSchoolPhysical)
		}
		buff.AttachProcTriggerCallback(&character.Unit, spend)

		equip := core.MakePermanent(character.RegisterAura(core.Aura{
			Label:    "Stoneform Shoulders - Enter Combat",
			ActionID: core.ActionID{ItemID: 13955},
			OnEncounterStart: func(_ *core.Aura, sim *core.Simulation) {
				buff.Activate(sim)
			},
		}))
		character.ItemSwap.RegisterProc(13955, equip)
	})

	// Zealot Plate
	// https://www.wowhead.com/forever/item=274651/zealot-plate
	//
	// When struck in combat, has a chance to energize you for 88 mana. This effect cannot occur more
	// than once every 15 sec.
	//
	// The item's effect names Zealotry 1294666, the mana gain, so the stat-bonus route cannot build it.
	// The row's ProcChance as an override sets it (EV-0115).
	shared.NewSpellDataEnergizeProc(shared.SpellDataProc{TriggerSpellID: 1294668},
		[]shared.ItemVariant{
			{ItemID: 274651, ItemName: "Zealot Plate"},
		})

	// Argent Defender
	// https://www.wowhead.com/forever/item=13243/argent-defender
	//
	// Has a chance when struck in combat to increase your chance to block by 18% for 10s. Undead that
	// strike you are 2 times as likely to activate this effect.
	//
	// The procs per minute and the versus bonus (EV-0126).
	shared.NewSpellDataAuraProc(shared.SpellDataProc{TriggerSpellID: 17350, BuffSpellID: 17351,
		Rate: func(character *core.Character, trigger *core.ProcTrigger) {
			row := spelldata.MustFind(17350)
			undead := shared.ItemVersusBonus(character, 13243, proto.VersusEffectKind_VersusEffectKindProcChance, row)
			trigger.ProcChance = 0
			trigger.DPM = character.NewAttackerSwingPPMManager(float64(row.RPPM), trigger.ProcMask, undead.Against)
		}},
		[]shared.ItemVariant{
			{ItemID: 13243, ItemName: "Argent Defender"},
		})
}

// TODO: The game files state no rate for Depleted Eye of Influence's 1297085, and none is measured on
// Forever.
// const depletedEyeOfInfluenceChance = 0 // TODO: waiting for PPM data

// The spell an item proc casts, as its row states it.
func procSpellConfig(character *core.Character, row *spelldata.Spell) core.SpellConfig {
	return procSpellConfigAs(character, row, spelldata.Magic(core.ProcMaskEmpty))
}

// The same for a spell of the kind given, such as spelldata.Melee.
func procSpellConfigAs(character *core.Character, row *spelldata.Spell, kind spelldata.SpellOpt) core.SpellConfig {
	config := spelldata.SpellConfig(&character.Unit, row, kind, spelldata.Proc(), spelldata.Triggered())
	return config
}

// A proc's A_PERIODIC_LEECH drain.
func registerLeechDrain(character *core.Character, row *spelldata.Spell, label string) *core.Spell {
	leech := row.FirstAura(dbcenums.A_PERIODIC_LEECH)
	healthMetrics := character.NewHealthMetrics(core.ActionID{SpellID: row.ID})
	config := procSpellConfig(character, row)
	config.Dot = spelldata.DotConfig(row, leech, spelldata.Label(label))
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		result := dot.Spell.CalcAndDealPeriodicDamage(sim, target, leech.Average(), row.TickOutcome(dot))
		shared.GainLeechHealth(sim, character, result.Damage, leech, healthMetrics)
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}
	return character.RegisterSpell(config)
}
