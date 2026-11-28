# EV-0231: Stratholme Holy Water logs damage on Undead targets alone

- Claim: Stratholme Holy Water's 17291, whose row names Undead (TargetCreatureType 32), logs damage on Undead
  enemies alone and no event of any kind on another enemy, neither a miss nor an immune; one use reaches up to 34
  Undead.
- Sources: Warcraft Logs, the 40 most recent of the 822 reports the sweep cache lists with 17291, whole reports, all
  fights: 4BhNKZ7zg6FLMkbn, NkBMTXFfzv1bZ3am, xTB6LXhAZP9GCw8j, YcFCX8r4pq1GBgwa, CBLzaq2VRMKZQ1JF, AJHbTt24NV31gr9x,
  cad1zxpmL6HrX94A, jQkz1MF3D6cAYmKv, KcnLAh8prDW37j2R, z2pxdRyNHVjQ8CKf, dJXGPaVLK2hFY71k, c3VjQZryJtqnxW4N,
  pXFZBWkhAYcxR9z7, rvPn71AMTwajKNdG, RWFN9TLVXKP8mCxp, Y2Tt1pzvkwA6aLZg, jr3d8GNpHwnavtPf, B8Md6DvZNXkbrxC1,
  z1qW38NkZF7QRw49, DfZCt67zGcT8rMXv, apdtqMV9gk6XNAfC, pc93Jt2YZGw1Tn7V, avnFMLCYq89B3KXw, 2rympNxvqj34QwPa,
  qhWbg68XDkZHALcn, dYWtjRGx4ck1KZmB, dK3hZrp79mJwjY1F, 8jZFPCGWvHQ26knA, Xxdpgy7KNcz1VkwT, wFgzpDZ6rxMT13Ry,
  R63P8NcWZxwt79Hv, 1G3NPXBkY7Vvx6jW, QgM8cAkp4KB6aPtY, vx8J2kaqzdVmHTtf, 1wH4rtmC6GpRnQfz, xTgvNMzQnc6wWCH1,
  vLqFydDAQkKpmjH4, a9ZnYNCQjXhPMcV2, gKrahydfZ7mJN9AD, RNgTjDm3FhCAznK2
- Game build: 1.60.1.71139 by the report dates (71139 from 2026-11-25 19:14 UTC); the hotfix each player ran is not
  recorded
- Date: 2026-11-26, the reports starting from 07:23 to 15:32 UTC
- Pull request: #1278

## Method

The fetch takes every event of ability 17291 in each report from friendly sources, the Warcraft Logs default: casts,
damage, misses and immunes. A use is a 17291 cast, and the targets it reached are the distinct targets of the same
source's 17291 events from the cast to 1,500 ms after it. A target counts as Undead by its NPC name; every name the
count met is a Stratholme or Scholomance Undead (Mindless Undead, skeletons, ghouls, banshees, gargoyles, spectres,
Baron Rivendare). The logs do not show whether another enemy stood in a use's area: a use that skips one logs the
same as a use with none in reach.

Scripts and outputs: `requests/forever-holywater-1126` in the log parser's folder. `fetch.mjs` fetches the events
and `look.py` counts them.

## Numbers

| Events                                                  | Count |
| ------------------------------------------------------- | ----: |
| Uses (17291 casts)                                      |   114 |
| Damage events                                           |   803 |
| Hits                                                    |   785 |
| Crits                                                   |    14 |
| Immune, on Undead (Wailing Banshee 2, Spectral Tutor 2) |     4 |
| Events of any kind on a non-Undead enemy                |     0 |
| Undead names among the targets                          |    35 |

| Targets reached by one use | Uses |
| -------------------------- | ---: |
| 0                          |    8 |
| 1 to 5                     |   71 |
| 6 to 10                    |   19 |
| 11 to 20                   |    4 |
| 21 to 34                   |   12 |

The most Undead one use reached is 34, in two uses.
