# R-0055: A summon appears where its row puts it and walks in from beside its owner as a pet does

- Skill: wowsims-consumables-items
- Code: `sim/common/shared/item_guardians.go`, `summonsBesideCaster`, `newGuardian`, `SummonGuardians` and
  `Guardian.ExecuteCustomRotation`; `sim/hunter/summon_hawk.go`, the hawk's `StartsInMeleeRange`
- Pull request: #1250
- Evidence: EV-0217

A summon puts its creature where its `E_SUMMON` effect's first implicit target says. A row that names a spot beside
the caster (`TARGET_DEST_CASTER` and its front, back, side, random and radius forms) brings the creature up as far
from the target as its owner stands at that moment, and the creature walks in at 7 yd/s the way a pet does:
into melee range when it swings, otherwise to its spell's range. It casts its spell only standing still within that
range, since the spell is a hardcast, and swings only once it stands in melee range. Every other row, one that puts
the creature at the target or at a spot the player picks, brings it up in melee range of the enemy.

Book of the Dead, Dragon's Call, the Violet Sorcerer's Vestments and every other use item the game files hold put
their creature beside the caster; Parachute-Priest Pager puts it at a spot the player picks, and the hunter's hawk at
the target (EV-0127). EV-0217 measured the walk for Book of the Dead's skeleton; the other creatures follow the same
rule unmeasured.
