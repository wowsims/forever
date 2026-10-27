# Keeping wowsims-spells true

| You changed…                      | Update                                                |
| --------------------------------- | ----------------------------------------------------- |
| added, renamed or removed a skill | "Which skill" in `SKILL.md` and the `AGENTS.md` table |

Then run `sh .github/skills/check_paths.sh .github/skills/wowsims-spells`.

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
