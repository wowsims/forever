import { Stat } from '@generated/proto/common';
import { Stats, UnitStat } from '@sim/proto/stats';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { BonusStatsLink } from './BonusStatsLink';

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);

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
});
