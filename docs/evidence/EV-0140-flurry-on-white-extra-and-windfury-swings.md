# EV-0140: Flurry on white swings, extra attacks and Windfury Totem's strike

- Claim: a white swing or an extra-attack swing spends a warrior Flurry charge whether it hits, misses or is avoided,
  and a crit leaves Flurry at 3 charges; Windfury Totem's strike spends no charge, and its crit sets Flurry to 3.
- Sources: combat log, a level 60 dual-wield Fury warrior with Flurry, an Upper Blackrock Spire group with Windfury
  Totem rank 3 (strike 10610), 2026-11-16 21:30 to 21:56 UTC
- Game build: 1.60.1.70938, hotfixes up to push 114115, relogged
- Date: 2026-11-16
- Pull request: #1080

The log is the one EV-0138 counts. A swing counts when Flurry was up and the warrior had no other own hit from 5 ms
before to 60 ms after it. A Flurry charge a swing spends logs 0 to 47 ms after it.

## White swings

| Swing           | Spends a charge | Spends none |
| --------------- | --------------: | ----------: |
| Hit             |             131 |           2 |
| Miss or avoided |              55 |           0 |

A crit ends at 3 charges every time: 53 log the refresh with no removal, and 7 log a removal and then the refresh in
the same ms.

## Extra attacks

Hand of Justice granted 13 extra swings. Seven settle the question; the other six are glancing blows with Flurry down
(3) or procs whose swing the log does not show (3).

| Extra swing | Cases | Flurry after it      |
| ----------- | ----: | -------------------- |
| Hit         |     3 | one charge spent     |
| Miss        |     1 | one charge spent     |
| Crit        |     3 | set to 3, no removal |

One of the hits logs 178 ms after the proc, so it may be the next white swing. Sword Specialization never procs in
the log.

## Windfury Totem's strike

Of 109 strikes, 22 crit. With Flurry up, 68 strikes have no removal after them and 4 have one that a white swing,
Heroic Strike or Cleave explains. The 2 others are avoided strikes whose removal logs in the miss's ms, with a crit's
refresh right after, so neither reads as the strike's spend. Only one landed strike has no other own hit within 50
ms, and no Flurry line follows it.

Of the 22 strike crits, 14 with no other crit nearby gave Flurry: 7 applied it and 7 refreshed it to 3. The other 8
had another crit nearby.
