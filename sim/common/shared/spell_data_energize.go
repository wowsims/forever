package shared

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// An on-use item whose spell restores the wearer's mana, rage or energy, at once or over time, on the
// item's cooldowns with the cast time and global cooldown the row states. A character without that
// bar does not register it. The cooldown manager uses it once the whole gain fits under the bar's
// maximum.
func NewSpellDataEnergizeOnUse(itemID int32) {
	registerSpellDataOnUseCooldown(itemID, func(character *core.Character, row *spelldata.Spell) (core.SpellConfig, core.MajorCooldown, bool) {
		effect := row.ProcEnergizeEffect()
		power := dbcenums.PowerType(effect.Misc)
		bar, ok := energizedBarOf(character, power, core.ActionID{ItemID: itemID})
		if !ok {
			return core.SpellConfig{}, core.MajorCooldown{}, false
		}

		amount := func(sim *core.Simulation) float64 {
			return barAmount(power, effect.Roll(sim))
		}

		config := spelldata.SpellConfig(&character.Unit, row)
		ticks := 1.0
		if effect.Aura == dbcenums.A_PERIODIC_ENERGIZE {
			config.Hot = spelldata.DotConfig(row, effect)
			config.Hot.SelfOnly = true
			config.Hot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
				bar.add(sim, amount(sim), bar.metrics)
			}
			config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				spell.SelfHot().Apply(sim)
			}
			ticks = float64(config.Hot.NumberOfTicks)
		} else {
			config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				bar.add(sim, amount(sim), bar.metrics)
			}
		}

		whole := barAmount(power, effect.Max()*ticks)

		return config, core.MajorCooldown{
			Type: bar.cdType,
			ShouldActivate: func(_ *core.Simulation, _ *core.Character) bool {
				return bar.room() >= whole
			},
		}, true
	})
}

// The bar a gain fills: the room left in it, the character's own gain on it metered under the item,
// and the cooldown type core files a gain of that bar under.
type energizedBar struct {
	room    func() float64
	add     func(*core.Simulation, float64, *core.ResourceMetrics)
	metrics *core.ResourceMetrics
	cdType  core.CooldownType
}

func energizedBarOf(character *core.Character, power dbcenums.PowerType, actionID core.ActionID) (energizedBar, bool) {
	switch {
	case power == dbcenums.POWER_MANA && character.HasManaBar():
		return energizedBar{
			room:    func() float64 { return character.MaxMana() - character.CurrentMana() },
			add:     character.AddMana,
			metrics: character.NewManaMetrics(actionID),
			cdType:  core.CooldownTypeMana,
		}, true
	case power == dbcenums.POWER_RAGE && character.HasRageBar():
		return energizedBar{
			room:    func() float64 { return character.MaximumRage() - character.CurrentRage() },
			add:     character.AddRage,
			metrics: character.NewRageMetrics(actionID),
			cdType:  core.CooldownTypeDPS,
		}, true
	case power == dbcenums.POWER_ENERGY && character.HasEnergyBar():
		return energizedBar{
			room:    func() float64 { return character.MaximumEnergy() - character.CurrentEnergy() },
			add:     character.AddEnergy,
			metrics: character.NewEnergyMetrics(actionID),
			cdType:  core.CooldownTypeDPS,
		}, true
	}
	return energizedBar{}, false
}

// Whether the power type is one of the bars a gain fills: mana, rage or energy.
func EnergizesABar(power dbcenums.PowerType) bool {
	switch power {
	case dbcenums.POWER_MANA, dbcenums.POWER_RAGE, dbcenums.POWER_ENERGY:
		return true
	}
	return false
}

// Why the spell a proc row fires cannot be built, or nil where it can: the row states no effect
// beside its proc, and the spell it fires puts mana, rage or energy into the wearer's bar at once or
// changes the cooldown of one of the wearer's spell categories.
func FiredSpellUnsupported(row *spelldata.Spell) []string {
	var unsupported []string
	if len(row.Effects) > 1 {
		unsupported = append(unsupported, fmt.Sprintf("%d states effects beside its proc, which the listener leaves out", row.ID))
	}
	fired := row.FirstAura(dbcenums.A_PROC_TRIGGER_SPELL).Trigger()
	gain, cooldown := fired.EnergizeEffect(), fired.CategoryCooldownEffect()

	switch {
	case fired == spelldata.Nil:
		unsupported = append(unsupported, fmt.Sprintf("%d fires no spell the store carries", row.ID))
	case gain != spelldata.NilEffect:
		if gain.Target[0] != dbcenums.TARGET_UNIT_CASTER {
			unsupported = append(unsupported, fmt.Sprintf("%d's gain lands on implicit target %d, not the wearer", fired.ID, gain.Target[0]))
		}
		if !EnergizesABar(dbcenums.PowerType(gain.Misc)) {
			unsupported = append(unsupported, fmt.Sprintf("%d fills power type %d, not mana, rage or energy", fired.ID, gain.Misc))
		}
	case cooldown != spelldata.NilEffect:
		if cooldown.Target[0] != dbcenums.TARGET_UNIT_CASTER {
			unsupported = append(unsupported, fmt.Sprintf("%d changes the cooldowns of implicit target %d, not the wearer", fired.ID, cooldown.Target[0]))
		}
	default:
		unsupported = append(unsupported, fmt.Sprintf("%d fires %d, which neither restores a resource nor changes a spell category's cooldown", row.ID, fired.ID))
	}
	return unsupported
}

// What the spell a proc row fires does, or nil where the wearer has no bar its gain fills. A
// cooldown change moves the whole category's timer, which is where a spell whose only recovery is
// its category's keeps its cooldown; the game files state a reduction as a negative time, Spark's
// 1306610 -100000.
func FiredSpellHandler(character *core.Character, actionID core.ActionID, row *spelldata.Spell) core.ProcHandler {
	fired := row.FirstAura(dbcenums.A_PROC_TRIGGER_SPELL).Trigger()
	if gain := fired.EnergizeEffect(); gain != spelldata.NilEffect {
		power := dbcenums.PowerType(gain.Misc)
		bar, ok := energizedBarOf(character, power, actionID)
		if !ok {
			return nil
		}
		return func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			bar.add(sim, barAmount(power, gain.Roll(sim)), bar.metrics)
		}
	}

	cooldown := fired.CategoryCooldownEffect()
	timer := character.CategoryTimer(cooldown.Misc)
	return func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		timer.Set(max(sim.CurrentTime, timer.ReadyAt()+cooldown.TimeValue()))
	}
}

func barAmount(power dbcenums.PowerType, v float64) float64 {
	if power.InTenths() {
		return v / 10
	}
	return v
}
