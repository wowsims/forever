import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { NameDisplay } from './NameDisplay';

const meta = {
	title: 'Features/Apl/NameDisplay',
	component: NameDisplay,
	tags: ['autodocs'],
	args: { name: 'Execute Phase', onRename: fn() },
} satisfies Meta<typeof NameDisplay>;

export default meta;
type Story = StoryObj<typeof meta>;

/** An action group's, a value variable's or, with `compact`, a placeholder's name inside a field. */
export const Default: Story = { args: { compact: false } };
