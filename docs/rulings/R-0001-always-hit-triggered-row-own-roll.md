# R-0001: Always Hit on a triggered row stops only that row's own roll

- Skill: wowsims-abilities-auras
- Code: `sim/core/spelldata/attributes.go`, `AlwaysHits`
- Pull request: #1141
- Evidence: EV-0164

`ATTR_EX_3_ALWAYS_HIT` on a triggered row stops only that row's own roll, so Judgement of Command keeps the miss roll of
its damage though its dummy 20968 carries the bit. A row the sim casts as a spell of its own, such as an item's on-hit
damage, honours its own bit: Iceblade Hacker's 1298414 and Warblade of Caer Darrow's 1298499 never miss.
