package core

import (
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	CooldownPriorityLow     = -1000
	CooldownPriorityDefault = 0
)

type CooldownType byte

const (
	CooldownTypeUnknown CooldownType = 0
	CooldownTypeMana    CooldownType = 1 << iota
	CooldownTypeDPS
	CooldownTypeExplosive
	CooldownTypeSurvival
	CooldownTypeHealing
)

func (ct CooldownType) Matches(other CooldownType) bool {
	return (ct & other) != 0
}

// Condition for whether a cooldown can/should be activated.
// Returning false prevents the cooldown from being activated.
type CooldownActivationCondition func(*Simulation, *Character) bool

// Function for activating a cooldown.
// Returns whether the activation was successful.
type CooldownActivation func(*Simulation, *Character)

type MajorCooldown struct {
	// Spell that is cast when this MCD is activated.
	Spell *Spell

	// Associated buff aura that is activated when the above spell is cast.
	BuffAura *StatBuffAura

	// Cooldowns with higher priority get used first. This is important when some
	// cooldowns have a non-zero cast time, which should go before an instant one.
	Priority int32

	// Internal category, used for filtering. For example, mages want to disable
	// all DPS cooldowns during their regen rotation.
	Type CooldownType

	// If False, the MCD will not be queued up before the GCD has fully expired.
	// This is desirable for auto-chaining multiple MCDs just before a GCD ability.
	AllowSpellQueueing bool

	// Whether the cooldown meets all optional conditions for activation. These
	// conditions will be ignored when the user specifies their own activation time.
	// This is for things like mana thresholds, which are optimizations for better
	// automatic timing.
	ShouldActivate CooldownActivationCondition

	// Autocast Other Cooldowns uses it only at a timing set in the cooldown settings; otherwise only the APL casts it.
	DisableAutocast bool

	// Fixed timings at which to use this cooldown. If these are specified, they
	// are used instead of ShouldActivate.
	timings []time.Duration

	// Number of times this MCD was used so far in the current iteration.
	numUsages int
}

// Whether Autocast Other Cooldowns can cast it: at a timing, or on its own.
func (mcd *MajorCooldown) autocasts() bool {
	return len(mcd.timings) > 0 || !mcd.DisableAutocast
}

func (mcd *MajorCooldown) ReadyAt() time.Duration {
	return mcd.Spell.ReadyAt()
}

func (mcd *MajorCooldown) IsReady(sim *Simulation) bool {
	return mcd.Spell.IsReady(sim)
}

func (mcd *MajorCooldown) TimeToReady(sim *Simulation) time.Duration {
	return mcd.Spell.TimeToReady(sim)
}

// Roughly how long until the next cast will happen, accounting for both spell CD and user-specified timings.
func (mcd *MajorCooldown) TimeToNextCast(sim *Simulation) time.Duration {
	timeToReady := mcd.TimeToReady(sim)
	if mcd.numUsages < len(mcd.timings) {
		timeToReady = max(timeToReady, mcd.timings[mcd.numUsages]-sim.CurrentTime)
	}
	return timeToReady
}

func (mcd *MajorCooldown) GetTimings() []time.Duration {
	return mcd.timings
}

// Uses this cooldown if its conditions pass, and returns whether it did.
func (mcd *MajorCooldown) TryActivate(sim *Simulation, character *Character) bool {
	return mcd.tryActivateHelper(sim, character)
}

func (mcd *MajorCooldown) shouldActivateHelper(sim *Simulation, character *Character) bool {
	if !mcd.Spell.CanCast(sim, character.CurrentTarget) {
		return false
	}

	// A major cooldown never leaves the form, timed or not; an APL cast of its spell does.
	if mcd.Spell.hasCastRequirement {
		if _, unshift := mcd.Spell.castRequirementFailure(sim, character.CurrentTarget); unshift {
			return false
		}
	}

	if mcd.numUsages < len(mcd.timings) {
		return sim.CurrentTime >= mcd.timings[mcd.numUsages]
	}

	if mcd.Type.Matches(CooldownTypeSurvival) && character.cooldownConfigs.HpPercentForDefensives != 0 {
		if character.CurrentHealthPercent() > character.cooldownConfigs.HpPercentForDefensives {
			return false
		}
	}

	return mcd.ShouldActivate(sim, character)
}

// Activates this MCD, if all the conditions pass.
// Returns whether the MCD was activated.
func (mcd *MajorCooldown) tryActivateHelper(sim *Simulation, character *Character) bool {
	shouldActivate := mcd.shouldActivateHelper(sim, character)

	if shouldActivate {
		character.setRotationCastEntryPath(character.currentEntryPath())
		if mcd.Spell.Flags.Matches(SpellFlagHelpful) {
			mcd.Spell.castForRotation(sim, &character.Unit, nil)
		} else {
			mcd.Spell.castForRotation(sim, character.CurrentTarget, nil)
		}

		mcd.numUsages++
		if sim.Log != nil {
			character.Log(sim, "Major cooldown used: %s", mcd.Spell.ActionID)
		}
	}

	return shouldActivate
}

type cooldownConfigs struct {
	Cooldowns              []*proto.Cooldown
	HpPercentForDefensives float64
}

type majorCooldownManager struct {
	// The Character whose cooldowns are being managed.
	character *Character

	// User-specified cooldown configs.
	cooldownConfigs cooldownConfigs

	// Cached list of major cooldowns sorted by priority, for resetting quickly.
	initialMajorCooldowns []MajorCooldown

	// Major cooldowns, ordered by next available. This should always contain
	// the same cooldowns as initialMajorCooldowns, but the order will change over
	// the course of the sim.
	majorCooldowns []*MajorCooldown
	minReady       time.Duration
}

func newMajorCooldownManager(cooldowns *proto.Cooldowns) majorCooldownManager {
	if cooldowns == nil {
		return majorCooldownManager{}
	}

	cooldownConfigs := cooldownConfigs{
		HpPercentForDefensives: cooldowns.HpPercentForDefensives,
	}
	for _, cooldownConfig := range cooldowns.Cooldowns {
		if cooldownConfig.Id != nil {
			cooldownConfigs.Cooldowns = append(cooldownConfigs.Cooldowns, cooldownConfig)
		}
	}

	return majorCooldownManager{
		cooldownConfigs: cooldownConfigs,
	}
}

func (mcdm *majorCooldownManager) initialize(character *Character) {
	mcdm.character = character
}

func (mcdm *majorCooldownManager) finalize() {
	// Match user-specified cooldown configs to existing cooldowns.
	for i := range mcdm.initialMajorCooldowns {
		mcd := &mcdm.initialMajorCooldowns[i]
		mcd.timings = []time.Duration{}

		for _, cooldownConfig := range mcdm.cooldownConfigs.Cooldowns {
			configID := ProtoToActionID(cooldownConfig.Id)
			if configID.SameAction(mcd.Spell.ActionID) {
				mcd.timings = make([]time.Duration, len(cooldownConfig.Timings))
				for t, timing := range cooldownConfig.Timings {
					mcd.timings[t] = DurationFromSeconds(timing)
				}
				break
			}
		}
	}

	mcdm.majorCooldowns = make([]*MajorCooldown, len(mcdm.initialMajorCooldowns))
}

func (mcdm *majorCooldownManager) reset(_ *Simulation) {
	for i := range mcdm.majorCooldowns {
		newMCD := &MajorCooldown{}
		*newMCD = mcdm.initialMajorCooldowns[i]
		mcdm.majorCooldowns[i] = newMCD
	}

	// For initial sorting.
	mcdm.UpdateMajorCooldowns()
}

// Registers a major cooldown to the Character, which will be automatically
// used when available.
func (mcdm *majorCooldownManager) AddMajorCooldown(mcd MajorCooldown) {
	if mcdm.character.Env.IsFinalized() {
		panic("Major cooldowns may not be added once finalized!")
	}
	if mcd.Spell == nil {
		panic("Major cooldown must have a Spell!")
	}

	if mcd.Spell.ActionID.ItemID != 0 {
		mcdm.character.ItemSwap.RegisterActive(mcd.Spell.ActionID.ItemID)
	}

	mcd.Spell.Flags |= SpellFlagAPL | SpellFlagMCD

	if (mcd.Type.Matches(CooldownTypeSurvival) && (mcd.Spell.DefaultCast.EffectiveTime() == 0)) || mcd.AllowSpellQueueing {
		mcd.Spell.Flags |= SpellFlagReactive
	}

	// A defensive cooldown that adds no damage or mana goes on its own only for a tank; a timing or an APL
	// action still uses it for anyone. The tank slots are read before any cooldown is added.
	defensiveOnly := mcd.Type.Matches(CooldownTypeSurvival) && !mcd.Type.Matches(CooldownTypeDPS|CooldownTypeMana)
	if defensiveOnly && !mcdm.character.isTank {
		mcd.DisableAutocast = true
	}
	if mcd.DisableAutocast {
		mcd.ShouldActivate = func(_ *Simulation, _ *Character) bool { return false }
	} else if mcd.ShouldActivate == nil {
		mcd.ShouldActivate = func(_ *Simulation, _ *Character) bool { return true }
	}

	mcdm.initialMajorCooldowns = append(mcdm.initialMajorCooldowns, mcd)
}

func (mcdm *majorCooldownManager) GetInitialMajorCooldown(actionID ActionID) MajorCooldown {
	for _, mcd := range mcdm.initialMajorCooldowns {
		if mcd.Spell.SameAction(actionID) {
			return mcd
		}
	}
	return MajorCooldown{}
}

func (mcdm *majorCooldownManager) removeInitialMajorCooldown(actionID ActionID) {
	// First, find and remove the exact match.
	removedIdx := -1
	var removedTimer *Timer
	for i, mcd := range mcdm.initialMajorCooldowns {
		if mcd.Spell.SameAction(actionID) {
			removedTimer = mcd.Spell.CD.Timer
			mcdm.initialMajorCooldowns = append(mcdm.initialMajorCooldowns[:i], mcdm.initialMajorCooldowns[i+1:]...)
			mcdm.majorCooldowns = mcdm.majorCooldowns[:len(mcdm.majorCooldowns)-1]
			removedIdx = i
			break
		}
	}

	if removedIdx < 0 || removedTimer == nil {
		return
	}

	// Also remove any other MCDs that share the same CD timer pointer (e.g.
	// Berserking 10%/15%/20%/25%/30% all share one timer).  This prevents
	// auto-casting a different variant while the player controls one manually.
	// Spells with independent timers (e.g. two trinkets that share no cooldown)
	// are left alone.
	for i := len(mcdm.initialMajorCooldowns) - 1; i >= 0; i-- {
		mcd := &mcdm.initialMajorCooldowns[i]
		if mcd.Spell.CD.Timer == removedTimer && mcd.Spell.SameActionIgnoreTag(actionID) {
			mcdm.initialMajorCooldowns = append(mcdm.initialMajorCooldowns[:i], mcdm.initialMajorCooldowns[i+1:]...)
			mcdm.majorCooldowns = mcdm.majorCooldowns[:len(mcdm.majorCooldowns)-1]
		}
	}
}

func (mcdm *majorCooldownManager) GetMajorCooldown(actionID ActionID) *MajorCooldown {
	for _, mcd := range mcdm.majorCooldowns {
		if mcd.Spell.SameAction(actionID) {
			return mcd
		}
	}
	return nil
}
func (mcdm *majorCooldownManager) GetMajorCooldownIgnoreTag(actionID ActionID) *MajorCooldown {
	for _, mcd := range mcdm.majorCooldowns {
		if mcd.Spell.SameActionIgnoreTag(actionID) {
			return mcd
		}
	}
	return nil
}

// Returns all MCDs.
func (mcdm *majorCooldownManager) GetMajorCooldowns() []*MajorCooldown {
	return mcdm.majorCooldowns
}

func (mcdm *majorCooldownManager) GetMajorCooldownIDs() []*proto.ActionID {
	ids := make([]*proto.ActionID, len(mcdm.initialMajorCooldowns))
	for i, mcd := range mcdm.initialMajorCooldowns {
		ids[i] = mcd.Spell.ActionID.ToProto()
	}
	return ids
}

func (mcdm *majorCooldownManager) GetMatchingStatBuffSpells(statTypesToMatch []stats.Stat) []*Spell {
	matchingSpells := make([]*Spell, 0, len(mcdm.initialMajorCooldowns))

	for _, mcd := range mcdm.initialMajorCooldowns {
		if mcd.BuffAura.BuffsMatchingStat(statTypesToMatch) && (mcd.Spell.DefaultCast.EffectiveTime() == 0) {
			matchingSpells = append(matchingSpells, mcd.Spell)
		}
	}

	return matchingSpells
}

func (mcdm *majorCooldownManager) GetMatchingStatBuffCooldownAuras(statTypesToMatch []stats.Stat) []*StatBuffAura {
	matchingAuras := make([]*StatBuffAura, 0, len(mcdm.initialMajorCooldowns))

	for _, mcd := range mcdm.initialMajorCooldowns {
		if mcd.BuffAura.BuffsMatchingStat(statTypesToMatch) && (mcd.Spell.DefaultCast.EffectiveTime() == 0) {
			matchingAuras = append(matchingAuras, mcd.BuffAura)
		}
	}

	return matchingAuras
}

// The first ready cooldown, by ready time and then priority, whose own conditions pass and that may
// fire at this point of the GCD: an explosive only while it runs and, when the shortest GCD would cover
// its cast, only while what is left of it does, any other once it is ready, a reactive one either way.
func (mcdm *majorCooldownManager) getFirstReadyMCD(sim *Simulation) *MajorCooldown {
	if sim.CurrentTime < mcdm.minReady {
		return nil
	}

	gcdLeft := mcdm.character.GCD.TimeToReady(sim)
	for _, mcd := range mcdm.majorCooldowns {
		if !mcd.autocasts() || !mcd.IsReady(sim) {
			continue
		}
		// Explicitly check for GCD because MCDs are usually desired to be cast immediately
		// before the next spell, rather than immediately after the previous spell. This is
		// true even for MCDs which do not require the GCD. The one exception to this rule
		// is explosives, which should instead be cast right *after* incurring a GCD, since
		// they are off-GCD but have small cast times. One whose cast fits in the shortest GCD
		// also waits while less of the GCD is left than its cast, so that it never delays the
		// next spell; a longer one, which the shortest GCD would not cover, goes whenever the
		// GCD runs.
		switch {
		case mcd.Spell.Flags.Matches(SpellFlagReactive):
		case mcd.Type.Matches(CooldownTypeExplosive):
			if castTime := mcd.Spell.CastTime(); gcdLeft == 0 || gcdLeft < castTime && castTime <= GCDMin {
				continue
			}
		case gcdLeft > 0:
			continue
		}
		if mcd.shouldActivateHelper(sim, mcdm.character) {
			return mcd
		}
	}

	return nil
}

// Re-sorts the cooldowns and caches the earliest time autocast can use one. Autocast calls it after each
// use, and anything else that changes a major cooldown's ready time must call it too.
func (mcdm *majorCooldownManager) UpdateMajorCooldowns() {
	mcdm.sort()
	mcdm.minReady = NeverExpires
	if i := slices.IndexFunc(mcdm.majorCooldowns, (*MajorCooldown).autocasts); i >= 0 {
		mcdm.minReady = mcdm.majorCooldowns[i].ReadyAt()
	}
}

func (mcdm *majorCooldownManager) sort() {
	slices.SortStableFunc(mcdm.majorCooldowns, func(m1, m2 *MajorCooldown) int {
		if r1, r2 := m1.ReadyAt(), m2.ReadyAt(); r1 != r2 {
			return int(r1 - r2)
		}
		return int(m2.Priority - m1.Priority)
	})
}

// Add a major cooldown to the given agent, which provides a temporary boost to a single stat.
// This is use for effects like Icon of the Silver Crescent and Bloodlust Brooch.
func RegisterTemporaryStatsOnUseCD(character *Character, auraLabel string, tempStats stats.Stats, duration time.Duration, config SpellConfig) *StatBuffAura {
	aura := character.NewTemporaryStatsAura(auraLabel, config.ActionID, tempStats, duration)
	cdType := aura.InferCDType()
	if config.Cast.CD.Duration > 0 {
		aura.Icd = &config.Cast.CD
	}

	config.Flags |= SpellFlagNoOnCastComplete
	config.ApplyEffects = func(sim *Simulation, _ *Unit, _ *Spell) {
		aura.Activate(sim)
	}

	spell := character.RegisterSpell(config)
	spell.RelatedSelfBuff = aura.Aura

	character.AddMajorCooldown(MajorCooldown{
		Spell:    spell,
		Type:     cdType,
		BuffAura: aura,
	})

	return aura
}
