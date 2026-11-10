import { EnchantType, EquipmentSpec, ItemSpec, ItemType } from '@generated/proto/common';
import { SimDatabase, SimEnchant } from '@generated/proto/db';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { Database } from './database';

// `Database.get` memoises its load in module scope, so every case resets the module registry to get
// a fresh memo rather than sharing one across tests.
// Fake timers are installed *after* the dynamic import, not in `beforeEach`: the module loader's own
// microtasks run on a frozen clock otherwise, and the retry case fails intermittently.
const freshDatabase = async () => {
	vi.resetModules();
	const { Database } = await import('./database');
	vi.useFakeTimers();
	return Database;
};

const okResponse = () => ({ json: () => Promise.resolve({}) });

const FIRST_SEEN_URL = '/forever/assets/database/item_first_seen.json';
// `Database.get` fetches the first-seen stamps beside db.json; this answers them with no entries so
// the inner mock sees and counts the db.json requests alone.
const withFirstSeen = (respond: (url: string) => Promise<unknown>) =>
	vi.fn((url: string) => (url === FIRST_SEEN_URL ? Promise.resolve(Response.json([])) : respond(url)));

// `runAllTimersAsync` drains what is scheduled when it is called; the retry schedules its next
// backoff from inside a rejection handler, so under load the third attempt can be scheduled after
// the drain has already finished. Advancing until the call count settles is deterministic.
const drainRetries = async (fetchMock: { mock: { calls: unknown[] } }, until: number) => {
	for (let i = 0; i < 20 && fetchMock.mock.calls.length < until; i++) await vi.advanceTimersByTimeAsync(1000);
};

describe('Database.get', () => {
	afterEach(() => {
		vi.useRealTimers();
		vi.unstubAllGlobals();
	});

	it('retries a failing load before giving up', async () => {
		const fetchMock = vi.fn().mockRejectedValueOnce(new Error('boom')).mockRejectedValueOnce(new Error('boom')).mockResolvedValue(okResponse());
		vi.stubGlobal('fetch', withFirstSeen(fetchMock));

		const Database = await freshDatabase();
		const loaded = expect(Database.get()).resolves.toBeDefined();
		await drainRetries(fetchMock, 3);
		await loaded;

		expect(fetchMock).toHaveBeenCalledTimes(3);
	}, 30000);

	// A memo that held the rejected promise for the page lifetime would make every later caller
	// await the same failure, and the app could not recover.
	it('does not memoize a failure — a later call tries again and can succeed', async () => {
		const fetchMock = vi.fn().mockRejectedValue(new Error('offline'));
		vi.stubGlobal('fetch', withFirstSeen(fetchMock));

		const Database = await freshDatabase();
		// The assertion is attached before the timers run: `get()` rejects while they are draining,
		// and a rejection with no handler yet attached surfaces as an unhandled rejection.
		const first = expect(Database.get()).rejects.toThrow('offline');
		await drainRetries(fetchMock, 3);
		await first;

		const attemptsWhileFailing = fetchMock.mock.calls.length;
		fetchMock.mockResolvedValue(okResponse());

		const attemptsBeforeSecond = fetchMock.mock.calls.length;
		const second = expect(Database.get()).resolves.toBeDefined();
		await drainRetries(fetchMock, attemptsBeforeSecond + 1);
		await second;
		expect(fetchMock.mock.calls.length).toBeGreaterThan(attemptsWhileFailing);
	}, 30000);

	it('keeps the consumables of db.json when it merges the leftovers', async () => {
		const bodies: Record<string, object> = {
			'/forever/assets/database/db.json': { consumables: [{ id: 5, name: 'Potion' }], itemIcons: [{ id: 5, name: 'Potion', icon: 'inv_potion' }] },
			'/forever/assets/database/leftover_db.json': { items: [{ id: 7, name: 'Grey Sword', scalingOptions: { 0: { ilvl: 5 } } }] },
		};
		const fetchMock = vi.fn((url: string) => Promise.resolve({ json: () => Promise.resolve(bodies[url]) }));
		vi.stubGlobal('fetch', withFirstSeen(fetchMock));

		const Database = await freshDatabase();
		const db = await Database.loadLeftoversIfNecessary(EquipmentSpec.create({ items: [ItemSpec.create({ id: 7 })] }));

		expect(fetchMock).toHaveBeenCalledWith('/forever/assets/database/leftover_db.json');
		expect(db.getItemById(7)?.name).toBe('Grey Sword');
		expect(db.getConsumable(5)?.name).toBe('Potion');
	}, 30000);

	it('shares one load between concurrent callers', async () => {
		const fetchMock = vi.fn().mockResolvedValue(okResponse());
		vi.stubGlobal('fetch', withFirstSeen(fetchMock));

		const Database = await freshDatabase();
		const [a, b] = await Promise.all([Database.get(), Database.get()]);

		expect(a).toBe(b);
		expect(fetchMock).toHaveBeenCalledTimes(1);
	}, 30000);

	it('reads the first-seen stamps beside db.json', async () => {
		const bodies: Record<string, unknown> = { '/forever/assets/database/db.json': {}, [FIRST_SEEN_URL]: [{ id: 7, firstSeen: 1700000000 }] };
		vi.stubGlobal(
			'fetch',
			vi.fn((url: string) => Promise.resolve(Response.json(bodies[url]))),
		);

		const Database = await freshDatabase();
		const db = await Database.get();

		expect(db.getItemFirstSeen(7)).toBe(1700000000);
		expect(db.getItemFirstSeen(8)).toBeUndefined();
	}, 30000);

	it('loads with no stamps and one warning when the first-seen file is missing', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((url: string) => Promise.resolve(url === FIRST_SEEN_URL ? new Response('Not found', { status: 404 }) : Response.json({}))),
		);
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});

		const Database = await freshDatabase();
		const db = await Database.get();

		expect(db.getItemFirstSeen(7)).toBeUndefined();
		expect(warn).toHaveBeenCalledOnce();
		warn.mockRestore();
	}, 30000);

	it('returns every row of an effect id once, across the slots it fits', async () => {
		const enchants = [
			{
				effectId: 1897,
				itemId: 2,
				name: 'Enchant 2H Weapon - Lesser Impact',
				type: ItemType.ItemTypeWeapon,
				enchantType: EnchantType.EnchantTypeTwoHand,
			},
			{ effectId: 1897, itemId: 3, name: 'Enchant Weapon - Superior Striking', type: ItemType.ItemTypeWeapon },
			{ effectId: 849, itemId: 1, name: 'Enchant Cloak - Minor Agility', type: ItemType.ItemTypeBack },
		];
		vi.stubGlobal('fetch', withFirstSeen(vi.fn().mockResolvedValue({ json: () => Promise.resolve({ enchants }) })));

		const Database = await freshDatabase();
		const db = await Database.get();

		expect(db.getEnchantsByEffectId(1897).map(enchant => enchant.name)).toEqual([enchants[0].name, enchants[1].name]);
	}, 30000);
});

describe('Database.mergeSimDatabases', () => {
	it('keeps the enchant rows of one effect id that differ in type', () => {
		const twoHand = SimEnchant.create({ effectId: 1897, type: ItemType.ItemTypeWeapon, enchantType: EnchantType.EnchantTypeTwoHand });
		const oneHand = SimEnchant.create({ effectId: 1897, type: ItemType.ItemTypeWeapon });

		const merged = Database.mergeSimDatabases(SimDatabase.create({ enchants: [twoHand] }), SimDatabase.create({ enchants: [oneHand, twoHand] }));

		expect(merged.enchants).toEqual([twoHand, oneHand]);
	});
});
