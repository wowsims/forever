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

// A tamed pet's own armor and health are the creature values for its level in the game files
// (core.CreatureArmorMaxLevel, core.CreatureHealthMaxLevel), and its own attack power is 30% of the creature
// dps (core.CreatureStatsByLevel) as attack power, 14 a point of dps, with nothing from Strength:
// the creature dps × 4.2 rounded down, with the owner's 10% rounded down beside it (EV-0070).
// TODO: Health is one reading (EV-0070). The creature health is kept until a second reading at another
// Stamina.
var petBaseStats = stats.Stats{
	stats.Health:      core.CreatureHealthMaxLevel,
	stats.Armor:       core.CreatureArmorMaxLevel,
	stats.AttackPower: petAttackPower,
}

var petAttackPower = 0.3 * 14 * core.CreatureStatsByLevel[core.CharacterLevel].AutoAttackDps

// The creature dps is the pet's auto attack damage with its own attack power in it, so the swing before
// attack power is the creature dps less that attack power over 14, over the 2.0 s, spanning 20% either side:
// a Cat's logged swings before armor are this swing plus attack power over 14, over the swing as sped up,
// with the Cat's +10% and happiness (EV-0071). The pet sheet shows the whole creature dps as the weapon and leaves
// happiness out, so its damage range is not the damage the pet deals.
var petSwingAverage = (core.CreatureStatsByLevel[core.CharacterLevel].AutoAttackDps - petAttackPower/14) * 2

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

	// What the pet casts in melee range, and what it casts to close the distance when it is not.
	abilities  []*core.Spell
	closers    []*core.Spell
	savageRend *core.Spell

	summonAttackPower float64 // The attack power and health inherited at the summon, kept while the pet is out.
	summonHealth      float64

	uptimePercent float64
	nextAutocast  time.Duration
}

func (hunter *Hunter) NewHunterPet() *HunterPet {
	family, ok := petFamilies[hunter.Options.PetType.String()]
	if !ok || hunter.Options.PetUptime <= 0 {
		return nil
	}

	hp := &HunterPet{
		hunterOwner: hunter,
		family:      family,
	}
	inherit := hunter.makeStatInheritance()
	hp.Pet = core.NewPet(core.PetConfig{
		Name:      family.Name,
		Owner:     &hunter.Character,
		BaseStats: petBaseStats,
		StatInheritance: func(ownerStats stats.Stats) stats.Stats {
			return hp.keepSummonStats(inherit(ownerStats))
		},
		EnabledOnStart: true,
		// Rapid Fire's melee haste reaches the pet this way; the owner's ranged haste does not.
		HasDynamicMeleeSpeedInheritance: true,
	})

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
// higher crit (inheritedCrit), the owner's ranged hit, 17/8 of it as spell hit, and the owner's
// expertise. The armor, attack power and crit are not linear in the owner's (EV-0070), so the pet
// snapshots them; registerStatInheritance says when.
// TODO: The health a point of Stamina is not read; see petBaseStats.
func (hunter *Hunter) makeStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		hit := inheritedHit(ownerStats)

		return stats.Stats{
			stats.Health:              ownerStats[stats.Stamina] * 2,
			stats.Armor:               ownerStats[stats.Armor] * 0.3,
			stats.AttackPower:         max(ownerStats[stats.AttackPower], ownerStats[stats.RangedAttackPower]) * 0.1,
			stats.PhysicalCritPercent: hunter.inheritedCrit(ownerStats),
			// https://github.com/ClassicWoWCommunity/forever-bugs/issues/394#issuecomment-6302980919
			// TODO: Not yet read on a pet sheet in game.
			stats.PhysicalHitPercent: hit,
			stats.SpellHitPercent:    hit * 17 / 8,
			// TODO: 415429 effect 17 (A_MOD_COMBAT_RESULT_CHANCE2, misc 4) states expertise without an
			// amount; the full share is unmeasured.
			stats.ExpertisePercent: ownerStats[stats.ExpertisePercent],
		}
	}
}

// The higher of the melee and ranged crit the in-game character sheet shows, at least 0. Melee is the
// owner's physical crit; ranged adds the ranged crit. Each loses 0.04% a point of its weapon's skill
// below the level-based maximum.
// TODO: A pet sheet read the owner's ranged crit whole at a bow skill just below the maximum
// (EV-0070), where this takes 0.04% off; a reading at a skill well below the maximum decides.
func (hunter *Hunter) inheritedCrit(ownerStats stats.Stats) float64 {
	mhSkillBonus := core.GetWeaponSkill(&hunter.Unit, hunter.GetMHWeapon(), proto.ItemSlot_ItemSlotMainHand)
	rangedSkillBonus := core.GetWeaponSkill(&hunter.Unit, hunter.GetRangedWeapon(), proto.ItemSlot_ItemSlotRanged)
	melee := ownerStats[stats.PhysicalCritPercent] + min(mhSkillBonus, 0)*0.04
	ranged := ownerStats[stats.PhysicalCritPercent] + ownerStats[stats.RangedCritPercent] + min(rangedSkillBonus, 0)*0.04
	return max(0, melee, ranged)
}

// The owner's ranged hit, at least 0: its physical hit plus its ranged-only hit.
func inheritedHit(ownerStats stats.Stats) float64 {
	return max(0, ownerStats[stats.PhysicalHitPercent]+ownerStats[stats.RangedHitPercent])
}

// The attack power and health the pet takes at the summon stay until it is dismissed; the rest follows
// the owner. The pet's attack power holds through Aspect of the Hawk toggles and a gear swap in one
// summon, its health through the owner's Stamina changes, and its armor follows each gear swap
// (EV-0072).
func (hp *HunterPet) keepSummonStats(inherited stats.Stats) stats.Stats {
	// The summon: Enable snapshots the owner before it marks the pet enabled.
	if !hp.IsEnabled() {
		hp.summonAttackPower, hp.summonHealth = inherited[stats.AttackPower], inherited[stats.Health]
	}
	inherited[stats.AttackPower] = hp.summonAttackPower
	inherited[stats.Health] = hp.summonHealth
	return inherited
}

// The server's HunterPetInheritance: the pet snapshots its owner's stats when summoned and again on
// any action it takes, at most once every 2 s, keeping the summon's attack power and health.
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
// Stamina, Natural Armor, Pet Hardiness and Pet Resistance change nothing the sim measures, and no
// Beast Training teaching spell in the game files teaches Pet Aggression.
func (hp *HunterPet) ApplyTalents() {
	opts := hp.hunterOwner.Options
	// The family's armor share is stated on armor from items, which a pet has none of: its armor is
	// its owner's, inherited, and the second effect is left out.
	spelldata.ParseStatic(&hp.Character, hp.family.Passive.Highest(), spelldata.SkipEffects(2))

	// An attack speed acts through a Simulation, which a static passive has none of, so these two
	// are auras the pet is never without. Only a Cat carries a Faster or Slower Attack passive; every
	// other family swings every 2 s.
	if opts.PetType == proto.HunterOptions_Cat {
		hp.applyAttackSpeed(petAttackSpeeds[opts.PetAttackSpeed].Highest())
	}
	// The same attack speed aura as Faster Attack: its "reduces damage" is the normalization.
	if opts.CobraReflexes {
		hp.applyAttackSpeed(cobraReflexesRank)
	}
}

// A pet attack speed is normalized: a faster swing deals less a swing, the damage a second unchanged.
func (hp *HunterPet) applyAttackSpeed(rank *spelldata.Spell) {
	if rank == spelldata.Nil {
		return
	}
	hp.applyPassiveAura(rank)
	hp.AutoAttacks.MHConfig().DamageMultiplier /= 1 + rank.Effect(dbcenums.A_MOD_ATTACKSPEED, 0).Percent()
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
	hp.nextAutocast = 0
}

func (hp *HunterPet) OnEncounterStart(_ *core.Simulation) {
}

func (hp *HunterPet) ExecuteCustomRotation(sim *core.Simulation) {
	target := hp.CurrentTarget
	// The pet starts beside the hunter and walks in to melee range, closing with Charge, Dash or Dive
	// when it has one, and casting on the way what reaches from where it stands, e.g. Lightning Breath.
	if hp.DistanceFromTarget > core.MaxMeleeRange {
		if !hp.castFirstReady(sim, target, hp.closers) {
			hp.castFirstReady(sim, target, hp.abilities)
		}
		if !hp.Moving && hp.DistanceFromTarget > core.MaxMeleeRange {
			hp.MoveTo(core.MaxMeleeRange-1, sim)
		}
		return
	}

	if sim.GetRemainingDurationPercent() < 1.0-hp.uptimePercent { // once fight is % completed, disable pet.
		hp.Disable(sim)
		return
	}

	// Autocast decides in 0.8 s steps, two of them after a cast, and picks evenly among what it can
	// cast; Savage Rend goes out only about one time in nine that it could (EV-0187).
	if sim.CurrentTime >= hp.nextAutocast {
		hp.nextAutocast = sim.CurrentTime + petAutocastStep
		if hp.castRandomReady(sim, target) {
			hp.nextAutocast += petAutocastStep
		}
	}
	hp.WaitUntil(sim, hp.nextAutocast)
}

const petAutocastStep = 800 * time.Millisecond

func (hp *HunterPet) castRandomReady(sim *core.Simulation, target *core.Unit) bool {
	var ready []*core.Spell
	for _, spell := range hp.abilities {
		if !spell.CanCast(sim, target) {
			continue
		}
		if spell == hp.savageRend {
			if sim.RandomFloat("Pet Savage Rend") < 0.113 {
				spell.Cast(sim, target)
				return true
			}
			continue
		}
		ready = append(ready, spell)
	}
	if len(ready) == 0 {
		return false
	}
	ready[int(sim.RandomFloat("Pet Autocast")*float64(len(ready)))].Cast(sim, target)
	return true
}

func (hp *HunterPet) castFirstReady(sim *core.Simulation, target *core.Unit, spells []*core.Spell) bool {
	for _, spell := range spells {
		if spell.CanCast(sim, target) {
			spell.Cast(sim, target)
			return true
		}
	}
	return false
}

// An ability by its row's name in the game files, so the generated family table needs no hand list
// beside it. A name not here stays unregistered: Cower, Growl and Shell Shield change nothing the sim
// measures, and Prowl's 40% slower walk in from the hunter costs more than its +50% first attack gains.
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
	case "Charge":
		hp.closers = append(hp.closers, hp.registerCharge(rank))
		return
	default:
		return
	}
	if rank.Name == "Savage Rend" {
		hp.savageRend = spell
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

// Attack power for the raid, every player and pet of any party, priced at the pet's level. A guardian
// gets none, as with totems, and the logs carry no buff on a guardian (EV-0199).
// A second hunter's wolf shares the aura the first registered on each unit.
// TODO: the row reaches only the raid within 15 yd of the pet, which the sim leaves out, and in the logs
// the hunter gets about half the howls (EV-0199).
func (hp *HunterPet) registerFuriousHowl(rank *spelldata.Spell) *core.Spell {
	auras := hp.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
		if pet, ok := hp.Env.Raid.GetPlayerFromUnit(unit).(core.PetAgent); ok && pet.GetPet().IsGuardian() {
			return nil
		}
		if aura := unit.GetAura(rank.Name); aura != nil {
			return aura
		}
		aura := unit.RegisterAura(spelldata.AuraConfig(rank))
		spelldata.ParseEffects(nil, aura, rank, spelldata.Level(hp.Level), spelldata.Exclusive(buffs.FuriousHowlCategory, true))
		return aura
	})

	config := hp.spellConfig(rank)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		auras.ActivateAll(sim)
	}
	config.RelatedAuraArrays = auras.ToMap()
	return hp.RegisterSpell(config)
}

// The pet runs to the enemy at three times its speed, and its next swing or ability takes the row's
// attack power. The stun the charge triggers changes nothing the sim measures.
func (hp *HunterPet) registerCharge(rank *spelldata.Spell) *core.Spell {
	buff := hp.RegisterAura(spelldata.AuraConfig(rank))
	spelldata.ParseEffects(&hp.Character, buff, rank, spelldata.Effects(2))
	buff.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, aura.MaxStacks)
	})
	spelldata.AttachChargeSpender(&hp.Character, buff, rank, "Charge - Spend")

	config := hp.spellConfig(rank, spelldata.Flags(core.SpellFlagMovesToTarget))
	run := core.MakeDashAura(&hp.Unit, "Charge - Run", config.ActionID, config.Cast.CD.Duration, nil)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		buff.Activate(sim)
		run.Activate(sim)
		hp.MoveTo(core.MaxMeleeRange-1, sim)
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
