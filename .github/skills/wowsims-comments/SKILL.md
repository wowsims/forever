---
name: wowsims-comments
description: 'Use when a comment in WoWSims Forever Go or TypeScript code is longer than a line or two, and before a commit that adds or rewrites one: a comment citing an EV- or R- id, the comment over a hand-written item, set or enchant effect, a doc comment on a core helper or a generator, a class ability explainer, a preset comment. Use it whenever asked to cut, trim, shorten or simplify comments, to remove comment bloat, or to review whether a comment says too much. Gives what a comment may hold, the size each kind gets, what is left alone, and the proof that a pass touched comments only. How a comment is worded is `wowsims-conventions`; what an EV- or R- file holds is `wowsims-evidence`.'
---

# Comment size

`wowsims-conventions` `## Comments` says how a comment is worded. This skill says how much of one stays. Run it over
every comment a change adds or rewrites, before the plain-prose pass, and over existing code when asked to cut.

## The test

A sentence stays only when a reader holding the code below it and the files it cites would miss it. Read the code and
open every cited `docs/evidence/` and `docs/rulings/` file, its Claim line first, before judging.

A sentence stays when it is:

- a trap: code that looks wrong and is right, an order that matters, a value a reader would "fix"
- where a number comes from, when the code states it as a literal: the row, the effect, the tooltip variable
- a game rule the code implements that no row, evidence file or ruling states
- a `TODO:` and what it waits for
- an `EV-` or `R-` id. Every id a file cites before the pass, it cites after

A sentence goes when it is:

- the next lines in words: "the aura's duration is the row's" over `Duration: row.Duration`
- what the cited file states: which hits roll, the rate, the cases the logs settled, counts, an example item
- what the row states and the code reads from the store; the editor hover shows the row
- what the code does not do, unless a reader would add it back
- the approach not taken: "attached to the window rather than registered as its own aura"
- what a callee does, or how a caller uses the function
- a second and third example, where one names the case
- item sources, vendors and gear comparisons in a preset comment; the database and the sims hold them
- a restatement of the paragraph above it
- history and comparisons with another game version, which `wowsims-conventions` already bars

A cut comment says less and never something else: no new fact, no stronger claim, no reworded rule. An id stays on
the clause its file attests; a cut that moves it onto a neighbouring claim makes that claim look measured. When
nothing is left, the comment goes.

## Sizes

| Comment                                  | What stays                                                      |
| ---------------------------------------- | --------------------------------------------------------------- |
| Cites an `EV-` or `R-` id                | One short sentence naming the rule, then the id, on one line    |
| Hand-written item, set or enchant effect | Name, link, tooltip, then the id line; one more line for a trap |
| Doc comment on a function or type        | One or two lines: what it is for and the rule it implements     |
| Inside a function body                   | One line; two when a trap needs its reason                      |
| Struct field, const, table row           | One trailing line                                               |
| Test                                     | One line: what the case pins and the row it reads               |

An id comment takes one of three shapes:

- A rate or a chance whose only source is a measurement gets a fixed label and nothing else:
  `// PPM measurement in EV-0130.` or `// Proc chance measurement in EV-0115.`
- Where the code under the comment shows the case (a test that casts the spell and asserts the result, a map that
  lists the spells, a name that says what the branch is), the comment points to the file by topic and does not retell
  the finding: `// Charge spending in EV-0149.` The topic is two to four words. "Moonfire spends the charge too
  (EV-0149)" over a test that casts Moonfire says nothing the test does not.
- Where the code does not show it (what the bare ids in a literal are, a trap, a limiting case), the comment is one
  whole sentence with the id in parentheses. It names its subject and gives every verb its object, so a reader holding
  only the comment and the code can tell what rolls, what is spent and on what. `// Heals that roll nothing
  (EV-0149).` has lost the proc it is about, and `// Spells the buff row does not name (EV-0149).` has lost what those
  spells do.

When unsure between the last two, point.

A comment whose ids back different facts gets one clause for each id. A trap or a rule no cited file states adds a
line to any of these sizes, and nothing else does.

A doc comment in `sim/core`, `sim/siml` or `tools/` runs longer: it keeps one sentence for each mechanic the body
implements, since class code and the generators are written against these helpers. Examples, repeats, caller notes
and what a cited file states still go. Elsewhere a doc comment runs longer only where the function is the one home of
several rules, with no evidence file, ruling or skill holding them, each rule one short sentence. A rule a second file
needs moves to a `docs/rulings/` file instead (`wowsims-evidence`).

```go
// Skullforge Reaver
// https://www.wowhead.com/forever/item=13361/skullforge-reaver
//
// Chance on hit: Drains target for 21 Shadow damage every 1 sec and transfers it to the caster.
// Lasts for 10 sec.
//
// PPM measurement in EV-0121.
```

Before the cut, five lines under the tooltip said which hand's swings roll the proc, that Windfury Totem hits do, that
a proc refreshes the debuff and that both effects crit. EV-0121 states all four.

## Left alone

- The name, Wowhead link and tooltip over an item or spell.
- A `TODO:`, a stub commented out under `panic("To be implemented")`, and table rows commented out until data comes.
- Directives: `//go:build`, `//go:generate`, `//nolint`, oxlint and `@ts-` lines.
- Doc comments on the methods of `sim/core/spelldata` types. The editor hover prints them
  (`tools/spelldata/printer/eval.go`) and `tools/spelldata/printer/eval_test.go` pins their words, so they are text
  for whoever writes class code.
- The JSDoc block over a React component and over its props. Storybook's Docs pages print them (`.storybook/main.ts`
  turns on docgen). A `//` comment in the same file takes the test.
- Comments in `proto/` and in generated files.

## A measurement with no id

A comment that tells of logs, an in-game check or an owner ruling and cites no id is not cut to fit. Search
`docs/evidence/` and `docs/rulings/` for its spell or item id. When a file holds the claim, cite it and cut to the id
line. When none does, leave the comment and report it: writing the file takes the answers `wowsims-evidence` asks for.
A count of game files rows is no measurement and takes the test like any other sentence.

## Proof

A cutting pass changes no code. Against the base it started from:

- `git diff -U0 <base> | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | grep -vE '^[+-][[:space:]]*(//|/\*|\*|$)'`
  lists the changed lines that are not whole comment lines. Each one left differs from its pair only after the `//`.
- `git diff <base> -- <file> | grep -E '^-' | grep -oE '\b(EV|R)-[0-9]{4}' | sort -u` prints no id that the same
  command over `^\+` lacks. An id the pass added is fine; one it lost is not.
- `gofmt -l sim tools` prints nothing, `go build ./...` passes, and `npx oxfmt . --check` is clean.
