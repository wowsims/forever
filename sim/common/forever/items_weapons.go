package forever

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
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

	// Skullforge Reaver
	// https://www.wowhead.com/forever/item=13361/skullforge-reaver
	//
	// Chance on hit: Drains target for 21 Shadow damage every 1 sec and transfers it to the caster.
	// Lasts for 10 sec.
	//
	// The game files state no rate for 17484; the item's chance-on-hit entry carries it, measured on
	// the sword's speed. Landed white swings and melee specials of the hand that holds it roll it alike,
	// Windfury Totem strikes too; the other hand's swings and ranged hits never do, and a proc on a
	// branded target refreshes the brand (EV-0121). 17484's first effect drains the target and its
	// second heals the wearer as a heal over time of its own. Both crit, as the row's Periodic Can Crit says.
	newEntryRateItemEffect(13361, 17484, func(character *core.Character, ppm float64) {
		row := spelldata.MustFind(17484)
		heal := row.FirstAura(dbcenums.A_PERIODIC_HEAL)
		healCrits := row.PeriodicCanCrit() && !row.CannotCrit()

		healConfig := procSpellConfig(character, row)
		healConfig.ProcMask = core.ProcMaskSpellHealing
		healConfig.ActionID.Tag = 1
		healConfig.Hot = spelldata.DotConfig(row, heal)
		healConfig.Hot.SelfOnly = true
		healConfig.Hot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			outcome := dot.OutcomeTick
			if healCrits {
				outcome = dot.Spell.OutcomeTickHealingCrit
			}
			dot.Spell.CalcAndDealPeriodicHealing(sim, target, heal.Roll(sim), outcome)
		}
		transfer := character.RegisterSpell(healConfig)

		config := procSpellConfig(character, row)
		config.Dot = spelldata.DotConfig(row, row.PeriodicDamageEffect())
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
				transfer.SelfHot().Apply(sim)
			}
			spell.DealOutcome(sim, result)
		}
		brand := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			brand.Cast(sim, result.Target)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(13361, ppm, 0)
		})
		trigger.Name = "Skullforge Reaver"
		trigger.ActionID = core.ActionID{ItemID: 13361}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(13361, character.MakeProcTriggerAura(trigger))
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
	// shields never do. A harmful spell that deals no damage, such as Counterspell or the start of
	// the Arcane Missiles channel, rolls it too; the sim leaves those out. Whether a miss rolls is
	// undetermined, so only landed hits roll (EV-0025).
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 21978, BuffSpellID: 27860,
		Rate: castTimePPM(spelldata.MustFind(21978), true)},
		[]shared.ItemVariant{
			{ItemID: 17780, ItemName: "Blade of Eternal Darkness"},
		})

	// Moonshadow Stave
	// https://www.wowhead.com/forever/item=22458/moonshadow-stave
	//
	// Harmful spell casts have a chance to deal 91 Arcane or Shadow damage.
	//
	// 1300435 rolls once for each enemy a harmful spell lands on, a damage over time effect or a debuff
	// that deals no damage included, apart from Curse of the Elements, Recklessness and Weakness and Hex
	// of Weakness, which the sim registers with no proc mask, and its 10 ms cooldown lets one landing
	// proc it once. The chance is the row's procs per minute on the cast time of the cast that launched
	// the spell, at least 1.5 s (castTimePPM). Damage over time ticks, misses, helpful spells, melee and spells fired as a proc,
	// Hurricane's ticks among them, never roll it. Each proc casts Shadow 1300437 or Arcane 1300438, one
	// in two each, and both fly to the target. In game a Hurricane cast rolls once for each enemy its
	// area reaches until one procs; the sim's Hurricane cast only starts its channel and lands on no
	// enemy, so it rolls nothing (EV-0115).
	core.NewItemEffect(22458, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(1300435)

		var bolts [2]*core.Spell
		for i, id := range []int32{1300437, 1300438} {
			damage := spelldata.MustFind(id).DamageEffect()
			config := procSpellConfig(character, spelldata.MustFind(id))
			config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.DealDamageAfterTravel(sim, spell.CalcDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeMagicHitAndCrit))
			}
			bolts[i] = character.RegisterSpell(config)
		}

		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			bolt := bolts[0]
			if sim.Proc(0.5, "Moonshadow Stave") {
				bolt = bolts[1]
			}
			bolt.Cast(sim, result.Target)
		}, castTimePPM(row, false))
		priced := trigger.ExtraCondition
		trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			return !firedAsAProc(spell) && priced(sim, spell, result)
		}
		trigger.Name = "Moonshadow Stave"
		trigger.ActionID = core.ActionID{ItemID: 22458}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(22458, character.MakeProcTriggerAura(trigger))
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
	// Logs fit that reading: half the column against other targets, the column against Orcs (EV-0119).
	core.NewItemEffect(11684, func(agent core.Agent) {
		registerExtraAttackVersus(agent.GetCharacter(), 11684, "Ironfoe", 1301046)
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
	// Landed wand shots roll it at the item entry's procs per minute on the wand's speed (EV-0119).
	//
	// 1294058 lowers the cost of a spell of the schools its mask names, and the next such spell that
	// pays mana spends its one charge. A free cast keeps it, and so does one the reduction itself
	// brings to 0. The listener is hand-written: the row's proc flag, cast successful, decodes to no
	// listener, and spelldata.AttachChargeSpender on cast complete would spend the charge on Shoot.
	newEntryRateItemEffect(274425, 1294058, func(character *core.Character, ppm float64) {
		if !character.HasManaBar() {
			return
		}
		row := spelldata.MustFind(1294058)
		effect := row.FirstAura(dbcenums.A_MOD_POWER_COST_SCHOOL)
		school := core.SpellSchool(effect.Misc)

		cheaper := character.AddDynamicMod(core.SpellModConfig{
			Kind:     core.SpellMod_PowerCost_Flat,
			School:   school,
			IntValue: int32(effect.Average()),
		})
		config := spelldata.AuraConfig(row, spelldata.Label("White Obsidian Wand"))
		config.OnGain = func(_ *core.Aura, _ *core.Simulation) {
			cheaper.Activate()
		}
		config.OnExpire = func(_ *core.Aura, _ *core.Simulation) {
			cheaper.Deactivate()
		}
		config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Cost != nil && spell.CurCast.Cost > 0 && spell.SpellSchool.Matches(school) {
				aura.Deactivate(sim)
			}
		}
		whiteObsidian := character.RegisterAura(config)

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1294055), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			whiteObsidian.Activate(sim)
		}, spelldata.PPM(ppm))
		trigger.Name = "White Obsidian Wand - Trigger"
		trigger.ActionID = core.ActionID{ItemID: 274425}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(274425, character.MakeProcTriggerAura(trigger))
	})

	// Sul'thraze the Lasher
	// https://www.wowhead.com/forever/item=9372/sulthraze-the-lasher
	//
	// Chance on hit: Lash the enemy with the rage of Sul'thraze, reducing their attack power by 180 for
	// 15s and dealing 159 Shadow damage. After 5s, the target suffers an additional 159 Shadow damage.
	//
	// A landed Rage of Sul'thraze 1294430 puts Shadow Lash 1294431 on the target, which deals its damage
	// once, when its 5 s run out. A Rage that lands while a Shadow Lash is still running puts up no
	// second one and leaves the running one's time alone (EV-0136). The Lash rolls no hit of its own and
	// can crit, as its row states.
	shared.DamageProcExtras[9372] = func(character *core.Character, config *core.SpellConfig) {
		row := spelldata.MustFind(1294431)
		lashConfig := procSpellConfig(character, row)
		lashConfig.Dot = spelldata.DotConfig(row, row.PeriodicDamageEffect())
		lashConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
		}
		lash := character.RegisterSpell(lashConfig)

		rage := config.ActionID
		core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Sul'thraze the Lasher - Shadow Lash",
			OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell.ActionID == rage && result.Landed() && !lash.Dot(result.Target).IsActive() {
					lash.Cast(sim, result.Target)
				}
			},
		}))
	}

	// Pendulum of Doom
	// https://www.wowhead.com/forever/item=9425/pendulum-of-doom
	//
	// Chance on hit: Delivers a fatal wound for 311 Physical damage. Deals 50% increased damage to
	// targets below 25% health.
	//
	// Fatal Wound 10373 deals the percent its second dummy states more to a target below the health
	// percent its third states. Hits of 1.5 times the plain range land late in a target's life, but the
	// logs state no health, so the threshold is the row's alone (EV-0136). The sim knows a target's
	// health only at 35% and 20%, so the bonus applies from 20%.
	// TODO: Give the bonus between 25% and 20% once the sim states a target's health there.
	shared.DamageProcExtras[9425] = func(_ *core.Character, config *core.SpellConfig) {
		bonus := 1 + spelldata.MustFind(10373).EffectN(2).Average()/100
		deal := config.ApplyEffects
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if !target.IsExecutePhase20() {
				deal(sim, target, spell)
				return
			}
			multiplier := spell.DamageMultiplier
			spell.DamageMultiplier *= bonus
			deal(sim, target, spell)
			spell.DamageMultiplier = multiplier
		}
	}

	// Flame Wrath
	// https://www.wowhead.com/forever/item=11809/flame-wrath
	//
	// Chance on hit: Shoots a ring of fire dealing 154 Fire damage to all nearby enemies, and envelops
	// the caster with a Fire Shield for 15s.
	//
	// Each proc also puts 16560 on the wearer for its duration, a later proc refreshing it (EV-0136):
	// it raises Fire resistance and deals its damage shield's amount to each attacker that strikes
	// the wearer.
	shared.DamageProcExtras[11809] = func(character *core.Character, config *core.SpellConfig) {
		row := spelldata.MustFind(16560)
		shield := character.RegisterAura(spelldata.AuraConfig(row, spelldata.Label("Flame Wrath Fire Shield")))
		spelldata.ParseEffects(character, shield, row, spelldata.Effects(spelldata.ItemAuraParsedOn(row, spelldata.AuraOnWearer)...))
		for _, effect := range spelldata.ItemAuraDamageShields(row) {
			buffs.AttachDamageShield(shield, row.SpellSchool(), effect.Average(), effect.Coeff(), spelldata.DamageShieldHits(row), spelldata.DamageShieldLockout(row))
		}
		deal := config.ApplyEffects
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			deal(sim, target, spell)
			shield.Activate(sim)
		}
	}

	// Gahz'rilla Fang
	// https://www.wowhead.com/forever/item=9467/gahzrilla-fang
	//
	// Chance on hit: Deal 84 Nature damage, increased by 2% per stack of Static Electricity. Deals 2
	// times as much damage to Aquatic enemies.
	//
	// 3742 deals the percent Static Electricity 1294477 states more for each stack of it on the wearer,
	// and each proc spends all of them but one (EV-0136).
	// TODO: Static Electricity builds while the wearer moves, at a rate no row states and the logs do
	// not settle in time between stacks; measure how far the wearer moves for a stack. Until then the
	// sim gives no stacks.
	shared.DamageProcExtras[9467] = func(character *core.Character, config *core.SpellConfig) {
		row := spelldata.MustFind(1294477)
		perStack := row.EffectN(1).Average() / 100
		static := character.RegisterAura(spelldata.AuraConfig(row))
		deal := config.ApplyEffects
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			stacks := static.GetStacks()
			multiplier := spell.DamageMultiplier
			spell.DamageMultiplier *= 1 + perStack*float64(stacks)
			deal(sim, target, spell)
			spell.DamageMultiplier = multiplier
			if stacks > 1 {
				static.SetStacks(sim, 1)
			}
		}
	}

	// The Chief's Enforcer
	// https://www.wowhead.com/forever/item=9477/the-chiefs-enforcer
	//
	// Chance on hit: Deal 294 Physical damage and stun for 1.5s. Trolls take 2 times as much damage
	// and stun duration.
	//
	// Every proc also casts 1294340 at the target, which rolls the spell hit table on its own and stuns
	// for its duration: a dodged, parried or missed strike stuns all the same (EV-0136). A stun-immune
	// target, a raid boss among them, takes none.
	shared.DamageProcExtras[9477] = func(character *core.Character, config *core.SpellConfig) {
		row := spelldata.MustFind(1294340)
		label := fmt.Sprintf("%s %d", row.Name, row.ID)
		stuns := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			if aura := target.GetAura(label); aura != nil {
				return aura
			}
			return target.RegisterStunAura(label, core.ActionID{SpellID: row.ID}, row.Duration())
		})
		stunConfig := procSpellConfig(character, row)
		if row.IsStun() {
			stunConfig.Flags |= core.SpellFlagStun
		}
		stunConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit).Landed() {
				core.ApplyStun(sim, stuns.Get(target))
			}
		}
		stun := character.RegisterSpell(stunConfig)
		deal := config.ApplyEffects
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			deal(sim, target, spell)
			stun.Cast(sim, target)
		}
	}

	// Iceblade Hacker
	// https://www.wowhead.com/forever/item=13952/iceblade-hacker
	//
	// Melee attacks with this weapon deal 41 Frost damage. Chance on hit: Deal an additional 105 Frost
	// damage and slow movement speed by 50% for 5s.
	//
	// Core takes one effect per item, so both procs are written here. 1298413 casts 1298414 on each landed
	// main-hand swing and special, on the melee table its row states. 1298416 strikes at the procs per
	// minute of the item's chance-on-hit entry on the axe's speed (EV-0136), and registers nothing while
	// the entry states none; its slow is movement the sim leaves out.
	EntryRateProcs[13952] = 1298416
	core.NewItemEffect(13952, func(agent core.Agent) {
		character := agent.GetCharacter()

		chillRow := spelldata.MustFind(1298414)
		chill := chillRow.DamageEffect()
		chillConfig := spelldata.SpellConfig(&character.Unit, chillRow, spelldata.Melee(core.ProcMaskEmpty), spelldata.Proc())
		if chillRow.IsAProc() {
			chillConfig.Flags |= core.SpellFlagProc
		}
		chillConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.DealDamageAfterTravel(sim, spell.CalcDamage(sim, target, chill.DirectAmount(sim), spell.OutcomeMeleeSpecialHitAndCrit))
		}
		chillSpell := character.RegisterSpell(chillConfig)

		chillTrigger := spelldata.ProcTrigger(character, spelldata.MustFind(1298413), func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			chillSpell.Cast(sim, result.Target)
		})
		chillTrigger.Name = "Iceblade Hacker"
		chillTrigger.ActionID = core.ActionID{ItemID: 13952}
		chillTrigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(13952, character.MakeProcTriggerAura(chillTrigger))

		ppm := weaponProcPPM(13952, 1298416)
		if ppm == 0 {
			return
		}
		boltRow := spelldata.MustFind(1298416)
		bolt := boltRow.DamageEffect()
		boltConfig := procSpellConfig(character, boltRow)
		boltConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.DealDamageAfterTravel(sim, spell.CalcDamage(sim, target, bolt.DirectAmount(sim), spell.OutcomeMagicHitAndCrit))
		}
		frostbolt := character.RegisterSpell(boltConfig)

		boltTrigger := spelldata.ProcTrigger(character, boltRow, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			frostbolt.Cast(sim, result.Target)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(13952, ppm, 0)
		})
		boltTrigger.Name = "Iceblade Hacker - Frostbolt"
		boltTrigger.ActionID = core.ActionID{ItemID: 13952}
		boltTrigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(13952, character.MakeProcTriggerAura(boltTrigger))
	})

	// Mana Channeling Wand
	// https://www.wowhead.com/forever/item=18483/mana-channeling-wand
	//
	// Chance on hit to restore 65 mana.
	//
	// Landed wand shots roll 1302348 at the procs per minute of the item's entry on the wand's speed,
	// and each proc restores the mana of 1302350, the spell 1302348 fires, not the 65 the tooltip
	// quotes from 1302348 (EV-0136). A wearer with no mana bar registers nothing.
	newEntryRateItemEffect(18483, 1302350, func(character *core.Character, ppm float64) {
		row := spelldata.MustFind(1302348)
		handler := shared.FiredSpellHandler(character, core.ActionID{ItemID: 18483}, row)
		if handler == nil {
			return
		}
		trigger := spelldata.ProcTrigger(character, row, handler, spelldata.PPM(ppm))
		trigger.Name = "Mana Channeling Wand"
		trigger.ActionID = core.ActionID{ItemID: 18483}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(18483, character.MakeProcTriggerAura(trigger))
	})
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

// An item's extra attack proc whose row states the chance against its versus targets in its
// ProcChance column, and in its versus bonus how many times likelier a hit on them is than one on any
// other target. The extra attack is the spell the row's first effect triggers. A hit on a versus
// target rolls the row's column and a hit on any other target the column divided by the factor.
func registerExtraAttackVersus(character *core.Character, itemID int32, name string, rowID int32) {
	row := spelldata.MustFind(rowID)
	extraAttacks := row.EffectN(1).Trigger()

	trigger := spelldata.ProcTrigger(character, row, nil)
	trigger.Name = name
	trigger.ActionID = core.ActionID{ItemID: itemID}
	versus := shared.ItemVersusBonus(character, itemID, proto.VersusEffectKind_VersusEffectKindProcChance, row)
	if versus == nil {
		panic(fmt.Sprintf("%d states no versus bonus for %s", rowID, name))
	}
	trigger.ProcChance = 0
	trigger.DPM = character.NewTargetScaledChanceManager(row.StatedChance()/versus.Multiplier, trigger.ProcMask, versus.Against)
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
// global cooldown's 1.5 s, and rolled by each direct spell hit that lands; with requireDamage, only by
// one that deals damage. The cast time is read where the cast deals the hit's result, which is the
// cast's end for a missile, so one still in flight keeps it when the next cast of the spell starts. It
// is the time the cast took after talents, Presence of Mind and cast speed. The chance depends on the
// hit, so Chance(1) leaves core no roll of its own and the condition draws it after the cooldown check
// and before the cooldown starts.
// TODO: Verify in game whether cast speed shortens the time the chance is priced on; Omen of Clarity's
// is priced on the unhasted time.
func castTimePPM(row *spelldata.Spell, requireDamage bool) spelldata.ProcOpt {
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
		trigger.RequireDamageDealt = requireDamage
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
