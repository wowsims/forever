# EV-0104: one Bloody Claws of Attack gives +1 Fist Weapons from its equip spell alone

- Claim: one Bloody Claws of Attack gives +1 Fist Weapons, the base points of its equip spell 1297442.
- Sources: the in-game tooltip and weapon skill of Bloody Claws of Attack 280068; no log
- Game build: 1.60.1.70938, after hotfix push 114060 (the owner's answer in #1129)
- Date: not recorded
- Pull request: #974, corrected in #1129

The game files at that push give Bloody Claws of Attack 280068 no weapon skill item stat; its +1 Fist Weapons is equip
spell 1297442 alone, A_MOD_SKILL on the Unarmed line. The in-game check shows +1. #974 wrote this file up as a tooltip
showing one +1 where item stat 98 and the equip spell both stated it, but the check ran after the hotfix that took item
stat 98 off the item, so it never compared the two.
