---
name: wowsims-apl
description: 'Use when a WoWSims Forever rotation (APL) misbehaves or changes: a preset line that never fires or fires every time it is ready, a cast skipped with no error, "does not know spell" or "No aura found" validations, a new or changed APL value or action in sim/core/apl_values_*.go or apl_actions_*.go and getting it into the rotation editor''s kind picker, a preset under ui/specs/<class>/<spec>/apls/ or its option groups, what a value reads at a hardcast end (Spell Is Casting, reaction time, stacks), Strict Multidot, Multiple CD Usages, Spell Expected DPET, a major cooldown waiting on the GCD, or hand-writing .apl.json with the JSON schema. Use it before trusting a class suite''s numbers after a preset edit. The order of a cast''s effects and procs is wowsims-core-mechanics; the rotation editor''s components and store are wowsims-ui; proving cast counts at runtime is wowsims-testing; judging a preset change is wowsims-review; reading a spell row is wowsims-spell-store; the map is wowsims-spells.'
---

# wowsims-apl

The rotation is an APL: a prepull list, a priority list and groups of actions with conditions, parsed from
`proto/apl.proto` by `sim/core/apl*.go` and edited in the UI or as a preset file. The rule this skill exists for: a
preset reference the sim does not resolve fails silently in the class suites, so a preset edit is trusted only after its
validations are dumped (§What fails silently). The rulings the code follows are in `references/rulings.md`; text here
follows `wowsims-conventions`.

## Where things live

| Path                                     | What it is                                                                          |
| ---------------------------------------- | ----------------------------------------------------------------------------------- |
| `sim/core/apl.go`                        | parsing, the validations per line, `getStats`, the cast-end pass                    |
| `sim/core/apl_action.go`                 | the action switch, condition resolution and pruning                                 |
| `sim/core/apl_value.go`                  | the value switch, `APLValue` and `DefaultAPLValueImpl`                              |
| `sim/core/apl_helpers.go`                | `GetAPLSpell`, `GetAPLAura`, `GetAPLDot`: id lookups and their warnings             |
| `sim/core/apl_values_spell.go`           | spell values; `apl_values_aura.go`, `apl_values_dot.go` and the rest by topic       |
| `sim/core/apl_actions_casting.go`        | Cast Spell, Multidot, Strict Multidot, Autocast Other Cooldowns, range warning      |
| `sim/core/apl_action_group_reference.go` | Group Reference; `apl_actions_sequences.go`, `apl_actions_misc.go` likewise         |
| `sim/warlock/apl_values.go`              | class values and actions; `sim/druid/apl_values.go`, `sim/shaman/apl_values.go`     |
| `ui/features/apl/model/value_kinds.ts`   | the kind picker's value entries; `action_kinds.ts`, `action_id_sets.ts`             |
| `ui/specs/`                              | presets, in `apls/` or `builds/` per spec (§Presets)                                |
| `sim/core/test_utils.go`                 | `GetAplRotation` and `GetTestBuildFromJSON`, which load a preset into a class suite |

| You are…                                                         | Open                                                    |
| ---------------------------------------------------------------- | ------------------------------------------------------- |
| writing a line that reads the running cast                       | `docs/rotation.md` §Which lines need `Spell Is Casting` |
| deciding what the pick at the cast end reads, or `Is Hit Capped` | `references/rulings.md`                                 |
| hand-editing an `.apl.json` with completion                      | `tools/APL.md`                                          |
| changing the rotation editor's components or store               | `wowsims-ui`                                            |

## Values and actions

A value or action reaches the rotation and the kind picker in five places, all in one change:

1. A message and a oneof field in `proto/apl.proto`; field numbers follow `wowsims-core-mechanics` §Protos.
2. A case in `newAPLValueWithContext` (`apl_value.go`) or `newAPLActionImpl` (`apl_action.go`). A class-only one goes in
   the class's `NewAPLValue` or `NewAPLAction`, which runs before core's switch.
3. The constructor in the topic file. It resolves ids through `rot.GetAPLSpell` or `rot.GetAPLAura`, which warn on a
   miss, and returns nil when they find nothing. A value embeds `DefaultAPLValueImpl` and implements the getter for its
   `Type()`; the other getters panic. Core's switch sets `Uuid` by reflection; a class `NewAPLValue` returns before it,
   so a class value passes `config.Uuid` along itself (`sim/shaman/apl_values.go`).
4. The entry in `valueKinds` or `actionKinds`: label, submenu, `includeIf` (prepull, spec) and fields. The mapped type
   fails `npm run type-check` without one.
5. The label and tooltip keys in `assets/locales/en/translation.json`, which `ui/i18n/locale_keys.test.ts` checks.

- **Nothing forces step 2.** The switch's `default` returns nil with no warning, so a oneof without a case is dropped
  like an unresolved value.
- **A field reads its value through `coerceTo`** (`apl_values_operators.go`), which leaves out a value it cannot read
  as the field's type, a yes/no value as a duration or text as a number, with a "Cannot read" warning. Move, Move
  Duration, Wait, Wait Until and a prepull Do At then drop the line with a warning of their own.
- **An action with a value field implements `GetAPLValues` and `ReplaceAPLValues`.** The first is how its values
  are finalized and scanned for placeholders, the second how a Group Reference fills the group's placeholders and
  variables in them (`apl_action_group_reference.go`); without it a placeholder there reads zero. A group variable
  that overrides a value variable is read as the value variable's type (`reResolveVariableRefs`).
- **`Spell Is Known`, `Aura Is Known` and `Talent Is Known` fold to a constant at parse.** A constant false condition
  prunes the line, a constant true one is no condition (`newAPLAction`).
- **The kind picker lists the spells flagged `SpellFlagAPL`** and every aura with an ActionID (`metadata` in
  `sim/core/unit.go`). A preset can still name an unflagged spell: `GetAPLSpell` falls back to any spell with the id.
- **A major cooldown a Cast Spell names leaves Autocast Other Cooldowns** (`removeFromMajorCooldowns`, `apl.go`),
  pruned lines included. Every other major cooldown, racials included, fires only through Autocast Other Cooldowns,
  whose Information validation lists them; a rotation without it never casts them.

## Presets

- **One file serves the UI and the suite, in one of two layouts.** `<spec>/apls/<name>.apl.json` is imported by the
  spec's `presets.ts` and loaded by `core.GetAplRotation` (`sim/warrior/dps/dps_test.go`); `<spec>/builds/<name>.build.json`
  holds it in `player.rotation`, passed to `makePresetAPLRotation` and loaded by `core.GetTestBuildFromJSON`
  (`sim/warlock/warlock_test.go`). Priest, elemental shaman and warlock use builds only, enhancement shaman both.
- **An option is a hidden group.** The priority list holds a Group Reference with `hide: true`, and the spec's
  `simpleRotation` finds it by group name and sets `hide` (`ui/specs/warrior/dps/spec.ts`). A group renamed in the JSON
  alone breaks its toggle without an error.
- **A suite runs a hidden group only when it un-hides it**, as `armsStanceDanceRotation` does in
  `sim/warrior/dps/dps_test.go`; a group no suite un-hides never reaches the `.results`.
- **A line is written against the pick at the cast end** (`docs/rotation.md`). After an edit, dump the validations
  (below) and judge the preset's cast counts from a debug log, not from the line as written.

## What fails silently

The class suites run a rotation whatever its validations say. Each of these runs a different rotation from the one the
preset names:

- **An unknown spell.** A Cast Spell, Multidot or value on a spell the unit lacks logs "does not know spell" and the line
  or value is dropped (`GetAPLSpell`). A `spellIsKnown` guard prunes the line, and the warning stays.
- **An aura under another ActionID.** An `auraIsActive` or `auraNumStacks` naming a spell id core registers under
  another ActionID logs "No aura found" and the value is dropped. An And drops that branch with a warning, an Or or a
  Not drops it silently, and a condition with nothing left reads true and the line runs every time it is ready,
  logging "Condition could not be resolved and will be treated as always true" (`sim/core/apl_action.go`); only the
  branch inside an Or or a Not is silent. An `auraIsKnown` on such an id folds to false and the line never runs.
- **A misspelled preset name in a suite.** `GetAplRotation` logs "no apl json file" and the suite runs no rotation
  (`GetTestBuildFromJSON` stops the run instead).

**Dump the validations** before trusting a preset: a throwaway `_test.go` in the class package builds the suite's player
with `core.SinglePlayerRaidProto` and the rotation from the suite's loader (`core.GetAplRotation`, or the `Rotation` of
`core.GetTestBuildFromJSON`'s config), calls `core.ComputeStats`, and prints every `Validations` entry of
`RotationStats`: `PrepullActions`, `PriorityList`, `Groups[i].Actions` and `UuidValidations`, as `bearAPLWarnings` in
`sim/druid/feralbear/apl_sweep_test.go` does. Run it as `wowsims-testing` §The gates says, then delete it.

## Timing

- **The pick at a hardcast end reads the running cast as landed**; what it sees and misses is `docs/rotation.md` §What
  the pick sees, and the order the landing runs in is `wowsims-core-mechanics` §Cast order.
- **A major cooldown's Cast Spell is ready only at a GCD-ready look** unless it is reactive or inside a sequence
  (`APLActionCastSpell.IsReady`), so an off-GCD cooldown can wait a GCD after its window opens; the first warrior major
  cooldown chance comes about 4.5 s into the fight (PR #653).
- **`Aura Is Inactive` with reaction time is the refresh check** for an aura the rotation keeps up:
  `not(Aura Is Active)` with reaction time reads a fresh landing as inactive and recasts (`APLValueAuraIsInactive`).
- **`Aura Num Stacks` with reaction time returns the previous stacks for a full reaction time after any change**, the
  player's own spend included, so a stack-gated line can re-cast right after its own cast; add an `Aura Is Active`
  check without reaction time (`APLValueAuraNumStacks`).
- **A prepull Do At is read once, when the rotation is built** (`newAPLRotation` in `sim/core/apl.go`), at the start
  distance and before any prepull buff, so a `Spell Travel Time` there is the travel from the start distance, even
  when an earlier prepull action moves the character. The hardcast itself rounds to the millisecond (`sim/core/cast.go`)
  and `Spell Cast Time` does not, so a Do At built from it can be up to half a millisecond off the cast end. A cast
  that comes that early is queued for the cast end (`QueueSpell` in `sim/core/spell_queueing.go`), so an action timed
  at a cast end still goes after the cast lands. A Do At that reads the encounter (`Current Time`, a unit's health)
  has no sim to read then: the line is left out with a warning that names the panic (`prepullTime`).

## Traps

- **An external buff copy registers as tag -1** (`actionID` in `sim/core/buffs/meta.go`): the raid's Battle Shout is
  25289 tag -1, the player's own copy tag 0. Berserking is 20554 untagged (`sim/core/racials.go`).
- **A talent gate reads `Talent Is Known`**, which holds for the rank or higher (`sim/core/apl_values_talent.go`). An
  `auraIsKnown` on a talent id finds an aura only where the class registers a `core.MakePermanent` marker aura tagged
  with the points.
- **Recklessness needs Berserker Stance** (ShapeshiftMask bit 18): Arms casts it only with the Stance Dance group on,
  and the dps suite never un-hides the Recklessness group, so the `.results` never show it (PR #653).

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The commands that
re-derive each claim, the "You changed → Update" table and the path check are in `references/maintenance.md`.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill or its `references/rulings.md` otherwise. A memory file is a pointer
to this skill, not a second copy.
