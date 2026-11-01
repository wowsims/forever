---
name: wowsims-core-mechanics
description: "Use when WoWSims Forever Go code depends on what sim/core does around a spell that its row cannot show: the order of applyEffects, OnCastComplete and proc handlers (TriggerImmediately, SpellBatchWindow), costs, miss refunds, combo points and charges, how damage scales, how many targets an area hits, what CastRequirement refuses, registration hooks, whose auras a tag reads, the ActionID tag of a weapon hand's copy, aura lifetimes, exclusive categories, hit tables, resists, dot tick outcomes, pets, rage and threat, target death and the target lifecycle, encounter designs, and proto field rules. Use it too when asked whether the sim models something (a racial, crit damage, resists, rage, stat scaling), when a handler sees or misses the cast that procced it, a refund goes wrong, another caster's dot counts, a registration panics, or a new proto bool reads false in saved settings. Reading a row is wowsims-spell-store; building an ability or aura is wowsims-abilities-auras; proc triggers are wowsims-procs; the pick at a cast end is wowsims-apl; runtime proofs are wowsims-testing; the map is wowsims-spells."
---

# wowsims-core-mechanics

A row says what a spell is; `sim/core` decides when its effects run, what the cast spends, which targets it reaches and
whose state it reads, so a spell that matches its row can still be wrong here. The rule this skill exists for: a claim
about core's timing (an order, a spend, a refund) is proven at runtime, never by reading the call order alone. The
rulings the code follows are in `references/rulings.md`; text here follows `wowsims-conventions`.

## Where things live

| Path                            | What it is                                                                   |
| ------------------------------- | ---------------------------------------------------------------------------- |
| `sim/core/cast.go`              | the cast: its checks, cost, GCD and cast time, then the effects              |
| `sim/core/spell.go`             | `RegisterSpell`, `CanCast`, `CanCompleteCast`, `applyEffects`, crit damage   |
| `sim/core/aura.go`              | aura registration and lifetime, tags, `NewAllyAuraArray`                     |
| `sim/core/aura_helpers.go`      | `core.ProcTrigger` and its deferral, `MakePermanent`                         |
| `sim/core/dot.go`               | `Dot.Apply`, ticks and refreshes                                             |
| `sim/core/spell_outcome.go`     | the outcome functions: hit, crit, block and the tick outcomes                |
| `sim/core/spell_result.go`      | `SpellChanceToMiss`, crit suppression, `DealCastResult`                      |
| `sim/core/spell_resistances.go` | partial resists and binary spells                                            |
| `sim/core/target.go`            | `NewAttackTable`: level difference, spell miss; `sim/core/attack.go` weapons |
| `sim/core/exclusive_effect.go`  | exclusive categories and effects                                             |
| `sim/core/energy.go`            | energy and combo points; `sim/core/rage.go` and `sim/core/mana.go` likewise  |
| `sim/core/pet.go`               | pets: enable and disable, stat inheritance, start distance                   |

| You are…                                                                     | Open                                                                        |
| ---------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| asking what the rotation sees at a hardcast end                              | `docs/rotation.md`, then `wowsims-apl`                                      |
| deciding a hit, roll, regen, OOM or pet rule the row leaves open             | `references/rulings.md`                                                     |
| chasing a trap: an item swap, weapon reach, `GainHealth`, pets, a proto bool | `references/traps.md`                                                       |
| looking for the review case behind a section below                           | `.github/skills/wowsims-review/references/defects.md`, the matching section |

## Cast order

- **`applyEffects` runs before `OnCastComplete`** (`cast.go`): `OnApplyEffects`, then the spell's `ApplyEffects`.
- **A proc handler runs inside `applyEffects` only with `TriggerImmediately`.** Without it, which is how
  `spelldata.ProcTrigger` builds one, it is a pending action one `SpellBatchWindow` (10 ms) later (`aura_helpers.go`),
  after `OnCastComplete`; a stepped test needs `sim.Step()` to see it fire.
- **`CallbackOnCastComplete` carries no result**, so `Outcome` and `RequireDamageDealt` are never read and no manager
  rolls (`aura_helpers.go`). A proc off the cast's hit roll uses `CallbackOnCastResult` (`spell_result.go`). Prove an order at runtime: `wowsims-testing` `## Proving a claim at runtime`.

## What a cast spends

- **Core spends the cost at the cast; a refund is the spell's own call.** `spell.IssueRefund(sim)` gives back the cost's
  `Refund` share and a mana cost refunds nothing (`mana.go`). `SpellConfig` sets the share from the row's `MissRefund()`
  (0.8 where the row flags Discount Power On Miss), the default unless verified spell-specific behaviour differs.
- **Combo points are not part of the cost.** The spell spends them through `SpendComboPoints` (`energy.go`), each
  resource on its own: a failed Ferocious Bite pays its base cost and keeps its points and its extra energy.
- **A one-charge buff is spent by the cast it modifies**; the spender is `wowsims-procs`. A spell's own charges
  (`MaxCharges`) are spent by `ConsumeCharge` and refuse the cast at zero.

## What scales

- **A flat weapon bonus is added before the weapon percent multiplies it** (`WeaponStrike` in `sim/core/spelldata/`).
- **`PointsPerResource` is the amount per combo point** (a parsed stat takes it at full points), and a tick that should
  scale with stacks or combo points reads them.
- **Crit damage** is `(base × CritMultiplierPct × CritDamageMultiplier × table − 1) × (CritMultiplierAdditive + 1) + 1`
  (`spell.go`).

## How many targets

- **An area effect reaches every enemy in it** (`Effect.HitsAnArea()`, `HitsACone()`, `HitsAroundTheCaster()` in
  `sim/core/spelldata/resolve_spell.go`). `MaxTargets` has no `SpellConfig` field: a caller that caps reads
  `s.MaxTargets`, and "up to $s3 targets" or `SPELLMOD_JUMP_TARGETS` is the total count.

## What the row forbids

- **Core checks `CastRequirement` on every cast**: form or stance, weapon or shield, "not usable in combat"
  (`castRequirementFailure` in `shapeshift.go`), in `Cast` and in `CanCompleteCast` (so in `CanCast`), before
  `ExtraCastCondition`, which never restates it (`wowsims-abilities-auras` §An ability; ranks: `auras-and-dots.md`).

## Hooks

- **A hook is attached once, at registration, on every path that needs it** (every attack that breaks stealth, a
  queued swing for every spec that registers it), and never in a reset, which runs every iteration. A rule every class
  needs goes in core once. `RegisterSpell` panics past 1000 spells, the sign of a per-iteration registration, and on a
  `DamageMultiplier` or `ThreatMultiplier` other than 0 with `ProcMaskUnknown` (`spell.go`).

## Whose state

- **An effect on several targets reads each hit target's state**, never the primary target's, and a map read for a key
  it may lack does not stand for 0.
- **Tags are every caster's.** `GetAurasWithTag` on a unit returns every registered aura with the tag, any caster's,
  active or not, and `HasAuraWithTag` is true for a registered aura that is not up (`aura.go`); filter by caster and use
  `GetActiveAuraWithTag` or `IsActive()`. A per-caster debuff's label and slots are `wowsims-abilities-auras` §An aura
  or dot.
- **A per-hand copy's `ActionID.Tag` is its hand's `ItemSlot`.** `ActionID.WithHand` (`sim/core/agent.go`) tags the
  main-hand copy 14 and the off-hand copy 15, the auto attack's two swings included, and the UI names those two tags as
  the hands (`ui/sim/proto/action_id/index.ts`). An extra attack is the main-hand swing tagged with the granting spell's
  id. Tags 1 and 2 are not hands: each spell or item gives them its own meaning, such as a queue spell or Weaponmaster's
  weapon branches.

## Auras and exclusive effects

- **`MakePermanent` sets `NeverExpires` and re-activates the aura in its `OnReset`** (`aura_helpers.go`); an aura with
  duration 0 panics on activation (`aura.go`).
- **A dot tick rolls its amount, not its hit** (`docs/spell_data/auras-and-dots.md`); a spell whose ticks roll hit
  passes `Spell.OutcomeTickMagicHitAndCrit` or `Dot.OutcomeTickMagicHit`. Always Hit on a triggered row:
  `.github/skills/wowsims-abilities-auras/references/rulings.md`.
- **Within an exclusive category the enabled effect with the highest `Priority` applies**; `YieldsTies` takes the
  category only by outbidding (`exclusive_effect.go`). The `SingleAura` reset is `wowsims-abilities-auras` §Buffs and
  debuffs.

## Targets and encounter designs

- **Zero active targets is a legal state** (`Target.Disable`, `moveRaidOffTarget` in `target.go`). Units on a disabled
  target move to target 0 when it is enabled, else the enabled target with the lowest index; with none left they keep
  the disabled target, their auto attacks stop and `Target.Enable` resumes them (`returnStrandedRaid`). Code that reads
  the current target may meet a disabled one. `Target.Disable` interrupts the target's own hardcast, so a boss that
  goes down mid-cast does not land the spell once it is back, and ends every raid unit's channel on the target with no
  further tick, for a downtime, cleared or not, and a death alike; a ground channel lives on its caster and runs on. A
  dot keeps ticking on a target in a downtime that does not clear its debuffs.
- **A death runs in one order** (`Target.Die`, `target_death.go`): the kill share is rolled once on the design rand,
  `OnTargetDies` runs for every raid unit, `OnKill` after it when the roll credits the raid, `Disable` last. A death
  listener gets a nil spell; an aura on the dying target is checked in `ExtraCondition`, which runs inside `Die`. The
  proc side is `docs/spell_data/procs.md`, which also says why a rated trigger hears no death.
- **Execute phases are per target** (`target_health.go`): an Execute-type check reads `result.Target`, never the
  encounter's first target. A plain encounter's targets cross 35% and 20% at the encounter's execute proportions of the
  fight (`resetExecutePhases`); a design's health curve sets the crossings of its boss and adds. A buff row a target's
  phase grants sits on the caster and carries to the next target (Decimation's 440873); a bonus whose rows state no buff
  counts per target (Quietus, Early Demise).
- **An encounter design is a timeline of fractions of the rolled fight duration** (`design_runner.go` and the
  `design_*.go` lanes). Jitter is rolled per occurrence at reset from `sim.designRand`, a stream apart from the
  sim's, so it does not move with gear. `DesignRepeat.count` is repeats after the first, 0 until the fight ends;
  overlapping downtimes on a target are a union; boss health keyframes get (0, 100) and (1, 0) added and are clamped;
  add waves take targets from a pool allocated at construction. Design input never panics: what construction drops
  is logged at each reset (`constructNotes`).
- **A pool target is a target of the results, after the encounter's own.** Its label stays `Target N`, which the log
  and the UI's log parser key on; its metrics carry the wave and add it spawns as (`UnitMetrics.design_add`, copied
  by the concurrent combiner too), and `EncounterMetrics.makeNew` in `ui/sim/proto/sim_result.ts` names it from
  them. Every slot is listed, whether or not it spawned in an iteration.
- **Fights are time-based**: there is no health-based fight length, and a target's health is a curve over the fight,
  not damage taken.

## Protos

- **On a merge, master keeps its field numbers and the branch renumbers its own new fields.** Nothing converts saved
  data (`ProtoVersion` in `proto/common.proto`), so renumbering a branch field costs nothing, while renumbering master's
  breaks every branch built on it.
- **No proto3 `optional` field**: the `npx protoc` behind `make ui/generated/proto/api.ts` rejects it. "Unset means X"
  is a plain field whose zero value means none or the default, or a message-typed field.
- **No `reserved` line before launch**, and a breaking change needs no migration; the full ruling is in
  `references/rulings.md`.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. The commands that
re-derive each claim, the "You changed → Update" table and the path check are in `references/maintenance.md`.

A ruling, trap or workflow fact learned while using this skill is written down in the same change that uses it, where
`AGENTS.md` says: a comment at the code when one file owns the rule, this skill or its `references/rulings.md` otherwise. A memory file is a pointer
to this skill, not a second copy.
