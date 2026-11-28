//go:build with_db

package forever_test

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/core/testagent"
)

func init() {
	forever.RegisterAllProcs()
	forever.RegisterAllEnchants()
	forever.RegisterAllOnUseCds()
	testagent.Register(proto.Player_HolyPaladin{}, proto.Spec_SpecHolyPaladin, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_HolyPaladin)
	}, func(agent *testagent.Agent) {
		agent.EnableManaBar()
	})
}

// A character wearing one item, against a level 60 target: the warrior stand-in, with a paladin the
// stand-in registered as the holy paladin, which has a mana bar, with a rogue the stand-in
// registered as the rogue, which shoots its ranged weapon, and with a druid or a shaman the healer
// stand-ins of items_relics_test.go.
func itemWearer(t *testing.T, class proto.Class, slot proto.ItemSlot, itemID int32) (*core.Simulation, *core.Character) {
	t.Helper()
	return itemSpecWearer(t, class, slot, &proto.ItemSpec{Id: itemID})
}

func itemSpecWearer(t *testing.T, class proto.Class, slot proto.ItemSlot, item *proto.ItemSpec) (*core.Simulation, *core.Character) {
	t.Helper()
	return itemWearerWithEquipment(t, class, equipping(slot, item))
}

// Equipment of one item in the slot and nothing else.
func equipping(slot proto.ItemSlot, item *proto.ItemSpec) *proto.EquipmentSpec {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[slot] = item
	return &proto.EquipmentSpec{Items: items}
}

func itemWearerWithEquipment(t *testing.T, class proto.Class, equipment *proto.EquipmentSpec) (*core.Simulation, *core.Character) {
	t.Helper()
	return itemWearerFacing(t, class, equipment, proto.MobType_MobTypeUnknown)
}

func itemWearerFacing(t *testing.T, class proto.Class, equipment *proto.EquipmentSpec, mobType proto.MobType) (*core.Simulation, *core.Character) {
	t.Helper()
	return itemWearerAgainst(t, class, equipment, &proto.Target{Name: "target", Level: 60, MobType: mobType})
}

func itemWearerAgainst(t *testing.T, class proto.Class, equipment *proto.EquipmentSpec, targets ...*proto.Target) (*core.Simulation, *core.Character) {
	t.Helper()
	return itemWearerInEncounter(t, class, equipment, &proto.Encounter{Targets: targets, Duration: 180})
}

// The same against a level 60 target, in an encounter in the areas.
func itemWearerIn(t *testing.T, class proto.Class, equipment *proto.EquipmentSpec, areas ...proto.AreaType) (*core.Simulation, *core.Character) {
	t.Helper()
	return itemWearerInEncounter(t, class, equipment,
		&proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180, AreaTypes: areas})
}

func itemWearerInEncounter(t *testing.T, class proto.Class, equipment *proto.EquipmentSpec, encounter *proto.Encounter, playerOpts ...func(*proto.Player)) (*core.Simulation, *core.Character) {
	t.Helper()
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceHuman,
		Equipment:   equipment,
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	for _, opt := range playerOpts {
		opt(player)
	}
	switch class {
	case proto.Class_ClassPaladin:
		core.WithSpec(player, &proto.Player_HolyPaladin{HolyPaladin: &proto.HolyPaladin{}})
	case proto.Class_ClassRogue:
		core.WithSpec(player, &proto.Player_Rogue{Rogue: &proto.Rogue{}})
	case proto.Class_ClassDruid:
		core.WithSpec(player, &proto.Player_RestorationDruid{RestorationDruid: &proto.RestorationDruid{}})
	case proto.Class_ClassShaman:
		core.WithSpec(player, &proto.Player_RestorationShaman{RestorationShaman: &proto.RestorationShaman{}})
	default:
		core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  encounter,
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Whether every cast landed within the effect's roll. The caller rules out crits.
func damageInRange(metrics core.SpellMetrics, effect *spelldata.Effect) bool {
	casts := float64(metrics.Casts)
	return metrics.Crits == 0 && metrics.TotalDamage >= casts*effect.Min() && metrics.TotalDamage <= casts*effect.Max()
}

// Zealot Plate 274651's 1294668 restores 1294666's mana when a melee swing taken lands, and nothing on
// a spell taken. It then rests for the row's cooldown.
func TestZealotPlateRestoresManaOnMeleeHitsTakenAndRests(t *testing.T) {
	editSpellRow(t, 1294668, func(s *spelldata.Spell) {
		s.ProcChance, s.ProcChanceSource = 100, spelldata.ProcChanceColumn
	})
	gain := spelldata.MustFind(1294666).EnergizeEffect()
	cooldown := spelldata.MustFind(1294668).ICD()
	if gain.Min() <= 0 || cooldown <= 0 {
		t.Fatalf("1294666 restores %v to %v mana and 1294668 rests %v, want a gain and a cooldown", gain.Min(), gain.Max(), cooldown)
	}
	sim, character := itemWearer(t, proto.Class_ClassPaladin, proto.ItemSlot_ItemSlotChest, 274651)
	enemy := sim.Encounter.ActiveTargetUnits[0]
	character.SpendMana(sim, 4*gain.Max(), character.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen}))
	restored := func(t *testing.T, spell *core.Spell) float64 {
		t.Helper()
		before := character.CurrentMana()
		takeHit(t, sim, character, spell)
		return character.CurrentMana() - before
	}

	if got := restored(t, fireSpell(enemy)); got != 0 {
		t.Errorf("a spell hit taken restored %v mana, want none", got)
	}
	start := sim.CurrentTime
	if got := restored(t, meleeSwing(enemy)); got < gain.Min() || got > gain.Max() {
		t.Errorf("a melee hit taken restored %v mana, want the row's %v to %v", got, gain.Min(), gain.Max())
	}
	if got := restored(t, meleeSwing(enemy)); got != 0 {
		t.Errorf("a melee hit taken inside the cooldown restored %v mana, want none", got)
	}
	// stepTo stops after 10,000 steps, fewer than the sim takes to cross the whole cooldown, so the
	// cooldown is crossed in 10 s legs.
	for at := start + 10*time.Second; at < start+cooldown; at += 10 * time.Second {
		stepTo(t, sim, at)
	}
	stepTo(t, sim, start+cooldown)
	if got := restored(t, meleeSwing(enemy)); got < gain.Min() {
		t.Errorf("a melee hit taken as the cooldown ends restored %v mana, want at least %v", got, gain.Min())
	}
}

// Argent Defender 13243's 17350 rolls the row's procs per minute on each landed melee hit taken,
// priced on the attacker's time between its own swings, a dual wielder's two hands together, and
// doubled by the row's second effect when an Undead strikes, and states that chance against its
// current target. Each case counts the procs of many hits
// on a row copy that rests no cooldown and carries a higher rate, so the cases sit far apart, within
// five standard deviations of the count.
func TestArgentDefenderPricesTheAttackersSwings(t *testing.T) {
	const ppm = 6
	row := editSpellRow(t, 17350, func(s *spelldata.Spell) { s.ICDMs, s.RPPM = 0, ppm })
	undead := row.EffectN(2).BasePoints
	if undead <= 1 {
		t.Fatalf("17350's second effect is %v, want a factor above 1 to tell Undead apart", undead)
	}

	for _, attacker := range []struct {
		name     string
		target   *proto.Target
		interval float64 // Seconds between the attacker's own swings.
		factor   float64
	}{
		{"1.0 s", &proto.Target{SwingSpeed: 1, MobType: proto.MobType_MobTypeHumanoid}, 1, 1},
		{"3.0 s", &proto.Target{SwingSpeed: 3, MobType: proto.MobType_MobTypeHumanoid}, 3, 1},
		{"3.0 s Undead", &proto.Target{SwingSpeed: 3, MobType: proto.MobType_MobTypeUndead}, 3, undead},
		{"2.0 s dual wield", &proto.Target{SwingSpeed: 2, DualWield: true, MobType: proto.MobType_MobTypeHumanoid}, 1, 1},
	} {
		t.Run(attacker.name, func(t *testing.T) {
			want := ppm * attacker.interval / 60 * attacker.factor
			attacker.target.Name, attacker.target.Level = "target", 60
			sim, character, buff := argentDefenderWearer(t, attacker.target)
			enemy := sim.Encounter.ActiveTargetUnits[0]
			if got := enemy.AutoAttacks.MainhandSwingSpeed().Seconds(); got != attacker.target.SwingSpeed {
				t.Fatalf("the attacker swings every %v s, want the target's %v s", got, attacker.target.SwingSpeed)
			}
			trigger := character.GetAura("Argent Defender")
			if trigger == nil || trigger.Dpm == nil {
				t.Fatal("13243 registers no proc on a procs per minute rate")
			}
			if got := trigger.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim); math.Abs(got-want) > 1e-9 {
				t.Errorf("the stated chance against the current target is %v, want %v", got, want)
			}
			const hits = 4000
			procs := 0
			for range hits {
				takeHit(t, sim, character, meleeSwing(enemy))
				if buff.IsActive() {
					procs++
					buff.Deactivate(sim)
				}
			}

			if spread := 5 * math.Sqrt(hits*want*(1-want)); math.Abs(float64(procs)-hits*want) > spread {
				t.Errorf("%d of %d landed melee hits taken procced it, want %v within %v", procs, hits, hits*want, spread)
			}
		})
	}
}

// Argent Defender 13243 raises the block chance by 17351's amount when a melee hit taken lands and
// deals damage, and nothing on a dodge, a hit the block took whole, a spell taken or a hit from an
// attacker with no melee auto attack. It then rests for 17350's cooldown. The row copy's rate procs
// on every hit from the 2 s attacker.
func TestArgentDefenderRaisesBlockAndRests(t *testing.T) {
	editSpellRow(t, 17350, func(s *spelldata.Spell) { s.RPPM = 60 })
	block := float64(spelldata.MustFind(17351).EffectN(1).BasePoints)
	cooldown := spelldata.MustFind(17350).ICD()
	if block <= 0 || cooldown <= 0 {
		t.Fatalf("17351 raises the block chance by %v and 17350 rests %v, want both", block, cooldown)
	}
	sim, character, buff := argentDefenderWearer(t,
		&proto.Target{Name: "swinging", Level: 60, SwingSpeed: 2},
		&proto.Target{Name: "still", Level: 60})
	enemy, still := sim.Encounter.ActiveTargetUnits[0], sim.Encounter.ActiveTargetUnits[1]

	dodgeSwing(t, sim, character, enemy)
	if buff.IsActive() {
		t.Error("a dodged melee swing raised the block chance, want nothing")
	}
	character.OnSpellHitTaken(sim, meleeSwing(enemy), &core.SpellResult{Target: &character.Unit, Outcome: core.OutcomeHit | core.OutcomeBlock})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if buff.IsActive() {
		t.Error("a melee hit the block took whole raised the block chance, want nothing")
	}
	takeHit(t, sim, character, fireSpell(enemy))
	if buff.IsActive() {
		t.Error("a spell hit taken raised the block chance, want nothing")
	}
	takeHit(t, sim, character, meleeSwing(still))
	if buff.IsActive() {
		t.Error("a melee hit from an attacker with no auto attack raised the block chance, want nothing")
	}
	start := sim.CurrentTime
	before := character.GetStats()
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := gained(character, before, stats.BlockPercent); !buff.IsActive() || !statNear(got, block) {
		t.Errorf("a melee hit taken raised the block chance by %v, want the row's %v", got, block)
	}
	buff.Deactivate(sim)
	takeHit(t, sim, character, meleeSwing(enemy))
	if buff.IsActive() {
		t.Error("a melee hit taken inside the cooldown raised the block chance, want nothing")
	}
	// stepTo stops after 10,000 steps, fewer than the sim takes to cross the whole cooldown, so the
	// cooldown is crossed in 10 s legs.
	for at := start + 10*time.Second; at < start+cooldown; at += 10 * time.Second {
		stepTo(t, sim, at)
	}
	stepTo(t, sim, start+cooldown)
	takeHit(t, sim, character, meleeSwing(enemy))
	if !buff.IsActive() {
		t.Error("a melee hit taken as the cooldown ends raised no block chance")
	}
}

// A warrior wearing Argent Defender 13243 against the targets, none of them tanked, so no target
// swings on its own, and the 17351 buff the shield applies.
func argentDefenderWearer(t *testing.T, targets ...*proto.Target) (*core.Simulation, *core.Character, *core.Aura) {
	t.Helper()
	sim, character := itemWearerAgainst(t, proto.Class_ClassWarrior, equipping(proto.ItemSlot_ItemSlotOffHand, &proto.ItemSpec{Id: 13243}), targets...)
	buff := character.GetAuraByID(core.ActionID{SpellID: 17351})
	if buff == nil {
		t.Fatal("13243 registers no 17351 buff")
	}
	return sim, character, buff
}

// Black Grasp of the Destroyer 22194 restores 29471's mana on each landed melee or ranged hit, and
// nothing on a spell; a wearer without mana registers nothing.
func TestBlackGraspOfTheDestroyerRestoresManaPerHit(t *testing.T) {
	gain := spelldata.MustFind(29471).EnergizeEffect().Average()
	sim, character := itemWearer(t, proto.Class_ClassPaladin, proto.ItemSlot_ItemSlotHands, 22194)
	character.SpendMana(sim, 2*gain, character.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen}))
	before := character.CurrentMana()

	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if got := character.CurrentMana() - before; gain <= 0 || got != gain {
		t.Errorf("a melee hit restored %v mana, want the row's %v", got, gain)
	}
	dealHit(t, sim, character, fireSpell(&character.Unit))
	if got := character.CurrentMana() - before; got != gain {
		t.Errorf("after a spell hit the mana restored is %v, want %v still", got, gain)
	}

	_, warrior := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotHands, 22194)
	if warrior.GetAura("Black Grasp of the Destroyer") != nil {
		t.Errorf("a wearer without mana registers the proc")
	}
}

// Swine Fists 10760's 1293783 casts 1293782's Nature damage on a hit against a Swine, and nothing on
// a hit against a creature of another family. The combat log names the item on the damage.
func TestSwineFistsDealDamageOnlyToSwine(t *testing.T) {
	if spelldata.MustFind(1293782).DamageEffect().Average() <= 0 {
		t.Fatal("1293782 states no damage")
	}
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotHands].Id = 10760
	player := &proto.Player{
		Name: "Wearer", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceHuman,
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter: &proto.Encounter{Duration: 180, Targets: []*proto.Target{
			{Name: "swine", Level: 60, MobType: proto.MobType_MobTypeBeast, CreatureFamily: proto.CreatureFamily_CreatureFamilySwine},
			{Name: "aquatic", Level: 60, MobType: proto.MobType_MobTypeBeast, CreatureFamily: proto.CreatureFamily_CreatureFamilyAquatic},
		}},
	}, simsignals.CreateSignals())
	sim.Reset()
	character := sim.Raid.Parties[0].Players[0].GetCharacter()
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	swine, aquatic := sim.Encounter.ActiveTargetUnits[0], sim.Encounter.ActiveTargetUnits[1]
	playerStats := &proto.PlayerStats{}
	character.FillPlayerStats(playerStats)
	if sheet := playerStats.VersusEffects; len(sheet) != 1 || sheet[0].ItemId != 10760 || sheet[0].Kind != proto.VersusEffectKind_VersusEffectKindProcOnly ||
		!slices.Equal(sheet[0].CreatureFamilies, []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilySwine}) {
		t.Errorf("the stats sheet lists %v, want the proc against Swine only", sheet)
	}
	damage := character.GetSpell(core.ActionID{SpellID: 1293782})
	if damage == nil {
		t.Fatal("10760 registers no damage spell")
	}

	lines := damageLines(sim, func() {
		for _, target := range []*core.Unit{swine, aquatic} {
			character.OnSpellHitDealt(sim, meleeSwing(&character.Unit), &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
			stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
		}
	})
	if want := ") [Versus Beast (Swine): {ItemID: 10760} proc only]"; len(lines) != 1 || !strings.HasSuffix(lines[0], want) {
		t.Errorf("the hits log %q, want one damage line ending with %q", lines, want)
	}
	if got := damage.SpellMetrics[swine.UnitIndex]; got.Casts != 1 || got.TotalDamage <= 0 {
		t.Errorf("a hit on the Swine cast the damage %d times for %v, want once for some damage", got.Casts, got.TotalDamage)
	}
	if got := damage.SpellMetrics[aquatic.UnitIndex]; got.Casts != 0 || got.TotalDamage != 0 {
		t.Errorf("a hit on the Aquatic target cast the damage %d times for %v, want none", got.Casts, got.TotalDamage)
	}
}

// Gloves of Holy Might 867's 1330855 casts 1325229's Holy damage on a hit, tripled against an Undead
// target. Both wearers roll from the same seed, so the target's creature type is all that differs.
func TestGlovesOfHolyMightTripleDamageAgainstUndead(t *testing.T) {
	smite := func(mobType proto.MobType) core.SpellMetrics {
		sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotHands, 867)
		character.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: 100, stats.SpellHitPercent: 100})
		target := character.CurrentTarget
		target.MobType = mobType
		damage := character.GetSpell(core.ActionID{SpellID: 1325229})
		if damage == nil {
			t.Fatal("867 registers no damage spell")
		}
		for hits := 0; damage.SpellMetrics[target.UnitIndex].Casts == 0; hits++ {
			if hits > 200 {
				t.Fatal("200 hits cast no damage")
			}
			dealHit(t, sim, character, meleeSwing(&character.Unit))
		}
		return damage.SpellMetrics[target.UnitIndex]
	}

	undead, humanoid := smite(proto.MobType_MobTypeUndead), smite(proto.MobType_MobTypeHumanoid)
	if humanoid.TotalDamage <= 0 {
		t.Fatalf("the proc dealt %v to a Humanoid, want some damage", humanoid.TotalDamage)
	}
	if math.Abs(undead.TotalDamage-3*humanoid.TotalDamage) > 1e-6 {
		t.Errorf("the proc dealt %v to an Undead target, want 3 times the %v it dealt to a Humanoid", undead.TotalDamage, humanoid.TotalDamage)
	}
}

// Thick Obsidian Breastplate 22196's 27539 has a chance, on a spell hit taken, to shield the wearer
// against that spell's school with the Fire row 27533's absorb roll for its duration. The chance has
// 27539's internal cooldown, and a Physical spell, which has no shield, does not spend it.
func TestThickObsidianBreastplateShieldsTheSchoolHit(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotChest, 22196)
	enemy := sim.Encounter.ActiveTargetUnits[0]
	fireShield := character.GetAuraByID(core.ActionID{SpellID: 27533})
	frostShield := character.GetAuraByID(core.ActionID{SpellID: 27534})
	if fireShield == nil || frostShield == nil {
		t.Fatal("22196 registers no Fire or Frost shield")
	}
	fire := spelldata.MustFind(27533)
	absorb := fire.AbsorbEffect()
	icd := spelldata.MustFind(27539).ICD()
	if icd <= fire.Duration() {
		t.Fatalf("27539's cooldown %v does not outlast the shield's %v, which the steps below need", icd, fire.Duration())
	}
	spell := func(school core.SpellSchool) *core.Spell {
		return &core.Spell{ProcMask: core.ProcMaskSpellDamage, SpellSchool: school, DefenseType: core.DefenseTypeMagic, Unit: enemy}
	}
	shieldsUp := func() int {
		up := 0
		for _, aura := range character.GetAuras() {
			if aura.IsActive() && aura.ActionID.SpellID >= 27533 && aura.ActionID.SpellID <= 27540 {
				up++
			}
		}
		return up
	}
	hitUntil := func(school core.SpellSchool, shield *core.Aura) {
		t.Helper()
		for range 200 {
			takeHit(t, sim, character, spell(school))
			if shield.IsActive() {
				return
			}
		}
		t.Fatalf("200 %v spells taken raised no %v shield", school, school)
	}

	for range 100 {
		takeHit(t, sim, character, spell(core.SpellSchoolPhysical))
		takeHit(t, sim, character, meleeSwing(enemy))
	}
	if got := shieldsUp(); got != 0 {
		t.Fatalf("Physical spells and melee hits raised %d shields, want none", got)
	}

	hitUntil(core.SpellSchoolFire, fireShield)
	raised := sim.CurrentTime
	if got := float64(fireShield.GetStacks()); got < absorb.Min() || got > absorb.Max() {
		t.Errorf("the Fire shield absorbs %v, want the row's %v to %v", got, absorb.Min(), absorb.Max())
	}
	if got := fireShield.Duration; got != fire.Duration() {
		t.Errorf("the Fire shield lasts %v, want the row's %v", got, fire.Duration())
	}
	stepTo(t, sim, raised+fire.Duration()+time.Millisecond)
	if fireShield.IsActive() {
		t.Errorf("the Fire shield is still up after the row's %v", fire.Duration())
	}

	for sim.CurrentTime < raised+icd-core.SpellBatchWindow {
		takeHit(t, sim, character, spell(core.SpellSchoolFrost))
	}
	if frostShield.IsActive() {
		t.Errorf("a Frost shield went up within the row's %v cooldown", icd)
	}

	stepTo(t, sim, raised+icd)
	hitUntil(core.SpellSchoolFrost, frostShield)
	if got := shieldsUp(); got != 1 {
		t.Errorf("%d shields are up after a Frost spell, want the Frost one alone", got)
	}
}

// Demon Forged Breastplate 12628's 16611 has a chance, on a melee hit taken, to cast 16608 on the
// attacker: the row's leech every period for its duration, each tick healing the wearer for the
// effect's share of the damage it dealt, times the wearer's healing taken multiplier.
func TestDemonForgedBreastplateDrainsTheAttacker(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotChest, 12628)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	character.PseudoStats.HealingTakenMultiplier = 0.5
	enemy := sim.Encounter.ActiveTargetUnits[0]
	drain := character.GetSpell(core.ActionID{SpellID: 16608})
	if drain == nil {
		t.Fatal("12628 registers no drain")
	}
	row := spelldata.MustFind(16608)
	leech := row.FirstAura(dbcenums.A_PERIODIC_LEECH)
	ticks := rowTickCount(row, leech)
	character.RemoveHealth(sim, 2*float64(ticks)*leech.Average())

	hits := 0
	for ; !drain.Dot(enemy).IsActive() && hits < 1000; hits++ {
		takeHit(t, sim, character, meleeSwing(enemy))
	}
	if hits == 1000 {
		t.Fatalf("1000 melee hits taken cast no drain")
	}
	dot := drain.Dot(enemy)
	if dot.BaseTickCount != ticks || dot.BaseTickLength != leech.Period() {
		t.Errorf("the drain ticks %d times every %v, want the row's %d every %v", dot.BaseTickCount, dot.BaseTickLength, ticks, leech.Period())
	}

	health := character.CurrentHealth()
	stepTo(t, sim, sim.CurrentTime+row.Duration()+time.Millisecond)
	metrics := drain.SpellMetrics[enemy.UnitIndex]
	if want := float64(ticks) * leech.Average(); want <= 0 || metrics.TotalDamage != want {
		t.Errorf("the drain dealt %v, want the row's %v", metrics.TotalDamage, want)
	}
	if got, want := character.CurrentHealth()-health, metrics.TotalDamage*float64(leech.Amplitude)*0.5; got != want {
		t.Errorf("the drain healed the wearer for %v, want the effect's share of the %v it dealt at half healing taken, %v", got, metrics.TotalDamage, want)
	}
}

// Maw of Min'loth 274608's 1295064 has a chance, on a melee hit taken, to cast 1295065 on the attacker,
// read from the generated registrations: the row's E_HEALTH_LEECH amount on the melee table, reduced
// by the attacker's armor, and the wearer gains the effect's Amplitude share of the damage dealt. A
// cast that misses, is dodged or is parried deals nothing and heals nothing.
func TestMawOfMinlothLeechesTheAttacker(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotOffHand, 274608)
	character.AddStatsDynamic(sim, stats.Stats{stats.PhysicalCritPercent: -character.GetStat(stats.PhysicalCritPercent)})
	character.PseudoStats.InFrontOfTarget = true
	enemy := sim.Encounter.ActiveTargetUnits[0]
	enemy.AddStatsDynamic(sim, stats.Stats{stats.Armor: 3000})
	bite := character.GetSpell(core.ActionID{SpellID: 1295065})
	if bite == nil {
		t.Fatal("274608 registers no Fatal Bite")
	}
	leech := spelldata.MustFind(1295065).ProcDamageEffect()
	if leech.Type != dbcenums.E_HEALTH_LEECH || leech.Average() <= 0 {
		t.Fatalf("1295065 deals through %v for %v, want an E_HEALTH_LEECH amount", leech.Type, leech.Average())
	}
	armor := character.AttackTables[enemy.UnitIndex].GetArmorDamageModifier(bite)
	if armor >= 1 {
		t.Fatalf("the target's armor keeps %v of the damage, want a reduction to test against", armor)
	}
	icd := spelldata.MustFind(1295064).ICD()

	// The hit takeHit deals removes health, so every hit starts at half health and a hit that casts
	// nothing shows what it costs; a cast's heal is what the wearer ends above that.
	struck := -1.0
	landed, avoided := 0, 0
	for hits := 0; hits < 5000 && (landed == 0 || avoided == 0); hits++ {
		character.SetCurrentHealthPercent(0.5)
		before := character.CurrentHealth()
		metrics := bite.SpellMetrics[enemy.UnitIndex]
		takeHit(t, sim, character, meleeSwing(enemy))
		after := bite.SpellMetrics[enemy.UnitIndex]
		if after.Casts == metrics.Casts {
			struck = before - character.CurrentHealth()
			continue
		}
		if struck < 0 {
			continue
		}
		dealt, healed := after.TotalDamage-metrics.TotalDamage, character.CurrentHealth()-before+struck
		stepTo(t, sim, sim.CurrentTime+icd)
		if after.Hits == metrics.Hits {
			avoided++
			if dealt != 0 || healed != 0 {
				t.Errorf("a cast that did not land dealt %v and healed the wearer for %v, want nothing", dealt, healed)
			}
			continue
		}
		landed++
		if want := leech.Average() * armor; !statNear(dealt, want) {
			t.Errorf("a landed cast dealt %v, want the row's %v reduced by armor to %v", dealt, leech.Average(), want)
		}
		if want := dealt * float64(leech.Amplitude); !statNear(healed, want) {
			t.Errorf("a landed cast healed the wearer for %v, want the effect's share of the %v it dealt, %v", healed, dealt, want)
		}
	}
	if landed == 0 || avoided == 0 {
		t.Fatalf("5000 melee hits taken landed %d casts and had %d avoided, want some of each", landed, avoided)
	}
}

func TestArathiLeatherBonusesActivateAtTheirPieceCounts(t *testing.T) {
	for _, set := range []struct {
		name string
		ids  []int32
	}{{"The Defiler's Purpose", []int32{20194, 20190, 20186}},
		{"The Highlander's Purpose", []int32{20059, 20045, 20052}}} {
		for count := 1; count <= 3; count++ {
			items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
			for i := range items {
				items[i] = &proto.ItemSpec{}
			}
			for i, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet} {
				if i < count {
					items[slot].Id = set.ids[i]
				}
			}
			sim, wearer := itemWearerWithEquipment(t, proto.Class_ClassWarrior, &proto.EquipmentSpec{Items: items})
			for _, bonus := range []struct {
				count int
				stat  stats.Stat
				row   int32
			}{{2, stats.Stamina, 7503}, {3, stats.PhysicalCritPercent, 7597}} {
				aura := wearer.GetAura(fmt.Sprintf("%s %dP", set.name, bonus.count))
				if count < bonus.count {
					if aura != nil && aura.IsActive() {
						t.Fatal("set bonus activated without enough pieces")
					}
					continue
				}
				if aura == nil || !aura.IsActive() {
					t.Fatal("equipped set bonus is inactive")
				}
				before := wearer.GetStat(bonus.stat)
				aura.Deactivate(sim)
				want := spelldata.MustFind(bonus.row).EffectN(1).BaseValue()
				if got := before - wearer.GetStat(bonus.stat); math.Abs(got-want) > 1e-8 {
					t.Errorf("%s bonus changed stat by %g, want %g", set.name, got, want)
				}
			}
		}
	}
}

// A row that deals its damage off its own trigger, read from the generated registrations: the
// A_PROC_TRIGGER_DAMAGE roll lands on the attacker of every melee hit the wearer blocks, whatever
// RecoveryTime the row states, and a hit not blocked deals nothing. Girdle of Extracted Answers
// 276742 carries 1308777 on an item and Thorium Shield Spike 1704 carries 16624 on an enchant, both
// Physical; Aegis of the Scarlet Commander 7726's 1293539 is Holy and rolls on the spell table.
func TestBlockedHitsDealTheRowsDamageToTheAttacker(t *testing.T) {
	for _, tc := range []struct {
		name    string
		equip   *proto.ItemSpec
		slot    proto.ItemSlot
		spellID int32
	}{
		{"Girdle of Extracted Answers", &proto.ItemSpec{Id: 276742}, proto.ItemSlot_ItemSlotWaist, 1308777},
		{"Thorium Shield Spike on Shard of the Splithooves", &proto.ItemSpec{Id: 10659, Enchant: 1704}, proto.ItemSlot_ItemSlotOffHand, 16624},
		{"Aegis of the Scarlet Commander", &proto.ItemSpec{Id: 7726}, proto.ItemSlot_ItemSlotOffHand, 1293539},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, character := itemSpecWearer(t, proto.Class_ClassWarrior, tc.slot, tc.equip)
			character.AddStatsDynamic(sim, stats.Stats{
				stats.PhysicalHitPercent: 100, stats.PhysicalCritPercent: -character.GetStat(stats.PhysicalCritPercent),
				stats.SpellHitPercent: 100, stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent),
			})
			enemy := sim.Encounter.ActiveTargetUnits[0]
			spike := character.GetSpell(core.ActionID{SpellID: tc.spellID})
			if spike == nil {
				t.Fatalf("%s registers no damage spell", tc.name)
			}
			row := spelldata.MustFind(tc.spellID)
			damage := row.FirstAura(dbcenums.A_PROC_TRIGGER_DAMAGE)
			if row.Cooldown() == 0 {
				t.Fatalf("%d states no RecoveryTime, so the check for no internal cooldown proves nothing", tc.spellID)
			}

			takeHit(t, sim, character, meleeSwing(enemy))
			if got := spike.SpellMetrics[enemy.UnitIndex].Casts; got != 0 {
				t.Errorf("a hit not blocked cast the damage %d times, want none", got)
			}

			block := func() {
				character.OnSpellHitTaken(sim, meleeSwing(enemy), &core.SpellResult{Target: &character.Unit, Outcome: core.OutcomeHit | core.OutcomeBlock, Damage: 50})
				stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
			}
			block()
			metrics := spike.SpellMetrics[enemy.UnitIndex]
			if metrics.Casts != 1 || !damageInRange(metrics, damage) {
				t.Errorf("a block cast the damage %d times for %v, want once for %v to %v", metrics.Casts, metrics.TotalDamage, damage.Min(), damage.Max())
			}
			block()
			if got := spike.SpellMetrics[enemy.UnitIndex].Casts; got != 2 {
				t.Errorf("a second block right after cast the damage %d times in all, want 2", got)
			}
		})
	}
}

// Stoneform Shoulders 13955 put up 1298231's flat Physical damage taken at the encounter start, and the
// first Physical hit taken that deals damage spends it: a spell hit and a dodged attack leave it up.
func TestStoneformShouldersReduceOnlyTheNextPhysicalHit(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotShoulder, 13955)
	reduction := spelldata.MustFind(1298231).EffectN(1).BasePoints
	before := character.PseudoStats.BonusPhysicalDamageTaken

	character.OnEncounterStart(sim)
	if got := character.PseudoStats.BonusPhysicalDamageTaken - before; reduction >= 0 || got != reduction {
		t.Fatalf("the encounter start changed the flat Physical damage taken by %v, want the row's %v", got, reduction)
	}

	enemy := sim.Encounter.ActiveTargetUnits[0]
	takeHit(t, sim, character, fireSpell(enemy))
	dodgeSwing(t, sim, character, enemy)
	if got := character.PseudoStats.BonusPhysicalDamageTaken - before; got != reduction {
		t.Fatalf("after a spell hit and a dodge the flat Physical damage taken is %v off, want %v still", got, reduction)
	}

	takeHit(t, sim, character, meleeSwing(enemy))
	if got := character.PseudoStats.BonusPhysicalDamageTaken - before; got != 0 {
		t.Errorf("after a Physical hit taken the flat Physical damage taken is %v off, want 0", got)
	}
}

// Skullflame Shield 1168 rolls both its procs on melee hits taken. 18815's Drain Life 18817 hits the
// attacker only while the wearer is in front of its target, and 18816's Flamestrike 18818 rolls either
// way and deals the row's damage to every enemy.
func TestSkullflameShieldDrainsOnlyInFrontAndBurnsEveryEnemy(t *testing.T) {
	targets := []*proto.Target{{Name: "first", Level: 60}, {Name: "second", Level: 60}, {Name: "third", Level: 60}}
	sim, character := itemWearerAgainst(t, proto.Class_ClassWarrior, equipping(proto.ItemSlot_ItemSlotOffHand, &proto.ItemSpec{Id: 1168}), targets...)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100, stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	attacker := sim.Encounter.ActiveTargetUnits[0]
	drainLife := character.GetSpell(core.ActionID{SpellID: 18817})
	flamestrike := character.GetSpell(core.ActionID{SpellID: 18818})
	if drainLife == nil || flamestrike == nil {
		t.Fatalf("1168 registers Drain Life %v and Flamestrike %v, want both", drainLife != nil, flamestrike != nil)
	}

	takeHits := func() {
		for range 2000 {
			character.SetCurrentHealthPercent(0.5)
			takeHit(t, sim, character, meleeSwing(attacker))
		}
	}

	character.PseudoStats.InFrontOfTarget = false
	takeHits()
	if got := drainLife.SpellMetrics[attacker.UnitIndex].Casts; got != 0 {
		t.Errorf("behind its target the wearer cast Drain Life %d times, want none", got)
	}
	casts := flamestrike.SpellMetrics[attacker.UnitIndex].Casts
	if casts == 0 {
		t.Fatal("behind its target the wearer cast no Flamestrike")
	}
	flame := spelldata.MustFind(18818).DamageEffect()
	for _, enemy := range sim.Encounter.ActiveTargetUnits {
		metrics := flamestrike.SpellMetrics[enemy.UnitIndex]
		hits := float64(metrics.Hits)
		if metrics.Hits != casts || metrics.TotalDamage < hits*flame.Min() || metrics.TotalDamage > hits*flame.Max() {
			t.Errorf("%s took %d Flamestrike hits for %v over %d casts, want one per cast within the row's roll", enemy.Label, metrics.Hits, metrics.TotalDamage, casts)
		}
	}

	character.PseudoStats.InFrontOfTarget = true
	takeHits()
	if got := drainLife.SpellMetrics[attacker.UnitIndex].Casts; got == 0 {
		t.Error("in front of its target the wearer cast no Drain Life")
	}
}
