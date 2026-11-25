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

// Shahram 10718, whom 16602 summons, casts one of his six spells and nothing else, picked by these
// weights (EV-0125).
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

// Shahram begins his hardcast 1.03 to 1.83 s after the summon (EV-0125); the sim draws the delay
// uniformly over that range.
const shahramCastStartMin = 1030 * time.Millisecond
const shahramCastStartMax = 1830 * time.Millisecond

// The most Will of Shahram stacks the wearer can carry. The game files state no cap, so this one is
// the sim's own. With every stack up, a new Will replaces the stack closest to fading.
const shahramWillMaxStacks = 5

func init() {
	// Blackblade of Shahram
	// https://www.wowhead.com/forever/item=12592/blackblade-of-shahram
	//
	// Chance on hit: Summons the infernal spirit of Shahram.
	//
	// The game files state no rate for 16602; the item's chance-on-hit entry carries the one measured
	// on the sword's speed. The sim summons no Shahram: each proc picks one of his spells, which lands
	// after his cast start and the picked row's hardcast. His spells take nothing from the wearer's
	// damage modifiers and put no threat on the wearer, and they roll the wearer's spell hit table;
	// Flames rolls the wearer's spell crit as well.
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

// One of Shahram's spells as the wearer casts it: out of the rotation, with no cast of its own, and
// with Shahram's threat, which is not the wearer's.
func shahramSpellConfig(character *core.Character, row *spelldata.Spell) core.SpellConfig {
	config := procSpellConfig(character, row)
	config.ThreatMultiplier = 0
	return config
}

// Flames of Shahram 16596 burns every enemy around Shahram, and can crit (EV-0125).
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

// Will of Shahram 16598 lands on Shahram's master, the wearer, as one more stack of the row's stats.
// Each stack fades the row's duration after its own landing (EV-0125), and the aura lasts until the
// latest one does.
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

	// When each stack fades, oldest first. A stack's fade is a pending action of its own, which drops
	// every stack due by then, so the action of a stack already replaced or expired drops nothing.
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

// Curse of Shahram 16597 slows the attacks of every enemy around Shahram it lands on; the movement
// slow does nothing in the sim.
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

// Might of Shahram 16600 stuns every enemy around Shahram it lands on, which a stun-immune enemy
// misses in full.
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

// Blessing of Shahram 16599 restores mana to each player of the wearer's party, never to a pet, every
// period for the row's duration; two Blessings tick side by side (EV-0125). Its heal is left out.
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

// Fist of Shahram 16601 lands on Shahram's master, the wearer, and raises the wearer's melee attack
// speed for the row's duration (EV-0125). A new Fist while one is up refreshes it.
func registerFistOfShahram(character *core.Character) *core.Spell {
	row := spelldata.MustFind(16601)
	// SpeedPseudoStats reads only the caster's own speed auras, and this one targets his master.
	effect := row.FirstAura(dbcenums.A_MOD_MELEE_HASTE_3)
	speed := make([]float64, stats.PseudoStatsLen)
	for _, pseudoStat := range spelldata.SpeedAuraPseudoStats[effect.Aura] {
		speed[pseudoStat] = effect.BasePoints
	}
	aura := character.RegisterAura(spelldata.AuraConfig(row)).AttachHastePseudoStats(speed)
	config := shahramSpellConfig(character, row)
	config.RelatedSelfBuff = aura
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	return character.RegisterSpell(config)
}
