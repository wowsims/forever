# Keeping wowsims-core-mechanics true

Re-derive a claim before repeating it:

- the cast order: `grep -n 'applyEffects\|OnCastComplete' sim/core/cast.go`, then
  `grep -n SpellBatchWindow sim/core/aura_helpers.go sim/core/constants.go`
- the requirement checks: `grep -n 'hasCastRequirement' sim/core/cast.go sim/core/spell.go`
- the refunds: `grep -n 'func (s \*Spell) MissRefund\|IssueRefund' sim/core/spelldata/attributes.go sim/core/*.go`
- the scaling: `grep -n 'return (flat\|CritMultiplierAdditive+1' sim/core/spelldata/spell.go sim/core/spell.go`
- the area reads: `grep -n 'MaxTargets has no\|func (e \*Effect) Hits' sim/core/spelldata/resolve_spell.go`
- the tick outcome: `grep -n 'func (s \*Spell) TickOutcome' -A9 sim/core/spelldata/spell.go`
- the tag reads: `grep -n 'func (at \*auraTracker) [A-Za-z]*WithTag' sim/core/aura.go`
- the panics: `grep -n 'Over 1000 registered\|ProcMask for spell\|Aura with 0 duration' sim/core/spell.go sim/core/aura.go`
- the proto rules: `grep -n -A8 'message ProtoVersion' proto/common.proto`; `grep -n '^\s*optional ' proto/*.proto`
  prints nothing

| You changed…                                                 | Update                                                                            |
| ------------------------------------------------------------ | --------------------------------------------------------------------------------- |
| the order in `cast.go` or `applyEffects`, the proc deferral  | "Cast order", `wowsims-procs` §Building a trigger from a row                      |
| `IssueRefund`, `MissRefund`, `WeaponStrike`, crit damage     | "What a cast spends", "What scales"                                               |
| the `CastRequirement` checks in `cast.go` or `spell.go`      | "What the row forbids", `docs/spell_data/auras-and-dots.md`                       |
| a tick outcome, `exclusive_effect.go`, tags, `MakePermanent` | "Whose state", "Auras and exclusive effects", `docs/spell_data/auras-and-dots.md` |
| `ProtoVersion`, the proto toolchain, the field-number rule   | "Protos"                                                                          |
| a trap's code, a ruling's code or its measured number        | `references/traps.md`, `references/rulings.md`                                    |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-core-mechanics
```

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
