import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { fn } from 'storybook/test';

import { FakeLink } from './FakeLink';

const meta = {
	title: 'Kit/FakeLink',
	component: FakeLink,
	tags: ['autodocs'],
	args: { onActivate: fn() },
	argTypes: { onActivate: { control: false } },
} satisfies Meta<typeof FakeLink>;

export default meta;
type Story = StoryObj<typeof meta>;

/** A row that activates anywhere: the layer sits last, and the button lifted above it keeps its own click. */
export const Row: Story = {
	render: args => (
		<div className="relative flex w-96 items-center gap-4 bg-gray-900 p-2 hover:bg-gray-800">
			<span>Click anywhere in the row</span>
			<Button size="sm" className="relative z-1 ml-auto">
				Own action
			</Button>
			<FakeLink {...args} />
		</div>
	),
};
