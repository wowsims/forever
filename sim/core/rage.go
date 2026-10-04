package core

import (
	"fmt"

	"github.com/wowsims/forever/sim/core/proto"
)

const ThreatPerRageGained = 5

// Rage from damage taken counts a hit before any of the defender's mitigation, with a physical hit counted
// as if armor reduced it by half, whatever the defender's armor.
const DamageTakenRageArmorMultiplier = 0.5

// Damage taken grants DamageTakenRageFactor * damage / c rage, where c is the rage conversion value at the defender's level (230.6 at 60).
// TODO: Ingame test needed -- a level 14 bear with 1032 health gained a third of this; 18 * damage / maximum health
// fits that bear and a level 14 warrior with 359 health alike.
const DamageTakenRageFactor = 2.5

func rageConversion(level int32) float64 {
	l := float64(level)
	return 0.0091107836*l*l + 3.225598133*l + 4.2652911
}

const (
	BaseRageHitFactor     = 3.46
	TwoHandRageHitFactor  = 4.5
	TwoHandRageMultiplier = TwoHandRageHitFactor / BaseRageHitFactor
	// A critical auto attack adds this share of its rage before the two-hand multiplier, for every rage
	// user (Rule of Rage (DND) states 100).
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

	RageRefundMetrics     *ResourceMetrics
	EncounterStartMetrics *ResourceMetrics
}

type RageBarOptions struct {
	MaxRage            float64
	StartingRage       float64
	BaseRageMultiplier float64
}

func (unit *Unit) EnableRageBar(options RageBarOptions) {
	rageFromDamageTakenMetrics := unit.NewRageMetrics(ActionID{OtherID: proto.OtherAction_OtherActionDamageTaken})
	damageTakenRageConversion := rageConversion(unit.Level)

	// Direct hits and periodic ticks from a unit grant rage alike.
	// TODO: Ingame test needed -- physical periodic damage (bleeds) counts at half like a physical hit; no
	// combat log covers a rage user taking a bleed.
	rageFromDamageTaken := func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
		if unit.GetCurrentPowerBar() != RageBar {
			return
		}
		if !result.Landed() {
			return
		}

		damage := result.PreMitigationDamage
		if spell.SpellSchool.Matches(SpellSchoolPhysical) {
			damage *= DamageTakenRageArmorMultiplier
		}
		generatedRage := DamageTakenRageFactor * damage / damageTakenRageConversion * unit.rageBar.damageTakenRageMultiplier

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
			if weapon.NormalizedSwingSpeed == TwoHandNormalizedSwingSpeed {
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
				}
				metrics = spell.ResourceMetrics
			}
			unit.AddRage(sim, generatedRage, metrics)
		},
		OnSpellHitTaken:       rageFromDamageTaken,
		OnPeriodicDamageTaken: rageFromDamageTaken,
	})

	// Not a real spell, just holds metrics from rage gain threat.
	unit.RegisterSpell(SpellConfig{
		ActionID: ActionID{OtherID: proto.OtherAction_OtherActionRageGain},
	})

	maxRage := max(100.0, options.MaxRage)

	unit.rageBar = rageBar{
		unit:                      unit,
		maxRage:                   maxRage,
		startingRage:              max(0, options.StartingRage),
		offHandRageMultiplier:     1,
		damageTakenRageMultiplier: 1,
		RageRefundMetrics:         unit.NewRageMetrics(ActionID{OtherID: proto.OtherAction_OtherActionRefund}),
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
	metrics.AddEvent(amount, newRage-rb.currentRage)

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

func (rb *rageBar) doneIteration() {
	if rb.unit == nil {
		return
	}

	rageGainSpell := rb.unit.GetSpell(ActionID{OtherID: proto.OtherAction_OtherActionRageGain})

	for _, resourceMetrics := range rb.unit.Metrics.resources {
		if resourceMetrics.Type != proto.ResourceType_ResourceTypeRage {
			continue
		}
		if resourceMetrics.ActionID.SameActionIgnoreTag(ActionID{OtherID: proto.OtherAction_OtherActionDamageTaken}) {
			continue
		}
		if resourceMetrics.ActionID.SameActionIgnoreTag(ActionID{OtherID: proto.OtherAction_OtherActionRefund}) {
			continue
		}
		if resourceMetrics.ActualGainForCurrentIteration() <= 0 {
			continue
		}

		// Need to exclude rage gained from white hits. Rather than have a manual list of all IDs that would
		// apply here (autos, WF attack, sword spec procs, etc), just check if the effect caused any damage.
		sourceSpell := rb.unit.GetSpell(resourceMetrics.ActionID)
		if sourceSpell != nil && sourceSpell.SpellMetrics[0].TotalDamage > 0 {
			continue
		}

		rageGainSpell.SpellMetrics[0].Casts += resourceMetrics.EventsForCurrentIteration()
		rageGainSpell.ApplyAOEThreatIgnoreMultipliers(resourceMetrics.ActualGainForCurrentIteration() * ThreatPerRageGained)
	}
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
