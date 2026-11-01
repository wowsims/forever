package forever

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
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

	// Sulfuras, Hand of Ragnaros
	// https://www.wowhead.com/forever/item=17182/sulfuras-hand-of-ragnaros
	//
	// Equip: Deals 5 Fire damage to anyone who strikes you with a melee attack.
	// Chance on hit: Hurls a fiery ball that causes 303 Fire damage and an additional 75 damage over
	// 10s.
	//
	// The game files state no rate for the Fireball; the item's chance-on-hit entry carries it.
	core.NewItemEffect(17182, func(agent core.Agent) {
		character := agent.GetCharacter()

		immolation := spelldata.MustFind(21142)
		shieldAura := core.MakePermanent(character.RegisterAura(spelldata.AuraConfig(immolation, spelldata.Label("Sulfuras, Hand of Ragnaros Immolation"))))
		for _, shield := range spelldata.ItemAuraDamageShields(immolation) {
			buffs.AttachDamageShield(shieldAura, immolation.SpellSchool(), shield.Average(), shield.Coeff())
		}
		character.ItemSwap.RegisterProc(17182, shieldAura)

		row := spelldata.MustFind(21162)
		damage := row.DamageEffect()
		config := procSpellConfig(character, row)
		config.Dot = spelldata.DotConfig(row, row.PeriodicDamageEffect())
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeMagicHitAndCrit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
		}
		fireball := character.RegisterSpell(config)

		ppm := weaponProcPPM(17182, row.ID)
		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			fireball.Cast(sim, result.Target)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(17182, ppm, 0)
		})
		trigger.Name = "Sulfuras, Hand of Ragnaros"
		trigger.ActionID = core.ActionID{ItemID: 17182}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(17182, character.MakeProcTriggerAura(trigger))
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
	core.NewItemEffect(7717, func(agent core.Agent) {
		character := agent.GetCharacter()

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

		ppm := weaponProcPPM(7717, row.ID)
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
	// The game files state no rate for Focus; the item's chance-on-hit entry carries it. The row makes
	// the cast instant with +100000% cast speed, which the sim states as no cast time at all, so the
	// global cooldown stays whole. The next spell with a cast time spends the one charge.
	core.NewItemEffect(2243, func(agent core.Agent) {
		character := agent.GetCharacter()
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

		ppm := weaponProcPPM(2243, row.ID)
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
	core.NewItemEffect(13040, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(29640)
		damage := row.DamageEffect()

		config := procSpellConfig(character, row)
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeRangedHitAndCrit)
		}
		shadowbolt := character.RegisterSpell(config)

		ppm := weaponProcPPM(13040, row.ID)
		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(29626), func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			shadowbolt.Cast(sim, result.Target)
		}, spelldata.PPM(ppm))
		trigger.Name = "Heartseeking Crossbow"
		trigger.ActionID = core.ActionID{ItemID: 13040}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(13040, character.MakeProcTriggerAura(trigger))
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

		var gain core.ProcHandler
		if character.HasRageBar() {
			rage := spelldata.MustFind(21186).EffectN(2).Tenths()
			metrics := character.NewRageMetrics(gainID)
			addRage := func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				character.AddRage(sim, rage, metrics)
			}
			if character.HasManaBar() {
				addMana := shared.FiredSpellHandler(character, gainID, row)
				gain = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if character.ShapeshiftForm == dbcenums.FORM_BEAR_FORM || character.ShapeshiftForm == dbcenums.FORM_DIRE_BEAR_FORM {
						addRage(sim, spell, result)
					} else if addMana != nil {
						addMana(sim, spell, result)
					}
				}
			} else {
				gain = addRage
			}
		} else {
			gain = shared.FiredSpellHandler(character, gainID, row)
		}
		if gain == nil {
			return
		}

		trigger := spelldata.ProcTrigger(character, row, gain)
		trigger.Name = "Spinal Reaper"
		trigger.ActionID = core.ActionID{ItemID: 288901}
		character.ItemSwap.RegisterProc(288901, character.MakeProcTriggerAura(trigger))
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

// The procs per minute the item's chance-on-hit entry for the spell carries.
func weaponProcPPM(itemID int32, spellID int32) float64 {
	for _, effect := range core.GetItemByID(itemID).ItemEffects {
		if effect.BuffId == spellID && effect.GetProc().GetPpm() > 0 {
			return effect.GetProc().GetPpm()
		}
	}
	panic(fmt.Sprintf("item %d states no procs per minute for spell %d", itemID, spellID))
}

// The multiplier an effect deals against the creature families it names, and 1 against any other
// target.
func creatureFamilyMultiplier(target *core.Unit, multiplier float64, families ...proto.CreatureFamily) float64 {
	if slices.Contains(families, target.CreatureFamily) {
		return multiplier
	}
	return 1
}
