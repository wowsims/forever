import { APLGroup, APLListItem, APLValueVariable } from '@generated/proto/apl';
import { describe, expect, it } from 'vitest';

import { aplClipFromText, aplClipToText, freeName } from './clipboard';

const row = () =>
	APLListItem.fromJson({
		notes: 'opener',
		action: {
			condition: { uuid: { value: 'cond' }, const: { val: '1' } },
			castSpell: { spellId: { spellId: 133 } },
		},
	});

describe('APL clipboard text', () => {
	it('reads back the action it wrote, with new uuids', () => {
		const pasted = aplClipFromText(aplClipToText({ kind: 'action', item: row() }));
		expect(pasted?.kind).toBe('action');
		const item = pasted!.item as APLListItem;
		expect(item.notes).toBe('opener');
		expect(item.action?.action).toEqual(row().action?.action);
		expect(item.action?.condition?.uuid?.value).not.toBe('cond');
	});

	it('reads back a row with no action chosen yet', () => {
		expect(aplClipFromText(aplClipToText({ kind: 'action', item: APLListItem.create({ action: {} }) }))?.kind).toBe('action');
	});

	it('reads back a group and a variable as their own kinds', () => {
		const group = APLGroup.create({ name: 'Cooldowns', actions: [row()] });
		const variable = APLValueVariable.fromJson({ name: 'low_mana', value: { const: { val: '1' } } });
		const pastedGroup = aplClipFromText(aplClipToText({ kind: 'group', item: group }));
		const pastedVariable = aplClipFromText(aplClipToText({ kind: 'variable', item: variable }));
		expect(pastedGroup?.kind).toBe('group');
		expect((pastedGroup!.item as APLGroup).actions).toHaveLength(1);
		expect(pastedVariable?.kind).toBe('variable');
		expect((pastedVariable!.item as APLValueVariable).name).toBe('low_mana');
	});

	it('drops fields this build does not know, so text from another build still pastes', () => {
		const pasted = aplClipFromText('{"kind": "group", "value": {"name": "Cooldowns", "fieldFromAnotherBuild": 1}}');
		expect(pasted?.kind).toBe('group');
		expect((pasted!.item as APLGroup).name).toBe('Cooldowns');
	});

	it('refuses text that is not a copied action, group or variable', () => {
		expect(aplClipFromText('not json')).toBeUndefined();
		expect(aplClipFromText('{}')).toBeUndefined();
		expect(aplClipFromText(APLListItem.toJsonString(row()))).toBeUndefined();
		expect(aplClipFromText('{"kind": "action", "value": {}}')).toBeUndefined();
		expect(aplClipFromText('{"kind": "group", "value": {"bogus": 1}}')).toBeUndefined();
		expect(aplClipFromText('{"kind": "toString", "value": {}}')).toBeUndefined();
	});
});

describe('freeName', () => {
	it('keeps a free name and otherwise takes the first free number', () => {
		expect(freeName('Cooldowns', ['Opener'])).toBe('Cooldowns');
		expect(freeName('Cooldowns', ['Cooldowns', 'Cooldowns (2)'])).toBe('Cooldowns (3)');
		expect(freeName('Cooldowns', ['Cooldowns', 'Cooldowns (3)'])).toBe('Cooldowns (2)');
	});

	it('numbers a numbered name from its base, never nesting the suffix', () => {
		expect(freeName('Berserk (2)', ['Berserk', 'Berserk (2)'])).toBe('Berserk (3)');
	});
});
