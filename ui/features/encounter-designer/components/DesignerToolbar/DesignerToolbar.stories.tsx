import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { DesignerToolbar } from './DesignerToolbar';

const meta = {
	title: 'Features/EncounterDesigner/DesignerToolbar',
	component: DesignerToolbar,
	tags: ['autodocs'],
	args: { snapSeconds: 1, onSnapChange: fn(), onZoomIn: fn(), onZoomOut: fn(), onFit: fn(), onAdd: fn() },
} satisfies Meta<typeof DesignerToolbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const SnapOff: Story = { args: { snapSeconds: 0 } };
