package overrides

// A spell a server-side handler casts off another, which no effect edge, $<id> token or skill-line
// row names. Reason and Source weigh the same as an Override's, and the generator refuses an entry
// without a reason.
type HandTrigger struct {
	Spell    int32
	Triggers int32
	Replaces int32 // A $<id> the tooltip names in place of Triggers, which the family's triggered ladder leaves out; 0 for none.
	Reason   string
	Source   string
}

// The trigger edges the generator adds by hand to the store.
var HandTriggers = []HandTrigger{
	{
		Spell:    20230,
		Triggers: 20240,
		Reason: "Retaliation's dummy aura (aura 4) casts the counterattack 20240: same name, class set and " +
			"icon, weapon damage with no base, a cost of 1 in SpellPower that is a tenth of a rage",
		Source: "game-files-rows",
	},
	{
		Spell:    1314103,
		Triggers: 1314104,
		Reason: "Improved Seal of Fury's talent aura (three dummies) casts the mana restore 1314104: same name " +
			"and icon, a mana energize whose tooltip reads its numbers off 1314103 and whose aura text is Restoring Mana",
		Source: "game-files-rows",
	},
	{
		Spell:    1311074,
		Triggers: 1311077,
		Reason: "Sanctified Judgement's talent aura (two dummies) casts the mana refund 1311077: same name and icon, " +
			"a mana energize whose tooltip reads Judging restores a portion of the Judged Seal's Mana cost",
		Source: "game-files-rows",
	},
	{
		Spell:    1310905,
		Triggers: 1280353,
		Replaces: 20924,
		Reason: "Consecrated Ground's talent aura (a dummy) rides the mark Consecration's ground puts on the first 4 enemies " +
			"inside it, 1280353: same name and class set, the bonus damage's dummy with its 0.042 coefficient and a " +
			"Holy damage from caster mod whose points the server sets from the talent; the tooltip names the " +
			"Consecration rank 20924 instead. In the log the mark lands at the cast and drops 0.6 sec after the last tick",
		Source: "combat log, Consecration Rank 2 with Consecrated Ground 2/2, 2026-10-17",
	},
	{
		Spell:    1323967,
		Triggers: 1323969,
		Reason: "Gore Drinker's talent aura (two dummies) applies the buff its tooltip describes: same name, class " +
			"set and icon, 3 charges heard on melee attacks, and its aura 231 casts the heal 1323968",
		Source: "game-files-rows",
	},
	{
		Spell:    24574,
		Triggers: 24575,
		Reason: "Zandalarian Hero Badge's on-use Brittle Armor applies the stacking buff its tooltip describes: " +
			"same name, 200 armor (A_MOD_ARMOR_NO_MODS) and 3 defense a stack, the 2000 and 30 of ten stacks, and " +
			"the script effect 24590 of its proc takes a stack off per hit taken",
		Source: "game-files-rows",
	},
	{
		Spell:    16389,
		Triggers: 16368,
		Reason: "Flametongue Totem Proc is a dummy holding the base points, and the hit logs in game as 16368 " +
			"Flametongue Attack: fire school damage with the magic defense type and no bonus coefficient",
		Source: "combat log, Flametongue Totem Rank 1, 2026-10-03",
	},
	{
		Spell:    408341,
		Triggers: 408423,
		Replaces: 8349,
		Reason:   "Fire Nova Rank 1 hits for 408423, the level 12 Fire Nova Rank 1 row, not the 8349 its tooltip reads",
		Source:   "combat log, Fire Nova Ranks 1 and 2, 2026-10-07",
	},
	{
		Spell:    408342,
		Triggers: 408424,
		Replaces: 8502,
		Reason:   "Fire Nova Rank 2 hits for 408424, the level 22 Fire Nova Rank 2 row, not the 8502 its tooltip reads",
		Source:   "combat log, Fire Nova Ranks 1 and 2, 2026-10-07",
	},
	{
		Spell:    408343,
		Triggers: 408426,
		Replaces: 8503,
		Reason:   "Fire Nova Rank 3 hits for 408426, the level 32 Fire Nova Rank 3 row, as ranks 1 and 2 hit for theirs",
		Source:   "combat log, Fire Nova Ranks 1 and 2, 2026-10-07",
	},
	{
		Spell:    408344,
		Triggers: 408427,
		Replaces: 11306,
		Reason:   "Fire Nova Rank 4 hits for 408427, the level 42 Fire Nova Rank 4 row, as ranks 1 and 2 hit for theirs",
		Source:   "combat log, Fire Nova Ranks 1 and 2, 2026-10-07",
	},
	{
		Spell:    408345,
		Triggers: 408428,
		Replaces: 11307,
		Reason:   "Fire Nova Rank 5 hits for 408428, the level 52 Fire Nova Rank 5 row, as ranks 1 and 2 hit for theirs",
		Source:   "combat log, Fire Nova Ranks 1 and 2, 2026-10-07",
	},
	{
		Spell:    1252526,
		Triggers: 1252543,
		Reason: "Caustic Smog Potion's cloud: 1252526's E_CREATE_AREATRIGGER (area trigger 39425) lasts 12 sec and " +
			"puts 1252543 on the enemies inside it; the tooltip reads $1252543s1 Nature damage every $1252543t1 sec, " +
			"and the potion's 1252525 names 1252543 among its tooltip references",
		Source: "tooltip",
	},
}
