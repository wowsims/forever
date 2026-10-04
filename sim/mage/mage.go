package mage

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{18, 17, 19}

type Mage struct {
	core.Character

	ClassSpellScaling float64

	Talents *proto.MageTalents
	Options *proto.MageOptions

	ArcaneChargesAura  *core.Aura
	ClearCasting       *core.Aura
	PresenceOfMindAura *core.Aura
	ArcanePowerAura    *core.Aura

	ImprovedScorchAuras core.AuraArray

	ArcaneBlast     *core.Spell
	Ignite          *core.Spell
	FireBlast       *core.Spell
	FlameOrbExplode *core.Spell
	Flamestrike     []*core.Spell
	FlamestrikeBW   *core.Spell
	FrostfireOrb    *core.Spell
	Pyroblast       *core.Spell
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
	// ImprovedScorchAuras is filled by the port of registerScorchSpell: the
	// game files' Fire Vulnerability (22959) is A_MOD_SCHOOL_MASK_DAMAGE_FROM_CASTER,
	// so it raises the fire damage this mage deals rather than the damage the
	// target takes: the port parses it onto each target's aura with
	// spelldata.ParseEffects and FromCaster.
	mage.registerPassives()
	mage.registerSpells()
}

func (mage *Mage) registerPassives() {
	mage.registerArcaneCharges()
}

func (mage *Mage) registerSpells() {
	mage.registerArcaneBlastSpell()
	mage.registerArcaneExplosionSpell()
	mage.registerArcaneMissilesSpell()
	mage.registerArmorSpells()
	mage.registerBlizzardSpell()
	mage.registerConeOfColdSpell()
	mage.registerFrostboltSpell()
	mage.registerEvocation()
	mage.registerFireballSpell()
	mage.registerFireBlastSpell()
	mage.registerFrostNovaSpell()
	mage.registerManaGems()
	mage.registerScorchSpell()

	FlameStrikeRankMap.Each(func(_ int32, rankConfig *spelldata.Spell) { mage.registerFlamestrike(rankConfig) })

	//TalentSpells
	mage.registerPresenceOfMindSpell()
	mage.registerArcanePowerSpell()

	mage.registerBlastWaveSpell()
	mage.registerPyroblastSpell()
	mage.registerCombustionSpell()

	mage.registerColdSnapSpell()
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

// Agent is a generic way to access underlying mage on any of the agents.
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
	MageSpellBlastWave
	MageSpellBlizzard
	MageSpellColdSnap
	MageSpellConeOfCold
	MageSpellEvocation
	MageSpellFireBlast
	MageSpellFireball
	MageSpellFlamestrike
	MageSpellFlamestrikeDot
	MageSpellFrostArmor
	MageSpellFrostbolt
	MageSpellFrostNova
	MageSpellIceBarrier
	MageSpellIceBlock
	MageSpellIceLance
	MageSpellIgnite
	MageSpellMageArmor
	MageSpellManaGems
	MageSpellMoltenArmor
	MageSpellPresenceOfMind
	MageSpellPyroblast
	MageSpellPyroblastDot
	MageSpellScorch
	MageSpellManaGem
	MageSpellCombustion

	// TODO: Forever abilities the sim does not model yet; see the stub file named for each.
	MageSpellFrostfireBolt

	MageSpellLast
	MageSpellsAll  = MageSpellLast<<1 - 1
	MageSpellFrost = MageSpellFrostbolt | MageSpellBlizzard | MageSpellFrostNova | MageSpellConeOfCold | MageSpellIceLance
	MageSpellFire  = MageSpellFireball | MageSpellCombustion |
		MageSpellFireBlast | MageSpellFlamestrike | MageSpellIgnite | MageSpellPyroblast | MageSpellScorch
	MageSpellsAllDamaging = MageSpellArcaneBlast | MageSpellArcaneExplosion | MageSpellArcaneMissilesTick | MageSpellBlizzard |
		MageSpellFireBlast | MageSpellFireball | MageSpellFlamestrike | MageSpellFrostbolt |
		MageSpellIceLance | MageSpellPyroblast | MageSpellPyroblastDot | MageSpellScorch
	MageSpellInstantCast = MageSpellArcaneMissilesCast | MageSpellArcaneMissilesTick | MageSpellFireBlast | MageSpellArcaneExplosion | MageSpellPyroblastDot |
		MageSpellCombustion | MageSpellConeOfCold | MageSpellIceLance | MageSpellManaGems | MageSpellPresenceOfMind
	MageSpellExtraResult = MageSpellArcaneMissilesTick | MageSpellBlizzard
	FireSpellIgnitable   = MageSpellFireball | MageSpellScorch | MageSpellPyroblast
)
