# Rulings

Each bullet is one rule the code follows: its id, the rule, where it lives, the pull request that settled it, and the
evidence that measured it (`docs/evidence/`). An id is never reused; a replaced rule keeps its bullet with a
`Replaced by` note.

- **R-0031** **An APL aura condition reads an aura only once it is active: a ProcTrigger declares no pending aura, and a proc
  that must land in the same batch as its hit sets `TriggerImmediately`, only with log proof.**
  `sim/core/aura_helpers.go`, PR #638.
- **R-0032** **The combat model is level 60 with target levels 58 to 63: attack tables come from the level difference, capped at
  +3, and per-hand weapon skill.** `sim/core/target.go`, PR #70.
- **R-0033** **`EffectVariance` is the full roll width: spell hits and heals roll the whole amount, power included, by ±V/2
  (`Spell.DirectVariance` after power), and physical effects roll the base by ±V/2, rounded to a whole number,
  before attack power.** `sim/core/spell_result.go`, PR #267; `sim/core/spelldata/effect.go` `RollAt`.
- **R-0034** **Every DoT tick rolls its effect's amount.** `sim/core/spelldata/resolve_aura.go`, `PeriodicDamageTick`, PR #573.
- **R-0035** **Tick misses count in `TargetedActionMetrics.tick_misses`, not `Misses`, and Consecration's own cast roll does not
  count (`OutcomeTickMagicHitNoHitCounter`).** `sim/core/spell_outcome.go`, PR #570.
- **R-0036** **Mana regenerates in 100 ms ticks; a paid channel holds regen until max(start + 5 s, end) and a free channel does
  not.** `sim/core/mana.go`, PR #249.
- **R-0037** **OOM runs while the highest-priority spell the APL wants (condition true, ready but for mana) is skipped for mana,
  whatever casts instead; it starts at the first free rotation pass (GCD ready, no hardcast) that skips one and ends at
  the first that skips none.** `sim/core/spell.go`, `meetsCost`, PR #371.
- **R-0038** **Dynamic pets (demons, hawks) refresh their inherited stats from the owner's totals at the pet heartbeat instead of
  running stat deltas through the inheritance.** `sim/core/pet.go`, PR #432.
- **R-0039** **The sim is not live, so a change that breaks saved APLs or settings needs no migration, only a note in the PR body;
  a removed proto field leaves no `reserved` line, new fields refill freed numbers and every `// NextIndex:` is the
  highest used number + 1.** `proto/common.proto`, PR #149.
