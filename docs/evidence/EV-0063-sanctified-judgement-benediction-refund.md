# EV-0063: Sanctified Judgement refunds the base cost with Benediction's cut added on

- Claim: Sanctified Judgement 3/3 refunds 60% of the judged seal's base mana cost, and Benediction's cut on the seal is
  added to the refund instead of taken from it: a 90 mana seal costs 81 and returns 59.4.
- Sources: a log reported on Discord, then measured in game on one character at every Benediction rank with Sanctified
  Judgement 3/3
- Game build: not recorded
- Date: not recorded
- Pull request: #507

Before the fix the sim refunded 48.6 mana on a 90 mana Seal of Righteousness with Benediction 5/5.

| Benediction | Seal mana paid | Refund | 60% x 90 x factor |
| ----------- | -------------- | ------ | ----------------- |
| 0/5         | 90             | 54     | x 1.00            |
| 1/5         | 88             | 55.08  | x 1.02            |
| 3/5         | 85             | 57.24  | x 1.06            |
| 5/5         | 81             | 59.40  | x 1.10            |

The same 60% of base x 1.1 showed on a 110 mana Seal of Light, a 65 mana Seal of the Crusader and the 20 mana rank 1
Seal of Righteousness in the earlier logs, and ranks 1 and 2 of the talent refund 20% and 40% of base without
Benediction. The chance is 100% at 3/3 and the refund fires on the cast, missed judgements included. With Twist of Light
as a second cut the sim multiplies the two before the flip (x 1.28 on 72% paid); the game has only been measured with one
cut, so that part is the model.
