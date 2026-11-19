package overrides

// A chance-on-hit item's procs per minute, which the game files keep outside the spell data: the
// spell's ProcChance column reads the 100 or 101 sentinel, or the spell has no proc row at all. The
// rate is the weapon's rather than the spell's, since two weapons can share a trigger spell at
// different rates, so it goes onto the item effect. Reason and Source weigh the same as an
// Override's; the database tests refuse a row without them, or one naming no effect of its item.
type WeaponProc struct {
	ItemID  int32
	SpellID int32
	PPM     float64
	Reason  string
	Source  string // "EV-<n>" (docs/evidence), "tbc-carryover", "vanilla-carryover" or "observed"
}

// Each rate is a carry-over or an observed fit unless its Source is the docs/evidence id of a measurement on this server. A hotfix that
// adds back a weapon whose row was dropped needs the row restored from git history, or the database tests that name
// the item fail the Update Items run.
var WeaponProcs = []WeaponProc{
	// Measured in end-game dungeon logs (EV-0003).
	{
		ItemID:  10750,
		SpellID: 1300356,
		PPM:     1.5,
		Reason:  "Lifeforce Dirk: no proc row",
		Source:  "EV-0003",
	},
	{
		ItemID:  11744,
		SpellID: 16433,
		PPM:     3.2,
		Reason:  "Bloodfist: no proc row",
		Source:  "EV-0003",
	},
	{
		ItemID:  11803,
		SpellID: 18086,
		PPM:     2,
		Reason:  "Force of Magma: no proc row",
		Source:  "EV-0003",
	},
	{
		ItemID:  12463,
		SpellID: 14118,
		PPM:     1.5,
		Reason:  "Drakefang Butcher: column reads 101",
		Source:  "EV-0003",
	},
	{
		ItemID:  12582,
		SpellID: 16528,
		PPM:     1.7,
		Reason:  "Keris of Zul'Serak: no proc row",
		Source:  "EV-0003",
	},
	{
		ItemID:  12969,
		SpellID: 17196,
		PPM:     0.9,
		Reason:  "Seeping Willow: column reads 101",
		Source:  "EV-0003",
	},
	{
		ItemID:  13183,
		SpellID: 18203,
		PPM:     1.7,
		Reason:  "Venomspitter: column reads 101",
		Source:  "EV-0003",
	},
	{
		ItemID:  13348,
		SpellID: 17483,
		PPM:     1.1,
		Reason:  "Demonshear: column reads 101",
		Source:  "EV-0003",
	},
	{
		ItemID:  13983,
		SpellID: 18289,
		PPM:     1.9,
		Reason:  "Gravestone War Axe: no proc row",
		Source:  "EV-0003",
	},
	{
		ItemID:  17730,
		SpellID: 21949,
		PPM:     1.2,
		Reason:  "Gatorbite Axe: column reads 101",
		Source:  "EV-0003",
	},
	{
		ItemID:  17753,
		SpellID: 29655,
		PPM:     3.5,
		Reason:  "Verdant Keeper's Aim: no proc row; melee hits do not roll it",
		Source:  "EV-0003",
	},
	{
		ItemID:  17766,
		SpellID: 21961,
		PPM:     1.8,
		Reason:  "Princess Theradras' Scepter: column reads 101",
		Source:  "EV-0003",
	},
	{
		ItemID:  17943,
		SpellID: 1294794,
		PPM:     3.2,
		Reason:  "Fist of Stone: equip trigger, column reads 100",
		Source:  "EV-0003",
	},
	{
		ItemID:  18388,
		SpellID: 1302279,
		PPM:     1.7,
		Reason:  "Stoneshatter: no proc row",
		Source:  "EV-0003",
	},
	{
		ItemID:  274270,
		SpellID: 1293568,
		PPM:     1.1,
		Reason:  "Weed Whacker: no proc row",
		Source:  "EV-0003",
	},
	{
		ItemID:  275441,
		SpellID: 1297921,
		PPM:     2.3,
		Reason:  "Cremation Longbow: no proc row",
		Source:  "EV-0003",
	},
	{
		ItemID:  276419,
		SpellID: 1299631,
		PPM:     3.6,
		Reason:  "Lightcrest Hammer: no proc row",
		Source:  "EV-0003",
	},
	// Measured the same way in combat logs of levelling players (EV-0022). A druid's Maul, Swipe, Claw and
	// Lacerate hits are left out as well, since they almost never give a chance-on-hit weapon's proc. The chance
	// a swing scales with the swing speed, as procs per minute do.
	{
		ItemID:  9478,
		SpellID: 1292573,
		PPM:     2,
		Reason:  "Ripsaw: no proc row",
		Source:  "EV-0022",
	},
	{
		ItemID:  9639,
		SpellID: 13532,
		PPM:     2,
		Reason:  "The Hand of Antu'sul: no proc row",
		Source:  "EV-0022",
	},
	{
		ItemID:  11086,
		SpellID: 11657,
		PPM:     2.9,
		Reason:  "Jang'thraze the Protector: no proc row",
		Source:  "EV-0022",
	},
	{
		ItemID:  17705,
		SpellID: 21919,
		PPM:     1,
		Reason:  "Thrash Blade: no proc row",
		Source:  "EV-0022",
	},
	{
		ItemID:  279852,
		SpellID: 1309967,
		PPM:     1.85,
		Reason:  "Rock on Stick: no proc row; measured on main-hand hits without extra attacks or haste",
		Source:  "EV-0050",
	},
	// Measured the same way in reports of dungeons below the end game, on players with no extra attacks, leaving
	// out the hits made under an attack speed aura and every swing that came sooner than the weapon's speed
	// (EV-0024).
	{
		ItemID:  287005,
		SpellID: 1324687,
		PPM:     0.6,
		Reason:  "Sword of Dismemberment: no proc row",
		Source:  "EV-0024",
	},
	// Measured on isolated hits in end-game and levelling dungeon reports: landed hits with no other hit of the
	// wearer close enough to have rolled the proc, from wearers with no off-hand weapon (EV-0119).
	{
		ItemID:  6220,
		SpellID: 1292152,
		PPM:     1.55,
		Reason:  "Meteor Shard: no proc row",
		Source:  "EV-0119",
	},
	{
		ItemID:  6472,
		SpellID: 1291663,
		PPM:     1.75,
		Reason:  "Stinging Viper: column reads 101",
		Source:  "EV-0119",
	},
	{
		ItemID:  6904,
		SpellID: 8313,
		PPM:     1.45,
		Reason:  "Bite of Serra'kis: column reads 101",
		Source:  "EV-0119",
	},
	{
		ItemID:  7717,
		SpellID: 9632,
		PPM:     1,
		Reason:  "Ravager: no proc row",
		Source:  "EV-0119",
	},
	{
		ItemID:  9418,
		SpellID: 12731,
		PPM:     1.95,
		Reason:  "Stoneslayer: no proc row",
		Source:  "EV-0119",
	},
	{
		ItemID:  9453,
		SpellID: 11790,
		PPM:     1.8,
		Reason:  "Toxic Revenger: column reads 101",
		Source:  "EV-0119",
	},
	{
		ItemID:  12974,
		SpellID: 14106,
		PPM:     0.6,
		Reason:  "The Black Knight: no proc row",
		Source:  "EV-0119",
	},
	{
		ItemID:  13040,
		SpellID: 29626,
		PPM:     2.7,
		Reason:  "Heartseeking Crossbow: equip trigger, column reads 100",
		Source:  "EV-0119",
	},
	{
		ItemID:  274425,
		SpellID: 1294055,
		PPM:     3.75,
		Reason:  "White Obsidian Wand: equip trigger, column reads 100; wand shots roll it",
		Source:  "EV-0119",
	},
	{
		ItemID:  279480,
		SpellID: 1306816,
		PPM:     1.5,
		Reason:  "Bigger Rock on Bigger Stick: no proc row",
		Source:  "EV-0119",
	},
	// Measured on off-hand wearers' white swings, each proc given to the hit logged right after it (EV-0120).
	{
		ItemID:  871,
		SpellID: 18797,
		PPM:     1.65,
		Reason:  "Flurry Axe: column reads 101",
		Source:  "EV-0120",
	},
	// Measured on main-hand white swings and melee casts of wearers holding it in the main hand (EV-0121).
	{
		ItemID:  13361,
		SpellID: 17484,
		PPM:     1.95,
		Reason:  "Skullforge Reaver: column reads 100",
		Source:  "EV-0121",
	},
	// Measured on a cat druid's white swings, priced on the 1.0 s paw (EV-0122).
	{
		ItemID:  274963,
		SpellID: 1295744,
		PPM:     1.45,
		Reason:  "Rot-Covered Harpoon: column reads 101",
		Source:  "EV-0122",
	},
	// The rate is fitted from logs (EV-0125, EV-0136), too few to call final. The "observed" Source keeps the
	// gear tab's notice up until more wearers validate it.
	{
		ItemID:  12592,
		SpellID: 16602,
		PPM:     1.15,
		Reason:  "Blackblade of Shahram: no proc row; a log fit not yet validated (EV-0125, EV-0136)",
		Source:  "observed",
	},
	// Fitted from logs on the landed white swings of the mace's wearers and rounded to 1.8 by the owner; their
	// Stormstrikes roll it at the same share (EV-0150). The "observed" Source keeps the gear tab's notice up until
	// more wearers validate it.
	{
		ItemID:  13393,
		SpellID: 17500,
		PPM:     1.8,
		Reason:  "Malown's Slam: no proc row; a log fit not yet validated (EV-0150)",
		Source:  "observed",
	},
	// Fitted from logs on a sword-and-board warrior's weapon hits and rounded to 1 by the owner (EV-0126).
	// The "observed" Source keeps the gear tab's notice up until more wearers validate it.
	{
		ItemID:  13246,
		SpellID: 17352,
		PPM:     1,
		Reason:  "Argent Avenger: no proc row; a log fit not yet validated (EV-0126)",
		Source:  "observed",
	},
	// Fitted from one wearer's landed melee hits; the owner rounds it to 1 (EV-0132). The "observed"
	// Source keeps the gear tab's notice up until more wearers validate it.
	{
		ItemID:  1263,
		SpellID: 17148,
		PPM:     1,
		Reason:  "Brain Hacker: column reads 101; a log fit not yet validated (EV-0132)",
		Source:  "observed",
	},
	// Measured on the landed hits each trigger hears, each priced on its weapon's base speed (EV-0136).
	{
		ItemID:  754,
		SpellID: 13519,
		PPM:     1,
		Reason:  "Shortsword of Vengeance: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  2099,
		SpellID: 29639,
		PPM:     4.3,
		Reason:  "Dwarven Hand Cannon: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  3078,
		SpellID: 1292647,
		PPM:     2.2,
		Reason:  "Naga Heartpiercer: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  2164,
		SpellID: 18107,
		PPM:     1.55,
		Reason:  "Gut Ripper: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  6469,
		SpellID: 29653,
		PPM:     2.05,
		Reason:  "Venomstrike: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  6831,
		SpellID: 1293407,
		PPM:     2,
		Reason:  "Black Menace: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  6953,
		SpellID: 1322218,
		PPM:     2.2,
		Reason:  "Verigan's Fist: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  9412,
		SpellID: 29644,
		PPM:     4.25,
		Reason:  "Galgann's Fireblaster: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  9467,
		SpellID: 3742,
		PPM:     2.1,
		Reason:  "Gahz'rilla Fang: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  9477,
		SpellID: 1294339,
		PPM:     1,
		Reason:  "The Chief's Enforcer: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  10567,
		SpellID: 29646,
		PPM:     2.9,
		Reason:  "Quillshooter: no proc row",
		Source:  "EV-0136",
	},
	{
		ItemID:  13952,
		SpellID: 1298416,
		PPM:     1.9,
		Reason:  "Iceblade Hacker: column reads 101",
		Source:  "EV-0136",
	},
	{
		ItemID:  267369,
		SpellID: 1282503,
		PPM:     2.25,
		Reason:  "Wolfsbane: no proc row",
		Source:  "EV-0136",
	},
	// Measured the same way with the reports up to 2026-11-17 added (EV-0146).
	{
		ItemID:  18324,
		SpellID: 1302166,
		PPM:     1.95,
		Reason:  "Waveslicer: no proc row",
		Source:  "EV-0146",
	},
	// Fitted the same way from the reports up to 2026-11-17, from too few wearers, too wide an interval
	// or wearers of one class only to call final (EV-0146). The "observed" Source keeps the gear tab's notice up until more wearers
	// validate it.
	{
		ItemID:  811,
		SpellID: 18104,
		PPM:     1.7,
		Reason:  "Axe of the Deep Woods: no proc row; a log fit not yet validated (EV-0146)",
		Source:  "observed",
	},
	{
		ItemID:  6660,
		SpellID: 8348,
		PPM:     1.05,
		Reason:  "Julie's Dagger: column reads 100; a log fit not yet validated (EV-0146)",
		Source:  "observed",
	},
	{
		ItemID:  12791,
		SpellID: 1300787,
		PPM:     2.05,
		Reason:  "Barman Shanker: column reads 101; a log fit not yet validated (EV-0146)",
		Source:  "observed",
	},
	{
		ItemID:  275439,
		SpellID: 1296730,
		PPM:     2.1,
		Reason:  "Blunderflame Bone Bow: no proc row; a log fit not yet validated (EV-0146)",
		Source:  "observed",
	},
	// Measured on every hit the class census finds rolling the weapon, each priced on its weapon's base speed (EV-0153,
	// EV-0155).
	{
		ItemID:  869,
		SpellID: 13752,
		PPM:     2.65,
		Reason:  "Dazzling Longsword: column reads 101",
		Source:  "EV-0153",
	},
	{
		ItemID:  870,
		SpellID: 18796,
		PPM:     0.85,
		Reason:  "Fiery War Axe: column reads 101",
		Source:  "EV-0153",
	},
	{
		ItemID:  1481,
		SpellID: 1292158,
		PPM:     2.05,
		Reason:  "Grimclaw: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  1727,
		SpellID: 13528,
		PPM:     1,
		Reason:  "Sword of Decay: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  2299,
		SpellID: 18199,
		PPM:     0.95,
		Reason:  "Burning War Axe: column reads 101",
		Source:  "EV-0153",
	},
	{
		ItemID:  2825,
		SpellID: 29638,
		PPM:     4.75,
		Reason:  "Bow of Searing Arrows: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  2942,
		SpellID: 13491,
		PPM:     2,
		Reason:  "Iron Knuckles: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  5752,
		SpellID: 16400,
		PPM:     2,
		Reason:  "Wyvern Tailspike: column reads 101; shares the trigger with Naraxis' Fang",
		Source:  "EV-0153",
	},
	{
		ItemID:  6681,
		SpellID: 1293183,
		PPM:     3.15,
		Reason:  "Thornspike: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  6909,
		SpellID: 1292657,
		PPM:     2.05,
		Reason:  "Strike of the Hydra: column reads 101",
		Source:  "EV-0153",
	},
	{
		ItemID:  7710,
		SpellID: 1293482,
		PPM:     1.9,
		Reason:  "Loksey's Training Stick: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  7961,
		SpellID: 9806,
		PPM:     2,
		Reason:  "Phantom Blade: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  8190,
		SpellID: 16405,
		PPM:     1,
		Reason:  "Hanzo Sword: column reads 101; shares the trigger with The Needler",
		Source:  "EV-0153",
	},
	{
		ItemID:  8225,
		SpellID: 13530,
		PPM:     1.95,
		Reason:  "Tainted Pierce: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  9379,
		SpellID: 1294243,
		PPM:     2.05,
		Reason:  "Sang'thraze the Deflector: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  9425,
		SpellID: 10373,
		PPM:     0.9,
		Reason:  "Pendulum of Doom: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  9446,
		SpellID: 1292851,
		PPM:     3.9,
		Reason:  "Electrocutioner Leg: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  9475,
		SpellID: 18206,
		PPM:     1.75,
		Reason:  "Diabolic Skiver: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  10761,
		SpellID: 1293790,
		PPM:     2.95,
		Reason:  "Coldrage Dagger: column reads 101",
		Source:  "EV-0153",
	},
	{
		ItemID:  10772,
		SpellID: 18075,
		PPM:     2,
		Reason:  "Glutton's Cleaver: column reads 101",
		Source:  "EV-0153",
	},
	{
		ItemID:  10804,
		SpellID: 18084,
		PPM:     3.1,
		Reason:  "Fist of the Damned: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  11121,
		SpellID: 16408,
		PPM:     2.9,
		Reason:  "Darkwater Talwar: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  11817,
		SpellID: 15602,
		PPM:     1.95,
		Reason:  "Lord General's Sword: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  13218,
		SpellID: 17331,
		PPM:     0.95,
		Reason:  "Fang of the Crystal Spider: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  14145,
		SpellID: 18381,
		PPM:     2,
		Reason:  "Cursed Felblade: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  17752,
		SpellID: 1292155,
		PPM:     1.95,
		Reason:  "Satyr's Lash: no proc row; shares the trigger with Black Malice",
		Source:  "EV-0153",
	},
	{
		ItemID:  272999,
		SpellID: 1291551,
		PPM:     2.9,
		Reason:  "Barbaric Crossbow: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  274293,
		SpellID: 1293714,
		PPM:     0.95,
		Reason:  "Spellsever Crossbow: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  279876,
		SpellID: 1309315,
		PPM:     1.95,
		Reason:  "Plaguefang: column reads 101",
		Source:  "EV-0153",
	},
	{
		ItemID:  279895,
		SpellID: 1322295,
		PPM:     2.8,
		Reason:  "Ironforge Greathammer: no proc row",
		Source:  "EV-0153",
	},
	{
		ItemID:  286430,
		SpellID: 1320269,
		PPM:     2.05,
		Reason:  "Toofbreaker: no proc row",
		Source:  "EV-0153",
	},
	// Fitted the same way from too few wearers, too wide an interval or wearers of one class only to call
	// final (EV-0153). The "observed" Source keeps the gear tab's notice up until more wearers validate it.
	{
		ItemID:  647,
		SpellID: 17152,
		PPM:     1.05,
		Reason:  "Destiny: no proc row; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  1982,
		SpellID: 18211,
		PPM:     2.1,
		Reason:  "Nightblade: no proc row; shares the trigger with Doombringer and Ebon Hand; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  2000,
		SpellID: 18091,
		PPM:     0.9,
		Reason:  "Archeus: column reads 101; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  2915,
		SpellID: 16415,
		PPM:     0.95,
		Reason:  "Taran Icebreaker: column reads 101; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  6738,
		SpellID: 16403,
		PPM:     1,
		Reason:  "Bleeding Crescent: column reads 101; shares the trigger with Silithid Ripper; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  8223,
		SpellID: 10351,
		PPM:     1.1,
		Reason:  "Blade of the Basilisk: no proc row; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  8708,
		SpellID: 1317432,
		PPM:     1.85,
		Reason:  "Hammer of Expertise: no proc row; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  9372,
		SpellID: 1294430,
		PPM:     1.4,
		Reason:  "Sul'thraze the Lasher: no proc row; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  11920,
		SpellID: 16414,
		PPM:     3.35,
		Reason:  "Wraith Scythe: no proc row; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  12590,
		SpellID: 16551,
		PPM:     0.55,
		Reason:  "Felstriker: no proc row; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  13953,
		SpellID: 1293654,
		PPM:     1.45,
		Reason:  "Silent Fang: no proc row; shares the trigger with Fight Club; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	// The one-hand white swings disagree with the rate over all hits on these weapons, so their rate is open and each
	// row keeps its earlier rate (EV-0153). The "observed" Source keeps the gear tab's notice up until the rate is
	// settled.
	{
		ItemID:  809,
		SpellID: 17504,
		PPM:     0.9,
		Reason:  "Bloodrazor: column reads 101; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  11816,
		SpellID: 1300783,
		PPM:     2,
		Reason:  "Angerforge's Battle Axe: no proc row; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	{
		ItemID:  13016,
		SpellID: 13318,
		PPM:     0.9,
		Reason:  "Killmaim: column reads 101; a log fit not yet validated (EV-0153)",
		Source:  "observed",
	},
	// Fitted the same way from too few wearers or too wide an interval to call final (EV-0136). The
	// "observed" Source keeps the gear tab's notice up until more wearers validate it.
	{
		ItemID:  810,
		SpellID: 13439,
		PPM:     4.55,
		Reason:  "Hammer of the Northern Wind: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  934,
		SpellID: 13524,
		PPM:     1.4,
		Reason:  "Stalvan's Reaper: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  937,
		SpellID: 18138,
		PPM:     0.9,
		Reason:  "Black Duskwood Staff: no proc row; shares the trigger with Shadowblade; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  1265,
		SpellID: 18208,
		PPM:     0.95,
		Reason:  "Scorpion Sting: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  1318,
		SpellID: 13480,
		PPM:     1.9,
		Reason:  "Night Reaver: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  1726,
		SpellID: 16401,
		PPM:     0.85,
		Reason:  "Poison-tipped Bone Spear: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  2163,
		SpellID: 18138,
		PPM:     1.05,
		Reason:  "Shadowblade: no proc row; shares the trigger with Black Duskwood Staff; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  2256,
		SpellID: 13440,
		PPM:     1.05,
		Reason:  "Skeletal Club: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  2291,
		SpellID: 17153,
		PPM:     1,
		Reason:  "Kang the Decapitator: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  2912,
		SpellID: 16409,
		PPM:     1,
		Reason:  "Claw of the Shadowmancer: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  3194,
		SpellID: 1292155,
		PPM:     2.05,
		Reason:  "Black Malice: no proc row; shares the trigger with Satyr's Lash; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  5197,
		SpellID: 1292008,
		PPM:     1.9,
		Reason:  "Cookie's Tenderizer: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  5243,
		SpellID: 1291686,
		PPM:     3.85,
		Reason:  "Firebelcher: equip trigger, column reads 100; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  5756,
		SpellID: 1293653,
		PPM:     2.25,
		Reason:  "Sliverblade: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  7730,
		SpellID: 18204,
		PPM:     2.25,
		Reason:  "Cobalt Crusher: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  7753,
		SpellID: 18200,
		PPM:     0.95,
		Reason:  "Bloodspiller: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  8006,
		SpellID: 13482,
		PPM:     3.6,
		Reason:  "The Ziggler: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  9485,
		SpellID: 1293005,
		PPM:     1,
		Reason:  "Vibroblade: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  9486,
		SpellID: 13527,
		PPM:     1.95,
		Reason:  "Supercharger Battle Axe: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  9651,
		SpellID: 18081,
		PPM:     1.95,
		Reason:  "Gryphon Rider's Stormhammer: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  10758,
		SpellID: 1293787,
		PPM:     2.1,
		Reason:  "X'caliboar: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  10797,
		SpellID: 16413,
		PPM:     2.1,
		Reason:  "Firebreather: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  10803,
		SpellID: 18088,
		PPM:     1.95,
		Reason:  "Blade of the Wretched: column reads 100; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  10805,
		SpellID: 1299869,
		PPM:     1.65,
		Reason:  "Eater of the Dead: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  11809,
		SpellID: 16559,
		PPM:     1.9,
		Reason:  "Flame Wrath: no proc row",
		Source:  "EV-0151",
	},
	{
		ItemID:  11902,
		SpellID: 18089,
		PPM:     1.95,
		Reason:  "Linken's Sword of Mastery: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  12992,
		SpellID: 1299850,
		PPM:     2.4,
		Reason:  "Searing Blade: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  13032,
		SpellID: 17510,
		PPM:     1,
		Reason:  "Sword of Corruption: column reads 100; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  13035,
		SpellID: 17511,
		PPM:     1.1,
		Reason:  "Serpent Slicer: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  13051,
		SpellID: 18214,
		PPM:     1,
		Reason:  "Witchfury: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  13054,
		SpellID: 14126,
		PPM:     1.1,
		Reason:  "Grim Reaper: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  13204,
		SpellID: 17315,
		PPM:     3.2,
		Reason:  "Bashguuder: column reads 101; shares the trigger with Rivenspike; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  13285,
		SpellID: 17407,
		PPM:     2.05,
		Reason:  "The Blackrock Slicer: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  13286,
		SpellID: 17315,
		PPM:     1.6,
		Reason:  "Rivenspike: column reads 101; shares the trigger with Bashguuder; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  13401,
		SpellID: 17505,
		PPM:     2.1,
		Reason:  "The Cruel Hand of Timmy: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  13984,
		SpellID: 18276,
		PPM:     2.2,
		Reason:  "Darrowspike: no proc row; shares the trigger with Bonechill Hammer; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  14024,
		SpellID: 19755,
		PPM:     2.35,
		Reason:  "Frightalon: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  14487,
		SpellID: 18276,
		PPM:     1.8,
		Reason:  "Bonechill Hammer: no proc row; shares the trigger with Darrowspike; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  15814,
		SpellID: 16406,
		PPM:     1.1,
		Reason:  "Hameya's Slayer: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  17738,
		SpellID: 21952,
		PPM:     2.45,
		Reason:  "Claw of Celebras: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  18483,
		SpellID: 1302348,
		PPM:     4.25,
		Reason:  "Mana Channeling Wand: equip trigger, column reads 100; wand shots roll it; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  273819,
		SpellID: 1292573,
		PPM:     2.8,
		Reason:  "Boneslicer: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  273841,
		SpellID: 1292679,
		PPM:     1.75,
		Reason:  "Twilight Maul: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  274265,
		SpellID: 1293518,
		PPM:     2.35,
		Reason:  "Brush Cutter: no proc row; shares the trigger with Deadhead Blade; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  274271,
		SpellID: 1293518,
		PPM:     2,
		Reason:  "Deadhead Blade: no proc row; shares the trigger with Brush Cutter; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  274484,
		SpellID: 1294226,
		PPM:     2.15,
		Reason:  "Axe of Inevitable Ruin: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  274952,
		SpellID: 1295736,
		PPM:     1.75,
		Reason:  "Jom-jom's Gabbar: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  281665,
		SpellID: 1312176,
		PPM:     3.15,
		Reason:  "Twilight Forgehammer: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  285332,
		SpellID: 1318250,
		PPM:     2.45,
		Reason:  "Puncturing Spear: no proc row; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  286650,
		SpellID: 1320925,
		PPM:     2.05,
		Reason:  "Pyromancer's Baton: column reads 101; a log fit not yet validated (EV-0136)",
		Source:  "observed",
	},
	{
		ItemID:  12798,
		SpellID: 16928,
		PPM:     1,
		Reason:  "Annihilator: column reads 101; TBC 1 PPM, unverified on Forever",
		Source:  "tbc-carryover",
	},
	{
		ItemID:  880,
		SpellID: 8552,
		PPM:     1.4, // TODO: Vanilla value - verify in-game
		Reason:  "Staff of Horrors: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  899,
		SpellID: 18077,
		PPM:     2, // TODO: Vanilla value - verify in-game
		Reason:  "Venom Web Fang: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  3336,
		SpellID: 18078,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Flesh Piercer: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  4090,
		SpellID: 13496,
		PPM:     2, // TODO: Vanilla value - verify in-game
		Reason:  "Mug O' Hurt: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  5182,
		SpellID: 18092,
		PPM:     0.5, // TODO: Vanilla value - verify in-game
		Reason:  "Shiver Blade: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  6622,
		SpellID: 8191,
		PPM:     1.8, // TODO: Vanilla value - verify in-game
		Reason:  "Sword of Zeal: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  10847,
		SpellID: 13049,
		PPM:     4.5,
		Reason:  "Dragon's Call: no proc row; measured on landed main-hand hits outside the 45 s cooldown",
		Source:  "EV-0021",
	},
	{
		ItemID:  12794,
		SpellID: 16921,
		PPM:     0.5, // TODO: Vanilla value - verify in-game
		Reason:  "Masterwork Stormhammer: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  13053,
		SpellID: 18211,
		PPM:     0.75, // TODO: Vanilla value - verify in-game
		Reason:  "Doombringer: no proc row; shares the trigger with Ebon Hand",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  13060,
		SpellID: 16405,
		PPM:     3, // TODO: Vanilla value - verify in-game
		Reason:  "The Needler: column reads 101; shares the trigger with Hanzo Sword",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  17073,
		SpellID: 21152,
		PPM:     2.1, // TODO: Vanilla value - verify in-game
		Reason:  "Earthshaker: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  17182,
		SpellID: 21162,
		PPM:     0.7, // TODO: Vanilla value - verify in-game
		Reason:  "Sulfuras, Hand of Ragnaros: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  17704,
		SpellID: 16407,
		PPM:     1.2, // TODO: Vanilla value - verify in-game
		Reason:  "Edge of Winter: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  17733,
		SpellID: 21951,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Fist of Stone: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  19019,
		SpellID: 21992,
		PPM:     6, // TODO: Vanilla value - verify in-game
		Reason:  "Thunderfury, Blessed Blade of the Windseeker: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  19099,
		SpellID: 18398,
		PPM:     1.4, // TODO: Vanilla value - verify in-game
		Reason:  "Glacial Blade: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  19169,
		SpellID: 23605,
		PPM:     2, // TODO: Vanilla value - verify in-game
		Reason:  "Nightfall: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  19170,
		SpellID: 18211,
		PPM:     0.4, // TODO: Vanilla value - verify in-game
		Reason:  "Ebon Hand: no proc row; shares the trigger with Doombringer",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  19323,
		SpellID: 23454,
		PPM:     0.5, // TODO: Vanilla value - verify in-game
		Reason:  "The Unstoppable Force: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  19324,
		SpellID: 1290950,
		PPM:     0.4, // TODO: Vanilla value - verify in-game
		Reason:  "The Lobotomizer: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  21134,
		SpellID: 26108,
		PPM:     0.4, // TODO: Vanilla value - verify in-game
		Reason:  "Dark Edge of Insanity: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  2243,
		SpellID: 18803,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Hand of Edward the Odd: column reads 100; vanilla figures differ, 1.0 from 21 procs in 757 hits against 2.0",
		Source:  "vanilla-carryover",
	},
	// Measured on main-hand white swings with no other hit or debuff of the wearer beside them, each proc given
	// to the hit it logs with (EV-0145).
	{
		ItemID:  12469,
		SpellID: 10370,
		PPM:     4.8,
		Reason:  "Mutilator: column reads 101",
		Source:  "EV-0145",
	},
	// TODO: The rates below are vanilla carry-overs, simulator assumptions or old community tests, and none
	// has been measured on this server. Replace each with a measured rate, its Source the id of a
	// docs/evidence file, when logs exist.
	{
		ItemID:  12777,
		SpellID: 16898,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Blazing Rapier: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  12790,
		SpellID: 16916,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Arcanite Champion: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  12792,
		SpellID: 18082,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Volcanic Hammer: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  12795,
		SpellID: 13318,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Blood Talon: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  12797,
		SpellID: 16927,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Frostguard: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  17193,
		SpellID: 21159,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Sulfuron Hammer: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  17223,
		SpellID: 21179,
		PPM:     1.5, // TODO: Vanilla value - verify in-game
		Reason:  "Thunderstrike: column reads 101; vanilla figures differ, 1.5 against 1.0",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  19100,
		SpellID: 23592,
		PPM:     1.4, // TODO: Vanilla value - verify in-game
		Reason:  "Electrified Dagger: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  19287,
		SpellID: 23689,
		PPM:     2, // TODO: Vanilla value - verify in-game
		Reason:  "Darkmoon Card: Heroism: column reads 100",
		Source:  "vanilla-carryover",
	},
	// TODO: waiting for PPM data. The rows below stay commented out until a rate is known.
	// {
	// 	ItemID:  1986,
	// 	SpellID: 18090,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Gutrender: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  3854,
	// 	SpellID: 13439,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Frost Tiger Blade: column reads 101; Hammer of the Northern Wind casts the same spell (EV-0136)",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  5028,
	// 	SpellID: 1312330,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Lord Sakrasis' Scepter: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  7959,
	// 	SpellID: 9796,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Blight: no proc row",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  7960,
	// 	SpellID: 9800,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Truesilver Champion: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  11607,
	// 	SpellID: 15280,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Dark Iron Sunderer: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  12769,
	// 	SpellID: 16871,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Bleakwood Hew: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  13146,
	// 	SpellID: 29647,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Shell Launcher Shotgun: no proc row",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  13262,
	// 	SpellID: 18112,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Ashbringer: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  16004,
	// 	SpellID: 29641,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Dark Iron Rifle: no proc row",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  17002,
	// 	SpellID: 17511,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Ichor Spitter: column reads 101; shares the trigger with Serpent Slicer",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  260186,
	// 	SpellID: 16784,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Darkblade: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  275645,
	// 	SpellID: 1297357,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Reforged Spear: no proc row",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  282561,
	// 	SpellID: 1314040,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Siegebreaker's Blaster: no proc row",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  284262,
	// 	SpellID: 1315767,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Howling Hide: no proc row",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  285102,
	// 	SpellID: 1317583,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Darkmist Dirk: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
}

// The procs-per-minute rate stated for a weapon's chance-on-hit spell, or zero where none is.
func WeaponProcPPM(itemID int32, spellID int32) float64 {
	for _, proc := range WeaponProcs {
		if proc.ItemID == itemID && proc.SpellID == spellID {
			return proc.PPM
		}
	}
	return 0
}
