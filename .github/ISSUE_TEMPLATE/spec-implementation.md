---
name: Spec implementation
about: Track implementing one spec, one checkbox per talent and per ability
title: '[Spec] Implement <Spec> <Class> for Forever'
labels: ''
assignees: ''
---

## Description

Everything needed to implement **<Spec> <Class>** for Forever: Go backend logic for abilities, talents,
gear interactions and simulation settings, and the frontend components for configuring the spec and
displaying its output. One checkbox per talent and per ability family.

Datamined from game build `<build>` (`wowf`), trait tree `<id>`: **N ability families**
and **N talents** after filtering. Counts and IDs are a starting point to verify against the game files,
not gospel.

<!--
Generate the lists below rather than typing them. See
.github/skills/wowsims-spec-implementation/SKILL.md - it names the extract, the filters and the traps.
Say which content you excluded and why, so a re-run that produces different counts is explainable.
-->

Season of Discovery content is excluded. <List what the build still carries that is not live here.>

## Talents

### <Tab> (N)

- [ ] **<Talent>** — N ranks, tier N — <what it does>

## Abilities

### <Tree> (N)

- [ ] <Ability> (`<rootSpellId>`) — N ranks, N rage, Ns CD

## Backend Tasks

- [ ] Research and data gathering for Forever abilities and talents
- [ ] Generate the spell data tables for this class from the Forever game files
- [ ] Identify any spells the extract found with no name or description
- [ ] Implement core abilities and rotations
- [ ] Implement the talent tree and choices
- [ ] Integrate Forever gear effects (tier sets)
- [ ] Update and verify stat scaling and interactions
- [ ] Implement buff and debuff interactions
- [ ] Add configuration options to simulation settings (optional)
- [ ] Develop and refine the APL
- [ ] Write unit and integration tests
- [ ] Run the `wowsims-review` pass over the change and fix what it finds before asking for review

## Frontend Tasks

- [ ] Update the configuration UI
- [ ] Display spec-specific simulation output, for new resources (optional)
- [ ] Check and update talent tree visualisation and selection
- [ ] Register presets for P1
- [ ] Publish as Alpha
