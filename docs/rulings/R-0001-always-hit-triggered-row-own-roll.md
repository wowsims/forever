# R-0001: Always Hit on a triggered spell stops only that spell's own roll

- Skill: wowsims-abilities-auras
- Code: `sim/core/spelldata/attributes.go`, `AlwaysHits`; `sim/common/shared/shared_utils.go`, `spellDataProcDamageSpell`
- Pull request: #1141
- Evidence: EV-0164

`ATTR_EX_3_ALWAYS_HIT` on a triggered spell stops only that spell's own roll. Judgement of Command's damage 20966
carries the bit and its dummy 20968 does not, so the judgement keeps the dummy's miss roll. A triggered spell the sim
casts as a spell of its own, such as an item's on-hit damage, honours its own bit: Iceblade Hacker's 1298414 and
Warblade of Caer Darrow's 1298499 never miss.
