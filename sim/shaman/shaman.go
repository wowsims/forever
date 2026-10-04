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

	shaman.WindfuryAPBonus = 475.0 //Base Windfury Bonus

	return shaman
}

func (shaman *Shaman) GetImbueProcMask(imbue proto.ShamanImbue) core.ProcMask {
	var mask core.ProcMask
	if shaman.SelfBuffs.ImbueMH == imbue || shaman.SelfBuffs.ImbueMHSwap == imbue {
		mask |= core.ProcMaskMeleeMH
	}
	if shaman.SelfBuffs.ImbueOH == imbue {
		mask |= core.ProcMaskMeleeOH
	}
	return mask
}

// Which buffs this shaman is using.
type SelfBuffs struct {
	ShieldProcrate float64
	ImbueMH        proto.ShamanImbue
	ImbueOH        proto.ShamanImbue
	ImbueMHSwap    proto.ShamanImbue
	ImbueOHSwap    proto.ShamanImbue
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

	WindfuryAPBonus float64

	Talents   *proto.ShamanTalents
	SelfBuffs SelfBuffs

	Totems *proto.ShamanTotems

	// The expiration time of each totem (earth, air, fire, water).
	TotemExpirations [4]time.Duration

	LavaBurst *core.Spell

	Stormstrike           *core.Spell
	StormstrikeCastResult *core.SpellResult

	LightningShieldAura *core.Aura
	WaterShieldAura     *core.Aura
	ShieldSelfProcSpell *core.Spell

	EarthShock *core.Spell
	FlameShock *core.Spell
	FrostShock *core.Spell

	StormStrikeDebuffAuras core.AuraArray

	MagmaTotem   *core.Spell
	SearingTotem *core.Spell
	FireNova     *core.Spell
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

func (shaman *Shaman) Initialize() {
	shaman.registerChainLightningSpell()
	shaman.registerLightningBoltSpell()
	shaman.registerShieldsSpells()
	shaman.registerMagmaTotemSpell()
	shaman.registerSearingTotemSpell()
	shaman.registerFireNovaSpell()
	shaman.registerWindfuryTotemSpell()
	shaman.registerStrengthOfEarthTotemSpell()
	shaman.registerGraceOfAirTotemSpell()
	shaman.registerManaSpringTotemSpell()
	shaman.registerShocks()
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
	SpellMaskFlametongueTotem
	SpellMaskStormstrikeCast
	SpellMaskStormstrikeDamage
	SpellMaskEarthShield
	SpellMaskFrostShock
	SpellMaskFlametongueWeapon
	SpellMaskWindfuryWeapon
	SpellMaskFrostbrandWeapon
	SpellMaskRockbiterWeapon
	SpellMaskElementalMastery
	SpellMaskShamanisticRage
	SpellMaskBasicTotem
	SpellMaskShieldSelfProc

	SpellMaskStormstrike = SpellMaskStormstrikeCast | SpellMaskStormstrikeDamage
	SpellMaskFlameShock  = SpellMaskFlameShockDirect | SpellMaskFlameShockDot
	SpellMaskFire        = SpellMaskFlameShock
	SpellMaskNature      = SpellMaskLightningBolt | SpellMaskLightningBoltOverload | SpellMaskChainLightning | SpellMaskChainLightningOverload | SpellMaskEarthShock
	SpellMaskFrost       = SpellMaskFrostShock
	SpellMaskOverload    = SpellMaskLightningBoltOverload | SpellMaskChainLightningOverload
	SpellMaskShock       = SpellMaskFlameShock | SpellMaskEarthShock | SpellMaskFrostShock
	SpellMaskFireTotem   = SpellMaskMagmaTotem | SpellMaskSearingTotem
	SpellMaskTotem       = SpellMaskFireTotem | SpellMaskBasicTotem
	SpellMaskImbue       = SpellMaskFrostbrandWeapon | SpellMaskWindfuryWeapon | SpellMaskFlametongueWeapon | SpellMaskRockbiterWeapon
)
