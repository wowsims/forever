# EV-0126: Argent Avenger's and Argent Defender's proc rates

- Claim: Argent Avenger's 17352 rolls about 1 proc per minute on the wearer's landed weapon hits, specials included.
  Argent Defender's 17350 procs about 0.5 times a minute on the attacker's swing interval, rolled by each landed melee
  hit taken that deals damage outside its 20 s cooldown, and about twice as often when an Undead strikes. The owner
  rules 1 proc per minute for the sword, and 0.5 procs per minute on the attacker's swing interval for the shield, twice
  that when an Undead strikes.
- Sources: public reports on forever.warcraftlogs.com, the fights whose combatant info lists the item on the wearer.
  Argent Avenger: ZfbY6RgWjx9hmpGz fights 1, 4, 9, 12, 13; 1z4vRDqGdM2cVPXk fight 1; Lpr76Vd3DbKWZqnh fights 1, 3;
  QHXdBJRj2GPpC3n1 fight 1; QxGpKz7knDvjWmYb fight 31; VjBTMHXxFvbthkw2 fight 1; NXMZV9xjvWpB7GtP fight 116;
  9zxHaWbYT3gqDQvF fight 17; z6LMfrpbwhyVCAJK fight 16; RGnyYzMgqCNdVwkT fight 21, left out below. Argent Defender,
  where a fight with no gear record counts too: VjBTMHXxFvbthkw2 fight 1; NXMZV9xjvWpB7GtP fight 107; 7tLY1pVK9wRWPABX
  fights 15 to 18 and 20; MWrPDAVFCBXzh8Z6 fights 1 to 8.
- Game build: 1.60.1.70938. Which hotfix push each report ran on and whether players relogged is not recorded.
- Date: 2026-11-13 07:09 to 2026-11-15 09:41 UTC
- Pull request: #1052

## Argent Avenger

The game files state no rate: 17352 has no proc row, and the item's chance-on-hit effect names it. A proc is one
application or refresh of 17352 on the wearer. Each logs just before the hit that rolled it, so a proc goes to the
wearer's next damage event that is not a tick, within 5 ms (356 of 389 procs), and otherwise to a cast within 5 ms
(Sunder Armor, which logs no damage). Hits are landed hits. PPM = chance × 60 / 2.2, the sword's speed, and intervals
are 95% Wilson intervals. A dual wielder's white swings are split between the hands by damage, and the split matches
the share the two weapons' speeds and Sword Specialization's extra swings predict. One dual-wielding warrior whose
white swings could not be split is left out.

The cleanest wearer is a warrior with the sword and a shield, in two reports, 138 procs:

| Hits                                                             | Procs / landed hits | Chance | PPM at 2.2 s |
| ---------------------------------------------------------------- | ------------------- | ------ | ------------ |
| White swings                                                     | 54 / 1,261          | 4.3%   | 1.17         |
| Revenge                                                          | 17 / 467            | 3.6%   | 0.99         |
| Cleave, per target                                               | 23 / 837            | 2.7%   | 0.75         |
| Sunder Armor, casts                                              | 14 / 478            | 2.9%   | 0.80         |
| Heroic Strike                                                    | 2 / 36              | 5.6%   | –            |
| Shield Slam                                                      | 0 / 498             | 0%     | –            |
| Thunder Clap                                                     | 0 / 597             | 0%     | –            |
| Weapon hits: white, Revenge, Cleave, Sunder Armor, Heroic Strike | 110 / 3,079         | 3.6%   | 0.97         |

Sunder Armor counts casts, so its chance is a floor. Warrior off hands also procced on Whirlwind's off-hand hit (25
procs). Windfury Totem strikes roll it as EV-0119 records: a proc a strike rolled logs 6 to 20 ms before the strike's
damage, outside the 5 ms window above, and crediting within 20 ms gives 19 procs on 453 landed strikes across all
wearers, where 21.5 are expected at the white chance.

White swings alone, by class and hand:

| Wearers                     | Procs / landed white swings | PPM at 2.2 s [95%] |
| --------------------------- | --------------------------- | ------------------ |
| Warrior main hand, 1 wearer | 54 / 1,261                  | 1.17 [0.90, 1.51]  |
| Warrior off hand, 3 wearers | 57 / 1,342                  | 1.16 [0.90, 1.49]  |
| Rogue off hand, 4 wearers   | 113 / 4,470                 | 0.69 [0.57, 0.83]  |

The warrior off-hand wearers read 0.79 to 1.60 each and the rogue off-hand wearers 0.40 to 0.74. Haste does not
explain the gap between the classes: both swing about 1.2 to 1.3 times as often as the base speed, and the rate is
priced on the base speed. Why rogues proc less is open. The owner sets the rate to 1 proc per minute, from the
sword-and-shield warrior's weapon hits.

## Argent Defender

17350 states a ProcChance of 100, a ProcTypeMask of 0x28 (melee hits taken) and a 20 s ProcCategoryRecovery, and its
tooltip says "Has a chance when struck in combat"; its second effect, a dummy of 2, is the tooltip's "Undead that
strike you are 2 times as likely". The 100 in the column is not the chance.

A trial is a melee hit the wearer takes outside the cooldown that deals damage after block and absorb: a white swing,
a school-melee swing (logged as ability -64 or -16) or a melee special. Partial blocks count. A fight counts unless
its combatant info lists gear without the shield; a fight with no gear record counts. The cooldown runs from the hit
a proc is credited to.

Each 17351 application logs 0 to 33 ms after the hit that rolled it (median 9 ms, 8 at 0 ms), the opposite of the
Avenger's order. A proc goes to the last landed melee hit in the 60 ms before it, and all 86 have one. Seven windows
held hits from more than one attacker, two of them with a different speed or creature type; crediting the earliest
moves nothing past the third decimal. The shortest gap between two procs is 20.76 s, and 806 hits inside a cooldown
drew none.

| Wearer  | Dungeons                     | Procs | Trials | From Undead |
| ------- | ---------------------------- | ----- | ------ | ----------- |
| Paladin | Blackrock Spire              | 8     | 403    | 0           |
| Warrior | Blackrock Spire              | 9     | 280    | 0           |
| Paladin | Blackrock Spire, Scholomance | 46    | 2,420  | 710         |
| Paladin | Dire Maul                    | 23    | 1,655  | 234         |
| All     |                              | 86    | 4,758  | 944         |

What rolls, outside the cooldown:

| Hits taken                                     | Procs / hits | Rate |
| ---------------------------------------------- | ------------ | ---- |
| White swings, plain                            | 37 / 1,662   | 2.2% |
| White swings, partly absorbed                  | 17 / 802     | 2.1% |
| White swings, partly blocked                   | 27 / 2,051   | 1.3% |
| White swings, partly resisted or crits         | 0 / 18       | 0    |
| Melee specials that deal damage                | 5 / 273      | 1.8% |
| Melee hits blocked or absorbed whole           | 0 / 567      | 0    |
| Melee misses, dodges, parries and immunes      | 0 / 2,623    | 0    |
| Spells, ranged hits and other non-melee damage | 0 / 985      | 0    |

Counted as trials, the 554 hits blocked or absorbed whole that the models see would draw 9.6 procs (P ≈ 1e-4).
Partial blocks proc at about 0.6 times the rate of unblocked hits (Mantel-Haenszel odds ratio 0.57 within wearer and
attacker, p ≈ 0.03), for no reason found; they stay in the trials.

### Attackers

Creature types come from Wowhead's classic NPC data for 191 NPCs, since the Forever pages could not be read; every
trial NPC has the same type in all three data environments read. An attacker's speed is its time between its own
swings at the wearer, measured from the logs: the mode of the gaps between its swings, refined to the median of the
gaps within 12% of it, with swings 100 ms apart merged as extra attacks. 81 of the 106 NPCs with a clean speed swing
every 2.0 s. Swings within 30 s of the warrior's Thunder Clap slow from 2.0 to 2.4 s and are left out of the speeds.
16 NPCs dual-wield, told apart by hits that alternate high and low and by gap pairs that sum to one hand's speed.
The 48 trials whose attacker has no speed, one proc among them, are left out of the wearer table's trials and every
fit; the wearer table's procs and the table of what rolls count them.

### Fits

Each model is fitted by maximum likelihood over the trials: chance = c × speed^β × k^u, with u marking an Undead
attacker. Flat is β = 0 and procs per minute β = 1, with PPM = 60c. Intervals are profile-likelihood 95% intervals.
Speed is per hand here; a dual wielder's per-swing gap is below.

| Model          | logL    | AIC    | Estimate [95%]                            |
| -------------- | ------- | ------ | ----------------------------------------- |
| Flat, k = 2    | -426.50 | 855.00 | 1.49% [1.20, 1.81]                        |
| PPM, k = 2     | -425.36 | 852.73 | 0.476 [0.384, 0.580]                      |
| Flat, k free   | -425.43 | 854.85 | 1.65% [1.28, 2.07], k 1.41 [0.87, 2.23]   |
| PPM, k free    | -424.23 | 852.46 | 0.530 [0.411, 0.664], k 1.40 [0.85, 2.17] |
| Speed^β, k = 2 | -425.31 | 854.62 | β 0.82 [-0.20, 1.94]                      |

The logs barely tell flat from procs per minute: 68% of the trials come from 2.0 s attackers, where both give the same
chance. On simulated data with the same attackers, the likelihood picks the true model 83 to 84% of the time.

| Trials                                | Trials / procs | Flat [95%]         | PPM [95%]         |
| ------------------------------------- | -------------- | ------------------ | ----------------- |
| All                                   | 4,758 / 85     | 1.49% [1.20, 1.81] | 0.48 [0.38, 0.58] |
| Single-weapon attackers               | 4,069 / 83     | 1.74% [1.40, 2.12] | 0.55 [0.44, 0.67] |
| One attacker on the wearer in the 5 s | 1,418 / 39     | 2.22% [1.61, 2.97] | 0.69 [0.50, 0.91] |
| Unique NPCs                           | 469 / 15       | 2.47% [1.44, 3.89] | 0.81 [0.47, 1.27] |
| Trash, two or more attackers          | 3,244 / 45     | 1.18% [0.87, 1.55] | 0.38 [0.28, 0.50] |
| Dual wielders at their per-swing gap  | 4,758 / 85     | 1.49%              | 0.52 [0.42, 0.63] |

The Undead factor is 1.75 [1.06, 2.72] on single-weapon attackers and 1.92 [1.14, 3.15] on those with a clean speed,
in line with the row's 2. Over all attackers it reads 1.41 [0.87, 2.23], pulled down by two dual-wielding Undead that
procced none of 249 hits.

Expected procs under each fitted model, k = 2:

| Attacker speed | Attacker | Mean speed | Trials | Procs | Flat | PPM  |
| -------------- | -------- | ---------- | ------ | ----- | ---- | ---- |
| Under 1.3 s    | Other    | 1.11       | 382    | 5     | 5.7  | 3.4  |
| Under 1.3 s    | Undead   | 1.01       | 77     | 0     | 2.3  | 1.2  |
| 1.3 to 1.7 s   | Other    | 1.49       | 639    | 6     | 9.5  | 7.6  |
| 1.3 to 1.7 s   | Undead   | 1.37       | 61     | 4     | 1.8  | 1.3  |
| 1.7 to 2.2 s   | Other    | 2.00       | 2,467  | 42    | 36.7 | 39.1 |
| 1.7 to 2.2 s   | Undead   | 2.00       | 788    | 17    | 23.5 | 25.0 |
| 2.2 to 2.7 s   | Other    | 2.41       | 290    | 10    | 4.3  | 5.5  |
| 2.2 to 2.7 s   | Undead   | 2.40       | 18     | 1     | 0.5  | 0.7  |
| 2.7 s and over | Other    | 3.61       | 36     | 0     | 0.5  | 1.0  |

### What neither model explains

- Dual-wielding attackers drew 2 procs in 689 trials. Flat expects 14.0 (P ≤ 1e-4), procs per minute on each hand's
  speed 12.6 (P = 3e-4), and procs per minute on the attacker's per-swing gap 6.8 (P = 0.03). The last reading fits best
  of the speed models (β = 1.43 [0.62, 2.40], logL 6.4 above flat), so the speed that matters is the attacker's time
  between its own swings. The shortfall holds at every damage size and every density of incoming hits.
- The chance per hit falls as more attackers hit the wearer: 2.75% with one attacker in the 5 s before the hit, 1.54%
  with two, 1.32% with three and 1.17% with four or more. As n^η it fits η = -0.50 [-0.90, -0.12] with speed in the
  model, and -0.35 [-0.78, 0.08] within trash. The time since the wearer's previous hit hardly matters (δ = 0.15 [0.00,
  0.30]), which rules out a cap on the wearer's procs per minute. The cause is open.
- Per-wearer rates against one pooled rate give p ≈ 0.11 (flat) and 0.22 (procs per minute). The wearers' own weapon
  speeds run against their rates, which rules out the wearer's swing speed.

The owner rules 0.5 procs per minute on the attacker's swing interval, from the pooled fit, with the Undead factor of
2 the row states and the 20 s cooldown.
