package forever

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Shahram 10718, whom 16602 summons, casts one of his six spells and nothing else, picked by these
// weights, which round his picks in the logs (EV-0123). Fist of Shahram 16601 hastes Shahram himself, so its pick does nothing for the
// wearer.
var shahramPicks = []struct {
	spellID int32
	weight  float64
}{
	{16596, 3}, // Flames of Shahram
	{16598, 3}, // Will of Shahram
	{16597, 1}, // Curse of Shahram
	{16599, 1}, // Blessing of Shahram
	{16600, 1}, // Might of Shahram
	{16601, 1}, // Fist of Shahram
}

// Shahram begins his hardcast a uniform 1.03 to 1.83 s after the summon (EV-0123).
const shahramCastStartMin = 1030 * time.Millisecond
const shahramCastStartMax = 1830 * time.Millisecond

// How many Will of Shahram copies the wearer can carry at once. Each summon's Will is a copy of its
// own on its own timer (EV-0123); with every copy up, a new one renews the copy closest to fading.
const shahramWillCopies = 5

func init() {
	// Blackblade of Shahram
	// https://www.wowhead.com/forever/item=12592/blackblade-of-shahram
	//
	// Chance on hit: Summons the infernal spirit of Shahram.
	//
	// The game files state no rate for 16602; the item's chance-on-hit entry carries the one measured
	// on the sword's speed. The sim summons no Shahram: each proc picks one of his spells, which lands
	// after his cast start and the picked row's hardcast. His spells take nothing from the wearer's
	// damage modifiers, never crit and put no threat on the wearer, and they roll the wearer's spell
	// hit table.
	newEntryRateItemEffect(12592, 16602, func(character *core.Character, ppm float64) {
		spells := map[int32]*core.Spell{
			16596: registerFlamesOfShahram(character),
			16598: registerWillOfShahram(character),
			16597: registerCurseOfShahram(character),
			16599: registerBlessingOfShahram(character),
			16600: registerMightOfShahram(character),
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
			pick := shahramPicks[len(shahramPicks)-1]
			for _, p := range shahramPicks {
				if roll -= p.weight; roll < 0 {
					pick = p
					break
				}
			}
			spell := spells[pick.spellID]
			if spell == nil {
				return
			}
			pa := sim.GetConsumedPendingActionFromPool()
			pa.NextActionAt = sim.CurrentTime + castStart + spelldata.MustFind(pick.spellID).CastTime()
			pa.OnAction = func(sim *core.Simulation) {
				spell.Cast(sim, target)
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
	config := spelldata.SpellConfig(&character.Unit, row, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc())
	if row.IsAProc() {
		config.Flags |= core.SpellFlagProc
	}
	config.ThreatMultiplier = 0
	return config
}

// Flames of Shahram 16596 burns every enemy around Shahram.
func registerFlamesOfShahram(character *core.Character) *core.Spell {
	row := spelldata.MustFind(16596)
	damage := row.DamageEffect()
	config := shahramSpellConfig(character, row)
	config.Flags |= core.SpellFlagIgnoreAttackerModifiers
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamage(sim, damage.DirectAmount(sim)*sim.Encounter.AOECapMultiplier(), spell.OutcomeMagicHit)
	}
	return character.RegisterSpell(config)
}

// Will of Shahram 16598 lands on Shahram's master, the wearer, as a copy of its own.
func registerWillOfShahram(character *core.Character) *core.Spell {
	row := spelldata.MustFind(16598)
	copies := make([]*core.Aura, shahramWillCopies)
	for i := range copies {
		copies[i] = character.RegisterAura(spelldata.AuraConfig(row, spelldata.Label(fmt.Sprintf("%s %d", row.Name, i+1))))
		spelldata.ParseEffects(character, copies[i], row)
	}
	config := shahramSpellConfig(character, row)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		renewed := copies[0]
		for _, aura := range copies {
			if !aura.IsActive() {
				renewed = aura
				break
			}
			if aura.RemainingDuration(sim) < renewed.RemainingDuration(sim) {
				renewed = aura
			}
		}
		renewed.Activate(sim)
	}
	return character.RegisterSpell(config)
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

// Blessing of Shahram 16599 restores mana to each member of the wearer's party, pets included, every
// period for the row's duration; two Blessings tick side by side (EV-0123). Its heal is left out.
func registerBlessingOfShahram(character *core.Character) *core.Spell {
	row := spelldata.MustFind(16599)
	mana := row.FirstAura(dbcenums.A_PERIODIC_ENERGIZE)
	type member struct {
		unit    *core.Unit
		metrics *core.ResourceMetrics
	}
	var members []member
	character.Env.RegisterPreFinalizeEffect(func() {
		for _, agent := range character.Party.PlayersAndPets {
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
