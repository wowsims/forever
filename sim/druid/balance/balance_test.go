package balance

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func init() {
	RegisterBalanceDruid()
	common.RegisterAllEffects()
}

func TestBalance(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassDruid,
			Race:       proto.Race_RaceNightElf,
			OtherRaces: []proto.Race{proto.Race_RaceTauren},
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: &proto.Player_BalanceDruid{
				BalanceDruid: &proto.BalanceDruid{
					Options: &proto.BalanceDruid_Options{
						ClassOptions: &proto.DruidOptions{},
					},
				},
			}},
			GearSet: core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "p1"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "prebis"),
			},
			Talents:          DefaultTalents,
			Rotation:         core.GetAplRotation("../../../ui/specs/druid/balance/apls", "default"),
			Consumables:      DefaultConsumables,
			Profession1:      proto.Profession_Enchanting,
			Profession2:      proto.Profession_Tailoring,
			StartingDistance: 20,
			ItemFilter: core.ItemFilter{
				WeaponTypes:       DefaultWeaponTypes,
				ArmorType:         DefaultArmorType,
				RangedWeaponTypes: DefaultRangedWeaponTypes,
			},
			EPReferenceStat: proto.Stat_StatSpellDamage,
			StatsToWeigh: []proto.Stat{
				proto.Stat_StatIntellect,
				proto.Stat_StatSpellDamage,
				proto.Stat_StatArcaneDamage,
				proto.Stat_StatNatureDamage,
				proto.Stat_StatSpellPiercing,
				proto.Stat_StatHitRating,
				proto.Stat_StatCritRating,
				proto.Stat_StatHasteRating,
				proto.Stat_StatMP5,
				proto.Stat_StatSpirit,
			},
		},
	}))
}

// The Moonkin preset on the Balance page.
const DefaultTalents = "5232220115501351--5053"

const DefaultArmorType = proto.ArmorType_ArmorTypeLeather

var DefaultWeaponTypes = []proto.WeaponType{
	proto.WeaponType_WeaponTypeDagger,
	proto.WeaponType_WeaponTypeMace,
	proto.WeaponType_WeaponTypeStaff,
	proto.WeaponType_WeaponTypeOffHand,
}

var DefaultRangedWeaponTypes = []proto.RangedWeaponType{
	proto.RangedWeaponType_RangedWeaponTypeIdol,
}

// The Balance page's consumes: Flask of Supreme Power, Greater Arcane Elixir, Elixir of Nature Power,
// Elixir of the Owl, Greater Mageblood Elixir, Lobster Stew, Brilliant Wizard Oil, Major Spellblasting
// Potion, Major Mana Potion, Demonic Rune and SAF-T Clever Dynamite.
var DefaultConsumables = &proto.ConsumesSpec{
	FlaskId:        13512,
	ElixirIds:      []int32{13454, 250343, 250337, 250341},
	FoodId:         13933,
	MhImbueId:      20749,
	OffensivePotId: 250937,
	ResourcePotId:  13444,
	ConjuredId:     12662,
	ExplosiveId:    260814,
}

// Self keeps the mana. Someone else still spends the cast and the cooldown; a dummy receives the mana.
func TestGiftedInnervateCastsWithoutTheMana(t *testing.T) {
	for _, c := range []struct {
		name     string
		gift     bool
		wantMana bool
	}{
		{"self", false, true},
		{"someone else", true, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, druid := balanceDruid(t, c.gift, 0)
			spell := druid.GetSpell(core.ActionID{SpellID: 29166})
			if spell == nil {
				t.Fatal("Innervate is not registered")
			}

			before := druid.ManaRegenPerSecondWhileNotCasting()
			dummy := sim.Raid.GetFirstTargetDummy()
			if dummy == nil {
				t.Fatal("Balance has no dummy")
			}

			target := sim.Encounter.ActiveTargetUnits[0]
			if !spell.Cast(sim, target) {
				t.Fatal("Innervate did not cast")
			}
			if spell.IsReady(sim) {
				t.Error("Innervate is still ready")
			}
			if got := spell.SpellMetrics[target.UnitIndex].Casts; got != 1 {
				t.Errorf("Innervate casts = %d, want 1", got)
			}

			active := druid.HasActiveAuraWithTag(buffs.InnervatesCategory)
			gotMana := druid.ManaRegenPerSecondWhileNotCasting() != before
			if active != c.wantMana || gotMana != c.wantMana {
				t.Errorf("mana buff active = %v, regen changed = %v, want %v", active, gotMana, c.wantMana)
			}
			if gotDummy := dummy.HasActiveAuraWithTag(buffs.InnervatesCategory); gotDummy == c.wantMana {
				t.Errorf("dummy Innervate = %v, want %v", gotDummy, !c.wantMana)
			}
		})
	}

	t.Run("someone else with an incoming innervate", func(t *testing.T) {
		sim, druid := balanceDruid(t, true, 1)
		spell := druid.GetSpell(core.ActionID{SpellID: 29166})
		if spell == nil {
			t.Fatal("Innervate is not registered")
		}
		incoming := druid.GetAura("Innervates (External)")
		if incoming == nil {
			t.Fatal("the incoming Innervate is not registered")
		}
		incoming.Activate(sim)
		if !druid.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Fatal("the incoming Innervate is not up")
		}
		target := sim.Encounter.ActiveTargetUnits[0]
		if !spell.Cast(sim, target) {
			t.Fatal("Innervate did not cast")
		}
		if spell.IsReady(sim) {
			t.Error("Innervate is still ready")
		}
		if got := spell.SpellMetrics[target.UnitIndex].Casts; got != 1 {
			t.Errorf("Innervate casts = %d, want 1", got)
		}
		dummy := sim.Raid.GetFirstTargetDummy()
		if dummy == nil {
			t.Fatal("Balance has no dummy")
		}
		if !dummy.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Error("the dummy has no Innervate")
		}
	})

	t.Run("cast at the dummy", func(t *testing.T) {
		sim, druid := balanceDruid(t, false, 0)
		spell := druid.GetSpell(core.ActionID{SpellID: 29166})
		if spell == nil {
			t.Fatal("Innervate is not registered")
		}
		dummy := sim.Raid.GetFirstTargetDummy()
		if dummy == nil {
			t.Fatal("Balance has no dummy")
		}
		before := druid.ManaRegenPerSecondWhileNotCasting()
		if !spell.Cast(sim, &dummy.Unit) {
			t.Fatal("Innervate did not cast")
		}
		if druid.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Error("the caster has Innervate")
		}
		if druid.ManaRegenPerSecondWhileNotCasting() != before {
			t.Error("the caster's regen changed")
		}
		if !dummy.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Error("the dummy has no Innervate")
		}
	})
}

func balanceDruid(t *testing.T, gift bool, incoming int32) (*core.Simulation, *core.Character) {
	t.Helper()
	target := &proto.UnitReference{}
	if gift {
		target = &proto.UnitReference{Type: proto.UnitReference_Player, Index: 1}
	}
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassDruid,
		Race:          proto.Race_RaceNightElf,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: DefaultTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		Buffs:         &proto.IndividualBuffs{Innervates: incoming},
	}, &proto.Player_BalanceDruid{
		BalanceDruid: &proto.BalanceDruid{
			Options: &proto.BalanceDruid_Options{
				ClassOptions: &proto.DruidOptions{InnervateTarget: target},
			},
		},
	})
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.TargetDummies = 1
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}
