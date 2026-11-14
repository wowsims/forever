# EV-0078: Demon armor includes Demon Armor and is fixed at the summon

- Claim: a demon inherits 35% of the warlock's full armor, Demon Armor included, and its armor is fixed when it is summoned while its attack power and spell damage follow the warlock's.
- Sources: screenshots of a level 30 Imp and Sayaad with a naked Undead level 30 warlock (talents `-0050003`: Demonic Embrace 5/5, Fel Vitality 3/3), Blood Pact rank 3 and Demon Armor rank 1 up; a Sayaad combat log, report code not recorded
- Game build: not recorded
- Date: 2026-10-15
- Pull request: #443

Imp. Warlock armor 272 (62 from Agility and 210 Demon Armor), Imp armor 490.

| Reading                                  | Imp's own armor | Share of the creature armor (1,212) | Level 20 Imp (no Demon Armor) |
| ---------------------------------------- | --------------- | ----------------------------------- | ----------------------------- |
| Demon Armor inherited (35% × 272 = 95.2) | 394.8           | 0.326                               | 0.325                         |
| Demon Armor left out                     | 468.3           | 0.386                               |                               |

With Demon Armor counted, the Imp's armor share is the same at levels 20 and 30, as its health (0.359 and 0.360) and weapon (0.514 and 0.512) shares are. The same holds for the Sayaad: own armor 765 at level 30 is 0.631 of the creature armor, as at level 20 (0.630), so the level 30 share of 0.692 read earlier does not hold.

Fixed at the summon. In a Sayaad log Demon Armor fell off (warlock armor 543 to 333) and the Sayaad kept its armor of 955 through 65 swings and casts, while its attack power followed the warlock's live (101 to 102 at its next Lash of Pain, 1.8 s after the warlock's spell damage went 30 to 36, with no re-summon). One screenshot (warlock 453 with Demon Armor, Sayaad 850) is 765 + 0.35 × (453 less 210) = 850.05 with the Sayaad summoned before Demon Armor went up.
