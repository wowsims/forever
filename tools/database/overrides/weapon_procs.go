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
	Source  string // "tbc-carryover", "vanilla-carryover", "wcl:<report>/<fight>", "issue #N"
}

// The rates measured on this server come first; every one after them is a carry-over, as its Source says. A
// hotfix that adds back a weapon whose row was dropped needs the row restored from git history, or the database
// tests that name the item fail the Update Items run.
var WeaponProcs = []WeaponProc{
	// Measured in end-game dungeon logs: the procs over the landed hits of the weapon carrying them, white swings
	// and specials alike, as the sim's weapon proc listens, at the weapon's speed (a druid's at the paw's). Shield
	// Slam, Shield Bash and Shank hits are left out, since they roll a chance-on-hit weapon about once in a thousand.
	{
		ItemID:  9418,
		SpellID: 12731,
		PPM:     2.5,
		Reason:  "Stoneslayer: no proc row; 619 procs on landed hits, 4 end-game dungeon runs, 1 player",
		Source:  "wcl:NHa2RqtjxAvPFwLg/8 and 3 more runs, 2026-11-12",
	},
	{
		ItemID:  10750,
		SpellID: 1300356,
		PPM:     1.5,
		Reason:  "Lifeforce Dirk: no proc row; 161 procs on landed hits, 2 end-game dungeon runs, 2 players, off hand",
		Source:  "wcl:jT7NDMWFfhKk1bdz/1 and 1 more run, 2026-11-12",
	},
	{
		ItemID:  11744,
		SpellID: 16433,
		PPM:     3.2,
		Reason:  "Bloodfist: no proc row; 91 procs on landed hits, 3 end-game dungeon runs, 2 players, off hand",
		Source:  "wcl:fFMPzh2LjnRTtaAk/1 and 2 more runs, 2026-11-12",
	},
	{
		ItemID:  11803,
		SpellID: 18086,
		PPM:     2,
		Reason:  "Force of Magma: no proc row; 398 procs on landed hits, 3 end-game dungeon runs, 2 players",
		Source:  "wcl:fgMRXWJKHzaLt2Gh/1 and 2 more runs, 2026-11-12",
	},
	{
		ItemID:  12463,
		SpellID: 14118,
		PPM:     1.5,
		Reason:  "Drakefang Butcher: column reads 101; 146 procs on landed hits, 1 end-game dungeon run, 1 player",
		Source:  "wcl:LQG6jWCnAP1gwvdt/1, 2026-11-12",
	},
	{
		ItemID:  12582,
		SpellID: 16528,
		PPM:     1.7,
		Reason:  "Keris of Zul'Serak: no proc row; 295 procs on landed hits, 6 end-game dungeon runs, 2 players, dual wielded",
		Source:  "wcl:cvRTFb1qhGBKNkCD/45 and 5 more runs, 2026-11-12",
	},
	{
		ItemID:  12969,
		SpellID: 17196,
		PPM:     0.9,
		Reason:  "Seeping Willow: column reads 101; 78 procs on landed hits, 4 end-game dungeon runs, 2 players, cat form, at the 1.0 s paw",
		Source:  "wcl:T79gRkJz4QVGbyYW/9 and 3 more runs, 2026-11-12",
	},
	{
		ItemID:  13183,
		SpellID: 18203,
		PPM:     1.7,
		Reason:  "Venomspitter: column reads 101; 157 procs on landed hits, 4 end-game dungeon runs, 3 players, dual wielded",
		Source:  "wcl:6AX1yxVMGKT7zbqL/1 and 3 more runs, 2026-11-12",
	},
	{
		ItemID:  13348,
		SpellID: 17483,
		PPM:     1.1,
		Reason:  "Demonshear: column reads 101; 67 procs on landed hits, 1 end-game dungeon run, 1 player",
		Source:  "wcl:DpGg6kA8nyMQ1Hzh/1, 2026-11-12",
	},
	{
		ItemID:  13983,
		SpellID: 18289,
		PPM:     1.9,
		Reason:  "Gravestone War Axe: no proc row; 116 procs on landed hits, 3 end-game dungeon runs, 1 player",
		Source:  "wcl:NRK61HZcDdJkrwXG/5 and 2 more runs, 2026-11-12",
	},
	{
		ItemID:  17730,
		SpellID: 21949,
		PPM:     1.2,
		Reason:  "Gatorbite Axe: column reads 101; 188 procs on landed hits, 11 end-game dungeon runs, 2 players",
		Source:  "wcl:kW9V4CMKfmBbtrRH/3 and 10 more runs, 2026-11-12",
	},
	{
		ItemID:  17753,
		SpellID: 29655,
		PPM:     3.5,
		Reason:  "Verdant Keeper's Aim: no proc row; 6839 procs on landed ranged hits, 33 end-game dungeon runs, 14 players; melee hits do not roll it",
		Source:  "wcl:DpGg6kA8nyMQ1Hzh/1 and 32 more runs, 2026-11-12",
	},
	{
		ItemID:  17766,
		SpellID: 21961,
		PPM:     1.8,
		Reason:  "Princess Theradras' Scepter: column reads 101; 307 procs on landed hits, 2 end-game dungeon runs, a shaman and a cat druid",
		Source:  "wcl:x3qQt7yf9mAgbTrZ/1 and 1 more run, 2026-11-12",
	},
	{
		ItemID:  18388,
		SpellID: 1302279,
		PPM:     1.7,
		Reason:  "Stoneshatter: no proc row; 146 procs on landed hits, 3 end-game dungeon runs, 1 player",
		Source:  "wcl:YH7n2xZqPKzWjQrR/6 and 2 more runs, 2026-11-12",
	},
	{
		ItemID:  274270,
		SpellID: 1293568,
		PPM:     1.1,
		Reason:  "Weed Whacker: no proc row; 119 procs on landed hits, 1 end-game dungeon run, 1 player",
		Source:  "wcl:LQG6jWCnAP1gwvdt/1, 2026-11-12",
	},
	{
		ItemID:  275441,
		SpellID: 1297921,
		PPM:     2.3,
		Reason:  "Cremation Longbow: no proc row; 292 procs on landed hits, 3 end-game dungeon runs, 1 player",
		Source:  "wcl:NgRFKH1PWzykqC9Q/12 and 2 more runs, 2026-11-12",
	},
	{
		ItemID:  276419,
		SpellID: 1299631,
		PPM:     3.6,
		Reason:  "Lightcrest Hammer: no proc row; 148 procs on landed hits, 1 end-game dungeon run, 1 player, main hand while dual wielding",
		Source:  "wcl:AHaVwhrgcm9NM2TZ/19, 2026-11-12",
	},
	{
		ItemID:  12798,
		SpellID: 16928,
		PPM:     1,
		Reason:  "Annihilator: column reads 101; TBC 1 PPM, unverified on Forever",
		Source:  "tbc-carryover",
	},
	{
		ItemID:  754,
		SpellID: 13519,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Shortsword of Vengeance: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  810,
		SpellID: 13439,
		PPM:     3.5, // TODO: Vanilla value - verify in-game
		Reason:  "Hammer of the Northern Wind: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  870,
		SpellID: 18796,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Fiery War Axe: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  871,
		SpellID: 18797,
		PPM:     1.9, // TODO: Vanilla value - verify in-game
		Reason:  "Flurry Axe: column reads 101",
		Source:  "vanilla-carryover",
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
		ItemID:  937,
		SpellID: 18138,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Black Duskwood Staff: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  1481,
		SpellID: 1292158,
		PPM:     1.25, // TODO: Vanilla value - verify in-game
		Reason:  "Grimclaw: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  1726,
		SpellID: 16401,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Poison-Tipped Bone Spear: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  2000,
		SpellID: 18091,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Archeus: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  2256,
		SpellID: 13440,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Skeletal Club: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  2912,
		SpellID: 16409,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Claw of the Shadowmancer: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  2942,
		SpellID: 13491,
		PPM:     2, // TODO: Vanilla value - verify in-game
		Reason:  "Iron Knuckles: no proc row",
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
		ItemID:  6220,
		SpellID: 1292152,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Meteor Shard: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  6472,
		SpellID: 1291663,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Stinging Viper: column reads 101",
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
		ItemID:  6660,
		SpellID: 8348,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Julie's Dagger: column reads 100",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  6904,
		SpellID: 8313,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Bite of Serra'kis: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  7730,
		SpellID: 18204,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Cobalt Crusher: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  7961,
		SpellID: 9806,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Phantom Blade: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  8190,
		SpellID: 16405,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Hanzo Sword: column reads 101; shares the trigger with The Needler",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  9453,
		SpellID: 11790,
		PPM:     0.8, // TODO: Vanilla value - verify in-game
		Reason:  "Toxic Revenger: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  11902,
		SpellID: 18089,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Linken's Sword of Mastery: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  12794,
		SpellID: 16921,
		PPM:     0.5, // TODO: Vanilla value - verify in-game
		Reason:  "Masterwork Stormhammer: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  12974,
		SpellID: 14106,
		PPM:     0.65, // TODO: Vanilla value - verify in-game
		Reason:  "The Black Knight: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  12992,
		SpellID: 1299850,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Searing Blade: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  13016,
		SpellID: 13318,
		PPM:     0.85, // TODO: Vanilla value - verify in-game
		Reason:  "Killmaim: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  13032,
		SpellID: 17510,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Sword of Corruption: column reads 100",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  13051,
		SpellID: 18214,
		PPM:     1.25, // TODO: Vanilla value - verify in-game
		Reason:  "Witchfury: no proc row",
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
		ItemID:  13054,
		SpellID: 14126,
		PPM:     0.87, // TODO: Vanilla value - verify in-game
		Reason:  "Grim Reaper: column reads 101",
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
		ItemID:  17705,
		SpellID: 21919,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Thrash Blade: no proc row",
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
		ItemID:  1982,
		SpellID: 18211,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Nightblade: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  2164,
		SpellID: 18107,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Gut Ripper: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  2243,
		SpellID: 18803,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Hand of Edward the Odd: column reads 100; vanilla figures differ, 1.0 from 21 procs in 757 hits against 2.0",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  2825,
		SpellID: 29638,
		PPM:     3.35, // TODO: Vanilla value - verify in-game
		Reason:  "Bow of Searing Arrows: no proc row",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  7717,
		SpellID: 9632,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Ravager: no proc row",
		Source:  "vanilla-carryover",
	},
	// TODO: The rates below are vanilla carry-overs, simulator assumptions or old community tests, and none
	// has been measured on this server. Replace each with a measured rate (Source wcl:<report>/<fight> or
	// issue #N) when logs exist.
	{
		ItemID:  9651,
		SpellID: 18081,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Gryphon Rider's Stormhammer: no proc row",
		Source:  "vanilla-carryover",
	},
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
		ItemID:  13035,
		SpellID: 17511,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Serpent Slicer: column reads 101",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  13040,
		SpellID: 29626,
		PPM:     3, // TODO: Vanilla value - verify in-game
		Reason:  "Heartseeking Crossbow: equip trigger, column reads 100",
		Source:  "vanilla-carryover",
	},
	{
		ItemID:  15814,
		SpellID: 16406,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Hameya's Slayer: column reads 101",
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
	// 	ItemID:  869,
	// 	SpellID: 13752,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Dazzling Longsword: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  1986,
	// 	SpellID: 18090,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Gutrender: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  3078,
	// 	SpellID: 1292647,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Naga Heartpiercer: no proc row",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  3854,
	// 	SpellID: 13439,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Frost Tiger Blade: column reads 101; Hammer of the Northern Wind casts the same spell at 3.5 PPM",
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
	// 	ItemID:  8708,
	// 	SpellID: 1317432,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Hammer of Expertise: no proc row",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  10772,
	// 	SpellID: 18075,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Glutton's Cleaver: column reads 101",
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
	// 	ItemID:  274425,
	// 	SpellID: 1294055,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "White Obsidian Wand: equip trigger, column reads 100",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  274963,
	// 	SpellID: 1295744,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Rot-Covered Harpoon: column reads 101",
	// 	Source:  "vanilla-carryover",
	// },
	// {
	// 	ItemID:  275439,
	// 	SpellID: 1296730,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Blunderflame Bone Bow: no proc row",
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
	// 	ItemID:  281665,
	// 	SpellID: 1312176,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Twilight Forgehammer: column reads 101",
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
