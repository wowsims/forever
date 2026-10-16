package shared

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Why the explosives builder cannot run the damage spell, or nil where it can: it puts a damage over
// time on the one target the spell is used on, so a row whose damage over time lands on every enemy
// in an area is not what it builds.
func DamageConsumableUnsupported(s *spelldata.Spell) []string {
	if periodic := s.PeriodicDamageEffect(); periodic != spelldata.NilEffect && periodic.HitsAnArea() {
		return []string{fmt.Sprintf("%d's damage over time lands on an area, which the damage builder applies to one target", s.ID)}
	}
	return nil
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
			config := spellDataOnUseDamageSpell(character, row)
			config.ActionID = core.ActionID{ItemID: itemID}
			config.Flags |= core.SpellFlagExplosive
			// The row of every Engineering explosive but Goblin Sapper Charge's 13241 marks it not usable
			// in a shapeshift form, which the row's cast requirement already carries. In game the sapper
			// cannot be used in one either: a druid leaves the form to throw it.
			if consumable.RequiredProfession == proto.Profession_Engineering {
				config.CastRequirement.NotShapeshifted = true
			}
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
