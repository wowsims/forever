---
name: wowsims-dependencies
description: 'Upgrade the toolchain and the libraries of WoWSims Forever: Node, Go, npm packages and Go modules. Use it for a version bump of any size, a framework or major upgrade, a Go or Node release, a package held back, a setup problem after a version change, and before touching package.json, package-lock.json, go.mod, .nvmrc or a version in a workflow.'
---

# Upgrading dependencies

A green unit run does not prove an upgrade: a package can change what ships without a type error. Upgrade one
group at a time and compare the build against `master`, not against what the tests happen to cover.

## Where a version lives

A toolchain bump changes all of these in one commit:

- **Node**: `.nvmrc`, `engines` and `volta` in `package.json`, `node-version` in `.github/workflows/`, `NODE_VERSION`
  in `Dockerfile`, `docs/installation.md`, `README.md`, `AGENTS.md`.
- **Go**: `go` and `toolchain` in `go.mod`, `go-version` in `.github/workflows/`, the `golang` images in `Dockerfile`,
  `docs/installation.md`, `README.md`, `AGENTS.md`.

Take the newest Node LTS from `https://nodejs.org/dist/index.json`, the entries with a codename in `lts`, and the
newest Go from `https://go.dev/dl/?mode=json`.

`tools/vscode-apl` and `tools/vscode-spelldata` run on VS Code's own Node and carry a committed build, so they move
with the VS Code version and not with this list.

## Order

1. **See what is behind.** `npm outdated`, and `go list -m -u all` for the modules.
2. **Go toolchain and modules.** `make sim/core/proto/api.pb.go binary_dist/dist.go`, then `go get -u ./...` and
   `go mod tidy`. A new Go runs a stricter `go vet` inside `go test`, so a test package can stop building. Every class
   suite has to match its `.results` unchanged: a changed result is a behaviour change to explain, not a file to update.
3. **Node.** Node 25 and later have a global `localStorage` that hides happy-dom's, and Vitest 5 keeps the test
   environment's: under Vitest 4 tests fail with `Cannot read properties of undefined (reading 'getItem')`.
4. **Patch and minor npm packages**, then each major on its own. Bump the packages in `dependencies` apart from those in
   `devDependencies`, so a failure names its group.
5. **Generated code.** Rerun `make ui/generated/proto/api.ts go-to-ts`. A diff under `ui/generated/` is a change to the
   shipped code.

## Proof

The gates in `wowsims-testing` run first, with `make skills-check`. When one fails, run it on a checkout of `master`
too: a Windows checkout fails some Go and SIML tests there for reasons of its own (CRLF, a space in the path), and only
a failure `master` does not have belongs to the upgrade.

An upgrade that touches a package the page ships also needs a comparison of two builds, `master` against the branch.
Build `storybook-static` and `dist` in a checkout of each, then:

```
node .github/skills/wowsims-dependencies/scripts/compare_builds.mjs --base <master checkout> --head <branch checkout>
```

It shoots every story and every tab of every spec page from both builds, retakes a story that differs to tell an
unstable one from a changed one, and sims each spec twice on `master` and once on the branch. A sim seeds itself at
random, so the branch's result is read against the gap between the two `master` runs. Every changed shot needs an
explanation.

Once the pull request has a preview, run the same script against it:

```
node .github/skills/wowsims-dependencies/scripts/compare_builds.mjs --smoke https://<preview>/forever/
```

Every spec page has to open, switch every tab and sim, its export link has to open on a new page, and its rotation has
to take a new action.

The script needs Playwright, which the repository does not depend on: `npm i --prefix <folder outside the repository>
playwright`, `npx playwright install chromium`, and `PLAYWRIGHT_MODULE=<that folder>/node_modules/playwright/index.mjs`.

## Held back

| Package                   | Held at | What a newer version does                                                                                  |
| ------------------------- | ------- | ---------------------------------------------------------------------------------------------------------- |
| `pako`, `@types/pako`     | 2.x     | 3.x has no default export, so `sim_links.ts` and `link.ts` throw when a link is made.                      |
| `@protobuf-ts/*`          | 2.9.x   | 2.11 generates a `create` that reads `this`, so a `create` passed as a callback (`useListItems`) throws.   |
| `@tanstack/react-virtual` | 3.14.10 | 3.14.24 renders no rows in `VirtualList` under happy-dom, and the real page's behaviour is not yet proved. |
| `highs`                   | 1.15.2  | A new version needs a new `ui/worker/highs.wasm`, so it goes through `make update-highs` on its own.       |

A package leaves this table in a change of its own, with its code adapted and the proof above repeated.
