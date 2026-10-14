package buffs

// What sim/core/buffs builds a generated buff with where the spell parse does not
// reach: the damage a shield deals back, the external cooldown a driven buff is
// cast on, and the flat bonus a set adds to a buff.

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/stats"
)

// The damage a damage shield deals back to whoever lands a melee hit, which
// also scales with the wearer's spell power by bonusCoefficient; the generated
// shields state none. category is the aura's own, which decides whether a
// second copy of the shield can sit next to it.
func NewDamageShield(unit *core.Unit, label string, actionID core.ActionID, duration time.Duration, category string, singleAura bool,
	school core.SpellSchool, damage float64, bonusCoefficient float64) *core.Aura {
	trigger := damageShieldTrigger(unit, label, actionID, school, damage, bonusCoefficient)

	aura := unit.GetOrRegisterAura(core.Aura{
		Label:      label,
		ActionID:   actionID,
		Duration:   core.TernaryDuration(duration > 0, duration, core.NeverExpires),
		BuildPhase: core.Ternary(actionID.Tag == -1, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone),
	}).AttachProcTrigger(trigger)

	// The shield has no stat to apply or remove: what the category decides is
	// which aura keeps its proc trigger, so the damage is the whole bid.
	if category != "" {
		aura.NewExclusiveEffect(category, singleAura, core.ExclusiveEffect{Priority: damage})
	}
	return aura
}

// The same damage shield on an aura that already exists, live while the aura is up: an item's buff
// whose row states a damage shield beside its other effects. The spell that deals the damage takes
// the aura's action id with its tag raised by 2.
func AttachDamageShield(aura *core.Aura, school core.SpellSchool, damage float64, bonusCoefficient float64) {
	aura.AttachProcTrigger(damageShieldTrigger(aura.Unit, aura.Label, aura.ActionID, school, damage, bonusCoefficient))
}

func damageShieldTrigger(unit *core.Unit, label string, actionID core.ActionID, school core.SpellSchool, damage float64,
	bonusCoefficient float64) core.ProcTrigger {
	procSpell := unit.RegisterSpell(core.SpellConfig{
		ActionID:    actionID.WithTag(actionID.Tag + 2),
		SpellSchool: school,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagBinary | core.SpellFlagNoPartialResist | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: bonusCoefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHit)
		},
	})

	return core.ProcTrigger{
		Name:     label + " Damage",
		Callback: core.CallbackOnSpellHitTaken,
		Outcome:  core.OutcomeLanded,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellSchool.Matches(core.SpellSchoolPhysical) {
				procSpell.Cast(sim, spell.Unit)
			}
		},
	}
}

// What a driver decides about a buff other players cast on this one: how many
// of them take turns, how long each waits before casting again, whether the
// cooldown is worth a mana, a damage or a survival slot, and when the sim
// should spend it.
type generatedExternalCD struct {
	NumSources     int32
	Cooldown       time.Duration
	Type           core.CooldownType
	ShouldActivate core.CooldownActivationCondition
}

// The external cooldown around a generated aura, approximated by NumSources
// casters taking turns. The aura carries the buff's own label, duration and
// effects; its tag is what tells the sim the buff is already up, so a row whose
// cooldown is driven states a category for the generator to tag it with.
func newGeneratedExternalCD(char *core.Character, aura *core.Aura, config generatedExternalCD) {
	if config.NumSources == 0 {
		return
	}

	// The external caster's copy of a permanent buff is up while the character
	// sheet is measured; a cooldown is not, so it keeps its effects out of the
	// stats the build phase collects.
	aura.BuildPhase = core.CharacterBuildPhaseNone

	registerExternalConsecutiveCDApproximation(char, externalConsecutiveCDApproximation{
		ActionID:         aura.ActionID,
		AuraTag:          aura.Tag,
		CooldownPriority: core.CooldownPriorityDefault,
		Type:             config.Type,
		AuraDuration:     aura.Duration,
		AuraCD:           config.Cooldown,
		ShouldActivate:   config.ShouldActivate,
		AddAura:          func(sim *core.Simulation, _ *core.Character) { aura.Activate(sim) },
		RelatedSelfBuff:  aura,
	}, config.NumSources)
}

// The second category the aura joins without an effect of its own, which is how
// the paladin auras exclude each other across schools. Only the player's own
// copy joins it: the external copy has to be able to sit next to the one the
// player casts.
func joinSharedCategory(aura *core.Aura, shared string, isPlayer bool) {
	if shared == "" || !isPlayer {
		return
	}
	aura.NewExclusiveEffect(shared, true, core.ExclusiveEffect{})
}

// AddGeneratedFlatBonus raises a generated buff the game files state as worth base
// to base+bonus, in both of the numbers a buff that does not stack keeps apart:
// what it applies and what it bids for its category. Three pieces of the
// warrior's tier 2 set are worth 30 more attack power on Battle Shout, and a
// shout worth 169 has to outbid one worth 139. The aura grants the extra amount
// for exactly as long as it holds the category, because a category that holds
// one aura at a time deactivates the copy it outbids.
//
// The aura belongs to the unit rather than to whoever raised it, so two party
// members wearing the same set ask for the same total and get it once: a call
// against an aura that already carries the bonus does nothing. One that finds
// some other amount there is reading a different buff's base and says so.
//
// The category has to hold one aura at a time. That is what makes attaching the
// extra amount to the aura the same thing as attaching it to the effect: the
// copy that loses the category is deactivated outright, so the aura is up
// exactly while its effect is the one applying.
//
// It writes Priority directly, which is only safe before the fight: an effect
// that is already active has to go through SetPriority to re-apply what it
// holds. Every caller runs while the character is being built.
func AddGeneratedFlatBonus(aura *core.Aura, stat stats.Stat, base float64, bonus float64) {
	if aura.MaxStacks > 0 {
		panic("a stacking aura re-prices its category effect on every stack, which would drop the bonus: " + aura.Label)
	}

	for _, effect := range aura.ExclusiveEffects {
		if effect.Category.Name != aura.Tag {
			continue
		}
		if !effect.Category.SingleAura {
			panic("a category that holds more than one aura cannot tell a flat bonus when to apply: " + aura.Label)
		}
		if effect.Priority == base+bonus {
			return
		}
		if effect.Priority != base {
			panic(fmt.Sprintf("%s bids %v, which is neither the %v it is worth nor the %v the bonus makes it",
				aura.Label, effect.Priority, base, base+bonus))
		}
		effect.Priority = base + bonus
		aura.AttachStatBuff(stat, bonus)
		return
	}

	panic("a flat bonus needs a buff that bids for its whole category: " + aura.Label)
}

// Used for approximating cooldowns applied by other players to you, such as
// innervate, power infusion, etc. This is specifically for buffs
// which can be consecutively applied multiple times to a single player.
type externalConsecutiveCDApproximation struct {
	ActionID         core.ActionID
	AuraTag          string
	CooldownPriority int32
	Type             core.CooldownType
	AuraDuration     time.Duration
	AuraCD           time.Duration

	// Callback for extra activation conditions.
	ShouldActivate core.CooldownActivationCondition

	// Applies the buff.
	AddAura           core.CooldownActivation
	RelatedSelfBuff   *core.Aura             // Used to attach the aura to the generic spell
	RelatedAuraArrays core.LabeledAuraArrays // Used to attach the aura to the generic spell
}

// numSources is the number of other players assigned to apply the buff to this player.
// E.g. the number of other druids in the group casting innervate.
func registerExternalConsecutiveCDApproximation(char *core.Character, config externalConsecutiveCDApproximation, numSources int32) {
	if numSources == 0 {
		panic("Need at least 1 source!")
	}

	var nextExternalIndex int

	externalTimers := make([]*core.Timer, numSources)
	for i := 0; i < int(numSources); i++ {
		externalTimers[i] = char.NewTimer()
	}
	sharedTimer := char.NewTimer()

	spell := char.RegisterSpell(core.SpellConfig{
		ActionID: config.ActionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagNoMetrics | core.SpellFlagNoLogs,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    sharedTimer,
				Duration: config.AuraDuration, // Assumes that multiple buffs are different sources.
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			if !externalTimers[nextExternalIndex].IsReady(sim) {
				return false
			}

			if char.HasActiveAuraWithTag(config.AuraTag) {
				return false
			}

			return true
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			config.AddAura(sim, char)
			externalTimers[nextExternalIndex].Set(sim.CurrentTime + config.AuraCD)

			nextExternalIndex = (nextExternalIndex + 1) % len(externalTimers)

			if externalTimers[nextExternalIndex].IsReady(sim) {
				sharedTimer.Set(sim.CurrentTime + config.AuraDuration)
			} else {
				sharedTimer.Set(sim.CurrentTime + externalTimers[nextExternalIndex].TimeToReady(sim))
			}
		},
		RelatedSelfBuff:   config.RelatedSelfBuff,
		RelatedAuraArrays: config.RelatedAuraArrays,
	})

	char.AddMajorCooldown(core.MajorCooldown{
		Spell:    spell,
		Priority: config.CooldownPriority,
		Type:     config.Type,

		ShouldActivate: config.ShouldActivate,
	})
}
