# Rulings

Each file here holds one rule the code follows that no single file owns: the rule, the code that follows it, the pull
request that settled it and the evidence behind it (`docs/evidence/`). A rule one spell, aura effect, item or component
owns is a comment at its implementation instead (`AGENTS.md`). Each subject skill's `references/rulings.md` indexes
the rulings it owns, and the spelldata hover shows the file over its id.

## Ids

- An id is `R-` and four digits, and the file is `R-0042-<slug>.md`. Take the next free number.
- An id is never reused or renumbered. When a newer rule replaces one, the old file stays, gains a
  `Replaced by: R-0057` line under its title, and every citation moves to the new id.
- A comment at code that follows a ruling may cite its id after the rule: `(R-0042)`.

## A file

```markdown
# R-0042: <the rule's gist, as a title>

- Skill: <the subject skill whose index lists it>
- Code: <the files and symbols that follow the rule, in backticks>
- Pull request: <#N, where the rule was settled>
- Evidence: <EV- ids, or none>

<The rule in full, present tense.>
```

`make skills-check` fails when a cited id has no file, two files share a number, or a file nothing cites lacks a
`Replaced by` line.
