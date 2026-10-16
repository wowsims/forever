package spelldata

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/stats"
)

// A warrior whose auras can move its stats during a test: a dynamic stat change needs an environment
// that is measuring stats.
func buffTestWarrior() *core.Character {
	character := parseWarrior()
	character.Env = &core.Environment{MeasuringStats: true}
	return character
}

func auraEffect(aura dbcenums.EffectAuraType, misc int32, points float64) Effect {
	return Effect{Type: dbcenums.E_APPLY_AURA, Aura: aura, Misc: misc, BasePoints: points}
}

func buffRow(id int32, effects ...Effect) *Spell {
	for i := range effects {
		effects[i].SpellID = id
		effects[i].Index = uint8(i)
	}
	return &Spell{ID: id, Name: fmt.Sprintf("Buff %d", id), DurationMs: -1, Effects: effects}
}

func categoryNames(aura *core.Aura) []string {
	var names []string
	for _, ee := range aura.ExclusiveEffects {
		names = append(names, ee.Category.Name)
	}
	return names
}

// The rows BuffAuras turns on, each read the way the table rows are, and none of them read without it.
func TestBuffAuraRows(t *testing.T) {
	cases := []struct {
		aura   dbcenums.EffectAuraType
		effect Effect
		kind   string
		value  float64
	}{
		{dbcenums.A_PERIODIC_ENERGIZE,
			Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PERIODIC_ENERGIZE, BasePoints: 10, PeriodMs: 2000},
			"stat MP5", 25},
		{dbcenums.A_MOD_ATTACK_POWER_PCT, auraEffect(dbcenums.A_MOD_ATTACK_POWER_PCT, 0, 10),
			"multiply-stat AttackPower", 1.1},
	}

	covered := map[dbcenums.EffectAuraType]bool{}
	for _, c := range cases {
		covered[c.aura] = true
	}
	for aura := range buffAuraTable {
		if !covered[aura] {
			t.Errorf("the buff table carries %s and no case reads it", auraName(aura))
		}
	}

	for _, c := range cases {
		t.Run(auraName(c.aura), func(t *testing.T) {
			row := buffRow(7000, c.effect)

			without := DryRun(row)
			if len(without.Applied) != 0 || len(without.Skipped) != 1 {
				t.Errorf("without BuffAuras the row attached %v", appliedKinds(without))
			}

			with := DryRun(row, BuffAuras())
			if len(with.Applied) != 1 {
				t.Fatalf("with BuffAuras the row attached %v", appliedKinds(with))
			}
			if got := with.Applied[0]; got.Kind != c.kind || math.Abs(got.Value-c.value) > 1e-9 {
				t.Errorf("attached %s %v, want %s %v", got.Kind, got.Value, c.kind, c.value)
			}
		})
	}
}

// The rage a warrior's own row ticks is not mana, and a buff row that ticks nothing has no rate.
func TestBuffManaTicksAreManaOnly(t *testing.T) {
	rage := buffRow(7010, Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PERIODIC_ENERGIZE,
		Misc: int32(dbcenums.POWER_RAGE), BasePoints: 10, PeriodMs: 1000})
	if parsed := DryRun(rage, BuffAuras()); len(parsed.Applied) != 0 {
		t.Errorf("a rage tick attached %v", appliedKinds(parsed))
	}

	untimed := buffRow(7011, Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PERIODIC_ENERGIZE, BasePoints: 10})
	if parsed := DryRun(untimed, BuffAuras()); len(parsed.Applied) != 0 {
		t.Errorf("a tick with no period attached %v", appliedKinds(parsed))
	}
}

// Flat damage taken has a field for physical and one for every spell school together, and nothing for
// a mask of some spell schools: Judgement of the Crusader's holy-only row is left out.
func TestFlatDamageTakenNeedsAField(t *testing.T) {
	if parsed := DryRun(buffRow(7020, auraEffect(dbcenums.A_MOD_DAMAGE_TAKEN, 2, 161))); len(parsed.Applied) != 0 {
		t.Errorf("a holy-only flat damage taken row attached %v", appliedKinds(parsed))
	}
}

// Demoralizing Shout rank 5 on a level 63 boss: without a character the unit's own level prices it,
// three levels too high, and Level prices it at the caster's.
func TestLevelPricesADebuffOnABoss(t *testing.T) {
	row := &Spell{ID: 11556, Name: "Demoralizing Shout", DurationMs: 45000,
		Effects: []Effect{{SpellID: 11556, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_ATTACK_POWER,
			BasePoints: -196, PPL: -1.399999976158142, BaseLevel: 54, MaxLevel: 64}}}

	boss := &core.Unit{Level: 63}

	if got := ParseEffects(nil, &core.Aura{Unit: boss}, row).Applied[0].Value; got != -209 {
		t.Fatalf("the unpriced debuff reads %v on a level 63 boss, want the boss's own -209", got)
	}
	if got := ParseEffects(nil, &core.Aura{Unit: boss}, row, Level(60)).Applied[0].Value; got != -205 {
		t.Errorf("Level(60) reads %v on a level 63 boss, want -205", got)
	}
	if got := DryRun(row).Applied[0].Value; got != -205 {
		t.Errorf("the dry run reads %v, want the -205 of core.CharacterLevel", got)
	}
}

// Mana Spring's talent adds a percentage of the tick and truncates it before the tick becomes mana
// per five: 10 a tick at +5% is still 10, at +25% it is 12.
func TestScaledByTruncatesTheRowAmount(t *testing.T) {
	row := buffRow(7030, Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PERIODIC_ENERGIZE,
		BasePoints: 10, PeriodMs: 2000})
	pct := func(points float64) *Effect {
		return &Effect{Aura: dbcenums.A_ADD_PCT_MODIFIER, BasePoints: points}
	}

	cases := []struct {
		mod  *Effect
		want float64
	}{
		{NilEffect, 25},
		{pct(5), 25},
		{pct(25), 30},
		{&Effect{Aura: dbcenums.A_ADD_FLAT_MODIFIER, BasePoints: 2}, 30},
	}
	for _, c := range cases {
		if got := DryRun(row, BuffAuras(), ScaledBy(c.mod)).Applied[0].Value; got != c.want {
			t.Errorf("ScaledBy(%v %v) reads %v, want %v", c.mod.Aura, c.mod.BasePoints, got, c.want)
		}
	}
}

// The talent prices the first effect the table takes, stepping over one it does not.
func TestScaledByPricesTheFirstAttachedEffect(t *testing.T) {
	row := buffRow(7031,
		auraEffect(dbcenums.A_DUMMY, 0, 100),
		auraEffect(dbcenums.A_MOD_ATTACK_POWER, 0, 100),
		auraEffect(dbcenums.A_MOD_RANGED_ATTACK_POWER, 0, 100))

	parsed := DryRun(row, ScaledBy(&Effect{Aura: dbcenums.A_ADD_PCT_MODIFIER, BasePoints: 10}))
	if len(parsed.Applied) != 2 || parsed.Applied[0].Value != 110 || parsed.Applied[1].Value != 100 {
		t.Errorf("attached %v, want the melee attack power alone at 110", parsed.Applied)
	}
}

// Expose Armor states its armor per combo point and nothing on the spell itself.
func TestFullComboPoints(t *testing.T) {
	row := buffRow(11198, Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RESISTANCE,
		Misc: miscArmor, PointsPerResource: -450})

	without := DryRun(row)
	if len(without.Applied) != 0 || len(without.Skipped) != 1 {
		t.Errorf("without FullComboPoints the row attached %v", appliedKinds(without))
	}
	if notes := without.SkippedNotes(row); len(notes) != 1 || !strings.Contains(notes[0], "per combo point") {
		t.Errorf("the note for the per-point row: %v", notes)
	}

	if got := DryRun(row, FullComboPoints()).Applied[0].Value; got != -2250 {
		t.Errorf("the finisher at full combo points reads %v, want -2250", got)
	}
}

// Atiesh is one aura for every staff in the party, so the flat amounts count once per staff and a
// multiplier, which cannot be counted, is left out.
func TestCountScalesTheAmounts(t *testing.T) {
	sim := &core.Simulation{}
	character := buffTestWarrior()
	before := character.GetStat(stats.SpellCritPercent)

	row := buffRow(7040,
		auraEffect(dbcenums.A_MOD_SPELL_CRIT_CHANCE, 0, 2),
		auraEffect(dbcenums.A_MOD_THREAT, 127, 10))

	aura := character.RegisterAura(AuraConfig(row))
	parsed := ParseEffects(character, aura, row, Count(3))
	if kinds := appliedKinds(parsed); len(kinds) != 1 || len(parsed.Skipped) != 1 {
		t.Fatalf("attached %v, want the crit alone", kinds)
	}

	aura.Activate(sim)
	if got := character.GetStat(stats.SpellCritPercent) - before; got != 6 {
		t.Errorf("three staves give %v spell crit, want 6", got)
	}
	aura.Deactivate(sim)
	if got := character.GetStat(stats.SpellCritPercent); got != before {
		t.Errorf("spell crit after the aura fell off: %v, want %v", got, before)
	}
}

// Concentration Aura's healing and mechanic rows state 0, and the buff leaves them out.
func TestSkipAuras(t *testing.T) {
	row := buffRow(19746,
		auraEffect(dbcenums.A_REDUCE_PUSHBACK, 127, 35),
		auraEffect(dbcenums.A_MOD_HEALING_PCT, 0, 0),
		auraEffect(dbcenums.A_MECHANIC_DURATION_MOD, 26, 0),
		auraEffect(dbcenums.A_MECHANIC_DURATION_MOD, 9, 0))

	all := DryRun(row)
	if kinds := appliedKinds(all); !slices.Equal(kinds, []string{"pushback", "healing-taken"}) {
		t.Errorf("without SkipAuras the row attached %v", kinds)
	}

	skipped := DryRun(row, SkipAuras(dbcenums.A_MOD_HEALING_PCT, dbcenums.A_MECHANIC_DURATION_MOD))
	if kinds := appliedKinds(skipped); !slices.Equal(kinds, []string{"pushback"}) || len(skipped.Skipped) != 0 {
		t.Errorf("with SkipAuras the row attached %v and reported %d", kinds, len(skipped.Skipped))
	}
	if notes := skipped.SkippedNotes(row); len(notes) != 0 {
		t.Errorf("the skipped auras are noted as left out: %v", notes)
	}
}

// Two copies of a buff whose category holds one aura at a time: the stronger one applies and shuts
// the weaker one off, and the category carries the bare name the hand-written members use.
func TestExclusiveSingleAura(t *testing.T) {
	sim := &core.Simulation{}
	character := buffTestWarrior()
	before := character.GetStat(stats.AttackPower)

	register := func(label string, points float64) *core.Aura {
		row := buffRow(7050, auraEffect(dbcenums.A_MOD_ATTACK_POWER, 0, points))
		aura := character.RegisterAura(AuraConfig(row, Label(label)))
		ParseEffects(character, aura, row, Exclusive("BattleShout", true))
		return aura
	}
	weaker, stronger := register("Weaker", 139), register("Stronger", 169)

	if names := categoryNames(weaker); !slices.Equal(names, []string{"BattleShout"}) {
		t.Fatalf("the aura bids in %v, want the bare category", names)
	}
	if got := weaker.ExclusiveEffects[0].Priority; got != 139 {
		t.Errorf("the weaker copy bids %v, want 139", got)
	}

	weaker.Activate(sim)
	stronger.Activate(sim)
	if weaker.IsActive() {
		t.Error("the weaker copy is still up next to the stronger one")
	}
	if got := character.GetStat(stats.AttackPower) - before; got != 169 {
		t.Errorf("attack power with both cast: +%v, want +169", got)
	}

	stronger.Deactivate(sim)
	if got := character.GetStat(stats.AttackPower); got != before {
		t.Errorf("attack power after both fell off: %v, want %v", got, before)
	}
}

// A stacking debuff bids nothing until its first stack and once per stack after it, and the amounts
// follow the bid.
func TestExclusiveFollowsTheStacks(t *testing.T) {
	sim := &core.Simulation{}
	target := buffTestWarrior()
	before := target.GetStat(stats.ArmorNoMods)

	row := buffRow(11597, auraEffect(dbcenums.A_MOD_ARMOR_NO_MODS, miscArmor, -450))
	row.MaxStack = 5
	aura := target.RegisterAura(AuraConfig(row))
	ParseEffects(nil, aura, row, Exclusive("MajorArmorReduction", true))

	effect := aura.ExclusiveEffects[0]
	if effect.Priority != 0 {
		t.Errorf("the debuff bids %v before its first stack, want 0", effect.Priority)
	}

	aura.Activate(sim)
	aura.AddStack(sim)
	aura.AddStack(sim)
	if effect.Priority != 900 {
		t.Errorf("the debuff bids %v at two stacks, want 900", effect.Priority)
	}
	if got := target.GetStat(stats.ArmorNoMods) - before; got != -900 {
		t.Errorf("armor at two stacks: %v, want -900", got)
	}

	aura.Deactivate(sim)
	if got := target.GetStat(stats.ArmorNoMods); got != before {
		t.Errorf("armor after the debuff fell off: %v, want %v", got, before)
	}
}

// Under a 10% armor multiplier, each stack of the A_MOD_ARMOR_NO_MODS debuff takes its full 450 off
// a finalized unit and gives it back on expiry.
func TestArmorNoModsStacksOutsideThePercentModifier(t *testing.T) {
	withRows(t, parseRows())
	sim := &core.Simulation{}
	target := parseWarrior()
	target.Env = &core.Environment{}
	ParseStatic(target, &Spell{ID: 6000, Name: "Armor Rows", ClassFlags: parseFamily, Effects: []Effect{
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RESISTANCE, Misc: miscArmor, BasePoints: 1000},
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_ARMOR_NO_MODS, Misc: miscArmor, BasePoints: 500},
		{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RESISTANCE_PCT, Misc: miscArmor, BasePoints: 10},
	}})

	row := buffRow(11597, auraEffect(dbcenums.A_MOD_ARMOR_NO_MODS, miscArmor, -450))
	row.MaxStack = 5
	aura := target.RegisterAura(AuraConfig(row))
	ParseEffects(nil, aura, row, Exclusive("MajorArmorReduction", true))

	target.Finalize()
	target.Env.State = core.Finalized
	before := target.GetStat(stats.Armor)

	aura.Activate(sim)
	for stacks := int32(1); stacks <= 2; stacks++ {
		aura.AddStack(sim)
		if got, want := target.GetStat(stats.Armor), before-450*float64(stacks); math.Abs(got-want) > 1e-9 {
			t.Errorf("armor at %d stacks: %v, want %v", stacks, got, want)
		}
	}
	aura.Deactivate(sim)
	if got := target.GetStat(stats.Armor); math.Abs(got-before) > 1e-9 {
		t.Errorf("armor after the debuff fell off: %v, want %v", got, before)
	}
}

// A multiplier bids how far it moves the field, whole-aura or per stat.
func TestExclusivePricesAMultiplierByItsDistanceFromOne(t *testing.T) {
	character := buffTestWarrior()
	row := buffRow(11581, auraEffect(dbcenums.A_MOD_MELEE_HASTE_3, 0, -20))

	aura := character.RegisterAura(AuraConfig(row))
	ParseEffects(character, aura, row, Exclusive("AtkSpdReduction", false))

	// -20 divides the speed by 1.2, which takes a sixth off it.
	if got := aura.ExclusiveEffects[0].Priority; math.Abs(got-1.0/6) > 1e-9 {
		t.Errorf("a 20%% slow bids %v, want 1/6", got)
	}
	if aura.ExclusiveEffects[0].Category.SingleAura {
		t.Error("the category turned single-aura, which the caller did not ask for")
	}
}

// Per stat, each stat and pseudo-stat bids alone under the names core's exclusive stat buffs build,
// so a stronger source of the same stat replaces the weaker one's amount without the aura dropping.
func TestExclusivePerStat(t *testing.T) {
	sim := &core.Simulation{}
	character := buffTestWarrior()
	before := character.GetStat(stats.Intellect)

	register := func(label string, points float64) *core.Aura {
		row := buffRow(7060, auraEffect(dbcenums.A_MOD_STAT, 3, points))
		aura := character.RegisterAura(AuraConfig(row, Label(label)))
		ParseEffects(character, aura, row, ExclusivePerStat("StatBuff"))
		return aura
	}
	brilliance, scroll := register("Brilliance", 31), register("Scroll", 20)

	if names := categoryNames(brilliance); !slices.Equal(names, []string{"StatBuffIntellectAdd"}) {
		t.Fatalf("the aura bids in %v, want StatBuffIntellectAdd", names)
	}

	brilliance.Activate(sim)
	scroll.Activate(sim)
	if !scroll.IsActive() {
		t.Error("the weaker aura was shut off, which only a single-aura category does")
	}
	if got := character.GetStat(stats.Intellect) - before; got != 31 {
		t.Errorf("intellect with both up: +%v, want +31", got)
	}

	infusion := buffRow(10060,
		auraEffect(dbcenums.A_MOD_HEALING_DONE_PERCENT, 126, 20),
		auraEffect(dbcenums.A_MOD_DAMAGE_PERCENT_DONE, 126, 20))
	aura := character.RegisterAura(AuraConfig(infusion))
	ParseEffects(character, aura, infusion, ExclusivePerStat("PowerInfusion"))

	want := []string{"PowerInfusionHealingDealtMultiplierMul", "PowerInfusionSchoolDamageDealtMultiplierMul"}
	if names := categoryNames(aura); !slices.Equal(names, want) {
		t.Errorf("the pseudo-stats bid in %v, want %v", names, want)
	}
	if got := aura.ExclusiveEffects[0].Priority; math.Abs(got-0.2) > 1e-9 {
		t.Errorf("a 20%% multiplier bids %v, want 0.2", got)
	}
}

// Each school's resistance bids in its school's category whatever else the buff does, so a 27 and a
// 60 of the same school do not add up. Armor and the main stats are not schools.
func TestSchoolResistances(t *testing.T) {
	sim := &core.Simulation{}
	character := buffTestWarrior()
	beforeFire, beforeFrost := character.GetStat(stats.FireResistance), character.GetStat(stats.FrostResistance)
	beforeArmor := character.GetStat(stats.Armor)

	wild := buffRow(21850,
		auraEffect(dbcenums.A_MOD_RESISTANCE, miscArmor, 385),
		auraEffect(dbcenums.A_MOD_RESISTANCE, 4, 27),
		auraEffect(dbcenums.A_MOD_RESISTANCE, 16, 27))
	wildAura := character.RegisterAura(AuraConfig(wild))
	ParseEffects(character, wildAura, wild, SchoolResistances())

	want := []string{"ResistanceFireFireResistanceAdd", "ResistanceFrostFrostResistanceAdd"}
	if names := categoryNames(wildAura); !slices.Equal(names, want) {
		t.Fatalf("the resistances bid in %v, want %v", names, want)
	}

	fire := buffRow(19900, auraEffect(dbcenums.A_MOD_RESISTANCE, 4, 60))
	fireAura := character.RegisterAura(AuraConfig(fire))
	ParseEffects(character, fireAura, fire, Exclusive("FireResistanceAura", true), SchoolResistances())

	want = []string{"ResistanceFireFireResistanceAdd", "FireResistanceAura"}
	if names := categoryNames(fireAura); !slices.Equal(names, want) {
		t.Fatalf("the resistance aura bids in %v, want %v", names, want)
	}
	if got := fireAura.ExclusiveEffects[1].Priority; got != 60 {
		t.Errorf("the resistance aura's own category bids %v, want the 60 it grants", got)
	}

	wildAura.Activate(sim)
	fireAura.Activate(sim)
	if got := character.GetStat(stats.FireResistance) - beforeFire; got != 60 {
		t.Errorf("fire resistance with both up: +%v, want +60", got)
	}
	if got := character.GetStat(stats.FrostResistance) - beforeFrost; got != 27 {
		t.Errorf("frost resistance: +%v, want +27", got)
	}
	if got := character.GetStat(stats.Armor) - beforeArmor; got != 385 {
		t.Errorf("armor: +%v, want +385", got)
	}
}

// The dry run answers what the parse attaches and skips, with no unit of the sim's to attach it to.
func TestDryRunMatchesTheParse(t *testing.T) {
	withRows(t, parseRows())

	for _, id := range []int32{1000, 1100, 1200, 1300, 1400} {
		character := parseWarrior()
		aura := character.RegisterAura(AuraConfig(Find(id)))
		parsed := ParseEffects(character, aura, Find(id))
		dry := DryRun(Find(id))

		if !slices.Equal(appliedKinds(dry), appliedKinds(parsed)) || len(dry.Skipped) != len(parsed.Skipped) {
			t.Errorf("spell %d: the dry run attached %v and skipped %d, the parse %v and %d",
				id, appliedKinds(dry), len(dry.Skipped), appliedKinds(parsed), len(parsed.Skipped))
		}
	}
}

// Hunter's Mark states the mark itself as an aura the sim has nothing for, beside the attack power.
func TestSkippedNotesNameWhatIsLeftOut(t *testing.T) {
	row := buffRow(14325,
		auraEffect(dbcenums.A_MOD_STALKED, 0, 0),
		auraEffect(dbcenums.A_RANGED_ATTACK_POWER_ATTACKER_BONUS, 0, 71))

	notes := DryRun(row).SkippedNotes(row)
	if len(notes) != 1 || notes[0] != "effect 1 A_MOD_STALKED(68) misc 0" {
		t.Errorf("notes %q, want the stalked effect alone", notes)
	}
	if notes := DryRun(Nil).SkippedNotes(Nil); len(notes) != 1 {
		t.Errorf("a missing row answered %q", notes)
	}
}
