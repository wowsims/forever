package dbc

import (
	"regexp"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/overrides"
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
	{name: "worgen", family: proto.CreatureFamily_CreatureFamilyWorgen},
	{name: "murloc", family: proto.CreatureFamily_CreatureFamilyMurloc},
	{name: "furbolg", family: proto.CreatureFamily_CreatureFamilyFurbolg},
	{name: "dwarves", family: proto.CreatureFamily_CreatureFamilyDwarf},
	{name: "dinosaur", family: proto.CreatureFamily_CreatureFamilyDinosaur},
	{name: "serpentine", family: proto.CreatureFamily_CreatureFamilySerpentine},
}

// The words a tooltip puts before the creatures an effect counts against: "damage to Plants",
// "tripled against Undead", "when fighting Undead and Demons", "if there is an elemental nearby".
var tooltipTargetLead = regexp.MustCompile(`(?i)\b(?:against|damage to|when fighting|versus|there is an?)\s+`)
var tooltipTargetSkip = regexp.MustCompile(`(?i)^(?:all|the|and|or|,)\b\s*|^,\s*`)

// The mob types and creature groups a tooltip names after a lead: the run of names, joined by "and",
// "or" and commas, that follows it. The run ends at the first other word, so "against Undead have a
// chance to summon a Fire Elemental" names the Undead alone.
func tooltipTargets(tooltip string) (mobTypes []proto.MobType, families []proto.CreatureFamily) {
	for _, lead := range tooltipTargetLead.FindAllStringIndex(tooltip, -1) {
		rest := strings.ToLower(tooltip[lead[1]:])
	run:
		for {
			if skip := tooltipTargetSkip.FindString(rest); skip != "" {
				rest = rest[len(skip):]
				continue
			}
			for _, target := range tooltipTargetNames {
				after, found := strings.CutPrefix(rest, target.name)
				if !found {
					continue
				}
				after = strings.TrimPrefix(after, "s")
				if after != "" && (after[0] >= 'a' && after[0] <= 'z') {
					continue
				}
				if target.mob != proto.MobType_MobTypeUnknown {
					mobTypes = append(mobTypes, target.mob)
				} else {
					families = append(families, target.family)
				}
				rest = strings.TrimLeft(after, " ")
				continue run
			}
			break
		}
	}
	return mobTypes, families
}

// The terrain areas, mob types and creature groups an item's effects name, which the gear picker
// filters by: the areas a spell requires or is stronger in, the mob types a versus aura counts
// against and the ones a spell can only hit. A tooltip states the rest, the creature groups among
// them, which no row holds (tooltipTargets). Every item effect counts, whatever triggers it, with
// each spell its trigger chain reaches. A spell that requires a single zone names no terrain area.
func ItemEffectTypes(itemID int32) ([]proto.AreaType, []proto.MobType, []proto.CreatureFamily) {
	var areaTypes []proto.AreaType
	var mobTypes []proto.MobType
	var families []proto.CreatureFamily
	addArea := func(group int32) {
		if areaType := spelldata.AreaTypeOfGroup(group); areaType != proto.AreaType_AreaTypeUnknown && !slices.Contains(areaTypes, areaType) {
			areaTypes = append(areaTypes, areaType)
		}
	}
	addMobs := func(found ...proto.MobType) {
		for _, mobType := range found {
			if !slices.Contains(mobTypes, mobType) {
				mobTypes = append(mobTypes, mobType)
			}
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
		for _, bonus := range overrides.AreaBonuses {
			if int(bonus.SpellID) == spellID {
				for _, group := range bonus.Groups {
					addArea(group)
				}
			}
		}
		addMobs(spelldata.MobTypes(spelldata.Find(int32(spellID)).TargetCreatureType)...)
		tooltipMobs, tooltipFamilies := tooltipTargets(spell.Description)
		addMobs(tooltipMobs...)
		for _, family := range tooltipFamilies {
			if !slices.Contains(families, family) {
				families = append(families, family)
			}
		}

		for _, effect := range effects {
			if slices.Contains(versusAuras, effect.EffectAura) && len(effect.EffectMiscValues) > 0 {
				addMobs(spelldata.MobTypes(int32(effect.EffectMiscValues[0]))...)
			}
			if effect.EffectTriggerSpell != 0 {
				visit(effect.EffectTriggerSpell)
			}
		}
	}
	for _, itemEffect := range dbcInstance.ItemEffectsByParentID[int(itemID)] {
		visit(itemEffect.SpellID)
	}
	addMobs(tooltipMobTypes[itemID]...)

	slices.Sort(areaTypes)
	slices.Sort(mobTypes)
	slices.Sort(families)
	return areaTypes, mobTypes, families
}
