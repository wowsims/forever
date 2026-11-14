package forever

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The Emerald Dragon Whelp 8776 that Dragon's Call's 13049 summons: a guardian at its owner's level that
// takes nothing from its owner, and never crits (EV-0021).
type emeraldDragonWhelp struct {
	core.Pet

	acidSpit   *core.Spell
	nextSpitAt time.Duration
}

// Health is 1.3 times the creature health for its level. Armor is the creature armor at
// ExpectedStatMod 1's share in the game files, 0.51 (EV-0021).
var whelpBaseStats = stats.Stats{
	stats.Health: 1.3 * core.CreatureHealthMaxLevel,
	stats.Armor:  0.51 * core.CreatureArmorMaxLevel,
}

// The 2.0 s swing is fitted at level 57 (EV-0021): 2 s of the creature dps for the level at 0.73,
// ExpectedStatMod 1's share, gives 83 ± 15%. That row is inferred from the whelp's armor matching its
// 0.51. A smaller share plus the whelp's attack power fits as well (0.5 gives 82), so the log does not
// tell the two apart.
//
// The log glances far more often than the sim's attack table, which glances 10% at that level, and a
// skill gap that gave the log's share would also give far more misses and parries than its 6% and 4%.
// No rule explains it, so the table stays and the swings read high (EV-0021).
var whelpSwingAverage = 0.73 * core.CreatureAutoAttackDpsMaxLevel * 2.0
var whelpSwing = core.Weapon{
	BaseDamageMin: whelpSwingAverage * (1 - 0.15),
	BaseDamageMax: whelpSwingAverage * (1 + 0.15),
	SwingSpeed:    2.0,
	MaxRange:      core.MaxMeleeRange,
}

// The whelp's waits are uniform draws: the first Acid Spit starts 0 to 1.8 s after the summon, each
// next wait is 0.2 to 3.5 s. A spit's hardcast holds the swing, as in the log, so the whelp swings
// only as it arrives and in the waits between casts (EV-0021).
const whelpFirstSpitMax = 1800 * time.Millisecond
const whelpSpitIdleMin = 200 * time.Millisecond
const whelpSpitIdleMax = 3500 * time.Millisecond

func whelpSpitIdle(sim *core.Simulation) time.Duration {
	return time.Duration(sim.RollWithLabel(float64(whelpSpitIdleMin), float64(whelpSpitIdleMax), "Whelp Spit Idle"))
}

func newEmeraldDragonWhelp(owner *core.Character) core.PetAgent {
	whelp := &emeraldDragonWhelp{Pet: core.NewPet(core.PetConfig{
		Name:      "Emerald Dragon Whelp",
		Owner:     owner,
		BaseStats: whelpBaseStats,
		StatInheritance: func(_ stats.Stats) stats.Stats {
			return stats.Stats{}
		},
		IsGuardian:         true,
		StartsInMeleeRange: true,
	})}
	whelp.EnableAutoAttacks(whelp, core.AutoAttackOptions{
		MainHand:       whelpSwing,
		AutoSwingMelee: true,
	})
	whelp.OnPetEnable = func(sim *core.Simulation) {
		whelp.nextSpitAt = sim.CurrentTime + time.Duration(sim.RollWithLabel(0, float64(whelpFirstSpitMax), "Whelp First Spit"))
	}
	return whelp
}

func (whelp *emeraldDragonWhelp) GetPet() *core.Pet                   { return &whelp.Pet }
func (whelp *emeraldDragonWhelp) Reset(_ *core.Simulation)            {}
func (whelp *emeraldDragonWhelp) OnEncounterStart(_ *core.Simulation) {}

// Acid Spit 9591 deals its base points as a percent of the creature dps for the caster's level, which
// is what the row's creature level scaling attribute (0x80000 on Attributes 0) means: 249.7 at level 57
// (EV-0021). Its 5340 mana at level 57 pays for more spits than a summon lasts, so the cast takes no
// cost.
func (whelp *emeraldDragonWhelp) Initialize() {
	whelp.Pet.Initialize()

	row := spelldata.MustFind(9591)
	baseDamage := row.DamageEffect().Percent() * core.CreatureAutoAttackDpsMaxLevel
	config := spelldata.SpellConfig(&whelp.Unit, row, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ManaCost = core.ManaCostOptions{}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		whelp.nextSpitAt = sim.CurrentTime + whelpSpitIdle(sim)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
	}
	whelp.acidSpit = whelp.RegisterSpell(config)
}

func (whelp *emeraldDragonWhelp) ExecuteCustomRotation(sim *core.Simulation) {
	if sim.CurrentTime < whelp.nextSpitAt {
		whelp.WaitUntil(sim, whelp.nextSpitAt)
	} else if !whelp.acidSpit.Cast(sim, whelp.CurrentTarget) {
		whelp.WaitUntil(sim, sim.CurrentTime+whelpSpitIdle(sim))
	}
}
