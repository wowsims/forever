package core

import (
	"fmt"

	"github.com/wowsims/forever/sim/core/proto"
)

// Booked in AddRage at gain time. Rage from white hits, hits taken and refunds carries
// ResourceMetrics.NoThreat and makes none; a new rage source that comes from damage sets NoThreat too.
const ThreatPerRageGained = 5

// Damage taken grants DamageTakenRageFactor * (1 - A) * damage / c rage, where damage is the hit before any of the
// defender's mitigation (armor, block and absorbs included), c is the rage conversion value at the attacker's level
// (230.6 at 60) and A is the armor reduction the game assumes in place of the defender's own. The defender's level,
// health and armor play no part.
const DamageTakenRageFactor = 2.5

// A by the attacker's level: 20% up to level 42, 30% at 44, 36% at 46 and 38% at 50, measured on combat logs, and
// 40% at 63, the top of the 20% to 40% range the game build's notes give. Levels between two rows interpolate; no
// log covers 43, 45, 47-49 or 61-63.
var damageTakenRageAssumedArmor = []struct {
	level     int32
	reduction float64
}{
	{level: 42, reduction: 0.20},
	{level: 44, reduction: 0.30},
	{level: 46, reduction: 0.36},
	{level: 50, reduction: 0.38},
	{level: 63, reduction: 0.40},
}

func damageTakenRageArmorReduction(attackerLevel int32) float64 {
	rows := damageTakenRageAssumedArmor
	if attackerLevel <= rows[0].level {
		return rows[0].reduction
	}
	for i := 1; i < len(rows); i++ {
		if attackerLevel <= rows[i].level {
			lo, hi := rows[i-1], rows[i]
			return lo.reduction + (hi.reduction-lo.reduction)*float64(attackerLevel-lo.level)/float64(hi.level-lo.level)
		}
	}
	return rows[len(rows)-1].reduction
}

func rageConversion(level int32) float64 {
	l := float64(level)
	return 0.0091107836*l*l + 3.225598133*l + 4.2652911
}

const (
	BaseRageHitFactor     = 3.46
	TwoHandRageHitFactor  = 4.5
	TwoHandRageMultiplier = TwoHandRageHitFactor / BaseRageHitFactor
	// A critical auto attack adds this share of its rage before the two-hand multiplier, for every rage
	// user (the game files' Rule of Rage (DND) 1322574 states 100; the spell store does not carry it).
	// TODO: Ingame test needed -- confirm with a dual-wield combat log that an off-hand crit's bonus is
	// the off-hand swing's own rage, the off-hand halving included.
	CritRageBonus = 1.0
)

type rageBar struct {
	unit *Unit

	maxRage      float64
	startingRage float64
	currentRage  float64

	offHandRageMultiplier     float64
	damageTakenRageMultiplier float64

	rageGainSpell *Spell

	RageRefundMetrics     *ResourceMetrics
	EncounterStartMetrics *ResourceMetrics
}

type RageBarOptions struct {
	MaxRage            float64
	StartingRage       float64
	BaseRageMultiplier float64
	TwoHandRage        bool // Auto attacks take the two-hand factor whatever the weapon's normalized speed.
}

func (unit *Unit) EnableRageBar(options RageBarOptions) {
	rageFromDamageTakenMetrics := unit.NewRageMetrics(ActionID{OtherID: proto.OtherAction_OtherActionDamageTaken})
	rageFromDamageTakenMetrics.NoThreat = true
	// Direct hits and periodic ticks from a unit grant rage alike.
	// TODO: Ingame test needed -- the rule is measured on melee hits. Spells land at a multiple of it the logs do
	// not explain yet: Immolate ticks, Shadow Bolt and Arcane Missiles about 1 times, Fireball and Mortal Strike
	// about 1.2 times, Poison ticks, Venom Sting and Lizard Bolt about 2 times.
	rageFromDamageTaken := func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
		if unit.GetCurrentPowerBar() != RageBar {
			return
		}
		if !result.Landed() {
			return
		}

		attackerLevel := spell.Unit.Level
		generatedRage := DamageTakenRageFactor * (1 - damageTakenRageArmorReduction(attackerLevel)) * result.PreMitigationDamage /
			rageConversion(attackerLevel) * unit.rageBar.damageTakenRageMultiplier

		unit.AddRage(sim, generatedRage, rageFromDamageTakenMetrics)
	}

	unit.SetCurrentPowerBar(RageBar)
	unit.RegisterAura(Aura{
		Label:    "RageBar",
		Duration: NeverExpires,
		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if unit.GetCurrentPowerBar() != RageBar {
				return
			}
			if !result.Landed() {
				return
			}

			hitFactor := BaseRageHitFactor
			handMultiplier := 1.0
			var weapon *Weapon

			switch spell.ProcMask {
			case ProcMaskMeleeMHAuto:
				weapon = unit.AutoAttacks.MH()
			case ProcMaskMeleeOHAuto:
				// OH hits generate 50% of the rage they would if they were MH hits
				hitFactor /= 2
				handMultiplier = unit.rageBar.offHandRageMultiplier
				weapon = unit.AutoAttacks.OH()
			default:
				return
			}

			// rage is normalized so it only depends on weapon swing speed and some multipliers
			swingRage := hitFactor * weapon.SwingSpeed * options.BaseRageMultiplier * handMultiplier
			generatedRage := swingRage
			if options.TwoHandRage || weapon.NormalizedSwingSpeed == TwoHandNormalizedSwingSpeed {
				generatedRage *= TwoHandRageMultiplier
			}
			// The crit bonus scales the swing's rage without the two-hand multiplier; the off-hand factors stay in it.
			if result.DidCrit() {
				generatedRage += CritRageBonus * swingRage
			}

			var metrics *ResourceMetrics
			if spell.Cost != nil {
				metrics = spell.Cost.ResourceCostImpl.(*RageCost).ResourceMetrics
			} else {
				if spell.ResourceMetrics == nil {
					spell.ResourceMetrics = spell.Unit.NewRageMetrics(spell.ActionID)
					spell.ResourceMetrics.NoThreat = true
				}
				metrics = spell.ResourceMetrics
			}
			unit.AddRage(sim, generatedRage, metrics)
		},
		OnSpellHitTaken:       rageFromDamageTaken,
		OnPeriodicDamageTaken: rageFromDamageTaken,
	})

	// Not a real spell, just holds metrics from rage gain threat.
	rageGainSpell := unit.RegisterSpell(SpellConfig{
		ActionID: ActionID{OtherID: proto.OtherAction_OtherActionRageGain},
	})

	maxRage := max(100.0, options.MaxRage)
	rageRefundMetrics := unit.NewRageMetrics(ActionID{OtherID: proto.OtherAction_OtherActionRefund})
	rageRefundMetrics.NoThreat = true

	unit.rageBar = rageBar{
		unit:                      unit,
		maxRage:                   maxRage,
		startingRage:              max(0, options.StartingRage),
		offHandRageMultiplier:     1,
		damageTakenRageMultiplier: 1,
		rageGainSpell:             rageGainSpell,
		RageRefundMetrics:         rageRefundMetrics,
		EncounterStartMetrics:     unit.NewRageMetrics(ActionID{OtherID: proto.OtherAction_OtherActionEncounterStart}),
	}
}

func (unit *Unit) HasRageBar() bool {
	return unit.rageBar.unit != nil
}

func (rb *rageBar) CurrentRage() float64 {
	return rb.currentRage
}

func (rb *rageBar) MaximumRage() float64 {
	return rb.maxRage
}

// Sets the most rage the bar holds. A raised maximum grants no rage; rage above a lowered one is lost
// and counted against the metrics. Without a Simulation there is no rage yet to lose.
func (rb *rageBar) SetMaximumRage(sim *Simulation, maxRage float64, metrics *ResourceMetrics) {
	rb.maxRage = maxRage

	if sim != nil && rb.currentRage > maxRage {
		rb.SpendRage(sim, rb.currentRage-maxRage, metrics)
	}
}

func (rb *rageBar) SetOffHandRageMultiplier(multiplier float64) {
	rb.offHandRageMultiplier = multiplier
}

func (rb *rageBar) MultiplyDamageTakenRageGen(multiplier float64) {
	rb.damageTakenRageMultiplier *= multiplier
}

func (rb *rageBar) AddRage(sim *Simulation, amount float64, metrics *ResourceMetrics) {
	if amount < 0 {
		panic("Trying to add negative rage!")
	}

	newRage := min(rb.currentRage+amount, rb.maxRage)
	gain := newRage - rb.currentRage
	metrics.AddEvent(amount, gain)

	if gain > 0 && !metrics.NoThreat {
		rb.addRageGainThreat(sim, gain)
	}

	if sim.Log != nil {
		rb.unit.Log(sim, "Gained %0.3f rage from %s (%0.3f --> %0.3f) of %0.0f total.", amount, metrics.ActionID, rb.currentRage, newRage, rb.maxRage)
	}

	rb.currentRage = newRage
	if !sim.Options.Interactive {
		rb.unit.ReactToEvent(sim, false, true)
	}
}

func (rb *rageBar) SpendRage(sim *Simulation, amount float64, metrics *ResourceMetrics) {
	if amount < 0 {
		panic("Trying to spend negative rage!")
	}

	newRage := rb.currentRage - amount
	metrics.AddEvent(-amount, -amount)

	if sim.Log != nil {
		rb.unit.Log(sim, "Spent %0.3f rage from %s (%0.3f --> %0.3f) of %0.0f total.", amount, metrics.ActionID, rb.currentRage, newRage, rb.maxRage)
	}

	rb.currentRage = newRage
}

func (rb *rageBar) ResetRageBar(sim *Simulation, rageToKeep float64) {
	if rb.currentRage > rageToKeep {
		rb.SpendRage(sim, rb.currentRage-rageToKeep, rb.EncounterStartMetrics)
	} else if rageToKeep > rb.currentRage {
		rb.AddRage(sim, rageToKeep-rb.currentRage, rb.EncounterStartMetrics)
	}
}

func (rb *rageBar) reset(_ *Simulation) {
	if rb.unit == nil {
		return
	}

	rb.currentRage = min(rb.startingRage, rb.maxRage)
	rb.damageTakenRageMultiplier = 1
}

// Each rage gained is ThreatPerRageGained threat times the unit's threat multiplier at the time, split
// between the targets in combat at the time.
func (rb *rageBar) addRageGainThreat(sim *Simulation, gain float64) {
	rb.rageGainSpell.applySplitThreat(sim, gain*ThreatPerRageGained*rb.rageGainSpell.unitThreatMultiplier())
	rb.rageGainSpell.SpellMetrics[0].Casts++
}

type RageCostOptions struct {
	Cost int32

	Refund        float64
	RefundMetrics *ResourceMetrics // Optional, will default to unit.RageRefundMetrics if not supplied.
}
type RageCost struct {
	Refund          float64
	RefundMetrics   *ResourceMetrics
	ResourceMetrics *ResourceMetrics
}

func newRageCost(spell *Spell, options RageCostOptions) *SpellCost {
	if options.Refund > 0 && options.RefundMetrics == nil {
		options.RefundMetrics = spell.Unit.RageRefundMetrics
	}

	return &SpellCost{
		spell:                   spell,
		BaseCost:                options.Cost,
		PercentModifier:         1,
		AdditivePercentModifier: 1,
		ResourceCostImpl: &RageCost{
			Refund:          options.Refund,
			RefundMetrics:   options.RefundMetrics,
			ResourceMetrics: spell.Unit.NewRageMetrics(spell.ActionID),
		},
	}
}

func (rc *RageCost) MeetsRequirement(_ *Simulation, spell *Spell) bool {
	spell.CurCast.Cost = spell.Cost.GetCurrentCost()
	return spell.Unit.CurrentRage() >= spell.CurCast.Cost
}
func (rc *RageCost) CostFailureReason(sim *Simulation, spell *Spell) string {
	return fmt.Sprintf("not enough rage (Current Rage = %0.03f, Rage Cost = %0.03f)", spell.Unit.CurrentRage(), spell.CurCast.Cost)
}
func (rc *RageCost) SpendCost(sim *Simulation, spell *Spell) {
	if spell.CurCast.Cost > 0 {
		spell.Unit.SpendRage(sim, spell.CurCast.Cost, rc.ResourceMetrics)
	}
}
func (rc *RageCost) IssueRefund(sim *Simulation, spell *Spell) {
	if rc.Refund > 0 && spell.CurCast.Cost > 0 {
		spell.Unit.AddRage(sim, rc.Refund*spell.CurCast.Cost, rc.RefundMetrics)
	}
}

func (spell *Spell) RageMetrics() *ResourceMetrics {
	return spell.Cost.ResourceCostImpl.(*RageCost).ResourceMetrics
}
