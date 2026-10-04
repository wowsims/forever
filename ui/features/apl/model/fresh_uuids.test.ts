import { APLListItem, APLValue } from '@generated/proto/apl';
import { describe, expect, it } from 'vitest';

import { freshUuids } from './fresh_uuids';

const constant = (val: string, uuid: string) => APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'const', const: { val } } });

describe('freshUuids', () => {
	it('gives every value in a copied row a new uuid, nested operands included, and keeps the content', () => {
		const original = APLListItem.create({
			action: {
				condition: APLValue.create({
					uuid: { value: 'and' },
					value: { oneofKind: 'and', and: { vals: [constant('1', 'one'), constant('2', 'two')] } },
				}),
				action: { oneofKind: 'wait', wait: { duration: constant('1s', 'wait') } },
			},
		});
		const copy = freshUuids(APLListItem.clone(original));

		const uuids = (row: APLListItem) => {
			const condition = row.action!.condition!;
			return [
				condition.uuid!.value,
				...(condition.value as any).and.vals.map((value: APLValue) => value.uuid!.value),
				(row.action!.action as any).wait.duration.uuid.value,
			];
		};
		expect(uuids(original)).toEqual(['and', 'one', 'two', 'wait']);
		expect(new Set([...uuids(original), ...uuids(copy)]).size).toBe(8);
		const withoutUuids = (row: APLListItem) => JSON.parse(JSON.stringify(row, (key, value) => (key === 'uuid' ? undefined : value)));
		expect(withoutUuids(copy)).toEqual(withoutUuids(original));
	});

	it('returns the message it was handed and leaves a value without a uuid alone', () => {
		const value = APLValue.create({ value: { oneofKind: 'const', const: { val: '1' } } });
		expect(freshUuids(value)).toBe(value);
		expect(value.uuid).toBeUndefined();
	});
});
