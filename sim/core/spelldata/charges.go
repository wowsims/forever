package spelldata

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// Attaches to a charged buff the listener that spends its charges, as the buff's row states it, and
// returns that listener. A cast spends one charge however many targets it hits. The spells that
// spend one are the ones the row's proc effect names; a row whose proc effect names none states them
// on its spell modifiers instead, and the charge goes to the spells the buff modifies.
//
// A channel is one cast: the charge stays up for every tick and every target, and is spent when the
// channel ends. A channel the buff's class flags name spends it once the channel has started, even
// when its cast lands nothing itself, and so does one already running when the charge comes. Wand Shoot
// is no such channel: each shot spends a charge like any hit, and so do other hits while wanding.
// TODO: A tick that lands after its channel ends, a missile still in flight, misses the charge.
func AttachChargeSpender(character *core.Character, buff *core.Aura, s *Spell, name string, opts ...ProcOpt) core.ProcTrigger {
	classFlags := chargeClassFlags(character, s)
	firstHit, resetFirstHit := core.OncePerCast()
	channelHit := false
	buff.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
		resetFirstHit()
		channel := character.ChanneledDot
		// A channel with no ticks left only waits out the GCD.
		channelHit = channel != nil && channel.RemainingTicks() > 0 && channel.Spell.MatchesFlags(classFlags)
	})

	character.OnChannelEnd(func(sim *core.Simulation, _ *core.Dot) {
		if channelHit && buff.IsActive() {
			channelHit = false
			buff.RemoveStack(sim)
		}
	})

	spender := ProcTrigger(character, s, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.RemoveStack(sim)
	}, opts...)
	spender.Name = name
	spender.ClassFlags = classFlags
	spender.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		// A death the row's flags name is no cast, so it spends a charge of its own.
		if spell == nil {
			return true
		}
		// The channel's own hit lands before its dot starts, its ticks while it runs. A hit the wearer
		// takes is not part of the channel.
		channel := character.ChanneledDot
		inChannel := channel != nil && !character.Wand.IsShoot(channel.Spell)
		if spell.Unit == &character.Unit && !character.Wand.IsShoot(spell) && (inChannel || spell.Flags.Matches(core.SpellFlagChanneled)) {
			channelHit = true
			return false
		}
		return firstHit(sim, spell)
	}

	buff.AttachProcTriggerCallback(&character.Unit, spender)

	if !spender.ClassFlags.IsZero() {
		buff.AttachProcTriggerCallback(&character.Unit, core.ProcTrigger{
			Name:               name + " - Channel",
			Callback:           core.CallbackOnCastComplete,
			SpellFlags:         core.SpellFlagChanneled,
			ClassFlags:         spender.ClassFlags,
			TriggerImmediately: true,
			ExtraCondition: func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
				return character.IsChanneling()
			},
			Handler: func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				channelHit = true
			},
		})
	}
	return spender
}

func chargeClassFlags(character *core.Character, s *Spell) core.ClassFlags {
	flags := rowClassFlags(s)
	if flags.IsZero() {
		flags = spellModClassFlags(s)
	}
	if othersFamily(character, flags) {
		return core.ClassFlags{}
	}
	return flags
}

func spellModClassFlags(s *Spell) core.ClassFlags {
	for i := range s.Effects {
		e := &s.Effects[i]
		if (e.Aura == dbcenums.A_ADD_FLAT_MODIFIER || e.Aura == dbcenums.A_ADD_PCT_MODIFIER) && !e.ClassFlags.IsZero() {
			return e.ClassFlags
		}
	}
	return core.ClassFlags{}
}
