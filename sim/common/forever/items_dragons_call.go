package forever

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The Emerald Dragon Whelp 8776 that Dragon's Call's 13049 summons: a guardian at its owner's level that
// takes nothing from its owner. A combat log of a level 57 owner read the same health, armor and
// attack power on all 40 summons while the owner's attack power ran from 981 to 1383, and its health,
// armor and Acid Spit fit the creature values for level 57. It never crits: the log shows one crit in
// 96 landed swings and none in 82 spits.
type emeraldDragonWhelp struct {
	core.Pet

	acidSpit   *core.Spell
	nextSpitAt time.Duration
}

// Health is 1.3 times the creature health for its level, measured at level 57 (5082). Armor is the
// creature armor at ExpectedStatMod 1's share in the game files, 0.51, which reads 1735 at level 57 as
// the log does. The log shows 174 attack power, which the swing below leaves out.
var whelpBaseStats = stats.Stats{
	stats.Health: 1.3 * core.CreatureHealthMaxLevel,
	stats.Armor:  0.51 * core.CreatureArmorMaxLevel,
}

// The 2.0 s swing is fitted at level 57, where the log's swings read 71 to 94 before armor, 80 on
// average: 2 s of the creature dps for the level at 0.73, ExpectedStatMod 1's share, gives 83 ± 15%.
// That row is inferred from the whelp's armor matching its 0.51. A smaller share plus the 174 attack
// power fits as well (0.5 gives 82), so the log does not tell the two apart.
//
// The log glances on 87 of 96 landed swings, and on 79% against targets at the whelp's own level, at
// about 57% of a hit. The sim's attack table glances 10% at that level, and a skill gap that gave the
// log's share would also give far more misses and parries than its 6% and 4%. No rule explains it, so
// the table stays and the swings read high; they were 13% of the whelp's damage.
var whelpSwingAverage = 0.73 * core.CreatureAutoAttackDpsMaxLevel * 2.0
var whelpSwing = core.Weapon{
	BaseDamageMin: whelpSwingAverage * (1 - 0.15),
	BaseDamageMax: whelpSwingAverage * (1 + 0.15),
	SwingSpeed:    2.0,
	MaxRange:      core.MaxMeleeRange,
}

// The log's median waits: the first Acid Spit starts 0.9 s after the summon and each next one 1.0 s
// after the last ends, a start every 4.0 s. A spit's hardcast holds the swing, which comes 2.0 s after
// the cast ends, so the whelp swings as it arrives and spits from then on.
const whelpFirstSpitDelay = 900 * time.Millisecond
const whelpSpitIdle = time.Second

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
		whelp.nextSpitAt = sim.CurrentTime + whelpFirstSpitDelay
	}
	return whelp
}

func (whelp *emeraldDragonWhelp) GetPet() *core.Pet                   { return &whelp.Pet }
func (whelp *emeraldDragonWhelp) Reset(_ *core.Simulation)            {}
func (whelp *emeraldDragonWhelp) OnEncounterStart(_ *core.Simulation) {}

// Acid Spit 9591 deals its base points as a percent of the creature dps for the caster's level, which
// is what the row's creature level scaling attribute (0x80000 on Attributes 0) means: 249.7 at level 57,
// where the log's 82 hits averaged 250.6. Its 5340 mana at level 57 pays for more spits than a summon
// lasts, so the cast takes no cost.
func (whelp *emeraldDragonWhelp) Initialize() {
	whelp.Pet.Initialize()

	row := spelldata.MustFind(9591)
	baseDamage := row.DamageEffect().Percent() * core.CreatureAutoAttackDpsMaxLevel
	config := spelldata.SpellConfig(&whelp.Unit, row, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ManaCost = core.ManaCostOptions{}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		whelp.nextSpitAt = sim.CurrentTime + whelpSpitIdle
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
	}
	whelp.acidSpit = whelp.RegisterSpell(config)
}

func (whelp *emeraldDragonWhelp) ExecuteCustomRotation(sim *core.Simulation) {
	if sim.CurrentTime < whelp.nextSpitAt {
		whelp.WaitUntil(sim, whelp.nextSpitAt)
	} else if !whelp.acidSpit.Cast(sim, whelp.CurrentTarget) {
		whelp.WaitUntil(sim, sim.CurrentTime+whelpSpitIdle)
	}
}
