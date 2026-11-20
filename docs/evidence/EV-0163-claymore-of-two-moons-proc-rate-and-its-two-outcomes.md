# EV-0163: Claymore of Two Moons' proc rate and its two outcomes

- Claim: Claymore of Two Moons procs at about 1.35 PPM on landed melee hits, read off paladins alone, so the rate is
  observed. Each proc casts Waxing Moon or puts up Waning Moon, one in two each, never both.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every report with a wearer: JZDFdhkvn3NP6xLm 3,
  Lpr76Vd3DbKWZqnh 3, YVWgj6ra27PvmNnx 1 and 3, YgNvCqfjP9GwXWyx 6; 9fRTQzPJXtrKgckV 1 logs the same run as
  JZDFdhkvn3NP6xLm 3 and counts once
- Game build: 1.60.1.70938 (YgNvCqfjP9GwXWyx and YVWgj6ra27PvmNnx, US; Lpr76Vd3DbKWZqnh, EU) and 1.60.1.70959
  (JZDFdhkvn3NP6xLm, US), by report date and region; the hotfix each player ran is not recorded. Hotfix push 17057296
  added the item with build 1.60.1.70862 on 2026-11-09, and no later push touches its rows or the rows of 1309550,
  1319378 and 1319382 (Raidbots' DBCache files from 2026-11-12 to 2026-11-20); those rows are the same on every build
  from 1.60.1.70862 to 1.60.1.70959
- Date: 2026-11-13 to 2026-11-18
- Pull request: not recorded

## Method

The wearers are 4 paladins, all with the item's 3.1 s two-hander in the main hand. A wearer counts between their first
and last proc, leaving both out, since two wearers swapped the weapon in and out. The run in YgNvCqfjP9GwXWyx is from
before the 2026-11-13 hotfix that stopped a paladin's Judgements rolling weapon procs, and holds 2 of the procs counted.

- Lone white swings: landed white swings with no other rolling hit of the wearer from 100 ms before to 600 ms after,
  killing blows left out. A Waning Moon at the swing (5 ms before to 30 ms after) or a Waxing Moon within 600 ms of it
  is a proc. This count needs no roll table and sets the rate.
- All hits: every landed roll, with a white swing, a seal's strike, Holy Strike and a Windfury Totem strike as one roll
  and Judgement of Command as two, as weapon procs roll for a paladin after the 2026-11-13 hotfix. That this proc hears
  the same rolls is not measured; the count checks the rate and sets nothing.

## Rate

| Count             | Procs / rolls | Chance [95%]       | PPM at 3.1 s      |
| ----------------- | ------------: | ------------------ | ----------------- |
| Lone white swings |      26 / 374 | 6.95% [4.79, 9.99] | 1.35 [0.93, 1.93] |
| All hits          |   219 / 3,189 | 6.87% [6.04, 7.80] | 1.33 [1.17, 1.51] |

| Wearer's report  | Rolls | Waxing Moon | Waning Moon | Chance |
| ---------------- | ----: | ----------: | ----------: | -----: |
| JZDFdhkvn3NP6xLm | 1,658 |          61 |          57 |   7.1% |
| Lpr76Vd3DbKWZqnh |    68 |           5 |           4 |  13.2% |
| YVWgj6ra27PvmNnx | 1,414 |          50 |          40 |   6.4% |
| YgNvCqfjP9GwXWyx |    49 |           2 |           0 |   4.1% |

Every wearer swings the same weapon, so a rate per minute and a flat chance per hit cannot be told apart here. The row's
ProcChance of 100 with no rate is read as procs per minute, as EV-0144 measured for such rows. The longest stretch with
no proc is 108 rolls (YVWgj6ra27PvmNnx). At this chance a given run of 108 rolls has none 0.05% of the time, and one
such run turns up somewhere in that report about 4% of the time, so the wearer may have swapped the weapon there with no
snapshot to show it.

## The two outcomes

Over all 227 procs, Waxing Moon is 122 and Waning Moon 105, a Waxing Moon share of 0.54 [0.47, 0.60]. In log order the
pairs read Waxing then Waxing 58, Waxing then Waning 62, Waning then Waxing 62 and Waning then Waning 41: each proc picks
on its own, with no alternation. No hit brings both. A killing blow logs no Waxing Moon (0 on 86 killing hits, 2 Waning
Moons), which reads Waxing Moon about 1 to 3% low against Waning Moon.

- Waning Moon refreshes and never stacks (17 refreshes), and fades 10.0 s after its last application or refresh (86 of
  88). It logs 16 to 17 ms after a white swing, beside a seal's or a Windfury Totem's strike.
- Waxing Moon lands 50 to 250 ms after its hit, the flight of a missile. It deals 189 before mitigation on nearly every
  hit, with no spell power, crits for 1.5 times (29 of 122), is partly resisted on 18 and missed none. Which crit
  chance it rolls is not settled at this count.
