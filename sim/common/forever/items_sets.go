package forever

import (
	"slices"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// A bonus whose rows the parse states whole, exposed to the APL under the first row's id. A row's spell
// modifiers name its class's spells by class flags, so on another class they reach nothing.
func parsedSetBonus(rows ...*spelldata.Spell) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		for _, row := range rows {
			spelldata.ParseEffects(agent.GetCharacter(), setBonusAura, row)
		}
		setBonusAura.ExposeToAPL(rows[0].ID)
	}
}

// The eight armor slots of a crafted or dungeon set.
var eightPieceSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotHead,
	proto.ItemSlot_ItemSlotShoulder,
	proto.ItemSlot_ItemSlotChest,
	proto.ItemSlot_ItemSlotWrist,
	proto.ItemSlot_ItemSlotHands,
	proto.ItemSlot_ItemSlotWaist,
	proto.ItemSlot_ItemSlotLegs,
	proto.ItemSlot_ItemSlotFeet,
}

// A Naxxramas tier 3 set: the eight armor pieces, which the game files open to every class, and the
// ring, which they keep to the set's class.
var tierThreeSlots = slices.Concat(eightPieceSlots, []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2})

// Blacksmithing - Plate
// https://www.wowhead.com/forever/item-set=321/imperial-plate
//
// Forever reworked the set: Imperial Plate Gauntlets (250589) joined the seven Classic pieces and
// every bonus is new. The 4-piece bonus keeps movement speed from dropping below 80% and has no
// place in the sim.
var ItemSetImperialPlate = core.NewItemSet(core.ItemSet{
	ID:    321,
	Name:  "Imperial Plate",
	Slots: eightPieceSlots,
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
	Slots: eightPieceSlots,
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
	Slots: eightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		3: parsedSetBonus(dungeonSpellDamage),
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
		6: parsedSetBonus(dungeonManaRegen),
	},
})

// (2) Set: all resistances, (4) Set: a chance to freeze an attacker when struck, and (8) Set: armor,
// none of which a boss fight tests.
var ItemSetSorcerersRegalia = core.NewItemSet(core.ItemSet{
	Name:  "Sorcerer's Regalia",
	ID:    517,
	Slots: eightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		6: parsedSetBonus(dungeonTwoDamage),
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
	Slots: eightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(dungeonResistances),
		3: parsedSetBonus(dungeonSpellDamage),
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
		6: parsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 2
// https://www.wowhead.com/forever/item-set=514/vestments-of-the-virtuous
//
// The 4-piece bonus (27778) shields the wearer when struck, which the sim does not model.
var ItemSetVestmentsOfTheVirtuous = core.NewItemSet(core.ItemSet{
	ID:    514,
	Name:  "Vestments of the Virtuous",
	Slots: eightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(dungeonResistances),
		6: parsedSetBonus(dungeonTwoDamage),
		8: parsedSetBonus(dungeonTwoArmor),
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
	Slots: eightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(dungeonResistances),
		3: parsedSetBonus(dungeonSpellDamage),
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
		6: parsedSetBonus(dungeonManaRegen),
	},
})

// Dungeon Set 2
var ItemSetDeathmistRaiment = core.NewItemSet(core.ItemSet{
	ID:    518,
	Name:  "Deathmist Raiment",
	Slots: eightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(dungeonResistances),
		// TODO: 4pc (27780) fears an attacker when struck, which the sim does not model.
		6: parsedSetBonus(dungeonTwoDamage),
		8: parsedSetBonus(dungeonTwoArmor),
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
	Slots: eightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(dungeonResistances),
		3: parsedSetBonus(dungeonSpellDamage, spelldata.MustFind(9336)), // +18 spell damage, +30 attack power
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
		6: parsedSetBonus(dungeonManaRegen),
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
	Slots: eightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(dungeonResistances),
		6: parsedSetBonus(spelldata.MustFind(9344), spelldata.MustFind(9334)), // +15 spell damage, +26 attack power
		8: parsedSetBonus(dungeonTwoArmor),
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
	Slots: eightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(dungeonResistances),
		3: parsedSetBonus(spelldata.MustFind(9336)), // +30 attack power
		5: rogueArmorEnergize("Shadowcraft 5-piece Proc"),
	},
})

// Dungeon Set 2
// https://www.wowhead.com/forever/item-set=512/darkmantle-armor
var ItemSetDarkmantleArmor = core.NewItemSet(core.ItemSet{
	ID:    512,
	Name:  "Darkmantle Armor",
	Slots: eightPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(dungeonResistances),
		4: rogueArmorEnergize("Darkmantle 4-piece Proc"),
		6: parsedSetBonus(spelldata.MustFind(14049)), // +40 attack power
		8: parsedSetBonus(dungeonTwoArmor),
	},
})

var (
	dreadnaughtRevenge        = spelldata.MustFind(28844) // 2pc: Revenge
	dreadnaughtTauntHit       = spelldata.MustFind(28843) // 4pc: Increased Spell Hit Chance
	dreadnaughtAbilityHit     = spelldata.MustFind(28842) // 6pc: Increased Hit Chance
	dreadnaughtCheatDeath     = spelldata.MustFind(28845) // 8pc: Cheat Death, the proc
	dreadnaughtCheatDeathBuff = spelldata.MustFind(28846) // 8pc: Cheat Death, the healing buff
)

// Tier 3
var ItemSetDreadnaughtsBattlegear = core.NewItemSet(core.ItemSet{
	Name:  "Dreadnaught's Battlegear",
	ID:    523,
	Slots: tierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(_ core.Agent, setBonusAura *core.Aura) {
			// A_ADD_FLAT_MODIFIER on the damage has no sim kind in the parse table, so only the
			// amount and the spell it names come off the row.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				ClassFlags: dreadnaughtRevenge.EffectN(1).ClassFlags,
				Kind:       core.SpellMod_BaseDamage_Flat,
				FloatValue: dreadnaughtRevenge.EffectN(1).BaseValue(),
			})
		},
		// TODO: both spells resolve on OutcomeAlwaysHit, so the bonus changes nothing until they
		// roll against the spell hit table.
		4: parsedSetBonus(dreadnaughtTauntHit),
		6: parsedSetBonus(dreadnaughtAbilityHit),
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			// Spell 28845 states that below 20% health, healing spells cast on you gain up to 160
			// healing (spell 28846) for 5 seconds.
			// TODO: the tank sim's modelled incoming healing bypasses the bonus; only heals a healer
			// unit casts take it.
			character := agent.GetCharacter()

			cheatDeath := character.RegisterAura(spelldata.AuraConfig(dreadnaughtCheatDeathBuff))
			spelldata.ParseEffects(character, cheatDeath, dreadnaughtCheatDeathBuff)

			trigger := spelldata.ProcTrigger(character, dreadnaughtCheatDeath,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					cheatDeath.Activate(sim)
				})
			trigger.Name = "Cheat Death - Trigger"
			// The health threshold is the whole of the tooltip's condition and the row states none
			// of it.
			trigger.ExtraCondition = func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
				return character.CurrentHealthPercent() < 0.2
			}

			setBonusAura.AttachProcTrigger(trigger)
		},
	},
})

var (
	cryptstalkerRapidFire      = spelldata.MustFind(28755) // 2pc: Rapid Fire lasts 4 s longer
	cryptstalkerAdrenalineProc = spelldata.MustFind(28752) // 6pc: Adrenaline Rush, the proc
	cryptstalkerAdrenalineMana = spelldata.MustFind(28753) // 6pc: Adrenaline Rush, the mana it grants
	cryptstalkerShotCost       = spelldata.MustFind(28751) // 8pc: Multi-Shot and Aimed Shot cost 20 less
)

// Tier 3
var ItemSetCryptstalkerArmor = core.NewItemSet(core.ItemSet{
	Name:  "Cryptstalker Armor",
	ID:    530,
	Slots: tierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(cryptstalkerRapidFire),
		// (4) Set: While your pet is active, increases Attack Power by 50 for both you and your pet.
		// The pet is not modelled.
		6: func(agent core.Agent, setBonusAura *core.Aura) {
			character := agent.GetCharacter()
			manaMetrics := character.NewManaMetrics(core.ActionID{SpellID: cryptstalkerAdrenalineMana.ID})
			mana := cryptstalkerAdrenalineMana.EnergizeEffect().Average()

			// 28752 states "always" on ranged hits and its proc effect names Multi-Shot alone; the
			// tooltip says your ranged critical hits, and the tooltip wins: the crit is the outcome
			// and the mask comes off.
			trigger := spelldata.ProcTrigger(character, cryptstalkerAdrenalineProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					character.AddMana(sim, mana, manaMetrics)
				})
			trigger.Name = "Cryptstalker Armor - 6PC"
			trigger.Outcome = core.OutcomeCrit
			trigger.ClassFlags = core.ClassFlags{}

			setBonusAura.AttachProcTrigger(trigger)
		},
		8: parsedSetBonus(cryptstalkerShotCost),
	},
})

var (
	faithRenewCost    = spelldata.MustFind(28807) // 2pc: Renew costs 12% less
	faithHealThreat   = spelldata.MustFind(28808) // 6pc: less threat from healing spells
	faithEpiphanyProc = spelldata.MustFind(28802) // 8pc: Epiphany, the proc
	faithEpiphanyBuff = spelldata.MustFind(28804) // 8pc: Epiphany, the mana regeneration
)

// Tier 3
// https://www.wowhead.com/forever/item-set=525/vestments-of-faith
//
// The 4-piece bonus (28809) shields the target of a Greater Heal crit with Armor of Faith, which
// the sim does not model.
var ItemSetVestmentsOfFaith = core.NewItemSet(core.ItemSet{
	ID:    525,
	Name:  "Vestments of Faith",
	Slots: tierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(faithRenewCost),
		6: parsedSetBonus(faithHealThreat),
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			// Each spell you cast can trigger an Epiphany, increasing your mana regeneration by 24
			// for 30 sec.
			character := agent.GetCharacter()
			epiphany := character.RegisterAura(spelldata.AuraConfig(faithEpiphanyBuff))
			spelldata.ParseEffects(character, epiphany, faithEpiphanyBuff)

			// The proc effect's class flags name Shadow Word: Pain alone; the tooltip says each spell
			// you cast, and the tooltip wins: the flags come off.
			trigger := spelldata.ProcTrigger(character, faithEpiphanyProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					epiphany.Activate(sim)
				})
			trigger.Name = "Vestments of Faith - 8PC"
			trigger.ClassFlags = core.ClassFlags{}

			setBonusAura.AttachProcTrigger(trigger)
			setBonusAura.ExposeToAPL(faithEpiphanyProc.ID)
		},
	},
})

var (
	frostfireEvocation     = spelldata.MustFind(28763) // 2pc: Evocation cooldown 60 s shorter
	frostfireVulnerability = spelldata.MustFind(28771) // 6pc: Elemental Vulnerability, the proc
	frostfireVulnDebuff    = spelldata.MustFind(28772) // 6pc: Elemental Vulnerability, the debuff
	frostfireNotThere      = spelldata.MustFind(28761) // 8pc: Not There, the proc
	frostfireNotThereBuff  = spelldata.MustFind(28762) // 8pc: Not There, the threat buff
)

// Tier 3
//
// (4) Set: Adaptive Warding raises a resistance when Mage Armor is struck by a spell, which a boss
// fight does not test.
var ItemSetFrostfireRegalia = core.NewItemSet(core.ItemSet{
	Name:  "Frostfire Regalia",
	ID:    526,
	Slots: tierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(frostfireEvocation),
		6: func(agent core.Agent, setBonusAura *core.Aura) {
			character := agent.GetCharacter()

			// A_MOD_DAMAGE_TAKEN rather than a from-caster aura: every caster's spells take it. "Up to
			// $s1" is spell damage against the target, which each spell takes at its own coefficient,
			// as Judgement of the Crusader does, so it is not parsed into the flat damage taken.
			// The row's one charge is spent by the next spell that hits the target.
			taken := frostfireVulnDebuff.EffectN(1)
			var takenSchools []stats.SchoolIndex
			for school := core.SpellSchoolHoly; school <= core.SpellSchoolArcane; school <<= 1 {
				if taken.Misc&int32(school) != 0 {
					takenSchools = append(takenSchools, school.SchoolIndex())
				}
			}
			debuffs := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
				aura := target.RegisterAura(spelldata.AuraConfig(frostfireVulnDebuff,
					spelldata.Label(frostfireVulnDebuff.Name+character.Label)))
				aura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
					for _, i := range takenSchools {
						target.PseudoStats.SchoolBonusSpellDamage[i] += taken.BaseValue()
					}
				}).ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
					for _, i := range takenSchools {
						target.PseudoStats.SchoolBonusSpellDamage[i] -= taken.BaseValue()
					}
				})

				consume := spelldata.ProcTrigger(character, frostfireVulnDebuff,
					func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
						aura.Deactivate(sim)
					})
				consume.Name = "Elemental Vulnerability - Consume" + character.Label
				consume.TriggerImmediately = true
				return aura.AttachProcTrigger(consume)
			})

			// The rate is the proc chance column. PROC_FLAG_DEAL_HARMFUL_SPELL decodes to the cast,
			// which names no target; the debuff goes on the one a damage spell landed on, the
			// tooltip's "your damage spells".
			trigger := spelldata.ProcTrigger(character, frostfireVulnerability,
				func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
					debuffs.Get(result.Target).Activate(sim)
				})
			trigger.Name = "Frostfire Regalia - 6PC"
			trigger.Callback = core.CallbackOnSpellHitDealt
			trigger.Outcome = core.OutcomeLanded
			trigger.RequireDamageDealt = false
			trigger.RequireDamageComponent = true

			setBonusAura.AttachProcTrigger(trigger)
		},
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			character := agent.GetCharacter()

			// A_MOD_THREAT -100% on the schools the misc value masks is a multiplier of 0, which the
			// parse cannot divide back out, so the buff holds the multipliers it zeroes and puts
			// them back.
			buff := character.RegisterAura(spelldata.AuraConfig(frostfireNotThereBuff))
			threatEffect := frostfireNotThereBuff.EffectN(1)
			var schools []stats.SchoolIndex
			for school := core.SpellSchoolHoly; school <= core.SpellSchoolArcane; school <<= 1 {
				if threatEffect.Misc&int32(school) != 0 {
					schools = append(schools, school.SchoolIndex())
				}
			}
			var held [stats.SchoolLen]float64
			buff.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
				for _, i := range schools {
					held[i] = character.PseudoStats.SchoolThreatMultiplier[i]
					character.PseudoStats.SchoolThreatMultiplier[i] *= core.MultiplierFromPercent(threatEffect.BaseValue())
				}
			}).ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
				for _, i := range schools {
					character.PseudoStats.SchoolThreatMultiplier[i] = held[i]
				}
			})

			// The row spends its one charge on the cast, but the sim reckons threat when a hit
			// deals, which a spell in flight does after its cast completes: the buff goes on the
			// next damage spell's hit.
			consume := spelldata.ProcTrigger(character, frostfireNotThereBuff,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					buff.Deactivate(sim)
				})
			consume.Name = "Not There - Consume"
			consume.Callback = core.CallbackOnSpellHitDealt
			consume.Outcome = core.OutcomeLanded
			consume.RequireDamageDealt = false
			consume.RequireDamageComponent = true
			consume.TriggerImmediately = true
			buff.AttachProcTrigger(consume)

			// The rate is the proc chance column. The proc effect's class flags name another
			// class's family, which the resolver drops. It listens to landed damage spells for the
			// consume's reason, so the hit that procs it keeps its threat and the next one loses it.
			trigger := spelldata.ProcTrigger(character, frostfireNotThere,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					buff.Activate(sim)
				})
			trigger.Name = "Frostfire Regalia - 8PC"
			trigger.Callback = core.CallbackOnSpellHitDealt
			trigger.Outcome = core.OutcomeLanded
			trigger.RequireDamageDealt = false
			trigger.RequireDamageComponent = true

			setBonusAura.AttachProcTrigger(trigger)
		},
	},
})

var (
	plagueCorruption = spelldata.MustFind(28829) // 4pc: Corruption's damage
	plagueThreat     = spelldata.MustFind(28746) // 6pc: threat
)

// Tier 3
var ItemSetPlagueheartRaiment = core.NewItemSet(core.ItemSet{
	ID:    529,
	Name:  "Plagueheart Raiment",
	Slots: tierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		// 2pc (28831) heals the wearer off Shadow Bolt hits, which does not affect sims.
		4: parsedSetBonus(plagueCorruption),
		// Effect 2 lowers the threat of Corruption, Immolate, Curse of Agony and Siphon Life. Effect 1,
		// 25% less threat from spell crits, has no place in the sim's threat.
		6: parsedSetBonus(plagueThreat),
		// 8pc (28830) takes 12% off Life Tap's health cost, which does not affect sims.
	},
})
