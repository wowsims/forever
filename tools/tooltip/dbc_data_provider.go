package tooltip

import (
	"math"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

type EffectMap map[int]dbc.SpellEffect
type DBCTooltipDataProvider struct {
	DBC        *dbc.DBC
	Modifiers  Modifiers   // The talents' spell modifiers the compiled texts leave to the player's points.
	Conditions *Conditions // What the compiled texts' conditions on spells, auras and passives read off the player.
}

func (d DBCTooltipDataProvider) GetModifiers() Modifiers {
	return d.Modifiers
}

func (d DBCTooltipDataProvider) GetConditions() *Conditions {
	return d.Conditions
}

func GetEffectByIndex(effects map[int]dbc.SpellEffect, index int) *dbc.SpellEffect {
	// quick check
	effect := effects[index]
	if effect.EffectIndex == index {
		return &effect
	}

	// did not find
	for _, e := range effects {
		if e.EffectIndex == index {
			return &e
		}
	}

	return nil
}

// GetSpellPPM implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetSpellPPM(spellId int64) float64 {
	// Table does not exist in TBC

	return float64(0)
}

// GetSpellProcCooldown implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetSpellProcCooldown(spellId int64) time.Duration {
	spell, ok := d.DBC.Spells[int(spellId)]
	if !ok {
		return 0
	}

	return time.Duration(spell.ProcCategoryRecovery) * time.Millisecond
}

func (d DBCTooltipDataProvider) GetSpellMaxTargets(spellId int64) int64 {
	spell, ok := d.DBC.Spells[int(spellId)]
	if !ok {
		return 0
	}

	return int64(spell.MaxTargets)
}

func (d DBCTooltipDataProvider) GetEffectAmplitude(spellId int64, effectIdx int64) float64 {
	effectEntries, ok := d.DBC.SpellEffects[int(spellId)]
	if !ok {
		return 0
	}

	effect := GetEffectByIndex(effectEntries, int(effectIdx))
	if effect == nil {
		return 0
	}

	return effect.EffectAmplitude
}

func (d DBCTooltipDataProvider) GetEffectChainAmplitude(spellId int64, effectIdx int64) float64 {
	effectEntries, ok := d.DBC.SpellEffects[int(spellId)]
	if !ok {
		return 0
	}

	effect := GetEffectByIndex(effectEntries, int(effectIdx))
	if effect == nil {
		return 0
	}

	return effect.EffectChainAmplitude
}

func (d DBCTooltipDataProvider) GetEffectPointsPerResource(spellId int64, effectIdx int64) float64 {
	effectEntries, ok := d.DBC.SpellEffects[int(spellId)]
	if !ok {
		return 0
	}

	effect := GetEffectByIndex(effectEntries, int(effectIdx))
	if effect == nil {
		return 0
	}

	return effect.EffectPointsPerResource
}

// GetEffectMaxTargets implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetEffectMaxTargets(spellId int64, effectIdx int64) int64 {
	effectEntries, ok := d.DBC.SpellEffects[int(spellId)]
	if !ok {
		return 0
	}

	effect := GetEffectByIndex(effectEntries, int(effectIdx))
	if effect == nil {
		return 0
	}

	return int64(effect.EffectChainTargets)
}

// GetSpellProcChance implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetSpellProcChance(spellId int64) float64 {
	spellEntry, ok := d.DBC.Spells[int(spellId)]
	if !ok {
		return 0
	}

	return float64(spellEntry.ProcChance)
}

func (d DBCTooltipDataProvider) GetSpecNum() int64 {
	return 0
}

func (d DBCTooltipDataProvider) GetSpellIcon(spellId int64) string {
	spellEntry, ok := d.DBC.Spells[int(spellId)]
	if !ok {
		return ""
	}

	return spellEntry.IconPath
}

// GetMainHandWeapon implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetMainHandWeapon() *core.Weapon {
	// Item: 103727 as dummy for now
	return &core.Weapon{
		BaseDamageMin: 10257,
		BaseDamageMax: 19050,
		SwingSpeed:    2.6,
	}
}

// GetOffHandWeapon implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetOffHandWeapon() *core.Weapon {
	return &core.Weapon{
		BaseDamageMin: 10257,
		BaseDamageMax: 19050,
		SwingSpeed:    2.6,
	}
}

func (d DBCTooltipDataProvider) GetPlayerLevel() float64 {
	return core.CharacterLevel
}

func (d DBCTooltipDataProvider) GetSpellDescription(spellId int64) string {
	spellEntry, ok := d.DBC.Spells[int(spellId)]
	if !ok {
		return ""
	}

	return spellEntry.Description
}

func (d DBCTooltipDataProvider) GetSpellName(spellId int64) string {
	spellEntry, ok := d.DBC.Spells[int(spellId)]
	if !ok {
		return ""
	}

	return spellEntry.NameLang
}

// GetAttackPower implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetAttackPower() float64 {
	return 1
}

// GetEffectScaledValue implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetEffectScaledValue(spellId int64, effectIdx int64) float64 {
	amount := d.GetEffectAmount(spellId, effectIdx)
	return amount.Base + d.GetAttackPower()*amount.AP + d.GetSpellDamage()*amount.SP
}

// An effect's amount at the character's level: Base for every player, plus AP times the attack power and
// SP times the spell power a damage or healing effect adds. Ranged, School and Heal name which attack power
// and which spell power those are.
type EffectAmount struct {
	Base, AP, SP float64
	Low, High    float64 // The ends of the roll at the character's level; both Base on an effect that does not roll.
	Ranged       bool    // SpellCategories.DefenseType 3: the attack power is ranged attack power.
	School       int32   // The spell's school mask: the spell power is that school's.
	Heal         bool    // A healing effect: the spell power is healing power.
	Kind         string  // A damage or healing effect's kind, direct or periodic, which the player's done multipliers and talents change; "" for any other.
}

// An effect the spell does not have reads 0.
func (d DBCTooltipDataProvider) GetEffectAmount(spellId int64, effectIdx int64) EffectAmount {
	spell, ok := d.DBC.Spells[int(spellId)]
	effect := GetEffectByIndex(d.DBC.SpellEffects[int(spellId)], int(effectIdx))
	if !ok || effect == nil {
		return EffectAmount{}
	}

	amount := EffectAmount{Ranged: spell.DefenseType == 3, School: spell.SchoolMask}
	amount.Base, amount.Low, amount.High = rowAmount(spell, effect)

	shouldScale := false
	switch effect.EffectType {
	case dbcenums.E_SCHOOL_DAMAGE:
		shouldScale, amount.Kind = true, amountDamage
	case dbcenums.E_HEAL:
		shouldScale, amount.Heal, amount.Kind = true, true, amountHeal

	case dbcenums.E_APPLY_AURA,
		dbcenums.E_APPLY_AREA_AURA_ENEMY,
		dbcenums.E_APPLY_AREA_AURA_FRIEND,
		dbcenums.E_APPLY_AREA_AURA_PARTY,
		dbcenums.E_APPLY_AREA_AURA_OWNER,
		dbcenums.E_APPLY_AREA_AURA_RAID,
		dbcenums.E_APPLY_AREA_AURA_PARTY_NONRANDOM,
		dbcenums.E_APPLY_AREA_AURA_PET,
		dbcenums.E_APPLY_AURA_ON_PET:
		switch effect.EffectAura {
		case dbcenums.A_PERIODIC_DAMAGE:
			shouldScale, amount.Kind = true, amountPeriodicDamage
		case dbcenums.A_PERIODIC_HEAL:
			shouldScale, amount.Heal, amount.Kind = true, true, amountPeriodicHeal
		}
	}

	if shouldScale {
		amount.AP = max(effect.BonusCoefficientFromAP, 0)
		amount.SP = max(effect.EffectBonusCoefficient, 0)
	}
	return amount
}

// An effect's amount and the ends of its roll at the character's level, read off the game files rows the way a
// tooltip writes them: the base points plus the gain for each level between the spell's BaseLevel and the
// character's, up to its MaxLevel, and EffectVariance spreading the base points half its width either side. Each
// rounds a half away from zero: Pyroblast rank 6's 191 ± 23.6 plus 12.6 reads 180 to 227. float32 as the spell
// store reads the same rows, whose own ends round outward for the sim.
func rowAmount(spell dbc.Spell, effect *dbc.SpellEffect) (base, low, high float64) {
	level := float32(core.CharacterLevel)
	if spell.MaxLevel > 0 {
		level = min(level, float32(spell.MaxLevel))
	}
	gain := max(level-float32(spell.BaseLevel), 0) * float32(effect.EffectRealPointsPerLevel)
	points := float32(effect.EffectBasePoints)
	side := float32(math.Abs(effect.EffectBasePoints)) * float32(effect.Variance) / 2
	round := func(v float32) float64 { return math.Round(float64(v)) }
	return round(points + gain), round(points - side + gain), round(points + side + gain)
}

func (d DBCTooltipDataProvider) GetEffectBonusCoefficient(spellId int64, effectIdx int64) float64 {
	effect := GetEffectByIndex(d.DBC.SpellEffects[int(spellId)], int(effectIdx))
	if effect == nil {
		return 0
	}
	return effect.EffectBonusCoefficient
}

func (d DBCTooltipDataProvider) GetSpellAuraDescription(spellId int64) string {
	return d.DBC.Spells[int(spellId)].AuraDescription
}

// GetDescriptionVariableString implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetDescriptionVariableString(spellId int64) string {
	spellEntry, ok := d.DBC.Spells[int(spellId)]
	if !ok {
		return ""
	}

	return spellEntry.Variables
}

// GetEffectBaseValue implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetEffectBaseValue(spellId int64, effectIdx int64) float64 {
	return d.GetEffectAmount(spellId, effectIdx).Low
}

// GetEffectMaxValue implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetEffectMaxValue(spellId int64, effectIdx int64) float64 {
	return d.GetEffectAmount(spellId, effectIdx).High
}

// GetEffectPeriod implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetEffectPeriod(spellId int64, effectIdx int64) time.Duration {
	effectEntries, ok := d.DBC.SpellEffects[int(spellId)]
	if !ok {
		return 0
	}

	effect := GetEffectByIndex(effectEntries, int(effectIdx))
	if effect == nil {
		return 0
	}

	return time.Duration(effect.EffectAuraPeriod) * time.Millisecond
}

// GetEffectRadius implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetEffectRadius(spellId int64, effectIdx int64) float64 {
	effectEntries, ok := d.DBC.SpellEffects[int(spellId)]
	if !ok {
		return 0
	}

	effect := GetEffectByIndex(effectEntries, int(effectIdx))
	if effect == nil {
		return 0
	}

	return effect.GetRadiusMax()
}

// GetSpellDuration implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetSpellDuration(spellId int64) time.Duration {
	spell, ok := d.DBC.Spells[int(spellId)]
	if !ok {
		return 0
	}

	if spell.Duration < 0 {
		return 0
	}

	return time.Duration(spell.Duration) * time.Millisecond
}

// The tooltips are rendered for no character, so an effect's amount is the one its row states, with no
// spell power on top: Arcane Detonation 23722 deals 200.
func (d DBCTooltipDataProvider) GetSpellDamage() float64 {
	return 0
}

// GetSpellRange implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetSpellRange(spellId int64) float64 {
	spell, ok := d.DBC.Spells[int(spellId)]
	if !ok {
		return 0
	}

	return float64(spell.MaxRange)
}

// GetStacks implements TooltipDataProvider.
func (d DBCTooltipDataProvider) GetSpellStacks(spellId int64) int64 {
	spell, ok := d.DBC.Spells[int(spellId)]
	if !ok {
		return 0
	}

	if spell.ProcCharges > 0 {
		return int64(spell.ProcCharges)
	}

	if spell.MaxCumulativeStacks > 0 {
		return int64(spell.MaxCumulativeStacks)
	}

	return 0
}

// $u: the stacks the aura reaches, or its charges where it does not stack (Reckoning stores 4 attacks).
func (d DBCTooltipDataProvider) GetSpellMaxStacks(spellId int64) int64 {
	if stacks := d.DBC.Spells[int(spellId)].MaxCumulativeStacks; stacks > 0 {
		return int64(stacks)
	}
	return d.GetSpellStacks(spellId)
}

// $q: an effect's first misc value, such as the mana Resurrection gives back.
func (d DBCTooltipDataProvider) GetEffectMiscValue(spellId int64, effectIdx int64) float64 {
	effect := GetEffectByIndex(d.DBC.SpellEffects[int(spellId)], int(effectIdx))
	if effect == nil || len(effect.EffectMiscValues) == 0 {
		return 0
	}
	return float64(effect.EffectMiscValues[0])
}

// $v: the highest level of target the spell affects (SpellTargetRestrictions.MaxTargetLevel).
func (d DBCTooltipDataProvider) GetSpellMaxTargetLevel(spellId int64) float64 {
	return float64(d.DBC.Spells[int(spellId)].MaxTargetLevel)
}

// HasAura implements TooltipDataProvider.
func (d DBCTooltipDataProvider) HasAura(auraId int64) bool {
	return true
}

// HasPassive implements TooltipDataProvider.
func (d DBCTooltipDataProvider) HasPassive(auraId int64) bool {
	return true
}

// IsMaleGender implements TooltipDataProvider.
func (d DBCTooltipDataProvider) IsMaleGender() bool {
	return true
}

// KnowsSpell implements TooltipDataProvider.
func (d DBCTooltipDataProvider) KnowsSpell(spellId int64) bool {
	return true
}

func (d DBCTooltipDataProvider) GetEffectEnchantValue(enchantId int64, effectIdx int64) float64 {
	enchantInfo, ok := d.DBC.EnchantsByEffectId[int(enchantId)]
	if !ok {
		return 0
	}

	return float64(enchantInfo.EffectPoints[effectIdx])
}

// A provider that reads one spell's effect amounts off another row: a talent rank's, whose base points the
// spell store's curve states rank by rank.
type RankProvider struct {
	DBCTooltipDataProvider
	SpellID int64
	Rank    *spelldata.Spell
}

func (p RankProvider) GetEffectAmount(spellId int64, effectIdx int64) EffectAmount {
	amount := p.DBCTooltipDataProvider.GetEffectAmount(spellId, effectIdx)
	if spellId == p.SpellID {
		if effect := p.Rank.EffectAtIndex(uint8(effectIdx)); effect != spelldata.NilEffect {
			amount.Base, amount.Low, amount.High = effect.Min(), effect.Min(), effect.Max()
		}
	}
	return amount
}

func (p RankProvider) GetEffectScaledValue(spellId int64, effectIdx int64) float64 {
	amount := p.GetEffectAmount(spellId, effectIdx)
	return amount.Base + p.GetAttackPower()*amount.AP + p.GetSpellDamage()*amount.SP
}

func (p RankProvider) GetEffectBaseValue(spellId int64, effectIdx int64) float64 {
	return p.GetEffectAmount(spellId, effectIdx).Low
}

func (p RankProvider) GetEffectMaxValue(spellId int64, effectIdx int64) float64 {
	return p.GetEffectAmount(spellId, effectIdx).High
}
