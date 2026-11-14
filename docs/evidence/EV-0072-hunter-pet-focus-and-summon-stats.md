# EV-0072: Pet focus regenerates 10 a second and the summon keeps attack power and health

- Claim: a pet's focus comes back at 10 a second, 1 every 0.1 s, and never shows above 100; its attack power and health are taken at the summon and kept, while its armor follows the owner.
- Sources: combat logs of 2026-10-20 (a level 30 Cat), 2026-11-07 and 2026-11-08 (hunter pets at levels 5 to 48); report codes not recorded
- Game build: not recorded
- Date: 2026-10-20, 2026-11-07 and 2026-11-08
- Pull request: #955

## Focus

A level 30 Cat that spends all its focus on 25 focus Claw casts one every 2.43 s in the log of 2026-10-20. The November pets read 11 to 13 a second with Bestial Discipline. The bar stays at 100: no logged pet at levels 5 to 48 shows more, although Hunter Pet Scaling (415429) carries a +50 max focus effect.

## Summon

In the logs of 2026-11-07 and 2026-11-08 the pet's attack power holds through 20 Aspect of the Hawk toggles and a gear swap in one summon. Its health holds through the owner's Stamina changes, and its armor follows each gear swap.

A sim run locally at the summon gave the pet 51.6 attack power, 10% of the owner's 516 ranged attack power with Trueshot Aura.
