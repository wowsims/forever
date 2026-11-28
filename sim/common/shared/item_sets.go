package shared

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// A bonus whose spells the parse states whole. A spell's modifiers name its class's spells by class
// flags, so on another class they reach nothing.
func ParsedSetBonus(rows ...*spelldata.Spell) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		character := agent.GetCharacter()
		for _, row := range rows {
			if IsSetBonusProc(row) {
				attachSetBonusProc(character, setBonusAura, row)
				continue
			}
			spelldata.ParseEffects(character, setBonusAura, row)
		}
		setBonusAura.ExposeToAPL(rows[0].ID)
	}
}

// What a set bonus proc fires, read off the spell it names.
type setBonusProcShape uint8

const (
	setBonusProcUnbuilt  setBonusProcShape = iota
	setBonusProcDamage                     // damage at once, a spell cast at the proc's target
	setBonusProcEnergize                   // mana, rage or energy into the wearer's bar
	setBonusProcAuras                      // auras on the wearer for the spell's duration
)

// A spell whose one effect triggers another spell off its own proc flags.
func IsSetBonusProc(row *spelldata.Spell) bool {
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

// Why a generated set cannot register the spell, or nil where it can. A proc spell of a class family whose
// proc effect names none of the family's spells is refused: the spells its tooltip names are not in the
// spell, so its listener would hear every spell (Symbols of Unending Life's 26107 refunds on finishing
// moves that miss).
func SetBonusUnsupported(row *spelldata.Spell) []string {
	if !IsSetBonusProc(row) {
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

// The proc of a spell IsSetBonusProc names, kept to the set bonus aura. A spell whose tooltip names the cast
// as a trigger and that hears melee swings and spells at procs per minute gets a listener for each. The
// pair shares no cooldown, and no such spell states one.
func attachSetBonusProc(character *core.Character, setBonusAura *core.Aura, row *spelldata.Spell) {
	fired := row.FiredSpell()
	castsRollTheRate := false
	trigger := spelldata.ProcTrigger(character, row, nil, spelldata.ItemProcChance(row),
		func(_ *core.Character, trigger *core.ProcTrigger) {
			if row.RPPM > 0 && row.ProcHint.Matches(core.ProcHintCastTrigger) &&
				trigger.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && trigger.ProcMask.Matches(core.ProcMaskSpellDamage) {
				castsRollTheRate = true
				// A spell's cast rolls the rate, so the swing listener hears neither its hit nor a heal's landing.
				trigger.ProcMask &^= core.ProcMaskSpellDamage | core.ProcMaskSpellHealing
			}
		})
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
		auras := newSpellDataAuras(character, fired, spelldata.AuraConfig(fired), 0)
		callback := trigger.Callback
		trigger.Handler = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			auras.activate(sim, procDamageTarget(character, callback, spell, result))
		}
	default:
		panic(fmt.Sprintf("set bonus %d fires %d, which no set bonus proc builds", row.ID, fired.ID))
	}
	setBonusAura.AttachProcTrigger(trigger)
	if castsRollTheRate {
		setBonusAura.AttachProcTrigger(SpellCastPPMTrigger(character, row, trigger.Handler))
	}
}

// A listener for the spell's procs per minute on spell casts. A cast of a damaging or healing spell rolls
// once, and so does any other cast whose spell costs mana. Any other cast never rolls (EV-0144). The
// chance depends on the cast, so Chance(1) leaves core no roll of its own and the condition draws it.
func SpellCastPPMTrigger(character *core.Character, row *spelldata.Spell, handler core.ProcHandler) core.ProcTrigger {
	trigger := spelldata.ProcTrigger(character, row, handler, spelldata.Chance(1))
	trigger.Name = row.Name + " - Cast"
	trigger.Callback = core.CallbackOnCastComplete
	trigger.ProcMask = core.ProcMaskUnknown
	trigger.ProcMaskExclude = core.ProcMaskMeleeOrRanged
	trigger.Outcome = core.OutcomeEmpty
	trigger.RequireDamageDealt = false
	flags := trigger.ProcSpellClassFlags
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		if !spell.ProcMask.Matches(core.ProcMaskSpellDamage|core.ProcMaskSpellHealing) && !costsMana(spell) {
			return false
		}
		return RollCastTimePPM(sim, character, row, flags, spell.CurCast.CastTime)
	}
	return trigger
}

// Whether the spell costs its caster mana before cost modifiers.
func costsMana(spell *core.Spell) bool {
	if spell.Cost == nil || spell.Cost.BaseCost <= 0 {
		return false
	}
	_, ok := spell.Cost.ResourceCostImpl.(*core.ManaCost)
	return ok
}

// Whether a cast that took castTime procs the spell's procs per minute: the rate on that time, at least
// the global cooldown's 1.5 s. The character's proc chance mods for flags raise it.
func RollCastTimePPM(sim *core.Simulation, character *core.Character, row *spelldata.Spell, flags core.ClassFlags, castTime time.Duration) bool {
	chance := core.CastTimePPMChance(float64(row.RPPM), castTime)
	return sim.Proc(character.ModifiedProcChance(flags, chance), row.Name)
}

// A set gen_db registers from the game files, each piece count's bonus parsed from its spells. A set
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
