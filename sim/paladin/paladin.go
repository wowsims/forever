package paladin

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{17, 16, 17}

type Paladin struct {
	core.Character

	Talents *proto.PaladinTalents

	Forbearance       *core.Aura
	RighteousFuryAura *core.Aura

	// Swift Judgement's free next Judgement, which Judgement spends.
	swiftJudgementAura *core.Aura

	Judgement *core.Spell

	// The seal the paladin is under, and the ranks it has cast so far this iteration.
	currentSeal *sealConfig

	// The judgement effects this paladin can put on enemies; its melee strikes refresh them.
	JudgementAuras []core.AuraArray

	// Twist of Light: one Echo per seal that leaves one, keyed by the Echo's spell id.
	echoes map[int32]*sealEcho

	// The mark Consecration lays on its bonus targets, one aura per enemy, which carries Consecrated Ground.
	consecrationMarks core.AuraArray
	// Consecrated Ground: the talent rank's points for the mark's damage from caster mod; nil without it.
	consecratedGroundPoints *spelldata.Effect

	// Light's Vigil: one entry per rank, so Holy Shock can find the vigil it consumes.
	lightsVigils []*lightsVigil

	// What gear adds to numbers the spells read as they register. Item effects and set bonuses
	// apply before Initialize, so the spells pick these up.
	sealOfTheCrusaderBonusAttackPower float64
	judgementOfTheCrusaderBonus       float64
	flashOfLightBonusHealing          float64
	holyShieldBlockValueMultiplier    float64
	forbearanceReduction              time.Duration
}

// Implemented by each Paladin spec.
type PaladinAgent interface {
	GetPaladin() *Paladin
}

func (paladin *Paladin) GetCharacter() *core.Character {
	return &paladin.Character
}

func (paladin *Paladin) GetPaladin() *Paladin {
	return paladin
}

func (paladin *Paladin) AddRaidBuffs(_ *proto.RaidBuffs) {
}

func (paladin *Paladin) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (paladin *Paladin) Initialize() {
	paladin.registerForbearance()
	paladin.registerRighteousFury()

	paladin.registerJudgement()
	paladin.registerSeals()
	paladin.registerAuras()

	HolyStrikeRankMap.Each(paladin.registerHolyStrike)
	paladin.registerConsecrationMark()
	ConsecrationRankMap.Each(paladin.registerConsecration)
	ExorcismRankMap.Each(paladin.registerExorcism)
	HammerOfWrathRankMap.Each(paladin.registerHammerOfWrath)
	HolyWrathRankMap.Each(paladin.registerHolyWrath)

	HolyLightRankMap.Each(paladin.registerHolyLight)
	FlashOfLightRankMap.Each(paladin.registerFlashOfLight)
	LayOnHandsRankMap.Each(paladin.registerLayOnHands)
}

func (paladin *Paladin) Reset(_ *core.Simulation) {
	paladin.currentSeal = nil
	for _, echo := range paladin.echoes {
		echo.seal = nil
	}
}

func (paladin *Paladin) OnEncounterStart(_ *core.Simulation) {
}

func NewPaladin(character *core.Character, talentsStr string, _ *proto.PaladinOptions) *Paladin {
	paladin := &Paladin{
		Character: *character,
		Talents:   &proto.PaladinTalents{},

		holyShieldBlockValueMultiplier: 1,
	}

	core.FillTalentsProto(paladin.Talents.ProtoReflect(), talentsStr, TalentTreeSizes)

	paladin.PseudoStats.CanParry = true

	paladin.EnableManaBar()

	paladin.EnableAutoAttacks(paladin, core.AutoAttackOptions{
		MainHand:       paladin.WeaponFromMainHand(),
		AutoSwingMelee: true,
	})

	paladin.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	paladin.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])
	paladin.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	return paladin
}
