---
name: wowsims-testing
description: 'Use when WoWSims Forever Go work has to be proven: running the gates run_tests.yml runs before a commit, the with_db tag, a class suite''s .results file, make update-tests, writing a test for an item, enchant, buff, proc rate or class rule and choosing its file, a snapshot test, or printing a value at runtime before claiming it. Use it too when a spec package panics with "No DB data for enchant", a package passes with none of its tests compiled, a .results file moves with no cause or disappears after update-tests, TestRegisteredEffects or a tools/database test turns red after a merge, a test skips instead of failing, a suite cannot see the spell a change touched, a test asserts a number the game files state, or a piped gate looks green. UI gates and component tests are wowsims-ui; regenerating generated files is wowsims-db-pipeline; the APL validation dump is wowsims-apl; judging a diff is wowsims-review; the port sequence is wowsims-spec-implementation; the spell skills map is wowsims-spells.'
---

# wowsims-testing

A sim change is proven by a run that could have failed on it. The rule this skill exists for: before reading a pass as
evidence, check that the suite exercises the change, that the tests compiled at all, and that no `.results` file moved
for a reason you did not intend. Text here follows `wowsims-conventions`.

## Where things live

| Path                                                    | What it is                                                                               |
| ------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `.github/workflows/run_tests.yml`                       | the CI gates: `build-ui`, `test-ui` (vitest, sharded) and `test` (Go, sharded)           |
| `makefile`                                              | `test`, `update-tests`, `spelldata-check`, `skills-check`                                |
| `sim/core/database_load.go`                             | the `with_db` file: loads `db.bin` into the sim's item, enchant and consumable tables    |
| `sim/core/test_suite.go`                                | runs a class suite, compares it against `<Suite>.results`, writes `<Suite>.results.tmp`  |
| `sim/core/test_generators.go`, `sim/core/test_utils.go` | the suite's test matrix (`AllItems` among it) and the helpers a throwaway test builds on |
| `sim/core/spelldata/spelldata_test.go`                  | `withRows`: a test's own copy of the store rows                                          |
| `sim/common/shared/spell_data_heal_proc_test.go`        | `editRow`: one row edited for the length of a test                                       |
| `sim/common/shared/spell_data_aura_test.go`             | the shared item, enchant and aura harness (`auraTesterAt`)                               |

| You are…                                                         | Open                                                                      |
| ---------------------------------------------------------------- | ------------------------------------------------------------------------- |
| running the UI gates, or reading vitest output                   | `.github/skills/wowsims-ui/references/verification.md`                    |
| red after a merge, a pull or on a fresh checkout                 | `references/stale-inputs.md`                                              |
| listing what a spell parse skipped (`SPELLDATA_REPORT=1`)        | `wowsims-abilities-auras`                                                 |
| dumping a preset's validations                                   | `wowsims-apl` §What fails silently                                        |
| conflicting on `db.bin`, or `go test ./sim/core/` will not build | `.github/skills/wowsims-db-pipeline/references/behind-master.md` §Merging |
| regenerating the store or the database before a test             | `wowsims-db-pipeline` `## Regenerating`                                   |
| a row that moves on its own, comparing runs, a preview           | `references/traps.md`                                                     |

## The gates

When the gates run and where their results go is `AGENTS.md` (Pull requests); judging the diff
before a commit is `wowsims-review`. Print the gates instead of quoting them from memory (the print gives the shard
counts); the list after the print says what each Go step of CI catches.

```
grep -n 'name:\|run:\|shard:' .github/workflows/run_tests.yml
node -e "console.log(require('./package.json').scripts)"
grep -n -A3 '^test:\|^update-tests' makefile
```

- **`make sim/core/proto/api.pb.go binary_dist/dist.go`**: the compiled protos and the dist stub; nothing under `sim/`
  builds without them.
- **`go test ./tools/database/... ./tools/gen_buffs_proto/...`**, untagged, on one shard: the buff manifest, the proto
  emitter and the buff generator. Tests that need `wowsims.db` or the DBC inputs skip there; run them locally.
- **`go test --tags=with_db` over `./sim/...`**, `GOARCH=amd64`: every class suite against its `.results`, the shared
  and core tests, and `TestProtoVersioning`, which shells out to `npx buf` and so needs `npm ci`.
- **`make skills-check`** in `build-ui`: every path, route and skill name the skills under `.github/skills/` mention.
  The format gate on markdown is `wowsims-conventions` `## Before committing`.

Locally, also run `go vet -tags=with_db ./sim/... ./tools/...`, `go test --tags=with_db ./tools/spelldata/`, and
`make spelldata-check` when a change touches spell data (it needs `wowsims.db`). A generator change regenerates and
diffs (`wowsims-db-pipeline`). One test, for a throwaway print or a single suite, runs with `-v`, since in package-list
mode `go test` drops a passing package's output, `SPELLDATA_REPORT` lines included:

```
go test --tags=with_db ./sim/<class>/... -run <Test> -count=1 -v
```

- **`with_db` decides what compiles.** Without the tag `sim/core/database_load.go` is left out and the item, enchant
  and consumable tables start empty: the class suites panic with "No DB data for enchant" (`FindAllEnchants` in
  `sim/core/test_generators.go`), which is the missing tag, not a data gap. Every test file that opens with
  `//go:build with_db` (`sim/common/forever/`, the feral suites' mechanics tests, `sim/common/registered_effects_test.go`)
  drops out too, so `go test ./sim/common/forever/` without the tag passes having compiled none of them.
- **Read the summary line, not the exit code.** A gate piped through `tail` or `grep` exits with the last command's
  status, so a red run looks green; run it bare, or check for `FAIL` or vitest's "Test Files" line before committing.

## .results discipline

Each class suite records its numbers in `<Suite>.results` beside it, under `sim/<class>/` or `sim/<class>/<spec>/`, and
every run writes `<Suite>.results.tmp`, which git ignores; its presence means nothing.

- **A file that cannot move proves nothing.** Before reading a pass as evidence, check the suite can see the change:
  `grep -rn <id> sim/<class> --include='*.results'`. Only the `*Coverage.results` files key cast counts by spell id;
  the other suites key by test name and `AllItems` by item id, so for most classes the grep finds nothing and the
  proof is a cast count or a value printed at runtime (§Proving a claim at runtime). Cast counts must be possible;
  cases: `.github/skills/wowsims-review/references/defects.md` §What the row forbids.
- **`make update-tests` deletes every `.results` before copying each `.results.tmp` over it**, so a suite that did not
  run loses its file, and the next run fails every one of its tests as unexpected. Run the tests, `cmp` each
  `.results.tmp` against its `.results`, and copy only the ones that moved on purpose; afterwards
  `git status --porcelain -- '*.results'` is empty unless a number was meant to move.
- **One isolated cause per move**, found by reverting one change at a time; rows that move on their own are in
  `references/traps.md`.
- **`.results` carry no OOM metric**: `TestDPS` records DPS, TPS, DTPS, HPS and TMI (`proto/test.proto`). Compare an OOM
  change with a throwaway test on the merge base and on the branch.

## What a test asserts

- **Tests add no production code**: the rule is in `AGENTS.md`, and `wowsims-review` step 2 checks for it. The
  settled UI exceptions are in `.github/skills/wowsims-ui/references/conventions.md`.
- **`DynamicProcManager.Chance` stays exported for the shared tests, which read an exact proc chance through it.**
  `sim/core/procs.go:35`, PR #187.
- **Logic, not data.** Outside `sim/<class>/` a test asserts what an effect does (what it modifies, when it fires, what
  it spends, how it scales), so a game build that changes an amount leaves it green. Expected values come from the raw
  row, `spelldata.MustFind(id).EffectN(i)`, not from the generated accessor the production path reads or a literal; a
  test over the real tables checks only data-independent invariants, never counts or specific ids.
- **Pin the precondition on a row copy** (`withRows`, `editRow`), never `t.Skip`; a needed amount that reads 0 fails
  the test. Names and messages quote no game-file numbers.
- **Where a new test goes.** A new class test outside the spec suites goes into `sim/<class>/<class>_test.go`, one test
  per fixed bug or class rule; a change adds no ability-specific test file under `sim/<class>/` or a spec folder, and
  proves an ability timing change with an uncommitted run stated in the pull request. A new test of cast requirement,
  form, equipment, buff or proc mechanics goes in the generic core or `sim/common/shared` tests. Item, enchant and
  generic aura behaviour goes in `sim/common/shared/spell_data_aura_test.go`; class items together in
  `sim/common/forever/class_items_test.go`; buffs and debuffs in `sim/core/buffs/meta_constructors_test.go` (shapes)
  and `sim/core/buffs/meta_test.go` (values). UI component tests are `wowsims-ui`.
- **A snapshot test writes its list when missing** and rewrites it under an `UPDATE_*` variable, and its path joins the
  snapshot block in `.gitignore`; the suites' `.results` stay tracked.

## Proving a claim at runtime

A claim about a value, a count, an order or a consumption (which handler sees which cast, when a charge goes) is
printed before it is made; reading the call order is not enough, and the orders themselves are
`wowsims-core-mechanics` `## Cast order`. Write a throwaway `_test.go` in the package under test, build the sim with
the helpers in `sim/core/test_utils.go` or the shared harness, `t.Logf` the value, run it as §The gates shows, and
delete the file before committing. Name the printed value in the claim ("`maxEnergy=110` at rank 2").

- **A proc rate never moves the class `.results`**: the `AllItems` rows do not proc chance-on-hit weapons. Prove a
  rate in `sim/common/shared` by building the listener and reading the chance or counting casts
  (`TestWeaponProcRateComesFromTheItemEffect` in `sim/common/shared/spell_data_weapon_rate_test.go`).

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The commands that
re-derive each claim, the "You changed → Update" table and the path check are in `references/maintenance.md`.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill otherwise. A memory file is a pointer
to this skill, not a second copy.
