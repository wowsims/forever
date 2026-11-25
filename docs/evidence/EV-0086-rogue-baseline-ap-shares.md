# EV-0086: Rogue attack power shares for Eviscerate, Garrote and the poisons

- Claim: Eviscerate takes 0.04 attack power a point per combo point, Garrote 0.03, each Deadly Poison tick 0.009 for
  each stack and Instant Poison 0.01.
- Sources: in-game tests; the Instant Poison share was confirmed by a tester; nothing else recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #140, whose body reads "Includes all talents and spells and default APL"
- Later count: Deadly Poison, §Count of the 2026-11-24 log

The Eviscerate and Garrote shares are marked as estimated from in-game tests. The Instant Poison share is marked as
confirmed by a tester. No counts were recorded.

## Deadly Poison

- Count: §Count of the 2026-11-24 log

## Count of the 2026-11-24 log

- Sources: combat log `WoWCombatLog-112426_191754`, one level 60 rogue, three Scholomance runs
- Game build: 1.60.1.71091; the hotfix push and whether the player relogged after it not recorded
- Date: 2026-11-24 19:17 to 23:25 UTC−5 (2026-11-25 00:17 to 04:25 UTC)
- Pull request: not recorded

The talent sets are EV-0192's. The third run's combat set takes Venom and Improved Poisons 4/5, and the others neither.
Each Deadly Poison IV tick (18 a stack in its row) is read by its base amount, before crit, with the stacks from its
aura events, outside Venom's buff and leaving out ticks partly resisted, absorbed or killing. The attack power is the
character sheet's: the logged attack power less 14 times the main-hand weapon's damage per second, 620 for this sword.
The per-stack amount is the slope of the ticks over 1 to 5 stacks.

| Part                 | Ticks | Sheet attack power, mean | Per stack | 18 + 0.009 × attack power | Read / formula |
| -------------------- | ----- | ------------------------ | --------- | ------------------------- | -------------- |
| First run, combat    | 471   | 912                      | 26.09     | 26.21                     | 0.995          |
| Second run, subtlety | 305   | 844                      | 25.38     | 25.59                     | 0.992          |
| Third run, combat    | 58    | 857                      | 25.23     | 25.71                     | 0.981          |
| Third run, subtlety  | 121   | 832                      | 25.74     | 25.49                     | 1.010          |

The two large parts differ by 68 attack power and by 0.71 a stack, where a 0.009 share a tick gives 0.61 and the same
share spread over the 4 ticks gives 0.15, so the share is per tick and no multiplier sits on top of it. At every stack
count from 1 to 5 the ticks read about 2 under the stacks times the formula (24.4 against 26.2 at one stack and 128.6
against 131.0 at five in the first run); that offset is not explained.
