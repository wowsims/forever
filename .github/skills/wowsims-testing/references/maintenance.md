# Keeping wowsims-testing true

Re-derive a claim before repeating it, the gates with the print commands in §The gates:

- the tagged test files: `grep -rl '^//go:build with_db' sim tools --include=*.go`; the panic without the tag:
  `grep -n 'No DB data' sim/core/test_generators.go`
- where `.results` live and which carry spell ids: `find sim -name '*.results'`, `grep -rl spell_id sim --include=*.results`
- what a `.results` row holds: `grep -n -A6 'message DpsTestResult' proto/test.proto`
- the ignored files: `grep -n 'results' .gitignore`, `grep -n -A3 '^# Baselines' .gitignore`
- the pin helpers: `grep -n 'func withRows\|func editRow' sim/core/spelldata/spelldata_test.go sim/common/shared/spell_data_heal_proc_test.go`
- `references/traps.md`: `grep -n 'RandomSeed + i' sim/core/sim.go`, `grep -n DiscardUnknown cmd/wowsimcli/cmd/basic_sim.go`

| You changed…                                                                                | Update                                    |
| ------------------------------------------------------------------------------------------- | ----------------------------------------- |
| a step in `run_tests.yml`, a make test target, `database_load.go` or a tag                  | §The gates                                |
| `test_suite.go`, `test_generators.go` or `proto/test.proto`                                 | §.results discipline                      |
| a snapshot list, a `withDBCInputs` test or an ignored input                                 | `references/stale-inputs.md`              |
| a pin helper, or where a kind of test lives                                                 | §What a test asserts, "Where things live" |
| `feralbear_test.go`, `mana_regen_test.go`, `encounter.ts`'s fallback target, `basic_sim.go` | `references/traps.md`                     |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-testing
```

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
