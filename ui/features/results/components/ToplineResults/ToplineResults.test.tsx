import { PlayerSpecs } from '@sim/player/specs/index';
import { renderWithHost } from '@testing/react';
import { mockSubscriptions } from '@testing/sim/subscriptions';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimResultData } from '../../model/result_data';
import { ToplineResults } from './ToplineResults';

vi.mock('@sim/state/subscriptions', async () => mockSubscriptions());

let result: SimResultData | null = null;

vi.mock('../../hooks/useSimResult', () => ({ useSimResult: () => result }));

const dist = (avg: number, stdev = 0) => ({ avg, stdev });

const resultFor = (spec: unknown) =>
	({
		result: {
			getRaidIndexedPlayers: () => [
				{
					spec,
					unitIndex: 0,
					getTargetIndex: () => null,
					dps: dist(1000, 10),
					tps: dist(500),
					dtps: dist(50),
					tmi: dist(2),
					hps: dist(20),
					tto: dist(30),
					chanceOfDeath: dist(1),
					secondsOomAvg: 4,
					actionsPerMinute: () => 42,
				},
			],
			getRaidIndexedActionMetrics: () => [],
			getTargets: () => [],
			result: { avgIterationDuration: 300 },
			encounterMetrics: { durationSeconds: 300 },
		},
		filter: {},
	}) as never;

beforeEach(() => {
	result = null;
});

const mount = () =>
	renderWithHost(<ToplineResults />, { sim: { getShowDamageMetrics: () => true, getShowThreatMetrics: () => true, getShowHealingMetrics: () => true } });

describe('ToplineResults', () => {
	it('renders an empty root before the first run', () => {
		const { container } = mount();
		const root = container.querySelector('[data-testid="topline-results-root"]')!;

		expect(root).toBeTruthy();
		expect(root.children).toHaveLength(0);
	});

	it('renders the one-row table once a result arrives', () => {
		result = resultFor(PlayerSpecs.Mage);
		const { container } = mount();

		expect([...container.querySelectorAll('th')].map(cell => cell.getAttribute('data-metric'))).toEqual([
			'dps',
			'tps',
			'apm',
			'dtps',
			'tmi',
			'cod',
			'tto',
			'hps',
			'oom',
		]);
	});

	it('drops the out-of-mana column for a class with no mana bar', () => {
		result = resultFor(PlayerSpecs.DpsWarrior);
		const { container } = mount();

		expect([...container.querySelectorAll('th')].map(cell => cell.getAttribute('data-metric'))).not.toContain('oom');
	});
});
