# State: one Zustand store per page

**Source of truth:** `ui/sim/state/sim_store.ts` for the shape, `ui/sim/state/subscriptions.ts` for the read side,
`ui/sim/state/batch.ts` for gating, `ui/sim/hooks/` for the React reads. `ui/sim/state/README.md` is the long-form
companion. Print the slice list from the code:

```
sed -n '/^export interface SimState {/,/^}/p' ui/sim/state/sim_store.ts
```

There is one store per page, `Sim.store`, built by `createSimStore()` and wrapped in `subscribeWithSelector`. It needs
no browser (see the `no-restricted-globals` ban in `layers.md`), which is what lets `ui/sim/sim.test.ts` build a `Sim`
against an in-memory `Env`.

## Contents

- Slices and who owns them
- The write path
- The read path
- Reading the store from React
- Bindings: `useInput`, hide/restore, kept selections, text fields, lists
- The dominant defect: reading state you did not subscribe to
- State vs events
- Writes that need a specific shape
- Adding a field
- Player lifetime

## Slices and who owns them

| Slice                   | Owner facade                       | Contents                                                                                            |
| ----------------------- | ---------------------------------- | --------------------------------------------------------------------------------------------------- |
| `ui`                    | `Sim`                              | show\* flags, language, `wasmConcurrency`: presentation                                             |
| `sim`                   | `Sim`                              | iterations, phase, faction, `fixedRngSeed`, filters, `lastUsedRngSeed(+version)`, `metadataVersion` |
| `encounter`             | `Encounter`                        | duration, execute proportions, `targets` (replace-on-write)                                         |
| `raid`                  | `Raid` / `Party`                   | buffs, debuffs, tanks, targetDummies, numActiveParties, `partyBuffs[5]`, `composition[5][5]`        |
| `players[storeKey]`     | `Player` / `ItemSwapSettings`      | the values, as `PlayerSlice`, plus a per-field version counter `v` keyed by `PLAYER_FIELDS`         |
| `reforge[storeKey]`     | `ReforgeSettings`                  | gear-optimizer settings (`REFORGE_FIELDS`) and counters                                             |
| `statWeights[storeKey]` | `StatWeightActionSettings`         | excluded stats, EP precision settings and a counter                                                 |
| `statSweep[storeKey]`   | `StatSweepSettings`                | the stat sweep's stats, range, step, plotted metric and precision settings, and a counter           |
| `bulk[storeKey]`        | `ui/sim/settings/bulk_settings.ts` | the batch's items, picker groups, limits, run flags and results, plus `v: { settings, items }`      |
| `runs`                  | `SimRuns` (`ui/sim/sim_runs.ts`)   | `{ isRunning, isAborting }` per `SimRunKind`: one user-visible operation, not one worker request    |

`runs` is keyed by `SimRunKind` and written only through `patchRun`. A bulk sim issues many worker requests under one
progress bar and one stop button (see `wowsims-bulk-sim`), which is why the unit is the operation.

Kept on the classes on purpose: the Party↔Player object graph (the store's `composition` is the notification source),
`aplRotation` (tracked by the `rotation` counter), caches, and server-derived metadata objects.

## The write path

A facade setter is `setX(value)`: it guards equality, then writes once through one of the writers in `sim_store.ts`:

- `patchSlice(store, 'sim', patch)` for the unkeyed slices
- `patchKeyed(store, 'players', key, patch, bumps)` for the per-entity slices, with `seedKeyed` / `deleteKeyed` for
  lifetime
- `patchRun(store, kind, patch)` for `runs`

One logical change is one `setState`: the value and its counter bump go in the same write, so a subscriber sees it once.
Values that must notify even when the new value compares equal (`epWeights`, `epRatios`, `currentStats`, `rotation`,
`itemSwap`, `lastUsedRngSeed`, the reforge and stat-weight counters) bump a version counter. Bump the counter rather
than content-hashing a value to force a notification. Features may not import the writers (lint-enforced); they go
through a facade.

## The read path

Subscribe through the helpers in `ui/sim/state/subscriptions.ts`, which come in two kinds:

- field level: `subscribePlayerField(player, 'gear')`, `subscribeSimField`, `subscribeUiField`,
  `subscribeEncounterField`, `subscribeRaidField`, `subscribeBulkField`, `subscribeRunState(sim, kind)`, …
- aggregate level: `subscribePlayerChange`, `subscribeEncounterChange`, `subscribeSimChange`, `subscribeStatsInputs`, …

`subscribeAll([...])` folds several selector sources into one tuple selector, so a write or batch that touches all of
them notifies once. List the current set with `/usr/bin/grep -n "^export function" ui/sim/state/subscriptions.ts`.

Every helper is batch-gated through `subscribeGated` (`ui/sim/state/batch.ts`): inside `batch(() => …)` a subscriber
is deferred and fires once at the end with the final state. An error thrown inside `batch()` is logged, not propagated,
and pending listeners still flush.

Aggregates leave out server-derived state on purpose: `PLAYER_CHANGE_FIELDS` drops `currentStats`, and `simSettingsKey`
drops the rng seed and metadata counters. Including them makes the stat recomputation re-trigger itself, a loop that
looks like a perf bug.

## Reading the store from React

Picking the wrong read is the most common defect here (see "The dominant defect" below). In order of preference:

**A plain selector.** `usePlayerStore('gear')`, `useSimStore('phase')`, `useUiStore('showEPValues')` in
`ui/sim/hooks/`. `PlayerSlice` holds the real values as stable references and `patchKeyed` replaces one only on a real
change, so Zustand's identity check is the whole subscription. `usePlayerStore` is typed on `keyof PlayerSlice`, not
on `PlayerField`: `PlayerField` is the version-counter list (`PLAYER_FIELDS`), which carries `rotation`, `itemSwap` and
`epRefStat` (no value in the slice) and omits the values those stand for. Use the slice for a value and the counter for
a notification. There is one player per page, read from `usePlayer()`.

**Store fields named as data.** `StoreField` (`ui/sim/hooks/useStoreField.ts`) is a bare player field or a scoped one
(`sim:` `ui:` `encounter:` `raid:` `reforge:` `bulk:` `statWeights:`). `useStoreField(field | [fields])` resolves it
against the host from context and folds an array into one `subscribeAll`, memoised on the names. Pass the result to
`useStoreSubscribe(useStoreField([...]), read)` when a component reads a facade method; `ConsumesPicker` does this
with `CONSUMABLE_ORDER_FIELDS`. Prefer it to a hand-built `subscribeAll([...])` in a render body: that builds a new
source each render, and `useStoreSubscribe` re-subscribes whenever the source identity changes. A named field is also
greppable and cannot capture the wrong object.

`StoreField` has two aggregates, `'player:*'` and `'encounter:*'`, which re-render a component on every write to the
slice. Use one only when nothing narrower works, and keep their number from growing:

```
/usr/bin/grep -rn "player:\*\|encounter:\*" --include=*.ts --include=*.tsx ui/ | /usr/bin/grep -v "useStoreField.ts\|\.test\."
```

**`useStoreSubscribe(source, read)`** (`ui/sim/hooks/useStoreSubscribe.ts`) for any other source: an APL row, a facade
with its own `.subscribe`, an `Emitter`. It caches the snapshot and re-reads only on a notification or a new source,
because `useSyncExternalStore` loops on a fresh object from `getSnapshot` and most model getters return one. So `read`
may touch only live model state: a prop or component state it captures is answered one notification late, for good. A
value keyed by a prop belongs in the render body or in a plain selector. The before-init form is
`useReadyStoreSubscribe(source, read, ready)` (`ui/sim/hooks/useReadyStoreSubscribe.ts`), which folds `ready` into the
subscription's identity, because `getSnapshot` runs on the first render and would otherwise cache a `null` forever.

**Context reads, not subscriptions.** `useSpecConfig()`, `useSpecPresets()`, `useSimHost()`, `usePlayer()`, `useSim()`
and the nullable `useOptionalSimHost()` (`ui/sim/context/SimHostContext.tsx`). The spec declaration is built once and
never replaced, so these re-render nothing; a reactive value read through one is right on mount and stale afterwards.

Composite hooks already exist: `useTalents()` (`ui/features/talents/hooks/useTalents.ts`) and
`useAplRotation(select)` (`ui/sim/hooks/useAplRotation.ts`), which keys on the `rotation` counter because the rotation
has no value in the store.

Outside React, cache a derived value with `memoLast()` (`ui/sim/utils/misc.ts`), a one-entry memo that recomputes when
one of its dependencies changes identity. Use it instead of a hand-written "recompute when these change" cache; the
consumables order and the stat-filter model use it.

## Bindings: `useInput`, hide/restore, kept selections, text fields, lists

`ui/features/inputs/` is the one place a kit picker meets the store. An `InputConfig` (`ui/features/inputs/input.ts`)
goes to a `Connected*` adapter (`ui/features/inputs/components/`), which calls `useInput`
(`ui/features/inputs/hooks/useInput.ts`) and hands the kit picker plain `value` / `onChange`. A config names its source
with `storeField` (preferred, resolved by `useStoreField`) or `storeSubscribe`; the `StoreBinding` type requires one,
and `storeSubscribe` wins when both are set. Omit both for an input its parent re-syncs (the nested APL pickers) or a
UI-local toggle; `useInput` then notifies itself after `setValue`.

`useInput` returns `{ value, setValue, hidden, disabled, revision }`. `revision` increments on every notification,
including ones that leave the value unchanged, so it is not a change counter. Its reader is `useHideZeroRestore`
(`ui/features/inputs/hooks/useHideZeroRestore.ts`), which uses it to tell a real notification from construction and a
StrictMode replay before it zeroes a hidden input or restores a shown one. The stored value comes back only onto a
source that still reads as zeroed, so a value set while hidden (a loaded build) is kept. `ConnectedIconPicker` and
`ConnectedIconEnumPicker` call it; the enum adapter also passes `onSettled` to reset a value whose option its
`showWhen` hides, and a user pick calls `clearStored()` first.

**Values that change with the model.** An icon-enum config's `values` may be an array or `(obj) => Array`, resolved
through `iconEnumValues` (`ui/features/inputs/components/ConnectedIconEnumPicker/utils.ts`) on every render and every
time the menu opens. That is how the consumables follow their EP order (`orderedByEp`, `consumableOrderOf`) without
the picker subscribing to the EP fields.

**Kept selections.** The icon, icon-enum and multi-icon configs implement `KeptInput.keptBy(obj, value)`
(`ui/features/inputs/input.ts`): true when the stat filter hides a selection that stays offered because it is
selected. The adapter turns that into the kit's single `keptTooltip` prop through `keptTooltipOf`; present means kept.
`IconSwatch` (`ui/ui-kit/IconSwatch/IconSwatch.tsx`) owns the dimmed swatch, the corner badge with the tooltip, and
the `data-kept` attribute on the swatch box. `valuePicked` and `inputPicked` (same `utils.ts`) answer whether a value or
an input holds a pick. The adapters derive the kept tooltip, the active input and `showWhen` in the render body, not
inside a `useStoreSubscribe` read: the stat filter hands in a new config when its stats change, and a read cached in
the snapshot would answer from the old config.

**Text fields.** `TextField` (`ui/ui-kit/TextField/TextField.tsx`) and its presets hold an uncommitted edit in
`useDraft` (`ui/ui-kit/hooks/useDraft.ts`), keyed on the formatted display string, so a same-value write or a
`revision` bump keeps a mid-edit draft. With the default `commit="blur"`, used by every connected text field, they
commit through `useCommitChange` (`ui/ui-kit/hooks/useCommitChange.ts`) on the native `change` event (blur after an
edit, or Enter), so typing writes nothing to the store until the edit is committed. `SearchField` commits per
keystroke or when typing pauses, and `ChoiceField` (`ui/ui-kit/ChoiceField/ChoiceField.tsx`) on selection.

**Lists.** `useListItems` (`ui/features/inputs/hooks/useListItems.ts`) binds an array field: `create`, `append`,
`remove`, `copy`, `move` and each row's `itemConfig` write one fresh array through `config.setValue` and never mutate the array
they read. Rows find their item by key (`ui/features/inputs/list_ids.ts`), so a row keeps its item through a reorder
and through a source that returns clones. The caller renders the array on the kit `Tree` (`ui/ui-kit/Tree/`) and wires
`move` to `sortable.onMove`.

## The dominant defect: reading state you did not subscribe to

A component subscribes to the fields it obviously uses, then calls a facade method that reads a field it did not name.
The render is right on mount and silently stale afterwards, with no error, warning or failing test.

The shape to look for: a component calling `player.hasProfession(...)` (which decides whether a profession-only enchant
is offered) while subscribing only to `gear`, or inside a `useMemo(…, [player, item])` with no profession subscription
in the file. Toggling the profession leaves the list stale until some other re-render. The fix is to subscribe to what
the method reads, here `profession1` and `profession2`. Naming fields at the subscription site (`storeField`,
`useStoreField`) puts the read set in one greppable place, which is the practical reason for the preference order
above.

No lint finds these: the dependency is only visible by following the call into `ui/sim/player/player.ts`. When you add
a facade call to a component, audit what that method reads.

`react-hooks/exhaustive-deps` gets the inverse case backwards. The `itemsToDisplay` memo in
`ui/features/gear/components/SelectorModal/ItemList.tsx` lists `filters` although its body never names it, because
`player.filterItemData` reads `sim.getFilters()` inside; the comment beside it says so. Treat a dependency the lint
calls unnecessary like a missing one: read the callee before deleting it.

## State vs events

- A value in the store changed: a store subscription. Class-side values (`aplRotation`, item swap, rng seed, reforge
  and stat-weight settings) are tracked by version counters; bump the counter on every write that should notify.
- Something happened (a sim result, a crash, a reference set, progress): an `Emitter<T>` from
  `ui/sim/state/events.ts`. `on` returns an unsubscribe function and `emit` fires synchronously, with no batching or
  dedup. Keep these out of the store, which would turn an event into a value that a late subscriber reads as current.

`Emitter.on` is an arrow property, so `storeSubscribe: () => this.changeEmitter.on` is valid.

## Writes that need a specific shape

**Rotation.** `player.modifyAplRotation(r => …)` mutates and then bumps the `rotation` counter, as
`player.touchRotation()` does. The APL pickers hold
references into the rotation tree, so they mutate in place and call `touchRotation()` themselves. A missed
`touchRotation()` leaves every reader stale, with no error.

**Targets.** `encounter.modifyTarget(i, draft => …)` hands the callback a clone and writes a new targets array. Mutate
only the draft: re-reading state inside the callback returns the old object and the edit disappears.

## Adding a field

1. Add it to its slice with a default in `sim_store.ts`. A player field also goes into `PLAYER_FIELDS` (the counters
   derive from that list), a reforge field into `REFORGE_FIELDS`.
2. The facade getter reads the store; the setter keeps an equality guard and writes through `patchSlice` /
   `patchKeyed`, passing the field in `bumps` if it must notify on an equal value.
3. React reads it through a plain selector or a `StoreField`; a non-React consumer calls the matching `subscribe*`
   helper and keeps the unsubscribe.
4. `npm run test:unit` covers the hooks' gating (`ui/sim/hooks/*.test.tsx`), but no suite asserts one notification per
   facade write, so count a new field's notifications yourself.

## Player lifetime

`Player.dispose()` unsubscribes its store reactions and drops its `players` / `reforge` / `statWeights` / `statSweep` /
`bulk` slices on the next tick. `Party.setPlayer` disposes a displaced player on the next microtask if it still has no party: moves
and swaps re-place the instance within the same task and are never disposed, and removals (`setPlayer(…, null)`) never
dispose. `Party.fromProto` spec changes and import temporaries call `dispose()` explicitly.
