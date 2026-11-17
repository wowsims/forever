package shared

// An item, enchant or set bonus whose proc rolls the rate of the spell it names unchanged, or the procs
// per minute its item's chance-on-hit entry states for that spell.
type GearProc struct {
	Gear      string // "item", "enchant" or "set bonus".
	ID        int32  // The item id, the enchant effect id or the set id.
	Pieces    int32  // The set bonus's piece count; 0 on an item or an enchant.
	SpellID   int32
	FromEntry bool // The rate is the item's entry's for the spell, and the effect registers nothing while it states none.
}

// The registered procs whose rate the generator can state on the gear tab. Hand-written gear records here
// from init, before the generated registrations run: the row builders when no Rate option replaces the
// rate, bespoke code through the functions below. A proc whose code scales the rate by target type,
// halves it by area or rolls another rate than the spell's records nothing, and its gear shows no rate.
var GearProcs []GearProc

// The item's proc rolls the spell's own rate.
func ItemRollsSpellRate(itemID int32, spellID int32) {
	GearProcs = append(GearProcs, GearProc{Gear: "item", ID: itemID, SpellID: spellID})
}

// The item's proc rolls the procs per minute its chance-on-hit entry states for the spell.
func ItemRollsEntryRate(itemID int32, spellID int32) {
	GearProcs = append(GearProcs, GearProc{Gear: "item", ID: itemID, SpellID: spellID, FromEntry: true})
}

// The enchant's proc rolls the spell's own rate.
func EnchantRollsSpellRate(effectID int32, spellID int32) {
	GearProcs = append(GearProcs, GearProc{Gear: "enchant", ID: effectID, SpellID: spellID})
}

// The set bonus's proc rolls the spell's own rate.
func SetBonusRollsSpellRate(setID int32, pieces int32, spellID int32) {
	GearProcs = append(GearProcs, GearProc{Gear: "set bonus", ID: setID, Pieces: pieces, SpellID: spellID})
}
