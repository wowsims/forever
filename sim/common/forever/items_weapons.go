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
	// The row is a channel: for its 9 s the wearer swings no auto attack and casts nothing. Each tick's
	// strike rolls the other weapon procs but not Bladestorm.
	newEntryRateItemEffect(7717, 9632, func(character *core.Character, ppm float64) {
		strikeRow := spelldata.MustFind(9633)
		strikeConfig := spelldata.SpellConfig(&character.Unit, strikeRow, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc(), spelldata.Triggered())
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
	// The sim states the row's +100000% cast speed as no cast time, so the global cooldown stays whole.
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
	shared.RegisterItemGuardian(10847, spelldata.MustFind(13049), emeraldDragonWhelp)
	newEntryRateItemEffect(10847, 13049, func(character *core.Character, ppm float64) {
		row := spelldata.MustFind(13049)
		whelps := shared.Guardians(character, row)
		config := spelldata.SpellConfig(&character.Unit, row, spelldata.Flags(core.SpellFlagNoOnCastComplete))
		config.ProcMask = core.ProcMaskEmpty
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			shared.SummonGuardians(sim, whelps, target)
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
	// No row states a cooldown.
	core.NewItemEffect(2801, func(agent core.Agent) {
		character := agent.GetCharacter()
		stackRow := spelldata.MustFind(1325509)
		perfectedRow := spelldata.MustFind(1325508)
		damage := perfectedRow.DamageEffect()
		threshold := stackRow.EffectN(2).Misc

		perfected := character.RegisterAura(spelldata.AuraConfig(perfectedRow, spelldata.Label("Blade of Hanna Perfected Form")))
		spelldata.ParseEffects(character, perfected, perfectedRow, spelldata.Effects(2))

		config := spelldata.SpellConfig(&character.Unit, perfectedRow, spelldata.Melee(core.ProcMaskEmpty), spelldata.Proc(), spelldata.Triggered())
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

	// Claymore of Two Moons
	// https://www.wowhead.com/forever/item=280080/claymore-of-two-moons
	//
	// Equip: Your attacks have a chance to Wax, dealing 189 Arcane damage or Wane, granting 40 Strength
	// for 10s.
	//
	// 1309550 states no odds: Wax or Wane, one in two (EV-0163). Waxing Moon takes a magic proc's spell
	// damage mask (R-0051), which R-0051 leaves open for a hand-written item spell.
	core.NewItemEffect(280080, func(agent core.Agent) {
		character := agent.GetCharacter()
		waxRow := spelldata.MustFind(1319378)
		waneRow := spelldata.MustFind(1319382)
		damage := waxRow.DamageEffect()

		config := procSpellConfig(character, waxRow)
		config.ProcMask = core.ProcMaskSpellDamage
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.DealDamageAfterTravel(sim, spell.CalcDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeMagicHitAndCrit))
		}
		wax := character.RegisterSpell(config)

		wane := character.RegisterAura(spelldata.AuraConfig(waneRow))
		parsed := spelldata.ParseEffects(character, wane, waneRow)

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1309550), func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			if sim.Proc(0.5, "Claymore of Two Moons") {
				wax.Cast(sim, result.Target)
			} else {
				wane.Activate(sim)
			}
		})
		trigger.Name = "Claymore of Two Moons"
		trigger.ActionID = core.ActionID{ItemID: 280080}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(280080, character.MakeProcTriggerAura(trigger))
		character.AddStatProcBuff(280080, &core.StatBuffAura{Aura: wane, BuffedStatTypes: parsed.Stats()}, false,
			character.ItemSwap.EligibleSlotsForItem(280080))
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
	// Measured rate (EV-0121).
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

	// Demonfork
	// https://www.wowhead.com/forever/item=12621/demonfork
	//
	// Chance on hit: Transfers 10 health every 1.0 seconds from the target to the caster for 5s.
	//
	// Measured rate (EV-0228). A proc refreshes the drain.
	newEntryRateItemEffect(12621, 16603, func(character *core.Character, ppm float64) {
		row := spelldata.MustFind(16603)
		drain := registerLeechDrain(character, row, "Demonfork")

		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			drain.Cast(sim, result.Target)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(12621, ppm, 0)
		})
		trigger.Name = "Demonfork"
		trigger.ActionID = core.ActionID{ItemID: 12621}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(12621, character.MakeProcTriggerAura(trigger))
	})

	// Malown's Slam
	// https://www.wowhead.com/forever/item=13393/malowns-slam
	//
	// Chance on hit: Knocks target silly for 2 sec and increases Strength by 53 for 20 sec.
	//
	// No effect of 17500 casts 17499, so the generator cannot pair them (EV-0150). The sim leaves the
	// stun out.
	shared.NewSpellDataAuraProc(shared.SpellDataProc{TriggerSpellID: 17500, BuffSpellID: 17499, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 13393, ItemName: "Malown's Slam"},
		})

	// Felstriker
	// https://www.wowhead.com/forever/item=12590/felstriker
	//
	// Chance on hit: All attacks are guaranteed to land and will be critical strikes for the next 3s.
	//
	// The Windfury Totem hit the proccing hit sets off rolls under it, and it lasts a batch window past
	// 3 s (EV-0225). An extra swing that hit sets off rolls under it too; that swing is not measured.
	shared.ProcTriggerExtras[12590] = func(trigger *core.ProcTrigger) {
		trigger.TriggerImmediately = true
	}
	shared.ProcBuffDurationExtras[12590] = core.SpellBatchWindow

	// Blade of Eternal Darkness
	// https://www.wowhead.com/forever/item=17780/blade-of-eternal-darkness
	//
	// Chance on landing a damaging spell to deal 112 Shadow damage and restore 104 mana to you.
	//
	// Harmful spells roll it on cast time; 21978 lacks Can Proc From Procs, so procs never do (EV-0025).
	row := spelldata.MustFind(21978)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 21978, BuffSpellID: 27860,
		Rate: func(character *core.Character, trigger *core.ProcTrigger) {
			castTimePPM(row)(character, trigger)
			harmfulSpells(character, trigger)
		}},
		[]shared.ItemVariant{
			{ItemID: 17780, ItemName: "Blade of Eternal Darkness"},
		})
	// A landed hit takes no Spell Piercing bonus (EV-0142).
	shared.DamageProcExtras[17780] = func(_ *core.Character, config *core.SpellConfig) {
		config.Flags |= core.SpellFlagNoPartialResist
	}

	// Moonshadow Stave
	// https://www.wowhead.com/forever/item=22458/moonshadow-stave
	//
	// Harmful spell casts have a chance to deal 91 Arcane or Shadow damage.
	//
	// Rate on cast time; each proc is Shadow or Arcane, one in two (EV-0115).
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
		}, castTimePPM(row), harmfulSpells)
		trigger.Name = "Moonshadow Stave"
		trigger.ActionID = core.ActionID{ItemID: 22458}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(22458, character.MakeProcTriggerAura(trigger))
	})

	// Kindling Stave
	// https://www.wowhead.com/forever/item=11750/kindling-stave
	//
	// Harmful spell casts have a chance to deal 70 Fire damage. Deals 3 times as much damage to Plants.
	//
	// Rate on cast time (EV-0219).
	core.NewItemEffect(11750, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(1300781)
		ignition := spelldata.MustFind(1300782)

		damage := ignition.DamageEffect()
		config := procSpellConfig(character, ignition)
		config.VersusBonus = shared.ItemVersusBonus(character, 11750, proto.VersusEffectKind_VersusEffectKindDamage, ignition)
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeMagicHitAndCrit)
		}
		fire := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			fire.Cast(sim, result.Target)
		}, castTimePPM(row), harmfulSpells)
		trigger.Name = "Kindling Stave"
		trigger.ActionID = core.ActionID{ItemID: 11750}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(11750, character.MakeProcTriggerAura(trigger))
	})

	// Spinal Reaper
	// https://www.wowhead.com/forever/item=288901/spinal-reaper
	//
	// Restores 150 mana or 20 rage when you kill a target that gives experience; this effect cannot
	// occur more than once every 10 seconds.
	//
	// A wearer with a rage bar and no mana bar gains the rage, and so does a druid in Bear or Dire Bear
	// Form at the killing blow; other forms gain the mana.
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
	// The row's ProcChance is the chance against Orcs, half of it against other targets (EV-0119).
	core.NewItemEffect(11684, func(agent core.Agent) {
		registerExtraAttackVersus(agent.GetCharacter(), 11684, "Ironfoe", 1301046)
	})

	// Naga Heartpiercer
	// https://www.wowhead.com/forever/item=3078/naga-heartpiercer
	//
	// Chance on hit: Wounds the target for 42 damage. Naga are Stunned for 2 sec.
	//
	// Rate (EV-0136); a missed Auto Shot never rolls it (EV-0206). It rolls and strikes at the shot's
	// release, physical on the melee table. The sim leaves the stun on Naga out.
	newEntryRateItemEffect(3078, 1292647, func(character *core.Character, ppm float64) {
		row := spelldata.MustFind(1292647)
		damage := row.DamageEffect()

		config := procSpellConfig(character, row)
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeMeleeSpecialHitAndCrit)
		}
		pierceHeart := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			pierceHeart.Cast(sim, result.Target)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.Callback = core.CallbackOnCastResult
			trigger.Outcome = core.OutcomeLanded
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(3078, ppm, 0)
		})
		trigger.Name = "Naga Heartpiercer"
		trigger.ActionID = core.ActionID{ItemID: 3078}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(3078, character.MakeProcTriggerAura(trigger))
	})

	// White Obsidian Wand
	// https://www.wowhead.com/forever/item=274425/white-obsidian-wand
	//
	// Chance on hit to reduce the mana cost of your next spell cast within 10s by 24.
	//
	// Rate on wand shots (EV-0119). A free cast keeps the charge, as does one the reduction brings to 0.
	// spelldata.AttachChargeSpender on cast complete would spend it on Shoot.
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
	// A Rage landing on a running Lash leaves it alone (EV-0136). The Trolls factor 1294430 states is the
	// Lash's alone, so the Rage hands its versus bonus over.
	shared.DamageProcExtras[9372] = func(character *core.Character, config *core.SpellConfig) {
		row := spelldata.MustFind(config.ActionID.SpellID).EffectN(4).Trigger()
		lashConfig := procSpellConfig(character, row)
		lashConfig.VersusBonus, config.VersusBonus = config.VersusBonus, nil
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

	// Barovian Family Sword
	// https://www.wowhead.com/forever/item=14541/barovian-family-sword
	//
	// Chance on hit: Deals 49 Shadow damage every 1.0 sec for 5s. All damage done is then transferred to
	// the caster.
	//
	// No hit roll, and each tick heals the amount the drain's row states (EV-0179).
	// TODO: the heal crits several times as often as the tick in the logs (EV-0179), and the sim rolls both
	// on the wearer's spell crit. Measure which chance each rolls.
	shared.DamageProcExtras[14541] = func(character *core.Character, config *core.SpellConfig) {
		row := spelldata.MustFind(config.ActionID.SpellID)
		drain := row.PeriodicDamageEffect()
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeAlwaysHit)
			spell.Dot(target).Apply(sim)
			spell.DealOutcome(sim, result)
		}

		linked := row.Triggered()
		if len(linked) != 1 || linked[0].HealEffect() == spelldata.NilEffect {
			panic(fmt.Sprintf("Barovian Family Sword: %d links to %d spells, want its heal alone", row.ID, len(linked)))
		}
		healConfig := procSpellConfigAs(character, linked[0], spelldata.Magic(core.ProcMaskSpellHealing))
		// The heal's row states no defense type, and a heal crits for the magic multiplier.
		healConfig.DefenseType = core.DefenseTypeMagic
		healConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, drain.Roll(sim), spell.OutcomeHealingCrit)
		}
		heal := character.RegisterSpell(healConfig)

		tick := config.Dot.OnTick
		config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			tick(sim, target, dot)
			heal.Cast(sim, &character.Unit)
		}
	}

	// Pendulum of Doom
	// https://www.wowhead.com/forever/item=9425/pendulum-of-doom
	//
	// Chance on hit: Delivers a fatal wound for 311 Physical damage. Deals 50% increased damage to
	// targets below 25% health.
	//
	// Diabolic Skiver
	// https://www.wowhead.com/forever/item=9475/diabolic-skiver
	//
	// Chance on hit: Delivers a fatal wound for 174 Physical damage. Deals 50% increased damage to
	// targets below 25% health.
	//
	// Gut Ripper
	// https://www.wowhead.com/forever/item=2164/gut-ripper
	//
	// Chance on hit: Delivers a fatal wound for 105 Physical damage. Deals 100% increased damage to
	// targets below 25% health.
	//
	// Blackhand Doomsaw
	// https://www.wowhead.com/forever/item=12583/blackhand-doomsaw
	//
	// Chance on hit: Delivers a fatal wound for 448 Physical damage. Deals 50% increased damage to
	// targets below 25% health.
	//
	// The health threshold is the row's alone (EV-0136).
	fatalWound := func(_ *core.Character, config *core.SpellConfig) {
		row := spelldata.MustFind(config.ActionID.SpellID)
		bonus := 1 + row.EffectN(2).Average()/100
		health := int32(row.EffectN(3).Average())
		switch health {
		case 35, 25, 20:
		default:
			panic(fmt.Sprintf("Fatal Wound: %d's third effect states %d%% health, which starts no execute phase", row.ID, health))
		}
		deal := config.ApplyEffects
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			multiplier := spell.DamageMultiplier
			if target.IsBelowExecuteHealth(health) {
				spell.DamageMultiplier *= bonus
			}
			deal(sim, target, spell)
			spell.DamageMultiplier = multiplier
		}
	}
	shared.DamageProcExtras[9425] = fatalWound
	shared.DamageProcExtras[9475] = fatalWound
	shared.DamageProcExtras[2164] = fatalWound
	shared.DamageProcExtras[12583] = fatalWound

	// Barman Shanker
	// https://www.wowhead.com/forever/item=12791/barman-shanker
	//
	// Chance on hit: Wounds the target causing them to bleed for 98 damage over 7 sec. Deals 2 times as
	// much damage if applied while behind the target.
	//
	// Effect 2 multiplies a Rend applied from behind (EV-0146); its ticks meet a tenth of armor (EV-0194).
	shared.DamageProcExtras[12791] = func(character *core.Character, config *core.SpellConfig) {
		if !character.PseudoStats.InFrontOfTarget {
			config.DamageMultiplier *= spelldata.MustFind(config.ActionID.SpellID).EffectN(2).Average()
		}
	}

	// Toofbreaker
	// https://www.wowhead.com/forever/item=286430/toofbreaker
	//
	// Chance on hit: Bash the target, causing 42 Physical damage. Deals 2 times as much damage if the
	// target is facing you.
	//
	// The target faces a wearer who stands in front of it, by the owner's ruling.
	shared.DamageProcExtras[286430] = func(character *core.Character, config *core.SpellConfig) {
		if character.PseudoStats.InFrontOfTarget {
			config.DamageMultiplier *= spelldata.MustFind(config.ActionID.SpellID).EffectN(2).Average()
		}
	}

	// Flame Wrath
	// https://www.wowhead.com/forever/item=11809/flame-wrath
	//
	// Chance on hit: Shoots a ring of fire dealing 154 Fire damage to all nearby enemies, and envelops
	// the caster with a Fire Shield for 15s.
	//
	// The generated spell does not cast 16559's second effect, so each proc puts the shield on here
	// (EV-0136). A druid's form attacks roll it as any other melee hit does (R-0056).
	shared.DamageProcExtras[11809] = func(character *core.Character, config *core.SpellConfig) {
		row := spelldata.MustFind(config.ActionID.SpellID).EffectN(2).Trigger()
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
	// Each proc spends all stacks but one (EV-0136).
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
	// The stun rolls its own hit, so an avoided strike stuns too (EV-0136).
	shared.DamageProcExtras[9477] = func(character *core.Character, config *core.SpellConfig) {
		row := spelldata.MustFind(1294340)
		label := fmt.Sprintf("%s %d", row.Name, row.ID)
		versus := config.VersusBonus
		stuns := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			if aura := target.GetAura(label); aura != nil {
				return aura
			}
			return target.RegisterStunAura(label, core.ActionID{SpellID: row.ID}, time.Duration(float64(row.Duration())*versus.Against(target)))
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
	// The generator keeps one proc per item, so both are written here: 1298414 on every landed main-hand
	// hit (EV-0195), which rolls crit alone (EV-0164), and 1298416 at its rate (EV-0136). The sim leaves the
	// slow out.
	core.NewItemEffect(13952, func(agent core.Agent) {
		character := agent.GetCharacter()

		chillRow := spelldata.MustFind(1298414)
		chill := chillRow.DamageEffect()
		chillConfig := spelldata.SpellConfig(&character.Unit, chillRow, spelldata.Melee(core.ProcMaskEmpty), spelldata.Proc(), spelldata.Triggered())
		var chillOutcome core.OutcomeApplier
		chillConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.DealDamageAfterTravel(sim, spell.CalcDamage(sim, target, chill.DirectAmount(sim), chillOutcome))
		}
		chillSpell := character.RegisterSpell(chillConfig)
		chillOutcome = chillSpell.OutcomeMeleeSpecialHitAndCrit
		if chillRow.AlwaysHits() {
			chillOutcome = chillSpell.OutcomeMeleeSpecialCritOnly
		}

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
	// Rate on wand shots; the mana is that of 1302350, which 1302348 fires, not the tooltip's 65 (EV-0136).
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

// Registers an item effect whose proc runs at the procs per minute the item's chance-on-hit entry states
// for spellID, since the game files state none; while the entry states none it registers nothing. The
// rate goes in overrides.WeaponProcs, never in a Spells PPM override: the entry carries only the first,
// while the generator's missing-effect check counts either as a rate.
func newEntryRateItemEffect(itemID int32, spellID int32, apply func(character *core.Character, ppm float64)) {
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

// An item's extra attack proc whose ProcChance column is the chance against its versus targets, and
// whose versus bonus is how many times likelier a hit on them is than one on any other target.
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

// Narrows the trigger to the wearer's harmful spells landing on an enemy: spell damage hits, and magic
// spells with no proc mask whose school is not Physical, such as the curses (EV-0025, EV-0219).
func harmfulSpells(character *core.Character, trigger *core.ProcTrigger) {
	condition := trigger.ExtraCondition
	trigger.ProcMask = core.ProcMaskUnknown
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
		harmful := spell.ProcMask.Matches(core.ProcMaskSpellDamage) ||
			(spell.ProcMask == core.ProcMaskEmpty && spell.DefenseType == core.DefenseTypeMagic &&
				spell.SpellSchool != core.SpellSchoolPhysical && !spell.Flags.Matches(core.SpellFlagPassiveSpell) &&
				spell.ActionID.ItemID == 0)
		return harmful && character.IsOpponent(result.Target) && (condition == nil || condition(sim, spell, result))
	}
}

// The row's procs per minute on the cast time behind each landed hit, at least 1.5 s (EV-0025, EV-0219).
// The cast time is read where the cast deals its result, a missile's at cast end, so one still in flight
// keeps it when the next cast of the spell starts. The chance depends on the hit, so Chance(1) leaves
// core no roll and the condition draws it after the cooldown check and before the cooldown starts.
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
		trigger.RequireDamageDealt = false
		flags := trigger.ProcSpellClassFlags
		trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			castTime, ok := castTimes[result]
			if !ok {
				castTime = spell.CurCast.CastTime
			}
			delete(castTimes, result)
			return spell.DefenseType != core.DefenseTypeNone && shared.RollCastTimePPM(sim, character, row, flags, castTime)
		}
		spelldata.Chance(1)(character, trigger)
	}
}
