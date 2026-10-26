package warrior

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

const stanceEffectCategory = "Stance"

var battleStanceRank = spellData.BattleStance.Highest()
var defensiveStanceRank = spellData.DefensiveStance.Highest()
var berserkerStanceRank = spellData.BerserkerStance.Highest()

// The passive each stance carries, which is where the game files state what standing in it is worth.
var battleStancePassive = spellData.BattleStancePassive.Highest()
var defensiveStancePassive = spellData.DefensiveStancePassive.Highest()
var berserkerStancePassive = spellData.BerserkerStancePassive.Highest()

// The three stance rows share category 47, so their 1 s category cooldown runs off one timer.
func (warrior *Warrior) makeStanceSpell(rank *spelldata.Spell, aura *core.Aura) *core.Spell {
	form := rank.ShapeshiftForm()
	rageMetrics := warrior.NewRageMetrics(aura.ActionID)
	maxRetainedRage := spellData.TacticalMastery.ValueAt(1) + spellData.ImprovedTacticalMastery.ValueAt(warrior.Talents.ImprovedTacticalMastery)

	config := spelldata.SpellConfig(&warrior.Unit, rank, spelldata.Flags(core.SpellFlagNoOnCastComplete|core.SpellFlagAPL))
	config.MenuGroup = "stances"
	aura.MenuGroup = "stances"

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.ShapeshiftForm != form
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		if warrior.WarriorInputs.StanceSnapshot {
			// Delayed, so same-GCD casts are affected by the current aura.
			// Alternatively, those casts could just (artificially) happen before the stance change.
			pa := sim.GetConsumedPendingActionFromPool()
			pa.NextActionAt = sim.CurrentTime + 10*time.Millisecond
			pa.OnAction = aura.Activate
			sim.AddPendingAction(pa)
		} else {
			aura.Activate(sim)
		}

		if warrior.CurrentRage() > maxRetainedRage {
			warrior.SpendRage(sim, warrior.CurrentRage()-maxRetainedRage, rageMetrics)
		}

		warrior.ShapeshiftForm = form
	}

	config.RelatedSelfBuff = aura

	return warrior.RegisterSpell(config)
}

func (warrior *Warrior) registerBattleStanceAura() *core.Aura {
	config := spelldata.AuraConfig(battleStanceRank)
	config.BuildPhase = core.Ternary(warrior.DefaultStance == proto.WarriorStance_WarriorStanceBattle, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone)

	aura := warrior.RegisterAura(config)
	spelldata.ParseEffects(&warrior.Character, aura, battleStancePassive)

	aura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})

	return aura
}

func (warrior *Warrior) registerDefensiveStanceAura() *core.Aura {
	config := spelldata.AuraConfig(defensiveStanceRank)
	config.BuildPhase = core.Ternary(warrior.DefaultStance == proto.WarriorStance_WarriorStanceDefensive, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone)

	aura := warrior.RegisterAura(config)
	spelldata.ParseEffects(&warrior.Character, aura, defensiveStancePassive)

	if warrior.Talents.Defiance > 0 {
		// The stance the talent's threat applies in is the aura it hangs on. The row states the
		// shield, but the parse never scopes to a shield, so it is the caller's condition, re-read
		// on an off-hand swap.
		defiance := spelldata.ParseEffects(&warrior.Character, aura,
			spellData.Defiance.Rank(warrior.Talents.Defiance),
			spelldata.Conditional(func() bool { return warrior.PseudoStats.CanBlock }))

		warrior.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}, func(sim *core.Simulation, _ proto.ItemSlot) {
			defiance.Refresh(sim)
		})
	}

	aura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})

	return aura
}

func (warrior *Warrior) registerBerserkerStanceAura() *core.Aura {
	config := spelldata.AuraConfig(berserkerStanceRank)
	config.BuildPhase = core.Ternary(warrior.DefaultStance == proto.WarriorStance_WarriorStanceBerserker, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone)

	aura := warrior.RegisterAura(config)
	// The row's fourth effect is an attack power percentage of 0, which has no sim kind and nothing
	// to apply either way.
	spelldata.ParseEffects(&warrior.Character, aura, berserkerStancePassive)

	aura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})

	return aura
}

func (warrior *Warrior) registerStances() {
	battleStanceAura := warrior.registerBattleStanceAura()
	defensiveStanceAura := warrior.registerDefensiveStanceAura()
	berserkerStanceAura := warrior.registerBerserkerStanceAura()
	warrior.BattleStance = warrior.makeStanceSpell(battleStanceRank, battleStanceAura)
	warrior.DefensiveStance = warrior.makeStanceSpell(defensiveStanceRank, defensiveStanceAura)
	warrior.BerserkerStance = warrior.makeStanceSpell(berserkerStanceRank, berserkerStanceAura)

	switch warrior.DefaultStance {
	case proto.WarriorStance_WarriorStanceBattle:
		core.MakePermanent(warrior.BattleStance.RelatedSelfBuff)
	case proto.WarriorStance_WarriorStanceDefensive:
		core.MakePermanent(warrior.DefensiveStance.RelatedSelfBuff)
	case proto.WarriorStance_WarriorStanceBerserker:
		core.MakePermanent(warrior.BerserkerStance.RelatedSelfBuff)
	}
}
