# R-0048: A charge buff hit by a channel is spent at that channel's end

- Skill: wowsims-procs
- Code: `sim/core/spelldata/charges.go`, `AttachChargeSpender`
- Pull request: #658
- Evidence: EV-0149, EV-0160

A charge buff hit by the wearer's channel is reserved and spent at that channel's end (`Unit.OnChannelEnd`), an
interrupted channel included, and a charge gained while a channel its class flags name is running is spent at that
channel's end.

This holds for a buff whose row names the spells that spend it by class flags, such as Revelation. A buff whose row
names no spells keeps nothing for a channel: the first hit it hears spends the charge as it lands, so Weakness
Analyzer's charge goes with the first Arcane Missiles missile that crits, mid-channel.
