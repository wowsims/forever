# EV-0125: Argent Avenger's and Argent Defender's proc rates

- Claim: Argent Avenger's 17352 rolls about 1 proc per minute on the wearer's landed weapon hits, specials included.
  Argent Defender's 17350 procs on about 1.6% of the landed melee swings taken outside its 20 s cooldown, and more often
  in a dungeon of Undead. The owner rules 1 proc per minute for the sword, and 1.5% a hit for the shield, 3% when an
  Undead strikes.
- Sources: public reports on forever.warcraftlogs.com, the fights whose combatant info lists the item on the wearer.
  Argent Avenger: ZfbY6RgWjx9hmpGz fights 1, 4, 9, 12, 13; 1z4vRDqGdM2cVPXk fight 1; Lpr76Vd3DbKWZqnh fights 1, 3;
  QHXdBJRj2GPpC3n1 fight 1; QxGpKz7knDvjWmYb fight 31; VjBTMHXxFvbthkw2 fight 1; NXMZV9xjvWpB7GtP fight 116;
  9zxHaWbYT3gqDQvF fight 17; z6LMfrpbwhyVCAJK fight 16; RGnyYzMgqCNdVwkT fight 21, left out below. Argent Defender:
  VjBTMHXxFvbthkw2 fight 1; NXMZV9xjvWpB7GtP fight 107; 7tLY1pVK9wRWPABX fights 15, 20; MWrPDAVFCBXzh8Z6 fights 1, 3, 8.
- Game build: not recorded; the reports were not checked against a build, and the hotfix state and relogs were not
  recorded either.
- Date: 2026-11-13 07:09 to 2026-11-15 09:41 UTC
- Pull request: not recorded yet

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
| Windfury Totem strikes                                           | 2 / 446             | 0.4%   | –            |
| Shield Slam                                                      | 0 / 498             | 0%     | –            |
| Thunder Clap                                                     | 0 / 597             | 0%     | –            |
| Weapon hits: white, Revenge, Cleave, Sunder Armor, Heroic Strike | 110 / 3,079         | 3.6%   | 0.97         |

Sunder Armor counts casts, so its chance is a floor. Warrior off hands also procced on Whirlwind's off-hand hit (25
procs). Windfury Totem strikes did not roll it, where EV-0119 records them rolling other weapons' procs.

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
strike you are 2 times as likely". A trial is a landed melee swing taken by the wearer more than 20 s after the
wearer's last proc; abilities taken are not counted. A proc is an application or refresh of 17351 logged from 50 ms
before to 150 ms after the swing.

| Wearer  | Dungeons                     | Procs | Landed swings taken outside the cooldown | Procs on them       |
| ------- | ---------------------------- | ----- | ---------------------------------------- | ------------------- |
| Paladin | Blackrock Spire              | 8     | 389                                      | 8 (2.1%)            |
| Warrior | Blackrock Spire              | 9     | 292                                      | 8 (2.7%)            |
| Paladin | Blackrock Spire, Scholomance | 44    | 2,465                                    | 41 (1.7%)           |
| Paladin | Dire Maul                    | 20    | 1,483                                    | 17 (1.1%)           |
| All     |                              | 81    | 4,629                                    | 74, 1.6% [1.3, 2.0] |

About 1.6% of the swings proc, so the 100 in the column is not the chance. The shortest gap between two procs of one wearer is 20.8 s, and the 748
landed swings taken inside 20 s of a proc procced nothing, where 12 were expected at 1.6%; that fits the 20 s
cooldown. The paladin who ran both dungeons procced on 26 of 1,129 swings in Scholomance (2.3%) and 15 of 1,336 in
Blackrock Spire (1.1%), in line with the tooltip's 2 times against Undead. The attacker's creature type was not read,
so this is a split by dungeon only.
