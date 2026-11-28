# EV-0111: Creature type item stats match the item tooltips, Seal of the Dawn's +78 among them

- Claim: the creature type attack power and spell damage item stats read as the item tooltips show them, and Seal of the
  Dawn 13209, whose +78 Attack Power against Undead comes from its equip spell 23930 alone, shows one +78.
- Sources: in-game item tooltips; no log
- Game build: not recorded
- Date: not recorded
- Pull request: #983

Every type name and value was checked against the item tooltips, for example Twitching Eye "+18 Attack Power against
Humanoids", Seal of the Dawn +78 against Undead, Courier's Clasp 10 spell damage against Humanoids and Enduring Chain
Vest 127 health per 5. Seal of the Dawn (23930) and Rune of the Dawn (24198, +46 spell damage against Undead) state no
creature type item stat; their equip spell holds the bonus, and the tooltip shows it once. No count is recorded.
