# EV-0071: A level 30 Cat's swing before attack power

- Claim: a tamed pet's swing is the creature dps less its own attack power over 14, and its logged hits are that swing plus attack power over 14, over the swing as sped up, with the Cat's +10% and happiness.
- Sources: one combat log of a level 30 Cat, Scarlet Monastery Graveyard
- Game build: not recorded
- Date: 2026-10-20
- Pull request: #789

The pet sheet shows the whole creature dps as the weapon and leaves happiness out, so its damage range is not the damage the pet deals. The creature dps already holds the pet's own attack power (30% of it), so the weapon is the creature dps less that attack power over 14, which is 70% of it.

The log holds 152 swings that are neither crits nor glancing blows, at 1.2 s with Unleashed Fury 5/5. Damage before armor at the attack power the log gives the Cat on each swing, against the two models:

| Pet attack power | Logged | Whole creature dps | Own attack power inside the creature dps |
| ---------------- | ------ | ------------------ | ---------------------------------------- |
| 119              | 42-55  | 53.6-72.3          | 42.4-55.6                                |
| 146              | 46-59  | 57.3-76.0          | 46.1-59.2                                |

Both models take the Cat's 10% and happiness's 25%. With the whole creature dps as the weapon, 131 of the 152 swings fall outside the range and the model runs about 32% high. With the own attack power taken out every swing lands inside the range. With those multipliers held fixed, the best fit for the weapon is 70.5% of the creature dps, with a plausible band of 70.0 to 71.0%. The runner-up model drops happiness and keeps the whole creature dps, but it needs an extra 0.976 multiplier and fits worse.

A hawk checked the same way: its logged attack power is exactly 80% of the owner's ranged attack power with none of its own, and all 142 of its swings that are neither crits nor glancing blows land inside the range.

Open points. It is one log, one pet, at level 30. The log has no talent data: Unleashed Fury 5/5 is the player's word and Focused Fire is assumed off. With Focused Fire 2/2 the fit comes out 2% high. The Cat's logged attack power (119 and 146) is lower than 30% of the creature dps plus 10% of the owner's ranged attack power gives (143 and 179), and it moves by 7.6% of the owner's ranged attack power between the two, where the pet sheets read 10%. In a log of 2026-10-09 it stays at 139 while the owner's ranged attack power goes from 876 to 1022.
