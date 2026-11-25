# EV-0209: Proc rate of Shell Launcher Shotgun from the 2026-11-24 sweep

- Claim: Shell Launcher Shotgun (item 13146) casts Flaming Shell 29647 at 3.75 procs per minute, priced on the weapon
  base speed of each ranged hit that rolls it. The rate is observed: no lone-unit rate checks it.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the 27 reports and fights §Reports lists, from every
  wearer stream on disk for the weapon
- Game build: 1.60.1.70938 to 1.60.1.71091 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 12 of the 27 reports start after hotfix push 114313 (2026-11-20 23:47 UTC), one of them after 2026-11-24 01:00
  UTC. The hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-24
- Pull request: #1231

## Method

The count is EV-0183 §Method, run again by the automated sweep of 2026-11-24 14:50 over every wearer stream on disk for
the weapon. Every wearer is a hunter, so the paladin rule does not apply. The rate is the result's `all_hits`, matched on
item and `effect_spell`, rounded to the nearest 0.05 at full precision; the result's lone-unit rates set no rate.

Scripts and inputs, in `forever-sweep-auto`: the run `runs/2026-11-24-1450`, whose `results.json` (`analyze.py`
version 19, cache `cache_v18`) holds the count; the split by push 114313 comes from that result's `runs_detail` and the
reports' start times in `meta.db`, each run's rate being the result's rate times its procs over its expected procs.

## Shell Launcher Shotgun (item 13146, spell 29647): 3.75, observed

4,519 procs on 31,398 rolling hits, 22 wearers, 27 runs, in the ranged slot: all hits 3.75 [3.65, 3.87], a chance of
14.4% per hit at the weapon's 2.3 s base speed. Of the hits, 19,163 are Auto Shots, 9,995 are abilities and 2,240 are
Serpent Sting landings. The rate is all hits. Observed: no lone-unit rate checks it. The proc is a missile that lands a
median 567 ms after its hit, 98th percentile 1,003 ms, window [-20, 1013] ms (983 isolated hits), so the lone Auto Shot
and special rates read on a window past 400 ms, and none checks the all-hits rate. Lone Auto Shots, which set no rate:
3.61 [3.45, 3.77], 1,115 procs on 8,047 lone units, 21 wearers. Specials, which set no rate: 3.90 [3.65, 4.16], 572
procs on 3,822 lone units, 22 wearers; Multi-Shot gives 4.00 (436 procs on 2,845), Serpent Sting's landing 4.27 (98 on
599), Arcane Shot 2.72 (38 on 364) and Aimed Shot 0 on 14. Dropping the swapped stretch (9 active minutes, 0.7% of hits,
in 1 run) moved the read from 3.73 to 3.75; 1 run logged in two reports counts once. Reports from 2026-11-13 to
2026-11-24.

The rate holds on both sides of push 114313, which changed the weapon procs of Execute and Mutilate and touches no
hunter ability: before it 3.79 [3.65, 3.93], 2,842 procs on 19,580 hits, 10 wearers, 15 runs; from it on 3.70 [3.52,
3.88], 1,677 procs on 11,818 hits, 12 wearers, 12 runs, one of them in a report of 2026-11-24 10:15 UTC, after the
2026-11-23 hotfix, with 44 procs on 318 hits. The intervals are Poisson on the procs.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Hunter |      22 | 4,519 | 31,398 | 3.75 |

## Reports

6rgZCnpQMGw78AW2 1; ADxF6TNdhRb2amJZ 1; NdbCqzpGA8Yr9ZfL 1, 3-15; 16kB9HrdYTtXcm8M 1-7; 2VpbrfHatzjZFkGP 1-2;
4BAHkN9fDaPCm8wT 1; 4qdnCVHXazcF9LNB 1-16; Dand1ft9zbqKcPXA 1; G3avtwxqpmc9ARFN 1-51; Gkb3rKHF6V9gz4pR 1-5;
KqyGLhrPVQ8pwWan 5-10; MqJFdTLDaVH9yWj7 34, 36-38; NW7Fx8QTHLD2v3fY 1-9; P1kwFWxZjfGDRqT8 1; YBDx8dkm1GCN9JT3 1-4;
Yzv1c7dBnLNhyXQT 1-9; bNwmGdPhzVQt9HDq 1-3, 5; cbkDQfXNPGj3tnWT 1; famzQN3D9FPdpqrY 1-16; hMdxA2bRGqr8ncvV 1-8;
q1dxCGHyak78KcRL 1-4; qBzJM1XDbcLymVvC 1-2; tVm8K9DfdRawv7nH 1-2; vGdq3ZrKyMAPaTJD 5-8; xvHfnhgdWpcB8Cy2 1-2;
yCT8NWQBMz9k6dYh 1-9; yT7QzCRkAD8F12bL 10-30.
