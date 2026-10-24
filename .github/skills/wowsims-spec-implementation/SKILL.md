---
name: wowsims-spec-implementation
description: 'Use when picking up a WoWSims Forever "[Spec] Implement ..." tracking issue, porting a class or spec onto the spell data store and resolvers, or producing the talent and ability checklist for a spec that has no issue yet: deciding which game files content is actually live in Forever, working the checklist family by family, writing its SIML rotations and gear sets, keeping a long-running class branch merged with master, and proving the work with tests that can actually see it. Use it whenever the task names a class or spec to implement or port, even without the word "issue". Single spell, proc or consumable fixes outside a port are wowsims-spells.'
argument-hint: 'Name the spec and the issue number to work from.'
---

# Implementing a spec for Forever

Text written here follows `wowsims-conventions`. The data API, the docs map and the generators are in
`wowsims-spells`; read it first. This skill is the part specific to a whole class: what belongs on the
checklist, how to work it, and how to know it is done.

## Where the checklist comes from

The checklists are datamined, not hand-written. The extract lives outside this repository (it is
tooling the sim never reads), so ask for its location rather than guessing a path. It writes one JSON per
class with `abilities`, `families`, `talents` and a `meta` block naming the game build and trait tree.

A family is an ability and all its ranks, keyed by its lowest rank's spell id, and the checklist has one
box per family: Heroic Strike has nine ranks and one box.

## Deciding what is actually live

The Forever build carries Season of Discovery content that is not live, and it is the biggest source of
wrong checklist entries: the spells are correctly extracted from the game files and simply not part of
the game being simulated. Compare the class against the SoD build by spell id (in both, SoD only,
Forever only) and drop the SoD-only ones. Entries with no name and no description are the first to
check; on the warrior every one of them was SoD-only, together with a whole Runes tree of
`S03 - Tuning and Overrides Passive` entries.

`gen_spelldata` already leaves out the rune abilities an `Engrave <slot> - <rune>` spell grants and
nothing else teaches, and the `// Not generated:` comment at the head of a class's
`spell_data_auto_gen.go` names them (`docs/spell_data/regenerating.md`). The issue's extract does not
filter them, so the comparison is still yours.

## Traps in the checklist

- **A new low rank shifts every rank above it.** Forever's Slam gains a rank 1 (1240193), so every older
  Slam id sits one rank lower while keeping its id. Match on spell id, never on rank number.
- **Descriptions carry in-game colour codes.** Strip `|CFFFFFFFF` and `|R` and collapse whitespace, or
  the issue body renders as noise.
- **Name and id answer different questions.** A talent missing from another build's ids may exist there
  under another id; check the other build's data before calling something new, rather than inferring it
  from what the sim registers.
- **Counts in an issue body go stale.** Treat them as a starting point, and say what was filtered and
  why, so a re-run with different numbers is explainable.

## Working the checklist

Start from `sim/<class>/spell_data_auto_gen.go` and take every number the game files state from it
instead of a literal; a literal that survives carries its reason on the line. Whether the class reads
rows by hand or through the resolvers decides the idiom (see `wowsims-spells`); moving a class onto the
resolvers follows `docs/spell_data/porting.md`, whose key steps are dumping the rows before touching a
call site and isolating every .results move by reverting one change. A talent naming a weapon, and the
one hit, crit and haste rating, are in `docs/spell_data/weapon-requirements.md`, including the
weapon-bound talents of the classes still to port.

Where the sim and the game files disagree, decide rather than assume ("When a sim number disagrees" in
`wowsims-spells`), and record the decision at the call site.

## Rotations, gear sets and builds

A spec's presets are files under `ui/specs/<class>/<spec>/`. The hunter writes them in SIML, which
`sim/siml/SPEC.md` describes, and the other specs keep JSON until their developers move them; `core.GetAplRotation`
and `core.GetGearSet` read an `.apl.siml` or `.gear.siml` when one is there and the JSON otherwise. To move a spec,
print each JSON file with `go run --tags=with_db ./cmd/wowsimcli siml print -kind apl|gear|siml -spec <field>`,
check it with `siml verify`, and write the spec's `defaults.siml` as the hunter's is written.

- `apls/<name>.apl.siml` holds a rotation. It opens with `siml 1` and `spec <field>`, the spec's field in the `spec`
  oneof of `Player` in `proto/api.proto`: `hunter`, `dps_warrior`, `enhancement_shaman`.
- `gear_sets/<name>.gear.siml` holds one slot row per item: `main_hand arcanite_reaper + weapon_crusader`.
- `builds/<name>.siml` holds a whole setup, or a preset build the
  site offers, which states only what it changes over `defaults.siml`. A `use` line takes a gear set or a rotation
  from its own file, by a path from the build's folder: `use "../gear_sets/p1.gear.siml"`.
- Rotations that share variables and defs `use` one file that holds only those, as the hunter's weave rotations use
  `apls/weave_common.apl.siml` (SPEC.md §9): its defs and variables read in each rotation, which adds its own and
  sets a shared variable's value with `use … with`.
- `defaults.siml` holds the setup the spec's page starts with, a whole setup in the same form, with what the page
  sets on top of it stated: the race, `reaction_time_ms 100`, the rotation type (`apl: type auto` when the spec has
  no default rotation), for a tank `in_front_of_target` and `tanks player1`, and for a healer `target_dummies 9`.
  `spec.ts` reads it through `PresetUtils.specDefaultsFromSettings`, and `wowsimcli siml sim` sims a rotation on it,
  with the preset build that uses the rotation loaded over it; `wowsimcli sim --infile` loads a preset build over it
  the same way.

Every spell, aura, item and enchant is written by its name in the name table, `sim/siml/names.json`:
`cast mortal_strike`, `target.aura(sunder_armor).num_stacks`, `cast item.goblin_sapper_charge`. A spell or aura has
a name once the sim registers it, so register the spells first, then run `make siml-names` and commit
the table with them. A `qualified:` line in its report needs a repin; `wowsims-spells`, "Names in SIML files", has
how the table, repins and the repin record work. The site's Rotation tab writes a rotation with
**Export SIML**. `make vscode-apl` builds the editor extension, which completes the names the file's spec registers
(`tools/vscode-apl/README.md`). A name the table does not hold fails, with the names close to it:

```
$ go run --tags=with_db ./cmd/wowsimcli siml parse x.apl.siml
x.apl.siml:5:8 unknown-name no spell is named mortal_strik; did you mean mortal_strike?
x.apl.siml does not load
```

`presets.ts` imports each file through `tools/vite/siml.mts` as the JSON of its message: an `APLRotation`, an
`EquipmentSpec`, or an `IndividualSimSettings` for a build.

```ts
import DefaultApl from './apls/default.apl.siml';
import TestingBuildJson from './builds/testing.siml';
import PreBiSGearJson from './gear_sets/prebis.gear.siml';

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);
export const PreBiSGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSGearJson);
export const TestingBuild = PresetUtils.makePresetBuildFromJSON('Testing', Spec.SpecEnhancementShaman, TestingBuildJson);
```

The Go suite reads the same files, for example `core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "arms")`
and `core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p1_arms")`.

A preset rotation is written against `docs/rotation.md`: the pick after a hardcast does not see that
cast's outcome, so a line that reads a proc, a stack or a buff of another id the landing changes carries
`not(Spell Is Casting(X))`, while a line refreshing the cast's own dot does not. Check the preset's cast
counts in a debug log before taking its DPS.

## Keeping the branch merged

A port lives on one branch for a while, and master moves under it.

- **Generated spell data conflicts:** take master's `sim/core/spelldata/spells_auto_gen.go` and
  `assets/db_inputs/spell_store_inputs.json`, then rerun `gen_spelldata -classes=<class>`
  (`docs/spell_data/partial-updates.md`). Merging the generated text by hand produces a store no input
  describes.
- **Proto field numbers and `db.json`:** master keeps its numbers and the branch renumbers, and a local
  database behind master is spliced rather than regenerated; both are in "Consumables and items" and
  "Regenerating" in `wowsims-spells`.

## Verifying

```
go test --tags=with_db ./sim/<class>/... -count=1
git status --porcelain -- '*.results'      # empty unless a number was meant to move
make spelldata-check                       # the generated files match the database (needs wowsims.db)
go run --tags=with_db ./cmd/wowsimcli siml fmt -check ui/specs/<class>   # each file reads and is in canonical form; -w writes it
```

`siml fmt -check` fails on a `.siml` file under `ui/specs` that does not read cleanly, is not in canonical form, sits
outside its class folder, or names a spell or aura its spec does not register; `go test --tags=with_db
./cmd/wowsimcli/...` runs it over every file under `ui/specs`. `./sim/...` includes `sim/siml/gen`, whose
`TestTableKeepsItsEntries` fails when a committed entry of the name table comes back changed and logs the names it
lacks; `make siml-names` adds them.

**A .results file that cannot move proves nothing.** Before reading a pass as evidence, grep the class's
.results for the spell id you changed: a suite whose rotation never casts the spell passes whatever the
spell does. Where the suite cannot see a change, dump each constructed row against the literal it
replaced, field by field; that is the check that finds the bugs.

Read the test summary line, not the exit code, and ignore the presence of `.results.tmp` files, which
every run writes.

**Before committing, run the `wowsims-review` pass over the diff**, ideally from a fresh agent: every
number and hand-built config against its row and tooltip, the rules the rows cannot show, and every
finding verified before it is fixed. The PR body carries its "checked, not findings" and "hand numbers
that stay" lists.
