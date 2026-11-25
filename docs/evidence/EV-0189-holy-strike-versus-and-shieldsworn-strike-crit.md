# EV-0189: Holy Strike's spell damage against a creature type, and the crits of Shieldsworn's strike

- Claim: Holy Strike Rank 8 (10333) takes none of the spell damage against the target's creature type, while Exorcism
  and Shieldsworn's strike (Holy Shield 456544) take it. Shieldsworn's strike crits for 1.5 times, at a rate that fits
  the paladin's spell crit chance, while Holy Shield's damage on a block (1239570) never crits.
- Sources: combat log `WoWCombatLog-112426_191754`, logged by a rogue in the group: three level 60 Scholomance runs,
  2026-11-24 19:17 to 23:25 in the player's local time (UTC−5), 2026-11-25 00:17 to 04:25 UTC. The counts are the
  Protection paladin's with Shieldsworn, in runs 2 and 3 (21:18 to 22:04 and 22:26 to 23:21 local); the Retribution
  paladin of run 1 wears no gear with spell damage against a creature type.
- Game build: 1.60.1.71091; the hotfix push the session ran on is not recorded, nor whether the player relogged after
  it
- Date: 2026-11-24
- Pull request: #1220

## Holy Strike against Undead

The paladin wears Rune of the Dawn (19812), whose equip spell 24198 states 46 spell damage against Undead; the log's
spell power field reads 69 on every event and holds none of it. Each Holy Strike hit with no resist, block or absorb
is read at the log's base amount, the amount before the crit, and set against the sim's formula without the 46, from
the attack power read off the paladin's white swings.
Hits on level 59 targets are left out (every Holy hit checked takes about 2 less there), and so are the targets the
creature map leaves untyped and Dark Shade, which the map calls Humanoid but Exorcism lands on. The residual on Undead
less the residual on the other typed targets is what the 46 adds; the normalization speed and the attack power offset
cancel in it.

| Hits              | Undead | Other | Undead less other    | Per point of the 46     | The 46 at 0.429 |
| ----------------- | -----: | ----: | -------------------- | ----------------------- | --------------: |
| Holy Strike 10333 |    185 |   104 | −0.69 [−2.76, +1.38] | −0.015 [−0.060, +0.030] |           +19.7 |

A second count of the same strikes, level by level, reads the same on both types: amount less the formula without the
46 is +3.9 on Undead and +3.1 on Humanoids at level 58 (66 and 48 hits), +4.6 and +4.3 at 59 (52 and 45), +6.2 and
+8.5 at 60 (88 and 20), and +11.4 and +11.2 at 61 (18 and 12).

Exorcism, which lands only on Undead, reads 0.995 [0.975, 1.015] of the sim's amount with the 46 at its 0.429 (57
hits) and 1.041 without it. Shieldsworn's strike, 221 plus 0.08 times the 69 spell power, reads 224.2 to 225.0 less
that spell power's share on each Undead target of levels 57 to 60 with 8 or more hits, and 220.3 to 221.1 on each such
Humanoid, about the 0.08 × 46 = 3.7 the bonus adds. Consecration and Holy Shield's block damage take it in this log as well, as EV-0141 and
EV-0143 record.

## Crits of Shieldsworn's strike

| Hit                            |  Hits | Crits | Rate                | Crit amount       |
| ------------------------------ | ----: | ----: | ------------------- | ----------------- |
| Shieldsworn's strike 456544    |   262 |    19 | 7.25% [4.69, 11.05] | 338 to 346, 1.50× |
| Holy Shield on a block 1239570 |   793 |     0 | 0% [0, 0.48]        |                   |
| Consecration's ticks           | 5,481 |   410 | 7.48% [6.81, 8.21]  | 1.54× of the base |

Non-crit strikes deal 225 to 244. The strike's row has the magic defense type and no Can't Crit flag, and the 1.5
times puts its crits on the spell table. Its rate matches the spell crit chance Consecration's ticks show; the interval
is wide enough to hold the paladin's melee crit as well, so the rate alone does not pick the table.
