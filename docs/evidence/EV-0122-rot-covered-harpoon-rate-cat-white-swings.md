# EV-0122: Rot-Covered Harpoon's rate from a cat druid's white swings

- Claim: Rot-Covered Harpoon's Rot rolls 1.45 times a minute on Cat Form white swings, priced on the 1.0 s paw. Landed
  white swings roll it; Cat Form specials did not.
- Sources: two local combat logs, WoWCombatLog-111426_205747 and WoWCombatLog-111526_065850: one druid in Cat Form
  and Dire Bear Form, levels 42 and 43, the harpoon in the main hand at level 43.
- Game build: 1.60.1.70938. Which hotfix push the session ran on and whether the player relogged is not recorded. The
  item updates committed on 2026-11-14 and 2026-11-15, through push 114102, change neither item 274963 nor spell 1295744.
- Date: 2026-11-14 23:42 to 2026-11-15 07:58, the player's local time
- Pull request: not yet opened

The harpoon is in the main hand from 23:42:07. A proc is one application or refresh of Rot 1295744 on a creature: 11 of
the 12 in the logs. The twelfth landed on a player after a Dire Bear Form swing and is left out. Each Rot lands in the
same millisecond as a landed white swing of the wearer on its target (12 of 12, 0 to 1 ms after the swing), so the
denominator is landed Cat Form white swings on creatures with the harpoon worn. PPM = chance × 60 / 1.0, the paw's
speed, since the sim prices a feral's weapon proc on the paw's speed; on the harpoon's 2.5 s it reads 0.57 [0.32, 1.01].
Intervals are 95% Wilson intervals. The rate is rounded to the nearest 0.05 step, as EV-0119 rounds its rates.

| Hits on creatures       | Procs / landed hits | Chance             | PPM on the paw [95%] |
| ----------------------- | ------------------- | ------------------ | -------------------- |
| Cat Form white swings   | 11 / 463            | 2.38% [1.33, 4.20] | 1.43 [0.80, 2.52]    |
| Cat Form specials       | 0 / 196             | 0%                 | –                    |
| Dire Bear Form white    | 0 / 8               | 0% [0, 32.4]       | –                    |
| Cat white, Nov 14 night | 3 / 28              | 10.7%              | –                    |
| Cat white, Nov 15 day   | 8 / 435             | 1.84%              | –                    |

Cat Form specials rolled no Rot: 4.7 were expected at the white swings' chance, and 0 has a chance of 0.009. EV-0119
records the same for Bigger Rock on Bigger Stick. The two sessions disagree, and 2 procs a minute (3.33% a swing) is not
ruled out: 11 or fewer of 463 has a chance of 0.15 (exact one-sided binomial). Example applications: 00:29:18 on
Tazzik, 07:05:00 on a Twilight Champion, 07:58:13 on a Twilight Subjugator.

Rot ticks 7 times, 1 s apart. Every tick on a creature that did not crit dealt 24 (45 of 45), the row's amount, against
the 25 of the tooltip. 9 of the 60 ticks crit, for 36 (×1.5). Rot is not a bleed: Claw (3 of 3) and Rake (2 of 2) on a
target with only Rot up deal the damage they deal without Rend and Tear.

The counts come from the local proc research scripts and tables under `feral-bear/scripts/` and `feral-bear/tables/`.
