import { DatabaseFilters, IndividualSimSettings, SimSettings, StatFilters } from '@generated/proto/ui';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { envelope, senderStatFilters, viewerFiltersWith, viewerStatFilters } from '@testing/sim/stat_filters';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { JSON_IMPORTER } from './json';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const document = (settings?: SimSettings) => IndividualSimSettings.toJsonString(envelope(settings));

// The viewer's filters before the import, their own stat filters among them.
const viewer = async (statFilters: StatFilters | undefined) => {
	const built = await makeIndividualSim();
	built.sim.setFilters({ ...built.sim.getFilters(), ...viewerFiltersWith(statFilters) });
	return built;
};

// What the same import leaves on a viewer the stat filters never touch, less the stat filters.
const withoutStatFiltersAfter = async (data: string) => {
	const { sim, host } = await viewer(undefined);
	host.fromProto(IndividualSimSettings.fromJsonString(data));
	return { ...sim.getFilters(), statFilters: undefined };
};

// A JSON export carries only the categories ticked in the exporter, and nothing in the file lists
// them. The importer has to apply exactly those, or every category left out is reset to empty.
const importedCategories = async (settings: IndividualSimSettings) => {
	const { host } = await makeIndividualSim();
	const fromProto = vi.spyOn(host, 'fromProto').mockImplementation(() => {});
	await JSON_IMPORTER.onImport(host, IndividualSimSettings.toJsonString(settings));
	return fromProto.mock.calls[0][1];
};

beforeAll(() => serveDatabase());

describe('JSON_IMPORTER', () => {
	it("keeps the viewer's stat filters over the ones the document carries", async () => {
		const { sim, host } = await viewer(viewerStatFilters);
		const data = document(SimSettings.create({ filters: DatabaseFilters.create({ minIlvl: 50, statFilters: senderStatFilters }) }));

		await JSON_IMPORTER.onImport(host, data);

		expect(StatFilters.equals(sim.getFilters().statFilters!, viewerStatFilters)).toBe(true);
		expect(sim.getFilters().minIlvl).toBe(50);
	});

	it.each([
		['filters without stat filters', document(SimSettings.create({ filters: DatabaseFilters.create({ minIlvl: 50 }) }))],
		['settings without filters', document(SimSettings.create({ iterations: 1000 }))],
		['no settings at all', document()],
	])("keeps the viewer's stat filters through a document with %s, and every other filter as before", async (_name, data) => {
		const { sim, host } = await viewer(viewerStatFilters);

		await JSON_IMPORTER.onImport(host, data);

		expect(StatFilters.equals(sim.getFilters().statFilters!, viewerStatFilters)).toBe(true);
		expect(DatabaseFilters.equals({ ...sim.getFilters(), statFilters: undefined }, await withoutStatFiltersAfter(data))).toBe(true);
	});

	it('leaves a viewer with no stat filters without any', async () => {
		const { sim, host } = await viewer(undefined);

		await JSON_IMPORTER.onImport(host, document(SimSettings.create({ filters: DatabaseFilters.create({ statFilters: senderStatFilters }) })));

		expect(sim.getFilters().statFilters).toBeUndefined();
	});

	it("carries the encounter's name from one sim's settings into another sim, and never as one of My encounters", async () => {
		const sender = await makeIndividualSim();
		sender.sim.encounter.setName('Phase two');
		sender.sim.encounter.setMyEncounter(true);
		const { sim, host } = await makeIndividualSim();

		await JSON_IMPORTER.onImport(host, IndividualSimSettings.toJsonString(sender.host.toProto()));

		expect(sim.encounter.getName()).toBe('Phase two');
		expect(sim.encounter.isMyEncounter()).toBe(false);
	});

	it('applies only the rotation from a rotation-only export', async () => {
		const settings = IndividualSimSettings.fromJson({ player: { cooldowns: {}, rotation: { type: 'TypeAuto' } } });
		expect(await importedCategories(settings)).toEqual([SimSettingCategories.Rotation]);
	});

	it('applies talents only when the export carries a talent string', async () => {
		expect(await importedCategories(IndividualSimSettings.fromJson({ player: { talentsString: '05' } }))).toEqual([SimSettingCategories.Talents]);
	});

	it('applies every category an export carries, and the UI settings only when present', async () => {
		const settings = IndividualSimSettings.fromJson({
			player: { equipment: {}, talentsString: '05', rotation: {}, consumables: {}, healingModel: {}, buffs: {} },
			raidBuffs: {},
			encounter: {},
		});
		expect(await importedCategories(settings)).toEqual([
			SimSettingCategories.Gear,
			SimSettingCategories.Talents,
			SimSettingCategories.Rotation,
			SimSettingCategories.Consumes,
			SimSettingCategories.Miscellaneous,
			SimSettingCategories.External,
			SimSettingCategories.Encounter,
		]);

		IndividualSimSettings.mergePartial(settings, { settings: {} });
		expect(await importedCategories(settings)).toContain(SimSettingCategories.UISettings);
	});
});
