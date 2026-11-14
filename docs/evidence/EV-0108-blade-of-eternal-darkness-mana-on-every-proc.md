# EV-0108: Blade of Eternal Darkness restores its mana with every proc, a full resist included

- Claim: Engulfing Shadows' 104 mana comes to the wearer with every proc of Blade of Eternal Darkness 17780, whether the
  damage lands or is fully resisted.
- Sources: Warcraft Logs Nq3C7bVGnvmyh2Pw, cvRTFb1qhGBKNkCD and JK1z96cCrj28tpBV, Forever dungeon reports of one mage,
  seven fights
- Game build: not recorded
- Date: not recorded
- Pull request: #1001

The three reports log 506 Engulfing Shadows mana gains of 104 and 499 damage events. Each mana gain is logged at the
proc, 0.1 to 1.4 s before the projectile's damage (median 0.3 s), every damage event has one, and both full resists
(cvRTFb1qhGBKNkCD fight 30 at 13026605, JK1z96cCrj28tpBV fight 2 at 3178478) restored the mana. No plain miss was
logged.
