package core

import (
	"slices"
	"testing"
)

// Masks taken from the game files rows named in each case.
func TestDecodeProcTypeMask(t *testing.T) {
	cases := []struct {
		name               string
		mask               [2]uint32
		hint               ProcHint
		callback           AuraCallback
		procMask           ProcMask
		outcome            HitOutcome
		requireDamageDealt bool
		unsupported        []string

		requireDamageComponent bool
	}{
		{
			name:               "Lightning Shield, every direct hit taken",
			mask:               [2]uint32{0x222A8, 0},
			callback:           CallbackOnSpellHitTaken,
			procMask:           ProcMaskMeleeWhiteHit | ProcMaskMeleeSpecial | ProcMaskRangedAuto | ProcMaskRangedSpecial | ProcMaskSpellDamage,
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:     "Striking, the harmful-spell bit alone",
			mask:     [2]uint32{0x10000, 0},
			callback: CallbackOnCastComplete,
			procMask: ProcMaskSpellDamage,
			outcome:  OutcomeEmpty,
		},
		{
			name:     "the damage-class-none bit beside the harmful-spell one, the same cast either way",
			mask:     [2]uint32{0x11000, 0},
			callback: CallbackOnCastComplete,
			procMask: ProcMaskSpellDamage,
			outcome:  OutcomeEmpty,
		},
		{
			name:               "the harmful-ability bit taken, which names no hits of its own",
			mask:               [2]uint32{0x2000, 0},
			callback:           CallbackOnSpellHitTaken,
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:     "the harmful-spell bit with a crit tooltip",
			mask:     [2]uint32{0x10000, 0},
			hint:     ProcHintCrit,
			callback: CallbackOnSpellHitDealt,
			procMask: ProcMaskSpellDamage,
			outcome:  OutcomeCrit,
		},
		{
			name:               "Dual Wield Specialization, melee swings restricted to the off hand",
			mask:               [2]uint32{0x800004, 0},
			callback:           CallbackOnSpellHitDealt,
			procMask:           ProcMaskMeleeOHAuto,
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:               "Bloodthrill, melee swings and abilities from the main hand alone",
			mask:               [2]uint32{0x14, 0x20},
			callback:           CallbackOnSpellHitDealt,
			procMask:           ProcMaskMeleeMH,
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:               "the unknown top bit",
			mask:               [2]uint32{0x80000000, 0},
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
			unsupported:        []string{"bit 31"},
		},
		{
			name:     "Darkmoon Card: Blue Dragon, a cast tooltip on both spell bits",
			mask:     [2]uint32{0x14000, 0},
			hint:     ProcHintCastTrigger,
			callback: CallbackOnCastComplete,
			procMask: ProcMaskSpellDamage | ProcMaskSpellHealing,
			outcome:  OutcomeEmpty,
		},
		{
			name:     "both spell bits with a heal tooltip",
			mask:     [2]uint32{0x14000, 0},
			hint:     ProcHintHeals,
			callback: CallbackOnSpellHitDealt | CallbackOnHealDealt,
			procMask: ProcMaskSpellDamage | ProcMaskSpellHealing,
			outcome:  OutcomeLanded,
		},
		{
			name:     "Grand Arcanist 1231152, both spell bits on a row with no tooltip",
			mask:     [2]uint32{0x14000, 0},
			callback: CallbackOnSpellHitDealt | CallbackOnHealDealt,
			procMask: ProcMaskSpellDamage | ProcMaskSpellHealing,
			outcome:  OutcomeLanded,
		},
		{
			name:     "Clearcasting 12536, every dealt hit and both spell bits with no heal tooltip",
			mask:     [2]uint32{0x15550, 0},
			callback: CallbackOnSpellHitDealt | CallbackOnHealDealt,
			procMask: ProcMaskMeleeSpecial | ProcMaskRangedAuto | ProcMaskRangedSpecial | ProcMaskSpellDamage | ProcMaskSpellHealing,
			outcome:  OutcomeLanded,
		},
		{
			name:     "the helpful bits with a healing-spells tooltip",
			mask:     [2]uint32{0x204000, 0},
			hint:     ProcHintHeals | ProcHintPureHeal,
			callback: CallbackOnHealDealt | CallbackOnPeriodicHealDealt,
			procMask: ProcMaskSpellHealing,
			outcome:  OutcomeLanded,
		},
		{
			name:     "the helpful-spell bit alone with a heal tooltip",
			mask:     [2]uint32{0x4000, 0},
			hint:     ProcHintHeals,
			callback: CallbackOnHealDealt,
			procMask: ProcMaskSpellHealing,
			outcome:  OutcomeLanded,
		},
		{
			name:               "Holy Power 28789, the helpful-spell bit alone with no heal tooltip",
			mask:               [2]uint32{0x4000, 0},
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:               "periodic damage taken",
			mask:               [2]uint32{0x80000, 0},
			callback:           CallbackOnPeriodicDamageTaken,
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:               "any damage taken",
			mask:               [2]uint32{0x100000, 0},
			callback:           CallbackOnSpellHitTaken | CallbackOnPeriodicDamageTaken,
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:               "a named ability, the hint the decoder leaves to the caller",
			mask:               [2]uint32{0x222A8, 0},
			hint:               ProcHintNamedAbility,
			callback:           CallbackOnSpellHitTaken,
			procMask:           ProcMaskMeleeWhiteHit | ProcMaskMeleeSpecial | ProcMaskRangedAuto | ProcMaskRangedSpecial | ProcMaskSpellDamage,
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:     "Battlegear of Wrath 23548, whose trigger is a block",
			mask:     [2]uint32{0x2A8, 0},
			hint:     ProcHintOutcomeTaken,
			callback: CallbackOnSpellHitTaken,
			procMask: ProcMaskMeleeWhiteHit | ProcMaskMeleeSpecial | ProcMaskRangedAuto | ProcMaskRangedSpecial,
			outcome:  OutcomeLanded,
		},
		{
			name:     "Recovery 1248761, whose trigger is the wearer's melee attack dodged or parried",
			mask:     [2]uint32{0x14, 0},
			hint:     ProcHintAttackDodged | ProcHintAttackParried,
			callback: CallbackOnSpellHitDealt,
			procMask: ProcMaskMeleeWhiteHit | ProcMaskMeleeSpecial,
			outcome:  OutcomeDodge | OutcomeParry,
		},
		{
			name:     "the wearer's attack parried alone",
			mask:     [2]uint32{0x14, 0},
			hint:     ProcHintAttackParried,
			callback: CallbackOnSpellHitDealt,
			procMask: ProcMaskMeleeWhiteHit | ProcMaskMeleeSpecial,
			outcome:  OutcomeParry,
		},
		{
			name:     "Vindication, melee hits restricted to damaging ones",
			mask:     [2]uint32{0x14, 0},
			hint:     ProcHintDamage,
			callback: CallbackOnSpellHitDealt,
			procMask: ProcMaskMeleeWhiteHit | ProcMaskMeleeSpecial,
			outcome:  OutcomeLanded,

			requireDamageComponent: true,
		},
		{
			name:     "Talisman of Ascendance, damage or healing spells",
			mask:     [2]uint32{0x15550, 0},
			hint:     ProcHintHeals | ProcHintDamage,
			callback: CallbackOnSpellHitDealt | CallbackOnHealDealt,
			procMask: ProcMaskMeleeSpecial | ProcMaskRangedAuto | ProcMaskRangedSpecial | ProcMaskSpellDamage | ProcMaskSpellHealing,
			outcome:  OutcomeLanded,

			requireDamageComponent: true,
		},
		{
			name:     "the harmful-spell bit alone with damage wording, still a cast",
			mask:     [2]uint32{0x10000, 0},
			hint:     ProcHintDamage,
			callback: CallbackOnCastComplete,
			procMask: ProcMaskSpellDamage,
			outcome:  OutcomeEmpty,
		},
		{
			name:               "Enrage, damage wording beside a taken mask",
			mask:               [2]uint32{0x222A8, 0},
			hint:               ProcHintDamage,
			callback:           CallbackOnSpellHitTaken,
			procMask:           ProcMaskMeleeWhiteHit | ProcMaskMeleeSpecial | ProcMaskRangedAuto | ProcMaskRangedSpecial | ProcMaskSpellDamage,
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:               "a word-1 bit",
			mask:               [2]uint32{0, 0x4},
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
			unsupported:        []string{"bit 34"},
		},
		{
			name:               "Spinal Reaper 21185, a killing blow",
			mask:               [2]uint32{0x2, 0},
			callback:           CallbackOnKill,
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:               "Lunar Idol 419288, a killing blow and a target dying",
			mask:               [2]uint32{0x2, 0x1},
			callback:           CallbackOnKill | CallbackOnTargetDies,
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
		},
		{
			name:               "the caster dying",
			mask:               [2]uint32{0x1000000, 0},
			outcome:            OutcomeLanded,
			requireDamageDealt: true,
			unsupported:        []string{"DEATH"},
		},
	}

	for _, tc := range cases {
		info := DecodeProcTypeMask(tc.mask, tc.hint)

		if info.Callback != tc.callback {
			t.Errorf("%s: Callback = %v, want %v", tc.name, info.Callback, tc.callback)
		}
		if info.ProcMask != tc.procMask {
			t.Errorf("%s: ProcMask = %v, want %v", tc.name, info.ProcMask, tc.procMask)
		}
		if info.Outcome != tc.outcome {
			t.Errorf("%s: Outcome = %d, want %d", tc.name, info.Outcome, tc.outcome)
		}
		if info.RequireDamageDealt != tc.requireDamageDealt {
			t.Errorf("%s: RequireDamageDealt = %v, want %v", tc.name, info.RequireDamageDealt, tc.requireDamageDealt)
		}
		if info.RequireDamageComponent != tc.requireDamageComponent {
			t.Errorf("%s: RequireDamageComponent = %v, want %v", tc.name, info.RequireDamageComponent, tc.requireDamageComponent)
		}
		if !slices.Equal(info.Unsupported, tc.unsupported) {
			t.Errorf("%s: Unsupported = %v, want %v", tc.name, info.Unsupported, tc.unsupported)
		}
	}
}

// A trigger built the way spelldata.ProcTrigger builds one from a row: the decoded listener, a stated
// chance of 1. The kill bit fires on a credited death alone, the target-dies bit on every death, and a
// mask carrying both fires once for each.
func TestDecodedDeathFlagsFireOnDie(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		mask                 [2]uint32
		uncredited, credited int
	}{
		{name: "the kill bit", mask: [2]uint32{0x2, 0}, uncredited: 0, credited: 1},
		{name: "the target-dies bit", mask: [2]uint32{0, 0x1}, uncredited: 1, credited: 1},
		{name: "both bits", mask: [2]uint32{0x2, 0x1}, uncredited: 1, credited: 2},
	} {
		fired := 0
		var trigger *Aura
		sim, target := setupTargetDeathSim(func(char *Character) {
			decoded := DecodeProcTypeMask(tc.mask, 0)
			trigger = char.MakeProcTriggerAura(ProcTrigger{
				Name:               "Decoded Death Trigger",
				Callback:           decoded.Callback,
				ProcMask:           decoded.ProcMask,
				Outcome:            decoded.Outcome,
				RequireDamageDealt: decoded.RequireDamageDealt,
				ProcChance:         1,
				TriggerImmediately: true,
				Handler: func(_ *Simulation, _ *Spell, _ *SpellResult) {
					fired++
				},
			})
		})
		trigger.Activate(sim)

		target.Die(sim, 0)
		if fired != tc.uncredited {
			t.Errorf("%s: fired %d times on a death with no kill credit, want %d", tc.name, fired, tc.uncredited)
		}

		fired = 0
		target.Enable(sim)
		target.Die(sim, 1)
		if fired != tc.credited {
			t.Errorf("%s: fired %d times on a credited kill, want %d", tc.name, fired, tc.credited)
		}
	}
}
