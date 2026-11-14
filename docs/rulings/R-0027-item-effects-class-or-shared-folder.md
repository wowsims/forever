# R-0027: A hand-written item effect lives in the folder of the classes that can use it

- Skill: wowsims-consumables-items
- Code: `tools/database/overrides.go`, `ApplyItemSetClasses`
- Pull request: #316, #539
- Evidence: not needed: a repo process rule

A hand-written item effect only one class can use lives in `sim/<class>/item_<kind>.go` (sets, librams, idols, totems,
trinkets, weapons); class-agnostic items live in `sim/common/forever/items_*.go`. A set the game files open to every
class is shared, in `sim/common/forever/items_sets.go` with `agent.GetCharacter()`, and so are the dungeon sets;
Naxxramas tier 3 armor is class-locked through `ItemSetClasses` and coded in the class folders.
