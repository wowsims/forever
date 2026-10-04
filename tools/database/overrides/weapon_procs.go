package overrides

// A chance-on-hit weapon's procs per minute, which the game files keep outside the spell data: the
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
