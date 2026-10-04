package hunter

import (
	"maps"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"

	"testing"
)

func init() {
	RegisterHunter()
	common.RegisterAllEffects()
}

func TestHunter(t *testing.T) {
	weaveRotation := core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default")
	weaveRotation.Label = "weave"

	turretRotation := core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default").Rotation
	stopWeaving(turretRotation)

	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassHunter,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceNightElf},
			GearSet:    core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1"),
			Talents:    DefaultMMTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "SV", Talents: DefaultSVTalents},
			},
			Consumables:      DefaultConsumables,
			SpecOptions:      core.SpecOptionsCombo{Label: "Default", SpecOptions: DefaultOptions},
			StartingDistance: 8,
			Profession1:      proto.Profession_Engineering,
			Profession2:      proto.Profession_Blacksmithing,

			Rotation: core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default"),

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeMail,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeBow,
					proto.RangedWeaponType_RangedWeaponTypeCrossbow,
					proto.RangedWeaponType_RangedWeaponTypeGun,
				},
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypePolearm,
					proto.WeaponType_WeaponTypeStaff,
					proto.WeaponType_WeaponTypeSword,
				},
				HandTypes: []proto.HandType{
					proto.HandType_HandTypeMainHand,
					proto.HandType_HandTypeOffHand,
					proto.HandType_HandTypeOneHand,
					proto.HandType_HandTypeTwoHand,
				},
			},
		},
	}))
}

var DefaultOptions = &proto.Player_Hunter{
	Hunter: &proto.Hunter{
		Options: &proto.Hunter_Options{
			ClassOptions: &proto.HunterOptions{
				Ammo:        proto.HunterOptions_ThoriumHeadedArrow,
				QuiverBonus: proto.HunterOptions_AncientSinewWrappedLamina,
				PetType:     proto.HunterOptions_PetNone,
			},
		},
	},
}

// Every pet family, fully trained, on one race, talent set and rotation: the pet's own suite, so
// that each family's abilities run without multiplying the hunter's. The talents are Beast Mastery,
// the pet's tree, and the rotation sends two hawks for each Arcane Shot, the two spells sharing a cooldown.
func TestHunterPets(t *testing.T) {
	turretRotation := core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default")
	turretRotation.Label = "turret"
	stopWeaving(turretRotation.Rotation)
	setVariable(turretRotation.Rotation, "Use Summon Hawk", "true")

	families := petOptions()
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:            proto.Class_ClassHunter,
			Race:             proto.Race_RaceOrc,
			GearSet:          core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1"),
			Talents:          DefaultBMTalents,
			Consumables:      DefaultConsumables,
			SpecOptions:      families[0],
			OtherSpecOptions: families[1:],
			StartingDistance: 8,
			Rotation:         turretRotation,
			ItemFilter:       core.ItemFilter{ArmorType: proto.ArmorType_ArmorTypeMail},
		},
	}))
}

// Auto Shot needs the bow, gun or crossbow its row names: a thrown weapon in the ranged slot fires
// none and starts no windup, the p1 bow fires them.
func TestAutoShotNeedsABowGunOrCrossbow(t *testing.T) {
	const thrown = 20814 // Master's Throwing Dagger
	if item := core.GetItemByID(thrown); item == nil || item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeThrown {
		t.Fatalf("item %d is not a thrown weapon", thrown)
	}
	windup := "Casting " + core.ActionID{OtherID: proto.OtherAction_OtherActionShoot}.String()

	autoShots := func(rangedID int32) (casts int32, windups int) {
		gear := googleProto.Clone(core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1").GearSet).(*proto.EquipmentSpec)
		if rangedID != 0 {
			gear.Items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: rangedID}
		}
		player := &proto.Player{
			Name:               "Auto Shot",
			Race:               proto.Race_RaceOrc,
			Class:              proto.Class_ClassHunter,
			Equipment:          gear,
			TalentsString:      DefaultMMTalents,
			Spec:               DefaultOptions,
			Rotation:           &proto.APLRotation{},
			DistanceFromTarget: 20,
		}
		result := core.RunRaidSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, Debug: true},
		})
		if result.Error != nil {
			t.Fatalf("the sim failed: %s", result.Error.Message)
		}
		for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
			if action.Id.GetOtherId() == proto.OtherAction_OtherActionShoot {
				for _, target := range action.Targets {
					casts += target.Casts
				}
			}
		}
		return casts, strings.Count(result.Logs, windup)
	}

	if casts, windups := autoShots(0); casts == 0 || windups == 0 {
		t.Errorf("the p1 bow fired %d Auto Shots after %d windups, want some", casts, windups)
	}
	if casts, windups := autoShots(thrown); casts != 0 || windups != 0 {
		t.Errorf("a thrown weapon fired %d Auto Shots after %d windups, want none", casts, windups)
	}
}

// Sniper Shot's buff carries Auto Shot's reach past the 41 yards Hawk Eye gives it, for three Shots: a
// hunter at 45 yards fires no Auto Shot until Sniper Shot lands, then lands at most three per Sniper
// Shot, since a miss keeps its charge.
func TestSniperShotExtendsAutoShotRange(t *testing.T) {
	shots := func(castSniperShot bool) (sniperShots, autoShots, landed int32) {
		rotation := &proto.APLRotation{Type: proto.APLRotation_TypeAPL}
		if castSniperShot {
			rotation.PriorityList = []*proto.APLListItem{{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{
				CastSpell: &proto.APLActionCastSpell{SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: sniperShotRank.ID}}},
			}}}}
		}
		player := &proto.Player{
			Name:               "Sniper Shot",
			Race:               proto.Race_RaceOrc,
			Class:              proto.Class_ClassHunter,
			Equipment:          core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1").GearSet,
			TalentsString:      DefaultMMTalents,
			Spec:               DefaultOptions,
			Rotation:           rotation,
			DistanceFromTarget: 45,
		}
		result := core.RunRaidSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
		})
		if result.Error != nil {
			t.Fatalf("the sim failed: %s", result.Error.Message)
		}
		for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
			for _, target := range action.Targets {
				switch {
				case action.Id.GetOtherId() == proto.OtherAction_OtherActionShoot:
					autoShots += target.Casts
					landed += target.Hits + target.Crits
				case action.Id.GetSpellId() == sniperShotRank.ID:
					sniperShots += target.Casts
				}
			}
		}
		return sniperShots, autoShots, landed
	}

	if _, autoShots, _ := shots(false); autoShots != 0 {
		t.Errorf("a hunter at 45 yards fired %d Auto Shots with no Sniper Shot, want none", autoShots)
	}
	sniperShots, _, landed := shots(true)
	if sniperShots == 0 || landed == 0 || landed > 3*sniperShots {
		t.Errorf("%d Sniper Shots let %d Auto Shots land from 45 yards, want some and at most 3 each", sniperShots, landed)
	}
}

// One sting per hunter on a target: a hunter's sting drops that hunter's other sting and leaves
// another hunter's. Two hunters' Scorpid Stings both stay up, and their -2% hit counts once.
func TestEachHunterKeepsOneStingPerTarget(t *testing.T) {
	player := func(name string) *proto.Player {
		return &proto.Player{Name: name, Race: proto.Race_RaceOrc, Class: proto.Class_ClassHunter,
			Equipment: &proto.EquipmentSpec{}, Spec: DefaultOptions, Rotation: &proto.APLRotation{}}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: &proto.Raid{
			Parties: []*proto.Party{{Players: []*proto.Player{player("A"), player("B")}, Buffs: &proto.PartyBuffs{}}},
			Buffs:   &proto.RaidBuffs{},
			Debuffs: &proto.Debuffs{},
		},
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{},
	}, simsignals.CreateSignals())
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	other := sim.Raid.Parties[0].Players[1].(*Hunter)
	target := sim.Encounter.AllTargetUnits[0]

	serpent := hunter.GetSpell(core.ActionID{SpellID: serpentStingRank.ID}).Dot(target)
	scorpid := target.GetAura("Scorpid Sting-" + hunter.Label)
	otherScorpid := target.GetAura("Scorpid Sting-" + other.Label)

	// What a landed sting does: drop the hunter's other sting, then go up.
	dropOtherSting(sim, otherScorpid)
	otherScorpid.Activate(sim)
	dropOtherSting(sim, scorpid)
	scorpid.Activate(sim)
	dropOtherSting(sim, serpent.Aura)
	serpent.Apply(sim)
	if scorpid.IsActive() {
		t.Error("Serpent Sting landed beside the hunter's own Scorpid Sting, want it dropped")
	}

	dropOtherSting(sim, scorpid)
	scorpid.Activate(sim)
	if serpent.IsActive() {
		t.Error("Scorpid Sting landed beside the hunter's own Serpent Sting, want it dropped")
	}
	if !otherScorpid.IsActive() {
		t.Error("the hunter's stings dropped the other hunter's Scorpid Sting, want it on the target")
	}
	if got := target.GetStats()[stats.PhysicalHitPercent]; got != -2 {
		t.Errorf("two hunters' Scorpid Stings moved hit by %v, want -2 once", got)
	}
}

// The hunter's own Hunter's Mark carries the mark's tag, which Expose Prey reads before it opens
// Mongoose Bite, and gives every attacker its ranged attack power.
func TestTheHuntersOwnMarkCarriesTheMarkTag(t *testing.T) {
	player := &proto.Player{
		Name:          "Mark",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassHunter,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: DefaultSVTalents,
		Spec:          DefaultOptions,
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	target := sim.Encounter.AllTargetUnits[0]

	mark := hunter.GetSpell(core.ActionID{SpellID: huntersMarkRank.ID})
	if mark == nil {
		t.Fatal("the hunter has no Hunter's Mark")
	}
	mark.SkipCastAndApplyEffects(sim, target)
	if !target.HasActiveAuraWithTag(buffs.HuntersMarkCategory) {
		t.Error("the hunter's mark is up but no aura tagged HuntersMark is, so Expose Prey never sees it")
	}
	if got, want := target.PseudoStats.BonusRangedAttackPower, buffs.HuntersMarkValue(0); got != want {
		t.Errorf("the mark gives attackers %v ranged attack power, want %v", got, want)
	}
}

// One combo per pet family, fully trained.
// The default rotation with its "Melee weave" variable turned off: the hunter stays at range.
func stopWeaving(rotation *proto.APLRotation) {
	setVariable(rotation, "Melee weave", "false")
}

func setVariable(rotation *proto.APLRotation, name string, value string) {
	for _, variable := range rotation.ValueVariables {
		if variable.Name == name {
			variable.Value = &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: value}}}
		}
	}
}

func petOptions() []core.SpecOptionsCombo {
	var combos []core.SpecOptionsCombo
	for _, value := range slices.Sorted(maps.Values(proto.HunterOptions_PetType_value)) {
		petType := proto.HunterOptions_PetType(value)
		if petType == proto.HunterOptions_PetNone {
			continue
		}
		combos = append(combos, core.SpecOptionsCombo{
			Label: petType.String(),
			SpecOptions: &proto.Player_Hunter{
				Hunter: &proto.Hunter{
					Options: &proto.Hunter_Options{
						ClassOptions: &proto.HunterOptions{
							PetType:        petType,
							PetUptime:      1,
							CobraReflexes:  true,
							PetAggression:  5,
							PetAttackSpeed: proto.HunterOptions_FasterAttackII,
						},
					},
				},
			},
		})
	}
	return combos
}

var DefaultBMTalents = "5320001505101251-3050552"
var DefaultMMTalents = "-30535525115023051-5"
var DefaultSVTalents = "-00505505-500250030050220151"

var DefaultConsumables = &proto.ConsumesSpec{
	ElixirIds:   []int32{13452, 250341}, // Elixir of the Mongoose, Greater Mageblood Elixir
	FoodId:      250072,                 // Swiftstrike Steak
	PotId:       250943,                 // Major Frenzy Potion
	ConjuredId:  12662,                  // Demonic Rune
	ExplosiveId: 18588,                  // Ez-Thro Dynamite II
	// Goblin Sapper Charge, Scroll of Agility IV, Scroll of Strength IV
	ExtraIds: []int32{10646, 10309, 10310},
}
