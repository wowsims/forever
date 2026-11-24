---
name: wowsims-bulk-sim
description: 'Use for any WoWSims Forever bulk sim ("Batch" tab) work, whether it runs natively, on the server or in the browser: candidate generation from BulkSettings and the enchant presets, Builds mode (bundles of gear set, talents, rotation and settings, run over ticked encounters), Drops mode (one new item per candidate, by source: `single_change`, `cull_to_baseline`), the low/medium/high and finalist stages, survivor culling, content-derived seeds, paired errors and tie groups in the results, progress, abort, combination counts, and the item database a bulk request carries to the workers. Also use it when a change to the sim request or its SimDatabase could reach a batch run. The gear optimizer (EP, stat caps, HiGHS) is `wowsims-reforge-optimizer`; React and store work outside the bulk feature is `wowsims-ui`.'
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
  slots), `item_keys.go` and `item_rules.go` (dedupe, equip rules), `single_change.go` (`single_change` sets),
  `candidates.go` (the `BulkCombinationCount`, `BulkCandidates` and `EnsureBulkSimCandidatesGenerated` entry
  points).
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
- Feature: `ui/features/bulk/model/run/run.ts` (`runBulkBatch`, cancel, combination refresh), `run/core_sim.ts`,
  `batch_encounter.ts`, `results/tie_chains.ts`, `gear/settings.ts`; components under `ui/features/bulk/components/`; the tab is
  `ui/app/tabs/BulkTabBody.tsx`. `ui/sim/settings/bulk_settings.ts` stores the settings, the enchant presets, the
  mode and the Builds picks.
- Builds mode, `ui/features/bulk/model/builds/`: `builds.ts` (the planner: `bulkGearSetSetups`, `bulkBundlePlans`,
  `bulkBuildPlan`, `bulkPlanRuns`, `bulkBuildsCount`), `bundles.ts` (the bundle list, its rows and storage),
  `encounters.ts`, `settings_entries.ts`, `gear_sets.ts`, `enchant_picks.ts`, `gear_setups.ts`, `talents.ts`,
  `rotations.ts`, `saved_refs.ts`; its results in `ui/features/bulk/model/results/` (`results.ts`, `results_csv.ts`,
  `compare.ts`). The plan, its `sameAs` and the count are built once per change in `ui/features/bulk/context/BulkBuildsContext.tsx`: `BulkBuildsProvider` wraps the Setup
  pane in `BulkTabBody.tsx`, and `useBulkBuildsPlan` (null outside Builds) hands them to the Builds column and the run
  card.
- Drops mode, `ui/features/bulk/model/drops/`: `sources.ts` (the source leaf keys, labels and tree), `items.ts` (the
  pickable list), `filters.ts` (the items dialog's search and filters), `settings.ts` (the setters and
  `dropsRequestSettings`), `results.ts` (`dropsGap`, rows grouped by
  item and by source). Components under `ui/features/bulk/components/`: `DropsSetup` (`DropsSources` tree,
  `DropsItemsCard` summary, `DropsItemsDialog`), `DropsRunCount`, `DropsResults` (`DropsSummary`, `DropsBossCard`,
  `DropsItemLabel`). `TotalIterations` is the run card's iterations row, shared by Gear, Builds and Drops, and
  `BulkDpsCell` the DPS, margin and difference of a Gear, Builds or Drops result row. A Drops result row is Gear's
  (`BulkResultRowView`), and `BulkResultSlots` draws the slot cells of both.
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
- **Explicit candidates (Builds):** `BulkSimOverrides.candidates` skip generation on both paths
  (`EnsureBulkSimCandidatesGenerated` returns early, the browser skips `getBulkCandidates`). They are deduped against
  the baseline and go out with no `BulkSettings`. Go generates candidates only for a request with bulk settings and no
  candidates, so with the Gear-mode settings a request whose candidates all repeat the baseline would have Go generate
  Gear-mode combinations.
- **The baseline:** `BulkSimOverrides.baselineGear` stands in for the live gear wherever `runBulkSim` reads the
  baseline: `player.equipment`, the seed hash, the `dedupeGearSets` check and the request database. Gear mode and a
  Builds batch's reference request, sent only while the Current bundle is ticked, leave it unset and sim the live gear.

Because both paths run the same algorithm, a change to staging, culling, statistics, carry-over or progress in
`sim/core/bulk` needs the same change in `ui/sim/wasm/bulk_sim`, and a new tuning constant goes through
`make go-to-ts` rather than being copied by hand.

Both paths build each candidate's request, the baseline's included, from the base request: the candidate's
equipment, imbues adjusted to its weapons, `saveAllValues` on. A Settings entry's main-hand and off-hand stones
therefore go out as saved; `applyBulkSettings` fits only the swap stones, to the entry's swap weapons.

## The request database

Only `with_db` builds load every item (`sim/core/database_load.go`). The server binary (`go build ./sim/web`) and
the WASM build do not use the tag, so they know only the items in the request's `SimDatabase`, while
`go test -tags with_db` sees every item. A green Go test therefore does not prove the request carries what it
needs: the generator fails with "selected bulk item … is missing from the database" for an item the request did
not ship, and core code reading an item the player does not wear gets nil.

- In Gear mode the bulk database is built from the baseline gear, `BulkSettings.items` (a selected spec's own
  enchant and suffix included) and the enchant pool (`makeBulkItemDatabaseFromSpecs`), on both paths, never from the
  generated candidates. An id a generated candidate can carry reaches the workers only through `BulkSettings`.
- A Builds request builds it from its baseline (`baselineGear`, or the live gear when unset) and its explicit
  candidates (`makeBulkGearDatabase`). A settings entry adds its swap gear and the items its party buffs read
  (`applyBulkSettings` in `ui/sim/sim.ts`).
- It is merged into the player's database (`Database.mergeSimDatabases`), never assigned over it. The player's
  database already holds the items party buffs read (`addPartyBuffItems` in `ui/sim/proto/buff_items.ts`, e.g. the
  party Darkspear's item); replacing it drops them silently. Any code path that rebuilds `player.database` either
  merges into it or calls `addPartyBuffItems` afterwards, as `runSim` with `options.gear` does.

## Candidates

- The baseline is `base_request.raid.parties[0].players[0].equipment`; `BulkGearCandidate.index` is a candidate's
  stable identity through every stage.
- `rawCombinations` is the mixed-radix index space of the combination table; `combinations` is the runnable count
  the UI shows. There is no set-bonus filter, so the two are equal for `BulkCombinationCount`. The count carries no
  iterations: the UI estimates the run card's iterations itself with the picked top results.
- Slot eligibility goes through `core.EligibleSlotsForItem` (shared with item swap), spec lookup through
  `core.PlayerProtoToSpecSafe`, request validation through `newGeneratorFromRequest`.
- Hand swap (`initSelectedItems`): with no weapon selected, the worn weapons stay in the hands they are worn in.
  Once any weapon is selected, a dual-wielder's worn weapons join the hand-weapon pool and every allowed
  main-hand/off-hand order is a candidate.
- `BulkSettings.single_change` (`single_change.go`) swaps in one new item per candidate and leaves the rest of the
  worn gear alone. A new item is a selected (id, random suffix) worn in no slot; a copy of a worn item yields
  nothing. Each candidate is built straight from the selection, not through the combination table. A ring or
  trinket goes into either slot. A placement is dropped when the item staying in the paired slot (other ring,
  other trinket, other hand) is the same Unique-Equipped item or shares its limit category (`excludesEachOther`).
  Enchant presets apply to new items only; with no pool entry that fits, the new item takes the slot's worn
  enchant when it fits, else none; into an empty slot it keeps its own. Worn items join no pool and are never
  re-enchanted. A two-hander goes in the main hand and empties the off hand. A one-hander under a worn two-hander
  goes in the main hand with the off hand empty. For a dual wielder, the one-hander worn in the new item's hand
  may also move to the other hand, keeping its own enchant. An off-hand-only item under a worn two-hander makes
  no candidate. The freeze fields, the weapon-type filters among them, have no effect. A set equal to the
  baseline or an earlier set is dropped, and the count equals the candidate list.
- Game rules the generator encodes: no reforging, gems or sockets (an `ItemSpec` is id, random suffix and
  enchant); no Titan's Grip, so a two-hander never shares the hands with an off-hand; the Ranged slot iterates like
  any non-weapon slot.
- A ring enchant needs Enchanting. For a player without it, `slotSpec` strips both ring slots' enchants, the
  baseline's and a selected ring's own, so no two candidates differ only by a ring enchant.
- The combination table resolves each choice's item once per physical slot, and a candidate's item specs share
  one backing array; keep per-candidate allocations out of that loop.

## Enchant presets

The bulk sim crosses enchants through its checked enchant presets. `enchantPresetPool` turns them into
`BulkSettings.enchants` on every Gear and Drops request (Finger 1 stands for both rings; Main Hand enchants go to both
hands); the pool is derived, never stored, and sorted so preset order does not change the seed. The presets live in
the bulk slice (`enchantPresets`, storage key `bulk-enchant-presets.v1`); checking a preset, or saving or deleting a
checked one, bumps the slice's `settings` counter so the combination count refreshes. Builds checks the same presets
on its own (`bulk-build-enchant-presets.v1`) and pools them per slot on every bundle's gear set in `gear_setups.ts`.
That pick holds preset names, so renaming a preset carries its pick over and deleting one drops it.

The gear optimizer's "Use enchant presets" option reads the same presets. That is the only thing `BulkSettings`
and `ReforgeSettings` share; keep the two apart in request shaping and serialization. The fit rule
(`core.EnchantAppliesToItem`, `core.FittingEnchantRow`) is shared with the optimizer too.

## Stages and statistics

- Every candidate runs the same seed sequence, so comparisons use paired errors over `AllValues`
  (`bulkSimPairedDpsError`), far tighter than per-result stdev.
- Iterations come from the stage configs (minimums raised by the error target), not from `SimOptions.iterations`.
  Below `BulkSimMinCombinations` candidates, or when it would sim fewer iterations, the run goes straight to the
  high stage. So does a request whose `top_results` covers every candidate (`shouldUseSingleStageBulkSim`): culling
  cannot change what it shows. The UI's `estimateBulkSimIterations` (`ui/sim/wasm/bulk_sim/estimate.ts`) takes the
  same `topResults` and follows the same rule.
- Low and medium stages cull to survivors, and the cull respects the cap: every culling stage raises its max
  survivors, and with them the soft cap, to at least the Top results N (`getBulkSimStageMaxSurvivors`) and keeps the
  top N by mean, never fewer than the stage minimum (`MinSurvivors`, 20 in the low stage, 5 in the medium). With N
  above 5 (`BulkSimDefaultTopResults`), every culling stage culls against the N-th best candidate
  (`bulkSimCullingCutoff`, `selectBulkSimSurvivors`, Go and wasm alike); at 5 or fewer, as in Gear without power tools,
  against the leader, the baseline included. A cap N below the candidate count therefore always returns N rows. The
  culling plan's estimate (`getBulkSimOptimisationIterationsUpperBound`) counts the same survivors, so the run card
  stays a true minimum.
  Culling intervals use the run's original candidate count, iteration targets the current stage's count. A stage that
  misses its target error gets up to `BulkSimMaxAdaptivePasses` extra segments with offset seeds, merged into the
  existing metrics, and the carry-over lets the next stage sim only the iteration delta for the survivors.
- Cull to baseline (`cull_to_baseline`, sent through `BulkSimOverrides.cullToBaseline`): each culling stage drops
  only the candidates whose baseline mean minus their own is more than `BulkSimZ95` paired errors
  (`cullBulkSimStage`, Go and wasm alike); an unpairable candidate stays. When more stay than the stage's max
  survivors for that count, the top-results cull runs on the above-cut set, soft cap included, so a culled row can
  still sit above the baseline.
  `top_results` and `all_results` are not read: the stages and the single-stage check plan for
  `BulkSimDefaultTopResults` (`resolveBulkSimTopResults`), so the estimate is the default one. Every candidate that
  reaches the high stage comes back in `top_results`; the finalist stage refines only the default top 5 and the
  baseline, and the rest follow at their high-stage numbers, paired against the high-stage baseline. Each culled one
  comes back in `culled_results` with its `culled_stage` and the metrics and paired error of the stage that culled
  it, `AllValues` stripped. Under 28 candidates the run is single-stage and `culled_results` is empty. A run that
  culls everyone still sims the baseline at the high stage. `runBulkSimCandidateBatchOnWorkers` returns at once on an
  empty batch. That is a TS-only guard, since the async queue drains only after a push (Go's batch returns empty on
  its own), and the one intended break of the line-for-line mirror.
- Finalist stage (`runBulkSimFinalistStage` / `runConcurrentBulkSimFinalistStage`): the top `topResults` and the
  baseline get lockstep extra iterations until every adjacent pair separates at `BulkSimZ95`, or
  `BulkSimFinalistMaxExtraIterationMultiplier` × the iterations so far is spent. Its results are the displayed set.
  `all_results` ranks every candidate, generated ones too, and the sim resolves it to its candidate count, so native
  Gear runs need no count from the UI. All three modes send the batch's top results (`BulkSimOverrides.topResults`)
  through `Sim.runBulkSim`; Drops sends the default 5 with `cullToBaseline`. An override of N goes out as
  `top_results` N, an override of 0 (All) sets `all_results`, and with no override the request carries `top_results` 0
  and no `all_results`, which the sim defaults to `BulkSimDefaultTopResults` (5).
- Top results: with power tools on, the run card holds a "Top results" number field and an All
  checkbox (`BulkTopResultsField`), one value per mode (`gearTopResults`, `buildsTopResults` on the slice, stored per
  spec under `bulk-top-results.v1` as `BulkTopResults`, 0 for All). Gear defaults to 5 and Builds to All. With power
  tools off Gear ranks 5 whatever is stored, and Builds is not reachable (`bulkTopResults` in `top_results.ts`). A cap
  at or above a request's candidates sims every one at the high stage and the finalist stage refines them all; a lower
  cap culls to what can reach the top N. Gear with All therefore sims every combination at the high stage. The control
  holds 1 up to the combinations limit (`BulkTopResultsFieldView`), which keeps it inside the request's int32, and a
  `bulk-top-results.v1` blob that does not parse loads as nothing stored.
- Each top result ships `paired_error_to_next_result` and `paired_error_to_baseline`, computed before `AllValues` are
  stripped (0 = could not pair). `BulkGearResult.paired_dps_diff_to_baseline` is the mean per-iteration DPS difference
  (the result minus the baseline) against the baseline that error pairs with, 0 when the runs could not pair; it can
  pair with an earlier stage's baseline than the shown one. A row whose paired difference decodes as 0 while its paired
  error is not 0 is read by the averages (`hasPairedDiffToBaseline`), since a decoded response holds 0 for a difference
  the backend never set. Strip `AllValues` by detaching the slice before cloning (`bulkSimCandidateResultToProto`),
  never by cloning megabytes and discarding them.
- `BulkSimZ95` in `sim/core/bulk/statistics.go` is the single significance threshold; `Z_95` and `zTest` in
  `ui/sim/utils/math.ts` read the generated `BULK_SIM_Z95`.

## Seed and results display

- `runBulkSim` hashes the baseline gear as sent (`baselineGear` when given, else the live gear), `BulkSettings` and,
  when one is passed, the explicit candidate list (each serialized and hashed, the digests combined) into
  `simOptions.randomSeed`: the same setup reproduces bit-identical results, any change draws a fresh sample. The
  encounter, settings and build overrides stay out of the hash, but a settings entry without Enchanting strips the
  ring enchants from the baseline and the candidates, which are hashed, so the seed moves with it. A fixed RNG seed
  takes precedence, and `lastUsedRngSeed` records the seed actually used.
- `buildTieChains` groups adjacent rows the paired test cannot separate, using the backend rank for "next result" and
  falling back to an unpaired `zTest` where no paired error exists. The unpaired test reads each row's own iteration
  count (`rowIterations`), the run's count for a row that carries none; against the baseline the same rule holds for the
  tie (`isTiedWithBaseline`) and the z (`baselineSignificance`). Each row's ± is `stDevToConf95`, the unpaired error of
  its own mean, a different quantity from the grouping test.
- A slot holding the same item with another enchant renders `GearChangeIcon`; a new item renders plainly;
  unchanged slots stay empty. Starting a run clears its mode's results and returns to the Setup pane; a finished run
  opens Results. Switching modes leaves Setup or Results open as it is (`BulkTabBody`).
- The batch has three modes. Gear crosses the picked items and the enchant presets under the current encounter,
  settings, talents and rotation. Drops sims each picked item from the ticked sources on its own (below). Builds has
  two columns (`BulkBuildsColumns`), Builds and Encounters, the first the wider, and runs every ticked bundle on every
  ticked encounter. The top navigation's batch tab reads "Batch" and opens a menu of the modes on mouseover, on a tap
  on a touch screen, or on Enter, Space or ArrowDown (`BulkModeMenu`, the kit `Menu` with `openOnHover` that Import and
  Export use, around the tab through `TabNavTab.renderTab`). The menu is shown to everyone: Gear and Drops always,
  Builds only with power tools on (`bulkMenuModes`, `showPowerTools` in the UI slice). Clicking the tab opens the
  batch tab in the current mode and does not toggle the menu, which would keep it open after the pointer leaves;
  picking a mode sets it, bumps `settings` so the count refreshes, and opens the batch tab. The mode is
  `BulkSlice.mode`, stored per spec under `bulk-mode.v1` (default Gear). `bulkRunMode` is the mode a batch runs in:
  a stored Builds runs as Gear with power tools off, and a stored Drops stays. The stored mode outlives power tools
  being turned off.
- Gear and Drops hold the Batch encounter block (`BatchEncounter`, `BulkSlice.batchEncounter`) in the run card, one
  dropdown stored per spec under `bulk-batch-encounter.v1` as a `BulkEncounterRef`. The empty ref is "Current
  encounter", and the run uses the Encounter tab. A picked preset or saved encounter from `pickableBulkEncounters`
  runs as `BulkSimOverrides.encounter` and is named in the results. A ref whose encounter is gone reads as Current
  encounter and stays stored. Builds shows no block (`BulkSettings` passes none while the Builds plan exists) and its
  plan never reads the ref. The block never writes the Encounter tab.
- **Drops** answers which drop from the ticked sources beats the worn gear. The Setup pane (`DropsSetup`) holds the
  Sources tree (`dropsSourceTree`) and the Items card; the run card holds `DropsRunCount` and the Batch encounter block,
  with no Top results field. Gear's Enchant presets group (`EnchantPresets`) is the one settings group under the run
  card, and its ticks are Gear's. The leaf keys are `drop:<zoneId>:<npcId>`, `drop:<zoneId>:other:<name>` (the world drop is
  `drop:0:other:World Drop`), `quest:<zoneId>`, `crafted:<Profession>` (the numeric enum), `rep:<repFactionId>` and
  `soldBy:<zoneId>`. A quest or vendor zone with no `zones` row folds into `quest:other` or `soldBy:0`; PvP and a source
  with nothing to key on have no key. The tree's categories are Dungeons, Raids and World drops (`drop_dungeon`,
  `drop_raid`, `drop_world`, by the drop's `DropSource.difficulty`: 1, 2, and 0 or unset), then Quest rewards, Crafted,
  Reputation and Vendors. The three drop categories share the `drop:` leaf keys, so the difficulty is not in a key: the
  tree reads it from the source, and a stored tick keeps its key whichever category shows it. Each drop category lists
  zones, each with its bosses and then its other drops, and after the zones the drops with no zone (World Drop). The Endgame
  dungeons quick pick ticks every zone of Dungeons whose level, as below, is `DROPS_ENDGAME_LEVEL` (55) or more
  (`DropsSetupData.dungeonLevels`). Drops lists what a level-60 character
  still wears (`dropsLevelItems`, `DROPS_MIN_LEVEL` 45): a dungeon's level (`dropsDungeonLevels`) is the highest `Player.requiredLevelOf` of
  every database item it drops, whatever the class or phase, and a dungeon below 45 leaves the tree. Raid drops stay,
  and any other source stays only on an item that needs level 45. The Sources tree and the item list hold the items that
  cutoff, the phase, the player's side and the class let through (`dropsPhaseItems`, `dropsFactionItems`,
  `dropsUsableItems`), random-suffix items left out, with no note of any of them. The gear picker's filters
  (`sim.getFilters()`) never reach Drops. `dropsSetupFor` keeps the cutoff until the database or its item count changes,
  and the items, their source keys and the tree until the phase, the side or the class changes too. A worn item is
  listed but never sent.
  The Items column is a summary card (`DropsItemsCard`): how many non-worn items are ticked of how many are listed, how
  many ticked ones the dialog's filters hide, and a chip per slot with a ticked item. Its Choose items button opens
  `DropsItemsDialog`, laid out like the gear picker: a search field, a Filters button and the By slot / By source switch
  on top, a rail of All and then one entry per slot or source with its ticked and listed non-worn counts, and the picked
  entry's rows in a `VirtualList`. Below `md` the rail is one row of chips that scrolls sideways above the list, the list
  fills the dialog, and each row's sources sit under its name, so the rows are taller there (`useMediumLayout`). All
  lists each item once, and a source's rows leave out that source. Arrow up and down move the rail pick (`railStep`,
  shared with the gear picker's slot rail), and Include shown and Exclude shown tick or untick the non-worn rows the
  search, the filters and the pick leave. The Filters button opens the gear picker's
  `FiltersMenuView` with only the rows a list of every slot can use (item level, stats, bonuses, armor, weapon and
  ranged), over Drops' own `DatabaseFilters` (`dropsFiltersOf`, the class's defaults until set); `Player.passesItemFilters`
  holds the checks both lists run, here for each item's first slot. Search and filters only change what the dialog
  shows: a ticked item they hide stays ticked and is sent, and neither reaches `BulkSettings`, the seed or the count.
  Four keys store the Drops picks per spec: `bulk-drops-sources.v1` (the ticked leaf keys, sorted),
  `bulk-drops-excluded.v1` (item ids unticked under them), `bulk-drops-view.v1` (`bosses` or `items`) and
  `bulk-drops-filters.v1` (the dialog's filters). The sources and excluded ids bump `settings` so the count refreshes;
  the view and the filters do not.
- The Drops request (`dropsRequestSettings`) is `BulkSettings` with `single_change` on, the picked item ids as `items`
  and Gear's enchant pool (`enchantPresetPool`). It sends `BulkSimOverrides.cullToBaseline` with
  `BULK_SIM_DEFAULT_TOP_RESULTS` (`bulkTopResults` returns it in Drops), so every candidate comes back once, in
  `top_results` or `culled_results`, and the run ranks every returned row, never All (`rankEncounterResults` with 0).
  The run card counts candidates, not combinations: the count leaves out the current gear, an all-worn pick counts 0 and
  Simulate stays off, while one candidate is enough to run (`canRunDropsBatch`, `canRunBatch` with a minimum of 1). Its
  iterations are the estimate at the default top results (`bulkDropsIterations`, `bulkGearIterations` at
  `BULK_SIM_DEFAULT_TOP_RESULTS`). The row for the iterations, `TotalIterations`, is shared with the Gear and Builds
  cards and carries the limit warning.
- Drops results: `dropsStarted` and `dropsResults` are the mode's own, like Gear's and Builds'. A run also stores
  `dropsRunSources` (the ticked leaf keys) and `dropsRunFaction` (`player.getFaction()`) as it started. The boss cards
  group by those, never by the tree or the race as they are now, and `dropsItemLookup` resolves items one id at a time
  (`getItemById`) through the faction filter but not the phase filter, so a run keeps its cards after a phase switch.
  `DropsResults` shows `DropsSummary`, then the Bosses view (one `DropsBossCard` per ticked leaf the rows reach, best
  gain first, cards with no upgrade last) or the Items view (Gear's result rows with no header and no rank, the new
  item's name, slot and sources in a `DropsItemLabel` beside its slot; the current gear's row and, per item, every
  row that is a proven upgrade (a ring upgrading in both slots lists both) or its best row when none is, filtered by the
  one source a card click picked; that click switches the
  stored view to Items, and the filter is not stored). Every % and every delta uses `dropsGap`, the paired difference
  (`paired_dps_diff_to_baseline`) on a row with a paired error and a paired difference, and the gap between the averages
  only on a row without them, never `row.avg - current.avg`: a culled row can pair with an earlier stage's baseline, so
  its average can sit below the current gear's while its paired difference is an upgrade. The CSV's `diff` and
  `diff_pct` take it too, and the margin at the row's own iteration count. An upgrade is a row `dropsGap` puts above the
  current gear and `isTiedWithBaseline` tells apart from it. An item's best row is its proven upgrade with the largest
  gain, else its best average. `DropsSummary.bestGroup` is null when `best` has no card among the run's sources, and the
  summary then shows the best drop without a best source; the best-source tile shows the best drop's gain. M in "N / M"
  counts the card's strip, the source's worn items too.
- A row culled before the finalist set (`isCulled`: `culledStage` set to a stage) is lower precision: no ±, no tight
  bracket, an empty `dps_margin_95` in the CSV, and it brackets only with the current gear. Drops rows show their DPS
  whatever the hide-damage display setting says, since every Drops output is a DPS ranking. In the CSV the `rank` column
  skips the current gear's row; Gear and Builds rank it like any other row. Drops has no Compare
  button.
- `bulkRunMode` (`ui/features/bulk/model/mode.ts`, from power tools and the stored mode) is the one gate. `runBulkBatch`
  reads it, and `useBulkMode` derives `builds` and `drops` from it (`bulkRunsBuilds`, `bulkRunsDrops`) for the Setup
  pane, which in Builds shows `BulkBuildsColumns` and in Drops `DropsSetup` in place of the item picker. Outside
  Builds the Builds picks stay stored and count again on the way back. Each mode keeps its own last results and
  started flag (`gearResults`/`dropsResults`/`buildsResults`, `gearStarted`/`dropsStarted`/`buildsStarted` on the
  slice): a run writes only the mode it started in, and the Results pane, the encounter list, the CSV export and
  Compare read the shown mode's through `useBulkResults` (`bulkModeResults`). A mode that never ran shows the hint to
  run one. Compare encounters stays tied to power tools and two or more finished encounters.
- The Batch tab waits for the sim (`useSimReady`) before it reads the database, and Simulate stays disabled until then.
- **Builds column** (`BuildBundles`): a bundle is one gear set, one talents, one rotation and one Settings tab entry,
  and is the unit that runs. The Current bundle (the live gear, talents, rotation and settings) heads the list, ticked
  by default; unticked, it is stored as `current_unchecked`. The user's bundles follow in stored order, each labelled
  by the name the user gave it (`BulkBuildBundle.name`) or, without one, from its parts (`bulkBundleList`: "<gear set> ·
  <talents> · <rotation> · <settings>", a Current part as "Current gear", "Current talents" and so on). A named row's
  label tooltip lists the four parts, after the "Same as …" reason on a disabled row. It opens on the label's hover, and
  on the box's focus only when the row is disabled; the reason and the parts describe the box and the row's buttons for
  keyboard and screen reader users either way. Each row has a pencil that
  reopens the dialog on its picks and name (saving replaces the row and keeps its checked state) and an × behind a
  confirm. The Current bundle has neither and cannot be named. New bundle opens `BulkBundleDialog` with an optional
  Name, whose placeholder is the label the picks give, over four pickers, Gear set, Talents, Rotation and Settings,
  each offering Current, then the spec's presets, then the user's saved entries (`bundleOptions`). The name is saved
  trimmed. Save is off while all four picks are Current, and while the name matches another bundle's label, the
  Current bundle's included, whatever its case (`takenBundleLabels`, `bundleNameTaken`); a stored ref no option holds
  starts its picker at Current. A part may sit in any number of bundles. A gear set preset is listed only when the
  player can load it (`enableWhen`), brings no bonus stats, and its `onLoad` never runs; a grouped one shows and runs
  as "<name> (<group>)", and its stored ref carries the group. Deleting a bundle moves focus to New bundle. A disabled
  "Same as …" row, here and in the Encounters column, keeps its box in the tab order (the kit `BooleanPicker`'s
  `focusableWhenDisabled`) and shows its reason on the box's focus as on the label's hover (`labelTooltipPlace`):
  right of the label in the Encounters column, below the row in the bundle list, clear of the row's pencil and ×.
- Bundle duplicates go by content (`flaggedBundles` in `bundles.ts`): gear identity, bonus stats, talents string, the
  rotation as picked (`rotationKey` in `rotations.ts`, so the current rotation and a saved one holding the same list
  stay apart) and the settings as `applyBulkSettings` sends them (`settingsContentKey`: an item swap that is off reads
  as none, and an entry without spec options reads with the live ones through `withLiveSpecOptions`), the current
  settings read by `useCurrentSettings` (`readSavedSettings` at run start). A bundle with the content of the Current
  bundle while that is ticked, or of an earlier checked one, is disabled and shown unchecked, "Same as <bundle>", keeps
  its stored `checked`, and does not run or count. With the Current bundle unticked, a copy of it runs.
- `bulkBuildPlan` gives `sameAs`, each bundle that does not run by stored index with the earlier running bundle it
  sims the same as (a twin of a twin resolves to the running one; -1 for the Current bundle), and the Builds column
  shows each disabled and unchecked, "Same as <bundle>", with its pencil and ×, like a content duplicate. A twin sims
  like an earlier running bundle once its rotations resolve (the same talents, the same settings by
  `settingsContentKey`, the current settings as the live ones from `BulkBuildReference.settings` and an entry without
  spec options with the live ones, and, per gear setup, the same gear identity, bonus stats and resolved rotation),
  such as an Auto and a Simple pick that generate the same list, or a settings entry that holds the live settings; a
  ticked Current bundle is always the earlier one, and an unticked one is not in the plans, so a twin of it runs. A
  merged bundle adds no setup to its group: an earlier bundle of the group already sends every one of its setups
  (`sentGearSetupKey` in `gear_setups.ts`), say gear that differs only by ring enchants under settings without
  Enchanting; it names the bundle that sends its first setup. Neither runs or counts.
- `pruneBulkBundles` runs from an effect in `BuildBundles` once the sim is ready, since no gear set is listed before
  then. It drops from storage every bundle with a saved part (`preset: false`) whose entry is gone, so a deleted or
  renamed entry does not bring the bundle back. A bundle whose preset part is not pickable stays stored, with no row,
  and does not run.
- Enchant presets sit under the bundles and apply to every bundle. Each slot keeps the gear set's own enchant (or
  none) or takes any fitting enchant a checked preset holds for it, as Gear mode's enchant pool does (Finger 1 for both
  rings, Main Hand for both hands), and every combination across the slots is a setup, the set with its own enchants
  first (`bulkGearSetups`). With no enchant preset checked, each gear set is one setup. An enchant preset with no
  enchants is not listed, and with none listed the column offers New preset. Saving the first preset removes New
  preset, so the dialog hands focus to the enchant picks list (`finalFocus`), whose first box is the new preset.
  `bulkGearSetSetups(player, enchantPicks)` builds each gear set's setups the first time a plan asks for them and hands
  them to every later bundle of that plan and to the next plan. It keeps setups only from the last plan, so a gear set
  the next plan does not hold drops its setups.
- **Encounters column** (`EncounterList`): a ticked list, the current encounter first, then the spec's preset
  encounters and the user's saved ones in the Saved Encounters panel's order; one with no encounter or no target is
  left out. The current encounter is checked by its empty ref in the stored refs, seeded on fresh storage, and runs
  unnamed. While the current encounter is ticked, an entry that sims like it (`bulkEncounterList` compares the
  encounters without their proto version, zone or target names) is disabled and shown unchecked, "Same as Current
  encounter", keeps its stored state, and does not run or count; with the current encounter unticked it runs under its
  own name. A later entry with the content of an earlier checked one is disabled the same way, "Same as <encounter>".
  A tick rebuilds the stored refs from every pickable ref (`checkedRefsAfter`), so their order follows the list.
- The plan (`builds.ts`): `bulkBundleRuns` gives the Current bundle first while it is ticked, then every checked
  bundle the list shows that duplicates none, in list order. An unticked Current bundle does not run, and nothing sims
  the live setup. `bulkBundlePlans(bundles, setupsOf)` gives each bundle its gear set's setups. `bulkBuildPlan` leaves
  the twins out and groups the bundles with the same talents, rotation and Settings picks by ref (`buildKey`), in the
  order of each group's first bundle. A group runs under that bundle's talents, rotation and settings, so a gear setup
  only ever sims with its own bundle's, and setups never collapse across groups. Within a group, setups equal by gear
  identity and bonus stats run as the first, and the group splits into one request per bonus stats and resolved
  rotation its setups give (`BulkSimOverrides.build`: talents string, rotation, bonus stats). Resolved rotations
  compare by what the sim reads (`simmedRotationKey`), so an Auto and a Simple pick that generate the same list are
  one request. Ring enchants are stripped per group by its settings' professions (Enchanting in the settings entry,
  the live professions for the current settings), each setup keeping the gear it holds as `ownGear`
  (`gearSetupsSentWith`). Encounter outer, group inner, each encounter's group one step; the requests do not depend on
  the encounter.
- A ticked Current bundle's group is group 0, and its first request is the reference request: the split with the
  live bonus stats and the rotation the live gear resolves to, or an empty one, sent without `baselineGear`, so its
  baseline is the live gear as equipped, the reference row every bundle ranks against. Its candidates leave out a
  setup that wears the live gear, which the baseline stands for. The live gear goes to the Current bundle's group by
  its index (-1), never by group number: with the Current bundle unticked there is no reference request. Every other
  request sims its first setup as its baseline (`baselineGear`, `BulkBuildRequest.baselineSetup`) and the rest as its
  candidates, so no request sims gear outside its own bundles.
- Simple and Auto rotations resolve through `Player.getResolvedAplRotationForBuild` once per rotation pick, talents and
  items without enchants (`generatorGearKey` in `builds.ts`), at run start and for the shared plan. That holds because
  no spec's rotation generator reads an enchant: the generators read the talents and the main hand's hand type. A
  generator that reads an enchant needs the enchants in `generatorGearKey`. Encounters do not enter the resolve, and a
  custom APL goes out as it is.
- `BulkSimOverrides.encounter` replaces the whole encounter on the request copy only. `BulkSimOverrides.build` sets the
  talents, rotation and bonus stats; a Simple rotation also sends its own `hpPercentForDefensives` (as
  `player.cooldowns`), and an Auto or APL one keeps the live threshold. `BulkSimOverrides.settings` is
  the Settings tab entry a request runs under; the current settings send none. `applyBulkSettings` writes the fields
  `applySavedSettings` loads, with its defaults: raid, party and player buffs, debuffs, consumables, race,
  professions, item swap, reaction time, channel clip delay, facing, distance and healing model, and the entry's spec
  options (`SavedSettings.spec`, the arms of `Player.spec`) when it has them for the live spec (`isSameSpecArm`); an
  entry saved without them, one holding another spec's, a spec settings preset that sets none, or an item swap preset
  (it never carries them) keeps the live ones, and `withLiveSpecOptions` (`ui/sim/proto/spec_functions.ts`) keys it
  with them; the Saved Settings panel marks such an entry active with them too. A field `applySavedSettings` gains is
  added there too.
- The run card (`CombinationsCount`, given the plan's count through `useBulkBuildsPlan`) shows the gear setups the
  running bundles hold, "N bundles × M encounters" with "K runs", and the total iterations. N counts the Current
  bundle while it is ticked and leaves out every bundle in `sameAs` (`bulkRunningPlans`); K counts the requests after
  splits. With the Current bundle unticked and no other bundle ticked, the card counts 0 bundles, shows the Builds
  hint and keeps Simulate off. With runs but only one build in total, it shows "Tick a second build to compare." and
  Simulate stays off. The iterations limit applies per request. Simulate needs no request over the
  combinations limit, and with power tools on none whose estimated iterations pass the iterations limit either
  (`canRunBatch`, `canRunBuildsBatch` in `selectors.ts`). Without power tools Gear gates on the combinations limit
  alone: past the iterations limit the card warns, without the Top results hint, and the batch still runs. Builds
  needs power tools, so it always gates. The card estimates every request with the batch's top results
  (`bulkBuildsCount`, and in Gear mode `bulkGearIterations` from the combination count). A request that runs the high
  stage alone costs the high-stage floor per candidate and baseline: with All, from 100 candidates a request passes
  the browser's 1,000,000 limit, the card warns, and Simulate is off. With power tools on, the warning ends with a
  hint to lower Top results. A stored negative top results loads as its mode's default. A column with nothing ticked
  leaves no run, and the card names each such column ("Tick at least one row under Builds", "… under Encounters").
- `rankEncounterResults` (`results.ts`) merges an encounter's requests into one list: the batch's top results across
  all of them (every row with All), plus the current gear's reference row, under the Current bundle's talents,
  rotation and settings, when a request is the reference request (`BulkFinishedRequest.reference`). Compare gives
  every ranked set a row. Only rows from the reference request keep the backend's paired errors; the others fall back
  to the unpaired test, and the tie test uses the fewest iterations of any request. Any other request's baseline is a
  row of its own setup (`BulkFinishedRequest.setups.baseline`, at that request's reference DPS). With the Current
  bundle unticked there is no reference request: the top results are setup rows alone, every row is unpaired, and the
  best row is the reference (`originalGearResults`, with `BulkResults.referenceIsCurrent` false). The best row shows
  no delta but keeps its items and Equip; only the current gear's row reads "Current gear" over its DPS and hides
  both. Every other row shows its loss against the reference with the usual delta and significance colours. Every
  delta's percentage, in the rows, the encounter list, the CSV's `diff_pct` and Compare, is of the reference's DPS
  (`formatDeltaText` in `ui/sim/utils/format.ts`, shared with the results sidebar): 1000 → 800 and 1000 → 1200 both
  read 20%. Rows list their item changes against the gear worn when the batch ran (`BulkResults.currentGear`),
  whichever row is the reference. A Builds row is only one a setup stands for, any other is dropped, and each carries
  its request's talents, rotation and settings. A setup that wears the current gear under other talents, rotation or
  settings reads "No gear changes". A row carries the setup's own gear (`BulkGearSetup.ownGear` when the request sent
  it without ring enchants), so one setup keys alike under every encounter and Compare gives it one row. A row's key
  (`bulkResultKey`) is its gear identity, bonus stats, talents string, rotation as picked and settings as sent.
  `results` holds one entry per finished encounter, the current one first, kept only when all its requests finished.
- A Builds row shows four pills, gear, talents, rotation and settings, each its part's name behind a kind icon
  (the kind as screen reader text), a Current part as "Current gear", "Current talents" and so on. A row from a named
  bundle shows one pill with the name in their place, its tooltip listing the four parts (`BulkSetupLines`) and a
  hidden copy of them as its description (`bulkSetupText`). `bulkBuildPlan` carries the name on copies of the bundle's
  gear setups (`BulkGearSetup.bundle`), so a setup two bundles send carries the first sender's. "N enchants differ"
  follows the gear pill when the setup changes any of the set's own enchants (`changedEnchantLabels`, the slots and
  enchants listed on hover, `BulkEnchantsDiffer`, as in Compare). Hovering or focusing the talent pill opens a tooltip
  (`BulkTalentsGrid`) with every talent that holds a point, as its icon and rank under its tree's name and points, the
  talent's name as screen reader text. The talent pill and the enchant count each carry a hidden copy of their
  tooltip as their description (`bulkTalentsText` for the talents), since a screen reader does not reach the tooltip.
  Equip (`equipAndAnnounce` in Gear, Builds and Drops: `equipBulkResult`, then the gear tab and a toast) applies, in
  one `batch`, the gear, talents, rotation and bonus stats of the row's gear setup, and the row's settings; a row
  under the current settings leaves the live settings alone. A row with no gear
  setup, every Gear-mode row and the Builds reference row, applies the gear only and leaves the talents alone. The
  CSV (`results_csv.ts`) has the columns `encounter`, `duration`, `variation`, `rank`, `current_gear`, `dps`,
  `dps_margin_95`, `diff`, `diff_pct`, `talents`, `talents_string`, `gear_set`, `changed_enchants` ("Slot: Enchant"
  per changed enchant, joined by "; "), `rotation`, `settings` and `bundle` (the bundle's name, empty without one),
  then four per slot. `diff` and `diff_pct` are
  against the reference row, and `current_gear` marks a row that wears the gear worn when the batch ran.
- Stored per spec: `bulk-build-bundles.v1` (`BulkBuildBundles`: the bundles in list order with their checked state and
  name, and `current_unchecked`), `bulk-encounters.v1` (`BulkEncounterSelection`: the checked encounter refs in list order),
  `bulk-build-enchant-presets.v1` (the enchant presets checked in Builds, apart from Gear mode's pick in
  `bulk-enchant-presets.v1`), `bulk-mode.v1` and `bulk-batch-encounter.v1` (the Batch encounter ref). None of the
  Builds picks bumps the slice's `settings` counter. Drops' four keys are under Drops above.
- With two or more finished encounters and power tools on, the encounter list offers Compare encounters: a table
  (`compareMatrix` in `ui/features/bulk/model/results/compare.ts`) with the current gear's row, then every set's, and a column
  per encounter by fight length. Each cell is the set's gain over the current gear (in % of
  the current gear's DPS), tinted on a diverging scale (`gainTint`, five steps either way, each a fifth of the table's
  largest gain or loss), with the encounter's best outlined and a dashed gap where the encounter did not rank the set;
  its tooltip gives the rank, DPS, fight-to-fight spread and gain. With the Current bundle unticked there is no
  current gear's row: each column measures against its own best set, so that cell reads 0.00%, outlined, and every
  other a loss; the caption says so (`compare.table_caption_best`) and the legend runs from "Loss" to "No change". A
  set keeps one "S<N>" number across all encounters, by `bulkResultKey`. Under its number a Builds set names every
  part of its bundle, the gear set with the enchants it changes, the talents, the rotation and the settings, or a named
  bundle's name with the parts in its tooltip; a
  Gear-mode set lists the items it changes (`changedItemLabels`); rings, trinkets or a dual wielder's weapons that
  only change places count as unchanged, as in the result rows (`ui/features/bulk/model/gear/swappable_slots.ts`).
- The Setup/Results tab strip spans the batch tab, and each pane has its own columns. Setup holds the settings
  sidebar, which in Builds is the run card alone. Results holds the ranking and the encounter list
  (`BulkEncounterResults`, a vertical, neutral-tone `ToggleGroup`), shown for one finished encounter too; picking a
  row shows that encounter's ranking. Each row gives the encounter's best DPS and its gain over the current gear;
  with the Current bundle unticked it gives the best DPS alone, under its own caption
  (`results.encounters_caption_best`). The list holds the CSV export, and Compare encounters from two encounters on.
- Simulate Batch in Builds needs more than one build in total (`BulkBuildsCount.builds > 1`, each request's baseline
  and candidates), as Gear mode needs more than one combination. One build over two encounters still runs; with the
  Current bundle unticked, one bundle of one setup over one encounter is one build and stays off.
- Progress is throttled (`BulkSimProgressThrottle` in Go, 100 ms in the browser emitter). Every stage, the
  finalist included, reports through the shared stage-metrics builder. A batch of more than one step lists its steps
  in the progress dialog as a table (`BulkProgressSteps`: Encounter, Settings, Talents, Rotation, status), one row per
  encounter and bundle group, named from the group's first bundle; Gear mode is one step and shows none. Each row
  carries its own progress line (`ProgressTrackerBar variant="line"` over the row separator), named by its encounter
  and talents ("Run progress: <encounter>, <talents>"): the iterations its requests simmed, counted from the ticks
  phase by phase (`requestIterationsProgress`), each request weighted by `estimateBulkSimIterations` of its candidates
  (`steps.progress` in `run.ts`). `steps.groups` holds each step's group, and `batchSecondsRemaining` predicts a step
  from its group's latest finished step, scaled by their weights, so the time left shows once every group has finished
  a step, after the first encounter. A one-step run, every Gear run included, has no finished step to predict from, so
  it shows the stage's own time left under its bar.

## Validation

```bash
make proto      # after a proto change
make go-to-ts   # after changing an exported constant or slot map in sim/core/bulk
go test -tags with_db -count=1 ./sim/core/bulk ./sim/web
npx vitest run ui/features/bulk ui/features/bulk/model/drops ui/sim/bulk ui/sim/wasm/bulk_sim/estimate.test.ts \
  ui/sim/wasm/bulk_sim/statistics.test.ts ui/sim/proto/buff_items.test.ts ui/features/enchant-presets \
  ui/sim/sim.test.ts ui/app/tabs/BulkTabBody.test.tsx ui/app/tabs/BulkTabBody.builds.test.tsx \
  ui/app/SimTabsSection.test.tsx
npm run type-check
```

Reproducibility: run the same `BulkSimRequest` twice through `BulkSimAsync` with the same seed; the results must
be bit-identical. For a request-database change, run a batch against the dev server, which has no `with_db`, not
only the Go tests.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill otherwise. A memory file is a pointer
to this skill, not a second copy.
