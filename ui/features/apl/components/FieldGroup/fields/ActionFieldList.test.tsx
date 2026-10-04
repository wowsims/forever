import { APLAction, APLRotation } from '@generated/proto/apl';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { dragEvent, place } from '@testing/apl/dom';
import { makePlayer } from '@testing/apl/player';
import { source } from '@testing/apl/source';
import { act, fireEvent, render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ActionFieldList } from './ActionFieldList';

vi.mock('@sim/state/subscriptions', async () => {
	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
	return mockSubscriptions((await import('@testing/apl/source')).source.subscribe);
});
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

let actions: Array<APLAction>;
let player: ReturnType<typeof makePlayer>;

const setup = () => {
	source.listeners.clear();
	actions = [APLAction.create({ action: { oneofKind: 'resetSequence', resetSequence: { sequenceName: 'seq' } } })];
	player = makePlayer(APLRotation.create(), () => source.notify());
};

const fieldConfig = () => ({
	id: 'actions',
	storeSubscribe: () => source.subscribe,
	getValue: () => actions,
	setValue: (_subject: unknown, next: Array<APLAction>) => {
		actions = next;
		source.notify();
	},
});

const mount = () =>
	render(
		<SimHostProvider host={{ player, rootElem: document.body } as never}>
			<ActionFieldList player={player as never} config={fieldConfig() as never} />
		</SimHostProvider>,
	);

const items = () => Array.from(document.querySelectorAll<HTMLElement>('[data-testid="apl-action-list-item"]'));
const sequenceNames = () => actions.map(action => (action.action.oneofKind === 'resetSequence' ? action.action.resetSequence.sequenceName : ''));

const clickItemAction = (index: number, testId: string) => {
	const trigger = items()[index].querySelector<HTMLButtonElement>(':scope > [data-testid="tree-item-header"] > [data-testid="tree-item-actions"]')!;
	trigger.focus();
	act(() => void fireEvent.click(trigger));
	const action = document.querySelector(`[data-testid="tree-item-actions-popup"][data-open] [data-testid="${testId}"]`)!;
	act(() => void fireEvent.click(action));
};

beforeEach(() => {
	document.body.innerHTML = '';
});

describe('ActionFieldList', () => {
	it('lays each action out as an inline row holding its action picker', () => {
		setup();
		mount();

		const list = document.querySelector('[data-testid="apl-action-list"]')!;
		expect(list.getAttribute('data-layout')).toBe('inline');
		expect(list.querySelectorAll('[data-testid="apl-action-list-item"] [data-testid="apl-action-picker-root"]')).toHaveLength(1);
	});

	it('copies a row, which a sequence needs as much as the priority list does', () => {
		setup();
		mount();

		clickItemAction(0, 'apl-action-list-copy');

		expect(actions).toHaveLength(2);
		// A clone, not the same message: editing the copy must not edit the row it came from.
		expect(actions[1]).not.toBe(actions[0]);
		expect(actions[1]).toEqual(actions[0]);
		expect(items()).toHaveLength(2);
	});

	it('deletes a row through its menu and adds an empty one through the add button', () => {
		setup();
		mount();

		clickItemAction(0, 'apl-action-list-delete');
		expect(actions).toEqual([]);
		expect(items()).toHaveLength(0);

		act(() => void fireEvent.click(document.querySelector('[data-testid="tree-add"]')!));
		expect(actions).toEqual([APLAction.create()]);
		expect(items()).toHaveLength(1);
	});

	it('moves a row by dragging its handle, in one write', () => {
		setup();
		actions = ['a', 'b', 'c'].map(sequenceName => APLAction.create({ action: { oneofKind: 'resetSequence', resetSequence: { sequenceName } } }));
		const config = fieldConfig();
		const setValue = vi.spyOn(config, 'setValue');
		render(
			<SimHostProvider host={{ player, rootElem: document.body } as never}>
				<ActionFieldList player={player as never} config={config as never} />
			</SimHostProvider>,
		);
		place(items());

		fireEvent.mouseDown(items()[0].querySelector('[data-testid="apl-action-list-handle"]')!);
		fireEvent.dragStart(items()[0], { dataTransfer: new DataTransfer() });
		dragEvent('drop', items()[2], 55);

		expect(setValue).toHaveBeenCalledTimes(1);
		expect(sequenceNames()).toEqual(['b', 'c', 'a']);
	});
});
