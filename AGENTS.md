# Agent instructions

WoWSims Forever: a Go simulator (`sim/`) with a React frontend (`ui/`), built on data extracted from the game
files (`tools/database`, `sim/core/spelldata`).

**Rule: tests and stories never add code to the sim or the UI.** No export, prop, parameter, branch, flag or
hook goes into `sim/` or `ui/` only so a unit test or a Storybook story can reach or cover something. A test
or story works through the code's real surface, and a fixture or stub it needs lives with the tests.

The project's skills live in `.github/skills/`. Read the one that matches the task before starting, and always
read `wowsims-conventions` before writing comments, docs, test messages, commit messages or GitHub text.

| Skill                         | Use it for                                                                                     |
| ----------------------------- | ---------------------------------------------------------------------------------------------- |
| `wowsims-conventions`         | How comments, docs, commits and PR text are written; "game files", not "client"; banned terms  |
| `wowsims-spells`              | Spell data: registering spells, talents, auras, procs from the game files; the spelldata store |
| `wowsims-spec-implementation` | Picking up a `[Spec]` tracking issue and porting a class                                       |
| `wowsims-review`              | Reviewing a spell change, your own before committing or a pull request, against the rows       |
| `wowsims-ui`                  | The frontend under `ui/`                                                                       |
| `wowsims-bulk-sim`            | Bulk sim, locally, on the server and in the browser                                            |
| `wowsims-reforge-optimizer`   | The gear optimizer (EP, stat caps, HiGHS)                                                      |
