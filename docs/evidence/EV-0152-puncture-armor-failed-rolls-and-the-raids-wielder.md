# EV-0152: Puncture Armor's failed hit rolls, and the raid's wielder against the sim's own

- Claim: Puncture Armor 17315 (Bashguuder, Rivenspike) and 1293005 (Vibroblade) roll the melee hit table in game and
  fail on a miss, a dodge or a parry. In the sim a main-hand wielder's Puncture Armor strike never misses and is dodged
  about 6.5% of the time, and the raid's copy with that dodge chance and one more roll off a swing lands within 2.4%
  of it.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), 48 reports with a wearer of Bashguuder or Rivenspike
  and 29 reports with a player who sources 1293005; the sim on master 4122e3519a
- Game build: 1.60.1.70862 to 1.60.1.70959 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-08 to 2026-11-18
- Pull request: #1106

## The logs

An attempt of Puncture Armor is what EV-0145 counts as one: a debuff applied, a stack added, a refresh that is not a
stack event's own, or the spell's 0-amount damage event, which is how a missed, dodged or parried one is logged. The
weapon and its hand come from the wearer's latest gear snapshot. The reports are the ones EV-0145 reads, downloaded by
2026-11-18. Vibroblade's are every report in which a player sources 1293005, counted whatever the hand.

| Weapon                | Reports | Wearers | Attempts | Landed | Missed | Dodged | Parried | Immune | Failed |
| --------------------- | ------: | ------: | -------: | -----: | -----: | -----: | ------: | -----: | -----: |
| Bashguuder, main hand |      11 |      11 |    1,170 |  1,087 |     15 |     28 |      38 |      2 |   7.1% |
| Bashguuder, off hand  |      12 |      12 |      991 |    952 |      2 |     27 |      10 |      0 |   3.9% |
| Rivenspike, main hand |      26 |      28 |    3,209 |  3,058 |     10 |     99 |      36 |      6 |   4.7% |
| Rivenspike, off hand  |       2 |       2 |      199 |    190 |      4 |      4 |       1 |      0 |   4.5% |
| Vibroblade (1293005)  |      29 |      29 |      649 |    578 |     21 |     27 |      23 |      0 |  10.9% |

The four rows of 17315 come to 5,569 attempts in 48 reports, of which 282 fail, 5.1%. Another 129 attempts come from
players whose snapshot holds neither weapon and are left out; 11 of them fail.

Examples: a dodged one in `AdfMXDLW29k6pGy1`, fight 1, at 00:08.635; a parried one in the same fight at 23:44.392; a
missed one in `DW9YkdvyrjtMBQnf`, fight 1, at 23:47.100.

## The sim's wielder

EV-0147 took the sim's own main-hand wielder as the model of the raid's copy, and read it while the player's proc
landed with no roll of its own. The player's proc is a strike with its own hit roll, and one a main-hand white swing
cast rolls the weapon once more (R-0040). Measured the way EV-0147 did, on master 4122e3519a: 180 s, one level 63
target from behind, 5,000 iterations, the raid's and party's buffs on, the raid's armor debuffs off but Sunder Armor.
The warrior is the pre-BiS Fury set with the weapon in the main hand, the rogue the Combat talents on the Sinister
Strike set.

| Wielder              | Strikes a minute | Landed a minute | Missed | Dodged |
| -------------------- | ---------------: | --------------: | -----: | -----: |
| Warrior, Bashguuder  |             8.20 |            7.68 |      0 |   6.4% |
| Warrior, Rivenspike  |             5.30 |            4.96 |      0 |   6.5% |
| Rogue, Rivenspike    |             5.35 |            5.10 |      0 |   4.6% |
| Warrior, Annihilator |                  |            2.89 |        |        |

Annihilator's Armor Shatter 16928 is on the magic table and is no strike, so its row is the landings alone; EV-0147
read 2.94. The warrior's 6.4% and 6.5% are the hit table's dodge for a target three levels above an attacker with its
level's weapon skill. The warrior's Bashguuder strikes come to 85.4 rolls a minute at the weapon's chance: 82.9 landed
main-hand hits, where EV-0147 read 83.4, and 2.5 rolls off the landed strikes of its main-hand swings.

The raid's copy in landings a minute, as `rollWielderHits` gives them: the swings at the weapon's base speed and 48
more hits, each at PPM × speed / 60.

| Wielder             | Sim's wielder | No roll of its own |  Dodged 6.5% | Dodged 6.5%, one more roll off a swing |
| ------------------- | ------------: | -----------------: | -----------: | -------------------------------------: |
| Warrior, Bashguuder |          7.68 |       7.81 (+1.7%) | 7.30 (-4.9%) |                           7.57 (-1.4%) |
| Warrior, Rivenspike |          4.96 |       5.31 (+7.2%) | 4.97 (+0.2%) |                           5.07 (+2.4%) |
| Rogue, Rivenspike   |          5.10 |       5.31 (+4.1%) | 4.97 (-2.6%) |                           5.07 (-0.5%) |

The dodge chance alone puts Bashguuder further from the sim's wielder than no roll does. With the roll off a swing
added all three are within 2.4%. The rogue's strike is dodged less than the warrior's, and the raid's copy takes the
warrior's 6.5%.
