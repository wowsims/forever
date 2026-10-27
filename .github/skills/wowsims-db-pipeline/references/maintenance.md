# Keeping wowsims-db-pipeline true

Re-derive a claim before repeating it:

- the make targets: `grep -n '^[a-z-]*:' makefile | grep -E 'db|spelldata'`, their flags: `grep -n '^DB2TOOL_FLAGS\|^WOW_DIR' makefile`
- the workflow steps: `grep -n 'name:\|make db' .github/workflows/check_hotfixes.yml .github/workflows/update_items.yml`
- the full order: `grep -n 'make \|go run' .github/workflows/update_db.yml`
- `gen_db`'s order: `grep -n 'RequireCommittedHotfixes\|RunOverrides\|WriteBinaryAndJson' tools/database/gen_db/main.go`
- the go-to-ts outputs: `grep -n '^AUTO_GEN_FILES_TS' makefile`
- the override tables: `grep -n '^var ' tools/database/overrides/*.go`
- the item level floor and the phase: `grep -n 'MIN_EFFECT_ILVL = \|Phase = 1' tools/database/gen_effects.go tools/database/gen_db/main.go`
- the PTR settings: `diff tools/database/generator-settings.json tools/database/ptr-generator-settings.json`
- who embeds `db.bin`: `grep -rn 'assets/database"' --include=*.go .`, then `head -1` of each file for its `//go:build`
- the fields `ReadDatabaseFromJson` keeps (no `Encounters`): `grep -n -A22 '^func ReadDatabaseFromJson' tools/database/database.go`

| You changed…                                                               | Update                                                                                                                                                       |
| -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| a make target, or the step order in `update_db.yml`                        | "Regenerating"                                                                                                                                               |
| a generator flag or the group rules                                        | `docs/spell_data/partial-updates.md` §Commands, "Regenerating"                                                                                               |
| an override field, table or stale rule                                     | `docs/spell_data/overrides.md`, "Overrides"                                                                                                                  |
| `RequireCommittedHotfixes`, `WriteBinaryAndJson` or `ReadDatabaseFromJson` | "A database behind master", `references/behind-master.md`                                                                                                    |
| `check_hotfixes.yml` or `update_items.yml`                                 | "Hotfixes", the Update Items paragraph in `docs/spell_data/partial-updates.md`                                                                               |
| `generator-settings.json`, its PTR copy, or `db2tool`'s flags or schema    | "Where things live", "Hotfixes", "Tables the generator does not ship", `references/game-files-tables.md`, `tools/db2tool/README.md` §generator-settings.json |
| `assets/database/loader.go` or a package importing it                      | `references/behind-master.md` §Merging                                                                                                                       |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-db-pipeline
```

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
