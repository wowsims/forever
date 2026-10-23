package spelldata

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The family the parse rows address, and the masks a modifier effect names inside it.
var (
	parseFamily     = core.ClassFlags{Family: 4, Mask: [4]uint32{0: 0x1}}
	parseTargetMask = core.ClassFlags{Family: 4, Mask: [4]uint32{0: 0x2}}
	parseOtherMask  = core.ClassFlags{Family: 4, Mask: [4]uint32{0: 0x80}}
)

// One row per shape the table has to answer for: the two modifier auras, a hit and dot modifier
// pair, a stacking buff, a charge buff and the passive a stance looks like.
func parseRows() []Spell {
	return []Spell{
		{
			ID: 1000, Name: "Mod Talent", ClassFlags: parseFamily,
			Effects: []Effect{
				{SpellID: 1000, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_ADD_PCT_MODIFIER,
					Misc: int32(dbcenums.SPELLMOD_ALL_EFFECTS), BasePoints: 15, ClassFlags: parseTargetMask},
				{SpellID: 1000, Index: 1, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_ADD_FLAT_MODIFIER,
					Misc: int32(dbcenums.SPELLMOD_COST), BasePoints: -30, ClassFlags: parseTargetMask},
				{SpellID: 1000, Index: 2, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_ADD_FLAT_MODIFIER,
					Misc: int32(dbcenums.SPELLMOD_CASTING_TIME), BasePoints: -500, ClassFlags: parseTargetMask},
				{SpellID: 1000, Index: 3, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_DUMMY,
					BasePoints: 7},
			},
		},
		{
			ID: 1100, Name: "Dot Talent", ClassFlags: parseFamily,
			Effects: []Effect{
				{SpellID: 1100, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_ADD_PCT_MODIFIER,
					Misc: int32(dbcenums.SPELLMOD_DAMAGE), BasePoints: 10, ClassFlags: parseTargetMask},
				{SpellID: 1100, Index: 1, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_ADD_PCT_MODIFIER,
					Misc: int32(dbcenums.SPELLMOD_DOT), BasePoints: 10, ClassFlags: parseTargetMask},
			},
		},
		{
			ID: 1200, Name: "Stacking Buff", DurationMs: 12000, MaxStack: 3, ClassFlags: parseFamily,
			Effects: []Effect{
				{SpellID: 1200, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_ADD_PCT_MODIFIER,
					Misc: int32(dbcenums.SPELLMOD_ALL_EFFECTS), BasePoints: 5, ClassFlags: parseTargetMask},
				{SpellID: 1200, Index: 1, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_THREAT,
					Misc: 127, BasePoints: 30},
			},
		},
		{
			ID: 1300, Name: "Charge Buff", DurationMs: 12000, ProcCharges: 3,
			Effects: []Effect{
				{SpellID: 1300, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_MELEE_HASTE_3,
					BasePoints: 25},
			},
		},
		{
			ID: 1400, Name: "Stance Passive", DurationMs: -1,
			Effects: []Effect{
				{SpellID: 1400, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_DAMAGE_PERCENT_DONE,
					Misc: 127, BasePoints: -10},
				{SpellID: 1400, Index: 1, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_DAMAGE_PERCENT_DONE,
					Misc: 1, BasePoints: 20},
				{SpellID: 1400, Index: 2, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_THREAT,
					Misc: 127, BasePoints: 30},
				{SpellID: 1400, Index: 3, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_ATTACK_POWER,
					BasePoints: 50},
				{SpellID: 1400, Index: 4, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PROC_TRIGGER_SPELL,
					TriggerID: 1300},
			},
		},
	}
}

// A character with its pseudo-stats and dependencies in place, which is what a stat buff and a
// multiplier need under them.
func parseCharacter(player *proto.Player) *core.Character {
	player.Name = "Parse Tester"
	player.Race = proto.Race_RaceOrc
	player.Equipment = &proto.EquipmentSpec{Items: []*proto.ItemSpec{}}

	character := core.NewCharacter(&core.Party{}, 0, player)
	return &character
}

func parseWarrior() *core.Character {
	return parseCharacter(&proto.Player{
		Class: proto.Class_ClassWarrior,
		Spec:  &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
	})
}

// A spell for a mod to land on, carrying the multiplier a registered damage spell starts at.
func parseSpell(character *core.Character, id int32, school core.SpellSchool, procMask core.ProcMask,
	flags core.ClassFlags) *core.Spell {
	return character.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: id},
		ClassFlags:       flags,
		ProcMask:         procMask,
		SpellSchool:      school,
		DamageMultiplier: 1,
	})
}

func appliedKinds(p *Parsed) []string {
	kinds := make([]string, len(p.Applied))
	for i, a := range p.Applied {
		kinds[i] = a.Kind
	}
	return kinds
}

func TestParseStaticBuildsTheModTable(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()
	character.EnableRageBar(core.RageBarOptions{})

	parsed := ParseStatic(character, Find(1000))

	want := []string{"SpellMod_DamageDone_Flat", "SpellMod_PowerCost_Flat", "SpellMod_CastTime_Flat"}
	if len(parsed.Applied) != len(want) {
		t.Fatalf("attached %v, want %v", appliedKinds(parsed), want)
	}
	for i, kind := range appliedKinds(parsed) {
		if kind != want[i] {
			t.Errorf("effect %d became %q, want %q", i+1, kind, want[i])
		}
	}

	if parsed.Applied[0].Value != 0.15 {
		t.Errorf("a percentage modifier is worth %v, want 0.15", parsed.Applied[0].Value)
	}
	if parsed.Applied[1].Value != -30 {
		t.Errorf("a cost modifier is worth %v, want the row's -30", parsed.Applied[1].Value)
	}
	if parsed.Applied[2].Value != -500 {
		t.Errorf("a cast time modifier is worth %v ms, want -500", parsed.Applied[2].Value)
	}

	if len(parsed.Skipped) != 1 || parsed.Skipped[0].Aura != dbcenums.A_DUMMY {
		t.Errorf("skipped %v, want the dummy effect alone", parsed.Skipped)
	}
}

// One cost modifier takes rage off a rage spell on the game files' 0-1000 bar, and the row's own number
// off a spell of another power, on the same caster.
func TestParseStaticCostInEachSpellsPower(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()
	character.EnableRageBar(core.RageBarOptions{})

	rage := character.RegisterSpell(core.SpellConfig{ActionID: core.ActionID{SpellID: 2000},
		ClassFlags: parseTargetMask, RageCost: core.RageCostOptions{Cost: 15}})
	energy := character.RegisterSpell(core.SpellConfig{ActionID: core.ActionID{SpellID: 2001},
		ClassFlags: parseTargetMask, EnergyCost: core.EnergyCostOptions{Cost: 40}})

	ParseStatic(character, Find(1000))

	if got := rage.Cost.FlatModifier; got != -3 {
		t.Errorf("the rage spell's cost moves by %d, want -3", got)
	}
	if got := energy.Cost.FlatModifier; got != -30 {
		t.Errorf("the energy spell's cost moves by %d, want the row's -30", got)
	}
}

// A mod reaches the spells its own effect names and no others.
func TestParseStaticCarriesTheClassFlags(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()
	character.EnableRageBar(core.RageBarOptions{})

	named := parseSpell(character, 2100, core.SpellSchoolPhysical, core.ProcMaskSpellDamage, parseTargetMask)
	elsewhere := parseSpell(character, 2101, core.SpellSchoolPhysical, core.ProcMaskSpellDamage, parseOtherMask)

	ParseStatic(character, Find(1000))

	if got := named.DamageMultiplierAdditive; got != 1.15 {
		t.Errorf("the spell the effect names: %v, want 1.15", got)
	}
	if got := elsewhere.DamageMultiplierAdditive; got != 1 {
		t.Errorf("a spell of the same family the effect does not name: %v, want 1", got)
	}
}

// An effect naming no spells at all means every spell of the caster's family.
func TestParseStaticFallsBackToTheWholeFamily(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	family := parseSpell(character, 2200, core.SpellSchoolPhysical, core.ProcMaskSpellDamage, parseOtherMask)
	outsider := parseSpell(character, 2201, core.SpellSchoolPhysical, core.ProcMaskSpellDamage,
		core.ClassFlags{Family: 9, Mask: [4]uint32{0: 0x80}})

	row := *Find(1100)
	row.Effects = []Effect{{SpellID: 1100, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_ADD_PCT_MODIFIER,
		Misc: int32(dbcenums.SPELLMOD_ALL_EFFECTS), BasePoints: 10}}

	ParseStatic(character, &row)

	if got := family.DamageMultiplierAdditive; got != 1.1 {
		t.Errorf("a spell of the caster's family: %v, want 1.1", got)
	}
	if got := outsider.DamageMultiplierAdditive; got != 1 {
		t.Errorf("a spell of another family: %v, want 1", got)
	}
}

// A talent that raises a spell and its dot states the bonus twice: SPELLMOD_DAMAGE reaches the hit
// alone and SPELLMOD_DOT the ticks, so each attaches its own mod.
func TestParseKeepsTheHitAndDotModifiersApart(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()
	spell := character.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 2150},
		ClassFlags:       parseTargetMask,
		ProcMask:         core.ProcMaskSpellDamage,
		SpellSchool:      core.SpellSchoolPhysical,
		DamageMultiplier: 1,
		Dot:              core.DotConfig{Aura: core.Aura{Label: "Parse Dot"}, IsAOE: true, NumberOfTicks: 1, TickLength: time.Second},
	})

	parsed := ParseStatic(character, Find(1100))

	kinds := appliedKinds(parsed)
	if len(kinds) != 2 || kinds[0] != "SpellMod_DirectDamageDone_Pct" || kinds[1] != "SpellMod_DotDamageDone_Pct" {
		t.Errorf("the pair became %v, want a direct damage mod and a dot damage mod", kinds)
	}
	if len(parsed.Skipped) != 0 {
		t.Errorf("skipped %d effects, want none", len(parsed.Skipped))
	}
	if got := spell.DirectDamageMultiplier; math.Abs(got-1.1) > 1e-9 {
		t.Errorf("the hit modifier: %v, want 1.1", got)
	}
	if got := spell.DamageMultiplierAdditive; got != 1 {
		t.Errorf("the multiplier the ticks share with the hit: %v, want 1", got)
	}
	if got := spell.AOEDot().PeriodicDamageMultiplier; math.Abs(got-1.1) > 1e-9 {
		t.Errorf("the dot modifier: %v, want 1.1", got)
	}
}

// Elemental Focus's Clearcasting parses to a percent cost mod: with its cost row taking the whole
// cost, the spells it names are free while it is up and cost what they did once it drops.
func TestGeneratedFreeCastRowGivesTheCostBack(t *testing.T) {
	withGeneratedStore(t)
	sim := &core.Simulation{}
	character := parseCharacter(&proto.Player{
		Class: proto.Class_ClassShaman,
		Spec:  &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{}},
	})

	row := *MustFind(16246)
	row.Effects = slices.Clone(row.Effects)
	costRow := row.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_COST))
	costRow.BasePoints = -100
	spell := character.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 2190},
		ClassFlags:  costRow.ClassFlags,
		SpellSchool: core.SpellSchoolNature,
		ManaCost:    core.ManaCostOptions{FlatCost: 120},
	})
	base := spell.Cost.GetCurrentCost()

	aura := character.RegisterAura(AuraConfig(&row))
	parsed := ParseEffects(character, aura, &row)
	if kinds := appliedKinds(parsed); len(kinds) != 1 || kinds[0] != "SpellMod_PowerCost_Pct" {
		t.Fatalf("attached %v, want the cost row alone", kinds)
	}

	aura.Activate(sim)
	if got := spell.Cost.GetCurrentCost(); got != 0 {
		t.Errorf("the cost while Clearcasting is up: %v, want 0", got)
	}
	aura.Deactivate(sim)
	if got := spell.Cost.GetCurrentCost(); got != base {
		t.Errorf("the cost after Clearcasting dropped: %v, want %v", got, base)
	}
}

// A hit row parsed with ShowAsSchoolHit shows on the sheet under the schools the caller names, and the
// hit still reaches only the spells the row's class flags name.
func TestParseStaticShowAsSchoolHit(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()
	named := parseSpell(character, 2160, core.SpellSchoolArcane, core.ProcMaskSpellDamage, parseTargetMask)
	other := parseSpell(character, 2161, core.SpellSchoolArcane, core.ProcMaskSpellDamage, parseOtherMask)

	ParseStatic(character, oneEffectRow(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_RESIST_MISS_CHANCE), 5),
		ShowAsSchoolHit(core.SpellSchoolArcane))

	if math.Abs(named.BonusHitPercent-5) > 1e-9 || other.BonusHitPercent != 0 {
		t.Errorf("hit on the named spell %v and on another arcane spell %v, want 5 and 0", named.BonusHitPercent, other.BonusHitPercent)
	}
	arcaneHit := character.GetPseudoStatsProto()[proto.PseudoStat_PseudoStatSchoolHitPercentArcane]
	if got := arcaneHit - character.GetStat(stats.SpellHitPercent); math.Abs(got-5) > 1e-9 {
		t.Errorf("the sheet shows %v arcane hit over spell hit, want 5", got)
	}
}

func TestParseStaticShowAsSchoolHitNeedsAHitRow(t *testing.T) {
	withRows(t, parseRows())
	defer func() {
		if recover() == nil {
			t.Error("ShowAsSchoolHit on a row with no hit modifier did not panic")
		}
	}()
	ParseStatic(parseWarrior(), oneEffectRow(dbcenums.A_MOD_DODGE_PERCENT, 0, 5), ShowAsSchoolHit(core.SpellSchoolArcane))
}

func TestParseReadsOnlyTheNamedEffects(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()
	character.EnableRageBar(core.RageBarOptions{})

	only := ParseStatic(character, Find(1000), Effects(2))
	if kinds := appliedKinds(only); len(kinds) != 1 || kinds[0] != "SpellMod_PowerCost_Flat" {
		t.Errorf("Effects(2) attached %v, want the cost modifier alone", kinds)
	}
	if len(only.Skipped) != 0 {
		t.Errorf("Effects(2) reported %d skipped effects, want none", len(only.Skipped))
	}

	rest := ParseStatic(character, Find(1000), SkipEffects(1, 4))
	if kinds := appliedKinds(rest); len(kinds) != 2 {
		t.Errorf("SkipEffects(1, 4) attached %v, want the two middle effects", kinds)
	}
}

// A conditional parse attaches everything and turns it on only while the condition holds.
func TestParseConditionalAndRefresh(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	baseAttackPower := character.GetStat(stats.AttackPower)

	allowed := false
	parsed := ParseStatic(character, Find(1400), Conditional(func() bool { return allowed }))

	if len(parsed.Applied) != 4 {
		t.Fatalf("attached %v, want the four the table knows", appliedKinds(parsed))
	}
	if got := character.PseudoStats.ThreatMultiplier; got != 1 {
		t.Errorf("threat multiplier while the condition is false: %v, want 1", got)
	}
	if got := character.GetStat(stats.AttackPower); got != baseAttackPower {
		t.Errorf("attack power while the condition is false: %v, want %v", got, baseAttackPower)
	}

	allowed = true
	parsed.Refresh(nil)
	if got := character.PseudoStats.ThreatMultiplier; got != 1.3 {
		t.Errorf("threat multiplier after the condition turned true: %v, want 1.3", got)
	}
	if got := character.PseudoStats.DamageDealtMultiplier; got != 0.9 {
		t.Errorf("damage dealt multiplier: %v, want 0.9", got)
	}
	if got := character.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]; got != 1.2 {
		t.Errorf("physical damage multiplier: %v, want 1.2", got)
	}
	if got := character.GetStat(stats.AttackPower); got != baseAttackPower+50 {
		t.Errorf("attack power after the condition turned true: %v, want %v", got, baseAttackPower+50)
	}

	allowed = false
	parsed.Refresh(nil)
	if got := character.PseudoStats.ThreatMultiplier; got != 1 {
		t.Errorf("threat multiplier after the condition turned false again: %v, want 1", got)
	}
	if got := character.GetStat(stats.AttackPower); got != baseAttackPower {
		t.Errorf("attack power after the condition turned false again: %v, want %v", got, baseAttackPower)
	}
}

// The proc effect is the one a port still has to wire, and the report names it.
func TestParseSkipsWhatTheTableDoesNotKnow(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	parsed := ParseStatic(character, Find(1400))

	if len(parsed.Skipped) != 1 || parsed.Skipped[0].Aura != dbcenums.A_PROC_TRIGGER_SPELL {
		t.Errorf("skipped %v, want the proc trigger alone", parsed.Skipped)
	}
}

// A rank the character does not have is the store's empty row, which attaches nothing.
func TestParseNilSpell(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	parsed := ParseStatic(character, Nil)
	if len(parsed.Applied) != 0 || len(parsed.Skipped) != 0 {
		t.Errorf("the empty row attached %v and skipped %v", parsed.Applied, parsed.Skipped)
	}
	parsed.Refresh(nil)

	if empty := ParseEffects(character, nil, Find(1000)); len(empty.Applied) != 0 {
		t.Errorf("a parse without an aura attached %v", appliedKinds(empty))
	}
}

func TestParseEffectsFollowsTheAuraAndItsStacks(t *testing.T) {
	withRows(t, parseRows())
	sim := &core.Simulation{}
	character := parseWarrior()

	spell := parseSpell(character, 2300, core.SpellSchoolPhysical, core.ProcMaskSpellDamage, parseTargetMask)
	aura := character.RegisterAura(AuraConfig(Find(1200)))
	parsed := ParseEffects(character, aura, Find(1200))

	if kinds := appliedKinds(parsed); len(kinds) != 1 || kinds[0] != "SpellMod_DamageDone_Flat" {
		t.Fatalf("attached %v, want the damage modifier alone", kinds)
	}
	// The threat multiplier cannot carry a per-stack value, so a stacking row leaves it to the caller.
	if len(parsed.Skipped) != 1 || parsed.Skipped[0].Aura != dbcenums.A_MOD_THREAT {
		t.Errorf("skipped %v, want the threat effect", parsed.Skipped)
	}

	if got := spell.DamageMultiplierAdditive; got != 1 {
		t.Errorf("the modifier before the aura is up: %v, want 1", got)
	}

	aura.Activate(sim)
	if got := spell.DamageMultiplierAdditive; got != 1.05 {
		t.Errorf("the modifier at one stack: %v, want 1.05", got)
	}

	aura.SetStacks(sim, 2)
	if got := spell.DamageMultiplierAdditive; got != 1.1 {
		t.Errorf("the modifier at two stacks: %v, want 1.1", got)
	}

	aura.Deactivate(sim)
	if got := spell.DamageMultiplierAdditive; got != 1 {
		t.Errorf("the modifier after the aura dropped: %v, want 1", got)
	}
}

// A row that states charges rather than cumulative stacks is worth its value once, however many
// charges the aura carries.
func TestParseEffectsChargesAreNotStacks(t *testing.T) {
	withRows(t, parseRows())
	sim := &core.Simulation{}
	character := parseWarrior()

	aura := character.RegisterAura(AuraConfig(Find(1300)))
	parsed := ParseEffects(character, aura, Find(1300))

	if kinds := appliedKinds(parsed); len(kinds) != 1 || kinds[0] != "melee-speed" {
		t.Fatalf("attached %v, want the haste row", kinds)
	}

	aura.Activate(sim)
	aura.SetStacks(sim, 3)

	if got := character.PseudoStats.MeleeSpeedMultiplier; got != 1.25 {
		t.Errorf("melee speed at three charges: %v, want the row's 1.25", got)
	}
}

// The rows whose value cannot follow the stacks are skipped on a stacking aura rather than attached
// at one stack, unless the caller says the value does not follow them.
func TestParseEffectsSkipsWhatCannotFollowTheStacks(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	row := *Find(1300)
	row.MaxStack = 3

	parsed := ParseEffects(character, character.RegisterAura(AuraConfig(&row)), &row)
	if len(parsed.Applied) != 0 || len(parsed.Skipped) != 1 {
		t.Errorf("a haste row on a stacking aura attached %v", appliedKinds(parsed))
	}

	ignored := ParseEffects(character, character.RegisterAura(core.Aura{Label: "Ignored Stacks"}), &row, IgnoreStacks())
	if len(ignored.Applied) != 1 {
		t.Errorf("IgnoreStacks attached %v, want the haste row", appliedKinds(ignored))
	}
}

// The aura every rank of a family points at takes a duration modifier once, however many ranks the
// modifier's mask names.
func TestParseStaticDurationModOnASharedAura(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	buff := character.RegisterAura(core.Aura{Label: "Shared Buff", Duration: time.Second * 10})
	for i := range 3 {
		character.RegisterSpell(core.SpellConfig{
			ActionID:        core.ActionID{SpellID: 2400 + int32(i)},
			ClassFlags:      parseTargetMask,
			ProcMask:        core.ProcMaskSpellDamage,
			SpellSchool:     core.SpellSchoolPhysical,
			RelatedSelfBuff: buff,
		})
	}

	row := Spell{
		ID: 3000, Name: "Duration Talent", ClassFlags: parseFamily,
		Effects: []Effect{{SpellID: 3000, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_ADD_FLAT_MODIFIER,
			Misc: int32(dbcenums.SPELLMOD_DURATION), BasePoints: 5000, ClassFlags: parseTargetMask}},
	}

	ParseStatic(character, &row)

	if got := buff.Duration; got != time.Second*15 {
		t.Errorf("the shared aura's duration: %v, want 15s", got)
	}
}

// Every row of the three tables, on a row of one effect: what it becomes and what it is worth. The
// haste rows run through ParseEffects, since the speeds they multiply need the Simulation an aura's
// gain hands over.
func TestEveryTableRow(t *testing.T) {
	// The caster's proc a proc chance row's class mask has to name.
	withRows(t, append(parseRows(), Spell{ID: 1500, Name: "Caster Proc", ProcChance: 10,
		ProcFlags: [2]uint32{0: 0x14}, ClassFlags: parseTargetMask,
		Effects: []Effect{{SpellID: 1500, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_DUMMY,
			Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER}}}}))

	const (
		flat = dbcenums.A_ADD_FLAT_MODIFIER
		pct  = dbcenums.A_ADD_PCT_MODIFIER
	)

	cases := []struct {
		aura   dbcenums.EffectAuraType
		misc   int32
		points float64
		kind   string
		value  float64
		onAura bool
	}{
		{flat, int32(dbcenums.SPELLMOD_DURATION), 5000, "SpellMod_Duration_Flat", 5000, false},
		{flat, int32(dbcenums.SPELLMOD_CHARGES), 1, "SpellMod_BuffMaxStacks_Flat", 1, false},
		{flat, int32(dbcenums.SPELLMOD_RANGE), 5, "SpellMod_Range_Flat", 5, false},
		{flat, int32(dbcenums.SPELLMOD_CRITICAL_CHANCE), 6, "SpellMod_BonusCrit_Percent", 6, false},
		{flat, int32(dbcenums.SPELLMOD_CASTING_TIME), -500, "SpellMod_CastTime_Flat", -500, false},
		{flat, int32(dbcenums.SPELLMOD_COOLDOWN), -2000, "SpellMod_Cooldown_Flat", -2000, false},
		{flat, int32(dbcenums.SPELLMOD_COST), -30, "SpellMod_PowerCost_Flat", -30, false},
		{flat, int32(dbcenums.SPELLMOD_RESIST_MISS_CHANCE), 3, "SpellMod_BonusHit_Percent", 3, false},
		{flat, int32(dbcenums.SPELLMOD_GLOBAL_COOLDOWN), -500, "SpellMod_GlobalCooldown_Flat", -500, false},
		{flat, int32(dbcenums.SPELLMOD_ACTIVATION_TIME), -500, "SpellMod_DotTickLength_Flat", -500, false},
		{flat, int32(dbcenums.SPELLMOD_CHANCE_OF_SUCCESS), 10, "SpellMod_ProcChance_Flat", 10, false},
		{flat, int32(dbcenums.SPELLMOD_EFFECT1), 20, "effect1-assumed-damage SpellMod_BaseDamage_Flat", 20, false},
		{flat, int32(dbcenums.SPELLMOD_EFFECT2), 20, "effect2-assumed-damage SpellMod_BaseDamage_Flat", 20, false},
		{flat, int32(dbcenums.SPELLMOD_EFFECT3), 20, "effect3-assumed-damage SpellMod_BaseDamage_Flat", 20, false},

		{pct, int32(dbcenums.SPELLMOD_DAMAGE), 15, "SpellMod_DirectDamageDone_Pct", 0.15, false},
		{pct, int32(dbcenums.SPELLMOD_ALL_EFFECTS), 15, "SpellMod_DamageDone_Flat", 0.15, false},
		{pct, int32(dbcenums.SPELLMOD_DURATION), 20, "SpellMod_DotBaseDuration_Pct", 0.2, false},
		{pct, int32(dbcenums.SPELLMOD_THREAT), -20, "SpellMod_ThreatMultiplier_Pct", -0.2, false},
		{pct, int32(dbcenums.SPELLMOD_CASTING_TIME), -10, "SpellMod_CastTime_Pct", -0.1, false},
		{pct, int32(dbcenums.SPELLMOD_COOLDOWN), -20, "SpellMod_Cooldown_Multiplier", 0.8, false},
		{pct, int32(dbcenums.SPELLMOD_COST), -20, "SpellMod_PowerCost_Pct", -0.2, false},
		{pct, int32(dbcenums.SPELLMOD_CRIT_DAMAGE_BONUS), 20, "SpellMod_CritMultiplier_Flat", 0.2, false},
		{pct, int32(dbcenums.SPELLMOD_CRITICAL_CHANCE), 7, "SpellMod_BonusCrit_Percent", 7, false},
		{pct, int32(dbcenums.SPELLMOD_DOT), 10, "SpellMod_DotDamageDone_Pct", 0.1, false},
		{pct, int32(dbcenums.SPELLMOD_CHANCE_OF_SUCCESS), 100, "SpellMod_ProcChance_Pct", 1, false},
		{pct, int32(dbcenums.SPELLMOD_PROC_COOLDOWN), -50, "SpellMod_ProcCooldown_Pct", -0.5, false},
		{pct, int32(dbcenums.SPELLMOD_GLOBAL_COOLDOWN), -10, "SpellMod_GlobalCooldown_Pct", -0.1, false},
		{pct, int32(dbcenums.SPELLMOD_RANGE), 10, "SpellMod_Range_Pct", 0.1, false},
		{pct, int32(dbcenums.SPELLMOD_EFFECT1), 10, "effect1-assumed-damage SpellMod_DamageDone_Flat", 0.1, false},
		{pct, int32(dbcenums.SPELLMOD_EFFECT2), 10, "effect2-assumed-damage SpellMod_DamageDone_Flat", 0.1, false},
		{pct, int32(dbcenums.SPELLMOD_EFFECT3), 10, "effect3-assumed-damage SpellMod_DamageDone_Flat", 0.1, false},

		{dbcenums.A_MOD_ATTACKSPEED, 0, 20, "attack-speed", 1.2, true},
		{dbcenums.A_MOD_THREAT, 127, 30, "threat", 1.3, false},
		{dbcenums.A_MOD_THREAT, 72, -30, "threat-by-school", 0.7, false},
		{dbcenums.A_MOD_DAMAGE_DONE, 1, 20, "stat PhysicalDamage", 20, false},
		{dbcenums.A_MOD_DAMAGE_DONE, 126, 20, "stat SpellDamage", 20, false},
		{dbcenums.A_MOD_DAMAGE_DONE, 127, 20, "stat PhysicalDamage, stat SpellDamage", 20, false},
		{dbcenums.A_MOD_DAMAGE_DONE, 4, 20, "stat FireDamage", 20, false},
		{dbcenums.A_MOD_RESISTANCE, 1, 100, "stat Armor", 100, false},
		{dbcenums.A_MOD_BASE_RESISTANCE_PCT, 1, 10, "equip-scaling Armor", 1.1, false},
		{dbcenums.A_MOD_RESISTANCE_PCT, 1, 8, "multiply-stat Armor", 1.08, false},
		{dbcenums.A_MOD_ARMOR_NO_MODS, 1, -450, "stat ArmorNoMods", -450, false},
		{dbcenums.A_MOD_RESISTANCE, 16, 30, "stat FrostResistance", 30, false},
		{dbcenums.A_MOD_RESISTANCE, 127, 30,
			"stat Armor, stat FireResistance, stat NatureResistance, stat FrostResistance, stat ShadowResistance, stat ArcaneResistance", 30, false},
		{dbcenums.A_MOD_STAT, -1, 10, "stat Strength, stat Agility, stat Stamina, stat Intellect, stat Spirit", 10, false},
		{dbcenums.A_MOD_STAT, 2, 10, "stat Stamina", 10, false},
		{dbcenums.A_MOD_INCREASE_HEALTH, 0, 200, "stat Health", 200, false},
		{dbcenums.A_MOD_MAX_HEALTH, 0, 560, "stat Health", 560, false},
		{dbcenums.A_MOD_INCREASE_HEALTH_PERCENT, 0, 10, "multiply-stat Health", 1.1, false},
		{dbcenums.A_MOD_PARRY_PERCENT, 0, 5, "stat ParryPercent", 5, false},
		{dbcenums.A_MOD_SKILL, miscDefense, 20, "stat DefenseRating", 20 * core.DefenseRatingPerDefenseLevel, false},
		{dbcenums.A_MOD_SKILL, 173, 5, "weapon-skill", 5, false},
		{dbcenums.A_MOD_DODGE_PERCENT, 0, 5, "stat DodgePercent", 5, false},
		{dbcenums.A_MOD_BLOCK_PERCENT, 0, 5, "stat BlockPercent", 5, false},
		{dbcenums.A_MOD_WEAPON_CRIT_PERCENT, 0, 5, "stat PhysicalCritPercent", 5, false},
		{dbcenums.A_MOD_HIT_CHANCE, 0, 3, "stat PhysicalHitPercent", 3, false},
		{dbcenums.A_MOD_SPELL_HIT_CHANCE, 0, 3, "stat SpellHitPercent", 3, false},
		{dbcenums.A_MOD_SPELL_CRIT_CHANCE, 0, 3, "stat SpellCritPercent", 3, false},
		{dbcenums.A_MOD_CRIT_PCT, 0, 2, "stat PhysicalCritPercent, stat SpellCritPercent", 2, false},
		{dbcenums.A_MOD_CASTING_SPEED_NOT_STACK, 0, 20, "cast-speed", 1.2, true},
		{dbcenums.A_MOD_DAMAGE_PERCENT_DONE, 127, -10, "damage-dealt", 0.9, false},
		{dbcenums.A_MOD_DAMAGE_PERCENT_DONE, 1, 10, "damage-dealt-by-school", 1.1, false},
		{dbcenums.A_MOD_DAMAGE_PERCENT_TAKEN, 127, -10, "damage-taken", 0.9, false},
		{dbcenums.A_MOD_DAMAGE_PERCENT_TAKEN, 4, -10, "damage-taken-by-school", 0.9, false},
		{dbcenums.A_MOD_POWER_REGEN, 0, 8, "stat MP5", 8, false},
		{dbcenums.A_MOD_HEALTH_REGEN_IN_COMBAT, 0, 15, "stat HP5", 15, false},
		{dbcenums.A_MOD_MANA_REGEN_INTERRUPT, 0, 50, "spirit-regen-casting", 0.5, false},
		{dbcenums.A_MOD_POWER_REGEN_PERCENT, 0, 400, "spirit-regen", 5, false},
		{dbcenums.A_MOD_ATTACK_POWER, 0, 50, "stat AttackPower", 50, false},
		{dbcenums.A_MOD_RANGED_ATTACK_POWER, 0, 50, "stat RangedAttackPower", 50, false},
		{dbcenums.A_MOD_HEALING, 0, 20, "healing-taken-flat", 20, false},
		{dbcenums.A_MOD_HEALING_PCT, 0, 10, "healing-taken", 1.1, false},
		{dbcenums.A_MOD_HEALING_DONE, 0, 30, "stat HealingPower", 30, false},
		{dbcenums.A_MOD_HEALING_DONE_PERCENT, 0, 10, "healing-dealt", 1.1, false},
		{dbcenums.A_MOD_OFFHAND_DAMAGE_PCT, 0, 25, "SpellMod_DamageDone_Pct", 0.25, false},
		{dbcenums.A_MOD_PERCENT_STAT, 4, 100, "multiply-stat Spirit", 2, false},
		{dbcenums.A_MOD_PERCENT_STAT, -1, 5,
			"multiply-stat Strength, multiply-stat Agility, multiply-stat Stamina, multiply-stat Intellect, multiply-stat Spirit", 1.05, false},
		{dbcenums.A_MOD_DAMAGE_TAKEN, 1, 8, "physical-damage-taken-flat", 8, false},
		{dbcenums.A_MOD_DAMAGE_TAKEN, 126, 8, "spell-damage-taken-flat", 8, false},
		{dbcenums.A_RANGED_ATTACK_POWER_ATTACKER_BONUS, 0, 71, "ranged-attack-power-attacker-bonus", 71, false},
		{dbcenums.A_REDUCE_PUSHBACK, 127, 35, "pushback", -0.35, false},
		{dbcenums.A_MOD_RATING, 4, 10, "stat DodgeRating", 10, false},
		{dbcenums.A_MOD_RATING, 256, 10, "stat CritRating", 10, false},
		{dbcenums.A_MOD_RATING, 768, 10, "stat CritRating", 10, false},
		{dbcenums.A_MOD_RATING, 224, 5, "stat HitRating", 5, false},
		{dbcenums.A_MOD_RATING, 1792, 21, "stat CritRating", 21, false},
		{dbcenums.A_MOD_RATING, 393216, 10, "stat HasteRating", 10, false},
		{dbcenums.A_MOD_RATING, 524288, 10, "stat HasteRating", 10, false},
		{dbcenums.A_MOD_RATING, 32 | 256, 10, "stat HitRating, stat CritRating", 10, false},
		{dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE, -1, 10,
			"multiply-stat Strength, multiply-stat Agility, multiply-stat Stamina, multiply-stat Intellect, multiply-stat Spirit", 1.1, false},
		{dbcenums.A_MECHANIC_DURATION_MOD, int32(dbcenums.MECHANIC_FEAR), -20, "fear-duration", 0.8, false},
		{dbcenums.A_MECHANIC_DURATION_MOD, int32(dbcenums.MECHANIC_STUN), -20, "stun-duration", 0.8, false},
		{dbcenums.A_MOD_EXPERTISE, 0, 5, "stat ExpertisePercent", 5, false},
		{dbcenums.A_MOD_MELEE_HASTE_3, 0, 25, "melee-speed", 1.25, true},
		{dbcenums.A_MOD_DAMAGE_TAKEN, 127, -10, "physical-damage-taken-flat, spell-damage-taken-flat", -10, false},
		{dbcenums.A_MOD_MELEE_DAMAGE_TAKEN_PCT, 0, -20, "melee-damage-taken", 0.8, false},
		{dbcenums.A_MOD_RANGED_DAMAGE_TAKEN, 1, -32, "ranged-damage-taken-flat", -32, false},
		{dbcenums.A_MOD_POWER_COST_SCHOOL_PCT, 126, 20, "SpellMod_PowerCost_Pct", 0.2, false},
		{dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, 4, 185, "mob-type-stat AttackPower", 185, true},
		{dbcenums.A_MOD_RANGED_ATTACK_POWER_VERSUS, 32, 99, "mob-type-stat RangedAttackPower", 99, true},
		{dbcenums.A_MOD_FLAT_SPELL_DAMAGE_VERSUS, 32, 58, "mob-type-stat SpellDamage", 58, true},
		{dbcenums.A_MOD_DAMAGE_DONE_VERSUS, 32, 1, "damage-dealt-versus", 1.01, false},
		{dbcenums.A_MOD_CRIT_PERCENT_VERSUS, 64, 3, "crit-damage-versus", 1.03, false},
		{dbcenums.A_SPELL_AURA_MOD_ARMOR_PENETRATION_PCT, 0, 9, "armor-ignore", 0.09, false},
		{dbcenums.A_MOD_RANGED_HASTE, 0, 40, "ranged-speed", 1.4, true},
		{dbcenums.A_MOD_MELEE_RANGED_HASTE_2, 0, 30, "attack-speed", 1.3, true},
		{dbcenums.A_MOD_ARMOR_PCT_FROM_STAT, 3, 50, "stat-share Intellect>Armor", 0.5, false},
		{dbcenums.A_MOD_ATTACK_POWER_OF_STAT_PERCENT, 1, 100, "stat-share Agility>AttackPower", 1, false},
		{dbcenums.A_MOD_RANGED_ATTACK_POWER_OF_STAT_PERCENT, 3, 20, "stat-share Intellect>RangedAttackPower", 0.2, false},
		{dbcenums.A_MOD_SPELL_DAMAGE_OF_STAT_PERCENT, 126, 8, "stat-share Spirit>SpellDamage", 0.08, false},
		{dbcenums.A_MOD_SPELL_HEALING_OF_STAT_PERCENT, 4, 25, "stat-share Spirit>HealingPower", 0.25, false},
		{dbcenums.A_MOD_MAX_POWER_PCT, 0, 15, "multiply-stat Mana", 1.15, false},
		{dbcenums.A_MOD_BLOCK_VALUE_PCT, 0, 30, "block-value", 1.3, false},
		{dbcenums.A_MOD_ATTACKER_MELEE_HIT_CHANCE, 0, -50, "physical-hit-taken", 50, false},
		{dbcenums.A_MOD_MAX_POWER, int32(dbcenums.POWER_RAGE), 300, "max-rage", 30, false},
		{dbcenums.A_MOD_MAX_POWER, int32(dbcenums.POWER_ENERGY), 10, "max-energy", 10, false},
		{dbcenums.A_MOD_INCREASE_ENERGY, int32(dbcenums.POWER_ENERGY), 10, "max-energy", 10, false},
		{dbcenums.A_MOD_MAX_POWER, int32(dbcenums.POWER_MANA), 2000, "stat Mana", 2000, false},
		{dbcenums.A_MOD_INCREASE_ENERGY, int32(dbcenums.POWER_MANA), 150, "stat Mana", 150, false},
		{dbcenums.A_MOD_SCHOOL_MASK_DAMAGE_FROM_CASTER, 32, 2, "caster-damage-by-school", 1.02, true},
		{dbcenums.A_MOD_SPELL_DAMAGE_FROM_CASTER, 0, 15, "caster-damage-by-spell", 1.15, true},
		{dbcenums.A_MOD_CRIT_CHANCE_FOR_CASTER_WITH_ABILITIES, 16, 2, "caster-crit-by-spell", 2, true},
		{dbcenums.A_MOD_TARGET_RESISTANCE, 1, -200, "stat ArmorPenetration", 200, false},
		{dbcenums.A_MOD_TARGET_RESISTANCE, 126, -15, "stat SpellPiercing", 15, false},
		{dbcenums.A_MOD_TARGET_RESISTANCE, 4, -10, "school-spell-piercing", 10, false},
	}

	// The aura that names its stat in the second misc value, which its case reads as Spirit.
	secondMisc := map[dbcenums.EffectAuraType]int32{dbcenums.A_MOD_SPELL_DAMAGE_OF_STAT_PERCENT: 4}

	// A row added to one of the three tables and to no case here would be invisible, so the cases
	// are read back as the coverage they are.
	auras := map[dbcenums.EffectAuraType]bool{}
	flatMods, pctMods := map[dbcenums.SpellModOp]bool{}, map[dbcenums.SpellModOp]bool{}
	for _, c := range cases {
		auras[c.aura] = true
		switch c.aura {
		case flat:
			flatMods[dbcenums.SpellModOp(c.misc)] = true
		case pct:
			pctMods[dbcenums.SpellModOp(c.misc)] = true
		}
	}
	for aura := range auraTable {
		if !auras[aura] {
			t.Errorf("the aura table carries %s and no case reads it", auraName(aura))
		}
	}
	for misc := range flatModTable {
		if !flatMods[misc] {
			t.Errorf("the flat modifier table carries misc %d and no case reads it", misc)
		}
	}
	for misc := range pctModTable {
		if !pctMods[misc] {
			t.Errorf("the percentage modifier table carries misc %d and no case reads it", misc)
		}
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s_misc%d", auraName(c.aura), c.misc), func(t *testing.T) {
			character := parseWarrior()
			if c.aura == dbcenums.A_MOD_MAX_POWER || c.aura == dbcenums.A_MOD_INCREASE_ENERGY {
				enablePowerBar(character, dbcenums.PowerType(c.misc))
			}
			row := Spell{ID: 5000, Name: "Table Row", DurationMs: 10000, ClassFlags: parseFamily,
				Effects: []Effect{{Type: dbcenums.E_APPLY_AURA, Aura: c.aura, Misc: c.misc, Misc2: secondMisc[c.aura],
					BasePoints: c.points, ClassFlags: parseTargetMask}}}

			// The character casts the aura it wears, which is all a row that reaches only its caster's
			// attacks needs to attach.
			var parsed *Parsed
			if c.onAura {
				parsed = ParseEffects(character, character.RegisterAura(AuraConfig(&row)), &row,
					FromCaster(&character.Unit, 0, 1))
			} else {
				parsed = ParseStatic(character, &row)
			}

			if len(parsed.Applied) == 0 {
				t.Fatalf("attached nothing and skipped %d, want an attachment", len(parsed.Skipped))
			}
			if got := strings.Join(appliedKinds(parsed), ", "); got != c.kind {
				t.Errorf("kinds = %q, want %q", got, c.kind)
			}
			if got := parsed.Applied[0].Value; math.Abs(got-c.value) > 1e-9 {
				t.Errorf("value = %v, want %v", got, c.value)
			}
		})
	}
}

func TestTotalStatPercentageReadsTheStatMask(t *testing.T) {
	allFive := "multiply-stat Strength, multiply-stat Agility, multiply-stat Stamina, multiply-stat Intellect, multiply-stat Spirit"
	cases := []struct {
		misc, misc2 int32
		kind        string
	}{
		{0, 8, "multiply-stat Intellect"},
		{2, 5, "multiply-stat Strength, multiply-stat Stamina"},
		{-1, 31, allFive},
		{4, 16, "multiply-stat Spirit"},
		{3, 0, "multiply-stat Intellect"},
	}
	for _, c := range cases {
		row := Spell{ID: 5000, Name: "Table Row", ClassFlags: parseFamily,
			Effects: []Effect{{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE,
				Misc: c.misc, Misc2: c.misc2, BasePoints: 10}}}
		parsed := ParseStatic(parseWarrior(), &row)
		if got := strings.Join(appliedKinds(parsed), ", "); got != c.kind {
			t.Errorf("misc %d and %d: kinds = %q, want %q", c.misc, c.misc2, got, c.kind)
		}
	}

	demonicEmbrace := Spell{ID: 5000, Name: "Table Row", ClassFlags: parseFamily,
		Effects: []Effect{
			{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE, Misc2: 4, BasePoints: 15},
			{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE, Misc: 4, BasePoints: -1},
		}}
	parsed := ParseStatic(parseWarrior(), &demonicEmbrace)
	if got, want := strings.Join(appliedKinds(parsed), ", "), "multiply-stat Stamina, multiply-stat Spirit"; got != want {
		t.Errorf("a mask effect beside a stat index effect: kinds = %q, want %q", got, want)
	}
}

// A row of one aura effect, in the shape the table reads.
func oneEffectRow(aura dbcenums.EffectAuraType, misc int32, points float64) *Spell {
	return &Spell{ID: 6000, Name: "One Effect", DurationMs: 10000, ClassFlags: parseFamily,
		Effects: []Effect{{Type: dbcenums.E_APPLY_AURA, Aura: aura, Misc: misc, BasePoints: points,
			ClassFlags: parseTargetMask}}}
}

// Each stat the table writes is stored in the units core reads it in, and those units differ: a
// rating, a percentage point and a fraction all come out of a game files percentage.
func TestParseStaticStatConventions(t *testing.T) {
	withRows(t, parseRows())

	dodge := parseWarrior()
	ParseStatic(dodge, oneEffectRow(dbcenums.A_MOD_DODGE_PERCENT, 0, 5))
	if got := dodge.GetDodgeFromRating(); math.Abs(got-0.05) > 1e-9 {
		t.Errorf("dodge from a 5 point row: %v, want 0.05", got)
	}

	block := parseWarrior()
	ParseStatic(block, oneEffectRow(dbcenums.A_MOD_BLOCK_PERCENT, 0, 5))
	if got := block.GetBlockFromRating(); math.Abs(got-0.05) > 1e-9 {
		t.Errorf("block from a 5 point row: %v, want 0.05", got)
	}

	crit := parseWarrior()
	baseCrit := crit.GetStat(stats.PhysicalCritPercent)
	ParseStatic(crit, oneEffectRow(dbcenums.A_MOD_WEAPON_CRIT_PERCENT, 0, 5))
	if got := crit.GetStat(stats.PhysicalCritPercent) - baseCrit; math.Abs(got-5) > 1e-9 {
		t.Errorf("crit from a 5 point row: %v percentage points, want 5", got)
	}

	parry := parseWarrior()
	ParseStatic(parry, oneEffectRow(dbcenums.A_MOD_PARRY_PERCENT, 0, 5))
	if got := parry.GetStat(stats.ParryPercent); math.Abs(got-5) > 1e-9 {
		t.Errorf("parry from a 5 point row: %v percentage points, want 5", got)
	}

	defense := parseWarrior()
	ParseStatic(defense, oneEffectRow(dbcenums.A_MOD_SKILL, miscDefense, 20))
	if got := defense.GetStat(stats.DefenseRating); math.Abs(got-20*core.DefenseRatingPerDefenseLevel) > 1e-9 {
		t.Errorf("defense rating from a 20 point Defense skill row: %v, want %v", got, 20*core.DefenseRatingPerDefenseLevel)
	}
	swords := parseWarrior()
	ParseStatic(swords, oneEffectRow(dbcenums.A_MOD_SKILL, 43, 5))
	if got := swords.PseudoStats.WeaponSkills[proto.WeaponSkill_WeaponSkillSwords]; got != 5 {
		t.Errorf("sword skill from a 5 point Swords skill row: %v, want 5", got)
	}
	fishing := parseWarrior()
	if parsed := ParseStatic(fishing, oneEffectRow(dbcenums.A_MOD_SKILL, 356, 25)); len(parsed.Applied) != 0 {
		t.Errorf("a Fishing skill row attached %v, want nothing", parsed.Applied)
	}

	expertise := parseWarrior()
	ParseStatic(expertise, oneEffectRow(dbcenums.A_MOD_EXPERTISE, 0, 5))
	if got := expertise.GetStat(stats.ExpertisePercent); math.Abs(got-5) > 1e-9 {
		t.Errorf("expertise from a 5 point row: %v percent, want 5", got)
	}
}

// The armor modifier scales the equipment share of the stat, which is what the tooltip states.
func TestParseStaticScalesEquippedArmor(t *testing.T) {
	withRows(t, parseRows())

	character := parseCharacter(&proto.Player{
		Class:      proto.Class_ClassWarrior,
		Spec:       &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
		BonusStats: &proto.UnitStats{Stats: bonusArmor(1000)},
	})

	if got := character.EquipStats()[stats.Armor]; got != 1000 {
		t.Fatalf("armor before the row: %v, want the bonus 1000", got)
	}

	ParseStatic(character, oneEffectRow(dbcenums.A_MOD_BASE_RESISTANCE_PCT, 1, 10))

	if got := character.EquipStats()[stats.Armor]; got != 1100 {
		t.Errorf("armor after the row: %v, want 1100", got)
	}
}

// The percent armor modifier scales the resistance aura's armor and leaves the no-mods aura's armor
// alone, which still counts in Armor.
func TestParseStaticArmorNoModsOutsideThePercentModifier(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()
	before := character.SortAndApplyStatDependencies(character.GetStats())[stats.Armor]

	ParseStatic(character, &Spell{ID: 6000, Name: "Armor Rows", ClassFlags: parseFamily, Effects: []Effect{
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RESISTANCE, Misc: miscArmor, BasePoints: 1000},
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_ARMOR_NO_MODS, Misc: miscArmor, BasePoints: 500},
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RESISTANCE_PCT, Misc: miscArmor, BasePoints: 10},
	}})

	got := character.SortAndApplyStatDependencies(character.GetStats())[stats.Armor]
	if want := (before+1000)*1.1 + 500; math.Abs(got-want) > 1e-9 {
		t.Errorf("armor: %v, want %v from %v before", got, want, before)
	}
}

// The no-mods armor row takes armor alone: a mask naming another school is skipped with a reason
// rather than parsed for its armor bit.
func TestParseArmorNoModsSkipsAMaskBeyondArmor(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	parsed := ParseStatic(character, &Spell{ID: 6000, Name: "Armor Rows", ClassFlags: parseFamily, Effects: []Effect{
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_ARMOR_NO_MODS, Misc: miscArmor | 4, BasePoints: 500},
	}})
	if len(parsed.Applied) != 0 || len(parsed.Skipped) != 1 || parsed.Skipped[0].Reason == "" {
		t.Errorf("attached %v and skipped %v, want the effect skipped with a reason", appliedKinds(parsed), parsed.Skipped)
	}
}

// A share of a stat reaches the stat it names, read from whichever misc value the aura keeps its
// stat in: the first for armor, attack power and healing, the second for spell damage, whose first is
// its schools.
func TestParseStaticStatShare(t *testing.T) {
	withRows(t, parseRows())

	row := &Spell{ID: 6100, Name: "Stat Share", ClassFlags: parseFamily, Effects: []Effect{
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_ARMOR_PCT_FROM_STAT, Misc: 3, BasePoints: 50},
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_SPELL_DAMAGE_OF_STAT_PERCENT, Misc: 126, Misc2: 4,
			BasePoints: 8},
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_SPELL_HEALING_OF_STAT_PERCENT, Misc: 4, BasePoints: 25},
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_ATTACK_POWER_OF_STAT_PERCENT, Misc: 1, BasePoints: 100},
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RANGED_ATTACK_POWER_OF_STAT_PERCENT, Misc: 3,
			BasePoints: 20},
	}}

	base, parsed := parseWarrior(), parseWarrior()
	ParseStatic(parsed, row)

	source := stats.Stats{stats.Agility: 100, stats.Intellect: 100, stats.Spirit: 100}
	gained := parsed.SortAndApplyStatDependencies(source).Subtract(base.SortAndApplyStatDependencies(source))

	for stat, want := range map[stats.Stat]float64{stats.Armor: 50, stats.SpellDamage: 8, stats.HealingPower: 25,
		stats.AttackPower: 100, stats.RangedAttackPower: 20} {
		if got := gained[stat]; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s from 100 Agility, Intellect and Spirit: %v, want %v", stat.StatName(), got, want)
		}
	}
}

func bonusArmor(amount float64) []float64 {
	values := make([]float64, len(proto.Stat_name))
	values[proto.Stat_StatArmor] = amount
	return values
}

// A parse narrowed to the dot modifier alone has no hit modifier to fold it into.
func TestParseFoldsOnlyIntoAHitModifierItReads(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	parsed := ParseStatic(character, Find(1100), Effects(2))

	if kinds := appliedKinds(parsed); len(kinds) != 1 || kinds[0] != "SpellMod_DotDamageDone_Pct" {
		t.Errorf("Effects(2) attached %v, want the dot modifier", kinds)
	}
}

// A cooldown multiplier cannot follow the stacks any more than the other multiplier rows can.
func TestParseSkipsACooldownMultiplierOnAStackingRow(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	row := oneEffectRow(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_COOLDOWN), -20)
	row.MaxStack = 3

	parsed := ParseEffects(character, character.RegisterAura(AuraConfig(row)), row)

	if len(parsed.Applied) != 0 || len(parsed.Skipped) != 1 {
		t.Errorf("a cooldown multiplier on a stacking aura attached %v", appliedKinds(parsed))
	}
}

// An aura that is already up when it is parsed takes what the rows say at once, except for the rows
// that read a Simulation to act.
func TestParseEffectsCatchesUpAnActiveAura(t *testing.T) {
	withRows(t, parseRows())
	sim := &core.Simulation{}
	character := parseWarrior()
	spell := parseSpell(character, 2500, core.SpellSchoolPhysical, core.ProcMaskSpellDamage, parseTargetMask)

	row := *Find(1200)
	row.MaxStack = 0

	aura := character.RegisterAura(AuraConfig(&row))
	aura.Activate(sim)
	ParseEffects(character, aura, &row)

	if got := spell.DamageMultiplierAdditive; got != 1.05 {
		t.Errorf("the modifier on an aura that was already up: %v, want 1.05", got)
	}
	if got := character.PseudoStats.ThreatMultiplier; got != 1.3 {
		t.Errorf("the threat multiplier on an aura that was already up: %v, want 1.3", got)
	}

	haste := character.RegisterAura(AuraConfig(Find(1300)))
	haste.Activate(sim)
	ParseEffects(character, haste, Find(1300))

	if got := character.PseudoStats.MeleeSpeedMultiplier; got != 1 {
		t.Errorf("melee speed on an aura that was already up: %v, want 1 until it is applied again", got)
	}
}

// A debuff sits on the enemy, and its amount is still the caster's: the value scales by the
// character's level, not by the level of the unit the aura is on.
func TestParseEffectsScalesByTheCharactersLevel(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	row := &Spell{ID: 6100, Name: "Enemy Debuff", DurationMs: 30000,
		Effects: []Effect{{SpellID: 6100, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_THREAT,
			Misc: 127, BasePoints: 30, PPL: 1, BaseLevel: 10}}}

	enemy := &core.Unit{Level: character.Level + 3}
	parsed := ParseEffects(character, &core.Aura{Unit: enemy}, row)

	if len(parsed.Applied) != 1 {
		t.Fatalf("the debuff attached %v, want the threat multiplier alone", appliedKinds(parsed))
	}
	// 30 plus a point for each of the 50 levels between the spell's own 10 and the character's 60.
	if got := parsed.Applied[0].Value; got != 1.8 {
		t.Errorf("a debuff on a level %d enemy is worth %v, want the character's 1.8",
			enemy.Level, got)
	}
}

// A multiplier of zero or less cannot be taken back off: expiry divides by it. The row is reported
// the way an unmapped one is rather than leaving the field at zero or at an infinity.
func TestParseSkipsANonPositiveMultiplier(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()

	row := &Spell{ID: 6200, Name: "Not There", DurationMs: 10000,
		Effects: []Effect{
			{SpellID: 6200, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_THREAT,
				Misc: 127, BasePoints: -100},
			{SpellID: 6200, Index: 1, Type: dbcenums.E_APPLY_AURA,
				Aura: dbcenums.A_MOD_BASE_RESISTANCE_PCT, Misc: miscArmor, BasePoints: -100},
		}}

	parsed := ParseStatic(character, row)

	if len(parsed.Applied) != 0 || len(parsed.Skipped) != 2 {
		t.Fatalf("the -100%% rows attached %v, want both of them skipped", appliedKinds(parsed))
	}
	if got := character.PseudoStats.ThreatMultiplier; got != 1 {
		t.Errorf("the threat multiplier is %v, want the 1 it started at", got)
	}
}

// A school cost modifier reaches the spells of the schools its mask names and no others.
func TestParseStaticSchoolCostModifier(t *testing.T) {
	withRows(t, parseRows())
	character := parseCharacter(&proto.Player{
		Class: proto.Class_ClassMage,
		Spec:  &proto.Player_Mage{Mage: &proto.Mage{}},
	})
	manaSpell := func(id int32, school core.SpellSchool) *core.Spell {
		return character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: id},
			SpellSchool: school,
			ManaCost:    core.ManaCostOptions{FlatCost: 100},
		})
	}
	arcane, physical := manaSpell(2600, core.SpellSchoolArcane), manaSpell(2601, core.SpellSchoolPhysical)

	ParseStatic(character, oneEffectRow(dbcenums.A_MOD_POWER_COST_SCHOOL_PCT, miscMagicSchool, 20))

	if got := arcane.Cost.GetCurrentCost(); math.Abs(got-120) > 1e-9 {
		t.Errorf("an arcane spell costs %v, want 120", got)
	}
	if got := physical.Cost.GetCurrentCost(); got != 100 {
		t.Errorf("a physical spell costs %v, want the 100 the magic mask leaves alone", got)
	}
}

// A threat modifier on some schools scales the whole threat of the spells whose first school it names,
// the damage share and the flat bonus, and leaves the other schools and the unit's multiplier alone.
func TestParseStaticSchoolThreatModifier(t *testing.T) {
	withRows(t, parseRows())
	character := parseWarrior()
	threatFrom := func(id int32, school core.SpellSchool) float64 {
		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: id},
			SpellSchool:      school,
			ProcMask:         core.ProcMaskSpellDamage,
			ThreatMultiplier: 1,
			FlatThreatBonus:  100,
		})
		return spell.ThreatFromDamage(nil, core.OutcomeHit, 100, nil)
	}

	ParseStatic(character, oneEffectRow(dbcenums.A_MOD_THREAT, int32(core.SpellSchoolNature|core.SpellSchoolArcane), -30))

	for _, c := range []struct {
		name   string
		school core.SpellSchool
		want   float64
	}{
		{"nature", core.SpellSchoolNature, 140},
		{"physical", core.SpellSchoolPhysical, 200},
		{"physical and nature", core.SpellSchoolPhysical | core.SpellSchoolNature, 200},
	} {
		if got := threatFrom(2700+int32(c.school), c.school); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("a %s spell's 100 damage and 100 flat make %v threat, want %v", c.name, got, c.want)
		}
	}
	if got := character.PseudoStats.ThreatMultiplier; got != 1 {
		t.Errorf("the unit's threat multiplier is %v, want the 1 a school mask leaves alone", got)
	}
}

// Flat damage taken lands on the physical field, the magic field or both, by the mask; a single magic
// school has no field of its own and is skipped.
func TestParseStaticFlatDamageTakenBySchool(t *testing.T) {
	withRows(t, parseRows())

	for _, c := range []struct {
		misc            int32
		physical, magic float64
	}{
		{miscAllSchools, -10, -10},
		{miscMagicSchool, 0, -10},
		{int32(core.SpellSchoolPhysical), -10, 0},
	} {
		character := parseWarrior()
		ParseStatic(character, oneEffectRow(dbcenums.A_MOD_DAMAGE_TAKEN, c.misc, -10))

		if got := character.PseudoStats.BonusPhysicalDamageTaken; got != c.physical {
			t.Errorf("misc %d: physical damage taken %v, want %v", c.misc, got, c.physical)
		}
		if got := character.PseudoStats.BonusSpellDamageTaken; got != c.magic {
			t.Errorf("misc %d: magic damage taken %v, want %v", c.misc, got, c.magic)
		}
	}

	fire := ParseStatic(parseWarrior(), oneEffectRow(dbcenums.A_MOD_DAMAGE_TAKEN, int32(core.SpellSchoolFire), -10))
	if len(fire.Applied) != 0 || len(fire.Skipped) != 1 {
		t.Errorf("a fire-only mask attached %v, want it skipped", appliedKinds(fire))
	}
}

// Target resistance by a mask of some schools lands on the per-school piercing of each school it names,
// its size whichever sign the row states, and leaves SpellPiercing and the other schools alone.
func TestParseStaticSchoolSpellPiercing(t *testing.T) {
	withRows(t, parseRows())

	for _, c := range []struct {
		misc        int32
		fire, frost float64
	}{
		{int32(core.SpellSchoolFire | core.SpellSchoolFrost), 10, 10},
	} {
		character := parseWarrior()
		ParseStatic(character, oneEffectRow(dbcenums.A_MOD_TARGET_RESISTANCE, c.misc, -10))

		piercing := character.PseudoStats.SchoolSpellPiercing
		if piercing[stats.SchoolIndexFire] != c.fire || piercing[stats.SchoolIndexFrost] != c.frost {
			t.Errorf("misc %d: fire and frost piercing %v and %v, want %v and %v", c.misc,
				piercing[stats.SchoolIndexFire], piercing[stats.SchoolIndexFrost], c.fire, c.frost)
		}
		if got := piercing[stats.SchoolIndexArcane] + piercing[stats.SchoolIndexShadow]; got != 0 {
			t.Errorf("misc %d: arcane and shadow piercing %v, want 0", c.misc, got)
		}
		if got := character.GetStat(stats.SpellPiercing); got != 0 {
			t.Errorf("misc %d: SpellPiercing %v, want 0", c.misc, got)
		}
	}
}

// A pet wears no equipment, so the armor modifier has no share of its armor to scale.
func TestParseSkipsEquipScalingOnAPet(t *testing.T) {
	withRows(t, parseRows())
	pet := parseWarrior()
	pet.Unit.Type = core.PetUnit

	parsed := ParseStatic(pet, oneEffectRow(dbcenums.A_MOD_BASE_RESISTANCE_PCT, miscArmor, 10))

	if len(parsed.Applied) != 0 || len(parsed.Skipped) != 1 {
		t.Errorf("the armor modifier on a pet attached %v, want it skipped", appliedKinds(parsed))
	}
}

// A power regen percentage on the energy bar multiplies the energy regen while the aura is up. A unit
// with no energy bar, the static path and a power type the sim keeps no regen multiplier for skip it.
func TestParseEnergyRegenPercentage(t *testing.T) {
	withRows(t, parseRows())
	sim := &core.Simulation{}
	energy := int32(dbcenums.POWER_ENERGY)

	character := parseWarrior()
	character.EnableEnergyBar(core.EnergyBarOptions{MaxEnergy: 100})
	row := oneEffectRow(dbcenums.A_MOD_POWER_REGEN_PERCENT, energy, 100)
	aura := character.RegisterAura(AuraConfig(row))
	parsed := ParseEffects(character, aura, row)

	if kinds := appliedKinds(parsed); len(kinds) != 1 || kinds[0] != "energy-regen" || parsed.Applied[0].Value != 2 {
		t.Fatalf("attached %v, want energy-regen at 2", parsed.Applied)
	}
	aura.Activate(sim)
	if got := character.CurrentEnergyRegenMultiplier(); got != 2 {
		t.Errorf("the energy regen multiplier while the aura is up: %v, want 2", got)
	}
	aura.Deactivate(sim)
	if got := character.CurrentEnergyRegenMultiplier(); got != 1 {
		t.Errorf("the energy regen multiplier after the aura dropped: %v, want 1", got)
	}

	noBar := parseWarrior()
	rage := oneEffectRow(dbcenums.A_MOD_POWER_REGEN_PERCENT, int32(dbcenums.POWER_RAGE), 100)
	for name, parsed := range map[string]*Parsed{
		"no energy bar": ParseEffects(noBar, noBar.RegisterAura(AuraConfig(row)), row),
		"static":        ParseStatic(character, row),
		"rage":          ParseEffects(character, character.RegisterAura(core.Aura{Label: "Rage Regen"}), rage),
	} {
		if len(parsed.Applied) != 0 || len(parsed.Skipped) != 1 {
			t.Errorf("%s: attached %v, want the effect skipped", name, appliedKinds(parsed))
		}
	}
}

func enablePowerBar(character *core.Character, power dbcenums.PowerType) {
	switch power {
	case dbcenums.POWER_RAGE:
		character.EnableRageBar(core.RageBarOptions{MaxRage: 100, BaseRageMultiplier: 1})
	case dbcenums.POWER_ENERGY:
		character.EnableEnergyBar(core.EnergyBarOptions{MaxEnergy: 100, MaxComboPoints: 5, HasNoRegen: true})
	}
}

// Boundless Rage's shape: the game files' 300 is 30 rage on top of the bar's 100. An aura carrying it
// keeps the rage held past the old maximum while it is up and clamps it back when it drops.
func TestParseMaxRage(t *testing.T) {
	row := oneEffectRow(dbcenums.A_MOD_MAX_POWER, int32(dbcenums.POWER_RAGE), 300)

	talent := parseWarrior()
	enablePowerBar(talent, dbcenums.POWER_RAGE)
	ParseStatic(talent, row)
	if got := talent.MaximumRage(); got != 130 {
		t.Errorf("maximum rage with the talent: %v, want 130", got)
	}

	sim := &core.Simulation{Options: &proto.SimOptions{Interactive: true}}
	character := parseWarrior()
	enablePowerBar(character, dbcenums.POWER_RAGE)
	aura := character.RegisterAura(AuraConfig(row))
	ParseEffects(character, aura, row)

	aura.Activate(sim)
	character.AddRage(sim, 120, character.RageRefundMetrics)
	if character.MaximumRage() != 130 || character.CurrentRage() != 120 {
		t.Fatalf("with the aura up: %v of %v rage, want 120 of 130", character.CurrentRage(), character.MaximumRage())
	}

	aura.Deactivate(sim)
	if character.MaximumRage() != 100 || character.CurrentRage() != 100 {
		t.Errorf("after the aura dropped: %v of %v rage, want 100 of 100", character.CurrentRage(), character.MaximumRage())
	}
}

// Mana is the Mana stat on any unit; rage and energy need the bar they raise, except on a dry run.
func TestParseMaxPowerByPowerType(t *testing.T) {
	for _, aura := range []dbcenums.EffectAuraType{dbcenums.A_MOD_MAX_POWER, dbcenums.A_MOD_INCREASE_ENERGY} {
		character := parseWarrior()
		before := character.GetStats()[stats.Mana]
		ParseStatic(character, oneEffectRow(aura, int32(dbcenums.POWER_MANA), 2000))
		if got := character.GetStats()[stats.Mana] - before; got != 2000 {
			t.Errorf("%s misc mana added %v Mana, want 2000", auraName(aura), got)
		}

		for _, power := range []dbcenums.PowerType{dbcenums.POWER_RAGE, dbcenums.POWER_ENERGY} {
			row := oneEffectRow(aura, int32(power), 10)
			if parsed := ParseStatic(parseWarrior(), row); len(parsed.Applied) != 0 {
				t.Errorf("%s misc %d on a unit without that bar attached %v", auraName(aura), power, appliedKinds(parsed))
			}
			if parsed := DryRun(row); len(parsed.Skipped) != 0 {
				t.Errorf("%s misc %d on a dry run skipped %d effects, want none", auraName(aura), power, len(parsed.Skipped))
			}
		}
	}
}
