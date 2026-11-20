# EV-0158: Health regen's tick, its HP5 credit and the in-combat base rate

- Claim: health regenerates on a per-player timer of about 2.02 s that keeps running at full health; each tick credits
  HP5 / 5 health a second over the tick, 0.404 × HP5; and no base health regenerates in combat.
- Sources: local combat logs (advanced logging) of players of every class from level 8 to 60, solo and in dungeon
  groups, with each pass's logs below; HP5 from the COMBATANT_INFO gear and aura lists, and Spirit and level from the
  StatLedger addon where a ledger covers the log
- Game build: 1.60.1 builds by log date: not recorded for the logs of 2026-09-27 to 2026-10-13, then 70245, 70532,
  70630 to 70707, 70755, 70790, 70840, 70862, 70888, 70938 and 70959 (the logs of 2026-11-17 on US servers and of
  2026-11-18); the hotfix each player ran is not recorded
- Date: 2026-09-27 to 2026-11-18
- Pull request: #655

## Method

A player's regen is the residual between two of their own advanced-logging health snapshots: the change in health,
minus heals (amount less overheal), plus damage taken (amount less overkill). A pair counts only when max health is the
same at both ends, health is below max at both ends and nobody died between. Deep combat means no gap over 4 s between
combat events; out of combat means 6 s or more after the last, with no food, drink or first aid. In combat the whole
residual is HP5, since base regen is 0 there, so a fixed credit of x a tick shows integer ticks alternating between
floor(x) and ceil(x), with the fraction carried.

Scripts, pickles and tables are in the research folder's `health-regen/`, one `run_<date>.sh` per pass.

## The tick

- Phase fits on three players of the first pass (116 logs, 2026-09-27 to 2026-10-23): 97 to 99% of positive residuals
  on a 2.02–2.03 s grid, 95.6% at 2.00 s and 88% at 2.10 s (1,200 positives over 129 windows on the best-sampled one).
  Two later players fit 97.5% at 2.01 s and 97.1% at 2.02 s.
- The timer runs per player with any phase. Damage taken does not reset it, and own casts credit no health between ticks
  (1 of 333 intervals after a cast against 1 of 370 other lines).
- Full health does not stop the timer. A hit that ends a stretch of 2.5 s or more at full health puts the first tick
  back on the grid from before the stretch:

| Logs                    | Cases, gap < 40 s, both arcs tight | Back within 0.3 s | Controls (hit while hurt) | A restarted timer gives |
| ----------------------- | ---------------------------------: | ----------------: | ------------------------: | ----------------------: |
| 179 logs to 2026-11-15  |                   794, 288 players |             92.9% |            94.5%, n 3,937 |                   29.6% |
| 20 logs, 11-13 to 11-18 |                                 77 |             90.9% |              94.7%, n 588 |                   29.6% |

One log of the second set is a second recording of a session the first set holds, so the two sets are not wholly
independent. The phase was lost on at most 1.6% of the first set's cases (95% upper 3.6%). The delay from the hit to the first
tick spreads evenly over the tick, with no fixed delay, in and out of combat.

- Both sets fit a mean period of 2.025 s slightly better than 2.02 s: at 2.02 s the mean phase residual drifts +0.04 to
  +0.11 s over 6–40 s gaps, at 2.025 s within ±0.03 s.

## HP5 per tick

Single in-combat ticks of players whose HP5 the logs state. Upper is the count of ticks at the higher of the two modal
values; each hypothesis predicts its share from the fraction of x × HP5. p is a two-sided binomial test.

| HP5 (source)         | Class, level | Ticks | Upper | 0.400 predicts | 0.404 predicts | p against 0.400 / 0.404 |
| -------------------- | ------------ | ----: | ----: | -------------: | -------------: | ----------------------- |
| 31 (gear)            | priest, 48   |   263 |   131 |          0.400 |          0.524 | pairs below             |
| 27 (two gear pieces) | paladin, 56  |   269 |   252 |          0.800 |          0.908 | 4.6e-10 / 0.115         |
| 12 (gear)            | warrior, 57  |   241 |   212 |          0.800 |          0.848 | 0.0015 / 0.195          |
| 8 (gear)             | paladin, 49  |   700 |   168 |          0.200 |          0.232 | 0.011 / 0.64            |
| 8 (gear)             | shaman, 39   |   264 |    73 |          0.200 |          0.232 | 0.0035 / 0.105          |
| 9, 14 (gear)         | priests      |   138 |    89 |          0.600 |    0.636–0.656 | 0.42–0.58 / 0.70–0.74   |
| 12 (gear)            | shaman, 55   |    94 |    71 |          0.800 |          0.848 | 0.34 / 0.025, mixed     |
| 5 (gear)             | hunter, 51   |   108 |     0 |              0 |          0.020 | 1 / 0.23                |

- The 31 HP5 priest's ticks of 12 or 13 average 12.51 over 288 intervals, 263 of them single (0.404 gives 12.52,
  0.400 gives 12.40). Of 173
  consecutive pairs, 2 are 12-12 and 7 are 13-13; a 12.40 credit gives about 35 12-12 pairs and no 13-13.
- Earlier wearers of 5, 8, 9 and 12 HP5 read 2.021, 3.247, 3.631 and 4.852–4.866 a tick, each nearer 0.404 × HP5
  than 0.400 × HP5 (the 9 HP5 one barely).
- The 12 HP5 shaman's histogram holds ticks of 1 to 3 beside the modal pair, so its count is mixed. A second 5 HP5
  hunter reads 3 on 6 of 127 ticks against none for the one above; 5 HP5 stays unresolved.
- Minor and Lesser Troll's Blood (2 and 6 HP5) read 0.822 and 2.419 a tick, and Demon Armor, Demon Skin and two gear
  pieces read about 0.4 × HP5 on fewer ticks.

## Base regen in combat

- First pass: 8 non-Troll loggers show no positive residual over about 46,000 s of deep combat.
- The ledgers and the logs to 2026-11-10: 137 non-Troll players over 110,456 s.
- The logs of 2026-11-05 to 11-14: 798 of 1,037 player rows under 0.2 a tick, mean 0.018 over about 180,000 s, after
  subtracting HP5. The rows above it carry HP5 the logs do not list (a paladin row of 1.5 to 2.4 a tick matches 0.4 × a
  gear piece of 4 to 6 HP5; a Retribution paladin whose gear is logged reads 0.00 to 0.11 over about 33,000 s).
