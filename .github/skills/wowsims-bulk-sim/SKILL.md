---
name: wowsims-bulk-sim
description: 'Use for any WoWSims Forever bulk sim ("Batch" tab) work, whether it runs natively, on the server or in the browser: candidate generation from BulkSettings and the enchant presets, the low/medium/high and finalist stages, survivor culling, content-derived seeds, paired errors and tie groups in the results, progress, abort, combination counts, and the item database a bulk request carries to the workers. Also use it when a change to the sim request or its SimDatabase could reach a batch run. The gear optimizer (EP, stat caps, HiGHS) is `wowsims-reforge-optimizer`; React and store work outside the bulk feature is `wowsims-ui`.'
argument-hint: 'Describe the bulk sim flow, stage or result behaviour, or validation task.'
---

# WoWSims Forever bulk sim

The bulk sim ranks gear sets built from the worn gear plus the items and enchants the user selects. One Go
implementation (`sim/core/bulk`) generates the candidates and runs the stages; the browser runs a TypeScript
mirror of the stage loop because the WASM sim cannot run `bulkSimAsync`.

## Where things live

- Protos: `proto/api.proto` (`BulkSimRequest`, `BulkSettings`, `BulkGearCandidate`, `BulkSimResult`,
  `BulkGearResult`, `BulkSimStageMetrics`, `BulkCombinationCountRequest`, `BulkCandidatesRequest`). `BulkSimStage`
  runs Low, Medium, High, Finalist, Complete; Error ends a failed run.
- Candidate generator, `sim/core/bulk`: `generator.go` (selected items, grouped slot pairs, combination table),
  `weapons.go` (main-hand/off-hand combos, imbue adjustment), `enchants.go` (enchant variants), `slots.go` (bulk
  slots), `item_keys.go` and `item_rules.go` (dedupe, equip rules), `candidates.go` (the `BulkCombinationCount`,
  `BulkCandidates` and `EnsureBulkSimCandidatesGenerated` entry points).
- Stage runner, `sim/core/bulk`: `bulk_sim.go` (`BulkSimAsync`, constants), `stage.go` (stage configs, adaptive
  passes, finalist stage), `statistics.go` (target iterations, culling, paired errors, `BulkSimZ95`), `batch.go`,
  `carry_over.go`, `merge.go`, `progress.go`, `estimate.go`.
- Server: `sim/web/bulk.go` generates the candidates when the request carries none, then hands over to
  `bulk.BulkSimAsync`; `sim/web/main.go` wires `/bulkSimAsync`, `/bulkCombinationCount` and `/bulkCandidates`.
  `sim/wasm/main.go` exports only `bulkCombinationCount` and `bulkCandidates`.
- Request building: `Sim.runBulkSim`, `getBulkCombinationCount` and `getBulkCandidates` in `ui/sim/sim.ts`;
  `ui/sim/bulk/gear_database.ts` (the request's SimDatabase); `ui/sim/bulk/utils.ts` (`dedupeGearSets`, stage
  helpers).
- Browser stage loop: `ui/sim/wasm/bulk_sim/` (`index`, `stage`, `batch`, `merge`, `statistics`, `progress`,
  `carry_over`, `estimate`, `types`), a line-for-line mirror of `sim/core/bulk`.
- Feature: `ui/features/bulk/model/run.ts` (`runBulkBatch`, cancel, combination refresh), `core_sim.ts`,
  `tie_chains.ts`, `settings.ts`; components under `ui/features/bulk/components/`; the tab is
  `ui/app/tabs/BulkTabBody.tsx`. `ui/sim/settings/bulk_settings.ts` stores the settings and the enchant presets.
- Enchant presets: `ui/features/enchant-presets/` (model, `EnchantPresets` tiles and dialog).
- Generated, never hand-edited: `ui/sim/bulk/constants_auto_gen.ts` (slot maps) and
  `ui/sim/wasm/bulk_sim/constants_auto_gen.ts` (tuning constants, stage ladder, `BULK_SIM_Z95`), written by
  `make go-to-ts` through `tools/database/gen_bulksim_constants.ts.go` from the exported consts in `sim/core/bulk`.

## Native and browser paths

`Sim.runBulkSim` picks the path by `isWasm()`:

- **Native or server:** the request goes out with no candidates; Go builds them in process and runs every stage.
  Nothing is emitted while the candidates are built, which can take minutes, so `sim/web/bulk.go` registers the
  abort signal before generating.
- **Browser (WASM workers):** the page asks the workers for the candidates (`getBulkCandidates`), dedupes them
  against the baseline (`dedupeGearSets`, keyed by `getGearIdentityKey`) and runs the stage loop in
  `ui/sim/wasm/bulk_sim`, one `raidSimAsync` per candidate segment. The WASM worker cannot run `bulkSimAsync`.

Because both paths run the same algorithm, a change to staging, culling, statistics, carry-over or progress in
`sim/core/bulk` needs the same change in `ui/sim/wasm/bulk_sim`, and a new tuning constant goes through
`make go-to-ts` rather than being copied by hand.

Both paths build each candidate's request from the base request: the candidate's equipment, imbues adjusted to
its weapons, `saveAllValues` on.

## The request database

Only `with_db` builds load every item (`sim/core/database_load.go`). The server binary (`go build ./sim/web`) and
the WASM build do not use the tag, so they know only the items in the request's `SimDatabase`, while
`go test -tags with_db` sees every item. A green Go test therefore does not prove the request carries what it
needs: the generator fails with "selected bulk item … is missing from the database" for an item the request did
not ship, and core code reading an item the player does not wear gets nil.

- The bulk database is built from the baseline gear, `BulkSettings.items` (a selected spec's own enchant and
  suffix included) and the enchant pool (`makeBulkItemDatabaseFromSpecs`), on both paths, never from the
  candidates. An id a candidate can carry reaches the workers only through `BulkSettings`.
- It is merged into the player's database (`Database.mergeSimDatabases`), never assigned over it. The player's
  database already holds the items party buffs read (`addPartyBuffItems` in `ui/sim/proto/buff_items.ts`, e.g. the
  party Darkspear's item); replacing it drops them silently. Any code path that rebuilds `player.database` either
  merges into it or calls `addPartyBuffItems` afterwards, as `runSim` with `options.gear` does.

## Candidates

- The baseline is `base_request.raid.parties[0].players[0].equipment`; `BulkGearCandidate.index` is a candidate's
  stable identity through every stage.
- `rawCombinations` is the mixed-radix index space of the combination table; `combinations` is the runnable count
  the UI shows. There is no set-bonus filter, so the two are equal for `BulkCombinationCount`.
- Slot eligibility goes through `core.EligibleSlotsForItem` (shared with item swap), spec lookup through
  `core.PlayerProtoToSpecSafe`, request validation through `newGeneratorFromRequest`.
- Hand swap (`initSelectedItems`): with no weapon selected, the worn weapons stay in the hands they are worn in.
  Once any weapon is selected, a dual-wielder's worn weapons join the hand-weapon pool and every allowed
  main-hand/off-hand order is a candidate.
- Game rules the generator encodes: no reforging, gems or sockets (an `ItemSpec` is id, random suffix and
  enchant); no Titan's Grip, so a two-hander never shares the hands with an off-hand; the Ranged slot iterates like
  any non-weapon slot.
- A ring enchant needs Enchanting. For a player without it, `slotSpec` strips both ring slots' enchants, the
  baseline's and a selected ring's own, so no two candidates differ only by a ring enchant.
- The combination table resolves each choice's item once per physical slot, and a candidate's item specs share
  one backing array; keep per-candidate allocations out of that loop.

## Enchant presets

The bulk sim crosses enchants through its checked enchant presets. `enchantPresetPool` turns them into
`BulkSettings.enchants` on every request (Finger 1 stands for both rings; Main Hand enchants go to both hands);
the pool is derived, never stored, and sorted so preset order does not change the seed. The presets live in the
bulk slice (`enchantPresets`, storage key `bulk-enchant-presets.v1`); checking a preset, or saving or deleting a
checked one, bumps the slice's `settings` counter so the combination count refreshes.

The gear optimizer's "Use enchant presets" option reads the same presets. That is the only thing `BulkSettings`
and `ReforgeSettings` share; keep the two apart in request shaping and serialization. The fit rule
(`core.EnchantAppliesToItem`, `core.FittingEnchantRow`) is shared with the optimizer too.

## Stages and statistics

- Every candidate runs the same seed sequence, so comparisons use paired errors over `AllValues`
  (`bulkSimPairedDpsError`), far tighter than per-result stdev.
- Iterations come from the stage configs (minimums raised by the error target), not from `SimOptions.iterations`.
  Below `BulkSimMinCombinations` candidates, or when it would sim fewer iterations, the run goes straight to the
  high stage.
- Low and medium stages cull to survivors. Culling intervals use the run's original candidate count, iteration
  targets the current stage's count. A stage that misses its target error gets up to `BulkSimMaxAdaptivePasses`
  extra segments with offset seeds, merged into the existing metrics, and the carry-over lets the next stage sim
  only the iteration delta for the survivors.
- Finalist stage (`runBulkSimFinalistStage` / `runConcurrentBulkSimFinalistStage`): the top `topResults` and the
  baseline get lockstep extra iterations until every adjacent pair separates at `BulkSimZ95`, or
  `BulkSimFinalistMaxExtraIterationMultiplier` × the iterations so far is spent. Its results are the displayed set.
- Each top result ships `paired_error_to_next_result` and `paired_error_to_baseline`, computed before `AllValues`
  are stripped (0 = could not pair). Strip `AllValues` by detaching the slice before cloning
  (`bulkSimCandidateResultToProto`), never by cloning megabytes and discarding them.
- `BulkSimZ95` in `sim/core/bulk/statistics.go` is the single significance threshold; `Z_95` and `zTest` in
  `ui/sim/utils/math.ts` read the generated `BULK_SIM_Z95`.

## Seed and results display

- `runBulkSim` hashes the baseline gear and `BulkSettings` (each serialized and hashed, the digests combined) into
  `simOptions.randomSeed`: the same setup reproduces bit-identical results, any change draws a fresh sample. A
  fixed RNG seed takes precedence, and `lastUsedRngSeed` records the seed actually used.
- `buildTieChains` groups adjacent rows the paired test cannot separate, using the backend rank for "next result"
  and falling back to an unpaired `zTest` where no paired error exists. Each row's ± is `stDevToConf95`, the
  unpaired error of its own mean, a different quantity from the grouping test.
- A slot holding the same item with another enchant renders `GearChangeIcon`; a new item renders plainly;
  unchanged slots stay empty. Starting a run clears the results and returns to the Setup pane; a finished run
  opens Results.
- The batch tab's Encounters group lists the spec's preset encounters and the user's saved encounters
  (`ui/features/bulk/model/encounters.ts`). `bulkEncounterRuns` decides what a batch runs, for `runBulkBatch` and the
  combination count alike: the current encounter, then every checked row the group's list (`bulkEncounterList`)
  neither leaves out (equal to the current encounter) nor flags (equal to an earlier checked row). Each run replaces
  the whole encounter on the request copy only (the encounter override of `Sim.runBulkSim`). Every encounter shares the
  one content seed, and `results` holds one entry per finished encounter, the current one first.
- The Encounters group and Compare encounters are power tools (`showPowerTools` in the UI slice). With power tools off
  the group is hidden, the combination count shows no encounter multiplier, and `runBulkBatch` runs the current
  encounter alone; `bulkEncounterRuns` holds that rule for both. The checked encounters stay stored and count again
  once power tools are back on. Results already on screen stay as they are.
- With two or more finished encounters and power tools on, the encounter list offers Compare encounters: a table
  (`compareMatrix` in `ui/features/bulk/model/compare.ts`) with the current gear's row, then every set's, and a column
  per encounter by fight length, health-based ones last. Each cell is the set's gain over the current gear (in % of the
  current gear's DPS), tinted on a diverging scale (`gainTint`, five steps either way, each a fifth of the table's
  largest gain or loss), with the encounter's best outlined and a dashed gap where the encounter did not rank the set;
  its tooltip gives the rank, DPS, fight-to-fight spread and gain. A set keeps one "S<N>" number across all
  encounters, by gear identity. Under its number a set lists the items it changes (`changedItemLabels`); rings,
  trinkets or a dual wielder's weapons that only change places count as unchanged, as in the result rows
  (`ui/features/bulk/model/swappable_slots.ts`).
- The Setup/Results tab strip spans the batch tab, and each pane has its own columns. Setup holds the settings
  sidebar. Results holds the ranking and, for a run with more than one finished encounter, an encounter list
  (`BulkEncounterResults`, a vertical, neutral-tone `ToggleGroup`); picking a row shows that encounter's ranking.
  With one result there is no list and the ranking takes the full width.
- Progress is throttled (`BulkSimProgressThrottle` in Go, 100 ms in the browser emitter). Every stage, the
  finalist included, reports through the shared stage-metrics builder.

## Validation

```bash
make proto      # after a proto change
make go-to-ts   # after changing an exported constant or slot map in sim/core/bulk
go test -tags with_db -count=1 ./sim/core/bulk ./sim/web
npx vitest run ui/features/bulk ui/sim/bulk ui/sim/proto/buff_items.test.ts ui/features/enchant-presets
npm run type-check
```

Reproducibility: run the same `BulkSimRequest` twice through `BulkSimAsync` with the same seed; the results must
be bit-identical. For a request-database change, run a batch against the dev server, which has no `with_db`, not
only the Go tests.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill otherwise. A memory file is a pointer
to this skill, not a second copy.
