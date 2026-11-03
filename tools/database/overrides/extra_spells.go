package overrides

// A spell the store carries although no class, item, enchant or set bonus reaches it. Reason and
// Source weigh the same as an Override's, and the generator refuses an entry without a reason.
type Extra struct {
	SpellID        int32
	Reason         string
	Source         string
	SpellDataGroup string // "racials" puts the row in the racials spell data group of a partial update; empty follows its class family.
}

// The spells the generator adds to the store's roots by hand.
var ExtraSpells = []Extra{
	{SpellID: 16979, Reason: "Bear form Feral Charge", Source: "game files"},
	{SpellID: 414924, Reason: "Cat form Feral Charge", Source: "game files"},
	{SpellID: 1293697, Reason: "Demonic Brand's Shadow damage when a demon consumes the brand; nothing links it to the talent", Source: "Spell 1293697 description"},
	{SpellID: 1293698, Reason: "Demonic Brand's Fire damage when a demon consumes the brand; nothing links it to the talent", Source: "Spell 1293698 description"},
	{SpellID: 1328570, Reason: "Holy Binding's self heal; nothing links it to the talent", Source: "Spell 1328570"},
	{SpellID: 1249755, Reason: "Summoning Disorientation, a 6 sec stun; no skill line row reaches it and no summon row triggers it", Source: "Spell 1249755"},
	// Enchant Weapon - Revelation 8217: its proc spell 1248806 names no trigger, and each class with
	// spells to crit has a buff of its own, which the sim picks by the wearer's class.
	{SpellID: 1248808, Reason: "Revelation, the mage buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323377, Reason: "Revelation, the priest buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323392, Reason: "Revelation, the warlock buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323400, Reason: "Revelation, the druid buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323410, Reason: "Revelation, the hunter buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323418, Reason: "Revelation, the shaman buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323419, Reason: "Revelation, the paladin buff of enchant 8217", Source: "owner decision"},
	// Shoot, the wand's autorepeat: no class row names it, and the mage, priest and warlock reach it through the equipped wand.
	{SpellID: 5019, Reason: "Shoot, the wand autorepeat of the mage, priest and warlock", Source: "owner decision"},
	// Rockbiter Weapon Proc, the threat Rockbiter Weapon (Passive) adds to each melee hit: a script
	// effect of the passive casts it, so no row names it.
	{SpellID: 20871, Reason: "Rockbiter Weapon Proc, the threat of the shaman's Rockbiter Weapon", Source: "owner decision"},
	{SpellID: 1311703, Reason: "Seal of Command's Echo for Unsealed Wrath; the talent's tooltip names only Echo of Fury", Source: "Spell 1311703"},
	{SpellID: 1311704, Reason: "Seal of Righteousness's Echo for Unsealed Wrath; the talent's tooltip names only Echo of Fury", Source: "Spell 1311704"},
	{SpellID: 1311705, Reason: "Seal of Justice's Echo for Unsealed Wrath; the talent's tooltip names only Echo of Fury", Source: "Spell 1311705"},
	{SpellID: 1328198, Reason: "Seal of Light's Echo for Unsealed Wrath; the talent's tooltip names only Echo of Fury", Source: "Spell 1328198"},
	{SpellID: 1328206, Reason: "Seal of Wisdom's Echo for Unsealed Wrath; the talent's tooltip names only Echo of Fury", Source: "Spell 1328206"},
	// Eureka!, the gnome racial: a row per class, each scoping its mods to that class's abilities,
	// which the sim picks by the gnome's class.
	{SpellID: 1259812, Reason: "Eureka!, the rogue row of the gnome racial", Source: "owner decision", SpellDataGroup: "racials"},
	{SpellID: 1259813, Reason: "Eureka!, the warrior row of the gnome racial", Source: "owner decision", SpellDataGroup: "racials"},
	{SpellID: 1259817, Reason: "Eureka!, the mage row of the gnome racial", Source: "owner decision", SpellDataGroup: "racials"},
	{SpellID: 1259821, Reason: "Eureka!, the warlock row of the gnome racial", Source: "owner decision", SpellDataGroup: "racials"},
	{SpellID: 1259823, Reason: "Eureka!, the priest row of the gnome racial", Source: "owner decision", SpellDataGroup: "racials"},
}
