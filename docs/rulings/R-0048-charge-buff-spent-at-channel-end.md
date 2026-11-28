# R-0048: A charge buff that names its spells is used at the end of a channel that hits it

- Skill: wowsims-procs
- Code: `sim/core/spelldata/charges.go`, `AttachChargeSpender`, `sim/common/forever/enchants.go`
- Pull request: #658
- Evidence: EV-0149, EV-0162

A charge buff hit by the wearer's channel is reserved and used at that channel's end (`Unit.OnChannelEnd`), an
interrupted channel included, and a charge gained while a channel its class flags name is running is used at that
channel's end.

This holds for a buff whose row names the spells that use it by class flags. Revelation keeps nothing for a channel's
cast: the channel's first tick uses its charge, whether the tick casts a spell of its own or is the channel's own
periodic damage, Arcane Missiles' charge goes to the first missile launched after the gain, and Penance's heals leave it
(EV-0149). A Revelation charge gained while a channel whose ticks cast a spell of their own runs goes to its next tick;
a charge gained during a channel of periodic ticks is not counted, and the sim gives it to the next tick too. A charge
no tick takes, in Volley or Penance's heal, is still used at the channel's end. A buff whose row names no spells keeps
nothing for a channel: the first hit it hears uses the charge as it lands, so Weakness Analyzer's charge goes with the
first Arcane Missiles missile that crits, mid-channel.
