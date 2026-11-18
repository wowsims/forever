package overrides

// Which roll an EnchantProc's Value is.
type RateKind uint8

const (
	RatePPM    RateKind = iota + 1 // Procs per minute, measured against the enchanted weapon's speed.
	RateChance                     // A chance per hit, 0 to 1.
)

func (k RateKind) String() string {
	switch k {
	case RatePPM:
		return "ppm"
	case RateChance:
		return "chance"
	}
	return "unknown rate kind"
}

// A permanent enchant's proc rate, which the game files keep outside the spell data: the slot's spell
// states no chance, or the 100 or 101 sentinel. The generator puts it onto the enchant's effect for
// the spell, where the sim reads it. Reason and Source weigh the same as an Override's.
type EnchantProc struct {
	EnchantID int32
	SpellID   int32
	Kind      RateKind
	Value     float64
	Reason    string
	Source    string // "EV-<n>" (docs/evidence), "tbc-carryover", "vanilla-carryover", "observed", "estimate"
}

var EnchantProcs = []EnchantProc{
	{
		EnchantID: 803,
		SpellID:   13897,
		Kind:      RatePPM,
		Value:     6, // TODO: Measure on a one-hander and in the off hand; EV-0130 has two two-handers.
		Reason:    "Fiery Weapon: 13897 states no proc chance, and the tooltip says 'often'",
		Source:    "EV-0130",
	},
	{
		EnchantID: 1900,
		SpellID:   20007,
		Kind:      RatePPM,
		Value:     1,
		Reason:    "Crusader: 20007 states no proc chance, and 458112's 100 is the sentinel beside 'often'; measured on the enchanted hands' landed melee hits",
		Source:    "EV-0135",
	},
	{
		EnchantID: 8217,
		SpellID:   1248806,
		Kind:      RateChance,
		Value:     0.08, // TODO: Model how the chance falls as the wearer's crit from gear and intellect rises, and how it differs by class (EV-0149).
		Reason: "Revelation: 1248806's 100 is the sentinel; the tooltip says the chance is diminished as the " +
			"wearer's critical strike chance increases, and states no number; the rate is the owner's round " +
			"figure from logs (EV-0129)",
		Source: "observed",
	},
}

// The rate stated for an enchant's proc spell, or nil where none is.
func EnchantProcRate(enchantID int32, spellID int32) *EnchantProc {
	for i := range EnchantProcs {
		if EnchantProcs[i].EnchantID == enchantID && EnchantProcs[i].SpellID == spellID {
			return &EnchantProcs[i]
		}
	}
	return nil
}
