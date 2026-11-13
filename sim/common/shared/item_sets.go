package shared

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// A bonus whose rows the parse states whole, exposed to the APL under the first row's id. A row's spell
// modifiers name its class's spells by class flags, so on another class they reach nothing. A row whose
// one effect is a proc trigger fires its spell off the row's own listener instead.
func ParsedSetBonus(rows ...*spelldata.Spell) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		character := agent.GetCharacter()
		for _, row := range rows {
			if isSetBonusProc(row) {
				attachSetBonusProc(character, setBonusAura, row)
				continue
			}
			spelldata.ParseEffects(character, setBonusAura, row)
		}
		setBonusAura.ExposeToAPL(rows[0].ID)
	}
}

// What a set bonus proc fires, read off the spell its row names.
type setBonusProcShape uint8

const (
	setBonusProcUnbuilt  setBonusProcShape = iota
	setBonusProcDamage                     // damage at once, a spell cast at the proc's target
	setBonusProcEnergize                   // mana, rage or energy into the wearer's bar
	setBonusProcAuras                      // auras on the wearer for the spell's duration
)

// A row whose one effect triggers another spell off the row's own proc flags.
func isSetBonusProc(row *spelldata.Spell) bool {
	return len(row.Effects) == 1 && row.Effects[0].Aura == dbcenums.A_PROC_TRIGGER_SPELL && row.FiredSpell() != spelldata.Nil
}

func shapeOfSetBonusProc(row *spelldata.Spell) setBonusProcShape {
	fired := row.FiredSpell()
	switch {
	case fired.ProcDamageEffect() != spelldata.NilEffect && fired.PeriodicDamageEffect() == spelldata.NilEffect:
		return setBonusProcDamage
	case fired.EnergizeEffect() != spelldata.NilEffect:
		return setBonusProcEnergize
	case !slices.ContainsFunc(fired.Effects, func(e spelldata.Effect) bool { return !spelldata.AppliesAura(e.Type) }):
		return setBonusProcAuras
	}
	return setBonusProcUnbuilt
}

// Why a generated set cannot register the row, or nil where it can. A proc row registers when its
// listener and rate are ones an item's proc takes, and the spell it fires deals damage at once, fills the
// wearer's bar or puts auras on the wearer alone. Any other row is the parse's to state.
//
// A proc row of a class family whose proc effect names none of the family's spells is refused too: the
// spells its tooltip names are not in the row, so its listener would hear every spell (Symbols of
// Unending Life's 26107 refunds on finishing moves that miss, Stormcaller's Garb's 26119 on three shaman
// spells).
func SetBonusUnsupported(row *spelldata.Spell) []string {
	if !isSetBonusProc(row) {
		return spelldata.SetBonusUnsupported(row)
	}

	unsupported := spelldata.ItemProcUnsupportedWithRate(row, false, 0, 0)
	if row.ClassFlags.Family != 0 && row.Effects[0].ClassFlags.IsZero() {
		unsupported = append(unsupported, fmt.Sprintf("names class family %d and none of its spells", row.ClassFlags.Family))
	}
	fired := row.FiredSpell()
	switch shapeOfSetBonusProc(row) {
	case setBonusProcDamage:
	case setBonusProcEnergize:
		unsupported = append(unsupported, FiredSpellUnsupported(row)...)
	case setBonusProcAuras:
		unsupported = append(unsupported, spelldata.ItemAuraUnsupported(fired, false)...)
		if len(spelldata.EffectsOn(fired, spelldata.AuraOnEnemy)) > 0 {
			unsupported = append(unsupported, fmt.Sprintf("%d puts auras on an enemy", fired.ID))
		}
	default:
		unsupported = append(unsupported, fmt.Sprintf("%d fires %d, which neither deals damage at once, fills a bar nor only applies auras", row.ID, fired.ID))
	}
	return unsupported
}

// The proc of a row isSetBonusProc names, kept to the set bonus aura: the row's listener, chance and
// cooldown, firing what its spell does.
func attachSetBonusProc(character *core.Character, setBonusAura *core.Aura, row *spelldata.Spell) {
	fired := row.FiredSpell()
	trigger := spelldata.ProcTrigger(character, row, nil, spelldata.ItemProcChance(row))
	switch shapeOfSetBonusProc(row) {
	case setBonusProcDamage:
		attachProcDamage(character, &trigger, spellDataProcDamageSpell(character, fired, true))
	case setBonusProcEnergize:
		trigger.Handler = FiredSpellHandler(character, trigger.ActionID, row)
		if trigger.Handler == nil {
			// The wearer has no bar the gain fills.
			return
		}
	case setBonusProcAuras:
		auras := newSpellDataAuras(character, fired, spelldata.AuraConfig(fired))
		callback := trigger.Callback
		trigger.Handler = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			auras.activate(sim, procDamageTarget(character, callback, spell, result))
		}
	default:
		panic(fmt.Sprintf("set bonus %d fires %d, which no set bonus proc builds", row.ID, fired.ID))
	}
	setBonusAura.AttachProcTrigger(trigger)
}

// A set gen_db registers from the game files, each piece count's bonus parsed from its rows. A set
// written by hand under the same id keeps it off.
func NewParsedItemSet(id int32, name string, slots []proto.ItemSlot, bonuses map[int32][]int32) {
	NewParsedProfessionItemSet(id, name, proto.Profession_ProfessionUnknown, slots, bonuses)
}

// A generated set whose ItemSet row gates the bonuses on a profession (RequiredSkill).
func NewParsedProfessionItemSet(id int32, name string, profession proto.Profession, slots []proto.ItemSlot, bonuses map[int32][]int32) {
	if slices.ContainsFunc(core.ItemSets(), func(set *core.ItemSet) bool { return set.ID == id }) {
		return
	}

	set := core.ItemSet{ID: id, Name: name, Slots: slots, RequiredProfession: profession, Bonuses: map[int32]core.ApplySetBonus{}}
	for pieces, spellIDs := range bonuses {
		set.Bonuses[pieces] = ParsedSetBonus(core.MapSlice(spellIDs, spelldata.MustFind)...)
	}
	core.NewItemSet(set)
}

// The eight armor slots of a crafted or dungeon set.
var EightPieceSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotHead,
	proto.ItemSlot_ItemSlotShoulder,
	proto.ItemSlot_ItemSlotChest,
	proto.ItemSlot_ItemSlotWrist,
	proto.ItemSlot_ItemSlotHands,
	proto.ItemSlot_ItemSlotWaist,
	proto.ItemSlot_ItemSlotLegs,
	proto.ItemSlot_ItemSlotFeet,
}

// A Naxxramas tier 3 set: the eight armor pieces and the ring, all worn only by the set's class.
var TierThreeSlots = slices.Concat(EightPieceSlots, []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2})
