import { BulkDuration } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';
import { patchSlice } from '@sim/state/sim_store';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

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
		expect(inputs()[2].getAttribute('aria-invalid')).toBeNull();

		act(() => patchSlice(player.sim.store, 'encounter', { duration: 200 }));

		expect(inputs()[0].value).toBe('200');
		expect(inputs()[2].getAttribute('aria-invalid')).toBe('true');
		expect(inputs()[3].getAttribute('aria-invalid')).toBe('true');
		expect(inputs()[0].getAttribute('aria-invalid')).toBeNull();
	});

	it('rejects a duplicate value, restores the old one and clears the outline after a moment', () => {
		vi.useFakeTimers();
		try {
			const player = mount([
				[90, 10],
				[300, 10],
			]);

			commit(inputs()[4], '90');

			expect(bulkState(player).durations.map(row => [row.duration, row.durationVariation])).toEqual([
				[90, 10],
				[300, 10],
			]);
			expect(inputs()[4].value).toBe('300');
			expect(inputs()[4].getAttribute('aria-invalid')).toBe('true');
			expect(inputs()[5].getAttribute('aria-invalid')).toBe('true');
			expect(inputs()[2].getAttribute('aria-invalid')).toBeNull();

			act(() => {
				vi.advanceTimersByTime(1500);
			});

			expect(inputs()[4].getAttribute('aria-invalid')).toBeNull();
			expect(inputs()[5].getAttribute('aria-invalid')).toBeNull();
		} finally {
			vi.useRealTimers();
		}
	});

	it('drops the rejected flash when a row is removed', () => {
		vi.useFakeTimers();
		try {
			mount([
				[90, 10],
				[300, 10],
			]);

			commit(inputs()[4], '90');
			expect(inputs()[4].getAttribute('aria-invalid')).toBe('true');

			fireEvent.click(screen.getAllByRole('button', { name: /bulk_tab.settings.durations.remove/ })[0]);

			expect(inputs().some(input => input.getAttribute('aria-invalid') === 'true')).toBe(false);
		} finally {
			vi.useRealTimers();
		}
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
		expect((screen.getByRole('button', { name: /bulk_tab.settings.durations.remove/ }) as HTMLButtonElement).disabled).toBe(true);
	});
});
