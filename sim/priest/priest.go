package priest

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{18, 17, 18}

type Priest struct {
	core.Character
	SelfBuffs
	Talents              *proto.PriestTalents
	PrayerOfHealingMode  proto.PrayerOfHealingMode
	PrayerOfMendingJumps int32
	PrayerOfMendingDelay float64
	HolyNovaHealTargets  int32

	Latency float64

	InnerFocusAura *core.Aura

	VampiricEmbrace *core.Spell

	powerInLightAuras core.AuraArray
	weakenedSoulAuras core.AuraArray
}

type SelfBuffs struct {
	PreShadowform bool
	Armor         proto.PriestOptions_Armor
}

func (priest *Priest) GetCharacter() *core.Character {
	return &priest.Character
}
func (priest *Priest) GetPriest() *Priest {
	return priest
}

func (priest *Priest) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (priest *Priest) Initialize() {
	spellData.HolyFire.Each(priest.registerHolyFireSpell)
	MindBlastRankMap.Each(priest.registerMindBlastSpell)
	ShadowWordPainRankMap.Each(priest.registerShadowWordPainSpell)
	ShadowWordDeathRankMap.Each(priest.registerShadowWordDeathSpell)
	SmiteRankMap.Each(priest.registerSmiteSpell)
	DevouringPlagueRankMap.Each(priest.registerDevouringPlagueSpell)
	spellData.InnerFire.Each(priest.registerInnerFire)
	spellData.Fade.Each(priest.registerFade)
	shared.RegisterShoot(&priest.Character)

	switch priest.Race {
	case proto.Race_RaceNightElf:
		StarshardsRankMap.Each(priest.registerStarshardsSpell)
		spellData.ElunesGrace.Each(priest.registerElunesGrace)
	case proto.Race_RaceTroll:
		spellData.HexOfWeakness.Each(priest.registerHexOfWeakness)
		spellData.Shadowguard.Each(priest.registerShadowguard)
	case proto.Race_RaceUndead:
		spellData.TouchOfWeakness.Each(priest.registerTouchOfWeakness)
		spellData.DarkSacrifice.Each(priest.registerDarkSacrifice)
	case proto.Race_RaceDwarf:
		spellData.Chastise.Each(priest.registerChastise)
		spellData.DesperatePrayer.Each(priest.registerDesperatePrayer)
	}

	priest.registerPrayerOfFortitude()
	priest.registerPrayerOfSpirit()
	priest.registerPrayerOfShadowProtection()
}

func (priest *Priest) InitializeHealingSpells() {
	priest.registerPowerWordShield()
	spellData.FlashHeal.Each(priest.registerFlashHeal)
	spellData.Heal.Each(priest.registerHeal)
	spellData.LesserHeal.Each(priest.registerLesserHeal)
	spellData.GreaterHeal.Each(priest.registerGreaterHeal)
	spellData.PrayerOfHealing.Each(priest.registerPrayerOfHealing)
	spellData.Renew.Each(priest.registerRenew)
}

func (priest *Priest) registerSelfBuffSpell(rank *spelldata.Spell, aura *core.Aura) *core.Spell {
	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
		if aura.MaxStacks > 0 {
			aura.SetStacks(sim, aura.MaxStacks)
		}
	}
	config.RelatedSelfBuff = aura

	return priest.RegisterSpell(config)
}

// A prayer's own copy on every player in the raid, up once the priest casts it. It bids in the same
// category as the raid buff of that name, so the two count once. The cast waits until nothing as
// strong holds the category, which the raid buff, permanent from the start, always does.
func (priest *Priest) registerPrayer(rank *spelldata.Spell, newAura func(*core.Unit, bool, int32) *core.Aura) *core.Spell {
	auras := priest.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
		return newAura(unit, true, 0)
	})
	ownAura := auras.Get(&priest.Unit)

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ExtraCastCondition = func(sim *core.Simulation, _ *core.Unit) bool {
		return ownAura.ShouldRefreshExclusiveEffects(sim, 0)
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		auras.ActivateAllPlayers(sim)
	}
	config.RelatedAuraArrays = auras.ToMap()

	return priest.RegisterSpell(config)
}

func (priest *Priest) periodicDamageSpellConfig(rank *spelldata.Spell, classSpellMask int64) core.SpellConfig {
	tick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = classSpellMask

	config.Dot = spelldata.DotConfig(rank, tick, spelldata.RankLabel(rank, &priest.Unit))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		if rank.PeriodicCanCrit() {
			return spell.CalcPeriodicDamage(sim, target, tick.Average(), spell.OutcomeExpectedMagicCrit)
		}
		return spell.CalcPeriodicDamage(sim, target, tick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	return config
}

func (priest *Priest) directDamageSpellConfig(rank *spelldata.Spell, classSpellMask int64) core.SpellConfig {
	baseDamage := rank.DamageEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = classSpellMask

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, baseDamage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, baseDamage.Average(), spell.OutcomeExpectedMagicCrit)
	}

	return config
}

func (priest *Priest) directHealSpellConfig(rank *spelldata.Spell, classSpellMask int64) core.SpellConfig {
	heal := rank.HealEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = classSpellMask
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		target = priest.healTarget(target)
		priest.withRenewedHope(sim, target, spell, func() {
			spell.CalcAndDealHealing(sim, target, heal.Average(), spell.OutcomeHealingCrit)
		})
	}

	return config
}

func (priest *Priest) healTarget(target *core.Unit) *core.Unit {
	if target == nil || target.Type == core.EnemyUnit {
		return &priest.Unit
	}
	return target
}

func (priest *Priest) Reset(_ *core.Simulation) {

}

func (priest *Priest) OnEncounterStart(sim *core.Simulation) {
}

func New(char *core.Character, selfBuffs SelfBuffs, talents string) *Priest {
	priest := &Priest{
		Character: *char,
		SelfBuffs: selfBuffs,
		Talents:   &proto.PriestTalents{},
	}

	priest.FillTalentsProto(priest.Talents.ProtoReflect(), talents, TalentTreeSizes)
	priest.EnableManaBar()
	priest.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	priest.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[char.Class])

	return priest
}

// Agent is a generic way to access underlying priest on any of the agents.
type PriestAgent interface {
	GetPriest() *Priest
}

func NewPriest(character *core.Character, options *proto.Player) *Priest {
	classOptions := options.GetDpsPriest().GetOptions().GetClassOptions()
	selfBuffs := SelfBuffs{
		PreShadowform: classOptions.GetPreShadowform(),
		Armor:         classOptions.GetArmor(),
	}

	basePriest := New(character, selfBuffs, options.TalentsString)
	basePriest.Latency = float64(basePriest.ChannelClipDelay.Milliseconds())

	return basePriest
}

func RegisterPriest() {
	core.RegisterAgentFactory(
		proto.Player_DpsPriest{},
		proto.Spec_SpecDpsPriest,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewPriest(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_DpsPriest)
			if !ok {
				panic("Invalid spec value for Priest!")
			}
			player.Spec = playerSpec
		},
	)
}

const (
	PriestSpellDevouringPlague int64 = 1 << iota
	PriestSpellHolyFire
	PriestSpellMindBlast
	PriestSpellMindFlay
	PriestSpellPenance
	PriestSpellShadowWordDeath
	PriestSpellShadowWordPain
	PriestSpellSmite

	PriestSpellFlashHeal
	PriestSpellLesserHeal
	PriestSpellHeal
	PriestSpellGreaterHeal
	PriestSpellBindingHeal
	PriestSpellRenew
	PriestSpellPowerWordShield
	PriestSpellHolyNova
	PriestSpellLightwell
	PriestSpellPrayerOfHealing
	PriestSpellPrayerOfMending

	PriestSpellCureDisease
	PriestSpellAbolishDisease

	PriestShadowSpells = PriestSpellDevouringPlague |
		PriestSpellShadowWordDeath |
		PriestSpellShadowWordPain |
		PriestSpellMindFlay |
		PriestSpellMindBlast
)
