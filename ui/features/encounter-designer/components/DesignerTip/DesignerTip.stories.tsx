import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { DesignerTip } from './DesignerTip';

const meta = {
	title: 'Features/EncounterDesigner/DesignerTip',
	component: DesignerTip,
	tags: ['autodocs'],
} satisfies Meta<typeof DesignerTip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Wave: Story = {
	args: { text: ["Adds #2 — Add wave 'Adds'", '1 add (Lvl 63 Mechanical)', 'Spawns 45s ± 2s', 'Lives 10s', 'Every 30s', 'Kill share 25 %'].join('\n') },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText("Adds #2 — Add wave 'Adds'")).toBeVisible();
		await expect(canvas.getByText('Kill share 25 %')).toBeVisible();
	},
};
