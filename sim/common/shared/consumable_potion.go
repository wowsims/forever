package shared

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Alchemists' Stone 13503's equip spell: A_DUMMY 20, "Increases the effects that healing and mana
// potions have on the wearer by $s1%."
const alchemistsStoneSpellID int32 = 17619

// A consumable whose on-use spell the potion builder runs on the user, or whose damage it deals to
// enemies (consumableDamageSpell), on the item's cooldown and the category it shares. A potion is a
// major cooldown only while the consumes pick it for an APL potion placeholder, the other potions are
// the rotation's; a conjured item (a healthstone, mana gem, rune or tea) is one only while it is the
// spec's chosen one; any other, such as the Flame Deflector's shield or the Scroll of Minor
// Evocation's mana, whenever the spec names it.
func NewSpellDataOnUseConsumable(itemID int32) {
	registerSpellDataConsumable(itemID)
}

// Why the consumable builders cannot run the spell, or nil where they can. A spell that damages
// enemies has to be one the damage builders run (DamageConsumableUnsupported), with nothing beside
// it on the user. Otherwise every effect has to land on the user and be a heal, an absorb, a gain of
// mana, rage or energy, damage to the user, or an aura the parser attaches, and the spell has to work
// outside a seated rest.
func ConsumableUnsupported(row *spelldata.Spell) []string {
	_, unsupported := consumablePartsOf(row)
	return unsupported
}

// Whether the parts of a consumable's spell the builder models restore power, and whether they heal
// or absorb.
func ConsumableRestores(row *spelldata.Spell) (power, survival bool) {
	parts, _ := consumablePartsOf(row)
	return len(parts.gains) > 0, parts.heal != nil || parts.absorb != nil
}

type consumableParts struct {
	damage     bool
	heal       *spelldata.Effect
	absorb     *spelldata.Effect
	gains      []*spelldata.Effect
	selfDamage *spelldata.Effect
	auraIdx    []int32
}

func (parts consumableParts) hasAura() bool {
	return len(parts.auraIdx) > 0
}

func (parts consumableParts) gainsOnly() bool {
	return !parts.damage && parts.heal == nil && parts.absorb == nil && !parts.hasAura()
}

func consumablePartsOf(row *spelldata.Spell) (consumableParts, []string) {
	var parts consumableParts
	var unsupported []string
	if row == spelldata.Nil {
		return parts, []string{"the spell has no row in the store"}
	}
	if ConsumableDamages(row) {
		parts.damage = true
		unsupported = DamageConsumableUnsupported(row)
		for i := range row.Effects {
			if LandsOnTheUser(&row.Effects[i]) {
				unsupported = append(unsupported, fmt.Sprintf("%s lands on the user beside the damage", row.Effects[i].Type))
			}
		}
		return parts, unsupported
	}

	for i := range row.Effects {
		e := &row.Effects[i]
		kind := e.Type.String()
		if e.Type == dbcenums.E_APPLY_AURA {
			kind = e.Aura.String()
		}

		switch {
		case e.Target[0] != dbcenums.TARGET_UNIT_CASTER:
			unsupported = append(unsupported, fmt.Sprintf("%s lands on implicit target %d, not the user", kind, e.Target[0]))
		case e.Type == dbcenums.E_HEAL, e.Type == dbcenums.E_HEAL_PCT,
			e.Type == dbcenums.E_APPLY_AURA && e.Aura == dbcenums.A_PERIODIC_HEAL:
			if parts.heal != nil {
				unsupported = append(unsupported, fmt.Sprintf("a second heal, %s", kind))
			}
			parts.heal = e
		case e.Type == dbcenums.E_APPLY_AURA && e.Aura == dbcenums.A_SCHOOL_ABSORB:
			if parts.absorb != nil {
				unsupported = append(unsupported, "a second absorb")
			}
			parts.absorb = e
		case e.Type == dbcenums.E_ENERGIZE, e.Type == dbcenums.E_ENERGIZE_PCT,
			e.Type == dbcenums.E_APPLY_AURA && e.Aura == dbcenums.A_PERIODIC_ENERGIZE:
			if EnergizesABar(dbcenums.PowerType(e.Misc)) {
				parts.gains = append(parts.gains, e)
			} else {
				unsupported = append(unsupported, fmt.Sprintf("%s fills power type %d", kind, e.Misc))
			}
		case e.Type == dbcenums.E_SCHOOL_DAMAGE:
			if parts.selfDamage != nil {
				unsupported = append(unsupported, "a second hit on the user")
			}
			parts.selfDamage = e
		case e.ParsedAura():
			parts.auraIdx = append(parts.auraIdx, int32(i+1))
		default:
			unsupported = append(unsupported, fmt.Sprintf("%s, which no consumable builder models", kind))
		}

		ticks := e.Type == dbcenums.E_APPLY_AURA && (e.Aura == dbcenums.A_PERIODIC_HEAL || e.Aura == dbcenums.A_PERIODIC_ENERGIZE)
		if ticks && (e.PeriodMs <= 0 || row.DurationMs <= 0) {
			unsupported = append(unsupported, fmt.Sprintf("%s states no period or no duration to tick over", kind))
		}
	}

	if row.WhileSeated() {
		unsupported = append(unsupported, fmt.Sprintf("%d restores only while the user sits, out of combat", row.ID))
	}
	if parts.heal != nil && parts.absorb != nil {
		unsupported = append(unsupported, "a heal beside an absorb")
	}
	if parts.hasAura() && row.DurationMs <= 0 {
		unsupported = append(unsupported, fmt.Sprintf("%d states no duration for its aura", row.ID))
	}
	return parts, unsupported
}

func registerSpellDataConsumable(itemID int32) {
	core.NewSpellDataConsumableEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		consumable := core.GetConsumableByID(itemID)
		itemEffect := consumable.OnUse()
		row := spelldata.MustFind(itemEffect.BuffId)
		parts, unsupported := consumablePartsOf(row)
		if len(unsupported) > 0 {
			panic(fmt.Sprintf("consumable %d casts %d, which the builder cannot run: %v", itemID, row.ID, unsupported))
		}

		actionID := core.ActionID{ItemID: itemID}
		potionBonus := 1.0
		if consumable.Type == proto.ConsumableType_ConsumableTypePotion && character.HasAlchStone() {
			potionBonus += spelldata.MustFind(alchemistsStoneSpellID).EffectN(1).Percent()
		}

		var config core.SpellConfig
		switch {
		case parts.damage:
			config = consumableDamageSpell(character, itemID, row)
		case parts.heal != nil:
			config = spellDataProcHealSpell(character, row, false)
			config.DamageMultiplier *= potionBonus
		case parts.absorb != nil:
			config = spellDataAbsorbSpell(character, row, itemID, false)
		default:
			config = spelldata.SpellConfig(&character.Unit, row)
		}
		config.ActionID = actionID
		config.Flags &^= core.SpellFlagPassiveSpell | core.SpellFlagProc
		config.Flags |= core.SpellFlagNoOnCastComplete | core.SpellFlagAPL
		switch consumable.Type {
		case proto.ConsumableType_ConsumableTypePotion:
			// Encounter-only, so a prepull potion is one of the potion placeholders and the rotation
			// cannot name one the consumes do not.
			config.Flags |= core.SpellFlagEncounterOnly | core.SpellFlagPotion
		case proto.ConsumableType_ConsumableTypeConjured:
			config.Flags |= core.SpellFlagConjured
		}

		config.Cast = consumableCast(character, itemEffect, row)
		sharedCD := config.Cast.SharedCD.Duration

		cdType := core.CooldownTypeUnknown
		if parts.damage {
			cdType = core.CooldownTypeDPS
		}
		var applyAura func(*core.Simulation)
		if parts.hasAura() {
			aura := character.GetOrRegisterAura(spelldata.AuraConfig(row))
			spelldata.ParseEffects(character, aura, row, spelldata.Effects(parts.auraIdx...))
			config.RelatedSelfBuff = aura
			applyAura = func(sim *core.Simulation) { aura.Activate(sim) }
			cdType = (&core.StatBuffAura{BuffedStatTypes: consumable.Stats.GetBuffedStatTypes()}).InferCDType()
		}

		var gains []*consumableGain
		for i := range row.Effects {
			e := &row.Effects[i]
			if e == parts.heal || e == parts.absorb {
				cdType = core.CooldownTypeSurvival
				continue
			}
			if !slices.Contains(parts.gains, e) {
				continue
			}
			gain := newConsumableGain(character, row, e, actionID, potionBonus)
			if gain == nil {
				continue
			}
			gains = append(gains, gain)
			switch {
			case gain.power != dbcenums.POWER_MANA:
				cdType = core.CooldownTypeDPS
			case cdType != core.CooldownTypeSurvival:
				cdType = core.CooldownTypeMana
			}
		}

		if parts.gainsOnly() && len(gains) == 0 {
			return
		}

		var selfDamage *core.Spell
		if parts.selfDamage != nil {
			selfConfig := spelldata.SpellConfig(&character.Unit, row, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc())
			selfConfig.ActionID = actionID
			selfConfig.Flags &^= core.SpellFlagHelpful
			selfConfig.DefenseType = damageDefenseType(selfConfig.DefenseType, selfConfig.SpellSchool, false)
			selfDamage = registerCasterDamageSpell(character, selfConfig, parts.selfDamage)
		}

		ownEffects := config.ApplyEffects
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if !parts.damage {
				target = &character.Unit
			}
			if ownEffects != nil {
				ownEffects(sim, target, spell)
			}
			for i := range gains {
				gains[i].apply(sim)
			}
			if selfDamage != nil {
				// Applied without a cast, so the hit logs no cast and draws no timeline row of its own, like
				// an explosive's hit on its thrower.
				selfDamage.ApplyEffects(sim, &character.Unit, selfDamage)
			}
			if applyAura != nil {
				applyAura(sim)
			}
			if sim.CurrentTime < 0 && sharedCD > 0 {
				spell.SharedCD.Set(sim.CurrentTime + sharedCD)
				character.UpdateMajorCooldowns()
			}
		}

		spell := character.RegisterSpell(config)

		consumables := character.Consumables
		mcd := core.MajorCooldown{Spell: spell, Type: cdType}
		if parts.gainsOnly() {
			mcd.ShouldActivate = func(sim *core.Simulation, character *core.Character) bool {
				for i := range gains {
					if !gains[i].fits() {
						return false
					}
				}
				return true
			}
		}

		switch consumable.Type {
		case proto.ConsumableType_ConsumableTypePotion:
			if !character.IsPickedPotion(itemID) {
				return
			}
		case proto.ConsumableType_ConsumableTypeConjured:
			// The spec's conjured item keeps the manager's default, so a healthstone waits for a tank
			// like any other defensive cooldown; another conjured item never fires on its own.
			if consumables.ConjuredId != itemID {
				mcd.Manual = true
			}
		}
		character.AddMajorCooldown(mcd)
	})
}

// The item's own cooldown and the category cooldown it shares, which every consumable builder reads.
// -1 on the item's row, and a category of 0, defer to the spell's row: the Spellblasting potions
// 250932-250937 state neither and share potion category 4 for 120 s through their spells. A category
// with no cooldown shares nothing, and is 0 here.
type consumableCooldowns struct {
	cooldownMs, category, categoryMs int32
}

func consumableCooldownsOf(itemEffect *proto.ItemEffect, row *spelldata.Spell) consumableCooldowns {
	onUse := itemEffect.GetOnUse()
	c := consumableCooldowns{cooldownMs: onUse.CooldownMs, category: onUse.CategoryId, categoryMs: onUse.CategoryCooldownMs}
	if c.cooldownMs < 0 {
		c.cooldownMs = row.CooldownMs
	}
	if c.category <= 0 {
		c.category = int32(row.Category)
	}
	if c.categoryMs < 0 {
		c.categoryMs = row.CategoryCooldownMs
	}
	if c.category <= 0 || c.categoryMs <= 0 {
		c.category, c.categoryMs = 0, 0
	}
	return c
}

func (c consumableCooldowns) longest() time.Duration {
	return core.DurationFromMillis(max(c.cooldownMs, c.categoryMs, 0))
}

// The cast the spell's row states, on the item's cooldown and the category cooldown it shares.
func consumableCast(character *core.Character, itemEffect *proto.ItemEffect, row *spelldata.Spell) core.CastConfig {
	c := consumableCooldownsOf(itemEffect, row)
	cast := spelldata.Cast(row)
	if c.cooldownMs > 0 {
		cast.CD = core.Cooldown{Timer: character.NewTimer(), Duration: core.DurationFromMillis(c.cooldownMs)}
	}
	if c.category > 0 {
		cast.SharedCD = core.Cooldown{Timer: character.GetOrInitSpellCategoryTimer(c.category), Duration: core.DurationFromMillis(c.categoryMs)}
	}
	return cast
}

// One gain of mana, rage or energy, at once or every period of the row's duration. A character
// without the bar gets none.
type consumableGain struct {
	power dbcenums.PowerType
	apply func(*core.Simulation)
	// Whether the whole gain fits under the bar's maximum, allowing for five seconds of mana regen.
	fits func() bool
}

func newConsumableGain(character *core.Character, row *spelldata.Spell, e *spelldata.Effect, actionID core.ActionID, potionBonus float64) *consumableGain {
	power := dbcenums.PowerType(e.Misc)
	var maximum, current func() float64
	var add func(*core.Simulation, float64, *core.ResourceMetrics)
	var metrics *core.ResourceMetrics
	bonus := 1.0
	switch {
	case power == dbcenums.POWER_MANA && character.HasManaBar():
		maximum, current, add, metrics = character.MaxMana, character.CurrentMana, character.AddMana, character.NewManaMetrics(actionID)
		bonus = potionBonus
	case power == dbcenums.POWER_RAGE && character.HasRageBar():
		maximum, current, add, metrics = character.MaximumRage, character.CurrentRage, character.AddRage, character.NewRageMetrics(actionID)
	case power == dbcenums.POWER_ENERGY && character.HasEnergyBar():
		maximum, current, add, metrics = character.MaximumEnergy, character.CurrentEnergy, character.AddEnergy, character.NewEnergyMetrics(actionID)
	default:
		return nil
	}

	amount := func(roll func(*spelldata.Effect) float64) float64 {
		if e.Type == dbcenums.E_ENERGIZE_PCT {
			return maximum() * e.Percent() * bonus
		}
		return barAmount(power, roll(e)) * bonus
	}

	ticks := 1
	gain := &consumableGain{power: power}
	if e.Aura == dbcenums.A_PERIODIC_ENERGIZE {
		ticks = int(row.DurationMs / e.PeriodMs)
		gain.apply = func(sim *core.Simulation) {
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   e.Period(),
				NumTicks: ticks,
				Priority: core.ActionPriorityDOT,
				OnAction: func(sim *core.Simulation) {
					add(sim, amount(func(e *spelldata.Effect) float64 { return e.Roll(sim) }), metrics)
				},
			})
		}
	} else {
		gain.apply = func(sim *core.Simulation) {
			add(sim, amount(func(e *spelldata.Effect) float64 { return e.Roll(sim) }), metrics)
		}
	}

	gain.fits = func() bool {
		room := maximum() - current()
		if power == dbcenums.POWER_MANA {
			room -= character.ManaRegenPerSecondWhileCasting() * 5
		}
		return room >= amount(func(e *spelldata.Effect) float64 { return e.Max() })*float64(ticks)
	}
	return gain
}
