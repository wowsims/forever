-- Borpa

-- Wisepaw Totem swings at 2.0, where the game files give it 3.2. Its damage comes from the same damage
-- table cell, so the range follows the speed.
UPDATE ItemSparse SET ItemDelay = 2000 WHERE ID = 274420;
