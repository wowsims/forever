# R-0048: A charge buff that names its spells is spent at the end of a channel that hits it

- Skill: wowsims-procs
- Code: `sim/core/spelldata/charges.go`, `AttachChargeSpender`
- Pull request: #658
- Evidence: EV-0149, EV-0162

A charge buff hit by the wearer's channel is reserved and spent at that channel's end (`Unit.OnChannelEnd`), an
interrupted channel included, and a charge gained while a channel its class flags name is running is spent at that
channel's end.

This holds for a buff whose row names the spells that spend it by class flags. Revelation keeps nothing for a channel's
cast: a channel whose ticks cast a spell of their own spends its charge on the first tick, and one whose ticks are its
own periodic damage leaves it to the next spell the buff affects (EV-0149). A Revelation charge gained while a channel
its class flags name is running is still spent at that channel's end, unless a tick takes it first. A buff whose row
names no spells keeps nothing for a channel: the first hit it hears spends the charge as it lands, so Weakness
Analyzer's charge goes with the first Arcane Missiles missile that crits, mid-channel.
