package hunter

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{16, 16, 18}

var autoShotRank = spellData.AutoShot.Highest()

type Hunter struct {
	core.Character

	Talents *proto.HunterTalents
	Options *proto.HunterOptions

	Pet      *HunterPet
	Hawks    []*HawkPet
	nextHawk int // The hawk the next Summon Hawk sends.

	RapidFire    *core.Spell
	RaptorStrike *core.Spell

	AspectOfTheHawkAura  *core.Aura
	AspectOfTheBeastAura *core.Aura
	MongooseBiteAura     *core.Aura

	rapidRecuperationAura  *core.Aura
	rapidRecuperationShare float64 // The share of regeneration while casting the aura grants, set by what granted it.
	rapidRecuperationSpend float64 // The share spending Rapid Killing grants.
}

func (hunter *Hunter) GetCharacter() *core.Character {
	return &hunter.Character
}

func (hunter *Hunter) GetHunter() *Hunter {
	return hunter
}

func RegisterHunter() {
	core.RegisterAgentFactory(
		proto.Player_Hunter{},
		proto.Spec_SpecHunter,
		func(character *core.Character, options *proto.Player, raid *proto.Raid) core.Agent {
			return NewHunter(character, options, options.GetHunter().Options.ClassOptions, raid)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Hunter)
			if !ok {
				panic("Invalid spec value for Hunter!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewHunter(character *core.Character, options *proto.Player, hunterOptions *proto.HunterOptions, raid *proto.Raid) *Hunter {
	hunter := &Hunter{
		Character: *character,
		Talents:   &proto.HunterTalents{},
		Options:   hunterOptions,
	}

	core.FillTalentsProto(hunter.Talents.ProtoReflect(), options.TalentsString, TalentTreeSizes)

	// The hunter casts its own mark, so a raid's permanent one would credit it before the cast.
	if raid.Debuffs != nil {
		raid.Debuffs.HuntersMark = false
	}

	hunter.PseudoStats.CanParry = true

	hunter.EnableManaBar()

	// Arrows add their damage per second to the ranged weapon, times its speed. A gun is taken to
	// fire them too, so no bullets are kept beside them.
	hunter.PseudoStats.BonusRangedDps += arrows[hunterOptions.Ammo.String()].DPS
	hunter.PseudoStats.RangedSpeedMultiplier *= core.SymmetricMultiplierFromPercent(quivers[hunterOptions.QuiverBonus.String()].Haste)

	hunter.EnableAutoAttacks(hunter, core.AutoAttackOptions{
		Ranged:          hunter.WeaponFromRanged(),
		MainHand:        hunter.WeaponFromMainHand(),
		OffHand:         hunter.WeaponFromOffHand(),
		ReplaceMHSwing:  hunter.TryRaptorStrike,
		AutoSwingRanged: true,
		AutoSwingMelee:  true,
	})

	// Auto Shot is row 75: its range, its missile speed, the bow, gun or crossbow it needs, and the
	// class flags that put the auto inside the talents that name it: Hawk Eye, Mortal Shots and
	// Deadly Aspects.
	rangedConfig := hunter.AutoAttacks.RangedConfig()
	rangedConfig.MinRange = core.MinRangedRange // Forever's, not the row's 8 yd
	rangedConfig.MaxRange = float64(autoShotRank.MaxRange)
	rangedConfig.MissileSpeed = float64(autoShotRank.Speed)
	rangedConfig.ClassFlags = autoShotRank.ClassFlags
	rangedConfig.CastRequirement = autoShotRank.CastRequirement()

	hunter.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	hunter.AddStatDependency(stats.Agility, stats.AttackPower, 1)
	hunter.AddStatDependency(stats.Agility, stats.RangedAttackPower, 2)
	hunter.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[hunter.Class])

	hunter.Pet = hunter.NewHunterPet()
	hunter.Hawks = hunter.newHawks()

	return hunter
}

func (hunter *Hunter) Initialize() {
	hunter.registerAimedShot()
	hunter.registerArcaneShot()
	hunter.registerMultiShot()
	hunter.registerSerpentSting()
	hunter.registerTraps()
	hunter.registerScorpidSting()
	hunter.registerRaptorStrike()
	hunter.registerMongooseBite()
	hunter.registerRapidFire()
	hunter.registerAspects()
	hunter.registerHuntersMark()
	hunter.addPvpGloves()
}

// A ranged cast shortens with ranged haste. Core's cast path applies spell haste, which the physical
// school's IgnoreHaste turns off, so the ranged multiplier is read here instead.
func (hunter *Hunter) hasteRangedCast(config *core.SpellConfig) {
	config.Cast.CastTime = func(spell *core.Spell) time.Duration {
		return hunter.ApplyRangedSpeed(spell.DefaultCast.CastTime)
	}
	config.Cast.ModifyCast = func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
		cast.CastTime = spell.CastTime()
	}
}

func (hunter *Hunter) stingTag() string {
	return "Sting-" + hunter.Label
}

// One sting per hunter on a target: applying either sting drops the hunter's other one.
func dropOtherSting(sim *core.Simulation, sting *core.Aura) {
	if active := sting.Unit.GetActiveAuraWithTag(sting.Tag); active != nil && active != sting {
		active.Deactivate(sim)
	}
}

func (hunter *Hunter) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	if hunter.Talents.TrueshotAura {
		raidBuffs.TrueshotAura = true
	}
}

func (hunter *Hunter) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (hunter *Hunter) Reset(_ *core.Simulation) {
	hunter.nextHawk = 0
}

func (hunter *Hunter) OnEncounterStart(sim *core.Simulation) {
}

// Agent is a generic way to access underlying hunter on any of the agents.
type HunterAgent interface {
	GetHunter() *Hunter
}
