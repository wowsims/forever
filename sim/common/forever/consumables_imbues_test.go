//go:build with_db

package forever_test

import (
	"testing"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	forever.RegisterAllTemporaryEnchantConsumables()
}

const (
	notchedShortsword  = 727  // a one-hand sword
	smallHandBlade     = 816  // a one-hand dagger
	aegisOfStormwind   = 1203 // a shield
	elementalSharpener = 18262
)

func weaponSlots(mainHand, offHand int32) []*proto.ItemSpec {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = mainHand
	items[proto.ItemSlot_ItemSlotOffHand].Id = offHand
	return items
}

func stonedWarrior(t *testing.T, mhImbueId, ohImbueId int32, swapped []*proto.ItemSpec) (*core.Simulation, *core.Character) {
	t.Helper()
	player := &proto.Player{
		Name: "Consumer", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc,
		Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
			ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker}}}},
		Equipment:   &proto.EquipmentSpec{Items: weaponSlots(notchedShortsword, smallHandBlade)},
		Consumables: &proto.ConsumesSpec{MhImbueId: mhImbueId, OhImbueId: ohImbueId},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	if swapped != nil {
		player.EnableItemSwap = true
		player.ItemSwap = &proto.ItemSwap{Items: swapped}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Elemental Sharpening Stone 18262's 22755 states A_MOD_WEAPON_CRIT_PERCENT 2 on the melee weapons
// its mask 42483 names, and each weapon carries its own copy: the crit reaches the attacks made with
// the hand the stone is on, and no other attack.
func TestElementalSharpeningStoneCritReachesItsOwnHand(t *testing.T) {
	_, bare := stonedWarrior(t, 0, 0, nil)
	for _, tc := range []struct {
		name                 string
		mhImbueId, ohImbueId int32
		wantMH, wantOH       float64
	}{
		{"main hand", elementalSharpener, 0, 2, 0},
		{"off hand", 0, elementalSharpener, 0, 2},
		{"both hands", elementalSharpener, elementalSharpener, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, stoned := stonedWarrior(t, tc.mhImbueId, tc.ohImbueId, nil)

			if delta := stoned.GetStat(stats.PhysicalCritPercent) - bare.GetStat(stats.PhysicalCritPercent); delta != 0 {
				t.Errorf("the stone adds %v%% crit to every attack, want none", delta)
			}
			mh := stoned.AutoAttacks.MHAuto().BonusCritPercent - bare.AutoAttacks.MHAuto().BonusCritPercent
			oh := stoned.AutoAttacks.OHAuto().BonusCritPercent - bare.AutoAttacks.OHAuto().BonusCritPercent
			if mh != tc.wantMH || oh != tc.wantOH {
				t.Errorf("the stone adds %v%% main-hand and %v%% off-hand crit, want %v and %v", mh, oh, tc.wantMH, tc.wantOH)
			}
		})
	}
}

func TestElementalSharpeningStoneSkipsAMagicDefenseSpell(t *testing.T) {
	_, character := stonedWarrior(t, elementalSharpener, 0, nil)

	magicSpell := character.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 990980},
		SpellSchool: core.SpellSchoolFrost,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		DefenseType: core.DefenseTypeMagic,
	})
	if magicSpell.BonusCritPercent != 0 {
		t.Errorf("a magic-defense spell off the stoned hand gains %v%% crit, want none", magicSpell.BonusCritPercent)
	}
	if mh := character.AutoAttacks.MHAuto().BonusCritPercent; mh != 2 {
		t.Errorf("the main-hand auto gains %v%% crit, want 2", mh)
	}
}

func TestElementalSharpeningStoneTurnsOffWhenTheStonedWeaponSwapsAway(t *testing.T) {
	for _, tc := range []struct {
		name    string
		swapped []*proto.ItemSpec
	}{
		{"swapped to a shield", weaponSlots(aegisOfStormwind, smallHandBlade)},
		{"swapped to an empty hand", weaponSlots(0, smallHandBlade)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, character := stonedWarrior(t, elementalSharpener, 0, tc.swapped)
			mhAuto := character.AutoAttacks.MHAuto()
			if mhAuto.BonusCritPercent != 2 {
				t.Fatalf("the main-hand auto gains %v%% crit before the swap, want 2", mhAuto.BonusCritPercent)
			}

			character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)

			if mhAuto.BonusCritPercent != 0 {
				t.Errorf("the main-hand auto keeps %v%% crit once the stoned weapon is swapped away, want none", mhAuto.BonusCritPercent)
			}
		})
	}
}

// Scroll of Imbue Baleflame 277487 puts enchant 8706 on a staff, which states 2 of the fire spell
// penetration stat 119, and Scroll of Imbue Balefrost 277496 puts 8714 on it, which states 3 of the
// frost one, 121. Each pierces its own school's resistance and neither adds SpellPiercing, which
// would reach every school.
func TestBaleImbuesPierceOneSchool(t *testing.T) {
	const staff = 873
	for _, c := range []struct {
		imbue       int32
		fire, frost float64
	}{
		{0, 0, 0},
		{277487, 2, 0},
		{277496, 0, 3},
	} {
		_, character := warriorWithConsumes(t, &proto.ConsumesSpec{MhImbueId: c.imbue}, staff, 0, proto.MobType_MobTypeUnknown)

		piercing := character.PseudoStats.SchoolSpellPiercing
		if piercing[stats.SchoolIndexFire] != c.fire || piercing[stats.SchoolIndexFrost] != c.frost {
			t.Errorf("imbue %d: fire and frost piercing %v and %v, want %v and %v", c.imbue,
				piercing[stats.SchoolIndexFire], piercing[stats.SchoolIndexFrost], c.fire, c.frost)
		}
		if got := character.GetStat(stats.SpellPiercing); got != 0 {
			t.Errorf("imbue %d: SpellPiercing %v, want 0", c.imbue, got)
		}
	}
}

// Scroll of Imbue Spark 277489 takes a one-hand sword and Scroll of Imbue Manablade 277498 a dagger.
// Spark's listener, which moves a spell category's cooldown, registers on any wearer; Manablade's,
// which restores mana, only on a wearer with a mana bar, which a warrior lacks.
func TestProcImbuesRegisterTheirListeners(t *testing.T) {
	for _, c := range []struct {
		imbue, weapon, enchant int32
		name                   string
		listens                bool
	}{
		{277489, notchedShortsword, 8707, "Scroll of Imbue Spark", true},
		{277498, smallHandBlade, 8715, "Scroll of Imbue Manablade", false},
	} {
		_, character := warriorWithConsumes(t, &proto.ConsumesSpec{MhWeaponImbueId: c.imbue}, c.weapon, 0, proto.MobType_MobTypeUnknown)

		if got := character.Equipment[proto.ItemSlot_ItemSlotMainHand].WeaponImbue; got != c.enchant {
			t.Errorf("%s puts enchant %d on the main hand, want %d", c.name, got, c.enchant)
		}
		if got := character.GetAura(c.name) != nil; got != c.listens {
			t.Errorf("%s registers a listener: %v, want %v", c.name, got, c.listens)
		}
	}
}
