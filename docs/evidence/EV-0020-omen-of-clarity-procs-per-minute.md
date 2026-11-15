# EV-0020: Omen of Clarity procs 2.0 times a minute

- Claim: Omen of Clarity procs at 2.0 per minute; the game files state none.
- Sources: combat logs of feral druids quoted in issue #225; report codes and counts per row not recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #248, which set the override to the measured rate; the logs are issue #225's

The pull request quotes these readings from those logs. Specials landed at about 8.6% with weapons of about 3.1 s, which a
fixed 2.5 s and the weapon's speed at about 1.6 procs per minute fit equally well, while 2.0 per minute on the weapon's
speed predicts about 10.4%. Every logged feral swung a slow weapon, so a dagger or one-hander cat log would settle it.
Applications of damage-free hits rolled it too: Faerie Fire 23 of 365 and Bash 2 of 22; Growl 0 of 150. The sim rolls
auto attacks at their own swing speed, specials at the equipped main-hand weapon's speed in every form, and spells at
their cast time with an instant at 1.5 s.
