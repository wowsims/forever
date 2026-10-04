// What the item procs the sim registered before it read them off the rows resolve to. The table is
// the sixteen registrations sim/common/forever/stat_bonus_procs_auto_gen.go carried at 60266be6f7,
// transcribed by hand, and every field is what that generated call stated. Two of them state a
// number the game files contradict, one a mask that leaves out the heals the game files' mask
// names, three are no longer registered at all, one required damage where a weapon proc needs only
// a landed hit, and three required damage where their rows now take any damage component; each says
// so here.
//
// A pin rather than a comparison: the generated file carries spell ids now, so there are no literals
// left to compare against, and what this guards is that resolving those ids still produces the
// listener the sim used to fire.

package spelldata

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
)

// One registration as the generated file stated it before the ids replaced it.
type liveProc struct {
	itemID         int32
	name           string
	triggerSpellID int32
	isWeaponProc   bool

	callback           core.AuraCallback
	procMask           core.ProcMask
	outcome            core.HitOutcome
	requireDamageDealt bool
	canProcFromProcs   bool

	requireDamageComponent bool

	// The rate the registration fired at. A procs-per-minute rate is a manager and no roll, which is
	// what ppm says.
	chance float64
	ppm    float64
	icdMs  int32

	// What the generated call stated for the chance where the game files contradict it, and why. The
	// table carries the resolved value, so every field is asserted either way; the note is what a
	// reader of this row needs, not a licence for the other fields to drift.
	statedChance float64
	chanceNote   string
	// Set where the rows refuse the proc, which is why the item registers no effect any more.
	unsupported string
}

const meleeAllMask = core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto |
	core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial

const shieldSpikeMask = meleeAllMask | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial

func liveProcs() []liveProc {
	shieldSpike := func(itemID int32, trigger int32, name string) liveProc {
		return liveProc{itemID: itemID, name: name, triggerSpellID: trigger,
			callback: core.CallbackOnSpellHitTaken, procMask: shieldSpikeMask,
			outcome: core.OutcomeLanded, requireDamageDealt: true, chance: 0.05}
	}
	// The generated call required damage. An override rules these to hits with a damage
	// component, so the landed application of a DoT that deals none rolls them too.
	weaponDamage := func(itemID int32, trigger int32, name string) liveProc {
		return liveProc{itemID: itemID, name: name, triggerSpellID: trigger,
			callback: core.CallbackOnSpellHitDealt, procMask: meleeAllMask,
			outcome: core.OutcomeLanded, requireDamageComponent: true, chance: 1}
	}

	const reissue = "the reissued shields carry trigger 1216968, whose Electrostatic Charge is a 20% roll, " +
		"while the generated call stated the 5% of the 13959 the six older shields share. The game files say " +
		"20% too, so the literal is what was out of date."

	procs := []liveProc{
		weaponDamage(12631, 7721, "Fiery Plate Gauntlets"),
		weaponDamage(12632, 16615, "Storm Gauntlets"),
		weaponDamage(17111, 7711, "Blazefury Medallion"),
		shieldSpike(18825, 13959, "Grand Marshal's Aegis"),
		shieldSpike(18826, 13959, "High Warlord's Shield Wall"),
		shieldSpike(234562, 1216968, "High Warlord's Shield Wall (reissue)"),
		shieldSpike(234588, 1216968, "Grand Marshal's Aegis (reissue)"),
		shieldSpike(272591, 13959, "Premier High Warlord's Shield Wall"),
		shieldSpike(272838, 13959, "Premier Grand Marshal's Aegis"),
		{
			// "Chance on hit" beside a column of 100 states no rate.
			itemID: 260205, name: "Highborne Research Tablet", triggerSpellID: 1318159,
			callback: core.CallbackOnSpellHitDealt,
			procMask: shieldSpikeMask | core.ProcMaskSpellDamage,
			outcome:  core.OutcomeLanded, chance: 1,
			unsupported: "states no rate",
		},
		{
			itemID: 285278, name: "Satchel of Dark Iron Bombs", triggerSpellID: 1318123,
			callback: core.CallbackOnSpellHitDealt, procMask: core.ProcMaskRangedAuto,
			outcome: core.OutcomeLanded, requireDamageDealt: true, chance: 1,
		},
		{
			// The generated call required damage. A weapon proc fires off any hit that lands, so a
			// special that deals none rolls it too.
			itemID: 12798, name: "Annihilator", triggerSpellID: 16928, isWeaponProc: true,
			callback: core.CallbackOnSpellHitDealt, procMask: core.ProcMaskUnknown,
			outcome: core.OutcomeLanded, ppm: 1,
		},
		{
			// The generated call stated ProcMaskSpellDamage alone. 23688's mask 0x14000 names helpful
			// spells beside harmful ones, so heal casts roll it too.
			itemID: 19288, name: "Darkmoon Card: Blue Dragon", triggerSpellID: 23688,
			callback: core.CallbackOnCastComplete, procMask: core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
			outcome: core.OutcomeEmpty, chance: 0.02,
		},
		{
			itemID: 21190, name: "Wrath of Cenarius", triggerSpellID: 25906,
			callback: core.CallbackOnCastComplete, procMask: core.ProcMaskSpellDamage,
			outcome: core.OutcomeEmpty, canProcFromProcs: true, chance: 0.05,
		},
		{
			itemID: 249473, name: "Dormant Heart of the Mountain", triggerSpellID: 1249118,
			callback: core.CallbackOnHealDealt, procMask: core.ProcMaskSpellHealing,
			outcome: core.OutcomeLanded, canProcFromProcs: true, chance: 1,
			unsupported: "named ability",
		},
		{
			// "Chance on harmful spell cast" beside a column of 100 states no rate.
			itemID: 275630, name: "Depleted Eye of Influence", triggerSpellID: 1297085,
			callback: core.CallbackOnCastComplete, procMask: core.ProcMaskSpellDamage,
			outcome: core.OutcomeEmpty, chance: 1,
			unsupported: "states no rate",
		},
	}

	for i := range procs {
		if procs[i].triggerSpellID == 1216968 {
			procs[i].statedChance = procs[i].chance
			procs[i].chance = 0.2
			procs[i].chanceNote = reissue
		}
	}

	return procs
}

func TestTheItemProcsResolveAsTheyWereRegistered(t *testing.T) {
	withGeneratedStore(t)

	for _, live := range liveProcs() {
		row := Find(live.triggerSpellID)
		if row == Nil {
			t.Errorf("%s (%d): the store does not carry trigger %d", live.name, live.itemID, live.triggerSpellID)
			continue
		}

		unsupported := ItemProcUnsupportedWithRate(row, live.isWeaponProc, live.ppm)
		if live.unsupported != "" {
			if len(unsupported) != 1 || unsupported[0] != live.unsupported {
				t.Errorf("%s (%d): unsupported = %v, want only %q, which is why it registers nothing now",
					live.name, live.itemID, unsupported, live.unsupported)
			}
			continue
		}
		if len(unsupported) > 0 {
			t.Errorf("%s (%d): the rows refuse a proc the sim registers: %v", live.name, live.itemID, unsupported)
			continue
		}

		character := &core.Character{}
		var opts []ProcOpt
		if live.isWeaponProc {
			opts = append(opts, WeaponProc())
		}
		trigger := ProcTrigger(character, row, nil, append(opts, procRateForTest(live.ppm))...)

		report := func(field string, got any, want any) {
			t.Helper()
			t.Errorf("%s (%d) %s: %v, the generated call stated %v", live.name, live.itemID, field, got, want)
		}

		if live.chanceNote != "" {
			t.Logf("%s (%d) ProcChance: %v, where the generated call stated %v - %s",
				live.name, live.itemID, live.chance, live.statedChance, live.chanceNote)
		}

		if trigger.Callback != live.callback {
			report("Callback", trigger.Callback, live.callback)
		}
		if trigger.ProcMask != live.procMask {
			report("ProcMask", trigger.ProcMask, live.procMask)
		}
		if trigger.Outcome != live.outcome {
			report("Outcome", trigger.Outcome, live.outcome)
		}
		if trigger.RequireDamageDealt != live.requireDamageDealt {
			report("RequireDamageDealt", trigger.RequireDamageDealt, live.requireDamageDealt)
		}
		if trigger.RequireDamageComponent != live.requireDamageComponent {
			report("RequireDamageComponent", trigger.RequireDamageComponent, live.requireDamageComponent)
		}
		if trigger.CanProcFromProcs != live.canProcFromProcs {
			report("CanProcFromProcs", trigger.CanProcFromProcs, live.canProcFromProcs)
		}
		if trigger.ProcChance != live.chance {
			report("ProcChance", trigger.ProcChance, live.chance)
		}
		if trigger.ICD != time.Millisecond*time.Duration(live.icdMs) {
			report("ICD", trigger.ICD, time.Millisecond*time.Duration(live.icdMs))
		}

		// A rate stated as procs per minute is a manager and no roll at all, so the manager is
		// asserted: a chance left beside a manager would gate the rate and the manager would never be
		// consulted.
		if live.ppm > 0 && trigger.DPM == nil {
			report("DPM", "no proc manager", live.ppm)
		}
	}
}

// The rate option sim/common/shared gives the resolver, copied rather than called: shared imports
// this package, so a test inside it cannot import shared back. A weapon's rate is its item effect's,
// measured on the weapon carrying it, so the manager is the dynamic one and the item is named to it
// - Annihilator is the only row here that has one.
func procRateForTest(ppm float64) ProcOpt {
	return func(character *core.Character, trigger *core.ProcTrigger) {
		if ppm <= 0 {
			return
		}

		trigger.ProcChance = 0
		trigger.DPM = character.NewDynamicLegacyProcForWeapon(12798, ppm, 0)
	}
}
