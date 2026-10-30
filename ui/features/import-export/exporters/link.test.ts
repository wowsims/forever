import { SimSettingCategories } from '@sim/constants/sim_settings';
import { tryParseUrlLocation } from '@sim/state/sim_links';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { senderStatFilters } from '@testing/sim/stat_filters';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { createLink } from './link';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

describe('createLink', () => {
	it('leaves the stat filters out of a link that carries the other filters', async () => {
		const { sim, host } = await makeIndividualSim();
		sim.setFilters({ ...sim.getFilters(), minIlvl: 50, statFilters: senderStatFilters });

		const url = new URL(createLink(host, [SimSettingCategories.Gear, SimSettingCategories.UISettings]));
		const filters = tryParseUrlLocation(url)!.settings.settings!.filters!;

		expect(filters.statFilters).toBeUndefined();
		expect(filters.minIlvl).toBe(50);
		expect(sim.getFilters().statFilters).toBeDefined();
	});

	it("carries the encounter's name, and leaves out that it is one of My encounters", async () => {
		const { sim, host } = await makeIndividualSim();
		sim.encounter.setName('Phase two');
		sim.encounter.setMyEncounter(true);

		const url = new URL(createLink(host, [SimSettingCategories.Encounter]));

		expect(tryParseUrlLocation(url)!.settings.encounter!.name).toBe('Phase two');
		expect(tryParseUrlLocation(url)!.settings.encounter!.myEncounter).toBe(false);
		expect(sim.encounter.isMyEncounter()).toBe(true);
	});
});
