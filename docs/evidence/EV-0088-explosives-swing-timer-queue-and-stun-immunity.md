# EV-0088: Throwing an explosive keeps the swing timer, and stun explosives come back immune on raid bosses

- Claim: an explosive throw leaves the melee swing timer running, so the next swing lands on its old schedule and a swing
  due during the cast lands in it; no swing landing in a dynamite or grenade cast was a queued attack; and Iron and
  Thorium Grenade, whose spells state the stun mechanic, come back immune on a stun-immune target, their damage
  included, while dynamite and sapper charges, which state no mechanic, hit the same bosses.
- Sources: Forever combat logs of every explosive thrown by a player who was meleeing (six throwers across five logs);
  Warcraft Logs from outside Forever: Classic Era Molten Core and Blackwing Lair (for the queue, 50 reports, warrior
  throws, duplicate events removed), a Fresh Naxxramas report, and TBC Hyjal and Black Temple; report codes not recorded
- Game build: not recorded for the Forever logs; not Forever for the rest
- Date: the Forever logs dated 2026-10-01 to 2026-10-12 in their names, and one named WoWCombatLog-091826_120108; the
  reports from outside Forever are not recorded
- Pull request: #783

## The swing timer and the queue

Each throw is checked against two models: the timer running on (next swing = last swing + weapon speed)
or the cast restarting it (cast end + weapon speed). The explosive rows carry no Do Not Reset Combat Timers (attributes
word 2 0x20000), so the sim had reset the swing on every throw.

| Log                                       | Explosive         | Throws | Error if the timer runs on                            | Error if the cast resets it |
| ----------------------------------------- | ----------------- | ------ | ----------------------------------------------------- | --------------------------- |
| WoWCombatLog-091826_120108                | Rough Copper Bomb | 13     | 10 within plus or minus 0.03 s, 3 at +0.73 to +0.98 s | -0.29 to -3.67 s            |
| 2026-10-12 mage log, fof frostbite        | Iron Grenade      | 3      | -0.08, -0.08, +0.57 s                                 | -1.42 to -2.79 s            |
| 2026-10-01 warrior log, 99 AP at level 13 | Rough Dynamite    | 2      | +0.04, +0.02 s                                        | -1.85, -1.89 s              |
| same log, a second thrower                | Rough Dynamite    | 1      | +0.08 s                                               | -1.83 s                     |
| 2026-10-08 mage log, fof frostbite        | Rough Copper Bomb | 1      | +0.03 s                                               | -2.47 s                     |
| 2026-10-12 warrior log, sword spec        | Large Copper Bomb | 1      | -0.01 s                                               | -1.52 s                     |

No throw reset the swing, and six had a white swing land inside the 1 s cast, so the cast does not hold the swing either.
The three late swings (+0.6 to +1.0 s) fit neither model; none is early. Sapper charges and instant explosives are not
measured, since the logs hold no melee thrower of them.

A combat log does not show the queue. In the Classic Era logs no swing that lands during an explosive's cast is a queued
attack, where a warrior's Heroic Strike or Cleave share over the same fights is 8 to 29%:

| Explosive       | Throws | Swings during the cast | Of them Heroic Strike or Cleave | Expected if the queue survived |
| --------------- | ------ | ---------------------- | ------------------------------- | ------------------------------ |
| Dense Dynamite  | 314    | 226                    | 0                               | about 59                       |
| Solid Dynamite  | 34     | 23                     | 0                               | about 6                        |
| Iron Grenade    | 30     | 16                     | 0                               | about 5                        |
| Thorium Grenade | 9      | 5                      | 0                               | about 0                        |
| Heavy Dynamite  | 3      | 1                      | 0                               | about 0                        |

No warrior threw a bomb in the sample. The first swing after a throw is a queued attack about as often for Dense Dynamite
(18%) as for the instant Goblin Sapper Charge (19%, 438 throws), since players queue again at once, so the logs cannot tell
a dropped queue from a held one, nor what the sapper does. The rule that dynamite, grenades and bombs drop the queue and
the sapper charge keeps it is the owner's ruling.

## Stun explosives on stun-immune targets

| Reading                                                        | Measured                |
| -------------------------------------------------------------- | ----------------------- |
| Thorium Grenade on bosses                                      | 7 of 7 immune, 0 damage |
| Iron Grenade and Thorium Grenade on stun-immune Naxxramas mobs | 5 of 5 immune           |

The level at which targets are stun-immune by default (63) is the owner's ruling.
