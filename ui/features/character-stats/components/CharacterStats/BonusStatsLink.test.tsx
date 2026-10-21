import { Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Stats, UnitStat } from '@sim/proto/stats';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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
const unitButton = (key: 'rating_suffix' | 'percent_label') => screen.getByRole('button', { name: i18n.t(`sidebar.character_stats.${key}`) });
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

	it.each([
		[Stat.StatHitRating, 20, 2, 2.5, 25],
		[Stat.StatCritRating, 14, 1, 1.071, 15],
		[Stat.StatHasteRating, 10, 1, 2.5, 25],
		[Stat.StatExpertiseRating, 10, 1, 0.64, 6],
		[Stat.StatDodgeRating, 24, 2, 1.5, 18],
	] as const)('accepts a percent bonus for rating stat %s, rounded to whole rating', async (stat, initialRating, initialPercent, percent, expectedRating) => {
		const { setBonusStats } = await openPicker(stat, initialRating);
		const percentButton = unitButton('percent_label');
		const ratingButton = unitButton('rating_suffix');
		expect(percentButton.getAttribute('aria-pressed')).toBe('true');
		expect(Number(textbox().value)).toBeCloseTo(initialPercent);

		await act(async () => {
			fireEvent.click(ratingButton);
		});
		expect(ratingButton.getAttribute('aria-pressed')).toBe('true');
		expect(textbox().value).toBe(String(initialRating));

		await act(async () => {
			fireEvent.click(percentButton);
		});
		textbox().value = String(percent);
		await act(async () => {
			fireEvent.change(textbox());
		});

		expect(setBonusStats).toHaveBeenCalledTimes(1);
		expect(setBonusStats.mock.calls[0][0].getStat(stat)).toBeCloseTo(expectedRating);
	});

	it('keeps the picker open and commits the edited value when switching units', async () => {
		const { setBonusStats, bonusStats } = await openPicker(Stat.StatHitRating, 20);

		await userEvent.clear(textbox());
		await userEvent.type(textbox(), '2.5');
		await userEvent.click(unitButton('rating_suffix'));

		await waitFor(() => expect(textbox().value).toBe('25'));
		expect(setBonusStats).toHaveBeenCalledTimes(1);
		expect(bonusStats().getStat(Stat.StatHitRating)).toBeCloseTo(25);
	});

	it('keeps the picker open when the pressed unit button never takes focus', async () => {
		await openPicker(Stat.StatHitRating, 20);
		const ratingButton = unitButton('rating_suffix');

		fireEvent.pointerDown(ratingButton);
		textbox().value = '2.5';
		fireEvent.change(textbox());
		await new Promise(resolve => requestAnimationFrame(resolve));
		fireEvent.click(ratingButton);

		await waitFor(() => expect(textbox().value).toBe('25'));
		await new Promise(resolve => setTimeout(resolve, 50));
		expect(document.querySelector('.bonus-stats-popover')?.classList.contains('react-tooltip__show')).toBe(true);
	});
});
