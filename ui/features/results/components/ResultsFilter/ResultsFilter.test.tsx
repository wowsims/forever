import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimResultData } from '../../model/result_data';
import { ResultsFilter } from './ResultsFilter';
import { ALL_UNITS, hasTarget, simResultFilter, targetOptions } from './utils';

let resultData: SimResultData | null = null;
vi.mock('../../hooks/useSimResult', () => ({ useSimResult: () => resultData }));

const resultWith = (...targets: Array<{ index: number; iconUrl?: string; classColor?: string }>) =>
	({
		result: {
			getRaidIndexedPlayers: () => [],
			getTargets: () => targets,
			getTargetWithEncounterIndex: (index: number) => targets.find(target => target.index === index) ?? null,
		},
		filter: {},
	}) as never as SimResultData;

const healingResultWith = (heals: Array<{ unitIndex: number; healing: number; shielding: number }>, healerIndex = 0) =>
	({
		getRaidIndexedPlayers: () => [
			{
				spec: { isHealingSpec: true },
				unitIndex: healerIndex,
				getHealingActions: () => [
					{
						targets: heals.map(({ unitIndex, healing, shielding }) => ({ data: { unitIndex }, healing, shielding })),
					},
				],
			},
		],
		getTargets: () => [],
		getTargetWithEncounterIndex: () => null,
	}) as never as SimResultData['result'];

const mount = (target: number, onTargetChange = vi.fn()) => render(<ResultsFilter target={target} onTargetChange={onTargetChange} />);

const trigger = () => within(screen.getByTestId('results-filter-root')).getByRole('combobox', { hidden: true }) as HTMLButtonElement;
const items = () => screen.queryAllByRole('option');
const open = () => userEvent.click(trigger());

beforeEach(() => {
	resultData = null;
});

describe('ResultsFilter', () => {
	it('stays disabled until a run has produced targets', () => {
		mount(ALL_UNITS);

		expect(trigger().disabled).toBe(true);
		expect(trigger().textContent).toBe('results_tab.details.unit');
	});

	it('shows all targets plus one option per target of the run, in encounter order', async () => {
		resultData = resultWith({ index: 0 }, { index: 1 }, { index: 2 });
		mount(ALL_UNITS);
		expect(trigger().disabled).toBe(false);

		await open();
		expect(items().map(item => item.textContent)).toEqual([
			'results_tab.details.all_targets',
			'results_tab.details.target_number',
			'results_tab.details.target_number',
			'results_tab.details.target_number',
		]);
		expect(items()[0].getAttribute('aria-selected')).toBe('true');
	});

	it("takes each target option's icon off the unit it stands for", async () => {
		resultData = resultWith({ index: 0, iconUrl: 'boss.jpg', classColor: 'warrior' });
		mount(ALL_UNITS);
		await open();

		expect(within(items()[1]).getByTestId('unit-picker-item-icon').getAttribute('src')).toBe('boss.jpg');
		expect(within(items()[0]).queryByTestId('unit-picker-item-icon')).toBeNull();
	});

	it('reports the encounter index of the chosen target, and ALL_UNITS for all targets', async () => {
		const onTargetChange = vi.fn();
		resultData = resultWith({ index: 0 }, { index: 1 });
		mount(ALL_UNITS, onTargetChange);

		await open();
		await userEvent.click(items()[2]);
		expect(onTargetChange).toHaveBeenCalledWith(1);

		await open();
		await userEvent.click(items()[0]);
		expect(onTargetChange).toHaveBeenLastCalledWith(ALL_UNITS);
	});

	it('shows the target the pane has selected', async () => {
		resultData = resultWith({ index: 0 }, { index: 1 });
		mount(1);
		await open();

		expect(items().map(item => item.getAttribute('aria-selected'))).toEqual(['false', 'false', 'true']);
	});
});

describe('the filter the pane emits', () => {
	it('is a null target for all units and the index otherwise', () => {
		expect(simResultFilter(ALL_UNITS)).toEqual({ target: null });
		expect(simResultFilter(0)).toEqual({ target: 0 });
		expect(simResultFilter(3)).toEqual({ target: 3 });
	});

	it('knows a target the run no longer has', () => {
		const { result } = resultWith({ index: 0 }, { index: 1 });

		expect(hasTarget(result, ALL_UNITS)).toBe(true);
		expect(hasTarget(result, 1)).toBe(true);
		expect(hasTarget(result, 2)).toBe(false);
	});
});

describe('healing target options', () => {
	it('labels the healer separately and numbers targets that received healing or shielding', () => {
		const result = healingResultWith([
			{ unitIndex: 0, healing: 20, shielding: 0 },
			{ unitIndex: 2, healing: 50, shielding: 0 },
			{ unitIndex: 4, healing: 0, shielding: 35 },
			{ unitIndex: 6, healing: 0, shielding: 0 },
		]);

		expect(targetOptions(result).map(option => [option.value, option.label])).toEqual([
			[ALL_UNITS, 'results_tab.details.all_targets'],
			[0, 'results_tab.details.healer_target'],
			[2, 'results_tab.details.heal_target_number'],
			[4, 'results_tab.details.heal_target_number'],
		]);
	});

	it('recognizes healed dummies as valid targets after a run', () => {
		const result = healingResultWith([{ unitIndex: 3, healing: 25, shielding: 0 }]);

		expect(hasTarget(result, 3)).toBe(true);
		expect(hasTarget(result, 2)).toBe(false);
	});
});

describe('a target the options do not hold', () => {
	it('leaves the trigger on the placeholder', async () => {
		resultData = resultWith({ index: 0 });
		mount(4);

		expect(trigger().textContent).toBe('results_tab.details.unit');
		await open();
		expect(items().every(item => item.getAttribute('aria-selected') === 'false')).toBe(true);
	});
});
