# R-0051: An item's magic proc is a spell hit to the listeners that hear procs

- Skill: wowsims-procs
- Code: `sim/common/shared/shared_utils.go`, `spellDataProcDamageSpell` and `attachProcDamage`;
  `sim/paladin/holy_shield.go`, Shieldsworn's strike; `sim/mage/combustion.go`, the charges' listener
- Pull request: not yet opened
- Evidence: EV-0149, EV-0161

The spell an item, enchant or set proc casts, built from its row by the shared builders (`NewSpellDataDamageProc`, a set
bonus proc, a consumable's combat spell, an absorb's break damage), takes `ProcMaskSpellDamage` when its row has the
magic defense type and deals its damage as a spell of its own, whether a weapon proc or an equip proc cast it. A
listener on spell damage hears it only when the listener can proc from procs, since the spell keeps `SpellFlagProc`:
Revelation, Insight and the paladin's Vengeance roll on it in game. A row with no defense type of its own, and damage an
aura deals off its own trigger (`A_PROC_TRIGGER_DAMAGE`: a shield spike, Holy Shield on a block), keep `ProcMaskEmpty`,
and none of those listeners rolls on them. A heal a proc casts takes `ProcMaskSpellHealing` and is heard the same way.

A hand-written listener that tests `ProcMaskSpellDamage` itself, rather than through a `ProcTrigger`, also leaves out
`SpellFlagProc` spells when its row cannot proc from procs, as Combustion's does.

Not settled: whether an imbue's hit (`ProcMaskSpellDamageProc`) belongs with these; an on-use item's spell, an
explosive's and a hand-written item spell (`NewProcDamageEffect`, Force Reactive Disk), which keep the mask their code
gives them; and a modifier keyed on the mask (Shatter, Fire Power), which reaches an item's magic proc with no
measurement behind it.
