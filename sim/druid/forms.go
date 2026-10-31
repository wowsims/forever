package druid

import (
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

type DruidForm uint8

const (
	Humanoid DruidForm = 1 << iota
	Bear
	Cat
	Moonkin
	Tree
	DireBear
)

func (form DruidForm) Matches(other DruidForm) bool {
	return (form & other) != 0
}

func (druid *Druid) setForm(form DruidForm) {
	switch {
	case form.Matches(DireBear):
		// Dire Bear keeps the bear bit so a bear-only cast and InForm(Bear) both match.
		druid.form = Bear | DireBear
		druid.ShapeshiftForm = shapeshiftOf(spellData.DireBearForm.Highest())
	case form.Matches(Bear):
		druid.form = Bear
		druid.ShapeshiftForm = shapeshiftOf(spellData.BearForm.Highest())
	case form.Matches(Cat):
		druid.form = Cat
		druid.ShapeshiftForm = shapeshiftOf(spellData.CatForm.Highest())
	case form.Matches(Moonkin):
		druid.form = Moonkin
		druid.ShapeshiftForm = shapeshiftOf(spellData.MoonkinForm.Highest())
	default:
		druid.form = Humanoid
		druid.ShapeshiftForm = 0
	}
}

// The shapeshift aura's misc value is the game files' form id.
func shapeshiftOf(rank *spelldata.Spell) dbcenums.ShapeshiftForm {
	return dbcenums.ShapeshiftForm(rank.EffectN(1).Misc)
}

func (druid *Druid) InForm(form DruidForm) bool {
	return druid.form.Matches(form)
}

func (druid *Druid) leaveForm(sim *core.Simulation) {
	switch {
	case druid.CatFormAura != nil && druid.CatFormAura.IsActive():
		druid.CatFormAura.Deactivate(sim)
	case druid.DireBearFormAura != nil && druid.DireBearFormAura.IsActive():
		druid.DireBearFormAura.Deactivate(sim)
	case druid.BearFormAura != nil && druid.BearFormAura.IsActive():
		druid.BearFormAura.Deactivate(sim)
	case druid.MoonkinFormAura != nil && druid.MoonkinFormAura.IsActive():
		druid.MoonkinFormAura.Deactivate(sim)
	}
}

func (druid *Druid) ClearForm(sim *core.Simulation) {
	druid.leaveForm(sim)
	druid.setForm(Humanoid)
	if druid.HasManaBar() {
		druid.SetCurrentPowerBar(core.ManaBar)
	}
}

func (druid *Druid) GetCatWeapon() core.Weapon {
	return druid.feralWeapon(1)
}

// SpellShapeshiftForm.DamageVariance is 0.4 for Cat, Bear, and Dire Bear.
const feralPawVariance = 0.4

// The paw uses the weapon's table DPS and form variance, with single-precision endpoints
// rounded before AP and flat enchant or stone damage.
func (druid *Druid) feralWeapon(swingSpeed float64) core.Weapon {
	paw := core.Weapon{
		BaseDamageMin:        1,
		BaseDamageMax:        1,
		SwingSpeed:           swingSpeed,
		NormalizedSwingSpeed: swingSpeed,
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		MaxRange:             core.MaxMeleeRange,
	}
	if item := druid.GetMHWeapon(); item != nil {
		mean := float32(item.UnroundedDPS * float32(swingSpeed*1000))
		mean = float32(mean * 0.001)
		half := float32(float32(feralPawVariance) / 2)
		flat := item.Enchant.WeaponDamage + item.TempEnchantWeaponDamage + druid.PseudoStats.BonusMHDps*swingSpeed
		paw.BaseDamageMin = max(1, math.Floor(float64(float32(mean*float32(1-half))))) + flat
		paw.BaseDamageMax = max(1, math.Round(float64(float32(mean*float32(1+half))))) + flat
	}
	return paw
}

// A bear's auto attack takes the two-hander rage factor: a combat log shows 11.25 rage a swing at the
// bear's 2.5 seconds, 4.5 rage per second of swing. No druid row states a rage multiplier of its own.
// TODO: Ingame test needed -- whether a bear holding a one-hander takes the one-hand factor instead, and
// whether a bear's crit doubles the whole swing as a two-hand warrior's does (measured on warriors only).
const BearRageIsTwoHanded = true

func (druid *Druid) GetBearWeapon() core.Weapon {
	return druid.feralWeapon(2.5)
}

func (druid *Druid) RegisterCatFormAura() {
	form := spellData.CatForm.Highest()
	druid.CatFormAura = druid.registerFormAura(form, druid.StartingForm.Matches(Cat))
	spelldata.ParseEffects(&druid.Character, druid.CatFormAura, spellData.CatFormTriggered.Highest())
	druid.CatFormAura.AttachStatDependency(druid.NewDynamicStatDependency(stats.FeralAttackPower, stats.AttackPower, 1))

	druid.CatFormAura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		druid.setForm(Cat)
		druid.SetFeralCombat(true)
		druid.SetCurrentPowerBar(core.EnergyBar)
		if !druid.Env.MeasuringStats {
			druid.AutoAttacks.SetMH(druid.GetCatWeapon())
		}
		druid.restoreCatEnergy(sim)
	})
	druid.CatFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.rememberCatEnergy(sim, druid.CurrentEnergy())
		druid.breakProwl(sim)
		druid.SetFeralCombat(false)
		druid.restoreHumanoidBody()
	})
}

// Bear Form (Passive2) 21178 is the bear forms' threat: A_MOD_THREAT on every school, 50 in the hotfixed
// game files. The spell store does not carry the row, as neither form's row names it.
// TODO: Read 21178 from the spell store once tools/database takes it for the druid.
const bearFormThreatMultiplier = 1.5

func (druid *Druid) RegisterBearFormAura() {
	druid.BearFormAura = druid.registerBearAura(spellData.BearForm.Highest(), spellData.BearFormTriggered.Highest(), Bear, druid.StartingForm == Bear)
	druid.DireBearFormAura = druid.registerBearAura(spellData.DireBearForm.Highest(), spellData.DireBearFormTriggered.Highest(), DireBear, druid.StartingForm.Matches(DireBear))
}

func (druid *Druid) registerBearAura(form *spelldata.Spell, passive *spelldata.Spell, kind DruidForm, starting bool) *core.Aura {
	aura := druid.registerFormAura(form, starting)
	aura.AttachStatDependency(druid.NewDynamicStatDependency(stats.FeralAttackPower, stats.AttackPower, 1))
	// Armor from items, base and bonus, health, attack power and the Faerie Fire spell mods are parsed.
	// TODO: Ingame research needed -- effect 4 of Bear Form and Dire Bear Form is aura 668, which core
	// does not name.
	spelldata.ParseEffects(&druid.Character, aura, passive)
	aura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		druid.pauseCatEnergy(sim)
		druid.setForm(kind)
		druid.SetFeralCombat(true)
		druid.SetCurrentPowerBar(core.RageBar)
		if !druid.Env.MeasuringStats {
			druid.AutoAttacks.SetMH(druid.GetBearWeapon())
			// A swing replacement also asks the rotation to queue an attack before the swing.
			// Only Bear forms can queue Maul.
			druid.AutoAttacks.SetReplaceMHSwing(druid.TryMaul)
		}
		druid.PseudoStats.ThreatMultiplier *= bearFormThreatMultiplier
	})
	aura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.PseudoStats.ThreatMultiplier /= bearFormThreatMultiplier
		druid.AutoAttacks.SetReplaceMHSwing(nil)
		druid.cancelMaulQueue(sim)
		if druid.HasEnergyBar() && !druid.Env.MeasuringStats {
			druid.EnergyPerTick = druid.furorEnergyPerTick
			druid.rememberCatEnergy(sim, druid.furorStoredEnergy)
		}
		druid.SetFeralCombat(false)
		druid.restoreHumanoidBody()
	})
	return aura
}

func (druid *Druid) registerFormAura(form *spelldata.Spell, starting bool) *core.Aura {
	return druid.RegisterAura(spelldata.AuraConfig(form, func(aura *core.Aura) {
		if starting {
			aura.BuildPhase = core.CharacterBuildPhaseBase
		}
	}))
}

// A change of form keeps the health percentage while the forms' health and Stamina move the maximum.
// It wraps the forms' handlers whole, so it runs in Initialize, after the talents have attached theirs.
func (druid *Druid) keepHealthPercentAcrossForms() {
	for _, aura := range []*core.Aura{druid.CatFormAura, druid.BearFormAura, druid.DireBearFormAura, druid.MoonkinFormAura} {
		if aura == nil {
			continue
		}
		onGain, onExpire := aura.OnGain, aura.OnExpire
		aura.OnGain = func(aura *core.Aura, sim *core.Simulation) {
			percent := druid.CurrentHealthPercent()
			onGain(aura, sim)
			druid.keepHealthPercent(percent)
		}
		aura.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
			percent := druid.CurrentHealthPercent()
			onExpire(aura, sim)
			druid.keepHealthPercent(percent)
		}
	}
}

func (druid *Druid) keepHealthPercent(percent float64) {
	if !druid.Env.MeasuringStats {
		druid.SetCurrentHealthPercent(percent)
	}
}

func (druid *Druid) restoreHumanoidBody() {
	if druid.Env != nil && !druid.Env.MeasuringStats && druid.AutoAttacks.MH() != nil {
		druid.AutoAttacks.SetMH(druid.WeaponFromMainHand())
	}
	if druid.HasManaBar() {
		druid.SetCurrentPowerBar(core.ManaBar)
	}
	druid.setForm(Humanoid)
}

func (druid *Druid) registerCatFormSpell() {
	druid.CatForm = druid.registerShapeshiftSpell(spellData.CatForm.Highest(), druid.CatFormAura)
}

func (druid *Druid) registerBearFormSpell() {
	druid.BearForm = druid.registerShapeshiftSpell(spellData.BearForm.Highest(), druid.BearFormAura)
	druid.DireBearForm = druid.registerShapeshiftSpell(spellData.DireBearForm.Highest(), druid.DireBearFormAura)
}

func (druid *Druid) registerShapeshiftSpell(rank *spelldata.Spell, aura *core.Aura) *DruidSpell {
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return aura != nil && (aura == druid.BearFormAura || aura == druid.DireBearFormAura || !aura.IsActive())
	}
	// A shift into Bear or Dire Bear Form starts the rage bar at 0, and Furor's rage comes on top. The
	// starting form is not cast, so the pull keeps the starting rage.
	var rageMetrics *core.ResourceMetrics
	if aura == druid.BearFormAura || aura == druid.DireBearFormAura {
		rageMetrics = druid.NewRageMetrics(core.ActionID{SpellID: rank.ID})
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		druid.leaveForm(sim)
		if rageMetrics != nil && druid.CurrentRage() > 0 {
			druid.SpendRage(sim, druid.CurrentRage(), rageMetrics)
		}
		aura.Activate(sim)
		if rageMetrics != nil {
			druid.furorRageOnEnterBear(sim)
		}
	}
	return druid.RegisterSpell(config)
}

// TODO: Verify a companion-triggered swing reset before adding a snake-weaving option.

// The rage chance, the tooltip's $m1: 0.2 at 1/5 through 1 at 5/5.
func (druid *Druid) furorFraction() float64 {
	return spellData.Furor.EffectAt(1).FractionAt(druid.Talents.Furor)
}

func (druid *Druid) rememberCatEnergy(sim *core.Simulation, energy float64) {
	if !druid.HasEnergyBar() {
		return
	}
	druid.furorStoredEnergy = energy
	druid.furorStoredValid = true
	druid.furorLeftCatAt = sim.CurrentTime
}

// Bear forms preserve the stored energy and pause regeneration until the form ends.
func (druid *Druid) pauseCatEnergy(sim *core.Simulation) {
	if !druid.HasEnergyBar() || druid.Env.MeasuringStats {
		return
	}
	energy := druid.CurrentEnergy()
	if druid.furorStoredValid {
		seconds := max(0, (sim.CurrentTime - druid.furorLeftCatAt).Seconds())
		energy = min(druid.furorStoredEnergy+10*seconds, 100)
	}
	druid.rememberCatEnergy(sim, energy)
	druid.furorEnergyPerTick = druid.EnergyPerTick
	druid.EnergyPerTick = 0
}

// The tooltip's rule, priced with $m2: the energy left in Cat Form plus ${10*($m2/100)} for each
// second spent outside Bear Form or Dire Bear Form, up to ${100*($m2/100)}. At 5/5 that is 10 a
// second, the Cat Form rate. With the talent untaken the cat comes back with none.
func (druid *Druid) restoreCatEnergy(sim *core.Simulation) {
	if !druid.HasEnergyBar() || druid.Env.MeasuringStats {
		return
	}
	// The energy bar ticks outside the bear forms, so a first entry with no bear form before it keeps
	// the pull's bar.
	// TODO: Ingame research needed -- what death or cancelling the form does to the stored energy, and
	// what a first entry later in the fight starts with.
	if !druid.furorStoredValid {
		return
	}
	seconds := max(0, (sim.CurrentTime - druid.furorLeftCatAt).Seconds())
	fraction := spellData.Furor.EffectAt(2).FractionAt(druid.Talents.Furor)
	gained := fraction * min(druid.furorStoredEnergy+10*seconds, 100)
	delta := gained - druid.CurrentEnergy()
	if delta > 0 {
		druid.AddEnergy(sim, delta, druid.furorEnergyMetrics())
	} else if delta < 0 {
		druid.SpendEnergy(sim, -delta, druid.furorEnergyMetrics())
	}
}

func (druid *Druid) furorRageOnEnterBear(sim *core.Simulation) {
	chance := druid.furorFraction()
	if chance == 0 || !druid.HasRageBar() || druid.Env.MeasuringStats {
		return
	}
	if !sim.Proc(chance, "Furor") {
		return
	}
	rage := spellData.FurorTriggered.Highest().EnergizeEffect().Tenths()
	druid.AddRage(sim, rage, druid.furorRageMetrics())
}

func (druid *Druid) furorEnergyMetrics() *core.ResourceMetrics {
	if druid.furorEnergyMetric == nil {
		druid.furorEnergyMetric = druid.NewEnergyMetrics(core.ActionID{SpellID: spellData.Furor.Highest().ID})
	}
	return druid.furorEnergyMetric
}

func (druid *Druid) furorRageMetrics() *core.ResourceMetrics {
	if druid.furorRageMetric == nil {
		druid.furorRageMetric = druid.NewRageMetrics(core.ActionID{SpellID: spellData.FurorTriggered.Highest().ID})
	}
	return druid.furorRageMetric
}

func (druid *Druid) RegisterMoonkinFormAura() {
	if !druid.Talents.MoonkinForm {
		return
	}
	druid.MoonkinFormAura = druid.registerFormAura(spellData.MoonkinForm.Highest(), druid.StartingForm.Matches(Moonkin))
	// Moonkin Form (Passive) 24905 states the form's armor from items, base and bonus. The crit of
	// Moonkin Aura 24907 is the raid buff AddRaidBuffs turns on.
	spelldata.ParseEffects(&druid.Character, druid.MoonkinFormAura, spellData.MoonkinFormTriggered.ByID(24905))
	// Of the form's own effects only Omen of Clarity's proc chance and proc cooldown have a sim kind.
	spelldata.ParseEffects(&druid.Character, druid.MoonkinFormAura, spellData.MoonkinForm.Highest(), spelldata.Effects(5, 6))
	druid.MoonkinFormAura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
		druid.setForm(Moonkin)
		druid.SetCurrentPowerBar(core.ManaBar)
	})
	druid.MoonkinFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.restoreHumanoidBody()
	})
}

func (druid *Druid) RegisterMoonkinFormSpell() {
	if druid.MoonkinFormAura == nil {
		return
	}
	druid.MoonkinForm = druid.registerShapeshiftSpell(spellData.MoonkinForm.Highest(), druid.MoonkinFormAura)
}
