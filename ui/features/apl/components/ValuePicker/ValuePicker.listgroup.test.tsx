import { ActionPicker } from '@features/apl/components/ActionPicker';
import { useApl } from '@features/apl/context/AplContext';
import { actionListGroup, APL_ROWS, aplItemId, canApplyMove, PRIORITY, valueListGroup } from '@features/apl/model/apply_move';
import { APLAction, APLRotation, APLValue } from '@generated/proto/apl';
import { makePlayer } from '@testing/apl/player';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ValuePicker } from './ValuePicker';

vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions(() => () => {}));
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const ListProbe = ({ testId }: { testId: string }) => <span data-testid={testId}>{useApl().listGroup}</span>;
vi.mock('@features/apl/components/FieldGroup/fields/ValueFieldList', () => ({ ValueFieldList: () => <ListProbe testId="values-probe" /> }));
vi.mock('@features/apl/components/FieldGroup/fields/ActionFieldList', () => ({ ActionFieldList: () => <ListProbe testId="actions-probe" /> }));

let player: ReturnType<typeof makePlayer>;

beforeEach(() => {
	document.body.innerHTML = '';
});

describe('the list group a picker hands its list field', () => {
	it('names an operand list after the value that owns it, minting the uuid in the same render', () => {
		const value = APLValue.create({ value: { oneofKind: 'and', and: { vals: [] } } });
		player = makePlayer(APLRotation.create());
		player.aplRotation.priorityList = [{ hide: false, notes: '', action: APLAction.create() }];
		player.aplRotation.priorityList[0].action!.condition = value;
		render(<ValuePicker player={player as never} config={{ getValue: () => value, setValue: () => {} }} />);

		expect(value.uuid?.value).toBeTruthy();
		expect(screen.getByTestId('values-probe').textContent).toBe(valueListGroup(value));
		expect(valueListGroup(value)).toBe(`apl-values:${value.uuid!.value}`);
	});

	it('names a sequence’s action list after the sequence, and the move resolves against the rotation', () => {
		const sequence = APLAction.create({ action: { oneofKind: 'sequence', sequence: { name: 'opener', actions: [] } } });
		player = makePlayer(APLRotation.create());
		player.aplRotation.priorityList = [{ hide: false, notes: '', action: sequence }];
		const actions = (sequence.action as any).sequence.actions as Array<APLAction>;
		actions.push(APLAction.create(), APLAction.create());
		render(<ActionPicker player={player as never} config={{ getValue: () => sequence, setValue: () => {} }} />);

		const group = screen.getByTestId('actions-probe').textContent!;
		expect(group).toBe(actionListGroup(sequence));
		const move = { from: { group, id: aplItemId(actions[0], 0), index: 0, parentId: 'row' }, to: { group, parentId: 'row', index: 2 } };
		expect(canApplyMove(player.aplRotation, move)).toBe(true);
		expect(canApplyMove(player.aplRotation, { ...move, to: { group: APL_ROWS, parentId: PRIORITY, index: 0 } })).toBe(false);
	});
});
