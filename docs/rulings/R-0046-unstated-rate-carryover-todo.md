# R-0046: A rate the game files do not state takes the carry-over value with a TODO

- Skill: wowsims-procs
- Code: `tools/database/gen_effects.go`, `rateOrigin`
- Pull request: #322
- Evidence: none

A rate the game files do not state takes the carry-over value (`vanilla-carryover`) with a TODO, or is written and
commented out waiting for a procs-per-minute rate, and the gear tab shows an in-game validation notice for every
unvalidated rate, driven by the override tables' `Source`; a measured rate's `Source` is its evidence id.
