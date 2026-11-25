# EV-0195: Every-hit main-hand procs fire on landed specials that deal no damage; rolled and either-hand procs do not

- Claim: A proc whose row fires on every hit with no roll and states the main-hand-only bit (ProcTypeMask word 1,
  0x20) fires on a landed main-hand special that deals no damage: Sunder Armor, Rend's application, Concussion Blow and
  Disarm cast Iceblade Hacker's 1298414 and Dreadfrost Saber's 1294940. A rolled equip proc (Hand of Justice 15600,
  Ironfoe 1301046) and an every-hit proc without the main-hand bit (Force Imbued Gauntlets 1302248) fire on those
  specials no more than on Taunt, which hears nothing.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the warrior wearer streams on disk for each item, in
  the fights whose gear lists the item: Iceblade Hacker 20 reports, Dreadfrost Saber 12, Force Imbued Gauntlets 21,
  Ironfoe 27, Hand of Justice 542
- Game build: 1.60.1.70938 to 1.60.1.71091 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-12 to 2026-11-24
- Pull request: #TBD

## Method

Each warrior wearer's own events in the fights whose gear lists the item. A trigger is a landed event: a damaging
special's damage event with a hit, crit, block, blocked crit or glance, and for a special that deals no damage the
debuff it puts up (applied, stacked or refreshed: Sunder Armor, Rend, Concussion Blow, Disarm, and Taunt as the
control). A trigger counts as answered when the item's event follows on the same target inside the window: the Frost
damage 330 to 500 ms after (Iceblade Hacker's 0.4 s travel), Dreadfrost Saber's and Force Imbued's damage −80 to
500 ms, an extra attack −80 to 80 ms. The background is the same window 1 s later. Wide windows catch the proc of a
white swing near the trigger, which is why Taunt reads above its background on the damage procs.

## Counts

Answered share, then background, in percent; n is the triggers counted.

| Trigger               | Iceblade Hacker     | Dreadfrost Saber    | Force Imbued Gauntlets | Ironfoe              | Hand of Justice       |
| --------------------- | ------------------- | ------------------- | ---------------------- | -------------------- | --------------------- |
| Sunder Armor          | 89.8 / 6.7 (2,136)  | 86.4 / 15.2 (1,728) | 32.9 / 18.9 (4,621)    | 0.65 / 0.13 (1,545)  | 0.19 / 0.08 (37,550)  |
| Rend, applied         | 98.3 / 7.7 (467)    | 86.3 / 12.0 (598)   | 40.0 / 18.3 (120)      | 0.90 / 0 (222)       | 0.31 / 0.04 (4,815)   |
| Concussion Blow       | 95.1 / 8.6 (222)    | 80.0 / 20.0 (30)    | 29.7 / 27.9 (226)      | 0 / 0 (57)           | 0.05 / 0.05 (1,964)   |
| Disarm                | 100 / 9.3 (43)      | 70.6 / 17.7 (17)    | 33.3 / 44.4 (9)        | 0 / 0 (12)           | 0.80 / 0 (251)        |
| Taunt (hears nothing) | 11.5 / 11.2 (767)   | 27.3 / 20.5 (297)   | 32.3 / 22.9 (961)      | 0.33 / 0 (303)       | 0.17 / 0.09 (9,528)   |
| Revenge               | 93.1 / 7.3 (2,993)  | 85.4 / 13.2 (793)   | 94.7 / 16.0 (2,556)    | 5.28 / 0.40 (757)    | 1.41 / 0.09 (33,834)  |
| Shield Slam           | 92.5 / 6.0 (3,460)  | 75.0 / 11.2 (232)   | 93.7 / 15.5 (3,347)    | 4.96 / 0.41 (987)    | 1.30 / 0.07 (38,363)  |
| Heroic Strike         | 95.3 / 4.7 (1,202)  | 84.9 / 26.2 (218)   | 96.6 / 14.1 (446)      | 5.89 / 0.11 (950)    | 1.95 / 0.10 (17,509)  |
| Main-hand white swing | 90.4 / 4.6 (11,606) | 88.7 / 23.5 (8,155) | 95.7 / 11.2 (10,833)   | 11.5 / 0.53 (23,581) | 3.07 / 0.15 (446,104) |

Iceblade Hacker and Dreadfrost Saber answer a special that deals no damage as often as Revenge, Shield Slam and Heroic
Strike. Ironfoe and Hand of Justice answer it at a tenth or less of a damaging special's share. Force Imbued Gauntlets
reads 14 points above its background on Sunder Armor and 22 on Rend's 120, against 9 on Taunt and 79 on Revenge: the
white swings its wide window catches, not a proc of its own.

Ironfoe is both rolled and an extra attack, so these counts do not tell which of the two keeps it off; Hand of Justice is
both as well.
