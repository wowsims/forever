package overrides

// Item summons that are not guardians, by the spell whose E_SUMMON names the creature, each with why. The game files
// hold no SummonProperties table to tell a companion, a sign or an area effect from a creature that fights, so these
// stay out of item_guardians_auto_gen.go by hand. A totem, which summons into a totem slot, stays out with no entry.
var NotGuardianSummons = map[int32]string{
	12766:   "Cloud of Chained Essence: an area trigger whose damage is its periodic aura 1300240, not a creature's",
	17567:   "Blood Parrot: a companion with no attack",
	1294258: "Ward of Zum'rah: its zombie adds below 1 DPS, which the cane's own gear notice says (EV-0124)",
	1306267: "Traveler's Mark: a sign that carries a message",
}
