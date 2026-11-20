# EV-0167: Bloodmail Regalia's bleed procs on white swings and Windfury strikes

- Claim: The Bloodmail Regalia 3-piece bonus (1299740) puts its bleed (1299739) on the target at about 10.7% per landed
  white swing or Windfury strike, about 1.9 to 2.0 PPM on the wearers' 3.0 to 3.9 s two-handers; Stormstrike never
  procs it. The weapon speeds are too close to tell a chance per hit from a PPM, and the sim takes the chance. The
  row's base of 7 matches the bleed's damage per second; nothing here shows it is the chance. All the wearers counted
  are shamans, so the rate is observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every fight with 3 or more pieces worn in:
  79QPzLcgZCKbJBkH, 7RzFd1VfWgjZ3wML, 82F6RVYwhapt4zLP, 8GFRaPBnKWf7Qrpb, 9tjaL4MdpWx6rmhT, JZDFdhkvn3NP6xLm,
  KTcgdHnN2ajkFfXP, PAbVZ6pGamwv7yJd, R6aLAPN83fB9mMkv, XTxHyBK9qJv3mrRV, cXf64Z1GR72MYnbF, dK9TAcBfjVmD2n1J,
  hv7fxbPQ1qXKHWZV, knyCYtGqvZDWzxg1, ymzMnpTZ2PqxALvH. 9fRTQzPJXtrKgckV and zPH4NRMVFqkaQyC8 log the run in
  JZDFdhkvn3NP6xLm again, and pKCQwXNvnRBA84DF the run in KTcgdHnN2ajkFfXP; each run counts once. A hunter wearer
  (9YwTfqWK8vDRFgLr, 21 melee hits, no proc) and a paladin wearer (cf12vkzydFqKnQPT) are left out of the rate.
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-16 to 2026-11-20
- Pull request: #1146

## Method

11 shamans with 3 to 5 pieces, gear checked per fight from CombatantInfo. Only events the wearer is the source of count
(their totems and pets are left out). A run logged in several reports counts once: two runs of one player whose damage
events share over 30% of their seconds are one run. A proc is an applydebuff or refreshdebuff of 1299739 by the wearer.
The debuff logs 17 to 50 ms after the hit (most at 25 to 38 ms), so a hit gets credit for a proc 20 ms before to 120 ms
after it. The bleed lands on the target, so killing blows are left out everywhere. Mobs are keyed on target id and
instance.

- What rolls (row mask 4, melee swings): groups of the wearer's melee hits with nothing else within 120 ms show it.
  The last column predicts each group from the lone white chance, with every white swing and Windfury strike rolling once.

| Group of hits                          | Groups | With a proc | Share  | Predicted at 10.9% a roll |
| -------------------------------------- | -----: | ----------: | ------ | ------------------------- |
| White swing alone                      |  3,676 |         401 | 10.91% | 10.9%                     |
| White + 2 Windfury Weapon strikes      |    713 |         209 | 29.31% | 29.3%                     |
| White + 1 Windfury Totem strike        |    551 |         113 | 20.51% | 20.6%                     |
| White + Windfury Totem + 2 Weapon      |    258 |          94 | 36.43% | 37.0%                     |
| Stormstrike (with its debuff) alone    |  1,186 |           0 | 0%     | 0%                        |
| Stormstrike + 2 Windfury strikes       |    210 |          40 | 19.05% | 20.6%                     |

- All hits: landed white swings, Windfury Weapon strikes and Windfury Totem strikes.
- Lone white swings: landed white swings with no other melee event of the wearer within 200 ms. With every event of
  the wearer as a neighbour (300 ms) the chance is the same, 10.88% (131 of 1,204), and lone spells (0 of 3,406), ticks
  (0 of 2,236) and debuff landings (0 of 586) never proc it.
- Swap check: between two procs, the rolls due at 10.7% each were summed. 7 of 1,077 such stretches had 5 or more due
  and no proc, where chance alone gives about 7 (e^-5 of them), so no stretch is dropped as a swap.

## Rate

| Count             |   Procs / hits | Chance [95%, bootstrap over wearers] | PPM [95%]         | Wearers |
| ----------------- | -------------: | ------------------------------------ | ----------------- | ------: |
| All hits          | 1,118 / 10,455 | 10.69% [10.03, 11.56]                | 1.91 [1.79, 1.99] |      11 |
| Lone white swings |    402 / 3,646 | 11.03% [10.03, 12.52]                | 1.97 [1.81, 2.16] |      11 |

Per wearer the lone white chance runs from 7.6% to 15.4%, with no trend over the 3.0 to 3.9 s range. A likelihood fit
of a flat chance against a PPM on lone white swings favours neither (log-likelihood difference 0.7, bootstrap 95% from
-0.7 to 2.7). As a check only, the paladin wearer reads 27 procs in 314 hits (8.6%).
