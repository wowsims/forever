# Styling, state and locating elements

Source of truth: `ui/styles/main.css` (the stylesheet body: layer order, `@import`s, `@utility`,
`@source`), the two entries that import it — `ui/styles/style.css` (production, which adds four
`@source not` lines) and `.storybook/storybook.css` (Storybook, which adds the matching `@source`
lines for three of them) — `ui/styles/theme/` (tokens, split by type — `colors.css`, `spacing.css`, `breakpoints.css`,
`typography.css`, `effects.css`, `z-index.css`, `vars.css`, `specs.css`, imported via `index.css`),
`ui/styles/base.css` (element defaults), `ui/styles/vendor.css` (the handful of selectors with no
JSX element to carry them), `ui/testing/tailwind/canonical-classes.mjs` /
`ui/testing/tailwind/class-hooks.mjs`, and the two tests that run them, `ui/canonical_classes.test.ts`
and `ui/no_class_hooks.test.ts`.

There is no SCSS and no Bootstrap anywhere in `ui/` (`find ui -name '*.scss'` is empty; neither
HTML entry links anything but `style.css`; icons are inline SVG from `@fortawesome/react-fontawesome`,
and `main.css` imports FontAwesome's stylesheet into the base layer). Every class in `ui/`
markup is a Tailwind utility, a `ui-*` composition class, an entry of the fixed allowlist
`ui/class_hook_allowlist.json` (`sim-ui`, `<spec>-sim-ui`, `group`/`peer` and their named forms,
and the react-tooltip class-as-theme names), or a
`data-testid`/`data-*` used only by tests and tools. `npm run test:unit` enforces it — see "The
gates" below for what each check covers.

## 1. Styling with Tailwind

Reach for a **named utility** first, including its variants and `!`-free forms. Arbitrary values
(`bg-[#123456]`, `text-[rgb(0,0,0)]`, `p-[10px]`) are a last resort, not a shortcut:

- **No arbitrary colours.** Every colour is a `--color-*` token in `ui/styles/theme/colors.css`'s
  `@theme static { … }` block, so it renders as `bg-primary`, `text-danger`, `border-surface-border`,
  and so on. If the colour you need has no token, add one to `@theme static` (named by role or hue,
  e.g. `--color-grey: #808080`) rather than writing `bg-[#808080]` — never drop a declaration because
  no token exists yet either; a missing visible property is a bug, not a simplification.
- **No arbitrary spacing where a scale step exists.** The spacing rhythm in this repo is even steps
  plus 1 — `1, 2, 4, 6, 8, 10, 12, 16, …`. The root is 14px below `fhd` (1921px) and 16px from `fhd`
  on, so one step is 3.5px or 4px and `p-2` is 7px or 8px. Conversions read a source px value at the
  14px root.
  Convert a source px value to the **nearest even step** (ties round down), rather than writing `[Npx]`
  — see §7 for the same rule at the rendering-diff gate. `gap-[3.5px]` → `gap-1`, `7px` → `p-2`,
  `14px` → `w-4`; a sub-step value that stands for a hairline (a border, a divider, a ruler tick)
  converts to the fixed `px` step instead — a `1.75px` border becomes `px`, not a rounded step. Keep a
  literal `[Npx]` only where no step and no hairline fits, and prefer a token over a repeated arbitrary
  value. `1rem` follows `ui/styles/base.css`'s `--font-size-root`, so a converted step scales with the
  root — a size the JavaScript relies on (the timeline, the log runner's virtual
  list) is pinned to an exact px value with a named `@theme` token instead, so it never scales with the
  root.
- **Canonical spelling only.** `node ui/testing/tailwind/canonical-classes.mjs` runs Tailwind's own
  `designSystem.canonicalizeCandidates()` — the same function IntelliSense's "the class X can be
  written as Y" lint uses — over every `className`/`clsx`/`*ClassName` string and every `@apply` in
  `ui/`, and fails if any token isn't already in its canonical form (`data-stuck:` not
  `data-[stuck]:`, `scrollbar-none` not `[scrollbar-width:none]`, a bare-number or named form over a
  bracket). Run it with `--write` to auto-fix, or `--json` for tooling.
- **CSS variables in class names.** Never write `-[var(--x)]`. If `--x` is a `@theme` token, use its
  named utility instead — `--z-index-*` tokens generate `z-<name>` directly (`z-dropdown`,
  `z-modal`), no paren form needed. For a non-token var, use the paren shorthand:
  `w-(--loader-width,60px)`, `text-(length:--btn-font-size)`. `var()` stays inside a class name only
  inside a genuine multi-term `calc()`/`min()`/gradient, or an arbitrary property (`[prop:…]`) that
  has no utility form.
- **No `clsx` for a static class list, but keep it on a const.** `clsx` is registered under
  `sortTailwindcss.functions` in `.oxfmtrc.json`, and oxfmt sorts **each argument separately** — an
  argument boundary is a sort barrier, so `clsx('text-red-500 flex', 'p-2 items-center')` formats to
  `clsx('flex text-red-500', 'items-center p-2')`, sorted within each group and never across them.
  On a `className` (or any attribute listed in `sortTailwindcss.attributes`) write one plain string
  instead: `className="flex items-center p-2 text-red-500"` gets a single canonical sort. The
  inverse holds for a **module-level const**, where a bare string is not sorted at all — there
  `clsx` is what makes the formatter look at it, so `const ROW_CLASS = clsx(…)`
  (`ui/features/gear/components/SelectorModal/ItemList.tsx`) must keep its call. Reach for `clsx`
  when an argument is conditional or a variable, never to wrap or line-break a fixed list.

Breakpoints are named tokens in `ui/styles/theme/breakpoints.css`: `sm md lg xl xxl xxxl fhd qhd uhd`
(`576px` through `3841px`; `fhd` is the 1921px `1080p` cutover, a layout breakpoint only: the root
font-size does not change there). All breakpoints are px, so they never depend on
the root.

**Spec themes are CSS variables, not per-spec CSS.** `ui/styles/theme/specs.css`'s `.sim-ui` block
declares `--color-primary`/`--color-primary-foreground` from `--theme-color`/`--theme-color-foreground`, and
one block per spec (e.g. `.dps-warrior-sim-ui { --theme-color: var(--color-class-warrior); … }` and
`.protection-warrior-sim-ui { --theme-color: var(--color-class-warrior); … }`, each with its own
background-image variable — every spec has its own background even where two specs of a class
share `--theme-color`) set those variables. A component never needs a
spec-aware variant —
`bg-primary`/`text-primary`/`border-primary` already re-theme per spec because the sim root carries
the spec's class. Derived shades (`--color-primary-hover/-active/-dampened/-disabled`) are
`color-mix(in srgb, var(--color-primary) …)` in `@theme static inline` — `inline` because they read a
var that's re-declared lower in the tree (see `ui/styles/theme/colors.css`'s comment-free but
documented block; the `inline` flag matters here specifically because plain `@theme` would resolve
`var(--color-primary)` at `:root`, before `.sim-ui` re-declares it).

## 2. `ui-*` composition classes

When a component needs a reusable multi-utility recipe, or a rule keyed on a state/descendant
selector Tailwind's utility-per-property model can't express cleanly, it gets a `ui-<component>-<part>`
class in a co-located `.css` file, `@import`ed from `ui/styles/main.css` (see the long list of
`@import '../ui-kit/…/X.css'` lines there — one per component that has one). Example,
`ui/ui-kit/Chip/Chip.css`:

```css
@layer components {
	.ui-chip {
		@apply flex rounded-full border border-primary p-0 text-ui font-normal text-white …;

		&:hover:not([data-active]) {
			@apply bg-primary-dampened;
		}

		&[data-active] {
			@apply bg-primary;

			& .ui-chip-delete {
				@apply text-primary-foreground;
			}
		}
	}
}
```

Rules:

- **A `.css` file selects only `ui-*` classes** — attribute (`[data-active]`), pseudo (`:hover`,
  `::after`), and element/descendant parts are allowed _inside_ a `ui-*` rule, but the file may never
  target a semantic hook class, a `data-testid`, or a raw element selector at the top level. The two
  exceptions are `ui/styles/theme/specs.css` (the spec theme selectors) and `vendor.css` (see below).
- Built with `@apply`, inside `@layer components`, matching the layer order declared in
  `main.css:1`; every co-located `.css` file `main.css` imports is written this way.
- No comments — including lint-disable directives — in any `ui/` `.css` file, same rule as `.ts`/`.tsx`.
- Names are `ui-<component>-<part-or-variant>` for both kit and feature components, so they can never
  collide with a retired hook name; tests never locate elements by them (see §5).
- `@utility` is reserved for a handful of atomic, non-component patterns where every carrier is
  enumerable and nothing else on any carrier sets the same properties. Only eight exist today, in
  five groups, all in `ui/styles/main.css`: `text-fluid-xl`/`text-fluid-5xl` (RFS-style fluid headings), `interactive`
  (cursor + user-select), `icon-sm` (a fixed icon box), `focus-ring`/`focus-ring-inset` (the
  outline-based focus treatment), and `active-underline`/`fade-in-out` (an animated underline and a
  Base UI enter/exit fade). Don't add a sixth without the same proof: enumerate every element that
  will carry it, and show nothing else touches those properties there.
- Native CSS nesting (`&:hover`, `&[data-active]`, `& .child`, `.parent &`, nested `@media`) is fine —
  Tailwind's build goes through Lightning CSS, which flattens it for the v4 browser baseline. There's
  no Sass-style suffix concatenation (`&-item` isn't legal), no variables/mixins/`@each`/`@extend`.
- `vendor.css` (imported before the utilities layer, unlayered) is the deliberate home for the small
  number of selectors that have no JSX element to carry a class — third-party internals like
  react-tooltip's build-hashed content node (`[class*='styles-module_content']`) or its `--rt-opacity`
  variable.

## 3. State: `data-*` attributes

Every stateful class (`active`, `stuck`, `selected`, `show`, `loading`, `open`, `expanded`, `disabled`, …)
is a `data-*` attribute, never a class — boolean presence (`data-active`) or a value
(`data-state="open"`), plus Base UI's own state attributes on its primitives (`data-popup-open`,
`data-highlighted`, `data-checked`, `data-starting-style`, `data-ending-style`, `data-disabled`, …).
Style with the matching variant:

```tsx
<button data-active={active ? '' : undefined} className="hover:bg-primary-dampened data-active:bg-primary …" />
```

- `data-active:` (bare-key presence — write it this way, not `data-[active]:`; the canonical-classes
  checker flags the bracketed form), `data-[state=open]:` for a valued attribute, `aria-*:`,
  `group-data-[…]:` for ancestor state (a marker `group`/`group/<name>` class sits directly on the
  ancestor; `group` and `peer` are plain marker classes, not utilities, and have no effect inside
  `@apply`), `in-data-[…]:` for a descendant-driven read.
- A retired state class never comes back as a styling hook. `ui/retired_class_names.json`
  lists the retired names, and `ui/no_class_hooks.test.ts` fails on any of them through
  `findRetiredClassNames()` — see "The gates".
- Perf-critical per-frame writers (the timeline ruler) use `element.toggleAttribute('data-*', cond)`
  instead of `classList.toggle` — same cost, same no-op when unchanged, and it's the attribute form
  the styling above already expects.

## 4. Visibility driven by settings

A user setting that decides whether something exists (show-threat/healing/damage-metrics toggles,
experimental settings, EP ratios, and the like) is a **conditional render**, never a CSS hide class or
a `data-hide-*` + `in-data-[…]:hidden` pair. Read the value from a hook and branch in JSX:

```tsx
const displayMetrics = useDisplayMetrics(sim);
const showThreatMetrics = displayMetrics.threat;
…
{showThreatMetrics && <ThreatColumn />}
```

`ui/sim/hooks/useDisplayMetrics.ts` is the existing example (consumed by
`ui/features/stat-weights/components/EpWeightsDialog/EpWeightsDialog.tsx`, which also filters its
column list with `!column.metric || displayMetrics[column.metric]` rather than hiding cells by CSS).
For input configs, the equivalent is `showWhen` on the input plus the toggle's own store subscription
(see any spec's `inputs.ts` for the `showWhen: player => …` shape) — never a class the element itself
renders unconditionally and then hides.

## 5. Locating elements — tests and tools

Prefer role and accessible name first (`getByRole('button', { name: … })`); reach for `data-testid`
only when no role fits. Never `querySelector('.some-class')`, `closest('.x')`, `classList.contains`,
or an exact class-list snapshot (`expect(el.className).toBe(…)`) in production code or tests. This
is a review rule on this tree, not a checked one.

- **`data-testid` value convention:** when a test or tool has to find an element whose class hook
  is retired, the testid carries the retired class name verbatim (`className="apl-list-item-picker"`
  → `data-testid="apl-list-item-picker"`), so the swap is a search-and-replace and every
  `\.name(\.|$)` regex a tool runs against the class matches the testid instead. An
  _unread_ hook is just deleted — don't add a testid nobody consumes.
- **`data-testid` is never a styling hook.** No `.css` rule, `@apply` selector, or TSX arbitrary
  variant may key off `[data-testid=…]`; `class-hooks.mjs`'s `findTestidStyling()` is the check for
  it, run by `ui/no_class_hooks.test.ts`. If a style needs to reach a child, give that child a prop
  (`triggerClassName`), a `ui-*` class of its own, or a real `data-*` state — never repoint a rule at
  its testid.
- **Kit components take a `testId` prop** (`Dialog`, `Popover`, `Tooltip`, …), forwarded to
  `data-testid` on the element that actually carries the identity (often not the root — `Dialog`'s
  `testId` lands on the popup, with `sim-dialog-portal`/`-backdrop`/`-viewport` fixed on the
  surrounding wrapper elements).

## 6. Dialogs, menus, popovers

Every overlay primitive (`Dialog`, `Popover`, `ConfirmPopover`, `Menu`, `Toast`, `ProgressTrackerDialog`)
wraps a Base UI component. Dismissal — Escape, outside click, nested-popup layering — is Base UI's
job: never add a hand-written `onKeyDown` that swallows Escape, and never portal a popup out of the
floating tree in a way that breaks topmost-first closing. If Escape or click-outside misbehaves, the
fix is almost always removing whatever broke the native layering, not adding logic on top of it.

## 7. The gates

The tree-wide tests at the `ui/` root run in `npm run test:unit`, all under
`// @vitest-environment node` because they read the tree off disk and need no DOM:

- **`ui/canonical_classes.test.ts`** runs `ui/testing/tailwind/canonical-classes.mjs` over every
  `ui/` class string and `@apply`. It rejects a non-canonical Tailwind token and a `var(...)` inside
  a class token. Run the script under node with `--write` to auto-fix, or `--json` for tooling.
- **`ui/no_class_hooks.test.ts`** runs the three checks in `ui/testing/tailwind/class-hooks.mjs`:
    - `findClassHooks()` flags every class token in a `ui/**/*.{ts,tsx}` file that is not a real
      Tailwind utility, not `ui-*` and not in the allowlist. Real utilities come from compiling
      against `ui/styles/style.css` with Tailwind's own design system, not from a hand-maintained
      list.
    - `findRetiredClassNames()` flags any name from `ui/retired_class_names.json` anywhere in `ui/`,
      CSS included.
    - `findTestidStyling()` flags a `[data-testid]` selector in CSS or in a TSX arbitrary variant.
- **`ui/no_raw_controls.test.ts`** checks that `ui/features`, `ui/app` and `ui/specs` render every
  control through a kit primitive (`Button`, `WowheadAnchor`, the menu items), so a production
  `.tsx` file there has no bare `<a>` or `<button>`. `ui/ui-kit` and `ui/design-system` are outside
  its scope because those primitives and their docs live there.
- **`ui/no_testid_wrappers.test.ts`** flags a `<div>`, `<span>` or `<section>` anywhere in `ui/`
  whose only attributes are `data-testid` and, optionally, `key`. Such an element exists only to
  carry the test id. Find the element by role, label or text instead, or move the test id onto the
  element that carries the identity (§5).
- **The two JSX checks read production `.tsx` only.** `no_raw_controls` and `no_testid_wrappers`
  skip `*.test.tsx` and `*.stories.tsx`, and blank whole-line comments first, so a tag named in
  prose never counts.
- **For the class checks, production means every file but a test.** They skip `*.test.ts(x)` and
  `ui/generated/` only, so a story (`*.stories.tsx`), `ui/testing/**` (decorators, fixtures,
  `react.tsx`) and the design-system code (`ui/design-system/components/`, `tokens/`) are held to
  the same class rules as a component. The checkers read `.ts`, `.tsx` and `.css`, so the `.mdx`
  pages are not scanned.
- **The allowlist, `ui/class_hook_allowlist.json`,** is `sim-ui` and the `<spec>-sim-ui` pattern
  that `ui/styles/theme/specs.css` selects, `group`/`group/*`, `peer`/`peer/*`,
  `sim-tooltip`/`sim-tooltip--unpadded`, and the three third-party-required names
  `tooltip-quick-swap`, `suggest-reforges-softcaps` and `bonus-stats-popover`. react-tooltip
  forwards a component's `className` as its own `data-theme` attribute and has no `data-testid`
  equivalent, so those three cannot become pure testids. An entry needs a `token` or a `pattern`
  plus a `reason`, and only a genuinely external constraint earns one; it is never a shortcut past
  a real conversion.
- **Stories, tests and the design system stay out of production CSS.** Tailwind's automatic source
  detection is on, so a class string in any file under `ui/` lands in the production bundle unless
  `ui/styles/style.css` excludes it. Its `@source not` lines exclude `**/*.stories.tsx`,
  `**/*.test.tsx`, `**/*.test.ts`, `ui/testing/storybook`, `ui/design-system` and `**/*.md` (so a
  code example in a doc like this one never generates a utility). `.storybook/storybook.css` adds
  the stories, `ui/testing/storybook` and `ui/design-system` back for Storybook. A new class-bearing
  file under `ui/testing/` outside `storybook/` needs its own `@source not`. For a kit or styling
  change, build with `npx vite build` and diff the sorted `dist/forever/bundle/*.style.css` against
  the parent commit's build; an unintended selector there is a leak.
- **Rendering is verified by diffing computed styles, not by reading classes.** Walk the DOM by
  child-index path from `document.body` plus tag, capture the computed style properties and the
  bounding box per node, and diff that against a build from before the change, which validates a
  class-hook removal for free. Cover the default state of each top-level tab and the sub-panes
  (sub-tabs, dialogs, modals, expanded rows), because a sweep that only opens top-level tabs at rest
  misses whole areas of the app.
    - Deleting a real `@utility` (`icon-sm`) as if it were a retired class hook grows the log
      runner's rows from 32.5px to 34px, and a sweep that never opens the log tab stays green. A
      custom `@utility` or a `ui-*` composition class is real styling. The class-hook gate and the
      hook census tell it apart from a hook by compiling every candidate against Tailwind's own
      design system (`ui/testing/tailwind/canonical-classes.mjs`) instead of a hand-written name
      list, so do not delete one because it looks like an old semantic class. Run the diff before
      calling a class-hook or utility change done: the class-hook gate proves a class is gone, but
      not that nothing rendered differently without it.
    - **A px value converts to the nearest even spacing step.** The scale here is 1 or an even
      integer (`1, 2, 4, 6, 8, 10, 12, 16, …`), and one step is 3.5px at the 14px root that
      conversions use. A source px value snaps to its nearest step, with ties rounding down (`7px` →
      `p-2`, `14px` → `w-4`). A sub-step value standing for a hairline (a border, a divider, a ruler
      tick) converts to the fixed `px` step instead of rounding. `findUnevenSpacing` in
      `ui/testing/tailwind/canonical-classes.mjs` flags any spacing-scale utility or `--spacing(n)`
      whose step is not `1` and not an even integer. A rendering diff this rule introduces is
      expected, since the target step is not always the source's exact px value. Keep an arbitrary
      `[Npx]` only where no step and no hairline fits, and add a token instead of an arbitrary value
      for anything repeated.
    - A class-removal commit and a DOM-structure commit never land together, because a
      position-keyed comparison misaligns on every node after an inserted or deleted element.
