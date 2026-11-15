# EV-0030: A Blizzard cast made with charges up spends no Fingers of Frost charge

- Claim: Blizzard, a channel, spends no Fingers of Frost charge, and its ticks grant charges.
- Sources: Warcraft Logs 7nzxFa4Y9bVmPNdv (no Frostbite) and jbM98hHnqwPCxNca (with Frostbite)
- Game build: not recorded
- Date: not recorded
- Pull request: #343

Seven Blizzard casts made with charges up spent none. With Frostbite the gain lands 0 to 1 ms before or between a tick's
damage lines, rolled per target, and never at the channel's cast. Raw logs without Frostbite put most Frostbolt procs 0 to
3 ms after the hit and a Cone of Cold proc in the same millisecond; the few procs about 100 ms after the hit are ruled
outliers. The Frostbolt order with Frostbite and the spending rules for direct damage spells and Arcane Missiles are the
owner's in-game observations, with no counts recorded.
