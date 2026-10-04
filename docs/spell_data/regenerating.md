# Regenerating and checking

```
go run ./tools/database/gen_spelldata              # rewrite every generated spell data file
go run ./tools/database/gen_spelldata -classes=druid  # take the druid from the database, keep every other class as committed
go run ./tools/database/gen_spelldata -check       # name the ones that are stale, write nothing
go run ./tools/database/gen_spelldata -unchecked   # write without the type-check below
make spelldata-check                               # the same check
```

The generator reads `tools/database/wowsims.db` and writes all of it in one pass:
`sim/core/spelldata/spells_auto_gen.go`, `sim/core/dbcenums/forms_auto_gen.go`, a
`sim/<class>/spell_data_auto_gen.go` of ladders into the store for every class, and the raid buffs,
which [Buffs and debuffs](buffs-and-debuffs.md) describes. It is a separate binary from `gen_db` on
purpose. `gen_db` imports the sim and the sim reads these files, so a stale one would stop the generator
that fixes it from compiling. For the same reason nothing is written until all of it type-checks. The
rendered bytes go to a staging directory first and are compiled through `go build -overlay` in the
place of the committed files. A failure leaves the tree exactly as it was and prints what the compiler
said. `-unchecked` skips that build. It is for a class file whose call sites cannot type-check against
what the file will state until they move: the file is written first and the call sites follow, and
`-check` closes the loop once the package builds again. `make db` runs the store before `gen_db` for
the same ordering reason, since `gen_db` classifies every item and enchant proc out of the store
compiled into it.

Nothing lists which spells to generate. The class files walk `dbc.Classes`, take each class's own skill
lines and emit a family for every spell whose subtext reads `Rank N`. The store starts from what those
families name, from the item, enchant and set-bonus tables and from the talent trees, and closes over
everything those reach. A Season of Discovery rune ability is left out. That is a family in which an
`Engrave <slot> - <rune>` spell reaches every spell, with no trainer or level row outside the rune lines
and no talent reaching it; Shadow Cleave is one. The undescribed dummy that a rune's action bar override
replaces, Metamorphosis : Demon Charge, is left out too, whatever its rows. A rune ability that a
trainer also teaches, such as Victory Rush or Incinerate, stays. The `// Not generated:` comment at the
head of the class file names each family that could not be resolved and each rune family.

`assets/db_inputs/spell_store_inputs.json`, committed beside the store, holds the game files rows the
store was built from. It also holds every `SpellShapeshiftForm` row whole, since a form names no spell
by which the store's closure could reach it. Despite the name it is gzip-compressed JSON, like
everything under `assets/db_inputs/dbc`; read it with `zcat`. It lets the store, the buffs and the
shapeshift forms all be rebuilt without the game files database, which is gitignored and comes from a
local WoW install. `-check` compares it too, because a capture that no longer matches the database
would still render the same committed files, and nothing else would notice it going stale.

What guards the outputs:

- `TestStoreRegeneratesFromTheCommittedInputs` in `tools/database` re-derives the closure, the hand
  links, the curves, the tooltip hints, the overrides and the emitter from that capture and asserts the
  committed store is what comes out. It needs no build tag and no database, so it runs everywhere.
  `TestBuffFilesRegenerateFromTheCommittedInputs` does the same for the buff files, and
  `TestFormsRegenerateFromTheCommittedInputs` for `sim/core/dbcenums/forms_auto_gen.go`.
- `sim/core/spelldata/snapshot_test.go` pins the shape and the counts of the committed store, and a
  handful of rows read out of the game files by hand, so a regeneration that moves a number says so there
  instead of in a sim result.
- The repository's `pre-commit` hook runs `-check` when the database is present and the commit touches
  the generator or one of its outputs.

The .results files are the last check, and the one that costs time: run the suites of the class that
changed, and diff the `.results` against the `.results.tmp` with a local diff helper or by hand. If a
port meant to be mechanical moves a .results number, the port is wrong; the number is not a new
baseline.

Taking only some classes or the racials from a new game build is described in
[Taking a new game build one spell data group at a time](partial-updates.md).
