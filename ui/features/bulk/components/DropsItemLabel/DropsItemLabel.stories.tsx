import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsRowsFixture } from '@testing/storybook/fixtures/bulk-drops';

import { DropsItemLabel } from './DropsItemLabel';

const [upgrade, ring] = dropsRowsFixture();

const meta = {
	title: 'Features/Bulk/DropsItemLabel',
	component: DropsItemLabel,
	tags: ['autodocs'],
	args: upgrade.label!,
	argTypes: { item: { control: false }, moreSources: { control: 'object' } },
} satisfies Meta<typeof DropsItemLabel>;

export default meta;
type Story = StoryObj<typeof meta>;

/** An item with a second source shows "+1" after its first, and lists both in its tooltip. */
export const Default: Story = {};

export const OneSource: Story = { args: ring.label! };

/** An item no source of the run drops names only its slot. */
export const NoSource: Story = { args: { source: '', moreSources: [] } };
