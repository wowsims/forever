# EV-0028: Revelation's mage charge and Flamestrike's ticks at level 60

- Claim: the mage's Revelation charge goes as the buffed cast completes, and Flamestrike's ground ticks are hits on each enemy that roll their partial resist against a tenth of the target's share.
- Sources: combat log, one level 60 mage with 10% sheet crit (Stratholme, area pulls); the ticks' partial resist names a
  count section and lists its own
- Game build: not recorded; the ticks' partial resist has its own
- Date: not recorded; the ticks' partial resist has its own
- Pull request: #990; the ticks' partial resist is #1155

## The mage charge

81 of 84 mage charges go in the same millisecond as the buffed cast's success and the other 3 expire. Every
enemy of the spending area cast crits (79 of 79), and one charge covers one cast. Both buffed Pyroblasts in the log spend
the charge and crit, though the row's crit modifier names neither Pyroblast nor Fireball.

## Flamestrike's ground ticks

- Count: §Count of the 2026-11-19 to 2026-11-21 reports

Each Flamestrike ground tick is a hit of the tick spell 1279990 on every enemy in the area. Revelation rolls on
20 of 230 non-crit tick hits (8.7% a hit; per tick it would be 22.5%), so it rolls per enemy. Other listeners against the
ticks:

| Listener                       | Log                                                                     |
| ------------------------------ | ----------------------------------------------------------------------- |
| Revelation                     | 20 of 230 tick hits                                                     |
| Ignite                         | applied or refreshed after 67 of 68 tick crits, 0 of 287 non-crit ticks |
| Arcane Concentration           | Clearcasting after 4 of 95 ticks                                        |
| Master of Elements, Heating Up | 0 of 41 tick crits                                                      |
| Revelation's crit and charge   | ticks under the buff crit 16 of 76, as without it; no tick spends it    |

Ticks roll their partial resist against a tenth of the target's share: on pairs of caster and mob with a share of 0.1 or
more, 90 of 390 ticks show one, and they resist 0.10 of the share. In the sim, a
Flamestrike-only rotation on 5 targets (60 s) goes from 3.0 to 8.6 Revelation procs a fight once the ticks hit each
enemy.

## Count of the 2026-11-19 to 2026-11-21 reports

- Sources: Warcraft Logs reports of Forever (zone 3000): the mages of 120 random reports with a mage started from
  2026-11-21 00:00 UTC and 33 streams of 2026-11-20 reports of mages with piercing (EV-0142's count of the same reports)
- Game build: 1.60.1.71037 with hotfixes 114309, 114310, 114312, 114313 and 114314, live on the evening of 2026-11-19
  UTC; the hotfix a run was on and relogs are not recorded
- Date: 2026-11-19 23:00 to 2026-11-21 UTC
- Pull request: #1155

A tick's share is its resisted damage over the damage before the resist, set against the caster's own share a on the
same mob and school from control nukes in the same report, on pairs with a of 0.1 or more: EV-0142's Method.
