import { type APLAction, APLRotation } from '@generated/proto/apl';
import { sequence, waitFor } from '@testing/apl/builders';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { type AplRevealRequest, aplReveals, requestAplReveal } from './reveal';

const rotation = APLRotation.create({
	priorityList: [{ action: waitFor('1s') }, { action: sequence([waitFor('2s'), waitFor('3s')]) }],
	groups: [{ name: 'execute', actions: [{ action: waitFor('4s') }], variables: [] }],
});
const ownSteps = () => (rotation.priorityList[1].action!.action as { sequence: { actions: Array<APLAction> } }).sequence.actions;

let sent: Array<AplRevealRequest>;
let off: () => void;

beforeEach(() => {
	sent = [];
	off = aplReveals.on(request => sent.push(request));
});

afterEach(() => off());

describe('requestAplReveal', () => {
	it('sends the rotation’s own row and the steps down to the one that fired', () => {
		const placed = requestAplReveal(rotation, [
			{ kind: 'priority', index: 1 },
			{ kind: 'step', index: 1 },
		]);

		expect(placed).toBe(true);
		expect(sent).toHaveLength(1);
		expect(sent[0].kind).toBe('priority');
		expect(sent[0].item).toBe(rotation.priorityList[1]);
		expect(sent[0].steps).toHaveLength(1);
		expect(sent[0].steps[0]).toBe(ownSteps()[1]);
	});

	it('sends the group row, not the priority row that referenced its group', () => {
		requestAplReveal(rotation, [
			{ kind: 'priority', index: 0 },
			{ kind: 'group', group: 0, index: 0 },
		]);

		expect(sent).toHaveLength(1);
		expect(sent[0].kind).toBe('group');
		expect(sent[0].item).toBe(rotation.groups[0].actions[0]);
	});

	it('sends a repeat of the same entry again', () => {
		requestAplReveal(rotation, [{ kind: 'priority', index: 0 }]);
		requestAplReveal(rotation, [{ kind: 'priority', index: 0 }]);

		expect(sent).toHaveLength(2);
	});

	it('sends nothing for an entry the rotation no longer holds', () => {
		expect(requestAplReveal(rotation, [{ kind: 'priority', index: 5 }])).toBe(false);
		expect(sent).toEqual([]);
	});
});
