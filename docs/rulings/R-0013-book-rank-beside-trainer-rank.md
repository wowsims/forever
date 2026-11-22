# R-0013: A rank only a class book teaches registers beside the trainer's top rank

- Skill: wowsims-abilities-auras
- Code: `ui/sim/proto/apl_utils.ts`, `sim/warrior/heroic_strike_cleave.go`, `sim/warrior/battle_shout.go`, `sim/warrior/revenge.go`, `sim/core/buffs/battle_shout.go`, `sim/core/buffs/drivers.go`, `sim/hunter/serpent_sting.go`, `sim/hunter/aspects.go`, `sim/shaman/totems.go`
- Pull request: #736, #746, #794, #1162
- Evidence: not needed: a sim design choice

A rank that only a class book teaches registers beside the trainer's top rank, and the rotation picks the rank by spell
id. Each simple rotation rank toggle defaults off, which casts the trainer's top rank, and the ids swap through
`withSpellIds`, which never swaps a tag -1 id (the raid's copy of a buff); in the rotation editor the spell dropdown
lists both ranks. The warlock presets cast the book rank: Shadow Bolt, Immolate and Corruption. The warrior presets
cast the trainer's top rank, so the simple rotation with its toggles off runs them as they are, and a toggle turned on
swaps in the book rank: Heroic Strike 11567 and 25286 (each with its own queue), Battle Shout 11551 and 25289, and in
the Protection preset, which has no simple rotation, Revenge 11601 and 25288 (one window, read by 5302). The hunter
and shaman presets cast the trainer's top rank too, because the books may not be there at launch:
Serpent Sting 13555 and 25295 (one sting to a hunter), Aspect of the Hawk 14322 and 25296, Strength of Earth Totem 10442
and 25361, Grace of Air Totem 10627 and 25359 (one totem to a slot), and shaman has no simple rotation. The mage, rogue
and druid presets do the same: Frostbolt 10181 and 25304, Fireball 10151 and 25306, Arcane Missiles 10212 and 25345,
Frost Ward 10177 and 28609, Backstab 11281 and 25300, Feint 11303 and 25302, Eviscerate 11300 and 31016, Ferocious Bite
22829 and 31018. A plain SIML name is the trainer's rank and a book rank takes a suffix, which
`sim/core/spelldata/book_ranks.go` decides, since a book rank's game files row states it like a trained one. Multi-Shot
has the one rank, 2643, so its Guide teaches nothing here. The raid's `battle_shout` has four states (off, Rank 6, Rank
7, Rank 7 with the tier 2 set; no Rank 6 with the set), and the warrior's own copies are labelled by rank.
