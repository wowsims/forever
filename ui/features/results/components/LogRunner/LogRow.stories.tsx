import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, resourceLog } from '@testing/storybook/fixtures/results-views';

import { LogRow } from './LogRow';

const meta = {
	title: 'Features/Results/LogRunner/Components/LogRow',
	component: LogRow,
	tags: ['autodocs'],
	args: { log: damageLog(65.25, 1420.5) },
	argTypes: { log: { control: false } },
} satisfies Meta<typeof LogRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Damage: Story = {};

export const Resource: Story = { args: { log: resourceLog(65.25, 60, 30) } };

export const BeforeThePull: Story = { args: { log: damageLog(-1.5, 0, { outcome: 'miss', effect: null }) } };
