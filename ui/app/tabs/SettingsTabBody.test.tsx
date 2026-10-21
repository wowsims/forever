import { BUFFS_CONFIG, DEBUFFS_CONFIG, DEBUFFS_MISC_CONFIG, GreaterBlessingOfSalvation, PARTY_BUFFS_CONFIG } from '@features/settings/model/buffs_debuffs';
import type { RenderableStatOptions } from '@features/settings/model/stat_options';
import { withSelection } from '@features/stat-filters/model/stat_filters';
import { APLRotation } from '@generated/proto/apl';
import { PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import { patchSlice } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { act, fireEvent, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

// Everything the tab mounts besides the buff and debuff sections; none of it is under test here.
vi.mock('@features/encounter', () => ({ EncounterPicker: () => null, SavedEncounter: () => null }));
vi.mock('@features/gear/components/SelectorModal', () => ({ SelectorModal: () => null }));
vi.mock('../PresetConfigurationPicker', () => ({ PresetConfigurationPicker: () => null }));
vi.mock('@sim/hooks/useSimReady', () => ({ useSimReady: () => true }));
vi.mock('@features/settings', () => ({
	BuffSection: ({ section, empty, filter, children }: { section: string; empty: boolean; filter?: ReactNode; children: ReactNode }) => (
		<section data-section={section} data-empty={empty}>
			{filter}
			{children}
		</section>
	),
	RaidBuffs: ({ options, miscOptions }: { options: RenderableStatOptions[]; miscOptions: RenderableStatOptions[] }) => {
		rendered.push([...options, ...miscOptions]);
		return null;
	},
	ConsumesFilters: () => null,
	ConsumesPicker: () => null,
	CustomSection: () => null,
	OtherSettings: () => null,
	PlayerSettings: () => null,
	SavedSettings: () => null,
}));

let rendered: RenderableStatOptions[][] = [];
const { SettingsTabBody } = await import('./SettingsTabBody');

const SPEC = {
	autoRotation: () => APLRotation.create(),
	presets: { epWeights: [] },
	displayStats: [UnitStat.fromStat(Stat.StatStrength)],
	// Stamina keeps Blood Pact on offer, so the Party Buffs section renders between the raid buffs and the debuffs.
	epStats: [Stat.StatAttackPower, Stat.StatStamina],
	includeBuffDebuffInputs: [] as unknown[],
	excludeBuffDebuffInputs: [] as unknown[],
	otherInputs: { inputs: [] },
	encounterPicker: { showExecuteProportion: false },
};

beforeAll(() => serveDatabase());

const setup = async (overrides: Partial<typeof SPEC> = {}) => {
	const config = { ...SPEC, ...overrides };
	const { sim, player } = await makeIndividualSim({ config });
	rendered = [];
	renderWithHost(<SettingsTabBody />, { sim, player, individualConfig: config, getStorageKey: (keyPart: string) => `test${keyPart}` });
	return { sim, player };
};

const lastRaidPartyAndDebuffLists = () => rendered.slice(-3);
const unpickedOfferCounts = () => lastRaidPartyAndDebuffLists().map(rows => rows.filter(row => !row.config.keptBy).length);
const rowCounts = () => lastRaidPartyAndDebuffLists().map(rows => rows.length);
const section = (name: string) => document.querySelector(`[data-section="${name}"]`);

const showAll = async (slot: 'buffs' | 'debuffs') => {
	await act(async () => {
		fireEvent.click(screen.getByTestId(`${slot}-filter-button`));
	});
	act(() => {
		fireEvent.click(screen.getByTestId(`${slot}-filter-all`));
	});
};

describe('SettingsTabBody buff and debuff filters', () => {
	it('offers every buff once the buff filter shows every stat, and leaves the debuffs to their own filter', async () => {
		await setup();
		const [raidBuffs, , debuffs] = unpickedOfferCounts();
		const rows = rowCounts();
		expect(raidBuffs).toBeLessThan(rows[0]);

		await showAll('buffs');
		expect(unpickedOfferCounts()).toEqual([rows[0], rows[1], debuffs]);
		expect(rowCounts()).toEqual(rows);
	});

	it('offers every debuff once the debuff filter shows every stat, and leaves the buffs to their own filter', async () => {
		await setup();
		const [raidBuffs, partyBuffs, debuffs] = unpickedOfferCounts();
		const rows = rowCounts();
		expect(debuffs).toBeLessThan(rows[2]);

		await showAll('debuffs');
		expect(unpickedOfferCounts()).toEqual([raidBuffs, partyBuffs, rows[2]]);
		expect(rowCounts()).toEqual(rows);
	});
});

describe('SettingsTabBody sections', () => {
	const noPartyBuffStats = { displayStats: [], epStats: [Stat.StatIntellect] };

	it('leaves Party Buffs out when the filter offers no party buff and none is picked', async () => {
		await setup(noPartyBuffStats);

		expect(section('party_buffs')).toBeNull();
	});

	it('shows Party Buffs while a party buff the filter hides is picked, and drops it once cleared', async () => {
		const { player } = await setup(noPartyBuffStats);
		const party = player.getParty()!;

		act(() => party.setBuffs(PartyBuffs.create({ bloodPact: true })));
		expect(section('party_buffs')).not.toBeNull();

		act(() => party.setBuffs(PartyBuffs.create()));
		expect(section('party_buffs')).toBeNull();
	});

	it('collapses the raid buffs body while the filter hides every stat and no raid buff is picked', async () => {
		const { sim, player } = await setup({ excludeBuffDebuffInputs: [GreaterBlessingOfSalvation] });
		expect(section('raid_buffs')?.getAttribute('data-empty')).toBe('false');

		act(() => sim.setFilters(withSelection(sim.getFilters(), 'buffs', [])));
		expect(section('raid_buffs')?.getAttribute('data-empty')).toBe('true');

		act(() => player.getRaid()!.setBuffs(RaidBuffs.create({ arcaneBrilliance: true })));
		expect(section('raid_buffs')?.getAttribute('data-empty')).toBe('false');
	});
});

describe('SettingsTabBody offered flags', () => {
	it('reads no option again for a write to an unrelated slice, and still follows the raid', async () => {
		const spies = [...BUFFS_CONFIG, ...PARTY_BUFFS_CONFIG, ...DEBUFFS_CONFIG, ...DEBUFFS_MISC_CONFIG]
			.flatMap(option => ('inputs' in option.config ? option.config.inputs : [option.config]))
			.filter(config => 'getValue' in config)
			.map(config => vi.spyOn(config as { getValue: (player: unknown) => unknown }, 'getValue'));
		const reads = () => spies.reduce((total, spy) => total + spy.mock.calls.length, 0);
		const { sim, player } = await setup({ excludeBuffDebuffInputs: [GreaterBlessingOfSalvation] });
		act(() => sim.setFilters(withSelection(sim.getFilters(), 'buffs', [])));
		expect(section('raid_buffs')?.getAttribute('data-empty')).toBe('true');

		const before = reads();
		act(() => patchSlice(sim.store, 'ui', {}));
		expect(reads()).toBe(before);

		act(() => player.getRaid()!.setBuffs(RaidBuffs.create({ arcaneBrilliance: true })));
		expect(reads()).toBeGreaterThan(before);
		expect(section('raid_buffs')?.getAttribute('data-empty')).toBe('false');
		vi.restoreAllMocks();
	});
});
