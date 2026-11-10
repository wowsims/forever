package buffmanifest

import "github.com/wowsims/forever/sim/core/proto"

// One row per proto field of RaidBuffs, PartyBuffs, IndividualBuffs and Debuffs, in the order the
// settings tab lists them, with the fields that have no input after the last one that does. A
// field's proto number is its position in its slice.
//
// A row's Ranks open with the rank no class book teaches, and a book teaches the last rank of a totem or
// a greater blessing, so the list opens with the trainer's top rank. Mark of the Wild, Power Word:
// Fortitude, Arcane Intellect and Shadow Protection read the single-target spell's top rank, whose
// amounts match the group spell's top rank.

var Raid = []BuffSpec{
	{
		Field:    "arcane_intellect",
		SpellID:  10157,
		Category: "StatBuff",
		Stats: []proto.Stat{
			proto.Stat_StatIntellect,
		},
	},
	{
		Field:    "prayer_of_spirit",
		SpellID:  27681,
		Category: "StatBuff",
		Stats: []proto.Stat{
			proto.Stat_StatSpirit,
		},
	},
	{
		Field:   "mark_of_the_wild",
		SpellID: 9885,
		Stats: []proto.Stat{
			proto.Stat_StatArmor,
			proto.Stat_StatStrength,
			proto.Stat_StatAgility,
			proto.Stat_StatIntellect,
			proto.Stat_StatSpirit,
			proto.Stat_StatStamina,
		},
	},
	{
		Field:    "power_word_fortitude",
		SpellID:  10938,
		Category: "StatBuff",
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
		},
	},
	{
		Field:   "shadow_protection",
		SpellID: 10958,
		Stats: []proto.Stat{
			proto.Stat_StatShadowResistance,
		},
	},
	{
		Field:          "fire_resistance_aura",
		SpellID:        19900,
		Category:       "FireResistanceAura",
		SharedCategory: "PaladinAura",
		SingleAura:     true,
		Stats: []proto.Stat{
			proto.Stat_StatFireResistance,
		},
	},
	{
		Field:          "frost_resistance_aura",
		SpellID:        19898,
		Category:       "FrostResistanceAura",
		SharedCategory: "PaladinAura",
		SingleAura:     true,
		Stats: []proto.Stat{
			proto.Stat_StatFrostResistance,
		},
	},
	{
		Field:          "shadow_resistance_aura",
		SpellID:        19896,
		Category:       "ShadowResistanceAura",
		SharedCategory: "PaladinAura",
		SingleAura:     true,
		Stats: []proto.Stat{
			proto.Stat_StatShadowResistance,
		},
	},
	{
		Field:   "fire_resistance_totem",
		SpellID: 10535,
		CastID:  10538,
		Stats: []proto.Stat{
			proto.Stat_StatFireResistance,
		},
	},
	{
		Field:   "frost_resistance_totem",
		SpellID: 10477,
		CastID:  10479,
		Stats: []proto.Stat{
			proto.Stat_StatFrostResistance,
		},
	},
	{
		Field:   "nature_resistance_totem",
		SpellID: 10599,
		CastID:  10601,
		Stats: []proto.Stat{
			proto.Stat_StatNatureResistance,
		},
	},
	{
		Field:    "strength_of_earth_totem",
		SpellID:  10441,
		CastID:   10442,
		Category: "StrengthOfEarthTotem",
		Ranks:    []BuffRank{{SpellID: 10441, CastID: 10442}, {SpellID: 25362, CastID: 25361}},
		Stats: []proto.Stat{
			proto.Stat_StatStrength,
		},
	},
	{
		Field:    "grace_of_air_totem",
		SpellID:  10626,
		CastID:   10627,
		Category: "GraceOfAirTotem",
		Ranks:    []BuffRank{{SpellID: 10626, CastID: 10627}, {SpellID: 25360, CastID: 25359}},
		Stats: []proto.Stat{
			proto.Stat_StatAgility,
		},
	},
	{
		Field:    "windfury_totem",
		SpellID:  10610,
		CastID:   10614,
		Kind:     KindProc,
		Category: "WindfuryTotem",
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	// The totem's raid aura procs 16389 on every main-hand auto attack; the fire damage is the driver's.
	{
		Field:    "flametongue_totem",
		SpellID:  15036,
		CastID:   16387,
		Kind:     KindProc,
		Category: "FlametongueTotem",
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	// Four states: off, the row's Rank 6 (the top rank the trainer teaches), Rank 7 from Manual of
	// Battle Shout VII, and Rank 7 from a warrior wearing three pieces of Battlegear of Wrath, whose
	// set spell 23563 adds 30 to every effect of the shout. No set spell is resolved, so
	// driveBattleShout picks the rank and adds the 30; ImpAction keeps 23563 in the store.
	{
		Field:      "battle_shout",
		SpellID:    11551,
		Category:   "BattleShout",
		SingleAura: true,
		Driver:     true,
		ImpAction:  &ActionRef{SpellID: 23563},
		Proto:      ProtoInt32,
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	// TODO: rank 5 states 50 ranged attack power and rank 4 states 75; in-game testing has to say
	// which one the top rank grants.
	{
		Field:   "trueshot_aura",
		SpellID: 20906,
		Stats: []proto.Stat{
			proto.Stat_StatRangedAttackPower,
		},
	},
	// A pet ability the hunter class family does not file, so the owner is stated. It does not stack
	// with Battle Shout: the two share one category and the stronger applies.
	{
		Field:      "furious_howl",
		SpellID:    24597,
		Owner:      proto.Class_ClassHunter,
		Category:   "BattleShout",
		SingleAura: true,
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	{
		Field:   "aspect_of_the_wild",
		SpellID: 20190,
		Stats: []proto.Stat{
			proto.Stat_StatNatureResistance,
		},
	},
	// Raises all crit, as Moonkin Aura below does; the two share one category, so only one counts.
	{
		Field:      "leader_of_the_pack",
		SpellID:    24932,
		CastID:     17007,
		Category:   "RaidCrit",
		SingleAura: true,
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
			proto.Stat_StatSpellDamage,
		},
	},
	{
		Field:      "moonkin_aura",
		SpellID:    24907,
		TooltipID:  24858,
		Category:   "RaidCrit",
		SingleAura: true,
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
			proto.Stat_StatSpellDamage,
		},
	},
}

var Party = []BuffSpec{
	// The Imp's copy and a ticked one bid per stat, both staying up, so the ticked one applies again
	// once the Imp leaves.
	{
		Field:    "blood_pact",
		SpellID:  11767,
		Category: "BloodPact",
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
		},
	},
	{
		Field:          "devotion_aura",
		SpellID:        10293,
		Category:       "DevotionAura",
		SharedCategory: "PaladinAura",
		SingleAura:     true,
		Stats: []proto.Stat{
			proto.Stat_StatArmor,
		},
	},
	{
		Field:    "mana_spring_totem",
		SpellID:  10494,
		CastID:   10497,
		Talent:   16187,
		Category: "ManaSpringTotem",
		Stats: []proto.Stat{
			proto.Stat_StatMP5,
		},
	},
	// The cast states the totem's 13 s life and the 5 min cooldown, the aura the mana.
	{
		Field:    "mana_tide_totems",
		SpellID:  1325257,
		CastID:   1325256,
		Kind:     KindExternalCD,
		Category: "ManaTideTotem",
		Stats: []proto.Stat{
			proto.Stat_StatMP5,
		},
		Label: "Mana Tide Totem",
	},
	// Driven because the damage scales with the providing paladin's Holy spell power.
	{
		Field:          "retribution_aura",
		SpellID:        10301,
		Category:       "RetributionAura",
		SharedCategory: "PaladinAura",
		SingleAura:     true,
		Driver:         true,
		Stats: []proto.Stat{
			proto.Stat_StatDefenseRating,
		},
	},
	{
		Field:  "retribution_aura_spell_power",
		Kind:   KindFlag,
		Proto:  ProtoDouble,
		Reason: "the Holy spell power of the paladin providing Retribution Aura, which driveRetributionAura scales the damage with; a sim input with no spell source, rendered under Other Inputs.",
	},
	{
		Field:          "concentration_aura",
		SpellID:        19746,
		Category:       "ConcentrationAura",
		SharedCategory: "PaladinAura",
		SingleAura:     true,
		Stats: []proto.Stat{
			proto.Stat_StatSpellDamage,
			proto.Stat_StatHealingPower,
		},
	},
	{
		Field:   "atiesh_mage",
		SpellID: 28142,
		Owner:   proto.Class_ClassMage,
		Kind:    KindItemCount,
		Label:   "Atiesh - Mage",
		Stats: []proto.Stat{
			proto.Stat_StatCritRating,
		},
	},
	{
		Field:   "atiesh_warlock",
		SpellID: 28143,
		Owner:   proto.Class_ClassWarlock,
		Kind:    KindItemCount,
		Label:   "Atiesh - Warlock",
		Stats: []proto.Stat{
			proto.Stat_StatSpellDamage,
			proto.Stat_StatHealingPower,
		},
	},
	{
		Field:   "atiesh_druid",
		SpellID: 28145,
		Owner:   proto.Class_ClassDruid,
		Kind:    KindItemCount,
		Label:   "Atiesh - Druid",
		Stats: []proto.Stat{
			proto.Stat_StatMP5,
		},
	},
	{
		Field:   "atiesh_priest",
		SpellID: 28144,
		Owner:   proto.Class_ClassPriest,
		Kind:    KindItemCount,
		Label:   "Atiesh - Priest",
		Stats: []proto.Stat{
			proto.Stat_StatHealingPower,
		},
	},
	{
		Field:    "stoneskin_totem",
		SpellID:  10405,
		CastID:   10408,
		Talent:   16258,
		Category: "StoneskinTotem",
		Stats: []proto.Stat{
			proto.Stat_StatDefenseRating,
		},
	},
	// Darkspear 12802's chance on hit lands 16939 on the wielder's party, so the party's copy is up
	// for the share of the fight driveDarkspear states. The wielder's own copy joins the category.
	{
		Field:      "darkspear",
		SpellID:    16939,
		Label:      "Darkspear",
		Category:   "Darkspear",
		SingleAura: true,
		Driver:     true,
		Stats: []proto.Stat{
			proto.Stat_StatCritRating,
		},
	},
	// Admiral's Hat 10030's use lasts 10 min on a 1 min cooldown, so a party member using it keeps
	// it up. The wearer's own copy joins the category.
	{
		Field:      "admirals_hat",
		SpellID:    12022,
		Category:   "AdmiralsHat",
		SingleAura: true,
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
		},
	},
	// A shadow priest's talent 1326159 raises the buff for 10 s on Mind Blast and a full Mind Flay;
	// the party's copy is up for the fight, and the field is the talent's rank. It bids per stat, so a
	// priest's stronger copy holds the Spirit while it is up and the party's holds it again after.
	{
		Field:    "dark_inspiration",
		SpellID:  1326168,
		Talent:   1326159,
		Category: "DarkInspiration",
		Proto:    ProtoInt32,
		Stats: []proto.Stat{
			proto.Stat_StatSpirit,
		},
	},
}

var Individual = []BuffSpec{
	{
		Field:   "greater_blessing_of_kings",
		SpellID: 25898,
		Stats: []proto.Stat{
			proto.Stat_StatAgility,
			proto.Stat_StatIntellect,
			proto.Stat_StatSpirit,
			proto.Stat_StatStamina,
			proto.Stat_StatStrength,
		},
	},
	{
		Field:   "greater_blessing_of_might",
		SpellID: 25782,
		Ranks:   []BuffRank{{SpellID: 25782}, {SpellID: 25916}},
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	{
		Field:   "greater_blessing_of_wisdom",
		SpellID: 25894,
		Ranks:   []BuffRank{{SpellID: 25894}, {SpellID: 25918}},
		Stats: []proto.Stat{
			proto.Stat_StatMP5,
		},
	},
	// The game files no longer have Greater Blessing of Salvation 25895 or Greater Blessing of Light
	// 25890, so the two fields stay in saved settings and do nothing.
	{
		Field:  "greater_blessing_of_salvation",
		Kind:   KindFlag,
		Reason: "Greater Blessing of Salvation 25895 is gone from the game files; the field stays for saved settings and applies nothing.",
	},
	{
		Field:  "greater_blessing_of_light",
		Kind:   KindFlag,
		Reason: "Greater Blessing of Light 25890 is gone from the game files; the field stays for saved settings and applies nothing.",
	},
	{
		Field:    "innervates",
		SpellID:  29166,
		Kind:     KindExternalCD,
		Category: "Innervate",
		Stats: []proto.Stat{
			proto.Stat_StatMP5,
		},
		Label: "Innervates",
	},
	{
		Field:    "power_infusions",
		SpellID:  10060,
		Kind:     KindExternalCD,
		Category: "PowerInfusion",
		Stats: []proto.Stat{
			proto.Stat_StatSpellDamage,
			proto.Stat_StatHealingPower,
		},
		Label: "Power Infusions",
	},
	// Driven because the damage scales with the providing druid's Nature spell power.
	{
		Field:      "thorns",
		SpellID:    9910,
		Category:   "Thorns",
		SingleAura: true,
		Driver:     true,
		Stats: []proto.Stat{
			proto.Stat_StatDefenseRating,
			proto.Stat_StatStamina,
		},
	},
	{
		Field:  "thorns_spell_power",
		Kind:   KindFlag,
		Proto:  ProtoDouble,
		Reason: "the Nature spell power of the druid providing Thorns, which driveThorns scales the damage with; a sim input with no spell source, rendered under Other Inputs.",
	},
}

var Debuffs = []BuffSpec{
	{
		Field:    "hunters_mark",
		SpellID:  14325,
		Category: "HuntersMark",
		Stats: []proto.Stat{
			proto.Stat_StatRangedAttackPower,
		},
	},
	// The Holy damage taken is applied by NewSpellDataJudgementOfTheCrusaderAuraRank in debuffs.go,
	// which the paladin's own ranks share.
	{
		Field:     "judgement_of_the_crusader",
		SpellID:   20303,
		TooltipID: 20308,
		Category:  "Judgement of the Crusader",
		Stats: []proto.Stat{
			proto.Stat_StatHolyDamage,
			proto.Stat_StatSpellDamage,
		},
	},
	{
		Field:     "judgement_of_light",
		SpellID:   20346,
		TooltipID: 20349,
		Kind:      KindProc,
		Stats: []proto.Stat{
			proto.Stat_StatDefenseRating,
		},
	},
	{
		Field:     "judgement_of_wisdom",
		SpellID:   20355,
		TooltipID: 20357,
		Kind:      KindProc,
		Stats: []proto.Stat{
			proto.Stat_StatMP5,
		},
	},
	{
		Field:    "curse_of_elements",
		SpellID:  1311680,
		Category: "CurseOfElements",
		Stats: []proto.Stat{
			proto.Stat_StatSpellDamage,
		},
	},
	{
		Field:    "curse_of_recklessness",
		SpellID:  11717,
		Category: "MinorArmorReduction", // TODO: verify in game that Curse of Recklessness, Faerie Fire and Dust Cloud share one slot.
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	{
		Field:    "faerie_fire",
		SpellID:  9907,
		Category: "MinorArmorReduction", // TODO: verify in game that Curse of Recklessness, Faerie Fire and Dust Cloud share one slot.
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	// The -450 armor is per combo point, so the raid config's debuff is the five-point finisher. A
	// rogue spending fewer needs a driver that prices the aura, and there is none.
	{
		Field:    "expose_armor",
		SpellID:  11198,
		Category: "MajorArmorReduction",
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	// Five stacks are worth nothing until they are there, so a driver ramps them.
	{
		Field:    "sunder_armor",
		SpellID:  11597,
		Category: "MajorArmorReduction",
		Driver:   true,
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	{
		Field:    "gift_of_arthas",
		SpellID:  11374,
		Category: "GiftOfArthasAura",
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
			proto.Stat_StatDefenseRating,
		},
	},
	{
		Field:    "demoralizing_roar",
		SpellID:  9898,
		Category: "Demoralizing",
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
			proto.Stat_StatDefenseRating,
		},
	},
	{
		Field:    "demoralizing_shout",
		SpellID:  11556,
		Category: "Demoralizing",
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
			proto.Stat_StatDefenseRating,
		},
	},
	{
		Field:    "thunder_clap",
		SpellID:  11581,
		Category: "AtkSpdReduction",
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
			proto.Stat_StatDefenseRating,
		},
	},
	{
		Field:    "insect_swarm",
		SpellID:  24977,
		Category: "HitChanceReduction", // TODO: verify in game that Scorpid Sting and Insect Swarm share one slot.
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
			proto.Stat_StatDefenseRating,
		},
	},
	{
		Field:    "scorpid_sting",
		SpellID:  3043,
		Category: "HitChanceReduction", // TODO: verify in game that Scorpid Sting and Insect Swarm share one slot.
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
			proto.Stat_StatDefenseRating,
		},
	},
	// Every flat attack-power reduction on a target bids in one category: the strongest applies.
	{
		Field:    "hex_of_weakness",
		SpellID:  19285,
		Category: "Demoralizing",
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
			proto.Stat_StatDefenseRating,
		},
	},
	{
		Field:    "demoralizing_screech",
		SpellID:  24579,
		Owner:    proto.Class_ClassHunter,
		Category: "Demoralizing",
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
			proto.Stat_StatDefenseRating,
		},
	},
	{
		Field:    "curse_of_weakness",
		SpellID:  11708,
		Category: "CurseOfWeakness",
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
			proto.Stat_StatDefenseRating,
		},
	},
	// The game files state one stack, and a stacking row's amount follows its stacks, so the raid's
	// copy takes it from a driver rather than sitting at zero.
	{
		Field:    "dust_cloud",
		SpellID:  1265904,
		CastID:   1265911,
		Category: "MinorArmorReduction", // TODO: verify in game that Curse of Recklessness, Faerie Fire and Dust Cloud share one slot.
		Driver:   true,
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	// Annihilator's proc keeps the stacks up, so the raid's copy holds all three from the start.
	{
		Field:    "armor_shatter",
		SpellID:  16928,
		Category: "ArmorShatter",
		Driver:   true,
		Stats: []proto.Stat{
			proto.Stat_StatAttackPower,
		},
	},
	// The talent's rank prices the debuff, which the talent states only in its tooltip.
	{
		Field:     "vindication",
		SpellID:   440667,
		TooltipID: 9452,
		Owner:     proto.Class_ClassPaladin,
		Talent:    9452,
		Proto:     ProtoBool,
		Category:  "Demoralizing",
		Stats: []proto.Stat{
			proto.Stat_StatStamina,
			proto.Stat_StatDefenseRating,
		},
	},
	// The resistance is 21992's own; the slow is its triggered 27648, which the driver bids in the
	// attack-speed category on an aura of its own so that a tie there leaves the resistance up.
	{
		Field:   "thunderfury",
		SpellID: 21992,
		Driver:  true,
		Stats: []proto.Stat{
			proto.Stat_StatNatureDamage,
			proto.Stat_StatStamina,
			proto.Stat_StatDefenseRating,
		},
	},
}
