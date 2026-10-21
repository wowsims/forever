package forever

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Blacksmithing - Plate
// https://www.wowhead.com/forever/item-set=321/imperial-plate
//
// Forever reworked the set: Imperial Plate Gauntlets (250589) joined the seven Classic pieces and
// every bonus is new. The 4-piece bonus keeps movement speed from dropping below 80% and has no
// place in the sim.
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
// New in Forever. The 4-piece bonus restores 20 health every 5 sec to party members within 10
// yards; the sim has no health regeneration stat, so it is left out.
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
	dungeonTwoArmor    = spelldata.MustFind(14803)  // Dungeon Set 2 8pc: +200 armor
)

var magisterManaProc = spelldata.MustFind(450527) // 5pc: Mana Proc on Cast, the proc

// (2) Set: all resistances, and (4) Set: a chance to freeze an attacker when struck, which a boss
// fight does not test.
var ItemSetMagistersRegalia = core.NewItemSet(core.ItemSet{
	Name:  "Magister's Regalia",
	ID:    181,
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		3: shared.ParsedSetBonus(dungeonSpellDamage),
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			character := agent.GetCharacter()
			manaMetrics := character.NewManaMetrics(core.ActionID{SpellID: dungeonSuddenMana.ID})
			mana := dungeonSuddenMana.EnergizeEffect().Average()

			// The rate is the proc chance column with no cooldown, rolled on every spell cast, the
			// tooltip's "your spellcasts".
			trigger := spelldata.ProcTrigger(character, magisterManaProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					character.AddMana(sim, mana, manaMetrics)
				})
			trigger.Name = "Magister's Regalia - 5PC"

			setBonusAura.AttachProcTrigger(trigger)
		},
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// (2) Set: all resistances, (4) Set: a chance to freeze an attacker when struck, and (8) Set: armor,
// none of which a boss fight tests.
var ItemSetSorcerersRegalia = core.NewItemSet(core.ItemSet{
	Name:  "Sorcerer's Regalia",
	ID:    517,
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		6: shared.ParsedSetBonus(dungeonTwoDamage),
	},
})

var devoutManaProc = spelldata.MustFind(450576) // 5pc: the proc

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
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your spellcasts have a 5% chance to energize you for 200 mana.
			character := agent.GetCharacter()
			manaMetrics := character.NewManaMetrics(core.ActionID{SpellID: dungeonSuddenMana.ID})
			mana := dungeonSuddenMana.EnergizeEffect().Average()

			trigger := spelldata.ProcTrigger(character, devoutManaProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					character.AddMana(sim, mana, manaMetrics)
				})
			trigger.Name = "Vestments of the Devout - 5PC"

			setBonusAura.AttachProcTrigger(trigger)
			setBonusAura.ExposeToAPL(devoutManaProc.ID)
		},
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 2
// https://www.wowhead.com/forever/item-set=514/vestments-of-the-virtuous
//
// The 4-piece bonus (27778) shields the wearer when struck, which the sim does not model.
var ItemSetVestmentsOfTheVirtuous = core.NewItemSet(core.ItemSet{
	ID:    514,
	Name:  "Vestments of the Virtuous",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		6: shared.ParsedSetBonus(dungeonTwoDamage),
		8: shared.ParsedSetBonus(dungeonTwoArmor),
	},
})

var (
	dreadmistHealProc = spelldata.MustFind(450585) // 5pc: the proc
	dreadmistHeal     = spelldata.MustFind(450583) // 5pc: Dark Reward, the heal it casts
)

// Dungeon Set 1
var ItemSetDreadmistRaiment = core.NewItemSet(core.ItemSet{
	ID:    183,
	Name:  "Dreadmist Raiment",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonSpellDamage),
		// TODO: 4pc (27780) fears an attacker when struck, which the sim does not model.
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			character := agent.GetCharacter()
			heal := dreadmistHeal.HealEffect()

			config := spelldata.SpellConfig(&character.Unit, dreadmistHeal, spelldata.Magic(core.ProcMaskSpellHealing), spelldata.Proc())
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
			trigger.Name = "Dreadmist Raiment - 5PC"

			setBonusAura.AttachProcTrigger(trigger)
		},
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 2
var ItemSetDeathmistRaiment = core.NewItemSet(core.ItemSet{
	ID:    518,
	Name:  "Deathmist Raiment",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		// TODO: 4pc (27780) fears an attacker when struck, which the sim does not model.
		6: shared.ParsedSetBonus(dungeonTwoDamage),
		8: shared.ParsedSetBonus(dungeonTwoArmor),
	},
})

var (
	wildheartProc = spelldata.MustFind(450608) // 5pc: the proc
	wildheartMana = spelldata.MustFind(27782)  // 5pc: the mana a spell cast restores
)

// Dungeon Set 1
// https://www.wowhead.com/forever/item-set=185/wildheart-raiment
//
// The 4-piece bonus is movement speed when struck, which has no place in the sim.
var ItemSetWildheartRaiment = core.NewItemSet(core.ItemSet{
	ID:    185,
	Name:  "Wildheart Raiment",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		3: shared.ParsedSetBonus(dungeonSpellDamage, spelldata.MustFind(9336)), // +18 spell damage, +30 attack power
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			// TODO: The proc also answers a melee swing with energy (27784, whose row restores mana) and
			// a melee hit taken with rage (27783); only the spell cast's mana is modelled.
			character := agent.GetCharacter()
			mana := wildheartMana.EnergizeEffect().Average()
			manaMetrics := character.NewManaMetrics(core.ActionID{SpellID: wildheartMana.ID})

			// The column reads 100 and the trigger clause states no number; the dummy effect's 5
			// is the chance. The row's mask carries the melee branches too, so the listener takes
			// the spell cast's shape by hand.
			trigger := spelldata.ProcTrigger(character, wildheartProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					character.AddMana(sim, mana, manaMetrics)
				}, spelldata.ChanceFrom(wildheartProc.EffectN(1)))
			trigger.Name = "Wildheart Raiment - 5PC"
			trigger.Callback = core.CallbackOnCastComplete
			trigger.ProcMask = core.ProcMaskSpellDamage | core.ProcMaskSpellHealing
			trigger.Outcome = core.OutcomeEmpty
			trigger.RequireDamageDealt = false

			setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(wildheartProc.ID)
		},
		6: shared.ParsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 2
// https://www.wowhead.com/forever/item-set=513/feralheart-raiment
//
// TODO: The 4-piece bonus returns 300 mana, 10 rage or 40 energy when struck. Its proc (27781) casts a
// dummy (27798) and only the tooltip states the amounts, so it is not modelled.
var ItemSetFeralheartRaiment = core.NewItemSet(core.ItemSet{
	ID:    513,
	Name:  "Feralheart Raiment",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		6: shared.ParsedSetBonus(spelldata.MustFind(9344), spelldata.MustFind(9334)), // +15 spell damage, +26 attack power
		8: shared.ParsedSetBonus(dungeonTwoArmor),
	},
})

var (
	rogueArmorEnergyProc = spelldata.MustFind(27787) // Shadowcraft 5pc, Darkmantle 4pc: Rogue Armor Energize, the proc
	rogueArmorEnergy     = spelldata.MustFind(27788) // the energy it restores
)

// The energy proc of the two rogue dungeon sets. The row states 100% on every melee hit and no rate
// per minute, so the sim keeps 1 per minute as a literal. A wearer with no energy bar gets nothing.
func rogueArmorEnergize(name string) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		character := agent.GetCharacter()
		gain := shared.FiredSpellHandler(character, core.ActionID{SpellID: rogueArmorEnergy.ID}, rogueArmorEnergyProc)
		if gain == nil {
			return
		}

		trigger := spelldata.ProcTrigger(character, rogueArmorEnergyProc, gain, spelldata.PPM(1.0))
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

// Dungeon Set 2
// https://www.wowhead.com/forever/item-set=512/darkmantle-armor
var ItemSetDarkmantleArmor = core.NewItemSet(core.ItemSet{
	ID:    512,
	Name:  "Darkmantle Armor",
	Slots: shared.EightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(dungeonResistances),
		4: rogueArmorEnergize("Darkmantle 4-piece Proc"),
		6: shared.ParsedSetBonus(spelldata.MustFind(14049)), // +40 attack power
		8: shared.ParsedSetBonus(dungeonTwoArmor),
	},
})
