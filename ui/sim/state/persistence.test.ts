import { Spec } from '@generated/proto/common';
import { DatabaseFilters, IndividualSimSettings, SimSettings, StatFilters } from '@generated/proto/ui';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { memoryEnv } from '@testing/sim/memory_env';
import {
	envelope,
	loadSettings,
	senderStatFilters as linkedStatFilters,
	STORAGE_KEY,
	viewerFiltersWith,
	viewerStatFilters as savedStatFilters,
} from '@testing/sim/stat_filters';
import { beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const savedWith = (statFilters: StatFilters | undefined) => envelope(SimSettings.create({ filters: DatabaseFilters.create(viewerFiltersWith(statFilters)) }));

// A link's hash, and the `i` parameter that names its categories: `u` is the one that carries settings.
const linkLocation = async (settings: IndividualSimSettings, categories: string | null) => {
	const deflated = new Blob([IndividualSimSettings.toBinary(settings).slice()]).stream().pipeThrough(new CompressionStream('deflate'));
	return {
		hash: '#' + btoa(String.fromCharCode(...new Uint8Array(await new Response(deflated).arrayBuffer()))),
		search: categories == null ? '' : `?i=${categories}`,
	};
};

const load = async (location: { hash: string; search: string } = { hash: '', search: '' }, saved = savedWith(savedStatFilters)) => {
	const env = memoryEnv(location);
	env.storage.setItem(STORAGE_KEY, IndividualSimSettings.toJsonString(saved));
	const built = await makeIndividualSim({ env });
	loadSettings(built);
	return built.sim.getFilters();
};

const lessStatFilters = (filters: DatabaseFilters) => ({ ...filters, statFilters: undefined });

beforeAll(() => serveDatabase());

describe('loadIndividualSettings', () => {
	it('restores the autosaved stat filters', async () => {
		expect(StatFilters.equals((await load()).statFilters!, savedStatFilters)).toBe(true);
	});

	it('keeps the autosaved stat filters over the ones a link carries, and takes its other filters', async () => {
		const linked = envelope(SimSettings.create({ filters: DatabaseFilters.create({ minIlvl: 50, statFilters: linkedStatFilters }) }));

		const filters = await load(await linkLocation(linked, 'gu'));

		expect(StatFilters.equals(filters.statFilters!, savedStatFilters)).toBe(true);
		expect(filters.minIlvl).toBe(50);
	});

	it.each([
		['filters without stat filters', envelope(SimSettings.create({ filters: DatabaseFilters.create({ minIlvl: 50 }) }))],
		['settings without filters', envelope(SimSettings.create({ iterations: 1000 }))],
		['no settings at all', envelope()],
	])('keeps the autosaved stat filters through a link with %s, and every other filter as before', async (_name, linked) => {
		const filters = await load(await linkLocation(linked, 'gu'));

		const withoutStatFiltersInPlay = await load(await linkLocation(linked, 'gu'), savedWith(undefined));
		expect(StatFilters.equals(filters.statFilters!, savedStatFilters)).toBe(true);
		expect(DatabaseFilters.equals(lessStatFilters(filters), lessStatFilters(withoutStatFiltersInPlay))).toBe(true);
	});

	it('leaves every autosaved filter as it was under a link that carries no settings category', async () => {
		const linked = envelope(SimSettings.create({ filters: DatabaseFilters.create({ minIlvl: 50, statFilters: linkedStatFilters }) }));

		expect(DatabaseFilters.equals(await load(await linkLocation(linked, null)), await load())).toBe(true);
	});
});

describe('Show power tools', () => {
	it('starts off', async () => {
		const { sim } = await makeIndividualSim();

		expect(sim.getShowPowerTools()).toBe(false);
	});

	it("is one setting for every spec page, outside a page's saved settings, and its defaults keep it", async () => {
		const env = memoryEnv();
		(await makeIndividualSim({ env })).sim.setShowPowerTools(true);

		const protection = await makeIndividualSim({ spec: Spec.SpecProtectionWarrior, env });
		expect(protection.sim.getShowPowerTools()).toBe(true);

		loadSettings(protection);
		protection.host.applyDefaults();
		expect(protection.sim.getShowPowerTools()).toBe(true);
		expect(env.storage.getItem(STORAGE_KEY)).not.toContain('showPowerTools');
	});

	it('stays off on the next page once turned off', async () => {
		const env = memoryEnv();
		(await makeIndividualSim({ env })).sim.setShowPowerTools(true);
		(await makeIndividualSim({ env })).sim.setShowPowerTools(false);

		expect((await makeIndividualSim({ env })).sim.getShowPowerTools()).toBe(false);
	});
});
