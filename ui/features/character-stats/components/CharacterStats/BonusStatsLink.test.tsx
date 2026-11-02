import { Stat } from '@generated/proto/common';
import { Stats, UnitStat } from '@sim/proto/stats';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { BonusStatsLink } from './BonusStatsLink';

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);

const openPicker = async (stat: Stat, rating: number) => {
	let bonusStats = new Stats().withStat(stat, rating);
	const setBonusStats = vi.fn((next: Stats) => (bonusStats = next));
	const player = { storeKey: 0, sim: { store: createSimStore() }, getBonusStats: () => bonusStats, setBonusStats };
	renderWithHost(<BonusStatsLink unitStat={UnitStat.fromStat(stat)} />, { player });
	await act(async () => {
		fireEvent.click(screen.getByTestId('add-bonus-stats'));
	});
	return { setBonusStats, bonusStats: () => bonusStats };
};
const textbox = () => screen.getByRole('textbox') as HTMLInputElement;

describe('BonusStatsLink', () => {
	it('opens a picker labelled with its stat, showing the bonus on it, and writes the entry', async () => {
		let bonusStats = new Stats().withStat(Stat.StatStrength, 5);
		const setBonusStats = vi.fn((next: Stats) => (bonusStats = next));
		const player = { storeKey: 0, sim: { store: createSimStore() }, getBonusStats: () => bonusStats, setBonusStats };
		renderWithHost(<BonusStatsLink unitStat={STRENGTH} />, { player });

		const button = screen.getByTestId('add-bonus-stats');
		expect(button.getAttribute('aria-label')).toMatch(/^sidebar\.character_stats\.bonus_prefix /);
		await act(async () => {
			fireEvent.click(button);
		});

		const input = (await screen.findByLabelText(button.getAttribute('aria-label')!, { selector: 'input' })) as HTMLInputElement;
		expect(input.id).toBe(`character-bonus-stat-${STRENGTH.getKey()}`);
		expect(input.value).toBe('5');
		await act(async () => {
			input.value = '20';
			fireEvent.change(input);
		});
		expect(setBonusStats).toHaveBeenCalledTimes(1);
		expect(setBonusStats.mock.calls[0][0].getStat(Stat.StatStrength)).toBe(20);
	});

	it('has no unit toggle', async () => {
		await openPicker(Stat.StatCritRating, 14);
		expect(screen.queryByTestId('bonus-stats-unit-toggle')).toBeNull();
	});

	it('shows a percent stat as its percent and stores a typed percent as a fractional rating', async () => {
		const { setBonusStats, bonusStats } = await openPicker(Stat.StatCritRating, 14);
		expect(Number(textbox().value)).toBeCloseTo(1);

		textbox().value = '0.33';
		await act(async () => {
			fireEvent.change(textbox());
		});

		expect(setBonusStats).toHaveBeenCalledTimes(1);
		expect(bonusStats().getStat(Stat.StatCritRating)).toBe(0.33 * 14);
		expect(Number(textbox().value)).toBeCloseTo(0.33);
	});

	it.each([Stat.StatDefenseRating, Stat.StatAgility])('takes points for %s', async stat => {
		const { setBonusStats } = await openPicker(stat, 7);
		expect(textbox().value).toBe('7');

		textbox().value = '12';
		await act(async () => {
			fireEvent.change(textbox());
		});

		expect(setBonusStats.mock.calls[0][0].getStat(stat)).toBe(12);
	});
});
