package shared

import (
	"fmt"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	auraBoltID     int32 = 992001
	auraStrikeID   int32 = 992002
	auraManaBoltID int32 = 992003

	// Nature Aligned, Natural Alignment Crystal 19344's on-use: magic spell damage, healing and the mana
	// cost of magic spells (A_MOD_POWER_COST_SCHOOL_PCT) raised on the wearer. The test's on-use puts
	// it on a 5 min cooldown in category 1141 for 20 s.
	natureAligned int32 = 23734

	// Obsidian Mail Tunic 22191's equip spell: A_MOD_DAMAGE_TAKEN of magic schools, on the wearer.
	spellDamageReduction int32 = 27518

	// Beastmaster's Boots 22061 and Treads 226881: 27206 periodically casts 27205,
	// A_MOD_DAMAGE_PERCENT_DONE on the pet. Beastmaster's Tunic 22060 and 226886: 27225 casts 27208,
	// A_MOD_BASE_RESISTANCE_PCT of armor on the pet, the same way. Beastmaster's Cap 22013: 27038 casts
	// 27039, A_MOD_INCREASE_HEALTH_PERCENT, and Beastmaster's Gloves 22015: 27043 casts 27042,
	// A_MOD_WEAPON_CRIT_PERCENT, both on the pet the same way.
	increasedPetDamage int32 = 27206
	increasedPetArmor  int32 = 27225
	increasedPetHealth int32 = 27038
	increasedPetCrit   int32 = 27043

	// Nightfall 19169's chance on hit: 23605, A_MOD_DAMAGE_PERCENT_TAKEN of magic schools on the
	// target. The tests hand the proc a chance of their own.
	spellVulnerability int32 = 23605

	// Sword of Zeal 6622's chance on hit: 8191, physical damage done (A_MOD_DAMAGE_DONE) and armor
	// (A_MOD_ARMOR_NO_MODS), which the test's item effect carries as 10 and 150. The tests hand the proc
	// a chance of their own.
	zeal int32 = 8191

	// Savior's Sacrifice, Ring of Saviors 1447's on-use: armor on the wearer (A_MOD_ARMOR_NO_MODS).
	saviorsSacrifice int32 = 18826

	// Lesser Arcane Amalgamation 1504's equip spell: armor on the wearer (A_MOD_ARMOR_NO_MODS).
	arcaneAmalgamationArmor int32 = 15390

	petClawID int32 = 992004
)

func init() {
	core.RegisterAgentFactory(
		proto.Player_RestorationShaman{},
		proto.Spec_SpecRestorationShaman,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			tester := &auraTester{Character: *character}
			tester.EnableAutoAttacks(tester, core.AutoAttackOptions{
				MainHand:       tester.WeaponFromMainHand(),
				AutoSwingMelee: true,
			})
			return tester
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_RestorationShaman)
		},
	)
	core.RegisterAgentFactory(
		proto.Player_Hunter{},
		proto.Spec_SpecHunter,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			owner := &auraTester{Character: *character}
			owner.pet = &testPet{Pet: core.NewPet(core.PetConfig{
				Name: "Test Pet", Owner: &owner.Character, EnabledOnStart: true,
				BaseStats:       stats.Stats{stats.Health: 1000},
				StatInheritance: func(stats.Stats) stats.Stats { return stats.Stats{} },
			})}
			owner.AddPet(owner.pet)
			return owner
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_Hunter)
		},
	)
}

// A wearer with an arcane bolt, a physical strike and an arcane bolt costing 100 mana, each dealing
// 100 that no resistance or armor reduces, and a pet where the class has one.
type auraTester struct {
	core.Character
	bolt, strike, manaBolt *core.Spell
	pet                    *testPet
}

// A pet with 1000 health whose claw deals 100 physical damage that armor does not reduce.
type testPet struct {
	core.Pet
	claw *core.Spell
}

func (p *testPet) GetPet() *core.Pet                        { return &p.Pet }
func (p *testPet) ApplyTalents()                            {}
func (p *testPet) Reset(_ *core.Simulation)                 {}
func (p *testPet) OnEncounterStart(_ *core.Simulation)      {}
func (p *testPet) ExecuteCustomRotation(_ *core.Simulation) {}

func (p *testPet) Initialize() {
	p.Pet.Initialize()
	p.claw = p.RegisterSpell(testHit(petClawID, core.SpellSchoolPhysical, core.ProcMaskMeleeMHSpecial))
}

func (c *auraTester) GetCharacter() *core.Character       { return &c.Character }
func (c *auraTester) ApplyTalents()                       {}
func (c *auraTester) Reset(_ *core.Simulation)            {}
func (c *auraTester) OnGCDReady(_ *core.Simulation)       {}
func (c *auraTester) OnEncounterStart(_ *core.Simulation) {}

func (c *auraTester) Initialize() {
	c.bolt = c.RegisterSpell(testHit(auraBoltID, core.SpellSchoolArcane, core.ProcMaskSpellDamage))
	c.strike = c.RegisterSpell(testHit(auraStrikeID, core.SpellSchoolPhysical, core.ProcMaskMeleeMHSpecial))

	manaBolt := testHit(auraManaBoltID, core.SpellSchoolArcane, core.ProcMaskSpellDamage)
	manaBolt.ManaCost = core.ManaCostOptions{FlatCost: 100}
	c.manaBolt = c.RegisterSpell(manaBolt)
}

func testHit(id int32, school core.SpellSchool, mask core.ProcMask) core.SpellConfig {
	return core.SpellConfig{
		ActionID:         core.ActionID{SpellID: id},
		SpellSchool:      school,
		ProcMask:         mask,
		Flags:            core.SpellFlagIgnoreResists,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		},
	}
}

func withAuraItem(id int32, itemType proto.ItemType, effects ...*proto.ItemEffect) {
	item := &proto.SimItem{Id: id, Name: "Test Aura Item", Type: itemType,
		ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}, ItemEffects: effects}
	if itemType == proto.ItemType_ItemTypeWeapon {
		item = testOneHander(id)
		item.ItemEffects = effects
	}
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{item}})
}

func auraPlayer(name string, class proto.Class, spec interface{}, slotted map[proto.ItemSlot]int32) *proto.Player {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	for slot, id := range slotted {
		items[slot] = &proto.ItemSpec{Id: id}
	}

	player := &proto.Player{
		Name:      name,
		Class:     class,
		Race:      proto.Race_RaceOrc,
		Buffs:     &proto.IndividualBuffs{},
		Equipment: &proto.EquipmentSpec{Items: items},
	}
	switch spec := spec.(type) {
	case *proto.Player_RestorationShaman:
		player.Spec = spec
	case *proto.Player_Hunter:
		player.Spec = spec
	}
	return player
}

func shaman(name string, slotted map[proto.ItemSlot]int32) *proto.Player {
	return auraPlayer(name, proto.Class_ClassShaman, &proto.Player_RestorationShaman{}, slotted)
}

func newAuraSim(players ...*proto.Player) *core.Simulation {
	return newAuraSimIn(nil, players...)
}

func newAuraSimIn(areas []proto.AreaType, players ...*proto.Player) *core.Simulation {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: players,
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:   []*proto.Target{{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon}},
			Duration:  180,
			AreaTypes: areas,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	for _, player := range sim.Raid.Parties[0].Players {
		player.GetCharacter().AutoAttacks.CancelAutoSwing(sim)
	}
	return sim
}

func auraTesterAt(sim *core.Simulation, index int) *auraTester {
	return sim.Raid.Parties[0].Players[index].(*auraTester)
}

func dealt(sim *core.Simulation, spell *core.Spell, target *core.Unit) float64 {
	before := spell.SpellMetrics[target.UnitIndex].TotalDamage
	spell.Cast(sim, target)
	return spell.SpellMetrics[target.UnitIndex].TotalDamage - before
}

// The multiplier an aura stating a percentage applies. A percentage of 0 leaves the test nothing to
// check, and fails it.
func percentMultiplier(t *testing.T, e *spelldata.Effect) float64 {
	t.Helper()
	needAmount(t, "the aura's percentage", e.Mean())
	return 1 + e.Mean()/100
}

// Natural Alignment Crystal raises the wearer's spell damage, healing and its spells' mana cost by the
// row's shares for the row's duration, and leaves a physical hit alone.
func TestNaturalAlignmentCrystal(t *testing.T) {
	const itemID int32 = 992101
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, testOnUse(natureAligned, 300000, 1141, 20000))
	NewSpellDataAuraOnUse(itemID)
	row := spelldata.MustFind(natureAligned)
	damage := percentMultiplier(t, row.FirstAura(dbcenums.A_MOD_DAMAGE_PERCENT_DONE))
	healing := percentMultiplier(t, row.FirstAura(dbcenums.A_MOD_HEALING_DONE_PERCENT))
	cost := percentMultiplier(t, row.FirstAura(dbcenums.A_MOD_POWER_COST_SCHOOL_PCT))

	sim := newAuraSim(shaman("Shaman", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID}))
	tester := auraTesterAt(sim, 0)
	target := sim.Encounter.ActiveTargetUnits[0]
	crystal := tester.GetSpell(core.ActionID{ItemID: itemID})
	if crystal == nil {
		t.Fatalf("the crystal registered no on-use")
	}

	check := func(when string, bolt, healing, cost float64) {
		t.Helper()
		if got := dealt(sim, tester.bolt, target); !core.WithinToleranceFloat64(bolt, got, 1e-6) {
			t.Errorf("%s: the arcane bolt dealt %v, want %v", when, got, bolt)
		}
		if got := dealt(sim, tester.strike, target); !core.WithinToleranceFloat64(100, got, 1e-6) {
			t.Errorf("%s: the physical strike dealt %v, want 100", when, got)
		}
		if got := tester.PseudoStats.HealingDealtMultiplier; !core.WithinToleranceFloat64(healing, got, 1e-6) {
			t.Errorf("%s: healing dealt multiplier %v, want %v", when, got, healing)
		}
		if got := tester.manaBolt.Cost.GetCurrentCost(); !core.WithinToleranceFloat64(cost, got, 1e-6) {
			t.Errorf("%s: the mana bolt costs %v, want %v", when, got, cost)
		}
	}

	check("before use", 100, 1, 100)

	if !crystal.Cast(sim, target) {
		t.Fatalf("the crystal could not be used")
	}
	start := sim.CurrentTime
	check("while Nature Aligned is up", 100*damage, healing, 100*cost)

	if got := crystal.CD.ReadyAt(); got != start+5*time.Minute {
		t.Errorf("the crystal is ready again at %v, want 5 min after its use at %v", got, start)
	}

	stepPast(t, sim, start+row.Duration()+time.Millisecond)
	check("after its duration", 100, 1, 100)
}

func hunter(name string, slotted map[proto.ItemSlot]int32) *proto.Player {
	return auraPlayer(name, proto.Class_ClassHunter, &proto.Player_Hunter{}, slotted)
}

// The bolt of auraTester, 100, once the wearer of 27518 has taken its row's amount off it.
func reducedBolt(t *testing.T) float64 {
	t.Helper()
	reduction := spelldata.MustFind(spellDamageReduction).FirstAura(dbcenums.A_MOD_DAMAGE_TAKEN).Mean()
	needAmount(t, "27518's reduction", reduction)
	return 100 + reduction
}

// Obsidian Mail Tunic takes its row's amount off every magic hit its wearer takes, and nothing off a
// physical one or off a hit on anyone else.
func TestObsidianMailTunicReducesAMagicHit(t *testing.T) {
	const itemID int32 = 992201
	withAuraItem(itemID, proto.ItemType_ItemTypeChest)
	NewSpellDataEquipAura(SpellDataProc{TriggerSpellID: spellDamageReduction},
		[]ItemVariant{{ItemID: itemID, ItemName: "Test Obsidian Mail"}})

	sim := newAuraSim(
		shaman("Wearer", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotChest: itemID}),
		shaman("Attacker", nil),
	)
	wearer, attacker := auraTesterAt(sim, 0), auraTesterAt(sim, 1)

	if got, want := dealt(sim, attacker.bolt, &wearer.Unit), reducedBolt(t); !core.WithinToleranceFloat64(want, got, 1e-6) {
		t.Errorf("an arcane bolt on the wearer dealt %v, want %v", got, want)
	}
	if got := dealt(sim, attacker.strike, &wearer.Unit); !core.WithinToleranceFloat64(100, got, 1e-6) {
		t.Errorf("a physical strike on the wearer dealt %v, want 100", got)
	}
	if got := dealt(sim, wearer.bolt, &attacker.Unit); !core.WithinToleranceFloat64(100, got, 1e-6) {
		t.Errorf("the wearer's arcane bolt on someone else dealt %v, want 100", got)
	}
}

// A row restricted to an area applies in an encounter there and nowhere else. 27518 states no area,
// so a copy requiring Forest and Grassland (group 9161) stands in for the rows that do.
func TestEquipAuraAppliesOnlyInItsArea(t *testing.T) {
	const itemID int32 = 992204
	editRow(t, spellDamageReduction, func(s *spelldata.Spell) { s.RequiredAreas = 9161 })
	withAuraItem(itemID, proto.ItemType_ItemTypeChest)
	NewSpellDataEquipAura(SpellDataProc{TriggerSpellID: spellDamageReduction},
		[]ItemVariant{{ItemID: itemID, ItemName: "Test Forest Mail"}})

	for _, c := range []struct {
		areas []proto.AreaType
		want  float64
	}{
		{nil, 100},
		{[]proto.AreaType{proto.AreaType_AreaTypeForestGrassland}, reducedBolt(t)},
	} {
		sim := newAuraSimIn(c.areas,
			shaman("Wearer", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotChest: itemID}),
			shaman("Attacker", nil),
		)
		wearer, attacker := auraTesterAt(sim, 0), auraTesterAt(sim, 1)
		if got := dealt(sim, attacker.bolt, &wearer.Unit); !core.WithinToleranceFloat64(c.want, got, 1e-6) {
			t.Errorf("in %v an arcane bolt on the wearer dealt %v, want %v", c.areas, got, c.want)
		}
	}
}

// Beastmaster's Boots raise the pet's damage by the share their triggered row states, and do nothing
// for a wearer with no pet. The Tunic's pet armor share has no equipment share on a pet to scale, so it
// registers nothing.
func TestBeastmastersBootsRaiseThePetsDamage(t *testing.T) {
	const bootsID, tunicID int32 = 992202, 992203
	withAuraItem(bootsID, proto.ItemType_ItemTypeFeet)
	withAuraItem(tunicID, proto.ItemType_ItemTypeChest)
	NewSpellDataEquipAura(SpellDataProc{TriggerSpellID: increasedPetDamage},
		[]ItemVariant{{ItemID: bootsID, ItemName: "Test Beastmaster's Boots"}})
	NewSpellDataEquipAura(SpellDataProc{TriggerSpellID: increasedPetArmor},
		[]ItemVariant{{ItemID: tunicID, ItemName: "Test Beastmaster's Tunic"}})

	gear := map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotFeet: bootsID, proto.ItemSlot_ItemSlotChest: tunicID}
	sim := newAuraSim(hunter("Owner", gear), hunter("Bare Owner", nil), shaman("Petless", gear))
	target := sim.Encounter.ActiveTargetUnits[0]
	owner, bare, petless := auraTesterAt(sim, 0), auraTesterAt(sim, 1), auraTesterAt(sim, 2)

	petAura := spelldata.MustFind(increasedPetDamage).FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL).Trigger()
	raised := percentMultiplier(t, petAura.FirstAura(dbcenums.A_MOD_DAMAGE_PERCENT_DONE))
	bareClaw := dealt(sim, bare.pet.claw, target)
	if got := dealt(sim, owner.pet.claw, target); !core.WithinToleranceFloat64(raised*bareClaw, got, 1e-6) {
		t.Errorf("the pet of the wearer clawed for %v, want %v times the %v of a pet without", got, raised, bareClaw)
	}
	if got := dealt(sim, owner.strike, target); !core.WithinToleranceFloat64(100, got, 1e-6) {
		t.Errorf("the wearer's own strike dealt %v, want the 100 a pet aura leaves alone", got)
	}
	if got := dealt(sim, petless.strike, target); !core.WithinToleranceFloat64(100, got, 1e-6) {
		t.Errorf("a wearer with no pet struck for %v, want 100", got)
	}
	if got := owner.pet.GetStat(stats.Armor); got != bare.pet.GetStat(stats.Armor) {
		t.Errorf("the tunic changed the pet's armor to %v", got)
	}
}

// Beastmaster's Cap raises the pet's maximum health and Beastmaster's Gloves its crit by what their
// triggered rows state, leaving the wearer's own alone, and do nothing for a wearer with no pet.
func TestBeastmastersCapAndGlovesRaiseThePetsHealthAndCrit(t *testing.T) {
	const capID, glovesID int32 = 992205, 992206
	withAuraItem(capID, proto.ItemType_ItemTypeHead)
	withAuraItem(glovesID, proto.ItemType_ItemTypeHands)
	NewSpellDataEquipAura(SpellDataProc{TriggerSpellID: increasedPetHealth},
		[]ItemVariant{{ItemID: capID, ItemName: "Test Beastmaster's Cap"}})
	NewSpellDataEquipAura(SpellDataProc{TriggerSpellID: increasedPetCrit},
		[]ItemVariant{{ItemID: glovesID, ItemName: "Test Beastmaster's Gloves"}})

	gear := map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotHead: capID, proto.ItemSlot_ItemSlotHands: glovesID}
	sim := newAuraSim(hunter("Owner", gear), hunter("Bare Owner", nil), shaman("Petless", gear), shaman("Bare Petless", nil))
	owner, bare := auraTesterAt(sim, 0), auraTesterAt(sim, 1)
	petless, barePetless := auraTesterAt(sim, 2), auraTesterAt(sim, 3)

	triggered := func(id int32) *spelldata.Spell {
		return spelldata.MustFind(id).FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL).Trigger()
	}
	health := percentMultiplier(t, triggered(increasedPetHealth).FirstAura(dbcenums.A_MOD_INCREASE_HEALTH_PERCENT))
	crit := triggered(increasedPetCrit).FirstAura(dbcenums.A_MOD_WEAPON_CRIT_PERCENT).Mean()
	needAmount(t, "27042's crit", crit)
	if got, bareHealth := owner.pet.MaxHealth(), bare.pet.MaxHealth(); bareHealth == 0 || !core.WithinToleranceFloat64(health*bareHealth, got, 1e-6) {
		t.Errorf("the pet of the wearer has %v maximum health, want %v times the %v of a pet without", got, health, bareHealth)
	}
	if got, bareCrit := owner.pet.GetStat(stats.PhysicalCritPercent), bare.pet.GetStat(stats.PhysicalCritPercent); !core.WithinToleranceFloat64(bareCrit+crit, got, 1e-6) {
		t.Errorf("the pet of the wearer has %v crit, want %v over the %v of a pet without", got, crit, bareCrit)
	}

	for _, c := range []struct {
		who          string
		wearer, bare *auraTester
	}{{"the wearer", owner, bare}, {"a wearer with no pet", petless, barePetless}} {
		if got, want := c.wearer.MaxHealth(), c.bare.MaxHealth(); got != want {
			t.Errorf("%s has %v maximum health, want the %v of one without", c.who, got, want)
		}
		if got, want := c.wearer.GetStat(stats.PhysicalCritPercent), c.bare.GetStat(stats.PhysicalCritPercent); got != want {
			t.Errorf("%s has %v crit, want the %v of one without", c.who, got, want)
		}
	}
}

// Lands the wielder's strike on the target, then steps past the batch window a proc's handler waits.
func strikeAndStep(t *testing.T, sim *core.Simulation, wielder *auraTester, target *core.Unit) {
	t.Helper()
	wielder.strike.Cast(sim, target)
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow+time.Millisecond)
}

// Nightfall's debuff raises the spell damage the target takes by the row's share for the row's
// duration, whoever casts at it, and leaves physical hits alone.
func TestNightfallRaisesTheTargetsSpellDamageTaken(t *testing.T) {
	const itemID int32 = 992301
	withAuraItem(itemID, proto.ItemType_ItemTypeWeapon,
		&proto.ItemEffect{BuffId: spellVulnerability, Effect: &proto.ItemEffect_Proc{Proc: &proto.ProcEffect{}}})
	everyHit(t, spellVulnerability)
	NewSpellDataAuraProc(SpellDataProc{TriggerSpellID: spellVulnerability, IsWeaponProc: true},
		[]ItemVariant{{ItemID: itemID, ItemName: "Test Nightfall"}})
	row := spelldata.MustFind(spellVulnerability)
	raised := 100 * percentMultiplier(t, row.FirstAura(dbcenums.A_MOD_DAMAGE_PERCENT_TAKEN))

	sim := newAuraSim(
		shaman("Wielder", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: itemID}),
		shaman("Caster", nil),
	)
	wielder, caster := auraTesterAt(sim, 0), auraTesterAt(sim, 1)
	target := sim.Encounter.ActiveTargetUnits[0]

	if got := dealt(sim, caster.bolt, target); !core.WithinToleranceFloat64(100, got, 1e-6) {
		t.Fatalf("before the proc a second caster's bolt dealt %v, want 100", got)
	}

	strikeAndStep(t, sim, wielder, target)
	debuff := target.GetAuraByID(core.ActionID{SpellID: spellVulnerability})
	if !debuff.IsActive() {
		t.Fatalf("a strike at 100%% chance put no Spell Vulnerability on the target")
	}
	if debuff.Duration != row.Duration() {
		t.Errorf("Spell Vulnerability lasts %v, want the row's %v", debuff.Duration, row.Duration())
	}

	if got := dealt(sim, caster.bolt, target); !core.WithinToleranceFloat64(raised, got, 1e-6) {
		t.Errorf("with the debuff up a second caster's bolt dealt %v, want %v", got, raised)
	}
	if got := dealt(sim, wielder.bolt, target); !core.WithinToleranceFloat64(raised, got, 1e-6) {
		t.Errorf("with the debuff up the wielder's bolt dealt %v, want %v", got, raised)
	}
	if got := dealt(sim, caster.strike, target); !core.WithinToleranceFloat64(100, got, 1e-6) {
		t.Errorf("with the debuff up a physical strike dealt %v, want 100", got)
	}

	// The wielder's bolt is a hit its 100% proc answers too, so the duration runs from the refresh it lands.
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow+time.Millisecond)
	stepPast(t, sim, debuff.ExpiresAt()+time.Millisecond)
	if debuff.IsActive() {
		t.Errorf("Spell Vulnerability is still up after its %v", row.Duration())
	}
	if got := dealt(sim, caster.bolt, target); !core.WithinToleranceFloat64(100, got, 1e-6) {
		t.Errorf("after the debuff a second caster's bolt dealt %v, want 100", got)
	}
}

// Sword of Zeal's buff adds the item effect's 10 to each physical hit and 150 armor for its row's
// duration.
func TestSwordOfZealAddsPhysicalDamageAndArmor(t *testing.T) {
	duration := spelldata.MustFind(zeal).Duration()
	const itemID int32 = 992302
	withAuraItem(itemID, proto.ItemType_ItemTypeWeapon, &proto.ItemEffect{
		BuffId:           zeal,
		EffectDurationMs: 15000,
		Effect:           &proto.ItemEffect_Proc{Proc: &proto.ProcEffect{}},
		ScalingOptions: map[int32]*proto.ScalingItemEffectProperties{0: {Stats: map[int32]float64{
			int32(proto.Stat_StatArmorNoMods): 150, int32(proto.Stat_StatPhysicalDamage): 10,
		}}},
	})
	everyHit(t, zeal)
	NewSpellDataProc(SpellDataProc{TriggerSpellID: zeal, IsWeaponProc: true},
		[]ItemVariant{{ItemID: itemID, ItemName: "Test Sword of Zeal"}})

	sim := newAuraSim(shaman("Wielder", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: itemID}))
	wielder := auraTesterAt(sim, 0)
	target := sim.Encounter.ActiveTargetUnits[0]
	armor := wielder.GetStat(stats.Armor)

	strikeAndStep(t, sim, wielder, target)
	procTime := sim.CurrentTime
	if got := wielder.GetStat(stats.Armor) - armor; !core.WithinToleranceFloat64(150, got, 1e-6) {
		t.Errorf("Zeal added %v armor, want 150", got)
	}
	if got := dealt(sim, wielder.strike, target); !core.WithinToleranceFloat64(110, got, 1e-6) {
		t.Errorf("with Zeal up a physical strike dealt %v, want 110", got)
	}
	if got := dealt(sim, wielder.bolt, target); !core.WithinToleranceFloat64(100, got, 1e-6) {
		t.Errorf("with Zeal up an arcane bolt dealt %v, want 100", got)
	}

	stepPast(t, sim, procTime+duration-time.Millisecond)
	if got := wielder.GetStat(stats.Armor) - armor; !core.WithinToleranceFloat64(150, got, 1e-6) {
		t.Errorf("Zeal's armor is %v just before its %v run out, want 150", got, duration)
	}
	// The strike above procced again, a batch window after it landed.
	stepPast(t, sim, procTime+duration+core.SpellBatchWindow+time.Millisecond)
	if got := wielder.GetStat(stats.Armor) - armor; got != 0 {
		t.Errorf("Zeal's armor is %v after its %v, want none", got, duration)
	}
	if got := dealt(sim, wielder.bolt, target); !core.WithinToleranceFloat64(100, got, 1e-6) {
		t.Errorf("after Zeal an arcane bolt dealt %v, want 100", got)
	}
}

// The buff at stacks, or unstacked for 0, adds want armor to the wearer, and none is left once it fades.
func checkBuffArmor(t *testing.T, sim *core.Simulation, wearer *auraTester, buff *core.Aura, stacks int32, want float64, what string) {
	t.Helper()
	before := wearer.GetStat(stats.Armor)
	buff.Activate(sim)
	if stacks > 0 {
		buff.SetStacks(sim, stacks)
	}
	if got := wearer.GetStat(stats.Armor) - before; !core.WithinToleranceFloat64(want, got, 1e-6) {
		t.Errorf("%s: the buff added %v armor, want %v", what, got, want)
	}
	buff.Deactivate(sim)
	if got := wearer.GetStat(stats.Armor) - before; !core.WithinToleranceFloat64(0, got, 1e-6) {
		t.Errorf("%s: %v armor is left after the buff fades, want none", what, got)
	}
}

// An item buff's ArmorNoMods, the stat the generator writes for a buff row's A_MOD_ARMOR_NO_MODS armor,
// stays outside the wearer's percent armor modifiers on an on-use, a proc and a stacking proc alike: the
// wearer gains it whole for each stack and loses as much when the buff fades. The same amount of Armor
// gains armor the modifiers scale. The wearer's neck doubles its armor.
func TestNoModsArmorFromAnItemStaysOutOfThePercentArmorModifiers(t *testing.T) {
	noModsArmorFixtures(t)
	armor := spelldata.MustFind(saviorsSacrifice).FirstAura(dbcenums.A_MOD_ARMOR_NO_MODS).BasePoints
	needAmount(t, "18826's armor", armor)

	// The stacking cases come last: they give the row a stack count for the rest of the test.
	const stacks = 3
	for i, tc := range []struct {
		stat   proto.Stat
		proc   bool
		stacks int32
		want   float64
	}{
		{proto.Stat_StatArmorNoMods, false, 0, armor},
		{proto.Stat_StatArmor, false, 0, armor * armorScaling},
		{proto.Stat_StatArmorNoMods, true, 0, armor},
		{proto.Stat_StatArmor, true, 0, armor * armorScaling},
		{proto.Stat_StatArmorNoMods, true, stacks, armor * stacks},
		{proto.Stat_StatArmor, true, stacks, armor * armorScaling * stacks},
	} {
		itemID := 992312 + int32(i)
		label := fmt.Sprintf("Test Armor %d", i)
		entry := &proto.ItemEffect{BuffId: saviorsSacrifice, BuffName: label, EffectDurationMs: 20000,
			ScalingOptions: map[int32]*proto.ScalingItemEffectProperties{0: {Stats: map[int32]float64{int32(tc.stat): armor}}}}
		slot := proto.ItemSlot_ItemSlotTrinket1
		if tc.proc {
			entry.Effect = &proto.ItemEffect_Proc{Proc: &proto.ProcEffect{}}
			withAuraItem(itemID, proto.ItemType_ItemTypeWeapon, entry)
			everyHit(t, saviorsSacrifice)
			if tc.stacks > 0 {
				editRow(t, saviorsSacrifice, func(s *spelldata.Spell) { s.MaxStack = int16(tc.stacks) })
			}
			NewSpellDataProc(SpellDataProc{TriggerSpellID: saviorsSacrifice, IsWeaponProc: true},
				[]ItemVariant{{ItemID: itemID, ItemName: label}})
			label += " Proc"
			slot = proto.ItemSlot_ItemSlotMainHand
		} else {
			entry.Effect = &proto.ItemEffect_OnUse{OnUse: &proto.OnUseEffect{CooldownMs: 120000}}
			withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, entry)
			NewSimpleStatActive(itemID)
		}

		sim := newAuraSim(shaman("Wearer", map[proto.ItemSlot]int32{slot: itemID, proto.ItemSlot_ItemSlotNeck: armorScalingNeckID}))
		wearer := auraTesterAt(sim, 0)
		buff := wearer.GetAura(label)
		if buff == nil || buff.IsActive() {
			t.Fatalf("%v (proc %v): the item registered no buff %q, or it is up before it is used", tc.stat, tc.proc, label)
		}
		checkBuffArmor(t, sim, wearer, buff, tc.stacks, tc.want, fmt.Sprintf("%v (proc %v, %d stacks)", tc.stat, tc.proc, tc.stacks))
	}
}

// A stacking on-use item's ArmorNoMods stays outside the percent armor modifiers for each stack, and
// the wearer loses it whole when the buff fades. The same amount of Armor gains armor the modifiers
// scale. The wearer's neck doubles its armor.
func TestNoModsArmorFromAStackingOnUseStaysOutOfThePercentArmorModifiers(t *testing.T) {
	noModsArmorFixtures(t)
	armor := spelldata.MustFind(saviorsSacrifice).FirstAura(dbcenums.A_MOD_ARMOR_NO_MODS).BasePoints
	needAmount(t, "18826's armor", armor)

	const stacks = 3
	for i, tc := range []struct {
		stat proto.Stat
		want float64
	}{
		{proto.Stat_StatArmorNoMods, armor * stacks},
		{proto.Stat_StatArmor, armor * armorScaling * stacks},
	} {
		itemID := 992331 + int32(i)
		name := fmt.Sprintf("Test Stacking Armor %d", i)
		withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, &proto.ItemEffect{
			BuffId: saviorsSacrifice, BuffName: name, EffectDurationMs: 20000, MaxCumulativeStacks: stacks,
			Effect:         &proto.ItemEffect_OnUse{OnUse: &proto.OnUseEffect{CooldownMs: 120000}},
			ScalingOptions: map[int32]*proto.ScalingItemEffectProperties{0: {Stats: map[int32]float64{int32(tc.stat): armor}}},
		})
		NewStackingStatBonusCD(StackingStatBonusCD{Name: name, ID: itemID, CD: 2 * time.Minute,
			Callback: core.CallbackOnSpellHitTaken, ProcMask: core.ProcMaskMelee})

		sim := newAuraSim(shaman("Wearer", map[proto.ItemSlot]int32{
			proto.ItemSlot_ItemSlotTrinket1: itemID, proto.ItemSlot_ItemSlotNeck: armorScalingNeckID}))
		wearer := auraTesterAt(sim, 0)
		buff := wearer.GetAura(name + " Proc")
		if buff == nil || buff.IsActive() {
			t.Fatalf("%v: the item registered no buff %q, or it is up before it is used", tc.stat, name+" Proc")
		}
		checkBuffArmor(t, sim, wearer, buff, stacks, tc.want, fmt.Sprintf("%v at %d stacks", tc.stat, stacks))
	}
}

const (
	noModsArmorEnchantID int32 = 992321 + iota
	plainArmorEnchantID
	testChestID
	armorScalingNeckID

	// The neck's scaling of armor from items, and its percent armor modifier.
	armorScaling = 2
)

// A test chest, an enchant for it of ArmorNoMods and one of the same Armor, and a neck that doubles
// the wearer's armor from items and then doubles its armor. The amount is the armor Lesser Arcane
// Amalgamation's equip spell states with A_MOD_ARMOR_NO_MODS.
func noModsArmorFixtures(t *testing.T) float64 {
	t.Helper()
	armor := spelldata.MustFind(arcaneAmalgamationArmor).FirstAura(dbcenums.A_MOD_ARMOR_NO_MODS).BasePoints
	needAmount(t, "15390's armor", armor)

	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{
			{Id: testChestID, Name: "Test Chest", Type: proto.ItemType_ItemTypeChest,
				ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}},
			{Id: armorScalingNeckID, Name: "Test Armor Scaling Neck", Type: proto.ItemType_ItemTypeNeck,
				ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}},
		},
		Enchants: []*proto.SimEnchant{
			{EffectId: noModsArmorEnchantID, Name: "Test Armor No Mods", Type: proto.ItemType_ItemTypeChest,
				Stats: stats.Stats{stats.ArmorNoMods: armor}.ToProtoArray()},
			{EffectId: plainArmorEnchantID, Name: "Test Armor", Type: proto.ItemType_ItemTypeChest,
				Stats: stats.Stats{stats.Armor: armor}.ToProtoArray()},
		},
	})
	if !core.HasItemEffect(armorScalingNeckID) {
		core.NewItemEffect(armorScalingNeckID, func(agent core.Agent) {
			character := agent.GetCharacter()
			character.ApplyEquipScaling(stats.Armor, armorScaling)
			character.MultiplyStat(stats.Armor, armorScaling)
		})
	}
	return armor
}

// A wearer of the scaling neck and the test chest with the enchant.
func chestWearer(enchantID int32) *proto.Player {
	player := shaman("Wearer", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotChest: testChestID, proto.ItemSlot_ItemSlotNeck: armorScalingNeckID})
	player.Equipment.Items[proto.ItemSlot_ItemSlotChest].Enchant = enchantID
	return player
}

// An enchant's ArmorNoMods, the stat the generator writes for an equip spell's A_MOD_ARMOR_NO_MODS
// armor, reaches the wearer whole past both the scaling of armor from items and a percent armor
// modifier, while the same amount of Armor is scaled by both.
func TestNoModsArmorFromAnEnchantStaysOutOfEveryArmorScaling(t *testing.T) {
	armor := noModsArmorFixtures(t)
	armorWith := func(enchantID int32) float64 {
		return auraTesterAt(newAuraSim(chestWearer(enchantID)), 0).GetStat(stats.Armor)
	}
	unenchanted := armorWith(0)
	if got := armorWith(noModsArmorEnchantID) - unenchanted; !core.WithinToleranceFloat64(armor, got, 1e-6) {
		t.Errorf("the ArmorNoMods enchant added %v armor, want its %v unscaled", got, armor)
	}
	if got, want := armorWith(plainArmorEnchantID)-unenchanted, armor*armorScaling*armorScaling; !core.WithinToleranceFloat64(want, got, 1e-6) {
		t.Errorf("the plain armor enchant added %v armor, want its %v scaled to %v", got, armor, want)
	}
}

// A pre-pull item swap that takes the chest with the ArmorNoMods enchant off takes that armor off
// whole, and the swap back puts it on whole, with the neck's scaling up throughout.
func TestItemSwapMovesArmorNoModsWhole(t *testing.T) {
	armor := noModsArmorFixtures(t)
	swapTo := func(doAt string, set proto.APLActionItemSwap_SwapSet) *proto.APLPrepullAction {
		return &proto.APLPrepullAction{
			DoAtValue: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: doAt}}},
			Action:    &proto.APLAction{Action: &proto.APLAction_ItemSwap{ItemSwap: &proto.APLActionItemSwap{SwapSet: set}}},
		}
	}
	player := chestWearer(noModsArmorEnchantID)
	swapItems := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range swapItems {
		swapItems[i] = &proto.ItemSpec{}
	}
	swapItems[proto.ItemSlot_ItemSlotChest] = &proto.ItemSpec{Id: testChestID}
	player.EnableItemSwap = true
	player.ItemSwap = &proto.ItemSwap{Items: swapItems}
	player.Rotation = &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PrepullActions: []*proto.APLPrepullAction{
		swapTo("-2s", proto.APLActionItemSwap_Swap1),
		swapTo("-1s", proto.APLActionItemSwap_Main),
	}}

	sim := newAuraSim(player)
	wearer := auraTesterAt(sim, 0)
	enchanted := wearer.GetStat(stats.Armor)
	sim.PrePull()
	stepPast(t, sim, -1500*time.Millisecond)
	if got := enchanted - wearer.GetStat(stats.Armor); !core.WithinToleranceFloat64(armor, got, 1e-6) {
		t.Errorf("swapping the chest off took %v armor, want the enchant's %v", got, armor)
	}
	stepPast(t, sim, -500*time.Millisecond)
	if got := wearer.GetStat(stats.Armor); !core.WithinToleranceFloat64(enchanted, got, 1e-6) {
		t.Errorf("after the swap back the wearer has %v armor, want the %v it started with", got, enchanted)
	}
}
