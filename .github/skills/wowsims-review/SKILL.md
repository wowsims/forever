---
name: wowsims-review
description: 'Use when reviewing any WoWSims Forever change, in sim/, ui/, tools/ or docs: your own branch before a commit or a pull request, or someone else''s pull request. Use it whenever a diff, branch or PR is to be checked or audited, even when the request does not say "review". A fixed pass: test-only production code, stale comments, tests that pin data, unworn item reads and orphans on every change; for spell, talent, aura, dot and proc code, every number and hand-built config against its game files row and tooltip. Each finding is verified and comes with a drop-in fix. Also covers the SIML name table and rotations such a change moves, and a change to the SIML grammar. How text is written is `wowsims-conventions`.'
argument-hint: 'Name the branch, the pull request number or the files to review.'
---

# Reviewing a change

The pass is the same on your own branch before a commit and on someone else's pull request. The author of a
change reads it the way they meant it, so a self-review works best handed to a fresh agent with the diff and
this skill; the author then verifies each finding that comes back the way step 7 says, and fixes it.

Every change gets steps 1, 2, 5, 7 and 8. Spell, talent, aura, dot and proc code under `sim/`, with the
overrides and class ladders it reads, also gets steps 3 and 4, and a change that moves the SIML name table, a rotation
or gear set under `ui/specs`, or the SIML language gets step 6. A UI change also runs the gates in
`.github/skills/wowsims-ui/references/verification.md`; generator internals are `wowsims-db-pipeline`'s subject.

`references/defects.md` is the catalog for the spell steps: each defect class, with a case a review of a port found
and the fix that reads the row. `references/rulings.md` holds the review's own rulings on orphans and kept helpers.

## 1. Inventory

```
git diff origin/master...HEAD --stat                 # what the change touches
git diff origin/master...HEAD -- <path>              # read every hunk, not the stat
gh pr diff <number> -R wowsims/forever               # someone else's PR; -R, since gh may default to an upstream remote
```

List every spell id, ladder and talent the diff registers or changes, every number literal it adds, and every
export, prop or parameter it adds. Read the PR body and the commit messages as claims to check, not as a
summary: "the generator refuses the placeholders once the game files state them" was written above five Go
constants that no generator reads.

## 2. Every change

- **Production code that only a test or story uses** (the rule is in `AGENTS.md`). For each new export, prop,
  parameter, branch or `index.ts` re-export, find a caller outside tests and stories; with none, the test changes
  instead (a fixture in `ui/testing/` or the test file, a stub at the module boundary).
- **A comment that claims what the code no longer does.** A comment naming a read, a row, a file or a
  dependency is checked against the code after the diff; one the change made untrue is a finding.
- **An item the request does not carry.** Check that a read of an item the player may not wear has the request
  carry it (`wowsims-consumables-items` `## The request database`) and fails loudly when it is missing, as
  `driveDarkspear` in `sim/core/buffs/drivers.go` panics.
- **Orphans the change made.** A function with no caller, a mask naming a spell nothing registers, an empty
  body the change left behind: remove them in the same change.
- **A live talent never becomes a silent no-op.** An unported talent or ability keeps its stub: the implementation
  commented out in place under `panic("To be implemented")` and a `// TODO:` naming what is missing, never a deletion
  or a silent return. Content Forever does not have is removed, not stubbed, once `tools/database/wowsims.db` shows no
  data for it. An aura registered with no effect, or a cooldown that does nothing, is a finding.
- **One copy of each rule.** Three channels differing only in their tick effect share one builder; a value
  built each iteration that never changes is built at registration.
- **Text.** Every comment, doc line, test message, commit message and PR text the change adds or rewrites is
  read against `wowsims-conventions`, its plain-prose pass included.

## 3. Each spell against its row

```
go run ./tools/spelldata <id>                     # header, effects, the ladder call that reaches it
go run ./tools/spelldata -config 'spelldata.SpellConfig(&<class>.Unit, <rank>, ...)' -package <class>   # what the resolver builds
sqlite3 tools/database/wowsims.db "select Description_lang, AuraDescription_lang from Spell where ID=<id>"   # the tooltip
sqlite3 tools/database/wowsims.db "select Attributes from SpellMisc where SpellID=<id>"   # every attribute word
SPELLDATA_REPORT=1 go test --tags=with_db ./sim/<class>/ -run <Test> -count=1 -v   # the effects the parse skipped; -v, or a passing run prints nothing
```

`go run ./tools/spelldata` prints only the attributes `sim/core/spelldata/attributes.go` reads; the
`SpellMisc` query shows the rest, such as "not usable in combat" (word 0, `0x10000000`). The Wowhead Forever
tooltip shows the values scaled to level, which the `$s1`-style tooltip text does not. Read the subject skill's
`references/` folder (`rulings.md`, or the traps file where there is none; `wowsims-spells` names the subject skills),
before judging a number the row does not decide.

Then make three comparisons for each spell:

1. **Every literal against the row.** A number the row states is a finding: the fix reads the accessor. A
   number the row does not state is an override in `tools/database/overrides/spell_overrides.go` where one of
   its kinds fits (an attack power coefficient is `APCoefDirect` or `APCoefPeriodic`), so the generator refuses
   it once the game files catch up; otherwise a literal with its reason and a `// TODO:` on the line.
2. **Every hand-built piece against the resolver.** A `core.SpellConfig`, aura, dot, `OnTick` or proc trigger
   written by hand is compared with what `SpellConfig`, `AuraConfig`, `DotConfig`, `ParseEffects` or
   `ProcTrigger` build from the row. Where they agree, the hand code goes. Where they differ, the tooltip
   decides which is right, and the decision is written at the call site.
3. **Every effect index against the tooltip.** The tooltip's `$s2` and `$m2` name effect 2. Code that reads
   another effect because its number happens to match is a finding even when nothing moves today.

## 4. Rules the rows cannot show

The first seven have a section in `references/defects.md`.

- **Core's order.** Check whether an `OnCastComplete` handler sees the cast that procced it, proven at runtime;
  the mechanic is in `wowsims-core-mechanics` `## Cast order`.
- **What a cast spends, and when.** Check combo points, energy and the miss refund one by one, and that a
  one-charge buff goes with the cast it modifies; `wowsims-core-mechanics` `## What a cast spends`.
- **What scales.** Check a tick against stacks and combo points, `PointsPerResource` and a flat weapon bonus;
  `wowsims-core-mechanics` `## What scales`.
- **How many targets.** Check that an area effect hits every enemy and an "up to" count is the total;
  `wowsims-core-mechanics` `## How many targets`.
- **What the row forbids.** Check that `ExtraCastCondition` restates no form, stance, weapon, shield or combat
  requirement; `wowsims-core-mechanics` `## What the row forbids`.
- **Hooks.** Check each hook is attached once, at registration, on every path that needs it;
  `wowsims-core-mechanics` `## Hooks`.
- **Whose state.** Check a multi-target effect reads each hit target, a tag counts only this caster's auras and a
  missing map key is not 0; `wowsims-core-mechanics` `## Whose state`.
- **What a preset line sees.** Check that a line reading something the running cast's landing changes
  carries `not(Spell Is Casting(X))` (`wowsims-apl` `## Timing`), and that its cast counts were judged as
  `wowsims-apl` `## Presets` says.

## 5. Tests

- **Logic, not data.** Check each test the change adds or edits against `wowsims-testing` `## What a test asserts`.
- **The class suites run** with no `t.Skip`, and a Forever talent string and gear set.
- **The suite can see the change.** Check its `.results` or a runtime print carries the spell the change touches;
  `wowsims-testing` `## .results discipline`.
- **The cast counts are possible.** Check them against the fight length and each spell's cooldown;
  `wowsims-testing` `## .results discipline`.
- **Each moved .results file has its own cause.** Check each move was isolated; `wowsims-testing`
  `## .results discipline`.

## 6. SIML

`sim/siml/SPEC.md` is the language. A spell change usually moves two kinds of SIML file.

**The name table, `sim/siml/names.json`.** A regeneration only adds names; any other change to a name is a repin,
which adds a line to `sim/siml/gen/repinned.txt`, and the pull request has to say it intends each one. How the table
and repins work is in `wowsims-spells`, "Names in SIML files". Check it with:

```
go run ./sim/siml/cmd/simlnames -base "$(git merge-base origin/master HEAD)"   # passes unless a name changed without its record line
git diff "$(git merge-base origin/master HEAD)" -- sim/siml/gen/repinned.txt   # the repins the change makes; read each one
```

**Rotations and gear sets, `ui/specs/**/*.siml`.** Their diff reads as SIML rows (`cast mortal_strike if …`,
`main_hand arcanite_reaper + weapon_crusader`), so read each changed row for what it does. Check that each name means
the thing intended: the table gives its id (`grep -n '"mortal_strike"' sim/siml/names.json`, one line per class and
namespace), and for a spell `go run ./tools/spelldata <id>` gives its row. The repository check,
`go run --tags=with_db ./cmd/wowsimcli siml fmt -check ui/specs`, fails on a spell or aura the file's spec does not
register, and on an enchant that does not fit its slot.

**A change to the protos.** SIML writes every APL kind and setup field from the protos (`SPEC.md` §8), so a new
action, value or field needs no SIML edit, and a change that edits `sim/siml` only to add one is a finding. What a
proto change does need: the `wowsims-siml` skill written again (`make siml`; `TestLLMTextIsCurrent` fails without it),
and the `.siml` files under `ui/specs` that use a kind or field the change takes out or renames, which the
repository check names, moved to the new form. A new value kind that reads a spell or an aura is named `<head>_<prop>`, with one
`ActionID` field and at most one `UnitReference`, so it takes the accessor form (`proto/apl.proto` says so at the top);
a finding otherwise.

**The language.** A change to the grammar, the typing or the printer states the rule in `SPEC.md` and adds a corpus
case under `sim/siml/testdata/` (`parse/`, `print/`, `format/` or `errors/`), which `sim/siml/conformance_test.go`
runs. A change with no case is a finding. So is a change that gives existing text another meaning without a new `siml`
major version (`SPEC.md` §14). A grammar change also updates the editor's partial parser, `sim/siml/lang_context.go`,
which reads a statement up to the cursor for completion; `go test --tags=with_db ./sim/siml/...` runs its cases
(`lang_db_test.go`).

```
go test ./sim/siml/...                     # the corpus; no database needed
go test --tags=with_db ./sim/siml/...      # also the test that the table keeps its entries
go test --tags=with_db ./cmd/wowsimcli/... # also the repository check of ui/specs
```

## 7. Verify every finding

A finding is reported only with evidence of at least one of these kinds, named in the finding:

- the row or tooltip line (`go run ./tools/spelldata <id>`, the `Spell` or `SpellMisc` query);
- the code path, as `file:line`;
- the numbers in a .results file;
- a value printed at runtime from a throwaway test.

A finding with none is dropped, and so is one the evidence contradicts; list it under "checked, not
findings" instead. A finding about when something happens (an order, a consumption, a refund) needs the
runtime print: reading the call order is how a withdrawn finding got posted. A suggested fix is applied in a
worktree and passes `go vet --tags=with_db ./sim/<class>/` (or the UI gates), and the number it claims is
printed at runtime before it is posted ("Checked: `maxEnergy=110` at rank 2"). Findings and their fixes are checked
on the PR head merged with origin/master, comparing `.results.tmp` at the merge with and without the fixes, since a
master database update can change a row with no textual conflict.

## 8. Report

One entry per finding, most severe first: a wrong number, then missing behaviour, then hand code the store
already states, then hygiene.

- `file:line`, and the defect in one bold sentence.
- The failure: concrete input and wrong output, with the numbers ("a 5-point Rupture ticks for 35 instead of
  58.65 before attack power").
- The fix, as code: a fenced block that drops in (the whole function or the changed block), not a
  description of one. For a spell, it reads the row through the store's accessors and resolvers.
- What was checked, and how.

Every identifier, expression, path and tooltip formula goes in backticks: a bare `weapon*1.55+80, weapon*3.5`
renders as italics and loses both asterisks. A comment posted from the shell takes its body from a file
(`gh api ... -F body=@file`), since backticks inside a quoted argument are command substitution; read the
posted comment back as HTML (`-H "Accept: application/vnd.github.html+json" --jq .body_html`) and check the
fences closed and no `<em>` stands where an asterisk was meant.

After the findings, two lists: **checked, not findings** (what was examined and holds), and **hand numbers
that stay**, each with its reason. On a self-review, fix the findings, rerun the gates, and put both lists in
the PR body.

The review ends with the `humanizer` skill: run it over the review text, the comments to post, and any PR body,
commit message or code comment the review added or rewrote, before anything is posted or pushed. When the
harness does not know the skill, run `npm run skills:install` first; it installs the skill for the agents on
this machine. The rewrite changes prose only; findings, numbers, code and paths stay.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The "You changed →
Update" table and the path check are in `references/maintenance.md`.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill or its `references/rulings.md` otherwise. A memory file is a pointer
to this skill, not a second copy.
