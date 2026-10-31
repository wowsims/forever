package druid

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{16, 20, 16}

type Druid struct {
	core.Character
	SelfBuffs

	Talents *proto.DruidTalents

	StartingForm DruidForm

	CannotShredTarget bool
	// Another raid member keeps a bleed on the target, so Rend and Tear always applies.
	AssumeBleedActive bool

	// Wolfshead Helm (8345): energy added to Shifting Power, and the share added to Enrage's instant rage.
	WolfsheadEnergyBonus float64
	WolfsheadEnrageBonus float64

	MHAutoSpell *core.Spell

	Barkskin             *DruidSpell
	Bash                 *DruidSpell
	ChallengingRoar      *DruidSpell
	Claw                 *DruidSpell
	Cower                *DruidSpell
	Dash                 *DruidSpell
	DemoralizingRoar     *DruidSpell
	FaerieFire           *DruidSpell
	FeralCharge          *DruidSpell
	FerociousBite        *DruidSpell
	Enrage               *DruidSpell
	EnrageAura           *core.Aura
	FrenziedRegeneration *DruidSpell
	Growl                *DruidSpell
	Hurricane            *DruidSpell
	Innervate            *DruidSpell
	InsectSwarm          *DruidSpell
	Lacerate             *DruidSpell
	Maul                 *DruidSpell
	Moonfire             *DruidSpell
	NaturesSwiftness     *DruidSpell
	Pounce               *DruidSpell
	PounceBleed          *DruidSpell
	PrimalBite           *DruidSpell
	Prowl                *DruidSpell
	Shadowmeld           *DruidSpell
	Rake                 *DruidSpell
	Ravage               *DruidSpell
	Rejuvenation         *DruidSpell
	Rip                  *DruidSpell
	ShiftingPower        *DruidSpell
	Shred                *DruidSpell
	Starfire             []*DruidSpell
	Thorns               *DruidSpell
	Swipe                *DruidSpell
	Wrath                *DruidSpell

	CatForm      *DruidSpell
	BearForm     *DruidSpell
	DireBearForm *DruidSpell
	MoonkinForm  *DruidSpell

	BearFormAura             *core.Aura
	CatFormAura              *core.Aura
	ClearcastingAura         *core.Aura
	OmenOfClarityAura        *core.Aura // The proc trigger, whose ICD says when Omen of Clarity can proc again.
	DireBearFormAura         *core.Aura
	DashAura                 *core.Aura
	FeralChargeMovementAura  *core.Aura
	FrenziedRegenerationAura *core.Aura
	DemoralizingRoarAuras    core.AuraArray
	FaerieFireAuras          core.AuraArray
	MoonkinFormAura          *core.Aura
	ProwlAura                *core.Aura
	ShadowmeldAura           *core.Aura

	form DruidForm

	// Stored Cat energy and the start of its regeneration interval outside Bear forms.
	furorEnergyPerTick float64
	furorStoredEnergy  float64
	furorStoredValid   bool
	furorLeftCatAt     time.Duration
	furorEnergyMetric  *core.ResourceMetrics
	furorRageMetric    *core.ResourceMetrics
	enrageMetric       *core.ResourceMetrics

	// Maul queue (fires on next auto-attack swing, like warrior Heroic Strike)
	maulQueueAura  *core.Aura
	maulQueueSpell *core.Spell
	maulRealismICD *core.Cooldown
}

type SelfBuffs struct {
	InnervateTarget *proto.UnitReference
}

func (druid *Druid) GetCharacter() *core.Character {
	return &druid.Character
}

func (druid *Druid) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	if druid.InForm(Cat|Bear) && druid.Talents.LeaderOfThePack {
		raidBuffs.LeaderOfThePack = true
	} else if druid.InForm(Moonkin) && druid.Talents.MoonkinForm {
		raidBuffs.MoonkinAura = true
	}
}

func (druid *Druid) RegisterSpell(config core.SpellConfig) *DruidSpell {
	return &DruidSpell{Spell: druid.Unit.RegisterSpell(config)}
}

func (druid *Druid) Initialize() {
	druid.setForm(druid.StartingForm)
	druid.keepHealthPercentAcrossForms()
	druid.AutoUnshift = druid.ClearForm
	druid.registerProwlBreakingCasts()

	druid.Env.RegisterPostFinalizeEffect(func() {
		druid.MHAutoSpell = druid.AutoAttacks.MHAuto()
	})

	druid.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, slot proto.ItemSlot) {
		switch {
		case druid.InForm(Cat):
			druid.AutoAttacks.SetMH(druid.GetCatWeapon())
		case druid.InForm(Bear):
			druid.AutoAttacks.SetMH(druid.GetBearWeapon())
		}
	})

	druid.RegisterBaselineSpells()
}

func (druid *Druid) RegisterBaselineSpells() {
	druid.registerInnervateCD()
	druid.registerFaerieFireSpell()
	druid.registerOmenOfClarity()
}

func (druid *Druid) RegisterBalanceSpells() {
	StarfireRankMap.Each(func(_ int32, r *spelldata.Spell) { druid.registerStarfireSpell(r) })
	druid.registerMoonfireSpell()
	druid.registerWrathSpell()
	druid.registerHurricaneSpell()
	druid.registerBarkskin()
	druid.RegisterCasterSpells()
}

// The feral specs take Thorns from the buff settings, so only the caster specs register it.
func (druid *Druid) RegisterCasterSpells() {
	druid.registerThornsSpell()
}

func (druid *Druid) RegisterFeralSpells() {
	druid.registerBarkskin()
	druid.registerCatFormSpell()
	druid.registerBearFormSpell()
	druid.registerClawSpell()
	druid.registerShredSpell()
	druid.registerRavageSpell()
	druid.registerRakeSpell()
	druid.registerProwlSpell()
	druid.registerShadowmeld()
	druid.registerPounceSpell()
	druid.registerRipSpell()
	druid.registerFerociousBiteSpell()
	druid.registerCowerSpell()
	druid.registerDashSpell()
	druid.registerMaulSpell()
	druid.registerSwipeBearSpell()
	druid.registerLacerateSpell()
	druid.registerPrimalBiteSpell()
	druid.registerShiftingPowerSpell()
	druid.registerEnrageSpell()
	druid.registerFrenziedRegenerationSpell()
	druid.registerDemoralizingRoarSpell()
	druid.registerBashSpell()
	druid.registerGrowlSpell()
	druid.registerChallengingRoarSpell()
	druid.registerFeralChargeSpell()
	druid.registerFeralMenuGroups()
}

func (druid *Druid) Reset(_ *core.Simulation) {
	druid.furorStoredValid = false
	druid.furorStoredEnergy = 0
	druid.furorLeftCatAt = 0
	druid.setForm(Humanoid)
}

func (druid *Druid) OnEncounterStart(sim *core.Simulation) {
}

func New(char *core.Character, form DruidForm, selfBuffs SelfBuffs, talents string) *Druid {
	druid := &Druid{
		Character:    *char,
		SelfBuffs:    selfBuffs,
		Talents:      &proto.DruidTalents{},
		StartingForm: form,
	}
	druid.setForm(form)

	druid.FillTalentsProto(druid.Talents.ProtoReflect(), talents, TalentTreeSizes)
	druid.EnableManaBar()

	druid.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	druid.AddStatDependency(stats.BonusArmor, stats.Armor, 1)
	druid.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[char.Class])

	return druid
}

type DruidSpell struct {
	*core.Spell
}

func (ds *DruidSpell) IsReady(sim *core.Simulation) bool {
	if ds == nil {
		return false
	}
	return ds.Spell.IsReady(sim)
}

func (ds *DruidSpell) CanCast(sim *core.Simulation, target *core.Unit) bool {
	if ds == nil {
		return false
	}
	return ds.Spell.CanCast(sim, target)
}

func (ds *DruidSpell) IsEqual(s *core.Spell) bool {
	if ds == nil || s == nil {
		return false
	}
	return ds.Spell == s
}

// Agent is a generic way to access underlying druid on any of the agents (for example balance druid.)
type DruidAgent interface {
	GetDruid() *Druid
}
