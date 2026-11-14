# EV-0110: A random suffix that reduces caster damage gives its weapon the reduced range

- Claim: a weapon carrying a random suffix whose bonus list has an ITEM_BONUS_REDUCED_CASTER_DPS row deals the reduced
  caster share of its damage table in game.
- Sources: the in-game damage of Brutal War Axe (3210) with "of the Physician" (12911), with two images on the pull request; no log
- Game build: not recorded
- Date: not recorded
- Pull request: #403

In game, Brutal War Axe (3210) "of the Physician" (12911) deals 43 to 66 damage, 15.1 DPS, where the sim gave it 58 to
88 with every suffix. The weapon-tier caster suffix lists (of the Sorcerer, Elder, Channeler, Physician, Prophet,
Hierophant, Magic, Healing) carry an ItemBonus row of Type 54 in the game files, the reduction the generator applies to
base items flagged REDUCED_CASTER_DPS (EV-0097). The lower lists on armor carry no such row.
