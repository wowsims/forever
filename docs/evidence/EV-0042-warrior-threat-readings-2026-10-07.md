# EV-0042: Sunder Armor, Heroic Strike and Cleave threat from the 2026-10-07 captures

- Claim: Sunder Armor Rank 1 makes 34 + 0.05 x attack power threat, Heroic Strike Rank 3 adds a flat 40 over its damage
  and Cleave Rank 2 about 46 split between its targets, each times the stance.
- Sources: threat capture by an addon that recorded the warrior's threat on the target at every threat update, paired
  with the combat log by script; report codes not recorded
- Game build: not recorded
- Date: 2026-10-07
- Pull request: #264

The game logs no threat, so each change in the recorded threat was paired with the combat log. Battle Stance multiplies
by 0.8 and Defensive Stance by 1.3; white hits in the same fights read exactly those.

Sunder Armor Rank 1 (threat effect 34 in the game files), clean updates where Sunder was the only action:

| Stance    | Sheet attack power | Threat per landed Sunder | n   | (34 + 0.05 x AP) x stance |
| --------- | ------------------ | ------------------------ | --- | ------------------------- |
| Battle    | 132                | 32                       | 2   | 32.48                     |
| Battle    | 158                | 33                       | 8   | 33.52                     |
| Battle    | 180                | 34                       | 7   | 34.40                     |
| Defensive | 170                | 55                       | 2   | 55.25                     |

Every landed cast counts the same, including stacks 2 to 5 and refreshes at 5. Parried, dodged and missed casts add
none. A coefficient from 0.049 to 0.053 fits all four rows.

Heroic Strike and Cleave add stance x (damage + k x spell level), with no attack power term:

| Spell                                 | Measured                                                                                    | k    |
| ------------------------------------- | ------------------------------------------------------------------------------------------- | ---- |
| Heroic Strike Rank 3 (spell level 16) | +40 to 40.6 over the damage, steady from 132 to 170 attack power, in both stances           | 2.54 |
| Heroic Strike Rank 4 (spell level 24) | +61                                                                                         | 2.54 |
| Cleave Rank 2 (spell level 30)        | about +46 on one target, about +23 on each of two: the total splits between the targets hit | 1.53 |

Flat threat proportional to spell level is how the game files state Sunder's own threat (34, 75, 117, 158 and 206 at
spell levels 10 to 58, 3.4 to 3.55 per level). The bonus damage of Heroic Strike and Cleave follows no such rule. An
extra threat of about 60 on white crits, whose source is unknown, was left out.
