# Keeping wowsims-spell-store true

Re-derive a claim before repeating it:

- which files are on the resolvers: `grep -l 'spelldata.SpellConfig\|spelldata.Parse' sim/*/*.go`
- the tool's arguments: `go run ./tools/spelldata` with no arguments prints the usage line (`usage` in
  `tools/spelldata/main.go`)
- the ladder readers: `grep -n '^func (l Ladder)' sim/core/spelldata/ladder.go`
- the effect readers: `grep -n '^func (e \*Effect)' sim/core/spelldata/effect.go`

| You changed…                                                    | Update                                                                                                                    |
| --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `sim/core/spelldata/ladder.go`                                  | "Reading a row", the `Highest()` trap, the ladder table (its `Highest()` row too) in `docs/spell_data/store.md`           |
| `sim/core/spelldata/effect.go`, `spell.go` or `store.go`        | "Reading a row", `docs/spell_data/store.md`, the `-expr` example in `tools/spelldata/README.md`                           |
| a class file moved between hand reads and the resolvers         | "Hand reads and resolvers", the opening sentences of `docs/spell_data/reading-by-hand.md` and `docs/spell_data/README.md` |
| the tool's arguments in `tools/spelldata/main.go`, or its hover | "The spelldata tool" and the command block in `tools/spelldata/README.md`                                                 |
| `discoverLadders` in `tools/database/gen_spell_data.go`         | `references/row-traps.md`                                                                                                 |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-spell-store
```

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
