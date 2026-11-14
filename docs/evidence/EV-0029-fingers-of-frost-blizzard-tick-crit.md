# EV-0029: Fingers of Frost gives Blizzard's ticks no crit

- Claim: Fingers of Frost's charges raise a mage's direct spells' crit chance and leave Blizzard's ticks at their own rate.
- Sources: foreverlogs.gg public reports, about 3,150 of 6,694 encounters, and a refetch of three mages for the charge counts; the five mages whose direct spells prove Shatter
- Game build: not recorded
- Date: not recorded
- Pull request: #761

Crit counts of the five mages whose direct spells prove Shatter. Frozen targets and ticks with Revelation up are left out.

| Crit           | No charges        | Charges gained mid-channel | Charges up at the cast |
| -------------- | ----------------- | -------------------------- | ---------------------- |
| Blizzard ticks | 29 of 438, 6.6%   | 6 of 188, 3.2%             | 18 of 267, 6.7%        |
| Direct spells  | 144 of 1925, 7.5% | not recorded               | 266 of 686, 38.8%      |

Charges give the same mages' direct spells about 31 more points of crit and their ticks none. Ticks with Revelation up crit
1386 of 1386, so unfiltered tick crit climbs through a channel with or without charges. One mage supplies 183 of the 188
mid-channel ticks and has the thinnest Shatter evidence (13 direct spells). The local raw logs hold too few ticks with
charges up to decide this alone.

Whether Shatter reaches ticks on really frozen targets is open; the sim never freezes a target.
