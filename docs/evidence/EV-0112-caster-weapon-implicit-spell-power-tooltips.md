# EV-0112: Caster weapons' implicit spell power matches their in-game tooltips

- Claim: a weapon flagged for implicit spell power or healing power carries the value the item level budget and the
  "of Magic" and "of Healing" allocations give, as its in-game tooltip shows.
- Sources: in-game item tooltips; no log
- Game build: not recorded
- Date: not recorded
- Pull request: #302

Every one of the 150 changed weapons matches its in-game tooltip exactly, healing and spell damage. Simone's
Cultivating Hammer 22380, for example, shows "Increases healing done by up to 139 and damage done by up to 46" in game.
Coldspire Staff (271215, 26 spell power in game) and Searing Dagger (273003, 18) show implicit spell power without
either flag.
