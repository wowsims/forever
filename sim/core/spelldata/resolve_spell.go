package spelldata

import (
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// An addition to the resolved config for what the game files do not state: the proc mask, the metrics
// the sim keeps and the flags a row cannot justify.
type SpellOpt func(*core.SpellConfig, *Spell)

// SpellCategories.StartRecoveryCategory of the global cooldown. It is the only category the store
// carries - 1763 rows state 133 and the rest state nothing - so a row outside it spends no GCD.
const globalCooldownCategory int16 = 133

// ImplicitTarget_0 values that name a friendly unit.
var helpfulTargets = []dbcenums.ImplicitTarget{
	dbcenums.TARGET_UNIT_CASTER, dbcenums.TARGET_UNIT_PET, dbcenums.TARGET_UNIT_CASTER_AREA_PARTY,
	dbcenums.TARGET_UNIT_TARGET_ALLY, dbcenums.TARGET_UNIT_SRC_AREA_ALLY, dbcenums.TARGET_UNIT_DEST_AREA_ALLY,
	dbcenums.TARGET_UNIT_SRC_AREA_PARTY, dbcenums.TARGET_UNIT_DEST_AREA_PARTY, dbcenums.TARGET_UNIT_TARGET_PARTY,
	dbcenums.TARGET_UNIT_TARGET_CHAINHEAL_ALLY, dbcenums.TARGET_UNIT_CASTER_AREA_RAID, dbcenums.TARGET_UNIT_TARGET_RAID,
}

// ImplicitTarget values that name an enemy unit, or a point or area chosen on one.
var enemyTargets = []dbcenums.ImplicitTarget{
	dbcenums.TARGET_UNIT_NEARBY_ENEMY, dbcenums.TARGET_UNIT_TARGET_ENEMY, dbcenums.TARGET_UNIT_SRC_AREA_ENEMY,
	dbcenums.TARGET_UNIT_DEST_AREA_ENEMY, dbcenums.TARGET_UNIT_CONE_ENEMY_24, dbcenums.TARGET_DEST_DYNOBJ_ENEMY,
	dbcenums.TARGET_DEST_TARGET_ENEMY, dbcenums.TARGET_UNIT_CONE_180_DEG_ENEMY, dbcenums.TARGET_UNIT_CONE_CASTER_TO_DEST_ENEMY,
	dbcenums.TARGET_UNIT_SRC_AREA_FURTHEST_ENEMY, dbcenums.TARGET_UNIT_AND_DEST_LAST_ENEMY,
	dbcenums.TARGET_UNIT_CASTER_AREA_ENEMY_CLUMP, dbcenums.TARGET_DEST_CASTER_ENEMY_CLUMP_CENTROID,
	dbcenums.TARGET_UNIT_RECT_CASTER_ENEMY, dbcenums.TARGET_UNIT_LINE_CASTER_TO_DEST_ENEMY,
}

// Whether either of the effect's implicit targets is an enemy.
func (e *Effect) HitsAnEnemy() bool {
	return TargetsAnEnemy(e.Target[0]) || TargetsAnEnemy(e.Target[1])
}

// Whether the implicit target names an enemy unit, or a point or area chosen on one.
func TargetsAnEnemy(target dbcenums.ImplicitTarget) bool {
	return slices.Contains(enemyTargets, target)
}

// The enemy targets that pick every enemy in a cone in front of the caster.
var coneEnemyTargets = []dbcenums.ImplicitTarget{
	dbcenums.TARGET_UNIT_CONE_ENEMY_24, dbcenums.TARGET_UNIT_CONE_180_DEG_ENEMY, dbcenums.TARGET_UNIT_CONE_CASTER_TO_DEST_ENEMY,
}

// The enemy targets that pick every enemy in an area, cone, rectangle or line rather than one unit.
var areaEnemyTargets = slices.Concat(coneEnemyTargets, []dbcenums.ImplicitTarget{
	dbcenums.TARGET_UNIT_SRC_AREA_ENEMY, dbcenums.TARGET_UNIT_DEST_AREA_ENEMY,
	dbcenums.TARGET_UNIT_SRC_AREA_FURTHEST_ENEMY, dbcenums.TARGET_UNIT_CASTER_AREA_ENEMY_CLUMP,
	dbcenums.TARGET_UNIT_RECT_CASTER_ENEMY, dbcenums.TARGET_UNIT_LINE_CASTER_TO_DEST_ENEMY,
})

// Whether either of the effect's implicit targets picks the enemies in an area. The area is often the
// second: Shard of the Fallen Star states TARGET_DEST_TARGET_ENEMY, then TARGET_UNIT_DEST_AREA_ENEMY.
func (e *Effect) HitsAnArea() bool {
	return slices.Contains(areaEnemyTargets, e.Target[0]) || slices.Contains(areaEnemyTargets, e.Target[1])
}

// Whether either of the effect's implicit targets picks the enemies in a cone in front of the caster.
func (e *Effect) HitsACone() bool {
	return slices.Contains(coneEnemyTargets, e.Target[0]) || slices.Contains(coneEnemyTargets, e.Target[1])
}

// The first implicit targets that centre an area on the caster.
var casterCentres = []dbcenums.ImplicitTarget{dbcenums.TARGET_SRC_CASTER, dbcenums.TARGET_DEST_CASTER}

// Whether the effect picks the enemies in an area centred on the caster: Goblin Sapper Charge's 13241
// states TARGET_SRC_CASTER, then TARGET_UNIT_SRC_AREA_ENEMY.
func (e *Effect) HitsAroundTheCaster() bool {
	return e.HitsAnArea() && slices.Contains(casterCentres, e.Target[0])
}

// What the game files state about a spell, as the fields core registers it through. The caller adds
// the proc mask, ApplyEffects and anything the game files do not carry to the returned value before
// handing it to RegisterSpell: the resolver fills the row's own fields and nothing else.
//
// The unit is needed for the cooldown timers, so a config is built where the sim has a character,
// not at package init. MaxTargets has no SpellConfig field; a caller that needs the row's cap reads
// s.MaxTargets itself.
func SpellConfig(unit *core.Unit, s *Spell, opts ...SpellOpt) core.SpellConfig {
	config := core.SpellConfig{
		ActionID:    core.ActionID{SpellID: s.ID},
		Rank:        rankOf(s),
		SpellSchool: s.SpellSchool(),
		DefenseType: s.DefenseTypeCore(),
		ClassFlags:  s.ClassFlags,
		Flags:       rowFlags(s),
		MinRange:    float64(s.MinRange),
		MaxRange:    float64(s.MaxRange),
		Cast:        castConfig(unit, s),

		MissileSpeed:       float64(s.Speed),
		MissileMinDuration: s.MissileMinDuration(),

		CastRequirement: s.CastRequirement(),
	}
	applyCost(&config, s)
	// A row with Fixed Travel Time flies for its Speed in seconds at any distance: Dragonfire Potion's
	// 1252554 lands 1 sec after use.
	if s.FixedTravelTime() {
		config.MissileSpeed = 0
		config.MissileMinDuration = max(config.MissileMinDuration, core.DurationFromSeconds(float64(s.Speed)))
	}
	// Forever moves the minimum range of a bow, gun or crossbow shot from the 8 yd the row states to
	// core.MinRangedRange.
	if s.MinRange != 0 && s.EquipClass == dbcenums.ITEM_CLASS_WEAPON && s.EquipSubclass&(dbcenums.ITEM_SUBCLASS_BIT_WEAPON_BOW|dbcenums.ITEM_SUBCLASS_BIT_WEAPON_GUNS|dbcenums.ITEM_SUBCLASS_BIT_WEAPON_CROSSBOW) != 0 {
		config.MinRange = core.MinRangedRange
	}
	// The row's range is to the caster's target; damage around the point the triggering spell chose
	// reaches anywhere.
	if damage := s.DamageEffect(); damage.HitsAnArea() && damage.Target[0] == dbcenums.TARGET_DEST_DEST {
		config.MaxRange = 0
	}
	if threat := s.firstOfType(dbcenums.E_THREAT, dbcenums.E_THREAT_ALL); threat != NilEffect {
		config.FlatThreatBonus = threat.Average()
		config.ThreatAPCoefficient = threat.APCoeff()
	} else {
		config.FlatThreatBonus = float64(s.FlatThreat)
	}
	if s.IsBleed() {
		config.DamageMultiplier = 1
		config.ThreatMultiplier = 1
	}

	// The row first, the caller's options on top, so an option sees what the row filled.
	for _, opt := range opts {
		opt(&config, s)
	}
	// An area centred on the caster, or a cone from it, reaches only the enemies inside it: a cast
	// reaches the radius, and a proc or tick has no range. Damage centred on the caster does so even
	// where the row states a range (Flamestrike's 18818 states 30 and reaches its 10 yd radius); a cone,
	// or an area on another effect, does so only where the row states none.
	if radius, ok := casterAreaRadius(s); ok && (config.MaxRange == 0 || s.DamageEffect().HitsAroundTheCaster()) {
		config.MaxRange = 0
		if !config.Flags.Matches(core.SpellFlagPassiveSpell) {
			config.MaxRange = float64(radius)
		}
	}
	return config
}

// The largest radius of the row's areas centred on the caster or cones from it, and whether it has
// one. A row with none takes the row its A_PERIODIC_TRIGGER_SPELL aura on the caster fires: Hellfire's
// channel fires Hellfire Effect, 10 yd around the warlock, every second.
func casterAreaRadius(s *Spell) (float32, bool) {
	radius, ok := ownCasterAreaRadius(s)
	if ok {
		return radius, true
	}
	for i := range s.Effects {
		if e := &s.Effects[i]; e.Aura == dbcenums.A_PERIODIC_TRIGGER_SPELL && e.Target[0] == dbcenums.TARGET_UNIT_CASTER {
			if r, found := ownCasterAreaRadius(e.Trigger()); found {
				radius, ok = max(radius, r), true
			}
		}
	}
	return radius, ok
}

func ownCasterAreaRadius(s *Spell) (radius float32, ok bool) {
	for i := range s.Effects {
		if e := &s.Effects[i]; e.HitsACone() || e.HitsAroundTheCaster() {
			radius, ok = max(radius, e.RadiusMax), true
		}
	}
	return radius, ok
}

// The proc mask, the melee metrics bucket and the multipliers an ability needs, for a physical spell.
func Melee(mask core.ProcMask) SpellOpt {
	return func(config *core.SpellConfig, s *Spell) {
		config.ProcMask = mask
		config.Flags |= core.SpellFlagMeleeMetrics | core.SpellFlagAPL
		config.DamageMultiplier = 1
		config.ThreatMultiplier = threatMultiplier(s)
		config.Cast.IgnoreHaste = true
	}
}

// The same for a spell that scales with spell power, whose share of it and whose roll the row states
// on the effect that deals the damage, or heals where the spell has no damaging effect.
func Magic(mask core.ProcMask) SpellOpt {
	return func(config *core.SpellConfig, s *Spell) {
		config.ProcMask = mask
		config.Flags |= core.SpellFlagAPL
		config.DamageMultiplier = 1
		config.ThreatMultiplier = threatMultiplier(s)
		config.BonusCoefficient = spellPowerCoeff(s)
		config.DirectVariance = amountEffect(s).DirectVariance()
	}
}

func threatMultiplier(s *Spell) float64 {
	if s.GeneratesNoThreat() {
		return 0
	}
	return 1
}

// A spell another spell or an aura casts: it is out of the rotation Melee and Magic put it in, it
// does not feed on-cast effects, and it has no cast, cost, cooldown, cast requirement or flat threat of its own even when it shares
// the row of the ability that casts it, the way the warrior's Blood Craze heal and Whirlwind's
// off-hand strike are registered. Its damage and healing are still measured, its casts are not: the
// metrics aggregator counts no cast for a passive spell, so a sub-spell whose casts the sim reports -
// the warrior's Deep Wounds and Retaliation's counterattack - sets the flags itself instead.
func Proc() SpellOpt {
	return func(config *core.SpellConfig, _ *Spell) {
		config.Flags |= core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete
		config.Flags &^= core.SpellFlagAPL
		config.Cast = core.CastConfig{}
		config.CastRequirement = core.CastRequirement{}
		config.ManaCost = core.ManaCostOptions{}
		config.RageCost = core.RageCostOptions{}
		config.EnergyCost = core.EnergyCostOptions{}
		config.FocusCost = core.FocusCostOptions{}
		config.FlatThreatBonus = 0
		config.ThreatAPCoefficient = 0
	}
}

// Flags the game files do not state, such as the sim's own metrics and rotation flags.
func Flags(flags core.SpellFlag) SpellOpt {
	return func(config *core.SpellConfig, _ *Spell) {
		config.Flags |= flags
	}
}

// RemoveFlags clears specific flags from the configuration that should not be present.
//
// Ex: ArcaneMissilesTickSpell is NotAProc so you cant use Proc() SpellOpt, so it ends up being castable through APL
//
// Usage: spelldata.RemoveFlags(core.SpellFlagAPL)
func RemoveFlags(flags core.SpellFlag) SpellOpt {
	return func(config *core.SpellConfig, _ *Spell) {
		config.Flags &^= flags
	}
}

// Splits one spell id into several actions, for a spell the sim registers more than once.
func Tag(tag int32) SpellOpt {
	return func(config *core.SpellConfig, _ *Spell) {
		config.ActionID.Tag = tag
	}
}

// The forms, caster auras, weapon or shield, out-of-combat state and target creature types the row
// requires. A shield is the only armor a castable spell asks for. The creature types limit the cast
// only on a spell cast on an enemy unit: Bestial Wrath's Beast is the pet it lands on, and Holy Wrath's
// Undead and Demon are the enemies its area hits.
//
// A ranked row whose only effect that does something is the periodic is refused while a higher rank
// of its dot is up: Shadow Word: Pain, Renew, Corruption, Rend, and Rip and Rupture, whose dummy
// effects do nothing by themselves. A row with any other effect casts, and only its periodic stays off.
func (s *Spell) CastRequirement() core.CastRequirement {
	r := core.CastRequirement{
		Forms:             s.StanceMask,
		ExcludedForms:     s.StanceExclude,
		CasterForm:        s.CastableInCasterForm(),
		NotShapeshifted:   s.NotShapeshifted(),
		CasterAura:        s.CasterAura,
		ExcludeCasterAura: s.ExcludeCasterAura,
		NotInCombat:       s.HasAttr(dbcenums.ATTR_INDEX_BASE, dbcenums.ATTR_NOT_IN_COMBAT),

		NotUnderHigherRank: s.RankNumber() > 0 && s.onlyPeriodicActs(),
	}
	switch s.EquipClass {
	case dbcenums.ITEM_CLASS_WEAPON:
		r.Weapon = core.EquipRequirement{SubclassMask: s.EquipSubclass, InvTypeMask: s.EquipInvType, Hand: RequiredHand(s)}
	case dbcenums.ITEM_CLASS_ARMOR:
		r.Shield = s.EquipSubclass&dbcenums.ITEM_SUBCLASS_BIT_ARMOR_SHIELD != 0
	}
	if s.EffectN(1).Target[0] == dbcenums.TARGET_UNIT_TARGET_ENEMY {
		for _, mobType := range MobTypes(s.TargetCreatureType) {
			r.TargetMobTypes |= 1 << mobType
		}
	}
	return r
}

// A periodic is an effect with a tick period, the one DotConfig builds a dot from.
func (s *Spell) onlyPeriodicActs() bool {
	var acting []*Effect
	for i := range s.Effects {
		if e := &s.Effects[i]; !e.isInert() {
			acting = append(acting, e)
		}
	}
	return len(acting) == 1 && acting[0].PeriodMs > 0
}

func (e *Effect) isInert() bool {
	return e.Type == dbcenums.E_DUMMY || e.Type == dbcenums.E_SCRIPT_EFFECT || e.Aura == dbcenums.A_DUMMY
}

func spellPowerCoeff(s *Spell) float64 {
	return amountEffect(s).Coeff()
}

func amountEffect(s *Spell) *Effect {
	if e := s.DamageEffect(); e != NilEffect {
		return e
	}
	return s.HealEffect()
}

// Spell.NameSubtext_lang, which is "Rank 4" on a ranked spell and a word like "Passive" or
// "Shapeshift" on a handful of others. A spell the game files show no rank on answers 0, which is what
// the APL UI reads as "this spell has no ranks".
func rankOf(s *Spell) int32 {
	digits, ranked := strings.CutPrefix(s.Rank, "Rank ")
	if !ranked {
		return 0
	}
	rank, err := strconv.Atoi(digits)
	if err != nil {
		return 0
	}
	return int32(rank)
}

// The flags the row's attributes and targets state. Everything else is the caller's: a flag the
// game files do not carry is not invented here.
func rowFlags(s *Spell) core.SpellFlag {
	var flags core.SpellFlag
	if s.IsPassive() {
		flags |= core.SpellFlagPassiveSpell
	}
	if s.IsChanneled() {
		flags |= core.SpellFlagChanneled
	}
	if s.SuppressesWeaponProcs() {
		flags |= core.SpellFlagSuppressWeaponProcs
	}
	if s.PushedBack() {
		flags |= core.SpellFlagPushback
	}
	if s.KeepsCombatTimers() {
		flags |= core.SpellFlagKeepsSwingTimer
	}
	// Helpful decides who the APL casts the spell on, so it follows the first effect's target. An
	// attack whose first effect is a self side-effect reads as helpful here and the caller clears it.
	if slices.Contains(helpfulTargets, s.EffectN(1).Target[0]) {
		flags |= core.SpellFlagHelpful
	}
	return flags
}

// The cast time and global cooldown the row states and whether haste shortens them, without the
// cooldowns: for a caller that runs the spell on cooldowns of its own.
func Cast(s *Spell) core.CastConfig {
	// Haste shortens a cast and the GCD it spends, which is the school's business rather than the hit
	// table's: the shouts, Taunt, Piercing Howl and Thunder Clap are physical spells the game files
	// put under the magic defense type, and they ignore haste like every other ability.
	cast := core.CastConfig{
		DefaultCast: core.Cast{CastTime: s.CastTime()},
		IgnoreHaste: s.SpellSchool() == core.SpellSchoolPhysical,
	}

	if s.StartRecoveryCategory == globalCooldownCategory {
		cast.DefaultCast.GCD = s.GCD()
	}
	return cast
}

func castConfig(unit *core.Unit, s *Spell) core.CastConfig {
	cast := Cast(s)

	switch {
	case s.CooldownMs > 0:
		cast.CD = core.Cooldown{Timer: unit.NewTimer(), Duration: s.Cooldown()}
		if s.CategoryCooldownMs > 0 {
			cast.SharedCD = core.Cooldown{Timer: categoryTimer(unit, s), Duration: s.CategoryCooldown()}
		}
	case s.CategoryCooldownMs > 0:
		cast.CD = core.Cooldown{Timer: categoryTimer(unit, s), Duration: s.CategoryCooldown()}
	}

	return cast
}

// The timer a category cooldown runs off. A category is a set of spells that share one cooldown, so
// the timer is the unit's for that category; 16 rows state a category cooldown without a category to
// share it with, and that is the spell's own recovery time.
func categoryTimer(unit *core.Unit, s *Spell) *core.Timer {
	if s.Category == 0 {
		return unit.NewTimer()
	}
	return unit.CategoryTimer(int32(s.Category))
}

// The cost out of the first bar the row states, and NonEmpty where the cast would otherwise read as
// an empty one, which is what a hand-written config sets on an off-GCD ability that spends a
// resource. It is not what keeps core's "Empty DefaultCast with a cost" panic away: core fills
// DefaultCast.Cost from the resolved cost before it tests for an empty cast. A row that spends
// nothing is left empty on purpose, so a passive or a proc spell keeps the cast path core gives those.
//
// Only the first bar: 45 rows state a second one - 44 of them combo points beside energy - and the
// sim has one cost per spell, so the caller spends the rest itself. A bar the sim does not model,
// which is the game files' -2 health on Bloodrage, resolves to no cost at all and leaves the cast
// as it found it rather than declaring a cost the spell does not take.
func applyCost(config *core.SpellConfig, s *Spell) {
	if len(s.Powers) == 0 {
		return
	}

	powerType := s.Powers[0].Type

	// Truncated rather than rounded: Retaliation states one rage-tenth and so costs nothing.
	cost := int32(s.PowerCost(powerType))
	costPct := float64(s.Powers[0].CostPct)

	switch powerType {
	case dbcenums.POWER_MANA:
		config.ManaCost = core.ManaCostOptions{FlatCost: cost}
		if costPct > 0 {
			config.ManaCost.BaseCostPercent = costPct
		}
	case dbcenums.POWER_RAGE:
		config.RageCost = core.RageCostOptions{Cost: cost, Refund: s.MissRefund()}
	case dbcenums.POWER_ENERGY:
		config.EnergyCost = core.EnergyCostOptions{Cost: cost, Refund: s.MissRefund()}
	case dbcenums.POWER_FOCUS:
		config.FocusCost = core.FocusCostOptions{Cost: cost, Refund: s.MissRefund()}
	default:
		return
	}

	spends := cost > 0 || costPct > 0
	if spends && config.Cast.DefaultCast.GCD == 0 && config.Cast.DefaultCast.CastTime == 0 {
		config.Cast.DefaultCast.NonEmpty = true
	}
}
