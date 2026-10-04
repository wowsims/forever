# Running the sim locally

**Source of truth:** `vite.config.mts`, `vite.build-workers.mts`, `tools/vite/spec_pages.mts`, the
`devserver` / `rundevserver` / `devmode` / `host` targets in the makefile, and for Storybook
`.storybook/main.ts` and the `storybook:dev` / `storybook:build` scripts in `package.json`. Read the
target before copying a command out of here:

```
/usr/bin/grep -n "^devserver:\|^rundevserver:\|^devmode:\|^host:" -A6 makefile
```

## The normal path

`npm start` is `cross-env WATCH=1 make devmode`: it builds the Go dev server, starts
`vite serve --host` and `vite.build-workers.mts` alongside it, and reloads the Go binary through
`air`. Use it unless you need something the sections below describe.

## Two things a plain `vite` dev server cannot do

**It cannot run sims.** The worker's backend fetch fails on localhost. Sims need the Go host.

**It half-renders spec pages with ZERO console errors unless the worker bundles exist first.**
`ui/sim/workers/worker_pool.ts` resolves the worker at `/forever/sim_worker.js`, and the dev server has
middleware that serves `/forever/**` out of `dist/forever/`. With no bundle there the probe simply hangs:
`waitForInit()` never resolves → `loadSettings` never runs → defaults are never applied → the page
looks structurally fine and is empty of settings. Build them first:

```
npx tsx vite.build-workers.mts     # → dist/forever/{local,net,sim}_worker.js (+ highs.wasm)
```

A fresh worktree also needs the gitignored generated files before anything type-checks or builds —
see the last section of `verification.md`.

Copying per-spec `ui/**/index.html` from a built checkout is **not** a step:
`tools/vite/spec_pages.mts` serves every spec URL itself from `ui/index_template.html`
through `transformIndexHtml`, and 301s the bare `/forever/<class>/<spec>` to the trailing-slash form the
way the production static host does. Only an _unknown_ spec URL falls through to vite's SPA
fallback, which serves the landing page — so "I got the landing page" means the URL is not in
`discoverSpecPages()`, i.e. there is no `ui/specs/<class>/<spec>/spec.ts(x)`.

## Running REAL sims

The Go host serves `dist/forever` from disk and answers the worker's sim requests natively — stats,
Simulate, reference swap and timeline tooltips all work.

```
make binary_dist/dist.go          # stub package so ./sim/web compiles — no built checkout needed
make wasm                         # dist/forever/lib.wasm.gz (needs Go)
npx tsx vite.build-workers.mts    # dist/forever/*_worker.js
npx vite build                    # dist/forever bundle + the 18 per-spec index.html
cp -r assets dist/forever/assets      # `make` does this; a plain `vite build` does not
make devserver                    # go build -o wowsimforever ./sim/web
./wowsimforever --usefs=true --launch=false --host=":3333"
```

Then `http://localhost:3333/forever/<class>/<spec>/`. `make rundevserver` runs the last two steps for
you, and `make dist/forever/.dirstamp` replaces the middle three (it also copies the assets).

Without `dist/forever/assets` the page half-renders with JSON parse errors from the DB fetch — that
error text is the signature of a missing asset copy, not of a broken database.

Do not `rm -rf dist`. `vite build` restores the bundles only; the database and assets come from
other tooling, and you will have broken the worktree. Copy `dist` to a scratch directory and serve
that on its own port if you need an isolated measurement.

## What to look at once it is up

A spec page that came up correctly has its settings pickers populated and its sidebar stats filled
in. An empty settings tab is the worker-probe hang above, not a rendering bug. Get the expected
shape by serving a known-good build of the parent branch on a second port and comparing the same
URL, rather than trusting a remembered element count — those numbers rot faster than anything else
in a document like this.

There is no browser or perf harness in this tree. If you need APL-edit timings, drive the page with
Playwright yourself and record the protocol in the PR.

## Storybook

```
npm run storybook:dev      # storybook dev -p 6006 --no-open → http://localhost:6006
npm run storybook:build    # storybook build --quiet → storybook-static/ (gitignored)
```

Both need the gitignored generated TS files like every other build (the story fixtures import
`@generated/proto/*`; see the last section of `verification.md`), but not the Go host, the worker
bundles or `dist/`: stories render components against stub hosts, not a running sim.
`node -e "console.log(require('./package.json').scripts)"` prints the exact commands.

`.storybook/main.ts` collects `ui/**/*.mdx` and `ui/**/*.stories.tsx`, and points the vite builder
at `.storybook/vite.config.ts` rather than the repo's `vite.config.mts`. The builder would load
`vite.config.mts` by default, and that config carries the site build's plugins (`specPages`,
`watchAndRun`, `checker`); the Storybook config takes only `UI_ALIASES` from it, plus React,
Tailwind and the i18next loader. Fonts and spec backgrounds are root-absolute under
`/forever/assets`, so `staticDirs` serves `assets/` there. `.storybook/preview.tsx` imports
`.storybook/storybook.css` (the Storybook stylesheet entry, see `ui/STYLING.md`), the i18n config
and `ui/shared/page_boot.ts`, stubs `gtag`, and adds the spec toolbar and the `withSimUi` frame;
`.storybook/preview-head.html` loads the Wowhead tooltip script the spec page loads.
