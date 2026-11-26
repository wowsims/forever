package core

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func newVersusTestUnit() (*Unit, *AttackTable, *AttackTable) {
	unit := &Unit{Type: PlayerUnit, Level: 60, auraTracker: newAuraTracker()}
	undead := &Unit{Type: EnemyUnit, Level: 63, MobType: proto.MobType_MobTypeUndead}
	humanoid := &Unit{Type: EnemyUnit, Level: 63, MobType: proto.MobType_MobTypeHumanoid}
	unit.AttackTables = []*AttackTable{NewAttackTable(unit, undead), NewAttackTable(unit, humanoid)}
	return unit, unit.AttackTables[0], unit.AttackTables[1]
}

func TestAddMobTypeBonusStatsAddsAgainstTheNamedTypes(t *testing.T) {
	unit, undead, humanoid := newVersusTestUnit()
	bonus := stats.Stats{stats.AttackPower: 150}

	unit.AddMobTypeBonusStats(ActionID{SpellID: 1}, bonus, proto.MobType_MobTypeUndead, proto.MobType_MobTypeDemon)
	if got := undead.MobTypeBonusStats[proto.MobType_MobTypeUndead][stats.AttackPower]; got != 150 {
		t.Errorf("Undead attack power %v, want 150", got)
	}
	if got := humanoid.MobTypeBonusStats[proto.MobType_MobTypeHumanoid][stats.AttackPower]; got != 0 {
		t.Errorf("Humanoid attack power %v, want 0", got)
	}

	unit.AddMobTypeBonusStats(ActionID{SpellID: 1}, bonus.Invert(), proto.MobType_MobTypeUndead)
	if got := undead.MobTypeBonusStats[proto.MobType_MobTypeUndead][stats.AttackPower]; got != 0 {
		t.Errorf("Undead attack power %v after the negative bonus, want 0", got)
	}
}

func TestMultiplyDamageVersusMultipliesOnlyTheNamedTypes(t *testing.T) {
	unit, undead, humanoid := newVersusTestUnit()

	unit.MultiplyDamageVersus(ActionID{SpellID: 1}, 1.05, proto.MobType_MobTypeUndead)
	unit.MultiplyCritDamageVersus(ActionID{SpellID: 1}, 1.05, proto.MobType_MobTypeUndead)
	unit.MultiplyDamageVersus(ActionID{SpellID: 1}, 1.02, proto.MobType_MobTypeDemon, proto.MobType_MobTypeUndead)
	want := 1.0
	want *= 1.05
	want *= 1.02
	if undead.DamageDealtMultiplier != want || undead.CritMultiplier != 1.05 {
		t.Errorf("Undead damage %v and crit %v, want %v and 1.05", undead.DamageDealtMultiplier, undead.CritMultiplier, want)
	}
	if humanoid.DamageDealtMultiplier != 1 || humanoid.CritMultiplier != 1 {
		t.Errorf("Humanoid damage %v and crit %v, want 1 and 1", humanoid.DamageDealtMultiplier, humanoid.CritMultiplier)
	}
}

func TestAddArmorIgnoreFactorAddsToEveryTable(t *testing.T) {
	unit, undead, humanoid := newVersusTestUnit()

	unit.AddArmorIgnoreFactor(0.15)
	unit.AddArmorIgnoreFactor(-0.05)
	for _, at := range []*AttackTable{undead, humanoid} {
		if !WithinToleranceFloat64(0.10, at.ArmorIgnoreFactor, 1e-12) {
			t.Errorf("armor ignored %v, want 0.10", at.ArmorIgnoreFactor)
		}
	}
}

func TestAttachMobTypeBonusStatsFollowsTheAura(t *testing.T) {
	unit, undead, _ := newVersusTestUnit()
	aura := unit.RegisterAura(Aura{Label: "Versus", Duration: NeverExpires}).
		AttachMobTypeBonusStats(ActionID{SpellID: 1}, stats.Stats{stats.SpellDamage: 58}, proto.MobType_MobTypeUndead)

	aura.OnGain(aura, nil)
	if got := undead.MobTypeBonusStats[proto.MobType_MobTypeUndead][stats.SpellDamage]; got != 58 {
		t.Errorf("spell damage against Undead %v while up, want 58", got)
	}
	aura.OnExpire(aura, nil)
	if got := undead.MobTypeBonusStats[proto.MobType_MobTypeUndead][stats.SpellDamage]; got != 0 {
		t.Errorf("spell damage against Undead %v after expiry, want 0", got)
	}
}

type versusLogTest struct {
	sim    *Simulation
	unit   *Character
	target *Unit
	strike *Spell
	logs   []string
}

func newVersusLogTest() *versusLogTest {
	sim := SetupFakeRageSim()
	return versusLogTestOn(sim, &sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Character)
}

func versusLogTestOn(sim *Simulation, unit *Character) *versusLogTest {
	vt := &versusLogTest{
		sim:    sim,
		unit:   unit,
		target: sim.Encounter.ActiveTargetUnits[0],
	}
	vt.strike = vt.unit.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 100},
		SpellSchool:      SpellSchoolPhysical,
		ProcMask:         ProcMaskMeleeMHSpecial,
		DefenseType:      DefenseTypeMelee,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	})
	sim.Log = func(message string, vals ...any) { vt.logs = append(vt.logs, fmt.Sprintf(message, vals...)) }
	return vt
}

func (vt *versusLogTest) deal(spell *Spell, outcome HitOutcome) {
	result := spell.CalcDamage(vt.sim, vt.target, 100, func(_ *Simulation, result *SpellResult, _ *AttackTable) { result.Outcome = outcome })
	spell.DealDamage(vt.sim, result)
}

func (vt *versusLogTest) tick(spell *Spell, outcome HitOutcome) {
	result := spell.CalcPeriodicDamage(vt.sim, vt.target, 100, func(_ *Simulation, result *SpellResult, _ *AttackTable) { result.Outcome = outcome })
	spell.DealPeriodicDamage(vt.sim, result)
}

func (vt *versusLogTest) damageLine(t *testing.T, spell *Spell, outcome HitOutcome) string {
	t.Helper()
	return vt.lineOf(t, func() { vt.deal(spell, outcome) })
}

func (vt *versusLogTest) lineOf(t *testing.T, hit func()) string {
	t.Helper()
	vt.logs = nil
	hit()
	for _, line := range vt.logs {
		if strings.Contains(line, "(Threat: ") {
			return line
		}
	}
	t.Fatalf("no damage line among %q", vt.logs)
	return ""
}

func TestDamageLogNamesTheVersusSourcesTheHitTook(t *testing.T) {
	vt := newVersusLogTest()
	humanoid, undead := proto.MobType_MobTypeHumanoid, proto.MobType_MobTypeUndead
	if vt.target.MobType != humanoid {
		t.Fatalf("the target is %v, want Humanoid", vt.target.MobType)
	}
	vt.unit.MultiplyDamageVersus(ActionID{ItemID: 11}, 1.05, humanoid)
	vt.unit.MultiplyCritDamageVersus(ActionID{SpellID: 33}, 1.1, humanoid)
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81, stats.SpellDamage: 40}, humanoid, undead)
	vt.unit.MultiplyDamageVersus(ActionID{ItemID: 44}, 2, undead)

	for _, c := range []struct {
		outcome HitOutcome
		want    string
	}{
		{OutcomeHit, ") [Versus Humanoid: {ItemID: 11} x1.05; {SpellID: 22} +81 AttackPower]"},
		{OutcomeCrit, ") [Versus Humanoid: {ItemID: 11} x1.05; {SpellID: 33} x1.1 crit; {SpellID: 22} +81 AttackPower]"},
	} {
		if line := vt.damageLine(t, vt.strike, c.outcome); !strings.HasSuffix(line, c.want) {
			t.Errorf("%v line %q, want it to end with %q", c.outcome, line, c.want)
		}
	}
	if line := vt.damageLine(t, vt.strike, OutcomeMiss); strings.Contains(line, "[Versus") {
		t.Errorf("a miss is tagged: %q", line)
	}
}

func TestDamageLogTagsNoSourceOfAnotherMobType(t *testing.T) {
	vt := newVersusLogTest()
	undead := proto.MobType_MobTypeUndead
	vt.unit.MultiplyDamageVersus(ActionID{ItemID: 11}, 1.05, undead)
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81}, undead)

	if line := vt.damageLine(t, vt.strike, OutcomeHit); !strings.HasSuffix(line, ")") || strings.Contains(line, "[Versus") {
		t.Errorf("a hit on a Humanoid is tagged with bonuses against Undead: %q", line)
	}
}

func TestDamageLogDropsASourceTakenBackOff(t *testing.T) {
	var aura *Aura
	sim, character := NewWeaponTestSim(t, proto.Race_RaceHuman, WeaponTestSlots(WeaponTestSwordID, 0, 0), nil, func(character *Character) {
		aura = character.RegisterAura(Aura{Label: "Test Humanoid Bane", Duration: NeverExpires}).
			AttachMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81}, proto.MobType_MobTypeHumanoid)
	})
	vt := versusLogTestOn(sim, character)
	vt.unit.MultiplyDamageVersus(ActionID{ItemID: 11}, 1.05, vt.target.MobType)
	vt.unit.MultiplyDamageVersus(ActionID{ItemID: 11}, 1/1.05, vt.target.MobType)

	aura.Activate(sim)
	if line, want := vt.damageLine(t, vt.strike, OutcomeHit), ") [Versus Humanoid: {SpellID: 22} +81 AttackPower]"; !strings.HasSuffix(line, want) {
		t.Errorf("line %q while the aura is up, want it to end with %q", line, want)
	}
	aura.Deactivate(sim)
	if line := vt.damageLine(t, vt.strike, OutcomeHit); strings.Contains(line, "[Versus") {
		t.Errorf("line %q after the aura fell off, want no tag", line)
	}
}

func TestDamageLogTagsATickByTheDotsBonusCoefficient(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	vt := versusLogTestOn(sim, &fa.Character)
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81, stats.SpellDamage: 40}, vt.target.MobType)

	line := vt.lineOf(t, func() { vt.tick(fa.Spell, OutcomeHit) })
	if want := ") [Versus Demon: {SpellID: 22} +40 SpellDamage]"; !strings.Contains(line, " tick ") || !strings.HasSuffix(line, want) {
		t.Errorf("tick line %q, want a tick ending with %q", line, want)
	}
}

func TestDamageLogNamesTheSpellsOwnVersusBonusAndTheFamily(t *testing.T) {
	vt := newVersusLogTest()
	vt.target.CreatureFamily = proto.CreatureFamily_CreatureFamilyOrc
	bolt := vt.unit.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 101},
		SpellSchool:      SpellSchoolFire,
		ProcMask:         ProcMaskSpellDamage,
		DefenseType:      DefenseTypeMagic,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,
		VersusBonus:      &VersusBonus{Multiplier: 3, Families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyOrc}},
	})
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81, stats.SpellDamage: 40}, vt.target.MobType)

	if line, want := vt.damageLine(t, bolt, OutcomeHit), ") [Versus Humanoid (Orc): {SpellID: 101} x3 base; {SpellID: 22} +40 SpellDamage]"; !strings.HasSuffix(line, want) {
		t.Errorf("line %q, want it to end with %q", line, want)
	}
}

func TestDamageLogListsTheStatsASpellDeclaresAlone(t *testing.T) {
	vt := newVersusLogTest()
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81, stats.SpellDamage: 40, stats.PhysicalDamage: 6}, vt.target.MobType)
	holyStrike := func(id int32, read []stats.Stat) *Spell {
		return vt.unit.RegisterSpell(SpellConfig{
			ActionID:         ActionID{SpellID: id},
			SpellSchool:      SpellSchoolHoly,
			ProcMask:         ProcMaskMeleeMHSpecial,
			DefenseType:      DefenseTypeMelee,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			VersusStats:      read,
		})
	}

	for _, c := range []struct {
		name  string
		spell *Spell
		want  string
	}{
		{"an undeclared melee hit", holyStrike(101, nil), ") [Versus Humanoid: {SpellID: 22} +81 AttackPower]"},
		{"a hit declaring spell damage alone", holyStrike(102, []stats.Stat{stats.SpellDamage}), ") [Versus Humanoid: {SpellID: 22} +40 SpellDamage]"},
		{"a hit declaring none", holyStrike(103, NoVersusStats), ")"},
	} {
		if line := vt.damageLine(t, c.spell, OutcomeHit); !strings.HasSuffix(line, c.want) || (c.want == ")" && strings.Contains(line, "[Versus")) {
			t.Errorf("%s: line %q, want it to end with %q", c.name, line, c.want)
		}
	}
}

func TestDamageLogTagsATickByTheDotsDeclaration(t *testing.T) {
	const hitTag = ") [Versus Humanoid: {SpellID: 22} +70 RangedAttackPower]"
	cases := []struct {
		name    string
		isAOE   bool
		dotRead []stats.Stat
		tick    string
		spell   *Spell
	}{
		{name: "a dot declaring spell damage", dotRead: []stats.Stat{stats.SpellDamage}, tick: ") [Versus Humanoid: {SpellID: 22} +40 SpellDamage]"},
		{name: "a dot declaring none", dotRead: NoVersusStats, tick: ")"},
		{name: "a dot declaring nothing", tick: hitTag},
		{name: "an area dot declaring none", isAOE: true, dotRead: NoVersusStats, tick: ")"},
	}
	sim, character := NewWeaponTestSim(t, proto.Race_RaceHuman, WeaponTestSlots(WeaponTestSwordID, 0, 0), nil, func(character *Character) {
		for i := range cases {
			id := int32(101 + i)
			cases[i].spell = character.RegisterSpell(SpellConfig{
				ActionID:         ActionID{SpellID: id},
				SpellSchool:      SpellSchoolFire,
				ProcMask:         ProcMaskSpellDamage,
				DefenseType:      DefenseTypeMagic,
				DamageMultiplier: 1,
				ThreatMultiplier: 1,
				VersusStats:      []stats.Stat{stats.RangedAttackPower},
				Dot: DotConfig{
					Aura:          Aura{Label: fmt.Sprintf("Test Trap %d", id)},
					IsAOE:         cases[i].isAOE,
					NumberOfTicks: 1,
					TickLength:    time.Second,
					VersusStats:   cases[i].dotRead,
				},
			})
		}
	})
	vt := versusLogTestOn(sim, character)
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.RangedAttackPower: 70, stats.SpellDamage: 40}, vt.target.MobType)

	for _, c := range cases {
		if line := vt.damageLine(t, c.spell, OutcomeHit); !strings.HasSuffix(line, hitTag) {
			t.Errorf("%s: hit line %q, want it to end with %q", c.name, line, hitTag)
		}
		line := vt.lineOf(t, func() { vt.tick(c.spell, OutcomeHit) })
		if !strings.Contains(line, " tick ") || !strings.HasSuffix(line, c.tick) || (c.tick == ")" && strings.Contains(line, "[Versus")) {
			t.Errorf("%s: tick line %q, want a tick ending with %q", c.name, line, c.tick)
		}
	}
}

func TestDamageLogNamesAnItemsStatsByTheItem(t *testing.T) {
	vt := newVersusLogTest()
	item := &Item{ID: 12345, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {MobTypeStats: []*proto.MobTypeStats{
		{MobTypes: []proto.MobType{vt.target.MobType}, Stats: map[int32]float64{int32(stats.AttackPower): 81}},
	}}}}
	vt.unit.addItemMobTypeStats(item, 1)

	if line, want := vt.damageLine(t, vt.strike, OutcomeHit), ") [Versus Humanoid: {ItemID: 12345} +81 AttackPower]"; !strings.HasSuffix(line, want) {
		t.Errorf("line %q, want it to end with %q", line, want)
	}
}

func TestDamageLogNamesACritChanceTheTargetsFamilyRaises(t *testing.T) {
	orcs := &VersusBonus{Multiplier: 2, Families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyOrc}}
	for _, c := range []struct {
		family   proto.CreatureFamily
		wantCrit float64
		want     string
	}{
		{proto.CreatureFamily_CreatureFamilyOrc, 10, ") [Versus Humanoid (Orc): {ItemID: 13965} +10% crit chance]"},
		{proto.CreatureFamily_CreatureFamilyUnknown, 5, ")"},
	} {
		vt := newVersusLogTest()
		vt.target.CreatureFamily = c.family
		table := vt.unit.AttackTables[vt.target.UnitIndex]
		vt.unit.AddCritChanceVersus(ActionID{ItemID: 13965}, vt.target, 5, orcs)

		if table.BonusCritPercent != c.wantCrit {
			t.Errorf("against %v the crit chance is %v, want %v", c.family, table.BonusCritPercent, c.wantCrit)
		}
		if line := vt.damageLine(t, vt.strike, OutcomeHit); !strings.HasSuffix(line, c.want) || (c.want == ")" && strings.Contains(line, "[Versus")) {
			t.Errorf("against %v line %q, want it to end with %q", c.family, line, c.want)
		}
		vt.unit.AddCritChanceVersus(ActionID{ItemID: 13965}, vt.target, -5, orcs)
		if line := vt.damageLine(t, vt.strike, OutcomeHit); table.BonusCritPercent != 0 || strings.Contains(line, "[Versus") {
			t.Errorf("against %v the crit chance is %v once taken back off and the line %q, want none and no tag", c.family, table.BonusCritPercent, line)
		}
	}
}

func TestDamageLogNamesTheProcThatCastTheSpell(t *testing.T) {
	dwarves := []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyDwarf}
	for _, c := range []struct {
		name   string
		mult   float64
		family proto.CreatureFamily
		want   string
	}{
		{"a likelier proc on a Dwarf", 3, proto.CreatureFamily_CreatureFamilyDwarf, ") [Versus Humanoid (Dwarf): {ItemID: 11815} x3 proc chance]"},
		{"a proc on Dwarves alone on a Dwarf", 0, proto.CreatureFamily_CreatureFamilyDwarf, ") [Versus Humanoid (Dwarf): {ItemID: 11815} proc only]"},
		{"a likelier proc on another target", 3, proto.CreatureFamily_CreatureFamilyUnknown, ")"},
	} {
		vt := newVersusLogTest()
		vt.target.CreatureFamily = c.family
		vt.strike.VersusProc = &VersusProc{Source: ActionID{ItemID: 11815}, VersusBonus: VersusBonus{Multiplier: c.mult, Families: dwarves}}

		if line := vt.damageLine(t, vt.strike, OutcomeHit); !strings.HasSuffix(line, c.want) || (c.want == ")" && strings.Contains(line, "[Versus")) {
			t.Errorf("%s: line %q, want it to end with %q", c.name, line, c.want)
		}
	}
}

func TestDamageWithoutALogBuildsNoTag(t *testing.T) {
	vt := newVersusLogTest()
	vt.sim.Log = nil
	hit := func() { vt.deal(vt.strike, OutcomeHit) }
	without := testing.AllocsPerRun(100, hit)
	vt.target.CreatureFamily = proto.CreatureFamily_CreatureFamilyOrc
	orcs := VersusBonus{Multiplier: 2, Families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyOrc}}
	vt.unit.MultiplyDamageVersus(ActionID{ItemID: 11}, 1.05, vt.target.MobType)
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81}, vt.target.MobType)
	vt.unit.AddCritChanceVersus(ActionID{ItemID: 13965}, vt.target, 5, &orcs)
	vt.strike.VersusProc = &VersusProc{Source: ActionID{ItemID: 11815}, VersusBonus: orcs}

	if with := testing.AllocsPerRun(100, hit); with != without {
		t.Errorf("a hit with no log allocates %v times under versus bonuses, %v without", with, without)
	}
}
