package core_test

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
	"github.com/wowsims/forever/sim/hunter"
	"github.com/wowsims/forever/sim/mage"
	"github.com/wowsims/forever/sim/paladin"
	"github.com/wowsims/forever/sim/priest"
	"github.com/wowsims/forever/sim/rogue"
	"github.com/wowsims/forever/sim/shaman"
	"github.com/wowsims/forever/sim/warlock"
	"github.com/wowsims/forever/sim/warrior"
)

func nakedPlayer(class proto.Class, race proto.Race) *proto.Player {
	player := &proto.Player{Class: class, Race: race, Equipment: &proto.EquipmentSpec{}}
	switch class {
	case proto.Class_ClassWarrior:
		player.Spec = &proto.Player_DpsWarrior{}
	case proto.Class_ClassPaladin:
		player.Spec = &proto.Player_RetributionPaladin{}
	case proto.Class_ClassPriest:
		player.Spec = &proto.Player_DpsPriest{}
	case proto.Class_ClassShaman:
		player.Spec = &proto.Player_ElementalShaman{}
	case proto.Class_ClassMage:
		player.Spec = &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{ClassOptions: &proto.MageOptions{}}}}
	case proto.Class_ClassWarlock:
		player.Spec = &proto.Player_Warlock{}
	case proto.Class_ClassDruid:
		player.Spec = &proto.Player_BalanceDruid{}
	case proto.Class_ClassRogue:
		player.Spec = &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.Rogue_Options{ClassOptions: &proto.RogueOptions{}}}}
	case proto.Class_ClassHunter:
		player.Spec = &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{ClassOptions: &proto.HunterOptions{}}}}
	}
	return player
}

func nakedSheet(t *testing.T, class proto.Class, race proto.Race) (*core.Character, stats.Stats) {
	t.Helper()
	player := nakedPlayer(class, race)
	raid := &proto.Raid{}
	party := &core.Party{Raid: &core.Raid{}}
	character := core.NewCharacter(party, 0, player)

	var built *core.Character
	switch class {
	case proto.Class_ClassWarrior:
		built = &warrior.NewWarrior(&character, &proto.WarriorOptions{}, "", warrior.WarriorInputs{}).Character
	case proto.Class_ClassPaladin:
		built = &paladin.NewPaladin(&character, "", nil).Character
	case proto.Class_ClassPriest:
		built = &priest.New(&character, priest.SelfBuffs{}, "").Character
	case proto.Class_ClassShaman:
		built = &shaman.NewShaman(&character, "", shaman.SelfBuffs{}).Character
	case proto.Class_ClassMage:
		built = &mage.NewMage(&character, player).Character
	case proto.Class_ClassWarlock:
		built = &warlock.NewWarlock(&character, player, &proto.WarlockOptions{}, raid).Character
	case proto.Class_ClassDruid:
		built = &druid.New(&character, druid.Humanoid, druid.SelfBuffs{}, "").Character
	case proto.Class_ClassRogue:
		built = &rogue.NewRogue(&character, player, "").Character
	case proto.Class_ClassHunter:
		built = &hunter.NewHunter(&character, player, player.GetHunter().Options.ClassOptions, raid).Character
	default:
		t.Fatalf("no constructor for %v", class)
	}

	built.FinalizeStatDeps()
	return built, built.ApplyStatDependencies(built.GetStats()).FloorGameStats()
}

func TestPetsGetTenHealthPerStaminaFromTheFirstPoint(t *testing.T) {
	owner := core.NewCharacter(&core.Party{Raid: &core.Raid{}}, 0, nakedPlayer(proto.Class_ClassWarrior, proto.Race_RaceHuman))
	ownerBase := core.BaseStats[core.BaseStatsKey{Race: proto.Race_RaceHuman, Class: proto.Class_ClassWarrior}]
	if got, want := owner.GetStats()[stats.Health], ownerBase[stats.Health]-180; got != want {
		t.Errorf("player health before stamina %v, want %v: the first 20 stamina give 1 health each", got, want)
	}

	pet := core.NewPet(core.PetConfig{
		Name:      "Test Pet",
		Owner:     &owner,
		BaseStats: stats.Stats{stats.Health: 1000, stats.Stamina: 50},
	})
	if got := pet.GetStats()[stats.Health]; got != 1000 {
		t.Errorf("pet health before stamina %v, want its base 1000", got)
	}
	pet.FinalizeStatDeps()
	if got := pet.ApplyStatDependencies(pet.GetStats())[stats.Health]; got != 1000+50*10 {
		t.Errorf("pet health %v, want %v", got, 1000+50*10)
	}
}

func TestNakedWarriorHealth(t *testing.T) {
	_, sheet := nakedSheet(t, proto.Class_ClassWarrior, proto.Race_RaceHuman)

	stamina := sheet[stats.Stamina]
	if stamina != 110 {
		t.Fatalf("human warrior stamina %v, want the class row's 110", stamina)
	}
	if want := 1689 + 20 + (stamina-20)*10; sheet[stats.Health] != want {
		t.Errorf("health %v, want %v: 10 per stamina beyond the first 20", sheet[stats.Health], want)
	}
}

func TestNakedAttackPowerPerClass(t *testing.T) {
	for _, tc := range []struct {
		class       proto.Class
		race        proto.Race
		attackPower func(s stats.Stats) float64
	}{
		{class: proto.Class_ClassWarrior, race: proto.Race_RaceHuman, attackPower: func(s stats.Stats) float64 { return 160 + 2*s[stats.Strength] }},
		{class: proto.Class_ClassPaladin, race: proto.Race_RaceHuman, attackPower: func(s stats.Stats) float64 { return 160 + 2*s[stats.Strength] }},
		{class: proto.Class_ClassPriest, race: proto.Race_RaceHuman, attackPower: func(s stats.Stats) float64 { return -10 + s[stats.Strength] }},
		{class: proto.Class_ClassShaman, race: proto.Race_RaceOrc, attackPower: func(s stats.Stats) float64 { return 100 + 2*s[stats.Strength] }},
		{class: proto.Class_ClassMage, race: proto.Race_RaceHuman, attackPower: func(s stats.Stats) float64 { return -10 + s[stats.Strength] }},
		{class: proto.Class_ClassWarlock, race: proto.Race_RaceHuman, attackPower: func(s stats.Stats) float64 { return -10 + s[stats.Strength] }},
		{class: proto.Class_ClassDruid, race: proto.Race_RaceTauren, attackPower: func(s stats.Stats) float64 { return -20 + 2*s[stats.Strength] }},
	} {
		t.Run(tc.class.String(), func(t *testing.T) {
			_, sheet := nakedSheet(t, tc.class, tc.race)
			if want := tc.attackPower(sheet); math.Abs(sheet[stats.AttackPower]-want) > 1e-9 {
				t.Errorf("attack power %v, want %v", sheet[stats.AttackPower], want)
			}
		})
	}
}
