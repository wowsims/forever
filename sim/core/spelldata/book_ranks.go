package spelldata

// The ranks a class book teaches, by spell id, each with the item that teaches it, then the buff auras of the totem
// book ranks and Greater Blessing rank 2. The game files state a book rank's SkillLineAbility row like a trained one,
// so this list is kept by hand. The group buffs (Gift of the Wild, Prayer of Fortitude, Arcane Brilliance, Prayer of
// Shadow Protection) stay out: a book teaches every rank of those, so no trained rank exists to prefer.
var bookRanks = map[int32]bool{
	25286: true, // Manual of Heroic Strike IX
	25289: true, // Manual of Battle Shout VII
	25288: true, // Manual of Revenge VI
	25307: true, // Grimoire of Shadow Bolt X
	25309: true, // Grimoire of Immolate VIII
	25311: true, // Grimoire of Corruption VII
	28610: true, // Grimoire of Shadow Ward IV
	25295: true, // Guide: Serpent Sting IX
	25296: true, // Guide: Aspect of the Hawk VII
	25361: true, // Tablet of Strength of Earth Totem V
	25362: true, // The buff of that rank
	25359: true, // Tablet of Grace of Air Totem III
	25360: true, // The buff of that rank
	25357: true, // Tablet of Healing Wave X
	29228: true, // Tablet of Flame Shock VI
	25304: true, // Tome of Frostbolt XI
	25306: true, // Tome of Fireball XII
	25345: true, // Tome of Arcane Missiles VIII
	28609: true, // Tome of Frost Ward V
	28612: true, // Tome of Conjure Food VII
	25298: true, // Book of Starfire VII
	25297: true, // Book of Healing Touch XI
	25299: true, // Book of Rejuvenation XI
	31018: true, // Book of Ferocious Bite V, which teaches rank 6
	25314: true, // Codex of Greater Heal V
	25316: true, // Codex of Prayer of Healing V
	25315: true, // Codex of Renew X
	25291: true, // Libram: Blessing of Might VII
	25290: true, // Libram: Blessing of Wisdom VI
	25916: true, // Greater Blessing of Might rank 2, the raid's book rank; no item teaches it directly
	25918: true, // Greater Blessing of Wisdom rank 2, the raid's book rank; no item teaches it directly
	25292: true, // Libram: Holy Light IX
	25300: true, // Handbook of Backstab IX
	25302: true, // Handbook of Feint V
	31016: true, // Manual of Eviscerate IX
}

func IsBookRank(id int32) bool {
	return bookRanks[id]
}
