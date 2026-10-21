---
name: wowsims-spells
description: 'Use for any WoWSims Forever work on game files data or the generators that write it: registering or fixing a spell, talent, aura, dot or proc from its row; reading the sim/core/spelldata store by hand or through the resolvers; consumables (db.json entries, mob type and zone conditions, potion classes, registrations); item and enchant effects; overrides; regenerating with gen_spelldata or gen_db; and deciding whether a sim number or the game files are wrong. Use it even when the task only says "fix this potion", "this talent is off" or "regenerate the DB". Picking up a [Spec] issue or porting a whole class is wowsims-spec-implementation, which builds on this skill.'
argument-hint: 'Describe the spell, talent, proc, consumable, item effect or generator task.'
---

# WoWSims Forever game files data

Text written here follows `wowsims-conventions`. This skill is the map; the detail lives in
`docs/spell_data/` (start at its README), `docs/consumables.md` and `tools/spelldata/README.md`. Read the
page for the shape you are building before writing code: the pages carry the traps found the hard way, and
the snippets in them are pinned by `sim/core/spelldata/example_test.go`.

## Where things live

| Path                                                 | What it is                                                                                                                                       |
| ---------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `sim/core/spelldata/spells_auto_gen.go`              | Generated store: every spell the sim can reach, in the game files' units. `sim/core` must not import `spelldata` (the store imports core).       |
| `sim/core/spelldata/*.go`                            | Hand-written accessors (`store.go`, `spell.go`, `effect.go`, `ladder.go`, `attributes.go`) and resolvers (`resolve_*.go`, `parse_effects*.go`).  |
| `sim/<class>/spell_data_auto_gen.go`                 | Generated: one `spelldata.Ladder` per family, no numbers of its own.                                                                             |
| `tools/database/overrides/spell_overrides.go`        | Numbers the game files do not state, each with a reason and a rule that refuses it once the game files catch up. Read it before quoting a value. |
| `tools/database/overrides.go`                        | Item, enchant and consumable overrides, allow and deny lists.                                                                                    |
| `assets/db_inputs/spell_store_inputs.json`           | The rows the store was built from, gzipped despite the name (`zcat`). Lets the store be regenerated and tested with no game files database.      |
| `assets/database/db.json`, `db.bin`, `leftover_db.*` | The item, enchant and consumable database `gen_db` writes; `db.bin` is the same `proto.UIDatabase` encoded.                                      |
| `sim/common/forever/*_auto_gen.go`                   | Generated item, enchant and consumable registrations; hand-written effects sit beside them (`items_*.go`, `consumables_*.go`).                   |
| `sim/common/shared`                                  | The builders that run consumables and item effects from their rows.                                                                              |

The warrior, hunter, mage, paladin and shaman build on the resolvers; the druid, priest, rogue and warlock
read rows by hand, one value at a time. Grep `spelldata.SpellConfig` / `spelldata.Parse` in the class
folder to see which way the code you touch goes, and match it.

## Reading a row and building from it

```go
var slamRank = spellData.Slam.Highest()          // a ladder: Rank(n), Highest(), ByID(id), Len(), Contains(spell)
var slamDamage = slamRank.DamageEffect().Average()
rank := spellData.Defiance.Rank(points)          // rank 0 is untaken and answers Nil, so no `if rank > 0`
spelldata.Find(id)                               // Nil where the store lacks it; chains safely
spelldata.MustFind(id)                           // panics instead, for a spell the caller depends on
```

The accessor tables are in `docs/spell_data/store.md`; reading values by hand, with worked examples, in
`reading-by-hand.md`. The resolvers (`SpellConfig`, `AuraConfig`, `DotConfig`, `ParseStatic`,
`ParseEffects`, `ProcTrigger`) each fill what the row states and nothing else; the caller adds
`ApplyEffects`, the proc mask, conditions and any flag the game files lack. Their options are the
exported funcs returning `SpellOpt`, `AuraOpt`, `ParseOpt` and `ProcOpt` in those files. Per shape:

- An ability, and where it may be cast: `building-an-ability.md`. `SpellConfig` fills `CastRequirement`
  (stance, form, weapon, shield, creature type, out of combat) and core enforces it, so a stance or weapon
  check in `ExtraCastCondition` duplicates the row and drifts from it.
- An aura or dot: `auras-and-dots.md`. A talent or buff's modifiers: `talents-and-auras.md`, including
  `FromCaster` for a debuff that raises only its caster's attacks and `ShowAsSchoolHit`.
  `SPELLDATA_REPORT=1` prints what a parse skipped, which is what still needs wiring by hand.
- A proc: `procs.md` (proc chance talents, `ProcChanceSource`, trigger names, item and enchant procs).
- A weapon-bound hit or crit talent: `weapon-requirements.md`.
- Raid buffs and debuffs: `buffs-and-debuffs.md`. Overrides: `overrides.md`.

Ask whether a spell is one ability with `spellData.X.Contains(spell)`, never by class flags: the game
files share mask bits between abilities (Mortal Strike and Bloodthirst), so a flag match hears both.

## Pitfalls the docs do not cover

`docs/spell_data/pitfalls.md` holds the rest (rage in tenths, `EffectN` vs `Effect.Index`, ranks the game
never grants, `Average` rounding, `ProcChance` 100, a parsed row doubling a hand line).

- **A periodic listener built from a row never hears a bleed.** `DecodeProcTypeMask` files
  `PROC_FLAG_DEAL_HARMFUL_PERIODIC` under `ProcMaskSpellDamage`, which a bleed tick does not carry.
- **A sub-second GCD is named twice.** `core.Cast.GCDMin` overrides the one-second floor, so a spell
  with a shorter GCD sets both `GCD` and `GCDMin`, or the cast clamps it back up.
- **Pick a talent's effect by what the call site does, never by which number fits.** A talent can state
  the same 1/2/3 ladder on two effects (a stat bonus and a crit reduction); the mod's `Kind` decides.
- **Migrating a per-point literal buys provenance, rarely accuracy.** Most percent talents scale
  linearly; the few that do not are why the rows are worth reading, so read the ladder rather than
  multiplying rank 1.
- **The `SPELLMOD_*` names are read off the talents that use them**, because the game files ship no list.
  A value no talent names means the data changed, not that a name is missing.
- **SQL array columns are 0-based, `EffectN` is 1-based.** `EffectMiscValue_0` in `wowsims.db` is the
  first effect's misc value; `EffectN(1)` is the first effect in the store.
- **core's `SpellSchool` bits are the game files'**, so a school mask is carried, not translated.
  `stats.SchoolIndex` (array positions) and `proto.SpellSchool` are separate enums.
- **A new parse row ships consumables.** `Effect.ParsedAura()` turns true for every misc value of the
  aura, and `gen_db` reads that as supported, so the next regeneration can add imbues and stat auras.
  Diff the generated consumables and check each new one against its row.

## Consumables and items

`docs/consumables.md` is the pipeline: catalogue, routing, what ships, exclusive groups, hand-written
effects, what the settings offer and the browser path. `gen_db -gen=db` runs, in order: the `.sql`
overrides, routing and `ConsumableOverrides`, `AddUnparsedStats`, `AddPotionClasses`, item and enchant
overrides, the filters, then the effect and registration generators.

- **Conditions are data, not code.** What an on-use grants only against some creature types is
  `Consumable.mob_type_stats` (`AddUnparsedStats` in `tools/database/consumable_unparsed_stats.go`, from
  `spelldata.VersusStat` and `spelldata.MobTypes(e.Misc)`), and what it grants only in some zones is
  `Consumable.zone_stats` (the Natural flasks, against the hand-kept `naturalFlaskZones` there). Neither
  goes into `stats`: the settings offer the consumable by them and weigh them in its EP order only against
  a matching target or zone (`orderedByEp` in `ui/sim/settings/consumables.ts`). `AddPotionClasses`
  runs after it and counts the mob type stats when classing a potion, so a versus-only potion is still
  offensive; a potion of no class fails `gen_db`.
- **Proto scalars are plain fields; 0 means none.** `ScalingItemProperties.unrounded_dps` is 0 for an
  item with no weapon table row and is read into `core.Item.UnroundedDPS` as it stands.
- **Only `with_db` builds load every item** (`sim/core/database_load.go`). The dev server and the WASM
  know only the items in the request's `SimDatabase`, so `core.GetItemByID` answers nil there for an item
  the player does not wear, while every test passes. Code that reads such an item (a party buff's item,
  `sim/core/buffs/drivers.go`) needs the UI to ship it: `PARTY_BUFF_ITEM_IDS` in
  `ui/sim/proto/buff_items.ts`.
- **On a merge, master keeps its proto field numbers** and the branch renumbers its own new fields.
  Nothing converts saved data (see `ProtoVersion` in `proto/common.proto`), so renumbering a branch field
  costs nothing, while renumbering master's breaks every other branch built on it.

## Regenerating

`docs/spell_data/regenerating.md` and `partial-updates.md` cover the generators, the checks and taking a
new game build one class at a time (`make db CLASSES=<class>`). Regenerating needs `tools/database/wowsims.db`,
gitignored and built by `make db` from a local WoW install; the store and `tools/database` tests run from
the committed inputs without it.

- **`gen_db` executes every numbered `tools/database/overrides/*.sql` against the database it opens.**
  Give a worktree its own copy of `wowsims.db`, never one shared with another checkout, so a run in one
  cannot change the rows another reads.
- **A local database behind master cannot regenerate items.** `RequireCommittedHotfixes` refuses a
  database older than the committed inputs' hotfixes, and regenerating from it would drop items only a
  hotfix adds. Splice only the intended entries into master's `db.json` and `db.bin` with a throwaway Go
  program that writes both through `protojson` / `proto.Marshal`: the sim embeds `db.bin`, the UI fetches
  either, and nothing derives one from the other, so a hand edit to one leaves the other behind. Then diff
  the JSON and `protoc -I proto --decode=proto.UIDatabase proto/ui.proto < assets/database/db.bin` against
  master's; the binary is not byte-stable, so only the decoded diff shows what moved.
- **Never delete a generated file before regenerating it.** `gen_db` imports the sim, so a class package
  missing its file breaks the generator that would recreate it.

```
go run ./tools/spelldata 11574                     # one row; -family warrior/Execute, -expr, -config, -hover, -lsp: see tools/spelldata/README.md
go run ./tools/database/gen_spelldata              # rewrite the store, the forms and every class file
go run ./tools/database/gen_spelldata -check       # name what is stale, write nothing (make spelldata-check)
go run tools/database/gen_db/*.go -outDir=./assets -gen=db   # db.json, db.bin, registrations, talent files
go test ./sim/core/spelldata/ ./tools/database/ -count=1     # no database needed
go test --tags=with_db ./sim/<class>/ -count=1     # without the tag: "No DB data for enchant"
git status --porcelain -- '*.results'              # empty unless a number was meant to move
```

`make update-tests` deletes every `.results` before promoting `.tmp`, so a test that cannot run locally
loses its file: check for a `D` afterwards, or `cp` the one `.tmp` you mean. An `rtk`-wrapped `go test`
can exit 0 on a failing run, so read the summary line. Before committing a spell change, run the
`wowsims-review` pass over the diff.

## When a sim number disagrees with the game files

The rows are a second opinion on every number the sim already had. Decide rather than assume, and when a
disagreement stays, say so at the call site.

- The sim is likely wrong when its value equals a different rank's or a sibling spell's, or breaks the
  die-sides ladder of the other ranks.
- The sim is likely right when the offset is uniform across unrelated abilities: that usually means the
  sim models something the game files' base excludes (a ranged-weapon component in every hunter cast
  time), not a run of slips. A coefficient read from another spell is not a disagreement either: a tick's
  coefficient belongs to the spell its description names.
- Where the row and the tooltip disagree, the tooltip can win (Flurry's haste ladder against its buff
  row's flat 30), and the comment names the wording it came from.
