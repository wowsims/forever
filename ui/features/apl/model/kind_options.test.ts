import { APLRotation } from '@generated/proto/apl';
import { makePlayer } from '@testing/apl/player';
import { describe, expect, it, vi } from 'vitest';

import { actionKinds } from './action_kinds';
import { actionKindOptions, valueKindOptions } from './kind_options';
import { valueKinds } from './value_kinds';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const player = (): any => makePlayer(APLRotation.create());

describe('kind options includeIf', () => {
	it.each<[string, (isPrepull: boolean) => Array<{ value: string }>, string]>([
		['valueKindOptions', isPrepull => valueKindOptions(player(), isPrepull, false), 'currentTime'],
		['actionKindOptions', isPrepull => actionKindOptions(player(), isPrepull), 'multidot'],
	])('%s drops a prepull-excluded kind when isPrepull is true, and keeps it otherwise', (_name, options, kind) => {
		expect(options(false).map(option => option.value)).toContain(kind);
		expect(options(true).map(option => option.value)).not.toContain(kind);
	});
});

describe('kind tooltip', () => {
	it('wraps the short description and appends the full description when one exists', () => {
		const options = valueKindOptions(player(), false, false);
		const withFull = options.find(option => option.value === 'const');
		expect(withFull?.tooltip).toBe(`<p>${valueKinds.const.shortDescription}</p> ${valueKinds.const.fullDescription}`);
	});

	it('is just the short description when there is no full description', () => {
		const options = valueKindOptions(player(), false, false);
		const shortOnly = options.find(option => option.value === 'cmp');
		expect(valueKinds.cmp.fullDescription).toBeUndefined();
		expect(shortOnly?.tooltip).toBe(valueKinds.cmp.shortDescription);
	});
});

describe('valueKindOptions dynamicStringResolver', () => {
	it('is included whenever not prepull, and resolves through the always-empty resolver', () => {
		const options = valueKindOptions(player(), false, false);
		const generic = options.find(option => option.value === 'currentGenericResource');

		expect(generic).toBeDefined();
		expect(generic?.label).toBe('');
		expect(generic?.tooltip).toBe('');
	});

	it('is excluded during prepull', () => {
		const options = valueKindOptions(player(), true, false);
		expect(options.map(option => option.value)).not.toContain('currentGenericResource');
	});
});

describe('table insertion order', () => {
	it.each<[string, Array<string>, () => Array<{ value: string }>]>([
		['valueKindOptions', Object.keys(valueKinds), () => valueKindOptions(player(), false, false)],
		['actionKindOptions', Object.keys(actionKinds), () => actionKindOptions(player(), false)],
	])('%s preserves its kind table order', (_name, order, options) => {
		const indices = options().map(option => order.indexOf(option.value));
		expect(indices).toEqual([...indices].sort((a, b) => a - b));
		expect(new Set(indices).size).toBe(indices.length);
	});
});
