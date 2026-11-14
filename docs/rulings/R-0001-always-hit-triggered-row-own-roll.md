# R-0001: Always Hit on a triggered row stops only that row's own roll

- Skill: wowsims-abilities-auras
- Code: `sim/core/spelldata/attributes.go`, `AlwaysHits`
- Pull request: none
- Evidence: pending

`ATTR_EX_3_ALWAYS_HIT` on a triggered row stops only that row's own roll; the sim skips the miss only when the row the
player casts carries the bit, so Judgement of Command keeps the miss roll of its dummy 20968.
