# EV-0092: Spell Piercing on DoT ticks follows the spell

- Claim: Spell Piercing on a tick follows the spell: Devouring Plague's, Pyroblast's and Starshards' ticks take the bonus
  at the caster's rate on direct hits and Frostfire Bolt's at about that rate, while Shadow Word: Pain's, Corruption's,
  Bane of Agony's, Immolate's and Fireball's ticks take none.
- Sources: Warcraft Logs reports of Forever (zone 3000) started 2026-11-19 23:03 to 2026-11-27 05:50 UTC: the damage
  of 9,019 casters in 5,352 reports: every mage of the reports read, every wearer of Blade of Eternal Darkness, the
  priests of reports that show Devouring Plague or Starshards, the casters of every class with 4 or more Spell Piercing
  on their gear and 8% of the other casters; the research folder `forever-piercing-posthotfix-1127`
  (`tables/reports.txt`) lists the reports
- Game build: 1.60.1.70959 or 1.60.1.71037 until 2026-11-20 18:23 UTC, 1.60.1.71037 until 2026-11-24 15:21 UTC (US) and
  2026-11-25 02:19 UTC (EU), 1.60.1.71091 until 2026-11-25 19:14 UTC and 1.60.1.71139 from then; all after the
  2026-11-19 hotfix (hotfixes 114309, 114310, 114312, 114313 and 114314). The hotfix a run was on and relogs are not
  recorded
- Date: 2026-11-19 23:03 to 2026-11-27 05:50 UTC
- Pull request: #1267

Warcraft Logs drops the negative resisted amount that Spell Piercing logs, so a bonus is a normal hit or tick with
nothing resisted whose damage is 1.1, 1.2 or 1.3 times the median of the same caster's other same-school hits on the
same mob within 10 seconds (within 2.5%), on mobs whose plain direct hits of that school show 3% or fewer partial
resists; a spell of two schools counts on mobs clean in both. Read against the hit before resists instead, one report,
cvRTFb1qhGBKNkCD, was checked one to one against a combat log of 1,425 hits before the 2026-11-19 hotfix: every −10% or −20% hit of the
combat log lands in the matching bin and no unpierced hit does; the median reading cancels the fixed +10% or +20% that
talents and debuffs put on whole spells, which look the same on one hit. Warlocks' Shadow hits 0 to 12 seconds after their own Shadow Bolt crit on the mob (Improved Shadow
Bolt) are left out. Ticks count from 50 damage before resists, except Frostfire Bolt's and Fireball's, which are 9 to 24:
these are read on whole numbers, the tick against 1.1, 1.2 or 1.3 times its damage before resists and the median of the
caster's other ticks of the spell on the mob, within 1 point, and a tick that fits two steps or none is left out. E is
the hits times the caster's own bonus rate on its class's other spells (mages: Arcane Explosion, Arcane Missiles, Arcane
Blast, Fire Blast, Blizzard, Fireball's and Pyroblast's hits, Scorch; priests: Mind Blast, Shadow Word: Death, Mind
Flay's ticks, Smite, Holy Fire's hit; warlocks: Conflagrate, Hellfire Effect, Incinerate, Bane of Havoc, Searing Pain,
Soul Fire, Immolate's hit), over casters with 30 or more such hits (mages 40) of which 5% or more (mages 12%) take a
bonus, 40 to 90% of those at +10% and one or more at +20% (mages: a fifth or more at +20%). The floor is the rate of
casters of the class without Spell Piercing on the same ticks. P at the full share is the Poisson chance of a count this
low at E plus the floor, and P at the floor that of a count this high at the floor's rate.

| Ticks                | With a bonus (E), +10% : +20% : +30%          | Over E | P at the full share | Floor         | P at the floor | Casters |
| -------------------- | --------------------------------------------- | ------ | ------------------- | ------------- | -------------- | ------- |
| Devouring Plague     | 3,454 of 33,099 (3,332.2), 2,480 : 970 : 4    | 1.04   | 0.12                | 29 of 4,996   | < 1e-15        | 123     |
| Pyroblast's DoT      | 9,833 of 32,617 (9,543.4), 6,879 : 2,947 : 7  | 1.03   | 0.11                | 14 of 1,101   | 3e-13          | 1,107   |
| Starshards (channel) | 205 of 1,826 (201.7), 134 : 71 : 0            | 1.02   | 0.42                | 5 of 1,369    | < 1e-15        | 13      |
| Frostfire Bolt's DoT | 426 of 2,121 (516.5), 282 : 136 : 8           | 0.82   | 2e-11               | 23 of 784     | 1e-14          | 164     |
| Fireball's DoT       | 159 of 7,009 (1,942.2), 150 : 8 : 1           | 0.08   | < 1e-15             | 40 of 1,218   | 1              | 274     |
| Immolate's DoT       | 104 of 35,152 (3,516.3), 102 : 2 : 0          | 0.03   | < 1e-15             | 3 of 2,212    | 1e-12          | 250     |
| Bane of Agony        | 1,777 of 262,090 (28,005.9), 1,491 : 257 : 29 | 0.06   | < 1e-15             | 124 of 26,543 | 2e-13          | 425     |
| Corruption           | 201 of 148,506 (16,032.0), 195 : 6 : 0        | 0.01   | < 1e-15             | 25 of 18,295  | 0.56           | 412     |
| Shadow Word: Pain    | 64 of 80,409 (8,111.3), 57 : 6 : 1            | 0.01   | < 1e-15             | 27 of 14,694  | 1              | 129     |

Immolate's and Bane of Agony's ticks run above the floor, at 0.03 and 0.06 of E. Bane of Agony is spell 980 to 11713,
which Warcraft Logs names Curse of Agony; its +30% bonuses are a step the roll does not give at these shares. Split at 2026-11-23 15:45, 2026-11-24 15:21 and 2026-11-25
19:14 UTC, its rate over E moves between 0.04 and 0.09, Fireball's between 0.03 and 0.11 and Corruption's between 0.008
and 0.015, apart at P < 0.01 but always far under the full share.

With E from the gear's piercing and level instead, 7.5 × p / (p + 8L + 19) a hit, Devouring Plague reads 3,354 of
34,272 (3,912.8, 0.86) and Starshards 188 of 1,923 (237.2, 0.79); casters other than mages read below 7.5 × s on
plain nukes as well (EV-0142, §The 1% cut).

Frostfire Bolt's ticks are 9, 13 or 19 damage before resists (ranks 1 to 3), where a step under 2 points does not
survive whole numbers. The rank 3 ticks (19 to 24 before resists) read 350 of 1,511 (E 364.5, 0.96), their +20% bonuses 103 against 121.5 at
the full share, and the floor 20 of 436, all at +10%; the ticks of 13 to 18 read 0.54 and those of 9 to 12 none of 54.
The same whole-number reading on known full-share ticks gives, at 50 or more before resists, Devouring Plague 0.97 and
Pyroblast 0.91 of E, and at 25 to 49, Devouring Plague 49 of 469 (E 82.2, 0.60) and Pyroblast 32 of 102 (E 28.9, 1.11),
so it keeps about 0.6 to 1.1 of a full-share tick's bonuses and Frostfire Bolt's 0.82 and 0.96 lie inside that range;
whether its ticks take exactly the full share is not settled.
Fireball's ticks are mostly 13 to 18 before resists: those read 136 of 5,596 (E 1,575.5) with 8 at +20%, where the full
share gives about 525 at +20%, and still about 280 at the 0.54 this reading keeps of Frostfire Bolt's ticks of that
size.

A tick at a tenth of the share and one that takes no bonus give the same count: a tenth of the share stays under the
1% that any bonus needs (EV-0142) up to 55 piercing at level 60.

No row property separates the two groups: aura type (Pyroblast's DoT is aura 3 like Shadow Word: Pain), dispel type,
the attribute words and the piercing auras (none carries a class mask) were all checked. Frostfire Bolt, Immolate and
Bane of Agony are binary spells (EV-0142).
