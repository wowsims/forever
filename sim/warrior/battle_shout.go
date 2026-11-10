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

// Rank 6 is the top rank the trainer teaches; Rank 7 comes from Manual of Battle Shout VII.
var battleShoutTrainerRank = spellData.BattleShout.ByID(11551)
var battleShoutBookRank = spellData.BattleShout.ByID(25289)

func (warrior *Warrior) registerBattleShout() {
	warrior.registerBattleShoutRank(battleShoutTrainerRank)
	warrior.registerBattleShoutRank(battleShoutBookRank)
}

func (warrior *Warrior) registerBattleShoutRank(rank *spelldata.Spell) {
	// A warrior that shouts builds its own copy. One that never shouts gets the
	// isPlayer=false constructor, whose aura is the raid's external copy, and the
	// raid's Battle Shout state sets what that copy is worth.
	castsOwnShout := warrior.UseBattleShout

	// Three pieces of Battlegear of Wrath add a flat 30 to the shout. HasBsT2 is
	// set by the user or by the set worn at the start. The set is worth the 30 on
	// the shout this warrior makes, so it raises what that copy applies and what
	// the cast is worth to the category together.
	battleShoutBase := buffs.BattleShoutRankValue(rank)
	battleShoutValue := battleShoutBase
	shoutsWithTheSet := castsOwnShout && warrior.HasBsT2
	if shoutsWithTheSet {
		battleShoutValue += buffs.BattleShoutT2Bonus
	}

	// The class option decides which rank's own copy the character sheet shows, as the rank the
	// rotation picks is unknown when the sheet is built. The rotation alone decides the rank cast.
	sheetRank := core.Ternary(warrior.HasBattleShoutRank7, battleShoutBookRank, battleShoutTrainerRank)
	ownBuildPhase := core.Ternary(castsOwnShout && rank == sheetRank, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone)

	auras := warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
		// The raid's Battle Shout registers the external copy of its rank before this runs, and
		// that copy keeps the build phase it was registered with.
		raidShout := !castsOwnShout && unit.GetAuraByID(core.ActionID{SpellID: rank.ID}.WithTag(-1)) != nil

		// Booming Voice widens the radius and lowers the cost only, so the aura takes no talent points.
		aura := buffs.NewSpellDataBattleShoutRankAura(unit, castsOwnShout, rank)
		if shoutsWithTheSet {
			buffs.AddGeneratedFlatBonus(aura, stats.AttackPower, battleShoutBase, buffs.BattleShoutT2Bonus)
		}
		if !raidShout {
			aura.BuildPhase = ownBuildPhase
		}
		return aura
	})

	// What holds the Battle Shout category decides whether shouting is worth a
	// global. Nothing there and the cast puts the buff up. This rank's own copy
	// there and the cast only refreshes it, which is worth a global once it is
	// about to run out. Anything else has to be outbid first, as a warrior without
	// the set would otherwise keep recasting a shout the category turns away
	// while a stronger external one is up.
	battleShoutCategory := warrior.GetExclusiveEffectCategory(buffs.BattleShoutCategory)

	config := spelldata.SpellConfig(&warrior.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	// TODO: Manual review needed -- spells 11551 and 25289 carry no threat effect; none is modelled
	// until measured in game.
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

	warrior.RegisterSpell(config)
}
