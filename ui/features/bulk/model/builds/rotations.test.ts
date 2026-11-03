import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { SavedRotation } from '@generated/proto/ui';
import type { PresetRotation } from '@sim/presets/types';
import type { IndividualSimHost } from '@sim/sim_host';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it, vi } from 'vitest';

import { bulkRotationEntries, pickableBulkRotations, rotationLabel } from './rotations';

vi.mock('@i18n/config', () => ({
	default: { t: (key: string) => key },
}));

const SAVED_KEY = 'saved-rotations';

// One list item whose notes tell rotations apart, with a uuid on its condition as the editor sets one.
const rotation = (notes: string, uuid?: string, type = APLRotationType.TypeAPL) =>
	APLRotation.create({
		type,
		priorityList: [{ notes, action: { condition: { uuid: uuid === undefined ? undefined : { value: uuid } } } }],
	});

const presetRotation = (name: string, proto: APLRotation | undefined, extra: Partial<PresetRotation> = {}): PresetRotation => ({
	name,
	rotation: SavedRotation.create({ rotation: proto }),
	...extra,
});

describe('bulkRotationEntries', () => {
	it('lists every preset rotation, then the saved ones', () => {
		const entries = bulkRotationEntries(
			[
				presetRotation('Single target', rotation('st')),
				presetRotation('Gated', rotation('gated'), { enableWhen: () => false }),
				presetRotation('Empty', undefined),
			],
			[{ name: 'Mine', data: SavedRotation.create({ rotation: rotation('mine') }) }],
		);

		expect(entries.map(e => [e.name, e.ref.preset, e.rotation.priorityList[0]?.notes])).toEqual([
			['Single target', true, 'st'],
			['Gated', true, 'gated'],
			['Empty', true, undefined],
			['Mine', false, 'mine'],
		]);
		expect(APLRotation.equals(entries[2].rotation, APLRotation.create())).toBe(true);
	});
});

describe('pickableBulkRotations', () => {
	it('reads the saved rotations the Saved Rotations panel stores, without their uuids', () => {
		const player = makeBulkPlayer();
		player.sim.env.storage.setItem(
			SAVED_KEY,
			JSON.stringify({ Mine: SavedRotation.toJson(SavedRotation.create({ rotation: rotation('mine', 'uuid-1') })) }),
		);
		const host = {
			player,
			sim: player.sim,
			individualConfig: { presets: { rotations: [presetRotation('Single target', rotation('st'))] } },
			getSavedRotationStorageKey: () => SAVED_KEY,
		} as unknown as IndividualSimHost<any>;

		const entries = pickableBulkRotations(host);

		expect(entries.map(e => e.name)).toEqual(['Single target', 'Mine']);
		expect(APLRotation.equals(entries[1].rotation, rotation('mine'))).toBe(true);
	});
});

describe('labels', () => {
	it('labels the current rotation and names every other one', () => {
		expect(rotationLabel({ isCurrent: true, name: '' })).toBe('bulk_tab.results.current_rotation');
		expect(rotationLabel({ isCurrent: false, name: 'Mine' })).toBe('Mine');
	});
});
