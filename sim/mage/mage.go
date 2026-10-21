package mage

import (
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{19, 17, 19}

type Mage struct {
	core.Character

	ClassSpellScaling float64

	Talents *proto.MageTalents
	Options *proto.MageOptions

	ArcaneChargesAura  *core.Aura
	ClearCasting       *core.Aura
	PresenceOfMindAura *core.Aura
	ArcanePowerAura    *core.Aura
	FingersOfFrostAura *core.Aura

	ImprovedScorchAuras core.AuraArray
	WintersChillAuras   core.AuraArray

	Ignite *core.Spell
}

func (mage *Mage) GetCharacter() *core.Character {
	return &mage.Character
}

func (mage *Mage) GetMage() *Mage {
	return mage
}

func RegisterMage() {
	core.RegisterAgentFactory(
		proto.Player_Mage{},
		proto.Spec_SpecMage,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewMage(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Mage)
			if !ok {
				panic("Invalid spec value for Mage!")
			}
			player.Spec = playerSpec
		},
	)
}

func (mage *Mage) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	raidBuffs.ArcaneBrilliance = true
}

func (mage *Mage) AddPartyBuffs(partyBuffs *proto.PartyBuffs) {
}

func (mage *Mage) Initialize() {
	mage.registerPassives()
	mage.registerSpells()
}

func (mage *Mage) registerPassives() {
	mage.registerArcaneCharges()
}

func (mage *Mage) registerSpells() {
	mage.registerArcaneExplosionSpell()
	mage.registerArcaneMissilesSpell()
	mage.registerArmorSpells()
	mage.registerBlizzardSpell()
	mage.registerConeOfColdSpell()
	mage.registerFrostboltSpell()
	mage.registerFrostBlastSpell()
	mage.registerEvocation()
	mage.registerFireballSpell()
	mage.registerFireBlastSpell()
	mage.registerFlamestrikeSpells()
	mage.registerFrostfireBoltSpell()
	mage.registerFrostNovaSpell()
	mage.registerManaGems()
	mage.registerScorchSpell()
	mage.registerDefensiveSpells()
	shared.RegisterShoot(&mage.Character)
}

// A damage effect's amount per target, for the AoE damage calls; core spreads a spell-school hit
// itself (DirectVariance).
func rollDamage(e *spelldata.Effect) core.BaseDamageCalculator {
	return func(sim *core.Simulation, _ *core.Spell) float64 {
		return e.DirectAmount(sim)
	}
}

func (mage *Mage) Reset(sim *core.Simulation) {
}

func (mage *Mage) OnEncounterStart(sim *core.Simulation) {
}

func NewMage(character *core.Character, options *proto.Player) *Mage {
	mageOptions := options.GetMage().Options.ClassOptions
	mage := &Mage{
		Character: *character,
		Talents:   &proto.MageTalents{},
		Options:   mageOptions,
	}

	core.FillTalentsProto(mage.Talents.ProtoReflect(), options.TalentsString, TalentTreeSizes)

	mage.EnableManaBar()
	mage.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	mage.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])

	// TODO: Forever drops Summon Water Elemental; the pet is never created until we know
	// whether the talent moved elsewhere.

	return mage
}

// The slots on the mage's attack table against a target for what raises only its own attacks there:
// one per FromCaster parse, and Fire Power.
const (
	DDBC_FireVulnerability int = iota
	DDBC_WintersChill
	DDBC_FirePower

	DDBC_Total
)

// The underlying mage of any of the mage agents.
type MageAgent interface {
	GetMage() *Mage
}

const (
	FireSpellMaxTimeUntilResult       = 750 * time.Millisecond
	MageSpellFlagNone           int64 = 0
	MageSpellArcaneBlast        int64 = 1 << iota
	MageSpellArcaneExplosion
	MageSpellArcanePower
	MageSpellArcaneMissilesCast
	MageSpellArcaneMissilesTick
	MageSpellArcaneWrath
	MageSpellBlastWave
	MageSpellBlizzard
	MageSpellColdSnap
	MageSpellConeOfCold
	MageSpellEvocation
	MageSpellFireBlast
	MageSpellFireball
	MageSpellFireballDot
	MageSpellFlamestrike
	MageSpellFlamestrikeDot
	MageSpellFrostArmor
	MageSpellIceArmor
	MageSpellFrostbolt
	MageSpellFrostBlast
	MageSpellFrostNova
	MageSpellIceBarrier
	MageSpellIceBlock
	MageSpellIceLance
	MageSpellIgnite
	MageSpellMageArmor
	MageSpellManaGems
	MageSpellPresenceOfMind
	MageSpellPyroblast
	MageSpellPyroblastDot
	MageSpellScorch
	MageSpellManaGem
	MageSpellCombustion

	MageSpellFrostfireBolt
	MageSpellFrostfireBoltDot

	MageSpellLast
	MageSpellsAll  = MageSpellLast<<1 - 1
	MageSpellFrost = MageSpellFrostbolt | MageSpellFrostBlast | MageSpellBlizzard | MageSpellFrostNova | MageSpellConeOfCold | MageSpellIceLance
	MageSpellFire  = MageSpellFireball | MageSpellCombustion |
		MageSpellFireBlast | MageSpellFlamestrike | MageSpellIgnite | MageSpellPyroblast | MageSpellScorch
	MageSpellsAllDamaging = MageSpellArcaneBlast | MageSpellArcaneExplosion | MageSpellArcaneMissilesTick | MageSpellArcaneWrath | MageSpellBlizzard |
		MageSpellBlastWave | MageSpellConeOfCold | MageSpellFireBlast | MageSpellFireball | MageSpellFireballDot |
		MageSpellFlamestrike | MageSpellFlamestrikeDot | MageSpellFrostbolt | MageSpellFrostBlast | MageSpellFrostfireBolt | MageSpellFrostfireBoltDot | MageSpellFrostNova |
		MageSpellIceLance | MageSpellPyroblast | MageSpellPyroblastDot | MageSpellScorch
	MageSpellInstantCast = MageSpellArcaneMissilesCast | MageSpellArcaneMissilesTick | MageSpellFireBlast | MageSpellArcaneExplosion | MageSpellPyroblastDot | MageSpellFireballDot |
		MageSpellCombustion | MageSpellConeOfCold | MageSpellIceLance | MageSpellManaGems | MageSpellPresenceOfMind
	MageSpellExtraResult = MageSpellArcaneMissilesTick | MageSpellBlizzard
)
