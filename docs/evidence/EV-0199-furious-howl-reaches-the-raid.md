# EV-0199: Furious Howl reaches the raid, players of every class and their pets

- Claim: Furious Howl (24597, and 24603 to 24605 below it) puts its attack power on the players of the raid, of every
  class, and on their pets, outside the hunter's party too: one howl reached up to nine other players. The hunter is
  among those it reaches on about half the howls.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the streams on disk of the hunters whose pet cast Furious
  Howl: 168 reports, too many to list. The howls that reached five or more other players are in the Upper Blackrock
  Spire fights GTWvLjXyf9qD2rwB 2, q3Hr8jghPC4QvV9k 1 and y46bmVXZzPrvwG81 1, and one in QHXdBJRj2GPpC3n1 1
- Game build: 1.60.1.70888 to 1.60.1.71037 by report date (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); the range spans hotfix push 114313 (2026-11-20 23:47 UTC); the hotfix each player ran is not recorded
- Date: 2026-11-11 to 2026-11-24, the last report starting at 11:21 UTC
- Pull request: #1233

## Method

For each Furious Howl the hunter's pet casts, the units it reaches are the targets of that pet's Furious Howl applies and
refreshes within 200 ms of the cast. Players are the report's actors with a class; the hunter and the hunter's own pet
are counted apart. A party holds the hunter and four others, so a howl that reaches five or more other players reaches
outside the party. The streams carry neither party membership nor positions, so the 15 yd radius the row states
cannot be read from them.

## Counts

Howls by the number of other players they reached:

| Fights                                       |  Howls |  None | 1 to 4 | 5 to 9 |
| -------------------------------------------- | -----: | ----: | -----: | -----: |
| Upper Blackrock Spire and General Drakkisath |    184 |   122 |     39 |     23 |
| All others                                   | 12,151 | 4,221 |  7,929 |      1 |

The 23 Upper Blackrock Spire howls are all on 71037. The one other howl is in fight 1 of QHXdBJRj2GPpC3n1, a Blackrock
Spire run on 70938.

Applies and refreshes on other players by class: Paladin 4,135, Druid 3,451, Shaman 3,173, Priest 2,353, Mage 2,261,
Warlock 1,744, Warrior 1,412, Rogue 1,398, Hunter 437. Of the 12,335 howls, 2,701 reached another player's pet, 6,442
reached the hunter and 10,326 the casting pet.

Howls that reached another player, by build: 70888 2 of 2, 70938 2,983 of 4,398, 70959 1,802 of 2,892, 71037 3,205 of
5,043.
