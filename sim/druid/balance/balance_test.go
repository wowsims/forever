package balance

import (
	"testing"
	"time"

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
			sim, druid := balanceDruid(t, c.gift, 0, 1)
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
		sim, druid := balanceDruid(t, true, 1, 1)
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
		sim, druid := balanceDruid(t, false, 0, 1)
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

func balanceDruid(t *testing.T, gift bool, incoming int32, targets int) (*core.Simulation, *core.Character) {
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
	encounter := core.MakeSingleTargetEncounter(0)
	for len(encounter.Targets) < targets {
		encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1, UseLabeledRands: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// One Hurricane tick rolls Omen of Clarity once, however many targets it lands on: three targets proc
// a little more often than one, since one of them nearly always lands, where a roll per target would
// nearly triple it.
func TestHurricaneTickRollsOmenOnce(t *testing.T) {
	const ticks = 2000
	procsOn := func(targets int) int {
		sim, _ := balanceDruid(t, false, 0, targets)
		druid := sim.Raid.Parties[0].Players[0].(*BalanceDruid)
		if len(sim.Encounter.ActiveTargetUnits) != targets {
			t.Fatalf("targets = %d, want %d", len(sim.Encounter.ActiveTargetUnits), targets)
		}
		procs := 0
		for range ticks {
			// The step clears the proc's cooldown. A successful roll starts that cooldown on the spot; the
			// Clearcasting aura follows a spell batch window later.
			sim.CurrentTime += time.Minute
			druid.Hurricane.RelatedDotSpell.Cast(sim, druid.CurrentTarget)
			if !druid.OmenOfClarityAura.Icd.IsReady(sim) {
				procs++
			}
		}
		return procs
	}

	one, three := procsOn(1), procsOn(3)
	if one < 100 {
		t.Fatalf("one target procs Omen of Clarity %d of %d ticks, too few to compare", one, ticks)
	}
	if float64(three) > float64(one)*1.3 {
		t.Errorf("three targets proc Omen of Clarity %d of %d ticks, one target %d; want about the same", three, ticks, one)
	}
}
