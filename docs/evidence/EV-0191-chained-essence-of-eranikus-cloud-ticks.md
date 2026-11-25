# EV-0191: Chained Essence of Eranikus's cloud and its ticks

- Claim: the cloud of Chained Essence of Eranikus (item 10455, use 12766) stays where it was made for 20 s and puts
  1300240 on every enemy inside it, 22 at once with no cap seen. Each enemy takes Nature damage every 2 s, 140 on level
  58 and 60 enemies, from 2 s after the debuff goes on, with no spell power added, and a crit deals 1.5 times. The
  cloud's end takes the debuff just before the tenth tick, so an enemy inside from the start takes 9.
- Sources: combat log `WoWCombatLog-112426_191754`, recorded by a level 60 rogue from 2026-11-24 19:17 to 23:25 UTC-5
  (2026-11-25 00:17 to 04:25 UTC). The item's wearer is a level 60 Enhancement shaman in the second of three
  Scholomance runs; the enemies are levels 58 to 61.
- Game build: 1.60.1.71091; the hotfix push the session ran on is not recorded, and whether the players relogged is
  unknown
- Date: 2026-11-24
- Pull request: #1222

The shaman used the item 3 times. Each use logs 12766's cast and buff and a summon of creature 8506, and then
1300240's cast, debuff and ticks on the enemies, all under the shaman. The creature logs nothing of its own. The
positions are the advanced log's x and y of each unit's nearest event, so a distance can be off by the distance the
unit moved in up to 2.5 s.

| Use | Enemies that took the debuff | Most holding it at once | Ticks | Shaman's spell power (log SP + 0.3 Int) |
| --- | ---------------------------- | ----------------------- | ----- | --------------------------------------- |
| 1   | 10                           | 4                       | 26    | 20 + 68 = 88                            |
| 2   | 29                           | 22                      | 93    | 88                                      |
| 3   | 2                            | 2                       | 9     | 88                                      |

- **The cloud stays at the spot of the use.** In use 2 the shaman was 35 yd from the spot by 16 s, and five enemies 2.7
  to 4.5 yd from the spot kept the debuff to the cloud's end. An enemy that walked out lost it: of the 27 over the three
  uses that lost it alive before the end, 25 stood 5.6 to 8.4 yd from the spot at their nearest position, one 4.7 yd and
  one 11.0 yd. Enemies that came in later took it then, up to 11.9 s after the use.
- **Every enemy inside takes it.** 16 enemies took it within 52 ms of use 2, and 22 held it at once.
- **The ticks start 2 s after the debuff goes on**, 2.00 to 2.02 s on all 35 enemies that ticked, and come every 2 s.
  The other 6 of the 41 took no tick: in use 1 five Spectral Projections took the debuff and lost it in the same
  instant at 18.155 s, and in use 3 a Rat died 0.9 s after it went on. The Spectral Tutor of use 1 took 5 ticks from
  8.68 s to 16.68 s, then logged no tick at 18.7 s and no removal.
- **The cloud's end takes the debuff.** 12766 leaves the shaman at 19.997 to 20.013 s, and the debuff leaves the
  enemies still inside at 20.02 to 20.024 s. It went on 37 to 64 ms after the use, so the tenth tick would fall at
  20.04 to 20.07 s, after the debuff is gone. Darkmaster Gandling, the one enemy inside for the whole cloud (0.064 s
  to 20.02 s), took 9 ticks. The Risen Protector of use 1 also took 9 but died at 19.0 s, before its tenth was due.
- **No spell power.** With 88 spell power, the 59 ticks on level 58 and 60 enemies that did not crit deal exactly 140,
  so a spell power share is below 0.006.
- **Amount by target level**, as logged:

| Target level | Ticks that did not crit | Amount   | Logged base amount |
| ------------ | ----------------------- | -------- | ------------------ |
| 58           | 54                      | 140      | 140                |
| 59           | 58                      | 139      | 138                |
| 60           | 5                       | 140      | 140                |
| 61           | 9                       | 143, 144 | 143                |

- **Crits:** 2 of 128 ticks, each 210 on a 140 tick, 1.5 times. Two crits give the multiplier and no rate.
- **No miss:** all 41 applications landed and none of the 128 ticks missed. That many on level 58 to 61 enemies cannot
  tell a spell hit roll from none. Resists: EV-0142.

Not measured: the cloud's radius from the game files (the area trigger 43368 is not in the spell store), the tooltip's
damage against Trolls (no Troll was in the cloud), and an enemy of level 62 or 63.

The counts come from `scripts/shaman_chained_essence.py` in the 2026-11-24 Scholomance research folder, which reads the
raw log around each use.
