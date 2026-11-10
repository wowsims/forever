---
name: wowsims-ui
description: "Work on the WoWSims Forever frontend, ui/ (React 19, TypeScript, Zustand, Base UI, Tailwind, Storybook). Use it for any change under ui/ or .storybook/, even a small one: a component, view or story, a kit picker or a Connected* adapter, a store field, subscription or hook, saved settings and links, an import the layer lint rejects, a spec page under ui/specs, a locale string, and before calling a UI change done (type-check, lint, vitest, the Storybook build, real-sim screenshots). Bulk sim work, its feature folder included, is wowsims-bulk-sim, and the gear optimizer's model (EP, stat caps, HiGHS) is wowsims-reforge-optimizer; the optimizer panel's layout and store wiring stay here."
---

# wowsims-ui

`ui/` is the frontend: React 19 and TypeScript, built by vite, no server-side rendering. The sim is Go, compiled to
WASM or served by the Go dev server behind `ui/worker/*`, and the UI speaks protobuf to it, so nothing in `ui/` models
combat.

Comments, docs, commit messages and PR text follow `wowsims-conventions`. The rule that tests and stories add no code to
`ui/` is in `AGENTS.md`; `references/layers.md` says where their fixtures go instead.

## Layers

Each layer imports only from the layers to its left. `ui/sim` and `ui/i18n` are peers and may import each other.
`.oxlintrc.json` enforces the direction, so a violation fails `npm run lint:js`; the exact bans and the ways they
surprise people are in `references/layers.md`.

```
generated → worker → {sim, i18n} → ui-kit → features → app → specs → pages
```

| Directory                  | Alias        | What lives there                                                                                                                                         |
| -------------------------- | ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ui/generated/`            | `@generated` | protobuf-ts output; git-ignored (like `sim/core/proto/*.pb.go`): regenerate, never hand-edit or commit it                                                |
| `ui/worker/`               | `@worker`    | the worker entries and the Go package that embeds `highs.wasm`                                                                                           |
| `ui/sim/`                  | `@sim`       | the model: the `Sim`/`Player`/`Raid`/`Encounter` facades, the Zustand store (`state/`) and its React bindings (`context/`, `hooks/`); no browser globals |
| `ui/i18n/`                 | `@i18n`      | i18next config and the label tables                                                                                                                      |
| `ui/ui-kit/`               | `@ui-kit`    | controlled, sim-agnostic widgets, one folder per component, each with a story                                                                            |
| `ui/features/<x>/`         | `@features`  | one folder per capability: `model/` (no DOM, no React), `components/`, `hooks/`; `ui/features/inputs/` binds the kit's pickers to the store              |
| `ui/app/`                  | `@app`       | the composition root: `SimHostObject` (`ui/app/individual_sim_ui.tsx`), the shell, tabs, header and landing page                                         |
| `ui/specs/<class>/<spec>/` | `@specs`     | spec data, one `spec.ts` per spec page                                                                                                                   |

Outside the arrow: `ui/styles/` (see `ui/STYLING.md`), `ui/design-system/` (the Storybook docs pages), `ui/testing/`
(helpers and fixtures for tests and stories only, alias `@testing`), `ui/shared/`, `ui/types/`, `ui/tracking/`, and
`.storybook/` at the repo root. `ui/README.md` has the full tree.

## What to build from

Look in Storybook before writing a component (`npm run storybook:dev`): every kit component, feature view and
presentational app piece has a story, so the stories are the catalogue. The view is built from Base UI
(`@base-ui/react`, every overlay and form primitive), react-tooltip behind `ui/ui-kit/Tooltip/`, TanStack Table and
Virtual for the results tables and `ui/ui-kit/VirtualList/`, react-i18next (`<Trans>` only where a locale string
carries markup) and Zustand. Styling is Tailwind utilities, co-located `ui-*` classes and `data-*` state attributes, all
in `ui/STYLING.md`; there is no SCSS or Bootstrap.

A few facts that shape the code: several classes have one spec page (`mage/dps`, `hunter/dps`), there are no glyphs,
`en` is the only locale, talents are three trees with tier and prerequisite rules (the picker lets a point be taken back
anyway and marks what that breaks, see `references/talents.md`), and there is no raid sim UI.

## Which file to open

| You are…                                                                                                     | Open                            |
| ------------------------------------------------------------------------------------------------------------ | ------------------------------- |
| moving a file, adding an import that lints, or deciding which directory something belongs in                 | `references/layers.md`          |
| adding a store field, wiring a picker, or chasing a component that renders stale data or re-renders too much | `references/state.md`           |
| touching saved settings, the URL hash, autosave, export/import or `IndividualSimSettings`                    | `references/persistence.md`     |
| about to call a change done, or reading a red or flaky gate                                                  | `references/verification.md`    |
| running real sims in a browser, taking screenshots, or running Storybook                                     | `references/running-locally.md` |
| adding a spec, formatting code, or about to reopen a decision that was already made                          | `references/conventions.md`     |
| touching the talent picker, the talents string, a talent preset, or the broken-talent rules                  | `references/talents.md`         |
| writing a component, a container and its view, or a story                                                    | `ui/README.md`                  |
| styling, component state as `data-*`, locating elements in a test, or a class gate failing                   | `ui/STYLING.md`                 |

## Done

`npm run type-check`, `npm run lint:js`, `npm run lint:css`, `npm run fmt`, `npm run test:unit`, `npm run test:locales`
and `npm run storybook:build`, all from the repo root after `npm ci`. `references/verification.md` says what each one
catches and what none of them do: no gate shows that a page rendered, so a change to rendering, layout or interaction
also gets real-sim screenshots (`references/running-locally.md`).

On this machine the shell wraps `diff`, `grep`, `find` and dev commands in `rtk`, which can compress output into a
summary. Use `/usr/bin/diff` (the wrapped one has reported different files as identical), and prefix `RTK_DISABLED=1`
or use `rtk proxy` when you grep output for a count, because a summarised line greps as zero.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The "You changed →
Update" table and the path check are in `references/maintenance.md`. Every reference opens with a _Source of truth_
line naming the code that defines the behaviour and a command that re-derives the claim. Copy no code and no counts
that the command can print.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill otherwise. A memory file is a pointer
to this skill, not a second copy.
