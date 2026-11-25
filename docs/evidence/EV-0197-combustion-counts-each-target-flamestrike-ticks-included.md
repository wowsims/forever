# EV-0197: Combustion counts each enemy a hit lands on, Flamestrike's ground ticks included

- Claim: Each landed Fire spell hit adds a stack of Combustion's crit (28682, up to 10), and each crit spends one of
  Combustion's three charges (11129), once per enemy hit. Each enemy a Flamestrike ground tick (1279983 to 1279990, 1314281) lands on adds a stack, and each tick crit spends a charge, as the hits of an area cast do. A tick the enemy
  is immune to adds nothing.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the streams on disk of the mages that gained 28682: 93
  streams in 29 reports, the reports and fights §Reports lists
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20
  18:23 UTC); the range spans hotfix push 114313 (2026-11-20 23:47 UTC); the hotfix each player ran is not recorded
- Date: 2026-11-14 to 2026-11-22
- Pull request: not yet opened

## Method

A Combustion window runs from 28682's apply to its removal. A crit is one of the mage's direct Fire damage events
(Fireball, Fire Blast, Scorch, Pyroblast, Flamestrike, Blast Wave and Frostfire Bolt; DoT ticks and Ignite left out)
with a crit outcome, a partially resisted crit included. Three counts of the crits each predict that the window ends at
the third crit, and a window fits a count when 28682 goes within 100 ms of that crit:

- Each hit: every crit event counts, one per enemy, ground ticks included.
- One per tick: a ground tick's crits on several enemies count once.
- No ticks: ground tick crits count nothing.

## Charges

| Build | Windows | Fit each hit | Fewer than 3 crits, mage died | Fewer than 3 crits, alive | With a tick crit | Each hit | One per tick | No ticks |
| ----- | ------: | -----------: | ----------------------------: | ------------------------: | ---------------: | -------: | -----------: | -------: |
| 70938 |      14 |           14 |                             0 |                         0 |                1 |        1 |            1 |        0 |
| 70959 |      54 |           53 |                             1 |                         0 |                7 |        7 |            4 |        2 |
| 71037 |     111 |          107 |                             2 |                         2 |               21 |       21 |           10 |        0 |
| All   |     179 |          174 |                             3 |                         2 |               29 |       29 |           15 |        2 |

No window outlives the third crit of the each-hit count. Two windows end with fewer than three crits and no death:
dr1BMh6pNxqbFJLf (1.6 s, two crits) and pAWhgCd1vP2nKHML (111 s, ending at a Flamestrike cast). In the 21 windows
where one area cast other than a tick crit on several enemies at once, the each-hit count fits 20 and a count of one per
cast fits 2.

## Stacks

The stack sample is small. A ground tick counts here when no other cast or damage of the mage falls within 25 ms of it.
16 such ticks that did not end the buff added one stack for each enemy they landed on, all within 60 ms; three of them
were single-target ticks the enemy was immune to, and they added none. In the 10 windows with a landed tick at least
100 ms before the end, the last stack is 1 plus the window's landed Fire hits, ticks included, in 9; left without the
ticks, the count fits none.

## Reports

Report codes and fights: 1xqRkXMpVBW82CHY 1; 26CZDjqdH93v4bfA 2; 7GcATyt6WRjBJvaQ 1; Awfm8RhtDBk71vzQ 1, 4;
BFPtcmbyLr71Kj63 1; bWRrmTpMycdvxhJa 1; bwYf1ptBqV9ym8HL 1; C32kRh9jA6DZVmWa 1; DfTMrcFAzq9ayQKt 1;
dKZ7WpqywDNBvVP1 1, 3; dr1BMh6pNxqbFJLf 1; fvq6hnmGpTkzBAbc 1; Gcb4J2hW137wT9MX 4; GK4DBfJQNq3a9FLY 1;
GMXVQBWT1Rw94mDN 2; gQGP6xRKvjM1789J 1; J8LjWZNM9xB7nYDm 2; MQckGPJgVb19mrT2 1, 6; N4738YLMnaXp9rw6 1;
pAWhgCd1vP2nKHML 1; pjYha813ZfDNnQT2 1; RHr283cN6AF4Za7D 1; T9FpP7vKQaL6nNj4 1, 5; TghGLcMjP1d6r8NV 1;
vnj8MpcGHydhY9Nt 1, 6; wr2Xh84BW6vRtbZ9 1; WRXxpjZcGbt4C16d 1; x4kKfzHQZ2CVjn1J 1, 4; XgQYbwGBZcHLKRxV 1, 3.
