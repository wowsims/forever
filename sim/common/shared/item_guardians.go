package shared

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// A creature an item or a set bonus summons as a guardian at its owner's level, as the logs measure it. The game files
// hold no creature row, only the summon's, which states the creature, how many come and for how long; so the
// creature's health, armor, swing and spells are stated as shares of the creature numbers for the level
// (core.CreatureHealthMaxLevel, core.CreatureArmorMaxLevel, core.CreatureAutoAttackDpsMaxLevel). The guardian takes
// nothing from its owner.
type GuardianCreature struct {
	Name   string        // The pet's name; with more than one up at once, numbered.
	Health float64       // Share of the creature health.
	Armor  float64       // Share of the creature armor.
	Swing  GuardianSwing // A zero Speed: the guardian never swings.
	Cast   *GuardianCast // The one spell the guardian casts between swings, or nil.
}

// A physical swing whose normal hit before armor is Share of Speed seconds of the creature dps, Variance either side.
type GuardianSwing struct {
	Share    float64
	Speed    float64
	Variance float64
}

// A spell the guardian casts on its target: the first within FirstMax of the summon, each next one a uniform wait of
// WaitMin to WaitMax after the last one's cast end. A hardcast holds the swing. The wait is drawn as the cast ends,
// before the cast's own rolls, and a cast that fails to start waits a fresh draw from then. Each draw goes under its
// label.
type GuardianCast struct {
	FirstMax              time.Duration
	WaitMin, WaitMax      time.Duration
	FirstLabel, WaitLabel string
	Config                func(guardian *Guardian) core.SpellConfig
}

type Guardian struct {
	core.Pet

	summon     *spelldata.Spell
	creature   GuardianCreature
	cast       *core.Spell
	nextCastAt time.Duration
}

// Builds the guardians an item summons with the item's owner, through core.NewItemPet: as many as the summon's row
// brings at once. Where more can be up than one summon brings, as when the summon rests less than its duration, each
// piece carrying one registers its own under its own name.
func RegisterItemGuardian(itemID int32, summon *spelldata.Spell, creature GuardianCreature) {
	count := summonCount(summon)
	builders := make([]func(owner *core.Character) core.PetAgent, count)
	for i := range builders {
		name := creature.Name
		if count > 1 {
			name = fmt.Sprintf("%s %d", creature.Name, i+1)
		}
		builders[i] = func(owner *core.Character) core.PetAgent {
			return newGuardian(owner, name, summon, creature)
		}
	}
	core.NewItemPet(itemID, builders...)
}

// A use item whose spell summons a guardian; sim/common/forever/item_guardians_auto_gen.go registers every one the
// game files hold. It registers nothing while creatures has no entry for the summon's creature, so the item keeps
// only its stats until the creature is measured. The use casts on the item's own cooldown and the category cooldown it
// shares, and puts the guardians up on the wearer's target.
func NewGuardianOnUse(itemID int32, summonID int32, creatures map[int32]GuardianCreature) {
	summon := spelldata.MustFind(summonID)
	creature, measured := creatures[summon.SummonEffect().Misc]
	if !measured {
		return
	}
	RegisterItemGuardian(itemID, summon, creature)
	registerSpellDataOnUse(itemID, core.CooldownTypeDPS, func(character *core.Character, row *spelldata.Spell) core.SpellConfig {
		guardians := Guardians(character, summon)
		config := spelldata.SpellConfig(&character.Unit, row, castBy(false))
		config.ProcMask = core.ProcMaskEmpty
		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			SummonGuardians(sim, guardians, character.CurrentTarget)
		}
		return config
	})
}

// How many creatures one cast of the summon's row puts up: its base points, one where the row states none.
func summonCount(summon *spelldata.Spell) int {
	return max(1, int(summon.SummonEffect().BasePoints))
}

func newGuardian(owner *core.Character, name string, summon *spelldata.Spell, creature GuardianCreature) *Guardian {
	guardian := &Guardian{summon: summon, creature: creature, Pet: core.NewPet(core.PetConfig{
		Name:  name,
		Owner: owner,
		BaseStats: stats.Stats{
			stats.Health: creature.Health * core.CreatureHealthMaxLevel,
			stats.Armor:  creature.Armor * core.CreatureArmorMaxLevel,
		},
		StatInheritance: func(_ stats.Stats) stats.Stats {
			return stats.Stats{}
		},
		IsGuardian:         true,
		StartsInMeleeRange: true,
	})}
	if swing := creature.Swing; swing.Speed > 0 {
		average := swing.Share * core.CreatureAutoAttackDpsMaxLevel * swing.Speed
		guardian.EnableAutoAttacks(guardian, core.AutoAttackOptions{
			MainHand: core.Weapon{
				BaseDamageMin: average * (1 - swing.Variance),
				BaseDamageMax: average * (1 + swing.Variance),
				SwingSpeed:    swing.Speed,
				MaxRange:      core.MaxMeleeRange,
			},
			AutoSwingMelee: true,
		})
	}
	return guardian
}

func (guardian *Guardian) GetPet() *core.Pet                   { return &guardian.Pet }
func (guardian *Guardian) Reset(_ *core.Simulation)            {}
func (guardian *Guardian) OnEncounterStart(_ *core.Simulation) {}

func (guardian *Guardian) Initialize() {
	guardian.Pet.Initialize()
	cast := guardian.creature.Cast
	if cast == nil {
		return
	}
	guardian.OnPetEnable = func(sim *core.Simulation) {
		guardian.nextCastAt = sim.CurrentTime + time.Duration(sim.RollWithLabel(0, float64(cast.FirstMax), cast.FirstLabel))
	}
	config := cast.Config(guardian)
	applyEffects := config.ApplyEffects
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		guardian.nextCastAt = sim.CurrentTime + guardian.castWait(sim)
		applyEffects(sim, target, spell)
	}
	guardian.cast = guardian.RegisterSpell(config)
}

func (guardian *Guardian) castWait(sim *core.Simulation) time.Duration {
	cast := guardian.creature.Cast
	return time.Duration(sim.RollWithLabel(float64(cast.WaitMin), float64(cast.WaitMax), cast.WaitLabel))
}

// A guardian with no spell only swings, so its rotation sleeps for as long as it can be up.
func (guardian *Guardian) ExecuteCustomRotation(sim *core.Simulation) {
	switch {
	case guardian.cast == nil:
		guardian.WaitUntil(sim, sim.CurrentTime+guardian.summon.Duration())
	case sim.CurrentTime < guardian.nextCastAt:
		guardian.WaitUntil(sim, guardian.nextCastAt)
	case !guardian.cast.Cast(sim, guardian.CurrentTarget):
		guardian.WaitUntil(sim, sim.CurrentTime+guardian.castWait(sim))
	}
}

// The guardians the character owns for a summon's row, in the order it built them.
func Guardians(character *core.Character, summon *spelldata.Spell) []*Guardian {
	var guardians []*Guardian
	for _, pet := range character.PetAgents {
		if guardian, ok := pet.(*Guardian); ok && guardian.summon.ID == summon.ID {
			guardians = append(guardians, guardian)
		}
	}
	return guardians
}

// Puts up as many guardians as the summon's row brings on the target for the row's duration. A summon takes the
// guardians that are not up first, then the first ones.
func SummonGuardians(sim *core.Simulation, guardians []*Guardian, target *core.Unit) {
	if len(guardians) == 0 {
		return
	}
	count := summonCount(guardians[0].summon)
	order := append(core.FilterSlice(guardians, func(g *Guardian) bool { return !g.IsEnabled() }),
		core.FilterSlice(guardians, func(g *Guardian) bool { return g.IsEnabled() })...)
	for _, guardian := range order[:min(count, len(order))] {
		guardian.CurrentTarget = target
		guardian.EnableWithTimeout(sim, guardian, guardian.summon.Duration())
	}
}
