# Verification: what to run, and what each gate actually catches

**Source of truth:** the `scripts` block in `package.json` and `.github/workflows/run_tests.yml`.
Never quote a command from memory; print the current list:

```
node -e "console.log(require('./package.json').scripts)"
```

## Before you call a UI change done

```
npm run type-check       # node_modules/typescript/bin/tsc --noEmit — the whole repo, tools/ and .storybook/ included
npm run lint:js          # npx oxlint --max-warnings 0 ./ui ./.storybook
npm run lint:css         # stylelint "./ui/**/*.css"
npm run fmt              # npx oxfmt . --check — the whole repo, not just ui/
npm run test:unit        # vitest run — happy-dom, ui/**/*.test.ts(x), every story included
npm run test:locales     # ajv, assets/locales/** against schemas/**
npm run storybook:build  # storybook build --quiet → storybook-static/
```

`npm run lint:js` runs with `--max-warnings 0`, so a warning fails it exactly like an error. Note
that oxlint 1.77 prints **nothing at all** on a clean run: an empty output with exit 0 is a pass,
not a misconfigured invocation.

Add `npm run test:locales` whenever you touched `assets/locales/**` or `schemas/**`, and
`npm run lint:css` (stylelint over `ui/**/*.css`) whenever you touched a stylesheet. There is no
SCSS in the tree.

What each one is actually for:

| Gate              | Catches                                                                                                                                                  | Does not catch                                                                  |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `type-check`      | broken specifiers, alias mismatches, spec-config shape drift                                                                                             | a legal import that violates the layer direction                                |
| `lint:js`         | layer violations (`no-restricted-imports`), browser globals in `ui/sim` and feature models, hook-rule breaks, import order                               | anything not listed in `.oxlintrc.json` — `categories.correctness` is **off**   |
| `test:unit`       | component and helper behaviour, the store hooks' gating, a story that fails to mount, and the tree-wide gates below                                      | anything without a `.test.ts(x)` beside it, and warnings — see the last section |
| `test:locales`    | a locale key with no matching property in its `additionalProperties: false` schema                                                                       | a key the schema allows and no locale file defines                              |
| `fmt`             | formatting across the repo, markdown included — `assets/**` and `**/test-fixtures/*.json` are the ignored trees                                          | anything `.gitignore` already hides, which is how generated output escapes it   |
| `storybook:build` | what only Storybook's own build sees: a broken `.storybook/` config, an MDX page that does not compile, a story module the browser bundle cannot resolve | a story that builds and renders wrong — open it in `npm run storybook`          |

**Zero `no-restricted-imports` errors is the bar**, not a count held steady: the layer rules are the
point of the layout, and `lint:js` is the only thing enforcing them anywhere.

## What CI runs

`.github/workflows/run_tests.yml` has two jobs:

- **`build-ui`**: `npm ci` and `test:locales`, then `fmt`, `lint:js`, `lint:css`, a
  `make ui/generated/proto/api.ts go-to-ts` generate step, `type-check`, `test:unit`,
  `storybook:build`, and finally `make dist/forever/.dirstamp`. On a pull request from this
  repository two more steps upload `dist` to Cloudflare Pages and keep one comment on the pull
  request pointing at `https://pr-<number>.wowsims-forever.pages.dev/forever/`, the built UI with
  real sims. Fork pull requests carry no secrets and skip both. `preview_cleanup.yml` deletes those
  deployments when the pull request closes.
- **`test`**: `go test --tags=with_db ./sim/...` in one job (four shards when there is more than one runner) — the Go sim, not the UI.

So every script in the list above runs in CI, cheapest first, which is why a formatting slip reports
in about a minute instead of behind the wasm build. The generate step has to come before `type-check`
because `ui/generated/proto/**` and `ui/sim/wasm/bulk_sim/constants_auto_gen.ts` are gitignored and
absent from a fresh checkout; the other three `*_auto_gen.ts` are tracked.

`make dist/forever/.dirstamp` is not a thin wrapper: through `dist/forever/bundle/.dirstamp` it runs
`tsc --noEmit`, `npx tsx vite.build-workers.mts` and `npx vite build`, and it also builds
`dist/forever/lib.wasm.gz`, `ui/generated/proto/api.ts` and the asset copies — so the build type-checks
a second time, through make, never by calling `vite build` directly.

Three of `test:unit`'s files are tree-wide gates rather than component tests, and they are the only
thing enforcing what they check: `ui/canonical_classes.test.ts` (no non-canonical Tailwind tokens,
no `var(...)` inside a class token) and `ui/no_class_hooks.test.ts` (no class hooks, no retired
vanilla class names, no `[data-testid]` styling) run under `// @vitest-environment node` — they read
the tree off disk and need no DOM — and read the data files `ui/retired_class_names.json` and
`ui/class_hook_allowlist.json` (the rules are `ui/STYLING.md` §7; stories, `ui/testing/**` and the
design-system code count as production there, and `.mdx` pages are not scanned). `ui/testing/storybook/stories.test.tsx` globs
every `*.stories.tsx`, composes each story with `.storybook/preview.tsx` and asserts it mounts under
happy-dom; a story opts out with `parameters: { smoke: false, smokeReason: '…' }`, and an opt-out
with no `smokeReason` fails. List the opt-outs with
`/usr/bin/grep -rn "smoke: false" ui --include=*.stories.tsx`.

`build-ui` is reused through `workflow_call` by `deploy.yml` and `release.yml`, both with
`needs: tests`, so every gate above also blocks a deploy and a release.

Read the current job list rather than trusting this section:

```
RTK_DISABLED=1 /usr/bin/grep -n "name:\|run:" .github/workflows/run_tests.yml
```

## Warnings are invisible unless you ask for them

vitest intercepts console output, so React and Base UI warnings do not reach stdout on a normal run
— `npm run test:unit` can be entirely green while the suite emits hundreds of them. They surface in
CI logs and nowhere else, which makes them look like a CI-only phenomenon. They are not:

```
RTK_DISABLED=1 npx vitest run --disableConsoleIntercept 2>&1 | grep -c "not wrapped in act"
```

Three counts to keep at zero:

- `not wrapped in act` — a state update landing outside an act window. Usually a promise resolving
  after the test body, or a store notify called bare between acts.
- `overlapping act` — the failure mode introduced by fixing the first one carelessly.
- `Base UI:` — the `nativeButton` contract, among others.

**Prefix `RTK_DISABLED=1` on anything whose output you grep for a count**, not just git. RTK rewrites
dev commands too, and a compressed one-line summary greps as zero — indistinguishable from success.

If a flag the count depends on might be silently ignored, pass a deliberately bogus one first:
`--disableConsoleInterceptXYZ` fails with ``CACError: Unknown option `--disableConsoleInterceptXYZ` ``,
so a clean run proves the real flag was accepted rather than dropped.

**Do not attribute a warning from its position in the output.** Workers write to a shared stdout, so
a warning header from one file routinely lands next to a stack frame from another, which pins the
warning on the wrong file. `--no-file-parallelism` does not fix it, because
the reporter's file markers and the console writes are separate streams. Bisect instead: run a
directory, then a file, then `-t 'test name'`, and read the count.

```
for f in ui/features/settings/model/*.test.ts*; do
  printf '%-48s ' "$f"
  RTK_DISABLED=1 npx vitest run "$f" --disableConsoleIntercept 2>&1 | grep -c 'not wrapped in act'
done
```

## Locales

`npm run test:locales` runs `test-locales.mjs`, which compiles each `schemas/<name>.schema.json`
with ajv and validates every `assets/locales/**/<name>.json` against it. Today that is `character`,
`talents` and `translation` — only `assets/locales/en/` ships, no second locale. There is no
`glyphs` schema or locale file (the proto has no `Glyphs` message). `gear.schema.json` exists but
matches no locale file, so it validates nothing; `assets/locales/en/updates.json` has no matching
schema either, so it is never validated by this script.

**Every schema sets `additionalProperties: false`**, so a new locale key needs a matching property in
`schemas/<name>.schema.json`. Skip the schema and CI fails on the very first job. Verify the
constraint rather than trusting it:

```
node -e "const j=require('./schemas/translation.schema.json'); const s=new Set(); (function w(o){if(!o||typeof o!=='object')return; if('additionalProperties' in o) s.add(String(o.additionalProperties)); for(const k in o) w(o[k]);})(j); console.log([...s])"
```

## What no gate covers

Rendering, layout and interaction. **Nothing above proves the page rendered**, so a change that only
moves DOM or CSS around can pass every command here, and a story that mounts proves only that it
mounts. When a change touches rendering, layout or interaction, drive the built page yourself —
build it, serve it against the Go host (`running-locally.md`, "Running REAL sims"), and compare
screenshots against a build of the parent commit rather than against a remembered element count.
Say in the PR what you ran or what you clicked, with the screenshots.

For a kit or stylesheet change, also diff the production CSS: `npx vite build`, then a sorted diff of
`dist/forever/bundle/*.style.css` against the parent commit's. A selector that appears with no
source change behind it is a class leaking in from a story, a fixture or a design-system page (see
the `@source not` lines in `ui/STYLING.md` §7).

## A fresh checkout needs generated files first

`npm run type-check` and every build need files that are gitignored and produced by Go tooling:
`ui/generated/proto/*` (`make ui/generated/proto/api.ts`) and the `*_auto_gen.ts` files
(`make go-to-ts`, which runs `go run ./tools/database/gen_db -gen=go-to-ts`) — today three:
`ui/sim/player/classes/capabilities_auto_gen.ts`, `ui/sim/bulk/constants_auto_gen.ts` and
`ui/sim/wasm/bulk_sim/constants_auto_gen.ts`. Copying them from a built checkout works and is
faster. Never run `gen_db` concurrently with another copy of itself.

The wart to watch for is a `AUTO_GEN_FILES_TS` entry the generator does not actually write: a stale
path there makes the prerequisite never appear, so every `make` that depends on it re-runs `gen_db`
and re-bundles even when nothing changed. Today the three agree — deleting all
three and running `make go-to-ts` brings back byte-identical copies, and a second `make go-to-ts`
then says "Nothing to be done". Re-derive rather than trusting this paragraph:

```
/usr/bin/grep -n AUTO_GEN_FILES_TS makefile | head -1
/usr/bin/grep -rn 'constants_auto_gen.ts\"\|capabilities_auto_gen.ts\"' tools/database/
```

A fourth generated TS file, `ui/sim/constants/missing_effects_auto_gen.ts`, is _not_ in that list and
is not written by `-gen=go-to-ts`: it is tracked, and `tools/database/gen_effects.go` rewrites it
during full database generation, which needs `assets/db_inputs`.
