# Evidence

Each file here holds measurements the code relies on: what was measured, where, and the numbers. Every measured item
or claim has one home, the first file that measured it, and its numbers live there alone; a table lives in one file
only, and a file that needs another's table cites its id. A measurement no pull request wrote up keeps a file of its
own. Code and skills cite a file by its id, and the spelldata hover shows the file over the id. These files are the
record of a measurement; an issue or a research note that tracks the same numbers links the id.

## Ids

- An id is `EV-` and four digits, and the file is `EV-0042-<slug>.md`. A new file takes the next free number, and
  only for items or claims no file holds yet.
- An id is never reused or renumbered.
- A new count of items or claims a file already holds rewrites that file: their rows and sections take the newest
  numbers in place, under the same id, and none of their older numbers stays in that file or in any other. A corrected
  reading of a claim is rewritten in its file the same way. No file says that another replaces part of it.
- A file whose every part has moved to other files keeps only its title and a `Replaced by: EV-0057` line under it,
  with no data, and every citation moves to the file that holds each part.
- Every citation names the home: an override row's `Source` and `Reason`, a comment, a skill, a ruling, another
  evidence file and issue #964.
- A comment cites the id after the rule it backs: `// A Pyroblast consumes every stack (EV-0005).` The rule stays in
  the comment; the counts, logs and method move here.
- An override row's `Source` is the bare id, or several separated by commas when two passes back the row:
  `Source: "EV-0003"`, `Source: "EV-0017, EV-0018"`. The generator reads such a Source as a measured rate. A rate fitted
  from too few logs to call final has the Source `observed`, and its `Reason` cites the id.

## What counts as evidence

A number a larger sample could change while the code stays the same: proc counts, rates, threat readings, hit counts,
"checked in game". A value the code implements stays in the code. These are not evidence and keep their own text:
a TODO to measure something, a value read from the game files, and the `Source` values `tooltip`, `owner decision`,
`Spell <id>`, `observed`, `vanilla-carryover` and `tbc-carryover`, and the other non-measurement `Source` values the
tables use (`game files`, `game-files-rows`, `Spell <id> description`, `owner decision, PR #99`).

## A file

```markdown
# EV-0042: <what was measured, as a title>

- Claim: <the rule the measurement settles, one sentence>
- Sources: <Warcraft Logs report codes with fight ids, or "combat log" with class, levels and date>
- Game build: <the full build and the hotfix the logs ran on: "1.60.1.70938, after the 2026-11-14 02:00 UTC hotfix, relogged">
- Date: <when the logs were taken>
- Pull request: <#N, where the measurement was written up; the commit hash beside it when the reading is in a commit message>

<The method: what was counted, what was left out, and why.>

<The numbers, as a table where there are several.>

## Count of the <first> to <last> reports

- Sources: <the count's reports, or where each of its sections lists them>
- Game build: <as above, for this count>
- Date: <when this count's logs were taken>
- Pull request: <#N, where this count was written up>

<The method, or the file and section whose method this count runs again.>
```

The header's Sources, Game build, Date and Pull request describe the count the file was written with. A later count
that rewrites some of its rows adds a count section like the one above at the end of the file, and each item it
rewrote names that section with a `- Count: §Count of the <first> to <last> reports` line under the heading of the
item's section; the item's row in a summary table moves to a table in the count section. The header's lines then say
that the items naming a count section have their own, and the Claim covers every item. A count that rewrites every row
of a file rewrites the header instead.

A file names logs by report code and fight, and a local log by class, levels, game build and date. It names no
character. Write only what the measurement recorded, taken from the text of the sites that cite it, an override row's
`Source` and `Reason`, the pull request's body and the commit messages on its branch: a field none of these records
reads "not recorded".

`make skills-check` fails when a cited id has no file, two files share a number, or a file nothing cites lacks a
`Replaced by` line.
