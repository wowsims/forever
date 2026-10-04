import { ActionId } from '@sim/proto/action_id';
import { deferFill, filled } from '@testing/sim/hooks';
import { act, render } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { useActionId } from './useActionId';

const Probe = ({ actionId }: { actionId?: ActionId }) => {
	const { iconUrl, name, href, ready } = useActionId(actionId);
	return <a href={href} data-ready={String(ready)} data-icon={iconUrl} title={name} />;
};

const anchor = (container: HTMLElement) => container.querySelector('a')!;

afterEach(() => vi.restoreAllMocks());

describe('useActionId', () => {
	it('renders an already-filled id on the first pass', () => {
		const { spy } = deferFill();
		const seen: string[] = [];
		const Recorder = ({ actionId }: { actionId: ActionId }) => {
			const { iconUrl } = useActionId(actionId);
			seen.push(iconUrl);
			return null;
		};

		render(<Recorder actionId={filled(ActionId.fromSpellId(1), 'Fireball', 'fireball.jpg')} />);

		expect(seen[0]).toBe('fireball.jpg');
		expect(spy).not.toHaveBeenCalled();
	});

	it('fills an unfilled id and reports ready', async () => {
		const { resolvers } = deferFill();
		const actionId = ActionId.fromSpellId(2);
		const { container } = render(<Probe actionId={actionId} />);

		expect(anchor(container).dataset.icon).toBe('');
		expect(anchor(container).dataset.ready).toBe('false');

		await act(async () => resolvers[0](filled(actionId, 'Frostbolt', 'frostbolt.jpg')));

		expect(anchor(container).dataset.icon).toBe('frostbolt.jpg');
		expect(anchor(container).title).toBe('Frostbolt');
		expect(anchor(container).dataset.ready).toBe('true');
	});

	it('stays silent when the fill rejects after the id is gone', async () => {
		const rejecters: Array<(error: Error) => void> = [];
		vi.spyOn(ActionId.prototype, 'fill').mockImplementation(() => new Promise<ActionId>((_resolve, reject) => rejecters.push(reject)));
		const log = vi.spyOn(console, 'error').mockImplementation(() => {});

		const { unmount } = render(<Probe actionId={ActionId.fromSpellId(2)} />);
		unmount();
		await act(async () => rejecters[0](new Error('no tooltip data')));

		expect(log).not.toHaveBeenCalled();
	});

	it('lets the current id win when an earlier fill resolves last', async () => {
		const { resolvers } = deferFill();
		const first = ActionId.fromSpellId(3);
		const second = ActionId.fromSpellId(4);
		const { container, rerender } = render(<Probe actionId={first} />);

		rerender(<Probe actionId={second} />);
		await act(async () => resolvers[1](filled(second, 'Second', 'second.jpg')));
		await act(async () => resolvers[0](filled(first, 'First', 'first.jpg')));

		expect(anchor(container).dataset.icon).toBe('second.jpg');
	});

	it('drops the previous icon in the render that changes the id', () => {
		deferFill();
		const { container, rerender } = render(<Probe actionId={filled(ActionId.fromSpellId(5), 'Old', 'old.jpg')} />);
		expect(anchor(container).dataset.icon).toBe('old.jpg');

		rerender(<Probe actionId={ActionId.fromSpellId(6)} />);

		expect(anchor(container).dataset.icon).toBe('');
	});

	it('tells two placeholder ids apart by name and icon, since neither carries an id', () => {
		deferFill();
		const { container, rerender } = render(<Probe actionId={ActionId.fromPetName('Cat')} />);
		expect(anchor(container).dataset.icon).toContain('ability_hunter_pet_cat');

		rerender(<Probe actionId={ActionId.fromPetName('Wolf')} />);

		expect(anchor(container).dataset.icon).toContain('ability_hunter_pet_wolf');
	});

	// ActionId.fromItemId is (itemId, tag?, randomSuffixId?), and the random suffix is its
	// url-bearing field: the href is derived before any fill, and re-derived when the id changes.
	it('knows the href without filling, and re-derives it when the random suffix changes', () => {
		deferFill();
		const { container, rerender } = render(<Probe actionId={ActionId.fromItemId(7, 0, 111)} />);
		expect(anchor(container).href).toContain('rand=111');

		rerender(<Probe actionId={ActionId.fromItemId(7, 0, 222)} />);

		expect(anchor(container).href).toContain('rand=222');
	});

	it('resolves an absent id to empty fields without a fill', () => {
		const { spy } = deferFill();
		const { container } = render(<Probe actionId={undefined} />);

		expect(anchor(container).dataset.icon).toBe('');
		expect(anchor(container).getAttribute('href')).toBe('');
		expect(anchor(container).dataset.ready).toBe('true');
		expect(spy).not.toHaveBeenCalled();
	});

	it('uses the spell url for a spell id', () => {
		deferFill();
		const { container } = render(<Probe actionId={ActionId.fromSpellId(8)} />);
		expect(anchor(container).href).toBe(ActionId.makeSpellUrl(8));
	});

	it('carries the rank into the spell url, and leaves it off at rank 0', () => {
		deferFill();
		const { container } = render(<Probe actionId={ActionId.fromSpellId(12297, 3)} />);
		expect(anchor(container).href).toContain('rank=3');
		expect(ActionId.makeSpellUrl(12297, 0)).not.toContain('rank=');
	});

	it('carries the trait definition alongside the rank, as wowhead expects', () => {
		deferFill();
		const { container } = render(<Probe actionId={ActionId.fromTalent(12297, 1, 135506)} />);
		expect(anchor(container).href).toContain('def=135506');
		expect(anchor(container).href).toContain('rank=1');
		expect(ActionId.makeSpellUrl(12297, 0, 0)).not.toContain('def=');
	});

	it('refetches when two talents share a spell id but not a trait definition', () => {
		const { spy } = deferFill();
		const { rerender } = render(<Probe actionId={ActionId.fromTalent(12297, 1, 135506)} />);
		expect(spy).toHaveBeenCalledTimes(1);

		rerender(<Probe actionId={ActionId.fromTalent(12297, 1, 999999)} />);
		expect(spy).toHaveBeenCalledTimes(2);
	});

	it('keeps the definition through fill, which rebuilds the id field by field', async () => {
		vi.spyOn(ActionId, 'getTooltipData').mockResolvedValue({ id: 12297, name: 'Anticipation', icon: 'ability_warrior_anticipation' } as never);

		const resolved = await ActionId.fromTalent(12297, 2, 135506).fill();

		expect(resolved.definitionId).toBe(135506);
		expect(resolved.rank).toBe(2);
		expect(ActionId.makeSpellUrl(resolved.spellId, resolved.rank, resolved.definitionId)).toContain('def=135506');
	});

	it('refetches when the rank changes, and stands pat when it does not', () => {
		const { spy } = deferFill();
		const { rerender } = render(<Probe actionId={ActionId.fromSpellId(12297, 1)} />);
		expect(spy).toHaveBeenCalledTimes(1);

		rerender(<Probe actionId={ActionId.fromSpellId(12297, 1)} />);
		expect(spy).toHaveBeenCalledTimes(1);

		rerender(<Probe actionId={ActionId.fromSpellId(12297, 2)} />);
		expect(spy).toHaveBeenCalledTimes(2);
	});

	it('leaves rank and definition out of ActionId identity, so equals() and equalityKey() agree', () => {
		const talent = ActionId.fromTalent(12297, 3, 135506);
		const plain = ActionId.fromSpellId(12297);

		expect(talent.equals(plain)).toBe(true);
		expect(talent.equalityKey()).toBe(plain.equalityKey());
	});
});
