# R-0051: An item's magic proc is a spell hit to the listeners that hear procs

- Skill: wowsims-procs
- Code: `sim/common/shared/shared_utils.go`, `spellDataProcDamageSpell` and `attachProcDamage`;
  `sim/paladin/holy_shield.go`, Shieldsworn's strike; `sim/mage/combustion.go`, the charges' listener;
  `sim/rogue/poisons.go`, `sim/shaman/weapon_imbues.go`, `sim/core/buffs/drivers.go`, the poisons' and imbues' hits
- Pull request: #1135
- Evidence: EV-0149, EV-0161, EV-0226

The spell an item, enchant or set proc casts, built from its row by the shared builders (`NewSpellDataDamageProc`, a set
bonus proc, a consumable's combat spell, an absorb's break damage), takes `ProcMaskSpellDamage` when its row has the
magic defense type and deals its damage as a spell of its own, whether a weapon proc or an equip proc cast it. A
listener on spell damage hears it only when the listener can proc from procs, since the spell keeps `SpellFlagProc`:
Revelation, Insight and the paladin's Vengeance roll on it in game. A row with no defense type of its own, such as a
shield spike's, keeps `ProcMaskEmpty`, and so does Holy Shield's damage on a block, which `holy_shield.go` builds; none
of those listeners rolls on them. A heal a proc casts takes `ProcMaskSpellHealing` and is heard the same way.

A weapon imbue's or a poison's hit (Flametongue Weapon and Totem, Frostbrand, Lightning Shield's orb, rogue poisons,
Shadowguard, Touch of Weakness) is the same: `ProcMaskSpellDamage` with `SpellFlagProc`, since none of their rows
carries Not A Proc. Lightning Shield's orb rolls Revelation (EV-0149); Flametongue Weapon's attack rolls it only at low
levels, and none from about level 32 up, as its spell level of 1 gives (EV-0149). Flametongue Totem's crits give
Vengeance as Fiery Weapon's do (EV-0226). Both listeners can proc from procs. Bolt's Charged
Eyeball, which cannot, hears no hit of Instant Poison, Lightning Shield, Shadowguard or Touch of Weakness; the rest
follow from their rows.

A hand-written listener that tests `ProcMaskSpellDamage` itself, rather than through a `ProcTrigger`, also leaves out
`SpellFlagProc` spells when its row cannot proc from procs, as Combustion's does.

Not settled: an on-use item's spell, an explosive's and a hand-written item spell (`NewProcDamageEffect`, Force Reactive
Disk), which keep the mask their code gives them; a modifier keyed on the mask (Shatter, Fire Power), which reaches an
item's magic proc or an imbue's hit with no measurement behind it; and whether Insight (8216) rolls on an imbue's or a
poison's hit, which the sim lets it do from its row's Can Proc From Procs (EV-0161 reads 8 of 1,093 lone Flametongue and
Frostbrand hits).
