import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';

import { ButtonGroup } from './ButtonGroup';

const meta = {
	title: 'Kit/ButtonGroup',
	component: ButtonGroup,
	tags: ['autodocs'],
	argTypes: {
		size: { control: 'select', options: [undefined, 'sm', 'xs'] },
		className: { control: 'text' },
	},
	render: args => (
		<ButtonGroup {...args}>
			<Button variant="outline-primary">Rotation</Button>
			<Button variant="primary">DPS</Button>
			<Button variant="outline-primary">Threat</Button>
		</ButtonGroup>
	),
} satisfies Meta<typeof ButtonGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Small: Story = { args: { size: 'sm' } };
export const Spaced: Story = { args: { gap: true } };
