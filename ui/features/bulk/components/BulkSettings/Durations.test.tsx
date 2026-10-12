import { BulkDuration } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';
import { patchSlice } from '@sim/state/sim_store';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Durations } from './Durations';

const mount = (extras: Array<[number, number]> = [], useHealth = false) => {
	const player = makeBulkPlayer();
	patchSlice(player.sim.store, 'encounter', { duration: 180, durationVariation: 5, useHealth });
	patchBulkState(player, { durations: extras.map(([duration, durationVariation]) => BulkDuration.create({ duration, durationVariation })) });
	const host = { player, sim: player.sim } as never;
	render(
		<SimHostProvider host={host}>
			<Durations />
		</SimHostProvider>,
	);
	return player;
};

const inputs = () => screen.getAllByRole('textbox') as HTMLInputElement[];

const commit = (input: HTMLInputElement, text: string) => {
	fireEvent.input(input, { target: { value: text } });
	fireEvent.change(input, { target: { value: text } });
	fireEvent.blur(input);
};

describe('Durations', () => {
	it('adds a row a minute after the last one', () => {
		const player = mount();
		fireEvent.click(screen.getByRole('button', { name: 'bulk_tab.settings.durations.add' }));
		expect(bulkState(player).durations).toEqual([BulkDuration.create({ duration: 240, durationVariation: 5 })]);
	});

	it('removes a row', () => {
		const player = mount([
			[90, 10],
			[300, 30],
		]);
		fireEvent.click(screen.getAllByRole('button', { name: /bulk_tab.settings.durations.remove/ })[0]);
		expect(bulkState(player).durations).toEqual([BulkDuration.create({ duration: 300, durationVariation: 30 })]);
	});

	it('the locked row follows the encounter', () => {
		const player = mount([[200, 5]]);
		expect(inputs()[0].value).toBe('180');
		expect(inputs()[0].disabled).toBe(true);
		expect(screen.queryByText('bulk_tab.settings.durations.duplicate')).toBeNull();

		act(() => patchSlice(player.sim.store, 'encounter', { duration: 200 }));

		expect(inputs()[0].value).toBe('200');
		expect(screen.getByText('bulk_tab.settings.durations.duplicate')).toBeTruthy();
	});

	it('clamps a typed duration to one second', () => {
		const player = mount([[90, 10]]);
		commit(inputs()[2], '0');
		expect(bulkState(player).durations[0].duration).toBe(1);
	});

	it('disables the group for a health fight', () => {
		mount([[90, 10]], true);
		expect(screen.getByText('bulk_tab.settings.durations.health_fight')).toBeTruthy();
		expect((screen.getByRole('button', { name: 'bulk_tab.settings.durations.add' }) as HTMLButtonElement).disabled).toBe(true);
		expect(inputs().every(input => input.disabled)).toBe(true);
	});
});
