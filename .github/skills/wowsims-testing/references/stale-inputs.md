# Stale ignored inputs

The Go tests read files git ignores, so a red test after a merge, a pull or on a fresh checkout is first a question
about those files. Rebuild them before calling master red.

## A fresh checkout

`sim/` builds and its suites run once `sim/core/proto/*.pb.go` and `binary_dist/dist.go` exist
(`make sim/core/proto/api.pb.go binary_dist/dist.go`) and `npm ci` has run; `db.bin` is committed. The generators need
`tools/database/wowsims.db`, `assets/db_inputs/dbc/` and `tools/db2tool/listfile.csv` too, and the `tools/database`
tests that read the database or the DBC inputs skip without them. The UI's generated files are in
`.github/skills/wowsims-ui/references/verification.md` §A fresh checkout needs generated files first.

## After a merge or a pull

- **Protos first.** A merge that touched `proto/` leaves the compiled `*.pb.go` behind; rerun
  `make sim/core/proto/api.pb.go binary_dist/dist.go`.
- **The DBC inputs.** The `tools/database` tests that call `withDBCInputs` (`tools/database/enchant_procs_test.go`)
  read the ignored `assets/db_inputs/dbc/*.json`: a missing copy skips them and a stale one fails them.
  Regenerate them as `wowsims-db-pipeline` `## Regenerating` says.
- **The snapshot lists.** `sim/common/forever/registered_effects.txt`, `tools/database/unsupported_procs.txt` and
  `tools/database/unsupported_consumables.txt` are local baselines their tests write when missing. After master
  brings new effects or rows, rewrite them with `UPDATE_REGISTERED_EFFECTS=1` (`TestRegisteredEffects`),
  `UPDATE_UNSUPPORTED_PROCS=1` and `UPDATE_UNSUPPORTED_CONSUMABLES=1` before reading their diff as a regression;
  each test's header line prints its exact command. A change to proc support in your own branch (a trigger newly
  built or refused) makes them stale the same way: the tests fail until you refresh the local copies.
- **The test cache.** Rerun the tools tests with `-count=1` after any of the above, so a cached pass from before the
  merge cannot stand in for a run.
