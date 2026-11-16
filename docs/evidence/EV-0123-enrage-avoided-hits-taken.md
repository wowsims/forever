# EV-0123: Enrage on avoided hits taken, across many warriors

- Claim: Enrage (12880) procs only on a hit taken that deals damage. Misses, dodges, parries, full resists, immunes,
  full blocks and ticks never proc it.
- Sources: public reports on forever.warcraftlogs.com, zone 3000, every fight of each: 3qbjvARyMLPDtcNg,
  4XH3ZGBf8FRnbpDj, 6DWXhYJRKVqrBjMx, 6amfKnRg7VMtDL1w, 79RVgZWM8dF2wkCp, 7LzJqTZRFj8ym36X, 7M3gGcH6bprxKqZY,
  7cpTkjaFB3mz6VgM, 7rGZbTJtpAqnBh2Y, 8G3K1TCaq4fLXhHz, 941apA2h6kRL8DcW, 9cQz2ymYMgfL7TwK, AYF7h4Kq9GTRJHmv,
  AvNwTdyBbDncFf2h, B4ak2F9hrztgQJH6, BDFj9VNrfXhxwQT3, BymPn39CrpcWbxv2, C8rwjpMbhYDga6FL, Cqkfvy2Lrdtc47zJ,
  D463FZBX1C9maJ7b, FPfGz7bjkvmyWXTB, FXaN14xP7VhbDpQM, GxTn6w2dD9MmXypP, HGNC6QbhXcTqvZ93, HT8dPLnkQbZtRXqF,
  JG8DxcCHLj6T9qXk, Jd2k6wT9zQCNbFj3, KYmwz2q1tG8yAcdP, NZrPmMxy3LXtj9G8, NfLb7C1qDZ8cHxFp, NmV6bvgJGTPHfQKF,
  P4Nv9AfZa8j1DdGr, Q1TkNaDV8BXf39dK, QVg7ANTY4B8RhDJM, QVxA76TD3LvH1ypa, RVgGFynN3KCYJHcA, RkFL4TH2rJnx7y8C,
  VXB4cGh1MrHD9y2k, WmFMHdtBrDhwRv2P, XJATCZLR1486QGY2, XWnJCQ1Mc38fRtwF, XdH8BcbwW6MQrYRg, Xv2QmnyK8RBxW9jZ,
  a2MybmTGLPr9AtqB, a3xGLZnNKTjQMFPY, aWZPH6nJ7xGw83Bf, bdYgfJk7Mh8GqZ4t, cT8wJVq316kAWpn9, cq2NnL1FrxgpCZjy,
  dTB9yPvkh4nXVAfK, daT4FcQgCYw2WBq6, dnXbDqMcL3pT1mr6, dxCHgDbw7z8PV3TK, fV1k4RHzhw7KDamQ, fn2Dc941K63CNvmj,
  fxvaWYAhcCZJ4pVT, hQztPF61ByCGgRwN, hf1pF42aPHMv9kLX, j6mDGcMhb7kFCpgn, k13frPZGVLQMtCzh, mFrhXgxtWDH7ZbGw,
  nyMqbY64tL7HhKQJ, pmMZRAD13CjqFVPW, pqmQRZAcyMa1L98X, qrwtb1GNA8W7L6Xm, qt9A4F718DcRGTQZ, tV6cx9nMN4782FX3,
  vGFaKgLT9mkn1Jdy, vKaqtxAX47zQDZVk, xcvKXzH7Rwk1Vb6J, xqnmW3BDpFb28cJ1, zL68yTpjZx2mAbgk, zmDY4LVX7N8vh9aR,
  zrALwGH82WM6Jcbx
- Game build: 1.60.1.70862 to 1.60.1.70938 by the report dates; the hotfix state and relogs were not recorded.
- Date: 2026-11-10 to 2026-11-15
- Pull request: #1049

The 74 reports were drawn at random from the zone's reports that log Enrage, and hold 100 warriors who gain it. For each
warrior, every hit taken is set against their Enrage applybuff and refreshbuff lines. A hit counts only when no other
hit on the same warrior falls within 50 ms of it, so a gain can belong to that hit alone; a gain counts for it when it
comes within 50 ms after. A window of 300 ms gives the same zeros and the same landed rates.

| Hit taken             | Hits   | Enrage gains | Rate  |
| --------------------- | ------ | ------------ | ----- |
| Miss                  | 8,363  | 0            | 0%    |
| Dodge                 | 5,554  | 0            | 0%    |
| Parry                 | 5,297  | 0            | 0%    |
| Full block, no damage | 1,915  | 0            | 0%    |
| Full resist           | 212    | 0            | 0%    |
| Immune                | 40     | 0            | 0%    |
| Tick                  | 5,335  | 0            | 0%    |
| Landed melee          | 50,663 | 13,331       | 26.3% |
| Landed spell          | 11,788 | 2,314        | 19.6% |

The landed rates are lower bounds: the landed rows keep hits fully absorbed, and a proc while Enrage is up does not
always log a refresh. The rates per hit come from EV-0050.
