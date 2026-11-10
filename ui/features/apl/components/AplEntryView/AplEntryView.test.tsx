import { AplListItem } from '@features/apl/components/AplListItem';
import { resolveAplEntry } from '@features/apl/model/apl_entry';
import { APLListItem, APLRotation } from '@generated/proto/apl';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { AplEntryPath } from '@sim/proto/combat_log';
import { after, constant, sequence, waitFor } from '@testing/apl/builders';
import { makePlayer } from '@testing/apl/player';
import { source } from '@testing/apl/source';
import { act, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, type Mock, vi } from 'vitest';

import { AplEntryView } from './AplEntryView';

vi.mock('@sim/state/subscriptions', async () => {
	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
	return mockSubscriptions((await import('@testing/apl/source')).source.subscribe);
});
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

/** Cast a spell once the fight is past 10 and 20 seconds: a row with an And condition. */
const castWhen = () =>
	APLListItem.create({
		action: {
			condition: { value: { oneofKind: 'and', and: { vals: [after('10s'), after('20s')] } } },
			action: { oneofKind: 'castSpell', castSpell: {} },
		},
	});

let player: ReturnType<typeof makePlayer>;
let notify: Mock<() => void>;

const setup = (rotation: APLRotation) => {
	source.listeners.clear();
	notify = vi.fn(() => source.notify());
	player = makePlayer(rotation, notify);
};

const mount = (path: AplEntryPath) =>
	render(
		<SimHostProvider host={{ player, rootElem: document.body } as never}>
			<AplEntryView entry={resolveAplEntry(player.aplRotation, path)!} />
		</SimHostProvider>,
	);

const view = () => screen.getByTestId('apl-entry-view');
const rowText = (root: ParentNode) => root.querySelector('[data-testid="apl-list-item-picker-root"]')!.textContent;
const inputValues = (root: Element) => Array.from(root.querySelectorAll('input')).map(input => input.value);

beforeEach(() => {
	document.body.innerHTML = '';
});

describe('AplEntryView', () => {
	it('draws a Cast Spell row with an And condition as the editor row draws it', async () => {
		setup(APLRotation.create({ priorityList: [castWhen()] }));
		const editor = render(
			<SimHostProvider host={{ player, rootElem: document.body } as never}>
				<AplListItem player={player as never} config={{ getValue: () => player.aplRotation.priorityList[0], setValue: () => {} } as never} />
			</SimHostProvider>,
		);
		mount([{ kind: 'priority', index: 0 }]);
		await act(async () => {});

		expect(rowText(view())).toBe(rowText(editor.container));
		expect(view().querySelectorAll('[data-testid="apl-action-condition"]')).toHaveLength(1);
		expect(view().querySelector('[data-disabled]')).toBeNull();
	});

	// happy-dom honours `inert` for focus only and still delivers a click to the trigger inside, which a browser does not.
	it('keeps the whole row out of reach: inert, and its pickers take no focus', async () => {
		setup(APLRotation.create({ priorityList: [castWhen()] }));
		mount([{ kind: 'priority', index: 0 }]);
		await act(async () => {});

		expect(view().hasAttribute('inert')).toBe(true);
		const trigger = view().querySelector<HTMLElement>('button, [role="combobox"]')!;
		act(() => trigger.focus());
		expect(document.activeElement).not.toBe(trigger);
	});

	it('writes nothing to the rotation it draws', async () => {
		const rotation = APLRotation.create({ priorityList: [castWhen()] });
		setup(rotation);
		const before = APLRotation.clone(rotation);
		mount([{ kind: 'priority', index: 0 }]);
		await act(async () => {});

		expect(rotation).toEqual(before);
		expect(notify).not.toHaveBeenCalled();
	});

	it('marks the step that fired and no other', async () => {
		// The first step holds a sequence of its own, whose second step is not the one that fired.
		setup(APLRotation.create({ priorityList: [{ action: sequence([sequence([waitFor('1s'), waitFor('2s')]), waitFor('3s'), waitFor('4s')]) }] }));
		mount([
			{ kind: 'priority', index: 0 },
			{ kind: 'step', index: 1 },
		]);
		await act(async () => {});

		const marked = view().querySelectorAll('[data-marked]');
		expect(marked).toHaveLength(1);
		expect(inputValues(marked[0])).toEqual(['3s']);
	});

	it('marks a step of a nested sequence, not the step that holds it', async () => {
		setup(APLRotation.create({ priorityList: [{ action: sequence([waitFor('1s'), sequence([waitFor('2s'), waitFor('3s')])]) }] }));
		mount([
			{ kind: 'priority', index: 0 },
			{ kind: 'step', index: 1 },
			{ kind: 'step', index: 0 },
		]);
		await act(async () => {});

		const marked = view().querySelectorAll('[data-marked]');
		expect(marked).toHaveLength(1);
		expect(inputValues(marked[0])).toEqual(['2s']);
	});

	it('draws a pre-pull row with its Do At time', async () => {
		setup(APLRotation.create({ prepullActions: [{ action: waitFor('1s'), doAtValue: constant('-2s') }] }));
		mount([{ kind: 'prepull', index: 0 }]);
		await act(async () => {});

		expect(screen.getByText('rotation_tab.apl.prepull_actions.do_at.label')).toBeDefined();
		expect(inputValues(view())).toEqual(['-2s', '1s']);
	});
});
