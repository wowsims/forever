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

	// The attack-power coefficient of the E_THREAT or E_THREAT_ALL effect, onto its Effect.APCoef.
	// Fails when the game files state a BonusCoefficientFromAP there.
	APCoefThreat

	// A flat threat bonus of Value times the spell's level, onto Spell.FlatThreat, for a family whose
	// ranks scale their threat with their level. Fails like FlatThreat, and beside a FlatThreat
	// override on the same spell.
	ThreatPerSpellLevel

	// The internal cooldown between two procs in milliseconds, onto Spell.ICDMs. Fails when the game
	// files state a SpellAuraOptions.ProcCategoryRecovery.
	ICDMs
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
	case APCoefThreat:
		return "APCoefThreat"
	case ThreatPerSpellLevel:
		return "ThreatPerSpellLevel"
	case ICDMs:
		return "ICDMs"
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

// A talent curve that states its size with the opposite sign to the row's own base points, where the
// tooltip reads the row's sign: the curve's values are negated. Stale once the curve's sign agrees
// with the row's, and refused where the spell has no curve at the position or the row's base is 0.
type CurveSign struct {
	SpellID int32
	Effect  int // the effect position as the tooltip counts it, from 1: $s2 is 2
	Reason  string
	Source  string
}

var CurveSigns = []CurveSign{
	{
		SpellID: 11160,
		Effect:  2,
		Reason:  "Frost Channeling: row -30, tooltip 'reduces the threat caused by your Frost spells by $s2%'",
		Source:  "tooltip, game build 1.60.1.70170",
	},
	{
		SpellID: 1311085,
		Effect:  2,
		Reason:  "Instrument of Law: row -20, tooltip 'reduces all threat you generate by $m2% while Righteous Fury is not active'",
		Source:  "tooltip, game build 1.60.1.70170",
	},
	{
		SpellID: 12962,
		Effect:  1,
		Reason:  "Iron Will: row -15, tooltip 'Reduces the duration of Stun and Fear effects inflicted on you by $s1%'",
		Source:  "tooltip, game build 1.60.1.70170",
	},
	{
		SpellID: 19572,
		Effect:  2,
		Reason:  "Improved Mend Pet: row -20, tooltip 'reduces the Mana cost by $s2%'",
		Source:  "tooltip, game build 1.60.1.70170",
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
	{
		SpellID: 350026,
		Field:   ICDMs,
		Value:   10000,
		Reason:  "Shadow Oil: no proc on 425 landed hits within 10 s of the last one, 80 on 633 after; procs at least 10.26 s apart, none on hits from 9.98 s",
		Source:  "wcl:Y4mTwRvPtqAn36CB/1-9, 2026-10-07",
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

	// Revenge's Attack Power shares, read off how its non-crit hits move with Attack Power inside each
	// log, with Defensive Stance's 10% and Improved Revenge taken out. Only Ranks 1 and 2 were measured;
	// the other ranks carry their share until a log at their level says otherwise.
	{
		SpellID: 6572,
		Field:   APCoefDirect,
		Value:   0.28,
		Reason:  "Revenge Rank 1: hits deal 22 +/- 15% + 0.28 x Attack Power, before armor",
		Source:  "combat logs, Revenge Rank 1, 2026-09-25 and 2026-09-29",
	},
	{
		SpellID: 6574,
		Field:   APCoefDirect,
		Value:   0.28,
		Reason:  "Revenge Rank 2: 0.24 to 0.30 x Attack Power with Improved Revenge 3/3 taken out, Rank 1's 0.28 within it",
		Source:  "combat log, Revenge Rank 2, 2026-10-03",
	},
	{
		SpellID: 7379,
		Field:   APCoefDirect,
		Value:   0.28, // TODO: Measure Revenge Rank 3 in game.
		Reason:  "Revenge Rank 3: Ranks 1 and 2's measured share",
		Source:  "combat logs, Revenge Ranks 1 and 2, 2026-09-25 to 2026-10-03",
	},
	{
		SpellID: 11600,
		Field:   APCoefDirect,
		Value:   0.28, // TODO: Measure Revenge Rank 4 in game.
		Reason:  "Revenge Rank 4: Ranks 1 and 2's measured share",
		Source:  "combat logs, Revenge Ranks 1 and 2, 2026-09-25 to 2026-10-03",
	},
	{
		SpellID: 11601,
		Field:   APCoefDirect,
		Value:   0.28, // TODO: Measure Revenge Rank 5 in game.
		Reason:  "Revenge Rank 5: Ranks 1 and 2's measured share",
		Source:  "combat logs, Revenge Ranks 1 and 2, 2026-09-25 to 2026-10-03",
	},
	{
		SpellID: 25288,
		Field:   APCoefDirect,
		Value:   0.28, // TODO: Measure Revenge Rank 6 in game.
		Reason:  "Revenge Rank 6: Ranks 1 and 2's measured share",
		Source:  "combat logs, Revenge Ranks 1 and 2, 2026-09-25 to 2026-10-03",
	},

	// Sunder Armor's threat grows with the sheet's Attack Power: measured on Rank 1 at 132, 158, 170 and
	// 180 Attack Power, in Battle and Defensive Stance. The other ranks carry its share until a capture
	// at their level says otherwise.
	{
		SpellID: 7386,
		Field:   APCoefThreat,
		Value:   0.05,
		Reason:  "Sunder Armor Rank 1: threat is 34 + 0.05 x Attack Power, times the stance, rounded down",
		Source:  "threat capture, Sunder Armor Rank 1, 2026-10-07",
	},
	{
		SpellID: 7405,
		Field:   APCoefThreat,
		Value:   0.05, // TODO: Measure Sunder Armor Rank 2 in game.
		Reason:  "Sunder Armor Rank 2: Rank 1's measured share",
		Source:  "threat capture, Sunder Armor Rank 1, 2026-10-07",
	},
	{
		SpellID: 8380,
		Field:   APCoefThreat,
		Value:   0.05, // TODO: Measure Sunder Armor Rank 3 in game.
		Reason:  "Sunder Armor Rank 3: Rank 1's measured share",
		Source:  "threat capture, Sunder Armor Rank 1, 2026-10-07",
	},
	{
		SpellID: 11596,
		Field:   APCoefThreat,
		Value:   0.05, // TODO: Measure Sunder Armor Rank 4 in game.
		Reason:  "Sunder Armor Rank 4: Rank 1's measured share",
		Source:  "threat capture, Sunder Armor Rank 1, 2026-10-07",
	},
	{
		SpellID: 11597,
		Field:   APCoefThreat,
		Value:   0.05, // TODO: Measure Sunder Armor Rank 5 in game.
		Reason:  "Sunder Armor Rank 5: Rank 1's measured share",
		Source:  "threat capture, Sunder Armor Rank 1, 2026-10-07",
	},

	// Heroic Strike and Cleave add threat beyond their damage in proportion to the rank's spell level,
	// as Sunder Armor's game files threat does (34 to 206, 3.4 to 3.55 per level), and none with
	// Attack Power: Heroic Strike Rank 3 held its bonus from 132 to 170 Attack Power. The bonus is
	// multiplied by the stance with the damage, and Cleave splits it between the targets it hits.
	// Ranks not measured carry the measured share.
	{
		SpellID: 78,
		Field:   ThreatPerSpellLevel,
		Value:   2.54, // TODO: Measure Heroic Strike Rank 1 in game.
		Reason:  "Heroic Strike Rank 1: Ranks 3 and 4's measured share",
		Source:  "threat capture, Heroic Strike Ranks 3 and 4, 2026-10-07",
	},
	{
		SpellID: 284,
		Field:   ThreatPerSpellLevel,
		Value:   2.54, // TODO: Measure Heroic Strike Rank 2 in game.
		Reason:  "Heroic Strike Rank 2: Ranks 3 and 4's measured share",
		Source:  "threat capture, Heroic Strike Ranks 3 and 4, 2026-10-07",
	},
	{
		SpellID: 285,
		Field:   ThreatPerSpellLevel,
		Value:   2.54,
		Reason:  "Heroic Strike Rank 3: threat is the damage + 40 to 40.6 at spell level 16, times the stance",
		Source:  "threat capture, Heroic Strike Rank 3, 2026-10-07",
	},
	{
		SpellID: 1608,
		Field:   ThreatPerSpellLevel,
		Value:   2.54,
		Reason:  "Heroic Strike Rank 4: threat is the damage + 61 at spell level 24, times the stance",
		Source:  "threat capture, Heroic Strike Rank 4, 2026-10-07",
	},
	{
		SpellID: 11564,
		Field:   ThreatPerSpellLevel,
		Value:   2.54, // TODO: Measure Heroic Strike Rank 5 in game.
		Reason:  "Heroic Strike Rank 5: Ranks 3 and 4's measured share",
		Source:  "threat capture, Heroic Strike Ranks 3 and 4, 2026-10-07",
	},
	{
		SpellID: 11565,
		Field:   ThreatPerSpellLevel,
		Value:   2.54, // TODO: Measure Heroic Strike Rank 6 in game.
		Reason:  "Heroic Strike Rank 6: Ranks 3 and 4's measured share",
		Source:  "threat capture, Heroic Strike Ranks 3 and 4, 2026-10-07",
	},
	{
		SpellID: 11566,
		Field:   ThreatPerSpellLevel,
		Value:   2.54, // TODO: Measure Heroic Strike Rank 7 in game.
		Reason:  "Heroic Strike Rank 7: Ranks 3 and 4's measured share",
		Source:  "threat capture, Heroic Strike Ranks 3 and 4, 2026-10-07",
	},
	{
		SpellID: 11567,
		Field:   ThreatPerSpellLevel,
		Value:   2.54, // TODO: Measure Heroic Strike Rank 8 in game.
		Reason:  "Heroic Strike Rank 8: Ranks 3 and 4's measured share",
		Source:  "threat capture, Heroic Strike Ranks 3 and 4, 2026-10-07",
	},
	{
		SpellID: 25286,
		Field:   ThreatPerSpellLevel,
		Value:   2.54, // TODO: Measure Heroic Strike Rank 9 in game.
		Reason:  "Heroic Strike Rank 9: Ranks 3 and 4's measured share",
		Source:  "threat capture, Heroic Strike Ranks 3 and 4, 2026-10-07",
	},
	{
		SpellID: 845,
		Field:   ThreatPerSpellLevel,
		Value:   1.53, // TODO: Measure Cleave Rank 1 in game.
		Reason:  "Cleave Rank 1: Rank 2's measured share",
		Source:  "threat capture, Cleave Rank 2, 2026-10-07",
	},
	{
		SpellID: 7369,
		Field:   ThreatPerSpellLevel,
		Value:   1.53,
		Reason:  "Cleave Rank 2: adds ~46 beyond its damage at spell level 30 on one target, ~23 on each of two, times the stance",
		Source:  "threat capture, Cleave Rank 2, 2026-10-07",
	},
	{
		SpellID: 11608,
		Field:   ThreatPerSpellLevel,
		Value:   1.53, // TODO: Measure Cleave Rank 3 in game.
		Reason:  "Cleave Rank 3: Rank 2's measured share",
		Source:  "threat capture, Cleave Rank 2, 2026-10-07",
	},
	{
		SpellID: 11609,
		Field:   ThreatPerSpellLevel,
		Value:   1.53, // TODO: Measure Cleave Rank 4 in game.
		Reason:  "Cleave Rank 4: Rank 2's measured share",
		Source:  "threat capture, Cleave Rank 2, 2026-10-07",
	},
	{
		SpellID: 20569,
		Field:   ThreatPerSpellLevel,
		Value:   1.53, // TODO: Measure Cleave Rank 5 in game.
		Reason:  "Cleave Rank 5: Rank 2's measured share",
		Source:  "threat capture, Cleave Rank 2, 2026-10-07",
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
