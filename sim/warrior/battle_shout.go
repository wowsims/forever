package warrior

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// How close to running out the warrior's own shout has to be before recasting it is worth a global.
const ShoutExpirationThreshold = time.Second * 3

var battleShoutRank = spellData.BattleShout.Highest()

func (warrior *Warrior) registerBattleShout() {
	// A warrior that shouts builds its own copy. One that never shouts gets the
	// isPlayer=false constructor, whose aura is the raid's external copy, and the
	// raid's Battle Shout state sets what that copy is worth.
	castsOwnShout := warrior.UseBattleShout

	// Three pieces of Battlegear of Wrath add a flat 30 to the shout. HasBsT2 is
	// set by the user or by the set worn at the start. The set is worth the 30 on
	// the shout this warrior makes, so it raises what that copy applies and what
	// the cast is worth to the category together.
	battleShoutBase := buffs.BattleShoutValue(0)
	battleShoutValue := battleShoutBase
	shoutsWithTheSet := castsOwnShout && warrior.HasBsT2
	if shoutsWithTheSet {
		battleShoutValue += buffs.BattleShoutT2Bonus
	}

	auras := warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
		// The raid's Battle Shout registers the external copy before this runs, and that copy
		// keeps the build phase it was registered with.
		raidShout := !castsOwnShout && unit.GetAuraByID(core.ActionID{SpellID: battleShoutRank.ID}.WithTag(-1)) != nil

		// Booming Voice widens the radius and lowers the cost only, so the aura takes no talent points.
		aura := buffs.NewSpellDataBattleShoutAura(unit, castsOwnShout, 0)
		if shoutsWithTheSet {
			buffs.AddGeneratedFlatBonus(aura, stats.AttackPower, battleShoutBase, buffs.BattleShoutT2Bonus)
		}
		if !raidShout {
			aura.BuildPhase = core.Ternary(castsOwnShout, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone)
		}
		return aura
	})

	// What holds the Battle Shout category decides whether shouting is worth a
	// global. Nothing there and the cast puts the buff up. This warrior's own
	// copy there and the cast only refreshes it, which is worth a global once it
	// is about to run out. Anything else has to be outbid first - a warrior
	// without the set would otherwise keep recasting a 139 shout the category
	// turns away while an external 169 one is up.
	battleShoutCategory := warrior.GetExclusiveEffectCategory(buffs.BattleShoutCategory)

	config := spelldata.SpellConfig(&warrior.Unit, battleShoutRank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	// TODO: Manual review needed -- spell 25289 carries no threat effect; none is modelled until
	// measured in game.
	config.FlatThreatBonus = 0

	config.ExtraCastCondition = func(sim *core.Simulation, _ *core.Unit) bool {
		active := battleShoutCategory.GetActiveEffect()
		if active == nil {
			return true
		}
		if active.Aura == auras.Get(&warrior.Unit) {
			return active.Aura.RemainingDuration(sim) <= ShoutExpirationThreshold
		}
		return battleShoutValue >= active.Priority
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
		auras.ActivateAllPlayers(sim)
	}

	config.RelatedAuraArrays = auras.ToMap()

	warrior.BattleShout = warrior.RegisterSpell(config)
}
