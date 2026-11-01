import type { JsonValue } from '@protobuf-ts/runtime';
import type { SavedDataCodec } from '@ui-kit/hooks/useSavedData';
import { describe, expect, it } from 'vitest';

import { MAX_NAME_LENGTH, planImport, tickedRows, withRowTicked } from './import_plan';

interface Data {
	value: number;
}

const codec: SavedDataCodec<Data> = {
	toJson: data => ({ value: data.value }),
	fromJson: json => {
		if (typeof json?.value !== 'number') throw new Error('not Data');
		return { value: json.value };
	},
};

const saved = (name: string, value: number) => ({ name, json: JSON.stringify({ value }) });

interface Sources {
	saved?: Array<ReturnType<typeof saved>>;
	rejectedNames?: string[];
	presetNames?: string[];
	check?: (data: Data) => void;
}

const plan = (entries: string | Record<string, JsonValue>, { saved = [], rejectedNames = [], ...sources }: Sources = {}) =>
	planImport(typeof entries === 'string' ? JSON.parse(entries) : entries, {
		codec,
		saved,
		storedNames: [...saved.map(({ name }) => name), ...rejectedNames],
		presetNames: [],
		...sources,
	});

const shape = (result: ReturnType<typeof plan>) => result.rows.map(({ saveAs, status, ticked }) => [saveAs, status, ticked]);

describe('planImport', () => {
	it('ticks a new name and keeps it', () => {
		expect(shape(plan({ Raid: { value: 1 } }))).toEqual([['Raid', 'new', true]]);
	});

	it('numbers a saved name with other contents', () => {
		expect(shape(plan({ Raid: { value: 1 } }, { saved: [saved('Raid', 2)] }))).toEqual([['Raid (2)', 'renamed', true]]);
	});

	it('skips the numbers already saved', () => {
		const result = plan({ Raid: { value: 1 } }, { saved: [saved('Raid', 2), saved('Raid (2)', 3), saved('Raid (3)', 4)] });

		expect(shape(result)).toEqual([['Raid (4)', 'renamed', true]]);
	});

	it('leaves identical contents under the saved name unticked, and saves it under a free name if ticked', () => {
		const [row] = plan({ Raid: { value: 1 } }, { saved: [saved('Raid', 1)] }).rows;

		expect(row).toMatchObject({ status: 'already_saved', savedAs: 'Raid', saveAs: 'Raid (2)', ticked: false });
	});

	it('finds identical contents under any saved name', () => {
		const [row] = plan({ Raid: { value: 1 } }, { saved: [saved('Crit test (2)', 1)] }).rows;

		expect(row).toMatchObject({ status: 'already_saved', savedAs: 'Crit test (2)', saveAs: 'Raid', ticked: false });
	});

	it('prefers the entry under the same name when several hold the contents', () => {
		const [row] = plan({ Raid: { value: 1 } }, { saved: [saved('Other', 1), saved('Raid', 1)] }).rows;

		expect(row.savedAs).toBe('Raid');
	});

	it('takes a saved name for already saved when the contents match another entry, whatever the clash', () => {
		const [row] = plan({ Raid: { value: 1 } }, { saved: [saved('Raid', 9), saved('Raid (2)', 1)] }).rows;

		expect(row).toMatchObject({ status: 'already_saved', savedAs: 'Raid (2)', saveAs: 'Raid (3)', ticked: false });
	});

	it('leaves a name for a new entry ahead of an already saved row that would only take it if ticked', () => {
		const result = plan({ ' Raid': { value: 1 }, Raid: { value: 2 } }, { saved: [saved('Other', 1)] });

		expect(shape(result)).toEqual([
			['Raid (2)', 'already_saved', false],
			['Raid', 'new', true],
		]);
	});

	it('takes two identical rows as new when no saved entry matches', () => {
		expect(shape(plan({ A: { value: 1 }, B: { value: 1 } }))).toEqual([
			['A', 'new', true],
			['B', 'new', true],
		]);
	});

	it('numbers a preset name', () => {
		expect(shape(plan({ 'Pre-Raid': { value: 1 } }, { presetNames: ['Pre-Raid'] }))).toEqual([['Pre-Raid (2)', 'renamed', true]]);
	});

	it('takes a name the codec rejected, and steps around it', () => {
		const result = plan({ X: { value: 1 } }, { rejectedNames: ['X'] });

		expect(shape(result)).toEqual([['X (2)', 'renamed', true]]);
	});

	it('steps around a name later in the same file', () => {
		const result = plan({ Raid: { value: 1 }, 'Raid (2)': { value: 2 } }, { saved: [saved('Raid', 9)] });

		expect(shape(result)).toEqual([
			['Raid (3)', 'renamed', true],
			['Raid (2)', 'new', true],
		]);
	});

	it('keeps two names apart that trim to one', () => {
		expect(shape(plan({ ' Raid': { value: 1 }, 'Raid ': { value: 2 } }))).toEqual([
			['Raid', 'new', true],
			['Raid (2)', 'renamed', true],
		]);
	});

	it('counts an entry the codec rejects and lists the good ones', () => {
		const result = plan({ Good: { value: 1 }, Bad: { value: 'x' }, Worse: 'x' });

		expect(shape(result)).toEqual([['Good', 'new', true]]);
		expect(result.unreadable).toBe(2);
	});

	it('counts an entry the check rejects', () => {
		const check = (data: Data) => {
			if (data.value < 0) throw new Error('negative');
		};

		expect(plan({ Good: { value: 1 }, Bad: { value: -1 } }, { check }).unreadable).toBe(1);
	});

	it.each([
		['empty', { '': { value: 1 } }],
		['blank', { '   ': { value: 1 } }],
		['over the cap', { ['x'.repeat(MAX_NAME_LENGTH + 1)]: { value: 1 } }],
		['constructor', { constructor: { value: 1 } }],
		['__proto__', '{"__proto__":{"value":1}}'],
	])('refuses a name that is %s', (_, entries) => {
		const result = plan(entries as Record<string, JsonValue>);

		expect(result.rows).toEqual([]);
		expect(result.unreadable).toBe(1);
	});

	it('takes a name as long as the cap', () => {
		expect(plan({ ['x'.repeat(MAX_NAME_LENGTH)]: { value: 1 } }).rows).toHaveLength(1);
	});

	it('does not see saved names on the prototype chain', () => {
		expect(shape(plan({ toString: { value: 1 } }))).toEqual([['toString', 'new', true]]);
	});

	it('ticks a row by index without touching the plan it was given', () => {
		const before = plan({ A: { value: 1 }, B: { value: 2 } });

		const after = withRowTicked(before, 0, false);

		expect(tickedRows(after).map(row => row.name)).toEqual(['B']);
		expect(tickedRows(before)).toHaveLength(2);
	});
});
