# EV-0137: Lightning Overload rolls Elemental Focus, and Elemental shaman talents in two logs

- Claim: a Lightning Overload's launch rolls Elemental Focus once, as a cast does, and keeps a Clearcasting charge
  that is up; Lightning Overload 3/3 overloads 30% of landed Lightning Bolts; each target a Chain Lightning lands on
  rolls its own overload at about a third of that, and the overload chains from that target; Lightning Mastery's buff
  lasts 15 s, no cast spends it, and Chain Lightning has no cooldown while it is up.
- Sources: combat log, an Orc Elemental shaman, level 49 (50 for the last cast), Lightning Overload 3/3, Elemental
  Focus, Lightning Mastery, solo in Felwood and Tanaris, 2026-11-15; combat log, an Elemental shaman, level 60,
  Lightning Overload 3/3, Elemental Focus, Lightning Mastery, in a Stratholme group, 2026-11-13
- Game build: 1.60.1.70938 for both logs; the 2026-11-13 log ran from 22:26 to 00:12 UTC, before the 2026-11-14
  02:00 UTC hotfix. The hotfix each log ran on is not recorded.
- Date: 2026-11-13 and 2026-11-15
- Pull request: #TBD

## Elemental Focus on the overload

A Clearcasting gain (`SPELL_AURA_APPLIED` of 16246) counts against a Lightning Bolt's impact when it lands from 5 ms
before to 30 ms after the bolt's damage line, with no own cast line between the impact and the gain, and Clearcasting is
down before the impact. An overload follows when the overload row of the bolt's rank hits the same target within 3 s
with no other bolt impact on it between. The level 49 log writes the gain in the impact's ms. The level 60 log writes it
5 to 25 ms after the bolt's damage line, while its gains from own casts share the cast's ms; a 5 ms window there counts
1 of 178.

| Log                  | Gain at the impact, overload follows | Gain at the impact, no overload |
| -------------------- | ------------------------------------ | ------------------------------- |
| level 49, 2026-11-15 | 8/70 = 11.4% [5.9, 21.0]             | 0/187 = 0.0% [0.0, 2.0]         |
| level 60, 2026-11-13 | 23/178 = 12.9% [8.8, 18.6]           | 1/378 = 0.3% [0.0, 1.5]         |
| Pooled               | 31/248 = 12.5% [8.9, 17.2]           | 1/565 = 0.2% [0.0, 1.0]         |

The gain sits between the bolt's impact and the overload's hit, which follows a median 0.64 s later in the level 49 log,
so it is the overload's launch that rolls. Chain Lightning's overload launches tens of ms after the cast in the level 60
log: of its Chain Lightning casts with Clearcasting down and no gain in the cast's ms, 7/45 = 15.6% [7.7, 28.8] gain 5
to 60 ms after the cast when an overload follows within 0.2 s, against 0/155 when none does. The level 49 log's Chain
Lightning overloads hit in the cast's ms or 90 to 110 ms later, so that window finds none of them (0/10).

The level 49 log's fit over every roll with Clearcasting down puts the chance at 0.122 [0.099, 0.148] over 656 rolls,
the overload launches 8/66 = 12.1% among them; heals, totems, shields and cooldowns gain it 0 of 274 times. An overload
launch with Clearcasting up kept the charge 5 of 5 times.

## Lightning Overload

Overloads per landed Lightning Bolt: level 49 75/281 = 26.7% [21.9, 32.2], level 60 198/613 = 32.3% [28.7, 36.1],
pooled 273/894 = 30.5% [27.6, 33.6].

Chain Lightning's overload starts on the target that rolled it and chains over the parent's targets. Overloads by the
position of the target they start on, against the landed hits there:

| Log      | Primary target | Jump targets |
| -------- | -------------- | ------------ |
| level 49 | 14/138         | 3/105        |
| level 60 | 28/268         | 40/407       |

Pooled, one chance on every landed target fits q = 0.093 [0.076, 0.112]; a separate factor on the jump targets reads
0.79 [0.54, 1.19]. Overloads only from the primary target cannot produce the 43 that start on a jump target, and one
roll per cast cannot produce the 6 casts with two overloads. The level 49 log's jump targets read low against the level
60 log's (Fisher p 0.018), cause unknown.

## Lightning Mastery

| Log      | Windows | Length        | Casts inside, most in one window | Shortest Chain Lightning gap inside |
| -------- | ------- | ------------- | -------------------------------- | ----------------------------------- |
| level 49 | 16      | 14.92-15.10 s | 7 bolts, 5 Chain Lightnings      | 1.65 s                              |
| level 60 | 17      | 15.00-15.02 s | 9 bolts, 7 Chain Lightnings      | 1.99 s                              |

No window ends in the ms of a cast. Outside the windows the level 49 log has no two Chain Lightnings under 8 s apart,
and the level 60 log none under 7.96 s where the first of the two is outside too. The level 60 log's nine shorter gaps
that end outside start with a cast in a window's last 3.1 s, so the window's edges are not settled.
