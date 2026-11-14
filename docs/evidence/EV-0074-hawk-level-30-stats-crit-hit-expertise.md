# EV-0074: A hawk's level 30 stats, and its crit, hit and expertise from the tamed pet's

- Claim: at level 30 a hawk has 5% of the creature health and no armor or attack power of its own, and takes 80% of the
  owner's ranged attack power and 30% of the owner's armor; its crit is half the tamed pet's, and its melee hit is the
  tamed pet's.
- Sources: four level 30 combat logs (2862 hawk swings, owner ranged attack power 407 to 1023), and logs with Ferocity
  5/5 and with none; report codes not recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #592, whose body is empty; the readings are in its commits 81651da664 and b89878e6f9

## Level 30

Read off 2862 hawk swings in four level 30 logs, a hawk has 5% of the creature health, no armor or attack
power of its own, 80% of the owner's ranged attack power and 30% of the owner's armor. A swing before armor was 3.3 to
6.7 plus 0.0179 a point of the owner's ranged attack power, the same line in every log.

## Crit

4340 swings with Ferocity 5/5 read 10 plus half the owner's crit against targets 2 levels below to 3 above it. Logs
with no Ferocity read half the tamed pet's.

## Hit

1% hit took a hawk from 1 miss in 32 swings to none in 515 against a target 8 levels below it.

The code takes the higher of the tamed pet's melee and ranged crit, with none from Agility, and the tamed pet's
expertise along with its hit; the commits record no count for those parts.
