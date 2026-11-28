package dbc

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/overrides"
)

// Walks the chain of spells reachable through SpellEffect.EffectTriggerSpell, visiting each
// spell at most once, since proc chains can loop back on themselves. The effects of a spell have to
// be read in index order or any traversal that stops at the first match becomes non-deterministic
// across runs.
type chainWalker struct {
	visited map[int]bool
}

func newChainWalker() *chainWalker {
	return &chainWalker{visited: map[int]bool{}}
}

// Returns spellID's effects in index order the first time spellID is reached, and nil on every
// later visit so a cyclic chain terminates.
func (w *chainWalker) effects(spellID int) []SpellEffect {
	if w.visited[spellID] {
		return nil
	}
	w.visited[spellID] = true
	return dbcInstance.SpellEffectsInOrder(spellID)
}

// A container aura and the stat aura it accumulates, with what makes a stack land. The container is
// reported separately because the window's duration and its proc flags belong to it rather than to
// the stat aura.
type stackingAura struct {
	Aura        *proto.ItemEffect
	ContainerID int
	Proc        *proto.ProcEffect
	// Set for the trinkets that start at full stacks and lose one per event instead of building
	// up. Read from the container's description, because the decrement is server script.
	Decays bool
}

// Resolves the stacking shape reachable from an item effect's spell, or nil when there is none.
//
// Blackened Naaru Sliver is the canonical case: the item's effect triggers a 20s aura that grants
// nothing, which in turn triggers 10 stacks of attack power with no duration of its own, because in
// game it drops when the window ends. resolveStatsSpell walks straight past the container to reach
// the stats, so the container has to be recovered by walking the chain a second time.
func buildStackingAura(rootSpellID, statsSpellID, itemLevel int) *stackingAura {
	containerID := newChainWalker().findStackingContainer(rootSpellID, statsSpellID)
	if containerID == 0 {
		return nil
	}

	// The aura the edge leads to is not always the one holding the stats: Zandalarian Hero Badge
	// names the real aura only in the container's description text.
	statAuraID := statsSpellID
	if len(collectStats(statAuraID, itemLevel).ToProtoMap()) == 0 {
		if referenced := referencedStatAura(containerID, itemLevel); referenced != 0 {
			statAuraID = referenced
		}
	}

	statAura := dbcInstance.Spells[statAuraID]
	if statAura.MaxCumulativeStacks < 2 {
		// One stack is not a stacking trinket.
		return nil
	}

	// Petrified Scarab ends here: nothing in the data names the aura that actually carries the stats,
	// and nil sends the effect to the missing-effects report to be implemented by hand.
	if len(collectStats(statAuraID, itemLevel).ToProtoMap()) == 0 {
		return nil
	}

	// A container with ProcChance above 100 is the "rate lives somewhere else" sentinel: the rate
	// comes from the hand-supplied rates keyed by the container's own spell, and with none there the
	// effect goes to the missing-effects report.
	container := dbcInstance.Spells[containerID]
	proc := &proto.ProcEffect{IcdMs: procIcdMs(container, dbcInstance.Spells[statAuraID])}
	switch {
	case container.ProcChance > 100:
		ppm := overrides.PPMFor(int32(containerID))
		if ppm == 0 {
			return nil
		}
		proc.ProcRate = &proto.ProcEffect_Ppm{Ppm: ppm}
	case container.ProcChance > 0:
		proc.ProcRate = &proto.ProcEffect_ProcChance{ProcChance: float64(container.ProcChance) / 100}
	}

	return &stackingAura{
		Aura: &proto.ItemEffect{
			BuffId:              int32(statAuraID),
			BuffName:            statAura.NameLang,
			MaxCumulativeStacks: statAura.MaxCumulativeStacks,
			ScalingOptions:      map[int32]*proto.ScalingItemEffectProperties{},
		},
		ContainerID: containerID,
		Proc:        proc,
		Decays:      stacksDecay.MatchString(container.Description),
	}
}

// A $<spellID><token> reference inside a spell description, e.g. the "$24575m1" in Zandalarian
// Hero Badge's "Increases your armor by ${$24575m1*10}". Four digits is the shortest real spell
// ID, which keeps the plain $s1 / $t1 / $d tokens out.
var descriptionSpellRef = regexp.MustCompile(`\$(\d{4,7})[a-zA-Z]`)

// Trinkets that start full and lose a stack per event say so in prose and nowhere else: the
// decrement is server script, with no effect or flag behind it.
var stacksDecay = regexp.MustCompile(`(?i)(bonus|this) is reduced by`)

// Resolves the stat aura a container names only in its description text, or 0 when the description
// references nothing that resolves to stats.
func referencedStatAura(containerSpellID int, itemLevel int) int {
	// Mining the English description is the only link there is, so an ambiguous one has to be
	// visible rather than silently resolved by ordering.
	chosen := 0
	// Counted per distinct spell: one aura referenced as both $24575m1 and $24575m2 is one candidate.
	resolved := map[int]bool{}

	for _, match := range descriptionSpellRef.FindAllStringSubmatch(dbcInstance.Spells[containerSpellID].Description, -1) {
		referenced, err := strconv.Atoi(match[1])
		if err != nil || referenced == containerSpellID || resolved[referenced] {
			continue
		}
		if dbcInstance.Spells[referenced].ID == 0 {
			continue
		}
		if len(collectStats(referenced, itemLevel).ToProtoMap()) > 0 {
			resolved[referenced] = true
			if chosen == 0 {
				chosen = referenced
			}
		}
	}

	if len(resolved) > 1 {
		fmt.Printf("WARN: spell %d's description names %d stat-resolving spells, taking the first (%d)\n", containerSpellID, len(resolved), chosen)
	}

	return chosen
}

// Finds the aura that bounds a stacking stat aura: the spell in the chain above statAuraID that
// triggers it and carries the duration it lives inside.
//
// The stat aura having no duration of its own is what separates this shape from an ordinary proc
// that happens to stack, where the buff expires on its own schedule and refreshing it is correct.
func (w *chainWalker) findStackingContainer(spellID, statAuraID int) int {
	statAura := dbcInstance.Spells[statAuraID]
	if statAura.MaxCumulativeStacks <= 0 || statAura.Duration > 0 {
		return 0
	}

	for _, se := range w.effects(spellID) {
		if se.EffectTriggerSpell == 0 {
			continue
		}

		if se.EffectTriggerSpell == statAuraID &&
			isProcTriggerAura(se.EffectAura) &&
			dbcInstance.Spells[spellID].Duration > 0 {
			return spellID
		}

		if found := w.findStackingContainer(se.EffectTriggerSpell, statAuraID); found != 0 {
			return found
		}
	}
	return 0
}

// The first spell of the store on the trigger chain from spellID, spellID itself included, that
// matches, or 0. A spell the store does not carry ends its branch.
func ResolveTriggered(spellID int, matches func(*spelldata.Spell) bool) int32 {
	return newChainWalker().resolveTriggered(spellID, matches)
}

func (w *chainWalker) resolveTriggered(spellID int, matches func(*spelldata.Spell) bool) int32 {
	s := spelldata.Find(int32(spellID))
	if s == spelldata.Nil || w.visited[spellID] {
		return 0
	}

	effects := w.effects(spellID)
	if matches(s) {
		return s.ID
	}
	for _, se := range effects {
		if !se.CastsItsTrigger() {
			continue
		}
		if found := w.resolveTriggered(se.EffectTriggerSpell, matches); found != 0 {
			return found
		}
	}
	return 0
}

func isProcTriggerAura(aura EffectAuraType) bool {
	return aura.IsProcTrigger()
}

// Attaches an accumulating aura to the effect that bounds it.
//
// The effect is rebased onto the container: resolveStatsSpell walked past it to reach the stats,
// leaving buff_id and the duration pointing at an aura that grants nothing by itself and never
// expires on its own.
func applyStackingAura(pe *proto.ItemEffect, stacking *stackingAura) {
	pe.StackingAura = stacking.Aura
	pe.StackTrigger = &proto.ItemEffect_StackProc{StackProc: stacking.Proc}
	pe.StacksDecay = stacking.Decays

	container := dbcInstance.Spells[stacking.ContainerID]
	pe.BuffId = int32(stacking.ContainerID)
	pe.BuffName = container.NameLang
	pe.EffectDurationMs = container.Duration
	// The stacks belong to the child; the container carries none and would otherwise leave the
	// flat stacking path in the sim competing with this one.
	pe.MaxCumulativeStacks = 0
}
