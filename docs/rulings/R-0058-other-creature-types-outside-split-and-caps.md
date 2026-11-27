# R-0058: Targets outside a row's creature types count toward no split, MaxTargets or AoE cap

- Skill: wowsims-core-mechanics
- Code: `sim/common/shared/shared_utils.go`, `calcMultiTargetDamage`
- Pull request: #1213
- Evidence: EV-0231

A multi-target effect whose row names creature types (TargetCreatureType) reaches only the enemies of those types. An
enemy of another type takes no roll, and it never counts toward the split of a row that splits its damage, toward the
row's MaxTargets or toward the AoE cap: the effect divides, caps and scales over the enemies of its types alone.
Stratholme Holy Water's 17291 names Undead, so among 40 targets with one Undead it deals that one its whole roll.

EV-0231 shows the no-roll part: 17291 logs no event of any kind on an enemy that is not Undead. The split and cap part
is not measured. Twelve uses reached 21 to 34 Undead, but the count recorded no per-hit damage by the number of
targets reached.

The owner ruled it on 2026-11-27 in #1213.
