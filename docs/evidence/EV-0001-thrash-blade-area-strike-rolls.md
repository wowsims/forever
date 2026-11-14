# EV-0001: Thrash Blade rolls once for each target an area strike lands on

- Claim: a chance-on-hit weapon rolls once for each target a Whirlwind or Cleave lands on.
- Sources: the end-game dungeon events of the 2026-11-12 weapon rate sweep (EV-0003)
- Game build: not recorded
- Date: 2026-11-12
- Pull request: #986

Each Whirlwind and Cleave cast is grouped by the number of targets it landed on, and a cast counts once when Thrash
Blade procced from it. One roll a cast would stay near 4% however many targets it lands on; one roll a target
predicts the last column.

| Ability, landed targets | Casts | Casts with a proc | Logs  | One roll a target |
| ----------------------- | ----- | ----------------- | ----- | ----------------- |
| Whirlwind, 1            | 1175  | 48                | 4.1%  | 4.0%              |
| Whirlwind, 2            | 1057  | 74                | 7.0%  | 7.8%              |
| Whirlwind, 3            | 819   | 90                | 11.0% | 11.5%             |
| Whirlwind, 4            | 826   | 114               | 13.8% | 15.1%             |
| Cleave, 1               | 940   | 37                | 3.9%  | 4.0%              |
| Cleave, 2               | 4661  | 318               | 6.8%  | 7.8%              |

Weaponmaster's sword roll is not covered here and stays once a cast.
