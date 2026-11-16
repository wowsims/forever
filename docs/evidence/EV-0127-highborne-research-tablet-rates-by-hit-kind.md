# EV-0127: Highborne Research Tablet rates on melee, ranged and spell hits

- Claim: Highborne Research Tablet (260205) procs about 2.6 times a minute on a weapon hit's own weapon speed, Auto
  Shot included, and at a flat chance of about 7% on a landed spell hit whatever the main hand's speed; periodic ticks
  and damage shields never proc it.
- Sources: Warcraft Logs, every player wearing the tablet in 55 reports: 1GBhzTR8CZwcJfFV, gBDMLv8J1K7nqkpd,
  FHqYCNcKntvDAjxg, 2HMyR6kfbWKzBhtv, JK1z96cCrj28tpBV, mhbXcFqf9xTaA1NP, JFqRThp6Z4HrnBwf, CF7ztDwY8pJ9KZhn,
  KqZ6xWAfDLGrVjYJ, A7ahGYbzQ8M1ZtF6, apNB26JmYPyhFWjH, rPpHYfq3tNnwJRMd, 8jzXTaqr732V9DxJ, byvnpga9tMq3BfX6,
  NXMZV9xjvWpB7GtP, 1X37tZRzgkmd9rNP, dxZCXJvPLKAB3cnQ, vyWAYhM3CFkRf12B, Zh9LTdGmgbW1BXw3, 1nTQBK4hmcaLbFZV,
  rphdM14FbnTBaj3V, 1ZyhaVDdFtTYWMcC, TkdjxNKfy9wWDhQp, AvNwTdyBbDncFf2h, 7tLY1pVK9wRWPABX, 9zxHaWbYT3gqDQvF,
  apcNj7QY8tPhLw2B, VkLX4CZnc68R2Gb1, wJtPdpxFGQCM9f3c, Q7V24RA9HTg6F38p, 369Yv2NcqrygTpwX, KG62VXF3a97CxtAg,
  gtNvbhF9aZCXyzVk, zmTNyXkfqtpgcM1H, XvjpYan3qxF9WRgz, 7rGZbTJtpAqnBh2Y, xbJBZj14NWCXrhAm, MWrPDAVFCBXzh8Z6,
  VLz78BfZyWa3Cxr2, Dbfd6avYpzwTCx3N, jTX2kZ1K7gzt3dcL, nyMqbY64tL7HhKQJ, 8GNt2hC7v1rWRXyc, cD6JqWvBhk19xL3r,
  BRxGa38JnVhPKw4A, jXTVfy3ApWKPYwDG, jrTKDZcxNBvtdAP6, H2T7NDqBgm6LxGYK, LyjHDc4nCYmBMRNX, QRCgnhck3KVq6vBT,
  4Tc3yBNVYn7wW8RJ, daT4FcQgCYw2WBq6, RY1yA2PFCBwb4MZa, 6kN78G2xfqy3vwTP, MypmYhB3RX9FJ7Kx; all fights
- Game build: 1.60.1.70888 for the 7 reports from 2026-11-10 to 2026-11-12 (the 2026-11-10 one possibly 1.60.1.70862)
  and 1.60.1.70938 for the 48 from 2026-11-13 to 2026-11-15, across the hotfixes from 113890 to 114102; the hotfix
  each player ran is not recorded
- Date: 2026-11-10 to 2026-11-15
- Pull request: #1053

## Method

The reports hold 76 wearer streams, 73 runs once a run logged in two reports is counted once, and 14,699 Ancient
Power (1265634) damage events. A run counts from the first fight its gear shows the tablet in. A stretch of more than
5 minutes of combat with no proc is left out as the tablet swapped away (63 of 2,851 combat minutes; one run stops
proccing after 10 minutes with no gear change logged).

Ancient Power is a missile, and the log has no event for its launch: its damage lands about 100 ms after a melee hit
and 300 to 800 ms after an Auto Shot lands. So a hit counts only when it is isolated, with no other damage event of the
wearer and no Ancient Power within 400 ms of it (melee) or 1,200 ms (ranged and spells), and it procced when Ancient
Power lands inside that window after it. Which hit a proc belongs to then has one answer, and a hit's timing does not
depend on its roll. Dual wielders' white swings are split by hand on their damage before crits. Druid melee is taken at
the form's speed, 1.0 s in Cat Form and 2.5 s in Bear Form, told apart by the white swing interval.

Only the wearer's own cast hits on their first target count as spells. Left out are an area spell's pulses (Volley,
Blizzard, Consecration, Rain of Fire, Arcane Explosion, Chain Lightning and the like), a second target of the same hit
within 10 ms, a hit with no cast event of its own, and Windfury, Seal and poison hits. An area spell's chance per target
is unmeasured.

Each rate carries two 95% intervals: one from the hits (Wilson for a chance, profile likelihood for a rate per minute)
and one from 2,000 draws of whole wearers, which also carries how much wearers differ. The margin is the wider half.
With no procs the wearer draws are all zero, so the bound from the hits is the one that holds.

## Rates

| Hits                                          | Procs / isolated hits | Wearers | Rate     | 95%, hits    | 95%, wearers | Margin    |
| --------------------------------------------- | --------------------- | ------- | -------- | ------------ | ------------ | --------- |
| Main hand, white swings and specials          | 1782 / 14061          | 45      | 2.71 PPM | 2.59 to 2.82 | 2.58 to 2.82 | ±0.13     |
| Main-hand white swings                        | 1135 / 8766           | 45      | 2.77 PPM | 2.62 to 2.92 | 2.63 to 2.91 | ±0.15     |
| Melee specials                                | 647 / 5295            | 40      | 2.60 PPM | 2.42 to 2.80 | 2.32 to 2.86 | ±0.28     |
| Off-hand white swings                         | 238 / 2538            | 12      | 2.94 PPM | 2.60 to 3.31 | 2.45 to 3.44 | ±0.50     |
| Druid forms, on the form's speed              | 206 / 3471            | 6       | 2.83 PPM | 2.47 to 3.22 | 2.45 to 3.53 | ±0.70     |
| Auto Shot, on the ranged weapon's speed       | 392 / 3265            | 14      | 2.51 PPM | 2.28 to 2.75 | 2.34 to 2.73 | ±0.24     |
| Ranged specials, on the ranged weapon's speed | 39 / 412              | 20      | 2.00 PPM | 1.46 to 2.65 | 1.35 to 3.45 | ±1.46     |
| Spells, direct and landed, as a flat chance   | 138 / 2023            | 16      | 6.82%    | 5.80 to 8.00 | 5.92 to 8.03 | ±1.21 pts |
| Wand Shoot, as a flat chance                  | 38 / 426              | 7       | 8.92%    | 6.57 to 12.0 | 7.09 to 11.1 | ±3.09 pts |
| Periodic ticks                                | 0 / 1307              | 36      | 0%       | 0 to 0.29    |              | +0.29 pts |
| Damage shields (Thorns, auras, shield spikes) | 0 / 379               | 29      | 0%       | 0 to 1.00    |              | +1.00 pts |

On the main hand the chance per hit rises with the weapon's speed, from 5.3% at 1.0 s to 15.3% at 3.5 s, and a rate
per minute fits the hits far better than one flat chance (log-likelihood 91 higher for white swings, 35 for
specials). Spell hits show no such rise: 5.5% at a 1.5 s main hand, 8.4% at 3.0 s, 6.1% at 3.5 s, and a rate per
minute on the main hand's speed fits them worse than a flat chance (1.51 PPM, log-likelihood 3.4 lower). Wand Shoot
spans too few speeds to tell the two apart.

## Auto Shot and haste

Hunters shot faster than their bow's speed, three in four Auto Shots 10 to 50% faster. Over 2,415 Auto Shots fired in an unbroken string (each 0.5 to
1.15 times the bow's speed after the one before), a rate on the bow's speed (2.54 PPM, 95% 2.27 to 2.82), a rate on the
hasted interval (3.06 PPM, 95% 2.74 to 3.40) and a flat 12.1% fit equally well, within 0.2 of each other in
log-likelihood. Most wearers shoot a 2.8 s bow, so the speeds span too little to tell them apart.

| Bow speed over interval | Shots | Procs | Chance | 95%          |
| ----------------------- | ----- | ----- | ------ | ------------ |
| below 1.10              | 405   | 61    | 15.06% | 11.9 to 18.9 |
| 1.10 to 1.25            | 1023  | 125   | 12.22% | 10.4 to 14.4 |
| 1.25 to 1.50            | 792   | 81    | 10.23% | 8.3 to 12.5  |
| 1.50 and above          | 195   | 26    | 13.33% | 9.3 to 18.8  |

## Damage

Ancient Power crits at the wearer's own rate, 12.4% over all procs and from 5.5% for warriors to 16.2% for druids, and
2.7% of procs miss or are resisted. The median hit per class runs from 70 to 81 against the row's 69 to 81, and a
mage's (77) sits inside the row's range as a warrior's (70) does, so spell power adds nothing.
