package database

import (
	"regexp"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/tools/database/dbc"
)

// Allows you to ignore certain Spell Effects that the Sim does not support.
// This prevents them from being added to the missing effects list to prevent confusion towards users.
// Empty array means ignore all effects of that type, otherwise it will be ignored
// based on EffectMiscValue_0
var IgnoreSpellEffectByAuraType = map[dbc.EffectAuraType][]int{
	dbcenums.A_MOD_MECHANIC_RESISTANCE: {},
	dbcenums.A_MOD_STEALTH:             {},
	dbcenums.A_MOD_STEALTH_DETECT:      {},
	dbcenums.A_MOD_STEALTH_LEVEL:       {},
	dbcenums.A_MOD_INVISIBILITY:        {},
	dbcenums.A_MOD_INVISIBILITY_DETECT: {},
	dbcenums.A_MOD_SKILL: {
		356, // Fishing Skill
		393, // Skinning Skill
	},
	dbcenums.A_MOD_INCREASE_MOUNTED_SPEED:        {},
	dbcenums.A_MOD_MOUNTED_SPEED_ALWAYS:          {},
	dbcenums.A_MOD_MOUNTED_SPEED_NOT_STACK:       {},
	dbcenums.A_MOD_INCREASE_MOUNTED_FLIGHT_SPEED: {},
	dbcenums.A_MOD_MOUNTED_FLIGHT_SPEED_ALWAYS:   {},
	dbcenums.A_TRANSFORM:                         {},
	dbcenums.A_MECHANIC_IMMUNITY:                 {},
	dbcenums.A_TRACK_CREATURES:                   {},
	dbcenums.A_TRACK_RESOURCES:                   {},
	dbcenums.A_FAR_SIGHT:                         {},
}

// Auras the sim has nothing to simulate for, which put a spell out of scope only where they are all
// it carries: Frostguard's Chilled 16927 slows movement beside a melee slow.
var IgnoreSpellEffectAloneByAuraType = []dbc.EffectAuraType{
	dbcenums.A_MOD_DECREASE_SPEED,
}

var IgnoreSpellEffectBySpellEffectType = map[dbc.SpellEffectType][]int{
	dbcenums.E_CREATE_ITEM:    {},
	dbcenums.E_SUMMON:         {},
	dbcenums.E_TELEPORT_UNITS: {},
}

// Spells that are flavour rather than mechanics, keyed by spell ID so that reporting them as
// missing effects can be suppressed without hiding anything real.
//
// Keyed per spell on purpose. The obvious shortcut - treat an effect made only of A_DUMMY auras
// as noise, which is what MoP does - does not hold in TBC, where A_DUMMY is how a whole class of
// real item effects is encoded: every relic, idol, libram and totem bonus, plus Zandalarian Hero
// Medallion, Thick Obsidian Breastplate and Ashtongue Talisman of Zeal. That rule removed 48
// entries here and only a handful of them were noise.
var IgnoreMissingEffectBySpellID = map[int]string{
	16372: "Seal of Ascension - no tooltip and no mechanic",
}

// Absorbs a server script restricts to spells the game files do not list, keyed by spell ID with the
// reason the generated file gives. 1287808, Onyxia Blood Talisman's, states 10000000000 beside the
// A_DUMMY that names the Dragon Breath spells.
var UnsupportedAbsorbBySpellID = map[int32]string{
	1287808: "the absorb of 10000000000 beside an A_DUMMY absorbs only the spells a script names, which the game files do not list",
}

// Items below MIN_EFFECT_ILVL whose effects are generated anyway, keyed by item ID with the reason a
// level 60 character still uses them.
var EffectFloorExemptions = map[int32]string{
	9449: "Manual Crowd Pummeler: a level 60 melee swaps it in for its on-use attack speed",
}

var OtherItemIdsToFetch = []string{}
var ConsumableOverrides = []*proto.Consumable{}

// Each Atiesh comes from its own class's quest, and nothing in the game files ties a staff to a class:
// AllowableClass is -1 on all four, and no spell creates them. The class follows the party aura each
// one carries.
var ItemOverrides = []*proto.UIItem{
	{
		Id:             22589, // Atiesh, Greatstaff of the Guardian: party spell crit
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	{
		Id:             22630, // Atiesh, Greatstaff of the Guardian: party spell damage
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	{
		Id:             22631, // Atiesh, Greatstaff of the Guardian: party healing
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	{
		Id:             22632, // Atiesh, Greatstaff of the Guardian: party mana regeneration
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	// Naxxramas tier 3 armor comes from class quests that take a token, and the game files leave every
	// piece open to every class: only the set's ring names its class. Each piece takes its ring's class.
	{
		Id:             22416, // Dreadnaught Breastplate
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id:             22417, // Dreadnaught Legplates
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id:             22418, // Dreadnaught Helmet
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id:             22419, // Dreadnaught Pauldrons
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id:             22420, // Dreadnaught Sabatons
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id:             22421, // Dreadnaught Gauntlets
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id:             22422, // Dreadnaught Waistguard
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id:             22423, // Dreadnaught Bracers
		ClassAllowlist: []proto.Class{proto.Class_ClassWarrior},
	},
	{
		Id:             22424, // Redemption Wristguards
		ClassAllowlist: []proto.Class{proto.Class_ClassPaladin},
	},
	{
		Id:             22425, // Redemption Tunic
		ClassAllowlist: []proto.Class{proto.Class_ClassPaladin},
	},
	{
		Id:             22426, // Redemption Handguards
		ClassAllowlist: []proto.Class{proto.Class_ClassPaladin},
	},
	{
		Id:             22427, // Redemption Legguards
		ClassAllowlist: []proto.Class{proto.Class_ClassPaladin},
	},
	{
		Id:             22428, // Redemption Headpiece
		ClassAllowlist: []proto.Class{proto.Class_ClassPaladin},
	},
	{
		Id:             22429, // Redemption Spaulders
		ClassAllowlist: []proto.Class{proto.Class_ClassPaladin},
	},
	{
		Id:             22430, // Redemption Boots
		ClassAllowlist: []proto.Class{proto.Class_ClassPaladin},
	},
	{
		Id:             22431, // Redemption Girdle
		ClassAllowlist: []proto.Class{proto.Class_ClassPaladin},
	},
	{
		Id:             22436, // Cryptstalker Tunic
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	{
		Id:             22437, // Cryptstalker Legguards
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	{
		Id:             22438, // Cryptstalker Headpiece
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	{
		Id:             22439, // Cryptstalker Spaulders
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	{
		Id:             22440, // Cryptstalker Boots
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	{
		Id:             22441, // Cryptstalker Handguards
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	{
		Id:             22442, // Cryptstalker Girdle
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	{
		Id:             22443, // Cryptstalker Wristguards
		ClassAllowlist: []proto.Class{proto.Class_ClassHunter},
	},
	{
		Id:             22476, // Bonescythe Breastplate
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	{
		Id:             22477, // Bonescythe Legplates
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	{
		Id:             22478, // Bonescythe Helmet
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	{
		Id:             22479, // Bonescythe Pauldrons
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	{
		Id:             22480, // Bonescythe Sabatons
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	{
		Id:             22481, // Bonescythe Gauntlets
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	{
		Id:             22482, // Bonescythe Waistguard
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	{
		Id:             22483, // Bonescythe Bracers
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue},
	},
	{
		Id:             22512, // Robe of Faith
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	{
		Id:             22513, // Leggings of Faith
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	{
		Id:             22514, // Circlet of Faith
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	{
		Id:             22515, // Shoulderpads of Faith
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	{
		Id:             22516, // Sandals of Faith
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	{
		Id:             22517, // Gloves of Faith
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	{
		Id:             22518, // Belt of Faith
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	{
		Id:             22519, // Bindings of Faith
		ClassAllowlist: []proto.Class{proto.Class_ClassPriest},
	},
	{
		Id:             22464, // Earthshatter Tunic
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	{
		Id:             22465, // Earthshatter Legguards
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	{
		Id:             22466, // Earthshatter Headpiece
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	{
		Id:             22467, // Earthshatter Spaulders
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	{
		Id:             22468, // Earthshatter Boots
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	{
		Id:             22469, // Earthshatter Handguards
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	{
		Id:             22470, // Earthshatter Girdle
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	{
		Id:             22471, // Earthshatter Wristguards
		ClassAllowlist: []proto.Class{proto.Class_ClassShaman},
	},
	{
		Id:             22496, // Frostfire Robe
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	{
		Id:             22497, // Frostfire Leggings
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	{
		Id:             22498, // Frostfire Circlet
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	{
		Id:             22499, // Frostfire Shoulderpads
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	{
		Id:             22500, // Frostfire Sandals
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	{
		Id:             22501, // Frostfire Gloves
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	{
		Id:             22502, // Frostfire Belt
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	{
		Id:             22503, // Frostfire Bindings
		ClassAllowlist: []proto.Class{proto.Class_ClassMage},
	},
	{
		Id:             22504, // Plagueheart Robe
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	{
		Id:             22505, // Plagueheart Leggings
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	{
		Id:             22506, // Plagueheart Circlet
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	{
		Id:             22507, // Plagueheart Shoulderpads
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	{
		Id:             22508, // Plagueheart Sandals
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	{
		Id:             22509, // Plagueheart Gloves
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	{
		Id:             22510, // Plagueheart Belt
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	{
		Id:             22511, // Plagueheart Bindings
		ClassAllowlist: []proto.Class{proto.Class_ClassWarlock},
	},
	{
		Id:             22488, // Dreamwalker Tunic
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	{
		Id:             22489, // Dreamwalker Legguards
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	{
		Id:             22490, // Dreamwalker Headpiece
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	{
		Id:             22491, // Dreamwalker Spaulders
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	{
		Id:             22492, // Dreamwalker Boots
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	{
		Id:             22493, // Dreamwalker Handguards
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	{
		Id:             22494, // Dreamwalker Girdle
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
	{
		Id:             22495, // Dreamwalker Wristguards
		ClassAllowlist: []proto.Class{proto.Class_ClassDruid},
	},
}

// Keep these sorted by item ID.
var ItemAllowList = map[int32]struct{}{
	1168:  {}, // Skullflame Shield
	2140:  {},
	2505:  {},
	8345:  {}, // Wolfshead Helm
	11815: {}, // Hand of Justice
	18168: {}, // Force Reactive Disk
	19337: {}, // The Black Book
}

// Keep these sorted by item ID.
var ItemDenyList = map[int32]struct{}{
	17782: {}, // talisman of the binding shard
	17783: {}, // talisman of the binding fragment
	17802: {}, // Deprecated version of Thunderfury
	18582: {},
	18583: {},
	18584: {},
	20003: {}, // Devilsaur Claws
	20005: {}, // Devilsaur Claws
	22736: {},
	23363: {}, // Titanic Breastplate
	24265: {},
	// Season of Discovery items the game files still carry. Forever gives what it keeps a new id.
	202256: {}, // Privateer's Ornate Pistol
	205218: {}, // Libram of Discovery
	205420: {}, // Libram of Judgement
	206376: {}, // Bulwark Icon
	206381: {}, // Dyadic Icon
	206382: {}, // Tempest Icon
	206383: {}, // Lithic Icon
	206384: {}, // Astral Icon
	206385: {}, // Clastic Icon
	206386: {}, // Galvanic Icon
	206387: {}, // Kajaric Icon
	206388: {}, // Sulfurous Icon
	206389: {}, // Brimstone Icon
	206390: {}, // Ichthyan Icon
	206391: {}, // Syzygic Icon
	206954: {}, // Idol of Ursine Rage
	208222: {}, // Old Guard Retaliator
	208414: {}, // Lunar Idol
	208689: {}, // Ferocious Idol
	208849: {}, // Libram of Blessings
	208851: {}, // Libram of Justice
	210195: {}, // Unbalanced Idol
	210534: {}, // Idol of the Wild
	210781: {}, // Phoenix Bindings
	211472: {}, // Libram of Banishment
	211500: {}, // Resilient Cloth Headband
	211856: {}, // Resilient Mail Coif
	211857: {}, // Resilient Leather Mask
	211940: {}, // Ecks'av's Tribal Guardian
	211941: {}, // Windwalker's Yari
	212580: {}, // Lorekeeper's Staff
	212581: {}, // Outrunner's Bow
	212582: {}, // Protector's Sword
	212583: {}, // Sentinel's Blade
	212584: {}, // Advisor's Gnarled Staff
	212585: {}, // Outrider's Bow
	212586: {}, // Legionnaire's Sword
	212587: {}, // Scout's Blade
	213105: {}, // Shawarmageddon
	213391: {}, // UNUSED Electrified Mithril Gauntlets
	213513: {}, // Libram of Deliverance
	213594: {}, // Idol of the Heckler
	215112: {}, // UNUSED - Cloak of Arcane Insulation
	215113: {}, // UNUSED - Razor-Lined Shoulderpads
	215365: {}, // Invoker's Mantle
	215366: {}, // Invoker's Cord
	217270: {}, // Turtle Scale Gloves
	217273: {}, // Golden Scale Gauntlets
	217287: {}, // Greater Mystic Wand
	217314: {}, // Moonsight Rifle
	217315: {}, // Precision Bow
	219345: {}, // Infernal Lasso
	220606: {}, // Idol of the Dream
	220794: {}, // Knight's Plate Hauberk
	220795: {}, // Knight-Lieutenant's Plate Pauldrons
	220796: {}, // Blood Guard's Plate Pauldrons
	220797: {}, // Knight's Plate Leggings
	220798: {}, // Stone Guard's Plate Leggings
	220799: {}, // Sergeant Major's Plate Greaves
	220800: {}, // First Sergeant's Plate Greaves
	220801: {}, // Stone Guard's Plate Armor
	220803: {}, // Blood Guard's Plate Helm
	220804: {}, // Knight-Lieutenant's Plate Helm
	220806: {}, // Sergeant Major's Plate Gauntlets
	220807: {}, // First Sergeant's Plate Gauntlets
	220808: {}, // Knight-Lieutenant's Imbued Pauldrons
	220809: {}, // Knight's Imbued Leggings
	220810: {}, // Knight-Lieutenant's Imbued Helmet
	220811: {}, // Sergeant Major's Imbued Greaves
	220812: {}, // Sergeant Major's Imbued Gauntlets
	220813: {}, // Knight's Imbued Armor
	220814: {}, // Sergeant Major's Lamellar Boots
	220815: {}, // Knight's Lamellar Chestplate
	220816: {}, // Knight's Lamellar Legplates
	220817: {}, // Sergeant Major's Lamellar Gauntlets
	220818: {}, // Knight-Lieutenant's Lamellar Pauldrons
	220819: {}, // Knight-Lieutenant's Lamellar Helm
	220820: {}, // Blood Guard's Mail Helmet
	220821: {}, // Blood Guard's Chain Helmet
	220822: {}, // Knight-Lieutenant's Chain Helmet
	220823: {}, // Blood Guard's Mail Epaulets
	220824: {}, // Blood Guard's Chain Epaulets
	220825: {}, // Knight-Lieutenant's Chain Epaulets
	220826: {}, // Stone Guard's Mail Armor
	220827: {}, // Stone Guard's Chain Armor
	220828: {}, // Knight's Chain Armor
	220829: {}, // Sergeant Major's Chain Gauntlets
	220830: {}, // First Sergeant's Chain Gauntlets
	220831: {}, // First Sergeant's Mail Gauntlets
	220832: {}, // Knight's Chain Legplates
	220833: {}, // Stone Guard's Chain Legplates
	220834: {}, // Stone Guard's Mail Legplates
	220835: {}, // First Sergeant's Mail Sabatons
	220836: {}, // First Sergeant's Chain Sabatons
	220837: {}, // Sergeant Major's Chain Sabatons
	220838: {}, // Stone Guard's Inscribed Chestpiece
	220839: {}, // Stone Guard's Inscribed Legplates
	220840: {}, // First Sergeant's Inscribed Sabatons
	220841: {}, // Blood Guard's Inscribed Shoulder Pads
	220842: {}, // Blood Guard's Inscribed Skullcap
	220843: {}, // First Sergeant's Inscribed Gauntlets
	220844: {}, // Stone Guard's Pulsing Breastplate
	220845: {}, // First Sergeant's Pulsing Gauntlets
	220846: {}, // First Sergeant's Pulsing Greaves
	220847: {}, // Stone Guard's Pulsing Legplates
	220848: {}, // Blood Guard's Pulsing Helmet
	220849: {}, // Blood Guard's Pulsing Shoulders
	220850: {}, // Knight-Lieutenant's Leather Headband
	220851: {}, // Blood Guard's Leather Headband
	220852: {}, // Knight-Lieutenant's Leather Shoulders
	220853: {}, // Blood Guard's Leather Shoulders
	220854: {}, // Knight's Leather Armor
	220855: {}, // Stone Guard's Leather Armor
	220856: {}, // Sergeant Major's Leather Gauntlets
	220857: {}, // First Sergeant's Leather Gauntlets
	220858: {}, // Knight's Leather Pants
	220859: {}, // Stone Guard's Leather Pants
	220860: {}, // Sergeant Major's Leather Boots
	220861: {}, // First Sergeant's Leather Boots
	220862: {}, // Sergeant Major's Crackling Leather Boots
	220863: {}, // First Sergeant's Crackling Leather Boots
	220864: {}, // Knight's Crackling Leather Leggings
	220865: {}, // Stone Guard's Crackling Leather Leggings
	220866: {}, // Sergeant Major's Crackling Leather Gauntlets
	220867: {}, // First Sergeant's Crackling Leather Gauntlets
	220868: {}, // Knight's Crackling Leather Tunic
	220869: {}, // Stone Guard's Crackling Leather Tunic
	220870: {}, // Knight-Lieutenant's Crackling Leather Spaulders
	220871: {}, // Blood Guard's Crackling Leather Spaulders
	220872: {}, // Knight-Lieutenant's Crackling Leather Helmet
	220873: {}, // Blood Guard's Crackling Leather Helmet
	220874: {}, // Knight-Lieutenant's Restored Leather Helm
	220875: {}, // Blood Guard's Restored Leather Helm
	220876: {}, // Knight-Lieutenant's Restored Leather Spaulders
	220877: {}, // Blood Guard's Restored Leather Spaulders
	220878: {}, // Knight's Restored Leather Jerkin
	220879: {}, // Stone Guard's Restored Leather Jerkin
	220880: {}, // Sergeant Major's Restored Leather Gloves
	220881: {}, // First Sergeant's Restored Leather Gloves
	220882: {}, // Knight's Restored Leather Leggings
	220883: {}, // Stone Guard's Restored Leather Leggings
	220884: {}, // Sergeant Major's Restored Leather Boots
	220885: {}, // First Sergeant's Restored Leather Boots
	220886: {}, // Knight's Dreadweave Vest
	220887: {}, // Knight-Lieutenant's Dreadweave Mantle
	220888: {}, // Knight's Dreadweave Leggings
	220889: {}, // Knight-Lieutenant's Dreadweave Hat
	220890: {}, // Sergeant Major's Dreadweave Gloves
	220891: {}, // Sergeant Major's Dreadweave Boots
	220892: {}, // Knight's Satin Armor
	220893: {}, // Knight's Satin Leggings
	220894: {}, // Knight-Lieutenant's Satin Pads
	220895: {}, // Sergeant Major's Satin Boots
	220896: {}, // Knight-Lieutenant's Satin Cover
	220897: {}, // Sergeant Major's Satin Gloves
	220898: {}, // First Sergeant's Satin Gloves
	220899: {}, // Blood Guard's Satin Cover
	220900: {}, // First Sergeant's Satin Boots
	220901: {}, // Blood Guard's Satin Pads
	220902: {}, // Stone Guard's Satin Leggings
	220903: {}, // Stone Guard's Satin Armor
	220904: {}, // Stone Guard's Dreadweave Vest
	220905: {}, // Blood Guard's Dreadweave Mantle
	220906: {}, // Stone Guard's Dreadweave Leggings
	220907: {}, // Blood Guard's Dreadweave Hat
	220908: {}, // First Sergeant's Dreadweave Gloves
	220909: {}, // First Sergeant's Dreadweave Boots
	220915: {}, // Idol of the Raging Shambler
	221344: {}, // Personal Spellbook
	221518: {}, // Whisper
	223073: {}, // Knight-Lieutenant's Mail Epaulets
	223074: {}, // Knight's Mail Legplates
	223075: {}, // Knight-Lieutenant's Mail Helmet
	223076: {}, // Sergeant Major's Mail Gauntlets
	223077: {}, // Sergeant Major's Mail Sabatons
	223078: {}, // Knight's Mail Armor
	225838: {}, // Voltaic Icon
	226772: {}, // Feralheart Sash
	226773: {}, // Feralheart Cowl
	226774: {}, // Feralheart Galoshes
	226775: {}, // Feralheart Wraps
	226776: {}, // Feralheart Vest
	226777: {}, // Feralheart Hands
	226778: {}, // Feralheart Spaulders
	226779: {}, // Feralheart Kilt
	226780: {}, // Feralheart Cord
	226781: {}, // Feralheart Sandals
	226782: {}, // Feralheart Bindings
	226783: {}, // Feralheart Embrace
	226784: {}, // Feralheart Gauntlets
	226785: {}, // Feralheart Mantle
	226786: {}, // Feralheart Headdress
	226787: {}, // Feralheart Pants
	226788: {}, // Feralheart Bands
	226789: {}, // Feralheart Girdle
	226790: {}, // Feralheart Epaulets
	226791: {}, // Feralheart Trousers
	226792: {}, // Feralheart Cap
	226793: {}, // Feralheart Fists
	226794: {}, // Feralheart Walkers
	226795: {}, // Feralheart Tunic
	226796: {}, // Feralheart Wristguards
	226797: {}, // Feralheart Waistguard
	226798: {}, // Feralheart Pauldrons
	226799: {}, // Feralheart Legguards
	226801: {}, // Feralheart Faceguard
	226802: {}, // Feralheart Grips
	226803: {}, // Feralheart Treads
	226804: {}, // Feralheart Armor
	226825: {}, // Darkmantle Tunic
	226826: {}, // Darkmantle Spaulders
	226827: {}, // Darkmantle Pants
	226828: {}, // Darkmantle Grips
	226829: {}, // Darkmantle Cap
	226830: {}, // Darkmantle Bracers
	226831: {}, // Darkmantle Footpads
	226832: {}, // Darkmantle Belt
	226857: {}, // Battleboots of Heroism
	226858: {}, // Spaulders of Heroism
	226859: {}, // Legplates of Heroism
	226860: {}, // Crown of Heroism
	226861: {}, // Gauntlets of Heroism
	226862: {}, // Breastplate of Heroism
	226863: {}, // Bracers of Heroism
	226864: {}, // Belt of Heroism
	226865: {}, // Wristguards of Heroism
	226866: {}, // Waistguard of Heroism
	226867: {}, // Pauldrons of Heroism
	226868: {}, // Legguards of Heroism
	226869: {}, // Faceguard of Heroism
	226870: {}, // Handguards of Heroism
	226871: {}, // Sabatons of Heroism
	226872: {}, // Chestguard of Heroism
	226881: {}, // Beastmaster's Treads
	226882: {}, // Beastmaster's Pants
	226883: {}, // Beastmaster's Gauntlets
	226884: {}, // Beastmaster's Mantle
	226885: {}, // Beastmaster's Bindings
	226886: {}, // Beastmaster's Tunic
	226887: {}, // Beastmaster's Cap
	226888: {}, // Beastmaster's Belt
	226905: {}, // Deathmist Belt
	226906: {}, // Deathmist Robe
	226907: {}, // Deathmist Bracers
	226908: {}, // Deathmist Sandals
	226909: {}, // Deathmist Mask
	226910: {}, // Deathmist Leggings
	226911: {}, // Deathmist Wraps
	226912: {}, // Deathmist Mantle
	226929: {}, // Sorcerer's Bindings
	226930: {}, // Sorcerer's Gauntlets
	226931: {}, // Sorcerer's Sandals
	226932: {}, // Sorcerer's Robes
	226933: {}, // Sorcerer's Leggings
	226934: {}, // Sorcerer's Belt
	226935: {}, // Sorcerer's Crown
	226936: {}, // Sorcerer's Mantle
	226945: {}, // Virtuous Robe
	226946: {}, // Virtuous Skirt
	226947: {}, // Virtuous Crown
	226948: {}, // Virtuous Belt
	226949: {}, // Virtuous Bracers
	226950: {}, // Virtuous Mitts
	226951: {}, // Virtuous Mantle
	226952: {}, // Virtuous Sandals
	226953: {}, // Virtuous Wraps
	226954: {}, // Virtuous Cord
	226955: {}, // Virtuous Epaulets
	226956: {}, // Virtuous Leggings
	226957: {}, // Virtuous Cowl
	226958: {}, // Virtuous Hands
	226959: {}, // Virtuous Slippers
	226960: {}, // Virtuous Gown
	226969: {}, // Soulforge Spaulders
	226970: {}, // Soulforge Bracers
	226971: {}, // Soulforge Belt
	226972: {}, // Soulforge Legplates
	226973: {}, // Soulforge Breastplate
	226974: {}, // Soulforge Warboots
	226975: {}, // Soulforge Gauntlets
	226976: {}, // Soulforge Greathelm
	226977: {}, // Soulforge Bindings
	226978: {}, // Soulforge Cord
	226979: {}, // Soulforge Epaulets
	226980: {}, // Soulforge Leggings
	226981: {}, // Soulforge Crown
	226982: {}, // Soulforge Fists
	226983: {}, // Soulforge Treads
	226984: {}, // Soulforge Embrace
	226985: {}, // Soulforge Wristguards
	226986: {}, // Soulforge Waistguard
	226987: {}, // Soulforge Pauldrons
	226988: {}, // Soulforge Legguards
	226989: {}, // Soulforge Faceguard
	226990: {}, // Soulforge Handguards
	226991: {}, // Soulforge Sabatons
	226992: {}, // Soulforge Chestguards
	227001: {}, // Bindings of The Five Thunders
	227002: {}, // Coif of The Five Thunders
	227003: {}, // Pauldrons of The Five Thunders
	227004: {}, // Vest of The Five Thunders
	227005: {}, // Kilt of The Five Thunders
	227006: {}, // Gauntlets of The Five Thunders
	227007: {}, // Slippers of The Five Thunders
	227008: {}, // Cord of The Five Thunders
	227009: {}, // Bracers of The Five Thunders
	227010: {}, // Sash of The Five Thunders
	227011: {}, // Mantle of The Five Thunders
	227012: {}, // Leggings of The Five Thunders
	227013: {}, // Crown of The Five Thunders
	227014: {}, // Grasp of The Five Thunders
	227015: {}, // Greaves of The Five Thunders
	227016: {}, // Tunic of The Five Thunders
	227017: {}, // Bands of The Five Thunders
	227018: {}, // Girdle of The Five Thunders
	227019: {}, // Spaulders of The Five Thunders
	227020: {}, // Legplates of The Five Thunders
	227021: {}, // Face of The Five Thunders
	227022: {}, // Fists of The Five Thunders
	227023: {}, // Treads of The Five Thunders
	227024: {}, // Chain of The Five Thunders
	227042: {}, // Champion's Plate Shoulders
	227043: {}, // Champion's Plate Helm
	227044: {}, // Lieutenant Commander's Plate Helm
	227045: {}, // Lieutenant Commander's Plate Shoulders
	227046: {}, // Knight-Captain's Plate Hauberk
	227047: {}, // Knight-Captain's Plate Leggings
	227048: {}, // Legionnaire's Plate Leggings
	227049: {}, // Legionnaire's Plate Hauberk
	227050: {}, // Blood Guard's Plate Gauntlets
	227051: {}, // Blood Guard's Plate Greaves
	227052: {}, // Knight-Lieutenant's Plate Greaves
	227053: {}, // Knight-Lieutenant's Plate Gauntlets
	227054: {}, // Lieutenant Commander's Leather Shoulders
	227055: {}, // Lieutenant Commander's Leather Helm
	227056: {}, // Champion's Leather Shoulders
	227057: {}, // Champion's Leather Helm
	227058: {}, // Knight-Captain's Leather Chestpiece
	227059: {}, // Legionnaire's Leather Legguards
	227060: {}, // Legionnaire's Leather Chestpiece
	227061: {}, // Knight-Captain's Leather Legguards
	227062: {}, // Blood Guard's Leather Walkers
	227063: {}, // Blood Guard's Leather Grips
	227064: {}, // Knight-Lieutenant's Leather Walkers
	227065: {}, // Knight-Lieutenant's Leather Grips
	227066: {}, // Lieutenant Commander's Chain Helm
	227067: {}, // Champion's Chain Helm
	227068: {}, // Lieutenant Commander's Chain Shoulders
	227069: {}, // Champion's Chain Shoulders
	227070: {}, // Knight-Captain's Chain Hauberk
	227071: {}, // Legionnaire's Chain Hauberk
	227072: {}, // Knight-Captain's Chain Legguards
	227073: {}, // Legionnaire's Chain Legguards
	227074: {}, // Blood Guard's Chain Greaves
	227075: {}, // Blood Guard's Chain Vices
	227076: {}, // Knight-Lieutenant's Chain Greaves
	227077: {}, // Knight-Lieutenant's Chain Vices
	227078: {}, // Champion's Chain Pauldrons
	227079: {}, // Legionnaire's Chain Legplates
	227080: {}, // Champion's Chain Greathelm
	227081: {}, // Blood Guard's Chain Grips
	227082: {}, // Blood Guard's Chain Sabatons
	227083: {}, // Legionnaire's Chain Armor
	227084: {}, // Lieutenant Commander's Chain Pauldrons
	227085: {}, // Knight-Captain's Chain Legplates
	227086: {}, // Lieutenant Commander's Chain Greathelm
	227087: {}, // Knight-Lieutenant's Chain Grips
	227088: {}, // Knight-Lieutenant's Chain Sabatons
	227089: {}, // Knight-Captain's Chain Armor
	227090: {}, // Champion's Dreadweave Cowl
	227091: {}, // Lieutenant Commander's Dreadweave Spaulders
	227092: {}, // Champion's Dreadweave Spaulders
	227093: {}, // Lieutenant Commander's Dreadweave Cowl
	227094: {}, // Legionnaire's Dreadweave Tunic
	227095: {}, // Knight-Captain's Dreadweave Legguards
	227096: {}, // Knight-Captain's Dreadweave Tunic
	227097: {}, // Legionnaire's Dreadweave Legguards
	227098: {}, // Blood Guard's Dreadweave Walkers
	227099: {}, // Blood Guard's Dreadweave Handwraps
	227100: {}, // Knight-Lieutenant's Dreadweave Handwraps
	227101: {}, // Knight-Lieutenant's Dreadweave Walkers
	227102: {}, // Lieutenant Commander's Silk Mantle
	227103: {}, // Lieutenant Commander's Silk Cowl
	227104: {}, // Champion's Silk Mantle
	227105: {}, // Champion's Silk Cowl
	227106: {}, // Legionnaire's Silk Tunic
	227107: {}, // Legionnaire's Silk Legguards
	227108: {}, // Knight-Captain's Silk Tunic
	227109: {}, // Knight-Captain's Silk Legguards
	227110: {}, // Blood Guard's Silk Walkers
	227111: {}, // Blood Guard's Silk Handwraps
	227112: {}, // Knight-Lieutenant's Silk Walkers
	227113: {}, // Knight-Lieutenant's Silk Handwraps
	227114: {}, // Knight-Lieutenant's Silk Gauntlets
	227115: {}, // Blood Guard's Silk Gauntlets
	227118: {}, // Champion's Satin Hood
	227119: {}, // Lieutenant Commander's Satin Mantle
	227120: {}, // Champion's Satin Mantle
	227121: {}, // Lieutenant Commander's Satin Hood
	227122: {}, // Knight-Captain's Satin Tunic
	227123: {}, // Legionnaire's Satin Legguards
	227124: {}, // Legionnaire's Satin Tunic
	227125: {}, // Knight-Captain's Satin Legguards
	227126: {}, // Blood Guard's Satin Handwraps
	227127: {}, // Blood Guard's Satin Walkers
	227128: {}, // Knight-Lieutenant's Satin Handwraps
	227129: {}, // Knight-Lieutenant's Satin Walkers
	227130: {}, // Champion's Satin Epaulets
	227131: {}, // Legionnaire's Satin Leggings
	227132: {}, // Champion's Satin Crown
	227133: {}, // Blood Guard's Satin Grips
	227134: {}, // Blood Guard's Satin Treads
	227135: {}, // Legionnaire's Satin Robe
	227136: {}, // Lieutenant Commander's Satin Epaulets
	227137: {}, // Knight-Captain's Satin Leggings
	227138: {}, // Lieutenant Commander's Satin Crown
	227139: {}, // Knight-Lieutenant's Satin Grips
	227140: {}, // Knight-Lieutenant's Satin Treads
	227141: {}, // Knight-Captain's Satin Robe
	227142: {}, // Knight-Captain's Lamellar Breastplate
	227143: {}, // Knight-Captain's Lamellar Leggings
	227144: {}, // Lieutenant Commander's Lamellar Headguard
	227145: {}, // Lieutenant Commander's Lamellar Shoulders
	227146: {}, // Knight-Lieutenant's Lamellar Sabatons
	227147: {}, // Knight-Lieutenant's Lamellar Gauntlets
	227148: {}, // Lieutenant Commander's Lamellar Pauldrons
	227149: {}, // Lieutenant Commander's Lamellar Helmet
	227150: {}, // Knight-Captain's Lamellar Legguards
	227151: {}, // Knight-Captain's Lamellar Chestplate
	227152: {}, // Knight-Lieutenant's Lamellar Gloves
	227153: {}, // Knight-Lieutenant's Lamellar Greaves
	227154: {}, // Champion's Mail Pauldrons
	227155: {}, // Champion's Mail Headguard
	227156: {}, // Legionnaire's Mail Legguards
	227157: {}, // Legionnaire's Mail Hauberk
	227158: {}, // Blood Guard's Mail Greaves
	227159: {}, // Blood Guard's Mail Vices
	227160: {}, // Champion's Mail Spaulders
	227161: {}, // Legionnaire's Mail Leggings
	227162: {}, // Champion's Mail Helm
	227163: {}, // Blood Guard's Mail Gauntlets
	227164: {}, // Blood Guard's Mail Sabatons
	227165: {}, // Legionnaire's Mail Breastplate
	227166: {}, // Champion's Mail Epaulets
	227167: {}, // Legionnaire's Mail Pants
	227168: {}, // Champion's Mail Skullcap
	227169: {}, // Blood Guard's Mail Gloves
	227170: {}, // Blood Guard's Mail Boots
	227171: {}, // Legionnaire's Mail Chestguard
	227172: {}, // Lieutenant Commander's Dragonhide Shoulders
	227173: {}, // Lieutenant Commander's Dragonhide Headguard
	227174: {}, // Champion's Dragonhide Headguard
	227175: {}, // Champion's Dragonhide Shoulders
	227176: {}, // Knight-Captain's Dragonhide Chestpiece
	227177: {}, // Legionnaire's Dragonhide Leggings
	227178: {}, // Knight-Captain's Dragonhide Leggings
	227179: {}, // Legionnaire's Dragonhide Chestpiece
	227180: {}, // Blood Guard's Dragonhide Grips
	227181: {}, // Blood Guard's Dragonhide Treads
	227182: {}, // Knight-Lieutenant's Dragonhide Treads
	227183: {}, // Knight-Lieutenant's Dragonhide Grips
	227184: {}, // Champion's Dragonhide Spaulders
	227185: {}, // Legionnaire's Dragonhide Pants
	227186: {}, // Champion's Dragonhide Helm
	227187: {}, // Blood Guard's Dragonhide Gloves
	227188: {}, // Blood Guard's Dragonhide Boots
	227189: {}, // Legionnaire's Dragonhide Armor
	227190: {}, // Lieutenant Commander's Dragonhide Spaulders
	227191: {}, // Knight-Captain's Dragonhide Pants
	227192: {}, // Lieutenant Commander's Dragonhide Helm
	227193: {}, // Knight-Lieutenant's Dragonhide Gloves
	227194: {}, // Knight-Lieutenant's Dragonhide Boots
	227195: {}, // Knight-Captain's Dragonhide Armor
	227196: {}, // Knight-Captain's Dragonhide Tunic
	227197: {}, // Knight-Lieutenant's Dragonhide Greaves
	227198: {}, // Knight-Lieutenant's Dragonhide Gauntlets
	227199: {}, // Lieutenant Commander's Dragonhide Headdress
	227200: {}, // Knight-Captain's Dragonhide Legguards
	227201: {}, // Lieutenant Commander's Dragonhide Pauldrons
	227202: {}, // Legionnaire's Dragonhide Tunic
	227203: {}, // Blood Guard's Dragonhide Greaves
	227204: {}, // Blood Guard's Dragonhide Gauntlets
	227205: {}, // Champion's Dragonhide Headdress
	227206: {}, // Legionnaire's Dragonhide Legguards
	227207: {}, // Champion's Dragonhide Pauldrons
	227444: {}, // Idol of the Huntress
	227803: {}, // Dire Warbear Harness
	227804: {}, // Dire Warbear Woolies
	227805: {}, // Ferocity of the Timbermaw
	227807: {}, // Dense Timbermaw Belt
	227808: {}, // Rugged Mantle of the Timbermaw
	227809: {}, // Studded Timbermaw Brawlers
	227810: {}, // Dense Timbermaw Boots
	227814: {}, // Radiant Girdle of the Dawn
	227815: {}, // Fine Dawn Treaders
	227816: {}, // Argent Elite Boots
	227817: {}, // Radiant Gloves of the Dawn
	227818: {}, // Glowing Mantle of the Dawn
	227859: {}, // Shimmering Dawnbringer Shoulders
	227862: {}, // Incandescent Mooncloth Boots
	227888: {}, // Argent Elite Shoulders
	227947: {}, // Golem Fitted Pauldrons
	228175: {}, // Libram of Holy Alacrity
	228176: {}, // Totem of Thunder
	228190: {}, // Knowledge of the Timbermaw
	231530: {}, // Warlord's Plate Armor
	231531: {}, // General's Plate Boots
	231532: {}, // General's Plate Gauntlets
	231533: {}, // General's Plate Leggings
	231534: {}, // Warlord's Plate Shoulders
	231535: {}, // Warlord's Plate Headpiece
	231536: {}, // Field Marshal's Plate Armor
	231537: {}, // Field Marshal's Plate Shoulderguards
	231538: {}, // Field Marshal's Plate Helm
	231539: {}, // Marshal's Plate Boots
	231540: {}, // Marshal's Plate Legguards
	231541: {}, // Marshal's Plate Gauntlets
	231543: {}, // Field Marshal's Leather Chestpiece
	231544: {}, // Marshal's Leather Handgrips
	231545: {}, // Field Marshal's Leather Mask
	231546: {}, // Marshal's Leather Footguards
	231547: {}, // Field Marshal's Leather Epaulets
	231548: {}, // Marshal's Leather Leggings
	231549: {}, // Warlord's Leather Breastplate
	231551: {}, // Warlord's Leather Spaulders
	231552: {}, // General's Leather Treads
	231553: {}, // Warlord's Leather Helm
	231554: {}, // General's Leather Legguards
	231555: {}, // General's Leather Mitts
	231557: {}, // Field Marshal's Chain Pauldrons
	231558: {}, // Marshal's Chain Legplates
	231560: {}, // Marshal's Chain Grips
	231561: {}, // Marshal's Chain Sabatons
	231562: {}, // Field Marshal's Chain Greathelm
	231563: {}, // Field Marshal's Chain Armor
	231564: {}, // General's Chain Sabatons
	231565: {}, // Warlord's Chain Pauldrons
	231566: {}, // Warlord's Chain Armor
	231567: {}, // General's Chain Legplates
	231568: {}, // Warlord's Chain Greathelm
	231569: {}, // General's Chain Grips
	231570: {}, // General's Chain Greaves
	231571: {}, // Warlord's Chain Helm
	231572: {}, // Warlord's Chain Shoulders
	231573: {}, // Warlord's Chain Hauberk
	231574: {}, // General's Chain Legguards
	231575: {}, // General's Chain Vices
	231576: {}, // Field Marshal's Chain Shoulders
	231577: {}, // Marshal's Chain Legguards
	231578: {}, // Marshal's Chain Vices
	231579: {}, // Marshal's Chain Greaves
	231580: {}, // Field Marshal's Chain Helm
	231581: {}, // Field Marshal's Chain Hauberk
	231582: {}, // Field Marshal's Dreadweave Robe
	231583: {}, // Field Marshal's Dreadweave Shoulders
	231584: {}, // Field Marshal's Coronal
	231585: {}, // Marshal's Dreadweave Boots
	231586: {}, // Marshal's Dreadweave Gloves
	231587: {}, // Marshal's Dreadweave Leggings
	231588: {}, // General's Dreadweave Pants
	231589: {}, // General's Dreadweave Gloves
	231590: {}, // Warlord's Dreadweave Hood
	231591: {}, // Warlord's Dreadweave Robe
	231592: {}, // Warlord's Dreadweave Mantle
	231593: {}, // General's Dreadweave Boots
	231594: {}, // Warlord's Silk Amice
	231595: {}, // General's Silk Trousers
	231596: {}, // Warlord's Silk Raiment
	231597: {}, // General's Silk Boots
	231599: {}, // General's Silk Gauntlets
	231600: {}, // General's Silk Handwraps
	231601: {}, // Warlord's Silk Cowl
	231602: {}, // Field Marshal's Silk Spaulders
	231603: {}, // Field Marshal's Silk Vestments
	231604: {}, // Field Marshal's Coronet
	231605: {}, // Marshal's Silk Leggings
	231606: {}, // Marshal's Silk Footwraps
	231608: {}, // Marshal's Silk Gauntlets
	231609: {}, // Marshal's Silk Handwraps
	231610: {}, // General's Satin Treads
	231611: {}, // Warlord's Satin Epaulets
	231612: {}, // Warlord's Satin Robes
	231613: {}, // General's Satin Grips
	231614: {}, // General's Satin Leggings
	231615: {}, // Warlord's Satin Crown
	231616: {}, // Field Marshal's Satin Crown
	231617: {}, // Marshal's Satin Grips
	231618: {}, // Field Marshal's Satin Robe
	231619: {}, // Marshal's Satin Leggings
	231620: {}, // Marshal's Satin Treads
	231621: {}, // Field Marshal's Satin Epaulets
	231622: {}, // Field Marshal's Satin Hood
	231623: {}, // Marshal's Satin Handwraps
	231624: {}, // Field Marshal's Satin Tunic
	231626: {}, // Marshal's Satin Legguards
	231627: {}, // Marshal's Satin Walkers
	231628: {}, // Field Marshal's Satin Mantle
	231630: {}, // General's Satin Walkers
	231631: {}, // Warlord's Satin Mantle
	231632: {}, // Warlord's Satin Tunic
	231633: {}, // General's Satin Handwraps
	231634: {}, // General's Satin Legguards
	231635: {}, // Warlord's Satin Hood
	231639: {}, // Marshal's Lamellar Legguards
	231640: {}, // Field Marshal's Lamellar Helmet
	231641: {}, // Field Marshal's Lamellar Chestplate
	231643: {}, // Marshal's Lamellar Gloves
	231645: {}, // Field Marshal's Lamellar Pauldrons
	231646: {}, // Marshal's Lamellar Greaves
	231647: {}, // Marshal's Lamellar Leggings
	231648: {}, // Field Marshal's Lamellar Headguard
	231649: {}, // Field Marshal's Lamellar Breastplate
	231650: {}, // Marshal's Lamellar Gauntlets
	231651: {}, // Field Marshal's Lamellar Shoulders
	231652: {}, // Marshal's Lamellar Sabatons
	231653: {}, // Warlord's Mail Hauberk
	231654: {}, // Warlord's Mail Pauldrons
	231655: {}, // General's Mail Vices
	231656: {}, // General's Mail Greaves
	231657: {}, // Warlord's Mail Headguard
	231658: {}, // General's Mail Legguards
	231659: {}, // Warlord's Mail Spaulders
	231660: {}, // General's Mail Gauntlets
	231661: {}, // General's Mail Sabatons
	231662: {}, // Warlord's Mail Breastplate
	231663: {}, // Warlord's Mail Helm
	231664: {}, // General's Mail Leggings
	231665: {}, // Warlord's Mail Epaulets
	231666: {}, // General's Mail Gloves
	231667: {}, // General's Mail Boots
	231668: {}, // Warlord's Mail Chestguard
	231669: {}, // Warlord's Mail Skullcap
	231670: {}, // General's Mail Pants
	231671: {}, // General's Dragonhide Greaves
	231672: {}, // Warlord's Dragonhide Pauldrons
	231673: {}, // General's Dragonhide Legguards
	231674: {}, // Warlord's Dragonhide Tunic
	231675: {}, // Warlord's Dragonhide Headdress
	231676: {}, // General's Dragonhide Gauntlets
	231677: {}, // General's Dragonhide Gloves
	231678: {}, // Warlord's Dragonhide Helm
	231679: {}, // Warlord's Dragonhide Armor
	231680: {}, // General's Dragonhide Pants
	231681: {}, // Warlord's Dragonhide Spaulders
	231682: {}, // General's Dragonhide Boots
	231683: {}, // General's Dragonhide Treads
	231684: {}, // Warlord's Dragonhide Shoulders
	231685: {}, // General's Dragonhide Leggings
	231686: {}, // Warlord's Dragonhide Chestpiece
	231687: {}, // Warlord's Dragonhide Headguard
	231688: {}, // General's Dragonhide Grips
	231689: {}, // Field Marshal's Dragonhide Headguard
	231690: {}, // Field Marshal's Dragonhide Chestpiece
	231691: {}, // Marshal's Dragonhide Leggings
	231692: {}, // Marshal's Dragonhide Treads
	231693: {}, // Field Marshal's Dragonhide Shoulders
	231694: {}, // Marshal's Dragonhide Grips
	231695: {}, // Field Marshal's Dragonhide Helm
	231696: {}, // Field Marshal's Dragonhide Armor
	231697: {}, // Marshal's Dragonhide Pants
	231698: {}, // Marshal's Dragonhide Boots
	231699: {}, // Field Marshal's Dragonhide Spaulders
	231700: {}, // Marshal's Dragonhide Gloves
	231701: {}, // Field Marshal's Dragonhide Headdress
	231702: {}, // Field Marshal's Dragonhide Tunic
	231703: {}, // Marshal's Dragonhide Legguards
	231704: {}, // Marshal's Dragonhide Greaves
	231705: {}, // Field Marshal's Dragonhide Pauldrons
	231706: {}, // Marshal's Dragonhide Gauntlets
	231877: {}, // Tyler's Balanced Bobble
	234016: {}, // Signet Ring of the Bronze Dragonflight
	234017: {}, // Signet Ring of the Bronze Dragonflight
	234018: {}, // Signet Ring of the Bronze Dragonflight
	234019: {}, // Signet Ring of the Bronze Dragonflight
	234020: {}, // Signet Ring of the Bronze Dragonflight
	234021: {}, // Signet Ring of the Bronze Dragonflight
	234022: {}, // Signet Ring of the Bronze Dragonflight
	234023: {}, // Signet Ring of the Bronze Dragonflight
	234024: {}, // Signet Ring of the Bronze Dragonflight
	234025: {}, // Signet Ring of the Bronze Dragonflight
	234026: {}, // Signet Ring of the Bronze Dragonflight
	234027: {}, // Signet Ring of the Bronze Dragonflight
	234028: {}, // Signet Ring of the Bronze Dragonflight
	234029: {}, // Signet Ring of the Bronze Dragonflight
	234030: {}, // Signet Ring of the Bronze Dragonflight
	234031: {}, // Signet Ring of the Bronze Dragonflight
	234032: {}, // Signet Ring of the Bronze Dragonflight
	234033: {}, // Signet Ring of the Bronze Dragonflight
	234034: {}, // Signet Ring of the Bronze Dragonflight
	234035: {}, // Signet Ring of the Bronze Dragonflight
	234198: {}, // Signet Ring of the Bronze Dragonflight
	234199: {}, // Signet Ring of the Bronze Dragonflight
	234200: {}, // Signet Ring of the Bronze Dragonflight
	234201: {}, // Signet Ring of the Bronze Dragonflight
	234202: {}, // Signet Ring of the Bronze Dragonflight
	234436: {}, // Signet Ring of the Bronze Dragonflight
	234437: {}, // Signet Ring of the Bronze Dragonflight
	234438: {}, // Signet Ring of the Bronze Dragonflight
	234439: {}, // Signet Ring of the Bronze Dragonflight
	234440: {}, // Signet Ring of the Bronze Dragonflight
	234542: {}, // High Warlord's Greatsword
	234543: {}, // High Warlord's Battle Axe
	234545: {}, // High Warlord's Pulverizer
	234546: {}, // High Warlord's Destroyer
	234547: {}, // High Warlord's Pig Sticker
	234548: {}, // High Warlord's Pig Poker
	234549: {}, // High Warlord's War Staff
	234550: {}, // High Warlord's Spellblade
	234551: {}, // High Warlord's Battle Mace
	234552: {}, // High Warlord's Blade
	234553: {}, // High Warlord's Quickblade
	234554: {}, // High Warlord's Cleaver
	234555: {}, // High Warlord's Bludgeon
	234556: {}, // High Warlord's Razor
	234557: {}, // High Warlord's Right Claw
	234558: {}, // High Warlord's Left Claw
	234559: {}, // High Warlord's Recurve
	234560: {}, // High Warlord's Crossbow
	234561: {}, // High Warlord's Street Sweeper
	234562: {}, // High Warlord's Shield Wall
	234563: {}, // High Warlord's Tome of Destruction
	234564: {}, // High Warlord's Tome of Mending
	234565: {}, // Grand Marshal's Claymore
	234566: {}, // Grand Marshal's Sunderer
	234567: {}, // Grand Marshal's Battle Hammer
	234568: {}, // Grand Marshal's Demolisher
	234569: {}, // Grand Marshal's Glaive
	234570: {}, // Grand Marshal's Polearm
	234571: {}, // Grand Marshal's Stave
	234574: {}, // Grand Marshal's Mageblade
	234576: {}, // Grand Marshal's Warhammer
	234578: {}, // Grand Marshal's Longsword
	234579: {}, // Grand Marshal's Swiftblade
	234580: {}, // Grand Marshal's Handaxe
	234581: {}, // Grand Marshal's Punisher
	234582: {}, // Grand Marshal's Dirk
	234583: {}, // Grand Marshal's Right Hand Blade
	234584: {}, // Grand Marshal's Left Hand Blade
	234585: {}, // Grand Marshal's Bullseye
	234586: {}, // Grand Marshal's Repeater
	234587: {}, // Grand Marshal's Hand Cannon
	234588: {}, // Grand Marshal's Aegis
	234589: {}, // Grand Marshal's Tome of Power
	234590: {}, // Grand Marshal's Tome of Restoration
	234964: {}, // Signet Ring of the Bronze Dragonflight
	234965: {}, // Signet Ring of the Bronze Dragonflight
	234966: {}, // Signet Ring of the Bronze Dragonflight
	234967: {}, // Signet Ring of the Bronze Dragonflight
	234968: {}, // Signet Ring of the Bronze Dragonflight
	235473: {}, // Grand Marshal's Barricade
	235474: {}, // High Warlord's Barricade
	235476: {}, // High Warlord's Hacker
	235477: {}, // High Warlord's Bonecracker
	235478: {}, // High Warlord's Shiv
	235479: {}, // Grand Marshal's Shiv
	235480: {}, // Grand Marshal's Bonecracker
	235481: {}, // Grand Marshal's Hacker
	237814: {}, // Sentinel's Lamellar Legguards
	237815: {}, // Sentinel's Silk Leggings
	237817: {}, // Sentinel's Lizardhide Pants
	237818: {}, // Sentinel's Leather Pants
	237819: {}, // Sentinel's Chain Leggings
	237820: {}, // Outrider's Silk Leggings
	237821: {}, // Outrider's Lizardhide Pants
	237822: {}, // Outrider's Leather Pants
	237823: {}, // Outrider's Mail Leggings
	237824: {}, // Outrider's Chain Leggings
	237825: {}, // Sentinel's Plate Legguards
	237826: {}, // Outrider's Plate Legguards
	239512: {}, // Lightbreaker Wrists
	239513: {}, // Lightbreaker Belt
	239514: {}, // Lightbreaker Grips
	239515: {}, // Lightbreaker Greaves
	239516: {}, // Lightbreaker Shoulders
	239517: {}, // Lightbreaker Greathelm
	239518: {}, // Lightbreaker Tassets
	239519: {}, // Lightbreaker Cuirass
	239520: {}, // Lightbreaker Bracers
	239521: {}, // Lightbreaker Waistguard
	239522: {}, // Lightbreaker Gauntlets
	239523: {}, // Lightbreaker Sabatons
	239524: {}, // Lightbreaker Pauldrons
	239525: {}, // Lightbreaker Helmet
	239526: {}, // Lightbreaker Legplates
	239527: {}, // Lightbreaker Breastplate
	239529: {}, // Dawnstalker Breastplate
	239530: {}, // Dawnstalker Greaves
	239531: {}, // Dawnstalker Gauntlets
	239532: {}, // Dawnstalker Visor
	239533: {}, // Dawnstalker Leggings
	239534: {}, // Dawnstalker Pauldrons
	239535: {}, // Dawnstalker Belt
	239536: {}, // Dawnstalker Vambraces
	239537: {}, // Dawnstalker Boots
	239538: {}, // Dawnstalker Girdle
	239539: {}, // Dawnstalker Handguards
	239540: {}, // Dawnstalker Headpiece
	239541: {}, // Dawnstalker Legguards
	239542: {}, // Dawnstalker Spaulders
	239543: {}, // Dawnstalker Tunic
	239544: {}, // Dawnstalker Wristguards
	239547: {}, // Duskwraith Wristguards
	239548: {}, // Duskwraith Chestguard
	239549: {}, // Duskwraith Grips
	239550: {}, // Duskwraith Mask
	239551: {}, // Duskwraith Leggings
	239552: {}, // Duskwraith Mantle
	239553: {}, // Duskwraith Treads
	239554: {}, // Duskwraith Belt
	239555: {}, // Duskwraith Bracers
	239556: {}, // Duskwraith Waistguard
	239557: {}, // Duskwraith Gauntlets
	239558: {}, // Duskwraith Sabatons
	239559: {}, // Duskwraith Pauldrons
	239560: {}, // Duskwraith Helmet
	239561: {}, // Duskwraith Legplates
	239562: {}, // Duskwraith Breastplate
	239565: {}, // Garb of Revelation
	239572: {}, // Boots of Revelation
	239574: {}, // Hands of Revelation
	239575: {}, // Crown of Revelation
	239577: {}, // Pants of Revelation
	239581: {}, // Mantle of Revelation
	239582: {}, // Girdle of Revelation
	239583: {}, // Wrists of Revelation
	239584: {}, // Gloves of Revelation
	239585: {}, // Circlet of Revelation
	239586: {}, // Shoulderpads of Revelation
	239587: {}, // Leggings of Revelation
	239588: {}, // Bindings of Revelation
	239589: {}, // Sandals of Revelation
	239590: {}, // Belt of Revelation
	239591: {}, // Robe of Revelation
	240020: {}, // Inquisition Legguards
	240021: {}, // Inquisition Spaulders
	240022: {}, // Inquisition Boots
	240023: {}, // Inquisition Vambraces
	240024: {}, // Inquisition Belt
	240025: {}, // Inquisition Shoulderplates
	240026: {}, // Inquisition Leggings
	240027: {}, // Inquisition Helmet
	240028: {}, // Inquisition Gloves
	240029: {}, // Inquisition Greaves
	240030: {}, // Inquisition Breastplate
	240031: {}, // Inquisition Bracers
	240032: {}, // Inquisition Waistguard
	240033: {}, // Inquisition Pauldrons
	240034: {}, // Inquisition Legplates
	240035: {}, // Inquisition Crown
	240036: {}, // Inquisition Gauntlets
	240037: {}, // Inquisition Sabatons
	240038: {}, // Inquisition Chestguard
	240039: {}, // Inquisition Tunic
	240040: {}, // Inquisition Headpiece
	240041: {}, // Inquisition Handguards
	240042: {}, // Inquisition Wristguards
	240043: {}, // Inquisition Girdle
	240044: {}, // Fireleaf Wristwraps
	240045: {}, // Fireleaf Waistguard
	240046: {}, // Fireleaf Mantle
	240047: {}, // Fireleaf Pants
	240048: {}, // Fireleaf Hood
	240049: {}, // Fireleaf Mitts
	240050: {}, // Fireleaf Boots
	240051: {}, // Fireleaf Garb
	240052: {}, // Fireleaf Bindings
	240053: {}, // Fireleaf Belt
	240054: {}, // Fireleaf Shoulderpads
	240055: {}, // Fireleaf Leggings
	240056: {}, // Fireleaf Circlet
	240057: {}, // Fireleaf Gloves
	240058: {}, // Fireleaf Sandals
	240059: {}, // Fireleaf Robe
	240060: {}, // Waywatcher Wraps
	240061: {}, // Waywatcher Sash
	240062: {}, // Waywatcher Shoulderpads
	240063: {}, // Waywatcher Trousers
	240064: {}, // Waywatcher Cowl
	240065: {}, // Waywatcher Grips
	240066: {}, // Waywatcher Stompers
	240067: {}, // Waywatcher Vest
	240068: {}, // Waywatcher Bindings
	240069: {}, // Waywatcher Cord
	240070: {}, // Waywatcher Mantle
	240071: {}, // Waywatcher Kilt
	240072: {}, // Waywatcher Hood
	240073: {}, // Waywatcher Mitts
	240074: {}, // Waywatcher Sandals
	240075: {}, // Waywatcher Leathers
	240076: {}, // Waywatcher Bracers
	240077: {}, // Waywatcher Waistguard
	240078: {}, // Waywatcher Pauldrons
	240079: {}, // Waywatcher Leggings
	240080: {}, // Waywatcher Headdress
	240081: {}, // Waywatcher Gauntlets
	240082: {}, // Waywatcher Sabatons
	240083: {}, // Waywatcher Armor
	240084: {}, // Waywatcher Wristguards
	240085: {}, // Waywatcher Girdle
	240086: {}, // Waywatcher Spaulders
	240087: {}, // Waywatcher Legguards
	240088: {}, // Waywatcher Headpiece
	240089: {}, // Waywatcher Handguards
	240090: {}, // Waywatcher Boots
	240091: {}, // Waywatcher Tunic
	240092: {}, // Soulcrusher Tunic
	240093: {}, // Soulcrusher Boots
	240095: {}, // Soulcrusher Handguards
	240096: {}, // Soulcrusher Headpiece
	240097: {}, // Soulcrusher Legguards
	240098: {}, // Soulcrusher Spaulders
	240099: {}, // Soulcrusher Girdle
	240100: {}, // Soulcrusher Wristguards
	240101: {}, // Soulcrusher Chestguard
	240102: {}, // Soulcrusher Sabatons
	240103: {}, // Soulcrusher Gloves
	240104: {}, // Soulcrusher Faceguard
	240105: {}, // Soulcrusher Legplates
	240106: {}, // Soulcrusher Pauldrons
	240107: {}, // Soulcrusher Waistguard
	240108: {}, // Soulcrusher Bracers
	240109: {}, // Soulcrusher Embrace
	240110: {}, // Soulcrusher Greaves
	240122: {}, // Soulcrusher Mitts
	240123: {}, // Soulcrusher Crown
	240124: {}, // Soulcrusher Kilt
	240125: {}, // Soulcrusher Mantle
	240126: {}, // Soulcrusher Cord
	240127: {}, // Soulcrusher Bindings
	240128: {}, // Soulcrusher Armor
	240129: {}, // Soulcrusher Treads
	240130: {}, // Soulcrusher Grips
	240131: {}, // Soulcrusher Helmet
	240134: {}, // Soulcrusher Leggings
	240135: {}, // Soulcrusher Epaulets
	240136: {}, // Soulcrusher Belt
	240137: {}, // Soulcrusher Vambraces
	240138: {}, // Heretic Robe
	240139: {}, // Heretic Sandals
	240140: {}, // Heretic Gloves
	240141: {}, // Heretic Cowl
	240142: {}, // Heretic Leggings
	240143: {}, // Heretic Shoulderpads
	240144: {}, // Heretic Belt
	240145: {}, // Heretic Bindings
	240146: {}, // Heretic Garb
	240147: {}, // Heretic Handguards
	240148: {}, // Heretic Hood
	240149: {}, // Heretic Pants
	240150: {}, // Heretic Mantle
	240151: {}, // Heretic Waistguard
	240152: {}, // Heretic Wristguards
	240153: {}, // Heretic Boots
	241089: {}, // Scarlet Dagger
	241090: {}, // Scarlet Great Axe
	// Forever test items the game files carry
	280304: {}, // Invisible Dagger
}

// Item Ids of consumables to allow
var ConsumableAllowList = []int32{
	7676,  // Thisle Tea
	9088,  // Gift of Arthas
	9155,  // Arcane Elixir
	9421,  // Major Healthstone
	13442, // Migty Rage Potion
	13452, // Elixir of the Mongoose
	13454, // Greater Arcane Elixir
	12662, // Demonic Rune
	5206,  // Bogling Root
	// The Natural flasks, whose zone bonus sim/common/forever registers by hand
	274273, // Flask of Natural Accuracy
	274274, // Flask of Natural Aggression
	274275, // Flask of Natural Precision
	274276, // Flask of Natural Swiftness
}
var ConsumableDenyList = []int32{
	// Not obtainable in the game
	5265,  // Watered-down Beer
	23578, // Diet McWeaksauce
	23579, // The McWeaksauce Classic
	23794, // Permanent Sheen of Zanza
	23795, // Permanent Spirit of Zanza
	23796, // Permanent Swiftness of Zanza
	// Mage familiar summons, which the settings leave out
	275069, // Scroll of Rat Familiar
	277483, // Scroll of Frog Familiar
	277493, // Scroll of Cat Familiar
	// Season of Discovery consumables the game files still carry
	206177, // Pine Nut Butter
	210708, // Elixir of Coalesced Regret
	211815, // Silverwing Battle Hymn
	211816, // Warsong Battle Drum
	211953, // Scroll of Arcane Recovery I
	211954, // Scroll of Arcane Accuracy I
	211955, // Scroll of Arcane Protection - Fire I
	211956, // Scroll of Arcane Protection - Frost I
	211957, // Scroll of Arcane Power I
	213564, // Scroll of Minor Evocation
	215127, // High-Yield Radiation Bomb
	217308, // Enchanted Sigil: Innovation
	217398, // Lesser Arcane Elixir
	217495, // Scroll of Cryoblast
	223913, // Major Healing Potion
	223914, // Greater Healing Potion
	227286, // Warmed Rumsey Rum with Cream
	228978, // Enchanted Sigil: Flowing Waters
	228980, // Conductive Shield Coating
	232436, // Darkclaw Bisque
	232438, // Smoked Redgill
	232611, // Magnificent Trollshine
	238637, // Prowler Steak
	238638, // Filet o' Flank
	238639, // Sunrise Omelette
}

// Exclusive groups the server keeps that the rows do not state: the game files carry no stacking
// table, and these items share no aura with their group and no category whose cooldown outlasts
// their buffs. A consumable named here takes this group instead of the one its rows give it.
var ConsumableExclusiveGroups = []ConsumableGroup{
	{Group: ElixirGroup(dbcenums.A_MOD_STAT, 0), Items: []int32{
		12451, // Juju Power
	}},
	{Group: ElixirGroup(dbcenums.A_MOD_ATTACK_POWER, 0), Items: []int32{
		12460, // Juju Might
	}},
	{Group: ConsumableGroupKey{Key: "Zanza", Name: "Zanza"}, Items: []int32{
		8410,  // R.O.I.D.S.
		8411,  // Lung Juice Cocktail
		8412,  // Ground Scorpok Assay
		8423,  // Cerebral Cortex Compound
		8424,  // Gizzard Gum
		20079, // Spirit of Zanza
	}},
	{Group: ConsumableGroupKey{Key: "Hit food", Name: "Hit food"}, Items: []int32{
		22237, // Dark Desire
		23327, // Fire-toasted Bun
	}},
}

// The world events by the ItemSparse.RequiredHoliday ids the items state.
var HolidayNames = map[int]string{
	324: "Hallow's End",
	327: "Lunar Festival",
	335: "Love is in the Air",
}

// Holiday consumables whose item states no RequiredHoliday: the Midsummer Fire Festival food only
// expires 14 days after it is bought.
var HolidayConsumables = map[int32]string{
	23211: "Midsummer Fire Festival", // Toasted Smorc
	23326: "Midsummer Fire Festival", // Midsummer Sausage
	23327: "Midsummer Fire Festival", // Fire-toasted Bun
}

// Raid buffs / debuffs
var SharedSpellsIcons = []int32{
	// Registered CD's
	10060,
	16190,
	29166,
	2825,

	17051,

	25898,

	20140,
	8071,

	14767,

	8075,

	20045,

	30808,
	19506,

	12861,
	18696,

	20245,
	5675,
	16206,

	17007,

	8512,
	29193,

	24907,

	3738,
	8227,

	16840,

	// Raid Debuffs
	8647,

	770,
	702,
	18180,

	12879,
	16862,

	20337,

	3043,

	1490,

	20271,

	11374,
	15235,

	// Raid buffs, debuffs and shared consumables.
	603,   // Curse of Doom
	688,   // Summon Imp
	691,   // Summon Felhunter
	697,   // Summon Voidwalker
	704,   // Curse of Recklessness
	706,   // Demon Armor
	712,   // Summon Succubus
	6117,  // Mage Armor
	7302,  // Ice Armor
	8024,  // Flametongue Weapon
	8033,  // Frostbrand Weapon
	8232,  // Windfury Weapon
	10538, // Fire Resistance Totem
	12470, // Fire Nova
	13339, // Fire Blast
	13376, // Fire Shield
	13889, // Minor Speed
	14325, // Hunter's Mark
	17768, // Wolfshead Helm
	18803, // Focus
	19615, // Frenzy Effect
	20554, // Berserking
	20572, // Blood Fury
	20574, // Savage
	20594, // Stoneform
	20597, // Ruthless
	23110, // Dash
	23563, // Enhanced Battle Shout
	25076, // Cobra Reflexes
	21564, // Prayer of Fortitude
	25895, // Mass Salvation
	25916, // Greater Blessing of Might
	25918, // Greater Blessing of Wisdom
	27681, // Prayer of Spirit
	27683, // Prayer of Shadow Protection
	26654, // Sweeping Strikes
	28142, // Power of the Guardian
	28143, // Power of the Guardian
	29414, // Haste

	1259686, // Skysight
	1259688, // Elemental Blessing
	1259705, // Read Ley Line
	1259719, // Earth Fury
	1259799, // Elune's Light
	1259812, // Eureka!
	1259813, // Eureka!
	1259817, // Eureka!
	1259821, // Eureka!
	1259823, // Eureka!
	1260198, // Touch of the Grave
	1270842, // Energized
	1299026, // Shatter Curse
}

// If any of these match the item name, don't include it.
var DenyListNameRegexes = []*regexp.Regexp{
	regexp.MustCompile(`30 Epic`),
	regexp.MustCompile(`130 Epic`),
	regexp.MustCompile(`63 Blue`),
	regexp.MustCompile(`63 Green`),
	regexp.MustCompile(`66 Epic`),
	regexp.MustCompile(`90 Epic`),
	regexp.MustCompile(`90 Green`),
	regexp.MustCompile(`Boots 1`),
	regexp.MustCompile(`Boots 2`),
	regexp.MustCompile(`Boots 3`),
	regexp.MustCompile(`Bracer 1`),
	regexp.MustCompile(`Bracer 2`),
	regexp.MustCompile(`Bracer 3`),
	regexp.MustCompile(`DB\d`),
	regexp.MustCompile(`DEPRECATED`),
	regexp.MustCompile(`OLD`),
	regexp.MustCompile(`Deprecated`),
	regexp.MustCompile(`Deprecated: Keanna`),
	regexp.MustCompile(`Indalamar`),
	regexp.MustCompile(`Monster -`),
	regexp.MustCompile(`NEW`),
	regexp.MustCompile(`PH`),
	regexp.MustCompile(`QR XXXX`),
	regexp.MustCompile(`TEST`),
	regexp.MustCompile(`Test`),
	regexp.MustCompile(`Enchant Template`),
	regexp.MustCompile(`Deleted`),
	regexp.MustCompile(`DELETED`),
	regexp.MustCompile(`zOLD`),
	regexp.MustCompile(`Archaic Spell`),
	regexp.MustCompile(`Well Repaired`),
	regexp.MustCompile(`Boss X`),
	regexp.MustCompile(`Adventurine`),
	regexp.MustCompile(`Sardonyx`),
	regexp.MustCompile(`Zyanite`),
	regexp.MustCompile(`zzold`),
	regexp.MustCompile(`Tom's`),
	regexp.MustCompile(`Stabilized Eternium Scope`),
	regexp.MustCompile(`\[TMP\]`),
	regexp.MustCompile(`\[DNT\]`),
	regexp.MustCompile(`Placeholder`),
	regexp.MustCompile(` DEP$`),
	regexp.MustCompile(` UNUSED$`),
}

var EnchantDenyListSpells = map[int32]struct{}{}
var EnchantDenyListItems = map[int32]struct{}{}

var EnchantDenyList = map[int32]struct{}{}

var EnchantAllowList = []int32{
	804,  // Enchant Cloak - Lesser Shadow Resistance
	684,  // Enchant Gloves - Major Strength
	963,  // Enchant Weapon - Major Striking
	1900, // Enchant Weapon - Crusader
	2564, // Weapon 15 Agi
	2583, // Presence of Might
	2588, // Presence of Sight
}

// Note: EffectId is required for all enchants, because they are
// used by various importers/exporters
// Note: ItemId, SpellId and Name are part of the enchant DB key, so they must
// match the generated enchant or the override is added as a separate entry.
var EnchantOverrides = []*proto.UIEnchant{
	// The head/leg resistance kits grant their stats through an equip spell that no
	// longer exists in the game files, so they have to be filled in by hand.
	{EffectId: 2681, ItemId: 22635, SpellId: 28162, Name: "Savage Guard", Stats: stats.Stats{stats.NatureResistance: 10}.ToProtoArray()},
	{EffectId: 2682, ItemId: 22636, SpellId: 28164, Name: "Ice Guard", Stats: stats.Stats{stats.FrostResistance: 10}.ToProtoArray()},
	{EffectId: 2683, ItemId: 22638, SpellId: 28166, Name: "Shadow Guard", Stats: stats.Stats{stats.ShadowResistance: 10}.ToProtoArray()},
}
