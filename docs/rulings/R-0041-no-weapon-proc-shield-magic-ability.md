# R-0041: No weapon proc hears a shield or magic DefenseType ability

- Skill: wowsims-procs
- Code: `sim/core/aura_helpers.go`, `canProcFrom`
- Pull request: #958
- Evidence: EV-0021

No weapon proc hears an ability whose row requires a shield (SpellEquippedItems), which strikes with the shield, or has
magic DefenseType: a weapon's chance-on-hit item effect (ItemEffect TriggerType 2) and a combat enchant both skip its
hits. Auto attacks are exempt, since a wand's shots carry magic DefenseType and cast weapon procs. Equip procs
(TriggerType 1) still hear these abilities. Shield Slam and Shield Bash require the shield, and Thunder Clap is Physical
with magic DefenseType. Non-damaging Sunder Armor, Concussion Blow and Rend applications proc Fiery Blaze and Dragon's
Call; Rend ticks proc neither.
