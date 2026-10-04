import { ActionId } from '@sim/proto/action_id';
import { deferFill, filled } from '@testing/sim/hooks';
import { act, render } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { useActionIds } from './useActionIds';

const seen: Array<ReturnType<typeof useActionIds>> = [];
const Probe = ({ actionIds }: { actionIds: Array<ActionId | undefined> }) => {
	seen.push(useActionIds(actionIds));
	return null;
};

afterEach(() => {
	seen.length = 0;
	vi.restoreAllMocks();
});

describe('useActionIds', () => {
	it('resolves every id in one pass, and leaves an absent one undefined', async () => {
		const { spy, resolvers } = deferFill();
		const first = ActionId.fromSpellId(1);
		const second = ActionId.fromItemId(2);
		render(<Probe actionIds={[first, undefined, second]} />);
		expect(seen.at(-1)).toEqual([undefined, undefined, undefined]);
		expect(spy).toHaveBeenCalledTimes(2);

		await act(async () => {
			resolvers[0](filled(first, 'First', 'first.jpg'));
			resolvers[1](filled(second, 'Second', 'second.jpg'));
		});

		expect(seen.at(-1)).toEqual([
			{ iconUrl: 'first.jpg', name: 'First', href: ActionId.makeSpellUrl(1) },
			undefined,
			{ iconUrl: 'second.jpg', name: 'Second', href: ActionId.makeItemUrl(2) },
		]);
	});

	it('resolves ids that already carry an icon on the first render, without a fill', () => {
		const { spy } = deferFill();
		render(<Probe actionIds={[filled(ActionId.fromSpellId(1), 'Fireball', 'fireball.jpg'), ActionId.fromPetName('Wolf')]} />);

		expect(seen[0][0]).toEqual({ iconUrl: 'fireball.jpg', name: 'Fireball', href: ActionId.makeSpellUrl(1) });
		expect(seen[0][1]?.iconUrl).toContain('ability_hunter_pet_wolf');
		expect(seen[0][1]?.href).toBeUndefined();
		expect(spy).not.toHaveBeenCalled();
	});

	it('drops a resolution that lands after unmount', async () => {
		const { resolvers } = deferFill();
		const actionId = ActionId.fromSpellId(1);
		const { unmount } = render(<Probe actionIds={[actionId]} />);
		unmount();
		const renders = seen.length;

		await act(async () => resolvers[0](filled(actionId, 'Late', 'late.jpg')));

		expect(seen).toHaveLength(renders);
	});

	it('does not refetch for a fresh array of the same ids, and returns the same array', () => {
		const { spy } = deferFill();
		const { rerender } = render(<Probe actionIds={[ActionId.fromSpellId(1), ActionId.fromSpellId(2)]} />);

		rerender(<Probe actionIds={[ActionId.fromSpellId(1), ActionId.fromSpellId(2)]} />);

		expect(spy).toHaveBeenCalledTimes(2);
		expect(seen[1]).toBe(seen[0]);
	});

	it('lets the current ids win when an earlier fill resolves last', async () => {
		const { resolvers } = deferFill();
		const first = ActionId.fromSpellId(1);
		const second = ActionId.fromSpellId(2);
		const { rerender } = render(<Probe actionIds={[first]} />);

		rerender(<Probe actionIds={[second]} />);
		await act(async () => resolvers[1](filled(second, 'Second', 'second.jpg')));
		await act(async () => resolvers[0](filled(first, 'First', 'first.jpg')));

		expect(seen.at(-1)).toEqual([{ iconUrl: 'second.jpg', name: 'Second', href: ActionId.makeSpellUrl(2) }]);
	});

	it('refetches when the rank changes, which toString() does not show', () => {
		const { spy } = deferFill();
		const { rerender } = render(<Probe actionIds={[ActionId.fromSpellId(12297, 1)]} />);

		rerender(<Probe actionIds={[ActionId.fromSpellId(12297, 2)]} />);

		expect(spy).toHaveBeenCalledTimes(2);
	});

	it('settles each id on its own: a rejected fill leaves its slot undefined, the others still resolve, and the rejection surfaces', async () => {
		const failure = new Error('no tooltip data');
		vi.spyOn(ActionId.prototype, 'fill').mockImplementation(async function (this: ActionId) {
			if (this.spellId === 1) throw failure;
			return filled(this, 'Second', 'second.jpg');
		});
		const unhandled: Array<unknown> = [];
		const onUnhandledRejection = (error: unknown) => unhandled.push(error);
		process.on('unhandledRejection', onUnhandledRejection);

		try {
			render(<Probe actionIds={[ActionId.fromSpellId(1), ActionId.fromSpellId(2)]} />);
			await act(async () => {});
			await act(async () => await new Promise(resolve => setImmediate(resolve)));

			expect(seen.at(-1)).toEqual([undefined, { iconUrl: 'second.jpg', name: 'Second', href: ActionId.makeSpellUrl(2) }]);
			expect(unhandled).toEqual([failure]);
		} finally {
			process.removeListener('unhandledRejection', onUnhandledRejection);
		}
	});
});
