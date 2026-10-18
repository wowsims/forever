package warlock

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

type WarlockPet struct {
	core.Pet

	owner  *Warlock
	summon proto.WarlockOptions_Summon

	AutoCastAbilities []*core.Spell
	MinMana           float64      // The mana the pet keeps before it casts again
	ownerAuras        []*core.Aura // the warlock's effects that last while this demon, and no other, is out
	partyAuras        []*core.Aura // the effects on the warlock's party that last while this demon is out
	summonStun        *core.Aura   // Summoning Disorientation, from the start of the warlock's summon until it runs out or the summon dismisses this demon
}

// A demon's level 60 stats and what its own stats convert into, from the team's pet model sheet where no
// share of the creature values below is known.
type petModel struct {
	baseStats stats.Stats // before the conversions below: Health and Mana are the base pools

	healthPerStamina float64
	manaPerIntellect float64
	mp5PerIntellect  float64
	apPerStrength    float64

	// Melee, for the demons that have it: the weapon's damage range and swing, and a multiplier on
	// everything the demon deals.
	meleeMin, meleeMax float64
	swingSpeed         float64
	damageMod          float64
}

// A demon's own health, armor and weapon are fixed shares of the creature values for its level in the game
// files (core.CreatureHealthMaxLevel, core.CreatureArmorMaxLevel, core.CreatureAutoAttackDpsMaxLevel), each
// demon with its own shares.
// TODO: The shares are read at levels 20 and 30 (the Voidwalker's at 30 only), not yet at 60.
//
// The Sayaad's average swing before attack power: 0.698 of the creature dps at 2.0 s. Its sheet's damage range
// spans 22.8% either side of the average at level 20 and 20.7% at 30, so 21.5%.
const sayaadSwingAverage = 0.698 * core.CreatureAutoAttackDpsMaxLevel * 2

// The Voidwalker's average swing before attack power and before its melee is halved: 0.494 of the creature dps
// at 2.0 s, spanning 26.4% either side. The best fit to three level 30 sheets at attack power 96, 108 and 116
// (16-22, 17-23 and 17-24).
const voidwalkerSwingAverage = 0.494 * core.CreatureAutoAttackDpsMaxLevel * 2

// The sheet gives no crit from Agility or Intellect, and models regen as MP5 and Intellect alone. A demon has
// no crit of its own: its crit is the warlock's (see petStatInheritance). The Imp's, Voidwalker's and Sayaad's
// own health and armor are shares of the creature values and the Voidwalker's and Sayaad's attack power is
// fixed, with nothing from Stamina or Strength; the Felhunter has no armor of its own.
var petModels = map[proto.WarlockOptions_Summon]petModel{
	proto.WarlockOptions_Imp: {
		baseStats: stats.Stats{
			stats.Strength:  122,
			stats.Agility:   35,
			stats.Stamina:   86,
			stats.Intellect: 264,
			stats.Spirit:    260,

			stats.Health: 0.359 * core.CreatureHealthMaxLevel,
			stats.Armor:  0.325 * core.CreatureArmorMaxLevel,
			stats.Mana:   576,
			stats.MP5:    95,
		},
		healthPerStamina: 0,
		manaPerIntellect: 5,
		mp5PerIntellect:  0.1,
		apPerStrength:    2,
		damageMod:        1,
	},
	proto.WarlockOptions_Voidwalker: {
		baseStats: meleeDemonStats(1.400 * core.CreatureHealthMaxLevel).Add(stats.Stats{
			stats.Armor: 1.352 * core.CreatureArmorMaxLevel,
			// TODO: 3 × level + 6 as the Sayaad's, read at level 30 only.
			stats.AttackPower: 186,
		}),
		healthPerStamina: 0,
		manaPerIntellect: 11.5,
		apPerStrength:    0,
		meleeMin:         voidwalkerSwingAverage * (1 - 0.264),
		meleeMax:         voidwalkerSwingAverage * (1 + 0.264),
		swingSpeed:       2,
		// Its melee is halved: at level 30, +20 attack power moved its sheet's damage by half of 20 × 2 / 14.
		damageMod: 0.5,
	},
	proto.WarlockOptions_Succubus: {
		baseStats: meleeDemonStats(0.699 * core.CreatureHealthMaxLevel).Add(stats.Stats{
			stats.Armor: 0.630 * core.CreatureArmorMaxLevel,
			// TODO: 3 × level + 6, a line through levels 20 and 30, not yet read at 60.
			stats.AttackPower: 186,
			stats.MP5:         95, // TODO: The Imp's base MP5; the Sayaad's in-combat regen is not measured.
		}),
		healthPerStamina: 0,
		manaPerIntellect: 11.5,
		mp5PerIntellect:  0.1, // TODO: The Imp's per-Intellect MP5; the Sayaad's in-combat regen is not measured.
		apPerStrength:    0,
		meleeMin:         sayaadSwingAverage * (1 - 0.215),
		meleeMax:         sayaadSwingAverage * (1 + 0.215),
		swingSpeed:       2,
		// In the logs Lash of Pain deals (16 + 0.429 × Spell Damage) × 1.10 under 5/5 Unholy Power, with no
		// multiplier of the Sayaad's own.
		damageMod: 1,
	},
	proto.WarlockOptions_Felhunter: {
		baseStats:        meleeDemonStats(0),
		healthPerStamina: 9.5,
		manaPerIntellect: 11.55,
		mp5PerIntellect:  0.02,
		apPerStrength:    2,
		meleeMin:         70,
		meleeMax:         97,
		swingSpeed:       2,
		damageMod:        0.8,
	},
}

// The Voidwalker, Succubus and Felhunter share their level 60 stats and base mana, each with its own base health.
// The Voidwalker's and Succubus's Stamina and Strength convert to nothing (their health and attack power are set
// directly).
func meleeDemonStats(baseHealth float64) stats.Stats {
	return stats.Stats{
		stats.Strength:  129,
		stats.Agility:   85,
		stats.Stamina:   234,
		stats.Intellect: 70,
		stats.Spirit:    150,

		stats.Health: baseHealth,
		stats.Mana:   1066,
	}
}

// What a demon takes from the warlock. Warlock Pet Scaling (416189) names the stats it inherits - health,
// attack power, spell damage, resistances, armor, hit, crit, haste, mana regen - but states no amounts,
// which the server supplies, so the shares are hand-kept. The pet model sheet gives Stamina and Intellect
// a share of 0, yet its totals give every demon more health and ~10,870 more mana than its own stats, and
// the same in-combat regen (its own MP5 only out of combat) - all from the warlock. Health is 7.5 a point of
// the warlock's Stamina.
// TODO: The warlock's share of mana and in-combat regen; resistances and haste.
func (warlock *Warlock) petStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		// The warlock's spell damage, as its own tooltips read it: the better of its two schools
		spellDamage := ownerStats[stats.SpellDamage] + max(ownerStats[stats.ShadowDamage], ownerStats[stats.FireDamage])

		return stats.Stats{
			stats.Health: ownerStats[stats.Stamina] * 7.5,
			// Demon Armor included: a level 30 Imp's own armor is the level 20 share only with it.
			// TODO: In game a demon's armor is fixed at the summon (it held when Demon Armor fell off),
			// while its attack power and spell damage follow the warlock's.
			stats.Armor:       ownerStats[stats.Armor] * 0.35,
			stats.AttackPower: spellDamage * 0.17,
			// A tenth: at 30 spell damage the warlock's tooltip reads "Increases your pet's Attack Power by 5
			// and Spell Damage by 3", and in the logs a Sayaad's Spell Damage is a tenth of the warlock's, plus
			// Demonic Knowledge's own share when taken.
			// TODO: Measured on the Sayaad only.
			stats.SpellDamage: spellDamage * 0.10,

			// TODO: Assumed 1:1, not yet measured.
			stats.PhysicalHitPercent: ownerStats[stats.SpellHitPercent],
			stats.SpellHitPercent:    ownerStats[stats.SpellHitPercent],

			// A demon takes the warlock's spell crit as its melee and spell crit, with no crit of its own:
			// https://github.com/ClassicWoWCommunity/forever-bugs/issues/331#issuecomment-6160342081
			// TODO: Not live yet: in game a demon still has its own crit (melee ~0.7%) and takes only the
			// warlock's "+X% critical strike" gear. Verify in game once the change ships.
			stats.PhysicalCritPercent: ownerStats[stats.SpellCritPercent],
			stats.SpellCritPercent:    ownerStats[stats.SpellCritPercent],
		}
	}
}

func (warlock *Warlock) registerPets() {
	warlock.Imp = warlock.registerPet(proto.WarlockOptions_Imp, "Imp")
	warlock.Voidwalker = warlock.registerPet(proto.WarlockOptions_Voidwalker, "Voidwalker")
	warlock.Succubus = warlock.registerPet(proto.WarlockOptions_Succubus, "Succubus")
	warlock.Felhunter = warlock.registerPet(proto.WarlockOptions_Felhunter, "Felhunter")

	// The demon out when the fight starts, if any; a summon in the APL changes it.
	startPet := warlock.demon(warlock.Options.Summon)
	warlock.ActivePet = startPet
	warlock.RegisterResetEffect(func(sim *core.Simulation) {
		warlock.ActivePet = startPet
	})
}

// The registered demon of that summon, nil for none.
func (warlock *Warlock) demon(summon proto.WarlockOptions_Summon) *WarlockPet {
	for _, pet := range warlock.demons() {
		if pet.summon == summon {
			return pet
		}
	}
	return nil
}

func (warlock *Warlock) registerPet(summon proto.WarlockOptions_Summon, name string) *WarlockPet {
	enabledOnStart := warlock.Options.Summon == summon
	model := petModels[summon]

	pet := &WarlockPet{
		owner:  warlock,
		summon: summon,
		Pet: core.NewPet(core.PetConfig{
			Name:            name,
			Owner:           &warlock.Character,
			BaseStats:       model.baseStats,
			StatInheritance: warlock.petStatInheritance(),
			IsDynamic:       true,
			EnabledOnStart:  enabledOnStart,
		}),
	}
	pet.EnableManaBar()

	// Core gives every character 10 health per Stamina and 2 armor per Agility, and the links add up, so
	// only the difference is added here. A pet gets none of the player's Intellect links from EnableManaBar.
	pet.AddStatDependency(stats.Stamina, stats.Health, model.healthPerStamina-10)
	pet.AddStatDependency(stats.Agility, stats.Armor, -2)
	pet.AddStatDependency(stats.Intellect, stats.Mana, model.manaPerIntellect)
	pet.AddStatDependency(stats.Intellect, stats.MP5, model.mp5PerIntellect)
	pet.AddStatDependency(stats.Strength, stats.AttackPower, model.apPerStrength)

	// Regen is MP5 and Intellect alone in the model: Spirit adds none.
	pet.PseudoStats.SpiritRegenMultiplier = 0
	pet.PseudoStats.DamageDealtMultiplier *= model.damageMod

	if model.meleeMax > 0 {
		pet.EnableAutoAttacks(pet, core.AutoAttackOptions{
			MainHand: core.Weapon{
				BaseDamageMin: model.meleeMin,
				BaseDamageMax: model.meleeMax,
				SwingSpeed:    model.swingSpeed,
			},
			AutoSwingMelee: true,
		})
	}

	// The warlock's effects that last while a demon is out, and those that last while this one is
	pet.OnPetEnable = func(sim *core.Simulation) {
		for _, aura := range warlock.petActiveAuras {
			aura.Activate(sim)
		}
		for _, aura := range pet.ownerAuras {
			aura.Activate(sim)
		}
	}
	pet.OnPetDisable = func(sim *core.Simulation) {
		for _, aura := range warlock.petActiveAuras {
			aura.Deactivate(sim)
		}
		for _, aura := range pet.ownerAuras {
			aura.Deactivate(sim)
		}
		for _, aura := range pet.partyAuras {
			aura.Deactivate(sim)
		}
	}

	warlock.AddPet(pet)
	return pet
}

// The four demons, each registered whichever is out.
func (warlock *Warlock) demons() []*WarlockPet {
	demons := []*WarlockPet{}
	for _, pet := range []*WarlockPet{warlock.Imp, warlock.Voidwalker, warlock.Succubus, warlock.Felhunter} {
		if pet != nil {
			demons = append(demons, pet)
		}
	}
	return demons
}

// An effect of the warlock's that lasts while a demon is out, gained and lost with it.
func (warlock *Warlock) addPetActiveAura(aura *core.Aura) {
	warlock.petActiveAuras = append(warlock.petActiveAuras, aura)
}

func (pet *WarlockPet) GetPet() *core.Pet {
	return &pet.Pet
}

// Each demon registers its own abilities, as the pet initializes.
func (pet *WarlockPet) Initialize() {
	pet.Pet.Initialize()

	// The Voidwalker's Torment and Suffering are threat, and the Felhunter's abilities utility: both melee only.
	switch pet.summon {
	case proto.WarlockOptions_Imp:
		pet.registerFirebolt()
	case proto.WarlockOptions_Succubus:
		pet.registerLashOfPain()
	}

	if pet.owner.DemonicBrandAuras != nil {
		pet.registerDemonicBrand()
	}

	// Summoning Disorientation, for its row's 6 sec. A summon cast in less, under Fel Domination or Master
	// Summoner, dismisses the demon before it runs out.
	disorientation := spelldata.AuraConfig(spelldata.MustFind(1249755))
	pet.summonStun = pet.RegisterStunAura(disorientation.Label, disorientation.ActionID, disorientation.Duration)
}

func (pet *WarlockPet) Reset(_ *core.Simulation) {
}

func (pet *WarlockPet) OnEncounterStart(_ *core.Simulation) {
}

// Casts the first ability that is ready once the pet has its MinMana; otherwise waits until it has
// the mana for the cheapest one, or until the GCD and the cooldown of an ability it has the mana for end.
func (pet *WarlockPet) ExecuteCustomRotation(sim *core.Simulation) {
	// A demon with nothing to cast only swings, which its auto attacks do on their own.
	if len(pet.AutoCastAbilities) == 0 {
		pet.WaitUntil(sim, core.NeverExpires)
		return
	}

	waitUntil := core.NeverExpires

	for _, spell := range pet.AutoCastAbilities {
		if pet.CurrentMana() >= pet.MinMana && spell.CanCast(sim, pet.CurrentTarget) {
			spell.Cast(sim, pet.CurrentTarget)
			return
		}

		needed := max(pet.MinMana, spell.Cost.GetCurrentCost()) - pet.CurrentMana()
		if regen := pet.ManaRegenPerSecondWhileCasting(); needed > 0 && regen > 0 {
			waitUntil = min(waitUntil, sim.CurrentTime+max(core.DurationFromSeconds(needed/regen), time.Millisecond*100))
		} else if needed <= 0 {
			// Enough mana, so the GCD or the cooldown holds the ability back
			if readyAt := max(spell.ReadyAt(), pet.NextGCDAt()); readyAt > sim.CurrentTime {
				waitUntil = min(waitUntil, readyAt)
			}
		}
	}

	// Held back by something with no readiness time of its own: look again shortly.
	if waitUntil == core.NeverExpires {
		waitUntil = sim.CurrentTime + time.Millisecond*100
	}
	pet.WaitUntil(sim, waitUntil)
}

// --- Imp ---

// The Imp's Blood Pact on each player in the warlock's party, the warlock among them and no pets, up while
// the Imp is out. With the Imp out at the start it counts in the build phase. It shares the party buff's
// category, so a Blood Pact ticked in the settings and the Imp's own are worth one buff. Each warlock's
// Imp puts up its own copy, and with two Imps out the strongest applies.
func (warlock *Warlock) registerBloodPact() {
	impAtStart := warlock.Imp.EnabledOnStart()
	auras := warlock.NewPartyAuraArray(func(member *core.Character) *core.Aura {
		aura := buffs.NewSpellDataBloodPactAuraFor(&member.Unit, &warlock.Unit)
		if impAtStart {
			aura.BuildPhase = core.CharacterBuildPhaseBuffs
			// The Imp comes out in the warlock's reset, before the players after it in the party reset.
			core.MakePermanent(aura)
		}
		return aura
	})
	for _, aura := range auras {
		if aura != nil {
			warlock.Imp.partyAuras = append(warlock.Imp.partyAuras, aura)
		}
	}
}

func (pet *WarlockPet) registerFirebolt() {
	rankedSpell := spellData.Firebolt.ByID(11763)
	fireboltHit := rankedSpell.DamageEffect()

	pet.MinMana = rankedSpell.Cost()

	config := spelldata.SpellConfig(&pet.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellImpFireBolt

	// Roll the damage and to hit, then land it after the bolt's travel time
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, fireboltHit.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
	}

	pet.AutoCastAbilities = append(pet.AutoCastAbilities, pet.RegisterSpell(config))
}

// --- Succubus ---

// The 12s cooldown is the row's category cooldown, which SpellConfig reads.
func (pet *WarlockPet) registerLashOfPain() {
	rankedSpell := spellData.LashOfPain.ByID(11780)
	lashHit := rankedSpell.DamageEffect()

	pet.MinMana = rankedSpell.Cost()

	config := spelldata.SpellConfig(&pet.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellSuccubusLashOfPain

	// Roll the damage and to hit
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, lashHit.Average(), spell.OutcomeMagicHitAndCrit)
	}

	pet.AutoCastAbilities = append(pet.AutoCastAbilities, pet.RegisterSpell(config))
}

// --- Demonic Brand ---

// Each of the demon's hits on a target the warlock branded consumes a charge and deals the brand's damage:
// Fire from the Imp (1293698), Shadow from the others (1293697). The rows state no amount; their
// description variables do: ((level - 26) * 1.5 + 14 to 17 + 0.078 * the warlock's spell power of the
// row's school), times the demon's damage multipliers. The demon's own multipliers apply them: Master
// Demonologist, Unholy Power, Soul Link and the demon's base damage modifier.
// TODO: Imp as Fire and the others as Shadow, every hit as an attack, and the hit and crit roll are assumed;
// the 1 Physical damage row (1293699), a threat marker, is left out. The base damage modifier and Soul Link
// may not reach the brand in game, since the rows' multiplier names only the pet's and the school's.
func (pet *WarlockPet) registerDemonicBrand() {
	warlock := pet.owner

	brandRow, schoolDamage := spelldata.MustFind(1293697), stats.ShadowDamage
	if pet.summon == proto.WarlockOptions_Imp {
		brandRow, schoolDamage = spelldata.MustFind(1293698), stats.FireDamage
	}
	brandHit := pet.RegisterSpell(spelldata.SpellConfig(&pet.Unit, brandRow, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc()))

	core.MakePermanent(pet.RegisterAura(core.Aura{
		Label: "Demonic Brand - Consume",
		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell == brandHit || !result.Landed() {
				return
			}
			brand := warlock.DemonicBrandAuras.Get(result.Target)
			if !brand.IsActive() {
				return
			}
			brand.RemoveStack(sim)
			if brand.GetStacks() == 0 {
				brand.Deactivate(sim)
			}

			spellPower := warlock.GetStat(stats.SpellDamage) + warlock.GetStat(schoolDamage)
			base := float64(warlock.Level-int32(brandRow.BaseLevel))*1.5 + 0.078*spellPower
			brandHit.CalcAndDealDamage(sim, result.Target, sim.Roll(base+14, base+17), brandHit.OutcomeMagicHitAndCrit)
		},
	}))
}
