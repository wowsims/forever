# Conventions, spec authoring, and decisions already made

**Source of truth:** `.oxfmtrc.json` for formatting, `ui/README.md` for authoring a spec,
`ui/STYLING.md` for styling/state/locating elements, `.oxlintrc.json` for anything that is actually
enforced.

## Formatting

`oxfmt` is the formatter, not prettier — prettier is not a dependency of this repo. `npm run fmt` is
`npx oxfmt . --check`, so its scope is the whole repo: the markdown in `.github/skills/`, the
workflows, `docker-compose.yml` and the JSON schemas are formatted too, and CI checks them. What it
skips is `ignorePatterns` in `.oxfmtrc.json`, all of it data rather than source: `assets/**`, whose
database is written by `gen_db`; `**/test-fixtures/*.json`; and the per-spec preset JSON and talent trees. Anything
`.gitignore` already hides is skipped as well, which is how generated output stays out.

```
node -e "console.log(require('./.oxfmtrc.json'))"
```

Today that is tabs, width 4, `printWidth` 160, single quotes, semicolons, trailing commas
everywhere, `arrowParens: "avoid"`. Import order is `simple-import-sort` through oxlint, so
`npm run lint:js:fix` sorts and `npm run fmt:fix` formats; run both on files you touched and on
nothing else.

oxfmt also sorts Tailwind classes, and **each `clsx` argument is sorted separately** — an argument
boundary is a sort barrier. So `clsx('text-red-500 flex', 'p-2 items-center')` formats to
`clsx('flex text-red-500', 'items-center p-2')`: sorted within each group, never across them. On a
`className` (or any attribute in `sortTailwindcss.attributes`) a static list should therefore be one
plain string — `className="flex items-center p-2 text-red-500"` — which gets a single canonical
sort and drops a render-time call. The inverse holds on a **module-level const**, where a bare
string is not sorted at all: there `clsx` is the only reason the formatter looks at it, so
`const ROW_CLASS = clsx(…)` in `ui/features/gear/components/SelectorModal/ItemList.tsx` keeps its
call and must not be "simplified" away. Use `clsx` when an argument is conditional or a variable,
never to wrap or line-break a fixed list.

`.oxfmtrc.json` deliberately ignores the preset JSON (`ui/**/apls/*.apl.json`,
`ui/**/builds/*.build.json`, `ui/**/gear_sets/*.gear.json` — check the current list with the command
above, since a new preset JSON category needs its own entry). Reformatting one of those is a diff
nobody can review. Its ignore list also names `ui/worker/highs.js`, which does not exist — the solver
comes from the npm `highs` package.

Stylesheets: Tailwind utilities, no SCSS or Bootstrap anywhere in `ui/`. A component that needs
reusable multi-utility classes or state/descendant selectors gets a co-located
`ui/ui-kit/<Name>/<Name>.css` of `ui-*` `@apply` classes, `@import`ed from `ui/styles/main.css`.
`npm run lint:css` is stylelint over `ui/**/*.css`. Full conventions — token policy, `ui-*` classes,
`data-*` state, locating elements, the class-hook gates — are in `ui/STYLING.md`.

## JSX dialect

React is the **only** dialect. `tsconfig.json` is `jsx: react-jsx` / `jsxImportSource: react` and
both vite configs use the automatic runtime, so every `.tsx` file is a React file and there is no
per-file opt-out: there is no `@jsxImportSource` pragma and no `tsx-vanilla` package.
A `.tsx` file with no JSX in it should be `.ts`.

## Fixtures and generated files

- **Never** regenerate a `.results` file or a `.test.json` reforge fixture to make a gate pass,
  and never commit one without asking. Diff first, then ask.
- **Never** hand-edit anything under `ui/generated/` or any `*_auto_gen.ts` — regenerate through the
  makefile.
- **Never** run `gen_db` concurrently with another copy of itself.

## Authoring a spec

`ui/README.md`'s "How to author a spec" is the long version; this is the shape and the traps. The
README is enforced by nothing, so prefer this file and the code where they disagree.

A spec is **data**. `ui/specs/<class>/<spec>/spec.ts` default-exports one `defineSpec({...})` and is
the only code file the spec owns besides `presets.ts` / `inputs.ts`. There is no `sim.ts`, no
per-spec `index.ts`, and no `SimHostObject` subclass anywhere — `SimHostObject` is concrete and
takes a `SpecDefinition<S>`, running the behaviour slots (`features` → `reforge` →
`derivedSettings`) as its last constructor statements.

Adding one is three edits and no page:

1. `ui/specs/<class>/<spec>/spec.ts`. Use `.ts`: **no spec carries real JSX.** Promote to `.tsx`
   only if a spec genuinely renders a React element.
2. A `PlayerSpec` class in `ui/sim/player/specs/<class>.ts` with its `launch: { phase, status }`,
   exported through `ui/sim/player/specs/index.ts`. That field is the single source of truth for
   launch status: `ui/app/header/SimTitleDropdown/` reads it on a spec page, and the landing page
   (`ui/app/landing/`, entered from `ui/app/landing_entry.tsx`) builds its class menu and its per-spec
   status badges from `PlayerSpecs` via `ui/app/landing/landing_classes.ts`, not a hand-written list.
3. A theme block in `ui/styles/theme/specs.css`. Every spec has a `--theme-color` shared per class
   and its own background image (`dps-warrior` and `protection-warrior` use two different files), so
   the file has two halves: class-grouped selector lists carry
   `--theme-color`/`--theme-color-foreground`, and a separate per-spec rule carries
   `--theme-background-image`. A new spec needs an entry in both halves (see `ui/STYLING.md`).

`tools/vite/spec_pages.mts` globs `ui/specs/*/*/spec.ts(x)`, so the page at `/forever/<class>/<spec>/`
appears with no build-config edit and no html anywhere. Check the count with
`/usr/bin/find ui/specs -name 'spec.ts' -o -name 'spec.tsx' | wc -l`.

Two traps:

- **A shared `DerivedSetting` must be typed `DerivedSetting<any>`.** `Player<S>` is invariant in
  `S`, so `DerivedSetting<A | B>` is not assignable to `DerivedSetting<A>` and tsc rejects it via
  `autoRotationGenerator`. Annotate the callback's `player` parameter to keep the body checked
  against the union. The input helpers _are_ generic-friendly across a union; only `DerivedSetting`
  is not.
- **A custom settings section is data, never DOM.** Declare `sections: [CustomSection]` (typed in
  `ui/sim/spec_config.ts`) and `ui/features/settings/components/CustomSection/CustomSection.tsx`
  renders it, from `ui/app/tabs/SettingsTabBody.tsx`, through the same `ContentBlock` + picker path
  the standard sections use. `sections` is the only shape.
  `ui/specs/shaman/shared/inputs.ts` is the worked example.

Rules shared by several specs of one class live in `ui/specs/<class>/shared/`
(`{inputs,presets,derived}.ts`). Constants used by more than one class — the melee-hit/expertise and
spell-hit `statCaps` builders, and the shared encounter protos (single-target and whichever raid
encounters more than one class's presets reference) — live in
`ui/sim/presets/{stat_caps,encounters}.ts` and export raw protos/`Stats` for a class's
`shared/presets.ts` to wrap, because `ui/sim` cannot import `@app`.

Proto-serialisable preset data lives as JSON, never a TS literal: gear, APLs, builds, EP weights and
talents under `ui/specs/<class>/<spec>/`. EP JSON stores enum fields **by name** (`"StatCritRating"`)
so the file survives a proto regeneration. `SavedTalents` is a single `talentsString` field
(no glyphs, no per-talent enum), so a talent preset is just that string, not enum-resolved JSON.
Anything that references a TS symbol or a callback — a computed EP preset via `.withStat()`, an
`onLoad` handler — stays a TS literal.

## Localization

The tree ships one locale, `en`. Every new user-facing string ships in `assets/locales/en/`, and a new key
needs a matching property in `schemas/<name>.schema.json` (`additionalProperties: false`). See
`verification.md`. There is no second locale to keep in register with, so a key added to `en` is the
whole change.

## Quirks that look like bugs

- There is no raid sim UI — the individual sim models a one-party raid, and the
  `Raid` and `Party` model classes in `ui/sim` exist for that reason, not because a fuller raid sim UI
  was cut.
- `PartyBuffs` (`proto/buffs.proto`) is a real, populated message here (`ferocious_inspiration`,
  `blood_pact`, `moonkin_aura`, `leader_of_the_pack`, …), and its party-buff slice, subscription and
  picker are live code, not vestigial paths to delete.

## Decisions already made — do not re-litigate without new evidence

| Decision                                                                                            | Why                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| --------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `import/no-cycle` is off                                                                            | 58 cycles (43 `ui/sim`, 14 `ui/features`, 1 `ui/i18n`) — a triage project, see `references/layers.md`. Cycles are caught by review                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Aliases are `tsconfig` `paths`, not `package.json` `imports`                                        | `tsc` under `moduleResolution: "bundler"` does not resolve `#foo/*`, so aliases live in `tsconfig` `paths`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| Only `ui/app` may `createRoot`                                                                      | lint-banned in ui-kit and features; one React root per page, not one per leaf                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `sections` is the only custom-section shape                                                         | one renderer, not two                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| Files move by hand                                                                                  | there is no move tool; move with `git mv` and re-sort imports (see `layers.md`)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `oxfmt`, not prettier                                                                               | prettier is not a dependency; do not add a second formatter for `ui/`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| The kit is controlled; the store binding is in `ui/features/inputs`                                 | a kit picker takes `value`/`onChange` and never names `@sim` state, so it renders in a story with no host; `InputConfig`, `useInput` and the `Connected*` adapters are the one place a picker meets the store (lint-enforced, see `references/layers.md`)                                                                                                                                                                                                                                                                                                                                               |
| Storybook is the component registry                                                                 | the kit components, feature views and the app's presentational pieces have stories, so the catalogue is generated from the code instead of a hand-kept list; there is no separate registry document or skill                                                                                                                                                                                                                                                                                                                                                                                            |
| Stories run as tests under happy-dom                                                                | `ui/testing/storybook/stories.test.tsx` mounts every story in `npm run test:unit`, so a story cannot rot unnoticed; an opt-out carries a `smokeReason`                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| The stylesheet body is `ui/styles/main.css`; `style.css` and `.storybook/storybook.css` are entries | one body, two `@source` sets: production excludes stories, `ui/testing/storybook` and `ui/design-system`, Storybook adds them back (`ui/STYLING.md` §7)                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| Views are props-only; containers keep the hooks                                                     | a view renders in a story from plain props, and the store reads stay in one greppable place per feature (`ui/README.md`, "Placement rules")                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| Tests pin behaviour and props, never classes                                                        | a test asserts what a user action or a prop changes: output, a callback, a store read or write. State is asserted through `data-*` or ARIA attributes, so a component with class-only state gets a `data-*` attribute beside the class. No "it renders" case where a story exists (the story smoke test mounts it), no case that only exercises Base UI, and one behaviour is tested at one layer: the kit component owns its interaction, a `Connected*` adapter only its binding, a container only its wiring when its view has a test. One test file per component, unless its `vi.mock` set differs |
