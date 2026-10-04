import type { Meta, StoryObj } from '@storybook/react-vite';

import { Notice } from './Notice';
import type { NoticeLevel } from './noticeLevel';

const LEVELS: NoticeLevel[] = ['info', 'warning', 'error'];

const meta = {
	title: 'Kit/Notice',
	component: Notice,
	tags: ['autodocs'],
	args: { level: 'warning', children: 'This item is not supported by the sim.' },
	argTypes: {
		level: { control: 'select', options: LEVELS },
		children: { control: 'text' },
		className: { control: 'text' },
	},
} satisfies Meta<typeof Notice>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Info: Story = { args: { level: 'info', children: 'Set bonuses count every equipped piece.' } };
export const Warning: Story = {};
export const ErrorLevel: Story = { name: 'Error', args: { level: 'error', children: 'The rotation has an invalid action.' } };

export const Levels: Story = {
	render: args => (
		<div className="flex flex-col gap-2">
			{LEVELS.map(level => (
				<Notice key={level} {...args} level={level}>
					A {level} notice.
				</Notice>
			))}
		</div>
	),
};
