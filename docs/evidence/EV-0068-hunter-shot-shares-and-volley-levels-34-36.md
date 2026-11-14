# EV-0068: Hunter shot attack power shares, the normalised shot bonus and Volley's waves at levels 34 and 36

- Claim: Multi-Shot and Aimed Shot hit 9 harder than the rows state, Arcane Shot takes 15% of ranged attack power, each Serpent Sting tick 5%, Explosive Trap's initial hit 10% with no share on its ticks, and Volley's waves land a tick apart from the cast.
- Sources: two combat logs of a Beast Mastery hunter at levels 34 and 36 (Redridge on 2026-11-07, Uldaman on 2026-11-08)
- Game build: not recorded
- Date: 2026-11-07 and 2026-11-08
- Pull request: #909; the Volley reading is in its commit 47a1687366

The game files state no attack power share on these rows and no bonus on the normalised shots. The shares below are the ones that fit the logs.

- Multi-Shot and Aimed Shot take 9 a hit on top of the rows' normalised weapon damage, across the whole weapon roll. The 9 was measured against Auto Shot on the same 2.8 speed bow with Jagged Arrows, which hits for its own damage. The 9 was read at levels 34 and 36, where a flat bonus and one that grows with ranged attack power predict the same hits, so the logs cannot tell them apart. Sniper Shot, the third normalised shot, was not in the logs and takes none.
- Arcane Shot takes 15% of ranged attack power.
- Each Serpent Sting tick takes 5% of ranged attack power, read at each tick. Spell power adds nothing.
- Explosive Trap's initial hit takes 10% of ranged attack power. Its ticks take no ranged attack power and roll no hit: none of 1451 ticks missed.

The hunter pet's Immolation Trap share is not measured; the site that gives it keeps a TODO.

Volley's waves land a tick of the cast row's periodic dummy apart, from the cast: 0.980 s at Wind Blessed's 2% cast
speed, and Quick Shots and Rapid Fire left that spacing alone. The pull request and its commits do not record that the
missile's flight delays the channel's aura but not the waves, so that part has no count.
