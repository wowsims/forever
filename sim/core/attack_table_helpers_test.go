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
	if got := undead.VersusStats[stats.AttackPower]; got != 150 {
		t.Errorf("Undead attack power %v, want 150", got)
	}
	if got := humanoid.VersusStats[stats.AttackPower]; got != 0 {
		t.Errorf("Humanoid attack power %v, want 0", got)
	}

	unit.AddMobTypeBonusStats(ActionID{SpellID: 1}, bonus.Invert(), proto.MobType_MobTypeUndead)
	if got := undead.VersusStats[stats.AttackPower]; got != 0 {
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
	if got := undead.VersusStats[stats.SpellDamage]; got != 58 {
		t.Errorf("spell damage against Undead %v while up, want 58", got)
	}
	aura.OnExpire(aura, nil)
	if got := undead.VersusStats[stats.SpellDamage]; got != 0 {
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
	vt.strike = versusTestSpell(vt.unit, 100, SpellSchoolPhysical, nil)
	sim.Log = func(message string, vals ...any) { vt.logs = append(vt.logs, fmt.Sprintf(message, vals...)) }
	return vt
}

// Deals 100 damage; the strike adds attack power to it, as a weapon strike does.
func (vt *versusLogTest) deal(spell *Spell, outcome HitOutcome) {
	baseDamage := 100.0
	if spell == vt.strike {
		baseDamage += spell.MeleeAttackPower(vt.target) / 14
	}
	result := spell.CalcDamage(vt.sim, vt.target, baseDamage, func(_ *Simulation, result *SpellResult, _ *AttackTable) { result.Outcome = outcome })
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
	lines := vt.linesOf(hit)
	if len(lines) == 0 {
		t.Fatalf("no damage line among %q", vt.logs)
	}
	return lines[0]
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
		{OutcomeHit, "Humanoid: {ItemID: 11} x1.05; {SpellID: 22} +81 AttackPower]"},
		{OutcomeCrit, "Humanoid: {ItemID: 11} x1.05; {SpellID: 33} x1.1 crit; {SpellID: 22} +81 AttackPower]"},
	} {
		if line := vt.damageLine(t, vt.strike, c.outcome); versusTag(line) != c.want {
			t.Errorf("%v line %q, want the tag %q", c.outcome, line, c.want)
		}
	}
	if line := vt.damageLine(t, vt.strike, OutcomeMiss); versusTag(line) != "" {
		t.Errorf("a miss is tagged: %q", line)
	}
}

func TestDamageLogTagsNoSourceOfAnotherMobType(t *testing.T) {
	vt := newVersusLogTest()
	undead := proto.MobType_MobTypeUndead
	vt.unit.MultiplyDamageVersus(ActionID{ItemID: 11}, 1.05, undead)
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81}, undead)

	if line := vt.damageLine(t, vt.strike, OutcomeHit); versusTag(line) != "" {
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
	if line, want := vt.damageLine(t, vt.strike, OutcomeHit), "Humanoid: {SpellID: 22} +81 AttackPower]"; versusTag(line) != want {
		t.Errorf("line %q while the aura is up, want the tag %q", line, want)
	}
	aura.Deactivate(sim)
	if line := vt.damageLine(t, vt.strike, OutcomeHit); versusTag(line) != "" {
		t.Errorf("line %q after the aura fell off, want no tag", line)
	}
}

func TestDamageLogTagsATickByTheDotsBonusCoefficient(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	vt := versusLogTestOn(sim, &fa.Character)
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81, stats.SpellDamage: 40}, vt.target.MobType)

	line := vt.lineOf(t, func() { vt.tick(fa.Spell, OutcomeHit) })
	if want := "Demon: {SpellID: 22} +40 SpellDamage]"; !strings.Contains(line, " tick ") || versusTag(line) != want {
		t.Errorf("tick line %q, want a tick with the tag %q", line, want)
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

	if line, want := vt.damageLine(t, bolt, OutcomeHit), "Humanoid (Orc): {SpellID: 101} x3 base; {SpellID: 22} +40 SpellDamage]"; versusTag(line) != want {
		t.Errorf("line %q, want the tag %q", line, want)
	}
}

// The tag body of a damage line, the text after " [Versus ", or "" when the line carries none.
func versusTag(line string) string {
	_, tag, _ := strings.Cut(line, " [Versus ")
	return tag
}

// A spell of the school on the melee table whose effects deal the damage baseDamage returns.
func versusTestHit(unit *Character, id int32, school SpellSchool, baseDamage func(*Spell, *Unit) float64) *Spell {
	return versusTestSpell(unit, id, school, func(sim *Simulation, target *Unit, spell *Spell) {
		spell.CalcAndDealDamage(sim, target, baseDamage(spell, target), spell.OutcomeAlwaysHitNoHitCounter)
	})
}

// A spell of the school on the melee table that runs apply.
func versusTestSpell(unit *Character, id int32, school SpellSchool, apply func(*Simulation, *Unit, *Spell)) *Spell {
	return unit.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: id},
		SpellSchool:      school,
		ProcMask:         ProcMaskMeleeMHSpecial,
		DefenseType:      DefenseTypeMelee,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects:     apply,
	})
}

func flatDamage(*Spell, *Unit) float64 { return 100 }

func attackPowerDamage(spell *Spell, target *Unit) float64 {
	return 100 + spell.MeleeAttackPower(target)/14
}

func TestDamageLogListsTheStatsTheDamageRead(t *testing.T) {
	vt := newVersusLogTest()
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81, stats.RangedAttackPower: 70, stats.SpellDamage: 40, stats.PhysicalDamage: 6}, vt.target.MobType)

	for i, c := range []struct {
		name       string
		school     SpellSchool
		baseDamage func(*Spell, *Unit) float64
		want       string
	}{
		{"a melee hit of flat holy damage", SpellSchoolHoly, flatDamage, ""},
		{"a flat physical strike, which takes the flat physical bonus", SpellSchoolPhysical, flatDamage, "Humanoid: {SpellID: 22} +6 PhysicalDamage]"},
		{"a holy hit on attack power", SpellSchoolHoly, attackPowerDamage, "Humanoid: {SpellID: 22} +81 AttackPower]"},
		{"a holy hit on ranged attack power", SpellSchoolHoly, func(spell *Spell, target *Unit) float64 {
			return 100 + spell.RangedAttackPower(target)/14
		}, "Humanoid: {SpellID: 22} +70 RangedAttackPower]"},
		{"a holy hit on attack power that adds spell damage by hand", SpellSchoolHoly, func(spell *Spell, target *Unit) float64 {
			return attackPowerDamage(spell, target) + spell.VersusStat(target, stats.SpellDamage)
		}, "Humanoid: {SpellID: 22} +40 SpellDamage; {SpellID: 22} +81 AttackPower]"},
	} {
		spell := versusTestHit(vt.unit, int32(101+i), c.school, c.baseDamage)
		line := vt.lineOf(t, func() { spell.SkipCastAndApplyEffects(vt.sim, vt.target) })
		if versusTag(line) != c.want {
			t.Errorf("%s: line %q, want the tag %q", c.name, line, c.want)
		}
	}
}

func TestDamageLogTagsATickByWhatTheTickRead(t *testing.T) {
	const hitTag = "Humanoid: {SpellID: 22} +70 RangedAttackPower]"
	rangedTick := func(sim *Simulation, target *Unit, dot *Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, 100+dot.Spell.RangedAttackPower(target)/14, dot.OutcomeTick)
	}
	flatTick := func(sim *Simulation, target *Unit, dot *Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, 100, dot.OutcomeTick)
	}
	// An area dot lives on its caster and ticks every enemy.
	flatAreaTick := func(sim *Simulation, _ *Unit, dot *Dot) {
		dot.Spell.CalcAndDealPeriodicAoeDamage(sim, 100, dot.OutcomeTick)
	}
	cases := []struct {
		name   string
		isAOE  bool
		onTick OnTick
		tick   string
		spell  *Spell
	}{
		{name: "a dot whose ticks read ranged attack power", onTick: rangedTick, tick: hitTag},
		{name: "a dot of flat ticks", onTick: flatTick, tick: ""},
		{name: "an area dot of flat ticks", isAOE: true, onTick: flatAreaTick, tick: ""},
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
				ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
					spell.CalcAndDealDamage(sim, target, 100+spell.RangedAttackPower(target)/14, spell.OutcomeAlwaysHitNoHitCounter)
				},
				Dot: DotConfig{
					Aura:          Aura{Label: fmt.Sprintf("Test Trap %d", id)},
					IsAOE:         cases[i].isAOE,
					NumberOfTicks: 1,
					TickLength:    time.Second,
					OnTick:        cases[i].onTick,
				},
			})
		}
	})
	vt := versusLogTestOn(sim, character)
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.RangedAttackPower: 70, stats.SpellDamage: 40}, vt.target.MobType)

	for _, c := range cases {
		if line := vt.lineOf(t, func() { c.spell.SkipCastAndApplyEffects(sim, vt.target) }); versusTag(line) != hitTag {
			t.Errorf("%s: hit line %q, want the tag %q", c.name, line, hitTag)
		}
		dot := c.spell.AOEDot()
		if !c.isAOE {
			dot = c.spell.Dot(vt.target)
		}
		line := vt.lineOf(t, func() { dot.TickOnce(sim) })
		if !strings.Contains(line, " tick ") || versusTag(line) != c.tick {
			t.Errorf("%s: tick line %q, want a tick with the tag %q", c.name, line, c.tick)
		}
	}
}

// A read no hit dealt, of damage only expected, of threat or before the cast, leaves no stat on a later hit.
func TestDamageLogNamesNoStatALeftoverReadTook(t *testing.T) {
	vt := newVersusLogTest()
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81}, vt.target.MobType)
	expected := vt.unit.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 101},
		SpellSchool:      SpellSchoolHoly,
		ProcMask:         ProcMaskMeleeMHSpecial,
		DefenseType:      DefenseTypeMelee,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ExpectedInitialDamage: func(sim *Simulation, target *Unit, spell *Spell, _ bool) *SpellResult {
			return spell.CalcDamage(sim, target, attackPowerDamage(spell, target), spell.OutcomeAlwaysHitNoHitCounter)
		},
	})
	threat := versusTestSpell(vt.unit, 102, SpellSchoolHoly, nil)
	threat.ThreatAPCoefficient = 1
	cast := versusTestHit(vt.unit, 103, SpellSchoolHoly, flatDamage)

	for _, c := range []struct {
		name string
		hit  func()
	}{
		{"a hit after its expected damage read attack power", func() {
			expected.ExpectedInitialDamage(vt.sim, vt.target)
			vt.deal(expected, OutcomeHit)
		}},
		{"a hit whose threat reads attack power", func() { vt.deal(threat, OutcomeHit) }},
		{"the next hit of that spell", func() { vt.deal(threat, OutcomeHit) }},
		{"a cast after attack power was read outside it", func() {
			cast.MeleeAttackPower(vt.target)
			cast.SkipCastAndApplyEffects(vt.sim, vt.target)
		}},
	} {
		if line := vt.lineOf(t, c.hit); versusTag(line) != "" {
			t.Errorf("%s: line %q, want no tag", c.name, line)
		}
	}
}

// The damage lines a hit logs, in order.
func (vt *versusLogTest) linesOf(hit func()) []string {
	vt.logs = nil
	hit()
	var lines []string
	for _, line := range vt.logs {
		if strings.Contains(line, "(Threat: ") {
			lines = append(lines, line)
		}
	}
	return lines
}

const attackPowerTag = "Humanoid: {SpellID: 22} +81 AttackPower]"

// Two hits one cast computes: the first reads attack power, the second deals flat damage.
func TestDamageLogNamesOnlyWhatEachHitRead(t *testing.T) {
	vt := newVersusLogTest()
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81}, vt.target.MobType)
	spell := versusTestSpell(vt.unit, 101, SpellSchoolHoly, func(sim *Simulation, target *Unit, spell *Spell) {
		read := spell.CalcDamage(sim, target, attackPowerDamage(spell, target), spell.OutcomeAlwaysHitNoHitCounter)
		flat := spell.CalcDamage(sim, target, 100, spell.OutcomeAlwaysHitNoHitCounter)
		spell.DealDamage(sim, read)
		spell.DealDamage(sim, flat)
	})

	lines := vt.linesOf(func() { spell.SkipCastAndApplyEffects(vt.sim, vt.target) })
	if len(lines) != 2 || versusTag(lines[0]) != attackPowerTag || versusTag(lines[1]) != "" {
		t.Errorf("lines %q, want the first tagged %q and the second untagged", lines, attackPowerTag)
	}
}

// A cleave or an area hit reads attack power once for all its targets, and the batched ones deal them
// after computing them all.
func TestDamageLogNamesAReadOnEveryResultItWentInto(t *testing.T) {
	request := fakeRageSimRequest()
	request.Encounter.Targets = append(request.Encounter.Targets, &proto.Target{Name: "second target", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2})
	sim := setupFakeRageSimFrom(request)
	vt := versusLogTestOn(sim, &sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Character)
	vt.unit.AddMobTypeBonusStats(ActionID{SpellID: 22}, stats.Stats{stats.AttackPower: 81}, vt.target.MobType)

	for i, c := range []struct {
		name string
		deal func(sim *Simulation, target *Unit, spell *Spell, baseDamage float64)
	}{
		{"a cleave dealt last target first", func(sim *Simulation, target *Unit, spell *Spell, baseDamage float64) {
			results := spell.CalcCleaveDamage(sim, target, 2, baseDamage, spell.OutcomeAlwaysHitNoHitCounter)
			for i := len(results) - 1; i >= 0; i-- {
				spell.DealDamage(sim, results[i])
			}
		}},
		{"a batched area hit", func(sim *Simulation, _ *Unit, spell *Spell, baseDamage float64) {
			spell.CalcAoeDamage(sim, baseDamage, spell.OutcomeAlwaysHitNoHitCounter)
			spell.DealBatchedAoeDamage(sim)
		}},
		{"an area hit dealt target by target", func(sim *Simulation, _ *Unit, spell *Spell, baseDamage float64) {
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeAlwaysHitNoHitCounter)
		}},
	} {
		spell := versusTestSpell(vt.unit, int32(101+i), SpellSchoolHoly, func(sim *Simulation, target *Unit, spell *Spell) {
			c.deal(sim, target, spell, attackPowerDamage(spell, target))
		})
		lines := vt.linesOf(func() { spell.SkipCastAndApplyEffects(sim, vt.target) })
		if len(lines) != 2 || versusTag(lines[0]) != attackPowerTag || versusTag(lines[1]) != attackPowerTag {
			t.Errorf("%s: lines %q, want two, each tagged %q", c.name, lines, attackPowerTag)
		}
	}
}

func TestDamageLogNamesAnItemsStatsByTheItem(t *testing.T) {
	vt := newVersusLogTest()
	item := &Item{ID: 12345, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {MobTypeStats: []*proto.MobTypeStats{
		{MobTypes: []proto.MobType{vt.target.MobType}, Stats: map[int32]float64{int32(stats.AttackPower): 81}},
	}}}}
	vt.unit.addItemMobTypeStats(item, 1)

	if line, want := vt.damageLine(t, vt.strike, OutcomeHit), "Humanoid: {ItemID: 12345} +81 AttackPower]"; versusTag(line) != want {
		t.Errorf("line %q, want the tag %q", line, want)
	}
}

func TestDamageLogNamesACritChanceTheTargetsFamilyRaises(t *testing.T) {
	orcs := &VersusBonus{Multiplier: 2, Families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyOrc}}
	for _, c := range []struct {
		family   proto.CreatureFamily
		wantCrit float64
		want     string
	}{
		{proto.CreatureFamily_CreatureFamilyOrc, 10, "Humanoid (Orc): {ItemID: 13965} +10% crit chance]"},
		{proto.CreatureFamily_CreatureFamilyUnknown, 5, ""},
	} {
		vt := newVersusLogTest()
		vt.target.CreatureFamily = c.family
		table := vt.unit.AttackTables[vt.target.UnitIndex]
		vt.unit.AddCritChanceVersus(ActionID{ItemID: 13965}, vt.target, 5, orcs)

		if table.BonusCritPercent != c.wantCrit {
			t.Errorf("against %v the crit chance is %v, want %v", c.family, table.BonusCritPercent, c.wantCrit)
		}
		if line := vt.damageLine(t, vt.strike, OutcomeHit); versusTag(line) != c.want {
			t.Errorf("against %v line %q, want the tag %q", c.family, line, c.want)
		}
		vt.unit.AddCritChanceVersus(ActionID{ItemID: 13965}, vt.target, -5, orcs)
		if line := vt.damageLine(t, vt.strike, OutcomeHit); table.BonusCritPercent != 0 || versusTag(line) != "" {
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
		{"a likelier proc on a Dwarf", 3, proto.CreatureFamily_CreatureFamilyDwarf, "Humanoid (Dwarf): {ItemID: 11815} x3 proc chance]"},
		{"a proc on Dwarves alone on a Dwarf", 0, proto.CreatureFamily_CreatureFamilyDwarf, "Humanoid (Dwarf): {ItemID: 11815} proc only]"},
		{"a likelier proc on another target", 3, proto.CreatureFamily_CreatureFamilyUnknown, ""},
	} {
		vt := newVersusLogTest()
		vt.target.CreatureFamily = c.family
		vt.strike.VersusProc = &VersusProc{Source: ActionID{ItemID: 11815}, VersusBonus: VersusBonus{Multiplier: c.mult, Families: dwarves}}

		if line := vt.damageLine(t, vt.strike, OutcomeHit); versusTag(line) != c.want {
			t.Errorf("%s: line %q, want the tag %q", c.name, line, c.want)
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
