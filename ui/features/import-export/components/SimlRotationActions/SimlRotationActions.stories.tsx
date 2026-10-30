import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';

import { SimlRotationActions } from './SimlRotationActions';

const meta = {
	title: 'Features/ImportExport/Components/SimlRotationActions',
	component: SimlRotationActions,
	tags: ['autodocs'],
	decorators: [withSimHost()],
} satisfies Meta<typeof SimlRotationActions>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The block the Rotation tab's sidebar ends with. */
export const Default: Story = {};
