import type { Meta, StoryObj } from '@storybook/react-vite';
import { CRIT_CAP_INFO } from '@testing/storybook/fixtures/character-stats';
import { Table } from '@ui-kit/Table';

import { CritCapRow } from './CritCapRow';

const meta = {
	title: 'Features/CharacterStats/Components/CritCapRow',
	component: CritCapRow,
	tags: ['autodocs'],
	args: { info: CRIT_CAP_INFO, text: '8.40% under' },
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
} satisfies Meta<typeof CritCapRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = { args: { info: { ...CRIT_CAP_INFO, specOffset: 1.5 }, pending: false } };
