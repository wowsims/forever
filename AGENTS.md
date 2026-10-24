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
| `wowsims-siml`                | SIML rotations, gear sets and setups: syntax, every action and value, CLI, language server     |
| `wowsims-bulk-sim`            | Bulk sim, locally, on the server and in the browser                                            |
| `wowsims-reforge-optimizer`   | The gear optimizer (EP, stat caps, HiGHS)                                                      |
