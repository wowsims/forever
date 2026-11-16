# EV-0135: Crusader procs per minute on the enchanted weapon's hits

- Claim: Enchant Weapon - Crusader (1900) gives Holy Strength (20007) about 1 time a minute on the enchanted weapon's
  landed melee hits, priced on that weapon's base speed, in either hand.
- Sources: Warcraft Logs, every player whose combatant info shows enchant 1900 on a weapon, in QxGpKz7knDvjWmYb,
  tQfhj2JbzZkYXFvx, VYBjgJx8aZtRhmpy, f3hyv2GtFW4b8wJV, KG62VXF3a97CxtAg, Rcanq4T9FhDg1JMy, zmDY4LVX7N8vh9aR,
  RkcV84XGnCzKWYbg, vPHb3a6c1KhknLrY, h3VmpYZ4Gt6xNJC7, rtpQk3BXbDnNGCj6, XLvj723CgZTJGy4h, Z78jbgTL9V1Yfahq,
  Qb1yWxa8vHCGL7Yr, vGtQdg49pCh7zx2H, dvkzFtPH3WR9C4Km, ak7AJXQCbD2YHT4n, GRnMHmK69BXfWVTj, mbZcgPavhdG7fKDW,
  JRQqNL6DkvWKYdPB, HxCQat38vmjnkpb6, 8zB1PhRQy64jHkXY, LR9WfMTdbr3ZNBFH, gmKvNqHFJLZXfdnP, CDKVcP6ghZJtv8AW,
  hVtxNMYzk4fAJGnR, 7fz9Kk8rY6jyQAbL, DW9YkdvyrjtMBQnf, qVvZxP3gNp12KWyb, zFXxNAn69HBGRaqg, YVWgj6ra27PvmNnx,
  1JLq46yg7wYNhan9, v92hQ4WGwV18yjMY and NdTXJzkWHLZrVQYm; all fights, a run left out when its weapons change between
  fights
- Game build: 1.60.1.70938; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-16
- Pull request: #1059

## Method

A proc is an application or refresh of Holy Strength (20007) on the wearer. The hits are the enchanted hands' landed,
non-periodic melee hits: on the main hand its white swings and every melee-defense hit (abilities and triggered hits
such as seals), on the off hand its white swings. A dual wielder's white swings are split between the hands by their
damage before crits. The procs expected at 1 per minute are each enchanted hand's hits times its weapon's base speed
over 60, so a wearer with the enchant on both hands counts both; the rate is the procs over that.

Druids are left out: in Cat Form they swing the paw, not the enchanted staff, and their two runs read 0.11 per minute
on the staff's speed. Runs that appear in two reports with the same counts are counted once.

## Rate

| Wearers' setup                | Runs | Procs / expected at 1 PPM | Rate     |
| ----------------------------- | ---- | ------------------------- | -------- |
| Every run                     | 27   | 1969 / 2068.5             | 0.95 PPM |
| Main hand, no off-hand weapon | 5    | 284 / 262.4               | 1.08 PPM |
| Main hand, dual wield         | 13   | 1188 / 1290.3             | 0.92 PPM |
| Both hands                    | 7    | 441 / 463.5               | 0.95 PPM |
| Off hand only                 | 2    | 56 / 52.4                 | 1.07 PPM |

The 95% interval of the pooled rate from 2,000 draws of whole runs is 0.91 to 1.00. By class: warriors 0.92, rogues
0.92, shamans 1.00, paladins 1.08, hunters 0.97.
