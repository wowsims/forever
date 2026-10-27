# Keeping wowsims-procs true

Re-derive a claim before repeating it:

- the trigger and its options: `grep -n 'func Proc\|) ProcOpt {' sim/core/spelldata/*.go`
- the decoder: `grep -n 'func DecodeProcTypeMask' sim/core/proc_types.go`
- the zero chance turned into 1: `grep -n 'ProcChance == 0' sim/core/aura_helpers.go`
- the item proc builders: `grep -n '^func NewSpellData[A-Za-z]*Proc' sim/common/shared/*.go`
- the bleed call sites: `grep -rn 'ProcMask = core.ProcMaskUnknown' sim/`
- the spender and extra attacks: `grep -rn 'func AttachChargeSpender\|ExtraAttackProc(' sim/core/`
- the override tables: `grep -nE '^var (WeaponProcs|EnchantProcs)' tools/database/overrides/*.go`
- the notice's origins and literals: `grep -n 'func rateOrigin\|var literalRates' tools/database/gen_effects.go`

| You changed…                                                | Update                                                                       |
| ----------------------------------------------------------- | ---------------------------------------------------------------------------- |
| a `ProcOpt` in `resolve_proc.go`                            | "Chance and masks", the option table in `docs/spell_data/procs.md`           |
| `DecodeProcTypeMask` or a `ProcHint`                        | "Traps", `docs/spell_data/procs.md` §The decoder and the tooltip hints       |
| `AttachChargeSpender` or `ExtraAttackProc`                  | "Building a trigger from a row"                                              |
| a `NewSpellData*Proc` builder or an override table's fields | "Item and enchant procs", `docs/spell_data/procs.md` §Item and enchant procs |
| `rateOrigin`, `literalRates` or `item_notices.tsx`          | "Measuring a rate", the carry-over ruling in `references/rulings.md`         |
| a ruling's code or its measured number                      | `references/rulings.md`                                                      |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-procs
```

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
