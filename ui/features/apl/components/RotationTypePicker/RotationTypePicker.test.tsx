import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { source } from '@testing/apl/source';
import { act, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { RotationTypePicker, RotationTypePickerView } from './RotationTypePicker';

vi.mock('@sim/state/subscriptions', async () => {
	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
	return mockSubscriptions((await import('@testing/apl/source')).source.subscribe);
});
vi.mock('@i18n/config', () => ({
	default: {
		t: (key: string) =>
			(
				{
					'rotation_tab.common.rotation_type.auto': 'Auto',
					'rotation_tab.common.rotation_type.simple': 'Simple',
					'rotation_tab.common.rotation_type.apl': 'APL',
				} as Record<string, string>
			)[key] ?? key,
	},
}));

/** The three members the picker reaches for on the host's player. */
class FakePlayer {
	constructor(
		private readonly simple: boolean,
		public type: APLRotationType = APLRotationType.TypeAuto,
	) {}
	hasSimpleRotationGenerator() {
		return this.simple;
	}
	getRotationType() {
		return this.type;
	}
	modifyAplRotation(apply: (rotation: { type: APLRotationType }) => void) {
		const draft = { type: this.type };
		apply(draft);
		this.type = draft.type;
		source.notify();
	}
}

const host = vi.hoisted(() => ({ current: null as unknown }));
vi.mock('@sim/context/SimHostContext', () => ({ useSimHost: () => host.current, useOptionalSimHost: () => host.current }));

const mount = (player: FakePlayer) => {
	host.current = { player };
	return render(<RotationTypePicker />);
};

const trigger = () => screen.getByRole('combobox');
const items = () => screen.queryAllByRole('option');
const open = () => userEvent.click(trigger());

beforeEach(() => source.listeners.clear());

describe('RotationTypePicker', () => {
	it.each([
		[true, ['Auto', 'Simple', 'APL']],
		[false, ['Auto', 'APL']],
	])('offers Simple only when the spec ships a rotation generator (hasSimple=%s)', async (hasSimple, expected) => {
		mount(new FakePlayer(hasSimple));
		await open();
		expect(items().map(item => item.textContent)).toEqual(expected);
	});

	it('shows the rotation type the player currently holds', () => {
		mount(new FakePlayer(true, APLRotationType.TypeAPL));

		expect(trigger().textContent).toBe('APL');
	});

	it('writes the chosen type through modifyAplRotation and follows the store back', async () => {
		const player = new FakePlayer(true);
		mount(player);
		await open();

		await userEvent.click(items()[2]);

		expect(player.type).toBe(APLRotationType.TypeAPL);
		expect(trigger().textContent).toBe('APL');
	});

	it('follows a rotation change made elsewhere', () => {
		const player = new FakePlayer(true);
		mount(player);
		expect(trigger().textContent).toBe('Auto');

		act(() => player.modifyAplRotation(rotation => (rotation.type = APLRotationType.TypeSimple)));
		expect(trigger().textContent).toBe('Simple');
	});
});

describe('RotationTypePickerView', () => {
	it('offers Auto, Simple and APL when the spec has a simple rotation', async () => {
		render(<RotationTypePickerView value={APLRotationType.TypeAuto} hasSimple onChange={() => {}} />);
		await open();

		expect(items().map(item => item.textContent)).toEqual(['Auto', 'Simple', 'APL']);
	});

	it('offers Auto and APL alone otherwise', async () => {
		render(<RotationTypePickerView value={APLRotationType.TypeAuto} hasSimple={false} onChange={() => {}} />);
		await open();

		expect(items().map(item => item.textContent)).toEqual(['Auto', 'APL']);
	});

	it('shows the given type on the trigger, which carries the fixed id', () => {
		render(<RotationTypePickerView value={APLRotationType.TypeAPL} hasSimple onChange={() => {}} />);

		expect(trigger().textContent).toBe('APL');
		expect(trigger().id).toBe('rotation-tab-rotation-type');
	});

	it('hands the chosen type to onChange', async () => {
		const onChange = vi.fn();
		render(<RotationTypePickerView value={APLRotationType.TypeAuto} hasSimple onChange={onChange} />);
		await open();

		await userEvent.click(items()[1]);
		expect(onChange).toHaveBeenCalledWith(APLRotationType.TypeSimple);
	});
});
