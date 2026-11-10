# Keeping wowsims-review true

| You changed…                                                                  | Update                                         |
| ----------------------------------------------------------------------------- | ---------------------------------------------- |
| what `go run ./tools/spelldata` prints, or `sim/core/spelldata/attributes.go` | step 3's commands and the paragraph under them |
| a defect class, or a `##` heading step 4 names in another skill               | `references/defects.md` and that step 4 bullet |

Then run `sh .github/skills/check_paths.sh .github/skills/wowsims-review`.

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
