# EV-0021: Dragon's Call, its whelp, the Highborne Research Tablet and weapon procs on shield hits, level 57

- Claim: Dragon's Call procs 4.5 times a minute behind a 45 s cooldown and summons a whelp that takes nothing from its owner; the Highborne Research Tablet procs off damaging hits only; and Fiery Blaze, Rock on Stick and Dragon's Call do not proc off Shield Slam, Shield Bash or Thunder Clap, while equip procs still do. Highborne Research Tablet is the exception: Thunder Clap never procs it (EV-0131).
- Sources: combat log, one level 57 Protection warrior with a 2.5 s main hand (Mount Hyjal, about 27 minutes in combat, StatLedger for gear and level); Warcraft Logs FHqYCNcKntvDAjxg, WRa9p2xhPnQKVG8C and t6bL9xfVY3hXyvHT for Fiery Blaze (3 players)
- Game build: not recorded
- Date: 2026-11-12 for the combat log; not recorded for the Warcraft Logs reports
- Pull request: #958

## Highborne Research Tablet (260205)

| Reading             | Measured                                |
| ------------------- | --------------------------------------- |
| Procs               | 70, from 647 landed damaging melee hits |
| Per hit             | 10.8% (8.5 to 13.5)                     |
| Per minute at 2.5 s | 2.6 (2.05 to 3.23)                      |
| Shortest gap        | 0.12 s, so no internal cooldown         |

The rate is the same on white hits, Shield Slam and Revenge. Rend and Concussion Blow applications, Rend ticks and damage
shields never proc it, so only a hit that deals damage rolls. An Auto Shot from a hunter procced it too. The chance on a
spell hit is unmeasured here; EV-0131 measures it, with the ranged and off-hand rates, in 55 Warcraft Logs reports.

## Dragon's Call (10847)

| Reading             | Measured                                                                                                       |
| ------------------- | -------------------------------------------------------------------------------------------------------------- |
| Procs               | 41, from 220 landed melee hits outside the cooldown (Shield Slam left out)                                     |
| Per hit             | 18.6% (13.7 to 24.4)                                                                                           |
| Per minute at 2.5 s | 4.5 (3.3 to 5.9)                                                                                               |
| Cooldown            | 45 s: shortest gaps 44.94, 45.03 and 45.16 s, and none of 73 hits in the last 5 s of a cooldown procced        |
| Shield Slam         | 1 proc in 50 landed outside the cooldown, about 9.5 expected; that one shares its millisecond with a white hit |

Revenge, Heroic Strike, Cleave and Victory Rush proc it at the same rate as white hits, and so do Rend and Concussion Blow
applications. Misses, dodges and parries do not. In the sim, 400 runs of 300 s gave 18.98% per landed hit outside the
cooldown and a shortest gap of 45.000 s.

## Weapon procs and the shield hits

Fiery Blaze is counted on the three Warcraft Logs reports and Dragon's Call on the level 57 log; the pull request does
not say which logs hold the other rows.

| Proc                                | Kind           | Shield Slam                        | Shield Bash       | Thunder Clap | White hits  |
| ----------------------------------- | -------------- | ---------------------------------- | ----------------- | ------------ | ----------- |
| Fiery Blaze (enchant 36), 3 players | combat enchant | 0 of 617                           | 0 of 78           | 0 of 436     | 212 of 1517 |
| Rock on Stick (279852)              | chance on hit  | 0 of 157                           |                   |              | 41 of 528   |
| Dragon's Call (10847)               | chance on hit  | 1 of 50                            | 0 of 4            |              |             |
| Dreadfrost Saber (274603)           | equip aura     | 60 of 105, and 54 of 80 in EV-0012 | 7 of 8 in EV-0012 |              | 208 of 311  |
| Highborne Research Tablet (260205)  | equip aura     | about 11%                          |                   |              | about 10%   |

Dragon's Call's one Shield Slam proc shares its millisecond with a white hit. Non-damaging Sunder Armor, Concussion
Blow and Rend applications proc Fiery Blaze and Dragon's Call; Rend ticks proc neither.

## The Emerald Dragon Whelp (8776) and Acid Spit (9591)

The log has 40 summons that read the same health, armor and attack power while the owner's attack power ran from 981 to
1383, and its health, armor and Acid Spit fit the creature values for level 57. The whelp never crits: one crit in 96
landed swings and none in 82 spits.

| Reading                            | Log (level 57 owner)                                                                                                                                                                           | Sim (level 60)                                                                        |
| ---------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| Level                              | the owner's                                                                                                                                                                                    | the owner's                                                                           |
| Health                             | 5082, which is 1.3 times the creature health at 57                                                                                                                                             | 1.3 times the creature health                                                         |
| Armor                              | 1735, which is 0.51 of the creature armor at 57 (ExpectedStatMod 1's share in the game files)                                                                                                  | 0.51 of the creature armor                                                            |
| Attack power                       | 174, which the swing leaves out                                                                                                                                                                | none                                                                                  |
| Swing                              | 71 to 94 before armor, 80 on average, at level 57; 0.73 of the creature dps for 2 s gives 83 plus or minus 15%; 0.5 with the 174 attack power gives 82, so the log does not tell the two apart | 2.0 s at 0.73                                                                         |
| Glancing                           | 87 of 96 landed swings, 79% against targets at the whelp's own level, at about 57% of a hit; the swings were 13% of the whelp's damage                                                         | the attack table's 10% glance, 6% miss and 4% parry; no rule explains the log's share |
| Acid Spit                          | 204 to 286, mean 250.6 over 82 hits; the row's 438.5% of the creature auto attack dps gives 249.7 at level 57                                                                                  | 265 mean at level 60                                                                  |
| Acid Spit mana                     | 5340 at level 57, more spits than a summon lasts                                                                                                                                               | no cost                                                                               |
| First spit                         | 0.00, 0.01, 0.89 and 2.59 s after the summon at the 0, 25, 50 and 75% marks over 41 summons                                                                                                    | uniform 0 to 1.8 s                                                                    |
| Wait from a cast end to next start | 0.20, 0.58, 0.99 and 1.88 s at the same marks over 69 waits, maximum 6.9 s, all with the target alive; mean 1.86 s over the 16 full summons that lasted 15 s with no target dying              | uniform 0.2 to 3.5 s, mean 1.85 s                                                     |
| Spits per summon                   | 2.31 (clean full summons)                                                                                                                                                                      | 2.86                                                                                  |
| Swings per summon                  | 2.94                                                                                                                                                                                           | 2.04                                                                                  |
| Damage per summon                  | 665 (clean full summons)                                                                                                                                                                       | 873, about 20% above the log scaled to level 60                                       |

No swing lands inside a cast that reaches its end, so a spit's hardcast holds the swing and the whelp swings only as it
arrives and in the waits between casts.

The Dragon's Call and Highborne Research Tablet rates are override rows on the trigger rows (10847 and 1318159) and the
Acid Spit row 9591 is added by hand, since a creature's spells are on the server and no row links it.
