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

	// The spell-power coefficient of the direct effect, onto Effect.SPCoef. Fails when the game files
	// state an EffectBonusCoefficient there.
	SPCoefDirect

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
	case SPCoefDirect:
		return "SPCoefDirect"
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
	Source  string // "tbc-carryover", "tooltip", "wcl:<report>/<fight>", "issue #N", "combat log, <spell and rank>, <date>"
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
// on every hit", or the spell has no proc row at all. Each row's Source says where its number came
// from.
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
	{
		SpellID: 16353,
		Field:   PPM,
		Value:   9, // TODO: Vanilla value - verify in-game
		Reason:  "Frostbrand Weapon: enchant 1668 casts 16353 with 0 in its amount, and 16353 has no proc row",
		Source:  "vanilla-carryover",
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
		SpellID: 350026,
		Field:   ICDMs,
		Value:   10000,
		Reason:  "Shadow Oil: no proc on 425 landed hits within 10 s of the last one, 80 on 633 after; procs at least 10.26 s apart, none on hits from 9.98 s",
		Source:  "wcl:Y4mTwRvPtqAn36CB/1-9, 2026-10-07",
	},

	// A spell-power coefficient the damage effect does not state.
	{
		SpellID: 1240047,
		Field:   SPCoefDirect,
		Value:   1.0 / 7, // TODO: Measure Ice Lance's spell power coefficient in game.
		Reason:  "Ice Lance: the damage effect states no coefficient, but the spell scales with spell power in game; TBC's 1/7 until measured",
		Source:  "tbc-carryover",
	},

	// Seal of Light and Judgement of Light, measured in game at level 30 on Rank 1 across two combat
	// logs, one with a 1.6 s weapon and one with a 3.7 s one. The seal healed on 463 of 1101 and 509
	// of 554 landed melee hits while it was up, which only a per-minute rate fits, and the judgement
	// on 825 of the 1693 landed while it was up, 534 of 1090 at 1.6 s and 256 of 537 at 3.7 s, which
	// is a flat chance. Neither fired on a dodge, a parry or a miss. The other ranks carry Rank 1's
	// numbers until a log at their level says otherwise, and Judgement of Wisdom carries Judgement of
	// Light's chance until a log measures it.
	{
		SpellID: 20165,
		Field:   PPM,
		Value:   15,
		Reason:  "Seal of Light Rank 1: 463 of 1101 landed hits at 1.6 s and 509 of 554 at 3.7 s; 15 per minute fits both, 14 fits neither and 16 fails the 3.7 s log",
		Source:  "combat log, Seal of Light Rank 1, 2026-10-14 and 2026-10-15",
	},
	{
		SpellID: 20347,
		Field:   PPM,
		Value:   15, // TODO: Measure Seal of Light Rank 2 in game.
		Reason:  "Seal of Light Rank 2: Rank 1's measured rate",
		Source:  "combat log, Seal of Light Rank 1, 2026-10-14 and 2026-10-15",
	},
	{
		SpellID: 20348,
		Field:   PPM,
		Value:   15, // TODO: Measure Seal of Light Rank 3 in game.
		Reason:  "Seal of Light Rank 3: Rank 1's measured rate",
		Source:  "combat log, Seal of Light Rank 1, 2026-10-14 and 2026-10-15",
	},
	{
		SpellID: 20349,
		Field:   PPM,
		Value:   15, // TODO: Measure Seal of Light Rank 4 in game.
		Reason:  "Seal of Light Rank 4: Rank 1's measured rate",
		Source:  "combat log, Seal of Light Rank 1, 2026-10-14 and 2026-10-15",
	},
	// Seal of Wisdom carries Seal of Light's measured rate until a log measures it. Every source that
	// gives the two seals a rate gives them the same one.
	{
		SpellID: 20166,
		Field:   PPM,
		Value:   15, // TODO: Measure Seal of Wisdom Rank 1 in game.
		Reason:  "Seal of Wisdom Rank 1: Seal of Light Rank 1's measured rate",
		Source:  "estimate",
	},
	{
		SpellID: 20356,
		Field:   PPM,
		Value:   15, // TODO: Measure Seal of Wisdom Rank 2 in game.
		Reason:  "Seal of Wisdom Rank 2: Seal of Light Rank 1's measured rate",
		Source:  "estimate",
	},
	{
		SpellID: 20357,
		Field:   PPM,
		Value:   15, // TODO: Measure Seal of Wisdom Rank 3 in game.
		Reason:  "Seal of Wisdom Rank 3: Seal of Light Rank 1's measured rate",
		Source:  "estimate",
	},
	{
		SpellID: 20185,
		Field:   ProcChancePct,
		Value:   50,
		Reason:  "Judgement of Light Rank 1: 825 of 1693 landed hits, 534 of 1090 at 1.6 s and 256 of 537 at 3.7 s",
		Source:  "combat log, Judgement of Light Rank 1, 2026-10-14 and 2026-10-15",
	},
	{
		SpellID: 20344,
		Field:   ProcChancePct,
		Value:   50, // TODO: Measure Judgement of Light Rank 2 in game.
		Reason:  "Judgement of Light Rank 2: Rank 1's measured chance",
		Source:  "combat log, Judgement of Light Rank 1, 2026-10-14 and 2026-10-15",
	},
	{
		SpellID: 20345,
		Field:   ProcChancePct,
		Value:   50, // TODO: Measure Judgement of Light Rank 3 in game.
		Reason:  "Judgement of Light Rank 3: Rank 1's measured chance",
		Source:  "combat log, Judgement of Light Rank 1, 2026-10-14 and 2026-10-15",
	},
	{
		SpellID: 20346,
		Field:   ProcChancePct,
		Value:   50, // TODO: Measure Judgement of Light Rank 4 in game.
		Reason:  "Judgement of Light Rank 4: Rank 1's measured chance",
		Source:  "combat log, Judgement of Light Rank 1, 2026-10-14 and 2026-10-15",
	},
	{
		SpellID: 20186,
		Field:   ProcChancePct,
		Value:   50, // TODO: Measure Judgement of Wisdom Rank 1 in game.
		Reason:  "Judgement of Wisdom Rank 1: Judgement of Light Rank 1's measured chance",
		Source:  "estimate",
	},
	{
		SpellID: 20354,
		Field:   ProcChancePct,
		Value:   50, // TODO: Measure Judgement of Wisdom Rank 2 in game.
		Reason:  "Judgement of Wisdom Rank 2: Judgement of Light Rank 1's measured chance",
		Source:  "estimate",
	},
	{
		SpellID: 20355,
		Field:   ProcChancePct,
		Value:   50, // TODO: Measure Judgement of Wisdom Rank 3 in game.
		Reason:  "Judgement of Wisdom Rank 3: Judgement of Light Rank 1's measured chance",
		Source:  "estimate",
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
	{
		SpellID: 16864,
		Field:   PPM,
		Value:   2,
		Reason:  "Omen of Clarity: 2.0 procs per minute, measured in combat logs; the game files state none",
		Source:  "issue #225",
	},
	{
		SpellID: 9896,
		Field:   APCoefPeriodic,
		Value:   0.01,
		Reason:  "Rip: 0.06 attack power per combo point over six ticks, capped at four points; higher-rank verification remains open",
		Source:  "Rip AP Scaling Testing, 2026-09-26, https://docs.google.com/spreadsheets/d/1fQiaAQzz1r_wS-2pD5e1Y6wZLxajru2R8GK5Uw1Lo60",
	},
	{
		SpellID: 1822,
		Field:   APCoefPeriodic,
		Value:   0.055,
		Reason:  "Rake Rank 1: periodic attack power share fitted to 136, 162, 217 and 237 AP with reported damage modifiers",
		Source:  "Aus and Chalkie, Rake and swipe AP scaling tests, 2026-10-01; PR #149 evidence notes",
	},
	{
		SpellID: 1823,
		Field:   APCoefPeriodic,
		Value:   0.055,
		Reason:  "Rake Rank 2: Rank 1's measured periodic share; TODO: verify at this rank",
		Source:  "Aus and Chalkie, Rake and swipe AP scaling tests, 2026-10-01; PR #149 evidence notes",
	},
	{
		SpellID: 1824,
		Field:   APCoefPeriodic,
		Value:   0.055,
		Reason:  "Rake Rank 3: Rank 1's measured periodic share; TODO: verify at this rank",
		Source:  "Aus and Chalkie, Rake and swipe AP scaling tests, 2026-10-01; PR #149 evidence notes",
	},
	{
		SpellID: 9904,
		Field:   APCoefPeriodic,
		Value:   0.055,
		Reason:  "Rake Rank 4: Rank 1's measured periodic share; TODO: verify at this rank",
		Source:  "Aus and Chalkie, Rake and swipe AP scaling tests, 2026-10-01; PR #149 evidence notes",
	},
	{
		SpellID: 9908,
		Field:   APCoefDirect,
		Value:   0.03,
		Reason:  "Swipe Rank 5: 3% of attack power on each hit; the game files state none",
		Source:  "Aus and Chalkie, Rake and swipe AP scaling tests, 2026-10-01; PR #149 evidence notes",
	},
	{
		SpellID: 22829,
		Field:   APCoefDirect,
		Value:   0.03,
		Reason:  "Placeholder: Ferocious Bite attack power per combo point; the game files state none",
		Source:  "classic-carryover",
	},

	// Revenge's Attack Power share, the $APR its tooltip states: 0.225 in Defensive Stance is the stance's
	// own 10%, which stays out of these values. Combat logs on Ranks 1 and 2 taken before that tooltip fit 0.28.
	{
		SpellID: 6572,
		Field:   APCoefDirect,
		Value:   0.25, // TODO: Measure Revenge Rank 1 in game.
		Reason:  "Revenge Rank 1: the tooltip's $APR outside Defensive Stance",
		Source:  "tooltip",
	},
	{
		SpellID: 6574,
		Field:   APCoefDirect,
		Value:   0.25, // TODO: Measure Revenge Rank 2 in game.
		Reason:  "Revenge Rank 2: the tooltip's $APR outside Defensive Stance",
		Source:  "tooltip",
	},
	{
		SpellID: 7379,
		Field:   APCoefDirect,
		Value:   0.25, // TODO: Measure Revenge Rank 3 in game.
		Reason:  "Revenge Rank 3: the tooltip's $APR outside Defensive Stance",
		Source:  "tooltip",
	},
	{
		SpellID: 11600,
		Field:   APCoefDirect,
		Value:   0.25, // TODO: Measure Revenge Rank 4 in game.
		Reason:  "Revenge Rank 4: the tooltip's $APR outside Defensive Stance",
		Source:  "tooltip",
	},
	{
		SpellID: 11601,
		Field:   APCoefDirect,
		Value:   0.25, // TODO: Measure Revenge Rank 5 in game.
		Reason:  "Revenge Rank 5: the tooltip's $APR outside Defensive Stance",
		Source:  "tooltip",
	},
	{
		SpellID: 25288,
		Field:   APCoefDirect,
		Value:   0.25, // TODO: Measure Revenge Rank 6 in game.
		Reason:  "Revenge Rank 6: the tooltip's $APR outside Defensive Stance",
		Source:  "tooltip",
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
