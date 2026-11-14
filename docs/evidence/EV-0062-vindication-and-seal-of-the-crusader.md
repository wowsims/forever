# EV-0062: Vindication 5 a minute on weapon speed, Seal of the Crusader 1.4 times faster

- Claim: Vindication procs 5 times a minute on the weapon's speed, and under Seal of the Crusader each swing's chance and
  damage fall by the seal's 1.4.
- Sources: Forever combat logs with Vindication 1/3, cancelling the buff after every proc so each one shows, at 1.6 and
  3.6 weapon speed; report codes not recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #409

The talent's row (9452) carries two dummy effects and a ProcChance column of 100.

| Weapon    | Seal          | Procs / landed swings | Chance |
| --------- | ------------- | --------------------- | ------ |
| 1.6 speed | Righteousness | 104 / 819             | 12.7%  |
| 1.6 speed | Crusader      | 67 / 648              | 10.3%  |
| 3.6 speed | Righteousness | 66 / 205              | 32.2%  |
| 3.6 speed | Crusader      | 80 / 394              | 20.3%  |

- 5 procs a minute on the weapon's speed fits all four (fit 5.0, 95% interval 4.5 to 5.5).
- Under Seal of the Crusader each swing's chance is multiplied by 0.715 (0.585 to 0.870), the seal's 1/1.4. A chance the
  seal leaves alone fits at p about 0.001.
- Seal of Righteousness's damage never procs it: in every proc with the seal's damage nearby, the buff is logged before
  that damage.
- Seal of the Crusader's white hits deal 1/1.4 of their damage, attack power's share included. The 3.6 speed axe's spread
  of hits narrows from 22 to 16 (23/1.4 is 16.4), and its average hit before armor is 109 under the seal against the 111
  that 1/1.4 predicts.
- The target's debuff (440667) is Magic in SpellCategories and the log shows `SPELL_MISSED ... MISS` for it: 10 misses in
  305 rolls, none dodged or parried. The paladin's buff comes with every proc.

Not measured: the talent's proc mask also hears melee specials (Holy Strike, Seal of Command's hits, the Judgements),
which roll the same chance, and haste other than the seal is taken to leave the chance alone.
