---
name: wowsims-reforge-optimizer
description: 'Use for any work on the WoWSims Forever gear optimizer, called the "reforge optimizer" in code although nothing is reforged: it picks one static enchant per slot to maximise EP within the stat caps. Covers EP weights and how percent pseudo-stats and weapon damage are valued, stat caps, undershoot caps and soft-cap breakpoints, which slots and enchants become candidates (proc enchants, frozen slots, Enchanting, "Use enchant presets"), the LP model and the HiGHS solve on the server and in the browser worker, the optimizer result cache, and /reforgeOptimizeAsync. The batch sim and its enchant presets are `wowsims-bulk-sim`; panel layout and store wiring outside the optimizer model are `wowsims-ui`.'
argument-hint: 'Describe the optimizer bug, cap or EP behaviour, or solver task.'
---

# WoWSims Forever gear optimizer

The optimizer chooses one static enchant per listed slot to maximise EP within the stat caps, by solving a
binary LP with HiGHS. The UI decides which slots and enchants are on offer; Go builds the model, solves it and
returns the gear. It is a single-gear solve: the batch sim never runs it.

## Where things live

- `sim/core/reforge_optimizer/`:
    - `optimizer.go`: `OptimizeAsync`, optimizer state, baseline build.
    - `enchants.go`: candidates, per-slot binaries, tie-break, pseudo-stat and weapon damage valuation.
    - `model.go`: the split between objective coefficients (EP) and cap coefficients (stat dependency graph).
    - `solver.go`: the solve and the cap-refinement loop (`solveModel`, `checkCaps`, `mipRelGap`).
    - `caps.go`: soft caps as gap-to-cap breakpoints, `checkWeights`.
    - `reforge_stats.go`: unit stat math, `resolveStatDelta`, rating/percent conversions.
    - `lp.go`: deterministic CPLEX LP text (`modelToLPFormat`, `appendLPTerm`).
    - `gear.go`: frozen slots, applying the picked enchants. `logging.go`: debug-only request dumps.
    - `highswasm.go` (native) and `highs_js.go` (browser): the HiGHS bridges.
- Endpoint: `/reforgeOptimizeAsync` in `sim/web/main.go`, which also calls `WarmUp()` at start so the first solve
  does not pay the module compile; the WASM sim exports `reforgeOptimize` and `reforgeOptimizeAsync`.
- UI model: `ui/features/reforge/model/reforge_optimizer.ts` (`createReforgeOptimizer`, the solve, the cache),
  `enchant_options.ts` (slot classification, candidate lists), `utils.ts` (soft-cap and breakpoint helpers); the
  panel is `ui/features/reforge/components/ReforgePanel/`. Settings: `ui/sim/settings/reforge_settings.ts`
  (`ReforgeSettings` proto). The request is built by `Sim.reforgeOptimize` in `ui/sim/sim.ts` and
  `ui/sim/state/reforge_request.ts`. Spec hooks (`getEPDefaults`, `updateSoftCaps`, ...) are described in
  `ui/README.md`.

## Candidates: the UI side

`classifyEnchantSlots` and `buildEnchantOptions` turn each equipped, unfrozen enchantable slot into either a
skipped slot or a `BulkEnchantSlotOptions` entry in `ReforgeOptimizeRequest.enchant_options`. The panel's skipped
list and the solve share one classification, rebuilt when the gear, Enchanting, the frozen slots or the presets
change.

- By default a slot lists the enchants the player can use on that item (class and Enchanting gated, fit by
  `enchantAppliesToItem`, the same rule as `core.EnchantAppliesToItem`) that are static (`proc_kind` None, with
  stats, pseudo-stats or weapon damage) and worth EP > 0 on the pre-cap weights.
- A slot whose item can take a proc enchant the sim implements is left to the batch sim and listed in the panel's
  skipped section with a link to it: a combat proc for every spec, a survival proc for tanks. Every melee weapon
  takes Crusader, so a melee weapon's slot is always skipped. A proc in `MISSING_ENCHANT_EFFECTS`
  (`ui/sim/constants/missing_effects_auto_gen.ts`) is not implemented, so it blocks nothing.
- "Use enchant presets" (`ReforgeSettings.use_enchant_presets`, off by default; only the UI reads it) draws the
  candidates from the batch sim's checked presets instead (`enchantPresetPool`; see `wowsims-bulk-sim`). A slot
  whose item takes enchants but none the presets list is skipped as unlisted; a listed proc the sim implements,
  combat or survival, skips the slot for every spec; otherwise the candidates are exactly the listed static
  enchants, with no EP filter. With the option on and no preset checked, Optimize Enchants is disabled.
- `proc_kind` comes from gen_db (`tools/database/dbc/enchant.go`, overrides in `enchantProcKindOverrides`).

## Candidates and baseline: the Go side

- `newReforgeOptimizer` clones the request and merges `player.database` into the core database before it
  resolves anything. Go drops only candidates that are not static, do not fit the item, or sit in a frozen slot;
  the slot's input enchant does not matter, so a listed slot wearing a proc is optimised from its options.
- The baseline is the input gear with each candidate slot's enchant cleared when that enchant is a static row
  fitting the item. Any other enchant (Minor Speed, a proc, one with no fitting row) stays, so a slot with no pick
  keeps it; a slot with no pick whose input enchant was static ends empty. One `ComputeStatsAndDeps` on the
  baseline gives the base stats and the stat dependency manager.
- With no candidates the optimizer returns the input gear unchanged, score 0, 0 passes, without building an
  environment (the UI skips the request entirely in that case).
- The item database trap: the server and WASM builds carry no `with_db` tag, so they know only the items and
  enchants the request ships. A candidate enchant or a slot's item the request did not ship is dropped silently,
  although a `go test -tags with_db` passes. `Sim.reforgeOptimize` builds the database from the gear and every
  candidate enchant id (`makeBulkGearDatabase`), then calls `addPartyBuffItems`, because it replaces the player's
  database rather than merging into it; keep that call on any rebuilt request database.

## The model and the solve

- One binary per (slot, candidate) and a per-slot row `<= 1`, so a slot may end with no pick when every candidate
  breaks a ceiling.
- Objective and caps live in separate coefficient spaces (`model.go`). The objective is EP: stats through
  `applyReforgeStat`; percent pseudo-stats as `Stats.computeEP` (`ui/sim/proto/stats.ts`) values them, by the
  percent's own weight or else as rating under the root rating's weight, with the part a rating's three hit, crit
  or haste channels share counted once and a rating's fallbacks counted once, as the largest; flat weapon damage as
  main-hand, off-hand or ranged DPS at the weapon's speed. The cap coefficients push the raw stat delta through
  the baseline's stat dependency manager (`resolveStatDelta`), so Agility counts toward Crit% and Dodge%.
- A school hit weight, pre-cap or post-cap, is per hit rating point, as the stat weights measure it, while every
  school hit value the model counts is in percent. `preCapWeights` and `computeReforgeSoftCaps` multiply those
  weights by `HitRatingPerHitPercent` (`weightScale`) before anything reads them, so `checkWeights` must run after
  them: the weight it moves from hit rating onto a capped school hit is already per percent.
- `enchantTieBreak` (twice for the slot's input enchant, none where the input enchant stays in the baseline) keeps
  the input enchant on an EP tie and fills a slot whose candidates are all worth 0. The solve passes an explicit
  `mipRelGap` far below HiGHS's default, which would otherwise swamp the tie-break.
- The score is the objective over the baseline without the tie-break, not the gain over the input gear.
- `checkCaps` sums the selected variables' cap coefficients; no sim runs between passes. Per pass: an exceeded
  hard cap gets a row, a ceiling `<=` when `UndershootCaps` names the stat, else `>=` at the cap with the stat's
  weight zeroed. A ceiling the baseline already exceeds allows none of that stat. If no hard cap fired, soft caps
  are consumed from the front, one per pass: a passed breakpoint adds a `>=` row and swaps in its post-cap EP;
  threshold caps drop after one pass. The loop re-solves until nothing fires and reports `PassesDone`.
- The cap coefficients must match what the character sheet would show; `TestCapSpaceDeltaMatchesTheSheet` in
  `sheet_delta_test.go` checks them against `core.ComputeStats`. A change to stat conversions or cap math belongs
  there first.
- On a HiGHS error, infeasible caps or a timeout without a solution, return an error; do not fall back to a
  different solve. Verbose logging stays behind `ReforgeOptimizeRequest.debug`.

## HiGHS

- The LP text is byte-exact and stable: variables are `x0…xn` in insertion order and every term goes through
  `appendLPTerm`. HiGHS breaks ties between equal-objective solutions by that text, so a layout change can change
  results.
- Native (`highswasm.go`, `//go:build !(js && wasm)`): `ui/worker/highs.wasm`, embedded by
  `ui/worker/highs_wasm_embed.go`, runs in pooled wazero runtimes; `WOWSIMS_HIGHS_WASM_RUNTIME_CONCURRENCY`
  overrides the pool size. Keep the file name: a `_wasm.go` suffix would make it a GOARCH=wasm-only file.
- Browser (`highs_js.go`): `ui/worker/sim_worker.ts` loads the npm `highs` package and exposes
  `globalThis.__wowsimsSolveHiGHSLP(lp, timeoutSeconds, mipRelGap)`, which the Go WASM calls.
- Both must run the same HiGHS build: `vite.build-workers.mts` and `highs_version_test.go` fail when the committed
  `highs.wasm` differs from the installed package. After changing the pinned `highs` version, run
  `make update-highs` and commit the result.

## The result cache

The UI caches optimized gear in IndexedDB for 14 days (`ui/sim/cache/reforge_cache.ts`, `ReforgeGearCache`). The
key is the gear key (`getReforgeCacheGearKey`: the gear identity plus the frozen slots) and a config hash over
three parts (`getReforgeConfigHash`): the solve-relevant player proto (`cacheRelevantPlayerProto`), the raid
buffs, party buffs and debuffs, and the optimizer config (`makeReforgeConfigRequestFields`, trimmed by
`cacheRelevantReforgeRequest`). A new field that changes a solve's output must reach one of these, or stale gear
is served silently; a field that does not must stay out, or every user's cache is invalidated by unrelated changes.

## Validation

```bash
go test -tags with_db -count=1 ./sim/core/reforge_optimizer ./sim/web
go test -tags with_db -count=1 -run TestEnchantProcKind ./tools/database
npx vitest run ui/features/reforge ui/features/enchant-presets
npm run type-check
```

The optimizer's Go tests use the class-free `sim/core/testagent` (registered as a protection warrior in
`sheet_delta_test.go`, with that spec's gear sets from `ui/specs`); the ones that build gear need the `with_db`
tag. For a change to
what the request carries, run the optimizer against the dev server as well.
