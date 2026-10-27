---
name: wowsims-spell-store
description: 'Use when Go code reads the WoWSims Forever spell store, sim/core/spelldata: picking a rank off a ladder (Rank, Highest, ByID, Len, Contains), Find or MustFind on an id, choosing between EffectN, EffectAt and Effect(aura, misc), reading Average, FractionAt or TenthsAt, following Trigger(), Triggered() or Refs() to the spell that does the work, asking whether a spell is one ability, telling whether a class file reads rows by hand or through the resolvers, or running go run ./tools/spelldata (an id, -family, -expr, -config) or the editor hover. Use it too when a value is one rank or one effect off, Highest() names the wrong spell, a misc value names the wrong stat, a ladder reader panics on a two-effect row, a ladder misses or doubles a rank, or a Wowhead number disagrees with the row. Building a spell, aura or dot from a row is wowsims-abilities-auras; a proc or trigger is wowsims-procs; regenerating the store or the database is wowsims-db-pipeline; the map of every spell skill is wowsims-spells.'
---

# wowsims-spell-store

The store holds every spell row the sim can reach, each field the game files' column in the game files' units, and
the accessors convert on the way out. The rule this skill exists for: name the rank and the effect you mean. A rank
comes from `Rank(n)`, `ByID(id)` or an edge off another row, an effect from its position or its aura, never from
whichever one happens to give the number you expected.

Text written here follows `wowsims-conventions`. The detail lives in the pages below; this skill routes to them and holds
the traps they do not carry.

## Where things live

| Path                                          | What it is                                                                                      |
| --------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| `sim/core/spelldata/spells_auto_gen.go`       | the generated store; `sim/core` must not import `spelldata`, since the store imports core       |
| `sim/core/spelldata/store.go`                 | `Find`, `MustFind`, `All`, `ByName`                                                             |
| `sim/core/spelldata/spell.go`                 | a row: `EffectN`, `Effect`, `FindEffect`, the role readers, `Triggered`, `Refs`, `ReagentCount` |
| `sim/core/spelldata/effect.go`                | an effect's value readers: `Average`, `Percent`, `Tenths`, `Roll`, `Trigger`                    |
| `sim/core/spelldata/ladder.go`                | `Ladder`, `Ranked`, `Talent` and the per-rank readers                                           |
| `sim/core/spelldata/attributes.go`            | the attribute bits the sim acts on                                                              |
| `sim/<class>/spell_data_auto_gen.go`          | a class's `spellData`: one ladder per family, no numbers of its own                             |
| `tools/database/gen_spell_data.go`            | `discoverLadders`, which decides which rows become a family's ranks                             |
| `tools/database/overrides/spell_overrides.go` | numbers the game files do not state                                                             |
| `tools/spelldata/`                            | the command-line reader and the language server behind the editor hover                         |

| You are…                                                                               | Open                                 |
| -------------------------------------------------------------------------------------- | ------------------------------------ |
| looking up an accessor, a unit, `Nil` chaining, class flags or a ladder reader         | `docs/spell_data/store.md`           |
| reading values into a hand-built config, a borrowed tick, or the inputs a dot captures | `docs/spell_data/reading-by-hand.md` |
| running the tool, reading its JSON, or setting up the hover in an editor               | `tools/spelldata/README.md`          |
| chasing a value that is off: rage in tenths, ranks the game never grants, the rounding | `docs/spell_data/pitfalls.md`        |
| building a ladder's ranks, reading a talent's curve, or checking a row against Wowhead | `references/row-traps.md`            |

## Reading a row

```go
var slamRank = spellData.Slam.Highest()          // a ladder: Rank(n), Highest(), ByID(id), Len(), Contains(spell)
var slamDamage = slamRank.DamageEffect().Average()
rank := spellData.Defiance.Rank(points)          // rank 0 is untaken and answers Nil, so no `if rank > 0`
spelldata.Find(id)                               // Nil where the store lacks it; chains safely
spelldata.MustFind(id)                           // panics instead, for a spell the caller depends on
```

- A ladder's `...At` readers panic on a rank with more than one effect; name the effect or let the parse read it
  (`docs/spell_data/pitfalls.md`).
- SQL array columns in `wowsims.db` are 0-based and `EffectN` is 1-based, so `EffectMiscValue_0` is `EffectN(1)`'s
  misc value.
- core's `SpellSchool` bits are the game files' bits, so a school mask is carried, not translated. `stats.SchoolIndex`
  (array positions) and `proto.SpellSchool` are separate enums.
- Ask whether a spell is one ability with `spellData.X.Contains(spell)`, on a `ProcTrigger` in its `ExtraCondition`,
  never with `ClassFlags`, which stay for talent and set filters (PR #93).

`docs/spell_data/store.md` has the rest: "Reaching an effect" for `EffectN`, `Effect(aura, misc)` and the role
readers, "Reading a value" for `Average` and the roll, "Attributes, edges and class flags" for `Trigger()`,
`Triggered()`, `Refs()` and why class flags hear several abilities. The per-level gain a `Ranked` ladder's readers skip
is in `docs/spell_data/reading-by-hand.md`.

## Hand reads and resolvers

The resolvers (`SpellConfig`, `AuraConfig`, `DotConfig`, `ParseStatic`, `ParseEffects`, `ProcTrigger`) fill what the
row states and nothing else. A hand read builds `core.SpellConfig` field by field from the row (PR #42). Which way a
file goes is read off the file, so match the file you touch:
`grep -l 'spelldata.SpellConfig\|spelldata.Parse' sim/<class>/*.go`.
Every class folder has files on the resolvers, and a resolver-built spell still reads some values by hand
(Consecration's ticks). The priest builds through the resolvers, not hand reads (PR #246).

## The spelldata tool

`tools/spelldata/README.md` has every mode, the JSON shape and the editor hover. Two modes answer the questions this
skill raises: `-expr` reads a ladder chain the way a class file writes it and marks the effect it read, and `-config`
states each field a resolver call fills and the step that filled it.

- The tool has no `-h`; it answers `unknown argument "-h"`. Run it with no arguments for the usage line.
- `go run ./tools/spelldata` needs the generated protobuf Go code, so a fresh checkout runs
  `make sim/core/proto/api.pb.go` first. The editor launcher does this itself.

## Traps

- **`Highest()` is the ladder's last entry, not its top rank**, and a database update that appends rows moves it; read
  a triggered row off the cast's top rank with `Trigger()` or `Refs()`. The ladder readers: `docs/spell_data/store.md`;
  the case: `.github/skills/wowsims-review/references/defects.md` §The wrong rank, the wrong effect.
- **A hand read of `Misc` gets the wrong stat** on an `A_MOD_TOTAL_STAT_PERCENTAGE` effect;
  `Effect.TotalStatPercentStats()` reads it. The case: Heart of the Wild in the same `defects.md` section.

The traps in the rows themselves (ladders keyed by name, granted copies, talent curves, summons, Wowhead's numbers
against the row) are in `references/row-traps.md`.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The commands that
re-derive each claim, the "You changed → Update" table and the path check are in `references/maintenance.md`.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill otherwise. A memory file is a pointer
to this skill, not a second copy.
