# EV-0171: Ice Lance's spell power coefficient

- Claim: Ice Lance Rank 6 (1240047) deals its game files damage (145 + 0.9 per level past 56, 136 to 161 at level 60)
  plus 0.115 of the spell power on gear. Frozen targets take 4 times the damage, and Ice Barrier (+5%), Frigid Power
  (+15%) and Ice Shards' crit damage (×2) apply to it as their rows state.
- Sources: combat logs of one level 60 Frost mage with StatLedger, the player's local time; geared casts from three log
  files, 2026-11-20 01:54 to 2026-11-21 07:36, casts at 0 spell power, 2026-11-21 08:40 to 08:45, and geared casts from
  three more log files (`WoWCombatLog-112226_225034`, `112326_045021`, `112326_195011`), 2026-11-22 23:26 to
  2026-11-24 08:07, with the same talents as the first set
- Game build: 1.60.1.70959 for the first geared set (StatLedger's logins read it up to 2026-11-21 06:18) and
  1.60.1.71037 from the 08:35 login on. The first geared set and the 0 spell power casts ran on the latest hotfix at
  the time, relogged; the second geared set on an older hotfix, which one not recorded
- Date: 2026-11-20 to 2026-11-24
- Pull request: #1150; the second geared set, the 0.115 reading and the Emblem's gain, #1199

## Method

Every Ice Lance cast by the mage is paired with its damage event on the same target. Spell power is the caster's on the
cast line. The damage is `baseAmount`, the amount before the crit and before any partial resist. Fingers of Frost counts
as frozen when one of its charges is removed within 50 ms of the cast; a Frost Nova, Freeze or Frostbite on the target
counts too. Ice Barrier, Frigid Power and Draconic Infused Emblem are read from the mage's auras at the cast. Each hit
is divided by the frozen, Ice Barrier and Frigid Power multipliers; hits on targets above level 60 are left out, since
those take a level factor.

## Counts

| Set                           | Hits | Spell power | Per hit, before modifiers | Range     |
| ----------------------------- | ---: | ----------: | ------------------------- | --------- |
| 0 spell power, no buffs       |   28 |           0 | 147.1 ± 1.4               | 136 – 159 |
| first geared, no Emblem buff  |  255 |   203 – 234 | 173.3 ± 0.5 (at 228.5)    | 161 – 186 |
| second geared, no Emblem buff |  185 |   218 – 257 | 176.4 ± 0.6 (at 242.4)    | 161 – 193 |
| second geared, Emblem buff    |   70 |   253 – 292 | 179.4 ± 0.9 (at 278.9)    | 166 – 194 |

The 0 spell power set fits the game files band (148.6 mean). Against it, the first geared set reads
(173.3 − 147.1) / 228.5 = 0.115 and the second (176.4 − 147.1) / 242.4 = 0.121; against the game files mean they read
0.108 and 0.115. The two readings against the 0 spell power set are good to about ±0.006, most of it from its 28
hits; the two against the game files mean are tighter, since that mean has no error of its own. The sim takes 0.115,
the middle of the four. 1/7 sits more than three standard errors above every reading.

The modifiers hold within 1.5% across the first geared set: 36 unfrozen hits with neither buff read 171.9, 696.6 frozen
(×4.05), 798.6 frozen with Frigid Power and 836.6 frozen with both. A crit deals 2.00 times `baseAmount` (134 crits,
2.00 to 2.01). Frost Nova without Fingers of Frost reads ×4 on its three hits.

## Draconic Infused Emblem

The Emblem's buff (1318930) adds 35 spell damage while it lasts, 35 more against Dragonkin. Its gain is the damage with
the buff up less the damage with it down in the same gear, each gear setup weighted by the inverse of its variance;
hits on Plagued Hatchlings, whelps and the only target whose name reads as Dragonkin, are left out. Ice Lance gains
+1.4 ± 1.0 over three gear setups in the first geared set and +2.8 ± 0.9 over five in the second, +2.2 ± 0.7 together,
where 0.115 gives +4.0. With the hatchlings kept the sets read +1.9 ± 1.1 and +3.0 ± 0.95, +2.5 ± 0.7 together. No
Frostbolt under the Emblem hit a hatchling, so their type is not confirmed from the logs. Frostbolt Rank 10 in the
second set gains +29.8 ± 1.6, where its 0.814 gives +28.5, on 354 hits with neither Ice Barrier nor Frigid Power, 122
of them under the Emblem.

Either way the two sets together sit three or more standard errors above no gain and two to three below the full
+4.0, so how much of the buff reaches Ice Lance is not settled. The sim adds all of it.
