package printer

import (
	"fmt"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// What the printer says about an effect whose type or aura no branch below words. The literal beside
// it is then the whole answer.
const unrecognised = "unrecognised shape"

func effectLines(s *spelldata.Spell) []Line {
	out := []Line{}
	for i := range s.Effects {
		e := &s.Effects[i]
		human := humanise(s, e)
		if human == "" {
			human = unrecognised
		}
		out = append(out, Line{Human: human, Literal: literal(e, i+1)})
	}
	return out
}

// The amount at the level the sim prices a cast at, which is Effect.Mean().
func value(e *spelldata.Effect) float64 {
	return e.Mean()
}

// The amount as the row states it: one number, or the ends of the roll where the row states a
// spread.
func amount(e *spelldata.Effect) string {
	if e.Variance == 0 {
		return number(value(e))
	}
	return number(e.Min()) + "–" + number(e.Max())
}

// A percentage the game files state as an integer, which is what Percent() divides by 100.
func percent(e *spelldata.Effect) string {
	return number(value(e)) + "%"
}

func signedPercent(e *spelldata.Effect) string {
	return signed(value(e)) + "%"
}

// Rage and energy sit on the game files' 0-1000 bar, which is what Tenths() divides by 10.
func tenths(e *spelldata.Effect) float64 {
	return value(e) / 10
}

func humanise(s *spelldata.Spell, e *spelldata.Effect) string {
	switch e.Type {
	case dbcenums.E_SCHOOL_DAMAGE:
		return sentence(join(amount(e), schoolName(s.SpellSchool()), "damage", targetPhrase(e)),
			scaling(e), coefficients(e))
	case dbcenums.E_HEAL:
		return sentence(join(amount(e), "healing", targetPhrase(e)), scaling(e), coefficients(e))

	case dbcenums.E_WEAPON_PERCENT_DAMAGE:
		return sentence(join(percent(e)+" weapon damage", targetPhrase(e)), scaling(e))
	case dbcenums.E_NORMALIZED_WEAPON_DMG:
		return sentence(join("normalised weapon damage", bonus(e), targetPhrase(e)), scaling(e))
	case dbcenums.E_WEAPON_DAMAGE:
		return sentence(join("weapon damage", bonus(e), targetPhrase(e)), scaling(e))
	case dbcenums.E_WEAPON_DAMAGE_NOSCHOOL:
		return sentence(join("weapon damage in the weapon's own school", bonus(e), targetPhrase(e)),
			scaling(e))

	case dbcenums.E_ENERGIZE:
		return sentence(join("restores", powerAmount(e), targetPhrase(e)), scaling(e))
	case dbcenums.E_ENERGIZE_PCT:
		return join("restores", percent(e), "of maximum", powerName(dbcenums.PowerType(e.Misc)), targetPhrase(e))

	case dbcenums.E_TRIGGER_SPELL, dbcenums.E_TRIGGER_SPELL_2:
		return join("casts", triggerPhrase(e), targetPhrase(e))

	case dbcenums.E_THREAT:
		return sentence(join(amount(e), "threat", targetPhrase(e)), coefficients(e))
	case dbcenums.E_INTERRUPT_CAST:
		return join("interrupts the cast", targetPhrase(e))
	case dbcenums.E_DISPEL:
		return join("dispels", amount(e), dispelName(e.Misc), "effect", targetPhrase(e))
	case dbcenums.E_SUMMON:
		return join("summons creature", fmt.Sprint(e.Misc), targetPhrase(e))
	case dbcenums.E_ENCHANT_ITEM_TEMPORARY:
		return join("applies temporary enchant", fmt.Sprint(e.Misc), targetPhrase(e))

	case dbcenums.E_DISPEL_MECHANIC:
		return join("dispels", mechanicName(e.Misc), targetPhrase(e))
	case dbcenums.E_CHARGE:
		return join("charges", targetPhrase(e))
	case dbcenums.E_ATTACK_ME:
		return join("taunts", targetPhrase(e))
	case dbcenums.E_ADD_EXTRA_ATTACKS:
		return join(number(value(e)), "extra attacks")
	case dbcenums.E_LEARN_SPELL:
		return join("teaches", triggerOrMisc(e), targetPhrase(e))
	case dbcenums.E_CREATE_ITEM:
		return join("creates item", fmt.Sprint(e.Misc), targetPhrase(e))

	case dbcenums.E_DUMMY:
		return sentence(join(dummyPhrase("effect", e), targetPhrase(e)), scaling(e))

	case dbcenums.E_APPLY_AURA:
		return humaniseAura(s, e)
	case dbcenums.E_APPLY_AREA_AURA_PARTY:
		return areaAura("party aura", s, e)
	case dbcenums.E_APPLY_AREA_AURA_RAID:
		return areaAura("raid aura", s, e)
	}
	return ""
}

func areaAura(kind string, s *spelldata.Spell, e *spelldata.Effect) string {
	if phrase := humaniseAura(s, e); phrase != "" {
		return kind + ": " + phrase
	}
	return ""
}

// An aura line is what the aura itself says, qualified by the columns every aura carries: the unit it
// reaches, the ticks its spell's duration fits and the level its amount is priced at.
func humaniseAura(s *spelldata.Spell, e *spelldata.Effect) string {
	phrase := auraPhrase(s, e)
	if phrase == "" {
		return ""
	}
	return sentence(join(phrase, auraTarget(e)), ticks(s, e), scaling(e), auraCoefficients(e), castOnly(s, e))
}

// A hit or crit aura under the rule spelldata.WeaponScope names for its row, which is the one the parse
// attaches it by. A per-weapon mod is physical crit or hit on the matching weapon's attacks.
func weaponBound(s *spelldata.Spell, e *spelldata.Effect, stat string) string {
	rule, scope := spelldata.WeaponScope(s, e)
	switch rule {
	case spelldata.WeaponRulePerWeapon:
		physical := "physical crit"
		if e.Aura == dbcenums.A_MOD_HIT_CHANCE {
			physical = "physical hit"
		}
		return join(signedPercent(e), physical+",", perWeapon(scope))
	case spelldata.WeaponRuleWhileEquipped:
		return join(signedPercent(e), stat+",", "while", weaponsPhrase(scope.EquipSubclass, scope.EquipInvType, scope.EquipHand),
			"is equipped: every spell and attack")
	}
	return join(signedPercent(e), stat)
}

// A damage done aura under the rule spelldata.WeaponScope names for its row. While equipped, a melee
// weapon also carries it to the melee-defense spells of the schools the mask leaves out.
func damageDonePhrase(s *spelldata.Spell, e *spelldata.Effect) string {
	phrase := join(signedPercent(e), schoolName(core.SpellSchool(e.Misc)), "damage done")
	rule, scope := spelldata.WeaponScope(s, e)
	switch rule {
	case spelldata.WeaponRulePerWeapon:
		return join(phrase+",", perWeapon(scope))
	case spelldata.WeaponRuleWhileEquipped:
		if scope.ProcMask.Matches(core.ProcMaskMelee) && core.SpellSchool(e.Misc) != allSchools {
			phrase += " and on the melee-defense spells of the other schools"
		}
		return join(phrase+",", "while", weaponsPhrase(scope.EquipSubclass, scope.EquipInvType, scope.EquipHand), "is equipped")
	}
	return phrase
}

func perWeapon(scope core.SpellModConfig) string {
	weapons := weaponsPhrase(scope.EquipSubclass, scope.EquipInvType, scope.EquipHand)
	if scope.ProcMask == core.ProcMaskRanged {
		return "ranged weapons only: attacks made with " + weapons
	}
	return "per weapon: only attacks made with " + weapons
}

// A hit, crit or rating aura on a row whose weapon requirement binds nothing the effect gives, since
// the aura reaches another unit.
func castOnly(s *spelldata.Spell, e *spelldata.Effect) string {
	if !spelldata.WeaponBindable(e.Aura) {
		return ""
	}
	if rule, _ := spelldata.WeaponScope(s, e); rule != spelldata.WeaponRuleNone ||
		s.EquipClass != dbcenums.ITEM_CLASS_WEAPON || e.Target[0] == dbcenums.TARGET_UNIT_CASTER {
		return ""
	}
	return "the weapon requirement only gates casting"
}

// A_MOD_RATING by the ratings its mask names, read the way the parse reads them. A row bound to a
// weapon, or a mask of ranged bits alone, is the percent its hit and crit are worth on those attacks,
// and its other ratings are not read.
func ratingPhrase(s *spelldata.Spell, e *spelldata.Effect) string {
	rule, scope := spelldata.WeaponScope(s, e)
	if rule != spelldata.WeaponRulePerWeapon && rule != spelldata.WeaponRuleRangedRatings {
		ratings := spelldata.RatingStats(e.Misc)
		if len(ratings) == 0 {
			return ""
		}
		return join(signed(value(e)), ratingNames(ratings))
	}

	scoped := spelldata.ScopedRatingPercents(e.Misc, value(e))
	if len(scoped) == 0 {
		return ""
	}
	var ratings, unread []stats.Stat
	var worth []string
	for _, rating := range scoped {
		ratings = append(ratings, rating.Stat)
		switch rating.Stat {
		case stats.HitRating:
			worth = append(worth, signed(rating.Percent)+"% hit")
		case stats.CritRating:
			worth = append(worth, signed(rating.Percent)+"% crit")
		default:
			unread = append(unread, rating.Stat)
		}
	}

	clause := "ranged attacks only"
	if rule == spelldata.WeaponRulePerWeapon {
		clause = perWeapon(scope)
	}
	phrase := join(signed(value(e)), ratingNames(ratings)+",", clause)
	if len(worth) > 0 {
		phrase += ", worth " + strings.Join(worth, " and ")
	}
	if len(unread) > 0 {
		phrase += "; the " + ratingNames(unread) + " is not read"
	}
	return phrase
}

// Ratings as one phrase: hit rating, or hit, crit and haste rating.
func ratingNames(ratings []stats.Stat) string {
	names := make([]string, 0, len(ratings))
	for _, stat := range ratings {
		names = append(names, strings.ToLower(strings.TrimSuffix(stat.StatName(), "Rating")))
	}
	if len(names) == 1 {
		return names[0] + " rating"
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1] + " rating"
}

// An aura on the caster is the ordinary case and says nothing worth a clause, so only an aura
// reaching somebody else names who.
func auraTarget(e *spelldata.Effect) string {
	if e.Target[0] == dbcenums.TARGET_UNIT_CASTER {
		return ""
	}
	return targetPhrase(e)
}

// The spell power and attack power shares, for the auras whose amount those columns really add to. On
// a modifier or a proc row the coefficient column carries a 1 that means nothing.
func auraCoefficients(e *spelldata.Effect) string {
	switch e.Aura {
	case dbcenums.A_PERIODIC_DAMAGE, dbcenums.A_PERIODIC_HEAL, dbcenums.A_PERIODIC_LEECH,
		dbcenums.A_SCHOOL_ABSORB, dbcenums.A_DAMAGE_SHIELD:
		return coefficients(e)
	}
	return ""
}

// The auras sim/core/spelldata/parse_effects.go knows how to attach, the ticking and proc ones
// resolve_aura.go and resolve_proc.go read, and the handful whose single column means one thing only.
// An aura whose columns need the tooltip to read is left to its literal: wording it would put a guess
// where the row states a fact.
func auraPhrase(s *spelldata.Spell, e *spelldata.Effect) string {
	switch e.Aura {
	case dbcenums.A_PERIODIC_DAMAGE:
		return join(amount(e), schoolName(s.SpellSchool()), "damage", every(e))
	case dbcenums.A_PERIODIC_HEAL:
		return join(amount(e), "healing", every(e))
	case dbcenums.A_PERIODIC_LEECH:
		return join("leeches", amount(e), every(e))
	case dbcenums.A_PERIODIC_ENERGIZE:
		return join("restores", powerAmount(e), every(e))
	case dbcenums.A_PERIODIC_TRIGGER_SPELL:
		return join("casts", triggerPhrase(e), every(e))
	case dbcenums.A_PROC_TRIGGER_SPELL:
		return join("casts", triggerPhrase(e), "on its trigger", onSpells(e))

	case dbcenums.A_DUMMY, dbcenums.A_PERIODIC_DUMMY:
		return join(dummyPhrase("aura", e), every(e), onSpells(e))

	case dbcenums.A_MOD_DECREASE_SPEED, dbcenums.A_MOD_INCREASE_SPEED:
		return join(signedPercent(e), "movement speed")
	case dbcenums.A_MOD_STUN:
		return "stuns"
	case dbcenums.A_MOD_ROOT:
		return "roots"
	case dbcenums.A_MOD_SILENCE:
		return "silences"
	case dbcenums.A_MOD_FEAR:
		return "fears"
	case dbcenums.A_MOD_TAUNT:
		return "taunts"

	case dbcenums.A_SCHOOL_ABSORB:
		return join("absorbs", amount(e), schoolName(core.SpellSchool(e.Misc)), "damage")
	case dbcenums.A_SCHOOL_IMMUNITY:
		return join("immune to", schoolName(core.SpellSchool(e.Misc)))
	case dbcenums.A_MECHANIC_IMMUNITY:
		return join("immune to", mechanicName(e.Misc))
	case dbcenums.A_DAMAGE_SHIELD:
		return sentence(join(amount(e), schoolName(s.SpellSchool()), "damage back to melee attackers"), scaling(e))

	case dbcenums.A_MOD_SKILL:
		return fmt.Sprintf("%s to skill %d", signed(value(e)), e.Misc)
	case dbcenums.A_MOD_INCREASE_ENERGY:
		return join(signed(value(e)), "maximum", powerName(dbcenums.PowerType(e.Misc)))
	case dbcenums.A_MOD_MANA_REGEN_INTERRUPT:
		return join(signedPercent(e), "mana regen while casting")
	case dbcenums.A_ADD_TARGET_TRIGGER:
		return join("casts", triggerPhrase(e), "when the target is hit", onSpells(e))
	case dbcenums.A_OVERRIDE_ACTIONBAR_SPELLS:
		return join("replaces an action bar spell with", triggerOrMisc(e))
	case dbcenums.A_MOD_SHAPESHIFT:
		return join("shifts into", formName(int(e.Misc)))
	case dbcenums.A_REFLECT_SPELLS_SCHOOL:
		return join("reflects", percent(e), "of", schoolName(core.SpellSchool(e.Misc)), "damage")

	case dbcenums.A_ADD_FLAT_MODIFIER:
		return join(namedOr(dbcenums.SpellModOp(e.Misc), "op %d"), flatModAmount(e), modTargets(s, e))
	case dbcenums.A_ADD_PCT_MODIFIER:
		return join(namedOr(dbcenums.SpellModOp(e.Misc), "op %d"), signedPercent(e), modTargets(s, e))

	case dbcenums.A_MOD_STAT:
		return join(signed(value(e)), statName(e.Misc))
	case dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE:
		if e.Misc2 != 0 {
			return join(signedPercent(e), statMaskName(e.Misc2))
		}
		return join(signedPercent(e), statName(e.Misc))

	case dbcenums.A_MOD_RESISTANCE, dbcenums.A_MOD_ARMOR_NO_MODS:
		return join(signed(value(e)), resistanceName(e.Misc))
	case dbcenums.A_MOD_BASE_RESISTANCE_PCT:
		return join(signedPercent(e), resistanceName(e.Misc), "from items")

	case dbcenums.A_MOD_ATTACK_POWER:
		return join(signed(value(e)), "attack power")
	case dbcenums.A_MOD_RANGED_ATTACK_POWER:
		return join(signed(value(e)), "ranged attack power")

	case dbcenums.A_MOD_DAMAGE_DONE:
		return join(signed(value(e)), schoolName(core.SpellSchool(e.Misc)), "damage done")
	case dbcenums.A_MOD_DAMAGE_PERCENT_DONE:
		return damageDonePhrase(s, e)
	case dbcenums.A_MOD_DAMAGE_PERCENT_TAKEN:
		return join(signedPercent(e), schoolName(core.SpellSchool(e.Misc)), "damage taken")
	case dbcenums.A_MOD_DAMAGE_TAKEN:
		return join(signed(value(e)), schoolName(core.SpellSchool(e.Misc)), "damage taken")

	case dbcenums.A_MOD_THREAT:
		return join(signedPercent(e), "threat")
	case dbcenums.A_MOD_POWER_COST_SCHOOL_PCT:
		return join(signedPercent(e), schoolName(core.SpellSchool(e.Misc)), "power cost")
	case dbcenums.A_MOD_ADDITIONAL_POWER_COST:
		return join(signed(value(e)), "extra", powerName(dbcenums.PowerType(e.Misc)), "per cast")

	case dbcenums.A_MOD_HIT_CHANCE:
		return weaponBound(s, e, "physical hit")
	case dbcenums.A_MOD_SPELL_HIT_CHANCE:
		return join(signedPercent(e), "spell hit")
	case dbcenums.A_MOD_WEAPON_CRIT_PERCENT:
		return weaponBound(s, e, "physical crit")
	case dbcenums.A_MOD_SPELL_CRIT_CHANCE:
		return join(signedPercent(e), "spell crit")
	case dbcenums.A_MOD_CRIT_PCT:
		return weaponBound(s, e, "physical and spell crit")
	case dbcenums.A_MOD_RATING:
		return ratingPhrase(s, e)
	case dbcenums.A_MOD_CRIT_DAMAGE_BONUS:
		return join(signedPercent(e), "crit damage")

	case dbcenums.A_MOD_CASTING_SPEED_NOT_STACK:
		return join(signedPercent(e), "cast speed")
	case dbcenums.A_MOD_MELEE_HASTE_3:
		return join(signedPercent(e), "melee haste")
	case dbcenums.A_MOD_ATTACKSPEED:
		return join(signedPercent(e), "attack speed")

	case dbcenums.A_MOD_HEALING_DONE:
		return join(signed(value(e)), "healing power")
	case dbcenums.A_MOD_HEALING_DONE_PERCENT:
		return join(signedPercent(e), "healing done")
	case dbcenums.A_MOD_HEALING_PCT:
		return join(signedPercent(e), "healing taken")
	case dbcenums.A_MOD_HEALING:
		return join(signed(value(e)), "healing taken")

	case dbcenums.A_MOD_POWER_REGEN:
		return join(signed(value(e)), powerName(dbcenums.PowerType(e.Misc)), "per 5 s")
	case dbcenums.A_MOD_INCREASE_HEALTH:
		return join(signed(value(e)), "health")
	case dbcenums.A_MOD_INCREASE_HEALTH_PERCENT:
		return join(signedPercent(e), "health")

	case dbcenums.A_MOD_BLOCK_PERCENT:
		return join(signedPercent(e), "block chance")
	case dbcenums.A_MOD_DODGE_PERCENT:
		return join(signedPercent(e), "dodge chance")
	case dbcenums.A_MOD_PARRY_PERCENT:
		return join(signedPercent(e), "parry chance")
	case dbcenums.A_MOD_OFFHAND_DAMAGE_PCT:
		return join(signedPercent(e), "off-hand damage")
	case dbcenums.A_MOD_EXPERTISE:
		return join(signed(value(e)), "expertise")

	case dbcenums.A_MECHANIC_DURATION_MOD:
		return join(signedPercent(e), mechanicName(e.Misc), "duration")
	}
	return ""
}

// A_ADD_FLAT_MODIFIER states its amount in the units of the property it names, which is what the
// flat table in parse_effects_table.go converts by op.
func flatModAmount(e *spelldata.Effect) string {
	switch dbcenums.SpellModOp(e.Misc) {
	case dbcenums.SPELLMOD_CASTING_TIME, dbcenums.SPELLMOD_COOLDOWN,
		dbcenums.SPELLMOD_GLOBAL_COOLDOWN, dbcenums.SPELLMOD_DURATION:
		return signed(value(e)/1000) + " s"
	case dbcenums.SPELLMOD_COST:
		// Which conversion applies is the caster's bar, not the row's: the flat table divides by ten
		// only for a unit with a rage bar, so both readings are stated.
		return fmt.Sprintf("%s (%s on a rage bar)", signed(value(e)), signed(tenths(e)))
	case dbcenums.SPELLMOD_CRITICAL_CHANCE, dbcenums.SPELLMOD_RESIST_MISS_CHANCE,
		dbcenums.SPELLMOD_CHANCE_OF_SUCCESS:
		return signedPercent(e)
	}
	return signed(value(e))
}

// A dummy row is the game files' "the server does this", so its number is stated and nothing is
// claimed about what the number means: it is the one shape a port has to read the tooltip for.
func dummyPhrase(kind string, e *spelldata.Effect) string {
	if value(e) == 0 {
		return "a dummy " + kind
	}
	return fmt.Sprintf("a dummy %s holding %s", kind, number(value(e)))
}

// A_OVERRIDE_ACTIONBAR_SPELLS names the replacement in its trigger column on some rows and in its
// misc value on others.
func triggerOrMisc(e *spelldata.Effect) string {
	if e.TriggerID != 0 {
		return triggerPhrase(e)
	}
	return spellRef(e.Misc)
}

// The spells a modifier effect reaches. An empty mask on a spell that has a family of its own is the
// game files' "every spell of my family", which is what modConfig() widens it to.
func modTargets(s *spelldata.Spell, e *spelldata.Effect) string {
	if phrase := classFlagsPhrase(e.ClassFlags); phrase != "" {
		return "on " + phrase
	}
	if s.ClassFlags.Family != 0 {
		return fmt.Sprintf("on every spell of family %d", s.ClassFlags.Family)
	}
	return "on every spell"
}

// The spells a proc or dummy effect names, where it names any: the same mask, read the way
// rowClassFlags() reads it.
func onSpells(e *spelldata.Effect) string {
	if phrase := classFlagsPhrase(e.ClassFlags); phrase != "" {
		return "from " + phrase
	}
	return ""
}

func triggerPhrase(e *spelldata.Effect) string {
	if e.TriggerID == 0 {
		return "nothing"
	}
	return spellRef(e.TriggerID)
}

// A weapon effect states a bonus on top of the swing rather than an amount, so a zero there is no
// bonus rather than no damage.
func bonus(e *spelldata.Effect) string {
	if value(e) == 0 {
		return ""
	}
	return signed(value(e))
}

// The level the amount is priced at, for a row whose per-level gain means the number is not the one
// the game files' own column states.
func scaling(e *spelldata.Effect) string {
	if e.PPL == 0 {
		return ""
	}
	return fmt.Sprintf("at level %d", core.CharacterLevel)
}

func coefficients(e *spelldata.Effect) string {
	return join(share(e.Coeff(), "spell power"), share(e.APCoeff(), "attack power"))
}

// A coefficient of exactly 1 is the game files' filler - 2,112 of the store's effects carry it, weapon
// effects and shapeshifts among them - so it is left to the literal rather than stated as a share.
func share(coeff float64, of string) string {
	if coeff == 0 || coeff == 1 {
		return ""
	}
	return signed(coeff) + " " + of
}

// The effect as the game files state it: the columns it fills, in the units the row keeps them in,
// so the words above can be checked against them.
func literal(e *spelldata.Effect, pos int) string {
	parts := []string{namedOr(e.Type, "E_%d")}
	if e.Aura != 0 {
		parts = append(parts, namedOr(e.Aura, "A_%d"))
	}
	parts = append(parts, "base="+number(e.BasePoints))

	add := func(format string, args ...any) {
		parts = append(parts, fmt.Sprintf(format, args...))
	}
	if e.PPL != 0 {
		add("ppl=%s", number(e.PPL))
	}
	if e.Variance != 0 {
		add("variance=%s", number(e.Variance))
	}
	if e.SPCoef != 0 {
		add("sp=%s", number(e.SPCoef))
	}
	if e.APCoef != 0 {
		add("ap=%s", number(e.APCoef))
	}
	if e.PeriodMs != 0 {
		add("period=%dms", e.PeriodMs)
	}
	if e.Amplitude != 0 {
		add("amplitude=%s", number(float64(e.Amplitude)))
	}
	if e.Misc != 0 {
		add("misc=%d", e.Misc)
	}
	if e.Misc2 != 0 {
		add("misc2=%d", e.Misc2)
	}
	if e.TriggerID != 0 {
		add("trigger=%d", e.TriggerID)
	}
	if e.ChainTargets != 0 {
		add("chain=%d", e.ChainTargets)
	}
	if e.RadiusMin != 0 {
		add("radiusMin=%s", number(float64(e.RadiusMin)))
	}
	if e.RadiusMax != 0 {
		add("radius=%s", number(float64(e.RadiusMax)))
	}
	if e.PointsPerResource != 0 {
		add("perResource=%s", number(float64(e.PointsPerResource)))
	}
	if e.Mechanic != 0 {
		add("mechanic=%d", e.Mechanic)
	}
	if !e.ClassFlags.IsZero() {
		add("family=%d", e.ClassFlags.Family)
		if mask := maskWords(e.ClassFlags); mask != "" {
			add("mask=%s", mask)
		}
	}
	// EffectIndex is the game files' own numbering and matches the position on all but 46 of the
	// store's rows, so it is stated only where the two disagree and EffectN(pos) is not
	// EffectIndex.
	if int(e.Index) != pos-1 {
		add("index=%d", e.Index)
	}
	add("target=[%d,%d]", e.Target[0], e.Target[1])
	return strings.Join(parts, " ")
}

func every(e *spelldata.Effect) string {
	if e.PeriodMs == 0 {
		return ""
	}
	return "every " + seconds(e.PeriodMs)
}

// How many times the aura ticks over its spell's duration, where both are stated.
func ticks(s *spelldata.Spell, e *spelldata.Effect) string {
	if s.DurationMs <= 0 || e.PeriodMs <= 0 {
		return ""
	}
	return fmt.Sprintf("%d ticks", s.DurationMs/e.PeriodMs)
}

// The unit or area the effect reaches. The game files state two implicit targets, and a pair naming a
// place and then a unit - Whirlwind's [22,15] - is worded by the unit.
func targetPhrase(e *spelldata.Effect) string {
	first, second := implicitTargets[e.Target[0]], implicitTargets[e.Target[1]]

	if first.place && (second.phrase != "" || e.Target[1] != 0) {
		return targetWord(e.Target[1], second)
	}
	if e.Target[0] == 0 {
		return ""
	}
	return targetWord(e.Target[0], first)
}

func targetWord(t dbcenums.ImplicitTarget, named implicitTarget) string {
	if named.phrase != "" {
		return named.phrase
	}
	if named.place {
		return ""
	}
	return fmt.Sprintf("to target %d", t)
}

// EffectMiscValue_0 is the power type on an energize effect, and rage is on the game files' 0-1000 bar
// the way a cost is.
func powerAmount(e *spelldata.Effect) string {
	amount := value(e)
	if dbcenums.PowerType(e.Misc) == dbcenums.POWER_RAGE {
		amount = tenths(e)
	}
	bar := powerName(dbcenums.PowerType(e.Misc))
	if amount == 1 {
		bar = strings.TrimSuffix(bar, "s")
	}
	return join(number(amount), bar)
}

// A_MOD_RESISTANCE, A_MOD_ARMOR_NO_MODS and A_MOD_BASE_RESISTANCE_PCT state armor as school 1 and a
// magic resistance as that school's own bit.
func resistanceName(misc int32) string {
	mask := core.SpellSchool(misc)
	if mask == core.SpellSchoolPhysical {
		return "armor"
	}
	names := schoolName(mask)
	if mask.Matches(core.SpellSchoolPhysical) {
		return strings.Replace(names, "physical", "armor", 1) + " resistance"
	}
	return names + " resistance"
}

// EffectMiscValue_0 on A_MECHANIC_DURATION_MOD, of which the parse table acts on two.
func mechanicName(misc int32) string {
	if misc >= 0 && misc <= 0xff {
		if name, ok := strings.CutPrefix(dbcenums.Mechanic(misc).String(), "MECHANIC_"); ok {
			return strings.ToLower(strings.ReplaceAll(name, "_", " "))
		}
	}
	return fmt.Sprintf("mechanic %d", misc)
}

func sentence(phrase string, notes ...string) string {
	kept := nonEmpty(notes)
	if len(kept) == 0 {
		return phrase
	}
	return phrase + " (" + strings.Join(kept, ", ") + ")"
}
