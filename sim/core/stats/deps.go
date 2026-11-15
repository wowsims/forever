package stats

import (
	"fmt"
	"math"
	"slices"
)

// This stat list is arranged such that evaluating dependencies in this order
// is always safe, given the set of dependencies actually used in the game.
//
// Note that many stats are omitted from this list, because they are not used
// in any dependencies.
var safeDepsOrder = []Stat{
	Strength,
	Agility,
	Stamina,
	Intellect,
	Spirit,
	DefenseRating,
	BonusArmor,
	ArmorNoMods,
	Armor,
	FeralAttackPower,
	AttackPower,
	RangedAttackPower,
	SpellDamage,
	HealingPower,
	Health,
	Mana,
	MP5,
	HasteRating,
	CritRating,
	HitRating,
	BlockRating,
	BlockPercent,
	DodgeRating,
	ParryRating,
	ExpertiseRating,
	BlockValue,
	ArmorPenetration,
	SpellPiercing,
	SpellCritPercent,
	PhysicalCritPercent,
	SpellHitPercent,
	PhysicalHitPercent,
	DodgePercent,
	ParryPercent,
	ReducedCritTakenPercent,
	ExpertisePercent,
}

func isSafeDep(s Stat) bool {
	return slices.Contains(safeDepsOrder, s)
}
func isValidDep(src Stat, dst Stat) bool {
	if !isSafeDep(src) || !isSafeDep(dst) {
		return false
	}

	// Check that src occurs before dst in the list
	for _, v := range safeDepsOrder {
		if v == src {
			return true
		} else if v == dst {
			return false
		}
	}
	return false
}
func validateDep(src Stat, dst Stat) {
	if !isValidDep(src, dst) {
		panic("Invalid stat dependency: " + src.StatName() + " --> " + dst.StatName())
	}
}

type StatDependency struct {
	dynamic bool
	enabled bool
	src     Stat
	dst     Stat

	// Note that amount is treated differently depending on whether src and dst
	// stats are the same.
	amount float64

	// Where non-zero, src counts in whole steps of this size only.
	step float64

	// Where set, src adds to dst whole, outside dst's own multipliers.
	unmultiplied bool
}

func (sd StatDependency) String() string {
	if sd.src == sd.dst {
		return fmt.Sprintf("%s *= %0.2f", sd.src.StatName(), sd.amount)
	} else {
		return fmt.Sprintf("%s += %s * %0.2f", sd.dst.StatName(), sd.src.StatName(), sd.amount)
	}
}

// Updates the "amount" field for a previously defined dep. Note that if this is
// called after the stats measurement phase, then ApplyStatDependencies() must
// be immediately called afterwards!
func (dep *StatDependency) UpdateValue(newAmount float64) {
	dep.amount = newAmount
}

// Manages dependencies between stats.
//
// Some examples:
// Increases your AP by 30% of your Int
// Increases agility by X%
// Reduces armor by 50%
type StatDependencyManager struct {
	deps      []*StatDependency
	finalized bool
}

func NewStatDependencyManager() StatDependencyManager {
	return StatDependencyManager{}
}

func (sdm *StatDependencyManager) AddStatDependency(src Stat, dst Stat, amount float64) {
	validateDep(src, dst)
	if sdm.IsFinalized() {
		panic("StatDependencyManager already finalized!")
	}
	if src == dst {
		panic("For same-stat dependencies, use MultiplyStat instead!")
	}

	sdm.deps = append(sdm.deps, &StatDependency{
		dynamic: false,
		enabled: true,
		src:     src,
		dst:     dst,
		amount:  amount,
	})
}

// Adds amount of dst per point of src, counting src in whole steps of step
// points only.
func (sdm *StatDependencyManager) AddFlooredStatDependency(src Stat, dst Stat, step float64, amount float64) {
	sdm.AddStatDependency(src, dst, amount)
	sdm.deps[len(sdm.deps)-1].step = step
}

// Adds all of src to dst after dst's own multipliers, which scale the rest of dst and leave src's
// share whole. Every dependency that reads dst counts all of it.
func (sdm *StatDependencyManager) AddUnmultipliedStatDependency(src Stat, dst Stat) {
	sdm.AddStatDependency(src, dst, 1)
	sdm.deps[len(sdm.deps)-1].unmultiplied = true
}

func (sdm *StatDependencyManager) MultiplyStat(s Stat, amount float64) {
	validateDep(s, s)
	if sdm.IsFinalized() {
		panic("StatDependencyManager already finalized!")
	}

	sdm.deps = append(sdm.deps, &StatDependency{
		dynamic: false,
		enabled: true,
		src:     s,
		dst:     s,
		amount:  amount,
	})
}

func (sdm *StatDependencyManager) NewDynamicStatDependency(src Stat, dst Stat, amount float64) *StatDependency {
	validateDep(src, dst)
	if sdm.IsFinalized() {
		panic("StatDependencyManager already finalized!")
	}
	if src == dst {
		panic("For same-stat dependencies, use NewDynamicMultiplyStat instead!")
	}

	dep := &StatDependency{
		dynamic: true,
		enabled: false,
		src:     src,
		dst:     dst,
		amount:  amount,
	}
	sdm.deps = append(sdm.deps, dep)
	return dep
}

func (sdm *StatDependencyManager) NewDynamicMultiplyStat(s Stat, amount float64) *StatDependency {
	validateDep(s, s)
	if sdm.IsFinalized() {
		panic("StatDependencyManager already finalized!")
	}

	dep := &StatDependency{
		dynamic: true,
		enabled: false,
		src:     s,
		dst:     s,
		amount:  amount,
	}
	sdm.deps = append(sdm.deps, dep)
	return dep
}

func (sdm *StatDependencyManager) sortDeps() {
	deps := make([]*StatDependency, 0, len(sdm.deps))

	// By looping through the stats in order of safeDeps, we guarantee proper
	// sorting of dependencies.
	for i, srcStat := range safeDepsOrder {
		for _, dstStat := range safeDepsOrder[i:] {
			// Combine all static deps into 1 for performance.
			startAmount := 0.0
			if srcStat == dstStat {
				startAmount = 1
			}

			amount := startAmount
			for _, dep := range sdm.deps {
				if dep.src != srcStat || dep.dst != dstStat {
					continue
				}

				if dep.dynamic || dep.step != 0 || dep.unmultiplied {
					// Dynamic deps need to remain separate, so
					// they can be turned on/off, floored ones
					// so they keep their step, and unmultiplied
					// ones so they stay outside the multipliers.
					deps = append(deps, dep)
				} else {
					if srcStat == dstStat {
						amount *= dep.amount
					} else {
						amount += dep.amount
					}
				}
			}

			if amount != startAmount {
				deps = append(deps, &StatDependency{
					enabled: true,
					src:     srcStat,
					dst:     dstStat,
					amount:  amount,
				})
			}
		}
	}

	sdm.deps = deps
}

func (sdm *StatDependencyManager) FinalizeStatDeps() {
	if sdm.IsFinalized() {
		panic("StatDependencyManager already finalized!")
	}
	sdm.sortDeps()
	sdm.finalized = true
}

func (sdm *StatDependencyManager) ResetStatDeps() {
	for _, dep := range sdm.deps {
		if dep.dynamic {
			dep.enabled = false
		}
	}
}

func (sdm *StatDependencyManager) IsFinalized() bool {
	return sdm.finalized
}

// ApplyStatDependenciesInPlace is ApplyStatDependencies without copying the array.
func (sdm *StatDependencyManager) ApplyStatDependenciesInPlace(s *Stats) {
	var unmultiplied Stats
	for _, dep := range sdm.deps {
		if dep.enabled {
			if dep.src == dep.dst {
				s[dep.dst] = (s[dep.dst]-unmultiplied[dep.dst])*dep.amount + unmultiplied[dep.dst]
			} else if dep.unmultiplied {
				s[dep.dst] += s[dep.src]
				unmultiplied[dep.dst] += s[dep.src]
			} else if dep.step != 0 {
				s[dep.dst] += math.Floor(s[dep.src]/dep.step) * dep.step * dep.amount
			} else if isFlooredGameStat[dep.src] {
				// The dep sort guarantees the source stat is final here, and
				// the game floors attributes before dependents consume them
				// (e.g. dodge is derived from the floored Agility).
				s[dep.dst] += math.Floor(s[dep.src]) * dep.amount
			} else {
				s[dep.dst] += s[dep.src] * dep.amount
			}
		}
	}
}

func (sdm *StatDependencyManager) ApplyStatDependencies(s Stats) Stats {
	sdm.ApplyStatDependenciesInPlace(&s)
	return s
}

func (sdm *StatDependencyManager) SortAndApplyStatDependencies(s Stats) Stats {
	sdm.sortDeps()
	return sdm.ApplyStatDependencies(s)
}

// GetStatDependencyCoeff returns the combined linear coefficient for how much of dst
// is gained per unit of src from all currently-enabled dependencies.
func (sdm *StatDependencyManager) GetStatDependencyCoeff(src, dst Stat) float64 {
	for _, dep := range sdm.deps {
		if dep.enabled && dep.src == src && dep.dst == dst {
			return dep.amount
		}
	}
	return 0
}

// Returns whether the state changed.
func (sdm *StatDependencyManager) EnableDynamicStatDep(dep *StatDependency) bool {
	if !dep.enabled {
		dep.enabled = true
		return true
	}
	return false
}

// Returns whether the state changed.
func (sdm *StatDependencyManager) DisableDynamicStatDep(dep *StatDependency) bool {
	if dep.enabled {
		dep.enabled = false
		return true
	}
	return false
}
