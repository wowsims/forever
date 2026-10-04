import { comparisonOperatorFieldConfig } from '@features/apl/model/field_descriptors';
import { type EnumTable, resolveField } from '@features/apl/model/field_specs';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import type { ChoiceFieldConfig } from '@features/inputs/configs';
import type { InputConfig } from '@features/inputs/input';
import { APLValueCompare_ComparisonOperator as ComparisonOperator } from '@generated/proto/apl';
import type { Player } from '@sim/player/player';
import { FakeSource } from '@testing/inputs';
import { act, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { EnumField } from './EnumField';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

type Source = FakeSource<number>;

// The comparison operator field as the APL resolves it: a proto enum table with a default label.
const spec = resolveField(comparisonOperatorFieldConfig('op'));
const table = (spec as { table: EnumTable }).table;

const binding = <C,>(id: string) =>
	({
		id,
		storeSubscribe: (source: Source) => source.subscribe,
		getValue: (source: Source) => source.value,
		setValue: (source: Source, next: number) => source.set(next),
	}) as unknown as C;

// The same enum as a settings input lists it: `{ name, value }` rows under `values`.
const settingsConfig: ChoiceFieldConfig<Source> = {
	...binding<ChoiceFieldConfig<Source>>('settings-op'),
	label: 'Operator',
	values: table.options.map(option => ({ name: option.label, value: option.value })),
};

const mount = (value: number) => {
	const apl = new FakeSource(value);
	const settings = new FakeSource(value);
	render(
		<>
			<EnumField player={apl as unknown as Player<any>} config={binding<InputConfig<Player<any>, number> & { id: string }>('apl-op')} table={table} />
			<ConnectedChoiceField modObject={settings} config={settingsConfig} />
		</>,
	);
	return { apl, settings, aplTrigger: document.getElementById('apl-op')!, settingsTrigger: document.getElementById('settings-op')! };
};

const openList = async (trigger: HTMLElement) => {
	await userEvent.click(trigger);
	const listbox = screen.getByRole('listbox');
	return { listbox, rows: within(listbox).getAllByRole('option') };
};

describe('EnumField and a settings enum', () => {
	it('show the stored number enum as the same option, and differ only in the look', () => {
		const { aplTrigger, settingsTrigger } = mount(ComparisonOperator.OpGe);

		expect(aplTrigger.textContent).toBe('rotation_tab.apl.operators.greater_than_or_equal');
		expect(settingsTrigger.textContent).toBe(aplTrigger.textContent);
		expect(aplTrigger.getAttribute('data-variant')).toBe('link');
		expect(settingsTrigger.getAttribute('data-variant')).toBe('select');
	});

	it('open the same list, with the stored value selected in both', async () => {
		const { aplTrigger, settingsTrigger } = mount(ComparisonOperator.OpGe);
		const strip = (html: string) => html.replace(/ id="[^"]*"| aria-[a-z]+="base-ui[^"]*"/g, '');

		const apl = await openList(aplTrigger);
		const aplHtml = strip(apl.listbox.outerHTML);
		const aplSelected = apl.rows.map(row => row.getAttribute('aria-selected'));
		await userEvent.keyboard('{Escape}');

		const settings = await openList(settingsTrigger);
		expect(strip(settings.listbox.outerHTML)).toBe(aplHtml);
		expect(settings.rows.map(row => row.getAttribute('aria-selected'))).toEqual(aplSelected);
		expect(aplSelected.indexOf('true')).toBe(table.options.findIndex(option => option.value === ComparisonOperator.OpGe));
	});

	it('write the picked row back to their sources as the same number', async () => {
		const { apl, settings, aplTrigger, settingsTrigger } = mount(ComparisonOperator.OpGe);
		const label = 'rotation_tab.apl.operators.less_than';

		await userEvent.click((await openList(aplTrigger)).rows.find(row => row.textContent === label)!);
		await userEvent.click((await openList(settingsTrigger)).rows.find(row => row.textContent === label)!);

		expect(apl.value).toBe(ComparisonOperator.OpLt);
		expect(settings.value).toBe(apl.value);
		expect(typeof settings.value).toBe('number');
		expect(aplTrigger.textContent).toBe(label);
		expect(settingsTrigger.textContent).toBe(label);
	});

	it('follow a change landing in their sources the same way', () => {
		const { apl, settings, aplTrigger, settingsTrigger } = mount(ComparisonOperator.OpGe);

		act(() => {
			apl.set(ComparisonOperator.OpNe);
			settings.set(ComparisonOperator.OpNe);
		});

		expect(aplTrigger.textContent).toBe('rotation_tab.apl.operators.not_equals');
		expect(settingsTrigger.textContent).toBe(aplTrigger.textContent);
	});

	it('select nothing and write nothing for a stored value outside the enum; the APL field names the default, the settings select shows its first row', async () => {
		const { apl, settings, aplTrigger, settingsTrigger } = mount(99);

		expect(aplTrigger.textContent).toBe(table.defaultLabel);
		expect(settingsTrigger.textContent).toBe(table.options[0].label);
		for (const trigger of [aplTrigger, settingsTrigger]) {
			const { rows } = await openList(trigger);
			expect(rows.every(row => row.getAttribute('aria-selected') === 'false')).toBe(true);
			await userEvent.keyboard('{Escape}');
		}
		expect([apl.value, settings.value]).toEqual([99, 99]);
	});
});
