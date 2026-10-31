package shared

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The auras a row puts on the wearer, on each of its pets and on an enemy, each carrying the effects
// that land on that unit. The wearer's and the pets' are registered from the config handed in; the
// enemy's is the row's own aura, one per enemy, shared by every character that applies it, so the
// first to register it parses it and the multiplier reaches every attacker's hits once. A row
// restricted to an area puts nothing on anyone in an encounter outside it.
type spellDataAuras struct {
	wearer      *core.Aura
	wearerStats []stats.Stat
	pets        []*core.Aura
	enemies     core.AuraArray
}

func newSpellDataAuras(character *core.Character, row *spelldata.Spell, config core.Aura, wearerOpts ...spelldata.ParseOpt) *spellDataAuras {
	auras := &spellDataAuras{}
	if row.RequiredAreas != 0 && !character.Env.Encounter.InArea(row.AreaType()) {
		return auras
	}

	if len(spelldata.EffectsOn(row, spelldata.AuraOnWearer)) > 0 {
		auras.wearer = character.RegisterAura(config)
		if effects := spelldata.ItemAuraParsedOn(row, spelldata.AuraOnWearer); len(effects) > 0 {
			auras.wearerStats = spelldata.ParseEffects(character, auras.wearer, row,
				append([]spelldata.ParseOpt{spelldata.Effects(effects...)}, wearerOpts...)...).Stats()
		}
		for _, shield := range spelldata.ItemAuraDamageShields(row) {
			buffs.AttachDamageShield(auras.wearer, row.SpellSchool(), shield.Average(), shield.Coeff())
		}
	}

	if effects := spelldata.ItemAuraParsedOn(row, spelldata.AuraOnPet); len(effects) > 0 {
		for _, pet := range character.Pets {
			if pet.IsGuardian() {
				continue
			}
			aura := pet.RegisterAura(config)
			spelldata.ParseEffects(&pet.Character, aura, row, spelldata.Effects(effects...))
			auras.pets = append(auras.pets, aura)
		}
	}

	if effects := spelldata.ItemAuraParsedOn(row, spelldata.AuraOnEnemy); len(effects) > 0 {
		auras.enemies = enemyDebuffAuras(character, row, config.Duration, effects)
	}

	return auras
}

func (auras *spellDataAuras) activate(sim *core.Simulation, target *core.Unit) {
	auras.wearer.Activate(sim)
	for _, aura := range auras.pets {
		if aura.Unit.IsEnabled() {
			aura.Activate(sim)
		}
	}
	if auras.enemies != nil {
		applyDebuff(sim, auras.enemies.Get(target))
	}
}

// An item or enchant proc whose buff applies auras: the wearer's, its pets' and the debuff on the unit
// the proc answers, for the buff's duration.
func NewSpellDataAuraProc(cfg SpellDataProc, variants []ItemVariant) {
	forEachSpellDataVariant(cfg, variants, registerSpellDataAuraProc)
}

func registerSpellDataAuraProc(cfg SpellDataProc) {
	registerSpellDataRowProc(cfg, nil, func(agent core.Agent, source effectSource, trigger *spelldata.Spell, buff *spelldata.Spell) {
		character := agent.GetCharacter()

		config := spelldata.AuraConfig(buff, spelldata.Label(cfg.Name+" Proc"))
		config.Duration = procBuffDuration(cfg, trigger, buff)
		auras := newSpellDataAuras(character, buff, config)

		proc := spellDataProcListener(character, cfg, source, trigger)
		if proc.ICD == 0 && buff.ID != trigger.ID {
			proc.ICD = buff.CategoryCooldown()
		}
		callback := proc.Callback
		proc.Handler = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			auras.activate(sim, procDamageTarget(character, callback, spell, result))
		}

		triggerAura := source.registerTrigger(character, proc)
		if auras.wearer == nil {
			return
		}

		// The registrations applySpellDataProc makes for a stat buff: an item swap drops an enchant's
		// buff with its weapon or shield, and the stat-proc APL values find a buff that moves stats.
		procAura := &core.StatBuffAura{Aura: auras.wearer, BuffedStatTypes: auras.wearerStats}
		source.registerWeaponEnchantBuff(character, procAura)
		if len(auras.wearerStats) > 0 {
			procAura.Icd = triggerAura.Icd
			character.AddStatProcBuff(source.id, procAura, source.isEnchant, source.eligibleSlots(character))
		}
	})
}

// An item or enchant whose equip spell applies auras: the row's auras on the wearer and on each
// summoned pet for as long as the item or the enchant is worn. An equip spell that only re-applies another aura on a period is read as
// that aura.
//
// The generator emits no equip spell that states spell modifiers: each class registers its own through
// this constructor in its item files, ported or not, and an item whose modifier has no sim value stays
// in the missing effects list.
func NewSpellDataEquipAura(cfg SpellDataProc, variants []ItemVariant) {
	forEachSpellDataVariant(cfg, variants, registerSpellDataEquipAura)
}

func registerSpellDataEquipAura(cfg SpellDataProc) {
	source := cfg.effectSource()

	// Soft fail to allow for overrides for bad effects
	if source.isAlreadyImplemented() {
		return
	}

	row := spelldata.EquipAuraRow(spelldata.MustFind(cfg.TriggerSpellID))

	source.registerEffect(func(agent core.Agent) {
		character := agent.GetCharacter()
		slots := source.eligibleSlots(character)

		// Checked on reset rather than made permanent: a pet resets after its owner's item swap has
		// settled, and would otherwise put the aura back on for an item only in the swap set.
		config := spelldata.AuraConfig(row, spelldata.Label(cfg.Name))
		config.Duration = core.NeverExpires
		// Worn from the start, the aura is part of the gear the character sheet shows.
		config.BuildPhase = core.Ternary(source.isEquipped(character, slots), core.CharacterBuildPhaseGear, core.CharacterBuildPhaseNone)
		config.OnReset = func(aura *core.Aura, sim *core.Simulation) {
			if source.isEquipped(character, slots) {
				aura.Activate(sim)
			}
		}

		var opts []spelldata.ParseOpt
		if hands := source.weaponHands(character, slots); hands != 0 {
			opts = append(opts, spelldata.OnHand(hands))
		}
		auras := newSpellDataAuras(character, row, config, opts...)
		for _, aura := range append([]*core.Aura{auras.wearer}, auras.pets...) {
			if aura != nil {
				source.registerProc(character, aura, slots)
			}
		}
	})
}

// An on-use item whose spell applies auras: the wearer's buff, its pets' and the debuff on the enemy
// it is used on, for the row's duration. One that lowers the wearer's threat is never used on its own:
// the sim gains nothing from less threat and a tank loses by it, so only a timing or an APL action
// that names it uses it.
func NewSpellDataAuraOnUse(itemID int32) {
	registerSpellDataOnUseCooldown(itemID, func(character *core.Character, row *spelldata.Spell) (core.SpellConfig, core.MajorCooldown, bool) {
		auras := newSpellDataAuras(character, row, spelldata.AuraConfig(row))
		cooldown := core.MajorCooldown{Type: onUseAuraCooldownType(row, auras)}
		if lowersThreat(row) {
			cooldown = core.MajorCooldown{Type: core.CooldownTypeUnknown, Manual: true}
		}
		return spellDataOnUseAuraSpell(character, row, auras), cooldown, true
	})
}

// The wearer's buff decides the cooldown: a damage shield makes it a survival cooldown, and its stats
// otherwise sort it the way a stat on-use is sorted.
func onUseAuraCooldownType(row *spelldata.Spell, auras *spellDataAuras) core.CooldownType {
	if len(spelldata.ItemAuraDamageShields(row)) > 0 {
		return core.CooldownTypeSurvival
	}
	return (&core.StatBuffAura{BuffedStatTypes: auras.wearerStats}).InferCDType()
}

func lowersThreat(row *spelldata.Spell) bool {
	for _, i := range spelldata.EffectsOn(row, spelldata.AuraOnWearer) {
		if e := row.EffectN(int(i)); e.Aura == dbcenums.A_MOD_THREAT && e.BasePoints < 0 {
			return true
		}
	}
	return false
}

// A buff the game spends by charges starts with its full count, and the hits its row names spend
// them: Weakness Analyzer's 1291101 drops on the first non-periodic crit.
func spellDataOnUseAuraSpell(character *core.Character, row *spelldata.Spell, auras *spellDataAuras) core.SpellConfig {
	charged := auras.wearer != nil && SpendsItsCharges(row)
	if charged {
		spelldata.AttachChargeSpender(character, auras.wearer, row, auras.wearer.Label+" Charge")
	}

	return core.SpellConfig{
		SpellSchool: row.SpellSchool(),
		ProcMask:    core.ProcMaskEmpty,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			auras.activate(sim, target)
			if charged {
				auras.wearer.SetStacks(sim, int32(row.ProcCharges))
			}
		},
	}
}

// Whether a buff on the wearer drops after its ProcCharges are spent, and its row names the hits that
// spend them. A row stating more than one stack as well is read as a stacking one, whose charges the
// aura cannot keep apart from its stacks.
func SpendsItsCharges(row *spelldata.Spell) bool {
	return row.ProcCharges > 0 && row.MaxStack <= 1 && len(spelldata.EffectsOn(row, spelldata.AuraOnWearer)) > 0 &&
		statesATrigger(row)
}
