package shared

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// An item or enchant proc whose spell puts a debuff on the enemy it lands on and does nothing else
// the sim models: Frostguard's Chilled 16927 slows its attacks, Annihilator's Armor Shatter 16928
// takes its armor. BuffSpellID names the spell where it is not the trigger.
func NewSpellDataDebuffProc(cfg SpellDataProc, variants []ItemVariant) {
	forEachSpellDataVariant(cfg, variants, registerSpellDataDebuffProc)
}

func registerSpellDataDebuffProc(cfg SpellDataProc) {
	registerSpellDataRowProc(cfg, nil, func(agent core.Agent, source effectSource, trigger *spelldata.Spell, debuff *spelldata.Spell) {
		applySpellDataDebuffProc(agent, cfg, source, trigger, debuff)
	})
}

func applySpellDataDebuffProc(agent core.Agent, cfg SpellDataProc, source effectSource, trigger *spelldata.Spell, debuff *spelldata.Spell) {
	character := agent.GetCharacter()
	debuffs := EnemyDebuffAuras(character, debuff, debuff.Duration(), debuff.DebuffEffects())

	config := spellDataProcListener(character, cfg, source, trigger)
	callback := config.Callback
	config.Handler = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if target := procDamageTarget(character, callback, spell, result); target != nil {
			applyDebuff(sim, debuffs.Get(target))
		}
	}
	config.TriggerImmediately = true

	source.registerTrigger(character, config)
}

// A re-application refreshes the duration and, on a row that stacks, adds a stack up to its count.
func applyDebuff(sim *core.Simulation, aura *core.Aura) {
	aura.Activate(sim)
	if aura.MaxStacks > 0 {
		aura.AddStack(sim)
	}
}

// Proc debuffs that bid in a raid debuff's category though no raid debuff reads their row. Vibroblade's
// Puncture Armor 1293005 is here for when its wielder proc is registered, which no item does yet. In
// game it never lands while Curse of Recklessness or Faerie Fire is on the target, and either one
// landing removes it (EV-0147); it is weaker than either, so bidding in their slot gives the target
// the same armor. Bashguuder and Rivenspike's 17315 is the raid's Puncture Armor row, whose player
// copy bids there already.
var debuffProcCategories = map[int32]string{
	1293005: buffs.FaerieFireCategory,
}

// The row's debuff on each enemy, carrying the effects at the given positions, for the duration given.
// The aura is the row's rather than the wearer's, so every character applying the row refreshes one
// debuff on the target, and the first to register it parses it. Each slow takes its exclusive
// category, attacks Thunder Clap's and casts Slow's, where only the strongest applies; every other
// effect applies to the target as the row states it, per stack. A row a raid debuff reads is that
// debuff's player copy instead, which bids in the raid debuff's category beside the raid's copy, or
// is the raid's copy where the raid's driver makes it the player's own, as Armor Shatter and Puncture
// Armor do when the player wields their weapon.
func EnemyDebuffAuras(character *core.Character, row *spelldata.Spell, duration time.Duration, effects []int32) core.AuraArray {
	config := spelldata.AuraConfig(row, spelldata.Label(fmt.Sprintf("%s %d", row.Name, row.ID)))
	config.Duration = duration
	slows := row.SlowEffects()

	return character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		if aura := buffs.NewSpellDataDebuffAuraBySpell(target, row.ID); aura != nil {
			return aura
		}
		if aura := target.GetAura(config.Label); aura != nil {
			return aura
		}

		aura := target.RegisterAura(config)
		var parsed []int32
		for _, i := range effects {
			if !slices.Contains(slows, i) {
				parsed = append(parsed, i)
				continue
			}
			effect := row.EffectN(int(i))
			slowedTime := core.MultiplierFromPercent(-effect.Average())
			if effect.ChangesCastSpeed() {
				buffs.CastSpeedReductionEffect(aura, slowedTime)
			} else {
				buffs.AtkSpeedReductionEffect(aura, slowedTime)
			}
		}
		if len(parsed) > 0 {
			opts := []spelldata.ParseOpt{spelldata.Effects(parsed...)}
			if category, ok := debuffProcCategories[row.ID]; ok {
				opts = append(opts, spelldata.Exclusive(category, false))
			}
			spelldata.ParseEffects(nil, aura, row, opts...)
		}
		return aura
	})
}

// The debuff a proc's damage spell puts on each enemy its hit lands on, or nil where the row states
// none. A spell that deals only damage over time has no hit, and its target takes the debuff with it.
func debuffOnLanding(character *core.Character, row *spelldata.Spell) afterDealt {
	effects := row.DebuffEffects()
	if len(effects) == 0 {
		return nil
	}

	debuffs := EnemyDebuffAuras(character, row, row.Duration(), effects)
	return func(sim *core.Simulation, _ *core.Spell, target *core.Unit, results core.SpellResultSlice) {
		if len(results) == 0 {
			applyDebuff(sim, debuffs.Get(target))
			return
		}
		for _, result := range results {
			if result.Landed() {
				applyDebuff(sim, debuffs.Get(result.Target))
			}
		}
	}
}
