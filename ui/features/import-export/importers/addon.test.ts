// The version request sent at module load, and what an import changes: the class is required, and
// anything missing from the pasted text stays as it was.
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Profession, Race } from '@generated/proto/common';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

// `getWSEVersion()` runs at module evaluation, so importing this module fires a request. Stubbed
// in `vi.hoisted`, which runs before the imports.
const fetchStub = vi.hoisted(() => {
	const stub = vi.fn((_url: string) => Promise.reject(new Error('offline')));
	(globalThis as unknown as { fetch: unknown }).fetch = stub;
	return stub;
});

import { ADDON_IMPORTER } from './addon';

describe('ADDON_IMPORTER', () => {
	// The failure is swallowed by design: no GitHub, no version warning, and the import still runs.
	it('asks GitHub for the addon version once, at module load', () => {
		expect(fetchStub).toHaveBeenCalledTimes(1);
		expect(fetchStub.mock.calls[0][0]).toBe('https://api.github.com/repos/wowsims/exporter/releases/latest');
	});

	it('rejects a payload that is not an addon export', async () => {
		await expect(ADDON_IMPORTER.onImport({} as never, 'not json')).rejects.toThrow('import.addon.error_invalid_export');
	});
});

const rotation = { type: 'TypeAPL', priorityList: [{ action: { autocastOtherCooldowns: {} } }] };

// A warrior sim with a race and professions already set.
const startingSim = async () => {
	const built = await makeIndividualSim();
	built.player.setRace(Race.RaceHuman);
	built.player.setProfessions([Profession.Engineering]);
	return built;
};

describe('ADDON_IMPORTER sections', () => {
	beforeAll(() => serveDatabase());

	it('applies a rotation on its own and leaves race, professions and gear as they were', async () => {
		const { player, host } = await startingSim();
		const gear = player.getGear();

		await ADDON_IMPORTER.onImport(host, JSON.stringify({ class: 'warrior', rotation }));

		expect(player.aplRotation.type).toBe(APLRotationType.TypeAPL);
		expect(player.aplRotation.priorityList).toHaveLength(1);
		expect(player.getRace()).toBe(Race.RaceHuman);
		expect(player.getProfessions()).toEqual([Profession.Engineering]);
		expect(player.getGear()).toBe(gear);
	});

	it('keeps the current rotation when the export has none', async () => {
		const { player, host } = await startingSim();
		player.setAplRotation(APLRotation.fromJson(rotation));

		await ADDON_IMPORTER.onImport(host, JSON.stringify({ class: 'warrior', race: 'Orc', gear: { items: [null] } }));

		expect(player.getRace()).toBe(Race.RaceOrc);
		expect(player.getProfessions()).toEqual([Profession.Engineering]);
		expect(player.aplRotation.priorityList).toHaveLength(1);
	});

	it('applies everything in a full export', async () => {
		const { player, host } = await startingSim();

		await ADDON_IMPORTER.onImport(
			host,
			JSON.stringify({
				class: 'warrior',
				race: 'Orc',
				professions: [{ name: 'Blacksmithing', level: 300 }],
				gear: { items: [] },
				rotation,
			}),
		);

		expect(player.getRace()).toBe(Race.RaceOrc);
		expect(player.getProfessions()).toEqual([Profession.Blacksmithing]);
		expect(player.aplRotation.priorityList).toHaveLength(1);
	});

	it.each([
		['names no class', { rotation }, 'import.errors.cannot_parse_class'],
		['is for another class', { class: 'mage', rotation }, 'import.errors.wrong_class'],
		['holds a rotation that does not parse', { class: 'warrior', rotation: 'not a rotation' }, 'import.addon.error_invalid_rotation'],
	])('changes nothing when the export %s', async (_name, json, message) => {
		const { player, host } = await startingSim();

		await expect(ADDON_IMPORTER.onImport(host, JSON.stringify(json))).rejects.toThrow(message);

		expect(player.aplRotation.priorityList).toHaveLength(0);
	});

	it.each([
		['has no section to import', JSON.stringify({ class: 'warrior', talents: '05' }), 'import.addon.error_nothing_to_import'],
		['is not an object', '[1]', 'import.addon.error_invalid_export'],
	])('rejects an export that %s', async (_name, data, message) => {
		const { host } = await startingSim();
		await expect(ADDON_IMPORTER.onImport(host, data)).rejects.toThrow(message);
	});
});
