# Traps in a hand-built spell

The defect classes of a spell built beside its row instead of from it. The cases reviews found for each are in
`.github/skills/wowsims-review/references/defects.md`, under the section named after the class.

- **Hand code the resolvers already build.** A duration, cooldown, cost, refund, tick count or tick the resolver fills
  is never typed beside it, and an `OnTick` equal to `DotConfig`'s default is redundant;
  `go run ./tools/spelldata -config '<call>' -package <class>` shows what the resolver fills. Cases: §Hand code the
  resolvers already build.
- **What scales.** A hand-built amount that scales other than the row says, by weapon bonus, combo point or stack: the
  rule is `wowsims-core-mechanics` `## What scales`. Cases: §What scales.
- **How many targets.** A hand-built area effect or target count that differs from the row's: the rule is
  `wowsims-core-mechanics` `## How many targets`. Cases: §How many targets.
- **What the row forbids.** A hand-built config that loses the row's requirements or restates them: the rule is
  `wowsims-abilities-auras` §An ability (a hand read: `docs/spell_data/building-an-ability.md`), core's check
  `wowsims-core-mechanics` `## What the row forbids`. Cases: §What the row forbids.
