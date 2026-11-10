# Keeping wowsims-spec-implementation true

| You changed…                                                     | Update                           |
| ---------------------------------------------------------------- | -------------------------------- |
| how `tools/database/gen_spell_data.go` skips Engrave-only spells | "Deciding what is actually live" |
| the `spelldata-check` target in `makefile`, or the test tags     | "Verifying"                      |
| a `##` heading this skill names in another skill                 | the pointer that names it        |

Then run `sh .github/skills/check_paths.sh .github/skills/wowsims-spec-implementation`.

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
