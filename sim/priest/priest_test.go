package priest

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	RegisterPriest()
	common.RegisterAllEffects()
}

// The Shadow Testing build's gear, talents and rotation, which the UI's Shadow Testing gear set,
// Standard talents and Default rotation read too.
var shadowTestingBuild = core.GetTestBuildFromJSON(proto.Class_ClassPriest, "../../ui/specs/priest/dps/builds", "testing",
	core.ItemFilter{}, nil, nil)

// The UI's defaults: Flask of Supreme Power, Major Spellblasting Potion, Nightfin Soup, Demonic Rune,
// Brilliant Wizard Oil and Ez-Thro Dynamite II.
var defaultConsumables = &proto.ConsumesSpec{
	FlaskId:        13512,
	OffensivePotId: 250937,
	FoodId:         13931,
	ConjuredId:     12662,
	MhImbueId:      20749,
	ExplosiveId:    18588,
}

func TestShadowPriest(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassPriest,
			Race:       proto.Race_RaceTroll,
			OtherRaces: []proto.Race{proto.Race_RaceUndead, proto.Race_RaceDwarf, proto.Race_RaceNightElf},

			SpecOptions: core.SpecOptionsCombo{
				Label: "Shadow",
				SpecOptions: &proto.Player_DpsPriest{
					DpsPriest: &proto.DpsPriest{
						Options: &proto.DpsPriest_Options{
							ClassOptions: &proto.PriestOptions{
								// Begin the sim already in Shadowform so the opener
								// doesn't spend a GCD casting it.
								PreShadowform: true,
								Armor:         proto.PriestOptions_InnerFire,
							},
						},
					},
				},
			},

			GearSet:     shadowTestingBuild.GearSet,
			Talents:     shadowTestingBuild.Talents,
			Consumables: defaultConsumables,
			Profession1: proto.Profession_Enchanting,
			Profession2: proto.Profession_Tailoring,

			Rotation: core.RotationCombo{Label: "default", Rotation: shadowTestingBuild.Rotation.Rotation},

			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeStaff,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeOffHand,
				},
				ArmorType: proto.ArmorType_ArmorTypeCloth,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeWand,
				},
				// Blacklist melee enchants that appear on cloth-relevant slots but
				// are never used by casters.
				EnchantBlacklist: []int32{2673, 3225, 3273},
			},
		},
	}))
}

// The priest's own Prayer of Fortitude bids in the raid buff's category. Alone, the cast adds the
// row's stamina for the row's mana; Arcane Brilliance lifts the priest's mana pool above the cost. With the raid buff up, the cast is turned away, and the priest's
// copy, put up anyway as another priest's cast would, adds nothing.
func TestPrayerOfFortitudeCountsOnceWithTheRaidBuff(t *testing.T) {
	rank := spellData.PrayerOfFortitude.Highest()
	rowStamina := rank.EffectN(1).BaseValue()

	newSim := func(t *testing.T, raidBuffs *proto.RaidBuffs) (*core.Simulation, *Priest, *core.Spell) {
		t.Helper()
		player := &proto.Player{
			Name:      "Prayer",
			Race:      proto.Race_RaceTroll,
			Class:     proto.Class_ClassPriest,
			Equipment: shadowTestingBuild.GearSet.GearSet,
			Spec: &proto.Player_DpsPriest{
				DpsPriest: &proto.DpsPriest{
					Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}},
				},
			},
			Rotation: &proto.APLRotation{},
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()

		priest := sim.Raid.Parties[0].Players[0].(*Priest)
		spell := priest.GetSpell(core.ActionID{SpellID: rank.ID})
		if spell == nil {
			t.Fatalf("the priest has no %s %d", rank.Name, rank.ID)
		}
		return sim, priest, spell
	}

	t.Run("alone the cast adds the row's stamina for the row's mana", func(t *testing.T) {
		sim, priest, spell := newSim(t, &proto.RaidBuffs{ArcaneBrilliance: true})
		stamina, mana := priest.GetStats()[stats.Stamina], priest.CurrentMana()

		if !spell.Cast(sim, &priest.Unit) {
			t.Fatal("the priest could not cast Prayer of Fortitude with nothing in its category")
		}
		if got := priest.GetStats()[stats.Stamina] - stamina; got != rowStamina {
			t.Errorf("the cast added %v stamina, want the game files' %v", got, rowStamina)
		}
		if got := mana - priest.CurrentMana(); got != rank.Cost() {
			t.Errorf("the cast spent %v mana, want the game files' %v", got, rank.Cost())
		}
	})

	t.Run("the raid buff holds the category", func(t *testing.T) {
		sim, priest, spell := newSim(t, &proto.RaidBuffs{PrayerOfFortitude: true})
		stamina := priest.GetStats()[stats.Stamina]

		if spell.Cast(sim, &priest.Unit) {
			t.Error("the priest cast Prayer of Fortitude while the raid buff holds its category")
		}
		own := priest.GetAura("Prayer of Fortitude (Player)")
		if own == nil {
			t.Fatal("the priest carries no aura labelled \"Prayer of Fortitude (Player)\"")
		}
		own.Activate(sim)
		if got := priest.GetStats()[stats.Stamina]; got != stamina {
			t.Errorf("stamina is %v with both copies up, want the raid buff's %v counted once", got, stamina)
		}
	})
}

// Spirit Tap hears a credited kill and needs Vampiric Embrace on the dying target. A death without
// the kill, and a kill of a target the priest has not afflicted, leave the buff off. The talent
// at full rank rolls 100%.
func TestSpiritTapProcsOnACreditedKillOfAnAfflictedTarget(t *testing.T) {
	player := &proto.Player{
		Name:      "Spirit Tap",
		Race:      proto.Race_RaceTroll,
		Class:     proto.Class_ClassPriest,
		Equipment: shadowTestingBuild.GearSet.GearSet,
		// Spirit Tap 5/5 and Vampiric Embrace in the Shadow tree.
		TalentsString: "--00500000001",
		Spec: &proto.Player_DpsPriest{
			DpsPriest: &proto.DpsPriest{
				Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}},
			},
		},
		Rotation: &proto.APLRotation{},
	}
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Targets = append(encounter.Targets, core.FreshDefaultTargetConfig())

	newSim := func(t *testing.T) (*core.Simulation, *Priest, *core.Aura) {
		t.Helper()
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  encounter,
			SimOptions: &proto.SimOptions{RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()

		priest := sim.Raid.Parties[0].Players[0].(*Priest)
		if priest.Talents.SpiritTap != 5 || priest.VampiricEmbrace == nil {
			t.Fatalf("the build has Spirit Tap %d and Vampiric Embrace %v, want 5 and the spell",
				priest.Talents.SpiritTap, priest.VampiricEmbrace)
		}
		buff := priest.GetAura(spellData.SpiritTapTriggered.Highest().Name)
		if buff == nil {
			t.Fatal("the priest carries no Spirit Tap buff")
		}
		return sim, priest, buff
	}
	afflict := func(sim *core.Simulation, priest *Priest, target *core.Target) {
		for _, auras := range priest.VampiricEmbrace.RelatedAuraArrays {
			auras.Get(&target.Unit).Activate(sim)
		}
	}

	t.Run("a credited kill of an afflicted target", func(t *testing.T) {
		sim, priest, buff := newSim(t)
		target := sim.Encounter.AllTargets[0]
		afflict(sim, priest, target)
		target.Die(sim, 1)
		if !buff.IsActive() {
			t.Error("Spirit Tap is off after a credited kill of a target with Vampiric Embrace")
		}
	})
	t.Run("a death without the kill", func(t *testing.T) {
		sim, priest, buff := newSim(t)
		target := sim.Encounter.AllTargets[0]
		afflict(sim, priest, target)
		target.Die(sim, 0)
		if buff.IsActive() {
			t.Error("Spirit Tap is on after a death the raid got no kill for")
		}
	})
	t.Run("a kill of a target without Vampiric Embrace", func(t *testing.T) {
		sim, priest, buff := newSim(t)
		afflict(sim, priest, sim.Encounter.AllTargets[1])
		sim.Encounter.AllTargets[0].Die(sim, 1)
		if buff.IsActive() {
			t.Error("Spirit Tap is on after a kill of a target the priest has not afflicted")
		}
	})
}

// Early Demise reads the health of each Shadow Word: Death's own target: the crit bonus is on for a cast
// at a target below its threshold and off for one at a target above it.
func TestEarlyDemiseFollowsTheCastsTarget(t *testing.T) {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Targets = append(encounter.Targets, core.FreshDefaultTargetConfig())
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name:      "Early Demise",
			Race:      proto.Race_RaceTroll,
			Class:     proto.Class_ClassPriest,
			Equipment: &proto.EquipmentSpec{},
			// Early Demise 2/2 in the Shadow tree.
			TalentsString: "--000000000000002",
			Spec: &proto.Player_DpsPriest{
				DpsPriest: &proto.DpsPriest{
					Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}},
				},
			},
			Rotation: &proto.APLRotation{},
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	priest := sim.Raid.Parties[0].Players[0].(*Priest)
	if priest.Talents.EarlyDemise != 2 {
		t.Fatalf("the build has Early Demise %d, want 2", priest.Talents.EarlyDemise)
	}
	rank := spellData.ShadowWordDeath.Highest()
	spell := priest.GetSpell(core.ActionID{SpellID: rank.ID})
	if spell == nil {
		t.Fatalf("the priest has no %s %d", rank.Name, rank.ID)
	}
	full, low := sim.Encounter.AllTargetUnits[0], sim.Encounter.AllTargets[1]
	low.SetHealthCurve(sim, 0, sim.Duration, core.HealthKnot{At: 0, Percent: 15}, core.HealthKnot{At: 1, Percent: 0})
	base := spell.BonusCritPercent
	bonus := spellData.EarlyDemise.EffectAt(1).ValueAt(2)

	for _, c := range []struct {
		target *core.Unit
		want   float64
	}{
		{full, 0},
		{&low.Unit, bonus},
		{full, 0},
	} {
		spell.SkipCastAndApplyEffects(sim, c.target)
		if got := spell.BonusCritPercent - base; math.Abs(got-c.want) > 1e-9 {
			t.Errorf("Shadow Word: Death at %s has %v bonus crit, want %v", c.target.Label, got, c.want)
		}
	}
}
