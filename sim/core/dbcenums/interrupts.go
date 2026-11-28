package dbcenums

// SpellInterrupts.InterruptFlags bits.
const (
	// Damage taken while casting pushes the cast back. The bombs state 0x5 without it, and the
	// grenades and dynamite have no row at all.
	SPELL_INTERRUPT_FLAG_PUSHBACK uint32 = 0x2
)
