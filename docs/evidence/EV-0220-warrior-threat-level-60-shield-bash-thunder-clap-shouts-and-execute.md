# EV-0220: Warrior threat at level 60: Shield Bash, Thunder Clap, the shouts and Execute

- Claim: In Defensive Stance with Defiance 3 a warrior's threat is 2.16 times the action's own; Shield Bash Rank 3 makes
  1.5 times its damage as threat and Thunder Clap Rank 6 1 times, Demoralizing Shout Rank 5 makes 47 threat a target and
  Battle Shout Rank 6 none, each times the stance, and an Execute makes 1.25 times its damage as threat, times the
  stance, which in Battle Stance is its damage.
- Sources: combat logs, ThreatLog and StatLedger of one protection warrior at level 60 in dungeon groups; no report
  codes; for the Berserker Stance Execute, a screenshot of another in-game threat addon on the same warrior
- Game build: 1.60.1.71091 on 2026-11-24 and 1.60.1.71139 from the 2026-11-25 18:17 login; the hotfix is not recorded;
  no reading moves between the two builds; the screenshot's build and hotfix are not recorded
- Date: 2026-11-24 to 2026-11-26
- Pull request: #1254

Talents: Defiance 3, Improved Revenge 3, Improved Thunder Clap 3, Improved Shield Bash 2, Booming Voice 5, Concussion
Blow and Shield Slam; no Improved Heroic Strike, Improved Bloodrage or Unbridled Wrath, and no Blessing of Salvation on
the warrior.

7,744 of the 8,169 threat log rows name no mob, since the logs are group content. Each row's mob is read from the combat
log instead: the one creature the warrior's swings and single-target abilities hit in the window. A window is kept when
the row and the one before it resolve to the same mob, the two are at most 3 s apart, and the row's label names the
casts the combat log shows; 4,899 of 5,402 windows pass. On the rows that name a mob, 158 of 164 windows resolved this
way find that mob. The addon's clock is set against the log once a login, 0.4 to 2.9 s apart, with no step inside one.

From each window's threat every term an earlier reading covers is taken out, each event rounded to the nearest whole
point: white swings, damage shields, item procs, Rend, rage, and Shield Slam, Revenge, Thunder Clap, Shield Bash,
Demoralizing Shout and Sunder Armor at the values below and in their homes (EV-0050, EV-0204, EV-0211). What is left is
the ability's own threat, divided by the stance. A strict window holds the ability with white swings, damage shields,
item procs and Rend alone; a loose window may hold the measured abilities as well. Intervals are 95%.

| Reading                                        | Measured                                                                                                                                                     |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Defensive Stance with Defiance 3, white swings | 2.161 [2.160, 2.162] times the damage (n 89); round(2.16 x damage) per swing exact on 59, against 0 for 1.8 and 2 for 2.18                                   |
| Shield Bash Rank 3 (1672)                      | 1.502 [1.502, 1.531] times the damage (n 9, strict); a flat bonus would read 0.22 [0.11, 1.50]                                                               |
| Thunder Clap Rank 6 (11581)                    | 1.014 [1.000, 1.015] times the damage on each mob (n 7, strict), 1.006 on the 4 with one target; 1 to 4 targets                                              |
| Demoralizing Shout Rank 5 (11556)              | 47 a target: 102 (round(2.16 x 47)) on 2 strict windows and on 10 of 20 rows whose label holds the shout alone, 4 more of them 162; round(2.16 x 48) on none |
| Battle Shout Rank 6 (11551)                    | 0: 0.0 on 4 of 5 strict windows, 0.9 on the fifth                                                                                                            |
| Execute (20647), Battle Stance                 | threat equal to the damage on 3 of 8 windows (600, 695 and 751); 3 of the other 5 hold a Windfury Totem hit, 1 a Shield Spike hit and 1 a stance change      |

Every Execute window in the logs is in Battle Stance, where 1.25 x 0.8 and a threat of 1 times the damage read the
same. The screenshot, taken 2026-11-26 in Berserker Stance (Berserker Stance Passive 7381 states -20% threat, so 0.8),
separates them. One threat update at Execute (20662) reads +1155, after a melee crit for 379 and an Execute for 852:
(1.25 x 852 + 379) x 0.8 is 1155.2. A threat of 1 times the damage with the swing gives 984.8, and the Execute alone
852 or 681.6, so neither fits. The update before it is not in the screenshot, so what the window holds is read from
the order of the chat lines; n 1.

Other readings in the same windows:

| Reading                   | Measured                                                               |
| ------------------------- | ---------------------------------------------------------------------- |
| Damage shields            | 0.995 [0.992, 1.006] times the damage (n 15)                           |
| Windfury Totem hits       | 1.065 [1.033, 1.118] times the damage (n 8)                            |
| Retaliation hits          | about 1 times the damage on 13 crowded windows                         |
| Disarm, Challenging Shout | Disarm 0 (n 1); Challenging Shout reads −4.2 to 8.5 on 8 of 11 windows |

Not measured: Revenge Rank 6, Heroic Strike Rank 9, Battle Shout Rank 7, Whirlwind, Overpower and Slam (never cast);
Mocking Blow, Hamstring, Rend Rank 7, Intercept, Piercing Howl and Taunt (too few windows); rage threat at level 60,
which dungeon threat lists blur.
