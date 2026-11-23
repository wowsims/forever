// Package overrides holds the numbers the spelldata store carries that the game files do
// not state. Each row names the spell, the field, the value and where the value comes from, and the
// generator refuses to run when the game files start stating one themselves - a value kept by hand
// after the data grew a column for it is a wrong number nobody would notice.
//
// A leaf package on purpose: tools/database imports it, it imports nothing back, and the folder it
// shares with the SQL overrides is read for *.sql alone.
package overrides

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

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

	// A flat threat bonus of Value times the spell's level, onto Spell.FlatThreat, for a rank that
	// carries a measured rank's share per spell level. Fails like FlatThreat, and beside a FlatThreat
	// override on the same spell.
	ThreatPerSpellLevel

	// The internal cooldown between two procs in milliseconds, onto Spell.ICDMs. Fails when the game
	// files state a SpellAuraOptions.ProcCategoryRecovery.
	ICDMs

	// The first word of the proc effect's class mask, onto its Effect.ClassFlags, for a mask the game
	// files state under a bit no spell of the family is filed under. Fails when the effect states
	// the same word itself.
	ProcClassMask
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
	case ProcClassMask:
		return "ProcClassMask"
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
	Source  string // "EV-<n>" (docs/evidence), "tbc-carryover", "vanilla-carryover", "observed", "tooltip", "owner decision", "estimate"
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
	{
		SpellID:            1321558,
		Groups:             []int32{9163, 9202},
		Multiplier:         2,
		DurationMultiplier: 1,
		Reason:             "Siphon the Grave (Rotmender's Raiment 4pc): 'Restores twice as much Mana in Haunted and Wasteland areas'; no companion row",
		Source:             "tooltip",
	},
}

// A factor a row deals, restores or procs against some targets, which only its tooltip names ("Deals
// $s2 times as much damage to Plants", "Attacks against Dwarves are $s2 times as likely to activate
// this effect"). A target of any of the mob types or families counts.
type VersusBonus struct {
	SpellID     int32
	EffectIndex uint8   // The effect whose base points are the factor: the tooltip's $s<index+1>.
	Multiplier  float64 // A factor no effect states; 0 reads EffectIndex's base points.
	MobTypes    []proto.MobType
	Families    []proto.CreatureFamily
	Reason      string
	Source      string
}

// No CreatureFamily holds Golems or Titan-forged, which the tooltips of 12731, 1302279 and 1294794
// name, so those rows count Earth Elementals alone.
var VersusBonuses = []VersusBonus{
	{
		SpellID:     1293674,
		EffectIndex: 0,
		MobTypes:    []proto.MobType{proto.MobType_MobTypeUndead},
		Reason:      "Enraging Light (Chain of the Scarlet Crusade 4pc): 'Deals $s1 times as much damage to Undead'",
		Source:      "tooltip",
	},
	{
		SpellID:     12731,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyEarthElemental},
		Reason:      "Strength of Stone: 'Deals $s2 times as much damage to Earth Elementals, Golems, and Titan-forged'",
		Source:      "tooltip",
	},
	{
		SpellID:     1302279,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyEarthElemental},
		Reason:      "Stoneshatter: 'Deals $s2 times as much damage to Earth Elementals, Golems, and Titan-forged'",
		Source:      "tooltip",
	},
	{
		SpellID:     1294794,
		EffectIndex: 0,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyEarthElemental},
		Reason:      "Fist of Stone: 'Restores $s1 times as much if attacking an Earth Elemental, Golem, or Titan-forged enemy'",
		Source:      "tooltip",
	},
	{
		SpellID:     1293568,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyPlant},
		Reason:      "Grass Killer: 'Deals $s2 times as much damage to Plants'",
		Source:      "tooltip",
	},
	{
		SpellID:     12257,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyPlant},
		Reason:      "Breath of Fire: 'Deals $s2 times as much damage to Plants'",
		Source:      "tooltip",
	},
	{
		SpellID:     17283,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyPlant},
		Reason:      "Smokey's Lighter: 'Deals $s2 times as much damage to Plants'",
		Source:      "tooltip",
	},
	{
		SpellID:     1295270,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyPlant},
		Reason:      "Prototype Pathcarver: 'Deals $s2 times as much damage to Plants'",
		Source:      "tooltip",
	},
	{
		SpellID:     1295039,
		EffectIndex: 2,
		MobTypes:    []proto.MobType{proto.MobType_MobTypeBeast},
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilySerpentine},
		Reason:      "Serpentbane: 'Inflicts $s3 times as much damage to Beast and Serpentine targets'",
		Source:      "tooltip",
	},
	{
		SpellID:     1318159,
		EffectIndex: 0,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyNaga, proto.CreatureFamily_CreatureFamilySatyr},
		Reason:      "Ancient Power: 'Deals $s1 times as much damage to Naga and Satyrs'",
		Source:      "tooltip",
	},
	{
		SpellID:     1311505,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyDinosaur},
		Reason:      "Grounded Electroshocks: 'Deals $s2 times as much damage to Dinosaurs'",
		Source:      "tooltip",
	},
	{
		SpellID:     14118,
		EffectIndex: 1,
		MobTypes:    []proto.MobType{proto.MobType_MobTypeDragonkin},
		Reason:      "Rend: 'Deals $s2 times as much damage to Dragonkin'",
		Source:      "tooltip",
	},
	{
		SpellID:     1299631,
		EffectIndex: 1,
		MobTypes:    []proto.MobType{proto.MobType_MobTypeUndead},
		Reason:      "Smite: 'Deals $s2 times as much damage to Undead'",
		Source:      "tooltip",
	},
	{
		SpellID:     16414,
		EffectIndex: 1,
		MobTypes:    []proto.MobType{proto.MobType_MobTypeUndead},
		Reason:      "Drain Life: 'Deals $s2 times as much damage to Undead'",
		Source:      "tooltip",
	},
	{
		SpellID:    1325229,
		Multiplier: 3,
		MobTypes:   []proto.MobType{proto.MobType_MobTypeUndead},
		Reason:     "Gloves of Holy Might: 'Damage is tripled against Undead'; neither 1325229 nor 1330855 states the 3",
		Source:     "tooltip",
	},
	{
		SpellID:     27656,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyOrc},
		Reason:      "Flame Lash: 'Deals 3 times as much damage to Orcs'",
		Source:      "tooltip",
	},
	{
		SpellID:     1295744,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyAquatic},
		Reason:      "Rot: 'Deals $s2 times as much damage to Aquatic enemies'",
		Source:      "tooltip",
	},
	{
		SpellID:     3742,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyAquatic},
		Reason:      "Gahz'rilla Fang: 'Deals $s2 times as much damage to Aquatic enemies'",
		Source:      "tooltip",
	},
	{
		SpellID:     23686,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyAquatic},
		Reason:      "Darkmoon Card: Maelstrom: 'Deals $s2 times as much damage to Aquatic enemies'",
		Source:      "tooltip",
	},
	{
		SpellID:     1294430,
		EffectIndex: 4,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyTroll},
		Reason:      "Rage of Sul'thraze: 'This additional Shadow damage deals $s5 times as much damage to Trolls'; Shadow Lash 1294431 deals it",
		Source:      "tooltip",
	},
	{
		SpellID:     1294339,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyTroll},
		Reason:      "Farraki Enforcement: 'Trolls take $s2 times as much damage and stun duration'",
		Source:      "tooltip",
	},
	{
		SpellID:     1300783,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyTroll, proto.CreatureFamily_CreatureFamilyOrc},
		Reason:      "Slaughter: 'Trolls and Orcs take $s2 times as much damage'",
		Source:      "tooltip",
	},
	{
		SpellID:     1317432,
		EffectIndex: 1,
		Families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyEarthElemental,
			proto.CreatureFamily_CreatureFamilyFireElemental, proto.CreatureFamily_CreatureFamilyMountainGiant},
		Reason: "Forging Strike: 'Deals $s2 times as much damage to Earth Elementals, Fire Elementals, and Mountain Giants'",
		Source: "tooltip",
	},
	{
		SpellID:     15600,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyDwarf},
		Reason:      "Hand of Justice: 'Attacks against Dwarves are $s2 times as likely to activate this effect'",
		Source:      "tooltip",
	},
	{
		SpellID:     1301046,
		EffectIndex: 1,
		Families:    []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyOrc},
		Reason:      "Fury of Forgewright: 'Attacks against Orcs are $s2 times as likely to activate this effect'",
		Source:      "tooltip",
	},
	{
		SpellID:     17350,
		EffectIndex: 1,
		MobTypes:    []proto.MobType{proto.MobType_MobTypeUndead},
		Reason:      "Argent Defender: 'Undead that strike you are $s2 times as likely to activate this effect'",
		Source:      "tooltip",
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

	// A relic's lock stated under a bit no spell of its class is filed under.
	{
		SpellID: 28849,
		Field:   ProcClassMask,
		Value:   128,
		Reason:  "Totem of Flowing Water: the dummy's mask is bit 5 of the shaman family, under which no spell is filed; Lesser Healing Wave, which the tooltip names, is bit 7",
		Source:  "tooltip",
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
		Reason:  "Shadow Oil: no proc on a hit within 10 s of the last one",
		Source:  "EV-0016",
	},

	// A spell-power coefficient the damage effect does not state.
	{
		SpellID: 1240047,
		Field:   SPCoefDirect,
		Value:   0.11,
		Reason:  "Ice Lance: the damage effect states no coefficient, but the spell scales with spell power in game (EV-0171)",
		Source:  "EV-0171",
	},

	// Seal of Light procs 15 a minute and Judgement of Light rolls a flat 50% (EV-0017, EV-0018). Neither
	// fires on a dodge, a parry or a miss. The ranks no log has measured carry the nearest measured rank's
	// numbers until one at their level says otherwise.
	{
		SpellID: 20165,
		Field:   PPM,
		Value:   15,
		Reason:  "Seal of Light Rank 1: a per-minute rate fits both weapon speeds",
		Source:  "EV-0017",
	},
	{
		SpellID: 20347,
		Field:   PPM,
		Value:   15,
		Reason:  "Seal of Light Rank 2: the measured rate, not a per-swing share of 15 per minute",
		Source:  "EV-0018",
	},
	{
		SpellID: 20348,
		Field:   PPM,
		Value:   15, // TODO: Measure Seal of Light Rank 3 in game.
		Reason:  "Seal of Light Rank 3: Ranks 1 and 2's measured rate",
		Source:  "EV-0017, EV-0018",
	},
	{
		SpellID: 20349,
		Field:   PPM,
		Value:   15, // TODO: Measure Seal of Light Rank 4 in game.
		Reason:  "Seal of Light Rank 4: Ranks 1 and 2's measured rate",
		Source:  "EV-0017, EV-0018",
	},
	// Seal of Wisdom Rank 1 measured the same rate as Seal of Light (EV-0018), and its other ranks carry it
	// until a log at their level says otherwise.
	{
		SpellID: 20166,
		Field:   PPM,
		Value:   15,
		Reason:  "Seal of Wisdom Rank 1: the measured rate, not a per-swing share of 15 per minute",
		Source:  "EV-0018",
	},
	{
		SpellID: 20356,
		Field:   PPM,
		Value:   15, // TODO: Measure Seal of Wisdom Rank 2 in game.
		Reason:  "Seal of Wisdom Rank 2: Rank 1's measured rate",
		Source:  "EV-0018",
	},
	{
		SpellID: 20357,
		Field:   PPM,
		Value:   15, // TODO: Measure Seal of Wisdom Rank 3 in game.
		Reason:  "Seal of Wisdom Rank 3: Rank 1's measured rate",
		Source:  "EV-0018",
	},
	{
		SpellID: 20185,
		Field:   ProcChancePct,
		Value:   50,
		Reason:  "Judgement of Light Rank 1: a flat chance on every landed hit, whatever the weapon speed",
		Source:  "EV-0017",
	},
	{
		SpellID: 20344,
		Field:   ProcChancePct,
		Value:   50,
		Reason:  "Judgement of Light Rank 2: the measured chance on landed melee swings; ranged and spell hits never heal",
		Source:  "EV-0018",
	},
	{
		SpellID: 20345,
		Field:   ProcChancePct,
		Value:   50, // TODO: Measure Judgement of Light Rank 3 in game.
		Reason:  "Judgement of Light Rank 3: Ranks 1 and 2's measured chance",
		Source:  "EV-0017, EV-0018",
	},
	{
		SpellID: 20346,
		Field:   ProcChancePct,
		Value:   50, // TODO: Measure Judgement of Light Rank 4 in game.
		Reason:  "Judgement of Light Rank 4: Ranks 1 and 2's measured chance",
		Source:  "EV-0017, EV-0018",
	},
	{
		SpellID: 20186,
		Field:   ProcChancePct,
		Value:   50,
		Reason:  "Judgement of Wisdom Rank 1: the measured chance on landed melee swings, direct spells and ranged shots",
		Source:  "EV-0018",
	},
	{
		SpellID: 20354,
		Field:   ProcChancePct,
		Value:   50, // TODO: Measure Judgement of Wisdom Rank 2 in game.
		Reason:  "Judgement of Wisdom Rank 2: Rank 1's measured chance",
		Source:  "EV-0018",
	},
	{
		SpellID: 20355,
		Field:   ProcChancePct,
		Value:   50, // TODO: Measure Judgement of Wisdom Rank 3 in game.
		Reason:  "Judgement of Wisdom Rank 3: Rank 1's measured chance",
		Source:  "EV-0018",
	},
	{
		SpellID: 1318931,
		Field:   ProcChancePct,
		Value:   8,
		Reason:  "Draconic Infused Emblem: 8% on each landed harmful spell hit; column reads 100, tooltip says 'Chance on harmful spellcast'",
		Source:  "EV-0115",
	},
	{
		SpellID: 1294668,
		Field:   ProcChancePct,
		Value:   5,
		Reason:  "Zealot Plate: 5% on each landed melee swing taken outside the 15 s cooldown; column reads 100, tooltip says 'When struck in combat, has a chance'",
		Source:  "EV-0115",
	},
	// Fitted from logs on landed melee hits taken that deal damage; the owner rules 0.5 from the pooled
	// fit (EV-0126). The "observed" Source keeps the gear tab's notice up until more wearers validate it.
	{
		SpellID: 17350,
		Field:   PPM,
		Value:   0.5,
		Reason:  "Argent Defender: the owner's 0.5 procs per minute on the attacker's swing interval, twice that when an Undead strikes, outside the 20 s cooldown; column reads 100, tooltip says 'Has a chance when struck'; a log fit not yet validated (EV-0126)",
		Source:  "observed",
	},

	// Attack Power shares measured against the sheet's Attack Power. Rend Ranks 2, 4 and 5 and Thunder Clap
	// Ranks 1 and 4 were measured; the other ranks carry the same share until a log at their level says
	// otherwise (EV-0019, EV-0048).
	{
		SpellID: 772,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 1: Rank 2's measured share per tick",
		Source:  "EV-0019",
	},
	{
		SpellID: 6546,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 2: ticks average 7 + 0.02 x Attack Power, read on every tick",
		Source:  "EV-0019",
	},
	{
		SpellID: 6547,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 3: Rank 2's measured share per tick",
		Source:  "EV-0019",
	},
	{
		SpellID: 6548,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 4: ticks take 0.02 x Attack Power, read on every tick",
		Source:  "EV-0048",
	},
	{
		SpellID: 11572,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 5: ticks take 0.02 x Attack Power, read on every tick",
		Source:  "EV-0048",
	},
	{
		SpellID: 11573,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 6: Rank 2's measured share per tick",
		Source:  "EV-0019",
	},
	{
		SpellID: 11574,
		Field:   APCoefPeriodic,
		Value:   0.02,
		Reason:  "Rend Rank 7: Rank 2's measured share per tick",
		Source:  "EV-0019",
	},
	{
		SpellID: 6343,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 1: hits deal 10 + 0.07 x Attack Power, rounded down, before armor",
		Source:  "EV-0019",
	},
	{
		SpellID: 8198,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 2: Rank 1's measured share",
		Source:  "EV-0019",
	},
	{
		SpellID: 8204,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 3: Rank 1's measured share",
		Source:  "EV-0019",
	},
	{
		SpellID: 8205,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 4: hits deal 0.07 x Attack Power, measured",
		Source:  "EV-0048",
	},
	{
		SpellID: 11580,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 5: Rank 1's measured share",
		Source:  "EV-0019",
	},
	{
		SpellID: 11581,
		Field:   APCoefDirect,
		Value:   0.07,
		Reason:  "Thunder Clap Rank 6: Rank 1's measured share",
		Source:  "EV-0019",
	},
	{
		SpellID: 16864,
		Field:   PPM,
		Value:   2,
		Reason:  "Omen of Clarity: 2.0 procs per minute, measured; the game files state none",
		Source:  "EV-0020",
	},
	{
		SpellID: 1318159,
		Field:   PPM,
		Value:   2.95,
		Reason:  "Highborne Research Tablet: 2.95 procs per minute on landed damaging hits, a weapon hit on its weapon's speed and a spell hit on its cast time; the game files state none",
		Source:  "EV-0131",
	},
	{
		SpellID: 27656,
		Field:   PPM,
		Value:   1.95,
		Reason:  "Heart of Wyrmthalak: 1.95 procs per minute on landed melee and ranged hits, on the attacking weapon's speed; the game files state none",
		Source:  "EV-0133",
	},
	{
		SpellID: 15594,
		Field:   ProcChancePct,
		Value:   6, // TODO: Find what moves the chance between attackers (EV-0154).
		Reason:  "Force of Will: 6% on each melee hit taken that deals damage, the measured chance to the whole percent; the game files state none",
		Source:  "EV-0154",
	},
	{
		SpellID: 1318323,
		Field:   PPM,
		Value:   1.2,
		Reason:  "Scented Runewood Brooch: 1.2 procs per minute on landed hits against Furbolg, a log fit not yet validated (EV-0136); the game files state none",
		Source:  "observed",
	},
	{
		SpellID: 1309550,
		Field:   PPM,
		Value:   1.35,
		Reason:  "Claymore of Two Moons: 1.35 procs per minute on landed melee hits, read off paladins alone (EV-0163); the game files state none",
		Source:  "observed",
	},
	{
		SpellID: 450587,
		Field:   PPM,
		Value:   1,
		Reason:  "Battlegear of Valor 5-piece, Battlegear of Heroism 4-piece: 1 proc per minute on landed white swings; the game files state none",
		Source:  "EV-0136, EV-0144",
	},
	{
		SpellID: 27787,
		Field:   PPM,
		Value:   0.8,
		Reason:  "Shadowcraft Armor 5-piece, Darkmantle Armor reissue 4-piece: 0.8 procs per minute on landed white swings; the game files state none",
		Source:  "EV-0144",
	},
	{
		SpellID: 450625,
		Field:   PPM,
		Value:   0.9,
		Reason:  "Lightforge Armor 5-piece, Soulforge Armor reissue 4-piece: 0.9 procs per minute on landed white swings and on spell casts priced on their cast time, the one rate that fits both, a log fit not yet validated (EV-0144); the game files state none",
		Source:  "observed",
	},
	{
		SpellID: 450626,
		Field:   PPM,
		Value:   1,
		Reason:  "The Elements 5-piece, The Five Thunders reissue 4-piece: 1 proc per minute on landed white swings and on spell casts priced on their cast time, a log fit not yet validated (EV-0144); the game files state none",
		Source:  "observed",
	},
	{
		SpellID: 450608,
		Field:   PPM,
		Value:   1,
		Reason:  "Wildheart Raiment 5-piece, Feralheart Raiment reissue 4-piece: 1 proc per minute on landed white swings, on spell casts priced on their cast time and on landed melee swings taken priced on the attacker's swing interval, a log fit not yet validated (EV-0144); the game files state none and the dummy effect's 5 is no chance",
		Source:  "observed",
	},
	{
		SpellID: 1324318,
		Field:   PPM,
		Value:   1,
		Reason:  "Violet Sorcerer's Vestments 4-piece: 1 proc per minute on each target a damaging spell lands on, priced on its cast time with instants at 1.5 s; the game files state none",
		Source:  "EV-0169",
	},
	{
		SpellID: 1324323,
		Field:   PPM,
		Value:   0.95,
		Reason:  "Violet Sorcerer's Vestments 5-piece: 0.95 procs per minute on each target a damaging spell lands on, priced on its cast time with instants at 1.5 s; the game files state none",
		Source:  "EV-0169",
	},
	{
		SpellID: 1324684,
		Field:   PPM,
		Value:   1,
		Reason:  "Defias Enforcer's Garb 5-piece: 1 proc per minute on landed melee hits of either hand, priced on the weapon that hit; the game files state none",
		Source:  "EV-0170",
	},
	{
		SpellID: 21978,
		Field:   PPM,
		Value:   1.8,
		Reason:  "Blade of Eternal Darkness: 1.8 procs per minute on the cast time of each landed harmful spell's cast, at least 1.5 s; the game files state none",
		Source:  "EV-0025",
	},
	{
		SpellID: 1300435,
		Field:   PPM,
		Value:   1.5,
		Reason:  "Moonshadow Stave: 1.5 procs per minute on the cast time of the cast behind each landed harmful spell, at least 1.5 s; column reads 100, tooltip says 'Harmful spell casts have a chance'",
		Source:  "EV-0115",
	},
	{
		SpellID: 1079,
		Field:   APCoefPeriodic,
		Value:   0.0114,
		Reason:  "Rip Rank 1: attack power share per tick per combo point, measured",
		Source:  "EV-0031",
	},
	{
		SpellID: 9492,
		Field:   APCoefPeriodic,
		Value:   0.0114,
		Reason:  "Rip Rank 2: 0.0114 * attack power * min(combo points, 4) per tick",
		Source:  "EV-0032",
	},
	{
		SpellID: 9493,
		Field:   APCoefPeriodic,
		Value:   0.0114,
		Reason:  "Rip Rank 3: attack power share per tick per combo point up to 4",
		Source:  "EV-0033",
	},
	{
		SpellID: 9752,
		Field:   APCoefPeriodic,
		Value:   0.0114,
		Reason:  "Rip Rank 4: attack power share per tick per combo point up to 4, not on all 5 points",
		Source:  "EV-0034",
	},
	{
		SpellID: 9894,
		Field:   APCoefPeriodic,
		Value:   0.0114,
		Reason:  "Rip Rank 5: attack power share per tick per combo point up to 4, not on all 5 points",
		Source:  "EV-0036",
	},
	{
		SpellID: 9896,
		Field:   APCoefPeriodic,
		Value:   0.0114, // TODO: Measure Rip Rank 6 in game.
		Reason:  "Rip Rank 6: Rank 1's measured share per tick per combo point",
		Source:  "EV-0031",
	},
	{
		SpellID: 1822,
		Field:   APCoefDirect,
		Value:   0.065,
		Reason:  "Rake Rank 1: measured direct hit share; one share fits the hit and the tick",
		Source:  "EV-0031",
	},
	{
		SpellID: 1822,
		Field:   APCoefPeriodic,
		Value:   0.065,
		Reason:  "Rake Rank 1: measured periodic share; one share fits the hit and the tick",
		Source:  "EV-0031",
	},
	{
		SpellID: 1823,
		Field:   APCoefDirect,
		Value:   0.065,
		Reason:  "Rake Rank 2: measured direct hit share",
		Source:  "EV-0033",
	},
	{
		SpellID: 1823,
		Field:   APCoefPeriodic,
		Value:   0.065,
		Reason:  "Rake Rank 2: measured periodic share",
		Source:  "EV-0033",
	},
	{
		SpellID: 1824,
		Field:   APCoefDirect,
		Value:   0.065,
		Reason:  "Rake Rank 3: measured direct hit share",
		Source:  "EV-0034",
	},
	{
		SpellID: 1824,
		Field:   APCoefPeriodic,
		Value:   0.065,
		Reason:  "Rake Rank 3: measured periodic share",
		Source:  "EV-0034",
	},
	{
		SpellID: 9904,
		Field:   APCoefDirect,
		Value:   0.065,
		Reason:  "Rake Rank 4: measured direct hit share",
		Source:  "EV-0036",
	},
	{
		SpellID: 9904,
		Field:   APCoefPeriodic,
		Value:   0.065,
		Reason:  "Rake Rank 4: measured periodic share",
		Source:  "EV-0036",
	},
	{
		SpellID: 9007,
		Field:   APCoefPeriodic,
		Value:   0.025,
		Reason:  "Pounce Bleed Rank 1: measured periodic share, fits every tick; the game files state none",
		Source:  "EV-0033",
	},
	{
		SpellID: 9824,
		Field:   APCoefPeriodic,
		Value:   0.025,
		Reason:  "Pounce Bleed Rank 2: measured periodic share, fits every tick; the game files state none",
		Source:  "EV-0034",
	},
	{
		SpellID: 9826,
		Field:   APCoefPeriodic,
		Value:   0.025,
		Reason:  "Pounce Bleed Rank 3: measured periodic share, fits every tick; the game files state none",
		Source:  "EV-0036",
	},
	{
		SpellID: 779,
		Field:   APCoefDirect,
		Value:   0.03,
		Reason:  "Swipe Rank 1: 3% of attack power on each hit, measured; the game files state none",
		Source:  "EV-0031",
	},
	{
		SpellID: 780,
		Field:   APCoefDirect,
		Value:   0.03,
		Reason:  "Swipe Rank 2: 3% of attack power on each hit, measured; the game files state none",
		Source:  "EV-0031",
	},
	{
		SpellID: 769,
		Field:   APCoefDirect,
		Value:   0.03,
		Reason:  "Swipe Rank 3: 3% of attack power on each hit, measured; the game files state none",
		Source:  "EV-0033",
	},
	{
		SpellID: 9754,
		Field:   APCoefDirect,
		Value:   0.03,
		Reason:  "Swipe Rank 4: 3% of attack power on each hit, measured; the game files state none",
		Source:  "EV-0034",
	},
	{
		SpellID: 9908,
		Field:   APCoefDirect,
		Value:   0.03,
		Reason:  "Swipe Rank 5: 3% of attack power on each hit; the game files state none",
		Source:  "EV-0038",
	},
	{
		SpellID: 1328618,
		Field:   APCoefDirect,
		Value:   0.037,
		Reason:  "Ferocious Bite Rank 1: attack power share per combo point, measured; the game files state none",
		Source:  "EV-0032",
	},
	{
		SpellID: 22568,
		Field:   APCoefDirect,
		Value:   0.037,
		Reason:  "Ferocious Bite Rank 2: attack power share per combo point, measured with Rank 3; the game files state none",
		Source:  "EV-0032",
	},
	{
		SpellID: 22827,
		Field:   APCoefDirect,
		Value:   0.037,
		Reason:  "Ferocious Bite Rank 3: attack power share per combo point, measured with Rank 2; the game files state none",
		Source:  "EV-0032",
	},
	{
		SpellID: 22828,
		Field:   APCoefDirect,
		Value:   0.037,
		Reason:  "Ferocious Bite Rank 4: attack power share per combo point, measured; the game files state none",
		Source:  "EV-0034",
	},
	{
		SpellID: 22829,
		Field:   APCoefDirect,
		Value:   0.037, // TODO: Measure Ferocious Bite Rank 5 in game.
		Reason:  "Ferocious Bite Rank 5: Ranks 1 to 3's measured share per combo point; the game files state none",
		Source:  "EV-0032",
	},
	{
		SpellID: 31018,
		Field:   APCoefDirect,
		Value:   0.037, // TODO: Measure Ferocious Bite Rank 6 in game.
		Reason:  "Ferocious Bite Rank 6: Ranks 1 to 3's measured share per combo point; the game files state none",
		Source:  "EV-0032",
	},

	// Revenge's Attack Power share, the $APR its tooltip states: 0.225 in Defensive Stance is the stance's
	// own 10%, which stays out of these values. Logs on Ranks 1 and 2 taken before that tooltip fit 0.28 (EV-0049).
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
		Value:   0.25,
		Reason:  "Revenge Rank 4: 0.25 of attack power outside Defensive Stance, measured",
		Source:  "EV-0048",
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

	// Sunder Armor's threat grows with the sheet's Attack Power (EV-0042). The other ranks carry its share
	// until a capture at their level says otherwise.
	{
		SpellID: 7386,
		Field:   APCoefThreat,
		Value:   0.05,
		Reason:  "Sunder Armor Rank 1: threat is 34 + 0.05 x Attack Power, times the stance, rounded down",
		Source:  "EV-0042",
	},
	{
		SpellID: 7405,
		Field:   APCoefThreat,
		Value:   0.05, // TODO: Measure Sunder Armor Rank 2 in game.
		Reason:  "Sunder Armor Rank 2: Rank 1's measured share",
		Source:  "EV-0042",
	},
	{
		SpellID: 8380,
		Field:   APCoefThreat,
		Value:   0.05,
		Reason:  "Sunder Armor Rank 3: measured 117 + 0.05 * sheet attack power",
		Source:  "EV-0048",
	},
	{
		SpellID: 11596,
		Field:   APCoefThreat,
		Value:   0.05, // TODO: Measure Sunder Armor Rank 4 in game.
		Reason:  "Sunder Armor Rank 4: Rank 1's measured share",
		Source:  "EV-0042",
	},
	{
		SpellID: 11597,
		Field:   APCoefThreat,
		Value:   0.05, // TODO: Measure Sunder Armor Rank 5 in game.
		Reason:  "Sunder Armor Rank 5: Rank 1's measured share",
		Source:  "EV-0042",
	},

	// Heroic Strike and Cleave add flat threat beyond their damage, and none with Attack Power (EV-0042).
	// The bonus is multiplied by the stance with the damage, and Cleave splits it between the targets it
	// hits. Heroic Strike Ranks 1 to 4's measured bonuses fit no single share per spell level (EV-0043);
	// Ranks 5 and 6 measure Rank 4's 2.5 a level (EV-0048, EV-0174), and Ranks 7 to 9 carry it. Cleave's
	// measured Ranks 1 and 2 fit 1.5 a level (EV-0042, EV-0045), Rank 3 measures 62 and Rank 4 80, and
	// Rank 5 carries Rank 4's 1.6 (EV-0050). Each rank's bonus is a
	// whole number: a measured bonus rounds to the nearest one, and each share times its spell level is
	// whole.
	{
		SpellID: 78,
		Field:   FlatThreat,
		Value:   17,
		Reason:  "Heroic Strike Rank 1: threat is the damage plus a measured flat bonus, times the stance",
		Source:  "EV-0043",
	},
	{
		SpellID: 284,
		Field:   FlatThreat,
		Value:   28,
		Reason:  "Heroic Strike Rank 2: threat is the damage plus a measured flat bonus, times the stance",
		Source:  "EV-0043",
	},
	{
		SpellID: 285,
		Field:   FlatThreat,
		Value:   40,
		Reason:  "Heroic Strike Rank 3: threat is the damage + 40, times the stance",
		Source:  "EV-0043",
	},
	{
		SpellID: 1608,
		Field:   FlatThreat,
		Value:   60,
		Reason:  "Heroic Strike Rank 4: threat is the damage + 60, times the stance",
		Source:  "EV-0004",
	},
	{
		SpellID: 11564,
		Field:   ThreatPerSpellLevel,
		Value:   2.5,
		Reason:  "Heroic Strike Rank 5: Rank 4's share per spell level, measured",
		Source:  "EV-0048",
	},
	{
		SpellID: 11565,
		Field:   FlatThreat,
		Value:   100,
		Reason:  "Heroic Strike Rank 6: threat is the damage + 100, times the stance",
		Source:  "EV-0174",
	},
	{
		SpellID: 11566,
		Field:   ThreatPerSpellLevel,
		Value:   2.5, // TODO: Measure Heroic Strike Rank 7 in game.
		Reason:  "Heroic Strike Rank 7: Rank 4's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 11567,
		Field:   ThreatPerSpellLevel,
		Value:   2.5, // TODO: Measure Heroic Strike Rank 8 in game.
		Reason:  "Heroic Strike Rank 8: Rank 4's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 25286,
		Field:   ThreatPerSpellLevel,
		Value:   2.5, // TODO: Measure Heroic Strike Rank 9 in game.
		Reason:  "Heroic Strike Rank 9: Rank 4's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 845,
		Field:   FlatThreat,
		Value:   30,
		Reason:  "Cleave Rank 1: adds 30 beyond its damage, split over the targets it lands on, times the stance",
		Source:  "EV-0045",
	},
	{
		SpellID: 7369,
		Field:   FlatThreat,
		Value:   46,
		Reason:  "Cleave Rank 2: adds 46 beyond its damage at spell level 30, split over the targets it lands on, times the stance",
		Source:  "EV-0042",
	},
	{
		SpellID: 11608,
		Field:   FlatThreat,
		Value:   62,
		Reason:  "Cleave Rank 3: adds 62 beyond its damage, split over the targets it lands on, times the stance",
		Source:  "EV-0050",
	},
	{
		SpellID: 11609,
		Field:   FlatThreat,
		Value:   80,
		Reason:  "Cleave Rank 4: adds 80 beyond its damage, split over the targets it lands on, times the stance",
		Source:  "EV-0050",
	},
	{
		SpellID: 20569,
		Field:   FlatThreat,
		Value:   96, // TODO: Measure Cleave Rank 5 in game.
		Reason:  "Cleave Rank 5: Rank 4's measured 1.6 a spell level at spell level 60",
		Source:  "EV-0050",
	},

	// Slam adds flat threat beyond its damage, times the stance: Rank 1's 30 at spell level 20, carried
	// as 1.5 a level to every rank and to the instant ranks Improved Slam swaps in (EV-0004).
	{
		SpellID: 1240193,
		Field:   ThreatPerSpellLevel,
		Value:   1.5, // TODO: Measure Slam Rank 1 in game.
		Reason:  "Slam Rank 1: threat is the damage + 30, times the stance",
		Source:  "EV-0004",
	},
	{
		SpellID: 1464,
		Field:   ThreatPerSpellLevel,
		Value:   1.5, // TODO: Measure Slam Rank 2 in game.
		Reason:  "Slam Rank 2: Rank 1's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 8820,
		Field:   ThreatPerSpellLevel,
		Value:   1.5, // TODO: Measure Slam Rank 3 in game.
		Reason:  "Slam Rank 3: Rank 1's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 11604,
		Field:   ThreatPerSpellLevel,
		Value:   1.5, // TODO: Measure Slam Rank 4 in game.
		Reason:  "Slam Rank 4: Rank 1's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 11605,
		Field:   ThreatPerSpellLevel,
		Value:   1.5, // TODO: Measure Slam Rank 5 in game.
		Reason:  "Slam Rank 5: Rank 1's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 1310196,
		Field:   ThreatPerSpellLevel,
		Value:   1.5, // TODO: Measure Improved Slam's Rank 1 in game.
		Reason:  "Slam Rank 1 under Improved Slam: Slam Rank 1's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 1310197,
		Field:   ThreatPerSpellLevel,
		Value:   1.5, // TODO: Measure Improved Slam's Rank 2 in game.
		Reason:  "Slam Rank 2 under Improved Slam: Slam Rank 1's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 1310198,
		Field:   ThreatPerSpellLevel,
		Value:   1.5, // TODO: Measure Improved Slam's Rank 3 in game.
		Reason:  "Slam Rank 3 under Improved Slam: Slam Rank 1's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 1310199,
		Field:   ThreatPerSpellLevel,
		Value:   1.5, // TODO: Measure Improved Slam's Rank 4 in game.
		Reason:  "Slam Rank 4 under Improved Slam: Slam Rank 1's measured share per spell level",
		Source:  "EV-0004",
	},
	{
		SpellID: 1310200,
		Field:   ThreatPerSpellLevel,
		Value:   1.5, // TODO: Measure Improved Slam's Rank 5 in game.
		Reason:  "Slam Rank 5 under Improved Slam: Slam Rank 1's measured share per spell level",
		Source:  "EV-0004",
	},

	// Shield Slam adds flat threat beyond its damage, times the stance, once on a crit too (EV-0004,
	// EV-0050). Rank 4 extends the line through Ranks 2 and 3 (4.33 a spell level from 48 to 54) to spell
	// level 60, rounded to the nearest whole number.
	{
		SpellID: 23922,
		Field:   FlatThreat,
		Value:   178,
		Reason:  "Shield Slam Rank 1: threat is the damage + 178, times the stance",
		Source:  "EV-0004",
	},
	{
		SpellID: 23923,
		Field:   FlatThreat,
		Value:   203,
		Reason:  "Shield Slam Rank 2: threat is the damage + 203, times the stance",
		Source:  "EV-0050",
	},
	{
		SpellID: 23924,
		Field:   FlatThreat,
		Value:   229,
		Reason:  "Shield Slam Rank 3: threat is the damage + 229, times the stance",
		Source:  "EV-0050",
	},
	{
		SpellID: 23925,
		Field:   FlatThreat,
		Value:   255, // TODO: Measure Shield Slam Rank 4 in game.
		Reason:  "Shield Slam Rank 4: the line through the measured Ranks 2 and 3 at spell level 60",
		Source:  "EV-0050",
	},
	{
		SpellID: 1324298,
		Field:   ProcChancePct,
		Value:   100,
		Reason:  "Krol'dok Resolve, Krol'dok Battlegear 4pc: 'Chance on block' heals on every blocked melee or ranged hit off its cooldown",
		Source:  "EV-0160",
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
