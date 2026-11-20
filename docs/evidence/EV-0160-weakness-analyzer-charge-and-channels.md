# EV-0160: What spends Weakness Analyzer's charge, inside a channel and out

- Claim: Weakness Analyzer's buff, Analyzing Weaknesses (1291101), drops on the first crit of a direct hit, and an
  Arcane Missiles missile is such a hit: the charge goes as that missile lands, with the channel still running.
  Periodic ticks never spend it, Mind Flay's and Drain Life's included, and neither do the waves of Blizzard, Rain of
  Fire and Hellfire.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every fight of the 142 reports the Reports section
  lists; the mid-channel drop below is `3WQwhXcFVtdHjrkZ` fight 1, source 20
- Game build: 1.60.1.70938 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114281; the hotfix
  each player ran is not recorded
- Date: 2026-11-14 to 2026-11-20
- Pull request: not recorded

## Method

The wearers are every mage and warlock whose gear in the report index lists Weakness Analyzer (272438), 165 in 141
reports, and one Shadow priest in `7fy93LdcwQhvF2gP`. A use is an application of 1291101; a use logged in two
reports (the same character and class, starting within 2 minutes) counts once. Each use runs to its first removal. A
crit is a damage event of the wearer with hit type 2, or 17 (a crit with a partial resist, whose damage reads crit
sized). A drop is credited to the last crit 0 to 40 ms before the removal; a drop with no crit in that window is the
buff running out or the wearer dying. Every other crit while the use ran counts as one that left the buff up. Script
and inputs: `forever-weakness-analyzer-1120/analyze.py` and `wanted_analyzer.json`, output in `out.txt`.

## Counts

| Class   | Uses | Ran out or died |
| ------- | ---: | --------------: |
| Warlock |  363 |             266 |
| Mage    |  241 |              33 |
| Priest  |   12 |               7 |

Hits by what their crits did, every class pooled:

| Hits                                                           | Crits before a drop | Crits with the buff up |
| -------------------------------------------------------------- | ------------------: | ---------------------: |
| Arcane Missiles missiles (10274)                               |                  40 |                     12 |
| Frostbolt, Shadow Bolt, Searing Pain, Arcane Blast, Ice Lance  |                 161 |                     25 |
| Mind Blast, Shadow Word: Death                                 |                   5 |                      0 |
| Mind Flay ticks (18807)                                        |                   0 |                      9 |
| Drain Life ticks (11700)                                       |                   0 |                    104 |
| Blizzard waves (1279949)                                       |                   0 |                    747 |
| Rain of Fire waves (1282385)                                   |                   0 |                    129 |
| Hellfire (11682 and ticks of 11684)                            |                   6 |                  1,919 |
| Corruption, Curse of Agony, Siphon Life, Wrack, Immolate ticks |                   4 |                  2,871 |
| Shadow Word: Pain, Devouring Plague ticks                      |                   0 |                     23 |

None of the 12 missile crits counted with the buff up was followed by the removal more than 40 ms later: each had
another crit of the wearer within 40 ms, and the drop went to that later crit. The 10 crits of Hellfire and the ticks
before a drop are among 4,800 crits of those hits, and fit crits that land beside the one that spent the charge.

In 24 of the 40 drops on a missile crit, more missiles of the same channel landed in the 1.2 s after the drop (one in
13, two in 11). In `3WQwhXcFVtdHjrkZ` fight 1 the second missile of a channel crits, the buff drops 13 ms later and
three more missiles land without it.

## Reports

`1MA3qWLGa8v2hdXP`, `1Tf47h2MvyczZbVK`, `21TnpDC6HBYPVWKz`, `2wPQmVqbdCcZW9gv`, `3Ra7Dtjhc9AXLCzJ`, `3WQwhXcFVtdHjrkZ`, `4VwxHWJ3pgNmvQY2`, `4XMGR6aQ2hd9Km7q`, `4qJWdMANzhgCBHVv`, `4zHJh1aptbByGFnj`, `6VCT9rpGZQzhmRjL`, `6kD7wbm3NaKW8RYP`, `6mpWRxtgXn8JhfkT`, `7LrX43AZapzF18JT`, `7fy93LdcwQhvF2gP`, `81LcPyjXVbm2xGa3`, `82WLwYNzvjB3bXM6`, `8FBqfvLyWTdKxbZh`, `8Gw7m39ZXFz1y2fv`, `8w4hqpXM7fQxJ2BK`, `8zB1PhRQy64jHkXY`, `9R8rhcAwz6V17fpF`, `9Y3Xbk6D28ThMHKd`, `9b71qzjAg6Myt34K`, `9tjaL4MdpWx6rmhT`, `ACTGrmZgQh38MvFf`, `AmWBdq2xcHLv3aYt`, `BX8n2xTycaKr4CGL`, `BkTfV9yAZJX8P7wF`, `D46Xm8ncR9WbTFrL`, `D6NFqLY8bVnkG4dB`, `DJvC7NkGn3HpPYhK`, `DZCc2gtwhW3FBKyx`, `DdKN3mXj6gH9yvpc`, `DjtAKCp2dN6TJvHR`, `DqGVcxR1QTKrMaP9`, `FDyCQTmfrdqW2hxa`, `FHA7TKpb3ncBqgMC`, `Fz2wJmfGKPnvTy9h`, `G7X1CkT8z6PrfjWM`, `GcbndBCyf1N9m7YT`, `Gx9NyLPCjwDbV3qc`, `HB7yCLD3wvVx4fj9`, `HRYM8C3jxmPGdbf2`, `HYmf1vTW6hVDPytK`, `HZP6h9AtRNK4n3Ck`, `HaDkgnNL6wTWrjhG`, `HfaAyZNt6rKWV7pY`, `HyfjpPNRaFhQ328n`, `JBxW16RqGXMPYH3d`, `JaC1tgnqAV8zN6LM`, `JbQyngR8NWXKardc`, `KG62VXF3a97CxtAg`, `KVLgwAr9bB1dJyNR`, `KXG2knhLR1jJqPft`, `KZYm38HcJx9wP6LQ`, `LQ9zMDqpmknBgAX8`, `LYw76XG1CFHQNdJz`, `Lr12yNYCdD3a7HQW`, `MBRmC1TnXq86k2Nb`, `MRJHKBWt2TCp17PG`, `N6B2APcTWnzHvfbq`, `N6cHzwkKGrAjqZfM`, `Q6FN7ALfWMhxBKvH`, `QNfaZ9MTnWpFrqb4`, `QXWcwaGZvCKjJFM4`, `Qyf4CGR1hgczwvp9`, `RmqHtxk2T8PLZ4QM`, `VKfYJ82cMTktxQzL`, `VMBZcafW2JLbT6dK`, `VYTnz7v9mbtgf3pF`, `WYTNrpf3VQZt7Xqw`, `WqpkNKafVtYjFBDg`, `Wy6TdamAk4ZxJ8MC`, `X4wDxLvF3YgfyKW7`, `X8PH2CG1hMkcfxYJ`, `XDQFCAzk7YHhmKvG`, `Xfzx4v1aMRnmbBph`, `XgQYbwGBZcHLKRxV`, `Xk697aGVLz4FWcDd`, `XvhwZD31Vz46JWG8`, `XzPyHxjZDvgMR9rA`, `Y8L1VRGqHjKJpbCM`, `Y9AvDptPmZHNMGjq`, `YGR7VPK1CQgc2Dnm`, `YHk7yC4L9qhmzc2f`, `YpQAGtmMa3b47XRk`, `Z1d98D2CFxk3cyh6`, `ZHpKVnhPDGyaTrvw`, `ZM2PjyvY7WDRAgzr`, `ZkgVjD9fcKFP8bnL`, `ZwvyzH9bxRN3ghCX`, `Zz3YcQk69P7N8rAn`, `b7y2YrKZh4kQ9cPv`, `bGTptRMWJyg4N7HQ`, `bKJcjWDRzvHpVmwa`, `bQw4YVFNJZBt2THG`, `c8mjK4TdzV2M7XLF`, `cKG7yDArxfqhJPp6`, `cav7VZrmPbNW3y29`, `cbTt1zxajFfNVPDA`, `dAG4yD6aXF7x9c2t`, `f1WqtFkXGn4pBzT6`, `fMLcHzyZvt7QTCmg`, `fjPTxNBrbkJ6v3Q1`, `gGBmQx9nC18ZD3d2`, `ghXzrk6YBFQ24p9a`, `hMv4dJgY3pnyVXTk`, `hZ3AnX6dp2TQavFx`, `hZLykw9dRnYbtvKP`, `hZPWBMjdQgVJp4AK`, `hfJzLYkNGarbVvWK`, `jr46gDtLzB8dfXnK`, `k1taHbwL76Cnxv4m`, `k3ZGwNbF9W7DgzLQ`, `kNwn6HXKVhCByZr3`, `kph1Df7CqFVxQ2PW`, `kqd9ZJbhaFKNLGnV`, `m76CKntQZWLYhTMg`, `mR9k6PLdgMKqvXJy`, `mTw3kz4cgqrv8JpA`, `mjrQyAZVTtxJvpN2`, `n3vFdHjhwJDC1fA2`, `nR1WGzY8pJKDxhQj`, `p7nAc2zT6qQbmCZB`, `pHqbZzVRTWy3N2nv`, `praQ172yvmdZNbgn`, `qCZwWn2pJ3zVGKkj`, `qZwYtNQLPG1xzkJW`, `qa8wcv4jLYZpGARK`, `qtG184bKfvyr3VXZ`, `rYhcR3zbKmCk4N6V`, `rkf3atgDPHQjXqTV`, `vdD1pLX6tkMamzqn`, `vwmWKDGFc14rzVJY`, `wrzB6GtjPaqCKJZW`, `xRK37aCvtVfj26kL`, `xkN3nXgd8RpHybzT`, `xvJWqXQh1nakDH36`, `yWFhcNBab8wYMQtG`, `ywLTngctk7WB24v1`, `zp9yVrtfTNZJCv7g`
