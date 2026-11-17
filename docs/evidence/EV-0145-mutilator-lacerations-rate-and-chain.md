# EV-0145: Mutilator's Lacerations rate, its own hit roll and its chain

- Claim: Mutilator 12469 rolls Lacerations 10370 on 19.9% of landed main-hand melee hits, 4.8 procs per minute on its
  2.5 s speed; Lacerations rolls the melee hit table itself, and 12.2% of them miss or are avoided; a Lacerations a
  white swing cast rolls the weapon once more at the same chance, and one cast by anything else rolls nothing; Fiery
  Weapon 803 rolls on a weapon proc's strike under the same rule.
- Sources: Warcraft Logs reports `AnPHcV64wg8DThzZ` (fights 1, 11, 12), `q62VRcw1m8fyHN4T` (fight 1),
  `Xa2MfbLKQCPYH6nr` (fights 1, 3, 5), `t38h6AJQWzRCHwZy` (fights 1, 3 to 6, 8 to 14), `BXhNKykZL7MDat8c` (fight 8),
  `nPLWN2VDTZMz7Qk6` (fight 1) and `nP4N7AqTvChFQ2GY` (fights 1, 3 to 10): three paladins, two warriors and two
  hunters in dungeons. For Fiery Weapon, the combat log of EV-0130: a Retribution paladin, levels 46 and 47
- Game build: 1.60.1.70888 for the four reports of 2026-11-11 and 2026-11-12, 1.60.1.70938 for the three of
  2026-11-14 and 2026-11-15; the hotfix each player ran is not recorded. The reports are dated against the build each
  region's version file named at the time: 70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11 02:19 UTC in the
  EU, 70938 from 2026-11-13 01:36 UTC in both. No hotfix record of build 70938 up to push 114140 names a row of
  10370 (spell, name, effect, misc, categories, aura options, levels) or the item's effect entry. The combat log ran
  1.60.1.70938 after the hotfix pushes EV-0130 names.
- Date: 2026-11-11 to 2026-11-15
- Pull request: #1088

## Method

One Lacerations attempt is a debuff application, a stack gained, a refresh with no stack event at its millisecond
(the debuff already at 5), or a 0-amount damage event of 10370, which is a Lacerations that missed, was dodged or was
parried. The seven wearers log 1,308 attempts.

An attempt is credited by time to what rolled it. A landed Lacerations logs within 2 ms of the hit that rolled it.
A failed one logs when that hit is cast, so for a hit that lands late (Judgement of Command, a seal's strike, the
Judgement debuff a white swing refreshes, a Windfury Totem strike) it sits 4 to 20 ms before the hit, and goes to the
next hit or debuff within 30 ms. All but one attempt has such an owner.

A hit counts when it is the wearer's only hit event on that target within 100 ms, no periodic tick falls there, it is
not a killing blow and a gear snapshot shows the axe. A white swing also needs no other debuff of the wearer at its
millisecond. A strike that never stands alone is counted per millisecond: the landed hits of one ability at one
millisecond with no other ability's hit or debuff there. One warrior's snapshots never show the axe, so that report
counts toward the attempts only.

## Chance per roll

| Roll                                                 | With an attempt | Rolls | Chance |
| ---------------------------------------------------- | --------------: | ----: | -----: |
| White swing, landed                                  |             259 | 1,299 |  19.9% |
| White swing, missed or avoided                       |               0 |   305 |     0% |
| Holy Strike                                          |              78 |   360 |  21.7% |
| Seal of Command's strike                             |              47 |   220 |  21.4% |
| Seal of Righteousness, one strike                    |              46 |   255 |  18.0% |
| Seal of Righteousness, three strikes                 |              22 |    48 |  45.8% |
| Windfury Totem's strike                              |              21 |   109 |  19.3% |
| Judgement of the Crusader refreshed by a white swing |              50 |   216 |  23.1% |
| Cleave                                               |              10 |    68 |  14.7% |

The white swing's 95% interval is 17.9% to 22.2%. The three wearers with 300 swings or more read 19.8%, 18.4% and
23.1%; the first ran build 70888 and the other two 70938. Every wearer has the same 2.5 s weapon, so the logs cannot
tell a flat chance from procs per minute. Three rolls at 19.9% give 48.6% for the three seal strikes. The Judgement
refresh reads 40 of 195 alone and 10 of 21 where Vindication goes on the target at the same millisecond.

Auto Shot (546), Volley (818), Multi-Shot (180), Arcane Shot (140), Hammer of Wrath (107), Thunder Clap (61),
Consecration (6,004 ticks) and Sweeping Strikes' copied hit (2 of 121) roll nothing. Rend's application rolls with no
damage event beside it.

Judgement of Command (122 of 344), Execute (14 of 36), Overpower (12 of 35) and Rend's application (33 of 103) read
above the white swing's chance. Judgement of Command and Execute show two attempts on one cast 12 times, which fits
two rolls a cast; the logs do not separate that from a chain.

## The hit roll of Lacerations

Of the 1,308 attempts 160 fail: 67 miss, 63 are dodged, 28 are parried and 2 meet an immune target.

## Chain

Attempts credited to one landed hit:

| Hit that rolled the first       | A first | A second | A third |
| ------------------------------- | ------: | -------: | ------: |
| White swing                     |     259 |       52 |       0 |
| Holy Strike, Cleave, Overpower  |     100 |        0 |       0 |
| Seal of Command's strike        |      47 |        0 |       0 |
| Seal of Righteousness, 1 strike |      46 |        0 |       0 |
| Windfury Totem's strike         |      21 |        0 |       0 |

A second follows a white swing's first 20.1% of the time (15.7% to 25.4%), the chance of the first roll. Two
independent rolls a swing that leave 19.9% of swings with an attempt would leave 14 of the 1,299 with two, and 52
have two. Of the 52 pairs 44 are two landed applications, 5 hold a failed one and 3 are two refreshes at 5 stacks.
The 5 mixed pairs fit a chain that only a landed Lacerations continues (6 expected) better than one a failed
Lacerations continues too (11 expected).

## Fiery Weapon on a weapon proc's strike

The combat log of EV-0130 has Fiery Weapon 803 on a two-handed mace whose proc is the bleed 1306816 and on a
two-handed sword whose proc is the strike 12731. Units and rolls are EV-0130's. A unit where one of the paladin's
Judgement debuffs is refreshed or Vindication goes on the target is left out, since a refresh rolls on its own. On
units with no line of the weapon's proc Fiery Weapon reads 34.3% a roll on the mace (370 of 1,080) and 36.8% on the
sword (176 of 478).

170 units hold one landed weapon proc and no failed one:

| Unit                | Units | Fiery Weapon strikes | Expected, no further roll | Expected, one more roll at the base chance | More strikes than rolls |
| ------------------- | ----: | -------------------: | ------------------------: | -----------------------------------------: | ----------------------: |
| With a white swing  |   119 |                  133 |                     122.4 |                                      164.1 |                       3 |
| A white swing alone |    14 |                    7 |                       4.9 |                                        9.7 |                       1 |
| No white swing      |    51 |                   25 |                      30.4 |                                       48.3 |                       0 |

The 14 are among the 119. The units with no white swing are a Holy Strike or a Judgement of Command, four of them
with a seal's or Windfury Totem's strike beside it.

In a unit with a white swing the weapon proc is the swing's with a probability of one over the unit's rolls, since
every roll has the same chance. Fitted by maximum likelihood on that split, a landed weapon proc a white swing cast
brings one more Fiery Weapon roll at 30% (9% to 53%), and one anything else cast brings one at 0% (0% to 10%). One more
roll at the base chance on every landed weapon proc fits worse by 26.5, as twice the log-likelihood, and one on a white
swing's proc alone fits within 0.2 of the best. EV-0130's 7.9% and 18.6% are the same units pooled over what cast the
proc.

Only 14 units are a white swing alone, so the 30% rests on the split of the mixed units. An oil's proc and the off hand
were not counted.
