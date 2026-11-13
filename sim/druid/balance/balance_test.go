package balance

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
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

// A Balance druid in a fresh sim with this many targets; a duration other than 0 resets the sim to run that long.
func newBalanceDruid(t *testing.T, targets int, duration time.Duration) (*core.Simulation, *BalanceDruid) {
	t.Helper()
	sim, _ := balanceDruid(t, false, 0, targets)
	if duration != 0 {
		sim.Cleanup()
		sim.BaseDuration = duration
		sim.Reset()
	}
	return sim, sim.Raid.Parties[0].Players[0].(*BalanceDruid)
}

// Steps the sim until the clock reaches until.
func stepUntil(sim *core.Simulation, until time.Duration) {
	for sim.CurrentTime < until && !sim.Step() {
	}
}

// How many of casts casts end with Clearcasting active 100 ms later. The buff makes each cast free;
// settle waits out the proc's cooldown.
func countOmenProcs(sim *core.Simulation, druid *BalanceDruid, casts int, cast, settle func(start time.Duration)) int {
	procs := 0
	for range casts {
		druid.ClearcastingAura.Activate(sim)
		start := sim.CurrentTime
		cast(start)
		stepUntil(sim, start+100*time.Millisecond)
		if druid.ClearcastingAura.IsActive() {
			procs++
		}
		settle(start)
	}
	return procs
}

// Hurricane's ticks roll no Omen of Clarity, and their crits grant no Nature's Grace.
func TestHurricaneTicksRollNoOmenAndGrantNoNaturesGrace(t *testing.T) {
	t.Run("Omen of Clarity", func(t *testing.T) {
		sim, druid := newBalanceDruid(t, 3, 0)
		for range 500 {
			// The step clears the proc's cooldown, which a successful roll starts on the spot.
			sim.CurrentTime += time.Minute
			druid.Hurricane.RelatedDotSpell.Cast(sim, druid.CurrentTarget)
			if !druid.OmenOfClarityAura.Icd.IsReady(sim) {
				t.Fatal("a Hurricane tick rolled Omen of Clarity")
			}
		}
	})

	t.Run("Nature's Grace", func(t *testing.T) {
		sim, druid := newBalanceDruid(t, 1, 0)
		naturesGrace := druid.GetAura("Nature's Grace")
		if naturesGrace == nil {
			t.Fatal("Nature's Grace is not registered")
		}
		druid.AddStatDynamic(sim, stats.SpellCritPercent, 100)
		tick := druid.Hurricane.RelatedDotSpell
		start := sim.CurrentTime
		tick.Cast(sim, druid.CurrentTarget)
		stepUntil(sim, start+time.Second)
		if tick.SpellMetrics[druid.CurrentTarget.UnitIndex].Crits == 0 {
			t.Fatal("the tick did not crit")
		}
		if naturesGrace.IsActive() {
			t.Error("a Hurricane tick crit granted Nature's Grace")
		}
	})
}

// The Hurricane cast rolls Omen of Clarity once for each target its area reaches, after it spends
// Clearcasting, so a cast that spends the buff can come out with a new one.
func TestHurricaneCastRollsOmenPerTarget(t *testing.T) {
	const casts = 400
	procsOn := func(targets int) int {
		sim, druid := newBalanceDruid(t, targets, casts*time.Minute)
		return countOmenProcs(sim, druid, casts, func(_ time.Duration) {
			if !druid.Hurricane.Cast(sim, druid.CurrentTarget) {
				t.Fatal("Hurricane did not cast")
			}
		}, func(_ time.Duration) {
			// The channel outlasts the proc's cooldown.
			for druid.Hurricane.AOEDot().IsActive() && !sim.Step() {
			}
		})
	}

	one, five := procsOn(1), procsOn(5)
	if one == 0 {
		t.Fatalf("a Hurricane cast that spends Clearcasting on one target never comes out with a new one in %d casts", casts)
	}
	if float64(five) < 2.5*float64(one) {
		t.Errorf("Hurricane procs Omen of Clarity %d of %d casts on five targets, %d on one; want about four times as often", five, casts, one)
	}
}

// A helpful cast such as Thorns rolls Omen of Clarity in Moonkin Form, after it spends Clearcasting;
// in caster form it rolls nothing, and neither does a shift into Moonkin Form.
func TestHelpfulCastsRollOmenInMoonkinFormOnly(t *testing.T) {
	const casts = 400
	newDruid := func() (*core.Simulation, *BalanceDruid) {
		return newBalanceDruid(t, 1, casts*10*time.Second)
	}
	procsIn := func(moonkin bool) int {
		sim, druid := newDruid()
		if !moonkin {
			druid.ClearForm(sim)
		}
		return countOmenProcs(sim, druid, casts, func(_ time.Duration) {
			if !druid.Thorns.Cast(sim, &druid.Unit) {
				t.Fatal("Thorns did not cast")
			}
		}, func(start time.Duration) {
			// The wait outlasts the proc's cooldown.
			stepUntil(sim, start+10*time.Second)
		})
	}

	if got := procsIn(true); got == 0 {
		t.Errorf("Thorns never procs Omen of Clarity in Moonkin Form in %d casts", casts)
	}
	if got := procsIn(false); got != 0 {
		t.Errorf("Thorns procs Omen of Clarity %d of %d casts in caster form, want none", got, casts)
	}

	sim, druid := newDruid()
	// A free shift, so the mana lasts; a roll would start the proc's cooldown on the spot.
	druid.MoonkinForm.Cost.PercentModifier = 0
	for range casts {
		druid.ClearForm(sim)
		start := sim.CurrentTime
		if !druid.MoonkinForm.Cast(sim, &druid.Unit) {
			t.Fatal("Moonkin Form did not cast")
		}
		if !druid.OmenOfClarityAura.Icd.IsReady(sim) {
			t.Fatal("a shift into Moonkin Form rolled Omen of Clarity")
		}
		stepUntil(sim, start+10*time.Second)
	}
}
