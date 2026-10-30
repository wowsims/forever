# Keeping wowsims-ui true

| You changed…                                                                                                                 | Update                                                 |
| ---------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| `.oxlintrc.json`, `tsconfig.json` `paths`, or a top-level `ui/` directory                                                    | `references/layers.md` and the layer map in `SKILL.md` |
| `ui/sim/state/sim_store.ts`, `subscriptions.ts`, `batch.ts`, `events.ts`, `ui/sim/hooks/*`, `ui/sim/context/`                | `references/state.md`                                  |
| `ui/features/inputs/**`, the binding layer between the store and the kit's pickers                                           | `references/state.md`                                  |
| `ui/sim/state/persistence.ts`, `serialization.ts`, `sim_links.ts`, `stat_filter_sharing.ts`                                  | `references/persistence.md`                            |
| a `package.json` script or `.github/workflows/run_tests.yml`                                                                 | `references/verification.md`                           |
| `vite.config.mts`, `vite.build-workers.mts`, `tools/vite/spec_pages.mts`, `tools/vite/siml.mts`, the makefile's dist targets | `references/running-locally.md`                        |
| `.storybook/**`                                                                                                              | `references/running-locally.md`                        |
| `.oxfmtrc.json`, how a spec is authored, or a decision recorded as settled                                                   | `references/conventions.md`                            |
| `ui/sim/talents/*.ts`, `ui/features/talents/**`, the broken-talent warning in `default_warnings.ts`                          | `references/talents.md`                                |

Then run the shared path check, which fails on any path this skill names that no longer exists and on any reference
that `SKILL.md` does not route to:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-ui
```

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
