import { APLRotation } from '@generated/proto/apl';
import { UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { patchSlice } from '@sim/state/sim_store';
import { makePlayer } from '@testing/apl/player';
import { act, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { UnitField } from './UnitField';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const config = { id: 'unit-field', getValue: () => undefined as UnitReference | undefined, setValue: () => {} };

let player: ReturnType<typeof makePlayer>;
let targets: Array<unknown>;
let pets: Array<unknown>;

const tree = (unitSet: 'targets' | 'players' | 'aura_sources') => (
	<SimHostProvider host={{ player, rootElem: document.body } as never}>
		<UnitField player={player as unknown as Player<any>} config={config} unitSet={unitSet} />
	</SimHostProvider>
);

const label = (key: string) => `rotation_tab.apl.helpers.unit_labels.${key}`;

const trigger = () => document.getElementById('unit-field')!;
const items = () => screen.queryAllByRole('menuitemradio').map(item => item.textContent);
const open = () => userEvent.click(trigger());

beforeEach(() => {
	document.body.innerHTML = '';
	targets = [];
	pets = [];
	player = makePlayer(APLRotation.create());
	player.sim.encounter = { targetsMetadata: { asList: () => targets.slice() } };
});

describe('UnitField', () => {
	it('offers the options of the unit set it was last rendered with, not the one it first read', async () => {
		const view = render(tree('targets'));
		await open();
		expect(items()).toEqual([label('current_target'), label('previous_target'), label('next_target')]);

		view.rerender(tree('players'));

		expect(items()).toEqual([label('current_target')]);
	});

	it('follows unit metadata, which is what puts a newly added target in the list', async () => {
		render(tree('targets'));
		await open();
		expect(items()).toHaveLength(3);

		await act(() => {
			targets.push({ getName: () => 'Target 1' });
			patchSlice(player.sim.store, 'sim', { metadataVersion: 1 });
		});

		expect(items()).toHaveLength(4);
	});

	it('keeps the same trigger node when a pet, and with it a submenu, arrives after mount', async () => {
		player.getPetMetadatas = () => ({ asList: () => pets.slice() });
		player.sim.raid.getActivePlayers = () => [];
		player.sim.getUnitMetadata = () => ({ getName: () => 'Player 1 - Shadowfiend' });
		render(tree('aura_sources'));
		const before = trigger();

		await act(() => {
			pets.push({});
			patchSlice(player.sim.store, 'sim', { metadataVersion: 1 });
		});

		expect(trigger()).toBe(before);
	});

	it('shows a pet whose icon is an ActionId with the url it resolves to', async () => {
		const pet = UnitReference.create({ type: UnitType.Pet, index: 0, owner: UnitReference.create({ type: UnitType.Self }) });
		const fill = vi.spyOn(ActionId.prototype, 'fill').mockImplementation(async function (this: ActionId) {
			return Object.assign(Object.create(ActionId.prototype), this, { name: 'Shadowfiend', iconUrl: 'shadowfiend.jpg' }) as ActionId;
		});
		player.getPetMetadatas = () => ({ asList: () => [{}] });
		player.sim.raid.getActivePlayers = () => [];
		player.sim.getUnitMetadata = () => ({ getName: () => 'Player 1 - Shadowfiend' });

		render(
			<SimHostProvider host={{ player, rootElem: document.body } as never}>
				<UnitField player={player as unknown as Player<any>} config={{ ...config, getValue: () => pet }} unitSet="aura_sources" />
			</SimHostProvider>,
		);
		await act(async () => {});

		const icon = within(trigger()).getByTestId('unit-picker-item-icon');
		expect(icon.getAttribute('src')).toBe('shadowfiend.jpg');
		fill.mockRestore();
	});
});
