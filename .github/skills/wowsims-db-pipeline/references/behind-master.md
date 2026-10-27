# Working from a database behind master

Try to match master's database first; splice only when it cannot match.

## Matching master's database

The Update Items pull request names the cache it applied (`Hotfixes: … (from <url>)`). Take the run behind the latest
`db.json` commit, not the push id, since several runs can record one push and still change `db.json`. Extract with
`db2tool --cdn --dbcache <that DBCache.bin>`, then prove it on a clean checkout of `origin/master`: a regeneration
there leaves `git status` clean. A matching push stamp alone proves nothing. A class branch carries its own `Builds`
and `Hotfixes` in `spell_store_inputs.json`, so read them on the branch you regenerate.

## Splicing

- **A store change.** Regenerate at the base and with the change, splice only the change's own fields by effect id
  into `spell_store_inputs.json` through `writeStoreInputs` and `renderStore` in a throwaway test, and prove it with
  `TestStoreRegeneratesFromTheCommittedInputs`. Restore `db.json` and `db.bin` when they hold drift only.
- **A database entry.** Splice it into master's `db.json` and `db.bin` with a throwaway Go program that writes both,
  `protojson.Unmarshal` into `proto.UIDatabase` then `proto.Marshal`. Nothing derives one file from the other, so a
  hand edit to one leaves the other behind. Diff the JSON, and diff
  `protoc -I proto --decode=proto.UIDatabase proto/ui.proto < assets/database/db.bin` against master's; the binary is
  not byte-stable, so only the decoded diff shows what moved.
- **Encounter presets only.** `database.ReadDatabaseFromJson`, set `db.Encounters = core.PresetEncounters`, call
  `WriteBinaryAndJson`, and check the `db.json` diff holds encounter lines only. Creature health and armor reach
  `core.Creature*MaxLevel` through `tools/base_stats_parser.py`.

## Merging master into a branch

- **Store conflicts.** Take master's `spells_auto_gen.go` and `spell_store_inputs.json` and rerun
  `gen_spelldata -classes=<class>` (`docs/spell_data/partial-updates.md` §Commands); a hand merge of generated text is
  a store no input describes.
- **A `db.bin` conflict.** `protojson.Unmarshal` the merged `db.json` into `proto.UIDatabase`, `proto.Marshal` it, and
  check the two proto-equal. Never `ReadDatabaseFromJson` there: it leaves out `Encounters`.
- **`db.bin` behind an unchanged `db.json`.** `WriteBinaryAndJson` skips both files when the new `db.json` equals the
  one on disk and `db.bin` exists. Delete `assets/database/db.bin` before `gen_db` to force the write; `gen_db` imports
  no package that embeds it, so it still builds. `assets/database/loader.go` embeds `db.bin` with no build tag, and
  the untagged `sim/core` tests load it (`sim/core/buffs_generated_test.go`), so `go test ./sim/core/` does not
  compile until `gen_db` has written the file again.
- **Items from the hotfix imports.** A local regeneration drops the items master took from them. Carry those over
  from master's `db.json` and check each section as master plus the branch's own rows.
- **A change ported in another shape.** Merged back, it duplicates with no conflict (struct fields, proto field
  numbers, helper files). Diff the ported files against `origin/master`, take master's, then regenerate
  `spell_store_inputs.json`.
- **Proto field numbers.** `wowsims-core-mechanics` `## Protos`.
