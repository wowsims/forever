import type { AplRun } from '@features/apl/model/apl_entry';
import { APLAction, APLRotation } from '@generated/proto/apl';
import i18n from '@i18n/config';
import type { AplEntryPath } from '@sim/proto/combat_log';
import { constant, sequence, waitFor } from '@testing/apl/builders';
import { makePlayer } from '@testing/apl/player';
import { renderWithHost } from '@testing/react';
import { castLog } from '@testing/storybook/fixtures/results-views';
import { act } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../../assets/locales/en/translation.json';
import { CastTooltip, type CastTooltipProps } from './CastTooltip';

vi.mock('@sim/state/subscriptions', async () => {
	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
	return mockSubscriptions((await import('@testing/apl/source')).source.subscribe);
});

const rotation = APLRotation.create({
	prepullActions: [{ action: waitFor('1s'), doAtValue: constant('-1s') }],
	priorityList: [
		...Array.from({ length: 5 }, () => ({ action: waitFor('9s') })),
		{ action: APLAction.create({ action: { oneofKind: 'groupReference', groupReference: { groupName: 'execute', variables: [] } } }) },
		{ action: sequence([waitFor('3s'), waitFor('4s'), waitFor('5s')]) },
	],
	groups: [{ name: 'execute', actions: [{ action: waitFor('6s') }, { action: waitFor('7s') }], variables: [] }],
});

const ran = (changed = false): AplRun => ({ rotation, changedSinceRun: () => changed });

const IN_GROUP: AplEntryPath = [
	{ kind: 'priority', index: 5 },
	{ kind: 'group', group: 0, index: 1 },
];

const render = async (props: CastTooltipProps) => {
	const view = renderWithHost(<CastTooltip {...props} />, {
		player: makePlayer(APLRotation.clone(rotation)),
		sim: { getShowDamageMetrics: () => true, getShowThreatMetrics: () => true, getShowHealingMetrics: () => true },
	});
	await act(async () => {});
	return view.container;
};

const entryInputs = (container: HTMLElement) =>
	Array.from(container.querySelectorAll('[data-testid="apl-entry-view"] input')).map(input => (input as HTMLInputElement).value);

// The test setup loads no locale, and the copy is what these cases pin, so the English bundle is loaded here.
beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

describe('CastTooltip', () => {
	it('shows the path, the entry and where a click takes it for a cast of an APL run', async () => {
		const container = await render({ log: castLog(12, { aplEntryPath: IN_GROUP }), aplRun: ran() });

		expect(container.textContent).toContain('Priority list #6 › Group execute › #2');
		expect(entryInputs(container)).toEqual(['7s']);
		expect(container.textContent).toContain('Click to open this line in Rotation');
	});

	it('draws the entry as it ran, and says the rotation changed instead of offering the click', async () => {
		const container = await render({ log: castLog(12, { aplEntryPath: IN_GROUP }), aplRun: ran(true) });

		expect(entryInputs(container)).toEqual(['7s']);
		expect(container.textContent).toContain('The rotation has changed since this run');
		expect(container.textContent).not.toContain('Click to open');
	});

	it('names a sequence step and a prepull row by their positions from 1', async () => {
		const step = await render({
			log: castLog(12, {
				aplEntryPath: [
					{ kind: 'priority', index: 6 },
					{ kind: 'step', index: 2 },
				],
			}),
			aplRun: ran(),
		});
		expect(step.textContent).toContain('Priority list #7 › Step 3');

		const prepull = await render({ log: castLog(-1, { aplEntryPath: [{ kind: 'prepull', index: 0 }] }), aplRun: ran() });
		expect(prepull.textContent).toContain('Prepull #1');
		expect(entryInputs(prepull)).toEqual(['-1s', '1s']);
	});

	it('leaves the tooltip as it is with no run rotation, for a cast no entry launched (every cast of an Auto or Simple run), and for an entry the run cannot place', async () => {
		const today = (await render({ log: castLog(12) })).innerHTML;

		expect((await render({ log: castLog(12, { aplEntryPath: IN_GROUP }) })).innerHTML).toBe(today);
		expect((await render({ log: castLog(12), aplRun: ran() })).innerHTML).toBe(today);
		expect((await render({ log: castLog(12, { aplEntryPath: [{ kind: 'priority', index: 40 }] }), aplRun: ran() })).innerHTML).toBe(today);
	});
});
