# EV-0185: A stat two elixirs grant counts only the strongest one, and the crit only one elixir's

- Claim: with several elixirs up, each stat counts only the strongest elixir's amount, and only one elixir's 2% crit
  counts.
- Sources: checked in game by the owner, no log
- Game build: the live build when it was reported on 2026-11-23 (the repo then recorded 1.60.1.71037, push 114332);
  whether the game was relogged after that hotfix is not recorded
- Date: not recorded; reported on 2026-11-23
- Pull request: #1185

With Elixir of the Grizzly (250351) up beside another elixir that grants 2% crit, Grizzly's buff tooltip reads "your
chance to critically hit is increased by 0%" while its Strength still reads 25.

The owner states that a stat two elixirs grant counts only at the strongest one's amount: Elixir of Ferocity's
(250350) Agility is overwritten by Elixir of the Mongoose's (13452), and its Strength by Juju Power's (12451) or Elixir
of Greater Strength's (9206).
