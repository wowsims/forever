import { APLRotation, APLValue } from '@generated/proto/apl';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { dragEvent, place } from '@testing/apl/dom';
import { makePlayer } from '@testing/apl/player';
import { source } from '@testing/apl/source';
import { act, fireEvent, render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ValueFieldList } from './ValueFieldList';

vi.mock('@sim/state/subscriptions', async () => {
	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
	return mockSubscriptions((await import('@testing/apl/source')).source.subscribe);
});
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const constant = (uuid: string, val: string) => APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'const', const: { val } } });

let values: Array<APLValue | undefined>;
let player: ReturnType<typeof makePlayer>;

const fieldConfig = () => ({
	id: 'vals',
	storeSubscribe: () => source.subscribe,
	getValue: () => values,
	setValue: (_subject: unknown, next: Array<APLValue | undefined>) => {
		values = next;
		source.notify();
	},
});

const mount = (initial: Array<APLValue | undefined>, config = fieldConfig()) => {
	source.listeners.clear();
	values = initial;
	player = makePlayer(APLRotation.create(), () => source.notify());
	return render(
		<SimHostProvider host={{ player, rootElem: document.body } as never}>
			<ValueFieldList player={player as never} config={config as never} />
		</SimHostProvider>,
	);
};

const list = () => document.querySelector('[data-testid="apl-value-list"]') as HTMLElement;
const items = () => Array.from(document.querySelectorAll<HTMLElement>('[data-testid="apl-value-list-item"]'));
const uuids = () => values.map(value => value?.uuid?.value);

const openMenu = (index: number) => {
	const trigger = items()[index].querySelector<HTMLButtonElement>(':scope > [data-testid="tree-item-header"] > [data-testid="tree-item-actions"]')!;
	trigger.focus();
	act(() => void fireEvent.click(trigger));
	return document.querySelector('[data-testid="tree-item-actions-popup"][data-open]') as HTMLElement;
};

const clickItemAction = (index: number, testId: string) => {
	const action = openMenu(index).querySelector(`[data-testid="${testId}"]`)!;
	act(() => void fireEvent.click(action));
};

beforeEach(() => {
	document.body.innerHTML = '';
});

describe('ValueFieldList', () => {
	it('renders one inline value row per operand', () => {
		mount([constant('a', '1'), constant('b', '2')]);

		expect(items()).toHaveLength(2);
		expect(list().getAttribute('data-layout')).toBe('inline');
		expect(items().map(item => item.querySelector('[data-testid="apl-value-picker-root"]'))).not.toContain(null);
	});

	it('copies a row to a clone beside it, with a uuid of its own', () => {
		mount([constant('a', '1'), constant('b', '2')]);

		clickItemAction(0, 'apl-value-list-copy');

		const [copied, ...rest] = uuids();
		expect(rest).toEqual(['a', 'b']);
		expect(copied).toBeTruthy();
		expect(copied).not.toBe('a');
		expect(values[0]).not.toBe(values[1]);
		expect(values[0]!.value).toEqual(values[1]!.value);
		expect(items()).toHaveLength(3);
	});

	it('deletes a row through its menu', () => {
		mount([constant('a', '1'), constant('b', '2')]);

		clickItemAction(1, 'apl-value-list-delete');

		expect(uuids()).toEqual(['a']);
		expect(items()).toHaveLength(1);
	});

	it('adds an empty value that carries a uuid', () => {
		mount([]);

		act(() => void fireEvent.click(document.querySelector('[data-testid="tree-add"]')!));

		expect(values).toHaveLength(1);
		expect(values[0]!.uuid!.value).toBeTruthy();
		expect(values[0]!.value.oneofKind).toBeUndefined();
	});

	it('stores a hole left in the list as an empty value with a uuid', () => {
		mount([constant('a', '1'), undefined]);

		clickItemAction(0, 'apl-value-list-delete');

		expect(values).toHaveLength(1);
		expect(values[0]).toBeDefined();
		expect(values[0]!.uuid!.value).toBeTruthy();
	});

	it.each([
		['offers to extract a value that has a kind', 0, true],
		['withholds the extract entry from an empty value', 1, false],
	])('%s', (_case, index, offered) => {
		mount([constant('a', '1'), APLValue.create({ uuid: { value: 'b' } })]);

		expect(openMenu(index).querySelector('[data-testid="apl-value-list-extract"]') !== null).toBe(offered);
	});

	it('moves a row by dragging its handle, in one write', () => {
		const config = fieldConfig();
		const setValue = vi.spyOn(config, 'setValue');
		mount([constant('a', '1'), constant('b', '2'), constant('c', '3')], config);
		place(items());

		fireEvent.mouseDown(items()[2].querySelector('[data-testid="apl-value-list-handle"]')!);
		fireEvent.dragStart(items()[2], { dataTransfer: new DataTransfer() });
		dragEvent('drop', items()[0], 0);

		expect(setValue).toHaveBeenCalledTimes(1);
		expect(uuids()).toEqual(['c', 'a', 'b']);
	});
});
