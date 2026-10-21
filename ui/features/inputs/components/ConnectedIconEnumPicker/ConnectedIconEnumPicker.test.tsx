import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import { FakeSource, filled } from '@testing/inputs';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { StrictMode } from 'react';
import { afterEach, beforeEach, describe, expect, expectTypeOf, it, vi } from 'vitest';

import { ConnectedIconEnumPicker } from './ConnectedIconEnumPicker';
import type { IconEnumPickerConfig } from './types';
import { iconEnumPickerShown } from './utils';

const frostId = filled(ActionId.fromSpellId(7302), 'Frost Armor', 'frost.jpg');
const moltenId = filled(ActionId.fromSpellId(30482), 'Molten Armor', 'molten.jpg');

/** Stands in for the player: the armor value, plus the flags the configs read. */
class Options extends FakeSource<number> {
	engineer = true;
	visible = true;
	enabled = true;
	columns = 2;
	constructor() {
		super(0);
	}
	override set(next: number) {
		// The real facades drop equal writes (`Player.setConsumes` returns early), so the guard mirrors that here.
		if (this.value === next) return;
		super.set(next);
	}
	setEngineer(next: boolean) {
		this.engineer = next;
		this.notify();
	}
	setVisible(next: boolean) {
		this.visible = next;
		this.notify();
	}
}

const configFor = (extra: Partial<IconEnumPickerConfig<Options, number>> = {}): IconEnumPickerConfig<Options, number> => ({
	values: [
		{ value: 0, tooltip: 'No Armor' },
		{ actionId: frostId, value: 1 },
		{ actionId: moltenId, value: 2, tooltip: 'Molten' },
	],
	equals: (a, b) => a === b,
	zeroValue: 0,
	storeSubscribe: options => options.subscribe,
	getValue: options => options.value,
	setValue: (options, next) => options.set(next),
	...extra,
});

const withEngineerOnly = () =>
	configFor({ values: [{ value: 0 }, { actionId: frostId, value: 1 }, { actionId: moltenId, value: 2, showWhen: (obj: Options) => obj.engineer }] });

const mount = (options: Options, config: IconEnumPickerConfig<Options, number> = configFor()) =>
	render(<ConnectedIconEnumPicker modObject={options} config={config} />);

const root = () => screen.queryByTestId('icon-enum-picker-root');
const button = () => root() as HTMLAnchorElement;
const menu = () => screen.queryByTestId('icon-enum-picker-menu') as HTMLUListElement | null;
const items = () => Array.from(menu()!.children) as HTMLLIElement[];
const optionAnchor = (index: number) => within(items()[index]).getByTestId('icon-enum-option') as HTMLAnchorElement;
const open = () =>
	act(() => {
		fireEvent.mouseEnter(button());
		fireEvent.mouseMove(button());
	});
const iconOf = (element: HTMLElement) => element.style.backgroundImage.replace(/^url\(['"]?|['"]?\)$/g, '');

beforeEach(() => {
	vi.spyOn(ActionId.prototype, 'fill').mockImplementation(async function (this: ActionId) {
		return this;
	});
});
afterEach(() => vi.restoreAllMocks());

describe('ConnectedIconEnumPicker', () => {
	it('resolves each value’s action id into its option’s icon and link', () => {
		const options = new Options();
		options.value = 2;
		mount(options);
		expect(iconOf(button())).toBe('molten.jpg');
		expect(button().getAttribute('href')).toBe(ActionId.makeSpellUrl(30482));

		open();
		expect(items().map((_item, index) => iconOf(optionAnchor(index)))).toEqual(['', 'frost.jpg', 'molten.jpg']);
		expect(optionAnchor(1).getAttribute('href')).toBe(ActionId.makeSpellUrl(7302));
	});

	it('shows a value’s buffActionId buff tooltip under its actionId’s icon, on the option and on the selection, from the start', async () => {
		const tooltipData = vi
			.spyOn(ActionId, 'makeSpellTooltipData')
			.mockImplementation(async (id, params) => `spell=${id}${params?.useBuffAura ? '&buff=1' : ''}`);
		const options = new Options();
		options.value = 2;
		mount(
			options,
			configFor({
				values: [{ value: 0 }, { actionId: frostId, value: 1 }, { actionId: moltenId, buffActionId: ActionId.fromSpellId(18789), value: 2 }],
			}),
		);
		await act(async () => {});
		expect(iconOf(button())).toBe('molten.jpg');
		expect(button().getAttribute('href')).toBe(ActionId.makeSpellUrl(18789));
		expect(button().getAttribute('data-wowhead')).toBe('spell=18789&buff=1');
		expect(tooltipData.mock.calls.map(([id]) => id)).toEqual([18789]);

		open();
		expect(button().getAttribute('data-wowhead')).toBe('spell=18789&buff=1');
		expect(optionAnchor(2).getAttribute('data-wowhead')).toBe('spell=18789&buff=1');
		await act(async () => {});
		expect(iconOf(optionAnchor(2))).toBe('molten.jpg');
		expect(optionAnchor(1).hasAttribute('data-wowhead')).toBe(false);
		expect(optionAnchor(1).getAttribute('href')).toBe(ActionId.makeSpellUrl(7302));
		expect(tooltipData).toHaveBeenCalledTimes(1);
	});

	it('leaves an option whose buff tooltip lookup fails without one, keeps the others, and surfaces the failure', async () => {
		const failure = new Error('lookup failed');
		vi.spyOn(ActionId, 'makeSpellTooltipData').mockImplementation(async id => {
			if (id === 18790) throw failure;
			return `spell=${id}`;
		});
		const unhandled: Array<unknown> = [];
		const onUnhandledRejection = (error: unknown) => unhandled.push(error);
		process.on('unhandledRejection', onUnhandledRejection);

		try {
			mount(
				new Options(),
				configFor({
					values: [
						{ value: 0 },
						{ actionId: moltenId, buffActionId: ActionId.fromSpellId(18789), value: 1 },
						{ actionId: frostId, buffActionId: ActionId.fromSpellId(18790), value: 2 },
					],
				}),
			);
			await act(async () => {});
			await act(async () => await new Promise(resolve => setImmediate(resolve)));

			open();
			expect(optionAnchor(1).getAttribute('data-wowhead')).toBe('spell=18789');
			expect(optionAnchor(2).hasAttribute('data-wowhead')).toBe(false);
			expect(unhandled).toEqual([failure]);
		} finally {
			process.removeListener('unhandledRejection', onUnhandledRejection);
		}
	});

	it('fills the selected option at mount, the others once the menu first opens, and a hidden one never', async () => {
		const fill = vi.spyOn(ActionId.prototype, 'fill').mockImplementation(async function (this: ActionId) {
			return filled(this, `Spell ${this.spellId}`, `${this.spellId}.jpg`);
		});
		const filledSpells = () => new Set(fill.mock.contexts.map(actionId => (actionId as ActionId).spellId));
		const options = new Options();
		options.value = 2;
		mount(
			options,
			configFor({
				values: [
					{ actionId: ActionId.fromSpellId(101), value: 1 },
					{ actionId: ActionId.fromSpellId(102), value: 2 },
					{ actionId: ActionId.fromSpellId(103), value: 3 },
					{ actionId: ActionId.fromSpellId(104), value: 4, showWhen: () => false },
				],
			}),
		);
		await act(async () => {});
		expect(filledSpells()).toEqual(new Set([102]));
		expect(iconOf(button())).toBe('102.jpg');

		open();
		await act(async () => {});
		expect(filledSpells()).toEqual(new Set([101, 102, 103]));
		expect(items().map((_item, index) => iconOf(optionAnchor(index)))).toEqual(['101.jpg', '102.jpg', '103.jpg']);
		expect(iconOf(button())).toBe('102.jpg');
	});

	it('writes the chosen value to the source', () => {
		const options = new Options();
		mount(options);
		open();
		act(() => {
			fireEvent.click(optionAnchor(2));
		});
		expect(options.value).toBe(2);
	});

	it('unmounts a value whose showWhen is false and brings the option back with it', () => {
		const options = new Options();
		mount(options, withEngineerOnly());
		open();
		expect(items()).toHaveLength(3);

		act(() => options.setEngineer(false));
		expect(items()).toHaveLength(2);

		act(() => options.setEngineer(true));
		expect(items()).toHaveLength(3);
		expect(optionAnchor(2).getAttribute('href')).toBe(ActionId.makeSpellUrl(30482));
	});

	it('zeroes the selection when the option holding it is hidden', () => {
		const options = new Options();
		options.value = 2;
		mount(options, withEngineerOnly());
		expect(options.value).toBe(2);

		act(() => options.setEngineer(false));
		expect(options.value).toBe(0);
	});

	it('zeroes once when the picker and the option holding the selection hide in the same notification', () => {
		const options = new Options();
		options.value = 2;
		const setValue = vi.fn((obj: Options, next: number) => obj.set(next));
		mount(
			options,
			configFor({
				showWhen: (obj: Options) => obj.visible,
				setValue,
				values: [{ value: 0 }, { actionId: frostId, value: 1 }, { actionId: moltenId, value: 2, showWhen: (obj: Options) => obj.engineer }],
			}),
		);

		act(() => {
			options.visible = false;
			options.engineer = false;
			options.notify();
		});
		expect(setValue.mock.calls.map(([, next]) => next)).toEqual([0]);

		// Back in view with the option still hidden: the put-aside value comes back and is zeroed in the same pass.
		act(() => options.setVisible(true));
		expect(options.value).toBe(0);
		expect(setValue.mock.calls.map(([, next]) => next)).toEqual([0, 2, 0]);
	});

	it('renders no picker at all unless some option carries an actionId and is shown', () => {
		// A list of colour-only values names nothing to show.
		mount(new Options(), configFor({ values: [{ value: 0, color: 'grey' }] }));
		expect(root()).toBeNull();
	});

	it('unmounts on its own showWhen, and puts the value aside across it', () => {
		const options = new Options();
		options.value = 2;
		mount(options, configFor({ showWhen: (obj: Options) => obj.visible }));
		expect(root()).toBeTruthy();

		act(() => options.setVisible(false));
		expect(root()).toBeNull();
		expect(options.value).toBe(0);

		// The source was still zero, so what was put aside comes back.
		act(() => options.setVisible(true));
		expect(root()).toBeTruthy();
		expect(options.value).toBe(2);
	});

	it('lets a value set while the picker was hidden beat the one put aside', () => {
		const options = new Options();
		options.value = 2;
		mount(options, configFor({ showWhen: (obj: Options) => obj.visible }));

		act(() => options.setVisible(false));
		expect(options.value).toBe(0);
		// A preset writing while the picker is away.
		act(() => options.set(1));
		act(() => options.setVisible(true));
		expect(options.value).toBe(1);
	});

	it('does not put a value aside for a picker that mounts hidden', () => {
		const options = new Options();
		options.value = 2;
		options.visible = false;
		// Store/restore is driven by the source subscription, so it never runs during construction, and
		// the revision guard keeps StrictMode's second mount render from reading as a transition.
		render(
			<StrictMode>
				<ConnectedIconEnumPicker modObject={options} config={configFor({ showWhen: (obj: Options) => obj.visible })} />
			</StrictMode>,
		);
		expect(root()).toBeNull();
		expect(options.value).toBe(2);

		// Nothing was put aside, so coming back is a no-op rather than a restore.
		act(() => options.setVisible(true));
		expect(options.value).toBe(2);
	});

	it('disables the picker on enableWhen', () => {
		mount(new Options(), configFor({ enableWhen: () => false }));
		expect(root()!.hasAttribute('data-disabled')).toBe(true);
	});

	it('reads a numColumns function off the mod object', () => {
		const options = new Options();
		options.columns = 4;
		mount(options, configFor({ numColumns: obj => obj.columns }));
		open();
		expect(menu()!.style.gridTemplateColumns).toBe('repeat(4, 1fr)');
	});

	it('reads a values function again on every opening of the menu, with no notification between', () => {
		const options = new Options();
		options.value = 1;
		const frost = { actionId: frostId, value: 1 };
		const molten = { actionId: moltenId, value: 2 };
		let ordered = [frost, molten];
		mount(options, configFor({ values: () => ordered }));
		const shownOrder = () => items().map((_item, index) => iconOf(optionAnchor(index)));
		open();
		expect(shownOrder()).toEqual(['frost.jpg', 'molten.jpg']);
		// Choosing the value already held closes the menu without a write, so nothing notifies.
		act(() => {
			fireEvent.click(optionAnchor(0));
		});
		expect(button().getAttribute('aria-expanded')).toBe('false');
		ordered = [molten, frost];
		open();
		expect(shownOrder()).toEqual(['molten.jpg', 'frost.jpg']);
	});

	it('falls back to backupIconUrl for a value the list does not carry', () => {
		const options = new Options();
		options.value = 99;
		mount(options, configFor({ backupIconUrl: () => moltenId }));

		expect(iconOf(button())).toBe('molten.jpg');
		expect(button().getAttribute('href')).toBe(ActionId.makeSpellUrl(30482));
		expect(button().hasAttribute('data-active')).toBe(true);
	});

	it('carries the config frame onto the field and the testId onto the swatch', () => {
		const { container } = render(<ConnectedIconEnumPicker modObject={new Options()} config={configFor({ label: 'Armor' })} testId="armor-root" />);
		const field = container.firstElementChild as HTMLElement;
		expect(screen.getByTestId('armor-root').tagName).toBe('A');
		expect(within(field).getByTestId('form-label').textContent).toBe('Armor');
	});

	it('marks the trigger kept while keptBy holds, and follows the source', () => {
		const options = new Options();
		options.set(2);
		mount(options, configFor({ keptBy: obj => obj.value === 2 }));
		expect(button().parentElement!.hasAttribute('data-kept')).toBe(true);
		expect(screen.getByTestId('icon-swatch-kept').getAttribute('data-tooltip-content')).toBe(i18n.t('stat_filters.kept_tooltip'));

		act(() => options.set(1));
		expect(button().parentElement!.hasAttribute('data-kept')).toBe(false);
		expect(screen.queryByTestId('icon-swatch-kept')).toBeNull();
	});
});

// Asserted directly because a caller outside the component depends on it: a `.consumes-row` hides
// when every picker in it is hidden, and it has only the configs to ask.
describe('iconEnumPickerShown', () => {
	const options = new Options();

	it('needs a value that carries an actionId', () => {
		expect(
			iconEnumPickerShown(
				configFor({
					values: [
						{ value: 0, color: 'grey' },
						{ value: 1, iconUrl: 'raw.jpg' },
					],
				}),
				options,
			),
		).toBe(false);
		expect(iconEnumPickerShown(configFor(), options)).toBe(true);
	});

	it('ignores an actionId value that is itself hidden', () => {
		const hiddenOption = configFor({ values: [{ value: 0 }, { actionId: frostId, value: 1, showWhen: () => false }] });
		expect(iconEnumPickerShown(hiddenOption, options)).toBe(false);
	});

	it('honours the config’s own showWhen as well', () => {
		expect(iconEnumPickerShown(configFor({ showWhen: () => false }), options)).toBe(false);
		expect(iconEnumPickerShown(configFor({ showWhen: () => true }), options)).toBe(true);
	});
});

describe('IconEnumPickerConfig', () => {
	type WithoutBinding = Omit<IconEnumPickerConfig<Options, number>, 'storeSubscribe' | 'storeField'>;
	type FieldOnly = WithoutBinding & { storeField: 'player:*' };

	it('will not accept a config that names neither a store subscription nor a store field', () => {
		expectTypeOf<WithoutBinding>().not.toExtend<IconEnumPickerConfig<Options, number>>();
		expectTypeOf<FieldOnly>().toExtend<IconEnumPickerConfig<Options, number>>();
	});
});
