# EV-0055: Thunder Clap rolls spell hit and spell crit

- Claim: Thunder Clap is Physical but Magic in SpellCategories; it rolls on the spell hit table and crits on spell crit
  chance for 1.5 times, and with no base spell crit a warrior without Totem of Wrath sees no crit.
- Sources: Warcraft Logs Fresh (fresh.warcraftlogs.com), 6 protection warriors; report codes not recorded. These are not
  Forever logs.
- Game build: not Forever
- Date: 2026-09-14
- Pull request: none found; the claim came in with commit 9df41474e8

The logs show full resists next to armor mitigation on Thunder Clap. Without Totem of Wrath they show no crit on it: 0 of
799 landed hits.
