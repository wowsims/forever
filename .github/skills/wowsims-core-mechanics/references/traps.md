# Traps in core

Each line is a trap of `sim/core` that a spell's row cannot warn about, under the section of `SKILL.md` it belongs to.

## What a cast spends

- **An item swap that moves a weapon slot extends the GCD to at least now + 1.5 s and restarts both swing timers**
  (`sim/core/item_swaps.go`, `SwapItems`). The cost is not measured in game, and an APL with swaps needs
  `enableItemSwap: true`.

## How many targets

- **A `core.Weapon{}` literal leaves `MaxRange` 0.** A melee swing's range check reads its spell's `MaxRange`
  (`MaxMeleeRange`, `IsInRange()` in `sim/core/attack.go`), but a walk wakes the unit where it crosses the weapon's own
  `MaxRange` (`sim/core/movement.go`), so a pet or form weapon that walks into melee range sets
  `MaxRange: core.MaxMeleeRange` (PR #578).

## Hooks

- **An aura's `OnExpire` also runs at iteration teardown** (`doneIteration` in `sim/core/aura.go`), so arrival logic
  hangs on `MovementEnd`, not on an expiry.
- **Shared code hooks a pet through `Pet.RegisterEnableEffect` and `RegisterDisableEffect`**, since `OnPetEnable` is
  one class-owned field (`sim/core/pet.go`).
- **`GainHealth` fires no heal callback** (`sim/core/health.go`), so a heal that must roll procs (Vampiric Embrace) is
  a real heal spell (PR #246).
- **A PseudoStats change made in a post-finalize effect is gone after the first `Reset`.** `unit.finalize` saves
  `initialPseudoStats` before `env.finalize` runs the post-finalize effects, and `reset` restores that copy
  (`sim/core/unit.go`). A fixed multiplier on a pet goes in before finalize, as Command's does in `applyRaceEffects`
  (`sim/core/racials.go`).
- **A bare stepped sim with no APL never swings** until `AutoAttacks.EnableMeleeSwing(sim)` (`sim/core/attack.go`).
- **A spec's own metric row goes through `spell.RecordCustomMetric(target, key, value)`** (label
  `results_tab.details.custom_metrics.<key>`), and a shield records through `spell.RecordShielding`; `Shield.Apply` has
  no callers (`sim/core/spell.go`, PR #510).

## Auras and exclusive effects

- **`Pet.Disable` expires only active auras**, which never-expiring auras never join, so a dismiss leaves permanent pet
  auras up (`sim/core/pet.go`).

## Protos

- **A new proto3 bool decodes as false from saved settings**, so a default-on toggle shows off for returning users
  (PR #243).
