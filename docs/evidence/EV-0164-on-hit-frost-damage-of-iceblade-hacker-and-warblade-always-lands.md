# EV-0164: The on-hit Frost damage of Iceblade Hacker and Warblade of Caer Darrow always lands and crits as melee

- Claim: Iceblade Hacker's 1298414 and Warblade of Caer Darrow's 1298499, cast by a proc on each landed hit, never miss
  and are never dodged, parried or blocked, as their rows' Always Hit states; they crit at the wearer's melee crit
  chance for twice the damage.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every wearer stream on disk for the two weapons; the
  reports and fights each section lists
- Game build: 1.60.1.70862 to 1.60.1.70959 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-10 to 2026-11-20
- Pull request: #1141, b50c27c9f5, 299aa8e1aa

## Method

Each wearer's own damage events in the fights where the weapon sits in the main hand, with a report's repeated events
counted once and periodic ticks left out. Every event of the proc spell is counted by its hit type: a crit is hit type 2
or 17 (a crit with a partial resist), landed is every event but an immune target. The wearer's melee specials (Sinister
Strike, Raptor Strike, Mongoose Bite, Wing Clip, Strider Kick, Cleave, Heroic Strike, Revenge, Shield Slam, Eviscerate,
Holy Strike and others) give the crit chance to compare, counted the same way over their landed hits. The travel time
is the delay from the wearer's last hit on the same target, on events 300 to 500 ms after one.

Iceblade Hacker counts the rogue, hunter and warrior wearers. Warblade of Caer Darrow is worn only by paladins in these
logs; its outcome counts check the rule and set no rate.

## Counts

| Spell                           | Wearers              | Events | Miss, dodge, parry, block | Partial resist | Immune | Crit / landed          | Specials' crit        |
| ------------------------------- | -------------------- | -----: | ------------------------: | -------------: | -----: | ---------------------- | --------------------- |
| Iceblade Hacker 1298414         | 13 runs, non-paladin | 12,210 |                         0 |            932 |    148 | 18.0% (2,166 / 12,062) | 18.4% (1,171 / 6,349) |
| Warblade of Caer Darrow 1298499 | 11 runs, paladin     | 18,990 |                         0 |          2,186 |      7 | 17.5% (3,313 / 18,983) | 17.3% (534 / 3,088)   |

A crit deals twice a hit: 82 against 41 for Iceblade Hacker without modifiers, 73 to 74 against 37 under Defensive
Stance. The damage lands a median 418 ms (Iceblade Hacker) and 412 ms (Warblade of Caer Darrow) after its hit,
98th percentile 437 and 436 ms, which the rows' 0.4 s minimum travel time gives.

The two warrior wearers of Iceblade Hacker read the same per report: 12.5% crit on 1,032 landed against 10.8% on 648
landed specials (B4wWqM2TbRmdQy7Y), and 17.2% on 1,782 against 16.0% on 594 (rTKYMd4yDvaL6Gn8).

## Reports and fights

Iceblade Hacker: A1yxMTmKatkf3HRJ 15; B4wWqM2TbRmdQy7Y 1; DNc6QALhxmGbCH3Y 1; DyMmJxc4t9pw6PAz 1; JQwbjnVYyGcP4rg8 3, 6;
KYmFPtyJ3gkzHfVM 102, 106, 111, 118, 120, 122, 127-128, 130; Kphk63jtr1wb9NyL 1; L6pXTqJAGPwD2tHc 3;
fZdz8MyCHKFkrRwB 1; gw8fpa1YVkv4WP32 1; rTKYMd4yDvaL6Gn8 1; xWDkmwXKrRcdM7Jv 1; xyGf9QY4dwT8JHMp 5.

Warblade of Caer Darrow: CgXG9bvFH6jJVZKk 1, 12, 15; DfTMrcFAzq9ayQKt 3; JQwbjnVYyGcP4rg8 1, 3, 6, 9, 13, 15, 17;
PayThRGFLKmbNcnp 7, 11; Q3XYxAvj7yDzp4c8 1, 11, 167, 184, 189, 195, 200, 203, 206, 209, 212, 215, 217;
dnDg7j9kRw6WtbML 52, 55, 58, 61, 64, 67-68; fxLMQjdb82cCHvmr 1; hZwALnTyCM9fzVrX 1; khfKmZYaDxgBn24t 1;
rQpg6YfW7cLRNAVM 1, 5; watv9ZMA2q3jGXgH 1, 8, 25.
