package core

import (
	"fmt"
	"math"
	"slices"
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

	manaRegenMultiplier          float64
	innervateManaRegenMultiplier float64

	spiritRegenBase      float64
	spiritRegenPerSpirit float64

	manaCastingMetrics    *ResourceMetrics
	manaNotCastingMetrics *ResourceMetrics

	manaRegenBoosts       map[*Aura]*ManaRegenBoost
	activeManaRegenBoosts []*ManaRegenBoost // In the order they were gained.
	// The regen per second without the active boosts, as of the last rate change.
	unboostedRegenWhileCasting    float64
	unboostedRegenWhileNotCasting float64

	JowManaMetrics    *ResourceMetrics
	VtManaMetrics     *ResourceMetrics
	JowiseManaMetrics *ResourceMetrics
	manaGainSpell     *Spell // Holds the threat mana gains make.

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
	mb.manaGainSpell.ApplySplitThreat(sim, gain*ThreatPerManaGained)
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

// Innervate multiplies Spirit regen and MP5, but a periodic energize aura ticks its own amount under
// it (EV-0114). Taking Innervate's surplus on that share back off keeps a
// factor of 1 bit-exact. Skyborne Energized's multiplier scales all of it.
func (unit *Unit) multipliedManaRegen(spiritRegen float64, innervate float64) float64 {
	return ((unit.MP5ManaRegenPerSecond()+spiritRegen)*innervate - unit.PseudoStats.EnergizeMP5/5.0*(innervate-1)) * unit.manaRegenMultiplier
}

// Returns the rate of mana regen per second, assuming this unit is
// considered to be casting.
func (unit *Unit) ManaRegenPerSecondWhileCasting() float64 {
	// The shares of spirit regen that continue while casting add up to at most all of it.
	return unit.multipliedManaRegen(unit.SpiritManaRegenPerSecond()*unit.PseudoStats.SpiritRegenMultiplier*min(unit.PseudoStats.SpiritRegenRateCasting, 1),
		unit.innervateManaRegenMultiplier)
}

// Returns the rate of mana regen per second, assuming this unit is
// considered to be not casting.
func (unit *Unit) ManaRegenPerSecondWhileNotCasting() float64 {
	return unit.multipliedManaRegen(unit.SpiritManaRegenPerSecond()*unit.PseudoStats.SpiritRegenMultiplier, unit.innervateManaRegenMultiplier)
}

// Recomputes the regen rates the mana bar ticks at, after crediting the regen so far at the old ones.
// Without a Simulation there is no regen to credit yet.
func (unit *Unit) UpdateManaRegenRates(sim *Simulation) {
	if sim != nil {
		unit.creditManaRegen(sim)
	}
	unit.manaRegenWhileCasting = unit.ManaRegenPerSecondWhileCasting()
	unit.manaRegenWhileNotCasting = unit.ManaRegenPerSecondWhileNotCasting()
	unit.updateUnboostedManaRegen()
}

// A buff that raises the mana regen for a while, such as Innervate or Evocation. The regen it adds
// goes to the buff's own metrics, so the regen rows show the unit's own regen and the buff's row shows
// what it added. The mana the unit gains is the same either way.
type ManaRegenBoost struct {
	metrics *ResourceMetrics

	SpiritMult   float64 // Its factor on PseudoStats.SpiritRegenMultiplier.
	Innervate    float64 // Its factor on the Innervate multiplier.
	CastingShare float64 // What it adds to PseudoStats.SpiritRegenRateCasting.

	// The regen per second it adds, as of the last rate change.
	addedWhileCasting    float64
	addedWhileNotCasting float64
}

// The tag on a boost's metrics. The aura often shares its spell's action ID, whose mana cost has
// metrics of its own (a druid's own Innervate), and two metrics under one ID do not combine.
const manaRegenBoostTag = 99997

// The boost the aura gives, made on the first call. Its metrics take the aura's spell, tagged
// manaRegenBoostTag, and auras of one spell (the player's copy, the external one) share them. A caller
// that changes a regen field while the aura is up tells the boost the same change.
func (unit *Unit) ManaRegenBoost(aura *Aura) *ManaRegenBoost {
	if boost := unit.manaRegenBoosts[aura]; boost != nil {
		return boost
	}
	id := aura.ActionID.WithTag(manaRegenBoostTag)
	var metrics *ResourceMetrics
	for _, other := range unit.manaRegenBoosts {
		if other.metrics.ActionID == id {
			metrics = other.metrics
			break
		}
	}
	if metrics == nil {
		metrics = unit.NewManaMetrics(id)
	}
	boost := &ManaRegenBoost{metrics: metrics, SpiritMult: 1, Innervate: 1}
	if unit.manaRegenBoosts == nil {
		unit.manaRegenBoosts = make(map[*Aura]*ManaRegenBoost)
	}
	unit.manaRegenBoosts[aura] = boost

	// The fields may change before the aura counts as up (an exclusive category applies them first) or
	// after, so listing and unlisting the boost credits the regen so far under the old split and then
	// recomputes it.
	aura.ApplyOnGain(func(_ *Aura, sim *Simulation) {
		unit.creditManaRegen(sim)
		unit.activeManaRegenBoosts = append(unit.activeManaRegenBoosts, boost)
		unit.UpdateManaRegenRates(sim)
	}).ApplyOnExpire(func(_ *Aura, sim *Simulation) {
		unit.creditManaRegen(sim)
		unit.activeManaRegenBoosts = slices.DeleteFunc(unit.activeManaRegenBoosts, func(b *ManaRegenBoost) bool { return b == boost })
		unit.UpdateManaRegenRates(sim)
	})
	return boost
}

// Takes the active boosts' factors back out of the regen rates, the most recently gained first: each
// boost adds the difference its own factors make on top of the boosts gained before it, and what is
// left is the unit's own regen. A boost whose factors are at 1 adds nothing, and so does one that lowers
// the regen: what it takes comes off the boosts gained before it and then off the unit's own regen, so
// the rows still add up to the regen.
func (unit *Unit) updateUnboostedManaRegen() {
	casting, notCasting := unit.manaRegenWhileCasting, unit.manaRegenWhileNotCasting
	spiritMult, share, innervate := unit.PseudoStats.SpiritRegenMultiplier, unit.PseudoStats.SpiritRegenRateCasting, unit.innervateManaRegenMultiplier
	spiritRegenBase := unit.SpiritManaRegenPerSecond()
	lostCasting, lostNotCasting := 0.0, 0.0
	for i := len(unit.activeManaRegenBoosts) - 1; i >= 0; i-- {
		boost := unit.activeManaRegenBoosts[i]
		spiritMult /= boost.SpiritMult
		share -= boost.CastingShare
		innervate /= boost.Innervate
		spiritRegen := spiritRegenBase * spiritMult
		withoutCasting := unit.multipliedManaRegen(spiritRegen*min(max(share, 0), 1), innervate)
		withoutNotCasting := unit.multipliedManaRegen(spiritRegen, innervate)
		addedCasting, addedNotCasting := casting-withoutCasting+lostCasting, notCasting-withoutNotCasting+lostNotCasting
		boost.addedWhileCasting, boost.addedWhileNotCasting = max(addedCasting, 0), max(addedNotCasting, 0)
		lostCasting, lostNotCasting = min(addedCasting, 0), min(addedNotCasting, 0)
		casting, notCasting = withoutCasting, withoutNotCasting
	}
	unit.unboostedRegenWhileCasting, unit.unboostedRegenWhileNotCasting = casting+lostCasting, notCasting+lostNotCasting
}

func (unit *Unit) MultiplyManaRegenSpeed(sim *Simulation, multiplier float64) {
	unit.manaRegenMultiplier *= multiplier
	unit.UpdateManaRegenRates(sim)
}

func (unit *Unit) MultiplyInnervateManaRegenSpeed(sim *Simulation, multiplier float64) {
	unit.innervateManaRegenMultiplier *= multiplier
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
	if seconds := (castingUntil - from).Seconds(); seconds > 0 {
		mb.addRegen(sim, seconds, true)
	}
	if seconds := (sim.CurrentTime - castingUntil).Seconds(); seconds > 0 {
		mb.addRegen(sim, seconds, false)
	}
}

// Adds the regen of these seconds as one gain, capped at max mana once, and splits what it records:
// each active boost's addition to its metrics and the rest to the regen row, each with its share of
// what the cap let through.
func (mb *manaBar) addRegen(sim *Simulation, seconds float64, casting bool) {
	unit := mb.unit
	amount, own, metrics := unit.manaRegenWhileNotCasting*seconds, unit.unboostedRegenWhileNotCasting*seconds, unit.manaNotCastingMetrics
	if casting {
		amount, own, metrics = unit.manaRegenWhileCasting*seconds, unit.unboostedRegenWhileCasting*seconds, unit.manaCastingMetrics
	}
	if amount <= 0 {
		return
	}
	if len(mb.activeManaRegenBoosts) == 0 || own >= amount {
		unit.AddMana(sim, amount, metrics)
		return
	}
	own = max(own, 0)

	oldMana := mb.currentMana
	newMana := min(oldMana+amount, unit.MaxMana())
	share := (newMana - oldMana) / amount
	metrics.AddEvent(own, own*share)
	for _, boost := range mb.activeManaRegenBoosts {
		added := boost.addedWhileNotCasting * seconds
		if casting {
			added = boost.addedWhileCasting * seconds
		}
		if added > 0 {
			boost.metrics.AddEvent(added, added*share)
		}
	}

	if sim.Log != nil {
		unit.Log(sim, "Gained %0.3f mana from %s, %0.3f of it from regen boosts (%0.3f --> %0.3f) of %0.0f total.",
			amount, metrics.ActionID, amount-own, oldMana, newMana, unit.MaxMana())
	}

	mb.currentMana = newMana
	unit.Metrics.ManaGained += newMana - oldMana
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
	mb.innervateManaRegenMultiplier = 1.0

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
	BaseCostPercent float64 // The cost of the spell as a percentage (0-100) of the unit's base mana, rounded to the nearest whole mana.
	FlatCost        int32   // Alternative to BaseCostPercent for giving a flat value.
	PercentModifier float64 // Will default to 1. PercentModifier stored as a float i.e. 40% reduction is (0.6 multiplier) to the base cost
}
type ManaCost struct {
	ResourceMetrics *ResourceMetrics
}

func newManaCost(spell *Spell, options ManaCostOptions) *SpellCost {
	return &SpellCost{
		spell:                   spell,
		BaseCost:                TernaryInt32(options.FlatCost > 0, options.FlatCost, int32(math.Round(options.BaseCostPercent*spell.Unit.BaseMana/100))),
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
