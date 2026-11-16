# EV-0073: Summon Hawk health, attack power and swing after the November 10 hotfix

EV-0132 replaces the swing's roll, read off live logs from level 20 to 60 after the hotfix had reached every realm.

- Claim: at level 48 after the 2026-11-10 hotfix a hawk takes 80% of the owner's attack power as the log shows it, and swings for 0.185 of the creature dps over 2.5 s, 22% either side, with 60 attack power a point of dps.
- Sources: three combat logs of one level 48 hunter taken after the 2026-11-10 hotfix (naked; geared; geared with Aspect of the Hawk), 357 swings in all; report codes not recorded
- Game build: not recorded
- Date: after 2026-11-10
- Pull request: #950

The hotfix raised the melee damage of the Summon Hawk guardian by about 80%. The row that scales a hawk, Hunter Guardian Hawk Pet Scaling 01 (1293586), is not in the store, so the swing is fitted to the logs.

| State                            | Sheet ranged attack power | Owner attack power in the log | Hawk attack power | Swing before armor |
| -------------------------------- | ------------------------- | ----------------------------- | ----------------- | ------------------ |
| Naked                            | 345                       | 352                           | 282               | 28-37              |
| Geared                           | 747                       | 1180                          | 944               | 56-64              |
| Geared, more ranged attack power | about 792                 | 1235                          | 988               | 58-66              |
| The same with Aspect of the Hawk | about 882                 | 1325                          | 1060              | 60-69              |

- The swing is a roll of 17 to 26.5 plus 0.0417 a point of the hawk's attack power, with Unleashed Fury 5/5 and Focused Fire 2/2, which reach the hawks. Logs before the hotfix read 0.0223 a point, so the attack power part is 1.8 times what it was. Over the 357 swings at 282, 944, 988 and 1060 attack power this is 0.185 of the creature dps over the 2.5 s, spanning 0.22 either side, and 60 attack power a point of dps.
- The hawk's attack power is 80% of the owner's attack power as the log shows it. That number is the ranged attack power plus an amount from the weapons: 7 with none, 433 with a 27.32 dps bow and a two-hander. The rule 14 a point of the ranged weapon's dps on a base of 7 gives 390 where the geared log shows 433, so the hawk's attack power is 3.7% low for that hunter (about 2.5% of its damage). A second hunter's log fits the rule to within 2. How the two-hander enters is not known. One level 48 hunter logged 352 at 345 ranged attack power with no weapons, and 1180 at 747 with a 27.3 dps bow.
- The roll is read at level 48 only and scaled to 60 by creature dps.
- Hawk health stays at 5% of creature health. The logs read 6.45% at 48; it does not move DPS.
- The 5% health, the no armor and attack power of its own, and the 30% of the owner's armor are the level 30 readings of #592 (EV-0074).
