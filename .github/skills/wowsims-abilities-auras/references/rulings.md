# Rulings

Each ruling is its own file in `docs/rulings/` (`docs/rulings/README.md`): the rule, the code that follows it, the pull
request that settled it and its evidence. This index lists the ones this skill owns.

## Rolls and outcomes

- **R-0001** Always Hit on a triggered spell stops only that spell's own roll: `docs/rulings/R-0001-always-hit-triggered-row-own-roll.md`
- **R-0002** A special can be dodged, parried or blocked only when it lacks No Active Defense: `docs/rulings/R-0002-special-dodge-follows-no-active-defense.md`
- **R-0003** A spell with magic DefenseType and no Always Hit rolls a magic hit even without damage: `docs/rulings/R-0003-magic-defense-spell-rolls-hit.md`
- **R-0004** Intercept's stun is its own spell and level plus 3 targets are immune to it: `docs/rulings/R-0004-intercept-stun-own-spell-immunity.md`
- **R-0005** Consecration ticks every second and Serpent Sting is a full ranged ability: `docs/rulings/R-0005-consecration-ticks-serpent-sting-ranged.md`
- **R-0006** DoTs do not snapshot and every tick uses the stats of the moment it lands: `docs/rulings/R-0006-dots-do-not-snapshot.md`
- **R-0007** Deep Wounds bleeds a share of weapon damage and its ticks never crit: `docs/rulings/R-0007-deep-wounds-bleed-no-tick-crit.md`

## Reading the row

- **R-0008** A talent's trait curve applies with the sign the game files state: `docs/rulings/R-0008-talent-curve-sign-follows-game-files.md`
- **R-0009** Seal of Righteousness keeps hand coefficients over the game files' zero: `docs/rulings/R-0009-seal-of-righteousness-tooltip-coefficients.md`
- **R-0010** Shield Bash and Shield Slam register as main-hand melee specials: `docs/rulings/R-0010-shield-specials-main-hand-special.md`
- **R-0011** A hand-built melee or ranged special states its DefenseType: `docs/rulings/R-0011-hand-built-special-states-defensetype.md`

## Registering

- **R-0012** Casters register every rank only when the implementation requires it: `docs/rulings/R-0012-casters-register-ranks-when-required.md`
- **R-0013** A rank only a class book teaches registers beside the trainer's top rank: `docs/rulings/R-0013-book-rank-beside-trainer-rank.md`
- **R-0014** A buff a class book raises is a rank select with 0 as off: `docs/rulings/R-0014-buff-book-rank-select.md`
- **R-0015** Devouring Plague is learnable by every race and Inner Focus restarts its cooldown: `docs/rulings/R-0015-devouring-plague-inner-focus-cooldown.md`
- **R-0016** A school hit talent shows on the sheet only by opting in: `docs/rulings/R-0016-school-hit-talent-opt-in-sheet.md`
- **R-0017** Weaponmaster covers a mace or staff in either hand without doubling: `docs/rulings/R-0017-weaponmaster-mace-staff-either-hand.md`

## Forms

- **R-0018** A shift into Bear Form empties the rage bar before Furor's rage: `docs/rulings/R-0018-bear-form-shift-empties-rage.md`
- **R-0019** A druid's change of form keeps the current health percentage: `docs/rulings/R-0019-form-change-keeps-health-percentage.md`
