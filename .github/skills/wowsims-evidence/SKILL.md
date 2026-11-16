---
name: wowsims-evidence
description: 'Use when WoWSims Forever work records something measured or ruled: a count from a combat log or a Warcraft Logs report, a rate, a threat or damage reading, an "I checked in game", or an owner statement such as "from now on every class does X". Use it when writing or changing a docs/evidence EV- file or a docs/rulings R- file, citing an EV- or R- id in a comment, an override Source or a skill, or when a rate, coefficient or rule in the sim changes because of logs. Use it too when the build, the hotfix, the report code or what was counted is unclear, when make skills-check names an evidence or ruling id, or when an id is already taken on another branch. How comments are worded is wowsims-conventions; proc rates themselves are wowsims-procs; the hotfix files and pushes are wowsims-db-pipeline.'
---

# Evidence and rulings

A measurement lives in one `docs/evidence/EV-NNNN-<slug>.md` file per pull request write-up; a rule no single file owns
lives in one `docs/rulings/R-NNNN-<slug>.md` file. Code, override rows and skills cite the id. The formats, the three
`Evidence` states of a ruling and the id rules are in `docs/evidence/README.md` and `docs/rulings/README.md`; read them
before writing either kind of file.

## Ask before you write

An evidence file is the record other contributors trust. A field filled with a guess, or with "not recorded" when the
user could have told you, is lost for good. So collect every field below first. Any field the source does not state
goes to the user as a question, and you wait for the answer before you write the file, edit a comment, or change a
value. "not recorded" is what you write after the user says they do not know.

| Field            | Where it can be read                                                                                      | Ask when                                                                                                                                                                                                                     |
| ---------------- | --------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Date and time    | a local log's file name, the player's local time: `WoWCombatLog-111326_201455.txt` is 2026-11-13 20:14:55 | the user names a day ("last night") or gives no log                                                                                                                                                                          |
| Game build       | nowhere in a log: its header states only `BUILD_VERSION,1.60.1`                                           | always, for the full build (`1.60.1.70938`), unless the user or the pull request states it                                                                                                                                   |
| Hotfix state     | nowhere in a log                                                                                          | always: which hotfix the session ran on, by date and time (UTC), and whether the player relogged after it; a hotfix reaches a running game only after a relog                                                                |
| Sources          | a report URL gives the code; the fights come from the user                                                | a report has no fight ids, or a local log has no class and levels                                                                                                                                                            |
| What was counted | the user's description                                                                                    | the denominator is unclear: landed hits or all swings, main hand only, swings inside the buff's window, extra attacks and haste left out (a weapon's PPM leaves out Windfury, Sword Specialization, Flurry and swing resets) |
| Scope            | the user's words                                                                                          | the rule could cover more than was measured: other ranks, other classes, other abilities, the off hand ("every melee class", "from now on")                                                                                  |

The game build and the hotfix go on the `Game build` line together: `1.60.1.70938, after the 2026-11-14 02:00 UTC hotfix,
relogged`.

When the count already disagrees with the sim or with an evidence file, ask about the counting method before anything
else; a count outside the buff's window or across two seals explains most surprises.

## What goes where

| The user gives you                                            | Write                                                                                                                     |
| ------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| counts from logs or a report                                  | an EV- file; the site states the rule and cites the id                                                                    |
| "checked in game", no counts                                  | an EV- file that says so, with build, hotfix and date asked for                                                           |
| a rule wider than what was measured ("every class", "always") | a ruling: a skill body line or an R- file whose `Evidence` line cites the EV- ids or reads `pending`; ask the scope first |
| a value change in the sim                                     | the EV- file first, then the change with its `.results` in its own pull request                                           |

The counts live in the EV- file alone. A comment, an override `Reason` and a ruling state the rule and cite the id; an
override's `Source` is the id itself, or `observed` for a rate fitted from too few logs to call final, whose `Reason`
cites the id.

An owner statement is a ruling: it holds whether or not a log backs it, and the evidence file records only what was
measured. Never change sim code on an owner statement whose scope you had to guess.

## Ids

Take the next free number across `origin/master` and every open pull request:
`git ls-tree origin/master docs/evidence/ docs/rulings/` and
`gh pr list -R wowsims/forever --state open --json files --jq '.[].files[].path' | grep -oE '(EV|R)-[0-9]{4}' | sort -u`.
Two branches that take one number fail `make skills-check` once both merge; the second renumbers before it merges. An id
already on master is never renumbered. A newer measurement that replaces a whole file adds `Replaced by: EV-NNNN` under
the old file's title; one that replaces part of it adds a sentence there naming what the new id replaces, and the rest
stands.

## Red flags

These thoughts mean stop and ask first:

| Thought                                                        | Reality                                                                        |
| -------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| "The questions are not blocking, I'll write and fill in later" | Nobody fills in later. Ask, wait, then write once.                             |
| "The log says 1.60.1, that's the build"                        | That is the major version. The build number and the hotfix come from the user. |
| "Today's hotfix must be the one they mean"                     | Ask which one, and whether they relogged after it.                             |
| "'Every melee class' obviously means classes with this talent" | Ask what the rule covers before writing a ruling or changing code.             |
| "The count fits the sim, nothing to ask"                       | The build, the hotfix and what was counted still belong in the file.           |

## Before committing

`make skills-check` runs `.github/skills/check_evidence.sh`: every cited id has its file, no number is used twice, every
file is cited, and every ruling states its evidence. A new file is invisible to the check until git tracks it
(`git add -N`). The spelldata hover shows an id's file in Go, JSON, TypeScript and Markdown; restart the language server
after adding files.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. Re-derive the log header
claim from any local combat log's first line, and the check's rules from `.github/skills/check_evidence.sh`. After an edit
run `make skills-check`, then `npm run skills:install`, and commit the refreshed `skills-lock.json` with the change.
