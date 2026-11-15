# EV-0121: Skullforge Reaver's rate from the order of the log lines

- Claim: Skullforge Reaver 13361 casts Skullforge Brand (17484) 1.95 times a minute on main-hand hits, white swings
  and melee casts alike, before the brand's own hit roll.
- Sources: public reports on forever.warcraftlogs.com, zone 3000, with the fights each wearer had it on. Rogue:
  123jYCvkVGrHfXWt fights 1 to 30, 1z4vRDqGdM2cVPXk 1 to 8, 2wMznDRvmkpbtHxA 1 to 7, 9zxHaWbYT3gqDQvF 17 to 18,
  CfHvwytKJ6P4q7ag 1 to 24, ZfbY6RgWjx9hmpGz 1 to 14, vZqVmLYQa6jwP3A9 5 to 21, vwYF4Z2C38cfKTDB 1 to 9. Warrior:
  NXMZV9xjvWpB7GtP 28 to 122. Hunter: NXMZV9xjvWpB7GtP 107 to 122.
- Game build: 1.60.1.70938; the hotfix state and relogs were not recorded. No hotfix on this build changed 17484, the
  item's on-hit entry or its 2.5 s speed.
- Date: 2026-11-14 to 2026-11-15
- Pull request: #1040

A wearer is a player whose combatant info lists the item, and a run that two players logged counts once. A proc is a
brand that landed (its applydebuff or refreshdebuff line) or one that failed (a damage line of 17484 that is not a tick,
resisted, immune or missed: 95 of them). Either line follows the damage line of the hit that rolled it in the same
millisecond: of 2,121 attributed procs, 1,961 share the hit's millisecond and 132 come within 10 ms after it. Each proc
therefore goes to the last landed hit logged before it within 50 ms, or failing that the first hit after it within
30 ms; 32 procs matched no hit. A melee cast that lands with no damage line, such as Cheap Shot, counts at its debuff
line.

Every wearer held the sword in the main hand with a weapon in the off hand. White swings are split by hand for each
run on their damage: a two-group split on the logarithm of the damage before armor, which the combat log gives before
the critical strike bonus. The rate is the chance times 60 over the sword's 2.5 s speed. A proc on a target that
already carries the brand logs refreshdebuff, so it refreshes the brand. The brand's drain and heal ticks crit (about 11%
of each).

| Hit                    | Procs / hits   | Chance | PPM [95%]         | Wearer bootstrap [95%] |
| ---------------------- | -------------- | ------ | ----------------- | ---------------------- |
| Main-hand white swing  | 955 / 11,599   | 8.23%  | 1.98 [1.86, 2.10] | [1.74, 2.06]           |
| Main-hand melee cast   | 906 / 11,490   | 7.89%  | 1.89 [1.78, 2.01] | [1.82, 2.05]           |
| Windfury Totem strike  | 248 / 3,731    | 6.65%  | 1.60 [1.41, 1.80] |                        |
| Off-hand white swing   | 12 / 12,870    | 0.09%  |                   |                        |
| Ranged-type hit        | 0 / 579        | 0%     |                   |                        |
| White swings and casts | 1,861 / 23,089 | 8.06%  | 1.93 [1.85, 2.02] | [1.82, 2.01]           |
| Every eligible hit     | 2,109 / 26,820 | 7.86%  | 1.89 [1.81, 1.97] | [1.59, 1.99]           |

The melee casts are Sinister Strike, Eviscerate, Whirlwind, Bloodthirst, Cleave and Cheap Shot's debuff. White swings
and melee casts roll at one rate. Windfury Totem strikes roll it too, at a lower rate (1.60, interval 1.41 to 1.80);
248 procs do not settle whether the strike rolls at a rate of its own. Off-hand swings give 12 procs in 12,870 and
ranged-type hits none in 579, against 8.23% for main-hand swings, so neither rolls it. The rate is the pooled white
swings and melee casts, 1.93, rounded to 1.95 at the nearest 0.05 step. The counts come from 6 wearers (4 rogues, 1
warrior and 1 hunter) over 9 runs.
