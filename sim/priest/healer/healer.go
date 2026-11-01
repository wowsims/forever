package healer

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/priest"
)

func RegisterHealerPriest() {
	core.RegisterAgentFactory(
		proto.Player_HealerPriest{},
		proto.Spec_SpecHealerPriest,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewHealerPriest(character, options)
		},
		func(player *proto.Player, spec any) {
			playerSpec, ok := spec.(*proto.Player_HealerPriest)
			if !ok {
				panic("Invalid spec value for Healer Priest!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewHealerPriest(character *core.Character, options *proto.Player) *HealerPriest {
	selfBuffs := priest.SelfBuffs{
		Armor: options.GetHealerPriest().GetOptions().GetClassOptions().GetArmor(),
	}

	basePriest := priest.New(character, selfBuffs, options.TalentsString)
	healerOptions := options.GetHealerPriest().GetOptions()
	basePriest.PrayerOfHealingMode = healerOptions.GetPrayerOfHealingMode()
	basePriest.PrayerOfMendingJumps = healerOptions.GetPrayerOfMendingJumps()
	basePriest.PrayerOfMendingDelay = 2.5
	if delay := healerOptions.GetPrayerOfMendingDelay(); delay > 0 {
		basePriest.PrayerOfMendingDelay = delay
	}
	basePriest.HolyNovaHealTargets = healerOptions.GetHolyNovaHealTargets()
	if basePriest.HolyNovaHealTargets == 0 {
		basePriest.HolyNovaHealTargets = 5
	}

	return &HealerPriest{Priest: basePriest}
}

type HealerPriest struct {
	*priest.Priest
}

func (healer *HealerPriest) GetPriest() *priest.Priest {
	return healer.Priest
}

func (healer *HealerPriest) Initialize() {
	healer.Priest.Initialize()
	healer.Priest.InitializeHealingSpells()
}

func (healer *HealerPriest) ApplyTalents() {
	healer.Priest.ApplyTalents()
}

func (healer *HealerPriest) Reset(sim *core.Simulation) {
	healer.Priest.Reset(sim)
}
