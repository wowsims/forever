# EV-0101: A weapon percent scales a strike's flat damage in the in-game tooltip

- Claim: the weapon percent of a weapon strike scales its flat weapon-damage roll too, so Holy Strike rank 1's 12 ± 1.5 at 25% shows as 2 to 4.
- Sources: in-game tooltips; no log
- Game build: not recorded
- Date: not recorded
- Pull request: #161

Holy Strike's tooltip reads "25% weapon damage plus an additional 2 to 4 as Holy damage" for rank 1 at level 20. Rank 3 at level 20 confirms it: 19 ± 2.375 rounds to 17 to 21, and at 32% the tooltip reads 5 to 7. Tooltips from another source round differently and do not match these in-game values; the pull request follows the game.

| Reading                      | Flat roll                    | Weapon percent | In-game tooltip |
| ---------------------------- | ---------------------------- | -------------- | --------------- |
| Holy Strike rank 1, level 20 | 12 ± 1.5                     | 25%            | 2 to 4          |
| Holy Strike rank 3, level 20 | 19 ± 2.375, rounded 17 to 21 | 32%            | 5 to 7          |
