---
name: wowsims-spec-implementation
description: 'Use when picking up a WoWSims Forever "[Spec] Implement ..." tracking issue, porting a class or spec onto the spell data store and resolvers, or producing the talent and ability checklist for a spec that has no issue yet: deciding which game files content is actually live in Forever, working the checklist family by family, keeping a long-running class branch merged with master, and proving the work with tests that can actually see it. Use it whenever the task names a class or spec to implement or port, even without the word "issue". Single spell, proc or consumable fixes outside a port are wowsims-spells.'
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
```

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
