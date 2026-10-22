import { Emitter } from '@sim/state/events';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { metricRows, metricsTooltipTable, RANKED_SPELLS, spellActionsRun, stubRankedSpellIcons } from '@testing/results-metrics';
import { act, fireEvent, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ResultChannel } from '../../model/result_channel';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);
vi.mock('../MetricsTable/MetricsActionCell', () => ({ MetricsActionCell: ({ name }: { name: string }) => <span>{name}</span> }));
vi.mock('../ResultsFilter', async importOriginal => ({
	...(await importOriginal<typeof import('../ResultsFilter')>()),
	ResultsFilter: () => <span data-testid="results-filter-root" />,
}));
vi.mock('../LogRunner', () => ({ LogRunner: () => <div /> }));
vi.mock('../Timeline', () => ({ Timeline: () => <div /> }));
vi.mock('../ToplineResults', () => ({ ToplineResults: () => <div /> }));
vi.mock('../DpsHistogram', () => ({ DpsHistogram: () => <div /> }));
vi.mock('../CastMetricsTable', () => ({ CastMetricsTable: () => <div /> }));
vi.mock('../ResourceMetricsTable', () => ({ ResourceMetricsTable: () => <div /> }));
vi.mock('../AuraMetricsTable', () => ({ AuraMetricsTable: () => <div /> }));

import { DetailedResults } from './DetailedResults';

const { FIREBALL_1, FIREBALL_2, FROSTBOLT } = RANKED_SPELLS;

const RANK = 'common.action_tags.rank';

let resultChannel: ResultChannel;
let currentChangeEmitter: Emitter<void>;

const renderPane = async () => {
	const run = spellActionsRun([
		{ spellId: FIREBALL_1, damage: 100 },
		{ spellId: FIREBALL_1, tag: 1, damage: 10 },
		{ spellId: FIREBALL_2, damage: 300 },
		{ spellId: FIREBALL_2, tag: 1, damage: 30 },
		{ spellId: FROSTBOLT, damage: 50 },
	]);
	const view = renderWithHost(<DetailedResults resultsManager={{ currentChangeEmitter, getRunData: () => ({ run }) } as never} />, {
		player: { getPlayerSpec: () => ({ isHealingSpec: false }) },
		sim: {
			store: createSimStore(),
			getShowDamageMetrics: () => true,
			getShowThreatMetrics: () => false,
			getShowHealingMetrics: () => false,
			getShowExperimental: () => false,
			getFixedRngSeed: () => 0,
			setFixedRngSeed: () => {},
		},
		resultChannel,
		disabled: false,
		runSingleIteration: () => Promise.resolve(undefined),
	});
	await act(async () => currentChangeEmitter.emit());
	return view;
};

const damageRows = (container: HTMLElement) =>
	metricRows(container.querySelector<HTMLElement>('[data-testid="damage-metrics-root"]')!).map(row => [
		row.cells[0].textContent,
		row.hasAttribute('data-parent'),
		row.hasAttribute('data-child'),
	]);

const RANKED = `Fireball (${RANK})`;
// Fireball names its tag 1 as the damage over time part.
const RANKED_DOT = `Fireball (common.action_tags.dot) (${RANK})`;

const toggle = (container: HTMLElement) => within(container.querySelector<HTMLElement>('[data-testid="results-filter"]')!).getByRole('checkbox');

beforeEach(() => {
	window.localStorage.clear();
	resultChannel = new ResultChannel();
	currentChangeEmitter = new Emitter<void>();
	stubRankedSpellIcons();
});

describe('DetailedResults rank grouping', () => {
	it('opens with the ranks of a spell under one row named without a rank and the tags of a rank under that rank, and lists every rank as its own row once the toggle is cleared', async () => {
		const { container } = await renderPane();

		expect(toggle(container).getAttribute('aria-checked')).toBe('true');
		await waitFor(() =>
			expect(damageRows(container)).toEqual([
				['Fireball', true, false],
				[RANKED, true, true],
				[RANKED, false, true],
				[RANKED_DOT, false, true],
				[RANKED, true, true],
				[RANKED, false, true],
				[RANKED_DOT, false, true],
				['Frostbolt', false, false],
			]),
		);

		await act(async () => {
			fireEvent.click(toggle(container));
		});

		expect(damageRows(container)).toEqual([
			[RANKED, true, false],
			[RANKED, false, true],
			[RANKED_DOT, false, true],
			[RANKED, true, false],
			[RANKED, false, true],
			[RANKED_DOT, false, true],
			['Frostbolt', false, false],
		]);
	});

	it('opens the damage breakdown on a tag row two levels down', async () => {
		const { container } = await renderPane();
		await waitFor(() => expect(damageRows(container)).toHaveLength(8));

		const tagCell = metricRows(container.querySelector<HTMLElement>('[data-testid="damage-metrics-root"]')!)[2].querySelector('td[data-primary-metric]')!;
		expect(tagCell.getAttribute('data-row-id')).toBe('0.1.0');
		fireEvent.mouseEnter(tagCell);
		await waitFor(() => expect(metricsTooltipTable()).toBeTruthy());
	});

	it('remembers a cleared toggle for the next time the pane opens', async () => {
		const first = await renderPane();
		await waitFor(() => expect(damageRows(first.container)).toHaveLength(8));
		await act(async () => {
			fireEvent.click(toggle(first.container));
		});
		first.unmount();

		const { container } = await renderPane();
		expect(toggle(container).getAttribute('aria-checked')).toBe('false');
		await waitFor(() => expect(damageRows(container)).toHaveLength(7));
		expect(damageRows(container)[0][0]).toBe(RANKED);
	});
});
