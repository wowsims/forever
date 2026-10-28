package warlock

import (
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{17, 19, 16}

type Warlock struct {
	core.Character
	Talents *proto.WarlockTalents
	Options *proto.WarlockOptions

	// Base Spells
	Corruption  *core.Spell
	DrainLife   *core.Spell
	DrainSoul   *core.Spell
	Hellfire    *core.Spell
	Immolate    *core.Spell
	Incinerate  *core.Spell
	RainOfFire  *core.Spell
	SearingPain *core.Spell
	ShadowBolt  *core.Spell
	Soulfire    *core.Spell

	LifeTap *core.Spell

	// Curses
	CurseOfElements          *core.Spell
	CurseOfElementsAuras     core.AuraArray
	CurseOfRecklessness      *core.Spell
	CurseOfRecklessnessAuras core.AuraArray
	CurseOfWeakness          *core.Spell
	CurseOfWeaknessAuras     core.AuraArray
	CurseOfWeaknessAmpAuras  core.AuraArray
	Curses                   []*core.Spell

	// Banes
	BaneOfAgony *core.Spell
	BaneOfDoom  *core.Spell
	Banes       []*core.Spell

	BaneOfHavoc      *core.Spell
	BaneOfHavocAuras core.AuraArray

	// Summons
	SummonImp        *core.Spell
	SummonVoidwalker *core.Spell
	SummonSuccubus   *core.Spell
	SummonFelhunter  *core.Spell

	// Talent Tree Spells
	AmplifyCurse     *core.Spell
	DemonicSacrifice *core.Spell
	Conflagrate      *core.Spell
	Shadowburn       *core.Spell
	SiphonLife       *core.Spell
	Wrack            *core.Spell

	// Auras
	DemonicSacrificeAuras  map[proto.WarlockOptions_Summon]*core.Aura // Demonic Sacrifice's effect, by the demon sacrificed
	FelDominationAura      *core.Aura
	AmplifyCurseAura       *core.Aura
	AmplifyCurseMultiplier float64

	ShadowVulnerabilityAuras core.AuraArray
	DemonicBrandAuras        core.AuraArray
	conflagrateKeepsImmolate float64 // Shadow and Flame's chance that Conflagrate leaves Immolate burning
	NightfallProcAura        *core.Aura
	ShadowEmbraceAura        *core.Aura
	DemonicKnowledgeAura     *core.Aura

	// Pets
	ActivePet  *WarlockPet
	Felhunter  *WarlockPet
	Imp        *WarlockPet
	Succubus   *WarlockPet
	Voidwalker *WarlockPet

	// Armors
	DemonArmor *core.Aura

	SoulShardsAura *core.Aura      // the Soul Shards left this iteration, as its stacks
	reagentWaivers []reagentWaiver // the effects that lift the reagents of some spells while up

	petActiveAuras  []*core.Aura                // the warlock's effects that last while a demon is out
	sacrificedDemon proto.WarlockOptions_Summon // the demon whose Demonic Sacrifice effect is up, NoSummon for none
	lifeTapPetShare float64                     // Demonic Energies: the share of Life Tap's mana the demon gains
}

func (warlock *Warlock) GetCharacter() *core.Character {
	return &warlock.Character
}

func (warlock *Warlock) GetWarlock() *Warlock {
	return warlock
}

func RegisterWarlock() {
	core.RegisterAgentFactory(
		proto.Player_Warlock{},
		proto.Spec_SpecWarlock,
		func(character *core.Character, options *proto.Player, raid *proto.Raid) core.Agent {
			return NewWarlock(character, options, options.GetWarlock().Options.ClassOptions, raid)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Warlock)
			if !ok {
				panic("Invalid spec value for Warlock!")
			}
			player.Spec = playerSpec
		},
	)
}

func (warlock *Warlock) Initialize() {

	// Curses
	warlock.registerCurseOfElements()
	warlock.registerBaneOfDoom()
	warlock.registerBaneOfAgony()
	warlock.registerCurseOfRecklessness()
	warlock.registerCurseOfWeakness()

	warlock.registerCorruption()
	warlock.registerDeathCoil()
	warlock.registerDrainLife()
	warlock.registerDrainSoul()
	warlock.registerHellfire()
	warlock.registerImmolate()
	warlock.registerLifeTap()
	warlock.registerRainOfFire()
	warlock.registerShadowBolt()
	warlock.registerSearingPain()

	warlock.registerSoulfire()
	warlock.registerSoulShards()
	shared.RegisterShoot(&warlock.Character)

	warlock.registerSummons()
	warlock.registerBloodPact()

	warlock.registerArmors()

	warlock.PseudoStats.SelfHealingMultiplier = 1.0
}

func (warlock *Warlock) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {

}

func (warlock *Warlock) Reset(sim *core.Simulation) {
}

func (warlock *Warlock) OnEncounterStart(sim *core.Simulation) {}

func NewWarlock(character *core.Character, options *proto.Player, warlockOptions *proto.WarlockOptions, raid *proto.Raid) *Warlock {
	warlock := &Warlock{
		Character: *character,
		Talents:   &proto.WarlockTalents{},
		Options:   warlockOptions,
	}

	warlock.FillTalentsProto(warlock.Talents.ProtoReflect(), options.TalentsString, TalentTreeSizes)

	warlock.resolveSacrifice()

	if raid.Debuffs != nil {
		// The assigned curse replaces the raid's copy of it, which is turned off. Matches GetAssignedCurse:
		// NoCurse turns off none, and any value it does not name is Elements.
		switch warlock.Options.CurseOptions {
		case proto.WarlockOptions_NoCurse:
		case proto.WarlockOptions_Recklessness:
			raid.Debuffs.CurseOfRecklessness = false
		case proto.WarlockOptions_Weakness:
			raid.Debuffs.CurseOfWeakness = false
		default:
			raid.Debuffs.CurseOfElements = false
		}
	}

	warlock.EnableManaBar()
	warlock.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	warlock.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])

	warlock.registerPets()

	return warlock
}

// Settles which demon is sacrificed and which is out, once, so everything after reads them as settled:
//   - Without Demonic Sacrifice, nothing is sacrificed.
//   - Without Demonic Pact, summoning any demon cancels the effect, so no demon is out.
//   - With it, the demon out may be another; resummoning the sacrificed one still cancels the effect.
func (warlock *Warlock) resolveSacrifice() {
	options := warlock.Options
	switch {
	case !warlock.Talents.DemonicSacrifice:
		options.SacrificedDemon = proto.WarlockOptions_NoSummon
	case options.SacrificedDemon == proto.WarlockOptions_NoSummon:
	case !warlock.Talents.DemonicPact:
		options.Summon = proto.WarlockOptions_NoSummon
	case options.Summon == options.SacrificedDemon:
		options.SacrificedDemon = proto.WarlockOptions_NoSummon
	}
}

// Returns the number of this warlock's affliction effects active on the target, for Soul Siphon.
// Dots are tagged "Affliction"; the tag is shared by every caster, but core labels each caster's dot aura
// "<label>-<caster unit index>", so the suffix picks out this warlock's own.
// Curses are the exception: their debuff's Tag is its exclusive category, so they cannot carry the tag,
// and are read through the Curses list instead.
// TODO: Confirm in-game that curses count toward Soul Siphon.
func (warlock *Warlock) AfflictionCount(target *core.Unit) float64 {
	own := "-" + strconv.Itoa(int(warlock.UnitIndex))
	count := 0.0
	for _, aura := range target.GetAurasWithTag("Affliction") {
		if aura.IsActive() && strings.HasSuffix(aura.Label, own) {
			count++
		}
	}
	for _, curse := range warlock.Curses {
		for _, auras := range curse.RelatedAuraArrays {
			if aura := auras.Get(target); aura != nil && aura.IsActive() {
				count++
			}
		}
	}
	return count
}

// Only one Bane per Warlock can be active on any one target, so a Bane landing takes the others off it.
func (warlock *Warlock) deactivateOtherBanes(sim *core.Simulation, bane *core.Spell, target *core.Unit) {
	deactivateOthersOnTarget(sim, warlock.Banes, bane, target)
}

// Only one Curse per Warlock can be active on any one target, so a Curse landing takes the others off it.
func (warlock *Warlock) deactivateOtherCurses(sim *core.Simulation, curse *core.Spell, target *core.Unit) {
	deactivateOthersOnTarget(sim, warlock.Curses, curse, target)
}

// Takes every spell in the provided group off of the target, except the landing spell.
func deactivateOthersOnTarget(sim *core.Simulation, group []*core.Spell, landing *core.Spell, target *core.Unit) {
	for _, other := range group {
		if other == landing {
			continue
		}
		if dot := other.Dot(target); dot != nil {
			dot.Deactivate(sim)
		}
		for _, auras := range other.RelatedAuraArrays {
			if aura := auras.Get(target); aura != nil {
				aura.Deactivate(sim)
			}
		}
	}
}

// Slots in a target's per-caster damage multipliers (core.EnableDamageDoneByCaster / AttachDDBC), one per
// warlock effect that raises only this warlock's damage against the target: a debuff, or Decimation's bonus
// against a target below 35%. Each needs its own slot so one debuff's expiry cannot clear another's, and every
// caller passes DDBC_Total so the slice is sized alike.
const (
	DDBC_ShadowVulnerability int = iota
	DDBC_Wrack
	DDBC_Decimation
	DDBC_Total
)

// Scales a result already calculated - a ramp, a per-target bonus - and keeps its threat in step with
// the damage, which the calculation set from the damage before the scaling.
//
// Landed warlock dots tick through spelldata's TickOutcome: they tested hit once, when applied,
// so a tick rolls only the crit the row's Periodic Can Crit allows (the Forever beta log has no
// SPELL_PERIODIC_MISSED). Area effects that tick, Hellfire and Rain of Fire, roll hit on every tick.
func scaleResult(sim *core.Simulation, spell *core.Spell, result *core.SpellResult, multiplier float64) {
	result.Damage *= multiplier
	result.Threat = spell.ThreatFromDamage(sim, result.Outcome, result.Damage, spell.Unit.AttackTables[result.Target.UnitIndex])
}

// Combines the class flags of several spells into one set, for filters the data doesn't name the spells for.
// All of them must share a family, as a warlock's spells do.
func combineClassFlags(flags ...core.ClassFlags) core.ClassFlags {
	var combined core.ClassFlags
	for _, f := range flags {
		if combined.Family != 0 && f.Family != combined.Family {
			panic("combineClassFlags: spells from different families")
		}
		combined.Family = f.Family
		for i := range combined.Mask {
			combined.Mask[i] |= f.Mask[i]
		}
	}
	return combined
}

// Agent is a generic way to access underlying warlock on any of the agents.
type WarlockAgent interface {
	GetWarlock() *Warlock
}

const (
	WarlockSpellFlagNone    int64 = 0
	WarlockSpellConflagrate int64 = 1 << iota
	WarlockSpellShadowBolt
	WarlockSpellImmolate
	WarlockSpellIncinerate
	WarlockSpellSoulFire
	WarlockSpellShadowBurn
	WarlockSpellLifeTap
	WarlockSpellCorruption
	WarlockSpellCurseOfElements
	WarlockSpellBaneOfAgony
	WarlockSpellBaneOfDoom
	WarlockSpellDrainLife
	WarlockSpellHellfire
	WarlockSpellSearingPain
	WarlockSpellImpFireBolt
	WarlockSpellSuccubusLashOfPain
	WarlockSpellRainOfFire
	WarlockSpellCurseOfRecklessness
	WarlockSpellCurseOfWeakness
	WarlockSpellSiphonLife
	WarlockSpellWrack
	WarlockSpellDrainSoul
	WarlockSpellDeathCoil
	WarlockSpellSummonDemon
	WarlockSpellAll int64 = 1<<iota - 1
)
