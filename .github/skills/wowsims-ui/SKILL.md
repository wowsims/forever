---
name: wowsims-ui
description: 'Work on the wowsims Forever frontend, ui/ (React 19 + TypeScript). Use when touching ui/sim (the Sim/Player/Raid/Encounter facades, the Zustand store in ui/sim/state, the store hooks in ui/sim/hooks, batching, persistence, the IndividualSimSettings envelope), ui/ui-kit, ui/features, ui/app, ui/i18n, ui/specs, ui/generated/proto, a story, ui/design-system or .storybook; when an import trips the oxlint layer rules; or when a UI change needs verifying (type-check, lint, vitest, a dev-server or real-sim smoke). Start here, then open the one file under references/ that owns the subject — the routing table is in this file.'
---

# wowsims-ui

Comments, docs, commit messages and PR text follow `wowsims-conventions`: "game files", not "client", no history in comments, no "golden".

`ui/` is the frontend: TypeScript, built by vite, no server-side rendering. The sim engine is
Go compiled to WASM behind `ui/worker/*`; the UI only ever speaks protobuf to it, so nothing in
`ui/` models combat.

Facts about this tree that change the docs, not just the paths: several classes have a single spec
page (one `mage/dps`, one `hunter/dps`), there are no glyphs anywhere in the proto or the locale set,
`en` is the only shipped locale, and talents are three free-form trees.

The view layer is **React 19**, and it is the only dialect: no file carries a `@jsxImportSource`
pragma, and there is no Bootstrap, no tippy and no `.scss` file anywhere in `ui/`. Both HTML entries
link a single `ui/styles/style.css`, and `form-control`/`d-none`/`btn` are retired class names, not
classes to reach for. Styling is Tailwind
utilities plus co-located `ui-*` composition classes — see `ui/STYLING.md`.

What the view is built from — reach for the existing thing before writing a new one:
**Base UI** (`@base-ui/react`) for Dialog, Menu, Popover, Tabs and Toast, and also for the Field,
Input, Button and Progress primitives the pickers are built on; **`react-tooltip`** behind
`ui/ui-kit/Tooltip/`; **`@tanstack/react-table`** and **`@tanstack/react-virtual`** for the results
tables, the log runner and `ui/ui-kit/VirtualList/`; **`react-i18next`** — `<Trans>` only where a
locale string carries markup, `i18n.t` everywhere else; and **Zustand**, which is the whole subject
of `references/state.md`.

**This skill documents everything below the view layer.** Storybook is the component registry:
every kit component, the feature views and the app's presentational pieces have a
`*.stories.tsx`, and `npm run storybook:dev` shows what exists before you write another. `ui/README.md`
has the kit contract, the container/view rule and the story conventions; `ui/STYLING.md` has the
co-located CSS.

## Layer map

Each layer may import only from layers to its left. `ui/sim` and `ui/i18n` are peers: both are
leaves everything above may depend on, and they may depend on each other.

```
generated → worker → {sim, i18n} → ui-kit → features → app → specs → pages
```

| Directory                  | Alias        | What lives there                                                                                                                                                                                                                                                                |
| -------------------------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ui/generated/`            | `@generated` | protobuf-ts output (`proto/**`) and `*_auto_gen.ts`. Tool output — never hand-edit, never lint (`ignorePatterns`)                                                                                                                                                               |
| `ui/worker/`               | `@worker`    | the three worker entries (`local_worker`, `net_worker`, `sim_worker`) plus the Go package that `go:embed`s `highs.wasm`                                                                                                                                                         |
| `ui/sim/`                  | `@sim`       | node-runnable model: the facades, `state/` (Zustand), `player/`, `raid/`, `settings/`, `talents/`, `presets/`, plus the React bindings in `context/` and `hooks/`. DOM-free and browser-global-free, **not** React-free                                                         |
| `ui/i18n/`                 | `@i18n`      | LEAF: i18next config, entity/label tables, `localization.ts`                                                                                                                                                                                                                    |
| `ui/ui-kit/`               | `@ui-kit`    | sim-agnostic, controlled React widgets, one folder per component (`TextField/`, `Field/`, `Table/`, …) with a story and a test each, plus `hooks/` and `utils/` — no `@sim` except `utils/**`, `constants/**`                                                                   |
| `ui/features/<x>/`         | `@features`  | twelve capabilities; `model/` is DOM-free _and_ React-free, `components/` + `hooks/` are React. Beside them, `ui/features/inputs/` binds the kit's pickers to the store (`InputConfig`, `useInput`, `Connected*`) and `ui/features/hooks/` holds shared hooks (`useSavedPanel`) |
| `ui/app/`                  | `@app`       | composition root: `SimHostObject` (`individual_sim_ui.tsx`), `SimApp`/`SimShell`/`SimTabs`, `header/`, `tabs/`, `landing/`, `spec_entry.tsx`, `browser_env.ts`                                                                                                                  |
| `ui/specs/<class>/<spec>/` | `@specs`     | spec data — one `spec.ts` per spec, may import everything                                                                                                                                                                                                                       |

Not layers, and not in the arrow: `ui/styles/` (`main.css` is the stylesheet, `style.css` its
production entry; tokens and the 18 spec themes as CSS variables — see `ui/STYLING.md`),
`ui/shared/` (two browser helpers — `pointer.ts` and `page_boot.ts`; the DOM helpers a component
reaches for are `ui/ui-kit/utils/dom.ts`), `ui/design-system/` (the twelve Storybook docs pages,
their `tokens/` parsers and `components/`), `ui/testing/` (`tailwind/`, the class checkers;
`storybook/`, the decorators, fixtures and the `stories.test.tsx` smoke test; `react.tsx`,
`renderWithHost`; tests and stories import it as `@testing/*`, production code may not), `ui/types/` (ambient `.d.ts`), `ui/tracking/` (the analytics shim), and
`ui/index.html` / `ui/index_template.html` at the root. `.storybook/` at the repo root is the
Storybook config (`main.ts`, `vite.config.ts`, `preview.tsx`, `storybook.css`). The React entry
points are `ui/app/spec_entry.tsx` and `ui/app/landing_entry.tsx`.

The direction is **enforced, not conventional** — `.oxlintrc.json` fails the build on a violation.
Read the exact bans, and the three ways they surprise people, in `references/layers.md` before
arguing with one.

## Which reference to open

| You are…                                                                                                     | Open                            |
| ------------------------------------------------------------------------------------------------------------ | ------------------------------- |
| moving a file, adding an import that lints, or deciding which directory something belongs in                 | `references/layers.md`          |
| adding a settings field, wiring a picker, or chasing a component that renders stale data or re-renders twice | `references/state.md`           |
| wiring a picker to a store field, or adding a `Connected*` adapter in `ui/features/inputs`                   | `references/state.md`           |
| touching saved settings, the URL hash, autosave, or `IndividualSimSettings`                                  | `references/persistence.md`     |
| about to call a change done, or wondering what CI actually runs                                              | `references/verification.md`    |
| running the sim in a browser, in a fresh worktree, or measuring a perf regression                            | `references/running-locally.md` |
| running or building Storybook, or changing `.storybook/`                                                     | `references/running-locally.md` |
| adding a spec, formatting code, or about to re-litigate a decision that was already made                     | `references/conventions.md`     |
| writing a component, a view or a story, or a design-system page                                              | `ui/README.md`                  |
| styling a component, adding state, or locating an element in a test/tool                                     | `ui/STYLING.md`                 |

## The short version of "done"

All from the repo root, all needing `npm ci` first:

```
npm run type-check     # tsc --noEmit over the whole repo, tools/ included
npm run lint:js        # oxlint on ui/ and .storybook/ with --max-warnings 0 — zero no-restricted-imports allowed
npm run lint:css       # stylelint ui/**/*.css
npm run fmt            # oxfmt . --check — the whole repo, not just ui/
npm run test:unit      # vitest + happy-dom, ui/**/*.test.ts(x) — every story mounts as a test too
npm run test:locales   # ajv, assets/locales/** against schemas/**
npm run build-storybook  # storybook build → storybook-static/
```

`references/verification.md` says what each command actually covers, which of them CI runs, and
which gate catches which class of mistake.

## Working on this tree

The whole repo type-checks, and every gate in `references/verification.md` runs against the layout
above. Two habits are about this machine:

- **Every diff uses `/usr/bin/diff`.** The bare `diff` here is wrapper-intercepted and has produced a
  false "files are identical" reading — never trust it. The same goes for `grep` and `find`, and
  `rtk` compresses some command output: if a result looks summarised rather than raw, re-run it
  through `rtk proxy` or redirect to a file and read that.
- **Rename detection needs the full `-- ui/` pathspec.** Files move across top-level `ui/`
  directories, so a pathspec naming only a source directory filters away every destination and
  reports plain deletions with zero renames — a convincing-looking "this file has no ancestor" that
  is purely an artefact of the query. Compare `<base>..<branch> -- ui/`.

## Keeping this skill true

Prose that nothing checks rots. Two rules keep this skill true.

**Point at code, not at copies of code.** Every reference file opens with a _Source of truth_ line
naming the file that defines the behaviour and a command that re-derives the claim. When the two
disagree, the code is right and the reference is a bug — fix it in the same commit as the code
change, the way you would a stale comment.

**Route by kind of change, so nobody has to re-read everything:**

| You changed…                                                                                                                                                                                                           | Update                                         |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| `.oxlintrc.json`, `tsconfig.json` `paths`, or a top-level `ui/` directory                                                                                                                                              | `references/layers.md` and the layer map above |
| `ui/sim/state/sim_store.ts`, `subscriptions.ts`, `batch.ts`, `events.ts`                                                                                                                                               | `references/state.md`                          |
| `ui/sim/hooks/*`, `ui/sim/context/SimHostContext.tsx`                                                                                                                                                                  | `references/state.md`                          |
| `ui/features/inputs/**` — the binding layer: `input.ts`, `hooks/useInput.ts`, `hooks/useConnectedField.ts`, `hooks/useHideZeroRestore.ts`, `configs.ts`, `input_helpers.ts`, `icon_inputs.ts`, `components/Connected*` | `references/state.md`                          |
| `ui/sim/state/persistence.ts`, `serialization.ts`, `sim_links.ts`                                                                                                                                                      | `references/persistence.md`                    |
| a `package.json` script or `.github/workflows/run_tests.yml`                                                                                                                                                           | `references/verification.md`                   |
| `vite.config.mts`, `vite.build-workers.mts`, `tools/vite/spec_pages.mts`, the makefile's dist targets                                                                                                                  | `references/running-locally.md`                |
| `.storybook/**`                                                                                                                                                                                                        | `references/running-locally.md`                |
| `.oxfmtrc.json`, `ui/specs/**` authoring, or a decision recorded as settled                                                                                                                                            | `references/conventions.md`                    |
| the kit contract, the container/view split, `*.stories.tsx` conventions, `ui/testing/storybook/**`, `ui/design-system/**`                                                                                              | `ui/README.md`                                 |
| `ui/styles/*.css`, a `ui-*` class, a class-hook gate, or how elements are located in tests/tools                                                                                                                       | `ui/STYLING.md`                                |

Then run the path check, which fails on any file this skill names that no longer exists:

```
.github/skills/wowsims-ui/check_paths.sh
```

There is no change log here. Facts about how the UI works go in a reference; the record of when
they changed stays in git (`git log --oneline -- .github/skills/wowsims-ui/`).
