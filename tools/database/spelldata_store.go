package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/tools/database/dbc"
)

// The store's rows as the generator builds them: the fields sim/core/spelldata.Spell, .Effect and
// .Power carry, mirrored here rather than imported from that package, with the generator's own
// bookkeeping alongside them and the JSON shape the captured game files rows are read back in.
//
// The mirror is not what keeps a generated file that does not compile from stopping the generator
// that rewrites it: this package imports the store for the item-proc routing in gen_effects.go, so
// a store that does not build takes gen_spelldata down with it. What keeps one out of the tree is
// the overlay gate in spelldata_write.go, which writes nothing until the sim compiles against the
// rendered files.
//
// Every value is the game files' own column, in the game files' units - the conversions live in the
// store's accessors - and the float fields are kept as the DB states them so the emitted number is
// the one the tooltip resolves from in game.
type storeSpell struct {
	ID   int32
	Name string
	Rank string

	School      core.SpellSchool
	Speed       float64
	MinDuration float64
	Attr        [17]uint32

	SpellLevel, BaseLevel, MaxLevel int16

	CastTimeMs, DurationMs int32
	MinRange, MaxRange     float64

	CooldownMs, CategoryCooldownMs, GCDMs int32

	Category, StartRecoveryCategory, ChargeCategory int16

	DefenseType, DispelType, Mechanic, PreventionType uint8

	MaxStack    int16
	ProcChance  uint8
	ProcCharges int16
	ProcFlags   [2]uint32
	ICDMs       int32

	// Procs per minute, which no game files row states: it is here only when an override supplies it.
	RPPM float64

	// A flat threat bonus for a spell the game files state no E_THREAT effect on, from an override.
	FlatThreat float64

	// An area bonus, from an override.
	AreaBonusGroups                        []int32
	AreaMultiplier, AreaDurationMultiplier float64

	ClassFlags core.ClassFlags

	InterruptFlags                  uint32
	AuraInterrupt, ChannelInterrupt [2]uint32

	StanceMask    uint64
	StanceExclude uint64

	CasterAura, ExcludeCasterAura int32

	MaxTargets         int16
	TargetCreatureType int32

	RequiredAreas int32

	EquipClass                  int8
	EquipSubclass, EquipInvType int32

	Labels []int16
	RefIDs []int32

	// Read off Spell.Description_lang by spelldata_hints.go rather than out of a column: the
	// tooltip's wording is what says whether ProcChance is a roll at all, which effect holds it, and
	// whether the damage is split among the targets.
	ProcChanceSource storeProcChanceSource
	ProcChanceEffect int8
	ProcHint         core.ProcHint
	SplitsDamage     bool

	Effects  []storeEffect
	Powers   []storePower
	Reagents []storeReagent

	// The generator's own bookkeeping, which the store has no field for. The procs-per-minute id is
	// what a PPM override is checked against, the tooltip flag what a proc chance override is, and
	// the notes are the reasons the emitted row states next to a hand-supplied value.
	procsPerMinuteID    int32
	tooltipStatesChance bool
	overrideNotes       []string
}

type storeEffect struct {
	ID, SpellID int32
	Index       uint8

	Type dbcenums.SpellEffectType
	Aura dbcenums.EffectAuraType

	BasePoints float64
	PPL        float64
	Variance   float64

	// The owning spell's SpellLevels, stamped on while the row is built so that the store's Average
	// does not look its owner up. Never captured: it is derived from the captured Levels row.
	BaseLevel, MaxLevel int16 `json:"-"`

	SPCoef  float64
	APCoef  float64
	PvpMult float64

	Amplitude float64
	PeriodMs  int32

	RadiusMin, RadiusMax float64

	Misc, Misc2 int32

	ClassFlags core.ClassFlags

	TriggerID int32

	ChainTargets int16
	ChainAmp     float64

	Mechanic uint8

	PointsPerResource float64

	Target [2]dbcenums.ImplicitTarget

	Attributes int32
}

type storePower struct {
	Type               int8
	Cost, CostPerLevel int32
	CostPct            float64
	PerSecond          int32
}

type storeReagent struct {
	Item  int32
	Count int16
}

// Every table the store reads, loaded once for the whole game files database rather than per spell:
// the closure walks the effects and descriptions of spells it has not selected yet, so the rows
// have to be there before the set of wanted ids is known. Exported because storeInputs embeds it,
// which is what writes the captured rows under these names.
type spellTables struct {
	Names        map[int32]string
	Subtexts     map[int32]string
	Descriptions map[int32]string

	Misc             map[int32]miscRow
	Levels           map[int32]levelsRow
	Cooldowns        map[int32]cooldownRow
	Categories       map[int32]categoryRow
	AuraOptions      map[int32]auraOptionRow
	ClassOptions     map[int32]core.ClassFlags
	Interrupts       map[int32]interruptRow
	Shapeshift       map[int32]shapeshiftRow
	AuraRestrictions map[int32]auraRestrictionRow
	Targets          map[int32]int16
	CreatureType     map[int32]int32
	Requirements     map[int32]int32
	Equipped         map[int32]equippedRow

	Labels   map[int32][]int16
	Powers   map[int32][]storePower
	Reagents map[int32][]storeReagent
	Effects  map[int32][]storeEffect

	// The spell granting the enchant, by each enchant equip spell that answers a hit.
	EnchantGrants map[int32]int32

	// The chance the enchantments casting a combat spell state for it, by the combat spell.
	EnchantChances map[int32]enchantChance

	// The tooltip references by spell id, filled as referencedIDs reads them: the closure asks for
	// the same spell on every pass over the store.
	refs map[int32][]int32
}

type miscRow struct {
	Attr                   [17]uint32
	School                 core.SpellSchool
	Speed                  float64
	CastTimeMs, DurationMs int32
	MinRange, MaxRange     float64
	MinDuration            float64 `json:",omitempty"` // Stated on a few dozen rows, so the capture leaves the rest without it.
}

type levelsRow struct {
	SpellLevel, BaseLevel, MaxLevel int16
}

type cooldownRow struct {
	CooldownMs, CategoryCooldownMs, GCDMs int32
}

type categoryRow struct {
	Category, StartRecoveryCategory, ChargeCategory   int16
	DefenseType, DispelType, Mechanic, PreventionType uint8
}

type auraOptionRow struct {
	MaxStack         int16
	ProcChance       uint8
	ProcCharges      int16
	ProcFlags        [2]uint32
	ICDMs            int32
	ProcsPerMinuteID int32
}

type interruptRow struct {
	InterruptFlags                  uint32
	AuraInterrupt, ChannelInterrupt [2]uint32
}

type shapeshiftRow struct {
	Mask, Exclude uint64
}

type auraRestrictionRow struct {
	CasterAura, ExcludeCasterAura int32
}

type equippedRow struct {
	Class              int8
	Subclass, InvTypes int32
}

type enchantChance struct {
	Chance   uint8
	Enchants []int32
}

func newSpellTables() *spellTables {
	return &spellTables{
		Names:            map[int32]string{},
		Subtexts:         map[int32]string{},
		Descriptions:     map[int32]string{},
		Misc:             map[int32]miscRow{},
		Levels:           map[int32]levelsRow{},
		Cooldowns:        map[int32]cooldownRow{},
		Categories:       map[int32]categoryRow{},
		AuraOptions:      map[int32]auraOptionRow{},
		ClassOptions:     map[int32]core.ClassFlags{},
		Interrupts:       map[int32]interruptRow{},
		Shapeshift:       map[int32]shapeshiftRow{},
		AuraRestrictions: map[int32]auraRestrictionRow{},
		Targets:          map[int32]int16{},
		CreatureType:     map[int32]int32{},
		Requirements:     map[int32]int32{},
		Equipped:         map[int32]equippedRow{},
		Labels:           map[int32][]int16{},
		Powers:           map[int32][]storePower{},
		Reagents:         map[int32][]storeReagent{},
		Effects:          map[int32][]storeEffect{},
		EnchantGrants:    map[int32]int32{},
		EnchantChances:   map[int32]enchantChance{},
	}
}

// A spell's rows exist once per difficulty on a few hundred spells - 29213 caps 20 targets at
// difficulty 0 and 10 at 186 - and difficulty 0 is the one the sim plays.
func loadSpellTables(db *sql.DB) (*spellTables, error) {
	t := newSpellTables()
	for _, load := range []func(*sql.DB) error{
		t.loadNames, t.loadMisc, t.loadLevels, t.loadCooldowns, t.loadCategories, t.loadAuraOptions,
		t.loadClassOptions, t.loadInterrupts, t.loadShapeshift, t.loadAuraRestrictions, t.loadTargetRestrictions,
		t.loadCastingRequirements, t.loadEquippedItems, t.loadLabels, t.loadPowers, t.loadReagents, t.loadEffects,
		t.loadEnchantGrants, t.loadEnchantChances,
	} {
		if err := load(db); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// The row the store carries for a spell, from the tables loaded above. The description is read here
// too, for the ids the tooltip names, but never stored: the store is data the sim reads, and the
// tooltip text is generator input.
func (t *spellTables) row(id int32) storeSpell {
	s := storeSpell{ID: id, Name: t.Names[id], Rank: t.Subtexts[id]}

	m := t.Misc[id]
	s.School, s.Speed, s.MinDuration, s.Attr = m.School, m.Speed, m.MinDuration, m.Attr
	s.CastTimeMs, s.DurationMs = m.CastTimeMs, m.DurationMs
	s.MinRange, s.MaxRange = m.MinRange, m.MaxRange

	// No SpellLevels row is the game files' "no level scaling", which is the spell's own and base
	// level at the cap and no maximum.
	if l, ok := t.Levels[id]; ok {
		s.SpellLevel, s.BaseLevel, s.MaxLevel = l.SpellLevel, l.BaseLevel, l.MaxLevel
	} else {
		s.SpellLevel, s.BaseLevel = RankLevel, RankLevel
	}

	c := t.Cooldowns[id]
	s.CooldownMs, s.CategoryCooldownMs, s.GCDMs = c.CooldownMs, c.CategoryCooldownMs, c.GCDMs

	cat := t.Categories[id]
	s.Category, s.StartRecoveryCategory, s.ChargeCategory = cat.Category, cat.StartRecoveryCategory, cat.ChargeCategory
	s.DefenseType, s.DispelType = cat.DefenseType, cat.DispelType
	s.Mechanic, s.PreventionType = cat.Mechanic, cat.PreventionType

	a := t.AuraOptions[id]
	s.MaxStack, s.ProcChance, s.ProcCharges = a.MaxStack, a.ProcChance, a.ProcCharges
	s.ProcFlags, s.ICDMs = a.ProcFlags, a.ICDMs
	s.procsPerMinuteID = a.ProcsPerMinuteID

	s.ClassFlags = t.ClassOptions[id]

	i := t.Interrupts[id]
	s.InterruptFlags, s.AuraInterrupt, s.ChannelInterrupt = i.InterruptFlags, i.AuraInterrupt, i.ChannelInterrupt

	ss := t.Shapeshift[id]
	s.StanceMask, s.StanceExclude = ss.Mask, ss.Exclude

	ar := t.AuraRestrictions[id]
	s.CasterAura, s.ExcludeCasterAura = ar.CasterAura, ar.ExcludeCasterAura

	s.MaxTargets = t.Targets[id]
	s.TargetCreatureType = t.CreatureType[id]
	s.RequiredAreas = t.Requirements[id]

	e := t.Equipped[id]
	s.EquipClass, s.EquipSubclass, s.EquipInvType = e.Class, e.Subclass, e.InvTypes

	s.Labels = t.Labels[id]
	s.RefIDs = t.referencedIDs(id)
	s.Powers = t.Powers[id]
	s.Reagents = t.Reagents[id]

	// The effect's class mask is read against the owning spell's family: the mask words alone name
	// nothing, since the same bit is a different spell in each family. The family is carried only
	// where the effect states a mask, because Matches needs an overlapping mask word as well and a
	// family on its own can never produce one.
	//
	// The spell's levels go onto every effect for the same reason the family does: the amount an
	// effect scales to is priced from them, and the effect is what the caller holds.
	s.Effects = make([]storeEffect, len(t.Effects[id]))
	copy(s.Effects, t.Effects[id])
	for i := range s.Effects {
		if !s.Effects[i].ClassFlags.IsZero() {
			s.Effects[i].ClassFlags.Family = s.ClassFlags.Family
		}
		s.Effects[i].BaseLevel, s.Effects[i].MaxLevel = s.BaseLevel, s.MaxLevel
	}

	return s
}

func (t *spellTables) loadNames(db *sql.DB) error {
	if err := eachRow(db, `SELECT ID, Name_lang FROM SpellName ORDER BY ID`, func(rows *sql.Rows) error {
		var id int32
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		t.Names[id] = name
		return nil
	}); err != nil {
		return err
	}

	return eachRow(db, `
		SELECT ID, COALESCE(NameSubtext_lang, ''), COALESCE(Description_lang, '')
		FROM Spell ORDER BY ID`, func(rows *sql.Rows) error {
		var id int32
		var subtext, description string
		if err := rows.Scan(&id, &subtext, &description); err != nil {
			return err
		}
		t.Subtexts[id] = subtext
		t.Descriptions[id] = description
		return nil
	})
}

func (t *spellTables) loadMisc(db *sql.DB) error {
	return eachRow(db, `
		SELECT m.SpellID,
		       m.Attributes_0, m.Attributes_1, m.Attributes_2, m.Attributes_3, m.Attributes_4,
		       m.Attributes_5, m.Attributes_6, m.Attributes_7, m.Attributes_8, m.Attributes_9,
		       m.Attributes_10, m.Attributes_11, m.Attributes_12, m.Attributes_13, m.Attributes_14,
		       m.Attributes_15, m.Attributes_16,
		       COALESCE(m.SchoolMask, 0), COALESCE(m.Speed, 0),
		       COALESCE(ct.Base, 0), COALESCE(d.Duration, 0),
		       COALESCE(r.RangeMin_0, 0), COALESCE(r.RangeMax_0, 0), COALESCE(m.MinDuration, 0)
		FROM SpellMisc m
		LEFT JOIN SpellCastTimes ct ON ct.ID = m.CastingTimeIndex
		LEFT JOIN SpellDuration d ON d.ID = m.DurationIndex
		LEFT JOIN SpellRange r ON r.ID = m.RangeIndex
		WHERE m.DifficultyID = 0
		ORDER BY m.SpellID`, func(rows *sql.Rows) error {
		var id int32
		var attr [17]int64
		var m miscRow
		var school int64
		dest := []any{&id}
		for i := range attr {
			dest = append(dest, &attr[i])
		}
		dest = append(dest, &school, &m.Speed, &m.CastTimeMs, &m.DurationMs, &m.MinRange, &m.MaxRange, &m.MinDuration)
		if err := rows.Scan(dest...); err != nil {
			return err
		}
		for i, word := range attr {
			m.Attr[i] = uint32(word)
		}
		m.School = core.SpellSchool(school)
		return putOnce(t.Misc, id, m, "SpellMisc rows at difficulty 0")
	})
}

// A table the game files state once per spell. A second row is an error rather than a silent
// overwrite: the store would carry whichever came last.
func putOnce[V any](into map[int32]V, id int32, v V, rows string) error {
	if _, dup := into[id]; dup {
		return fmt.Errorf("spell %d has two %s", id, rows)
	}
	into[id] = v
	return nil
}

func (t *spellTables) loadLevels(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, SpellLevel, BaseLevel, MaxLevel
		FROM SpellLevels WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var l levelsRow
		if err := rows.Scan(&id, &l.SpellLevel, &l.BaseLevel, &l.MaxLevel); err != nil {
			return err
		}
		return putOnce(t.Levels, id, l, "SpellLevels rows at difficulty 0")
	})
}

func (t *spellTables) loadCooldowns(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(RecoveryTime, 0), COALESCE(CategoryRecoveryTime, 0),
		       COALESCE(StartRecoveryTime, 0)
		FROM SpellCooldowns WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var c cooldownRow
		if err := rows.Scan(&id, &c.CooldownMs, &c.CategoryCooldownMs, &c.GCDMs); err != nil {
			return err
		}
		return putOnce(t.Cooldowns, id, c, "SpellCooldowns rows at difficulty 0")
	})
}

func (t *spellTables) loadCategories(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(Category, 0), COALESCE(StartRecoveryCategory, 0),
		       COALESCE(ChargeCategory, 0), COALESCE(DefenseType, 0), COALESCE(DispelType, 0),
		       COALESCE(Mechanic, 0), COALESCE(PreventionType, 0)
		FROM SpellCategories WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var c categoryRow
		var defense, dispel, mechanic, prevention int64
		if err := rows.Scan(&id, &c.Category, &c.StartRecoveryCategory, &c.ChargeCategory,
			&defense, &dispel, &mechanic, &prevention); err != nil {
			return err
		}
		c.DefenseType, c.DispelType = uint8(defense), uint8(dispel)
		c.Mechanic, c.PreventionType = uint8(mechanic), uint8(prevention)
		return putOnce(t.Categories, id, c, "SpellCategories rows at difficulty 0")
	})
}

func (t *spellTables) loadAuraOptions(db *sql.DB) error {
	// ProcTypeMask_0 and _1 are generated columns over the JSON array, so they are named here rather
	// than reached through SELECT *.
	return eachRow(db, `
		SELECT SpellID, COALESCE(CumulativeAura, 0), COALESCE(ProcChance, 0), COALESCE(ProcCharges, 0),
		       COALESCE(ProcTypeMask_0, 0), COALESCE(ProcTypeMask_1, 0), COALESCE(ProcCategoryRecovery, 0),
		       COALESCE(SpellProcsPerMinuteID, 0)
		FROM SpellAuraOptions WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var a auraOptionRow
		var chance, mask0, mask1 int64
		if err := rows.Scan(&id, &a.MaxStack, &chance, &a.ProcCharges, &mask0, &mask1, &a.ICDMs,
			&a.ProcsPerMinuteID); err != nil {
			return err
		}
		a.ProcChance = uint8(chance)
		a.ProcFlags = [2]uint32{uint32(mask0), uint32(mask1)}
		return putOnce(t.AuraOptions, id, a, "SpellAuraOptions rows at difficulty 0")
	})
}

func (t *spellTables) loadClassOptions(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(SpellClassSet, 0), COALESCE(SpellClassMask_0, 0),
		       COALESCE(SpellClassMask_1, 0), COALESCE(SpellClassMask_2, 0), COALESCE(SpellClassMask_3, 0)
		FROM SpellClassOptions ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var f core.ClassFlags
		var mask [4]int64
		if err := rows.Scan(&id, &f.Family, &mask[0], &mask[1], &mask[2], &mask[3]); err != nil {
			return err
		}
		for i, word := range mask {
			f.Mask[i] = uint32(word)
		}
		return putOnce(t.ClassOptions, id, f, "SpellClassOptions rows")
	})
}

func (t *spellTables) loadInterrupts(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(InterruptFlags, 0), COALESCE(AuraInterruptFlags_0, 0), COALESCE(AuraInterruptFlags_1, 0),
		       COALESCE(ChannelInterruptFlags_0, 0), COALESCE(ChannelInterruptFlags_1, 0)
		FROM SpellInterrupts WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var flags, aura0, aura1, channel0, channel1 int64
		if err := rows.Scan(&id, &flags, &aura0, &aura1, &channel0, &channel1); err != nil {
			return err
		}
		return putOnce(t.Interrupts, id, interruptRow{
			InterruptFlags:   uint32(flags),
			AuraInterrupt:    [2]uint32{uint32(aura0), uint32(aura1)},
			ChannelInterrupt: [2]uint32{uint32(channel0), uint32(channel1)},
		}, "SpellInterrupts rows at difficulty 0")
	})
}

// The two mask words are one 64-bit set of forms, which is how core reads a stance mask.
func (t *spellTables) loadShapeshift(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(ShapeshiftMask_0, 0), COALESCE(ShapeshiftMask_1, 0),
		       COALESCE(ShapeshiftExclude_0, 0), COALESCE(ShapeshiftExclude_1, 0)
		FROM SpellShapeshift ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var low, high, excludeLow, excludeHigh int64
		if err := rows.Scan(&id, &low, &high, &excludeLow, &excludeHigh); err != nil {
			return err
		}
		return putOnce(t.Shapeshift, id, shapeshiftRow{
			Mask:    uint64(uint32(low)) | uint64(uint32(high))<<32,
			Exclude: uint64(uint32(excludeLow)) | uint64(uint32(excludeHigh))<<32,
		}, "SpellShapeshift rows")
	})
}

// The aura a spell needs on its caster to be cast, and the one that bars it, where the game files
// state neither as a form: Shifting Power names Cat Form here beside its SpellShapeshift mask.
func (t *spellTables) loadAuraRestrictions(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(CasterAuraSpell, 0), COALESCE(ExcludeCasterAuraSpell, 0)
		FROM SpellAuraRestrictions WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var r auraRestrictionRow
		if err := rows.Scan(&id, &r.CasterAura, &r.ExcludeCasterAura); err != nil {
			return err
		}
		return putOnce(t.AuraRestrictions, id, r, "SpellAuraRestrictions rows at difficulty 0")
	})
}

func (t *spellTables) loadTargetRestrictions(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(MaxTargets, 0), COALESCE(TargetCreatureType, 0)
		FROM SpellTargetRestrictions WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var maxTargets int16
		var creatureType int32
		if err := rows.Scan(&id, &maxTargets, &creatureType); err != nil {
			return err
		}
		if err := putOnce(t.Targets, id, maxTargets, "SpellTargetRestrictions rows at difficulty 0"); err != nil {
			return err
		}
		if creatureType != 0 {
			t.CreatureType[id] = creatureType
		}
		return nil
	})
}

func (t *spellTables) loadCastingRequirements(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(RequiredAreasID, 0)
		FROM SpellCastingRequirements WHERE RequiredAreasID != 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id, areas int32
		if err := rows.Scan(&id, &areas); err != nil {
			return err
		}
		if _, dup := t.Requirements[id]; dup {
			return fmt.Errorf("spell %d has two SpellCastingRequirements rows naming an area group", id)
		}
		t.Requirements[id] = areas
		return nil
	})
}

func (t *spellTables) loadEquippedItems(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(EquippedItemClass, 0), COALESCE(EquippedItemSubclass, 0),
		       COALESCE(EquippedItemInvTypes, 0)
		FROM SpellEquippedItems ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var class int64
		var e equippedRow
		if err := rows.Scan(&id, &class, &e.Subclass, &e.InvTypes); err != nil {
			return err
		}
		e.Class = int8(class)
		return putOnce(t.Equipped, id, e, "SpellEquippedItems rows")
	})
}

func (t *spellTables) loadLabels(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, LabelID FROM SpellLabel ORDER BY SpellID, LabelID`, func(rows *sql.Rows) error {
		var id int32
		var label int16
		if err := rows.Scan(&id, &label); err != nil {
			return err
		}
		t.Labels[id] = append(t.Labels[id], label)
		return nil
	})
}

// Every bar the spell costs from, in the game files' order, so Power(t) finds the one it asks for.
func (t *spellTables) loadPowers(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(PowerType, 0), COALESCE(ManaCost, 0), COALESCE(ManaCostPerLevel, 0),
		       COALESCE(PowerCostPct, 0), COALESCE(ManaPerSecond, 0)
		FROM SpellPower ORDER BY SpellID, OrderIndex`, func(rows *sql.Rows) error {
		var id int32
		var powerType int64
		var p storePower
		if err := rows.Scan(&id, &powerType, &p.Cost, &p.CostPerLevel, &p.CostPct, &p.PerSecond); err != nil {
			return err
		}
		p.Type = int8(powerType)
		t.Powers[id] = append(t.Powers[id], p)
		return nil
	})
}

// Reagent and ReagentCount are JSON arrays of 8 slots, read in slot order; an empty slot is item 0.
func (t *spellTables) loadReagents(db *sql.DB) error {
	return eachRow(db, `
		SELECT s.SpellID, item.value, count.value
		FROM SpellReagents s, json_each(s.Reagent) item
		JOIN json_each(s.ReagentCount) count ON count.key = item.key
		WHERE item.value != 0 AND count.value != 0
		ORDER BY s.SpellID, item.key`, func(rows *sql.Rows) error {
		var id int32
		var r storeReagent
		if err := rows.Scan(&id, &r.Item, &r.Count); err != nil {
			return err
		}
		t.Reagents[id] = append(t.Reagents[id], r)
		return nil
	})
}

// EffectMiscValue_0/_1, EffectSpellClassMask_0..3, EffectRadiusIndex_0/_1 and ImplicitTarget_0/_1 are
// generated columns over the JSON arrays, so each is named rather than reached through SELECT *.
// RadiusMin and RadiusMax are SpellRadius' own columns, both from the radius slot with the larger
// maximum, the first on a tie: most areas around the caster state theirs in the second slot only
// (Arcane Explosion's [22,15]). Two radius rows state a Radius with no maximum, and those read as
// zero here rather than as a radius the game files do not state there.
func (t *spellTables) loadEffects(db *sql.DB) error {
	return eachRow(db, `
		SELECT e.SpellID, e.ID, e.EffectIndex, e.Effect, e.EffectAura,
		       COALESCE(e.EffectBasePointsF, 0), COALESCE(e.EffectRealPointsPerLevel, 0),
		       COALESCE(e.Variance, 0), COALESCE(e.EffectBonusCoefficient, 0),
		       COALESCE(e.BonusCoefficientFromAP, 0), COALESCE(e.PvpMultiplier, 0),
		       COALESCE(e.EffectAmplitude, 0), COALESCE(e.EffectAuraPeriod, 0),
		       CASE WHEN COALESCE(r1.RadiusMax, 0) > COALESCE(r0.RadiusMax, 0)
		            THEN COALESCE(r1.RadiusMin, 0) ELSE COALESCE(r0.RadiusMin, 0) END,
		       MAX(COALESCE(r0.RadiusMax, 0), COALESCE(r1.RadiusMax, 0)),
		       COALESCE(e.EffectMiscValue_0, 0), COALESCE(e.EffectMiscValue_1, 0),
		       COALESCE(e.EffectSpellClassMask_0, 0), COALESCE(e.EffectSpellClassMask_1, 0),
		       COALESCE(e.EffectSpellClassMask_2, 0), COALESCE(e.EffectSpellClassMask_3, 0),
		       COALESCE(e.EffectTriggerSpell, 0), COALESCE(e.EffectChainTargets, 0),
		       COALESCE(e.EffectChainAmplitude, 0), COALESCE(e.EffectMechanic, 0),
		       COALESCE(e.EffectPointsPerResource, 0),
		       COALESCE(e.ImplicitTarget_0, 0), COALESCE(e.ImplicitTarget_1, 0),
		       COALESCE(e.EffectAttributes, 0)
		FROM SpellEffect e
		LEFT JOIN SpellRadius r0 ON r0.ID = e.EffectRadiusIndex_0
		LEFT JOIN SpellRadius r1 ON r1.ID = e.EffectRadiusIndex_1
		WHERE e.DifficultyID = 0
		ORDER BY e.SpellID, e.EffectIndex, e.ID`, func(rows *sql.Rows) error {
		var e storeEffect
		var index, mask [4]int64
		var mechanic, target0, target1 int64
		if err := rows.Scan(&e.SpellID, &e.ID, &index[0], &e.Type, &e.Aura,
			&e.BasePoints, &e.PPL, &e.Variance, &e.SPCoef, &e.APCoef, &e.PvpMult,
			&e.Amplitude, &e.PeriodMs, &e.RadiusMin, &e.RadiusMax,
			&e.Misc, &e.Misc2,
			&mask[0], &mask[1], &mask[2], &mask[3],
			&e.TriggerID, &e.ChainTargets, &e.ChainAmp, &mechanic, &e.PointsPerResource,
			&target0, &target1, &e.Attributes); err != nil {
			return err
		}
		e.Index = uint8(index[0])
		for i, word := range mask {
			e.ClassFlags.Mask[i] = uint32(word)
		}
		e.Mechanic = uint8(mechanic)
		e.Target = [2]dbcenums.ImplicitTarget{dbcenums.ImplicitTarget(target0), dbcenums.ImplicitTarget(target1)}

		// The rows arrive in index order, so a second row at one index is the one just read. The
		// store would carry both, and the position EffectN counts by would answer the first.
		if prior := t.Effects[e.SpellID]; len(prior) > 0 && prior[len(prior)-1].Index == e.Index {
			return fmt.Errorf("spell %d states effect index %d twice, as rows %d and %d",
				e.SpellID, e.Index, prior[len(prior)-1].ID, e.ID)
		}

		t.Effects[e.SpellID] = append(t.Effects[e.SpellID], e)
		return nil
	})
}

// The three effect slots of every SpellItemEnchantment row, one row per slot.
const enchantSlotsCTE = `
	WITH slots AS (
		SELECT ID, Effect_0 AS Effect, EffectPointsMin_0 AS Points, EffectArg_0 AS SpellID FROM SpellItemEnchantment
		UNION ALL SELECT ID, Effect_1, EffectPointsMin_1, EffectArg_1 FROM SpellItemEnchantment
		UNION ALL SELECT ID, Effect_2, EffectPointsMin_2, EffectArg_2 FROM SpellItemEnchantment)`

// The equip spells (Effect 3) of each SpellItemEnchantment row that answer a hit, with the
// E_ENCHANT_ITEM spell granting the enchant. Several spells can grant one enchant - a recipe
// re-taught by a later expansion - and several enchants can share an equip spell; the lowest grant
// is the one kept, as enchantGrantEffects keeps it.
func (t *spellTables) loadEnchantGrants(db *sql.DB) error {
	auras := make([]string, len(dbc.EnchantProcAuras))
	for i, aura := range dbc.EnchantProcAuras {
		auras[i] = strconv.Itoa(int(aura))
	}

	return eachRow(db, enchantSlotsCTE+fmt.Sprintf(`
		SELECT s.SpellID, min(g.SpellID)
		FROM slots s
		JOIN SpellEffect g ON g.Effect = %d AND g.EffectMiscValue_0 = s.ID
		WHERE s.Effect = %d AND s.SpellID > 0 AND EXISTS (
			SELECT 1 FROM SpellEffect a
			WHERE a.SpellID = s.SpellID AND a.DifficultyID = 0 AND a.EffectAura IN (%s))
		GROUP BY s.SpellID ORDER BY s.SpellID`,
		dbcenums.E_ENCHANT_ITEM, dbc.ITEM_ENCHANTMENT_EQUIP_SPELL, strings.Join(auras, ", ")),
		func(rows *sql.Rows) error {
			var id, grant int32
			if err := rows.Scan(&id, &grant); err != nil {
				return err
			}
			t.EnchantGrants[id] = grant
			return nil
		})
}

// The chance each combat spell (Effect 1) of a SpellItemEnchantment row is cast at, which the
// enchantment states in EffectPointsMin rather than the spell in its own column: Fiery Blaze 36
// casts 6297 at 15. Enchantments that cast one spell at two chances are an error: the store has one
// row to state it on.
func (t *spellTables) loadEnchantChances(db *sql.DB) error {
	return eachRow(db, enchantSlotsCTE+fmt.Sprintf(`
		SELECT SpellID, Points, ID FROM slots
		WHERE Effect = %d AND SpellID > 0 AND Points > 0
		ORDER BY SpellID, ID`, dbc.ITEM_ENCHANTMENT_COMBAT_SPELL), func(rows *sql.Rows) error {
		var id, enchant int32
		var chance uint8
		if err := rows.Scan(&id, &chance, &enchant); err != nil {
			return err
		}
		stated, ok := t.EnchantChances[id]
		if ok && stated.Chance != chance {
			return fmt.Errorf("enchantments %v cast spell %d at %d%% and enchantment %d at %d%%",
				stated.Enchants, id, stated.Chance, enchant, chance)
		}
		t.EnchantChances[id] = enchantChance{Chance: chance, Enchants: append(stated.Enchants, enchant)}
		return nil
	})
}

// The spell ids the tooltip names, in the order it names them, each once. An id no SpellName row
// carries is left out: the closure walks the same tokens and only follows the ones that are spells.
func (t *spellTables) referencedIDs(id int32) []int32 {
	if refs, read := t.refs[id]; read {
		return refs
	}

	var refs []int32
	seen := map[int32]bool{}
	for _, m := range descriptionSpellRef.FindAllStringSubmatch(t.Descriptions[id], -1) {
		ref := parseSpellID(m[1])
		if ref == 0 || ref == id || seen[ref] {
			continue
		}
		seen[ref] = true
		if _, named := t.Names[ref]; named {
			refs = append(refs, ref)
		}
	}

	if t.refs == nil {
		t.refs = map[int32][]int32{}
	}
	t.refs[id] = refs
	return refs
}

func eachRow(db *sql.DB, query string, scan func(*sql.Rows) error) error {
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("%s: %w", query, err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := scan(rows); err != nil {
			return fmt.Errorf("%s: %w", query, err)
		}
	}
	return rows.Err()
}
