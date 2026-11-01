package database

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"iter"
	"maps"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/database/overrides"
	"github.com/wowsims/forever/tools/tooltip"
)

// Sets the minimum itemlevel that should be considered for this expansions
const MIN_EFFECT_ILVL = 40

// Whether the item's effects are generated: it is at or above MIN_EFFECT_ILVL, or EffectFloorExemptions
// names it.
func aboveEffectFloor(parsed *proto.UIItem) bool {
	_, exempt := EffectFloorExemptions[parsed.Id]
	return exempt || parsed.ScalingOptions[0].Ilvl >= MIN_EFFECT_ILVL
}

// Whether the item's effect is written by hand: registered as an item effect, or applied by core where
// it reads the item, as Alchemist's Stone's bonus to the potions its wearer drinks.
func implementedByHand(itemID int32) bool {
	return core.HasItemEffect(itemID) || slices.Contains(core.AlchStoneItemIDs, itemID)
}

// Enchantment IDs at or below this are not generated.
const MIN_ENCHANT_EFFECT_ID = 0

func isGeneratableEnchant(effectID int32) bool {
	return effectID > MIN_ENCHANT_EFFECT_ID
}

type ProcInfo struct {
	Outcome             core.HitOutcome
	Callback            core.AuraCallback
	ProcMask            core.ProcMask
	MaxCumulativeStacks int32
	RequireDamageDealt  bool
	ClassSpellsOnly     bool
	// See core.ProcTrigger.
	RequireDamageComponent bool
	// The listener's Can Proc From Procs attribute: it also hears hits from proc spells.
	CanProcFromProcs bool
	// A "Chance on hit" item effect or a combat enchant, cast by the game off every eligible
	// weapon hit regardless of proc flags. Decided by the trigger type, not by the item slot.
	IsWeaponProc bool
	// An aura proc carrying Aura Is Weapon Proc, which skips hits that suppress weapon procs.
	HonoursWeaponProcSuppression bool
}

// A weapon proc ignores proc-ness and already skips hits that suppress weapon procs, so the two
// aura-side fields would be dead or redundant next to it. Cleared so the generated file states
// one rule per listener.
func (info *ProcInfo) setIsWeaponProc(isWeaponProc bool) {
	info.IsWeaponProc = isWeaponProc
	if isWeaponProc {
		info.CanProcFromProcs = false
		info.HonoursWeaponProcSuppression = false
	}
}

// Entry represents a effect with its Item ID, Spell ID and display name.
type Variant struct {
	ID      int
	SpellID int
	Name    string
}

type Entry struct {
	Variants  []*Variant
	Tooltip   []string
	ProcInfo  ProcInfo
	Supported bool
	// What adds a stack while the window is open, for the trinkets whose effect carries a
	// separate accumulating aura. Nil for everything else.
	StackProcInfo *ProcInfo
	// Set for an on-use whose window accumulates a separate aura, which needs the stacking
	// helper rather than the flat one. Carries what that helper cannot read from the database.
	StackingOnUse *StackingOnUse
	// Set for effects an ignore list deliberately excludes. These emit a comment only, so that
	// skipping them is visible in the generated file rather than silent.
	Skipped bool
	// Set when the proc is resolved from the game files' own rows at run time, which is every proc
	// but the two shapes that need more than the rows state: a window that accumulates a second
	// aura, and an effect a hand-written constructor already covers.
	Proc *ProcRouting
	// What a registered on-use leaves out, stated beside its call.
	NotSimulated string
}

// The literals a stacking on-use needs in the generated call. Everything else - stacks,
// per-stack stats, which aura is which - the helper reads from the database at runtime.
type StackingOnUse struct {
	Name       string
	CooldownMs int32
}

// The two spells an item or enchant proc is resolved from at run time, and what the rows say the
// sim cannot model. The reasons are the ones sim/core/spelldata answers, so the registration the
// generator writes and the listener the sim builds come from the same reading.
type ProcRouting struct {
	TriggerSpellID int
	// Zero where the trigger's own row is the buff.
	BuffSpellID int
	// A "Chance on hit" item effect or a combat enchant, cast by the game off every eligible weapon
	// hit whatever the row's proc flags say.
	IsWeaponProc bool
	// A weapon's own procs-per-minute rate, stated on its item effect; zero where the rows state one.
	ItemPPM float64
	Shape   ProcShape
	// Empty when the rows state enough to build the listener.
	Unsupported []string
	// What the rows resolve to, for the reader of the generated file.
	Summary string
}

type ProcShape uint8

const (
	// A buff granting stats, or on an on-use none of the shapes below.
	ShapeStats ProcShape = iota
	// Deals damage instead of granting an aura: there is no buff to build.
	ShapeDamage
	// Heals the wearer.
	ShapeHeal
	// Shields the wearer with an absorb.
	ShapeAbsorb
	// An on-use that raises the wearer's speeds.
	ShapeSpeed
	// Restores the wearer's mana, rage or energy.
	ShapeEnergize
	// Puts a debuff on the enemy it lands on.
	ShapeDebuff
	// Applies auras on the wearer, its pets or an enemy.
	ShapeAura
	// An equip spell's auras, kept up for as long as the item is worn rather than applied by a proc.
	ShapeEquip
	// Grants the wearer extra attacks: main-hand swings under the granting spell's name.
	ShapeExtraAttack
	// An on-use aura on the wearer that casts area damage on every enemy each period.
	ShapePeriodicDamage
)

var procShapes = [...]struct {
	procConstructor  string
	onUseConstructor string
	onUseGroup       string
	unsupported      func(*spelldata.Spell) []string
	alwaysNamesBuff  bool
}{
	ShapeStats:          {"NewSpellDataProc", "NewSimpleStatActive", "", nil, false},
	ShapeDamage:         {"NewSpellDataDamageProc", "NewSpellDataDamageOnUse", "Damage", damageUnsupported, true},
	ShapeHeal:           {"NewSpellDataHealProc", "NewSpellDataHealOnUse", "Heals", healUnsupported, true},
	ShapeAbsorb:         {"NewSpellDataAbsorbProc", "NewSpellDataAbsorbOnUse", "Absorbs", absorbUnsupported, true},
	ShapeSpeed:          {"NewSpellDataProc", "NewSpellDataSpeedOnUse", "Speed", nil, false},
	ShapeEnergize:       {"NewSpellDataProc", "NewSpellDataEnergizeOnUse", "Resources", nil, false},
	ShapeDebuff:         {"NewSpellDataDebuffProc", "NewSimpleStatActive", "", debuffUnsupported, false},
	ShapeAura:           {"NewSpellDataAuraProc", "NewSpellDataAuraOnUse", "Auras", procAuraUnsupported, false},
	ShapeEquip:          {"NewSpellDataEquipAura", "NewSimpleStatActive", "", nil, false},
	ShapeExtraAttack:    {"NewSpellDataExtraAttackProc", "NewSimpleStatActive", "", extraAttackUnsupported, false},
	ShapePeriodicDamage: {"NewSpellDataProc", "NewSpellDataPeriodicDamageOnUse", "Damage", nil, false},
}

func (r *ProcRouting) Supported() bool {
	return len(r.Unsupported) == 0
}

// Renders the reasons as the generated file states them, one clause per shape.
func (r *ProcRouting) Reason() string {
	return strings.Join(r.Unsupported, "; ")
}

// The rows behind an item effect, by the two ids the sim will look up: the spell carrying the proc
// and the spell it applies. A weapon's own procs-per-minute rate, stated on its item effect, answers
// a trigger whose row states none.
func routeProc(triggerSpellID int, buffSpellID int, isWeaponProc bool, itemPPM float64) *ProcRouting {
	routing := &ProcRouting{TriggerSpellID: triggerSpellID, IsWeaponProc: isWeaponProc, ItemPPM: itemPPM}
	if buffSpellID != triggerSpellID {
		routing.BuffSpellID = buffSpellID
	}

	trigger := spelldata.Find(int32(triggerSpellID))
	routing.Unsupported = spelldata.ItemProcUnsupportedWithRate(trigger, isWeaponProc, itemPPM)
	routing.Summary = routing.summarize(buffSpellID)

	return routing
}

func (r *ProcRouting) summarize(buffSpellID int) string {
	return procSummary(r.TriggerSpellID, spelldata.Find(int32(r.TriggerSpellID)), buffSpellID, r.ItemPPM)
}

// A proc that grants no stats: it casts or applies spellID, which the game files hang below the
// trigger or state on it, and whose row says what the sim cannot model.
func (r *ProcRouting) as(shape ProcShape, spellID int32) {
	r.Shape = shape
	r.BuffSpellID = int(spellID)
	if !procShapes[shape].alwaysNamesBuff && r.BuffSpellID == r.TriggerSpellID {
		r.BuffSpellID = 0
	}
	r.Unsupported = append(r.Unsupported, procShapes[shape].unsupported(spelldata.Find(spellID))...)
	r.Summary = r.summarize(int(spellID))
}

func pickNoStatShape(spellID int, debuffID int32) (ProcShape, int32) {
	if damage := dbc.ResolveDamageEffect(spellID); damage != nil {
		return ShapeDamage, int32(damage.SpellID)
	}
	if dot := dbc.ResolveTriggered(spellID, dealsPeriodicDamage); dot != 0 {
		return ShapeDamage, dot
	}
	if heal := dbc.ResolveTriggered(spellID, heals); heal != 0 {
		return ShapeHeal, heal
	}
	if absorb := dbc.ResolveTriggered(spellID, absorbs); absorb != 0 {
		return ShapeAbsorb, absorb
	}
	if spelldata.Find(int32(spellID)).ExtraAttacksEffect() != spelldata.NilEffect {
		return ShapeExtraAttack, int32(spellID)
	}
	if spelldata.Find(debuffID).DebuffsTheTarget() {
		return ShapeDebuff, debuffID
	}
	return ShapeStats, 0
}

func extraAttackUnsupported(row *spelldata.Spell) []string {
	var unsupported []string
	effect := row.ExtraAttacksEffect()
	if effect.Target[0] != dbcenums.TARGET_UNIT_CASTER {
		unsupported = append(unsupported, "the extra attacks are not the wearer's")
	}
	if effect.BasePoints < 1 {
		unsupported = append(unsupported, "states no attack count")
	}
	return unsupported
}

// A row dealing damage over time alone has to tick on an enemy: Blood Talon's Rend 13318 bleeds the
// target every 3 s for 30 s.
func damageUnsupported(damage *spelldata.Spell) []string {
	periodic := damage.PeriodicDamageEffect()
	switch {
	case damage == spelldata.Nil:
		return []string{"the damage spell has no row in the store"}
	case damage.DamageEffect() != spelldata.NilEffect:
		return nil
	case periodic == spelldata.NilEffect:
		return []string{"the damage spell's row states no damage"}
	}

	var unsupported []string
	if !periodic.HitsAnEnemy() {
		unsupported = append(unsupported, fmt.Sprintf("the damage lands on implicit target %d, not an enemy", periodic.Target[0]))
	}
	return append(unsupported, tickUnsupported("damage", damage, periodic)...)
}

func debuffUnsupported(debuff *spelldata.Spell) []string {
	if debuff.DurationMs <= 0 {
		return []string{"the debuff states no duration"}
	}
	return nil
}

func procAuraUnsupported(buff *spelldata.Spell) []string {
	return spelldata.ItemAuraUnsupported(buff, false)
}

func healUnsupported(heal *spelldata.Spell) []string {
	return healTargetUnsupported(heal)
}

// The same, where the heal may also land on one of the targets named, which reach the wearer: an
// on-use heal cast on a friendly target is cast on the wearer, the one friendly unit the sim has.
func healTargetUnsupported(heal *spelldata.Spell, wearer ...dbcenums.ImplicitTarget) []string {
	effect := heal.ProcHealEffect()
	unsupported := wearerTarget("heal", effect, wearer...)
	if effect.Aura == dbcenums.A_PERIODIC_HEAL {
		unsupported = append(unsupported, tickUnsupported("heal", heal, effect)...)
	}
	return unsupported
}

func wearerTarget(what string, effect *spelldata.Effect, wearer ...dbcenums.ImplicitTarget) []string {
	if effect.Target[0] == dbcenums.TARGET_UNIT_CASTER || slices.Contains(wearer, effect.Target[0]) {
		return nil
	}
	return []string{fmt.Sprintf("the %s lands on implicit target %d, not the wearer", what, effect.Target[0])}
}

func tickUnsupported(what string, s *spelldata.Spell, effect *spelldata.Effect) []string {
	if effect.PeriodMs > 0 && s.DurationMs > 0 {
		return nil
	}
	return []string{fmt.Sprintf("the %s over time states no period or no duration to tick over", what)}
}

func energizeUnsupported(s *spelldata.Spell, effect *spelldata.Effect) []string {
	unsupported := wearerTarget("gain", effect)
	if !shared.EnergizesABar(dbcenums.PowerType(effect.Misc)) {
		unsupported = append(unsupported, fmt.Sprintf("the gain fills power type %d, not mana, rage or energy", effect.Misc))
	}
	if effect.Aura == dbcenums.A_PERIODIC_ENERGIZE {
		unsupported = append(unsupported, tickUnsupported("gain", s, effect)...)
	}
	return unsupported
}

func absorbUnsupported(absorb *spelldata.Spell) []string {
	effect := absorb.AbsorbEffect()
	unsupported := wearerTarget("absorb", effect)
	if reason, ok := UnsupportedAbsorbBySpellID[absorb.ID]; ok {
		unsupported = append(unsupported, reason)
	}
	if absorb.DurationMs == 0 {
		unsupported = append(unsupported, "the absorb states no duration")
	}
	breakDamage, breaks := shared.AbsorbBreakDamage(absorb)
	if breaks {
		unsupported = append(unsupported, damageUnsupported(spelldata.Find(breakDamage))...)
	}
	for _, id := range shared.AbsorbLinkedDamage(absorb) {
		if !breaks || id != breakDamage {
			unsupported = append(unsupported, fmt.Sprintf("the damage of %d (%s) the absorb's row names is not simulated", id, spelldata.Find(id).Name))
		}
	}
	return unsupported
}

func dealsPeriodicDamage(s *spelldata.Spell) bool {
	return s.PeriodicDamageEffect() != spelldata.NilEffect
}

func heals(s *spelldata.Spell) bool {
	return s.ProcHealEffect() != spelldata.NilEffect
}

func absorbs(s *spelldata.Spell) bool {
	return s.AbsorbEffect() != spelldata.NilEffect
}

// The buff the proc applies has to last for something: an aura of no duration is one the sim
// refuses to activate, and the game files state it on either row.
func (r *ProcRouting) requireABuffDuration() {
	trigger := spelldata.Find(int32(r.TriggerSpellID))
	buff := trigger
	if r.BuffSpellID != 0 {
		buff = spelldata.Find(int32(r.BuffSpellID))
	}

	if buff == spelldata.Nil {
		r.Unsupported = append(r.Unsupported, "the buff has no row in the store")
		return
	}

	if buff.DurationMs == 0 && trigger.DurationMs == 0 {
		r.Unsupported = append(r.Unsupported, "neither row states how long the buff lasts")
	}
}

// What the rows resolve to, as the sim's own constants, so the generated file states the reading
// rather than leaving it to be looked up.
func procSummary(triggerSpellID int, trigger *spelldata.Spell, buffSpellID int, itemPPM float64) string {
	if trigger == spelldata.Nil {
		return fmt.Sprintf("trigger %d is not in the store", triggerSpellID)
	}

	rate := procRateSummary(trigger)
	if itemPPM > 0 {
		rate = fmt.Sprintf("%v ppm on the item", itemPPM)
	}
	decoded := core.DecodeProcTypeMask(trigger.ProcFlags, trigger.ProcHint)
	summary := fmt.Sprintf("trigger %d (%s, %s, %s)", trigger.ID,
		rate, asCoreCallback(decoded.Callback), asCoreProcMask(decoded.ProcMask))

	if int(trigger.ID) != buffSpellID {
		summary += fmt.Sprintf(" -> buff %d", buffSpellID)
	}

	return summary
}

func procRateSummary(trigger *spelldata.Spell) string {
	switch {
	case trigger.RPPM > 0:
		return fmt.Sprintf("%v ppm", trigger.RPPM)
	case trigger.ItemProcRollsTheColumn():
		return fmt.Sprintf("%d%%, the column over effect %d's %v%%", trigger.ProcChance, trigger.ProcChanceEffect, trigger.StatedChance()*100)
	case trigger.ProcChanceSource == spelldata.ProcChanceEffectN:
		return fmt.Sprintf("effect %d's chance", trigger.ProcChanceEffect)
	case trigger.ProcChanceSource == spelldata.ProcChanceAlways:
		return "every time"
	case trigger.ProcChanceSource == spelldata.ProcChancePPM:
		return "no stated rate"
	default:
		return fmt.Sprintf("%d%%", trigger.ProcChance)
	}
}

// Group holds a category of effects.
type Group struct {
	Name    string
	Entries []*Entry
}

type MissingItemEffect struct {
	ItemID  int32
	Name    string
	Effects []Variant
}

var missingEffectsMap = map[string]map[int32]MissingItemEffect{
	"EnchantEffects": {},
	"ItemEffects":    {},
}

type EffectParseResult byte

const (
	EffectParseResultInvalid     EffectParseResult = iota // Returned when the effect is invalid for the current parameters
	EffectParseResultUnsupported                          // Returned when the effect could be parsed but is not supported for effect generation
	EffectParseResultSuccess                              // Returned when the effect was parsed successfuly
	EffectParseResultRefused                              // Returned when the effect was parsed, is not supported, and said so in an entry of its own
)

func GenerateEffectsFile(groups []*Group, outFile string, templateString string) error {
	if _, err := os.Stat(outFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to check file %s: %w", outFile, err)
	}

	// Ensure groups and entries are sorted
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	for _, grp := range groups {
		sort.Slice(grp.Entries, func(i, j int) bool {
			if grp.Entries[i].Supported != grp.Entries[j].Supported {
				return !grp.Entries[i].Supported
			}

			return entryOrder(grp.Entries[i], grp.Entries[j])
		})
	}

	funcMap := map[string]any{
		"asCoreCallback": asCoreCallback,
		"asCoreProcMask": asCoreProcMask,
		"asCoreOutcome":  asCoreOutcome,
		"formatStrings":  formatStrings,
	}
	tmpl := template.Must(template.New("effects").Funcs(funcMap).Parse(templateString))

	// An empty generated file must not import anything: gen_db links sim/common, so unused
	// imports in a file it just wrote break the very build the next run needs.
	hasEntries := false
	for _, grp := range groups {
		if len(grp.Entries) > 0 {
			hasEntries = true
			break
		}
	}

	hasStacking := false
	// A registration resolved from the game files' rows names no core constant, so a file whose live
	// entries are all of that shape must not import core: gen_db links the sim, and an unused import
	// in a file it just wrote breaks the build the next run needs.
	usesCore := false
	// And nothing at all is imported by a file whose every entry is commented out.
	hasLive := false
	for _, grp := range groups {
		for _, entry := range grp.Entries {
			if entry.StackingOnUse != nil {
				hasStacking = true
			}
			if entry.Skipped || !entry.Supported {
				continue
			}
			hasLive = true
			if entry.Proc == nil {
				usesCore = true
			}
		}
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, map[string]interface{}{"Groups": groups, "HasEntries": hasEntries, "HasStacking": hasStacking, "UsesCore": usesCore, "HasLive": hasLive}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// The template cannot indent commented-out blocks or blank lines the way gofmt wants, so
	// format the result. Otherwise every regeneration reverts whatever formatted the file last
	// and the diff is hundreds of whitespace-only lines.
	out := rendered.Bytes()
	if formatted, err := format.Source(out); err != nil {
		fmt.Printf("WARN: generated %s is not valid Go, writing unformatted: %v\n", outFile, err)
	} else {
		out = formatted
	}

	if err := os.WriteFile(outFile, out, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outFile, err)
	}

	return nil
}

// Whether two entries are resolved from the same rows, which is what lets them be emitted as one
// call with a variant list.
func sameProcRows(a *Entry, b *Entry) bool {
	if a.Proc == nil || b.Proc == nil {
		return a.Proc == b.Proc
	}

	return a.Proc.TriggerSpellID == b.Proc.TriggerSpellID && a.Proc.BuffSpellID == b.Proc.BuffSpellID
}

// A total order over entries. Sorting on the item or enchant ID alone is not one: an item with
// two effects yields two entries sharing that ID, and sort.Slice is not stable, so their order
// in the generated file flipped between runs.
func entryOrder(a *Entry, b *Entry) bool {
	if a.Variants[0].ID != b.Variants[0].ID {
		return a.Variants[0].ID < b.Variants[0].ID
	}
	if a.Variants[0].SpellID != b.Variants[0].SpellID || a.Proc == nil || b.Proc == nil {
		return a.Variants[0].SpellID < b.Variants[0].SpellID
	}
	return a.Proc.TriggerSpellID < b.Proc.TriggerSpellID
}

// Escapes a rendered tooltip for use inside a double-quoted TypeScript string. Tooltips
// routinely span several lines, which would otherwise produce a file that does not parse.
func jsString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}

const missingEffectsFileName = "ui/sim/constants/missing_effects_auto_gen.ts"

// A proc rate on gear that the game files do not state and nobody has measured on Forever.
type UnvalidatedRate struct {
	ID     int32 // The item id, the enchant effect id or the set id.
	Pieces int32 // The set bonus's piece count; 0 on an item or an enchant.
	Name   string
	Kind   overrides.RateKind
	Origin string // "vanilla", "tbc" or "unknown".
}

type UnvalidatedRates struct {
	Items      []UnvalidatedRate
	Enchants   []UnvalidatedRate
	SetBonuses []UnvalidatedRate
}

// A rate the sim keeps as a literal in code, outside the override tables, which the notices still
// have to name. Gear is "item", "enchant" or "set bonus".
type literalRate struct {
	Gear   string
	ID     int32
	Pieces int32
	Kind   overrides.RateKind
	Origin string
	Reason string
}

var literalRates = []literalRate{
	{
		Gear:   "item",
		ID:     272432,
		Kind:   overrides.RatePPM,
		Origin: "unknown",
		Reason: "Totem of the Storm: half of Maelstrom Weapon's rate per minute, which the shaman talent " +
			"code keeps as its own literal; the item cannot carry it without the talent reading it from the item",
	},
	{
		Gear:   "set bonus",
		ID:     184,
		Pieces: 5,
		Kind:   overrides.RatePPM,
		Origin: "unknown",
		Reason: "Shadowcraft Armor: the energy proc's row states 100% on every melee hit with no rate per minute, " +
			"so the sim keeps 1 per minute as a literal",
	},
	{
		Gear:   "set bonus",
		ID:     1676,
		Pieces: 4,
		Kind:   overrides.RatePPM,
		Origin: "unknown",
		Reason: "Darkmantle Armor: the energy proc's row states 100% on every melee hit with no rate per minute, " +
			"so the sim keeps 1 per minute as a literal",
	},
}

// The origin a rate's Source names, or "" for one measured on Forever, which needs no notice.
func rateOrigin(source string) string {
	switch {
	case strings.HasPrefix(source, "wcl:"), strings.HasPrefix(source, "issue #"), strings.HasPrefix(source, "combat log, "):
		return ""
	case source == "vanilla-carryover":
		return "vanilla"
	case source == "tbc-carryover":
		return "tbc"
	}
	return "unknown"
}

// The hand-kept rates the notices weigh, and the gear casting each spell a spell rate sits on.
type rateInputs struct {
	WeaponProcs   []overrides.WeaponProc
	Spells        []overrides.Override
	EnchantProcs  []overrides.EnchantProc
	Literals      []literalRate
	SpellItems    map[int32][]int32    // The items casting each spell, by spell id.
	SpellEnchants map[int32][]int32    // The enchant effect ids casting each spell, by spell id.
	SpellSets     map[int32][][2]int32 // The set id and piece count of each set bonus granting the spell, by spell id.
}

// A spell override's rate kind, or 0 for an override that is not a rate.
func spellRateKind(o overrides.Override) overrides.RateKind {
	switch o.Field {
	case overrides.PPM:
		return overrides.RatePPM
	case overrides.ProcChancePct:
		return overrides.RateChance
	}
	return 0
}

// The shipped tables, with the items, enchants and set bonuses the game files have casting each rated spell.
func shippedRateInputs(instance *dbc.DBC) rateInputs {
	in := rateInputs{
		WeaponProcs:   overrides.WeaponProcs,
		Spells:        overrides.Spells,
		EnchantProcs:  overrides.EnchantProcs,
		Literals:      literalRates,
		SpellItems:    map[int32][]int32{},
		SpellEnchants: map[int32][]int32{},
		SpellSets:     map[int32][][2]int32{},
	}
	for _, o := range overrides.Spells {
		if spellRateKind(o) == 0 {
			continue
		}
		for itemID, effects := range instance.ItemEffectsByParentID {
			if slices.ContainsFunc(effects, func(e dbc.ItemEffect) bool { return e.SpellID == int(o.SpellID) }) {
				in.SpellItems[o.SpellID] = append(in.SpellItems[o.SpellID], int32(itemID))
			}
		}
		for effectID, enchant := range instance.EnchantsByEffectId {
			if slices.Contains(enchant.EffectArgs, int(o.SpellID)) {
				in.SpellEnchants[o.SpellID] = append(in.SpellEnchants[o.SpellID], int32(effectID))
			}
		}
		for setID, spells := range instance.ItemSetSpells {
			for _, spell := range spells {
				if spell.SpellID == int(o.SpellID) {
					in.SpellSets[o.SpellID] = append(in.SpellSets[o.SpellID], [2]int32{int32(setID), int32(spell.Threshold)})
				}
			}
		}
	}
	return in
}

// Every hand-kept proc rate reaching an item, an enchant or a set bonus in the database whose Source
// is not a measurement on Forever, once per piece of gear and in id order: the weapon rates, the
// spell rates through the items, enchants and set bonuses that cast the spell, the enchant rates, and the rates
// still kept as literals in code. Two rates of different kind or origin on one piece is an error.
func unvalidatedRates(in rateInputs, db *WowDatabase) (UnvalidatedRates, error) {
	items := map[int32]UnvalidatedRate{}
	enchants := map[int32]UnvalidatedRate{}
	setBonuses := map[[2]int32]UnvalidatedRate{}
	var conflict error
	check := func(listed UnvalidatedRate, ok bool, rate UnvalidatedRate) {
		if ok && (listed.Kind != rate.Kind || listed.Origin != rate.Origin) && conflict == nil {
			conflict = fmt.Errorf("%s (%d) carries two rates, a %s %s one and a %s %s one",
				rate.Name, rate.ID, listed.Origin, listed.Kind, rate.Origin, rate.Kind)
		}
	}
	put := func(list map[int32]UnvalidatedRate, rate UnvalidatedRate) {
		listed, ok := list[rate.ID]
		check(listed, ok, rate)
		list[rate.ID] = rate
	}
	putSetBonus := func(rate UnvalidatedRate) {
		key := [2]int32{rate.ID, rate.Pieces}
		listed, ok := setBonuses[key]
		check(listed, ok, rate)
		setBonuses[key] = rate
	}
	enchantByEffectID := map[int32]*proto.UIEnchant{}
	for _, enchant := range db.Enchants {
		enchantByEffectID[enchant.EffectId] = enchant
	}
	setNames := setNamesOf(db)

	for _, proc := range in.WeaponProcs {
		item, ok := db.Items[proc.ItemID]
		if origin := rateOrigin(proc.Source); origin != "" && ok {
			put(items, UnvalidatedRate{ID: item.Id, Name: item.Name, Kind: overrides.RatePPM, Origin: origin})
		}
	}

	for _, o := range in.Spells {
		kind, origin := spellRateKind(o), rateOrigin(o.Source)
		if kind == 0 || origin == "" {
			continue
		}
		for _, itemID := range in.SpellItems[o.SpellID] {
			if item, ok := db.Items[itemID]; ok {
				put(items, UnvalidatedRate{ID: itemID, Name: item.Name, Kind: kind, Origin: origin})
			}
		}
		for _, effectID := range in.SpellEnchants[o.SpellID] {
			if enchant, ok := enchantByEffectID[effectID]; ok {
				put(enchants, UnvalidatedRate{ID: effectID, Name: enchant.Name, Kind: kind, Origin: origin})
			}
		}
		for _, bonus := range in.SpellSets[o.SpellID] {
			if name, ok := setNames[bonus[0]]; ok {
				putSetBonus(UnvalidatedRate{ID: bonus[0], Pieces: bonus[1], Name: name, Kind: kind, Origin: origin})
			}
		}
	}

	for _, proc := range in.EnchantProcs {
		enchant, ok := enchantByEffectID[proc.EnchantID]
		if origin := rateOrigin(proc.Source); origin != "" && ok {
			put(enchants, UnvalidatedRate{ID: proc.EnchantID, Name: enchant.Name, Kind: proc.Kind, Origin: origin})
		}
	}

	for _, literal := range in.Literals {
		rate := UnvalidatedRate{ID: literal.ID, Pieces: literal.Pieces, Kind: literal.Kind, Origin: literal.Origin}
		switch literal.Gear {
		case "item":
			if item, ok := db.Items[literal.ID]; ok {
				rate.Name = item.Name
				put(items, rate)
			}
		case "enchant":
			if enchant, ok := enchantByEffectID[literal.ID]; ok {
				rate.Name = enchant.Name
				put(enchants, rate)
			}
		case "set bonus":
			if name, ok := setNames[literal.ID]; ok {
				rate.Name = name
				putSetBonus(rate)
			}
		}
	}

	byID := func(list []UnvalidatedRate) []UnvalidatedRate {
		sort.Slice(list, func(i, j int) bool {
			return list[i].ID < list[j].ID || list[i].ID == list[j].ID && list[i].Pieces < list[j].Pieces
		})
		return list
	}
	return UnvalidatedRates{
		Items:      byID(slices.Collect(maps.Values(items))),
		Enchants:   byID(slices.Collect(maps.Values(enchants))),
		SetBonuses: byID(slices.Collect(maps.Values(setBonuses))),
	}, conflict
}

// A set bonus the game files grant at a piece count the sim registers no effect for.
type MissingSetBonus struct {
	ID       int32 // The set id.
	Pieces   int32
	Name     string
	Spells   []int32
	Tooltips []string // One per spell.
}

// Every set bonus of a set with items in the database whose piece count no registered set covers,
// in set id and piece count order. A set registered with ID 0 is matched by name.
func missingSetBonuses(registered []*core.ItemSet, setSpells map[int][]dbc.ItemSetSpell, db *WowDatabase, tooltipOf func(spellID int) string) []MissingSetBonus {
	var missing []MissingSetBonus
	for setID, name := range setNamesOf(db) {
		implemented := map[int32]bool{}
		for _, set := range registered {
			if set.ID == setID || set.ID == 0 && (set.Name == name || set.AlternativeName == name) {
				for pieces := range set.Bonuses {
					implemented[pieces] = true
				}
			}
		}
		byPieces := map[int32]*MissingSetBonus{}
		for _, row := range setSpells[int(setID)] {
			pieces := int32(row.Threshold)
			if implemented[pieces] {
				continue
			}
			bonus, ok := byPieces[pieces]
			if !ok {
				bonus = &MissingSetBonus{ID: setID, Pieces: pieces, Name: name}
				byPieces[pieces] = bonus
			}
			bonus.Spells = append(bonus.Spells, int32(row.SpellID))
			bonus.Tooltips = append(bonus.Tooltips, tooltipOf(row.SpellID))
		}
		for _, bonus := range byPieces {
			missing = append(missing, *bonus)
		}
	}

	sort.Slice(missing, func(i, j int) bool {
		return missing[i].ID < missing[j].ID || missing[i].ID == missing[j].ID && missing[i].Pieces < missing[j].Pieces
	})
	return missing
}

// The name of every set with an item in the database, by set id.
func setNamesOf(db *WowDatabase) map[int32]string {
	names := map[int32]string{}
	for _, item := range db.Items {
		if item.SetId != 0 {
			names[item.SetId] = item.SetName
		}
	}
	return names
}

// A list in set id order, cut into one run per set.
func groupBySet[T any](list []T, setID func(T) int32) [][]T {
	var sets [][]T
	for _, entry := range list {
		if len(sets) == 0 || setID(sets[len(sets)-1][0]) != setID(entry) {
			sets = append(sets, nil)
		}
		sets[len(sets)-1] = append(sets[len(sets)-1], entry)
	}
	return sets
}

// The spell's tooltip as the missing-effects list shows it, or its name where the tooltip does not render.
func setBonusTooltip(instance *dbc.DBC, spellID int) string {
	spell := instance.Spells[spellID]
	if rendered, _ := tooltip.ParseTooltip(spell.Description, tooltip.DBCTooltipDataProvider{DBC: instance}, int64(spellID)); rendered != nil {
		if text := rendered.String(); text != "" {
			return text
		}
	}
	return spell.NameLang
}

func GenerateMissingEffectsFile(instance *dbc.DBC, db *WowDatabase) error {
	if _, err := os.Stat(missingEffectsFileName); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to check file %s: %w", missingEffectsFileName, err)
	}

	rates, err := unvalidatedRates(shippedRateInputs(instance), db)
	if err != nil {
		return err
	}
	setBonuses := missingSetBonuses(core.ItemSets(), instance.ItemSetSpells, db, func(spellID int) string {
		return setBonusTooltip(instance, spellID)
	})

	f, err := os.Create(missingEffectsFileName)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", missingEffectsFileName, err)
	}
	defer f.Close()

	return renderMissingEffects(f, rates, setBonuses)
}

func renderMissingEffects(w io.Writer, rates UnvalidatedRates, setBonuses []MissingSetBonus) error {
	funcMap := map[string]any{
		"asCoreCallback": asCoreCallback,
		"asCoreProcMask": asCoreProcMask,
		"asCoreOutcome":  asCoreOutcome,
		"formatStrings":  formatStrings,
		"jsString":       jsString,
	}
	tmpl := template.Must(template.New("missingEffects").Funcs(funcMap).Parse(TmplStrMissingEffects))
	sets := groupBySet(rates.SetBonuses, func(r UnvalidatedRate) int32 { return r.ID })
	missingSets := groupBySet(setBonuses, func(b MissingSetBonus) int32 { return b.ID })
	data := map[string]any{
		"ItemEffects":               missingEffectsMap["ItemEffects"],
		"EnchantEffects":            missingEffectsMap["EnchantEffects"],
		"MissingSetBonuses":         missingSets,
		"UnvalidatedRateItems":      rates.Items,
		"UnvalidatedRateEnchants":   rates.Enchants,
		"UnvalidatedRateSetBonuses": sets,
	}
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func GenerateEnchantEffects(instance *dbc.DBC, db *WowDatabase) {
	groupMapProc := map[string]Group{}
	enchantSpellEffects := enchantGrantEffects(instance.SpellEffectsById)

	// The raw rows repeat an enchant once per recipe name, and an enchant registers once.
	generated := map[int32]bool{}
	for _, enchant := range instance.Enchants {
		if generated[int32(enchant.EffectId)] {
			continue
		}
		parsed, slots := enchant.ToProtoWithSlots()
		if _, ok := db.Enchants[EnchantToDBKey(parsed)]; !ok {
			continue
		}
		generated[parsed.EffectId] = true

		TryParseEnchantEffect(parsed, slots, groupMapProc, instance, enchantSpellEffects)
		storeUnmappedSpeedEnchant(instance, enchant, parsed, enchantSpellEffects)
		storeWeaponScopedEnchant(instance, enchant, parsed, enchantSpellEffects, groupMapProc)
	}

	var procGroups []*Group
	for _, grp := range groupMapProc {
		procGroups = append(procGroups, &grp)
	}

	GenerateEffectsFile(procGroups, "sim/common/forever/enchants_auto_gen.go", TmplStrEnchant)
}

var speedAuras = []dbcenums.EffectAuraType{
	dbcenums.A_MOD_MELEE_HASTE, dbcenums.A_MOD_MELEE_RANGED_HASTE, dbcenums.A_HASTE_SPELLS,
	dbcenums.A_MOD_MELEE_HASTE_2, dbcenums.A_MOD_RANGED_HASTE_2,
}

// These attack and cast speed auras reach no enchant field, so an enchant whose equip spell states
// one is listed as missing its effect unless it is implemented by hand.
func storeUnmappedSpeedEnchant(instance *dbc.DBC, enchant dbc.Enchant, parsed *proto.UIEnchant, enchantSpellEffects map[int]*dbc.SpellEffect) {
	grant, ok := enchantSpellEffects[int(parsed.EffectId)]
	if !ok || core.HasEnchantEffect(parsed.EffectId) {
		return
	}

	for spellID := range enchantEquipSpellIDs(enchant) {
		for _, spellEffect := range instance.SpellEffectsInOrder(spellID) {
			if spellEffect.EffectType == dbcenums.E_APPLY_AURA && slices.Contains(speedAuras, spellEffect.EffectAura) {
				StoreMissingEffect("EnchantEffects", parsed.Name, Variant{
					ID:      int(parsed.EffectId),
					Name:    renderSpellTooltip(instance, grant.SpellID) + " (an attack or cast speed aura, which no enchant stat carries)",
					SpellID: spellID,
				})
				return
			}
		}
	}
}

// An equip spell whose weapon requirement dbc.EquipSpellFolds leaves to run time reaches no enchant
// field, so the enchant registers the spell's auras as an equip aura, kept up while the enchant is
// worn. One the rows cannot build, and one whose enchant already registers a proc, is listed as
// missing its effect unless it is implemented by hand.
func storeWeaponScopedEnchant(instance *dbc.DBC, enchant dbc.Enchant, parsed *proto.UIEnchant, enchantSpellEffects map[int]*dbc.SpellEffect, groupMapProc map[string]Group) {
	grant, ok := enchantSpellEffects[int(parsed.EffectId)]
	if !ok || core.HasEnchantEffect(parsed.EffectId) {
		return
	}

	for spellID := range enchantEquipSpellIDs(enchant) {
		if dbc.EquipSpellFolds(spellID) {
			continue
		}

		rendered := renderSpellTooltip(instance, grant.SpellID)
		equip := spelldata.Find(int32(spellID))
		row := spelldata.EquipAuraRow(equip)
		if !appliesAnItemAura(row) {
			StoreMissingEffect("EnchantEffects", parsed.Name, Variant{
				ID:      int(parsed.EffectId),
				Name:    rendered + " (a weapon-scoped equip spell whose auras no equip aura builds)",
				SpellID: spellID,
			})
			return
		}

		routing := equipAuraRouting(instance, equip, row)
		if enchantRegisters(groupMapProc, parsed.EffectId) {
			routing.Unsupported = append(routing.Unsupported, "another slot of the enchant registers")
		}

		variant := &Variant{ID: int(parsed.EffectId), Name: parsed.Name, SpellID: grant.SpellID}
		if !appendEquipEntry(groupMapProc, "Enchants", rendered, variant, routing) {
			StoreMissingEffect("EnchantEffects", parsed.Name, Variant{ID: int(parsed.EffectId), Name: rendered, SpellID: spellID})
		}
		return
	}
}

func enchantEquipSpellIDs(enchant dbc.Enchant) iter.Seq[int] {
	return func(yield func(int) bool) {
		for idx, effect := range enchant.Effects {
			if effect == dbc.ITEM_ENCHANTMENT_EQUIP_SPELL && idx < len(enchant.EffectArgs) && !yield(enchant.EffectArgs[idx]) {
				return
			}
		}
	}
}

// Files an equip aura's entry under the group, and answers whether it registers.
func appendEquipEntry(groupMapProc map[string]Group, group string, rendered string, variant *Variant, routing *ProcRouting) bool {
	grp := groupMapProc[group]
	grp.Name = group
	grp.Entries = append(grp.Entries, &Entry{
		Tooltip:   strings.Split(rendered, "\n"),
		Variants:  []*Variant{variant},
		Proc:      routing,
		Supported: routing.Supported(),
	})
	groupMapProc[group] = grp
	return routing.Supported()
}

func enchantRegisters(groupMapProc map[string]Group, effectID int32) bool {
	return slices.ContainsFunc(groupMapProc["Enchants"].Entries, func(entry *Entry) bool {
		return entry.Supported && entry.Variants[0].ID == int(effectID)
	})
}

// The E_ENCHANT_ITEM effect that grants each enchant, by enchant ID. Only that effect names an
// enchant in its first misc value; any other effect's misc value is a school, a stat or a power.
//
// Several spells can grant the same enchant -- an enchant that was re-taught
// by a later expansion's recipe has one spell per version, up to five here.
// Map iteration order is randomized, so keep the lowest spell ID rather than
// letting whichever one is visited last win and churn the generated file. The store's
// loadEnchantGrants keeps the same one.
func enchantGrantEffects(effects map[int]dbc.SpellEffect) map[int]*dbc.SpellEffect {
	grants := map[int]*dbc.SpellEffect{}
	for _, effect := range effects {
		if effect.EffectType != dbcenums.E_ENCHANT_ITEM {
			continue
		}
		enchantID := effect.EffectMiscValues[0]
		if existing, ok := grants[enchantID]; ok && existing.SpellID <= effect.SpellID {
			continue
		}
		grants[enchantID] = &effect
	}
	return grants
}

// Names the ignore-list rule that excluded an effect, for the comment emitted in the generated
// file. Returns "" when nothing excludes it.
func ignoredEffectReason(instance *dbc.DBC, effectID int) string {
	for _, effect := range instance.SpellEffectsInOrder(effectID) {
		if params, ok := IgnoreSpellEffectByAuraType[effect.EffectAura]; ok {
			if len(params) == 0 || slices.Contains(params, effect.EffectMiscValues[0]) {
				return fmt.Sprintf("ignored aura type %d", effect.EffectAura)
			}
		}

		if params, ok := IgnoreSpellEffectBySpellEffectType[effect.EffectType]; ok {
			if len(params) == 0 || slices.Contains(params, effect.EffectMiscValues[0]) {
				return fmt.Sprintf("ignored effect type %d", effect.EffectType)
			}
		}
	}

	if effects := instance.SpellEffectsInOrder(effectID); carriesOnlyAloneIgnoredAuras(effects) {
		return fmt.Sprintf("ignored aura type %d", effects[0].EffectAura)
	}

	return ""
}

func carriesOnlyAloneIgnoredAuras(effects []dbc.SpellEffect) bool {
	for _, effect := range effects {
		if !slices.Contains(IgnoreSpellEffectAloneByAuraType, effect.EffectAura) {
			return false
		}
	}
	return len(effects) > 0
}

// Records an effect excluded by an ignore list so the generated file documents it. Kept in its
// own group: variant merging is per-group, so these cannot affect whether a real effect's
// variant set is emitted live or commented.
func storeSkippedEffect(id int32, name string, buffID int32, instance *dbc.DBC, groupMap map[string]Group) {
	grp, exists := groupMap["Skipped"]
	if !exists {
		grp = Group{Name: "Skipped"}
	}

	buffName := instance.Spells[int(buffID)].NameLang
	grp.Entries = append(grp.Entries, &Entry{
		Skipped:  true,
		Variants: []*Variant{{ID: int(id), Name: name, SpellID: int(buffID)}},
		Tooltip: []string{fmt.Sprintf("%s: %q (%d) - %s",
			name, buffName, buffID, ignoredEffectReason(instance, int(buffID)))},
	})
	groupMap["Skipped"] = grp
}

func ItemEffectIsSupported(instance *dbc.DBC, effectID int) bool {
	supported := true
	if effects, ok := instance.SpellEffects[effectID]; ok {
		for _, effect := range effects {
			if params, ok := IgnoreSpellEffectByAuraType[effect.EffectAura]; ok {
				if len(params) == 0 {
					supported = false
					break
				} else {
					if slices.Contains(params, effect.EffectMiscValues[0]) {
						supported = false
					}
				}
			}

			if params, ok := IgnoreSpellEffectBySpellEffectType[effect.EffectType]; ok {
				if len(params) == 0 {
					supported = false
					break
				} else {
					if slices.Contains(params, effect.EffectMiscValues[0]) {
						supported = false
					}
				}
			}
		}
	}
	return supported && !carriesOnlyAloneIgnoredAuras(instance.SpellEffectsInOrder(effectID))
}

func GenerateItemEffects(instance *dbc.DBC, db *WowDatabase) {
	groupMapOnUse := map[string]Group{}
	groupMapProc := map[string]Group{}

	// Example loop over your items
	for _, parsed := range db.Items {
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)

		for _, itemEffect := range parsed.ItemEffects {
			if !ItemEffectIsSupported(instance, int(itemEffect.BuffId)) {
				// Commented into the generated file rather than dropped. These are deliberately
				// out of scope - summons, teleports, created items - but an item whose only
				// effect is skipped otherwise vanished with no trace, while a sibling marker
				// aura on the same item got reported as missing instead.
				skippedGroup := groupMapProc
				if itemEffect.GetOnUse() != nil {
					skippedGroup = groupMapOnUse
				}
				storeSkippedEffect(parsed.Id, parsed.Name, itemEffect.BuffId, instance, skippedGroup)
				continue
			}

			switch TryParseOnUseEffect(parsed, itemEffect, instance, groupMapOnUse) {
			case EffectParseResultSuccess, EffectParseResultRefused:
				continue
			}

			switch TryParseEquipAuraEffect(parsed, itemEffect, instance, groupMapProc) {
			case EffectParseResultSuccess, EffectParseResultRefused:
				continue
			}

			switch TryParseProcEffect(parsed, itemEffect, instance, groupMapProc) {
			case EffectParseResultSuccess, EffectParseResultRefused:
			default:
				ParseTooltipForMissingEffect(parsed, itemEffect, instance, groupMapProc, "Procs")
			}
		}
	}

	refuseEquipBesideAnotherEffect(groupMapOnUse, groupMapProc)

	// Sorting done in GenerateEffectsFile
	var onUseGroups []*Group
	for _, grp := range groupMapOnUse {
		onUseGroups = append(onUseGroups, &grp)
	}

	// Merge variants
	var procGroups []*Group
	needsStatPostfix := map[string]bool{}
	for _, grp := range groupMapProc {
		newEntries := []*Entry{}
		entryGroupings := map[string]*Entry{}

		// sort entries first to make tooltip generation consistent for variants
		sort.Slice(grp.Entries, func(i, j int) bool {
			return entryOrder(grp.Entries[i], grp.Entries[j])
		})

		for _, entry := range grp.Entries {
			var idx int64 = 0
			added := false

			// Make sure to only group by name and proc mask, each proc mask will create it's own sub group
			// A variant set is emitted as one call, so its members also have to name the same rows:
			// the reissued PvP shields share a name and a buff and carry different triggers.
			for _, group := range entryGroupings {
				if group.Variants[0].Name == entry.Variants[0].Name {
					idx++
					if group.ProcInfo.ProcMask == entry.ProcInfo.ProcMask && sameProcRows(group, entry) {
						group.AddVariant(entry.Variants[0])
						added = true
						break
					}
				}
			}

			if !added {
				groupName := entry.Variants[0].Name
				if idx > 0 {
					needsStatPostfix[groupName] = true
					groupName += "(" + strconv.FormatInt(idx, 10) + ")"
				}

				newEntries = append(newEntries, entry)
				entryGroupings[entry.Variants[0].Name] = entry
			}
		}

		grp.Entries = newEntries
		procGroups = append(procGroups, &grp)
	}

	updateNames := func(entries []*Entry) {
		for _, entry := range entries {
			for _, variant := range entry.Variants {
				if _, ok := needsStatPostfix[variant.Name]; ok {
					item := db.Items[int32(variant.ID)]
					for _, itemEffect := range item.ItemEffects {
						variant.Name += " - " + GetEffectStatString(itemEffect)
					}
				}

				variant.Name += BuildItemNamePostfix(variant.ID, instance)
			}
		}
	}

	// Update Item names
	for _, grp := range onUseGroups {
		updateNames(grp.Entries)
	}

	for _, grp := range procGroups {
		updateNames(grp.Entries)
	}

	GenerateEffectsFile(onUseGroups, "sim/common/forever/stat_bonus_cds_auto_gen.go", TmplStrOnUse)
	GenerateEffectsFile(procGroups, "sim/common/forever/stat_bonus_procs_auto_gen.go", TmplStrProc)
}

// An item registers one effect, and the sim drops every registration after the first, so an equip aura
// beside another of the item's effects that registers would drop it or be dropped: Sulfuras' damage
// shield beside its Fireball proc. The equip aura is refused and listed as missing; the other stays.
func refuseEquipBesideAnotherEffect(groupMapOnUse map[string]Group, groupMapProc map[string]Group) {
	registered := map[int]int{}
	for _, groupMap := range []map[string]Group{groupMapOnUse, groupMapProc} {
		for _, grp := range groupMap {
			for _, entry := range grp.Entries {
				if entry.Supported && !entry.Skipped {
					registered[entry.Variants[0].ID]++
				}
			}
		}
	}

	for _, entry := range groupMapProc["Equip"].Entries {
		variant := entry.Variants[0]
		if !entry.Supported || registered[variant.ID] < 2 {
			continue
		}
		entry.Proc.Unsupported = append(entry.Proc.Unsupported, "another effect of the item registers")
		entry.Supported = false
		StoreMissingEffect("ItemEffects", variant.Name, Variant{ID: variant.ID, Name: strings.Join(entry.Tooltip, "\n"), SpellID: variant.SpellID})
	}
}

func GenerateItemEffectRandomPropPoints(instance *dbc.DBC, db *WowDatabase) {
	for id, allocMap := range instance.RandomPropertiesByIlvl {
		ilvl := int32(id)
		if ilvl < core.MinIlvl || ilvl > core.MaxIlvl {
			continue
		}
		db.ItemEffectRandPropPoints[ilvl] = &proto.ItemEffectRandPropPoints{
			Ilvl:           ilvl,
			RandPropPoints: allocMap[proto.ItemQuality_ItemQualityEpic][0],
		}
	}
}

func BuildItemNamePostfix(itemId int, instance *dbc.DBC) string {
	namePostfix := ""

	if item, ok := instance.Items[itemId]; ok {
		if len(item.NameDescription) > 0 && item.NameDescription != "Heroic" {
			namePostfix += " (" + item.NameDescription + ")"
		}

		if item.Flags1.Has(dbc.HORDE_SPECIFIC) {
			namePostfix += " (Horde)"
		}

		if item.Flags1.Has(dbc.ALLIANCE_SPECIFIC) {
			namePostfix += " (Alliance)"
		}
	}

	return namePostfix
}

func TryParseProcEffect(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMapProc map[string]Group) EffectParseResult {
	if itemEffect.GetProc() != nil && aboveEffectFloor(parsed) {
		// Effect was already manually implemented
		if implementedByHand(parsed.Id) {
			return EffectParseResultSuccess
		}

		renderedTooltip, rendered := renderItemEffectTooltip(parsed, itemEffect, instance)

		grp, exists := groupMapProc["Procs"]
		if !exists {
			grp = Group{Name: "Procs"}
		}

		if rendered {
			entry := Entry{Tooltip: strings.Split(renderedTooltip, "\n"), Variants: []*Variant{{ID: int(parsed.Id), Name: parsed.Name, SpellID: int(itemEffect.BuffId)}}}
			entry.ProcInfo, entry.Supported = BuildProcInfo(parsed, int(itemEffect.BuffId), instance, renderedTooltip)

			entry.StackProcInfo = buildStackProcInfo(itemEffect, instance, renderedTooltip)

			// entry.Supported speaks only for the trigger that opens the window, so the stack side
			// has to be refused separately. Without this the template omits StackCallback,
			// attachStackTrigger early-returns on the empty callback and the stat aura activates at
			// zero stacks with no duration of its own - a trinket whose window opens and never gains
			// a stack, worth nothing and reported as implemented. The on-use path refuses the same
			// shape further down.
			if itemEffect.StackingAura != nil && itemEffect.GetStackProc() != nil && entry.StackProcInfo == nil {
				entry.Supported = false
			}

			// A stat-buff proc carries its two spell ids and nothing else: what it hears, how often
			// and for how long are the rows' to say, at run time, through the same decision this
			// reads here - and that decision is the whole of it, rather than the tooltip reading
			// BuildProcInfo does for the shapes below. The two that need more than the rows state
			// stay where they are: a window accumulating a second aura, and an effect with no stats.
			grantsStats := len(dbc.EffectStats(itemEffect)) > 0
			if itemEffect.StackingAura == nil && grantsStats {
				entry.Proc = routeItemProc(parsed, itemEffect)
				if entry.Proc != nil {
					entry.Proc.requireABuffDuration()
					entry.Supported = entry.Proc.Supported()
				}
			}

			// An effect that resolves no stats may still deal flat damage, heal or shield the wearer,
			// debuff the enemy it lands on, or apply auras the stat path cannot state, each a shape of
			// its own rather than a reason to refuse: there is no buff, so the proc casts the spell the
			// game files hang below its trigger, read from that spell's own row.
			if !grantsStats {
				shape, spellID := pickNoStatShape(int(itemEffect.BuffId), itemEffect.BuffId)
				if shape == ShapeStats && appliesAnItemAura(spelldata.Find(itemEffect.BuffId)) {
					shape, spellID = ShapeAura, itemEffect.BuffId
				}
				if shape != ShapeStats {
					entry.Proc = routeItemProc(parsed, itemEffect)
					if entry.Proc != nil {
						entry.Proc.as(shape, spellID)
						entry.Supported = entry.Proc.Supported()
					}
				}
			}

			if (!grantsStats && entry.Proc == nil) || !entry.Supported {
				StoreMissingEffect("ItemEffects", parsed.Name, Variant{
					ID:      int(parsed.Id),
					Name:    renderedTooltip,
					SpellID: int(itemEffect.BuffId),
				})

				// A proc the rows themselves refuse is emitted with the reason they gave, so the
				// generated file says why rather than leaving a shapeless commented block behind.
				if entry.Proc != nil && !entry.Proc.Supported() {
					grp.Entries = append(grp.Entries, &entry)
					groupMapProc["Procs"] = grp
					return EffectParseResultRefused
				}

				return EffectParseResultUnsupported
			}

			grp.Entries = append(grp.Entries, &entry)
			groupMapProc["Procs"] = grp

			return EffectParseResultSuccess
		} else {
			return EffectParseResultUnsupported
		}
	}

	// check if the item has any kind of proc as we only support stat proc parsing right now
	if effects, ok := instance.ItemEffectsByParentID[int(parsed.Id)]; ok && aboveEffectFloor(parsed) {
		for _, effect := range effects {
			if SpellHasTriggerEffect(effect.SpellID, instance) {
				return EffectParseResultUnsupported
			}
		}
	}

	return EffectParseResultInvalid
}

// The rows an item effect names: the game files' ItemEffect row carries the spell with the proc on it,
// and the shipped entry carries the spell that applies the stats.
func routeItemProc(parsed *proto.UIItem, itemEffect *proto.ItemEffect) *ProcRouting {
	effect := dbc.GetItemEffectForBuffID(int(parsed.Id), int(itemEffect.BuffId))
	if effect == nil {
		return nil
	}

	return routeProc(effect.SpellID, int(itemEffect.BuffId), effect.TriggerType == dbc.ITEM_SPELLTRIGGER_CHANCE_ON_HIT,
		itemEffect.GetProc().GetPpm())
}

// An equip spell with no stats that applies an aura the stat path cannot state, read from the row it
// keeps up. One the rows cannot build is refused in an entry of its own, with the reasons.
func TryParseEquipAuraEffect(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMapProc map[string]Group) EffectParseResult {
	if !aboveEffectFloor(parsed) || implementedByHand(parsed.Id) || len(dbc.EffectStats(itemEffect)) > 0 {
		return EffectParseResultInvalid
	}

	effect := dbc.GetItemEffectForBuffID(int(parsed.Id), int(itemEffect.BuffId))
	if effect == nil || effect.TriggerType != dbc.ITEM_SPELLTRIGGER_ON_EQUIP {
		return EffectParseResultInvalid
	}

	equip := spelldata.Find(itemEffect.BuffId)
	row := spelldata.EquipAuraRow(equip)
	if row.FirstAura(equipAuraKinds...) == spelldata.NilEffect {
		return EffectParseResultInvalid
	}

	routing := equipAuraRouting(instance, equip, row)

	rendered := itemEffectTooltip(parsed, itemEffect, instance)
	variant := &Variant{ID: int(parsed.Id), Name: parsed.Name, SpellID: int(itemEffect.BuffId)}
	if appendEquipEntry(groupMapProc, "Equip", rendered, variant, routing) {
		return EffectParseResultSuccess
	}

	StoreMissingEffect("ItemEffects", parsed.Name, Variant{ID: int(parsed.Id), Name: rendered, SpellID: int(itemEffect.BuffId)})
	return EffectParseResultRefused
}

// The equip aura shape for an item's or an enchant's equip spell, read from the row it keeps up.
func equipAuraRouting(instance *dbc.DBC, equip *spelldata.Spell, row *spelldata.Spell) *ProcRouting {
	routing := &ProcRouting{
		TriggerSpellID: int(equip.ID),
		Shape:          ShapeEquip,
		Unsupported:    spelldata.ItemAuraUnsupported(row, true),
		Summary:        fmt.Sprintf("equip: %d", equip.ID),
	}
	if row != equip {
		routing.Summary += fmt.Sprintf(" keeps %d up", row.ID)
	}
	routing.Summary += fmt.Sprintf(" (%s)", spellEffectKinds(instance, int(row.ID)))
	return routing
}

func TryParseOnUseEffect(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMap map[string]Group) EffectParseResult {
	// Effect was already manually implemented
	if implementedByHand(parsed.Id) {
		return EffectParseResultSuccess
	}

	if itemEffect.GetOnUse() != nil && aboveEffectFloor(parsed) {
		if itemEffect.GetOnUse().CooldownMs < 0 && itemEffect.GetOnUse().CategoryCooldownMs < 0 {
			return EffectParseResultUnsupported
		}

		if len(dbc.EffectStats(itemEffect)) == 0 || shieldsWithItsStats(parsed, itemEffect, instance) {
			return parseOnUseSpell(parsed, itemEffect, instance, groupMap)
		}

		groupName := GetEffectStatString(itemEffect)
		grp, exists := groupMap[groupName]
		if !exists {
			grp = Group{Name: groupName}
		}

		entry := &Entry{Variants: []*Variant{{ID: int(parsed.Id), Name: parsed.Name, SpellID: int(itemEffect.BuffId)}}, Supported: true}
		grp.Entries = append(grp.Entries, entry)
		groupMap[groupName] = grp

		// A stacking on-use keeps its stats on the accumulating aura, so the flat check below
		// would call it unsupported, and the flat helper would grant nothing.
		stacking := itemEffect.StackingAura
		if stacking != nil && len(stacking.GetScalingOptions()[0].GetStats()) > 0 {
			entry.StackProcInfo = buildStackProcInfo(itemEffect, instance, "")
			if entry.StackProcInfo == nil {
				entry.Supported = false
				return EffectParseResultUnsupported
			}
			entry.StackingOnUse = &StackingOnUse{
				Name:       parsed.Name,
				CooldownMs: itemEffect.GetOnUse().CooldownMs,
			}
			return EffectParseResultSuccess
		}

		if notSimulated := statOnUseLeftOut(itemEffect.BuffId, instance); notSimulated != "" {
			entry.NotSimulated = notSimulated
			StoreMissingEffect("ItemEffects", parsed.Name, Variant{
				ID:      int(parsed.Id),
				Name:    itemEffectTooltip(parsed, itemEffect, instance),
				SpellID: int(itemEffect.BuffId),
			})
		}

		return EffectParseResultSuccess
	}

	return EffectParseResultInvalid
}

// What the flat on-use leaves out of its buff, which grants the buff's stats and nothing else: the proc
// an A_PROC_TRIGGER_SPELL casts (Aegis of Preservation 23780's heal on every hit taken, 23781), and an
// absorb or a damage shield that no shape carries with the stats.
func statOnUseLeftOut(buffID int32, instance *dbc.DBC) string {
	buff := spelldata.Find(buffID)
	var left []string
	if trigger := buff.FirstAura(dbcenums.A_PROC_TRIGGER_SPELL).TriggerID; trigger != 0 {
		left = append(left, fmt.Sprintf("the proc the buff carries, %d (%s)", trigger, spellEffectKinds(instance, int(trigger))))
	}
	for _, aura := range []dbcenums.EffectAuraType{dbcenums.A_SCHOOL_ABSORB, dbcenums.A_DAMAGE_SHIELD} {
		if e := buff.FirstAura(aura); e != spelldata.NilEffect {
			left = append(left, fmt.Sprintf("the buff's %v %v", aura, e.BasePoints))
		}
	}
	return strings.Join(left, "; ")
}

// Whether a stat on-use's buff also states a shield that a shape can carry with its stats: a damage
// shield the item aura shape attaches (Heart of the Scale's 17275 deals 20 Fire to melee attackers
// beside its Fire resistance), or an absorb the absorb shape holds the stats on (Blazing Emblem's 13744
// absorbs 465 Fire beside its 33 Fire resistance).
func shieldsWithItsStats(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC) bool {
	if itemEffect.StackingAura != nil {
		return false
	}
	buff := spelldata.Find(itemEffect.BuffId)
	shape := ShapeAura
	switch {
	case buff.AbsorbEffect() != spelldata.NilEffect:
		shape = ShapeAbsorb
	case buff.FirstAura(dbcenums.A_DAMAGE_SHIELD) == spelldata.NilEffect:
		return false
	}
	routing := routeOnUse(parsed, itemEffect, instance)
	return routing.Shape == shape && routing.Supported()
}

// An on-use with no stats to grant, resolved from the spell it casts. One the sim cannot build from
// that is refused in an entry of its own, with the reason, and listed as missing.
func parseOnUseSpell(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMap map[string]Group) EffectParseResult {
	routing := routeOnUse(parsed, itemEffect, instance)
	entry := &Entry{
		Variants:  []*Variant{{ID: int(parsed.Id), Name: parsed.Name, SpellID: int(itemEffect.BuffId)}},
		Proc:      routing,
		Supported: routing.Supported(),
	}

	groupName := ""
	if entry.Supported {
		groupName = procShapes[routing.Shape].onUseGroup
	}
	grp := groupMap[groupName]
	grp.Name = groupName
	grp.Entries = append(grp.Entries, entry)
	groupMap[groupName] = grp

	if entry.Supported {
		return EffectParseResultSuccess
	}

	if _, ignored := IgnoreMissingEffectBySpellID[int(itemEffect.BuffId)]; !ignored {
		StoreMissingEffect("ItemEffects", parsed.Name, Variant{
			ID:      int(parsed.Id),
			Name:    itemEffectTooltip(parsed, itemEffect, instance),
			SpellID: int(itemEffect.BuffId),
		})
	}
	return EffectParseResultRefused
}

// The spell an on-use casts, read from the store the way the sim reads it: damage on the enemy it is
// used on, a heal, an absorb, mana, rage or energy for the wearer, or a buff raising the wearer's
// speeds. What else the row does - a root, a stun - is left out, and the summary names it.
func routeOnUse(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC) *ProcRouting {
	return routeOnUseSpell(int(parsed.Id), itemEffect.BuffId, instance)
}

func routeOnUseSpell(itemID int, buffID int32, instance *dbc.DBC) *ProcRouting {
	spellID := int(buffID)
	routing := &ProcRouting{TriggerSpellID: spellID}
	s := spelldata.Find(buffID)
	direct, periodic, heal, absorb, energize := s.DamageEffect(), s.PeriodicDamageEffect(), s.ProcHealEffect(), s.AbsorbEffect(), s.ProcEnergizeEffect()

	switch {
	case !castsOnUse(itemID, spellID, instance):
		routing.Unsupported = append(routing.Unsupported, fmt.Sprintf("the item's on-use row does not cast %d itself", spellID))
	case s == spelldata.Nil:
		routing.Unsupported = append(routing.Unsupported, "the spell has no row in the store")
	case direct != spelldata.NilEffect || periodic != spelldata.NilEffect:
		routing.Shape = ShapeDamage
		routing.Summary = onUseSummary(s, direct, periodic)
		if direct != spelldata.NilEffect && direct.Type != dbcenums.E_SCHOOL_DAMAGE {
			routing.Unsupported = append(routing.Unsupported, fmt.Sprintf("the direct damage is %v rather than an amount", direct.Type))
		}
		for _, e := range []*spelldata.Effect{direct, periodic} {
			if e != spelldata.NilEffect && !e.HitsAnEnemy() {
				routing.Unsupported = append(routing.Unsupported, fmt.Sprintf("the damage lands on implicit target %d, not an enemy", e.Target[0]))
			}
		}
		if periodic != spelldata.NilEffect {
			routing.Unsupported = append(routing.Unsupported, tickUnsupported("damage", s, periodic)...)
		}
	case heal != spelldata.NilEffect:
		routing.Shape = ShapeHeal
		routing.Summary = onUseSummary(s, heal)
		routing.Unsupported = append(routing.Unsupported, healTargetUnsupported(s, dbcenums.TARGET_UNIT_TARGET_ALLY)...)
	case absorb != spelldata.NilEffect:
		routing.Shape = ShapeAbsorb
		routing.Summary = onUseSummary(s, append([]*spelldata.Effect{absorb}, statsBesideTheAbsorb(instance, s)...)...)
		routing.Unsupported = append(routing.Unsupported, absorbUnsupported(s)...)
	case energize != spelldata.NilEffect:
		routing.Shape = ShapeEnergize
		routing.Summary = onUseSummary(s, energize)
		routing.Unsupported = append(routing.Unsupported, energizeUnsupported(s, energize)...)
	case len(s.SpeedEffects()) > 0:
		routing.Shape = ShapeSpeed
		routing.Summary = onUseSummary(s, s.SpeedEffects()...)
		if s.DurationMs <= 0 {
			routing.Unsupported = append(routing.Unsupported, "the speed buff states no duration")
		}
	case s.FirstAura(onUseAuraKinds...) != spelldata.NilEffect:
		routing.Shape = ShapeAura
		routing.Summary = onUseSummary(s, itemAuraModelled(s)...)
		routing.Unsupported = append(routing.Unsupported, procAuraUnsupported(s)...)
		if s.ProcCharges > 0 && !shared.SpendsItsCharges(s) {
			routing.Unsupported = append(routing.Unsupported, fmt.Sprintf("the buff drops after ProcCharges %d, which the aura does not spend", s.ProcCharges))
		}
	case triggersDamagePeriodically(s):
		periodic := s.FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL)
		routing.Shape = ShapePeriodicDamage
		routing.Summary = fmt.Sprintf("%s; every %v for %v: %d (%s)", onUseSummary(s, periodic),
			periodic.Period(), s.Duration(), periodic.TriggerID, spellEffectKinds(instance, int(periodic.TriggerID)))
		routing.Unsupported = append(routing.Unsupported, periodicDamageUnsupported(s, periodic)...)
	default:
		routing.Unsupported = append(routing.Unsupported,
			fmt.Sprintf("%d deals no damage and heals no one (%s)", spellID, spellEffectKinds(instance, spellID)))
	}

	return routing
}

// Whether the row's A_PERIODIC_TRIGGER_SPELL casts a row that deals damage: Thunderbrew's Boot Flask's
// 12257 casts 12278, 75 Fire to the enemies in front of the wearer, every 1 s for 5 s.
func triggersDamagePeriodically(s *spelldata.Spell) bool {
	periodic := s.FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL)
	return periodic != spelldata.NilEffect && spelldata.Find(periodic.TriggerID).DamageEffect() != spelldata.NilEffect
}

// The aura ticks on the wearer, and the row it casts deals an amount to every enemy in an area or cone,
// which the shape casts at the current target: a single-target row would leave it unclear which enemy
// a tick picks.
func periodicDamageUnsupported(s *spelldata.Spell, periodic *spelldata.Effect) []string {
	unsupported := append(wearerTarget("periodic aura", periodic), tickUnsupported("damage", s, periodic)...)
	damage := spelldata.Find(periodic.TriggerID).DamageEffect()
	if damage.Type != dbcenums.E_SCHOOL_DAMAGE {
		unsupported = append(unsupported, fmt.Sprintf("the triggered damage is %v rather than an amount", damage.Type))
	}
	if !damage.HitsAnArea() {
		unsupported = append(unsupported, fmt.Sprintf("the triggered damage lands on implicit target %d, not every enemy in an area", damage.Target[0]))
	}
	return unsupported
}

// The row's effects whose amounts the item's effect entry carries as stats, which the absorb on-use
// grants for as long as its shield holds.
func statsBesideTheAbsorb(instance *dbc.DBC, s *spelldata.Spell) []*spelldata.Effect {
	var effects []*spelldata.Effect
	for _, se := range instance.SpellEffectsInOrder(int(s.ID)) {
		if se.GrantsStats() && se.EffectAura != dbcenums.A_PERIODIC_TRIGGER_SPELL {
			effects = append(effects, s.EffectAtIndex(uint8(se.EffectIndex)))
		}
	}
	return effects
}

// Whether the item's own on-use row names the spell, rather than a spell that reaches it through a
// trigger: the sim casts the spell the effect names.
func castsOnUse(itemID int, spellID int, instance *dbc.DBC) bool {
	return slices.ContainsFunc(instance.ItemEffectsByParentID[itemID], func(e dbc.ItemEffect) bool {
		return e.TriggerType == dbc.ITEM_SPELLTRIGGER_ON_USE && e.SpellID == spellID
	})
}

func onUseSummary(s *spelldata.Spell, modelled ...*spelldata.Effect) string {
	var cast, left []string
	for i := range s.Effects {
		e := &s.Effects[i]
		kind := effectKind(e.Type, e.Aura)
		if slices.Contains(modelled, e) {
			cast = append(cast, kind)
		} else {
			left = append(left, kind)
		}
	}

	summary := fmt.Sprintf("on use: %d (%s)", s.ID, strings.Join(cast, ", "))
	if len(left) > 0 {
		summary += fmt.Sprintf("; not simulated: %s", strings.Join(left, ", "))
	}
	return summary
}

// The item effect's tooltip as the missing-effects list shows it, or the spell's name where the
// tooltip does not render.
func itemEffectTooltip(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC) string {
	if rendered, ok := renderItemEffectTooltip(parsed, itemEffect, instance); ok {
		return rendered
	}
	return instance.Spells[int(itemEffect.BuffId)].NameLang
}

func renderItemEffectTooltip(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC) (string, bool) {
	tooltipString, id := dbc.GetItemEffectSpellTooltip(int(parsed.Id), int(itemEffect.BuffId))
	rendered, _ := tooltip.ParseTooltip(tooltipString, tooltip.DBCTooltipDataProvider{DBC: instance}, int64(id))
	if rendered == nil {
		return "", false
	}
	return rendered.String(), true
}

// The constructor an on-use routed from its spell registers through. A refused one that is neither
// damage nor a heal names the stat constructor in its commented call.
func (r *ProcRouting) OnUseConstructor() string {
	return procShapes[r.Shape].onUseConstructor
}

// The constructor an item or enchant proc registers through.
func (r *ProcRouting) ProcConstructor() string {
	return procShapes[r.Shape].procConstructor
}

// The auras the item aura shape is emitted for: the damage, healing, cost, armor, health, stat and threat
// multipliers, the flat damage taken, the hit, crit, block, dodge and parry percents, and skill points, none
// of which an item states as a stat. The percents reach the item stats only through an equip spell that
// folds; an on-use or a weapon-scoped equip spell applies them as an aura.
var itemAuraKinds = []dbcenums.EffectAuraType{
	dbcenums.A_MOD_THREAT,
	dbcenums.A_MOD_DAMAGE_PERCENT_DONE,
	dbcenums.A_MOD_DAMAGE_PERCENT_TAKEN,
	dbcenums.A_MOD_HEALING_DONE_PERCENT,
	dbcenums.A_MOD_POWER_COST_SCHOOL_PCT,
	dbcenums.A_MOD_DAMAGE_TAKEN,
	dbcenums.A_MOD_BASE_RESISTANCE_PCT,
	dbcenums.A_MOD_INCREASE_HEALTH_PERCENT,
	dbcenums.A_MOD_PERCENT_STAT,
	dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE,
	dbcenums.A_MOD_HIT_CHANCE,
	dbcenums.A_MOD_WEAPON_CRIT_PERCENT,
	dbcenums.A_MOD_SPELL_HIT_CHANCE,
	dbcenums.A_MOD_SPELL_CRIT_CHANCE,
	dbcenums.A_MOD_CRIT_PCT,
	dbcenums.A_MOD_BLOCK_PERCENT,
	dbcenums.A_MOD_DODGE_PERCENT,
	dbcenums.A_MOD_PARRY_PERCENT,
	dbcenums.A_MOD_SKILL,
}

// The auras an item's equip spell is emitted as an aura for, beside itemAuraKinds: a damage shield kept
// up while the item is worn. A proc's or an enchant's damage shield row is left out: its proc flags state
// when the shield deals damage, which the proc shapes would read as when it is applied.
var equipAuraKinds = append(slices.Clone(itemAuraKinds), dbcenums.A_DAMAGE_SHIELD)

// The auras an on-use alone is emitted as an aura for, beside equipAuraKinds: the spell modifiers, which
// many equip spells state too (class necks, relics, PvP gloves) where the item aura shape is not wanted,
// and the attack speed an on-use slows an enemy by.
var onUseAuraKinds = append(slices.Clone(equipAuraKinds),
	dbcenums.A_ADD_FLAT_MODIFIER,
	dbcenums.A_ADD_PCT_MODIFIER,
	dbcenums.A_MOD_MELEE_RANGED_HASTE_2,
)

func appliesAnItemAura(s *spelldata.Spell) bool {
	return s.FirstAura(itemAuraKinds...) != spelldata.NilEffect
}

// The effects the item aura shape attaches: the ones it parses on each unit and the damage shields on the
// wearer. A movement speed change is left out.
func itemAuraModelled(s *spelldata.Spell) []*spelldata.Effect {
	effects := spelldata.ItemAuraDamageShields(s)
	for _, target := range []spelldata.AuraTarget{spelldata.AuraOnWearer, spelldata.AuraOnPet, spelldata.AuraOnEnemy} {
		for _, i := range spelldata.ItemAuraParsedOn(s, target) {
			effects = append(effects, s.EffectN(int(i)))
		}
	}
	return effects
}

func TryParseEnchantEffect(enchant *proto.UIEnchant, slots []dbc.EnchantProcSlot, groupMapProc map[string]Group, instance *dbc.DBC, enchantSpellEffects map[int]*dbc.SpellEffect) {
	if len(slots) == 0 || !isGeneratableEnchant(enchant.EffectId) {
		return
	}

	// Effect was already manually implemented
	if core.HasEnchantEffect(enchant.EffectId) {
		return
	}

	enchantingSpell, ok := enchantSpellEffects[int(enchant.EffectId)]
	if !ok {
		return
	}

	renderedTooltip := renderSpellTooltip(instance, enchantingSpell.SpellID)

	grp, exists := groupMapProc["Enchants"]
	if !exists {
		grp = Group{Name: "Enchants"}
	}

	for _, routing := range routeEnchantProcs(slots, instance) {
		grp.Entries = append(grp.Entries, &Entry{
			Tooltip:   strings.Split(renderedTooltip, "\n"),
			Variants:  []*Variant{{ID: int(enchant.EffectId), Name: enchant.Name, SpellID: int(enchantingSpell.SpellID)}},
			Proc:      routing,
			Supported: routing.Supported(),
		})

		if !routing.Supported() {
			StoreMissingEffect("EnchantEffects", enchant.Name, Variant{
				ID:      int(enchant.EffectId),
				Name:    renderedTooltip,
				SpellID: int(enchant.SpellId),
			})
		}
	}
	groupMapProc["Enchants"] = grp
}

// The rows an enchant's procs are resolved from, one per slot. A combat spell and an equip aura that
// apply the same spell are one proc stated twice (Crusader 1900), so one of the two is kept: the
// combat spell, unless only the aura states a rate. An enchant registers once, so where two slots
// could, the first does.
func routeEnchantProcs(slots []dbc.EnchantProcSlot, instance *dbc.DBC) []*ProcRouting {
	routings := make([]*ProcRouting, len(slots))
	for i, slot := range slots {
		routings[i] = routeEnchantSlot(slot, instance)
	}

	for i, combat := range slots {
		if !combat.IsCombatSpell {
			continue
		}
		for j, aura := range slots {
			if aura.IsCombatSpell || routings[j] == nil || aura.AppliesSpellID != combat.AppliesSpellID {
				continue
			}

			kept, dropped := i, j
			if statesNoRate(routings[i]) && !statesNoRate(routings[j]) {
				kept, dropped = j, i
			}
			routings[kept].Summary += fmt.Sprintf("; slot spell %d applies the same spell and is not registered", slots[dropped].SpellID)
			routings[dropped] = nil
			break
		}
	}

	var kept []*ProcRouting
	registers := false
	for _, routing := range routings {
		if routing == nil {
			continue
		}
		if routing.Supported() && registers {
			routing.Unsupported = append(routing.Unsupported, "another slot of the enchant registers")
		}
		registers = registers || routing.Supported()
		kept = append(kept, routing)
	}

	return kept
}

func renderSpellTooltip(instance *dbc.DBC, spellID int) string {
	tooltip, _ := tooltip.ParseTooltip(instance.Spells[spellID].Description, tooltip.DBCTooltipDataProvider{DBC: instance}, int64(spellID))
	return tooltip.String()
}

func statesNoRate(routing *ProcRouting) bool {
	return slices.Contains(routing.Unsupported, spelldata.ReasonStatesNoRate)
}

// One slot's proc. The trigger is the slot's spell; the buff is what the shipped entry says it
// applies, where it resolves stats, and otherwise the spell it deals damage through or the auras the
// slot applies.
func routeEnchantSlot(slot dbc.EnchantProcSlot, instance *dbc.DBC) *ProcRouting {
	applied := slot.AppliesSpellID
	if applied == 0 {
		applied = slot.SpellID
	}

	effect := slot.Effect
	hasStats := effect != nil

	buffSpellID := slot.SpellID
	if hasStats {
		buffSpellID = int(effect.BuffId)
	}

	routing := routeProc(slot.SpellID, buffSpellID, slot.IsCombatSpell, 0)
	trigger := spelldata.Find(int32(slot.SpellID))
	if !slot.IsCombatSpell {
		routing.Unsupported = spelldata.EnchantAuraUnsupported(trigger)
	}

	shape, spellID := ShapeStats, int32(0)
	if !hasStats {
		shape, spellID = pickNoStatShape(slot.SpellID, int32(applied))
		if shape == ShapeStats && appliesAnItemAura(spelldata.Find(int32(applied))) {
			shape, spellID = ShapeAura, int32(applied)
		}
	}

	if hasStats {
		routing.requireABuffDuration()
	} else if shape != ShapeStats {
		routing.as(shape, spellID)
		if mask := spelldata.Find(spellID).TargetCreatureType; shape == ShapeDamage && mask != 0 {
			routing.Unsupported = append(routing.Unsupported,
				fmt.Sprintf("the damage spell hits %s (TargetCreatureType %d) only", creatureTypeNames(mask), mask))
		}
	} else {
		routing.Unsupported = append(routing.Unsupported,
			fmt.Sprintf("the enchant's effect entry resolves no stats from %d (%s)", applied, spellEffectKinds(instance, applied)))
	}

	if !slot.IsCombatSpell && trigger.ProcHint.Matches(core.ProcHintAttackAvoided) {
		decoded := core.DecodeProcTypeMask(trigger.ProcFlags, trigger.ProcHint)
		routing.Summary += fmt.Sprintf("; the enchant's tooltip restricts it to %s", asCoreOutcome(decoded.Outcome))
	}

	return routing
}

// The game files' CreatureType ids, which TargetCreatureType names as 1 << (id - 1). The damage
// proc does not read the restriction.
var creatureTypes = []string{1: "beasts", 2: "dragonkin", 3: "demons", 4: "elementals", 5: "giants",
	6: "undead", 7: "humanoids", 8: "critters", 9: "mechanicals"}

func creatureTypeNames(mask int32) string {
	var names []string
	for id := 1; id <= 32; id++ {
		if mask&(1<<(id-1)) == 0 {
			continue
		}
		if id < len(creatureTypes) {
			names = append(names, creatureTypes[id])
		} else {
			names = append(names, fmt.Sprintf("creature type %d", id))
		}
	}
	return strings.Join(names, " and ")
}

func spellEffectKinds(instance *dbc.DBC, spellID int) string {
	var kinds []string
	for _, effect := range instance.SpellEffectsInOrder(spellID) {
		kinds = append(kinds, effectKind(effect.EffectType, effect.EffectAura))
	}
	return strings.Join(kinds, ", ")
}

func effectKind(typ dbcenums.SpellEffectType, aura dbcenums.EffectAuraType) string {
	if typ == dbcenums.E_APPLY_AURA {
		return aura.String()
	}
	return typ.String()
}

// "Often", "sometimes" and "occasionally" are how a tooltip says its proc has a rate the rows do not
// carry. "No more often than" states a cooldown instead.
var rateWordMatcher = regexp.MustCompile(`(?i)\b(often|sometimes|occasionally)\b`)
var cooldownWordingMatcher = regexp.MustCompile(`(?i)more often than`)

func ParseTooltipForMissingEffect(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMap map[string]Group, groupMapName string) {
	// Effect was already manually implemented
	if implementedByHand(parsed.Id) {
		return
	}

	renderedTooltip, rendered := renderItemEffectTooltip(parsed, itemEffect, instance)
	if !rendered {
		return
	}

	// An item below the floor takes no entry in the generated file: the variant grouping merges
	// entries by item name, and a Supported: false entry can comment out a live registration
	// of the same name above the floor. It is still reported as missing its effect, since the
	// gear picker has no floor.
	if aboveEffectFloor(parsed) {
		grp, exists := groupMap[groupMapName]
		if !exists {
			grp = Group{Name: groupMapName}
		}

		grp.Entries = append(grp.Entries, &Entry{
			Tooltip:   strings.Split(renderedTooltip, "\n"),
			Supported: false,
			Variants: []*Variant{
				{
					ID:      int(parsed.Id),
					Name:    parsed.Name,
					SpellID: int(itemEffect.BuffId),
				},
			},
		})
		groupMap[groupMapName] = grp
	}

	// Flavour auras carry no mechanic worth implementing and only add noise to the
	// report. Suppressing the report only, never the group entry: those
	// Supported: false entries take part in the variant grouping that decides whether
	// a whole variant set is emitted live or commented, so dropping one can flip a
	// real registration.
	if _, ignored := IgnoreMissingEffectBySpellID[int(itemEffect.BuffId)]; ignored {
		return
	}

	StoreMissingEffect("ItemEffects", parsed.Name, Variant{
		ID:      int(parsed.Id),
		Name:    renderedTooltip,
		SpellID: int(itemEffect.BuffId),
	})
}

// A critical strike named as the trigger. The first clause reads the wording that puts something
// after the crit ("critical strikes have a", "critical hits grant"), which is why the character
// after it may not open "by", "for" or "chance" - those three are how a tooltip states a magnitude
// instead. The second reads the trigger stated from the caster's side, "your critical strikes",
// where the plural is what separates it from the magnitudes: those all read "critical strike
// chance", "critical strike damage" or "critical strike rating", in the singular. Some tooltips
// capitalize the term ("Your Critical Strikes with direct damage spells").
var critMatcher = regexp.MustCompile(`critical ([^\s]+|damage,?)( chance)? [^fbc]|[Yy]our [a-z ]{0,20}[Cc]ritical [Ss]trikes`)
var pureHealMatcher = regexp.MustCompile(`healing spells`)

// A trigger clause restricted to spells with a damage component: "your damaging spells", "damage
// spells", "Your next 5 damage or healing spells", "abilities that deal damage".
var damageTriggerMatcher = regexp.MustCompile(`(?i)\bdamaging\b|\bdamage (or healing )?spells?\b|\bdeals? damage\b`)
var hasHealMatcher = regexp.MustCompile(`heal(ing)?[^,]`)
var hasGenericMatcher = regexp.MustCompile(`a spell`)

// A trigger condition stated as an attack outcome. Deliberately matches the condition clause rather
// than the words themselves: "increases your dodge rating" is a stat on hundreds of items, while
// "when one of your spells is resisted" is a trigger. "After a block" is the third clause the
// game files use, on the Battlegear of Wrath parry (23548); "after you dodge" is not in here because
// the only rows wording it that way are Counterattack's, where it gates the button rather than a
// proc. A miss is not in here either: the one row naming one, 456394, means its own attack missing
// rather than an attack on it.
var outcomeConditionMatcher = regexp.MustCompile(`(?i)when .{0,60}?(is|are) resisted|((each|every) time|when|whenever) you (block|dodge|parry)|after a (block|dodge|parry)`)

// The wearer's own attack dodged or parried, named as the trigger: Recovery's "when you are Parried
// or Dodged". The same words also state a magnitude - "reduces the chance for your attacks to be
// dodged or parried" - so only the condition clause counts.
var attackAvoidedMatcher = regexp.MustCompile(`(?i)((each|every) time|when|whenever) (you are|your (melee )?attacks? (is|are)) (dodged|parried)((,? or|,) (dodged|parried))*`)
var attackDodgedMatcher = regexp.MustCompile(`(?i)dodged`)
var attackParriedMatcher = regexp.MustCompile(`(?i)parried`)

func attackAvoidedHints(tooltip string) core.ProcHint {
	var hints core.ProcHint
	for _, clause := range attackAvoidedMatcher.FindAllString(tooltip, -1) {
		if attackDodgedMatcher.MatchString(clause) {
			hints |= core.ProcHintAttackDodged
		}
		if attackParriedMatcher.MatchString(clause) {
			hints |= core.ProcHintAttackParried
		}
	}
	return hints
}

// A tooltip stating that the effect only happens sometimes: "a chance to", "Chance on hit", "Chance on
// harmful spell cast". Where the data pairs that with a 100% rate, the real rate is the one thing the
// data does not carry.
var statedChanceMatcher = regexp.MustCompile(`(?i)chance (to|of|when|on)|has a chance`)

// What a sentence names as feeding the proc. A clause stating a chance says nothing about a rate
// unless it also says what the chance is rolled on or what it does: Force Reactive Disk's "This also
// has a chance of damaging the shield" is the shield's durability, in a sentence of its own, and the
// -ing is what keeps it apart from the damage a proc deals.
var procTriggerClauseMatcher = regexp.MustCompile(`(?i)melee|ranged|attack|swing|strike|cast|spell|whenever|each time|on hit|block|weapon|damage\b`)

// A chance that is the magnitude rather than the rate: "increases the critical effect chance of your
// Lesser Healing Wave", "grants increased chance to Block", "Chance to trigger Overload increased by
// an additional 5%". The word the effect modifies sits within a clause of the chance either way.
var increasedChanceMatcher = regexp.MustCompile(`(?i)increase[sd]?[^.]{0,40}?chance|chance[^.]{0,30}?increase[sd]?`)

// Where one statement ends and the next begins. Crude on purpose: an abbreviation or a decimal
// splits a sentence in two, and neither changes which half carries the trigger.
var sentenceBreak = regexp.MustCompile(`[.!?](?:\s|$)|\r?\n`)

// Whether the tooltip says the effect only happens sometimes and leaves the rate unsaid. A stated
// chance is read from the sentence that states the trigger rather than from the whole text:
// "increases your critical strike chance" is a magnitude on hundreds of spells, and a chance in a
// sentence that names nothing the proc fires on is about something else entirely.
func tooltipStatesAnUnknownRate(description string) bool {
	for _, sentence := range sentenceBreak.Split(description, -1) {
		if !statedChanceMatcher.MatchString(sentence) {
			continue
		}
		if !procTriggerClauseMatcher.MatchString(sentence) {
			continue
		}
		if increasedChanceMatcher.MatchString(sentence) {
			continue
		}

		return true
	}

	return rateWordMatcher.MatchString(description) &&
		rateWordMatcher.MatchString(cooldownWordingMatcher.ReplaceAllString(description, ""))
}

// Wording that names the cast itself as the trigger rather than the spell landing:
// "each time you cast a spell", "chance on successful spellcast", "chance on spell cast".
var castTriggerMatcher = regexp.MustCompile(`(?i)you cast|on spell ?cast|spellcast`)

// A trigger clause restricted to one named ability: "Your Shock spells", "Your Moonfire ability",
// "Your casts of Greater Heal", "Your Shadow Bolt has", "When you cast Flash of Light". The capital
// is what carries the meaning - an unrestricted trigger reads "your spell critical strikes" or
// "each time you cast a spell", with nothing capitalized to name.
// The game files write a conditional list where the item shows one name - Eternal Power's "Your casts
// of $?s2060[Greater Heal]..." - so the "casts of" clause reads the possessive alone and leaves what
// follows to the tooltip renderer. The other two clauses need the capital: "when you cast a spell"
// and "your melee attacks have" name no ability.
var namedAbilityMatcher = regexp.MustCompile(`[Yy]our [A-Z][A-Za-z']*( [A-Z][A-Za-z']*)* (spell|spells|ability|abilities|has|have)` +
	`|[Yy]our casts of` +
	`|[Ww]hen you cast [A-Z]`)

// What core.DecodeProcTypeMask cannot read off the mask: the trigger wording around it. The named
// ability and outcome condition bits are read here as well, for the store's rows, and say nothing
// new to the item generator: BuildSpellProcInfo refuses a tooltip carrying either before it decodes
// anything, so the outcome bit the decode does read never reaches an item.
func procTooltipHints(tooltip string) core.ProcHint {
	var hints core.ProcHint

	if castTriggerMatcher.MatchString(tooltip) {
		hints |= core.ProcHintCastTrigger
	}

	if critMatcher.MatchString(tooltip) {
		hints |= core.ProcHintCrit
	}

	// An unrestricted "a spell" counts as heal evidence too: next to a helpful-spell bit it is the
	// wording of a proc that fires off any spell the character casts, healing included.
	if hasHealMatcher.MatchString(tooltip) || hasGenericMatcher.MatchString(tooltip) {
		hints |= core.ProcHintHeals
	}

	// "damage or healing spells" names healing spells without being restricted to them.
	if damageTriggerMatcher.MatchString(tooltip) {
		hints |= core.ProcHintDamage
	} else if pureHealMatcher.MatchString(tooltip) {
		hints |= core.ProcHintPureHeal
	}

	if namedAbilityMatcher.MatchString(tooltip) {
		hints |= core.ProcHintNamedAbility
	}

	if outcomeConditionMatcher.MatchString(tooltip) {
		hints |= core.ProcHintOutcomeTaken
	}

	hints |= attackAvoidedHints(tooltip)

	return hints
}

// Derives what adds a stack to an accumulating aura, from the container spell rather than from
// the one that opens the window. buff_id is the container by then: the parser rebases the effect
// onto it precisely because that is where the duration and these proc flags live.
func buildStackProcInfo(itemEffect *proto.ItemEffect, instance *dbc.DBC, tooltip string) *ProcInfo {
	if itemEffect.StackingAura == nil || itemEffect.GetStackProc() == nil {
		return nil
	}

	container, ok := instance.Spells[int(itemEffect.BuffId)]
	if !ok {
		return nil
	}

	// The container's own tooltip is what states what adds a stack, and the store row carries its
	// reading along with any hint an override supplies.
	hint := procTooltipHints(tooltip)
	if row := spelldata.Find(container.ID); row != spelldata.Nil {
		hint = row.ProcHint
	}

	info, supported := buildSpellProcInfo(&container, tooltip, proto.ItemType_ItemTypeUnknown, hint)
	if !supported {
		return nil
	}
	return &info
}

func BuildProcInfo(parsed *proto.UIItem, itemEffectID int, instance *dbc.DBC, tooltip string) (ProcInfo, bool) {
	itemEffect := dbc.GetItemEffectForBuffID(int(parsed.Id), itemEffectID)
	if itemEffect == nil {
		return ProcInfo{}, false
	}

	// if we have multiple spells find the first that has a proc aura assigned
	procId := itemEffect.SpellID
	procSpell, ok := instance.Spells[int(procId)]
	if !ok {
		panic(fmt.Sprintf("Could not find proc aura %d spell for item effect %d.\n", procId, parsed.Id))
	}

	// A "Chance on hit" effect is a weapon proc: the game casts it off the hit itself and never
	// consults a ProcTypeMask, so it takes the on-hit shape whatever the proc spell says.
	isWeaponProc := itemEffect.TriggerType == dbc.ITEM_SPELLTRIGGER_CHANCE_ON_HIT

	itemType := proto.ItemType_ItemTypeUnknown
	if isWeaponProc {
		itemType = proto.ItemType_ItemTypeWeapon
	}

	procInfo, supported := BuildSpellProcInfo(&procSpell, tooltip, itemType)
	procInfo.setIsWeaponProc(isWeaponProc)

	if SpellHasDummyEffect(int(procId), instance) {
		return procInfo, false
	}

	return procInfo, supported
}

func BuildSpellProcInfo(procSpell *dbc.Spell, tooltip string, itemType proto.ItemType) (ProcInfo, bool) {
	return buildSpellProcInfo(procSpell, tooltip, itemType, procTooltipHints(tooltip))
}

func buildSpellProcInfo(procSpell *dbc.Spell, tooltip string, itemType proto.ItemType, hint core.ProcHint) (ProcInfo, bool) {
	var info = ProcInfo{
		RequireDamageDealt:  true,
		MaxCumulativeStacks: procSpell.MaxCumulativeStacks,
	}

	requiresOutcome := true
	onHitProc := false

	// On hit proc
	if itemType == proto.ItemType_ItemTypeWeapon {
		onHitProc = true
		info.Callback |= core.CallbackOnSpellHitDealt
		info.ProcMask |= core.ProcMaskUnknown
	}

	if itemType == proto.ItemType_ItemTypeRanged {
		info.Callback |= core.CallbackOnSpellHitDealt
		info.ProcMask |= core.ProcMaskRanged
	}

	if procSpell.OnlyProcsFromClassAbilities() {
		info.ClassSpellsOnly = true
	}

	// A spell-family filter the generated proc cannot reproduce. Two sources of evidence for the
	// same thing: the spell names a family in SpellClassMask, or its tooltip names one. On item
	// procs TBC stores the mask as either nil or all-zero and keeps the real filter server-side,
	// so for those the tooltip is the only evidence there is - "Your Shock spells", "Your Moonfire
	// ability". Generating one anyway would proc it off every spell instead of that one.
	if slices.ContainsFunc(procSpell.SpellClassMask, func(mask int) bool { return mask != 0 }) ||
		namedAbilityMatcher.MatchString(tooltip) {
		return info, false
	}

	// An outcome the game files do not record. ProcTypeMask has no dodge, parry, block or resist bit -
	// TrinityCore keeps those in a HitMask of its own - so a tooltip stating one as the trigger
	// condition is the only evidence there is. Eye of Magtheridon fires on a resisted spell, and
	// without this it generates as a 100%-per-cast buff. Crit is deliberately not in here: it is the
	// one outcome the generated shape can express, and critMatcher below assigns it.
	if outcomeConditionMatcher.MatchString(tooltip) {
		return info, false
	}

	// A buff the game spends by charges rather than by time: World Breaker's 900 crit rating is
	// consumed by the next two melee hits, which is ProcCharges 2 on 36111. The generated shape
	// knows only durations, so it would hold the buff for the whole window instead. On-use effects
	// are unaffected - their charges are uses of the item, and they do not come through here.
	if procSpell.ProcCharges > 0 {
		return info, false
	}

	// The bit table, decoded in core so the sim reads the same mask the same way. Bits the decode
	// does not model come back in Unsupported and are dropped here rather than refused: a mask
	// naming one next to bits that do model something still generates a listener, and that
	// listener is deliberately the narrower trigger: it hears the hits the sim knows and stays
	// silent on the rest.
	//
	// One rule of the bit table cannot live in core. The heal branch's cleanup for a helpful-only
	// mask also has to clear the OnSpellHitDealt seeded above for a ranged item, which the decode
	// cannot see and this merge cannot undo. No ranged item carries a helpful-only mask with a
	// healing tooltip, and the pureHeal strip below still covers the "healing spells" wording.
	if !onHitProc && len(procSpell.ProcTypeMask) > 0 {
		decoded := core.DecodeProcTypeMask(
			[2]uint32{uint32(procSpell.ProcTypeMask[0]), uint32(procSpell.ProcTypeMask[1])},
			hint,
		)

		info.Callback |= decoded.Callback
		info.ProcMask |= decoded.ProcMask
		info.RequireDamageDealt = decoded.RequireDamageDealt
		info.RequireDamageComponent = decoded.RequireDamageComponent
		info.Outcome = decoded.Outcome
		requiresOutcome = false
	}

	// Proc-ness is a flag on the listener, not a hit kind in the mask. ProcMaskSpellDamageProc is
	// deliberately never emitted: that is the weapon-imbue shape (Flametongue, rogue poisons), kept
	// a distinct hit kind so that Shiffar's Nexus-Horn, Wrath of Cenarius and Robe of the Elder
	// Scribes proc off Ignite, Elemental Overload and Hurricane's DoT rather than off imbue crits.
	info.CanProcFromProcs = procSpell.CanProcFromProcs()
	info.HonoursWeaponProcSuppression = procSpell.IsWeaponProcAura()

	if requiresOutcome {
		if critMatcher.MatchString(tooltip) {
			info.Outcome = core.OutcomeCrit
		} else {
			info.Outcome = core.OutcomeLanded
		}
	}

	// check for pure healing spell
	if pureHealMatcher.MatchString(tooltip) {
		info.Callback &= ^core.CallbackOnSpellHitDealt
		info.Callback &= ^core.CallbackOnPeriodicDamageDealt
	}

	// A trigger with no callback never fires, and factory_ProcStatBonusEffect returns early on
	// exactly that, so such an effect cannot be generated. The test used to also require an
	// empty Outcome and ProcMask, which it can never have: when requiresOutcome is true the
	// Outcome is set a few lines up, and when it is false the whole term is false, so nothing
	// was ever refused here and empty-callback effects were emitted as live registrations that
	// silently did nothing.
	return info, info.Callback != core.CallbackEmpty
}

func StoreMissingEffect(effectType string, name string, variant Variant) {
	if missingEffectsMap[effectType] == nil {
		missingEffectsMap[effectType] = map[int32]MissingItemEffect{}
	}
	id := int32(variant.ID)
	if missingEffectsMap[effectType][id].Effects == nil {
		missingEffectsMap[effectType][id] = MissingItemEffect{
			ItemID:  id,
			Name:    name,
			Effects: []Variant{},
		}
	}
	itemEntry := missingEffectsMap[effectType][id]
	haveEffect := false
	for _, effect := range itemEntry.Effects {
		if effect.SpellID == variant.SpellID {
			haveEffect = true
			break
		}
	}
	if haveEffect {
		return
	}

	itemEntry.Effects = append(
		itemEntry.Effects,
		variant,
	)
	missingEffectsMap[effectType][id] = itemEntry
}

func asCoreCallback(callback core.AuraCallback) string {
	callbacks := []string{}
	for i := range 32 {
		callbackFlag := core.AuraCallback(1 << i)
		if callbackFlag >= core.CallbackLast {
			break
		}

		if callback.Matches(callbackFlag) {
			callbacks = append(callbacks, "core."+callbackFlag.String())
		}
	}

	if len(callbacks) == 0 {
		return "core.CallbackEmpty"
	}

	return strings.Join(callbacks, " | ")
}

func asCoreProcMask(procMask core.ProcMask) string {
	procs := []string{}
	for i := range 32 {
		procFlag := core.ProcMask(1 << i)
		if procFlag >= core.ProcMaskLast {
			break
		}

		if procMask.Matches(procFlag) {
			procs = append(procs, "core."+procFlag.String())
		}
	}

	if len(procs) == 0 {
		return "core.ProcMaskUnknown"
	}
	return strings.Join(procs, " | ")
}

func asCoreOutcome(outcome core.HitOutcome) string {
	if outcome == core.OutcomeCrit {
		return "core.OutcomeCrit"
	}

	if outcome.Matches(core.OutcomeLanded) {
		return "core.OutcomeLanded"
	}

	var avoided []string
	if outcome.Matches(core.OutcomeDodge) {
		avoided = append(avoided, "core.OutcomeDodge")
	}
	if outcome.Matches(core.OutcomeParry) {
		avoided = append(avoided, "core.OutcomeParry")
	}
	if len(avoided) > 0 {
		return strings.Join(avoided, " | ")
	}

	return "core.OutcomeEmpty"
}

func (entry *Entry) AddVariant(variant *Variant) {
	entry.Variants = append(entry.Variants, variant)
	sort.Slice(entry.Variants, func(i, j int) bool {
		return entry.Variants[i].ID < entry.Variants[j].ID
	})
}
