# EV-0185: Elixirs count each stat once, and the raid buffs of Intellect, Spirit, Stamina and crit overwrite them

- Claim: with several elixirs up, each stat counts only the strongest elixir's amount, and only one elixir's 2% crit
  counts. Arcane Intellect overwrites the elixirs' Intellect, Divine Spirit their Spirit, Power Word: Fortitude their
  Stamina, and Moonkin Aura and Leader of the Pack their crit. Scroll of Intellect does not stack with Arcane Brilliance and stacks with the elixirs. Juju
  Power removes Scroll of Strength's buff and blocks a new one. Mark of the Wild stacks with the elixirs. Greater Mageblood Elixir stacks with Mana Spring Totem, which raises no MP5 on the
  character sheet.
- Sources: checked in game by the owner, no log
- Game build: the live build when it was reported on 2026-11-23 (the repo then recorded 1.60.1.71037, push 114332);
  whether the game was relogged after that hotfix is not recorded
- Date: not recorded; reported on 2026-11-23
- Pull request: #1185
- Later count: the raid buffs, the scroll, Mark of the Wild and Mana Spring Totem, §Count of the 2026-11-23 in-game check
- Later count: Power Word: Fortitude, §Count of the 2026-11-23 Fortitude check
- Later count: Juju Power and Scroll of Strength, §Count of the 2026-11-23 Juju Power check

With Elixir of the Grizzly (250351) up beside another elixir that grants 2% crit, Grizzly's buff tooltip reads "your
chance to critically hit is increased by 0%" while its Strength still reads 25.

The owner states that a stat two elixirs grant counts only at the strongest one's amount: Elixir of Ferocity's
(250350) Agility is overwritten by Elixir of the Mongoose's (13452), and its Strength by Juju Power's (12451) or Elixir
of Greater Strength's (9206).

## Count of the 2026-11-23 in-game check

- Sources: a player's in-game check, which the owner relayed with screenshots; no log
- Game build: the live build when it was reported on 2026-11-23; the hotfix it ran on and whether the game was
  relogged are not recorded
- Date: not recorded; reported on 2026-11-23
- Pull request: #1185

The player read the character sheet's Intellect, Spirit and MP5 and the buff tooltips with one buff added at a time.

| Up                                                        | Character sheet           | Tooltip                                          |
| --------------------------------------------------------- | ------------------------- | ------------------------------------------------ |
| Arcane Intellect and Divine Spirit                        | 262 Intellect, 245 Spirit |                                                  |
| and Elixir of Sages (1250941: 25 Spirit, 2% crit)         | 262 Intellect, 245 Spirit | Sages: "Your spirit is increased by 0"           |
| and Elixir of Cunning (1250918: 25 Agility, 25 Intellect) | 262 Intellect             | Cunning still reads 25 Intellect                 |
| no Mana Spring Totem                                      | 195 MP5                   |                                                  |
| Mana Spring Totem                                         | 195 MP5                   | Mana Spring: "gives you 10 mana every 5 seconds" |
| Greater Mageblood Elixir                                  | 215 MP5                   |                                                  |

The owner's summary of the check: an elixir's crit does not stack with Moonkin Aura or Leader of the Pack, the
Intellect elixirs do not stack with Arcane Intellect, the Spirit elixirs do not stack with Divine Spirit, and Scroll of
Intellect does not stack with Arcane Brilliance but does stack with the elixirs. Greater Mageblood Elixir appears to
stack with Mana Spring Totem. The owner reports that Mark of the Wild stacks fully with the elixirs.

The owner's ruling R-0054 counts Scroll of Intellect with the elixirs all the same, the strongest amount alone.

## Count of the 2026-11-23 Fortitude check

- Sources: checked in game by the owner, no log
- Game build: the live build when it was reported on 2026-11-23; the hotfix it ran on and whether the game was
  relogged are not recorded
- Date: not recorded; reported on 2026-11-23
- Pull request: #1185

The owner reports that Power Word: Fortitude and the Stamina of Elixir of Brute Force (13453) do not stack. What was
read in game is not recorded.

## Count of the 2026-11-23 Juju Power check

- Sources: checked in game by the owner, no log
- Game build: the live build when it was reported on 2026-11-23; the hotfix it ran on and whether the game was
  relogged are not recorded
- Date: not recorded; reported on 2026-11-23
- Pull request: #TBD

The owner reports that Scroll of Strength and Juju Power (12451) do not stack: applying Juju Power removes the scroll's
Strength buff, and a new Scroll of Strength cannot be applied while Juju Power is up. Which rank of the scroll was used
is not recorded.
