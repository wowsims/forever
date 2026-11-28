package forever

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Spell weights in EV-0125.
var shahramPicks = []struct {
	spellID  int32
	weight   float64
	register func(*core.Character) *core.Spell
}{
	{16596, 7, registerFlamesOfShahram},
	{16598, 5, registerWillOfShahram},
	{16597, 2, registerCurseOfShahram},
	{16599, 2, registerBlessingOfShahram},
	{16600, 2, registerMightOfShahram},
	{16601, 2, registerFistOfShahram},
}

// Hardcast start in EV-0125.
const shahramCastStartMin = 1030 * time.Millisecond
const shahramCastStartMax = 1830 * time.Millisecond

// The most Will of Shahram stacks the wearer can carry. No cap in game files, so this one is
// the sim's own.
const shahramWillMaxStacks = 5

func init() {
	// Blackblade of Shahram
	// https://www.wowhead.com/forever/item=12592/blackblade-of-shahram
	//
	// Chance on hit: Summons the infernal spirit of Shahram.
	//
	// PPM measurement in EV-0125.
	// The sim summons no Shahram: each proc picks one of his spells.
	newEntryRateItemEffect(12592, 16602, func(character *core.Character, ppm float64) {
		type pickedSpell struct {
			spell    *core.Spell
			hardcast time.Duration
		}
		spells := make([]pickedSpell, len(shahramPicks))
		for i, pick := range shahramPicks {
			spells[i] = pickedSpell{pick.register(character), spelldata.MustFind(pick.spellID).CastTime()}
		}
		var total float64
		for _, pick := range shahramPicks {
			total += pick.weight
		}

		row := spelldata.MustFind(16602)
		config := spelldata.SpellConfig(&character.Unit, row, spelldata.Flags(core.SpellFlagNoOnCastComplete))
		config.ProcMask = core.ProcMaskEmpty
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			castStart := time.Duration(sim.RollWithLabel(float64(shahramCastStartMin), float64(shahramCastStartMax), "Shahram Cast Start"))
			roll := sim.RollWithLabel(0, total, "Shahram Pick")
			picked := len(shahramPicks) - 1
			for i, p := range shahramPicks {
				if roll -= p.weight; roll < 0 {
					picked = i
					break
				}
			}
			pick := spells[picked]
			pa := sim.GetConsumedPendingActionFromPool()
			pa.NextActionAt = sim.CurrentTime + castStart + pick.hardcast
			pa.OnAction = func(sim *core.Simulation) {
				pick.spell.Cast(sim, target)
			}
			sim.AddPendingAction(pa)
		}
		summon := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			summon.Cast(sim, result.Target)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(12592, ppm, 0)
		})
		trigger.Name = "Blackblade of Shahram"
		trigger.ActionID = core.ActionID{ItemID: 12592}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(12592, character.MakeProcTriggerAura(trigger))
	})
}

// One of Shahram's spells as the wearer casts it, with Shahram's threat, which is not the wearer's.
func shahramSpellConfig(character *core.Character, row *spelldata.Spell) core.SpellConfig {
	config := procSpellConfig(character, row)
	config.ThreatMultiplier = 0
	return config
}

// Crit rolls in EV-0125.
func registerFlamesOfShahram(character *core.Character) *core.Spell {
	row := spelldata.MustFind(16596)
	damage := row.DamageEffect()
	config := shahramSpellConfig(character, row)
	config.Flags |= core.SpellFlagIgnoreAttackerModifiers
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamage(sim, damage.DirectAmount(sim)*sim.Encounter.AOECapMultiplier(), spell.OutcomeMagicHitAndCrit)
	}
	return character.RegisterSpell(config)
}

// Stack fades in EV-0125.
func registerWillOfShahram(character *core.Character) *core.Spell {
	row := spelldata.MustFind(16598)
	effect := row.EffectN(1)
	var perStack stats.Stats
	for _, stat := range effect.FlatStats() {
		perStack[stat] = effect.MeanAt(character.Level)
	}
	config := spelldata.AuraConfig(row)
	config.MaxStacks = shahramWillMaxStacks
	aura := core.MakeStackingAura(character, core.StackingStatAura{Aura: config, BonusPerStack: perStack})

	// When each stack fades, oldest first. Each fade is a pending action that drops every stack due by
	// then, so the fade of a stack already replaced or expired drops nothing.
	var fades []time.Duration
	aura.ApplyOnReset(func(_ *core.Aura, _ *core.Simulation) {
		fades = fades[:0]
	}).ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
		fades = fades[:0]
	})
	fade := func(sim *core.Simulation) {
		for len(fades) > 0 && fades[0] <= sim.CurrentTime {
			fades = slices.Delete(fades, 0, 1)
			aura.RemoveStack(sim)
		}
	}

	spell := shahramSpellConfig(character, row)
	spell.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		if len(fades) == shahramWillMaxStacks {
			fades = slices.Delete(fades, 0, 1)
		}
		fades = append(fades, sim.CurrentTime+row.Duration())
		aura.Activate(sim)
		aura.AddStack(sim)

		pa := sim.GetConsumedPendingActionFromPool()
		pa.NextActionAt = sim.CurrentTime + row.Duration()
		// A stack fades before the rotation acts at the same instant, the way an aura's expiry does.
		pa.Priority = core.ActionPriorityDOT
		pa.OnAction = fade
		sim.AddPendingAction(pa)
	}
	return character.RegisterSpell(spell)
}

// Slows the attacks of every enemy around Shahram it lands on; the movement slow does nothing in the sim.
func registerCurseOfShahram(character *core.Character) *core.Spell {
	row := spelldata.MustFind(16597)
	debuffs := shared.EnemyDebuffAuras(character, row, row.Duration(), row.DebuffEffects())
	config := shahramSpellConfig(character, row)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, target := range sim.Encounter.ActiveTargetUnits {
			if spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit).Landed() {
				debuffs.Get(target).Activate(sim)
			}
		}
	}
	return character.RegisterSpell(config)
}

// Stuns every enemy around Shahram it lands on.
func registerMightOfShahram(character *core.Character) *core.Spell {
	row := spelldata.MustFind(16600)
	label := fmt.Sprintf("%s %d", row.Name, row.ID)
	stuns := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		if aura := target.GetAura(label); aura != nil {
			return aura
		}
		return target.RegisterStunAura(label, core.ActionID{SpellID: row.ID}, row.Duration())
	})
	config := shahramSpellConfig(character, row)
	if row.IsStun() {
		config.Flags |= core.SpellFlagStun
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, target := range sim.Encounter.ActiveTargetUnits {
			if spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit).Landed() {
				core.ApplyStun(sim, stuns.Get(target))
			}
		}
	}
	return character.RegisterSpell(config)
}

// Mana targets in EV-0125. Its heal is left out.
func registerBlessingOfShahram(character *core.Character) *core.Spell {
	row := spelldata.MustFind(16599)
	mana := row.FirstAura(dbcenums.A_PERIODIC_ENERGIZE)
	type member struct {
		unit    *core.Unit
		metrics *core.ResourceMetrics
	}
	var members []member
	character.Env.RegisterPreFinalizeEffect(func() {
		for _, agent := range character.Party.Players {
			if unit := &agent.GetCharacter().Unit; unit.HasManaBar() {
				members = append(members, member{unit, unit.NewManaMetrics(core.ActionID{SpellID: row.ID})})
			}
		}
	})
	config := shahramSpellConfig(character, row)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   mana.Period(),
			NumTicks: int(row.Duration() / mana.Period()),
			OnAction: func(sim *core.Simulation) {
				for _, m := range members {
					if m.unit.IsEnabled() {
						m.unit.AddMana(sim, mana.Average(), m.metrics)
					}
				}
			},
		})
	}
	return character.RegisterSpell(config)
}

// Two Fists run side by side (EV-0125), and their speed counts once, an owner ruling. Every copy lasts
// the spell's duration, so the newest fades last and one aura refreshed by each Fist holds the same
// speed for the same time.
func registerFistOfShahram(character *core.Character) *core.Spell {
	row := spelldata.MustFind(16601)
	aura := character.RegisterAura(spelldata.AuraConfig(row))
	spelldata.ParseEffects(character, aura, row)
	config := shahramSpellConfig(character, row)
	config.RelatedSelfBuff = aura
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	return character.RegisterSpell(config)
}
