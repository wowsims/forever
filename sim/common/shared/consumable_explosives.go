package shared

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Why the damage builders cannot run the spell, or nil where they can. A cloud has to state how long
// it lasts, and its damage has to tick, land on an enemy and state Always Hit, since the cloud rolls
// no hit for it.
func DamageConsumableUnsupported(s *spelldata.Spell) []string {
	cloud, damage := DamageCloudOf(s)
	if cloud == spelldata.Nil {
		return nil
	}
	var unsupported []string
	periodic := damage.PeriodicDamageEffect()
	if cloud.DurationMs <= 0 {
		unsupported = append(unsupported, fmt.Sprintf("the cloud %d states no duration", cloud.ID))
	}
	if periodic.PeriodMs <= 0 {
		unsupported = append(unsupported, fmt.Sprintf("%d states no period for its damage over time", damage.ID))
	}
	if !periodic.HitsAnEnemy() {
		unsupported = append(unsupported, fmt.Sprintf("%d's damage over time lands on implicit target %d, not an enemy", damage.ID, periodic.Target[0]))
	}
	if !damage.AlwaysHits() {
		unsupported = append(unsupported, fmt.Sprintf("%d rolls a hit, which the cloud does not roll", damage.ID))
	}
	return unsupported
}

// The cloud a consumable's missile leaves where it lands, and the damage over time the cloud puts on
// the enemies inside it: the spell's E_TRIGGER_MISSILE fires a spell whose E_CREATE_AREATRIGGER is
// the cloud, and the damage is the first spell the cloud triggers that deals periodic damage
// (Caustic Smog Potion's 1252525 fires 1252526, which puts 1252543 on the enemies). Nil and Nil
// where the spell leaves no such cloud: Disorienting Smog Potion's cloud deals no damage.
func DamageCloudOf(row *spelldata.Spell) (cloud, damage *spelldata.Spell) {
	for i := range row.Effects {
		if row.Effects[i].Type != dbcenums.E_TRIGGER_MISSILE {
			continue
		}
		trigger := row.Effects[i].Trigger()
		if !slices.ContainsFunc(trigger.Effects, func(e spelldata.Effect) bool { return e.Type == dbcenums.E_CREATE_AREATRIGGER }) {
			continue
		}
		for _, s := range trigger.Triggered() {
			if s.PeriodicDamageEffect() != spelldata.NilEffect {
				return trigger, s
			}
		}
	}
	return spelldata.Nil, spelldata.Nil
}

// Whether a consumable's spell deals damage to enemies: a hit or a damage over time of its own, or a
// cloud's.
func ConsumableDamages(row *spelldata.Spell) bool {
	if cloud, _ := DamageCloudOf(row); cloud != spelldata.Nil {
		return true
	}
	return row.DamagesAnEnemy()
}

// The spell a consumable that damages enemies registers: its cloud's, or the damage its own row deals.
func consumableDamageSpell(character *core.Character, itemID int32, row *spelldata.Spell) core.SpellConfig {
	if cloud, damage := DamageCloudOf(row); cloud != spelldata.Nil {
		return cloudDamageSpell(character, itemID, row, cloud, damage)
	}

	return spellDataOnUseDamageSpell(character, row)
}

// A cloud's damage. The row's missile flies to the target, at its speed and for no less than its
// minimum flight time, and the cloud it leaves there puts its damage over time on every enemy for
// as long as the cloud lasts: Caustic Smog Potion's 1252543 ticks every second for 1252526's 12 sec.
// The damage states Always Hit (DamageConsumableUnsupported), so it goes on each enemy unrolled. The
// ticks count under the item, tagged 2 apart from the damage to the user's tag 1
// (registerCasterDamageSpell).
func cloudDamageSpell(character *core.Character, itemID int32, row, cloud, damage *spelldata.Spell) core.SpellConfig {
	lasting := *damage
	lasting.DurationMs = cloud.DurationMs
	tickConfig := spelldata.SpellConfig(&character.Unit, &lasting, damageShape(&lasting), castBy(true))
	tickConfig.ActionID = core.ActionID{ItemID: itemID}.WithTag(2)
	tickConfig.DefenseType = damageDefenseType(tickConfig.DefenseType, tickConfig.SpellSchool, false)
	tickConfig.Dot = spelldata.DotConfig(&lasting, lasting.PeriodicDamageEffect())
	tickSpell := character.RegisterSpell(tickConfig)

	config := spelldata.SpellConfig(&character.Unit, row, castBy(false))
	config.ProcMask = core.ProcMaskEmpty
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.WaitTravelTime(sim, tickSpell.ApplyAllDots)
	}
	return config
}

// An explosive used from the bar: the damage its on-use spell deals, on the item's cooldown and the
// category cooldown every explosive shares, and the damage the spell deals to the thrower where the
// row states one. Only a character with the profession and skill the item requires can use it.
func NewSpellDataExplosive(itemID int32) {
	core.NewSpellDataConsumableEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		consumable := core.GetConsumableByID(itemID)
		if !character.MeetsProfession(consumable.RequiredProfession, consumable.ReqProfRank) {
			return
		}

		for _, itemEffect := range consumable.OnUseEffects() {
			row := spelldata.MustFind(itemEffect.BuffId)
			config := consumableDamageSpell(character, itemID, row)
			config.ActionID = core.ActionID{ItemID: itemID}
			config.Flags |= core.SpellFlagExplosive
			config.Cast = consumableCast(character, itemEffect, row)
			if self := row.CasterDamageEffect(); self != spelldata.NilEffect {
				selfSpell := registerCasterDamageSpell(character, core.SpellConfig{
					ActionID:         config.ActionID,
					SpellSchool:      config.SpellSchool,
					DefenseType:      config.DefenseType,
					ProcMask:         core.ProcMaskSpellDamage,
					Flags:            core.SpellFlagExplosive,
					DamageMultiplier: 1,
					ThreatMultiplier: 1,
				}, self)
				applyEffects := config.ApplyEffects
				config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					applyEffects(sim, target, spell)
					selfSpell.ApplyEffects(sim, &character.Unit, selfSpell)
				}
			}

			// An explosive with a cooldown of its own, such as the sapper, goes before the ones only the
			// shared category holds back, which would otherwise delay it by that category's cooldown.
			priority := int32(core.CooldownPriorityLow)
			if config.Cast.CD.Duration > 0 {
				priority += 10
			}
			character.AddMajorCooldown(core.MajorCooldown{
				Spell:    character.RegisterSpell(config),
				Type:     core.CooldownTypeDPS | core.CooldownTypeExplosive,
				Priority: priority,
			})
		}
	})
}

// The damage a consumable's spell deals its user, rolled from e, as a spell of its own under config,
// tagged 1, so that it reaches the character as a harmful spell hit, which is what a listener on spell
// damage taken hears.
func registerCasterDamageSpell(character *core.Character, config core.SpellConfig, e *spelldata.Effect) *core.Spell {
	config.ActionID = config.ActionID.WithTag(1)
	config.DirectVariance = e.DirectVariance()
	var outcome core.OutcomeApplier
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, e.DirectAmount(sim), outcome)
	}
	spell := character.RegisterSpell(config)
	outcome = casterDamageOutcome(spell, spelldata.Find(e.SpellID).CannotCrit())
	return spell
}

// Damage a spell deals to its own caster always lands; it can still crit, on the crit of its defense
// type, unless the row says it cannot.
func casterDamageOutcome(spell *core.Spell, cannotCrit bool) core.OutcomeApplier {
	switch {
	case cannotCrit:
		return spell.OutcomeAlwaysHit
	case spell.DefenseType == core.DefenseTypeMelee:
		return spell.OutcomeMeleeSpecialCritOnly
	case spell.DefenseType == core.DefenseTypeRanged:
		return spell.OutcomeRangedCritOnly
	default:
		return spell.OutcomeMagicCrit
	}
}
