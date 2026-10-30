import { ActionID } from '@generated/proto/common';
import { IconData } from '@generated/proto/ui';
import { Database } from '@sim/proto/database';
import { ActionMetrics, SimResult } from '@sim/proto/sim_result';
import {
	actionMetric as metric,
	metricRows as rows,
	metricsTooltipRowNames,
	metricsTooltipTable as tooltipTable,
	mockActionMetricsMerge,
	playerMetricsResult,
	spellActionsRun,
} from '@testing/results-metrics';
import { fireEvent, render, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimResultData } from '../../model/result_data';
import { HealingMetricsTable } from './HealingMetricsTable';

let result: SimResultData | null = null;
let showThreatMetrics = true;

vi.mock('../../hooks/useSimResult', () => ({ useSimResult: () => result }));
vi.mock('../MetricsTable/MetricsActionCell', () => ({
	MetricsActionCell: ({
		name,
		actionId,
		expandable,
		onToggle,
	}: {
		name: string;
		actionId: { iconUrl?: string };
		expandable: boolean;
		onToggle: () => void;
	}) =>
		expandable ? (
			<button data-icon={actionId.iconUrl} onClick={onToggle}>
				{name}
			</button>
		) : (
			<span data-icon={actionId.iconUrl}>{name}</span>
		),
}));
vi.mock('@sim/context/SimHostContext', async importOriginal => ({
	...(await importOriginal<typeof import('@sim/context/SimHostContext')>()),
	// `useDisplayMetrics` reads all three flags through one subscription; only threat varies per case here.
	useSim: () => ({ getShowThreatMetrics: () => showThreatMetrics, getShowDamageMetrics: () => true, getShowHealingMetrics: () => true }),
}));
vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions());

const playerResult = playerMetricsResult('getHealingActions');

const taggedActionId = (spellId: number, tag: number, name: string) => ({
	spellId,
	itemId: 0,
	tag,
	name,
	toStringIgnoringTag: () => `${spellId}`,
	toString: () => `${spellId}:${tag}`,
	equals: (other: { spellId: number; itemId: number; tag: number }) => other.spellId === spellId && other.itemId === 0 && other.tag === tag,
	withoutTag() {
		return { ...this, tag: 0 };
	},
});

const sourceId = (spellId: number) => ({ spellId, itemId: 0, tag: 0 });

// The actions on one player unit, whose action list a proc row reads its source's name from.
const onePlayer = (actions: Array<ActionMetrics>) => {
	const unit = { isPet: false, petActionId: null, actions };
	actions.forEach(action => Object.assign(action, { unit }));
	return actions;
};

describe('HealingMetricsTable', () => {
	beforeEach(() => {
		result = null;
		showThreatMetrics = true;
		mockActionMetricsMerge('healing', 'hps');
	});

	it('builds the twelve-column shell before any result', () => {
		const { container } = render(<HealingMetricsTable />);

		expect(container.querySelector('[data-testid="healing-metrics-root"]')).toBeTruthy();
		expect(container.querySelectorAll('thead th')).toHaveLength(12);
		expect(rows(container)).toHaveLength(0);
	});

	it('drops an action with no healing and opens sorted by HPS descending', () => {
		result = playerResult([metric('Healing Touch', { hps: 4 }), metric('Rejuvenation', { hps: 12 }), metric('Melee', { hps: 0 })]);
		const { container } = render(<HealingMetricsTable />);

		expect(rows(container).map(row => row.cells[0].textContent)).toEqual(['Rejuvenation', 'Healing Touch']);
	});

	it('keeps a row per source under the proc spell, named after its source, and leaves Penance tag 1 a normal action', () => {
		result = playerResult(
			onePlayer([
				metric('Prayer of Healing', { actionId: taggedActionId(100, 0, 'Prayer of Healing'), hps: 8 }),
				metric('Flash Heal', { actionId: taggedActionId(300, 0, 'Flash Heal'), hps: 2 }),
				metric('Divine Aegis', { actionId: taggedActionId(900, 100, 'Divine Aegis'), sourceActionId: sourceId(100), hps: 4 }),
				metric('Divine Aegis', { actionId: taggedActionId(900, 300, 'Divine Aegis'), sourceActionId: sourceId(300), hps: 3 }),
				metric('Penance', { actionId: taggedActionId(200, 1, 'Penance'), hps: 6 }),
			]),
		);
		const { container } = render(<HealingMetricsTable />);

		expect(rows(container).map(row => row.cells[0].textContent)).toEqual([
			'Prayer of Healing',
			'Divine Aegis',
			'Prayer of Healing',
			'Flash Heal',
			'Penance',
			'Flash Heal',
		]);
	});

	it('names the joined row of real proc metrics after the proc spell, with its icon, and each source row after its source', async () => {
		vi.mocked(ActionMetrics.merge).mockRestore();
		const spells: Record<number, { name: string; icon: string; rank?: number }> = {
			2060: { name: 'Heal', icon: 'spell_holy_heal', rank: 3 },
			2061: { name: 'Flash Heal', icon: 'spell_holy_flashheal', rank: 4 },
			431624: { name: 'Divine Aegis', icon: 'spell_holy_devineaegis' },
		};
		vi.spyOn(Database, 'getSpellIconData').mockImplementation(spellId => Promise.resolve(IconData.create({ id: spellId, ...spells[spellId] })));
		const run = spellActionsRun([
			{ spellId: 2060, damage: 0 },
			{ spellId: 2061, damage: 0 },
			{ spellId: 431624, tag: 2060, damage: 0 },
			{ spellId: 431624, tag: 2061, damage: 0 },
		]);
		const actions = run.result.raidMetrics!.parties[0].players[0].actions;
		[400, 300, 250, 100].forEach((healing, index) => (actions[index].targets[0].healing = healing));
		actions[2].sourceId = ActionID.create({ rawId: { oneofKind: 'spellId', spellId: 2060 } });
		actions[3].sourceId = ActionID.create({ rawId: { oneofKind: 'spellId', spellId: 2061 } });
		result = { result: await SimResult.makeNew(run.request, run.result), filter: {} } as unknown as SimResultData;
		const { container } = render(<HealingMetricsTable />);

		const names = rows(container).map(row => [
			row.cells[0].textContent,
			row.cells[0]
				.querySelector('[data-icon]')
				?.getAttribute('data-icon')
				?.match(/(\w+)\.jpg$/)?.[1],
		]);
		expect(names).toEqual([
			['Heal (common.action_tags.rank)', 'spell_holy_heal'],
			['Divine Aegis', 'spell_holy_devineaegis'],
			['Heal (common.action_tags.rank)', 'spell_holy_heal'],
			['Flash Heal (common.action_tags.rank)', 'spell_holy_flashheal'],
			['Flash Heal (common.action_tags.rank)', 'spell_holy_flashheal'],
		]);
	});

	it('keeps a lone source under the proc spell, named after a damage spell that heals nothing itself', () => {
		result = playerResult(
			onePlayer([
				metric('Shadow Word: Pain', { actionId: taggedActionId(589, 0, 'Shadow Word: Pain'), hps: 0 }),
				metric('Vampiric Embrace', { actionId: taggedActionId(15290, 589, 'Vampiric Embrace'), sourceActionId: sourceId(589), hps: 5 }),
			]),
		);
		const { container } = render(<HealingMetricsTable />);

		expect(rows(container).map(row => row.cells[0].textContent)).toEqual(['Vampiric Embrace', 'Shadow Word: Pain']);
	});

	it('draws the shielding overlay as a second bar fill', () => {
		result = playerResult([metric('Healing Touch')]);
		const { container } = render(<HealingMetricsTable />);

		expect(container.querySelectorAll('td[data-primary-metric] [data-testid="metrics-total-bar-fill"]')).toHaveLength(2);
	});

	it('drops the overlay bar when nothing is shielded', () => {
		result = playerResult([metric('Healing Touch', { shielding: 0 })]);
		const { container } = render(<HealingMetricsTable />);

		expect(container.querySelectorAll('td[data-primary-metric] [data-testid="metrics-total-bar-fill"]')).toHaveLength(1);
	});

	it('opens the healing breakdown with its average, min and max columns from the primary-metric cell', async () => {
		result = playerResult([metric('Healing Touch')]);
		const { container } = render(<HealingMetricsTable />);

		fireEvent.mouseEnter(container.querySelector('td[data-primary-metric]')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect([...tooltipTable()!.querySelectorAll('thead th')].map(th => th.textContent)).toEqual([
			'results_tab.details.tooltip_table.type',
			'results_tab.details.tooltip_table.amount',
			'results_tab.details.tooltip_table.min',
			'results_tab.details.tooltip_table.average',
			'results_tab.details.tooltip_table.max',
		]);
		expect(metricsTooltipRowNames()).toEqual(['results_tab.details.attack_types.hit', 'results_tab.details.attack_types.critical_hit']);
	});

	it('adds a spec metrics group to the healing breakdown, with the share the action gives each value', async () => {
		result = playerResult([
			metric('Prayer of Healing', { customMetrics: [{ key: 'divine_aegis', value: 30, percentage: 30, average: 410, min: 300, max: 520 }] }),
		]);
		const { container } = render(<HealingMetricsTable />);

		fireEvent.mouseEnter(container.querySelector('td[data-primary-metric]')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(metricsTooltipRowNames()).toEqual([
			'results_tab.details.attack_types.hit',
			'results_tab.details.attack_types.critical_hit',
			'results_tab.details.custom_metrics.group',
			'divine_aegis',
		]);
		const aegis = [...tooltipTable()!.querySelectorAll<HTMLTableRowElement>('tbody tr')].at(-1)!;
		expect(aegis.cells[1].firstElementChild?.firstElementChild?.textContent).toMatch(/^30(\.0+)?\s?%$/);
	});

	it('keeps a threat tooltip shut while threat metrics are hidden, and opens it once they are shown', async () => {
		result = playerResult([metric('Healing Touch')]);
		showThreatMetrics = false;
		const hidden = render(<HealingMetricsTable />);

		fireEvent.mouseEnter(hidden.container.querySelector('td.text-success')!);
		await waitFor(() => expect(hidden.container.querySelector('td.text-success')?.getAttribute('data-tooltip-id')).toBeTruthy());
		expect(tooltipTable()).toBeNull();
		hidden.unmount();

		showThreatMetrics = true;
		const shown = render(<HealingMetricsTable />);
		fireEvent.mouseEnter(shown.container.querySelector('td.text-success')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(tooltipTable()!.textContent).toContain('results_tab.details.attack_types.threat');
	});

	it('names the two groups of the hits breakdown', async () => {
		result = playerResult([metric('Healing Touch', { landedTicks: 4, ticks: 4 })]);
		const { container } = render(<HealingMetricsTable />);

		fireEvent.mouseEnter(rows(container)[0].cells[6]);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect([...tooltipTable()!.querySelectorAll('tbody th')].map(th => th.textContent)).toEqual([
			'results_tab.details.columns.hits',
			'results_tab.details.tooltip_table.ticks',
		]);
	});

	it('anchors all three header tooltips on their own header cell', () => {
		const { container } = render(<HealingMetricsTable />);
		const headers = [...container.querySelectorAll('thead th')];

		expect(headers[5].getAttribute('data-tooltip-id')).toBe('healing-metrics-avg-cast-header');
		expect(headers[6].getAttribute('data-tooltip-id')).toBe('healing-metrics-hits-header');
		expect(headers[7].getAttribute('data-tooltip-id')).toBe('healing-metrics-avg-hit-header');
		expect(headers.map(th => th.getAttribute('data-tooltip-content')).filter(Boolean)).toHaveLength(3);
	});

	it('leaves no tooltip behind when the table unmounts while one is open', async () => {
		result = playerResult([metric('Healing Touch')]);
		const { container, unmount } = render(<HealingMetricsTable />);

		fireEvent.mouseEnter(container.querySelector('td[data-primary-metric]')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());

		unmount();
		expect(document.querySelectorAll('.sim-tooltip')).toHaveLength(0);
	});

	it('renders no bar denominator at all before a result, rather than -Infinity', () => {
		const { container } = render(<HealingMetricsTable />);
		expect(container.querySelectorAll('[data-testid="metrics-total-bar-fill"]')).toHaveLength(0);
	});
});
