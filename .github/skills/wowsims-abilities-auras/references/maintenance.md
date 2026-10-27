# Keeping wowsims-abilities-auras true

Re-derive a claim before repeating it:

- the resolver options: `grep -n '^func [A-Z][A-Za-z]*(' sim/core/spelldata/resolve_spell.go sim/core/spelldata/resolve_aura.go`
- the parse options: `grep -n '^func [A-Z][A-Za-z]*(.*ParseOpt' sim/core/spelldata/parse_effects.go`
- the GCD floor: `grep -n 'GCDMin' sim/core/cast.go sim/core/constants.go`
- the `SPELLMOD_*` names: `grep -n 'SpellModOp = ' sim/core/dbcenums/spellmods.go`
- the `FromCaster` slots: `grep -rn 'FromCaster(\|AttachDDBC(' sim/<class>/`
- the `SingleAura` reset: `grep -n 'SingleAura = ' sim/core/exclusive_effect.go`
- `Magic()`'s variance: `grep -n 'DirectVariance' sim/core/spelldata/resolve_spell.go`
- the stance forms: `grep -n 'stanceForms' sim/core/dbcenums/forms_auto_gen.go`

| You changed…                                                         | Update                                                                               |
| -------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| a `SpellOpt` or `AuraOpt` in `resolve_spell.go` or `resolve_aura.go` | "An ability", the option table in `docs/spell_data/building-an-ability.md`           |
| a class's `FromCaster` or `AttachDDBC` slots                         | "An aura or dot", the `FromCaster` example in `docs/spell_data/talents-and-auras.md` |
| `DotConfig`, `PeriodicDamageTick` or `TickOutcome`                   | "An aura or dot", `docs/spell_data/auras-and-dots.md`                                |
| a `ParseOpt`, `parse_effects_table.go` or the skipped report         | "A talent's modifiers", `docs/spell_data/talents-and-auras.md`                       |
| `weapon_scope.go`                                                    | `docs/spell_data/weapon-requirements.md`, the weapon ruling                          |
| `sim/core/buffs/` or `tools/database/buffmanifest/`                  | "Buffs and debuffs", `docs/spell_data/buffs-and-debuffs.md`, issue #36               |
| a ruling's code or its measured number                               | `references/rulings.md`                                                              |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-abilities-auras
```

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
