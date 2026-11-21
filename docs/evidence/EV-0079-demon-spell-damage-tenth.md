# EV-0079: A demon's spell damage is a tenth of the warlock's

- Claim: a Sayaad's Spell Damage is a tenth of the warlock's, plus Demonic Knowledge's own share when taken.
- Sources: the warlock's tooltip and a combat log of a Sayaad (no Unholy Power, no Demonic Knowledge); report code not
  recorded
- Game build: not recorded
- Date: 2026-10-14 for the Sayaad log
- Pull request: #121; the reading is in its commit c3d9ea9516

At 30 spell damage the warlock's tooltip reads "Increases your pet's Attack Power by 5 and Spell Damage by 3". The Sayaad
logged 3 Spell Damage at the warlock's 30 and 9 at about 96, and its Lash of Pain dealt 17.28 = 16 + 0.429 x 3. The
Sayaad's 42 at 126 (EV-0077) is the tenth plus Demonic Knowledge 3/3 (+30 at level 30): 12.6 + 30 = 42.6. The reading
was made on the Sayaad only.
