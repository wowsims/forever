package forever

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	// Jom Gabbar
	// Use: Increases attack power by 65 and an additional 65 every 2 sec. Lasts 20 sec. (2 Min Cooldown)
	core.NewItemEffect(23570, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionID := core.ActionID{SpellID: 29602}
		duration := time.Second * 20
		bonusPerStack := stats.Stats{
			stats.AttackPower:       65,
			stats.RangedAttackPower: 65,
		}

		jomGabbarAura := character.GetOrRegisterAura(core.Aura{
			Label:     "Jom Gabbar",
			ActionID:  actionID,
			Duration:  duration,
			MaxStacks: 10,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period:          time.Second * 2,
					NumTicks:        10,
					Priority:        core.ActionPriorityAuto,
					TickImmediately: true,
					OnAction: func(sim *core.Simulation) {
						aura.AddStack(sim)
					},
				})
			},
			OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
				character.AddStatsDynamic(sim, bonusPerStack.Multiply(float64(newStacks-oldStacks)))
			},
		})

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID: actionID,
			Flags:    core.SpellFlagNoOnCastComplete,
			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
				SharedCD: core.Cooldown{
					Timer:    character.GetOffensiveTrinketCD(),
					Duration: duration,
				},
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				jomGabbarAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeDPS,
		})
	})

	// Mark of the Champion (physical): +150 AP vs Undead and Demons.
	core.NewItemEffect(23206, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label:    "Mark of the Champion (Physical)",
			ActionID: core.ActionID{ItemID: 23206},
		})).AttachMobTypeBonusStats(stats.Stats{stats.AttackPower: 150, stats.RangedAttackPower: 150},
			proto.MobType_MobTypeUndead, proto.MobType_MobTypeDemon)
		character.ItemSwap.RegisterProc(23206, aura)
	})

	// Mark of the Champion (spell): +85 spell damage vs Undead and Demons.
	core.NewItemEffect(23207, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label:    "Mark of the Champion (Spell)",
			ActionID: core.ActionID{ItemID: 23207},
		})).AttachMobTypeBonusStats(stats.Stats{stats.SpellDamage: 85}, proto.MobType_MobTypeUndead, proto.MobType_MobTypeDemon)
		character.ItemSwap.RegisterProc(23207, aura)
	})

	// Lifegiving Gem: Gift of Life 23725 raises maximum health by 15% for 20 s (23782) and heals for as
	// much (23783). Both are scripted: the dummy states neither, and 23782 states 0 for its
	// A_MOD_INCREASE_HEALTH and is not in the spell store, so the 15% is the tooltip's and the 20 s is
	// 23782's duration. When the raise fades, only health above the lower maximum goes.
	core.NewItemEffect(19341, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionID := core.ActionID{SpellID: 23725}
		healthMetrics := character.NewHealthMetrics(core.ActionID{SpellID: 23783})

		var bonusHealth float64
		aura := character.RegisterAura(core.Aura{
			Label:    "Gift of Life",
			ActionID: actionID,
			Duration: time.Second * 20,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				bonusHealth = character.MaxHealth() * 0.15
				character.AddStatsDynamic(sim, stats.Stats{stats.Health: bonusHealth})
				character.GainHealth(sim, bonusHealth, healthMetrics)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				character.AddStatsDynamic(sim, stats.Stats{stats.Health: -bonusHealth})
				if character.CurrentHealth() > character.MaxHealth() {
					// Healing caps health at the maximum, so a heal of 0 trims the excess without
					// counting it as damage taken.
					character.GainHealth(sim, 0, healthMetrics)
				}
			},
		})

		buff := &core.StatBuffAura{Aura: aura, BuffedStatTypes: []stats.Stat{stats.Health}}
		registerItemOnUse(character, 19341, core.CooldownTypeSurvival, buff, func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			aura.Activate(sim)
		})
	})

	// The four buffs below go on at full strength and lose a stack each time their trigger fires. Each
	// takes its trigger and duration from the on-use row; the stacks are a second, scripted row that
	// is not in the spell store, so the count and the amount a stack carries are stated here with
	// the row they come from.

	// Zandalarian Hero Badge: 2000 armor and 30 defense, less 200 armor and 3 defense for each melee
	// or ranged hit taken. Brittle Armor 24575: 200 armor and 3 Defense a stack, 10 stacks.
	core.NewItemEffect(19948, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(24574)
		perStack := stats.Stats{stats.Armor: 200, stats.DefenseRating: 3 * core.DefenseRatingPerDefenseLevel}
		registerCountdownOnUse(character, 19948, row, 10, perStack, core.CooldownTypeSurvival, spelldata.ProcTrigger(character, row, nil))
	})

	// Zandalarian Hero Medallion: 40 melee and ranged damage, less 2 for each hit dealt. Restless
	// Strength 24662: 2 physical damage a stack, 20 stacks.
	core.NewItemEffect(19949, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(24661)
		perStack := stats.Stats{stats.PhysicalDamage: 2}
		registerCountdownOnUse(character, 19949, row, 20, perStack, core.CooldownTypeDPS, spelldata.ProcTrigger(character, row, nil))
	})

	// Zandalarian Hero Charm: 204 spell damage and 408 healing, less 17 and 34 for each spell cast.
	// Unstable Power 24659: 17 spell damage and 34 healing a stack, 12 stacks. The on-use row states
	// no trigger; the stacks' row states its ProcTypeMask, whose melee and ranged bits the tooltip's
	// "every time you cast a spell" leaves out, so the trigger is a damage or healing spell's cast.
	core.NewItemEffect(19950, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(24658)
		perStack := stats.Stats{stats.SpellDamage: 17, stats.HealingPower: 34}
		trigger := core.ProcTrigger{
			Name:     row.Name,
			Callback: core.CallbackOnCastComplete,
			ProcMask: core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
		}
		registerCountdownOnUse(character, 19950, row, 12, perStack, core.CooldownTypeDPS, trigger)
	})

	// Petrified Scarab: 100 to every resistance, less 10 for each hostile spell that lands. Mercurial
	// Shield 26464: 10 to the five resistances a stack, 10 stacks.
	core.NewItemEffect(21685, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(26463)
		perStack := stats.Stats{
			stats.ArcaneResistance: 10,
			stats.FireResistance:   10,
			stats.FrostResistance:  10,
			stats.NatureResistance: 10,
			stats.ShadowResistance: 10,
		}
		registerCountdownOnUse(character, 21685, row, 10, perStack, core.CooldownTypeSurvival, spelldata.ProcTrigger(character, row, nil))
	})

	// Everlook Delivery Bot: 1295271 marks the target, and once as many spells or attacks as its first
	// two effects bound have landed on it, 1295633 deals its Fire damage to it and the mark goes. The
	// count is rolled when the mark goes on, and the mark's ProcFlags state what counts. The tooltip
	// halves the count in snowy areas, which the row names no area for, so the sim does not.
	core.NewItemEffect(274760, func(agent core.Agent) {
		character := agent.GetCharacter()
		markRow := spelldata.MustFind(1295271)
		deliveryRow := spelldata.MustFind(1295633)
		fewest, most := int32(markRow.EffectN(1).BaseValue()), int32(markRow.EffectN(2).BaseValue())

		deliveryConfig := spelldata.SpellConfig(&character.Unit, deliveryRow, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc())
		if deliveryRow.IsAProc() {
			deliveryConfig.Flags |= core.SpellFlagProc
		}
		damage := deliveryRow.DamageEffect()
		cannotCrit := deliveryRow.CannotCrit()
		deliveryConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			outcome := spell.OutcomeMagicHitAndCrit
			if cannotCrit {
				outcome = spell.OutcomeMagicHit
			}
			result := spell.CalcDamage(sim, target, damage.DirectAmount(sim), outcome)
			spell.DealDamageAfterTravel(sim, result)
		}
		delivery := character.RegisterSpell(deliveryConfig)

		marks := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			var needed int32
			mark := target.RegisterAura(core.Aura{
				Label:     fmt.Sprintf("%s %s", markRow.Name, character.Label),
				ActionID:  core.ActionID{SpellID: markRow.ID},
				Duration:  markRow.Duration(),
				MaxStacks: int32(markRow.MaxStack),
				OnGain: func(_ *core.Aura, sim *core.Simulation) {
					needed = fewest + int32(sim.RandomFloat("Everlook Delivery Bot")*float64(most-fewest+1))
				},
			})
			mark.AttachProcTriggerCallback(target, spelldata.ProcTrigger(character, markRow, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				if !mark.IsActive() {
					return
				}
				mark.AddStack(sim)
				if mark.GetStacks() >= needed {
					mark.Deactivate(sim)
					delivery.Cast(sim, target)
				}
			}))
			return mark
		})

		registerItemOnUse(character, 274760, core.CooldownTypeDPS, nil, func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			marks.Get(target).Activate(sim)
		})
	})

	// Darkmoon Card: Heroism
	// https://www.wowhead.com/forever/item=19287/darkmoon-card-heroism
	//
	// Sometimes heals bearer of 120 to 180 damage when damaging an enemy in melee.
	//
	// The game files state no rate for 23689; the item's entry carries it.
	core.NewItemEffect(19287, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(23682)
		heal := row.HealEffect()

		config := spelldata.SpellConfig(&character.Unit, row, spelldata.Magic(core.ProcMaskSpellHealing), spelldata.Proc())
		if row.IsAProc() {
			config.Flags |= core.SpellFlagProc
		}
		// The row states no defense type; a heal crits for the magic multiplier.
		config.DefenseType = core.DefenseTypeMagic
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, heal.DirectAmount(sim), spell.OutcomeHealingCrit)
		}
		heroism := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(23689), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			heroism.Cast(sim, &character.Unit)
		}, spelldata.PPM(weaponProcPPM(19287, row.ID)))
		trigger.Name = "Darkmoon Card: Heroism"
		trigger.ActionID = core.ActionID{ItemID: 19287}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(19287, character.MakeProcTriggerAura(trigger))
	})

	// Highborne Research Tablet
	// https://www.wowhead.com/forever/item=260205/highborne-research-tablet
	//
	// Chance on hit to deal 69 to 81 Arcane damage. Deals 2 times as much damage to Naga and Satyrs.
	//
	// TODO: waiting for PPM data
	//
	// The 2 times is the trigger 1318159's own value; 1265634 states only the damage.
	// core.NewItemEffect(260205, func(agent core.Agent) {
	// 	character := agent.GetCharacter()
	// 	row := spelldata.MustFind(1265634)
	// 	triggerRow := spelldata.MustFind(1318159)
	// 	damage := row.DamageEffect()
	// 	familyMultiplier := triggerRow.Effects[0].BaseValue()
	//
	// 	config := procSpellConfig(character, row)
	// 	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		amount := damage.DirectAmount(sim) * creatureFamilyMultiplier(target, familyMultiplier,
	// 			proto.CreatureFamily_CreatureFamilyNaga, proto.CreatureFamily_CreatureFamilySatyr)
	// 		spell.CalcAndDealDamage(sim, target, amount, spell.OutcomeMagicHitAndCrit)
	// 	}
	// 	ancientPower := character.RegisterSpell(config)
	//
	// 	trigger := spelldata.ProcTrigger(character, triggerRow, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
	// 		ancientPower.Cast(sim, result.Target)
	// 	}, spelldata.Chance(highborneResearchTabletChance))
	// 	trigger.Name = "Highborne Research Tablet"
	// 	trigger.ActionID = core.ActionID{ItemID: 260205}
	// 	trigger.TriggerImmediately = true
	// 	character.ItemSwap.RegisterProc(260205, character.MakeProcTriggerAura(trigger))
	// })

	// Dormant Heart of the Mountain
	// https://www.wowhead.com/forever/item=249473/dormant-heart-of-the-mountain
	//
	// Your casts of Greater Heal in combat grant up to 40 increased healing and up to 13 increased
	// damage for 15s.
	//
	// The tooltip names the wearer's class's heal: Greater Heal, Healing Touch, Healing Wave or Holy
	// Light. 1249118's row names none of them, so the heals are matched by their ranks.
	core.NewItemEffect(249473, func(agent core.Agent) {
		character := agent.GetCharacter()
		buffRow := spelldata.MustFind(1249119)
		buff := character.RegisterAura(spelldata.AuraConfig(buffRow, spelldata.Label("Dormant Heart of the Mountain Proc")))
		parsed := spelldata.ParseEffects(character, buff, buffRow)

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1249118), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Activate(sim)
		})
		trigger.Name = "Dormant Heart of the Mountain"
		trigger.ActionID = core.ActionID{ItemID: 249473}
		trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return slices.ContainsFunc(eternalPowerHeals, func(heal spelldata.Ladder) bool { return heal.Contains(spell) })
		}
		character.ItemSwap.RegisterProc(249473, character.MakeProcTriggerAura(trigger))
		character.AddStatProcBuff(249473, &core.StatBuffAura{Aura: buff, BuffedStatTypes: parsed.Stats()}, false,
			character.ItemSwap.EligibleSlotsForItem(249473))
	})
}

// TODO: The game files state no rate for these procs, and none is measured on Forever.
// const highborneResearchTabletChance = 0 // TODO: waiting for PPM data

// The ranks of the four heals Dormant Heart of the Mountain answers.
var eternalPowerHeals = []spelldata.Ladder{
	spelldata.Ranked(2060, 10963, 10964, 10965, 25314),                                  // Greater Heal
	spelldata.Ranked(5185, 5186, 5187, 5188, 5189, 6778, 8903, 9758, 9888, 9889, 25297), // Healing Touch
	spelldata.Ranked(331, 332, 547, 913, 939, 959, 8005, 10395, 10396, 25357),           // Healing Wave
	spelldata.Ranked(635, 639, 647, 1026, 1042, 3472, 10328, 10329, 25292),              // Holy Light
}

// A buff that goes on with all its stacks, each worth perStack, for the row's duration, and loses one
// each time the trigger fires while it is up.
func registerCountdownOnUse(character *core.Character, itemID int32, row *spelldata.Spell, stacks int32, perStack stats.Stats,
	cdType core.CooldownType, trigger core.ProcTrigger) {
	buff := core.MakeStackingAura(character, core.StackingStatAura{
		Aura: core.Aura{
			Label:     row.Name,
			ActionID:  core.ActionID{SpellID: row.ID},
			Duration:  row.Duration(),
			MaxStacks: stacks,
		},
		BonusPerStack: perStack,
	})

	trigger.Handler = func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.RemoveStack(sim)
	}
	buff.AttachProcTriggerCallback(&character.Unit, trigger)

	registerItemOnUse(character, itemID, cdType, buff, func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		buff.Activate(sim)
		buff.SetStacks(sim, stacks)
	})
}

// The item's use as a major cooldown, on the cooldowns its on-use effect states. The use completes no
// cast, so a buff that counts casts does not count the use that applied it.
func registerItemOnUse(character *core.Character, itemID int32, cdType core.CooldownType, buff *core.StatBuffAura, apply core.ApplySpellResults) {
	effects := shared.ItemEffects(itemID)
	i := slices.IndexFunc(effects, func(effect *proto.ItemEffect) bool { return effect.GetOnUse() != nil })
	if i < 0 {
		panic(fmt.Sprintf("item %d has no on-use effect", itemID))
	}

	config := core.SpellConfig{
		ActionID:     core.ActionID{ItemID: itemID},
		Flags:        core.SpellFlagNoOnCastComplete,
		Cast:         shared.OnUseCast(character, effects[i]),
		ApplyEffects: apply,
	}
	if buff != nil {
		config.RelatedSelfBuff = buff.Aura
	}

	character.AddMajorCooldown(core.MajorCooldown{
		Spell:    character.RegisterSpell(config),
		Type:     cdType,
		BuffAura: buff,
	})
}
