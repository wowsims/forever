# Taking a new game build one spell data group at a time

A new game build moves every class at once: talents leave the trees, rows change, and each class's
code has to follow before its tests pass. `-classes`, `-racials` and `-shared` take some spell data groups of the
database and keep every other row as the committed inputs state it, so each class moves onto the build
in its own pull request:

```
make db CLASSES=druid                                # extract, then gen_spelldata and gen_db with -classes=druid
go run ./tools/database/gen_spelldata -classes=druid -racials
go run tools/database/gen_db/*.go -outDir=./assets -gen=db -classes=druid
```

A run chooses which of the nine classes it takes (by package name), whether it takes `racials`, and
whether it takes `shared` (the shapeshift forms, and the extra spells and buff manifest rows no class
owns). `shared` is taken unless `-shared=false` keeps it as committed. One spell data group cannot be
chosen: every run takes `items` (the item, enchant and set-bonus rows, with the item database `gen_db`
writes) from the database, because the item database is not split and the store's item rows have to
match it. So `-classes=druid` moves the druid, the items and the shared rows, `-racials` moves the
racials, the items and the shared rows, and `-classes=none -shared=false` moves the items alone.

| Spell data group                                                                        | Chosen with                                       |
| --------------------------------------------------------------------------------------- | ------------------------------------------------- |
| `warrior`, `paladin`, `hunter`, `rogue`, `priest`, `shaman`, `mage`, `warlock`, `druid` | `-classes` (`CLASSES=` in make)                   |
| `racials`                                                                               | `-racials` (`RACIALS=1` in make)                  |
| `shared`                                                                                | taken unless `-shared=false` (`SHARED=0` in make) |
| `items`                                                                                 | not chosen: every run takes it                    |

A row belongs to the first of these that holds:

1. it is an `overrides.ExtraSpells` entry with `SpellDataGroup: "racials"`, the Eureka! rows;
2. it states a class family: its class flags are that class's, so a set bonus or an enchant's class
   buff moves with the class's own spells;
3. a group's roots reach it. A row that a taken group reaches moves. When a group that is not taken
   also reaches it and the row changes, it is printed as a shared row.

Every run, full or partial, prints what it changes against the committed inputs, one block per group it
takes: the build it moves from and to, the counts, and each spell added, changed (with the tables whose
rows differ) or removed.

```
spelldata: warrior 1.60.1.70124 -> 1.60.1.70170: 5 added, 40 changed, 4 removed
spelldata:   + 1323963 Furious Precision
spelldata:   ~ 23881 Bloodthirst (Effects)
spelldata:   ~ 12321 Booming Voice (Descriptions, Effects, TraitPoints)
spelldata:   - 1310236 Boundless Rage
spelldata: items 1.60.1.70124 -> 1.60.1.70170: 6 added, 225 changed, 3 removed
spelldata: shared row 10612 Windfury Totem changed, also reached by shaman
```

Two rules come before all three. A committed root of a group that is not taken never moves: that
class file names it and keeps the row it was written against. A spell that the new build adds and a
taken row or root reaches always comes from the new build, since no committed file names it.

A class that is not taken keeps its `sim/<class>/spell_data_auto_gen.go`. A class the run cannot read
from the new build yet is reported, and the run goes on. `gen_db -classes` writes only the taken
classes' talent protos, TS configs and tree JSON. The hunter's ammo file goes with the hunter, though
the item database always ships the new build's arrows.

A class that is not ported yet breaks only its own packages and `sim`, which registers every class. The
tests of shared code (`sim/common`, `sim/core/bulk`, `sim/core/reforge_optimizer`) build their players
with `sim/core/testagent`, a character with no class code, and do not borrow a class's agent. A test
that runs real specs, such as the reforge optimizer's presets, lives in `sim` with every class.

While the inputs record one game build, a run that names no groups takes all of them, which is a full
regeneration. Once a partial update leaves groups on different builds, such a run takes only the
groups on the database's build and prints the rest, so `make db`, `make spelldata` or the
`all_items.go` rule never moves a class that has not been ported. The inputs record the game build each
group was taken from (`Builds`), the highest hotfix push it was taken with (`Hotfixes`), both read
from the `BuildInfo` table `db2tool` writes, and the roots of each group (`SpellDataGroupRoots`). The next partial update keeps the other groups by those roots. A racial row on a
different build from the class whose flags it states is printed as a warning.

A game files table the store starts reading reaches only the groups a run takes. The other groups keep
their rows as the inputs commit them, without that table, until a later run takes them.

`-check` renders the groups whose recorded build is the database's and compares those. A group on
another build is printed as not ported and does not fail the check. Its class file is not compared,
since only its own build could render it. Items are always compared. A re-extraction that picked up
new hotfixes fails `-check` on the current groups until they are regenerated. Inputs that record no
groups are split the way the database splits the same ids, and a root that names no group there is
printed as adopted into `shared`.

A database extracted with older hotfixes than the inputs record is refused. When a group the run
takes is on the database's game build and records a higher push than the database's, `gen_spelldata`
(`-check` included) stops before writing anything and names the groups, and `gen_db` does the same
for the groups it takes, `items` always among them. Taking such a database would drop every row
only a later hotfix adds. Log into the game to refresh `DBCache.bin`, or pass a newer cache with
`make db DB2TOOL_FLAGS="--dbcache path/to/DBCache.bin"`, and run `make db` again. No flag skips the
check. `db2tool` prints the push it applied as `Hotfixes: build 70235 up to push 112426`.

Two pull requests that each take a class both rewrite the store and the gzipped inputs. Resolve the
conflict by taking the base branch's copies and running `gen_spelldata -classes=<class>` again; the merge
is deterministic.

## Commands

`-classes` takes `warrior`, `paladin`, `hunter`, `rogue`, `priest`, `shaman`, `mage`, `warlock` and
`druid`, comma-separated, or `all` or `none`; `-racials` takes the racials; `-shared=false` keeps the
shared group as committed. Any of them makes the run a partial update, and `-racials` or
`-shared=false` with no `-classes` takes no class. No flag can leave out `items`: every run, partial
or full, takes it from the database.

```
make db                                   # extract, then gen_spelldata and gen_db, default groups
make db CLASSES=druid                     # take the druid, with items and shared
make db CLASSES=druid,mage RACIALS=1
make db RACIALS=1                         # the racials, with items and shared, no class
make db CLASSES=none                      # items and shared alone
make db CLASSES=none SHARED=0             # the items alone, as the Update Items workflow runs it
make db CLASSES=all RACIALS=1             # every group, also while the builds are mixed
make db CLASSES=warrior WOW_DIR="/mnt/e/World of Warcraft"
make db CLASSES=mage DB2TOOL_FLAGS="--cdn --dbcache caches/DBCache.bin"
make ptrdb CLASSES=druid                  # the same from the PTR settings
make spelldata CLASSES=hunter             # gen_spelldata alone, no extraction
make spelldata-check                      # stale files exit 1; groups not ported are listed and pass
```

```
go run ./tools/database/gen_spelldata                               # default groups
go run ./tools/database/gen_spelldata -classes=druid
go run ./tools/database/gen_spelldata -classes="druid, mage" -racials
go run ./tools/database/gen_spelldata -racials
go run ./tools/database/gen_spelldata -classes=none -shared=false   # the items alone
go run ./tools/database/gen_spelldata -classes=druid -unchecked     # write without the type-check
go run ./tools/database/gen_spelldata -check                        # compare, write nothing
go run ./tools/database/gen_spelldata -dbPath /path/to/other.db -classes=druid

go run tools/database/gen_db/*.go -outDir=./assets -gen=db                 # talent files of the default groups
go run tools/database/gen_db/*.go -outDir=./assets -gen=db -classes=druid  # the druid's talent files only
```

`-classes`, `-racials` and `-shared` do not combine with `-check` or `-everyClassSpell`. `gen_db`
takes only `-classes`, since it has no racial or shared files, and always writes the item database.

The Update Items workflow (`.github/workflows/update_items.yml`) extracts the current build from the
CDN (product `wowf`) with the latest hotfixes and runs `make db CLASSES=none SHARED=0`, so its pull
request moves the items and leaves every class, the racials and the shared rows on their committed
builds. Raidbots labels the live game's files `forever`. A new one can be for an older build than
the CDN serves, so `--carry-hotfixes` carries such a file onto the extracted build (the rules are in
`tools/db2tool/README.md`). The last beta file, `classic_beta` build 70532, is carried under the live
one on every run: it holds about 4,300 item rows the beta looked up that the live file lacks. A row
the live file marks invalid still drops out, since the extracted build's own file applies last. The
Check Hotfixes workflow dispatches Update Items every hour that a new `forever` file appears for the
CDN build or an older one; a pull request with the hotfixes applied that passes every test is merged
and deployed without review.

`db2tool` writes the `BuildInfo` table on every extraction: the game build, the product and
`HotfixPush`, the highest hotfix push it applied (0 when it applied none, as in offline mode without
`--dbcache` or with `--no-hotfixes`; with `--carry-hotfixes`, the highest across every cache that
applied). In offline mode pass the full version, `--build 1.60.1.70170`; a bare build number is
recorded as given, with a warning. Stamp a database extracted without `BuildInfo` instead of
extracting it again, since a new extraction can pick up hotfixes the committed rows do not have.
Stamp the push `db2tool` printed for that extraction, or 0 when it is not known; a database whose
`BuildInfo` has no `HotfixPush` column reads as 0:

```
sqlite3 tools/database/wowsims.db "CREATE TABLE BuildInfo (Version TEXT NOT NULL, Product TEXT NOT NULL, HotfixPush INTEGER NOT NULL); INSERT INTO BuildInfo VALUES ('1.60.1.70235','wowf',112426);"
sqlite3 tools/database/wowsims.db "SELECT * FROM BuildInfo"
```

The build and hotfix push each spell data group is on, and the conflict resolution between two class
pull requests:

```
zcat assets/db_inputs/spell_store_inputs.json | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['Builds'], d.get('Hotfixes'))"

git checkout origin/master -- sim/core/spelldata/spells_auto_gen.go assets/db_inputs/spell_store_inputs.json
go run ./tools/database/gen_spelldata -classes=mage
```

A new game build, start to finish:

```
make db RACIALS=1                         # the update pull request: items, shared and racials
make db CLASSES=druid                     # each class on its own branch, once the update is merged
make spelldata-check                      # the classes still to port
go test ./tools/database/ -count=1 -run 'Merge|MoveRow|SpellDataGroups'
go test --tags=with_db -count=1 ./sim/druid/...
```
