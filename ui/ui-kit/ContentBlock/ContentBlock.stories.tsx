import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';

import { ContentBlock } from './ContentBlock';

const meta = {
	title: 'Kit/ContentBlock',
	component: ContentBlock,
	tags: ['autodocs'],
	args: {
		config: {},
		children: <p>Patchwerk, 180 seconds, one target.</p>,
		testId: 'story-content-block',
	},
	argTypes: {
		children: { control: false },
		headerChildren: { control: false },
	},
	render: args => (
		<div className="flex flex-col">
			<ContentBlock {...args} />
			<p className="text-secondary">Next block</p>
		</div>
	),
} satisfies Meta<typeof ContentBlock>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithHeader: Story = {
	args: {
		config: { header: { title: 'Encounter' } },
		headerChildren: (
			<Button variant="link" size="sm" className="ml-auto">
				Reset
			</Button>
		),
	},
};

export const WithTooltip: Story = {
	args: { config: { header: { title: 'Encounter', tooltip: 'The fight the sim runs: <b>one boss</b>, fixed duration.' } } },
};

export const Flush: Story = { args: { config: { header: { title: 'Encounter' } }, flush: true } };

export const WithoutBody: Story = { args: { config: { header: { title: 'Encounter' }, withoutBody: true } } };
