import { BulkTalentRef, SavedTalents } from '@generated/proto/ui';
import { PlayerSpecs } from '@sim/player/specs';
import { bulkState, loadStoredBulkTalents } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it, vi } from 'vitest';

import { bulkTalentEntries, bulkTalentList, bulkTalentRuns, formatTalentsSummary, pickableBulkTalents, setBulkTalentChecked, talentsLabel } from './talents';

vi.mock('@i18n/config', () => ({
	default: { t: (key: string) => key },
}));

const SPEC = PlayerSpecs.DpsWarrior;
const SAVED_KEY = 'saved-talents';
const ARMS = '3030500130';
const FURY = '-0505000540501';
const LIVE = '30305001302';

const saved = (talentsString: string) => SavedTalents.create({ talentsString });

const ref = (name: string, preset = false) => BulkTalentRef.create({ name, preset });

const entry = (name: string, talentsString: string, preset = false) => ({ ref: ref(name, preset), name, talentsString });

const entries = [entry('Arms', ARMS, true), entry('Fury', FURY)];

const writeSaved = (player: ReturnType<typeof makeBulkPlayer>, record: Record<string, string>) =>
	player.sim.env.storage.setItem(
		SAVED_KEY,
		JSON.stringify(Object.fromEntries(Object.entries(record).map(([name, s]) => [name, SavedTalents.toJson(saved(s))]))),
	);

const hostWith = (presets: Array<{ name: string; data: SavedTalents }>, record: Record<string, string>) => {
	const player = makeBulkPlayer();
	writeSaved(player, record);
	return {
		player,
		host: {
			player,
			sim: player.sim,
			individualConfig: { presets: { talents: presets } },
			getSavedTalentsStorageKey: () => SAVED_KEY,
		} as unknown as IndividualSimHost<any>,
	};
};

describe('bulkTalentEntries', () => {
	it('lists the presets, then the saved setups, in panel order', () => {
		expect(
			bulkTalentEntries([{ name: 'Arms', data: saved(ARMS) }], [{ name: 'Fury', data: saved(FURY) }]).map(e => [e.name, e.ref.preset, e.talentsString]),
		).toEqual([
			['Arms', true, ARMS],
			['Fury', false, FURY],
		]);
	});
});

describe('bulkTalentRuns', () => {
	it('runs the current talents first, then each checked setup in list order', () => {
		expect(bulkTalentRuns(SPEC, LIVE, entries, [ref('Fury'), ref('Arms', true)], true)).toEqual([
			{ name: '', talentsString: LIVE, isCurrent: true },
			{ name: 'Arms', talentsString: ARMS, isCurrent: false },
			{ name: 'Fury', talentsString: FURY, isCurrent: false },
		]);
	});

	it('runs the current talents alone with power tools off, whatever is checked', () => {
		expect(bulkTalentRuns(SPEC, LIVE, entries, [ref('Arms', true)], false)).toEqual([{ name: '', talentsString: LIVE, isCurrent: true }]);
	});

	it('skips a checked setup equal to an earlier checked one', () => {
		const runs = bulkTalentRuns(SPEC, LIVE, [entry('Arms', ARMS, true), entry('Arms again', `${ARMS}0`)], [ref('Arms', true), ref('Arms again')], true);

		expect(runs.map(run => run.name)).toEqual(['', 'Arms']);
	});
});

describe('bulkTalentList', () => {
	it('hides a setup equal to the current talents once both strings are normalised', () => {
		expect(bulkTalentList(SPEC, LIVE, [entry('Same', `${LIVE}000--`)], [ref('Same')])).toEqual([]);
	});

	it('flags a checked setup equal to an earlier checked one after normalising', () => {
		expect(
			bulkTalentList(
				SPEC,
				LIVE,
				[entry('Arms', ARMS, true), entry('Arms again', `${ARMS}0`), entry('Fury', FURY)],
				[ref('Arms', true), ref('Arms again')],
			).map(item => [item.name, item.checked, item.duplicate]),
		).toEqual([
			['Arms', true, false],
			['Arms again', true, true],
			['Fury', false, false],
		]);
	});

	it('tells a preset from a saved setup of the same name', () => {
		const list = bulkTalentList(SPEC, LIVE, [entry('Arms', ARMS, true), entry('Arms', FURY)], [ref('Arms', true)]);

		expect(list.map(item => [item.talentsString, item.checked])).toEqual([
			[ARMS, true],
			[FURY, false],
		]);
	});
});

describe('pickableBulkTalents', () => {
	it('resolves refs at call time: a deleted setup drops out, an edited one carries its new string', () => {
		const { player, host } = hostWith([{ name: 'Arms', data: saved(ARMS) }], { Fury: FURY, Arms2: ARMS });

		expect(pickableBulkTalents(host).map(e => [e.name, e.talentsString])).toEqual([
			['Arms', ARMS],
			['Fury', FURY],
			['Arms2', ARMS],
		]);

		writeSaved(player, { Arms2: LIVE });

		expect(pickableBulkTalents(host).map(e => [e.name, e.talentsString])).toEqual([
			['Arms', ARMS],
			['Arms2', LIVE],
		]);
	});
});

describe('setBulkTalentChecked', () => {
	it('stores the checked refs in list order so they load back', () => {
		const player = makeBulkPlayer();
		const pickable = entries.map(e => e.ref);

		setBulkTalentChecked(player, ref('Fury'), true, pickable);
		setBulkTalentChecked(player, ref('Arms', true), true, pickable);

		expect(loadStoredBulkTalents(player)?.talents).toEqual([ref('Arms', true), ref('Fury')]);
		expect(bulkState(player).talents).toEqual([ref('Arms', true), ref('Fury')]);
	});

	it('drops a checked ref whose setup is no longer pickable', () => {
		const player = makeBulkPlayer();
		setBulkTalentChecked(player, ref('Fury'), true, [ref('Fury')]);

		setBulkTalentChecked(player, ref('Arms', true), true, [ref('Arms', true)]);

		expect(bulkState(player).talents).toEqual([ref('Arms', true)]);
	});
});

describe('labels', () => {
	it('labels the current talents and names every other setup', () => {
		expect(talentsLabel({ isCurrent: true, name: '' })).toBe('bulk_tab.results.current_talents');
		expect(talentsLabel({ isCurrent: false, name: 'Arms' })).toBe('Arms');
	});

	it('summarises a setup by its string, or says it spends no points', () => {
		expect(formatTalentsSummary(ARMS)).toBe(ARMS);
		expect(formatTalentsSummary('')).toBe('bulk_tab.settings.talents.no_points');
	});
});
