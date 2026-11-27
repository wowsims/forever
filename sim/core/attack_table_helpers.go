package core

import (
	"maps"
	"math"
	"math/bits"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The attack tables exist once the environment is set up, so the Unit helpers act on the tables of
// that moment: a caller runs them from a post-finalize effect or an aura's gain and expire.

// What the unit has against one mob type. The attack tables hold it only for the defenders the
// encounter has, so the helpers below keep it per mob type as well, from the moment the tables exist,
// for the stats sheet to list every mob type whatever the targets are. Nothing in combat reads it.
type mobTypeBonus struct {
	stats            stats.Stats
	damageMultiplier float64
	critMultiplier   float64
}

// One source's share of the unit's bonus against one mob type, for the combat log and the stats sheet
// to name the sources of a versus bonus. A share taken back off stays in the list at its neutral value.
type versusSource struct {
	source  ActionID
	mobType proto.MobType
	mobTypeBonus
}

// Changes the unit's bonus against each of the mob types once, however often the list names one, as
// the attack tables count it, and the source's share of it.
func (unit *Unit) changeMobTypeBonuses(source ActionID, mobTypes []proto.MobType, change func(bonus *mobTypeBonus)) {
	if unit.AttackTables == nil {
		return
	}
	if unit.mobTypeBonuses == nil {
		unit.mobTypeBonuses = make(map[proto.MobType]*mobTypeBonus)
	}
	for i, mobType := range mobTypes {
		if slices.Index(mobTypes, mobType) < i {
			continue
		}
		bonus := unit.mobTypeBonuses[mobType]
		if bonus == nil {
			bonus = &mobTypeBonus{damageMultiplier: 1, critMultiplier: 1}
			unit.mobTypeBonuses[mobType] = bonus
		}
		change(bonus)
		change(&unit.versusSource(source, mobType).mobTypeBonus)
	}
}

func (unit *Unit) versusSource(source ActionID, mobType proto.MobType) *versusSource {
	for _, share := range unit.versusSources {
		if share.source == source && share.mobType == mobType {
			return share
		}
	}
	share := &versusSource{source: source, mobType: mobType, mobTypeBonus: mobTypeBonus{damageMultiplier: 1, critMultiplier: 1}}
	unit.versusSources = append(unit.versusSources, share)
	return share
}

// Adds stats the unit has against every defender of the mob types; a negative bonus takes them away.
// The source is the item or spell the combat log names for them.
func (unit *Unit) AddMobTypeBonusStats(source ActionID, bonus stats.Stats, mobTypes ...proto.MobType) {
	for _, at := range unit.AttackTables {
		if mobType := at.Defender.MobType; slices.Contains(mobTypes, mobType) {
			at.MobTypeBonusStats[mobType] = at.MobTypeBonusStats[mobType].Add(bonus)
		}
	}
	unit.changeMobTypeBonuses(source, mobTypes, func(versus *mobTypeBonus) { versus.stats = versus.stats.Add(bonus) })
}

// An item's stats against mob types, every copy worn counting; sign -1 takes them off.
func (character *Character) addItemMobTypeStats(item *Item, sign float64) {
	for _, versus := range item.ScalingOptions[0].GetMobTypeStats() {
		character.AddMobTypeBonusStats(ActionID{ItemID: item.ID}, stats.FromProtoMap(versus.Stats).Multiply(sign), versus.MobTypes...)
	}
}

// Multiplies the damage the unit deals to every defender of the mob types; 1/mult takes the
// multiplier back off. The source is the item or spell the combat log names for it.
func (unit *Unit) MultiplyDamageVersus(source ActionID, mult float64, mobTypes ...proto.MobType) {
	for _, at := range unit.AttackTables {
		if slices.Contains(mobTypes, at.Defender.MobType) {
			at.DamageDealtMultiplier *= mult
		}
	}
	unit.changeMobTypeBonuses(source, mobTypes, func(versus *mobTypeBonus) { versus.damageMultiplier *= mult })
}

// Multiplies the crit damage the unit deals to every defender of the mob types on top of the damage
// multiplier; 1/mult takes it back off. The source is the item or spell the combat log names for it.
func (unit *Unit) MultiplyCritDamageVersus(source ActionID, mult float64, mobTypes ...proto.MobType) {
	for _, at := range unit.AttackTables {
		if slices.Contains(mobTypes, at.Defender.MobType) {
			at.CritMultiplier *= mult
		}
	}
	unit.changeMobTypeBonuses(source, mobTypes, func(versus *mobTypeBonus) { versus.critMultiplier *= mult })
}

// One source's crit chance against a defender, which the defender's mob type or family grants.
type versusCritChance struct {
	source   ActionID
	percent  float64
	byFamily bool
}

// Adds crit chance to the unit's hits on the target: percent, times the bonus's factor against it. Where
// the bonus names the target, the combat log names the source for all of it as a versus effect. A
// negative percent takes it away.
func (unit *Unit) AddCritChanceVersus(source ActionID, target *Unit, percent float64, versus *VersusBonus) {
	at := unit.AttackTables[target.UnitIndex]
	mult := versus.Against(target)
	at.BonusCritPercent += percent * mult
	if mult == 1 {
		return
	}
	for _, share := range at.versusCritChances {
		if share.source == source {
			share.percent += percent * mult
			return
		}
	}
	at.versusCritChances = append(at.versusCritChances, &versusCritChance{source: source, percent: percent * mult, byFamily: versus.namesFamilyOf(target)})
}

// Adds to the share of every defender's armor the unit's attacks ignore; a negative delta takes it away.
func (unit *Unit) AddArmorIgnoreFactor(delta float64) {
	for _, at := range unit.AttackTables {
		at.ArmorIgnoreFactor += delta
	}
}

// The stats against the mob types while the aura is up, named in the combat log by the source.
// Returns parent aura for chaining
func (parentAura *Aura) AttachMobTypeBonusStats(source ActionID, bonus stats.Stats, mobTypes ...proto.MobType) *Aura {
	parentAura.ApplyOnGain(func(aura *Aura, _ *Simulation) {
		aura.Unit.AddMobTypeBonusStats(source, bonus, mobTypes...)
	})

	parentAura.ApplyOnExpire(func(aura *Aura, _ *Simulation) {
		aura.Unit.AddMobTypeBonusStats(source, bonus.Invert(), mobTypes...)
	})

	if parentAura.IsActive() {
		parentAura.Unit.AddMobTypeBonusStats(source, bonus, mobTypes...)
	}

	return parentAura
}

// A bonus taken back off can leave a rounding error, which counts as none.
const versusEpsilon = 1e-9

// The stats as a proto map, leaving out any that only a rounding error left behind.
func versusStatMap(bonus stats.Stats) map[int32]float64 {
	statMap := map[int32]float64{}
	for stat, value := range bonus[:stats.ProtoStatsLen] {
		if math.Abs(value) > versusEpsilon {
			statMap[int32(stat)] = value
		}
	}
	return statMap
}

func (bonus *mobTypeBonus) isNone(statMap map[int32]float64) bool {
	return len(statMap) == 0 && math.Abs(bonus.damageMultiplier-1) <= versusEpsilon && math.Abs(bonus.critMultiplier-1) <= versusEpsilon
}

// The unit's bonuses against each mob type, leaving out a mob type with none. Each lists its sources'
// shares, leaving out a share that nets to none.
func (unit *Unit) mobTypeBonusesProto() []*proto.MobTypeBonus {
	var bonuses []*proto.MobTypeBonus
	for _, mobType := range slices.Sorted(maps.Keys(unit.mobTypeBonuses)) {
		bonus := unit.mobTypeBonuses[mobType]
		statMap := versusStatMap(bonus.stats)
		if bonus.isNone(statMap) {
			continue
		}
		var sources []*proto.MobTypeBonusSource
		for _, share := range unit.versusSources {
			if share.mobType != mobType {
				continue
			}
			if shareStats := versusStatMap(share.stats); !share.isNone(shareStats) {
				sources = append(sources, &proto.MobTypeBonusSource{
					ActionId:             share.source.ToProto(),
					Stats:                shareStats,
					DamageMultiplier:     share.damageMultiplier,
					CritDamageMultiplier: share.critMultiplier,
				})
			}
		}
		bonuses = append(bonuses, &proto.MobTypeBonus{
			MobType:              mobType,
			Stats:                statMap,
			DamageMultiplier:     bonus.damageMultiplier,
			CritDamageMultiplier: bonus.critMultiplier,
			Sources:              sources,
		})
	}
	return bonuses
}

// The tag the combat log puts after "(Threat: X)" on a damage line, naming every versus bonus the hit
// took against its target, or "" when it took none:
//
//	[Versus Undead: {ItemID: 12345} x1.05; {SpellID: 17670} x2 base; {ItemID: 12345} +81 AttackPower]
//
// The tag opens with a space and "[Versus ", then the target's mob type, then " (<creature family>)"
// when an entry matched the target by its family, then ": ". The entries follow, separated by "; ",
// each an ActionID in the log's {SpellID: n} or {ItemID: n} form, a space, and one effect:
//   - "x<m>": a damage multiplier against the mob type, unless the spell ignores attacker modifiers
//     or damage versus creature types;
//   - "x<m> crit": a crit damage multiplier against the mob type, on a crit only;
//   - "+<n>% crit chance" or "-<n>% crit chance": crit chance against the target
//     (AddCritChanceVersus);
//   - "x<m> base": the spell's own VersusBonus on its base damage, under the spell's ActionID;
//   - "x<m> proc chance": the proc that cast the spell is m times as likely against the target
//     (VersusProc), under the proc's ActionID;
//   - "proc only": the proc that cast the spell fires only against targets such as this one;
//   - "+<n> <stat>" or "-<n> <stat>": a stat against the mob type the hit's damage read (VersusStat).
//
// A bonus an item grants, from its stats, its equip spell, its proc or its on-use, names the item; one
// from an enchant, a set bonus, a consumable, a talent or a racial names its spell. The kinds come in
// that order, each in the order the sources were added, and one source's stats in the order of the stat
// enum. Mob types and families are their enum names without the prefix, numbers are rounded to 4
// decimals, and no part holds ";" or "]", so the tag ends the line at its only "]". A snapshot dot
// tick names the sources up at the tick, not the ones its snapshot took.
func (spell *Spell) versusLogTag(result *SpellResult) string {
	if !result.Landed() || result.Damage <= 0 {
		return ""
	}
	target := result.Target
	var entries []string
	if !spell.Flags.Matches(SpellFlagIgnoreAttackerModifiers | SpellFlagNoDamageVersus) {
		for _, share := range spell.Unit.versusSources {
			if share.mobType == target.MobType && math.Abs(share.damageMultiplier-1) > versusEpsilon {
				entries = append(entries, share.source.String()+" x"+versusNumber(share.damageMultiplier))
			}
		}
	}
	if result.DidCrit() {
		for _, share := range spell.Unit.versusSources {
			if share.mobType == target.MobType && math.Abs(share.critMultiplier-1) > versusEpsilon {
				entries = append(entries, share.source.String()+" x"+versusNumber(share.critMultiplier)+" crit")
			}
		}
	}
	byFamily := false
	for _, share := range spell.Unit.AttackTables[target.UnitIndex].versusCritChances {
		if math.Abs(share.percent) > versusEpsilon {
			entries = append(entries, share.source.String()+" "+signedVersusNumber(share.percent)+"% crit chance")
			byFamily = byFamily || share.byFamily
		}
	}
	if mult := spell.VersusBonus.Against(target); mult != 1 {
		entries = append(entries, spell.ActionID.String()+" x"+versusNumber(mult)+" base")
		byFamily = byFamily || spell.VersusBonus.namesFamilyOf(target)
	}
	if proc := spell.VersusProc; proc != nil {
		if mult := proc.Against(target); mult == 0 {
			entries = append(entries, proc.Source.String()+" proc only")
		} else if mult != 1 {
			entries = append(entries, proc.Source.String()+" x"+versusNumber(mult)+" proc chance")
		}
		byFamily = byFamily || proc.namesFamilyOf(target)
	}
	for _, share := range spell.Unit.versusSources {
		if share.mobType != target.MobType {
			continue
		}
		for read := result.versusRead; read != 0; read &= read - 1 {
			stat := stats.Stat(bits.TrailingZeros64(read))
			if value := share.stats[stat]; math.Abs(value) > versusEpsilon {
				entries = append(entries, share.source.String()+" "+signedVersusNumber(value)+" "+stat.StatName())
			}
		}
	}
	if len(entries) == 0 {
		return ""
	}
	family := ""
	if byFamily {
		family = " (" + strings.TrimPrefix(target.CreatureFamily.String(), "CreatureFamily") + ")"
	}
	return " [Versus " + strings.TrimPrefix(target.MobType.String(), "MobType") + family + ": " + strings.Join(entries, "; ") + "]"
}

func versusNumber(v float64) string {
	return strconv.FormatFloat(math.Round(v*1e4)/1e4, 'f', -1, 64)
}

func signedVersusNumber(v float64) string {
	num := versusNumber(v)
	if v > 0 {
		num = "+" + num
	}
	return num
}
