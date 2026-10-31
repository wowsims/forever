package spelldata

import "github.com/wowsims/forever/sim/core/dbcenums"

// Whether the spell sets bit in Attributes[word]. A word outside the game files' 17 reads as unset.
func (s *Spell) HasAttr(word int, bit uint32) bool {
	if word < 0 || word >= len(s.Attr) {
		return false
	}
	return s.Attr[word]&bit != 0
}

func (s *Spell) IsPassive() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_BASE, dbcenums.ATTR_PASSIVE)
}

func (s *Spell) NotShapeshifted() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_BASE, dbcenums.ATTR_NOT_SHAPESHIFTED)
}

func (s *Spell) CastableInCasterForm() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_2, dbcenums.ATTR_EX_2_CASTABLE_IN_CASTER_FORM)
}

func (s *Spell) IsChanneled() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_1, dbcenums.ATTR_EX_1_IS_CHANNELLED|dbcenums.ATTR_EX_1_IS_SELF_CHANNELLED)
}

// Whether damage taken while casting the spell pushes its cast back.
func (s *Spell) PushedBack() bool {
	return s.InterruptFlags&dbcenums.SPELL_INTERRUPT_FLAG_PUSHBACK != 0 &&
		!s.HasAttr(dbcenums.ATTR_INDEX_EX_6, dbcenums.ATTR_EX_6_NO_PUSHBACK)
}

// Whether completing a cast of the spell leaves the melee swing timer running instead of starting
// a fresh swing.
func (s *Spell) KeepsCombatTimers() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_2, dbcenums.ATTR_EX_2_DO_NOT_RESET_COMBAT_TIMERS)
}

func (s *Spell) IsBleed() bool {
	return s.Mechanic == dbcenums.MECHANIC_BLEED
}

func (s *Spell) IsStun() bool {
	return s.Mechanic == dbcenums.MECHANIC_STUN
}

// Whether the spell's periodic damage is a bleed: the mechanic sits on the spell, or on the effect
// alone where the spell also hits directly.
func (s *Spell) AppliesBleed() bool {
	return s.IsBleed() || s.PeriodicDamageEffect().Mechanic == dbcenums.MECHANIC_BLEED
}

func (s *Spell) RefundsOnMiss() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_1, dbcenums.ATTR_EX_1_DISCOUNT_POWER_ON_MISS)
}

func (s *Spell) GeneratesNoThreat() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_1, dbcenums.ATTR_EX_1_NO_THREAT)
}

func (s *Spell) PeriodicCanCrit() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_8, dbcenums.ATTR_EX_8_PERIODIC_CAN_CRIT)
}

func (s *Spell) CanProcFromProcs() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_3, dbcenums.ATTR_EX_3_CAN_PROC_FROM_PROCS)
}

// Whether only a player can carry the spell's aura, which leaves a pet without it.
func (s *Spell) OnlyOnPlayers() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_3, dbcenums.ATTR_EX_3_ONLY_ON_PLAYER)
}

func (s *Spell) ClassSpellsOnly() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_12, dbcenums.ATTR_EX_12_ONLY_PROC_FROM_CLASS_ABILITIES)
}

// Whether the game files bar the spell from critting, which picks the no-crit hit table.
func (s *Spell) CannotCrit() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_2, dbcenums.ATTR_EX_2_CANT_CRIT)
}

// Whether the game files flag the spell Always Hit, which picks an outcome with no miss roll.
func (s *Spell) AlwaysHits() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_3, dbcenums.ATTR_EX_3_ALWAYS_HIT)
}

// Whether the game files flag the spell No Active Defense, which keeps dodge, parry and block out of
// its hit roll.
func (s *Spell) NoActiveDefense() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_BASE, dbcenums.ATTR_NO_ACTIVE_DEFENSE)
}

// Whether the game files flag the spell Fixed Travel Time, which makes its Speed the missile's flight
// time in seconds rather than yards per second.
func (s *Spell) FixedTravelTime() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_9, dbcenums.ATTR_EX_9_FIXED_TRAVEL_TIME)
}

// Whether the spell's hits count as a proc, which is what a listener that cannot proc from procs
// skips. The game files state the negative, so this answers it as the sim's SpellFlagProc reads.
func (s *Spell) IsAProc() bool {
	return !s.HasAttr(dbcenums.ATTR_INDEX_EX_3, dbcenums.ATTR_EX_3_NOT_A_PROC)
}

func (s *Spell) SuppressesWeaponProcs() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_4, dbcenums.ATTR_EX_4_SUPPRESS_WEAPON_PROCS)
}

// Whether the spell's aura hears hits the way a weapon proc does, which means skipping the hits of
// a spell flagged Suppress Weapon Procs. The flag sits on the listener; SuppressesWeaponProcs above
// sits on the spell whose hits are skipped.
func (s *Spell) IsWeaponProcAura() bool {
	return s.HasAttr(dbcenums.ATTR_INDEX_EX_6, dbcenums.ATTR_EX_6_AURA_IS_WEAPON_PROC)
}

// The share of the cost a miss refunds, for RageCostOptions.Refund: 80% where the game files flag
// Discount Power On Miss, nothing otherwise.
func (s *Spell) MissRefund() float64 {
	if s.RefundsOnMiss() {
		return 0.8
	}
	return 0
}
