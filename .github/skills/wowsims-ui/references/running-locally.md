# Running the sim locally

**Source of truth:** `vite.config.mts`, `vite.build-workers.mts`, `tools/vite/spec_pages.mts`, `tools/vite/siml.mts`,
`tools/vite/build_info.mts`, the `devserver` / `rundevserver` / `devmode` / `host` targets in the makefile,
`ui/worker/local_worker.ts`, and for Storybook `.storybook/main.ts` and the `storybook:*` scripts in `package.json`. Read
a target before copying a command from here:

```
/usr/bin/grep -n "^devserver:\|^rundevserver:\|^devmode:\|^host:" -A8 makefile
```

## The normal path

`npm start` is `cross-env WATCH=1 make devmode`: it builds the Go dev server, runs `vite serve --host` and the worker
build beside it, and reloads the Go binary through `air`. Use it unless you need one of the setups below.

## `.siml` imports need Go

`tools/vite/siml.mts` reads every `.siml` import (the presets under `ui/specs`) with the Go reader: the first one
starts `wowsimcli siml serve -names sim/siml/names.json`, one per vite or vitest process, which answers each file
over stdin and stdout until the process ends. It runs the binary `$WOWSIMCLI` names, else `bin/wowsimcli` when that
file exists, else `go run ./cmd/wowsimcli` from the repository root, so the dev server, `vitest`, Storybook and
`vite build` all need Go on the PATH. `make` builds `bin/wowsimcli` and rebuilds it when a Go file under `sim`,
`cmd/wowsimcli` or `tools/spelldata` changes, and skips it when `$WOWSIMCLI` is set (`make bin/wowsimcli` on its own); the plugin does not check its age, so after a Go change
outside `make`, rebuild it or delete it. It stays out of `dist`, which the pull request preview uploads whole to
Cloudflare Pages, and Pages refuses a file over 25 MiB. `go run` costs about 0.3 s once the build cache holds the package, and a few
seconds before; CI builds the binary once and exports `WOWSIMCLI`:

```
go build -o /tmp/wowsimcli ./cmd/wowsimcli && export WOWSIMCLI=/tmp/wowsimcli
```

No `with_db` tag is needed: the server reads the name table from `sim/siml/names.json`, again when it changes. A
change to the Go reader itself takes effect on the next vite start, once the binary it runs is rebuilt.

## Two things a plain vite dev server cannot do

**Run sims on its own.** Under `vite serve`, `/forever/sim_worker.js` is mapped to `ui/worker/local_worker.ts`, which
sends every sim request to a Go host at `http://localhost:3333`. Without one, sims fail.

**Render spec pages without the worker bundles.** `ui/sim/workers/worker_pool.ts` loads the worker from
`/forever/sim_worker.js`, which the dev server serves out of `dist/forever/`. With no bundle there, `waitForInit()`
never resolves, `loadSettings` never runs and the page looks structurally fine but has no settings, with no console
error. Build the workers first:

```
npx tsx vite.build-workers.mts     # dist/forever/{local,net,sim}_worker.js and highs.wasm
```

A fresh worktree also needs the generated files (`verification.md`, last section).

`tools/vite/spec_pages.mts` serves every spec URL from `ui/index_template.html`, and redirects the bare
`/forever/<class>/<spec>` to the trailing-slash form as the production host does. Only an unknown spec URL falls through
to the landing page, so getting the landing page means there is no `ui/specs/<class>/<spec>/spec.ts(x)` for that URL.

## The release a page shows

`tools/vite/build_info.mts` reads the release when vite loads its config: the commit (`BUILD_COMMIT`, else
`git rev-parse HEAD`, else `GITHUB_SHA`), the UTC build date, the game build stamped on each part of
`assets/db_inputs/spell_store_inputs.json`, and each spec page's last presets and engine changes, over the paths listed
above `specReleases`. A shallow clone has no changes to read, so the build workflows check out the full history.

`vite.config.mts` defines the lot as `import.meta.env.VITE_BUILD_INFO`. The landing page shows the game build, the build
date and the commit beside its title; a spec page shows the game build and the date under the sidebar's stats, and its
info button opens the sim build with the commit and the spec's presets and engine changes. The makefile's bundle target depends on the checkout's `HEAD` and
`index`, so a commit or checkout rebuilds it. vitest and Storybook define no `VITE_BUILD_INFO`; the stories and tests
fill `BUILD_INFO` from `ui/testing/build_info.ts`.

## Running real sims

The Go host serves `dist/forever` from disk and answers the worker's sim requests natively, so stats, Simulate,
reference swap and timeline tooltips all work. The short form:

```
make dist/forever/.dirstamp       # wasm, workers, the vite bundle, protos, the SIML name table and the asset copy
make devserver                    # go build -o wowsimforever ./sim/web
tail -f /dev/null | ./wowsimforever --usefs=true --launch=false --host=":<port>"
```

Then open `http://localhost:<port>/forever/<class>/<spec>/`. `make rundevserver` runs the server on `:3333` for you.
Start the server with stdin held open (`tail -f /dev/null |`, run in the background): with stdin closed it answers once
and then waits on its command prompt. Pick a free port, since other sessions may run their own server, and stop yours by
its PID. `pkill -f` or `pkill -x wowsimforever` would also kill other sessions' servers and their stdin holders.

Without `dist/forever/assets` the page half-renders with JSON parse errors from the database fetch: that is a missing
asset copy (`make` does it; a plain `vite build` does not), not a broken database. Do not `rm -rf dist`: `vite build`
restores only the bundles, and the database and assets come from other tooling. Copy `dist` to a scratch directory and
serve that on its own port if you need an isolated build, for example the parent commit's for a comparison.

## Screenshots

A change to rendering, layout or interaction gets screenshots of the real sim, because no gate shows the page. Drive it
with headless `playwright-core`. It is not a dependency of this repo, so use a local install whose browser build is in
`~/.cache/ms-playwright` (a `chromium_headless_shell-*` matching that install's version). What works:

- Block `**/*wowhead.com/**` with `page.route` so tooltips and icons do not hang the page or redirect it.
- Load with `waitUntil: 'load'` and then wait about 8 seconds; the page never reaches `networkidle`.
- Icon pickers open on hover, and their trigger is a Wowhead link, so a click navigates away. Hover, never click.
- Screenshot the element (`locator.screenshot()`) rather than the page, and find it by its `data-testid`
  (`ui/STYLING.md` §5).
- Headless Chromium may draw overlay scrollbars, so a layout that depends on scrollbar width can differ from a desktop
  browser.
- To load a realistic build, edit the page's `…__currentSettings__` localStorage key, or paste an
  `IndividualSimSettings` JSON into Import → JSON, and reload.

Compare against a build of the parent commit on a second port rather than a remembered element count.

There is no perf harness in this tree. For APL-edit timings, drive the page the same way and record the protocol in the
PR.

## Storybook

```
npm run storybook:dev      # http://localhost:6006
npm run storybook:build    # storybook-static/ (gitignored)
```

Both need the generated TS files like every other build (the story fixtures import `@generated/proto/*`), but not the
Go host, the worker bundles or `dist/`: stories render against stub hosts.

`.storybook/main.ts` collects `ui/**/*.mdx` and `ui/**/*.stories.tsx` and points the vite builder at
`.storybook/vite.config.ts`. The builder would load `vite.config.mts` by default, which carries the site build's plugins;
the Storybook config takes only `UI_ALIASES` from it, plus React, Tailwind, the i18next loader and the SIML plugin (`simlPlugin`, which
`vitest.config.mts` registers too). `staticDirs` serves
`assets/` under `/forever/assets`, where fonts and spec backgrounds are root-absolute. `.storybook/preview.tsx` imports
`.storybook/storybook.css` (the Storybook stylesheet entry, see `ui/STYLING.md`), the i18n config and
`ui/shared/page_boot.ts`, stubs `gtag`, and adds the spec toolbar and the `withSimUi` frame;
`.storybook/preview-head.html` loads the Wowhead tooltip script.
