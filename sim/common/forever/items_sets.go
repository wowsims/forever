package forever

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func arathiLeatherSet(id int32, name string) *core.ItemSet {
	return core.NewItemSet(core.ItemSet{
		ID: id, Name: name,
		Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet},
		Bonuses: map[int32]core.ApplySetBonus{
			2: shared.ParsedSetBonus(spelldata.MustFind(7503)),
			3: shared.ParsedSetBonus(spelldata.MustFind(7597)),
		},
	})
}

// Arathi Basin reputation - Horde leather
// https://www.wowhead.com/forever/item-set=486/the-defilers-purpose
var ItemSetDefilersPurpose = arathiLeatherSet(486, "The Defiler's Purpose")

// Arathi Basin reputation - Alliance leather
// https://www.wowhead.com/forever/item-set=471/the-highlanders-purpose
var ItemSetHighlandersPurpose = arathiLeatherSet(471, "The Highlander's Purpose")

// Blacksmithing - Plate
// https://www.wowhead.com/forever/item-set=321/imperial-plate
//
// The 4-piece bonus keeps movement speed from dropping below 80% and has no place in the sim.
var ItemSetImperialPlate = core.NewItemSet(core.ItemSet{
	ID:    321,
	Name:  "Imperial Plate",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increased Defense +7.
			setBonusAura.
				AttachStatBuff(stats.DefenseRating, 7*core.DefenseRatingPerDefenseLevel).
				ExposeToAPL(13385)
		},
		3: func(_ core.Agent, setBonusAura *core.Aura) {
			// Improves your chance to hit by 1%.
			setBonusAura.
				AttachStatsBuff(stats.Stats{stats.PhysicalHitPercent: 1, stats.SpellHitPercent: 1}).
				ExposeToAPL(1251990)
		},
		5: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increased Strength +20.
			setBonusAura.
				AttachStatBuff(stats.Strength, 20).
				ExposeToAPL(1251984)
		},
		6: func(_ core.Agent, setBonusAura *core.Aura) {
			// Improves your chance to get a critical strike by 1%.
			setBonusAura.
				AttachStatsBuff(stats.Stats{stats.PhysicalCritPercent: 1, stats.SpellCritPercent: 1}).
				ExposeToAPL(1251991)
		},
	},
})

// Blacksmithing - Plate
// https://www.wowhead.com/forever/item-set=1968/blessed-plate
//
// The 4-piece bonus restores 20 health every 5 sec to party members within 10 yards; the sim has no
// health regeneration stat, so it is left out.
var ItemSetBlessedPlate = core.NewItemSet(core.ItemSet{
	ID:    1968,
	Name:  "Blessed Plate",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increased Shadow Resistance +10.
			setBonusAura.
				AttachStatBuff(stats.ShadowResistance, 10).
				ExposeToAPL(14673)
		},
		3: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increased Strength +20.
			setBonusAura.
				AttachStatBuff(stats.Strength, 20).
				ExposeToAPL(1252009)
		},
		5: func(_ core.Agent, setBonusAura *core.Aura) {
			// Improves your chance to hit by 2%.
			setBonusAura.
				AttachStatsBuff(stats.Stats{stats.PhysicalHitPercent: 2, stats.SpellHitPercent: 2}).
				ExposeToAPL(1252013)
		},
		6: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increases damage done by Holy spells and effects by up to 29.
			setBonusAura.
				AttachStatBuff(stats.HolyDamage, 29).
				ExposeToAPL(21518)
		},
	},
})

// The rows the dungeon sets share.
var (
	dungeonResistances = spelldata.MustFind(18679)  // 2pc: +8 all resistances
	dungeonSpellDamage = spelldata.MustFind(9346)   // Dungeon Set 1 3pc: +18 spell damage
	dungeonManaRegen   = spelldata.MustFind(18378)  // Dungeon Set 1 6pc: 8 mana per 5 sec
	dungeonSuddenMana  = spelldata.MustFind(450554) // Dungeon Set 1 5pc: Sudden Insight, the mana Magister's and Devout procs restore
	dungeonTwoDamage   = spelldata.MustFind(14047)  // Dungeon Set 2 cloth 6pc: +23 spell damage and healing
	dungeonTwoAttack   = spelldata.MustFind(14049)  // Dungeon Set 2 6pc: +40 attack power
)

var (
	magisterManaProc     = spelldata.MustFind(450527) // Magister's 5pc: Mana Proc on Cast, the proc
	beaststalkerManaProc = spelldata.MustFind(450577) // Beaststalker 5pc: the proc on melee and ranged autoattacks
)

// The Sudden Insight mana proc of Magister's Regalia, Beaststalker Armor and their reissues.
func suddenInsightMana(proc *spelldata.Spell, name string) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		character := agent.GetCharacter()
		manaMetrics := character.NewManaMetrics(core.ActionID{SpellID: dungeonSuddenMana.ID})
		mana := dungeonSuddenMana.EnergizeEffect().Average()

		// The rate is the proc chance column with no cooldown, rolled on whatever the row's mask hears.
		trigger := spelldata.ProcTrigger(character, proc,
			func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				character.AddMana(sim, mana, manaMetrics)
			})
		trigger.Name = name

		setBonusAura.AttachProcTrigger(trigger)
	}
}

// (2) Set: all resistances, and (4) Set: a chance to freeze an attacker when struck, which a boss
// fight does not test.
var ItemSetMagistersRegalia = core.NewItemSet(core.ItemSet{
	Name:  "Magister's Regalia",
	ID:    181,
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		3: shared.ParsedSetBonus(dungeonSpellDamage),
		5: suddenInsightMana(magisterManaProc, "Magister's Regalia - 5PC"),
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 2, Season of Discovery reissue
//
// The 5-piece bonus (27867) freezes an attacker when struck, which a boss fight does not test.
var ItemSetSorcerersRegaliaReissue = core.NewItemSet(core.ItemSet{
	ID:    1671,
	Name:  "Sorcerer's Regalia",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonManaRegen),
		4: suddenInsightMana(magisterManaProc, "Sorcerer's Regalia - 4PC"),
		6: shared.ParsedSetBonus(dungeonTwoDamage),
	},
})

var devoutManaProc = spelldata.MustFind(450576) // 5pc: the proc

// The mana proc of Vestments of the Devout and of the Vestments of the Virtuous reissue.
func devoutManaRestore(name string) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		// Your spellcasts have a 5% chance to energize you for 200 mana.
		character := agent.GetCharacter()
		manaMetrics := character.NewManaMetrics(core.ActionID{SpellID: dungeonSuddenMana.ID})
		mana := dungeonSuddenMana.EnergizeEffect().Average()

		trigger := spelldata.ProcTrigger(character, devoutManaProc,
			func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				character.AddMana(sim, mana, manaMetrics)
			})
		trigger.Name = name

		setBonusAura.AttachProcTrigger(trigger)
		setBonusAura.ExposeToAPL(devoutManaProc.ID)
	}
}

// Dungeon Set 1
// https://www.wowhead.com/forever/item-set=182/vestments-of-the-devout
//
// The 4-piece bonus (27778) shields the wearer when struck, which the sim does not model.
var ItemSetVestmentsOfTheDevout = core.NewItemSet(core.ItemSet{
	ID:    182,
	Name:  "Vestments of the Devout",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonSpellDamage),
		5: devoutManaRestore("Vestments of the Devout - 5PC"),
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 2, Season of Discovery reissue
//
// The 5-piece bonus (27778) shields the wearer when struck, which the sim does not model.
var ItemSetVestmentsOfTheVirtuousReissue = core.NewItemSet(core.ItemSet{
	ID:    1675,
	Name:  "Vestments of the Virtuous",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonManaRegen),
		4: devoutManaRestore("Vestments of the Virtuous - 4PC"),
		6: shared.ParsedSetBonus(dungeonTwoDamage),
	},
})

var (
	dreadmistHealProc = spelldata.MustFind(450585) // 5pc: the proc
	dreadmistHeal     = spelldata.MustFind(450583) // 5pc: Dark Reward, the heal it casts
)

// The heal proc of Dreadmist Raiment and of the Deathmist Raiment reissue.
func dreadmistDarkReward(name string) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		character := agent.GetCharacter()
		heal := dreadmistHeal.HealEffect()

		config := spelldata.SpellConfig(&character.Unit, dreadmistHeal, spelldata.Magic(core.ProcMaskSpellHealing), spelldata.Proc(), spelldata.Triggered())
		// The row states no defense type, and a heal crits for the magic multiplier.
		config.DefenseType = core.DefenseTypeMagic
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, heal.DirectAmount(sim), spell.OutcomeHealingCrit)
		}
		darkReward := character.RegisterSpell(config)

		// The proc chance column with no cooldown, rolled on every spell cast.
		trigger := spelldata.ProcTrigger(character, dreadmistHealProc,
			func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				darkReward.Cast(sim, &character.Unit)
			})
		trigger.Name = name

		setBonusAura.AttachProcTrigger(trigger)
	}
}

// Dungeon Set 1
var ItemSetDreadmistRaiment = core.NewItemSet(core.ItemSet{
	ID:    183,
	Name:  "Dreadmist Raiment",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonSpellDamage),
		// TODO: 4pc (27780) fears an attacker when struck, which the sim does not model.
		5: dreadmistDarkReward("Dreadmist Raiment - 5PC"),
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 2, Season of Discovery reissue
var ItemSetDeathmistRaimentReissue = core.NewItemSet(core.ItemSet{
	ID:    1681,
	Name:  "Deathmist Raiment",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonManaRegen),
		4: dreadmistDarkReward("Deathmist Raiment - 4PC"),
		// TODO: 5pc (27780) fears an attacker when struck, which the sim does not model.
		6: shared.ParsedSetBonus(dungeonTwoDamage),
	},
})

var (
	wildheartProc    = spelldata.MustFind(450608)  // 5pc: the proc
	wildheartMana    = spelldata.MustFind(27782)   // 5pc: the mana a spell cast restores
	wildheartEnergy  = spelldata.MustFind(27784)   // 5pc: the energy over 5 s a landed melee swing restores
	wildheartRage    = spelldata.MustFind(27783)   // 5pc: the rage a landed melee swing taken restores
	shiftingPower    = spelldata.MustFind(1322605) // 5pc: a mana-costing cast that rolls nothing
	wildHeartTrigger = spelldata.MustFind(1302708) // 4pc: Wild Heart, the proc
	wildHeartBuff    = spelldata.MustFind(1302707) // 4pc: the movement speed it gives
)

// The energize proc of Wildheart Raiment and of the Feralheart Raiment reissue; the dummy effect's 5 is
// no chance.
// PPM measurement in EV-0144.
// Which casts roll in EV-0144.
func wildheartEnergize(name string) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		character := agent.GetCharacter()
		mana := wildheartMana.EnergizeEffect().Average()
		manaMetrics := character.NewManaMetrics(core.ActionID{SpellID: wildheartMana.ID})

		trigger := shared.SpellCastPPMTrigger(character, wildheartProc,
			func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				character.AddMana(sim, mana, manaMetrics)
			})
		trigger.Name = name
		roll := trigger.ExtraCondition
		trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			row := spelldata.Find(spell.SpellID)
			spendsRageOrEnergy := len(row.Powers) > 0 && (row.Powers[0].Type == dbcenums.POWER_RAGE || row.Powers[0].Type == dbcenums.POWER_ENERGY)
			return row.ShapeshiftForm() == 0 && !spendsRageOrEnergy && row.ID != shiftingPower.ID && roll(sim, spell, result)
		}

		setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(wildheartProc.ID)
		attachWildheartSwingEnergy(character, setBonusAura, name)
		attachWildheartStruckRage(character, setBonusAura, name)
	}
}

// The energy listener.
// Swings and ticks in EV-0144.
// TODO: The manager reads the main hand's speed when it is built, so a druid that shifts between Cat Form
// and a bear form keeps the paw of the form it started in.
func attachWildheartSwingEnergy(character *core.Character, setBonusAura *core.Aura, name string) {
	if !character.HasEnergyBar() {
		return
	}
	tick := wildheartEnergy.ProcEnergizeEffect()
	amount := tick.Average()
	metrics := character.NewEnergyMetrics(core.ActionID{SpellID: wildheartEnergy.ID})

	config := procSpellConfig(character, wildheartEnergy)
	config.Hot = spelldata.DotConfig(wildheartEnergy, tick)
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
		if character.GetCurrentPowerBar() != core.ManaBar {
			character.AddEnergy(sim, amount, metrics)
		}
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.SelfHot().Apply(sim)
	}
	energy := character.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(character, wildheartProc,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			energy.Cast(sim, &character.Unit)
		},
		landedHits(core.ProcMaskMeleeWhiteHit))
	trigger.Name = name + " - Swing"
	setBonusAura.AttachProcTrigger(trigger)
}

// The rage listener: a melee ability taken never rolls the proc, though the row names it (EV-0144). The
// rage fills the bar only while rage is the current one, since a shift into a bear form empties the bar.
func attachWildheartStruckRage(character *core.Character, setBonusAura *core.Aura, name string) {
	if !character.HasRageBar() {
		return
	}
	rage := wildheartRage.EnergizeEffect().Tenths()
	metrics := character.NewRageMetrics(core.ActionID{SpellID: wildheartRage.ID})

	trigger := spelldata.ProcTrigger(character, wildheartProc,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			if character.GetCurrentPowerBar() == core.RageBar {
				character.AddRage(sim, rage, metrics)
			}
		},
		func(_ *core.Character, trigger *core.ProcTrigger) {
			trigger.Callback = core.CallbackOnSpellHitTaken
			trigger.ProcMask = core.ProcMaskMeleeWhiteHit
			trigger.DPM = character.NewAttackerSwingPPMManager(float64(wildheartProc.RPPM), trigger.ProcMask,
				func(*core.Unit) float64 { return 1 })
		})
	trigger.Name = name + " - Struck"
	setBonusAura.AttachProcTrigger(trigger)
}

// Wild Heart, Wildheart Raiment's 4-piece and the Feralheart Raiment reissue's 5-piece: when struck in
// combat, has a 5% chance to increase your movement speed by 10% for 8 sec. Can only occur once every
// 60 sec. A landed hit taken rolls the row's chance outside its cooldown (EV-0144). The spell parse
// leaves the buff's speed aura out, so the buff multiplies the movement speed itself.
func wildHeart(name string) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		character := agent.GetCharacter()
		speed := 1 + wildHeartBuff.FirstAura(dbcenums.A_MOD_SPEED_ALWAYS).Percent()
		buff := character.RegisterAura(spelldata.AuraConfig(wildHeartBuff))
		buff.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
			character.MultiplyMovementSpeed(sim, speed)
		})
		buff.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
			character.MultiplyMovementSpeed(sim, 1/speed)
		})

		trigger := spelldata.ProcTrigger(character, wildHeartTrigger, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Activate(sim)
		})
		trigger.Name = name
		setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(wildHeartTrigger.ID)
	}
}

// Dungeon Set 1
// https://www.wowhead.com/forever/item-set=185/wildheart-raiment
var ItemSetWildheartRaiment = core.NewItemSet(core.ItemSet{
	ID:    185,
	Name:  "Wildheart Raiment",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonSpellDamage, spelldata.MustFind(9336)), // +18 spell damage, +30 attack power
		4: wildHeart("Wildheart Raiment - 4PC"),
		// TODO: Ingame test needed -- whether a rogue gains the energy and a warrior the rage.
		5: wildheartEnergize("Wildheart Raiment - 5PC"),
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 1
// https://www.wowhead.com/forever/item-set=186/beaststalker-armor
//
// The 4-piece bonus removes a root when struck, which has no place in the sim.
var ItemSetBeaststalkerArmor = core.NewItemSet(core.ItemSet{
	ID:    186,
	Name:  "Beaststalker Armor",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(spelldata.MustFind(9336)), // +30 attack power
		5: suddenInsightMana(beaststalkerManaProc, "Beaststalker Armor - 5PC"),
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 1
// https://www.wowhead.com/forever/item-set=187/the-elements
//
// PPM measurement in EV-0144.
// The 4-piece bonus disarms an attacker when struck, which a boss fight does not test.
var ItemSetTheElements = core.NewItemSet(core.ItemSet{
	ID:    187,
	Name:  "The Elements",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonSpellDamage),
		5: shared.ParsedSetBonus(spelldata.MustFind(450626)), // The Furious Storm
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 1
// https://www.wowhead.com/forever/item-set=188/lightforge-armor
//
// PPM measurement in EV-0144.
// The 4-piece bonus silences a caster when struck by a spell, which a boss fight does not test.
var ItemSetLightforgeArmor = core.NewItemSet(core.ItemSet{
	ID:    188,
	Name:  "Lightforge Armor",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonSpellDamage),
		5: shared.ParsedSetBonus(spelldata.MustFind(450625)), // Crusader's Wrath
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 1
// https://www.wowhead.com/forever/item-set=189/battlegear-of-valor
//
// The 4-piece bonus removes a disarm when struck and the 6-piece bonus restores health over time,
// neither of which has a place in the sim.
var ItemSetBattlegearOfValor = core.NewItemSet(core.ItemSet{
	ID:    189,
	Name:  "Battlegear of Valor",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(spelldata.MustFind(13198)), // +15 Strength
		5: warriorsResolve("Battlegear of Valor - 5PC"),
	},
})

var (
	warriorsResolveProc = spelldata.MustFind(450587) // Valor 5pc, Heroism reissue 4pc: the proc
	warriorsResolveHeal = spelldata.MustFind(450589) // Warrior's Resolve, the heal and the rage it casts
)

// The proc of Battlegear of Valor and of the Battlegear of Heroism reissue. Heroic Strike and Cleave
// never roll it (EV-0136).
func warriorsResolve(name string) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		character := agent.GetCharacter()
		heal := warriorsResolveHeal.HealEffect()

		config := spelldata.SpellConfig(&character.Unit, warriorsResolveHeal, spelldata.Magic(core.ProcMaskSpellHealing), spelldata.Proc(), spelldata.Triggered())
		// The row states no defense type, and a heal crits for the magic multiplier.
		config.DefenseType = core.DefenseTypeMagic
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, heal.DirectAmount(sim), spell.OutcomeHealingCrit)
		}
		resolve := character.RegisterSpell(config)
		rage := shared.FiredSpellHandler(character, core.ActionID{SpellID: warriorsResolveHeal.ID}, warriorsResolveProc)

		trigger := spelldata.ProcTrigger(character, warriorsResolveProc, func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			resolve.Cast(sim, &character.Unit)
			if rage != nil {
				rage(sim, spell, result)
			}
		})
		trigger.Name = name
		setBonusAura.AttachProcTrigger(trigger)
	}
}

// Dungeon Set 2, Season of Discovery reissue
// https://www.wowhead.com/forever/item-set=1667/feralheart-raiment
//
// The four groups of pieces count toward the same bonuses (EV-0144).
var ItemSetFeralheartRaimentReissue = core.NewItemSet(core.ItemSet{
	ID:    1667,
	Name:  "Feralheart Raiment",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonManaRegen),
		4: wildheartEnergize("Feralheart Raiment - 4PC"),
		5: wildHeart("Feralheart Raiment - 5PC"),
		6: shared.ParsedSetBonus(dungeonTwoAttack, dungeonTwoDamage),
	},
})

var (
	rogueArmorEnergyProc = spelldata.MustFind(27787) // Shadowcraft 5pc, Darkmantle 4pc: Rogue Armor Energize, the proc
	rogueArmorEnergy     = spelldata.MustFind(27788) // the energy it restores
)

// The energy proc of the two rogue dungeon sets.
// PPM measurement in EV-0144.
// A Windfury Totem hit rolls it as a main-hand swing (EV-0136).
func rogueArmorEnergize(name string) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		character := agent.GetCharacter()
		gain := shared.FiredSpellHandler(character, core.ActionID{SpellID: rogueArmorEnergy.ID}, rogueArmorEnergyProc)
		if gain == nil {
			return
		}

		trigger := spelldata.ProcTrigger(character, rogueArmorEnergyProc, gain)
		trigger.Name = name
		setBonusAura.AttachProcTrigger(trigger)
	}
}

// Dungeon Set 1
// https://www.wowhead.com/forever/item-set=184/shadowcraft-armor
//
// The 4-piece bonus removes a snare when struck and the 6-piece bonus restores health over time,
// neither of which has a place in the sim.
var ItemSetShadowcraftArmor = core.NewItemSet(core.ItemSet{
	ID:    184,
	Name:  "Shadowcraft Armor",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(spelldata.MustFind(9336)), // +30 attack power
		5: rogueArmorEnergize("Shadowcraft 5-piece Proc"),
	},
})

// Dungeon Set 2, Season of Discovery reissue
//
// The 3-piece bonus restores health over time and the 5-piece bonus removes a snare when struck,
// neither of which has a place in the sim.
var ItemSetDarkmantleArmorReissue = core.NewItemSet(core.ItemSet{
	ID:    1676,
	Name:  "Darkmantle Armor",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		4: rogueArmorEnergize("Darkmantle 4-piece Proc"),
		6: shared.ParsedSetBonus(dungeonTwoAttack),
	},
})

// Dungeon Set 2, Season of Discovery reissue
//
// The 5-piece bonus removes a root when struck, which has no place in the sim.
var ItemSetBeastmasterArmorReissue = core.NewItemSet(core.ItemSet{
	ID:    1669,
	Name:  "Beastmaster Armor",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonManaRegen),
		4: suddenInsightMana(beaststalkerManaProc, "Beastmaster Armor - 4PC"),
		6: shared.ParsedSetBonus(dungeonTwoAttack),
	},
})

// Dungeon Set 2, Season of Discovery reissue
//
// PPM measurement in EV-0144.
// The 4-piece bonus is Lightforge Armor's 5-piece proc (450625). The 5-piece bonus silences a caster when
// struck by a spell, which a boss fight does not test.
var ItemSetSoulforgeArmorReissue = core.NewItemSet(core.ItemSet{
	ID:    1673,
	Name:  "Soulforge Armor",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonManaRegen),
		4: shared.ParsedSetBonus(spelldata.MustFind(450625)), // Crusader's Wrath
		6: shared.ParsedSetBonus(dungeonTwoDamage),
	},
})

// Dungeon Set 2, Season of Discovery reissue
//
// PPM measurement in EV-0144.
// The 4-piece bonus is The Elements' 5-piece proc (450626). The 5-piece bonus disarms an attacker when
// struck, which a boss fight does not test.
var ItemSetTheFiveThundersReissue = core.NewItemSet(core.ItemSet{
	ID:    1679,
	Name:  "The Five Thunders",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonManaRegen),
		4: shared.ParsedSetBonus(spelldata.MustFind(450626)), // The Furious Storm
		6: shared.ParsedSetBonus(dungeonTwoDamage),
	},
})

// Dungeon Set 2, Season of Discovery reissue
//
// PPM measurement in EV-0144.
// The 4-piece bonus is Battlegear of Valor's 5-piece proc (450587). The 3-piece bonus restores health over
// time and the 5-piece bonus removes a disarm when struck, neither of which has a place in the sim.
var ItemSetBattlegearOfHeroismReissue = core.NewItemSet(core.ItemSet{
	ID:    1778,
	Name:  "Battlegear of Heroism",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		4: warriorsResolve("Battlegear of Heroism - 4PC"),
		6: shared.ParsedSetBonus(spelldata.MustFind(14492)), // +20 Strength
	},
})

// Scholomance - Leather
// https://www.wowhead.com/forever/item-set=121/cadaverous-garb
//
// The 3-piece bonus also shortens movement impairing effects on the wearer (1292546), which has no
// place in the sim.
var ItemSetCadaverousGarb = core.NewItemSet(core.ItemSet{
	ID:   121,
	Name: "Cadaverous Garb",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist,
		proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet},
	Bonuses: map[int32]core.ApplySetBonus{
		3: shared.ParsedSetBonus(spelldata.MustFind(9140)),   // +10 attack power
		4: shared.ParsedSetBonus(spelldata.MustFind(18676)),  // +5 all resistances
		5: shared.ParsedSetBonus(spelldata.MustFind(460230)), // 2% hit
	},
})

var (
	kroldokResolveProc = spelldata.MustFind(1324298) // 4pc: the proc
	kroldokResolveHeal = spelldata.MustFind(1324296) // 4pc: Krol'dok Resolve, the heal it casts
	ogreDominanceProc  = spelldata.MustFind(1324309) // 5pc: the proc
	ogreDominanceBuff  = spelldata.MustFind(1324308) // 5pc: Ogre Dominance, the Stamina it stacks
)

// Krol'dok Stronghold - Plate
// https://www.wowhead.com/forever/item-set=2135/krol-dok-battlegear
//
// (4) Set: Chance on block to heal for 18. Healing is doubled in Strongholds and Cities.
// (5) Set: Taking damage has a high chance to increase your Stamina by 1 for 30s, stacking up to 20 times.
//
// Heals and stacks in EV-0160.
// The 5-piece buff also makes the wearer larger and turns them into an ogre at 20 stacks, which has no
// place in the sim.
var ItemSetKroldokBattlegear = core.NewItemSet(core.ItemSet{
	ID:   2135,
	Name: "Krol'dok Battlegear",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(1324287)), // +100 armor
		3: shared.ParsedSetBonus(spelldata.MustFind(1324288)), // +5 defense
		4: kroldokResolve,
		5: ogreDominance,
	},
})

func kroldokResolve(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	config := spelldata.SpellConfig(&character.Unit, kroldokResolveHeal, spelldata.Magic(core.ProcMaskSpellHealing), spelldata.Proc(), spelldata.Triggered())
	amount := kroldokResolveHeal.HealEffect().DirectAmount
	if character.Env.Encounter.InArea(proto.AreaType_AreaTypeStrongholdsCities) {
		stronghold := kroldokResolveHeal.EffectN(2).BasePoints
		amount = func(*core.Simulation) float64 { return stronghold }
		config.DirectVariance = 0
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealHealing(sim, target, amount(sim), spell.OutcomeHealingCrit)
	}
	resolve := character.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(character, kroldokResolveProc, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		resolve.Cast(sim, &character.Unit)
	})
	trigger.Name = "Krol'dok Battlegear - 4PC"
	setBonusAura.AttachProcTrigger(trigger)
}

func ogreDominance(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	buff := character.RegisterAura(spelldata.AuraConfig(ogreDominanceBuff))
	spelldata.ParseEffects(character, buff, ogreDominanceBuff)

	trigger := spelldata.ProcTrigger(character, ogreDominanceProc, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.Activate(sim)
		buff.AddStack(sim)
	})
	trigger.Name = "Krol'dok Battlegear - 5PC"
	trigger.RequireDamageDealt = false
	setBonusAura.AttachProcTrigger(trigger)
}

var furbolgTeachings = spelldata.MustFind(1294087)

// https://www.wowhead.com/forever/item-set=2071/teachings-of-the-furbolgs
//
// (2) Set: Your attacks have a chance to increase your Attack Power by 40 and your healing spells have a
// chance to increase your Intellect by 20 for 15s.
//
// No effect of 1294087 casts the healing half, Wisdom of the Timbermaw 1294070, so a heal procs nothing;
// that row also states Strength where the tooltip says Intellect.
var ItemSetTeachingsOfTheFurbolgs = core.NewItemSet(core.ItemSet{
	ID:    2071,
	Name:  "Teachings of the Furbolgs",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2},
	Bonuses: map[int32]core.ApplySetBonus{
		// TODO: Ingame test needed -- which melee and ranged hits roll it.
		2: shared.ParsedSetBonus(furbolgTeachings),
	},
})

var (
	deviousStrikeProc  = spelldata.MustFind(1292028) // 4pc: the proc
	deviousStrikeBleed = spelldata.MustFind(1292029) // 4pc: Devious Strike, the bleed it casts
)

// https://www.wowhead.com/forever/item-set=161/defias-leather
//
// (4) Set: Grants your melee attacks a 5% chance when striking from behind to hit a vital point, causing
// the target to Bleed for 75 Physical damage over 5s.
var ItemSetDefiasLeather = core.NewItemSet(core.ItemSet{
	ID:   161,
	Name: "Defias Leather",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist,
		proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(14707)),   // +5 Arcane Resistance
		3: shared.ParsedSetBonus(spelldata.MustFind(1292025)), // +15 attack power against Humanoids
		4: deviousStrike,
		5: shared.ParsedSetBonus(spelldata.MustFind(7534)), // +1 Daggers
	},
})

func deviousStrike(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	config := procSpellConfigAs(character, deviousStrikeBleed, spelldata.Melee(core.ProcMaskEmpty))
	// Armor on ticks in EV-0194.
	config.Flags &^= core.SpellFlagTenthResistance | core.SpellFlagApplyArmorReduction
	config.Dot = spelldata.DotConfig(deviousStrikeBleed, deviousStrikeBleed.PeriodicDamageEffect())
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}
	bleed := character.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(character, deviousStrikeProc, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		bleed.Cast(sim, result.Target)
	})
	trigger.Name = "Defias Leather - 4PC"
	trigger.ExtraCondition = func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
		return !character.PseudoStats.InFrontOfTarget
	}
	setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(deviousStrikeProc.ID)
}

var (
	enragingLightProc = spelldata.MustFind(1293674) // 4pc: the proc
	enragingLightBuff = spelldata.MustFind(1293673) // 4pc: Enraging Light, the buff it puts up
	enragedLight      = spelldata.MustFind(1293675) // 4pc: Enraged Light, the Holy damage the buff deals
)

// https://www.wowhead.com/forever/item-set=163/chain-of-the-scarlet-crusade
//
// (4) Set: Taking damage has a 15% chance to grant Enraging Light for 10s, causing your melee attacks to
// deal 20 Holy damage. Deals 3 times as much damage to Undead.
//
// Triggering hits in EV-0173.
var ItemSetChainOfTheScarletCrusade = core.NewItemSet(core.ItemSet{
	ID:   163,
	Name: "Chain of the Scarlet Crusade",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotHands,
		proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(14673)), // +10 Shadow Resistance
		3: shared.ParsedSetBonus(spelldata.MustFind(18074)), // +30 attack power against Undead
		4: enragingLight,
		5: shared.ParsedSetBonus(spelldata.MustFind(432639)), // 1% hit
		// TODO: 6pc (1293670) shields the wearer when a hit leaves them at or below 20% health, which the
		// sim does not model.
	},
})

func enragingLight(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	damage := enragedLight.DamageEffect()
	config := procSpellConfigAs(character, enragedLight, spelldata.Magic(core.ProcMaskSpellDamage))
	if versus, ok := enragingLightProc.VersusBonus(); ok {
		config.VersusBonus = &versus
	}
	var outcome core.OutcomeApplier
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), outcome)
	}
	light := character.RegisterSpell(config)
	// The row states Always Hit: the damage never misses and can still crit.
	outcome = light.OutcomeMagicHitAndCrit
	if enragedLight.AlwaysHits() {
		outcome = light.OutcomeMagicCrit
	}

	buff := character.RegisterAura(spelldata.AuraConfig(enragingLightBuff))
	strike := spelldata.ProcTrigger(character, enragingLightBuff, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		light.Cast(sim, result.Target)
	})
	strike.Name = "Enraging Light - Strike"
	buff.AttachProcTrigger(strike)

	trigger := spelldata.ProcTrigger(character, enragingLightProc, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.Activate(sim)
	})
	trigger.Name = "Chain of the Scarlet Crusade - 4PC"
	setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(enragingLightProc.ID)
}

var deathboneExpertise = spelldata.MustFind(1213289) // 5pc: the dodge and the parry it takes off

// https://www.wowhead.com/forever/item-set=124/deathbone-guardian
//
// (5) Set: Reduces the chance for your attacks to be dodged or parried by 2%.
//
// The two effects state the same percent, one for the dodge and one for the parry, which the sim's
// expertise lowers together.
var ItemSetDeathboneGuardian = core.NewItemSet(core.ItemSet{
	ID:   124,
	Name: "Deathbone Guardian",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist,
		proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(7516)),    // +3 defense
		3: shared.ParsedSetBonus(spelldata.MustFind(1299743)), // a spell damage surge when struck
		4: shared.ParsedSetBonus(spelldata.MustFind(18676)),   // +5 all resistances
		// TODO: Ingame test needed -- whether the dodge and parry reduction works as expertise.
		5: func(_ core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.ExpertisePercent, -deathboneExpertise.EffectN(1).BaseValue()).
				ExposeToAPL(deathboneExpertise.ID)
		},
	},
})

var (
	siphonTheGraveProc = spelldata.MustFind(1321560) // 4pc: the proc
	siphonTheGrave     = spelldata.MustFind(1321558) // 4pc: Siphon the Grave, the mana it restores over time
)

// https://www.wowhead.com/forever/item-set=2133/rotmenders-raiment
//
// (4) Set: When your Mana falls below 15%, restore 200 Mana over 10s. Restores twice as much Mana in
// Haunted and Wasteland areas. May only occur once every 5 min.
//
// The amount takes 1321558's area bonus, which an override supplies: the game files state no doubling
// for the Haunted and Wasteland areas the tooltip names.
var ItemSetRotmendersRaiment = core.NewItemSet(core.ItemSet{
	ID:   2133,
	Name: "Rotmender's Raiment",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist,
		proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(14668)), // +5 Shadow Resistance
		3: shared.ParsedSetBonus(spelldata.MustFind(9106)),  // +10 Intellect
		// TODO: Ingame test needed -- whether any cast that leaves mana below 15% triggers it.
		4: siphonTheGraveMana,
		// TODO: 5pc (1321587) heals over time off a heal, at a rate its row does not state.
	},
})

func siphonTheGraveMana(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	if !character.HasManaBar() {
		return
	}
	tick := siphonTheGrave.ProcEnergizeEffect()
	area, _ := siphonTheGrave.AreaBonus(&character.Env.Encounter)
	mana := tick.Average() * area
	metrics := character.NewManaMetrics(core.ActionID{SpellID: siphonTheGrave.ID})

	config := procSpellConfig(character, siphonTheGrave)
	config.Hot = spelldata.DotConfig(siphonTheGrave, tick)
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
		character.AddMana(sim, mana, metrics)
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.SelfHot().Apply(sim)
	}
	siphon := character.RegisterSpell(config)

	threshold := siphonTheGraveProc.EffectN(1).Percent()
	trigger := spelldata.ProcTrigger(character, siphonTheGraveProc, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		siphon.Cast(sim, &character.Unit)
	})
	trigger.Name = "Rotmender's Raiment - 4PC"
	trigger.ExtraCondition = func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
		return character.CurrentManaPercent() < threshold
	}
	setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(siphonTheGraveProc.ID)
}

var (
	spidersKissProc   = spelldata.MustFind(17332) // 2pc: the proc
	spidersKissDebuff = spelldata.MustFind(17333) // 2pc: Spider's Kiss, the root and the armor it takes
)

// https://www.wowhead.com/forever/item-set=65/spiders-kiss
//
// (2) Set: Chance on Hit: Immobilizes the target and lowers their armor by 100 for 10s.
//
// The root has no place in the sim. The armor is the row's own debuff in no armor category, so it stacks
// with every armor debuff the raid puts up.
var ItemSetSpidersKiss = core.NewItemSet(core.ItemSet{
	ID:    65,
	Name:  "Spider's Kiss",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand},
	Bonuses: map[int32]core.ApplySetBonus{
		// TODO: Ingame test needed -- whether the armor reduction stacks with other armor debuffs.
		2: spidersKiss,
	},
})

func spidersKiss(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	debuffs := shared.EnemyDebuffAuras(character, spidersKissDebuff, spidersKissDebuff.Duration(), spidersKissDebuff.DebuffEffects())

	trigger := spelldata.ProcTrigger(character, spidersKissProc, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		debuffs.Get(result.Target).Activate(sim)
	})
	trigger.Name = "Spider's Kiss - 2PC"
	setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(spidersKissProc.ID)
}

// https://www.wowhead.com/forever/item-set=2131/partners-in-crime
//
// (2) Set: +22 Strength.
// (2) Set: Chance on melee swing to become Angry, increasing Strength by 50 for 10s.
var ItemSetPartnersInCrime = core.NewItemSet(core.ItemSet{
	ID:    2131,
	Name:  "Partners in Crime",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2},
	Bonuses: map[int32]core.ApplySetBonus{
		// TODO: 1317039 also puts Angry 1317040 up on a melee swing or ability. Its row states no rate, and
		// the rate is being measured.
		2: shared.ParsedSetBonus(spelldata.MustFind(1317038)), // +22 Strength
	},
})

var (
	roarOfTheCrowdProc = spelldata.MustFind(1301123) // 4pc: the proc
	roarOfTheCrowd     = spelldata.MustFind(1301124) // 4pc: Roar of the Crowd, the haste it puts up
)

// Proc chance measurement in EV-0166.
// The row states none.
const roarOfTheCrowdChance = 0.072

// https://www.wowhead.com/forever/item-set=1/the-gladiator
//
// (4) Set: Chance on hit to gain 5% attack speed for 10s. Attacks while no allies are within 20 yds are 2
// times as likely to activate this effect.
//
// Which hits roll in EV-0166.
// The wearer always has an ally within 20 yd, so the chance never doubles. The 2-piece bonus lowers the
// hit chance of Fear effects, which has no place in the sim.
var ItemSetTheGladiator = core.NewItemSet(core.ItemSet{
	ID:   1,
	Name: "The Gladiator",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet},
	Bonuses: map[int32]core.ApplySetBonus{
		3: shared.ParsedSetBonus(spelldata.MustFind(9331)), // +20 attack power
		4: roarOfTheCrowdHaste,
		5: shared.ParsedSetBonus(spelldata.MustFind(1314795)), // 1% crit
	},
})

func roarOfTheCrowdHaste(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	buff := character.RegisterAura(spelldata.AuraConfig(roarOfTheCrowd))
	spelldata.ParseEffects(character, buff, roarOfTheCrowd)

	trigger := spelldata.ProcTrigger(character, roarOfTheCrowdProc, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.Activate(sim)
	}, spelldata.Chance(roarOfTheCrowdChance), landedHits(core.ProcMaskMelee))
	trigger.Name = "The Gladiator - 4PC"
	setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(roarOfTheCrowdProc.ID)
}

var (
	bloodmailProc        = spelldata.MustFind(1299740) // 3pc: the proc
	bloodmailBleed       = spelldata.MustFind(1299739) // 3pc: Bloodmail, the bleed it casts
	windfuryWeaponStrike = spelldata.MustFind(439440)  // Windfury Weapon hit
)

// Proc chance measurement in EV-0167.
// The row states none.
const bloodmailChance = 0.107

// https://www.wowhead.com/forever/item-set=123/bloodmail-regalia
//
// (3) Set: Grants your melee attacks a chance to wound the target, causing them to Bleed for 70 damage over
// 10s.
//
// Which hits roll in EV-0167.
var ItemSetBloodmailRegalia = core.NewItemSet(core.ItemSet{
	ID:   123,
	Name: "Bloodmail Regalia",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist,
		proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(9140)), // +10 attack power
		// TODO: Ingame test needed -- whether Windfury hits proc it, and whether it is a chance per swing
		// or procs per minute.
		3: bloodmailWound,
		4: shared.ParsedSetBonus(spelldata.MustFind(18676)),   // +5 all resistances
		5: shared.ParsedSetBonus(spelldata.MustFind(1299738)), // 1.5% crit
	},
})

func bloodmailWound(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	config := procSpellConfigAs(character, bloodmailBleed, spelldata.Melee(core.ProcMaskEmpty))
	config.Dot = spelldata.DotConfig(bloodmailBleed, bloodmailBleed.PeriodicDamageEffect())
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeAlwaysHit)
		spell.Dot(target).Apply(sim)
		spell.DealOutcome(sim, result)
	}
	bleed := character.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(character, bloodmailProc, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		bleed.Cast(sim, result.Target)
	}, spelldata.Chance(bloodmailChance), landedHits(core.ProcMaskMelee))
	trigger.Name = "Bloodmail Regalia - 3PC"
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) || spell.Flags.Matches(core.SpellFlagCountsAsSwing) ||
			spell.SpellID == windfuryWeaponStrike.ID
	}
	setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(bloodmailProc.ID)
}

var (
	necropileDrainProc = spelldata.MustFind(1299737) // 3pc: the proc
	necropileDrain     = spelldata.MustFind(1299735) // 3pc: Necropile Drain, the leech it casts
)

// Proc chance measurement in EV-0168.
// The row states none.
const necropileDrainChance = 0.139

// https://www.wowhead.com/forever/item-set=122/necropile-raiment
//
// (3) Set: Harmful spell casts have a chance to steal 28 life from the target enemy.
//
// Which landings roll in EV-0168.
// The row's 10 ms cooldown lets an area spell proc it once.
var ItemSetNecropileRaiment = core.NewItemSet(core.ItemSet{
	ID:   122,
	Name: "Necropile Raiment",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist,
		proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(1299734)), // 0.5% hit
		// TODO: Ingame test needed -- whether every harmful spell has the same chance.
		3: necropileDrainLeech,
		4: shared.ParsedSetBonus(spelldata.MustFind(18676)), // +5 all resistances
		5: shared.ParsedSetBonus(dungeonTwoDamage),
	},
})

func necropileDrainLeech(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	leech := necropileDrain.DamageEffect()
	healthMetrics := character.NewHealthMetrics(core.ActionID{SpellID: necropileDrain.ID})
	config := procSpellConfigAs(character, necropileDrain, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, leech.DirectAmount(sim), spell.OutcomeAlwaysHit)
		shared.GainLeechHealth(sim, character, result.Damage, leech, healthMetrics)
	}
	drain := character.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(character, necropileDrainProc, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		drain.Cast(sim, result.Target)
	}, spelldata.Chance(necropileDrainChance), landedHits(core.ProcMaskSpellDamage))
	trigger.Name = "Necropile Raiment - 3PC"
	setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(necropileDrainProc.ID)
}

// Every landed hit of the mask, one that deals no damage included.
func landedHits(mask core.ProcMask) spelldata.ProcOpt {
	return func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.Callback = core.CallbackOnSpellHitDealt
		trigger.ProcMask = mask
		trigger.Outcome = core.OutcomeLanded
		trigger.RequireDamageDealt = false
	}
}

// The row's amount, or in the tooltip's area the amount its second effect states.
func secondEffectInArea(character *core.Character, row *spelldata.Spell, area proto.AreaType) spelldata.ParseOpt {
	if !character.Env.Encounter.InArea(area) {
		return spelldata.Count(1)
	}
	return spelldata.Count(row.EffectN(2).BaseValue() / row.EffectN(1).BaseValue())
}

var (
	violetMeditationProc = spelldata.MustFind(1324318) // 4pc: the proc
	violetMeditation     = spelldata.MustFind(1324315) // 4pc: Violet Meditation, the regen it puts up
	arcaneElementalProc  = spelldata.MustFind(1324323) // 5pc: the proc
	arcaneElementalCall  = spelldata.MustFind(1324322) // 5pc: Summon Arcane Elemental
)

// https://www.wowhead.com/forever/item-set=2134/violet-sorcerers-vestments
//
// (4) Set: Chance on damaging spell to allow 15% of your Mana regeneration to continue while casting for 15s.
// This Mana Regeneration is doubled in Strongholds and Cities.
// (5) Set: Chance on damaging spell to summon an Arcane elemental to aid you in combat.
//
// PPM measurement in EV-0169.
var ItemSetVioletSorcerersVestments = core.NewItemSet(core.ItemSet{
	ID:   2134,
	Name: "Violet Sorcerer's Vestments",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands,
		proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(1324311)), // +5 Arcane Resistance
		3: shared.ParsedSetBonus(spelldata.MustFind(1324312)), // +10 Spirit
		4: violetMeditationRegen,
		// TODO: Ingame test needed -- the elemental's swing at level 60.
		5: summonArcaneElemental,
	},
})

func violetMeditationRegen(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	if !character.HasManaBar() {
		return
	}
	buff := character.RegisterAura(spelldata.AuraConfig(violetMeditation))
	spelldata.ParseEffects(character, buff, violetMeditation,
		secondEffectInArea(character, violetMeditation, proto.AreaType_AreaTypeStrongholdsCities))

	trigger := spelldata.ProcTrigger(character, violetMeditationProc, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.Activate(sim)
	}, castTimePPM(violetMeditationProc))
	trigger.Name = "Violet Sorcerer's Vestments - 4PC"
	trigger.ProcMask = core.ProcMaskSpellDamage
	setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(violetMeditationProc.ID)
}

func summonArcaneElemental(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	elementals := shared.Guardians(character, arcaneElementalCall)

	// A summon takes an elemental that is not up; the 15 s rest leaves one free whenever a summon comes.
	config := procSpellConfig(character, arcaneElementalCall)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
		shared.SummonGuardians(sim, elementals, target)
	}
	call := character.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(character, arcaneElementalProc, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		call.Cast(sim, result.Target)
	}, castTimePPM(arcaneElementalProc))
	trigger.Name = "Violet Sorcerer's Vestments - 5PC"
	trigger.ProcMask = core.ProcMaskSpellDamage
	setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(arcaneElementalProc.ID)
}

var (
	defiasJuggernautProc = spelldata.MustFind(1324684) // 5pc: the proc
	defiasJuggernaut     = spelldata.MustFind(1324683) // 5pc: Defias Juggernaut, the crit it puts up
)

// https://www.wowhead.com/forever/item-set=2136/defias-enforcers-garb
//
// (5) Set: Melee attacks have a chance to increase your Critical Strike chance by 2% for 20s. This effect is
// doubled in swamps.
//
// The row names the main hand alone, but off-hand swings proc it as often (EV-0170).
var ItemSetDefiasEnforcersGarb = core.NewItemSet(core.ItemSet{
	ID:   2136,
	Name: "Defias Enforcer's Garb",
	Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(1324673)), // +10 Fire Resistance
		3: shared.ParsedSetBonus(spelldata.MustFind(1324674)), // +15 Stamina
		4: shared.ParsedSetBonus(spelldata.MustFind(1324675)), // +1 Swords
		5: defiasJuggernautCrit,
	},
})

func defiasJuggernautCrit(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	buff := character.RegisterAura(spelldata.AuraConfig(defiasJuggernaut))
	spelldata.ParseEffects(character, buff, defiasJuggernaut,
		secondEffectInArea(character, defiasJuggernaut, proto.AreaType_AreaTypeSwamp))

	// The mask goes in before the rate, so the procs per minute are priced on either hand's swings.
	trigger := spelldata.ProcTrigger(character, defiasJuggernautProc, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.Activate(sim)
	}, landedHits(core.ProcMaskMelee))
	trigger.Name = "Defias Enforcer's Garb - 5PC"
	setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(defiasJuggernautProc.ID)
}
