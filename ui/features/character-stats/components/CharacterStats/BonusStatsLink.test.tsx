import { PseudoStat, Stat } from '@generated/proto/common';
import { Stats, UnitStat } from '@sim/proto/stats';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, cleanup, fireEvent, screen } from '@testing-library/react';
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

	it('marks a percent stat field with the percent sign and a points stat field without', async () => {
		await openPicker(Stat.StatCritRating, 14);
		expect(textbox().labels![0].textContent).toMatch(/ \(sidebar\.character_stats\.percent_suffix\)$/);
		cleanup();

		await openPicker(Stat.StatAgility, 14);
		expect(textbox().labels![0].textContent).not.toContain('percent_suffix');
	});

	it('takes a school hit row as a percent, labelled with the percent sign and stored as typed', async () => {
		let bonusStats = new Stats().withPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire, 2);
		const setBonusStats = vi.fn((next: Stats) => (bonusStats = next));
		const player = { storeKey: 0, sim: { store: createSimStore() }, getBonusStats: () => bonusStats, setBonusStats };
		renderWithHost(<BonusStatsLink unitStat={UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire)} />, { player });
		await act(async () => {
			fireEvent.click(screen.getByTestId('add-bonus-stats'));
		});
		expect(textbox().labels![0].textContent).toMatch(/ \(sidebar\.character_stats\.percent_suffix\)$/);
		expect(Number(textbox().value)).toBe(2);

		textbox().value = '1.5';
		await act(async () => {
			fireEvent.change(textbox());
		});

		expect(bonusStats.getPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire)).toBe(1.5);
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

	it.each([
		['defense', Stat.StatDefenseRating],
		['agility', Stat.StatAgility],
	])('takes points for %s', async (_name, stat) => {
		const { setBonusStats } = await openPicker(stat, 7);
		expect(textbox().value).toBe('7');

		textbox().value = '12';
		await act(async () => {
			fireEvent.change(textbox());
		});

		expect(setBonusStats.mock.calls[0][0].getStat(stat)).toBe(12);
	});
});
