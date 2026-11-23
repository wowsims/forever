package core

import (
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

type DynamicProcManager struct {
	procMasks   []ProcMask
	procChances []DynamicProc
}

type DynamicProc interface {
	Reset()
	Chance(sim *Simulation) float64
	Proc(sim *Simulation, label string) bool
}

func (dpm *DynamicProcManager) Reset() {
	for _, proc := range dpm.procChances {
		proc.Reset()
	}
}

// Returns whether the effect procced.
func (dpm *DynamicProcManager) Proc(sim *Simulation, procMask ProcMask, label string) bool {
	for i, m := range dpm.procMasks {
		if m.Matches(procMask) {
			return dpm.procChances[i].Proc(sim, label)
		}
	}

	return false
}

func (dpm *DynamicProcManager) Chance(procMask ProcMask, sim *Simulation) float64 {
	chance, _ := dpm.chanceFor(procMask, sim)
	return chance
}

// The chance the manager states for a hit of this proc mask, and whether any of its masks hears it.
func (dpm *DynamicProcManager) chanceFor(procMask ProcMask, sim *Simulation) (float64, bool) {
	for i, m := range dpm.procMasks {
		if m.Matches(procMask) {
			return dpm.procChances[i].Chance(sim), true
		}
	}

	return 0, false
}

// A chance that depends on the hit being rolled. A trigger asks it about each hit; Chance answers
// without one.
type hitDynamicProc interface {
	DynamicProc
	ChanceOn(sim *Simulation, spell *Spell, result *SpellResult) float64
}

// The chance for this hit from the entry the proc mask the listener hears it with matches, asked
// about the hit itself when the entry prices on it, and whether any entry hears it.
func (dpm *DynamicProcManager) chanceOn(sim *Simulation, procMask ProcMask, spell *Spell, result *SpellResult) (float64, bool) {
	for i, m := range dpm.procMasks {
		if m.Matches(procMask) {
			if hit, ok := dpm.procChances[i].(hitDynamicProc); ok {
				return hit.ChanceOn(sim, spell, result), true
			}
			return dpm.procChances[i].Chance(sim), true
		}
	}

	return 0, false
}

// A procs-per-minute rate on hits taken, priced on the attacker's time between its own swings at
// the hit, slows included: its main hand's interval, or both hands together when it dual-wields. An
// attacker with no melee auto attack has no interval and never procs it. scale multiplies the chance
// per attacker. Read without a hit, the chance is the one against the wearer's current target.
func (character *Character) NewAttackerSwingPPMManager(ppm float64, procMask ProcMask, scale func(attacker *Unit) float64) *DynamicProcManager {
	return &DynamicProcManager{
		procMasks:   []ProcMask{procMask},
		procChances: []DynamicProc{attackerSwingProc{character: character, ppm: ppm, scale: scale}},
	}
}

type attackerSwingProc struct {
	character *Character
	ppm       float64
	scale     func(attacker *Unit) float64
}

func (ap attackerSwingProc) Reset() {}
func (ap attackerSwingProc) Chance(_ *Simulation) float64 {
	if ap.character.CurrentTarget == nil {
		return 0
	}
	return ap.chanceAgainst(ap.character.CurrentTarget)
}
func (ap attackerSwingProc) ChanceOn(_ *Simulation, spell *Spell, _ *SpellResult) float64 {
	return ap.chanceAgainst(spell.Unit)
}
func (ap attackerSwingProc) Proc(sim *Simulation, label string) bool {
	return sim.Proc(ap.Chance(sim), label)
}

func (ap attackerSwingProc) chanceAgainst(attacker *Unit) float64 {
	attacks := &attacker.AutoAttacks
	interval := attacks.MainhandSwingSpeed().Seconds()
	if attacks.IsDualWielding {
		interval = 1 / (1/interval + 1/attacks.OffhandSwingSpeed().Seconds())
	}
	return ap.ppm * interval / 60 * ap.scale(attacker)
}

// A fixed chance per hit, multiplied by scale for the hit's target. Read without a hit, the chance is
// the one against the wearer's current target.
func (character *Character) NewTargetScaledChanceManager(chance float64, procMask ProcMask, scale func(target *Unit) float64) *DynamicProcManager {
	return &DynamicProcManager{
		procMasks:   []ProcMask{procMask},
		procChances: []DynamicProc{targetScaledProc{character: character, chance: chance, scale: scale}},
	}
}

type targetScaledProc struct {
	character *Character
	chance    float64
	scale     func(target *Unit) float64
}

func (tp targetScaledProc) Reset() {}
func (tp targetScaledProc) Chance(_ *Simulation) float64 {
	if tp.character.CurrentTarget == nil {
		return 0
	}
	return tp.chance * tp.scale(tp.character.CurrentTarget)
}
func (tp targetScaledProc) ChanceOn(_ *Simulation, _ *Spell, result *SpellResult) float64 {
	return tp.chance * tp.scale(result.Target)
}
func (tp targetScaledProc) Proc(sim *Simulation, label string) bool {
	return sim.Proc(tp.Chance(sim), label)
}

// PPMManager for static ProcMasks
func (character *Character) NewLegacyPPMManager(ppm float64, procMask ProcMask) *DynamicProcManager {
	dpm := character.newDynamicWeaponProcManager(ppm, 0, procMask)

	character.RegisterItemSwapCallback(AllWeaponSlots(), func(sim *Simulation, slot proto.ItemSlot) {
		dpm = character.newDynamicWeaponProcManager(ppm, 0, procMask)
	})

	return &dpm
}

// A procs-per-minute rate for weapon hits beside the same rate for spell hits, each spell hit priced
// on its own cast time and never under the 1.5 s global cooldown, so an instant or a channel prices
// at 1.5 s. The weapon hits NewLegacyPPMManager measures keep its speeds: a swing's, and with no auto
// attack the wand's for a ranged hit. Every other hit of the mask except a ranged auto attack rolls
// the rate on the main-hand weapon's own speed, read at the roll, whether or not the character swings
// it: a spell on the ranged hit table with no ranged swing or wand behind it, such as Hammer of Wrath.
// With no main-hand weapon it is the sim's unarmed weapon's speed.
func (character *Character) NewLegacyPPMManagerWithCastTimeSpells(ppm float64, procMask ProcMask) *DynamicProcManager {
	return character.legacyPPMManagerWithSpells(ppm, procMask, castTimeProc{ppm: ppm})
}

// A procs-per-minute rate for weapon hits beside the same rate for every other hit of the mask, on
// the main-hand weapon's own speed read at the roll. The weapon hits NewLegacyPPMManager measures keep
// its speeds: a swing's, and with no auto attack the wand's for a ranged hit. A spell hit, and any
// hit no swing prices except a ranged auto attack, rolls the rate on the main-hand weapon's speed
// whether or not the character swings it. With no main-hand weapon it is the sim's unarmed weapon's
// speed.
func (character *Character) NewLegacyPPMManagerWithMainHandSpells(ppm float64, procMask ProcMask) *DynamicProcManager {
	return character.legacyPPMManagerWithSpells(ppm, procMask, mainHandWeaponSpeedProc{character: character, ppm: ppm})
}

// The weapon hits of the mask on NewLegacyPPMManager's speeds, spell hits on spells, and every
// other hit of the mask except a ranged auto attack on the main-hand weapon's speed.
func (character *Character) legacyPPMManagerWithSpells(ppm float64, procMask ProcMask, spells DynamicProc) *DynamicProcManager {
	weaponHits := procMask &^ ProcMaskSpellDamage
	build := func() DynamicProcManager {
		dpm := character.newDynamicWeaponProcManager(ppm, 0, weaponHits)
		dpm.procMasks = append(dpm.procMasks, procMask&ProcMaskSpellDamage, weaponHits&^ProcMaskRangedAuto)
		dpm.procChances = append(dpm.procChances, spells, mainHandWeaponSpeedProc{character: character, ppm: ppm})
		return dpm
	}
	dpm := build()

	character.RegisterItemSwapCallback(AllWeaponSlots(), func(sim *Simulation, slot proto.ItemSlot) {
		dpm = build()
	})

	return &dpm
}

type mainHandWeaponSpeedProc struct {
	character *Character
	ppm       float64
}

func (mp mainHandWeaponSpeedProc) Reset() {}
func (mp mainHandWeaponSpeedProc) Chance(_ *Simulation) float64 {
	return mp.ppm / 60 * mp.character.WeaponFromMainHand().SwingSpeed
}
func (mp mainHandWeaponSpeedProc) Proc(sim *Simulation, label string) bool {
	return sim.Proc(mp.Chance(sim), label)
}

// PPMManager for static ProcMasks and no item swap callback
func (character *Character) NewStaticLegacyPPMManager(ppm float64, procMask ProcMask) *DynamicProcManager {
	dpm := character.newDynamicWeaponProcManager(ppm, 0, procMask)

	return &dpm
}

// A procs-per-minute rate on the main hand's speed as WeaponSpeedMultiplier leaves it: Seal of the
// Crusader's faster swings each carry 1/1.4 of the chance, so the swings' procs a minute hold, and
// haste leaves the chance alone. Both are read at each roll, so a weapon swap needs no callback.
func (character *Character) NewWeaponSpeedPPMManager(ppm float64, procMask ProcMask) *DynamicProcManager {
	return &DynamicProcManager{
		procMasks:   []ProcMask{procMask},
		procChances: []DynamicProc{weaponSpeedProc{unit: &character.Unit, ppm: ppm}},
	}
}

type weaponSpeedProc struct {
	unit *Unit
	ppm  float64
}

func (wp weaponSpeedProc) Reset() {}
func (wp weaponSpeedProc) Chance(_ *Simulation) float64 {
	return wp.ppm / 60 * wp.unit.AutoAttacks.MH().SwingSpeed / wp.unit.PseudoStats.WeaponSpeedMultiplier
}
func (wp weaponSpeedProc) Proc(sim *Simulation, label string) bool {
	return sim.Proc(wp.Chance(sim), label)
}

// Dynamic Proc Manager for static ProcMasks and no item swap callback
func (character *Character) NewFixedProcChanceManager(fixedProcChance float64, procMask ProcMask) *DynamicProcManager {
	dpm := character.newDynamicWeaponProcManager(0, fixedProcChance, procMask)

	return &dpm
}

// Dynamic Proc Manager for dynamic ProcMasks on weapon enchants
func (character *Character) NewDynamicLegacyProcForEnchant(effectID int32, ppm float64, fixedProcChance float64) *DynamicProcManager {
	return character.newDynamicProcManagerWithDynamicProcMask(ppm, fixedProcChance, func() ProcMask {
		return character.getCurrentProcMaskForWeaponEnchant(effectID)
	})
}

// Dynamic Proc Manager for a weapon enchant heard on its own ProcMask: the mask's weapon hits roll
// on the hands carrying the enchant at their own speeds, and nothing else rolls.
func (character *Character) NewDynamicLegacyProcForEnchantWithMask(effectID int32, ppm float64, procMask ProcMask) *DynamicProcManager {
	return character.newDynamicProcManagerWithDynamicProcMask(ppm, 0, func() ProcMask {
		return procMask & character.getCurrentProcMaskForWeaponEnchant(effectID)
	})
}

// Dynamic Proc Manager for dynamic ProcMasks on weapon temp enchants
func (character *Character) NewDynamicLegacyProcForTempEnchant(effectID int32, ppm float64, fixedProcChanceFn func(ProcMask) float64) *DynamicProcManager {
	return character.newDynamicProcManagerWithDynamicProcMaskDynamicProcChance(ppm, fixedProcChanceFn, func() ProcMask {
		return character.getCurrentProcMaskForWeaponTempEnchant(effectID)
	})
}

// Dynamic Proc Manager for dynamic ProcMasks on weapon effects
func (character *Character) NewDynamicLegacyProcForWeapon(itemID int32, ppm float64, fixedProcChance float64) *DynamicProcManager {
	return character.newDynamicProcManagerWithDynamicProcMask(ppm, fixedProcChance, func() ProcMask {
		return character.getCurrentProcMaskForWeaponEffect(itemID)
	})
}

func (character *Character) newDynamicProcManagerWithDynamicProcMask(ppm float64, fixedProcChance float64, procMaskFn func() ProcMask) *DynamicProcManager {
	dpm := character.newDynamicWeaponProcManager(ppm, fixedProcChance, procMaskFn())
	character.RegisterItemSwapCallback(AllWeaponSlots(), func(sim *Simulation, slot proto.ItemSlot) {
		dpm = character.newDynamicWeaponProcManager(ppm, fixedProcChance, procMaskFn())
	})

	return &dpm
}

func (character *Character) newDynamicProcManagerWithDynamicProcMaskDynamicProcChance(ppm float64, fixedProcChanceFn func(ProcMask) float64, procMaskFn func() ProcMask) *DynamicProcManager {
	procMask := procMaskFn()
	dpm := character.newDynamicWeaponProcManager(ppm, fixedProcChanceFn(procMask), procMask)
	character.RegisterItemSwapCallback(AllWeaponSlots(), func(sim *Simulation, slot proto.ItemSlot) {
		procMask := procMaskFn()
		dpm = character.newDynamicWeaponProcManager(ppm, fixedProcChanceFn(procMask), procMask)
	})

	return &dpm
}

func (character *Character) newDynamicWeaponProcManager(ppm float64, fixedProcChance float64, procMask ProcMask) DynamicProcManager {
	if (ppm != 0) && (fixedProcChance != 0) {
		panic("Cannot simultaneously specify both a ppm and a fixed proc chance!")
	}

	aa := character.AutoAttacks
	rangedSpeed := aa.ranged.SwingSpeed
	if !aa.AutoSwingMelee && !aa.AutoSwingRanged {
		// With no auto attack, only ranged hits roll, on the ranged slot's weapon speed: a caster's wand shots.
		procMask &= ProcMaskRanged
		rangedSpeed = character.WeaponFromRanged().SwingSpeed
	}

	dpm := DynamicProcManager{procMasks: make([]ProcMask, 0, 2), procChances: []DynamicProc{}}

	chances := make([]staticProc, 0, 2)
	mergeOrAppend := func(speed float64, mask ProcMask) {
		if speed == 0 || mask == 0 {
			return
		}

		for i, proc := range chances {
			if proc.chance == speed {
				dpm.procMasks[i] |= mask
				return
			}
		}

		dpm.procMasks = append(dpm.procMasks, mask)
		chances = append(chances, staticProc{chance: speed})
	}

	mergeOrAppend(aa.mh.SwingSpeed, procMask&^ProcMaskRanged&^ProcMaskMeleeOH) // "everything else", even if not explicitly flagged MH
	mergeOrAppend(aa.offHandProcSpeed(), procMask&ProcMaskMeleeOH)
	mergeOrAppend(rangedSpeed, procMask&ProcMaskRanged)

	for i := range chances {
		if fixedProcChance != 0 {
			chances[i].chance = fixedProcChance
		} else {
			chances[i].chance *= ppm / 60
		}

		dpm.procChances = append(dpm.procChances, chances[i])
	}

	return dpm
}

// The rate on a spell hit's cast time, as the cast that landed it set it, 1.5 s at least. Read without
// a hit, the chance is an instant's.
type castTimeProc struct {
	ppm float64
}

func (cp castTimeProc) Reset()                       {}
func (cp castTimeProc) Chance(_ *Simulation) float64 { return cp.chanceFor(0) }
func (cp castTimeProc) ChanceOn(_ *Simulation, spell *Spell, _ *SpellResult) float64 {
	return cp.chanceFor(spell.CurCast.CastTime)
}
func (cp castTimeProc) Proc(sim *Simulation, label string) bool {
	return sim.Proc(cp.Chance(sim), label)
}
func (cp castTimeProc) chanceFor(castTime time.Duration) float64 {
	return cp.ppm * max(castTime, GCDDefault).Seconds() / 60
}

type staticProc struct {
	chance float64
}

func (sp staticProc) Reset()                       {}
func (sp staticProc) Chance(_ *Simulation) float64 { return sp.chance }
func (sp staticProc) Proc(sim *Simulation, label string) bool {
	return sim.Proc(sp.chance, label)
}
