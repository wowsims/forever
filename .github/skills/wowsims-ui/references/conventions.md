# Conventions, spec authoring, and decisions already made

**Source of truth:** `.oxfmtrc.json` for formatting, `.oxlintrc.json` for what is enforced, `ui/STYLING.md` for
styling, component state and locating elements, `ui/README.md` ("How to author a spec") for the spec surface, and
`ui/specs/` itself for what a spec looks like.

## Formatting

`oxfmt` is the formatter; prettier is not a dependency. `npm run fmt` is `npx oxfmt . --check` over the whole repo,
including the markdown in `.github/skills/`, the workflows and the JSON schemas, so run it after editing any of them.
It skips `ignorePatterns` in `.oxfmtrc.json` (generated files, `assets/**` and the per-spec preset and talent JSON,
whose reformatting would be a diff nobody can review) and anything `.gitignore` hides. A new preset JSON directory
needs its own ignore entry. Print the config:

```
node -e "console.log(require('./.oxfmtrc.json'))"
```

Import order is `simple-import-sort` through oxlint, so `npm run lint:js:fix` sorts and `npm run fmt:fix` formats. Run
both on the files you touched and nothing else.

oxfmt also sorts Tailwind classes, per `clsx` argument and in the attributes `sortTailwindcss.attributes` lists. What
that means for writing a class list (one plain string on a `className`, `clsx` kept on a module-level const) is in
`ui/STYLING.md` §1.

## JSX

Every `.tsx` file is React: `tsconfig.json` sets `jsx: react-jsx` / `jsxImportSource: react`, and there is no per-file
pragma. A `.tsx` file with no JSX in it should be `.ts`.

## Generated files and fixtures

- Regenerate a `.results` file or a `.test.json` optimizer fixture only when the change means to move it, and never to
  make a gate pass. Diff first and ask before committing one.
- Never hand-edit anything under `ui/generated/` or any `*_auto_gen.ts`; regenerate through the makefile.

## Authoring a spec

`ui/README.md` has the long version. A spec is data: `ui/specs/<class>/<spec>/spec.ts` default-exports one
`defineSpec({...})`, beside `presets.ts` and, where needed, `inputs.ts`. There is no per-spec page, `index.ts` or
`SimHostObject` subclass: `SimHostObject` is concrete, takes a `SpecDefinition<S>`, and runs the behaviour slots
(`features`, then `reforge`, then `derivedSettings`) at the end of its constructor.

Adding a spec is three edits:

1. `ui/specs/<class>/<spec>/spec.ts`. Use `.ts`; promote to `.tsx` only if the spec renders a React element.
2. A `PlayerSpec` class in `ui/sim/player/specs/<class>.ts` with its `launch: { phase, status }`, exported through
   `ui/sim/player/specs/index.ts`. That field is the only record of launch status: the spec page's
   `ui/app/header/SimTitleDropdown/` and the landing page's class menu (`ui/app/landing/landing_classes.ts`) both read
   it from `PlayerSpecs`.
3. A theme block in `ui/styles/theme/specs.css`, in both halves: the class-grouped `--theme-color` lists and the
   per-spec `--theme-background-image` rule (`ui/STYLING.md` §1).

`tools/vite/spec_pages.mts` globs `ui/specs/*/*/spec.ts(x)`, so the page appears at `/forever/<class>/<spec>/` with no
build-config edit and no HTML file.

Preset data: gear sets, APLs and builds are JSON under the spec folder (`gear_sets/`, `apls/`, `builds/`), built into
presets by the `make*` helpers in `ui/app/preset_utils.ts`. EP weights, talents (one `talentsString`, since there are
no glyphs and no per-talent enum) and anything that needs a callback such as `onLoad` stay TS in `presets.ts`. Rules
that several specs of one class share live in `ui/specs/<class>/shared/`; only the classes that need it have one.

**Every new gearset passes `devOnly: true`** to `makePresetGear`, and a `makePresetBuildFromJSON` build takes that
gearset as its `gear`, until the gearset is released. Without the flag it shows on wowsims.com: `spec_entry.tsx` drops
dev-only gear outside the dev server and the Cloudflare Pages previews (`isPreviewOrDevMode` in `ui/sim/utils/env.ts`).
Removing the flag is the release, so a change does it only when asked to release that gearset.

Two traps:

- **A shared `DerivedSetting` must be typed `DerivedSetting<any>`.** `Player<S>` is invariant in `S`, so
  `DerivedSetting<A | B>` is not assignable to `DerivedSetting<A>`. Annotate the callback's `player` parameter to keep
  the body checked against the union.
- **A custom settings section is data, never DOM.** Declare it in `sections` as a `CustomSection`
  (`ui/sim/spec_config.ts`); `ui/features/settings/components/CustomSection/CustomSection.tsx` renders it from
  `ui/app/tabs/SettingsTabBody.tsx` through the same path as the standard sections. `CursesSection` in
  `ui/specs/warlock/dps/inputs.ts` is a worked example.

## Localization

`en` is the only locale. A new user-facing string goes into `assets/locales/en/`, and its key needs a matching property
in `schemas/<name>.schema.json` (`verification.md`, "Locales"). With no second locale, that is the whole change.

## Code that looks dead but is not

- There is no raid sim UI. The individual sim models a one-party raid, which is why `Raid` and `Party` exist in
  `ui/sim`.
- `PartyBuffs` (`proto/buffs.proto`) is a populated message, and its party-buff slice, subscription and picker are live.
- A kit option with no production caller may have been requested ahead of use (Drawer `side`, the TabNav badge): list
  such options and ask before removing kit API.
- A container's exported view, the exports the design-system pages use and Popover `container` are settled exceptions
  to the rule that tests and stories add no production code (PR #187).

## Decisions already made

Reopen one only with new evidence.

| Decision                                                            | Why                                                                                                                                                                                                                                                                                                                  |
| ------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `import/no-cycle` is off                                            | the existing cycles are a triage project (`layers.md`); review catches new ones                                                                                                                                                                                                                                      |
| Aliases are `tsconfig` `paths`                                      | `tsc` under `moduleResolution: "bundler"` does not resolve `package.json` `#foo/*` imports                                                                                                                                                                                                                           |
| Only `ui/app` may `createRoot`                                      | one React root per page, not one per leaf; lint-banned in the kit and features                                                                                                                                                                                                                                       |
| `sections` is the only custom-section shape                         | one renderer                                                                                                                                                                                                                                                                                                         |
| `oxfmt`, not prettier                                               | one formatter for the repo                                                                                                                                                                                                                                                                                           |
| The kit is controlled; the store binding is in `ui/features/inputs` | a kit picker takes `value` / `onChange` and never names `@sim` state, so it renders in a story with no host; `InputConfig`, `useInput` and the `Connected*` adapters are the one place a picker meets the store                                                                                                      |
| Storybook is the component registry                                 | the catalogue comes from the code, so there is no hand-kept list to go stale                                                                                                                                                                                                                                         |
| Stories run as tests                                                | `ui/testing/storybook/stories.test.tsx` mounts every story in `npm run test:unit`, so a story cannot rot unnoticed                                                                                                                                                                                                   |
| One stylesheet body, `ui/styles/main.css`, with two entries         | Tailwind merges `@source not` across the whole import graph, so production excludes stories, `ui/testing/storybook` and `ui/design-system` in one entry and Storybook adds them back in the other (`ui/STYLING.md` §7); after an `@source` change, check both bundles                                                |
| Views are props-only; containers keep the hooks                     | a view renders in a story from plain props, and the store reads stay in one greppable place per feature (`ui/README.md`, "Placement rules")                                                                                                                                                                          |
| Component state is a `data-*` attribute                             | styling and tests key on the same attribute (`data-kept`, `data-customised`, …), never on a state class (`ui/STYLING.md` §3)                                                                                                                                                                                         |
| Tests pin behaviour and props, never classes                        | a test asserts what a user action or a prop changes: output, a callback, a store read or write, a `data-*` or ARIA attribute. No "it renders" case where a story exists, no case that only exercises Base UI, and one behaviour is tested at one layer. One test file per component unless its `vi.mock` set differs |
| UI tests drive the top-level feature component                      | the dialog or panel a user opens, through its real surface; its sub-components get stories only, and `model/` logic keeps its unit tests                                                                                                                                                                             |
| Presentational kit props get a story, not a test                    | a border, padding or a variant's data attribute is shown in a story or matrix; tests pin behaviour                                                                                                                                                                                                                   |
| One folder per feature component                                    | every component under `ui/features/<x>/components/` gets its own folder with its story, a dialog's sub-components included                                                                                                                                                                                           |
| Only base scale tokens hold literals                                | spacing, typography, colour, radius and shadow are the base scale (`ui/styles/theme/spacing.css` and its siblings); every component, primitive or feature token aliases a base token, and a value with no base match is raised, not added as a literal                                                               |
