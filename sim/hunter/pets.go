package hunter

import (
	"cmp"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var cobraReflexesRank = spellData.CobraReflexesTriggered.Highest()

// The Faster Attack and Slower Attack passives a tamed creature carries on top of the 2 s swing,
// by the option that names them.
var petAttackSpeeds = map[proto.HunterOptions_PetAttackSpeed]spelldata.Ladder{
	proto.HunterOptions_SlowerAttackIII: spellData.SlowerAttackIII,
	proto.HunterOptions_SlowerAttackII:  spellData.SlowerAttackII,
	proto.HunterOptions_FasterAttackI:   spellData.FasterAttackI,
	proto.HunterOptions_FasterAttackII:  spellData.FasterAttackII,
	proto.HunterOptions_FasterAttackIII: spellData.FasterAttackIII,
	proto.HunterOptions_FasterAttackIV:  spellData.FasterAttackIV,
	proto.HunterOptions_FasterAttackV:   spellData.FasterAttackV,
	proto.HunterOptions_FasterAttackVI:  spellData.FasterAttackVI,
	proto.HunterOptions_FasterAttackVII: spellData.FasterAttackVII,
}

// A tamed pet's own armor, health and weapon are the creature values for its level in the game files
// (core.CreatureArmorMaxLevel, core.CreatureHealthMaxLevel, core.CreatureAutoAttackDpsMaxLevel), and its own
// attack power is 30% of the creature dps as attack power, 14 a point of dps, with nothing from Strength:
// 36, 61 and 103 on the sheets of a level 11 Carrion Bird, a level 17 Cat and a level 30 pet, each the
// creature dps × 4.2 rounded down, with the owner's 10% rounded down beside it.
// TODO: Health is one reading, 550 with Great Stamina 2 at 35 owner Stamina, which the creature health
// fits with ~1.1 health a point of the owner's Stamina, and half of it with the warlock's 7.5. The creature
// health is kept until a second reading at another Stamina.
var petBaseStats = stats.Stats{
	stats.Health:      core.CreatureHealthMaxLevel,
	stats.Armor:       core.CreatureArmorMaxLevel,
	stats.AttackPower: 0.3 * 14 * core.CreatureAutoAttackDpsMaxLevel,
}

// The swing before attack power is the creature dps over the 2.0 s, spanning 20% either side: a level 11
// Carrion Bird's 20-27 at 44 attack power and 2.0 s, and a level 17 Cat's 28-37 and 27-37 at 75 and 72,
// swinging at 1.5 s with Faster Attack III, are the creature dps plus attack power over 14, over the swing
// as sped up, with the Cat's +10%. The sheet leaves happiness out.
const petSwingAverage = core.CreatureAutoAttackDpsMaxLevel * 2

var petSwing = core.Weapon{
	BaseDamageMin: petSwingAverage * (1 - 0.2),
	BaseDamageMax: petSwingAverage * (1 + 0.2),
	SwingSpeed:    2,
	MaxRange:      core.MaxMeleeRange,
}

type HunterPet struct {
	core.Pet

	hunterOwner *Hunter
	family      generatedPetFamily

	// What the pet casts in melee range, longest cooldown first, and what it casts to close the
	// distance when it is not.
	abilities []*core.Spell
	closers   []*core.Spell

	uptimePercent float64
}

func (hunter *Hunter) NewHunterPet() *HunterPet {
	family, ok := petFamilies[hunter.Options.PetType.String()]
	if !ok || hunter.Options.PetUptime <= 0 {
		return nil
	}

	hp := &HunterPet{
		Pet: core.NewPet(core.PetConfig{
			Name:                  family.Name,
			Owner:                 &hunter.Character,
			BaseStats:             petBaseStats,
			StatInheritance:       hunter.makeStatInheritance(),
			EnabledOnStart:        true,
			StartsAtOwnerDistance: true,
		}),
		hunterOwner: hunter,
		family:      family,
	}

	// Bestial Discipline's effect 1 raises every effect of the pet's focus passive: +10% a rank.
	hp.EnableFocusBar(spellData.BestialDiscipline.EffectAt(1).MultiplierAt(hunter.Talents.BestialDiscipline))

	// Normalized in ApplyTalents to the Faster or Slower Attack passive's swing.
	hp.EnableAutoAttacks(hp, core.AutoAttackOptions{
		MainHand:       petSwing,
		AutoSwingMelee: true,
	})

	// Happiness
	hp.PseudoStats.DamageDealtMultiplier *= 1.25

	hunter.AddPet(hp)
	return hp
}

// Hunter Pet Scaling (415429) states every inherited stat without an amount, the server pricing
// them: 2 health a point of stamina, 30% of armor, 10% of the higher attack power, the owner's
// higher crit (inheritedCrit), the owner's higher hit and the owner's expertise, the armor, attack
// power and crit read on a level 17 pet sheet at two gear states. Not linear, so the pet snapshots them;
// registerStatInheritance says when.
// TODO: The health a point of Stamina is not read; see petBaseStats.
func (hunter *Hunter) makeStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		return stats.Stats{
			stats.Health:              ownerStats[stats.Stamina] * 2,
			stats.Armor:               ownerStats[stats.Armor] * 0.3,
			stats.AttackPower:         max(ownerStats[stats.AttackPower], ownerStats[stats.RangedAttackPower]) * 0.1,
			stats.PhysicalCritPercent: hunter.inheritedCrit(ownerStats),
			// The owner's hit whole, the higher of melee and ranged, as crit is.
			stats.PhysicalHitPercent: max(0, ownerStats[stats.PhysicalHitPercent], ownerStats[stats.PhysicalHitPercent]+ownerStats[stats.RangedHitPercent]),
			// TODO: 415429 effect 17 (A_MOD_COMBAT_RESULT_CHANCE2, misc 4) states expertise without an
			// amount; the full share is unmeasured.
			stats.ExpertisePercent: ownerStats[stats.ExpertisePercent],
		}
	}
}

// The higher of the melee and ranged crit the in-game character sheet shows, at least 0. Melee is the
// owner's physical crit; ranged adds the ranged crit. Each loses 0.04% a point of its weapon's skill
// below the level-based maximum.
// TODO: A level 17 pet sheet read the owner's ranged crit whole at 84 of 85 bow skill, where this
// takes 0.04% off; a reading at a skill well below the maximum decides.
func (hunter *Hunter) inheritedCrit(ownerStats stats.Stats) float64 {
	mhSkillBonus := core.GetWeaponSkill(&hunter.Unit, hunter.GetMHWeapon())
	rangedSkillBonus := core.GetWeaponSkill(&hunter.Unit, hunter.GetRangedWeapon())
	melee := ownerStats[stats.PhysicalCritPercent] + min(mhSkillBonus, 0)*0.04
	ranged := ownerStats[stats.PhysicalCritPercent] + ownerStats[stats.RangedCritPercent] + min(rangedSkillBonus, 0)*0.04
	return max(0, melee, ranged)
}

// The server's HunterPetInheritance: the pet snapshots its owner's stats when summoned and again on
// any action it takes, at most once every 2 s.
func (hp *HunterPet) registerStatInheritance() {
	hp.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Hunter Pet Inheritance",
		Callback: core.CallbackOnSpellHitDealt | core.CallbackOnCastComplete,
		ICD:      time.Second * 2,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			hp.RefreshInheritedStats(sim)
		},
	})
}

// The family's own damage, armor and health, and what Beast Training taught the pet that changes
// its damage: each is a passive the pet is never without, applied as its effects read. Great
// Stamina, Natural Armor, Pet Hardiness and Pet Resistance change nothing the sim measures.
func (hp *HunterPet) ApplyTalents() {
	opts := hp.hunterOwner.Options
	// The family's armor share is stated on armor from items, which a pet has none of: its armor is
	// its owner's, inherited, and the second effect is left out.
	spelldata.ParseStatic(&hp.Character, hp.family.Passive.Highest(), spelldata.SkipEffects(2))
	spelldata.ParseStatic(&hp.Character, spellData.PetAggression.Rank(opts.PetAggression))

	// An attack speed acts through a Simulation, which a static passive has none of, so these two
	// are auras the pet is never without.
	attackSpeed := petAttackSpeeds[opts.PetAttackSpeed].Highest()
	hp.applyPassiveAura(attackSpeed)
	if attackSpeed != spelldata.Nil {
		// Normalized: a faster swing deals less a swing, the damage a second unchanged.
		hp.AutoAttacks.MHConfig().DamageMultiplier /= 1 + attackSpeed.Effect(dbcenums.A_MOD_ATTACKSPEED, 0).Percent()
	}
	if opts.CobraReflexes {
		hp.applyPassiveAura(cobraReflexesRank)
		// TODO: In-game testing required. "but reduces damage" states no number; TBC's 15% kept.
		hp.AutoAttacks.MHConfig().DamageMultiplier *= 0.85
	}
}

func (hp *HunterPet) applyPassiveAura(rank *spelldata.Spell) {
	if rank == spelldata.Nil {
		return
	}
	aura := hp.RegisterAura(spelldata.AuraConfig(rank, spelldata.Permanent()))
	spelldata.ParseEffects(&hp.Character, aura, rank)
}

func (hp *HunterPet) GetPet() *core.Pet {
	return &hp.Pet
}

func (hp *HunterPet) Initialize() {
	hp.Pet.Initialize()
	hp.registerStatInheritance()
	for _, ladder := range hp.family.Abilities {
		hp.registerAbility(ladder.Highest())
	}
	slices.SortStableFunc(hp.abilities, func(a, b *core.Spell) int {
		return cmp.Compare(max(b.CD.Duration, b.SharedCD.Duration), max(a.CD.Duration, a.SharedCD.Duration))
	})
}

func (hp *HunterPet) Reset(sim *core.Simulation) {
	hp.uptimePercent = min(1, max(0, hp.hunterOwner.Options.PetUptime))
}

func (hp *HunterPet) OnEncounterStart(_ *core.Simulation) {
}

func (hp *HunterPet) ExecuteCustomRotation(sim *core.Simulation) {
	target := hp.CurrentTarget
	if hp.DistanceFromTarget > core.MaxMeleeRange {
		for _, closer := range hp.closers {
			if closer.CanCast(sim, target) {
				closer.Cast(sim, target)
				break
			}
		}
		if !hp.Moving {
			hp.MoveTo(core.MaxMeleeRange-1, sim)
		}
		return
	}

	if sim.GetRemainingDurationPercent() < 1.0-hp.uptimePercent { // once fight is % completed, disable pet.
		hp.Disable(sim)
		return
	}

	for _, spell := range hp.abilities {
		if spell.CanCast(sim, target) {
			spell.Cast(sim, target)
			return
		}
		// A cooldown that is up but short of focus is waited for rather than spent on a filler.
		if (spell.CD.Duration > 0 || spell.SharedCD.Duration > 0) && spell.IsReady(sim) {
			return
		}
	}
}

// An ability by its row's name in the game files, so the generated family table needs no hand list
// beside it. A name not here stays unregistered: Cower, Growl, Prowl and Shell Shield change nothing
// the sim measures, and the Boar's Charge is not modelled.
func (hp *HunterPet) registerAbility(rank *spelldata.Spell) {
	var spell *core.Spell
	switch rank.Name {
	case "Bite", "Claw", "Swipe", "Dismember", "Pinch", "Mine!", "Lightning Breath", "Thunderstomp", "Lava Breath":
		spell = hp.registerStrike(rank, nil)
	case "Demoralizing Screech":
		spell = hp.registerDemoralizingScreech(rank)
	case "Web", "Savage Rend", "Tendon Rip":
		spell = hp.registerDot(rank)
	case "Scorpid Poison":
		spell = hp.registerScorpidPoison(rank)
	case "Dust Cloud":
		spell = hp.registerDustCloud(rank)
	case "Furious Howl":
		spell = hp.registerFuriousHowl(rank)
	case "Trickster's Dance":
		spell = hp.registerSelfBuff(rank)
	case "Dash", "Dive":
		hp.closers = append(hp.closers, hp.registerSprint(rank))
		return
	default:
		return
	}
	hp.abilities = append(hp.abilities, spell)
}

// The row's config for a spell of the pet's: out of the owner's rotation, and cast only while the
// pet is up.
func (hp *HunterPet) spellConfig(rank *spelldata.Spell, opts ...spelldata.SpellOpt) core.SpellConfig {
	config := spelldata.SpellConfig(&hp.Unit, rank, opts...)
	config.Flags &^= core.SpellFlagAPL
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return hp.IsEnabled()
	}
	return config
}

// The proc mask and metrics a damaging row wants: a physical row is a melee special, any other a
// spell of its school.
func schoolOpt(rank *spelldata.Spell) spelldata.SpellOpt {
	if rank.SpellSchool() == core.SpellSchoolPhysical {
		return spelldata.Melee(core.ProcMaskMeleeMHSpecial)
	}
	return spelldata.Magic(core.ProcMaskSpellDamage)
}

// How the row's application is rolled: a landed-or-not roll on the row's defense type.
func hitOutcome(spell *core.Spell, rank *spelldata.Spell) core.OutcomeApplier {
	switch rank.DefenseTypeCore() {
	case core.DefenseTypeMagic:
		return spell.OutcomeMagicHit
	case core.DefenseTypeRanged:
		return spell.OutcomeRangedHit
	default:
		return spell.OutcomeMeleeSpecialHit
	}
}

func hitAndCritOutcome(spell *core.Spell, rank *spelldata.Spell) core.OutcomeApplier {
	switch rank.DefenseTypeCore() {
	case core.DefenseTypeMagic:
		return spell.OutcomeMagicHitAndCrit
	case core.DefenseTypeRanged:
		return spell.OutcomeRangedHitAndCrit
	default:
		return spell.OutcomeMeleeSpecialHitAndCrit
	}
}

// How many enemies the damage reaches: a chain states its count, an area around the pet takes the
// row's cap, anything else hits the target alone.
func targetsOf(rank *spelldata.Spell, damage *spelldata.Effect) int32 {
	if damage.ChainTargets > 0 {
		return int32(damage.ChainTargets)
	}
	if damage.Target[0] == dbcenums.TARGET_UNIT_SRC_AREA_ENEMY && rank.MaxTargets > 0 {
		return int32(rank.MaxTargets)
	}
	return 1
}

// A direct hit, on one enemy or several, with the row's roll. onCast runs after the damage whatever
// it did, for a row whose second effect is not the hit's.
func (hp *HunterPet) registerStrike(rank *spelldata.Spell, onCast func(sim *core.Simulation, target *core.Unit)) *core.Spell {
	damage := rank.DamageEffect()
	targets := targetsOf(rank, damage)

	config := hp.spellConfig(rank, schoolOpt(rank))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if targets > 1 {
			spell.CalcCleaveDamageWithVariance(sim, target, targets, hitAndCritOutcome(spell, rank),
				func(sim *core.Simulation, spell *core.Spell) float64 { return damage.DirectAmount(sim) })
			spell.DealBatchedAoeDamage(sim)
		} else {
			spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), hitAndCritOutcome(spell, rank))
		}
		if onCast != nil {
			onCast(sim, target)
		}
	}
	return hp.RegisterSpell(config)
}

// The hit, then the attack power every enemy around the pet loses, in the slot the raid's
// Demoralizing Shout and Roar share.
func (hp *HunterPet) registerDemoralizingScreech(rank *spelldata.Spell) *core.Spell {
	auras := hp.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataDemoralizingScreechAura(target, true, 0)
	})
	spell := hp.registerStrike(rank, func(sim *core.Simulation, _ *core.Unit) {
		for _, enemy := range sim.Encounter.ActiveTargetUnits {
			auras.Get(enemy).Activate(sim)
		}
	})
	spell.RelatedAuraArrays = auras.ToMap()
	return spell
}

// A periodic effect the row's roll puts on the enemy.
func (hp *HunterPet) registerDot(rank *spelldata.Spell) *core.Spell {
	config := hp.spellConfig(rank, schoolOpt(rank))
	config.Dot = spelldata.DotConfig(rank, rank.PeriodicEffect())
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, hitOutcome(spell, rank))
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}
	return hp.RegisterSpell(config)
}

// The same, stacking: each application refreshes the poison and adds a stack, and a tick deals the
// row's amount once per stack.
func (hp *HunterPet) registerScorpidPoison(rank *spelldata.Spell) *core.Spell {
	tick := rank.PeriodicEffect()

	config := hp.spellConfig(rank, schoolOpt(rank))
	config.Dot = spelldata.DotConfig(rank, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.Average()*float64(dot.GetStacks()), rank.TickOutcome(dot))
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, hitOutcome(spell, rank))
		if result.Landed() {
			dot := spell.Dot(target)
			stacks := dot.GetStacks() + 1
			dot.Apply(sim)
			dot.SetStacks(sim, stacks)
		}
		spell.DealOutcome(sim, result)
	}
	return hp.RegisterSpell(config)
}

// The armor the enemy loses, in the slot the raid's Dust Cloud, Faerie Fire and Curse of
// Recklessness share: the debuff lands beside the others, and only the strongest one's armor counts.
func (hp *HunterPet) registerDustCloud(rank *spelldata.Spell) *core.Spell {
	auras := hp.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		// TODO: verify in game that Curse of Recklessness, Faerie Fire and Dust Cloud share one slot.
		return buffs.NewSpellDataDustCloudAura(target, true, 0)
	})

	config := hp.spellConfig(rank)
	// The row states one stack, and the armor follows the stacks.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		aura := auras.Get(target)
		aura.Activate(sim)
		aura.AddStack(sim)
	}
	config.RelatedAuraArrays = auras.ToMap()
	return hp.RegisterSpell(config)
}

// The pet's and the hunter's attack power. TODO: the row's target is the raid within 15 yd.
func (hp *HunterPet) registerFuriousHowl(rank *spelldata.Spell) *core.Spell {
	var auras []*core.Aura
	for _, character := range []*core.Character{&hp.Character, &hp.hunterOwner.Character} {
		aura := character.RegisterAura(spelldata.AuraConfig(rank))
		spelldata.ParseEffects(character, aura, rank, spelldata.Exclusive(buffs.FuriousHowlCategory, true))
		auras = append(auras, aura)
	}

	config := hp.spellConfig(rank)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		for _, aura := range auras {
			aura.Activate(sim)
		}
	}
	return hp.RegisterSpell(config)
}

// A buff on the pet whose effects the row states.
func (hp *HunterPet) registerSelfBuff(rank *spelldata.Spell) *core.Spell {
	aura := hp.RegisterAura(spelldata.AuraConfig(rank))
	spelldata.ParseEffects(&hp.Character, aura, rank)

	config := hp.spellConfig(rank)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	return hp.RegisterSpell(config)
}

// Dash and Dive: the pet's movement speed for the row's duration.
func (hp *HunterPet) registerSprint(rank *spelldata.Spell) *core.Spell {
	aura := hp.RegisterAura(spelldata.AuraConfig(rank))
	aura.NewActiveMovementSpeedEffect(rank.EffectN(1).Percent())

	config := hp.spellConfig(rank)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	return hp.RegisterSpell(config)
}
