package database

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// What the tooltip says about a proc that its columns cannot, baked onto the row at generation.
// Spell.Description_lang is generator input only - the store carries the reading, never the text -
// so the sim reads a decided shape instead of re-parsing English at runtime.
//
// The four shapes are the ones docs/spell_data/procs.md names: the column is the
// roll, an effect's value is the roll, the column is a sentinel for "fires on its own condition",
// or nothing states a rate at all and an override owes one.

// The tooltip's own name for SpellAuraOptions.ProcChance. A chance stated as "$<id>h" is another
// spell's column and cannot match, since the id sits between the $ and the h. The chance is
// sometimes scaled - "${$h/2}% chance" - which still says the column is where it comes from.
var tooltipOwnChance = regexp.MustCompile(`\$h`)

// A chance stated as an effect's value: "a $m2% chance to generate Rage". The word has to follow
// the token, because the same token is how a tooltip states a magnitude: Flurry's "$m1% melee
// attack speed" and Deep Wounds' "$m1% of your weapon's average damage" are the no-roll shape, and
// reading either as a chance turns an aura that always fires into a percentage. The digit is the
// game files' EffectIndex plus one.
var tooltipEffectChance = regexp.MustCompile(`\$[ms]([123])%\s+chance`)

// Damage divided among the targets hit, which the server knows per spell and no column states:
// Everlook Pathcarver's "split between up to $s3 nearby enemies", the Meteors' "divided up evenly
// among all affected targets" and Shard of the Fallen Star's "$s1 total Fire damage".
var tooltipSplitsDamage = regexp.MustCompile(`split between|divided up evenly|\$s\d total \w+ damage`)

// Reads the proc shape and the damage split off the tooltip and the aura columns and writes them
// onto the row.
func applyTooltipHints(t *spellTables, s *storeSpell) {
	description := t.Descriptions[s.ID]

	ownChance := tooltipOwnChance.MatchString(description)

	s.SplitsDamage = tooltipSplitsDamage.MatchString(description)
	s.ProcHint = procTooltipHints(description)
	s.ProcChanceSource, s.ProcChanceEffect = procChanceSource(description, ownChance, s)
	s.tooltipStatesChance = ownChance || s.ProcChanceSource == procChanceEffectN

	if grant, ok := t.EnchantGrants[s.ID]; ok && description == "" {
		applyEnchantGrantHints(t.Descriptions[grant], s)
	}
}

// What an enchant's grant states about the equip spell it hangs on a hit. The grant describes the
// enchant rather than the spell's mask, so only what no mask can state is read off it: a named
// ability, an outcome the mask has no bit for, the wearer's attack dodged or parried, and whether a
// column of 100 is a rate the rows do not carry. Which hits feed the proc stays the mask's to say.
func applyEnchantGrantHints(grant string, s *storeSpell) {
	s.ProcHint |= procTooltipHints(grant) & (core.ProcHintNamedAbility | core.ProcHintOutcomeTaken | core.ProcHintAttackAvoided)

	if s.ProcChanceSource == procChanceAlways && tooltipStatesAnUnknownRate(grant) {
		s.ProcChanceSource = procChancePPM
	}
}

// A combat spell's roll as its enchantments state it, which is the one the game casts it at: the
// spell's own column is no roll there, and on Fiery Blaze's 6297 there is none. A tooltip stating a
// chance of its own that the enchantments contradict is an error.
func applyEnchantChance(t *spellTables, s *storeSpell) error {
	stated, ok := t.EnchantChances[s.ID]
	if !ok {
		return nil
	}
	if s.tooltipStatesChance && (s.ProcChanceSource != procChanceColumn || s.ProcChance != stated.Chance) {
		return fmt.Errorf("spell %d states its own proc chance in the tooltip, and enchantments %v state %d%%",
			s.ID, stated.Enchants, stated.Chance)
	}

	s.ProcChance = stated.Chance
	s.ProcChanceSource, s.ProcChanceEffect = procChanceColumn, 0
	s.overrideNotes = append(s.overrideNotes,
		fmt.Sprintf("enchantment: ProcChance %d -- EffectPointsMin of SpellItemEnchantment %s", stated.Chance, joinIDs(stated.Enchants)))
	return nil
}

func procChanceSource(description string, ownChance bool, s *storeSpell) (storeProcChanceSource, int8) {
	if ownChance {
		return procChanceColumn, 0
	}

	// A tooltip names several effects - Shield Specialization states "$s1% block" and "$m2% chance
	// to generate rage" - so the one the chance is on is the one carrying the proc aura, not the
	// first one the text mentions.
	//
	// The tooltip counts by the game files' EffectIndex and the store by position, which differ on the
	// 46 rows whose indices have gaps. The position is what is emitted, since EffectN is how the
	// sim reads the effect back.
	for _, m := range tooltipEffectChance.FindAllStringSubmatch(description, -1) {
		index := uint8(m[1][0]-'0') - 1
		if position, ok := s.effectPosition(index); ok && isProcEffect(&s.Effects[position]) {
			return procChanceEffectN, int8(position + 1)
		}
	}

	// 100 and 101 are both the game files' "no roll here": the aura fires whenever its own condition
	// is met, and on a spell that is not a proc at all they mean nothing.
	//
	// Except where the tooltip says the effect only happens sometimes. A proc whose text reads
	// "Chance to strike your ranged target" next to a column of 100 is the same convention as the
	// 101 sentinel - the rate lives outside the spell data - so it is a rate somebody owes rather
	// than a proc on every hit.
	if s.ProcChance == 100 || s.ProcChance == 101 {
		if s.triggersAProc() && tooltipStatesAnUnknownRate(description) {
			return procChancePPM, 0
		}
		return procChanceAlways, 0
	}

	// A proc aura with no chance anywhere is the shape whose rate the game files do not carry, so it
	// has to come from an override into RPPM.
	if s.ProcChance == 0 && s.triggersAProc() {
		return procChancePPM, 0
	}

	return procChanceColumn, 0
}

// An effect a tooltip can state a chance for: the two trigger auras, and the dummy the server hangs
// a hand-written proc off - Furor's and Unbridled Wrath's chances both sit on one.
func isProcEffect(e *storeEffect) bool {
	return e.Aura.IsProcTrigger() || e.Aura == dbcenums.A_DUMMY
}

// Whether the spell fires something through the game's own proc machinery, which is what makes a
// missing rate a rate somebody owes. The dummy is deliberately out: Rip, Rupture and Arcane
// Missiles all carry one, and none of them is a proc.
func (s *storeSpell) triggersAProc() bool {
	for i := range s.Effects {
		if s.Effects[i].Aura.IsProcTrigger() {
			return true
		}
	}
	return false
}

// Where in the row the effect the game files put under an index sits: EffectIndex is the tooltip's
// numbering and has gaps, while the slice is packed.
func (s *storeSpell) effectPosition(index uint8) (int, bool) {
	for i := range s.Effects {
		if s.Effects[i].Index == index {
			return i, true
		}
	}
	return 0, false
}

// The store's ProcChanceSource, mirrored here the way the row structs are. The order is the store's
// own: the emitted name comes from this table, not from the number.
type storeProcChanceSource uint8

const (
	procChanceColumn storeProcChanceSource = iota
	procChanceEffectN
	procChanceAlways
	procChancePPM
)

func (s storeProcChanceSource) String() string {
	switch s {
	case procChanceEffectN:
		return "ProcChanceEffectN"
	case procChanceAlways:
		return "ProcChanceAlways"
	case procChancePPM:
		return "ProcChancePPM"
	default:
		return "ProcChanceColumn"
	}
}

// The hint bits as the generated file names them, so a row states which words the reading came
// from rather than a number.
func formatProcHint(hint core.ProcHint) string {
	names := []struct {
		bit  core.ProcHint
		name string
	}{
		{core.ProcHintCastTrigger, "core.ProcHintCastTrigger"},
		{core.ProcHintCrit, "core.ProcHintCrit"},
		{core.ProcHintHeals, "core.ProcHintHeals"},
		{core.ProcHintPureHeal, "core.ProcHintPureHeal"},
		{core.ProcHintNamedAbility, "core.ProcHintNamedAbility"},
		{core.ProcHintOutcomeTaken, "core.ProcHintOutcomeTaken"},
		{core.ProcHintAttackDodged, "core.ProcHintAttackDodged"},
		{core.ProcHintAttackParried, "core.ProcHintAttackParried"},
		{core.ProcHintDamage, "core.ProcHintDamage"},
	}

	var set []string
	for _, n := range names {
		if hint.Matches(n.bit) {
			set = append(set, n.name)
		}
	}
	return strings.Join(set, " | ")
}
