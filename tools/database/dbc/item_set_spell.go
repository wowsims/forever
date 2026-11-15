package dbc

// A spell an item set grants once enough of its pieces are worn.
type ItemSetSpell struct {
	ItemSetID     int // Item set ID
	Threshold     int // Pieces worn
	SpellID       int // Spell ID
	RequiredSkill int // The skill line the set's ItemSet row gates its bonuses on, 0 for none
}
