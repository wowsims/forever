# R-0048: A charge buff hit by a channel is spent at that channel's end

- Skill: wowsims-procs
- Code: `sim/core/spelldata/charges.go`, `AttachChargeSpender`
- Pull request: #658
- Evidence: EV-0149

A charge buff hit by the wearer's channel is reserved and spent at that channel's end (`Unit.OnChannelEnd`), an
interrupted channel included, and a charge gained while a channel its class flags name is running is spent at that
channel's end.
