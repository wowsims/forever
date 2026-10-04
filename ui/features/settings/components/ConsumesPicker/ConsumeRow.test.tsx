import type { IconEnumPickerConfig } from '@features/inputs/configs';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { createSimStore, patchKeyed, PLAYER_FIELDS, type PlayerSlice, seedKeyed, type SimStore, zeroVersions } from '@sim/state/sim_store';
import { act, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ConsumeRow, ConsumeRowView } from './ConsumeRow';

const KEY = 2;

/** Stands in for the player: the one flag the configs below ask about, over a store the row can select from. */
class Options {
	engineer = true;
	readonly storeKey = KEY;
	readonly sim: { store: SimStore };

	constructor() {
		const store = createSimStore();
		seedKeyed(store, 'players', KEY, { profession1: 0, profession2: 0, v: zeroVersions(PLAYER_FIELDS) } as unknown as PlayerSlice);
		this.sim = { store };
	}

	changeProfession(next: number) {
		patchKeyed(this.sim.store, 'players', KEY, { profession1: next }, ['profession1']);
	}

	changeSomethingElse() {
		patchKeyed(this.sim.store, 'players', KEY, { name: 'Other' }, ['name']);
	}
}

// `iconEnumPickerShown` is satisfied only by a value that carries an actionId *and* is shown, so
// this is the shape of the engineering explosives: one real option behind a profession check.
const configFor = (shown: (options: Options) => boolean): IconEnumPickerConfig<Options, number> =>
	({
		values: [{ value: 0 }, { actionId: {} as never, value: 1, showWhen: shown }],
		equals: (a: number, b: number) => a === b,
		zeroValue: 0,
		storeSubscribe: () => (() => () => {}) as never,
		getValue: () => 0,
		setValue: () => {},
	}) as IconEnumPickerConfig<Options, number>;

const row = (options: Options, configs: Array<IconEnumPickerConfig<Options, number>>) => {
	render(
		<SimHostProvider host={{ player: options } as never}>
			<ConsumeRow name="engineering" configs={configs as never}>
				<div className="consumes-engi" />
			</ConsumeRow>
		</SimHostProvider>,
	);
	return screen.getByTestId('consumes-row');
};

const isHidden = (element: Element) => element.hasAttribute('hidden');

describe('ConsumeRow', () => {
	it('captions the row with its consumable group', () => {
		row(new Options(), [configFor(() => true)]);

		expect(screen.getByRole('group', { name: 'settings_tab.consumables.engineering.title' })).toBe(screen.getByTestId('consumes-row'));
	});

	// Vanilla's updateRow toggled a `hide` class and nothing more, so the pickers inside a hidden
	// row stayed mounted and kept zeroing and restoring the field they are bound to. Unmounting
	// instead would strand a selection the player can no longer use.
	it('hides the row when every picker in it is hidden, without unmounting it, and shows it again', () => {
		const options = new Options();
		const rowElem = row(options, [configFor(opts => opts.engineer), configFor(opts => opts.engineer)]);
		expect(isHidden(rowElem)).toBe(false);

		options.engineer = false;
		act(() => options.changeProfession(1));
		expect(screen.getByTestId('consumes-row')).toBe(rowElem);
		expect(isHidden(rowElem)).toBe(true);
		expect(document.querySelector('.consumes-engi')).not.toBeNull();

		options.engineer = true;
		act(() => options.changeProfession(2));
		expect(isHidden(rowElem)).toBe(false);
	});

	it('keeps the row shown while any one of its pickers is', () => {
		const options = new Options();
		const element = row(options, [configFor(opts => opts.engineer), configFor(() => true)]);

		options.engineer = false;
		act(() => options.changeProfession(1));
		expect(isHidden(element)).toBe(false);
	});

	it('hides a row that names no pickers', () => {
		const element = row(new Options(), []);
		expect(isHidden(element)).toBe(true);
	});

	it('re-evaluates on any player change, not only a profession change', () => {
		// The rows gate on more than professions — imbues on gear and the party's Windfury Totem, the
		// rest on the spec's stats — so this subscribes to the whole player-change aggregate rather
		// than naming individual fields the way the profession-only rows would suggest.
		const options = new Options();
		const element = row(options, [configFor(opts => opts.engineer)]);

		options.engineer = false;
		act(() => options.changeSomethingElse());
		expect(isHidden(element)).toBe(true);

		options.engineer = true;
		act(() => options.changeProfession(1));
		expect(isHidden(element)).toBe(false);
	});
});

describe('ConsumeRowView', () => {
	it('lays the caption out inline before the pickers', () => {
		render(
			<ConsumeRowView label="Elixirs" hidden={false}>
				<div data-testid="pickers" />
			</ConsumeRowView>,
		);
		const rowElem = screen.getByTestId('consumes-row');

		expect(rowElem.getAttribute('data-layout')).toBe('inline');
		expect(rowElem.children[0].textContent).toBe('Elixirs');
		expect(rowElem.children[1]).toBe(screen.getByTestId('pickers'));
	});
});
