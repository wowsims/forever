// Package overrides holds the numbers the spelldata store carries that the game files do
// not state. Each row names the spell, the field, the value and where the value comes from, and the
// generator refuses to run when the game files start stating one themselves - a value kept by hand
// after the data grew a column for it is a wrong number nobody would notice.
//
// A leaf package on purpose: tools/database imports it, it imports nothing back, and the folder it
// shares with the SQL overrides is read for *.sql alone.
package overrides

import "github.com/wowsims/forever/sim/core"

// Which field of the row the value belongs on. Every one of them names a rule the generator checks
// before it writes the value, so an override outlives its reason no longer than the next
// regeneration.
type Field uint8

const (
	// Procs per minute, onto Spell.RPPM. Fails when SpellAuraOptions.SpellProcsPerMinuteID stops
	// being zero on the row.
	PPM Field = iota + 1

	// A flat threat bonus, onto Spell.FlatThreat. Fails when the spell gains an E_THREAT or
	// E_THREAT_ALL effect.
	FlatThreat

	// The attack-power coefficient of the direct effect, onto Effect.APCoef. Fails when the game files
	// state a BonusCoefficientFromAP there.
	APCoefDirect

	// The same for the ticking effect.
	APCoefPeriodic

	// A proc chance in percent, onto Spell.ProcChance. Fails when the tooltip states a chance of
	// its own through $h or $mN, because then the game files already say where the roll is.
	ProcChancePct

	// The aura's duration in milliseconds. Fails when the game files state a SpellDuration.
	DurationMs

	// The ProcHint bits the tooltip's wording did not yield, which Value carries as a bit set. It
	// replaces what the wording produced rather than adding to it.
	Hint

	// The stat mask of the A_MOD_TOTAL_STAT_PERCENTAGE effect, onto its second misc value, one bit
	// per stat index. Fails when the effect states either misc value itself, since 0 and 0 is
	// Strength by index.
	StatMask
)

func (f Field) String() string {
	switch f {
	case PPM:
		return "PPM"
	case FlatThreat:
		return "FlatThreat"
	case APCoefDirect:
		return "APCoefDirect"
	case APCoefPeriodic:
		return "APCoefPeriodic"
	case ProcChancePct:
		return "ProcChancePct"
	case DurationMs:
		return "DurationMs"
	case Hint:
		return "Hint"
	case StatMask:
		return "StatMask"
	}
	return "unknown field"
}

// One hand-supplied value. Reason and Source are not documentation: they are what the next reader
// has to weigh the number against, so the generator refuses a row without a reason.
type Override struct {
	SpellID int32
	Field   Field
	Value   float64
	Reason  string
	Source  string // "tbc-carryover", "tooltip", "wcl:<report>/<fight>", "issue #N"
}

// Tooltip-only area bonuses ("doubled in X areas"). Groups are AreaGroup ids; any counts.
type AreaBonus struct {
	SpellID            int32
	Groups             []int32
	Multiplier         float64
	DurationMultiplier float64
	Reason             string
	Source             string
}

var AreaBonuses = []AreaBonus{
	{
		SpellID:            1249113,
		Groups:             []int32{9203},
		Multiplier:         2,
		DurationMultiplier: 1,
		Reason:             "Molten Fury: 'This effect is doubled in Volcanic areas'; no companion row, and 9203 is the Blackrock group no spell requires",
		Source:             "tooltip",
	},
	{
		SpellID:            1249110,
		Groups:             []int32{9164},
		Multiplier:         2,
		DurationMultiplier: 1,
		Reason:             "Chthonic Power: 'This effect is doubled in Snowy areas'; no companion row",
		Source:             "tooltip",
	},
	{
		SpellID:            1287571,
		Groups:             []int32{9161},
		Multiplier:         2,
		DurationMultiplier: 2,
		Reason:             "Monkey Business: 'Effect and duration are doubled in Forest and Grassland areas'; no companion row",
		Source:             "tooltip",
	},
	{
		SpellID:            1291105,
		Groups:             []int32{9326},
		Multiplier:         2,
		DurationMultiplier: 1,
		Reason:             "Stabilized: 'This effect is doubled in Strongholds and Cities'; no companion row",
		Source:             "tooltip",
	},
	{
		SpellID:            1293820,
		Groups:             []int32{9326},
		Multiplier:         2,
		DurationMultiplier: 1,
		Reason:             "Toy Soldier: 'This effect is doubled in Strongholds and Cities'; no companion row",
		Source:             "tooltip",
	},
}

// The procs whose rate the game files keep outside the spell data and that belong to the spell
// rather than to one weapon: a chance-on-hit weapon's rate is its own and lives in WeaponProcs. The
// ProcChance column reads the 100 or 101 sentinel, which the store would otherwise take for "fires
// on every hit". Each rate is a carry-over, as its Source says; none has been measured on this server.
var Spells = []Override{
	{
		SpellID: 23686,
		Field:   PPM,
		Value:   1,
		Reason:  "Darkmoon Card: Maelstrom: column reads 100, tooltip says 'Chance to strike'",
		Source:  "tbc-carryover",
	},
	{
		SpellID: 26480,
		Field:   PPM,
		Value:   10,
		Reason:  "Badge of the Swarmguard: stack accumulator inside the on-use window",
		Source:  "tbc-carryover",
	},

	// Melee procs whose tooltips name no trigger condition, ruled to hear only hits with a damage
	// component: a hit that deals damage, or the landed application of a DoT such as Rend's.
	{
		SpellID: 7711,
		Field:   Hint,
		Value:   float64(core.ProcHintDamage),
		Reason:  "Blazefury Medallion: fires on hits with a damage component",
		Source:  "owner decision, PR #99",
	},
	{
		SpellID: 7721,
		Field:   Hint,
		Value:   float64(core.ProcHintDamage),
		Reason:  "Fiery Plate Gauntlets: fires on hits with a damage component",
		Source:  "owner decision, PR #99",
	},
	{
		SpellID: 16615,
		Field:   Hint,
		Value:   float64(core.ProcHintDamage),
		Reason:  "Storm Gauntlets: fires on hits with a damage component",
		Source:  "owner decision, PR #99",
	},
	{
		SpellID: 23686,
		Field:   Hint,
		Value:   float64(core.ProcHintDamage),
		Reason:  "Darkmoon Card: Maelstrom: fires on hits with a damage component",
		Source:  "owner decision, PR #99",
	},
	{
		SpellID: 26480,
		Field:   Hint,
		Value:   float64(core.ProcHintDamage),
		Reason:  "Badge of the Swarmguard: stacks on hits with a damage component",
		Source:  "owner decision, PR #99",
	},
	{
		SpellID: 1299796,
		Field:   StatMask,
		Value:   16, // TODO: Verify in game that Insight's proc doubles Spirit, not Strength.
		Reason:  "Insight: misc 0 and 0 is Strength by index, the tooltip reads 'Increases your Spirit by 100%'",
		Source:  "tooltip",
	},
	{
		SpellID: 1248751,
		Field:   StatMask,
		Value:   16, // TODO: Verify in game that Mystic Mushroom raises Spirit, not Strength.
		Reason:  "Increased Spirit: misc 0 and 0 is Strength by index, the tooltip reads 'Increases Spirit by 5%'",
		Source:  "tooltip",
	},

	// Attack Power shares measured in game at levels 13 and 14, against the sheet's Attack Power from
	// 99 to 154. Only Rend Rank 2 and Thunder Clap Rank 1 were measured; the other ranks carry the
	// same share until a log at their level says otherwise.
	{
		SpellID: 772,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 1: Rank 2's measured share per tick",
		Source:  "combat log, Rend Rank 2, 2026-10-01",
	},
	{
		SpellID: 6546,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 2: ticks average 7 + 0.02 x Attack Power, read on every tick",
		Source:  "combat log, Rend Rank 2, 2026-10-01",
	},
	{
		SpellID: 6547,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 3: Rank 2's measured share per tick",
		Source:  "combat log, Rend Rank 2, 2026-10-01",
	},
	{
		SpellID: 6548,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 4: Rank 2's measured share per tick",
		Source:  "combat log, Rend Rank 2, 2026-10-01",
	},
	{
		SpellID: 11572,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 5: Rank 2's measured share per tick",
		Source:  "combat log, Rend Rank 2, 2026-10-01",
	},
	{
		SpellID: 11573,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 6: Rank 2's measured share per tick",
		Source:  "combat log, Rend Rank 2, 2026-10-01",
	},
	{
		SpellID: 11574,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 7: Rank 2's measured share per tick",
		Source:  "combat log, Rend Rank 2, 2026-10-01",
	},
	{
		SpellID: 6343,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 1: hits deal 10 + 0.07 x Attack Power, rounded down, before armor",
		Source:  "combat log, Thunder Clap Rank 1, 2026-10-01",
	},
	{
		SpellID: 8198,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 2: Rank 1's measured share",
		Source:  "combat log, Thunder Clap Rank 1, 2026-10-01",
	},
	{
		SpellID: 8204,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 3: Rank 1's measured share",
		Source:  "combat log, Thunder Clap Rank 1, 2026-10-01",
	},
	{
		SpellID: 8205,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 4: Rank 1's measured share",
		Source:  "combat log, Thunder Clap Rank 1, 2026-10-01",
	},
	{
		SpellID: 11580,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 5: Rank 1's measured share",
		Source:  "combat log, Thunder Clap Rank 1, 2026-10-01",
	},
	{
		SpellID: 11581,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 6: Rank 1's measured share",
		Source:  "combat log, Thunder Clap Rank 1, 2026-10-01",
	},
}

// The procs-per-minute rate stated for a spell, or zero where none is. The item database reads it
// for the same spells the store does: a container whose ProcChance column is the "rate lives
// elsewhere" sentinel has no rate anywhere else either.
func PPMFor(spellID int32) float64 {
	for _, override := range Spells {
		if override.SpellID == spellID && override.Field == PPM {
			return override.Value
		}
	}
	return 0
}
