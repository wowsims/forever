# EV-0037: Bear form health and armor, Faerie Fire and Lacerate threat, bear rage, Furor and the swing timer, druid logs of levels 25 to 60

- Claim: the bear forms' health is flat and follows the level, Bear Form's 180% and Dire Bear Form's 360% armor apply once to the items' armor and leave the armor from Agility and buffs whole, Faerie Fire's threat rule holds at the level 60 rank and Lacerate's 10/3 at the later ranks, a bear's white hit takes the two-hander rage factor, Furor 5/5 and 0/5 restore Cat Form energy on every entry, and an instant cast leaves the swing timer running.
- Sources: two sets of logs (combat logs, StatLedger and ThreatLog). Source A is a druid at level 36 in two sessions, with a two-handed mace and then a dagger, and Furor at 0/5, 1/5 and 5/5. Source B is a Tauren feral druid levelled from 25 to 60 without Furor, with StatLedger and ThreatLog all the way and combat logs at levels 52 to 56 including a Sunken Temple run.
- Game build: 1.60.1; no build number recorded
- Date: source A 2026-11-12 and 2026-11-13; source B 2026-11-06 to 2026-11-13
- Pull request: #984

Rage and energy come from the advanced log fields, which hold power in tenths. Threat comes from ThreatLog entries whose actions are the one ability, matched to the combat log's damage where there is one.

Each section below is one rule, with the source that measured it.

## A bear's white hit takes the two-hander rage factor whatever it holds, and a crit doubles it

In source A, 15 of 17 clean hits with a two-handed mace (Defender of Dalaran) gave 11.2 to 11.3 rage, glancing blows
included, and crits 22.5. With a dagger (Poniard), 12 clean hits gave 11.2 to 11.3 and 3 crits 22.4 to 22.5.

## An instant cast leaves the swing timer running (Faerie Fire)

In source A, under a 15% haste buff, 7 of 7 casts between Bear Form swings left gaps of 2.09 to 2.22 s, against 2.18 s
without a cast.

## The forms' health is flat and follows the level

Bear Form adds 20 plus 18 a level above 10, up to 560 at 40, and Dire Bear Form 600 plus 32 a level above 40, 1240 at 60. Source A reads +488 at both 2249 and 2339 maximum health. In source B every level matches the rows times 1.05, a
Tauren's Endurance, from 304 at 25 through 588 at 40 and 798 at 45 (34 of 36) to 1302 at 60.

## Endurance scales all health, flat health added later included

In source B the Tauren's bear forms add 5% more than their rows at every level from 25 to 60. The sim matches: printed at
runtime, Dire Bear Form adds 1302 to a Tauren and 1240 to a Night Elf at 60.

## Bear Form's 180% and Dire Bear Form's 360% armor apply once and leave the armor from Agility and buffs whole

Each applies to its own part of the items' armor. In source A, 621 armor from items gained 1118, with Mark of the Wild
up or not. In source B the forms read 1.799 to 1.802 and 3.599 to 3.601 times the items' base armor at every level, and
the bonus armor rose by the same factor times the gear's bonus armor (0 to 205).

## Furor 5/5 follows its tooltip

In source A, 5 of 5 Bear Form entries gave 10 rage. Cat Form came back at 80 and 87 energy where the rule gives 81 and 86,
after a shift and after cancelling Cat Form. Stints of 5.4 s and 8 s in Bear Form added nothing.

## Every Cat Form entry takes Furor's share, the first included

Both sources show it. At 0/5, 275 of 275 entries started at 0 (-1.3 to +1.5 energy), from caster, Travel, Aquatic and Dire
Bear Form, after 2 s to over 10 minutes out of Cat Form, and the first after each of four logins. At 1/5, the first Cat
Form in a long while came back near 20, the 1/5 cap. 247 of 247 Cat Form entries sit between -1.32 and +1.52 energy.

## Faerie Fire's rule holds at the level 60 rank

The rule is twice the level times the form's multiplier in either bear form, and none in Cat Form. In source B, 9907 at
levels 56 to 59 read exactly round(2 × level × 2.15) on 18 of 20 first entries on a mob whose only action is Faerie
Fire, and at 60 a mid-fight entry whose only action was Faerie Fire rose by exactly 258. Over levels 37 to 59, both
bear forms together, 176 of 198 first entries sit within 0.6 of 2 × level × 2.15. 21178 is +115% threat, so 2.15.

## Lacerate's 10/3 holds for the later ranks in Dire Bear Form

It has no flat part, and each hit or tick takes the multiplier of the form the druid is in when it lands. The readings
are source B's. Enrage's rage, which the combat log never shows, adds about 21.5 threat a second and spoils a window.
Rank 4 at levels 53 to 56, in windows with no Enrage tick, Blood Frenzy proc or other mob, fits 10/3 with a spread of
0.4 on 21 windows, and 1.3 on 53. A tick that landed after a shift to caster form threatened for 48 × 10/3. Ranks 3
and 5, from the ThreatLog alone, show tick values at 10/3 × 2.15 per stack (rank 5: 41 entries at 1, 2, 3 and 5 stacks,
against about 5 expected), not moving with attack power.

## Swipe strikes three targets, its own among them, and a miss takes one of the three

In source A, 21 casts with four mobs in front each struck exactly three. In source B, none of 1,830 casts struck more
than three; with four or more mobs on the bear 556 of 575 struck three, and the 81 with a miss, dodge or parry each had
three results; 567 of 588 in the last 3 s.

## Rage from a spell taken is about twice the attacker's melee, and a weapon special about the melee rate

Both sources contribute. The values are medians, with deltas next to the druid's own casts and under Barkskin dropped.
Attackers without mana, whose melee gives 1.01 times the rule (691 hits), give 2.11 on Poison ticks (15), 2.0 on
Fevered Plague and 1.82 on Venom Spit. Mana users, whose melee gives 0.61 (83), give 1.20 on Shadow Bolt (7), 1.24 on
Shadow Word: Pain ticks (11) and 1.16 on Lightning Bolt (4), and about 1.0 on 7 hits at level 36. Among weapon specials
from attackers without mana, Shield Slam reads 1.05, Strike 0.97 and Shoot 0.95; a mana user's Shoot reads 0.57 to 0.66
against its melee's 0.59, and Cleave about 1.3 (11) at level 36. Mortal Strike reads 1.2 to 1.3.

## Other readings

Some checks changed nothing. Ability crits do 2.2 times a hit and white crits 2.0 times (Swipe crits of 100 to 108 against hits of 46 to 49 on the same mobs, all of base 64). A Blood Frenzy proc reads about 6 threat a mob in multi-mob pulls. Enrage's rage, 2 a second, never shows in the combat log; it accounts for jumps of 2.0 rage between snapshots and about 21.5 threat a tick (2 × 5 × 2.15) in the ThreatLog.

The pull request leaves four points open: what death does to the stored Furor energy; why a spell gives twice the attacker's melee rage and why Cleave and Mortal Strike read 1.2 to 1.3; why in the Sunken Temple fights every ability threatened as if it did 1 more damage than the log shows; and Lacerate's ranks 3 and 5, which rest on ThreatLog values alone.
