import { beforeEach, describe, expect, it, vi } from 'vitest';

import { deleteRuns } from './stat_sweep_history';

// happy-dom has no IndexedDB, so `idb` hands back a stub database that records the calls made on it.
const idb = vi.hoisted(() => {
	const remove = vi.fn(async (_id: string) => {});
	const transaction = vi.fn((_store: string, _mode: string) => ({ store: { delete: remove }, done: Promise.resolve() }));
	const close = vi.fn();
	return { remove, transaction, close, openDB: vi.fn(async () => ({ transaction, close })) };
});
vi.mock('idb', () => ({ openDB: idb.openDB }));

beforeEach(() => {
	vi.clearAllMocks();
});

describe('deleteRuns', () => {
	it('deletes every run in one readwrite transaction on one connection, and closes it', async () => {
		await deleteRuns(['first', 'second', 'third']);

		expect(idb.openDB).toHaveBeenCalledTimes(1);
		expect(idb.transaction).toHaveBeenCalledExactlyOnceWith('runs', 'readwrite');
		expect(idb.remove.mock.calls.map(call => call[0])).toEqual(['first', 'second', 'third']);
		expect(idb.close).toHaveBeenCalledTimes(1);
	});

	it('closes the connection when a delete fails', async () => {
		idb.remove.mockRejectedValueOnce(new Error('failed'));

		await expect(deleteRuns(['first'])).rejects.toThrow('failed');
		expect(idb.close).toHaveBeenCalledTimes(1);
	});
});
