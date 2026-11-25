# EV-0195: Every-hit main-hand procs fire on landed specials that deal no damage; rolled and either-hand procs do not

- Claim: A proc whose row fires on every hit with no roll and states the main-hand-only bit (ProcTypeMask word 1,
  0x20) fires on a landed main-hand special that deals no damage: Sunder Armor, Rend's application, Concussion Blow and
  Disarm cast Iceblade Hacker's 1298414 and Dreadfrost Saber's 1294940. A rolled equip proc (Hand of Justice 15600,
  Ironfoe 1301046) and an every-hit proc without the main-hand bit (Force Imbued Gauntlets 1302248) fire on those
  specials at a quarter or less of their share on a damaging special, Hand of Justice's Disarm on 251 triggers aside.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the warrior wearer streams on disk for each item, in
  the fights whose gear lists the item: Iceblade Hacker 20 reports, Dreadfrost Saber 12, Force Imbued Gauntlets 21,
  Ironfoe 27, Hand of Justice 542; and the rogue wearers of Iceblade Hacker in the fights whose gear puts it in the main
  hand, 6 reports; the reports and fights §Reports lists
- Game build: 1.60.1.70862 to 1.60.1.71091 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-10 to 2026-11-24
- Pull request: #1226

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

Iceblade Hacker and Dreadfrost Saber answer a special that deals no damage about as often as Revenge, Shield Slam and
Heroic Strike. Ironfoe and Hand of Justice answer it at a quarter or less of their share on a damaging special, apart
from Hand of Justice's Disarm on 251 triggers. Force Imbued Gauntlets
reads 14 points above its background on Sunder Armor and 22 on Rend's 120, against 9 on Taunt and 79 on Revenge: the
white swings its wide window catches, not a proc of its own.

Ironfoe is both rolled and an extra attack, so these counts do not tell which of the two keeps it off; Hand of Justice is
both as well.

## Rogues

The same count on the rogue wearers of Iceblade Hacker, in the fights whose gear puts it in the main hand: Garrote 93.9 /
12.1 (33), Kidney Shot 68.2 / 9.1 (22), Rupture 73.3 / 20.0 (15), against Sinister Strike 78.3 / 25.3 (3,152) and
Eviscerate 66.4 / 11.5 (547). No Death Mark landing falls in these fights.

## Reports

Report codes and fights, from the fights whose gear lists the item.

Iceblade Hacker, warriors: 3mavCA2tLRc9qwMy 4, 21, 26, 28, 29, 41; 3r4pzgfc172HvbwC 2, 9, 12, 17, 19, 21;
7fGLbPgqwYDRr2xt 41; 7wGT81AkLYy6dHZF 1; AP4vBCgQj78qD6zW 3; B4wWqM2TbRmdQy7Y 1; CmZh47GDdxbMPRqg 3; DbQwz3av89BpZYfn 1;
Ft3TAra68jH4CK2G 1; K3xmjRCnMpV2c4qr 42; Pk2mFQLyAvxRXVcb 2; QFktMKaqvbLWP89d 1; Wn9tVQrwFb2MGChP 36;
YnbJrPgHjLMqwCWx 6, 9, 13, 18, 23, 27, 31, 32, 38, 48, 53; ZGrhmwyRv7HKpqX1 2; gc2BTyJdxXAR7h4D 4; mtgfAK36pdbGM7wH 1, 3;
rJaTCp796xwvqL1g 1, 4; rTKYMd4yDvaL6Gn8 1; thqTyPArmkxfMnK7 5, 23, 29, 36.

Iceblade Hacker, rogues: 8HvfKkArXpmbhwTF 1; A1yxMTmKatkf3HRJ 15; DNc6QALhxmGbCH3Y 1; Hn8yXmQgGJdZawL3 1;
JQwbjnVYyGcP4rg8 3, 6; KYmFPtyJ3gkzHfVM 102, 106, 111, 118, 120, 122, 127, 128, 130.

Dreadfrost Saber: 1NBKq2mDWghwR69z 4; 4ynKGp8XCQPNmJLF 5, 6; BFWZDP47JX9ynkCh 1; HqwQhrTj28a19Dfx 1; QJRvHKtmWqNG6hc7 7;
WYrCyJkdfMXw84Ph 13; mFTa4Zwkcnf1GCVx 1; mzFnf3hbjVK6kYBX 5; rL2hdXv8cKt3PQM4 1; v2aNxQ7mGTdkJryV 1;
vgrGc2Zt6BzQ849V 1, 4, 6; ynVw7ZgMx4tjF3fQ 15.

Force Imbued Gauntlets: 2DC7xQNm9wHJnVF4 6; 4f2ygdYCzFAvqc6K 6, 10, 13; 9JZtRr8xj46N3fng 1, 3; 9nX7rz3h2bvPJBaH 1;
B1WXm7QGV4hdHLbK 32; Bg6fnXdYLqHtRhwp 1; DH3dT6tYLKmMrNz8 6; Fhdty4J8jrDV7cbn 1, 3; HmF9nyBdrvAL1wKt 1, 24;
JmTh2g1fbrqZaRwH 1; MB1yTdQ9xmRXPaK8 1, 11; RJXQC67jbzATWdx1 1; Tmbn84VwC13A7kDr 39; W3DMV9g1FxbkAjaP 1;
aZKTvVcCHLwhgXzM 6; bMVJtkRrhZpYfa8d 2; f36RaB12wyDcTCGY 1; h4gWnT3HLBbYzMcK 1; k1taHbwL76Cnxv4m 4;
m6htAf2g7FvD3WVj 1, 3, 5, 7, 9, 11, 13, 15, 17; rYgfPHhG4DKxZF3p 1.

Ironfoe: 1MA3qWLGa8v2hdXP 3, 41, 42; 6fxJHmdyR8DTLhBt 3, 17, 23, 25, 35, 37, 40, 44; BFtRGWmNv8JV1Tbq 2;
DTMvbWAqcHGPn8xg 11, 14, 18, 26, 29, 31, 40, 43; F27ZvazHdY4RjVAQ 1; FCfzwGJ2NqBZtYcb 1; FLtpV1h8GTz6rgAc 1;
Hn8yXmQgGJdZawL3 1; KLRwMraFATG1XnWZ 1; PqK7NWmpYF1zA4fT 155; Yvtw3nWdD7Tz2pxh 1, 5; Z6Xm92K3zgfrqM84 1, 6;
aFtMHKAG1vjTCwcq 4, 15, 21, 26, 28, 34; apMdm4GJH32L1DVC 1; c486my2FLjHPbDX7 22;
c9yNWjx1PdfgFrmX 8, 12, 16, 21, 31, 34, 36, 42, 44; daTq1vcjC3QBntb8 4, 21, 27, 35, 37; fprTQwWyDJK7gamC 3;
gT3zDPraq8KMxH6b 1, 5; gdBnpzCGaJ2THt3y 1; gmfpNjd3zBwkr64b 2; n7dy3G4VmhZrzF6R 7; nXZWfk1jK83YATyP 4;
tvCnHVb9xhpAT2XK 1; v8L6NkQTzMV2mWq3 5; vT1rMa2cNGXLfqp7 9, 12; vYjL3q712CyTGkQh 1, 15.

Hand of Justice: 542 reports, too many to list; every warrior wearer stream on disk with the item in the fight's gear,
16GhDNRmVn8KWjBT and 1AJgnPFRhjQtacVW among them.
