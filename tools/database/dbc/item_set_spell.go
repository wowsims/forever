package dbc

// A spell an item set grants once enough of its pieces are worn.
type ItemSetSpell struct {
	ItemSetID int // Item set ID
	Threshold int // Pieces worn
	SpellID   int // Spell ID
}
