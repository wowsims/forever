# Persistence and serialization

**Source of truth:** `ui/sim/state/persistence.ts` (load order and autosave), `ui/sim/state/serialization.ts` (the
`IndividualSimSettings` envelope), `ui/sim/state/sim_links.ts` (URL-hash import), `ui/sim/state/stat_filter_sharing.ts`
(stat filters stay with the viewer). The load order is written as a comment at the top of `persistence.ts`; when this
file and the code disagree, the code is right:

```
sed -n '1,10p' ui/sim/state/persistence.ts
```

All four are browser-free: storage and `location` arrive through the injected `Env` (`sim.env`), never through globals.

## The load order is a contract

`loadIndividualSettings(host, opts)` runs the sequence inside `sim.applyLoadedSettings(() => batch(...))`, which holds
back the character-stat recomputation until the load ends and then runs it once:

1. `host.applyDefaults()`
2. saved settings from `env.storage`, read with `fromJsonStringLenient` (`ui/sim/proto/lenient_json.ts`). A plain
   `ignoreUnknownFields` read leaves a oneof holding an enum name the protos do not list with an undefined value, and
   the next `toBinary` throws. Parse failures are warned and swallowed, so a corrupt blob cannot brick the page.
3. the URL-hash link, via `tryParseUrlLocation(env.location)`. It comes after step 2 so that a partial link (rotation
   only, say) keeps the saved settings for the other categories. `withViewerStatFilters` puts the viewer's stat
   filters into the link's envelope first.
4. `env.location.setHash('')`
5. `player.setName('Player')`
6. `opts.autosaveSubscribe(schedulePersist)`, last, so initialization does not re-store what it just loaded
7. `opts.statWeightSettings.load()`, then `opts.statSweepSettings.load()`

Then one explicit `persist()` outside the batch, because the autosave subscription only sees changes made after it
existed.

Autosave is debounced by `AUTOSAVE_DEBOUNCE_MS`, because serializing and storing the whole envelope on every keystroke
was a visible cost per APL edit. `env.onPageHide` flushes a pending write. `persistTimer` is declared before the load
batch on purpose: the batch's flush can already schedule a persist, and a later declaration is a TDZ error that only
fires on a page with saved settings.

## Storage keys

The caller builds the key and hands it in; the state layer never derives one. `SimHostObject.getStorageKey(postfix)`
(`ui/app/individual_sim_ui.tsx`) builds keys over `specStorageKey` (`ui/sim/state/storage_keys.ts`), which prefixes the
spec's key. The suffixes live in `ui/sim/state/storage_keys.ts` (the saved-data slots and
`STAT_WEIGHT_SETTINGS_STORAGE_KEY`) and in `persistence.ts` (`SETTINGS_STORAGE_SUFFIX` for the settings envelope).

A few keys skip the spec prefix, so every sim shares them: `SHARED_SAVED_ENCOUNTER_STORAGE_KEY` (saved encounters),
and in `ui/sim/sim.ts` the app-wide preferences `WASM_CONCURRENCY_STORAGE_KEY` and `SHOW_POWER_TOOLS_STORAGE_KEY`.
`Sim` reads those two in its constructor and writes them in their setters; they stay out of the settings envelope, so
a spec's Restore Defaults leaves them alone. A new key goes through `getStorageKey` unless it must be shared across
specs, and then says so in a comment, since the shared form is easy to copy by accident.

## The envelope

`individualSimSettingsToProto(ctx, exportCategories?)` and `applyIndividualSimSettings(...)` in `serialization.ts`
build and apply `IndividualSimSettings`; `SimHostObject.toProto` / `fromProto` wrap them. The envelope covers more than
the store: `IndividualSimSerializationContext` also carries the gear-optimizer settings and the EP reference-stat
selections. There are no load-time migrations of older envelopes.

Category filtering is shared with link import through `SimSettingCategories`. `LINK_DEFAULT_CATEGORIES`
(`sim_links.ts`) is every category except `UISettings`, so show-flags and language do not travel in a link.

**Stat filters belong to the viewer.** `DatabaseFilters.stat_filters` never travels: the link and JSON exporters
(`ui/features/import-export/exporters/link.ts`, `ui/features/import-export/exporters/json.ts`) strip it with
`withoutStatFilters` whatever the categories, and the JSON importer (`ui/features/import-export/importers/json.ts`)
and the link import write the viewer's filters into the envelope with `withViewerStatFilters` before `fromProto`.
Both helpers change the envelope they are given and return it. An envelope without filters needs nothing, since the
defaults it loads keep the viewer's. The autosaved envelope keeps them. There are five per-panel filters (EP,
character stats, buffs, debuffs, consumables), an unset one meaning the spec default; a pick a filter hides stays
offered, dimmed with a kept badge, and Restore defaults keeps the filters.

## Moving saved entries between browsers

`ui/features/saved-data-transfer/` exports and imports a saved panel's own entries, behind a cog in the panel's header
that shows with power tools on. The gear, talents, rotation, settings and encounter panels have it. EP weights, enchant
presets and Builds bundles do not.

The text is `wowsims-saved:v1:` followed by the base64 of zlib-wrapped deflate (`CompressionStream('deflate')`) of the
JSON `{ type, spec, entries }`. `type` is `gear`, `talents`, `rotation`, `settings` or `encounter`; `spec` is the `Spec`
enum name; `entries` is `Record<name, toJson(data)>`, the record `useSavedData` stores. A downloaded file holds the same
text. `model/transfer_text.ts` encodes and decodes it. A format change raises the version, and the import refuses a
version it does not know.

Both dialogs run on the shells in `ui/features/import-export/components`: `Exporter` with a picker as its children, and
`Importer` with a `preview`, the step any importer can add between the text and the import. Every refusal is an
`ImportProblems` under the text.

Names and contents come from other people, so the import treats them as untrusted:

- A name is trimmed, and an empty one, one over `MAX_NAME_LENGTH` or a `__proto__`-style one counts as unreadable.
- The text, the inflated JSON and the entry count are capped.
- An entry has to pass the panel's codec and, for gear, encounters and talents, the check in
  `model/saved_data_types.ts`. Gear with more items than their slots hold throws in `lookupEquipmentSpec`, which the
  Builds tab calls while it renders. A talents string goes into the Batch CSV as it stands, so it has to be digits and
  dashes.
- A saved name never goes into a string that `LocaleHtml` reads. A `FieldLabel` tooltip given as a string goes through
  `LocaleHtml`, so a tooltip quoting a saved name is passed as an element (`BulkCheckList`, `BulkBundleList`).
- A saved name in a CSV goes through `csvText` (`ui/sim/utils/csv.ts`), so a spreadsheet reads it as text. The guard is
  per column because a talents string can open with a dash.

An entry whose contents a saved entry holds, under any name, is unticked and labelled already saved, so importing a file
twice saves nothing new. Ticking it saves it under a free name. A name in use, a preset's name or the name of a stored
entry the codec rejects saves under the next free number, so an import never overwrites anything.

Gear and encounters import on any spec. Talents, rotations and settings are refused from another spec, the last because
they hold spec options. Several entries go in through `saveMany`: `save` reads the record as of the last render, so a
loop of them keeps only the last.

## What is gated

No suite pins the serialized shape. The persistence, import-export and shared-build tests go through
`serialization.ts` via the `makeIndividualSim` fixture (`ui/testing/sim/individual_sim.ts`), but they check what a
round trip keeps, not the envelope as a whole. A change that alters the envelope can pass every gate, so serialize a
default page before and after the change and diff the two.

Two `Sim` round-trip quirks matter for that diff:

- `Sim.toProto` collapses all-selected filter arrays to `[]` and `Sim.fromProto` re-expands them, so the form is a fixed
  point only from the second pass. Canonicalize once before comparing.
- `Sim.fromProto` mutates its argument's filters in place. Serialize before round-tripping.
