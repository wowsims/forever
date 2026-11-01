import { SavedTalents } from '@generated/proto/ui';
import type { IndividualSimHost } from '@sim/sim_host';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it, vi } from 'vitest';

import { bulkTalentEntries, pickableBulkTalents, talentsLabel } from './talents';

vi.mock('@i18n/config', () => ({
	default: { t: (key: string) => key },
}));

const SAVED_KEY = 'saved-talents';
const ARMS = '3030500130';
const FURY = '-0505000540501';
const LIVE = '30305001302';

const saved = (talentsString: string) => SavedTalents.create({ talentsString });

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

describe('labels', () => {
	it('labels the current talents and names every other setup', () => {
		expect(talentsLabel({ isCurrent: true, name: '' })).toBe('bulk_tab.results.current_talents');
		expect(talentsLabel({ isCurrent: false, name: 'Arms' })).toBe('Arms');
	});
});
