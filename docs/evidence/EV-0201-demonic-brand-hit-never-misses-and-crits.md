# EV-0201: Demonic Brand's hit never misses and is never fully resisted; it crits and takes partial resists

- Claim: The brand's damage (Shadow 1293697, Fire 1293698), one line for each charge a demon's hit consumes, rolls no
  hit: none of about 19,900 lines misses or is fully resisted, while the same demons' Firebolt and Lash of Pain do. It
  crits and takes partial resists. The Fire row comes from the Imp and the Shadow row from the other demons, and every
  melee, Firebolt and Lash of Pain hit on a branded target deals it.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the 80 warlock streams on disk in the reports whose
  abilities list either row; the brand deals damage in the 62 reports and fights §Reports lists
- Game build: 1.60.1.70938 to 1.60.1.71037 by event time (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); the hotfix each player ran is not recorded
- Date: 2026-11-13 10:03 to 2026-11-22 23:07 UTC
- Pull request: #1233

## Method

Every damage line of the two rows in each warlock's stream; the demon is the source. A hitType 16 or 17 line carries a
`resisted` amount, so 16 is a partly resisted hit and 17 a partly resisted crit. A crit is hitType 2 or 17. A miss
would read hitType 0 and a full resist 14. The control is the same demons' own Firebolt, Lash of Pain and melee lines in
the same streams.

## Counts

| Row                    | Build | Hit (1) | Crit (2) | Partial (16) | Partial crit (17) | Immune (10) | 12  | Miss (0) | Resist (14) | Crit share |
| ---------------------- | ----- | ------- | -------- | ------------ | ----------------- | ----------- | --- | -------- | ----------- | ---------- |
| Shadow 1293697         | 70938 | 11,160  | 958      | 93           | 8                 | 19          | 12  | 0        | 0           | 7.9%       |
| Shadow 1293697         | 70959 | 1,199   | 100      | 30           | 7                 | 6           | 0   | 0        | 0           | 8.0%       |
| Shadow 1293697         | 71037 | 66      | 15       | 0            | 0                 | 0           | 0   | 0        | 0           | 18.5%      |
| Fire 1293698           | 70938 | 1,197   | 95       | 53           | 6                 | 0           | 0   | 0        | 0           | 7.5%       |
| Fire 1293698           | 70959 | 4,061   | 457      | 312          | 38                | 0           | 0   | 0        | 0           | 10.2%      |
| Shadow 1293697, in all | all   | 12,425  | 1,073    | 123          | 15                | 25          | 12  | 0        | 0           | 8.0%       |
| Fire 1293698, in all   | all   | 5,258   | 552      | 365          | 44                | 0           | 0   | 0        | 0           | 9.6%       |

The crit share is of the lines that dealt damage (hitType 1, 2, 16 and 17). The 12 hitType 12 lines deal 0, on three
mobs in two reports (7aVBrLNq89nfgvdM and 6hVmGCdjHTYcMJ71); Firebolt takes the same hitType on 7 lines.

The same demons' own attacks in the same streams do miss and get fully resisted: Firebolt 245 misses and 353 full
resists in 11,852 lines, Lash of Pain 56 misses in 5,267, melee 365 misses, 980 dodges and 173 parries in 30,050.

## School and charges

Each of the 28 demons that deals the Fire row also casts Firebolt, and none of the 45 that deal the Shadow row does.

A demon's landed hit on a target inside a window of the brand's debuff (1293696, applydebuff to removedebuff) is
answered when a brand line on the same target falls within 50 ms: melee 11,503 of 11,577, Firebolt 6,018 of 6,096 and
Lash of Pain 1,869 of 1,883. Retribution Aura and Thorns lines logged with a demon as their source are left out.

## Reports

Report codes and the fights in which the brand deals damage.

Shadow: 12RB4fbmrp3YZjPF 1, 2, 3, 5; 28RBXDnLZCWNpHf9 2; 6hVmGCdjHTYcMJ71 1, 4, 5; 78wWZfmgrJYzFXvD 34;
7aVBrLNq89nfgvdM 1, 3, 4, 6, 8, 9, 10; 7mnVXCgpyxkbtZGq 1, 2, 3, 4, 5, 6, 10, 11; 8jzXTaqr732V9DxJ 1, 3;
AYwLfvqh9QyNj6zJ 1; BX8n2xTycaKr4CGL 1; GTJw7YXampLrWcgM 1, 2, 3, 4, 7, 8, 9, 10, 11, 12, 13; HYmf1vTW6hVDPytK 13;
JT3mkMDXFYy4QbNq 1; JyXKg38NTr9GvmYA 1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 15, 16, 18; KPkDVWCvjLw2BqNR 1, 3;
LYw76XG1CFHQNdJz 4; Lr12yNYCdD3a7HQW 1, 2, 3, 5, 6, 7; M7cz2xknrqXVAKjD 1;
MXdJx7rWLQn2Av9P 1, 3, 4, 5, 6, 11, 13, 15, 16, 17, 18, 19; NXMZV9xjvWpB7GtP 122; PGXN8fK9FphR1n4v 1, 2;
PGgDMwaVk9ptQLnZ 1, 7, 8, 10; T3hVZkfLcnp42MzP 1; TbB8q4RMK3LmYJNv 1, 2, 3, 4, 5, 6, 7; VmX6tMZzjQdfJHkT 1, 2, 3;
WM4F6BHqNpYLh2vr 2; WhPpmnzB6rR3gD4Z 1, 2, 3, 5, 7, 9, 10, 11, 12, 15, 16, 17, 18, 19, 21, 22, 23, 25;
YvNphABXf269cGFJ 1, 2, 3, 4; kr3AzhZ9FVB1nG2D 1, 3; mdyXBDfw2LZvtn3c 1, 3, 4, 5, 6, 7, 8, 9, 10, 11; mzZNwpb3TjMFXGgH 1,
3, 4, 5, 6, 7, 8; nR1WGzY8pJKDxhQj 4, 10, 11, 16; ncxwTy7JbYHRZPDG 1; nhYyv7NG3L982RaB 1, 3, 6, 7, 9; pJvWftwmdkbq6T9h
1, 3; praQ172yvmdZNbgn 2; qZwYtNQLPG1xzkJW 1, 2, 3, 5, 6, 7; rCnfgxYFVXLHvjZp 1; wrzB6GtjPaqCKJZW 1;
y3wkJZnA7gv4qzfX 1, 2, 6, 7, 8, 9, 10, 11, 13, 14, 15, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30;
zLacx86Thr3AMtqn 1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15.

Fire: 1MA3qWLGa8v2hdXP 43; 1kWMb9cQ67dJhrZz 2; 28RBXDnLZCWNpHf9 1, 2, 7, 8, 9; 2nZF4L6jRhwzGQa9 9, 12;
3Ra7Dtjhc9AXLCzJ 1, 5, 7, 9; 82WLwYNzvjB3bXM6 1; 8zB1PhRQy64jHkXY 1; 9b71qzjAg6Myt34K 1; DZCc2gtwhW3FBKyx 1, 2;
KVLgwAr9bB1dJyNR 1; LQ9zMDqpmknBgAX8 1, 2; LYw76XG1CFHQNdJz 4, 8, 9, 10; NXMZV9xjvWpB7GtP 122; Q3kAgpdcGJ27VPvT 3;
YpQAGtmMa3b47XRk 1; cbTt1zxajFfNVPDA 1, 3; hZLykw9dRnYbtvKP 1; k3ZGwNbF9W7DgzLQ 1, 3; mjrQyAZVTtxJvpN2 1;
ncxwTy7JbYHRZPDG 1; rYhcR3zbKmCk4N6V 1, 5, 6, 7, 8, 9, 11, 12, 13, 15, 16, 18, 19, 21, 22; vGgpbB8kmCjrRKf9 1;
vYbjafXydK2z8chx 1, 3, 4, 6, 7; vfBK4pHa9TtyG67w 1, 3; wrzB6GtjPaqCKJZW 1, 2, 3, 6, 7, 8, 9, 10;
xvJWqXQh1nakDH36 1, 2, 3, 4, 5, 6, 9, 10, 12, 13; y3wkJZnA7gv4qzfX 1; ywLTngctk7WB24v1 1, 3.
