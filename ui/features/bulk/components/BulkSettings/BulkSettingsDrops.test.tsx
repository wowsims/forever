import { saveEnchantPreset } from '@features/enchant-presets/model/enchant_presets';
import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { ItemSlot, Race } from '@generated/proto/common';
import { ContentPhase, UIEnchant } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { Gear } from '@sim/proto/gear';
import { raceToFaction } from '@sim/proto/utils';
import { bulkState, loadStoredBulkDropsCombinedGain, patchBulkState } from '@sim/settings/bulk_settings';
import { patchSlice, PLAYER_FIELDS, seedKeyed, type SimStore, zeroVersions } from '@sim/state/sim_store';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { BLACKROCK_SPIRE, dropsItemsFixture, dropsNamesFixture, REND, WORN_ITEM_ID } from '@testing/storybook/fixtures/bulk-drops';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { dropsRequestSettings, setDropsSources } from '../../model/drops/settings';
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
const CLOAK_AGILITY = 849;
const text = (testId: string) => screen.getByTestId(testId).textContent;

const mount = () => {
	localStorage.clear();
	const items = dropsItemsFixture();
	// The worn Doomsaw is a two-hander.
	const gear = new Gear({ [WORN_SLOTS[0]]: makeEquippedItem({ item: items.find(item => item.id === WORN_ITEM_ID)! }) });
	const player: Player<any> = makeBulkPlayer({
		getGear: () => gear,
		getFaction: () => raceToFaction[player.sim.store.getState().players[player.storeKey].race as Race],
		getSpec: () => PlayerSpecs.DpsWarrior.specID,
		getClass: () => PlayerSpecs.DpsWarrior.classID,
		getEnchants: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotBack ? [UIEnchant.create({ effectId: CLOAK_AGILITY })] : []),
		sim: {
			isNative: false,
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
	seedKeyed(player.sim.store, 'players', player.storeKey, { race: Race.RaceHuman, gear, v: zeroVersions(PLAYER_FIELDS) } as never);
	patchBulkState(player, { mode: 'drops' });
	render(
		<SimHostProvider host={{ player, sim: player.sim, getStorageKey: (key: string) => `test-${key}` } as never}>
			<BulkSettings />
		</SimHostProvider>,
	);
	return player;
};

describe('BulkSettings Drops run card on the fixture database', () => {
	it('counts the picked items off the ticked source and turns Find next drop on once candidates are counted', () => {
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

	it('notes the one-handers and off-hand items the worn two-hander skips', () => {
		const player = mount();
		expect(screen.queryAllByTestId('drops-run-two-hander-note')).toHaveLength(0);

		act(() => setDropsSources(player, [`drop:${BLACKROCK_SPIRE}:${REND}`]));

		// Felstriker; the Lionheart Gauntlets go in no hand.
		expect(screen.getAllByTestId('drops-run-two-hander-note').map(note => note.textContent)).toEqual([
			english.bulk_tab.drops.run.two_hander_skipped_one.replace('{{count}}', '1'),
		]);
	});

	it('stores the Combined gain tick, off until ticked, and leaves the count alone', async () => {
		const player = mount();
		const box = screen.getByRole('checkbox', { name: english.bulk_tab.drops.run.combined_gain });
		const version = bulkState(player).v.settings;
		expect(box.getAttribute('aria-checked')).toBe('false');

		await userEvent.click(box);

		expect(bulkState(player).dropsCombinedGain).toBe(true);
		expect(loadStoredBulkDropsCombinedGain(player)).toBe(true);
		expect(bulkState(player).v.settings).toBe(version);
		expect(box.getAttribute('aria-checked')).toBe('true');
	});

	it('sends the enchants of a preset ticked in its Enchant presets group, which ticks it for Gear too', async () => {
		const player = mount();
		act(() => {
			saveEnchantPreset(player, { name: 'Agility', slots: [{ slot: ItemSlot.ItemSlotBack, effectIds: [CLOAK_AGILITY] }] });
			setDropsSources(player, [`drop:${BLACKROCK_SPIRE}:${REND}`]);
		});
		fireEvent.click(screen.getByTestId('bulk-settings-group-enchants-trigger'));

		await userEvent.click(within(screen.getByTestId('bulk-enchant-presets')).getByRole('checkbox'));

		expect(bulkState(player).enchantPresets.map(preset => [preset.name, preset.checked])).toEqual([['Agility', true]]);
		expect(dropsRequestSettings(player).enchants).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [CLOAK_AGILITY] })]);
	});
});
