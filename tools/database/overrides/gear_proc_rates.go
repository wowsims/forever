package overrides

// Which kind of gear a GearProcRate names.
type GearKind uint8

const (
	GearItem     GearKind = iota + 1 // An item, by item id.
	GearEnchant                      // An enchant, by enchant effect id.
	GearSetBonus                     // A set bonus, by set id and piece count.
)

func (g GearKind) String() string {
	switch g {
	case GearItem:
		return "item"
	case GearEnchant:
		return "enchant"
	case GearSetBonus:
		return "set bonus"
	}
	return "unknown gear kind"
}

// The rate the gear tab shows for hand-written gear whose code rolls something its rows do not state: it
// scales the chance by target type, halves it by area, adds a flat chance or splits the procs. The generator
// reads the gear's rate from its proc rows, and a row here replaces it. With no Kind the gear states no single
// rate and shows none; with one it shows Rate, though gear listed unvalidated or missing its effect still shows
// none. Only the generator reads this table; the sim rolls what its code rolls. Reason and Source weigh the
// same as an Override's.
type GearProcRate struct {
	Gear   GearKind
	ID     int32    // The item id, the enchant effect id or the set id.
	Pieces int32    // The set bonus's piece count; 0 on an item or an enchant.
	Kind   RateKind // 0 where the gear states no single rate.
	Rate   float64  // Procs per minute, or a chance from 0 to 1.
	Reason string
	Source string // "EV-<n>" (docs/evidence), "tooltip", "tbc-carryover", "vanilla-carryover" or "observed"; empty only with no Kind.
}

var GearProcRates = []GearProcRate{
	{
		Gear:   GearItem,
		ID:     11684,
		Reason: "Ironfoe: 3%, or the column's 6% against Orcs (NewTargetScaledChanceManager)",
		Source: "EV-0119",
	},
	{
		Gear:   GearItem,
		ID:     13243,
		Reason: "Argent Defender: 0.5 ppm on the attacker's swing interval, or twice that when an Undead strikes (NewAttackerSwingPPMManager)",
		Source: "EV-0126",
	},
	{
		Gear:   GearItem,
		ID:     260205,
		Reason: "Highborne Research Tablet: 2.6 ppm on weapon hits plus a flat 7% per spell hit",
		Source: "EV-0131",
	},
	{
		Gear:   GearItem,
		ID:     280032,
		Reason: "Bolt's Charged Eyeball: 15%, and half the procs land on the wearer",
	},
	{
		Gear:   GearItem,
		ID:     280080,
		Kind:   RatePPM,
		Rate:   1.35,
		Reason: "Claymore of Two Moons: 1.35 ppm, each proc Waxing Moon or Waning Moon at even odds",
		Source: "observed",
	},
}
