import type { Meta, StoryObj } from '@storybook/react-vite';

import { UnlaunchedNotice } from './UnlaunchedNotice';

const meta = {
	title: 'Features/Results/SimResultsPanel/Components/UnlaunchedNotice',
	component: UnlaunchedNotice,
	tags: ['autodocs'],
} satisfies Meta<typeof UnlaunchedNotice>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
