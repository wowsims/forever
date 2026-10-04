package spelldata

// The generated store as the package loads it, rather than the fixture the rest of the tests run
// against: the shape every accessor depends on (ids in search order, effects in position order) and
// a handful of rows read out of the game files by hand, so that a regeneration which moves a
// number says so here instead of in a sim result.

import (
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
	if s.SpellLevel != 4 || s.MaxLevel != 8 {
		t.Errorf("Frostbolt rank 1 is taught at %d and stops scaling at %d, want 4 and 8",
			s.SpellLevel, s.MaxLevel)
	}
	if got := s.PowerCost(0); got != 25 {
		t.Errorf("Frostbolt rank 1 costs %v mana, want 25", got)
	}

	damage := s.EffectN(2)
	if damage.BasePoints != 19 || damage.PPL != 0.5 {
		t.Errorf("Frostbolt's damage effect is %v base + %v per level, want 19 + 0.5",
			damage.BasePoints, damage.PPL)
	}
	if damage.SPCoef < 0.4069 || damage.SPCoef > 0.4071 {
		t.Errorf("Frostbolt's spell power coefficient is %v, want 0.407", damage.SPCoef)
	}
	// 19 plus half a point for each of the four levels between the rank's own 4 and its cap of 8.
	if got := damage.AverageAt(60); got != 21 {
		t.Errorf("Frostbolt rank 1 averages %v at level 60, want 21", got)
	}
}

// Per-level scaling counts from the spell's BaseLevel: Vindication's debuff states BaseLevel 0 and
// SpellLevel 1, and the game shows the BaseLevel amount.
func TestAverageCountsLevelsFromBaseLevel(t *testing.T) {
	withGeneratedStore(t)

	vindication := MustFind(440667).EffectN(1)
	for _, c := range []struct {
		level int32
		want  float64
	}{{16, -50}, {20, -64}, {60, -204}} {
		if got := vindication.AverageAt(c.level); got != c.want {
			t.Errorf("Vindication's debuff at %d is %v, want %v", c.level, got, c.want)
		}
	}
	// BaseLevel equals SpellLevel on ordinary ranks, so their amounts do not move.
	if got := MustFind(11556).EffectN(1).AverageAt(60); got != -205 {
		t.Errorf("Demoralizing Shout at 60 is %v, want -205", got)
	}
}

// Every row but 350026 at 0 and 60 is an in-game tooltip. Wrath rank 3 (5178), Wildvine Potion's 11387
// and Minor Recombobulator's 18805 end their base points on halves, which round to the even end: 27.5
// and 32.5 are 28 and 32, 0.5 and 1499.5 are 0 and 1500.
func TestRollEndsMatchTheTooltip(t *testing.T) {
	withGeneratedStore(t)

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

	// Wrath rank 3 at 20 rolls 32 to 37, which averages 34.5 where Average floors 34.2 to 34.
	if got := MustFind(5178).DamageEffect().MeanAt(20); got != 34.5 {
		t.Errorf("Wrath rank 3 at 20 averages %v, want 34.5", got)
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
	if s.ProcCharges != 3 {
		t.Errorf("Lightning Shield holds %d charges, want 3", s.ProcCharges)
	}
	if s.ICDMs != 3500 {
		t.Errorf("Lightning Shield's internal cooldown is %dms, want 3500", s.ICDMs)
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

// Lightning Strike's rate is not in the game files at all: Darkmoon Card: Maelstrom's column reads
// 100 beside a tooltip that says "chance", and an override supplies the PPM.
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

// Frostbolt and Slam state 15, whose 0x2 pushes the cast back; EZ-Thro Bronze Mortar states 5.
func TestGeneratedInterruptFlags(t *testing.T) {
	withGeneratedStore(t)

	for id, want := range map[int32]uint32{116: 15, 11605: 15, 1269268: 5} {
		s := MustFind(id)
		if s.InterruptFlags != want || s.PushedBack() != (want == 15) {
			t.Errorf("%s (%d) states interrupt flags %d, pushed back %v; want %d", s.Name, id,
				s.InterruptFlags, s.PushedBack(), want)
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

// Justice Battleplate's 4-piece bonus 1301076 is a 189 row masked to 8388608.
func TestGeneratedRatingRowParsesExpertise(t *testing.T) {
	withGeneratedStore(t)

	parsed := ParseStatic(parseWarrior(), MustFind(1301076))
	if len(parsed.Applied) != 1 {
		t.Fatalf("attached %v, want one attachment", appliedKinds(parsed))
	}
	if got := parsed.Applied[0].Kind; got != "stat ExpertiseRating" {
		t.Errorf("kind = %q, want %q", got, "stat ExpertiseRating")
	}
	if got := parsed.Applied[0].Value; got != 12 {
		t.Errorf("value = %v, want 12", got)
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
