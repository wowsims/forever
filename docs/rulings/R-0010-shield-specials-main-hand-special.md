# R-0010: Shield Bash and Shield Slam register as main-hand melee specials

- Skill: wowsims-abilities-auras
- Code: `sim/warrior/shield_bash.go`, `sim/warrior/talents_protection.go`, `sim/warrior/thunder_clap.go`, `sim/core/buffs/drivers.go`
- Pull request: none
- Evidence: EV-0096

Shield Bash and Shield Slam register as `ProcMaskMeleeMHSpecial`, though their rows require the shield: the main-hand
weapon's equip procs hear them, and the main-hand weapon skill sets their miss and parry, as measured in game (EV-0096).
The game files give no hand. The shield stays their cast requirement, and Windfury Totem rolls on them as
on any melee ability. The shield requirement also keeps every chance-on-hit item effect and combat enchant from hearing
them, as Thunder Clap's magic DefenseType does for it (R-0041).
