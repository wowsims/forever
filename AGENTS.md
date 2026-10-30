# Agent instructions

WoWSims Forever: a Go simulator (`sim/`) with a React frontend (`ui/`), built on data extracted from the game
files (`tools/database`, `sim/core/spelldata`).

The hunter's rotations, gear sets and sim setups are SIML files (`*.siml`); the other specs keep theirs as JSON
until each moves, and the Go suites and the site read either. `sim/siml/SPEC.md` is the language, and the
`wowsims-siml` skill a short guide to it with every action and value, where its tools are and how to start its
language server. A change to a proto, `proto/apl.proto` most of all, reaches SIML on its own; run `make siml` after
one, and fix the `.siml` files the check names when a kind or field is taken out or renamed.

**Rule: tests and stories never add code to the sim or the UI.** No export, prop, parameter, branch, flag or
hook goes into `sim/` or `ui/` only so a unit test or a Storybook story can reach or cover something. A test
or story works through the code's real surface, and a fixture or stub it needs lives with the tests. The
settled exceptions are listed in `wowsims-testing` §What a test asserts and
`.github/skills/wowsims-ui/references/conventions.md`.

The project's skills live in `.github/skills/`. Read the one that matches the task before starting, and always
read `wowsims-conventions` before writing comments, docs, test messages, commit messages or GitHub text.
`make skills-check` verifies every path and skill name the skills mention. `npm run skills:install` (part of
`make setup`) installs these skills and the `humanizer` skill for the agents on this machine. A change that edits
a skill reruns it and commits the refreshed `skills-lock.json` with the change: the agents read the installed
copy, and the lockfile records each skill's hash.

| Skill                         | Use it for                                                                                     |
| ----------------------------- | ---------------------------------------------------------------------------------------------- |
| `wowsims-conventions`         | How comments, docs, commits and PR text are written; "game files", not "client"; banned terms  |
| `wowsims-spells`              | The map: which spell data skill to read, where things live, when the sim and the row disagree  |
| `wowsims-spell-store`         | Reading the spell store: ladders, ranks, effects, hand reads and resolvers                     |
| `wowsims-abilities-auras`     | Building abilities, auras, dots, talents and buffs from their rows                             |
| `wowsims-procs`               | Proc triggers: chances, PPM, ICDs, charges, item, enchant and set procs                        |
| `wowsims-consumables-items`   | Items, enchants and consumables in the database, and the effects registered for them           |
| `wowsims-db-pipeline`         | Regenerating the store and the database, overrides, hotfixes, a database behind master         |
| `wowsims-core-mechanics`      | What `sim/core` does around a spell: cast order, spends, scaling, targets, hooks, auras, state |
| `wowsims-apl`                 | Rotations (APL): preset lines, timing, values, why a cast never fires                          |
| `wowsims-testing`             | Proving Go work: the CI gates, .results files, what a test asserts, runtime proofs             |
| `wowsims-spec-implementation` | Picking up a `[Spec]` tracking issue and porting a class                                       |
| `wowsims-review`              | Reviewing a spell change, your own before committing or a pull request, against the rows       |
| `wowsims-ui`                  | The frontend under `ui/`                                                                       |
| `wowsims-siml`                | SIML rotations, gear sets and setups: syntax, every action and value, CLI, language server     |
| `wowsims-bulk-sim`            | Bulk sim, locally, on the server and in the browser                                            |
| `wowsims-reforge-optimizer`   | The gear optimizer (EP, stat caps, HiGHS)                                                      |

Pull requests:

- **Review and simplify when asked.** Once a PR is open, ask the user before running the review pass
  (`wowsims-review`) and the simplify pass; fix their findings before the PR leaves draft.
- **Gates before every commit.** The CI gate set (`wowsims-testing` `## The gates`) runs locally before every commit,
  and the PR body states its results.
- **One worktree, one draft PR.** A big change runs in its own worktree and branch, with one draft PR opened at the
  first commit and its body kept current as each slice lands.

A ruling, trap or workflow fact learned while working is written down in the same change. A rule one spell, aura
effect, item or component owns is a comment at its implementation (the rule in present tense, no pull request number);
a rule no single file owns goes into the subject skill, or its `references/rulings.md`.
