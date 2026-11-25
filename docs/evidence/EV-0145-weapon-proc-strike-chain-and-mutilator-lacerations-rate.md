# EV-0145: A weapon proc's strike and what it rolls, and Mutilator's Lacerations rate and hit roll

EV-0156 dates the reading that a white swing's Judgement refresh rolls a weapon proc on its own: it holds to
2026-11-13, and since that day's hotfix the refresh rolls nothing.

- Claim: Mutilator 12469 rolls Lacerations 10370 on 19.9% of landed main-hand white swings, 4.8 procs per minute on
  its 2.5 s speed; Lacerations rolls the melee hit table itself, and 12.2% of them fail; a weapon proc's strike that a
  main-hand white swing of the weapon cast rolls the weapon once more at the same chance and never a third time, on 23
  weapons of the table, Mutilator among them; one cast by an ability, a seal's strike, a Windfury hit, an off-hand swing
  or a druid's swing in Cat, Bear or Dire Bear Form rolls nothing; Toofbreak 1320269, whose row states a 100 ms
  cooldown, brings no second; Judgement of Command rolled a weapon proc twice a cast up to hotfix push 114399 (2026-11-25) and rolls it once from
  that push (EV-0155); Fiery Weapon 803 rolls on a
  weapon proc's strike under the same rule.
- Sources: Warcraft Logs reports `AnPHcV64wg8DThzZ` (fights 1, 11, 12), `q62VRcw1m8fyHN4T` (fight 1),
  `Xa2MfbLKQCPYH6nr` (fights 1, 3, 5), `t38h6AJQWzRCHwZy` (fights 1, 3 to 6, 8 to 14), `BXhNKykZL7MDat8c` (fight 8),
  `nPLWN2VDTZMz7Qk6` (fight 1) and `nP4N7AqTvChFQ2GY` (fights 1, 3 to 10): three paladins, two warriors and two
  hunters in dungeons. For Fiery Weapon, the combat log of EV-0130: a Retribution paladin, levels 46 and 47. For the
  chain on the other weapons, the 2,861 reports of forever.warcraftlogs.com zone 3000 from 2026-11-06 to 2026-11-18
  in which a player procs one of the weapons that section names; the section says how they are chosen and names
  example units
- Game build: 1.60.1.70888 for the four reports of 2026-11-11 and 2026-11-12, 1.60.1.70938 for the three of
  2026-11-14 and 2026-11-15; the hotfix each player ran is not recorded. The reports are dated against the build each
  region's version file named at the time: 70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11 02:19 UTC in the
  EU, 70938 from 2026-11-13 01:36 UTC in both. No hotfix record of build 70938 up to push 114140 names a row of
  10370 (spell, name, effect, misc, categories, aura options, levels) or the item's effect entry. The combat log ran
  1.60.1.70938 after the hotfix pushes EV-0130 names. The 2,861 reports run from 1.60.1.70862 or older to
  1.60.1.70959, split by build in their section.
- Date: 2026-11-06 to 2026-11-18
- Pull request: #1088

## Method

One Lacerations attempt is a debuff application, a stack gained, a refresh with no stack event at its millisecond
(the debuff already at 5), or a 0-amount damage event of 10370, which is a Lacerations that missed, was dodged or was
parried. The seven wearers log 1,308 attempts.

An attempt is credited by time to what rolled it. A landed Lacerations logs within 2 ms of the hit that rolled it.
A failed one logs when that hit is cast, so for a hit that lands late (Judgement of Command, a seal's strike, the
Judgement debuff a white swing refreshes, a Windfury Totem hit) it sits 4 to 20 ms before the hit, and goes to the
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
| Windfury Totem hit                                   |              21 |   109 |  19.3% |
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
above the white swing's chance. Judgement of Command and Execute show two attempts on one cast 12 times, which is a
second roll and no chain (the section on the other weapons). Raptor Strike reads below it here (0 of 21, at most
15.5%) and rolls on the other weapons.

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
| Windfury Totem hit              |      21 |        0 |       0 |

A second follows a white swing's first 20.1% of the time (15.7% to 25.4%), the chance of the first roll. Two
independent rolls a swing that leave 19.9% of swings with an attempt would leave 14 of the 1,299 with two, and 52
have two. Of the 52 pairs 44 are two landed applications, 5 hold a failed one and 3 are two refreshes at 5 stacks.
The 5 mixed pairs fit a chain that only a landed Lacerations continues (6 expected) better than one a failed
Lacerations continues too (11 expected).

## The chain on the other weapons

The reports are every zone 3000 report downloaded by 2026-11-18 in which a gear snapshot holds one of 27 weapons in
the main or off hand, or a player sources one of their proc spells. Of the 27, 25 are in the table, and Blood Talon
12795 and The Needler 13060 have no wearer. Killmaim 13016 and Boneslicer 273819 are not among the 27: they carry the
proc spells of Blood Talon and Ripsaw, and are in the table through them. An attempt of a proc is its damage event
where the row has a damage effect, and otherwise what an attempt of Lacerations is, with a damage event on an immune
target counted as a failed one. A debuff or a hit of a second spell with the proc's own name, such as the stun beside
Farraki Enforcement, is part of the proc.

The wearer's hits and debuffs on one target form a unit while the gap between two stays within 100 ms. A lone white
swing is a unit whose only hit is one landed white swing, with no other debuff of the wearer. A lone ability is the
same for one landed melee ability, its own debuff allowed. The weapon and its hand come from the latest gear snapshot,
and a run two reports hold counts once.

Lone main-hand white swings whose first attempt landed, druids left out, and lone abilities other than Judgement of
Command and Execute whose first attempt landed. The first column of numbers is the share of lone white swings with an
attempt, for wearers with no weapon in the off hand:

| Weapon                             | Proc    | A first, per swing | White swing, landed first | With a second | With a third | Ability, landed first | With a second |
| ---------------------------------- | ------- | -----------------: | ------------------------: | ------------: | -----------: | --------------------: | ------------: |
| Bigger Rock on Bigger Stick 279480 | 1306816 |              10.2% |                    17,079 | 1,829 (10.7%) |            0 |                 9,261 |             3 |
| Rock on Stick 279852               | 1309967 |               8.4% |                     8,314 |    609 (7.3%) |            0 |                 8,037 |             0 |
| Toofbreaker 286430                 | 1320269 |               6.0% |                     6,329 |             0 |            0 |                 3,684 |             0 |
| Ripsaw 9478                        | 1292573 |               8.0% |                     4,888 |    378 (7.7%) |            0 |                 5,180 |             0 |
| Princess Theradras' Scepter 17766  | 21961   |              10.6% |                     2,514 |   268 (10.7%) |            0 |                 1,597 |             0 |
| Stoneslayer 9418                   | 12731   |              11.6% |                     2,488 |   293 (11.8%) |            0 |                 2,848 |             0 |
| Gatorbite Axe 17730                | 21949   |               9.4% |                     1,553 |    122 (7.9%) |            0 |                   868 |             0 |
| The Chief's Enforcer 9477          | 1294339 |               5.4% |                     1,335 |     62 (4.6%) |            0 |                   623 |             0 |
| Boneslicer 273819                  | 1292573 |               8.6% |                       705 |     65 (9.2%) |            0 |                   310 |             0 |
| Killmaim 13016                     | 13318   |               4.3% |                       457 |     14 (3.1%) |            0 |                   210 |             0 |
| Rivenspike 13286                   | 17315   |                    |                       356 |     31 (8.7%) |            0 |                   454 |             0 |
| Bashguuder 13204                   | 17315   |               7.6% |                       353 |     26 (7.4%) |            0 |                   264 |             0 |
| Glutton's Cleaver 10772            | 18075   |               6.3% |                       332 |     21 (6.3%) |            0 |                   154 |             0 |
| Pendulum of Doom 9425              | 10373   |               6.3% |                       301 |     20 (6.6%) |            0 |                   286 |             0 |
| Hammer of Expertise 8708           | 1317432 |               6.5% |                       272 |     23 (8.5%) |            0 |                   183 |             0 |
| Kang the Decapitator 2291          | 17153   |               5.3% |                       271 |     19 (7.0%) |            0 |                   110 |             0 |
| The Blackrock Slicer 13285         | 17407   |              12.9% |                       229 |    24 (10.5%) |            0 |                   145 |             0 |
| Gut Ripper 2164                    | 18107   |                    |                       225 |      9 (4.0%) |            0 |                   141 |             0 |
| Mutilator 12469                    | 10370   |              19.1% |                       197 |    45 (22.8%) |            0 |                   105 |             0 |
| Axe of Inevitable Ruin 274484      | 1294226 |               9.5% |                       182 |     16 (8.8%) |            0 |                    75 |             0 |
| Bloodrazor 809                     | 17504   |               5.2% |                       165 |      9 (5.5%) |            0 |                   180 |             0 |
| Grim Reaper 13054                  | 14126   |               5.0% |                       106 |      3 (2.8%) |            0 |                    60 |             0 |
| Drakefang Butcher 12463            | 14118   |               8.1% |                        89 |      5 (5.6%) |            0 |                    49 |             0 |
| Brain Hacker 1263                  | 17148   |               4.8% |                        61 |      4 (6.6%) |            0 |                    35 |             0 |
| Hanzo Sword 8190                   | 16405   |               2.1% |                        16 |             0 |            0 |                     5 |             0 |
| Bloodfist 11744                    | 16433   |                    |                         8 |             0 |            0 |                    15 |             0 |
| Hameya's Slayer 15814              | 16406   |                    |                         1 |             0 |            0 |                     0 |             0 |

Hanzo Sword, Bloodfist and Hameya's Slayer are worn in the off hand almost only.

Example units, by report, fight and time into the fight:

| Unit                                                             | Report             | Fight | Time      |
| ---------------------------------------------------------------- | ------------------ | ----: | --------- |
| A white swing with two Strength of Stone                         | `2zXhygVYpkrMZGxT` |     1 | 5:26.266  |
| Three attempts: two at a white swing, one at a Judgement refresh | `6T3M1HRAxyntv9DL` |     1 | 4:23.811  |
| A druid's white swing in a form with one Bigger Rock to the Face | `1WHqx7bPa8mkGydJ` |     1 | 3:37.180  |
| A white swing with one Toofbreak                                 | `1BNXynxVbH4kGt8R` |     1 | 0:44.852  |
| One of the four off-hand white swings with two Rock to the Face  | `7xNZAm1c9gy4nVXB` |     1 | 90:30.259 |

### No third

With Toofbreaker left out, 42,497 landed firsts bring 3,895 seconds, 9.2%, which is about each weapon's chance of
a first, and no third. A chain that went on at the same chance would have left about 378 thirds. Bigger Rock on
Bigger Stick, Rock on Stick, Stoneslayer, Ripsaw, Princess Theradras' Scepter, Mutilator and Gatorbite Axe would each
have shown about 10 or more; the other weapons rest on the pooled count.

1,110 units hold three attempts or more, and each holds as many rolls beside them, with Judgement of Command counted
as two. The 13 with one white swing, no ability and three attempts show the third roll: a Judgement debuff the swing
refreshes 26 to 39 ms later in 11, a Rend application in one and a Cheap Shot in one.

By the build each region's version file named when the report started:

| Report started (UTC)                                            | White swing, landed first | With a second | Druid, landed first | With a second | Toofbreaker | With a second |
| --------------------------------------------------------------- | ------------------------: | ------------: | ------------------: | ------------: | ----------: | ------------: |
| before 2026-11-10 15:11, 1.60.1.70862 or older                  |                     1,737 |    158 (9.1%) |               2,164 |             1 |         122 |             0 |
| to 2026-11-13 01:36, 70888 (the EU from 2026-11-11 02:19)       |                     7,155 |    654 (9.1%) |               4,360 |             1 |         689 |             0 |
| to 2026-11-17 15:19, 70938                                      |                    29,459 |  2,701 (9.2%) |              11,424 |             1 |       4,885 |             0 |
| after it, 70959 in the US and, from 2026-11-18 02:04, in the EU |                     4,146 |    382 (9.2%) |               1,652 |             0 |         633 |             0 |

The druid columns are every druid's lone white swing, whatever form the tracking below gives it. The version file's
listing starts on 2026-11-07 with 1.60.1.70862.

### Abilities and triggered hits

With Toofbreaker left out as above, a lone ability's landed first brings 3 seconds in 31,195, over 28 abilities. The
largest by units with an attempt are
Stormstrike (8,420), Holy Strike (6,729), Sinister Strike (6,073), Cleave (5,966) and Revenge (3,873). Raptor Strike
rolls: 409 of 5,306 lone hits have an attempt.

A triggered hit that lands after the swing that cast it is read by millisecond: the landed attempts within 2 ms of a
landed triggered hit with no other hit or debuff of the wearer within 3 ms of it.

| Triggered hit                  | One landed attempt | Two |
| ------------------------------ | -----------------: | --: |
| Seal of Command's strike 20424 |              3,021 |   0 |
| Seal of Righteousness, strike  |              3,449 |   1 |
| Windfury Totem hit             |              8,058 |   1 |

Windfury Weapon's two swings land together, so they are read off the units of one landed Stormstrike and two landed
swings. Of the 3,147 with an attempt, 2,797 hold one, 339 two and 11 three. Three rolls with no chain give 2,816, 318
and 12; a chain off the two swings gives 2,655, 444 and 46, with 3 at four.

### Judgement of Command and Execute

A lone Judgement of Command has an attempt on 7,560 of 38,665 casts and two on 453, priced per weapon and class on the
white swing's chance. One roll a cast gives 3,784 with an attempt; two rolls give 7,184, and 383 with two. A lone
Holy Strike on the same weapons has 6,440 of 64,800 against 5,740 for one roll, with 2 holding two.

Execute sits between: 917 of 6,984 with an attempt and 46 with two, against 699 for one roll and 1,325 and 72 for
two. A missed Execute with nothing else in its unit has an attempt beside it 122 times, so Execute has a second roll
that does not need its damage to land. Its chance is not settled here.

### A druid's form

A druid's lone white swing in a form brings no second: 0 of 8,616 landed firsts in Cat Form, 0 of 8,364 in Dire Bear
Form and 0 of 1,256 in Bear Form. A lone swing has an attempt 3.3% of the time in Cat Form, 7.3% in Dire Bear Form
and 10.2% in Bear Form, and a second at those chances would make about 1,000. The form is
the druid's latest form buff in the log. Out of a form the tracking is not reliable enough to read: it places 742
swings in Travel Form.

### The off hand

A lone white swing of a wearer who holds the weapon in the off hand alone brings 4 seconds on 8,871 landed firsts.

### Toofbreak

Toofbreaker's 6,329 landed firsts bring no second. Toofbreak 1320269 states a cooldown of 100 ms in the game files,
which no other proc in the table does. Two Toofbreak events in a row of one wearer, on any target, sit less than
100 ms apart 5 times in 36,381. Strength of Stone's sit that close 2,896 times in 27,242 and Ripsaw's Wound 1,962
times in 39,297, through the chain and through hits on two targets.

### Barman Shanker

Barman Shanker 12791 took its rate after the 27 were chosen (EV-0146), and its Rend 1300787 follows the same rule in
the sim. It is counted here on its own, the same way, and is in neither the table nor the totals above. The 11
reports downloaded by 2026-11-18 that show it run from 2026-11-13 to 2026-11-17, and every wearer is a rogue with it
in the main hand. A lone white swing brings 6 seconds on 125 landed firsts, 4.8% (2.2% to 10.1%), against 6.5% a
landed hit at EV-0146's rate, and no third. A lone ability brings none on 41 landed firsts, where two or three would be
there if an ability chained, so the abilities are too few to settle it for this weapon alone. A white swing with two
Rend is in `p6mVRBzvWMqnQZFc`, fight 2, at 21:44.354.

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
with a seal's strike or a Windfury Totem hit beside it.

In a unit with a white swing the weapon proc is the swing's with a probability of one over the unit's rolls, since
every roll has the same chance. Fitted by maximum likelihood on that split, a landed weapon proc a white swing cast
brings one more Fiery Weapon roll at 30% (9% to 53%), and one anything else cast brings one at 0% (0% to 10%). One more
roll at the base chance on every landed weapon proc fits worse by 26.5, as twice the log-likelihood, and one on a white
swing's proc alone fits within 0.2 of the best. EV-0130's last table pools the same units over what cast the proc.

Only 14 units are a white swing alone, so the 30% rests on the split of the mixed units. An oil's proc and Fiery
Weapon in the off hand were not counted.
