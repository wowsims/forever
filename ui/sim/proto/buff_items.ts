import { Debuffs, PartyBuffs } from '@generated/proto/buffs';
import { SimDatabase } from '@generated/proto/db';

import type { Database } from './database';
import { Gear } from './gear';

// The items whose data the sim reads for a party buff. Without the full item database the sim
// knows only the items the request carries, so the request carries these while the buff is on.
export const PARTY_BUFF_ITEM_IDS: Partial<Record<keyof PartyBuffs, readonly number[]>> = {
	darkspear: [12802],
};

// The weapons whose swing speed and procs per minute the sim reads for a raid debuff, carried the same way. Puncture
// Armor carries all three weapons its states name.
export const DEBUFF_ITEM_IDS: Partial<Record<keyof Debuffs, readonly number[]>> = {
	armorShatter: [12798],
	punctureArmor: [13204, 13286, 9485],
};

const addItems = <Message extends object>(
	table: Partial<Record<keyof Message, readonly number[]>>,
	database: SimDatabase,
	db: Database,
	message: Message | undefined,
) => {
	const itemIds = Object.entries(table).flatMap(([field, ids]) => (message?.[field as keyof Message] ? (ids as readonly number[]) : []));
	for (const id of itemIds) {
		const item = db.getItemById(id);
		if (item && !database.items.some(carried => carried.id === id)) database.items.push(Gear.itemToDB(item));
	}
};

export const addPartyBuffItems = (database: SimDatabase, db: Database, partyBuffs: PartyBuffs | undefined) =>
	addItems(PARTY_BUFF_ITEM_IDS, database, db, partyBuffs);

export const addDebuffItems = (database: SimDatabase, db: Database, debuffs: Debuffs | undefined) => addItems(DEBUFF_ITEM_IDS, database, db, debuffs);
