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

// Each rate is a carry-over, as its Source says; none has been measured on this server.
var WeaponProcs = []WeaponProc{
	{
		ItemID:  12798,
		SpellID: 16928,
		PPM:     1,
		Reason:  "Annihilator: column reads 101; TBC 1 PPM, unverified on Forever",
		Source:  "tbc-carryover",
	},
	{
		ItemID:  810,
		SpellID: 13439,
		PPM:     3.5, // TODO: Vanilla value - verify in-game
		Reason:  "Hammer of the Northern Wind: column reads 101",
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
		ItemID:  7961,
		SpellID: 9806,
		PPM:     1, // TODO: Vanilla value - verify in-game
		Reason:  "Phantom Blade: column reads 101",
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
	// 	ItemID:  17753,
	// 	SpellID: 29655,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Verdant Keeper's Aim: no proc row",
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
	// 	ItemID:  275441,
	// 	SpellID: 1297921,
	// 	PPM:     0, // TODO: waiting for PPM data
	// 	Reason:  "Cremation Longbow: no proc row",
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
