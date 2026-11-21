# EV-0140: Flurry on white swings and extra attacks

- Claim: a white swing spends a warrior Flurry charge whether it hits, misses or is avoided, and so did an
  extra-attack swing that hit or missed; a crit leaves Flurry at 3 charges. A Windfury Totem hit spends no charge,
  and its crit applies or refreshes Flurry (EV-0138).
- Sources: combat log, a level 60 dual-wield Fury warrior with Flurry, an Upper Blackrock Spire group with Windfury
  Totem rank 3 (hit 10610), 2026-11-16 21:30 to 21:56 UTC
- Game build: 1.60.1.70938, hotfixes up to push 114115, relogged
- Date: 2026-11-16
- Pull request: #1079

The log is the one EV-0138 counts. A white swing counts when Flurry was up, the warrior had no other own hit or cast
from 5 ms before to 60 ms after it, and no Hand of Justice proc came in the 400 ms before it; its Flurry lines are the
ones logged up to 50 ms after it. A Flurry charge a swing spends logs 0 to 47 ms after it. An extra swing is the first
own swing after the proc; Fiery Weapon and Flame Lash, which are no melee hits, may fall near it.

## White swings

| Swing           | Spends a charge | Spends none |
| --------------- | --------------: | ----------: |
| Hit             |             131 |           2 |
| Miss or avoided |              55 |           0 |

The two hits that spend none log no Flurry line within 50 ms. A crit ends at 3 charges every time: 53 log the refresh
with no removal, and 7 log a removal and then the refresh in the same ms.

## Extra attacks

Hand of Justice procs 13 times. Five settle the question. Three are glancing blows with Flurry down, and five show no
swing 8 to 17 ms after the proc, where every other proc's swing logs: in two a Heroic Strike or Cleave is cast 11 ms
after it, in two only a Windfury Totem hit follows, 9 and 12 ms after it, and in one nothing does.

| Extra swing | Cases | Flurry after it      |
| ----------- | ----: | -------------------- |
| Hit         |     1 | one charge spent     |
| Miss        |     1 | one charge spent     |
| Crit        |     3 | set to 3, no removal |

Sword Specialization never procs in the log.

## Windfury Totem hits

EV-0138 counts the Windfury hits in this log: a Windfury hit spends no Flurry charge, and its crit applies or refreshes
Flurry.
