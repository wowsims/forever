package core

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The internals the tests of the generated buffs read. Those tests sit in package core_test, beside
// this file, because they import sim/core/buffs, which imports core; that import is also what
// registers the buff hooks for every other test in this directory.

func NewGeneratedBuffTestCharacter() *Character {
	char := &Character{
		Unit: Unit{
			Type:        PlayerUnit,
			Level:       60,
			auraTracker: newAuraTracker(),
			Env:         &Environment{},
		},
	}
	char.PseudoStats = stats.NewPseudoStats()
	return char
}

func NewGeneratedDebuffTestTarget() *Unit {
	target := &Unit{
		Type:        EnemyUnit,
		Index:       0,
		Level:       63,
		auraTracker: newAuraTracker(),
		Env:         &Environment{MeasuringStats: true, Raid: &Raid{}},
	}
	target.PseudoStats = stats.NewPseudoStats()
	return target
}

func ApplyBuffEffects(char *Character, raid *proto.RaidBuffs, party *proto.PartyBuffs, individual *proto.IndividualBuffs) {
	registeredBuffs().ApplyBuffs(char, raid, party, individual)
}

func ApplyDebuffEffects(target *Unit, debuffs *proto.Debuffs, raid *proto.Raid) {
	registeredBuffs().ApplyDebuffs(target, debuffs, raid)
}

func (character *Character) ApplyBuildPhaseAuras(phase CharacterBuildPhase) {
	character.applyBuildPhaseAuras(phase)
}

func (unit *Unit) SetStats(s stats.Stats) {
	unit.stats = s
}

func (category *ExclusiveCategory) Effects() []*ExclusiveEffect {
	return category.effects
}

// Narrows the loaded consumables to these ids until the returned function restores the rest, for a
// test that runs the sim on the consumables a browser request carries.
func KeepOnlyConsumables(ids []int32) (restore func()) {
	dbMu.Lock()
	defer dbMu.Unlock()
	all := consumablesByID
	consumablesByID = make(map[int32]Consumable, len(ids))
	for _, id := range ids {
		if consumable, ok := all[id]; ok {
			consumablesByID[id] = consumable
		}
	}
	return func() {
		dbMu.Lock()
		consumablesByID = all
		dbMu.Unlock()
	}
}

// Narrows the loaded items to these ids until the returned function restores the rest, for a test
// that runs the sim on the items a browser request carries.
func KeepOnlyItems(ids []int32) (restore func()) {
	dbMu.Lock()
	defer dbMu.Unlock()
	all := itemsByID
	itemsByID = make(map[int32]Item, len(ids))
	for _, id := range ids {
		if item, ok := all[id]; ok {
			itemsByID[id] = item
		}
	}
	return func() {
		dbMu.Lock()
		itemsByID = all
		dbMu.Unlock()
	}
}
