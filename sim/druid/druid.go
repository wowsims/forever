package druid

import (
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

	WolfsheadEnergyBonus float64 // Wolfshead Helm (8345): +20 energy on shift into Cat Form
	WolfsheadRageBonus   float64 // Wolfshead Helm (8345): +5 rage on shift into Bear Form

	MHAutoSpell *core.Spell

	Barkskin             *DruidSpell
	Dash                 *DruidSpell
	DemoralizingRoar     *DruidSpell
	FaerieFire           *DruidSpell
	FaerieFireFeral      *DruidSpell
	FerociousBite        *DruidSpell
	Enrage               *DruidSpell
	EnrageAura           *core.Aura
	FrenziedRegeneration *DruidSpell
	Hurricane            *DruidSpell
	Innervate            *DruidSpell
	InsectSwarm          *DruidSpell
	Lacerate             *DruidSpell
	MangleBear           *DruidSpell
	Maul                 *DruidSpell
	Moonfire             *DruidSpell
	NaturesSwiftness     *DruidSpell
	Prowl                *DruidSpell
	Rake                 *DruidSpell
	Ravage               *DruidSpell
	Rejuvenation         *DruidSpell
	Rip                  *DruidSpell
	Shred                *DruidSpell
	Starfire             []*DruidSpell
	Swipe                *DruidSpell
	Wrath                *DruidSpell

	CatForm     *DruidSpell
	BearForm    *DruidSpell
	MoonkinForm *DruidSpell

	BearFormAura             *core.Aura
	CatFormAura              *core.Aura
	ClearcastingAura         *core.Aura
	DashAura                 *core.Aura
	FrenziedRegenerationAura *core.Aura
	DemoralizingRoarAuras    core.AuraArray
	FaerieFireAuras          core.AuraArray
	MangleAuras              core.AuraArray
	MoonkinFormAura          *core.Aura
	ProwlAura                *core.Aura

	form DruidForm

	IntensityEnrageRageBonus float64

	// Furor: chance to gain energy/rage when shifting into Cat/Bear Form.
	FurorProcChance float64

	// Maul queue (fires on next auto-attack swing, like warrior Heroic Strike)
	maulQueueAura  *core.Aura
	maulQueueSpell *core.Spell
	maulRealismICD *core.Cooldown
}

const (
	DruidSpellFlagNone        int64 = 0
	DruidSpellEntanglingRoots int64 = 1 << iota
	DruidSpellDemoralizingRoar
	DruidSpellFaerieFire
	DruidSpellFaerieFireFeral
	DruidSpellHurricane
	DruidSpellFerociousBite
	DruidSpellFrenziedRegeneration
	DruidSpellInnervate
	DruidSpellInsectSwarm
	DruidSpellLacerate
	DruidSpellMangleBear
	DruidSpellMaul
	DruidSpellMoonfireInitial
	DruidSpellMoonfireDoT
	DruidSpellRake
	DruidSpellRavage
	DruidSpellRip
	DruidSpellShred
	DruidSpellStarfire
	DruidSpellSwipe
	DruidSpellThorns
	DruidSpellWrath
	DruidSpellEnrage
	DruidSpellCatForm
	DruidSpellBearForm

	DruidSpellHealingTouch
	DruidSpellRegrowth
	DruidSpellLifebloom
	DruidSpellRejuvenation
	DruidSpellTranquility
	DruidSpellMarkOfTheWild
	DruidSpellSwiftmend
	DruidSpellCenarionWard

	// TODO: Forever abilities the sim does not model yet; see the stub file named for each.
	DruidSpellRevive

	DruidSpellLast
	DruidSpellsAll = DruidSpellLast<<1 - 1

	DruidSpellMoonfire           = DruidSpellMoonfireInitial | DruidSpellMoonfireDoT
	DruidSpellDoT                = DruidSpellMoonfireDoT | DruidSpellInsectSwarm
	DruidSpellHoT                = DruidSpellRejuvenation | DruidSpellLifebloom | DruidSpellRegrowth
	DruidSpellInstant            = DruidSpellMoonfire | DruidSpellFaerieFire
	DruidSpellMangle             = DruidSpellMangleBear
	DruidSpellBuilder            = DruidSpellMangle | DruidSpellShred | DruidSpellRake | DruidSpellRavage
	DruidSpellFinisher           = DruidSpellFerociousBite | DruidSpellRip
	DruidArcaneSpells            = DruidSpellMoonfire | DruidSpellMoonfireDoT | DruidSpellStarfire
	DruidNatureSpells            = DruidSpellWrath | DruidSpellHurricane | DruidSpellInsectSwarm
	DruidHealingNonInstantSpells = DruidSpellHealingTouch | DruidSpellRegrowth
	DruidHealingSpells           = DruidHealingNonInstantSpells | DruidSpellRejuvenation | DruidSpellLifebloom | DruidSpellSwiftmend
	DruidDamagingSpells          = DruidArcaneSpells | DruidNatureSpells
)

type SelfBuffs struct {
	InnervateTarget *proto.UnitReference
}

func (druid *Druid) GetCharacter() *core.Character {
	return &druid.Character
}

func (druid *Druid) AddPartyBuffs(partyBuffs *proto.PartyBuffs) {
	if druid.InForm(Cat|Bear) && druid.Talents.LeaderOfThePack {
		partyBuffs.LeaderOfThePack = true
	} else if druid.InForm(Moonkin) && druid.Talents.MoonkinForm {
		partyBuffs.MoonkinAura = true
	}
}

func (druid *Druid) RegisterSpell(config core.SpellConfig) *DruidSpell {
	return &DruidSpell{Spell: druid.Unit.RegisterSpell(config)}
}

func (druid *Druid) Initialize() {
	druid.setForm(druid.StartingForm)
	druid.AutoUnshift = druid.ClearForm

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
	druid.registerThornsSpell()
	druid.registerFormBreakingConsumes()
}

// registerFormBreakingConsumes patches ApplyEffects on potions, conjured items,
// and engineering explosives to drop Bear/Cat form when used. These spells all
// carry SpellFlagNoOnCastComplete, so OnCastComplete aura hooks never fire for
// them — we must wrap ApplyEffects directly instead.
func (druid *Druid) registerFormBreakingConsumes() {
	druid.Env.RegisterPostFinalizeEffect(func() {
		breakFlags := core.SpellFlagPotion | core.SpellFlagConjured | core.SpellFlagExplosive
		for _, spell := range druid.Spellbook {
			if !spell.Flags.Matches(breakFlags) {
				continue
			}
			prev := spell.ApplyEffects
			spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, sp *core.Spell) {
				prev(sim, target, sp)
				if druid.InForm(Bear) || druid.InForm(Cat) {
					druid.ClearForm(sim)
				}
			}
		}
	})
}

func (druid *Druid) RegisterBalanceSpells() {
	StarfireRankMap.Each(func(_ int32, r *spelldata.Spell) { druid.registerStarfireSpell(r) })
	druid.registerMoonfireSpell()
	druid.registerWrathSpell()
	druid.registerHurricaneSpell()
	druid.registerFaerieFireSpell()
}

func (druid *Druid) RegisterFeralCatSpells() {
	druid.registerCatFormSpell()

	// Forever has no Cat-form Mangle.
	druid.registerRakeSpell()
	druid.registerRipSpell()
	druid.registerFerociousBiteSpell()
	// TODO: Forever drops Faerie Fire (Feral); see registerFaerieFireFeralSpell.
	druid.registerShredSpell()
}

func (druid *Druid) RegisterFeralTankSpells() {
	druid.registerBearFormSpell()
	druid.registerBarkskin()
	druid.registerDemoralizingRoarSpell()
	// TODO: Forever drops Faerie Fire (Feral); see registerFaerieFireFeralSpell.
	druid.registerEnrageSpell()
	druid.registerFrenziedRegenerationSpell()
	druid.registerLacerateSpell()
	druid.registerPrimalBiteSpell()
	druid.registerMaulSpell()
	druid.registerSwipeBearSpell()
}

func (druid *Druid) Reset(_ *core.Simulation) {
	druid.setForm(druid.StartingForm)
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

	core.FillTalentsProto(druid.Talents.ProtoReflect(), talents, TalentTreeSizes)
	druid.EnableManaBar()

	druid.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	druid.AddStatDependency(stats.BonusArmor, stats.Armor, 1)
	druid.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[char.Class])

	return druid
}

type DruidSpell struct {
	*core.Spell

	// Optional fields used in snapshotting calculations
	CurrentSnapshotPower float64
	NewSnapshotPower     float64
	ShortName            string
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

func (druid *Druid) UpdateBleedPower(bleedSpell *DruidSpell, sim *core.Simulation, target *core.Unit, updateCurrent bool, updateNew bool) {
	snapshotPower := bleedSpell.ExpectedTickDamage(sim, target)

	if updateCurrent {
		bleedSpell.CurrentSnapshotPower = snapshotPower

		if sim.Log != nil {
			druid.Log(sim, "%s Snapshot Power: %.1f", bleedSpell.ShortName, snapshotPower)
		}
	}

	if updateNew {
		bleedSpell.NewSnapshotPower = snapshotPower

		if (sim.Log != nil) && !updateCurrent {
			druid.Log(sim, "%s Projected Power: %.1f", bleedSpell.ShortName, snapshotPower)
		}
	}
}

// Agent is a generic way to access underlying druid on any of the agents (for example balance druid.)
type DruidAgent interface {
	GetDruid() *Druid
}
