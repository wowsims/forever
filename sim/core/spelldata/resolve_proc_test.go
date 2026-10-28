package spelldata

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// One row per shape the proc resolver has to answer for: the four chance sources, an override-baked
// procs-per-minute rate with and without a mask to measure it on, and a mask the decoder reads past.
func procRows() []Spell {
	return []Spell{
		{
			ID: 2000, Name: "Column Chance", ProcChance: 5, ProcChanceSource: ProcChanceColumn,
			ProcFlags: [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_SWING | dbcenums.PROC_FLAG_DEAL_MELEE_ABILITY},
			ICDMs:     45000,
			Attr: [17]uint32{
				dbcenums.ATTR_INDEX_EX_3:  dbcenums.ATTR_EX_3_CAN_PROC_FROM_PROCS,
				dbcenums.ATTR_INDEX_EX_6:  dbcenums.ATTR_EX_6_AURA_IS_WEAPON_PROC,
				dbcenums.ATTR_INDEX_EX_12: dbcenums.ATTR_EX_12_ONLY_PROC_FROM_CLASS_ABILITIES,
			},
			Effects: []Effect{
				{SpellID: 2000, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PROC_TRIGGER_SPELL,
					TriggerID: 2900, ClassFlags: core.ClassFlags{Family: 4, Mask: [4]uint32{0: 0x20}}},
			},
		},
		{
			ID: 2100, Name: "Sentinel Chance", ProcChance: 101, ProcChanceSource: ProcChanceAlways,
			ProcFlags: [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_SWING},
		},
		{
			ID: 2200, Name: "Effect Chance", ProcChance: 60, ProcChanceSource: ProcChanceEffectN,
			ProcChanceEffect: 2,
			ProcFlags:        [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_SWING},
			Effects: []Effect{
				{SpellID: 2200, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PROC_TRIGGER_SPELL, TriggerID: 2900},
				{SpellID: 2200, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_DUMMY, BasePoints: 12},
			},
		},
		{
			ID: 2300, Name: "Unstated Rate", ProcChance: 101, ProcChanceSource: ProcChancePPM,
			ProcFlags: [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_SWING},
		},
		{
			ID: 2400, Name: "Override Rate", ProcChance: 101, ProcChanceSource: ProcChancePPM, RPPM: 3,
			ProcFlags: [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_SWING},
		},
		{
			ID: 2500, Name: "Maskless Override Rate", ProcChance: 101, ProcChanceSource: ProcChancePPM, RPPM: 1,
		},
		{
			ID: 2600, Name: "Named Ability", ProcChance: 101, ProcChanceSource: ProcChanceAlways,
			ProcFlags: [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_SWING | dbcenums.PROC_FLAG_DEATH},
			ProcHint:  core.ProcHintNamedAbility,
		},
		{
			ID: 2700, Name: "Cheat Death", ProcChance: 100, ProcChanceSource: ProcChanceAlways,
			ProcFlags: [2]uint32{0: dbcenums.PROC_FLAG_TAKE_MELEE_SWING},
			Effects: []Effect{
				{SpellID: 2700, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PROC_TRIGGER_SPELL,
					TriggerID: 2900, ClassFlags: core.ClassFlags{Family: 5, Mask: [4]uint32{0: 0x8}}},
			},
		},
		{
			ID: 2800, Name: "Potion Family", ProcChance: 100, ProcChanceSource: ProcChanceAlways,
			ProcFlags: [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_SWING},
			Effects: []Effect{
				{SpellID: 2800, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PROC_TRIGGER_SPELL,
					TriggerID: 2900, ClassFlags: core.ClassFlags{Family: 13, Mask: [4]uint32{0: 0x1}}},
			},
		},
	}
}

// The proc managers are all a resolved trigger needs off the character, and one with no weapons
// swinging answers an empty manager rather than reaching for an environment.
func testCharacter() *core.Character {
	return &core.Character{}
}

func noopHandler(*core.Simulation, *core.Spell, *core.SpellResult) {}

func TestProcTriggerFromARow(t *testing.T) {
	withRows(t, procRows())
	trigger := ProcTrigger(testCharacter(), Find(2000), noopHandler)

	if trigger.Name != "Column Chance" {
		t.Errorf("name = %q, want the spell's name", trigger.Name)
	}
	if trigger.ActionID != (core.ActionID{SpellID: 2000}) {
		t.Errorf("ActionID = %v, want spell 2000", trigger.ActionID)
	}
	if trigger.Callback != core.CallbackOnSpellHitDealt {
		t.Errorf("callback = %d, want the decoded hit-dealt", trigger.Callback)
	}
	if trigger.ProcMask != core.ProcMaskMeleeWhiteHit|core.ProcMaskMeleeSpecial {
		t.Errorf("proc mask = %d, want the decoded melee swings and abilities", trigger.ProcMask)
	}
	if trigger.Outcome != core.OutcomeLanded {
		t.Errorf("outcome = %d, want landed", trigger.Outcome)
	}
	if !trigger.RequireDamageDealt {
		t.Error("a melee listener has to require damage dealt, the way the decoder states it")
	}
	if trigger.ICD != time.Second*45 {
		t.Errorf("ICD = %v, want the row's 45s", trigger.ICD)
	}
	if !trigger.CanProcFromProcs || !trigger.ClassSpellsOnly {
		t.Errorf("attribute gates = %v/%v, want both set from the row", trigger.CanProcFromProcs, trigger.ClassSpellsOnly)
	}
	if !trigger.SpellFlagsExclude.Matches(core.SpellFlagSuppressWeaponProcs) {
		t.Error("an aura marked Aura Is Weapon Proc has to skip hits that suppress weapon procs")
	}
	if trigger.ClassFlags != (core.ClassFlags{Family: 4, Mask: [4]uint32{0: 0x20}}) {
		t.Errorf("class flags = %v, want the proc effect's", trigger.ClassFlags)
	}
	if trigger.Handler == nil {
		t.Error("the handler the caller passed was dropped")
	}
}

// A weapon proc rolls off any hit of its weapon that lands, a special that deals no damage
// included, and off nothing that misses or is avoided.
func TestWeaponProcHearsALandedHitThatDealsNoDamage(t *testing.T) {
	withRows(t, procRows())

	procs := 0
	trigger := ProcTrigger(testCharacter(), Find(2100), func(*core.Simulation, *core.Spell, *core.SpellResult) {
		procs++
	}, WeaponProc())
	trigger.TriggerImmediately = true

	aura := &core.Aura{}
	aura.AttachProcTriggerCallback(&core.Unit{}, trigger)

	sim := &core.Simulation{}
	special := &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial}
	for _, row := range []struct {
		name    string
		outcome core.HitOutcome
		want    int
	}{
		{"a landed special that deals no damage", core.OutcomeHit, 1},
		{"a miss", core.OutcomeMiss, 0},
		{"a dodge", core.OutcomeDodge, 0},
		{"a parry", core.OutcomeParry, 0},
	} {
		procs = 0
		aura.OnSpellHitDealt(aura, sim, special, &core.SpellResult{Outcome: row.outcome})
		if procs != row.want {
			t.Errorf("%s rolled the proc %d times, want %d", row.name, procs, row.want)
		}
	}
}

// The column is the roll only where the source says so, and the 101 sentinel is not a roll at all.
func TestProcTriggerChanceBySource(t *testing.T) {
	withRows(t, procRows())
	character := testCharacter()

	if got := ProcTrigger(character, Find(2000), noopHandler).ProcChance; got != 0.05 {
		t.Errorf("column chance = %v, want the row's 5%%", got)
	}
	if got := ProcTrigger(character, Find(2100), noopHandler).ProcChance; got != 1 {
		t.Errorf("sentinel chance = %v, want 1", got)
	}
	if got := ProcTrigger(character, Find(2200), noopHandler).ProcChance; got != 0.12 {
		t.Errorf("effect chance = %v, want the second effect's 12%%, not the column's 60", got)
	}
}

// A row that states no rate is the shape whose real rate lives outside the spell data, so it has to
// be given one rather than firing on every hit.
func TestProcTriggerUnstatedRatePanics(t *testing.T) {
	withRows(t, procRows())

	requirePanic(t, "states no proc chance", func() {
		ProcTrigger(testCharacter(), Find(2300), noopHandler)
	})
}

func TestProcTriggerRateOptions(t *testing.T) {
	withRows(t, procRows())
	character := testCharacter()

	ppm := ProcTrigger(character, Find(2300), noopHandler, PPM(2))
	if ppm.DPM == nil {
		t.Error("PPM() left the trigger without a proc manager")
	}
	if ppm.ProcChance != 0 {
		t.Errorf("PPM() left a chance of %v behind, which would gate the manager", ppm.ProcChance)
	}

	// The options run after the fill, so they win over whatever the row stated.
	if got := ProcTrigger(character, Find(2000), noopHandler, Chance(0.25)).ProcChance; got != 0.25 {
		t.Errorf("chance = %v, want the caller's 25%%", got)
	}
	if got := ProcTrigger(character, Find(2000), noopHandler, ChanceFrom(Find(2200).EffectN(2))).ProcChance; got != 0.12 {
		t.Errorf("chance = %v, want the effect's 12%%", got)
	}
}

// An override-baked rate is a manager rather than a roll, and it needs a mask to measure hits on.
func TestProcTriggerOverriddenProcsPerMinute(t *testing.T) {
	withRows(t, procRows())

	trigger := ProcTrigger(testCharacter(), Find(2400), noopHandler)
	if trigger.DPM == nil {
		t.Error("a row with a procs-per-minute override got no proc manager")
	}
	if trigger.ProcChance != 0 {
		t.Errorf("chance = %v, want none next to the manager", trigger.ProcChance)
	}

	requirePanic(t, "no proc mask", func() {
		ProcTrigger(testCharacter(), Find(2500), noopHandler)
	})
}

// The manager measures the mask the trigger ends on. An option that blanks it - the weapon-proc
// shape, where the weapon says which hits count - leaves nothing to measure, so the row's own rate
// cannot stand in for one.
func TestProcTriggerRateFollowsTheOptionsMask(t *testing.T) {
	withRows(t, procRows())

	requirePanic(t, "no proc mask", func() {
		ProcTrigger(testCharacter(), Find(2400), noopHandler, func(_ *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcMask = core.ProcMaskUnknown
		})
	})

	// Narrowing the mask is not blanking it: the rate is still measurable, on the narrower mask.
	narrowed := ProcTrigger(testCharacter(), Find(2400), noopHandler, func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcMask = core.ProcMaskMeleeMH
	})
	if narrowed.DPM == nil {
		t.Error("narrowing the mask lost the proc manager")
	}
}

// A weapon proc's mask comes from the weapon rather than from the row, so the caller supplies the
// manager and the rate check is satisfied by it.
func TestProcTriggerMasklessRateFromTheCaller(t *testing.T) {
	withRows(t, procRows())
	character := testCharacter()

	trigger := ProcTrigger(character, Find(2500), noopHandler, func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.DPM = character.NewDynamicLegacyProcForWeapon(12798, 1, 0)
	})

	if trigger.DPM == nil {
		t.Error("the caller's manager was dropped")
	}
}

// A death rolls the row's plain chance, which a rated trigger does not keep, so a row hearing hits and
// killing blows keeps the kill only where it states a chance: a rate from the row or the caller drops it.
func TestRatedTriggerHearsNoDeath(t *testing.T) {
	rows := procRows()
	for i := range rows {
		if rows[i].ID == 2100 || rows[i].ID == 2400 {
			rows[i].ProcFlags[0] |= dbcenums.PROC_FLAG_KILL
			rows[i].ProcFlags[1] |= dbcenums.PROC_FLAG_2_TARGET_DIES
		}
	}
	withRows(t, rows)
	character := testCharacter()
	deaths := core.CallbackOnKill | core.CallbackOnTargetDies

	for _, tc := range []struct {
		name      string
		trigger   core.ProcTrigger
		hearsDies bool
	}{
		{"a stated chance", ProcTrigger(character, Find(2100), noopHandler), true},
		{"a rate the row states", ProcTrigger(character, Find(2400), noopHandler), false},
		{"a rate the caller states", ProcTrigger(character, Find(2100), noopHandler, PPM(2)), false},
	} {
		if !tc.trigger.Callback.Matches(core.CallbackOnSpellHitDealt) {
			t.Errorf("%s: callback = %d, want the hit listener kept", tc.name, tc.trigger.Callback)
		}
		if got := tc.trigger.Callback&deaths == deaths; got != tc.hearsDies {
			t.Errorf("%s: hears deaths %v, want %v", tc.name, got, tc.hearsDies)
		}
	}
}

// The trigger is still built for a row the decode reads past.
func TestProcTriggerBuildsPastUnsupportedBitsAndHints(t *testing.T) {
	withRows(t, procRows())

	trigger := ProcTrigger(testCharacter(), Find(2600), noopHandler)
	if trigger.Callback != core.CallbackOnSpellHitDealt {
		t.Errorf("callback = %d, want the supported bits to still build a listener", trigger.Callback)
	}
}

// A class mask is a filter inside one family, so a mask from another family names no spell the
// character can cast. The row is Dreadnaught's 8pc 28845, whose proc effect sits in the warlock's
// family 5 and is worn by every class.
func TestProcTriggerDropsAClassMaskOfAnotherFamily(t *testing.T) {
	withRows(t, procRows())

	warrior := &core.Character{Class: proto.Class_ClassWarrior}
	if got := ProcTrigger(warrior, Find(2700), noopHandler).ClassFlags; !got.IsZero() {
		t.Errorf("class flags = %v, want none: family 5 names nothing a warrior casts", got)
	}

	warlock := &core.Character{Class: proto.Class_ClassWarlock}
	if got := ProcTrigger(warlock, Find(2700), noopHandler).ClassFlags; got != (core.ClassFlags{Family: 5, Mask: [4]uint32{0: 0x8}}) {
		t.Errorf("class flags = %v, want the row's, which is the warlock's own family", got)
	}

	// A class the game files state no family for is no evidence that the mask is somebody else's.
	if got := ProcTrigger(testCharacter(), Find(2700), noopHandler).ClassFlags; got.IsZero() {
		t.Error("class flags were dropped for a character whose class states no family")
	}
}

// A family no class files its spells under names spells everyone can use - 13 is the potions - so
// the mask stays whatever the wearer's class is.
func TestProcTriggerKeepsAMaskOutsideEveryClassFamily(t *testing.T) {
	withRows(t, procRows())

	warrior := &core.Character{Class: proto.Class_ClassWarrior}
	if got := ProcTrigger(warrior, Find(2800), noopHandler).ClassFlags; got != (core.ClassFlags{Family: 13, Mask: [4]uint32{0: 0x1}}) {
		t.Errorf("class flags = %v, want the row's: family 13 is nobody's class family", got)
	}
}

// The family every class files its spells under, against one spell of each in the real store.
func TestClassSpellFamilies(t *testing.T) {
	withGeneratedStore(t)

	for _, row := range []struct {
		class   proto.Class
		spellID int32
		name    string
	}{
		{proto.Class_ClassMage, 10, "Blizzard"},
		{proto.Class_ClassWarrior, 78, "Heroic Strike"},
		{proto.Class_ClassWarlock, 686, "Shadow Bolt"},
		{proto.Class_ClassPriest, 17, "Power Word: Shield"},
		{proto.Class_ClassDruid, 8921, "Moonfire"},
		{proto.Class_ClassRogue, 53, "Backstab"},
		{proto.Class_ClassHunter, 3044, "Arcane Shot"},
		{proto.Class_ClassPaladin, 20271, "Judgement"},
		{proto.Class_ClassShaman, 403, "Lightning Bolt"},
	} {
		if got := MustFind(row.spellID).ClassFlags.Family; got != core.ClassSpellFamilies[row.class] {
			t.Errorf("%s (%d) files under family %d, and the table says class %v is family %d",
				row.name, row.spellID, got, row.class, core.ClassSpellFamilies[row.class])
		}
	}

	if len(core.ClassSpellFamilies) != 9 {
		t.Errorf("the table holds %d classes, and the game files have 9", len(core.ClassSpellFamilies))
	}
}

// A trigger carries its row's own class flags, which is what a SPELLMOD_CHANCE_OF_SUCCESS names:
// Improved Poisons 14113 raises the chance Instant Poison 8680 rolls by the points it states, and
// Moonkin Form 24858 raises the chance of Omen of Clarity 16864 by its percent while the form is up.
func TestAProcChanceModReachesTheTriggerBuiltFromTheRowItNames(t *testing.T) {
	withGeneratedStore(t)

	rogue := parseCharacter(&proto.Player{Class: proto.Class_ClassRogue, Spec: &proto.Player_Rogue{Rogue: &proto.Rogue{}}})
	instantPoison := MustFind(8680)
	trigger := ProcTrigger(rogue, instantPoison, noopHandler, WeaponProc())
	if trigger.ProcSpellClassFlags != instantPoison.ClassFlags {
		t.Fatalf("the trigger's own class flags = %v, want the row's %v", trigger.ProcSpellClassFlags, instantPoison.ClassFlags)
	}
	base := float64(instantPoison.ProcChance) / 100
	if got := rogue.ModifiedProcChance(trigger.ProcSpellClassFlags, trigger.ProcChance); got != base {
		t.Errorf("Instant Poison's chance before the talent: %v, want %v", got, base)
	}

	improvedPoisons := MustFind(14113)
	parsed := ParseStatic(rogue, improvedPoisons)
	if kinds := appliedKinds(parsed); len(kinds) == 0 || kinds[0] != "SpellMod_ProcChance_Flat" {
		t.Fatalf("Improved Poisons attached %v, want the proc chance mod first", kinds)
	}
	if got, want := rogue.ModifiedProcChance(trigger.ProcSpellClassFlags, trigger.ProcChance), base+improvedPoisons.EffectN(1).Percent(); math.Abs(got-want) > 1e-9 {
		t.Errorf("Instant Poison's chance with the talent: %v, want %v", got, want)
	}

	sim := &core.Simulation{}
	druid := parseCharacter(&proto.Player{Class: proto.Class_ClassDruid, Spec: &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{}}})
	moonkin := MustFind(24858)
	form := druid.RegisterAura(AuraConfig(moonkin))
	ParseEffects(druid, form, moonkin, Effects(5))
	omen := MustFind(16864).ClassFlags

	form.Activate(sim)
	if got, want := druid.ModifiedProcChance(omen, 0.1), 0.1*(1+moonkin.EffectN(5).Percent()); math.Abs(got-want) > 1e-9 {
		t.Errorf("Omen of Clarity's chance in Moonkin Form: %v, want %v", got, want)
	}
	form.Deactivate(sim)
	if got := druid.ModifiedProcChance(omen, 0.1); got != 0.1 {
		t.Errorf("Omen of Clarity's chance out of the form: %v, want 0.1", got)
	}
}

// Frenzy 19621 names the proc of Tamed Pet Passive 20784, a row the store does not hold, and
// Improved Judgement of Light 23564 the Judgement of Light debuff 20185 on the target. Both attach to
// the caster: a pet's trigger counts its owner's mods, and the debuff's trigger names the paladin.
func TestAProcChanceModReachesThePetAndTheTargetSideTrigger(t *testing.T) {
	withGeneratedStore(t)

	hunter := parseCharacter(&proto.Player{Class: proto.Class_ClassHunter, Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{}}})
	hunter.Party.Raid = &core.Raid{}
	pet := core.NewPet(core.PetConfig{Name: "Test Pet", Owner: hunter,
		StatInheritance: func(stats.Stats) stats.Stats { return stats.Stats{} }})
	frenzy := MustFind(19621)
	if kinds := appliedKinds(ParseStatic(hunter, frenzy)); len(kinds) != 1 || kinds[0] != "SpellMod_ProcChance_Flat" {
		t.Fatalf("Frenzy attached %v, want the proc chance mod", kinds)
	}
	if got, want := pet.ModifiedProcChance(frenzy.EffectN(1).ClassFlags, 0), frenzy.EffectN(1).Percent(); math.Abs(got-want) > 1e-9 {
		t.Errorf("the pet's Frenzy chance with the talent: %v, want %v", got, want)
	}

	paladin := parseCharacter(&proto.Player{Class: proto.Class_ClassPaladin, Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{}}})
	improvedJudgement := MustFind(23564)
	if kinds := appliedKinds(ParseStatic(paladin, improvedJudgement)); len(kinds) != 1 || kinds[0] != "SpellMod_ProcChance_Flat" {
		t.Fatalf("Improved Judgement of Light attached %v, want the proc chance mod", kinds)
	}
	if got, want := paladin.ModifiedProcChance(MustFind(20185).ClassFlags, 0.5), 0.5+improvedJudgement.EffectN(1).Percent(); math.Abs(got-want) > 1e-9 {
		t.Errorf("Judgement of Light's chance with the talent: %v, want %v", got, want)
	}
}
