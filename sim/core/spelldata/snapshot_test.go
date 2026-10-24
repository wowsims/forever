package spelldata

// The generated store as the package loads it, rather than the fixture the rest of the tests run
// against: the shape every accessor depends on (ids in search order, effects in position order) and
// the links, masks and effect kinds of a handful of rows; no amount is pinned.

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// Installs the generated data for one test and puts the fixture back afterwards, since TestMain
// swaps the fixture in for the whole package.
func withGeneratedStore(t *testing.T) {
	t.Helper()
	t.Cleanup(generatedStore())
}

func TestGeneratedStoreShape(t *testing.T) {
	withGeneratedStore(t)

	// The bounds are wide enough that a regeneration adding a class or a patch's spells does not fail
	// the gate. Whether the committed store is the one the committed inputs produce is
	// TestStoreRegeneratesFromTheCommittedInputs' job.
	rows := All()
	if len(rows) < 5000 || len(rows) > 12000 {
		t.Fatalf("the store holds %d spells, which is outside the [5000, 12000] a whole store has",
			len(rows))
	}

	effects, gapped := 0, 0
	for i, s := range rows {
		if i > 0 && s.ID <= rows[i-1].ID {
			t.Fatalf("spell %d follows %d: Find binary searches, so the ids have to ascend",
				s.ID, rows[i-1].ID)
		}

		effects += len(s.Effects)
		packed := true
		for position := range s.Effects {
			e := &s.Effects[position]
			if e.SpellID != s.ID {
				t.Errorf("spell %d effect %d says it belongs to spell %d", s.ID, position, e.SpellID)
			}
			if position > 0 && e.Index <= s.Effects[position-1].Index {
				t.Errorf("spell %d has effect index %d after %d - the effects are out of game files order",
					s.ID, e.Index, s.Effects[position-1].Index)
			}
			if int(e.Index) != position {
				packed = false
			}
		}
		if !packed {
			gapped++
		}
	}

	if effects < 6000 || effects > 16000 {
		t.Errorf("the store holds %d effects, which is outside the [6000, 16000] a whole store has",
			effects)
	}
	// The game files' EffectIndex has gaps, which is why EffectN counts by position and the store
	// keeps the game files' own number alongside.
	if gapped == 0 {
		t.Errorf("no row states an EffectIndex that is not its position - EffectN's by-position count " +
			"no longer has a gap to step over")
	}
}

// Frostbolt rank 1: the row the package comment's units are read off.
func TestGeneratedFrostbolt(t *testing.T) {
	withGeneratedStore(t)

	s := MustFind(116)
	if s.Name != "Frostbolt" || s.Rank != "Rank 1" {
		t.Errorf("spell 116 is %q %q, want Frostbolt Rank 1", s.Name, s.Rank)
	}
	if s.School != 16 {
		t.Errorf("Frostbolt's school mask is %d, want 16 (frost)", s.School)
	}
	if got := s.DamageEffect(); got != s.EffectN(2) {
		t.Errorf("Frostbolt's damage is effect %d, want the second by position", got.Index)
	}
}

// Per-level scaling counts from the spell's BaseLevel, which can sit below the level it is taught at:
// the copy of Vindication's debuff 440667 here is an effect at BaseLevel 0 on a spell taught at 1,
// and the game shows the BaseLevel amount.
func TestAverageCountsLevelsFromBaseLevel(t *testing.T) {
	vindication := &Effect{SpellID: 440667, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_ATTACK_POWER,
		BasePoints: 6, PPL: -3.5, MaxLevel: 60}
	for _, c := range []struct {
		level int32
		want  float64
	}{{16, -50}, {20, -64}, {60, -204}} {
		if got := vindication.AverageAt(c.level); got != c.want {
			t.Errorf("Vindication's debuff at %d is %v, want %v", c.level, got, c.want)
		}
	}
	// BaseLevel equals SpellLevel on ordinary ranks, so their amounts do not move: Demoralizing Shout
	// 11556's.
	demoralizingShout := &Effect{SpellID: 11556, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_ATTACK_POWER,
		BasePoints: -196, PPL: -1.399999976158142, BaseLevel: 54, MaxLevel: 64}
	if got := demoralizingShout.AverageAt(60); got != -205 {
		t.Errorf("Demoralizing Shout at 60 is %v, want -205", got)
	}
}

// Copies of game files rows whose ends, every one but 350026's at 0 and 60, are an in-game tooltip.
// Wrath rank 3 (5178), Wildvine Potion's 11387 and Minor Recombobulator's 18805 end their base points
// on halves, which round to the even end: 27.5 and 32.5 are 28 and 32, 0.5 and 1499.5 are 0 and 1500.
func TestRollEndsMatchTheTooltip(t *testing.T) {
	withRows(t, []Spell{
		{ID: 2912, Name: "Starfire", School: 64, Effects: []Effect{
			{SpellID: 2912, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 81, PPL: 1.2999999523162842,
				Variance: 0.20202019810676575, BaseLevel: 20, MaxLevel: 25},
		}},
		{ID: 5177, Name: "Wrath", School: 8, Effects: []Effect{
			{SpellID: 5177, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 23, PPL: 0.6000000238418579,
				Variance: 0.14814814925193787, BaseLevel: 6, MaxLevel: 12},
		}},
		{ID: 5178, Name: "Wrath", School: 8, Effects: []Effect{
			{SpellID: 5178, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 30, PPL: 0.699999988079071,
				Variance: 0.1666666716337204, BaseLevel: 14, MaxLevel: 20},
		}},
		{ID: 11387, Name: "Wildvine Potion", School: 1, Effects: []Effect{
			{SpellID: 11387, Type: dbcenums.E_HEAL, BasePoints: 750, Variance: 1.9986666440963745},
			{SpellID: 11387, Index: 1, Type: dbcenums.E_ENERGIZE, BasePoints: 750, Variance: 1.9986666440963745},
		}},
		{ID: 18805, Name: "Recombobulate", School: 64, Effects: []Effect{
			{SpellID: 18805, Type: dbcenums.E_DISPEL_MECHANIC, BasePoints: 1, MaxLevel: 30},
			{SpellID: 18805, Index: 1, Type: dbcenums.E_HEAL, BasePoints: 270, Variance: 0.5, MaxLevel: 30},
			{SpellID: 18805, Index: 2, Type: dbcenums.E_ENERGIZE, BasePoints: 128, Variance: 0.5, MaxLevel: 30},
		}},
		{ID: 350025, Name: "Frostbolt", School: 16, Effects: []Effect{
			{SpellID: 350025, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_DECREASE_SPEED, BasePoints: -40, MaxLevel: 12},
			{SpellID: 350025, Index: 1, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 33, PPL: 0.699999988079071,
				Variance: 0.12121211737394333, MaxLevel: 12},
		}},
		{ID: 350026, Name: "Shadow Bolt", School: 32, Effects: []Effect{
			{SpellID: 350026, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 52, PPL: 0.8999999761581421,
				Variance: 0.1538461595773697, MaxLevel: 17},
		}},
	})

	for _, c := range []struct {
		spell    int32
		effect   int
		level    int32
		min, max float64
	}{
		{350026, 1, 0, 48, 56},
		{350026, 1, 14, 60, 69},
		{350026, 1, 60, 63, 72},
		{350025, 2, 14, 39, 44},
		{5177, 1, 20, 24, 29},
		{2912, 1, 20, 73, 89},
		{5178, 1, 20, 32, 37},
		{11387, 1, 35, 0, 1500},
		{18805, 2, 60, 202, 338},
		{18805, 3, 60, 96, 160},
	} {
		effect := MustFind(c.spell).EffectN(c.effect)
		if got, want := [2]float64{effect.MinAt(c.level), effect.MaxAt(c.level)}, [2]float64{c.min, c.max}; got != want {
			t.Errorf("%d effect %d at %d rolls %v, want %v", c.spell, c.effect, c.level, got, want)
		}
	}

	// Wrath rank 3 at 20 shows 32 to 37, but a cast rolls its whole amount around Average, which keeps
	// the fraction, so before spell power it averages 34.2 rather than the tooltip's middle.
	if got, want := MustFind(5178).DamageEffect().MeanAt(20), float64(float32(34.2)); got != want {
		t.Errorf("Wrath rank 3 at 20 averages %v, want %v", got, want)
	}
}

// A physical hit rolls its base points by the whole Variance either side: the copy of Revenge rank 1
// below, 22 ± 3.4, rolls 19 to 25. A spell-school hit and a heal of any school, Healing Potion's
// included, leave their roll to core, which spreads the whole amount by DirectVariance; an energize
// keeps the half-width roll.
func TestRollWidthByEffectKind(t *testing.T) {
	withRows(t, []Spell{{ID: 6572, Name: "Revenge", School: 1, Effects: []Effect{
		{SpellID: 6572, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 22, Variance: 0.1538461595773697, BaseLevel: 14},
	}}})
	copied := MustFind(6572).DamageEffect()
	if got, want := [2]float64{copied.MinAt(60), copied.MaxAt(60)}, [2]float64{19, 25}; got != want {
		t.Errorf("Revenge rank 1 rolls %v at 60, want %v", got, want)
	}

	withGeneratedStore(t)
	revenge := MustFind(6572).DamageEffect()
	if got := revenge.DirectVariance(); got != 0 {
		t.Errorf("Revenge rank 1 hands core a DirectVariance of %v, want 0", got)
	}

	for _, effect := range []*Effect{MustFind(5178).DamageEffect(), MustFind(439).HealEffect()} {
		if got := effect.DirectVariance(); got != effect.Variance || got == 0 {
			t.Errorf("%d hands core a DirectVariance of %v, want its Variance %v", effect.SpellID, got, effect.Variance)
		}
	}

	if got := MustFind(437).EnergizeEffect().DirectVariance(); got != 0 {
		t.Errorf("Restore Mana hands core a DirectVariance of %v, want 0", got)
	}
}

// Lightning Shield rank 1: the proc columns, and the two edges a proc row is read through. The aura
// triggers the shared dispatcher 26545, which every rank triggers; the rank's own damage spell is
// named by the tooltip instead, which is what RefIDs carries.
func TestGeneratedLightningShield(t *testing.T) {
	withGeneratedStore(t)

	s := MustFind(324)
	if s.ProcFlags[0] != 0x222A8 || s.ProcFlags[1] != 0 {
		t.Errorf("Lightning Shield's proc mask is %#x/%#x, want 0x222a8/0", s.ProcFlags[0], s.ProcFlags[1])
	}
	if s.ProcChanceSource != ProcChanceAlways {
		t.Errorf("Lightning Shield's proc chance source is %d, want ProcChanceAlways", s.ProcChanceSource)
	}
	if got := s.EffectN(1).TriggerID; got != 26545 {
		t.Errorf("Lightning Shield's aura triggers %d, want the shared dispatcher 26545", got)
	}
	if len(s.RefIDs) != 1 || s.RefIDs[0] != 26364 {
		t.Errorf("Lightning Shield's tooltip names %v, want [26364]", s.RefIDs)
	}

	if !drives(MustFind(26545), 324) {
		t.Errorf("spell 324 does not drive 26545, so a rank's dispatcher cannot be reached from it")
	}
}

// Stolen Power's companion row requires Forest and Grassland; Staff of Westfall's group is a single zone.
func TestGeneratedRequiredAreas(t *testing.T) {
	withGeneratedStore(t)

	if s := MustFind(1318002); s.RequiredAreas != 9161 || s.AreaType() != proto.AreaType_AreaTypeForestGrassland {
		t.Errorf("Stolen Power 1318002 requires area group %d (%v), want 9161 (ForestGrassland)", s.RequiredAreas, s.AreaType())
	}
	if s := MustFind(1287561); s.RequiredAreas != 0 || s.AreaType() != proto.AreaType_AreaTypeUnknown {
		t.Errorf("Stolen Power 1287561 requires area group %d, want none", s.RequiredAreas)
	}
	if s := MustFind(1292011); s.RequiredAreas != 9071 || s.AreaType() != proto.AreaType_AreaTypeUnknown {
		t.Errorf("Staff of Westfall requires area group %d (%v), want 9071 (Unknown)", s.RequiredAreas, s.AreaType())
	}
}

// Area bonuses come from the override table, not the game files.
func TestGeneratedAreaBonus(t *testing.T) {
	withGeneratedStore(t)

	if s := MustFind(1249113); len(s.AreaBonusGroups) != 1 || s.AreaBonusGroups[0] != 9203 || s.AreaMultiplier != 2 || s.AreaDurationMultiplier != 1 {
		t.Errorf("Molten Fury's area bonus reads %v x%v (duration x%v), want [9203] x2 (duration x1)", s.AreaBonusGroups, s.AreaMultiplier, s.AreaDurationMultiplier)
	}
	if s := MustFind(1287571); s.AreaMultiplier != 2 || s.AreaDurationMultiplier != 2 {
		t.Errorf("Monkey Business's area bonus reads x%v (duration x%v), want x2 (duration x2)", s.AreaMultiplier, s.AreaDurationMultiplier)
	}
}

// Arcane Missiles rank 2: a periodic trigger reaches its tick spell through the effect.
func TestGeneratedTriggerResolves(t *testing.T) {
	withGeneratedStore(t)

	if got := MustFind(5144).EffectN(1).Trigger().ID; got != 7269 {
		t.Errorf("Arcane Missiles rank 2 ticks spell %d, want 7269", got)
	}
}

// Retaliation's counterattack is a link the game files do not state and overrides.HandTriggers
// supplies, so it has to come back out of Triggered() and Drivers().
func TestGeneratedHandLink(t *testing.T) {
	withGeneratedStore(t)

	var ids []int32
	for _, s := range MustFind(20230).Triggered() {
		ids = append(ids, s.ID)
	}
	if !slices.Contains(ids, 20240) {
		t.Errorf("Retaliation triggers %v, want the counterattack 20240 among them", ids)
	}

	ids = nil
	for _, s := range MustFind(20240).Drivers() {
		ids = append(ids, s.ID)
	}
	if !slices.Contains(ids, 20230) {
		t.Errorf("the counterattack is driven by %v, want Retaliation 20230 among them", ids)
	}
}

// Flurry is a trait talent: one spell, five ranks, and the per-rank numbers on a curve.
func TestGeneratedTalentCurve(t *testing.T) {
	withGeneratedStore(t)

	curve := curves[12319]
	if len(curve) != 1 || len(curve[0]) != 5 {
		t.Fatalf("Flurry's curve is %v, want one effect row of five ranks", curve)
	}

	ladder := Talent(12319, 5)
	if got := ladder.Rank(3).EffectN(1).BaseValue(); got != curve[0][2] {
		t.Errorf("Flurry rank 3 states %v, want the curve's third value %v", got, curve[0][2])
	}
	if got := ladder.Rank(0); got != Nil {
		t.Errorf("an untaken talent reads %v, want Nil", got)
	}
}

// Lightning Strike's rate is not in the game files at all: Darkmoon Card: Maelstrom's column states
// no roll beside a tooltip that says "chance", and an override supplies the PPM.
func TestGeneratedOverrideRow(t *testing.T) {
	withGeneratedStore(t)

	s := MustFind(23686)
	if s.RPPM != 1 {
		t.Errorf("Lightning Strike's RPPM is %v, want the override's 1", s.RPPM)
	}
	if s.ProcChanceSource != ProcChancePPM {
		t.Errorf("Lightning Strike's proc chance source is %d, want ProcChancePPM", s.ProcChanceSource)
	}
}

// Frostbolt and Slam state the flag that pushes the cast back, and Slam's row also carries the
// attribute that keeps it from being pushed back. EZ-Thro Bronze Mortar states neither.
func TestGeneratedInterruptFlags(t *testing.T) {
	withGeneratedStore(t)

	for id, want := range map[int32]struct {
		flags      uint32
		pushedBack bool
	}{116: {15, true}, 11605: {15, false}, 1269268: {5, false}} {
		s := MustFind(id)
		if s.InterruptFlags != want.flags || s.PushedBack() != want.pushedBack {
			t.Errorf("%s (%d) states interrupt flags %d, pushed back %v; want %d, pushed back %v", s.Name, id,
				s.InterruptFlags, s.PushedBack(), want.flags, want.pushedBack)
		}
	}
}

// Wrath is castable in Moonkin Form and excludes Tree Form, Healing Touch the reverse. Shifting Power
// states its Cat Form requirement twice, as a caster aura beside its stance mask.
func TestGeneratedStanceAndAuraRestriction(t *testing.T) {
	withGeneratedStore(t)

	wrath := MustFind(5176)
	if wrath.StanceMask != 0x40000000 || wrath.StanceExclude != 0x2 {
		t.Errorf("Wrath's stance mask is %#x exclude %#x, want 0x40000000 exclude 0x2",
			wrath.StanceMask, wrath.StanceExclude)
	}

	healingTouch := MustFind(5185)
	if healingTouch.StanceMask != 0x2 || healingTouch.StanceExclude != 0x40000000 {
		t.Errorf("Healing Touch's stance mask is %#x exclude %#x, want 0x2 exclude 0x40000000",
			healingTouch.StanceMask, healingTouch.StanceExclude)
	}

	shiftingPower := MustFind(1322605)
	if shiftingPower.CasterAura != 768 || shiftingPower.StanceMask != 0x1 {
		t.Errorf("Shifting Power's caster aura is %d and stance mask %#x, want 768 and 0x1",
			shiftingPower.CasterAura, shiftingPower.StanceMask)
	}

	for id, want := range map[int32]dbcenums.ShapeshiftForm{
		2457: dbcenums.FORM_BATTLE_STANCE, 71: dbcenums.FORM_DEFENSIVE_STANCE, 2458: dbcenums.FORM_BERSERKER_STANCE,
		768: dbcenums.FORM_CAT_FORM, 9634: dbcenums.FORM_DIRE_BEAR_FORM, 24858: dbcenums.FORM_MOONKIN_FORM,
	} {
		if got := MustFind(id).ShapeshiftForm(); got != want {
			t.Errorf("spell %d puts the caster in form %d, want %d", id, got, want)
		}
	}
}

// Battle Stance's and Shadowform's configs state the form they shift to, with no menu group.
func TestConfigsCarryTheFormData(t *testing.T) {
	withGeneratedStore(t)

	battleStance := SpellConfig(&core.Unit{}, MustFind(2457))
	if battleStance.ShiftsTo != dbcenums.FORM_BATTLE_STANCE || battleStance.MenuGroup != "" {
		t.Errorf("Battle Stance ShiftsTo = %d, MenuGroup = %q; want %d and none", battleStance.ShiftsTo, battleStance.MenuGroup, dbcenums.FORM_BATTLE_STANCE)
	}
	shadowform := SpellConfig(&core.Unit{}, MustFind(15473))
	if shadowform.ShiftsTo != dbcenums.FORM_SHADOWFORM || shadowform.MenuGroup != "" {
		t.Errorf("Shadowform ShiftsTo = %d, MenuGroup = %q; want %d and none", shadowform.ShiftsTo, shadowform.MenuGroup, dbcenums.FORM_SHADOWFORM)
	}
}

// Justice Battleplate's 4-piece bonus 1301076 is an A_MOD_RATING row masked to the expertise bit.
func TestGeneratedRatingRowParsesExpertise(t *testing.T) {
	withGeneratedStore(t)

	row := MustFind(1301076)
	parsed := ParseStatic(parseWarrior(), row)
	if len(parsed.Applied) != 1 {
		t.Fatalf("attached %v, want one attachment", appliedKinds(parsed))
	}
	if got := parsed.Applied[0].Kind; got != "stat ExpertiseRating" {
		t.Errorf("kind = %q, want %q", got, "stat ExpertiseRating")
	}
	if got, want := parsed.Applied[0].Value, row.EffectN(1).Mean(); got != want {
		t.Errorf("value = %v, want the row's %v", got, want)
	}
}

func TestGeneratedStoreMisses(t *testing.T) {
	withGeneratedStore(t)

	if Find(0) != Nil {
		t.Errorf("Find(0) answers a row, want Nil")
	}
	if Find(-1) != Nil {
		t.Errorf("Find(-1) answers a row, want Nil")
	}

	defer func() {
		message, ok := recover().(string)
		if !ok {
			t.Fatalf("MustFind(0) did not panic on an id the store does not carry")
		}
		if !strings.Contains(message, "spelldata: spell 0 is not in the store") {
			t.Errorf("MustFind(0) panicked with %q, want the store's own message", message)
		}
	}()
	MustFind(0)
}

func drives(s *Spell, id int32) bool {
	for _, driver := range s.Drivers() {
		if driver.ID == id {
			return true
		}
	}
	return false
}
