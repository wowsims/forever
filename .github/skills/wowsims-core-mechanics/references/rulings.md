# Rulings

Each ruling is its own file in `docs/rulings/` (`docs/rulings/README.md`): the rule, the code that follows it, the pull
request that settled it and its evidence. This index lists the ones this skill owns.

- **R-0031** An APL aura condition reads an aura only once it is active: `docs/rulings/R-0031-apl-aura-condition-active-only.md`
- **R-0032** The combat model is level 60 against targets of level 58 to 63: `docs/rulings/R-0032-combat-model-level-60-targets.md`
- **R-0033** EffectVariance is the full roll width: `docs/rulings/R-0033-effect-variance-full-roll-width.md`
- **R-0034** Every DoT tick rolls its effect's amount: `docs/rulings/R-0034-dot-tick-rolls-effect-amount.md`
- **R-0035** Tick misses count in tick_misses and not in Misses: `docs/rulings/R-0035-tick-misses-own-counter.md`
- **R-0036** Mana regenerates in 100 ms ticks and a paid channel holds regen: `docs/rulings/R-0036-mana-regen-ticks-channel-hold.md`
- **R-0037** OOM runs while the highest-priority wanted spell is skipped for mana: `docs/rulings/R-0037-oom-skipped-spell-for-mana.md`
- **R-0038** Dynamic pets refresh their inherited stats from the owner at the pet heartbeat: `docs/rulings/R-0038-dynamic-pets-refresh-inherited-stats.md`
- **R-0039** The sim is not live, so breaking saved settings needs no migration: `docs/rulings/R-0039-sim-not-live-no-migration.md`
