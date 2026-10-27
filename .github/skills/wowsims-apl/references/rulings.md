# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

- **At a player's hardcast end the rotation picks at the cast-end timestamp, before the cast's outcome applies, and a
  pushback re-queues the pick; the major cooldowns that go ahead of it and the units that look after the landing are
  `docs/rotation.md` §The flow.** `sim/core/gcd.go`, PR #640.
- **During the hardcast the pick reads the cast's own dot, own-id aura, cooldown and mana as landed; the full reading
  and the lines that need `Spell Is Casting` are `docs/rotation.md`.** `sim/core/apl_values_aura.go`, PR #640.
- **Include Reaction Time delays every aura change the same way, the unit's own casts included; during a hardcast the
  cast's own aura reads as landed instead.** `sim/core/apl_values_aura.go`, PR #638, PR #640.
- **`Is Hit Capped` and `Is Expertise Capped` are per spell, current target only and yes or no: melee caps at the
  special-attack soft cap without the dual-wield penalty, the table follows the spell's `DefenseType` (proc mask for
  none), auto shots roll the ranged table, and the 1e-9 tolerance stays in the APL values.**
  `sim/core/apl_values_spell.go`, PR #608.
