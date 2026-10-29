//go:build with_db

package forever_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/core/testagent"
)

func init() {
	testagent.Register(proto.Player_HolyPaladin{}, proto.Spec_SpecHolyPaladin, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_HolyPaladin)
	}, func(agent *testagent.Agent) {
		agent.EnableManaBar()
	})
}

// A character wearing one item, against a level 60 target: the warrior stand-in, with a paladin the
// stand-in registered as the holy paladin, which has a mana bar, or with a rogue the stand-in
// registered as the rogue, which shoots its ranged weapon.
func itemWearer(t *testing.T, class proto.Class, slot proto.ItemSlot, itemID int32) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[slot].Id = itemID
	return itemWearerWithEquipment(t, class, &proto.EquipmentSpec{Items: items})
}

func itemWearerWithEquipment(t *testing.T, class proto.Class, equipment *proto.EquipmentSpec) (*core.Simulation, *core.Character) {
	t.Helper()
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceHuman,
		Equipment:   equipment,
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	switch class {
	case proto.Class_ClassPaladin:
		core.WithSpec(player, &proto.Player_HolyPaladin{HolyPaladin: &proto.HolyPaladin{}})
	case proto.Class_ClassRogue:
		core.WithSpec(player, &proto.Player_Rogue{Rogue: &proto.Rogue{}})
	default:
		core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Whether every cast landed within the effect's roll. The caller rules out crits.
func damageInRange(metrics core.SpellMetrics, effect *spelldata.Effect) bool {
	casts := float64(metrics.Casts)
	return metrics.Crits == 0 && metrics.TotalDamage >= casts*effect.Min() && metrics.TotalDamage <= casts*effect.Max()
}

// Aegis of the Scarlet Commander 7726's 1293539 deals its row's Holy roll to the attacker of a melee
// hit the wearer blocks, at most once per the row's RecoveryTime; a hit not blocked deals nothing.
func TestAegisOfTheScarletCommanderDealsHolyDamageOnBlock(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotOffHand, 7726)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100, stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	enemy := sim.Encounter.ActiveTargetUnits[0]
	retribution := character.GetSpell(core.ActionID{SpellID: 1293539})
	if retribution == nil {
		t.Fatal("7726 registers no damage spell")
	}
	row := spelldata.MustFind(1293539)
	damage := row.FirstAura(dbcenums.A_PROC_TRIGGER_DAMAGE)
	lockout := row.Cooldown()
	if lockout <= 0 {
		t.Fatalf("1293539 states no RecoveryTime")
	}

	takeHit(t, sim, character, meleeSwing(enemy))
	if got := retribution.SpellMetrics[enemy.UnitIndex].Casts; got != 0 {
		t.Errorf("a hit not blocked cast the damage %d times, want none", got)
	}

	block := func() {
		character.OnSpellHitTaken(sim, meleeSwing(enemy), &core.SpellResult{Target: &character.Unit, Outcome: core.OutcomeHit | core.OutcomeBlock, Damage: 50})
		stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	}
	firstBlock := sim.CurrentTime
	block()
	metrics := retribution.SpellMetrics[enemy.UnitIndex]
	if metrics.Casts != 1 || !damageInRange(metrics, damage) {
		t.Errorf("a block cast the damage %d times for %v, want once for %v to %v", metrics.Casts, metrics.TotalDamage, damage.Min(), damage.Max())
	}
	block()
	if got := retribution.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
		t.Errorf("a second block at once cast the damage %d times in all, want once still", got)
	}
	stepTo(t, sim, firstBlock+lockout-core.SpellBatchWindow)
	block()
	if got := retribution.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
		t.Errorf("a block just inside the row's %v cast the damage %d times in all, want once still", lockout, got)
	}
	stepTo(t, sim, firstBlock+lockout)
	block()
	if got := retribution.SpellMetrics[enemy.UnitIndex].Casts; got != 2 {
		t.Errorf("a block the row's %v after the first cast the damage %d times in all, want 2", lockout, got)
	}
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
// effect's share of the damage it dealt.
func TestDemonForgedBreastplateDrainsTheAttacker(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotChest, 12628)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
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
	if got, want := character.CurrentHealth()-health, metrics.TotalDamage*float64(leech.Amplitude); got != want {
		t.Errorf("the drain healed the wearer for %v, want the effect's share of the %v it dealt, %v", got, metrics.TotalDamage, want)
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
