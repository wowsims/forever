# Keeping wowsims-apl true

Re-derive a claim before repeating it:

- the files: `ls sim/core/apl_values_*.go sim/core/apl_actions_*.go`
- the validations: `grep -n 'Validation' sim/core/apl.go | head`, `grep -n 'ValidationMessage' sim/core/apl_helpers.go`
- the silent drops: `grep -n -A2 'default:' sim/core/apl_value.go sim/core/apl_action.go`, then
  `grep -n 'could not be resolved' sim/core/apl_action.go sim/core/apl_values_operators.go`
- the constant folds: `grep -n 'func (rot \*APLRotation) newValue[A-Za-z]*IsKnown' sim/core/apl_values_*.go`
- the kind picker's lists: `grep -n 'IsCastable' sim/core/unit.go`
- the major cooldowns: `grep -n SpellFlagMCD sim/core/apl_actions_casting.go`, `grep -n removeFromMajorCooldowns sim/core/apl.go`
- the preset loaders: `grep -n 'func GetAplRotation\|func GetTestBuildFromJSON' sim/core/test_utils.go`
- the layouts: `ls -d ui/specs/*/*/apls ui/specs/*/*/builds`; Pet Attack: `grep -rln petAttack ui/specs/*/*/apls ui/specs/*/*/builds`
- the option groups: `grep -n 'groupName ===' ui/specs/*/*/spec.ts`
- the dump's shape: `grep -n 'func bearAPLWarnings' sim/druid/feralbear/apl_sweep_test.go`

| You changed…                                                       | Update                                                                    |
| ------------------------------------------------------------------ | ------------------------------------------------------------------------- |
| the value or action switch, a constructor's nil path, `valueKinds` | "Values and actions"; a value's or action's rule: `references/rulings.md` |
| a lookup's warning, condition resolution, a preset loader          | "What fails silently"                                                     |
| the pick at the cast end, `readLanding`, `Spell Is Casting`        | `docs/rotation.md`, "Timing"                                              |
| `APLActionCastSpell.IsReady`, the reaction-time reads              | "Timing"                                                                  |
| a preset group name, a `simpleRotation` toggle, a suite's un-hide  | "Presets"                                                                 |
| `proto/apl.proto` or the `.vscode/settings.json` schema mapping    | `tools/APL.md`                                                            |

Then run the shared path check, `sh .github/skills/check_paths.sh .github/skills/wowsims-apl`.

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
