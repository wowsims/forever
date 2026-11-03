package warrior

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{17, 17, 18}

type WarriorInputs struct {
	UseBattleShout bool
	DefaultStance  proto.WarriorStance

	StartingRage   float64
	QueueDelay     int32
	StanceSnapshot bool
	HasBsT2        bool

	HasBattleShoutRank7 bool
}

// The game files' SpellClassOptions for the Sweeping Strikes copy, a sim-only sub-spell that takes its
// parent's flags, the game files having no row for it.
var SpellFlagsSweepingStrikes = spellData.SweepingStrikes.Highest().ClassFlags

type Warrior struct {
	core.Character

	ClassSpellScaling float64

	Talents *proto.WarriorTalents

	WarriorInputs

	// Current state
	thunderClapEffectBonus float64

	DemoralizingShout *core.Spell
	BattleStance      *core.Spell
	DefensiveStance   *core.Spell
	BerserkerStance   *core.Spell

	Rend                            *core.Spell
	DeepWounds                      *core.Spell
	MortalStrike                    *core.Spell
	SweepingStrikesNormalizedAttack *core.Spell

	Cleave            *core.Spell
	MockingBlow       *core.Spell
	ChallengingShout  *core.Spell
	IntimidatingShout *core.Spell
	Disarm            *core.Spell
	Taunt             *core.Spell
	VictoryRush       *core.Spell

	curQueueAura       *core.Aura
	curQueuedAutoSpell *core.Spell
	queuedRealismICD   *core.Cooldown
	pendingQueue       *core.PendingAction

	EnrageAura      *core.Aura
	GoreDrinkerAura *core.Aura

	SweepingStrikesAura *core.Aura
	OverpowerAura       *core.Aura

	DemoralizingShoutAuras core.AuraArray
	SunderArmorAuras       core.AuraArray
}

func (warrior *Warrior) GetCharacter() *core.Character {
	return &warrior.Character
}

func (warrior *Warrior) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
}

func (warrior *Warrior) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (warrior *Warrior) Initialize() {
	warrior.registerRecklessness()
	warrior.registerShieldWall()
	warrior.registerRetaliation()

	warrior.registerBerserkerRage()
	warrior.registerBloodrage()
	warrior.registerCharge()
	warrior.registerIntercept()
	warrior.registerPummel()
	warrior.registerHamstring()
	warrior.registerDisarm()
	warrior.registerTaunt()

	warrior.registerRend()
	warrior.registerSunderArmor()
	warrior.registerHeroicStrike()
	warrior.registerCleave()
	warrior.registerOverpower()
	warrior.registerSlam()
	warrior.registerWhirlwind()
	warrior.registerExecute()
	warrior.registerThunderClap()
	warrior.registerRevenge()
	warrior.registerShieldBlock()
	warrior.registerShieldBash()
	warrior.registerMockingBlow()
	warrior.registerVictoryRush()

	warrior.registerStances()
	warrior.registerBattleShout()
	warrior.registerDemoralizingShout()
	warrior.registerChallengingShout()
	warrior.registerIntimidatingShout()
}

func (warrior *Warrior) Reset(_ *core.Simulation) {
	warrior.curQueueAura = nil
	warrior.curQueuedAutoSpell = nil
	warrior.pendingQueue = nil

	switch warrior.DefaultStance {
	case proto.WarriorStance_WarriorStanceBattle:
		warrior.ShapeshiftForm = battleStanceRank.ShapeshiftForm()
	case proto.WarriorStance_WarriorStanceDefensive:
		warrior.ShapeshiftForm = defensiveStanceRank.ShapeshiftForm()
	case proto.WarriorStance_WarriorStanceBerserker:
		warrior.ShapeshiftForm = berserkerStanceRank.ShapeshiftForm()
	}
}

func (warrior *Warrior) OnEncounterStart(sim *core.Simulation) {}

func (warrior *Warrior) GetMainHandType() proto.HandType {
	mh := warrior.GetMHWeapon()

	if mh != nil && (mh.HandType == proto.HandType_HandTypeTwoHand) {
		return proto.HandType_HandTypeTwoHand
	}

	return proto.HandType_HandTypeOneHand
}

func NewWarrior(character *core.Character, options *proto.WarriorOptions, talents string, inputs WarriorInputs) *Warrior {
	warrior := &Warrior{
		Character:     *character,
		Talents:       &proto.WarriorTalents{},
		WarriorInputs: inputs,
	}
	warrior.FillTalentsProto(warrior.Talents.ProtoReflect(), talents, TalentTreeSizes)

	warrior.EnableRageBar(core.RageBarOptions{
		MaxRage:            100,
		BaseRageMultiplier: 1,
		StartingRage:       inputs.StartingRage,
	})

	warrior.EnableAutoAttacks(warrior, core.AutoAttackOptions{
		MainHand:       warrior.WeaponFromMainHand(),
		OffHand:        warrior.WeaponFromOffHand(),
		AutoSwingMelee: true,
		ReplaceMHSwing: warrior.TryHSOrCleave,
	})
	warrior.AutoAttacks.SetQueuedSwingCanceler(warrior.cancelQueuedAttack)

	warrior.PseudoStats.CanParry = true

	warrior.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	warrior.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])
	warrior.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	// Without a delay the sim re-queues Heroic Strike faster than a player can, which lands
	// back-to-back hits around extra attacks.
	warrior.queuedRealismICD = &core.Cooldown{
		Timer:    warrior.NewTimer(),
		Duration: time.Millisecond * time.Duration(inputs.QueueDelay),
	}

	return warrior
}

// Drops the queued Heroic Strike or Cleave, and a press still waiting out the Queue Delay.
func (warrior *Warrior) cancelQueuedAttack(sim *core.Simulation) {
	if !warrior.pendingQueue.IsConsumed() {
		warrior.pendingQueue.Cancel(sim)
		warrior.pendingQueue = nil
	}
	if warrior.curQueueAura != nil {
		warrior.curQueueAura.Deactivate(sim)
	}
}

func (warrior *Warrior) CastNormalizedSweepingStrikesAttack(results core.SpellResultSlice, sim *core.Simulation) {
	if warrior.SweepingStrikesAura != nil && warrior.SweepingStrikesAura.IsActive() {
		for _, result := range results {
			if result.Landed() {
				warrior.SweepingStrikesNormalizedAttack.Cast(sim, warrior.Env.NextActiveTargetUnit(result.Target))
				warrior.SweepingStrikesAura.RemoveStack(sim)
				break
			}
		}
	}
}

// Agent is a generic way to access underlying warrior on any of the agents.
type WarriorAgent interface {
	GetWarrior() *Warrior
}
