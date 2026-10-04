import type { Meta, StoryObj } from '@storybook/react-vite';
import { CRIT_IMMUNITY_INFO } from '@testing/storybook/fixtures/character-stats';
import { Table } from '@ui-kit/Table';

import { CritImmunityRow } from './CritImmunityRow';

const meta = {
	title: 'Features/CharacterStats/Components/CritImmunityRow',
	component: CritImmunityRow,
	tags: ['autodocs'],
	args: { info: CRIT_IMMUNITY_INFO, text: 'Immune' },
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
} satisfies Meta<typeof CritImmunityRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = { args: { pending: false } };
