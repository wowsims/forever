package dbc

import (
	"regexp"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The auras whose misc value is a creature type mask.
var versusAuras = []EffectAuraType{
	dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS,
	dbcenums.A_MOD_RANGED_ATTACK_POWER_VERSUS,
	dbcenums.A_MOD_FLAT_SPELL_DAMAGE_VERSUS,
	dbcenums.A_MOD_DAMAGE_DONE_VERSUS,
	dbcenums.A_MOD_CRIT_PERCENT_VERSUS,
}

// The items with no item effect row, whose mob types only the item tooltip states.
var tooltipMobTypes = map[int32][]proto.MobType{
	23206: {proto.MobType_MobTypeDemon, proto.MobType_MobTypeUndead}, // Mark of the Champion: attack power "when fighting Undead and Demons"
	23207: {proto.MobType_MobTypeDemon, proto.MobType_MobTypeUndead}, // Mark of the Champion: spell damage "against Undead and Demons"
}

// What a spell tooltip calls the creatures an effect counts against, each also read with a plural s.
// A name that holds another comes first, so "Earth Elementals" is no Elemental.
var tooltipTargetNames = []struct {
	name   string
	mob    proto.MobType
	family proto.CreatureFamily
}{
	{name: "earth elemental", family: proto.CreatureFamily_CreatureFamilyEarthElemental},
	{name: "fire elemental", family: proto.CreatureFamily_CreatureFamilyFireElemental},
	{name: "mountain giant", family: proto.CreatureFamily_CreatureFamilyMountainGiant},
	{name: "undead", mob: proto.MobType_MobTypeUndead},
	{name: "demon", mob: proto.MobType_MobTypeDemon},
	{name: "beast", mob: proto.MobType_MobTypeBeast},
	{name: "dragonkin", mob: proto.MobType_MobTypeDragonkin},
	{name: "elemental", mob: proto.MobType_MobTypeElemental},
	{name: "giant", mob: proto.MobType_MobTypeGiant},
	{name: "humanoid", mob: proto.MobType_MobTypeHumanoid},
	{name: "mechanical", mob: proto.MobType_MobTypeMechanical},
	{name: "naga", family: proto.CreatureFamily_CreatureFamilyNaga},
	{name: "satyr", family: proto.CreatureFamily_CreatureFamilySatyr},
	{name: "aquatic", family: proto.CreatureFamily_CreatureFamilyAquatic},
	{name: "swine", family: proto.CreatureFamily_CreatureFamilySwine},
	{name: "plant", family: proto.CreatureFamily_CreatureFamilyPlant},
	{name: "troll", family: proto.CreatureFamily_CreatureFamilyTroll},
	{name: "orc", family: proto.CreatureFamily_CreatureFamilyOrc},
	{name: "ogre", family: proto.CreatureFamily_CreatureFamilyOgre},
	{name: "wolves", family: proto.CreatureFamily_CreatureFamilyWolf},
	{name: "wolf", family: proto.CreatureFamily_CreatureFamilyWolf},
	{name: "worgen", family: proto.CreatureFamily_CreatureFamilyWorgen},
	{name: "murloc", family: proto.CreatureFamily_CreatureFamilyMurloc},
	{name: "furbolg", family: proto.CreatureFamily_CreatureFamilyFurbolg},
	{name: "dwarves", family: proto.CreatureFamily_CreatureFamilyDwarf},
	{name: "dwarf", family: proto.CreatureFamily_CreatureFamilyDwarf},
	{name: "dinosaur", family: proto.CreatureFamily_CreatureFamilyDinosaur},
	{name: "serpentine", family: proto.CreatureFamily_CreatureFamilySerpentine},
}

// The words a tooltip puts before the creatures an effect counts against: "damage to Plants",
// "tripled against Undead", "when fighting Undead and Demons", "if there is an elemental nearby".
var tooltipTargetLead = regexp.MustCompile(`\b(?:against|damage to|when fighting|versus|there is an?|if attacking an?)\s+`)

// A sentence that opens with the creatures: "Trolls and Orcs take 2 times as much damage",
// "Undead that strike you are 3 times as likely to trigger this effect".
var tooltipTargetSentence = regexp.MustCompile(`(?:^|[.!]\s+)`)
var tooltipTargetVerb = regexp.MustCompile(`^(?:targets |enemies )?(?:take|that strike)\b`)
var tooltipTargetSkip = regexp.MustCompile(`^(?:(?:all|the|and|or)\b|,)\s*`)

// The run of names at the start of lowered text, joined by "and", "or" and commas, and the text after
// it. The run ends at the first other word, so "undead have a chance to summon a fire elemental"
// names the Undead alone.
func tooltipNameRun(rest string) (mobTypes []proto.MobType, families []proto.CreatureFamily, after string) {
	for {
		if skip := tooltipTargetSkip.FindString(rest); skip != "" {
			rest = rest[len(skip):]
			continue
		}
		matched := false
		for _, target := range tooltipTargetNames {
			tail, found := strings.CutPrefix(rest, target.name)
			if !found {
				continue
			}
			tail = strings.TrimPrefix(tail, "s")
			if tail != "" && tail[0] >= 'a' && tail[0] <= 'z' {
				continue
			}
			if target.mob != proto.MobType_MobTypeUnknown {
				mobTypes = append(mobTypes, target.mob)
			} else {
				families = append(families, target.family)
			}
			rest, matched = strings.TrimLeft(tail, " "), true
			break
		}
		if !matched {
			return mobTypes, families, rest
		}
	}
}

// The mob types and creature groups a tooltip states a bonus against: the names after a lead, and
// the names that open a sentence about what they take or do. "Only works against Beasts" limits an
// effect and states no bonus, so it names none.
func tooltipTargets(tooltip string) (mobTypes []proto.MobType, families []proto.CreatureFamily) {
	lower := strings.ToLower(tooltip)
	for _, lead := range tooltipTargetLead.FindAllStringIndex(lower, -1) {
		if strings.HasSuffix(lower[:lead[0]], "only works ") {
			continue
		}
		mobs, fams, _ := tooltipNameRun(lower[lead[1]:])
		mobTypes, families = append(mobTypes, mobs...), append(families, fams...)
	}
	for _, start := range tooltipTargetSentence.FindAllStringIndex(lower, -1) {
		if mobs, fams, after := tooltipNameRun(lower[start[1]:]); tooltipTargetVerb.MatchString(after) {
			mobTypes, families = append(mobTypes, mobs...), append(families, fams...)
		}
	}
	return mobTypes, families
}

// The terrain areas, mob types and creature groups an item's effects state a bonus in or against,
// which the gear picker filters by: the areas a spell requires or is stronger in and the mob types a
// versus aura counts against. A tooltip states the rest, the creature groups among them, which no row
// holds (tooltipTargets). Every item effect counts, whatever triggers it, with each spell its trigger
// chain reaches. A spell that requires a single zone names no terrain area, and one that can only
// hit a mob type (TargetCreatureType: a stun, a charm) states no bonus against it.
func ItemEffectTypes(itemID int32) ([]proto.AreaType, []proto.MobType, []proto.CreatureFamily) {
	var areaTypes []proto.AreaType
	var mobTypes []proto.MobType
	var families []proto.CreatureFamily
	addArea := func(group int32) {
		if areaType := spelldata.AreaTypeOfGroup(group); areaType != proto.AreaType_AreaTypeUnknown {
			areaTypes = append(areaTypes, areaType)
		}
	}

	w := newChainWalker()
	var visit func(spellID int)
	visit = func(spellID int) {
		if w.visited[spellID] {
			return
		}
		effects := w.effects(spellID)
		spell := dbcInstance.Spells[spellID]

		addArea(spell.RequiredAreasID)
		for _, group := range spelldata.Find(int32(spellID)).AreaBonusGroups {
			addArea(group)
		}
		tooltipMobs, tooltipFamilies := tooltipTargets(spell.Description)
		mobTypes = append(mobTypes, tooltipMobs...)
		families = append(families, tooltipFamilies...)

		for _, effect := range effects {
			if slices.Contains(versusAuras, effect.EffectAura) && len(effect.EffectMiscValues) > 0 {
				mobTypes = append(mobTypes, spelldata.MobTypes(int32(effect.EffectMiscValues[0]))...)
			}
			if effect.EffectTriggerSpell != 0 {
				visit(effect.EffectTriggerSpell)
			}
		}
	}
	for _, itemEffect := range dbcInstance.ItemEffectsByParentID[int(itemID)] {
		visit(itemEffect.SpellID)
	}
	mobTypes = append(mobTypes, tooltipMobTypes[itemID]...)

	slices.Sort(areaTypes)
	slices.Sort(mobTypes)
	slices.Sort(families)
	return slices.Compact(areaTypes), slices.Compact(mobTypes), slices.Compact(families)
}
