package shared

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	core.RegisterAgentFactory(
		proto.Player_EnhancementShaman{},
		proto.Spec_SpecEnhancementShaman,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			wielder := &testCaster{Character: *character}
			wielder.EnableAutoAttacks(wielder, core.AutoAttackOptions{
				MainHand:       wielder.WeaponFromMainHand(),
				OffHand:        wielder.WeaponFromOffHand(),
				AutoSwingMelee: true,
			})
			return wielder
		},
		func(player *proto.Player, spec any) {
			player.Spec = spec.(*proto.Player_EnhancementShaman)
		},
	)
}

// Hit Weapon +2, 430391: A_MOD_HIT_CHANCE on the wearer, restricted to melee weapons.
const hitWeapon int32 = 430391

// A weapon enchant whose equip spell is weapon-scoped puts its aura up while the enchant is worn, and
// the aura's hit reaches the attacks made with the hand the enchant is on, while that hand holds a
// weapon its row names. The row is narrowed to daggers and freed of its area, and the enchant sits on
// the off-hand dagger beside a main-hand sword or a second dagger.
func TestWeaponScopedEnchantReachesTheWeaponItNames(t *testing.T) {
	const swordID, daggerID, enchantID int32 = 990981, 990982, 990983

	editRow(t, hitWeapon, func(s *spelldata.Spell) {
		s.RequiredAreas = 0
		s.EquipSubclass = dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS
	})

	dagger := testOneHander(daggerID)
	dagger.Name, dagger.WeaponType, dagger.WeaponSpeed = "Test Dagger", proto.WeaponType_WeaponTypeDagger, 1.8
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{testOneHander(swordID), dagger},
		Enchants: []*proto.SimEnchant{{EffectId: enchantID, Name: "Test Hit Enchant",
			Type: proto.ItemType_ItemTypeWeapon}},
	})
	registerSpellDataEquipAura(SpellDataProc{Name: "Test Hit Enchant", EnchantID: enchantID, TriggerSpellID: hitWeapon})
	hit := spelldata.MustFind(hitWeapon).EffectN(1).Mean()
	needAmount(t, "430391's hit", hit)

	for _, tc := range []struct {
		name           string
		mainHand       int32
		enchanted      bool
		wantMH, wantOH float64
	}{
		{"enchanted", swordID, true, 0, hit},
		{"not enchanted", swordID, false, 0, 0},
		{"enchanted beside a main-hand dagger", daggerID, true, 0, hit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			offHand := &proto.ItemSpec{Id: daggerID}
			if tc.enchanted {
				offHand.Enchant = enchantID
			}
			sim := core.NewSim(&proto.RaidSimRequest{
				SimOptions: &proto.SimOptions{RandomSeed: 1},
				Raid: &proto.Raid{Parties: []*proto.Party{{
					Players: []*proto.Player{{
						Name:      "Wielder",
						Class:     proto.Class_ClassShaman,
						Race:      proto.Race_RaceTroll,
						Buffs:     &proto.IndividualBuffs{},
						Spec:      &proto.Player_EnhancementShaman{},
						Equipment: &proto.EquipmentSpec{Items: testHands(&proto.ItemSpec{Id: tc.mainHand}, offHand)},
					}},
					Buffs: &proto.PartyBuffs{},
				}}},
				Encounter: &proto.Encounter{
					Targets:  []*proto.Target{{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon}},
					Duration: 180,
				},
			}, simsignals.CreateSignals())
			sim.Reset()
			wielder := sim.Raid.Parties[0].Players[0].(*testCaster)

			if !wielder.AutoAttacks.IsDualWielding {
				t.Fatal("the two one-handers should dual wield")
			}
			mh, oh := wielder.AutoAttacks.MHAuto().BonusHitPercent, wielder.AutoAttacks.OHAuto().BonusHitPercent
			if mh != tc.wantMH || oh != tc.wantOH {
				t.Errorf("main-hand and off-hand hit = %0.2f%% and %0.2f%%, want %0.2f%% and %0.2f%%",
					mh, oh, tc.wantMH, tc.wantOH)
			}
		})
	}
}
