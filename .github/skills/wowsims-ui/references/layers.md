# Layers: where code goes and what may import what

**Source of truth:** `.oxlintrc.json` (the bans) and `tsconfig.json` `paths` (the aliases). Both are
machine-enforced on every `npm run lint:js` / `npm run type-check`, so they cannot drift from the
build the way prose can. Dump the current bans rather than trusting this file:

```
node -e "const c=require('./.oxlintrc.json'); c.overrides.forEach(o => { const r = o.rules?.['no-restricted-imports'] ?? o.rules?.['no-restricted-globals']; if (r) console.log(JSON.stringify(o.files), JSON.stringify(r)); })"
```

`ui/README.md` has the orientation prose — the tree, placement rules, how to author a spec.
Read it for shape, but prefer `.oxlintrc.json` for rules: the README is enforced by nothing, so it
can drift from the build silently.

## The bans, as configured

| Scope                       | May not import                                                                                       | Notes                                                                                                                                                                                                                 |
| --------------------------- | ---------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ui/sim/**`                 | `@ui-kit` `@features` `@app` `@specs`, `@testing`                                                    | **`allowTypeImports: true`** — see below                                                                                                                                                                              |
| `ui/i18n/**/*.{ts,tsx}`     | `@app` `@features` `@ui-kit` `@specs/**`, `@testing`                                                 | `@sim` is allowed; no type-import escape hatch                                                                                                                                                                        |
| `ui/ui-kit/**`              | `@features` `@app` `@specs`, `react-dom/client`, `@sim` except `utils/**` `constants/**`, `@testing` | only the shell may create a React root; the `@sim` group is below                                                                                                                                                     |
| `ui/features/**`            | `@app` `@specs`, plus `react-dom/client`, `@testing`                                                 |                                                                                                                                                                                                                       |
| `ui/features/*/model/**`    | `@app` `@specs`, plus `react` and `react-dom`, `@testing`                                            | a model layer is framework-free, not merely DOM-free                                                                                                                                                                  |
| `ui/features/**/*.{ts,tsx}` | `patchSlice` `patchKeyed` `seedKeyed` `deleteKeyed` from `**/state/sim_store*`                       | write through a facade instead                                                                                                                                                                                        |
| `ui/features/**/*.test.*`   | `@app` `@specs`, plus `react-dom/client`                                                             | a test may seed the store with the writers; a model test may use React                                                                                                                                                |
| `ui/app/**`                 | `@specs`, `@testing`                                                                                 | specs import app, never the reverse                                                                                                                                                                                   |
| tests and stories           | the bans of their layer, without `@testing`                                                          | `ui/{sim,i18n}/**/*.test.*`, `ui/{ui-kit,app}/**/*.{test,stories}.*` and `ui/features/**/*.stories.tsx` restate their layer's groups; the two features test blocks also drop the store writers, as the row above says |
| `ui/design-system/**`       | `@features` `@app` `@specs`                                                                          | the docs pages sit beside the kit: `@sim` and `@ui-kit` allowed                                                                                                                                                       |
| `ui/testing/**`             | `@app` `@specs`                                                                                      | the story fixtures and decorators may reach features                                                                                                                                                                  |

The kit's `@sim` group is a positive list, because a negated glob inside a group is unproven here:
`hooks` `context` `state` `player` `raid` `proto` `settings` `presets` `talents` `bulk` `cache`
`wasm` `workers`, and the top-level `sim` `sim_host` `sim_runs` `sim_signal_manager`
`spec_config` modules. A new top-level entry under `ui/sim` has to join it, and a directory with an
`index.ts` needs both forms, `@sim/<dir>` and `@sim/<dir>/**`, since the bare specifier resolves to
it (`wasm` and the six directories listed first have both). It covers the kit's tests
and stories too, with no type-import escape hatch; a store binding belongs in `ui/features/inputs`.

Global bans (`no-restricted-globals`, error): `window` `document` `localStorage` `location`
`navigator` in `ui/sim/**` and in `ui/features/*/model/**`. Anything that needs one takes an
injected `Env` (`ui/sim/state/env.ts`; `sim.env` at runtime, `ui/app/browser_env.ts` in the browser,
an in-memory `makeEnv()` in `ui/sim/sim.test.ts`).

**Naming a DOM type is not the same as being a view.** A function that takes an `HTMLElement` it is
_handed_ and enriches it — with an icon, an href, a Wowhead dataset — can still belong in the sim
layer: it never creates an element and it touches no browser global, so it satisfies the rule. Do
not "tidy" a sim-layer function that names `HTMLElement` into `features/` on sight — the test for
this layer is browser globals and element _creation_, not whether `HTMLElement` appears in a
signature.

## Three ways these rules surprise people

**`import type` helps in `ui/sim/**` and nowhere else.** That block, and only that block, sets
`allowTypeImports: true`, so `ui/sim` may name an `@app`/`@features`/`@ui-kit` type as long as the
import is erased. No other block sets it, and `no-restricted-imports` bans the _specifier_ rather
than the binding, so `import type { SimHostObject } from '@app/individual_sim_ui'` inside
`ui/ui-kit` is still a violation even though nothing survives to runtime. Confirm which blocks have
it before you rely on it:

```
node -e "console.log(require('./.oxlintrc.json').overrides.filter(o => o.rules?.['no-restricted-imports']).map(o => [o.files, JSON.stringify(o.rules['no-restricted-imports']).includes('allowTypeImports')]))"
```

**Patterns must use `**`, not `*`.** oxlint matches one path segment per `*`, so `@features/*`
silently misses `@features/gear/components/SelectorModal/ItemList` and the ban looks like it is
working when it is not.

**Scoping is an explicit file list.** Negated globs in `overrides.files` break scoping in oxlint, so
a new top-level `ui/` directory needs its own override block — you cannot exclude it with `!`. And
`categories` is `{"correctness": "off"}`: a rule that is not listed explicitly does not run at all.
That is why `react-hooks/rules-of-hooks` appears in the root `rules` block by name; deleting the
line silently disables it.

**A later block replaces a rule, it does not add to it.** When two blocks match a file and both set
`no-restricted-imports`, the later block's options win whole; the groups are not merged, and
`typescript/no-restricted-imports` is the same rule under another name. `excludeFiles` exists, but a
block that excludes the tests leaves them with none of its bans. So the `@testing` ban sits in each
production block, and each test and story block that follows repeats its layer's groups without it.
A new group in a layer block goes into that layer's test and story blocks as well.

## Aliases

| Alias          | Resolves to      |
| -------------- | ---------------- |
| `@sim/*`       | `ui/sim/*`       |
| `@generated/*` | `ui/generated/*` |
| `@worker/*`    | `ui/worker/*`    |
| `@ui-kit/*`    | `ui/ui-kit/*`    |
| `@features/*`  | `ui/features/*`  |
| `@app/*`       | `ui/app/*`       |
| `@specs/*`     | `ui/specs/*`     |
| `@i18n/*`      | `ui/i18n/*`      |
| `@testing/*`   | `ui/testing/*`   |

Check with `node -e "console.log(require('./tsconfig.json').compilerOptions.paths)"`. The same table
is `UI_ALIASES` in `vite.config.mts`, which `getBaseConfig` (inherited by the worker builds),
`vitest.config.mts` and `.storybook/vite.config.ts` spread into `resolve.alias`. The two must agree; a mismatch shows up as a type-check that passes and a
build that cannot resolve.

The aliases are `tsconfig` `paths` rather than Node's `package.json` `imports` field because `tsc`
under `moduleResolution: "bundler"` does not resolve `#foo/*` subpaths.

## Crossing layers without importing downward

A lower layer that needs to talk to a shell uses a **narrow host interface**, never the class. They
all live in `@sim/sim_host` (`ui/sim/sim_host.ts`) — today `SimWarning`, `SimHost`,
`IndividualSimHost<Spec>`, and the `isIndividualSimHost()` predicate that replaces
`instanceof SimHostObject` (a runtime check an interface cannot express).
`SimHostObject implements IndividualSimHost` (`ui/app/individual_sim_ui.tsx`) — the `implements`
clause is what keeps the interfaces honest, so add to the interface and let the compiler find the
host. From React the host arrives through context rather than a prop: `useSimHost()`, `usePlayer()`,
`useSim()` in `ui/sim/context/SimHostContext.tsx`, which is why `ui/sim` owns React context despite
being the bottom layer.

The per-spec config schema is `@sim/spec_config` (`ui/sim/spec_config.ts`): `IndividualSimUIConfig`,
`InputConfig`, `InputSection`, `CustomSection`, `OtherDefaults`, `Settings`, `SpecDefinition`,
`SpecBehaviors`, `DerivedSetting`, `defineSpec`, `registerSpecConfig`, `itemSwapEnabledSpecs`. Its
`InputConfig` is a union of the typed picker configs in `@features/inputs/input_helpers`, and its
icon inputs come from `@features/inputs/icon_inputs`; it names those, the reforge options and the
kit's `SavedDataConfig` through `import type` only, which the `ui/sim/**` block's
`allowTypeImports` permits and a value import would break. List what `ui/sim` names upward with
`/usr/bin/grep -rn "@ui-kit\|@features\|@app" ui/sim --include=*.ts --include=*.tsx | /usr/bin/grep -v "\.test\."`.

## The module-evaluation-order trap

`ui/sim/player/specs/<class>.ts` calls `getSpecSitePath()` (from `@sim/constants/other`) at **module
scope**, to build each `PlayerSpec`'s `simLink`. If one of those files ever imports a _value_ from a
non-leaf module, the resulting cycle lets a bundler evaluate `ui/sim/player/specs/index.ts` before the
class modules it aggregates, `specToPlayerSpec` gets built from `undefined`s, and every
`PlayerSpecs.fromProto` silently returns `undefined` — which surfaces as a crash inside
`new Player`, far from the cause.

Rule: `player/specs/*` and `player/classes/*` import leaf modules only.

`import/no-cycle` is deliberately **not** enabled. It reports **58** cycles — 43 in `ui/sim`, 14 in
`ui/features`, 1 in `ui/i18n` — a triage project, not a switch to flip. Re-derive the number before arguing either way:

```
RTK_DISABLED=1 npx oxlint --import-plugin -D import/no-cycle ./ui 2>&1 | /usr/bin/grep -c 'import(no-cycle)'
```

Until someone does that work, cycles are caught by review, not by lint.

## Moving a file

There is no move tool. Move by hand with `git mv`, then repair specifiers:
emit alias form when the importer and the target end up in different top-level `ui/` directories and
a relative specifier otherwise, and run `npx oxlint ./ui --fix` to re-sort the imports the rewrite
disturbs. Then the gates in `verification.md` — a move is exactly the change that passes
`type-check` and only fails at runtime.
