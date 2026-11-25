# EV-0205: Soul Siphon reads the effects on the target at each tick of Wrack and Drain Life

- Claim: a Soul Siphon channel's tick takes 12% of its base per Corruption, Siphon Life and Bane of Agony of the warlock on
  the target, added together, counted when the tick lands; the damage does not snapshot at the channel's start.
- Sources: a combat log of an Affliction warlock at levels 46 to 47 in Zul'Farrak (728 Wrack ticks and 224 Drain Life
  ticks); report code not recorded.
- Game build: 1.60.1.70938 by the log's date; the hotfix the player ran is not recorded
- Date: 2026-11-14
- Pull request: #1232

Method: the warlock's Corruption, Siphon Life and Bane of Agony on each target are tracked from their aura applications
and removals. Two consecutive ticks of one channel on one target, 1.0 s apart (within 0.15 s), count as a pair when
neither crit and neither killed the target. A pair is compared by the ratio of its second tick's damage to its first.
Where the number of effects is the same at both ticks, the ratio is 1. Where it falls by one between the ticks, a
snapshot would keep 1.0, and a read at each tick gives `(1 + 0.12 x n2) / (1 + 0.12 x n1)`.

| Pair           | Wrack pairs | Ratio range    | Drain Life pairs | Ratio range    | Read at each tick | Snapshot |
| -------------- | ----------- | -------------- | ---------------- | -------------- | ----------------- | -------- |
| 3 effects to 2 | 23          | 0.907 to 0.915 | 3                | 0.900 to 0.917 | 0.9118            | 1.0      |
| 2 effects to 1 | 8           | 0.890 to 0.907 | 1                | 0.899          | 0.9032            | 1.0      |
| Same count     | 388         | 0.983 to 1.011 | 121              | 0.987 to 1.013 | 1.0               | 1.0      |

No pair has the count rise between its ticks, and Drain Soul's 11 ticks form no pair.
