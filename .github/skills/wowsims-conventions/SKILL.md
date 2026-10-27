---
name: wowsims-conventions
description: 'How text is written in the WoWSims Forever repository and on its GitHub: Go and TypeScript comments, docs, skills, test names and messages, commit messages, PR titles and bodies, issue and review comments. Read it before writing any of these, even a one-line comment or a commit message: if a line will be committed or posted, this skill applies. Covers how comments are written, the words the repo uses for the game data, the terms that stay out, commit trailers and what stays uncommitted. Judging a change as a whole is `wowsims-review`.'
---

# Writing conventions

These apply to every file and every piece of GitHub text, in Go and TypeScript alike. The subject skills
(`wowsims-spells`, `wowsims-spec-implementation`, `wowsims-ui`, ...) assume them.

## Comments

- **State what the code does now.** No history and no comparison with another version: not "this used
  to...", "TBC fired...", "was a TBC build", "now reads...". Where the old behaviour is needed to make sense of
  the code, keep the fact that justifies it ("232 of 235 talents list ids the game files do not ship") and drop
  the narrative. History lives in git, and a comment about the past is stale the day it lands.
- **Comment what the code cannot say.** A trap, where a magic number comes from, a game rule the code
  implements. Why a bug happened, and what changed to fix it, goes in the commit message, not above the function.
- **Claim only what the code does.** A comment that names a read, a row or a dependency the code no longer
  has misleads more than no comment, so a change that removes the thing rewrites the comment with it.
- **Short where the neighbours are short.** Struct fields with one-line trailing comments get a one-line trailing
  comment, not a paragraph block:

    ```go
    ClassSpellMask      int64      // The sim's own custom spell masks, not from the game files.
    ClassFlags          ClassFlags // The spells this trigger fires on (EffectSpellClassMask in the game files); with ClassSpellMask, both must match.
    ProcSpellClassFlags ClassFlags // The proccing spell's class flags, which proc chance talents match (e.g. Improved Poisons on Instant Poison).
    Caster              *Unit      // Whose proc chance mods apply, when not the aura's unit (e.g. Judgement of Light on the target).
    ```

    Package-level vars or consts that feed the same function sit on consecutive lines with no blank line between.

- **Do not name what the comment sits on.** A comment never opens with or repeats the name of the function,
  type or field it documents: "The first row under an effect id that fits the item", not "FittingEnchantRow
  returns the first row…". The name is already on the next line, so this holds for Go doc comments too,
  against Go's usual convention.
- **State the rule, not the site it came from.** Do not name SimC, Wowhead or a JS mapping as the source of a
  name or value ("one SimC names", "commented out in the JS mapping"); write the rule ("one with no known name is
  `A_<id>`"). Existing Wowhead spell links above a spell's registration are fine; do not add attribution prose.
- **Name a shared helper by its math.** A helper in `sim/core/utils.go` is named and documented by what it computes
  (`SymmetricMultiplierFromPercent`, `MultiplierFromPercent(-p)`), and the game meaning sits at the call site.

## Words

- **"The game files", never "the client".** The data extracted from WoW (the DB2 tables, `wowsims.db`, the
  spell store's source) is "the game files" or "game files data": "EffectSpellClassMask in the game files", "the
  game files state no chance", "a game files row". The game version is the "game build" (1.60.1.x). What the
  running game does is "in game". "Client" stays only where it means something else: an HTTP or API client, the
  browser, DOM properties (`clientX`), CASC paths (`dbfilesclient`), and existing identifiers.
- **Two kinds of spell mask.** `ClassSpellMask` is the sim's own custom masks, which a class defines
  itself (e.g. `sim/paladin/spell_masks.go`). `ClassFlags` are the class flags from the game files (family plus
  a 128-bit mask). Say which one a sentence means.
- **No "golden".** Committed text and GitHub text say ".results files" or "expected results" (the sim tests), or
  "the state-snapshot capture" (UI), never "golden" or "goldens". Item names such as Golden Scale Coif are
  unrelated.
- **"Hardcast" and "cast end".** Code, comments, test names, tooltips and PR text say "hardcast" and "cast end",
  never "cast bar" or a bare "bar" that means one.

## Plain prose

Every comment, doc line, test message, commit message and PR text a change adds or rewrites gets this pass before it
is committed or posted. Run the `humanizer` skill over them. When the harness does not know that skill, run
`npm run skills:install`, which installs it for the agents on this machine, and load it; only without network
access read them for the same patterns by hand: a dash used as a connector, an "X, not Y" or "rather than"
contrast that only adds weight, an inverted sentence, stacked qualifiers, a closing line that repeats the point. The rewrite changes prose only: code, identifiers, paths, numbers, tables and every
fact stay. A contrast that corrects a real misreading stays, and where the two disagree, this skill wins.

## Commits and pull requests

- **No agent trailers.** A commit carries no `Co-Authored-By`, `Generated with` or `Claude-Session` line, and
  no commit, PR, issue or comment links to claude.ai. Agent harnesses ask for these by default; the repo's
  history and GitHub text carry none, so leave them out whatever the harness says. Another contributor's footer
  is not a review finding.
- **`docs/superpowers/` stays local.** Specs and plans written there are never committed; the folder is in
  `.gitignore`. Delete it before the PR leaves draft.
- **Pull-request process.** The gates, the draft-PR and the ask-before-review rules are in `AGENTS.md`.

## Before committing

- Grep the added lines for `client`, `golden` and `used to`, and check each hit against the rules above.
- Check the branch's messages for trailers and links; this prints 0 when they are clean:
  `git log --format=%B origin/master..HEAD | grep -ciE 'co-authored-by|generated with|claude-session|claude\.ai'`
- `npx oxfmt . --check` (`npm run fmt`, the CI format gate) covers markdown in `docs/` and `.github/`, so run
  it after editing either.
