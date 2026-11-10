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

		// The row's ProcTypeMask hears melee and ranged hits taken; the block is the tooltip's, and a
		// block that leaves no damage still fires it.
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
	// 1293783 casts 1293782 on every hit, and the rows deal its Nature damage to any target. The
	// tooltip limits it to Swine, so the trigger fires only on a hit against one. The tooltip reads
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

	// Thick Obsidian Breastplate
	// https://www.wowhead.com/forever/item=22196/thick-obsidian-breastplate
	//
	// When struck by a non-periodic damage spell you have a 30% chance of getting a 6s spell shield
	// that absorbs 400 of that school of damage.
	//
	// 27539 is a scripted dummy carrying the chance, the 10 s internal cooldown and the spell hits
	// taken it hears. The shield is one row per school; only the Fire one, 27533, is in the spell
	// store, and the other five state the same absorb, roll and duration on their own school.
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
		row := spelldata.MustFind(16608)
		leech := row.FirstAura(dbcenums.A_PERIODIC_LEECH)
		healthMetrics := character.NewHealthMetrics(core.ActionID{SpellID: row.ID})

		config := procSpellConfig(character, row)
		config.Dot = spelldata.DotConfig(row, leech, spelldata.Label("Demon Forged Breastplate"))
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
		drain := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(16611), func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			drain.Cast(sim, spell.Unit)
		})
		trigger.Name = "Demon Forged Breastplate"
		trigger.ActionID = core.ActionID{ItemID: 12628}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(12628, character.MakeProcTriggerAura(trigger))
	})

	// Stoneform Shoulders
	// https://www.wowhead.com/forever/item=13955/stoneform-shoulders
	//
	// After entering combat, reduce the next instance of Physical damage taken within 20s by 60.
	//
	// 1298230 applies 1298231 on entering combat, which is the encounter start in the sim. 1298231
	// states the 60 as flat Physical damage taken and no charges: the first Physical hit taken that
	// deals damage, of the direct hits its ProcTypeMask hears, takes the 60 off and removes it.
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
}

// TODO: The game files state no rate for Depleted Eye of Influence's 1297085, and none is measured on
// Forever.
// const depletedEyeOfInfluenceChance = 0 // TODO: waiting for PPM data

// The spell an item proc casts, as its row states it: out of the rotation, no cast or cost of its
// own, and a proc unless the row says it is not.
func procSpellConfig(character *core.Character, row *spelldata.Spell) core.SpellConfig {
	config := spelldata.SpellConfig(&character.Unit, row, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc())
	if row.IsAProc() {
		config.Flags |= core.SpellFlagProc
	}
	return config
}
