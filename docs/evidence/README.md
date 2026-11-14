# Evidence

Each file here holds one pull request's write-up of a measurement the code relies on: what was measured, where, and the
numbers. Several log sets one pull request wrote up are sections of its file, and a table lives in one file only; a
file that needs another's table cites its id. A measurement no pull request wrote up keeps a file of its own. Code and
skills cite a file by its id, and the spelldata hover shows the file over the id. These files are the record of a
measurement; an issue or a research note that tracks the same numbers links the id.

## Ids

- An id is `EV-` and four digits, and the file is `EV-0042-<slug>.md`. Take the next free number.
- An id is never reused or renumbered. When a newer measurement replaces one, the old file stays, gains a
  `Replaced by: EV-0057` line under its title, and every citation moves to the new id.
- A comment cites the id after the rule it backs: `// A Pyroblast consumes every stack (EV-0005).` The rule stays in
  the comment; the counts, logs and method move here.
- An override row's `Source` is the bare id, or several separated by commas when two passes back the row:
  `Source: "EV-0003"`, `Source: "EV-0017, EV-0018"`. The generator reads such a Source as a measured rate.

## What counts as evidence

A number a larger sample could change while the code stays the same: proc counts, rates, threat readings, hit counts,
"checked in game". A value the code implements stays in the code. These are not evidence and keep their own text:
a TODO to measure something, a value read from the game files, and the `Source` values `tooltip`, `owner decision`,
`Spell <id>`, `estimate`, `vanilla-carryover` and `tbc-carryover`, and the other non-measurement `Source` values the
tables use (`game files`, `game-files-rows`, `Spell <id> description`, `owner decision, PR #99`).

## A file

```markdown
# EV-0042: <what was measured, as a title>

- Claim: <the rule the measurement settles, one sentence>
- Sources: <Warcraft Logs report codes with fight ids, or "combat log" with class, levels and date>
- Game build: <1.60.1.x, when known>
- Date: <when the logs were taken>
- Pull request: <#N, where the measurement was written up; the commit hash beside it when the reading is in a commit message>

<The method: what was counted, what was left out, and why.>

<The numbers, as a table where there are several.>
```

A file names logs by report code and fight, and a local log by class, levels, game build and date. It names no
character. Write only what the measurement recorded, taken from the text of the sites that cite it, an override row's
`Source` and `Reason`, the pull request's body and the commit messages on its branch: a field none of these records
reads "not recorded".

`make skills-check` fails when a cited id has no file, two files share a number, or a file nothing cites lacks a
`Replaced by` line.
