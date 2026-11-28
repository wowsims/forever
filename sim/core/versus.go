package core

import (
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

// The mob type each creature family belongs to. A target of a family and no mob type of its own counts
// as the family's mob type in every versus check: a Fire Elemental with none is an Elemental. A target's
// own mob type stays its only one, so an undead Troll stays Undead. Aquatic spans Beasts, Elementals and
// Humanoids and Serpentine Beasts and Humanoids, so neither gives one. make go-to-ts writes the UI's copy
// (CREATURE_FAMILY_MOB_TYPES) from this table.
var creatureFamilyMobTypes = [...]proto.MobType{
	proto.CreatureFamily_CreatureFamilyUnknown:        proto.MobType_MobTypeUnknown,
	proto.CreatureFamily_CreatureFamilyNaga:           proto.MobType_MobTypeHumanoid,
	proto.CreatureFamily_CreatureFamilySatyr:          proto.MobType_MobTypeDemon,
	proto.CreatureFamily_CreatureFamilyAquatic:        proto.MobType_MobTypeUnknown,
	proto.CreatureFamily_CreatureFamilyEarthElemental: proto.MobType_MobTypeElemental,
	proto.CreatureFamily_CreatureFamilyFireElemental:  proto.MobType_MobTypeElemental,
	proto.CreatureFamily_CreatureFamilyMountainGiant:  proto.MobType_MobTypeGiant,
	proto.CreatureFamily_CreatureFamilySwine:          proto.MobType_MobTypeBeast,
	proto.CreatureFamily_CreatureFamilyPlant:          proto.MobType_MobTypeElemental,
	proto.CreatureFamily_CreatureFamilyTroll:          proto.MobType_MobTypeHumanoid,
	proto.CreatureFamily_CreatureFamilyOrc:            proto.MobType_MobTypeHumanoid,
	proto.CreatureFamily_CreatureFamilyOgre:           proto.MobType_MobTypeHumanoid,
	proto.CreatureFamily_CreatureFamilyWolf:           proto.MobType_MobTypeBeast,
	proto.CreatureFamily_CreatureFamilyWorgen:         proto.MobType_MobTypeHumanoid,
	proto.CreatureFamily_CreatureFamilyMurloc:         proto.MobType_MobTypeHumanoid,
	proto.CreatureFamily_CreatureFamilyFurbolg:        proto.MobType_MobTypeHumanoid,
	proto.CreatureFamily_CreatureFamilyDwarf:          proto.MobType_MobTypeHumanoid,
	proto.CreatureFamily_CreatureFamilyDinosaur:       proto.MobType_MobTypeBeast,
	proto.CreatureFamily_CreatureFamilySerpentine:     proto.MobType_MobTypeUnknown,
}

// The target's own mob type, or its family's where it has none.
func targetMobType(options *proto.Target) proto.MobType {
	if options.MobType != proto.MobType_MobTypeUnknown || int(options.CreatureFamily) >= len(creatureFamilyMobTypes) {
		return options.MobType
	}
	return creatureFamilyMobTypes[options.CreatureFamily]
}

// The targets a versus effect names: any of the mob types or creature families.
type VersusTargets struct {
	MobTypes []proto.MobType
	Families []proto.CreatureFamily
}

// Whether the unit is of one of the mob types or families.
func (targets VersusTargets) Matches(unit *Unit) bool {
	return slices.Contains(targets.MobTypes, unit.MobType) || slices.Contains(targets.Families, unit.CreatureFamily)
}

// Whether the targets name no mob type and no family.
func (targets VersusTargets) IsEmpty() bool {
	return len(targets.MobTypes) == 0 && len(targets.Families) == 0
}

// Whether any of the units is one of the targets.
func (targets VersusTargets) AnyIn(units []*Unit) bool {
	return slices.ContainsFunc(units, targets.Matches)
}

// A factor a spell deals or restores against a target of any of the mob types or creature families,
// once however many of them the target is. The game files state the factor on a row and name the
// targets only in its tooltip.
type VersusBonus struct {
	VersusTargets
	Multiplier float64
}

// The factor against the target, 1 where it is none of the mob types and families or there is no bonus.
func (bonus *VersusBonus) Against(target *Unit) float64 {
	if bonus != nil && bonus.Matches(target) {
		return bonus.Multiplier
	}
	return 1
}

func (bonus *VersusBonus) namesFamilyOf(target *Unit) bool {
	return bonus != nil && slices.Contains(bonus.Families, target.CreatureFamily)
}

// The proc that casts a spell, where it fires only against the targets the bonus names or more often
// against them, for the combat log to name on the spell's hits on them. The bonus's Multiplier is how
// many times likelier the proc is against them, 0 where it fires against no other target.
type VersusProc struct {
	Source ActionID
	VersusBonus
}

// The factor a source's spell deals, restores or adds against the mob types and families, kept for the
// stats sheet under the source: an item, or the spell of a set bonus or an enchant. A damage spell
// applies it as its SpellConfig.VersusBonus; any other amount is multiplied by Against at the call site.
func (character *Character) RegisterVersus(source ActionID, kind proto.VersusEffectKind, bonus VersusBonus) *VersusBonus {
	character.versusEffects = append(character.versusEffects, &proto.VersusEffect{Source: source.ToProto(), Kind: kind,
		MobTypes: bonus.MobTypes, CreatureFamilies: bonus.Families, Multiplier: bonus.Multiplier})
	return &bonus
}

// The proc of a source's spell that fires against the targets alone, kept for the stats sheet.
func (character *Character) RegisterVersusProc(source ActionID, targets VersusTargets) *VersusProc {
	bonus := character.RegisterVersus(source, proto.VersusEffectKind_VersusEffectKindProcOnly, VersusBonus{VersusTargets: targets})
	return &VersusProc{Source: source, VersusBonus: *bonus}
}

// Runs apply, which registers the effects of a set bonus or of items only the item swap set holds, and
// leaves the factors it registers off the stats sheet.
func (character *Character) registerSwapOnly(apply func()) {
	start := len(character.versusEffects)
	apply()
	character.versusEffects = character.versusEffects[:start]
}

// The registered factors of the gear the character wears, each item source named by its worn item.
func (character *Character) versusEffectsProto() []*proto.VersusEffect {
	var effects []*proto.VersusEffect
	for _, effect := range character.versusEffects {
		if itemID := effect.Source.GetItemId(); itemID != 0 {
			slot := slices.IndexFunc(character.Equipment[:], func(item Item) bool { return item.ID == itemID })
			if slot < 0 {
				continue
			}
			effect.ItemId, effect.ItemName = itemID, character.Equipment[slot].Name
		}
		effects = append(effects, effect)
	}
	return effects
}
