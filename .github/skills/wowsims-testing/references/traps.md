# Traps outside the gates

Numbers that move or differ for a known reason, and CI behaviour a pull request runs into.

## Rows that move on their own

- `TestFeralBear`'s StatWeights-Default row (`sim/druid/feralbear/feralbear_test.go`) divides by an Agility weight
  within noise of 0, so any timing change moves it.
- `TestSpiritRegenIsTheSameForEveryClass` (`sim/core/mana_regen_test.go`) holds a naked Dwarf paladin's Spirit, so it
  moves with the Dwarf race offsets.
- **A duration aura that a plain fight puts on a target moves the class `.results`**, even when nothing reads it: its
  expiry adds a step to the fight and reorders the events at that time. Per-target state that only a target's death
  reads is kept as an end time per target instead, as Shadowburn's Soul Shard window is (`sim/warlock/shadowburn.go`).

## Comparing runs

- **Seeds overlap.** Iteration i runs on `RandomSeed + i` (`sim/core/sim.go`), so runs on nearby seeds share most
  iterations; space independent seeds by at least the iteration count.
- **The CLI drops unknown JSON fields** (`DiscardUnknown` in `cmd/wowsimcli/cmd/basic_sim.go`), so a buff under the
  wrong group (`partyBuffs` for a `RaidBuffs` field) vanishes with no error.
- **UI and CLI tank numbers differ.** A tank fixture may lack `inFrontOfTarget` (no dodge, parry or block) or
  `inspirationUptime`, and the UI's fallback target in `ui/sim/raid/encounter.ts` carries BlockValue 54 where the
  database Raid Target (`sim/encounters/register_all.go`) has none; the sim does not enforce `LimitCategory`.

## Pull request previews

- Every same-repository pull request gets a Cloudflare Pages preview from `build-ui`. `pull_request` runs the
  workflow as the merge commit holds it, so a stacked pull request whose head and base predate a workflow change runs
  the old one until master is merged into its head; a run started before a secret existed sees it empty.
- A second pull request that carries another's commits and lands as a merge commit closes the original as merged, and
  `preview_cleanup.yml` deletes its preview; a squash or rebase merge keeps it.
