# EV-0006: Weapon procs skip shield and magic DefenseType abilities

- Claim: no chance-on-hit item effect or combat enchant hears an ability that requires a shield or has magic
  DefenseType; equip procs still hear them.
- Sources: combat logs of Protection warriors; report codes not recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #958

| Proc                          | Shield Slam         | Shield Bash | Thunder Clap | White main-hand hits |
| ----------------------------- | ------------------- | ----------- | ------------ | -------------------- |
| Fiery Blaze (enchant 36)      | 0 of 617            | 0 of 78     | 0 of 436     | 14.0%                |
| Rock on Stick                 | 0 of 157            |             |              |                      |
| Dragon's Call                 | 1 of 50             | 0 of 4      |              |                      |
| Dreadfrost Saber's equip proc | 54 of 80, 60 of 105 | 7 of 8      |              |                      |

Dragon's Call's one Shield Slam proc shares its millisecond with a white hit. Non-damaging Sunder Armor, Concussion
Blow and Rend applications proc Fiery Blaze and Dragon's Call; Rend ticks proc neither.
