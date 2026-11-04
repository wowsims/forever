import { RaidSimRequest, RaidSimResult } from '@generated/proto/api';
import { ItemSlot, RangedWeaponType } from '@generated/proto/common';
import i18n from '@i18n/config';
import { type CombatLog, parseAll } from '@sim/proto/combat_log';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import type { SimResultData } from '../../model/result_data';
import { Movement } from './Movement';

const state = vi.hoisted(() => ({ result: null as SimResultData | null }));
vi.mock('../../hooks/useSimResult', () => ({ useSimResult: () => state.result }));

// One player `p` at raid index 0, 12 yd out at the start, with a cat and a ranged weapon of the given type; `logs` is
// the run's log as the sim prints it.
const runOf = async (text: string, ranged = RangedWeaponType.RangedWeaponTypeBow): Promise<SimResultData> => {
	const logs = await parseAll(RaidSimResult.create({ logs: text.trim().replace(/^\s+/gm, '') }));
	const of = (pet: boolean) => logs.filter((log: CombatLog) => log.source && log.source.isPet === pet);
	return {
		result: {
			request: RaidSimRequest.create({
				requestId: 'a',
				raid: {
					parties: [
						{
							players: [
								{
									distanceFromTarget: 12,
									equipment: {
										items: Array.from({ length: ItemSlot.ItemSlotRanged + 1 }, (_, slot) => ({
											id: slot === ItemSlot.ItemSlotRanged ? 1 : 0,
										})),
									},
									database: { items: [{ id: 1, rangedWeaponType: ranged }] },
								},
							],
						},
					],
				},
			}),
			logs,
			firstIterationDuration: 30,
			getRaidIndexedPlayers: () => [{ name: 'p', index: 0, logs: of(false), pets: [{ name: 'Cat', logs: of(true) }] }],
			getTargets: () => [{ name: 'Target 1' }],
		},
		filter: {},
	} as unknown as SimResultData;
};

const WEAVE = `
	[-10.00] [p (#1)] Walking from 12.00 to 25.00 yards over 2.000s
	[-8.00] [p (#1)] Stopped walking at 25.00 yards
	[0.00] [p (#1)] Walking from 25.00 to 4.00 yards over 3.000s
	[0.00] [p (#1) - Cat] Walking from 12.00 to 4.00 yards over 1.000s
	[1.00] [p (#1) - Cat] Stopped walking at 4.00 yards
	[3.00] [p (#1)] Stopped walking at 4.00 yards
`;

const slider = () => screen.getByRole('slider');
const zone = (container: HTMLElement) => container.querySelector('[data-zone]');

beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});
beforeEach(() => {
	state.result = null;
});

describe('Movement', () => {
	it('starts at the prepull and reads the player out at the slider time', async () => {
		state.result = await runOf(WEAVE);
		const { container } = render(<Movement active />);

		expect(screen.getByRole('img', { name: 'The player and pets around the target' })).toBeTruthy();
		expect(slider().getAttribute('aria-valuetext')).toBe('-10.00 s');
		expect(screen.getByText('p: 12.00 yd')).toBeTruthy();
		expect(zone(container)?.textContent).toBe('At range');

		fireEvent.change(slider(), { target: { value: '1.5' } });
		expect(slider().getAttribute('aria-valuetext')).toBe('1.50 s');
		expect(screen.getByText('p: 14.50 yd')).toBeTruthy();

		fireEvent.change(slider(), { target: { value: '2.5' } });
		expect(zone(container)?.getAttribute('data-zone')).toBe('dead_zone');

		fireEvent.change(slider(), { target: { value: '30' } });
		expect(screen.getByText('p: 4.00 yd')).toBeTruthy();
		expect(zone(container)?.textContent).toBe('In melee range');
	});

	it('has no dead zone for a weapon that fires from anywhere', async () => {
		state.result = await runOf(WEAVE, RangedWeaponType.RangedWeaponTypeThrown);
		const { container } = render(<Movement active />);
		fireEvent.change(slider(), { target: { value: '2.5' } });
		expect(zone(container)?.getAttribute('data-zone')).toBe('ranged');
	});

	it('says so instead of drawing when no unit walked', async () => {
		state.result = await runOf('[1.00] [p (#1)] Dynamic stat change: {"DodgeRating": 12.000,}');
		render(<Movement active />);
		expect(screen.getByText('No unit walked in this run.')).toBeTruthy();
		expect(screen.queryByRole('img')).toBeNull();
	});

	it('holds a run that lands while the tab is closed', async () => {
		state.result = await runOf(WEAVE);
		const { rerender } = render(<Movement active={false} />);
		expect(screen.queryByRole('slider')).toBeNull();
		rerender(<Movement active />);
		expect(screen.getByRole('slider')).toBeTruthy();
	});
});
