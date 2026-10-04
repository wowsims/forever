import type { Meta, StoryObj } from '@storybook/react-vite';
import { PET, PLAYER, TARGET } from '@testing/storybook/fixtures/results-views';

import { EntityLabel } from './EntityLabel';

const meta = {
	title: 'Features/Results/EntityLabel',
	component: EntityLabel,
	tags: ['autodocs'],
	args: { entity: PLAYER },
	argTypes: { entity: { control: false } },
} satisfies Meta<typeof EntityLabel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Player: Story = {};

export const Target: Story = { args: { entity: TARGET } };

export const Pet: Story = { args: { entity: PET } };
