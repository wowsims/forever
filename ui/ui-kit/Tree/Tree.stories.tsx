import type { Meta, StoryObj } from '@storybook/react-vite';
import { type ReactNode, useState } from 'react';
import { expect, fireEvent, fn, screen, userEvent, within } from 'storybook/test';

import { Tree } from './index';
import type { TreeDensity, TreeDropRef, TreeLayout, TreeMove, TreeRef } from './types';

const DENSITIES: Array<TreeDensity> = ['default', 'compact'];
const LAYOUTS: Array<TreeLayout> = ['default', 'inline'];

interface TreeStoryProps {
	density?: TreeDensity;
	layout?: TreeLayout;
	disabled?: boolean;
	onAdd?: () => void;
	onDelete?: () => void;
	onMove?: (move: TreeMove) => void;
}

const ROWS = ['Opener', 'Main priority', 'Cooldowns'];

const Row = ({ id, title, onDelete, testId, children }: { id: string; title: string; onDelete?: () => void; testId?: string; children?: ReactNode }) => (
	<Tree.Item id={id} testId={testId}>
		<Tree.Item.Header>
			<Tree.Item.Handle label={`Move ${title}`} />
			<Tree.Item.Title>{title}</Tree.Item.Title>
			<Tree.Item.Trigger label={`Toggle ${title}`} />
			<Tree.Item.Actions label={`${title} actions`}>
				<Tree.Item.Action icon="copy" label="Copy" />
				<Tree.Item.Action icon="xmark" label="Delete" tone="danger" onClick={onDelete} />
			</Tree.Item.Actions>
		</Tree.Item.Header>
		<Tree.Item.Body>The body of {title}.</Tree.Item.Body>
		{children}
	</Tree.Item>
);

const TreeStory = ({ density, layout, disabled, onAdd = () => {}, onDelete }: TreeStoryProps) => (
	<Tree.Root density={density} layout={layout} disabled={disabled}>
		<Tree.Items>
			{ROWS.map(title => (
				<Row key={title} id={title} title={title} onDelete={onDelete} />
			))}
		</Tree.Items>
		<Tree.AddButton label="New row" onClick={onAdd} />
	</Tree.Root>
);

const InlineTree = ({ onDelete }: { onDelete?: () => void }) => (
	<Tree.Root layout="inline">
		<Tree.Items>
			{ROWS.map(title => (
				<Tree.Item key={title} id={title}>
					<Tree.Item.Header>
						<Tree.Item.Handle label={`Move ${title}`} />
						<Tree.Item.Aside>ON</Tree.Item.Aside>
						<Tree.Item.Actions label={`${title} actions`}>
							<Tree.Item.Action icon="xmark" label="Delete" tone="danger" onClick={onDelete} />
						</Tree.Item.Actions>
					</Tree.Item.Header>
					<Tree.Item.Body>
						The body of {title}, on the header row. The handle, the aside and the menu stay on its first line when it wraps.
					</Tree.Item.Body>
				</Tree.Item>
			))}
		</Tree.Items>
		<Tree.AddButton label="New row" icon="plus" tone="success" onClick={() => {}} />
	</Tree.Root>
);

const isValues = (group: string) => group.startsWith('values-');

const BodyListTree = ({ onMove }: { onMove?: (move: TreeMove) => void }) => {
	const [lists, setLists] = useState<Record<string, Array<string>>>({ Opener: ['Potion', 'Charge'], Burst: [] });
	const move = (next: TreeMove) => {
		onMove?.(next);
		setLists(current => {
			const from = next.from.parentId!;
			const to = next.to.parentId!;
			const moved = current[from][next.from.index];
			const source = current[from].filter((_, index) => index !== next.from.index);
			const at = from === to && next.to.index > next.from.index ? next.to.index - 1 : next.to.index;
			return { ...current, [from]: source, [to]: insertAt(from === to ? source : current[to], moved, at) };
		});
	};
	return (
		<Tree.Root group="rows" sortable={{ onMove: move, canDrop: (from, to) => isValues(from.group) === isValues(to.group) }}>
			<Tree.Items>
				{Object.entries(lists).map(([id, values]) => (
					<Tree.Item key={id} id={id} testId={`row-${id}`}>
						<Tree.Item.Header>
							<Tree.Item.Title>{id}</Tree.Item.Title>
						</Tree.Item.Header>
						<Tree.Item.Body>
							<Tree.Root group={`values-${id}`} layout="inline" density="compact" testId={`list-${id}`}>
								<Tree.Items>
									{values.map(value => (
										<Tree.Item key={value} id={value} testId={`value-${value}`}>
											<Tree.Item.Header>
												<Tree.Item.Handle label={`Move ${value}`} />
											</Tree.Item.Header>
											<Tree.Item.Body>{value}</Tree.Item.Body>
										</Tree.Item>
									))}
								</Tree.Items>
								<Tree.AddButton label={`New value in ${id}`} icon="plus" tone="success" onClick={() => {}} />
							</Tree.Root>
						</Tree.Item.Body>
					</Tree.Item>
				))}
			</Tree.Items>
		</Tree.Root>
	);
};

interface Group {
	id: string;
	rows: Array<string>;
}

const GROUPS: Array<Group> = [
	{ id: 'Opener', rows: ['Potion', 'Charge'] },
	{ id: 'Burst', rows: ['Recklessness', 'Execute'] },
];

const insertAt = <T,>(list: Array<T>, item: T, index: number) => [...list.slice(0, index), item, ...list.slice(index)];

const applyMove = (groups: Array<Group>, { from, to }: TreeMove): Array<Group> => {
	const shift = from.parentId === to.parentId && to.index > from.index ? 1 : 0;
	if (from.parentId === null) {
		return insertAt(
			groups.filter(group => group.id !== from.id),
			groups[from.index],
			to.index - shift,
		);
	}
	return groups
		.map(group => (group.id === from.parentId ? { ...group, rows: group.rows.filter(row => row !== from.id) } : group))
		.map(group => (group.id === to.parentId ? { ...group, rows: insertAt(group.rows, from.id, to.index - shift) } : group));
};

const sameLevel = (from: TreeRef, to: TreeDropRef) => (from.parentId === null) === (to.parentId === null);

const SortableTree = ({ onMove }: { onMove?: (move: TreeMove) => void }) => {
	const [groups, setGroups] = useState(GROUPS);
	return (
		<Tree.Root
			group="rotation"
			sortable={{
				onMove: move => {
					onMove?.(move);
					setGroups(current => applyMove(current, move));
				},
				canDrop: sameLevel,
			}}>
			<Tree.Items>
				{groups.map(group => (
					<Row key={group.id} id={group.id} title={group.id} testId={`row-${group.id}`}>
						<Tree.Item.Children>
							<Tree.Items>
								{group.rows.map(row => (
									<Row key={row} id={row} title={row} testId={`row-${row}`} />
								))}
							</Tree.Items>
						</Tree.Item.Children>
					</Row>
				))}
			</Tree.Items>
		</Tree.Root>
	);
};

const meta = {
	title: 'Kit/Tree',
	component: TreeStory,
	tags: ['autodocs'],
	args: { onAdd: fn(), onDelete: fn(), onMove: fn() },
	argTypes: {
		density: { control: 'select', options: DENSITIES },
		layout: { control: 'select', options: LAYOUTS },
	},
} satisfies Meta<typeof TreeStory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Nested: Story = {
	render: ({ onDelete }) => (
		<Tree.Root>
			<Tree.Items>
				<Row id="group" title="Group" onDelete={onDelete}>
					<Tree.Item.Children>
						<Tree.Items>
							<Row id="sequence" title="Sequence" onDelete={onDelete}>
								<Tree.Item.Children>
									<Tree.Items>
										<Row id="step-1" title="Step 1" onDelete={onDelete} />
										<Row id="step-2" title="Step 2" onDelete={onDelete} />
									</Tree.Items>
								</Tree.Item.Children>
							</Row>
						</Tree.Items>
					</Tree.Item.Children>
				</Row>
			</Tree.Items>
		</Tree.Root>
	),
};

export const Compact: Story = { args: { density: 'compact' } };

export const Inline: Story = {
	render: ({ onDelete }) => <InlineTree onDelete={onDelete} />,
};

export const Collapsed: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const toggle = canvas.getByRole('button', { name: 'Toggle Opener' });
		await expect(toggle).toHaveAttribute('aria-expanded', 'true');
		await userEvent.click(toggle);
		await expect(toggle).toHaveAttribute('aria-expanded', 'false');
	},
};

export const Disabled: Story = { args: { disabled: true } };

export const WithActions: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Opener actions' }));
		await userEvent.click(await screen.findByRole('button', { name: 'Delete' }));
		await expect(args.onDelete).toHaveBeenCalledTimes(1);
	},
};

export const Sortable: Story = {
	render: ({ onMove }) => <SortableTree onMove={onMove} />,
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await fireEvent.mouseDown(canvas.getByRole('button', { name: 'Move Potion' }));
		await fireEvent.dragStart(canvas.getByTestId('row-Potion'), { dataTransfer: new DataTransfer() });
		await fireEvent.drop(canvas.getByTestId('row-Recklessness'), { dataTransfer: new DataTransfer(), clientY: 0 });

		await expect(args.onMove).toHaveBeenCalledTimes(1);
		await expect(args.onMove).toHaveBeenCalledWith({
			from: { group: 'rotation', id: 'Potion', index: 0, parentId: 'Opener' },
			to: { group: 'rotation', parentId: 'Burst', index: 0 },
		});
		await expect(within(canvas.getByTestId('row-Burst')).getByTestId('row-Potion')).toBeInTheDocument();
	},
};

/** A list in each row's body, of its own group: it indents like children, and an empty one still takes a drop. */
export const BodyList: Story = {
	render: ({ onMove }) => <BodyListTree onMove={onMove} />,
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByTestId('list-Burst').hasAttribute('data-in-body')).toBe(true);
		await fireEvent.mouseDown(canvas.getByRole('button', { name: 'Move Potion' }));
		await fireEvent.dragStart(canvas.getByTestId('value-Potion'), { dataTransfer: new DataTransfer() });
		await fireEvent.drop(canvas.getByTestId('list-Burst'), { dataTransfer: new DataTransfer(), clientY: 0 });

		await expect(args.onMove).toHaveBeenCalledWith({
			from: { group: 'values-Opener', id: 'Potion', index: 0, parentId: 'Opener' },
			to: { group: 'values-Burst', parentId: 'Burst', index: 0 },
		});
		await expect(within(canvas.getByTestId('list-Burst')).getByTestId('value-Potion')).toBeInTheDocument();
	},
};

export const InsideRowFlex: Story = {
	render: args => (
		<div className="flex items-start gap-4">
			<span>Rotation</span>
			<TreeStory {...args} />
		</div>
	),
};

export const Narrow: Story = {
	render: ({ onDelete }) => (
		<div className="w-64">
			<InlineTree onDelete={onDelete} />
		</div>
	),
};
