import type { Meta, StoryObj } from '@storybook/react-vite';
import { AVOIDANCE_INFO } from '@testing/storybook/fixtures/character-stats';
import { Table } from '@ui-kit/Table';

import { AvoidanceRow } from './AvoidanceRow';

const meta = {
	title: 'Features/CharacterStats/Components/AvoidanceRow',
	component: AvoidanceRow,
	tags: ['autodocs'],
	args: { info: AVOIDANCE_INFO, withHolyShield: false, hasParry: true, hasBlock: true },
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
} satisfies Meta<typeof AvoidanceRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = { args: { withHolyShield: true, pending: false } };
