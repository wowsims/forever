package enhancement

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/shaman"
)

func RegisterEnhancementShaman() {
	core.RegisterAgentFactory(
		proto.Player_EnhancementShaman{},
		proto.Spec_SpecEnhancementShaman,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewEnhancementShaman(character, options)
		},
		func(player *proto.Player, spec any) {
			playerSpec, ok := spec.(*proto.Player_EnhancementShaman)
			if !ok {
				panic("Invalid spec value for Enhancement Shaman!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewEnhancementShaman(character *core.Character, options *proto.Player) *EnhancementShaman {
	enhOptions := options.GetEnhancementShaman().Options

	selfBuffs := shaman.SelfBuffs{
		ShieldProcrate: enhOptions.ClassOptions.ShieldProcrate,
		ImbueMH:        enhOptions.ClassOptions.ImbueMh,
		ImbueMHSwap:    enhOptions.ClassOptions.ImbueMhSwap,
	}

	enh := &EnhancementShaman{
		Shaman: shaman.NewShaman(character, options.TalentsString, selfBuffs),
	}

	// Shamans do not dual wield, so only the main hand swings.
	enh.EnableAutoAttacks(enh, core.AutoAttackOptions{
		MainHand:       enh.WeaponFromMainHand(),
		AutoSwingMelee: true,
	})

	if !enh.HasMHWeapon() {
		enh.SelfBuffs.ImbueMH = proto.ShamanImbue_NoImbue
	}

	return enh
}

type EnhancementShaman struct {
	*shaman.Shaman
}

func (enh *EnhancementShaman) GetShaman() *shaman.Shaman {
	return enh.Shaman
}

// The melee crit cap counts Elemental Devastation's crit as taken. The buff is up for almost every
// swing, and the sheet shows its crit only while it is up.
func (enh *EnhancementShaman) ModifyMeleeCritCap(info *proto.MeleeCritCapInfo) {
	info.SpecOffset += enh.ElementalDevastationCritPercent()
}

func (enh *EnhancementShaman) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	enh.Shaman.AddRaidBuffs(raidBuffs)
}

func (enh *EnhancementShaman) ApplyTalents() {
	enh.Shaman.ApplyTalents()
}

func (enh *EnhancementShaman) Initialize() {
	enh.Shaman.Initialize()
}

func (enh *EnhancementShaman) Reset(sim *core.Simulation) {
	enh.Shaman.Reset(sim)
}
