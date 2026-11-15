# EV-0121: Skullforge Reaver's rate from the order of the log lines

- Claim: Skullforge Reaver 13361's Skullforge Brand (17484) procs 1.85 times a minute on main-hand hits, white swings
  and melee casts alike.
- Sources: public reports on forever.warcraftlogs.com, zone 3000, with the fights each wearer had it on. Rogue:
  123jYCvkVGrHfXWt fights 1 to 30, 1z4vRDqGdM2cVPXk 1 to 8, 2wMznDRvmkpbtHxA 1 to 7, 9zxHaWbYT3gqDQvF 17 to 18,
  CfHvwytKJ6P4q7ag 1 to 24, ZfbY6RgWjx9hmpGz 1 to 14, vZqVmLYQa6jwP3A9 5 to 21, vwYF4Z2C38cfKTDB 1 to 9. Warrior:
  NXMZV9xjvWpB7GtP 28 to 122. Hunter: NXMZV9xjvWpB7GtP 107 to 122.
- Game build: 1.60.1.70938; the hotfix state and relogs were not recorded. No hotfix on this build changed 17484, the
  item's on-hit entry or its 2.5 s speed.
- Date: 2026-11-14 to 2026-11-15
- Pull request: not recorded

A wearer is a player whose combatant info lists the item, and a run that two wearers logged counts once. The brand's
applydebuff or refreshdebuff line follows the damage line of the hit that rolled it in the same millisecond: of 2,027
attributed procs, 1,888 share the hit's millisecond and 132 come within 10 ms after it. Each proc therefore goes to
the last landed hit logged before it within 50 ms, or failing that the first hit after it within 30 ms; 32 procs
matched no hit. A melee cast that lands with no damage line, such as Cheap Shot, counts at its debuff line.

Every wearer held the sword in the main hand with a weapon in the off hand. White swings are split by hand for each
run on their damage: a two-group split on the log of the damage before armor, which the log gives before the critical
strike bonus. The rate is the chance times 60 over the sword's 2.5 s speed. A proc on a target that already carries
the brand logs refreshdebuff, so it refreshes the brand.

| Hit                    | Procs / hits   | Chance | PPM [95%]         | Wearer bootstrap [95%] |
| ---------------------- | -------------- | ------ | ----------------- | ---------------------- |
| Main-hand white swing  | 915 / 11,599   | 7.89%  | 1.89 [1.78, 2.01] | [1.65, 1.97]           |
| Main-hand melee cast   | 861 / 11,490   | 7.49%  | 1.80 [1.69, 1.92] | [1.73, 1.96]           |
| Windfury Totem strike  | 242 / 3,731    | 6.49%  | 1.56 [1.38, 1.76] |                        |
| Off-hand white swing   | 9 / 12,870     | 0.07%  |                   |                        |
| Ranged-type hit        | 0 / 579        | 0%     |                   |                        |
| White swings and casts | 1,776 / 23,089 | 7.69%  | 1.85              |                        |
| Every eligible hit     | 2,018 / 26,820 | 7.52%  | 1.81              |                        |

The melee casts are Sinister Strike, Eviscerate, Whirlwind, Bloodthirst, Cleave and Cheap Shot's debuff. White swings
and melee casts roll at one rate. Windfury Totem strikes roll it too, at a lower count (1.56, interval 1.38 to 1.76);
whether the strike rolls at a rate of its own is not settled by 242 procs. Off-hand swings and
ranged-type hits never roll it. The rate is the pooled white swings and melee casts, 1.85. The counts come from 6
wearers (4 rogues, 1 warrior and 1 hunter) over 9 runs.
