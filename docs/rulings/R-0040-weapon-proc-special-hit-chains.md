# R-0040: A weapon proc's strike a main-hand white swing cast rolls weapon procs once more

- Skill: wowsims-procs
- Code: `sim/common/shared/shared_utils.go`, `weaponProcStrikeMask`, `procDamageHandler` and `applySpellDataDamageProc`;
  `sim/common/shared/spell_data_debuff_proc.go`, `attachDebuffStrike`
- Pull request: #352, #1088
- Evidence: EV-0145

A weapon proc with melee or ranged DefenseType is a special hit, whether it deals a direct hit, applies
damage over time alone or applies a debuff alone. It rolls the hit table of its DefenseType, and what it applies lands
only where that roll does.

A melee strike a weapon proc casts off a main-hand white swing of the weapon is heard by weapon procs, its own weapon
included, by combat enchants and by oils. One it casts off anything else is heard by none of them: an ability, a
seal's strike, a Windfury Totem hit, an off-hand hit, the paw's swing of a druid in Cat, Bear or Dire Bear Form, or
another strike. A strike a white swing cast therefore rolls its weapon once more, and the strike that roll casts rolls
nothing.

A strike with a cooldown of its own is not cast again while it runs. It brings no second, and a hit that
lands inside the cooldown casts none either, a Windfury Totem hit off the same swing included: Toofbreak 1320269
states 100 ms.

A magic weapon proc's spell is a spell hit (R-0051), equip and aura procs keep the `SpellFlagProc` gate, and Windfury
never hears a strike.

The measurement covers the weapon's own proc in the main hand and in the off hand, and Fiery Weapon in the main hand.
The same flag decides for the other combat enchants and for oils, which no log has counted. A ranged strike rolls
ranged weapon procs whatever cast it, an extra attack's swing is a white swing, and the strike each tick of Ravager's
Bladestorm casts rolls the other weapon procs; none of the three is measured. A form other than Cat, Bear and Dire
Bear Form is not measured either, and its swing counts as the weapon's.
