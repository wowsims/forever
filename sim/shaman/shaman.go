package shaman

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{16, 18, 16}

const (
	SpellFlagShamanSpell = core.SpellFlagAgentReserved1
	SpellFlagShock       = core.SpellFlagAgentReserved2
	SpellFlagInstant     = core.SpellFlagAgentReserved3
	SpellFlagFocusable   = core.SpellFlagAgentReserved4
)

func NewShaman(character *core.Character, talents string, selfBuffs SelfBuffs) *Shaman {
	shaman := &Shaman{
		Character: *character,
		Talents:   &proto.ShamanTalents{},
		Totems:    &proto.ShamanTotems{},
		SelfBuffs: selfBuffs,
	}

	core.FillTalentsProto(shaman.Talents.ProtoReflect(), talents, TalentTreeSizes)

	// Add Shaman stat dependencies
	shaman.AddStatDependency(stats.BonusArmor, stats.Armor, 1)
	shaman.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[shaman.Class])
	shaman.EnableManaBarWithModifier()

	shaman.AddStatDependency(stats.Strength, stats.AttackPower, 2.0)

	shaman.windfuryWeaponMultiplier = 1
	shaman.rockbiterWeaponMultiplier = 1

	return shaman
}

// Which buffs this shaman is using.
type SelfBuffs struct {
	ShieldProcrate float64
	ImbueMH        proto.ShamanImbue
	ImbueMHSwap    proto.ShamanImbue
}

// Indexes into NextTotemDrops for self buffs
const (
	AirTotem int = iota
	EarthTotem
	FireTotem
	WaterTotem
)

// Shaman represents a shaman character.
type Shaman struct {
	core.Character

	// Elemental Weapons' share of the imbues' attack power, and the threat Spirit Weapons adds to
	// Rockbiter Weapon's effect 2.
	windfuryWeaponMultiplier     float64
	rockbiterWeaponMultiplier    float64
	spiritWeaponsRockbiterThreat float64

	// Windfury Weapon's extra strikes, which spend no Flurry charge.
	windfuryWeaponStrike *core.Spell

	// The equipped main hand and the swap set's, which the imbue options name.
	mainHandID     int32
	swapMainHandID int32

	// The imbue on each of those main hands as the sim runs: the options set them at reset, and an
	// imbue's cast replaces the one on the main hand in use.
	imbues        [2]proto.ShamanImbue
	imbueAuras    [proto.ShamanImbue_RockbiterWeapon + 1]*core.Aura // Active while the main hand in use carries that imbue.
	onImbueChange []func(sim *core.Simulation)

	Talents   *proto.ShamanTalents
	SelfBuffs SelfBuffs

	Totems *proto.ShamanTotems

	// The expiration time of each totem (earth, air, fire, water).
	TotemExpirations [4]time.Duration

	LavaBurst []*core.Spell // One per rank in castableRanks.

	Stormstrike *core.Spell

	LightningShield     *core.Spell
	LightningShieldAura *core.Aura
	WaterShieldAura     *core.Aura
	// Spends a charge of the shield that is up the way a hit taken does, for the shield proc rate option.
	ShieldSelfProc func(sim *core.Simulation)

	EarthShock []*core.Spell // One per rank in castableRanks.
	FlameShock []*core.Spell
	FrostShock []*core.Spell

	StormstrikeDebuffAuras core.AuraArray

	MagmaTotem   *core.Spell
	SearingTotem *core.Spell
	FireNova     *core.Spell

	WindfuryTotem    *core.Spell
	FlametongueTotem *core.Spell
	// Always nil: its registrar is commented out in totems.go, upstream of the fork.
	HealingStreamTotem *core.Spell

	// The aura of the totem standing in each element's slot, indexed like TotemExpirations.
	TotemAuras [4]*core.Aura
}

// Implemented by each Shaman spec.
type ShamanAgent interface {
	core.Agent

	// The Shaman controlled by this Agent.
	GetShaman() *Shaman
}

func (shaman *Shaman) GetCharacter() *core.Character {
	return &shaman.Character
}

func (shaman *Shaman) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
}

// The shocks, Lightning Bolt, Chain Lightning and Lava Burst register every rank in castableRanks; the
// other spells register their highest rank only.
func (shaman *Shaman) Initialize() {
	shaman.registerChainLightningSpell()
	shaman.registerLightningBoltSpell()
	shaman.registerShieldsSpells()
	shaman.registerMagmaTotemSpell()
	shaman.registerSearingTotemSpell()
	shaman.registerFireNovaSpell()
	shaman.registerWindfuryTotemSpell()
	shaman.registerFlametongueTotemSpell()
	shaman.registerFrostResistanceTotemSpell()
	shaman.registerStrengthOfEarthTotemSpell()
	shaman.registerGraceOfAirTotemSpell()
	shaman.registerManaSpringTotemSpell()
	shaman.registerShocks()

	// In Initialize so that the imbues read the multipliers the talents set.
	shaman.registerImbues()
}

func (shaman *Shaman) ApplyTalents() {
	shaman.registerElementalTalents()
	shaman.registerEnhancementTalents()
	shaman.registerRestorationTalents()
}

func (shaman *Shaman) Reset(sim *core.Simulation) {
	shaman.TotemExpirations[FireTotem] = -10 * time.Hour
	shaman.TotemExpirations[AirTotem] = -10 * time.Hour
	shaman.TotemExpirations[EarthTotem] = -10 * time.Hour
	shaman.TotemExpirations[WaterTotem] = -10 * time.Hour

	shaman.imbues = [2]proto.ShamanImbue{shaman.SelfBuffs.ImbueMH, shaman.SelfBuffs.ImbueMHSwap}
	shaman.applyImbues(sim)
}

func (shaman *Shaman) OnEncounterStart(sim *core.Simulation) {
	shaman.startShieldProcPeriodicAction(sim)
}

// The chance is the talent's ranked dummy, 3/7/10, which the tooltip of 408438 reads as $m1; the
// effect's base points, 30, are what the talent's ranks replace.
func (shaman *Shaman) GetOverloadChance() float64 {
	return spellData.LightningOverload.FractionAt(shaman.Talents.LightningOverload)
}

const (
	SpellMaskNone             int64 = 0
	SpellMaskFlameShockDirect int64 = 1 << iota
	SpellMaskFlameShockDot
	SpellMaskLightningBolt
	SpellMaskLightningBoltOverload
	SpellMaskChainLightning
	SpellMaskChainLightningOverload
	SpellMaskEarthShock
	SpellMaskLightningShield
	SpellMaskMagmaTotem
	SpellMaskSearingTotem
	SpellMaskFireNova
	SpellMaskStormstrike
	SpellMaskEarthShield
	SpellMaskFrostShock
	SpellMaskFlametongueWeapon
	SpellMaskFrostbrandWeapon
	SpellMaskElementalMastery
	SpellMaskShamanisticRage
	SpellMaskBasicTotem

	SpellMaskFlameShock = SpellMaskFlameShockDirect | SpellMaskFlameShockDot
	SpellMaskFire       = SpellMaskFlameShock
	SpellMaskNature     = SpellMaskLightningBolt | SpellMaskLightningBoltOverload | SpellMaskChainLightning | SpellMaskChainLightningOverload | SpellMaskEarthShock
	SpellMaskFrost      = SpellMaskFrostShock
	SpellMaskOverload   = SpellMaskLightningBoltOverload | SpellMaskChainLightningOverload
	SpellMaskShock      = SpellMaskFlameShock | SpellMaskEarthShock | SpellMaskFrostShock
	SpellMaskFireTotem  = SpellMaskMagmaTotem | SpellMaskSearingTotem
	SpellMaskTotem      = SpellMaskFireTotem | SpellMaskBasicTotem
)
