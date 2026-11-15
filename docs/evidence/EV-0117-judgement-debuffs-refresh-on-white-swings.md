# EV-0117: Judgement debuffs refresh on white swings and Windfury Totem's strike

- Claim: a paladin's landed white swing and Windfury Totem's strike refresh the paladin's Judgement of the Crusader,
  Judgement of Light and Judgement of Wisdom on their target, and Holy Strike, Seal of Command's hit, Judgement of
  Command, Judgement of Righteousness, Judgement of Fury and Hammer of Wrath refresh none.
- Sources: combat log, Retribution paladin, levels 43 to 46, with a two-handed mace and, for part of the first log, a
  one-handed mace and a shield; two logs: the open world, with three rank 1 Windfury Totem strikes, and the open world
  followed by a five-player dungeon under a rank 2 Windfury Totem
- Game build: 1.60.1.70938, after the 2026-11-14 02:00 UTC hotfix, relogged
- Date: 2026-11-14
- Pull request: #1024

A refresh is a `SPELL_AURA_REFRESH` line of the paladin's own Judgement of the Crusader, Judgement of Light or Judgement
of Wisdom on the hit's target, within 120 ms of the hit. A target counts when it had carried the debuff for 0.2 seconds
and kept it for 0.2 seconds more. A hit is lone when no other swing, Holy Strike, seal hit, judgement, Hammer of Wrath
or Windfury Totem strike of the paladin's, landed or avoided, falls within 150 ms either side on any target; a white
swing's own Seal of Righteousness and Seal of Fury hits do not count against it. Consecration ticks and item procs were
not looked at. Judgement of Fury's own debuff is its taunt, which no hit refreshes, and is left out.

| Hit                                                      | First log                | Second log |
| -------------------------------------------------------- | ------------------------ | ---------- |
| Lone white swing                                         | 396 of 404               | 12 of 12   |
| Lone Holy Strike                                         | 0 of 173                 | 0 of 13    |
| Lone Judgement of Command                                | 0 of 69                  | 0 of 3     |
| Lone Judgement of Righteousness                          | 0 of 75                  | 0 of 7     |
| Lone Judgement of Fury                                   | 0 of 23                  | none cast  |
| Lone Hammer of Wrath                                     | 0 of 2                   | 0 of 6     |
| White swing with Seal of Command hits and nothing else   | 136 of 137 show one line | 12 of 12   |
| Windfury Totem's strike, a refresh line after the strike | not separable            | 13 of 13   |

Of the 8 white swings with no line inside the window, 2 show one 120 to 600 ms later, and 7 were with the one-hander, as
was every Judgement of Fury. A white swing that Seal of Command hits follow never shows a second line, the 14 swings
with two or three hits from Echo of Command among them.

In the second log a proc's line follows its swing by 25 to 45 ms, the strike follows its trigger by as much, and the
refresh a strike brings follows 5 to 80 ms after the strike. The 13 strikes came after a Holy Strike (4), after a Seal
of Command hit (5) and after a white swing whose own refresh was already logged (4). In the first log a proc's line
shares its swing's millisecond, so the three rank 1 strikes there cannot be told from their trigger. Two of them came
with a Holy Strike, and both show a refresh that no lone Holy Strike shows.

The paladin had Sacred Arbiter, whose tooltip has Holy Strike refresh all Judgement effects on the target. No Holy
Strike wrote a refresh line. Whether the talent extends the debuffs without one is not measured: they last 40 seconds,
and none ran out with a Holy Strike after its last logged refresh.
