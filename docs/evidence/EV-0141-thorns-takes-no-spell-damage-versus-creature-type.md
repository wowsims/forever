# EV-0141: Thorns takes no spell damage against a creature type

- Claim: an item's spell damage against a creature type (Courier's Clasp 279833, +10 against Humanoids) reaches every
  damage spell and tick of a Balance druid's except Thorns, whose hits take the druid's other spell power at Thorns'
  coefficient and none of the bonus; an Elemental shaman's Lightning Shield orbs take it like the shaman's other spells.
  A Protection paladin's own Retribution Aura takes none of it either, while the same paladin's Consecration and Holy
  Shield take all of it on the same targets; the sim applies the rule to every damage shield, as it does the target's
  flat school bonus.
- Sources: combat log, a Tauren Balance druid, levels 54 to 56, Courier's Clasp worn all day, 2026-11-16 06:53 to 17:40
  and 22:21 to 22:48 (two logs); combat log, an Orc Elemental shaman, levels 50 and 51, Courier's Clasp worn from 19:11,
  2026-11-16 17:01 to 2026-11-17 00:37; combat log, a level 60 Protection paladin in Stratholme wearing Courier's Clasp
  in every combatant info line, 2026-11-13 14:26 to 16:12. Times are the player's local time.
- Game build: 1.60.1.70938 for the druid's logs; the shaman's log states only 1.60.1. The hotfix each log ran on, and
  whether the player relogged after it, is not recorded.
- Date: 2026-11-16
- Pull request: #1078

## Method

The druid's logs carry no creature type, so a target counts as a Humanoid by its Insect Swarm rank 4 ticks: per target
name and level, the most common offset of a tick's base amount from floor((25 + 0.158 × spell power) × the druid's
multipliers), with spell power read from the druid's own log field, which leaves the clasp out. Most targets read 0,
and 38 target cells read +2, the share 10 spell power gives a tick there. For each row kind the extra spell power every
hit on those targets allows is the intersection of the intervals its base amount allows; the same intersection on the
targets reading 0 is the control. A tick or a Thorns hit fits a reading when floor((row amount + coefficient × (spell
power + extra)) × multipliers) equals its base amount, which is the amount before a crit. Thorns rank 6 is 18 + 0.06 ×
spell power.

## Druid, targets reading +2

| Row                             | Hits | Extra spell power all hits allow | Fits + 0 | Fits + 10 |
| ------------------------------- | ---: | -------------------------------- | -------: | --------: |
| Insect Swarm rank 4 tick        |  771 | 9.40 to 11.30                    |        0 |       771 |
| Moonfire rank 9 tick            |  517 | 9.04 to 10.79                    |        0 |       517 |
| Entangling Roots rank 5 tick    |   34 | 3.03 to 17.89                    |        0 |        34 |
| Nature's Grasp roots 19971 tick |    2 | 2.64 to 16.93                    |        0 |         2 |
| Hurricane rank 2 hit            |  632 | 4.80 to 11.39                    |      581 |       632 |
| Thorns rank 6                   |  313 | −1.25 to 1.32                    |      313 |       100 |

On the targets reading 0 the intervals hold 0 and leave out 10: Moonfire rank 9 ticks −0.50 to 2.47 (727 hits),
Hurricane rank 2 −0.08 to 0.78 (1,173), Thorns rank 6 −0.39 to 1.32 (568). Insect Swarm's interval there is empty
(1,147 ticks): no single extra value fits every tick. One Thorns hit adds 0.6 for 10 spell power
before the multipliers and the floor, so a single hit often fits both readings; the 313 hits together leave no room for
the bonus. No Thorns rank 5 hit landed on a target reading +2. Direct hits agree: Wrath rank 8 non-crits on the +2 targets
read 1.0241 ± 0.0033 of the expected amount at the logged spell power and 1.0006 ± 0.0033 with 10 more (197 hits),
Starfire rank 5 1.0189 and 1.0035 (105).

## Shaman

Hits on Humanoids after the clasp went on read 10.8 spell power more than the log's spell power field over 326 hits
with no other group debuff on the target: Earth Shock 10.3, Flame Shock ticks 10.6, Frost Shock 9.6, Lightning Bolt
11.3 and Lightning Shield's orbs 10.5. Beasts, elementals, golems and spiders read 0 ± 1.5, and Humanoids before the
clasp went on read 1.00 of the expected amount.

## Paladin

The paladin's own Retribution Aura hits fit floor(20 + 0.06 × spell power) at the logged spell power. On Humanoids, 5 of
the 206 hits where +0 and +10 spell power predict different amounts fit +10 (2.4%, 95% interval 1.0 to 5.6%); on Undead,
the control, 9 of 750 do. On the same targets in the same log, Consecration's spell power ticks fit +10 on 266 of 266
discriminating ticks and Holy Shield 1239570 on 168 of 168, so the clasp counts there and the targets are Humanoids.
Creature types come from the NPC tooltips, since the game files list only companions.

