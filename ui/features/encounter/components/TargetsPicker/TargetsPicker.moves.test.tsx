import { Target as TargetProto } from '@generated/proto/common';
import type { Encounter } from '@sim/raid/encounter';
import { createEvent, fireEvent, render, screen } from '@testing-library/react';
import { Tree } from '@ui-kit/Tree';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { TargetsPicker } from './TargetsPicker';

const source = vi.hoisted(() => {
	const listeners = new Set<() => void>();
	return {
		listeners,
		subscribe: (onChange: () => void) => {
			listeners.add(onChange);
			return () => listeners.delete(onChange);
		},
		notify: () => Array.from(listeners).forEach(listener => listener()),
	};
});
vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions(source.subscribe));
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('../../../../tracking/utils', () => ({ trackEvent: vi.fn() }));

// Each target row holds two draggable lists of its own, one in a foreign group and one in the
// targets' group, so their drags reach the targets list's shared controller.
vi.mock('./TargetPicker', () => ({
	TargetPicker: ({ targetIndex }: { targetIndex: number }) => (
		<>
			{['foreign', 'targets'].map(group => (
				<Tree.Root key={group} group={group}>
					<Tree.Items>
						{['a', 'b'].map(name => (
							<Tree.Item key={name} id={`${group}-${targetIndex}-${name}`} testId={`${group}-${targetIndex}-${name}`}>
								<Tree.Item.Header>
									<Tree.Item.Handle label="move" testId={`${group}-${targetIndex}-${name}-handle`} />
								</Tree.Item.Header>
							</Tree.Item>
						))}
					</Tree.Items>
				</Tree.Root>
			))}
		</>
	),
}));

class FakeEncounter {
	constructor(public targets: Array<TargetProto>) {}
	getTargets() {
		return this.targets;
	}
	setTargets(next: Array<TargetProto>) {
		this.targets = next;
		source.notify();
	}
}

const levels = (encounter: FakeEncounter) => encounter.targets.map(target => target.level);

// happy-dom lays nothing out, so each row gets a 20px box, `index` rows from the top.
const place = (testId: string, index: number) =>
	vi.spyOn(screen.getByTestId(testId), 'getBoundingClientRect').mockReturnValue(new DOMRect(0, index * 20, 20, 20));
const grab = (testId: string) => {
	fireEvent.mouseDown(screen.getByTestId(`${testId}-handle`));
	fireEvent.dragStart(screen.getByTestId(testId), { dataTransfer: new DataTransfer() });
};
// happy-dom's DragEvent is a plain Event, which drops `clientY` from its init.
const dropOn = (testId: string, y: number) => {
	const target = screen.getByTestId(testId);
	const event = createEvent.drop(target, { dataTransfer: new DataTransfer() });
	Object.defineProperty(event, 'clientY', { value: y });
	fireEvent(target, event);
};

beforeEach(() => {
	source.listeners.clear();
});

describe('TargetsPicker moves', () => {
	it('ignores a move inside a nested list of another group', () => {
		const encounter = new FakeEncounter([TargetProto.create({ level: 60 }), TargetProto.create({ level: 61 })]);
		render(<TargetsPicker encounter={encounter as unknown as Encounter} />);
		const setTargets = vi.spyOn(encounter, 'setTargets');
		place('foreign-0-a', 0);
		place('foreign-0-b', 1);

		grab('foreign-0-b');
		dropOn('foreign-0-a', 0);

		expect(setTargets).not.toHaveBeenCalled();
		expect(levels(encounter)).toEqual([60, 61]);
	});

	it('ignores a move of a row that is not a target, even in the targets’ own group', () => {
		const encounter = new FakeEncounter([TargetProto.create({ level: 60 }), TargetProto.create({ level: 61 })]);
		render(<TargetsPicker encounter={encounter as unknown as Encounter} />);
		const setTargets = vi.spyOn(encounter, 'setTargets');
		place('targets-1-a', 0);
		place('targets-1-b', 1);

		grab('targets-1-a');
		dropOn('targets-1-b', 35);

		expect(setTargets).not.toHaveBeenCalled();
		expect(levels(encounter)).toEqual([60, 61]);
	});
});
