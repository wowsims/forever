package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

const ThreatPerManaGained = 0.5

// Mana regenerates continuously in game; the bar credits it every 100 ms, and at once whenever the
// rate it regenerates at changes.
const manaTickDuration = time.Second / 10

type manaBar struct {
	unit     *Unit
	BaseMana float64

	currentMana float64

	manaRegenMultiplier float64

	spiritRegenBase      float64
	spiritRegenPerSpirit float64

	manaCastingMetrics    *ResourceMetrics
	manaNotCastingMetrics *ResourceMetrics
	JowManaMetrics        *ResourceMetrics
	VtManaMetrics         *ResourceMetrics
	JowiseManaMetrics     *ResourceMetrics
	manaGainSpell         *Spell // Holds the threat mana gains make.

	ReplenishmentAura *Aura

	agent        Agent         // Told of every mana tick; nil while the bar does not tick.
	lastManaTick time.Duration // The regen is credited up to here.
	nextManaTick time.Duration
	paidBy       *Spell // The spell that last paid mana, until a channel it starts claims the payment.

	// OOM while the rotation skips a spell it wants for its mana cost, whatever it casts instead. A pass
	// the unit starts free to act updates it.
	oom          bool
	oomStartedAt time.Duration
	costBlocked  bool // The current rotation pass skipped a spell for its mana cost.
}

// EnableManaBar will setup caster stat dependencies (int->mana and int->spellcrit)
// as well as enable the mana gain action to regenerate mana.
// It will then enable mana gain metrics for reporting.
func (character *Character) EnableManaBar() {
	character.EnableManaBarWithModifier()
	character.Unit.SetCurrentPowerBar(ManaBar)
}

func (character *Character) EnableManaBarWithModifier() {

	if character.Unit.Type == PlayerUnit {
		// Pets might have different scaling so let them handle their scaling
		character.AddStatDependency(stats.Intellect, stats.SpellCritPercent, CritPerIntMaxLevel[character.Class])

		// Assumes all units have >= 20 intellect.
		// See https://wowwiki-archive.fandom.com/wiki/Base_mana.
		// Subtract out the non-linear part of the formula separately, so that weird
		// mana values are not included when using the stat dependency manager.
		character.AddStat(stats.Mana, 20-15*20)
		character.AddStatDependency(stats.Intellect, stats.Mana, 15)
	}

	character.manaGainSpell = character.RegisterSpell(SpellConfig{
		ActionID: ActionID{OtherID: proto.OtherAction_OtherActionManaGain},
	})

	// Spirit restores 6.25 + Spirit/8 mana per second
	character.spiritRegenBase, character.spiritRegenPerSpirit = 6.25, 1.0/8

	character.BaseMana = character.GetBaseStats()[stats.Mana]
	character.Unit.manaBar.unit = &character.Unit
	character.manaCastingMetrics = character.NewManaMetrics(ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 1})
	character.manaNotCastingMetrics = character.NewManaMetrics(ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 2})
}

// SetSpiritManaRegen sets the spirit regen to base + Spirit*perSpirit mana per second.
// It replaces the class default, so call it after EnableManaBar.
func (character *Character) SetSpiritManaRegen(base float64, perSpirit float64) {
	character.spiritRegenBase, character.spiritRegenPerSpirit = base, perSpirit
}

func (unit *Unit) HasManaBar() bool {
	return unit.manaBar.unit != nil
}

// Empty handler so Agents don't have to provide one if they have no logic to add.
func (unit *Unit) OnManaTick(sim *Simulation) {}

// Gets the Maxiumum mana including bonus and temporary affects that would increase your mana pool.
func (unit *Unit) MaxMana() float64 {
	return unit.stats[stats.Mana]
}
func (unit *Unit) CurrentMana() float64 {
	return unit.currentMana
}
func (unit *Unit) CurrentManaPercent() float64 {
	return unit.CurrentMana() / unit.MaxMana()
}

func (unit *Unit) AddMana(sim *Simulation, amount float64, metrics *ResourceMetrics) {
	if amount < 0 {
		panic("Trying to add negative mana!")
	}

	oldMana := unit.CurrentMana()
	newMana := min(oldMana+amount, unit.MaxMana())

	metrics.AddEvent(amount, newMana-oldMana)

	if sim.Log != nil {
		unit.Log(sim, "Gained %0.3f mana from %s (%0.3f --> %0.3f) of %0.0f total.", amount, metrics.ActionID, oldMana, newMana, unit.MaxMana())
	}

	unit.currentMana = newMana
	unit.Metrics.ManaGained += newMana - oldMana
	unit.addManaGainThreat(sim, newMana-oldMana, metrics)
}

// Each mana gained is ThreatPerManaGained threat, split between the targets in combat at the time.
// Regen makes none, and neither does a source marked NoThreat. Vampiric Touch's mana is the priest's
// threat, which the priest code adds.
func (mb *manaBar) addManaGainThreat(sim *Simulation, gain float64, metrics *ResourceMetrics) {
	if gain <= 0 || mb.manaGainSpell == nil || metrics.NoThreat ||
		metrics.ActionID.SameActionIgnoreTag(ActionID{OtherID: proto.OtherAction_OtherActionManaRegen}) ||
		metrics.ActionID.SameActionIgnoreTag(ActionID{SpellID: 34917}) {
		return
	}
	mb.manaGainSpell.SpellMetrics[0].Casts++
	mb.manaGainSpell.applySplitThreat(sim, gain*ThreatPerManaGained)
}

func (unit *Unit) SpendMana(sim *Simulation, amount float64, metrics *ResourceMetrics) {
	if amount < 0 {
		panic("Trying to spend negative mana!")
	}

	newMana := unit.CurrentMana() - amount
	metrics.AddEvent(-amount, -amount)

	if sim.Log != nil {
		unit.Log(sim, "Spent %0.3f mana from %s (%0.3f --> %0.3f) of %0.0f total.", amount, metrics.ActionID, unit.CurrentMana(), newMana, unit.MaxMana())
	}

	unit.currentMana = newMana
	unit.Metrics.ManaSpent += amount
}

func (mb *manaBar) doneIteration(sim *Simulation) {
	if mb.unit == nil {
		return
	}

	mb.EndOOMEvent(sim)
}

// Returns the rate of mana regen per second from mp5.
func (unit *Unit) MP5ManaRegenPerSecond() float64 {
	return unit.stats[stats.MP5] / 5.0
}

// Returns the rate of mana regen per second from spirit.
func (unit *Unit) SpiritManaRegenPerSecond() float64 {
	return unit.spiritRegenBase + unit.stats[stats.Spirit]*unit.spiritRegenPerSpirit
}

// Returns the rate of mana regen per second, assuming this unit is
// considered to be casting.
func (unit *Unit) ManaRegenPerSecondWhileCasting() float64 {
	regenRate := unit.MP5ManaRegenPerSecond()

	// The shares of spirit regen that continue while casting add up to at most all of it.
	regenRate += unit.SpiritManaRegenPerSecond() * unit.PseudoStats.SpiritRegenMultiplier * min(unit.PseudoStats.SpiritRegenRateCasting, 1)

	regenRate *= unit.manaRegenMultiplier

	return regenRate
}

// Returns the rate of mana regen per second, assuming this unit is
// considered to be not casting.
func (unit *Unit) ManaRegenPerSecondWhileNotCasting() float64 {
	regenRate := unit.MP5ManaRegenPerSecond()

	regenRate += unit.SpiritManaRegenPerSecond() * unit.PseudoStats.SpiritRegenMultiplier

	regenRate *= unit.manaRegenMultiplier

	return regenRate
}

// Recomputes the regen rates the mana bar ticks at, after crediting the regen so far at the old ones.
// Without a Simulation there is no regen to credit yet.
func (unit *Unit) UpdateManaRegenRates(sim *Simulation) {
	if sim != nil {
		unit.creditManaRegen(sim)
	}
	unit.manaRegenWhileCasting = unit.ManaRegenPerSecondWhileCasting()
	unit.manaRegenWhileNotCasting = unit.ManaRegenPerSecondWhileNotCasting()
}

func (unit *Unit) MultiplyManaRegenSpeed(sim *Simulation, multiplier float64) {
	unit.manaRegenMultiplier *= multiplier
	unit.UpdateManaRegenRates(sim)
}

// Credits the regen since the last tick. Up to the end of the five-second rule window, or all of it
// while a paid channel runs, it comes at the casting rate, and after that at the full rate.
func (mb *manaBar) creditManaRegen(sim *Simulation) {
	if mb.agent == nil || sim.CurrentTime <= mb.lastManaTick {
		return
	}
	unit := mb.unit
	from := mb.lastManaTick
	mb.lastManaTick = sim.CurrentTime
	if !unit.IsEnabled() {
		return
	}

	castingUntil := sim.CurrentTime
	if !unit.channelPaidMana() {
		castingUntil = max(from, min(unit.PseudoStats.FiveSecondRuleRefreshTime, sim.CurrentTime))
	}
	if amount := unit.manaRegenWhileCasting * (castingUntil - from).Seconds(); amount > 0 {
		unit.AddMana(sim, amount, unit.manaCastingMetrics)
	}
	if amount := unit.manaRegenWhileNotCasting * (sim.CurrentTime - castingUntil).Seconds(); amount > 0 {
		unit.AddMana(sim, amount, unit.manaNotCastingMetrics)
	}
}

func (mb *manaBar) RunTask(sim *Simulation) time.Duration {
	if sim.CurrentTime < mb.nextManaTick {
		return mb.nextManaTick
	}
	mb.creditManaRegen(sim)
	if mb.unit.IsEnabled() {
		mb.agent.OnManaTick(sim)
	}
	mb.nextManaTick = sim.CurrentTime + manaTickDuration
	return mb.nextManaTick
}

// Starts the ticks, with the regen counting from the start of the prepull.
func (mb *manaBar) enable(sim *Simulation, agent Agent) {
	mb.agent = agent
	mb.lastManaTick = sim.Environment.PrepullStartTime()
	mb.nextManaTick = mb.lastManaTick + manaTickDuration
	sim.AddTask(mb)
	sim.RescheduleTask(mb.nextManaTick)
}

// Whether the unit is channeling a spell it paid mana for. The five seconds count from the channel's
// start, but regen stays at the casting rate until the channel ends, however it ends.
func (unit *Unit) channelPaidMana() bool {
	return unit.ChanneledDot != nil && unit.ChanneledDot.paidMana
}

// Whether the starting channel's cast paid mana: its own spell's, or the spell whose RelatedDotSpell it is.
func (mb *manaBar) claimChannelPayment(dot *Dot) bool {
	paidBy := mb.paidBy
	mb.paidBy = nil
	return paidBy != nil && (paidBy == dot.Spell || paidBy.RelatedDotSpell == dot.Spell)
}

// Returns how long the unit waits for its regen to reach desiredMana, up to the tick that credits it,
// assuming it spends no mana in the meantime. The regen runs at the casting rate for the rest of the
// five-second rule window or of a paid channel, whichever ends later, and at the full rate after it.
func (unit *Unit) TimeUntilManaRegen(sim *Simulation, desiredMana float64) time.Duration {
	manaNeeded := desiredMana - unit.CurrentMana()
	if manaNeeded <= 0 {
		return 0
	}

	window := max(unit.PseudoStats.FiveSecondRuleRefreshTime-sim.CurrentTime, 0)
	if unit.channelPaidMana() {
		window = max(window, unit.ChanneledDot.RemainingDuration(sim))
	}

	var regenTime time.Duration
	if castingMana := unit.manaRegenWhileCasting * window.Seconds(); manaNeeded <= castingMana {
		regenTime = DurationFromSeconds(manaNeeded / unit.manaRegenWhileCasting)
	} else if unit.manaRegenWhileNotCasting > 0 {
		regenTime = window + DurationFromSeconds((manaNeeded-castingMana)/unit.manaRegenWhileNotCasting)
	} else {
		return NeverExpires
	}

	reachedAt := sim.CurrentTime + regenTime
	if reachedAt <= unit.nextManaTick {
		return unit.nextManaTick - sim.CurrentTime
	}
	ticks := (reachedAt - unit.nextManaTick + manaTickDuration - 1) / manaTickDuration
	return unit.nextManaTick + ticks*manaTickDuration - sim.CurrentTime
}

func (sim *Simulation) enableManaTicks() {
	for _, party := range sim.Raid.Parties {
		for _, player := range party.Players {
			character := player.GetCharacter()
			if character.HasManaBar() {
				character.manaBar.enable(sim, player)
			}

			for _, petAgent := range character.PetAgents {
				if pet := petAgent.GetPet(); pet.HasManaBar() {
					pet.manaBar.enable(sim, petAgent)
				}
			}
		}
	}
}

func (mb *manaBar) reset() {
	if mb.unit == nil {
		return
	}

	mb.currentMana = mb.unit.MaxMana()

	mb.manaRegenMultiplier = 1.0

	mb.oom = false
	mb.paidBy = nil
}

func (mb *manaBar) IsOOM() bool {
	return mb.oom
}

// Pets keep no OOM metrics, so they never go OOM.
func (mb *manaBar) StartOOMEvent(sim *Simulation) {
	if mb.oom || mb.unit.Type == PetUnit {
		return
	}
	mb.oom = true
	mb.oomStartedAt = sim.CurrentTime
	mb.unit.Metrics.MarkOOM(sim)
}
func (mb *manaBar) EndOOMEvent(sim *Simulation) {
	if !mb.oom {
		return
	}
	mb.oom = false
	mb.unit.Metrics.AddOOMTime(sim, sim.CurrentTime-mb.oomStartedAt)
}

type ManaCostOptions struct {
	BaseCostPercent float64 // The cost of the spell as a percentage (0-100) of the unit's base mana.
	FlatCost        int32   // Alternative to BaseCostPercent for giving a flat value.
	PercentModifier float64 // Will default to 1. PercentModifier stored as a float i.e. 40% reduction is (0.6 multiplier) to the base cost
}
type ManaCost struct {
	ResourceMetrics *ResourceMetrics
}

func newManaCost(spell *Spell, options ManaCostOptions) *SpellCost {
	return &SpellCost{
		spell:                   spell,
		BaseCost:                TernaryInt32(options.FlatCost > 0, options.FlatCost, int32(options.BaseCostPercent*spell.Unit.BaseMana)/100),
		PercentModifier:         TernaryFloat64(options.PercentModifier == 0, 1, options.PercentModifier),
		AdditivePercentModifier: 1,
		ResourceCostImpl: &ManaCost{
			ResourceMetrics: spell.Unit.NewManaMetrics(spell.ActionID),
		},
	}
}

// A running hardcast pays its cost when it completes; the mana left after that is what counts.
func (unit *Unit) manaAfterHardcast(sim *Simulation) float64 {
	mana := unit.CurrentMana()
	if unit.hardcastRuns(sim) {
		mana -= unit.Hardcast.Cost
	}
	return mana
}

// For a rotation that checks a mana cost itself: the current pass skips a spell it wants for its cost.
func (mb *manaBar) SkipForMana() {
	mb.costBlocked = true
}

func (mc *ManaCost) MeetsRequirement(sim *Simulation, spell *Spell) bool {
	spell.CurCast.Cost = spell.Cost.GetCurrentCost()
	return spell.CurCast.Cost <= 0 || spell.Unit.manaAfterHardcast(sim) >= spell.CurCast.Cost
}
func (mc *ManaCost) CostFailureReason(sim *Simulation, spell *Spell) string {
	return fmt.Sprintf("not enough mana (Mana After Cast = %0.03f, Mana Cost = %0.03f)", spell.Unit.manaAfterHardcast(sim), spell.CurCast.Cost)
}
func (mc *ManaCost) SpendCost(sim *Simulation, spell *Spell) {
	spell.Unit.paidBy = nil
	if spell.CurCast.Cost > 0 {
		if spell.Unit.PseudoStats.FiveSecondRuleRefreshTime < sim.CurrentTime {
			spell.Unit.creditManaRegen(sim)
		}
		spell.Unit.SpendMana(sim, spell.CurCast.Cost, mc.ResourceMetrics)
		spell.Unit.paidBy = spell
		spell.Unit.PseudoStats.FiveSecondRuleRefreshTime = max(sim.CurrentTime+time.Second*5, spell.Unit.Hardcast.Expires)
	}
}
func (mc *ManaCost) IssueRefund(_ *Simulation, _ *Spell) {}
