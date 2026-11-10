package core

import (
	"sync"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

const (
	WeaponTestSwordID int32 = 990901 + iota
	WeaponTestDaggerID
	WeaponTestShieldID
	WeaponTestBowID
	weaponTestAxeID
)

// A rogue that swings every weapon it holds and has no talents beyond what NewWeaponTestSim hands it.
// It and the test weapons register on first use, so only a test that asks for them has them.
type weaponTester struct {
	Character
}

var (
	registerWeaponTester sync.Once
	weaponTestTalents    func(*Character)
)

func (wt *weaponTester) GetCharacter() *Character { return &wt.Character }
func (wt *weaponTester) Initialize()              {}
func (wt *weaponTester) ApplyTalents() {
	if weaponTestTalents != nil {
		weaponTestTalents(&wt.Character)
	}
}
func (wt *weaponTester) Reset(_ *Simulation)            {}
func (wt *weaponTester) OnGCDReady(_ *Simulation)       {}
func (wt *weaponTester) OnEncounterStart(_ *Simulation) {}

func ensureWeaponTester() {
	registerWeaponTester.Do(func() {
		RegisterAgentFactory(
			proto.Player_Rogue{},
			proto.Spec_SpecRogue,
			func(character *Character, _ *proto.Player, _ *proto.Raid) Agent {
				wt := &weaponTester{Character: *character}
				wt.EnableAutoAttacks(wt, AutoAttackOptions{
					MainHand:        wt.WeaponFromMainHand(),
					OffHand:         wt.WeaponFromOffHand(),
					Ranged:          wt.WeaponFromRanged(),
					AutoSwingMelee:  true,
					AutoSwingRanged: true,
				})
				return wt
			},
			func(player *proto.Player, spec any) {
				player.Spec = spec.(*proto.Player_Rogue)
			},
		)

		damage := map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 50, WeaponDamageMax: 90}}
		AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
			{Id: WeaponTestSwordID, Name: "Test Sword", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand, WeaponSpeed: 2.6, ScalingOptions: damage},
			{Id: WeaponTestDaggerID, Name: "Test Dagger", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand, WeaponSpeed: 1.8, ScalingOptions: damage},
			{Id: WeaponTestShieldID, Name: "Test Shield", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}},
			{Id: WeaponTestBowID, Name: "Test Bow", Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow, WeaponSpeed: 3.0, ScalingOptions: damage},
			{Id: weaponTestAxeID, Name: "Test Axe", Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeOneHand, WeaponSpeed: 2.6, ScalingOptions: damage},
		}})
	})
}

func WeaponTestSlots(mainHand, offHand, ranged int32) []*proto.ItemSpec {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: mainHand}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHand}
	items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: ranged}
	return items
}

func weaponTestEncounter() *proto.Encounter {
	return &proto.Encounter{
		Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
		Duration: 180,
	}
}

func weaponTestRaid(race proto.Race, equipped, swapped []*proto.ItemSpec) *proto.Raid {
	ensureWeaponTester()
	player := &proto.Player{
		Name:      "Weapon Tester",
		Class:     proto.Class_ClassRogue,
		Race:      race,
		Buffs:     &proto.IndividualBuffs{},
		Spec:      &proto.Player_Rogue{},
		Equipment: &proto.EquipmentSpec{Items: equipped},
		Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	if swapped != nil {
		player.EnableItemSwap = true
		player.ItemSwap = &proto.ItemSwap{Items: swapped}
	}
	return &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}}
}

// A reset sim of the weapon tester, with talents applied as its talents.
func NewWeaponTestSim(t testing.TB, race proto.Race, equipped, swapped []*proto.ItemSpec, talents func(*Character)) (*Simulation, *Character) {
	t.Helper()
	weaponTestTalents = talents
	defer func() { weaponTestTalents = nil }()

	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       weaponTestRaid(race, equipped, swapped),
		Encounter:  weaponTestEncounter(),
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}
