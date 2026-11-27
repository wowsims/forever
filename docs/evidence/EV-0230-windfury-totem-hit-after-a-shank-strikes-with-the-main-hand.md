# EV-0230: A Windfury Totem hit after a Shank strikes with the main-hand weapon

- Claim: the Windfury Totem hit a rogue's Shank sets off deals the main-hand weapon's damage, as the hit after a
  main-hand swing or ability does, though Shank strikes with the off hand.
- Sources: a combat log of two dual-wielding rogues in a group, levels not recorded: one with a 2.0 s dagger in the main
  hand and a 1.7 s axe in the off hand (Barman Shanker 12791, Bone Slicing Hatchet 18737), one with two 1.8 s daggers
  (Glacial Blade 19099, The Shadowfoot Stabber 24222)
- Game build: 1.60.1.71091 (the 1.8 s daggers, 2026-11-25 14:33 to 15:07 UTC) and 1.60.1.71139 (the dagger and axe,
  2026-11-25 21:21 to 23:27 UTC), after hotfix push 114399, relogged
- Date: 2026-11-25
- Pull request: #1280

Each Windfury Totem hit that is neither a crit nor a glance is matched to the rogue's last hit on any target 5 to 60 ms
before it: a Shank, a main-hand white swing or a main-hand ability. Its damage before armor is divided by the median of
the rogue's plain main-hand white swings within 60 s of it. The log does not name the hand of a white swing, so swings
are told apart by their damage less the attack power part (Otsu's split). The interval is the mean's 95% normal interval.

| Weapons             | Hit before        | Totem hits | Totem hit / main-hand white swing |
| ------------------- | ----------------- | ---------: | --------------------------------- |
| dagger 2.0, axe 1.7 | Shank             |         23 | 1.00 [0.97, 1.03]                 |
| dagger 2.0, axe 1.7 | main-hand swing   |        145 | 1.01 [0.99, 1.02]                 |
| dagger 2.0, axe 1.7 | main-hand ability |         88 | 1.00 [0.98, 1.01]                 |
| two daggers 1.8     | Shank             |         17 | 0.98 [0.94, 1.02]                 |
| two daggers 1.8     | main-hand swing   |         67 | 0.99 [0.97, 1.00]                 |
| two daggers 1.8     | main-hand ability |         50 | 1.01 [0.99, 1.03]                 |

The off-hand weapon's full damage, its average plus the same attack power part over its own speed, would read about
0.88 against the main-hand swing with the dagger and axe, so that pair settles the claim; with the two daggers it would
read about 0.95, inside the Shank row's interval.
