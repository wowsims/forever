package forever

import (
	"slices"
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	// Graverobber's Shovel
	//
	// Equip 1319047: 12 attack power and 12 ranged attack power, which the item's own stats carry,
	// and effects 3 and 4, 12 of each against Undead.
	core.NewItemEffect(273817, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(1319047)
		aura := core.MakePermanent(character.RegisterAura(spelldata.AuraConfig(row)))
		spelldata.ParseEffects(character, aura, row, spelldata.Effects(3, 4))
		character.ItemSwap.RegisterProc(273817, aura)
	})

	// Ravager
	// https://www.wowhead.com/forever/item=7717/ravager
	//
	// Chance on hit: Attack all nearby enemies, dealing weapon damage plus 5 every 3 sec for 9 sec, and
	// reducing your own movement speed by 80%.
	//
	// The game files state no rate for Bladestorm; the item's chance-on-hit entry carries it. The row
	// is a channel: it cuts off a cast in progress, and for its 9 s the wearer swings no auto attack
	// and casts nothing. Each tick casts 9633, a main-hand strike at every enemy, which rolls the other
	// weapon procs but not Bladestorm.
	newEntryRateItemEffect(7717, 9632, func(character *core.Character, ppm float64) {
		strikeRow := spelldata.MustFind(9633)
		strikeConfig := spelldata.SpellConfig(&character.Unit, strikeRow, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc())
		if strikeRow.IsAProc() {
			strikeConfig.Flags |= core.SpellFlagProc
		}
		strikeConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := strikeRow.WeaponStrike(sim, spell, spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
			spell.CalcAoeDamage(sim, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			spell.DealBatchedAoeDamage(sim)
		}
		strike := character.RegisterSpell(strikeConfig)

		row := spelldata.MustFind(9632)
		config := procSpellConfig(character, row)
		config.Dot = spelldata.DotConfig(row, row.EffectN(1), spelldata.Label("Ravager Bladestorm"))
		config.Dot.SelfOnly = true
		config.Dot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
			strike.Cast(sim, character.CurrentTarget)
		}
		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			character.Interrupt(sim)
			channel := spell.AOEDot()
			channel.Apply(sim)
			character.AutoAttacks.PauseMeleeBy(sim, channel.Duration)
			character.AutoAttacks.DelayRangedUntil(sim, sim.CurrentTime+channel.Duration)
		}
		bladestorm := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			bladestorm.Cast(sim, result.Target)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(7717, ppm, 0)
		})
		trigger.Name = "Ravager"
		trigger.ActionID = core.ActionID{ItemID: 7717}
		trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spell != strike && !bladestorm.AOEDot().IsActive()
		}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(7717, character.MakeProcTriggerAura(trigger))
	})

	// Hand of Edward the Odd
	// https://www.wowhead.com/forever/item=2243/hand-of-edward-the-odd
	//
	// Chance on hit: Next spell cast within 4 sec will cast instantly.
	//
	// The game files state no rate for Focus; the proc runs at the one the item's chance-on-hit entry
	// states, and registers nothing while it states none. The row makes the cast instant with +100000%
	// cast speed, which the sim states as no cast time at all, so the global cooldown stays whole. The
	// next spell with a cast time spends the one charge.
	newEntryRateItemEffect(2243, 18803, func(character *core.Character, ppm float64) {
		row := spelldata.MustFind(18803)

		instant := character.AddDynamicMod(core.SpellModConfig{
			Kind:       core.SpellMod_CastTime_Pct,
			FloatValue: -1,
		})
		config := spelldata.AuraConfig(row, spelldata.Label("Hand of Edward the Odd Focus"))
		config.OnGain = func(_ *core.Aura, _ *core.Simulation) {
			instant.Activate()
		}
		config.OnExpire = func(_ *core.Aura, _ *core.Simulation) {
			instant.Deactivate()
		}
		config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.DefaultCast.CastTime > 0 && spell.CurCast.CastTime == 0 {
				aura.Deactivate(sim)
			}
		}
		focus := character.RegisterAura(config)

		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			focus.Activate(sim)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(2243, ppm, 0)
		})
		trigger.Name = "Hand of Edward the Odd"
		trigger.ActionID = core.ActionID{ItemID: 2243}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(2243, character.MakeProcTriggerAura(trigger))
	})

	// Dragon's Call
	// https://www.wowhead.com/forever/item=10847/dragons-call
	//
	// Chance on hit: Calls forth an Emerald Dragon Whelp to protect you in battle for a short period of
	// time.
	//
	// The game files state no rate for 13049; the item's chance-on-hit entry carries the one measured on
	// the sword's speed. Any landed hit rolls it, Rend's and Concussion Blow's applications included. The
	// row's category cooldown starts at the proc, and no hit rolls while it runs. The cast summons the
	// Emerald Dragon Whelp 8776 its E_SUMMON names, on the hit enemy, for the row's duration.
	core.NewItemPet(10847, newEmeraldDragonWhelp)
	newEntryRateItemEffect(10847, 13049, func(character *core.Character, ppm float64) {
		var whelp *emeraldDragonWhelp
		for _, pet := range character.PetAgents {
			if w, ok := pet.(*emeraldDragonWhelp); ok {
				whelp = w
				break
			}
		}
		row := spelldata.MustFind(13049)
		config := spelldata.SpellConfig(&character.Unit, row, spelldata.Flags(core.SpellFlagNoOnCastComplete))
		config.ProcMask = core.ProcMaskEmpty
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			whelp.CurrentTarget = target
			whelp.EnableWithTimeout(sim, whelp, row.Duration())
		}
		dragonsCall := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			dragonsCall.Cast(sim, result.Target)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(10847, ppm, 0)
		})
		trigger.Name = "Dragon's Call"
		trigger.ActionID = core.ActionID{ItemID: 10847}
		trigger.ExtraCondition = func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
			return dragonsCall.CD.IsReady(sim)
		}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(10847, character.MakeProcTriggerAura(trigger))
	})

	// Blade of Hanna
	// https://www.wowhead.com/forever/item=2801/blade-of-hanna
	//
	// Equip: Each melee hit grants you a stack of Improving Form. At 30 stacks perfect your form,
	// inflicting 490 Physical damage to your current target and increasing all your stats by 20 for
	// 20 sec.
	//
	// 1325535 adds a stack of 1325509 on each main-hand swing. At the stack count 1325509's second
	// effect names, 1325508 deals its damage, applies its stats and removes the stacks. No row states
	// a cooldown.
	core.NewItemEffect(2801, func(agent core.Agent) {
		character := agent.GetCharacter()
		stackRow := spelldata.MustFind(1325509)
		perfectedRow := spelldata.MustFind(1325508)
		damage := perfectedRow.DamageEffect()
		threshold := stackRow.EffectN(2).Misc

		perfected := character.RegisterAura(spelldata.AuraConfig(perfectedRow, spelldata.Label("Blade of Hanna Perfected Form")))
		spelldata.ParseEffects(character, perfected, perfectedRow, spelldata.Effects(2))

		config := spelldata.SpellConfig(&character.Unit, perfectedRow, spelldata.Melee(core.ProcMaskEmpty), spelldata.Proc())
		if perfectedRow.IsAProc() {
			config.Flags |= core.SpellFlagProc
		}
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeMeleeSpecialHitAndCrit)
			perfected.Activate(sim)
		}
		perfectForm := character.RegisterSpell(config)

		stacks := character.RegisterAura(spelldata.AuraConfig(stackRow, spelldata.Label("Blade of Hanna Improving Form")))

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1325535), func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			stacks.Activate(sim)
			stacks.AddStack(sim)
			if stacks.GetStacks() >= threshold {
				stacks.Deactivate(sim)
				perfectForm.Cast(sim, result.Target)
			}
		})
		trigger.Name = "Blade of Hanna"
		trigger.ActionID = core.ActionID{ItemID: 2801}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(2801, character.MakeProcTriggerAura(trigger))
	})

	// Heartseeking Crossbow
	// https://www.wowhead.com/forever/item=13040/heartseeking-crossbow
	//
	// Chance to strike your ranged target with a Shadowbolt for 13 to 19 Shadow damage.
	//
	// The game files state no rate for 29626; the item's entry carries it, measured on the
	// crossbow's speed.
	newEntryRateItemEffect(13040, 29640, func(character *core.Character, ppm float64) {
		row := spelldata.MustFind(29640)
		damage := row.DamageEffect()

		config := procSpellConfig(character, row)
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeRangedHitAndCrit)
		}
		shadowbolt := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(29626), func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			shadowbolt.Cast(sim, result.Target)
		}, spelldata.PPM(ppm))
		trigger.Name = "Heartseeking Crossbow"
		trigger.ActionID = core.ActionID{ItemID: 13040}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(13040, character.MakeProcTriggerAura(trigger))
	})

	// Blade of Eternal Darkness
	// https://www.wowhead.com/forever/item=17780/blade-of-eternal-darkness
	//
	// Chance on landing a damaging spell to deal 112 Shadow damage and restore 104 mana to you.
	//
	// Each direct spell hit that lands on an enemy rolls 21978, an area spell once for each enemy it
	// lands on, and the row's 0.1 s cooldown lets one landing proc it once. The chance is the row's
	// procs per minute on the cast time of the cast that launched the hit, at least 1.5 s
	// (castTimePPM): an instant, a triggered spell such as an Arcane Missiles missile, and a cast that
	// Presence of Mind or Heating Up cut short roll 4.5%, a 3 s Fireball 9%. Flamestrike's ground
	// ticks are direct hits and roll it; damage over time ticks, Ignite, wand shots and damage
	// shields never do. In game a harmful spell that deals no damage, such as Counterspell or the
	// start of the Arcane Missiles channel, rolls it too, 11 of the 554 procs; the sim leaves those
	// out. One mage's 554 procs in Forever dungeon logs fit 1.8 per minute (1.65 to 1.96); fitted per
	// tier, 4.5% (4.1 to 5.0) on instants, triggered spells and 1.5 s casts and 9.1% (7.5 to 10.9) on
	// 3 s casts. Misses are undetermined, with 2 in the sample, so only landed hits roll.
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 21978, BuffSpellID: 27860,
		Rate: castTimePPM(spelldata.MustFind(21978))},
		[]shared.ItemVariant{
			{ItemID: 17780, ItemName: "Blade of Eternal Darkness"},
		})

	// Spinal Reaper
	// https://www.wowhead.com/forever/item=288901/spinal-reaper
	//
	// Restores 150 mana or 20 rage when you kill a target that gives experience; this effect cannot
	// occur more than once every 10 seconds.
	//
	// 21185 hears the killing blow and fires 21186, whose effect 1 states the mana and effect 2 the
	// rage in tenths. A wearer with a rage bar and no mana bar gains the rage, and so does a druid in
	// Bear Form or Dire Bear Form at the killing blow; other forms gain the mana.
	core.NewItemEffect(288901, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(21185)
		gainID := core.ActionID{SpellID: 21186}

		addMana := shared.FiredSpellHandler(character, gainID, row)
		gain := addMana
		if character.HasRageBar() {
			rage := spelldata.MustFind(21186).EffectN(2).Tenths()
			metrics := character.NewRageMetrics(gainID)
			gain = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if addMana == nil || character.ShapeshiftForm == dbcenums.FORM_BEAR_FORM || character.ShapeshiftForm == dbcenums.FORM_DIRE_BEAR_FORM {
					character.AddRage(sim, rage, metrics)
				} else {
					addMana(sim, spell, result)
				}
			}
		}
		if gain == nil {
			return
		}

		trigger := spelldata.ProcTrigger(character, row, gain)
		trigger.Name = "Spinal Reaper"
		trigger.ActionID = core.ActionID{ItemID: 288901}
		character.ItemSwap.RegisterProc(288901, character.MakeProcTriggerAura(trigger))
	})

	// Ironfoe
	// https://www.wowhead.com/forever/item=11684/ironfoe
	//
	// Attacks have a chance to grant 2 extra attacks on your next swing. Attacks against Orcs are 2 times
	// as likely to activate this effect.
	//
	// The tooltip states no chance. It follows Hand of Justice's shape: 1301046's ProcChance is the
	// chance against an Orc and its second effect, a dummy, is the "2 times as likely".
	// TODO: Measure the rate in game, against Orcs and against other targets.
	core.NewItemEffect(11684, func(agent core.Agent) {
		registerLikelierExtraAttack(agent.GetCharacter(), 11684, "Ironfoe", 1301046, proto.CreatureFamily_CreatureFamilyOrc)
	})

	// Naga Heartpiercer
	// https://www.wowhead.com/forever/item=3078/naga-heartpiercer
	//
	// Chance on hit: Wounds the target for 42 damage. Naga are Stunned for 2 sec.
	//
	// TODO: waiting for PPM data
	//
	// Pierce Heart rolls on the cast result of a shot that lands, at the shot's release, and strikes
	// then. 1292647 is physical on the melee table. The stun on Naga is crowd control the sim leaves out.
	// TODO: Ingame test needed: whether a missed shot rolls it too.
	// core.NewItemEffect(3078, func(agent core.Agent) {
	// 	character := agent.GetCharacter()
	// 	row := spelldata.MustFind(1292647)
	// 	damage := row.DamageEffect()
	//
	// 	config := procSpellConfig(character, row)
	// 	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeMeleeSpecialHitAndCrit)
	// 	}
	// 	pierceHeart := character.RegisterSpell(config)
	//
	// 	ppm := weaponProcPPM(3078, row.ID)
	// 	trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
	// 		pierceHeart.Cast(sim, result.Target)
	// 	}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
	// 		trigger.Callback = core.CallbackOnCastResult
	// 		trigger.Outcome = core.OutcomeLanded
	// 		trigger.ProcChance = 0
	// 		trigger.DPM = character.NewDynamicLegacyProcForWeapon(3078, ppm, 0)
	// 	})
	// 	trigger.Name = "Naga Heartpiercer"
	// 	trigger.ActionID = core.ActionID{ItemID: 3078}
	// 	trigger.TriggerImmediately = true
	// 	character.ItemSwap.RegisterProc(3078, character.MakeProcTriggerAura(trigger))
	// })

	// White Obsidian Wand
	// https://www.wowhead.com/forever/item=274425/white-obsidian-wand
	//
	// Chance on hit to reduce the mana cost of your next spell cast within 10s by 24.
	//
	// TODO: waiting for PPM data
	//
	// 1294058 lowers the cost of a spell of the schools its mask names, and the next such spell
	// with a cost spends its one charge.
	// core.NewItemEffect(274425, func(agent core.Agent) {
	// 	character := agent.GetCharacter()
	// 	if !character.HasManaBar() {
	// 		return
	// 	}
	// 	row := spelldata.MustFind(1294058)
	// 	effect := row.FirstAura(dbcenums.A_MOD_POWER_COST_SCHOOL)
	// 	school := core.SpellSchool(effect.Misc)
	//
	// 	cheaper := character.AddDynamicMod(core.SpellModConfig{
	// 		Kind:     core.SpellMod_PowerCost_Flat,
	// 		School:   school,
	// 		IntValue: int32(effect.Average()),
	// 	})
	// 	config := spelldata.AuraConfig(row, spelldata.Label("White Obsidian Wand"))
	// 	config.OnGain = func(_ *core.Aura, _ *core.Simulation) {
	// 		cheaper.Activate()
	// 	}
	// 	config.OnExpire = func(_ *core.Aura, _ *core.Simulation) {
	// 		cheaper.Deactivate()
	// 	}
	// 	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
	// 		if spell.Cost != nil && spell.SpellSchool.Matches(school) {
	// 			aura.Deactivate(sim)
	// 		}
	// 	}
	// 	whiteObsidian := character.RegisterAura(config)
	//
	// 	ppm := weaponProcPPM(274425, row.ID)
	// 	trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1294055), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
	// 		whiteObsidian.Activate(sim)
	// 	}, spelldata.PPM(ppm))
	// 	trigger.Name = "White Obsidian Wand"
	// 	trigger.ActionID = core.ActionID{ItemID: 274425}
	// 	trigger.TriggerImmediately = true
	// 	character.ItemSwap.RegisterProc(274425, character.MakeProcTriggerAura(trigger))
	// })

	// Rot-Covered Harpoon
	// https://www.wowhead.com/forever/item=274963/rot-covered-harpoon
	//
	// Chance on hit: Diseases target for 25 Nature damage every 1 sec for 7 sec. Deals 2 times as much
	// damage to Aquatic enemies.
	//
	// TODO: waiting for PPM data
	// core.NewItemEffect(274963, func(agent core.Agent) {
	// 	character := agent.GetCharacter()
	// 	row := spelldata.MustFind(1295744)
	// 	tick := row.PeriodicDamageEffect()
	// 	familyMultiplier := row.Effects[1].BaseValue()
	//
	// 	config := procSpellConfig(character, row)
	// 	config.Dot = spelldata.DotConfig(row, tick)
	// 	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 		amount := tick.Average() * creatureFamilyMultiplier(target, familyMultiplier, proto.CreatureFamily_CreatureFamilyAquatic)
	// 		dot.Spell.CalcAndDealPeriodicDamage(sim, target, amount, row.TickOutcome(dot))
	// 	}
	// 	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
	// 		if result.Landed() {
	// 			spell.Dot(target).Apply(sim)
	// 		}
	// 		spell.DealOutcome(sim, result)
	// 	}
	// 	rot := character.RegisterSpell(config)
	//
	// 	ppm := weaponProcPPM(274963, row.ID)
	// 	trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
	// 		rot.Cast(sim, result.Target)
	// 	}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
	// 		trigger.ProcChance = 0
	// 		trigger.DPM = character.NewDynamicLegacyProcForWeapon(274963, ppm, 0)
	// 	})
	// 	trigger.Name = "Rot-Covered Harpoon"
	// 	trigger.ActionID = core.ActionID{ItemID: 274963}
	// 	trigger.TriggerImmediately = true
	// 	character.ItemSwap.RegisterProc(274963, character.MakeProcTriggerAura(trigger))
	// })

	// Hammer of Expertise
	// https://www.wowhead.com/forever/item=8708/hammer-of-expertise
	//
	// Chance on hit: Deal 112 Fire damage. Deals 3 times as much damage to Earth Elementals, Fire
	// Elementals, and Mountain Giants.
	//
	// TODO: waiting for PPM data
	//
	// 1317432 states Physical on the melee table; the tooltip's Fire is a spell hit.
	// core.NewItemEffect(8708, func(agent core.Agent) {
	// 	character := agent.GetCharacter()
	// 	row := spelldata.MustFind(1317432)
	// 	damage := row.DamageEffect()
	// 	familyMultiplier := row.Effects[1].BaseValue()
	//
	// 	config := procSpellConfig(character, row)
	// 	config.SpellSchool = core.SpellSchoolFire
	// 	config.DefenseType = core.DefenseTypeMagic
	// 	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		amount := damage.DirectAmount(sim) * creatureFamilyMultiplier(target, familyMultiplier,
	// 			proto.CreatureFamily_CreatureFamilyEarthElemental, proto.CreatureFamily_CreatureFamilyFireElemental,
	// 			proto.CreatureFamily_CreatureFamilyMountainGiant)
	// 		spell.CalcAndDealDamage(sim, target, amount, spell.OutcomeMagicHitAndCrit)
	// 	}
	// 	forgingStrike := character.RegisterSpell(config)
	//
	// 	ppm := weaponProcPPM(8708, row.ID)
	// 	trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
	// 		forgingStrike.Cast(sim, result.Target)
	// 	}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
	// 		trigger.ProcChance = 0
	// 		trigger.DPM = character.NewDynamicLegacyProcForWeapon(8708, ppm, 0)
	// 	})
	// 	trigger.Name = "Hammer of Expertise"
	// 	trigger.ActionID = core.ActionID{ItemID: 8708}
	// 	trigger.TriggerImmediately = true
	// 	character.ItemSwap.RegisterProc(8708, character.MakeProcTriggerAura(trigger))
	// })
}

// The spell each hand-written proc reads its item's procs per minute for, by item id. The generator
// lists an item whose entry states no rate for that spell as missing its effect.
var EntryRateProcs = map[int32]int32{}

// Registers an item effect whose proc runs at the procs per minute the item's chance-on-hit entry
// states for spellID, since the game files state no rate for it. While the entry states none, the
// effect registers nothing and the item keeps only its stats; no rate is made up for the proc.
func newEntryRateItemEffect(itemID int32, spellID int32, apply func(character *core.Character, ppm float64)) {
	EntryRateProcs[itemID] = spellID
	core.NewItemEffect(itemID, func(agent core.Agent) {
		if ppm := weaponProcPPM(itemID, spellID); ppm > 0 {
			apply(agent.GetCharacter(), ppm)
		}
	})
}

// The procs per minute the item's chance-on-hit entry for the spell carries, or 0 where it states none.
func weaponProcPPM(itemID int32, spellID int32) float64 {
	for _, effect := range core.GetItemByID(itemID).ItemEffects {
		if effect.BuffId == spellID && effect.GetProc().GetPpm() > 0 {
			return effect.GetProc().GetPpm()
		}
	}
	return 0
}

// An item's extra attack proc whose row states the chance against one creature family in its
// ProcChance column, and on its second effect how many times likelier a hit on that family is than
// one on any other target. The extra attack is the spell the row's first effect triggers. A hit on
// the family rolls the row's column and a hit on any other target the column divided by the second
// effect; the condition draws that one roll per hit under the item's name, so the trigger's own
// chance is 1 and draws nothing, and the row's cooldown starts only on a proc.
func registerLikelierExtraAttack(character *core.Character, itemID int32, name string, rowID int32, family proto.CreatureFamily) {
	row := spelldata.MustFind(rowID)
	extraAttacks := row.EffectN(1).Trigger()
	factor := row.EffectN(2).BasePoints
	elsewhere := row.StatedChance() / factor

	trigger := spelldata.ProcTrigger(character, row, nil)
	trigger.Name = name
	trigger.ActionID = core.ActionID{ItemID: itemID}
	trigger.ProcChance = 1
	trigger.ExtraCondition = func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) bool {
		return sim.Proc(elsewhere*creatureFamilyMultiplier(result.Target, factor, family), name)
	}
	onInit := character.ExtraAttackProc(&trigger, extraAttacks.ID, int(extraAttacks.ExtraAttacksEffect().BasePoints))
	aura := character.MakeProcTriggerAura(trigger)
	aura.ApplyOnInit(onInit)
	character.ItemSwap.RegisterProc(itemID, aura)
}

// The factor an effect applies against the creature families it names, and 1 against any other
// target.
func creatureFamilyMultiplier(target *core.Unit, multiplier float64, families ...proto.CreatureFamily) float64 {
	if slices.Contains(families, target.CreatureFamily) {
		return multiplier
	}
	return 1
}

// The row's procs per minute, priced on the cast time of the cast that launched the hit, at least the
// global cooldown's 1.5 s, and rolled by each direct spell hit that deals damage. The cast time is
// read where the cast deals the hit's result, which is the cast's end for a missile, so one still in
// flight keeps it when the next cast of the spell starts. It is the time the cast took after talents,
// Presence of Mind and cast speed. The chance depends on the hit, so Chance(1) leaves core no roll of
// its own and the condition draws it after the cooldown check and before the cooldown starts.
// TODO: Verify in game whether cast speed shortens the time the chance is priced on; Omen of Clarity's
// is priced on the unhasted time.
func castTimePPM(row *spelldata.Spell) spelldata.ProcOpt {
	return func(character *core.Character, trigger *core.ProcTrigger) {
		castTimes := map[*core.SpellResult]time.Duration{}
		core.MakePermanent(character.RegisterAura(core.Aura{
			Label: row.Name + " Cast Times",
			OnReset: func(_ *core.Aura, _ *core.Simulation) {
				clear(castTimes)
			},
			OnCastResult: func(_ *core.Aura, _ *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				castTimes[result] = spell.CurCast.CastTime
			},
		}))

		trigger.Callback = core.CallbackOnSpellHitDealt
		trigger.Outcome = core.OutcomeLanded
		trigger.RequireDamageDealt = true
		flags := trigger.ProcSpellClassFlags
		trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			castTime, ok := castTimes[result]
			if !ok {
				castTime = spell.CurCast.CastTime
			}
			delete(castTimes, result)
			chance := float64(row.RPPM) * max(castTime, core.GCDDefault).Seconds() / 60
			return sim.Proc(character.ModifiedProcChance(flags, chance), row.Name)
		}
		spelldata.Chance(1)(character, trigger)
	}
}
