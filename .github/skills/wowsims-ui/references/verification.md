# Verification: what to run, and what each gate catches

**Source of truth:** the `scripts` block in `package.json` and `.github/workflows/run_tests.yml`. Print the current
commands instead of quoting them from memory:

```
node -e "console.log(require('./package.json').scripts)"
RTK_DISABLED=1 /usr/bin/grep -n "name:\|run:" .github/workflows/run_tests.yml
```

## Before you call a UI change done

Run these from the repo root after `npm ci`, and before every commit: the owner relies on the local gates, so a PR check
is not a substitute.

```
npm run type-check       # tsc --noEmit over the whole repo, tools/ and .storybook/ included
npm run lint:js          # oxlint --max-warnings 0 on ui/ and .storybook/
npm run lint:css         # stylelint over ui/**/*.css
npm run fmt              # oxfmt . --check, the whole repo including markdown
npm run test:unit        # vitest under happy-dom, ui/**/*.test.ts(x), every story included
npm run test:locales     # ajv, assets/locales/** against schemas/**
npm run storybook:build  # storybook build into storybook-static/
```

`lint:js` runs with `--max-warnings 0`, so a warning fails it like an error, and on a clean run oxlint prints nothing:
an empty output with exit 0 is a pass.

| Gate              | Catches                                                                                                             | Does not catch                                                            |
| ----------------- | ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `type-check`      | broken specifiers, alias mismatches, spec-config shape drift                                                        | a legal import that breaks the layer direction                            |
| `lint:js`         | layer violations, browser globals in `ui/sim` and feature models, hook-rule breaks, import order                    | any rule not listed in `.oxlintrc.json` (`categories.correctness` is off) |
| `test:unit`       | component and helper behaviour, the store hooks' gating, a story that fails to mount, and the tree-wide gates below | code without a test beside it, and warnings (see below)                   |
| `test:locales`    | a locale key with no property in its `additionalProperties: false` schema                                           | a key the schema allows that no locale file defines                       |
| `fmt`             | formatting across the repo; `.oxfmtrc.json` `ignorePatterns` and `.gitignore` are the exceptions                    | anything those two hide                                                   |
| `storybook:build` | a broken `.storybook/` config, an MDX page that does not compile, a story module the browser bundle cannot resolve  | a story that builds and renders wrong: open it in `npm run storybook:dev` |

Zero `no-restricted-imports` errors is the bar: `lint:js` is the only thing enforcing the layer rules.

## Reading the result

**Never pipe the gate.** `npm run test:unit | tail` exits with `tail`'s status, so a red suite looks green to the shell
and to any chain that greps the summary. Run the gate bare, or check the "Test Files" line for `failed` before you
commit.

**A single failure under load may be a flake.** The full suite has occasionally failed one interaction test under load
(for example `ui/features/bulk/components/BulkSettings/BulkSettings.test.tsx` or
`ui/features/gear/components/FiltersMenu/FiltersMenu.test.tsx`) that passes alone. Rerun that file alone, then the
whole suite once; call the change red only if it fails again.

## The tree-wide gates

The `ui/*.test.ts` files at the root of `ui/` are not component tests: they scan the tree from disk under
`// @vitest-environment node` and are the only enforcement of what they check. `ui/STYLING.md` §7 describes the class
gates (`ui/canonical_classes.test.ts`, `ui/no_class_hooks.test.ts`) and the raw-control gate
(`ui/no_raw_controls.test.ts`). `ui/no_testid_wrappers.test.ts` rejects a `div`, `span` or `section` whose only
attributes are `data-testid` and `key`: a test finds such an element by role, label or text, or the test id moves onto
the element that carries the identity. List them with `ls ui/*.test.ts`.

`ui/testing/storybook/stories.test.tsx` mounts every `*.stories.tsx` under happy-dom with `.storybook/preview.tsx`
applied. A story opts out with `parameters: { smoke: false, smokeReason: '…' }`, and an opt-out without a reason fails.
List the opt-outs with `/usr/bin/grep -rn "smoke: false" ui --include=*.stories.tsx`.

## What CI runs

`.github/workflows/run_tests.yml` has three jobs, `build-ui`, `test-ui` and `test`:

- **`build-ui`**: `make skills-check`, `npm ci` and `test:locales`, then `fmt`, `lint:js`, `lint:css`,
  `make ui/generated/proto/api.ts go-to-ts`, a `go build` of `wowsimcli` exported as `WOWSIMCLI` for the SIML imports
  (`running-locally.md`), `type-check`, the SIML extension's `npm run check` and `npm test` in `tools/vscode-apl`,
  `storybook:build`, and `make dist/forever/.dirstamp`, with `BUILD_COMMIT` set to the pull request's head so the page's
  release names it rather than the merge commit. A pull request from this
  repository then gets its build deployed as a Cloudflare Pages preview at a hashed `*.wowsims-forever.pages.dev`
  address, linked from one comment on the pull request; fork pull requests skip it. `preview_cleanup.yml` deletes the
  deployments when the pull request closes.
- **`test-ui`**: `test:unit`, sharded, each shard after `make ui/generated/proto/api.ts go-to-ts` and the same
  `wowsimcli` build for the SIML imports.
- **`test`**: the Go side, sharded; its steps are `wowsims-testing` `## The gates`.

The cheap UI gates run first, so a formatting slip reports in about a minute. `deploy.yml` and `release.yml` reuse the
workflow through `workflow_call` with `needs: tests`, so every gate also blocks a deploy and a release.

`make dist/forever/.dirstamp` runs `tsc --noEmit`, `npx tsx vite.build-workers.mts` and `npx vite build` through
`dist/forever/bundle/.dirstamp`, and also builds `dist/forever/lib.wasm.gz`, `ui/generated/proto/api.ts` and the asset
copies. Build the site through make rather than by calling `vite build` alone.

## Warnings are invisible unless you ask for them

vitest intercepts console output, so React and Base UI warnings do not reach stdout: `npm run test:unit` can be green
while the suite emits hundreds of them. Count them:

```
RTK_DISABLED=1 npx vitest run --disableConsoleIntercept 2>&1 | /usr/bin/grep -c "not wrapped in act"
```

Keep three counts at zero: `not wrapped in act` (a state update outside an act window, usually a promise resolving after
the test body or a bare store notify between acts), `overlapping act` (what a careless fix for the first produces), and
`Base UI:` (the `nativeButton` contract, among others).

Prefix `RTK_DISABLED=1` on anything whose output you count: `rtk` compresses dev commands too, and a one-line summary
greps as zero. If a flag might be silently dropped, pass a bogus variant first (`--disableConsoleInterceptXYZ` fails
with `Unknown option`), so a clean run proves the real flag was accepted.

Do not attribute a warning from its position in the output. Workers share stdout, so a warning from one file often
lands next to a stack frame from another, and `--no-file-parallelism` does not fix it. Bisect instead: a directory,
then a file, then `-t 'test name'`, reading the count each time.

```
for f in ui/features/settings/model/*.test.ts*; do
  printf '%-48s ' "$f"
  RTK_DISABLED=1 npx vitest run "$f" --disableConsoleIntercept 2>&1 | /usr/bin/grep -c 'not wrapped in act'
done
```

## Locales

`npm run test:locales` runs `test-locales.mjs`, which compiles each `schemas/<name>.schema.json` and validates every
`assets/locales/**/<name>.json` against it. Only `assets/locales/en/` ships. A locale file with no schema of the same
name is never validated, and a schema with no locale file validates nothing; compare `ls schemas` with
`ls assets/locales/en`.

Every schema sets `additionalProperties: false`, so a new locale key needs a matching property in
`schemas/<name>.schema.json`, or CI fails in its first step.

## What no gate covers

Rendering, layout and interaction. A change that only moves DOM or CSS can pass every command above, and a story that
mounts proves only that it mounts. For such a change, build the page, serve it against the Go host and take screenshots
(`running-locally.md`), comparing against a build of the parent commit rather than a remembered element count. Say in
the PR what you ran or clicked, with the screenshots.

For a kit or stylesheet change, also diff the production CSS against the parent commit's, as `ui/STYLING.md` §7
describes: a new selector with no source change behind it is a class leaking in from a story, fixture or docs page.

## A fresh checkout needs generated files first

`type-check` and every build need files that Go tooling generates and a fresh checkout lacks: `ui/generated/proto/*`
(`make ui/generated/proto/api.ts`) and the untracked `*_auto_gen.ts` files among the ones `make go-to-ts` writes
(`go run ./tools/database/gen_db -gen=go-to-ts`). Copying them from a built checkout also works. Every build,
`vitest` run and Storybook build also needs Go itself, for the `.siml` imports (`running-locally.md`). The makefile's
`AUTO_GEN_FILES_TS` must list exactly what that generator writes; a stale entry makes every dependent `make` rerun
`gen_db`. Compare the generator's outputs with the makefile and with what git tracks:

```
/usr/bin/grep -n '^AUTO_GEN_FILES_TS ' makefile
/usr/bin/grep -rn 'auto_gen.ts"' tools/database/*.go
/usr/bin/git ls-files 'ui/**/*_auto_gen.ts'
```

The tracked ones that `go-to-ts` does not write come from the full database generation, which needs the game files
database that CI does not have, so they are committed. Never hand-edit any of them, and never run two `gen_db`
processes at once.
