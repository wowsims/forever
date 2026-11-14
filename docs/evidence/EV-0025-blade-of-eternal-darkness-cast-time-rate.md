# EV-0025: Blade of Eternal Darkness procs on cast time

- Claim: Blade of Eternal Darkness 17780 procs 1.8 times a minute on the cast time of the cast that launched each landed spell hit, at least 1.5 s, and its Engulfing Shadows mana comes with every proc, a full resist included.
- Sources: Warcraft Logs Nq3C7bVGnvmyh2Pw, cvRTFb1qhGBKNkCD and JK1z96cCrj28tpBV, Forever dungeon reports of one mage
- Game build: not recorded
- Date: 2026-11-14
- Pull request: #1007, with the mana reading first written up in #1001

The blade is confirmed in the main hand by combatant info in 12 fights; the other fights with procs are inferred from
those procs. There are 554 procs, each tied to the event at its timestamp.

| Trigger                                                                                                   | Procs |
| --------------------------------------------------------------------------------------------------------- | ----: |
| AoE casts (Arcane Explosion 192, Flamestrike first hit 45, Frost Nova 11, Cone of Cold 6)                 |   256 |
| Single-target casts (Arcane Wrath 45, Fire Blast 34, Arcane Blast 32, Fireball 18, Pyroblast 5, Scorch 5) |   145 |
| Flamestrike's periodic spell 1279990                                                                      |    75 |
| Arcane Missiles missiles                                                                                  |    58 |
| Counterspell, Arcane Missiles channel starts, mixed                                                       |    20 |
| DoT ticks (0 of 488), Ignite (0 of 377), wand Shoot (0 of 132)                                            |     0 |

One roll per target, not per landing. On short-tier AoE the per-landing rate rises with the targets hit: 3.8% at 1
target, 14.0% at 3, 20.0% at 4, 44.3% at 9 or more. Rolling 4.7% per target predicts 4.7%, 13.4%, 17.5% and 44.8%, where
one roll per landing predicts a flat 17.5%. Log-likelihood is -1420 per target against -1578 per landing.

Cast time, not weapon speed. Single-target landings proc 4.4% on instants and triggered spells (93 of 2119), 3.4% on 1.5 s
casts (29 of 842) and 9.3% on casts of about 2.95 s (60 of 644). Pyroblast cut to 1.48 s by Heating Up procs 2.1% (3 of
140). Arcane Wrath with Presence of Mind procs 4.0% (2 of 50), against 9.0% hardcast. The fit is 1.80 per minute
(1.65 to 1.96). Against one flat chance per target the likelihood-ratio statistic is 35.2 (p about 3e-9). The blade's
weapon speed is 1.5 s, so a manager on the main hand's speed would price every hit at 4.5%; the 3 s tier rules that out.
Fitted per tier, the rate is 4.5% (4.1 to 5.0) on instants, triggered spells and 1.5 s casts and 9.1% (7.5 to 10.9) on 3 s
casts.

Harmful spells that deal no damage also rolled in game (Counterspell 5 of 92, Arcane Missiles channel starts 6 of 219),
11 of the 554 procs; the sim leaves them out. Misses are undetermined, since the sample has 2 true misses or full
resists, so only landed hits roll.

Caveats: one player and mage spells only. Only two cast-time levels exist (about 1.5 s and 2.95 s), so a two-tier rule
fits as well as the proportional one. Haste is unmeasured. One low-level fight (Scarlet Monastery Library) has 2 procs
against 7.5 expected.

Mana. The same three reports (seven fights) log 506 Engulfing Shadows mana gains of 104 and 499 damage events. Each mana
gain is logged at the proc, 0.1 to 1.4 s before the projectile's damage (median 0.3 s), every damage event has one, and
both full resists (cvRTFb1qhGBKNkCD fight 30 at 13026605, JK1z96cCrj28tpBV fight 2 at 3178478) restored the mana. No
plain miss was logged.
