import type { Meta, StoryObj } from '@storybook/react-vite';
import { MISS_INFO } from '@testing/storybook/fixtures/character-stats';
import { Table } from '@ui-kit/Table';

import { MissRow } from './MissRow';

const meta = {
	title: 'Features/CharacterStats/Components/MissRow',
	component: MissRow,
	tags: ['autodocs'],
	args: { info: MISS_INFO },
	argTypes: { info: { control: false } },
	decorators: [
		Story => (
			<Table variant="plain" className="max-w-64">
				<tbody>
					<Story />
				</tbody>
			</Table>
		),
	],
} satisfies Meta<typeof MissRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = { args: { pending: false } };
