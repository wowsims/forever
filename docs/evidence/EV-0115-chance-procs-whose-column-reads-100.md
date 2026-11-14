# EV-0115: Draconic Infused Emblem, Zealot Plate and Moonshadow Stave proc rates

- Claim: Draconic Infused Emblem 22268 procs 8% on each landed harmful spell hit, Zealot Plate 274651 procs 5% on
  each landed melee swing taken outside its 15 s cooldown, and Moonshadow Stave 22458 procs 1.5 times a minute on the
  cast time of each landed harmful spell, at least 1.5 s.
- Sources: Forever dungeon reports on Warcraft Logs, listed per item below with the fights each wearer had the item on
- Game build: 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938, by report date (counts per item below); hotfix state not
  recorded. No hotfix changed 1318930, 1318931, 1294666, 1294668, 1300435, 1300437 or 1300438 on these builds.
- Date: 2026-11-09 to 2026-11-14
- Pull request: #1016

Wearers come from a search of 3,226 zone 3000 reports for each proc's spell ids. Combatant info gives the gear, once per
dungeon run; a wearer's fights run from the first combatant info that shows the item to the first that does not. The
proc column of every trigger reads 100, a sentinel the tooltips ("Chance on harmful spellcast", "When struck in combat,
has a chance", "Harmful spell casts have a chance") contradict.

## Draconic Infused Emblem

Reports: 1ZyhaVDdFtTYWMcC/1-12 (two mages), NRK61HZcDdJkrwXG/7-9 (warlock), NXMZV9xjvWpB7GtP/13 (warlock),
T79gRkJz4QVGbyYW/14-15 (priest). RXzv1k3V4PgaWhqj is a second upload of the T79gRkJz4QVGbyYW run and is left out.
Builds: 70888 (2 reports), 70938 (2).

A proc counts each apply or refresh of 1318930. The trigger's 0x10000 rolls once for each target a harmful spell lands
on: mages proc 17 to 20% a cast from Arcane Explosion and Flamestrike and 8% a hit, like the other classes.

| Wearer  | Procs | Counted against                   |  Rate |
| ------- | ----: | --------------------------------- | ----: |
| Mage    |   231 | 2,725 direct hits                 |  8.5% |
| Mage    |   135 | 1,682 direct hits                 |  8.0% |
| Warlock |    20 | 188 hits                          | 10.6% |
| Warlock |    12 | 148 hits and debuff applications  |  8.1% |
| Priest  |    37 | 491 harmful casts (single target) |  7.5% |
| Pooled  |   435 | about 5,200                       |  8.3% |

The pooled interval is ±0.75%. Per ability, hardcasts roll like instants: Frostbolt 9.1% (22 of 241), Arcane Blast 5.3%
(12 of 228), Pyroblast 4.7% (4 of 85), Arcane Explosion 7.9% (164 of 2,079), Arcane Missiles missiles 7.5% (37 of 493).
Procs a minute differ threefold by class (mages 2.1 to 2.4, the priest 1.1, a warlock 0.8), so the rate is no procs per
minute. Mages show two procs in one millisecond on area hits, so there is no cooldown. Hellfire Effect never procs it
(0 of 1,991 hits); Bane of Havoc copies proc it 4 times in 558. Misses are undetermined.

## Zealot Plate

Reports, 40 paladins: 2HMyR6kfbWKzBhtv/1, 2bdPF6mYxvJgZXK7/1-5, 369Yv2NcqrygTpwX/1-5, 3LwjMABn6RCd1Kz8/1-2,
4YzXaGtg81WdqM2N/1-15 (two wearers), 7vacy8gRZM9CHm16/1-11, 8vJjYft73m9Chzkc/14, CF7ztDwY8pJ9KZhn/1-12,
CGF4WpJPjbhvKg9y/1-2, JFqRThp6Z4HrnBwf/1, MC2GmHchpfVkNqwD/14-47, NXMZV9xjvWpB7GtP/10-99, NgRFKH1PWzykqC9Q/12-15,
PB6zGcVNtTZrMKXb/1-6, RPFnkAX7pH6qQ2Mm/7, VCaNv2TFXx6y1dYZ/1-9, VDCA7xJ8GyNfrtcW/1, VyXwH1Wfq2bc3xZ6/1-21,
WrBf9NpqGhTcm21x/1, XxqdrnVN9DCZcL8T/1-9, bP6qxJTBpdHf47MA/1-4, fAThJd7KGmX3Pkcg/1-2, fXyK8kCNdGxtQ67c/1-2,
g6JdFkQKfR7YyP9B/1-3, gBDMLv8J1K7nqkpd/10-70 (two wearers), gJtby1TP3nmA2DFp/1-29, hR6pzT1HkaLym2nd/1-2,
rBWb9Nyt4YpPFGH7/1-23, rDJL7nqX1VZHm9zk/1-8, rtQwP64YL7GqR1Hf/1-12, tY2jp7cVGaQx8nDw/1-9, wJtPdpxFGQCM9f3c/1-2,
wa9fNCxj8JZLMDcq/5-7, x378BTAPqhCJryVW/2-12, xWKHB8qbYAcgJVRn/1-6, xbfy4DQZ6PWCFXYr/1-10, ymcTt8vprBnJFC9H/1-23,
zptFCDJMLT7QGxvK/1-3. Builds: 70862 (5 reports), 70888 (8), 70938 (25).

A proc counts each mana gain of 1294666. The denominator is landed auto swings taken (hit, crit, block, glancing,
crushing) at least 15 s after the wearer's last proc; mob melee specials taken are left out of it, while their procs
stay in, so the rate is a little high.

| Procs | Landed melee swings taken outside the cooldown | All landed melee swings taken | Rate        |
| ----: | ---------------------------------------------: | ----------------------------: | ----------- |
| 1,661 |                                         31,232 |                        51,802 | 5.3% ± 0.2% |

The shortest gap between two procs of one wearer is 14.97 s, so the row's 15 s cooldown holds.

## Moonshadow Stave

Reports, 10 druids: 24kgnZAXbKTW89x6/1-2, 7cpTkjaFB3mz6VgM/272-492, VqZ38vXydB7gn1J9/1-5, WKgT9mPJHnYa14Zy/1,
cvRTFb1qhGBKNkCD/1-24, f36RaB12wyDcTCGY/1-3, hCfYgdqVk7QDRZGN/1-3, rPpHYfq3tNnwJRMd/5-19, x3njCmpkVYf7WDgh/1-39,
z1kHPnyGtX7g4QqC/3. Builds: 70888 (4 reports), 70938 (6).

A proc counts each hit of 1300437 (Shadow) or 1300438 (Arcane), 359 in all: 165 Shadow and 194 Arcane, an even split
within chance (z = −1.5). Both are projectiles, so a proc lands one
flight after the spell that rolled it; each is matched to its spell by target and Wrath's flight time to that target.
The logs carry no positions. Two wearer windows end in long stretches with no proc (959 and 487 casts) and are cut at
their last proc.

The chance is procs per minute times the cast time of the launching cast, at least 1.5 s, divided by 60, rolled when a
harmful spell lands on a target. The fit is 1.52 per minute (bootstrap 1.27 to 1.75; profile likelihood 1.36 to 1.69):
3.8% on an instant or a 1.5 s Wrath, 6 to 9% on Starfire.

| Model                                      | Log-likelihood |
| ------------------------------------------ | -------------: |
| Procs per minute on cast time, per landing |        −3881.3 |
| Flat 4.05% per landing                     |        −3887.7 |
| Procs per minute per cast, misses rolling  |        −3886.6 |
| Procs per minute, Hurricane ticks rolling  |          −4226 |
| Hurricane cast rolling once                |        −3932.6 |

| Spell            | Procs seen | Expected at 1.52 per minute |
| ---------------- | ---------: | --------------------------: |
| Wrath            |       97.7 |                       105.7 |
| Starfire         |       45.9 |                        45.4 |
| Moonfire         |       38.0 |                        34.0 |
| Insect Swarm     |       59.4 |                        46.2 |
| Entangling Roots |        6.2 |                         7.0 |
| Faerie Fire      |        3.8 |                         9.0 |
| Hurricane cast   |       85.4 |                        89.8 |

Nothing else rolls it: Hurricane ticks 0 of 16,615, damage over time ticks 0.7 of 10,625, misses and resists 0 of 159,
helpful casts 0 of 3,103, Thorns and melee 0 of 1,122. A Hurricane cast procs it on 85 of 591 casts and never twice, so
it rolls once for each target until one procs. Starfire carries the cast-time signal alone (about 46 procs from three
wearers), so a flat chance with Starfire as an exception fits nearly as well. Without the window cut the fit is 1.26;
without the two lowest wearers, 1.56. 18 procs stay unmatched.
