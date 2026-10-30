import type { AplRun } from '@features/apl/model/apl_entry';
import { RaidSimRequest } from '@generated/proto/api';
import { APLRotation, APLRotation_Type } from '@generated/proto/apl';
import type { AplTestPlayer } from '@testing/apl/player';
import { renderWithHost } from '@testing/react';
import { makeAplStoryPlayer } from '@testing/storybook/fixtures/apl';
import { fireEvent, within } from '@testing-library/react';
import { act } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimResultData } from '../../model/result_data';
import { Timeline } from './Timeline';

const state = vi.hoisted(() => ({
	result: null as SimResultData | null,
	models: [] as Array<unknown>,
	aplRuns: [] as Array<unknown>,
	specs: [] as Array<unknown>,
	built: 0,
	charted: 0,
}));

vi.mock('../../hooks/useSimResult', () => ({ useSimResult: () => state.result }));
vi.mock('../../model/timeline/rotation', () => ({
	buildRotationModel: ({ player }: { player: { id: string } }) => {
		state.built++;
		return { id: player.id, duration: 10, rows: [], sections: [], byKey: new Map() };
	},
}));
vi.mock('./chart/build', () => ({
	chartSpec: (player: { id: string }) => {
		state.charted++;
		return { id: player.id, datasets: [], scales: {}, annotations: null, duration: 10 };
	},
}));
vi.mock('./rotation/RotationTrack', () => ({
	RotationTrack: ({ model, aplRun }: { model: unknown; aplRun: unknown }) => {
		state.models.push(model);
		state.aplRuns.push(aplRun);
		return <div data-testid="rotation-pane" />;
	},
}));
vi.mock('./chart/TimelineChart', () => ({
	TimelineChart: ({ spec }: { spec: unknown }) => {
		state.specs.push(spec);
		return <div data-testid="timeline-chart" />;
	},
}));

const resultFor = (requestId: string, playerId: string, filter: unknown = {}) =>
	({
		result: {
			request: { requestId },
			result: { firstIterationDuration: 10 },
			getRaidIndexedPlayers: () => [{ id: playerId }],
			getTargets: () => [{ index: 0 }],
		},
		filter,
	}) as unknown as SimResultData;

// A run of one player at raid index 0; the request holds its own copy of `rotation`, as a real request does.
const runOf = (requestId: string, rotation: APLRotation) =>
	({
		...resultFor(requestId, 'p1'),
		result: {
			...resultFor(requestId, 'p1').result,
			request: RaidSimRequest.create({ requestId, raid: { parties: [{ players: [{ rotation }] }] } }),
			getRaidIndexedPlayers: () => [{ id: 'p1', index: 0 }],
		},
	}) as unknown as SimResultData;

const viewButton = (container: HTMLElement, view: 'rotation' | 'dps') =>
	within(container).getByRole('button', { name: `results_tab.details.timeline.chart_types.${view}` });

const emit = (next: SimResultData | null, rerender: () => void) => {
	state.result = next;
	act(rerender);
};

beforeEach(() => {
	state.result = null;
	state.models.length = 0;
	state.aplRuns.length = 0;
	state.specs.length = 0;
	state.built = 0;
	state.charted = 0;
});

// The APL test player bumps the store's rotation counter on every write, as the real one does.
const livePlayer = (rotation = APLRotation.create({ type: APLRotation_Type.TypeAPL })): AplTestPlayer => {
	const player = makeAplStoryPlayer(rotation);
	player.getResolvedAplRotation = () => player.aplRotation;
	return player;
};

const mount = (active: boolean, player = livePlayer()) => {
	const view = renderWithHost(<Timeline active={active} />, { player });
	return { ...view, again: (next = active) => view.rerender(<Timeline active={next} />) };
};

const lastAplRun = () => state.aplRuns.at(-1) as AplRun | undefined;

describe('Timeline', () => {
	it('holds a run that lands while the tab is closed, and draws it when the tab opens', () => {
		const view = mount(false);
		emit(resultFor('a', 'p1'), view.again);
		expect(state.models.at(-1)).toBeNull();

		act(() => view.again(true));
		expect(state.models.at(-1)).toMatchObject({ id: 'p1' });
	});

	it('builds the rotation once for two emits of the same run under the same filter', () => {
		const view = mount(true);
		emit(resultFor('a', 'p1'), view.again);
		expect(state.built).toBe(1);

		emit(resultFor('a', 'p2'), view.again);
		expect(state.built).toBe(1);
		expect(state.models.at(-1)).toMatchObject({ id: 'p1' });
	});

	it('rebuilds when the filter changes, because the rows are the filtered units', () => {
		const view = mount(true);
		emit(resultFor('a', 'p1', { unit: 0 }), view.again);
		emit(resultFor('a', 'p2', { unit: 1 }), view.again);
		expect(state.built).toBe(2);
		expect(state.models.at(-1)).toMatchObject({ id: 'p2' });
	});

	it('draws nothing for a cleared result, whose raid is empty', () => {
		state.result = { ...resultFor('a', 'p1'), result: { ...resultFor('a', 'p1').result, getRaidIndexedPlayers: () => [] } } as unknown as SimResultData;
		mount(true);
		expect(state.models.at(-1)).toBeNull();
		expect(state.built).toBe(0);
	});

	it('survives a model that cannot be built', () => {
		const logged = vi.spyOn(console, 'log').mockImplementation(() => {});
		state.result = {
			...resultFor('a', 'p1'),
			result: {
				...resultFor('a', 'p1').result,
				getTargets: () => {
					throw new Error('no targets');
				},
			},
		} as unknown as SimResultData;
		mount(true);
		expect(state.models.at(-1)).toBeNull();
		expect(logged).toHaveBeenCalled();
	});

	it('leaves the chart series unbuilt until the DPS view is picked, then keeps them', () => {
		const view = mount(true);
		emit(resultFor('a', 'p1'), view.again);
		expect(state.specs).toHaveLength(0);
		expect(state.charted).toBe(0);

		act(() => {
			fireEvent.click(viewButton(view.container, 'dps'));
		});
		expect(state.charted).toBe(1);
		expect(state.specs.at(-1)).toMatchObject({ id: 'p1' });

		act(() => {
			fireEvent.click(viewButton(view.container, 'rotation'));
		});
		act(() => {
			fireEvent.click(viewButton(view.container, 'dps'));
		});
		expect(state.charted).toBe(1);
	});

	it('hands the track the rotation an APL run ran, unchanged right after the run', () => {
		const rotation = APLRotation.create({ type: APLRotation_Type.TypeAPL, priorityList: [{ notes: 'ran' }] });
		const view = mount(true, livePlayer(APLRotation.clone(rotation)));
		emit(runOf('a', rotation), view.again);

		expect(lastAplRun()?.rotation.priorityList[0].notes).toBe('ran');
		expect(lastAplRun()?.changedSinceRun()).toBe(false);
	});

	it('marks the run changed once the rotation is edited, keeping the rotation as it ran, without drawing the track again', () => {
		const rotation = APLRotation.create({ type: APLRotation_Type.TypeAPL, priorityList: [{ notes: 'ran' }] });
		const player = livePlayer(APLRotation.clone(rotation));
		const view = mount(true, player);
		emit(runOf('a', rotation), view.again);
		expect(lastAplRun()?.changedSinceRun()).toBe(false);
		const drawn = state.aplRuns.length;

		act(() => player.modifyAplRotation(live => live.priorityList.push({ notes: 'added' } as never)));
		expect(state.aplRuns).toHaveLength(drawn);
		expect(lastAplRun()?.changedSinceRun()).toBe(true);
		expect(lastAplRun()?.rotation.priorityList).toHaveLength(1);
	});

	it('mounts only the pane that is showing', () => {
		const view = mount(true);
		expect(view.container.querySelector('[data-testid="dps-resources-plot"]')).toBeNull();
		expect(view.container.querySelector('[data-testid="rotation-plot"]')).toBeTruthy();

		act(() => {
			fireEvent.click(viewButton(view.container, 'dps'));
		});
		expect(view.container.querySelector('[data-testid="dps-resources-plot"]')).toBeTruthy();
		expect(view.container.querySelector('[data-testid="rotation-plot"]')).toBeNull();
	});
});
