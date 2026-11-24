import { ItemSlot, Race } from '@generated/proto/common';
import { ContentPhase, DatabaseFilters } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { Gear } from '@sim/proto/gear';
import { raceToFaction } from '@sim/proto/utils';
import { patchBulkState } from '@sim/settings/bulk_settings';
import { patchSlice, PLAYER_FIELDS, seedKeyed, type SimStore, zeroVersions } from '@sim/state/sim_store';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { BLACKROCK_SPIRE, dropsItemsFixture, dropsNamesFixture, REND, WORN_ITEM_ID } from '@testing/storybook/fixtures/bulk-drops';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { act, render, screen } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { setDropsSources } from '../../model/drops/settings';
import { BulkSettings } from './BulkSettings';

// The Drops pick comes off the fixture database and the live gear: nothing here mocks `useDropsSetup`.
vi.mock('@sim/hooks/useSimReady', () => ({ useSimReady: () => true }));
vi.mock('@features/settings/hooks/useCurrentSettings', () => ({ useCurrentSettings: () => null }));
vi.mock('../BatchEncounter', () => ({ BatchEncounter: () => null }));

beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

const names = dropsNamesFixture();
const WORN_SLOTS = [ItemSlot.ItemSlotMainHand];
const text = (testId: string) => screen.getByTestId(testId).textContent;

const mount = () => {
	localStorage.clear();
	const items = dropsItemsFixture();
	const player: Player<any> = makeBulkPlayer({
		getGear: () => new Gear({ [WORN_SLOTS[0]]: makeEquippedItem({ item: items.find(item => item.id === WORN_ITEM_ID)! }) }),
		getFaction: () => raceToFaction[player.sim.store.getState().players[player.storeKey].race as Race],
		getSpec: () => PlayerSpecs.DpsWarrior.specID,
		getClass: () => PlayerSpecs.DpsWarrior.classID,
		sim: {
			isNative: false,
			getFilters(this: { store: SimStore }) {
				return DatabaseFilters.clone(this.store.getState().sim.filters);
			},
			getPhase(this: { store: SimStore }) {
				return this.store.getState().sim.phase;
			},
			db: {
				getAllItems: () => items,
				getZone: (id: number) => ({ name: names.zone(id) }),
				getNpc: (id: number) => ({ name: names.npc(id) }),
				getRepFaction: () => undefined,
			},
		},
	});
	patchSlice(player.sim.store, 'sim', { phase: ContentPhase.ContentPhase1 });
	seedKeyed(player.sim.store, 'players', player.storeKey, { race: Race.RaceHuman, gear: new Gear({}), v: zeroVersions(PLAYER_FIELDS) } as never);
	patchBulkState(player, { mode: 'drops' });
	render(
		<SimHostProvider host={{ player, sim: player.sim, getStorageKey: (key: string) => `test-${key}` } as never}>
			<BulkSettings />
		</SimHostProvider>,
	);
	return player;
};

describe('BulkSettings Drops run card on the fixture database', () => {
	it('counts the picked items off the ticked source and turns Simulate Drops on once candidates are counted', () => {
		const player = mount();
		const simulate = () => screen.getByTestId('bulk-settings-btn') as HTMLButtonElement;

		expect(text('drops-run-picked')).toBe('0');
		expect(simulate().disabled).toBe(true);

		// Rend drops Felstriker, the worn Doomsaw (not picked) and the Alliance-only Lionheart Gauntlets, which a Human keeps.
		act(() => setDropsSources(player, [`drop:${BLACKROCK_SPIRE}:${REND}`]));

		expect(text('drops-run-picked')).toBe('2');
		expect(simulate().disabled).toBe(true);

		act(() => patchBulkState(player, { combinations: 2 }));

		expect(simulate().disabled).toBe(false);
	});
});
