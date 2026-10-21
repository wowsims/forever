//go:build with_db

package forever_test

import (
	"testing"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
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

func stonedWarrior(t *testing.T, consumes *proto.ConsumesSpec, swapped []*proto.ItemSpec) (*core.Simulation, *core.Character) {
	t.Helper()
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       stonedRaid(consumes, swapped),
		Encounter:  stonedEncounter,
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

var stonedEncounter = &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180}

func stonedRaid(consumes *proto.ConsumesSpec, swapped []*proto.ItemSpec) *proto.Raid {
	player := &proto.Player{
		Name: "Consumer", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc,
		Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
			ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker}}}},
		Equipment:   &proto.EquipmentSpec{Items: weaponSlots(notchedShortsword, smallHandBlade)},
		Consumables: consumes,
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	if swapped != nil {
		player.EnableItemSwap = true
		player.ItemSwap = &proto.ItemSwap{Items: swapped}
	}
	return &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}}
}

// The A_MOD_WEAPON_CRIT_PERCENT amount of 22755, the spell Elemental Sharpening Stone 18262 puts on
// the weapon.
func stoneCrit(t *testing.T) float64 {
	t.Helper()
	crit := spelldata.MustFind(22755).EffectN(1).Average()
	if crit == 0 {
		t.Fatal("22755 states no crit for the stone to add")
	}
	return crit
}

// Elemental Sharpening Stone 18262's 22755 states A_MOD_WEAPON_CRIT_PERCENT on the melee weapon
// subclasses its mask names, and each weapon carries its own copy: the crit reaches the attacks made
// with the hand the stone is on, and no other attack.
func TestElementalSharpeningStoneCritReachesItsOwnHand(t *testing.T) {
	_, bare := stonedWarrior(t, &proto.ConsumesSpec{}, nil)
	crit := stoneCrit(t)
	for _, tc := range []struct {
		name                 string
		mhImbueId, ohImbueId int32
		mhStones, ohStones   float64
	}{
		{"main hand", elementalSharpener, 0, 1, 0},
		{"off hand", 0, elementalSharpener, 0, 1},
		{"both hands", elementalSharpener, elementalSharpener, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, stoned := stonedWarrior(t, &proto.ConsumesSpec{MhImbueId: tc.mhImbueId, OhImbueId: tc.ohImbueId}, nil)

			if delta := stoned.GetStat(stats.PhysicalCritPercent) - bare.GetStat(stats.PhysicalCritPercent); delta != 0 {
				t.Errorf("the stone adds %v%% crit to every attack, want none", delta)
			}
			wantMH, wantOH := crit*tc.mhStones, crit*tc.ohStones
			mh := stoned.AutoAttacks.MHAuto().BonusCritPercent - bare.AutoAttacks.MHAuto().BonusCritPercent
			oh := stoned.AutoAttacks.OHAuto().BonusCritPercent - bare.AutoAttacks.OHAuto().BonusCritPercent
			if mh != wantMH || oh != wantOH {
				t.Errorf("the stone adds %v%% main-hand and %v%% off-hand crit, want %v and %v", mh, oh, wantMH, wantOH)
			}
		})
	}
}

// The stone's crit on the main hand shows on the character sheet as a consumable: the consumes
// column gains it over the buffs column, and the gear column is a bare warrior's.
func TestElementalSharpeningStoneCritShowsAsAConsumableOnTheSheet(t *testing.T) {
	sheet := func(consumes *proto.ConsumesSpec) *proto.PlayerStats {
		request := &proto.ComputeStatsRequest{Raid: stonedRaid(consumes, nil), Encounter: stonedEncounter}
		return core.ComputeStats(request).RaidStats.Parties[0].Players[0]
	}
	bare, stoned := sheet(&proto.ConsumesSpec{}), sheet(&proto.ConsumesSpec{MhImbueId: elementalSharpener})

	crit := proto.PseudoStat_PseudoStatMeleeCritPercent
	if got, want := stoned.GearStats.PseudoStats[crit], bare.GearStats.PseudoStats[crit]; got != want {
		t.Errorf("the gear column shows %v%% main-hand crit, want the bare warrior's %v%%", got, want)
	}
	gained := (stoned.ConsumesStats.PseudoStats[crit] - stoned.BuffsStats.PseudoStats[crit]) -
		(bare.ConsumesStats.PseudoStats[crit] - bare.BuffsStats.PseudoStats[crit])
	if want := stoneCrit(t); !core.WithinToleranceFloat64(want, gained, 1e-9) {
		t.Errorf("the consumes column gains %v%% main-hand crit from the stone, want %v%%", gained, want)
	}
}

func TestElementalSharpeningStoneSkipsAMagicDefenseSpell(t *testing.T) {
	_, character := stonedWarrior(t, &proto.ConsumesSpec{MhImbueId: elementalSharpener}, nil)

	magicSpell := character.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 990980},
		SpellSchool: core.SpellSchoolFrost,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		DefenseType: core.DefenseTypeMagic,
	})
	if magicSpell.BonusCritPercent != 0 {
		t.Errorf("a magic-defense spell off the stoned hand gains %v%% crit, want none", magicSpell.BonusCritPercent)
	}
	if mh, want := character.AutoAttacks.MHAuto().BonusCritPercent, stoneCrit(t); mh != want {
		t.Errorf("the main-hand auto gains %v%% crit, want %v", mh, want)
	}
}

func TestElementalSharpeningStoneTurnsOffWhenTheStonedWeaponSwapsAway(t *testing.T) {
	for _, tc := range []struct {
		name    string
		swapped []*proto.ItemSpec
	}{
		{"swapped to a shield", weaponSlots(aegisOfStormwind, smallHandBlade)},
		{"swapped to an empty hand", weaponSlots(0, smallHandBlade)},
		{"swapped to a bare sword the stone fits", weaponSlots(notchedShortsword, smallHandBlade)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, character := stonedWarrior(t, &proto.ConsumesSpec{MhImbueId: elementalSharpener}, tc.swapped)
			mhAuto := character.AutoAttacks.MHAuto()
			if want := stoneCrit(t); mhAuto.BonusCritPercent != want {
				t.Fatalf("the main-hand auto gains %v%% crit before the swap, want %v", mhAuto.BonusCritPercent, want)
			}

			character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)

			if mhAuto.BonusCritPercent != 0 {
				t.Errorf("the main-hand auto keeps %v%% crit once the stoned weapon is swapped away, want none", mhAuto.BonusCritPercent)
			}
		})
	}
}

// A stone the spec names for the item swap set's main hand reaches the main-hand attacks while that
// weapon is worn, and no attack of the main set's.
func TestElementalSharpeningStoneOnTheSwapWeaponTurnsOnWhenSwappedIn(t *testing.T) {
	sim, character := stonedWarrior(t, &proto.ConsumesSpec{SwapMhImbueId: elementalSharpener}, weaponSlots(notchedShortsword, smallHandBlade))
	mhAuto, ohAuto := character.AutoAttacks.MHAuto(), character.AutoAttacks.OHAuto()
	if mhAuto.BonusCritPercent != 0 || ohAuto.BonusCritPercent != 0 {
		t.Fatalf("the main set's autos gain %v%% and %v%% crit, want none", mhAuto.BonusCritPercent, ohAuto.BonusCritPercent)
	}

	character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	if want := stoneCrit(t); mhAuto.BonusCritPercent != want || ohAuto.BonusCritPercent != 0 {
		t.Errorf("with the stoned swap weapon worn the autos gain %v%% and %v%% crit, want %v on the main hand alone",
			mhAuto.BonusCritPercent, ohAuto.BonusCritPercent, want)
	}

	character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Main, false)
	if mhAuto.BonusCritPercent != 0 {
		t.Errorf("the main-hand auto keeps %v%% crit once the main set is worn again, want none", mhAuto.BonusCritPercent)
	}
}

// Scroll of Imbue Baleflame 277487 puts enchant 8706 on a staff, which states the fire spell
// penetration stat 119, and Scroll of Imbue Balefrost 277496 puts 8714 on it, which states the frost
// one, 121. Each pierces its own school's resistance and neither adds SpellPiercing, which would
// reach every school.
func TestBaleImbuesPierceOneSchool(t *testing.T) {
	const staff = 873
	pierces := func(amount float64, want bool) bool {
		if want {
			return amount > 0
		}
		return amount == 0
	}
	for _, c := range []struct {
		imbue       int32
		fire, frost bool
	}{
		{0, false, false},
		{277487, true, false},
		{277496, false, true},
	} {
		_, character := warriorWithConsumes(t, &proto.ConsumesSpec{MhImbueId: c.imbue}, staff, 0, proto.MobType_MobTypeUnknown)

		piercing := character.PseudoStats.SchoolSpellPiercing
		fire, frost := piercing[stats.SchoolIndexFire], piercing[stats.SchoolIndexFrost]
		if !pierces(fire, c.fire) || !pierces(frost, c.frost) {
			t.Errorf("imbue %d: fire and frost piercing %v and %v, want fire %v and frost %v", c.imbue,
				fire, frost, c.fire, c.frost)
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
		imbue, weapon int32
		name          string
		listens       bool
	}{
		{277489, notchedShortsword, "Scroll of Imbue Spark", true},
		{277498, smallHandBlade, "Scroll of Imbue Manablade", false},
	} {
		_, character := warriorWithConsumes(t, &proto.ConsumesSpec{MhWeaponImbueId: c.imbue}, c.weapon, 0, proto.MobType_MobTypeUnknown)

		enchant := core.GetConsumableByID(c.imbue).TemporaryEnchant
		if enchant == nil {
			t.Fatalf("%s states no temporary enchant", c.name)
		}
		if got := character.Equipment[proto.ItemSlot_ItemSlotMainHand].WeaponImbue; got != enchant.EffectID {
			t.Errorf("%s puts enchant %d on the main hand, want %d", c.name, got, enchant.EffectID)
		}
		if got := character.GetAura(c.name) != nil; got != c.listens {
			t.Errorf("%s registers a listener: %v, want %v", c.name, got, c.listens)
		}
	}
}
